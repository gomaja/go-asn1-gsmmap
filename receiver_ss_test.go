package gsmmap

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime"
)

// TS 29.002 V19.1.0 §17.1.2 and §17.6.6: roamingNotAllowed has a
// semantically mandatory parameter, unlike illegalSubscriber.
func TestReceiverRoamingNotAllowedRequiresParameter(t *testing.T) {
	for _, data := range [][]byte{nil, {}} {
		got, err := ParseReturnErrorParameter(MapErrorRoamingNotAllowed, data)
		if got != nil || !errors.Is(err, ErrRoamingNotAllowedParameterMissing) {
			t.Errorf("empty parameter: got %v, %v", got, err)
		}
		got, err = ParseReturnErrorParameter(MapErrorIllegalSubscriber, data)
		if got != nil || err != nil {
			t.Errorf("optional parameter: got %v, %v", got, err)
		}
	}
	w := &gsm_map.RoamingNotAllowedParam{RoamingNotAllowedCause: gsm_map.RoamingNotAllowedCause(0)}
	got, err := ParseReturnErrorParameter(MapErrorRoamingNotAllowed, strictBER(t, w))
	param, ok := got.(*RoamingNotAllowedParam)
	if err != nil || !ok || param.RoamingNotAllowedCause == nil || *param.RoamingNotAllowedCause != 0 {
		t.Fatalf("valid parameter: got %v, %v", got, err)
	}
}

// TS 29.002 V19.1.0 §17.7.1 Ext-ForwOptions and InterCUG-Restrictions.
func TestReceiverForwardingOptionsDiscardLaterOctets(t *testing.T) {
	for _, options := range []HexBytes{{0xfc}, {0xfc, 0}, {0xfc, 0xff}, {0xfc, 1, 2}, {0xfc, 1, 2, 3}, {0xfc, 1, 2, 3, 4}} {
		t.Run(fmt.Sprintf("%x", []byte(options)), func(t *testing.T) {
			in := &InsertSubscriberDataArg{ProvisionedSS: []ExtSSInfo{{ForwardingInfo: &ExtForwInfo{
				SsCode: 0x21, ForwardingFeatureList: []ExtForwFeature{{SsStatus: HexBytes{1}, ForwardingOptions: HexBytes{0xec}}},
			}}}}
			data, err := in.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			var input gsm_map.InsertSubscriberDataArg
			if err := input.UnmarshalBER(data); err != nil {
				t.Fatal(err)
			}
			*input.ProvisionedSS.Values[0].ForwardingInfo.ForwardingFeatureList.Values[0].ForwardingOptions = gsm_map.ExtForwOptions(options)
			data = strictBER(t, &input)
			got, err := ParseInsertSubscriberData(data)
			if err != nil {
				t.Fatal(err)
			}
			info := got.ProvisionedSS[0].ForwardingInfo
			if info.SsCode != 0x21 || !slices.Equal(info.ForwardingFeatureList[0].ForwardingOptions, HexBytes{0xfc}) {
				t.Fatalf("received forwarding info: %+v", info)
			}
			if _, err := got.Marshal(); !errors.Is(err, ErrExtForwOptionsUnusedBits) {
				t.Fatalf("reserved bits: %v", err)
			}
		})
	}
}

func TestReceiverCUGRestrictionsIgnoreUnsupportedBits(t *testing.T) {
	for _, value := range []byte{0, 1, 2, 3, 0xfc, 0xff} {
		t.Run(fmt.Sprintf("%02x", value), func(t *testing.T) {
			idx := 42
			in := &InsertSubscriberDataArg{ProvisionedSS: []ExtSSInfo{{CugInfo: &CUGInfo{
				CugFeatureList: []CUGFeature{{InterCUGRestrictions: value & 3, PreferentialCUGIndex: &idx}},
			}}}}
			data, err := in.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			var w gsm_map.InsertSubscriberDataArg
			if err := w.UnmarshalBER(data); err != nil {
				t.Fatal(err)
			}
			w.ProvisionedSS.Values[0].CugInfo.CugFeatureList.Values[0].InterCUGRestrictions = gsm_map.InterCUGRestrictions{value}
			data = strictBER(t, &w)
			got, err := ParseInsertSubscriberData(data)
			if err != nil {
				t.Fatal(err)
			}
			feature := got.ProvisionedSS[0].CugInfo.CugFeatureList[0]
			if feature.InterCUGRestrictions != value&3 || feature.PreferentialCUGIndex == nil || *feature.PreferentialCUGIndex != idx {
				t.Fatalf("received CUG feature: %+v", feature)
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.8 EMLPP-Priority: spare 7..15 map to 4.
func TestReceiverCallPriorityMapsSpareValues(t *testing.T) {
	for priority := -1; priority <= 15; priority++ {
		t.Run(fmt.Sprint(priority), func(t *testing.T) {
			psi := &ProvideSubscriberInfo{IMSI: "001010123456789", RequestedInfo: RequestedInfo{SubscriberState: true}}
			sri := &Sri{MSISDN: "31612345678", GmscOrGsmSCFAddress: "31612345679", InterrogationType: InterrogationBasicCall}
			if priority >= 0 {
				psi.CallPriority, sri.CallPriority = &priority, &priority
			}
			var psiData, sriData []byte
			if priority >= 7 {
				if _, err := psi.Marshal(); !errors.Is(err, ErrEMLPPPriorityOutOfRange) {
					t.Fatal(err)
				}
				if _, err := sri.Marshal(); !errors.Is(err, ErrEMLPPPriorityOutOfRange) {
					t.Fatal(err)
				}
				psi.CallPriority, sri.CallPriority = nil, nil
				pw := semMust(convertProvideSubscriberInfoToArg(psi))
				sw := semMust(convertSriToArg(sri))
				pw.CallPriority, sw.CallPriority = new(int64(priority)), new(int64(priority))
				psiData, sriData = strictBER(t, pw), strictBER(t, sw)
			} else {
				psiData = semMust(psi.Marshal())
				sriData = semMust(sri.Marshal())
			}
			var psiWire gsm_map.ProvideSubscriberInfoArg
			if err := psiWire.UnmarshalBER(psiData); err != nil {
				t.Fatal(err)
			}
			var sriWire gsm_map.SendRoutingInfoArg
			if err := sriWire.UnmarshalBER(sriData); err != nil {
				t.Fatal(err)
			}
			for name, sent := range map[string]*int64{"PSI": psiWire.CallPriority, "SRI": sriWire.CallPriority} {
				if priority < 0 {
					if sent != nil {
						t.Errorf("%s sender added priority", name)
					}
				} else if sent == nil || *sent != int64(priority) {
					t.Errorf("%s sender changed priority", name)
				}
			}
			psiGot, err := ParseProvideSubscriberInfo(psiData)
			if err != nil {
				t.Fatal(err)
			}
			sriGot, err := ParseSri(sriData)
			if err != nil {
				t.Fatal(err)
			}
			for name, got := range map[string]*int{"PSI": psiGot.CallPriority, "SRI": sriGot.CallPriority} {
				want := priority
				if want >= 7 {
					want = 4
				}
				if priority < 0 {
					if got != nil {
						t.Errorf("%s absent priority became %v", name, got)
					}
				} else if got == nil || *got != want {
					t.Errorf("%s priority: got %v, want %d", name, got, want)
				}
			}
			if priority >= 0 {
				w := &gsm_map.InsertSubscriberDataArg{ProvisionedSS: &gsm_map.ExtSSInfoList{Values: []gsm_map.ExtSSInfo{
					gsm_map.NewExtSSInfoEmlppInfo(gsm_map.EMLPPInfo{MaximumentitledPriority: int64(priority), DefaultPriority: int64(priority)}),
				}}}
				got, err := ParseInsertSubscriberData(strictBER(t, w))
				if err != nil {
					t.Fatal(err)
				}
				want := priority
				if want >= 7 {
					want = 4
				}
				info := got.ProvisionedSS[0].EmlppInfo
				if info.DefaultPriority != want || info.MaximumEntitledPriority != want {
					t.Errorf("EMLPPInfo: %+v, want %d", info, want)
				}
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.8 Ext-SS-Status: discard ununderstood octets 2..5,
// retaining every bit of octet 1 across all six carriers.
func TestReceiverSSStatusDiscardLaterOctets(t *testing.T) {
	for _, status := range []HexBytes{{1}, {1, 0xff}, {1, 2, 3, 4, 5}, {0xf1, 0xff}} {
		t.Run(fmt.Sprintf("%x", []byte(status)), func(t *testing.T) {
			in := &InsertSubscriberDataArg{
				ProvisionedSS: []ExtSSInfo{
					{ForwardingInfo: &ExtForwInfo{SsCode: 0x21, ForwardingFeatureList: []ExtForwFeature{{SsStatus: HexBytes{1}}}}},
					{CallBarringInfo: &ExtCallBarInfo{SsCode: 0x91, CallBarringFeatureList: []ExtCallBarringFeature{{SsStatus: HexBytes{1}}}}},
					{SsData: &ExtSSData{SsCode: 0x11, SsStatus: HexBytes{1}}},
				},
				McSSInfo: &MCSSInfo{SsCode: 0x11, SsStatus: HexBytes{1}, NbrSB: 2, NbrUser: 1},
				LcsInformation: &LCSInformation{
					LcsPrivacyExceptionList: LCSPrivacyExceptionList{{SsCode: 0xa1, SsStatus: HexBytes{1}}},
					MolrList:                MOLRList{{SsCode: 0xb0, SsStatus: HexBytes{1}}},
				},
			}
			data, err := in.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			var input gsm_map.InsertSubscriberDataArg
			if err := input.UnmarshalBER(data); err != nil {
				t.Fatal(err)
			}
			input.ProvisionedSS.Values[0].ForwardingInfo.ForwardingFeatureList.Values[0].SsStatus = gsm_map.ExtSSStatus(status)
			input.ProvisionedSS.Values[1].CallBarringInfo.CallBarringFeatureList.Values[0].SsStatus = gsm_map.ExtSSStatus(status)
			input.ProvisionedSS.Values[2].SsData.SsStatus = gsm_map.ExtSSStatus(status)
			input.McSSInfo.SsStatus = gsm_map.ExtSSStatus(status)
			input.LcsInformation.LcsPrivacyExceptionList.Values[0].SsStatus = gsm_map.ExtSSStatus(status)
			input.LcsInformation.MolrList.Values[0].SsStatus = gsm_map.ExtSSStatus(status)
			data = strictBER(t, &input)
			got, err := ParseInsertSubscriberData(data)
			if err != nil {
				t.Fatal(err)
			}
			for name, value := range map[string]HexBytes{
				"forwarding": got.ProvisionedSS[0].ForwardingInfo.ForwardingFeatureList[0].SsStatus,
				"barring":    got.ProvisionedSS[1].CallBarringInfo.CallBarringFeatureList[0].SsStatus,
				"ss data":    got.ProvisionedSS[2].SsData.SsStatus,
				"MCSS":       got.McSSInfo.SsStatus,
				"privacy":    got.LcsInformation.LcsPrivacyExceptionList[0].SsStatus,
				"MOLR":       got.LcsInformation.MolrList[0].SsStatus,
			} {
				if !slices.Equal(value, status[:1]) {
					t.Errorf("%s status = %x, want %x", name, []byte(value), []byte(status[:1]))
				}
			}
			if status[0]&0xf0 != 0 {
				if _, err := got.Marshal(); !errors.Is(err, ErrExtSSStatusUnusedBits) {
					t.Fatalf("reserved status bits: %v", err)
				}
				return
			}
			out, err := got.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			var w gsm_map.InsertSubscriberDataArg
			if err := w.UnmarshalBER(out); err != nil {
				t.Fatal(err)
			}
			for _, value := range [][]byte{
				w.ProvisionedSS.Values[0].ForwardingInfo.ForwardingFeatureList.Values[0].SsStatus,
				w.ProvisionedSS.Values[1].CallBarringInfo.CallBarringFeatureList.Values[0].SsStatus,
				w.ProvisionedSS.Values[2].SsData.SsStatus, w.McSSInfo.SsStatus,
				w.LcsInformation.LcsPrivacyExceptionList.Values[0].SsStatus, w.LcsInformation.MolrList.Values[0].SsStatus,
			} {
				if !slices.Equal(value, status[:1]) {
					t.Error("Marshal reintroduced discarded status octets")
				}
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.11 ExtensionSet: "Unsupported private extensions
// shall be discarded if received." Only the caller knows which it supports,
// so an opaque PDP context keeps them, with every other extension.
func TestOpaquePDPContextKeepsExtensions(t *testing.T) {
	ctx := gsm_map.PDPContextInfo{
		PdpContextIdentifier: 1, PdpType: gsm_map.PDPType{0xf1, 0x21},
		ExtensionContainer: &gsm_map.ExtensionContainer{
			PrivateExtensionList: &gsm_map.PrivateExtensionList{Values: []gsm_map.PrivateExtension{{ExtId: runtime.MustObjectIdentifierFromArcs(1, 2, 3)}}},
			PcsExtensions:        &gsm_map.PCSExtensions{ExtCount_: 1, ExtPresent_: []bool{true}, ExtData_: [][]byte{{0x85, 1, 0xaa}}},
			ExtCount_:            1, ExtPresent_: []bool{true}, ExtData_: [][]byte{{0x86, 1, 0xbb}},
		},
		ExtCount_: 1, ExtPresent_: []bool{true}, ExtData_: [][]byte{{0x9e, 1, 0xcc}},
	}
	input := strictBER(t, &ctx)
	for _, reachable := range []bool{false, true} {
		for _, eps := range []bool{false, true} {
			for _, api := range []string{"PSI", "ATI", "SRI"} {
				t.Run(fmt.Sprintf("%s/reachable=%v/EPS=%v", api, reachable, eps), func(t *testing.T) {
					state := &PsSubscriberState{}
					if reachable {
						state.PsPDPActiveReachableForPaging = []HexBytes{input}
					} else {
						state.PsPDPActiveNotReachableForPaging = []HexBytes{input}
					}
					si := SubscriberInfo{}
					if eps {
						si.EpsSubscriberState = state
					} else {
						si.PsSubscriberState = state
					}
					var data []byte
					var got *SubscriberInfo
					var err error
					switch api {
					case "PSI":
						data, err = (&ProvideSubscriberInfoRes{SubscriberInfo: si}).Marshal()
					case "ATI":
						data, err = (&AnyTimeInterrogationRes{SubscriberInfo: si}).Marshal()
					case "SRI":
						data, err = (&SriResp{SubscriberInfo: &si}).Marshal()
					}
					if err != nil {
						t.Fatal(err)
					}
					// Sender-supplied properly located private extensions still survive Marshal.
					var wireSI *gsm_map.SubscriberInfo
					switch api {
					case "PSI":
						var w gsm_map.ProvideSubscriberInfoRes
						if err := w.UnmarshalBER(data); err != nil {
							t.Fatal(err)
						}
						wireSI = &w.SubscriberInfo
						var p *ProvideSubscriberInfoRes
						p, err = ParseProvideSubscriberInfoRes(data)
						if p != nil {
							got = &p.SubscriberInfo
						}
					case "ATI":
						var w gsm_map.AnyTimeInterrogationRes
						if err := w.UnmarshalBER(data); err != nil {
							t.Fatal(err)
						}
						wireSI = &w.SubscriberInfo
						var p *AnyTimeInterrogationRes
						p, err = ParseAnyTimeInterrogationRes(data)
						if p != nil {
							got = &p.SubscriberInfo
						}
					case "SRI":
						var w gsm_map.SendRoutingInfoRes
						if err := w.UnmarshalBER(data); err != nil {
							t.Fatal(err)
						}
						wireSI = w.SubscriberInfo
						var p *SriResp
						p, err = ParseSriResp(data)
						if p != nil {
							got = p.SubscriberInfo
						}
					}
					if err != nil {
						t.Fatal(err)
					}
					wireState := wireSI.PsSubscriberState
					parsedState := got.PsSubscriberState
					if eps {
						wireState, parsedState = wireSI.EpsSubscriberState, got.EpsSubscriberState
					}
					wireList := wireState.PsPDPActiveNotReachableForPaging
					parsedList := parsedState.PsPDPActiveNotReachableForPaging
					if reachable {
						wireList, parsedList = wireState.PsPDPActiveReachableForPaging, parsedState.PsPDPActiveReachableForPaging
					}
					if wireList.Values[0].ExtensionContainer.PrivateExtensionList == nil {
						t.Fatal("sender private extension lost")
					}
					var parsed gsm_map.PDPContextInfo
					if err := parsed.UnmarshalBER(parsedList[0]); err != nil {
						t.Fatal(err)
					}
					if parsed.ExtensionContainer == nil || parsed.ExtensionContainer.PrivateExtensionList == nil {
						t.Errorf("private extension lost: %+v", parsed.ExtensionContainer)
					}
					if parsed.ExtensionContainer == nil || parsed.ExtensionContainer.PcsExtensions == nil || len(parsed.ExtData_) != 1 ||
						len(parsed.ExtensionContainer.ExtData_) != 1 || len(parsed.ExtensionContainer.PcsExtensions.ExtData_) != 1 {
						t.Fatal("unrelated extensions lost")
					}
					if parsed.PdpContextIdentifier != 1 || !slices.Equal(parsed.PdpType, ctx.PdpType) ||
						!slices.Equal(parsed.ExtData_[0], ctx.ExtData_[0]) || !slices.Equal(parsed.ExtensionContainer.ExtData_[0], ctx.ExtensionContainer.ExtData_[0]) ||
						!slices.Equal(parsed.ExtensionContainer.PcsExtensions.ExtData_[0], ctx.ExtensionContainer.PcsExtensions.ExtData_[0]) {
						t.Error("ordinary context or unrelated extension fields lost")
					}
				})
			}
		}
	}
}

func TestReceiverOctetHelpersReturnIndependentCopies(t *testing.T) {
	for name, normalize := range map[string]func([]byte) HexBytes{
		"Ext-SS-Status":   keepFirstOctet,
		"Ext-ForwOptions": keepFirstOctet,
	} {
		t.Run(name, func(t *testing.T) {
			for _, empty := range [][]byte{nil, {}} {
				if got := normalize(empty); got != nil {
					t.Errorf("empty input produced %x", []byte(got))
				}
			}
			input := []byte{0xf1, 0xff}
			got := normalize(input)
			got[0] = 0
			if input[0] != 0xf1 {
				t.Fatal("receiver normalization mutated wire storage")
			}
		})
	}
}
