// convert_map_error.go
//
// Converters from the gsm_map.*Param wire forms of the MAP ReturnError
// parameters to the public diagnostic types defined in gsmmap.go. The
// ParseReturnErrorParameter dispatcher lives in parse_map_error.go.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// AbsentSubscriberSMParam — TS 29.002 MAP-ER-DataTypes.asn (errorCode 6)
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
// UnknownSubscriberParam — TS 29.002 MAP-ER-DataTypes.asn (errorCode 1)
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
// CallBarredParam — TS 29.002 MAP-ER-DataTypes.asn (errorCode 13)
// ============================================================================

func convertWireToCallBarredParam(w *gsm_map.CallBarredParam) (*CallBarredParam, error) {
	if w == nil {
		return nil, nil
	}
	out := &CallBarredParam{}
	switch w.Choice {
	case gsm_map.CallBarredParamChoiceCallBarringCause:
		if w.CallBarringCause == nil {
			return nil, fmt.Errorf("CallBarredParam: choice=CallBarringCause but payload is nil")
		}
		v := *w.CallBarringCause
		// CallBarringCause is non-extensible (TS 29.002 MAP-ER-DataTypes.asn);
		// reject out-of-range values per project convention.
		// go-asn1 does not enforce ENUMERATED membership: https://github.com/gomaja/go-asn1/issues/81.
		if int64(v) < 0 || int64(v) > 1 {
			return nil, fmt.Errorf("CallBarredParam.CallBarringCause=%d: must be 0..1 per TS 29.002 MAP-ER-DataTypes.asn", v)
		}
		out.CallBarringCause = &v
	case gsm_map.CallBarredParamChoiceExtensibleCallBarredParam:
		if w.ExtensibleCallBarredParam == nil {
			return nil, fmt.Errorf("CallBarredParam: choice=ExtensibleCallBarredParam but payload is nil")
		}
		ext, err := convertWireToExtensibleCallBarredParam(w.ExtensibleCallBarredParam)
		if err != nil {
			return nil, fmt.Errorf("CallBarredParam.ExtensibleCallBarredParam: %w", err)
		}
		out.ExtensibleCallBarredParam = ext
	default:
		return nil, fmt.Errorf("CallBarredParam: unsupported choice %d", w.Choice)
	}
	return out, nil
}

func convertWireToExtensibleCallBarredParam(w *gsm_map.ExtensibleCallBarredParam) (*ExtensibleCallBarredParam, error) {
	if w == nil {
		return nil, nil
	}
	out := &ExtensibleCallBarredParam{
		UnauthorisedMessageOriginator: nullPtrToBool(w.UnauthorisedMessageOriginator),
		AnonymousCallRejection:        nullPtrToBool(w.AnonymousCallRejection),
	}
	if w.CallBarringCause != nil {
		v := *w.CallBarringCause
		// go-asn1 does not enforce ENUMERATED membership: https://github.com/gomaja/go-asn1/issues/81.
		if int64(v) < 0 || int64(v) > 1 {
			return nil, fmt.Errorf("ExtensibleCallBarredParam.CallBarringCause=%d: must be 0..1 per TS 29.002 MAP-ER-DataTypes.asn", v)
		}
		out.CallBarringCause = &v
	}
	return out, nil
}

// ============================================================================
// SystemFailureParam — TS 29.002 MAP-ER-DataTypes.asn (errorCode 34)
// ============================================================================

// isValidNetworkResource reports whether v is one of the NetworkResource
// values of 3GPP TS 29.002 V19.1.0 §17.7.8, plmn (0) to rss (7). The type is
// not extensible.
// go-asn1 does not enforce ENUMERATED membership: https://github.com/gomaja/go-asn1/issues/81.
func isValidNetworkResource(v gsm_map.NetworkResource) bool {
	return v >= gsm_map.NetworkResourcePlmn && v <= gsm_map.NetworkResourceRss
}

func convertWireToSystemFailureParam(w *gsm_map.SystemFailureParam) (*SystemFailureParam, error) {
	if w == nil {
		return nil, nil
	}
	out := &SystemFailureParam{}
	switch w.Choice {
	case gsm_map.SystemFailureParamChoiceNetworkResource:
		if w.NetworkResource == nil {
			return nil, fmt.Errorf("SystemFailureParam: choice=NetworkResource but payload is nil")
		}
		v := *w.NetworkResource
		if !isValidNetworkResource(v) {
			return nil, fmt.Errorf("SystemFailureParam.NetworkResource=%d: %w", v, ErrNetworkResourceInvalid)
		}
		out.NetworkResource = &v
	case gsm_map.SystemFailureParamChoiceExtensibleSystemFailureParam:
		if w.ExtensibleSystemFailureParam == nil {
			return nil, fmt.Errorf("SystemFailureParam: choice=ExtensibleSystemFailureParam but payload is nil")
		}
		ext, err := convertWireToExtensibleSystemFailureParam(w.ExtensibleSystemFailureParam)
		if err != nil {
			return nil, fmt.Errorf("SystemFailureParam.ExtensibleSystemFailureParam: %w", err)
		}
		out.ExtensibleSystemFailureParam = ext
	default:
		return nil, fmt.Errorf("SystemFailureParam: unsupported choice %d", w.Choice)
	}
	return out, nil
}

func convertWireToExtensibleSystemFailureParam(w *gsm_map.ExtensibleSystemFailureParam) (*ExtensibleSystemFailureParam, error) {
	if w == nil {
		return nil, nil
	}
	out := &ExtensibleSystemFailureParam{}
	if w.NetworkResource != nil {
		v := *w.NetworkResource
		if !isValidNetworkResource(v) {
			return nil, fmt.Errorf("ExtensibleSystemFailureParam.NetworkResource=%d: %w", v, ErrNetworkResourceInvalid)
		}
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
	return out, nil
}

// ============================================================================
// RoamingNotAllowedParam — TS 29.002 MAP-ER-DataTypes.asn (errorCode 8)
// ============================================================================

func convertWireToRoamingNotAllowedParam(w *gsm_map.RoamingNotAllowedParam) (*RoamingNotAllowedParam, error) {
	if w == nil {
		return nil, nil
	}
	out := &RoamingNotAllowedParam{}
	if w.AdditionalRoamingNotAllowedCause != nil {
		// 3GPP TS 29.002 V19.1.0 §17.7.7: "if the
		// additionalRoamingNotallowedCause is received by the MSC/VLR or SGSN
		// then the roamingNotAllowedCause shall be discarded."
		v := *w.AdditionalRoamingNotAllowedCause
		out.AdditionalRoamingNotAllowedCause = &v
		return out, nil
	}
	// RoamingNotAllowedCause is non-extensible per TS 29.002
	// MAP-ER-DataTypes.asn with non-contiguous values: 0
	// (plmnRoamingNotAllowed) and 3 (operatorDeterminedBarring).
	// go-asn1 does not enforce ENUMERATED membership: https://github.com/gomaja/go-asn1/issues/81.
	switch w.RoamingNotAllowedCause {
	case gsm_map.RoamingNotAllowedCausePlmnRoamingNotAllowed,
		gsm_map.RoamingNotAllowedCauseOperatorDeterminedBarring:
	default:
		return nil, fmt.Errorf("RoamingNotAllowedParam.RoamingNotAllowedCause=%d: must be 0 (plmnRoamingNotAllowed) or 3 (operatorDeterminedBarring) per TS 29.002 MAP-ER-DataTypes.asn", w.RoamingNotAllowedCause)
	}
	c := w.RoamingNotAllowedCause
	out.RoamingNotAllowedCause = &c
	return out, nil
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
// AbsentSubscriberParam — TS 29.002 MAP-ER-DataTypes.asn (errorCode 27)
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
