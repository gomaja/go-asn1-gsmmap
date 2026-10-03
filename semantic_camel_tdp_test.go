// semantic_camel_tdp_test.go
//
// 3GPP TS 29.002 V19.1.0 §17.7.1 allows one entry per trigger detection
// point in each CAMEL TDP data list:
//
//	"O-BcsmCamelTDPDataList shall not contain more than one instance of
//	O-BcsmCamelTDPData containing the same value for
//	o-BcsmTriggerDetectionPoint."
//
// and the same for T-BcsmCamelTDPDataList (t-BcsmTriggerDetectionPoint),
// SMS-CAMEL-TDP-DataList (sms-TriggerDetectionPoint) and
// GPRS-CamelTDPDataList (gprs-TriggerDetectionPoint). Marshal rejects a
// repeated TDP; Parse rejects one among the entries the receiver keeps, so
// an entry it ignores for an unlisted TDP does not count.
package gsmmap

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// tdpListCase is one TDP data list in one message that carries it. marshal
// sets the list to one entry per TDP and marshals the message. parse sets
// the wire list of a valid message to one entry per TDP, parses the strict
// BER and returns the TDPs of the decoded list.
type tdpListCase struct {
	name    string
	field   string  // the list name errors carry
	keep    []int64 // TDPs the receiver keeps, as many as the list admits
	ignored int64   // a TDP the receiver ignores the whole entry for
	marshal func(tdps []int64) error
	parse   func(t *testing.T, tdps []int64) ([]int64, error)
}

func sriRespGmsc(g GmscCamelSubscriptionInfo) *SriResp {
	return &SriResp{
		IMSI:                "204080012345678",
		ExtendedRoutingInfo: &ExtendedRoutingInfo{CamelRoutingInfo: &CamelRoutingInfo{GmscCamelSubscriptionInfo: g}},
	}
}

func oTDPList(tdps []int64) []OBcsmCamelTDPData {
	out := make([]OBcsmCamelTDPData, len(tdps))
	for i, tdp := range tdps {
		out[i] = oTDPData(OBcsmTriggerDetectionPoint(tdp), int64(i+1))
	}
	return out
}

func tTDPList(tdps []int64) []TBcsmCamelTDPData {
	out := make([]TBcsmCamelTDPData, len(tdps))
	for i, tdp := range tdps {
		out[i] = tTDPData(TBcsmTriggerDetectionPoint(tdp), int64(i+1))
	}
	return out
}

func smsTDPList(tdps []int64) []SMSCAMELTDPData {
	out := make([]SMSCAMELTDPData, len(tdps))
	for i, tdp := range tdps {
		out[i] = smsTDPData(SMSTriggerDetectionPoint(tdp), int64(i+1))
	}
	return out
}

func gprsTDPList(tdps []int64) GPRSCamelTDPDataList {
	out := make(GPRSCamelTDPDataList, len(tdps))
	for i, tdp := range tdps {
		out[i] = gprsTDPData(GPRSTriggerDetectionPoint(tdp), int64(i+1))
	}
	return out
}

// wireTDPList returns one copy of first per TDP, with that TDP set by
// setTDP and a distinct service key.
func wireTDPList[W any](first W, tdps []int64, set func(*W, int64, int64)) []W {
	out := make([]W, len(tdps))
	for i, tdp := range tdps {
		out[i] = first
		set(&out[i], tdp, int64(i+1))
	}
	return out
}

func oTDPs(l []OBcsmCamelTDPData) []int64 {
	var out []int64
	for _, d := range l {
		out = append(out, int64(d.OBcsmTriggerDetectionPoint))
	}
	return out
}

func tTDPs(l []TBcsmCamelTDPData) []int64 {
	var out []int64
	for _, d := range l {
		out = append(out, int64(d.TBcsmTriggerDetectionPoint))
	}
	return out
}

func smsTDPs(l []SMSCAMELTDPData) []int64 {
	var out []int64
	for _, d := range l {
		out = append(out, int64(d.SmsTriggerDetectionPoint))
	}
	return out
}

func setOWire(e *gsm_map.OBcsmCamelTDPData, tdp, sk int64) {
	e.OBcsmTriggerDetectionPoint, e.ServiceKey = OBcsmTriggerDetectionPoint(tdp), sk
}

func setTWire(e *gsm_map.TBcsmCamelTDPData, tdp, sk int64) {
	e.TBcsmTriggerDetectionPoint, e.ServiceKey = TBcsmTriggerDetectionPoint(tdp), sk
}

func setSMSWire(e *gsm_map.SMSCAMELTDPData, tdp, sk int64) {
	e.SmsTriggerDetectionPoint, e.ServiceKey = SMSTriggerDetectionPoint(tdp), sk
}

func setGPRSWire(e *gsm_map.GPRSCamelTDPData, tdp, sk int64) {
	e.GprsTriggerDetectionPoint, e.ServiceKey = gsm_map.GPRSTriggerDetectionPoint(tdp), sk
}

func parseISDBytes(t *testing.T, w *gsm_map.InsertSubscriberDataArg) (*InsertSubscriberDataArg, error) {
	t.Helper()
	return ParseInsertSubscriberData(strictBER(t, w))
}

// smsTDPListCase is the SMS-CAMEL-TDP-DataList of one SMS-CSI of a
// camelISD. Only keep is listed there, so the list holds one entry; the
// other SMS TDP is ignored in this CSI.
func smsTDPListCase(name string, keep, ignored SMSTriggerDetectionPoint,
	csi func(a *InsertSubscriberDataArg) *SMSCSI,
	wire func(w *gsm_map.InsertSubscriberDataArg) *gsm_map.SMSCSI,
) tdpListCase {
	return tdpListCase{
		name: name, field: "SmsCAMELTDPDataList",
		keep: []int64{int64(keep)}, ignored: int64(ignored),
		marshal: func(tdps []int64) error {
			a := camelISD()
			csi(a).SmsCAMELTDPDataList = smsTDPList(tdps)
			_, err := a.Marshal()
			return err
		},
		parse: func(t *testing.T, tdps []int64) ([]int64, error) {
			w := isdWire(t, camelISD())
			l := wire(w).SmsCAMELTDPDataList
			l.Values = wireTDPList(l.Values[0], tdps, setSMSWire)
			got, err := parseISDBytes(t, w)
			if err != nil {
				return nil, err
			}
			return smsTDPs(csi(got).SmsCAMELTDPDataList), nil
		},
	}
}

func tdpListCases() []tdpListCase {
	collected, delivery := SMSTriggerDetectionPointSmsCollectedInfo, SMSTriggerDetectionPointSmsDeliveryRequest
	return []tdpListCase{
		{
			name: "VLR O-CSI", field: "OBcsmCamelTDPDataList",
			keep:    []int64{int64(OBcsmTriggerCollectedInfo), int64(OBcsmTriggerRouteSelectFailure)},
			ignored: 3,
			marshal: func(tdps []int64) error {
				a := camelISD()
				a.VlrCamelSubscriptionInfo.OCSI.OBcsmCamelTDPDataList = oTDPList(tdps)
				_, err := a.Marshal()
				return err
			},
			parse: func(t *testing.T, tdps []int64) ([]int64, error) {
				w := isdWire(t, camelISD())
				l := w.VlrCamelSubscriptionInfo.OCSI.OBcsmCamelTDPDataList
				l.Values = wireTDPList(l.Values[0], tdps, setOWire)
				got, err := parseISDBytes(t, w)
				if err != nil {
					return nil, err
				}
				return oTDPs(got.VlrCamelSubscriptionInfo.OCSI.OBcsmCamelTDPDataList), nil
			},
		},
		{
			name: "GMSC O-CSI", field: "OBcsmCamelTDPDataList",
			keep:    []int64{int64(OBcsmTriggerCollectedInfo), int64(OBcsmTriggerRouteSelectFailure)},
			ignored: 3,
			marshal: func(tdps []int64) error {
				g := gmscCamel()
				g.OCSI.OBcsmCamelTDPDataList = oTDPList(tdps)
				_, err := sriRespGmsc(g).Marshal()
				return err
			},
			parse: func(t *testing.T, tdps []int64) ([]int64, error) {
				w := sriRespWire(t, gmscCamel())
				l := w.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo.OCSI.OBcsmCamelTDPDataList
				l.Values = wireTDPList(l.Values[0], tdps, setOWire)
				got, err := ParseSriResp(strictBER(t, w))
				if err != nil {
					return nil, err
				}
				return oTDPs(got.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo.OCSI.OBcsmCamelTDPDataList), nil
			},
		},
		{
			name: "VLR VT-CSI", field: "TBcsmCamelTDPDataList",
			keep:    []int64{int64(TBcsmTriggerTermAttemptAuthorized), int64(TBcsmTriggerTBusy), int64(TBcsmTriggerTNoAnswer)},
			ignored: 15,
			marshal: func(tdps []int64) error {
				a := camelISD()
				a.VlrCamelSubscriptionInfo.VtCSI.TBcsmCamelTDPDataList = tTDPList(tdps)
				_, err := a.Marshal()
				return err
			},
			parse: func(t *testing.T, tdps []int64) ([]int64, error) {
				w := isdWire(t, camelISD())
				l := w.VlrCamelSubscriptionInfo.VtCSI.TBcsmCamelTDPDataList
				l.Values = wireTDPList(l.Values[0], tdps, setTWire)
				got, err := parseISDBytes(t, w)
				if err != nil {
					return nil, err
				}
				return tTDPs(got.VlrCamelSubscriptionInfo.VtCSI.TBcsmCamelTDPDataList), nil
			},
		},
		{
			name: "GMSC T-CSI", field: "TBcsmCamelTDPDataList",
			keep:    []int64{int64(TBcsmTriggerTermAttemptAuthorized), int64(TBcsmTriggerTBusy), int64(TBcsmTriggerTNoAnswer)},
			ignored: 15,
			marshal: func(tdps []int64) error {
				g := gmscCamel()
				g.TCSI.TBcsmCamelTDPDataList = tTDPList(tdps)
				_, err := sriRespGmsc(g).Marshal()
				return err
			},
			parse: func(t *testing.T, tdps []int64) ([]int64, error) {
				w := sriRespWire(t, gmscCamel())
				l := w.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo.TCSI.TBcsmCamelTDPDataList
				l.Values = wireTDPList(l.Values[0], tdps, setTWire)
				got, err := ParseSriResp(strictBER(t, w))
				if err != nil {
					return nil, err
				}
				return tTDPs(got.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo.TCSI.TBcsmCamelTDPDataList), nil
			},
		},
		smsTDPListCase("VLR MO-SMS-CSI", collected, delivery,
			func(a *InsertSubscriberDataArg) *SMSCSI { return a.VlrCamelSubscriptionInfo.MoSmsCSI },
			func(w *gsm_map.InsertSubscriberDataArg) *gsm_map.SMSCSI { return w.VlrCamelSubscriptionInfo.MoSmsCSI }),
		smsTDPListCase("VLR MT-SMS-CSI", delivery, collected,
			func(a *InsertSubscriberDataArg) *SMSCSI { return a.VlrCamelSubscriptionInfo.MtSmsCSI },
			func(w *gsm_map.InsertSubscriberDataArg) *gsm_map.SMSCSI { return w.VlrCamelSubscriptionInfo.MtSmsCSI }),
		smsTDPListCase("SGSN MO-SMS-CSI", collected, delivery,
			func(a *InsertSubscriberDataArg) *SMSCSI { return a.SgsnCAMELSubscriptionInfo.MoSmsCSI },
			func(w *gsm_map.InsertSubscriberDataArg) *gsm_map.SMSCSI { return w.SgsnCAMELSubscriptionInfo.MoSmsCSI }),
		smsTDPListCase("SGSN MT-SMS-CSI", delivery, collected,
			func(a *InsertSubscriberDataArg) *SMSCSI { return a.SgsnCAMELSubscriptionInfo.MtSmsCSI },
			func(w *gsm_map.InsertSubscriberDataArg) *gsm_map.SMSCSI { return w.SgsnCAMELSubscriptionInfo.MtSmsCSI }),
		{
			name: "SGSN GPRS-CSI", field: "GPRSCamelTDPDataList",
			keep: []int64{
				int64(GPRSTDPAttach), int64(GPRSTDPAttachChangeOfPosition), int64(GPRSTDPPdpContextEstablishment),
				int64(GPRSTDPPdpContextEstablishmentAcknowledgement), int64(GPRSTDPPdpContextChangeOfPosition),
			},
			ignored: 13,
			marshal: func(tdps []int64) error {
				a := camelISD()
				a.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList = gprsTDPList(tdps)
				_, err := a.Marshal()
				return err
			},
			parse: func(t *testing.T, tdps []int64) ([]int64, error) {
				w := isdWire(t, camelISD())
				l := w.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList
				l.Values = wireTDPList(l.Values[0], tdps, setGPRSWire)
				got, err := parseISDBytes(t, w)
				if err != nil {
					return nil, err
				}
				var out []int64
				for _, d := range got.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList {
					out = append(out, int64(d.GprsTriggerDetectionPoint))
				}
				return out, nil
			},
		},
	}
}

// wantDuplicateTDP checks that err is ErrCamelDuplicateTriggerDetectionPoint
// for entry index of the list field.
func wantDuplicateTDP(t *testing.T, err error, field string, index int) {
	t.Helper()
	if !errors.Is(err, ErrCamelDuplicateTriggerDetectionPoint) {
		t.Fatalf("err = %v, want ErrCamelDuplicateTriggerDetectionPoint", err)
	}
	if at := fmt.Sprintf("%s[%d]", field, index); !strings.Contains(err.Error(), at) {
		t.Errorf("err = %v, want it to name %s", err, at)
	}
}

func TestMarshalRejectsDuplicateTriggerDetectionPoint(t *testing.T) {
	for _, c := range tdpListCases() {
		t.Run(c.name, func(t *testing.T) {
			// Every TDP the list admits once, then the first again.
			wantDuplicateTDP(t, c.marshal(slices.Concat(c.keep, c.keep[:1])), c.field, len(c.keep))
			// The same TDP twice in a row.
			for _, tdp := range c.keep {
				wantDuplicateTDP(t, c.marshal([]int64{tdp, tdp}), c.field, 1)
			}
		})
	}
}

// Each listed TDP once is accepted and round-trips. An SMS-CSI lists a
// single TDP, so its list holds one entry.
func TestMarshalDistinctTriggerDetectionPoints(t *testing.T) {
	for _, c := range tdpListCases() {
		t.Run(c.name, func(t *testing.T) {
			if err := c.marshal(c.keep); err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			got, err := c.parse(t, c.keep)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			wantEqual(t, c.field, c.keep, got)
		})
	}
}

func TestParseRejectsDuplicateTriggerDetectionPoint(t *testing.T) {
	for _, c := range tdpListCases() {
		t.Run(c.name, func(t *testing.T) {
			_, err := c.parse(t, slices.Concat(c.keep, c.keep[:1]))
			wantDuplicateTDP(t, err, c.field, len(c.keep))
			for _, tdp := range c.keep {
				_, err := c.parse(t, []int64{tdp, tdp})
				wantDuplicateTDP(t, err, c.field, 1)
			}
			// An ignored entry before the pair: the error names the wire
			// index of the repeated entry.
			_, err = c.parse(t, []int64{c.ignored, c.keep[0], c.keep[0]})
			wantDuplicateTDP(t, err, c.field, 2)
		})
	}
}

// Entries the receiver ignores do not count: two ignored entries with the
// same TDP, around a kept one, are not a duplicate, and an ignored entry
// with a TDP listed elsewhere (sms-DeliveryRequest in an MO-SMS-CSI) does
// not clash with the kept entry.
func TestParseIgnoredEntriesAreNotDuplicates(t *testing.T) {
	for _, c := range tdpListCases() {
		t.Run(c.name, func(t *testing.T) {
			for _, tc := range []struct{ wire, want []int64 }{
				{[]int64{c.ignored, c.keep[0], c.ignored}, c.keep[:1]},
				{[]int64{c.ignored, c.ignored, c.keep[0]}, c.keep[:1]},
				{append([]int64{c.ignored}, c.keep...), c.keep},
			} {
				got, err := c.parse(t, tc.wire)
				if err != nil {
					t.Fatalf("wire TDPs %v: Parse: %v", tc.wire, err)
				}
				wantEqual(t, fmt.Sprintf("wire TDPs %v", tc.wire), tc.want, got)
			}
		})
	}
}
