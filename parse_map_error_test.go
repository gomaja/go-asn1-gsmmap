// parse_map_error_test.go
//
// Tests for the MAP ReturnError parameter parsers and the
// ParseReturnErrorParameter dispatcher.
package gsmmap

import (
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1-gsmmap/tbcd"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// =============================================================================
// AbsentSubscriberSMParam (errorCode 6)
// =============================================================================

func TestParseAbsentSubscriberSMParamRoundTrip(t *testing.T) {
	// Build a wire-form fixture, marshal it, then parse it back.
	imsiTBCD, _ := tbcd.Encode("001010123456789")
	wireImsi := gsm_map.IMSI(imsiTBCD)
	diag := gsm_map.AbsentSubscriberDiagnosticSM(1) // imsiDetached
	addDiag := gsm_map.AbsentSubscriberDiagnosticSM(4) // msPurged-ForNonGPRS
	wire := &gsm_map.AbsentSubscriberSMParam{
		AbsentSubscriberDiagnosticSM:           &diag,
		AdditionalAbsentSubscriberDiagnosticSM: &addDiag,
		Imsi:                                   &wireImsi,
	}
	data, err := wire.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER fixture: %v", err)
	}

	got, err := ParseAbsentSubscriberSMParam(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.AbsentSubscriberDiagnosticSM == nil || *got.AbsentSubscriberDiagnosticSM != 1 {
		t.Errorf("AbsentSubscriberDiagnosticSM: want 1 (imsiDetached), got %v", got.AbsentSubscriberDiagnosticSM)
	}
	if got.AdditionalAbsentSubscriberDiagnosticSM == nil || *got.AdditionalAbsentSubscriberDiagnosticSM != 4 {
		t.Errorf("AdditionalAbsentSubscriberDiagnosticSM: want 4, got %v", got.AdditionalAbsentSubscriberDiagnosticSM)
	}
	if got.IMSI != "001010123456789" {
		t.Errorf("IMSI: want %q, got %q", "001010123456789", got.IMSI)
	}
}

func TestParseAbsentSubscriberSMParamEmpty(t *testing.T) {
	// Empty SEQUENCE — all optional fields absent. Should decode cleanly.
	wire := &gsm_map.AbsentSubscriberSMParam{}
	data, err := wire.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	got, err := ParseAbsentSubscriberSMParam(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got == nil {
		t.Fatal("got nil result for empty SEQUENCE")
	}
	if got.AbsentSubscriberDiagnosticSM != nil || got.IMSI != "" {
		t.Errorf("expected zero-value struct, got %+v", got)
	}
}

// =============================================================================
// UnknownSubscriberParam (errorCode 1)
// =============================================================================

func TestParseUnknownSubscriberParamRoundTrip(t *testing.T) {
	diag := gsm_map.UnknownSubscriberDiagnosticImsiUnknown
	wire := &gsm_map.UnknownSubscriberParam{UnknownSubscriberDiagnostic: &diag}
	data, err := wire.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	got, err := ParseUnknownSubscriberParam(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.UnknownSubscriberDiagnostic == nil || *got.UnknownSubscriberDiagnostic != gsm_map.UnknownSubscriberDiagnosticImsiUnknown {
		t.Fatalf("UnknownSubscriberDiagnostic: want imsiUnknown, got %v", got.UnknownSubscriberDiagnostic)
	}
	if got.UnknownSubscriberDiagnostic.String() != "imsiUnknown" {
		t.Errorf("String(): want %q, got %q", "imsiUnknown", got.UnknownSubscriberDiagnostic.String())
	}
}

// =============================================================================
// CallBarredParam (errorCode 13)
// =============================================================================

func TestParseCallBarredParamLegacyChoiceRoundTrip(t *testing.T) {
	wire := gsm_map.NewCallBarredParamCallBarringCause(gsm_map.CallBarringCauseOperatorBarring)
	data, err := wire.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	got, err := ParseCallBarredParam(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.CallBarringCause == nil {
		t.Fatal("CallBarringCause should be set on legacy alternative")
	}
	if *got.CallBarringCause != gsm_map.CallBarringCauseOperatorBarring {
		t.Errorf("want operatorBarring, got %v", *got.CallBarringCause)
	}
	if got.ExtensibleCallBarredParam != nil {
		t.Error("ExtensibleCallBarredParam should be nil on legacy alternative")
	}
}

func TestParseCallBarredParamExtensibleChoiceRoundTrip(t *testing.T) {
	cause := gsm_map.CallBarringCauseBarringServiceActive
	wire := gsm_map.NewCallBarredParamExtensibleCallBarredParam(gsm_map.ExtensibleCallBarredParam{
		CallBarringCause:              &cause,
		UnauthorisedMessageOriginator: &struct{}{},
	})
	data, err := wire.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	got, err := ParseCallBarredParam(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.CallBarringCause != nil {
		t.Error("CallBarringCause (legacy) should be nil on extensible alternative")
	}
	if got.ExtensibleCallBarredParam == nil {
		t.Fatal("ExtensibleCallBarredParam should be set on extensible alternative")
	}
	if got.ExtensibleCallBarredParam.CallBarringCause == nil ||
		*got.ExtensibleCallBarredParam.CallBarringCause != gsm_map.CallBarringCauseBarringServiceActive {
		t.Errorf("nested CallBarringCause: want barringServiceActive, got %v",
			got.ExtensibleCallBarredParam.CallBarringCause)
	}
	if !got.ExtensibleCallBarredParam.UnauthorisedMessageOriginator {
		t.Error("UnauthorisedMessageOriginator should be true (NULL flag set)")
	}
	if got.ExtensibleCallBarredParam.AnonymousCallRejection {
		t.Error("AnonymousCallRejection should be false (NULL flag absent)")
	}
}

// =============================================================================
// SystemFailureParam (errorCode 34)
// =============================================================================

func TestParseSystemFailureParamLegacyChoiceRoundTrip(t *testing.T) {
	wire := gsm_map.NewSystemFailureParamNetworkResource(gsm_map.NetworkResourceVlr)
	data, err := wire.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	got, err := ParseSystemFailureParam(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.NetworkResource == nil || *got.NetworkResource != gsm_map.NetworkResourceVlr {
		t.Fatalf("NetworkResource: want vlr, got %v", got.NetworkResource)
	}
	if got.NetworkResource.String() != "vlr" {
		t.Errorf("String(): want %q, got %q", "vlr", got.NetworkResource.String())
	}
	if got.ExtensibleSystemFailureParam != nil {
		t.Error("ExtensibleSystemFailureParam should be nil on legacy alternative")
	}
}

func TestParseSystemFailureParamExtensibleChoiceRoundTrip(t *testing.T) {
	nr := gsm_map.NetworkResourceHlr
	addnr := gsm_map.AdditionalNetworkResourceMme
	wire := gsm_map.NewSystemFailureParamExtensibleSystemFailureParam(gsm_map.ExtensibleSystemFailureParam{
		NetworkResource:           &nr,
		AdditionalNetworkResource: &addnr,
	})
	data, err := wire.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	got, err := ParseSystemFailureParam(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.NetworkResource != nil {
		t.Error("NetworkResource (legacy) should be nil on extensible alternative")
	}
	if got.ExtensibleSystemFailureParam == nil {
		t.Fatal("ExtensibleSystemFailureParam should be set")
	}
	if got.ExtensibleSystemFailureParam.NetworkResource == nil ||
		*got.ExtensibleSystemFailureParam.NetworkResource != gsm_map.NetworkResourceHlr {
		t.Errorf("nested NetworkResource: want hlr, got %v",
			got.ExtensibleSystemFailureParam.NetworkResource)
	}
	if got.ExtensibleSystemFailureParam.AdditionalNetworkResource == nil ||
		*got.ExtensibleSystemFailureParam.AdditionalNetworkResource != gsm_map.AdditionalNetworkResourceMme {
		t.Errorf("AdditionalNetworkResource: want mme, got %v",
			got.ExtensibleSystemFailureParam.AdditionalNetworkResource)
	}
}

// =============================================================================
// RoamingNotAllowedParam (errorCode 8)
// =============================================================================

func TestParseRoamingNotAllowedParamRoundTrip(t *testing.T) {
	addCause := gsm_map.AdditionalRoamingNotAllowedCauseSupportedRATTypesNotAllowed
	wire := &gsm_map.RoamingNotAllowedParam{
		RoamingNotAllowedCause:           gsm_map.RoamingNotAllowedCausePlmnRoamingNotAllowed,
		AdditionalRoamingNotAllowedCause: &addCause,
	}
	data, err := wire.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	got, err := ParseRoamingNotAllowedParam(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.RoamingNotAllowedCause != gsm_map.RoamingNotAllowedCausePlmnRoamingNotAllowed {
		t.Errorf("RoamingNotAllowedCause: want plmnRoamingNotAllowed, got %v", got.RoamingNotAllowedCause)
	}
	if got.AdditionalRoamingNotAllowedCause == nil ||
		*got.AdditionalRoamingNotAllowedCause != gsm_map.AdditionalRoamingNotAllowedCauseSupportedRATTypesNotAllowed {
		t.Errorf("AdditionalRoamingNotAllowedCause: want supportedRAT-TypesNotAllowed, got %v",
			got.AdditionalRoamingNotAllowedCause)
	}
}

// =============================================================================
// FacilityNotSupParam (errorCode 21)
// =============================================================================

func TestParseFacilityNotSupParamRoundTrip(t *testing.T) {
	wire := &gsm_map.FacilityNotSupParam{
		ShapeOfLocationEstimateNotSupported:          &struct{}{},
		NeededLcsCapabilityNotSupportedInServingNode: &struct{}{},
	}
	data, err := wire.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	got, err := ParseFacilityNotSupParam(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !got.ShapeOfLocationEstimateNotSupported {
		t.Error("ShapeOfLocationEstimateNotSupported should be true")
	}
	if !got.NeededLcsCapabilityNotSupportedInServingNode {
		t.Error("NeededLcsCapabilityNotSupportedInServingNode should be true")
	}
}

// =============================================================================
// Empty-payload error params (UnauthorizedRequestingNetwork, TeleservNotProv, DataMissing)
// =============================================================================

func TestParseEmptyParams(t *testing.T) {
	cases := []struct {
		name string
		fn   func([]byte) (any, error)
	}{
		{"UnauthorizedRequestingNetwork", func(data []byte) (any, error) {
			return ParseUnauthorizedRequestingNetworkParam(data)
		}},
		{"TeleservNotProv", func(data []byte) (any, error) {
			return ParseTeleservNotProvParam(data)
		}},
		{"DataMissing", func(data []byte) (any, error) {
			return ParseDataMissingParam(data)
		}},
	}
	// Empty-SEQUENCE BER encoding.
	emptySeq := []byte{0x30, 0x00}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.fn(emptySeq)
			if err != nil {
				t.Errorf("Parse: %v", err)
			}
			if out == nil || reflect.ValueOf(out).IsNil() {
				t.Error("expected non-nil result for empty SEQUENCE")
			}
		})
	}
}

// =============================================================================
// ParseReturnErrorParameter dispatcher
// =============================================================================

func TestParseReturnErrorParameterDispatch(t *testing.T) {
	diag := gsm_map.AbsentSubscriberDiagnosticSM(1)
	wire := &gsm_map.AbsentSubscriberSMParam{AbsentSubscriberDiagnosticSM: &diag}
	data, err := wire.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	out, err := ParseReturnErrorParameter(6, data) // errorCode 6 = absentSubscriberSM
	if err != nil {
		t.Fatalf("ParseReturnErrorParameter: %v", err)
	}
	got, ok := out.(*AbsentSubscriberSMParam)
	if !ok {
		t.Fatalf("dispatcher returned %T, want *AbsentSubscriberSMParam", out)
	}
	if got.AbsentSubscriberDiagnosticSM == nil || *got.AbsentSubscriberDiagnosticSM != 1 {
		t.Errorf("dispatched value lost diagnostic field: %+v", got)
	}
}

func TestParseReturnErrorParameterUnknownErrorCode(t *testing.T) {
	// errorCode outside the handled set: returns (nil, nil) so callers
	// can dispatch unconditionally.
	out, err := ParseReturnErrorParameter(999, []byte{0x30, 0x00})
	if err != nil {
		t.Errorf("unknown errorCode: want nil error, got %v", err)
	}
	if out != nil {
		t.Errorf("unknown errorCode: want nil result, got %T", out)
	}
}

func TestParseReturnErrorParameterEmptyData(t *testing.T) {
	// Empty data: returns (nil, nil) regardless of errorCode.
	out, err := ParseReturnErrorParameter(6, nil)
	if err != nil {
		t.Errorf("empty data: want nil error, got %v", err)
	}
	if out != nil {
		t.Errorf("empty data: want nil result, got %T", out)
	}
}

// Build minimal valid BER fixtures per errorCode so the dispatcher
// smoke test asserts both routing and a successful decode. CHOICE
// types (CallBarred, SystemFailure) reject empty SEQUENCEs, so
// they need a selected alternative.
func buildDispatcherFixtures(t *testing.T) map[int64][]byte {
	t.Helper()
	fixtures := make(map[int64][]byte)

	// errorCode=1 (UnknownSubscriberParam): empty SEQUENCE works
	// (all fields optional).
	fixtures[1] = []byte{0x30, 0x00}

	// errorCode=6 (AbsentSubscriberSMParam): empty SEQUENCE works.
	fixtures[6] = []byte{0x30, 0x00}

	// errorCode=8 (RoamingNotAllowedParam): mandatory cause field —
	// build a wire fixture with cause=plmnRoamingNotAllowed.
	rna := &gsm_map.RoamingNotAllowedParam{
		RoamingNotAllowedCause: gsm_map.RoamingNotAllowedCausePlmnRoamingNotAllowed,
	}
	rnaData, err := rna.MarshalBER()
	if err != nil {
		t.Fatalf("RoamingNotAllowedParam fixture: %v", err)
	}
	fixtures[8] = rnaData

	// errorCode=11 (TeleservNotProvParam): empty SEQUENCE works.
	fixtures[11] = []byte{0x30, 0x00}

	// errorCode=13 (CallBarredParam): CHOICE — provide legacy alt.
	cb := gsm_map.NewCallBarredParamCallBarringCause(gsm_map.CallBarringCauseOperatorBarring)
	cbData, err := cb.MarshalBER()
	if err != nil {
		t.Fatalf("CallBarredParam fixture: %v", err)
	}
	fixtures[13] = cbData

	// errorCode=21 (FacilityNotSupParam): empty SEQUENCE works.
	fixtures[21] = []byte{0x30, 0x00}

	// errorCode=34 (SystemFailureParam): CHOICE — provide legacy alt.
	sf := gsm_map.NewSystemFailureParamNetworkResource(gsm_map.NetworkResourceVlr)
	sfData, err := sf.MarshalBER()
	if err != nil {
		t.Fatalf("SystemFailureParam fixture: %v", err)
	}
	fixtures[34] = sfData

	// errorCode=35 (DataMissingParam): empty SEQUENCE works.
	fixtures[35] = []byte{0x30, 0x00}

	// errorCode=52 (UnauthorizedRequestingNetworkParam): empty SEQUENCE.
	fixtures[52] = []byte{0x30, 0x00}

	return fixtures
}

func TestParseReturnErrorParameterAllDispatchedTypes(t *testing.T) {
	// Each errorCode listed in the dispatcher comment must route
	// successfully to the documented concrete type. Use minimal
	// valid BER fixtures (CHOICE types need a selected alternative)
	// so a routing regression is caught even on CHOICE types.
	fixtures := buildDispatcherFixtures(t)
	cases := []struct {
		errorCode int64
		want      string // type name
	}{
		{1, "*gsmmap.UnknownSubscriberParam"},
		{6, "*gsmmap.AbsentSubscriberSMParam"},
		{8, "*gsmmap.RoamingNotAllowedParam"},
		{11, "*gsmmap.TeleservNotProvParam"},
		{13, "*gsmmap.CallBarredParam"},
		{21, "*gsmmap.FacilityNotSupParam"},
		{34, "*gsmmap.SystemFailureParam"},
		{35, "*gsmmap.DataMissingParam"},
		{52, "*gsmmap.UnauthorizedRequestingNetworkParam"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			data, ok := fixtures[tc.errorCode]
			if !ok {
				t.Fatalf("missing fixture for errorCode=%d", tc.errorCode)
			}
			out, err := ParseReturnErrorParameter(tc.errorCode, data)
			if err != nil {
				t.Fatalf("errorCode=%d: unexpected parse error: %v", tc.errorCode, err)
			}
			gotType := reflect.TypeOf(out).String()
			if gotType != tc.want {
				t.Errorf("errorCode=%d: dispatcher returned %s, want %s", tc.errorCode, gotType, tc.want)
			}
		})
	}
}
