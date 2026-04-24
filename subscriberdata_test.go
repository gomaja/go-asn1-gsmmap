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
)

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
	got := convertWireToZoneCodeList(wire)
	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
	}
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
