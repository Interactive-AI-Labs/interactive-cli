package output

import (
	"bytes"
	"testing"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintReplicaList(t *testing.T) {
	tests := []struct {
		name     string
		replicas []clients.ReplicaInfo
		want     string
	}{
		{
			name:     "empty list prints headers only",
			replicas: []clients.ReplicaInfo{},
			want:     "NAME   STATUS   CPU   MEMORY   STARTED\n",
		},
		{
			name: "ready replica with status",
			replicas: []clients.ReplicaInfo{
				{
					Name:      "web-abc123",
					Status:    "Running",
					Phase:     "Running",
					Ready:     true,
					CPU:       "100m",
					Memory:    "128Mi",
					StartTime: "2024-01-01T00:00:00Z",
				},
			},
			want: "NAME         STATUS            CPU    MEMORY   STARTED\n" +
				"web-abc123   Running [Ready]   100m   128Mi    2024-01-01T00:00:00Z\n",
		},
		{
			name: "not-ready replica shows Not Ready",
			replicas: []clients.ReplicaInfo{
				{
					Name:   "worker-xyz",
					Status: "CrashLoopBackOff",
					Ready:  false,
					CPU:    "50m",
					Memory: "64Mi",
				},
			},
			want: "NAME         STATUS                         CPU   MEMORY   STARTED\n" +
				"worker-xyz   CrashLoopBackOff [Not Ready]   50m   64Mi     \n",
		},
		{
			name: "falls back to phase when status is empty",
			replicas: []clients.ReplicaInfo{
				{
					Name:  "api-pod",
					Phase: "Pending",
					Ready: false,
				},
			},
			want: "NAME      STATUS                CPU   MEMORY   STARTED\n" +
				"api-pod   Pending [Not Ready]                  \n",
		},
		{
			name: "shows Unknown when both status and phase are empty",
			replicas: []clients.ReplicaInfo{
				{
					Name:  "mystery-pod",
					Ready: false,
				},
			},
			want: "NAME          STATUS                CPU   MEMORY   STARTED\n" +
				"mystery-pod   Unknown [Not Ready]                  \n",
		},
		{
			name: "multiple replicas",
			replicas: []clients.ReplicaInfo{
				{Name: "pod-1", Status: "Running", Ready: true, CPU: "100m", Memory: "128Mi"},
				{Name: "pod-2", Status: "Pending", Ready: false, CPU: "0m", Memory: "0Mi"},
			},
			want: "NAME    STATUS                CPU    MEMORY   STARTED\n" +
				"pod-1   Running [Ready]       100m   128Mi    \n" +
				"pod-2   Pending [Not Ready]   0m     0Mi      \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintReplicaList(&buf, tt.replicas)
			if err != nil {
				t.Fatalf("PrintReplicaList() error = %v, want nil", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
