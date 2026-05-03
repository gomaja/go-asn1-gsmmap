// convert_psl_res.go
//
// Top-level converter for ProvideSubscriberLocationRes (opCode 83) and
// the ServingNodeAddress CHOICE codec referenced by PSL-Res's
// targetServingNodeForHandover field. PR E of the staged PSL
// implementation: completes the ProvideSubscriberLocation operation
// after PRs #43/#44/#45/#46 (PSL-Arg).

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// CGI/SAI and LAI fixed-length sizes per TS 29.002 MAP-CommonDataTypes.asn.
const (
	pslResCellGlobalIdLen = 7
	pslResLAILen          = 5

	// DiameterIdentity SIZE 9..255 per TS 29.002 MAP-MS-DataTypes.asn:1434
	// (also used for DiameterName/Realm and ServingNodeAddress.MmeName).
	diameterIdentityMinLen = 9
	diameterIdentityMaxLen = 255
)

// ============================================================================
// ServingNodeAddress CHOICE codec
// ============================================================================
//
// ServingNodeAddress is a CHOICE between MscNumber, SgsnNumber, and
// MmeName per TS 29.002 MAP-LCS-DataTypes.asn (used in PSL-Res
// targetServingNodeForHandover field). Per the existing CHOICE pattern
// (see AdditionalNumber, CancelLocationIdentity), the selected
// alternative is inferred from which field is set:
//   - non-empty MscNumber digits → MscNumber alternative
//   - non-empty SgsnNumber digits → SgsnNumber alternative
//   - non-empty MmeName octets → MmeName alternative

func convertServingNodeAddressToWire(s *ServingNodeAddress) (*gsm_map.ServingNodeAddress, error) {
	if s == nil {
		return nil, nil
	}
	mscSet := s.MscNumber != ""
	sgsnSet := s.SgsnNumber != ""
	mmeSet := len(s.MmeName) > 0
	count := 0
	if mscSet {
		count++
	}
	if sgsnSet {
		count++
	}
	if mmeSet {
		count++
	}
	switch {
	case count == 0:
		return nil, ErrServingNodeAddressNoAlt
	case count > 1:
		return nil, ErrServingNodeAddressMultipleAlts
	}

	switch {
	case mscSet:
		isdn, err := encodeAddressField(s.MscNumber, s.MscNumberNature, s.MscNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding ServingNodeAddress.MscNumber: %w", err)
		}
		v := gsm_map.NewServingNodeAddressMscNumber(gsm_map.ISDNAddressString(isdn))
		return &v, nil
	case sgsnSet:
		isdn, err := encodeAddressField(s.SgsnNumber, s.SgsnNumberNature, s.SgsnNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding ServingNodeAddress.SgsnNumber: %w", err)
		}
		v := gsm_map.NewServingNodeAddressSgsnNumber(gsm_map.ISDNAddressString(isdn))
		return &v, nil
	default: // mmeSet
		if len(s.MmeName) < diameterIdentityMinLen || len(s.MmeName) > diameterIdentityMaxLen {
			return nil, fmt.Errorf("ServingNodeAddress.MmeName len=%d: %w", len(s.MmeName), ErrServingNodeAddressMmeNameSize)
		}
		v := gsm_map.NewServingNodeAddressMmeNumber(gsm_map.DiameterIdentity(s.MmeName))
		return &v, nil
	}
}

func convertWireToServingNodeAddress(w *gsm_map.ServingNodeAddress) (*ServingNodeAddress, error) {
	if w == nil {
		return nil, nil
	}
	out := &ServingNodeAddress{}
	switch w.Choice {
	case gsm_map.ServingNodeAddressChoiceMscNumber:
		if w.MscNumber == nil {
			return nil, ErrServingNodeAddressNoAlt
		}
		s, nature, plan, err := decodeAddressField([]byte(*w.MscNumber))
		if err != nil {
			return nil, fmt.Errorf("decoding ServingNodeAddress.MscNumber: %w", err)
		}
		if s == "" {
			return nil, fmt.Errorf("ServingNodeAddress.MscNumber: present wire ISDN-AddressString decoded to empty digits")
		}
		out.MscNumber = s
		out.MscNumberNature = nature
		out.MscNumberPlan = plan
	case gsm_map.ServingNodeAddressChoiceSgsnNumber:
		if w.SgsnNumber == nil {
			return nil, ErrServingNodeAddressNoAlt
		}
		s, nature, plan, err := decodeAddressField([]byte(*w.SgsnNumber))
		if err != nil {
			return nil, fmt.Errorf("decoding ServingNodeAddress.SgsnNumber: %w", err)
		}
		if s == "" {
			return nil, fmt.Errorf("ServingNodeAddress.SgsnNumber: present wire ISDN-AddressString decoded to empty digits")
		}
		out.SgsnNumber = s
		out.SgsnNumberNature = nature
		out.SgsnNumberPlan = plan
	case gsm_map.ServingNodeAddressChoiceMmeNumber:
		if w.MmeNumber == nil {
			return nil, ErrServingNodeAddressNoAlt
		}
		mme := HexBytes(*w.MmeNumber)
		if len(mme) < diameterIdentityMinLen || len(mme) > diameterIdentityMaxLen {
			return nil, fmt.Errorf("ServingNodeAddress.MmeName len=%d: %w", len(mme), ErrServingNodeAddressMmeNameSize)
		}
		out.MmeName = mme
	default:
		return nil, ErrServingNodeAddressNoAlt
	}
	return out, nil
}

// ============================================================================
// CellIdOrSai CHOICE codec (CGI/SAI 7 octets vs LAI 5 octets)
// ============================================================================

func convertCellIdOrSaiToWire(cgi, lai HexBytes) (*gsm_map.CellGlobalIdOrServiceAreaIdOrLAI, error) {
	cgiSet := len(cgi) > 0
	laiSet := len(lai) > 0
	if cgiSet && laiSet {
		return nil, ErrPSLResCellGlobalIdAndLAIMutex
	}
	if !cgiSet && !laiSet {
		return nil, nil
	}
	if cgiSet {
		if len(cgi) != pslResCellGlobalIdLen {
			return nil, fmt.Errorf("CellGlobalId len=%d: %w", len(cgi), ErrPSLResCellGlobalIdSize)
		}
		v := gsm_map.NewCellGlobalIdOrServiceAreaIdOrLAICellGlobalIdOrServiceAreaIdFixedLength(
			gsm_map.CellGlobalIdOrServiceAreaIdFixedLength(cgi),
		)
		return &v, nil
	}
	// laiSet
	if len(lai) != pslResLAILen {
		return nil, fmt.Errorf("LAI len=%d: %w", len(lai), ErrPSLResLAIInvalidSize)
	}
	v := gsm_map.NewCellGlobalIdOrServiceAreaIdOrLAILaiFixedLength(gsm_map.LAIFixedLength(lai))
	return &v, nil
}

func convertWireToCellIdOrSai(w *gsm_map.CellGlobalIdOrServiceAreaIdOrLAI) (cgi, lai HexBytes, err error) {
	if w == nil {
		return nil, nil, nil
	}
	switch w.Choice {
	case gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceCellGlobalIdOrServiceAreaIdFixedLength:
		if w.CellGlobalIdOrServiceAreaIdFixedLength == nil {
			return nil, nil, nil
		}
		b := HexBytes(*w.CellGlobalIdOrServiceAreaIdFixedLength)
		if len(b) != pslResCellGlobalIdLen {
			return nil, nil, fmt.Errorf("CellGlobalId len=%d: %w", len(b), ErrPSLResCellGlobalIdSize)
		}
		return b, nil, nil
	case gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceLaiFixedLength:
		if w.LaiFixedLength == nil {
			return nil, nil, nil
		}
		b := HexBytes(*w.LaiFixedLength)
		if len(b) != pslResLAILen {
			return nil, nil, fmt.Errorf("LAI len=%d: %w", len(b), ErrPSLResLAIInvalidSize)
		}
		return nil, b, nil
	default:
		return nil, nil, nil
	}
}

// ============================================================================
// ProvideSubscriberLocationRes top-level
// ============================================================================
//
// Decoder validation rules mirror PSL-Arg's:
//   - Fixed-domain identifiers (CGI/SAI/LAI byte sizes,
//     UtranBaroPressureMeas range, AccuracyFulfilmentIndicator range
//     when surfaced): rejected when out of range, symmetric with the
//     encoder.
//   - LocationEstimate is mandatory; empty wire payload is rejected.
//   - Extensible enum AccuracyFulfilmentIndicator: encoder-strict 0..1,
//     decoder-lenient (preserves unknown values per Postel).
//   - ExtensionContainer at tag [1]: dropped (opaque metadata not
//     surfaced; see ProvideSubscriberLocationRes doc).

func convertProvideSubscriberLocationResToWire(r *ProvideSubscriberLocationRes) (*gsm_map.ProvideSubscriberLocationRes, error) {
	if r == nil {
		return nil, ErrPSLResNil
	}
	if len(r.LocationEstimate) == 0 {
		return nil, ErrPSLResLocationEstimateMissing
	}
	if len(r.LocationEstimate) < ExtGeographicalInformationMinLen || len(r.LocationEstimate) > ExtGeographicalInformationMaxLen {
		return nil, fmt.Errorf("ProvideSubscriberLocationRes.LocationEstimate len=%d: %w", len(r.LocationEstimate), ErrExtGeographicalInformationSize)
	}

	out := &gsm_map.ProvideSubscriberLocationRes{
		LocationEstimate: gsm_map.ExtGeographicalInformation(r.LocationEstimate),
	}

	if r.AgeOfLocationEstimate != nil {
		v := gsm_map.AgeOfLocationInformation(*r.AgeOfLocationEstimate)
		out.AgeOfLocationEstimate = &v
	}
	if len(r.AddLocationEstimate) > 0 {
		if len(r.AddLocationEstimate) < AddGeographicalInformationMinLen || len(r.AddLocationEstimate) > AddGeographicalInformationMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.AddLocationEstimate len=%d: %w", len(r.AddLocationEstimate), ErrAddGeographicalInformationSize)
		}
		v := gsm_map.AddGeographicalInformation(r.AddLocationEstimate)
		out.AddLocationEstimate = &v
	}
	out.DeferredmtLrResponseIndicator = boolToNullPtr(r.DeferredmtLrResponseIndicator)

	if len(r.GeranPositioningData) > 0 {
		if len(r.GeranPositioningData) < PositioningDataInformationMinLen || len(r.GeranPositioningData) > PositioningDataInformationMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.GeranPositioningData len=%d: %w", len(r.GeranPositioningData), ErrPositioningDataInformationSize)
		}
		v := gsm_map.PositioningDataInformation(r.GeranPositioningData)
		out.GeranPositioningData = &v
	}
	if len(r.UtranPositioningData) > 0 {
		if len(r.UtranPositioningData) < UtranPositioningDataInfoMinLen || len(r.UtranPositioningData) > UtranPositioningDataInfoMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.UtranPositioningData len=%d: %w", len(r.UtranPositioningData), ErrUtranPositioningDataInfoSize)
		}
		v := gsm_map.UtranPositioningDataInfo(r.UtranPositioningData)
		out.UtranPositioningData = &v
	}

	cellChoice, err := convertCellIdOrSaiToWire(r.CellGlobalId, r.LAI)
	if err != nil {
		return nil, fmt.Errorf("ProvideSubscriberLocationRes.CellIdOrSai: %w", err)
	}
	out.CellIdOrSai = cellChoice

	out.SaiPresent = boolToNullPtr(r.SaiPresent)

	if r.AccuracyFulfilmentIndicator != nil {
		v := *r.AccuracyFulfilmentIndicator
		// AccuracyFulfilmentIndicator is extensible (TS 29.002:457);
		// encoder strict (0..1), decoder lenient.
		if int64(v) < 0 || int64(v) > 1 {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.AccuracyFulfilmentIndicator=%d: %w", v, ErrAccuracyFulfilmentIndicatorInvalid)
		}
		out.AccuracyFulfilmentIndicator = &v
	}
	if len(r.VelocityEstimate) > 0 {
		if len(r.VelocityEstimate) < VelocityEstimateMinLen || len(r.VelocityEstimate) > VelocityEstimateMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.VelocityEstimate len=%d: %w", len(r.VelocityEstimate), ErrVelocityEstimateSize)
		}
		v := gsm_map.VelocityEstimate(r.VelocityEstimate)
		out.VelocityEstimate = &v
	}
	out.MoLrShortCircuitIndicator = boolToNullPtr(r.MoLrShortCircuitIndicator)

	if len(r.GeranGANSSpositioningData) > 0 {
		if len(r.GeranGANSSpositioningData) < GeranGANSSpositioningDataMinLen || len(r.GeranGANSSpositioningData) > GeranGANSSpositioningDataMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.GeranGANSSpositioningData len=%d: %w", len(r.GeranGANSSpositioningData), ErrGeranGANSSpositioningDataSize)
		}
		v := gsm_map.GeranGANSSpositioningData(r.GeranGANSSpositioningData)
		out.GeranGANSSpositioningData = &v
	}
	if len(r.UtranGANSSpositioningData) > 0 {
		if len(r.UtranGANSSpositioningData) < UtranGANSSpositioningDataMinLen || len(r.UtranGANSSpositioningData) > UtranGANSSpositioningDataMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.UtranGANSSpositioningData len=%d: %w", len(r.UtranGANSSpositioningData), ErrUtranGANSSpositioningDataSize)
		}
		v := gsm_map.UtranGANSSpositioningData(r.UtranGANSSpositioningData)
		out.UtranGANSSpositioningData = &v
	}

	if r.TargetServingNodeForHandover != nil {
		v, err := convertServingNodeAddressToWire(r.TargetServingNodeForHandover)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.TargetServingNodeForHandover: %w", err)
		}
		out.TargetServingNodeForHandover = v
	}

	if len(r.UtranAdditionalPositioningData) > 0 {
		if len(r.UtranAdditionalPositioningData) < UtranAdditionalPositioningDataMinLen || len(r.UtranAdditionalPositioningData) > UtranAdditionalPositioningDataMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.UtranAdditionalPositioningData len=%d: %w", len(r.UtranAdditionalPositioningData), ErrUtranAdditionalPositioningDataSize)
		}
		v := gsm_map.UtranAdditionalPositioningData(r.UtranAdditionalPositioningData)
		out.UtranAdditionalPositioningData = &v
	}
	if r.UtranBaroPressureMeas != nil {
		v := *r.UtranBaroPressureMeas
		if v < UtranBaroPressureMeasMin || v > UtranBaroPressureMeasMax {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.UtranBaroPressureMeas=%d: %w", v, ErrUtranBaroPressureMeasOutOfRange)
		}
		out.UtranBaroPressureMeas = &v
	}
	if len(r.UtranCivicAddress) > 0 {
		v := gsm_map.UtranCivicAddress(r.UtranCivicAddress)
		out.UtranCivicAddress = &v
	}
	return out, nil
}

func convertWireToProvideSubscriberLocationRes(w *gsm_map.ProvideSubscriberLocationRes) (*ProvideSubscriberLocationRes, error) {
	if w == nil {
		return nil, ErrPSLResNil
	}
	if len(w.LocationEstimate) == 0 {
		return nil, ErrPSLResLocationEstimateMissing
	}
	if len(w.LocationEstimate) < ExtGeographicalInformationMinLen || len(w.LocationEstimate) > ExtGeographicalInformationMaxLen {
		return nil, fmt.Errorf("ProvideSubscriberLocationRes.LocationEstimate len=%d: %w", len(w.LocationEstimate), ErrExtGeographicalInformationSize)
	}

	out := &ProvideSubscriberLocationRes{
		LocationEstimate: ExtGeographicalInformation(w.LocationEstimate),
	}

	if w.AgeOfLocationEstimate != nil {
		v := int64(*w.AgeOfLocationEstimate)
		out.AgeOfLocationEstimate = &v
	}
	if w.AddLocationEstimate != nil {
		if len(*w.AddLocationEstimate) < AddGeographicalInformationMinLen || len(*w.AddLocationEstimate) > AddGeographicalInformationMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.AddLocationEstimate len=%d: %w", len(*w.AddLocationEstimate), ErrAddGeographicalInformationSize)
		}
		out.AddLocationEstimate = AddGeographicalInformation(*w.AddLocationEstimate)
	}
	out.DeferredmtLrResponseIndicator = nullPtrToBool(w.DeferredmtLrResponseIndicator)

	if w.GeranPositioningData != nil {
		if len(*w.GeranPositioningData) < PositioningDataInformationMinLen || len(*w.GeranPositioningData) > PositioningDataInformationMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.GeranPositioningData len=%d: %w", len(*w.GeranPositioningData), ErrPositioningDataInformationSize)
		}
		out.GeranPositioningData = PositioningDataInformation(*w.GeranPositioningData)
	}
	if w.UtranPositioningData != nil {
		if len(*w.UtranPositioningData) < UtranPositioningDataInfoMinLen || len(*w.UtranPositioningData) > UtranPositioningDataInfoMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.UtranPositioningData len=%d: %w", len(*w.UtranPositioningData), ErrUtranPositioningDataInfoSize)
		}
		out.UtranPositioningData = UtranPositioningDataInfo(*w.UtranPositioningData)
	}

	cgi, lai, err := convertWireToCellIdOrSai(w.CellIdOrSai)
	if err != nil {
		return nil, fmt.Errorf("ProvideSubscriberLocationRes.CellIdOrSai: %w", err)
	}
	out.CellGlobalId = cgi
	out.LAI = lai

	out.SaiPresent = nullPtrToBool(w.SaiPresent)

	if w.AccuracyFulfilmentIndicator != nil {
		v := *w.AccuracyFulfilmentIndicator
		out.AccuracyFulfilmentIndicator = &v
	}
	if w.VelocityEstimate != nil {
		if len(*w.VelocityEstimate) < VelocityEstimateMinLen || len(*w.VelocityEstimate) > VelocityEstimateMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.VelocityEstimate len=%d: %w", len(*w.VelocityEstimate), ErrVelocityEstimateSize)
		}
		out.VelocityEstimate = VelocityEstimate(*w.VelocityEstimate)
	}
	out.MoLrShortCircuitIndicator = nullPtrToBool(w.MoLrShortCircuitIndicator)

	if w.GeranGANSSpositioningData != nil {
		if len(*w.GeranGANSSpositioningData) < GeranGANSSpositioningDataMinLen || len(*w.GeranGANSSpositioningData) > GeranGANSSpositioningDataMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.GeranGANSSpositioningData len=%d: %w", len(*w.GeranGANSSpositioningData), ErrGeranGANSSpositioningDataSize)
		}
		out.GeranGANSSpositioningData = GeranGANSSpositioningData(*w.GeranGANSSpositioningData)
	}
	if w.UtranGANSSpositioningData != nil {
		if len(*w.UtranGANSSpositioningData) < UtranGANSSpositioningDataMinLen || len(*w.UtranGANSSpositioningData) > UtranGANSSpositioningDataMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.UtranGANSSpositioningData len=%d: %w", len(*w.UtranGANSSpositioningData), ErrUtranGANSSpositioningDataSize)
		}
		out.UtranGANSSpositioningData = UtranGANSSpositioningData(*w.UtranGANSSpositioningData)
	}

	if w.TargetServingNodeForHandover != nil {
		v, err := convertWireToServingNodeAddress(w.TargetServingNodeForHandover)
		if err != nil {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.TargetServingNodeForHandover: %w", err)
		}
		out.TargetServingNodeForHandover = v
	}

	if w.UtranAdditionalPositioningData != nil {
		if len(*w.UtranAdditionalPositioningData) < UtranAdditionalPositioningDataMinLen || len(*w.UtranAdditionalPositioningData) > UtranAdditionalPositioningDataMaxLen {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.UtranAdditionalPositioningData len=%d: %w", len(*w.UtranAdditionalPositioningData), ErrUtranAdditionalPositioningDataSize)
		}
		out.UtranAdditionalPositioningData = UtranAdditionalPositioningData(*w.UtranAdditionalPositioningData)
	}
	if w.UtranBaroPressureMeas != nil {
		v := *w.UtranBaroPressureMeas
		if v < UtranBaroPressureMeasMin || v > UtranBaroPressureMeasMax {
			return nil, fmt.Errorf("ProvideSubscriberLocationRes.UtranBaroPressureMeas=%d: %w", v, ErrUtranBaroPressureMeasOutOfRange)
		}
		out.UtranBaroPressureMeas = &v
	}
	if w.UtranCivicAddress != nil {
		out.UtranCivicAddress = UtranCivicAddress(*w.UtranCivicAddress)
	}
	return out, nil
}
