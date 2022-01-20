# Faucet frontend

- Frontend works well enough even without javascript
- Only runtime javascript dependency is google's recaptcha, so maintenance can be minimal


## Run

```sh
yarn
yarn dev

edit faucet-backend.toml
  web_root = "../faucet-frontend/dist"

(cd ../faucet-backend && go1.17.6 build && ./faucet-backend)

open http://localhost:8080/
```


## Configure

See [./.env](./.env) for overridable environment variables (note: overriding via `.env.local` or `.env.development` files doesn't seem to work).


## Build for production

```sh
rm -r ./dist
CAPTCHA_SITE_KEY=___ yarn build
```
