```sh
mkdir data
(cd ./data && oasis-node registry entity init)

(cd ./faucet-backend && go1.17.6 build)
(cd ./faucet-frontend && yarn)

(cd ./faucet-frontend && CAPTCHA_SITE_KEY=6LcNnAQeAAAAAHkPRq4iQls3gTFwhLGn0Djxz8Uw yarn dev) # dev-only key
./faucet-backend/faucet-backend -f ./dev-backend-config.toml
# Transfer some TEST tokens to printed address.

Open http://localhost:8080/
```

- It says `failed to fund account: failed to submit transaction` when it is out of funds.

- This uses captcha site key and secret intended for development only. Generated at https://www.google.com/recaptcha/admin/site/503618573:
  - type: reCAPTCHA v2; "I'm not a robot" Checkbox
  - domains: 127.0.0.1 and localhost
