package netutil

import (
	"errors"
	"fmt"
	"net"
	"net/http"
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

func ServePortHTTP(server *http.ServeMux, host string, port int, autoIncrPort bool, watchTimeout time.Duration, watch func(port int)) error {
	return ServePort(host, port, autoIncrPort, watchTimeout, watch, func(port int) error {
		return http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), server)
	})
}

// suggested watch timeout: 500ms
func ServePort(host string, port int, autoIncrPort bool, watchTimeout time.Duration, watch func(port int), doWithPort func(port int) error) error {
	for {
		addr := net.JoinHostPort(host, strconv.Itoa(port))
		serving, err := IsTCPAddrServing(addr, 20*time.Millisecond)
		if err != nil {
			return err
		}
		if serving {
			if !autoIncrPort {
				return fmt.Errorf("bind %s failed: address in use", addr)
			}
			port++
			continue
		}

		// open url after 500ms, waiting for port opening to check if error
		portErr := make(chan struct{})
		if watchTimeout > 0 && watch != nil {
			go watchSignalWithinTimeout(watchTimeout, portErr, func() {
				watch(port)
			})
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

// get the local area network IP address by ipType
func GetLocalHostByIPType(ipType string) []string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("\nnet.InterfaceAddrs failed, err: %v\n", err.Error())
		return nil
	}
	// ipv4 + ipv6
	var ips []string
	for _, address := range addresses {
		if ipNet, ok := address.(*net.IPNet); ok {
			// IPv4
			if ipNet.IP.To4() != nil && (ipType == "all" || ipType == "ipv4") {
				ips = append(ips, ipNet.IP.String())
			}
			// IPv6
			if ipNet.IP.To16() != nil && ipNet.IP.To4() == nil && (ipType == "all" || ipType == "ipv6") {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}
	if len(ips) == 0 {
		fmt.Println("no network interface found")
		return nil
	}
	return ips
}

// provide default values for host and port
func GetHostAndIP(bindStr string, portStr string) (host string, port int) {
	// default host=localhost, ipv4 127.0.0.1 or ipv6 ::1
	host = "localhost"
	port = 7070

	if portStr != "" {
		parsePort, err := strconv.ParseInt(portStr, 10, 64)
		if err == nil {
			port = int(parsePort)
		}
	}

	localIPs := GetLocalHostByIPType("all")
	if localIPs == nil {
		return
	}
	// add default router
	localIPs = append(localIPs, []string{"0.0.0.0", "::"}...)
	for _, localIP := range localIPs {
		if localIP == bindStr {
			host = bindStr
			break
		}
	}
	// parse ipv6
	ip := net.ParseIP(host)
	if ip != nil {
		// IP is valid, check if it is IPv4 or IPv6
		if ip.To4() == nil {
			// ip is not a v4 addr, must be v6
			host = fmt.Sprintf("[%s]", host)
		}
	}
	return
}

// build URL based on host and port
func BuildAndDisplayURL(host string, port int) string {
	url := fmt.Sprintf("http://%s:%d", host, port)
	fmt.Println("Server listen at:")
	fmt.Printf("-  Open: %s \r\n", url)
	fmt.Printf("-  Local: http://localhost:%d \r\n", port)
	// only print ipv4
	if host == "0.0.0.0" {
		ipv4IPs := GetLocalHostByIPType("ipv4")
		for _, ipv4IP := range ipv4IPs {
			fmt.Printf("-  Network: http://%s:%d \r\n", ipv4IP, port)
		}
	}
	return url
}
