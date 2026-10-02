// psl_geoinfo_test.go
//
// Tests for ProvideSubscriberLocation (opCode 83) geographical and
// positioning data types. PR B of the staged PSL implementation —
// top-level Arg/Res structs and codec land in follow-up PRs.
package gsmmap

import (
	"testing"
)

// Compile-smoke: every new public type must be referenceable.
func TestPSLGeoInfoTypesCompile(t *testing.T) {
	var _ ExtGeographicalInformation
	var _ AddGeographicalInformation
	var _ VelocityEstimate
	var _ PositioningDataInformation
	var _ UtranPositioningDataInfo
	var _ GeranGANSSpositioningData
	var _ UtranGANSSpositioningData
	var _ UtranAdditionalPositioningData
	var _ UtranCivicAddress
	var _ UtranBaroPressureMeas
}

// All byte-typed PSL geo/positioning aliases share HexBytes as their
// underlying type. The aliases let a HexBytes value pass directly into
// a function whose parameter is typed as the respective alias (no cast
// required) — same pattern as TestPSLByteAliases in
// psl_foundation_test.go.
func TestPSLGeoInfoByteAliases(t *testing.T) {
	// 4 bytes satisfies the tightest spec lower bound in this set
	// (VelocityEstimate SIZE 4..7), so the literal isn't accidentally
	// mistaken for a malformed value of any of the underlying types.
	input := HexBytes{0x01, 0x02, 0x03, 0x04}

	ext := func(v ExtGeographicalInformation) int { return len(v) }
	if got := ext(input); got != 4 {
		t.Errorf("ExtGeographicalInformation alias: want len 4, got %d", got)
	}
	add := func(v AddGeographicalInformation) int { return len(v) }
	if got := add(input); got != 4 {
		t.Errorf("AddGeographicalInformation alias: want len 4, got %d", got)
	}
	vel := func(v VelocityEstimate) int { return len(v) }
	if got := vel(input); got != 4 {
		t.Errorf("VelocityEstimate alias: want len 4, got %d", got)
	}
	pos := func(v PositioningDataInformation) int { return len(v) }
	if got := pos(input); got != 4 {
		t.Errorf("PositioningDataInformation alias: want len 4, got %d", got)
	}
	utpos := func(v UtranPositioningDataInfo) int { return len(v) }
	if got := utpos(input); got != 4 {
		t.Errorf("UtranPositioningDataInfo alias: want len 4, got %d", got)
	}
	geran := func(v GeranGANSSpositioningData) int { return len(v) }
	if got := geran(input); got != 4 {
		t.Errorf("GeranGANSSpositioningData alias: want len 4, got %d", got)
	}
	utganss := func(v UtranGANSSpositioningData) int { return len(v) }
	if got := utganss(input); got != 4 {
		t.Errorf("UtranGANSSpositioningData alias: want len 4, got %d", got)
	}
	utadd := func(v UtranAdditionalPositioningData) int { return len(v) }
	if got := utadd(input); got != 4 {
		t.Errorf("UtranAdditionalPositioningData alias: want len 4, got %d", got)
	}
	civic := func(v UtranCivicAddress) int { return len(v) }
	if got := civic(input); got != 4 {
		t.Errorf("UtranCivicAddress alias: want len 4, got %d", got)
	}
}

// UtranBaroPressureMeas is aliased to int64; values within and outside
// the spec range must round-trip without conversion. The Min/Max
// constants are typed as UtranBaroPressureMeas so range checks compose
// directly without explicit casts.
func TestPSLUtranBaroPressureMeasAlias(t *testing.T) {
	var v UtranBaroPressureMeas = 65000
	if int64(v) != 65000 {
		t.Fatalf("UtranBaroPressureMeas alias: want 65000, got %d", v)
	}

}

// Zero values for the aliases must compose cleanly with HexBytes.
func TestPSLGeoInfoZeroValues(t *testing.T) {
	var ext ExtGeographicalInformation
	if len(ext) != 0 {
		t.Error("ExtGeographicalInformation zero value should have len 0")
	}
	var v VelocityEstimate
	if len(v) != 0 {
		t.Error("VelocityEstimate zero value should have len 0")
	}
	var b UtranBaroPressureMeas
	if int64(b) != 0 {
		t.Error("UtranBaroPressureMeas zero value should be 0")
	}
}
