# ledger.transactions-fetcher
Fetch transactions from banks and report them to ledger

## Running locally

Add google client-id/secret to config

Setup the db:
```
go run ./cmd/storage/ -cmd setup
```

Get oauth code. To do this you need to generate code grant url, navigate it and authenticate with ledger known account:
```
go run ./cmd/auth/ -cmd auth-url
```

Register the user:
```
go run ./cmd/auth/ -cmd register-user -code "XXX"
```
Where XXX is a code obtained with a previous step

Prepare merchant config. Create file config/fetchers/\<email\>.json:
```
{
    "UserID": "<email>",
    "Merchants": {
        "<account-id>": {
            "ID": "<merchant-id>",
            "Password": "<merchant-password>",
            "BankAccount": "<bank-account>"
        }
    }
}
```

Here and below:
* `<account-id>` is a ledger account id

Fetch transactions:

```
go run ./cmd/fetch-transactions/ -bank=pbanua2x -acc <account-id> -days 5 -user <email> | npx pino-pretty
```

Sync fetched transactions:

```
go run ./cmd/ledger/ -cmd sync -user <email> -account <account-id> | npx pino-pretty
```

## Dev

### Generated mocks

Some mocks are generated with `mockgen`. Generate commands are added to Makefile (see mockgen target). Please add new mocks there.

Make sure to regenerate mocks if updating interfaces (e.g make mockgen).

## Prod

### Naïve approach

Create a folder on a target server, something like: ~/ledger-services/prod/transactions-fetcher

Create a `env` file with contents similar to below:
```
APP_ENV=production
CONFIG_REMOTE_STORAGE=local
GOOGLE_CLIENT_ID=<xxx>.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=<yyy>
```

Create `fetch.sh` file with contents similar to below:
```
#!/usr/bin/env sh

set -e

days=${days:-5}

fetch-transactions -acc <ledger-account-1> -days ${days} -user <user@email.com>
fetch-transactions -acc <ledger-account-2> -days ${days} -user <user@email.com>
fetch-transactions -acc <ledger-account-3> -days ${days} -user <user@email.com>
```

Create `sync.sh` file with contents similar to below:
```
#!/usr/bin/env sh

set -e

ledger -cmd sync -user <user@email.com> -account <ledger-account-1>
ledger -cmd sync -user <user@email.com> -account <ledger-account-2>
ledger -cmd sync -user <user@email.com> -account <ledger-account-3>
```

Copy `bin/create-containers.sh` to the folder and run it. To rerun it remove containers first:
```
docker rm transactions-fetcher-fetch
docker rm transactions-fetcher-sync
docker rm transactions-fetcher-shell
```

Setup local db:
```
mkdir db
docker run \
    --env-file env \
    -v ${PWD}/db:/go/src/db \
    --rm \
    evgenymyasishchev/transactions-fetcher:latest storage -cmd setup
```

Create user fetchers config file `config/fetchers/user@email.com.json`:
```
{
    "UserID": "user@email.com",
    "Merchants": {
        "<ledger-account-x>": {
            "ID": "<id>",
            "Password": "<pass>",
            "BankAccount": "<bank account>"
        }
    }
}
```

Now fetch transactions:
```
docker start -ai transactions-fetcher-fetch
```

And sync fetched transactions:
```
docker start -ai transactions-fetcher-sync
```

Cron it every 10 minutes:

```
*/10 * * * * docker start -ai transactions-fetcher-fetch && docker start -ai transactions-fetcher-sync
```