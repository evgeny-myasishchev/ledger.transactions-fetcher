package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

// JSONMarshalToReader marshal JSON or panic. To be used for tests only
func JSONMarshalToReader(v interface{}) io.Reader {
	payload, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(payload)
}

// JSONUnmarshalBuffer unmarshal provided buffer. To be used for tests only
func JSONUnmarshalBuffer(buffer *bytes.Buffer, v interface{}) {
	if err := json.Unmarshal(buffer.Bytes(), v); err != nil {
		panic(err)
	}
}

// JSONUnmarshalReader unmarshal provided reader
// returns true if succeeded
func JSONUnmarshalReader(t *testing.T, reader io.Reader, v interface{}) bool {
	if !assert.NotNil(t, reader) {
		return false
	}
	buffer, err := ioutil.ReadAll(reader)
	if !assert.NoError(t, err) {
		return false
	}
	if err := json.Unmarshal(buffer, v); !assert.NoError(t, err) {
		return false
	}
	return true
}
