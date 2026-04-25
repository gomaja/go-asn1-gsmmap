package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// AllocationRetentionPriority — TS 29.002 MAP-MS-DataTypes.asn:1420
// ============================================================================

func convertAllocationRetentionPriorityToWire(a *AllocationRetentionPriority) *gsm_map.AllocationRetentionPriority {
	if a == nil {
		return nil
	}
	out := &gsm_map.AllocationRetentionPriority{PriorityLevel: a.PriorityLevel}
	if a.PreEmptionCapability != nil {
		v := *a.PreEmptionCapability
		out.PreEmptionCapability = &v
	}
	if a.PreEmptionVulnerability != nil {
		v := *a.PreEmptionVulnerability
		out.PreEmptionVulnerability = &v
	}
	return out
}

func convertWireToAllocationRetentionPriority(w *gsm_map.AllocationRetentionPriority) *AllocationRetentionPriority {
	if w == nil {
		return nil
	}
	out := &AllocationRetentionPriority{PriorityLevel: w.PriorityLevel}
	if w.PreEmptionCapability != nil {
		v := *w.PreEmptionCapability
		out.PreEmptionCapability = &v
	}
	if w.PreEmptionVulnerability != nil {
		v := *w.PreEmptionVulnerability
		out.PreEmptionVulnerability = &v
	}
	return out
}

// ============================================================================
// EPSQoSSubscribed — TS 29.002 MAP-MS-DataTypes.asn:1380
// ============================================================================

func convertEPSQoSSubscribedToWire(q *EPSQoSSubscribed) (*gsm_map.EPSQoSSubscribed, error) {
	if q == nil {
		return nil, nil
	}
	if q.QosClassIdentifier < 1 || q.QosClassIdentifier > 9 {
		return nil, fmt.Errorf("%w (got %d)", ErrQoSClassIdentifierOutOfRange, q.QosClassIdentifier)
	}
	return &gsm_map.EPSQoSSubscribed{
		QosClassIdentifier:          gsm_map.QoSClassIdentifier(q.QosClassIdentifier),
		AllocationRetentionPriority: *convertAllocationRetentionPriorityToWire(&q.AllocationRetentionPriority),
	}, nil
}

func convertWireToEPSQoSSubscribed(w *gsm_map.EPSQoSSubscribed) (*EPSQoSSubscribed, error) {
	if w == nil {
		return nil, nil
	}
	qci, err := narrowInt64Range(int64(w.QosClassIdentifier), 1, 9, "EPSQoSSubscribed.QosClassIdentifier")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQoSClassIdentifierOutOfRange, err)
	}
	return &EPSQoSSubscribed{
		QosClassIdentifier:          qci,
		AllocationRetentionPriority: *convertWireToAllocationRetentionPriority(&w.AllocationRetentionPriority),
	}, nil
}

// ============================================================================
// SpecificAPNInfo / SpecificAPNInfoList — TS 29.002 MAP-MS-DataTypes.asn:1398-1408
// PdnGwIdentity is the pre-existing public type (gsmmap.go:366) shared with
// UpdateGprsLocation; convertPdnGwIdentityToWire / convertWireToPdnGwIdentity
// in convert_updategprsloc.go enforce the strict spec sizes (IPv4=4, IPv6=16)
// and the "at least one of IPv4Address, IPv6Address, or Name" rule.
// ============================================================================

func convertSpecificAPNInfoToWire(s *SpecificAPNInfo) (*gsm_map.SpecificAPNInfo, error) {
	if s == nil {
		return nil, nil
	}
	if err := validateAPN(s.Apn, "SpecificAPNInfo.Apn"); err != nil {
		return nil, err
	}
	gw, err := convertPdnGwIdentityToWire(&s.PdnGwIdentity)
	if err != nil {
		return nil, fmt.Errorf("SpecificAPNInfo.PdnGwIdentity: %w", err)
	}
	return &gsm_map.SpecificAPNInfo{
		Apn:           gsm_map.APN(s.Apn),
		PdnGwIdentity: *gw,
	}, nil
}

func convertWireToSpecificAPNInfo(w *gsm_map.SpecificAPNInfo) (*SpecificAPNInfo, error) {
	if w == nil {
		return nil, nil
	}
	if err := validateAPN(HexBytes(w.Apn), "SpecificAPNInfo.Apn"); err != nil {
		return nil, err
	}
	gw, err := convertWireToPdnGwIdentity(&w.PdnGwIdentity)
	if err != nil {
		return nil, fmt.Errorf("SpecificAPNInfo.PdnGwIdentity: %w", err)
	}
	return &SpecificAPNInfo{
		Apn:           HexBytes(w.Apn),
		PdnGwIdentity: *gw,
	}, nil
}

func convertSpecificAPNInfoListToWire(list SpecificAPNInfoList) (gsm_map.SpecificAPNInfoList, error) {
	if list == nil {
		return nil, nil
	}
	if int64(len(list)) < 1 || int64(len(list)) > gsm_map.MaxNumOfSpecificAPNInfos {
		return nil, fmt.Errorf("%w (got %d)", ErrSpecificAPNInfoListSize, len(list))
	}
	out := make(gsm_map.SpecificAPNInfoList, len(list))
	for i, s := range list {
		w, err := convertSpecificAPNInfoToWire(&s)
		if err != nil {
			return nil, fmt.Errorf("SpecificAPNInfoList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToSpecificAPNInfoList(w gsm_map.SpecificAPNInfoList) (SpecificAPNInfoList, error) {
	if w == nil {
		return nil, nil
	}
	if int64(len(w)) < 1 || int64(len(w)) > gsm_map.MaxNumOfSpecificAPNInfos {
		return nil, fmt.Errorf("%w (got %d)", ErrSpecificAPNInfoListSize, len(w))
	}
	out := make(SpecificAPNInfoList, len(w))
	for i, s := range w {
		v, err := convertWireToSpecificAPNInfo(&s)
		if err != nil {
			return nil, fmt.Errorf("SpecificAPNInfoList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

// ============================================================================
// WLANOffloadability — TS 29.002 MAP-MS-DataTypes
// ============================================================================

func convertWLANOffloadabilityToWire(o *WLANOffloadability) (*gsm_map.WLANOffloadability, error) {
	if o == nil {
		return nil, nil
	}
	if o.WlanOffloadabilityEUTRAN != nil {
		if v := *o.WlanOffloadabilityEUTRAN; v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (EUTRAN got %d)", ErrWLANOffloadabilityIndicationInvalid, v)
		}
	}
	if o.WlanOffloadabilityUTRAN != nil {
		if v := *o.WlanOffloadabilityUTRAN; v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (UTRAN got %d)", ErrWLANOffloadabilityIndicationInvalid, v)
		}
	}
	out := &gsm_map.WLANOffloadability{}
	if o.WlanOffloadabilityEUTRAN != nil {
		v := gsm_map.WLANOffloadabilityIndication(*o.WlanOffloadabilityEUTRAN)
		out.WlanOffloadabilityEUTRAN = &v
	}
	if o.WlanOffloadabilityUTRAN != nil {
		v := gsm_map.WLANOffloadabilityIndication(*o.WlanOffloadabilityUTRAN)
		out.WlanOffloadabilityUTRAN = &v
	}
	return out, nil
}

func convertWireToWLANOffloadability(w *gsm_map.WLANOffloadability) (*WLANOffloadability, error) {
	if w == nil {
		return nil, nil
	}
	out := &WLANOffloadability{}
	if w.WlanOffloadabilityEUTRAN != nil {
		v := WLANOffloadabilityIndication(*w.WlanOffloadabilityEUTRAN)
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (EUTRAN got %d)", ErrWLANOffloadabilityIndicationInvalid, v)
		}
		out.WlanOffloadabilityEUTRAN = &v
	}
	if w.WlanOffloadabilityUTRAN != nil {
		v := WLANOffloadabilityIndication(*w.WlanOffloadabilityUTRAN)
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (UTRAN got %d)", ErrWLANOffloadabilityIndicationInvalid, v)
		}
		out.WlanOffloadabilityUTRAN = &v
	}
	return out, nil
}

// ============================================================================
// APNConfiguration — TS 29.002 MAP-MS-DataTypes.asn:1327
// ============================================================================

func convertAPNConfigurationToWire(a *APNConfiguration) (*gsm_map.APNConfiguration, error) {
	if a == nil {
		return nil, nil
	}
	if a.ContextId < 1 || int64(a.ContextId) > gsm_map.MaxNumOfPDPContexts {
		return nil, fmt.Errorf("%w (got %d)", ErrPDPContextIdOutOfRange, a.ContextId)
	}
	if len(a.PdnType) != 1 {
		return nil, fmt.Errorf("%w (got %d)", ErrPDNTypeInvalidSize, len(a.PdnType))
	}
	if err := validateAPN(a.Apn, "APNConfiguration.Apn"); err != nil {
		return nil, err
	}
	if a.ServedPartyIPIPv4Address != nil {
		if err := validatePDPAddress(a.ServedPartyIPIPv4Address, "APNConfiguration.ServedPartyIPIPv4Address"); err != nil {
			return nil, err
		}
	}
	if a.ServedPartyIPIPv6Address != nil {
		if err := validatePDPAddress(a.ServedPartyIPIPv6Address, "APNConfiguration.ServedPartyIPIPv6Address"); err != nil {
			return nil, err
		}
	}
	if a.ChargingCharacteristics != nil && len(a.ChargingCharacteristics) != 2 {
		return nil, fmt.Errorf("%w (got %d)", ErrPDPChargingCharsInvalidSize, len(a.ChargingCharacteristics))
	}
	if a.ApnOiReplacement != nil {
		if err := validateAPNOIReplacement(a.ApnOiReplacement, "APNConfiguration.ApnOiReplacement"); err != nil {
			return nil, err
		}
	}
	if a.RestorationPriority != nil && len(a.RestorationPriority) != 1 {
		return nil, fmt.Errorf("%w (got %d)", ErrRestorationPriorityInvalidSize, len(a.RestorationPriority))
	}
	if a.SCEFID != nil {
		if err := validateFQDN(a.SCEFID, "APNConfiguration.SCEFID"); err != nil {
			return nil, err
		}
	}
	if a.PdnGwAllocationType != nil {
		if v := *a.PdnGwAllocationType; v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrPDNGWAllocationTypeInvalid, v)
		}
	}
	if a.SiptoPermission != nil {
		if v := *a.SiptoPermission; v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrSIPTOPermissionInvalid, v)
		}
	}
	if a.LipaPermission != nil {
		if v := *a.LipaPermission; v < 0 || v > 2 {
			return nil, fmt.Errorf("%w (got %d)", ErrLIPAPermissionInvalid, v)
		}
	}
	if a.SiptoLocalNetworkPermission != nil {
		if v := *a.SiptoLocalNetworkPermission; v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrSIPTOLocalNetworkPermissionInvalid, v)
		}
	}
	if a.NIDDMechanism != nil {
		if v := *a.NIDDMechanism; v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrNIDDMechanismInvalid, v)
		}
	}
	if a.PdnConnectionContinuity != nil {
		if v := *a.PdnConnectionContinuity; v < 0 || v > 2 {
			return nil, fmt.Errorf("%w (got %d)", ErrPDNConnectionContinuityInvalid, v)
		}
	}

	qos, err := convertEPSQoSSubscribedToWire(&a.EpsQosSubscribed)
	if err != nil {
		return nil, fmt.Errorf("APNConfiguration.EpsQosSubscribed: %w", err)
	}

	out := &gsm_map.APNConfiguration{
		ContextId:           gsm_map.ContextId(a.ContextId),
		PdnType:             gsm_map.PDNType(a.PdnType),
		Apn:                 gsm_map.APN(a.Apn),
		EpsQosSubscribed:    *qos,
		VplmnAddressAllowed: boolToNullPtr(a.VplmnAddressAllowed),
		NonIPPDNTypeIndicator: boolToNullPtr(a.NonIPPDNTypeIndicator),
	}
	if a.ServedPartyIPIPv4Address != nil {
		v := gsm_map.PDPAddress(a.ServedPartyIPIPv4Address)
		out.ServedPartyIPIPv4Address = &v
	}
	if a.PdnGwIdentity != nil {
		gw, err := convertPdnGwIdentityToWire(a.PdnGwIdentity)
		if err != nil {
			return nil, fmt.Errorf("APNConfiguration.PdnGwIdentity: %w", err)
		}
		out.PdnGwIdentity = gw
	}
	if a.PdnGwAllocationType != nil {
		v := gsm_map.PDNGWAllocationType(*a.PdnGwAllocationType)
		out.PdnGwAllocationType = &v
	}
	if a.ChargingCharacteristics != nil {
		v := gsm_map.ChargingCharacteristics(a.ChargingCharacteristics)
		out.ChargingCharacteristics = &v
	}
	if a.Ambr != nil {
		ambr, err := convertAMBRToWire(a.Ambr)
		if err != nil {
			return nil, fmt.Errorf("APNConfiguration.Ambr: %w", err)
		}
		out.Ambr = ambr
	}
	if a.SpecificAPNInfoList != nil {
		sl, err := convertSpecificAPNInfoListToWire(a.SpecificAPNInfoList)
		if err != nil {
			return nil, fmt.Errorf("APNConfiguration.SpecificAPNInfoList: %w", err)
		}
		out.SpecificAPNInfoList = sl
	}
	if a.ServedPartyIPIPv6Address != nil {
		v := gsm_map.PDPAddress(a.ServedPartyIPIPv6Address)
		out.ServedPartyIPIPv6Address = &v
	}
	if a.ApnOiReplacement != nil {
		v := gsm_map.APNOIReplacement(a.ApnOiReplacement)
		out.ApnOiReplacement = &v
	}
	if a.SiptoPermission != nil {
		v := gsm_map.SIPTOPermission(*a.SiptoPermission)
		out.SiptoPermission = &v
	}
	if a.LipaPermission != nil {
		v := gsm_map.LIPAPermission(*a.LipaPermission)
		out.LipaPermission = &v
	}
	if a.RestorationPriority != nil {
		v := gsm_map.RestorationPriority(a.RestorationPriority)
		out.RestorationPriority = &v
	}
	if a.SiptoLocalNetworkPermission != nil {
		v := gsm_map.SIPTOLocalNetworkPermission(*a.SiptoLocalNetworkPermission)
		out.SiptoLocalNetworkPermission = &v
	}
	if a.WlanOffloadability != nil {
		wo, err := convertWLANOffloadabilityToWire(a.WlanOffloadability)
		if err != nil {
			return nil, fmt.Errorf("APNConfiguration.WlanOffloadability: %w", err)
		}
		out.WlanOffloadability = wo
	}
	if a.NIDDMechanism != nil {
		v := gsm_map.NIDDMechanism(*a.NIDDMechanism)
		out.NIDDMechanism = &v
	}
	if a.SCEFID != nil {
		v := gsm_map.FQDN(a.SCEFID)
		out.SCEFID = &v
	}
	if a.PdnConnectionContinuity != nil {
		v := gsm_map.PDNConnectionContinuity(*a.PdnConnectionContinuity)
		out.PdnConnectionContinuity = &v
	}
	return out, nil
}

func convertWireToAPNConfiguration(w *gsm_map.APNConfiguration) (*APNConfiguration, error) {
	if w == nil {
		return nil, nil
	}
	id, err := narrowInt64Range(int64(w.ContextId), 1, gsm_map.MaxNumOfPDPContexts, "APNConfiguration.ContextId")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPDPContextIdOutOfRange, err)
	}
	if len(w.PdnType) != 1 {
		return nil, fmt.Errorf("%w (got %d)", ErrPDNTypeInvalidSize, len(w.PdnType))
	}
	if err := validateAPN(HexBytes(w.Apn), "APNConfiguration.Apn"); err != nil {
		return nil, err
	}

	qos, err := convertWireToEPSQoSSubscribed(&w.EpsQosSubscribed)
	if err != nil {
		return nil, fmt.Errorf("APNConfiguration.EpsQosSubscribed: %w", err)
	}

	out := &APNConfiguration{
		ContextId:             id,
		PdnType:               HexBytes(w.PdnType),
		Apn:                   HexBytes(w.Apn),
		EpsQosSubscribed:      *qos,
		VplmnAddressAllowed:   nullPtrToBool(w.VplmnAddressAllowed),
		NonIPPDNTypeIndicator: nullPtrToBool(w.NonIPPDNTypeIndicator),
	}
	if w.ServedPartyIPIPv4Address != nil {
		if err := validatePDPAddress(HexBytes(*w.ServedPartyIPIPv4Address), "APNConfiguration.ServedPartyIPIPv4Address"); err != nil {
			return nil, err
		}
		out.ServedPartyIPIPv4Address = HexBytes(*w.ServedPartyIPIPv4Address)
	}
	if w.PdnGwIdentity != nil {
		gw, err := convertWireToPdnGwIdentity(w.PdnGwIdentity)
		if err != nil {
			return nil, fmt.Errorf("APNConfiguration.PdnGwIdentity: %w", err)
		}
		out.PdnGwIdentity = gw
	}
	if w.PdnGwAllocationType != nil {
		v := PDNGWAllocationType(*w.PdnGwAllocationType)
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrPDNGWAllocationTypeInvalid, v)
		}
		out.PdnGwAllocationType = &v
	}
	if w.ChargingCharacteristics != nil {
		if len(*w.ChargingCharacteristics) != 2 {
			return nil, fmt.Errorf("%w (got %d)", ErrPDPChargingCharsInvalidSize, len(*w.ChargingCharacteristics))
		}
		out.ChargingCharacteristics = HexBytes(*w.ChargingCharacteristics)
	}
	if w.Ambr != nil {
		ambr, err := convertWireToAMBR(w.Ambr)
		if err != nil {
			return nil, fmt.Errorf("APNConfiguration.Ambr: %w", err)
		}
		out.Ambr = ambr
	}
	if w.SpecificAPNInfoList != nil {
		sl, err := convertWireToSpecificAPNInfoList(w.SpecificAPNInfoList)
		if err != nil {
			return nil, fmt.Errorf("APNConfiguration.SpecificAPNInfoList: %w", err)
		}
		out.SpecificAPNInfoList = sl
	}
	if w.ServedPartyIPIPv6Address != nil {
		if err := validatePDPAddress(HexBytes(*w.ServedPartyIPIPv6Address), "APNConfiguration.ServedPartyIPIPv6Address"); err != nil {
			return nil, err
		}
		out.ServedPartyIPIPv6Address = HexBytes(*w.ServedPartyIPIPv6Address)
	}
	if w.ApnOiReplacement != nil {
		if err := validateAPNOIReplacement(HexBytes(*w.ApnOiReplacement), "APNConfiguration.ApnOiReplacement"); err != nil {
			return nil, err
		}
		out.ApnOiReplacement = HexBytes(*w.ApnOiReplacement)
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
	if w.WlanOffloadability != nil {
		wo, err := convertWireToWLANOffloadability(w.WlanOffloadability)
		if err != nil {
			return nil, fmt.Errorf("APNConfiguration.WlanOffloadability: %w", err)
		}
		out.WlanOffloadability = wo
	}
	if w.NIDDMechanism != nil {
		v := NIDDMechanism(*w.NIDDMechanism)
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("%w (got %d)", ErrNIDDMechanismInvalid, v)
		}
		out.NIDDMechanism = &v
	}
	if w.SCEFID != nil {
		if err := validateFQDN(HexBytes(*w.SCEFID), "APNConfiguration.SCEFID"); err != nil {
			return nil, err
		}
		out.SCEFID = HexBytes(*w.SCEFID)
	}
	if w.PdnConnectionContinuity != nil {
		v := PDNConnectionContinuity(*w.PdnConnectionContinuity)
		if v < 0 || v > 2 {
			return nil, fmt.Errorf("%w (got %d)", ErrPDNConnectionContinuityInvalid, v)
		}
		out.PdnConnectionContinuity = &v
	}
	return out, nil
}

// ============================================================================
// EPSDataList / APNConfigurationProfile / EPSSubscriptionData
// — TS 29.002 MAP-MS-DataTypes.asn:1283-1325
// ============================================================================

func convertEPSDataListToWire(list EPSDataList) (gsm_map.EPSDataList, error) {
	if list == nil {
		return nil, nil
	}
	if int64(len(list)) < 1 || int64(len(list)) > gsm_map.MaxNumOfAPNConfigurations {
		return nil, fmt.Errorf("%w (got %d)", ErrEPSDataListSize, len(list))
	}
	out := make(gsm_map.EPSDataList, len(list))
	for i, a := range list {
		w, err := convertAPNConfigurationToWire(&a)
		if err != nil {
			return nil, fmt.Errorf("EPSDataList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToEPSDataList(w gsm_map.EPSDataList) (EPSDataList, error) {
	if w == nil {
		return nil, nil
	}
	if int64(len(w)) < 1 || int64(len(w)) > gsm_map.MaxNumOfAPNConfigurations {
		return nil, fmt.Errorf("%w (got %d)", ErrEPSDataListSize, len(w))
	}
	out := make(EPSDataList, len(w))
	for i, a := range w {
		v, err := convertWireToAPNConfiguration(&a)
		if err != nil {
			return nil, fmt.Errorf("EPSDataList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

func convertAPNConfigurationProfileToWire(p *APNConfigurationProfile) (*gsm_map.APNConfigurationProfile, error) {
	if p == nil {
		return nil, nil
	}
	if p.DefaultContext < 1 || int64(p.DefaultContext) > gsm_map.MaxNumOfPDPContexts {
		return nil, fmt.Errorf("%w (got %d)", ErrPDPContextIdOutOfRange, p.DefaultContext)
	}
	if p.AdditionalDefaultContext != nil {
		if v := *p.AdditionalDefaultContext; v < 1 || int64(v) > gsm_map.MaxNumOfPDPContexts {
			return nil, fmt.Errorf("%w (additional, got %d)", ErrPDPContextIdOutOfRange, v)
		}
	}
	if p.EpsDataList == nil {
		return nil, ErrAPNConfigurationProfileMissingList
	}
	dl, err := convertEPSDataListToWire(p.EpsDataList)
	if err != nil {
		return nil, err
	}
	out := &gsm_map.APNConfigurationProfile{
		DefaultContext:           gsm_map.ContextId(p.DefaultContext),
		EpsDataList:              dl,
		CompleteDataListIncluded: boolToNullPtr(p.CompleteDataListIncluded),
	}
	if p.AdditionalDefaultContext != nil {
		v := gsm_map.ContextId(*p.AdditionalDefaultContext)
		out.AdditionalDefaultContext = &v
	}
	return out, nil
}

func convertWireToAPNConfigurationProfile(w *gsm_map.APNConfigurationProfile) (*APNConfigurationProfile, error) {
	if w == nil {
		return nil, nil
	}
	def, err := narrowInt64Range(int64(w.DefaultContext), 1, gsm_map.MaxNumOfPDPContexts, "APNConfigurationProfile.DefaultContext")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPDPContextIdOutOfRange, err)
	}
	if w.EpsDataList == nil {
		return nil, ErrAPNConfigurationProfileMissingList
	}
	dl, err := convertWireToEPSDataList(w.EpsDataList)
	if err != nil {
		return nil, err
	}
	out := &APNConfigurationProfile{
		DefaultContext:           def,
		EpsDataList:              dl,
		CompleteDataListIncluded: nullPtrToBool(w.CompleteDataListIncluded),
	}
	if w.AdditionalDefaultContext != nil {
		add, err := narrowInt64Range(int64(*w.AdditionalDefaultContext), 1, gsm_map.MaxNumOfPDPContexts, "APNConfigurationProfile.AdditionalDefaultContext")
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrPDPContextIdOutOfRange, err)
		}
		out.AdditionalDefaultContext = &add
	}
	return out, nil
}

func convertEPSSubscriptionDataToWire(e *EPSSubscriptionData) (*gsm_map.EPSSubscriptionData, error) {
	if e == nil {
		return nil, nil
	}
	if e.ApnOiReplacement != nil {
		if err := validateAPNOIReplacement(e.ApnOiReplacement, "EPSSubscriptionData.ApnOiReplacement"); err != nil {
			return nil, err
		}
	}
	if e.RfspId != nil {
		if v := *e.RfspId; v < 1 || v > MaxRFSPID {
			return nil, fmt.Errorf("%w (got %d)", ErrRFSPIDOutOfRange, v)
		}
	}
	out := &gsm_map.EPSSubscriptionData{
		MpsCSPriority:    boolToNullPtr(e.MpsCSPriority),
		MpsEPSPriority:   boolToNullPtr(e.MpsEPSPriority),
		SubscribedVsrvcc: boolToNullPtr(e.SubscribedVsrvcc),
	}
	if e.ApnOiReplacement != nil {
		v := gsm_map.APNOIReplacement(e.ApnOiReplacement)
		out.ApnOiReplacement = &v
	}
	if e.RfspId != nil {
		v := gsm_map.RFSPID(*e.RfspId)
		out.RfspId = &v
	}
	if e.Ambr != nil {
		ambr, err := convertAMBRToWire(e.Ambr)
		if err != nil {
			return nil, fmt.Errorf("EPSSubscriptionData.Ambr: %w", err)
		}
		out.Ambr = ambr
	}
	if e.ApnConfigurationProfile != nil {
		acp, err := convertAPNConfigurationProfileToWire(e.ApnConfigurationProfile)
		if err != nil {
			return nil, fmt.Errorf("EPSSubscriptionData.ApnConfigurationProfile: %w", err)
		}
		out.ApnConfigurationProfile = acp
	}
	if e.StnSr != "" {
		isdn, err := encodeAddressField(e.StnSr, e.StnSrNature, e.StnSrPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding EPSSubscriptionData.StnSr: %w", err)
		}
		v := gsm_map.ISDNAddressString(isdn)
		out.StnSr = &v
	}
	return out, nil
}

func convertWireToEPSSubscriptionData(w *gsm_map.EPSSubscriptionData) (*EPSSubscriptionData, error) {
	if w == nil {
		return nil, nil
	}
	out := &EPSSubscriptionData{
		MpsCSPriority:    nullPtrToBool(w.MpsCSPriority),
		MpsEPSPriority:   nullPtrToBool(w.MpsEPSPriority),
		SubscribedVsrvcc: nullPtrToBool(w.SubscribedVsrvcc),
	}
	if w.ApnOiReplacement != nil {
		if err := validateAPNOIReplacement(HexBytes(*w.ApnOiReplacement), "EPSSubscriptionData.ApnOiReplacement"); err != nil {
			return nil, err
		}
		out.ApnOiReplacement = HexBytes(*w.ApnOiReplacement)
	}
	if w.RfspId != nil {
		v, err := narrowInt64Range(int64(*w.RfspId), 1, MaxRFSPID, "EPSSubscriptionData.RfspId")
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrRFSPIDOutOfRange, err)
		}
		out.RfspId = &v
	}
	if w.Ambr != nil {
		ambr, err := convertWireToAMBR(w.Ambr)
		if err != nil {
			return nil, fmt.Errorf("EPSSubscriptionData.Ambr: %w", err)
		}
		out.Ambr = ambr
	}
	if w.ApnConfigurationProfile != nil {
		acp, err := convertWireToAPNConfigurationProfile(w.ApnConfigurationProfile)
		if err != nil {
			return nil, fmt.Errorf("EPSSubscriptionData.ApnConfigurationProfile: %w", err)
		}
		out.ApnConfigurationProfile = acp
	}
	if w.StnSr != nil {
		s, nature, plan, err := decodeAddressField([]byte(*w.StnSr))
		if err != nil {
			return nil, fmt.Errorf("decoding EPSSubscriptionData.StnSr: %w", err)
		}
		out.StnSr = s
		out.StnSrNature = nature
		out.StnSrPlan = plan
	}
	return out, nil
}
