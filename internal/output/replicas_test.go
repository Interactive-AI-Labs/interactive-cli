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

func TestPrintReplicaDescribe(t *testing.T) {
	tests := []struct {
		name   string
		status *clients.ReplicaStatus
		want   string
	}{
		{
			name: "no last termination state",
			status: &clients.ReplicaStatus{
				Name:         "web-abc123",
				Status:       "Running",
				Ready:        true,
				StartTime:    "2024-01-01T00:00:00Z",
				RestartCount: 0,
			},
			want: "\nName:          web-abc123\n" +
				"Status:        Running\n" +
				"Ready:         Yes\n" +
				"Start Time:    2024-01-01T00:00:00Z\n" +
				"Restart Count: 0\n",
		},
		{
			name: "with last termination state all fields",
			status: &clients.ReplicaStatus{
				Name:         "worker-xyz",
				Status:       "Running",
				Ready:        true,
				RestartCount: 3,
				LastTerminationState: &clients.ReplicaLastTermination{
					Reason:     "OOMKilled",
					ExitCode:   137,
					StartedAt:  "2024-01-01T00:00:00Z",
					FinishedAt: "2024-01-01T01:00:00Z",
				},
			},
			want: "\nName:          worker-xyz\n" +
				"Status:        Running\n" +
				"Ready:         Yes\n" +
				"Restart Count: 3\n" +
				"\nLast Termination State:\n" +
				"  Reason:      OOMKilled\n" +
				"  Exit Code:   137\n" +
				"  Started At:  2024-01-01T00:00:00Z\n" +
				"  Finished At: 2024-01-01T01:00:00Z\n",
		},
		{
			name: "last termination state without timestamps",
			status: &clients.ReplicaStatus{
				Name:         "api-pod",
				Status:       "CrashLoopBackOff",
				Ready:        false,
				RestartCount: 5,
				LastTerminationState: &clients.ReplicaLastTermination{
					Reason:   "Error",
					ExitCode: 1,
				},
			},
			want: "\nName:          api-pod\n" +
				"Status:        CrashLoopBackOff\n" +
				"Ready:         No\n" +
				"Restart Count: 5\n" +
				"\nLast Termination State:\n" +
				"  Reason:      Error\n" +
				"  Exit Code:   1\n",
		},
		{
			name: "last termination state with exit code zero",
			status: &clients.ReplicaStatus{
				Name:         "job-pod",
				Status:       "Running",
				Ready:        true,
				RestartCount: 1,
				LastTerminationState: &clients.ReplicaLastTermination{
					Reason:   "Completed",
					ExitCode: 0,
				},
			},
			want: "\nName:          job-pod\n" +
				"Status:        Running\n" +
				"Ready:         Yes\n" +
				"Restart Count: 1\n" +
				"\nLast Termination State:\n" +
				"  Reason:      Completed\n" +
				"  Exit Code:   0\n",
		},
		{
			name: "last termination state with resources and events",
			status: &clients.ReplicaStatus{
				Name:         "full-pod",
				Status:       "Running",
				Ready:        true,
				StartTime:    "2024-06-01T12:00:00Z",
				RestartCount: 1,
				LastTerminationState: &clients.ReplicaLastTermination{
					Reason:     "OOMKilled",
					ExitCode:   137,
					FinishedAt: "2024-06-01T11:59:00Z",
				},
				Resources: &clients.ReplicaResources{
					CPU:    "250m",
					Memory: "512Mi",
				},
				Events: []clients.ReplicaEvent{
					{Type: "Warning", Reason: "OOMKilling", Count: 1, Message: "Memory limit exceeded", LastTimestamp: "2024-06-01T11:59:00Z"},
				},
			},
			want: "\nName:          full-pod\n" +
				"Status:        Running\n" +
				"Ready:         Yes\n" +
				"Start Time:    2024-06-01T12:00:00Z\n" +
				"Restart Count: 1\n" +
				"\nLast Termination State:\n" +
				"  Reason:      OOMKilled\n" +
				"  Exit Code:   137\n" +
				"  Finished At: 2024-06-01T11:59:00Z\n" +
				"\nResources:\n" +
				"  CPU:    250m\n" +
				"  Memory: 512Mi\n" +
				"\nEvents:\n" +
				"TYPE      REASON       COUNT   MESSAGE                 LAST SEEN\n" +
				"Warning   OOMKilling   1       Memory limit exceeded   2024-06-01T11:59:00Z\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintReplicaDescribe(&buf, tt.status)
			if err != nil {
				t.Fatalf("PrintReplicaDescribe() error = %v, want nil", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
