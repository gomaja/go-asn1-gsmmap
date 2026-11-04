// convert_map_error.go
//
// Converters between the MAP ReturnError wire parameters and the public
// diagnostic types defined in gsmmap.go. The
// ParseReturnErrorParameter dispatcher lives in parse_map_error.go.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// AbsentSubscriberSMParam — TS 29.002 §17.7.7 (errorCode 6)
// ============================================================================

func convertWireToAbsentSubscriberSMParam(w *gsm_map.AbsentSubscriberSMParam) (*AbsentSubscriberSMParam, error) {
	if w == nil {
		return nil, nil
	}
	out := &AbsentSubscriberSMParam{}
	// The upstream AbsentSubscriberDiagnosticSM is a plain int64; the public
	// field uses the named type, which has a String method.
	if w.AbsentSubscriberDiagnosticSM != nil {
		v := AbsentSubscriberDiagnosticSM(*w.AbsentSubscriberDiagnosticSM)
		out.AbsentSubscriberDiagnosticSM = &v
	}
	if w.AdditionalAbsentSubscriberDiagnosticSM != nil {
		v := AbsentSubscriberDiagnosticSM(*w.AdditionalAbsentSubscriberDiagnosticSM)
		out.AdditionalAbsentSubscriberDiagnosticSM = &v
	}
	if w.Imsi != nil {
		imsi, err := decodeIdentityDigits(identityIMSI, *w.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding AbsentSubscriberSMParam.IMSI: %w", err)
		}
		out.IMSI = imsi
	}
	if w.RequestedRetransmissionTime != nil {
		out.RequestedRetransmissionTime = HexBytes(*w.RequestedRetransmissionTime)
	}
	if w.UserIdentifierAlert != nil {
		uid, err := decodeIdentityDigits(identityIMSI, *w.UserIdentifierAlert)
		if err != nil {
			return nil, fmt.Errorf("decoding AbsentSubscriberSMParam.UserIdentifierAlert: %w", err)
		}
		out.UserIdentifierAlert = uid
	}
	return out, nil
}

// ============================================================================
// UnknownSubscriberParam — TS 29.002 §17.7.7 (errorCode 1)
// ============================================================================

func convertWireToUnknownSubscriberParam(w *gsm_map.UnknownSubscriberParam) (*UnknownSubscriberParam, error) {
	if w == nil {
		return nil, nil
	}
	out := &UnknownSubscriberParam{}
	// 3GPP TS 29.002 V19.1.0 §17.7.7 UnknownSubscriberDiagnostic: "if unknown
	// values are received in UnknownSubscriberDiagnostic they shall be
	// discarded".
	if v := w.UnknownSubscriberDiagnostic; v != nil {
		switch *v {
		case gsm_map.UnknownSubscriberDiagnosticImsiUnknown,
			gsm_map.UnknownSubscriberDiagnosticGprsEpsSubscriptionUnknown,
			gsm_map.UnknownSubscriberDiagnosticNpdbMismatch:
			d := *v
			out.UnknownSubscriberDiagnostic = &d
		}
	}
	return out, nil
}

// ============================================================================
// CallBarredParam — TS 29.002 §17.7.7 (errorCode 13)
// ============================================================================

func convertWireToCallBarredParam(w *gsm_map.CallBarredParam) (*CallBarredParam, error) {
	if w == nil {
		return nil, nil
	}
	out := &CallBarredParam{}
	switch w.Choice {
	case gsm_map.CallBarredParamChoiceCallBarringCause:
		v := *w.CallBarringCause
		out.CallBarringCause = &v
	case gsm_map.CallBarredParamChoiceExtensibleCallBarredParam:
		ext := convertWireToExtensibleCallBarredParam(w.ExtensibleCallBarredParam)
		out.ExtensibleCallBarredParam = ext
	default:
		return nil, fmt.Errorf("%w: CallBarredParam: unsupported choice %d", ErrCallBarredParamUnknownAlternative, w.Choice)
	}
	return out, nil
}

func convertWireToExtensibleCallBarredParam(w *gsm_map.ExtensibleCallBarredParam) *ExtensibleCallBarredParam {
	if w == nil {
		return nil
	}
	out := &ExtensibleCallBarredParam{
		UnauthorisedMessageOriginator: nullPtrToBool(w.UnauthorisedMessageOriginator),
		AnonymousCallRejection:        nullPtrToBool(w.AnonymousCallRejection),
	}
	if w.CallBarringCause != nil {
		v := *w.CallBarringCause
		out.CallBarringCause = &v
	}
	return out
}

// ============================================================================
// SystemFailureParam — TS 29.002 §17.7.7 (errorCode 34)
// ============================================================================

func convertWireToSystemFailureParam(w *gsm_map.SystemFailureParam) (*SystemFailureParam, error) {
	if w == nil {
		return nil, nil
	}
	out := &SystemFailureParam{}
	switch w.Choice {
	case gsm_map.SystemFailureParamChoiceNetworkResource:
		v := *w.NetworkResource
		out.NetworkResource = &v
	case gsm_map.SystemFailureParamChoiceExtensibleSystemFailureParam:
		ext := convertWireToExtensibleSystemFailureParam(w.ExtensibleSystemFailureParam)
		out.ExtensibleSystemFailureParam = ext
	default:
		return nil, fmt.Errorf("%w: SystemFailureParam: unsupported choice %d", ErrSystemFailureParamUnknownAlternative, w.Choice)
	}
	return out, nil
}

func convertWireToExtensibleSystemFailureParam(w *gsm_map.ExtensibleSystemFailureParam) *ExtensibleSystemFailureParam {
	if w == nil {
		return nil
	}
	out := &ExtensibleSystemFailureParam{}
	if w.NetworkResource != nil {
		v := *w.NetworkResource
		out.NetworkResource = &v
	}
	// 3GPP TS 29.002 V19.1.0 §17.7.8 AdditionalNetworkResource: "if unknown
	// value is received in AdditionalNetworkResource it shall be ignored."
	if v := w.AdditionalNetworkResource; v != nil && *v >= gsm_map.AdditionalNetworkResourceSgsn && *v <= gsm_map.AdditionalNetworkResourceMme {
		r := *v
		out.AdditionalNetworkResource = &r
	}
	// 3GPP TS 29.002 V19.1.0 §17.7.7 FailureCauseParam: "if unknown value is
	// received in FailureCauseParam it shall be ignored".
	if v := w.FailureCauseParam; v != nil && *v == gsm_map.FailureCauseParamLimitReachedOnNumberOfConcurrentLocationRequests {
		c := *v
		out.FailureCauseParam = &c
	}
	return out
}

// ============================================================================
// RoamingNotAllowedParam — TS 29.002 §17.7.7 (errorCode 8)
// ============================================================================

func convertWireToRoamingNotAllowedParam(w *gsm_map.RoamingNotAllowedParam) *RoamingNotAllowedParam {
	if w == nil {
		return nil
	}
	out := &RoamingNotAllowedParam{RoamingNotAllowedCause: w.RoamingNotAllowedCause}
	if w.AdditionalRoamingNotAllowedCause != nil {
		v := *w.AdditionalRoamingNotAllowedCause
		out.AdditionalRoamingNotAllowedCause = &v
	}
	return out
}

// ============================================================================
// UnauthorizedRequestingNetworkParam — TS 29.002 (errorCode 52)
// FacilityNotSupParam — TS 29.002 (errorCode 21)
// TeleservNotProvParam — TS 29.002 (errorCode 11)
// DataMissingParam — TS 29.002 (errorCode 35)
// ============================================================================
//
// These types carry only ExtensionContainer (which we don't surface)
// or a small set of NULL flags. Decoders are minimal pass-through.

func convertWireToUnauthorizedRequestingNetworkParam(w *gsm_map.UnauthorizedRequestingNetworkParam) (*UnauthorizedRequestingNetworkParam, error) {
	if w == nil {
		return nil, nil
	}
	return &UnauthorizedRequestingNetworkParam{}, nil
}

func convertWireToFacilityNotSupParam(w *gsm_map.FacilityNotSupParam) (*FacilityNotSupParam, error) {
	if w == nil {
		return nil, nil
	}
	return &FacilityNotSupParam{
		ShapeOfLocationEstimateNotSupported:          nullPtrToBool(w.ShapeOfLocationEstimateNotSupported),
		NeededLcsCapabilityNotSupportedInServingNode: nullPtrToBool(w.NeededLcsCapabilityNotSupportedInServingNode),
	}, nil
}

func convertWireToTeleservNotProvParam(w *gsm_map.TeleservNotProvParam) (*TeleservNotProvParam, error) {
	if w == nil {
		return nil, nil
	}
	return &TeleservNotProvParam{}, nil
}

func convertWireToDataMissingParam(w *gsm_map.DataMissingParam) (*DataMissingParam, error) {
	if w == nil {
		return nil, nil
	}
	return &DataMissingParam{}, nil
}

// ============================================================================
// AbsentSubscriberParam — TS 29.002 §17.7.7 (errorCode 27)
// ============================================================================

func convertWireToAbsentSubscriberParam(w *gsm_map.AbsentSubscriberParam) (*AbsentSubscriberParam, error) {
	if w == nil {
		return nil, nil
	}
	out := &AbsentSubscriberParam{}
	// 3GPP TS 29.002 V19.1.0 §17.7.7 AbsentSubscriberReason: "at reception of
	// other values than the ones listed the AbsentSubscriberReason shall be
	// ignored."
	if v := w.AbsentSubscriberReason; v != nil && *v >= gsm_map.AbsentSubscriberReasonImsiDetach && *v <= gsm_map.AbsentSubscriberReasonBusySubscriber {
		r := *v
		out.AbsentSubscriberReason = &r
	}
	return out, nil
}

// ============================================================================
// IllegalSubscriberParam, IllegalEquipmentParam, UnexpectedDataParam —
// TS 29.002 §17.7.7 (errorCodes 9, 12, 36)
// ============================================================================

func convertWireToIllegalSubscriberParam(w *gsm_map.IllegalSubscriberParam) (*IllegalSubscriberParam, error) {
	if w == nil {
		return nil, nil
	}
	return &IllegalSubscriberParam{}, nil
}

func convertWireToIllegalEquipmentParam(w *gsm_map.IllegalEquipmentParam) (*IllegalEquipmentParam, error) {
	if w == nil {
		return nil, nil
	}
	return &IllegalEquipmentParam{}, nil
}

func convertWireToUnexpectedDataParam(w *gsm_map.UnexpectedDataParam) (*UnexpectedDataParam, error) {
	if w == nil {
		return nil, nil
	}
	return &UnexpectedDataParam{UnexpectedSubscriber: w.UnexpectedSubscriber != nil}, nil
}

func convertAbsentSubscriberSMParamToWire(p *AbsentSubscriberSMParam) (*gsm_map.AbsentSubscriberSMParam, error) {
	if p == nil {
		return nil, ErrReturnErrorParameterNil
	}
	w := &gsm_map.AbsentSubscriberSMParam{}
	if p.AbsentSubscriberDiagnosticSM != nil {
		w.AbsentSubscriberDiagnosticSM = new(int64(*p.AbsentSubscriberDiagnosticSM))
	}
	if p.AdditionalAbsentSubscriberDiagnosticSM != nil {
		w.AdditionalAbsentSubscriberDiagnosticSM = new(int64(*p.AdditionalAbsentSubscriberDiagnosticSM))
	}
	if p.IMSI != "" {
		v, err := encodeIdentityDigits(identityIMSI, p.IMSI)
		if err != nil {
			return nil, fmt.Errorf("encoding AbsentSubscriberSMParam.IMSI: %w", err)
		}
		w.Imsi = &v
	}
	if len(p.RequestedRetransmissionTime) > 0 {
		v := gsm_map.Time(p.RequestedRetransmissionTime)
		w.RequestedRetransmissionTime = &v
	}
	if p.UserIdentifierAlert != "" {
		v, err := encodeIdentityDigits(identityIMSI, p.UserIdentifierAlert)
		if err != nil {
			return nil, fmt.Errorf("encoding AbsentSubscriberSMParam.UserIdentifierAlert: %w", err)
		}
		w.UserIdentifierAlert = &v
	}
	return w, nil
}

func convertUnknownSubscriberParamToWire(p *UnknownSubscriberParam) (*gsm_map.UnknownSubscriberParam, error) {
	if p == nil {
		return nil, ErrReturnErrorParameterNil
	}
	// 3GPP TS 29.002 V19.1.0 §17.7.7: send only defined UnknownSubscriberDiagnostic values.
	if v := p.UnknownSubscriberDiagnostic; v != nil {
		switch *v {
		case UnknownSubscriberDiagnosticImsiUnknown, UnknownSubscriberDiagnosticGprsEpsSubscriptionUnknown, UnknownSubscriberDiagnosticNpdbMismatch:
		default:
			return nil, ErrUnknownSubscriberDiagnosticInvalid
		}
	}
	return &gsm_map.UnknownSubscriberParam{UnknownSubscriberDiagnostic: p.UnknownSubscriberDiagnostic}, nil
}

func convertCallBarredParamToWire(p *CallBarredParam) (*gsm_map.CallBarredParam, error) {
	if p == nil {
		return nil, ErrReturnErrorParameterNil
	}
	// 3GPP TS 29.002 V19.1.0 §17.7.7 CallBarredParam CHOICE.
	if p.CallBarringCause != nil && p.ExtensibleCallBarredParam != nil {
		return nil, ErrCallBarredParamMultipleAlternatives
	}
	if p.CallBarringCause != nil {
		w := gsm_map.NewCallBarredParamCallBarringCause(*p.CallBarringCause)
		return &w, nil
	}
	if p.ExtensibleCallBarredParam == nil {
		return nil, ErrCallBarredParamNoAlternative
	}
	ext := p.ExtensibleCallBarredParam
	// 3GPP TS 29.002 V19.1.0 §17.7.7: the two indications shall be mutually exclusive.
	if ext.UnauthorisedMessageOriginator && ext.AnonymousCallRejection {
		return nil, ErrExtensibleCallBarredParamMutuallyExclusiveIndicators
	}
	w := gsm_map.NewCallBarredParamExtensibleCallBarredParam(gsm_map.ExtensibleCallBarredParam{
		CallBarringCause:              ext.CallBarringCause,
		UnauthorisedMessageOriginator: boolToNullPtr(ext.UnauthorisedMessageOriginator),
		AnonymousCallRejection:        boolToNullPtr(ext.AnonymousCallRejection),
	})
	return &w, nil
}

func convertSystemFailureParamToWire(p *SystemFailureParam) (*gsm_map.SystemFailureParam, error) {
	if p == nil {
		return nil, ErrReturnErrorParameterNil
	}
	// 3GPP TS 29.002 V19.1.0 §17.7.7 SystemFailureParam CHOICE.
	if p.NetworkResource != nil && p.ExtensibleSystemFailureParam != nil {
		return nil, ErrSystemFailureParamMultipleAlternatives
	}
	if p.NetworkResource != nil {
		w := gsm_map.NewSystemFailureParamNetworkResource(*p.NetworkResource)
		return &w, nil
	}
	if p.ExtensibleSystemFailureParam == nil {
		return nil, ErrSystemFailureParamNoAlternative
	}
	ext := p.ExtensibleSystemFailureParam
	// 3GPP TS 29.002 V19.1.0 §17.7.8: send only defined AdditionalNetworkResource values.
	if v := ext.AdditionalNetworkResource; v != nil && (*v < AdditionalNetworkResourceSgsn || *v > AdditionalNetworkResourceMme) {
		return nil, ErrAdditionalNetworkResourceInvalid
	}
	// 3GPP TS 29.002 V19.1.0 §17.7.7: send only defined FailureCauseParam values.
	if v := ext.FailureCauseParam; v != nil && *v != FailureCauseParamLimitReachedOnNumberOfConcurrentLocationRequests {
		return nil, ErrFailureCauseParamInvalid
	}
	w := gsm_map.NewSystemFailureParamExtensibleSystemFailureParam(gsm_map.ExtensibleSystemFailureParam{
		NetworkResource: ext.NetworkResource, AdditionalNetworkResource: ext.AdditionalNetworkResource, FailureCauseParam: ext.FailureCauseParam,
	})
	return &w, nil
}

func convertRoamingNotAllowedParamToWire(p *RoamingNotAllowedParam) (*gsm_map.RoamingNotAllowedParam, error) {
	if p == nil {
		return nil, ErrReturnErrorParameterNil
	}
	// 3GPP TS 29.002 V19.1.0 §17.7.7: send only defined AdditionalRoamingNotAllowedCause values.
	if v := p.AdditionalRoamingNotAllowedCause; v != nil && *v != AdditionalRoamingNotAllowedCauseSupportedRATTypesNotAllowed {
		return nil, ErrAdditionalRoamingNotAllowedCauseInvalid
	}
	return &gsm_map.RoamingNotAllowedParam{RoamingNotAllowedCause: p.RoamingNotAllowedCause, AdditionalRoamingNotAllowedCause: p.AdditionalRoamingNotAllowedCause}, nil
}

func convertAbsentSubscriberParamToWire(p *AbsentSubscriberParam) (*gsm_map.AbsentSubscriberParam, error) {
	if p == nil {
		return nil, ErrReturnErrorParameterNil
	}
	// 3GPP TS 29.002 V19.1.0 §17.7.7: send only defined AbsentSubscriberReason values.
	if v := p.AbsentSubscriberReason; v != nil && (*v < AbsentSubscriberReasonImsiDetach || *v > AbsentSubscriberReasonBusySubscriber) {
		return nil, ErrAbsentSubscriberReasonInvalid
	}
	return &gsm_map.AbsentSubscriberParam{AbsentSubscriberReason: p.AbsentSubscriberReason}, nil
}

func convertUnauthorizedRequestingNetworkParamToWire(p *UnauthorizedRequestingNetworkParam) (*gsm_map.UnauthorizedRequestingNetworkParam, error) {
	if p == nil {
		return nil, ErrReturnErrorParameterNil
	}
	return &gsm_map.UnauthorizedRequestingNetworkParam{}, nil
}

func convertTeleservNotProvParamToWire(p *TeleservNotProvParam) (*gsm_map.TeleservNotProvParam, error) {
	if p == nil {
		return nil, ErrReturnErrorParameterNil
	}
	return &gsm_map.TeleservNotProvParam{}, nil
}

func convertDataMissingParamToWire(p *DataMissingParam) (*gsm_map.DataMissingParam, error) {
	if p == nil {
		return nil, ErrReturnErrorParameterNil
	}
	return &gsm_map.DataMissingParam{}, nil
}

func convertIllegalSubscriberParamToWire(p *IllegalSubscriberParam) (*gsm_map.IllegalSubscriberParam, error) {
	if p == nil {
		return nil, ErrReturnErrorParameterNil
	}
	return &gsm_map.IllegalSubscriberParam{}, nil
}

func convertIllegalEquipmentParamToWire(p *IllegalEquipmentParam) (*gsm_map.IllegalEquipmentParam, error) {
	if p == nil {
		return nil, ErrReturnErrorParameterNil
	}
	return &gsm_map.IllegalEquipmentParam{}, nil
}

func convertFacilityNotSupParamToWire(p *FacilityNotSupParam) (*gsm_map.FacilityNotSupParam, error) {
	if p == nil {
		return nil, ErrReturnErrorParameterNil
	}
	return &gsm_map.FacilityNotSupParam{ShapeOfLocationEstimateNotSupported: boolToNullPtr(p.ShapeOfLocationEstimateNotSupported), NeededLcsCapabilityNotSupportedInServingNode: boolToNullPtr(p.NeededLcsCapabilityNotSupportedInServingNode)}, nil
}

func convertUnexpectedDataParamToWire(p *UnexpectedDataParam) (*gsm_map.UnexpectedDataParam, error) {
	if p == nil {
		return nil, ErrReturnErrorParameterNil
	}
	return &gsm_map.UnexpectedDataParam{UnexpectedSubscriber: boolToNullPtr(p.UnexpectedSubscriber)}, nil
}
