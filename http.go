package wordfeud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// roundtrip executes an HTTP request to path using method and unmarshalls the Content of the response
// into a new C, a pointer to which is returned.
//
// If session is set to anything but the empty string, it will be included in the "Cookie" header of the request.
// If body is not nil, it will be marshalled to JSON and sent as the request body.
func roundtrip[C any](c *Client, method string, path string, session SessionID, body any) (*C, error) {
	var b []byte
	if body != nil {
		var err error
		b, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshalling request body: %v", err)
		}
	}
	res, err := c.request(method, path, session, b)
	if err != nil {
		return nil, err
	}
	if res.Content == nil {
		return nil, fmt.Errorf("response body content field is empty")
	}

	t := *new(C)
	err = json.Unmarshal(res.Content, &t)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling response body: %v", err)
	}
	return &t, nil
}

type response struct {
	Content json.RawMessage
	Header  http.Header
}

// request executes an HTTP request to path using method. It reads the response body in full and
// returns a response.
//
// If session is set to anything but the empty string, it will be included in the Cookie header of the request.
// If body is not nil, it will be marshalled to JSON and sent as the request body.
//
// The Wordfeud API does not follow convention when it comes to what HTTP status codes are returned.
// To alleviate this, part of the response body is eagerly parsed in the search of errors, even if
// the HTTP status code is 200. This method will return a non-nil error if the response body "status"
// field is equal to "error" (or if the status code is not 200).
func (c *Client) request(method string, path string, session SessionID, body []byte) (*response, error) {
	// Trailing slash is required.
	url := fmt.Sprintf("%s/%s/", c.baseURL, strings.Trim(path, "/"))
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}
	req.Header.Add("Accept", "application/json")
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	if session != "" {
		req.Header.Add("Cookie", session.cookie())
	}

	res, err := c.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %v", err)
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			panic(fmt.Errorf("closing response body: %v", err))
		}
	}(res.Body)

	var b bytes.Buffer
	_, err = io.Copy(&b, res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	bodyBytes := b.Bytes()

	if len(bodyBytes) == 0 {
		if res.StatusCode == http.StatusOK {
			return &response{Content: nil, Header: res.Header}, nil
		}
		return nil, fmt.Errorf("status code %d (no body)", res.StatusCode)
	}

	// Internal Server Errors are sent as text/html instead of JSON (which is sent as text/plain(!))
	// so in that case we bail here and include the body in the error message.
	if res.StatusCode >= http.StatusInternalServerError {
		return nil, fmt.Errorf("status code %d: %s", res.StatusCode, bodyBytes)
	}

	var responseBody struct {
		Status  string          `json:"status"`
		Content json.RawMessage `json:"content"`
	}

	err = json.Unmarshal(bodyBytes, &responseBody)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling response body (status code %d): %v", res.StatusCode, err)
	}
	if responseBody.Status == "error" {
		var e *apiError
		err = json.Unmarshal(responseBody.Content, &e)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling error response (status code %d): %v", res.StatusCode, err)
		}
		return nil, convertToSentinel(e)
	}

	r := &response{
		Content: responseBody.Content,
		Header:  res.Header,
	}
	return r, nil
}

type apiError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e *apiError) Error() string {
	msg := e.Type
	if e.Message != "" {
		msg = fmt.Sprintf("%s: %s", msg, e.Message)
	}
	return msg
}

var cookieRegexp = regexp.MustCompile(`sessionid=(.+?)(?:;|$)`)

func extractSessionID(r *response) (SessionID, error) {
	cookie := r.Header.Get("Set-Cookie")
	if cookie == "" {
		return "", fmt.Errorf("cookie not found in response")
	}

	matches := cookieRegexp.FindStringSubmatch(cookie)
	if len(matches) < 2 {
		return "", fmt.Errorf("sessionid not found in cookie")
	}
	return SessionID(matches[1]), nil
}
