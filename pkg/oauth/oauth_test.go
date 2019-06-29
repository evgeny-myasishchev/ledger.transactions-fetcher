package oauth

import (
	"fmt"
	"testing"

	"github.com/bxcodec/faker/v3"

	"github.com/stretchr/testify/assert"
)

func Test_googleOAuthClient_BuildCodeGrantURL(t *testing.T) {
	type fields struct {
		clientID string
	}
	clientID := "client-id-" + faker.Word()
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "build code flow url",
			fields: fields{clientID: clientID},
			want: fmt.Sprint(
				"https://accounts.google.com/o/oauth2/v2/auth?",
				"response_type=code",
				"&client_id="+clientID,
				"&redirect_uri=urn:ietf:wg:oauth:2.0:oob",
				"&scope=email",
				"&access_type=offline",
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewGoogleOAuth(WithClientSecrets(tt.fields.clientID, ""))
			got := c.BuildCodeGrantURL()
			assert.Equal(t, tt.want, got)
		})
	}
}
