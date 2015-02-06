// Copyright 2015 The go-fritzbox AUTHORS. All rights reserved.
//
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fritzbox

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the FRITZ!Box client being tested.
	client *Client

	// server is a test HTTP server used to provide mock responses.
	server *httptest.Server
)

// setup sets up a test HTTP Server along with a fritzbox.Client that  is
// configured to talk to the that test server. Tests should register handlers
// on mux which provide mock responses.
func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewClient(nil)
	client.session = NewSession(client)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

// teardown closes the test server
func teardown() {
	server.Close()
}

func TestNewClient(t *testing.T) {
	c := NewClient(nil)
	if got, want := c.BaseURL.String(), defaultBaseURL; got != want {
		t.Errorf("NewClient BaseURL is %v, want %v", got, want)
	}
}

func TestNewRequest(t *testing.T) {
	c := NewClient(nil)
	c.session = NewSession(c)

	inSession, outSession := "abc", "?sid=abc"
	c.session.Sid = inSession

	inURL, outURL := "/test", defaultBaseURL+"test"+outSession
	inData, outData := url.Values{"test": {"test"}}, "test=test"

	req, _ := c.NewRequest("GET", inURL, inData)
	if got, want := req.URL.String(), outURL; got != want {
		t.Errorf("NewRequest(%q) URL is %v, want %v", inURL, got, want)
	}
	body, _ := ioutil.ReadAll(req.Body)
	if got, want := string(body), outData; got != want {
		t.Errorf("NewRequest(%q) Query is %v, want %v", inData, got, want)
	}
}

func TestNewRequest_invalidURL(t *testing.T) {
	c := NewClient(nil)
	inURL := "%"
	_, err := c.NewRequest("GET", inURL, nil)
	if err == nil {
		t.Error("Expected error to be returned")
	}
}

func TestNewRequest_invalidHTTPRequest(t *testing.T) {
	c := NewClient(nil)
	inURL := "%"
	_, err := c.NewRequest("GET", inURL, nil)
	if err == nil {
		t.Error("Expected error to be returned")
	}
}

func TestDo_JSON(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := "GET"; m != r.Method {
			t.Errorf("Request method = %v, want %v", r.Method, m)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8;")
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	body := new(foo)
	client.Do(req, body)
	want := &foo{"a"}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("Response body = %v, want %v", body, want)
	}
}

func TestDo_XML(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := "GET"; m != r.Method {
			t.Errorf("Request method = %v, want %v", r.Method, m)
		}
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, `<foo><A>a</A></foo>`)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	body := new(foo)
	client.Do(req, body)
	want := &foo{"a"}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("Response body = %v, want %v", body, want)
	}
}

func TestDo_SessionExpired(t *testing.T) {
	setup()
	defer teardown()
	client.session.Expires = time.Now()
	req, _ := client.NewRequest("GET", "/", nil)
	_, err := client.Do(req, nil)
	if err == nil {
		t.Error("Expected error to be returned")
	}
}

func TestDo_HTTPError(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	_, err := client.Do(req, nil)

	if err == nil {
		t.Error("Expected HTTP 400 error.")
	}
}

func TestClientAuth(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, `
		<SessionInfo>
      <SID>ff88e4d39354992f</SID>
      <Challenge>1234567z</Challenge>
    </SessionInfo>`)
	})
	client.session = &Session{
		client: client,
		Sid:    DefaultSid,
	}

	err := client.Auth("username", "äbc")

	if err != nil {
		t.Errorf("Auth unexpected error %v", err)
	}

	want := "ff88e4d39354992f"

	if client.session.Sid != want {
		t.Errorf("Auth sid is %s, want %s", client.session.Sid, want)
	}
}

func TestClientAuth_Invalid(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, `
		<SessionInfo>
      <SID>`+DefaultSid+`</SID>
      <Challenge>1234567z</Challenge>
    </SessionInfo>`)
	})
	client.session = &Session{
		client: client,
		Sid:    DefaultSid,
	}
	err := client.Auth("username", "äbc")
	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	want := DefaultSid

	if client.session.Sid != want {
		t.Errorf("Auth sid is %s, want %s", client.session.Sid, want)
	}
}

func TestClientAuth_NewSession(t *testing.T) {
	c := NewClient(nil)
	c.Auth("username", "äbc")
	if c.session.Sid != DefaultSid {
		t.Errorf("Auth sid is %s, want %s", c.session.Sid, DefaultSid)
	}
}

func TestClientClose(t *testing.T) {
	c := &Client{
		session: &Session{
			Sid: "abc",
		},
	}

	c.Close()

	if c.session.Sid != DefaultSid {
		t.Errorf("Close sid is %s, want %v", c.session.Sid, DefaultSid)
	}
}
