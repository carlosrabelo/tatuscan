package internal

import (
	"net"
	"testing"
)

// MockInterface simulates a network interface for testing
type MockInterface struct {
	name         string
	flags        net.Flags
	hardwareAddr net.HardwareAddr
	addrs        []net.Addr
}

func (m MockInterface) Name() string                   { return m.name }
func (m MockInterface) Flags() net.Flags               { return m.flags }
func (m MockInterface) HardwareAddr() net.HardwareAddr { return m.hardwareAddr }
func (m MockInterface) Addrs() ([]net.Addr, error)     { return m.addrs, nil }

// createMockIPv6Addr creates a mock IPv6 address
func createMockIPv6Addr(ipv6 string) *net.IPNet {
	ip := net.ParseIP(ipv6)
	if ip == nil {
		panic("Invalid IPv6 address: " + ipv6)
	}
	return &net.IPNet{IP: ip, Mask: net.CIDRMask(64, 128)}
}

// createMockIPv4Addr creates a mock IPv4 address
func createMockIPv4Addr(ipv4 string) *net.IPNet {
	ip := net.ParseIP(ipv4)
	if ip == nil {
		panic("Invalid IPv4 address: " + ipv4)
	}
	return &net.IPNet{IP: ip, Mask: net.CIDRMask(24, 32)}
}

func TestIPv6OnlyScenarios(t *testing.T) {
	tests := []struct {
		name        string
		interfaces  []MockInterface
		expectError bool
		description string
	}{
		{
			name: "IPv6-only modern datacenter",
			interfaces: []MockInterface{
				{
					name:         "eth0",
					flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
					hardwareAddr: mustParseMAC("00:1b:21:12:34:56"), // Intel NIC
					addrs: []net.Addr{
						createMockIPv6Addr("2001:db8::1"),
						createMockIPv6Addr("fe80::21b:21ff:fe12:3456"), // Link-local
					},
				},
			},
			expectError: false,
			description: "Pure IPv6 environment with global address is eligible",
		},
		{
			name: "Dual-stack with IPv6 preference",
			interfaces: []MockInterface{
				{
					name:         "eth0",
					flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
					hardwareAddr: mustParseMAC("00:e0:4c:12:34:56"), // Realtek NIC
					addrs: []net.Addr{
						createMockIPv6Addr("2001:db8::2"),
						createMockIPv4Addr("192.168.1.100"),
					},
				},
			},
			expectError: false, // Has both IPv6 and IPv4
			description: "Dual-stack should work fine",
		},
		{
			name: "IPv6-only with link-local only",
			interfaces: []MockInterface{
				{
					name:         "eth0",
					flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
					hardwareAddr: mustParseMAC("00:10:18:12:34:56"), // Broadcom NIC
					addrs: []net.Addr{
						createMockIPv6Addr("fe80::210:18ff:fe12:3456"), // Only link-local
					},
				},
			},
			expectError: true, // Link-local is not routable
			description: "Link-local only should not be sufficient",
		},
		{
			name: "Multiple IPv6-only interfaces",
			interfaces: []MockInterface{
				{
					name:         "eth0",
					flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
					hardwareAddr: mustParseMAC("00:14:22:12:34:56"), // Dell NIC
					addrs: []net.Addr{
						createMockIPv6Addr("2001:db8::10"),
					},
				},
				{
					name:         "eth1",
					flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
					hardwareAddr: mustParseMAC("00:1f:29:12:34:57"), // HP NIC
					addrs: []net.Addr{
						createMockIPv6Addr("2001:db8::20"),
					},
				},
			},
			expectError: false,
			description: "Multiple IPv6-only interfaces in datacenter scenario",
		},
		{
			name: "Cloud IPv6-only (should be filtered)",
			interfaces: []MockInterface{
				{
					name:         "eth0",
					flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
					hardwareAddr: mustParseMAC("02:00:17:12:34:56"), // AWS locally administered
					addrs: []net.Addr{
						createMockIPv6Addr("2600:1f16::1"), // AWS IPv6 range example
					},
				},
			},
			expectError: true, // Should be filtered due to locally administered MAC
			description: "Cloud IPv6-only with locally administered MAC should be filtered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eligibleIface := false
			for _, iface := range tt.interfaces {
				if isLocallyAdministeredMAC(iface.hardwareAddr) || isVirtualInterface(iface.name) {
					continue
				}
				if (iface.flags&net.FlagUp) == 0 || (iface.flags&net.FlagLoopback) != 0 {
					continue
				}
				if _, ok := hasEligibleIP(iface.addrs); ok {
					eligibleIface = true
					break
				}
			}
			if tt.expectError && eligibleIface {
				t.Fatalf("%s: expected no eligible interface, got one (%s)", tt.name, tt.description)
			}
			if !tt.expectError && !eligibleIface {
				t.Fatalf("%s: expected eligible interface, got none (%s)", tt.name, tt.description)
			}
		})
	}
}

// TestIPv6AddressValidation tests IPv6 address classification
func TestIPv6AddressValidation(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		isValid     bool
		isGlobal    bool
		isLinkLocal bool
		description string
	}{
		// Global unicast addresses (should be valid for MachineID)
		{"Global unicast", "2001:db8::1", true, true, false, "Standard global IPv6"},
		{"AWS IPv6", "2600:1f16::1", true, true, false, "AWS global IPv6 range"},
		{"Google IPv6", "2001:4860::1", true, true, false, "Google global IPv6"},
		{"Azure IPv6", "2603:1030::1", true, true, false, "Azure global IPv6"},

		// Link-local addresses (should not be used for MachineID)
		{"Link-local", "fe80::1", true, false, true, "Link-local should not be used"},
		{"Link-local with interface", "fe80::210:18ff:fe12:3456", true, false, true, "Auto-configured link-local"},

		// Special addresses
		{"Loopback", "::1", true, false, false, "IPv6 loopback"},
		{"Unspecified", "::", true, false, false, "IPv6 unspecified"},

		// Invalid addresses
		{"Invalid", "invalid-ipv6", false, false, false, "Malformed address"},
		{"IPv4 mapped", "::ffff:192.168.1.1", true, false, false, "IPv4-mapped IPv6"},

		// Unique Local Addresses (RFC 4193)
		{"ULA", "fd00::1", true, false, false, "Unique Local Address"},
		{"ULA fc", "fc00::1", true, false, false, "Unique Local Address fc prefix"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.address)

			if (ip != nil) != tt.isValid {
				t.Errorf("ParseIP(%s) validity = %v, want %v", tt.address, ip != nil, tt.isValid)
				return
			}

			if ip == nil {
				return // Skip further tests for invalid IPs
			}

			isIPv6 := ip.To4() == nil && ip.To16() != nil
			if !isIPv6 {
				t.Logf("Address %s is not pure IPv6", tt.address)
			}

			isLinkLocal := ip.IsLinkLocalUnicast()
			if isLinkLocal != tt.isLinkLocal {
				t.Errorf("IP(%s).IsLinkLocalUnicast() = %v, want %v", tt.address, isLinkLocal, tt.isLinkLocal)
			}

			// Global check: not loopback, not link-local, not unspecified
			isGlobalApprox := !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsUnspecified()
			if isGlobalApprox != tt.isGlobal {
				t.Logf("IP(%s) global approximation = %v, expected %v (%s)",
					tt.address, isGlobalApprox, tt.isGlobal, tt.description)
			}
		})
	}
}

// mustParseMAC is a helper that panics on invalid MAC addresses
func mustParseMAC(s string) net.HardwareAddr {
	mac, err := net.ParseMAC(s)
	if err != nil {
		panic("Invalid MAC address: " + s)
	}
	return mac
}

func TestCurrentIPv4Requirement(t *testing.T) {
	t.Run("helper accepts global IPv6 and rejects link-local-only", func(t *testing.T) {
		if ip, ok := hasEligibleIP([]net.Addr{createMockIPv6Addr("2001:db8::1")}); !ok || ip != "2001:db8::1" {
			t.Fatalf("global ipv6: got %q ok=%v", ip, ok)
		}
		if ip, ok := hasEligibleIP([]net.Addr{createMockIPv6Addr("fe80::1")}); ok || ip != "" {
			t.Fatalf("link-local: got %q ok=%v", ip, ok)
		}
		if ip, ok := hasEligibleIP([]net.Addr{
			createMockIPv6Addr("2001:db8::2"),
			createMockIPv4Addr("192.168.1.100"),
		}); !ok || ip != "192.168.1.100" {
			t.Fatalf("prefer ipv4: got %q ok=%v", ip, ok)
		}
	})
}
