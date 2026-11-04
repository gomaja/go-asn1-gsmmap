package gsmmap

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	identity "github.com/gomaja/go-3gpp-identity"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// Wires follow 3GPP TS 29.002 V19.1.0 §§17.7.7–17.7.8.
func returnErrorMarshalCases() []struct {
	name  string
	code  MapErrorCode
	value marshaler
	wire  string
} {
	return []struct {
		name  string
		code  MapErrorCode
		value marshaler
		wire  string
	}{
		{"unknown empty", 1, &UnknownSubscriberParam{}, "3000"},                                                                                                     // empty SEQUENCE
		{"unknown IMSI", 1, &UnknownSubscriberParam{UnknownSubscriberDiagnostic: new(UnknownSubscriberDiagnosticImsiUnknown)}, "30030a0100"},                        // diagnostic 0
		{"unknown subscription", 1, &UnknownSubscriberParam{UnknownSubscriberDiagnostic: new(UnknownSubscriberDiagnosticGprsEpsSubscriptionUnknown)}, "30030a0101"}, // diagnostic 1
		{"unknown NPDB", 1, &UnknownSubscriberParam{UnknownSubscriberDiagnostic: new(UnknownSubscriberDiagnosticNpdbMismatch)}, "30030a0102"},                       // diagnostic extension 2
		{"SMS empty", 6, &AbsentSubscriberSMParam{}, "3000"},
		{"SMS IMSI", 6, &AbsentSubscriberSMParam{IMSI: "001010"}, "30058103000101"},                                                                                                                                                // [1] IMSI
		{"SMS user identifier", 6, &AbsentSubscriberSMParam{UserIdentifierAlert: "001010123456789"}, "300a830800010121436587f9"},                                                                                                   // [3] IMSI
		{"SMS retransmission", 6, &AbsentSubscriberSMParam{RequestedRetransmissionTime: HexBytes{1, 2, 3, 4}}, "3006820401020304"},                                                                                                 // [2] Time
		{"SMS primary diagnostic", 6, &AbsentSubscriberSMParam{AbsentSubscriberDiagnosticSM: new(AbsentSubscriberDiagnosticSM(1))}, "3003020101"},                                                                                  // INTEGER 1
		{"SMS additional diagnostic", 6, &AbsentSubscriberSMParam{AdditionalAbsentSubscriberDiagnosticSM: new(AbsentSubscriberDiagnosticSM(4))}, "3003800104"},                                                                     // [0] INTEGER 4
		{"SMS shortest identities", 6, &AbsentSubscriberSMParam{IMSI: "001010", UserIdentifierAlert: "001010"}, "300a81030001018303000101"},                                                                                        // two 3-octet IMSIs
		{"SMS diagnostics", 6, &AbsentSubscriberSMParam{AbsentSubscriberDiagnosticSM: new(AbsentSubscriberDiagnosticSM(255)), AdditionalAbsentSubscriberDiagnosticSM: new(AbsentSubscriberDiagnosticSM(0))}, "3007020200ff800100"}, // primary 255, additional 0
		{"SMS full", 6, &AbsentSubscriberSMParam{AbsentSubscriberDiagnosticSM: new(AbsentSubscriberDiagnosticSM(0)), AdditionalAbsentSubscriberDiagnosticSM: new(AbsentSubscriberDiagnosticSM(255)), IMSI: "001010123456789", RequestedRetransmissionTime: HexBytes{1, 2, 3, 4}, UserIdentifierAlert: "001010123456789"}, "3021020100800200ff810800010121436587f9820401020304830800010121436587f9"}, // 3+4+10+6+10 = 33 octets
		{"roaming mandatory", 8, &RoamingNotAllowedParam{}, "30030a0100"},
		{"roaming full", 8, &RoamingNotAllowedParam{RoamingNotAllowedCause: RoamingNotAllowedCauseOperatorDeterminedBarring, AdditionalRoamingNotAllowedCause: new(AdditionalRoamingNotAllowedCauseSupportedRATTypesNotAllowed)}, "30060a0103800100"}, // base 3, additional 0
		{"illegal subscriber", 9, &IllegalSubscriberParam{}, "3000"},
		{"teleservice", 11, &TeleservNotProvParam{}, "3000"},
		{"illegal equipment", 12, &IllegalEquipmentParam{}, "3000"},
		{"call service", 13, &CallBarredParam{CallBarringCause: new(CallBarringCauseBarringServiceActive)}, "0a0100"}, // bare ENUMERATED
		{"call operator", 13, &CallBarredParam{CallBarringCause: new(CallBarringCauseOperatorBarring)}, "0a0101"},
		{"call extensible empty", 13, &CallBarredParam{ExtensibleCallBarredParam: &ExtensibleCallBarredParam{}}, "3000"},
		{"call originator", 13, &CallBarredParam{ExtensibleCallBarredParam: &ExtensibleCallBarredParam{CallBarringCause: new(CallBarringCauseOperatorBarring), UnauthorisedMessageOriginator: true}}, "30050a01018100"}, // cause and [1] NULL
		{"call cause only", 13, &CallBarredParam{ExtensibleCallBarredParam: &ExtensibleCallBarredParam{CallBarringCause: new(CallBarringCauseBarringServiceActive)}}, "30030a0100"},                                     // untagged cause
		{"call anonymous", 13, &CallBarredParam{ExtensibleCallBarredParam: &ExtensibleCallBarredParam{AnonymousCallRejection: true}}, "30028200"},                                                                       // [2] NULL
		{"facility empty", 21, &FacilityNotSupParam{}, "3000"},
		{"facility shape", 21, &FacilityNotSupParam{ShapeOfLocationEstimateNotSupported: true}, "30028000"},                                                        // [0] NULL
		{"facility capability", 21, &FacilityNotSupParam{NeededLcsCapabilityNotSupportedInServingNode: true}, "30028100"},                                          // [1] NULL
		{"facility full", 21, &FacilityNotSupParam{ShapeOfLocationEstimateNotSupported: true, NeededLcsCapabilityNotSupportedInServingNode: true}, "300480008100"}, // [0], [1] NULL
		{"absent empty", 27, &AbsentSubscriberParam{}, "3000"},
		{"system bare", 34, &SystemFailureParam{NetworkResource: new(NetworkResourceHlr)}, "0a0101"},
		{"system empty", 34, &SystemFailureParam{ExtensibleSystemFailureParam: &ExtensibleSystemFailureParam{}}, "3000"},
		{"system full", 34, &SystemFailureParam{ExtensibleSystemFailureParam: &ExtensibleSystemFailureParam{NetworkResource: new(NetworkResourceVlr), AdditionalNetworkResource: new(AdditionalNetworkResourceMme), FailureCauseParam: new(FailureCauseParamLimitReachedOnNumberOfConcurrentLocationRequests)}}, "30090a0102800107810100"}, // 3 ENUMERATED TLVs
		{"data missing", 35, &DataMissingParam{}, "3000"},
		{"unexpected empty", 36, &UnexpectedDataParam{}, "3000"},
		{"unexpected subscriber", 36, &UnexpectedDataParam{UnexpectedSubscriber: true}, "30028000"}, // [0] NULL
		{"unauthorized network", 52, &UnauthorizedRequestingNetworkParam{}, "3000"},
	}
}

func TestReturnErrorMarshalRoundTrip(t *testing.T) {
	for _, c := range returnErrorMarshalCases() {
		t.Run(c.name, func(t *testing.T) {
			data, err := c.value.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(data, mustHex(t, c.wire)) {
				t.Fatalf("Marshal = %x, want %s", data, c.wire)
			}
			got, err := ParseReturnErrorParameter(c.code, data)
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, c.name, any(c.value), got)
		})
	}
	// Each remaining enumerator is a one-octet value in its tagged TLV.
	for v, wires := range [][3]string{
		{"0a0100", "3003800100", "30030a0100"}, // plmn, sgsn
		{"0a0101", "3003800101", "30030a0101"}, // hlr, ggsn
		{"0a0102", "3003800102", "30030a0102"}, // vlr, gmlc
		{"0a0103", "3003800103", "30030a0103"}, // pvlr, gsmSCF
		{"0a0104", "3003800104", "30030a0104"}, // controllingMSC, nplr
		{"0a0105", "3003800105", "30030a0105"}, // vmsc, auc
		{"0a0106", "3003800106", "30030a0106"}, // eir, ue
		{"0a0107", "3003800107", "30030a0107"}, // rss, mme
	} {
		for _, c := range []struct {
			value marshaler
			wire  string
		}{
			{&SystemFailureParam{NetworkResource: new(NetworkResource(v))}, wires[0]},
			{&SystemFailureParam{ExtensibleSystemFailureParam: &ExtensibleSystemFailureParam{AdditionalNetworkResource: new(AdditionalNetworkResource(v))}}, wires[1]},
			{&SystemFailureParam{ExtensibleSystemFailureParam: &ExtensibleSystemFailureParam{NetworkResource: new(NetworkResource(v))}}, wires[2]},
		} {
			data, err := c.value.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(data, mustHex(t, c.wire)) {
				t.Fatalf("got %x want %s", data, c.wire)
			}
			got, err := ParseReturnErrorParameter(34, data)
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "system", any(c.value), got)
		}
	}
	for v, wire := range []string{
		"3003800100", // imsiDetach
		"3003800101", // restrictedArea
		"3003800102", // noPageResponse
		"3003800103", // purgedMS
		"3003800104", // mtRoamingRetry
		"3003800105", // busySubscriber
	} {
		c := &AbsentSubscriberParam{AbsentSubscriberReason: new(AbsentSubscriberReason(v))}
		data, err := c.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, mustHex(t, wire)) {
			t.Fatalf("reason %d: %x", v, data)
		}
		got, err := ParseReturnErrorParameter(27, data)
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "absent", any(c), got)
	}
}

func TestReturnErrorMarshalNil(t *testing.T) {
	for code, typ := range dispatchedParamTypes {
		v := reflect.Zero(typ).Interface().(marshaler)
		data, err := v.Marshal()
		if !errors.Is(err, ErrReturnErrorParameterNil) || data != nil {
			t.Errorf("code %d: %x, %v", code, data, err)
		}
	}
}

func TestReturnErrorMarshalSenderRules(t *testing.T) {
	cases := []struct {
		value marshaler
		want  error
	}{
		{&CallBarredParam{}, ErrCallBarredParamNoAlternative},
		{&CallBarredParam{CallBarringCause: new(CallBarringCauseBarringServiceActive), ExtensibleCallBarredParam: &ExtensibleCallBarredParam{}}, ErrCallBarredParamMultipleAlternatives},
		{&SystemFailureParam{}, ErrSystemFailureParamNoAlternative},
		{&SystemFailureParam{NetworkResource: new(NetworkResourceHlr), ExtensibleSystemFailureParam: &ExtensibleSystemFailureParam{}}, ErrSystemFailureParamMultipleAlternatives},
		{&CallBarredParam{ExtensibleCallBarredParam: &ExtensibleCallBarredParam{UnauthorisedMessageOriginator: true, AnonymousCallRejection: true}}, ErrExtensibleCallBarredParamMutuallyExclusiveIndicators},
		{&UnknownSubscriberParam{UnknownSubscriberDiagnostic: new(UnknownSubscriberDiagnostic(3))}, ErrUnknownSubscriberDiagnosticInvalid},
		{&SystemFailureParam{ExtensibleSystemFailureParam: &ExtensibleSystemFailureParam{AdditionalNetworkResource: new(AdditionalNetworkResource(8))}}, ErrAdditionalNetworkResourceInvalid},
		{&SystemFailureParam{ExtensibleSystemFailureParam: &ExtensibleSystemFailureParam{FailureCauseParam: new(FailureCauseParam(1))}}, ErrFailureCauseParamInvalid},
		{&AbsentSubscriberParam{AbsentSubscriberReason: new(AbsentSubscriberReason(6))}, ErrAbsentSubscriberReasonInvalid},
		{&RoamingNotAllowedParam{AdditionalRoamingNotAllowedCause: new(AdditionalRoamingNotAllowedCause(1))}, ErrAdditionalRoamingNotAllowedCauseInvalid},
		{&AbsentSubscriberSMParam{IMSI: "12345"}, identity.ErrIMSILength},
		{&AbsentSubscriberSMParam{UserIdentifierAlert: "12345"}, identity.ErrIMSILength},
	}
	for _, v := range []int64{-1, 128} {
		cases = append(cases,
			struct {
				value marshaler
				want  error
			}{&UnknownSubscriberParam{UnknownSubscriberDiagnostic: new(UnknownSubscriberDiagnostic(v))}, ErrUnknownSubscriberDiagnosticInvalid},
			struct {
				value marshaler
				want  error
			}{&SystemFailureParam{ExtensibleSystemFailureParam: &ExtensibleSystemFailureParam{AdditionalNetworkResource: new(AdditionalNetworkResource(v))}}, ErrAdditionalNetworkResourceInvalid},
			struct {
				value marshaler
				want  error
			}{&SystemFailureParam{ExtensibleSystemFailureParam: &ExtensibleSystemFailureParam{FailureCauseParam: new(FailureCauseParam(v))}}, ErrFailureCauseParamInvalid},
			struct {
				value marshaler
				want  error
			}{&AbsentSubscriberParam{AbsentSubscriberReason: new(AbsentSubscriberReason(v))}, ErrAbsentSubscriberReasonInvalid},
			struct {
				value marshaler
				want  error
			}{&RoamingNotAllowedParam{AdditionalRoamingNotAllowedCause: new(AdditionalRoamingNotAllowedCause(v))}, ErrAdditionalRoamingNotAllowedCauseInvalid},
		)
	}
	for _, c := range cases {
		_, err := c.value.Marshal()
		if !errors.Is(err, c.want) {
			t.Errorf("%+v: %v, want %v", c.value, err, c.want)
		}
	}
}

func TestReturnErrorMarshalIdentities(t *testing.T) {
	for _, c := range []struct {
		digits string
		want   error
	}{
		{"12345", identity.ErrIMSILength},
		{"1234567890123456", identity.ErrIMSILength},
		{"00101a123456789", identity.ErrNotDigits},
	} {
		for _, value := range []*AbsentSubscriberSMParam{{IMSI: c.digits}, {UserIdentifierAlert: c.digits}} {
			_, err := value.Marshal()
			if !errors.Is(err, c.want) {
				t.Errorf("%+v: %v, want %v", value, err, c.want)
			}
		}
	}
}

func TestReturnErrorMarshalCodecConstraints(t *testing.T) {
	cases := []struct {
		value            marshaler
		path, constraint string
	}{
		{&AbsentSubscriberSMParam{AbsentSubscriberDiagnosticSM: new(AbsentSubscriberDiagnosticSM(-1))}, "absentSubscriberDiagnosticSM", "(0..255)"},
		{&AbsentSubscriberSMParam{AbsentSubscriberDiagnosticSM: new(AbsentSubscriberDiagnosticSM(256))}, "absentSubscriberDiagnosticSM", "(0..255)"},
		{&AbsentSubscriberSMParam{AdditionalAbsentSubscriberDiagnosticSM: new(AbsentSubscriberDiagnosticSM(-1))}, "additionalAbsentSubscriberDiagnosticSM", "(0..255)"},
		{&AbsentSubscriberSMParam{AdditionalAbsentSubscriberDiagnosticSM: new(AbsentSubscriberDiagnosticSM(256))}, "additionalAbsentSubscriberDiagnosticSM", "(0..255)"},
		{&AbsentSubscriberSMParam{RequestedRetransmissionTime: HexBytes{1, 2, 3}}, "requestedRetransmissionTime", "SIZE (4)"},
		{&AbsentSubscriberSMParam{RequestedRetransmissionTime: HexBytes{1, 2, 3, 4, 5}}, "requestedRetransmissionTime", "SIZE (4)"},
		{&CallBarredParam{CallBarringCause: new(CallBarringCause(2))}, "callBarringCause", "ENUMERATED {0, 1}"},
		{&CallBarredParam{ExtensibleCallBarredParam: &ExtensibleCallBarredParam{CallBarringCause: new(CallBarringCause(2))}}, "extensibleCallBarredParam.callBarringCause", "ENUMERATED {0, 1}"},
		{&SystemFailureParam{NetworkResource: new(NetworkResource(8))}, "networkResource", "ENUMERATED {0, 1, 2, 3, 4, 5, 6, 7}"},
		{&SystemFailureParam{ExtensibleSystemFailureParam: &ExtensibleSystemFailureParam{NetworkResource: new(NetworkResource(8))}}, "extensibleSystemFailureParam.networkResource", "ENUMERATED {0, 1, 2, 3, 4, 5, 6, 7}"},
		{&RoamingNotAllowedParam{RoamingNotAllowedCause: RoamingNotAllowedCause(1)}, "roamingNotAllowedCause", "ENUMERATED {0, 3}"},
	}
	for _, c := range cases {
		_, err := c.value.Marshal()
		wantValidationError(t, err, c.path, c.constraint)
	}
	// An empty optional time is absent.
	data, err := (&AbsentSubscriberSMParam{RequestedRetransmissionTime: HexBytes{}}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "empty time", mustHex(t, "3000"), data)
}

// Unknown extension enumerators are discarded only where §17.7.7/§17.7.8 says so.
func TestReturnErrorReceiverRules(t *testing.T) {
	cases := []struct {
		code MapErrorCode
		wire string
		want any
	}{
		{1, "30030a0103", &UnknownSubscriberParam{}},                                                           // unknown diagnostic
		{34, "3003800108", &SystemFailureParam{ExtensibleSystemFailureParam: &ExtensibleSystemFailureParam{}}}, // unknown additional resource
		{34, "3003810101", &SystemFailureParam{ExtensibleSystemFailureParam: &ExtensibleSystemFailureParam{}}}, // unknown failure cause
		{27, "3003800106", &AbsentSubscriberParam{}},                                                           // unknown absent reason
		{8, "30060a0103800101", &RoamingNotAllowedParam{RoamingNotAllowedCause: RoamingNotAllowedCauseOperatorDeterminedBarring, AdditionalRoamingNotAllowedCause: new(AdditionalRoamingNotAllowedCause(1))}}, // retain additional cause
		{13, "300481008200", &CallBarredParam{ExtensibleCallBarredParam: &ExtensibleCallBarredParam{UnauthorisedMessageOriginator: true, AnonymousCallRejection: true}}},                                      // retain conflicting NULLs
	}
	for _, c := range cases {
		got, err := ParseReturnErrorParameter(c.code, mustHex(t, c.wire))
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "receiver", c.want, got)
	}
}

func returnErrorSemLenientCases() []semLenientCase {
	return []semLenientCase{
		{identity.ErrIMSILength, asParser(parseAbsentSubscriberSMParam), func(*testing.T) semWire {
			return &gsm_map.AbsentSubscriberSMParam{Imsi: new(gsm_map.IMSI{0x00, 0x01, 0xf1})}
		}},
		{identity.ErrIMSILength, asParser(parseAbsentSubscriberSMParam), func(*testing.T) semWire {
			return &gsm_map.AbsentSubscriberSMParam{UserIdentifierAlert: new(gsm_map.IMSI{0x00, 0x01, 0xf1})}
		}},
		{ErrAdditionalRoamingNotAllowedCauseInvalid, asParser(parseRoamingNotAllowedParam), func(*testing.T) semWire {
			return &gsm_map.RoamingNotAllowedParam{AdditionalRoamingNotAllowedCause: new(gsm_map.AdditionalRoamingNotAllowedCause(1))}
		}},
		{ErrExtensibleCallBarredParamMutuallyExclusiveIndicators, asParser(parseCallBarredParam), func(*testing.T) semWire {
			v := gsm_map.NewCallBarredParamExtensibleCallBarredParam(gsm_map.ExtensibleCallBarredParam{UnauthorisedMessageOriginator: &struct{}{}, AnonymousCallRejection: &struct{}{}})
			return &v
		}},
	}
}

func TestReturnErrorUnknownChoices(t *testing.T) {
	if _, err := convertWireToCallBarredParam(&gsm_map.CallBarredParam{Choice: 99}); !errors.Is(err, ErrCallBarredParamUnknownAlternative) {
		t.Errorf("call barred: %v", err)
	}
	if _, err := convertWireToSystemFailureParam(&gsm_map.SystemFailureParam{Choice: 99}); !errors.Is(err, ErrSystemFailureParamUnknownAlternative) {
		t.Errorf("system failure: %v", err)
	}
}

func TestReturnErrorParseIdentities(t *testing.T) {
	cases := []struct {
		wire string
		want error
	}{
		{"300581030001ff", identity.ErrIMSILength}, // IMSI four digits, 3 octets with filler
		{"300583030001ff", identity.ErrIMSILength}, // userIdentifierAlert four digits
		{"30058103ffffff", identity.ErrEmpty},      // IMSI all filler
		{"30058303ffffff", identity.ErrEmpty},      // userIdentifierAlert all filler
		{"300581030001a1", identity.ErrNotDigits},  // IMSI nondecimal
		{"300583030001a1", identity.ErrNotDigits},  // userIdentifierAlert nondecimal
	}
	for _, c := range cases {
		v, err := ParseReturnErrorParameter(MapErrorAbsentSubscriberSM, mustHex(t, c.wire))
		if v != nil || !errors.Is(err, c.want) {
			t.Errorf("%s: %v, %v; want nil, %v", c.wire, v, err, c.want)
		}
	}
	for _, c := range []struct{ wire, path string }{
		{"300481020001", "imsi"},
		{"300483020001", "userIdentifierAlert"},
	} {
		_, err := ParseReturnErrorParameter(MapErrorAbsentSubscriberSM, mustHex(t, c.wire))
		wantValidationError(t, err, c.path, "SIZE (3..8)")
	}
}
