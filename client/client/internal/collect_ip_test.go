//go:build windows || linux || darwin

package internal

import (
	"net"
	"testing"
)

func TestIsEligibleIP(t *testing.T) {
	cases := []struct {
		name string
		ip   string
		want bool
	}{
		{"ipv4", "192.168.1.10", true},
		{"ipv4 loopback", "127.0.0.1", false},
		{"ipv6 global", "2001:db8::1", true},
		{"ipv6 link-local", "fe80::1", false},
		{"ipv6 loopback", "::1", false},
		{"ipv6 ula", "fd00::1", true},
		{"nil", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var ip net.IP
			if tc.ip != "" {
				ip = net.ParseIP(tc.ip)
			}
			if got := isEligibleIP(ip); got != tc.want {
				t.Fatalf("isEligibleIP(%q)=%v want %v", tc.ip, got, tc.want)
			}
		})
	}
}

func TestHasEligibleIP(t *testing.T) {
	t.Run("prefer ipv4 over ipv6", func(t *testing.T) {
		ip, ok := hasEligibleIP([]net.Addr{
			createMockIPv6Addr("2001:db8::1"),
			createMockIPv4Addr("10.0.0.5"),
		})
		if !ok || ip != "10.0.0.5" {
			t.Fatalf("got %q ok=%v", ip, ok)
		}
	})
	t.Run("ipv6 only global", func(t *testing.T) {
		ip, ok := hasEligibleIP([]net.Addr{
			createMockIPv6Addr("fe80::1"),
			createMockIPv6Addr("2001:db8::9"),
		})
		if !ok || ip != "2001:db8::9" {
			t.Fatalf("got %q ok=%v", ip, ok)
		}
	})
	t.Run("link-local only", func(t *testing.T) {
		ip, ok := hasEligibleIP([]net.Addr{createMockIPv6Addr("fe80::abcd")})
		if ok || ip != "" {
			t.Fatalf("got %q ok=%v", ip, ok)
		}
	})
}
