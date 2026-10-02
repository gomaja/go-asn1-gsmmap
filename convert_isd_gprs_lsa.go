package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// validateExtQoSHierarchy enforces the spec hierarchy from
// MAP-MS-DataTypes.asn:1534-1538: Ext2 requires Ext, Ext3 requires
// Ext2, Ext4 requires Ext3. Each parameter is true when the
// corresponding Ext{N}-QoS-Subscribed field is present.
func validateExtQoSHierarchy(ext, ext2, ext3, ext4 bool) error {
	if ext2 && !ext {
		return fmt.Errorf("%w: Ext2 set without Ext", ErrExtQoSHierarchyViolated)
	}
	if ext3 && !ext2 {
		return fmt.Errorf("%w: Ext3 set without Ext2", ErrExtQoSHierarchyViolated)
	}
	if ext4 && !ext3 {
		return fmt.Errorf("%w: Ext4 set without Ext3", ErrExtQoSHierarchyViolated)
	}
	return nil
}

// ============================================================================
// AMBR — TS 29.002 MAP-MS-DataTypes.asn:1386
// ============================================================================

func convertAMBRToWire(a *AMBR) (*gsm_map.AMBR, error) {
	if a == nil {
		return nil, nil
	}
	if a.MaxRequestedBandwidthUL < 0 || a.MaxRequestedBandwidthDL < 0 {
		return nil, fmt.Errorf("%w (UL=%d, DL=%d)", ErrAMBRBandwidthOutOfRange, a.MaxRequestedBandwidthUL, a.MaxRequestedBandwidthDL)
	}
	out := &gsm_map.AMBR{
		MaxRequestedBandwidthUL: bigIntFromInt64(a.MaxRequestedBandwidthUL),
		MaxRequestedBandwidthDL: bigIntFromInt64(a.MaxRequestedBandwidthDL),
	}
	if a.ExtendedMaxRequestedBandwidthUL != nil {
		if *a.ExtendedMaxRequestedBandwidthUL < 0 {
			return nil, fmt.Errorf("%w (extended UL=%d)", ErrAMBRBandwidthOutOfRange, *a.ExtendedMaxRequestedBandwidthUL)
		}
		out.ExtendedMaxRequestedBandwidthUL = bigIntFromInt64(*a.ExtendedMaxRequestedBandwidthUL)
	}
	if a.ExtendedMaxRequestedBandwidthDL != nil {
		if *a.ExtendedMaxRequestedBandwidthDL < 0 {
			return nil, fmt.Errorf("%w (extended DL=%d)", ErrAMBRBandwidthOutOfRange, *a.ExtendedMaxRequestedBandwidthDL)
		}
		out.ExtendedMaxRequestedBandwidthDL = bigIntFromInt64(*a.ExtendedMaxRequestedBandwidthDL)
	}
	return out, nil
}

func convertWireToAMBR(w *gsm_map.AMBR) (*AMBR, error) {
	if w == nil {
		return nil, nil
	}
	ul, err := int64FromBigInt(w.MaxRequestedBandwidthUL, "AMBR.MaxRequestedBandwidthUL")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAMBRBandwidthOutOfRange, err)
	}
	dl, err := int64FromBigInt(w.MaxRequestedBandwidthDL, "AMBR.MaxRequestedBandwidthDL")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAMBRBandwidthOutOfRange, err)
	}
	if ul < 0 || dl < 0 {
		return nil, fmt.Errorf("%w (UL=%d, DL=%d)", ErrAMBRBandwidthOutOfRange, ul, dl)
	}
	out := &AMBR{
		MaxRequestedBandwidthUL: ul,
		MaxRequestedBandwidthDL: dl,
	}
	if w.ExtendedMaxRequestedBandwidthUL != nil {
		v, err := int64FromBigInt(w.ExtendedMaxRequestedBandwidthUL, "AMBR.ExtendedMaxRequestedBandwidthUL")
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrAMBRBandwidthOutOfRange, err)
		}
		if v < 0 {
			return nil, fmt.Errorf("%w (extended UL=%d)", ErrAMBRBandwidthOutOfRange, v)
		}
		out.ExtendedMaxRequestedBandwidthUL = &v
	}
	if w.ExtendedMaxRequestedBandwidthDL != nil {
		v, err := int64FromBigInt(w.ExtendedMaxRequestedBandwidthDL, "AMBR.ExtendedMaxRequestedBandwidthDL")
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrAMBRBandwidthOutOfRange, err)
		}
		if v < 0 {
			return nil, fmt.Errorf("%w (extended DL=%d)", ErrAMBRBandwidthOutOfRange, v)
		}
		out.ExtendedMaxRequestedBandwidthDL = &v
	}
	return out, nil
}

// ============================================================================
// PDP-Context — TS 29.002 MAP-MS-DataTypes.asn:1522
// ============================================================================

func convertPDPContextToWire(p *PDPContext) (*gsm_map.PDPContext, error) {
	if p == nil {
		return nil, nil
	}

	if err := validateExtQoSHierarchy(
		p.ExtQoSSubscribed != nil,
		p.Ext2QoSSubscribed != nil,
		p.Ext3QoSSubscribed != nil,
		p.Ext4QoSSubscribed != nil,
	); err != nil {
		return nil, err
	}

	if p.ExtPdpAddress != nil {
		if p.PdpAddress == nil {
			return nil, ErrExtPDPAddressWithoutPDPAddress
		}
	}

	if p.SiptoPermission != nil {
		if *p.SiptoPermission < 0 || *p.SiptoPermission > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrSIPTOPermissionInvalid, *p.SiptoPermission)
		}
	}
	if p.SiptoLocalNetworkPermission != nil {
		if *p.SiptoLocalNetworkPermission < 0 || *p.SiptoLocalNetworkPermission > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrSIPTOLocalNetworkPermissionInvalid, *p.SiptoLocalNetworkPermission)
		}
	}
	if p.LipaPermission != nil {
		if *p.LipaPermission < 0 || *p.LipaPermission > 2 {
			return nil, fmt.Errorf("%w (got %d)", ErrLIPAPermissionInvalid, *p.LipaPermission)
		}
	}
	if p.NIDDMechanism != nil {
		if *p.NIDDMechanism < 0 || *p.NIDDMechanism > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrNIDDMechanismInvalid, *p.NIDDMechanism)
		}
	}

	out := &gsm_map.PDPContext{
		PdpContextId:  gsm_map.ContextId(p.PdpContextId),
		PdpType:       gsm_map.PDPType(p.PdpType),
		QosSubscribed: gsm_map.QoSSubscribed(p.QosSubscribed),
		Apn:           gsm_map.APN(p.Apn),
	}
	if p.PdpAddress != nil {
		v := gsm_map.PDPAddress(p.PdpAddress)
		out.PdpAddress = &v
	}
	out.VplmnAddressAllowed = boolToNullPtr(p.VplmnAddressAllowed)
	if p.ExtQoSSubscribed != nil {
		v := gsm_map.ExtQoSSubscribed(p.ExtQoSSubscribed)
		out.ExtQoSSubscribed = &v
	}
	if p.PdpChargingCharacteristics != nil {
		v := gsm_map.ChargingCharacteristics(p.PdpChargingCharacteristics)
		out.PdpChargingCharacteristics = &v
	}
	if p.Ext2QoSSubscribed != nil {
		v := gsm_map.Ext2QoSSubscribed(p.Ext2QoSSubscribed)
		out.Ext2QoSSubscribed = &v
	}
	if p.Ext3QoSSubscribed != nil {
		v := gsm_map.Ext3QoSSubscribed(p.Ext3QoSSubscribed)
		out.Ext3QoSSubscribed = &v
	}
	if p.Ext4QoSSubscribed != nil {
		v := gsm_map.Ext4QoSSubscribed(p.Ext4QoSSubscribed)
		out.Ext4QoSSubscribed = &v
	}
	if p.ApnOiReplacement != nil {
		v := gsm_map.APNOIReplacement(p.ApnOiReplacement)
		out.ApnOiReplacement = &v
	}
	if p.ExtPdpType != nil {
		v := gsm_map.ExtPDPType(p.ExtPdpType)
		out.ExtPdpType = &v
	}
	if p.ExtPdpAddress != nil {
		v := gsm_map.PDPAddress(p.ExtPdpAddress)
		out.ExtPdpAddress = &v
	}
	if p.Ambr != nil {
		ambr, err := convertAMBRToWire(p.Ambr)
		if err != nil {
			return nil, fmt.Errorf("PDPContext.Ambr: %w", err)
		}
		out.Ambr = ambr
	}
	if p.SiptoPermission != nil {
		v := gsm_map.SIPTOPermission(*p.SiptoPermission)
		out.SiptoPermission = &v
	}
	if p.LipaPermission != nil {
		v := gsm_map.LIPAPermission(*p.LipaPermission)
		out.LipaPermission = &v
	}
	if p.RestorationPriority != nil {
		v := gsm_map.RestorationPriority(p.RestorationPriority)
		out.RestorationPriority = &v
	}
	if p.SiptoLocalNetworkPermission != nil {
		v := gsm_map.SIPTOLocalNetworkPermission(*p.SiptoLocalNetworkPermission)
		out.SiptoLocalNetworkPermission = &v
	}
	if p.NIDDMechanism != nil {
		v := gsm_map.NIDDMechanism(*p.NIDDMechanism)
		out.NIDDMechanism = &v
	}
	if p.SCEFID != nil {
		v := gsm_map.FQDN(p.SCEFID)
		out.SCEFID = &v
	}
	return out, nil
}

func convertWireToPDPContext(w *gsm_map.PDPContext) (*PDPContext, error) {
	if w == nil {
		return nil, nil
	}
	id := int(w.PdpContextId)

	if err := validateExtQoSHierarchy(
		w.ExtQoSSubscribed != nil,
		w.Ext2QoSSubscribed != nil,
		w.Ext3QoSSubscribed != nil,
		w.Ext4QoSSubscribed != nil,
	); err != nil {
		return nil, err
	}
	out := &PDPContext{
		PdpContextId:        id,
		PdpType:             HexBytes(w.PdpType),
		QosSubscribed:       HexBytes(w.QosSubscribed),
		Apn:                 HexBytes(w.Apn),
		VplmnAddressAllowed: nullPtrToBool(w.VplmnAddressAllowed),
	}
	if w.PdpAddress != nil {
		out.PdpAddress = HexBytes(*w.PdpAddress)
	}
	if w.ExtQoSSubscribed != nil {
		out.ExtQoSSubscribed = HexBytes(*w.ExtQoSSubscribed)
	}
	if w.PdpChargingCharacteristics != nil {
		out.PdpChargingCharacteristics = HexBytes(*w.PdpChargingCharacteristics)
	}
	if w.Ext2QoSSubscribed != nil {
		out.Ext2QoSSubscribed = HexBytes(*w.Ext2QoSSubscribed)
	}
	if w.Ext3QoSSubscribed != nil {
		out.Ext3QoSSubscribed = HexBytes(*w.Ext3QoSSubscribed)
	}
	if w.Ext4QoSSubscribed != nil {
		out.Ext4QoSSubscribed = HexBytes(*w.Ext4QoSSubscribed)
	}
	if w.ApnOiReplacement != nil {
		out.ApnOiReplacement = HexBytes(*w.ApnOiReplacement)
	}
	if w.ExtPdpType != nil {
		out.ExtPdpType = HexBytes(*w.ExtPdpType)
	}
	if w.ExtPdpAddress != nil {
		if w.PdpAddress == nil {
			return nil, ErrExtPDPAddressWithoutPDPAddress
		}

		out.ExtPdpAddress = HexBytes(*w.ExtPdpAddress)
	}
	if w.Ambr != nil {
		ambr, err := convertWireToAMBR(w.Ambr)
		if err != nil {
			return nil, fmt.Errorf("PDPContext.Ambr: %w", err)
		}
		out.Ambr = ambr
	}
	if w.SiptoPermission != nil {
		v := SIPTOPermission(*w.SiptoPermission)
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrSIPTOPermissionInvalid, v)
		}
		out.SiptoPermission = &v
	}
	if w.LipaPermission != nil {
		v := LIPAPermission(*w.LipaPermission)
		if v < 0 || v > 2 {
			return nil, fmt.Errorf("%w (got %d)", ErrLIPAPermissionInvalid, v)
		}
		out.LipaPermission = &v
	}
	if w.RestorationPriority != nil {
		out.RestorationPriority = HexBytes(*w.RestorationPriority)
	}
	if w.SiptoLocalNetworkPermission != nil {
		v := SIPTOLocalNetworkPermission(*w.SiptoLocalNetworkPermission)
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrSIPTOLocalNetworkPermissionInvalid, v)
		}
		out.SiptoLocalNetworkPermission = &v
	}
	if w.NIDDMechanism != nil {
		v := NIDDMechanism(*w.NIDDMechanism)
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrNIDDMechanismInvalid, v)
		}
		out.NIDDMechanism = &v
	}
	if w.SCEFID != nil {
		out.SCEFID = HexBytes(*w.SCEFID)
	}
	return out, nil
}

// ============================================================================
// GPRSDataList / GPRSSubscriptionData — TS 29.002 MAP-MS-DataTypes.asn:1517-1595
// ============================================================================

func convertGPRSDataListToWire(list GPRSDataList) (*gsm_map.GPRSDataList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.GPRSDataList{Values: make([]gsm_map.PDPContext, len(list))}
	for i, p := range list {
		w, err := convertPDPContextToWire(&p)
		if err != nil {
			return nil, fmt.Errorf("GPRSDataList[%d]: %w", i, err)
		}
		out.Values[i] = *w
	}
	return &out, nil
}

func convertWireToGPRSDataList(w *gsm_map.GPRSDataList) (GPRSDataList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(GPRSDataList, len(w.Values))
	for i, p := range w.Values {
		v, err := convertWireToPDPContext(&p)
		if err != nil {
			return nil, fmt.Errorf("GPRSDataList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

func convertGPRSSubscriptionDataToWire(g *GPRSSubscriptionData) (*gsm_map.GPRSSubscriptionData, error) {
	if g == nil {
		return nil, nil
	}
	if g.GprsDataList == nil {
		return nil, ErrGPRSSubscriptionDataMissingList
	}
	gdl, err := convertGPRSDataListToWire(g.GprsDataList)
	if err != nil {
		return nil, err
	}
	out := &gsm_map.GPRSSubscriptionData{
		GprsDataList:             gdl,
		CompleteDataListIncluded: boolToNullPtr(g.CompleteDataListIncluded),
	}
	if g.ApnOiReplacement != nil {
		v := gsm_map.APNOIReplacement(g.ApnOiReplacement)
		out.ApnOiReplacement = &v
	}
	return out, nil
}

func convertWireToGPRSSubscriptionData(w *gsm_map.GPRSSubscriptionData) (*GPRSSubscriptionData, error) {
	if w == nil {
		return nil, nil
	}
	if w.GprsDataList == nil {
		return nil, ErrGPRSSubscriptionDataMissingList
	}
	gdl, err := convertWireToGPRSDataList(w.GprsDataList)
	if err != nil {
		return nil, err
	}
	out := &GPRSSubscriptionData{
		GprsDataList:             gdl,
		CompleteDataListIncluded: nullPtrToBool(w.CompleteDataListIncluded),
	}
	if w.ApnOiReplacement != nil {
		out.ApnOiReplacement = HexBytes(*w.ApnOiReplacement)
	}
	return out, nil
}

// ============================================================================
// LSAData / LSADataList / LSAInformation — TS 29.002
// MAP-MS-DataTypes.asn:1706-1726
// ============================================================================

func convertLSADataToWire(l *LSAData) (*gsm_map.LSAData, error) {
	if l == nil {
		return nil, nil
	}

	return &gsm_map.LSAData{
		LsaIdentity:            gsm_map.LSAIdentity(l.LsaIdentity),
		LsaAttributes:          gsm_map.LSAAttributes(l.LsaAttributes),
		LsaActiveModeIndicator: boolToNullPtr(l.LsaActiveModeIndicator),
	}, nil
}

func convertWireToLSAData(w *gsm_map.LSAData) (*LSAData, error) {
	if w == nil {
		return nil, nil
	}

	return &LSAData{
		LsaIdentity:            HexBytes(w.LsaIdentity),
		LsaAttributes:          HexBytes(w.LsaAttributes),
		LsaActiveModeIndicator: nullPtrToBool(w.LsaActiveModeIndicator),
	}, nil
}

func convertLSADataListToWire(list LSADataList) (*gsm_map.LSADataList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.LSADataList{Values: make([]gsm_map.LSAData, len(list))}
	for i, l := range list {
		w, err := convertLSADataToWire(&l)
		if err != nil {
			return nil, fmt.Errorf("LSADataList[%d]: %w", i, err)
		}
		out.Values[i] = *w
	}
	return &out, nil
}

func convertWireToLSADataList(w *gsm_map.LSADataList) (LSADataList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(LSADataList, len(w.Values))
	for i, l := range w.Values {
		v, err := convertWireToLSAData(&l)
		if err != nil {
			return nil, fmt.Errorf("LSADataList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

func convertLSAInformationToWire(l *LSAInformation) (*gsm_map.LSAInformation, error) {
	if l == nil {
		return nil, nil
	}
	if l.LsaOnlyAccessIndicator != nil {
		v := *l.LsaOnlyAccessIndicator
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrLSAOnlyAccessIndicatorInvalid, v)
		}
	}
	out := &gsm_map.LSAInformation{
		CompleteDataListIncluded: boolToNullPtr(l.CompleteDataListIncluded),
	}
	if l.LsaOnlyAccessIndicator != nil {
		v := gsm_map.LSAOnlyAccessIndicator(*l.LsaOnlyAccessIndicator)
		out.LsaOnlyAccessIndicator = &v
	}
	if l.LsaDataList != nil {
		ldl, err := convertLSADataListToWire(l.LsaDataList)
		if err != nil {
			return nil, err
		}
		out.LsaDataList = ldl
	}
	return out, nil
}

func convertWireToLSAInformation(w *gsm_map.LSAInformation) (*LSAInformation, error) {
	if w == nil {
		return nil, nil
	}
	out := &LSAInformation{
		CompleteDataListIncluded: nullPtrToBool(w.CompleteDataListIncluded),
	}
	if w.LsaOnlyAccessIndicator != nil {
		v := LSAOnlyAccessIndicator(*w.LsaOnlyAccessIndicator)
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrLSAOnlyAccessIndicatorInvalid, v)
		}
		out.LsaOnlyAccessIndicator = &v
	}
	if w.LsaDataList != nil {
		ldl, err := convertWireToLSADataList(w.LsaDataList)
		if err != nil {
			return nil, err
		}
		out.LsaDataList = ldl
	}
	return out, nil
}
