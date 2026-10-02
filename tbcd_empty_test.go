package gsmmap

import (
	"encoding/hex"
	"errors"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// An identity whose octets are all filler holds no digits. decodeIdentityDigits
// is the single place that rejects it, with ErrIdentityEmpty.
func TestParseAllFillerIdentityRejected(t *testing.T) {
	fill := []byte{0xFF, 0xFF, 0xFF}
	t.Run("CancelLocation imsi", func(t *testing.T) {
		arg := gsm_map.CancelLocationArg{Identity: gsm_map.NewIdentityImsi(gsm_map.IMSI(fill))}
		data, err := arg.MarshalBER()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ParseCancelLocation(data); !errors.Is(err, ErrIdentityEmpty) {
			t.Errorf("err = %v, want ErrIdentityEmpty", err)
		}
	})
	t.Run("MoFsm SM-RP-DA imsi", func(t *testing.T) {
		golden, err := hex.DecodeString(forwardSMFuzzSeeds[0])
		if err != nil {
			t.Fatal(err)
		}
		var arg gsm_map.MOForwardSMArg
		if err := arg.UnmarshalBER(golden); err != nil {
			t.Fatal(err)
		}
		arg.SmRPDA = gsm_map.NewSMRPDAImsi(gsm_map.IMSI(fill))
		data, err := arg.MarshalBER()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ParseMoFsm(data); !errors.Is(err, ErrIdentityEmpty) {
			t.Errorf("err = %v, want ErrIdentityEmpty", err)
		}
	})
	t.Run("ProvideSubscriberLocation imsi and imei", func(t *testing.T) {
		mlc := gsm_map.ISDNAddressString{0x91, 0x13, 0x16, 0x32, 0x54, 0x76, 0x98}
		good := gsm_map.IMSI{0x02, 0x04, 0x08, 0x00, 0x21, 0x43, 0x65, 0xf7}
		empty := gsm_map.IMSI(fill)
		emptyImei := gsm_map.IMEI{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		lt := gsm_map.LocationType{LocationEstimateType: gsm_map.LocationEstimateTypeCurrentLocation}
		for name, w := range map[string]*gsm_map.ProvideSubscriberLocationArg{
			"imsi": {LocationType: lt, MlcNumber: mlc, Imsi: &empty},
			"imei": {LocationType: lt, MlcNumber: mlc, Imsi: &good, Imei: &emptyImei},
		} {
			data, err := w.MarshalBER()
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			if _, err := ParseProvideSubscriberLocation(data); !errors.Is(err, ErrIdentityEmpty) {
				t.Errorf("%s: err = %v, want ErrIdentityEmpty", name, err)
			}
		}
	})
}

func TestIdentityEmptyNilAndEmpty(t *testing.T) {
	for _, in := range [][]byte{nil, {}, {0xff}} {
		if _, err := decodeIdentityDigits(in); !errors.Is(err, ErrIdentityEmpty) {
			t.Errorf("decodeIdentityDigits(%#v) err = %v, want ErrIdentityEmpty", in, err)
		}
	}
}

// An all-filler GroupId is valid only together with a LongGroupId.
func TestWireGroupIdAllFiller(t *testing.T) {
	fill := gsm_map.GroupId{0xff, 0xff, 0xff}
	if _, err := convertWireToVoiceBroadcastData(&gsm_map.VoiceBroadcastData{Groupid: fill}); !errors.Is(err, ErrGroupIdDecodedEmpty) {
		t.Errorf("VBS: err = %v, want ErrGroupIdDecodedEmpty", err)
	}
	if _, err := convertWireToVoiceGroupCallData(&gsm_map.VoiceGroupCallData{GroupId: fill}); !errors.Is(err, ErrGroupIdDecodedEmpty) {
		t.Errorf("VGCS: err = %v, want ErrGroupIdDecodedEmpty", err)
	}
	lg := gsm_map.LongGroupId{0x21, 0x43, 0x65, 0xf7}
	v, err := convertWireToVoiceBroadcastData(&gsm_map.VoiceBroadcastData{Groupid: fill, LongGroupId: &lg})
	if err != nil || v.GroupId != "" || v.LongGroupId != "1234567" {
		t.Errorf("VBS with LongGroupId: %+v, %v", v, err)
	}
	g, err := convertWireToVoiceGroupCallData(&gsm_map.VoiceGroupCallData{GroupId: fill, LongGroupId: &lg})
	if err != nil || g.GroupId != "" || g.LongGroupId != "1234567" {
		t.Errorf("VGCS with LongGroupId: %+v, %v", g, err)
	}
}
