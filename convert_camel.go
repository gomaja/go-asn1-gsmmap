package gsmmap

import (
	"fmt"
	"slices"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// --- CAMEL subscription info converters ---
//
// Full field-for-field conversion between the public CAMEL types and the
// go-asn1 wire types. Replaces the earlier opaque-HexBytes stubs that
// silently dropped CAMEL subscription data on round-trip.

// isValidOBcsmTDP reports whether v is a defined originating BCSM TDP.
func isValidOBcsmTDP(v OBcsmTriggerDetectionPoint) bool {
	switch v {
	case OBcsmTriggerCollectedInfo, OBcsmTriggerRouteSelectFailure:
		return true
	}
	return false
}

// isValidTBcsmTDP reports whether v is a defined terminating BCSM TDP.
func isValidTBcsmTDP(v TBcsmTriggerDetectionPoint) bool {
	switch v {
	case TBcsmTriggerTermAttemptAuthorized, TBcsmTriggerTBusy, TBcsmTriggerTNoAnswer:
		return true
	}
	return false
}

// isValidDefaultCallHandling reports whether v is a defined value.
func isValidDefaultCallHandling(v DefaultCallHandling) bool {
	switch v {
	case DefaultCallHandlingContinueCall, DefaultCallHandlingReleaseCall:
		return true
	}
	return false
}

// isValidCallTypeCriteria reports whether v is a defined value.
func isValidCallTypeCriteria(v CallTypeCriteria) bool {
	switch v {
	case CallTypeCriteriaForwarded, CallTypeCriteriaNotForwarded:
		return true
	}
	return false
}

// isValidMatchType reports whether v is a defined value.
func isValidMatchType(v MatchType) bool {
	switch v {
	case MatchTypeInhibiting, MatchTypeEnabling:
		return true
	}
	return false
}

// validateCamelCapabilityHandling enforces the 1..4 phase range on encode:
// a sender uses only the defined CAMEL phases 1 to 4.
func validateCamelCapabilityHandling(p *int) error {
	if p == nil {
		return nil
	}
	if *p < 1 || *p > 4 {
		return fmt.Errorf("%w (got %d)", ErrCamelCapabilityHandlingOutOfRange, *p)
	}
	return nil
}

// camelCapabilityHandlingFromWire applies 3GPP TS 29.002 V19.1.0 §17.7.1
// CamelCapabilityHandling: "reception of values greater than 4 shall be
// treated as CAMEL phase 4." The codec has already enforced INTEGER (1..16).
func camelCapabilityHandlingFromWire(w gsm_map.CamelCapabilityHandling) *int {
	v := int(min(w, 4))
	return &v
}

// defaultCallHandlingFromWire applies 3GPP TS 29.002 V19.1.0 §17.7.1
// DefaultCallHandling: "reception of values in range 2-31 shall be treated
// as "continueCall"" and "reception of values greater than 31 shall be
// treated as "releaseCall"". A negative value lies outside both ranges and
// is rejected.
func defaultCallHandlingFromWire(w DefaultCallHandling) (DefaultCallHandling, error) {
	switch {
	case w < 0:
		return 0, fmt.Errorf("%w (got %d)", ErrCamelInvalidDefaultCallHandling, w)
	case w == DefaultCallHandlingContinueCall, w == DefaultCallHandlingReleaseCall:
		return w, nil
	case w <= 31:
		return DefaultCallHandlingContinueCall, nil
	default:
		return DefaultCallHandlingReleaseCall, nil
	}
}

// convertIgnorableWireList converts each wire entry with conv, keeping the
// entries conv returns and dropping those the receiver ignores (conv returns
// nil). It returns nil when every entry is ignored. Errors carry field and
// the wire index.
func convertIgnorableWireList[W, T any](field string, ws []W, conv func(*W) (*T, error)) ([]T, error) {
	var out []T
	for i := range ws {
		v, err := conv(&ws[i])
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", field, i, err)
		}
		if v != nil {
			out = append(out, *v)
		}
	}
	return out, nil
}

// convertOBcsmTDPDataToWire encodes a single O-BCSM TDP entry.
func convertOBcsmTDPDataToWire(d *OBcsmCamelTDPData) (gsm_map.OBcsmCamelTDPData, error) {
	if !isValidOBcsmTDP(d.OBcsmTriggerDetectionPoint) {
		return gsm_map.OBcsmCamelTDPData{}, ErrCamelInvalidOTriggerPoint
	}

	if d.GsmSCFAddress == "" {
		return gsm_map.OBcsmCamelTDPData{}, ErrCamelMissingGsmSCFAddress
	}
	if !isValidDefaultCallHandling(d.DefaultCallHandling) {
		return gsm_map.OBcsmCamelTDPData{}, ErrCamelInvalidDefaultCallHandling
	}
	addr, err := encodeAddressField(d.GsmSCFAddress, d.GsmSCFAddressNature, d.GsmSCFAddressPlan)
	if err != nil {
		return gsm_map.OBcsmCamelTDPData{}, fmt.Errorf("encoding GsmSCFAddress: %w", err)
	}
	return gsm_map.OBcsmCamelTDPData{
		OBcsmTriggerDetectionPoint: d.OBcsmTriggerDetectionPoint,
		ServiceKey:                 gsm_map.ServiceKey(d.ServiceKey),
		GsmSCFAddress:              gsm_map.ISDNAddressString(addr),
		DefaultCallHandling:        d.DefaultCallHandling,
	}, nil
}

// convertWireToOBcsmTDPData decodes a single wire O-BCSM TDP entry. It
// returns nil, without examining the other fields, for an entry the
// receiver ignores: 3GPP TS 29.002 V19.1.0 §17.7.1
// O-BcsmTriggerDetectionPoint, "For O-BcsmCamelTDPData sequences containing
// this parameter with any other value than the ones listed the receiver
// shall ignore the whole O-BcsmCamelTDPData sequence."
func convertWireToOBcsmTDPData(w *gsm_map.OBcsmCamelTDPData) (*OBcsmCamelTDPData, error) {
	if !isValidOBcsmTDP(w.OBcsmTriggerDetectionPoint) {
		return nil, nil
	}
	sk := int64(w.ServiceKey)

	dch, err := defaultCallHandlingFromWire(w.DefaultCallHandling)
	if err != nil {
		return nil, err
	}
	digits, nature, plan, err := decodeAddressField(w.GsmSCFAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding GsmSCFAddress: %w", err)
	}
	if digits == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}
	return &OBcsmCamelTDPData{
		OBcsmTriggerDetectionPoint: w.OBcsmTriggerDetectionPoint,
		ServiceKey:                 sk,
		GsmSCFAddress:              digits,
		GsmSCFAddressNature:        nature,
		GsmSCFAddressPlan:          plan,
		DefaultCallHandling:        dch,
	}, nil
}

// convertTBcsmTDPDataToWire encodes a single T-BCSM TDP entry.
func convertTBcsmTDPDataToWire(d *TBcsmCamelTDPData) (gsm_map.TBcsmCamelTDPData, error) {
	if !isValidTBcsmTDP(d.TBcsmTriggerDetectionPoint) {
		return gsm_map.TBcsmCamelTDPData{}, ErrCamelInvalidTTriggerPoint
	}

	if d.GsmSCFAddress == "" {
		return gsm_map.TBcsmCamelTDPData{}, ErrCamelMissingGsmSCFAddress
	}
	if !isValidDefaultCallHandling(d.DefaultCallHandling) {
		return gsm_map.TBcsmCamelTDPData{}, ErrCamelInvalidDefaultCallHandling
	}
	addr, err := encodeAddressField(d.GsmSCFAddress, d.GsmSCFAddressNature, d.GsmSCFAddressPlan)
	if err != nil {
		return gsm_map.TBcsmCamelTDPData{}, fmt.Errorf("encoding GsmSCFAddress: %w", err)
	}
	return gsm_map.TBcsmCamelTDPData{
		TBcsmTriggerDetectionPoint: d.TBcsmTriggerDetectionPoint,
		ServiceKey:                 gsm_map.ServiceKey(d.ServiceKey),
		GsmSCFAddress:              gsm_map.ISDNAddressString(addr),
		DefaultCallHandling:        d.DefaultCallHandling,
	}, nil
}

// convertWireToTBcsmTDPData decodes a single wire T-BCSM TDP entry. It
// returns nil, without examining the other fields, for an entry the
// receiver ignores: 3GPP TS 29.002 V19.1.0 §17.7.1
// T-BcsmTriggerDetectionPoint, "For T-BcsmCamelTDPData sequences containing
// this parameter with any other value than the ones listed above, the
// receiver shall ignore the whole T-BcsmCamelTDPData sequence."
func convertWireToTBcsmTDPData(w *gsm_map.TBcsmCamelTDPData) (*TBcsmCamelTDPData, error) {
	if !isValidTBcsmTDP(w.TBcsmTriggerDetectionPoint) {
		return nil, nil
	}
	sk := int64(w.ServiceKey)

	dch, err := defaultCallHandlingFromWire(w.DefaultCallHandling)
	if err != nil {
		return nil, err
	}
	digits, nature, plan, err := decodeAddressField(w.GsmSCFAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding GsmSCFAddress: %w", err)
	}
	if digits == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}
	return &TBcsmCamelTDPData{
		TBcsmTriggerDetectionPoint: w.TBcsmTriggerDetectionPoint,
		ServiceKey:                 sk,
		GsmSCFAddress:              digits,
		GsmSCFAddressNature:        nature,
		GsmSCFAddressPlan:          plan,
		DefaultCallHandling:        dch,
	}, nil
}

// convertDestinationNumberCriteriaToWire encodes the DestinationNumberCriteria
// SEQUENCE, enforcing that at least one of the two lists is present.
func convertDestinationNumberCriteriaToWire(c *DestinationNumberCriteria) (*gsm_map.DestinationNumberCriteria, error) {
	if !isValidMatchType(c.MatchType) {
		return nil, ErrCamelInvalidMatchType
	}
	if len(c.DestinationNumberList) == 0 && len(c.DestinationNumberLengthList) == 0 {
		return nil, ErrCamelMissingDestinationNumberCriteria
	}
	out := &gsm_map.DestinationNumberCriteria{
		MatchType: c.MatchType,
	}
	if len(c.DestinationNumberList) > 0 {
		list := gsm_map.DestinationNumberList{Values: make([]gsm_map.ISDNAddressString, len(c.DestinationNumberList))}
		for i, n := range c.DestinationNumberList {
			if n.Digits == "" {
				return nil, fmt.Errorf("DestinationNumberList[%d]: %w", i, ErrCamelMissingDestinationNumber)
			}
			enc, err := encodeAddressField(n.Digits, n.Nature, n.Plan)
			if err != nil {
				return nil, fmt.Errorf("DestinationNumberList[%d]: %w", i, err)
			}
			list.Values[i] = gsm_map.ISDNAddressString(enc)
		}
		out.DestinationNumberList = &list
	}
	if len(c.DestinationNumberLengthList) > 0 {
		list := gsm_map.DestinationNumberLengthList{Values: make([]int64, len(c.DestinationNumberLengthList))}
		for i, l := range c.DestinationNumberLengthList {
			list.Values[i] = int64(l)
		}
		out.DestinationNumberLengthList = &list
	}
	return out, nil
}

// convertWireToDestinationNumberCriteria decodes the criteria SEQUENCE.
// Mirrors the encoder's "at least one list" rule so malformed peer input
// can't produce a criteria SEQUENCE with neither list populated.
func convertWireToDestinationNumberCriteria(w *gsm_map.DestinationNumberCriteria) (*DestinationNumberCriteria, error) {
	mt := MatchType(w.MatchType)
	if !isValidMatchType(mt) {
		return nil, ErrCamelInvalidMatchType
	}
	if (w.DestinationNumberList == nil || len(w.DestinationNumberList.Values) == 0) &&
		(w.DestinationNumberLengthList == nil || len(w.DestinationNumberLengthList.Values) == 0) {
		return nil, ErrCamelMissingDestinationNumberCriteria
	}
	out := &DestinationNumberCriteria{MatchType: mt}
	if w.DestinationNumberList != nil && len(w.DestinationNumberList.Values) > 0 {
		list := make([]ISDNNumber, len(w.DestinationNumberList.Values))
		for i, n := range w.DestinationNumberList.Values {
			digits, nature, plan, err := decodeAddressField(n)
			if err != nil {
				return nil, fmt.Errorf("DestinationNumberList[%d]: %w", i, err)
			}
			if digits == "" {
				return nil, fmt.Errorf("DestinationNumberList[%d]: %w", i, ErrCamelMissingDestinationNumber)
			}
			list[i] = ISDNNumber{Digits: digits, Nature: nature, Plan: plan}
		}
		out.DestinationNumberList = list
	}
	if w.DestinationNumberLengthList != nil && len(w.DestinationNumberLengthList.Values) > 0 {
		list := make([]int, len(w.DestinationNumberLengthList.Values))
		for i, l := range w.DestinationNumberLengthList.Values {
			list[i] = int(l)
		}
		out.DestinationNumberLengthList = list
	}
	return out, nil
}

// convertOBcsmTDPCriteriaToWire encodes an O-BCSM TDP criteria entry.
func convertOBcsmTDPCriteriaToWire(c *OBcsmCamelTDPCriteria) (gsm_map.OBcsmCamelTDPCriteria, error) {
	if !isValidOBcsmTDP(c.OBcsmTriggerDetectionPoint) {
		return gsm_map.OBcsmCamelTDPCriteria{}, ErrCamelInvalidOTriggerPoint
	}
	out := gsm_map.OBcsmCamelTDPCriteria{
		OBcsmTriggerDetectionPoint: c.OBcsmTriggerDetectionPoint,
	}
	if c.DestinationNumberCriteria != nil {
		dnc, err := convertDestinationNumberCriteriaToWire(c.DestinationNumberCriteria)
		if err != nil {
			return gsm_map.OBcsmCamelTDPCriteria{}, fmt.Errorf("DestinationNumberCriteria: %w", err)
		}
		out.DestinationNumberCriteria = dnc
	}
	if len(c.BasicServiceCriteria) > 0 {
		bsc := gsm_map.BasicServiceCriteria{Values: make([]gsm_map.ExtBasicServiceCode, len(c.BasicServiceCriteria))}
		for i := range c.BasicServiceCriteria {
			wv, err := convertExtBasicServiceCodeToWire(&c.BasicServiceCriteria[i])
			if err != nil {
				return gsm_map.OBcsmCamelTDPCriteria{}, fmt.Errorf("BasicServiceCriteria[%d]: %w", i, err)
			}
			bsc.Values[i] = *wv
		}
		out.BasicServiceCriteria = &bsc
	}
	if c.CallTypeCriteria != nil {
		if !isValidCallTypeCriteria(*c.CallTypeCriteria) {
			return gsm_map.OBcsmCamelTDPCriteria{}, ErrCamelInvalidCallTypeCriteria
		}
		ctc := *c.CallTypeCriteria
		out.CallTypeCriteria = &ctc
	}
	if len(c.OCauseValueCriteria) > 0 {
		list := gsm_map.OCauseValueCriteria{Values: make([]gsm_map.CauseValue, len(c.OCauseValueCriteria))}
		for i, v := range c.OCauseValueCriteria {
			if v < 0 || v > 127 {
				return gsm_map.OBcsmCamelTDPCriteria{}, fmt.Errorf("OCauseValueCriteria[%d]: %w", i, ErrCamelInvalidCauseValue)
			}
			list.Values[i] = gsm_map.CauseValue{byte(v)}
		}
		out.OCauseValueCriteria = &list
	}
	return out, nil
}

// convertWireToOBcsmTDPCriteria decodes an O-BCSM TDP criteria entry. It
// returns nil, without examining the other fields, for an entry the
// receiver ignores: 3GPP TS 29.002 V19.1.0 §17.7.1
// O-BcsmTriggerDetectionPoint, "For O-BcsmCamelTDP-Criteria sequences
// containing this parameter with any other value than the ones listed the
// receiver shall ignore the whole O-BcsmCamelTDP-Criteria sequence."
func convertWireToOBcsmTDPCriteria(w *gsm_map.OBcsmCamelTDPCriteria) (*OBcsmCamelTDPCriteria, error) {
	if !isValidOBcsmTDP(w.OBcsmTriggerDetectionPoint) {
		return nil, nil
	}
	out := &OBcsmCamelTDPCriteria{OBcsmTriggerDetectionPoint: w.OBcsmTriggerDetectionPoint}
	if w.DestinationNumberCriteria != nil {
		dnc, err := convertWireToDestinationNumberCriteria(w.DestinationNumberCriteria)
		if err != nil {
			return nil, fmt.Errorf("DestinationNumberCriteria: %w", err)
		}
		out.DestinationNumberCriteria = dnc
	}
	if w.BasicServiceCriteria != nil && len(w.BasicServiceCriteria.Values) > 0 {
		bsc := make([]ExtBasicServiceCode, len(w.BasicServiceCriteria.Values))
		for i := range w.BasicServiceCriteria.Values {
			pv, err := convertWireToExtBasicServiceCode(&w.BasicServiceCriteria.Values[i])
			if err != nil {
				return nil, fmt.Errorf("BasicServiceCriteria[%d]: %w", i, err)
			}
			bsc[i] = *pv
		}
		out.BasicServiceCriteria = bsc
	}
	if w.CallTypeCriteria != nil {
		ctc := CallTypeCriteria(*w.CallTypeCriteria)
		if !isValidCallTypeCriteria(ctc) {
			return nil, ErrCamelInvalidCallTypeCriteria
		}
		out.CallTypeCriteria = &ctc
	}
	if w.OCauseValueCriteria != nil {
		list := make([]int, len(w.OCauseValueCriteria.Values))
		for i, b := range w.OCauseValueCriteria.Values {
			// CauseValue is OCTET STRING (SIZE(1)); reject any other length
			// rather than silently normalising missing/extra octets.
			if len(b) != 1 {
				return nil, fmt.Errorf("OCauseValueCriteria[%d]: %w", i, ErrCamelInvalidCauseValueOctetLength)
			}
			v := int(b[0])
			if v > 127 {
				return nil, fmt.Errorf("OCauseValueCriteria[%d]: %w", i, ErrCamelInvalidCauseValue)
			}
			list[i] = v
		}
		out.OCauseValueCriteria = list
	}
	return out, nil
}

// convertTBcsmTDPCriteriaToWire encodes a T-BCSM TDP criteria entry.
func convertTBcsmTDPCriteriaToWire(c *TBcsmCamelTDPCriteria) (gsm_map.TBCSMCAMELTDPCriteria, error) {
	if !isValidTBcsmTDP(c.TBcsmTriggerDetectionPoint) {
		return gsm_map.TBCSMCAMELTDPCriteria{}, ErrCamelInvalidTTriggerPoint
	}
	out := gsm_map.TBCSMCAMELTDPCriteria{
		TBCSMTriggerDetectionPoint: c.TBcsmTriggerDetectionPoint,
	}
	if len(c.BasicServiceCriteria) > 0 {
		bsc := gsm_map.BasicServiceCriteria{Values: make([]gsm_map.ExtBasicServiceCode, len(c.BasicServiceCriteria))}
		for i := range c.BasicServiceCriteria {
			wv, err := convertExtBasicServiceCodeToWire(&c.BasicServiceCriteria[i])
			if err != nil {
				return gsm_map.TBCSMCAMELTDPCriteria{}, fmt.Errorf("BasicServiceCriteria[%d]: %w", i, err)
			}
			bsc.Values[i] = *wv
		}
		out.BasicServiceCriteria = &bsc
	}
	if len(c.TCauseValueCriteria) > 0 {
		list := gsm_map.TCauseValueCriteria{Values: make([]gsm_map.CauseValue, len(c.TCauseValueCriteria))}
		for i, v := range c.TCauseValueCriteria {
			if v < 0 || v > 127 {
				return gsm_map.TBCSMCAMELTDPCriteria{}, fmt.Errorf("TCauseValueCriteria[%d]: %w", i, ErrCamelInvalidCauseValue)
			}
			list.Values[i] = gsm_map.CauseValue{byte(v)}
		}
		out.TCauseValueCriteria = &list
	}
	return out, nil
}

// convertWireToTBcsmTDPCriteria decodes a T-BCSM TDP criteria entry.
func convertWireToTBcsmTDPCriteria(w *gsm_map.TBCSMCAMELTDPCriteria) (TBcsmCamelTDPCriteria, error) {
	tdp := TBcsmTriggerDetectionPoint(w.TBCSMTriggerDetectionPoint)
	if !isValidTBcsmTDP(tdp) {
		return TBcsmCamelTDPCriteria{}, ErrCamelInvalidTTriggerPoint
	}
	out := TBcsmCamelTDPCriteria{TBcsmTriggerDetectionPoint: tdp}
	if w.BasicServiceCriteria != nil && len(w.BasicServiceCriteria.Values) > 0 {
		bsc := make([]ExtBasicServiceCode, len(w.BasicServiceCriteria.Values))
		for i := range w.BasicServiceCriteria.Values {
			pv, err := convertWireToExtBasicServiceCode(&w.BasicServiceCriteria.Values[i])
			if err != nil {
				return TBcsmCamelTDPCriteria{}, fmt.Errorf("BasicServiceCriteria[%d]: %w", i, err)
			}
			bsc[i] = *pv
		}
		out.BasicServiceCriteria = bsc
	}
	if w.TCauseValueCriteria != nil {
		list := make([]int, len(w.TCauseValueCriteria.Values))
		for i, b := range w.TCauseValueCriteria.Values {
			if len(b) != 1 {
				return TBcsmCamelTDPCriteria{}, fmt.Errorf("TCauseValueCriteria[%d]: %w", i, ErrCamelInvalidCauseValueOctetLength)
			}
			v := int(b[0])
			if v > 127 {
				return TBcsmCamelTDPCriteria{}, fmt.Errorf("TCauseValueCriteria[%d]: %w", i, ErrCamelInvalidCauseValue)
			}
			list[i] = v
		}
		out.TCauseValueCriteria = list
	}
	return out, nil
}

// convertOCSIToWire encodes an O-CSI.
func convertOCSIToWire(o *OCSI) (*gsm_map.OCSI, error) {
	if err := validateCamelCapabilityHandling(o.CamelCapabilityHandling); err != nil {
		return nil, err
	}
	list := gsm_map.OBcsmCamelTDPDataList{Values: make([]gsm_map.OBcsmCamelTDPData, len(o.OBcsmCamelTDPDataList))}
	for i := range o.OBcsmCamelTDPDataList {
		w, err := convertOBcsmTDPDataToWire(&o.OBcsmCamelTDPDataList[i])
		if err != nil {
			return nil, fmt.Errorf("OBcsmCamelTDPDataList[%d]: %w", i, err)
		}
		list.Values[i] = w
	}
	out := &gsm_map.OCSI{OBcsmCamelTDPDataList: &list}
	if o.CamelCapabilityHandling != nil {
		v := gsm_map.CamelCapabilityHandling(int64(*o.CamelCapabilityHandling))
		out.CamelCapabilityHandling = &v
	}
	out.NotificationToCSE = boolToNullPtr(o.NotificationToCSE)
	out.CsiActive = boolToNullPtr(o.CsiActive)
	return out, nil
}

// convertWireToOCSI decodes a wire O-CSI. It returns nil when the receiver
// ignores every O-BcsmCamelTDPData (convertWireToOBcsmTDPData): an O-CSI
// arms its TDPs only through O-BcsmCamelTDPDataList, SIZE (1..10), so with
// none left the receiver holds no O-CSI.
func convertWireToOCSI(w *gsm_map.OCSI) (*OCSI, error) {
	list, err := convertIgnorableWireList("OBcsmCamelTDPDataList", w.OBcsmCamelTDPDataList.Values, convertWireToOBcsmTDPData)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, nil
	}
	out := &OCSI{OBcsmCamelTDPDataList: list}
	if w.CamelCapabilityHandling != nil {
		out.CamelCapabilityHandling = camelCapabilityHandlingFromWire(*w.CamelCapabilityHandling)
	}
	out.NotificationToCSE = nullPtrToBool(w.NotificationToCSE)
	out.CsiActive = nullPtrToBool(w.CsiActive)
	return out, nil
}

// convertTCSIToWire encodes a T-CSI.
func convertTCSIToWire(t *TCSI) (*gsm_map.TCSI, error) {
	if err := validateCamelCapabilityHandling(t.CamelCapabilityHandling); err != nil {
		return nil, err
	}
	list := gsm_map.TBcsmCamelTDPDataList{Values: make([]gsm_map.TBcsmCamelTDPData, len(t.TBcsmCamelTDPDataList))}
	for i := range t.TBcsmCamelTDPDataList {
		w, err := convertTBcsmTDPDataToWire(&t.TBcsmCamelTDPDataList[i])
		if err != nil {
			return nil, fmt.Errorf("TBcsmCamelTDPDataList[%d]: %w", i, err)
		}
		list.Values[i] = w
	}
	out := &gsm_map.TCSI{TBcsmCamelTDPDataList: &list}
	if t.CamelCapabilityHandling != nil {
		v := gsm_map.CamelCapabilityHandling(int64(*t.CamelCapabilityHandling))
		out.CamelCapabilityHandling = &v
	}
	out.NotificationToCSE = boolToNullPtr(t.NotificationToCSE)
	out.CsiActive = boolToNullPtr(t.CsiActive)
	return out, nil
}

// convertWireToTCSI decodes a wire T-CSI or VT-CSI. It returns nil when the
// receiver ignores every T-BcsmCamelTDPData (convertWireToTBcsmTDPData): a
// T-CSI arms its TDPs only through T-BcsmCamelTDPDataList, SIZE (1..10), so
// with none left the receiver holds no T-CSI.
func convertWireToTCSI(w *gsm_map.TCSI) (*TCSI, error) {
	list, err := convertIgnorableWireList("TBcsmCamelTDPDataList", w.TBcsmCamelTDPDataList.Values, convertWireToTBcsmTDPData)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, nil
	}
	out := &TCSI{TBcsmCamelTDPDataList: list}
	if w.CamelCapabilityHandling != nil {
		out.CamelCapabilityHandling = camelCapabilityHandlingFromWire(*w.CamelCapabilityHandling)
	}
	out.NotificationToCSE = nullPtrToBool(w.NotificationToCSE)
	out.CsiActive = nullPtrToBool(w.CsiActive)
	return out, nil
}

// convertDPAnalysedInfoCriteriumToWire encodes one D-CSI entry.
func convertDPAnalysedInfoCriteriumToWire(c *DPAnalysedInfoCriterium) (gsm_map.DPAnalysedInfoCriterium, error) {
	if c.DialledNumber == "" {
		return gsm_map.DPAnalysedInfoCriterium{}, ErrCamelMissingDialledNumber
	}

	if c.GsmSCFAddress == "" {
		return gsm_map.DPAnalysedInfoCriterium{}, ErrCamelMissingGsmSCFAddress
	}
	if !isValidDefaultCallHandling(c.DefaultCallHandling) {
		return gsm_map.DPAnalysedInfoCriterium{}, ErrCamelInvalidDefaultCallHandling
	}
	dn, err := encodeAddressField(c.DialledNumber, c.DialledNumberNature, c.DialledNumberPlan)
	if err != nil {
		return gsm_map.DPAnalysedInfoCriterium{}, fmt.Errorf("encoding DialledNumber: %w", err)
	}
	sc, err := encodeAddressField(c.GsmSCFAddress, c.GsmSCFAddressNature, c.GsmSCFAddressPlan)
	if err != nil {
		return gsm_map.DPAnalysedInfoCriterium{}, fmt.Errorf("encoding GsmSCFAddress: %w", err)
	}
	return gsm_map.DPAnalysedInfoCriterium{
		DialledNumber:       gsm_map.ISDNAddressString(dn),
		ServiceKey:          gsm_map.ServiceKey(c.ServiceKey),
		GsmSCFAddress:       gsm_map.ISDNAddressString(sc),
		DefaultCallHandling: c.DefaultCallHandling,
	}, nil
}

// convertWireToDPAnalysedInfoCriterium decodes a single D-CSI entry.
func convertWireToDPAnalysedInfoCriterium(w *gsm_map.DPAnalysedInfoCriterium) (DPAnalysedInfoCriterium, error) {
	sk := int64(w.ServiceKey)

	dch, err := defaultCallHandlingFromWire(w.DefaultCallHandling)
	if err != nil {
		return DPAnalysedInfoCriterium{}, err
	}
	dnDigits, dnNature, dnPlan, err := decodeAddressField(w.DialledNumber)
	if err != nil {
		return DPAnalysedInfoCriterium{}, fmt.Errorf("decoding DialledNumber: %w", err)
	}
	if dnDigits == "" {
		return DPAnalysedInfoCriterium{}, ErrCamelMissingDialledNumber
	}
	scDigits, scNature, scPlan, err := decodeAddressField(w.GsmSCFAddress)
	if err != nil {
		return DPAnalysedInfoCriterium{}, fmt.Errorf("decoding GsmSCFAddress: %w", err)
	}
	if scDigits == "" {
		return DPAnalysedInfoCriterium{}, ErrCamelMissingGsmSCFAddress
	}
	return DPAnalysedInfoCriterium{
		DialledNumber:       dnDigits,
		DialledNumberNature: dnNature,
		DialledNumberPlan:   dnPlan,
		ServiceKey:          sk,
		GsmSCFAddress:       scDigits,
		GsmSCFAddressNature: scNature,
		GsmSCFAddressPlan:   scPlan,
		DefaultCallHandling: dch,
	}, nil
}

// convertDCSIToWire encodes a D-CSI.
func convertDCSIToWire(d *DCSI) (*gsm_map.DCSI, error) {
	if err := validateCamelCapabilityHandling(d.CamelCapabilityHandling); err != nil {
		return nil, err
	}
	out := &gsm_map.DCSI{}
	if len(d.DPAnalysedInfoCriteriaList) > 0 {
		list := gsm_map.DPAnalysedInfoCriteriaList{Values: make([]gsm_map.DPAnalysedInfoCriterium, len(d.DPAnalysedInfoCriteriaList))}
		for i := range d.DPAnalysedInfoCriteriaList {
			w, err := convertDPAnalysedInfoCriteriumToWire(&d.DPAnalysedInfoCriteriaList[i])
			if err != nil {
				return nil, fmt.Errorf("DPAnalysedInfoCriteriaList[%d]: %w", i, err)
			}
			list.Values[i] = w
		}
		out.DpAnalysedInfoCriteriaList = &list
	}
	if d.CamelCapabilityHandling != nil {
		v := gsm_map.CamelCapabilityHandling(int64(*d.CamelCapabilityHandling))
		out.CamelCapabilityHandling = &v
	}
	out.NotificationToCSE = boolToNullPtr(d.NotificationToCSE)
	out.CsiActive = boolToNullPtr(d.CsiActive)
	return out, nil
}

// convertWireToDCSI decodes a D-CSI.
func convertWireToDCSI(w *gsm_map.DCSI) (*DCSI, error) {
	out := &DCSI{}
	if w.DpAnalysedInfoCriteriaList != nil {
		out.DPAnalysedInfoCriteriaList = make([]DPAnalysedInfoCriterium, len(w.DpAnalysedInfoCriteriaList.Values))
		for i := range w.DpAnalysedInfoCriteriaList.Values {
			c, err := convertWireToDPAnalysedInfoCriterium(&w.DpAnalysedInfoCriteriaList.Values[i])
			if err != nil {
				return nil, fmt.Errorf("DpAnalysedInfoCriteriaList[%d]: %w", i, err)
			}
			out.DPAnalysedInfoCriteriaList[i] = c
		}
	}
	if w.CamelCapabilityHandling != nil {
		out.CamelCapabilityHandling = camelCapabilityHandlingFromWire(*w.CamelCapabilityHandling)
	}
	out.NotificationToCSE = nullPtrToBool(w.NotificationToCSE)
	out.CsiActive = nullPtrToBool(w.CsiActive)
	return out, nil
}

// convertGmscCamelSubInfoToWire converts the public GmscCamelSubscriptionInfo
// to its wire-level representation, including every nested CSI and criteria
// list. Replaces the earlier stub that silently dropped all CAMEL data.
func convertGmscCamelSubInfoToWire(g *GmscCamelSubscriptionInfo) (gsm_map.GmscCamelSubscriptionInfo, error) {
	out := gsm_map.GmscCamelSubscriptionInfo{}
	if g.TCSI != nil {
		t, err := convertTCSIToWire(g.TCSI)
		if err != nil {
			return gsm_map.GmscCamelSubscriptionInfo{}, fmt.Errorf("TCSI: %w", err)
		}
		out.TCSI = t
	}
	if g.OCSI != nil {
		o, err := convertOCSIToWire(g.OCSI)
		if err != nil {
			return gsm_map.GmscCamelSubscriptionInfo{}, fmt.Errorf("OCSI: %w", err)
		}
		out.OCSI = o
	}
	if g.DCSI != nil {
		d, err := convertDCSIToWire(g.DCSI)
		if err != nil {
			return gsm_map.GmscCamelSubscriptionInfo{}, fmt.Errorf("DCSI: %w", err)
		}
		out.DCsi = d
	}
	if len(g.OBcsmCamelTDPCriteriaList) > 0 {
		list := gsm_map.OBcsmCamelTDPCriteriaList{Values: make([]gsm_map.OBcsmCamelTDPCriteria, len(g.OBcsmCamelTDPCriteriaList))}
		for i := range g.OBcsmCamelTDPCriteriaList {
			w, err := convertOBcsmTDPCriteriaToWire(&g.OBcsmCamelTDPCriteriaList[i])
			if err != nil {
				return gsm_map.GmscCamelSubscriptionInfo{}, fmt.Errorf("OBcsmCamelTDPCriteriaList[%d]: %w", i, err)
			}
			list.Values[i] = w
		}
		out.OBcsmCamelTDPCriteriaList = &list
	}
	if len(g.TBcsmCamelTDPCriteriaList) > 0 {
		list := gsm_map.TBCSMCAMELTDPCriteriaList{Values: make([]gsm_map.TBCSMCAMELTDPCriteria, len(g.TBcsmCamelTDPCriteriaList))}
		for i := range g.TBcsmCamelTDPCriteriaList {
			w, err := convertTBcsmTDPCriteriaToWire(&g.TBcsmCamelTDPCriteriaList[i])
			if err != nil {
				return gsm_map.GmscCamelSubscriptionInfo{}, fmt.Errorf("TBcsmCamelTDPCriteriaList[%d]: %w", i, err)
			}
			list.Values[i] = w
		}
		out.TBCSMCAMELTDPCriteriaList = &list
	}
	return out, nil
}

// convertWireToGmscCamelSubInfo converts a wire GmscCamelSubscriptionInfo back
// into the public type. Replaces the earlier stub that silently dropped all
// CAMEL data on decode.
func convertWireToGmscCamelSubInfo(w *gsm_map.GmscCamelSubscriptionInfo) (GmscCamelSubscriptionInfo, error) {
	out := GmscCamelSubscriptionInfo{}
	if w.TCSI != nil {
		t, err := convertWireToTCSI(w.TCSI)
		if err != nil {
			return GmscCamelSubscriptionInfo{}, fmt.Errorf("TCSI: %w", err)
		}
		out.TCSI = t
	}
	if w.OCSI != nil {
		o, err := convertWireToOCSI(w.OCSI)
		if err != nil {
			return GmscCamelSubscriptionInfo{}, fmt.Errorf("OCSI: %w", err)
		}
		out.OCSI = o
	}
	if w.DCsi != nil {
		d, err := convertWireToDCSI(w.DCsi)
		if err != nil {
			return GmscCamelSubscriptionInfo{}, fmt.Errorf("DCSI: %w", err)
		}
		out.DCSI = d
	}
	if w.OBcsmCamelTDPCriteriaList != nil {
		// Absent when the receiver ignores every entry.
		list, err := convertIgnorableWireList("OBcsmCamelTDPCriteriaList", w.OBcsmCamelTDPCriteriaList.Values, convertWireToOBcsmTDPCriteria)
		if err != nil {
			return GmscCamelSubscriptionInfo{}, err
		}
		out.OBcsmCamelTDPCriteriaList = list
	}
	if w.TBCSMCAMELTDPCriteriaList != nil {
		list := make([]TBcsmCamelTDPCriteria, len(w.TBCSMCAMELTDPCriteriaList.Values))
		for i := range w.TBCSMCAMELTDPCriteriaList.Values {
			c, err := convertWireToTBcsmTDPCriteria(&w.TBCSMCAMELTDPCriteriaList.Values[i])
			if err != nil {
				return GmscCamelSubscriptionInfo{}, fmt.Errorf("TBCSMCAMELTDPCriteriaList[%d]: %w", i, err)
			}
			list[i] = c
		}
		out.TBcsmCamelTDPCriteriaList = list
	}
	return out, nil
}

// --- VlrCamelSubscriptionInfo sub-types (MAP-MS-DataTypes.asn:2183) ---

func convertSSCSIToWire(s *SSCSI) (*gsm_map.SSCSI, error) {
	if s.GsmSCFAddress == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}
	addr, err := encodeAddressField(s.GsmSCFAddress, s.GsmSCFNature, s.GsmSCFPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding SS-CSI.GsmSCFAddress: %w", err)
	}
	events := gsm_map.SSEventList{Values: make([]gsm_map.SSCode, len(s.SsEventList))}
	for i, c := range s.SsEventList {
		events.Values[i] = gsm_map.SSCode{byte(c)}
	}
	return &gsm_map.SSCSI{
		SsCamelData: gsm_map.SSCamelData{
			SsEventList:   &events,
			GsmSCFAddress: gsm_map.ISDNAddressString(addr),
		},
		NotificationToCSE: boolToNullPtr(s.NotificationToCSE),
		CsiActive:         boolToNullPtr(s.CsiActive),
	}, nil
}

func convertWireToSSCSI(w *gsm_map.SSCSI) (*SSCSI, error) {
	events := w.SsCamelData.SsEventList

	digits, nat, plan, err := decodeAddressField(w.SsCamelData.GsmSCFAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding SS-CSI.GsmSCFAddress: %w", err)
	}
	if digits == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}
	ssList := make([]SsCode, len(events.Values))
	for i, b := range events.Values {
		// go-asn1 does not enforce SEQUENCE OF element SIZE: https://github.com/gomaja/go-asn1/issues/79.
		if len(b) != 1 {
			return nil, fmt.Errorf("SS-CSI.SsEventList[%d]: SsCode must be 1 octet, got %d", i, len(b))
		}
		ssList[i] = SsCode(b[0])
	}
	return &SSCSI{
		SsEventList:       ssList,
		GsmSCFAddress:     digits,
		GsmSCFNature:      nat,
		GsmSCFPlan:        plan,
		NotificationToCSE: nullPtrToBool(w.NotificationToCSE),
		CsiActive:         nullPtrToBool(w.CsiActive),
	}, nil
}

func convertMCSIToWire(m *MCSI) (*gsm_map.MCSI, error) {
	if m.GsmSCFAddress == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}
	addr, err := encodeAddressField(m.GsmSCFAddress, m.GsmSCFNature, m.GsmSCFPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding M-CSI.GsmSCFAddress: %w", err)
	}
	triggers := gsm_map.MobilityTriggers{Values: make([]gsm_map.MMCode, len(m.MobilityTriggers))}
	for i, b := range m.MobilityTriggers {
		triggers.Values[i] = gsm_map.MMCode{b}
	}
	return &gsm_map.MCSI{
		MobilityTriggers:  &triggers,
		ServiceKey:        gsm_map.ServiceKey(m.ServiceKey),
		GsmSCFAddress:     gsm_map.ISDNAddressString(addr),
		NotificationToCSE: boolToNullPtr(m.NotificationToCSE),
		CsiActive:         boolToNullPtr(m.CsiActive),
	}, nil
}

func convertWireToMCSI(w *gsm_map.MCSI) (*MCSI, error) {
	sk := int64(w.ServiceKey)

	digits, nat, plan, err := decodeAddressField(w.GsmSCFAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding M-CSI.GsmSCFAddress: %w", err)
	}
	if digits == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}
	triggers := make([]byte, len(w.MobilityTriggers.Values))
	for i, mm := range w.MobilityTriggers.Values {
		if len(mm) != 1 {
			return nil, fmt.Errorf("M-CSI.MobilityTriggers[%d]: %w", i, ErrCamelInvalidMobilityTriggerOctet)
		}
		triggers[i] = mm[0]
	}
	return &MCSI{
		MobilityTriggers:  triggers,
		ServiceKey:        sk,
		GsmSCFAddress:     digits,
		GsmSCFNature:      nat,
		GsmSCFPlan:        plan,
		NotificationToCSE: nullPtrToBool(w.NotificationToCSE),
		CsiActive:         nullPtrToBool(w.CsiActive),
	}, nil
}

// The SMS trigger detection point an SMS-CSI or MT-smsCAMELTDP-Criteria
// carries depends on where it sits; 3GPP TS 29.002 V19.1.0 §17.7.1
// SMS-TriggerDetectionPoint has the receiver ignore any other value:
//
//	"If this parameter is received with any other value than
//	sms-CollectedInfo in an SMS-CAMEL-TDP-Data sequence contained in
//	mo-sms-CSI, then the receiver shall ignore the whole SMS-CAMEL-TDP-Data
//	sequence."
//
//	"If this parameter is received with any other value than
//	sms-DeliveryRequest in an SMS-CAMEL-TDP-Data sequence contained in
//	mt-sms-CSI then the receiver shall ignore the whole SMS-CAMEL-TDP-Data
//	sequence."
//
//	"If this parameter is received with any other value than
//	sms-DeliveryRequest in an MT-smsCAMELTDP-Criteria sequence then the
//	receiver shall ignore the whole MT-smsCAMELTDP-Criteria sequence."
//
// These subsume the clause's general rule for values other than the listed
// ones. The encoders accept only the same value, so the wire never carries
// an entry the receiver would drop.
const (
	// moSMSTriggerDetectionPoint is the TDP of every mo-sms-CSI entry.
	moSMSTriggerDetectionPoint = SMSTriggerDetectionPointSmsCollectedInfo
	// mtSMSTriggerDetectionPoint is the TDP of every mt-sms-CSI entry and
	// of every MT-smsCAMELTDP-Criteria.
	mtSMSTriggerDetectionPoint = SMSTriggerDetectionPointSmsDeliveryRequest
)

func isValidDefaultSMSHandling(v DefaultSMSHandling) bool {
	return v == DefaultSMSHandlingContinueTransaction ||
		v == DefaultSMSHandlingReleaseTransaction
}

func isValidMTSMSTPDUType(v MTSMSTPDUType) bool {
	return v == MTSMSTPDUTypeSmsDELIVER ||
		v == MTSMSTPDUTypeSmsSUBMITREPORT ||
		v == MTSMSTPDUTypeSmsSTATUSREPORT
}

// convertSMSCAMELTDPDataToWire encodes one SMS-CAMEL-TDP-Data of an SMS-CSI
// whose only permitted trigger detection point is tdp.
func convertSMSCAMELTDPDataToWire(d *SMSCAMELTDPData, tdp SMSTriggerDetectionPoint) (gsm_map.SMSCAMELTDPData, error) {
	if d.SmsTriggerDetectionPoint != tdp {
		return gsm_map.SMSCAMELTDPData{}, fmt.Errorf("%w (got %d)", ErrCamelInvalidSMSTriggerDetectionPoint, d.SmsTriggerDetectionPoint)
	}

	if d.GsmSCFAddress == "" {
		return gsm_map.SMSCAMELTDPData{}, ErrCamelMissingGsmSCFAddress
	}
	if !isValidDefaultSMSHandling(d.DefaultSMSHandling) {
		return gsm_map.SMSCAMELTDPData{}, ErrCamelInvalidDefaultSMSHandling
	}
	addr, err := encodeAddressField(d.GsmSCFAddress, d.GsmSCFNature, d.GsmSCFPlan)
	if err != nil {
		return gsm_map.SMSCAMELTDPData{}, fmt.Errorf("encoding SMS-CAMEL-TDP-Data.GsmSCFAddress: %w", err)
	}
	return gsm_map.SMSCAMELTDPData{
		SmsTriggerDetectionPoint: d.SmsTriggerDetectionPoint,
		ServiceKey:               gsm_map.ServiceKey(d.ServiceKey),
		GsmSCFAddress:            gsm_map.ISDNAddressString(addr),
		DefaultSMSHandling:       d.DefaultSMSHandling,
	}, nil
}

// convertWireToSMSCAMELTDPData decodes an SMS-CAMEL-TDP-Data entry of an
// SMS-CSI whose only permitted trigger detection point is tdp. It returns
// nil, without examining the other fields, for an entry with any other
// trigger detection point, which the receiver ignores (see
// moSMSTriggerDetectionPoint). The decoder applies the lenient rules
// documented on DefaultSMSHandling (values 2..31 → continueTransaction,
// values >31 → releaseTransaction).
func convertWireToSMSCAMELTDPData(w *gsm_map.SMSCAMELTDPData, tdp SMSTriggerDetectionPoint) (*SMSCAMELTDPData, error) {
	if w.SmsTriggerDetectionPoint != tdp {
		return nil, nil
	}
	sk := int64(w.ServiceKey)

	digits, nat, plan, err := decodeAddressField(w.GsmSCFAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding SMS-CAMEL-TDP-Data.GsmSCFAddress: %w", err)
	}
	if digits == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}
	// DefaultSMSHandling lenient mapping per TS 29.002 §8.8.1:
	//   0 = continueTransaction
	//   1 = releaseTransaction
	//   2..31 → continueTransaction
	//   >31  → releaseTransaction
	// Apply the mapping in int64 space first so a wire value larger than
	// platform int (e.g. on a 32-bit build) still follows the spec's
	// >31 → releaseTransaction rule instead of erroring on the narrow.
	// The post-mapping value is always 0 or 1, which fits in int on every
	// supported platform.
	dsh64 := int64(w.DefaultSMSHandling)
	var dsh DefaultSMSHandling
	switch {
	case dsh64 == 0:
		dsh = DefaultSMSHandlingContinueTransaction
	case dsh64 == 1:
		dsh = DefaultSMSHandlingReleaseTransaction
	case dsh64 >= 2 && dsh64 <= 31:
		dsh = DefaultSMSHandlingContinueTransaction
	case dsh64 > 31:
		dsh = DefaultSMSHandlingReleaseTransaction
	default:
		// Negative values aren't covered by the spec exception clause;
		// reject them.
		return nil, ErrCamelInvalidDefaultSMSHandling
	}
	return &SMSCAMELTDPData{
		SmsTriggerDetectionPoint: w.SmsTriggerDetectionPoint,
		ServiceKey:               sk,
		GsmSCFAddress:            digits,
		GsmSCFNature:             nat,
		GsmSCFPlan:               plan,
		DefaultSMSHandling:       dsh,
	}, nil
}

// convertSMSCSIToWire encodes an SMS-CSI whose only permitted trigger
// detection point is tdp: moSMSTriggerDetectionPoint for mo-sms-CSI,
// mtSMSTriggerDetectionPoint for mt-sms-CSI.
func convertSMSCSIToWire(s *SMSCSI, tdp SMSTriggerDetectionPoint) (*gsm_map.SMSCSI, error) {
	if s.CamelCapabilityHandling == nil {
		return nil, ErrCamelSMSCSIMissingCapabilityHandling
	}
	if err := validateCamelCapabilityHandling(s.CamelCapabilityHandling); err != nil {
		return nil, err
	}
	list := gsm_map.SMSCAMELTDPDataList{Values: make([]gsm_map.SMSCAMELTDPData, len(s.SmsCAMELTDPDataList))}
	for i := range s.SmsCAMELTDPDataList {
		w, err := convertSMSCAMELTDPDataToWire(&s.SmsCAMELTDPDataList[i], tdp)
		if err != nil {
			return nil, fmt.Errorf("SmsCAMELTDPDataList[%d]: %w", i, err)
		}
		list.Values[i] = w
	}
	cch := gsm_map.CamelCapabilityHandling(int64(*s.CamelCapabilityHandling))
	return &gsm_map.SMSCSI{
		SmsCAMELTDPDataList:     &list,
		CamelCapabilityHandling: &cch,
		NotificationToCSE:       boolToNullPtr(s.NotificationToCSE),
		CsiActive:               boolToNullPtr(s.CsiActive),
	}, nil
}

// convertWireToSMSCSI decodes an SMS-CSI whose only permitted trigger
// detection point is tdp: moSMSTriggerDetectionPoint for mo-sms-CSI,
// mtSMSTriggerDetectionPoint for mt-sms-CSI. It returns nil when the
// receiver ignores every SMS-CAMEL-TDP-Data: an SMS-CSI arms its TDP only
// through SMS-CAMEL-TDP-DataList, SIZE (1..10), so with none left the
// receiver holds no SMS-CSI.
func convertWireToSMSCSI(w *gsm_map.SMSCSI, tdp SMSTriggerDetectionPoint) (*SMSCSI, error) {
	if w.SmsCAMELTDPDataList == nil {
		return nil, ErrCamelSMSCSIMissingTDPData
	}
	if w.CamelCapabilityHandling == nil {
		return nil, ErrCamelSMSCSIMissingCapabilityHandling
	}
	decode := func(d *gsm_map.SMSCAMELTDPData) (*SMSCAMELTDPData, error) {
		return convertWireToSMSCAMELTDPData(d, tdp)
	}
	list, err := convertIgnorableWireList("SmsCAMELTDPDataList", w.SmsCAMELTDPDataList.Values, decode)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, nil
	}
	return &SMSCSI{
		SmsCAMELTDPDataList:     list,
		CamelCapabilityHandling: camelCapabilityHandlingFromWire(*w.CamelCapabilityHandling),
		NotificationToCSE:       nullPtrToBool(w.NotificationToCSE),
		CsiActive:               nullPtrToBool(w.CsiActive),
	}, nil
}

func convertMTSmsCAMELTDPCriteriaToWire(c *MTSmsCAMELTDPCriteria) (gsm_map.MTSmsCAMELTDPCriteria, error) {
	if c.SmsTriggerDetectionPoint != mtSMSTriggerDetectionPoint {
		return gsm_map.MTSmsCAMELTDPCriteria{}, fmt.Errorf("%w (got %d)", ErrCamelInvalidSMSTriggerDetectionPoint, c.SmsTriggerDetectionPoint)
	}
	out := gsm_map.MTSmsCAMELTDPCriteria{
		SmsTriggerDetectionPoint: c.SmsTriggerDetectionPoint,
	}
	if c.TpduTypeCriterion != nil {
		tpdu := gsm_map.TPDUTypeCriterion{Values: make([]gsm_map.MTSMSTPDUType, len(c.TpduTypeCriterion))}
		for i, t := range c.TpduTypeCriterion {
			if !isValidMTSMSTPDUType(t) {
				return gsm_map.MTSmsCAMELTDPCriteria{}, fmt.Errorf("TpduTypeCriterion[%d]: %w", i, ErrCamelInvalidMTSMSTPDUType)
			}
			tpdu.Values[i] = t
		}
		out.TpduTypeCriterion = &tpdu
	}
	return out, nil
}

// convertWireToMTSmsCAMELTDPCriteria decodes an MT-smsCAMELTDP-Criteria. It
// returns nil for an entry whose trigger detection point is not
// mtSMSTriggerDetectionPoint, which the receiver ignores.
func convertWireToMTSmsCAMELTDPCriteria(w *gsm_map.MTSmsCAMELTDPCriteria) (*MTSmsCAMELTDPCriteria, error) {
	if w.SmsTriggerDetectionPoint != mtSMSTriggerDetectionPoint {
		return nil, nil
	}
	out := &MTSmsCAMELTDPCriteria{SmsTriggerDetectionPoint: w.SmsTriggerDetectionPoint}
	if w.TpduTypeCriterion != nil {
		// 3GPP TS 29.002 V19.1.0 §17.7.1 MT-SMS-TPDU-Type: "For
		// TPDU-TypeCriterion sequences containing this parameter with any
		// other value than the ones listed above the receiver shall ignore
		// the whole TPDU-TypeCriterion sequence." The criterion is OPTIONAL,
		// so ignoring it leaves the entry without one.
		out.TpduTypeCriterion = slices.Clone(w.TpduTypeCriterion.Values)
		for _, t := range out.TpduTypeCriterion {
			if !isValidMTSMSTPDUType(t) {
				out.TpduTypeCriterion = nil
				break
			}
		}
	}
	return out, nil
}

func convertVlrCamelSubscriptionInfoToWire(v *VlrCamelSubscriptionInfo) (*gsm_map.VlrCamelSubscriptionInfo, error) {
	out := &gsm_map.VlrCamelSubscriptionInfo{}
	if v.OCSI != nil {
		w, err := convertOCSIToWire(v.OCSI)
		if err != nil {
			return nil, fmt.Errorf("OCSI: %w", err)
		}
		out.OCSI = w
	}
	if v.SsCSI != nil {
		w, err := convertSSCSIToWire(v.SsCSI)
		if err != nil {
			return nil, fmt.Errorf("SsCSI: %w", err)
		}
		out.SsCSI = w
	}
	if v.OBcsmCamelTDPCriteriaList != nil {
		list := gsm_map.OBcsmCamelTDPCriteriaList{Values: make([]gsm_map.OBcsmCamelTDPCriteria, len(v.OBcsmCamelTDPCriteriaList))}
		for i := range v.OBcsmCamelTDPCriteriaList {
			w, err := convertOBcsmTDPCriteriaToWire(&v.OBcsmCamelTDPCriteriaList[i])
			if err != nil {
				return nil, fmt.Errorf("OBcsmCamelTDPCriteriaList[%d]: %w", i, err)
			}
			list.Values[i] = w
		}
		out.OBcsmCamelTDPCriteriaList = &list
	}
	out.TifCSI = boolToNullPtr(v.TifCSI)
	if v.MCSI != nil {
		w, err := convertMCSIToWire(v.MCSI)
		if err != nil {
			return nil, fmt.Errorf("MCSI: %w", err)
		}
		out.MCSI = w
	}
	if v.MoSmsCSI != nil {
		w, err := convertSMSCSIToWire(v.MoSmsCSI, moSMSTriggerDetectionPoint)
		if err != nil {
			return nil, fmt.Errorf("MoSmsCSI: %w", err)
		}
		out.MoSmsCSI = w
	}
	if v.VtCSI != nil {
		w, err := convertTCSIToWire(v.VtCSI)
		if err != nil {
			return nil, fmt.Errorf("VtCSI: %w", err)
		}
		out.VtCSI = w
	}
	if v.TBcsmCamelTDPCriteriaList != nil {
		list := gsm_map.TBCSMCAMELTDPCriteriaList{Values: make([]gsm_map.TBCSMCAMELTDPCriteria, len(v.TBcsmCamelTDPCriteriaList))}
		for i := range v.TBcsmCamelTDPCriteriaList {
			w, err := convertTBcsmTDPCriteriaToWire(&v.TBcsmCamelTDPCriteriaList[i])
			if err != nil {
				return nil, fmt.Errorf("TBcsmCamelTDPCriteriaList[%d]: %w", i, err)
			}
			list.Values[i] = w
		}
		out.TBCSMCAMELTDPCriteriaList = &list
	}
	if v.DCSI != nil {
		w, err := convertDCSIToWire(v.DCSI)
		if err != nil {
			return nil, fmt.Errorf("DCSI: %w", err)
		}
		out.DCSI = w
	}
	if v.MtSmsCSI != nil {
		w, err := convertSMSCSIToWire(v.MtSmsCSI, mtSMSTriggerDetectionPoint)
		if err != nil {
			return nil, fmt.Errorf("MtSmsCSI: %w", err)
		}
		out.MtSmsCSI = w
	}
	if v.MtSmsCAMELTDPCriteriaList != nil {
		list := gsm_map.MTSmsCAMELTDPCriteriaList{Values: make([]gsm_map.MTSmsCAMELTDPCriteria, len(v.MtSmsCAMELTDPCriteriaList))}
		for i := range v.MtSmsCAMELTDPCriteriaList {
			w, err := convertMTSmsCAMELTDPCriteriaToWire(&v.MtSmsCAMELTDPCriteriaList[i])
			if err != nil {
				return nil, fmt.Errorf("MtSmsCAMELTDPCriteriaList[%d]: %w", i, err)
			}
			list.Values[i] = w
		}
		out.MtSmsCAMELTDPCriteriaList = &list
	}
	return out, nil
}

func convertWireToVlrCamelSubscriptionInfo(w *gsm_map.VlrCamelSubscriptionInfo) (*VlrCamelSubscriptionInfo, error) {
	out := &VlrCamelSubscriptionInfo{TifCSI: nullPtrToBool(w.TifCSI)}
	if w.OCSI != nil {
		d, err := convertWireToOCSI(w.OCSI)
		if err != nil {
			return nil, fmt.Errorf("OCSI: %w", err)
		}
		out.OCSI = d
	}
	if w.SsCSI != nil {
		d, err := convertWireToSSCSI(w.SsCSI)
		if err != nil {
			return nil, fmt.Errorf("SsCSI: %w", err)
		}
		out.SsCSI = d
	}
	if w.OBcsmCamelTDPCriteriaList != nil {
		// Per spec SIZE(1..10), a non-nil empty wire list is malformed.
		// Match the encoder's strictness.
		// Absent when the receiver ignores every entry.
		list, err := convertIgnorableWireList("OBcsmCamelTDPCriteriaList", w.OBcsmCamelTDPCriteriaList.Values, convertWireToOBcsmTDPCriteria)
		if err != nil {
			return nil, err
		}
		out.OBcsmCamelTDPCriteriaList = list
	}
	if w.MCSI != nil {
		d, err := convertWireToMCSI(w.MCSI)
		if err != nil {
			return nil, fmt.Errorf("MCSI: %w", err)
		}
		out.MCSI = d
	}
	if w.MoSmsCSI != nil {
		d, err := convertWireToSMSCSI(w.MoSmsCSI, moSMSTriggerDetectionPoint)
		if err != nil {
			return nil, fmt.Errorf("MoSmsCSI: %w", err)
		}
		out.MoSmsCSI = d
	}
	if w.VtCSI != nil {
		d, err := convertWireToTCSI(w.VtCSI)
		if err != nil {
			return nil, fmt.Errorf("VtCSI: %w", err)
		}
		out.VtCSI = d
	}
	if w.TBCSMCAMELTDPCriteriaList != nil {
		list := make([]TBcsmCamelTDPCriteria, len(w.TBCSMCAMELTDPCriteriaList.Values))
		for i := range w.TBCSMCAMELTDPCriteriaList.Values {
			d, err := convertWireToTBcsmTDPCriteria(&w.TBCSMCAMELTDPCriteriaList.Values[i])
			if err != nil {
				return nil, fmt.Errorf("TBcsmCamelTDPCriteriaList[%d]: %w", i, err)
			}
			list[i] = d
		}
		out.TBcsmCamelTDPCriteriaList = list
	}
	if w.DCSI != nil {
		d, err := convertWireToDCSI(w.DCSI)
		if err != nil {
			return nil, fmt.Errorf("DCSI: %w", err)
		}
		out.DCSI = d
	}
	if w.MtSmsCSI != nil {
		d, err := convertWireToSMSCSI(w.MtSmsCSI, mtSMSTriggerDetectionPoint)
		if err != nil {
			return nil, fmt.Errorf("MtSmsCSI: %w", err)
		}
		out.MtSmsCSI = d
	}
	if w.MtSmsCAMELTDPCriteriaList != nil {
		// Absent when the receiver ignores every entry.
		list, err := convertIgnorableWireList("MtSmsCAMELTDPCriteriaList", w.MtSmsCAMELTDPCriteriaList.Values, convertWireToMTSmsCAMELTDPCriteria)
		if err != nil {
			return nil, err
		}
		out.MtSmsCAMELTDPCriteriaList = list
	}
	return out, nil
}
