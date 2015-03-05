# go-fritzbox#

go-fritzbox is a Go client libary for accessing a FRITZ!Box (>=FRITZ!OS 5.50)

**Documentation:** [![GoDoc](https://godoc.org/github.com/philippfranke/go-fritzbox/fritzbox?status.svg)](https://godoc.org/github.com/philippfranke/go-fritzbox/fritzbox)

**Build Status:** [![Build Status](https://travis-ci.org/philippfranke/go-fritzbox.svg?branch=master)](https://travis-ci.org/philippfranke/go-fritzbox)

go-fritzbox requires Go version 1.1 or greater.

## Usage ##
```go
import "github.com/philippfranke/go-fritzbox/fritzbox
```

Construct a new FRITZ!Box client, then use auth method in order to log in.
For example, to access Fritz!Box as user "Peter":

```go
client := fritzbox.NewClient(nil)
err := client.Auth("Peter", "Passw0rD!")
```

## Access remote FRITZ!Box over SSL ##
The recommended way to access a FRITZ!Box over SSL is using a valid SSL
certificate, but you can always skip validation. See [http docs][] for complete
instruction on using `http.Client`.

```go
url, _ := url.Parse("https://example.com")
// !!! Not recommended !!!
tr := &http.Transport{
  TLSClientConfig: &tls.Config{
    InsecureSkipVerify: true,
  },
}
c := &http.Client{Transport: tr}

client := fritzbox.NewClient(cl)
client.BaseURL = url

// Login
err := client.Auth("Peter", "Passw0rD!")
```

## License ##
This library is distributed under the MIT-style license found in the [LICENSE](./LICENSE)
file.


[http docs]: http://golang.org/pkg/net/http/
