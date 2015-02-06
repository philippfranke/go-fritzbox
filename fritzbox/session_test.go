// Copyright 2015 The go-fritzbox AUTHORS. All rights reserved.
//
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fritzbox

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	want := NewClient(nil)
	s := NewSession(want)
	if s.client != want {
		t.Errorf("newSession Client is %v, want %v", s.client, want)
	}
}

func TestSessionOpen(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/login_sid.lua", func(w http.ResponseWriter, r *http.Request) {
		if m := "GET"; m != r.Method {
			t.Errorf("Request method = %v, want %v", r.Method, m)
		}
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, `
			<SessionInfo>
        <SID>0000000000000000</SID>
        <Challenge>1234567z</Challenge>
        <BlockTime>0</BlockTime>
        <Rights/>
      </SessionInfo>`)
	})

	s := NewSession(client)

	s.Open()

	want := &Session{
		Sid:       "0000000000000000",
		Challenge: "1234567z",
	}

	if want.Sid != s.Sid {
		t.Errorf("OpenSession Sid is %s, want %s", s.Sid, want.Sid)
	}

	if want.Challenge != s.Challenge {
		t.Errorf("OpenSession Challenge is %s, want %#v", s.Challenge, want.Challenge)
	}
}

// Auth test cases
var testsAuth = []struct {
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

func TestSessionAuth(t *testing.T) {

	for _, want := range testsAuth {
		setup()
		defer teardown()

		mux.HandleFunc("/login_sid.lua", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/xml")
			fmt.Fprint(w, `
			<SessionInfo>
        <SID>`+want.Sid+`</SID>
        <Challenge>1234567z</Challenge>
        <BlockTime>0</BlockTime>
        <Rights/>
      </SessionInfo>`)
		})

		s := &Session{
			client:    client,
			Sid:       DefaultSid,
			Challenge: "1234567z",
			BlockTime: 0,
		}

		err := s.Auth("Username", want.Password)
		if want.Error {
			if err != ErrInvalidCred {
				t.Error("Expected error to be returned")
			}
			if s.Sid != want.Sid {
				t.Errorf("SessionAuth Sid is %s, want %s", s.Sid, want.Sid)
			}
		} else {
			if err != nil {
				t.Errorf("SessionAuth unexpected error %v", err)
			}
			if s.Sid != want.Sid {
				t.Errorf("SessionAuth Sid is %s, want %s", s.Sid, want.Sid)
			}
		}

	}
}

func TestSessionClose(t *testing.T) {
	s := &Session{
		Sid:       "ff88e4d39354992f",
		Challenge: "1234567z",
		BlockTime: 0,
	}

	s.Close()
	if s.Sid != DefaultSid {
		t.Errorf("SessionClose Sid is %s, want %s", s.Sid, DefaultSid)
	}
}

func TestIsExpired(t *testing.T) {
	// TODO: better test
	s := &Session{
		Sid:       "ff88e4d39354992f",
		Challenge: "1234567z",
		BlockTime: 0,
		Expires:   time.Now().Add(time.Second * -5),
	}

	if !s.IsExpired() {
		t.Errorf("SessionExpires isExpired is %t, want %t", s.IsExpired(), true)
	}

	s.Expires = time.Now().Add(time.Second * 5)

	if s.IsExpired() {
		t.Errorf("SessionExpires isExpired is %t, want %t", s.IsExpired(), false)
	}
}

func TestRefresh(t *testing.T) {
	s := &Session{}
	s.Refresh()
	diff := time.Now().Add(DefaultExpires).Sub(s.Expires)
	if diff > (500 * time.Microsecond) {
		t.Errorf("SessionRefresh Expires is %s, want %s, diff %s", s.Expires, time.Now().Add(DefaultExpires), diff)
	}
	s.Expires = time.Now().Add(time.Second * -5)
	if err := s.Refresh(); err == nil {
		t.Error("Expected error to be returned")
	}

}

// Responses test cases
var testsChallenge = []struct {
	Challenge string
	Password  string
	Want      string
}{
	{
		Challenge: "1234567z",
		Password:  "äbc",
		Want:      "1234567z-9e224a41eeefa284df7bb0f26c2913e2",
	},
	{
		Challenge: "1234567z",
		Password:  "äbz€",
		Want:      "1234567z-eefe7c82f8f122671950682d0be94e52",
	},
}

func TestComputeResponse(t *testing.T) {
	for _, c := range testsChallenge {
		r, err := computeResponse(c.Challenge, c.Password)
		if err != nil {
			t.Fatalf("computeResponse unexpected error %v", err)
		}

		if r != c.Want {
			t.Errorf("computeResponse response is %s, want %s", r, c.Want)
		}
	}
}
