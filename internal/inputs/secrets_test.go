package inputs

import (
	"testing"
)

func TestValidateSecretValue(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "valid unquoted value",
			key:     "MY_KEY",
			value:   "myvalue",
			wantErr: false,
		},
		{
			name:    "valid value with internal quotes",
			key:     "MY_KEY",
			value:   `say "hello"`,
			wantErr: false,
		},
		{
			name:    "valid empty value",
			key:     "MY_KEY",
			value:   "",
			wantErr: false,
		},
		{
			name:    "valid single character",
			key:     "MY_KEY",
			value:   "x",
			wantErr: false,
		},
		{
			name:    "invalid double quoted value",
			key:     "MY_KEY",
			value:   `"myvalue"`,
			wantErr: true,
		},
		{
			name:    "invalid single quoted value",
			key:     "MY_KEY",
			value:   `'myvalue'`,
			wantErr: true,
		},
		{
			name:    "valid mixed quotes",
			key:     "MY_KEY",
			value:   `"myvalue'`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecretValue(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSecretValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSecretData(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name: "all valid values",
			data: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			wantErr: false,
		},
		{
			name:    "empty map",
			data:    map[string]string{},
			wantErr: false,
		},
		{
			name: "one invalid double quoted value",
			data: map[string]string{
				"KEY1": "value1",
				"KEY2": `"quoted"`,
			},
			wantErr: true,
		},
		{
			name: "one invalid single quoted value",
			data: map[string]string{
				"KEY1": `'quoted'`,
				"KEY2": "value2",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecretData(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSecretData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsQuoted(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		quote byte
		want  bool
	}{
		{
			name:  "double quoted",
			s:     `"hello"`,
			quote: '"',
			want:  true,
		},
		{
			name:  "single quoted",
			s:     `'hello'`,
			quote: '\'',
			want:  true,
		},
		{
			name:  "not quoted",
			s:     "hello",
			quote: '"',
			want:  false,
		},
		{
			name:  "only start quote",
			s:     `"hello`,
			quote: '"',
			want:  false,
		},
		{
			name:  "empty string",
			s:     "",
			quote: '"',
			want:  false,
		},
		{
			name:  "wrong quote type",
			s:     `"hello"`,
			quote: '\'',
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isQuoted(tt.s, tt.quote); got != tt.want {
				t.Errorf("isQuoted() = %v, want %v", got, tt.want)
			}
		})
	}
}
