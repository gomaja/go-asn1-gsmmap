package gsmmap

import (
	"fmt"

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

// validateCamelCapabilityHandling enforces the 1..4 phase range.
func validateCamelCapabilityHandling(p *int) error {
	if p == nil {
		return nil
	}
	if *p < 1 || *p > 4 {
		return ErrCamelInvalidCamelCapabilityHandling
	}
	return nil
}

// convertOBcsmTDPDataToWire encodes a single O-BCSM TDP entry.
func convertOBcsmTDPDataToWire(d *OBcsmCamelTDPData) (gsm_map.OBcsmCamelTDPData, error) {
	if !isValidOBcsmTDP(d.OBcsmTriggerDetectionPoint) {
		return gsm_map.OBcsmCamelTDPData{}, ErrCamelInvalidOTriggerPoint
	}
	if d.ServiceKey < 0 || d.ServiceKey > 2147483647 {
		return gsm_map.OBcsmCamelTDPData{}, ErrCamelInvalidServiceKey
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

// convertWireToOBcsmTDPData decodes a single wire O-BCSM TDP entry.
// Mirrors the encoder's range/mandatory checks to reject malformed peer
// input rather than silently surfacing an invalid struct to the caller.
func convertWireToOBcsmTDPData(w *gsm_map.OBcsmCamelTDPData) (OBcsmCamelTDPData, error) {
	tdp := OBcsmTriggerDetectionPoint(w.OBcsmTriggerDetectionPoint)
	if !isValidOBcsmTDP(tdp) {
		return OBcsmCamelTDPData{}, ErrCamelInvalidOTriggerPoint
	}
	sk := int64(w.ServiceKey)
	if sk < 0 || sk > 2147483647 {
		return OBcsmCamelTDPData{}, ErrCamelInvalidServiceKey
	}
	dch := DefaultCallHandling(w.DefaultCallHandling)
	if !isValidDefaultCallHandling(dch) {
		return OBcsmCamelTDPData{}, ErrCamelInvalidDefaultCallHandling
	}
	digits, nature, plan, err := decodeAddressField(w.GsmSCFAddress)
	if err != nil {
		return OBcsmCamelTDPData{}, fmt.Errorf("decoding GsmSCFAddress: %w", err)
	}
	if digits == "" {
		return OBcsmCamelTDPData{}, ErrCamelMissingGsmSCFAddress
	}
	return OBcsmCamelTDPData{
		OBcsmTriggerDetectionPoint: tdp,
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
	if d.ServiceKey < 0 || d.ServiceKey > 2147483647 {
		return gsm_map.TBcsmCamelTDPData{}, ErrCamelInvalidServiceKey
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

// convertWireToTBcsmTDPData decodes a single wire T-BCSM TDP entry.
func convertWireToTBcsmTDPData(w *gsm_map.TBcsmCamelTDPData) (TBcsmCamelTDPData, error) {
	tdp := TBcsmTriggerDetectionPoint(w.TBcsmTriggerDetectionPoint)
	if !isValidTBcsmTDP(tdp) {
		return TBcsmCamelTDPData{}, ErrCamelInvalidTTriggerPoint
	}
	sk := int64(w.ServiceKey)
	if sk < 0 || sk > 2147483647 {
		return TBcsmCamelTDPData{}, ErrCamelInvalidServiceKey
	}
	dch := DefaultCallHandling(w.DefaultCallHandling)
	if !isValidDefaultCallHandling(dch) {
		return TBcsmCamelTDPData{}, ErrCamelInvalidDefaultCallHandling
	}
	digits, nature, plan, err := decodeAddressField(w.GsmSCFAddress)
	if err != nil {
		return TBcsmCamelTDPData{}, fmt.Errorf("decoding GsmSCFAddress: %w", err)
	}
	if digits == "" {
		return TBcsmCamelTDPData{}, ErrCamelMissingGsmSCFAddress
	}
	return TBcsmCamelTDPData{
		TBcsmTriggerDetectionPoint: tdp,
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
		list := make(gsm_map.DestinationNumberList, len(c.DestinationNumberList))
		for i, n := range c.DestinationNumberList {
			if n.Digits == "" {
				return nil, fmt.Errorf("DestinationNumberList[%d]: %w", i, ErrCamelMissingDestinationNumber)
			}
			enc, err := encodeAddressField(n.Digits, n.Nature, n.Plan)
			if err != nil {
				return nil, fmt.Errorf("DestinationNumberList[%d]: %w", i, err)
			}
			list[i] = gsm_map.ISDNAddressString(enc)
		}
		out.DestinationNumberList = list
	}
	if len(c.DestinationNumberLengthList) > 0 {
		list := make(gsm_map.DestinationNumberLengthList, len(c.DestinationNumberLengthList))
		for i, l := range c.DestinationNumberLengthList {
			if l < 1 || l > 15 {
				return nil, fmt.Errorf("DestinationNumberLengthList[%d]: %w", i, ErrCamelInvalidDestinationNumberLength)
			}
			list[i] = int64(l)
		}
		out.DestinationNumberLengthList = list
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
	if len(w.DestinationNumberList) == 0 && len(w.DestinationNumberLengthList) == 0 {
		return nil, ErrCamelMissingDestinationNumberCriteria
	}
	out := &DestinationNumberCriteria{MatchType: mt}
	if len(w.DestinationNumberList) > 0 {
		list := make([]ISDNNumber, len(w.DestinationNumberList))
		for i, n := range w.DestinationNumberList {
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
	if len(w.DestinationNumberLengthList) > 0 {
		list := make([]int, len(w.DestinationNumberLengthList))
		for i, l := range w.DestinationNumberLengthList {
			if l < 1 || l > 15 {
				return nil, fmt.Errorf("DestinationNumberLengthList[%d]: %w", i, ErrCamelInvalidDestinationNumberLength)
			}
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
		bsc := make(gsm_map.BasicServiceCriteria, len(c.BasicServiceCriteria))
		for i := range c.BasicServiceCriteria {
			wv, err := convertExtBasicServiceCodeToWire(&c.BasicServiceCriteria[i])
			if err != nil {
				return gsm_map.OBcsmCamelTDPCriteria{}, fmt.Errorf("BasicServiceCriteria[%d]: %w", i, err)
			}
			bsc[i] = *wv
		}
		out.BasicServiceCriteria = bsc
	}
	if c.CallTypeCriteria != nil {
		if !isValidCallTypeCriteria(*c.CallTypeCriteria) {
			return gsm_map.OBcsmCamelTDPCriteria{}, ErrCamelInvalidCallTypeCriteria
		}
		ctc := *c.CallTypeCriteria
		out.CallTypeCriteria = &ctc
	}
	if len(c.OCauseValueCriteria) > 0 {
		if len(c.OCauseValueCriteria) > 5 {
			return gsm_map.OBcsmCamelTDPCriteria{}, ErrCamelInvalidCauseValueListSize
		}
		list := make(gsm_map.OCauseValueCriteria, len(c.OCauseValueCriteria))
		for i, v := range c.OCauseValueCriteria {
			if v < 0 || v > 127 {
				return gsm_map.OBcsmCamelTDPCriteria{}, fmt.Errorf("OCauseValueCriteria[%d]: %w", i, ErrCamelInvalidCauseValue)
			}
			list[i] = gsm_map.CauseValue{byte(v)}
		}
		out.OCauseValueCriteria = list
	}
	return out, nil
}

// convertWireToOBcsmTDPCriteria decodes an O-BCSM TDP criteria entry.
func convertWireToOBcsmTDPCriteria(w *gsm_map.OBcsmCamelTDPCriteria) (OBcsmCamelTDPCriteria, error) {
	tdp := OBcsmTriggerDetectionPoint(w.OBcsmTriggerDetectionPoint)
	if !isValidOBcsmTDP(tdp) {
		return OBcsmCamelTDPCriteria{}, ErrCamelInvalidOTriggerPoint
	}
	out := OBcsmCamelTDPCriteria{OBcsmTriggerDetectionPoint: tdp}
	if w.DestinationNumberCriteria != nil {
		dnc, err := convertWireToDestinationNumberCriteria(w.DestinationNumberCriteria)
		if err != nil {
			return OBcsmCamelTDPCriteria{}, fmt.Errorf("DestinationNumberCriteria: %w", err)
		}
		out.DestinationNumberCriteria = dnc
	}
	if len(w.BasicServiceCriteria) > 0 {
		bsc := make([]ExtBasicServiceCode, len(w.BasicServiceCriteria))
		for i := range w.BasicServiceCriteria {
			pv, err := convertWireToExtBasicServiceCode(&w.BasicServiceCriteria[i])
			if err != nil {
				return OBcsmCamelTDPCriteria{}, fmt.Errorf("BasicServiceCriteria[%d]: %w", i, err)
			}
			bsc[i] = *pv
		}
		out.BasicServiceCriteria = bsc
	}
	if w.CallTypeCriteria != nil {
		ctc := CallTypeCriteria(*w.CallTypeCriteria)
		if !isValidCallTypeCriteria(ctc) {
			return OBcsmCamelTDPCriteria{}, ErrCamelInvalidCallTypeCriteria
		}
		out.CallTypeCriteria = &ctc
	}
	if w.OCauseValueCriteria != nil {
		// OCauseValueCriteria is SIZE(1..5) per maxNumOfCAMEL-O-CauseValueCriteria;
		// a non-nil empty slice means the tag was on the wire with zero elements,
		// which violates the lower bound.
		if len(w.OCauseValueCriteria) < 1 || len(w.OCauseValueCriteria) > 5 {
			return OBcsmCamelTDPCriteria{}, ErrCamelInvalidCauseValueListSize
		}
		list := make([]int, len(w.OCauseValueCriteria))
		for i, b := range w.OCauseValueCriteria {
			// CauseValue is OCTET STRING (SIZE(1)); reject any other length
			// rather than silently normalising missing/extra octets.
			if len(b) != 1 {
				return OBcsmCamelTDPCriteria{}, fmt.Errorf("OCauseValueCriteria[%d]: %w", i, ErrCamelInvalidCauseValueOctetLength)
			}
			v := int(b[0])
			if v > 127 {
				return OBcsmCamelTDPCriteria{}, fmt.Errorf("OCauseValueCriteria[%d]: %w", i, ErrCamelInvalidCauseValue)
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
		bsc := make(gsm_map.BasicServiceCriteria, len(c.BasicServiceCriteria))
		for i := range c.BasicServiceCriteria {
			wv, err := convertExtBasicServiceCodeToWire(&c.BasicServiceCriteria[i])
			if err != nil {
				return gsm_map.TBCSMCAMELTDPCriteria{}, fmt.Errorf("BasicServiceCriteria[%d]: %w", i, err)
			}
			bsc[i] = *wv
		}
		out.BasicServiceCriteria = bsc
	}
	if len(c.TCauseValueCriteria) > 0 {
		if len(c.TCauseValueCriteria) > 5 {
			return gsm_map.TBCSMCAMELTDPCriteria{}, ErrCamelInvalidCauseValueListSize
		}
		list := make(gsm_map.TCauseValueCriteria, len(c.TCauseValueCriteria))
		for i, v := range c.TCauseValueCriteria {
			if v < 0 || v > 127 {
				return gsm_map.TBCSMCAMELTDPCriteria{}, fmt.Errorf("TCauseValueCriteria[%d]: %w", i, ErrCamelInvalidCauseValue)
			}
			list[i] = gsm_map.CauseValue{byte(v)}
		}
		out.TCauseValueCriteria = list
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
	if len(w.BasicServiceCriteria) > 0 {
		bsc := make([]ExtBasicServiceCode, len(w.BasicServiceCriteria))
		for i := range w.BasicServiceCriteria {
			pv, err := convertWireToExtBasicServiceCode(&w.BasicServiceCriteria[i])
			if err != nil {
				return TBcsmCamelTDPCriteria{}, fmt.Errorf("BasicServiceCriteria[%d]: %w", i, err)
			}
			bsc[i] = *pv
		}
		out.BasicServiceCriteria = bsc
	}
	if w.TCauseValueCriteria != nil {
		if len(w.TCauseValueCriteria) < 1 || len(w.TCauseValueCriteria) > 5 {
			return TBcsmCamelTDPCriteria{}, ErrCamelInvalidCauseValueListSize
		}
		list := make([]int, len(w.TCauseValueCriteria))
		for i, b := range w.TCauseValueCriteria {
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
	if len(o.OBcsmCamelTDPDataList) < 1 || len(o.OBcsmCamelTDPDataList) > 10 {
		return nil, ErrCamelInvalidTDPDataListSize
	}
	if err := validateCamelCapabilityHandling(o.CamelCapabilityHandling); err != nil {
		return nil, err
	}
	list := make(gsm_map.OBcsmCamelTDPDataList, len(o.OBcsmCamelTDPDataList))
	for i := range o.OBcsmCamelTDPDataList {
		w, err := convertOBcsmTDPDataToWire(&o.OBcsmCamelTDPDataList[i])
		if err != nil {
			return nil, fmt.Errorf("OBcsmCamelTDPDataList[%d]: %w", i, err)
		}
		list[i] = w
	}
	out := &gsm_map.OCSI{OBcsmCamelTDPDataList: list}
	if o.CamelCapabilityHandling != nil {
		v := gsm_map.CamelCapabilityHandling(int64(*o.CamelCapabilityHandling))
		out.CamelCapabilityHandling = &v
	}
	out.NotificationToCSE = boolToNullPtr(o.NotificationToCSE)
	out.CsiActive = boolToNullPtr(o.CsiActive)
	return out, nil
}

// convertWireToOCSI decodes a wire O-CSI.
func convertWireToOCSI(w *gsm_map.OCSI) (*OCSI, error) {
	if len(w.OBcsmCamelTDPDataList) < 1 || len(w.OBcsmCamelTDPDataList) > 10 {
		return nil, ErrCamelInvalidTDPDataListSize
	}
	out := &OCSI{
		OBcsmCamelTDPDataList: make([]OBcsmCamelTDPData, len(w.OBcsmCamelTDPDataList)),
	}
	for i := range w.OBcsmCamelTDPDataList {
		d, err := convertWireToOBcsmTDPData(&w.OBcsmCamelTDPDataList[i])
		if err != nil {
			return nil, fmt.Errorf("OBcsmCamelTDPDataList[%d]: %w", i, err)
		}
		out.OBcsmCamelTDPDataList[i] = d
	}
	if w.CamelCapabilityHandling != nil {
		// Range-check the wire int64 before narrowing to Go int so 32-bit
		// builds can't truncate out-of-range values into the valid 1..4
		// window.
		v64 := int64(*w.CamelCapabilityHandling)
		if v64 < 1 || v64 > 4 {
			return nil, ErrCamelInvalidCamelCapabilityHandling
		}
		v := int(v64)
		out.CamelCapabilityHandling = &v
	}
	out.NotificationToCSE = nullPtrToBool(w.NotificationToCSE)
	out.CsiActive = nullPtrToBool(w.CsiActive)
	return out, nil
}

// convertTCSIToWire encodes a T-CSI.
func convertTCSIToWire(t *TCSI) (*gsm_map.TCSI, error) {
	if len(t.TBcsmCamelTDPDataList) < 1 || len(t.TBcsmCamelTDPDataList) > 10 {
		return nil, ErrCamelInvalidTDPDataListSize
	}
	if err := validateCamelCapabilityHandling(t.CamelCapabilityHandling); err != nil {
		return nil, err
	}
	list := make(gsm_map.TBcsmCamelTDPDataList, len(t.TBcsmCamelTDPDataList))
	for i := range t.TBcsmCamelTDPDataList {
		w, err := convertTBcsmTDPDataToWire(&t.TBcsmCamelTDPDataList[i])
		if err != nil {
			return nil, fmt.Errorf("TBcsmCamelTDPDataList[%d]: %w", i, err)
		}
		list[i] = w
	}
	out := &gsm_map.TCSI{TBcsmCamelTDPDataList: list}
	if t.CamelCapabilityHandling != nil {
		v := gsm_map.CamelCapabilityHandling(int64(*t.CamelCapabilityHandling))
		out.CamelCapabilityHandling = &v
	}
	out.NotificationToCSE = boolToNullPtr(t.NotificationToCSE)
	out.CsiActive = boolToNullPtr(t.CsiActive)
	return out, nil
}

// convertWireToTCSI decodes a wire T-CSI.
func convertWireToTCSI(w *gsm_map.TCSI) (*TCSI, error) {
	if len(w.TBcsmCamelTDPDataList) < 1 || len(w.TBcsmCamelTDPDataList) > 10 {
		return nil, ErrCamelInvalidTDPDataListSize
	}
	out := &TCSI{
		TBcsmCamelTDPDataList: make([]TBcsmCamelTDPData, len(w.TBcsmCamelTDPDataList)),
	}
	for i := range w.TBcsmCamelTDPDataList {
		d, err := convertWireToTBcsmTDPData(&w.TBcsmCamelTDPDataList[i])
		if err != nil {
			return nil, fmt.Errorf("TBcsmCamelTDPDataList[%d]: %w", i, err)
		}
		out.TBcsmCamelTDPDataList[i] = d
	}
	if w.CamelCapabilityHandling != nil {
		v64 := int64(*w.CamelCapabilityHandling)
		if v64 < 1 || v64 > 4 {
			return nil, ErrCamelInvalidCamelCapabilityHandling
		}
		v := int(v64)
		out.CamelCapabilityHandling = &v
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
	if c.ServiceKey < 0 || c.ServiceKey > 2147483647 {
		return gsm_map.DPAnalysedInfoCriterium{}, ErrCamelInvalidServiceKey
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
	if sk < 0 || sk > 2147483647 {
		return DPAnalysedInfoCriterium{}, ErrCamelInvalidServiceKey
	}
	dch := DefaultCallHandling(w.DefaultCallHandling)
	if !isValidDefaultCallHandling(dch) {
		return DPAnalysedInfoCriterium{}, ErrCamelInvalidDefaultCallHandling
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
		if len(d.DPAnalysedInfoCriteriaList) > 10 {
			return nil, ErrCamelInvalidDPAnalysedInfoListSize
		}
		list := make(gsm_map.DPAnalysedInfoCriteriaList, len(d.DPAnalysedInfoCriteriaList))
		for i := range d.DPAnalysedInfoCriteriaList {
			w, err := convertDPAnalysedInfoCriteriumToWire(&d.DPAnalysedInfoCriteriaList[i])
			if err != nil {
				return nil, fmt.Errorf("DPAnalysedInfoCriteriaList[%d]: %w", i, err)
			}
			list[i] = w
		}
		out.DpAnalysedInfoCriteriaList = list
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
		if len(w.DpAnalysedInfoCriteriaList) < 1 || len(w.DpAnalysedInfoCriteriaList) > 10 {
			return nil, ErrCamelInvalidDPAnalysedInfoListSize
		}
		out.DPAnalysedInfoCriteriaList = make([]DPAnalysedInfoCriterium, len(w.DpAnalysedInfoCriteriaList))
		for i := range w.DpAnalysedInfoCriteriaList {
			c, err := convertWireToDPAnalysedInfoCriterium(&w.DpAnalysedInfoCriteriaList[i])
			if err != nil {
				return nil, fmt.Errorf("DpAnalysedInfoCriteriaList[%d]: %w", i, err)
			}
			out.DPAnalysedInfoCriteriaList[i] = c
		}
	}
	if w.CamelCapabilityHandling != nil {
		v64 := int64(*w.CamelCapabilityHandling)
		if v64 < 1 || v64 > 4 {
			return nil, ErrCamelInvalidCamelCapabilityHandling
		}
		v := int(v64)
		out.CamelCapabilityHandling = &v
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
		if len(g.OBcsmCamelTDPCriteriaList) > 10 {
			return gsm_map.GmscCamelSubscriptionInfo{}, ErrCamelInvalidCriteriaListSize
		}
		list := make(gsm_map.OBcsmCamelTDPCriteriaList, len(g.OBcsmCamelTDPCriteriaList))
		for i := range g.OBcsmCamelTDPCriteriaList {
			w, err := convertOBcsmTDPCriteriaToWire(&g.OBcsmCamelTDPCriteriaList[i])
			if err != nil {
				return gsm_map.GmscCamelSubscriptionInfo{}, fmt.Errorf("OBcsmCamelTDPCriteriaList[%d]: %w", i, err)
			}
			list[i] = w
		}
		out.OBcsmCamelTDPCriteriaList = list
	}
	if len(g.TBcsmCamelTDPCriteriaList) > 0 {
		if len(g.TBcsmCamelTDPCriteriaList) > 10 {
			return gsm_map.GmscCamelSubscriptionInfo{}, ErrCamelInvalidCriteriaListSize
		}
		list := make(gsm_map.TBCSMCAMELTDPCriteriaList, len(g.TBcsmCamelTDPCriteriaList))
		for i := range g.TBcsmCamelTDPCriteriaList {
			w, err := convertTBcsmTDPCriteriaToWire(&g.TBcsmCamelTDPCriteriaList[i])
			if err != nil {
				return gsm_map.GmscCamelSubscriptionInfo{}, fmt.Errorf("TBcsmCamelTDPCriteriaList[%d]: %w", i, err)
			}
			list[i] = w
		}
		out.TBCSMCAMELTDPCriteriaList = list
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
		if len(w.OBcsmCamelTDPCriteriaList) < 1 || len(w.OBcsmCamelTDPCriteriaList) > 10 {
			return GmscCamelSubscriptionInfo{}, ErrCamelInvalidCriteriaListSize
		}
		list := make([]OBcsmCamelTDPCriteria, len(w.OBcsmCamelTDPCriteriaList))
		for i := range w.OBcsmCamelTDPCriteriaList {
			c, err := convertWireToOBcsmTDPCriteria(&w.OBcsmCamelTDPCriteriaList[i])
			if err != nil {
				return GmscCamelSubscriptionInfo{}, fmt.Errorf("OBcsmCamelTDPCriteriaList[%d]: %w", i, err)
			}
			list[i] = c
		}
		out.OBcsmCamelTDPCriteriaList = list
	}
	if w.TBCSMCAMELTDPCriteriaList != nil {
		if len(w.TBCSMCAMELTDPCriteriaList) < 1 || len(w.TBCSMCAMELTDPCriteriaList) > 10 {
			return GmscCamelSubscriptionInfo{}, ErrCamelInvalidCriteriaListSize
		}
		list := make([]TBcsmCamelTDPCriteria, len(w.TBCSMCAMELTDPCriteriaList))
		for i := range w.TBCSMCAMELTDPCriteriaList {
			c, err := convertWireToTBcsmTDPCriteria(&w.TBCSMCAMELTDPCriteriaList[i])
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

// maxNumOfCamelTDPData is the spec upper bound on CAMEL TDP data lists
// and related criteria lists (O-BCSM / T-BCSM / SMS).
const maxNumOfCamelTDPData = 10

// maxNumOfCamelSSEvents is the spec upper bound on SS-EventList.
const maxNumOfCamelSSEvents = 10

// maxNumOfMobilityTriggers is the spec upper bound on MobilityTriggers.
const maxNumOfMobilityTriggers = 10

// maxNumOfMTSmsCamelCriteria is the spec upper bound on MT-smsCAMELTDP-CriteriaList.
const maxNumOfMTSmsCamelCriteria = 5

// maxNumOfTPDUTypes is the spec upper bound on TPDUTypeCriterion.
const maxNumOfTPDUTypes = 5

func convertSSCSIToWire(s *SSCSI) (*gsm_map.SSCSI, error) {
	if len(s.SsEventList) < 1 || len(s.SsEventList) > maxNumOfCamelSSEvents {
		return nil, ErrCamelInvalidSSEventListSize
	}
	if s.GsmSCFAddress == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}
	addr, err := encodeAddressField(s.GsmSCFAddress, s.GsmSCFNature, s.GsmSCFPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding SS-CSI.GsmSCFAddress: %w", err)
	}
	events := make(gsm_map.SSEventList, len(s.SsEventList))
	for i, c := range s.SsEventList {
		events[i] = gsm_map.SSCode{byte(c)}
	}
	return &gsm_map.SSCSI{
		SsCamelData: gsm_map.SSCamelData{
			SsEventList:   events,
			GsmSCFAddress: gsm_map.ISDNAddressString(addr),
		},
		NotificationToCSE: boolToNullPtr(s.NotificationToCSE),
		CsiActive:         boolToNullPtr(s.CsiActive),
	}, nil
}

func convertWireToSSCSI(w *gsm_map.SSCSI) (*SSCSI, error) {
	events := w.SsCamelData.SsEventList
	if len(events) < 1 || len(events) > maxNumOfCamelSSEvents {
		return nil, ErrCamelInvalidSSEventListSize
	}
	digits, nat, plan, err := decodeAddressField(w.SsCamelData.GsmSCFAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding SS-CSI.GsmSCFAddress: %w", err)
	}
	if digits == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}
	ssList := make([]SsCode, len(events))
	for i, b := range events {
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
	if len(m.MobilityTriggers) < 1 || len(m.MobilityTriggers) > maxNumOfMobilityTriggers {
		return nil, ErrCamelInvalidMobilityTriggersSize
	}
	if m.ServiceKey < 0 || m.ServiceKey > 2147483647 {
		return nil, ErrCamelInvalidServiceKey
	}
	if m.GsmSCFAddress == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}
	addr, err := encodeAddressField(m.GsmSCFAddress, m.GsmSCFNature, m.GsmSCFPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding M-CSI.GsmSCFAddress: %w", err)
	}
	triggers := make(gsm_map.MobilityTriggers, len(m.MobilityTriggers))
	for i, b := range m.MobilityTriggers {
		triggers[i] = gsm_map.MMCode{b}
	}
	return &gsm_map.MCSI{
		MobilityTriggers:  triggers,
		ServiceKey:        gsm_map.ServiceKey(m.ServiceKey),
		GsmSCFAddress:     gsm_map.ISDNAddressString(addr),
		NotificationToCSE: boolToNullPtr(m.NotificationToCSE),
		CsiActive:         boolToNullPtr(m.CsiActive),
	}, nil
}

func convertWireToMCSI(w *gsm_map.MCSI) (*MCSI, error) {
	if len(w.MobilityTriggers) < 1 || len(w.MobilityTriggers) > maxNumOfMobilityTriggers {
		return nil, ErrCamelInvalidMobilityTriggersSize
	}
	sk := int64(w.ServiceKey)
	if sk < 0 || sk > 2147483647 {
		return nil, ErrCamelInvalidServiceKey
	}
	digits, nat, plan, err := decodeAddressField(w.GsmSCFAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding M-CSI.GsmSCFAddress: %w", err)
	}
	if digits == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}
	triggers := make([]byte, len(w.MobilityTriggers))
	for i, mm := range w.MobilityTriggers {
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

func isValidSMSTriggerDetectionPoint(v SMSTriggerDetectionPoint) bool {
	return v == SMSTriggerDetectionPointSmsCollectedInfo ||
		v == SMSTriggerDetectionPointSmsDeliveryRequest
}

func isValidDefaultSMSHandling(v DefaultSMSHandling) bool {
	return v == DefaultSMSHandlingContinueTransaction ||
		v == DefaultSMSHandlingReleaseTransaction
}

func isValidMTSMSTPDUType(v MTSMSTPDUType) bool {
	return v == MTSMSTPDUTypeSmsDELIVER ||
		v == MTSMSTPDUTypeSmsSUBMITREPORT ||
		v == MTSMSTPDUTypeSmsSTATUSREPORT
}

func convertSMSCAMELTDPDataToWire(d *SMSCAMELTDPData) (gsm_map.SMSCAMELTDPData, error) {
	if !isValidSMSTriggerDetectionPoint(d.SmsTriggerDetectionPoint) {
		return gsm_map.SMSCAMELTDPData{}, ErrCamelInvalidSMSTriggerDetectionPoint
	}
	if d.ServiceKey < 0 || d.ServiceKey > 2147483647 {
		return gsm_map.SMSCAMELTDPData{}, ErrCamelInvalidServiceKey
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

// convertWireToSMSCAMELTDPData decodes an SMS-CAMEL-TDP-Data entry.
// Per spec exception handling, the decoder applies the lenient rules
// documented on DefaultSMSHandling (values 2..31 → continueTransaction,
// values >31 → releaseTransaction). An unknown SmsTriggerDetectionPoint
// is treated strictly: it returns an error and the caller (the list
// converter convertWireToSMSCSI) propagates it via fmt.Errorf, rejecting
// the entire SMS-CSI sequence rather than silently dropping the entry.
func convertWireToSMSCAMELTDPData(w *gsm_map.SMSCAMELTDPData) (*SMSCAMELTDPData, error) {
	// Narrow the int64 wire enum into Go int with overflow detection so
	// crafted values like 1+2^32 can't wrap to a valid trigger on 32-bit
	// builds before isValidSMSTriggerDetectionPoint runs.
	tdpRaw, err := narrowInt64(int64(w.SmsTriggerDetectionPoint))
	if err != nil {
		return nil, fmt.Errorf("SmsTriggerDetectionPoint: %w", err)
	}
	tdp := SMSTriggerDetectionPoint(tdpRaw)
	if !isValidSMSTriggerDetectionPoint(tdp) {
		return nil, ErrCamelInvalidSMSTriggerDetectionPoint
	}
	sk := int64(w.ServiceKey)
	if sk < 0 || sk > 2147483647 {
		return nil, ErrCamelInvalidServiceKey
	}
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
		SmsTriggerDetectionPoint: tdp,
		ServiceKey:               sk,
		GsmSCFAddress:            digits,
		GsmSCFNature:             nat,
		GsmSCFPlan:               plan,
		DefaultSMSHandling:       dsh,
	}, nil
}

func convertSMSCSIToWire(s *SMSCSI) (*gsm_map.SMSCSI, error) {
	// Spec 8.8.1: both SmsCAMELTDPDataList and CamelCapabilityHandling
	// shall be present in an SMS-CSI sequence. Encoder enforces that
	// and distinguishes "missing" (a §8.8.1 violation) from "oversize"
	// (a SIZE(1..10) violation).
	if len(s.SmsCAMELTDPDataList) < 1 {
		return nil, ErrCamelSMSCSIMissingTDPData
	}
	if len(s.SmsCAMELTDPDataList) > maxNumOfCamelTDPData {
		return nil, ErrCamelInvalidSMSTDPDataListSize
	}
	if s.CamelCapabilityHandling == nil {
		return nil, ErrCamelSMSCSIMissingCapabilityHandling
	}
	if err := validateCamelCapabilityHandling(s.CamelCapabilityHandling); err != nil {
		return nil, err
	}
	list := make(gsm_map.SMSCAMELTDPDataList, len(s.SmsCAMELTDPDataList))
	for i := range s.SmsCAMELTDPDataList {
		w, err := convertSMSCAMELTDPDataToWire(&s.SmsCAMELTDPDataList[i])
		if err != nil {
			return nil, fmt.Errorf("SmsCAMELTDPDataList[%d]: %w", i, err)
		}
		list[i] = w
	}
	cch := gsm_map.CamelCapabilityHandling(int64(*s.CamelCapabilityHandling))
	return &gsm_map.SMSCSI{
		SmsCAMELTDPDataList:     list,
		CamelCapabilityHandling: &cch,
		NotificationToCSE:       boolToNullPtr(s.NotificationToCSE),
		CsiActive:               boolToNullPtr(s.CsiActive),
	}, nil
}

func convertWireToSMSCSI(w *gsm_map.SMSCSI) (*SMSCSI, error) {
	if len(w.SmsCAMELTDPDataList) < 1 {
		return nil, ErrCamelSMSCSIMissingTDPData
	}
	if len(w.SmsCAMELTDPDataList) > maxNumOfCamelTDPData {
		return nil, ErrCamelInvalidSMSTDPDataListSize
	}
	if w.CamelCapabilityHandling == nil {
		return nil, ErrCamelSMSCSIMissingCapabilityHandling
	}
	v64 := int64(*w.CamelCapabilityHandling)
	if v64 < 1 || v64 > 4 {
		return nil, ErrCamelInvalidCamelCapabilityHandling
	}
	cch := int(v64)
	list := make([]SMSCAMELTDPData, len(w.SmsCAMELTDPDataList))
	for i := range w.SmsCAMELTDPDataList {
		d, err := convertWireToSMSCAMELTDPData(&w.SmsCAMELTDPDataList[i])
		if err != nil {
			return nil, fmt.Errorf("SmsCAMELTDPDataList[%d]: %w", i, err)
		}
		list[i] = *d
	}
	return &SMSCSI{
		SmsCAMELTDPDataList:     list,
		CamelCapabilityHandling: &cch,
		NotificationToCSE:       nullPtrToBool(w.NotificationToCSE),
		CsiActive:               nullPtrToBool(w.CsiActive),
	}, nil
}

func convertMTSmsCAMELTDPCriteriaToWire(c *MTSmsCAMELTDPCriteria) (gsm_map.MTSmsCAMELTDPCriteria, error) {
	if !isValidSMSTriggerDetectionPoint(c.SmsTriggerDetectionPoint) {
		return gsm_map.MTSmsCAMELTDPCriteria{}, ErrCamelInvalidSMSTriggerDetectionPoint
	}
	out := gsm_map.MTSmsCAMELTDPCriteria{
		SmsTriggerDetectionPoint: c.SmsTriggerDetectionPoint,
	}
	if c.TpduTypeCriterion != nil {
		if len(c.TpduTypeCriterion) < 1 || len(c.TpduTypeCriterion) > maxNumOfTPDUTypes {
			return gsm_map.MTSmsCAMELTDPCriteria{}, ErrCamelInvalidTPDUTypeCriterionSize
		}
		tpdu := make(gsm_map.TPDUTypeCriterion, len(c.TpduTypeCriterion))
		for i, t := range c.TpduTypeCriterion {
			if !isValidMTSMSTPDUType(t) {
				return gsm_map.MTSmsCAMELTDPCriteria{}, fmt.Errorf("TpduTypeCriterion[%d]: %w", i, ErrCamelInvalidMTSMSTPDUType)
			}
			tpdu[i] = t
		}
		out.TpduTypeCriterion = tpdu
	}
	return out, nil
}

func convertWireToMTSmsCAMELTDPCriteria(w *gsm_map.MTSmsCAMELTDPCriteria) (*MTSmsCAMELTDPCriteria, error) {
	tdpRaw, err := narrowInt64(int64(w.SmsTriggerDetectionPoint))
	if err != nil {
		return nil, fmt.Errorf("SmsTriggerDetectionPoint: %w", err)
	}
	tdp := SMSTriggerDetectionPoint(tdpRaw)
	if !isValidSMSTriggerDetectionPoint(tdp) {
		return nil, ErrCamelInvalidSMSTriggerDetectionPoint
	}
	out := &MTSmsCAMELTDPCriteria{SmsTriggerDetectionPoint: tdp}
	if w.TpduTypeCriterion != nil {
		if len(w.TpduTypeCriterion) < 1 || len(w.TpduTypeCriterion) > maxNumOfTPDUTypes {
			return nil, ErrCamelInvalidTPDUTypeCriterionSize
		}
		tpdu := make([]MTSMSTPDUType, len(w.TpduTypeCriterion))
		for i, t := range w.TpduTypeCriterion {
			ttRaw, err := narrowInt64(int64(t))
			if err != nil {
				return nil, fmt.Errorf("TpduTypeCriterion[%d]: %w", i, err)
			}
			tt := MTSMSTPDUType(ttRaw)
			if !isValidMTSMSTPDUType(tt) {
				return nil, fmt.Errorf("TpduTypeCriterion[%d]: %w", i, ErrCamelInvalidMTSMSTPDUType)
			}
			tpdu[i] = tt
		}
		out.TpduTypeCriterion = tpdu
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
		// Spec SIZE(1..10) — reject a non-nil empty list rather than
		// silently omitting it. Matches the PR #29 pattern.
		if len(v.OBcsmCamelTDPCriteriaList) < 1 || len(v.OBcsmCamelTDPCriteriaList) > maxNumOfCamelTDPData {
			return nil, ErrCamelInvalidCriteriaListSize
		}
		list := make(gsm_map.OBcsmCamelTDPCriteriaList, len(v.OBcsmCamelTDPCriteriaList))
		for i := range v.OBcsmCamelTDPCriteriaList {
			w, err := convertOBcsmTDPCriteriaToWire(&v.OBcsmCamelTDPCriteriaList[i])
			if err != nil {
				return nil, fmt.Errorf("OBcsmCamelTDPCriteriaList[%d]: %w", i, err)
			}
			list[i] = w
		}
		out.OBcsmCamelTDPCriteriaList = list
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
		w, err := convertSMSCSIToWire(v.MoSmsCSI)
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
		if len(v.TBcsmCamelTDPCriteriaList) < 1 || len(v.TBcsmCamelTDPCriteriaList) > maxNumOfCamelTDPData {
			return nil, ErrCamelInvalidCriteriaListSize
		}
		list := make(gsm_map.TBCSMCAMELTDPCriteriaList, len(v.TBcsmCamelTDPCriteriaList))
		for i := range v.TBcsmCamelTDPCriteriaList {
			w, err := convertTBcsmTDPCriteriaToWire(&v.TBcsmCamelTDPCriteriaList[i])
			if err != nil {
				return nil, fmt.Errorf("TBcsmCamelTDPCriteriaList[%d]: %w", i, err)
			}
			list[i] = w
		}
		out.TBCSMCAMELTDPCriteriaList = list
	}
	if v.DCSI != nil {
		w, err := convertDCSIToWire(v.DCSI)
		if err != nil {
			return nil, fmt.Errorf("DCSI: %w", err)
		}
		out.DCSI = w
	}
	if v.MtSmsCSI != nil {
		w, err := convertSMSCSIToWire(v.MtSmsCSI)
		if err != nil {
			return nil, fmt.Errorf("MtSmsCSI: %w", err)
		}
		out.MtSmsCSI = w
	}
	if v.MtSmsCAMELTDPCriteriaList != nil {
		if len(v.MtSmsCAMELTDPCriteriaList) < 1 || len(v.MtSmsCAMELTDPCriteriaList) > maxNumOfMTSmsCamelCriteria {
			return nil, ErrCamelInvalidMTSmsCAMELCriteriaSize
		}
		list := make(gsm_map.MTSmsCAMELTDPCriteriaList, len(v.MtSmsCAMELTDPCriteriaList))
		for i := range v.MtSmsCAMELTDPCriteriaList {
			w, err := convertMTSmsCAMELTDPCriteriaToWire(&v.MtSmsCAMELTDPCriteriaList[i])
			if err != nil {
				return nil, fmt.Errorf("MtSmsCAMELTDPCriteriaList[%d]: %w", i, err)
			}
			list[i] = w
		}
		out.MtSmsCAMELTDPCriteriaList = list
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
		// Match the encoder's strictness (and the PR #29 pattern).
		if len(w.OBcsmCamelTDPCriteriaList) < 1 || len(w.OBcsmCamelTDPCriteriaList) > maxNumOfCamelTDPData {
			return nil, ErrCamelInvalidCriteriaListSize
		}
		list := make([]OBcsmCamelTDPCriteria, len(w.OBcsmCamelTDPCriteriaList))
		for i := range w.OBcsmCamelTDPCriteriaList {
			d, err := convertWireToOBcsmTDPCriteria(&w.OBcsmCamelTDPCriteriaList[i])
			if err != nil {
				return nil, fmt.Errorf("OBcsmCamelTDPCriteriaList[%d]: %w", i, err)
			}
			list[i] = d
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
		d, err := convertWireToSMSCSI(w.MoSmsCSI)
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
		if len(w.TBCSMCAMELTDPCriteriaList) < 1 || len(w.TBCSMCAMELTDPCriteriaList) > maxNumOfCamelTDPData {
			return nil, ErrCamelInvalidCriteriaListSize
		}
		list := make([]TBcsmCamelTDPCriteria, len(w.TBCSMCAMELTDPCriteriaList))
		for i := range w.TBCSMCAMELTDPCriteriaList {
			d, err := convertWireToTBcsmTDPCriteria(&w.TBCSMCAMELTDPCriteriaList[i])
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
		d, err := convertWireToSMSCSI(w.MtSmsCSI)
		if err != nil {
			return nil, fmt.Errorf("MtSmsCSI: %w", err)
		}
		out.MtSmsCSI = d
	}
	if w.MtSmsCAMELTDPCriteriaList != nil {
		if len(w.MtSmsCAMELTDPCriteriaList) < 1 || len(w.MtSmsCAMELTDPCriteriaList) > maxNumOfMTSmsCamelCriteria {
			return nil, ErrCamelInvalidMTSmsCAMELCriteriaSize
		}
		list := make([]MTSmsCAMELTDPCriteria, len(w.MtSmsCAMELTDPCriteriaList))
		for i := range w.MtSmsCAMELTDPCriteriaList {
			d, err := convertWireToMTSmsCAMELTDPCriteria(&w.MtSmsCAMELTDPCriteriaList[i])
			if err != nil {
				return nil, fmt.Errorf("MtSmsCAMELTDPCriteriaList[%d]: %w", i, err)
			}
			list[i] = *d
		}
		out.MtSmsCAMELTDPCriteriaList = list
	}
	return out, nil
}
