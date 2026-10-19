package gsmmap

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
)

// TS 29.002 V19.1.0 §17.7.1 CancelLocationArg restricts sending both
// MTRF flags but does not prescribe their rejection on receipt.
func TestReceiverCancelLocationMTRFFlags(t *testing.T) {
	for _, flags := range [][2]bool{{true, true}, {true, false}, {false, true}, {false, false}} {
		t.Run(fmt.Sprint(flags), func(t *testing.T) {
			w := &gsm_map.CancelLocationArg{Identity: gsm_map.NewIdentityImsi(semAddrIMSI), MtrfSupportedAndAuthorized: boolToNullPtr(flags[0]), MtrfSupportedAndNotAuthorized: boolToNullPtr(flags[1])}
			got, err := ParseCancelLocation(strictBER(t, w))
			if err != nil {
				t.Fatal(err)
			}
			if got.MtrfSupportedAndAuthorized != flags[0] || got.MtrfSupportedAndNotAuthorized != flags[1] {
				t.Fatalf("flags lost: %+v", got)
			}
			_, err = got.Marshal()
			if flags[0] && flags[1] {
				if !errors.Is(err, ErrCancelLocMtrfBothSet) {
					t.Fatalf("Marshal: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.1 TypeOfUpdate specifies sender applicability.
func TestReceiverCancelLocationTypeOfUpdate(t *testing.T) {
	for _, ct := range []*CancellationType{nil, new(CancellationTypeSubscriptionWithdraw), new(CancellationType(99)), new(CancellationTypeUpdateProcedure), new(CancellationTypeInitialAttachProcedure)} {
		t.Run(fmt.Sprint(ct), func(t *testing.T) {
			tou := TypeOfUpdateMmeChange
			w := &gsm_map.CancelLocationArg{Identity: gsm_map.NewIdentityImsi(semAddrIMSI), CancellationType: ct, TypeOfUpdate: &tou}
			got, err := ParseCancelLocation(strictBER(t, w))
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "CancellationType", ct, got.CancellationType)
			wantEqual(t, "TypeOfUpdate", &tou, got.TypeOfUpdate)
			_, err = got.Marshal()
			switch {
			case ct != nil && *ct == 99:
				if !errors.Is(err, ErrCancelLocInvalidCancellationType) {
					t.Fatalf("Marshal: %v", err)
				}
			case ct == nil || *ct == CancellationTypeSubscriptionWithdraw:
				if !errors.Is(err, ErrCancelLocTypeOfUpdateNotApplicable) {
					t.Fatalf("Marshal: %v", err)
				}
			default:
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func mobilityGprsWire(t *testing.T) *gsm_map.UpdateGprsLocationArg {
	t.Helper()
	w, err := convertUpdateGprsLocationToArg(&UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "12345", SGSNAddress: "192.0.2.2", SmsOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

// TS 29.002 V19.1.0 §17.7.1 PDN-GW-Update explicitly requires the HSS
// to ignore the enclosing EPS-Info when pdn-gw-Identity is missing.
func TestReceiverEPSInfoMissingPDNIdentity(t *testing.T) {
	for _, tc := range []struct {
		name   string
		update gsm_map.PDNGWUpdate
		keep   bool
	}{
		{"empty", gsm_map.PDNGWUpdate{}, false},
		{"apn only", gsm_map.PDNGWUpdate{Apn: new([]byte{1, 'a'})}, false},
		{"context only", gsm_map.PDNGWUpdate{ContextId: new(int64(1))}, false},
		{"identity without apn or context", gsm_map.PDNGWUpdate{PdnGwIdentity: &gsm_map.PDNGWIdentity{PdnGwIpv4Address: new(gsm_map.PDPAddress{192, 0, 2, 1})}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := mobilityGprsWire(t)
			e := gsm_map.NewEPSInfoPdnGwUpdate(tc.update)
			w.EpsInfo = &e
			got, err := ParseUpdateGprsLocation(strictBER(t, w))
			if err != nil {
				t.Fatal(err)
			}
			if (got.EpsInfo != nil) != tc.keep {
				t.Fatalf("EPS-Info = %+v, keep %v", got.EpsInfo, tc.keep)
			}
			if got.IMSI != "001010123456789" || got.SgsnNumber != "12345" || got.SGSNAddress != "192.0.2.2" || !got.SmsOnly {
				t.Fatalf("surrounding fields lost: %+v", got)
			}
			if tc.keep && (got.EpsInfo.PdnGwUpdate.PdnGwIdentity == nil || got.EpsInfo.PdnGwUpdate.APN != nil || got.EpsInfo.PdnGwUpdate.ContextID != nil) {
				t.Fatalf("identity-only update changed: %+v", got.EpsInfo)
			}
		})
	}
}

// 3GPP TS 29.002 V19.1.0 §17.7.1 discards unknown ISR bit assignments.
// ITU-T X.680 (02/2021) §22.7 permits trimming named-bit zero padding.
func TestReceiverISRUnknownBits(t *testing.T) {
	for bitLen := 3; bitLen <= 8; bitLen++ {
		for _, b := range []byte{0, 0xE0, 0xFF, 0x10, 0x08, 0x04, 0x02, 0x01} {
			mask := byte(0xFF << uint(8-bitLen))
			raw := b & mask
			t.Run(fmt.Sprintf("%d/%02x", bitLen, raw), func(t *testing.T) {
				w := mobilityGprsWire(t)
				e := gsm_map.NewEPSInfoIsrInformation(runtime.BitString{Bytes: []byte{raw}, BitLength: bitLen})
				w.EpsInfo = &e
				got, err := ParseUpdateGprsLocation(strictBER(t, w))
				if err != nil {
					t.Fatal(err)
				}
				wantEqual(t, "EpsInfo", &EpsInfo{IsrInformation: &IsrInformation{
					UpdateLocation: raw&0x80 != 0, CancelSGSN: raw&0x40 != 0, InitialAttachIndicator: raw&0x20 != 0,
				}}, got.EpsInfo)

			})
		}
	}
}

// TS 29.002 V19.1.0 §17.7.1 Ext-AccessRestrictionData permits 1..32
// bits without a receiver rule discarding its unnamed positions. The
// decoder removes trailing 0 bits of the named BIT STRING (ITU-T X.680
// (02/2021) §22.7), so Parse reports no BitLength and Marshal sends the
// value without them.
func TestReceiverExtAccessRestrictionUnknownBits(t *testing.T) {
	for _, tc := range []struct {
		raw     []byte
		bits    int
		unknown HexBytes
		sent    runtime.BitString
	}{
		{[]byte{0x20}, 3, HexBytes{0x20}, runtime.BitString{Bytes: []byte{0x20}, BitLength: 3}},
		{[]byte{0, 0, 0, 1}, 32, HexBytes{0, 0, 0, 1}, runtime.BitString{Bytes: []byte{0, 0, 0, 1}, BitLength: 32}},
		{[]byte{0xE0}, 3, HexBytes{0x20}, runtime.BitString{Bytes: []byte{0xE0}, BitLength: 3}},
		{[]byte{0xC0, 0, 0, 0}, 32, nil, runtime.BitString{Bytes: []byte{0xC0}, BitLength: 2}},
		{[]byte{0}, 1, nil, runtime.BitString{Bytes: []byte{0}, BitLength: 1}},
	} {
		t.Run(fmt.Sprintf("%x/%d", tc.raw, tc.bits), func(t *testing.T) {
			bs := runtime.BitString{Bytes: tc.raw, BitLength: tc.bits}
			got, err := ParseInsertSubscriberData(strictBER(t, &gsm_map.InsertSubscriberDataArg{ExtAccessRestrictionData: &bs}))
			if err != nil {
				t.Fatal(err)
			}
			r := got.ExtAccessRestrictionData
			if r == nil {
				t.Fatal("missing restrictions")
			}
			if r.NrAsSecondaryRATNotAllowed != (tc.raw[0]&0x80 != 0) || r.UnlicensedSpectrumAsSecondaryRATNotAllowed != (tc.raw[0]&0x40 != 0) || !bytes.Equal(r.UnknownBits, tc.unknown) {
				t.Fatalf("restrictions = %+v", r)
			}
			data, err := got.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			var w gsm_map.InsertSubscriberDataArg
			if err = w.UnmarshalBER(data); err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "wire restrictions", &tc.sent, w.ExtAccessRestrictionData)
		})
	}
}

// TS 29.002 V19.1.0 §17.7.1 SupportedRAT-Types names bits 0 to 5, nb-iot
// being bit 5: "bits 6 to 7 shall be ignored if received and not
// understood". The decoder removes trailing 0 bits (ITU-T X.680 (02/2021)
// §22.7), down to the SIZE (2..8) lower bound.
func TestReceiverSupportedRATTypesNBIoT(t *testing.T) {
	for _, tc := range []struct {
		raw  byte
		bits int
		want SupportedRATTypes
		sent runtime.BitString
	}{
		{0x04, 6, SupportedRATTypes{NBIoT: true}, runtime.BitString{Bytes: []byte{0x04}, BitLength: 6}},
		{0x0C, 6, SupportedRATTypes{EUTRAN: true, NBIoT: true}, runtime.BitString{Bytes: []byte{0x0C}, BitLength: 6}},
		{0x07, 8, SupportedRATTypes{NBIoT: true}, runtime.BitString{Bytes: []byte{0x04}, BitLength: 6}},
		{0x83, 8, SupportedRATTypes{UTRAN: true}, runtime.BitString{Bytes: []byte{0x80}, BitLength: 2}},
	} {
		t.Run(fmt.Sprintf("%02x/%d", tc.raw, tc.bits), func(t *testing.T) {
			w := mobilityGprsWire(t)
			w.SgsnCapability = &gsm_map.SGSNCapability{SupportedRATTypesIndicator: &runtime.BitString{Bytes: []byte{tc.raw}, BitLength: tc.bits}}
			got, err := ParseUpdateGprsLocation(strictBER(t, w))
			if err != nil {
				t.Fatal(err)
			}
			if got.SGSNCapability == nil {
				t.Fatal("missing SGSNCapability")
			}
			wantEqual(t, "SupportedRATTypesIndicator", &tc.want, got.SGSNCapability.SupportedRATTypesIndicator)
			data, err := got.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			var sent gsm_map.UpdateGprsLocationArg
			if err := sent.UnmarshalBER(data); err != nil {
				t.Fatal(err)
			}
			if sent.SgsnCapability == nil {
				t.Fatal("missing sgsn-Capability")
			}
			wantEqual(t, "sent supportedRAT-TypesIndicator", &tc.sent, sent.SgsnCapability.SupportedRATTypesIndicator)
		})
	}
}

// TS 29.002 V19.1.0 §17.7.1 PDP-Context QoS predecessor and extended
// address requirements restrict sending; independent received fields remain inspectable.
func TestReceiverPDPContextSenderDependencies(t *testing.T) {
	for _, tc := range []struct {
		name string
		mut  func(*gsm_map.PDPContext)
		want error
	}{
		{"Ext2 without Ext", func(w *gsm_map.PDPContext) { w.ExtQoSSubscribed = nil }, ErrExtQoSHierarchyViolated},
		{"Ext3 without Ext2", func(w *gsm_map.PDPContext) { w.Ext2QoSSubscribed = nil }, ErrExtQoSHierarchyViolated},
		{"Ext4 without Ext3", func(w *gsm_map.PDPContext) { w.Ext3QoSSubscribed = nil }, ErrExtQoSHierarchyViolated},
		{"extended IPv4 without base", func(w *gsm_map.PDPContext) {
			w.PdpAddress = nil
			v := gsm_map.PDPAddress{192, 0, 2, 1}
			w.ExtPdpAddress = &v
		}, ErrExtPDPAddressWithoutPDPAddress},
		{"extended IPv6 without base", func(w *gsm_map.PDPContext) {
			w.PdpAddress = nil
			v := make([]byte, 16)
			w.ExtPdpAddress = &v
		}, ErrExtPDPAddressWithoutPDPAddress},
		{"complete", func(*gsm_map.PDPContext) {}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := makePDPContext()
			w, err := convertPDPContextToWire(&p)
			if err != nil {
				t.Fatal(err)
			}
			tc.mut(w)
			arg := &gsm_map.InsertSubscriberDataArg{GprsSubscriptionData: &gsm_map.GPRSSubscriptionData{GprsDataList: &gsm_map.GPRSDataList{Values: []gsm_map.PDPContext{*w}}}}
			got, err := ParseInsertSubscriberData(strictBER(t, arg))
			if err != nil {
				t.Fatal(err)
			}
			r := got.GprsSubscriptionData.GprsDataList[0]
			want := func(v *[]byte) HexBytes {
				if v == nil {
					return nil
				}
				return HexBytes(*v)
			}
			wantEqual(t, "ExtQoS", want(w.ExtQoSSubscribed), r.ExtQoSSubscribed)
			wantEqual(t, "Ext2QoS", want(w.Ext2QoSSubscribed), r.Ext2QoSSubscribed)
			wantEqual(t, "Ext3QoS", want(w.Ext3QoSSubscribed), r.Ext3QoSSubscribed)
			wantEqual(t, "Ext4QoS", want(w.Ext4QoSSubscribed), r.Ext4QoSSubscribed)
			if w.PdpAddress == nil && r.PdpAddress != nil {
				t.Fatal("base address invented")
			}
			if w.ExtPdpAddress != nil {
				wantEqual(t, "extended address", HexBytes(*w.ExtPdpAddress), r.ExtPdpAddress)
			}
			_, err = got.Marshal()
			if !errors.Is(err, tc.want) {
				t.Fatalf("Marshal = %v, want %v", err, tc.want)
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.1 DestinationNumberCriteria restricts list
// absence in sent components without prescribing receiver rejection.
func TestReceiverDestinationNumberCriteriaListAbsence(t *testing.T) {
	for _, where := range []string{"ISD", "SRI"} {
		for _, lists := range []int{0, 1, 2, 3} {
			t.Run(fmt.Sprintf("%s/%d", where, lists), func(t *testing.T) {
				d := &gsm_map.DestinationNumberCriteria{MatchType: MatchTypeInhibiting}
				if lists&1 != 0 {
					d.DestinationNumberList = &gsm_map.DestinationNumberList{Values: []gsm_map.ISDNAddressString{{0x91, 0x21}}}
				}
				if lists&2 != 0 {
					d.DestinationNumberLengthList = &gsm_map.DestinationNumberLengthList{Values: []int64{2}}
				}
				criteria := &gsm_map.OBcsmCamelTDPCriteriaList{Values: []gsm_map.OBcsmCamelTDPCriteria{{OBcsmTriggerDetectionPoint: OBcsmTriggerCollectedInfo, DestinationNumberCriteria: d}}}
				var r *DestinationNumberCriteria
				var senderErr error
				if where == "ISD" {
					w := &gsm_map.InsertSubscriberDataArg{VlrCamelSubscriptionInfo: &gsm_map.VlrCamelSubscriptionInfo{OBcsmCamelTDPCriteriaList: criteria}}
					got, err := ParseInsertSubscriberData(strictBER(t, w))
					if err != nil {
						t.Fatal(err)
					}
					r = got.VlrCamelSubscriptionInfo.OBcsmCamelTDPCriteriaList[0].DestinationNumberCriteria
					_, senderErr = got.Marshal()
				} else {
					w := &gsm_map.SendRoutingInfoRes{ExtendedRoutingInfo: new(gsm_map.NewExtendedRoutingInfoCamelRoutingInfo(gsm_map.CamelRoutingInfo{GmscCamelSubscriptionInfo: gsm_map.GmscCamelSubscriptionInfo{OBcsmCamelTDPCriteriaList: criteria}}))}
					got, err := ParseSriResp(strictBER(t, w))
					if err != nil {
						t.Fatal(err)
					}
					r = got.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo.OBcsmCamelTDPCriteriaList[0].DestinationNumberCriteria
					_, senderErr = got.Marshal()
				}
				if r == nil || r.MatchType != MatchTypeInhibiting || (r.DestinationNumberList != nil) != (lists&1 != 0) || (r.DestinationNumberLengthList != nil) != (lists&2 != 0) {
					t.Fatalf("criteria changed: %+v", r)
				}
				if lists == 0 {
					if !errors.Is(senderErr, ErrCamelMissingDestinationNumberCriteria) {
						t.Fatalf("Marshal: %v", senderErr)
					}
				} else if senderErr != nil {
					t.Fatal(senderErr)
				}
			})
		}
	}
}

// Repeated triggers identify separate received descriptors; each descriptor's
// service key and position remain available after Parse (TS 29.002 §17.7.1).
func TestReceiverRepeatedTDPDescriptors(t *testing.T) {
	for _, family := range []string{"O", "T", "SMS", "GPRS"} {
		t.Run(family, func(t *testing.T) {
			w := isdWire(t, camelISD())
			switch family {
			case "O":
				l := w.VlrCamelSubscriptionInfo.OCSI.OBcsmCamelTDPDataList
				l.Values = wireTDPList(l.Values[0], []int64{2, 2}, setOWire)
			case "T":
				l := w.VlrCamelSubscriptionInfo.VtCSI.TBcsmCamelTDPDataList
				l.Values = wireTDPList(l.Values[0], []int64{12, 12}, setTWire)
			case "SMS":
				l := w.VlrCamelSubscriptionInfo.MoSmsCSI.SmsCAMELTDPDataList
				l.Values = wireTDPList(l.Values[0], []int64{1, 1}, setSMSWire)
			case "GPRS":
				l := w.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList
				l.Values = wireTDPList(l.Values[0], []int64{1, 1}, setGPRSWire)
			}
			got, err := ParseInsertSubscriberData(strictBER(t, w))
			if err != nil {
				t.Fatal(err)
			}
			var keys []int64
			switch family {
			case "O":
				for _, e := range got.VlrCamelSubscriptionInfo.OCSI.OBcsmCamelTDPDataList {
					keys = append(keys, e.ServiceKey)
				}
			case "T":
				for _, e := range got.VlrCamelSubscriptionInfo.VtCSI.TBcsmCamelTDPDataList {
					keys = append(keys, e.ServiceKey)
				}
			case "SMS":
				for _, e := range got.VlrCamelSubscriptionInfo.MoSmsCSI.SmsCAMELTDPDataList {
					keys = append(keys, e.ServiceKey)
				}
			case "GPRS":
				for _, e := range got.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList {
					keys = append(keys, e.ServiceKey)
				}
			}
			wantEqual(t, "descriptor service keys in received order", []int64{1, 2}, keys)
			if _, err := got.Marshal(); !errors.Is(err, ErrCamelDuplicateTriggerDetectionPoint) {
				t.Fatalf("Marshal: %v", err)
			}
		})
	}
}

// A present empty list violates ASN SIZE, unlike an absent list; these
// constraints still belong to the codec (TS 29.002 V19.1.0 §17.7.1).
func TestReceiverDestinationNumberCriteriaEmptyPresentLists(t *testing.T) {
	for _, where := range []string{"ISD", "SRI"} {
		for _, numbers := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/numbers=%t", where, numbers), func(t *testing.T) {
				d := &gsm_map.DestinationNumberCriteria{MatchType: MatchTypeInhibiting}
				if numbers {
					d.DestinationNumberList = &gsm_map.DestinationNumberList{}
				} else {
					d.DestinationNumberLengthList = &gsm_map.DestinationNumberLengthList{}
				}
				l := &gsm_map.OBcsmCamelTDPCriteriaList{Values: []gsm_map.OBcsmCamelTDPCriteria{{OBcsmTriggerDetectionPoint: OBcsmTriggerCollectedInfo, DestinationNumberCriteria: d}}}
				var err error
				if where == "ISD" {
					_, err = ParseInsertSubscriberData(tolerantBER(t, &gsm_map.InsertSubscriberDataArg{VlrCamelSubscriptionInfo: &gsm_map.VlrCamelSubscriptionInfo{OBcsmCamelTDPCriteriaList: l}}))
				} else {
					_, err = ParseSriResp(tolerantBER(t, &gsm_map.SendRoutingInfoRes{ExtendedRoutingInfo: new(gsm_map.NewExtendedRoutingInfoCamelRoutingInfo(gsm_map.CamelRoutingInfo{GmscCamelSubscriptionInfo: gsm_map.GmscCamelSubscriptionInfo{OBcsmCamelTDPCriteriaList: l}}))}))
				}
				var constraint *ber.ValidationError
				if !errors.As(err, &constraint) {
					t.Fatalf("Parse: %v, want codec SIZE error", err)
				}
			})
		}
	}
}
