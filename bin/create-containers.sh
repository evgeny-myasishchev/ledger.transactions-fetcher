#!/usr/bin/env sh

IMAGE=evgenymyasishchev/transactions-fetcher:latest

COMMON_PARAMS="--env-file ${PWD}/env \
              -v ${PWD}/config/fetchers:/go/src/config/fetchers \
              -v ${PWD}/fetch.sh:/go/src/fetch.sh \
              -v ${PWD}/sync.sh:/go/src/sync.sh \
              -v ${PWD}/db:/go/src/db \
              ${IMAGE}"

docker create --name=transactions-fetcher-fetch \
              ${COMMON_PARAMS} \
              ./fetch.sh

docker create --name=transactions-fetcher-sync \
              ${COMMON_PARAMS} \
              ./sync.sh
