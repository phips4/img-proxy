package internal

import (
	"net"
	"testing"
)

func TestFindIp(t *testing.T) {
	tests := []struct {
		name    string
		args    []net.Addr
		want    string
		wantErr bool
	}{
		{
			name: "IPv4 address found",
			args: []net.Addr{
				&net.IPNet{
					IP:   net.IPv4(192, 168, 0, 1),
					Mask: net.CIDRMask(24, 32),
				},
			},
			want:    "192.168.0.1",
			wantErr: false,
		},
		{
			name:    "No address found",
			args:    []net.Addr{},
			want:    "",
			wantErr: true,
		},
		{
			name: "IPv6 address found",
			args: []net.Addr{
				&net.IPNet{
					IP:   net.ParseIP("2001:db8::1"),
					Mask: net.CIDRMask(64, 128),
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Loopback address found",
			args: []net.Addr{
				&net.IPNet{
					IP:   net.IPv4(127, 0, 0, 1),
					Mask: net.CIDRMask(8, 32),
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindIp(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindIp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindIp() got = %v, want %v", got, tt.want)
			}
		})
	}
}
