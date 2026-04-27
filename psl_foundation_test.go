// psl_foundation_test.go
//
// Tests for ProvideSubscriberLocation (opCode 83) foundation types.
// PR A of the staged PSL implementation — top-level Arg/Res structs and
// converters land in follow-up PRs.
package gsmmap

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// Compile-smoke: every new public type must be referenceable.
func TestPSLTypesCompile(t *testing.T) {
	var _ LocationEstimateType
	var _ DeferredLocationEventType
	var _ LocationType
	var _ LCSClientType
	var _ LCSFormatIndicator
	var _ LCSClientName
	var _ LCSRequestorID
	var _ LCSClientID
	var _ ResponseTimeCategory
	var _ ResponseTime
	var _ LCSQoS
	var _ PrivacyCheckRelatedAction
	var _ LCSPrivacyCheck
	var _ LCSCodeword
	var _ AccuracyFulfilmentIndicator
	var _ SupportedGADShapes
	var _ LCSPriority
	var _ LCSReferenceNumber

	// Constants exist and resolve to the correct upstream values.
	_ = LocationEstimateCurrentLocation
	_ = LocationEstimateCurrentOrLastKnownLocation
	_ = LocationEstimateInitialLocation
	_ = LocationEstimateActivateDeferredLocation
	_ = LocationEstimateCancelDeferredLocation
	_ = LocationEstimateNotificationVerificationOnly

	_ = LCSClientTypeEmergencyServices
	_ = LCSClientTypeValueAddedServices
	_ = LCSClientTypePlmnOperatorServices
	_ = LCSClientTypeLawfulInterceptServices

	_ = LCSFormatLogicalName
	_ = LCSFormatEMailAddress
	_ = LCSFormatMsisdn
	_ = LCSFormatUrl
	_ = LCSFormatSipUrl

	_ = ResponseTimeLowdelay
	_ = ResponseTimeDelaytolerant

	_ = PrivacyCheckAllowedWithoutNotification
	_ = PrivacyCheckAllowedWithNotification
	_ = PrivacyCheckAllowedIfNoResponse
	_ = PrivacyCheckRestrictedIfNoResponse
	_ = PrivacyCheckNotAllowed

	_ = AccuracyFulfilmentRequestedAccuracyFulfilled
	_ = AccuracyFulfilmentRequestedAccuracyNotFulfilled
}

// Aliased enums must resolve to the same numeric values as upstream so
// callers can use either local or upstream names interchangeably.
func TestPSLEnumsAliasUpstream(t *testing.T) {
	cases := []struct {
		name  string
		local int64
		upstr int64
	}{
		{"LocationEstimateCurrentLocation", int64(LocationEstimateCurrentLocation), int64(gsm_map.LocationEstimateTypeCurrentLocation)},
		{"LocationEstimateNotificationVerificationOnly", int64(LocationEstimateNotificationVerificationOnly), int64(gsm_map.LocationEstimateTypeNotificationVerificationOnly)},
		{"LCSClientTypeEmergencyServices", int64(LCSClientTypeEmergencyServices), int64(gsm_map.LCSClientTypeEmergencyServices)},
		{"LCSClientTypeLawfulInterceptServices", int64(LCSClientTypeLawfulInterceptServices), int64(gsm_map.LCSClientTypeLawfulInterceptServices)},
		{"LCSFormatLogicalName", int64(LCSFormatLogicalName), int64(gsm_map.LCSFormatIndicatorLogicalName)},
		{"LCSFormatSipUrl", int64(LCSFormatSipUrl), int64(gsm_map.LCSFormatIndicatorSipUrl)},
		{"ResponseTimeLowdelay", int64(ResponseTimeLowdelay), int64(gsm_map.ResponseTimeCategoryLowdelay)},
		{"ResponseTimeDelaytolerant", int64(ResponseTimeDelaytolerant), int64(gsm_map.ResponseTimeCategoryDelaytolerant)},
		{"PrivacyCheckAllowedWithoutNotification", int64(PrivacyCheckAllowedWithoutNotification), int64(gsm_map.PrivacyCheckRelatedActionAllowedWithoutNotification)},
		{"PrivacyCheckNotAllowed", int64(PrivacyCheckNotAllowed), int64(gsm_map.PrivacyCheckRelatedActionNotAllowed)},
		{"AccuracyFulfilmentRequestedAccuracyFulfilled", int64(AccuracyFulfilmentRequestedAccuracyFulfilled), int64(gsm_map.AccuracyFulfilmentIndicatorRequestedAccuracyFulfilled)},
		{"AccuracyFulfilmentRequestedAccuracyNotFulfilled", int64(AccuracyFulfilmentRequestedAccuracyNotFulfilled), int64(gsm_map.AccuracyFulfilmentIndicatorRequestedAccuracyNotFulfilled)},
	}
	for _, tc := range cases {
		if tc.local != tc.upstr {
			t.Errorf("%s: local=%d upstream=%d", tc.name, tc.local, tc.upstr)
		}
	}
}

// LCSPriority / LCSReferenceNumber are HexBytes aliases. The alias
// relationship lets HexBytes flow through unchanged; tested by passing
// a HexBytes value through a function whose parameter is typed as the
// alias. Size enforcement happens in the codec (PR D).
func TestPSLByteAliases(t *testing.T) {
	priority := func(v LCSPriority) int { return len(v) }
	if got := priority(HexBytes{0x00}); got != 1 {
		t.Fatalf("LCSPriority alias: want len 1, got %d", got)
	}
	ref := func(v LCSReferenceNumber) int { return len(v) }
	if got := ref(HexBytes{0xff}); got != 1 {
		t.Fatalf("LCSReferenceNumber alias: want len 1, got %d", got)
	}
}

// Sentinel errors must be non-nil, distinct, and detectable through
// errors.Is when wrapped via %w (the typical caller pattern).
func TestPSLSentinelErrors(t *testing.T) {
	sentinels := []error{
		ErrLocationEstimateTypeInvalid,
		ErrLCSClientTypeInvalid,
		ErrLCSFormatIndicatorInvalid,
		ErrPrivacyCheckRelatedActionInvalid,
		ErrAccuracyFulfilmentIndicatorInvalid,
		ErrResponseTimeCategoryInvalid,
		ErrLCSPriorityInvalidSize,
		ErrLCSReferenceNumberInvalidSize,
		ErrHorizontalAccuracyInvalidSize,
		ErrVerticalAccuracyInvalidSize,
		ErrLCSCodewordStringSize,
		ErrLCSClientNameNameStringSize,
		ErrLCSRequestorIDStringSize,
		ErrDeferredLocationEventTypeSize,
		ErrLCSClientIDDialedByMSEmpty,
	}
	seen := make(map[error]int, len(sentinels))
	for i, s := range sentinels {
		if s == nil {
			t.Errorf("sentinel #%d is nil", i)
			continue
		}
		if j, dup := seen[s]; dup {
			t.Errorf("sentinel #%d aliases sentinel #%d (same error value)", i, j)
		}
		seen[s] = i
		wrapped := fmt.Errorf("psl wrapper: %w", s)
		if !errors.Is(wrapped, s) {
			t.Errorf("sentinel #%d not detectable through errors.Is when wrapped with %%w", i)
		}
	}
}

// Spec constants must resolve to the values defined in TS 29.002.
func TestPSLSpecConstants(t *testing.T) {
	if LCSCodewordStringMaxLen != 20 {
		t.Errorf("LCSCodewordStringMaxLen: want 20 per maxLCSCodewordStringLength, got %d", LCSCodewordStringMaxLen)
	}
	if NameStringMaxLen != 63 {
		t.Errorf("NameStringMaxLen: want 63 per maxNameStringLength, got %d", NameStringMaxLen)
	}
	if RequestorIDStringMaxLen != 63 {
		t.Errorf("RequestorIDStringMaxLen: want 63 per maxRequestorIDStringLength, got %d", RequestorIDStringMaxLen)
	}
}

// Foundation struct shapes must be zero-value safe so the public API
// can be constructed incrementally before the codec lands.
func TestPSLZeroValues(t *testing.T) {
	var lt LocationType
	if lt.LocationEstimateType != 0 {
		t.Error("LocationType zero value should have LocationEstimateType=0")
	}
	if lt.DeferredLocationEventType != nil {
		t.Error("LocationType zero value should have nil DeferredLocationEventType")
	}

	var qos LCSQoS
	if qos.HorizontalAccuracy != nil {
		t.Error("LCSQoS zero value should have nil HorizontalAccuracy")
	}
	if qos.ResponseTime != nil {
		t.Error("LCSQoS zero value should have nil ResponseTime")
	}
	if qos.VerticalCoordinateRequest || qos.VelocityRequest {
		t.Error("LCSQoS zero value should have NULL flags = false")
	}

	var id LCSClientID
	if id.LcsClientType != 0 {
		t.Error("LCSClientID zero value should have LcsClientType=0")
	}
	if id.LcsClientDialedByMS != "" {
		t.Error("LCSClientID zero value should have empty LcsClientDialedByMS digits")
	}

	var d DeferredLocationEventType
	if d.MsAvailable || d.EnteringIntoArea || d.LeavingFromArea || d.BeingInsideArea || d.PeriodicLDR {
		t.Error("DeferredLocationEventType zero value should have all bits false")
	}

	var g SupportedGADShapes
	if g.EllipsoidPoint || g.EllipsoidArc {
		t.Error("SupportedGADShapes zero value should have all bits false")
	}
}
