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
		MaxRequestedBandwidthUL: gsm_map.Bandwidth(a.MaxRequestedBandwidthUL),
		MaxRequestedBandwidthDL: gsm_map.Bandwidth(a.MaxRequestedBandwidthDL),
	}
	if a.ExtendedMaxRequestedBandwidthUL != nil {
		if *a.ExtendedMaxRequestedBandwidthUL < 0 {
			return nil, fmt.Errorf("%w (extended UL=%d)", ErrAMBRBandwidthOutOfRange, *a.ExtendedMaxRequestedBandwidthUL)
		}
		v := gsm_map.BandwidthExt(*a.ExtendedMaxRequestedBandwidthUL)
		out.ExtendedMaxRequestedBandwidthUL = &v
	}
	if a.ExtendedMaxRequestedBandwidthDL != nil {
		if *a.ExtendedMaxRequestedBandwidthDL < 0 {
			return nil, fmt.Errorf("%w (extended DL=%d)", ErrAMBRBandwidthOutOfRange, *a.ExtendedMaxRequestedBandwidthDL)
		}
		v := gsm_map.BandwidthExt(*a.ExtendedMaxRequestedBandwidthDL)
		out.ExtendedMaxRequestedBandwidthDL = &v
	}
	return out, nil
}

func convertWireToAMBR(w *gsm_map.AMBR) (*AMBR, error) {
	if w == nil {
		return nil, nil
	}
	if w.MaxRequestedBandwidthUL < 0 || w.MaxRequestedBandwidthDL < 0 {
		return nil, fmt.Errorf("%w (UL=%d, DL=%d)", ErrAMBRBandwidthOutOfRange, w.MaxRequestedBandwidthUL, w.MaxRequestedBandwidthDL)
	}
	out := &AMBR{
		MaxRequestedBandwidthUL: int64(w.MaxRequestedBandwidthUL),
		MaxRequestedBandwidthDL: int64(w.MaxRequestedBandwidthDL),
	}
	if w.ExtendedMaxRequestedBandwidthUL != nil {
		v := int64(*w.ExtendedMaxRequestedBandwidthUL)
		if v < 0 {
			return nil, fmt.Errorf("%w (extended UL=%d)", ErrAMBRBandwidthOutOfRange, v)
		}
		out.ExtendedMaxRequestedBandwidthUL = &v
	}
	if w.ExtendedMaxRequestedBandwidthDL != nil {
		v := int64(*w.ExtendedMaxRequestedBandwidthDL)
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
	if p.PdpContextId < 1 || int64(p.PdpContextId) > gsm_map.MaxNumOfPDPContexts {
		return nil, fmt.Errorf("%w (got %d)", ErrPDPContextIdOutOfRange, p.PdpContextId)
	}
	if len(p.PdpType) != 2 {
		return nil, fmt.Errorf("%w (got %d)", ErrPDPTypeInvalidSize, len(p.PdpType))
	}
	if len(p.QosSubscribed) != 3 {
		return nil, fmt.Errorf("%w (got %d)", ErrQoSSubscribedInvalidSize, len(p.QosSubscribed))
	}
	if err := validateAPN(p.Apn, "PDPContext.Apn"); err != nil {
		return nil, err
	}
	if p.ExtQoSSubscribed != nil && (len(p.ExtQoSSubscribed) < 1 || len(p.ExtQoSSubscribed) > 9) {
		return nil, fmt.Errorf("%w (got %d)", ErrExtQoSSubscribedInvalidSize, len(p.ExtQoSSubscribed))
	}
	if p.Ext2QoSSubscribed != nil && (len(p.Ext2QoSSubscribed) < 1 || len(p.Ext2QoSSubscribed) > 3) {
		return nil, fmt.Errorf("%w (got %d)", ErrExt2QoSSubscribedInvalidSize, len(p.Ext2QoSSubscribed))
	}
	if p.Ext3QoSSubscribed != nil && (len(p.Ext3QoSSubscribed) < 1 || len(p.Ext3QoSSubscribed) > 2) {
		return nil, fmt.Errorf("%w (got %d)", ErrExt3QoSSubscribedInvalidSize, len(p.Ext3QoSSubscribed))
	}
	if p.Ext4QoSSubscribed != nil && len(p.Ext4QoSSubscribed) != 1 {
		return nil, fmt.Errorf("%w (got %d)", ErrExt4QoSSubscribedInvalidSize, len(p.Ext4QoSSubscribed))
	}
	if err := validateExtQoSHierarchy(
		p.ExtQoSSubscribed != nil,
		p.Ext2QoSSubscribed != nil,
		p.Ext3QoSSubscribed != nil,
		p.Ext4QoSSubscribed != nil,
	); err != nil {
		return nil, err
	}
	if p.PdpAddress != nil && (len(p.PdpAddress) < 1 || len(p.PdpAddress) > 16) {
		return nil, fmt.Errorf("%w (got %d)", ErrPDPAddressInvalidSize, len(p.PdpAddress))
	}
	if p.ExtPdpType != nil && len(p.ExtPdpType) != 2 {
		return nil, fmt.Errorf("%w (got %d)", ErrExtPDPTypeInvalidSize, len(p.ExtPdpType))
	}
	if p.ExtPdpAddress != nil {
		if p.PdpAddress == nil {
			return nil, ErrExtPDPAddressWithoutPDPAddress
		}
		if len(p.ExtPdpAddress) < 1 || len(p.ExtPdpAddress) > 16 {
			return nil, fmt.Errorf("%w (got %d)", ErrExtPDPAddressInvalidSize, len(p.ExtPdpAddress))
		}
	}
	if p.PdpChargingCharacteristics != nil && len(p.PdpChargingCharacteristics) != 2 {
		return nil, fmt.Errorf("%w (got %d)", ErrPDPChargingCharsInvalidSize, len(p.PdpChargingCharacteristics))
	}
	if p.ApnOiReplacement != nil {
		if err := validateAPNOIReplacement(p.ApnOiReplacement, "PDPContext.ApnOiReplacement"); err != nil {
			return nil, err
		}
	}
	if p.RestorationPriority != nil && len(p.RestorationPriority) != 1 {
		return nil, fmt.Errorf("%w (got %d)", ErrRestorationPriorityInvalidSize, len(p.RestorationPriority))
	}
	if p.SCEFID != nil {
		if err := validateFQDN(p.SCEFID, "PDPContext.SCEFID"); err != nil {
			return nil, err
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
	if p.VplmnAddressAllowed {
		out.VplmnAddressAllowed = &struct{}{}
	}
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
	id, err := narrowInt64Range(int64(w.PdpContextId), 1, gsm_map.MaxNumOfPDPContexts, "PDPContext.PdpContextId")
	if err != nil {
		// Wrap to preserve encode/decode symmetry on errors.Is(...).
		return nil, fmt.Errorf("%w: %v", ErrPDPContextIdOutOfRange, err)
	}
	if len(w.PdpType) != 2 {
		return nil, fmt.Errorf("%w (got %d)", ErrPDPTypeInvalidSize, len(w.PdpType))
	}
	if len(w.QosSubscribed) != 3 {
		return nil, fmt.Errorf("%w (got %d)", ErrQoSSubscribedInvalidSize, len(w.QosSubscribed))
	}
	if err := validateAPN(HexBytes(w.Apn), "PDPContext.Apn"); err != nil {
		return nil, err
	}
	if w.ExtQoSSubscribed != nil && (len(*w.ExtQoSSubscribed) < 1 || len(*w.ExtQoSSubscribed) > 9) {
		return nil, fmt.Errorf("%w (got %d)", ErrExtQoSSubscribedInvalidSize, len(*w.ExtQoSSubscribed))
	}
	if w.Ext2QoSSubscribed != nil && (len(*w.Ext2QoSSubscribed) < 1 || len(*w.Ext2QoSSubscribed) > 3) {
		return nil, fmt.Errorf("%w (got %d)", ErrExt2QoSSubscribedInvalidSize, len(*w.Ext2QoSSubscribed))
	}
	if w.Ext3QoSSubscribed != nil && (len(*w.Ext3QoSSubscribed) < 1 || len(*w.Ext3QoSSubscribed) > 2) {
		return nil, fmt.Errorf("%w (got %d)", ErrExt3QoSSubscribedInvalidSize, len(*w.Ext3QoSSubscribed))
	}
	if w.Ext4QoSSubscribed != nil && len(*w.Ext4QoSSubscribed) != 1 {
		return nil, fmt.Errorf("%w (got %d)", ErrExt4QoSSubscribedInvalidSize, len(*w.Ext4QoSSubscribed))
	}
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
		VplmnAddressAllowed: w.VplmnAddressAllowed != nil,
	}
	if w.PdpAddress != nil {
		if len(*w.PdpAddress) < 1 || len(*w.PdpAddress) > 16 {
			return nil, fmt.Errorf("%w (got %d)", ErrPDPAddressInvalidSize, len(*w.PdpAddress))
		}
		out.PdpAddress = HexBytes(*w.PdpAddress)
	}
	if w.ExtQoSSubscribed != nil {
		out.ExtQoSSubscribed = HexBytes(*w.ExtQoSSubscribed)
	}
	if w.PdpChargingCharacteristics != nil {
		if len(*w.PdpChargingCharacteristics) != 2 {
			return nil, fmt.Errorf("%w (got %d)", ErrPDPChargingCharsInvalidSize, len(*w.PdpChargingCharacteristics))
		}
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
		if err := validateAPNOIReplacement(HexBytes(*w.ApnOiReplacement), "PDPContext.ApnOiReplacement"); err != nil {
			return nil, err
		}
		out.ApnOiReplacement = HexBytes(*w.ApnOiReplacement)
	}
	if w.ExtPdpType != nil {
		if len(*w.ExtPdpType) != 2 {
			return nil, fmt.Errorf("%w (got %d)", ErrExtPDPTypeInvalidSize, len(*w.ExtPdpType))
		}
		out.ExtPdpType = HexBytes(*w.ExtPdpType)
	}
	if w.ExtPdpAddress != nil {
		if w.PdpAddress == nil {
			return nil, ErrExtPDPAddressWithoutPDPAddress
		}
		if len(*w.ExtPdpAddress) < 1 || len(*w.ExtPdpAddress) > 16 {
			return nil, fmt.Errorf("%w (got %d)", ErrExtPDPAddressInvalidSize, len(*w.ExtPdpAddress))
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
		if len(*w.RestorationPriority) != 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrRestorationPriorityInvalidSize, len(*w.RestorationPriority))
		}
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
		if err := validateFQDN(HexBytes(*w.SCEFID), "PDPContext.SCEFID"); err != nil {
			return nil, err
		}
		out.SCEFID = HexBytes(*w.SCEFID)
	}
	return out, nil
}

// ============================================================================
// GPRSDataList / GPRSSubscriptionData — TS 29.002 MAP-MS-DataTypes.asn:1517-1595
// ============================================================================

func convertGPRSDataListToWire(list GPRSDataList) (gsm_map.GPRSDataList, error) {
	if list == nil {
		return nil, nil
	}
	if int64(len(list)) < 1 || int64(len(list)) > gsm_map.MaxNumOfPDPContexts {
		return nil, fmt.Errorf("%w (got %d)", ErrGPRSDataListSize, len(list))
	}
	out := make(gsm_map.GPRSDataList, len(list))
	for i, p := range list {
		w, err := convertPDPContextToWire(&p)
		if err != nil {
			return nil, fmt.Errorf("GPRSDataList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToGPRSDataList(w gsm_map.GPRSDataList) (GPRSDataList, error) {
	if w == nil {
		return nil, nil
	}
	if int64(len(w)) < 1 || int64(len(w)) > gsm_map.MaxNumOfPDPContexts {
		return nil, fmt.Errorf("%w (got %d)", ErrGPRSDataListSize, len(w))
	}
	out := make(GPRSDataList, len(w))
	for i, p := range w {
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
	out := &gsm_map.GPRSSubscriptionData{GprsDataList: gdl}
	if g.CompleteDataListIncluded {
		out.CompleteDataListIncluded = &struct{}{}
	}
	if g.ApnOiReplacement != nil {
		if err := validateAPNOIReplacement(g.ApnOiReplacement, "GPRSSubscriptionData.ApnOiReplacement"); err != nil {
			return nil, err
		}
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
		CompleteDataListIncluded: w.CompleteDataListIncluded != nil,
	}
	if w.ApnOiReplacement != nil {
		if err := validateAPNOIReplacement(HexBytes(*w.ApnOiReplacement), "GPRSSubscriptionData.ApnOiReplacement"); err != nil {
			return nil, err
		}
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
	if len(l.LsaIdentity) != 3 {
		return nil, fmt.Errorf("%w (got %d)", ErrLSAIdentityInvalidSize, len(l.LsaIdentity))
	}
	if len(l.LsaAttributes) != 1 {
		return nil, fmt.Errorf("%w (got %d)", ErrLSAAttributesInvalidSize, len(l.LsaAttributes))
	}
	out := &gsm_map.LSAData{
		LsaIdentity:   gsm_map.LSAIdentity(l.LsaIdentity),
		LsaAttributes: gsm_map.LSAAttributes(l.LsaAttributes),
	}
	if l.LsaActiveModeIndicator {
		out.LsaActiveModeIndicator = &struct{}{}
	}
	return out, nil
}

func convertWireToLSAData(w *gsm_map.LSAData) (*LSAData, error) {
	if w == nil {
		return nil, nil
	}
	if len(w.LsaIdentity) != 3 {
		return nil, fmt.Errorf("%w (got %d)", ErrLSAIdentityInvalidSize, len(w.LsaIdentity))
	}
	if len(w.LsaAttributes) != 1 {
		return nil, fmt.Errorf("%w (got %d)", ErrLSAAttributesInvalidSize, len(w.LsaAttributes))
	}
	return &LSAData{
		LsaIdentity:            HexBytes(w.LsaIdentity),
		LsaAttributes:          HexBytes(w.LsaAttributes),
		LsaActiveModeIndicator: w.LsaActiveModeIndicator != nil,
	}, nil
}

func convertLSADataListToWire(list LSADataList) (gsm_map.LSADataList, error) {
	if list == nil {
		return nil, nil
	}
	if int64(len(list)) < 1 || int64(len(list)) > gsm_map.MaxNumOfLSAs {
		return nil, fmt.Errorf("%w (got %d)", ErrLSADataListSize, len(list))
	}
	out := make(gsm_map.LSADataList, len(list))
	for i, l := range list {
		w, err := convertLSADataToWire(&l)
		if err != nil {
			return nil, fmt.Errorf("LSADataList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToLSADataList(w gsm_map.LSADataList) (LSADataList, error) {
	if w == nil {
		return nil, nil
	}
	if int64(len(w)) < 1 || int64(len(w)) > gsm_map.MaxNumOfLSAs {
		return nil, fmt.Errorf("%w (got %d)", ErrLSADataListSize, len(w))
	}
	out := make(LSADataList, len(w))
	for i, l := range w {
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
	out := &gsm_map.LSAInformation{}
	if l.CompleteDataListIncluded {
		out.CompleteDataListIncluded = &struct{}{}
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
		CompleteDataListIncluded: w.CompleteDataListIncluded != nil,
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
