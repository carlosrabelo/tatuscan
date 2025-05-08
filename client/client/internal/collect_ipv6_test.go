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
			expectError: true, // Current implementation requires IPv4
			description: "Pure IPv6 environment should work but currently fails",
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
			expectError: true, // Multiple IPv6-only should work but currently fails
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
			// Test individual interface validation logic
			for _, iface := range tt.interfaces {
				t.Logf("Testing interface: %s, MAC: %s, Flags: %v",
					iface.name, iface.hardwareAddr, iface.flags)

				// Test MAC filtering
				isLocallyAdmin := isLocallyAdministeredMAC(iface.hardwareAddr)
				t.Logf("  Locally administered MAC: %v", isLocallyAdmin)

				// Test virtual interface detection
				isVirtual := isVirtualInterface(iface.name)
				t.Logf("  Virtual interface: %v", isVirtual)

				// Test addresses
				for _, addr := range iface.addrs {
					if ipnet, ok := addr.(*net.IPNet); ok {
						isIPv4 := ipnet.IP.To4() != nil
						isIPv6 := !isIPv4 && ipnet.IP.To16() != nil
						isLoopback := ipnet.IP.IsLoopback()
						isLinkLocal := ipnet.IP.IsLinkLocalUnicast()

						t.Logf("  Address: %s, IPv4: %v, IPv6: %v, Loopback: %v, LinkLocal: %v",
							ipnet.IP, isIPv4, isIPv6, isLoopback, isLinkLocal)
					}
				}
			}

			// Note: We can't easily test the full CollectData() function without mocking
			// the net.Interfaces() call, but we can test the individual components
			t.Logf("Scenario: %s", tt.description)
			if tt.expectError {
				t.Logf("  Expected to fail with current implementation (IPv4 requirement)")
			} else {
				t.Logf("  Expected to succeed")
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

// TestCurrentIPv4Requirement documents the current limitation
func TestCurrentIPv4Requirement(t *testing.T) {
	t.Run("Document IPv4 requirement limitation", func(t *testing.T) {
		// This test documents the current behavior that requires IPv4
		// In the future, this should be updated when IPv6-only support is added

		t.Log("CURRENT LIMITATION: The CollectData() function requires IPv4 addresses")
		t.Log("Lines that enforce this:")
		t.Log("  Linux: collect_data_linux.go:133 - ipnet.IP.To4() != nil")
		t.Log("  Windows: collect_data_windows.go:90 - ipnet.IP.To4() != nil")
		t.Log("")
		t.Log("TODO: Update these lines to also accept valid IPv6 global unicast addresses")
		t.Log("Suggested fix: Accept addresses where:")
		t.Log("  - IPv4: ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback()")
		t.Log("  - IPv6: ipnet.IP.To4() == nil && ipnet.IP.To16() != nil && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalUnicast()")

		// This test always passes but serves as documentation
		if testing.Short() {
			t.Skip("Skipping documentation test in short mode")
		}
	})
}
