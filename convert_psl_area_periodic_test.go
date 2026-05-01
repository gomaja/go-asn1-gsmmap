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
		{"plmnId max", &Area{
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
	_, err := convertAreaToWire(&Area{
		AreaType:           AreaTypeCountryCode,
		AreaIdentification: HexBytes{0x01}, // too small (min 2)
	})
	if !errors.Is(err, ErrAreaIdentificationSize) {
		t.Errorf("encode 1 octet: want ErrAreaIdentificationSize, got %v", err)
	}
	tooBig := make(HexBytes, 8) // too big (max 7)
	_, err = convertAreaToWire(&Area{
		AreaType:           AreaTypeCountryCode,
		AreaIdentification: tooBig,
	})
	if !errors.Is(err, ErrAreaIdentificationSize) {
		t.Errorf("encode 8 octets: want ErrAreaIdentificationSize, got %v", err)
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
	_, err := convertAreaListToWire(AreaList{})
	if !errors.Is(err, ErrAreaListSize) {
		t.Errorf("want ErrAreaListSize for empty list, got %v", err)
	}
}

func TestAreaListOversizedRejected(t *testing.T) {
	tooMany := make(AreaList, AreaListMaxEntries+1)
	for i := range tooMany {
		tooMany[i] = Area{AreaType: AreaTypeCountryCode, AreaIdentification: HexBytes{0x01, 0x02}}
	}
	_, err := convertAreaListToWire(tooMany)
	if !errors.Is(err, ErrAreaListSize) {
		t.Errorf("want ErrAreaListSize for 11 entries, got %v", err)
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
	_, err := convertAreaEventInfoToWire(&AreaEventInfo{
		AreaDefinition: AreaDefinition{
			AreaList: AreaList{{AreaType: AreaTypeCountryCode, AreaIdentification: HexBytes{0x01, 0x02}}},
		},
		IntervalTime: &bad,
	})
	if !errors.Is(err, ErrIntervalTimeOutOfRange) {
		t.Errorf("encode IntervalTime=0: want ErrIntervalTimeOutOfRange, got %v", err)
	}

	tooBig := IntervalTime(IntervalTimeMax + 1)
	_, err = convertAreaEventInfoToWire(&AreaEventInfo{
		AreaDefinition: AreaDefinition{
			AreaList: AreaList{{AreaType: AreaTypeCountryCode, AreaIdentification: HexBytes{0x01, 0x02}}},
		},
		IntervalTime: &tooBig,
	})
	if !errors.Is(err, ErrIntervalTimeOutOfRange) {
		t.Errorf("encode IntervalTime=32768: want ErrIntervalTimeOutOfRange, got %v", err)
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
		name    string
		in      *PeriodicLDRInfo
		wantErr error
	}{
		{"amount below min", &PeriodicLDRInfo{ReportingAmount: 0, ReportingInterval: 1}, ErrReportingAmountOutOfRange},
		{"amount above max", &PeriodicLDRInfo{ReportingAmount: ReportingAmountMax + 1, ReportingInterval: 1}, ErrReportingAmountOutOfRange},
		{"interval below min", &PeriodicLDRInfo{ReportingAmount: 1, ReportingInterval: 0}, ErrReportingIntervalOutOfRange},
		{"interval above max", &PeriodicLDRInfo{ReportingAmount: 1, ReportingInterval: ReportingIntervalMax + 1}, ErrReportingIntervalOutOfRange},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := convertPeriodicLDRInfoToWire(tc.in)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("want %v, got %v", tc.wantErr, err)
			}
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
	_, err := convertReportingPLMNToWire(&ReportingPLMN{
		PlmnId: HexBytes{0x01, 0x02}, // too short (must be exactly 3)
	})
	if !errors.Is(err, ErrPlmnIdInvalidSize) {
		t.Errorf("want ErrPlmnIdInvalidSize, got %v", err)
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
	_, err := convertReportingPLMNListToWire(&ReportingPLMNList{
		PlmnList: PLMNList{},
	})
	if !errors.Is(err, ErrPLMNListSize) {
		t.Errorf("want ErrPLMNListSize for empty list, got %v", err)
	}
}

func TestReportingPLMNListOversizedRejected(t *testing.T) {
	tooMany := make(PLMNList, PLMNListMaxEntries+1)
	for i := range tooMany {
		tooMany[i] = ReportingPLMN{PlmnId: HexBytes{0x32, 0xf4, 0x10}}
	}
	_, err := convertReportingPLMNListToWire(&ReportingPLMNList{PlmnList: tooMany})
	if !errors.Is(err, ErrPLMNListSize) {
		t.Errorf("want ErrPLMNListSize for 21 entries, got %v", err)
	}
}

func TestPSLAreaPeriodicNilPassThrough(t *testing.T) {
	if w, err := convertAreaToWire(nil); err != nil || w != nil {
		t.Errorf("Area nil: got w=%v err=%v", w, err)
	}
	if w, err := convertAreaDefinitionToWire(nil); err != nil || w != nil {
		t.Errorf("AreaDefinition nil: got w=%v err=%v", w, err)
	}
	if w, err := convertAreaEventInfoToWire(nil); err != nil || w != nil {
		t.Errorf("AreaEventInfo nil: got w=%v err=%v", w, err)
	}
	if w, err := convertPeriodicLDRInfoToWire(nil); err != nil || w != nil {
		t.Errorf("PeriodicLDRInfo nil: got w=%v err=%v", w, err)
	}
	if w, err := convertReportingPLMNToWire(nil); err != nil || w != nil {
		t.Errorf("ReportingPLMN nil: got w=%v err=%v", w, err)
	}
	if w, err := convertReportingPLMNListToWire(nil); err != nil || w != nil {
		t.Errorf("ReportingPLMNList nil: got w=%v err=%v", w, err)
	}
}
