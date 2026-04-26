package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// convertExtSSInfoListToWire / convertWireToExtSSInfoList — list helpers
// for the ProvisionedSS field. Per TS 29.002 the list size is bounded by
// the spec but the package convention is to validate the underlying
// per-entry constraints.

func convertExtSSInfoListToWire(list []ExtSSInfo) (gsm_map.ExtSSInfoList, error) {
	if list == nil {
		return nil, nil
	}
	if len(list) < 1 || len(list) > 30 {
		return nil, fmt.Errorf("provisionedSS (Ext-SS-InfoList): must contain 1..30 entries (got %d)", len(list))
	}
	out := make(gsm_map.ExtSSInfoList, len(list))
	for i, e := range list {
		w, err := convertExtSSInfoToWire(&e)
		if err != nil {
			return nil, fmt.Errorf("ProvisionedSS[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToExtSSInfoList(w gsm_map.ExtSSInfoList) ([]ExtSSInfo, error) {
	if w == nil {
		return nil, nil
	}
	if len(w) < 1 || len(w) > 30 {
		return nil, fmt.Errorf("provisionedSS (Ext-SS-InfoList): must contain 1..30 entries (got %d)", len(w))
	}
	out := make([]ExtSSInfo, len(w))
	for i, e := range w {
		v, err := convertWireToExtSSInfo(&e)
		if err != nil {
			return nil, fmt.Errorf("ProvisionedSS[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

// ============================================================================
// InsertSubscriberDataArg ↔ wire converter
// ============================================================================

func convertInsertSubscriberDataArgToWire(a *InsertSubscriberDataArg) (*gsm_map.InsertSubscriberDataArg, error) {
	if a == nil {
		return nil, fmt.Errorf("InsertSubscriberDataArg: argument must not be nil")
	}
	out := &gsm_map.InsertSubscriberDataArg{
		RoamingRestrictionDueToUnsupportedFeature:      boolToNullPtr(a.RoamingRestrictionDueToUnsupportedFeature),
		RoamingRestrictedInSgsnDueToUnsupportedFeature: boolToNullPtr(a.RoamingRestrictedInSgsnDueToUnsupportedFeature),
		LmuIndicator:                                   boolToNullPtr(a.LmuIndicator),
		UeReachabilityRequestIndicator:                 boolToNullPtr(a.UeReachabilityRequestIndicator),
		VplmnLIPAAllowed:                               boolToNullPtr(a.VplmnLIPAAllowed),
		PsAndSMSOnlyServiceProvision:                   boolToNullPtr(a.PsAndSMSOnlyServiceProvision),
		SmsInSGSNAllowed:                               boolToNullPtr(a.SmsInSGSNAllowed),
		CsToPsSRVCCAllowedIndicator:                    boolToNullPtr(a.CsToPsSRVCCAllowedIndicator),
		PcscfRestorationRequest:                        boolToNullPtr(a.PcscfRestorationRequest),
		UserPlaneIntegrityProtectionIndicator:          boolToNullPtr(a.UserPlaneIntegrityProtectionIndicator),
		IabOperationAllowedIndicator:                   boolToNullPtr(a.IabOperationAllowedIndicator),
	}

	if len(a.IMSI) > 0 {
		v := gsm_map.IMSI(a.IMSI)
		out.Imsi = &v
	}
	if a.MSISDN != "" {
		isdn, err := encodeAddressField(a.MSISDN, a.MSISDNNature, a.MSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding MSISDN: %w", err)
		}
		v := gsm_map.ISDNAddressString(isdn)
		out.Msisdn = &v
	}
	if len(a.Category) > 0 {
		if len(a.Category) != 1 {
			return nil, fmt.Errorf("InsertSubscriberDataArg.Category: must be exactly 1 octet (got %d)", len(a.Category))
		}
		v := gsm_map.Category(a.Category)
		out.Category = &v
	}
	if a.SubscriberStatus != nil {
		v := gsm_map.SubscriberStatus(*a.SubscriberStatus)
		out.SubscriberStatus = &v
	}
	if a.BearerServiceList != nil {
		out.BearerServiceList = make(gsm_map.BearerServiceList, len(a.BearerServiceList))
		for i, b := range a.BearerServiceList {
			if len(b) < 1 || len(b) > 5 {
				return nil, fmt.Errorf("BearerServiceList[%d]: ExtBearerServiceCode must be 1..5 octets (got %d)", i, len(b))
			}
			out.BearerServiceList[i] = gsm_map.ExtBearerServiceCode(b)
		}
	}
	if a.TeleserviceList != nil {
		out.TeleserviceList = make(gsm_map.TeleserviceList, len(a.TeleserviceList))
		for i, t := range a.TeleserviceList {
			if len(t) < 1 || len(t) > 5 {
				return nil, fmt.Errorf("TeleserviceList[%d]: ExtTeleserviceCode must be 1..5 octets (got %d)", i, len(t))
			}
			out.TeleserviceList[i] = gsm_map.ExtTeleserviceCode(t)
		}
	}
	if a.ProvisionedSS != nil {
		l, err := convertExtSSInfoListToWire(a.ProvisionedSS)
		if err != nil {
			return nil, err
		}
		out.ProvisionedSS = l
	}
	if a.OdbData != nil {
		w, err := convertODBDataToWire(a.OdbData)
		if err != nil {
			return nil, fmt.Errorf("OdbData: %w", err)
		}
		out.OdbData = w
	}
	if a.RegionalSubscriptionData != nil {
		l, err := convertZoneCodeListToWire(a.RegionalSubscriptionData)
		if err != nil {
			return nil, err
		}
		out.RegionalSubscriptionData = l
	}
	if a.VbsSubscriptionData != nil {
		l, err := convertVBSDataListToWire(a.VbsSubscriptionData)
		if err != nil {
			return nil, err
		}
		out.VbsSubscriptionData = l
	}
	if a.VgcsSubscriptionData != nil {
		l, err := convertVGCSDataListToWire(a.VgcsSubscriptionData)
		if err != nil {
			return nil, err
		}
		out.VgcsSubscriptionData = l
	}
	if a.VlrCamelSubscriptionInfo != nil {
		w, err := convertVlrCamelSubscriptionInfoToWire(a.VlrCamelSubscriptionInfo)
		if err != nil {
			return nil, fmt.Errorf("VlrCamelSubscriptionInfo: %w", err)
		}
		out.VlrCamelSubscriptionInfo = w
	}
	if a.NaeaPreferredCI != nil {
		out.NaeaPreferredCI = convertNaeaPreferredCIToWire(a.NaeaPreferredCI)
	}
	if a.GprsSubscriptionData != nil {
		w, err := convertGPRSSubscriptionDataToWire(a.GprsSubscriptionData)
		if err != nil {
			return nil, fmt.Errorf("GprsSubscriptionData: %w", err)
		}
		out.GprsSubscriptionData = w
	}
	if a.NetworkAccessMode != nil {
		v := gsm_map.NetworkAccessMode(*a.NetworkAccessMode)
		out.NetworkAccessMode = &v
	}
	if a.LsaInformation != nil {
		w, err := convertLSAInformationToWire(a.LsaInformation)
		if err != nil {
			return nil, fmt.Errorf("LsaInformation: %w", err)
		}
		out.LsaInformation = w
	}
	if a.LcsInformation != nil {
		w, err := convertLCSInformationToWire(a.LcsInformation)
		if err != nil {
			return nil, fmt.Errorf("LcsInformation: %w", err)
		}
		out.LcsInformation = w
	}
	if a.IstAlertTimer != nil {
		v := gsm_map.ISTAlertTimerValue(*a.IstAlertTimer)
		out.IstAlertTimer = &v
	}
	if len(a.SuperChargerSupportedInHLR) > 0 {
		if len(a.SuperChargerSupportedInHLR) < 1 || len(a.SuperChargerSupportedInHLR) > 6 {
			return nil, fmt.Errorf("SuperChargerSupportedInHLR (AgeIndicator): must be 1..6 octets (got %d)", len(a.SuperChargerSupportedInHLR))
		}
		v := gsm_map.AgeIndicator(a.SuperChargerSupportedInHLR)
		out.SuperChargerSupportedInHLR = &v
	}
	if a.McSSInfo != nil {
		w, err := convertMCSSInfoToWire(a.McSSInfo)
		if err != nil {
			return nil, fmt.Errorf("McSSInfo: %w", err)
		}
		out.McSSInfo = w
	}
	if len(a.CsAllocationRetentionPriority) > 0 {
		if len(a.CsAllocationRetentionPriority) != 1 {
			return nil, fmt.Errorf("CsAllocationRetentionPriority: must be exactly 1 octet (got %d)", len(a.CsAllocationRetentionPriority))
		}
		v := gsm_map.CSAllocationRetentionPriority(a.CsAllocationRetentionPriority)
		out.CsAllocationRetentionPriority = &v
	}
	if a.SgsnCAMELSubscriptionInfo != nil {
		w, err := convertSGSNCAMELSubscriptionInfoToWire(a.SgsnCAMELSubscriptionInfo)
		if err != nil {
			return nil, fmt.Errorf("SgsnCAMELSubscriptionInfo: %w", err)
		}
		out.SgsnCAMELSubscriptionInfo = w
	}
	if len(a.ChargingCharacteristics) > 0 {
		if len(a.ChargingCharacteristics) != 2 {
			return nil, fmt.Errorf("ChargingCharacteristics: must be exactly 2 octets (got %d)", len(a.ChargingCharacteristics))
		}
		v := gsm_map.ChargingCharacteristics(a.ChargingCharacteristics)
		out.ChargingCharacteristics = &v
	}
	if a.AccessRestrictionData != nil {
		bs := convertAccessRestrictionDataToBitString(a.AccessRestrictionData)
		out.AccessRestrictionData = &bs
	}
	if a.IcsIndicator != nil {
		v := *a.IcsIndicator
		out.IcsIndicator = &v
	}
	if a.EpsSubscriptionData != nil {
		w, err := convertEPSSubscriptionDataToWire(a.EpsSubscriptionData)
		if err != nil {
			return nil, fmt.Errorf("EpsSubscriptionData: %w", err)
		}
		out.EpsSubscriptionData = w
	}
	if a.CsgSubscriptionDataList != nil {
		l, err := convertCSGSubscriptionDataListToWire(a.CsgSubscriptionDataList)
		if err != nil {
			return nil, err
		}
		out.CsgSubscriptionDataList = l
	}
	if a.SgsnNumber != "" {
		isdn, err := encodeAddressField(a.SgsnNumber, a.SgsnNumberNature, a.SgsnNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SgsnNumber: %w", err)
		}
		v := gsm_map.ISDNAddressString(isdn)
		out.SgsnNumber = &v
	}
	if len(a.MmeName) > 0 {
		v := gsm_map.DiameterIdentity(a.MmeName)
		out.MmeName = &v
	}
	if a.SubscribedPeriodicRAUTAUtimer != nil {
		v := gsm_map.SubscribedPeriodicRAUTAUtimer(*a.SubscribedPeriodicRAUTAUtimer)
		out.SubscribedPeriodicRAUTAUtimer = &v
	}
	if a.MdtUserConsent != nil {
		v := *a.MdtUserConsent
		out.MdtUserConsent = &v
	}
	if a.SubscribedPeriodicLAUtimer != nil {
		v := gsm_map.SubscribedPeriodicLAUtimer(*a.SubscribedPeriodicLAUtimer)
		out.SubscribedPeriodicLAUtimer = &v
	}
	if a.VplmnCsgSubscriptionDataList != nil {
		l, err := convertVPLMNCSGSubscriptionDataListToWire(a.VplmnCsgSubscriptionDataList)
		if err != nil {
			return nil, err
		}
		out.VplmnCsgSubscriptionDataList = l
	}
	if a.AdditionalMSISDN != "" {
		isdn, err := encodeAddressField(a.AdditionalMSISDN, a.AdditionalMSISDNNature, a.AdditionalMSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding AdditionalMSISDN: %w", err)
		}
		v := gsm_map.ISDNAddressString(isdn)
		out.AdditionalMSISDN = &v
	}
	if a.AdjacentAccessRestrictionDataList != nil {
		l, err := convertAdjacentAccessRestrictionDataListToWire(a.AdjacentAccessRestrictionDataList)
		if err != nil {
			return nil, err
		}
		out.AdjacentAccessRestrictionDataList = l
	}
	if a.ImsiGroupIdList != nil {
		l, err := convertIMSIGroupIdListToWire(a.ImsiGroupIdList)
		if err != nil {
			return nil, err
		}
		out.ImsiGroupIdList = l
	}
	if len(a.UeUsageType) > 0 {
		v := gsm_map.UEUsageType(a.UeUsageType)
		out.UeUsageType = &v
	}
	if a.DlBufferingSuggestedPacketCount != nil {
		v := gsm_map.DLBufferingSuggestedPacketCount(*a.DlBufferingSuggestedPacketCount)
		out.DlBufferingSuggestedPacketCount = &v
	}
	if a.ResetIdList != nil {
		l, err := convertResetIdListToWire(a.ResetIdList)
		if err != nil {
			return nil, err
		}
		out.ResetIdList = l
	}
	if a.EDRXCycleLengthList != nil {
		l, err := convertEDRXCycleLengthListToWire(a.EDRXCycleLengthList)
		if err != nil {
			return nil, err
		}
		out.EDRXCycleLengthList = l
	}
	if a.ExtAccessRestrictionData != nil {
		bs := convertExtAccessRestrictionDataToBitString(a.ExtAccessRestrictionData)
		out.ExtAccessRestrictionData = &bs
	}
	return out, nil
}

func convertWireToInsertSubscriberDataArg(w *gsm_map.InsertSubscriberDataArg) (*InsertSubscriberDataArg, error) {
	if w == nil {
		return nil, fmt.Errorf("InsertSubscriberDataArg: wire pointer must not be nil")
	}
	out := &InsertSubscriberDataArg{
		RoamingRestrictionDueToUnsupportedFeature:      nullPtrToBool(w.RoamingRestrictionDueToUnsupportedFeature),
		RoamingRestrictedInSgsnDueToUnsupportedFeature: nullPtrToBool(w.RoamingRestrictedInSgsnDueToUnsupportedFeature),
		LmuIndicator:                                   nullPtrToBool(w.LmuIndicator),
		UeReachabilityRequestIndicator:                 nullPtrToBool(w.UeReachabilityRequestIndicator),
		VplmnLIPAAllowed:                               nullPtrToBool(w.VplmnLIPAAllowed),
		PsAndSMSOnlyServiceProvision:                   nullPtrToBool(w.PsAndSMSOnlyServiceProvision),
		SmsInSGSNAllowed:                               nullPtrToBool(w.SmsInSGSNAllowed),
		CsToPsSRVCCAllowedIndicator:                    nullPtrToBool(w.CsToPsSRVCCAllowedIndicator),
		PcscfRestorationRequest:                        nullPtrToBool(w.PcscfRestorationRequest),
		UserPlaneIntegrityProtectionIndicator:          nullPtrToBool(w.UserPlaneIntegrityProtectionIndicator),
		IabOperationAllowedIndicator:                   nullPtrToBool(w.IabOperationAllowedIndicator),
	}

	if w.Imsi != nil {
		out.IMSI = HexBytes(*w.Imsi)
	}
	if w.Msisdn != nil {
		s, nature, plan, err := decodeAddressField([]byte(*w.Msisdn))
		if err != nil {
			return nil, fmt.Errorf("decoding MSISDN: %w", err)
		}
		out.MSISDN = s
		out.MSISDNNature = nature
		out.MSISDNPlan = plan
	}
	if w.Category != nil {
		if len(*w.Category) != 1 {
			return nil, fmt.Errorf("InsertSubscriberDataArg.Category: must be exactly 1 octet (got %d)", len(*w.Category))
		}
		out.Category = HexBytes(*w.Category)
	}
	if w.SubscriberStatus != nil {
		v := SubscriberStatus(*w.SubscriberStatus)
		out.SubscriberStatus = &v
	}
	if w.BearerServiceList != nil {
		out.BearerServiceList = make([]HexBytes, len(w.BearerServiceList))
		for i, b := range w.BearerServiceList {
			if len(b) < 1 || len(b) > 5 {
				return nil, fmt.Errorf("BearerServiceList[%d]: must be 1..5 octets (got %d)", i, len(b))
			}
			out.BearerServiceList[i] = HexBytes(b)
		}
	}
	if w.TeleserviceList != nil {
		out.TeleserviceList = make([]HexBytes, len(w.TeleserviceList))
		for i, t := range w.TeleserviceList {
			if len(t) < 1 || len(t) > 5 {
				return nil, fmt.Errorf("TeleserviceList[%d]: must be 1..5 octets (got %d)", i, len(t))
			}
			out.TeleserviceList[i] = HexBytes(t)
		}
	}
	if w.ProvisionedSS != nil {
		l, err := convertWireToExtSSInfoList(w.ProvisionedSS)
		if err != nil {
			return nil, err
		}
		out.ProvisionedSS = l
	}
	if w.OdbData != nil {
		out.OdbData = convertWireToODBData(w.OdbData)
	}
	if w.RegionalSubscriptionData != nil {
		l, err := convertWireToZoneCodeList(w.RegionalSubscriptionData)
		if err != nil {
			return nil, err
		}
		out.RegionalSubscriptionData = l
	}
	if w.VbsSubscriptionData != nil {
		l, err := convertWireToVBSDataList(w.VbsSubscriptionData)
		if err != nil {
			return nil, err
		}
		out.VbsSubscriptionData = l
	}
	if w.VgcsSubscriptionData != nil {
		l, err := convertWireToVGCSDataList(w.VgcsSubscriptionData)
		if err != nil {
			return nil, err
		}
		out.VgcsSubscriptionData = l
	}
	if w.VlrCamelSubscriptionInfo != nil {
		v, err := convertWireToVlrCamelSubscriptionInfo(w.VlrCamelSubscriptionInfo)
		if err != nil {
			return nil, fmt.Errorf("VlrCamelSubscriptionInfo: %w", err)
		}
		out.VlrCamelSubscriptionInfo = v
	}
	if w.NaeaPreferredCI != nil {
		out.NaeaPreferredCI = convertWireToNaeaPreferredCI(w.NaeaPreferredCI)
	}
	if w.GprsSubscriptionData != nil {
		v, err := convertWireToGPRSSubscriptionData(w.GprsSubscriptionData)
		if err != nil {
			return nil, fmt.Errorf("GprsSubscriptionData: %w", err)
		}
		out.GprsSubscriptionData = v
	}
	if w.NetworkAccessMode != nil {
		v := NetworkAccessMode(*w.NetworkAccessMode)
		out.NetworkAccessMode = &v
	}
	if w.LsaInformation != nil {
		v, err := convertWireToLSAInformation(w.LsaInformation)
		if err != nil {
			return nil, fmt.Errorf("LsaInformation: %w", err)
		}
		out.LsaInformation = v
	}
	if w.LcsInformation != nil {
		v, err := convertWireToLCSInformation(w.LcsInformation)
		if err != nil {
			return nil, fmt.Errorf("LcsInformation: %w", err)
		}
		out.LcsInformation = v
	}
	if w.IstAlertTimer != nil {
		v := int64(*w.IstAlertTimer)
		out.IstAlertTimer = &v
	}
	if w.SuperChargerSupportedInHLR != nil {
		if len(*w.SuperChargerSupportedInHLR) < 1 || len(*w.SuperChargerSupportedInHLR) > 6 {
			return nil, fmt.Errorf("SuperChargerSupportedInHLR (AgeIndicator): must be 1..6 octets (got %d)", len(*w.SuperChargerSupportedInHLR))
		}
		out.SuperChargerSupportedInHLR = HexBytes(*w.SuperChargerSupportedInHLR)
	}
	if w.McSSInfo != nil {
		v, err := convertWireToMCSSInfo(w.McSSInfo)
		if err != nil {
			return nil, fmt.Errorf("McSSInfo: %w", err)
		}
		out.McSSInfo = v
	}
	if w.CsAllocationRetentionPriority != nil {
		if len(*w.CsAllocationRetentionPriority) != 1 {
			return nil, fmt.Errorf("CsAllocationRetentionPriority: must be exactly 1 octet (got %d)", len(*w.CsAllocationRetentionPriority))
		}
		out.CsAllocationRetentionPriority = HexBytes(*w.CsAllocationRetentionPriority)
	}
	if w.SgsnCAMELSubscriptionInfo != nil {
		v, err := convertWireToSGSNCAMELSubscriptionInfo(w.SgsnCAMELSubscriptionInfo)
		if err != nil {
			return nil, fmt.Errorf("SgsnCAMELSubscriptionInfo: %w", err)
		}
		out.SgsnCAMELSubscriptionInfo = v
	}
	if w.ChargingCharacteristics != nil {
		if len(*w.ChargingCharacteristics) != 2 {
			return nil, fmt.Errorf("ChargingCharacteristics: must be exactly 2 octets (got %d)", len(*w.ChargingCharacteristics))
		}
		out.ChargingCharacteristics = HexBytes(*w.ChargingCharacteristics)
	}
	if w.AccessRestrictionData != nil {
		out.AccessRestrictionData = convertBitStringToAccessRestrictionData(*w.AccessRestrictionData)
	}
	if w.IcsIndicator != nil {
		v := *w.IcsIndicator
		out.IcsIndicator = &v
	}
	if w.EpsSubscriptionData != nil {
		v, err := convertWireToEPSSubscriptionData(w.EpsSubscriptionData)
		if err != nil {
			return nil, fmt.Errorf("EpsSubscriptionData: %w", err)
		}
		out.EpsSubscriptionData = v
	}
	if w.CsgSubscriptionDataList != nil {
		l, err := convertWireToCSGSubscriptionDataList(w.CsgSubscriptionDataList)
		if err != nil {
			return nil, err
		}
		out.CsgSubscriptionDataList = l
	}
	if w.SgsnNumber != nil {
		s, nature, plan, err := decodeAddressField([]byte(*w.SgsnNumber))
		if err != nil {
			return nil, fmt.Errorf("decoding SgsnNumber: %w", err)
		}
		out.SgsnNumber = s
		out.SgsnNumberNature = nature
		out.SgsnNumberPlan = plan
	}
	if w.MmeName != nil {
		out.MmeName = HexBytes(*w.MmeName)
	}
	if w.SubscribedPeriodicRAUTAUtimer != nil {
		v := int64(*w.SubscribedPeriodicRAUTAUtimer)
		out.SubscribedPeriodicRAUTAUtimer = &v
	}
	if w.MdtUserConsent != nil {
		v := *w.MdtUserConsent
		out.MdtUserConsent = &v
	}
	if w.SubscribedPeriodicLAUtimer != nil {
		v := int64(*w.SubscribedPeriodicLAUtimer)
		out.SubscribedPeriodicLAUtimer = &v
	}
	if w.VplmnCsgSubscriptionDataList != nil {
		l, err := convertWireToVPLMNCSGSubscriptionDataList(w.VplmnCsgSubscriptionDataList)
		if err != nil {
			return nil, err
		}
		out.VplmnCsgSubscriptionDataList = l
	}
	if w.AdditionalMSISDN != nil {
		s, nature, plan, err := decodeAddressField([]byte(*w.AdditionalMSISDN))
		if err != nil {
			return nil, fmt.Errorf("decoding AdditionalMSISDN: %w", err)
		}
		out.AdditionalMSISDN = s
		out.AdditionalMSISDNNature = nature
		out.AdditionalMSISDNPlan = plan
	}
	if w.AdjacentAccessRestrictionDataList != nil {
		l, err := convertWireToAdjacentAccessRestrictionDataList(w.AdjacentAccessRestrictionDataList)
		if err != nil {
			return nil, err
		}
		out.AdjacentAccessRestrictionDataList = l
	}
	if w.ImsiGroupIdList != nil {
		l, err := convertWireToIMSIGroupIdList(w.ImsiGroupIdList)
		if err != nil {
			return nil, err
		}
		out.ImsiGroupIdList = l
	}
	if w.UeUsageType != nil {
		out.UeUsageType = HexBytes(*w.UeUsageType)
	}
	if w.DlBufferingSuggestedPacketCount != nil {
		v := int64(*w.DlBufferingSuggestedPacketCount)
		out.DlBufferingSuggestedPacketCount = &v
	}
	if w.ResetIdList != nil {
		l, err := convertWireToResetIdList(w.ResetIdList)
		if err != nil {
			return nil, err
		}
		out.ResetIdList = l
	}
	if w.EDRXCycleLengthList != nil {
		l, err := convertWireToEDRXCycleLengthList(w.EDRXCycleLengthList)
		if err != nil {
			return nil, err
		}
		out.EDRXCycleLengthList = l
	}
	if w.ExtAccessRestrictionData != nil {
		out.ExtAccessRestrictionData = convertBitStringToExtAccessRestrictionData(*w.ExtAccessRestrictionData)
	}
	return out, nil
}

// ============================================================================
// InsertSubscriberDataRes ↔ wire converter
// ============================================================================

func convertInsertSubscriberDataResToWire(r *InsertSubscriberDataRes) (*gsm_map.InsertSubscriberDataRes, error) {
	if r == nil {
		return nil, fmt.Errorf("InsertSubscriberDataRes: argument must not be nil")
	}
	out := &gsm_map.InsertSubscriberDataRes{}
	if r.TeleserviceList != nil {
		out.TeleserviceList = make(gsm_map.TeleserviceList, len(r.TeleserviceList))
		for i, t := range r.TeleserviceList {
			if len(t) < 1 || len(t) > 5 {
				return nil, fmt.Errorf("Res.TeleserviceList[%d]: must be 1..5 octets (got %d)", i, len(t))
			}
			out.TeleserviceList[i] = gsm_map.ExtTeleserviceCode(t)
		}
	}
	if r.BearerServiceList != nil {
		out.BearerServiceList = make(gsm_map.BearerServiceList, len(r.BearerServiceList))
		for i, b := range r.BearerServiceList {
			if len(b) < 1 || len(b) > 5 {
				return nil, fmt.Errorf("Res.BearerServiceList[%d]: must be 1..5 octets (got %d)", i, len(b))
			}
			out.BearerServiceList[i] = gsm_map.ExtBearerServiceCode(b)
		}
	}
	if r.SsList != nil {
		out.SsList = make(gsm_map.SSList, len(r.SsList))
		for i, c := range r.SsList {
			out.SsList[i] = gsm_map.SSCode{byte(c)}
		}
	}
	if r.OdbGeneralData != nil {
		bs := convertODBGeneralDataToBitString(r.OdbGeneralData)
		out.OdbGeneralData = &bs
	}
	if r.RegionalSubscriptionResponse != nil {
		v := gsm_map.RegionalSubscriptionResponse(*r.RegionalSubscriptionResponse)
		out.RegionalSubscriptionResponse = &v
	}
	if r.SupportedCamelPhases != nil {
		bs := convertCamelPhasesToBitString(r.SupportedCamelPhases)
		out.SupportedCamelPhases = &bs
	}
	if r.OfferedCamel4CSIs != nil {
		bs := convertOfferedCamel4CSIsToBitString(r.OfferedCamel4CSIs)
		out.OfferedCamel4CSIs = &bs
	}
	if r.SupportedFeatures != nil {
		bs := convertSupportedFeaturesToBitString(r.SupportedFeatures)
		out.SupportedFeatures = &bs
	}
	if r.ExtSupportedFeatures != nil {
		bs := convertExtSupportedFeaturesToBitString(r.ExtSupportedFeatures)
		out.ExtSupportedFeatures = &bs
	}
	return out, nil
}

func convertWireToInsertSubscriberDataRes(w *gsm_map.InsertSubscriberDataRes) (*InsertSubscriberDataRes, error) {
	if w == nil {
		return nil, fmt.Errorf("InsertSubscriberDataRes: wire pointer must not be nil")
	}
	out := &InsertSubscriberDataRes{}
	if w.TeleserviceList != nil {
		out.TeleserviceList = make([]HexBytes, len(w.TeleserviceList))
		for i, t := range w.TeleserviceList {
			if len(t) < 1 || len(t) > 5 {
				return nil, fmt.Errorf("Res.TeleserviceList[%d]: must be 1..5 octets (got %d)", i, len(t))
			}
			out.TeleserviceList[i] = HexBytes(t)
		}
	}
	if w.BearerServiceList != nil {
		out.BearerServiceList = make([]HexBytes, len(w.BearerServiceList))
		for i, b := range w.BearerServiceList {
			if len(b) < 1 || len(b) > 5 {
				return nil, fmt.Errorf("Res.BearerServiceList[%d]: must be 1..5 octets (got %d)", i, len(b))
			}
			out.BearerServiceList[i] = HexBytes(b)
		}
	}
	if w.SsList != nil {
		out.SsList = make([]SsCode, len(w.SsList))
		for i, c := range w.SsList {
			if len(c) == 0 {
				return nil, fmt.Errorf("Res.SsList[%d]: SS-Code must be present (1 octet)", i)
			}
			out.SsList[i] = SsCode(c[0])
		}
	}
	if w.OdbGeneralData != nil {
		out.OdbGeneralData = convertBitStringToODBGeneralData(*w.OdbGeneralData)
	}
	if w.RegionalSubscriptionResponse != nil {
		v := RegionalSubscriptionResponse(*w.RegionalSubscriptionResponse)
		out.RegionalSubscriptionResponse = &v
	}
	if w.SupportedCamelPhases != nil {
		out.SupportedCamelPhases = convertBitStringToCamelPhases(*w.SupportedCamelPhases)
	}
	if w.OfferedCamel4CSIs != nil {
		out.OfferedCamel4CSIs = convertBitStringToOfferedCamel4CSIs(*w.OfferedCamel4CSIs)
	}
	if w.SupportedFeatures != nil {
		out.SupportedFeatures = convertBitStringToSupportedFeatures(*w.SupportedFeatures)
	}
	if w.ExtSupportedFeatures != nil {
		out.ExtSupportedFeatures = convertBitStringToExtSupportedFeatures(*w.ExtSupportedFeatures)
	}
	return out, nil
}

// ============================================================================
// Public Parse + Marshal entry points (opCode 7)
// ============================================================================

// ParseInsertSubscriberData decodes BER-encoded bytes into an
// InsertSubscriberDataArg.
func ParseInsertSubscriberData(data []byte) (*InsertSubscriberDataArg, error) {
	var arg gsm_map.InsertSubscriberDataArg
	if err := arg.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding InsertSubscriberDataArg: %w", err)
	}
	return convertWireToInsertSubscriberDataArg(&arg)
}

// ParseInsertSubscriberDataRes decodes BER-encoded bytes into an
// InsertSubscriberDataRes.
func ParseInsertSubscriberDataRes(data []byte) (*InsertSubscriberDataRes, error) {
	var res gsm_map.InsertSubscriberDataRes
	if err := res.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding InsertSubscriberDataRes: %w", err)
	}
	return convertWireToInsertSubscriberDataRes(&res)
}

// Marshal encodes InsertSubscriberDataArg into BER-encoded bytes.
func (a *InsertSubscriberDataArg) Marshal() ([]byte, error) {
	arg, err := convertInsertSubscriberDataArgToWire(a)
	if err != nil {
		return nil, fmt.Errorf("converting InsertSubscriberDataArg: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding InsertSubscriberDataArg: %w", err)
	}
	return data, nil
}

// Marshal encodes InsertSubscriberDataRes into BER-encoded bytes.
func (r *InsertSubscriberDataRes) Marshal() ([]byte, error) {
	res, err := convertInsertSubscriberDataResToWire(r)
	if err != nil {
		return nil, fmt.Errorf("converting InsertSubscriberDataRes: %w", err)
	}
	data, err := res.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding InsertSubscriberDataRes: %w", err)
	}
	return data, nil
}
