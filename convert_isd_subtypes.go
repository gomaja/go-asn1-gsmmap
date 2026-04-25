package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/runtime"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// MC-SS-Info — TS 29.002 MAP-CommonDataTypes.asn:627
// ============================================================================

func convertMCSSInfoToWire(m *MCSSInfo) (*gsm_map.MCSSInfo, error) {
	if m == nil {
		return nil, nil
	}
	if err := validateExtSSStatus(m.SsStatus, "MCSSInfo.SsStatus"); err != nil {
		return nil, err
	}
	if int64(m.NbrSB) < 2 || int64(m.NbrSB) > gsm_map.MaxNumOfMCBearers {
		return nil, fmt.Errorf("%w (got %d)", ErrMCSSInfoNbrSBOutOfRange, m.NbrSB)
	}
	if int64(m.NbrUser) < 1 || int64(m.NbrUser) > gsm_map.MaxNumOfMCBearers {
		return nil, fmt.Errorf("%w (got %d)", ErrMCSSInfoNbrUserOutOfRange, m.NbrUser)
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
	if err := validateExtSSStatus(HexBytes(w.SsStatus), "MCSSInfo.SsStatus"); err != nil {
		return nil, err
	}
	nbrSB, err := narrowInt64Range(w.NbrSB, 2, gsm_map.MaxNumOfMCBearers, "MCSSInfo.NbrSB")
	if err != nil {
		return nil, err
	}
	nbrUser, err := narrowInt64Range(w.NbrUser, 1, gsm_map.MaxNumOfMCBearers, "MCSSInfo.NbrUser")
	if err != nil {
		return nil, err
	}
	if len(w.SsCode) == 0 {
		return nil, ErrMCSSInfoMissingSsCode
	}
	return &MCSSInfo{
		SsCode:   SsCode(w.SsCode[0]),
		SsStatus: HexBytes(w.SsStatus),
		NbrSB:    nbrSB,
		NbrUser:  nbrUser,
	}, nil
}

// ============================================================================
// CSG-SubscriptionData / CSG-SubscriptionDataList / VPLMN-CSG-SubscriptionDataList
// — TS 29.002 MAP-MS-DataTypes.asn:1259-1274
// ============================================================================

func convertCSGSubscriptionDataToWire(c *CSGSubscriptionData) (*gsm_map.CSGSubscriptionData, error) {
	if c == nil {
		return nil, nil
	}
	// CSG-Id is exactly 27 bits → ceil(27/8) = 4 octets. Caller must set
	// the bit length explicitly; silent coercion of 0 has been removed
	// to prevent encode/decode round-trip mutation (0 → 27).
	if c.CsgIdBitLength != CSGIdBitLength || len(c.CsgId) != (CSGIdBitLength+7)/8 {
		return nil, fmt.Errorf("%w (got %d octets, %d bits)", ErrCSGIdInvalidSize, len(c.CsgId), c.CsgIdBitLength)
	}
	if c.PlmnId != nil {
		if err := validatePlmnId(c.PlmnId, "CSGSubscriptionData.PlmnId"); err != nil {
			return nil, err
		}
	}
	out := &gsm_map.CSGSubscriptionData{
		CsgId: runtime.BitString{Bytes: append([]byte(nil), c.CsgId...), BitLength: CSGIdBitLength},
	}
	if len(c.ExpirationDate) > 0 {
		t := gsm_map.Time(c.ExpirationDate)
		out.ExpirationDate = &t
	}
	if c.LipaAllowedAPNList != nil {
		if len(c.LipaAllowedAPNList) < 1 || int64(len(c.LipaAllowedAPNList)) > gsm_map.MaxNumOfLIPAAllowedAPN {
			return nil, fmt.Errorf("%w (got %d)", ErrLipaAllowedAPNListSize, len(c.LipaAllowedAPNList))
		}
		out.LipaAllowedAPNList = make(gsm_map.LIPAAllowedAPNList, len(c.LipaAllowedAPNList))
		for i, apn := range c.LipaAllowedAPNList {
			if err := validateAPN(apn, fmt.Sprintf("CSGSubscriptionData.LipaAllowedAPNList[%d]", i)); err != nil {
				return nil, err
			}
			out.LipaAllowedAPNList[i] = gsm_map.APN(apn)
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
	if w.CsgId.BitLength != CSGIdBitLength {
		return nil, fmt.Errorf("%w (got %d bits)", ErrCSGIdInvalidSize, w.CsgId.BitLength)
	}
	out := &CSGSubscriptionData{
		CsgId:          HexBytes(append([]byte(nil), w.CsgId.Bytes...)),
		CsgIdBitLength: w.CsgId.BitLength,
	}
	if w.ExpirationDate != nil {
		out.ExpirationDate = HexBytes(*w.ExpirationDate)
	}
	if w.LipaAllowedAPNList != nil {
		if len(w.LipaAllowedAPNList) < 1 || int64(len(w.LipaAllowedAPNList)) > gsm_map.MaxNumOfLIPAAllowedAPN {
			return nil, fmt.Errorf("%w (got %d)", ErrLipaAllowedAPNListSize, len(w.LipaAllowedAPNList))
		}
		out.LipaAllowedAPNList = make([]HexBytes, len(w.LipaAllowedAPNList))
		for i, apn := range w.LipaAllowedAPNList {
			if err := validateAPN(HexBytes(apn), fmt.Sprintf("CSGSubscriptionData.LipaAllowedAPNList[%d]", i)); err != nil {
				return nil, err
			}
			out.LipaAllowedAPNList[i] = HexBytes(apn)
		}
	}
	if w.PlmnId != nil {
		if err := validatePlmnId(HexBytes(*w.PlmnId), "CSGSubscriptionData.PlmnId"); err != nil {
			return nil, err
		}
		out.PlmnId = HexBytes(*w.PlmnId)
	}
	return out, nil
}

func convertCSGSubscriptionDataListToWire(list CSGSubscriptionDataList) (gsm_map.CSGSubscriptionDataList, error) {
	if list == nil {
		return nil, nil
	}
	if len(list) < 1 || len(list) > MaxNumOfCSGSubscriptions {
		return nil, fmt.Errorf("%w (got %d)", ErrCSGSubscriptionDataListSize, len(list))
	}
	out := make(gsm_map.CSGSubscriptionDataList, len(list))
	for i, csd := range list {
		w, err := convertCSGSubscriptionDataToWire(&csd)
		if err != nil {
			return nil, fmt.Errorf("CSGSubscriptionDataList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToCSGSubscriptionDataList(w gsm_map.CSGSubscriptionDataList) (CSGSubscriptionDataList, error) {
	if w == nil {
		return nil, nil
	}
	if len(w) < 1 || len(w) > MaxNumOfCSGSubscriptions {
		return nil, fmt.Errorf("%w (got %d)", ErrCSGSubscriptionDataListSize, len(w))
	}
	out := make(CSGSubscriptionDataList, len(w))
	for i, csd := range w {
		c, err := convertWireToCSGSubscriptionData(&csd)
		if err != nil {
			return nil, fmt.Errorf("CSGSubscriptionDataList[%d]: %w", i, err)
		}
		out[i] = *c
	}
	return out, nil
}

// VPLMN-CSG-SubscriptionDataList shares the wire shape with CSG-SubscriptionDataList.
func convertVPLMNCSGSubscriptionDataListToWire(list VPLMNCSGSubscriptionDataList) (gsm_map.CSGSubscriptionDataList, error) {
	return convertCSGSubscriptionDataListToWire(CSGSubscriptionDataList(list))
}

func convertWireToVPLMNCSGSubscriptionDataList(w gsm_map.CSGSubscriptionDataList) (VPLMNCSGSubscriptionDataList, error) {
	out, err := convertWireToCSGSubscriptionDataList(w)
	if err != nil {
		return nil, err
	}
	return VPLMNCSGSubscriptionDataList(out), nil
}

// ============================================================================
// AdjacentAccessRestrictionData / AdjacentAccessRestrictionDataList
// — TS 29.002 MAP-MS-DataTypes.asn:1475-1483
// ============================================================================

func convertAdjacentAccessRestrictionDataToWire(a *AdjacentAccessRestrictionData) (*gsm_map.AdjacentAccessRestrictionData, error) {
	if a == nil {
		return nil, nil
	}
	if err := validatePlmnId(a.PlmnId, "AdjacentAccessRestrictionData.PlmnId"); err != nil {
		return nil, err
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
	if err := validatePlmnId(HexBytes(w.PlmnId), "AdjacentAccessRestrictionData.PlmnId"); err != nil {
		return nil, err
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

func convertAdjacentAccessRestrictionDataListToWire(list AdjacentAccessRestrictionDataList) (gsm_map.AdjacentAccessRestrictionDataList, error) {
	if list == nil {
		return nil, nil
	}
	if len(list) < 1 || len(list) > MaxNumOfAdjacentPLMN {
		return nil, fmt.Errorf("%w (got %d)", ErrAdjacentAccessRestrictionListSize, len(list))
	}
	out := make(gsm_map.AdjacentAccessRestrictionDataList, len(list))
	for i, a := range list {
		w, err := convertAdjacentAccessRestrictionDataToWire(&a)
		if err != nil {
			return nil, fmt.Errorf("AdjacentAccessRestrictionDataList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToAdjacentAccessRestrictionDataList(w gsm_map.AdjacentAccessRestrictionDataList) (AdjacentAccessRestrictionDataList, error) {
	if w == nil {
		return nil, nil
	}
	if len(w) < 1 || len(w) > MaxNumOfAdjacentPLMN {
		return nil, fmt.Errorf("%w (got %d)", ErrAdjacentAccessRestrictionListSize, len(w))
	}
	out := make(AdjacentAccessRestrictionDataList, len(w))
	for i, a := range w {
		v, err := convertWireToAdjacentAccessRestrictionData(&a)
		if err != nil {
			return nil, fmt.Errorf("AdjacentAccessRestrictionDataList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

// ============================================================================
// IMSI-GroupId / IMSI-GroupIdList — TS 29.002 MAP-MS-DataTypes.asn:1242-1252
// ============================================================================

func convertIMSIGroupIdToWire(g *IMSIGroupId) (*gsm_map.IMSIGroupId, error) {
	if g == nil {
		return nil, nil
	}
	if err := validatePlmnId(g.PlmnId, "IMSIGroupId.PlmnId"); err != nil {
		return nil, err
	}
	if len(g.LocalGroupID) < 1 || len(g.LocalGroupID) > 10 {
		return nil, fmt.Errorf("%w (got %d)", ErrLocalGroupIDInvalidSize, len(g.LocalGroupID))
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
	if err := validatePlmnId(HexBytes(w.PlmnId), "IMSIGroupId.PlmnId"); err != nil {
		return nil, err
	}
	if w.GroupServiceId < 0 || w.GroupServiceId > 0xFFFFFFFF {
		return nil, fmt.Errorf("%w (got %d)", ErrIMSIGroupServiceIDOverflow, w.GroupServiceId)
	}
	if len(w.LocalGroupID) < 1 || len(w.LocalGroupID) > 10 {
		return nil, fmt.Errorf("%w (got %d)", ErrLocalGroupIDInvalidSize, len(w.LocalGroupID))
	}
	return &IMSIGroupId{
		GroupServiceID: uint32(w.GroupServiceId),
		PlmnId:         HexBytes(w.PlmnId),
		LocalGroupID:   HexBytes(w.LocalGroupID),
	}, nil
}

func convertIMSIGroupIdListToWire(list IMSIGroupIdList) (gsm_map.IMSIGroupIdList, error) {
	if list == nil {
		return nil, nil
	}
	if len(list) < 1 || len(list) > MaxNumOfIMSIGroupId {
		return nil, fmt.Errorf("%w (got %d)", ErrIMSIGroupIdListSize, len(list))
	}
	out := make(gsm_map.IMSIGroupIdList, len(list))
	for i, g := range list {
		w, err := convertIMSIGroupIdToWire(&g)
		if err != nil {
			return nil, fmt.Errorf("IMSIGroupIdList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToIMSIGroupIdList(w gsm_map.IMSIGroupIdList) (IMSIGroupIdList, error) {
	if w == nil {
		return nil, nil
	}
	if len(w) < 1 || len(w) > MaxNumOfIMSIGroupId {
		return nil, fmt.Errorf("%w (got %d)", ErrIMSIGroupIdListSize, len(w))
	}
	out := make(IMSIGroupIdList, len(w))
	for i, g := range w {
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
// — TS 29.002 MAP-MS-DataTypes.asn:1207-1218
// ============================================================================

func convertEDRXCycleLengthToWire(e *EDRXCycleLength) (*gsm_map.EDRXCycleLength, error) {
	if e == nil {
		return nil, nil
	}
	if len(e.EDRXCycleLengthValue) != 1 {
		return nil, fmt.Errorf("%w (got %d)", ErrEDRXCycleLengthValueSize, len(e.EDRXCycleLengthValue))
	}
	return &gsm_map.EDRXCycleLength{
		RatType:              gsm_map.UsedRATType(e.RatType),
		EDRXCycleLengthValue: gsm_map.EDRXCycleLengthValue(e.EDRXCycleLengthValue),
	}, nil
}

func convertWireToEDRXCycleLength(w *gsm_map.EDRXCycleLength) (*EDRXCycleLength, error) {
	if w == nil {
		return nil, nil
	}
	if len(w.EDRXCycleLengthValue) != 1 {
		return nil, fmt.Errorf("%w (got %d)", ErrEDRXCycleLengthValueSize, len(w.EDRXCycleLengthValue))
	}
	// UsedRATType is extensible — preserve unknown values (Postel's law).
	return &EDRXCycleLength{
		RatType:              UsedRATType(w.RatType),
		EDRXCycleLengthValue: HexBytes(w.EDRXCycleLengthValue),
	}, nil
}

func convertEDRXCycleLengthListToWire(list EDRXCycleLengthList) (gsm_map.EDRXCycleLengthList, error) {
	if list == nil {
		return nil, nil
	}
	if len(list) < 1 || len(list) > MaxNumOfEDRXCycleLength {
		return nil, fmt.Errorf("%w (got %d)", ErrEDRXCycleLengthListSize, len(list))
	}
	out := make(gsm_map.EDRXCycleLengthList, len(list))
	for i, e := range list {
		w, err := convertEDRXCycleLengthToWire(&e)
		if err != nil {
			return nil, fmt.Errorf("EDRXCycleLengthList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToEDRXCycleLengthList(w gsm_map.EDRXCycleLengthList) (EDRXCycleLengthList, error) {
	if w == nil {
		return nil, nil
	}
	if len(w) < 1 || len(w) > MaxNumOfEDRXCycleLength {
		return nil, fmt.Errorf("%w (got %d)", ErrEDRXCycleLengthListSize, len(w))
	}
	out := make(EDRXCycleLengthList, len(w))
	for i, e := range w {
		v, err := convertWireToEDRXCycleLength(&e)
		if err != nil {
			return nil, fmt.Errorf("EDRXCycleLengthList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

// ============================================================================
// Reset-Id-List — TS 29.002 MAP-MS-DataTypes.asn:1223-1227
// Reset-Id is a leaf OCTET STRING (SIZE 1..4); only the list needs validation.
// ============================================================================

func convertResetIdListToWire(list ResetIdList) (gsm_map.ResetIdList, error) {
	if list == nil {
		return nil, nil
	}
	if len(list) < 1 || len(list) > MaxNumOfResetId {
		return nil, fmt.Errorf("%w (got %d)", ErrResetIdListSize, len(list))
	}
	out := make(gsm_map.ResetIdList, len(list))
	for i, r := range list {
		if len(r) < 1 || len(r) > MaxResetIdOctets {
			return nil, fmt.Errorf("ResetIdList[%d]: %w (got %d)", i, ErrResetIdInvalidSize, len(r))
		}
		out[i] = gsm_map.ResetId(r)
	}
	return out, nil
}

func convertWireToResetIdList(w gsm_map.ResetIdList) (ResetIdList, error) {
	if w == nil {
		return nil, nil
	}
	if len(w) < 1 || len(w) > MaxNumOfResetId {
		return nil, fmt.Errorf("%w (got %d)", ErrResetIdListSize, len(w))
	}
	out := make(ResetIdList, len(w))
	for i, r := range w {
		if len(r) < 1 || len(r) > MaxResetIdOctets {
			return nil, fmt.Errorf("ResetIdList[%d]: %w (got %d)", i, ErrResetIdInvalidSize, len(r))
		}
		out[i] = HexBytes(r)
	}
	return out, nil
}
