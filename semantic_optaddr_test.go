// semantic_optaddr_test.go
//
// A present OPTIONAL AddressString or ISDN-AddressString (3GPP TS 29.002
// V19.1.0 §17.7.8) may hold only its nature/plan octet, or only filler. The
// public field holds the digits with "" meaning absent, so such an address
// would parse to "absent" and Marshal would drop it. Parse rejects it with
// the field's "...DecodedEmpty" sentinel; one digit is the shortest address
// that parses, and it marshals back to the same octets.
package gsmmap

import (
	"bytes"
	"errors"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func semOptionalAddressCases() []semAddressCase {
	ptr := func(a []byte) *gsm_map.ISDNAddressString {
		v := a
		return &v
	}
	imsi := func() *gsm_map.IMSI {
		v := semAddrIMSI
		return &v
	}
	return []semAddressCase{
		{"Reset SendingNodeNumber HLR", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.ResetArg{SendingNodenumber: gsm_map.NewSendingNodeNumberHlrNumber(a)}
		}, semParse(ParseResetArg), ErrSendingNodeNumberHLRNumberDecodedEmpty},
		{"Reset SendingNodeNumber CSS", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.ResetArg{SendingNodenumber: gsm_map.NewSendingNodeNumberCssNumber(a)}
		}, semParse(ParseResetArg), ErrSendingNodeNumberCSSNumberDecodedEmpty},

		// TS 29.002 V19.1.0 §17.7.6 uses an ISDN address in each list CHOICE.
		{"ReportSMDeliveryStatus failedSMServingNodes networkNode-Number", func(t *testing.T, a []byte) semWire {
			w := semMust(convertReportSMDeliveryStatusToArg(minimalReportSMDeliveryStatus()))
			w.FailedSMServingNodes = &gsm_map.SMServingNodeAddressList{Values: []gsm_map.SMServingNodeAddress{gsm_map.NewSMServingNodeAddressNetworkNodeNumber(a)}}
			return w
		}, semParse(ParseReportSMDeliveryStatus), ErrSMServingNodeAddressNetworkNodeNumberDecodedEmpty},
		{"ReportSMDeliveryStatusRes registeredSMServingNodes networkNode-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.ReportSMDeliveryStatusRes{RegisteredSMServingNodes: &gsm_map.SMServingNodeAddressList{Values: []gsm_map.SMServingNodeAddress{gsm_map.NewSMServingNodeAddressNetworkNodeNumber(a)}}}
		}, semParse(ParseReportSMDeliveryStatusRes), ErrSMServingNodeAddressNetworkNodeNumberDecodedEmpty},
		{"ReportSMDeliveryStatusRes storedMSISDN", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.ReportSMDeliveryStatusRes{StoredMSISDN: ptr(a)}
		}, semParse(ParseReportSMDeliveryStatusRes), ErrReportSMDeliveryStatusResStoredMSISDNDecodedEmpty},

		{"AlertServiceCentre newSGSN-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.AlertServiceCentreArg{Msisdn: semTwoDigits, ServiceCentreAddress: semTwoDigits, NewSGSNNumber: ptr(a)}
		}, semParse(ParseAlertServiceCentre), ErrAscNewSGSNNumberDecodedEmpty},
		{"AlertServiceCentre newMME-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.AlertServiceCentreArg{Msisdn: semTwoDigits, ServiceCentreAddress: semTwoDigits, NewMMENumber: ptr(a)}
		}, semParse(ParseAlertServiceCentre), ErrAscNewMMENumberDecodedEmpty},
		{"AlertServiceCentre newMSC-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.AlertServiceCentreArg{Msisdn: semTwoDigits, ServiceCentreAddress: semTwoDigits, NewMSCNumber: ptr(a)}
		}, semParse(ParseAlertServiceCentre), ErrAscNewMSCNumberDecodedEmpty},
		{"CancelLocation newMSC-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.CancelLocationArg{Identity: gsm_map.NewIdentityImsi(semAddrIMSI), NewMSCNumber: ptr(a)}
		}, semParse(ParseCancelLocation), ErrCancelLocNewMSCNumberDecodedEmpty},
		{"CancelLocation newVLR-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.CancelLocationArg{Identity: gsm_map.NewIdentityImsi(semAddrIMSI), NewVLRNumber: ptr(a)}
		}, semParse(ParseCancelLocation), ErrCancelLocNewVLRNumberDecodedEmpty},
		{"PurgeMS vlr-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.PurgeMSArg{Imsi: semAddrIMSI, VlrNumber: ptr(a)}
		}, semParse(ParsePurgeMS), ErrPurgeMSVlrNumberDecodedEmpty},
		{"PurgeMS sgsn-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.PurgeMSArg{Imsi: semAddrIMSI, SgsnNumber: ptr(a)}
		}, semParse(ParsePurgeMS), ErrPurgeMSSgsnNumberDecodedEmpty},
		{"SendRoutingInfoRes vmsc-Address", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.SendRoutingInfoRes{Imsi: imsi(), VmscAddress: ptr(a)}
		}, semParse(ParseSriRes), ErrSriResVmscAddressDecodedEmpty},
		{"SendRoutingInfoRes msisdn", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.SendRoutingInfoRes{Imsi: imsi(), Msisdn: ptr(a)}
		}, semParse(ParseSriRes), ErrSriResMSISDNDecodedEmpty},
		{"SendRoutingInfoRes forwardingData forwardedToNumber", func(_ *testing.T, a []byte) semWire {
			eri := gsm_map.NewExtendedRoutingInfoRoutingInfo(gsm_map.NewRoutingInfoForwardingData(gsm_map.ForwardingData{ForwardedToNumber: ptr(a)}))
			return &gsm_map.SendRoutingInfoRes{Imsi: imsi(), ExtendedRoutingInfo: &eri}
		}, semParse(ParseSriRes), ErrForwardingDataForwardedToNumberDecodedEmpty},
		{"RoutingInfoForSMRes smsf-3gpp-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.RoutingInfoForSMRes{Imsi: semAddrIMSI, LocationInfoWithLMSI: gsm_map.LocationInfoWithLMSI{NetworkNodeNumber: semTwoDigits, Smsf3gppNumber: ptr(a)}}
		}, semParse(ParseSriSmRes), ErrSriSmResSmsf3gppNumberDecodedEmpty},
		{"RoutingInfoForSMRes smsf-non-3gpp-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.RoutingInfoForSMRes{Imsi: semAddrIMSI, LocationInfoWithLMSI: gsm_map.LocationInfoWithLMSI{NetworkNodeNumber: semTwoDigits, SmsfNon3gppNumber: ptr(a)}}
		}, semParse(ParseSriSmRes), ErrSriSmResSmsfNon3gppNumberDecodedEmpty},
		{"AnyTimeInterrogationRes locationInformation vlr-number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{LocationInformation: &gsm_map.LocationInformation{VlrNumber: ptr(a)}}}
		}, semParse(ParseAnyTimeInterrogationRes), ErrLocationInformationVlrNumberDecodedEmpty},
		{"AnyTimeInterrogationRes locationInformation msc-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{LocationInformation: &gsm_map.LocationInformation{MscNumber: ptr(a)}}}
		}, semParse(ParseAnyTimeInterrogationRes), ErrLocationInformationMscNumberDecodedEmpty},
		{"AnyTimeInterrogationRes locationInformationGPRS sgsn-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{LocationInformationGPRS: &gsm_map.LocationInformationGPRS{SgsnNumber: ptr(a)}}}
		}, semParse(ParseAnyTimeInterrogationRes), ErrLocationInformationGPRSSgsnNumberDecodedEmpty},
		{"AnyTimeInterrogationRes mnpInfoRes msisdn", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{MnpInfoRes: &gsm_map.MNPInfoRes{Msisdn: ptr(a)}}}
		}, semParse(ParseAnyTimeInterrogationRes), ErrMnpInfoResMSISDNDecodedEmpty},
		{"UpdateGprsLocation mmeNumberforMTSMS", func(t *testing.T, a []byte) semWire {
			w, err := convertUpdateGprsLocationToArg(&UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "31612345678", SGSNAddress: "192.168.31.1"})
			if err != nil {
				t.Fatalf("convertUpdateGprsLocationToArg: %v", err)
			}
			w.MmeNumberforMTSMS = ptr(a)
			return w
		}, semParse(ParseUpdateGprsLocation), ErrUpdateGprsLocationMmeNumberForMTSMSDecodedEmpty},
		{"MtFsm smsGmscAddress", func(t *testing.T, a []byte) semWire {
			w, err := convertMtFsmToArg(semMtFsm(t))
			if err != nil {
				t.Fatalf("convertMtFsmToArg: %v", err)
			}
			w.SmsGmscAddress = ptr(a)
			return w
		}, semParse(ParseMtFsm), ErrMtFsmSmsGmscAddressDecodedEmpty},
		{"InsertSubscriberData msisdn", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.InsertSubscriberDataArg{Msisdn: ptr(a)}
		}, semParse(ParseInsertSubscriberDataArg), ErrIsdMSISDNDecodedEmpty},
		{"SubscriberLocationReport msisdn", func(t *testing.T, a []byte) semWire {
			w, err := convertSubscriberLocationReportArgToWire(semSLRArg())
			if err != nil {
				t.Fatalf("convertSubscriberLocationReportArgToWire: %v", err)
			}
			w.Msisdn = ptr(a)
			return w
		}, semParse(ParseSubscriberLocationReportArg), ErrSLRArgMSISDNDecodedEmpty},
		{"SubscriberLocationReport na-ESRD", func(t *testing.T, a []byte) semWire {
			w, err := convertSubscriberLocationReportArgToWire(semSLRArg())
			if err != nil {
				t.Fatalf("convertSubscriberLocationReportArgToWire: %v", err)
			}
			w.NaESRD = ptr(a)
			return w
		}, semParse(ParseSubscriberLocationReportArg), ErrSLRArgNaESRDDecodedEmpty},
		{"SubscriberLocationReport na-ESRK", func(t *testing.T, a []byte) semWire {
			w, err := convertSubscriberLocationReportArgToWire(semSLRArg())
			if err != nil {
				t.Fatalf("convertSubscriberLocationReportArgToWire: %v", err)
			}
			w.NaESRK = ptr(a)
			return w
		}, semParse(ParseSubscriberLocationReportArg), ErrSLRArgNaESRKDecodedEmpty},
		{"InsertSubscriberData sgsn-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.InsertSubscriberDataArg{SgsnNumber: ptr(a)}
		}, semParse(ParseInsertSubscriberDataArg), ErrIsdSgsnNumberDecodedEmpty},
		{"InsertSubscriberData additionalMSISDN", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.InsertSubscriberDataArg{AdditionalMSISDN: ptr(a)}
		}, semParse(ParseInsertSubscriberDataArg), ErrIsdAdditionalMSISDNDecodedEmpty},
		{"InsertSubscriberData eps-SubscriptionData stn-sr", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.InsertSubscriberDataArg{EpsSubscriptionData: &gsm_map.EPSSubscriptionData{StnSr: ptr(a)}}
		}, semParse(ParseInsertSubscriberDataArg), ErrEPSSubscriptionDataStnSrDecodedEmpty},
		{"InsertSubscriberData Ext-ForwFeature forwardedToNumber", func(t *testing.T, a []byte) semWire {
			w := semForwISD(t)
			w.ProvisionedSS.Values[0].ForwardingInfo.ForwardingFeatureList.Values[0].ForwardedToNumber = ptr(a)
			return w
		}, semParse(ParseInsertSubscriberDataArg), ErrExtForwFeatureForwardedToNumberDecodedEmpty},
		{"InsertSubscriberData Ext-ForwFeature longForwardedToNumber", func(t *testing.T, a []byte) semWire {
			w := semForwISD(t)
			v := a
			w.ProvisionedSS.Values[0].ForwardingInfo.ForwardingFeatureList.Values[0].LongForwardedToNumber = &v
			return w
		}, semParse(ParseInsertSubscriberDataArg), ErrExtForwFeatureLongForwardedToNumberDecodedEmpty},
		{"SendRoutingInfoRes forwardingData longForwardedToNumber", func(_ *testing.T, a []byte) semWire {
			v := a
			eri := gsm_map.NewExtendedRoutingInfoRoutingInfo(gsm_map.NewRoutingInfoForwardingData(gsm_map.ForwardingData{LongForwardedToNumber: &v}))
			return &gsm_map.SendRoutingInfoRes{ExtendedRoutingInfo: &eri}
		}, semParse(ParseSriRes), ErrForwardingDataLongForwardedToNumberDecodedEmpty},
		{"ProvideSubscriberLocation lcsClientExternalID externalAddress", func(t *testing.T, a []byte) semWire {
			arg := semPSLArg()
			arg.LcsClientID = &LCSClientID{LcsClientType: LCSClientTypeEmergencyServices, LcsClientExternalID: &LCSClientExternalID{}}
			w, err := convertProvideSubscriberLocationArgToWire(arg)
			if err != nil {
				t.Fatalf("convertProvideSubscriberLocationArgToWire: %v", err)
			}
			w.LcsClientID.LcsClientExternalID.ExternalAddress = ptr(a)
			return w
		}, semParse(ParseProvideSubscriberLocationArg), ErrLCSClientExternalIDExternalAddressDecodedEmpty},
	}
}

// semForwISD returns an InsertSubscriberData wire message with one
// forwarding feature and no forwarded-to number.
func semForwISD(t *testing.T) *gsm_map.InsertSubscriberDataArg {
	t.Helper()
	w, err := convertInsertSubscriberDataArgToWire(&InsertSubscriberDataArg{ProvisionedSS: []ExtSSInfo{{
		ForwardingInfo: &ExtForwInfo{SsCode: 0x21, ForwardingFeatureList: []ExtForwFeature{{SsStatus: HexBytes{0x05}}}},
	}}})
	if err != nil {
		t.Fatalf("convertInsertSubscriberDataArgToWire: %v", err)
	}
	return w
}

func TestParseDigitlessOptionalAddress(t *testing.T) {
	for _, c := range semOptionalAddressCases() {
		for name, addr := range map[string][]byte{"no digits": semNoDigits, "filler only": {0x91, 0xFF}} {
			t.Run(c.name+"/"+name, func(t *testing.T) {
				data, err := c.build(t, addr).MarshalBER()
				if err != nil {
					t.Fatalf("MarshalBER: %v", err)
				}
				if _, err := c.parse(data); !errors.Is(err, c.want) {
					t.Errorf("Parse: err = %v, want %v", err, c.want)
				}
			})
		}
		t.Run(c.name+"/one digit", func(t *testing.T) {
			data, err := c.build(t, semOneDigit).MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}
			v, err := c.parse(data)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			out, err := v.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if !bytes.Equal(out, data) {
				t.Errorf("Marshal = %x, want %x", out, data)
			}
		})
	}
}
