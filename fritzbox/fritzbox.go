// Copyright 2015 The go-fritzbox AUTHORS. All rights reserved.
//
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fritzbox

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultBaseURL = "http://fritz.box/"
)

type Client struct {
	client  *http.Client
	BaseURL *url.URL

	session *Session
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		client:  httpClient,
		BaseURL: baseURL,
	}

	return c
}

// Auth authenticates with username and password
func (c *Client) Auth(username, password string) error {
	s := newSession(c)

	if err := s.Open(); err != nil {
		return err
	}

	if err := s.Auth(username, password); err != nil {
		return err
	}

	c.session = s
	return nil
}

// Session returns the current SessionId or Sid
func (c *Client) Session() string {
	return c.session.Sid
}

func (c *Client) NewRequest(method, urlStr string,
	data url.Values) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	if c.session != nil {
		values := u.Query()
		values.Set("sid", c.session.Sid)
		u.RawQuery = values.Encode()
	}

	var buf io.Reader
	if data != nil {
		buf = strings.NewReader(data.Encode())
	}
	req, err := http.NewRequest(method, u.String(), buf)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	if c.session != nil {
		if err := c.session.Refresh(); err != nil {
			return nil, err
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if v != nil {
		err = xml.NewDecoder(resp.Body).Decode(v)
	}

	return resp, err
}
