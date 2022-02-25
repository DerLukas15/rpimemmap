//go:build !arm
// +build !arm

package rpimemmap

import (
	"github.com/pkg/errors"
)

func vcOpen(size uint32) (int, error) {
	return 0, errors.New("Sorry no char device for unix at the moment. I couldn't make it work for now")
}

func vcClose() error {
	return nil
}
