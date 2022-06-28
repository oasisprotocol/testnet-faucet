package main

import (
	"context"
	"time"

	consensus "github.com/oasisprotocol/oasis-core/go/consensus/api"
	consensusTx "github.com/oasisprotocol/oasis-core/go/consensus/api/transaction"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"

	"github.com/oasisprotocol/oasis-sdk/client-sdk/go/config"
	"github.com/oasisprotocol/oasis-sdk/client-sdk/go/connection"
	"github.com/oasisprotocol/oasis-sdk/client-sdk/go/modules/consensusaccounts"
	"github.com/oasisprotocol/oasis-sdk/client-sdk/go/types"
)

// From https://pkg.go.dev/github.com/oasisprotocol/oasis-sdk/client-sdk/go@v0.2.0/config#pkg-variables
var paratimeIdToName = map[string]string{
	"0000000000000000000000000000000000000000000000000000000000000000": "cipher",
	"00000000000000000000000000000000000000000000000072c8215e60d5bca7": "emerald",
}

// Returns the name of the paratime corresponding to paratimeId.
func (svc *Service) paratimeName(paratimeId string) string {
	for name, paratime := range svc.network.ParaTimes.All {
		if paratimeId == paratime.ID {
			return name
		}
	}
	// Should never happen since the input has been validated
	// by the frontend.
	svc.log.Printf("bank/paratime: unknown paratime id %s",
		paratimeId,
	)
	return "unknown_paratime"
}

type FundRequest struct {
	ParaTime *config.ParaTime
	Account  *types.Address

	ConsensusAmount *types.Quantity
	ParaTimeAmount  *types.BaseUnits
}

func (svc *Service) BankWorker() {
	svc.log.Printf("bank: started")

	// XXX: Wire into termination.
	ctx := context.Background()

	var (
		conn connection.Connection
		err  error
	)
	for {
		svc.log.Printf("bank: attempting to connect to gRPC endpoint")
		if conn, err = connection.Connect(ctx, svc.network); err != nil {
			svc.log.Printf("bank: failed to connect to node: %v", err)
			time.Sleep(15 * time.Second)
			continue
		}
		break
	}

	// Refill the allowances.
	svc.RefillAllowances(ctx, conn)

	svc.log.Printf("bank: connected to gRPC endpoint")

	// Mark as ready to accept requests.
	close(svc.readyCh)

	refillTicker := time.NewTicker(1 * time.Hour)
	for {
		select {
		case req := <-svc.fundRequestCh:
			// Note: Access control, validation, and non-debug logging is
			// handled by the frontend.
			if req.ParaTime == nil {
				svc.FundConsensusRequest(ctx, conn, req)
			} else {
				svc.FundParaTimeRequest(ctx, conn, req)
			}
		case <-refillTicker.C:
			svc.RefillAllowances(ctx, conn)
		case <-svc.quitCh:
			return
		}
	}
}

func (svc *Service) FundConsensusRequest(ctx context.Context, conn connection.Connection, req *FundRequest) {
	defer svc.ClearAddress(req.Account)

	var elapsed time.Duration
	start := time.Now()

	xfer := staking.Transfer{
		To:     req.Account.ConsensusAddress(),
		Amount: *req.ConsensusAmount,
	}
	tx := staking.NewTransferTx(0, new(consensusTx.Fee), &xfer)
	if err := svc.SignAndSubmitConsensusTx(ctx, conn, tx); err != nil {
		svc.log.Printf("bank/consesus: failed to submit tx (%v: %v): %v",
			xfer.To.String(),
			xfer.Amount.String(),
			err,
		)
		svc.metrics.Requests.WithLabelValues("consensus", "failure").Inc()
		return
	}

	svc.log.Printf("bank/consensus: request successful: %v: %v TEST",
		xfer.To.String(),
		xfer.Amount.String(),
	)

	elapsed = time.Since(start)
	svc.metrics.RequestLatencies.WithLabelValues("consensus").Observe(elapsed.Seconds())
	svc.metrics.Requests.WithLabelValues("consensus", "success").Inc()
}

func (svc *Service) FundParaTimeRequest(ctx context.Context, conn connection.Connection, req *FundRequest) {
	var submitOk bool
	defer func() {
		if !submitOk {
			svc.ClearAddress(req.Account)
		}
	}()

	var elapsed time.Duration
	start := time.Now()
	reqParatimeName := svc.paratimeName(req.ParaTime.ID)

	// Just asssume that there is sufficient allowance, and that the periodic
	// refill adequately handles keeping the allowance topped off.

	depositBody := &consensusaccounts.Deposit{
		To:     req.Account,
		Amount: *req.ParaTimeAmount,
	}
	tx := consensusaccounts.NewDepositTx(nil, depositBody)
	watcher, err := svc.SignAndSubmitMetaTx(ctx, conn, req.ParaTime, tx)
	if err != nil {
		svc.log.Printf("bank/paratime: failed to submit tx (%v: %v): %v",
			depositBody.To.String(),
			depositBody.Amount.String(),
			err,
		)
		svc.metrics.Requests.WithLabelValues(reqParatimeName, "failure").Inc()
		return
	}

	submitOk = true
	go func() {
		defer func() {
			svc.ClearAddress(req.Account)
		}()

		ev := <-watcher.ResultCh
		if ev == nil {
			svc.log.Printf("bank/paratime: failed to wait for event: %v", watcher.Context.Err())
			svc.metrics.Requests.WithLabelValues(reqParatimeName, "failure").Inc()
			return
		}

		if !ev.IsSuccess() {
			svc.log.Printf("bank/paratime: tx failed with error: module: %s code: %d",
				ev.Error.Module,
				ev.Error.Code,
			)
			svc.metrics.Requests.WithLabelValues(reqParatimeName, "failure").Inc()
			return
		}

		svc.log.Printf("bank/paratime: request successful: %v: %v TEST",
			depositBody.To.String(),
			depositBody.Amount.String(),
		)

		elapsed = time.Since(start)
		svc.metrics.RequestLatencies.WithLabelValues(reqParatimeName).Observe(elapsed.Seconds())
		svc.metrics.Requests.WithLabelValues(reqParatimeName, "success").Inc()
	}()
}

func (svc *Service) RefillAllowances(ctx context.Context, conn connection.Connection) {
	// Failures are ignored under the assumption that there is sufficient allowance
	// already.
	svc.log.Printf("bank: refilling allowances")

	// Query the existing allowances.
	consensusAccount, err := conn.Consensus().Staking().Account(ctx, &staking.OwnerQuery{
		Height: consensus.HeightLatest,
		Owner:  svc.address,
	})
	if err != nil {
		svc.log.Printf("bank: failed to query funding account: %v", err)
		return
	}
	svc.metrics.Balances.WithLabelValues("consensus").Set(float64(consensusAccount.General.Balance.ToBigInt().Uint64()))

	for ptName, pt := range svc.network.ParaTimes.All {
		ptAddr := staking.NewRuntimeAddress(pt.Namespace())
		allowance := consensusAccount.General.Allowances[ptAddr]

		svc.metrics.Balances.WithLabelValues(ptName).Set(float64(allowance.ToBigInt().Uint64()))

		svc.log.Printf("refill: %v allowance: %v", ptName, allowance)

		// Figure out if we need to increase.
		toFund := svc.cfg.TargetAllowance.Clone()
		if err = toFund.Sub(&allowance); err != nil || toFund.IsZero() {
			svc.log.Printf("bank: paratime '%s' already has sufficient allowance: %v", ptName, allowance)
			continue
		}

		// Build the staking transaction.
		allow := staking.Allow{
			Beneficiary:  ptAddr,
			Negative:     false,
			AmountChange: *toFund,
		}
		tx := staking.NewAllowTx(0, new(consensusTx.Fee), &allow)
		if err := svc.SignAndSubmitConsensusTx(ctx, conn, tx); err != nil {
			svc.log.Printf("bank: failed to add allowance to paratime '%s': %v", ptName, err)
		}
	}
}
