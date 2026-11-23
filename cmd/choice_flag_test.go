package cmd

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestChoiceFlags(t *testing.T) {
	tests := []struct {
		path []string
		flag string
		want []string
	}{
		{
			[]string{"router", "models", "list"},
			"input-modality",
			[]string{"audio", "file", "image", "text", "video"},
		},
		{
			[]string{"router", "models", "list"},
			"output-modality",
			[]string{
				"audio", "decisions", "embeddings", "image", "rerank",
				"speech", "text", "transcription", "video",
			},
		},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.path, " ")+" --"+tt.flag, func(t *testing.T) {
			cmd, _, err := rootCmd.Find(tt.path)
			if err != nil {
				t.Fatalf("rootCmd.Find() error = %v", err)
			}
			f := cmd.Flags().Lookup(tt.flag)
			if f == nil {
				t.Fatalf("flag --%s not found", tt.flag)
			}
			suffix := " (" + strings.Join(tt.want, "|") + ")"
			if !strings.HasSuffix(f.Usage, suffix) {
				t.Errorf("usage = %q, want suffix %q", f.Usage, suffix)
			}
			complete, ok := cmd.GetFlagCompletionFunc(tt.flag)
			if !ok {
				t.Fatalf("no completion registered for --%s", tt.flag)
			}
			got, directive := complete(cmd, nil, "")
			if !slices.Equal(got, tt.want) {
				t.Errorf("completions = %v, want %v", got, tt.want)
			}
			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("directive = %v, want NoFileComp", directive)
			}
		})
	}
}

var choiceSuffix = regexp.MustCompile(`\(([^()|\s]+(?:\|[^()|\s]+)+)\)$`)

// Help that lists choices must also complete exactly those choices.
func TestChoiceFlagsComplete(t *testing.T) {
	found := 0
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		c.LocalFlags().VisitAll(func(f *pflag.Flag) {
			m := choiceSuffix.FindStringSubmatch(f.Usage)
			if m == nil {
				return
			}
			found++
			where := c.CommandPath() + " --" + f.Name
			complete, ok := c.GetFlagCompletionFunc(f.Name)
			if !ok {
				t.Errorf("%s: help lists choices but nothing completes them", where)
				return
			}
			got, directive := complete(c, nil, "")
			want := strings.Split(m[1], "|")
			if !slices.Equal(got, want) || directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf(
					"%s: completes %v (%v), want %v without files",
					where,
					got,
					directive,
					want,
				)
			}
		})
		for _, child := range c.Commands() {
			walk(child)
		}
	}
	walk(rootCmd)
	if found == 0 {
		t.Fatal("no flag lists choices; the help format changed")
	}
}
