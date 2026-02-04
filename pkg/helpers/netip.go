package helpers

import "net/netip"

func ParseIPAddress(s string) (*netip.Addr, error) {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return nil, err
	}
	return &addr, nil
}
