package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_param_emptyValue(t *testing.T) {
	type args struct {
		p param
	}
	tests := []struct {
		name string
		args args
		want paramValue
	}{
		{
			name: "str value",
			args: args{p: StringParam{}},
			want: StringVal{val: new(string)},
		},
		{
			name: "int value",
			args: args{p: IntParam{}},
			want: IntVal{val: new(int)},
		},
		{
			name: "bool value",
			args: args{p: BoolParam{}},
			want: BoolVal{val: new(bool)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.p.emptyValue()
			assert.Equal(t, tt.want, got)
		})
	}
}
