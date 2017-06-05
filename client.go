package bitbank

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	PublicEndpoint  = "https://public.bitbank.cc/"
	PrivateEndpoint = "https://api.bitbank.cc/v1/"
)

type Client struct {
	PublicBaseURL  *url.URL
	PrivateBaseURL *url.URL
	// Services
	Ticker *TickerService
	Pubnub *PubnubService
}

type BaseResponseJSON struct {
	Success int
	Data    interface{}
}

// NewClient creates new Bitfinex.com API client.
func NewClient() *Client {
	publicBaseURL, _ := url.Parse(PublicEndpoint)
	privateBaseURL, _ := url.Parse(PrivateEndpoint)

	c := &Client{PublicBaseURL: publicBaseURL, PrivateBaseURL: privateBaseURL}
	c.Ticker = &TickerService{client: c}
	c.Pubnub = NewPubnubService(c)

	return c
}

// NewRequest create new API request. Relative url can be provided in refURL.
func (c *Client) newRequest(method string, refURL string, params url.Values) (*http.Request, error) {
	rel, err := url.Parse(refURL)
	if err != nil {
		return nil, err
	}
	if params != nil {
		rel.RawQuery = params.Encode()
	}
	var req *http.Request
	u := c.PublicBaseURL.ResolveReference(rel)
	req, err = http.NewRequest(method, u.String(), nil)

	if err != nil {
		return nil, err
	}

	return req, nil
}

var httpDo = func(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

// Do executes API request created by NewRequest method or custom *http.Request.
func (c *Client) do(req *http.Request, v interface{}) (*Response, error) {
	resp, err := httpDo(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := newResponse(resp)

	err = checkResponse(response)
	if err != nil {
		// Return response in case caller need to debug it.
		return response, err
	}

	if v != nil {
		err = json.Unmarshal(response.Body, v)
		if err != nil {
			return response, err
		}
	}

	return response, nil
}

// Response is wrapper for standard http.Response and provides
// more methods.
type Response struct {
	Response *http.Response
	Body     []byte
}

// newResponse creates new wrapper.
func newResponse(r *http.Response) *Response {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		body = []byte(`Error reading body:` + err.Error())
	}

	return &Response{r, body}
}

// String converts response body to string.
// An empty string will be returned if error.
func (r *Response) String() string {
	return string(r.Body)
}

// ErrorResponse is the custom error type that is returned if the API returns an
// error.
type ErrorResponse struct {
	Response *Response
	Message  string
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d, %v",
		r.Response.Response.Request.Method,
		r.Response.Response.Request.URL,
		r.Response.Response.StatusCode,
		r.Message,
	)
}

type APIError struct {
	Code int
}

// checkResponse checks response status code and response
// for errors.
func checkResponse(r *Response) error {
	if c := r.Response.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}

	// Try to decode error message
	d := &APIError{}
	j := &BaseResponseJSON{Data: d}
	err := json.Unmarshal(r.Body, j)

	if err != nil {
		errorResponse.Message = "Error decoding response error message. " +
			"Please see response body for more information."
	} else {
		errorResponse.Message = "API error code: " + strconv.Itoa(d.Code) +
			". Please see https://docs.bitbank.cc/error_code/ for more information."
	}

	return errorResponse
}
