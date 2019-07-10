package dal

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"

	// This has to be here to let go mods work work
	_ "github.com/mattn/go-sqlite3"
)

type nowFn func() time.Time

func defaultNowFn() time.Time {
	return time.Now()
}

type sqlStorage struct {
	db    *sql.DB
	nowFn nowFn
}

func (s *sqlStorage) Setup(ctx context.Context) error {
	logger.Info(ctx, "Setup SQL storage")
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS users(
	email nvarchar(30) NOT NULL PRIMARY KEY,
	refresh_token NTEXT NOT NULL,
	id_token NTEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS transactions(
	id nvarchar(50) NOT NULL PRIMARY KEY,
	amount nvarchar(255) NOT NULL,
	date nvarchar(255) NOT NULL,
	comment text NOT NULL,
	account_id nvarchar(255) NOT NULL,
	type_id integer(8) NOT NULL,
	created_at timestamp NOT NULL,
	synced_at timestamp NULL
);
`)
	return errors.Wrap(err, "Failed to setup storage")
}

// TODO: Add context (for others as well)
func (s *sqlStorage) GetAuthTokenByEmail(ctx context.Context, email string) (*AuthTokenDTO, error) {
	res, err := s.db.Query(`
	SELECT 
		email, id_token, refresh_token
	FROM USERS WHERE email = $1`, email)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	if !res.Next() {
		if res.Err() != nil {
			return nil, res.Err()
		}
		return nil, fmt.Errorf("Unknown user: %v", email)
	}

	result := &AuthTokenDTO{}
	if err := res.Scan(
		&result.Email,
		&result.IDToken,
		&result.RefreshToken,
	); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *sqlStorage) SaveAuthToken(ctx context.Context, token *AuthTokenDTO) error {
	if _, err := s.db.Exec(`
	INSERT INTO users(email, id_token, refresh_token)
	VALUES($1, $2, $3)
	ON CONFLICT(email) DO UPDATE 
	SET id_token=$2, refresh_token=$3
	`,
		token.Email, token.IDToken, token.RefreshToken); err != nil {
		return err
	}
	return nil
}

func (s *sqlStorage) SavePendingTransaction(ctx context.Context, trx *PendingTransactionDTO) error {
	if _, err := s.db.ExecContext(ctx, `
	INSERT INTO transactions(
		id,
		amount,
		date,
		comment,
		account_id,
		type_id,
		created_at,
		synced_at
	)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT(id) DO UPDATE 
	SET amount=$2, date=$3, comment=$4, account_id=$5, type_id=$6, synced_at=$8
	`,
		trx.ID, trx.Amount, trx.Date, trx.Comment,
		trx.AccountID, trx.TypeID, s.nowFn().UTC(), trx.SyncedAt); err != nil {
		return errors.Wrapf(err, "Failed to save transaction: %v, %v (%v)", trx.Amount, trx.Date, trx.Comment)
	}
	return nil
}

func (s *sqlStorage) FindNotSyncedTransactions(ctx context.Context, accountID string) ([]PendingTransactionDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
	SELECT 
		id, amount, date, comment, account_id, type_id, created_at, synced_at
	FROM transactions 
	WHERE account_id=$1 and synced_at IS NULL
	`, accountID)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to query not synced transactions")
	}

	defer rows.Close()

	trxs := []PendingTransactionDTO{}
	for rows.Next() {
		trx := &PendingTransactionDTO{}
		if err := rows.Scan(
			&trx.ID,
			&trx.Amount,
			&trx.Date,
			&trx.Comment,
			&trx.AccountID,
			&trx.TypeID,
			&trx.CreatedAt,
			&trx.SyncedAt,
		); err != nil {
			return nil, errors.Wrap(err, "Failed to scan trx")
		}
		trxs = append(trxs, *trx)
	}

	return trxs, nil
}

// SQLStorageOpt is an option of SQL storage
type SQLStorageOpt func(s *sqlStorage)

// WithSQLDb will set an explicit db instance for a storage
func WithSQLDb(db *sql.DB) SQLStorageOpt {
	return func(s *sqlStorage) {
		s.db = db
	}
}

// NewSQLStorage returns an instance of a local storage
func NewSQLStorage(opts ...SQLStorageOpt) (Storage, error) {
	storage := &sqlStorage{
		nowFn: defaultNowFn,
	}
	for _, opt := range opts {
		opt(storage)
	}
	return storage, nil
}
