package auth

import (
	"context"
	"testing"
)

func Test_service_RegisterUser(t *testing.T) {
	type fields struct {
		oauthClient OAuthClient
	}
	type args struct {
		ctx       context.Context
		oauthCode string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				oauthClient: tt.fields.oauthClient,
			}
			if err := svc.RegisterUser(tt.args.ctx, tt.args.oauthCode); (err != nil) != tt.wantErr {
				t.Errorf("service.RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
