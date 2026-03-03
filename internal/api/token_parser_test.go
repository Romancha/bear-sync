package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConsumerTokens(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    map[string]string
		wantErr string
	}{
		{
			name: "single consumer",
			raw:  "openclaw:token1",
			want: map[string]string{"openclaw": "token1"},
		},
		{
			name: "multiple consumers",
			raw:  "openclaw:token1,myapp:token2",
			want: map[string]string{"openclaw": "token1", "myapp": "token2"},
		},
		{
			name: "three consumers",
			raw:  "openclaw:secret1,myapp:secret2,another:secret3",
			want: map[string]string{"openclaw": "secret1", "myapp": "secret2", "another": "secret3"},
		},
		{
			name: "whitespace trimming",
			raw:  " openclaw : token1 , myapp : token2 ",
			want: map[string]string{"openclaw": "token1", "myapp": "token2"},
		},
		{
			name:    "empty string",
			raw:     "",
			wantErr: "no consumer tokens configured",
		},
		{
			name:    "whitespace only",
			raw:     "   ",
			wantErr: "no consumer tokens configured",
		},
		{
			name:    "missing colon",
			raw:     "openclaw-token1",
			wantErr: "invalid consumer token entry \"openclaw-token1\": expected format \"name:token\"",
		},
		{
			name:    "missing colon in second entry",
			raw:     "openclaw:token1,badentry",
			wantErr: "invalid consumer token entry \"badentry\": expected format \"name:token\"",
		},
		{
			name:    "empty name",
			raw:     ":token1",
			wantErr: "invalid consumer token entry \":token1\": name and token must not be empty",
		},
		{
			name:    "empty token",
			raw:     "openclaw:",
			wantErr: "invalid consumer token entry \"openclaw:\": name and token must not be empty",
		},
		{
			name:    "duplicate consumer name",
			raw:     "openclaw:token1,openclaw:token2",
			wantErr: "duplicate consumer name \"openclaw\"",
		},
		{
			name: "token containing colon",
			raw:  "openclaw:tok:en:1",
			want: map[string]string{"openclaw": "tok:en:1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConsumerTokens(tt.raw)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
