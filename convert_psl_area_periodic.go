// convert_psl_area_periodic.go
//
// Converters for the PSL-Arg area-event tree, periodic LDR info, and
// reporting-PLMN list. PR D3 of the staged ProvideSubscriberLocation
// (opCode 83) implementation, building on PRs #43 (leaf converters +
// BIT STRING codecs) and #44 (LCS-Client identifier tree).
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
// Area — TS 29.002 MAP-LCS-DataTypes.asn:332
// ============================================================================

func convertAreaToWire(a *Area) (*gsm_map.Area, error) {
	if a == nil {
		return nil, nil
	}
	if int64(a.AreaType) < 0 || int64(a.AreaType) > 5 {
		return nil, fmt.Errorf("Area.AreaType=%d: %w", a.AreaType, ErrAreaTypeInvalid)
	}
	if len(a.AreaIdentification) < AreaIdentificationMinLen || len(a.AreaIdentification) > AreaIdentificationMaxLen {
		return nil, fmt.Errorf("Area.AreaIdentification len=%d: %w", len(a.AreaIdentification), ErrAreaIdentificationSize)
	}
	return &gsm_map.Area{
		AreaType:           a.AreaType,
		AreaIdentification: gsm_map.AreaIdentification(a.AreaIdentification),
	}, nil
}

func convertWireToArea(w *gsm_map.Area) (*Area, error) {
	if w == nil {
		return nil, nil
	}
	if len(w.AreaIdentification) < AreaIdentificationMinLen || len(w.AreaIdentification) > AreaIdentificationMaxLen {
		return nil, fmt.Errorf("Area.AreaIdentification len=%d: %w", len(w.AreaIdentification), ErrAreaIdentificationSize)
	}
	// AreaType is extensible (TS 29.002:337); decoder lenient,
	// preserving unknown values per Postel.
	return &Area{
		AreaType:           w.AreaType,
		AreaIdentification: HexBytes(w.AreaIdentification),
	}, nil
}

// ============================================================================
// AreaList — TS 29.002 MAP-LCS-DataTypes.asn:328 (SIZE 1..maxNumOfAreas=10)
// ============================================================================

func convertAreaListToWire(list AreaList) (gsm_map.AreaList, error) {
	if len(list) < AreaListMinEntries || len(list) > AreaListMaxEntries {
		return nil, fmt.Errorf("AreaList len=%d: %w", len(list), ErrAreaListSize)
	}
	out := make(gsm_map.AreaList, 0, len(list))
	for i := range list {
		w, err := convertAreaToWire(&list[i])
		if err != nil {
			return nil, fmt.Errorf("AreaList[%d]: %w", i, err)
		}
		out = append(out, *w)
	}
	return out, nil
}

func convertWireToAreaList(w gsm_map.AreaList) (AreaList, error) {
	if len(w) < AreaListMinEntries || len(w) > AreaListMaxEntries {
		return nil, fmt.Errorf("AreaList len=%d: %w", len(w), ErrAreaListSize)
	}
	out := make(AreaList, 0, len(w))
	for i := range w {
		area, err := convertWireToArea(&w[i])
		if err != nil {
			return nil, fmt.Errorf("AreaList[%d]: %w", i, err)
		}
		out = append(out, *area)
	}
	return out, nil
}

// ============================================================================
// AreaDefinition — TS 29.002 MAP-LCS-DataTypes.asn:324
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

func convertWireToAreaDefinition(w *gsm_map.AreaDefinition) (*AreaDefinition, error) {
	if w == nil {
		return nil, nil
	}
	list, err := convertWireToAreaList(w.AreaList)
	if err != nil {
		return nil, fmt.Errorf("AreaDefinition: %w", err)
	}
	return &AreaDefinition{AreaList: list}, nil
}

// ============================================================================
// AreaEventInfo — TS 29.002 MAP-LCS-DataTypes.asn:318
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
		// OccurrenceInfo is extensible (TS 29.002:361); encoder
		// strict (0..1), decoder lenient.
		if int64(v) < 0 || int64(v) > 1 {
			return nil, fmt.Errorf("AreaEventInfo.OccurrenceInfo=%d: %w", v, ErrOccurrenceInfoInvalid)
		}
		out.OccurrenceInfo = &v
	}
	if a.IntervalTime != nil {
		v := *a.IntervalTime
		if v < IntervalTimeMin || v > IntervalTimeMax {
			return nil, fmt.Errorf("AreaEventInfo.IntervalTime=%d: %w", v, ErrIntervalTimeOutOfRange)
		}
		out.IntervalTime = &v
	}
	return out, nil
}

func convertWireToAreaEventInfo(w *gsm_map.AreaEventInfo) (*AreaEventInfo, error) {
	if w == nil {
		return nil, nil
	}
	def, err := convertWireToAreaDefinition(&w.AreaDefinition)
	if err != nil {
		return nil, fmt.Errorf("AreaEventInfo.AreaDefinition: %w", err)
	}
	out := &AreaEventInfo{AreaDefinition: *def}
	if w.OccurrenceInfo != nil {
		v := *w.OccurrenceInfo
		out.OccurrenceInfo = &v
	}
	if w.IntervalTime != nil {
		v := *w.IntervalTime
		if v < IntervalTimeMin || v > IntervalTimeMax {
			return nil, fmt.Errorf("AreaEventInfo.IntervalTime=%d: %w", v, ErrIntervalTimeOutOfRange)
		}
		out.IntervalTime = &v
	}
	return out, nil
}

// ============================================================================
// PeriodicLDRInfo — TS 29.002 MAP-LCS-DataTypes.asn:369
// ============================================================================
//
// Per spec at lines 375-376: ReportingInterval × ReportingAmount must
// not exceed 8639999 (99 days, 23 hours, 59 minutes, 59 seconds) for
// compatibility with OMA MLP and RLP.

func convertPeriodicLDRInfoToWire(p *PeriodicLDRInfo) (*gsm_map.PeriodicLDRInfo, error) {
	if p == nil {
		return nil, nil
	}
	if p.ReportingAmount < ReportingAmountMin || p.ReportingAmount > ReportingAmountMax {
		return nil, fmt.Errorf("PeriodicLDRInfo.ReportingAmount=%d: %w", p.ReportingAmount, ErrReportingAmountOutOfRange)
	}
	if p.ReportingInterval < ReportingIntervalMin || p.ReportingInterval > ReportingIntervalMax {
		return nil, fmt.Errorf("PeriodicLDRInfo.ReportingInterval=%d: %w", p.ReportingInterval, ErrReportingIntervalOutOfRange)
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
	if w.ReportingAmount < ReportingAmountMin || w.ReportingAmount > ReportingAmountMax {
		return nil, fmt.Errorf("PeriodicLDRInfo.ReportingAmount=%d: %w", w.ReportingAmount, ErrReportingAmountOutOfRange)
	}
	if w.ReportingInterval < ReportingIntervalMin || w.ReportingInterval > ReportingIntervalMax {
		return nil, fmt.Errorf("PeriodicLDRInfo.ReportingInterval=%d: %w", w.ReportingInterval, ErrReportingIntervalOutOfRange)
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
// ReportingPLMN — TS 29.002 MAP-LCS-DataTypes.asn:414
// ============================================================================

func convertReportingPLMNToWire(r *ReportingPLMN) (*gsm_map.ReportingPLMN, error) {
	if r == nil {
		return nil, nil
	}
	if err := validatePlmnId(r.PlmnId, "ReportingPLMN.PlmnId"); err != nil {
		return nil, err
	}
	out := &gsm_map.ReportingPLMN{
		PlmnId: gsm_map.PLMNId(r.PlmnId),
	}
	if r.RanTechnology != nil {
		v := *r.RanTechnology
		// RANTechnology is extensible (TS 29.002:420); encoder strict
		// (0..1), decoder lenient.
		if int64(v) < 0 || int64(v) > 1 {
			return nil, fmt.Errorf("ReportingPLMN.RanTechnology=%d: %w", v, ErrRANTechnologyInvalid)
		}
		out.RanTechnology = &v
	}
	out.RanPeriodicLocationSupport = boolToNullPtr(r.RanPeriodicLocationSupport)
	return out, nil
}

func convertWireToReportingPLMN(w *gsm_map.ReportingPLMN) (*ReportingPLMN, error) {
	if w == nil {
		return nil, nil
	}
	if err := validatePlmnId(HexBytes(w.PlmnId), "ReportingPLMN.PlmnId"); err != nil {
		return nil, err
	}
	out := &ReportingPLMN{
		PlmnId: HexBytes(w.PlmnId),
	}
	if w.RanTechnology != nil {
		v := *w.RanTechnology
		out.RanTechnology = &v
	}
	out.RanPeriodicLocationSupport = nullPtrToBool(w.RanPeriodicLocationSupport)
	return out, nil
}

// ============================================================================
// PLMNList — TS 29.002 MAP-LCS-DataTypes.asn:409 (SIZE 1..maxNumOfReportingPLMN=20)
// ============================================================================

func convertPLMNListToWire(list PLMNList) (gsm_map.PLMNList, error) {
	if len(list) < PLMNListMinEntries || len(list) > PLMNListMaxEntries {
		return nil, fmt.Errorf("PLMNList len=%d: %w", len(list), ErrPLMNListSize)
	}
	out := make(gsm_map.PLMNList, 0, len(list))
	for i := range list {
		w, err := convertReportingPLMNToWire(&list[i])
		if err != nil {
			return nil, fmt.Errorf("PLMNList[%d]: %w", i, err)
		}
		out = append(out, *w)
	}
	return out, nil
}

func convertWireToPLMNList(w gsm_map.PLMNList) (PLMNList, error) {
	if len(w) < PLMNListMinEntries || len(w) > PLMNListMaxEntries {
		return nil, fmt.Errorf("PLMNList len=%d: %w", len(w), ErrPLMNListSize)
	}
	out := make(PLMNList, 0, len(w))
	for i := range w {
		plmn, err := convertWireToReportingPLMN(&w[i])
		if err != nil {
			return nil, fmt.Errorf("PLMNList[%d]: %w", i, err)
		}
		out = append(out, *plmn)
	}
	return out, nil
}

// ============================================================================
// ReportingPLMNList — TS 29.002 MAP-LCS-DataTypes.asn:404
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

func convertWireToReportingPLMNList(w *gsm_map.ReportingPLMNList) (*ReportingPLMNList, error) {
	if w == nil {
		return nil, nil
	}
	list, err := convertWireToPLMNList(w.PlmnList)
	if err != nil {
		return nil, fmt.Errorf("ReportingPLMNList: %w", err)
	}
	return &ReportingPLMNList{
		PlmnListPrioritized: nullPtrToBool(w.PlmnListPrioritized),
		PlmnList:            list,
	}, nil
}
