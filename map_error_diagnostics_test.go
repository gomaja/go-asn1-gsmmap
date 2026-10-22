// map_error_diagnostics_test.go
//
// Tests for the wrapper-level MAP ReturnError diagnostic types. PR F1
// of the staged ReturnError.Parameter implementation — parsers and the
// dispatcher land in follow-up PRs.
package gsmmap

import (
	"testing"
)

// Compile-smoke: every new public type must be referenceable.
func TestMapErrorParamTypesCompile(t *testing.T) {
	var _ AbsentSubscriberSMParam
	var _ UnknownSubscriberParam
	var _ CallBarredParam
	var _ ExtensibleCallBarredParam
	var _ SystemFailureParam
	var _ ExtensibleSystemFailureParam
	var _ RoamingNotAllowedParam
	var _ UnauthorizedRequestingNetworkParam
	var _ FacilityNotSupParam
	var _ TeleservNotProvParam
	var _ DataMissingParam
	var _ AbsentSubscriberParam
}

// Diagnostic-enum fields must keep their named upstream types so
// callers can call String() without dropping down to gsm_map.* — the
// whole point of this surface. Cover every enum referenced from the
// wrapper *Param types so a future upstream rename or visibility
// regression on String() is caught at test time, not by callers.
func TestMapErrorParamDiagnosticEnumStringers(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		// CallBarringCause (from CallBarredParam, ExtensibleCallBarredParam).
		{"CallBarringCauseBarringServiceActive", CallBarringCauseBarringServiceActive.String(), "barringServiceActive"},
		{"CallBarringCauseOperatorBarring", CallBarringCauseOperatorBarring.String(), "operatorBarring"},

		// RoamingNotAllowedCause (from RoamingNotAllowedParam). Spec lists
		// plmnRoamingNotAllowed and operatorDeterminedBarring as the two
		// canonical values.
		{"RoamingNotAllowedCausePlmnRoamingNotAllowed", RoamingNotAllowedCausePlmnRoamingNotAllowed.String(), "plmnRoamingNotAllowed"},
		{"RoamingNotAllowedCauseOperatorDeterminedBarring", RoamingNotAllowedCauseOperatorDeterminedBarring.String(), "operatorDeterminedBarring"},

		// AdditionalRoamingNotAllowedCause (from RoamingNotAllowedParam).
		{"AdditionalRoamingNotAllowedCauseSupportedRATTypesNotAllowed", AdditionalRoamingNotAllowedCauseSupportedRATTypesNotAllowed.String(), "supportedRAT-TypesNotAllowed"},

		// NetworkResource (from SystemFailureParam, ExtensibleSystemFailureParam).
		{"NetworkResourcePlmn", NetworkResourcePlmn.String(), "plmn"},
		{"NetworkResourceHlr", NetworkResourceHlr.String(), "hlr"},
		{"NetworkResourceVlr", NetworkResourceVlr.String(), "vlr"},

		// FailureCauseParam (from ExtensibleSystemFailureParam).
		{"FailureCauseParamLimitReachedOnNumberOfConcurrentLocationRequests", FailureCauseParamLimitReachedOnNumberOfConcurrentLocationRequests.String(), "limitReachedOnNumberOfConcurrentLocationRequests"},

		// UnknownSubscriberDiagnostic (from UnknownSubscriberParam).
		{"UnknownSubscriberDiagnosticImsiUnknown", UnknownSubscriberDiagnosticImsiUnknown.String(), "imsiUnknown"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s: want %q, got %q", tc.name, tc.want, tc.got)
		}
	}
}

// Zero values for the new structs must compose cleanly. Each field is
// either a pointer (nil) or a primitive zero — no required fields are
// dereferenced at zero-value time.
func TestMapErrorParamZeroValues(t *testing.T) {
	var asm AbsentSubscriberSMParam
	if asm.AbsentSubscriberDiagnosticSM != nil {
		t.Error("AbsentSubscriberSMParam zero: AbsentSubscriberDiagnosticSM should be nil")
	}
	if asm.IMSI != "" {
		t.Error("AbsentSubscriberSMParam zero: IMSI should be empty")
	}

	var us UnknownSubscriberParam
	if us.UnknownSubscriberDiagnostic != nil {
		t.Error("UnknownSubscriberParam zero: UnknownSubscriberDiagnostic should be nil")
	}

	var cb CallBarredParam
	if cb.CallBarringCause != nil || cb.ExtensibleCallBarredParam != nil {
		t.Error("CallBarredParam zero: both CHOICE alternatives should be nil")
	}

	var sf SystemFailureParam
	if sf.NetworkResource != nil || sf.ExtensibleSystemFailureParam != nil {
		t.Error("SystemFailureParam zero: both CHOICE alternatives should be nil")
	}

	var rna RoamingNotAllowedParam
	if rna.RoamingNotAllowedCause != RoamingNotAllowedCausePlmnRoamingNotAllowed {
		t.Error("RoamingNotAllowedParam zero: cause should be plmnRoamingNotAllowed")
	}
	if rna.AdditionalRoamingNotAllowedCause != nil {
		t.Error("RoamingNotAllowedParam zero: additional cause should be nil")
	}

	var fns FacilityNotSupParam
	if fns.ShapeOfLocationEstimateNotSupported || fns.NeededLcsCapabilityNotSupportedInServingNode {
		t.Error("FacilityNotSupParam zero: NULL flags should be false")
	}
}

// Diagnostic-enum fields are pointers so that "absent" is distinguishable
// from "value 0". Verify the pointer indirection works as expected.
func TestMapErrorParamPointerSemantics(t *testing.T) {
	cause := CallBarringCauseOperatorBarring
	cb := CallBarredParam{
		CallBarringCause: &cause,
	}
	if cb.CallBarringCause == nil {
		t.Fatal("CallBarringCause should not be nil after assignment")
	}
	if *cb.CallBarringCause != CallBarringCauseOperatorBarring {
		t.Errorf("CallBarringCause: want OperatorBarring, got %v", *cb.CallBarringCause)
	}

	// String() works on the pointer dereferenced value.
	if cb.CallBarringCause.String() != "operatorBarring" {
		t.Errorf("CallBarringCause.String(): want %q, got %q", "operatorBarring", cb.CallBarringCause.String())
	}
}
