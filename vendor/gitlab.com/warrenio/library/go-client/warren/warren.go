package warren

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

/* API Client */

// Client Schema for Warren API Client
type Client struct {
	client *http.Client

	BaseURL *url.URL
	// LocationSlug must be set if managing resources that are location specific.
	// Can be ignored if there are no locations or just one location defined.
	LocationSlug string
	ApiToken     string

	VirtualMachine *VirtualMachineService
	Location       *LocationService
	Network        *NetworkService
	BlockStorage   *BlockStorageService
}

type ClientBuilder struct {
	_ApiUrl       *string
	_ApiToken     *string
	_LocationSlug *string
	_client       *http.Client
}

/* Common data structures */

// ResponseError Schema for client api call response errors
type ResponseError struct {
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

/* Builder */

func (cb *ClientBuilder) ApiUrl(baseUrl string) *ClientBuilder {
	cb._ApiUrl = &baseUrl
	return cb
}

func (cb *ClientBuilder) ApiToken(token string) *ClientBuilder {
	cb._ApiToken = &token
	return cb
}

func (cb *ClientBuilder) LocationSlug(slug string) *ClientBuilder {
	cb._LocationSlug = &slug
	return cb
}

func (cb *ClientBuilder) Client(client *http.Client) *ClientBuilder {
	cb._client = client
	return cb
}

func (cb *ClientBuilder) Build() (*Client, error) {
	return NewClientFromBuilder(cb)
}

/* Helper functions */

// NewClient builds and returns a Warren API client. Providing a custom httpClient is optional.
func NewClient(baseUrl string, apiToken string, httpClient *http.Client) (*Client, error) {
	return (&ClientBuilder{}).
		ApiUrl(baseUrl).
		ApiToken(apiToken).
		Client(httpClient).
		Build()
}

func NewClientFromBuilder(builder *ClientBuilder) (*Client, error) {
	if builder._ApiUrl == nil {
		return nil, fmt.Errorf("ApiUrl must not be empty")
	}
	u, err := url.Parse(*builder._ApiUrl)
	if err != nil {
		return nil, err
	}
	httpClient := builder._client
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	apiToken := ""
	if builder._ApiToken != nil {
		apiToken = *builder._ApiToken
	}
	c := &Client{client: httpClient, BaseURL: u, ApiToken: apiToken}
	if builder._LocationSlug != nil {
		c.LocationSlug = *builder._LocationSlug
	}
	c.VirtualMachine = &VirtualMachineService{client: c}
	c.Location = &LocationService{client: c}
	c.Network = &NetworkService{client: c}
	c.BlockStorage = &BlockStorageService{client: c}
	return c, nil
}

// New returns a pointer to a copy of the input value. New("text") returns a pointer to the string "text".
func New[T any](v T) *T {
	return &v
}

/* Internal stuff */

// ApiCall Schema for API call structure
type ApiCall struct {
	method       string
	slug         string
	path         string
	responseData any
	headers      map[string]string
	queryParams  map[string]string
	formParams   map[string]string
	jsonBody     any
}

// BuildBody Helper function for building API call body
func BuildBody(params ApiCall) (*strings.Reader, string, error) {
	contentType := ""
	var body *strings.Reader
	// What kind of data there is?
	// Form parameters
	if len(params.formParams) > 0 {
		contentType = "application/x-www-form-urlencoded"
		data := url.Values{}
		for key, value := range params.formParams {
			data.Set(key, value)
		}
		body = strings.NewReader(data.Encode())
	}
	// JSON body
	if params.jsonBody != nil {
		contentType = "application/json; charset=utf-8"
		bodyBytes, err := json.Marshal(params.jsonBody)
		if err != nil {
			return nil, "", fmt.Errorf("failed to convert jsonBody to JSON: %w", err)
		}
		body = strings.NewReader(string(bodyBytes[:]))
	}
	if body == nil {
		body = strings.NewReader("")
	}
	return body, contentType, nil
}

// Call Function to execute API calls to defined datacenter locations and resources
func (c *Client) Call(params ApiCall) error {
	slug := ""
	if len(c.LocationSlug) > 0 {
		slug = fmt.Sprintf("/%s", c.LocationSlug)
	}
	if len(params.slug) > 0 {
		slug = fmt.Sprintf("/%s", params.slug)
	}
	rel := &url.URL{Path: fmt.Sprintf("/v1%s%s", slug, params.path)}
	fullUrl := c.BaseURL.ResolveReference(rel)
	query := fullUrl.Query()
	for key, value := range params.queryParams {
		query.Add(key, value)
	}
	fullUrl.RawQuery = query.Encode()
	body, contentType, err := BuildBody(params)
	req, err := http.NewRequest(params.method, fullUrl.String(), body)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("apikey", c.ApiToken)
	if len(contentType) > 0 {
		req.Header.Set("Content-Type", contentType)
	}
	for key, value := range params.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call HTTP request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Print("Failed to close response body", err)
		}
	}(resp.Body)
	reqId := resp.Header.Get("X-Warren-Correlation-Id")
	if resp.StatusCode >= 300 {
		var errMess = new(ResponseError)
		buf := new(strings.Builder)
		_, err := io.Copy(buf, resp.Body)
		if err != nil {
			return fmt.Errorf("[%d] failed to read error response: %w %s", resp.StatusCode, err, reqId)
		}
		if buf.Len() == 0 {
			return fmt.Errorf("[%d] empty error response %s", resp.StatusCode, reqId)
		}
		err = json.NewDecoder(strings.NewReader(buf.String())).Decode(errMess)
		if err != nil {
			return fmt.Errorf("[%d] failed to parse error response, body: %s, err: %w %s", resp.StatusCode, buf.String(), err, reqId)
		}
		if len(errMess.Message) == 0 && len(errMess.Errors) == 0 {
			return fmt.Errorf("[%d] failed to parse meaningful data from error response, body: %s %s", resp.StatusCode, buf.String(), reqId)
		}
		return fmt.Errorf("[%d] %v, %v, %v", resp.StatusCode, errMess.Message, errMess.Errors, reqId)
	}
	if params.responseData != nil {
		err = json.NewDecoder(resp.Body).Decode(params.responseData)
		if err != nil {
			return fmt.Errorf("failed to parse response: %w %s", err, reqId)
		}
	}
	return nil
}
