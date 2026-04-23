// mandatory_fields_test.go
//
// Tests for mandatory-field validation on the encode path. Per 3GPP
// TS 29.002, each MAP operation's argument SEQUENCE declares some
// fields as non-OPTIONAL. The converters must reject empty caller
// input for those fields instead of silently emitting malformed wire
// bytes (e.g. a zero-length IMSI or a missing VLR number).
package gsmmap

import (
	"encoding/hex"
	"errors"
	"testing"
)

// UpdateLocationArg has mandatory imsi, msc-Number, vlr-Number
// (MAP-MS-DataTypes.asn:256-259).
func TestUpdateLocationMandatoryFields(t *testing.T) {
	base := func() *UpdateLocation {
		return &UpdateLocation{
			IMSI:      "204080012345678",
			MSCNumber: "31600000001",
			VLRNumber: "31600000002",
		}
	}

	t.Run("MissingIMSI", func(t *testing.T) {
		u := base()
		u.IMSI = ""
		_, err := u.Marshal()
		if err == nil {
			t.Fatal("expected error for missing IMSI")
		}
		if !errors.Is(err, ErrUpdateLocationMissingIMSI) {
			t.Errorf("expected ErrUpdateLocationMissingIMSI, got: %v", err)
		}
	})

	t.Run("MissingMSCNumber", func(t *testing.T) {
		u := base()
		u.MSCNumber = ""
		_, err := u.Marshal()
		if err == nil {
			t.Fatal("expected error for missing MSCNumber")
		}
		if !errors.Is(err, ErrUpdateLocationMissingMSCNumber) {
			t.Errorf("expected ErrUpdateLocationMissingMSCNumber, got: %v", err)
		}
	})

	t.Run("MissingVLRNumber", func(t *testing.T) {
		u := base()
		u.VLRNumber = ""
		_, err := u.Marshal()
		if err == nil {
			t.Fatal("expected error for missing VLRNumber")
		}
		if !errors.Is(err, ErrUpdateLocationMissingVLRNumber) {
			t.Errorf("expected ErrUpdateLocationMissingVLRNumber, got: %v", err)
		}
	})

	t.Run("AllPresent", func(t *testing.T) {
		if _, err := base().Marshal(); err != nil {
			t.Errorf("all mandatory fields present: unexpected error: %v", err)
		}
	})
}

// MT-ForwardSM-Arg uses SM-RP-DA and SM-RP-OA CHOICE variants. The
// public MtFsm struct hardcodes the IMSI and ServiceCentreAddressOA
// alternatives, so those string fields must be non-empty.
func TestMtFsmMandatoryFields(t *testing.T) {
	// Parse a known-valid MT-FSM (same hex as TestMtFsmFullStressRoundTrip)
	// to obtain a populated TPDU value; synthesizing one from scratch would
	// entangle this test with the sms/tpdu builder API.
	knownHex := "3077800832140080803138f684069169318488880463040b916971101174f40000422182612464805bd2e2b1252d467ff6de6c47efd96eb6a1d056cb0d69b49a10269c098537586e96931965b260d15613da72c29b91261bde72c6a1ad2623d682b5996d58331271375a0d1733eee4bd98ec768bd966b41c0d"
	knownBytes, err := hex.DecodeString(knownHex)
	if err != nil {
		t.Fatalf("hex decode: %v", err)
	}
	parsed, err := ParseMtFsm(knownBytes)
	if err != nil {
		t.Fatalf("ParseMtFsm: %v", err)
	}
	base := func() *MtFsm {
		m := *parsed // copy
		return &m
	}

	t.Run("MissingIMSI", func(t *testing.T) {
		m := base()
		m.IMSI = ""
		_, err := m.Marshal()
		if err == nil {
			t.Fatal("expected error for missing IMSI")
		}
		if !errors.Is(err, ErrMtFsmMissingIMSI) {
			t.Errorf("expected ErrMtFsmMissingIMSI, got: %v", err)
		}
	})

	t.Run("MissingServiceCentreAddressOA", func(t *testing.T) {
		m := base()
		m.ServiceCentreAddressOA = ""
		_, err := m.Marshal()
		if err == nil {
			t.Fatal("expected error for missing ServiceCentreAddressOA")
		}
		if !errors.Is(err, ErrMtFsmMissingServiceCentreAddressOA) {
			t.Errorf("expected ErrMtFsmMissingServiceCentreAddressOA, got: %v", err)
		}
	})

	t.Run("AllPresent", func(t *testing.T) {
		data, err := base().Marshal()
		if err != nil {
			t.Fatalf("all mandatory fields present: unexpected error: %v", err)
		}
		if len(data) == 0 {
			t.Error("expected non-empty wire output")
		}
	})
}

// RoutingInfoForSM-Arg has mandatory msisdn and serviceCentreAddress
// (MAP-SM-DataTypes.asn:63-66).
func TestSriSmMandatoryFields(t *testing.T) {
	base := func() *SriSm {
		return &SriSm{
			MSISDN:               "31612345678",
			SmRpPri:              true,
			ServiceCentreAddress: "31611111111",
		}
	}

	t.Run("MissingMSISDN", func(t *testing.T) {
		s := base()
		s.MSISDN = ""
		_, err := s.Marshal()
		if err == nil {
			t.Fatal("expected error for missing MSISDN")
		}
		if !errors.Is(err, ErrSriSmMissingMSISDN) {
			t.Errorf("expected ErrSriSmMissingMSISDN, got: %v", err)
		}
	})

	t.Run("MissingServiceCentreAddress", func(t *testing.T) {
		s := base()
		s.ServiceCentreAddress = ""
		_, err := s.Marshal()
		if err == nil {
			t.Fatal("expected error for missing ServiceCentreAddress")
		}
		if !errors.Is(err, ErrSriSmMissingServiceCentreAddress) {
			t.Errorf("expected ErrSriSmMissingServiceCentreAddress, got: %v", err)
		}
	})

	t.Run("AllPresent", func(t *testing.T) {
		if _, err := base().Marshal(); err != nil {
			t.Errorf("all mandatory fields present: unexpected error: %v", err)
		}
	})
}
