FROM golang:1.22 as backend_builder

WORKDIR /usr/src/app

COPY faucet-backend/ /usr/src/app

RUN go build .

FROM node:20

WORKDIR /oasis/frontend

COPY faucet-frontend/ /oasis/frontend

RUN yarn

COPY deploy/frontend-serve.sh /oasis/bin/frontend-serve.sh
COPY --from=backend_builder /usr/src/app/faucet-backend /oasis/bin/faucet-backend
