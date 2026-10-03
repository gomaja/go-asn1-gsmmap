// convert_psl_arg.go
//
// Top-level converter for ProvideSubscriberLocationArg (opCode 83): wires
// the leaf, LCS-Client, and area-event/periodic/PLMN-list converters into a
// single arg encoder/decoder pair. Marshal()/Parse() entry points live in
// marshal.go / parse.go.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

	"github.com/gomaja/go-asn1-gsmmap/gsn"
)

// lcsPriorityNormal is LCS-Priority 1, normal priority; 0 is the highest
// (3GPP TS 29.002 V19.1.0 §17.7.13).
const lcsPriorityNormal = 0x01

// convertProvideSubscriberLocationArgToWire builds the wire-form
// gsm_map.ProvideSubscriberLocationArg from the public type. Semantic
// validation errors carry field context and the relevant sentinel.
func convertProvideSubscriberLocationArgToWire(a *ProvideSubscriberLocationArg) (*gsm_map.ProvideSubscriberLocationArg, error) {
	if a == nil {
		return nil, ErrPSLArgNil
	}

	// Mandatory: LocationType.
	loc, err := convertLocationTypeToWire(&a.LocationType)
	if err != nil {
		return nil, fmt.Errorf("ProvideSubscriberLocationArg.LocationType: %w", err)
	}

	// Mandatory: MlcNumber digits.
	if a.MlcNumber == "" {
		return nil, ErrPSLArgMlcNumberEmpty
	}
	mlcWire, err := encodeAddressField(a.MlcNumber, a.MlcNumberNature, a.MlcNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding ProvideSubscriberLocationArg.MlcNumber: %w", err)
	}

	out := &gsm_map.ProvideSubscriberLocationArg{
		LocationType: *loc,
		MlcNumber:    mlcWire,
	}

	// Optional fields.
	if a.LcsClientID != nil {
		v, err := convertLCSClientIDToWire(a.LcsClientID)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsClientID: %w", err)
		}
		out.LcsClientID = v
	}
	out.PrivacyOverride = boolToNullPtr(a.PrivacyOverride)

	if a.IMSI != "" {
		imsiBytes, err := encodeIdentityDigits(identityIMSI, a.IMSI)
		if err != nil {
			return nil, fmt.Errorf("encoding ProvideSubscriberLocationArg.IMSI: %w", err)
		}
		v := imsiBytes
		out.Imsi = &v
	}
	if a.MSISDN != "" {
		isdn, err := encodeAddressField(a.MSISDN, a.MSISDNNature, a.MSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding ProvideSubscriberLocationArg.MSISDN: %w", err)
		}
		v := isdn
		out.Msisdn = &v
	}
	if len(a.LMSI) > 0 {
		v := gsm_map.LMSI(a.LMSI)
		out.Lmsi = &v
	}
	if a.IMEI != "" {
		imeiBytes, err := encodeIdentityDigits(identityIMEI, a.IMEI)
		if err != nil {
			return nil, fmt.Errorf("encoding ProvideSubscriberLocationArg.IMEI: %w", err)
		}
		v := imeiBytes
		out.Imei = &v
	}
	if len(a.LcsPriority) > 0 {
		// §17.7.13 LCS-Priority: 0 is the highest and 1 the normal priority;
		// "all other values treated as 1", so only 0 and 1 are sent. The
		// codec checks SIZE (1).
		if len(a.LcsPriority) == 1 && a.LcsPriority[0] > lcsPriorityNormal {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsPriority=%x: %w", []byte(a.LcsPriority), ErrLCSPriorityInvalid)
		}
		v := gsm_map.LCSPriority(a.LcsPriority)
		out.LcsPriority = &v
	}
	if a.LcsQoS != nil {
		v, err := convertLCSQoSToWire(a.LcsQoS)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsQoS: %w", err)
		}
		out.LcsQoS = v
	}
	if a.SupportedGADShapes != nil {
		bs := convertSupportedGADShapesToBitString(a.SupportedGADShapes)
		out.SupportedGADShapes = &bs
	}
	if len(a.LcsReferenceNumber) > 0 {
		v := gsm_map.LCSReferenceNumber(a.LcsReferenceNumber)
		out.LcsReferenceNumber = &v
	}
	if a.LcsServiceTypeID != nil {
		v := *a.LcsServiceTypeID

		w := v
		out.LcsServiceTypeID = &w
	}
	if a.LcsCodeword != nil {
		out.LcsCodeword = convertLCSCodewordToWire(a.LcsCodeword)
	}
	if a.LcsPrivacyCheck != nil {
		v, err := convertLCSPrivacyCheckToWire(a.LcsPrivacyCheck)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsPrivacyCheck: %w", err)
		}
		out.LcsPrivacyCheck = v
	}
	if a.AreaEventInfo != nil {
		v, err := convertAreaEventInfoToWire(a.AreaEventInfo)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.AreaEventInfo: %w", err)
		}
		out.AreaEventInfo = v
	}
	if a.HGmlcAddress != "" {
		gsnAddr, err := gsn.Build(a.HGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("encoding ProvideSubscriberLocationArg.HGmlcAddress: %w", err)
		}
		v := gsnAddr
		out.HGmlcAddress = &v
	}
	out.MoLrShortCircuitIndicator = boolToNullPtr(a.MoLrShortCircuitIndicator)

	if a.PeriodicLDRInfo != nil {
		v, err := convertPeriodicLDRInfoToWire(a.PeriodicLDRInfo)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.PeriodicLDRInfo: %w", err)
		}
		out.PeriodicLDRInfo = v
	}
	if a.ReportingPLMNList != nil {
		v, err := convertReportingPLMNListToWire(a.ReportingPLMNList)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.ReportingPLMNList: %w", err)
		}
		out.ReportingPLMNList = v
	}

	return out, nil
}

// convertWireToProvideSubscriberLocationArg unmarshals the wire-form
// struct back to the public type. Validation rules:
//   - Round-trip safety: present-but-empty MlcNumber/MSISDN decoded
//     values are rejected (cannot round-trip through the string API).
//   - An unrecognized LocationEstimateType or PrivacyCheckRelatedAction,
//     or a DeferredLocationEventType bit other than msAvailable(0) to
//     periodicLDR(4), rejects the argument, as does an unrecognized
//     LCSClientType without
//     privacyOverride (3GPP TS 29.002 V19.1.0 §17.7.13); the caller
//     answers with unexpected data value. With privacyOverride an
//     unrecognized LCSClientType is kept.
//   - Other extensible enums (LCSFormatIndicator,
//     AccuracyFulfilmentIndicator, AreaType, OccurrenceInfo,
//     RANTechnology): unknown values preserved per Postel; encoder-side
//     strictness lives in the leaf converters.
//   - ExtensionContainer at tag [8]: dropped (opaque metadata not
//     surfaced; see ProvideSubscriberLocationArg doc).
func convertWireToProvideSubscriberLocationArg(w *gsm_map.ProvideSubscriberLocationArg) (*ProvideSubscriberLocationArg, error) {
	if w == nil {
		return nil, ErrPSLArgNil
	}

	// 3GPP TS 29.002 V19.1.0 §17.7.13: "a ProvideSubscriberLocation-Arg
	// containing an unrecognized LocationEstimateType shall be rejected by
	// the receiver with a return error cause of unexpected data value". The
	// clause says the same of PrivacyCheckRelatedAction, and of LCSClientType
	// unless the client uses the privacy override.
	if v := w.LocationType.LocationEstimateType; !isRecognizedLocationEstimateType(v) {
		return nil, fmt.Errorf("ProvideSubscriberLocationArg.LocationType.LocationEstimateType=%d: %w", v, ErrLocationEstimateTypeUnrecognized)
	}
	// §17.7.13 DeferredLocationEventType: "a ProvideSubscriberLocation-Arg
	// containing other values than listed above in DeferredLocationEventType
	// shall be rejected by the receiver with a return error cause of
	// unexpected data value".
	if d := w.LocationType.DeferredLocationEventType; d != nil && hasUnlistedDeferredLocationEvent(*d) {
		return nil, fmt.Errorf("ProvideSubscriberLocationArg.LocationType.DeferredLocationEventType=%x/%d: %w", d.Bytes, d.BitLength, ErrDeferredLocationEventTypeUnrecognized)
	}
	if p := w.LcsPrivacyCheck; p != nil {
		if !isRecognizedPrivacyCheckRelatedAction(p.CallSessionUnrelated) {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsPrivacyCheck.CallSessionUnrelated=%d: %w", p.CallSessionUnrelated, ErrPrivacyCheckRelatedActionUnrecognized)
		}
		if r := p.CallSessionRelated; r != nil && !isRecognizedPrivacyCheckRelatedAction(*r) {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsPrivacyCheck.CallSessionRelated=%d: %w", *r, ErrPrivacyCheckRelatedActionUnrecognized)
		}
	}
	if err := checkLCSClientType(w.LcsClientID, w.PrivacyOverride != nil); err != nil {
		return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsClientID: %w", err)
	}

	loc := convertWireToLocationType(&w.LocationType)
	mlcStr, mlcNature, mlcPlan, err := decodeAddressField(w.MlcNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding ProvideSubscriberLocationArg.MlcNumber: %w", err)
	}
	if mlcStr == "" {
		return nil, ErrPSLArgMlcNumberEmpty
	}

	out := &ProvideSubscriberLocationArg{
		LocationType:    *loc,
		MlcNumber:       mlcStr,
		MlcNumberNature: mlcNature,
		MlcNumberPlan:   mlcPlan,
	}

	if w.LcsClientID != nil {
		v, err := convertWireToLCSClientID(w.LcsClientID)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsClientID: %w", err)
		}
		out.LcsClientID = v
	}
	out.PrivacyOverride = nullPtrToBool(w.PrivacyOverride)

	if w.Imsi != nil {
		imsi, err := decodeIdentityDigits(identityIMSI, *w.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding ProvideSubscriberLocationArg.IMSI: %w", err)
		}
		out.IMSI = imsi
	}
	if w.Msisdn != nil {
		s, nature, plan, err := decodeAddressField(*w.Msisdn)
		if err != nil {
			return nil, fmt.Errorf("decoding ProvideSubscriberLocationArg.MSISDN: %w", err)
		}
		if s == "" {
			return nil, ErrPSLArgMSISDNDecodedEmpty
		}
		out.MSISDN = s
		out.MSISDNNature = nature
		out.MSISDNPlan = plan
	}
	if w.Lmsi != nil {
		out.LMSI = HexBytes(*w.Lmsi)
	}
	if w.Imei != nil {
		imei, err := decodeIdentityDigits(identityIMEI, *w.Imei)
		if err != nil {
			return nil, fmt.Errorf("decoding ProvideSubscriberLocationArg.IMEI: %w", err)
		}
		out.IMEI = imei
	}
	if w.LcsPriority != nil {
		// §17.7.13 LCS-Priority: "all other values treated as 1".
		out.LcsPriority = LCSPriority(*w.LcsPriority)
		if len(out.LcsPriority) == 1 && out.LcsPriority[0] > lcsPriorityNormal {
			out.LcsPriority = LCSPriority{lcsPriorityNormal}
		}
	}
	if w.LcsQoS != nil {
		v, err := convertWireToLCSQoS(w.LcsQoS)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsQoS: %w", err)
		}
		out.LcsQoS = v
	}
	if w.SupportedGADShapes != nil {
		out.SupportedGADShapes = convertBitStringToSupportedGADShapes(*w.SupportedGADShapes)
	}
	if w.LcsReferenceNumber != nil {
		out.LcsReferenceNumber = LCSReferenceNumber(*w.LcsReferenceNumber)
	}
	if w.LcsServiceTypeID != nil {
		v := *w.LcsServiceTypeID

		out.LcsServiceTypeID = &v
	}
	if w.LcsCodeword != nil {
		out.LcsCodeword = convertWireToLCSCodeword(w.LcsCodeword)
	}
	out.LcsPrivacyCheck = convertWireToLCSPrivacyCheck(w.LcsPrivacyCheck)
	if w.AreaEventInfo != nil {
		out.AreaEventInfo = convertWireToAreaEventInfo(w.AreaEventInfo)
	}
	if w.HGmlcAddress != nil {
		addr, err := gsn.Parse(*w.HGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("decoding ProvideSubscriberLocationArg.HGmlcAddress: %w", err)
		}
		out.HGmlcAddress = addr
	}
	out.MoLrShortCircuitIndicator = nullPtrToBool(w.MoLrShortCircuitIndicator)

	if w.PeriodicLDRInfo != nil {
		v, err := convertWireToPeriodicLDRInfo(w.PeriodicLDRInfo)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.PeriodicLDRInfo: %w", err)
		}
		out.PeriodicLDRInfo = v
	}
	if w.ReportingPLMNList != nil {
		out.ReportingPLMNList = convertWireToReportingPLMNList(w.ReportingPLMNList)
	}

	return out, nil
}
