// convert_map_error.go
//
// Converters between the wrapper-level MAP ReturnError diagnostic
// types (defined in gsmmap.go) and their gsm_map.*Param wire forms.
// PR F2 of the staged ReturnError.Parameter implementation, building
// on PR #49 (types only). Parse* helpers and the
// ParseReturnErrorParameter dispatcher live in parse.go.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// ============================================================================
// AbsentSubscriberSMParam — TS 29.002 MAP-ER-DataTypes.asn (errorCode 6)
// ============================================================================

func convertWireToAbsentSubscriberSMParam(w *gsm_map.AbsentSubscriberSMParam) (*AbsentSubscriberSMParam, error) {
	if w == nil {
		return nil, nil
	}
	out := &AbsentSubscriberSMParam{}
	if w.AbsentSubscriberDiagnosticSM != nil {
		v := *w.AbsentSubscriberDiagnosticSM
		out.AbsentSubscriberDiagnosticSM = &v
	}
	if w.AdditionalAbsentSubscriberDiagnosticSM != nil {
		v := *w.AdditionalAbsentSubscriberDiagnosticSM
		out.AdditionalAbsentSubscriberDiagnosticSM = &v
	}
	if w.Imsi != nil {
		imsi, err := tbcd.Decode(*w.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding AbsentSubscriberSMParam.IMSI: %w", err)
		}
		if imsi == "" {
			return nil, fmt.Errorf("AbsentSubscriberSMParam.IMSI: present wire field decoded to empty digits; presence cannot round-trip through string-based API")
		}
		out.IMSI = imsi
	}
	if w.RequestedRetransmissionTime != nil {
		out.RequestedRetransmissionTime = HexBytes(*w.RequestedRetransmissionTime)
	}
	if w.UserIdentifierAlert != nil {
		uid, err := tbcd.Decode(*w.UserIdentifierAlert)
		if err != nil {
			return nil, fmt.Errorf("decoding AbsentSubscriberSMParam.UserIdentifierAlert: %w", err)
		}
		if uid == "" {
			return nil, fmt.Errorf("AbsentSubscriberSMParam.UserIdentifierAlert: present wire field decoded to empty digits; presence cannot round-trip through string-based API")
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
	if w.UnknownSubscriberDiagnostic != nil {
		v := *w.UnknownSubscriberDiagnostic
		out.UnknownSubscriberDiagnostic = &v
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
		out.NetworkResource = &v
	}
	if w.AdditionalNetworkResource != nil {
		v := *w.AdditionalNetworkResource
		out.AdditionalNetworkResource = &v
	}
	if w.FailureCauseParam != nil {
		v := *w.FailureCauseParam
		out.FailureCauseParam = &v
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
	// RoamingNotAllowedCause is non-extensible per TS 29.002
	// MAP-ER-DataTypes.asn with non-contiguous values: 0
	// (plmnRoamingNotAllowed) and 3 (operatorDeterminedBarring).
	switch w.RoamingNotAllowedCause {
	case gsm_map.RoamingNotAllowedCausePlmnRoamingNotAllowed,
		gsm_map.RoamingNotAllowedCauseOperatorDeterminedBarring:
		// valid
	default:
		return nil, fmt.Errorf("RoamingNotAllowedParam.RoamingNotAllowedCause=%d: must be 0 (plmnRoamingNotAllowed) or 3 (operatorDeterminedBarring) per TS 29.002 MAP-ER-DataTypes.asn", w.RoamingNotAllowedCause)
	}
	out := &RoamingNotAllowedParam{
		RoamingNotAllowedCause: w.RoamingNotAllowedCause,
	}
	if w.AdditionalRoamingNotAllowedCause != nil {
		v := *w.AdditionalRoamingNotAllowedCause
		out.AdditionalRoamingNotAllowedCause = &v
	}
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
	if w.AbsentSubscriberReason != nil {
		v := *w.AbsentSubscriberReason
		out.AbsentSubscriberReason = &v
	}
	return out, nil
}
