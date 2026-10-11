package gsn

import (
	"errors"
	"testing"
)

// TS 23.003 V20.1.0 §5.1: the header gives the address length in octets.
func TestMutationGSNTruncatedAddress(t *testing.T) {
	for _, tc := range []struct {
		name    string
		data    []byte
		want    error
		message string
	}{
		{"empty", nil, ErrAddressTooShort, "gsn: address too short"},
		{"IPv4 header", []byte{4}, ErrDataTooShort, "gsn: data too short: expected 5 bytes, got 1"},
		{"IPv4 missing octet", []byte{4, 192, 0, 2}, ErrDataTooShort, "gsn: data too short: expected 5 bytes, got 4"},
		{"IPv6 header", []byte{0x50}, ErrDataTooShort, "gsn: data too short: expected 17 bytes, got 1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.data)
			if got != "" || !errors.Is(err, tc.want) {
				t.Fatalf("Parse = %q, %v; want empty, %v", got, err, tc.want)
			}
			if err.Error() != tc.message {
				t.Fatalf("error = %q, want %q", err, tc.message)
			}
		})
	}
}
