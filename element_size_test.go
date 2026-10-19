package gsmmap

import (
	"bytes"
	"fmt"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// elementSizeParser adapts a Parse function to the table: a nil result
// becomes a nil interface, and a result returned with an error is kept.
func elementSizeParser[T any](parse func([]byte) (*T, error)) func([]byte) (any, error) {
	return func(data []byte) (any, error) {
		got, err := parse(data)
		if got == nil {
			return nil, err
		}
		return got, err
	}
}

// SS-Code has SIZE (1), 3GPP TS 29.002 V19.1.0 §17.7.5.
// SS-EventList is in §17.7.1; SendRoutingInfoRes is in §17.7.3.
func TestSSCodeElementSize(t *testing.T) {
	ssCSI := func(codes []SsCode) *InsertSubscriberDataArg {
		return &InsertSubscriberDataArg{VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{
			SsCSI: &SSCSI{SsEventList: codes, GsmSCFAddress: testGsmSCF},
		}}
	}
	for _, tc := range []struct {
		name  string
		msg   func([]SsCode) marshaler
		wire  func(*testing.T, []byte) berMarshaler
		parse func([]byte) (any, error)
		path  string
	}{
		{
			name: "SS-CSI ss-EventList",
			msg:  func(codes []SsCode) marshaler { return ssCSI(codes) },
			wire: func(t *testing.T, bad []byte) berMarshaler {
				w := isdWire(t, ssCSI([]SsCode{SsCodeECT, SsCodeCD}))
				w.VlrCamelSubscriptionInfo.SsCSI.SsCamelData.SsEventList.Values[1] = bad
				return w
			},
			parse: elementSizeParser(ParseInsertSubscriberData),
			path:  "vlrCamelSubscriptionInfo.ss-CSI.ss-CamelData.ss-EventList.element[1]",
		},
		{
			name: "InsertSubscriberDataRes ss-List",
			msg:  func(codes []SsCode) marshaler { return &InsertSubscriberDataRes{SsList: codes} },
			wire: func(_ *testing.T, bad []byte) berMarshaler {
				return &gsm_map.InsertSubscriberDataRes{SsList: &gsm_map.SSList{Values: []gsm_map.SSCode{{byte(SsCodeECT)}, bad}}}
			},
			parse: elementSizeParser(ParseInsertSubscriberDataRes),
			path:  "ss-List.element[1]",
		},
		{
			name: "SendRoutingInfoRes ss-List",
			msg:  func(codes []SsCode) marshaler { return &SriResp{SsList: codes} },
			wire: func(_ *testing.T, bad []byte) berMarshaler {
				return &gsm_map.SendRoutingInfoRes{SsList: &gsm_map.SSList{Values: []gsm_map.SSCode{{byte(SsCodeECT)}, bad}}}
			},
			parse: elementSizeParser(ParseSriResp),
			path:  "ss-List.element[1]",
		},
		{
			name: "SendRoutingInfoRes ss-List2",
			msg:  func(codes []SsCode) marshaler { return &SriResp{SsList2: codes} },
			wire: func(_ *testing.T, bad []byte) berMarshaler {
				return &gsm_map.SendRoutingInfoRes{SsList2: &gsm_map.SSList{Values: []gsm_map.SSCode{{byte(SsCodeECT)}, bad}}}
			},
			parse: elementSizeParser(ParseSriResp),
			path:  "ss-List2.element[1]",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("round trip", func(t *testing.T) {
				in := tc.msg([]SsCode{SsCodeECT, SsCodeCD})
				data, err := in.Marshal()
				if err != nil {
					t.Fatalf("Marshal: %v", err)
				}
				got, err := tc.parse(data)
				if err != nil {
					t.Fatalf("Parse: %v", err)
				}
				wantEqual[any](t, "message", in, got)
			})
			for _, n := range []int{0, 2} {
				t.Run(fmt.Sprintf("decode %d octets", n), func(t *testing.T) {
					data := tolerantBER(t, tc.wire(t, bytes.Repeat([]byte{byte(SsCodeCD)}, n)))
					got, err := tc.parse(data)
					wantValidationError(t, err, tc.path, "SIZE (1)")
					if got != nil {
						t.Errorf("Parse = %v, want nil", got)
					}
				})
			}
		})
	}
}

// EPLMN-List and AdjacentPLMN-List are in 3GPP TS 29.002 V19.1.0 §17.7.1;
// their PLMN-Id elements have SIZE (3), §17.7.8.
func TestPLMNIDElementSize(t *testing.T) {
	updateLocation := func(entries []HexBytes) marshaler {
		return &UpdateLocation{
			IMSI: "001010000000001", MscNumber: "1234", VlrNumber: "5678", EplmnList: entries,
		}
	}
	updateGprsLocation := func(entries []HexBytes, adjacent bool) marshaler {
		m := &UpdateGprsLocation{IMSI: "001010000000001", SgsnNumber: "1234", SGSNAddress: "192.0.2.1"}
		if adjacent {
			m.AdjacentPLMNList = entries
		} else {
			m.EplmnList = entries
		}
		return m
	}
	for _, tc := range []struct {
		name  string
		msg   func([]HexBytes) marshaler
		wire  func([]byte) berMarshaler
		parse func([]byte) (any, error)
		path  string
	}{
		{
			name: "UpdateLocation EPLMN-List",
			msg:  updateLocation,
			wire: func(bad []byte) berMarshaler {
				return &gsm_map.UpdateLocationArg{
					Imsi:      gsm_map.IMSI{0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0xf1},
					MscNumber: gsm_map.ISDNAddressString{0x91, 0x21, 0x43},
					VlrNumber: gsm_map.ISDNAddressString{0x91, 0x65, 0x87},
					EplmnList: &gsm_map.EPLMNList{Values: []gsm_map.PLMNId{{0x21, 0xf3, 0x54}, bad}},
				}
			},
			parse: elementSizeParser(ParseUpdateLocation),
			path:  "eplmn-List.element[1]",
		},
		{
			name: "UpdateGprsLocation EPLMN-List",
			msg:  func(entries []HexBytes) marshaler { return updateGprsLocation(entries, false) },
			wire: func(bad []byte) berMarshaler {
				return &gsm_map.UpdateGprsLocationArg{
					Imsi:        gsm_map.IMSI{0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0xf1},
					SgsnNumber:  gsm_map.ISDNAddressString{0x91, 0x21, 0x43},
					SgsnAddress: gsm_map.GSNAddress{0x04, 192, 0, 2, 1},
					EplmnList:   &gsm_map.EPLMNList{Values: []gsm_map.PLMNId{{0x21, 0xf3, 0x54}, bad}},
				}
			},
			parse: elementSizeParser(ParseUpdateGprsLocation),
			path:  "eplmn-List.element[1]",
		},
		{
			name: "UpdateGprsLocation AdjacentPLMN-List",
			msg:  func(entries []HexBytes) marshaler { return updateGprsLocation(entries, true) },
			wire: func(bad []byte) berMarshaler {
				return &gsm_map.UpdateGprsLocationArg{
					Imsi:             gsm_map.IMSI{0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0xf1},
					SgsnNumber:       gsm_map.ISDNAddressString{0x91, 0x21, 0x43},
					SgsnAddress:      gsm_map.GSNAddress{0x04, 192, 0, 2, 1},
					AdjacentPLMNList: &gsm_map.AdjacentPLMNList{Values: []gsm_map.PLMNId{{0x21, 0xf3, 0x54}, bad}},
				}
			},
			parse: elementSizeParser(ParseUpdateGprsLocation),
			path:  "adjacentPLMN-List.element[1]",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("round trip", func(t *testing.T) {
				in := tc.msg([]HexBytes{{0x21, 0xf3, 0x54}, {0x13, 0xf0, 0x62}})
				data, err := in.Marshal()
				if err != nil {
					t.Fatalf("Marshal: %v", err)
				}
				got, err := tc.parse(data)
				if err != nil {
					t.Fatalf("Parse: %v", err)
				}
				wantEqual[any](t, "message", in, got)
			})
			for _, n := range []int{0, 2, 4} {
				bad := bytes.Repeat([]byte{0x21}, n)
				t.Run(fmt.Sprintf("encode %d octets", n), func(t *testing.T) {
					_, err := tc.msg([]HexBytes{{0x21, 0xf3, 0x54}, bad}).Marshal()
					wantValidationError(t, err, tc.path, "SIZE (3)")
				})
				t.Run(fmt.Sprintf("decode %d octets", n), func(t *testing.T) {
					got, err := tc.parse(tolerantBER(t, tc.wire(bad)))
					wantValidationError(t, err, tc.path, "SIZE (3)")
					if got != nil {
						t.Errorf("Parse = %v, want nil", got)
					}
				})
			}
		})
	}
}
