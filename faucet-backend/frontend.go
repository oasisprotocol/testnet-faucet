package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/oasisprotocol/oasis-sdk/client-sdk/go/helpers"
	"github.com/oasisprotocol/oasis-sdk/client-sdk/go/types"
)

const (
	queryParaTime          = "paratime"
	queryAccount           = "account"
	queryAmount            = "amount"
	queryRecaptchaResponse = "g-recaptcha-response"
)

var accountPrefixes = map[string][]string{
	"":         []string{"oasis"}, // Consensus.
	"emerald":  []string{"0x"},
	"cipher":   []string{"oasis"},
	"sapphire": []string{"0x", "oasis"},
}

// isValidAccountPrefixForParaTime checks if the given account address string has a valid
// prefix for use with the given paratime name based on the accountPrefixes map.
func isValidAccountPrefixForParaTime(paraTimeStr string, accountStr string) (bool, error) {
	prefixes, ok := accountPrefixes[strings.ToLower(paraTimeStr)]
	if !ok {
		return false, fmt.Errorf("frontend: unknown paratime type")
	}
	for _, p := range prefixes {
		if strings.HasPrefix(accountStr, p) {
			return true, nil
		}
	}
	return false, nil
}

func (svc *Service) TestAndSetAddress(addr *types.Address) bool {
	svc.dedupLock.Lock()
	defer svc.dedupLock.Unlock()

	addrStr := addr.String()

	ret := svc.dedupMap[addrStr]
	svc.dedupMap[addrStr] = true
	return ret
}

func (svc *Service) ClearAddress(addr *types.Address) {
	svc.dedupLock.Lock()
	defer svc.dedupLock.Unlock()

	svc.dedupMap[addr.String()] = false
}

func (svc *Service) FrontendWorker() {
	defer func() {
		close(svc.doneCh)
	}()

	svc.log.Printf("frontend: started")

	// Register API endpoints.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/fund", svc.OnFundRequest)
	if svc.cfg.WebRoot != "" {
		mux.Handle("/", http.FileServer(http.Dir(svc.cfg.WebRoot)))
	}

	srv := &http.Server{
		Addr:    svc.cfg.ListenAddr,
		Handler: mux,
	}

	// Wait till the part that does the actual heavy lifting is initialized.
	<-svc.readyCh

	svc.log.Printf("frontend: bank ready, starting HTTP server")

	// Serve.
	go func() {
		defer close(svc.quitCh)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		svc.log.Printf("frontend: user requested termination")

		if err := srv.Shutdown(context.Background()); err != nil {
			svc.log.Printf("frontend: failed graceful HTTP server shutdown: %v", err)
		}
	}()
	switch {
	case svc.cfg.TLSCertFile != "" || svc.cfg.TLSKeyFile != "":
		if err := srv.ListenAndServeTLS(svc.cfg.TLSCertFile, svc.cfg.TLSKeyFile); err != http.ErrServerClosed {
			svc.log.Printf("frontend: failed to start HTTPs server: %v", err)
			return
		}
	default:
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			svc.log.Printf("frontend: failed to start HTTP server: %v", err)
			return
		}
	}

	// Wait till all pending requests have been serviced.
	<-svc.quitCh
}

// onFundRequest handles a funding request.  The expect request is a POST of
// the form `https://host:port/api/v1/fund&account=CONSENSUS_ACCOUNT_ID&amount=TOKENS`.
func (svc *Service) OnFundRequest(w http.ResponseWriter, req *http.Request) {
	writeResult := func(statusCode int, result error) {
		type fundResponse struct {
			Result string `json:"result"`
		}

		w.WriteHeader(statusCode)
		b, _ := json.Marshal(&fundResponse{
			Result: result.Error(),
		})
		_, _ = w.Write(b)
	}

	// Ensure the user is POSTing, if auth is enabled.
	authEnabled := svc.cfg.RecaptchaSharedSecret != ""
	if authEnabled {
		if req.Method != http.MethodPost {
			svc.log.Printf("frontend: invalid http method: '%v'", req.Method)
			writeResult(
				http.StatusMethodNotAllowed,
				fmt.Errorf("invalid http method: '%v'", req.Method),
			)
			return
		}
	}

	// Parse the query and POST form (combined).
	if err := req.ParseForm(); err != nil {
		svc.log.Printf("frontend: invalid http request: %v", err)
		writeResult(
			http.StatusBadRequest,
			fmt.Errorf("invalid http request, failed to parse query/form"),
		)
		return
	}

	var (
		err     error
		fundReq FundRequest
	)

	// ParaTime/Account
	paraTimeStr := strings.TrimSpace(req.Form.Get(queryParaTime))
	accountStr := strings.TrimSpace(req.Form.Get(queryAccount))

	prefixValid, err := isValidAccountPrefixForParaTime(paraTimeStr, accountStr)
	if err != nil {
		svc.log.Printf("frontend: invalid paratime: '%v'", paraTimeStr)
		writeResult(
			http.StatusInternalServerError,
			fmt.Errorf("failed to fund account: invalid paratime: '%v'", paraTimeStr),
		)
		return
	}

	if paraTimeStr != "" {
		// Paratime account
		fundReq.ParaTime = svc.network.ParaTimes.All[paraTimeStr]
		if fundReq.ParaTime == nil {
			svc.log.Printf("frontend: invalid paratime: '%v'", paraTimeStr)
			writeResult(
				http.StatusInternalServerError,
				fmt.Errorf("failed to fund account: invalid paratime: '%v'", paraTimeStr),
			)
			return
		}
		if !prefixValid {
			svc.log.Printf("frontend: account not a paratime address: '%v'", accountStr)
			writeResult(
				http.StatusInternalServerError,
				fmt.Errorf("failed to fund account: invalid account: not a paratime address"),
			)
			return
		}
	} else if !prefixValid {
		// Consensus account
		svc.log.Printf("frontend: account not an oasis address: '%v'", accountStr)
		writeResult(
			http.StatusInternalServerError,
			fmt.Errorf("failed to fund account: invalid account: not an oasis address"),
		)
		return
	}

	if fundReq.Account, fundReq.EthAccount, err = helpers.ResolveEthOrOasisAddress(accountStr); err != nil {
		svc.log.Printf("frontend: invalid account '%v': %v", accountStr, err)
		writeResult(
			http.StatusInternalServerError,
			fmt.Errorf("failed to fund account: invalid account: '%v'", accountStr),
		)
		return
	}

	// Amount
	amountStr := strings.TrimSpace(req.Form.Get(queryAmount))
	switch fundReq.ParaTime {
	case nil:
		if fundReq.ConsensusAmount, err = helpers.ParseConsensusDenomination(
			svc.network,
			amountStr,
		); err != nil {
			svc.log.Printf("frontend: invalid amount '%v': %v", amountStr, err)
			writeResult(
				http.StatusInternalServerError,
				fmt.Errorf("failed to fund account: invalid amount: '%v'", amountStr),
			)
			return
		}
		if !svc.cfg.MaxConsensusFundAmount.IsZero() {
			max := svc.cfg.MaxConsensusFundAmount.Clone()
			if err = max.Sub(fundReq.ConsensusAmount); err != nil {
				svc.log.Printf("frontend: excessive consensus amount: %v", fundReq.ConsensusAmount)
				writeResult(
					http.StatusInternalServerError,
					fmt.Errorf("failed to fund account: excessive consensus amount: '%v'", amountStr),
				)
				return
			}
		}
	default:
		if fundReq.ParaTimeAmount, err = helpers.ParseParaTimeDenomination(
			fundReq.ParaTime,
			amountStr,
			types.NativeDenomination, // XXX: Make this configurable.
		); err != nil {
			svc.log.Printf("frontend: invalid amount '%v': %v", amountStr, err)
			writeResult(
				http.StatusInternalServerError,
				fmt.Errorf("failed to fund account: invalid amount: '%v'", amountStr),
			)
			return
		}
		if maxStr := svc.cfg.MaxParatimeFundAmount; maxStr != "" {
			max, err := helpers.ParseParaTimeDenomination(
				fundReq.ParaTime,
				maxStr,
				types.NativeDenomination,
			)
			if err != nil {
				svc.log.Printf("frontend: invalid maximum amount '%v': %v", maxStr, err)
				writeResult(
					http.StatusInternalServerError,
					fmt.Errorf("failed to fund account: per-paratime max misconfigured"),
				)
				return
			}
			if err = max.Amount.Sub(&fundReq.ParaTimeAmount.Amount); err != nil {
				svc.log.Printf("frontend: excessive paratime amount: %v", fundReq.ParaTimeAmount)
				writeResult(
					http.StatusInternalServerError,
					fmt.Errorf("failed to fund account: excessive paratime amount: '%v'", amountStr),
				)
				return
			}
		}
	}

	// Handle reCAPTCHA integration, if enabled.
	if authEnabled {
		// Technically not a query, but the server has a unified view of
		// POST form and query fields.
		if err = svc.CheckRecaptcha(req.Form.Get(queryRecaptchaResponse)); err != nil {
			svc.log.Printf("frontend: reCAPTCHA failed: %v", err)
			writeResult(
				http.StatusForbidden,
				fmt.Errorf("failed to verify reCAPTCHA"),
			)
			return
		}
	}

	// Ensure the address does not have a request in-flight already.
	if svc.TestAndSetAddress(fundReq.Account) {
		// User is being a greedy asshole, fail.
		writeResult(
			http.StatusForbidden,
			fmt.Errorf("funding request already pending, try again later"),
		)
		return
	}

	// Attempt to fund the address.
	select {
	case svc.fundRequestCh <- &fundReq:
	default:
		// Queue backlog full, fail early.
		svc.ClearAddress(fundReq.Account)
		writeResult(
			http.StatusInternalServerError,
			fmt.Errorf("temporary failure, try again later"),
		)
		return
	}

	svc.log.Printf("frontend: request enqueued: [%v]%v: %v TEST", paraTimeStr, accountStr, amountStr)

	writeResult(
		http.StatusOK,
		fmt.Errorf("funding request submitted"),
	)
}
