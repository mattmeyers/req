package req

import (
	"bytes"
	"fmt"
	"io"
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

func (c *Client) Do(req Request) error {
	url := fmt.Sprintf("http://localhost:8080%s", req.Path)
	httpReq, err := http.NewRequest(req.Method, url, bytes.NewBufferString(req.Body))
	if err != nil {
		return err
	}

	res, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", resBody)

	return nil
}
