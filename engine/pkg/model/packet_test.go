package model

import "testing"

func TestProtocolName(t *testing.T) {
	tests := []struct {
		protocol uint8
		want     string
	}{
		{protocol: 0, want: "PROTO_0"},
		{protocol: 1, want: "ICMP"},
		{protocol: 2, want: "PROTO_2"},
		{protocol: 6, want: "TCP"},
		{protocol: 17, want: "UDP"},
		{protocol: 255, want: "PROTO_255"},
	}
	for _, test := range tests {
		if got := ProtocolName(test.protocol); got != test.want {
			t.Errorf("ProtocolName(%d) = %q, want %q", test.protocol, got, test.want)
		}
	}
}
