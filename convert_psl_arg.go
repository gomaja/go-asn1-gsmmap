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

// convertProvideSubscriberLocationArgToWire builds the wire-form
// gsm_map.ProvideSubscriberLocationArg from the public type. Validates
// every field; the first error is returned with field context wrapped
// via %w on the relevant sentinel.
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
		MlcNumber:    gsm_map.ISDNAddressString(mlcWire),
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
		v := gsm_map.IMSI(imsiBytes)
		out.Imsi = &v
	}
	if a.MSISDN != "" {
		isdn, err := encodeAddressField(a.MSISDN, a.MSISDNNature, a.MSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding ProvideSubscriberLocationArg.MSISDN: %w", err)
		}
		v := gsm_map.ISDNAddressString(isdn)
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
		v := gsm_map.IMEI(imeiBytes)
		out.Imei = &v
	}
	if len(a.LcsPriority) > 0 {
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

		w := gsm_map.LCSServiceTypeID(v)
		out.LcsServiceTypeID = &w
	}
	if a.LcsCodeword != nil {
		v, err := convertLCSCodewordToWire(a.LcsCodeword)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsCodeword: %w", err)
		}
		out.LcsCodeword = v
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
		v := gsm_map.GSNAddress(gsnAddr)
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
//   - Extensible enums (LocationEstimateType, LCSClientType,
//     LCSFormatIndicator, AccuracyFulfilmentIndicator, AreaType,
//     OccurrenceInfo, RANTechnology): unknown values preserved per
//     Postel; encoder-side strictness lives in the leaf converters.
//   - ExtensionContainer at tag [8]: dropped (opaque metadata not
//     surfaced; see ProvideSubscriberLocationArg doc).
func convertWireToProvideSubscriberLocationArg(w *gsm_map.ProvideSubscriberLocationArg) (*ProvideSubscriberLocationArg, error) {
	if w == nil {
		return nil, ErrPSLArgNil
	}

	loc, err := convertWireToLocationType(&w.LocationType)
	if err != nil {
		return nil, fmt.Errorf("ProvideSubscriberLocationArg.LocationType: %w", err)
	}
	mlcStr, mlcNature, mlcPlan, err := decodeAddressField([]byte(w.MlcNumber))
	if err != nil {
		return nil, fmt.Errorf("decoding ProvideSubscriberLocationArg.MlcNumber: %w", err)
	}
	if mlcStr == "" {
		return nil, ErrPSLArgMlcNumberDecodedEmpty
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
		s, nature, plan, err := decodeAddressField([]byte(*w.Msisdn))
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
		out.LcsPriority = LCSPriority(*w.LcsPriority)
	}
	if w.LcsQoS != nil {
		v, err := convertWireToLCSQoS(w.LcsQoS)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsQoS: %w", err)
		}
		out.LcsQoS = v
	}
	if w.SupportedGADShapes != nil {
		v, err := convertBitStringToSupportedGADShapes(*w.SupportedGADShapes)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.SupportedGADShapes: %w", err)
		}
		out.SupportedGADShapes = v
	}
	if w.LcsReferenceNumber != nil {
		out.LcsReferenceNumber = LCSReferenceNumber(*w.LcsReferenceNumber)
	}
	if w.LcsServiceTypeID != nil {
		v := int64(*w.LcsServiceTypeID)

		out.LcsServiceTypeID = &v
	}
	if w.LcsCodeword != nil {
		v, err := convertWireToLCSCodeword(w.LcsCodeword)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsCodeword: %w", err)
		}
		out.LcsCodeword = v
	}
	out.LcsPrivacyCheck = convertWireToLCSPrivacyCheck(w.LcsPrivacyCheck)
	if w.AreaEventInfo != nil {
		v, err := convertWireToAreaEventInfo(w.AreaEventInfo)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.AreaEventInfo: %w", err)
		}
		out.AreaEventInfo = v
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
		v, err := convertWireToReportingPLMNList(w.ReportingPLMNList)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.ReportingPLMNList: %w", err)
		}
		out.ReportingPLMNList = v
	}

	return out, nil
}
