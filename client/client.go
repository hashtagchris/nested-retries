package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/cenkalti/backoff/v4"
)

const retries = 3

func GetDepth(ctx context.Context, port int) (int64, error) {
	boff := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), retries)
	boff = backoff.WithContext(boff, ctx)

	return backoff.RetryWithData(func() (int64, error) {
		return do(ctx, port)
	}, boff)
}

func do(ctx context.Context, port int) (int64, error) {
	url := fmt.Sprintf("http://localhost:%d", port)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	switch resp.StatusCode / 100 {
	case 2:
		// success!
	case 5:
		// retry on 5xx responses
		return 0, fmt.Errorf("unexpected response code %d", resp.StatusCode)
	default:
		// don't retry on 3xx and 4xx responses
		return 0, backoff.Permanent(fmt.Errorf("unexpected response code %d", resp.StatusCode))
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
