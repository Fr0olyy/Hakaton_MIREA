package services

import (
	"errors"
	"fmt"
)

var ErrBadRequest = errors.New("bad request")

func badRequest(message string) error {
	return fmt.Errorf("%w: %s", ErrBadRequest, message)
}

func IsBadRequest(err error) bool {
	return errors.Is(err, ErrBadRequest)
}
