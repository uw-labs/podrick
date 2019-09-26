package podman_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/cenkalti/backoff"

	"github.com/uw-labs/podrick"
	// Register auto-runtime
	_ "github.com/uw-labs/podrick/runtimes/podman"
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
	ctr, err := podrick.StartContainer("docker.io/kennethreitz/httpbin", "latest", "80",
		podrick.WithLogger((*testLogger)(t)),
	)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer func() {
		cErr := ctr.Close()
		if cErr != nil {
			t.Fatal(cErr)
		}
	}()
	u := &url.URL{
		Scheme: "http",
		Host:   ctr.Address(),
		Path:   "get",
	}

	bk := backoff.NewExponentialBackOff()
	bk.MaxElapsedTime = 10 * time.Second
	err = backoff.RetryNotify(
		func() error {
			_, err := http.Get(u.String())
			return err
		},
		bk,
		func(err error, next time.Duration) {
			t.Logf("Failed to connect to container, retrying in %s", next.Truncate(time.Millisecond))
		},
	)
	if err != nil {
		t.Fatal(err)
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
