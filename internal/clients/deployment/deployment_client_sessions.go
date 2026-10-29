package deployment

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// AgentSession is a conversation session on an agent, with Raw as the platform returned it.
type AgentSession struct {
	ID         string          `json:"id"`
	CustomerID string          `json:"customerId"`
	CreatedAt  string          `json:"createdAt"`
	Title      *string         `json:"title"`
	Mode       string          `json:"mode"`
	Labels     []string        `json:"labels"`
	Raw        json.RawMessage `json:"-"`
}

type AgentSessionList struct {
	Items      []AgentSession  `json:"items"`
	TotalCount int             `json:"totalCount"`
	HasMore    bool            `json:"hasMore"`
	NextCursor *string         `json:"nextCursor"`
	Raw        json.RawMessage `json:"-"`
}

// AgentEvent is one event in a conversation session; Data's shape depends on Kind.
type AgentEvent struct {
	ID        string          `json:"id"`
	Source    string          `json:"source"`
	Kind      string          `json:"kind"`
	Offset    int             `json:"offset"`
	CreatedAt string          `json:"createdAt"`
	TraceID   string          `json:"traceId"`
	Data      json.RawMessage `json:"data"`
	Raw       json.RawMessage `json:"-"`
}

type AgentEventList struct {
	Items []AgentEvent    `json:"items"`
	Raw   json.RawMessage `json:"-"`
}

// MessageData holds the accumulated text and generation status of a message event.
type MessageData struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// StatusData is the data of a status event.
type StatusData struct {
	Status string `json:"status"`
}

const (
	EventKindMessage = "message"
	EventKindStatus  = "status"

	EventSourceCustomer = "customer"
	EventSourceAgent    = "agent"
)

type AgentSessionCreateRequest struct {
	CustomerID string   `json:"customerId,omitempty"`
	Title      string   `json:"title,omitempty"`
	Labels     []string `json:"labels,omitempty"`
}

type AgentSessionListOptions struct {
	CustomerID string
	Labels     []string
	Limit      int
	Cursor     string
	Sort       string
}

// AgentEventListOptions filters an event list; a nil MaxOffset and a zero Limit are unset.
type AgentEventListOptions struct {
	MinOffset int
	MaxOffset *int
	Limit     int
	Source    string
	Kinds     string
	TraceID   string
}

// AgentError is a refusal from the agent or the platform in front of it.
type AgentError struct {
	Status int
	Detail string
}

func (e *AgentError) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("agent returned %d", e.Status)
	}
	return fmt.Sprintf("agent returned %d: %s", e.Status, e.Detail)
}

// maxEventStreamLine leaves room for a line's wrapper around an event at the platform's 4 MiB cap.
const maxEventStreamLine = 4<<20 + 64<<10

// ErrEventStreamEnded is a followed stream that closed before it was done; following again resumes it.
var ErrEventStreamEnded = errors.New("event stream ended")

func agentSessionsPath(orgID, projectID, agentName string) string {
	return fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/agents/%s/sessions",
		url.PathEscape(orgID),
		url.PathEscape(projectID),
		url.PathEscape(agentName),
	)
}

func agentEventsPath(orgID, projectID, agentName, sessionID string) string {
	sessions := agentSessionsPath(orgID, projectID, agentName)
	return sessions + "/" + url.PathEscape(sessionID) + "/events"
}

func (c *DeploymentClient) CreateAgentSession(
	ctx context.Context,
	orgID, projectID, agentName string,
	req AgentSessionCreateRequest,
	allowGreeting bool,
) (*AgentSession, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode session request: %w", err)
	}
	path := agentSessionsPath(orgID, projectID, agentName)
	if allowGreeting {
		path += "?allowGreeting=true"
	}
	raw, err := c.agentRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	return DecodeAgentSession(raw)
}

func (c *DeploymentClient) ListAgentSessions(
	ctx context.Context,
	orgID, projectID, agentName string,
	opts AgentSessionListOptions,
) (*AgentSessionList, error) {
	path := agentSessionsPath(orgID, projectID, agentName)
	if query := AgentSessionListQuery(opts); len(query) > 0 {
		path += "?" + query.Encode()
	}
	raw, err := c.agentRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return DecodeAgentSessionList(raw)
}

// CreateAgentEvent posts an event body as the caller wrote it.
func (c *DeploymentClient) CreateAgentEvent(
	ctx context.Context,
	orgID, projectID, agentName, sessionID string,
	body json.RawMessage,
) (*AgentEvent, error) {
	raw, err := c.agentRequest(
		ctx,
		http.MethodPost,
		agentEventsPath(orgID, projectID, agentName, sessionID),
		body,
	)
	if err != nil {
		return nil, err
	}
	return DecodeAgentEvent(raw)
}

func (c *DeploymentClient) ListAgentEvents(
	ctx context.Context,
	orgID, projectID, agentName, sessionID string,
	opts AgentEventListOptions,
) (*AgentEventList, error) {
	path := agentEventsPath(orgID, projectID, agentName, sessionID) +
		"?" + AgentEventListQuery(opts).Encode()
	raw, err := c.agentRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return DecodeAgentEventList(raw)
}

func (c *DeploymentClient) GetAgentEvent(
	ctx context.Context,
	orgID, projectID, agentName, sessionID, eventID string,
) (*AgentEvent, error) {
	path := agentEventsPath(orgID, projectID, agentName, sessionID) + "/" + url.PathEscape(eventID)
	raw, err := c.agentRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return DecodeAgentEvent(raw)
}

// FollowAgentEvents delivers new events as the agent records them until ctx ends,
// onEvent returns an error, or the stream fails or ends. There is no automatic
// reconnection, as with FollowReplay; a stream that ends returns ErrEventStreamEnded.
func (c *DeploymentClient) FollowAgentEvents(
	ctx context.Context,
	orgID, projectID, agentName, sessionID string,
	opts AgentEventListOptions,
	onEvent func(AgentEvent) error,
) error {
	query := AgentEventListQuery(opts)
	query.Set("follow", "true")
	path := agentEventsPath(orgID, projectID, agentName, sessionID) + "?" + query.Encode()
	if err := c.followAgentStream(ctx, path, onEvent); err != nil {
		return err
	}
	return ErrEventStreamEnded
}

// FollowAgentEvent delivers each update of one event until the agent reports it
// complete, returning its final state.
func (c *DeploymentClient) FollowAgentEvent(
	ctx context.Context,
	orgID, projectID, agentName, sessionID, eventID string,
	onUpdate func(AgentEvent) error,
) (*AgentEvent, error) {
	path := agentEventsPath(orgID, projectID, agentName, sessionID) +
		"/" + url.PathEscape(eventID) + "?follow=true"
	var last *AgentEvent
	err := c.followAgentStream(ctx, path, func(ev AgentEvent) error {
		last = &ev
		return onUpdate(ev)
	})
	if err != nil {
		return nil, err
	}
	if last == nil || last.Streaming() {
		return nil, ErrEventStreamEnded
	}
	return last, nil
}

// followAgentStream reads one NDJSON stream until it closes, whether the agent
// finished or the platform's stream window ended.
func (c *DeploymentClient) followAgentStream(
	ctx context.Context,
	path string,
	onEvent func(AgentEvent) error,
) error {
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("event request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return DecodeAgentError(resp.StatusCode, raw)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), maxEventStreamLine)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		ev, done, err := DecodeAgentEventLine(line)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		if ev == nil {
			continue
		}
		if err := onEvent(*ev); err != nil {
			return err
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read the event stream: %w", err)
	}
	return nil
}

// agentRequest sends a request to the platform's agent routes and returns the body of a success.
func (c *DeploymentClient) agentRequest(
	ctx context.Context,
	method, path string,
	body []byte,
) ([]byte, error) {
	req, err := c.newRequest(ctx, method, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if body != nil {
		clients.SetJSONBody(req, body)
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("agent request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the agent response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, DecodeAgentError(resp.StatusCode, raw)
	}
	return raw, nil
}

func AgentSessionListQuery(opts AgentSessionListOptions) url.Values {
	query := url.Values{}
	if opts.CustomerID != "" {
		query.Set("customerId", opts.CustomerID)
	}
	for _, label := range opts.Labels {
		query.Add("labels", label)
	}
	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		query.Set("cursor", opts.Cursor)
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	return query
}

func AgentEventListQuery(opts AgentEventListOptions) url.Values {
	query := url.Values{"minOffset": {strconv.Itoa(opts.MinOffset)}}
	if opts.MaxOffset != nil {
		query.Set("maxOffset", strconv.Itoa(*opts.MaxOffset))
	}
	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Source != "" {
		query.Set("source", opts.Source)
	}
	if opts.Kinds != "" {
		query.Set("kinds", opts.Kinds)
	}
	if opts.TraceID != "" {
		query.Set("traceId", opts.TraceID)
	}
	return query
}

func DecodeAgentSession(raw []byte) (*AgentSession, error) {
	var session AgentSession
	if err := json.Unmarshal(raw, &session); err != nil {
		return nil, fmt.Errorf("failed to decode session: %w", err)
	}
	session.Raw = raw
	return &session, nil
}

func DecodeAgentSessionList(raw []byte) (*AgentSessionList, error) {
	var list AgentSessionList
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("failed to decode session list: %w", err)
	}
	list.Raw = raw
	return &list, nil
}

func DecodeAgentEvent(raw []byte) (*AgentEvent, error) {
	var event AgentEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, fmt.Errorf("failed to decode event: %w", err)
	}
	event.Raw = raw
	return &event, nil
}

func DecodeAgentEventList(raw []byte) (*AgentEventList, error) {
	var list struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("failed to decode event list: %w", err)
	}
	items := make([]AgentEvent, 0, len(list.Items))
	for _, item := range list.Items {
		event, err := DecodeAgentEvent(item)
		if err != nil {
			return nil, err
		}
		items = append(items, *event)
	}
	return &AgentEventList{Items: items, Raw: raw}, nil
}

// DecodeAgentEventLine reads one line of a follow stream: an event, a heartbeat
// (nil event), the end of the stream window (done), or the stream's error.
func DecodeAgentEventLine(line []byte) (event *AgentEvent, done bool, err error) {
	var record struct {
		Event   json.RawMessage `json:"event"`
		Error   string          `json:"error"`
		Timeout bool            `json:"timeout"`
	}
	if err := json.Unmarshal(line, &record); err != nil {
		return nil, false, fmt.Errorf("failed to decode event stream: %w", err)
	}
	switch {
	case record.Error != "":
		return nil, false, errors.New(record.Error)
	case record.Timeout:
		return nil, true, nil
	case len(record.Event) > 0:
		event, err := DecodeAgentEvent(record.Event)
		return event, false, err
	}
	return nil, false, nil
}

// DecodeAgentError reads the agent's error bodies ({"detail": text | 422 list}) or the platform's ({"message": text}).
func DecodeAgentError(status int, body []byte) *AgentError {
	return &AgentError{Status: status, Detail: agentErrorDetail(body)}
}

// Message returns a message event's data, or false for any other event.
func (e AgentEvent) Message() (MessageData, bool) {
	if e.Kind != EventKindMessage {
		return MessageData{}, false
	}
	var data MessageData
	if err := json.Unmarshal(e.Data, &data); err != nil {
		return MessageData{}, false
	}
	return data, true
}

// Streaming reports whether the agent is still adding to a message event.
func (e AgentEvent) Streaming() bool {
	data, ok := e.Message()
	return ok && data.Status == "streaming"
}

// Status returns a status event's status, or empty for any other event.
func (e AgentEvent) Status() string {
	if e.Kind != EventKindStatus {
		return ""
	}
	var data StatusData
	if err := json.Unmarshal(e.Data, &data); err != nil {
		return ""
	}
	return data.Status
}
