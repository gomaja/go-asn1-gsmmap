package gsmmap

import (
	"errors"
	"fmt"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// 3GPP TS 29.002 V19.1.0 §17.7.1 MM-Code lists the CS domain and PS domain
// MM events, then: "If the MSC receives any other MM-code than the ones
// listed above for the CS domain, then the MSC shall ignore that MM-code.
// If the SGSN receives any other MM-code than the ones listed above for the
// PS domain, then the SGSN shall ignore that MM-code." M-CSI goes to the VLR
// (CS domain) and MG-CSI to the SGSN (PS domain).

var (
	csMMCodes = []MMCode{
		MMCodeLocationUpdateInSameVLR, MMCodeLocationUpdateToOtherVLR, MMCodeIMSIAttach,
		MMCodeMSInitiatedIMSIDetach, MMCodeNetworkInitiatedIMSIDetach,
	}
	psMMCodes = []MMCode{
		MMCodeRouteingAreaUpdateInSameSGSN, MMCodeRouteingAreaUpdateToOtherSGSNUpdateFromNewSGSN,
		MMCodeRouteingAreaUpdateToOtherSGSNDisconnectByDetach, MMCodeGPRSAttach, MMCodeMSInitiatedGPRSDetach,
		MMCodeNetworkInitiatedGPRSDetach, MMCodeNetworkInitiatedTransferToMSNotReachableForPaging,
	}
)

func mmCodeISD(m, mg []MMCode) *InsertSubscriberDataArg {
	a := camelISD()
	if m != nil {
		a.VlrCamelSubscriptionInfo.MCSI = &MCSI{MobilityTriggers: m, ServiceKey: 8, GsmSCFAddress: testGsmSCF}
	}
	if mg != nil {
		a.SgsnCAMELSubscriptionInfo.MgCsi = &MGCSI{MobilityTriggers: mg, ServiceKey: 9, GsmSCFAddress: testGsmSCF}
	}
	return a
}

func TestMMCodeMarshal(t *testing.T) {
	in := mmCodeISD(csMMCodes, psMMCodes)
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseInsertSubscriberData(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	wantEqual(t, "MCSI", in.VlrCamelSubscriptionInfo.MCSI, got.VlrCamelSubscriptionInfo.MCSI)
	wantEqual(t, "MgCsi", in.SgsnCAMELSubscriptionInfo.MgCsi, got.SgsnCAMELSubscriptionInfo.MgCsi)

	for _, c := range []MMCode{0x05, 0x7f, 0x80, 0x86, 0xff} {
		if _, err := mmCodeISD([]MMCode{MMCodeIMSIAttach, c}, nil).Marshal(); !errors.Is(err, ErrMCSIMMCodeInvalid) {
			t.Errorf("M-CSI MM-Code 0x%02x: err = %v, want ErrMCSIMMCodeInvalid", c, err)
		}
	}
	for _, c := range []MMCode{0x00, 0x04, 0x7f, 0x87, 0xff} {
		if _, err := mmCodeISD(nil, []MMCode{MMCodeGPRSAttach, c}).Marshal(); !errors.Is(err, ErrMGCSIMMCodeInvalid) {
			t.Errorf("MG-CSI MM-Code 0x%02x: err = %v, want ErrMGCSIMMCodeInvalid", c, err)
		}
	}
}

// mmCodeWire is a camelISD wire whose M-CSI and MG-CSI carry the given
// MobilityTriggers entries.
func mmCodeWire(t *testing.T, m, mg [][]byte) []byte {
	t.Helper()
	w := isdWire(t, mmCodeISD(csMMCodes, psMMCodes))
	w.VlrCamelSubscriptionInfo.MCSI.MobilityTriggers = &gsm_map.MobilityTriggers{Values: m}
	w.SgsnCAMELSubscriptionInfo.MgCsi.MobilityTriggers = &gsm_map.MobilityTriggers{Values: mg}
	return strictBER(t, w)
}

func TestMMCodeParseIgnoresOtherDomain(t *testing.T) {
	octets := func(b ...byte) [][]byte {
		out := make([][]byte, len(b))
		for i := range b {
			out[i] = []byte{b[i]}
		}
		return out
	}
	for _, tc := range []struct {
		name          string
		m, mg         [][]byte
		wantM, wantMG []MMCode // nil: no CSI left
	}{
		{"listed", octets(0x00, 0x04), octets(0x80, 0x86), []MMCode{0x00, 0x04}, []MMCode{0x80, 0x86}},
		{"mixed", octets(0x05, 0x01, 0x83, 0x02), octets(0x02, 0x81, 0x87, 0x85, 0xff), []MMCode{0x01, 0x02}, []MMCode{0x81, 0x85}},
		{"none left", octets(0x80, 0x05), octets(0x00, 0x87), nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := mmCodeWire(t, tc.m, tc.mg)
			got, err := ParseInsertSubscriberData(data)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if m := got.VlrCamelSubscriptionInfo.MCSI; tc.wantM == nil {
				if m != nil {
					t.Errorf("MCSI = %+v, want nil", m)
				}
			} else {
				wantEqual(t, "M-CSI MobilityTriggers", tc.wantM, m.MobilityTriggers)
			}
			if mg := got.SgsnCAMELSubscriptionInfo.MgCsi; tc.wantMG == nil {
				if mg != nil {
					t.Errorf("MgCsi = %+v, want nil", mg)
				}
			} else {
				wantEqual(t, "MG-CSI MobilityTriggers", tc.wantMG, mg.MobilityTriggers)
			}
			wantEqual(t, "OCSI", camelISD().VlrCamelSubscriptionInfo.OCSI, got.VlrCamelSubscriptionInfo.OCSI)
			if err := checkParseRoundTrip("InsertSubscriberData", asParser(ParseInsertSubscriberData), data); err != nil {
				t.Error(err)
			}
		})
	}
}

// MM-Code is OCTET STRING (SIZE (1)); go-asn1 does not enforce SEQUENCE OF
// element SIZE (https://github.com/gomaja/go-asn1/issues/79).
func TestMMCodeParseRejectsWrongSize(t *testing.T) {
	for _, bad := range [][]byte{{}, {0x00, 0x00}} {
		for _, tc := range []struct {
			name  string
			m, mg [][]byte
		}{
			{"M-CSI", [][]byte{{0x00}, bad}, [][]byte{{0x80}}},
			{"MG-CSI", [][]byte{{0x00}}, [][]byte{{0x80}, bad}},
		} {
			t.Run(fmt.Sprintf("%s %d octets", tc.name, len(bad)), func(t *testing.T) {
				if got, err := ParseInsertSubscriberData(mmCodeWire(t, tc.m, tc.mg)); !errors.Is(err, ErrMMCodeInvalidSize) || got != nil {
					t.Errorf("got %v, %v; want ErrMMCodeInvalidSize", got, err)
				}
			})
		}
	}
}
