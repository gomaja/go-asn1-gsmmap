package gsmmap

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/runtime/ber"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func senderMobilityGPRS() *UpdateGprsLocation {
	return &UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "12345", SGSNAddress: "192.0.2.2"}
}

func senderMobilityRoundTrip(t *testing.T, in marshaler, parse func([]byte) (marshaler, error)) {
	t.Helper()
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(in, got) {
		t.Fatalf("round trip: got %#v, want %#v", got, in)
	}
}

// TS 29.002 V19.1.0 §17.7.1 SupportedLCS-CapabilitySets forbids SGSN set1.
func TestSenderSGSNCapabilityLCS(t *testing.T) {
	for _, caps := range []*SupportedLCSCapabilitySets{nil, {}, {LcsCapabilitySet2: true}, {LcsCapabilitySet3: true}, {LcsCapabilitySet4: true}, {LcsCapabilitySet5: true}, {LcsCapabilitySet1: true}} {
		in := senderMobilityGPRS()
		in.SGSNCapability = &SGSNCapability{SupportedLCSCapabilitySets: caps}
		if caps != nil && caps.LcsCapabilitySet1 {
			if _, err := in.Marshal(); !errors.Is(err, ErrSGSNCapabilityLCSSet1NotAllowed) {
				t.Fatalf("set1: %v", err)
			}
			continue
		}
		senderMobilityRoundTrip(t, in, asParser(ParseUpdateGprsLocation))
	}
	in := &UpdateLocation{IMSI: "001010123456789", MscNumber: "12345", VlrNumber: "12346", VlrCapability: &VlrCapability{SupportedLCSCapabilitySets: &SupportedLCSCapabilitySets{LcsCapabilitySet1: true}}}
	senderMobilityRoundTrip(t, in, asParser(ParseUpdateLocation))
}

// TS 29.002 V19.1.0 §17.7.1 PDN-GW-Update requires identity and APN or ContextID.
func TestSenderPdnGwUpdateRequiredFields(t *testing.T) {
	id := &PdnGwIdentity{IPv4Address: HexBytes{192, 0, 2, 1}}
	for _, tc := range []struct {
		p    PdnGwUpdate
		want error
	}{
		{PdnGwUpdate{}, ErrPdnGwUpdateMissingIdentity}, {PdnGwUpdate{APN: HexBytes{1, 'a'}}, ErrPdnGwUpdateMissingIdentity},
		{PdnGwUpdate{PdnGwIdentity: id}, ErrPdnGwUpdateMissingAPNOrContextID},
		{PdnGwUpdate{PdnGwIdentity: id, APN: HexBytes{1, 'a'}}, nil}, {PdnGwUpdate{PdnGwIdentity: id, ContextID: new(1)}, nil},
		{PdnGwUpdate{PdnGwIdentity: id, ContextID: new(1), APN: HexBytes{1, 'a'}}, nil},
	} {
		in := senderMobilityGPRS()
		in.EpsInfo = &EpsInfo{PdnGwUpdate: &tc.p}
		if tc.want != nil {
			if _, err := in.Marshal(); !errors.Is(err, tc.want) {
				t.Fatalf("%+v: %v, want %v", tc.p, err, tc.want)
			}
			continue
		}
		senderMobilityRoundTrip(t, in, asParser(ParseUpdateGprsLocation))
	}
}

// TS 29.002 V19.1.0 §17.7.1 ISR-Information assigns only bits0..2.
func TestSenderISRInformationBits(t *testing.T) {
	for bits := 3; bits <= 8; bits++ {
		for _, b := range []byte{0, 0x80, 0x40, 0x20, 0xe0, 0x10, 0x08, 0x04, 0x02, 0x01} {
			in := senderMobilityGPRS()
			in.EpsInfo = &EpsInfo{IsrInformation: HexBytes{b}, IsrInformationBits: bits}
			effective := b & byte(0xff<<uint(8-bits))
			if effective&0x1f != 0 {
				if _, err := in.Marshal(); !errors.Is(err, ErrISRInformationUnknownBits) {
					t.Fatalf("%02x/%d: %v", b, bits, err)
				}
				continue
			}
			data, err := in.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			got, err := ParseUpdateGprsLocation(data)
			if err != nil {
				t.Fatal(err)
			}
			if got.EpsInfo == nil || got.EpsInfo.IsrInformationBits != bits || !reflect.DeepEqual(got.EpsInfo.IsrInformation, HexBytes{effective}) {
				t.Fatalf("%02x/%d: %+v", b, bits, got.EpsInfo)
			}
		}
	}
}

func senderMobilityPDN(v byte) *InsertSubscriberDataArg {
	return &InsertSubscriberDataArg{EpsSubscriptionData: &EPSSubscriptionData{ApnConfigurationProfile: &APNConfigurationProfile{DefaultContext: 1, EpsDataList: EPSDataList{{ContextId: 1, PdnType: HexBytes{v}, Apn: HexBytes{1, 'a'}, EpsQosSubscribed: EPSQoSSubscribed{QosClassIdentifier: 1}}}}}}
}

// TS 29.002 V19.1.0 §17.7.1 PDN-Type assigns only octets01..04.
func TestSenderPDNType(t *testing.T) {
	for v := 0; v < 256; v++ {
		in := senderMobilityPDN(byte(v))
		if v < 1 || v > 4 {
			if _, err := in.Marshal(); !errors.Is(err, ErrPDNTypeInvalid) {
				t.Fatalf("%02x: %v", v, err)
			}
			continue
		}
		senderMobilityRoundTrip(t, in, asParser(ParseInsertSubscriberData))
	}
	for _, v := range []byte{0, 5, 0xff} {
		w := semMust(convertInsertSubscriberDataArgToWire(senderMobilityPDN(1)))
		w.EpsSubscriptionData.ApnConfigurationProfile.EpsDataList.Values[0].PdnType = []byte{v}
		got, err := ParseInsertSubscriberData(strictBER(t, w))
		if err != nil {
			t.Fatal(err)
		}
		if got.EpsSubscriptionData.ApnConfigurationProfile.EpsDataList[0].PdnType[0] != v {
			t.Fatal("received octet changed")
		}
	}
}

func senderMobilityPDP() PDPContext {
	return PDPContext{PdpContextId: 1, PdpType: HexBytes{0xf1, 0x21}, QosSubscribed: HexBytes{0, 0, 0}, Apn: HexBytes{1, 'a'}}
}
func senderMobilityISD(p PDPContext) *InsertSubscriberDataArg {
	return &InsertSubscriberDataArg{GprsSubscriptionData: &GPRSSubscriptionData{GprsDataList: GPRSDataList{p}}}
}

func senderMobilityCarrier(api string, si SubscriberInfo) (marshaler, func([]byte) (marshaler, error)) {
	switch api {
	case "ATI":
		return &AnyTimeInterrogationRes{SubscriberInfo: si}, asParser(ParseAnyTimeInterrogationRes)
	case "PSI":
		return &ProvideSubscriberInfoRes{SubscriberInfo: si}, asParser(ParseProvideSubscriberInfoRes)
	default:
		return &SriResp{SubscriberInfo: &si}, asParser(ParseSriResp)
	}
}
func senderMobilityReceivedSubscriber(t *testing.T, api string, si gsm_map.SubscriberInfo, in marshaler, want error) {
	t.Helper()
	var wire semWire
	switch api {
	case "ATI":
		wire = &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: si}
	case "PSI":
		wire = &gsm_map.ProvideSubscriberInfoRes{SubscriberInfo: si}
	default:
		wire = &gsm_map.SendRoutingInfoRes{SubscriberInfo: &si}
	}
	_, parse := senderMobilityCarrier(api, SubscriberInfo{})
	got, err := parse(strictBER(t, wire))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(in, got) {
		t.Fatalf("received value changed: got %#v, want %#v", got, in)
	}
	if _, err := got.Marshal(); !errors.Is(err, want) {
		t.Fatalf("Marshal received: %v, want %v", err, want)
	}
}

func senderMobilityPDPInfo(t *testing.T, c gsm_map.PDPContextInfo, eps, reachable bool) SubscriberInfo {
	t.Helper()
	state := &PsSubscriberState{}
	if reachable {
		state.PsPDPActiveReachableForPaging = []HexBytes{strictBER(t, &c)}
	} else {
		state.PsPDPActiveNotReachableForPaging = []HexBytes{strictBER(t, &c)}
	}
	if eps {
		return SubscriberInfo{EpsSubscriberState: state}
	}
	return SubscriberInfo{PsSubscriberState: state}
}
func senderMobilityCheckPDPInfo(t *testing.T, c gsm_map.PDPContextInfo, want error) {
	t.Helper()
	for _, api := range []string{"ATI", "PSI", "SRI"} {
		for _, eps := range []bool{false, true} {
			for _, reachable := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/eps=%v/reachable=%v", api, eps, reachable), func(t *testing.T) {
					in, parse := senderMobilityCarrier(api, senderMobilityPDPInfo(t, c, eps, reachable))
					if want != nil {
						if _, err := in.Marshal(); !errors.Is(err, want) {
							t.Fatalf("Marshal: %v, want %v", err, want)
						}
						list := &gsm_map.PDPContextInfoList{Values: []gsm_map.PDPContextInfo{c}}
						state := gsm_map.NewPSSubscriberStatePsPDPActiveNotReachableForPaging(list)
						if reachable {
							state = gsm_map.NewPSSubscriberStatePsPDPActiveReachableForPaging(list)
						}
						wireInfo := gsm_map.SubscriberInfo{PsSubscriberState: &state}
						if eps {
							wireInfo = gsm_map.SubscriberInfo{EpsSubscriberState: &state}
						}
						senderMobilityReceivedSubscriber(t, api, wireInfo, in, want)
					} else {
						senderMobilityRoundTrip(t, in, parse)
					}
				})
			}
		}
	}
}

// TS 29.002 V19.1.0 §17.7.1 PDP-Type/Ext-PDP-Type; TS 29.060 V19.0.0 §7.7.27 Figures35..39.
func TestSenderPDPTypes(t *testing.T) {
	for _, v := range []HexBytes{{0xf0, 1}, {0xf0, 2}, {0xf1, 0x21}, {0xf1, 0x57}, {0, 0}, {1, 0x21}, {0xf2, 0x21}, {0xf1, 0}, {0xf1, 0x8d}} {
		t.Run(fmt.Sprintf("base/%x", v), func(t *testing.T) {
			p := senderMobilityPDP()
			p.PdpType = v
			c := gsm_map.PDPContextInfo{PdpContextIdentifier: 1, PdpType: v}
			var want error
			if (v[0] != 0xf0 || v[1] != 1 && v[1] != 2) && (v[0] != 0xf1 || v[1] != 0x21 && v[1] != 0x57) {
				want = ErrPDPTypeInvalid
			}
			if want != nil {
				if _, err := senderMobilityISD(p).Marshal(); !errors.Is(err, want) {
					t.Fatal(err)
				}
			} else {
				senderMobilityRoundTrip(t, senderMobilityISD(p), asParser(ParseInsertSubscriberData))
			}
			senderMobilityCheckPDPInfo(t, c, want)
		})
	}
	for _, v := range []HexBytes{nil, {0xf1, 0x8d}, {0, 0}, {0xf0, 1}, {0xf1, 0x21}, {0xf1, 0x57}, {1, 0x8d}} {
		t.Run(fmt.Sprintf("extension/%x", v), func(t *testing.T) {
			p := senderMobilityPDP()
			p.ExtPdpType = v
			c := gsm_map.PDPContextInfo{PdpContextIdentifier: 1, PdpType: p.PdpType}
			if v != nil {
				ext := []byte(v)
				c.ExtPdpType = &ext
			}
			var want error
			if v != nil && (v[0] != 0xf1 || v[1] != 0x8d) {
				want = ErrExtPDPTypeInvalid
			}
			if want != nil {
				if _, err := senderMobilityISD(p).Marshal(); !errors.Is(err, want) {
					t.Fatal(err)
				}
			} else {
				senderMobilityRoundTrip(t, senderMobilityISD(p), asParser(ParseInsertSubscriberData))
			}
			senderMobilityCheckPDPInfo(t, c, want)
		})
	}
}

// TS 29.002 V19.1.0 §17.7.1 PDP-Context/PDP-ContextInfo paired addresses use different IP families.
func TestSenderPDPAddressFamilies(t *testing.T) {
	for _, pair := range [][2]int{{4, 4}, {16, 16}, {1, 4}, {4, 1}, {4, 16}, {16, 4}} {
		t.Run(fmt.Sprint(pair), func(t *testing.T) {
			p := senderMobilityPDP()
			p.PdpAddress = make(HexBytes, pair[0])
			p.ExtPdpAddress = make(HexBytes, pair[1])
			var want error
			if pair != [2]int{4, 16} && pair != [2]int{16, 4} {
				want = ErrPDPAddressFamilyCombinationInvalid
			}
			if want != nil {
				if _, err := senderMobilityISD(p).Marshal(); !errors.Is(err, want) {
					t.Fatal(err)
				}
			} else {
				senderMobilityRoundTrip(t, senderMobilityISD(p), asParser(ParseInsertSubscriberData))
			}
			base, ext := []byte(p.PdpAddress), []byte(p.ExtPdpAddress)
			senderMobilityCheckPDPInfo(t, gsm_map.PDPContextInfo{PdpContextIdentifier: 1, PdpType: p.PdpType, PdpAddress: &base, ExtPdpAddress: &ext}, want)
		})
	}
	senderMobilityCheckPDPInfo(t, gsm_map.PDPContextInfo{PdpContextIdentifier: 1, PdpType: []byte{0xf1, 0x21}, ExtPdpAddress: new([]byte{192, 0, 2, 1})}, ErrPDPContextInfoExtPDPAddressWithoutPDPAddress)
	senderMobilityCheckPDPInfo(t, gsm_map.PDPContextInfo{PdpContextIdentifier: 1, PdpType: []byte{0xf1, 0x21}, PdpAddress: new([]byte{192, 0, 2, 1})}, nil)
}

// TS 29.002 V19.1.0 §17.7.1 PDP-ContextInfo requires each QoS2/3/4 predecessor.
func TestSenderPDPContextInfoQoSHierarchy(t *testing.T) {
	for _, family := range []string{"Subscribed", "Requested", "Negotiated"} {
		for level := 2; level <= 4; level++ {
			t.Run(fmt.Sprintf("%s/%d", family, level), func(t *testing.T) {
				c := gsm_map.PDPContextInfo{PdpContextIdentifier: 1, PdpType: []byte{0xf1, 0x21}}
				v := reflect.ValueOf(&c).Elem()
				set := func(n int) {
					name := "Qos"
					size := 3
					if n > 1 {
						name += fmt.Sprint(n)
						size = 1
					}
					name += family
					if n == 1 {
						size = 9
					}
					b := make([]byte, size)
					v.FieldByName(name).Set(reflect.ValueOf(&b))
				}
				set(level)
				senderMobilityCheckPDPInfo(t, c, ErrPDPContextInfoQoSWithoutPredecessor)
				for n := level - 1; n >= 1; n-- {
					set(n)
				}
				senderMobilityCheckPDPInfo(t, c, nil)
			})
		}
	}
}

// TS 29.002 V19.1.0 §17.7.1 IMS-VoiceOverPS-SessionsInd forbids unknown only in PSI results.
func TestSenderPSIIMSVoiceUnknown(t *testing.T) {
	for _, api := range []string{"ATI", "PSI", "SRI"} {
		for _, v := range []*ImsVoiceOverPSSessionsIndication{nil, new(IMSVoiceOverPSNotSupported), new(IMSVoiceOverPSSupported), new(IMSVoiceOverPSUnknown)} {
			in, parse := senderMobilityCarrier(api, SubscriberInfo{ImsVoiceOverPSSessionsIndication: v})
			if api == "PSI" && v != nil && *v == IMSVoiceOverPSUnknown {
				if _, err := in.Marshal(); !errors.Is(err, ErrPSIResImsVoiceOverPSSessionsIndicationUnknown) {
					t.Fatal(err)
				}
				senderMobilityReceivedSubscriber(t, api, gsm_map.SubscriberInfo{ImsVoiceOverPSSessionsIndication: v}, in, ErrPSIResImsVoiceOverPSSessionsIndicationUnknown)
			} else {
				senderMobilityRoundTrip(t, in, parse)
			}
		}
	}
}

// TS 29.002 V19.1.0 §17.7.1 NumberPortabilityStatus restricts values4/5 to ATI.
func TestSenderNumberPortabilityATIOnly(t *testing.T) {
	for _, api := range []string{"ATI", "PSI", "SRI"} {
		for _, v := range []*NumberPortabilityStatus{nil, new(MnpNotKnownToBePorted), new(MnpOwnNumberPortedOut), new(MnpForeignNumberPortedToForeignNetwork), new(MnpOwnNumberNotPortedOut), new(MnpForeignNumberPortedIn)} {
			in, parse := senderMobilityCarrier(api, SubscriberInfo{MnpInfoRes: &MnpInfoRes{NumberPortabilityStatus: v}})
			if api != "ATI" && v != nil && (*v == 4 || *v == 5) {
				if _, err := in.Marshal(); !errors.Is(err, ErrNumberPortabilityStatusATIOnly) {
					t.Fatal(err)
				}
				senderMobilityReceivedSubscriber(t, api, gsm_map.SubscriberInfo{MnpInfoRes: &gsm_map.MNPInfoRes{NumberPortabilityStatus: v}}, in, ErrNumberPortabilityStatusATIOnly)
			} else {
				senderMobilityRoundTrip(t, in, parse)
			}
			if api == "SRI" {
				top := &SriResp{NumberPortabilityStatus: v}
				if v != nil && (*v == 4 || *v == 5) {
					if _, err := top.Marshal(); !errors.Is(err, ErrNumberPortabilityStatusATIOnly) {
						t.Fatal(err)
					}
					got, err := ParseSriResp(strictBER(t, &gsm_map.SendRoutingInfoRes{NumberPortabilityStatus: v}))
					if err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(top, got) {
						t.Fatal("received top-level status changed")
					}
				} else {
					senderMobilityRoundTrip(t, top, asParser(ParseSriResp))
				}
			}
		}
	}
}

// TS 29.002 V19.1.0 §17.7.1 RequestedInfo states all five local presence conditions.
func TestSenderRequestedInfoDependencies(t *testing.T) {
	cases := []struct {
		ri   RequestedInfo
		want error
	}{
		{RequestedInfo{CurrentLocation: true}, ErrRequestedInfoCurrentLocationWithoutLocationInformation},
		{RequestedInfo{LocationInformationEPSSupported: true}, ErrRequestedInfoEPSSupportedWithoutLocationInformation},
		{RequestedInfo{RequestedNodes: &RequestedNodes{MME: true}, RequestedDomain: new(CsDomain)}, ErrRequestedInfoRequestedNodesInCSDomain},
		{RequestedInfo{ServingNodeIndication: true}, ErrRequestedInfoServingNodeWithoutLocationInformation},
		{RequestedInfo{LocationInformation: true, CurrentLocation: true, ServingNodeIndication: true}, ErrRequestedInfoServingNodeWithCurrentLocation},
		{RequestedInfo{}, nil}, {RequestedInfo{LocationInformation: true, CurrentLocation: true, LocationInformationEPSSupported: true}, nil},
		{RequestedInfo{RequestedNodes: &RequestedNodes{MME: true}}, nil}, {RequestedInfo{RequestedNodes: &RequestedNodes{MME: true}, RequestedDomain: new(PsDomain)}, nil},
		{RequestedInfo{LocationInformation: true, ServingNodeIndication: true}, nil},
	}
	for _, api := range []string{"ATI", "PSI"} {
		for i, tc := range cases {
			t.Run(fmt.Sprintf("%s/%d", api, i), func(t *testing.T) {
				var in marshaler
				var parse func([]byte) (marshaler, error)
				if api == "ATI" {
					in = &AnyTimeInterrogation{SubscriberIdentity: SubscriberIdentity{IMSI: "001010123456789"}, GsmSCFAddress: "12345", RequestedInfo: tc.ri}
					parse = asParser(ParseAnyTimeInterrogation)
				} else {
					in = &ProvideSubscriberInfo{IMSI: "001010123456789", RequestedInfo: tc.ri}
					parse = asParser(ParseProvideSubscriberInfo)
				}
				if tc.want != nil {
					if _, err := in.Marshal(); !errors.Is(err, tc.want) {
						t.Fatalf("%v, want %v", err, tc.want)
					}
				} else {
					senderMobilityRoundTrip(t, in, parse)
				}
			})
		}
	}
}

// TS 29.002 V19.1.0 §17.7.1 GeodeticInformation; Q.763 (1999) §3.88(c), Figure82.1.
func TestSenderGeodeticInformationShape(t *testing.T) {
	for _, shape := range []byte{0, 1, 2, 3, 4, 5, 0x7f, 0x81, 0xff} {
		for _, family := range []string{"CS", "EPS", "GPRS", "5GS"} {
			for _, api := range []string{"ATI", "PSI", "SRI"} {
				t.Run(fmt.Sprintf("%s/%s/%02x", api, family, shape), func(t *testing.T) {
					b := make(HexBytes, 10)
					b[1] = shape
					si := SubscriberInfo{}
					switch family {
					case "CS":
						si.LocationInformation = &CSLocationInformation{GeodeticInformation: b}
					case "EPS":
						si.LocationInformationEPS = &EPSLocationInformation{GeodeticInformation: b}
					case "GPRS":
						si.LocationInformationGPRS = &GPRSLocationInformation{GeodeticInformation: b}
					case "5GS":
						si.LocationInformation5GS = &LocationInformation5GS{GeodeticInformation: b}
					}
					in, parse := senderMobilityCarrier(api, si)
					if shape&0x7f != 1 {
						if _, err := in.Marshal(); !errors.Is(err, ErrGeodeticInformationShapeInvalid) {
							t.Fatal(err)
						}
						raw := []byte(b)
						w := gsm_map.SubscriberInfo{}
						switch family {
						case "CS":
							w.LocationInformation = &gsm_map.LocationInformation{GeodeticInformation: &raw}
						case "EPS":
							w.LocationInformationEPS = &gsm_map.LocationInformationEPS{GeodeticInformation: &raw}
						case "GPRS":
							w.LocationInformationGPRS = &gsm_map.LocationInformationGPRS{GeodeticInformation: &raw}
						case "5GS":
							w.LocationInformation5GS = &gsm_map.LocationInformation5GS{GeodeticInformation: &raw}
						}
						senderMobilityReceivedSubscriber(t, api, w, in, ErrGeodeticInformationShapeInvalid)
					} else {
						senderMobilityRoundTrip(t, in, parse)
					}
				})
			}
		}
	}
}

// TS 29.002 V19.1.0 §17.7.8 TBCD-STRING permits filler only after the final odd digit.
func TestSenderRouteingNumberTBCD(t *testing.T) {
	for _, v := range []HexBytes{nil, {0x21}, {0xf1}, {0xba, 0xdc, 0xfe}, {0x1f}, {0xff}, {0x21, 0xff}, {0xf1, 0x21}} {
		for _, api := range []string{"ATI", "PSI", "SRI"} {
			t.Run(fmt.Sprintf("%s/%x", api, v), func(t *testing.T) {
				in, parse := senderMobilityCarrier(api, SubscriberInfo{MnpInfoRes: &MnpInfoRes{RouteingNumber: v}})
				bad := reflect.DeepEqual(v, HexBytes{0x1f}) || reflect.DeepEqual(v, HexBytes{0xff}) || reflect.DeepEqual(v, HexBytes{0x21, 0xff}) || reflect.DeepEqual(v, HexBytes{0xf1, 0x21})
				if bad {
					if _, err := in.Marshal(); !errors.Is(err, ErrRouteingNumberTBCDInvalid) {
						t.Fatal(err)
					}
					raw := []byte(v)
					senderMobilityReceivedSubscriber(t, api, gsm_map.SubscriberInfo{MnpInfoRes: &gsm_map.MNPInfoRes{RouteingNumber: &raw}}, in, ErrRouteingNumberTBCDInvalid)
				} else {
					senderMobilityRoundTrip(t, in, parse)
				}
			})
		}
	}
}

// TS 29.002 V19.1.0 §17.7.8 EMLPP-Priority reserves values7..15.
func TestSenderCallPriority(t *testing.T) {
	for _, api := range []string{"PSI", "SRI"} {
		for v := -1; v <= 16; v++ {
			t.Run(fmt.Sprintf("%s/%d", api, v), func(t *testing.T) {
				var in marshaler
				var parse func([]byte) (marshaler, error)
				if api == "PSI" {
					in = &ProvideSubscriberInfo{IMSI: "001010123456789", CallPriority: &v}
					parse = asParser(ParseProvideSubscriberInfo)
				} else {
					in = &Sri{MSISDN: "12345", GmscOrGsmSCFAddress: "12346", CallPriority: &v}
					parse = asParser(ParseSri)
				}
				_, err := in.Marshal()
				switch {
				case v >= 7 && v <= 15:
					if !errors.Is(err, ErrEMLPPPriorityOutOfRange) {
						t.Fatal(err)
					}
				case v < 0 || v > 15:
					var c *ber.ConstraintError
					if !errors.As(err, &c) || errors.Is(err, ErrEMLPPPriorityOutOfRange) {
						t.Fatalf("want codec constraint, got %v", err)
					}
				default:
					senderMobilityRoundTrip(t, in, parse)
				}
			})
		}
	}
}

func senderMobilityLenientCases() []semLenientCase {
	gprs := func(t *testing.T) *gsm_map.UpdateGprsLocationArg {
		return semMust(convertUpdateGprsLocationToArg(senderMobilityGPRS()))
	}
	pdp := func(t *testing.T) *gsm_map.InsertSubscriberDataArg {
		return semMust(convertInsertSubscriberDataArgToWire(senderMobilityISD(senderMobilityPDP())))
	}
	info := func(c gsm_map.PDPContextInfo) *gsm_map.AnyTimeInterrogationRes {
		state := gsm_map.NewPSSubscriberStatePsPDPActiveReachableForPaging(&gsm_map.PDPContextInfoList{Values: []gsm_map.PDPContextInfo{c}})
		return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{PsSubscriberState: &state}}
	}
	requested := func(ri gsm_map.RequestedInfo) func(*testing.T) semWire {
		return func(*testing.T) semWire {
			return &gsm_map.ProvideSubscriberInfoArg{Imsi: semAddrIMSI, RequestedInfo: ri}
		}
	}
	return []semLenientCase{
		{ErrSGSNCapabilityLCSSet1NotAllowed, asParser(ParseUpdateGprsLocation), func(t *testing.T) semWire {
			w := gprs(t)
			bs := convertLCSCapsToBitString(&SupportedLCSCapabilitySets{LcsCapabilitySet1: true})
			w.SgsnCapability = &gsm_map.SGSNCapability{SupportedLCSCapabilitySets: &bs}
			return w
		}},
		{ErrPdnGwUpdateMissingAPNOrContextID, asParser(ParseUpdateGprsLocation), func(t *testing.T) semWire {
			w := gprs(t)
			e := gsm_map.NewEPSInfoPdnGwUpdate(gsm_map.PDNGWUpdate{PdnGwIdentity: &gsm_map.PDNGWIdentity{PdnGwIpv4Address: new([]byte{192, 0, 2, 1})}})
			w.EpsInfo = &e
			return w
		}},
		{ErrPDNTypeInvalid, asParser(ParseInsertSubscriberData), func(t *testing.T) semWire {
			w := semMust(convertInsertSubscriberDataArgToWire(senderMobilityPDN(1)))
			w.EpsSubscriptionData.ApnConfigurationProfile.EpsDataList.Values[0].PdnType = []byte{0}
			return w
		}},
		{ErrPDPAddressFamilyCombinationInvalid, asParser(ParseInsertSubscriberData), func(t *testing.T) semWire {
			w := pdp(t)
			c := &w.GprsSubscriptionData.GprsDataList.Values[0]
			c.PdpAddress = new([]byte{192, 0, 2, 1})
			c.ExtPdpAddress = new([]byte{192, 0, 2, 2})
			return w
		}},
		{ErrPDPTypeInvalid, asParser(ParseAnyTimeInterrogationRes), func(*testing.T) semWire {
			return info(gsm_map.PDPContextInfo{PdpContextIdentifier: 1, PdpType: []byte{0, 0}})
		}},
		{ErrExtPDPTypeInvalid, asParser(ParseAnyTimeInterrogationRes), func(*testing.T) semWire {
			return info(gsm_map.PDPContextInfo{PdpContextIdentifier: 1, PdpType: []byte{0xf1, 0x21}, ExtPdpType: new([]byte{0, 0})})
		}},
		{ErrPSIResImsVoiceOverPSSessionsIndicationUnknown, asParser(ParseProvideSubscriberInfoRes), func(*testing.T) semWire {
			return &gsm_map.ProvideSubscriberInfoRes{SubscriberInfo: gsm_map.SubscriberInfo{ImsVoiceOverPSSessionsIndication: new(IMSVoiceOverPSUnknown)}}
		}},
		{ErrNumberPortabilityStatusATIOnly, asParser(ParseProvideSubscriberInfoRes), func(*testing.T) semWire {
			return &gsm_map.ProvideSubscriberInfoRes{SubscriberInfo: gsm_map.SubscriberInfo{MnpInfoRes: &gsm_map.MNPInfoRes{NumberPortabilityStatus: new(MnpOwnNumberNotPortedOut)}}}
		}},
		{ErrRequestedInfoCurrentLocationWithoutLocationInformation, asParser(ParseProvideSubscriberInfo), requested(gsm_map.RequestedInfo{CurrentLocation: &struct{}{}})},
		{ErrRequestedInfoEPSSupportedWithoutLocationInformation, asParser(ParseProvideSubscriberInfo), requested(gsm_map.RequestedInfo{LocationInformationEPSSupported: &struct{}{}})},
		{ErrRequestedInfoRequestedNodesInCSDomain, asParser(ParseProvideSubscriberInfo), func(*testing.T) semWire {
			bs := convertRequestedNodesToBitString(&RequestedNodes{MME: true})
			return &gsm_map.ProvideSubscriberInfoArg{Imsi: semAddrIMSI, RequestedInfo: gsm_map.RequestedInfo{RequestedNodes: &bs, RequestedDomain: new(CsDomain)}}
		}},
		{ErrRequestedInfoServingNodeWithoutLocationInformation, asParser(ParseProvideSubscriberInfo), requested(gsm_map.RequestedInfo{ServingNodeIndication: &struct{}{}})},
		{ErrRequestedInfoServingNodeWithCurrentLocation, asParser(ParseProvideSubscriberInfo), requested(gsm_map.RequestedInfo{LocationInformation: &struct{}{}, CurrentLocation: &struct{}{}, ServingNodeIndication: &struct{}{}})},
		{ErrGeodeticInformationShapeInvalid, asParser(ParseAnyTimeInterrogationRes), func(*testing.T) semWire {
			return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{LocationInformation: &gsm_map.LocationInformation{GeodeticInformation: new(make([]byte, 10))}}}
		}},
		{ErrPDPContextInfoQoSWithoutPredecessor, asParser(ParseAnyTimeInterrogationRes), func(*testing.T) semWire {
			return info(gsm_map.PDPContextInfo{PdpContextIdentifier: 1, PdpType: []byte{0xf1, 0x21}, Qos2Subscribed: new([]byte{1})})
		}},
		{ErrPDPContextInfoExtPDPAddressWithoutPDPAddress, asParser(ParseAnyTimeInterrogationRes), func(*testing.T) semWire {
			return info(gsm_map.PDPContextInfo{PdpContextIdentifier: 1, PdpType: []byte{0xf1, 0x21}, ExtPdpAddress: new([]byte{192, 0, 2, 1})})
		}},
		{ErrRouteingNumberTBCDInvalid, asParser(ParseAnyTimeInterrogationRes), func(*testing.T) semWire {
			return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{MnpInfoRes: &gsm_map.MNPInfoRes{RouteingNumber: new([]byte{0xff})}}}
		}},
	}
}
