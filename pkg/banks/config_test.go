package banks

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/bxcodec/faker/v3"

	"github.com/stretchr/testify/assert"
)

func ensureTmpDir(name string) string {
	var tmpDir = path.Join("..", "..", "tmp", name)
	os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		panic(err)
	}
	return tmpDir
}

func TestFSFetcherConfig_GetUserConfig(t *testing.T) {
	type fields struct {
		configDir string
	}
	type args struct {
		userID   string
		receiver interface{}
	}
	type testCase struct {
		name   string
		fields fields
		args   args
		run    func(t *testing.T, err error)
	}

	type userConfig struct {
		Prop1 string
		Prop2 string
	}

	configDir := ensureTmpDir("bank-config-test")
	writeConfig := func(name string, value interface{}) bool {
		buffer, err := json.Marshal(value)
		if !assert.NoError(t, err) {
			return false
		}

		if err := ioutil.WriteFile(path.Join(configDir, name), buffer, os.ModePerm); !assert.NoError(t, err) {
			return false
		}

		return true
	}

	tests := []func() testCase{
		func() testCase {
			userID := faker.Word()
			want := userConfig{
				Prop1: faker.Word(),
				Prop2: faker.Word(),
			}
			if !writeConfig(userID+".json", want) {
				panic("Failed to write config")
			}
			receiver := userConfig{}
			return testCase{
				name:   "fetch existing config",
				fields: fields{configDir: configDir},
				args:   args{userID: userID, receiver: &receiver},
				run: func(t *testing.T, err error) {
					if !assert.Nil(t, err) {
						return
					}
					assert.Equal(t, want, receiver)
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewFSFetcherConfig(tt.fields.configDir)
			err := cfg.GetUserConfig(context.Background(), tt.args.userID, tt.args.receiver)
			tt.run(t, err)
		})
	}
}
