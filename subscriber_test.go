package gsmmap

import (
	"errors"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// PS-SubscriberState and UserCSGInformation: TS 29.002 V19.1.0 §17.7.1.
func TestSubscriberAlternatives(t *testing.T) {
	ctx := gsm_map.PDPContextInfo{PdpContextIdentifier: 1, PdpType: gsm_map.PDPType{0x01, 0x21}}
	contextBER := HexBytes(strictBER(t, &ctx))
	reason := NotReachableReason(1)
	cases := []struct {
		name  string
		state PsSubscriberState
	}{
		{"detached", PsSubscriberState{PsDetached: true}},
		{"attached-unreachable", PsSubscriberState{PsAttachedNotReachableForPaging: true}},
		{"attached-reachable", PsSubscriberState{PsAttachedReachableForPaging: true}},
		{"pdp-unreachable", PsSubscriberState{PsPDPActiveNotReachableForPaging: []HexBytes{contextBER}}},
		{"pdp-reachable", PsSubscriberState{PsPDPActiveReachableForPaging: []HexBytes{contextBER}}},
		{"network-unreachable", PsSubscriberState{NetDetNotReachable: &reason}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{PsSubscriberState: &tc.state, EpsSubscriberState: &tc.state}}
			data, err := in.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			out, err := ParseAnyTimeInterrogationRes(data)
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "subscriber states", in, out)
		})
		t.Run("multiple-"+tc.name, func(t *testing.T) {
			state := tc.state
			state.NotProvidedFromSGSNorMME = true
			_, err := (&AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{PsSubscriberState: &state}}).Marshal()
			if !errors.Is(err, ErrPsSubscriberStateMultipleAlternatives) {
				t.Fatalf("want %v, got %v", ErrPsSubscriberStateMultipleAlternatives, err)
			}
		})
	}
}

// GPRSMSClass and UserCSGInformation optional octets: TS 29.002 V19.1.0 §17.7.1.
func TestSubscriberOptionalOctets(t *testing.T) {
	csg := &UserCSGInformation{CsgID: HexBytes{0x12, 0x34, 0x56, 0x60}, CsgIDBits: 27, AccessMode: HexBytes{1}, CMI: HexBytes{1}}
	in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{
		GprsMSClass:             &GprsMSClass{MSNetworkCapability: HexBytes{1}, MSRadioAccessCapability: HexBytes{2, 3}},
		LocationInformation:     &CSLocationInformation{UserCSGInformation: csg},
		LocationInformationGPRS: &GPRSLocationInformation{UserCSGInformation: csg},
	}}
	data, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "optional octets", in, out)
}

// IMSI SIZE (3..8) is checked before MNP conversion: TS 29.002 V19.1.0 §17.7.8.
func TestSubscriberEmptyMNPIMSI(t *testing.T) {
	empty := gsm_map.IMSI{}
	w := &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{MnpInfoRes: &gsm_map.MNPInfoRes{Imsi: &empty}}}
	_, err := ParseAnyTimeInterrogationRes(tolerantBER(t, w))
	wantConstraintError(t, err, "subscriberInfo.mnpInfoRes.imsi", "SIZE (3..8)")
}
