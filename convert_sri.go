package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// --- SRI (SendRoutingInfo) full converters ---

func validateSri(s *Sri) error {
	if s.MSISDN == "" {
		return ErrSriMissingMSISDN
	}
	if s.GmscOrGsmSCFAddress == "" {
		return ErrSriMissingGmsc
	}
	if s.InterrogationType != InterrogationBasicCall && s.InterrogationType != InterrogationForwarding {
		return ErrSriInvalidInterrogationType
	}
	if s.NumberOfForwarding != nil {
		if *s.NumberOfForwarding < 1 || *s.NumberOfForwarding > 5 {
			return ErrSriInvalidNumberOfForwarding
		}
	}
	if s.OrCapability != nil {
		if *s.OrCapability < 1 || *s.OrCapability > 127 {
			return ErrSriInvalidOrCapability
		}
	}
	if len(s.CallReferenceNumber) > 8 {
		return ErrSriInvalidCallReferenceNumber
	}
	return nil
}

func convertSriToArg(s *Sri) (*gsm_map.SendRoutingInfoArg, error) {
	if err := validateSri(s); err != nil {
		return nil, err
	}

	msisdn, err := encodeAddressField(s.MSISDN, s.MSISDNNature, s.MSISDNPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding MSISDN: %w", err)
	}
	gmsc, err := encodeAddressField(s.GmscOrGsmSCFAddress, s.GmscNature, s.GmscPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding GmscOrGsmSCFAddress: %w", err)
	}

	arg := &gsm_map.SendRoutingInfoArg{
		Msisdn:              gsm_map.ISDNAddressString(msisdn),
		InterrogationType:   gsm_map.InterrogationType(int64(s.InterrogationType)),
		GmscOrGsmSCFAddress: gsm_map.ISDNAddressString(gmsc),
	}

	// CugCheckInfo
	if s.CugCheckInfo != nil {
		arg.CugCheckInfo = convertCugCheckInfoToWire(s.CugCheckInfo)
	}

	// NumberOfForwarding
	if s.NumberOfForwarding != nil {
		v := int64(*s.NumberOfForwarding)
		arg.NumberOfForwarding = &v
	}

	// OrInterrogation
	arg.OrInterrogation = boolToNullPtr(s.OrInterrogation)

	// OrCapability
	if s.OrCapability != nil {
		v := int64(*s.OrCapability)
		arg.OrCapability = &v
	}

	// CallReferenceNumber
	if len(s.CallReferenceNumber) > 0 {
		v := gsm_map.CallReferenceNumber(s.CallReferenceNumber)
		arg.CallReferenceNumber = &v
	}

	// ForwardingReason
	if s.ForwardingReason != nil {
		v := gsm_map.ForwardingReason(int64(*s.ForwardingReason))
		arg.ForwardingReason = &v
	}

	// BasicServiceGroup
	if s.BasicServiceGroup != nil {
		bsg, err := convertExtBasicServiceCodeToWire(s.BasicServiceGroup)
		if err != nil {
			return nil, fmt.Errorf("encoding BasicServiceGroup: %w", err)
		}
		arg.BasicServiceGroup = bsg
	}

	// BasicServiceGroup2
	if s.BasicServiceGroup2 != nil {
		bsg2, err := convertExtBasicServiceCodeToWire(s.BasicServiceGroup2)
		if err != nil {
			return nil, fmt.Errorf("encoding BasicServiceGroup2: %w", err)
		}
		arg.BasicServiceGroup2 = bsg2
	}

	// NetworkSignalInfo
	if s.NetworkSignalInfo != nil {
		arg.NetworkSignalInfo = convertExternalSignalInfoToWire(s.NetworkSignalInfo)
	}

	// NetworkSignalInfo2
	if s.NetworkSignalInfo2 != nil {
		arg.NetworkSignalInfo2 = convertExternalSignalInfoToWire(s.NetworkSignalInfo2)
	}

	// CamelInfo
	if s.CamelInfo != nil {
		arg.CamelInfo = convertSriCamelInfoToWire(s.CamelInfo)
	}

	// SuppressionOfAnnouncement
	if s.SuppressionOfAnnouncement {
		v := gsm_map.SuppressionOfAnnouncement{}
		arg.SuppressionOfAnnouncement = &v
	}

	// AlertingPattern
	if len(s.AlertingPattern) > 0 {
		v := gsm_map.AlertingPattern(s.AlertingPattern)
		arg.AlertingPattern = &v
	}

	// CcbsCall
	arg.CcbsCall = boolToNullPtr(s.CcbsCall)

	// SupportedCCBSPhase
	if s.SupportedCCBSPhase != nil {
		v := int64(*s.SupportedCCBSPhase)
		arg.SupportedCCBSPhase = &v
	}

	// AdditionalSignalInfo
	if s.AdditionalSignalInfo != nil {
		arg.AdditionalSignalInfo = convertExtExternalSignalInfoToWire(s.AdditionalSignalInfo)
	}

	// IstSupportIndicator
	if s.IstSupportIndicator != nil {
		if *s.IstSupportIndicator < 0 || *s.IstSupportIndicator > 1 {
			return nil, fmt.Errorf("IstSupportIndicator out of range 0..1: %d", *s.IstSupportIndicator)
		}
		v := gsm_map.ISTSupportIndicator(int64(*s.IstSupportIndicator))
		arg.IstSupportIndicator = &v
	}

	// PrePagingSupported
	arg.PrePagingSupported = boolToNullPtr(s.PrePagingSupported)

	// CallDiversionTreatmentIndicator
	if len(s.CallDiversionTreatmentIndicator) > 0 {
		v := gsm_map.CallDiversionTreatmentIndicator(s.CallDiversionTreatmentIndicator)
		arg.CallDiversionTreatmentIndicator = &v
	}

	// LongFTNSupported
	arg.LongFTNSupported = boolToNullPtr(s.LongFTNSupported)

	// SuppressVTCSI
	arg.SuppressVTCSI = boolToNullPtr(s.SuppressVTCSI)

	// SuppressIncomingCallBarring
	arg.SuppressIncomingCallBarring = boolToNullPtr(s.SuppressIncomingCallBarring)

	// GsmSCFInitiatedCall
	arg.GsmSCFInitiatedCall = boolToNullPtr(s.GsmSCFInitiatedCall)

	// SuppressMTSS
	if s.SuppressMTSS != nil {
		v := convertSuppressMTSSToBitString(s.SuppressMTSS)
		arg.SuppressMTSS = &v
	}

	// MtRoamingRetrySupported
	arg.MtRoamingRetrySupported = boolToNullPtr(s.MtRoamingRetrySupported)

	// CallPriority
	if s.CallPriority != nil {
		v := int64(*s.CallPriority)
		arg.CallPriority = &v
	}

	return arg, nil
}

func convertArgToSri(arg *gsm_map.SendRoutingInfoArg) (*Sri, error) {
	msisdn, msisdnNature, msisdnPlan, err := decodeAddressField(arg.Msisdn)
	if err != nil {
		return nil, fmt.Errorf("decoding MSISDN: %w", err)
	}

	gmsc, gmscNature, gmscPlan, err := decodeAddressField(arg.GmscOrGsmSCFAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding GmscOrGsmSCFAddress: %w", err)
	}

	// InterrogationType — 0 (basicCall) or 1 (forwarding) per TS 29.002.
	it, err := narrowInt64Range(int64(arg.InterrogationType), 0, 1, "InterrogationType")
	if err != nil {
		return nil, err
	}

	s := &Sri{
		MSISDN:              msisdn,
		MSISDNNature:        msisdnNature,
		MSISDNPlan:          msisdnPlan,
		InterrogationType:   InterrogationType(it),
		GmscOrGsmSCFAddress: gmsc,
		GmscNature:          gmscNature,
		GmscPlan:            gmscPlan,
	}

	// CugCheckInfo
	if arg.CugCheckInfo != nil {
		s.CugCheckInfo = convertWireToCugCheckInfo(arg.CugCheckInfo)
	}

	// NumberOfForwarding — 1..5 per TS 29.002.
	if arg.NumberOfForwarding != nil {
		v, err := narrowInt64Range(*arg.NumberOfForwarding, 1, 5, "NumberOfForwarding")
		if err != nil {
			return nil, err
		}
		s.NumberOfForwarding = &v
	}

	// OrInterrogation
	s.OrInterrogation = nullPtrToBool(arg.OrInterrogation)

	// OrCapability — 1..127 per TS 29.002.
	if arg.OrCapability != nil {
		v, err := narrowInt64Range(*arg.OrCapability, 1, 127, "OrCapability")
		if err != nil {
			return nil, err
		}
		s.OrCapability = &v
	}

	// CallReferenceNumber — OCTET STRING (SIZE(1..8)) per TS 29.002.
	if arg.CallReferenceNumber != nil {
		if len(*arg.CallReferenceNumber) > 8 {
			return nil, fmt.Errorf("CallReferenceNumber must be 1..8 octets, got %d", len(*arg.CallReferenceNumber))
		}
		s.CallReferenceNumber = HexBytes(*arg.CallReferenceNumber)
	}

	// ForwardingReason — 0..2 per TS 29.002.
	if arg.ForwardingReason != nil {
		v, err := narrowInt64Range(int64(*arg.ForwardingReason), 0, 2, "ForwardingReason")
		if err != nil {
			return nil, err
		}
		fr := ForwardingReason(v)
		s.ForwardingReason = &fr
	}

	// BasicServiceGroup
	if arg.BasicServiceGroup != nil {
		bsg, err := convertWireToExtBasicServiceCode(arg.BasicServiceGroup)
		if err != nil {
			return nil, fmt.Errorf("decoding BasicServiceGroup: %w", err)
		}
		s.BasicServiceGroup = bsg
	}

	// BasicServiceGroup2
	if arg.BasicServiceGroup2 != nil {
		bsg2, err := convertWireToExtBasicServiceCode(arg.BasicServiceGroup2)
		if err != nil {
			return nil, fmt.Errorf("decoding BasicServiceGroup2: %w", err)
		}
		s.BasicServiceGroup2 = bsg2
	}

	// NetworkSignalInfo
	if arg.NetworkSignalInfo != nil {
		s.NetworkSignalInfo = convertWireToExternalSignalInfo(arg.NetworkSignalInfo)
	}

	// NetworkSignalInfo2
	if arg.NetworkSignalInfo2 != nil {
		s.NetworkSignalInfo2 = convertWireToExternalSignalInfo(arg.NetworkSignalInfo2)
	}

	// CamelInfo
	if arg.CamelInfo != nil {
		s.CamelInfo = convertWireToSriCamelInfo(arg.CamelInfo)
	}

	// SuppressionOfAnnouncement
	s.SuppressionOfAnnouncement = arg.SuppressionOfAnnouncement != nil

	// AlertingPattern
	if arg.AlertingPattern != nil {
		s.AlertingPattern = HexBytes(*arg.AlertingPattern)
	}

	// CcbsCall
	s.CcbsCall = nullPtrToBool(arg.CcbsCall)

	// SupportedCCBSPhase — 1..5 per TS 29.002 (CCBS phase 1..5).
	if arg.SupportedCCBSPhase != nil {
		v, err := narrowInt64Range(*arg.SupportedCCBSPhase, 1, 5, "SupportedCCBSPhase")
		if err != nil {
			return nil, err
		}
		s.SupportedCCBSPhase = &v
	}

	// AdditionalSignalInfo
	if arg.AdditionalSignalInfo != nil {
		s.AdditionalSignalInfo = convertWireToExtExternalSignalInfo(arg.AdditionalSignalInfo)
	}

	// IstSupportIndicator — 0..1 per TS 29.002.
	if arg.IstSupportIndicator != nil {
		v, err := narrowInt64Range(int64(*arg.IstSupportIndicator), 0, 1, "IstSupportIndicator")
		if err != nil {
			return nil, err
		}
		s.IstSupportIndicator = &v
	}

	// PrePagingSupported
	s.PrePagingSupported = nullPtrToBool(arg.PrePagingSupported)

	// CallDiversionTreatmentIndicator
	if arg.CallDiversionTreatmentIndicator != nil {
		s.CallDiversionTreatmentIndicator = HexBytes(*arg.CallDiversionTreatmentIndicator)
	}

	// LongFTNSupported
	s.LongFTNSupported = nullPtrToBool(arg.LongFTNSupported)

	// SuppressVTCSI
	s.SuppressVTCSI = nullPtrToBool(arg.SuppressVTCSI)

	// SuppressIncomingCallBarring
	s.SuppressIncomingCallBarring = nullPtrToBool(arg.SuppressIncomingCallBarring)

	// GsmSCFInitiatedCall
	s.GsmSCFInitiatedCall = nullPtrToBool(arg.GsmSCFInitiatedCall)

	// SuppressMTSS
	if arg.SuppressMTSS != nil && arg.SuppressMTSS.BitLength > 0 {
		s.SuppressMTSS = convertBitStringToSuppressMTSS(*arg.SuppressMTSS)
	}

	// MtRoamingRetrySupported
	s.MtRoamingRetrySupported = nullPtrToBool(arg.MtRoamingRetrySupported)

	// CallPriority — EMLPP-Priority 0..15 per TS 29.002.
	if arg.CallPriority != nil {
		v, err := narrowInt64Range(int64(*arg.CallPriority), 0, 15, "CallPriority")
		if err != nil {
			return nil, err
		}
		s.CallPriority = &v
	}

	return s, nil
}

// --- SRI Response (SendRoutingInfoRes) full converters ---

func convertSriRespToRes(s *SriResp) (*gsm_map.SendRoutingInfoRes, error) {
	imsiBytes, err := tbcd.Encode(s.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	out := &gsm_map.SendRoutingInfoRes{
		Imsi: (*gsm_map.IMSI)(&imsiBytes),
	}

	// ExtendedRoutingInfo
	if s.ExtendedRoutingInfo != nil {
		eri, err := convertExtendedRoutingInfoToWire(s.ExtendedRoutingInfo)
		if err != nil {
			return nil, fmt.Errorf("encoding ExtendedRoutingInfo: %w", err)
		}
		out.ExtendedRoutingInfo = eri
	}

	// CugCheckInfo
	if s.CugCheckInfo != nil {
		out.CugCheckInfo = convertCugCheckInfoToWire(s.CugCheckInfo)
	}

	// CugSubscriptionFlag
	out.CugSubscriptionFlag = boolToNullPtr(s.CugSubscriptionFlag)

	// SubscriberInfo
	if s.SubscriberInfo != nil {
		si, err := convertSubscriberInfoToWire(s.SubscriberInfo)
		if err != nil {
			return nil, fmt.Errorf("encoding SubscriberInfo: %w", err)
		}
		out.SubscriberInfo = si
	}

	// SsList
	if len(s.SsList) > 0 {
		out.SsList = make(gsm_map.SSList, len(s.SsList))
		for i, c := range s.SsList {
			out.SsList[i] = gsm_map.SSCode{byte(c)}
		}
	}

	// BasicService
	if s.BasicService != nil {
		bs, err := convertExtBasicServiceCodeToWire(s.BasicService)
		if err != nil {
			return nil, fmt.Errorf("encoding BasicService: %w", err)
		}
		out.BasicService = bs
	}

	// ForwardingInterrogationRequired
	out.ForwardingInterrogationRequired = boolToNullPtr(s.ForwardingInterrogationRequired)

	// VmscAddress
	if s.VmscAddress != "" {
		enc, err := encodeAddressField(s.VmscAddress, s.VmscNature, s.VmscPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding VmscAddress: %w", err)
		}
		v := gsm_map.ISDNAddressString(enc)
		out.VmscAddress = &v
	}

	// NaeaPreferredCI
	if s.NaeaPreferredCI != nil {
		out.NaeaPreferredCI = convertNaeaPreferredCIToWire(s.NaeaPreferredCI)
	}

	// CcbsIndicators
	if s.CcbsIndicators != nil {
		out.CcbsIndicators = convertCcbsIndicatorsToWire(s.CcbsIndicators)
	}

	// Msisdn
	if s.MSISDN != "" {
		enc, err := encodeAddressField(s.MSISDN, s.MSISDNNature, s.MSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding MSISDN: %w", err)
		}
		v := gsm_map.ISDNAddressString(enc)
		out.Msisdn = &v
	}

	// NumberPortabilityStatus
	if s.NumberPortabilityStatus != nil {
		v := gsm_map.NumberPortabilityStatus(int64(*s.NumberPortabilityStatus))
		out.NumberPortabilityStatus = &v
	}

	// IstAlertTimer
	out.IstAlertTimer = intPtrTo64(s.IstAlertTimer)

	// SupportedCamelPhasesInVMSC
	if s.SupportedCamelPhasesInVMSC != nil {
		bs := convertCamelPhasesToBitString(s.SupportedCamelPhasesInVMSC)
		out.SupportedCamelPhasesInVMSC = &bs
	}

	// OfferedCamel4CSIsInVMSC
	if s.OfferedCamel4CSIsInVMSC != nil {
		bs := convertOfferedCamel4CSIsToBitString(s.OfferedCamel4CSIsInVMSC)
		out.OfferedCamel4CSIsInVMSC = &bs
	}

	// RoutingInfo2
	if s.RoutingInfo2 != nil {
		ri, err := convertRoutingInfoToWire(s.RoutingInfo2)
		if err != nil {
			return nil, fmt.Errorf("encoding RoutingInfo2: %w", err)
		}
		out.RoutingInfo2 = ri
	}

	// SsList2
	if len(s.SsList2) > 0 {
		out.SsList2 = make(gsm_map.SSList, len(s.SsList2))
		for i, c := range s.SsList2 {
			out.SsList2[i] = gsm_map.SSCode{byte(c)}
		}
	}

	// BasicService2
	if s.BasicService2 != nil {
		bs2, err := convertExtBasicServiceCodeToWire(s.BasicService2)
		if err != nil {
			return nil, fmt.Errorf("encoding BasicService2: %w", err)
		}
		out.BasicService2 = bs2
	}

	// AllowedServices
	if s.AllowedServices != nil {
		bs := convertAllowedServicesToBitString(s.AllowedServices)
		out.AllowedServices = &bs
	}

	// UnavailabilityCause
	if s.UnavailabilityCause != nil {
		v := gsm_map.UnavailabilityCause(int64(*s.UnavailabilityCause))
		out.UnavailabilityCause = &v
	}

	// ReleaseResourcesSupported
	out.ReleaseResourcesSupported = boolToNullPtr(s.ReleaseResourcesSupported)

	// GsmBearerCapability
	if s.GsmBearerCapability != nil {
		out.GsmBearerCapability = convertExternalSignalInfoToWire(s.GsmBearerCapability)
	}

	return out, nil
}

func convertResToSriResp(res *gsm_map.SendRoutingInfoRes) (*SriResp, error) {
	out := &SriResp{}

	// Imsi
	if res.Imsi != nil {
		imsi, err := tbcd.Decode(*res.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding IMSI: %w", err)
		}
		out.IMSI = imsi
	}

	// ExtendedRoutingInfo
	if res.ExtendedRoutingInfo != nil {
		eri, err := convertWireToExtendedRoutingInfo(res.ExtendedRoutingInfo)
		if err != nil {
			return nil, fmt.Errorf("decoding ExtendedRoutingInfo: %w", err)
		}
		out.ExtendedRoutingInfo = eri
	}

	// CugCheckInfo
	if res.CugCheckInfo != nil {
		out.CugCheckInfo = convertWireToCugCheckInfo(res.CugCheckInfo)
	}

	// CugSubscriptionFlag
	out.CugSubscriptionFlag = nullPtrToBool(res.CugSubscriptionFlag)

	// SubscriberInfo
	if res.SubscriberInfo != nil {
		si, err := convertWireToSubscriberInfo(res.SubscriberInfo)
		if err != nil {
			return nil, fmt.Errorf("decoding SubscriberInfo: %w", err)
		}
		out.SubscriberInfo = si
	}

	// SsList — each SS-Code is OCTET STRING (SIZE(1)) per 3GPP TS 29.002.
	if len(res.SsList) > 0 {
		out.SsList = make([]SsCode, len(res.SsList))
		for i, c := range res.SsList {
			if len(c) != 1 {
				return nil, fmt.Errorf("SsList[%d]: SS-Code must be exactly 1 octet, got %d", i, len(c))
			}
			out.SsList[i] = SsCode(c[0])
		}
	}

	// BasicService
	if res.BasicService != nil {
		bs, err := convertWireToExtBasicServiceCode(res.BasicService)
		if err != nil {
			return nil, fmt.Errorf("decoding BasicService: %w", err)
		}
		out.BasicService = bs
	}

	// ForwardingInterrogationRequired
	out.ForwardingInterrogationRequired = nullPtrToBool(res.ForwardingInterrogationRequired)

	// VmscAddress
	if res.VmscAddress != nil {
		digits, nat, pl, err := decodeAddressField(*res.VmscAddress)
		if err != nil {
			return nil, fmt.Errorf("decoding VmscAddress: %w", err)
		}
		out.VmscAddress = digits
		out.VmscNature = nat
		out.VmscPlan = pl
	}

	// NaeaPreferredCI
	if res.NaeaPreferredCI != nil {
		out.NaeaPreferredCI = convertWireToNaeaPreferredCI(res.NaeaPreferredCI)
	}

	// CcbsIndicators
	if res.CcbsIndicators != nil {
		out.CcbsIndicators = convertWireToCcbsIndicators(res.CcbsIndicators)
	}

	// Msisdn
	if res.Msisdn != nil {
		digits, nat, pl, err := decodeAddressField(*res.Msisdn)
		if err != nil {
			return nil, fmt.Errorf("decoding MSISDN: %w", err)
		}
		out.MSISDN = digits
		out.MSISDNNature = nat
		out.MSISDNPlan = pl
	}

	// NumberPortabilityStatus — defined values 0,1,2,4,5 per TS 29.002.
	if res.NumberPortabilityStatus != nil {
		v64 := int64(*res.NumberPortabilityStatus)
		switch NumberPortabilityStatus(v64) {
		case MnpNotKnownToBePorted, MnpOwnNumberPortedOut, MnpForeignNumberPortedToForeignNetwork,
			MnpOwnNumberNotPortedOut, MnpForeignNumberPortedIn:
		default:
			return nil, fmt.Errorf("NumberPortabilityStatus has undefined value %d", v64)
		}
		v := NumberPortabilityStatus(v64)
		out.NumberPortabilityStatus = &v
	}

	// IstAlertTimer
	istAlert, err := int64PtrTo(res.IstAlertTimer)
	if err != nil {
		return nil, fmt.Errorf("decoding IstAlertTimer: %w", err)
	}
	out.IstAlertTimer = istAlert

	// SupportedCamelPhasesInVMSC
	if res.SupportedCamelPhasesInVMSC != nil && res.SupportedCamelPhasesInVMSC.BitLength > 0 {
		out.SupportedCamelPhasesInVMSC = convertBitStringToCamelPhases(*res.SupportedCamelPhasesInVMSC)
	}

	// OfferedCamel4CSIsInVMSC
	if res.OfferedCamel4CSIsInVMSC != nil && res.OfferedCamel4CSIsInVMSC.BitLength > 0 {
		out.OfferedCamel4CSIsInVMSC = convertBitStringToOfferedCamel4CSIs(*res.OfferedCamel4CSIsInVMSC)
	}

	// RoutingInfo2
	if res.RoutingInfo2 != nil {
		ri, err := convertWireToRoutingInfo(res.RoutingInfo2)
		if err != nil {
			return nil, fmt.Errorf("decoding RoutingInfo2: %w", err)
		}
		out.RoutingInfo2 = ri
	}

	// SsList2
	if len(res.SsList2) > 0 {
		out.SsList2 = make([]SsCode, len(res.SsList2))
		for i, c := range res.SsList2 {
			if len(c) != 1 {
				return nil, fmt.Errorf("SsList2[%d]: SS-Code must be exactly 1 octet, got %d", i, len(c))
			}
			out.SsList2[i] = SsCode(c[0])
		}
	}

	// BasicService2
	if res.BasicService2 != nil {
		bs2, err := convertWireToExtBasicServiceCode(res.BasicService2)
		if err != nil {
			return nil, fmt.Errorf("decoding BasicService2: %w", err)
		}
		out.BasicService2 = bs2
	}

	// AllowedServices
	if res.AllowedServices != nil && res.AllowedServices.BitLength > 0 {
		out.AllowedServices = convertBitStringToAllowedServices(*res.AllowedServices)
	}

	// UnavailabilityCause — 1..6 per TS 29.002.
	if res.UnavailabilityCause != nil {
		v, err := narrowInt64Range(int64(*res.UnavailabilityCause), 1, 6, "UnavailabilityCause")
		if err != nil {
			return nil, err
		}
		uc := UnavailabilityCause(v)
		out.UnavailabilityCause = &uc
	}

	// ReleaseResourcesSupported
	out.ReleaseResourcesSupported = nullPtrToBool(res.ReleaseResourcesSupported)

	// GsmBearerCapability
	if res.GsmBearerCapability != nil {
		out.GsmBearerCapability = convertWireToExternalSignalInfo(res.GsmBearerCapability)
	}

	return out, nil
}
