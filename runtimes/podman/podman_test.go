package podman_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/uw-labs/podrick"
	_ "github.com/uw-labs/podrick/runtimes/podman" // Register auto-runtime
)

type jsonResp struct {
	Headers struct {
		Accept         string `json:"Accept"`
		AcceptEncoding string `json:"Accept-Encoding"`
		Host           string `json:"Host"`
		Origin         string `json:"Origin"`
		Referer        string `json:"Referer"`
		UserAgent      string `json:"User-Agent"`
	} `json:"headers"`
	Origin string `json:"origin"`
	URL    string `json:"url"`
}

func TestHTTPBin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	lc := func(address string) error {
		u := &url.URL{
			Scheme: "http",
			Host:   address,
			Path:   "get",
		}
		_, err := http.Get(u.String())
		return err
	}
	ctr, err := podrick.StartContainer(ctx, "docker.io/kennethreitz/httpbin", "latest", "80",
		podrick.WithLogger((*testLogger)(t)),
		podrick.WithLivenessCheck(lc),
	)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer func() {
		cErr := ctr.Close(context.Background())
		if cErr != nil {
			t.Fatal(cErr)
		}
	}()
	u := &url.URL{
		Scheme: "http",
		Host:   ctr.Address(),
		Path:   "get",
	}

	resp, err := http.Get(u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected StatusCode %d", resp.StatusCode)
	}

	var r jsonResp
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Fatal(err)
	}

	if r.URL != u.String() {
		t.Errorf("Unexpected request URL: got %q, wanted %q", r.URL, u.String())
	}
}

type testLogger testing.T

func (t *testLogger) Trace(msg string, fields ...map[string]interface{}) {
	msg = "] " + msg
	if len(fields) > 0 {
		for k, v := range fields[0] {
			msg = fmt.Sprintf(k+": %v, ", v) + msg
		}
	}
	msg = "TRACE [" + msg
	t.Logf(msg)
}

func (t *testLogger) Debug(msg string, fields ...map[string]interface{}) {
	msg = "] " + msg
	if len(fields) > 0 {
		for k, v := range fields[0] {
			msg = fmt.Sprintf(k+": %v, ", v) + msg
		}
	}
	msg = "DEBUG [" + msg
	t.Logf(msg)
}

func (t *testLogger) Info(msg string, fields ...map[string]interface{}) {
	msg = "] " + msg
	if len(fields) > 0 {
		for k, v := range fields[0] {
			msg = fmt.Sprintf(k+": %v, ", v) + msg
		}
	}
	msg = "INFO [" + msg
	t.Logf(msg)
}

func (t *testLogger) Warn(msg string, fields ...map[string]interface{}) {
	msg = "] " + msg
	if len(fields) > 0 {
		for k, v := range fields[0] {
			msg = fmt.Sprintf(k+": %v, ", v) + msg
		}
	}
	msg = "WARN [" + msg
	t.Logf(msg)
}

func (t *testLogger) Error(msg string, fields ...map[string]interface{}) {
	msg = "] " + msg
	if len(fields) > 0 {
		for k, v := range fields[0] {
			msg = fmt.Sprintf(k+": %v, ", v) + msg
		}
	}
	msg = "ERROR [" + msg
	t.Logf(msg)
}
