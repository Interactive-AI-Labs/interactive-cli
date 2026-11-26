package cmd

import (
	"encoding/json"
	"testing"
)

func TestPerformanceAgentsOrSelected(t *testing.T) {
	tests := []struct {
		name     string
		selected string
		flag     []string
		want     string
	}{
		{
			"typed agents win",
			"sel",
			[]string{"a", "b"},
			`{"agents":["a","b"],"filters":[{"column":"Tags","type":"arrayOptions","operator":"any of","value":["agent:a","agent:b"]}]}`,
		},
		{
			"selected agent used",
			"sel",
			nil,
			`{"agents":["sel"],"filters":[{"column":"Tags","type":"arrayOptions","operator":"any of","value":["agent:sel"]}]}`,
		},
		{"neither means every agent", "", nil, `{"agents":null,"filters":null}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saveSelection(t, tt.selected)
			agents, err := performanceAgentsOrSelected(tt.flag, "", "")
			if err != nil {
				t.Fatal(err)
			}
			got, _ := json.Marshal(
				map[string]any{"agents": agents, "filters": agentTagFilters(agents)},
			)
			if string(got) != tt.want {
				t.Errorf("got %s\nwant %s", got, tt.want)
			}
		})
	}
}
