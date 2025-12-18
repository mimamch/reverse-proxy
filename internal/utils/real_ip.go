package utils

import (
	"net"
	"net/http"
)

var cloudflareCIDRs = []string{
	// IPv4
	"173.245.48.0/20",
	"103.21.244.0/22",
	"103.22.200.0/22",
	"103.31.4.0/22",
	"141.101.64.0/18",
	"108.162.192.0/18",
	"190.93.240.0/20",
	"188.114.96.0/20",
	"197.234.240.0/22",
	"198.41.128.0/17",
	"162.158.0.0/15",
	"104.16.0.0/13",
	"104.24.0.0/14",
	"172.64.0.0/13",
	"131.0.72.0/22",

	// IPv6
	"2400:cb00::/32",
	"2606:4700::/32",
	"2803:f800::/32",
	"2405:b500::/32",
	"2405:8100::/32",
	"2a06:98c0::/29",
	"2c0f:f248::/32",
}

var cfIPNets []*net.IPNet

func InitTrustedCIDRs() error {
	for _, cidr := range cloudflareCIDRs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return err
		}
		cfIPNets = append(cfIPNets, ipnet)
	}
	return nil
}

func isFromCloudflare(ip net.IP) bool {

	for _, ipnet := range cfIPNets {
		if ipnet.Contains(ip) {
			return true
		}
	}
	return false
}

func GetRealClientIP(r *http.Request) net.IP {
	remoteIPStr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil
	}

	remoteIP := net.ParseIP(remoteIPStr)
	if remoteIP == nil {
		return nil
	}

	if isFromCloudflare(remoteIP) {
		if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
			if parsed := net.ParseIP(cfIP); parsed != nil {
				return parsed
			}
		}
	}

	// fallback
	return remoteIP
}
