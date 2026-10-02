// subscriberdata_test.go
//
// Tests for SubscriberData sub-struct converters (ISD PR B).
// Each converter pair gets a round-trip test covering the minimal
// valid case, the full-featured case, and the relevant validation
// errors.
package gsmmap

import (
	"bytes"
	"errors"
	"testing"

	"github.com/gomaja/go-asn1-gsmmap/tbcd"
	"github.com/gomaja/go-asn1/runtime"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/google/go-cmp/cmp"
)

// gsmMapEmptyZoneCodeList returns a non-nil, zero-length wire ZoneCodeList
// to exercise the decoder's SIZE(1..10) lower-bound check.
func gsmMapEmptyZoneCodeList() *gsm_map.ZoneCodeList {
	return &gsm_map.ZoneCodeList{Values: []gsm_map.ZoneCode{}}
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
				GroupId:                  "abc*#1",
				BroadcastInitEntitlement: true,
			},
		},
		{
			name: "withLongGroupId",
			in: &VoiceBroadcastData{
				GroupId:     "", // six fillers on the wire when LongGroupId is present
				LongGroupId: "1234abc#",
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
	t.Run("emptyGroupIdWithLongId", func(t *testing.T) {
		// TS 29.002: groupId is filled with six TBCD fillers when the
		// longGroupId is present, so an empty GroupId is the valid form.
		w, err := convertVoiceBroadcastDataToWire(&VoiceBroadcastData{LongGroupId: "1234abc#"})
		if err != nil {
			t.Fatalf("encode: %v", err)
		}
		if !bytes.Equal(w.Groupid, []byte{0xff, 0xff, 0xff}) {
			t.Errorf("GroupId wire = %x, want ffffff", []byte(w.Groupid))
		}
	})
	t.Run("nonFillerGroupIdWithLongId", func(t *testing.T) {
		_, err := convertVoiceBroadcastDataToWire(&VoiceBroadcastData{
			GroupId:     "123456",
			LongGroupId: "1234abc#",
		})
		if !errors.Is(err, ErrGroupIdFillerRequired) {
			t.Errorf("want ErrGroupIdFillerRequired, got %v", err)
		}
	})
	t.Run("wrongLengthGroupId", func(t *testing.T) {
		// 7 digits = 4 TBCD octets, but GroupId is TBCD-STRING (SIZE (3)).
		_, err := convertVoiceBroadcastDataToWire(&VoiceBroadcastData{GroupId: "1234567"})
		if !errors.Is(err, ErrGroupIdInvalidEncodedLength) {
			t.Errorf("want ErrGroupIdInvalidEncodedLength, got %v", err)
		}
	})
	t.Run("wrongLengthLongGroupId", func(t *testing.T) {
		// 9 digits = 5 TBCD octets, but Long-GroupId is TBCD-STRING (SIZE (4)).
		_, err := convertVoiceBroadcastDataToWire(&VoiceBroadcastData{
			LongGroupId: "123456789",
		})
		if !errors.Is(err, ErrLongGroupIdInvalidEncodedLength) {
			t.Errorf("want ErrLongGroupIdInvalidEncodedLength, got %v", err)
		}
	})
	t.Run("hexFillerRejected", func(t *testing.T) {
		// 'f' is not a TBCD digit: the filler is a nibble, not a character.
		for _, gid := range []string{"ffffff", "FFFFFF", "12345f", "abcdef"} {
			_, err := convertVoiceBroadcastDataToWire(&VoiceBroadcastData{GroupId: gid})
			if !errors.Is(err, tbcd.ErrInvalidCharacter) {
				t.Errorf("GroupId %q: want tbcd.ErrInvalidCharacter, got %v", gid, err)
			}
		}
	})
	t.Run("shortGroupIdFillerPadded", func(t *testing.T) {
		// "When Group-Id is less than six characters in length, the TBCD
		// filler (1111) is used to fill unused half octets."
		for gid, want := range map[string][]byte{
			"12345": {0x21, 0x43, 0xf5},
			"1234":  {0x21, 0x43, 0xff},
			"1":     {0xf1, 0xff, 0xff},
		} {
			w, err := convertVoiceBroadcastDataToWire(&VoiceBroadcastData{GroupId: gid})
			if err != nil {
				t.Fatalf("GroupId %q: %v", gid, err)
			}
			if !bytes.Equal(w.Groupid, want) {
				t.Errorf("GroupId %q wire = %x, want %x", gid, []byte(w.Groupid), want)
			}
			got, err := convertWireToVoiceBroadcastData(w)
			if err != nil || got.GroupId != gid {
				t.Errorf("GroupId %q round trip = %q, %v", gid, got.GroupId, err)
			}
		}
	})
	t.Run("misplacedFillerOnWire", func(t *testing.T) {
		w := &gsm_map.VoiceBroadcastData{Groupid: []byte{0x21, 0xff, 0x65}}
		if _, err := convertWireToVoiceBroadcastData(w); !errors.Is(err, tbcd.ErrMisplacedFiller) {
			t.Errorf("want tbcd.ErrMisplacedFiller, got %v", err)
		}
		lg := gsm_map.LongGroupId{0x21, 0xff, 0x65, 0x87}
		w = &gsm_map.VoiceBroadcastData{Groupid: []byte{0x21, 0x43, 0x65}, LongGroupId: &lg}
		if _, err := convertWireToVoiceBroadcastData(w); !errors.Is(err, tbcd.ErrMisplacedFiller) {
			t.Errorf("LongGroupId: want tbcd.ErrMisplacedFiller, got %v", err)
		}
	})
	t.Run("allFillerLongGroupIdOnWire", func(t *testing.T) {
		lg := gsm_map.LongGroupId{0xff, 0xff, 0xff, 0xff}
		w := &gsm_map.VoiceBroadcastData{Groupid: []byte{0xff, 0xff, 0xff}, LongGroupId: &lg}
		if _, err := convertWireToVoiceBroadcastData(w); !errors.Is(err, ErrLongGroupIdDecodedEmpty) {
			t.Errorf("want ErrLongGroupIdDecodedEmpty, got %v", err)
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
			GroupId:     "abc*#1",
			LongGroupId: "1234abc#",
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
			LongGroupId: "123456789", // 5 octets, at most 4
		})
		if !errors.Is(err, ErrLongGroupIdInvalidEncodedLength) {
			t.Errorf("want ErrLongGroupIdInvalidEncodedLength, got %v", err)
		}
	})
}

// Per the VoiceGroupCallData.AdditionalInfo godoc, the HexBytes
// representation is byte-aligned only — a wire BIT STRING whose
// BitLength is not a multiple of 8 has its sub-byte trailing bits
// discarded. The decoder uses BitLength/8 (floor), not ceiling.
func TestVoiceGroupCallDataAdditionalInfoSubByteDiscarded(t *testing.T) {
	// Wire carries 9 bits in 2 bytes; floor(9/8) = 1, so only the
	// first byte survives the decode.
	bs := runtime.BitString{Bytes: []byte{0xAB, 0x80}, BitLength: 9}
	w := &gsm_map.VoiceGroupCallData{
		GroupId:        []byte{0x21, 0x43, 0x65}, // TBCD of "123456"
		AdditionalInfo: &bs,
	}
	got, err := convertWireToVoiceGroupCallData(w)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := HexBytes{0xAB}
	if diff := cmp.Diff(want, got.AdditionalInfo); diff != "" {
		t.Errorf("sub-byte truncation (-want +got):\n%s", diff)
	}
}

// A LongGroupId of fewer than eight characters is padded with TBCD filler
// on the wire, and the padding is not part of the value on decode.
func TestLongGroupIdFillerPaddingRoundTrips(t *testing.T) {
	for _, lg := range []string{"1234567", "12345", "1", "*#abc012"} {
		in := &VoiceBroadcastData{LongGroupId: lg}
		wire, err := convertVoiceBroadcastDataToWire(in)
		if err != nil {
			t.Fatalf("%q encode: %v", lg, err)
		}
		if len(*wire.LongGroupId) != LongGroupIdOctets {
			t.Errorf("%q: wire length %d, want %d", lg, len(*wire.LongGroupId), LongGroupIdOctets)
		}
		got, err := convertWireToVoiceBroadcastData(wire)
		if err != nil {
			t.Fatalf("%q decode: %v", lg, err)
		}
		if got.LongGroupId != lg || got.GroupId != "" {
			t.Errorf("round-trip: got GroupId %q LongGroupId %q, want empty and %q", got.GroupId, got.LongGroupId, lg)
		}
	}
}

func TestVBSDataListRoundTrip(t *testing.T) {
	in := VBSDataList{
		{GroupId: "123456"},
		{GroupId: "abc*#1", BroadcastInitEntitlement: true},
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
				GroupId: "abc*#1",
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
				GroupId:     "", // six fillers on the wire
				LongGroupId: "1234abc#",
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
			GroupId:                 "abc*#1",
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
