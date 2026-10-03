package gsmmap

import (
	"errors"
	"fmt"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// 3GPP TS 29.002 V19.1.0 §17.7.1 SS-EventList lists ectSS-Code,
// multiPTYSS-Code, cdSS-Code and ccbsSS-Code, then: "all other SS codes
// shall be ignored", "When SS-CSI is sent to the VLR, it shall not contain a
// marking for ccbs. If the VLR receives SS-CSI containing a marking for
// ccbs, the VLR shall discard the ccbs marking in SS-CSI." The package
// carries SS-CSI only in the VlrCamelSubscriptionInfo of
// InsertSubscriberData, which goes to the VLR.

func ssCSIISD(codes ...SsCode) *InsertSubscriberDataArg {
	a := camelISD()
	a.VlrCamelSubscriptionInfo.SsCSI = &SSCSI{SsEventList: codes, GsmSCFAddress: testGsmSCF}
	return a
}

func TestSSCSIEventListMarshal(t *testing.T) {
	in := ssCSIISD(SsCodeECT, SsCodeMultiPTY, SsCodeCD)
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseInsertSubscriberData(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	wantEqual(t, "SsCSI", in.VlrCamelSubscriptionInfo.SsCSI, got.VlrCamelSubscriptionInfo.SsCSI)

	if _, err := ssCSIISD(SsCodeECT, SsCodeCCBS).Marshal(); !errors.Is(err, ErrSSCSICCBSToVLR) {
		t.Errorf("ccbs: err = %v, want ErrSSCSICCBSToVLR", err)
	}
	for _, c := range []SsCode{0x00, 0x11, 0x21, 0x30, 0x32, 0xff} {
		if _, err := ssCSIISD(SsCodeECT, c).Marshal(); !errors.Is(err, ErrSSEventUnlisted) {
			t.Errorf("SS-Code 0x%02x: err = %v, want ErrSSEventUnlisted", c, err)
		}
	}
}

func TestSSCSIEventListParse(t *testing.T) {
	wire := func(t *testing.T, codes ...byte) []byte {
		w := isdWire(t, ssCSIISD(SsCodeECT))
		w.VlrCamelSubscriptionInfo.SsCSI.SsCamelData.SsEventList.Values = nil
		for _, c := range codes {
			w.VlrCamelSubscriptionInfo.SsCSI.SsCamelData.SsEventList.Values = append(w.VlrCamelSubscriptionInfo.SsCSI.SsCamelData.SsEventList.Values, gsm_map.SSCode{c})
		}
		return strictBER(t, w)
	}
	for _, tc := range []struct {
		codes []byte
		want  []SsCode // nil: no SS-CSI left
	}{
		{[]byte{0x31, 0x51, 0x24}, []SsCode{SsCodeECT, SsCodeMultiPTY, SsCodeCD}},
		{[]byte{0x11, 0x31, 0x44, 0x24, 0x00}, []SsCode{SsCodeECT, SsCodeCD}},
		{[]byte{0x44}, nil},
		{[]byte{0x11, 0x44, 0xff}, nil},
	} {
		t.Run(fmt.Sprintf("%x", tc.codes), func(t *testing.T) {
			data := wire(t, tc.codes...)
			got, err := ParseInsertSubscriberData(data)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			v := got.VlrCamelSubscriptionInfo
			if tc.want == nil {
				if v.SsCSI != nil {
					t.Errorf("SsCSI = %+v, want nil", v.SsCSI)
				}
			} else {
				wantEqual(t, "SsEventList", tc.want, v.SsCSI.SsEventList)
			}
			// The rest of the subscription info is kept.
			wantEqual(t, "OCSI", camelISD().VlrCamelSubscriptionInfo.OCSI, v.OCSI)
			if err := checkParseRoundTrip("InsertSubscriberData", asParser(ParseInsertSubscriberData), data); err != nil {
				t.Error(err)
			}
		})
	}
}
