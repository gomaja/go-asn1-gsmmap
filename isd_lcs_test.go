package gsmmap

import (
	"errors"
	"strings"
	"testing"

	"github.com/gomaja/go-tbcd"

	"github.com/gomaja/go-asn1-gsmmap/address"
)

// TS 29.002 V19.1.0 §17.7.1 defines GMLC-List as ISDN-AddressString
// entries (§17.7.8). The public GMLCList contract rejects digitless entries
// on decode as well as encode, and preserves an absent list as nil.
func TestLCSGMLCListParseNegative(t *testing.T) {
	for _, tc := range []struct {
		name, wire string
		want       GMLCList
		err        error
	}{
		{"nil converter", "", nil, nil},
		// ISD contains LCSInformation [22], then GMLC-List [0]. Entry
		// TLVs 3/4/5 imply list content 3/4/5, LCS content 5/6/7 and
		// outer content 7/8/9 respectively. Header 91 has no digits;
		// 1f puts filler before a digit, which TBCD forbids.
		{"digitless entry", "3007b605a003040191", nil, ErrGMLCListEntryEmpty},
		{"misplaced filler", "3008b606a0040402911f", nil, tbcd.ErrMisplacedFiller},
		{"valid entry", "3009b607a0050403912143", GMLCList{{Digits: "1234", Nature: address.NatureInternational, Plan: address.PlanISDN}}, nil},
		// Empty LCSInformation content, so GMLC-List is absent.
		{"nil list", "3002b600", nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wire == "" {
				got, err := convertWireToGMLCList(nil)
				if got != nil || err != nil {
					t.Fatalf("nil list = %+v, %v; want nil, nil", got, err)
				}
				return
			}
			got, err := ParseInsertSubscriberDataArg(mustHex(t, tc.wire))
			if tc.err != nil {
				if got != nil || !errors.Is(err, tc.err) {
					t.Fatalf("Parse = %+v, %v; want nil, %v", got, err, tc.err)
				}
				if !strings.Contains(err.Error(), "GMLCList[0]") {
					t.Errorf("error must name list entry: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			wantEqual(t, "ISD", &InsertSubscriberDataArg{LcsInformation: &LCSInformation{GmlcList: tc.want}}, got)
		})
	}
}

// Extension external clients and PLMN clients: TS 29.002 V19.1.0 §17.7.1.
func TestLCSExtExternalClients(t *testing.T) {
	for _, clients := range []ExtExternalClientList{nil, {makeExternalClient()}} {
		privacy := LCSPrivacyClass{SsCode: 0xb3, SsStatus: HexBytes{1}, ExtExternalClientList: clients}
		if clients != nil {
			privacy.ExternalClientList = ExternalClientList{makeExternalClient(), makeExternalClient(), makeExternalClient(), makeExternalClient(), makeExternalClient()}
		}
		in := &InsertSubscriberDataArg{LcsInformation: &LCSInformation{LcsPrivacyExceptionList: LCSPrivacyExceptionList{privacy, {SsCode: 0xb4, SsStatus: HexBytes{1}, PlmnClientList: PLMNClientList{LCSClientTargetMSsubscribedService}}}}}
		data, err := in.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		out, err := ParseInsertSubscriberDataArg(data)
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "privacy clients", in, out)
	}
}
