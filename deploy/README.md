# Docker Deployment Build

To build, run the following from the root of the repository:

## Building

```bash
docker build -f deploy/Dockerfile .
```

## Usage

The docker image will contain two critical files:

* `/oasis/bin/faucet-backend` - This is the binary for serving the faucet backend
* `/oasis/bin/frontend-serve.sh` - This is a script that will build the frontend
  static files for serving. This is intended for use as an init-container in
  kubernetes. If a persistent volume is used to store the compiled frontend
  files, this script will only run when there are changes in the source
  directory. This script expects 2 environment variables:
  * This file is called like: `/oasis/bin/frontend-serve.sh [src_dir] [dest_dir]`
  * The following environment variables are expected:
    * `CAPTCHA_SITE_KEY` - This is the recaptcha site key
    * `REQUEST_AMOUNT` - This is the amount of tokens to request during funding
