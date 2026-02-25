package output

import (
	"bytes"
	"testing"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintVectorStoreList(t *testing.T) {
	tests := []struct {
		name   string
		stores []clients.VectorStoreInfo
		want   string
	}{
		{
			name:   "empty list prints message",
			stores: []clients.VectorStoreInfo{},
			want:   "No vector stores found.\n",
		},
		{
			name:   "nil list prints message",
			stores: nil,
			want:   "No vector stores found.\n",
		},
		{
			name: "single store",
			stores: []clients.VectorStoreInfo{
				{
					VectorStoreName: "my-store",
					Status:          "Running",
					SecretName:      "my-store-secret",
				},
			},
			want: "NAME       STATUS    SECRET\n" +
				"my-store   Running   my-store-secret\n",
		},
		{
			name: "multiple stores",
			stores: []clients.VectorStoreInfo{
				{
					VectorStoreName: "store-a",
					Status:          "Running",
					SecretName:      "secret-a",
				},
				{
					VectorStoreName: "store-b",
					Status:          "Pending",
					SecretName:      "secret-b",
				},
			},
			want: "NAME      STATUS    SECRET\n" +
				"store-a   Running   secret-a\n" +
				"store-b   Pending   secret-b\n",
		},
		{
			name: "store with empty secret",
			stores: []clients.VectorStoreInfo{
				{
					VectorStoreName: "no-secret",
					Status:          "Running",
					SecretName:      "",
				},
			},
			want: "NAME        STATUS    SECRET\n" +
				"no-secret   Running   \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintVectorStoreList(&buf, tt.stores)
			if err != nil {
				t.Fatalf("PrintVectorStoreList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintVectorStoreDescribe(t *testing.T) {
	tests := []struct {
		name  string
		store *clients.DescribeVectorStoreResponse
		want  string
	}{
		{
			name: "full describe output",
			store: &clients.DescribeVectorStoreResponse{
				VectorStoreName: "my-store",
				Status:          "ready",
				EngineVersion:   "POSTGRES_15",
				CreatedAt:       "2025-01-15T10:30:00Z",
				Resources: clients.VectorStoreResources{
					CPU:    4,
					Memory: 16,
				},
				Storage: clients.VectorStoreStorage{
					Size:            100,
					AutoResize:      true,
					AutoResizeLimit: 500,
				},
				HA: true,
				Backups: clients.VectorStoreBackupConfig{
					Enabled:   true,
					StartTime: "03:00",
				},
				SecretName: "my-store-credentials",
			},
			want: "Name:            my-store\n" +
				"Status:          ready\n" +
				"Engine Version:  POSTGRES_15\n" +
				"Created At:      2025-01-15T10:30:00Z\n" +
				"HA:              Yes\n" +
				"Backups:         Yes\n" +
				"Backup Time:     03:00\n" +
				"Secret:          my-store-credentials\n" +
				"\n" +
				"Resources:\n" +
				"  CPU:               4\n" +
				"  Memory:            16.00 GB\n" +
				"Storage:\n" +
				"  Size:              100 GB\n" +
				"  Auto Resize:       Yes\n" +
				"  Auto Resize Limit: 500 GB\n",
		},
		{
			name: "minimal store without optional fields",
			store: &clients.DescribeVectorStoreResponse{
				VectorStoreName: "basic-store",
				Status:          "creating",
				EngineVersion:   "POSTGRES_14",
				Resources: clients.VectorStoreResources{
					CPU:    2,
					Memory: 8,
				},
				Storage: clients.VectorStoreStorage{
					Size: 20,
				},
			},
			want: "Name:            basic-store\n" +
				"Status:          creating\n" +
				"Engine Version:  POSTGRES_14\n" +
				"HA:              No\n" +
				"Backups:         No\n" +
				"\n" +
				"Resources:\n" +
				"  CPU:               2\n" +
				"  Memory:            8.00 GB\n" +
				"Storage:\n" +
				"  Size:              20 GB\n" +
				"  Auto Resize:       No\n",
		},
		{
			name: "store with auto resize disabled shows no limit",
			store: &clients.DescribeVectorStoreResponse{
				VectorStoreName: "no-resize",
				Status:          "ready",
				EngineVersion:   "POSTGRES_15",
				CreatedAt:       "2025-02-01T00:00:00Z",
				Resources: clients.VectorStoreResources{
					CPU:    2,
					Memory: 4,
				},
				Storage: clients.VectorStoreStorage{
					Size:       50,
					AutoResize: false,
				},
				SecretName: "no-resize-credentials",
			},
			want: "Name:            no-resize\n" +
				"Status:          ready\n" +
				"Engine Version:  POSTGRES_15\n" +
				"Created At:      2025-02-01T00:00:00Z\n" +
				"HA:              No\n" +
				"Backups:         No\n" +
				"Secret:          no-resize-credentials\n" +
				"\n" +
				"Resources:\n" +
				"  CPU:               2\n" +
				"  Memory:            4.00 GB\n" +
				"Storage:\n" +
				"  Size:              50 GB\n" +
				"  Auto Resize:       No\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintVectorStoreDescribe(&buf, tt.store)
			if err != nil {
				t.Fatalf("PrintVectorStoreDescribe() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
