// convert_psl.go
//
// Converters for ProvideSubscriberLocation (opCode 83) leaf SEQUENCE
// types and BIT STRING surrogates. The container converters (LCSClientID,
// AreaEventInfo, PeriodicLDRInfo, ReportingPLMNList) and the top-level
// ProvideSubscriberLocationArg/Res live in the other convert_psl_*.go files.
//
// Converters for semantic sender rules return the sentinels in gsmmap.go.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/runtime"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// BIT STRING surrogates
// ============================================================================

// DeferredLocationEventType (BIT STRING SIZE 1..16, 5 named bits) per
// TS 29.002 MAP-LCS-DataTypes.asn:165.
//
// Encode rule: BitLength is the position of the highest set bit + 1
// (minimum 1 to satisfy the SIZE 1..16 lower bound).
func convertDeferredLocationEventTypeToBitString(d *DeferredLocationEventType) runtime.BitString {
	var b byte
	bitLen := 1
	if d.MsAvailable {
		b |= 0x80
	}
	if d.EnteringIntoArea {
		b |= 0x40
		if bitLen < 2 {
			bitLen = 2
		}
	}
	if d.LeavingFromArea {
		b |= 0x20
		if bitLen < 3 {
			bitLen = 3
		}
	}
	if d.BeingInsideArea {
		b |= 0x10
		if bitLen < 4 {
			bitLen = 4
		}
	}
	if d.PeriodicLDR {
		b |= 0x08
		if bitLen < 5 {
			bitLen = 5
		}
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: bitLen}
}

// convertBitStringToDeferredLocationEventType decodes the 5 named bits and
// ignores the others. A ProvideSubscriberLocation-Arg setting another bit is
// rejected first (hasUnlistedDeferredLocationEvent).
func convertBitStringToDeferredLocationEventType(bs runtime.BitString) *DeferredLocationEventType {
	return &DeferredLocationEventType{
		MsAvailable:      bs.Has(0),
		EnteringIntoArea: bs.Has(1),
		LeavingFromArea:  bs.Has(2),
		BeingInsideArea:  bs.Has(3),
		PeriodicLDR:      bs.Has(4),
	}
}

// hasUnlistedDeferredLocationEvent reports whether bs sets a bit past
// periodicLDR(4), the last value 3GPP TS 29.002 V19.1.0 §17.7.13
// DeferredLocationEventType lists.
func hasUnlistedDeferredLocationEvent(bs runtime.BitString) bool {
	for bit := 5; bit < bs.BitLength; bit++ {
		if bs.Has(bit) {
			return true
		}
	}
	return false
}

// SupportedGADShapes (BIT STRING SIZE 7..16, 7 named bits) per TS 29.002
// MAP-LCS-DataTypes.asn:280.
//
// Encode rule: always emit 7 bits to satisfy the SIZE 7..16 lower bound,
// even when no flag is set.
func convertSupportedGADShapesToBitString(g *SupportedGADShapes) runtime.BitString {
	var b byte
	if g.EllipsoidPoint {
		b |= 0x80
	}
	if g.EllipsoidPointWithUncertaintyCircle {
		b |= 0x40
	}
	if g.EllipsoidPointWithUncertaintyEllipse {
		b |= 0x20
	}
	if g.Polygon {
		b |= 0x10
	}
	if g.EllipsoidPointWithAltitude {
		b |= 0x08
	}
	if g.EllipsoidPointWithAltitudeAndUncertaintyEllipsoid {
		b |= 0x04
	}
	if g.EllipsoidArc {
		b |= 0x02
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: 7}
}

// convertBitStringToSupportedGADShapes decodes the 7 named bits.
// Bits past them are tolerated and ignored on decode.
func convertBitStringToSupportedGADShapes(bs runtime.BitString) *SupportedGADShapes {
	g := &SupportedGADShapes{}
	g.EllipsoidPoint = bs.Has(0)
	g.EllipsoidPointWithUncertaintyCircle = bs.Has(1)
	g.EllipsoidPointWithUncertaintyEllipse = bs.Has(2)
	g.Polygon = bs.Has(3)
	g.EllipsoidPointWithAltitude = bs.Has(4)
	g.EllipsoidPointWithAltitudeAndUncertaintyEllipsoid = bs.Has(5)
	g.EllipsoidArc = bs.Has(6)
	return g
}

// ============================================================================
// LocationType — TS 29.002 MAP-LCS-DataTypes.asn:148
// ============================================================================

// isRecognizedLocationEstimateType reports whether v is one of the
// LocationEstimateType values 3GPP TS 29.002 V19.1.0 §17.7.13 lists,
// currentLocation(0) to notificationVerificationOnly(5).
func isRecognizedLocationEstimateType(v LocationEstimateType) bool {
	return v >= LocationEstimateCurrentLocation && v <= LocationEstimateNotificationVerificationOnly
}

func convertLocationTypeToWire(l *LocationType) (*gsm_map.LocationType, error) {
	if l == nil {
		return nil, nil
	}
	out := &gsm_map.LocationType{
		LocationEstimateType: l.LocationEstimateType,
	}
	if !isRecognizedLocationEstimateType(l.LocationEstimateType) {
		return nil, fmt.Errorf("LocationType.LocationEstimateType=%d: %w", l.LocationEstimateType, ErrLocationEstimateTypeInvalid)
	}
	if l.DeferredLocationEventType != nil {
		bs := convertDeferredLocationEventTypeToBitString(l.DeferredLocationEventType)
		out.DeferredLocationEventType = &bs
	}
	return out, nil
}

func convertWireToLocationType(w *gsm_map.LocationType) *LocationType {
	if w == nil {
		return nil
	}
	out := &LocationType{
		LocationEstimateType: w.LocationEstimateType,
	}
	if w.DeferredLocationEventType != nil {
		out.DeferredLocationEventType = convertBitStringToDeferredLocationEventType(*w.DeferredLocationEventType)
	}
	return out
}

// ============================================================================
// LCSCodeword — TS 29.002 MAP-LCS-DataTypes.asn:293
// ============================================================================

func convertLCSCodewordToWire(c *LCSCodeword) *gsm_map.LCSCodeword {
	if c == nil {
		return nil
	}

	out := &gsm_map.LCSCodeword{
		DataCodingScheme:  gsm_map.USSDDataCodingScheme{byte(c.DataCodingScheme)},
		LcsCodewordString: gsm_map.LCSCodewordString(c.LcsCodewordString),
	}
	return out
}

func convertWireToLCSCodeword(w *gsm_map.LCSCodeword) *LCSCodeword {
	if w == nil {
		return nil
	}
	dcs := USSDDataCodingScheme(w.DataCodingScheme[0])
	return &LCSCodeword{
		DataCodingScheme:  dcs,
		LcsCodewordString: HexBytes(w.LcsCodewordString),
	}
}

// ============================================================================
// LCSPrivacyCheck — TS 29.002 MAP-LCS-DataTypes.asn:302
// ============================================================================

// isRecognizedPrivacyCheckRelatedAction reports whether v is one of the
// PrivacyCheckRelatedAction values 3GPP TS 29.002 V19.1.0 §17.7.13 lists,
// allowedWithoutNotification(0) to notAllowed(4).
func isRecognizedPrivacyCheckRelatedAction(v PrivacyCheckRelatedAction) bool {
	return v >= PrivacyCheckAllowedWithoutNotification && v <= PrivacyCheckNotAllowed
}

func convertLCSPrivacyCheckToWire(p *LCSPrivacyCheck) (*gsm_map.LCSPrivacyCheck, error) {
	if p == nil {
		return nil, nil
	}
	if !isRecognizedPrivacyCheckRelatedAction(p.CallSessionUnrelated) {
		return nil, fmt.Errorf("LCSPrivacyCheck.CallSessionUnrelated=%d: %w", p.CallSessionUnrelated, ErrPrivacyCheckRelatedActionInvalid)
	}
	out := &gsm_map.LCSPrivacyCheck{
		CallSessionUnrelated: p.CallSessionUnrelated,
	}
	if p.CallSessionRelated != nil {
		v := *p.CallSessionRelated
		if !isRecognizedPrivacyCheckRelatedAction(v) {
			return nil, fmt.Errorf("LCSPrivacyCheck.CallSessionRelated=%d: %w", v, ErrPrivacyCheckRelatedActionInvalid)
		}
		out.CallSessionRelated = &v
	}
	return out, nil
}

// convertWireToLCSPrivacyCheck copies both PrivacyCheckRelatedActions. The
// ProvideSubscriberLocation-Arg decoder rejects an unrecognized one
// (ErrPrivacyCheckRelatedActionUnrecognized).
func convertWireToLCSPrivacyCheck(w *gsm_map.LCSPrivacyCheck) *LCSPrivacyCheck {
	if w == nil {
		return nil
	}
	out := &LCSPrivacyCheck{
		CallSessionUnrelated: w.CallSessionUnrelated,
	}
	if w.CallSessionRelated != nil {
		v := *w.CallSessionRelated
		out.CallSessionRelated = &v
	}
	return out
}

// ============================================================================
// ResponseTime — TS 29.002 MAP-LCS-DataTypes.asn:261
// ============================================================================
//
// ResponseTimeCategory is an extensible ENUMERATED with a spec exception
// clause: unrecognized values shall be treated as delaytolerant(1) on
// decode. Decoder applies the exception clause; encoder is strict
// (lowdelay or delaytolerant only).

func convertResponseTimeToWire(r *ResponseTime) (*gsm_map.ResponseTime, error) {
	if r == nil {
		return nil, nil
	}
	if r.ResponseTimeCategory != ResponseTimeLowdelay && r.ResponseTimeCategory != ResponseTimeDelaytolerant {
		return nil, fmt.Errorf("ResponseTime.ResponseTimeCategory=%d: %w", r.ResponseTimeCategory, ErrResponseTimeCategoryInvalid)
	}
	return &gsm_map.ResponseTime{
		ResponseTimeCategory: r.ResponseTimeCategory,
	}, nil
}

func convertWireToResponseTime(w *gsm_map.ResponseTime) *ResponseTime {
	if w == nil {
		return nil
	}
	cat := w.ResponseTimeCategory
	// Per TS 29.002 MAP-LCS-DataTypes.asn:270-271, an unrecognized value
	// shall be treated the same as delaytolerant(1).
	if cat != ResponseTimeLowdelay && cat != ResponseTimeDelaytolerant {
		cat = ResponseTimeDelaytolerant
	}
	return &ResponseTime{
		ResponseTimeCategory: cat,
	}
}

// ============================================================================
// LCSQoS — TS 29.002 MAP-LCS-DataTypes.asn:237
// ============================================================================

func convertLCSQoSToWire(q *LCSQoS) (*gsm_map.LCSQoS, error) {
	if q == nil {
		return nil, nil
	}
	out := &gsm_map.LCSQoS{}

	if len(q.HorizontalAccuracy) > 0 {
		// Spec mandates bit 8 = 0 (TS 29.002 MAP-LCS-DataTypes.asn:250):
		// only the low 7 bits encode the uncertainty code per TS 23.032.
		if q.HorizontalAccuracy[0]&0x80 != 0 {
			return nil, fmt.Errorf("LCSQoS.HorizontalAccuracy=0x%02x: %w", q.HorizontalAccuracy[0], ErrHorizontalAccuracyReservedBit)
		}
		v := gsm_map.HorizontalAccuracy(q.HorizontalAccuracy)
		out.HorizontalAccuracy = &v
	}
	out.VerticalCoordinateRequest = boolToNullPtr(q.VerticalCoordinateRequest)
	if len(q.VerticalAccuracy) > 0 {
		// Spec mandates bit 8 = 0 (TS 29.002 MAP-LCS-DataTypes.asn:256):
		// only the low 7 bits encode the vertical uncertainty code per TS 23.032.
		if q.VerticalAccuracy[0]&0x80 != 0 {
			return nil, fmt.Errorf("LCSQoS.VerticalAccuracy=0x%02x: %w", q.VerticalAccuracy[0], ErrVerticalAccuracyReservedBit)
		}
		v := gsm_map.VerticalAccuracy(q.VerticalAccuracy)
		out.VerticalAccuracy = &v
	}
	if q.ResponseTime != nil {
		rt, err := convertResponseTimeToWire(q.ResponseTime)
		if err != nil {
			return nil, fmt.Errorf("LCSQoS.ResponseTime: %w", err)
		}
		out.ResponseTime = rt
	}
	out.VelocityRequest = boolToNullPtr(q.VelocityRequest)
	return out, nil
}

func convertWireToLCSQoS(w *gsm_map.LCSQoS) (*LCSQoS, error) {
	if w == nil {
		return nil, nil
	}
	out := &LCSQoS{}
	if w.HorizontalAccuracy != nil {
		if (*w.HorizontalAccuracy)[0]&0x80 != 0 {
			return nil, fmt.Errorf("LCSQoS.HorizontalAccuracy=0x%02x: %w", (*w.HorizontalAccuracy)[0], ErrHorizontalAccuracyReservedBit)
		}
		out.HorizontalAccuracy = HexBytes(*w.HorizontalAccuracy)
	}
	out.VerticalCoordinateRequest = nullPtrToBool(w.VerticalCoordinateRequest)
	if w.VerticalAccuracy != nil {
		if (*w.VerticalAccuracy)[0]&0x80 != 0 {
			return nil, fmt.Errorf("LCSQoS.VerticalAccuracy=0x%02x: %w", (*w.VerticalAccuracy)[0], ErrVerticalAccuracyReservedBit)
		}
		out.VerticalAccuracy = HexBytes(*w.VerticalAccuracy)
	}
	if w.ResponseTime != nil {
		out.ResponseTime = convertWireToResponseTime(w.ResponseTime)
	}
	out.VelocityRequest = nullPtrToBool(w.VelocityRequest)
	return out, nil
}
