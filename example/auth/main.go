// Copyright 2015 The go-fritzbox AUTHORS. All rights reserved.
//
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/philippfranke/go-fritzbox/fritzbox"
)

func main() {
	fmt.Printf("Connect to local FRITZ!Box! \n \n")

	// Ignore SSL certificates because it's self-signed
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	cl := &http.Client{Transport: tr}

	c := fritzbox.NewClient(cl)

	var username, password string

	fmt.Print("Enter username: ")
	fmt.Scan(&username)
	fmt.Println("\u2757  Caution: Reading password from STDIN with echoing \u2757")
	fmt.Print("Enter password: ")
	fmt.Scan(&password)

	if err := c.Auth(username, password); err != nil {
		log.Fatalf("Auth failed: %v", err)
	}

	fmt.Printf("Successfully logged in! \n \t Session ID: %s \n", c)
}
