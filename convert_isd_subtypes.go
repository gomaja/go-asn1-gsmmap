package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// MC-SS-Info — 3GPP TS 29.002 V19.1.0 §17.7.8
// ============================================================================

func convertMCSSInfoToWire(m *MCSSInfo) (*gsm_map.MCSSInfo, error) {
	if m == nil {
		return nil, nil
	}

	return &gsm_map.MCSSInfo{
		SsCode:   gsm_map.SSCode{byte(m.SsCode)},
		SsStatus: gsm_map.ExtSSStatus(m.SsStatus),
		NbrSB:    int64(m.NbrSB),
		NbrUser:  int64(m.NbrUser),
	}, nil
}

func convertWireToMCSSInfo(w *gsm_map.MCSSInfo) (*MCSSInfo, error) {
	if w == nil {
		return nil, nil
	}

	nbrSB := int(w.NbrSB)
	nbrUser := int(w.NbrUser)

	return &MCSSInfo{
		SsCode:   SsCode(w.SsCode[0]),
		SsStatus: HexBytes(w.SsStatus),
		NbrSB:    nbrSB,
		NbrUser:  nbrUser,
	}, nil
}

// ============================================================================
// CSG-SubscriptionData / CSG-SubscriptionDataList / VPLMN-CSG-SubscriptionDataList
// — 3GPP TS 29.002 V19.1.0 §17.7.1
// ============================================================================

func convertCSGSubscriptionDataToWire(c *CSGSubscriptionData) (*gsm_map.CSGSubscriptionData, error) {
	if c == nil {
		return nil, nil
	}
	// CSG-Id SIZE (27) is checked by the BER codec; bitStringToWire also
	// enforces the octet count (3GPP TS 29.002 V19.1.0 §17.7.8, X.690 §8.6.2).
	csgID, err := bitStringToWire("CSGSubscriptionData.CsgID", c.CsgID, c.CsgIDBits)
	if err != nil {
		return nil, err
	}
	out := &gsm_map.CSGSubscriptionData{
		CsgId: csgID,
	}
	if len(c.ExpirationDate) > 0 {
		t := gsm_map.Time(c.ExpirationDate)
		out.ExpirationDate = &t
	}
	if c.LipaAllowedAPNList != nil {
		out.LipaAllowedAPNList = &gsm_map.LIPAAllowedAPNList{Values: make([]gsm_map.APN, len(c.LipaAllowedAPNList))}
		for i, apn := range c.LipaAllowedAPNList {
			if err := validateAPN(apn, fmt.Sprintf("CSGSubscriptionData.LipaAllowedAPNList[%d]", i)); err != nil {
				return nil, err
			}
			out.LipaAllowedAPNList.Values[i] = gsm_map.APN(apn)
		}
	}
	if c.PlmnId != nil {
		p := gsm_map.PLMNId(c.PlmnId)
		out.PlmnId = &p
	}
	return out, nil
}

func convertWireToCSGSubscriptionData(w *gsm_map.CSGSubscriptionData) (*CSGSubscriptionData, error) {
	if w == nil {
		return nil, nil
	}

	out := &CSGSubscriptionData{
		CsgID:     HexBytes(append([]byte(nil), w.CsgId.Bytes...)),
		CsgIDBits: w.CsgId.BitLength,
	}
	if w.ExpirationDate != nil {
		out.ExpirationDate = HexBytes(*w.ExpirationDate)
	}
	if w.LipaAllowedAPNList != nil {
		out.LipaAllowedAPNList = make([]HexBytes, len(w.LipaAllowedAPNList.Values))
		for i, apn := range w.LipaAllowedAPNList.Values {
			if err := validateAPN(HexBytes(apn), fmt.Sprintf("CSGSubscriptionData.LipaAllowedAPNList[%d]", i)); err != nil {
				return nil, err
			}
			out.LipaAllowedAPNList[i] = HexBytes(apn)
		}
	}
	if w.PlmnId != nil {
		out.PlmnId = HexBytes(*w.PlmnId)
	}
	return out, nil
}

func convertCSGSubscriptionDataListToWire(list CSGSubscriptionDataList) (*gsm_map.CSGSubscriptionDataList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.CSGSubscriptionDataList{Values: make([]gsm_map.CSGSubscriptionData, len(list))}
	for i, csd := range list {
		w, err := convertCSGSubscriptionDataToWire(&csd)
		if err != nil {
			return nil, fmt.Errorf("CSGSubscriptionDataList[%d]: %w", i, err)
		}
		out.Values[i] = *w
	}
	return &out, nil
}

func convertWireToCSGSubscriptionDataList(w *gsm_map.CSGSubscriptionDataList) (CSGSubscriptionDataList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(CSGSubscriptionDataList, len(w.Values))
	for i, csd := range w.Values {
		c, err := convertWireToCSGSubscriptionData(&csd)
		if err != nil {
			return nil, fmt.Errorf("CSGSubscriptionDataList[%d]: %w", i, err)
		}
		out[i] = *c
	}
	return out, nil
}

func convertVPLMNCSGSubscriptionDataListToWire(list VPLMNCSGSubscriptionDataList) (*gsm_map.VPLMNCSGSubscriptionDataList, error) {
	if list == nil {
		return nil, nil
	}

	out := &gsm_map.VPLMNCSGSubscriptionDataList{Values: make([]gsm_map.CSGSubscriptionData, len(list))}
	for i, csd := range list {
		w, err := convertCSGSubscriptionDataToWire(&csd)
		if err != nil {
			return nil, fmt.Errorf("CSGSubscriptionDataList[%d]: %w", i, err)
		}
		out.Values[i] = *w
	}
	return out, nil
}

func convertWireToVPLMNCSGSubscriptionDataList(w *gsm_map.VPLMNCSGSubscriptionDataList) (VPLMNCSGSubscriptionDataList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(VPLMNCSGSubscriptionDataList, len(w.Values))
	for i, csd := range w.Values {
		c, err := convertWireToCSGSubscriptionData(&csd)
		if err != nil {
			return nil, fmt.Errorf("CSGSubscriptionDataList[%d]: %w", i, err)
		}
		out[i] = *c
	}
	return out, nil
}

// ============================================================================
// AdjacentAccessRestrictionData / AdjacentAccessRestrictionDataList
// — 3GPP TS 29.002 V19.1.0 §17.7.1
// ============================================================================

func convertAdjacentAccessRestrictionDataToWire(a *AdjacentAccessRestrictionData) (*gsm_map.AdjacentAccessRestrictionData, error) {
	if a == nil {
		return nil, nil
	}

	out := &gsm_map.AdjacentAccessRestrictionData{
		PlmnId:                gsm_map.PLMNId(a.PlmnId),
		AccessRestrictionData: convertAccessRestrictionDataToBitString(&a.AccessRestrictionData),
	}
	if a.ExtAccessRestrictionData != nil {
		bs := convertExtAccessRestrictionDataToBitString(a.ExtAccessRestrictionData)
		out.ExtAccessRestrictionData = &bs
	}
	return out, nil
}

func convertWireToAdjacentAccessRestrictionData(w *gsm_map.AdjacentAccessRestrictionData) (*AdjacentAccessRestrictionData, error) {
	if w == nil {
		return nil, nil
	}

	out := &AdjacentAccessRestrictionData{
		PlmnId:                HexBytes(w.PlmnId),
		AccessRestrictionData: *convertBitStringToAccessRestrictionData(w.AccessRestrictionData),
	}
	if w.ExtAccessRestrictionData != nil {
		out.ExtAccessRestrictionData = convertBitStringToExtAccessRestrictionData(*w.ExtAccessRestrictionData)
	}
	return out, nil
}

func convertAdjacentAccessRestrictionDataListToWire(list AdjacentAccessRestrictionDataList) (*gsm_map.AdjacentAccessRestrictionDataList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.AdjacentAccessRestrictionDataList{Values: make([]gsm_map.AdjacentAccessRestrictionData, len(list))}
	for i, a := range list {
		w, err := convertAdjacentAccessRestrictionDataToWire(&a)
		if err != nil {
			return nil, fmt.Errorf("AdjacentAccessRestrictionDataList[%d]: %w", i, err)
		}
		out.Values[i] = *w
	}
	return &out, nil
}

func convertWireToAdjacentAccessRestrictionDataList(w *gsm_map.AdjacentAccessRestrictionDataList) (AdjacentAccessRestrictionDataList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(AdjacentAccessRestrictionDataList, len(w.Values))
	for i, a := range w.Values {
		v, err := convertWireToAdjacentAccessRestrictionData(&a)
		if err != nil {
			return nil, fmt.Errorf("AdjacentAccessRestrictionDataList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

// ============================================================================
// IMSI-GroupId / IMSI-GroupIdList — 3GPP TS 29.002 V19.1.0 §17.7.1
// ============================================================================

func convertIMSIGroupIdToWire(g *IMSIGroupId) (*gsm_map.IMSIGroupId, error) {
	if g == nil {
		return nil, nil
	}

	return &gsm_map.IMSIGroupId{
		GroupServiceId: int64(g.GroupServiceID),
		PlmnId:         gsm_map.PLMNId(g.PlmnId),
		LocalGroupID:   gsm_map.LocalGroupID(g.LocalGroupID),
	}, nil
}

func convertWireToIMSIGroupId(w *gsm_map.IMSIGroupId) (*IMSIGroupId, error) {
	if w == nil {
		return nil, nil
	}

	return &IMSIGroupId{
		GroupServiceID: uint32(w.GroupServiceId),
		PlmnId:         HexBytes(w.PlmnId),
		LocalGroupID:   HexBytes(w.LocalGroupID),
	}, nil
}

func convertIMSIGroupIdListToWire(list IMSIGroupIdList) (*gsm_map.IMSIGroupIdList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.IMSIGroupIdList{Values: make([]gsm_map.IMSIGroupId, len(list))}
	for i, g := range list {
		w, err := convertIMSIGroupIdToWire(&g)
		if err != nil {
			return nil, fmt.Errorf("IMSIGroupIdList[%d]: %w", i, err)
		}
		out.Values[i] = *w
	}
	return &out, nil
}

func convertWireToIMSIGroupIdList(w *gsm_map.IMSIGroupIdList) (IMSIGroupIdList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(IMSIGroupIdList, len(w.Values))
	for i, g := range w.Values {
		v, err := convertWireToIMSIGroupId(&g)
		if err != nil {
			return nil, fmt.Errorf("IMSIGroupIdList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

// ============================================================================
// EDRX-Cycle-Length / EDRX-Cycle-Length-List
// — 3GPP TS 29.002 V19.1.0 §17.7.1
// ============================================================================

func convertEDRXCycleLengthToWire(e *EDRXCycleLength) (*gsm_map.EDRXCycleLength, error) {
	if e == nil {
		return nil, nil
	}
	if !isListedUsedRATType(e.RatType) {
		return nil, fmt.Errorf("EDRXCycleLength.RatType=%d: %w", e.RatType, ErrUsedRATTypeInvalid)
	}

	return &gsm_map.EDRXCycleLength{
		RatType:              e.RatType,
		EDRXCycleLengthValue: gsm_map.EDRXCycleLengthValue(e.EDRXCycleLengthValue),
	}, nil
}

func convertWireToEDRXCycleLength(w *gsm_map.EDRXCycleLength) (*EDRXCycleLength, error) {
	if w == nil {
		return nil, nil
	}

	// UsedRatType is an extensible enum: an unknown value is kept (3GPP TS
	// 29.002 V19.1.0 §17.1.4) and Marshal refuses it.
	return &EDRXCycleLength{
		RatType:              w.RatType,
		EDRXCycleLengthValue: HexBytes(w.EDRXCycleLengthValue),
	}, nil
}

func convertEDRXCycleLengthListToWire(list EDRXCycleLengthList) (*gsm_map.EDRXCycleLengthList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.EDRXCycleLengthList{Values: make([]gsm_map.EDRXCycleLength, len(list))}
	for i, e := range list {
		w, err := convertEDRXCycleLengthToWire(&e)
		if err != nil {
			return nil, fmt.Errorf("EDRXCycleLengthList[%d]: %w", i, err)
		}
		out.Values[i] = *w
	}
	return &out, nil
}

func convertWireToEDRXCycleLengthList(w *gsm_map.EDRXCycleLengthList) (EDRXCycleLengthList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(EDRXCycleLengthList, len(w.Values))
	for i, e := range w.Values {
		v, err := convertWireToEDRXCycleLength(&e)
		if err != nil {
			return nil, fmt.Errorf("EDRXCycleLengthList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

// ============================================================================
// Reset-Id-List — 3GPP TS 29.002 V19.1.0 §17.7.1
// Reset-Id is a leaf OCTET STRING (SIZE 1..4).
// ============================================================================

func convertResetIdListToWire(list ResetIdList) (*gsm_map.ResetIdList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.ResetIdList{Values: make([]gsm_map.ResetId, len(list))}
	for i, r := range list {
		if len(r) < 1 || len(r) > 4 {
			return nil, fmt.Errorf("ResetIdList[%d]: %w (got %d)", i, ErrResetIdInvalidSize, len(r))
		}
		out.Values[i] = gsm_map.ResetId(r)
	}
	return &out, nil
}

func convertWireToResetIdList(w *gsm_map.ResetIdList) (ResetIdList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(ResetIdList, len(w.Values))
	for i, r := range w.Values {
		if len(r) < 1 || len(r) > 4 {
			return nil, fmt.Errorf("ResetIdList[%d]: %w (got %d)", i, ErrResetIdInvalidSize, len(r))
		}
		out[i] = HexBytes(r)
	}
	return out, nil
}
