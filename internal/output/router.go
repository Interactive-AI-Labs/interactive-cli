package output

import (
	"fmt"
	"io"
	"net/url"
)

const routerDocumentationURL = "https://docs.interactive.ai/llm-router"

type RouterEndpoint struct {
	Method      string `json:"method"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type RouterEndpoints struct {
	ChatCompletions RouterEndpoint `json:"chatCompletions"`
	Responses       RouterEndpoint `json:"responses"`
	Embeddings      RouterEndpoint `json:"embeddings"`
	Rerank          RouterEndpoint `json:"rerank"`
	Models          RouterEndpoint `json:"models"`
	Health          RouterEndpoint `json:"health"`
}

type RouterInfo struct {
	BaseURL          string          `json:"baseUrl"`
	Endpoints        RouterEndpoints `json:"endpoints"`
	DocumentationURL string          `json:"documentationUrl"`
}

func NewRouterInfo(hostname string) (RouterInfo, error) {
	u, err := url.Parse(hostname)
	if err != nil {
		return RouterInfo{}, fmt.Errorf("failed to parse API hostname %q: %w", hostname, err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") ||
		u.Host == "" ||
		u.User != nil ||
		(u.Path != "" && u.Path != "/") ||
		u.RawQuery != "" ||
		u.Fragment != "" {
		return RouterInfo{}, fmt.Errorf("invalid API hostname %q", hostname)
	}

	u.Path = ""
	baseURL := u.JoinPath("api", "v1").String()
	endpoint := func(method, path, description string) RouterEndpoint {
		return RouterEndpoint{
			Method:      method,
			URL:         baseURL + path,
			Description: description,
		}
	}

	return RouterInfo{
		BaseURL: baseURL,
		Endpoints: RouterEndpoints{
			ChatCompletions: endpoint(
				"POST",
				"/chat/completions",
				"Generate chat responses, optionally returning them as they are created.",
			),
			Responses: endpoint(
				"POST",
				"/responses",
				"Generate model responses from text, messages, or tool results.",
			),
			Embeddings: endpoint(
				"POST",
				"/embeddings",
				"Convert text into numbers that can be compared by meaning.",
			),
			Rerank: endpoint(
				"POST",
				"/rerank",
				"Sort candidate documents by their relevance to a query.",
			),
			Models: endpoint(
				"GET",
				"/models",
				"List models available for the request region.",
			),
			Health: endpoint(
				"GET",
				"/health/llm-router",
				"Check whether the inference router is healthy.",
			),
		},
		DocumentationURL: routerDocumentationURL,
	}, nil
}

func PrintRouterInfo(out io.Writer, info RouterInfo) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Base URL:\t%s\n", info.BaseURL)
	fmt.Fprintf(w, "Documentation:\t%s\n", info.DocumentationURL)
	fmt.Fprintln(w, "Endpoints:")

	endpoints := []struct {
		name     string
		endpoint RouterEndpoint
	}{
		{"Chat Completions", info.Endpoints.ChatCompletions},
		{"Responses", info.Endpoints.Responses},
		{"Embeddings", info.Endpoints.Embeddings},
		{"Rerank", info.Endpoints.Rerank},
		{"Models", info.Endpoints.Models},
		{"Health", info.Endpoints.Health},
	}
	for _, item := range endpoints {
		fmt.Fprintf(w, "  %s:\n", item.name)
		fmt.Fprintf(w, "    Method:\t%s\n", item.endpoint.Method)
		fmt.Fprintf(w, "    URL:\t%s\n", item.endpoint.URL)
		fmt.Fprintf(w, "    Description:\t%s\n", item.endpoint.Description)
	}
	return w.Flush()
}
