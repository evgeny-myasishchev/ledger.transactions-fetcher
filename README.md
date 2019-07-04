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