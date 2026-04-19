package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// --- SRI-SM helper converters ---

func convertAdditionalNumberToWire(a *AdditionalNumber) (*gsm_map.AdditionalNumber, error) {
	hasMsc := a.MscNumber != ""
	hasSgsn := a.SgsnNumber != ""
	switch {
	case hasMsc && hasSgsn:
		return nil, ErrSriChoiceMultipleAlternatives
	case hasMsc:
		encoded, err := encodeAddressField(a.MscNumber, a.MscNumberNature, a.MscNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding MscNumber: %w", err)
		}
		v := gsm_map.NewAdditionalNumberMscNumber(gsm_map.ISDNAddressString(encoded))
		return &v, nil
	case hasSgsn:
		encoded, err := encodeAddressField(a.SgsnNumber, a.SgsnNumberNature, a.SgsnNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SgsnNumber: %w", err)
		}
		v := gsm_map.NewAdditionalNumberSgsnNumber(gsm_map.ISDNAddressString(encoded))
		return &v, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertWireToAdditionalNumber(w *gsm_map.AdditionalNumber) (*AdditionalNumber, error) {
	an := &AdditionalNumber{}
	switch w.Choice {
	case gsm_map.AdditionalNumberChoiceMscNumber:
		if w.MscNumber == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		num, nature, plan, err := decodeAddressField(*w.MscNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding MscNumber: %w", err)
		}
		an.MscNumber = num
		an.MscNumberNature = nature
		an.MscNumberPlan = plan
	case gsm_map.AdditionalNumberChoiceSgsnNumber:
		if w.SgsnNumber == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		num, nature, plan, err := decodeAddressField(*w.SgsnNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding SgsnNumber: %w", err)
		}
		an.SgsnNumber = num
		an.SgsnNumberNature = nature
		an.SgsnNumberPlan = plan
	default:
		return nil, ErrSriChoiceNoAlternative
	}
	return an, nil
}

func convertNetworkNodeDiameterAddressToWire(n *NetworkNodeDiameterAddress) *gsm_map.NetworkNodeDiameterAddress {
	return &gsm_map.NetworkNodeDiameterAddress{
		DiameterName:  gsm_map.DiameterIdentity(n.DiameterName),
		DiameterRealm: gsm_map.DiameterIdentity(n.DiameterRealm),
	}
}

func convertWireToNetworkNodeDiameterAddress(w *gsm_map.NetworkNodeDiameterAddress) *NetworkNodeDiameterAddress {
	return &NetworkNodeDiameterAddress{
		DiameterName:  HexBytes(w.DiameterName),
		DiameterRealm: HexBytes(w.DiameterRealm),
	}
}

func convertCorrelationIDToWire(c *SriSmCorrelationID) (*gsm_map.CorrelationID, error) {
	if len(c.SipUriB) == 0 {
		return nil, ErrSriSmMissingSipUriB
	}
	out := &gsm_map.CorrelationID{
		SipUriB: gsm_map.SIPURI(c.SipUriB),
	}
	if len(c.HlrID) > 0 {
		v := gsm_map.HLRId(c.HlrID)
		out.HlrId = &v
	}
	if len(c.SipUriA) > 0 {
		v := gsm_map.SIPURI(c.SipUriA)
		out.SipUriA = &v
	}
	return out, nil
}

func convertWireToCorrelationID(w *gsm_map.CorrelationID) *SriSmCorrelationID {
	c := &SriSmCorrelationID{
		SipUriB: HexBytes(w.SipUriB),
	}
	if w.HlrId != nil {
		c.HlrID = HexBytes(*w.HlrId)
	}
	if w.SipUriA != nil {
		c.SipUriA = HexBytes(*w.SipUriA)
	}
	return c
}

func convertIpSmGwGuidanceToWire(g *IpSmGwGuidance) (*gsm_map.IPSMGWGuidance, error) {
	if g.MinimumDeliveryTimeValue < MinSmDeliveryTimer || g.MinimumDeliveryTimeValue > MaxSmDeliveryTimer ||
		g.RecommendedDeliveryTimeValue < MinSmDeliveryTimer || g.RecommendedDeliveryTimeValue > MaxSmDeliveryTimer {
		return nil, ErrSriSmInvalidDeliveryTimerValue
	}
	return &gsm_map.IPSMGWGuidance{
		MinimumDeliveryTimeValue:     gsm_map.SMDeliveryTimerValue(g.MinimumDeliveryTimeValue),
		RecommendedDeliveryTimeValue: gsm_map.SMDeliveryTimerValue(g.RecommendedDeliveryTimeValue),
	}, nil
}

func convertWireToIpSmGwGuidance(w *gsm_map.IPSMGWGuidance) *IpSmGwGuidance {
	return &IpSmGwGuidance{
		MinimumDeliveryTimeValue:     int(w.MinimumDeliveryTimeValue),
		RecommendedDeliveryTimeValue: int(w.RecommendedDeliveryTimeValue),
	}
}

// --- UpdateLocation helpers ---

func convertSuperChargerInfoToWire(s *SuperChargerInfo) (*gsm_map.SuperChargerInfo, error) {
	hasSend := s.SendSubscriberData
	hasStored := len(s.SubscriberDataStored) > 0
	if hasSend && hasStored {
		return nil, ErrSuperChargerInfoMultipleAlternatives
	}
	if !hasSend && !hasStored {
		return nil, ErrSuperChargerInfoNoAlternative
	}
	if hasSend {
		v := gsm_map.NewSuperChargerInfoSendSubscriberData(struct{}{})
		return &v, nil
	}
	v := gsm_map.NewSuperChargerInfoSubscriberDataStored(gsm_map.AgeIndicator(s.SubscriberDataStored))
	return &v, nil
}

func convertWireToSuperChargerInfo(w *gsm_map.SuperChargerInfo) (*SuperChargerInfo, error) {
	if w.SendSubscriberData != nil && w.SubscriberDataStored != nil {
		return nil, ErrSuperChargerInfoMultipleAlternatives
	}
	out := &SuperChargerInfo{}
	if w.SendSubscriberData != nil {
		out.SendSubscriberData = true
	} else if w.SubscriberDataStored != nil {
		out.SubscriberDataStored = HexBytes(*w.SubscriberDataStored)
	} else {
		return nil, ErrSuperChargerInfoNoAlternative
	}
	return out, nil
}

func convertAddInfoToWire(a *AddInfo) (*gsm_map.ADDInfo, error) {
	imeisvBytes, err := tbcd.Encode(a.IMEISV)
	if err != nil {
		return nil, fmt.Errorf("encoding IMEISV: %w", err)
	}
	out := &gsm_map.ADDInfo{
		Imeisv:                   gsm_map.IMEI(imeisvBytes),
		SkipSubscriberDataUpdate: boolToNullPtr(a.SkipSubscriberDataUpdate),
	}
	return out, nil
}

func convertWireToAddInfo(w *gsm_map.ADDInfo) (*AddInfo, error) {
	imeisv, err := tbcd.Decode(w.Imeisv)
	if err != nil {
		return nil, fmt.Errorf("decoding IMEISV: %w", err)
	}
	return &AddInfo{
		IMEISV:                   imeisv,
		SkipSubscriberDataUpdate: nullPtrToBool(w.SkipSubscriberDataUpdate),
	}, nil
}

// --- SRI nested SEQUENCE helpers ---

func convertForwardingDataToWire(f *ForwardingData) (*gsm_map.ForwardingData, error) {
	out := &gsm_map.ForwardingData{}
	if f.ForwardedToNumber != "" {
		enc, err := encodeAddressField(f.ForwardedToNumber, f.ForwardedToNumberNature, f.ForwardedToNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding ForwardedToNumber: %w", err)
		}
		as := gsm_map.ISDNAddressString(enc)
		out.ForwardedToNumber = &as
	}
	if len(f.ForwardedToSubaddress) > 0 {
		sa := gsm_map.ISDNSubaddressString(f.ForwardedToSubaddress)
		out.ForwardedToSubaddress = &sa
	}
	if len(f.ForwardingOptions) > 0 {
		fo := gsm_map.ForwardingOptions(f.ForwardingOptions)
		out.ForwardingOptions = &fo
	}
	if len(f.LongForwardedToNumber) > 0 {
		ln := gsm_map.FTNAddressString(f.LongForwardedToNumber)
		out.LongForwardedToNumber = &ln
	}
	return out, nil
}

func convertWireToForwardingData(w *gsm_map.ForwardingData) (*ForwardingData, error) {
	out := &ForwardingData{}
	if w.ForwardedToNumber != nil {
		digits, nat, pl, err := decodeAddressField(*w.ForwardedToNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding ForwardedToNumber: %w", err)
		}
		out.ForwardedToNumber = digits
		out.ForwardedToNumberNature = nat
		out.ForwardedToNumberPlan = pl
	}
	if w.ForwardedToSubaddress != nil {
		out.ForwardedToSubaddress = HexBytes(*w.ForwardedToSubaddress)
	}
	if w.ForwardingOptions != nil {
		out.ForwardingOptions = HexBytes(*w.ForwardingOptions)
	}
	if w.LongForwardedToNumber != nil {
		out.LongForwardedToNumber = HexBytes(*w.LongForwardedToNumber)
	}
	return out, nil
}

func convertCcbsIndicatorsToWire(c *CcbsIndicators) *gsm_map.CCBSIndicators {
	return &gsm_map.CCBSIndicators{
		CcbsPossible:          boolToNullPtr(c.CcbsPossible),
		KeepCCBSCallIndicator: boolToNullPtr(c.KeepCCBSCallIndicator),
	}
}

func convertWireToCcbsIndicators(w *gsm_map.CCBSIndicators) *CcbsIndicators {
	return &CcbsIndicators{
		CcbsPossible:          nullPtrToBool(w.CcbsPossible),
		KeepCCBSCallIndicator: nullPtrToBool(w.KeepCCBSCallIndicator),
	}
}

func convertCugCheckInfoToWire(c *CugCheckInfo) *gsm_map.CUGCheckInfo {
	return &gsm_map.CUGCheckInfo{
		CugInterlock:      gsm_map.CUGInterlock(c.CugInterlock),
		CugOutgoingAccess: boolToNullPtr(c.CugOutgoingAccess),
	}
}

func convertWireToCugCheckInfo(w *gsm_map.CUGCheckInfo) *CugCheckInfo {
	return &CugCheckInfo{
		CugInterlock:      HexBytes(w.CugInterlock),
		CugOutgoingAccess: nullPtrToBool(w.CugOutgoingAccess),
	}
}

// --- SRI CHOICE helpers ---

func convertExtBasicServiceCodeToWire(e *ExtBasicServiceCode) (*gsm_map.ExtBasicServiceCode, error) {
	hasBearer := len(e.ExtBearerService) > 0
	hasTele := len(e.ExtTeleservice) > 0
	switch {
	case hasBearer && hasTele:
		return nil, ErrSriChoiceMultipleAlternatives
	case hasBearer:
		v := gsm_map.NewExtBasicServiceCodeExtBearerService(gsm_map.ExtBearerServiceCode(e.ExtBearerService))
		return &v, nil
	case hasTele:
		v := gsm_map.NewExtBasicServiceCodeExtTeleservice(gsm_map.ExtTeleserviceCode(e.ExtTeleservice))
		return &v, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertWireToExtBasicServiceCode(w *gsm_map.ExtBasicServiceCode) (*ExtBasicServiceCode, error) {
	switch w.Choice {
	case gsm_map.ExtBasicServiceCodeChoiceExtBearerService:
		if w.ExtBearerService == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		return &ExtBasicServiceCode{ExtBearerService: HexBytes(*w.ExtBearerService)}, nil
	case gsm_map.ExtBasicServiceCodeChoiceExtTeleservice:
		if w.ExtTeleservice == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		return &ExtBasicServiceCode{ExtTeleservice: HexBytes(*w.ExtTeleservice)}, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertRoutingInfoToWire(r *RoutingInfo) (*gsm_map.RoutingInfo, error) {
	hasRoaming := r.RoamingNumber != ""
	hasFwd := r.ForwardingData != nil
	switch {
	case hasRoaming && hasFwd:
		return nil, ErrSriChoiceMultipleAlternatives
	case hasRoaming:
		enc, err := encodeAddressField(r.RoamingNumber, r.RoamingNumberNature, r.RoamingNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding RoamingNumber: %w", err)
		}
		v := gsm_map.NewRoutingInfoRoamingNumber(gsm_map.ISDNAddressString(enc))
		return &v, nil
	case hasFwd:
		fw, err := convertForwardingDataToWire(r.ForwardingData)
		if err != nil {
			return nil, err
		}
		v := gsm_map.NewRoutingInfoForwardingData(*fw)
		return &v, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertWireToRoutingInfo(w *gsm_map.RoutingInfo) (*RoutingInfo, error) {
	switch w.Choice {
	case gsm_map.RoutingInfoChoiceRoamingNumber:
		if w.RoamingNumber == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		digits, nat, pl, err := decodeAddressField(*w.RoamingNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding RoamingNumber: %w", err)
		}
		return &RoutingInfo{RoamingNumber: digits, RoamingNumberNature: nat, RoamingNumberPlan: pl}, nil
	case gsm_map.RoutingInfoChoiceForwardingData:
		if w.ForwardingData == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		fd, err := convertWireToForwardingData(w.ForwardingData)
		if err != nil {
			return nil, err
		}
		return &RoutingInfo{ForwardingData: fd}, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertExtendedRoutingInfoToWire(e *ExtendedRoutingInfo) (*gsm_map.ExtendedRoutingInfo, error) {
	hasRI := e.RoutingInfo != nil
	hasCamel := e.CamelRoutingInfo != nil
	switch {
	case hasRI && hasCamel:
		return nil, ErrSriChoiceMultipleAlternatives
	case hasRI:
		ri, err := convertRoutingInfoToWire(e.RoutingInfo)
		if err != nil {
			return nil, err
		}
		v := gsm_map.NewExtendedRoutingInfoRoutingInfo(*ri)
		return &v, nil
	case hasCamel:
		cri, err := convertCamelRoutingInfoToWire(e.CamelRoutingInfo)
		if err != nil {
			return nil, err
		}
		v := gsm_map.NewExtendedRoutingInfoCamelRoutingInfo(*cri)
		return &v, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertWireToExtendedRoutingInfo(w *gsm_map.ExtendedRoutingInfo) (*ExtendedRoutingInfo, error) {
	switch w.Choice {
	case gsm_map.ExtendedRoutingInfoChoiceRoutingInfo:
		if w.RoutingInfo == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		ri, err := convertWireToRoutingInfo(w.RoutingInfo)
		if err != nil {
			return nil, err
		}
		return &ExtendedRoutingInfo{RoutingInfo: ri}, nil
	case gsm_map.ExtendedRoutingInfoChoiceCamelRoutingInfo:
		if w.CamelRoutingInfo == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		cri, err := convertWireToCamelRoutingInfo(w.CamelRoutingInfo)
		if err != nil {
			return nil, err
		}
		return &ExtendedRoutingInfo{CamelRoutingInfo: cri}, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertCamelRoutingInfoToWire(c *CamelRoutingInfo) (*gsm_map.CamelRoutingInfo, error) {
	gi, err := convertGmscCamelSubInfoToWire(&c.GmscCamelSubscriptionInfo)
	if err != nil {
		return nil, fmt.Errorf("GmscCamelSubscriptionInfo: %w", err)
	}
	out := &gsm_map.CamelRoutingInfo{GmscCamelSubscriptionInfo: gi}
	if c.ForwardingData != nil {
		fd, err := convertForwardingDataToWire(c.ForwardingData)
		if err != nil {
			return nil, err
		}
		out.ForwardingData = fd
	}
	return out, nil
}

func convertWireToCamelRoutingInfo(w *gsm_map.CamelRoutingInfo) (*CamelRoutingInfo, error) {
	gi, err := convertWireToGmscCamelSubInfo(&w.GmscCamelSubscriptionInfo)
	if err != nil {
		return nil, fmt.Errorf("GmscCamelSubscriptionInfo: %w", err)
	}
	out := &CamelRoutingInfo{GmscCamelSubscriptionInfo: gi}
	if w.ForwardingData != nil {
		fd, err := convertWireToForwardingData(w.ForwardingData)
		if err != nil {
			return nil, err
		}
		out.ForwardingData = fd
	}
	return out, nil
}

// --- SRI remaining helpers ---

func convertExternalSignalInfoToWire(e *ExternalSignalInfo) *gsm_map.ExternalSignalInfo {
	return &gsm_map.ExternalSignalInfo{
		ProtocolId: gsm_map.ProtocolId(int64(e.ProtocolID)),
		SignalInfo: gsm_map.SignalInfo(e.SignalInfo),
	}
}

func convertWireToExternalSignalInfo(w *gsm_map.ExternalSignalInfo) *ExternalSignalInfo {
	return &ExternalSignalInfo{
		ProtocolID: int(w.ProtocolId),
		SignalInfo: HexBytes(w.SignalInfo),
	}
}

func convertExtExternalSignalInfoToWire(e *ExtExternalSignalInfo) *gsm_map.ExtExternalSignalInfo {
	return &gsm_map.ExtExternalSignalInfo{
		ExtProtocolId: gsm_map.ExtProtocolId(int64(e.ExtProtocolID)),
		SignalInfo:    gsm_map.SignalInfo(e.SignalInfo),
	}
}

func convertWireToExtExternalSignalInfo(w *gsm_map.ExtExternalSignalInfo) *ExtExternalSignalInfo {
	return &ExtExternalSignalInfo{
		ExtProtocolID: int(w.ExtProtocolId),
		SignalInfo:    HexBytes(w.SignalInfo),
	}
}

func convertSriCamelInfoToWire(c *SriCamelInfo) *gsm_map.CamelInfo {
	out := &gsm_map.CamelInfo{
		SupportedCamelPhases: convertCamelPhasesToBitString(&c.SupportedCamelPhases),
		SuppressTCSI:         boolToNullPtr(c.SuppressTCSI),
	}
	if c.OfferedCamel4CSIs != nil {
		bs := convertOfferedCamel4CSIsToBitString(c.OfferedCamel4CSIs)
		out.OfferedCamel4CSIs = &bs
	}
	return out
}

func convertWireToSriCamelInfo(w *gsm_map.CamelInfo) *SriCamelInfo {
	out := &SriCamelInfo{
		SupportedCamelPhases: *convertBitStringToCamelPhases(w.SupportedCamelPhases),
		SuppressTCSI:         nullPtrToBool(w.SuppressTCSI),
	}
	if w.OfferedCamel4CSIs != nil {
		out.OfferedCamel4CSIs = convertBitStringToOfferedCamel4CSIs(*w.OfferedCamel4CSIs)
	}
	return out
}

func convertNaeaPreferredCIToWire(n *NaeaPreferredCI) *gsm_map.NAEAPreferredCI {
	return &gsm_map.NAEAPreferredCI{NaeaPreferredCIC: gsm_map.NAEACIC(n.NaeaPreferredCIC)}
}

func convertWireToNaeaPreferredCI(w *gsm_map.NAEAPreferredCI) *NaeaPreferredCI {
	return &NaeaPreferredCI{NaeaPreferredCIC: HexBytes(w.NaeaPreferredCIC)}
}
