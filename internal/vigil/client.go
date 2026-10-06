// Package vigil is a small read client for a local Vigil server.
package vigil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client reads events from Vigil's log API.
type Client struct {
	BaseURL   string
	ProjectID string
	HTTP      *http.Client
}

func New(baseURL, projectID string) *Client {
	return &Client{
		BaseURL:   baseURL,
		ProjectID: projectID,
		HTTP:      &http.Client{Timeout: 30 * time.Second},
	}
}

// Envelope mirrors the subset of Vigil's stored event we consume.
type Envelope struct {
	SchemaVersion int             `json:"schema_version"`
	ProjectID     string          `json:"project_id"`
	Kind          string          `json:"kind"`
	TS            string          `json:"ts"`
	Source        string          `json:"source"`
	TraceID       string          `json:"trace_id"`
	Level         string          `json:"level"`
	Name          string          `json:"name"`
	Attrs         json.RawMessage `json:"attrs"`
	Body          json.RawMessage `json:"body"`
}

// Event is one stored Vigil event.
type Event struct {
	EventID    string `json:"event_id"`
	ReceivedAt string `json:"received_at"`
	Envelope
}

// LogList is one page of the /api/logs response.
type LogList struct {
	Events []Event `json:"events"`
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
	Total  int     `json:"total"`
}

// AttrsMap decodes an event's attrs object into a map.
func (e Event) AttrsMap() map[string]any {
	out := map[string]any{}
	if len(e.Attrs) == 0 {
		return out
	}
	_ = json.Unmarshal(e.Attrs, &out)
	return out
}

func (e Event) Time() time.Time {
	ts, err := time.Parse(time.RFC3339Nano, e.TS)
	if err != nil {
		return time.Time{}
	}
	return ts
}

func (c *Client) logsURL(params url.Values) string {
	q := url.Values{}
	for k, vs := range params {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	q.Set("project_id", c.ProjectID)
	return c.BaseURL + "/api/logs?" + q.Encode()
}

// Logs fetches all events matching namePrefix within the window, paginating
// until exhausted. namePrefix filters client-side to avoid depending on
// Vigil's exact name-match semantics.
func (c *Client) Logs(from time.Time, namePrefix string) ([]Event, error) {
	var all []Event
	page := 1
	for {
		params := url.Values{}
		params.Set("from", from.Format(time.RFC3339))
		params.Set("limit", "100")
		params.Set("page", strconv.Itoa(page))

		resp, err := c.HTTP.Get(c.logsURL(params))
		if err != nil {
			return nil, fmt.Errorf("vigil request: %w", err)
		}
		var list LogList
		decodeErr := json.NewDecoder(resp.Body).Decode(&list)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("vigil /api/logs: status %d", resp.StatusCode)
		}
		if decodeErr != nil {
			return nil, fmt.Errorf("vigil decode: %w", decodeErr)
		}
		if len(list.Events) == 0 {
			break
		}
		for _, ev := range list.Events {
			if namePrefix == "" || len(ev.Name) >= len(namePrefix) && ev.Name[:len(namePrefix)] == namePrefix {
				all = append(all, ev)
			}
		}
		if len(all) >= list.Total && list.Total > 0 {
			break
		}
		page++
		if page > 1000 { // safety valve
			break
		}
	}
	return all, nil
}
