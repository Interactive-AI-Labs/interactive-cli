package cmd

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

// TestBindSkillConfigFlags pins the on-the-wire shape of the config payload
// against what interactive-chat's skill_loader reads at runtime
// (config = {"skill": {"description": "...", "intents": [...]}}).
func TestBindSkillConfigFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want map[string]any
	}{
		{
			name: "no flags omits config entirely",
			args: nil,
			want: nil,
		},
		{
			name: "description only",
			args: []string{"--description", "Summarize a trace"},
			want: map[string]any{
				"skill": map[string]any{
					"description": "Summarize a trace",
				},
			},
		},
		{
			name: "intents only",
			args: []string{"--intents", "summarize trace", "--intents", "explain trace"},
			want: map[string]any{
				"skill": map[string]any{
					"intents": []string{"summarize trace", "explain trace"},
				},
			},
		},
		{
			name: "both flags",
			args: []string{
				"--description", "Summarize a trace",
				"--intents", "summarize trace",
				"--intents", "explain trace",
			},
			want: map[string]any{
				"skill": map[string]any{
					"description": "Summarize a trace",
					"intents":     []string{"summarize trace", "explain trace"},
				},
			},
		},
		{
			// Guards against pflag.StringSliceVar regressions: a comma in a
			// single repeated value must stay one intent, not be split.
			name: "comma inside repeated intent is preserved",
			args: []string{"--intents", "summarize, then explain"},
			want: map[string]any{
				"skill": map[string]any{
					"intents": []string{"summarize, then explain"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "fake", RunE: func(*cobra.Command, []string) error { return nil }}
			build := bindSkillConfigFlags(cmd)
			cmd.SetArgs(tt.args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("flag parsing failed: %v", err)
			}
			got := build()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("config mismatch\ngot:  %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}
