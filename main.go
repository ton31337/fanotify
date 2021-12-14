package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/sys/unix"
)

type FanotifyEventMetadata struct {
	Len         uint32
	Version     uint8
	Reserved    uint8
	MetadataLen uint16
	Mask        uint64
	Fd          int32
	Pid         int32
}

type FanotifyResponse struct {
	Fd       int32
	Response uint32
}

type FanotifyMetadata struct {
	File          *os.File
	Data          *bufio.Reader
	EventMetadata FanotifyEventMetadata
}

func (fem *FanotifyEventMetadata) GetPath() (string, error) {
	path, err := os.Readlink(
		filepath.Join(
			"/proc/self/fd",
			strconv.FormatUint(uint64(fem.Fd), 10),
		),
	)

	if err != nil {
		return "", err
	}

	return path, nil
}

func (fem *FanotifyEventMetadata) Response() FanotifyResponse {
	return FanotifyResponse{
		Fd:       fem.Fd,
		Response: unix.FAN_ALLOW,
	}
}

func FanotifyRead(fd int) (*FanotifyMetadata, error) {
	meta := FanotifyEventMetadata{}

	file := os.NewFile(uintptr(fd), "")
	if file == nil {
		return nil, errors.New("failed creating temporary file")
	}

	buf := bufio.NewReader(file)
	if buf == nil {
		return nil, errors.New("failed creating buffer")
	}

	err := binary.Read(buf, binary.LittleEndian, &meta)
	if err != nil {
		return nil, err
	}

	return &FanotifyMetadata{
		File:          file,
		Data:          buf,
		EventMetadata: meta,
	}, nil
}

func FanotifyPoll(fd int, stopFirst bool, callback func(data interface{})) error {
	for {
		pfd := []unix.PollFd{
			{
				Fd:     0,
				Events: unix.POLLIN,
			},
			{
				Fd:     int32(fd),
				Events: unix.POLLIN,
			},
		}

		_, err := unix.Poll(pfd, -1)
		if err != nil {
			return err
		}

		if pfd[1].Revents == unix.POLLIN {
			fm, err := FanotifyRead(fd)
			if err != nil {
				return err
			}

			path, err := fm.EventMetadata.GetPath()
			if err != nil {
				return err
			}

			callback(path)

			binary.Write(fm.File,
				binary.LittleEndian,
				fm.EventMetadata.Response())

			if stopFirst {
				return err
			}
		}
	}
}
