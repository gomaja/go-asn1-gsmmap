package gsmmap

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

type ssSenderPath struct {
	name    string
	message func(HexBytes) marshaler
	parse   func([]byte) (marshaler, error)
}

func ssSenderISD(info ExtSSInfo) *InsertSubscriberDataArg {
	return &InsertSubscriberDataArg{ProvisionedSS: []ExtSSInfo{info}}
}

func ssSenderPrivacy(c LCSPrivacyClass) *InsertSubscriberDataArg {
	return &InsertSubscriberDataArg{LcsInformation: &LCSInformation{LcsPrivacyExceptionList: LCSPrivacyExceptionList{c}}}
}

func ssSenderSri() *Sri {
	return &Sri{MSISDN: "31612345678", GmscOrGsmSCFAddress: "31600000001"}
}

func ssSenderForwardingPaths(field string) []ssSenderPath {
	return []ssSenderPath{
		{"ISD", func(raw HexBytes) marshaler {
			f := ExtForwFeature{SsStatus: HexBytes{1}}
			if field == "options" {
				f.ForwardingOptions = raw
			} else {
				f.ForwardedToSubaddress = raw
			}
			return ssSenderISD(ExtSSInfo{ForwardingInfo: &ExtForwInfo{SsCode: 0x21, ForwardingFeatureList: []ExtForwFeature{f}}})
		}, asParser(ParseInsertSubscriberData)},
		{"SRI", func(raw HexBytes) marshaler {
			f := &ForwardingData{}
			if field == "options" {
				f.ForwardingOptions = raw
			} else {
				f.ForwardedToSubaddress = raw
			}
			return &SriResp{ExtendedRoutingInfo: &ExtendedRoutingInfo{RoutingInfo: &RoutingInfo{ForwardingData: f}}}
		}, asParser(ParseSriResp)},
	}
}

func ssSenderAccepted(t *testing.T, message marshaler, parse func([]byte) (marshaler, error)) {
	t.Helper()
	data, err := message.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := checkParseRoundTrip("sender control", parse, data); err != nil {
		t.Fatal(err)
	}
}

// 3GPP TS 29.002 V19.1.0 §17.7.1 ForwardingOptions and Ext-ForwOptions unused bits.
func TestSenderForwardingOptions(t *testing.T) {
	for _, path := range ssSenderForwardingPaths("options") {
		for code := 0; code <= 255; code++ {
			t.Run(fmt.Sprintf("%s/%02x", path.name, code), func(t *testing.T) {
				m := path.message(HexBytes{byte(code)})
				if code&0x13 != 0 {
					if _, err := m.Marshal(); !errors.Is(err, ErrExtForwOptionsUnusedBits) {
						t.Fatalf("got %v", err)
					}
				} else {
					ssSenderAccepted(t, m, path.parse)
				}
			})
		}
	}
}

// 3GPP TS 29.002 V19.1.0 §17.7.8 ISDN-SubaddressString header.
func TestSenderISDNSubaddress(t *testing.T) {
	for _, path := range ssSenderForwardingPaths("subaddress") {
		for code := 0; code <= 255; code++ {
			t.Run(fmt.Sprintf("%s/%02x", path.name, code), func(t *testing.T) {
				m := path.message(HexBytes{byte(code), 0x21})
				if code != 0x80 && code != 0x88 && code != 0xa0 && code != 0xa8 {
					if _, err := m.Marshal(); !errors.Is(err, ErrISDNSubaddressHeaderInvalid) {
						t.Fatalf("got %v", err)
					}
				} else {
					ssSenderAccepted(t, m, path.parse)
				}
			})
		}
	}
}

// 3GPP TS 29.002 V19.1.0 §17.7.1 InterCUG-Restrictions bits 8..3.
func TestSenderInterCUGRestrictions(t *testing.T) {
	for code := 0; code <= 255; code++ {
		m := ssSenderISD(ExtSSInfo{CugInfo: &CUGInfo{CugFeatureList: []CUGFeature{{InterCUGRestrictions: byte(code)}}}})
		if code > 3 {
			if _, err := m.Marshal(); !errors.Is(err, ErrInterCUGRestrictionsUnusedBits) {
				t.Fatalf("%02x: got %v", code, err)
			}
		} else {
			ssSenderAccepted(t, m, asParser(ParseInsertSubscriberData))
		}
	}
}

func ssSenderStatusPaths() []ssSenderPath {
	parse := asParser(ParseInsertSubscriberData)
	return []ssSenderPath{
		{"forwarding", func(raw HexBytes) marshaler {
			return ssSenderISD(ExtSSInfo{ForwardingInfo: &ExtForwInfo{SsCode: 0x21, ForwardingFeatureList: []ExtForwFeature{{SsStatus: raw}}}})
		}, parse},
		{"barring", func(raw HexBytes) marshaler {
			return ssSenderISD(ExtSSInfo{CallBarringInfo: &ExtCallBarInfo{SsCode: 0x92, CallBarringFeatureList: []ExtCallBarringFeature{{SsStatus: raw}}}})
		}, parse},
		{"SS data", func(raw HexBytes) marshaler {
			return ssSenderISD(ExtSSInfo{SsData: &ExtSSData{SsCode: 0x11, SsStatus: raw}})
		}, parse},
		{"privacy", func(raw HexBytes) marshaler { return ssSenderPrivacy(LCSPrivacyClass{SsCode: 0xb1, SsStatus: raw}) }, parse},
		{"MOLR", func(raw HexBytes) marshaler {
			return &InsertSubscriberDataArg{LcsInformation: &LCSInformation{MolrList: MOLRList{{SsCode: 0xc1, SsStatus: raw}}}}
		}, parse},
		{"multicall", func(raw HexBytes) marshaler {
			return &InsertSubscriberDataArg{McSSInfo: &MCSSInfo{SsCode: 0x45, SsStatus: raw, NbrSB: 2, NbrUser: 1}}
		}, parse},
	}
}

// 3GPP TS 29.002 V19.1.0 §17.7.8 Ext-SS-Status unused bits and reserved octets.
func TestSenderExtSSStatus(t *testing.T) {
	for _, path := range ssSenderStatusPaths() {
		for code := 0; code <= 255; code++ {
			t.Run(fmt.Sprintf("%s/%02x", path.name, code), func(t *testing.T) {
				m := path.message(HexBytes{byte(code)})
				if code&0xf0 != 0 {
					if _, err := m.Marshal(); !errors.Is(err, ErrExtSSStatusUnusedBits) {
						t.Fatalf("got %v", err)
					}
				} else {
					ssSenderAccepted(t, m, path.parse)
				}
			})
		}
		for _, raw := range []HexBytes{{1, 0xff}, {1, 0, 0, 0, 0}} {
			if _, err := path.message(raw).Marshal(); !errors.Is(err, ErrExtSSStatusReservedOctets) {
				t.Fatalf("%s/%x: got %v", path.name, raw, err)
			}
		}
	}
}

func ssSenderCodePaths() []ssSenderPath {
	parse := asParser(ParseInsertSubscriberData)
	return []ssSenderPath{
		{"forwarding", func(raw HexBytes) marshaler {
			return ssSenderISD(ExtSSInfo{ForwardingInfo: &ExtForwInfo{SsCode: SsCode(raw[0]), ForwardingFeatureList: []ExtForwFeature{{SsStatus: HexBytes{1}}}}})
		}, parse},
		{"barring", func(raw HexBytes) marshaler {
			return ssSenderISD(ExtSSInfo{CallBarringInfo: &ExtCallBarInfo{SsCode: SsCode(raw[0]), CallBarringFeatureList: []ExtCallBarringFeature{{SsStatus: HexBytes{1}}}}})
		}, parse},
		{"SS data", func(raw HexBytes) marshaler {
			return ssSenderISD(ExtSSInfo{SsData: &ExtSSData{SsCode: SsCode(raw[0]), SsStatus: HexBytes{1}}})
		}, parse},
		{"privacy", func(raw HexBytes) marshaler {
			return ssSenderPrivacy(LCSPrivacyClass{SsCode: SsCode(raw[0]), SsStatus: HexBytes{1}})
		}, parse},
		{"MOLR", func(raw HexBytes) marshaler {
			return &InsertSubscriberDataArg{LcsInformation: &LCSInformation{MolrList: MOLRList{{SsCode: SsCode(raw[0]), SsStatus: HexBytes{1}}}}}
		}, parse},
		{"multicall", func(raw HexBytes) marshaler {
			return &InsertSubscriberDataArg{McSSInfo: &MCSSInfo{SsCode: SsCode(raw[0]), SsStatus: HexBytes{1}, NbrSB: 2, NbrUser: 1}}
		}, parse},
		{"ISD result", func(raw HexBytes) marshaler { return &InsertSubscriberDataRes{SsList: []SsCode{SsCode(raw[0])}} }, asParser(ParseInsertSubscriberDataRes)},
		{"SRI SsList", func(raw HexBytes) marshaler { return &SriResp{SsList: []SsCode{SsCode(raw[0])}} }, asParser(ParseSriResp)},
		{"SRI SsList2", func(raw HexBytes) marshaler { return &SriResp{SsList2: []SsCode{SsCode(raw[0])}} }, asParser(ParseSriResp)},
	}
}

// 3GPP TS 29.002 V19.1.0 §17.7.5 reserved SS-Codes and ccbs-A/ccbs-B operations.
func TestSenderSSCode(t *testing.T) {
	reserved := []byte{0x00, 0x10, 0x15, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x30, 0x32, 0x40, 0x50, 0x60, 0x70, 0x80, 0xa0}
	for _, path := range ssSenderCodePaths() {
		for _, code := range reserved {
			if path.name == "ISD result" {
				ssSenderAccepted(t, path.message(HexBytes{code}), path.parse)
				continue
			}
			if _, err := path.message(HexBytes{code}).Marshal(); !errors.Is(err, ErrSSCodeReserved) {
				t.Fatalf("%s/%02x: got %v", path.name, code, err)
			}
		}
		for _, code := range []byte{0x20, 0x28, 0x90, 0x41, 0x42, 0x45, 0x43, 0x44} {
			m := path.message(HexBytes{code})
			if (path.name == "SRI SsList" || path.name == "SRI SsList2") && (code == 0x43 || code == 0x44) {
				if _, err := m.Marshal(); !errors.Is(err, ErrSSCodeOperationInvalid) {
					t.Fatalf("%s/%02x: got %v", path.name, code, err)
				}
			} else {
				ssSenderAccepted(t, m, path.parse)
			}
		}
	}
}

// 3GPP TS 29.002 V19.1.0 §17.7.1 LCS-PrivacyClass; codes in §17.7.5.
func TestSenderLCSPrivacyApplicability(t *testing.T) {
	notification := NotifyLocationAllowed
	fields := []struct {
		name    string
		set     func(*LCSPrivacyClass)
		allowed func(SsCode) bool
		want    error
	}{
		{"notification", func(c *LCSPrivacyClass) { c.NotificationToMSUser = &notification }, func(code SsCode) bool { return code == 0xb2 || code == 0xb3 }, ErrLCSPrivacyNotificationNotApplicable},
		{"clients empty", func(c *LCSPrivacyClass) { c.ExternalClientList = ExternalClientList{} }, func(code SsCode) bool { return code == 0xb2 || code == 0xb3 }, ErrLCSPrivacyExternalClientListNotApplicable},
		{"clients", func(c *LCSPrivacyClass) {
			c.ExternalClientList = ExternalClientList{{NotificationToMSUser: &notification}}
		}, func(code SsCode) bool { return code == 0xb2 || code == 0xb3 }, ErrLCSPrivacyExternalClientListNotApplicable},
		{"PLMN", func(c *LCSPrivacyClass) { c.PlmnClientList = PLMNClientList{LCSClientBroadcastService} }, func(code SsCode) bool { return code == 0xb4 }, ErrLCSPrivacyPLMNClientListNotApplicable},
		{"services", func(c *LCSPrivacyClass) { c.ServiceTypeList = ServiceTypeList{{ServiceTypeIdentity: 0}} }, func(code SsCode) bool { return code == 0xb5 }, ErrLCSPrivacyServiceTypeListNotApplicable},
	}
	for _, code := range []SsCode{0xb0, 0xb1, 0xb2, 0xb3, 0xb4, 0xb5} {
		ssSenderAccepted(t, ssSenderPrivacy(LCSPrivacyClass{SsCode: code, SsStatus: HexBytes{1}}), asParser(ParseInsertSubscriberData))
		for _, field := range fields {
			t.Run(fmt.Sprintf("%s/%02x", field.name, code), func(t *testing.T) {
				c := LCSPrivacyClass{SsCode: code, SsStatus: HexBytes{1}}
				field.set(&c)
				m := ssSenderPrivacy(c)
				if field.allowed(code) {
					ssSenderAccepted(t, m, asParser(ParseInsertSubscriberData))
				} else if _, err := m.Marshal(); !errors.Is(err, field.want) {
					t.Fatalf("got %v", err)
				}
			})
		}
	}
	for baseCount := 0; baseCount <= 5; baseCount++ {
		for extCount := 1; extCount <= 7; extCount++ {
			c := LCSPrivacyClass{SsCode: 0xb3, SsStatus: HexBytes{1}, ExternalClientList: make(ExternalClientList, baseCount), ExtExternalClientList: make(ExtExternalClientList, extCount)}
			if baseCount+extCount <= 5 {
				if _, err := ssSenderPrivacy(c).Marshal(); !errors.Is(err, ErrLCSPrivacyExtendedClientCountInvalid) {
					t.Fatalf("%d+%d: got %v", baseCount, extCount, err)
				}
			} else {
				ssSenderAccepted(t, ssSenderPrivacy(c), asParser(ParseInsertSubscriberData))
			}
		}
	}
}

// 3GPP TS 29.002 V19.1.0 §17.7.8 AlertingPattern reserved assignments.
func TestSenderSriAlertingPattern(t *testing.T) {
	for code := 0; code <= 255; code++ {
		m := ssSenderSri()
		m.AlertingPattern = HexBytes{byte(code)}
		if code == 0 || code == 1 || code == 2 || (code >= 4 && code <= 8) {
			ssSenderAccepted(t, m, asParser(ParseSri))
		} else if _, err := m.Marshal(); !errors.Is(err, ErrAlertingPatternReserved) {
			t.Fatalf("%02x: got %v", code, err)
		}
	}
}

// 3GPP TS 29.002 V19.1.0 §17.7.8 LCSServiceTypeID navigation is reserved.
func TestSenderLCSServiceTypeID(t *testing.T) {
	for _, value := range []int64{0, 7, 8, 9, 64, 127} {
		psl := semPSLArg()
		psl.LcsServiceTypeID = &value
		slr := minimalSLRArg()
		slr.LcsServiceTypeID = &value
		isd := ssSenderPrivacy(LCSPrivacyClass{SsCode: 0xb5, SsStatus: HexBytes{1}, ServiceTypeList: ServiceTypeList{{ServiceTypeIdentity: value}}})
		for _, path := range []struct {
			message marshaler
			parse   func([]byte) (marshaler, error)
		}{{psl, asParser(ParseProvideSubscriberLocation)}, {slr, asParser(ParseSubscriberLocationReport)}, {isd, asParser(ParseInsertSubscriberData)}} {
			if value == 8 {
				if _, err := path.message.Marshal(); !errors.Is(err, ErrLCSServiceTypeIDReserved) {
					t.Fatalf("%T: got %v", path.message, err)
				}
			} else {
				ssSenderAccepted(t, path.message, path.parse)
			}
		}
	}
}

func ssSenderServicePaths(bearer bool) []ssSenderPath {
	service := func(raw HexBytes) ExtBasicServiceCode {
		if bearer {
			return ExtBasicServiceCode{ExtBearerService: raw}
		}
		return ExtBasicServiceCode{ExtTeleservice: raw}
	}
	isdParse := asParser(ParseInsertSubscriberData)
	sriParse := asParser(ParseSri)
	resParse := asParser(ParseSriResp)
	paths := []ssSenderPath{
		{"ISD list", func(raw HexBytes) marshaler {
			a := &InsertSubscriberDataArg{}
			if bearer {
				a.BearerServiceList = []HexBytes{raw}
			} else {
				a.TeleserviceList = []HexBytes{raw}
			}
			return a
		}, isdParse},
		{"ISD result list", func(raw HexBytes) marshaler {
			a := &InsertSubscriberDataRes{}
			if bearer {
				a.BearerServiceList = []HexBytes{raw}
			} else {
				a.TeleserviceList = []HexBytes{raw}
			}
			return a
		}, asParser(ParseInsertSubscriberDataRes)},
		{"forwarding basic service", func(raw HexBytes) marshaler {
			s := service(raw)
			return ssSenderISD(ExtSSInfo{ForwardingInfo: &ExtForwInfo{SsCode: 0x21, ForwardingFeatureList: []ExtForwFeature{{SsStatus: HexBytes{1}, BasicService: &s}}}})
		}, isdParse},
		{"barring basic service", func(raw HexBytes) marshaler {
			s := service(raw)
			return ssSenderISD(ExtSSInfo{CallBarringInfo: &ExtCallBarInfo{SsCode: 0x92, CallBarringFeatureList: []ExtCallBarringFeature{{SsStatus: HexBytes{1}, BasicService: &s}}}})
		}, isdParse},
		{"CUG feature", func(raw HexBytes) marshaler {
			s := service(raw)
			return ssSenderISD(ExtSSInfo{CugInfo: &CUGInfo{CugFeatureList: []CUGFeature{{BasicService: &s}}}})
		}, isdParse},
		{"CUG subscription", func(raw HexBytes) marshaler {
			return ssSenderISD(ExtSSInfo{CugInfo: &CUGInfo{CugSubscriptionList: []CUGSubscription{{CugInterlock: HexBytes{1, 2, 3, 4}, BasicServiceGroupList: []ExtBasicServiceCode{service(raw)}}}}})
		}, isdParse},
		{"SS data service list", func(raw HexBytes) marshaler {
			return ssSenderISD(ExtSSInfo{SsData: &ExtSSData{SsCode: 0x11, SsStatus: HexBytes{1}, BasicServiceGroupList: []ExtBasicServiceCode{service(raw)}}})
		}, isdParse},
		{"SRI request basic service", func(raw HexBytes) marshaler {
			a := ssSenderSri()
			s := service(raw)
			a.BasicServiceGroup = &s
			return a
		}, sriParse},
		{"SRI request basic service2", func(raw HexBytes) marshaler {
			a := ssSenderSri()
			s := service(raw)
			a.BasicServiceGroup2 = &s
			return a
		}, sriParse},
		{"SRI result basic service", func(raw HexBytes) marshaler { s := service(raw); return &SriResp{BasicService: &s} }, resParse},
		{"SRI result basic service2", func(raw HexBytes) marshaler { s := service(raw); return &SriResp{BasicService2: &s} }, resParse},
	}
	for _, originating := range []bool{true, false} {
		name := "terminating"
		if originating {
			name = "originating"
		}
		paths = append(paths, ssSenderPath{"ISD CAMEL " + name, func(raw HexBytes) marshaler {
			v := &VlrCamelSubscriptionInfo{}
			if originating {
				v.OBcsmCamelTDPCriteriaList = []OBcsmCamelTDPCriteria{{OBcsmTriggerDetectionPoint: OBcsmTriggerCollectedInfo, BasicServiceCriteria: []ExtBasicServiceCode{service(raw)}}}
			} else {
				v.TBcsmCamelTDPCriteriaList = []TBcsmCamelTDPCriteria{{TBcsmTriggerDetectionPoint: TBcsmTriggerTermAttemptAuthorized, BasicServiceCriteria: []ExtBasicServiceCode{service(raw)}}}
			}
			return &InsertSubscriberDataArg{VlrCamelSubscriptionInfo: v}
		}, isdParse})
		paths = append(paths, ssSenderPath{"SRI CAMEL " + name, func(raw HexBytes) marshaler {
			g := &GmscCamelSubscriptionInfo{}
			if originating {
				g.OBcsmCamelTDPCriteriaList = []OBcsmCamelTDPCriteria{{OBcsmTriggerDetectionPoint: OBcsmTriggerCollectedInfo, BasicServiceCriteria: []ExtBasicServiceCode{service(raw)}}}
			} else {
				g.TBcsmCamelTDPCriteriaList = []TBcsmCamelTDPCriteria{{TBcsmTriggerDetectionPoint: TBcsmTriggerTermAttemptAuthorized, BasicServiceCriteria: []ExtBasicServiceCode{service(raw)}}}
			}
			return &SriResp{ExtendedRoutingInfo: &ExtendedRoutingInfo{CamelRoutingInfo: &CamelRoutingInfo{GmscCamelSubscriptionInfo: *g}}}
		}, resParse})
	}
	return paths
}

// 3GPP TS 29.002 V19.1.0 §17.7.9 and §17.7.10 reserve later code octets,
// restrict compound groups to call-independent SS and assign PLMN codes d0..df.
func TestSenderServiceCodes(t *testing.T) {
	for _, bearer := range []bool{false, true} {
		reserved := ErrExtTeleserviceCodeReservedOctets
		compound := ErrCompoundTeleserviceCodeNotAllowed
		compoundCodes := []byte{0x70, 0x80}
		if bearer {
			reserved = ErrExtBearerServiceCodeReservedOctets
			compound = ErrCompoundBearerServiceCodeNotAllowed
			compoundCodes = []byte{0x50, 0x58, 0x60, 0x68}
		}
		for _, path := range ssSenderServicePaths(bearer) {
			t.Run(fmt.Sprintf("bearer=%t/%s", bearer, path.name), func(t *testing.T) {
				for _, raw := range []HexBytes{{0x11, 1}, {0x11, 0, 0, 0, 0}} {
					if path.name == "ISD result list" {
						ssSenderAccepted(t, path.message(raw), path.parse)
						continue
					}
					if _, err := path.message(raw).Marshal(); !errors.Is(err, reserved) {
						t.Fatalf("%x: got %v", raw, err)
					}
				}
				for _, code := range compoundCodes {
					if path.name == "ISD result list" {
						ssSenderAccepted(t, path.message(HexBytes{code}), path.parse)
						continue
					}
					if _, err := path.message(HexBytes{code}).Marshal(); !errors.Is(err, compound) {
						t.Fatalf("%02x: got %v", code, err)
					}
				}
				if bearer {
					for _, code := range []byte{0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xe1, 0xf1} {
						if path.name == "ISD result list" {
							ssSenderAccepted(t, path.message(HexBytes{code}), path.parse)
							continue
						}
						if _, err := path.message(HexBytes{code}).Marshal(); !errors.Is(err, ErrBearerServiceCodeUnusedBit) {
							t.Fatalf("%02x: got %v", code, err)
						}
					}
				}
				for _, code := range []byte{0x00, 0x10, 0x11, 0x20, 0xd0, 0xd1, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xdb, 0xdc, 0xdd, 0xde, 0xdf} {
					ssSenderAccepted(t, path.message(HexBytes{code}), path.parse)
				}
			})
		}
	}
}

// 3GPP TS 29.002 V19.1.0 §§17.7.1, 17.7.5, 17.7.8-17.7.10 sender rules
// retain these values on reception.
func ssSenderLenientCases() []semLenientCase {
	privacy := func(change func(*gsm_map.LCSPrivacyClass)) func(*testing.T) semWire {
		return func(*testing.T) semWire {
			w := semMust(convertInsertSubscriberDataArgToWire(ssSenderPrivacy(LCSPrivacyClass{SsCode: 0xb1, SsStatus: HexBytes{1}})))
			change(&w.LcsInformation.LcsPrivacyExceptionList.Values[0])
			return w
		}
	}
	service := func(bearer bool, raw []byte) func(*testing.T) semWire {
		return func(*testing.T) semWire {
			w := &gsm_map.InsertSubscriberDataArg{}
			if bearer {
				w.BearerServiceList = &gsm_map.BearerServiceList{Values: []gsm_map.ExtBearerServiceCode{raw}}
			} else {
				w.TeleserviceList = &gsm_map.TeleserviceList{Values: []gsm_map.ExtTeleserviceCode{raw}}
			}
			return w
		}
	}
	parse := asParser(ParseInsertSubscriberData)
	return []semLenientCase{
		{ErrExtForwOptionsUnusedBits, asParser(ParseSriResp), func(*testing.T) semWire {
			w := semMust(convertSriRespToRes(&SriResp{ExtendedRoutingInfo: &ExtendedRoutingInfo{RoutingInfo: &RoutingInfo{ForwardingData: &ForwardingData{ForwardingOptions: HexBytes{0x00}}}}}))
			*w.ExtendedRoutingInfo.RoutingInfo.ForwardingData.ForwardingOptions = gsm_map.ForwardingOptions{0x13}
			return w
		}},
		{ErrISDNSubaddressHeaderInvalid, asParser(ParseSriResp), func(*testing.T) semWire {
			w := semMust(convertSriRespToRes(&SriResp{ExtendedRoutingInfo: &ExtendedRoutingInfo{RoutingInfo: &RoutingInfo{ForwardingData: &ForwardingData{ForwardedToSubaddress: HexBytes{0x80, 0x21}}}}}))
			*w.ExtendedRoutingInfo.RoutingInfo.ForwardingData.ForwardedToSubaddress = gsm_map.ISDNSubaddressString{0x90, 0x21}
			return w
		}},
		{ErrLCSPrivacyNotificationNotApplicable, parse, privacy(func(c *gsm_map.LCSPrivacyClass) {
			v := gsm_map.NotificationToMSUserNotifyLocationAllowed
			c.NotificationToMSUser = &v
		})},
		{ErrLCSPrivacyExternalClientListNotApplicable, parse, privacy(func(c *gsm_map.LCSPrivacyClass) {
			c.ExternalClientList = &gsm_map.ExternalClientList{Values: []gsm_map.ExternalClient{}}
		})},
		{ErrLCSPrivacyPLMNClientListNotApplicable, parse, privacy(func(c *gsm_map.LCSPrivacyClass) {
			c.PlmnClientList = &gsm_map.PLMNClientList{Values: []gsm_map.LCSClientInternalID{gsm_map.LCSClientInternalIDBroadcastService}}
		})},
		{ErrLCSPrivacyExtendedClientCountInvalid, parse, privacy(func(c *gsm_map.LCSPrivacyClass) {
			c.ExtExternalClientList = &gsm_map.ExtExternalClientList{Values: []gsm_map.ExternalClient{{ClientIdentity: gsm_map.LCSClientExternalID{}}}}
		})},
		{ErrLCSPrivacyServiceTypeListNotApplicable, parse, privacy(func(c *gsm_map.LCSPrivacyClass) {
			c.ServiceTypeList = &gsm_map.ServiceTypeList{Values: []gsm_map.ServiceType{{ServiceTypeIdentity: 0}}}
		})},
		{ErrSSCodeReserved, asParser(ParseSriResp), func(*testing.T) semWire {
			return &gsm_map.SendRoutingInfoRes{SsList: &gsm_map.SSList{Values: []gsm_map.SSCode{{0x15}}}}
		}},
		{ErrSSCodeOperationInvalid, asParser(ParseSriResp), func(*testing.T) semWire {
			return &gsm_map.SendRoutingInfoRes{SsList: &gsm_map.SSList{Values: []gsm_map.SSCode{{0x43}}}}
		}},
		{ErrLCSServiceTypeIDReserved, asParser(ParseProvideSubscriberLocation), func(*testing.T) semWire {
			w := semMust(convertProvideSubscriberLocationArgToWire(semPSLArg()))
			v := int64(8)
			w.LcsServiceTypeID = &v
			return w
		}},
		{ErrExtSSStatusUnusedBits, parse, privacy(func(c *gsm_map.LCSPrivacyClass) { c.SsStatus = gsm_map.ExtSSStatus{0xf1} })},
		{ErrExtTeleserviceCodeReservedOctets, parse, service(false, []byte{0x11, 1})},
		{ErrExtBearerServiceCodeReservedOctets, parse, service(true, []byte{0x11, 1})},
		{ErrCompoundTeleserviceCodeNotAllowed, parse, service(false, []byte{0x70})},
		{ErrBearerServiceCodeUnusedBit, parse, service(true, []byte{0x81})},
		{ErrCompoundBearerServiceCodeNotAllowed, parse, service(true, []byte{0x50})},
	}
}

// TS 29.002 V19.1.0 §8.8.1.3 returns unsupported service codes received in ISD.
func TestSenderISDResultEcho(t *testing.T) {
	for code := 0; code <= 255; code++ {
		for length := 1; length <= 5; length++ {
			raw := make(HexBytes, length)
			raw[0] = byte(code)
			for i := 1; i < length; i++ {
				raw[i] = byte(i)
			}
			in := &InsertSubscriberDataRes{SsList: []SsCode{SsCode(code)}, TeleserviceList: []HexBytes{raw}, BearerServiceList: []HexBytes{raw}}
			data, err := in.Marshal()
			if err != nil {
				t.Fatalf("%02x/%d Marshal: %v", code, length, err)
			}
			got, err := ParseInsertSubscriberDataRes(data)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if !reflect.DeepEqual(in, got) {
				t.Fatalf("%02x/%d result changed: got %+v, want %+v", code, length, got, in)
			}
		}
	}
	for _, bearer := range []bool{false, true} {
		for _, raw := range []HexBytes{{}, {0x11, 0, 0, 0, 0, 0}} {
			in := &InsertSubscriberDataRes{}
			if bearer {
				in.BearerServiceList = []HexBytes{raw}
			} else {
				in.TeleserviceList = []HexBytes{raw}
			}
			if _, err := in.Marshal(); err == nil {
				t.Fatalf("bearer=%t/%x: accepted invalid SIZE", bearer, raw)
			}
		}
	}
}

// TS 29.002 V19.1.0 §17.7.1 Ext-ForwOptions reserves octets 2..5.
func TestSenderExtForwOptionsReservedOctets(t *testing.T) {
	path := ssSenderForwardingPaths("options")[0]
	for length := 2; length <= 5; length++ {
		raw := make(HexBytes, length)
		raw[0] = 4
		if _, err := path.message(raw).Marshal(); !errors.Is(err, ErrExtForwOptionsReservedOctets) {
			t.Fatalf("length %d: got %v", length, err)
		}
		w := semMust(convertInsertSubscriberDataArgToWire(path.message(HexBytes{4}).(*InsertSubscriberDataArg)))
		w.ProvisionedSS.Values[0].ForwardingInfo.ForwardingFeatureList.Values[0].ForwardingOptions = new(gsm_map.ExtForwOptions(raw))
		got, err := ParseInsertSubscriberData(strictBER(t, w))
		if err != nil {
			t.Fatal(err)
		}
		options := got.ProvisionedSS[0].ForwardingInfo.ForwardingFeatureList[0].ForwardingOptions
		if !reflect.DeepEqual(options, HexBytes{4}) {
			t.Fatalf("Parse kept reserved octets: %x", options)
		}
		if _, err := got.Marshal(); err != nil {
			t.Fatalf("Marshal after Parse: %v", err)
		}
	}
	for _, path := range ssSenderForwardingPaths("options") {
		for _, raw := range []HexBytes{{}, {4, 0}, {4, 0, 0, 0, 0, 0}} {
			if path.name == "SRI" && len(raw) == 0 {
				continue
			}
			if _, err := path.message(raw).Marshal(); err == nil {
				t.Fatalf("%s/%x: accepted invalid SIZE", path.name, raw)
			}
		}
	}
}

func TestSenderISDResultEchoReceivedCodes(t *testing.T) {
	w := semMust(convertInsertSubscriberDataArgToWire(ssSenderISD(ExtSSInfo{SsData: &ExtSSData{SsCode: 0x21, SsStatus: HexBytes{1}}})))
	w.ProvisionedSS.Values[0].SsData.SsCode = gsm_map.SSCode{0x15}
	w.TeleserviceList = &gsm_map.TeleserviceList{Values: []gsm_map.ExtTeleserviceCode{{0x11, 0}}}
	w.BearerServiceList = &gsm_map.BearerServiceList{Values: []gsm_map.ExtBearerServiceCode{{0x90}}}
	received, err := ParseInsertSubscriberData(strictBER(t, w))
	if err != nil {
		t.Fatal(err)
	}
	in := &InsertSubscriberDataRes{SsList: []SsCode{received.ProvisionedSS[0].SsData.SsCode}, TeleserviceList: received.TeleserviceList, BearerServiceList: received.BearerServiceList}
	senderMobilityRoundTrip(t, in, asParser(ParseInsertSubscriberDataRes))
}
