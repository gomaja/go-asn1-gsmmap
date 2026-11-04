package gsn

import (
	"encoding/hex"
	"errors"
	"testing"
)

func TestBuildParseIPv4(t *testing.T) {
	data, err := Build("192.168.1.1")
	if err != nil {
		t.Fatalf("Build IPv4 error: %v", err)
	}

	expected := "04c0a80101"
	if hex.EncodeToString(data) != expected {
		t.Errorf("Build IPv4 = %s, want %s", hex.EncodeToString(data), expected)
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse IPv4 error: %v", err)
	}
	if parsed != "192.168.1.1" {
		t.Errorf("Parse IPv4 = %s, want 192.168.1.1", parsed)
	}
}

func TestBuildParseIPv6(t *testing.T) {
	data, err := Build("2001:db8::1")
	if err != nil {
		t.Fatalf("Build IPv6 error: %v", err)
	}

	expected := "5020010db8000000000000000000000001"
	if hex.EncodeToString(data) != expected {
		t.Errorf("Build IPv6 = %s, want %s", hex.EncodeToString(data), expected)
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse IPv6 error: %v", err)
	}
	if parsed != "2001:db8::1" {
		t.Errorf("Parse IPv6 = %s, want 2001:db8::1", parsed)
	}
}

// TS 23.003 V20.1.0 §5.1 codes the address type and length explicitly, so
// an IPv6-typed address that maps an IPv4 address (RFC 4291 §2.5.5.2) keeps
// its family: Parse renders it as ::ffff:a.b.c.d (RFC 5952 §5) and Build sends
// that back as the same 16 octets, where the dotted quad alone is sent as 4.
func TestBuildParseIPv4Mapped(t *testing.T) {
	const mapped = "5000000000000000000000ffffc0000201"
	data, err := hex.DecodeString(mapped)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if parsed != "::ffff:192.0.2.1" {
		t.Errorf("Parse = %s, want ::ffff:192.0.2.1", parsed)
	}
	built, err := Build(parsed)
	if err != nil {
		t.Fatalf("Build(%s): %v", parsed, err)
	}
	if hex.EncodeToString(built) != mapped {
		t.Errorf("Build(%s) = %x, want %s", parsed, built, mapped)
	}
	built, err = Build("192.0.2.1")
	if err != nil {
		t.Fatalf("Build IPv4: %v", err)
	}
	if hex.EncodeToString(built) != "04c0000201" {
		t.Errorf("Build IPv4 = %x, want 04c0000201", built)
	}
}

func TestBuildInvalidIP(t *testing.T) {
	for _, s := range []string{"not-an-ip", "fe80::1%eth0", "192.0.2", ""} {
		if _, err := Build(s); !errors.Is(err, ErrInvalidIP) {
			t.Errorf("Build(%q) = %v, want ErrInvalidIP", s, err)
		}
	}
}

func TestParseTooShort(t *testing.T) {
	_, err := Parse(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

// TS 23.003 V20.1.0 §5.1: the header gives the address length in octets.
func TestGSNTruncatedAddress(t *testing.T) {
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
