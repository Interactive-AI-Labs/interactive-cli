package cmd

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRouterCommandPaths(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "router", got: routerCmd.CommandPath(), want: "iai router"},
		{name: "router info", got: routerInfoCmd.CommandPath(), want: "iai router info"},
		{name: "models", got: modelsCmd.CommandPath(), want: "iai router models"},
		{name: "models list", got: modelsListCmd.CommandPath(), want: "iai router models list"},
		{name: "models get", got: modelsGetCmd.CommandPath(), want: "iai router models get"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, tt.got); diff != "" {
				t.Errorf("command path mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRouterUsesDefaultParentHelp(t *testing.T) {
	got := struct {
		Runnable       bool
		HasSubcommands bool
	}{
		Runnable:       routerCmd.Runnable(),
		HasSubcommands: routerCmd.HasSubCommands(),
	}
	want := struct {
		Runnable       bool
		HasSubcommands bool
	}{HasSubcommands: true}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parent behavior mismatch (-want +got):\n%s", diff)
	}
}
