package netutil

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"sort"
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

func FindListenablePort(host string, port int) (int, error) {
	for {
		addr := net.JoinHostPort(host, strconv.Itoa(port))
		serving, err := IsTCPAddrServing(addr, 20*time.Millisecond)
		if err != nil {
			return 0, err
		}
		if serving {
			port++
			continue
		}

		return port, nil
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
func GetLocalHostByIPType(ipType string) ([]string, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
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
	sort.Slice(ips, func(i, j int) bool {
		if ips[i] == "127.0.0.1" {
			return true
		}
		if ips[j] == "127.0.0.1" {
			return false
		}
		return ips[i] < ips[j]
	})
	return ips, nil
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

	localIPs, err := GetLocalHostByIPType("all")
	if err != nil || localIPs == nil {
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

// get URL based on host and port
func GetURLToOpen(host string, port int) (primary string, extra []string) {
	addr := host
	if host == "0.0.0.0" {
		// only get ipv4
		ipv4IPs, err := GetLocalHostByIPType("ipv4")
		if err == nil && len(ipv4IPs) > 0 {
			addr = ipv4IPs[0]
			for i := 1; i < len(ipv4IPs); i++ {
				extra = append(extra, fmt.Sprintf("http://%s:%d", ipv4IPs[i], port))
			}
		}
	}
	primary = fmt.Sprintf("http://%s:%d", addr, port)
	return
}

func PrintUrls(url string, extra ...string) {
	if len(extra) == 0 {
		fmt.Printf("Server listen at %s\n", url)
		return
	}
	if len(extra) > 0 {
		fmt.Printf("Server listen at:\n")
		fmt.Printf("  %s\n", url)
		for _, remain := range extra {
			fmt.Printf("  %s\n", remain)
		}
		return
	}
}
