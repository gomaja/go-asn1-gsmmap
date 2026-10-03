package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1-gsmmap/gsn"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// --- UpdateGprsLocation ---

func convertUpdateGprsLocationToArg(u *UpdateGprsLocation) (*gsm_map.UpdateGprsLocationArg, error) {
	if u.SgsnNumber == "" {
		return nil, ErrUpdateGprsLocationMissingSgsnNumber
	}
	imsiBytes, err := encodeIdentityDigits(identityIMSI, u.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	sgsnNumber, err := encodeAddressField(u.SgsnNumber, u.SgsnNumberNature, u.SgsnNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding SgsnNumber: %w", err)
	}

	sgsnAddr, err := gsn.Build(u.SGSNAddress)
	if err != nil {
		return nil, fmt.Errorf("encoding SGSNAddress: %w", err)
	}

	arg := &gsm_map.UpdateGprsLocationArg{
		Imsi:        imsiBytes,
		SgsnNumber:  sgsnNumber,
		SgsnAddress: sgsnAddr,
	}

	if u.SGSNCapability != nil {
		sgsnCap, err := convertSGSNCapabilityToWire(u.SGSNCapability)
		if err != nil {
			return nil, fmt.Errorf("SGSNCapability: %w", err)
		}
		arg.SgsnCapability = sgsnCap
	}

	// [1] informPreviousNetworkEntity / [2] psLCSNotSupportedByUE
	arg.InformPreviousNetworkEntity = boolToNullPtr(u.InformPreviousNetworkEntity)
	arg.PsLCSNotSupportedByUE = boolToNullPtr(u.PsLCSNotSupportedByUE)

	// [3] v-gmlc-Address
	if u.VGmlcAddress != "" {
		gsnAddr, err := gsn.Build(u.VGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("encoding VGmlcAddress: %w", err)
		}
		v := gsnAddr
		arg.VGmlcAddress = &v
	}

	// [4] add-info
	if u.AddInfo != nil {
		ai, err := convertAddInfoToWire(u.AddInfo)
		if err != nil {
			return nil, fmt.Errorf("AddInfo: %w", err)
		}
		arg.AddInfo = ai
	}

	// [5] eps-info
	if u.EpsInfo != nil {
		ei, err := convertEpsInfoToWire(u.EpsInfo)
		if err != nil {
			return nil, fmt.Errorf("EpsInfo: %w", err)
		}
		arg.EpsInfo = ei
	}

	// [6]..[13] simple NULL flags
	arg.ServingNodeTypeIndicator = boolToNullPtr(u.ServingNodeTypeIndicator)
	arg.SkipSubscriberDataUpdate = boolToNullPtr(u.SkipSubscriberDataUpdate)

	// [8] usedRatType — Used-RAT-Type per TS 29.002 (extensible enum):
	// only a listed value is sent.
	if u.UsedRatType != nil {
		v := *u.UsedRatType
		if !isListedUsedRATType(v) {
			return nil, fmt.Errorf("UsedRatType=%d: %w", v, ErrUsedRATTypeInvalid)
		}
		arg.UsedRATType = &v
	}

	arg.GprsSubscriptionDataNotNeeded = boolToNullPtr(u.GprsSubscriptionDataNotNeeded)
	arg.NodeTypeIndicator = boolToNullPtr(u.NodeTypeIndicator)
	arg.AreaRestricted = boolToNullPtr(u.AreaRestricted)
	arg.UeReachableIndicator = boolToNullPtr(u.UeReachableIndicator)
	arg.EpsSubscriptionDataNotNeeded = boolToNullPtr(u.EpsSubscriptionDataNotNeeded)

	// [14] ue-SRVCC-Capability — extensible enum per TS 29.002
	// 3GPP TS 29.002 V19.1.0 §17.7.1: only a listed value is sent.
	if u.UeSrvccCapability != nil {
		v := *u.UeSrvccCapability
		if v != UeSrvccNotSupported && v != UeSrvccSupported {
			return nil, fmt.Errorf("UeSrvccCapability=%d: %w", v, ErrUESRVCCCapabilityInvalid)
		}
		arg.UeSrvccCapability = &v
	}

	// [15] eplmn-List
	if len(u.EplmnList) > 0 {
		list := gsm_map.EPLMNList{Values: make([]gsm_map.PLMNId, len(u.EplmnList))}
		for i, raw := range u.EplmnList {
			// go-asn1 does not enforce SEQUENCE OF element SIZE: https://github.com/gomaja/go-asn1/issues/79.
			if len(raw) != 3 {
				return nil, fmt.Errorf("UpdateGprsLocation: EplmnList[%d] length %d: %w", i, len(raw), ErrPLMNIdInvalidLength)
			}
			list.Values[i] = gsm_map.PLMNId(raw)
		}
		arg.EplmnList = &list
	}

	// [16] mme-Number-for-MT-SMS
	if u.MmeNumberForMTSMS != "" {
		mme, err := encodeAddressField(u.MmeNumberForMTSMS, u.MmeNumberForMTSMSNature, u.MmeNumberForMTSMSPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding MmeNumberForMTSMS: %w", err)
		}
		v := mme
		arg.MmeNumberforMTSMS = &v
	}

	// [17] smsRegisterRequest — extensible enum per TS 29.002
	// 3GPP TS 29.002 V19.1.0 §17.7.1; preserve unknown values per Postel's law.
	if u.SmsRegisterRequest != nil {
		v := *u.SmsRegisterRequest
		// SMSRegisterRequest is extensible: only a listed value is sent.
		if v < SmsRegistrationRequired || v > SmsRegistrationNoPreference {
			return nil, fmt.Errorf("SmsRegisterRequest=%d: %w", v, ErrSMSRegisterRequestInvalid)
		}
		arg.SmsRegisterRequest = &v
	}

	arg.SmsOnly = boolToNullPtr(u.SmsOnly)

	// [19] names the SGSN by FQDN; [20] carries an FQDN or realm
	// (RFC 6733 §4.3.1).
	if len(u.SgsnName) > 0 {
		v := gsm_map.DiameterIdentity(u.SgsnName)
		arg.SgsnName = &v
	}
	if len(u.SgsnRealm) > 0 {
		v := gsm_map.DiameterIdentity(u.SgsnRealm)
		arg.SgsnRealm = &v
	}

	arg.LgdSupportIndicator = boolToNullPtr(u.LgdSupportIndicator)
	arg.RemovalofMMERegistrationforSMS = boolToNullPtr(u.RemovalofMMERegistrationforSMS)

	// [23] adjacentPLMNList
	if len(u.AdjacentPLMNList) > 0 {
		list := gsm_map.AdjacentPLMNList{Values: make([]gsm_map.PLMNId, len(u.AdjacentPLMNList))}
		for i, raw := range u.AdjacentPLMNList {
			// go-asn1 does not enforce SEQUENCE OF element SIZE: https://github.com/gomaja/go-asn1/issues/79.
			if len(raw) != 3 {
				return nil, fmt.Errorf("UpdateGprsLocation: AdjacentPLMNList[%d] length %d: %w", i, len(raw), ErrPLMNIdInvalidLength)
			}
			list.Values[i] = gsm_map.PLMNId(raw)
		}
		arg.AdjacentPLMNList = &list
	}

	return arg, nil
}

func convertArgToUpdateGprsLocation(arg *gsm_map.UpdateGprsLocationArg) (*UpdateGprsLocation, error) {
	imsi, err := decodeIdentityDigits(identityIMSI, arg.Imsi)
	if err != nil {
		return nil, fmt.Errorf("decoding IMSI: %w", err)
	}

	sgsnNum, sgsnNature, sgsnPlan, err := decodeAddressField(arg.SgsnNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding SgsnNumber: %w", err)
	}
	if sgsnNum == "" {
		return nil, ErrUpdateGprsLocationMissingSgsnNumber
	}

	sgsnAddr, err := gsn.Parse(arg.SgsnAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding SGSNAddress: %w", err)
	}

	u := &UpdateGprsLocation{
		IMSI:             imsi,
		SgsnNumber:       sgsnNum,
		SgsnNumberNature: sgsnNature,
		SgsnNumberPlan:   sgsnPlan,
		SGSNAddress:      sgsnAddr,
	}

	if arg.SgsnCapability != nil {
		sc, err := convertWireToSGSNCapability(arg.SgsnCapability)
		if err != nil {
			return nil, fmt.Errorf("SGSNCapability: %w", err)
		}
		u.SGSNCapability = sc
	}

	u.InformPreviousNetworkEntity = nullPtrToBool(arg.InformPreviousNetworkEntity)
	u.PsLCSNotSupportedByUE = nullPtrToBool(arg.PsLCSNotSupportedByUE)

	if arg.VGmlcAddress != nil {
		addr, err := gsn.Parse(*arg.VGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("decoding VGmlcAddress: %w", err)
		}
		u.VGmlcAddress = addr
	}

	if arg.AddInfo != nil {
		ai, err := convertWireToAddInfo(arg.AddInfo)
		if err != nil {
			return nil, fmt.Errorf("AddInfo: %w", err)
		}
		u.AddInfo = ai
	}

	if arg.EpsInfo != nil {
		ei, err := convertWireToEpsInfo(arg.EpsInfo)
		if err != nil {
			return nil, fmt.Errorf("EpsInfo: %w", err)
		}
		u.EpsInfo = ei
	}

	u.ServingNodeTypeIndicator = nullPtrToBool(arg.ServingNodeTypeIndicator)
	u.SkipSubscriberDataUpdate = nullPtrToBool(arg.SkipSubscriberDataUpdate)

	// UsedRATType — extensible enum per TS 29.002; an unknown value is kept
	// (3GPP TS 29.002 V19.1.0 §17.1.4) and Marshal refuses it.
	if arg.UsedRATType != nil {
		v := *arg.UsedRATType
		u.UsedRatType = &v
	}

	u.GprsSubscriptionDataNotNeeded = nullPtrToBool(arg.GprsSubscriptionDataNotNeeded)
	u.NodeTypeIndicator = nullPtrToBool(arg.NodeTypeIndicator)
	u.AreaRestricted = nullPtrToBool(arg.AreaRestricted)
	u.UeReachableIndicator = nullPtrToBool(arg.UeReachableIndicator)
	u.EpsSubscriptionDataNotNeeded = nullPtrToBool(arg.EpsSubscriptionDataNotNeeded)

	// UeSrvccCapability — extensible enum per TS 29.002; an unknown value is
	// kept (3GPP TS 29.002 V19.1.0 §17.1.4) and Marshal refuses it.
	if arg.UeSrvccCapability != nil {
		v := *arg.UeSrvccCapability
		u.UeSrvccCapability = &v
	}

	if arg.EplmnList != nil && len(arg.EplmnList.Values) > 0 {
		list := make([]HexBytes, len(arg.EplmnList.Values))
		for i, plmn := range arg.EplmnList.Values {
			// go-asn1 does not enforce SEQUENCE OF element SIZE: https://github.com/gomaja/go-asn1/issues/79.
			if len(plmn) != 3 {
				return nil, fmt.Errorf("UpdateGprsLocation: EplmnList[%d] length %d: %w", i, len(plmn), ErrPLMNIdInvalidLength)
			}
			list[i] = HexBytes(plmn)
		}
		u.EplmnList = list
	}

	if arg.MmeNumberforMTSMS != nil {
		mme, nature, plan, err := decodeAddressWithDigits(*arg.MmeNumberforMTSMS, ErrUpdateGprsLocationMmeNumberForMTSMSDecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding MmeNumberForMTSMS: %w", err)
		}
		u.MmeNumberForMTSMS = mme
		u.MmeNumberForMTSMSNature = nature
		u.MmeNumberForMTSMSPlan = plan
	}

	// SmsRegisterRequest — extensible enum per TS 29.002; an unknown value
	// is kept (3GPP TS 29.002 V19.1.0 §17.1.4) and Marshal refuses it.
	if arg.SmsRegisterRequest != nil {
		v := *arg.SmsRegisterRequest
		u.SmsRegisterRequest = &v
	}

	u.SmsOnly = nullPtrToBool(arg.SmsOnly)

	if arg.SgsnName != nil {
		u.SgsnName = HexBytes(*arg.SgsnName)
	}
	if arg.SgsnRealm != nil {
		u.SgsnRealm = HexBytes(*arg.SgsnRealm)
	}

	u.LgdSupportIndicator = nullPtrToBool(arg.LgdSupportIndicator)
	u.RemovalofMMERegistrationforSMS = nullPtrToBool(arg.RemovalofMMERegistrationforSMS)

	if arg.AdjacentPLMNList != nil && len(arg.AdjacentPLMNList.Values) > 0 {
		list := make([]HexBytes, len(arg.AdjacentPLMNList.Values))
		for i, plmn := range arg.AdjacentPLMNList.Values {
			// go-asn1 does not enforce SEQUENCE OF element SIZE: https://github.com/gomaja/go-asn1/issues/79.
			if len(plmn) != 3 {
				return nil, fmt.Errorf("UpdateGprsLocation: AdjacentPLMNList[%d] length %d: %w", i, len(plmn), ErrPLMNIdInvalidLength)
			}
			list[i] = HexBytes(plmn)
		}
		u.AdjacentPLMNList = list
	}

	return u, nil
}

// --- SGSNCapability ---

func convertSGSNCapabilityToWire(s *SGSNCapability) (*gsm_map.SGSNCapability, error) {
	out := &gsm_map.SGSNCapability{}

	out.SolsaSupportIndicator = boolToNullPtr(s.SolsaSupportIndicator)

	if s.SuperChargerSupportedInServingNetworkEntity != nil {
		sc, err := convertSuperChargerInfoToWire(s.SuperChargerSupportedInServingNetworkEntity)
		if err != nil {
			return nil, fmt.Errorf("SuperChargerInfo: %w", err)
		}
		out.SuperChargerSupportedInServingNetworkEntity = sc
	}

	out.GprsEnhancementsSupportIndicator = boolToNullPtr(s.GprsEnhancementsSupportIndicator)

	if s.SupportedCamelPhases != nil {
		bs := convertCamelPhasesToBitString(s.SupportedCamelPhases)
		out.SupportedCamelPhases = &bs
	}

	if s.SupportedLCSCapabilitySets != nil {
		bs := convertLCSCapsToBitString(s.SupportedLCSCapabilitySets)
		out.SupportedLCSCapabilitySets = &bs
	}

	if s.OfferedCamel4CSIs != nil {
		bs := convertOfferedCamel4CSIsToBitString(s.OfferedCamel4CSIs)
		out.OfferedCamel4CSIs = &bs
	}

	out.SmsCallBarringSupportIndicator = boolToNullPtr(s.SmsCallBarringSupportIndicator)

	if s.SupportedRATTypesIndicator != nil {
		bs := convertSupportedRATTypesToBitString(s.SupportedRATTypesIndicator)
		out.SupportedRATTypesIndicator = &bs
	}

	if s.SupportedFeatures != nil {
		bs := convertSupportedFeaturesToBitString(s.SupportedFeatures)
		out.SupportedFeatures = &bs
	}

	out.TAdsDataRetrieval = boolToNullPtr(s.TAdsDataRetrieval)

	if s.HomogeneousSupportOfIMSVoiceOverPSSessions != nil {
		v := *s.HomogeneousSupportOfIMSVoiceOverPSSessions
		out.HomogeneousSupportOfIMSVoiceOverPSSessions = &v
	}

	out.CancellationTypeInitialAttach = boolToNullPtr(s.CancellationTypeInitialAttach)
	out.MsisdnLessOperationSupported = boolToNullPtr(s.MsisdnLessOperationSupported)
	out.UpdateofHomogeneousSupportOfIMSVoiceOverPSSessions = boolToNullPtr(s.UpdateofHomogeneousSupportOfIMSVoiceOverPSSessions)
	out.ResetIdsSupported = boolToNullPtr(s.ResetIdsSupported)

	if s.ExtSupportedFeatures != nil {
		bs := convertExtSupportedFeaturesToBitString(s.ExtSupportedFeatures)
		out.ExtSupportedFeatures = &bs
	}

	return out, nil
}

func convertWireToSGSNCapability(w *gsm_map.SGSNCapability) (*SGSNCapability, error) {
	out := &SGSNCapability{}

	out.SolsaSupportIndicator = nullPtrToBool(w.SolsaSupportIndicator)

	if w.SuperChargerSupportedInServingNetworkEntity != nil {
		sc, err := convertWireToSuperChargerInfo(w.SuperChargerSupportedInServingNetworkEntity)
		if err != nil {
			return nil, fmt.Errorf("SuperChargerInfo: %w", err)
		}
		out.SuperChargerSupportedInServingNetworkEntity = sc
	}

	out.GprsEnhancementsSupportIndicator = nullPtrToBool(w.GprsEnhancementsSupportIndicator)

	if w.SupportedCamelPhases != nil {
		out.SupportedCamelPhases = convertBitStringToCamelPhases(*w.SupportedCamelPhases)
	}

	if w.SupportedLCSCapabilitySets != nil {
		out.SupportedLCSCapabilitySets = convertBitStringToLCSCaps(*w.SupportedLCSCapabilitySets)
	}

	if w.OfferedCamel4CSIs != nil {
		out.OfferedCamel4CSIs = convertBitStringToOfferedCamel4CSIs(*w.OfferedCamel4CSIs)
	}

	out.SmsCallBarringSupportIndicator = nullPtrToBool(w.SmsCallBarringSupportIndicator)

	if w.SupportedRATTypesIndicator != nil {
		out.SupportedRATTypesIndicator = convertBitStringToSupportedRATTypes(*w.SupportedRATTypesIndicator)
	}

	if w.SupportedFeatures != nil {
		out.SupportedFeatures = convertBitStringToSupportedFeatures(*w.SupportedFeatures)
	}

	out.TAdsDataRetrieval = nullPtrToBool(w.TAdsDataRetrieval)

	if w.HomogeneousSupportOfIMSVoiceOverPSSessions != nil {
		v := *w.HomogeneousSupportOfIMSVoiceOverPSSessions
		out.HomogeneousSupportOfIMSVoiceOverPSSessions = &v
	}

	out.CancellationTypeInitialAttach = nullPtrToBool(w.CancellationTypeInitialAttach)
	out.MsisdnLessOperationSupported = nullPtrToBool(w.MsisdnLessOperationSupported)
	out.UpdateofHomogeneousSupportOfIMSVoiceOverPSSessions = nullPtrToBool(w.UpdateofHomogeneousSupportOfIMSVoiceOverPSSessions)
	out.ResetIdsSupported = nullPtrToBool(w.ResetIdsSupported)

	if w.ExtSupportedFeatures != nil {
		out.ExtSupportedFeatures = convertBitStringToExtSupportedFeatures(*w.ExtSupportedFeatures)
	}

	return out, nil
}

// --- EpsInfo CHOICE ---

func convertEpsInfoToWire(e *EpsInfo) (*gsm_map.EPSInfo, error) {
	hasPdn := e.PdnGwUpdate != nil
	hasIsr := e.IsrInformationBits != 0 || len(e.IsrInformation) > 0
	if hasPdn && hasIsr {
		return nil, ErrSriChoiceMultipleAlternatives
	}
	if !hasPdn && !hasIsr {
		return nil, ErrSriChoiceNoAlternative
	}
	if hasPdn {
		pgu, err := convertPdnGwUpdateToWire(e.PdnGwUpdate)
		if err != nil {
			return nil, err
		}
		v := gsm_map.NewEPSInfoPdnGwUpdate(*pgu)
		return &v, nil
	}
	// IsrInformation is BIT STRING (SIZE(3..8)) per TS 29.002 §17.7.1.
	bs, err := bitStringToWire("EpsInfo.IsrInformation", e.IsrInformation, e.IsrInformationBits)
	if err != nil {
		return nil, err
	}
	v := gsm_map.NewEPSInfoIsrInformation(bs)
	return &v, nil
}

func convertWireToEpsInfo(w *gsm_map.EPSInfo) (*EpsInfo, error) {
	switch w.Choice {
	case gsm_map.EPSInfoChoicePdnGwUpdate:
		if w.PdnGwUpdate == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		pgw, err := convertWireToPdnGwUpdate(w.PdnGwUpdate)
		if err != nil {
			return nil, err
		}
		return &EpsInfo{PdnGwUpdate: pgw}, nil
	case gsm_map.EPSInfoChoiceIsrInformation:
		if w.IsrInformation == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		bits := w.IsrInformation.BitLength
		return &EpsInfo{
			IsrInformation:     HexBytes(append([]byte(nil), w.IsrInformation.Bytes...)),
			IsrInformationBits: bits,
		}, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertPdnGwUpdateToWire(p *PdnGwUpdate) (*gsm_map.PDNGWUpdate, error) {
	out := &gsm_map.PDNGWUpdate{}
	if len(p.APN) > 0 {
		apn := append([]byte(nil), p.APN...)
		out.Apn = &apn
	}
	if p.PdnGwIdentity != nil {
		id, err := convertPdnGwIdentityToWire(p.PdnGwIdentity)
		if err != nil {
			return nil, err
		}
		out.PdnGwIdentity = id
	}
	if p.ContextID != nil {
		v := int64(*p.ContextID)
		out.ContextId = &v
	}
	return out, nil
}

// convertWireToPdnGwUpdate copies a wire PDNGWUpdate. ContextId has
// range 1..50 in 3GPP TS 29.002 V19.1.0 §17.7.1.
func convertWireToPdnGwUpdate(w *gsm_map.PDNGWUpdate) (*PdnGwUpdate, error) {
	out := &PdnGwUpdate{}
	if w.Apn != nil {
		out.APN = HexBytes(append([]byte(nil), (*w.Apn)...))
	}
	if w.PdnGwIdentity != nil {
		pid, err := convertWireToPdnGwIdentity(w.PdnGwIdentity)
		if err != nil {
			return nil, err
		}
		out.PdnGwIdentity = pid
	}
	if w.ContextId != nil {
		v := int(*w.ContextId)
		out.ContextID = &v
	}
	return out, nil
}

func convertPdnGwIdentityToWire(p *PdnGwIdentity) (*gsm_map.PDNGWIdentity, error) {
	if len(p.IPv4Address) > 0 && len(p.IPv4Address) != 4 {
		return nil, fmt.Errorf("%w: PdnGwIdentity: IPv4Address must be exactly 4 octets, got %d", ErrPdnGwIdentityIPv4AddressInvalidLength, len(p.IPv4Address))
	}
	if len(p.IPv6Address) > 0 && len(p.IPv6Address) != 16 {
		return nil, fmt.Errorf("%w: PdnGwIdentity: IPv6Address must be exactly 16 octets, got %d", ErrPdnGwIdentityIPv6AddressInvalidLength, len(p.IPv6Address))
	}
	if len(p.IPv4Address) == 0 && len(p.IPv6Address) == 0 && len(p.Name) == 0 {
		return nil, fmt.Errorf("%w: PdnGwIdentity: at least one of IPv4Address, IPv6Address, or Name must be set", ErrPdnGwIdentityAddressMissing)
	}
	out := &gsm_map.PDNGWIdentity{}
	if len(p.IPv4Address) > 0 {
		v := append([]byte(nil), p.IPv4Address...)
		out.PdnGwIpv4Address = &v
	}
	if len(p.IPv6Address) > 0 {
		v := append([]byte(nil), p.IPv6Address...)
		out.PdnGwIpv6Address = &v
	}
	if len(p.Name) > 0 {
		v := append([]byte(nil), p.Name...)
		out.PdnGwName = &v
	}
	return out, nil
}

// convertWireToPdnGwIdentity decodes a wire PDNGWIdentity, mirroring the
// encoder's validation: IPv4=4 octets, IPv6=16 octets, and at least one
// of the three identity fields must be present.
func convertWireToPdnGwIdentity(w *gsm_map.PDNGWIdentity) (*PdnGwIdentity, error) {
	out := &PdnGwIdentity{}
	if w.PdnGwIpv4Address != nil {
		ip4 := append([]byte(nil), (*w.PdnGwIpv4Address)...)
		if len(ip4) != 4 {
			return nil, fmt.Errorf("%w: PdnGwIdentity: IPv4Address must be exactly 4 octets, got %d", ErrPdnGwIdentityIPv4AddressInvalidLength, len(ip4))
		}
		out.IPv4Address = HexBytes(ip4)
	}
	if w.PdnGwIpv6Address != nil {
		ip6 := append([]byte(nil), (*w.PdnGwIpv6Address)...)
		if len(ip6) != 16 {
			return nil, fmt.Errorf("%w: PdnGwIdentity: IPv6Address must be exactly 16 octets, got %d", ErrPdnGwIdentityIPv6AddressInvalidLength, len(ip6))
		}
		out.IPv6Address = HexBytes(ip6)
	}
	if w.PdnGwName != nil {
		out.Name = HexBytes(append([]byte(nil), (*w.PdnGwName)...))
	}
	if len(out.IPv4Address) == 0 && len(out.IPv6Address) == 0 && len(out.Name) == 0 {
		return nil, fmt.Errorf("%w: PdnGwIdentity: at least one of IPv4Address, IPv6Address, or Name must be present", ErrPdnGwIdentityAddressMissing)
	}
	return out, nil
}

// --- UpdateGprsLocationRes ---

func convertUpdateGprsLocationResToRes(u *UpdateGprsLocationRes) (*gsm_map.UpdateGprsLocationRes, error) {
	if u.HlrNumber == "" {
		return nil, ErrUpdateGprsLocationResMissingHlrNumber
	}
	hlr, err := encodeAddressField(u.HlrNumber, u.HlrNumberNature, u.HlrNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding HlrNumber: %w", err)
	}

	return &gsm_map.UpdateGprsLocationRes{
		HlrNumber:                  hlr,
		AddCapability:              boolToNullPtr(u.AddCapability),
		SgsnMmeSeparationSupported: boolToNullPtr(u.SgsnMmeSeparationSupported),
		MmeRegisteredforSMS:        boolToNullPtr(u.MmeRegisteredforSMS),
	}, nil
}

func convertResToUpdateGprsLocationRes(res *gsm_map.UpdateGprsLocationRes) (*UpdateGprsLocationRes, error) {
	hlr, nature, plan, err := decodeAddressField(res.HlrNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding HlrNumber: %w", err)
	}
	if hlr == "" {
		return nil, ErrUpdateGprsLocationResMissingHlrNumber
	}

	return &UpdateGprsLocationRes{
		HlrNumber:                  hlr,
		HlrNumberNature:            nature,
		HlrNumberPlan:              plan,
		AddCapability:              nullPtrToBool(res.AddCapability),
		SgsnMmeSeparationSupported: nullPtrToBool(res.SgsnMmeSeparationSupported),
		MmeRegisteredforSMS:        nullPtrToBool(res.MmeRegisteredforSMS),
	}, nil
}
