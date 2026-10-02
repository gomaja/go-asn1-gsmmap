// convert_psl_area_periodic_test.go
//
// Tests for the PSL-Arg area-event tree, periodic LDR info, and
// reporting-PLMN list converters.
package gsmmap

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// Area
// ============================================================================

func TestAreaRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *Area
	}{
		{"countryCode min", &Area{
			AreaType:           AreaTypeCountryCode,
			AreaIdentification: HexBytes{0x01, 0x02},
		}},
		{"utranCellId max", &Area{
			AreaType:           AreaTypeUtranCellId,
			AreaIdentification: HexBytes{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertAreaToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := convertWireToArea(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", tc.in, out)
			}
		})
	}
}

func TestAreaOutOfRangeTypeRejected(t *testing.T) {
	_, err := convertAreaToWire(&Area{
		AreaType:           AreaType(99),
		AreaIdentification: HexBytes{0x01, 0x02},
	})
	if !errors.Is(err, ErrAreaTypeInvalid) {
		t.Errorf("want ErrAreaTypeInvalid, got %v", err)
	}
}

func TestAreaIdentificationSizeRejected(t *testing.T) {
	_, err := strictWire(convertAreaToWire(&Area{
		AreaType:           AreaTypeCountryCode,
		AreaIdentification: HexBytes{0x01}, // too small (min 2)
	}))
	if !isConstraint(err) {
		t.Errorf("encode 1 octet: want BER constraint error, got %v", err)
	}
	tooBig := make(HexBytes, 8) // too big (max 7)
	_, err = strictWire(convertAreaToWire(&Area{
		AreaType:           AreaTypeCountryCode,
		AreaIdentification: tooBig,
	}))
	if !isConstraint(err) {
		t.Errorf("encode 8 octets: want BER constraint error, got %v", err)
	}
}

// ============================================================================
// AreaList
// ============================================================================

func TestAreaListRoundTrip(t *testing.T) {
	in := AreaList{
		{AreaType: AreaTypeCountryCode, AreaIdentification: HexBytes{0x01, 0x02}},
		{AreaType: AreaTypePlmnId, AreaIdentification: HexBytes{0x03, 0x04, 0x05}},
	}
	wire, err := convertAreaListToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	out, err := convertWireToAreaList(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", in, out)
	}
}

func TestAreaListEmptyRejected(t *testing.T) {
	_, err := strictWire(convertAreaListToWire(AreaList{}))
	if !isConstraint(err) {
		t.Errorf("want BER constraint error for empty list, got %v", err)
	}
}

func TestAreaListOversizedRejected(t *testing.T) {
	tooMany := make(AreaList, 10+1)
	for i := range tooMany {
		tooMany[i] = Area{AreaType: AreaTypeCountryCode, AreaIdentification: HexBytes{0x01, 0x02}}
	}
	_, err := strictWire(convertAreaListToWire(tooMany))
	if !isConstraint(err) {
		t.Errorf("want BER constraint error for 11 entries, got %v", err)
	}
}

// ============================================================================
// AreaEventInfo
// ============================================================================

func TestAreaEventInfoRoundTrip(t *testing.T) {
	occ := OccurrenceMultipleTimeEvent
	intv := IntervalTime(120)
	cases := []struct {
		name string
		in   *AreaEventInfo
	}{
		{"minimal", &AreaEventInfo{
			AreaDefinition: AreaDefinition{
				AreaList: AreaList{{AreaType: AreaTypeCountryCode, AreaIdentification: HexBytes{0x01, 0x02}}},
			},
		}},
		{"full population", &AreaEventInfo{
			AreaDefinition: AreaDefinition{
				AreaList: AreaList{
					{AreaType: AreaTypePlmnId, AreaIdentification: HexBytes{0x01, 0x02, 0x03}},
					{AreaType: AreaTypeUtranCellId, AreaIdentification: HexBytes{0x04, 0x05, 0x06, 0x07}},
				},
			},
			OccurrenceInfo: &occ,
			IntervalTime:   &intv,
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertAreaEventInfoToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := convertWireToAreaEventInfo(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", tc.in, out)
			}
		})
	}
}

func TestAreaEventInfoIntervalTimeOutOfRangeRejected(t *testing.T) {
	bad := IntervalTime(0) // below min
	_, err := strictWire(convertAreaEventInfoToWire(&AreaEventInfo{
		AreaDefinition: AreaDefinition{
			AreaList: AreaList{{AreaType: AreaTypeCountryCode, AreaIdentification: HexBytes{0x01, 0x02}}},
		},
		IntervalTime: &bad,
	}))
	if !isConstraint(err) {
		t.Errorf("encode IntervalTime=0: want BER constraint error, got %v", err)
	}

	tooBig := IntervalTime(32767 + 1)
	_, err = strictWire(convertAreaEventInfoToWire(&AreaEventInfo{
		AreaDefinition: AreaDefinition{
			AreaList: AreaList{{AreaType: AreaTypeCountryCode, AreaIdentification: HexBytes{0x01, 0x02}}},
		},
		IntervalTime: &tooBig,
	}))
	if !isConstraint(err) {
		t.Errorf("encode IntervalTime=32768: want BER constraint error, got %v", err)
	}
}

func TestAreaEventInfoOccurrenceInfoOutOfRangeRejected(t *testing.T) {
	bad := OccurrenceInfo(99)
	_, err := convertAreaEventInfoToWire(&AreaEventInfo{
		AreaDefinition: AreaDefinition{
			AreaList: AreaList{{AreaType: AreaTypeCountryCode, AreaIdentification: HexBytes{0x01, 0x02}}},
		},
		OccurrenceInfo: &bad,
	})
	if !errors.Is(err, ErrOccurrenceInfoInvalid) {
		t.Errorf("want ErrOccurrenceInfoInvalid, got %v", err)
	}
}

// ============================================================================
// PeriodicLDRInfo
// ============================================================================

func TestPeriodicLDRInfoRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *PeriodicLDRInfo
	}{
		{"minimum values", &PeriodicLDRInfo{ReportingAmount: 1, ReportingInterval: 1}},
		{"typical", &PeriodicLDRInfo{ReportingAmount: 10, ReportingInterval: 60}},
		{"product cap boundary", &PeriodicLDRInfo{
			ReportingAmount:   PeriodicLDRProductMax,
			ReportingInterval: 1,
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertPeriodicLDRInfoToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := convertWireToPeriodicLDRInfo(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", tc.in, out)
			}
		})
	}
}

func TestPeriodicLDRInfoOutOfRangeRejected(t *testing.T) {
	cases := []struct {
		name string
		in   *PeriodicLDRInfo
		path string
	}{
		{"amount below min", &PeriodicLDRInfo{ReportingAmount: 0, ReportingInterval: 1}, "reportingAmount"},
		{"amount above max", &PeriodicLDRInfo{ReportingAmount: 8639999 + 1, ReportingInterval: 1}, "reportingAmount"},
		{"interval below min", &PeriodicLDRInfo{ReportingAmount: 1, ReportingInterval: 0}, "reportingInterval"},
		{"interval above max", &PeriodicLDRInfo{ReportingAmount: 1, ReportingInterval: 8639999 + 1}, "reportingInterval"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire := &gsm_map.PeriodicLDRInfo{ReportingAmount: gsm_map.ReportingAmount(tc.in.ReportingAmount), ReportingInterval: gsm_map.ReportingInterval(tc.in.ReportingInterval)}
			_, err := wire.MarshalBER()
			wantConstraintError(t, err, tc.path, "(1..8639999)")
		})
	}
}

// Spec-mandated cap: ReportingAmount × ReportingInterval ≤ 8639999
// (TS 29.002 MAP-LCS-DataTypes.asn:375-376).
func TestPeriodicLDRInfoProductCapRejected(t *testing.T) {
	in := &PeriodicLDRInfo{ReportingAmount: 1000, ReportingInterval: 10000} // 10,000,000 > cap
	_, err := convertPeriodicLDRInfoToWire(in)
	if !errors.Is(err, ErrPeriodicLDRProductExceeded) {
		t.Errorf("encode: want ErrPeriodicLDRProductExceeded, got %v", err)
	}

	w := &gsm_map.PeriodicLDRInfo{ReportingAmount: 1000, ReportingInterval: 10000}
	_, err = convertWireToPeriodicLDRInfo(w)
	if !errors.Is(err, ErrPeriodicLDRProductExceeded) {
		t.Errorf("decode: want ErrPeriodicLDRProductExceeded, got %v", err)
	}
}

// ============================================================================
// ReportingPLMN
// ============================================================================

func TestReportingPLMNRoundTrip(t *testing.T) {
	tech := RANTechnologyUmts
	cases := []struct {
		name string
		in   *ReportingPLMN
	}{
		{"plmnId only", &ReportingPLMN{
			PlmnId: HexBytes{0x32, 0xf4, 0x10},
		}},
		{"with tech", &ReportingPLMN{
			PlmnId:        HexBytes{0x32, 0xf4, 0x10},
			RanTechnology: &tech,
		}},
		{"with periodic support", &ReportingPLMN{
			PlmnId:                     HexBytes{0x32, 0xf4, 0x10},
			RanPeriodicLocationSupport: true,
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertReportingPLMNToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := convertWireToReportingPLMN(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", tc.in, out)
			}
		})
	}
}

func TestReportingPLMNInvalidPlmnIdRejected(t *testing.T) {
	_, err := strictWire(convertReportingPLMNToWire(&ReportingPLMN{
		PlmnId: HexBytes{0x01, 0x02}, // too short (must be exactly 3)
	}))
	if !isConstraint(err) {
		t.Errorf("want BER constraint error, got %v", err)
	}
}

func TestReportingPLMNRanTechnologyOutOfRangeRejected(t *testing.T) {
	bad := RANTechnology(99)
	_, err := convertReportingPLMNToWire(&ReportingPLMN{
		PlmnId:        HexBytes{0x32, 0xf4, 0x10},
		RanTechnology: &bad,
	})
	if !errors.Is(err, ErrRANTechnologyInvalid) {
		t.Errorf("want ErrRANTechnologyInvalid, got %v", err)
	}
}

// ============================================================================
// ReportingPLMNList
// ============================================================================

func TestReportingPLMNListRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *ReportingPLMNList
	}{
		{"single entry", &ReportingPLMNList{
			PlmnList: PLMNList{{PlmnId: HexBytes{0x32, 0xf4, 0x10}}},
		}},
		{"prioritized + 2 entries", &ReportingPLMNList{
			PlmnListPrioritized: true,
			PlmnList: PLMNList{
				{PlmnId: HexBytes{0x32, 0xf4, 0x10}},
				{PlmnId: HexBytes{0x62, 0xf2, 0x20}},
			},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertReportingPLMNListToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := convertWireToReportingPLMNList(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", tc.in, out)
			}
		})
	}
}

func TestReportingPLMNListEmptyListRejected(t *testing.T) {
	_, err := strictWire(convertReportingPLMNListToWire(&ReportingPLMNList{
		PlmnList: PLMNList{},
	}))
	if !isConstraint(err) {
		t.Errorf("want BER constraint error for empty list, got %v", err)
	}
}

func TestReportingPLMNListOversizedRejected(t *testing.T) {
	tooMany := make(PLMNList, 20+1)
	for i := range tooMany {
		tooMany[i] = ReportingPLMN{PlmnId: HexBytes{0x32, 0xf4, 0x10}}
	}
	_, err := strictWire(convertReportingPLMNListToWire(&ReportingPLMNList{PlmnList: tooMany}))
	if !isConstraint(err) {
		t.Errorf("want BER constraint error for 21 entries, got %v", err)
	}
}

func TestPSLAreaPeriodicNilPassThrough(t *testing.T) {
	if w, err := convertAreaToWire(nil); err != nil || w != nil {
		t.Errorf("AreaToWire nil: got w=%v err=%v", w, err)
	}
	if w, err := convertAreaDefinitionToWire(nil); err != nil || w != nil {
		t.Errorf("AreaDefinitionToWire nil: got w=%v err=%v", w, err)
	}
	if w, err := convertAreaEventInfoToWire(nil); err != nil || w != nil {
		t.Errorf("AreaEventInfoToWire nil: got w=%v err=%v", w, err)
	}
	if w, err := convertPeriodicLDRInfoToWire(nil); err != nil || w != nil {
		t.Errorf("PeriodicLDRInfoToWire nil: got w=%v err=%v", w, err)
	}
	if w, err := convertReportingPLMNToWire(nil); err != nil || w != nil {
		t.Errorf("ReportingPLMNToWire nil: got w=%v err=%v", w, err)
	}
	if w, err := convertReportingPLMNListToWire(nil); err != nil || w != nil {
		t.Errorf("ReportingPLMNListToWire nil: got w=%v err=%v", w, err)
	}

	// Decode-side nil pass-through.
	if out, err := convertWireToArea(nil); err != nil || out != nil {
		t.Errorf("WireToArea nil: got out=%v err=%v", out, err)
	}
	if out, err := convertWireToAreaDefinition(nil); err != nil || out != nil {
		t.Errorf("WireToAreaDefinition nil: got out=%v err=%v", out, err)
	}
	if out, err := convertWireToAreaEventInfo(nil); err != nil || out != nil {
		t.Errorf("WireToAreaEventInfo nil: got out=%v err=%v", out, err)
	}
	if out, err := convertWireToPeriodicLDRInfo(nil); err != nil || out != nil {
		t.Errorf("WireToPeriodicLDRInfo nil: got out=%v err=%v", out, err)
	}
	if out, err := convertWireToReportingPLMN(nil); err != nil || out != nil {
		t.Errorf("WireToReportingPLMN nil: got out=%v err=%v", out, err)
	}
	if out, err := convertWireToReportingPLMNList(nil); err != nil || out != nil {
		t.Errorf("WireToReportingPLMNList nil: got out=%v err=%v", out, err)
	}
}

// Lock in the decoder-side leniency contract: extensible enums must
// preserve unknown values per Postel even though encoders are strict.
// (Symmetric encoder strict-rejection tests for these enums live in
// TestAreaOutOfRangeTypeRejected, TestAreaEventInfoOccurrenceInfoOutOfRangeRejected,
// and TestReportingPLMNRanTechnologyOutOfRangeRejected.)
func TestPSLAreaPeriodicDecoderLenientForExtensibleEnums(t *testing.T) {
	// AreaType — extensible (TS 29.002:337).
	w := &gsm_map.Area{
		AreaType:           gsm_map.AreaType(99),
		AreaIdentification: gsm_map.AreaIdentification{0x01, 0x02},
	}
	got, err := convertWireToArea(w)
	if err != nil {
		t.Fatalf("AreaType=99 decode: unexpected error %v", err)
	}
	if int64(got.AreaType) != 99 {
		t.Errorf("AreaType not preserved: want 99, got %d", got.AreaType)
	}

	// OccurrenceInfo — extensible (TS 29.002:361).
	occ := gsm_map.OccurrenceInfo(99)
	wAEI := &gsm_map.AreaEventInfo{
		AreaDefinition: gsm_map.AreaDefinition{
			AreaList: &gsm_map.AreaList{Values: []gsm_map.Area{
				{AreaType: gsm_map.AreaTypeCountryCode, AreaIdentification: gsm_map.AreaIdentification{0x01, 0x02}},
			}},
		},
		OccurrenceInfo: &occ,
	}
	gotAEI, err := convertWireToAreaEventInfo(wAEI)
	if err != nil {
		t.Fatalf("OccurrenceInfo=99 decode: unexpected error %v", err)
	}
	if gotAEI.OccurrenceInfo == nil || int64(*gotAEI.OccurrenceInfo) != 99 {
		t.Errorf("OccurrenceInfo not preserved: got %v", gotAEI.OccurrenceInfo)
	}

	// RANTechnology — extensible (TS 29.002:420).
	tech := gsm_map.RANTechnology(99)
	wRP := &gsm_map.ReportingPLMN{
		PlmnId:        gsm_map.PLMNId{0x32, 0xf4, 0x10},
		RanTechnology: &tech,
	}
	gotRP, err := convertWireToReportingPLMN(wRP)
	if err != nil {
		t.Fatalf("RanTechnology=99 decode: unexpected error %v", err)
	}
	if gotRP.RanTechnology == nil || int64(*gotRP.RanTechnology) != 99 {
		t.Errorf("RanTechnology not preserved: got %v", gotRP.RanTechnology)
	}
}
