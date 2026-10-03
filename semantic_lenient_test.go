// semantic_lenient_test.go
//
// Every error FuzzParse accepts from Marshal (strictEncodeErrors) marks a
// value the package deliberately decodes but does not send. Each case below
// parses such a value and checks that Marshal reports exactly that error, so
// the allow-list holds no error a parsed value cannot produce.
package gsmmap

import (
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

	"github.com/gomaja/go-asn1-gsmmap/address"
)

type semLenientCase struct {
	want  error
	parse func([]byte) (marshaler, error)
	// wire returns a valid wire message with the lenient value set.
	wire func(t *testing.T) semWire
}

// semMust returns v; a fixture that fails to build is a broken test.
func semMust[T any](v T, err error) T {
	if err != nil {
		panic("building the fixture: " + err.Error())
	}
	return v
}

func semLenientCases() []semLenientCase {
	pslWithClient := func(t *testing.T) *gsm_map.ProvideSubscriberLocationArg {
		a := semPSLArg()
		a.LcsClientID = &LCSClientID{
			LcsClientType: LCSClientTypeEmergencyServices,
			LcsClientName: &LCSClientName{DataCodingScheme: USSDDataCodingSchemeGSM7, NameString: HexBytes{0x41}, LcsFormatIndicator: new(LCSFormatIndicator)},
		}
		a.LcsPrivacyCheck = &LCSPrivacyCheck{CallSessionUnrelated: PrivacyCheckAllowedWithoutNotification}
		a.AreaEventInfo = &AreaEventInfo{
			AreaDefinition: AreaDefinition{AreaList: AreaList{{AreaType: AreaTypeCountryCode, AreaIdentification: AreaIdentification{0x21, 0xF3}}}},
			OccurrenceInfo: new(OccurrenceInfo),
		}
		a.ReportingPLMNList = &ReportingPLMNList{PlmnList: PLMNList{{PlmnId: HexBytes{0x21, 0xF3, 0x54}, RanTechnology: new(RANTechnology)}}}
		return semMust(convertProvideSubscriberLocationArgToWire(a))
	}
	cch := semCamelPhase
	camel := func(t *testing.T) *gsm_map.InsertSubscriberDataArg {
		return semMust(convertInsertSubscriberDataArgToWire(&InsertSubscriberDataArg{
			VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{
				OCSI: &OCSI{
					OBcsmCamelTDPDataList: []OBcsmCamelTDPData{{
						OBcsmTriggerDetectionPoint: OBcsmTriggerCollectedInfo, ServiceKey: 1,
						GsmSCFAddress: "31611111111", GsmSCFAddressNature: address.NatureInternational, GsmSCFAddressPlan: address.PlanISDN,
					}},
					CamelCapabilityHandling: &cch,
				},
				MoSmsCSI:                  &SMSCSI{SmsCAMELTDPDataList: []SMSCAMELTDPData{semSMSTDPData(SMSTriggerDetectionPointSmsCollectedInfo)}},
				TBcsmCamelTDPCriteriaList: []TBcsmCamelTDPCriteria{{TBcsmTriggerDetectionPoint: TBcsmTriggerTermAttemptAuthorized}},
			},
			SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{GprsCSI: &GPRSCSI{GprsCamelTDPDataList: GPRSCamelTDPDataList{{
				GprsTriggerDetectionPoint: GPRSTDPAttach, ServiceKey: 1, GsmSCFAddress: "31611111111",
			}}}},
		}))
	}
	// cancel sets CancellationType ct and, when tou is not nil, TypeOfUpdate.
	cancel := func(ct int64, tou *gsm_map.TypeOfUpdate) func(*testing.T) semWire {
		return func(*testing.T) semWire {
			c := gsm_map.CancellationType(ct)
			return &gsm_map.CancelLocationArg{Identity: gsm_map.NewIdentityImsi(semAddrIMSI), CancellationType: &c, TypeOfUpdate: tou}
		}
	}
	unknownUpdate := gsm_map.TypeOfUpdate(2)
	return []semLenientCase{
		{ErrAlertingPatternReserved, asParser(ParseUSSDArg), func(t *testing.T) semWire {
			w := semMust(convertUSSDArgToWire(&USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: HexBytes{0x41}}))
			w.AlertingPattern = &gsm_map.AlertingPattern{0x03}
			return w
		}},
		{ErrIMEISpareDigitNotZero, asParser(ParseAnyTimeInterrogationRes), func(t *testing.T) semWire {
			imei := semTBCD(t, "490154203237518")
			return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{Imei: &imei}}
		}},
		{ErrCancelLocInvalidCancellationType, asParser(ParseCancelLocation), cancel(3, nil)},
		{ErrCancelLocInvalidTypeOfUpdate, asParser(ParseCancelLocation), cancel(0, &unknownUpdate)},
		{ErrCamelInvalidTTriggerPoint, asParser(ParseInsertSubscriberData), func(t *testing.T) semWire {
			w := camel(t)
			w.VlrCamelSubscriptionInfo.TBCSMCAMELTDPCriteriaList.Values[0].TBCSMTriggerDetectionPoint = 15
			return w
		}},
		{ErrCamelInvalidDefaultCallHandling, asParser(ParseInsertSubscriberData), func(t *testing.T) semWire {
			w := camel(t)
			w.VlrCamelSubscriptionInfo.OCSI.OBcsmCamelTDPDataList.Values[0].DefaultCallHandling = -1
			return w
		}},
		{ErrCamelInvalidDefaultSMSHandling, asParser(ParseInsertSubscriberData), func(t *testing.T) semWire {
			w := camel(t)
			w.VlrCamelSubscriptionInfo.MoSmsCSI.SmsCAMELTDPDataList.Values[0].DefaultSMSHandling = -1
			return w
		}},
		{ErrDefaultGPRSHandlingInvalid, asParser(ParseInsertSubscriberData), func(t *testing.T) semWire {
			w := camel(t)
			w.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList.Values[0].DefaultSessionHandling = -1
			return w
		}},
		{ErrSaiInvalidRequestingNodeType, asParser(ParseSendAuthenticationInfo), func(*testing.T) semWire {
			r := gsm_map.RequestingNodeType(-1)
			return &gsm_map.SendAuthenticationInfoArg{Imsi: semAddrIMSI, NumberOfRequestedVectors: 1, RequestingNodeType: &r}
		}},
		{ErrRequestedDomainInvalid, asParser(ParseAnyTimeInterrogation), func(t *testing.T) semWire {
			w := semMust(convertATIToArg(&AnyTimeInterrogation{
				SubscriberIdentity: SubscriberIdentity{IMSI: "001010123456789"},
				RequestedInfo:      RequestedInfo{SubscriberState: true},
				GsmSCFAddress:      "31611111111",
			}))
			d := gsm_map.DomainType(-1)
			w.RequestedInfo.RequestedDomain = &d
			return w
		}},
		{ErrISTSupportIndicatorInvalid, asParser(ParseSri), func(t *testing.T) semWire {
			w := semMust(convertSriToArg(&Sri{MSISDN: "31612345678", GmscOrGsmSCFAddress: "31600000001"}))
			ist := gsm_map.ISTSupportIndicator(-1)
			w.IstSupportIndicator = &ist
			return w
		}},
		{ErrUnavailabilityCauseInvalid, asParser(ParseSriResp), func(*testing.T) semWire {
			uc := gsm_map.UnavailabilityCause(7)
			return &gsm_map.SendRoutingInfoRes{UnavailabilityCause: &uc}
		}},
		{ErrLCSClientInternalIDInvalid, asParser(ParseProvideSubscriberLocation), func(t *testing.T) semWire {
			w := pslWithClient(t)
			id := gsm_map.LCSClientInternalID(5)
			w.LcsClientID.LcsClientInternalID = &id
			return w
		}},
		{ErrLCSClientTypeInvalid, asParser(ParseProvideSubscriberLocation), func(t *testing.T) semWire {
			w := pslWithClient(t)
			w.LcsClientID.LcsClientType = 4
			w.PrivacyOverride = &struct{}{}
			return w
		}},
		{ErrLCSFormatIndicatorInvalid, asParser(ParseProvideSubscriberLocation), func(t *testing.T) semWire {
			w := pslWithClient(t)
			f := gsm_map.LCSFormatIndicator(5)
			w.LcsClientID.LcsClientName.LcsFormatIndicator = &f
			return w
		}},
		{ErrAreaTypeInvalid, asParser(ParseProvideSubscriberLocation), func(t *testing.T) semWire {
			w := pslWithClient(t)
			w.AreaEventInfo.AreaDefinition.AreaList.Values[0].AreaType = 6
			return w
		}},
		{ErrOccurrenceInfoInvalid, asParser(ParseProvideSubscriberLocation), func(t *testing.T) semWire {
			w := pslWithClient(t)
			o := gsm_map.OccurrenceInfo(2)
			w.AreaEventInfo.OccurrenceInfo = &o
			return w
		}},
		{ErrRANTechnologyInvalid, asParser(ParseProvideSubscriberLocation), func(t *testing.T) semWire {
			w := pslWithClient(t)
			r := gsm_map.RANTechnology(2)
			w.ReportingPLMNList.PlmnList.Values[0].RanTechnology = &r
			return w
		}},
		{ErrAccuracyFulfilmentIndicatorInvalid, asParser(ParseProvideSubscriberLocationRes), func(t *testing.T) semWire {
			w := semMust(convertProvideSubscriberLocationResToWire(&ProvideSubscriberLocationRes{LocationEstimate: ExtGeographicalInformation{0x10, 0, 0, 0, 0, 0, 0}}))
			a := gsm_map.AccuracyFulfilmentIndicator(2)
			w.AccuracyFulfilmentIndicator = &a
			return w
		}},
	}
}

func TestStrictEncodeErrorsAreReachable(t *testing.T) {
	cases := semLenientCases()
	covered := map[error]bool{}
	for _, c := range cases {
		t.Run(c.want.Error(), func(t *testing.T) {
			data, err := c.wire(t).MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}
			v, err := c.parse(data)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			_, err = v.Marshal()
			if !errors.Is(err, c.want) {
				t.Fatalf("Marshal: err = %v, want %v", err, c.want)
			}
			if !isStrictEncodeError(err) {
				t.Errorf("isStrictEncodeError(%v) = false", err)
			}
		})
		covered[c.want] = true
	}
	for _, e := range strictEncodeErrors {
		if !covered[e] {
			t.Errorf("strictEncodeErrors holds %v, which no case reaches", e)
		}
	}
}

// Errors outside the allow-list are not accepted, even when they wrap a
// field name the list mentions.
func TestIsStrictEncodeErrorIsExact(t *testing.T) {
	for _, err := range []error{
		ErrIMSIInvalidLength,
		ErrIMEIInvalidLength,
		ErrAscMissingMSISDN,
		errors.New("cancelLocation: CancellationType must be one of updateProcedure(0)"),
	} {
		if isStrictEncodeError(err) {
			t.Errorf("isStrictEncodeError(%v) = true", err)
		}
	}
}

type semFakeMsg struct{ err error }

func (m *semFakeMsg) Marshal() ([]byte, error) { return nil, m.err }

// The FuzzParse property fails a parsed value that does not marshal, unless
// Marshal reports an allow-listed error.
func TestCheckParseRoundTripMarshalErrors(t *testing.T) {
	for _, tc := range []struct {
		err  error
		pass bool
	}{
		{ErrIMSIInvalidLength, false},
		{fmt.Errorf("wrapped: %w", ErrAscMissingMSISDN), false},
		{ErrAlertingPatternReserved, true},
		{fmt.Errorf("CancellationType: %w", ErrCancelLocInvalidCancellationType), true},
	} {
		parse := func([]byte) (marshaler, error) { return &semFakeMsg{tc.err}, nil }
		err := checkParseRoundTrip("fake", parse, []byte{0x30, 0x00})
		if (err == nil) != tc.pass {
			t.Errorf("Marshal error %v: checkParseRoundTrip = %v, want pass %t", tc.err, err, tc.pass)
		}
	}
}

// 3GPP TS 29.002 V19.1.0 §17.7.3 SendRoutingInfoRes: "IMSI must be present
// if SendRoutingInfoRes is not segmented. If the TC-Result-NL segmentation
// option is taken the IMSI must be present in one segmented transmission of
// SendRoutingInfoRes." A segment without the IMSI marshals and parses.
func TestSriRespWithoutIMSI(t *testing.T) {
	in := &SriResp{MSISDN: "31612345678"}
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var w gsm_map.SendRoutingInfoRes
	if err := w.UnmarshalBER(data); err != nil {
		t.Fatalf("UnmarshalBER: %v", err)
	}
	if w.Imsi != nil {
		t.Errorf("wire imsi = %x, want absent", *w.Imsi)
	}
	got, err := ParseSriResp(data)
	if err != nil {
		t.Fatalf("ParseSriResp: %v", err)
	}
	semWantEqual(t, "SriResp", in, got)

	// The FuzzParse seed that found it: a SendRoutingInfoRes holding no
	// field the package reads.
	seed, err := hex.DecodeString("a30804032143650a0103")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkParseRoundTrip("SriResp", asParser(ParseSriResp), seed); err != nil {
		t.Error(err)
	}
}
