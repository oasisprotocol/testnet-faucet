# Faucet frontend

## Setup

```sh
yarn


cd ../faucet-backend
mkdir ./data
initialize entity in ./data and transfer some TEST tokens

edit faucet-backend.toml
  data_dir = "./data"
  web_root = "../faucet-frontend/dist"
  recaptcha_shared_secret = dev-only secret from .env
```


## Run

```sh
yarn dev

(cd ../faucet-backend && go1.17.6 build && ./faucet-backend)

open http://localhost:8080/
```


## Build for production

```sh
rm -r .parcel-cache
rm -r dist
CAPTCHA_SITE_KEY=___ yarn build
```
