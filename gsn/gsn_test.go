package gsn

import (
	"encoding/hex"
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

func TestBuildInvalidIP(t *testing.T) {
	_, err := Build("not-an-ip")
	if err == nil {
		t.Fatal("expected error for invalid IP")
	}
}

func TestParseTooShort(t *testing.T) {
	_, err := Parse(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}
