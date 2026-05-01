// convert_psl_lcsclientid.go
//
// Converters for the LCS-Client identifier tree referenced by the
// ProvideSubscriberLocation (opCode 83) Arg's lcs-ClientID field.
// PR D2 of the staged PSL implementation, building on PR #43 (leaf
// converters + BIT STRING codecs).
//
// Container converters land in this file:
//   - LCSClientName (USSD-DataCodingScheme + NameString + optional
//     LCSFormatIndicator)
//   - LCSRequestorID (USSD-DataCodingScheme + RequestorIDString +
//     optional LCSFormatIndicator)
//   - LCSClientID (LcsClientType enum + 6 optional sub-fields:
//     LcsClientExternalID (existing converter from convert_isd_lcs.go),
//     LcsClientDialedByMS (AddressString digits + Nature/Plan triple),
//     LcsClientInternalID (existing alias),
//     LcsClientName (this PR), LcsAPN (HexBytes opaque),
//     LcsRequestorID (this PR))

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// LCSClientName — TS 29.002 MAP-LCS-DataTypes.asn:199
// ============================================================================

func convertLCSClientNameToWire(c *LCSClientName) (*gsm_map.LCSClientName, error) {
	if c == nil {
		return nil, nil
	}
	if len(c.NameString) < 1 || len(c.NameString) > NameStringMaxLen {
		return nil, fmt.Errorf("LCSClientName.NameString len=%d: %w", len(c.NameString), ErrLCSClientNameNameStringSize)
	}
	out := &gsm_map.LCSClientName{
		DataCodingScheme: gsm_map.USSDDataCodingScheme{c.DataCodingScheme},
		NameString:       gsm_map.NameString(c.NameString),
	}
	if c.LcsFormatIndicator != nil {
		v := *c.LcsFormatIndicator
		// LCSFormatIndicator is extensible (TS 29.002:224); encoder
		// strict, decoder lenient (consistent with LcsClientType,
		// LocationEstimateType, etc.).
		if int64(v) < 0 || int64(v) > 4 {
			return nil, fmt.Errorf("LCSClientName.LcsFormatIndicator=%d: %w", v, ErrLCSFormatIndicatorInvalid)
		}
		out.LcsFormatIndicator = &v
	}
	return out, nil
}

func convertWireToLCSClientName(w *gsm_map.LCSClientName) (*LCSClientName, error) {
	if w == nil {
		return nil, nil
	}
	if len(w.DataCodingScheme) != 1 {
		return nil, fmt.Errorf("LCSClientName.DataCodingScheme len=%d: %w", len(w.DataCodingScheme), ErrUSSDDataCodingSchemeInvalidSize)
	}
	if len(w.NameString) < 1 || len(w.NameString) > NameStringMaxLen {
		return nil, fmt.Errorf("LCSClientName.NameString len=%d: %w", len(w.NameString), ErrLCSClientNameNameStringSize)
	}
	out := &LCSClientName{
		DataCodingScheme: w.DataCodingScheme[0],
		NameString:       HexBytes(w.NameString),
	}
	if w.LcsFormatIndicator != nil {
		v := *w.LcsFormatIndicator
		out.LcsFormatIndicator = &v
	}
	return out, nil
}

// ============================================================================
// LCSRequestorID — TS 29.002 MAP-LCS-DataTypes.asn:214
// ============================================================================

func convertLCSRequestorIDToWire(r *LCSRequestorID) (*gsm_map.LCSRequestorID, error) {
	if r == nil {
		return nil, nil
	}
	if len(r.RequestorIDString) < 1 || len(r.RequestorIDString) > RequestorIDStringMaxLen {
		return nil, fmt.Errorf("LCSRequestorID.RequestorIDString len=%d: %w", len(r.RequestorIDString), ErrLCSRequestorIDStringSize)
	}
	out := &gsm_map.LCSRequestorID{
		DataCodingScheme:  gsm_map.USSDDataCodingScheme{r.DataCodingScheme},
		RequestorIDString: gsm_map.RequestorIDString(r.RequestorIDString),
	}
	if r.LcsFormatIndicator != nil {
		v := *r.LcsFormatIndicator
		if int64(v) < 0 || int64(v) > 4 {
			return nil, fmt.Errorf("LCSRequestorID.LcsFormatIndicator=%d: %w", v, ErrLCSFormatIndicatorInvalid)
		}
		out.LcsFormatIndicator = &v
	}
	return out, nil
}

func convertWireToLCSRequestorID(w *gsm_map.LCSRequestorID) (*LCSRequestorID, error) {
	if w == nil {
		return nil, nil
	}
	if len(w.DataCodingScheme) != 1 {
		return nil, fmt.Errorf("LCSRequestorID.DataCodingScheme len=%d: %w", len(w.DataCodingScheme), ErrUSSDDataCodingSchemeInvalidSize)
	}
	if len(w.RequestorIDString) < 1 || len(w.RequestorIDString) > RequestorIDStringMaxLen {
		return nil, fmt.Errorf("LCSRequestorID.RequestorIDString len=%d: %w", len(w.RequestorIDString), ErrLCSRequestorIDStringSize)
	}
	out := &LCSRequestorID{
		DataCodingScheme:  w.DataCodingScheme[0],
		RequestorIDString: HexBytes(w.RequestorIDString),
	}
	if w.LcsFormatIndicator != nil {
		v := *w.LcsFormatIndicator
		out.LcsFormatIndicator = &v
	}
	return out, nil
}

// ============================================================================
// LCSClientID — TS 29.002 MAP-LCS-DataTypes.asn:178
// ============================================================================
//
// LcsClientType is an extensible ENUMERATED (TS 29.002:188); encoder is
// strict (0..3), decoder preserves unknown values per Postel.
// LcsClientDialedByMS is an AddressString surfaced as digits +
// Nature/Plan triple consistent with the rest of the public API; empty
// digits = absent.

func convertLCSClientIDToWire(c *LCSClientID) (*gsm_map.LCSClientID, error) {
	if c == nil {
		return nil, nil
	}
	if int64(c.LcsClientType) < 0 || int64(c.LcsClientType) > 3 {
		return nil, fmt.Errorf("LCSClientID.LcsClientType=%d: %w", c.LcsClientType, ErrLCSClientTypeInvalid)
	}
	out := &gsm_map.LCSClientID{
		LcsClientType: c.LcsClientType,
	}
	if c.LcsClientExternalID != nil {
		ext, err := convertLCSClientExternalIDToWire(c.LcsClientExternalID)
		if err != nil {
			return nil, fmt.Errorf("LCSClientID.LcsClientExternalID: %w", err)
		}
		out.LcsClientExternalID = ext
	}
	// Symmetry with the decode-side ErrLCSClientIDDialedByMSEmpty
	// invariant: empty digits combined with non-zero Nature/Plan
	// indicates a caller bug (Nature/Plan are only meaningful when
	// digits are present).
	if c.LcsClientDialedByMS == "" {
		if c.LcsClientDialedByMSNature != 0 || c.LcsClientDialedByMSPlan != 0 {
			return nil, fmt.Errorf("LCSClientID.LcsClientDialedByMS: %w", ErrLCSClientIDDialedByMSEmpty)
		}
	} else {
		isdn, err := encodeAddressField(c.LcsClientDialedByMS, c.LcsClientDialedByMSNature, c.LcsClientDialedByMSPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding LCSClientID.LcsClientDialedByMS: %w", err)
		}
		v := gsm_map.AddressString(isdn)
		out.LcsClientDialedByMS = &v
	}
	if c.LcsClientInternalID != nil {
		v := *c.LcsClientInternalID
		// LCSClientInternalID is a non-extensible enum (0..4) per
		// TS 29.002 MAP-CommonDataTypes.asn; validate symmetrically
		// with PLMNClientList's per-entry check
		// (ErrLCSClientInternalIDInvalid).
		if int64(v) < 0 || int64(v) > 4 {
			return nil, fmt.Errorf("LCSClientID.LcsClientInternalID=%d: %w", v, ErrLCSClientInternalIDInvalid)
		}
		out.LcsClientInternalID = &v
	}
	if c.LcsClientName != nil {
		nm, err := convertLCSClientNameToWire(c.LcsClientName)
		if err != nil {
			return nil, fmt.Errorf("LCSClientID.LcsClientName: %w", err)
		}
		out.LcsClientName = nm
	}
	if len(c.LcsAPN) > 0 {
		if err := validateAPN(c.LcsAPN, "LCSClientID.LcsAPN"); err != nil {
			return nil, err
		}
		v := gsm_map.APN(c.LcsAPN)
		out.LcsAPN = &v
	}
	if c.LcsRequestorID != nil {
		rid, err := convertLCSRequestorIDToWire(c.LcsRequestorID)
		if err != nil {
			return nil, fmt.Errorf("LCSClientID.LcsRequestorID: %w", err)
		}
		out.LcsRequestorID = rid
	}
	return out, nil
}

func convertWireToLCSClientID(w *gsm_map.LCSClientID) (*LCSClientID, error) {
	if w == nil {
		return nil, nil
	}
	out := &LCSClientID{
		LcsClientType: w.LcsClientType,
	}
	if w.LcsClientExternalID != nil {
		ext, err := convertWireToLCSClientExternalID(w.LcsClientExternalID)
		if err != nil {
			return nil, fmt.Errorf("LCSClientID.LcsClientExternalID: %w", err)
		}
		out.LcsClientExternalID = ext
	}
	if w.LcsClientDialedByMS != nil {
		s, nature, plan, err := decodeAddressField([]byte(*w.LcsClientDialedByMS))
		if err != nil {
			return nil, fmt.Errorf("decoding LCSClientID.LcsClientDialedByMS: %w", err)
		}
		// Per the project convention (e.g., ErrIsdMSISDNDecodedEmpty),
		// an explicitly present wire AddressString that decodes to
		// empty digits cannot round-trip through the string-based API.
		if s == "" {
			return nil, fmt.Errorf("LCSClientID.LcsClientDialedByMS: %w", ErrLCSClientIDDialedByMSEmpty)
		}
		out.LcsClientDialedByMS = s
		out.LcsClientDialedByMSNature = nature
		out.LcsClientDialedByMSPlan = plan
	}
	if w.LcsClientInternalID != nil {
		v := *w.LcsClientInternalID
		if int64(v) < 0 || int64(v) > 4 {
			return nil, fmt.Errorf("LCSClientID.LcsClientInternalID=%d: %w", v, ErrLCSClientInternalIDInvalid)
		}
		out.LcsClientInternalID = &v
	}
	if w.LcsClientName != nil {
		nm, err := convertWireToLCSClientName(w.LcsClientName)
		if err != nil {
			return nil, fmt.Errorf("LCSClientID.LcsClientName: %w", err)
		}
		out.LcsClientName = nm
	}
	if w.LcsAPN != nil {
		apn := HexBytes(*w.LcsAPN)
		if err := validateAPN(apn, "LCSClientID.LcsAPN"); err != nil {
			return nil, err
		}
		out.LcsAPN = apn
	}
	if w.LcsRequestorID != nil {
		rid, err := convertWireToLCSRequestorID(w.LcsRequestorID)
		if err != nil {
			return nil, fmt.Errorf("LCSClientID.LcsRequestorID: %w", err)
		}
		out.LcsRequestorID = rid
	}
	return out, nil
}
