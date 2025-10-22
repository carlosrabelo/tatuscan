package internal

import (
	"net"
	"testing"
)

func TestIsLocallyAdministeredMAC(t *testing.T) {
	tests := []struct {
		name     string
		mac      string
		expected bool
	}{
		// Cloud/virtual MACs (locally administered - bit 0x02 set)
		{"AWS typical", "02:00:17:12:34:56", true},
		{"Azure typical", "02:42:ac:11:00:02", true},
		{"Docker default", "02:42:ac:11:00:03", true},
		{"Vagrant typical", "02:00:4c:4f:4f:50", true},
		{"Generic locally admin", "06:00:00:00:00:01", true},
		{"VMware vNIC", "02:50:56:12:34:56", true},

		// Physical/OUI MACs (globally unique - bit 0x02 not set)
		{"Intel NIC", "00:1b:21:12:34:56", false},
		{"Realtek NIC", "00:e0:4c:12:34:56", false},
		{"Broadcom NIC", "00:10:18:12:34:56", false},
		{"Dell NIC", "00:14:22:12:34:56", false},
		{"HP NIC", "00:1f:29:12:34:56", false},
		{"Cisco NIC", "00:1d:71:12:34:56", false},

		// Edge cases
		{"Empty MAC", "", false},
		{"All zeros", "00:00:00:00:00:00", false},
		{"All ones first octet", "ff:ff:ff:ff:ff:ff", true}, // 0xff has bit 0x02 set
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mac == "" {
				result := isLocallyAdministeredMAC(net.HardwareAddr{})
				if result != tt.expected {
					t.Errorf("isLocallyAdministeredMAC(empty) = %v, want %v", result, tt.expected)
				}
				return
			}

			mac, err := net.ParseMAC(tt.mac)
			if err != nil {
				t.Fatalf("Failed to parse MAC %s: %v", tt.mac, err)
			}

			result := isLocallyAdministeredMAC(mac)
			if result != tt.expected {
				t.Errorf("isLocallyAdministeredMAC(%s) = %v, want %v", tt.mac, result, tt.expected)
			}
		})
	}
}

func TestIsVirtualInterface(t *testing.T) {
	tests := []struct {
		name      string
		ifaceName string
		expected  bool
	}{
		// Cloud provider interfaces
		{"AWS ENI", "eth0", false}, // Physical in cloud
		{"AWS EFA", "efa0", false}, // AWS Elastic Fabric Adapter is physical-like

		// Container/Docker interfaces
		{"Docker bridge", "docker0", true},
		{"Docker veth", "veth123abc", true},
		{"Docker bridge custom", "br-1234567890ab", true},

		// VM interfaces
		{"VMware", "vmnet1", true},
		{"VirtualBox host", "VirtualBox Host-Only Ethernet Adapter", true},
		{"Hyper-V", "Hyper-V Virtual Ethernet Adapter", true},

		// VPN interfaces
		{"WireGuard", "wg0", true},
		{"OpenVPN tun", "tun0", true},
		{"OpenVPN tap", "tap0", true},
		{"WireGuard with name", "wireguard-peer", true},

		// Physical interfaces (should not be filtered)
		{"Ethernet", "eth0", false},
		{"Ethernet alt", "enp0s3", false},
		{"WiFi", "wlan0", false},
		{"WiFi alt", "wlp2s0", false},
		{"Intel", "eno1", false},
		{"Realtek", "enp3s0", false},

		// Edge cases
		{"Empty name", "", false},
		{"Loopback", "lo", false}, // Handled separately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVirtualInterface(tt.ifaceName)
			if result != tt.expected {
				t.Errorf("isVirtualInterface(%s) = %v, want %v", tt.ifaceName, result, tt.expected)
			}
		})
	}
}

// TestCloudMACFiltering tests the complete filtering logic for cloud environments
func TestCloudMACFiltering(t *testing.T) {
	cloudMACs := []struct {
		name         string
		mac          string
		shouldFilter bool
		reason       string
	}{
		// AWS EC2 typical scenarios
		{"AWS EC2 primary", "02:00:17:12:34:56", true, "locally administered"},
		{"AWS EC2 secondary", "02:00:17:98:76:54", true, "locally administered"},

		// Azure VM scenarios
		{"Azure VM", "02:42:ac:11:00:02", true, "locally administered"},

		// Google Cloud scenarios
		{"GCP VM", "02:42:ac:11:00:03", true, "locally administered"},

		// Docker scenarios
		{"Docker container", "02:42:ac:11:00:04", true, "locally administered"},

		// Physical NICs that should NOT be filtered
		{"AWS bare metal Intel", "00:1b:21:12:34:56", false, "globally unique OUI"},
		{"Azure bare metal Mellanox", "00:02:c9:12:34:56", false, "globally unique OUI"},
		{"On-premise Intel", "00:e0:4c:12:34:56", false, "globally unique OUI"},
	}

	for _, tt := range cloudMACs {
		t.Run(tt.name, func(t *testing.T) {
			mac, err := net.ParseMAC(tt.mac)
			if err != nil {
				t.Fatalf("Failed to parse MAC %s: %v", tt.mac, err)
			}

			isLocallyAdmin := isLocallyAdministeredMAC(mac)

			if isLocallyAdmin != tt.shouldFilter {
				t.Errorf("MAC %s (%s): isLocallyAdministered = %v, expected %v (reason: %s)",
					tt.mac, tt.name, isLocallyAdmin, tt.shouldFilter, tt.reason)
			}
		})
	}
}

// BenchmarkMACFiltering benchmarks the MAC filtering performance
func BenchmarkMACFiltering(b *testing.B) {
	testMACs := []string{
		"02:00:17:12:34:56", // AWS
		"00:1b:21:12:34:56", // Intel
		"02:42:ac:11:00:02", // Docker
		"00:e0:4c:12:34:56", // Realtek
	}

	macs := make([]net.HardwareAddr, len(testMACs))
	for i, macStr := range testMACs {
		mac, _ := net.ParseMAC(macStr)
		macs[i] = mac
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, mac := range macs {
			isLocallyAdministeredMAC(mac)
		}
	}
}
