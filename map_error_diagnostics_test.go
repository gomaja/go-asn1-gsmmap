// map_error_diagnostics_test.go
//
// Tests for the wrapper-level MAP ReturnError diagnostic types. PR F1
// of the staged ReturnError.Parameter implementation — parsers and the
// dispatcher land in follow-up PRs.
package gsmmap

import (
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
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
}

// Diagnostic-enum fields must keep their named upstream types so
// callers can call String() without dropping down to gsm_map.* — the
// whole point of this surface.
func TestMapErrorParamDiagnosticEnumStringers(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "CallBarringCause.String()",
			got:  gsm_map.CallBarringCauseBarringServiceActive.String(),
			want: "barringServiceActive",
		},
		{
			name: "CallBarringCause operator",
			got:  gsm_map.CallBarringCauseOperatorBarring.String(),
			want: "operatorBarring",
		},
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
	if rna.RoamingNotAllowedCause != 0 {
		t.Error("RoamingNotAllowedParam zero: cause should be 0")
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
	cause := gsm_map.CallBarringCauseOperatorBarring
	cb := CallBarredParam{
		CallBarringCause: &cause,
	}
	if cb.CallBarringCause == nil {
		t.Fatal("CallBarringCause should not be nil after assignment")
	}
	if *cb.CallBarringCause != gsm_map.CallBarringCauseOperatorBarring {
		t.Errorf("CallBarringCause: want OperatorBarring, got %v", *cb.CallBarringCause)
	}

	// String() works on the pointer dereferenced value.
	if cb.CallBarringCause.String() != "operatorBarring" {
		t.Errorf("CallBarringCause.String(): want %q, got %q", "operatorBarring", cb.CallBarringCause.String())
	}
}
