// Copyright 2015 The go-fritzbox AUTHORS. All rights reserved.
//
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fritzbox

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

var testClient *Client

func init() {
	testClient = NewClient(nil)
}

func newTestServer(body string) (*httptest.Server, *url.URL) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, body)
	}))
	ts_url, _ := url.Parse(ts.URL)
	return ts, ts_url
}

func TestNewSession(t *testing.T) {
	s := newSession(testClient)
	if s.client != testClient {
		t.Errorf("New session: different clients: %v, got %v", testClient, s.client)
	}
}

func TestOpen(t *testing.T) {
	ts, ts_url := newTestServer(`
      <SessionInfo>
        <SID>0000000000000000</SID>
        <Challenge>1234567z</Challenge>
        <BlockTime>0</BlockTime>
        <Rights/>
      </SessionInfo>`)
	defer ts.Close()
	testClient.BaseURL = ts_url
	s := newSession(testClient)
	if err := s.Open(); err != nil {
		t.Errorf("Open session failed: %v", err)
	}

	expect := &Session{
		XMLName: xml.Name{
			Space: "",
			Local: "SessionInfo",
		},
		Sid:       "0000000000000000",
		Challenge: "1234567z",
		BlockTime: 0,
	}

	if expect.Sid != s.Sid {
		t.Errorf("Open session failed: %#v, got %#v", expect.Sid, s.Sid)
	}

	if expect.Challenge != s.Challenge {
		t.Errorf("Open session failed: %#v, got %#v", expect.Challenge, s.Challenge)
	}

}

var testAuths = []struct {
	Sid      string
	Password string
	Error    bool
}{
	{
		Sid:      "ff88e4d39354992f",
		Password: "äbc",
		Error:    false,
	},
	{
		Sid:      DefaultSid,
		Password: "äbz",
		Error:    true,
	},
}

func TestAuth(t *testing.T) {
	for _, auth := range testAuths {
		ts, ts_url := newTestServer(`
      <SessionInfo>
        <SID>` + auth.Sid + `</SID>
        <Challenge>1234567z</Challenge>
        <BlockTime>0</BlockTime>
        <Rights/>
      </SessionInfo>`)
		defer ts.Close()
		testClient.BaseURL = ts_url

		s := &Session{
			client:    testClient,
			Sid:       DefaultSid,
			Challenge: "1234567z",
			BlockTime: 0,
		}

		err := s.Auth("Username", auth.Password)
		if auth.Error {
			if err != ErrInvalidCred {
				t.Errorf("Auth didn't fail: %v, got %v", ErrInvalidCred, err)
			}
			if s.Sid != auth.Sid {
				t.Errorf("Auth session failed: %v, got %v", auth.Sid, s.Sid)
			}
		} else {
			if err != nil {
				t.Errorf("Auth failed: %v", err)
			}
			if s.Sid != auth.Sid {
				t.Errorf("Auth session failed: %v, got %v", auth.Sid, s.Sid)
			}
		}

	}
}

func TestClose(t *testing.T) {
	s := &Session{
		Sid:       "ff88e4d39354992f",
		Challenge: "1234567z",
		BlockTime: 0,
	}

	s.Close()
	if s.Sid != DefaultSid {
		t.Errorf("Session close failed: %s, got %s", DefaultSid, s.Sid)
	}
}

func TestIsExpired(t *testing.T) {
	s := &Session{
		Sid:       "ff88e4d39354992f",
		Challenge: "1234567z",
		BlockTime: 0,
		Expires:   time.Now().Add(time.Second * -5),
	}

	if !s.IsExpired() {
		t.Errorf("Session expires failed: %t, got %t", true, s.IsExpired())
	}

	s.Expires = time.Now().Add(time.Second * 5)

	if s.IsExpired() {
		t.Errorf("Session expires failed: %t, got %t", false, s.IsExpired())
	}
}

func TestRefresh(t *testing.T) {
	s := &Session{}
	s.Refresh()
	diff := time.Now().Add(DefaultExpires).Sub(s.Expires)
	if diff > (500 * time.Microsecond) {
		t.Errorf("Session refresh failed: %s, got %s, diff %s", time.Now().Add(DefaultExpires), s.Expires, diff)
	}
	s.Expires = time.Now().Add(time.Second * -5)
	if err := s.Refresh(); err == nil {
		t.Errorf("Session refresh failed: %v, got %v", ErrExpiredSess, err)
	}

}

func TestComputeResponse(t *testing.T) {
	r, err := computeResponse("1234567z", "äbc")
	if err != nil {
		t.Fatalf("computeResponse failed: %v", err)
	}

	expect := "1234567z-9e224a41eeefa284df7bb0f26c2913e2"

	if r != expect {
		t.Errorf("computeResponse failed: %s, got %s", expect, r)
	}
}
