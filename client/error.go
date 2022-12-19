package client

import "fmt"

type ResponseCodeError struct {
	ResponseCode int
	StatusRange  int
}

func (r ResponseCodeError) Error() string {
	return fmt.Sprintf("Unexpected response code %d", r.ResponseCode)
}
