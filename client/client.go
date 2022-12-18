package client

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func GetDepth(port int) (int64, error) {
	url := fmt.Sprintf("http://localhost:%d", port)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("unexpected response code %d", resp.StatusCode)
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
