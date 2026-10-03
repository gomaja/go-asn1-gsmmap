// semantic_address_test.go
//
// A mandatory AddressString or ISDN-AddressString (3GPP TS 29.002 V19.1.0
// §17.7.8, SIZE (1..maxAddressLength)) may arrive with only its
// nature/plan octet. The public types hold the digits as a string, where ""
// means the address is missing, and Marshal rejects a missing mandatory
// address; Parse rejects the digitless one with the same sentinel, so a
// parsed message always marshals again. A CHOICE alternative that decodes
// to no digits would select nothing and is rejected with its own
// "...DecodedEmpty" sentinel.
package gsmmap

import (
	"bytes"
	"errors"
	"testing"

	"github.com/gomaja/go-asn1/runtime/ber"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

var (
	semNoDigits  = []byte{0x91}                                                 // international, ISDN, no digits
	semOneDigit  = []byte{0x91, 0xF1}                                           // "1"
	semAddrIMSI  = gsm_map.IMSI{0x00, 0x01, 0x10, 0x32, 0x54, 0x76, 0x98, 0xF0} // "001001234567890"
	semTwoDigits = []byte{0x91, 0x21}                                           // "12"
)

type semWire interface {
	MarshalBER(...ber.EncodeOption) ([]byte, error)
}

type semAddressCase struct {
	name string
	// build returns the wire message with addr in the field under test.
	build func(t *testing.T, addr []byte) semWire
	// parse parses data.
	parse func(data []byte) (semMarshaler, error)
	want  error
}

type semMarshaler interface{ Marshal() ([]byte, error) }

func semParse[T semMarshaler](parse func([]byte) (T, error)) func([]byte) (semMarshaler, error) {
	return func(data []byte) (semMarshaler, error) {
		v, err := parse(data)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func semAddressCases() []semAddressCase {
	addl := func(msc bool, addr []byte) *gsm_map.AdditionalNumber {
		v := gsm_map.NewAdditionalNumberSgsnNumber(addr)
		if msc {
			v = gsm_map.NewAdditionalNumberMscNumber(addr)
		}
		return &v
	}
	sriSmRes := func(nnn []byte, an, third *gsm_map.AdditionalNumber) semWire {
		return &gsm_map.RoutingInfoForSMRes{Imsi: semAddrIMSI, LocationInfoWithLMSI: gsm_map.LocationInfoWithLMSI{
			NetworkNodeNumber: nnn, AdditionalNumber: an, ThirdNumber: third,
		}}
	}
	slr := func(t *testing.T, an *gsm_map.AdditionalNumber) semWire {
		w, err := convertSubscriberLocationReportArgToWire(semSLRArg())
		if err != nil {
			t.Fatalf("convertSubscriberLocationReportArgToWire: %v", err)
		}
		w.LcsLocationInfo.AdditionalNumber = an
		return w
	}
	return []semAddressCase{
		{"ReportSMDeliveryStatus msisdn", func(t *testing.T, a []byte) semWire {
			w, err := convertReportSMDeliveryStatusToArg(minimalReportSMDeliveryStatus())
			if err != nil {
				t.Fatal(err)
			}
			w.Msisdn = a
			return w
		}, semParse(ParseReportSMDeliveryStatus), ErrReportSMDeliveryStatusMSISDNEmpty},
		{"ReportSMDeliveryStatus serviceCentreAddress", func(t *testing.T, a []byte) semWire {
			w, err := convertReportSMDeliveryStatusToArg(minimalReportSMDeliveryStatus())
			if err != nil {
				t.Fatal(err)
			}
			w.ServiceCentreAddress = a
			return w
		}, semParse(ParseReportSMDeliveryStatus), ErrReportSMDeliveryStatusSCAEmpty},
		{"ProvideSubscriberLocation mlc-Number", func(t *testing.T, a []byte) semWire {
			w, err := convertProvideSubscriberLocationArgToWire(semPSLArg())
			if err != nil {
				t.Fatal(err)
			}
			w.MlcNumber = a
			return w
		}, semParse(ParseProvideSubscriberLocation), ErrPSLArgMlcNumberEmpty},
		{"SendRoutingInfoForLCS mlc-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.RoutingInfoForLCSArg{MlcNumber: a, TargetMS: gsm_map.NewSubscriberIdentityImsi(semAddrIMSI)}
		}, semParse(ParseSriLcs), ErrSriLcsMlcNumberEmpty},
		{"SubscriberLocationReport networkNode-Number", func(t *testing.T, a []byte) semWire {
			w, err := convertSubscriberLocationReportArgToWire(semSLRArg())
			if err != nil {
				t.Fatal(err)
			}
			w.LcsLocationInfo.NetworkNodeNumber = a
			return w
		}, semParse(ParseSubscriberLocationReport), ErrLCSLocationInfoNetworkNodeEmpty},
		{"UpdateGprsLocation sgsn-Number", func(t *testing.T, a []byte) semWire {
			w, err := convertUpdateGprsLocationToArg(&UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "12", SGSNAddress: "192.0.2.1"})
			if err != nil {
				t.Fatal(err)
			}
			w.SgsnNumber = a
			return w
		}, semParse(ParseUpdateGprsLocation), ErrUpdateGprsLocationMissingSGSNNumber},
		{"UpdateGprsLocationRes hlr-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.UpdateGprsLocationRes{HlrNumber: a}
		}, semParse(ParseUpdateGprsLocationRes), ErrUpdateGprsLocationResMissingHLRNumber},
		{"GPRS-CSI gsmSCF-Address", func(t *testing.T, a []byte) semWire {
			w, err := convertInsertSubscriberDataArgToWire(&InsertSubscriberDataArg{SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{GprsCSI: &GPRSCSI{GprsCamelTDPDataList: GPRSCamelTDPDataList{makeGPRSCamelTDPData()}}}})
			if err != nil {
				t.Fatal(err)
			}
			w.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList.Values[0].GsmSCFAddress = a
			return w
		}, semParse(ParseInsertSubscriberData), ErrCamelMissingGsmSCFAddress},
		{"MG-CSI gsmSCF-Address", func(t *testing.T, a []byte) semWire {
			w, err := convertInsertSubscriberDataArgToWire(&InsertSubscriberDataArg{SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{MgCsi: makeMGCSI()}})
			if err != nil {
				t.Fatal(err)
			}
			w.SgsnCAMELSubscriptionInfo.MgCsi.GsmSCFAddress = a
			return w
		}, semParse(ParseInsertSubscriberData), ErrCamelMissingGsmSCFAddress},
		{"AlertServiceCentre msisdn", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.AlertServiceCentreArg{Msisdn: a, ServiceCentreAddress: semTwoDigits}
		}, semParse(ParseAlertServiceCentre), ErrAscMissingMSISDN},
		{"AlertServiceCentre serviceCentreAddress", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.AlertServiceCentreArg{Msisdn: semTwoDigits, ServiceCentreAddress: a}
		}, semParse(ParseAlertServiceCentre), ErrAscMissingServiceCentreAddress},
		{"SendRoutingInfo msisdn", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.SendRoutingInfoArg{Msisdn: a, GmscOrGsmSCFAddress: semTwoDigits}
		}, semParse(ParseSri), ErrSriMissingMSISDN},
		{"SendRoutingInfo gmsc-OrGsmSCF-Address", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.SendRoutingInfoArg{Msisdn: semTwoDigits, GmscOrGsmSCFAddress: a}
		}, semParse(ParseSri), ErrSriMissingGmsc},
		{"SendRoutingInfoRes roamingNumber", func(_ *testing.T, a []byte) semWire {
			eri := gsm_map.NewExtendedRoutingInfoRoutingInfo(gsm_map.NewRoutingInfoRoamingNumber(a))
			imsi := semAddrIMSI
			return &gsm_map.SendRoutingInfoRes{Imsi: &imsi, ExtendedRoutingInfo: &eri}
		}, semParse(ParseSriResp), ErrRoutingInfoRoamingNumberDecodedEmpty},
		{"RoutingInfoForSM msisdn", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.RoutingInfoForSMArg{Msisdn: a, SmRPPRI: true, ServiceCentreAddress: semTwoDigits}
		}, semParse(ParseSriSm), ErrSriSmMissingMSISDN},
		{"RoutingInfoForSM serviceCentreAddress", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.RoutingInfoForSMArg{Msisdn: semTwoDigits, SmRPPRI: true, ServiceCentreAddress: a}
		}, semParse(ParseSriSm), ErrSriSmMissingServiceCentreAddress},
		{"RoutingInfoForSMRes networkNode-Number", func(_ *testing.T, a []byte) semWire {
			return sriSmRes(a, nil, nil)
		}, semParse(ParseSriSmResp), ErrSriSmRespMissingNetworkNodeNumber},
		{"RoutingInfoForSMRes additional-Number msc-Number", func(_ *testing.T, a []byte) semWire {
			return sriSmRes(semTwoDigits, addl(true, a), nil)
		}, semParse(ParseSriSmResp), ErrAdditionalNumberMscNumberDecodedEmpty},
		{"RoutingInfoForSMRes additional-Number sgsn-Number", func(_ *testing.T, a []byte) semWire {
			return sriSmRes(semTwoDigits, addl(false, a), nil)
		}, semParse(ParseSriSmResp), ErrAdditionalNumberSgsnNumberDecodedEmpty},
		{"RoutingInfoForSMRes thirdNumber msc-Number", func(_ *testing.T, a []byte) semWire {
			return sriSmRes(semTwoDigits, nil, addl(true, a))
		}, semParse(ParseSriSmResp), ErrAdditionalNumberMscNumberDecodedEmpty},
		{"SubscriberLocationReport additional-Number sgsn-Number", func(t *testing.T, a []byte) semWire {
			return slr(t, addl(false, a))
		}, semParse(ParseSubscriberLocationReport), ErrAdditionalNumberSgsnNumberDecodedEmpty},
		{"UpdateLocation msc-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.UpdateLocationArg{Imsi: semAddrIMSI, MscNumber: a, VlrNumber: semTwoDigits}
		}, semParse(ParseUpdateLocation), ErrUpdateLocationMissingMSCNumber},
		{"UpdateLocation vlr-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.UpdateLocationArg{Imsi: semAddrIMSI, MscNumber: semTwoDigits, VlrNumber: a}
		}, semParse(ParseUpdateLocation), ErrUpdateLocationMissingVLRNumber},
		{"UpdateLocationRes hlr-Number", func(_ *testing.T, a []byte) semWire {
			return &gsm_map.UpdateLocationRes{HlrNumber: a}
		}, semParse(ParseUpdateLocationRes), ErrUpdateLocationResMissingHLRNumber},
		{"AnyTimeInterrogation gsmSCF-Address", func(_ *testing.T, a []byte) semWire {
			v := struct{}{}
			return &gsm_map.AnyTimeInterrogationArg{
				SubscriberIdentity: gsm_map.NewSubscriberIdentityImsi(semAddrIMSI),
				RequestedInfo:      gsm_map.RequestedInfo{SubscriberState: &v},
				GsmSCFAddress:      a,
			}
		}, semParse(ParseAnyTimeInterrogation), ErrAtiMissingGsmSCFAddress},
	}
}

func TestParseDigitlessMandatoryAddress(t *testing.T) {
	for _, c := range semAddressCases() {
		t.Run(c.name+"/no digits", func(t *testing.T) {
			data, err := c.build(t, semNoDigits).MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}
			if _, err := c.parse(data); !errors.Is(err, c.want) {
				t.Errorf("Parse: err = %v, want %v", err, c.want)
			}
		})
		t.Run(c.name+"/filler only", func(t *testing.T) {
			data, err := c.build(t, []byte{0x91, 0xFF}).MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}
			if _, err := c.parse(data); !errors.Is(err, c.want) {
				t.Errorf("Parse: err = %v, want %v", err, c.want)
			}
		})
		// One digit is the shortest address that parses, and it marshals
		// back to the same octets.
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

// The mandatory addresses that Marshal used to emit without digits.
func TestMarshalMissingMandatoryAddress(t *testing.T) {
	for _, tc := range []struct {
		name string
		msg  interface{ Marshal() ([]byte, error) }
		want error
	}{
		{"UpdateLocationRes HLRNumber", &UpdateLocationRes{}, ErrUpdateLocationResMissingHLRNumber},
		{"SriSmResp NetworkNodeNumber", &SriSmResp{IMSI: "001010123456789"}, ErrSriSmRespMissingNetworkNodeNumber},
		{"AnyTimeInterrogation GsmSCFAddress", &AnyTimeInterrogation{
			SubscriberIdentity: SubscriberIdentity{IMSI: "001010123456789"},
			RequestedInfo:      RequestedInfo{SubscriberState: true},
		}, ErrAtiMissingGsmSCFAddress},
		{"ReportSMDeliveryStatus MSISDN", &ReportSMDeliveryStatus{ServiceCentreAddress: "12"}, ErrReportSMDeliveryStatusMSISDNEmpty},
		{"ReportSMDeliveryStatus SCA", &ReportSMDeliveryStatus{MSISDN: "12"}, ErrReportSMDeliveryStatusSCAEmpty},
		{"ProvideSubscriberLocation MlcNumber", &ProvideSubscriberLocationArg{}, ErrPSLArgMlcNumberEmpty},
		{"SendRoutingInfoForLCS MlcNumber", &SriLcs{TargetMS: SubscriberIdentity{IMSI: "001010123456789"}}, ErrSriLcsMlcNumberEmpty},
		{"SubscriberLocationReport NetworkNodeNumber", func() *SubscriberLocationReportArg {
			v := minimalSLRArg()
			v.LcsLocationInfo.NetworkNodeNumber = ""
			return v
		}(), ErrLCSLocationInfoNetworkNodeEmpty},
		{"UpdateGprsLocation SgsnNumber", &UpdateGprsLocation{IMSI: "001010123456789", SGSNAddress: "192.0.2.1"}, ErrUpdateGprsLocationMissingSGSNNumber},
		{"UpdateGprsLocationRes HLRNumber", &UpdateGprsLocationRes{}, ErrUpdateGprsLocationResMissingHLRNumber},
		{"GPRS-CSI gsmSCF-Address", &InsertSubscriberDataArg{SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{GprsCSI: &GPRSCSI{GprsCamelTDPDataList: GPRSCamelTDPDataList{{GprsTriggerDetectionPoint: GPRSTDPAttach}}}}}, ErrCamelMissingGsmSCFAddress},
		{"MG-CSI gsmSCF-Address", &InsertSubscriberDataArg{SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{MgCsi: &MGCSI{MobilityTriggers: []MMCode{MMCodeGPRSAttach}}}}, ErrCamelMissingGsmSCFAddress},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.msg.Marshal(); !errors.Is(err, tc.want) {
				t.Errorf("Marshal: err = %v, want %v", err, tc.want)
			}
		})
	}
}
