package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
)

const retries = 3
const timeout = 10 * time.Second

func GetDepth(ctx context.Context, port int) (int64, error) {
	exp := backoff.NewExponentialBackOff()
	boff := backoff.WithMaxRetries(exp, retries)
	boff = backoff.WithContext(boff, ctx)

	return backoff.RetryNotifyWithData(func() (int64, error) {
		return do(ctx, port)
	}, boff, notify)
}

func do(ctx context.Context, port int) (int64, error) {
	url := fmt.Sprintf("http://localhost:%d", port)

	// Use a 10 second timeout for the http request
	// Another option is to set the response header timeout on the transport
	ctx, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	statusRange := resp.StatusCode / 100
	if statusRange != 2 {
		err := ResponseCodeError{resp.StatusCode, statusRange}
		if statusRange == 5 {
			return 0, err
		} else {
			return 0, backoff.Permanent(err)
		}
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	depth, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		return 0, err
	}

	return depth, nil
}

func notify(err error, dur time.Duration) {
	// log.Printf("Error: %s, Duration: %s", err, dur)
}
