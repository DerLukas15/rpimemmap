package rpimemmap

import (
	"syscall"

	"github.com/pkg/errors"
)

func vcOpen(size uint32) (int, error) {
	fileDesc, err := syscall.Open("/dev/vcio", syscall.O_RDONLY, 0)
	if err != nil {
		return 0, errors.Wrap(err, "mbox open")
	}
	return fileDesc, nil
}

func vcClose(handle int) error {
	err := syscall.Close(handle)
	if err != nil {
		return errors.Wrap(err, "mbox close")
	}
	return nil
}
