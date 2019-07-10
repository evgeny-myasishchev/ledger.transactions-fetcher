# ledger.transactions-fetcher
Fetch transactions from banks and report them to ledger

## Running locally

Add google client-id/secret to config

Setup the db:
```
go run cmd/storage/main.go -cmd setup
```

Get oauth code. To do this you need to generate code grant url, navigate it and authenticate with ledger known account:
```
go run cmd/auth/*.go -cmd auth-url
```

Register the user:
```
go run cmd/auth/*.go -cmd register-user -code "XXX"
```
Where XXX is a code obtained with a previous step

PBANUA2X: Prepare merchant config. Create file config/fetchers/\<email\>.json:
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
go run cmd/fetch-transactions/main.go -acc <account-id> -days 5 -user <email> | npx pino-pretty
```

Sync fetched transactions:

```
go run cmd/ledger/main.go -cmd sync -user <email> -account <account-id> | npx pino-pretty
```

## Dev

### Generated mocks

Some mocks are generated with `mockgen`. Generate commands are added to Makefile (see mockgen target). Please add new mocks there.

Make sure to regenerate mocks if updating interfaces (e.g make mockgen).