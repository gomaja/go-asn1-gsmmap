// convert_psl_area_periodic.go
//
// Converters for the PSL-Arg area-event tree, periodic LDR info, and
// reporting-PLMN list of ProvideSubscriberLocation (opCode 83).
//
// Container converters added:
//   - Area / AreaList / AreaDefinition / AreaEventInfo
//   - PeriodicLDRInfo (with the spec ReportingInterval × ReportingAmount
//     ≤ 8639999 product cap enforced)
//   - ReportingPLMN / PLMNList / ReportingPLMNList

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// Area — 3GPP TS 29.002 V19.1.0 §17.7.13
// ============================================================================

func convertAreaToWire(a *Area) (*gsm_map.Area, error) {
	if a == nil {
		return nil, nil
	}
	if int64(a.AreaType) < 0 || int64(a.AreaType) > 5 {
		return nil, fmt.Errorf("Area.AreaType=%d: %w", a.AreaType, ErrAreaTypeInvalid)
	}

	return &gsm_map.Area{
		AreaType:           a.AreaType,
		AreaIdentification: gsm_map.AreaIdentification(a.AreaIdentification),
	}, nil
}

func convertWireToArea(w *gsm_map.Area) *Area {
	if w == nil {
		return nil
	}

	// AreaType preserves unknown extensions (3GPP TS 29.002 V19.1.0 §17.1.4).
	return &Area{
		AreaType:           w.AreaType,
		AreaIdentification: HexBytes(w.AreaIdentification),
	}
}

// ============================================================================
// AreaList — 3GPP TS 29.002 V19.1.0 §17.7.13 (SIZE 1..maxNumOfAreas=10)
// ============================================================================

func convertAreaListToWire(list AreaList) (*gsm_map.AreaList, error) {
	out := gsm_map.AreaList{Values: make([]gsm_map.Area, 0, len(list))}
	for i := range list {
		w, err := convertAreaToWire(&list[i])
		if err != nil {
			return nil, fmt.Errorf("AreaList[%d]: %w", i, err)
		}
		out.Values = append(out.Values, *w)
	}
	return &out, nil
}

func convertWireToAreaList(w *gsm_map.AreaList) AreaList {
	if w == nil {
		w = &gsm_map.AreaList{}
	}

	out := make(AreaList, 0, len(w.Values))
	for i := range w.Values {
		out = append(out, *convertWireToArea(&w.Values[i]))
	}
	return out
}

// ============================================================================
// AreaDefinition — 3GPP TS 29.002 V19.1.0 §17.7.13
// ============================================================================

func convertAreaDefinitionToWire(d *AreaDefinition) (*gsm_map.AreaDefinition, error) {
	if d == nil {
		return nil, nil
	}
	list, err := convertAreaListToWire(d.AreaList)
	if err != nil {
		return nil, fmt.Errorf("AreaDefinition: %w", err)
	}
	return &gsm_map.AreaDefinition{AreaList: list}, nil
}

func convertWireToAreaDefinition(w *gsm_map.AreaDefinition) *AreaDefinition {
	if w == nil {
		return nil
	}
	return &AreaDefinition{AreaList: convertWireToAreaList(w.AreaList)}
}

// ============================================================================
// AreaEventInfo — 3GPP TS 29.002 V19.1.0 §17.7.13
// ============================================================================

func convertAreaEventInfoToWire(a *AreaEventInfo) (*gsm_map.AreaEventInfo, error) {
	if a == nil {
		return nil, nil
	}
	def, err := convertAreaDefinitionToWire(&a.AreaDefinition)
	if err != nil {
		return nil, fmt.Errorf("AreaEventInfo.AreaDefinition: %w", err)
	}
	out := &gsm_map.AreaEventInfo{AreaDefinition: *def}
	if a.OccurrenceInfo != nil {
		v := *a.OccurrenceInfo
		// OccurrenceInfo is extensible (3GPP TS 29.002 V19.1.0 §17.7.13); encoder
		// strict (0..1), decoder lenient.
		if int64(v) < 0 || int64(v) > 1 {
			return nil, fmt.Errorf("AreaEventInfo.OccurrenceInfo=%d: %w", v, ErrOccurrenceInfoInvalid)
		}
		out.OccurrenceInfo = &v
	}
	if a.IntervalTime != nil {
		v := *a.IntervalTime

		out.IntervalTime = &v
	}
	return out, nil
}

func convertWireToAreaEventInfo(w *gsm_map.AreaEventInfo) *AreaEventInfo {
	if w == nil {
		return nil
	}
	def := convertWireToAreaDefinition(&w.AreaDefinition)
	out := &AreaEventInfo{AreaDefinition: *def}
	if w.OccurrenceInfo != nil {
		v := *w.OccurrenceInfo
		out.OccurrenceInfo = &v
	}
	if w.IntervalTime != nil {
		v := *w.IntervalTime

		out.IntervalTime = &v
	}
	return out
}

// ============================================================================
// PeriodicLDRInfo — 3GPP TS 29.002 V19.1.0 §17.7.13
// ============================================================================
//
// 3GPP TS 29.002 V19.1.0 §17.7.13: "reportingInterval x reportingAmount
// shall not exceed 8639999 (99 days, 23 hours, 59 minutes and 59 seconds)"
// for compatibility with OMA MLP and RLP.

func convertPeriodicLDRInfoToWire(p *PeriodicLDRInfo) (*gsm_map.PeriodicLDRInfo, error) {
	if p == nil {
		return nil, nil
	}

	if p.ReportingAmount*p.ReportingInterval > PeriodicLDRProductMax {
		return nil, fmt.Errorf("PeriodicLDRInfo: ReportingAmount(%d) × ReportingInterval(%d) = %d: %w",
			p.ReportingAmount, p.ReportingInterval, p.ReportingAmount*p.ReportingInterval, ErrPeriodicLDRProductExceeded)
	}
	return &gsm_map.PeriodicLDRInfo{
		ReportingAmount:   p.ReportingAmount,
		ReportingInterval: p.ReportingInterval,
	}, nil
}

func convertWireToPeriodicLDRInfo(w *gsm_map.PeriodicLDRInfo) (*PeriodicLDRInfo, error) {
	if w == nil {
		return nil, nil
	}

	if w.ReportingAmount*w.ReportingInterval > PeriodicLDRProductMax {
		return nil, fmt.Errorf("PeriodicLDRInfo: ReportingAmount(%d) × ReportingInterval(%d) = %d: %w",
			w.ReportingAmount, w.ReportingInterval, w.ReportingAmount*w.ReportingInterval, ErrPeriodicLDRProductExceeded)
	}
	return &PeriodicLDRInfo{
		ReportingAmount:   w.ReportingAmount,
		ReportingInterval: w.ReportingInterval,
	}, nil
}

// ============================================================================
// ReportingPLMN — 3GPP TS 29.002 V19.1.0 §17.7.13
// ============================================================================

func convertReportingPLMNToWire(r *ReportingPLMN) (*gsm_map.ReportingPLMN, error) {
	if r == nil {
		return nil, nil
	}

	out := &gsm_map.ReportingPLMN{
		PlmnId: gsm_map.PLMNId(r.PlmnId),
	}
	if r.RanTechnology != nil {
		v := *r.RanTechnology
		// RANTechnology is extensible (3GPP TS 29.002 V19.1.0 §17.7.13); encoder strict
		// (0..1), decoder lenient.
		if int64(v) < 0 || int64(v) > 1 {
			return nil, fmt.Errorf("ReportingPLMN.RanTechnology=%d: %w", v, ErrRANTechnologyInvalid)
		}
		out.RanTechnology = &v
	}
	out.RanPeriodicLocationSupport = boolToNullPtr(r.RanPeriodicLocationSupport)
	return out, nil
}

func convertWireToReportingPLMN(w *gsm_map.ReportingPLMN) *ReportingPLMN {
	if w == nil {
		return nil
	}

	out := &ReportingPLMN{
		PlmnId: HexBytes(w.PlmnId),
	}
	if w.RanTechnology != nil {
		v := *w.RanTechnology
		out.RanTechnology = &v
	}
	out.RanPeriodicLocationSupport = nullPtrToBool(w.RanPeriodicLocationSupport)
	return out
}

// ============================================================================
// PLMNList — 3GPP TS 29.002 V19.1.0 §17.7.13 (SIZE 1..maxNumOfReportingPLMN=20)
// ============================================================================

func convertPLMNListToWire(list PLMNList) (*gsm_map.PLMNList, error) {
	out := gsm_map.PLMNList{Values: make([]gsm_map.ReportingPLMN, 0, len(list))}
	for i := range list {
		w, err := convertReportingPLMNToWire(&list[i])
		if err != nil {
			return nil, fmt.Errorf("PLMNList[%d]: %w", i, err)
		}
		out.Values = append(out.Values, *w)
	}
	return &out, nil
}

func convertWireToPLMNList(w *gsm_map.PLMNList) PLMNList {
	if w == nil {
		w = &gsm_map.PLMNList{}
	}

	out := make(PLMNList, 0, len(w.Values))
	for i := range w.Values {
		out = append(out, *convertWireToReportingPLMN(&w.Values[i]))
	}
	return out
}

// ============================================================================
// ReportingPLMNList — 3GPP TS 29.002 V19.1.0 §17.7.13
// ============================================================================

func convertReportingPLMNListToWire(r *ReportingPLMNList) (*gsm_map.ReportingPLMNList, error) {
	if r == nil {
		return nil, nil
	}
	list, err := convertPLMNListToWire(r.PlmnList)
	if err != nil {
		return nil, fmt.Errorf("ReportingPLMNList: %w", err)
	}
	out := &gsm_map.ReportingPLMNList{
		PlmnList: list,
	}
	out.PlmnListPrioritized = boolToNullPtr(r.PlmnListPrioritized)
	return out, nil
}

func convertWireToReportingPLMNList(w *gsm_map.ReportingPLMNList) *ReportingPLMNList {
	if w == nil {
		return nil
	}
	return &ReportingPLMNList{
		PlmnListPrioritized: nullPtrToBool(w.PlmnListPrioritized),
		PlmnList:            convertWireToPLMNList(w.PlmnList),
	}
}
