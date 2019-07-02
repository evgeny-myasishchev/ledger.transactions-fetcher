package storage

import (
	"database/sql"
	"fmt"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/oauth"

	// This has to be here to let go mods work work
	_ "github.com/mattn/go-sqlite3"
)

type sqlStorage struct {
	db *sql.DB
}

func (s *sqlStorage) Setup() error {
	_, err := s.db.Exec(`
CREATE TABLE users(
	email nvarchar(30) NOT NULL PRIMARY KEY,
	access_token NTEXT NOT NULL,
	refresh_token NTEXT NOT NULL,
	id_token NTEXT NOT NULL,
	expires_in int NOT NULL
)
`)
	return err
}

// TODO: Add context (for others as well)
func (s *sqlStorage) GetAccessTokenByEmail(userEmail string) (*oauth.AccessToken, error) {
	res, err := s.db.Query(`
	SELECT 
		access_token, refresh_token, id_token, expires_in
	FROM USERS WHERE email = $1`, userEmail)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	if !res.Next() {
		if res.Err() != nil {
			return nil, res.Err()
		}
		return nil, fmt.Errorf("Unknown user: %v", userEmail)
	}

	result := &oauth.AccessToken{}
	if err := res.Scan(
		&result.AccessToken,
		&result.RefreshToken,
		&result.IDToken,
		&result.ExpiresIn,
	); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *sqlStorage) SaveAccessToken(token *oauth.AccessToken) error {
	idToken, err := token.ExtractIDTokenDetails()
	if err != nil {
		return err
	}
	if _, err := s.db.Exec(`
	INSERT INTO users(email, access_token, refresh_token, id_token, expires_in)
	VALUES($1, $2, $3, $4, $5)
	ON CONFLICT(email) DO UPDATE 
	SET access_token=$2, refresh_token=$3, id_token=$4, expires_in=$5
	`,
		idToken.Email, token.AccessToken, token.RefreshToken,
		token.IDToken, token.ExpiresIn); err != nil {
		return err
	}
	return nil
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
	storage := &sqlStorage{}
	for _, opt := range opts {
		opt(storage)
	}
	return storage, nil
}
