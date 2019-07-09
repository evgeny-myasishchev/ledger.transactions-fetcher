package types

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// IDToken is a jwt encoded ID token string
type IDToken string

// Value is an underlying string
func (t IDToken) Value() string {
	return string(t)
}

// IDTokenDetails represents details related to ID token
type IDTokenDetails struct {
	Email   string `json:"email"`
	Expires int64  `json:"exp"`
}

// ExtractIDTokenDetails will decode an ID token and get it's details
func (t IDToken) ExtractIDTokenDetails() (*IDTokenDetails, error) {
	parts := strings.Split(t.Value(), ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("Unexpected ID token structure. Should have 3 segments, got: %v", len(parts))
	}
	decoder := json.NewDecoder(base64.NewDecoder(base64.RawURLEncoding, strings.NewReader(parts[1])))
	var details IDTokenDetails
	if err := decoder.Decode(&details); err != nil {
		return nil, err
	}
	return &details, nil
}
