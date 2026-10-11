package gsmmap

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gomaja/go-asn1-gsmmap/address"
)

// GMLC-List and DestinationNumberList elements have ISDN-AddressString
// SIZE (1..9), 3GPP TS 29.002 V19.1.0 §17.7.8.

// isdnOctets returns an ISDN-AddressString of n octets: the nature/plan
// octet and n-1 octets of digits.
func isdnOctets(n int) []byte {
	b := []byte{0x91}
	for range n - 1 {
		b = append(b, 0x21)
	}
	return b[:n]
}

func gmlcISD(digits string) *InsertSubscriberDataArg {
	return &InsertSubscriberDataArg{LcsInformation: &LCSInformation{
		GmlcList: GMLCList{{Digits: digits, Nature: address.NatureInternational, Plan: address.PlanISDN}},
	}}
}

func destinationCriteria(digits string) []OBcsmCamelTDPCriteria {
	return []OBcsmCamelTDPCriteria{{
		OBcsmTriggerDetectionPoint: OBcsmTriggerCollectedInfo,
		DestinationNumberCriteria: &DestinationNumberCriteria{
			MatchType:             MatchTypeEnabling,
			DestinationNumberList: []ISDNNumber{{Digits: digits, Nature: address.NatureInternational, Plan: address.PlanISDN}},
		},
	}}
}

func destinationISD(digits string) *InsertSubscriberDataArg {
	a := camelISD()
	a.VlrCamelSubscriptionInfo.OBcsmCamelTDPCriteriaList = destinationCriteria(digits)
	return a
}

func destinationSriResp(digits string) *SriResp {
	g := gmscCamel()
	g.OBcsmCamelTDPCriteriaList = destinationCriteria(digits)
	return &SriResp{
		IMSI:                "204080012345678",
		ExtendedRoutingInfo: &ExtendedRoutingInfo{CamelRoutingInfo: &CamelRoutingInfo{GmscCamelSubscriptionInfo: g}},
	}
}

func TestMarshalAddressListEntrySize(t *testing.T) {
	max := strings.Repeat("1", 16) // 8 TBCD octets after nature/plan: 9 octets
	for _, tc := range []struct {
		name string
		msg  func(string) marshaler
		path string
	}{
		{"GMLC-List", func(d string) marshaler { return gmlcISD(d) }, "lcsInformation.gmlc-List.element[0]"},
		{"DestinationNumberList ISD", func(d string) marshaler { return destinationISD(d) }, "vlrCamelSubscriptionInfo.o-BcsmCamelTDP-CriteriaList.element[0].destinationNumberCriteria.destinationNumberList.element[0]"},
		{"DestinationNumberList SRI", func(d string) marshaler { return destinationSriResp(d) }, "extendedRoutingInfo.camelRoutingInfo.gmscCamelSubscriptionInfo.o-BcsmCamelTDP-CriteriaList.element[0].destinationNumberCriteria.destinationNumberList.element[0]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.msg(max).Marshal(); err != nil {
				t.Errorf("16 digits: %v", err)
			}
			for _, n := range []int{17, 19} {
				if _, err := tc.msg(strings.Repeat("1", n)).Marshal(); !matchesConstraint(err, tc.path, "SIZE (1..9)") {
					t.Errorf("%d digits: err = %v, want %v", n, err, tc.path)
				}
			}
		})
	}
}

func TestParseAddressListEntrySize(t *testing.T) {
	gmlcWire := func(t *testing.T, entry []byte) berMarshaler {
		w := isdWire(t, gmlcISD("31612345678"))
		w.LcsInformation.GmlcList.Values[0] = entry
		return w
	}
	isdDestWire := func(t *testing.T, entry []byte) berMarshaler {
		w := isdWire(t, destinationISD("31612345678"))
		w.VlrCamelSubscriptionInfo.OBcsmCamelTDPCriteriaList.Values[0].DestinationNumberCriteria.DestinationNumberList.Values[0] = entry
		return w
	}
	sriDestWire := func(t *testing.T, entry []byte) berMarshaler {
		w := sriRespWire(t, destinationSriResp("31612345678").ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo)
		w.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo.OBcsmCamelTDPCriteriaList.Values[0].DestinationNumberCriteria.DestinationNumberList.Values[0] = entry
		return w
	}
	for _, tc := range []struct {
		name  string
		wire  func(*testing.T, []byte) berMarshaler
		parse func([]byte) (marshaler, error)
		path  string
	}{
		{"GMLC-List", gmlcWire, asParser(ParseInsertSubscriberData), "lcsInformation.gmlc-List.element[0]"},
		{"DestinationNumberList ISD", isdDestWire, asParser(ParseInsertSubscriberData), "vlrCamelSubscriptionInfo.o-BcsmCamelTDP-CriteriaList.element[0].destinationNumberCriteria.destinationNumberList.element[0]"},
		{"DestinationNumberList SRI", sriDestWire, asParser(ParseSriResp), "extendedRoutingInfo.camelRoutingInfo.gmscCamelSubscriptionInfo.o-BcsmCamelTDP-CriteriaList.element[0].destinationNumberCriteria.destinationNumberList.element[0]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.parse(strictBER(t, tc.wire(t, isdnOctets(9)))); err != nil {
				t.Errorf("9 octets: %v", err)
			}
			for _, n := range []int{0, 10, 11} {
				data := tolerantBER(t, tc.wire(t, isdnOctets(n)))
				if got, err := tc.parse(data); !matchesConstraint(err, tc.path, "SIZE (1..9)") || got != nil {
					t.Errorf("%d octets: got %v, %v; want %v", n, got, err, tc.path)
				}
				if err := checkParseRoundTrip(fmt.Sprintf("%s %d", tc.name, n), tc.parse, data); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
