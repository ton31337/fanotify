package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

func FanotifyCallBack(data interface{}) {
	// dummy callback function
	fmt.Printf("Data: %s\n", data)
}

func TestFanotify(t *testing.T) {
	fd, _ := unix.FanotifyInit(unix.FAN_CLOEXEC|
		unix.FAN_CLASS_CONTENT|
		unix.FAN_NONBLOCK,
		uint(os.O_RDONLY|
			unix.O_LARGEFILE))

	unix.FanotifyMark(fd,
		unix.FAN_MARK_ADD|unix.FAN_MARK_MOUNT,
		unix.FAN_MODIFY|unix.FAN_CLOSE_WRITE,
		unix.AT_FDCWD,
		"/tmp")

	err := FanotifyPoll(fd, true, FanotifyCallBack)
	assert.NoError(t, err)
}
