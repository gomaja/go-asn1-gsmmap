package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/runtime"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/gsn"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// --- UpdateGprsLocation ---

func convertUpdateGprsLocationToArg(u *UpdateGprsLocation) (*gsm_map.UpdateGprsLocationArg, error) {
	if u.IMSI == "" {
		return nil, fmt.Errorf("UpdateGprsLocation: IMSI is mandatory and must be non-empty")
	}
	if u.SGSNNumber == "" {
		return nil, fmt.Errorf("UpdateGprsLocation: SGSNNumber is mandatory and must be non-empty")
	}
	imsiBytes, err := tbcd.Encode(u.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	sgsnNumber, err := encodeAddressField(u.SGSNNumber, u.SGSNNature, u.SGSNPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding SGSNNumber: %w", err)
	}

	sgsnAddr, err := gsn.Build(u.SGSNAddress)
	if err != nil {
		return nil, fmt.Errorf("encoding SGSNAddress: %w", err)
	}

	arg := &gsm_map.UpdateGprsLocationArg{
		Imsi:        gsm_map.IMSI(imsiBytes),
		SgsnNumber:  gsm_map.ISDNAddressString(sgsnNumber),
		SgsnAddress: gsm_map.GSNAddress(sgsnAddr),
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
		v := gsm_map.GSNAddress(gsnAddr)
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

	// [8] usedRatType — Used-RAT-Type 0..5 per TS 29.002.
	if u.UsedRatType != nil {
		if *u.UsedRatType < 0 || *u.UsedRatType > 5 {
			return nil, fmt.Errorf("UsedRATType out of range 0..5: %d", *u.UsedRatType)
		}
		v := *u.UsedRatType
		arg.UsedRATType = &v
	}

	arg.GprsSubscriptionDataNotNeeded = boolToNullPtr(u.GprsSubscriptionDataNotNeeded)
	arg.NodeTypeIndicator = boolToNullPtr(u.NodeTypeIndicator)
	arg.AreaRestricted = boolToNullPtr(u.AreaRestricted)
	arg.UeReachableIndicator = boolToNullPtr(u.UeReachableIndicator)
	arg.EpsSubscriptionDataNotNeeded = boolToNullPtr(u.EpsSubscriptionDataNotNeeded)

	// [14] ue-SRVCC-Capability — 0..1 per TS 29.002.
	if u.UeSrvccCapability != nil {
		if *u.UeSrvccCapability < 0 || *u.UeSrvccCapability > 1 {
			return nil, fmt.Errorf("UeSrvccCapability out of range 0..1: %d", *u.UeSrvccCapability)
		}
		v := *u.UeSrvccCapability
		arg.UeSrvccCapability = &v
	}

	// [15] eplmn-List
	if len(u.EplmnList) > 0 {
		list := make(gsm_map.EPLMNList, len(u.EplmnList))
		for i, raw := range u.EplmnList {
			if len(raw) != 3 {
				return nil, fmt.Errorf("UpdateGprsLocation: EplmnList[%d] PLMNId must be exactly 3 octets, got %d", i, len(raw))
			}
			list[i] = gsm_map.PLMNId(raw)
		}
		arg.EplmnList = list
	}

	// [16] mme-Number-for-MT-SMS
	if u.MmeNumberForMTSMS != "" {
		mme, err := encodeAddressField(u.MmeNumberForMTSMS, u.MmeNumberForMTSMSNature, u.MmeNumberForMTSMSPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding MmeNumberForMTSMS: %w", err)
		}
		v := gsm_map.ISDNAddressString(mme)
		arg.MmeNumberforMTSMS = &v
	}

	// [17] smsRegisterRequest — 0..2 per TS 29.002.
	if u.SmsRegisterRequest != nil {
		if *u.SmsRegisterRequest < 0 || *u.SmsRegisterRequest > 2 {
			return nil, fmt.Errorf("SmsRegisterRequest out of range 0..2: %d", *u.SmsRegisterRequest)
		}
		v := *u.SmsRegisterRequest
		arg.SmsRegisterRequest = &v
	}

	arg.SmsOnly = boolToNullPtr(u.SmsOnly)

	// [19]/[20] DiameterIdentity
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
		list := make(gsm_map.AdjacentPLMNList, len(u.AdjacentPLMNList))
		for i, raw := range u.AdjacentPLMNList {
			if len(raw) != 3 {
				return nil, fmt.Errorf("UpdateGprsLocation: AdjacentPLMNList[%d] PLMNId must be exactly 3 octets, got %d", i, len(raw))
			}
			list[i] = gsm_map.PLMNId(raw)
		}
		arg.AdjacentPLMNList = list
	}

	return arg, nil
}

func convertArgToUpdateGprsLocation(arg *gsm_map.UpdateGprsLocationArg) (*UpdateGprsLocation, error) {
	if len(arg.Imsi) == 0 {
		return nil, fmt.Errorf("UpdateGprsLocation: IMSI is mandatory and must be non-empty")
	}
	imsi, err := tbcd.Decode(arg.Imsi)
	if err != nil {
		return nil, fmt.Errorf("decoding IMSI: %w", err)
	}
	if imsi == "" {
		return nil, fmt.Errorf("UpdateGprsLocation: IMSI decoded to empty string")
	}

	sgsnNum, sgsnNature, sgsnPlan, err := decodeAddressField(arg.SgsnNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding SGSNNumber: %w", err)
	}
	if sgsnNum == "" {
		return nil, fmt.Errorf("UpdateGprsLocation: SGSNNumber decoded to empty string")
	}

	sgsnAddr, err := gsn.Parse(arg.SgsnAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding SGSNAddress: %w", err)
	}

	u := &UpdateGprsLocation{
		IMSI:        imsi,
		SGSNNumber:  sgsnNum,
		SGSNNature:  sgsnNature,
		SGSNPlan:    sgsnPlan,
		SGSNAddress: sgsnAddr,
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

	if arg.UsedRATType != nil {
		v := *arg.UsedRATType
		if v < 0 || v > 5 {
			return nil, fmt.Errorf("UsedRATType out of range 0..5: %d", v)
		}
		u.UsedRatType = &v
	}

	u.GprsSubscriptionDataNotNeeded = nullPtrToBool(arg.GprsSubscriptionDataNotNeeded)
	u.NodeTypeIndicator = nullPtrToBool(arg.NodeTypeIndicator)
	u.AreaRestricted = nullPtrToBool(arg.AreaRestricted)
	u.UeReachableIndicator = nullPtrToBool(arg.UeReachableIndicator)
	u.EpsSubscriptionDataNotNeeded = nullPtrToBool(arg.EpsSubscriptionDataNotNeeded)

	if arg.UeSrvccCapability != nil {
		v := *arg.UeSrvccCapability
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("UeSrvccCapability out of range 0..1: %d", v)
		}
		u.UeSrvccCapability = &v
	}

	if len(arg.EplmnList) > 0 {
		list := make([]HexBytes, len(arg.EplmnList))
		for i, plmn := range arg.EplmnList {
			if len(plmn) != 3 {
				return nil, fmt.Errorf("UpdateGprsLocation: EplmnList[%d] PLMNId must be exactly 3 octets, got %d", i, len(plmn))
			}
			list[i] = HexBytes(plmn)
		}
		u.EplmnList = list
	}

	if arg.MmeNumberforMTSMS != nil {
		mme, nature, plan, err := decodeAddressField(*arg.MmeNumberforMTSMS)
		if err != nil {
			return nil, fmt.Errorf("decoding MmeNumberForMTSMS: %w", err)
		}
		u.MmeNumberForMTSMS = mme
		u.MmeNumberForMTSMSNature = nature
		u.MmeNumberForMTSMSPlan = plan
	}

	if arg.SmsRegisterRequest != nil {
		v := *arg.SmsRegisterRequest
		if v < 0 || v > 2 {
			return nil, fmt.Errorf("SmsRegisterRequest out of range 0..2: %d", v)
		}
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

	if len(arg.AdjacentPLMNList) > 0 {
		list := make([]HexBytes, len(arg.AdjacentPLMNList))
		for i, plmn := range arg.AdjacentPLMNList {
			if len(plmn) != 3 {
				return nil, fmt.Errorf("UpdateGprsLocation: AdjacentPLMNList[%d] PLMNId must be exactly 3 octets, got %d", i, len(plmn))
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

	if s.SupportedFeaturesBits > 0 {
		if len(s.SupportedFeatures) == 0 || s.SupportedFeaturesBits > len(s.SupportedFeatures)*8 {
			return nil, fmt.Errorf("SGSNCapability: SupportedFeaturesBits (%d) inconsistent with bytes (%d)", s.SupportedFeaturesBits, len(s.SupportedFeatures))
		}
		bs := runtime.BitString{Bytes: append([]byte(nil), s.SupportedFeatures...), BitLength: s.SupportedFeaturesBits}
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

	if s.ExtSupportedFeaturesBits > 0 {
		if len(s.ExtSupportedFeatures) == 0 || s.ExtSupportedFeaturesBits > len(s.ExtSupportedFeatures)*8 {
			return nil, fmt.Errorf("SGSNCapability: ExtSupportedFeaturesBits (%d) inconsistent with bytes (%d)", s.ExtSupportedFeaturesBits, len(s.ExtSupportedFeatures))
		}
		bs := runtime.BitString{Bytes: append([]byte(nil), s.ExtSupportedFeatures...), BitLength: s.ExtSupportedFeaturesBits}
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

	if w.SupportedCamelPhases != nil && w.SupportedCamelPhases.BitLength > 0 {
		out.SupportedCamelPhases = convertBitStringToCamelPhases(*w.SupportedCamelPhases)
	}

	if w.SupportedLCSCapabilitySets != nil && w.SupportedLCSCapabilitySets.BitLength > 0 {
		out.SupportedLCSCapabilitySets = convertBitStringToLCSCaps(*w.SupportedLCSCapabilitySets)
	}

	if w.OfferedCamel4CSIs != nil && w.OfferedCamel4CSIs.BitLength > 0 {
		out.OfferedCamel4CSIs = convertBitStringToOfferedCamel4CSIs(*w.OfferedCamel4CSIs)
	}

	out.SmsCallBarringSupportIndicator = nullPtrToBool(w.SmsCallBarringSupportIndicator)

	if w.SupportedRATTypesIndicator != nil && w.SupportedRATTypesIndicator.BitLength > 0 {
		if w.SupportedRATTypesIndicator.BitLength < 2 || w.SupportedRATTypesIndicator.BitLength > 8 {
			return nil, fmt.Errorf("SGSNCapability: SupportedRATTypes BitLength must be 2..8, got %d", w.SupportedRATTypesIndicator.BitLength)
		}
		out.SupportedRATTypesIndicator = convertBitStringToSupportedRATTypes(*w.SupportedRATTypesIndicator)
	}

	if w.SupportedFeatures != nil && w.SupportedFeatures.BitLength > 0 {
		// BitString capacity must be consistent with Bytes length.
		if int64(w.SupportedFeatures.BitLength) > int64(len(w.SupportedFeatures.Bytes))*8 {
			return nil, fmt.Errorf("SGSNCapability: SupportedFeatures BitLength %d exceeds len(Bytes)*8 = %d",
				w.SupportedFeatures.BitLength, len(w.SupportedFeatures.Bytes)*8)
		}
		out.SupportedFeatures = HexBytes(append([]byte(nil), w.SupportedFeatures.Bytes...))
		out.SupportedFeaturesBits = w.SupportedFeatures.BitLength
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

	if w.ExtSupportedFeatures != nil && w.ExtSupportedFeatures.BitLength > 0 {
		if int64(w.ExtSupportedFeatures.BitLength) > int64(len(w.ExtSupportedFeatures.Bytes))*8 {
			return nil, fmt.Errorf("SGSNCapability: ExtSupportedFeatures BitLength %d exceeds len(Bytes)*8 = %d",
				w.ExtSupportedFeatures.BitLength, len(w.ExtSupportedFeatures.Bytes)*8)
		}
		out.ExtSupportedFeatures = HexBytes(append([]byte(nil), w.ExtSupportedFeatures.Bytes...))
		out.ExtSupportedFeaturesBits = w.ExtSupportedFeatures.BitLength
	}

	return out, nil
}

// --- EpsInfo CHOICE ---

func convertEpsInfoToWire(e *EpsInfo) (*gsm_map.EPSInfo, error) {
	hasPdn := e.PdnGwUpdate != nil
	hasIsr := e.IsrInformationBits > 0
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
	// IsrInformation is BIT STRING (SIZE(1..8)) per TS 29.002.
	if e.IsrInformationBits < 1 || e.IsrInformationBits > 8 {
		return nil, fmt.Errorf("EpsInfo: IsrInformationBits must be 1..8, got %d", e.IsrInformationBits)
	}
	if len(e.IsrInformation) == 0 || e.IsrInformationBits > len(e.IsrInformation)*8 {
		return nil, fmt.Errorf("EpsInfo: IsrInformationBits (%d) inconsistent with bytes (%d)", e.IsrInformationBits, len(e.IsrInformation))
	}
	bs := runtime.BitString{Bytes: append([]byte(nil), e.IsrInformation...), BitLength: e.IsrInformationBits}
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
		// IsrInformation is BIT STRING (SIZE(1..8)) per TS 29.002.
		bits := w.IsrInformation.BitLength
		if bits < 1 || bits > 8 {
			return nil, fmt.Errorf("IsrInformation BitLength must be 1..8, got %d", bits)
		}
		if int64(bits) > int64(len(w.IsrInformation.Bytes))*8 {
			return nil, fmt.Errorf("IsrInformation BitLength %d exceeds len(Bytes)*8 = %d", bits, len(w.IsrInformation.Bytes)*8)
		}
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
		apn := gsm_map.APN(append([]byte(nil), p.APN...))
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
		if *p.ContextID < 1 || *p.ContextID > 50 {
			return nil, fmt.Errorf("ContextId out of range 1..50: %d", *p.ContextID)
		}
		v := gsm_map.ContextId(int64(*p.ContextID))
		out.ContextId = &v
	}
	return out, nil
}

// convertWireToPdnGwUpdate decodes a wire PDNGWUpdate, validating
// ContextId per TS 29.272 (1..50 context range).
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
		v, err := narrowInt64Range(*w.ContextId, 1, 50, "ContextId")
		if err != nil {
			return nil, err
		}
		out.ContextID = &v
	}
	return out, nil
}

func convertPdnGwIdentityToWire(p *PdnGwIdentity) (*gsm_map.PDNGWIdentity, error) {
	if len(p.IPv4Address) > 0 && len(p.IPv4Address) != 4 {
		return nil, fmt.Errorf("PdnGwIdentity: IPv4Address must be exactly 4 octets, got %d", len(p.IPv4Address))
	}
	if len(p.IPv6Address) > 0 && len(p.IPv6Address) != 16 {
		return nil, fmt.Errorf("PdnGwIdentity: IPv6Address must be exactly 16 octets, got %d", len(p.IPv6Address))
	}
	if len(p.IPv4Address) == 0 && len(p.IPv6Address) == 0 && len(p.Name) == 0 {
		return nil, fmt.Errorf("PdnGwIdentity: at least one of IPv4Address, IPv6Address, or Name must be set")
	}
	out := &gsm_map.PDNGWIdentity{}
	if len(p.IPv4Address) > 0 {
		v := gsm_map.PDPAddress(append([]byte(nil), p.IPv4Address...))
		out.PdnGwIpv4Address = &v
	}
	if len(p.IPv6Address) > 0 {
		v := gsm_map.PDPAddress(append([]byte(nil), p.IPv6Address...))
		out.PdnGwIpv6Address = &v
	}
	if len(p.Name) > 0 {
		v := gsm_map.FQDN(append([]byte(nil), p.Name...))
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
			return nil, fmt.Errorf("PdnGwIdentity: IPv4Address must be exactly 4 octets, got %d", len(ip4))
		}
		out.IPv4Address = HexBytes(ip4)
	}
	if w.PdnGwIpv6Address != nil {
		ip6 := append([]byte(nil), (*w.PdnGwIpv6Address)...)
		if len(ip6) != 16 {
			return nil, fmt.Errorf("PdnGwIdentity: IPv6Address must be exactly 16 octets, got %d", len(ip6))
		}
		out.IPv6Address = HexBytes(ip6)
	}
	if w.PdnGwName != nil {
		out.Name = HexBytes(append([]byte(nil), (*w.PdnGwName)...))
	}
	if len(out.IPv4Address) == 0 && len(out.IPv6Address) == 0 && len(out.Name) == 0 {
		return nil, fmt.Errorf("PdnGwIdentity: at least one of IPv4Address, IPv6Address, or Name must be present")
	}
	return out, nil
}

// --- UpdateGprsLocationRes ---

func convertUpdateGprsLocationResToRes(u *UpdateGprsLocationRes) (*gsm_map.UpdateGprsLocationRes, error) {
	if u.HLRNumber == "" {
		return nil, fmt.Errorf("UpdateGprsLocationRes: HLRNumber is mandatory and must be non-empty")
	}
	hlr, err := encodeAddressField(u.HLRNumber, u.HLRNumberNature, u.HLRNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding HLRNumber: %w", err)
	}

	return &gsm_map.UpdateGprsLocationRes{
		HlrNumber:                  gsm_map.ISDNAddressString(hlr),
		AddCapability:              boolToNullPtr(u.AddCapability),
		SgsnMmeSeparationSupported: boolToNullPtr(u.SgsnMmeSeparationSupported),
		MmeRegisteredforSMS:        boolToNullPtr(u.MmeRegisteredforSMS),
	}, nil
}

func convertResToUpdateGprsLocationRes(res *gsm_map.UpdateGprsLocationRes) (*UpdateGprsLocationRes, error) {
	hlr, nature, plan, err := decodeAddressField(res.HlrNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding HLRNumber: %w", err)
	}
	if hlr == "" {
		return nil, fmt.Errorf("UpdateGprsLocationRes: HLRNumber is mandatory and must be non-empty")
	}

	return &UpdateGprsLocationRes{
		HLRNumber:                  hlr,
		HLRNumberNature:            nature,
		HLRNumberPlan:              plan,
		AddCapability:              nullPtrToBool(res.AddCapability),
		SgsnMmeSeparationSupported: nullPtrToBool(res.SgsnMmeSeparationSupported),
		MmeRegisteredforSMS:        nullPtrToBool(res.MmeRegisteredforSMS),
	}, nil
}
