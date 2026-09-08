package replay

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

type fakeSecrets struct {
	secrets map[string]map[string]string
	err     error
	fetched []string
}

func (f *fakeSecrets) GetSecret(
	_ context.Context, _, _, name string,
) (*deployment.SecretInfo, error) {
	f.fetched = append(f.fetched, name)
	if f.err != nil {
		return nil, f.err
	}
	return &deployment.SecretInfo{Name: name, Data: f.secrets[name]}, nil
}

func describedAgent(refs ...string) *deployment.DescribeAgentResponse {
	var secretRefs []deployment.SecretRef
	for _, r := range refs {
		secretRefs = append(secretRefs, deployment.SecretRef{SecretName: r})
	}
	return &deployment.DescribeAgentResponse{
		AgentConfig: map[string]any{"runtime": map[string]any{"api_key": "${AGENT_API_KEY}"}},
		SecretRefs:  secretRefs,
	}
}

func TestResolveBearer(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("from-secret"))

	t.Run("flag wins and reads nothing", func(t *testing.T) {
		secrets := &fakeSecrets{}
		var errW bytes.Buffer
		got, err := ResolveBearer(
			context.Background(),
			secrets,
			"o",
			"p",
			describedAgent("s1"),
			"from-flag",
			"from-env",
			&errW,
		)
		if err != nil || got != "from-flag" || len(secrets.fetched) != 0 || errW.Len() != 0 {
			t.Errorf("got %q err %v fetched %v stderr %q", got, err, secrets.fetched, errW.String())
		}
	})

	t.Run("env var second", func(t *testing.T) {
		got, err := ResolveBearer(
			context.Background(),
			&fakeSecrets{},
			"o",
			"p",
			describedAgent("s1"),
			"",
			"from-env",
			&bytes.Buffer{},
		)
		if err != nil || got != "from-env" {
			t.Errorf("got %q err %v", got, err)
		}
	})

	t.Run("agent env list before secrets", func(t *testing.T) {
		agent := describedAgent("s1")
		agent.Env = []deployment.EnvVar{{Name: "AGENT_API_KEY", Value: "literal"}}
		secrets := &fakeSecrets{}
		var errW bytes.Buffer
		got, err := ResolveBearer(context.Background(), secrets, "o", "p", agent, "", "", &errW)
		if err != nil || got != "literal" || len(secrets.fetched) != 0 {
			t.Errorf("got %q err %v fetched %v", got, err, secrets.fetched)
		}
		if errW.String() != "using AGENT_API_KEY from the agent's env\n" {
			t.Errorf("stderr = %q", errW.String())
		}
	})

	t.Run("first secret with the key wins, later ones never fetched", func(t *testing.T) {
		secrets := &fakeSecrets{secrets: map[string]map[string]string{
			"services-dev":  {"OTHER": "x"},
			"platform-dev":  {"AGENT_API_KEY": encoded},
			"databases-dev": {"AGENT_API_KEY": "should-not-be-read"},
		}}
		var errW bytes.Buffer
		got, err := ResolveBearer(
			context.Background(), secrets, "o", "p",
			describedAgent("services-dev", "platform-dev", "databases-dev"), "", "", &errW,
		)
		if err != nil || got != "from-secret" {
			t.Fatalf("got %q err %v", got, err)
		}
		if strings.Join(secrets.fetched, ",") != "services-dev,platform-dev" {
			t.Errorf("fetched = %v", secrets.fetched)
		}
		if errW.String() != "using AGENT_API_KEY from secret platform-dev\n" {
			t.Errorf("stderr = %q", errW.String())
		}
		if strings.Contains(errW.String(), "from-secret") ||
			strings.Contains(errW.String(), encoded) {
			t.Errorf("stderr leaks the key: %q", errW.String())
		}
	})

	t.Run("raw value kept when not base64", func(t *testing.T) {
		secrets := &fakeSecrets{
			secrets: map[string]map[string]string{"s1": {"AGENT_API_KEY": "not base64!"}},
		}
		got, err := ResolveBearer(
			context.Background(), secrets, "o", "p", describedAgent("s1"), "", "", &bytes.Buffer{},
		)
		if err != nil || got != "not base64!" {
			t.Errorf("got %q err %v", got, err)
		}
	})

	t.Run("no env ref in config", func(t *testing.T) {
		agent := &deployment.DescribeAgentResponse{AgentConfig: map[string]any{}}
		_, err := ResolveBearer(
			context.Background(),
			&fakeSecrets{},
			"o",
			"p",
			agent,
			"",
			"",
			&bytes.Buffer{},
		)
		if err == nil || !strings.Contains(err.Error(), "runtime.api_key") {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("key in no secret", func(t *testing.T) {
		secrets := &fakeSecrets{secrets: map[string]map[string]string{"s1": {"OTHER": "x"}}}
		_, err := ResolveBearer(
			context.Background(), secrets, "o", "p", describedAgent("s1"), "", "", &bytes.Buffer{},
		)
		if err == nil || !strings.Contains(err.Error(), "not found in the agent's env or secrets") {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("secret read failure surfaces", func(t *testing.T) {
		secrets := &fakeSecrets{err: errors.New("forbidden")}
		_, err := ResolveBearer(
			context.Background(), secrets, "o", "p", describedAgent("s1"), "", "", &bytes.Buffer{},
		)
		if err == nil || !strings.Contains(err.Error(), `failed to read secret "s1"`) ||
			!strings.Contains(err.Error(), "forbidden") {
			t.Errorf("error = %v", err)
		}
	})
}
