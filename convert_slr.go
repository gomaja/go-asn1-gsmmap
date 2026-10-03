// convert_slr.go
//
// Converters for SubscriberLocationReport (opCode 86) sub-types:
// LCSLocationInfo and DeferredmtLrData. The top-level
// SubscriberLocationReportArg/Res converters are in convert_slr_arg.go and
// convert_slr_res.go.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// LCSLocationInfo — TS 29.002 MAP-LCS-DataTypes.asn
// ============================================================================

func convertLCSLocationInfoToWire(l *LCSLocationInfo) (*gsm_map.LCSLocationInfo, error) {
	if l == nil {
		return nil, nil
	}
	if l.NetworkNodeNumber == "" {
		return nil, ErrLCSLocationInfoNetworkNodeEmpty
	}
	nodeWire, err := encodeAddressField(l.NetworkNodeNumber, l.NetworkNodeNumberNature, l.NetworkNodeNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding LCSLocationInfo.NetworkNodeNumber: %w", err)
	}

	out := &gsm_map.LCSLocationInfo{
		NetworkNodeNumber: nodeWire,
	}

	if len(l.LMSI) > 0 {
		v := gsm_map.LMSI(l.LMSI)
		out.Lmsi = &v
	}
	out.GprsNodeIndicator = boolToNullPtr(l.GprsNodeIndicator)

	if l.AdditionalNumber != nil {
		an, err := convertAdditionalNumberToWire(l.AdditionalNumber)
		if err != nil {
			return nil, fmt.Errorf("LCSLocationInfo.AdditionalNumber: %w", err)
		}
		out.AdditionalNumber = an
	}
	if l.SupportedLCSCapabilitySets != nil {
		bs := convertLCSCapsToBitString(l.SupportedLCSCapabilitySets)
		out.SupportedLCSCapabilitySets = &bs
	}
	if l.AdditionalLCSCapabilitySets != nil {
		bs := convertLCSCapsToBitString(l.AdditionalLCSCapabilitySets)
		out.AdditionalLCSCapabilitySets = &bs
	}
	if len(l.MmeName) > 0 {
		v := gsm_map.DiameterIdentity(l.MmeName)
		out.MmeName = &v
	}
	if len(l.AaaServerName) > 0 {
		v := gsm_map.DiameterIdentity(l.AaaServerName)
		out.AaaServerName = &v
	}
	if len(l.SgsnName) > 0 {
		v := gsm_map.DiameterIdentity(l.SgsnName)
		out.SgsnName = &v
	}
	if len(l.SgsnRealm) > 0 {
		v := gsm_map.DiameterIdentity(l.SgsnRealm)
		out.SgsnRealm = &v
	}
	return out, nil
}

func convertWireToLCSLocationInfo(w *gsm_map.LCSLocationInfo) (*LCSLocationInfo, error) {
	if w == nil {
		return nil, nil
	}
	node, nature, plan, err := decodeAddressField(w.NetworkNodeNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding LCSLocationInfo.NetworkNodeNumber: %w", err)
	}
	if node == "" {
		return nil, ErrLCSLocationInfoNetworkNodeDecodedEmpty
	}

	out := &LCSLocationInfo{
		NetworkNodeNumber:       node,
		NetworkNodeNumberNature: nature,
		NetworkNodeNumberPlan:   plan,
	}

	if w.Lmsi != nil {
		out.LMSI = HexBytes(*w.Lmsi)
	}
	out.GprsNodeIndicator = nullPtrToBool(w.GprsNodeIndicator)

	if w.AdditionalNumber != nil {
		an, err := convertWireToAdditionalNumber(w.AdditionalNumber)
		if err != nil {
			return nil, fmt.Errorf("LCSLocationInfo.AdditionalNumber: %w", err)
		}
		out.AdditionalNumber = an
	}
	// Direct wire conversion treats zero-length capability sets as absent;
	// BER enforces SIZE (2..16) (3GPP TS 29.002 V19.1.0 §17.7.1).
	if w.SupportedLCSCapabilitySets != nil && w.SupportedLCSCapabilitySets.BitLength > 0 {
		out.SupportedLCSCapabilitySets = convertBitStringToLCSCaps(*w.SupportedLCSCapabilitySets)
	}
	if w.AdditionalLCSCapabilitySets != nil && w.AdditionalLCSCapabilitySets.BitLength > 0 {
		out.AdditionalLCSCapabilitySets = convertBitStringToLCSCaps(*w.AdditionalLCSCapabilitySets)
	}
	if w.MmeName != nil {
		mme := HexBytes(*w.MmeName)

		out.MmeName = mme
	}
	if w.AaaServerName != nil {
		aaa := HexBytes(*w.AaaServerName)

		out.AaaServerName = aaa
	}
	if w.SgsnName != nil {
		sgsn := HexBytes(*w.SgsnName)

		out.SgsnName = sgsn
	}
	if w.SgsnRealm != nil {
		realm := HexBytes(*w.SgsnRealm)

		out.SgsnRealm = realm
	}
	return out, nil
}

// ============================================================================
// DeferredmtLrData — TS 29.002 MAP-LCS-DataTypes.asn:673
// ============================================================================
//
// LcsLocationInfo may be present only if TerminationCause indicates
// mt-lrRestart per spec. That invariant is the caller's responsibility;
// the codec preserves whatever is set, since intermediaries may relay
// data they don't fully validate.

// isRecognizedTerminationCause reports whether v is one of the
// TerminationCause values 3GPP TS 29.002 V19.1.0 §17.7.13 lists, normal(0)
// to networkTermination(9). The encoder sends only these; the decoder treats
// any other value as errorundefined(1).
func isRecognizedTerminationCause(v TerminationCause) bool {
	return v >= TerminationNormal && v <= TerminationNetworkTermination
}

func convertDeferredmtLrDataToWire(d *DeferredmtLrData) (*gsm_map.DeferredmtLrData, error) {
	if d == nil {
		return nil, nil
	}
	bs := convertDeferredLocationEventTypeToBitString(&d.DeferredLocationEventType)
	out := &gsm_map.DeferredmtLrData{
		DeferredLocationEventType: bs,
	}
	if d.TerminationCause != nil {
		v := *d.TerminationCause
		if !isRecognizedTerminationCause(v) {
			return nil, fmt.Errorf("DeferredmtLrData.TerminationCause=%d: %w", v, ErrTerminationCauseInvalid)
		}
		out.TerminationCause = &v
	}
	if d.LcsLocationInfo != nil {
		li, err := convertLCSLocationInfoToWire(d.LcsLocationInfo)
		if err != nil {
			return nil, fmt.Errorf("DeferredmtLrData.LcsLocationInfo: %w", err)
		}
		out.LcsLocationInfo = li
	}
	return out, nil
}

func convertWireToDeferredmtLrData(w *gsm_map.DeferredmtLrData) (*DeferredmtLrData, error) {
	if w == nil {
		return nil, nil
	}
	det := convertBitStringToDeferredLocationEventType(w.DeferredLocationEventType)
	out := &DeferredmtLrData{
		DeferredLocationEventType: *det,
	}
	if w.TerminationCause != nil {
		// 3GPP TS 29.002 V19.1.0 §17.7.13 TerminationCause: "an
		// unrecognized value shall be treated the same as value 1
		// (errorundefined)".
		v := *w.TerminationCause
		if !isRecognizedTerminationCause(v) {
			v = TerminationErrorundefined
		}
		out.TerminationCause = &v
	}
	if w.LcsLocationInfo != nil {
		li, err := convertWireToLCSLocationInfo(w.LcsLocationInfo)
		if err != nil {
			return nil, fmt.Errorf("DeferredmtLrData.LcsLocationInfo: %w", err)
		}
		out.LcsLocationInfo = li
	}
	return out, nil
}
