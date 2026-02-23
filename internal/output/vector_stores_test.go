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
