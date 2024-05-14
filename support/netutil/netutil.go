package netutil

import (
	"errors"
	"net"
	"strconv"
	"syscall"
	"time"
)

func IsTCPAddrServing(url string, timeout time.Duration) (bool, error) {
	conn, err := net.DialTimeout("tcp", url, timeout)
	if err != nil {
		return false, nil
	}
	conn.Close()
	return true, nil
}

func ServePort(port int, autoIncrPort bool, watchTimeout time.Duration, watch func(port int), doWithPort func(port int) error) error {
	for {
		// open url after 500ms, waiting for port opening to check if error
		portErr := make(chan struct{})
		if watchTimeout > 0 && watch != nil {
			go watchSignalWithinTimeout(watchTimeout, portErr, func() {
				watch(port)
			})
		}

		serving, err := IsTCPAddrServing(net.JoinHostPort("localhost", strconv.Itoa(port)), 20*time.Millisecond)
		if err != nil {
			return err
		}
		if serving {
			continue
		}
		err = doWithPort(port)
		if err == nil {
			return nil
		}
		close(portErr)
		var syscallErr syscall.Errno
		if autoIncrPort && errors.As(err, &syscallErr) && syscallErr == syscall.EADDRINUSE {
			port++
			continue
		}
		return err
	}
}

// executing action
func watchSignalWithinTimeout(timeout time.Duration, errSignal chan struct{}, action func()) {
	select {
	case <-time.After(timeout):
	case <-errSignal:
		return
	}
	action()
}
