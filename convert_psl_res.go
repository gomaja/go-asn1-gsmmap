// convert_psl_res.go
//
// Top-level converter for ProvideSubscriberLocationRes (opCode 83) and
// the ServingNodeAddress CHOICE codec referenced by PSL-Res's
// targetServingNodeForHandover field.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// ServingNodeAddress CHOICE codec
// ============================================================================
//
// ServingNodeAddress is a CHOICE between MscNumber, SgsnNumber, and
// MmeNumber per TS 29.002 MAP-LCS-DataTypes.asn (used in PSL-Res
// targetServingNodeForHandover field). Per the existing CHOICE pattern
// (see AdditionalNumber, CancelLocationIdentity), the selected
// alternative is inferred from which field is set:
//   - non-empty MscNumber digits → MscNumber alternative
//   - non-empty SgsnNumber digits → SgsnNumber alternative
//   - non-empty MmeNumber octets → MmeNumber alternative

func convertServingNodeAddressToWire(s *ServingNodeAddress) (*gsm_map.ServingNodeAddress, error) {
	if s == nil {
		return nil, nil
	}
	mscSet := s.MscNumber != ""
	sgsnSet := s.SgsnNumber != ""
	mmeSet := len(s.MmeNumber) > 0
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
		v := gsm_map.NewServingNodeAddressMscNumber(isdn)
		return &v, nil
	case sgsnSet:
		isdn, err := encodeAddressField(s.SgsnNumber, s.SgsnNumberNature, s.SgsnNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding ServingNodeAddress.SgsnNumber: %w", err)
		}
		v := gsm_map.NewServingNodeAddressSgsnNumber(isdn)
		return &v, nil
	default: // mmeSet

		v := gsm_map.NewServingNodeAddressMmeNumber(gsm_map.DiameterIdentity(s.MmeNumber))
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
		s, nature, plan, err := decodeAddressField(*w.MscNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding ServingNodeAddress.MscNumber: %w", err)
		}
		if s == "" {
			return nil, ErrServingNodeAddressMscNumberDecodedEmpty
		}
		out.MscNumber = s
		out.MscNumberNature = nature
		out.MscNumberPlan = plan
	case gsm_map.ServingNodeAddressChoiceSgsnNumber:
		if w.SgsnNumber == nil {
			return nil, ErrServingNodeAddressNoAlt
		}
		s, nature, plan, err := decodeAddressField(*w.SgsnNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding ServingNodeAddress.SgsnNumber: %w", err)
		}
		if s == "" {
			return nil, ErrServingNodeAddressSgsnNumberDecodedEmpty
		}
		out.SgsnNumber = s
		out.SgsnNumberNature = nature
		out.SgsnNumberPlan = plan
	case gsm_map.ServingNodeAddressChoiceMmeNumber:
		if w.MmeNumber == nil {
			return nil, ErrServingNodeAddressNoAlt
		}
		mme := HexBytes(*w.MmeNumber)

		out.MmeNumber = mme
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
		v := gsm_map.NewCellGlobalIdOrServiceAreaIdOrLAICellGlobalIdOrServiceAreaIdFixedLength(
			gsm_map.CellGlobalIdOrServiceAreaIdFixedLength(cgi),
		)
		return &v, nil
	}
	// laiSet

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
			return nil, nil, fmt.Errorf("CellIdOrSai: choice=CGI/SAI but payload is nil: %w", ErrPSLResCellIdOrSaiInvalidChoice)
		}
		b := HexBytes(*w.CellGlobalIdOrServiceAreaIdFixedLength)

		return b, nil, nil
	case gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceLaiFixedLength:
		if w.LaiFixedLength == nil {
			return nil, nil, fmt.Errorf("CellIdOrSai: choice=LAI but payload is nil: %w", ErrPSLResCellIdOrSaiInvalidChoice)
		}
		b := HexBytes(*w.LaiFixedLength)

		return nil, b, nil
	default:
		return nil, nil, fmt.Errorf("CellIdOrSai: unsupported choice %d: %w", w.Choice, ErrPSLResCellIdOrSaiInvalidChoice)
	}
}

// ============================================================================
// ProvideSubscriberLocationRes top-level
// ============================================================================
//
// Decoder behavior:
//   - Extensible enum AccuracyFulfilmentIndicator: encoder-strict 0..1,
//     decoder-lenient (preserves unknown values per Postel).
//   - ExtensionContainer at tag [1]: dropped (opaque metadata not
//     surfaced; see ProvideSubscriberLocationRes doc).

func convertProvideSubscriberLocationResToWire(r *ProvideSubscriberLocationRes) (*gsm_map.ProvideSubscriberLocationRes, error) {
	if r == nil {
		return nil, ErrPSLResNil
	}

	out := &gsm_map.ProvideSubscriberLocationRes{
		LocationEstimate: gsm_map.ExtGeographicalInformation(r.LocationEstimate),
	}

	if r.AgeOfLocationEstimate != nil {
		v := *r.AgeOfLocationEstimate
		out.AgeOfLocationEstimate = &v
	}
	if len(r.AddLocationEstimate) > 0 {
		v := gsm_map.AddGeographicalInformation(r.AddLocationEstimate)
		out.AddLocationEstimate = &v
	}
	out.DeferredmtLrResponseIndicator = boolToNullPtr(r.DeferredmtLrResponseIndicator)

	if len(r.GeranPositioningData) > 0 {
		v := gsm_map.PositioningDataInformation(r.GeranPositioningData)
		out.GeranPositioningData = &v
	}
	if len(r.UtranPositioningData) > 0 {
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
		v := gsm_map.VelocityEstimate(r.VelocityEstimate)
		out.VelocityEstimate = &v
	}
	out.MoLrShortCircuitIndicator = boolToNullPtr(r.MoLrShortCircuitIndicator)

	if len(r.GeranGANSSpositioningData) > 0 {
		v := gsm_map.GeranGANSSpositioningData(r.GeranGANSSpositioningData)
		out.GeranGANSSpositioningData = &v
	}
	if len(r.UtranGANSSpositioningData) > 0 {
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
		v := gsm_map.UtranAdditionalPositioningData(r.UtranAdditionalPositioningData)
		out.UtranAdditionalPositioningData = &v
	}
	if r.UtranBaroPressureMeas != nil {
		v := *r.UtranBaroPressureMeas

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

	out := &ProvideSubscriberLocationRes{
		LocationEstimate: ExtGeographicalInformation(w.LocationEstimate),
	}

	if w.AgeOfLocationEstimate != nil {
		v := *w.AgeOfLocationEstimate
		out.AgeOfLocationEstimate = &v
	}
	if w.AddLocationEstimate != nil {
		out.AddLocationEstimate = AddGeographicalInformation(*w.AddLocationEstimate)
	}
	out.DeferredmtLrResponseIndicator = nullPtrToBool(w.DeferredmtLrResponseIndicator)

	if w.GeranPositioningData != nil {
		out.GeranPositioningData = PositioningDataInformation(*w.GeranPositioningData)
	}
	if w.UtranPositioningData != nil {
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
		out.VelocityEstimate = VelocityEstimate(*w.VelocityEstimate)
	}
	out.MoLrShortCircuitIndicator = nullPtrToBool(w.MoLrShortCircuitIndicator)

	if w.GeranGANSSpositioningData != nil {
		out.GeranGANSSpositioningData = GeranGANSSpositioningData(*w.GeranGANSSpositioningData)
	}
	if w.UtranGANSSpositioningData != nil {
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
		out.UtranAdditionalPositioningData = UtranAdditionalPositioningData(*w.UtranAdditionalPositioningData)
	}
	if w.UtranBaroPressureMeas != nil {
		v := *w.UtranBaroPressureMeas

		out.UtranBaroPressureMeas = &v
	}
	if w.UtranCivicAddress != nil {
		out.UtranCivicAddress = UtranCivicAddress(*w.UtranCivicAddress)
	}
	return out, nil
}
