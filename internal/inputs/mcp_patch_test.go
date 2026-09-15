package inputs

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/google/go-cmp/cmp"
)

func TestBuildMcpUpdatePatch(t *testing.T) {
	credential, empty := "rotated", ""
	for _, tt := range []struct {
		name          string
		input         McpUpdateInput
		flags         []string
		want, wantErr string
	}{
		{name: "no fields", wantErr: "no fields to update; pass at least one flag"},
		{name: "memory stays partial", input: McpUpdateInput{Workload: platform.McpWorkload{Memory: "1G", Endpoint: true}, Auth: platform.McpAuth{Type: "none"}}, flags: []string{"memory"}, want: `{"workload":{"memory":"1G"}}`},
		{name: "invalid values reach the API", flags: []string{"memory", "port"}, want: `{"workload":{"memory":"","port":0}}`},
		{name: "public enabled", input: McpUpdateInput{Workload: platform.McpWorkload{Endpoint: true}}, flags: []string{"endpoint"}, want: `{"workload":{"endpoint":true}}`},
		{name: "public disabled", flags: []string{"endpoint"}, want: `{"workload":{"endpoint":false}}`},
		{name: "image", input: McpUpdateInput{Workload: platform.McpWorkload{Image: "tools:2"}}, flags: []string{"image-name", "image-tag"}, want: `{"workload":{"image":"tools:2"}}`},
		{name: "assign stack", input: McpUpdateInput{Workload: platform.McpWorkload{StackId: "tools"}}, flags: []string{"stack-id"}, want: `{"workload":{"stack_id":"tools"}}`},
		{name: "clear stack", input: McpUpdateInput{ClearStackID: true}, want: `{"workload":{"stack_id":null}}`},
		{name: "empty stack", flags: []string{"stack-id"}, want: `{"workload":{"stack_id":""}}`},
		{name: "conflicting stack flags", input: McpUpdateInput{ClearStackID: true}, flags: []string{"stack-id"}, wantErr: "--clear-stack-id cannot be combined with --stack-id"},
		{name: "env replacement", input: McpUpdateInput{EnvVars: []string{"ENV=dev", "EMPTY="}}, flags: []string{"env"}, want: `{"workload":{"env":[{"name":"ENV","value":"dev"},{"name":"EMPTY","value":""}]}}`},
		{name: "secret replacement", input: McpUpdateInput{Workload: platform.McpWorkload{SecretRefs: []string{"first", "second"}}}, flags: []string{"secret"}, want: `{"workload":{"secret_refs":["first","second"]}}`},
		{name: "clear env", input: McpUpdateInput{ClearEnv: true}, want: `{"workload":{"env":[]}}`},
		{name: "clear secrets", input: McpUpdateInput{ClearSecret: true}, want: `{"workload":{"secret_refs":[]}}`},
		{name: "credential only", input: McpUpdateInput{Auth: platform.McpAuth{Credential: &credential}}, flags: []string{"credential"}, want: `{"auth":{"credential":"rotated"}}`},
		{name: "credential from stdin", input: McpUpdateInput{Auth: platform.McpAuth{Credential: &credential}}, flags: []string{"credential-stdin"}, want: `{"auth":{"credential":"rotated"}}`},
		{name: "clear authentication", input: McpUpdateInput{Auth: platform.McpAuth{Type: "none", Credential: &empty}}, flags: []string{"auth-type", "credential"}, want: `{"auth":{"credential":"","type":"none"}}`},
		{name: "empty prefix", input: McpUpdateInput{Auth: platform.McpAuth{HeaderPrefix: &empty}}, flags: []string{"auth-header-prefix"}, want: `{"auth":{"header_prefix":""}}`},
		{name: "description only", input: McpUpdateInput{Description: "notes"}, flags: []string{"description"}, want: `{"description":"notes"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			patch, err := BuildMcpUpdatePatch(
				tt.input,
				func(flag string) bool { return slices.Contains(tt.flags, flag) },
			)
			message, got := "", ""
			if err != nil {
				message = err.Error()
			} else {
				encoded, err := json.Marshal(patch)
				if err != nil {
					t.Fatal(err)
				}
				got = string(encoded)
			}
			if diff := cmp.Diff([]string{tt.want, tt.wantErr}, []string{got, message}); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
