package testing

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// EncodeUnsignedJWT will encode given payload as JWT token string
// Without any signature, empty header and footer
func EncodeUnsignedJWT(t *testing.T, payload interface{}) (string, error) {
	tokenData := bytes.Buffer{}
	if _, err := tokenData.WriteString(base64.StdEncoding.EncodeToString([]byte(`{"alg":"none"}`))); !assert.NoError(t, err) {
		return "", err
	}
	if _, err := tokenData.WriteString("."); !assert.NoError(t, err) {
		return "", err
	}
	marshaledPayload, err := json.Marshal(&payload)
	if !assert.NoError(t, err) {
		return "", err
	}
	if _, err := tokenData.WriteString(base64.StdEncoding.EncodeToString(marshaledPayload)); !assert.NoError(t, err) {
		return "", err
	}
	if _, err := tokenData.WriteString("."); !assert.NoError(t, err) {
		return "", err
	}
	if _, err := tokenData.WriteString(base64.StdEncoding.EncodeToString([]byte(`{}`))); !assert.NoError(t, err) {
		return "", err
	}
	return tokenData.String(), nil
}
