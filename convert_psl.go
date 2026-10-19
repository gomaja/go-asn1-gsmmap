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

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime"
)

// ============================================================================
// BIT STRING surrogates
// ============================================================================

// DeferredLocationEventType (BIT STRING SIZE 1..16, 5 named bits) per
// 3GPP TS 29.002 V19.1.0 §17.7.13.
//
// go-asn1 chooses the shortest permitted length (ITU-T X.680 (02/2021) §22.7).
func convertDeferredLocationEventTypeToBitString(d *DeferredLocationEventType) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.DeferredLocationEventTypeMsAvailable:      d.MsAvailable,
		gsm_map.DeferredLocationEventTypeEnteringIntoArea: d.EnteringIntoArea,
		gsm_map.DeferredLocationEventTypeLeavingFromArea:  d.LeavingFromArea,
		gsm_map.DeferredLocationEventTypeBeingInsideArea:  d.BeingInsideArea,
		gsm_map.DeferredLocationEventTypePeriodicLDR:      d.PeriodicLDR,
	}
	return gsm_map.NewDeferredLocationEventType(namedBitPositions(bits)...)
}

// convertBitStringToDeferredLocationEventType decodes the 5 named bits and
// ignores the others. A ProvideSubscriberLocation-Arg setting another bit is
// rejected first (hasUnlistedDeferredLocationEvent).
func convertBitStringToDeferredLocationEventType(bs runtime.BitString) *DeferredLocationEventType {
	return &DeferredLocationEventType{
		MsAvailable:      bs.Has(gsm_map.DeferredLocationEventTypeMsAvailable),
		EnteringIntoArea: bs.Has(gsm_map.DeferredLocationEventTypeEnteringIntoArea),
		LeavingFromArea:  bs.Has(gsm_map.DeferredLocationEventTypeLeavingFromArea),
		BeingInsideArea:  bs.Has(gsm_map.DeferredLocationEventTypeBeingInsideArea),
		PeriodicLDR:      bs.Has(gsm_map.DeferredLocationEventTypePeriodicLDR),
	}
}

// hasUnlistedDeferredLocationEvent reports whether bs sets a bit past
// periodicLDR(4), the last value 3GPP TS 29.002 V19.1.0 §17.7.13
// DeferredLocationEventType lists.
func hasUnlistedDeferredLocationEvent(bs runtime.BitString) bool {
	for bit := gsm_map.DeferredLocationEventTypePeriodicLDR + 1; bit < bs.BitLength; bit++ {
		if bs.Has(bit) {
			return true
		}
	}
	return false
}

// SupportedGADShapes (BIT STRING SIZE 7..16, 7 named bits) per 3GPP TS 29.002 V19.1.0 §17.7.13.
//
// go-asn1 chooses the shortest permitted length (ITU-T X.680 (02/2021) §22.7).
func convertSupportedGADShapesToBitString(g *SupportedGADShapes) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.SupportedGADShapesEllipsoidPoint:                                   g.EllipsoidPoint,
		gsm_map.SupportedGADShapesEllipsoidPointWithUncertaintyCircle:              g.EllipsoidPointWithUncertaintyCircle,
		gsm_map.SupportedGADShapesEllipsoidPointWithUncertaintyEllipse:             g.EllipsoidPointWithUncertaintyEllipse,
		gsm_map.SupportedGADShapesPolygon:                                          g.Polygon,
		gsm_map.SupportedGADShapesEllipsoidPointWithAltitude:                       g.EllipsoidPointWithAltitude,
		gsm_map.SupportedGADShapesEllipsoidPointWithAltitudeAndUncertaintyElipsoid: g.EllipsoidPointWithAltitudeAndUncertaintyEllipsoid,
		gsm_map.SupportedGADShapesEllipsoidArc:                                     g.EllipsoidArc,
	}
	return gsm_map.NewSupportedGADShapes(namedBitPositions(bits)...)
}

// convertBitStringToSupportedGADShapes decodes the 7 named bits.
// Bits past them are tolerated and ignored on decode.
func convertBitStringToSupportedGADShapes(bs runtime.BitString) *SupportedGADShapes {
	g := &SupportedGADShapes{}
	g.EllipsoidPoint = bs.Has(gsm_map.SupportedGADShapesEllipsoidPoint)
	g.EllipsoidPointWithUncertaintyCircle = bs.Has(gsm_map.SupportedGADShapesEllipsoidPointWithUncertaintyCircle)
	g.EllipsoidPointWithUncertaintyEllipse = bs.Has(gsm_map.SupportedGADShapesEllipsoidPointWithUncertaintyEllipse)
	g.Polygon = bs.Has(gsm_map.SupportedGADShapesPolygon)
	g.EllipsoidPointWithAltitude = bs.Has(gsm_map.SupportedGADShapesEllipsoidPointWithAltitude)
	g.EllipsoidPointWithAltitudeAndUncertaintyEllipsoid = bs.Has(gsm_map.SupportedGADShapesEllipsoidPointWithAltitudeAndUncertaintyElipsoid)
	g.EllipsoidArc = bs.Has(gsm_map.SupportedGADShapesEllipsoidArc)
	return g
}

// ============================================================================
// LocationType — 3GPP TS 29.002 V19.1.0 §17.7.13
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
		bs, err := convertDeferredLocationEventTypeToBitString(l.DeferredLocationEventType)
		if err != nil {
			return nil, fmt.Errorf("encoding DeferredLocationEventType: %w", err)
		}
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
// LCSCodeword — 3GPP TS 29.002 V19.1.0 §17.7.13
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
// LCSPrivacyCheck — 3GPP TS 29.002 V19.1.0 §17.7.13
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
// ResponseTime — 3GPP TS 29.002 V19.1.0 §17.7.13
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
	// Per 3GPP TS 29.002 V19.1.0 §17.7.13, an unrecognized value
	// shall be treated the same as delaytolerant(1).
	if cat != ResponseTimeLowdelay && cat != ResponseTimeDelaytolerant {
		cat = ResponseTimeDelaytolerant
	}
	return &ResponseTime{
		ResponseTimeCategory: cat,
	}
}

// ============================================================================
// LCSQoS — 3GPP TS 29.002 V19.1.0 §17.7.13
// ============================================================================

func convertLCSQoSToWire(q *LCSQoS) (*gsm_map.LCSQoS, error) {
	if q == nil {
		return nil, nil
	}
	out := &gsm_map.LCSQoS{}

	if len(q.HorizontalAccuracy) > 0 {
		// Spec mandates bit 8 = 0 (3GPP TS 29.002 V19.1.0 §17.7.13):
		// only the low 7 bits encode the uncertainty code per TS 23.032.
		if q.HorizontalAccuracy[0]&0x80 != 0 {
			return nil, fmt.Errorf("LCSQoS.HorizontalAccuracy=0x%02x: %w", q.HorizontalAccuracy[0], ErrHorizontalAccuracyReservedBit)
		}
		v := gsm_map.HorizontalAccuracy(q.HorizontalAccuracy)
		out.HorizontalAccuracy = &v
	}
	out.VerticalCoordinateRequest = boolToNullPtr(q.VerticalCoordinateRequest)
	if len(q.VerticalAccuracy) > 0 {
		// Spec mandates bit 8 = 0 (3GPP TS 29.002 V19.1.0 §17.7.13):
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
	if q.LcsQosClass != nil {
		c := *q.LcsQosClass
		if c != LCSQoSClassBestEffort && c != LCSQoSClassAssured {
			return nil, fmt.Errorf("LCSQoS.LcsQosClass=%d: %w", c, ErrLCSQoSClassInvalid)
		}
		out.LcsQosClass = &c
	}
	return out, nil
}

func convertWireToLCSQoS(w *gsm_map.LCSQoS) (*LCSQoS, error) {
	if w == nil {
		return nil, nil
	}
	out := &LCSQoS{}
	if w.HorizontalAccuracy != nil {
		out.HorizontalAccuracy = HexBytes(*w.HorizontalAccuracy)
	}
	out.VerticalCoordinateRequest = nullPtrToBool(w.VerticalCoordinateRequest)
	if w.VerticalAccuracy != nil {
		out.VerticalAccuracy = HexBytes(*w.VerticalAccuracy)
	}
	if w.ResponseTime != nil {
		out.ResponseTime = convertWireToResponseTime(w.ResponseTime)
	}
	out.VelocityRequest = nullPtrToBool(w.VelocityRequest)
	if w.LcsQosClass != nil {
		// 3GPP TS 29.002 V19.1.0 §17.7.13 LCS-QoS-Class: "an unrecognized
		// value shall be treated the same as value 0 (bestEffort)".
		c := *w.LcsQosClass
		if c != LCSQoSClassAssured {
			c = LCSQoSClassBestEffort
		}
		out.LcsQosClass = &c
	}
	return out, nil
}
