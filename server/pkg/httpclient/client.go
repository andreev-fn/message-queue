package httpclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"server/pkg/httpmodels"
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	baseURL  string
	httpDoer HTTPDoer
}

func NewClient(baseURL string, httpDoer HTTPDoer) *Client {
	if httpDoer == nil {
		httpDoer = &http.Client{
			Timeout: time.Second * 5,
		}
	}
	return &Client{
		baseURL:  baseURL,
		httpDoer: httpDoer,
	}
}

func (c *Client) PrepareMessages(reqDTO httpmodels.PublishRequest) (httpmodels.PublishResponse, error) {
	var respDTO httpmodels.PublishResponse

	if err := c.doRequest("/messages/prepare", reqDTO, &respDTO); err != nil {
		return nil, err
	}

	return respDTO, nil
}

func (c *Client) PublishMessages(reqDTO httpmodels.PublishRequest) (httpmodels.PublishResponse, error) {
	var respDTO httpmodels.PublishResponse

	if err := c.doRequest("/messages/publish", reqDTO, &respDTO); err != nil {
		return nil, err
	}

	return respDTO, nil
}

func (c *Client) ReleaseMessages(reqDTO httpmodels.ReleaseRequest) error {
	var respDTO httpmodels.OkResponse

	if err := c.doRequest("/messages/release", reqDTO, &respDTO); err != nil {
		return err
	}

	return c.checkOkResponse(respDTO)
}

func (c *Client) CheckMessages(reqDTO httpmodels.CheckRequest) (httpmodels.CheckResponse, error) {
	var respDTO httpmodels.CheckResponse

	if err := c.doRequest("/messages/check", reqDTO, &respDTO); err != nil {
		return nil, err
	}

	return respDTO, nil
}

func (c *Client) ConsumeMessages(reqDTO httpmodels.ConsumeRequest) (httpmodels.ConsumeResponse, error) {
	var respDTO httpmodels.ConsumeResponse

	if err := c.doRequest("/messages/consume", reqDTO, &respDTO); err != nil {
		return nil, err
	}

	return respDTO, nil
}

func (c *Client) AckMessages(reqDTO httpmodels.AckRequest) error {
	var respDTO httpmodels.OkResponse

	if err := c.doRequest("/messages/ack", reqDTO, &respDTO); err != nil {
		return err
	}

	return c.checkOkResponse(respDTO)
}

func (c *Client) NackMessages(reqDTO httpmodels.NackRequest) error {
	var respDTO httpmodels.OkResponse

	if err := c.doRequest("/messages/nack", reqDTO, &respDTO); err != nil {
		return err
	}

	return c.checkOkResponse(respDTO)
}

func (c *Client) RedirectMessages(reqDTO httpmodels.RedirectRequest) error {
	var respDTO httpmodels.OkResponse

	if err := c.doRequest("/messages/redirect", reqDTO, &respDTO); err != nil {
		return err
	}

	return c.checkOkResponse(respDTO)
}

func (c *Client) doRequest(method string, reqDTO any, respDTO any) error {
	body, err := json.Marshal(reqDTO)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	fullURL, err := url.JoinPath(c.baseURL, method)
	if err != nil {
		return fmt.Errorf("url.JoinPath: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("http.NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpDoer.Do(req)
	if err != nil {
		return fmt.Errorf("httpDoer.Do: %w", err)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		return fmt.Errorf("unexpected content type: %s; body: %s", contentType, string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		var errDTO httpmodels.ErrorResponse
		err = json.Unmarshal(respBody, &errDTO)
		if err != nil {
			return fmt.Errorf("json.Unmarshal: %w; body: %s", err, string(respBody))
		}

		return errors.New(errDTO.Error)
	}

	if err = json.Unmarshal(respBody, &respDTO); err != nil {
		return fmt.Errorf("json.Unmarshal: %w; body: %s", err, string(respBody))
	}

	return nil
}

func (c *Client) checkOkResponse(respDTO httpmodels.OkResponse) error {
	if !respDTO.Ok {
		return errors.New("malformed ok response")
	}

	return nil
}
