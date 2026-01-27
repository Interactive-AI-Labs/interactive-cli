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
		errMsg  string
	}{
		{
			name:    "valid value",
			key:     "API_KEY",
			value:   "my-secret-value",
			wantErr: false,
		},
		{
			name:    "valid value with internal quotes",
			key:     "MY_KEY",
			value:   `say "hello"`,
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
			name:    "value wrapped in double quotes",
			key:     "MY_KEY",
			value:   `"myvalue"`,
			wantErr: true,
			errMsg:  `value for key "MY_KEY" should not be wrapped in double quotes`,
		},
		{
			name:    "value wrapped in single quotes",
			key:     "MY_KEY",
			value:   `'myvalue'`,
			wantErr: true,
			errMsg:  `value for key "MY_KEY" should not be wrapped in single quotes`,
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

func TestIsValidEnvVarName(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{
			name:     "mixed case with numbers and underscores",
			varName:  "_My_Var_123",
			expected: true,
		},
		{
			name:     "uppercase with underscores",
			varName:  "API_KEY",
			expected: true,
		},
		{
			name:     "lowercase only",
			varName:  "myvar",
			expected: true,
		},
		{
			name:     "single underscore",
			varName:  "_",
			expected: true,
		},
		{
			name:     "single letter",
			varName:  "x",
			expected: true,
		},
		{
			name:     "empty string",
			varName:  "",
			expected: false,
		},
		{
			name:     "starts with number",
			varName:  "123VAR",
			expected: false,
		},
		{
			name:     "contains hyphen",
			varName:  "MY-VAR",
			expected: false,
		},
		{
			name:     "contains space",
			varName:  "MY VAR",
			expected: false,
		},
		{
			name:     "contains special char",
			varName:  "MY$VAR",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidEnvVarName(tt.varName)
			if result != tt.expected {
				t.Errorf("IsValidEnvVarName(%q) = %v, want %v", tt.varName, result, tt.expected)
			}
		})
	}
}

