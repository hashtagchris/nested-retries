package client

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/cenkalti/backoff/v4"
)

const retries = 3

func GetDepth(port int) (int64, error) {
	boff := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), retries)

	return backoff.RetryWithData(func() (int64, error) {
		return do(port)
	}, boff)
}

func do(port int) (int64, error) {
	url := fmt.Sprintf("http://localhost:%d", port)

	resp, err := http.Get(url)
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
