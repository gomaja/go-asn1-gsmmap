// subscriberdata_test.go
//
// Tests for SubscriberData sub-struct converters (ISD PR B).
// Each converter pair gets a round-trip test covering the minimal
// valid case, the full-featured case, and the relevant validation
// errors.
package gsmmap

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// gsmMapEmptyZoneCodeList returns a non-nil, zero-length wire ZoneCodeList
// to exercise the decoder's SIZE(1..10) lower-bound check.
func gsmMapEmptyZoneCodeList() gsm_map.ZoneCodeList {
	return gsm_map.ZoneCodeList{}
}

// --- ODBData ---

func TestODBDataRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *ODBData
	}{
		{
			name: "generalOnly",
			in:   &ODBData{OdbGeneralData: &ODBGeneralData{AllOGCallsBarred: true}},
		},
		{
			name: "generalAndHPLMN",
			in: &ODBData{
				OdbGeneralData: &ODBGeneralData{InternationalOGCallsBarred: true, AllECTBarred: true},
				OdbHPLMNData:   &ODBHPLMNData{PLMNSpecificBarringType3: true},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertODBDataToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			got := convertWireToODBData(wire)
			if diff := cmp.Diff(tc.in, got); diff != "" {
				t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestODBDataValidation(t *testing.T) {
	_, err := convertODBDataToWire(&ODBData{}) // missing general data
	if !errors.Is(err, ErrODBDataMissingGeneralData) {
		t.Errorf("want ErrODBDataMissingGeneralData, got %v", err)
	}
}

// --- ZoneCodeList ---

func TestZoneCodeListRoundTrip(t *testing.T) {
	in := ZoneCodeList{
		ZoneCode{0x12, 0x34},
		ZoneCode{0xab, 0xcd},
	}
	wire, err := convertZoneCodeListToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := convertWireToZoneCodeList(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
	}
}

// Decoder must treat an empty (non-nil) wire list as malformed, per the
// spec's SIZE(1..N) constraint — encode and decode share the same bounds.
func TestZoneCodeListDecoderEnforcesBounds(t *testing.T) {
	t.Run("nilReturnsNil", func(t *testing.T) {
		got, err := convertWireToZoneCodeList(nil)
		if err != nil || got != nil {
			t.Errorf("nil wire: got (%v, %v), want (nil, nil)", got, err)
		}
	})
	t.Run("emptyNonNil", func(t *testing.T) {
		_, err := convertWireToZoneCodeList(gsmMapEmptyZoneCodeList())
		if !errors.Is(err, ErrZoneCodeListInvalidSize) {
			t.Errorf("want ErrZoneCodeListInvalidSize, got %v", err)
		}
	})
}

func TestZoneCodeListValidation(t *testing.T) {
	t.Run("emptyList", func(t *testing.T) {
		if _, err := convertZoneCodeListToWire(nil); !errors.Is(err, ErrZoneCodeListInvalidSize) {
			t.Errorf("want ErrZoneCodeListInvalidSize, got %v", err)
		}
	})
	t.Run("tooManyEntries", func(t *testing.T) {
		big := make(ZoneCodeList, MaxNumOfZoneCodes+1)
		for i := range big {
			big[i] = ZoneCode{0, 0}
		}
		if _, err := convertZoneCodeListToWire(big); !errors.Is(err, ErrZoneCodeListInvalidSize) {
			t.Errorf("want ErrZoneCodeListInvalidSize, got %v", err)
		}
	})
	t.Run("shortEntry", func(t *testing.T) {
		bad := ZoneCodeList{ZoneCode{0x12}} // 1 octet, need 2
		if _, err := convertZoneCodeListToWire(bad); !errors.Is(err, ErrZoneCodeInvalidSize) {
			t.Errorf("want ErrZoneCodeInvalidSize, got %v", err)
		}
	})
}

// --- VoiceBroadcastData / VBSDataList ---

func TestVoiceBroadcastDataRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *VoiceBroadcastData
	}{
		{
			name: "groupIdOnly",
			in:   &VoiceBroadcastData{GroupId: "123456"},
		},
		{
			name: "withEntitlement",
			in: &VoiceBroadcastData{
				GroupId:                  "abcdef",
				BroadcastInitEntitlement: true,
			},
		},
		{
			name: "withLongGroupId",
			in: &VoiceBroadcastData{
				GroupId:     "ffffff", // filler required when LongGroupId is present
				LongGroupId: "1234abcd",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertVoiceBroadcastDataToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			got, err := convertWireToVoiceBroadcastData(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if diff := cmp.Diff(tc.in, got); diff != "" {
				t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestVoiceBroadcastDataValidation(t *testing.T) {
	t.Run("missingGroupId", func(t *testing.T) {
		_, err := convertVoiceBroadcastDataToWire(&VoiceBroadcastData{})
		if !errors.Is(err, ErrGroupIdMissingWithoutLong) {
			t.Errorf("want ErrGroupIdMissingWithoutLong, got %v", err)
		}
	})
	t.Run("missingFillerWithLongId", func(t *testing.T) {
		_, err := convertVoiceBroadcastDataToWire(&VoiceBroadcastData{LongGroupId: "1234abcd"})
		if !errors.Is(err, ErrGroupIdFillerRequired) {
			t.Errorf("want ErrGroupIdFillerRequired, got %v", err)
		}
	})
	t.Run("nonFillerGroupIdWithLongId", func(t *testing.T) {
		_, err := convertVoiceBroadcastDataToWire(&VoiceBroadcastData{
			GroupId:     "123456",
			LongGroupId: "1234abcd",
		})
		if !errors.Is(err, ErrGroupIdFillerRequired) {
			t.Errorf("want ErrGroupIdFillerRequired, got %v", err)
		}
	})
	t.Run("wrongLengthGroupId", func(t *testing.T) {
		// 4 hex chars = 2 TBCD octets, but spec demands exactly 3.
		_, err := convertVoiceBroadcastDataToWire(&VoiceBroadcastData{GroupId: "1234"})
		if !errors.Is(err, ErrGroupIdInvalidEncodedLength) {
			t.Errorf("want ErrGroupIdInvalidEncodedLength, got %v", err)
		}
	})
	t.Run("wrongLengthLongGroupId", func(t *testing.T) {
		// 6 hex chars = 3 TBCD octets, but spec demands exactly 4.
		_, err := convertVoiceBroadcastDataToWire(&VoiceBroadcastData{
			GroupId:     "ffffff",
			LongGroupId: "123456",
		})
		if !errors.Is(err, ErrLongGroupIdInvalidEncodedLength) {
			t.Errorf("want ErrLongGroupIdInvalidEncodedLength, got %v", err)
		}
	})
	t.Run("fillerGroupIdCaseInsensitive", func(t *testing.T) {
		// Spec filler is six 'f' nibbles; accept uppercase too.
		_, err := convertVoiceBroadcastDataToWire(&VoiceBroadcastData{
			GroupId:     "FFFFFF",
			LongGroupId: "1234abcd",
		})
		if err != nil {
			t.Errorf("FFFFFF filler should be accepted case-insensitively: %v", err)
		}
	})
}

func TestVoiceGroupCallDataValidation(t *testing.T) {
	t.Run("missingGroupId", func(t *testing.T) {
		_, err := convertVoiceGroupCallDataToWire(&VoiceGroupCallData{})
		if !errors.Is(err, ErrGroupIdMissingWithoutLong) {
			t.Errorf("want ErrGroupIdMissingWithoutLong, got %v", err)
		}
	})
	t.Run("nonFillerGroupIdWithLongId", func(t *testing.T) {
		_, err := convertVoiceGroupCallDataToWire(&VoiceGroupCallData{
			GroupId:     "abcdef",
			LongGroupId: "1234abcd",
		})
		if !errors.Is(err, ErrGroupIdFillerRequired) {
			t.Errorf("want ErrGroupIdFillerRequired, got %v", err)
		}
	})
	t.Run("additionalInfoTooLong", func(t *testing.T) {
		big := make(HexBytes, MaxAdditionalInfoOctets+1)
		_, err := convertVoiceGroupCallDataToWire(&VoiceGroupCallData{
			GroupId:        "123456",
			AdditionalInfo: big,
		})
		if !errors.Is(err, ErrAdditionalInfoTooLong) {
			t.Errorf("want ErrAdditionalInfoTooLong, got %v", err)
		}
	})
	t.Run("additionalInfoAtBoundary", func(t *testing.T) {
		// Exactly MaxAdditionalInfoOctets bytes must be accepted.
		ok := make(HexBytes, MaxAdditionalInfoOctets)
		_, err := convertVoiceGroupCallDataToWire(&VoiceGroupCallData{
			GroupId:        "123456",
			AdditionalInfo: ok,
		})
		if err != nil {
			t.Errorf("%d-octet AdditionalInfo should be accepted: %v", MaxAdditionalInfoOctets, err)
		}
	})
	t.Run("wrongLengthLongGroupId", func(t *testing.T) {
		_, err := convertVoiceGroupCallDataToWire(&VoiceGroupCallData{
			GroupId:     "ffffff",
			LongGroupId: "123456", // 3 octets, need 4
		})
		if !errors.Is(err, ErrLongGroupIdInvalidEncodedLength) {
			t.Errorf("want ErrLongGroupIdInvalidEncodedLength, got %v", err)
		}
	})
}

// LongGroupId's trailing 'f' nibble must survive the round-trip — the
// raw nibble-swap decoder must not strip it (tbcd.Decode would).
func TestLongGroupIdTrailingFRoundTrips(t *testing.T) {
	// LongGroupId = "1234567f" ends with a legitimate 'f' nibble;
	// tbcd.Decode would have silently turned this into "1234567".
	in := &VoiceBroadcastData{
		GroupId:     "ffffff",
		LongGroupId: "1234567f",
	}
	wire, err := convertVoiceBroadcastDataToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := convertWireToVoiceBroadcastData(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.LongGroupId != in.LongGroupId {
		t.Errorf("LongGroupId round-trip: got %q, want %q", got.LongGroupId, in.LongGroupId)
	}
}

func TestVBSDataListRoundTrip(t *testing.T) {
	in := VBSDataList{
		{GroupId: "123456"},
		{GroupId: "abcdef", BroadcastInitEntitlement: true},
	}
	wire, err := convertVBSDataListToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := convertWireToVBSDataList(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
	}
}

func TestVBSDataListValidation(t *testing.T) {
	t.Run("emptyList", func(t *testing.T) {
		if _, err := convertVBSDataListToWire(nil); !errors.Is(err, ErrVBSDataListInvalidSize) {
			t.Errorf("want ErrVBSDataListInvalidSize, got %v", err)
		}
	})
	t.Run("tooManyEntries", func(t *testing.T) {
		big := make(VBSDataList, MaxNumOfVBSGroupIds+1)
		for i := range big {
			big[i] = VoiceBroadcastData{GroupId: "123456"}
		}
		if _, err := convertVBSDataListToWire(big); !errors.Is(err, ErrVBSDataListInvalidSize) {
			t.Errorf("want ErrVBSDataListInvalidSize, got %v", err)
		}
	})
}

// --- VoiceGroupCallData / VGCSDataList ---

func TestVoiceGroupCallDataRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *VoiceGroupCallData
	}{
		{
			name: "groupIdOnly",
			in:   &VoiceGroupCallData{GroupId: "123456"},
		},
		{
			name: "withAdditionalSubscriptions",
			in: &VoiceGroupCallData{
				GroupId: "abcdef",
				AdditionalSubscriptions: &AdditionalSubscriptions{
					PrivilegedUplinkRequest: true,
					EmergencyReset:          true,
				},
			},
		},
		{
			name: "withAdditionalInfo",
			in: &VoiceGroupCallData{
				GroupId:        "123456",
				AdditionalInfo: HexBytes{0x80, 0x40, 0x20},
			},
		},
		{
			name: "withLongGroupId",
			in: &VoiceGroupCallData{
				GroupId:     "ffffff", // filler
				LongGroupId: "1234abcd",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertVoiceGroupCallDataToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			got, err := convertWireToVoiceGroupCallData(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if diff := cmp.Diff(tc.in, got); diff != "" {
				t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestVGCSDataListRoundTrip(t *testing.T) {
	in := VGCSDataList{
		{GroupId: "123456"},
		{
			GroupId:                 "abcdef",
			AdditionalSubscriptions: &AdditionalSubscriptions{EmergencyUplinkRequest: true},
		},
	}
	wire, err := convertVGCSDataListToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := convertWireToVGCSDataList(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
	}
}

func TestVGCSDataListValidation(t *testing.T) {
	t.Run("emptyList", func(t *testing.T) {
		if _, err := convertVGCSDataListToWire(nil); !errors.Is(err, ErrVGCSDataListInvalidSize) {
			t.Errorf("want ErrVGCSDataListInvalidSize, got %v", err)
		}
	})
	t.Run("tooManyEntries", func(t *testing.T) {
		big := make(VGCSDataList, MaxNumOfVGCSGroupIds+1)
		for i := range big {
			big[i] = VoiceGroupCallData{GroupId: "123456"}
		}
		if _, err := convertVGCSDataListToWire(big); !errors.Is(err, ErrVGCSDataListInvalidSize) {
			t.Errorf("want ErrVGCSDataListInvalidSize, got %v", err)
		}
	})
}

// --- AdditionalSubscriptions BIT STRING ---

func TestAdditionalSubscriptionsRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *AdditionalSubscriptions
	}{
		{name: "empty", in: &AdditionalSubscriptions{}},
		{name: "allSet", in: &AdditionalSubscriptions{true, true, true}},
		{name: "onlyPrivileged", in: &AdditionalSubscriptions{PrivilegedUplinkRequest: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bs := convertAdditionalSubscriptionsToBitString(tc.in)
			if bs.BitLength < 3 {
				t.Errorf("BitLength=%d, want >=3 (spec min)", bs.BitLength)
			}
			got := convertBitStringToAdditionalSubscriptions(bs)
			if diff := cmp.Diff(tc.in, got); diff != "" {
				t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
