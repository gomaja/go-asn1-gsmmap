package gsmmap

import (
	"bytes"
	"errors"
	"testing"

	"github.com/gomaja/go-asn1/runtime"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// TS 29.002 V19.1.0 §17.7.8: each FTN-AddressString carries its own
// nature and numbering plan octet.
func TestExtForwFeatureDistinctAddressHeaders(t *testing.T) {
	w := semForwISD(t)
	f := &w.ProvisionedSS.Values[0].ForwardingInfo.ForwardingFeatureList.Values[0]
	short := []byte{0x91, 0x21}
	long := []byte{0xa1, 0x43}
	f.ForwardedToNumber = &short
	f.LongForwardedToNumber = &long
	data, err := w.MarshalBER()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseInsertSubscriberData(data)
	if err != nil {
		t.Fatal(err)
	}
	got, err := parsed.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("long FTN header changed from a1 to 91: got %x, want %x", got, data)
	}
}

func TestForwardingDataLongFTNAddress(t *testing.T) {
	short := []byte{0x91, 0x21}
	long := []byte{0xa1, 0x43}
	eri := gsm_map.NewExtendedRoutingInfoRoutingInfo(gsm_map.NewRoutingInfoForwardingData(gsm_map.ForwardingData{ForwardedToNumber: &short, LongForwardedToNumber: &long}))
	w := &gsm_map.SendRoutingInfoRes{ExtendedRoutingInfo: &eri}
	data, err := w.MarshalBER()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseSriResp(data)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := got.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, data) {
		t.Errorf("ForwardingData FTN headers changed: got %x, want %x", encoded, data)
	}
}

func TestFeatureBitsAcrossMessages(t *testing.T) {
	ext := &ExtSupportedFeatures{UnlicensedSpectrumAsSecondaryRAT: true, UnknownBits: HexBytes{0x40}, BitLength: 40}
	for _, tc := range []struct {
		name    string
		marshal func() ([]byte, error)
		parse   func([]byte) (*ExtSupportedFeatures, []byte, error)
	}{
		{"SGSNCapability", func() ([]byte, error) {
			return (&UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "12", SGSNAddress: "192.0.2.1", SGSNCapability: &SGSNCapability{SupportedFeatures: &SupportedFeatures{OdbAllApn: true, NrAsSecondaryRAT: true}, ExtSupportedFeatures: ext}}).Marshal()
		}, func(b []byte) (*ExtSupportedFeatures, []byte, error) {
			x, e := ParseUpdateGprsLocation(b)
			if e != nil {
				return nil, nil, e
			}
			out, e := x.Marshal()
			return x.SGSNCapability.ExtSupportedFeatures, out, e
		}},
		{"InsertSubscriberDataRes", func() ([]byte, error) { return (&InsertSubscriberDataRes{ExtSupportedFeatures: ext}).Marshal() }, func(b []byte) (*ExtSupportedFeatures, []byte, error) {
			x, e := ParseInsertSubscriberDataRes(b)
			if e != nil {
				return nil, nil, e
			}
			out, e := x.Marshal()
			return x.ExtSupportedFeatures, out, e
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.marshal()
			if err != nil {
				t.Fatal(err)
			}
			got, encoded, err := tc.parse(data)
			if err != nil {
				t.Fatal(err)
			}
			if got.BitLength != 40 || len(got.UnknownBits) != 1 || got.UnknownBits[0] != 0x40 || !got.UnlicensedSpectrumAsSecondaryRAT {
				t.Errorf("ExtSupportedFeatures = %+v", got)
			}
			if !bytes.Equal(encoded, data) {
				t.Errorf("feature bits changed: got %x, want %x", encoded, data)
			}
		})
	}
}

func TestSupportedFeaturesTrailingZeroBits(t *testing.T) {
	w, err := convertUpdateGprsLocationToArg(&UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "12", SGSNAddress: "192.0.2.1", SGSNCapability: &SGSNCapability{SupportedFeatures: &SupportedFeatures{}}})
	if err != nil {
		t.Fatal(err)
	}
	w.SgsnCapability.SupportedFeatures = &runtime.BitString{Bytes: []byte{0, 0, 0, 0, 0}, BitLength: 40}
	data, err := w.MarshalBER()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseUpdateGprsLocation(data)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := got.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, data) {
		t.Errorf("SupportedFeatures length changed: got %x, want %x", encoded, data)
	}
	r, err := convertInsertSubscriberDataResToWire(&InsertSubscriberDataRes{SupportedFeatures: &SupportedFeatures{}})
	if err != nil {
		t.Fatal(err)
	}
	r.SupportedFeatures = &runtime.BitString{Bytes: []byte{0, 0, 0, 0, 0}, BitLength: 40}
	data, err = r.MarshalBER()
	if err != nil {
		t.Fatal(err)
	}
	res, err := ParseInsertSubscriberDataRes(data)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err = res.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, data) {
		t.Errorf("ISD SupportedFeatures length changed: got %x, want %x", encoded, data)
	}
}

func TestGroupIdMissingSentinelPublic(t *testing.T) {
	for _, tc := range []struct {
		name    string
		makeArg func() *InsertSubscriberDataArg
		mutate  func(*gsm_map.InsertSubscriberDataArg)
	}{
		{"VBS", func() *InsertSubscriberDataArg {
			return &InsertSubscriberDataArg{VbsSubscriptionData: VBSDataList{{GroupId: "123"}}}
		}, func(w *gsm_map.InsertSubscriberDataArg) {
			w.VbsSubscriptionData.Values[0].Groupid = gsm_map.GroupId{0xff, 0xff, 0xff}
		}},
		{"VGCS", func() *InsertSubscriberDataArg {
			return &InsertSubscriberDataArg{VgcsSubscriptionData: VGCSDataList{{GroupId: "123"}}}
		}, func(w *gsm_map.InsertSubscriberDataArg) {
			w.VgcsSubscriptionData.Values[0].GroupId = gsm_map.GroupId{0xff, 0xff, 0xff}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			arg := tc.makeArg()
			if tc.name == "VBS" {
				arg.VbsSubscriptionData[0].GroupId = ""
			} else {
				arg.VgcsSubscriptionData[0].GroupId = ""
			}
			if _, err := arg.Marshal(); !errors.Is(err, ErrGroupIdMissingWithoutLong) {
				t.Errorf("Marshal: %v", err)
			}
			w, err := convertInsertSubscriberDataArgToWire(tc.makeArg())
			if err != nil {
				t.Fatal(err)
			}
			tc.mutate(w)
			data, err := w.MarshalBER()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParseInsertSubscriberData(data); !errors.Is(err, ErrGroupIdMissingWithoutLong) {
				t.Errorf("Parse: %v", err)
			}
		})
	}
}

func TestCSGIdPublicConstraint(t *testing.T) {
	for _, bits := range []int{26, 27, 28} {
		arg := &InsertSubscriberDataArg{CsgSubscriptionDataList: CSGSubscriptionDataList{{CsgID: HexBytes{0x12, 0x34, 0x56, 0x60}, CsgIDBits: bits}}}
		data, err := arg.Marshal()
		if bits == 27 {
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParseInsertSubscriberData(data); err != nil {
				t.Fatal(err)
			}
			continue
		}
		wantConstraintError(t, err, "csg-Id", "SIZE (27)")
	}
	user := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformation: &CSLocationInformation{
		VlrNumber: "12", UserCSGInformation: &UserCSGInformation{CsgID: HexBytes{0x12, 0x34, 0x56, 0x60}, CsgIDBits: 26},
	}}}
	_, err := user.Marshal()
	wantConstraintError(t, err, "csg-Id", "SIZE (27)")
}

func TestISDIstAlertTimerUsesInt(t *testing.T) {
	timer := 30
	data, err := (&InsertSubscriberDataArg{IstAlertTimer: &timer}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseInsertSubscriberData(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.IstAlertTimer == nil || *got.IstAlertTimer != timer {
		t.Errorf("IstAlertTimer = %v, want %d", got.IstAlertTimer, timer)
	}
}

func TestSAIMissingIMSISentinelPublic(t *testing.T) {
	if _, err := (&SendAuthenticationInfo{}).Marshal(); !errors.Is(err, ErrIdentityEmpty) {
		t.Errorf("Marshal: %v", err)
	}
	w := &gsm_map.SendAuthenticationInfoArg{Imsi: gsm_map.IMSI{0xff, 0xff, 0xff}, NumberOfRequestedVectors: 1}
	data, err := w.MarshalBER()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseSendAuthenticationInfo(data); !errors.Is(err, ErrIdentityEmpty) {
		t.Errorf("Parse: %v", err)
	}
}

func TestUpdateGprsLocationMissingIMSI(t *testing.T) {
	_, err := (&UpdateGprsLocation{SgsnNumber: "12", SGSNAddress: "192.0.2.1"}).Marshal()
	if !errors.Is(err, ErrIdentityEmpty) {
		t.Fatalf("Marshal: err = %v, want ErrIdentityEmpty", err)
	}
}

// BER does not require the unused bits of a BIT STRING's last octet to be
// zero (ITU-T X.690 §11.2.1 makes it a DER/CER rule), so they carry no
// feature bit.
func TestExtSupportedFeaturesIgnoresPaddingBits(t *testing.T) {
	// InsertSubscriberDataRes ext-SupportedFeatures [10]: one bit, '1'B, with
	// the seven unused bits of the octet set to 0000001.
	wire := []byte{0x30, 0x04, 0x8a, 0x02, 0x07, 0x81}
	got, err := ParseInsertSubscriberDataRes(wire)
	if err != nil {
		t.Fatal(err)
	}
	ext := got.ExtSupportedFeatures
	if ext == nil || !ext.UnlicensedSpectrumAsSecondaryRAT || ext.UnknownBits != nil {
		t.Fatalf("ExtSupportedFeatures = %+v, want only UnlicensedSpectrumAsSecondaryRAT", ext)
	}
	out, err := got.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{0x30, 0x04, 0x8a, 0x02, 0x07, 0x80}; !bytes.Equal(out, want) {
		t.Errorf("Marshal = %x, want %x", out, want)
	}
}

// The unused bits of a BIT STRING's last octet carry no value in BER (ITU-T
// X.690 §11.2.1 makes zero padding a DER/CER rule): a received value with
// them set equals the same value with them clear.
func TestBitStringPaddingBitsCleared(t *testing.T) {
	// 27 bits of CSG-Id 0x12345620 with the five padding bits set.
	padded := runtime.BitString{Bytes: []byte{0x12, 0x34, 0x56, 0x3F}, BitLength: 27}
	want := HexBytes{0x12, 0x34, 0x56, 0x20}

	csg, err := convertWireToCSGSubscriptionData(&gsm_map.CSGSubscriptionData{CsgId: padded})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(csg.CsgID, want) || csg.CsgIDBits != 27 {
		t.Errorf("CSGSubscriptionData.CsgID = %x (%d bits), want %x (27 bits)", csg.CsgID, csg.CsgIDBits, want)
	}

	user := convertWireToUserCSGInformation(&gsm_map.UserCSGInformation{CsgId: padded})
	if !bytes.Equal(user.CsgID, want) || user.CsgIDBits != 27 {
		t.Errorf("UserCSGInformation.CsgID = %x (%d bits), want %x (27 bits)", user.CsgID, user.CsgIDBits, want)
	}

	// ISR-Information, 3 bits '101'B with the five padding bits set.
	isr := runtime.BitString{Bytes: []byte{0xBF}, BitLength: 3}
	eps, err := convertWireToEpsInfo(&gsm_map.EPSInfo{Choice: gsm_map.EPSInfoChoiceIsrInformation, IsrInformation: &isr})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(eps.IsrInformation, HexBytes{0xA0}) || eps.IsrInformationBits != 3 {
		t.Errorf("EpsInfo.IsrInformation = %x (%d bits), want a0 (3 bits)", eps.IsrInformation, eps.IsrInformationBits)
	}
}

// CellGlobalIdOrServiceAreaIdOrLAI is a CHOICE: a location with both a cell
// identity and an LAI cannot be encoded, and none of them is dropped.
func TestLocationCellIdAndLAIBothSet(t *testing.T) {
	cgi := HexBytes{0x00, 0xf1, 0x10, 0x00, 0x01, 0x00, 0x02}
	lai := HexBytes{0x00, 0xf1, 0x10, 0x00, 0x01}
	for _, tc := range []struct {
		name string
		info SubscriberInfo
	}{
		{"CS", SubscriberInfo{LocationInformation: &CSLocationInformation{CellGlobalId: cgi, LAI: lai}}},
		{"GPRS", SubscriberInfo{LocationInformationGPRS: &GPRSLocationInformation{CellGlobalId: cgi, LAI: lai}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := (&AnyTimeInterrogationRes{SubscriberInfo: tc.info}).Marshal()
			if !errors.Is(err, ErrCellGlobalIdOrServiceAreaIdOrLAIMultipleAlternatives) {
				t.Fatalf("Marshal: err = %v, want ErrCellGlobalIdOrServiceAreaIdOrLAIMultipleAlternatives", err)
			}
		})
	}
}

// Marshal sends zero padding bits whatever the caller's octets hold, so a
// value marshals to the same octets before and after a round trip.
func TestBitStringToWireClearsPaddingBits(t *testing.T) {
	bs, err := bitStringToWire("AdditionalInfo", []byte{0xAB, 0xFF}, 9)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bs.Bytes, []byte{0xAB, 0x80}) || bs.BitLength != 9 {
		t.Errorf("bitStringToWire = %x (%d bits), want ab80 (9 bits)", bs.Bytes, bs.BitLength)
	}
	in := &InsertSubscriberDataArg{VgcsSubscriptionData: VGCSDataList{{GroupId: "123456", AdditionalInfo: HexBytes{0xFF}, AdditionalInfoBits: 1}}}
	first, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseInsertSubscriberData(first)
	if err != nil {
		t.Fatal(err)
	}
	second, err := parsed.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Errorf("Marshal → Parse → Marshal changed the octets: %x, then %x", first, second)
	}
}
