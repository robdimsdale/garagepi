package test_helpers

import (
	"errors"
	"io"
)

type ErrCloser struct {
	io.Reader
}

func (e ErrCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("ReadError")
}

func (e ErrCloser) Close() error {
	return nil
}
