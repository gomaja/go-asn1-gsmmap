// convert_psl_lcsclientid_test.go
//
// Tests for the LCS-Client identifier tree converters
// (LCSClientName, LCSRequestorID, LCSClientID).
package gsmmap

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// LCSClientName
// ============================================================================

func TestLCSClientNameRoundTrip(t *testing.T) {
	fmtMsisdn := LCSFormatMsisdn
	cases := []struct {
		name string
		in   *LCSClientName
	}{
		{"minimal", &LCSClientName{
			DataCodingScheme: 0x0f,
			NameString:       HexBytes{0x41, 0x42, 0x43},
		}},
		{"with format indicator", &LCSClientName{
			DataCodingScheme:   0x00,
			NameString:         HexBytes{0x4e, 0x41, 0x4d, 0x45},
			LcsFormatIndicator: &fmtMsisdn,
		}},
		{"max-length name string (63 octets)", &LCSClientName{
			DataCodingScheme: 0x0f,
			NameString:       make(HexBytes, NameStringMaxLen),
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertLCSClientNameToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := convertWireToLCSClientName(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", tc.in, out)
			}
		})
	}
}

func TestLCSClientNameNilPassesThrough(t *testing.T) {
	wire, err := convertLCSClientNameToWire(nil)
	if err != nil || wire != nil {
		t.Errorf("nil → nil expected, got wire=%v err=%v", wire, err)
	}
	out, err := convertWireToLCSClientName(nil)
	if err != nil || out != nil {
		t.Errorf("nil → nil expected, got out=%v err=%v", out, err)
	}
}

func TestLCSClientNameEmptyNameStringRejected(t *testing.T) {
	_, err := convertLCSClientNameToWire(&LCSClientName{
		DataCodingScheme: 0x0f,
		NameString:       HexBytes{},
	})
	if !errors.Is(err, ErrLCSClientNameNameStringSize) {
		t.Errorf("want ErrLCSClientNameNameStringSize on encode, got %v", err)
	}
}

func TestLCSClientNameOversizedNameStringRejected(t *testing.T) {
	tooBig := make(HexBytes, NameStringMaxLen+1)
	_, err := convertLCSClientNameToWire(&LCSClientName{
		DataCodingScheme: 0x0f,
		NameString:       tooBig,
	})
	if !errors.Is(err, ErrLCSClientNameNameStringSize) {
		t.Errorf("want ErrLCSClientNameNameStringSize on encode, got %v", err)
	}
}

func TestLCSClientNameWireDataCodingSchemeMustBeOneOctet(t *testing.T) {
	w := &gsm_map.LCSClientName{
		DataCodingScheme: gsm_map.USSDDataCodingScheme{0x0f, 0x10}, // too long
		NameString:       gsm_map.NameString{0x41},
	}
	_, err := convertWireToLCSClientName(w)
	if !errors.Is(err, ErrUSSDDataCodingSchemeInvalidSize) {
		t.Errorf("want ErrUSSDDataCodingSchemeInvalidSize, got %v", err)
	}
}

// ============================================================================
// LCSRequestorID
// ============================================================================

func TestLCSRequestorIDRoundTrip(t *testing.T) {
	fmtUrl := LCSFormatUrl
	cases := []struct {
		name string
		in   *LCSRequestorID
	}{
		{"minimal", &LCSRequestorID{
			DataCodingScheme:  0x0f,
			RequestorIDString: HexBytes{0x52, 0x49, 0x44},
		}},
		{"with format indicator", &LCSRequestorID{
			DataCodingScheme:   0x00,
			RequestorIDString:  HexBytes{0x55, 0x52, 0x4c},
			LcsFormatIndicator: &fmtUrl,
		}},
		{"max-length requestor string (63 octets)", &LCSRequestorID{
			DataCodingScheme:  0x0f,
			RequestorIDString: make(HexBytes, RequestorIDStringMaxLen),
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertLCSRequestorIDToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := convertWireToLCSRequestorID(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", tc.in, out)
			}
		})
	}
}

func TestLCSRequestorIDNilPassesThrough(t *testing.T) {
	wire, err := convertLCSRequestorIDToWire(nil)
	if err != nil || wire != nil {
		t.Errorf("nil → nil expected, got wire=%v err=%v", wire, err)
	}
	out, err := convertWireToLCSRequestorID(nil)
	if err != nil || out != nil {
		t.Errorf("nil → nil expected, got out=%v err=%v", out, err)
	}
}

func TestLCSRequestorIDEmptyStringRejected(t *testing.T) {
	_, err := convertLCSRequestorIDToWire(&LCSRequestorID{
		DataCodingScheme:  0x0f,
		RequestorIDString: HexBytes{},
	})
	if !errors.Is(err, ErrLCSRequestorIDStringSize) {
		t.Errorf("want ErrLCSRequestorIDStringSize, got %v", err)
	}
}

func TestLCSRequestorIDOversizedStringRejected(t *testing.T) {
	tooBig := make(HexBytes, RequestorIDStringMaxLen+1)
	_, err := convertLCSRequestorIDToWire(&LCSRequestorID{
		DataCodingScheme:  0x0f,
		RequestorIDString: tooBig,
	})
	if !errors.Is(err, ErrLCSRequestorIDStringSize) {
		t.Errorf("want ErrLCSRequestorIDStringSize, got %v", err)
	}
}

func TestLCSRequestorIDWireDataCodingSchemeMustBeOneOctet(t *testing.T) {
	w := &gsm_map.LCSRequestorID{
		DataCodingScheme:  gsm_map.USSDDataCodingScheme{0x0f, 0x10},
		RequestorIDString: gsm_map.RequestorIDString{0x41},
	}
	_, err := convertWireToLCSRequestorID(w)
	if !errors.Is(err, ErrUSSDDataCodingSchemeInvalidSize) {
		t.Errorf("want ErrUSSDDataCodingSchemeInvalidSize, got %v", err)
	}
}

// ============================================================================
// LCSClientID
// ============================================================================

func TestLCSClientIDRoundTrip(t *testing.T) {
	internalID := LCSClientBroadcastService
	cases := []struct {
		name string
		in   *LCSClientID
	}{
		{"minimal (LcsClientType only)", &LCSClientID{
			LcsClientType: LCSClientTypeEmergencyServices,
		}},
		{"with internal ID", &LCSClientID{
			LcsClientType:       LCSClientTypeValueAddedServices,
			LcsClientInternalID: &internalID,
		}},
		{"with dialed-by-MS", &LCSClientID{
			LcsClientType:             LCSClientTypePlmnOperatorServices,
			LcsClientDialedByMS:       "31612345678",
			LcsClientDialedByMSNature: 0x10, // International
			LcsClientDialedByMSPlan:   0x01, // ISDN
		}},
		{"with external ID", &LCSClientID{
			LcsClientType: LCSClientTypeLawfulInterceptServices,
			LcsClientExternalID: &LCSClientExternalID{
				ExternalAddress:       "1234567890",
				ExternalAddressNature: 0x10, // International
				ExternalAddressPlan:   0x01, // ISDN
			},
		}},
		{"with name", &LCSClientID{
			LcsClientType: LCSClientTypeEmergencyServices,
			LcsClientName: &LCSClientName{
				DataCodingScheme: 0x0f,
				NameString:       HexBytes{0x41, 0x42, 0x43},
			},
		}},
		{"with APN", &LCSClientID{
			LcsClientType: LCSClientTypeEmergencyServices,
			LcsAPN:        HexBytes{0x03, 'i', 'm', 's', 0x01, 0x49},
		}},
		{"with requestor ID", &LCSClientID{
			LcsClientType: LCSClientTypeEmergencyServices,
			LcsRequestorID: &LCSRequestorID{
				DataCodingScheme:  0x0f,
				RequestorIDString: HexBytes{0x52, 0x49, 0x44},
			},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertLCSClientIDToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := convertWireToLCSClientID(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", tc.in, out)
			}
		})
	}
}

func TestLCSClientIDFullPopulationRoundTrip(t *testing.T) {
	internalID := LCSClientTargetMSsubscribedService
	fmtMsisdn := LCSFormatMsisdn
	in := &LCSClientID{
		LcsClientType: LCSClientTypeValueAddedServices,
		LcsClientExternalID: &LCSClientExternalID{
			ExternalAddress:       "9876543210",
			ExternalAddressNature: 0x10, // International
			ExternalAddressPlan:   0x01, // ISDN
		},
		LcsClientDialedByMS:       "112",
		LcsClientDialedByMSNature: 0x20, // National
		LcsClientDialedByMSPlan:   0x01, // ISDN
		LcsClientInternalID:       &internalID,
		LcsClientName: &LCSClientName{
			DataCodingScheme:   0x00,
			NameString:         HexBytes{0x4e, 0x41, 0x4d, 0x45},
			LcsFormatIndicator: &fmtMsisdn,
		},
		LcsAPN: HexBytes{0x05, 't', 'e', 's', 't', '1'},
		LcsRequestorID: &LCSRequestorID{
			DataCodingScheme:  0x0f,
			RequestorIDString: HexBytes{0x52, 0x49, 0x44},
		},
	}
	wire, err := convertLCSClientIDToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	out, err := convertWireToLCSClientID(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("full round-trip mismatch:\n in=%+v\nout=%+v", in, out)
	}
}

func TestLCSClientIDNilPassesThrough(t *testing.T) {
	wire, err := convertLCSClientIDToWire(nil)
	if err != nil || wire != nil {
		t.Errorf("nil → nil expected, got wire=%v err=%v", wire, err)
	}
	out, err := convertWireToLCSClientID(nil)
	if err != nil || out != nil {
		t.Errorf("nil → nil expected, got out=%v err=%v", out, err)
	}
}

func TestLCSClientIDOutOfRangeTypeRejected(t *testing.T) {
	_, err := convertLCSClientIDToWire(&LCSClientID{
		LcsClientType: LCSClientType(99),
	})
	if !errors.Is(err, ErrLCSClientTypeInvalid) {
		t.Errorf("want ErrLCSClientTypeInvalid, got %v", err)
	}
}

// Per project convention, an explicitly-present wire AddressString that
// decodes to empty digits cannot round-trip through the string-based
// API and must be flagged on decode.
func TestLCSClientIDDialedByMSWireEmptyDigitsRejected(t *testing.T) {
	// Construct a wire AddressString header byte (extension=1,
	// nature=international(1), plan=ISDN(1)) with no TBCD digits.
	emptyAddr := gsm_map.AddressString{0x91}
	w := &gsm_map.LCSClientID{
		LcsClientType:       gsm_map.LCSClientTypeEmergencyServices,
		LcsClientDialedByMS: &emptyAddr,
	}
	_, err := convertWireToLCSClientID(w)
	if !errors.Is(err, ErrLCSClientIDDialedByMSEmpty) {
		t.Errorf("want ErrLCSClientIDDialedByMSEmpty, got %v", err)
	}
}

// Symmetric encode-side check: empty digits combined with non-zero
// Nature/Plan must surface ErrLCSClientIDDialedByMSEmpty rather than
// silently dropping the field.
func TestLCSClientIDDialedByMSEncodeEmptyWithNaturePlanRejected(t *testing.T) {
	_, err := convertLCSClientIDToWire(&LCSClientID{
		LcsClientType:             LCSClientTypeEmergencyServices,
		LcsClientDialedByMS:       "",
		LcsClientDialedByMSNature: 0x10,
	})
	if !errors.Is(err, ErrLCSClientIDDialedByMSEmpty) {
		t.Errorf("Nature set with empty digits: want ErrLCSClientIDDialedByMSEmpty, got %v", err)
	}
	_, err = convertLCSClientIDToWire(&LCSClientID{
		LcsClientType:           LCSClientTypeEmergencyServices,
		LcsClientDialedByMS:     "",
		LcsClientDialedByMSPlan: 0x01,
	})
	if !errors.Is(err, ErrLCSClientIDDialedByMSEmpty) {
		t.Errorf("Plan set with empty digits: want ErrLCSClientIDDialedByMSEmpty, got %v", err)
	}
}

// LcsClientInternalID is a non-extensible enum (0..4); validate
// symmetrically on encode and decode.
func TestLCSClientIDInternalIDOutOfRangeRejected(t *testing.T) {
	bad := LCSClientInternalID(99)
	_, err := convertLCSClientIDToWire(&LCSClientID{
		LcsClientType:       LCSClientTypeEmergencyServices,
		LcsClientInternalID: &bad,
	})
	if !errors.Is(err, ErrLCSClientInternalIDInvalid) {
		t.Errorf("encode: want ErrLCSClientInternalIDInvalid, got %v", err)
	}

	wireBad := gsm_map.LCSClientInternalID(99)
	_, err = convertWireToLCSClientID(&gsm_map.LCSClientID{
		LcsClientType:       gsm_map.LCSClientTypeEmergencyServices,
		LcsClientInternalID: &wireBad,
	})
	if !errors.Is(err, ErrLCSClientInternalIDInvalid) {
		t.Errorf("decode: want ErrLCSClientInternalIDInvalid, got %v", err)
	}
}

// LcsAPN must satisfy APN SIZE(2..63) per TS 29.002 MAP-MS-DataTypes.asn.
// Use the shared validateAPN helper for symmetry with PDPContext etc.
func TestLCSClientIDAPNSizeValidation(t *testing.T) {
	// 1-octet APN is too small (spec minimum is 2).
	_, err := convertLCSClientIDToWire(&LCSClientID{
		LcsClientType: LCSClientTypeEmergencyServices,
		LcsAPN:        HexBytes{0x01},
	})
	if err == nil {
		t.Error("encode: 1-octet APN should be rejected")
	}

	// 64-octet APN is too large.
	tooBig := make(HexBytes, 64)
	_, err = convertLCSClientIDToWire(&LCSClientID{
		LcsClientType: LCSClientTypeEmergencyServices,
		LcsAPN:        tooBig,
	})
	if err == nil {
		t.Error("encode: 64-octet APN should be rejected")
	}

	// Decode-side parity.
	tooSmallWire := gsm_map.APN{0x01}
	_, err = convertWireToLCSClientID(&gsm_map.LCSClientID{
		LcsClientType: gsm_map.LCSClientTypeEmergencyServices,
		LcsAPN:        &tooSmallWire,
	})
	if err == nil {
		t.Error("decode: 1-octet APN should be rejected")
	}
}

// LcsFormatIndicator (extensible enum, encoder strict, decoder lenient)
// — validate the encoder rejects out-of-range values for both
// LCSClientName and LCSRequestorID.
func TestLCSFormatIndicatorEncoderStrict(t *testing.T) {
	bad := LCSFormatIndicator(99)

	_, err := convertLCSClientNameToWire(&LCSClientName{
		DataCodingScheme:   0x0f,
		NameString:         HexBytes{0x41},
		LcsFormatIndicator: &bad,
	})
	if !errors.Is(err, ErrLCSFormatIndicatorInvalid) {
		t.Errorf("LCSClientName encode: want ErrLCSFormatIndicatorInvalid, got %v", err)
	}

	_, err = convertLCSRequestorIDToWire(&LCSRequestorID{
		DataCodingScheme:   0x0f,
		RequestorIDString:  HexBytes{0x41},
		LcsFormatIndicator: &bad,
	})
	if !errors.Is(err, ErrLCSFormatIndicatorInvalid) {
		t.Errorf("LCSRequestorID encode: want ErrLCSFormatIndicatorInvalid, got %v", err)
	}
}
