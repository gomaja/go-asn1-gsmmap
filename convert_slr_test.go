// convert_slr_test.go
//
// Tests for the SLR sub-type converters: LCSLocationInfo and
// DeferredmtLrData. Round-trip + targeted negative cases.
package gsmmap

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/runtime"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// runtimeBitString0 is a zero-length BIT STRING used to exercise the
// BitLength==0 = "absent" decode path.
var runtimeBitString0 = runtime.BitString{Bytes: []byte{}, BitLength: 0}

// =============================================================================
// LCSLocationInfo
// =============================================================================

func TestLCSLocationInfoRoundTrip(t *testing.T) {
	mme := HexBytes("mme1.example.com")          // 16 octets, within 9..255
	aaa := HexBytes("aaa.example.com")           // 15 octets
	sgsnName := HexBytes("sgsn.example.com")     // 16 octets
	sgsnRealm := HexBytes("epc.mnc001.mcc204.x") // 19 octets

	cases := []struct {
		name string
		in   *LCSLocationInfo
	}{
		{"minimal (NetworkNodeNumber only)", &LCSLocationInfo{
			NetworkNodeNumber:       "31650000000",
			NetworkNodeNumberNature: 0x10,
			NetworkNodeNumberPlan:   0x01,
		}},
		{"with LMSI + GprsNodeIndicator", &LCSLocationInfo{
			NetworkNodeNumber:       "31650000000",
			NetworkNodeNumberNature: 0x10,
			NetworkNodeNumberPlan:   0x01,
			LMSI:                    HexBytes{0x01, 0x02, 0x03, 0x04},
			GprsNodeIndicator:       true,
		}},
		{"with AdditionalNumber (SGSN)", &LCSLocationInfo{
			NetworkNodeNumber:       "31650000000",
			NetworkNodeNumberNature: 0x10,
			NetworkNodeNumberPlan:   0x01,
			AdditionalNumber: &AdditionalNumber{
				SgsnNumber:       "31660000000",
				SgsnNumberNature: 0x10,
				SgsnNumberPlan:   0x01,
			},
		}},
		{"with LCS capability sets", &LCSLocationInfo{
			NetworkNodeNumber:       "31650000000",
			NetworkNodeNumberNature: 0x10,
			NetworkNodeNumberPlan:   0x01,
			SupportedLCSCapabilitySets: &SupportedLCSCapabilitySets{
				LcsCapabilitySet1: true,
				LcsCapabilitySet3: true,
			},
			AdditionalLCSCapabilitySets: &SupportedLCSCapabilitySets{
				LcsCapabilitySet2: true,
			},
		}},
		{"with all Diameter names", &LCSLocationInfo{
			NetworkNodeNumber:       "31650000000",
			NetworkNodeNumberNature: 0x10,
			NetworkNodeNumberPlan:   0x01,
			MmeName:                 mme,
			AaaServerName:           aaa,
			SgsnName:                sgsnName,
			SgsnRealm:               sgsnRealm,
		}},
		{"full population", &LCSLocationInfo{
			NetworkNodeNumber:       "31650000000",
			NetworkNodeNumberNature: 0x10,
			NetworkNodeNumberPlan:   0x01,
			LMSI:                    HexBytes{0x0a, 0x0b, 0x0c, 0x0d},
			GprsNodeIndicator:       true,
			AdditionalNumber: &AdditionalNumber{
				MscNumber:       "31640000000",
				MscNumberNature: 0x10,
				MscNumberPlan:   0x01,
			},
			SupportedLCSCapabilitySets:  &SupportedLCSCapabilitySets{LcsCapabilitySet1: true},
			AdditionalLCSCapabilitySets: &SupportedLCSCapabilitySets{LcsCapabilitySet5: true},
			MmeName:                     mme,
			AaaServerName:               aaa,
			SgsnName:                    sgsnName,
			SgsnRealm:                   sgsnRealm,
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertLCSLocationInfoToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := convertWireToLCSLocationInfo(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", tc.in, out)
			}
		})
	}
}

func TestLCSLocationInfoNilPassesThrough(t *testing.T) {
	if w, err := convertLCSLocationInfoToWire(nil); err != nil || w != nil {
		t.Errorf("encode nil: got w=%v err=%v", w, err)
	}
	if out, err := convertWireToLCSLocationInfo(nil); err != nil || out != nil {
		t.Errorf("decode nil: got out=%v err=%v", out, err)
	}
}

func TestLCSLocationInfoEmptyNetworkNodeRejected(t *testing.T) {
	_, err := convertLCSLocationInfoToWire(&LCSLocationInfo{})
	if !errors.Is(err, ErrLCSLocationInfoNetworkNodeEmpty) {
		t.Errorf("encode empty NetworkNodeNumber: want ErrLCSLocationInfoNetworkNodeEmpty, got %v", err)
	}
}

func TestLCSLocationInfoLMSISizeRejected(t *testing.T) {
	_, err := convertLCSLocationInfoToWire(&LCSLocationInfo{
		NetworkNodeNumber:       "31650000000",
		NetworkNodeNumberNature: 0x10,
		NetworkNodeNumberPlan:   0x01,
		LMSI:                    HexBytes{0x01, 0x02, 0x03}, // 3 octets, must be 4
	})
	if !errors.Is(err, ErrLCSLocationInfoLMSIInvalidSize) {
		t.Errorf("encode LMSI=3: want ErrLCSLocationInfoLMSIInvalidSize, got %v", err)
	}
}

func TestLCSLocationInfoDiameterIdentitySizeRejected(t *testing.T) {
	base := func() *LCSLocationInfo {
		return &LCSLocationInfo{
			NetworkNodeNumber:       "31650000000",
			NetworkNodeNumberNature: 0x10,
			NetworkNodeNumberPlan:   0x01,
		}
	}
	cases := []struct {
		name    string
		mutate  func(*LCSLocationInfo)
		wantErr error
	}{
		{"MmeName too short", func(l *LCSLocationInfo) { l.MmeName = HexBytes("short") }, ErrLCSLocationInfoMmeNameSize},
		{"AaaServerName too short", func(l *LCSLocationInfo) { l.AaaServerName = HexBytes("short") }, ErrLCSLocationInfoAaaServerNameSize},
		{"SgsnName too short", func(l *LCSLocationInfo) { l.SgsnName = HexBytes("short") }, ErrLCSLocationInfoSgsnNameSize},
		{"SgsnRealm too short", func(l *LCSLocationInfo) { l.SgsnRealm = HexBytes("short") }, ErrLCSLocationInfoSgsnRealmSize},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := base()
			tc.mutate(in)
			_, err := convertLCSLocationInfoToWire(in)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("want %v, got %v", tc.wantErr, err)
			}
		})
	}
}

// A present-but-empty BIT STRING (BitLength==0) for the LCS capability
// sets must decode as absent, matching the convert_updateloc.go /
// convert_updategprsloc.go precedent. Without the BitLength>0 guard a
// zero-length wire value would surface an all-false surrogate that
// cannot round-trip.
func TestLCSLocationInfoWireEmptyCapabilitySetsTreatedAsAbsent(t *testing.T) {
	w := &gsm_map.LCSLocationInfo{
		NetworkNodeNumber:           gsm_map.ISDNAddressString{0x91, 0x13, 0x05, 0x00, 0x00, 0x00, 0xf0}, // 31650000000
		SupportedLCSCapabilitySets:  &runtimeBitString0,
		AdditionalLCSCapabilitySets: &runtimeBitString0,
	}
	out, err := convertWireToLCSLocationInfo(w)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.SupportedLCSCapabilitySets != nil {
		t.Error("zero-length SupportedLCSCapabilitySets should decode as absent (nil)")
	}
	if out.AdditionalLCSCapabilitySets != nil {
		t.Error("zero-length AdditionalLCSCapabilitySets should decode as absent (nil)")
	}
}

// Decoder rejects an empty-digits NetworkNodeNumber for round-trip fidelity.
func TestLCSLocationInfoWireEmptyNodeDecodedEmptyRejected(t *testing.T) {
	emptyAddr := gsm_map.ISDNAddressString{0x91} // header-only, no digits
	w := &gsm_map.LCSLocationInfo{NetworkNodeNumber: emptyAddr}
	_, err := convertWireToLCSLocationInfo(w)
	if !errors.Is(err, ErrLCSLocationInfoNetworkNodeDecodedEmpty) {
		t.Errorf("want ErrLCSLocationInfoNetworkNodeDecodedEmpty, got %v", err)
	}
}

// =============================================================================
// DeferredmtLrData
// =============================================================================

func TestDeferredmtLrDataRoundTrip(t *testing.T) {
	tc1 := TerminationMtLrRestart
	cases := []struct {
		name string
		in   *DeferredmtLrData
	}{
		{"event type only", &DeferredmtLrData{
			DeferredLocationEventType: DeferredLocationEventType{MsAvailable: true},
		}},
		{"with termination cause", &DeferredmtLrData{
			DeferredLocationEventType: DeferredLocationEventType{EnteringIntoArea: true},
			TerminationCause:          &tc1,
		}},
		{"with location info (mt-lrRestart)", &DeferredmtLrData{
			DeferredLocationEventType: DeferredLocationEventType{PeriodicLDR: true},
			TerminationCause:          &tc1,
			LcsLocationInfo: &LCSLocationInfo{
				NetworkNodeNumber:       "31650000000",
				NetworkNodeNumberNature: 0x10,
				NetworkNodeNumberPlan:   0x01,
			},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertDeferredmtLrDataToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := convertWireToDeferredmtLrData(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", tc.in, out)
			}
		})
	}
}

func TestDeferredmtLrDataNilPassesThrough(t *testing.T) {
	if w, err := convertDeferredmtLrDataToWire(nil); err != nil || w != nil {
		t.Errorf("encode nil: got w=%v err=%v", w, err)
	}
	if out, err := convertWireToDeferredmtLrData(nil); err != nil || out != nil {
		t.Errorf("decode nil: got out=%v err=%v", out, err)
	}
}

func TestDeferredmtLrDataTerminationCauseOutOfRangeRejected(t *testing.T) {
	bad := TerminationCause(99)
	_, err := convertDeferredmtLrDataToWire(&DeferredmtLrData{
		DeferredLocationEventType: DeferredLocationEventType{MsAvailable: true},
		TerminationCause:          &bad,
	})
	if !errors.Is(err, ErrTerminationCauseInvalid) {
		t.Errorf("want ErrTerminationCauseInvalid, got %v", err)
	}
}

// TerminationCause is extensible — decoder preserves unknown values
// per Postel even though the encoder rejects them.
func TestDeferredmtLrDataTerminationCauseDecoderLenient(t *testing.T) {
	bad := gsm_map.TerminationCause(99)
	w := &gsm_map.DeferredmtLrData{
		DeferredLocationEventType: convertDeferredLocationEventTypeToBitString(&DeferredLocationEventType{MsAvailable: true}),
		TerminationCause:          &bad,
	}
	out, err := convertWireToDeferredmtLrData(w)
	if err != nil {
		t.Fatalf("decode TerminationCause=99: unexpected error %v", err)
	}
	if out.TerminationCause == nil || int64(*out.TerminationCause) != 99 {
		t.Errorf("decoder leniency: want 99 preserved, got %v", out.TerminationCause)
	}
}
