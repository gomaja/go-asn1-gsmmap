package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime"
)

func convertDeleteSubscriberDataArgToWire(v *DeleteSubscriberDataArg) (*gsm_map.DeleteSubscriberDataArg, error) {
	if v == nil {
		return nil, ErrDeleteSubscriberDataArgNil
	}
	imsi, err := encodeIdentityDigits(identityIMSI, v.IMSI)
	if err != nil {
		return nil, fmt.Errorf("IMSI: %w", err)
	}
	w := &gsm_map.DeleteSubscriberDataArg{Imsi: imsi,
		RoamingRestrictionDueToUnsupportedFeature:       boolToNullPtr(v.RoamingRestrictionDueToUnsupportedFeature),
		VbsGroupIndication:                              boolToNullPtr(v.VbsGroupIndication),
		VgcsGroupIndication:                             boolToNullPtr(v.VgcsGroupIndication),
		CamelSubscriptionInfoWithdraw:                   boolToNullPtr(v.CamelSubscriptionInfoWithdraw),
		RoamingRestrictedInSgsnDueToUnsuppportedFeature: boolToNullPtr(v.RoamingRestrictedInSgsnDueToUnsupportedFeature),
		GmlcListWithdraw:                                boolToNullPtr(v.GmlcListWithdraw),
		IstInformationWithdraw:                          boolToNullPtr(v.IstInformationWithdraw),
		ChargingCharacteristicsWithdraw:                 boolToNullPtr(v.ChargingCharacteristicsWithdraw),
		StnSrWithdraw:                                   boolToNullPtr(v.StnSrWithdraw),
		ApnOiReplacementWithdraw:                        boolToNullPtr(v.ApnOiReplacementWithdraw),
		CsgSubscriptionDeleted:                          boolToNullPtr(v.CsgSubscriptionDeleted),
		SubscribedPeriodicTAURAUTimerWithdraw:           boolToNullPtr(v.SubscribedPeriodicTAURAUTimerWithdraw),
		SubscribedPeriodicLAUTimerWithdraw:              boolToNullPtr(v.SubscribedPeriodicLAUTimerWithdraw),
		SubscribedVsrvccWithdraw:                        boolToNullPtr(v.SubscribedVsrvccWithdraw),
		VplmnCsgSubscriptionDeleted:                     boolToNullPtr(v.VplmnCsgSubscriptionDeleted),
		AdditionalMSISDNWithdraw:                        boolToNullPtr(v.AdditionalMSISDNWithdraw),
		CsToPsSRVCCWithdraw:                             boolToNullPtr(v.CsToPsSRVCCWithdraw),
		ImsiGroupIdListWithdraw:                         boolToNullPtr(v.ImsiGroupIdListWithdraw),
		UserPlaneIntegrityProtectionWithdraw:            boolToNullPtr(v.UserPlaneIntegrityProtectionWithdraw),
		DlBufferingSuggestedPacketCountWithdraw:         boolToNullPtr(v.DlBufferingSuggestedPacketCountWithdraw),
		UeUsageTypeWithdraw:                             boolToNullPtr(v.UeUsageTypeWithdraw),
		ResetIdsWithdraw:                                boolToNullPtr(v.ResetIdsWithdraw),
		IabOperationWithdraw:                            boolToNullPtr(v.IabOperationWithdraw),
	}
	if v.BasicServiceList != nil {
		w.BasicServiceList = &gsm_map.BasicServiceList{Values: make([]gsm_map.ExtBasicServiceCode, len(v.BasicServiceList))}
		for i := range v.BasicServiceList {
			b, err := convertExtBasicServiceCodeToWire(&v.BasicServiceList[i], senderDeleteSubscriberData)
			if err != nil {
				return nil, fmt.Errorf("BasicServiceList[%d]: %w", i, err)
			}
			w.BasicServiceList.Values[i] = *b
		}
	}
	if v.SsList != nil {
		w.SsList = &gsm_map.SSList{Values: make([]gsm_map.SSCode, len(v.SsList))}
		for i, code := range v.SsList {
			if err := validateSSCodeToWire(code, senderDeleteSubscriberData); err != nil {
				return nil, fmt.Errorf("SsList[%d]: %w", i, err)
			}
			w.SsList.Values[i] = []byte{byte(code)}
		}
	}
	if len(v.RegionalSubscriptionIdentifier) > 0 {
		b := gsm_map.ZoneCode(v.RegionalSubscriptionIdentifier)
		w.RegionalSubscriptionIdentifier = &b
	}
	if v.GprsSubscriptionDataWithdraw != nil {
		w.GprsSubscriptionDataWithdraw, err = convertGPRSSubscriptionDataWithdrawToWire(v.GprsSubscriptionDataWithdraw)
		if err != nil {
			return nil, fmt.Errorf("GprsSubscriptionDataWithdraw: %w", err)
		}
	}
	if v.EpsSubscriptionDataWithdraw != nil {
		w.EpsSubscriptionDataWithdraw, err = convertEPSSubscriptionDataWithdrawToWire(v.EpsSubscriptionDataWithdraw)
		if err != nil {
			return nil, fmt.Errorf("EpsSubscriptionDataWithdraw: %w", err)
		}
	}
	if v.LsaInformationWithdraw != nil {
		w.LsaInformationWithdraw, err = convertLSAInformationWithdrawToWire(v.LsaInformationWithdraw)
		if err != nil {
			return nil, fmt.Errorf("LsaInformationWithdraw: %w", err)
		}
	}
	if v.SpecificCSIWithdraw != nil {
		w.SpecificCSIWithdraw, err = convertSpecificCSIWithdrawToWire(v.SpecificCSIWithdraw)
		if err != nil {
			return nil, fmt.Errorf("SpecificCSIWithdraw: %w", err)
		}
	}
	return w, nil
}

func convertWireToDeleteSubscriberDataArg(w *gsm_map.DeleteSubscriberDataArg) (*DeleteSubscriberDataArg, error) {
	imsi, err := decodeIdentityDigits(identityIMSI, w.Imsi)
	if err != nil {
		return nil, fmt.Errorf("IMSI: %w", err)
	}
	v := &DeleteSubscriberDataArg{IMSI: imsi,
		RoamingRestrictionDueToUnsupportedFeature:      nullPtrToBool(w.RoamingRestrictionDueToUnsupportedFeature),
		VbsGroupIndication:                             nullPtrToBool(w.VbsGroupIndication),
		VgcsGroupIndication:                            nullPtrToBool(w.VgcsGroupIndication),
		CamelSubscriptionInfoWithdraw:                  nullPtrToBool(w.CamelSubscriptionInfoWithdraw),
		RoamingRestrictedInSgsnDueToUnsupportedFeature: nullPtrToBool(w.RoamingRestrictedInSgsnDueToUnsuppportedFeature),
		GmlcListWithdraw:                               nullPtrToBool(w.GmlcListWithdraw),
		IstInformationWithdraw:                         nullPtrToBool(w.IstInformationWithdraw),
		ChargingCharacteristicsWithdraw:                nullPtrToBool(w.ChargingCharacteristicsWithdraw),
		StnSrWithdraw:                                  nullPtrToBool(w.StnSrWithdraw),
		ApnOiReplacementWithdraw:                       nullPtrToBool(w.ApnOiReplacementWithdraw),
		CsgSubscriptionDeleted:                         nullPtrToBool(w.CsgSubscriptionDeleted),
		SubscribedPeriodicTAURAUTimerWithdraw:          nullPtrToBool(w.SubscribedPeriodicTAURAUTimerWithdraw),
		SubscribedPeriodicLAUTimerWithdraw:             nullPtrToBool(w.SubscribedPeriodicLAUTimerWithdraw),
		SubscribedVsrvccWithdraw:                       nullPtrToBool(w.SubscribedVsrvccWithdraw),
		VplmnCsgSubscriptionDeleted:                    nullPtrToBool(w.VplmnCsgSubscriptionDeleted),
		AdditionalMSISDNWithdraw:                       nullPtrToBool(w.AdditionalMSISDNWithdraw),
		CsToPsSRVCCWithdraw:                            nullPtrToBool(w.CsToPsSRVCCWithdraw),
		ImsiGroupIdListWithdraw:                        nullPtrToBool(w.ImsiGroupIdListWithdraw),
		UserPlaneIntegrityProtectionWithdraw:           nullPtrToBool(w.UserPlaneIntegrityProtectionWithdraw),
		DlBufferingSuggestedPacketCountWithdraw:        nullPtrToBool(w.DlBufferingSuggestedPacketCountWithdraw),
		UeUsageTypeWithdraw:                            nullPtrToBool(w.UeUsageTypeWithdraw),
		ResetIdsWithdraw:                               nullPtrToBool(w.ResetIdsWithdraw),
		IabOperationWithdraw:                           nullPtrToBool(w.IabOperationWithdraw),
	}
	if w.BasicServiceList != nil {
		v.BasicServiceList = make([]ExtBasicServiceCode, len(w.BasicServiceList.Values))
		for i := range w.BasicServiceList.Values {
			b, err := convertWireToExtBasicServiceCode(&w.BasicServiceList.Values[i])
			if err != nil {
				return nil, fmt.Errorf("BasicServiceList[%d]: %w", i, err)
			}
			v.BasicServiceList[i] = *b
		}
	}
	if w.SsList != nil {
		v.SsList = make([]SsCode, len(w.SsList.Values))
		for i, code := range w.SsList.Values {
			v.SsList[i] = SsCode(code[0])
		}
	}
	if w.RegionalSubscriptionIdentifier != nil {
		v.RegionalSubscriptionIdentifier = HexBytes(*w.RegionalSubscriptionIdentifier)
	}
	if w.GprsSubscriptionDataWithdraw != nil {
		v.GprsSubscriptionDataWithdraw, err = convertWireToGPRSSubscriptionDataWithdraw(w.GprsSubscriptionDataWithdraw)
		if err != nil {
			return nil, fmt.Errorf("GprsSubscriptionDataWithdraw: %w", err)
		}
	}
	if w.EpsSubscriptionDataWithdraw != nil {
		v.EpsSubscriptionDataWithdraw, err = convertWireToEPSSubscriptionDataWithdraw(w.EpsSubscriptionDataWithdraw)
		if err != nil {
			return nil, fmt.Errorf("EpsSubscriptionDataWithdraw: %w", err)
		}
	}
	if w.LsaInformationWithdraw != nil {
		v.LsaInformationWithdraw, err = convertWireToLSAInformationWithdraw(w.LsaInformationWithdraw)
		if err != nil {
			return nil, fmt.Errorf("LsaInformationWithdraw: %w", err)
		}
	}
	if w.SpecificCSIWithdraw != nil {
		v.SpecificCSIWithdraw = convertWireToSpecificCSIWithdraw(*w.SpecificCSIWithdraw)
	}
	return v, nil
}

func convertDeleteSubscriberDataResToWire(v *DeleteSubscriberDataRes) (*gsm_map.DeleteSubscriberDataRes, error) {
	if v == nil {
		return nil, ErrDeleteSubscriberDataResNil
	}
	// 3GPP TS 29.002 V19.1.0 §8.8.2.3 excludes these two ASN.1 values.
	// The codec checks non-extensible ENUMERATED membership.
	if v.RegionalSubscriptionResponse != nil && (*v.RegionalSubscriptionResponse == RegionalSubscriptionResponseTooManyZoneCodes || *v.RegionalSubscriptionResponse == RegionalSubscriptionResponseZoneCodesConflict) {
		return nil, ErrDeleteSubscriberDataResRegionalSubscriptionResponseInvalid
	}
	return &gsm_map.DeleteSubscriberDataRes{RegionalSubscriptionResponse: v.RegionalSubscriptionResponse}, nil
}

func convertWireToDeleteSubscriberDataRes(w *gsm_map.DeleteSubscriberDataRes) (*DeleteSubscriberDataRes, error) {
	return &DeleteSubscriberDataRes{RegionalSubscriptionResponse: w.RegionalSubscriptionResponse}, nil
}

func convertGPRSSubscriptionDataWithdrawToWire(v *GPRSSubscriptionDataWithdraw) (*gsm_map.GPRSSubscriptionDataWithdraw, error) {
	var w gsm_map.GPRSSubscriptionDataWithdraw
	switch {
	case v.AllGPRSData && v.ContextIDList != nil:
		return nil, ErrGPRSSubscriptionDataWithdrawMultipleAlternatives
	case v.AllGPRSData:
		w = gsm_map.NewGPRSSubscriptionDataWithdrawAllGPRSData(struct{}{})
	case v.ContextIDList != nil:
		list := &gsm_map.ContextIdList{Values: make([]gsm_map.ContextId, len(v.ContextIDList))}
		for i, id := range v.ContextIDList {
			list.Values[i] = gsm_map.ContextId(id)
		}
		w = gsm_map.NewGPRSSubscriptionDataWithdrawContextIdList(list)
	default:
		return nil, ErrGPRSSubscriptionDataWithdrawNoAlternative
	}
	return &w, nil
}

func convertWireToGPRSSubscriptionDataWithdraw(w *gsm_map.GPRSSubscriptionDataWithdraw) (*GPRSSubscriptionDataWithdraw, error) {
	switch w.Choice {
	case gsm_map.GPRSSubscriptionDataWithdrawChoiceAllGPRSData:
		return &GPRSSubscriptionDataWithdraw{AllGPRSData: true}, nil
	case gsm_map.GPRSSubscriptionDataWithdrawChoiceContextIdList:
		list := make([]int, len(w.ContextIdList.Values))
		for i, id := range w.ContextIdList.Values {
			list[i] = int(id)
		}
		return &GPRSSubscriptionDataWithdraw{ContextIDList: list}, nil
	default:
		return nil, ErrGPRSSubscriptionDataWithdrawUnknownAlternative
	}
}

func convertEPSSubscriptionDataWithdrawToWire(v *EPSSubscriptionDataWithdraw) (*gsm_map.EPSSubscriptionDataWithdraw, error) {
	var w gsm_map.EPSSubscriptionDataWithdraw
	switch {
	case v.AllEPSData && v.ContextIDList != nil:
		return nil, ErrEPSSubscriptionDataWithdrawMultipleAlternatives
	case v.AllEPSData:
		w = gsm_map.NewEPSSubscriptionDataWithdrawAllEPSData(struct{}{})
	case v.ContextIDList != nil:
		list := &gsm_map.ContextIdList{Values: make([]gsm_map.ContextId, len(v.ContextIDList))}
		for i, id := range v.ContextIDList {
			list.Values[i] = gsm_map.ContextId(id)
		}
		w = gsm_map.NewEPSSubscriptionDataWithdrawContextIdList(list)
	default:
		return nil, ErrEPSSubscriptionDataWithdrawNoAlternative
	}
	return &w, nil
}

func convertWireToEPSSubscriptionDataWithdraw(w *gsm_map.EPSSubscriptionDataWithdraw) (*EPSSubscriptionDataWithdraw, error) {
	switch w.Choice {
	case gsm_map.EPSSubscriptionDataWithdrawChoiceAllEPSData:
		return &EPSSubscriptionDataWithdraw{AllEPSData: true}, nil
	case gsm_map.EPSSubscriptionDataWithdrawChoiceContextIdList:
		list := make([]int, len(w.ContextIdList.Values))
		for i, id := range w.ContextIdList.Values {
			list[i] = int(id)
		}
		return &EPSSubscriptionDataWithdraw{ContextIDList: list}, nil
	default:
		return nil, ErrEPSSubscriptionDataWithdrawUnknownAlternative
	}
}

func convertLSAInformationWithdrawToWire(v *LSAInformationWithdraw) (*gsm_map.LSAInformationWithdraw, error) {
	var w gsm_map.LSAInformationWithdraw
	switch {
	case v.AllLSAData && v.LSAIdentityList != nil:
		return nil, ErrLSAInformationWithdrawMultipleAlternatives
	case v.AllLSAData:
		w = gsm_map.NewLSAInformationWithdrawAllLSAData(struct{}{})
	case v.LSAIdentityList != nil:
		list := &gsm_map.LSAIdentityList{Values: make([]gsm_map.LSAIdentity, len(v.LSAIdentityList))}
		for i, id := range v.LSAIdentityList {
			list.Values[i] = gsm_map.LSAIdentity(id)
		}
		w = gsm_map.NewLSAInformationWithdrawLsaIdentityList(list)
	default:
		return nil, ErrLSAInformationWithdrawNoAlternative
	}
	return &w, nil
}

func convertWireToLSAInformationWithdraw(w *gsm_map.LSAInformationWithdraw) (*LSAInformationWithdraw, error) {
	switch w.Choice {
	case gsm_map.LSAInformationWithdrawChoiceAllLSAData:
		return &LSAInformationWithdraw{AllLSAData: true}, nil
	case gsm_map.LSAInformationWithdrawChoiceLsaIdentityList:
		list := make([]HexBytes, len(w.LsaIdentityList.Values))
		for i, id := range w.LsaIdentityList.Values {
			list[i] = HexBytes(id)
		}
		return &LSAInformationWithdraw{LSAIdentityList: list}, nil
	default:
		return nil, ErrLSAInformationWithdrawUnknownAlternative
	}
}

func convertSpecificCSIWithdrawToWire(v *SpecificCSIWithdraw) (*gsm_map.SpecificCSIWithdraw, error) {
	// 3GPP TS 29.002 V19.1.0 §17.7.1: "Bit 8 and bits 11-13 are only applicable
	// for the NoteSubscriberDataModified operation".
	if v.TCSI {
		return nil, ErrSpecificCSIWithdrawTCSINotApplicable
	}
	bits := []bool{
		gsm_map.SpecificCSIWithdrawOCsi:     v.OCSI,
		gsm_map.SpecificCSIWithdrawSsCsi:    v.SsCSI,
		gsm_map.SpecificCSIWithdrawTifCsi:   v.TifCSI,
		gsm_map.SpecificCSIWithdrawDCsi:     v.DCSI,
		gsm_map.SpecificCSIWithdrawVtCsi:    v.VtCSI,
		gsm_map.SpecificCSIWithdrawMoSmsCsi: v.MoSmsCSI,
		gsm_map.SpecificCSIWithdrawMCsi:     v.MCSI,
		gsm_map.SpecificCSIWithdrawGprsCsi:  v.GprsCSI,
		gsm_map.SpecificCSIWithdrawMtSmsCsi: v.MtSmsCSI,
		gsm_map.SpecificCSIWithdrawMgCsi:    v.MgCSI,
	}
	positions := namedBitPositions(bits)
	// 3GPP TS 29.002 V19.1.0 §8.8.2.3: one or more specific elements shall be withdrawn.
	if len(positions) == 0 {
		return nil, ErrSpecificCSIWithdrawEmpty
	}
	w, err := gsm_map.NewSpecificCSIWithdraw(positions...)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func convertWireToSpecificCSIWithdraw(w runtime.BitString) *SpecificCSIWithdraw {
	// 3GPP TS 29.002 V19.1.0 §17.7.1: "bits 11 to 31 shall be ignored if
	// received by a non-IP Multimedia Core Network entity".
	return &SpecificCSIWithdraw{
		OCSI:     w.Has(gsm_map.SpecificCSIWithdrawOCsi),
		SsCSI:    w.Has(gsm_map.SpecificCSIWithdrawSsCsi),
		TifCSI:   w.Has(gsm_map.SpecificCSIWithdrawTifCsi),
		DCSI:     w.Has(gsm_map.SpecificCSIWithdrawDCsi),
		VtCSI:    w.Has(gsm_map.SpecificCSIWithdrawVtCsi),
		MoSmsCSI: w.Has(gsm_map.SpecificCSIWithdrawMoSmsCsi),
		MCSI:     w.Has(gsm_map.SpecificCSIWithdrawMCsi),
		GprsCSI:  w.Has(gsm_map.SpecificCSIWithdrawGprsCsi),
		TCSI:     w.Has(gsm_map.SpecificCSIWithdrawTCsi),
		MtSmsCSI: w.Has(gsm_map.SpecificCSIWithdrawMtSmsCsi),
		MgCSI:    w.Has(gsm_map.SpecificCSIWithdrawMgCsi),
	}
}
