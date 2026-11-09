package cmd

import "testing"

func TestRoutineFollowAuthHint(t *testing.T) {
	const hint = "Authentication: reuse the same --api-key argument when following; it is omitted above.\n"
	tests := []struct {
		name, apiKey, defaultAPIKey, token, want string
	}{
		{"login cookies", "", "", "", ""},
		{"environment key", "environment-secret", "environment-secret", "", ""},
		{"flag key", "flag-secret", "", "", hint},
		{"override environment key", "flag-secret", "environment-secret", "", hint},
		{"clear environment key", "", "environment-secret", "", hint},
		{"bearer token takes precedence", "flag-secret", "environment-secret", "token-secret", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := routineFollowAuthHint(tt.apiKey, tt.defaultAPIKey, tt.token); got != tt.want {
				t.Errorf("hint = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHostnameOverride(t *testing.T) {
	tests := []struct {
		name, value, defaultValue, want string
	}{
		{"platform default", "https://app.interactive.ai", "https://app.interactive.ai", ""},
		{
			"operator default",
			"https://deployment.interactive.ai",
			"https://deployment.interactive.ai",
			"",
		},
		{"dev environment default", "https://dev.interactive.ai", "https://dev.interactive.ai", ""},
		{"local environment default", "http://localhost:8000", "http://localhost:8000", ""},
		{"schemeless environment default", "https://dev.interactive.ai", "dev.interactive.ai", ""},
		{
			"local override",
			"http://localhost:8000",
			"https://deployment.dev.interactive.ai",
			"http://localhost:8000",
		},
		{
			"platform override",
			"https://dev.interactive.ai",
			"https://app.interactive.ai",
			"https://dev.interactive.ai",
		},
		{
			"HTTP override of schemeless HTTPS default",
			"http://dev.interactive.ai",
			"dev.interactive.ai",
			"http://dev.interactive.ai",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hostnameOverride(tt.value, tt.defaultValue); got != tt.want {
				t.Errorf("override = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAgentRoutineRunArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"agent and folder routine", []string{"my-agent", "billing/refund"}, ""},
		{"no args", nil, "accepts 2 arg(s), received 0"},
		{"missing routine", []string{"my-agent"}, "accepts 2 arg(s), received 1"},
		{
			"extra input argument",
			[]string{"my-agent", "refund", "input.json"},
			"accepts 2 arg(s), received 3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := agentRoutineRunCmd.Args(agentRoutineRunCmd, tt.args)
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Errorf("error = %q, want %q", gotErr, tt.wantErr)
			}
		})
	}
}

func TestAgentRoutineCommandRegistration(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"plural", []string{"agents", "routines", "run"}, "iai agents routines run"},
		{"singular aliases", []string{"agent", "routine", "run"}, "iai agents routines run"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _, err := rootCmd.Find(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if got := cmd.CommandPath(); got != tt.want {
				t.Errorf("command = %q, want %q", got, tt.want)
			}
		})
	}
}
