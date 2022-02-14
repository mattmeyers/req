package req

import (
	"bytes"
	"net/http"
)

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{},
	}
}

func (c *Client) Do(req Request) (*http.Request, *http.Response, error) {
	httpReq, err := http.NewRequest(req.Method, req.Path, bytes.NewBufferString(req.Body))
	if err != nil {
		return nil, nil, err
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	res, err := c.client.Do(httpReq)
	if err != nil {
		return nil, nil, err
	}

	return httpReq, res, nil
}
