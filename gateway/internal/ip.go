package internal

import (
	"errors"
	"net"
)

var ErrNoIp = errors.New("no ip found")

func FindIp(addrs []net.Addr) (string, error) {
	if len(addrs) == 0 {
		return "", ErrNoIp
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}
	return "", ErrNoIp
}
