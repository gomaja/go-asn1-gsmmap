// psl_area_periodic_test.go
//
// Tests for ProvideSubscriberLocation (opCode 83) area-event,
// periodic, reporting-PLMN, and serving-node-address types. PR C of
// the staged PSL implementation — top-level Arg/Res structs and codec
// land in follow-up PRs.
package gsmmap

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// Compile-smoke: every new public type must be referenceable.
func TestPSLAreaPeriodicTypesCompile(t *testing.T) {
	var _ AreaType
	var _ AreaIdentification
	var _ Area
	var _ AreaList
	var _ AreaDefinition
	var _ OccurrenceInfo
	var _ IntervalTime
	var _ AreaEventInfo
	var _ ReportingAmount
	var _ ReportingInterval
	var _ PeriodicLDRInfo
	var _ RANTechnology
	var _ ReportingPLMN
	var _ PLMNList
	var _ ReportingPLMNList
	var _ TerminationCause
	var _ ServingNodeAddress

	// Constants exist (compile-smoke). Numeric equivalence to upstream
	// values is verified by TestPSLAreaPeriodicEnumsAliasUpstream below.
	_ = AreaTypeCountryCode
	_ = AreaTypePlmnId
	_ = AreaTypeLocationAreaId
	_ = AreaTypeRoutingAreaId
	_ = AreaTypeCellGlobalId
	_ = AreaTypeUtranCellId

	_ = OccurrenceOneTimeEvent
	_ = OccurrenceMultipleTimeEvent

	_ = RANTechnologyGsm
	_ = RANTechnologyUmts

	_ = TerminationNormal
	_ = TerminationErrorundefined
	_ = TerminationInternalTimeout
	_ = TerminationCongestion
	_ = TerminationMtLrRestart
	_ = TerminationPrivacyViolation
	_ = TerminationShapeOfLocationEstimateNotSupported
	_ = TerminationSubscriberTermination
	_ = TerminationUETermination
	_ = TerminationNetworkTermination

}

// Aliased enums must resolve to the same numeric values as upstream so
// callers can use either local or upstream names interchangeably.
func TestPSLAreaPeriodicEnumsAliasUpstream(t *testing.T) {
	cases := []struct {
		name  string
		local int64
		upstr int64
	}{
		{"AreaTypeCountryCode", int64(AreaTypeCountryCode), int64(gsm_map.AreaTypeCountryCode)},
		{"AreaTypeUtranCellId", int64(AreaTypeUtranCellId), int64(gsm_map.AreaTypeUtranCellId)},
		{"OccurrenceOneTimeEvent", int64(OccurrenceOneTimeEvent), int64(gsm_map.OccurrenceInfoOneTimeEvent)},
		{"OccurrenceMultipleTimeEvent", int64(OccurrenceMultipleTimeEvent), int64(gsm_map.OccurrenceInfoMultipleTimeEvent)},
		{"RANTechnologyGsm", int64(RANTechnologyGsm), int64(gsm_map.RANTechnologyGsm)},
		{"RANTechnologyUmts", int64(RANTechnologyUmts), int64(gsm_map.RANTechnologyUmts)},
		{"TerminationNormal", int64(TerminationNormal), int64(gsm_map.TerminationCauseNormal)},
		{"TerminationNetworkTermination", int64(TerminationNetworkTermination), int64(gsm_map.TerminationCauseNetworkTermination)},
	}
	for _, tc := range cases {
		if tc.local != tc.upstr {
			t.Errorf("%s: local=%d upstream=%d", tc.name, tc.local, tc.upstr)
		}
	}
}

// AreaIdentification is a HexBytes alias — a HexBytes literal flows
// directly into a function whose parameter is typed as the alias.
func TestPSLAreaIdentificationAlias(t *testing.T) {
	check := func(v AreaIdentification) int { return len(v) }
	if got := check(HexBytes{0x01, 0x02, 0x03}); got != 3 {
		t.Errorf("AreaIdentification alias: want len 3, got %d", got)
	}
}

// IntervalTime, ReportingAmount, ReportingInterval are int64 aliases —
// values flow without conversion; range-bound constants match the spec.
func TestPSLAreaPeriodicIntegerAliases(t *testing.T) {
	var iv IntervalTime = 60
	if iv != 60 {
		t.Errorf("IntervalTime alias: want 60, got %d", iv)
	}

	var amt ReportingAmount = 10
	var ivl ReportingInterval = 60
	if amt != 10 || ivl != 60 {
		t.Errorf("ReportingAmount/Interval aliases: want 10/60, got %d/%d", amt, ivl)
	}

	if PeriodicLDRProductMax != 8639999 {
		t.Errorf("PeriodicLDRProductMax: want 8639999, got %d", PeriodicLDRProductMax)
	}
}

// Sentinel errors must be defined, distinct, and detectable through
// errors.Is when wrapped via %w.
func TestPSLAreaPeriodicSentinelErrors(t *testing.T) {
	sentinels := []error{
		ErrAreaTypeInvalid,
		ErrOccurrenceInfoInvalid,
		ErrPeriodicLDRProductExceeded,
		ErrRANTechnologyInvalid,
		ErrTerminationCauseInvalid,
		ErrServingNodeAddressMultipleAlts,
		ErrServingNodeAddressNoAlt,
	}
	seen := make(map[error]int, len(sentinels))
	for i, s := range sentinels {
		if s == nil {
			t.Errorf("sentinel #%d is nil", i)
			continue
		}
		if j, dup := seen[s]; dup {
			t.Errorf("sentinel #%d aliases sentinel #%d (same error value)", i, j)
		}
		seen[s] = i
		wrapped := fmt.Errorf("psl wrapper: %w", s)
		if !errors.Is(wrapped, s) {
			t.Errorf("sentinel #%d not detectable through errors.Is when wrapped with %%w", i)
		}
	}
}

// Foundation struct shapes must be zero-value safe so the public API
// can be constructed incrementally before the codec lands.
func TestPSLAreaPeriodicZeroValues(t *testing.T) {
	var aei AreaEventInfo
	if len(aei.AreaDefinition.AreaList) != 0 {
		t.Error("AreaEventInfo zero value should have empty AreaList")
	}
	if aei.OccurrenceInfo != nil || aei.IntervalTime != nil {
		t.Error("AreaEventInfo zero value should have nil OccurrenceInfo/IntervalTime")
	}

	var ldr PeriodicLDRInfo
	if ldr.ReportingAmount != 0 || ldr.ReportingInterval != 0 {
		t.Error("PeriodicLDRInfo zero value should have zero ReportingAmount/Interval")
	}

	var rpl ReportingPLMNList
	if rpl.PlmnListPrioritized {
		t.Error("ReportingPLMNList zero value should have PlmnListPrioritized=false")
	}
	if len(rpl.PlmnList) != 0 {
		t.Error("ReportingPLMNList zero value should have empty PlmnList")
	}

	var sna ServingNodeAddress
	if sna.MscNumber != "" || sna.SgsnNumber != "" {
		t.Error("ServingNodeAddress zero value should have empty MscNumber/SgsnNumber digits")
	}
	if sna.MmeNumber != nil {
		t.Error("ServingNodeAddress zero value should have nil MmeNumber")
	}
}
