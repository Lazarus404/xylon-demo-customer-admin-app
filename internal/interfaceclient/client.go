package interfaceclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client forwards Control admin calls to xylon-minimal-customer-interface.
type Client struct {
	BaseURL    string
	CustomerID string
	HTTPClient *http.Client
}

// New builds a client with a 5s timeout.
func New(baseURL, customerID string) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		CustomerID: customerID,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// Result is the raw interface HTTP response to pass through.
type Result struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

// Presence GETs session presence via the platform interface.
func (c *Client) Presence(sessionID string) (Result, error) {
	return c.do(http.MethodGet, sessionID, "presence", nil, nil)
}

// History GETs session history; label may be empty.
func (c *Client) History(sessionID, label string) (Result, error) {
	q := url.Values{}
	if label != "" {
		q.Set("label", label)
	}
	return c.do(http.MethodGet, sessionID, "history", q, nil)
}

// Inject POSTs data inject body.
func (c *Client) Inject(sessionID string, body []byte) (Result, error) {
	return c.do(http.MethodPost, sessionID, "data", nil, body)
}

// Moderate POSTs moderate body.
func (c *Client) Moderate(sessionID string, body []byte) (Result, error) {
	return c.do(http.MethodPost, sessionID, "moderate", nil, body)
}

// RecordStart POSTs record/start body.
func (c *Client) RecordStart(sessionID string, body []byte) (Result, error) {
	return c.do(http.MethodPost, sessionID, "record/start", nil, body)
}

// RecordStop POSTs record/stop body.
func (c *Client) RecordStop(sessionID string, body []byte) (Result, error) {
	return c.do(http.MethodPost, sessionID, "record/stop", nil, body)
}

// RecordGet GETs record status.
func (c *Client) RecordGet(sessionID string) (Result, error) {
	return c.do(http.MethodGet, sessionID, "record", nil, nil)
}

// SessionsList GETs the tenant live session list via the platform interface.
func (c *Client) SessionsList() (Result, error) {
	return c.doCustomer(http.MethodGet, "sessions", nil)
}

// FleetSummary GETs fleet summary via the platform interface.
func (c *Client) FleetSummary() (Result, error) {
	return c.doCustomer(http.MethodGet, "fleet/summary", nil)
}

// FleetNodes GETs fleet nodes via the platform interface.
func (c *Client) FleetNodes() (Result, error) {
	return c.doCustomer(http.MethodGet, "fleet/nodes", nil)
}

// FleetAccepting POSTs drain/undrain body via the platform interface.
func (c *Client) FleetAccepting(body []byte) (Result, error) {
	return c.doCustomer(http.MethodPost, "fleet/accepting", body)
}

func (c *Client) doCustomer(method, suffix string, body []byte) (Result, error) {
	if c.BaseURL == "" {
		return Result{}, fmt.Errorf("interface_base_url empty")
	}
	if c.CustomerID == "" {
		return Result{}, fmt.Errorf("customer_id empty")
	}
	u := c.BaseURL + fmt.Sprintf("/v1/customers/%s/%s", url.PathEscape(c.CustomerID), suffix)
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return Result{}, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return Result{}, err
	}
	return Result{
		StatusCode:  res.StatusCode,
		ContentType: res.Header.Get("Content-Type"),
		Body:        raw,
	}, nil
}

func (c *Client) do(method, sessionID, action string, query url.Values, body []byte) (Result, error) {
	if c.BaseURL == "" {
		return Result{}, fmt.Errorf("interface_base_url empty")
	}
	if c.CustomerID == "" {
		return Result{}, fmt.Errorf("customer_id empty")
	}
	path := fmt.Sprintf("/v1/customers/%s/sessions/%s/%s",
		url.PathEscape(c.CustomerID), url.PathEscape(sessionID), action)
	u := c.BaseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return Result{}, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return Result{}, err
	}
	return Result{
		StatusCode:  res.StatusCode,
		ContentType: res.Header.Get("Content-Type"),
		Body:        raw,
	}, nil
}
