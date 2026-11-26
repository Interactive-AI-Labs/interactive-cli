package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
)

// AgentOverviewOptions are the query parameters the agent Overview page sends.
type AgentOverviewOptions struct {
	Window string `url:"window,omitempty"`
	From   string `url:"from,omitempty"`
	To     string `url:"to,omitempty"`
	Groups bool   `url:"groups"`
}

type MetricDelta struct {
	Value    *float64 `json:"value"`
	Previous *float64 `json:"previous"`
	DeltaPct *float64 `json:"deltaPct"`
}

// Figure is a money amount, or kind "unavailable" when the platform cannot price it.
type Figure struct {
	Kind  string `json:"kind"`
	Money *struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"money"`
}

type CostMetric struct {
	Value    Figure   `json:"value"`
	Previous Figure   `json:"previous"`
	DeltaPct *float64 `json:"deltaPct"`
}

type NameCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type AgentOverview struct {
	DisplayName string `json:"displayName"`
	Window      struct {
		Key  string `json:"key"`
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"window"`
	Performance struct {
		Sessions              MetricDelta `json:"sessions"`
		SessionLatencyMs      MetricDelta `json:"sessionLatencyMs"`
		Traces                MetricDelta `json:"traces"`
		TraceLatencyMs        MetricDelta `json:"traceLatencyMs"`
		AvgTracesPerSession   MetricDelta `json:"avgTracesPerSession"`
		AvgIterationsPerTrace MetricDelta `json:"avgIterationsPerTrace"`
		Observations          MetricDelta `json:"observations"`
		Scores                MetricDelta `json:"scores"`
	} `json:"performance"`
	Cost struct {
		Agent            CostMetric `json:"agent"`
		InteractiveAgent CostMetric `json:"interactiveAgent"`
		LLM              CostMetric `json:"llm"`
		Tokens           *struct {
			Rated int `json:"rated"`
		} `json:"tokens"`
	} `json:"cost"`
	Errors struct {
		Sessions []struct {
			SessionID  string `json:"sessionId"`
			ErrorCount int    `json:"errorCount"`
		} `json:"sessions"`
		Traces []struct {
			TraceID    string `json:"traceId"`
			ErrorCount int    `json:"errorCount"`
		} `json:"traces"`
		Types []struct {
			ObservationType  string `json:"observationType"`
			Message          string `json:"message"`
			ObservationCount int    `json:"observationCount"`
		} `json:"types"`
	} `json:"errors"`
	Usage struct {
		DocumentsReferenced []NameCount `json:"documentsReferenced"`
		PoliciesMatched     []NameCount `json:"policiesMatched"`
		RoutineTriggers     []NameCount `json:"routineTriggers"`
		ToolCalls           []NameCount `json:"toolCalls"`
	} `json:"usage"`
}

func (c *APIClient) GetAgentOverview(
	ctx context.Context,
	orgID, projectID, agent string,
	opts AgentOverviewOptions,
) (AgentOverview, json.RawMessage, error) {
	var overview AgentOverview
	req, err := c.newRequest(ctx, http.MethodGet,
		evalBasePath(orgID, projectID)+"/agents/"+url.PathEscape(agent)+"/overview")
	if err != nil {
		return overview, nil, fmt.Errorf("failed to create request: %w", err)
	}
	q, err := query.Values(opts)
	if err != nil {
		return overview, nil, fmt.Errorf("failed to encode query parameters: %w", err)
	}
	req.URL.RawQuery = q.Encode()
	return decodeBody[AgentOverview](c, req, "get agent overview")
}

// PerformanceFilter is a filter of the project home's agent table.
type PerformanceFilter struct {
	Column   string   `json:"column"`
	Type     string   `json:"type"`
	Operator string   `json:"operator"`
	Value    []string `json:"value"`
}

type AgentPerformanceBody struct {
	FromTimestamp string              `json:"fromTimestamp"`
	ToTimestamp   string              `json:"toTimestamp"`
	Filters       []PerformanceFilter `json:"filters,omitempty"`
	Groups        bool                `json:"groups"`
	AgentIDs      []string            `json:"agentIds,omitempty"`
}

type AgentPerformance struct {
	Agents []struct {
		AgentID        string   `json:"agent_id"`
		Name           string   `json:"name"`
		Traces         *int     `json:"traces"`
		Sessions       *int     `json:"sessions"`
		TraceLatency   *float64 `json:"trace_latency"`
		SessionLatency *float64 `json:"session_latency"`
	} `json:"agents"`
	AgentCosts []struct {
		AgentID string `json:"agent_id"`
		Total   Figure `json:"total"`
	} `json:"agentCosts"`
	CostTotal Figure `json:"costTotal"`
}

func (c *APIClient) GetAgentPerformance(
	ctx context.Context,
	orgID, projectID string,
	body AgentPerformanceBody,
) (AgentPerformance, json.RawMessage, error) {
	req, err := c.newJSONRequest(ctx, http.MethodPost,
		evalBasePath(orgID, projectID)+"/agent-performance", body)
	if err != nil {
		return AgentPerformance{}, nil, fmt.Errorf("failed to create request: %w", err)
	}
	perf, raw, err := decodeBody[AgentPerformance](c, req, "get agent performance")
	if err != nil || len(body.AgentIDs) == 0 {
		return perf, raw, err
	}

	// agent-performance's costTotal is every agent's; cost-total prices only the chosen ones.
	req, err = c.newJSONRequest(ctx, http.MethodPost,
		evalBasePath(orgID, projectID)+"/agent-performance/cost-total", body)
	if err != nil {
		return perf, nil, fmt.Errorf("failed to create request: %w", err)
	}
	var total json.RawMessage
	if perf.CostTotal, total, err = decodeBody[Figure](c, req, "get agent cost total"); err != nil {
		return perf, nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return perf, nil, fmt.Errorf("failed to decode get agent performance response: %w", err)
	}
	fields["costTotal"] = total
	raw, err = json.Marshal(fields)
	return perf, raw, err
}

// decodeBody reads a response the platform returns without the success envelope.
func decodeBody[T any](c *APIClient, req *http.Request, action string) (T, json.RawMessage, error) {
	var v T
	body, err := c.doAndRead(req, action)
	if err != nil {
		return v, nil, err
	}
	if err := json.Unmarshal(body, &v); err != nil {
		return v, nil, fmt.Errorf("failed to decode %s response: %w", action, err)
	}
	return v, json.RawMessage(body), nil
}

type AgentActivityOptions struct {
	Page   int      `url:"page,omitempty"`
	Limit  int      `url:"limit,omitempty"`
	Search string   `url:"q,omitempty"`
	Actors []string `url:"actor,omitempty"`
}

type AgentActivity struct {
	Activity []struct {
		Time    string         `json:"time"`
		Action  string         `json:"action"`
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
		Actor   struct {
			ID      string `json:"id"`
			Display string `json:"display"`
		} `json:"actor"`
	} `json:"activity"`
	Meta PageMeta `json:"meta"`
}

func (c *APIClient) ListAgentActivity(
	ctx context.Context,
	orgID, projectID, agent string,
	opts AgentActivityOptions,
) (AgentActivity, json.RawMessage, error) {
	return doList[AgentActivity](c, ctx,
		evalBasePath(orgID, projectID)+"/agents/"+url.PathEscape(agent)+"/activity",
		opts, "list agent activity")
}
