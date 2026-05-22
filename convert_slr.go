// convert_slr.go
//
// Converters for SubscriberLocationReport (opCode 86) sub-types:
// LCSLocationInfo and DeferredmtLrData. PR G2 of the staged SLR
// implementation, building on PR #53 (foundation types). The
// top-level SubscriberLocationReportArg/Res converters and
// Marshal/Parse entry points land in subsequent PRs.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// diameterIdentityMinLen / diameterIdentityMaxLen (9..255 per RFC 6733,
// TS 29.002 MAP-MS-DataTypes.asn:1434) are defined in convert_psl_res.go.

// ============================================================================
// LCSLocationInfo — TS 29.002 MAP-LCS-DataTypes.asn
// ============================================================================

// validateDiameterIdentity checks an optional DiameterIdentity FQDN
// field (9..255 octets); empty/nil is treated as absent by the caller.
func validateDiameterIdentity(b HexBytes, sentinel error) error {
	if len(b) < diameterIdentityMinLen || len(b) > diameterIdentityMaxLen {
		return fmt.Errorf("len=%d: %w", len(b), sentinel)
	}
	return nil
}

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
		NetworkNodeNumber: gsm_map.ISDNAddressString(nodeWire),
	}

	if len(l.LMSI) > 0 {
		if len(l.LMSI) != 4 {
			return nil, fmt.Errorf("LCSLocationInfo.LMSI len=%d: %w", len(l.LMSI), ErrLCSLocationInfoLMSIInvalidSize)
		}
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
		if err := validateDiameterIdentity(l.MmeName, ErrLCSLocationInfoMmeNameSize); err != nil {
			return nil, fmt.Errorf("LCSLocationInfo.MmeName: %w", err)
		}
		v := gsm_map.DiameterIdentity(l.MmeName)
		out.MmeName = &v
	}
	if len(l.AaaServerName) > 0 {
		if err := validateDiameterIdentity(l.AaaServerName, ErrLCSLocationInfoAaaServerNameSize); err != nil {
			return nil, fmt.Errorf("LCSLocationInfo.AaaServerName: %w", err)
		}
		v := gsm_map.DiameterIdentity(l.AaaServerName)
		out.AaaServerName = &v
	}
	if len(l.SgsnName) > 0 {
		if err := validateDiameterIdentity(l.SgsnName, ErrLCSLocationInfoSgsnNameSize); err != nil {
			return nil, fmt.Errorf("LCSLocationInfo.SgsnName: %w", err)
		}
		v := gsm_map.DiameterIdentity(l.SgsnName)
		out.SgsnName = &v
	}
	if len(l.SgsnRealm) > 0 {
		if err := validateDiameterIdentity(l.SgsnRealm, ErrLCSLocationInfoSgsnRealmSize); err != nil {
			return nil, fmt.Errorf("LCSLocationInfo.SgsnRealm: %w", err)
		}
		v := gsm_map.DiameterIdentity(l.SgsnRealm)
		out.SgsnRealm = &v
	}
	return out, nil
}

func convertWireToLCSLocationInfo(w *gsm_map.LCSLocationInfo) (*LCSLocationInfo, error) {
	if w == nil {
		return nil, nil
	}
	node, nature, plan, err := decodeAddressField([]byte(w.NetworkNodeNumber))
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
		if len(*w.Lmsi) != 4 {
			return nil, fmt.Errorf("LCSLocationInfo.LMSI len=%d: %w", len(*w.Lmsi), ErrLCSLocationInfoLMSIInvalidSize)
		}
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
	// Guard with BitLength > 0 so a present-but-empty BIT STRING is
	// treated as absent — a zero-length wire value can't round-trip
	// through the struct-of-bools surrogate. Matches the existing
	// pattern in convert_updateloc.go / convert_updategprsloc.go.
	if w.SupportedLCSCapabilitySets != nil && w.SupportedLCSCapabilitySets.BitLength > 0 {
		out.SupportedLCSCapabilitySets = convertBitStringToLCSCaps(*w.SupportedLCSCapabilitySets)
	}
	if w.AdditionalLCSCapabilitySets != nil && w.AdditionalLCSCapabilitySets.BitLength > 0 {
		out.AdditionalLCSCapabilitySets = convertBitStringToLCSCaps(*w.AdditionalLCSCapabilitySets)
	}
	if w.MmeName != nil {
		mme := HexBytes(*w.MmeName)
		if err := validateDiameterIdentity(mme, ErrLCSLocationInfoMmeNameSize); err != nil {
			return nil, fmt.Errorf("LCSLocationInfo.MmeName: %w", err)
		}
		out.MmeName = mme
	}
	if w.AaaServerName != nil {
		aaa := HexBytes(*w.AaaServerName)
		if err := validateDiameterIdentity(aaa, ErrLCSLocationInfoAaaServerNameSize); err != nil {
			return nil, fmt.Errorf("LCSLocationInfo.AaaServerName: %w", err)
		}
		out.AaaServerName = aaa
	}
	if w.SgsnName != nil {
		sgsn := HexBytes(*w.SgsnName)
		if err := validateDiameterIdentity(sgsn, ErrLCSLocationInfoSgsnNameSize); err != nil {
			return nil, fmt.Errorf("LCSLocationInfo.SgsnName: %w", err)
		}
		out.SgsnName = sgsn
	}
	if w.SgsnRealm != nil {
		realm := HexBytes(*w.SgsnRealm)
		if err := validateDiameterIdentity(realm, ErrLCSLocationInfoSgsnRealmSize); err != nil {
			return nil, fmt.Errorf("LCSLocationInfo.SgsnRealm: %w", err)
		}
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
		// TerminationCause is extensible (TS 29.002:696); encoder
		// strict (0..9), decoder lenient.
		if int64(v) < 0 || int64(v) > 9 {
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
	det, err := convertBitStringToDeferredLocationEventType(w.DeferredLocationEventType)
	if err != nil {
		return nil, fmt.Errorf("DeferredmtLrData.DeferredLocationEventType: %w", err)
	}
	out := &DeferredmtLrData{
		DeferredLocationEventType: *det,
	}
	if w.TerminationCause != nil {
		v := *w.TerminationCause
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
