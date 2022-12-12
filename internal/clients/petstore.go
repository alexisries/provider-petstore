package petstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ResourceNotFoundException struct {
	Message *string
}

func (e *ResourceNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ResourceNotFoundException) ErrorCode() string { return "ResourceNotFoundException" }
func (e *ResourceNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}

func errorFromStatusCode(statusCode int, message string) (err error) {
	switch statusCode {
	case 404:
		err = &ResourceNotFoundException{Message: &message}
	default:
		err = errors.New(message)
	}
	return err
}

func IsErrorNotFound(err error) bool {
	var notFoundError *ResourceNotFoundException
	return errors.As(err, &notFoundError)
}

type Config struct {
	server string
}

func GetConfig(url string) *Config {
	return &Config{
		server: url,
	}
}

type Client struct {
	config  *Config
	context context.Context
}

func New(cfg *Config) *Client {
	return &Client{
		config:  cfg,
		context: context.TODO(),
	}
}

func (c *Client) prepareRequest(path string, method string, body []byte) (*http.Request, error) {
	queryURL := c.config.server + path
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(c.context, method, queryURL, bodyReader)
	if err != nil {
		return nil, err
	}

	switch method {
	case "GET":
		req.Header.Set("Accept", "application/json")
	case "PUT", "POST":
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) DoRequest(path string, method string, body []byte) (*http.Response, error) {
	req, err := c.prepareRequest(path, method, body)
	if err != nil {
		return nil, err
	}
	fmt.Println(req.URL)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		var errMsg string
		var err error
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)
		if err == nil {
			errMsg = string(resBody)
		}
		err = errorFromStatusCode(res.StatusCode, errMsg)
		return nil, err
	}
	return res, nil
}

// String returns a pointer value for the string value passed in.
func String(v string) *string {
	return &v
}

// Int64 returns a pointer value for the int64 value passed in.
func Int64(v int64) *int64 {
	return &v
}
