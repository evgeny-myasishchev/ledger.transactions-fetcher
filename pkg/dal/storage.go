package dal

import "context"

// AuthTokenDTO is a DTO to store user auth token
type AuthTokenDTO struct {
	Email        string
	IDToken      string
	RefreshToken string
}

// Storage is a persistance layer
type Storage interface {
	Setup(ctx context.Context) error
	GetAuthTokenByEmail(ctx context.Context, email string) (*AuthTokenDTO, error)
	SaveAuthToken(ctx context.Context, token *AuthTokenDTO) error
}
