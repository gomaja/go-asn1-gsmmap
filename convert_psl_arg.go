// convert_psl_arg.go
//
// Top-level converter for ProvideSubscriberLocationArg (opCode 83).
// PR D4 of the staged PSL implementation: wires the leaf, LCS-Client,
// and area-event/periodic/PLMN-list converters from PRs #43, #44, and
// #45 into a single arg encoder/decoder pair. Marshal()/Parse() entry
// points live in marshal.go / parse.go.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// IMSI/LMSI/IMEI sizes per TS 29.002 MAP-CommonDataTypes.asn.
const (
	pslIMSIMinLen = 3
	pslIMSIMaxLen = 8
	pslLMSILen    = 4
	pslIMEILen    = 8

	// LCSServiceTypeID INTEGER (0..127) per TS 29.002
	// MAP-CommonDataTypes.asn:436.
	pslLcsServiceTypeIDMin int64 = 0
	pslLcsServiceTypeIDMax int64 = 127
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

	if len(a.IMSI) > 0 {
		if len(a.IMSI) < pslIMSIMinLen || len(a.IMSI) > pslIMSIMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.IMSI len=%d: %w", len(a.IMSI), ErrPSLArgIMSIInvalidSize)
		}
		v := gsm_map.IMSI(a.IMSI)
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
		if len(a.LMSI) != pslLMSILen {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LMSI len=%d: %w", len(a.LMSI), ErrPSLArgLMSIInvalidSize)
		}
		v := gsm_map.LMSI(a.LMSI)
		out.Lmsi = &v
	}
	if len(a.IMEI) > 0 {
		if len(a.IMEI) != pslIMEILen {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.IMEI len=%d: %w", len(a.IMEI), ErrPSLArgIMEIInvalidSize)
		}
		v := gsm_map.IMEI(a.IMEI)
		out.Imei = &v
	}
	if len(a.LcsPriority) > 0 {
		if len(a.LcsPriority) != 1 {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsPriority len=%d: %w", len(a.LcsPriority), ErrLCSPriorityInvalidSize)
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
		if len(a.LcsReferenceNumber) != 1 {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsReferenceNumber len=%d: %w", len(a.LcsReferenceNumber), ErrLCSReferenceNumberInvalidSize)
		}
		v := gsm_map.LCSReferenceNumber(a.LcsReferenceNumber)
		out.LcsReferenceNumber = &v
	}
	if a.LcsServiceTypeID != nil {
		v := *a.LcsServiceTypeID
		if v < pslLcsServiceTypeIDMin || v > pslLcsServiceTypeIDMax {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsServiceTypeID=%d: %w", v, ErrPSLArgLcsServiceTypeIDOutOfRange)
		}
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
	if len(a.HGmlcAddress) > 0 {
		v := gsm_map.GSNAddress(a.HGmlcAddress)
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
// struct back to the public type. Symmetric validation: rejects out-
// of-range values consistent with the encoder; preserves unknown
// values for extensible enums per Postel.
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
		if len(*w.Imsi) < pslIMSIMinLen || len(*w.Imsi) > pslIMSIMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.IMSI len=%d: %w", len(*w.Imsi), ErrPSLArgIMSIInvalidSize)
		}
		out.IMSI = HexBytes(*w.Imsi)
	}
	if w.Msisdn != nil {
		s, nature, plan, err := decodeAddressField([]byte(*w.Msisdn))
		if err != nil {
			return nil, fmt.Errorf("decoding ProvideSubscriberLocationArg.MSISDN: %w", err)
		}
		out.MSISDN = s
		out.MSISDNNature = nature
		out.MSISDNPlan = plan
	}
	if w.Lmsi != nil {
		if len(*w.Lmsi) != pslLMSILen {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LMSI len=%d: %w", len(*w.Lmsi), ErrPSLArgLMSIInvalidSize)
		}
		out.LMSI = HexBytes(*w.Lmsi)
	}
	if w.Imei != nil {
		if len(*w.Imei) != pslIMEILen {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.IMEI len=%d: %w", len(*w.Imei), ErrPSLArgIMEIInvalidSize)
		}
		out.IMEI = HexBytes(*w.Imei)
	}
	if w.LcsPriority != nil {
		if len(*w.LcsPriority) != 1 {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsPriority len=%d: %w", len(*w.LcsPriority), ErrLCSPriorityInvalidSize)
		}
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
		if len(*w.LcsReferenceNumber) != 1 {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsReferenceNumber len=%d: %w", len(*w.LcsReferenceNumber), ErrLCSReferenceNumberInvalidSize)
		}
		out.LcsReferenceNumber = LCSReferenceNumber(*w.LcsReferenceNumber)
	}
	if w.LcsServiceTypeID != nil {
		v := int64(*w.LcsServiceTypeID)
		if v < pslLcsServiceTypeIDMin || v > pslLcsServiceTypeIDMax {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsServiceTypeID=%d: %w", v, ErrPSLArgLcsServiceTypeIDOutOfRange)
		}
		out.LcsServiceTypeID = &v
	}
	if w.LcsCodeword != nil {
		v, err := convertWireToLCSCodeword(w.LcsCodeword)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsCodeword: %w", err)
		}
		out.LcsCodeword = v
	}
	if w.LcsPrivacyCheck != nil {
		v, err := convertWireToLCSPrivacyCheck(w.LcsPrivacyCheck)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.LcsPrivacyCheck: %w", err)
		}
		out.LcsPrivacyCheck = v
	}
	if w.AreaEventInfo != nil {
		v, err := convertWireToAreaEventInfo(w.AreaEventInfo)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationArg.AreaEventInfo: %w", err)
		}
		out.AreaEventInfo = v
	}
	if w.HGmlcAddress != nil {
		out.HGmlcAddress = HexBytes(*w.HGmlcAddress)
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
