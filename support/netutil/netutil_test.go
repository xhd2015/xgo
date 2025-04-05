package netutil

import (
	"fmt"
	"testing"
)

func TestGetLocalHostByIPType(t *testing.T) {

	tests := []struct {
		name   string
		ipType string
		want   bool
	}{
		{name: "Test with all IPs", ipType: "all", want: true},
		{name: "Test with IPv4", ipType: "ipv4", want: true},
		{name: "Test unknown ipType", ipType: "test", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ips, err := GetLocalHostByIPType(tt.ipType)
			if err != nil {
				t.Fatalf("get local host ip: %v", err)
			}
			if got := len(ips); got > 0 != tt.want {
				t.Errorf("GetLocalHostByIPType(%v) = %v; want %v", tt.ipType, got, tt.want)
			}
		})
	}
}

func TestGetHostAndIP(t *testing.T) {
	ipv4IPs, err := GetLocalHostByIPType("ipv4")
	if err != nil {
		t.Fatalf("get local host ip: %v", err)
	}
	var tests = []struct {
		bindStr  string
		portStr  string
		wantHost string
		wantPort int
	}{
		{"", "", "localhost", 7070},            // default host and port
		{ipv4IPs[0], "", ipv4IPs[0], 7070},     // provide host
		{"", "8080", "localhost", 8080},        // provide port
		{ipv4IPs[0], "8080", ipv4IPs[0], 8080}, // provide host and port
	}

	for _, tt := range tests {
		host, port := GetHostAndIP(tt.bindStr, tt.portStr)
		if host != tt.wantHost || port != tt.wantPort {
			t.Errorf("GetHostAndIP(%v, %v) => (%v, %v), want (%v, %v)", tt.bindStr, tt.portStr, host, port, tt.wantHost, tt.wantPort)
		}
	}
}

func TestGetURLToOpen(t *testing.T) {
	ipv4IPs, err := GetLocalHostByIPType("ipv4")
	if err != nil {
		t.Fatalf("get local host ip: %v", err)
	}
	url := fmt.Sprintf("http://%s:7070", ipv4IPs[0])
	var tests = []struct {
		host    string
		port    int
		wantURL string
	}{
		{"localhost", 7070, "http://localhost:7070"},
		{"127.0.0.1", 7070, "http://127.0.0.1:7070"},
		{"0.0.0.0", 7070, "http://127.0.0.1:7070"},
		{ipv4IPs[0], 7070, url},
		{"::1", 7070, "http://::1:7070"},
	}

	for _, tt := range tests {
		gotURL, _ := GetURLToOpen(tt.host, tt.port)
		if gotURL != tt.wantURL {
			t.Errorf("BuildAndDisplayURL(%v, %v) => %v, want %v", tt.host, tt.port, gotURL, tt.wantURL)
		}
	}
}
