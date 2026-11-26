package platform

import "testing"

func TestWithCostTotal(t *testing.T) {
	tests := []struct{ raw, want string }{
		{`null`, `{"costTotal":1}`},
		{`{}`, `{"costTotal":1}`},
		{`{"agents":[],"costTotal":9}`, `{"agents":[],"costTotal":1}`},
	}
	for _, tt := range tests {
		got, err := withCostTotal([]byte(tt.raw), []byte(`1`))
		if err != nil || string(got) != tt.want {
			t.Errorf("withCostTotal(%s) = %s, %v; want %s", tt.raw, got, err, tt.want)
		}
	}
}
