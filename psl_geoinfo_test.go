// psl_geoinfo_test.go
//
// Tests for ProvideSubscriberLocation (opCode 83) geographical and
// positioning data types. PR B of the staged PSL implementation —
// top-level Arg/Res structs and codec land in follow-up PRs.
package gsmmap

import (
	"errors"
	"fmt"
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
// underlying type. The aliases let callers pass HexBytes values
// through unchanged; tested by passing HexBytes through functions
// whose parameters are typed as the respective alias.
func TestPSLGeoInfoByteAliases(t *testing.T) {
	cases := []struct {
		name   string
		passer func(HexBytes) int
	}{
		{"ExtGeographicalInformation", func(v HexBytes) int { var x ExtGeographicalInformation = v; return len(x) }},
		{"AddGeographicalInformation", func(v HexBytes) int { var x AddGeographicalInformation = v; return len(x) }},
		{"VelocityEstimate", func(v HexBytes) int { var x VelocityEstimate = v; return len(x) }},
		{"PositioningDataInformation", func(v HexBytes) int { var x PositioningDataInformation = v; return len(x) }},
		{"UtranPositioningDataInfo", func(v HexBytes) int { var x UtranPositioningDataInfo = v; return len(x) }},
		{"GeranGANSSpositioningData", func(v HexBytes) int { var x GeranGANSSpositioningData = v; return len(x) }},
		{"UtranGANSSpositioningData", func(v HexBytes) int { var x UtranGANSSpositioningData = v; return len(x) }},
		{"UtranAdditionalPositioningData", func(v HexBytes) int { var x UtranAdditionalPositioningData = v; return len(x) }},
		{"UtranCivicAddress", func(v HexBytes) int { var x UtranCivicAddress = v; return len(x) }},
	}
	for _, tc := range cases {
		input := HexBytes{0x01, 0x02, 0x03}
		if got := tc.passer(input); got != 3 {
			t.Errorf("%s alias: want len 3, got %d", tc.name, got)
		}
	}
}

// UtranBaroPressureMeas is aliased to int64; values within and outside
// the spec range must round-trip without conversion.
func TestPSLUtranBaroPressureMeasAlias(t *testing.T) {
	var v UtranBaroPressureMeas = 65000
	if int64(v) != 65000 {
		t.Fatalf("UtranBaroPressureMeas alias: want 65000, got %d", v)
	}
	if int64(UtranBaroPressureMeasMin) != 30000 {
		t.Errorf("UtranBaroPressureMeasMin: want 30000, got %d", UtranBaroPressureMeasMin)
	}
	if int64(UtranBaroPressureMeasMax) != 115000 {
		t.Errorf("UtranBaroPressureMeasMax: want 115000, got %d", UtranBaroPressureMeasMax)
	}
}

// Sentinel errors must be defined, distinct, and detectable through
// errors.Is when wrapped via %w.
func TestPSLGeoInfoSentinelErrors(t *testing.T) {
	sentinels := []error{
		ErrExtGeographicalInformationSize,
		ErrAddGeographicalInformationSize,
		ErrVelocityEstimateSize,
		ErrPositioningDataInformationSize,
		ErrUtranPositioningDataInfoSize,
		ErrGeranGANSSpositioningDataSize,
		ErrUtranGANSSpositioningDataSize,
		ErrUtranAdditionalPositioningDataSize,
		ErrUtranBaroPressureMeasOutOfRange,
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

// Spec-derived size constants must match TS 29.002.
func TestPSLGeoInfoSpecConstants(t *testing.T) {
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"ExtGeographicalInformationMaxLen (maxExt-GeographicalInformation, asn:518)", ExtGeographicalInformationMaxLen, 20},
		{"AddGeographicalInformationMaxLen (maxAdd-GeographicalInformation, asn:619)", AddGeographicalInformationMaxLen, 91},
		{"VelocityEstimateMinLen (asn:522)", VelocityEstimateMinLen, 4},
		{"VelocityEstimateMaxLen (asn:522)", VelocityEstimateMaxLen, 7},
		{"PositioningDataInformationMinLen (asn:552)", PositioningDataInformationMinLen, 2},
		{"PositioningDataInformationMaxLen (maxPositioningDataInformation, asn:557)", PositioningDataInformationMaxLen, 10},
		{"UtranPositioningDataInfoMinLen (asn:560)", UtranPositioningDataInfoMinLen, 3},
		{"UtranPositioningDataInfoMaxLen (maxUtranPositioningDataInfo, asn:565)", UtranPositioningDataInfoMaxLen, 11},
		{"GeranGANSSpositioningDataMinLen (asn:568)", GeranGANSSpositioningDataMinLen, 2},
		{"GeranGANSSpositioningDataMaxLen (maxGeranGANSSpositioningData, asn:573)", GeranGANSSpositioningDataMaxLen, 10},
		{"UtranGANSSpositioningDataMaxLen (maxUtranGANSSpositioningData, asn:581)", UtranGANSSpositioningDataMaxLen, 9},
		{"UtranAdditionalPositioningDataMaxLen (maxUtranAdditionalPositioningData, asn:589)", UtranAdditionalPositioningDataMaxLen, 8},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s: want %d, got %d", tc.name, tc.want, tc.got)
		}
	}
}

// Zero values for the aliases must compose cleanly with HexBytes.
func TestPSLGeoInfoZeroValues(t *testing.T) {
	var ext ExtGeographicalInformation
	if ext != nil {
		t.Error("ExtGeographicalInformation zero value should be nil")
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
