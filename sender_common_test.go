package gsmmap

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// TS 29.002 V19.1.0 §17.7.8 assigns six address natures and seven numbering plans.
func TestSenderAddressAssignments(t *testing.T) {
	for _, tc := range append(semAddressCases(), semOptionalAddressCases()...) {
		t.Run(tc.name, func(t *testing.T) {
			for _, nature := range []byte{0, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70} {
				senderAddressAssignment(t, tc, nature|1, nature != 0x50 && nature != 0x70, ErrAddressNatureInvalid)
			}
			for plan := byte(0); plan < 16; plan++ {
				valid := plan == 0 || plan == 1 || plan == 3 || plan == 4 || plan == 6 || plan == 8 || plan == 9
				senderAddressAssignment(t, tc, 0x10|plan, valid, ErrAddressPlanInvalid)
			}
		})
	}
}

func senderAddressAssignment(t *testing.T, tc semAddressCase, header byte, valid bool, want error) {
	t.Helper()
	raw := []byte{0x80 | header, 0x21}
	wire := tc.build(t, raw)
	b, err := wire.MarshalBER()
	if err != nil {
		t.Fatal(err)
	}
	v, err := tc.parse(b)
	if err != nil {
		t.Fatalf("parse header %02x: %v", header, err)
	}
	b, err = v.Marshal()
	if !valid {
		if !errors.Is(err, want) {
			t.Fatalf("header %02x: %v, want %v", header, err, want)
		}
		return
	}
	if err != nil {
		t.Fatalf("header %02x: %v", header, err)
	}
	if _, err = tc.parse(b); err != nil {
		t.Fatal(err)
	}
}

type senderOctetCase struct {
	name  string
	build func(*testing.T, HexBytes) marshaler
	parse func([]byte) (marshaler, error)
}

func senderUpdateLocation() *UpdateLocation {
	return &UpdateLocation{IMSI: "001010123456789", MscNumber: "1", VlrNumber: "2"}
}

func senderDiameterCases() []senderOctetCase {
	return []senderOctetCase{
		{"AlertServiceCentre.SmsGmscDiameterAddress.DiameterName", func(t *testing.T, b HexBytes) marshaler {
			v := &AlertServiceCentre{MSISDN: "1", ServiceCentreAddress: "2"}
			v.SmsGmscDiameterAddress = &NetworkNodeDiameterAddress{DiameterName: b, DiameterRealm: HexBytes("node.test")}
			return v
		}, asParser(ParseAlertServiceCentre)},
		{"AlertServiceCentre.SmsGmscDiameterAddress.DiameterRealm", func(t *testing.T, b HexBytes) marshaler {
			v := &AlertServiceCentre{MSISDN: "1", ServiceCentreAddress: "2"}
			v.SmsGmscDiameterAddress = &NetworkNodeDiameterAddress{DiameterRealm: b, DiameterName: HexBytes("node.test")}
			return v
		}, asParser(ParseAlertServiceCentre)},
		{"AlertServiceCentre.NewSGSNDiameterAddress.DiameterName", func(t *testing.T, b HexBytes) marshaler {
			v := &AlertServiceCentre{MSISDN: "1", ServiceCentreAddress: "2"}
			v.NewSGSNDiameterAddress = &NetworkNodeDiameterAddress{DiameterName: b, DiameterRealm: HexBytes("node.test")}
			return v
		}, asParser(ParseAlertServiceCentre)},
		{"AlertServiceCentre.NewSGSNDiameterAddress.DiameterRealm", func(t *testing.T, b HexBytes) marshaler {
			v := &AlertServiceCentre{MSISDN: "1", ServiceCentreAddress: "2"}
			v.NewSGSNDiameterAddress = &NetworkNodeDiameterAddress{DiameterRealm: b, DiameterName: HexBytes("node.test")}
			return v
		}, asParser(ParseAlertServiceCentre)},
		{"AlertServiceCentre.NewMMEDiameterAddress.DiameterName", func(t *testing.T, b HexBytes) marshaler {
			v := &AlertServiceCentre{MSISDN: "1", ServiceCentreAddress: "2"}
			v.NewMMEDiameterAddress = &NetworkNodeDiameterAddress{DiameterName: b, DiameterRealm: HexBytes("node.test")}
			return v
		}, asParser(ParseAlertServiceCentre)},
		{"AlertServiceCentre.NewMMEDiameterAddress.DiameterRealm", func(t *testing.T, b HexBytes) marshaler {
			v := &AlertServiceCentre{MSISDN: "1", ServiceCentreAddress: "2"}
			v.NewMMEDiameterAddress = &NetworkNodeDiameterAddress{DiameterRealm: b, DiameterName: HexBytes("node.test")}
			return v
		}, asParser(ParseAlertServiceCentre)},
		{"SriSmResp.LocationInfoWithLMSI.NetworkNodeDiameterAddress.DiameterName", func(t *testing.T, b HexBytes) marshaler {
			v := &SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "1"}}
			v.LocationInfoWithLMSI.NetworkNodeDiameterAddress = &NetworkNodeDiameterAddress{DiameterName: b, DiameterRealm: HexBytes("node.test")}
			return v
		}, asParser(ParseSriSmResp)},
		{"SriSmResp.LocationInfoWithLMSI.NetworkNodeDiameterAddress.DiameterRealm", func(t *testing.T, b HexBytes) marshaler {
			v := &SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "1"}}
			v.LocationInfoWithLMSI.NetworkNodeDiameterAddress = &NetworkNodeDiameterAddress{DiameterRealm: b, DiameterName: HexBytes("node.test")}
			return v
		}, asParser(ParseSriSmResp)},
		{"SriSmResp.LocationInfoWithLMSI.AdditionalNetworkNodeDiameterAddress.DiameterName", func(t *testing.T, b HexBytes) marshaler {
			v := &SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "1"}}
			v.LocationInfoWithLMSI.AdditionalNetworkNodeDiameterAddress = &NetworkNodeDiameterAddress{DiameterName: b, DiameterRealm: HexBytes("node.test")}
			return v
		}, asParser(ParseSriSmResp)},
		{"SriSmResp.LocationInfoWithLMSI.AdditionalNetworkNodeDiameterAddress.DiameterRealm", func(t *testing.T, b HexBytes) marshaler {
			v := &SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "1"}}
			v.LocationInfoWithLMSI.AdditionalNetworkNodeDiameterAddress = &NetworkNodeDiameterAddress{DiameterRealm: b, DiameterName: HexBytes("node.test")}
			return v
		}, asParser(ParseSriSmResp)},
		{"SriSmResp.LocationInfoWithLMSI.ThirdNetworkNodeDiameterAddress.DiameterName", func(t *testing.T, b HexBytes) marshaler {
			v := &SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "1"}}
			v.LocationInfoWithLMSI.ThirdNetworkNodeDiameterAddress = &NetworkNodeDiameterAddress{DiameterName: b, DiameterRealm: HexBytes("node.test")}
			return v
		}, asParser(ParseSriSmResp)},
		{"SriSmResp.LocationInfoWithLMSI.ThirdNetworkNodeDiameterAddress.DiameterRealm", func(t *testing.T, b HexBytes) marshaler {
			v := &SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "1"}}
			v.LocationInfoWithLMSI.ThirdNetworkNodeDiameterAddress = &NetworkNodeDiameterAddress{DiameterRealm: b, DiameterName: HexBytes("node.test")}
			return v
		}, asParser(ParseSriSmResp)},
		{"SriSmResp.LocationInfoWithLMSI.Smsf3gppDiameterAddress.DiameterName", func(t *testing.T, b HexBytes) marshaler {
			v := &SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "1"}}
			v.LocationInfoWithLMSI.Smsf3gppDiameterAddress = &NetworkNodeDiameterAddress{DiameterName: b, DiameterRealm: HexBytes("node.test")}
			return v
		}, asParser(ParseSriSmResp)},
		{"SriSmResp.LocationInfoWithLMSI.Smsf3gppDiameterAddress.DiameterRealm", func(t *testing.T, b HexBytes) marshaler {
			v := &SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "1"}}
			v.LocationInfoWithLMSI.Smsf3gppDiameterAddress = &NetworkNodeDiameterAddress{DiameterRealm: b, DiameterName: HexBytes("node.test")}
			return v
		}, asParser(ParseSriSmResp)},
		{"SriSmResp.LocationInfoWithLMSI.SmsfNon3gppDiameterAddress.DiameterName", func(t *testing.T, b HexBytes) marshaler {
			v := &SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "1"}}
			v.LocationInfoWithLMSI.SmsfNon3gppDiameterAddress = &NetworkNodeDiameterAddress{DiameterName: b, DiameterRealm: HexBytes("node.test")}
			return v
		}, asParser(ParseSriSmResp)},
		{"SriSmResp.LocationInfoWithLMSI.SmsfNon3gppDiameterAddress.DiameterRealm", func(t *testing.T, b HexBytes) marshaler {
			v := &SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "1"}}
			v.LocationInfoWithLMSI.SmsfNon3gppDiameterAddress = &NetworkNodeDiameterAddress{DiameterRealm: b, DiameterName: HexBytes("node.test")}
			return v
		}, asParser(ParseSriSmResp)},
		{"MtFsm.SmsGmscDiameterAddress.DiameterName", func(t *testing.T, b HexBytes) marshaler {
			v := semMtFsm(t)
			v.SmsGmscDiameterAddress = &NetworkNodeDiameterAddress{DiameterName: b, DiameterRealm: HexBytes("node.test")}
			return v
		}, asParser(ParseMtFsm)},
		{"MtFsm.SmsGmscDiameterAddress.DiameterRealm", func(t *testing.T, b HexBytes) marshaler {
			v := semMtFsm(t)
			v.SmsGmscDiameterAddress = &NetworkNodeDiameterAddress{DiameterRealm: b, DiameterName: HexBytes("node.test")}
			return v
		}, asParser(ParseMtFsm)},
		{"SubscriberLocationReportArg.MmeName", func(_ *testing.T, b HexBytes) marshaler { v := semSLRArg(); v.LcsLocationInfo.MmeName = b; return v }, asParser(ParseSubscriberLocationReport)},
		{"SubscriberLocationReportArg.AaaServerName", func(_ *testing.T, b HexBytes) marshaler {
			v := semSLRArg()
			v.LcsLocationInfo.AaaServerName = b
			return v
		}, asParser(ParseSubscriberLocationReport)},
		{"SubscriberLocationReportArg.SgsnName", func(_ *testing.T, b HexBytes) marshaler { v := semSLRArg(); v.LcsLocationInfo.SgsnName = b; return v }, asParser(ParseSubscriberLocationReport)},
		{"SubscriberLocationReportArg.SgsnRealm", func(_ *testing.T, b HexBytes) marshaler { v := semSLRArg(); v.LcsLocationInfo.SgsnRealm = b; return v }, asParser(ParseSubscriberLocationReport)},
		{"SriLcsResp.MmeName", func(_ *testing.T, b HexBytes) marshaler {
			v := &SriLcsResp{TargetMS: SubscriberIdentity{IMSI: "001010123456789"}, LcsLocationInfo: LCSLocationInfo{NetworkNodeNumber: "1"}}
			v.LcsLocationInfo.MmeName = b
			return v
		}, asParser(ParseSriLcsResp)},
		{"SriLcsResp.AaaServerName", func(_ *testing.T, b HexBytes) marshaler {
			v := &SriLcsResp{TargetMS: SubscriberIdentity{IMSI: "001010123456789"}, LcsLocationInfo: LCSLocationInfo{NetworkNodeNumber: "1"}}
			v.LcsLocationInfo.AaaServerName = b
			return v
		}, asParser(ParseSriLcsResp)},
		{"SriLcsResp.SgsnName", func(_ *testing.T, b HexBytes) marshaler {
			v := &SriLcsResp{TargetMS: SubscriberIdentity{IMSI: "001010123456789"}, LcsLocationInfo: LCSLocationInfo{NetworkNodeNumber: "1"}}
			v.LcsLocationInfo.SgsnName = b
			return v
		}, asParser(ParseSriLcsResp)},
		{"SriLcsResp.SgsnRealm", func(_ *testing.T, b HexBytes) marshaler {
			v := &SriLcsResp{TargetMS: SubscriberIdentity{IMSI: "001010123456789"}, LcsLocationInfo: LCSLocationInfo{NetworkNodeNumber: "1"}}
			v.LcsLocationInfo.SgsnRealm = b
			return v
		}, asParser(ParseSriLcsResp)},
		{"ProvideSubscriberInfoRes.EPS.MmeName", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{LocationInformationEPS: &EPSLocationInformation{MmeName: b}}}
		}, asParser(ParseProvideSubscriberInfoRes)},
		{"SriResp.EPS.MmeName", func(_ *testing.T, b HexBytes) marshaler {
			return &SriResp{SubscriberInfo: &SubscriberInfo{LocationInformationEPS: &EPSLocationInformation{MmeName: b}}}
		}, asParser(ParseSriResp)},
		{"UpdateLocation.MmeDiameterAddress.name", func(_ *testing.T, b HexBytes) marshaler {
			v := senderUpdateLocation()
			v.MmeDiameterAddress = &NetworkNodeDiameterAddress{DiameterName: b, DiameterRealm: HexBytes("realm.test")}
			return v
		}, asParser(ParseUpdateLocation)},
		{"UpdateLocation.MmeDiameterAddress.realm", func(_ *testing.T, b HexBytes) marshaler {
			v := senderUpdateLocation()
			v.MmeDiameterAddress = &NetworkNodeDiameterAddress{DiameterName: HexBytes("node.test"), DiameterRealm: b}
			return v
		}, asParser(ParseUpdateLocation)},
		{"UpdateGprsLocation.SgsnName", func(_ *testing.T, b HexBytes) marshaler {
			return &UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "1", SGSNAddress: "127.0.0.1", SgsnName: b}
		}, asParser(ParseUpdateGprsLocation)},
		{"UpdateGprsLocation.SgsnRealm", func(_ *testing.T, b HexBytes) marshaler {
			return &UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "1", SGSNAddress: "127.0.0.1", SgsnRealm: b}
		}, asParser(ParseUpdateGprsLocation)},
		{"InsertSubscriberData.MmeName", func(_ *testing.T, b HexBytes) marshaler { return &InsertSubscriberDataArg{MmeName: b} }, asParser(ParseInsertSubscriberData)},
		{"ATI.EPS.MmeName", func(_ *testing.T, b HexBytes) marshaler {
			return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformationEPS: &EPSLocationInformation{MmeName: b}}}
		}, asParser(ParseAnyTimeInterrogationRes)},
		{"PSL.ServingNode.MmeNumber", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberLocationRes{LocationEstimate: ExtGeographicalInformation{0, 0, 0, 0, 0, 0, 0}, TargetServingNodeForHandover: &ServingNodeAddress{MmeNumber: b}}
		}, asParser(ParseProvideSubscriberLocationRes)},
	}
}

// TS 29.002 V19.1.0 §17.7.8 and RFC 6733 §4.3.1 require ASCII DiameterIdentity.
func TestSenderDiameterASCII(t *testing.T) {
	for _, tc := range senderDiameterCases() {
		t.Run(tc.name, func(t *testing.T) {
			for pos := 0; pos < 9; pos++ {
				for invalid := 0x80; invalid <= 0xff; invalid++ {
					b := HexBytes("node.test")
					b[pos] = byte(invalid)
					if _, err := tc.build(t, b).Marshal(); !errors.Is(err, ErrDiameterIdentityNotASCII) {
						t.Fatalf("%02x: %v", invalid, err)
					}
				}
			}
			for _, b := range []HexBytes{HexBytes("node.test"), {0, 1, 2, 3, 4, 5, 6, 0x7f, 0x20}} {
				wire, err := tc.build(t, b).Marshal()
				if err != nil {
					t.Fatal(err)
				}
				if _, err = tc.parse(wire); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func senderPLMNCases() []senderOctetCase {
	return []senderOctetCase{
		{"UpdateGprsLocation.EplmnList[1]", func(_ *testing.T, b HexBytes) marshaler {
			return &UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "1", SGSNAddress: "127.0.0.1", EplmnList: []HexBytes{{0x21, 0xf3, 0x54}, b}}
		}, asParser(ParseUpdateGprsLocation)},
		{"UpdateGprsLocation.AdjacentPLMNList[1]", func(_ *testing.T, b HexBytes) marshaler {
			return &UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "1", SGSNAddress: "127.0.0.1", AdjacentPLMNList: []HexBytes{{0x21, 0xf3, 0x54}, b}}
		}, asParser(ParseUpdateGprsLocation)},
		{"UpdateLocation.PagingArea.LAI", func(t *testing.T, b HexBytes) marshaler {
			v := senderUpdateLocation()
			lai := gsm_map.NewLocationAreaLaiFixedLength(gsm_map.LAIFixedLength(append(append(HexBytes{}, b...), 0, 0)))
			raw, err := lai.MarshalBER()
			if err != nil {
				t.Fatal(err)
			}
			v.PagingArea = []HexBytes{raw}
			return v
		}, asParser(ParseUpdateLocation)},
		{"AnyTimeInterrogationRes.LocationInformation.CellGlobalId", func(_ *testing.T, b HexBytes) marshaler {
			return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformation: &CSLocationInformation{CellGlobalId: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}
		}, asParser(ParseAnyTimeInterrogationRes)},
		{"AnyTimeInterrogationRes.LocationInformation.LAI", func(_ *testing.T, b HexBytes) marshaler {
			return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformation: &CSLocationInformation{LAI: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}}}
		}, asParser(ParseAnyTimeInterrogationRes)},
		{"AnyTimeInterrogationRes.LocationInformationGPRS.CellGlobalId", func(_ *testing.T, b HexBytes) marshaler {
			return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformationGPRS: &GPRSLocationInformation{CellGlobalId: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}
		}, asParser(ParseAnyTimeInterrogationRes)},
		{"AnyTimeInterrogationRes.LocationInformationGPRS.LAI", func(_ *testing.T, b HexBytes) marshaler {
			return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformationGPRS: &GPRSLocationInformation{LAI: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}}}
		}, asParser(ParseAnyTimeInterrogationRes)},
		{"AnyTimeInterrogationRes.LocationInformationGPRS.RouteingAreaIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformationGPRS: &GPRSLocationInformation{RouteingAreaIdentity: append(append(HexBytes{}, b...), make(HexBytes, 3)...)}}}
		}, asParser(ParseAnyTimeInterrogationRes)},
		{"AnyTimeInterrogationRes.LocationInformationEPS.EUtranCellGlobalIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformationEPS: &EPSLocationInformation{EUtranCellGlobalIdentity: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}
		}, asParser(ParseAnyTimeInterrogationRes)},
		{"AnyTimeInterrogationRes.LocationInformationEPS.TrackingAreaIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformationEPS: &EPSLocationInformation{TrackingAreaIdentity: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}}}
		}, asParser(ParseAnyTimeInterrogationRes)},
		{"AnyTimeInterrogationRes.LocationInformation5GS.NrCellGlobalIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{NrCellGlobalIdentity: append(append(HexBytes{}, b...), make(HexBytes, 5)...)}}}
		}, asParser(ParseAnyTimeInterrogationRes)},
		{"AnyTimeInterrogationRes.LocationInformation5GS.EUtranCellGlobalIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{EUtranCellGlobalIdentity: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}
		}, asParser(ParseAnyTimeInterrogationRes)},
		{"AnyTimeInterrogationRes.LocationInformation5GS.TrackingAreaIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{TrackingAreaIdentity: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}}}
		}, asParser(ParseAnyTimeInterrogationRes)},
		{"AnyTimeInterrogationRes.LocationInformation5GS.VplmnID", func(_ *testing.T, b HexBytes) marshaler {
			return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{VplmnID: append(append(HexBytes{}, b...), make(HexBytes, 0)...)}}}
		}, asParser(ParseAnyTimeInterrogationRes)},
		{"AnyTimeInterrogationRes.LocationInformation5GS.NrTrackingAreaIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{NrTrackingAreaIdentity: append(append(HexBytes{}, b...), make(HexBytes, 3)...)}}}
		}, asParser(ParseAnyTimeInterrogationRes)},
		{"ProvideSubscriberInfoRes.LocationInformation.CellGlobalId", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{LocationInformation: &CSLocationInformation{CellGlobalId: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}
		}, asParser(ParseProvideSubscriberInfoRes)},
		{"ProvideSubscriberInfoRes.LocationInformation.LAI", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{LocationInformation: &CSLocationInformation{LAI: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}}}
		}, asParser(ParseProvideSubscriberInfoRes)},
		{"ProvideSubscriberInfoRes.LocationInformationGPRS.CellGlobalId", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{LocationInformationGPRS: &GPRSLocationInformation{CellGlobalId: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}
		}, asParser(ParseProvideSubscriberInfoRes)},
		{"ProvideSubscriberInfoRes.LocationInformationGPRS.LAI", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{LocationInformationGPRS: &GPRSLocationInformation{LAI: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}}}
		}, asParser(ParseProvideSubscriberInfoRes)},
		{"ProvideSubscriberInfoRes.LocationInformationGPRS.RouteingAreaIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{LocationInformationGPRS: &GPRSLocationInformation{RouteingAreaIdentity: append(append(HexBytes{}, b...), make(HexBytes, 3)...)}}}
		}, asParser(ParseProvideSubscriberInfoRes)},
		{"ProvideSubscriberInfoRes.LocationInformationEPS.EUtranCellGlobalIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{LocationInformationEPS: &EPSLocationInformation{EUtranCellGlobalIdentity: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}
		}, asParser(ParseProvideSubscriberInfoRes)},
		{"ProvideSubscriberInfoRes.LocationInformationEPS.TrackingAreaIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{LocationInformationEPS: &EPSLocationInformation{TrackingAreaIdentity: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}}}
		}, asParser(ParseProvideSubscriberInfoRes)},
		{"ProvideSubscriberInfoRes.LocationInformation5GS.NrCellGlobalIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{NrCellGlobalIdentity: append(append(HexBytes{}, b...), make(HexBytes, 5)...)}}}
		}, asParser(ParseProvideSubscriberInfoRes)},
		{"ProvideSubscriberInfoRes.LocationInformation5GS.EUtranCellGlobalIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{EUtranCellGlobalIdentity: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}
		}, asParser(ParseProvideSubscriberInfoRes)},
		{"ProvideSubscriberInfoRes.LocationInformation5GS.TrackingAreaIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{TrackingAreaIdentity: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}}}
		}, asParser(ParseProvideSubscriberInfoRes)},
		{"ProvideSubscriberInfoRes.LocationInformation5GS.VplmnID", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{VplmnID: append(append(HexBytes{}, b...), make(HexBytes, 0)...)}}}
		}, asParser(ParseProvideSubscriberInfoRes)},
		{"ProvideSubscriberInfoRes.LocationInformation5GS.NrTrackingAreaIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{NrTrackingAreaIdentity: append(append(HexBytes{}, b...), make(HexBytes, 3)...)}}}
		}, asParser(ParseProvideSubscriberInfoRes)},
		{"SriResp.LocationInformation.CellGlobalId", func(_ *testing.T, b HexBytes) marshaler {
			return &SriResp{SubscriberInfo: &SubscriberInfo{LocationInformation: &CSLocationInformation{CellGlobalId: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}
		}, asParser(ParseSriResp)},
		{"SriResp.LocationInformation.LAI", func(_ *testing.T, b HexBytes) marshaler {
			return &SriResp{SubscriberInfo: &SubscriberInfo{LocationInformation: &CSLocationInformation{LAI: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}}}
		}, asParser(ParseSriResp)},
		{"SriResp.LocationInformationGPRS.CellGlobalId", func(_ *testing.T, b HexBytes) marshaler {
			return &SriResp{SubscriberInfo: &SubscriberInfo{LocationInformationGPRS: &GPRSLocationInformation{CellGlobalId: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}
		}, asParser(ParseSriResp)},
		{"SriResp.LocationInformationGPRS.LAI", func(_ *testing.T, b HexBytes) marshaler {
			return &SriResp{SubscriberInfo: &SubscriberInfo{LocationInformationGPRS: &GPRSLocationInformation{LAI: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}}}
		}, asParser(ParseSriResp)},
		{"SriResp.LocationInformationGPRS.RouteingAreaIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &SriResp{SubscriberInfo: &SubscriberInfo{LocationInformationGPRS: &GPRSLocationInformation{RouteingAreaIdentity: append(append(HexBytes{}, b...), make(HexBytes, 3)...)}}}
		}, asParser(ParseSriResp)},
		{"SriResp.LocationInformationEPS.EUtranCellGlobalIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &SriResp{SubscriberInfo: &SubscriberInfo{LocationInformationEPS: &EPSLocationInformation{EUtranCellGlobalIdentity: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}
		}, asParser(ParseSriResp)},
		{"SriResp.LocationInformationEPS.TrackingAreaIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &SriResp{SubscriberInfo: &SubscriberInfo{LocationInformationEPS: &EPSLocationInformation{TrackingAreaIdentity: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}}}
		}, asParser(ParseSriResp)},
		{"SriResp.LocationInformation5GS.NrCellGlobalIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &SriResp{SubscriberInfo: &SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{NrCellGlobalIdentity: append(append(HexBytes{}, b...), make(HexBytes, 5)...)}}}
		}, asParser(ParseSriResp)},
		{"SriResp.LocationInformation5GS.EUtranCellGlobalIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &SriResp{SubscriberInfo: &SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{EUtranCellGlobalIdentity: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}
		}, asParser(ParseSriResp)},
		{"SriResp.LocationInformation5GS.TrackingAreaIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &SriResp{SubscriberInfo: &SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{TrackingAreaIdentity: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}}}
		}, asParser(ParseSriResp)},
		{"SriResp.LocationInformation5GS.VplmnID", func(_ *testing.T, b HexBytes) marshaler {
			return &SriResp{SubscriberInfo: &SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{VplmnID: append(append(HexBytes{}, b...), make(HexBytes, 0)...)}}}
		}, asParser(ParseSriResp)},
		{"SriResp.LocationInformation5GS.NrTrackingAreaIdentity", func(_ *testing.T, b HexBytes) marshaler {
			return &SriResp{SubscriberInfo: &SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{NrTrackingAreaIdentity: append(append(HexBytes{}, b...), make(HexBytes, 3)...)}}}
		}, asParser(ParseSriResp)},
		{"PSLRes.CellGlobalId", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberLocationRes{LocationEstimate: ExtGeographicalInformation{0, 0, 0, 0, 0, 0, 0}, CellGlobalId: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}
		}, asParser(ParseProvideSubscriberLocationRes)},
		{"SLRArg.CellGlobalId", func(_ *testing.T, b HexBytes) marshaler {
			v := semSLRArg()
			v.IMSI = "001010123456789"
			v.CellGlobalId = append(append(HexBytes{}, b...), make(HexBytes, 4)...)
			return v
		}, asParser(ParseSubscriberLocationReport)},
		{"PSLRes.LAI", func(_ *testing.T, b HexBytes) marshaler {
			return &ProvideSubscriberLocationRes{LocationEstimate: ExtGeographicalInformation{0, 0, 0, 0, 0, 0, 0}, LAI: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}
		}, asParser(ParseProvideSubscriberLocationRes)},
		{"SLRArg.LAI", func(_ *testing.T, b HexBytes) marshaler {
			v := semSLRArg()
			v.IMSI = "001010123456789"
			v.LAI = append(append(HexBytes{}, b...), make(HexBytes, 2)...)
			return v
		}, asParser(ParseSubscriberLocationReport)},
		{"PSLArg.ReportingPLMNList[1]", func(_ *testing.T, b HexBytes) marshaler {
			v := semPSLArg()
			v.IMSI = "001010123456789"
			v.ReportingPLMNList = &ReportingPLMNList{PlmnList: PLMNList{{PlmnId: HexBytes{0x21, 0xf3, 0x54}}, {PlmnId: b}}}
			return v
		}, asParser(ParseProvideSubscriberLocation)},
		{"PSLArg.Area.AreaTypePlmnId", func(_ *testing.T, b HexBytes) marshaler {
			v := semPSLArg()
			v.IMSI = "001010123456789"
			v.AreaEventInfo = &AreaEventInfo{AreaDefinition: AreaDefinition{AreaList: AreaList{{AreaType: AreaTypePlmnId, AreaIdentification: append(append(HexBytes{}, b...), make(HexBytes, 0)...)}}}}
			return v
		}, asParser(ParseProvideSubscriberLocation)},
		{"PSLArg.Area.AreaTypeLocationAreaId", func(_ *testing.T, b HexBytes) marshaler {
			v := semPSLArg()
			v.IMSI = "001010123456789"
			v.AreaEventInfo = &AreaEventInfo{AreaDefinition: AreaDefinition{AreaList: AreaList{{AreaType: AreaTypeLocationAreaId, AreaIdentification: append(append(HexBytes{}, b...), make(HexBytes, 2)...)}}}}
			return v
		}, asParser(ParseProvideSubscriberLocation)},
		{"PSLArg.Area.AreaTypeRoutingAreaId", func(_ *testing.T, b HexBytes) marshaler {
			v := semPSLArg()
			v.IMSI = "001010123456789"
			v.AreaEventInfo = &AreaEventInfo{AreaDefinition: AreaDefinition{AreaList: AreaList{{AreaType: AreaTypeRoutingAreaId, AreaIdentification: append(append(HexBytes{}, b...), make(HexBytes, 3)...)}}}}
			return v
		}, asParser(ParseProvideSubscriberLocation)},
		{"PSLArg.Area.AreaTypeCellGlobalId", func(_ *testing.T, b HexBytes) marshaler {
			v := semPSLArg()
			v.IMSI = "001010123456789"
			v.AreaEventInfo = &AreaEventInfo{AreaDefinition: AreaDefinition{AreaList: AreaList{{AreaType: AreaTypeCellGlobalId, AreaIdentification: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}}
			return v
		}, asParser(ParseProvideSubscriberLocation)},
		{"PSLArg.Area.AreaTypeUtranCellId", func(_ *testing.T, b HexBytes) marshaler {
			v := semPSLArg()
			v.IMSI = "001010123456789"
			v.AreaEventInfo = &AreaEventInfo{AreaDefinition: AreaDefinition{AreaList: AreaList{{AreaType: AreaTypeUtranCellId, AreaIdentification: append(append(HexBytes{}, b...), make(HexBytes, 4)...)}}}}
			return v
		}, asParser(ParseProvideSubscriberLocation)},
		{"UpdateLocation.EplmnList[1]", func(_ *testing.T, b HexBytes) marshaler {
			v := senderUpdateLocation()
			v.EplmnList = []HexBytes{{0x21, 0xf3, 0x54}, b}
			return v
		}, asParser(ParseUpdateLocation)},
		{"SAI.RequestingPLMNId", func(_ *testing.T, b HexBytes) marshaler {
			return &SendAuthenticationInfo{IMSI: "001010123456789", NumberOfRequestedVectors: 1, RequestingPLMNId: b}
		}, asParser(ParseSendAuthenticationInfo)},
		{"ISD.AdjacentAccessRestrictionDataList", func(_ *testing.T, b HexBytes) marshaler {
			return &InsertSubscriberDataArg{AdjacentAccessRestrictionDataList: AdjacentAccessRestrictionDataList{{PlmnId: b}}}
		}, asParser(ParseInsertSubscriberData)},
		{"ISD.ImsiGroupIdList", func(_ *testing.T, b HexBytes) marshaler {
			return &InsertSubscriberDataArg{ImsiGroupIdList: IMSIGroupIdList{{PlmnId: b, LocalGroupID: HexBytes{1}}}}
		}, asParser(ParseInsertSubscriberData)},
		{"ISD.CsgSubscriptionDataList", func(_ *testing.T, b HexBytes) marshaler {
			return &InsertSubscriberDataArg{CsgSubscriptionDataList: CSGSubscriptionDataList{{CsgID: HexBytes{0, 0, 0, 0}, CsgIDBits: 27, PlmnId: b}}}
		}, asParser(ParseInsertSubscriberData)},
		{"ISD.VplmnCsgSubscriptionDataList", func(_ *testing.T, b HexBytes) marshaler {
			return &InsertSubscriberDataArg{VplmnCsgSubscriptionDataList: VPLMNCSGSubscriptionDataList{{CsgID: HexBytes{0, 0, 0, 0}, CsgIDBits: 27, PlmnId: b}}}
		}, asParser(ParseInsertSubscriberData)},
	}
}

// TS 29.002 V19.1.0 §17.7.8: PLMN, CGI/SAI and LAI use decimal MCC/MNC digits.
func TestSenderPLMNDigits(t *testing.T) {
	for _, tc := range senderPLMNCases() {
		t.Run(tc.name, func(t *testing.T) {
			for pos := 0; pos < 6; pos++ {
				for d := byte(0); d < 16; d++ {
					b := HexBytes{0x21, 0x63, 0x54}
					shift := uint((pos % 2) * 4)
					b[pos/2] = (b[pos/2] &^ (0xf << shift)) | d<<shift
					wire, err := tc.build(t, b).Marshal()
					if d < 10 || pos == 3 && d == 15 {
						if err != nil {
							t.Fatal(err)
						}
						if _, err = tc.parse(wire); err != nil {
							t.Fatal(err)
						}
						continue
					}
					if !errors.Is(err, ErrPLMNIdentityInvalidDigits) {
						t.Fatalf("digit%d=%x: %v", pos, d, err)
					}
				}
			}
			for _, b := range []HexBytes{{0x21, 0xf3, 0x54}, {0x21, 0x63, 0x54}} {
				wire, err := tc.build(t, b).Marshal()
				if err != nil {
					t.Fatal(err)
				}
				if _, err = tc.parse(wire); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.13 AreaIdentification codes MCC and any included MNC.
func TestSenderAreaCountryMCC(t *testing.T) {
	for unused := byte(0); unused < 16; unused++ {
		b := HexBytes{0x21, unused<<4 | 3}
		v := semPSLArg()
		v.IMSI = "001010123456789"
		v.AreaEventInfo = &AreaEventInfo{AreaDefinition: AreaDefinition{AreaList: AreaList{{AreaType: AreaTypeCountryCode, AreaIdentification: b}}}}
		wire, err := v.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		p, err := ParseProvideSubscriberLocation(wire)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(p.AreaEventInfo.AreaDefinition.AreaList[0].AreaIdentification, b) {
			t.Fatal("country MCC changed")
		}
	}
	for pos := 0; pos < 3; pos++ {
		for digit := byte(10); digit < 16; digit++ {
			b := HexBytes{0x21, 0xf3}
			shift := uint((pos % 2) * 4)
			b[pos/2] = b[pos/2]&^(0xf<<shift) | digit<<shift
			v := semPSLArg()
			v.IMSI = "001010123456789"
			v.AreaEventInfo = &AreaEventInfo{AreaDefinition: AreaDefinition{AreaList: AreaList{{AreaType: AreaTypeCountryCode, AreaIdentification: b}}}}
			if _, err := v.Marshal(); !errors.Is(err, ErrPLMNIdentityInvalidDigits) {
				t.Fatalf("%x: %v", b, err)
			}
		}
	}
}

func senderCommonLenientCases() []semLenientCase {
	update := func() *gsm_map.UpdateLocationArg {
		return &gsm_map.UpdateLocationArg{Imsi: semAddrIMSI, MscNumber: semOneDigit, VlrNumber: semOneDigit}
	}
	cases := []semLenientCase{
		{ErrDiameterIdentityNotASCII, asParser(ParseUpdateLocation), func(*testing.T) semWire {
			w := update()
			w.MmeDiameterAddress = &gsm_map.NetworkNodeDiameterAddress{DiameterName: gsm_map.DiameterIdentity{0x80, 'o', 'd', 'e', '.', 't', 'e', 's', 't'}, DiameterRealm: gsm_map.DiameterIdentity("realm.test")}
			return w
		}},
		{ErrPLMNIdentityInvalidDigits, asParser(ParseUpdateLocation), func(*testing.T) semWire {
			w := update()
			w.EplmnList = &gsm_map.EPLMNList{Values: []gsm_map.PLMNId{{0xaa, 0xff, 0xaa}}}
			return w
		}},
	}
	for _, tc := range []struct {
		header byte
		want   error
	}{{0xd1, ErrAddressNatureInvalid}, {0x92, ErrAddressPlanInvalid}} {
		cases = append(cases, semLenientCase{tc.want, asParser(ParseUpdateLocation), func(*testing.T) semWire { w := update(); w.MscNumber = []byte{tc.header, 0x21}; return w }})
	}
	return cases
}

// TS 29.002 V19.1.0 §17.7.13 AreaIdentification assigns octet 3 to MNC digits.
func TestSenderAreaCountryIncludedMNC(t *testing.T) {
	for length := 3; length <= 7; length++ {
		for _, prefix := range []HexBytes{{0x21, 0xf3, 0x54}, {0x21, 0x63, 0x54}, {0x21, 0xa3, 0x54}, {0x21, 0xf3, 0xfa}, {0x21, 0xf3, 0xf4}} {
			raw := make(HexBytes, length)
			copy(raw, prefix)
			in := senderPSLArg()
			in.AreaEventInfo = &AreaEventInfo{AreaDefinition: AreaDefinition{AreaList: AreaList{{AreaType: AreaTypeCountryCode, AreaIdentification: raw}}}}
			data, err := in.Marshal()
			if prefix[1] == 0xa3 || prefix[2] == 0xfa || prefix[2] == 0xf4 {
				if !errors.Is(err, ErrPLMNIdentityInvalidDigits) {
					t.Fatalf("%x: got %v", raw, err)
				}
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			got, err := ParseProvideSubscriberLocation(data)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got.AreaEventInfo.AreaDefinition.AreaList[0].AreaIdentification, raw) {
				t.Fatalf("country area changed: %x", raw)
			}
		}
	}
}

func TestSenderPLMNListErrorContexts(t *testing.T) {
	for _, field := range []string{"EplmnList", "AdjacentPLMNList"} {
		t.Run(field, func(t *testing.T) {
			in := senderMobilityGPRS()
			list := []HexBytes{{0x21, 0xf3, 0x54}, {0x21, 0xf3, 0xfa}}
			if field == "EplmnList" {
				in.EplmnList = list
			} else {
				in.AdjacentPLMNList = list
			}
			_, err := in.Marshal()
			if !errors.Is(err, ErrPLMNIdentityInvalidDigits) || !strings.Contains(err.Error(), "UpdateGprsLocation."+field+"[1]") {
				t.Fatalf("wrong field context: %v", err)
			}
		})
	}
	in := senderUpdateLocation()
	in.EplmnList = []HexBytes{{0x21, 0xf3, 0x54}, {0x21, 0xf3, 0xfa}}
	_, err := in.Marshal()
	if !errors.Is(err, ErrPLMNIdentityInvalidDigits) || !strings.Contains(err.Error(), "UpdateLocation.EplmnList[1]") {
		t.Fatalf("wrong field context: %v", err)
	}
}
