// convert_slr_arg.go
//
// Top-level converter for SubscriberLocationReportArg (opCode 86).
// PR G3 of the staged SLR implementation: wires the foundation types
// (PR #53) and sub-converters (PR #54) — together with the LCS leaf,
// CellIdOrSai, ServingNodeAddress, and PeriodicLDRInfo converters
// already built for ProvideSubscriberLocation — into a single arg
// encoder/decoder pair. Marshal()/Parse() entry points live in
// marshal.go / parse.go.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

	"github.com/gomaja/go-asn1-gsmmap/gsn"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// LcsEvent value bounds per TS 29.002 MAP-LCS-DataTypes.asn:681
// (ENUMERATED 0..5, extensible). Encoder strict, decoder lenient.
const (
	slrLcsEventMin int64 = 0
	slrLcsEventMax int64 = 5
)

// convertSubscriberLocationReportArgToWire builds the wire-form
// gsm_map.SubscriberLocationReportArg from the public type. Validates
// every field; the first error is returned with field context wrapped
// via %w on the relevant sentinel.
//
// Per spec, one of MSISDN or IMSI must be present. That cross-field
// invariant is the caller's responsibility — the encoder does not
// reject its absence, since reporting nodes may relay partial data and
// the receiving GMLC correlates the subscriber by other means.
func convertSubscriberLocationReportArgToWire(a *SubscriberLocationReportArg) (*gsm_map.SubscriberLocationReportArg, error) {
	if a == nil {
		return nil, ErrSLRArgNil
	}

	// Mandatory: LcsEvent (extensible enum; encoder strict 0..5).
	if int64(a.LcsEvent) < slrLcsEventMin || int64(a.LcsEvent) > slrLcsEventMax {
		return nil, fmt.Errorf("SubscriberLocationReportArg.LcsEvent=%d: %w", int64(a.LcsEvent), ErrLCSEventInvalid)
	}

	// Mandatory: LcsClientID.
	clientID, err := convertLCSClientIDToWire(&a.LcsClientID)
	if err != nil {
		return nil, fmt.Errorf("SubscriberLocationReportArg.LcsClientID: %w", err)
	}

	// Mandatory: LcsLocationInfo.
	locInfo, err := convertLCSLocationInfoToWire(&a.LcsLocationInfo)
	if err != nil {
		return nil, fmt.Errorf("SubscriberLocationReportArg.LcsLocationInfo: %w", err)
	}

	out := &gsm_map.SubscriberLocationReportArg{
		LcsEvent:        a.LcsEvent,
		LcsClientID:     *clientID,
		LcsLocationInfo: *locInfo,
	}

	// [0] msisdn
	if a.MSISDN != "" {
		isdn, err := encodeAddressField(a.MSISDN, a.MSISDNNature, a.MSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SubscriberLocationReportArg.MSISDN: %w", err)
		}
		v := gsm_map.ISDNAddressString(isdn)
		out.Msisdn = &v
	}
	// [1] imsi
	if a.IMSI != "" {
		if len(a.IMSI) < pslIMSIDigitsMin || len(a.IMSI) > pslIMSIDigitsMax {
			return nil, fmt.Errorf("SubscriberLocationReportArg.IMSI digits=%d: %w", len(a.IMSI), ErrSLRArgIMSIInvalidSize)
		}
		imsiBytes, err := tbcd.Encode(a.IMSI)
		if err != nil {
			return nil, fmt.Errorf("encoding SubscriberLocationReportArg.IMSI: %w", err)
		}
		v := gsm_map.IMSI(imsiBytes)
		out.Imsi = &v
	}
	// [2] imei
	if a.IMEI != "" {
		if len(a.IMEI) != pslIMEIDigits {
			return nil, fmt.Errorf("SubscriberLocationReportArg.IMEI digits=%d: %w", len(a.IMEI), ErrSLRArgIMEIInvalidSize)
		}
		imeiBytes, err := tbcd.Encode(a.IMEI)
		if err != nil {
			return nil, fmt.Errorf("encoding SubscriberLocationReportArg.IMEI: %w", err)
		}
		v := gsm_map.IMEI(imeiBytes)
		out.Imei = &v
	}
	// [3] na-ESRD
	if a.NaESRD != "" {
		isdn, err := encodeAddressField(a.NaESRD, a.NaESRDNature, a.NaESRDPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SubscriberLocationReportArg.NaESRD: %w", err)
		}
		v := gsm_map.ISDNAddressString(isdn)
		out.NaESRD = &v
	}
	// [4] na-ESRK
	if a.NaESRK != "" {
		isdn, err := encodeAddressField(a.NaESRK, a.NaESRKNature, a.NaESRKPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SubscriberLocationReportArg.NaESRK: %w", err)
		}
		v := gsm_map.ISDNAddressString(isdn)
		out.NaESRK = &v
	}
	// [5] locationEstimate
	if len(a.LocationEstimate) > 0 {
		if len(a.LocationEstimate) < ExtGeographicalInformationMinLen || len(a.LocationEstimate) > ExtGeographicalInformationMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.LocationEstimate len=%d: %w", len(a.LocationEstimate), ErrExtGeographicalInformationSize)
		}
		v := gsm_map.ExtGeographicalInformation(a.LocationEstimate)
		out.LocationEstimate = &v
	}
	// [6] ageOfLocationEstimate
	if a.AgeOfLocationEstimate != nil {
		v := gsm_map.AgeOfLocationInformation(*a.AgeOfLocationEstimate)
		out.AgeOfLocationEstimate = &v
	}
	// [7] slr-ArgExtensionContainer: opaque metadata; not surfaced.
	// [8] add-LocationEstimate
	if len(a.AddLocationEstimate) > 0 {
		if len(a.AddLocationEstimate) < AddGeographicalInformationMinLen || len(a.AddLocationEstimate) > AddGeographicalInformationMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.AddLocationEstimate len=%d: %w", len(a.AddLocationEstimate), ErrAddGeographicalInformationSize)
		}
		v := gsm_map.AddGeographicalInformation(a.AddLocationEstimate)
		out.AddLocationEstimate = &v
	}
	// [9] deferredmt-lrData
	if a.DeferredmtLrData != nil {
		v, err := convertDeferredmtLrDataToWire(a.DeferredmtLrData)
		if err != nil {
			return nil, fmt.Errorf("SubscriberLocationReportArg.DeferredmtLrData: %w", err)
		}
		out.DeferredmtLrData = v
	}
	// [10] lcs-ReferenceNumber (OCTET STRING SIZE 1)
	if len(a.LcsReferenceNumber) > 0 {
		if len(a.LcsReferenceNumber) != 1 {
			return nil, fmt.Errorf("SubscriberLocationReportArg.LcsReferenceNumber len=%d: %w", len(a.LcsReferenceNumber), ErrLCSReferenceNumberInvalidSize)
		}
		v := gsm_map.LCSReferenceNumber(a.LcsReferenceNumber)
		out.LcsReferenceNumber = &v
	}
	// [11] geranPositioningData
	if len(a.GeranPositioningData) > 0 {
		if len(a.GeranPositioningData) < PositioningDataInformationMinLen || len(a.GeranPositioningData) > PositioningDataInformationMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.GeranPositioningData len=%d: %w", len(a.GeranPositioningData), ErrPositioningDataInformationSize)
		}
		v := gsm_map.PositioningDataInformation(a.GeranPositioningData)
		out.GeranPositioningData = &v
	}
	// [12] utranPositioningData
	if len(a.UtranPositioningData) > 0 {
		if len(a.UtranPositioningData) < UtranPositioningDataInfoMinLen || len(a.UtranPositioningData) > UtranPositioningDataInfoMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.UtranPositioningData len=%d: %w", len(a.UtranPositioningData), ErrUtranPositioningDataInfoSize)
		}
		v := gsm_map.UtranPositioningDataInfo(a.UtranPositioningData)
		out.UtranPositioningData = &v
	}
	// [13] cellIdOrSai (CHOICE CGI/SAI vs LAI)
	if len(a.CellGlobalId) > 0 && len(a.LAI) > 0 {
		return nil, fmt.Errorf("SubscriberLocationReportArg.CellIdOrSai: %w", ErrSLRArgCellGlobalIdAndLAIMutex)
	}
	cellChoice, err := convertCellIdOrSaiToWire(a.CellGlobalId, a.LAI)
	if err != nil {
		return nil, fmt.Errorf("SubscriberLocationReportArg.CellIdOrSai: %w", err)
	}
	out.CellIdOrSai = cellChoice
	// [14] h-gmlc-Address
	if a.HGmlcAddress != "" {
		gsnAddr, err := gsn.Build(a.HGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("encoding SubscriberLocationReportArg.HGmlcAddress: %w", err)
		}
		v := gsm_map.GSNAddress(gsnAddr)
		out.HGmlcAddress = &v
	}
	// [15] lcsServiceTypeID
	if a.LcsServiceTypeID != nil {
		v := *a.LcsServiceTypeID
		if v < pslLcsServiceTypeIDMin || v > pslLcsServiceTypeIDMax {
			return nil, fmt.Errorf("SubscriberLocationReportArg.LcsServiceTypeID=%d: %w", v, ErrSLRArgLcsServiceTypeIDOutOfRange)
		}
		w := gsm_map.LCSServiceTypeID(v)
		out.LcsServiceTypeID = &w
	}
	// [17] sai-Present / [18] pseudonymIndicator (NULL flags)
	out.SaiPresent = boolToNullPtr(a.SaiPresent)
	out.PseudonymIndicator = boolToNullPtr(a.PseudonymIndicator)
	// [19] accuracyFulfilmentIndicator (extensible enum; encoder strict 0..1)
	if a.AccuracyFulfilmentIndicator != nil {
		v := *a.AccuracyFulfilmentIndicator
		if int64(v) < 0 || int64(v) > 1 {
			return nil, fmt.Errorf("SubscriberLocationReportArg.AccuracyFulfilmentIndicator=%d: %w", v, ErrAccuracyFulfilmentIndicatorInvalid)
		}
		out.AccuracyFulfilmentIndicator = &v
	}
	// [20] velocityEstimate
	if len(a.VelocityEstimate) > 0 {
		if len(a.VelocityEstimate) < VelocityEstimateMinLen || len(a.VelocityEstimate) > VelocityEstimateMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.VelocityEstimate len=%d: %w", len(a.VelocityEstimate), ErrVelocityEstimateSize)
		}
		v := gsm_map.VelocityEstimate(a.VelocityEstimate)
		out.VelocityEstimate = &v
	}
	// [21] sequenceNumber
	if a.SequenceNumber != nil {
		v := *a.SequenceNumber
		if v < SequenceNumberMin || v > SequenceNumberMax {
			return nil, fmt.Errorf("SubscriberLocationReportArg.SequenceNumber=%d: %w", v, ErrSequenceNumberOutOfRange)
		}
		out.SequenceNumber = &v
	}
	// [22] periodicLDRInfo
	if a.PeriodicLDRInfo != nil {
		v, err := convertPeriodicLDRInfoToWire(a.PeriodicLDRInfo)
		if err != nil {
			return nil, fmt.Errorf("SubscriberLocationReportArg.PeriodicLDRInfo: %w", err)
		}
		out.PeriodicLDRInfo = v
	}
	// [23] mo-lrShortCircuitIndicator (NULL flag)
	out.MoLrShortCircuitIndicator = boolToNullPtr(a.MoLrShortCircuitIndicator)
	// [24] geranGANSSpositioningData
	if len(a.GeranGANSSpositioningData) > 0 {
		if len(a.GeranGANSSpositioningData) < GeranGANSSpositioningDataMinLen || len(a.GeranGANSSpositioningData) > GeranGANSSpositioningDataMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.GeranGANSSpositioningData len=%d: %w", len(a.GeranGANSSpositioningData), ErrGeranGANSSpositioningDataSize)
		}
		v := gsm_map.GeranGANSSpositioningData(a.GeranGANSSpositioningData)
		out.GeranGANSSpositioningData = &v
	}
	// [25] utranGANSSpositioningData
	if len(a.UtranGANSSpositioningData) > 0 {
		if len(a.UtranGANSSpositioningData) < UtranGANSSpositioningDataMinLen || len(a.UtranGANSSpositioningData) > UtranGANSSpositioningDataMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.UtranGANSSpositioningData len=%d: %w", len(a.UtranGANSSpositioningData), ErrUtranGANSSpositioningDataSize)
		}
		v := gsm_map.UtranGANSSpositioningData(a.UtranGANSSpositioningData)
		out.UtranGANSSpositioningData = &v
	}
	// [26] targetServingNodeForHandover
	if a.TargetServingNodeForHandover != nil {
		v, err := convertServingNodeAddressToWire(a.TargetServingNodeForHandover)
		if err != nil {
			return nil, fmt.Errorf("SubscriberLocationReportArg.TargetServingNodeForHandover: %w", err)
		}
		out.TargetServingNodeForHandover = v
	}
	// [27] utranAdditionalPositioningData
	if len(a.UtranAdditionalPositioningData) > 0 {
		if len(a.UtranAdditionalPositioningData) < UtranAdditionalPositioningDataMinLen || len(a.UtranAdditionalPositioningData) > UtranAdditionalPositioningDataMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.UtranAdditionalPositioningData len=%d: %w", len(a.UtranAdditionalPositioningData), ErrUtranAdditionalPositioningDataSize)
		}
		v := gsm_map.UtranAdditionalPositioningData(a.UtranAdditionalPositioningData)
		out.UtranAdditionalPositioningData = &v
	}
	// [28] utranBaroPressureMeas
	if a.UtranBaroPressureMeas != nil {
		v := *a.UtranBaroPressureMeas
		if v < UtranBaroPressureMeasMin || v > UtranBaroPressureMeasMax {
			return nil, fmt.Errorf("SubscriberLocationReportArg.UtranBaroPressureMeas=%d: %w", v, ErrUtranBaroPressureMeasOutOfRange)
		}
		out.UtranBaroPressureMeas = &v
	}
	// [29] utranCivicAddress
	if len(a.UtranCivicAddress) > 0 {
		v := gsm_map.UtranCivicAddress(a.UtranCivicAddress)
		out.UtranCivicAddress = &v
	}

	return out, nil
}

// convertWireToSubscriberLocationReportArg unmarshals the wire-form
// struct back to the public type. Validation rules mirror the encoder
// and PSL's:
//   - Fixed-domain identifiers (IMSI/IMEI digit counts, LcsReferenceNumber
//     byte size, LcsServiceTypeID/SequenceNumber/UtranBaroPressureMeas
//     ranges, positioning/estimate byte sizes): rejected when out of
//     range, symmetric with the encoder.
//   - Round-trip safety: present-but-empty MSISDN/IMSI/IMEI/NaESRD/NaESRK
//     decoded values are rejected (cannot round-trip through the
//     string-based public API).
//   - Extensible enums (LcsEvent, AccuracyFulfilmentIndicator, and those
//     inside LcsClientID/LcsLocationInfo leaves): unknown values
//     preserved per Postel; encoder-side strictness lives here and in the
//     leaf converters.
//   - slr-ArgExtensionContainer at tag [7]: dropped (opaque metadata not
//     surfaced; see SubscriberLocationReportArg doc).
func convertWireToSubscriberLocationReportArg(w *gsm_map.SubscriberLocationReportArg) (*SubscriberLocationReportArg, error) {
	if w == nil {
		return nil, ErrSLRArgNil
	}

	clientID, err := convertWireToLCSClientID(&w.LcsClientID)
	if err != nil {
		return nil, fmt.Errorf("SubscriberLocationReportArg.LcsClientID: %w", err)
	}
	locInfo, err := convertWireToLCSLocationInfo(&w.LcsLocationInfo)
	if err != nil {
		return nil, fmt.Errorf("SubscriberLocationReportArg.LcsLocationInfo: %w", err)
	}

	out := &SubscriberLocationReportArg{
		LcsEvent:        w.LcsEvent,
		LcsClientID:     *clientID,
		LcsLocationInfo: *locInfo,
	}

	if w.Msisdn != nil {
		s, nature, plan, err := decodeAddressField([]byte(*w.Msisdn))
		if err != nil {
			return nil, fmt.Errorf("decoding SubscriberLocationReportArg.MSISDN: %w", err)
		}
		if s == "" {
			return nil, ErrSLRArgMSISDNDecodedEmpty
		}
		out.MSISDN = s
		out.MSISDNNature = nature
		out.MSISDNPlan = plan
	}
	if w.Imsi != nil {
		imsi, err := tbcd.Decode(*w.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding SubscriberLocationReportArg.IMSI: %w", err)
		}
		if imsi == "" {
			return nil, ErrSLRArgIMSIDecodedEmpty
		}
		if len(imsi) < pslIMSIDigitsMin || len(imsi) > pslIMSIDigitsMax {
			return nil, fmt.Errorf("SubscriberLocationReportArg.IMSI digits=%d: %w", len(imsi), ErrSLRArgIMSIInvalidSize)
		}
		out.IMSI = imsi
	}
	if w.Imei != nil {
		imei, err := tbcd.Decode(*w.Imei)
		if err != nil {
			return nil, fmt.Errorf("decoding SubscriberLocationReportArg.IMEI: %w", err)
		}
		if imei == "" {
			return nil, ErrSLRArgIMEIDecodedEmpty
		}
		if len(imei) != pslIMEIDigits {
			return nil, fmt.Errorf("SubscriberLocationReportArg.IMEI digits=%d: %w", len(imei), ErrSLRArgIMEIInvalidSize)
		}
		out.IMEI = imei
	}
	if w.NaESRD != nil {
		s, nature, plan, err := decodeAddressField([]byte(*w.NaESRD))
		if err != nil {
			return nil, fmt.Errorf("decoding SubscriberLocationReportArg.NaESRD: %w", err)
		}
		if s == "" {
			return nil, ErrSLRArgNaESRDDecodedEmpty
		}
		out.NaESRD = s
		out.NaESRDNature = nature
		out.NaESRDPlan = plan
	}
	if w.NaESRK != nil {
		s, nature, plan, err := decodeAddressField([]byte(*w.NaESRK))
		if err != nil {
			return nil, fmt.Errorf("decoding SubscriberLocationReportArg.NaESRK: %w", err)
		}
		if s == "" {
			return nil, ErrSLRArgNaESRKDecodedEmpty
		}
		out.NaESRK = s
		out.NaESRKNature = nature
		out.NaESRKPlan = plan
	}
	if w.LocationEstimate != nil {
		if len(*w.LocationEstimate) < ExtGeographicalInformationMinLen || len(*w.LocationEstimate) > ExtGeographicalInformationMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.LocationEstimate len=%d: %w", len(*w.LocationEstimate), ErrExtGeographicalInformationSize)
		}
		out.LocationEstimate = ExtGeographicalInformation(*w.LocationEstimate)
	}
	if w.AgeOfLocationEstimate != nil {
		v := int64(*w.AgeOfLocationEstimate)
		out.AgeOfLocationEstimate = &v
	}
	if w.AddLocationEstimate != nil {
		if len(*w.AddLocationEstimate) < AddGeographicalInformationMinLen || len(*w.AddLocationEstimate) > AddGeographicalInformationMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.AddLocationEstimate len=%d: %w", len(*w.AddLocationEstimate), ErrAddGeographicalInformationSize)
		}
		out.AddLocationEstimate = AddGeographicalInformation(*w.AddLocationEstimate)
	}
	if w.DeferredmtLrData != nil {
		v, err := convertWireToDeferredmtLrData(w.DeferredmtLrData)
		if err != nil {
			return nil, fmt.Errorf("SubscriberLocationReportArg.DeferredmtLrData: %w", err)
		}
		out.DeferredmtLrData = v
	}
	if w.LcsReferenceNumber != nil {
		if len(*w.LcsReferenceNumber) != 1 {
			return nil, fmt.Errorf("SubscriberLocationReportArg.LcsReferenceNumber len=%d: %w", len(*w.LcsReferenceNumber), ErrLCSReferenceNumberInvalidSize)
		}
		out.LcsReferenceNumber = LCSReferenceNumber(*w.LcsReferenceNumber)
	}
	if w.GeranPositioningData != nil {
		if len(*w.GeranPositioningData) < PositioningDataInformationMinLen || len(*w.GeranPositioningData) > PositioningDataInformationMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.GeranPositioningData len=%d: %w", len(*w.GeranPositioningData), ErrPositioningDataInformationSize)
		}
		out.GeranPositioningData = PositioningDataInformation(*w.GeranPositioningData)
	}
	if w.UtranPositioningData != nil {
		if len(*w.UtranPositioningData) < UtranPositioningDataInfoMinLen || len(*w.UtranPositioningData) > UtranPositioningDataInfoMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.UtranPositioningData len=%d: %w", len(*w.UtranPositioningData), ErrUtranPositioningDataInfoSize)
		}
		out.UtranPositioningData = UtranPositioningDataInfo(*w.UtranPositioningData)
	}
	cgi, lai, err := convertWireToCellIdOrSai(w.CellIdOrSai)
	if err != nil {
		return nil, fmt.Errorf("SubscriberLocationReportArg.CellIdOrSai: %w", err)
	}
	out.CellGlobalId = cgi
	out.LAI = lai

	if w.HGmlcAddress != nil {
		addr, err := gsn.Parse(*w.HGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("decoding SubscriberLocationReportArg.HGmlcAddress: %w", err)
		}
		out.HGmlcAddress = addr
	}
	if w.LcsServiceTypeID != nil {
		v := int64(*w.LcsServiceTypeID)
		if v < pslLcsServiceTypeIDMin || v > pslLcsServiceTypeIDMax {
			return nil, fmt.Errorf("SubscriberLocationReportArg.LcsServiceTypeID=%d: %w", v, ErrSLRArgLcsServiceTypeIDOutOfRange)
		}
		out.LcsServiceTypeID = &v
	}
	out.SaiPresent = nullPtrToBool(w.SaiPresent)
	out.PseudonymIndicator = nullPtrToBool(w.PseudonymIndicator)
	if w.AccuracyFulfilmentIndicator != nil {
		v := *w.AccuracyFulfilmentIndicator
		out.AccuracyFulfilmentIndicator = &v
	}
	if w.VelocityEstimate != nil {
		if len(*w.VelocityEstimate) < VelocityEstimateMinLen || len(*w.VelocityEstimate) > VelocityEstimateMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.VelocityEstimate len=%d: %w", len(*w.VelocityEstimate), ErrVelocityEstimateSize)
		}
		out.VelocityEstimate = VelocityEstimate(*w.VelocityEstimate)
	}
	if w.SequenceNumber != nil {
		v := *w.SequenceNumber
		if v < SequenceNumberMin || v > SequenceNumberMax {
			return nil, fmt.Errorf("SubscriberLocationReportArg.SequenceNumber=%d: %w", v, ErrSequenceNumberOutOfRange)
		}
		out.SequenceNumber = &v
	}
	if w.PeriodicLDRInfo != nil {
		v, err := convertWireToPeriodicLDRInfo(w.PeriodicLDRInfo)
		if err != nil {
			return nil, fmt.Errorf("SubscriberLocationReportArg.PeriodicLDRInfo: %w", err)
		}
		out.PeriodicLDRInfo = v
	}
	out.MoLrShortCircuitIndicator = nullPtrToBool(w.MoLrShortCircuitIndicator)
	if w.GeranGANSSpositioningData != nil {
		if len(*w.GeranGANSSpositioningData) < GeranGANSSpositioningDataMinLen || len(*w.GeranGANSSpositioningData) > GeranGANSSpositioningDataMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.GeranGANSSpositioningData len=%d: %w", len(*w.GeranGANSSpositioningData), ErrGeranGANSSpositioningDataSize)
		}
		out.GeranGANSSpositioningData = GeranGANSSpositioningData(*w.GeranGANSSpositioningData)
	}
	if w.UtranGANSSpositioningData != nil {
		if len(*w.UtranGANSSpositioningData) < UtranGANSSpositioningDataMinLen || len(*w.UtranGANSSpositioningData) > UtranGANSSpositioningDataMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.UtranGANSSpositioningData len=%d: %w", len(*w.UtranGANSSpositioningData), ErrUtranGANSSpositioningDataSize)
		}
		out.UtranGANSSpositioningData = UtranGANSSpositioningData(*w.UtranGANSSpositioningData)
	}
	if w.TargetServingNodeForHandover != nil {
		v, err := convertWireToServingNodeAddress(w.TargetServingNodeForHandover)
		if err != nil {
			return nil, fmt.Errorf("SubscriberLocationReportArg.TargetServingNodeForHandover: %w", err)
		}
		out.TargetServingNodeForHandover = v
	}
	if w.UtranAdditionalPositioningData != nil {
		if len(*w.UtranAdditionalPositioningData) < UtranAdditionalPositioningDataMinLen || len(*w.UtranAdditionalPositioningData) > UtranAdditionalPositioningDataMaxLen {
			return nil, fmt.Errorf("SubscriberLocationReportArg.UtranAdditionalPositioningData len=%d: %w", len(*w.UtranAdditionalPositioningData), ErrUtranAdditionalPositioningDataSize)
		}
		out.UtranAdditionalPositioningData = UtranAdditionalPositioningData(*w.UtranAdditionalPositioningData)
	}
	if w.UtranBaroPressureMeas != nil {
		v := *w.UtranBaroPressureMeas
		if v < UtranBaroPressureMeasMin || v > UtranBaroPressureMeasMax {
			return nil, fmt.Errorf("SubscriberLocationReportArg.UtranBaroPressureMeas=%d: %w", v, ErrUtranBaroPressureMeasOutOfRange)
		}
		out.UtranBaroPressureMeas = &v
	}
	if w.UtranCivicAddress != nil {
		out.UtranCivicAddress = UtranCivicAddress(*w.UtranCivicAddress)
	}

	return out, nil
}
