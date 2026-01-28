package inputs

import (
	"testing"
)

func TestValidateSecretKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid key",
			key:     "API_KEY",
			wantErr: false,
		},
		{
			name:    "empty string",
			key:     "",
			wantErr: true,
			errMsg:  "key name cannot be empty",
		},
		{
			name:    "whitespace only",
			key:     "   ",
			wantErr: true,
			errMsg:  "key name cannot be empty",
		},
		{
			name:    "invalid key",
			key:     "123VAR",
			wantErr: true,
			errMsg:  `key name "123VAR" is not a valid environment variable name`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecretKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSecretKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ValidateSecretKey() error message = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateSecretValue(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid value",
			key:     "API_KEY",
			value:   "my-secret-value",
			wantErr: false,
		},
		{
			name:    "empty value",
			key:     "API_KEY",
			value:   "",
			wantErr: true,
			errMsg:  `value for key "API_KEY" cannot be empty`,
		},
		{
			name:    "whitespace only value",
			key:     "API_KEY",
			value:   "   ",
			wantErr: true,
			errMsg:  `value for key "API_KEY" cannot be empty`,
		},
		{
			name:    "value wrapped in quotes",
			key:     "MY_KEY",
			value:   `"myvalue"`,
			wantErr: true,
			errMsg:  `value for key "MY_KEY" should not be wrapped in double quotes`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecretValue(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSecretValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ValidateSecretValue() error message = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestIsValidEnvVarName(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "empty string",
			s:    "",
			want: false,
		},
		{
			name: "single uppercase letter",
			s:    "A",
			want: true,
		},
		{
			name: "single lowercase letter",
			s:    "z",
			want: true,
		},
		{
			name: "single underscore",
			s:    "_",
			want: true,
		},
		{
			name: "starts with digit",
			s:    "1VAR",
			want: false,
		},
		{
			name: "contains underscore",
			s:    "MY_VAR",
			want: true,
		},
		{
			name: "contains digit",
			s:    "VAR1",
			want: true,
		},
		{
			name: "contains hyphen",
			s:    "MY-VAR",
			want: false,
		},
		{
			name: "contains dot",
			s:    "my.var",
			want: false,
		},
		{
			name: "contains special char",
			s:    "VAR$NAME",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidEnvVarName(tt.s); got != tt.want {
				t.Errorf("isValidEnvVarName(%q) = %v, want %v", tt.s, got, tt.want)
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
			name:  "single character",
			s:     `"`,
			quote: '"',
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
