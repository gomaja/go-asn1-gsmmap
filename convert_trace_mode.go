package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func convertActivateTraceModeArgToWire(a *ActivateTraceModeArg) (*gsm_map.ActivateTraceModeArg, error) {
	if a == nil {
		return nil, ErrActivateTraceModeArgNil
	}
	w := &gsm_map.ActivateTraceModeArg{TraceReference: a.TraceReference, TraceType: int64(a.TraceType)}
	if a.IMSI != "" {
		b, err := encodeIdentityDigits(identityIMSI, a.IMSI)
		if err != nil {
			return nil, fmt.Errorf("encoding IMSI: %w", err)
		}
		w.Imsi = &b
	}
	if a.OMCID != "" {
		b, err := encodeAddressField(a.OMCID, a.OMCIDNature, a.OMCIDPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding OMCID: %w", err)
		}
		w.OmcId = &b
	}
	if len(a.TraceReference2) > 0 {
		b := []byte(a.TraceReference2)
		w.TraceReference2 = &b
	}
	var err error
	if a.TraceDepthList != nil {
		w.TraceDepthList, err = convertTraceDepthListToWire(a.TraceDepthList)
		if err != nil {
			return nil, fmt.Errorf("TraceDepthList: %w", err)
		}
	}
	if a.TraceNETypeList != nil {
		b, e := convertTraceNETypeListToBitString(a.TraceNETypeList)
		if e != nil {
			return nil, e
		}
		w.TraceNETypeList = &b
	}
	if a.TraceInterfaceList != nil {
		w.TraceInterfaceList, err = convertTraceInterfaceListToWire(a.TraceInterfaceList)
		if err != nil {
			return nil, fmt.Errorf("TraceInterfaceList: %w", err)
		}
	}
	if a.TraceEventList != nil {
		w.TraceEventList, err = convertTraceEventListToWire(a.TraceEventList)
		if err != nil {
			return nil, fmt.Errorf("TraceEventList: %w", err)
		}
	}
	if err := encodeOptionalGSN(a.TraceCollectionEntity, "TraceCollectionEntity", &w.TraceCollectionEntity); err != nil {
		return nil, err
	}
	if a.MDTConfiguration != nil {
		w.MdtConfiguration, err = convertMDTConfigurationToWire(a.MDTConfiguration)
		if err != nil {
			return nil, fmt.Errorf("MDTConfiguration: %w", err)
		}
	}
	return w, nil
}

func convertWireToActivateTraceModeArg(w *gsm_map.ActivateTraceModeArg) (*ActivateTraceModeArg, error) {
	a := &ActivateTraceModeArg{TraceReference: HexBytes(w.TraceReference), TraceType: int(w.TraceType)}
	if w.Imsi != nil {
		s, err := decodeIdentityDigits(identityIMSI, *w.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding IMSI: %w", err)
		}
		a.IMSI = s
	}
	if w.OmcId != nil {
		s, n, p, err := decodeAddressWithDigits(*w.OmcId, ErrActivateTraceModeArgOMCIDDecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding OMCID: %w", err)
		}
		a.OMCID = s
		a.OMCIDNature = n
		a.OMCIDPlan = p
	}
	if w.TraceReference2 != nil {
		a.TraceReference2 = HexBytes(*w.TraceReference2)
	}
	if w.TraceDepthList != nil {
		a.TraceDepthList = convertWireToTraceDepthList(w.TraceDepthList)
	}
	if w.TraceNETypeList != nil {
		a.TraceNETypeList = convertBitStringToTraceNETypeList(*w.TraceNETypeList)
	}
	if w.TraceInterfaceList != nil {
		a.TraceInterfaceList = convertWireToTraceInterfaceList(w.TraceInterfaceList)
	}
	if w.TraceEventList != nil {
		a.TraceEventList = convertWireToTraceEventList(w.TraceEventList)
	}
	s, err := decodeOptionalGSN(w.TraceCollectionEntity, "TraceCollectionEntity")
	if err != nil {
		return nil, err
	}
	a.TraceCollectionEntity = s
	if w.MdtConfiguration != nil {
		a.MDTConfiguration = convertWireToMDTConfiguration(w.MdtConfiguration)
	}
	return a, nil
}

func convertActivateTraceModeResToWire(r *ActivateTraceModeRes) (*gsm_map.ActivateTraceModeRes, error) {
	if r == nil {
		return nil, ErrActivateTraceModeResNil
	}
	return &gsm_map.ActivateTraceModeRes{TraceSupportIndicator: boolToNullPtr(r.TraceSupportIndicator)}, nil
}

func convertWireToActivateTraceModeRes(w *gsm_map.ActivateTraceModeRes) *ActivateTraceModeRes {
	return &ActivateTraceModeRes{TraceSupportIndicator: nullPtrToBool(w.TraceSupportIndicator)}
}

func convertDeactivateTraceModeArgToWire(a *DeactivateTraceModeArg) (*gsm_map.DeactivateTraceModeArg, error) {
	if a == nil {
		return nil, ErrDeactivateTraceModeArgNil
	}
	w := &gsm_map.DeactivateTraceModeArg{TraceReference: a.TraceReference}
	if a.IMSI != "" {
		b, err := encodeIdentityDigits(identityIMSI, a.IMSI)
		if err != nil {
			return nil, fmt.Errorf("encoding IMSI: %w", err)
		}
		w.Imsi = &b
	}
	if len(a.TraceReference2) > 0 {
		b := []byte(a.TraceReference2)
		w.TraceReference2 = &b
	}
	return w, nil
}

func convertWireToDeactivateTraceModeArg(w *gsm_map.DeactivateTraceModeArg) (*DeactivateTraceModeArg, error) {
	a := &DeactivateTraceModeArg{TraceReference: HexBytes(w.TraceReference)}
	if w.Imsi != nil {
		s, err := decodeIdentityDigits(identityIMSI, *w.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding IMSI: %w", err)
		}
		a.IMSI = s
	}
	if w.TraceReference2 != nil {
		a.TraceReference2 = HexBytes(*w.TraceReference2)
	}
	return a, nil
}

func convertDeactivateTraceModeResToWire(r *DeactivateTraceModeRes) (*gsm_map.DeactivateTraceModeRes, error) {
	if r == nil {
		return nil, ErrDeactivateTraceModeResNil
	}
	return &gsm_map.DeactivateTraceModeRes{}, nil
}

func convertWireToDeactivateTraceModeRes(_ *gsm_map.DeactivateTraceModeRes) *DeactivateTraceModeRes {
	return &DeactivateTraceModeRes{}
}

// 3GPP TS 29.002 V19.1.0 §17.7.2 defines the extensible depth enumerations,
// restricts medium to RNC, and requires a matching base for each extension.
func validateTraceDepthPair(base *TraceDepth, ext *TraceDepthExtension, rnc bool) error {
	if base != nil {
		switch *base {
		case TraceDepthMinimum, TraceDepthMedium, TraceDepthMaximum:
		default:
			return ErrTraceDepthInvalid
		}
		if !rnc && *base == TraceDepthMedium {
			return ErrTraceDepthMediumNonRNC
		}
	}
	if ext != nil {
		switch *ext {
		case TraceDepthExtensionMinimumWithoutVendorSpecificExtension, TraceDepthExtensionMediumWithoutVendorSpecificExtension, TraceDepthExtensionMaximumWithoutVendorSpecificExtension:
		default:
			return ErrTraceDepthExtensionInvalid
		}
		if !rnc && *ext == TraceDepthExtensionMediumWithoutVendorSpecificExtension {
			return ErrTraceDepthExtensionMediumNonRNC
		}
		if base == nil || int64(*base) != int64(*ext) {
			return ErrTraceDepthExtensionMismatch
		}
	}
	return nil
}

func convertTraceDepthListToWire(v *TraceDepthList) (*gsm_map.TraceDepthList, error) {
	if err := validateTraceDepthPair(v.MscSTraceDepth, v.MscSTraceDepthExtension, false); err != nil {
		return nil, fmt.Errorf("MscSTraceDepth: %w", err)
	}
	if err := validateTraceDepthPair(v.MgwTraceDepth, v.MgwTraceDepthExtension, false); err != nil {
		return nil, fmt.Errorf("MgwTraceDepth: %w", err)
	}
	if err := validateTraceDepthPair(v.SgsnTraceDepth, v.SgsnTraceDepthExtension, false); err != nil {
		return nil, fmt.Errorf("SgsnTraceDepth: %w", err)
	}
	if err := validateTraceDepthPair(v.GgsnTraceDepth, v.GgsnTraceDepthExtension, false); err != nil {
		return nil, fmt.Errorf("GgsnTraceDepth: %w", err)
	}
	if err := validateTraceDepthPair(v.RncTraceDepth, v.RncTraceDepthExtension, true); err != nil {
		return nil, fmt.Errorf("RncTraceDepth: %w", err)
	}
	if err := validateTraceDepthPair(v.BmscTraceDepth, v.BmscTraceDepthExtension, false); err != nil {
		return nil, fmt.Errorf("BmscTraceDepth: %w", err)
	}
	if err := validateTraceDepthPair(v.MmeTraceDepth, v.MmeTraceDepthExtension, false); err != nil {
		return nil, fmt.Errorf("MmeTraceDepth: %w", err)
	}
	if err := validateTraceDepthPair(v.SgwTraceDepth, v.SgwTraceDepthExtension, false); err != nil {
		return nil, fmt.Errorf("SgwTraceDepth: %w", err)
	}
	if err := validateTraceDepthPair(v.PgwTraceDepth, v.PgwTraceDepthExtension, false); err != nil {
		return nil, fmt.Errorf("PgwTraceDepth: %w", err)
	}
	if err := validateTraceDepthPair(v.ENBTraceDepth, v.ENBTraceDepthExtension, false); err != nil {
		return nil, fmt.Errorf("ENBTraceDepth: %w", err)
	}
	return &gsm_map.TraceDepthList{
		MscSTraceDepth:          v.MscSTraceDepth,
		MgwTraceDepth:           v.MgwTraceDepth,
		SgsnTraceDepth:          v.SgsnTraceDepth,
		GgsnTraceDepth:          v.GgsnTraceDepth,
		RncTraceDepth:           v.RncTraceDepth,
		BmscTraceDepth:          v.BmscTraceDepth,
		MmeTraceDepth:           v.MmeTraceDepth,
		SgwTraceDepth:           v.SgwTraceDepth,
		PgwTraceDepth:           v.PgwTraceDepth,
		ENBTraceDepth:           v.ENBTraceDepth,
		MscSTraceDepthExtension: v.MscSTraceDepthExtension,
		MgwTraceDepthExtension:  v.MgwTraceDepthExtension,
		SgsnTraceDepthExtension: v.SgsnTraceDepthExtension,
		GgsnTraceDepthExtension: v.GgsnTraceDepthExtension,
		RncTraceDepthExtension:  v.RncTraceDepthExtension,
		BmscTraceDepthExtension: v.BmscTraceDepthExtension,
		MmeTraceDepthExtension:  v.MmeTraceDepthExtension,
		SgwTraceDepthExtension:  v.SgwTraceDepthExtension,
		PgwTraceDepthExtension:  v.PgwTraceDepthExtension,
		ENBTraceDepthExtension:  v.ENBTraceDepthExtension,
	}, nil
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: non-RNC medium values become minimum.
// Undefined enumeration values and mismatching base/extension pairs are retained.
func traceDepthFromWire(v *TraceDepth, rnc bool) *TraceDepth {
	if v == nil {
		return nil
	}
	out := *v
	if !rnc && out == TraceDepthMedium {
		out = TraceDepthMinimum
	}
	return &out
}

func traceDepthExtensionFromWire(v *TraceDepthExtension, rnc bool) *TraceDepthExtension {
	if v == nil {
		return nil
	}
	out := *v
	if !rnc && out == TraceDepthExtensionMediumWithoutVendorSpecificExtension {
		out = TraceDepthExtensionMinimumWithoutVendorSpecificExtension
	}
	return &out
}

func convertWireToTraceDepthList(v *gsm_map.TraceDepthList) *TraceDepthList {
	return &TraceDepthList{
		MscSTraceDepth:          traceDepthFromWire(v.MscSTraceDepth, false),
		MgwTraceDepth:           traceDepthFromWire(v.MgwTraceDepth, false),
		SgsnTraceDepth:          traceDepthFromWire(v.SgsnTraceDepth, false),
		GgsnTraceDepth:          traceDepthFromWire(v.GgsnTraceDepth, false),
		RncTraceDepth:           traceDepthFromWire(v.RncTraceDepth, true),
		BmscTraceDepth:          traceDepthFromWire(v.BmscTraceDepth, false),
		MmeTraceDepth:           traceDepthFromWire(v.MmeTraceDepth, false),
		SgwTraceDepth:           traceDepthFromWire(v.SgwTraceDepth, false),
		PgwTraceDepth:           traceDepthFromWire(v.PgwTraceDepth, false),
		ENBTraceDepth:           traceDepthFromWire(v.ENBTraceDepth, false),
		MscSTraceDepthExtension: traceDepthExtensionFromWire(v.MscSTraceDepthExtension, false),
		MgwTraceDepthExtension:  traceDepthExtensionFromWire(v.MgwTraceDepthExtension, false),
		SgsnTraceDepthExtension: traceDepthExtensionFromWire(v.SgsnTraceDepthExtension, false),
		GgsnTraceDepthExtension: traceDepthExtensionFromWire(v.GgsnTraceDepthExtension, false),
		RncTraceDepthExtension:  traceDepthExtensionFromWire(v.RncTraceDepthExtension, true),
		BmscTraceDepthExtension: traceDepthExtensionFromWire(v.BmscTraceDepthExtension, false),
		MmeTraceDepthExtension:  traceDepthExtensionFromWire(v.MmeTraceDepthExtension, false),
		SgwTraceDepthExtension:  traceDepthExtensionFromWire(v.SgwTraceDepthExtension, false),
		PgwTraceDepthExtension:  traceDepthExtensionFromWire(v.PgwTraceDepthExtension, false),
		ENBTraceDepthExtension:  traceDepthExtensionFromWire(v.ENBTraceDepthExtension, false),
	}
}

func convertTraceInterfaceListToWire(v *TraceInterfaceList) (*gsm_map.TraceInterfaceList, error) {
	w := &gsm_map.TraceInterfaceList{}
	if v.MscSList != nil {
		b, err := convertMSCSInterfaceListToBitString(v.MscSList)
		if err != nil {
			return nil, fmt.Errorf("MscSList: %w", err)
		}
		w.MscSList = &b
	}
	if v.MgwList != nil {
		b, err := convertMGWInterfaceListToBitString(v.MgwList)
		if err != nil {
			return nil, fmt.Errorf("MgwList: %w", err)
		}
		w.MgwList = &b
	}
	if v.SgsnList != nil {
		b, err := convertSGSNInterfaceListToBitString(v.SgsnList)
		if err != nil {
			return nil, fmt.Errorf("SgsnList: %w", err)
		}
		w.SgsnList = &b
	}
	if v.GgsnList != nil {
		b, err := convertGGSNInterfaceListToBitString(v.GgsnList)
		if err != nil {
			return nil, fmt.Errorf("GgsnList: %w", err)
		}
		w.GgsnList = &b
	}
	if v.RncList != nil {
		b, err := convertRNCInterfaceListToBitString(v.RncList)
		if err != nil {
			return nil, fmt.Errorf("RncList: %w", err)
		}
		w.RncList = &b
	}
	if v.BmscList != nil {
		b, err := convertBMSCInterfaceListToBitString(v.BmscList)
		if err != nil {
			return nil, fmt.Errorf("BmscList: %w", err)
		}
		w.BmscList = &b
	}
	if v.MmeList != nil {
		b, err := convertMMEInterfaceListToBitString(v.MmeList)
		if err != nil {
			return nil, fmt.Errorf("MmeList: %w", err)
		}
		w.MmeList = &b
	}
	if v.SgwList != nil {
		b, err := convertSGWInterfaceListToBitString(v.SgwList)
		if err != nil {
			return nil, fmt.Errorf("SgwList: %w", err)
		}
		w.SgwList = &b
	}
	if v.PgwList != nil {
		b, err := convertPGWInterfaceListToBitString(v.PgwList)
		if err != nil {
			return nil, fmt.Errorf("PgwList: %w", err)
		}
		w.PgwList = &b
	}
	if v.ENBList != nil {
		b, err := convertENBInterfaceListToBitString(v.ENBList)
		if err != nil {
			return nil, fmt.Errorf("ENBList: %w", err)
		}
		w.ENBList = &b
	}
	return w, nil
}

func convertWireToTraceInterfaceList(w *gsm_map.TraceInterfaceList) *TraceInterfaceList {
	v := &TraceInterfaceList{}
	if w.MscSList != nil {
		v.MscSList = convertBitStringToMSCSInterfaceList(*w.MscSList)
	}
	if w.MgwList != nil {
		v.MgwList = convertBitStringToMGWInterfaceList(*w.MgwList)
	}
	if w.SgsnList != nil {
		v.SgsnList = convertBitStringToSGSNInterfaceList(*w.SgsnList)
	}
	if w.GgsnList != nil {
		v.GgsnList = convertBitStringToGGSNInterfaceList(*w.GgsnList)
	}
	if w.RncList != nil {
		v.RncList = convertBitStringToRNCInterfaceList(*w.RncList)
	}
	if w.BmscList != nil {
		v.BmscList = convertBitStringToBMSCInterfaceList(*w.BmscList)
	}
	if w.MmeList != nil {
		v.MmeList = convertBitStringToMMEInterfaceList(*w.MmeList)
	}
	if w.SgwList != nil {
		v.SgwList = convertBitStringToSGWInterfaceList(*w.SgwList)
	}
	if w.PgwList != nil {
		v.PgwList = convertBitStringToPGWInterfaceList(*w.PgwList)
	}
	if w.ENBList != nil {
		v.ENBList = convertBitStringToENBInterfaceList(*w.ENBList)
	}
	return v
}

func convertTraceEventListToWire(v *TraceEventList) (*gsm_map.TraceEventList, error) {
	w := &gsm_map.TraceEventList{}
	if v.MscSList != nil {
		b, err := convertMSCSEventListToBitString(v.MscSList)
		if err != nil {
			return nil, fmt.Errorf("MscSList: %w", err)
		}
		w.MscSList = &b
	}
	if v.MgwList != nil {
		b, err := convertMGWEventListToBitString(v.MgwList)
		if err != nil {
			return nil, fmt.Errorf("MgwList: %w", err)
		}
		w.MgwList = &b
	}
	if v.SgsnList != nil {
		b, err := convertSGSNEventListToBitString(v.SgsnList)
		if err != nil {
			return nil, fmt.Errorf("SgsnList: %w", err)
		}
		w.SgsnList = &b
	}
	if v.GgsnList != nil {
		b, err := convertGGSNEventListToBitString(v.GgsnList)
		if err != nil {
			return nil, fmt.Errorf("GgsnList: %w", err)
		}
		w.GgsnList = &b
	}
	if v.BmscList != nil {
		b, err := convertBMSCEventListToBitString(v.BmscList)
		if err != nil {
			return nil, fmt.Errorf("BmscList: %w", err)
		}
		w.BmscList = &b
	}
	if v.MmeList != nil {
		b, err := convertMMEEventListToBitString(v.MmeList)
		if err != nil {
			return nil, fmt.Errorf("MmeList: %w", err)
		}
		w.MmeList = &b
	}
	if v.SgwList != nil {
		b, err := convertSGWEventListToBitString(v.SgwList)
		if err != nil {
			return nil, fmt.Errorf("SgwList: %w", err)
		}
		w.SgwList = &b
	}
	if v.PgwList != nil {
		b, err := convertPGWEventListToBitString(v.PgwList)
		if err != nil {
			return nil, fmt.Errorf("PgwList: %w", err)
		}
		w.PgwList = &b
	}
	return w, nil
}

func convertWireToTraceEventList(w *gsm_map.TraceEventList) *TraceEventList {
	v := &TraceEventList{}
	if w.MscSList != nil {
		v.MscSList = convertBitStringToMSCSEventList(*w.MscSList)
	}
	if w.MgwList != nil {
		v.MgwList = convertBitStringToMGWEventList(*w.MgwList)
	}
	if w.SgsnList != nil {
		v.SgsnList = convertBitStringToSGSNEventList(*w.SgsnList)
	}
	if w.GgsnList != nil {
		v.GgsnList = convertBitStringToGGSNEventList(*w.GgsnList)
	}
	if w.BmscList != nil {
		v.BmscList = convertBitStringToBMSCEventList(*w.BmscList)
	}
	if w.MmeList != nil {
		v.MmeList = convertBitStringToMMEEventList(*w.MmeList)
	}
	if w.SgwList != nil {
		v.SgwList = convertBitStringToSGWEventList(*w.SgwList)
	}
	if w.PgwList != nil {
		v.PgwList = convertBitStringToPGWEventList(*w.PgwList)
	}
	return v
}

func convertMDTConfigurationToWire(v *MDTConfiguration) (*gsm_map.MDTConfiguration, error) {
	w := &gsm_map.MDTConfiguration{JobType: v.JobType}
	if v.AreaScope != nil {
		a, err := convertAreaScopeToWire(v.AreaScope)
		if err != nil {
			return nil, err
		}
		w.AreaScope = a
	}
	if len(v.ListOfMeasurements) > 0 {
		b := []byte(v.ListOfMeasurements)
		w.ListOfMeasurements = &b
	}
	if len(v.ReportingTrigger) > 0 {
		b := []byte(v.ReportingTrigger)
		w.ReportingTrigger = &b
	}
	w.ReportInterval = v.ReportInterval
	w.ReportAmount = v.ReportAmount
	w.EventThresholdRSRP = intPtrTo64(v.EventThresholdRSRP)
	w.EventThresholdRSRQ = intPtrTo64(v.EventThresholdRSRQ)
	w.LoggingInterval = v.LoggingInterval
	w.LoggingDuration = v.LoggingDuration
	w.MeasurementPeriodUMTS = v.MeasurementPeriodUMTS
	w.MeasurementPeriodLTE = v.MeasurementPeriodLTE
	w.CollectionPeriodRRMUMTS = v.CollectionPeriodRRMUMTS
	w.CollectionPeriodRRMLTE = v.CollectionPeriodRRMLTE
	if len(v.PositioningMethod) > 0 {
		b := []byte(v.PositioningMethod)
		w.PositioningMethod = &b
	}
	if len(v.MeasurementQuantity) > 0 {
		b := []byte(v.MeasurementQuantity)
		w.MeasurementQuantity = &b
	}
	w.EventThreshold1F = intPtrTo64(v.EventThreshold1F)
	w.EventThreshold1I = intPtrTo64(v.EventThreshold1I)
	if v.MDTAllowedPLMNList != nil {
		a, err := traceIdentityListToWire(v.MDTAllowedPLMNList, "MDTAllowedPLMNList")
		if err != nil {
			return nil, err
		}
		w.MdtAllowedPLMNList = &gsm_map.MDTAllowedPLMNIdList{Values: a}
	}
	return w, nil
}

func convertWireToMDTConfiguration(w *gsm_map.MDTConfiguration) *MDTConfiguration {
	v := &MDTConfiguration{JobType: w.JobType}
	if w.AreaScope != nil {
		v.AreaScope = convertWireToAreaScope(w.AreaScope)
	}
	if w.ListOfMeasurements != nil {
		v.ListOfMeasurements = HexBytes(*w.ListOfMeasurements)
	}
	if w.ReportingTrigger != nil {
		v.ReportingTrigger = HexBytes(*w.ReportingTrigger)
	}
	v.ReportInterval = w.ReportInterval
	v.ReportAmount = w.ReportAmount
	v.EventThresholdRSRP = int64PtrTo(w.EventThresholdRSRP)
	v.EventThresholdRSRQ = int64PtrTo(w.EventThresholdRSRQ)
	v.LoggingInterval = w.LoggingInterval
	v.LoggingDuration = w.LoggingDuration
	v.MeasurementPeriodUMTS = w.MeasurementPeriodUMTS
	v.MeasurementPeriodLTE = w.MeasurementPeriodLTE
	v.CollectionPeriodRRMUMTS = w.CollectionPeriodRRMUMTS
	v.CollectionPeriodRRMLTE = w.CollectionPeriodRRMLTE
	if w.PositioningMethod != nil {
		v.PositioningMethod = HexBytes(*w.PositioningMethod)
	}
	if w.MeasurementQuantity != nil {
		v.MeasurementQuantity = HexBytes(*w.MeasurementQuantity)
	}
	v.EventThreshold1F = int64PtrTo(w.EventThreshold1F)
	v.EventThreshold1I = int64PtrTo(w.EventThreshold1I)
	if w.MdtAllowedPLMNList != nil {
		v.MDTAllowedPLMNList = traceIdentityListFromWire(w.MdtAllowedPLMNList.Values)
	}
	return v
}

func convertAreaScopeToWire(v *AreaScope) (*gsm_map.AreaScope, error) {
	w := &gsm_map.AreaScope{}
	if v.CgiList != nil {
		a, err := traceIdentityListToWire(v.CgiList, "AreaScope.CgiList")
		if err != nil {
			return nil, err
		}
		w.CgiList = &gsm_map.CGIList{Values: a}
	}
	if v.EUtranCgiList != nil {
		a, err := traceIdentityListToWire(v.EUtranCgiList, "AreaScope.EUtranCgiList")
		if err != nil {
			return nil, err
		}
		w.EUtranCgiList = &gsm_map.EUTRANCGIList{Values: a}
	}
	if v.RoutingAreaIdList != nil {
		a, err := traceIdentityListToWire(v.RoutingAreaIdList, "AreaScope.RoutingAreaIdList")
		if err != nil {
			return nil, err
		}
		w.RoutingAreaIdList = &gsm_map.RoutingAreaIdList{Values: a}
	}
	if v.LocationAreaIdList != nil {
		a, err := traceIdentityListToWire(v.LocationAreaIdList, "AreaScope.LocationAreaIdList")
		if err != nil {
			return nil, err
		}
		w.LocationAreaIdList = &gsm_map.LocationAreaIdList{Values: a}
	}
	if v.TrackingAreaIdList != nil {
		a, err := traceIdentityListToWire(v.TrackingAreaIdList, "AreaScope.TrackingAreaIdList")
		if err != nil {
			return nil, err
		}
		w.TrackingAreaIdList = &gsm_map.TrackingAreaIdList{Values: a}
	}
	return w, nil
}

func convertWireToAreaScope(w *gsm_map.AreaScope) *AreaScope {
	v := &AreaScope{}
	if w.CgiList != nil {
		v.CgiList = traceIdentityListFromWire(w.CgiList.Values)
	}
	if w.EUtranCgiList != nil {
		v.EUtranCgiList = traceIdentityListFromWire(w.EUtranCgiList.Values)
	}
	if w.RoutingAreaIdList != nil {
		v.RoutingAreaIdList = traceIdentityListFromWire(w.RoutingAreaIdList.Values)
	}
	if w.LocationAreaIdList != nil {
		v.LocationAreaIdList = traceIdentityListFromWire(w.LocationAreaIdList.Values)
	}
	if w.TrackingAreaIdList != nil {
		v.TrackingAreaIdList = traceIdentityListFromWire(w.TrackingAreaIdList.Values)
	}
	return v
}

// The common identities begin with PLMN digit coding (3GPP TS 29.002
// V19.1.0 §17.7.8); use the same sender check as other fields of these types.
func traceIdentityListToWire(list []HexBytes, field string) ([][]byte, error) {
	out := make([][]byte, len(list))
	for i, raw := range list {
		if err := validatePLMNIdentityDigits(raw, fmt.Sprintf("%s[%d]", field, i)); err != nil {
			return nil, err
		}
		out[i] = raw
	}
	return out, nil
}

func traceIdentityListFromWire(list [][]byte) []HexBytes {
	out := make([]HexBytes, len(list))
	for i, raw := range list {
		out[i] = HexBytes(raw)
	}
	return out
}
