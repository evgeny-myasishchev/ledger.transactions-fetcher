#!/usr/bin/env sh

IMAGE=evgenymyasishchev/transactions-fetcher:latest

docker create --name=transactions-fetcher \
              -v ${PWD}/config/fetchers:/go/src/config/fetchers \
              -v ${PWD}/db:/go/src/db \
              ${IMAGE}