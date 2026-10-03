// parse_map_error.go
//
// ParseReturnErrorParameter, the decoder of the BER-encoded parameter of a
// TCAP ReturnError component, and the per-error parsers it dispatches to.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ParseReturnErrorParameter decodes the BER-encoded parameter of a TCAP
// ReturnError component: errorCode is the MAP local error code of the
// component and data its parameter. The concrete type of the returned value
// depends on errorCode (TS 29.002 §17.6.6 for the errors, §17.7.7 for their
// parameters):
//
//	MapErrorUnknownSubscriber             (1)  → *UnknownSubscriberParam
//	MapErrorAbsentSubscriberSM            (6)  → *AbsentSubscriberSMParam
//	MapErrorRoamingNotAllowed             (8)  → *RoamingNotAllowedParam
//	MapErrorIllegalSubscriber             (9)  → *IllegalSubscriberParam
//	MapErrorTeleserviceNotProvisioned     (11) → *TeleservNotProvParam
//	MapErrorIllegalEquipment              (12) → *IllegalEquipmentParam
//	MapErrorCallBarred                    (13) → *CallBarredParam
//	MapErrorFacilityNotSupported          (21) → *FacilityNotSupParam
//	MapErrorAbsentSubscriber              (27) → *AbsentSubscriberParam
//	MapErrorSystemFailure                 (34) → *SystemFailureParam
//	MapErrorDataMissing                   (35) → *DataMissingParam
//	MapErrorUnexpectedDataValue           (36) → *UnexpectedDataParam
//	MapErrorUnauthorizedRequestingNetwork (52) → *UnauthorizedRequestingNetworkParam
//
// unknownAlphabet (71) and ussd-Busy (72) have no parameter. For them, for
// any error code not listed above, and when data is empty, the result is
// (nil, nil), so a caller can pass every ReturnError without branching first.
func ParseReturnErrorParameter(errorCode MapErrorCode, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	switch errorCode {
	case MapErrorUnknownSubscriber:
		return dispatched(parseUnknownSubscriberParam(data))
	case MapErrorAbsentSubscriberSM:
		return dispatched(parseAbsentSubscriberSMParam(data))
	case MapErrorRoamingNotAllowed:
		return dispatched(parseRoamingNotAllowedParam(data))
	case MapErrorIllegalSubscriber:
		return dispatched(parseIllegalSubscriberParam(data))
	case MapErrorTeleserviceNotProvisioned:
		return dispatched(parseTeleservNotProvParam(data))
	case MapErrorIllegalEquipment:
		return dispatched(parseIllegalEquipmentParam(data))
	case MapErrorCallBarred:
		return dispatched(parseCallBarredParam(data))
	case MapErrorFacilityNotSupported:
		return dispatched(parseFacilityNotSupParam(data))
	case MapErrorAbsentSubscriber:
		return dispatched(parseAbsentSubscriberParam(data))
	case MapErrorSystemFailure:
		return dispatched(parseSystemFailureParam(data))
	case MapErrorDataMissing:
		return dispatched(parseDataMissingParam(data))
	case MapErrorUnexpectedDataValue:
		return dispatched(parseUnexpectedDataParam(data))
	case MapErrorUnauthorizedRequestingNetwork:
		return dispatched(parseUnauthorizedRequestingNetworkParam(data))
	default:
		return nil, nil
	}
}

// dispatched returns the result of a per-error parser as an untyped any: a
// nil *T becomes a nil interface, so a failed or empty decode never yields a
// non-nil any holding a nil pointer.
func dispatched[T any](v *T, err error) (any, error) {
	if err != nil || v == nil {
		return nil, err
	}
	return v, nil
}

// parseAbsentSubscriberSMParam decodes BER-encoded bytes into an
// AbsentSubscriberSMParam (errorCode 6).
func parseAbsentSubscriberSMParam(data []byte) (*AbsentSubscriberSMParam, error) {
	var w gsm_map.AbsentSubscriberSMParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding AbsentSubscriberSMParam: %w", err)
	}
	return convertWireToAbsentSubscriberSMParam(&w)
}

// parseUnknownSubscriberParam decodes BER-encoded bytes into an
// UnknownSubscriberParam (errorCode 1).
func parseUnknownSubscriberParam(data []byte) (*UnknownSubscriberParam, error) {
	var w gsm_map.UnknownSubscriberParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding UnknownSubscriberParam: %w", err)
	}
	return convertWireToUnknownSubscriberParam(&w)
}

// parseCallBarredParam decodes BER-encoded bytes into a CallBarredParam
// (errorCode 13). Handles both the bare CallBarringCause and
// extensible (ExtensibleCallBarredParam) CHOICE variants.
func parseCallBarredParam(data []byte) (*CallBarredParam, error) {
	var w gsm_map.CallBarredParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding CallBarredParam: %w", err)
	}
	return convertWireToCallBarredParam(&w)
}

// parseSystemFailureParam decodes BER-encoded bytes into a
// SystemFailureParam (errorCode 34). Handles both the bare NetworkResource
// alone) and extensible (ExtensibleSystemFailureParam) CHOICE variants.
func parseSystemFailureParam(data []byte) (*SystemFailureParam, error) {
	var w gsm_map.SystemFailureParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding SystemFailureParam: %w", err)
	}
	return convertWireToSystemFailureParam(&w)
}

// parseRoamingNotAllowedParam decodes BER-encoded bytes into a
// RoamingNotAllowedParam (errorCode 8).
func parseRoamingNotAllowedParam(data []byte) (*RoamingNotAllowedParam, error) {
	var w gsm_map.RoamingNotAllowedParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding RoamingNotAllowedParam: %w", err)
	}
	return convertWireToRoamingNotAllowedParam(&w)
}

// parseUnauthorizedRequestingNetworkParam decodes BER-encoded bytes
// into an UnauthorizedRequestingNetworkParam (errorCode 52).
func parseUnauthorizedRequestingNetworkParam(data []byte) (*UnauthorizedRequestingNetworkParam, error) {
	var w gsm_map.UnauthorizedRequestingNetworkParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding UnauthorizedRequestingNetworkParam: %w", err)
	}
	return convertWireToUnauthorizedRequestingNetworkParam(&w)
}

// parseFacilityNotSupParam decodes BER-encoded bytes into a
// FacilityNotSupParam (errorCode 21).
func parseFacilityNotSupParam(data []byte) (*FacilityNotSupParam, error) {
	var w gsm_map.FacilityNotSupParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding FacilityNotSupParam: %w", err)
	}
	return convertWireToFacilityNotSupParam(&w)
}

// parseTeleservNotProvParam decodes BER-encoded bytes into a
// TeleservNotProvParam (errorCode 11).
func parseTeleservNotProvParam(data []byte) (*TeleservNotProvParam, error) {
	var w gsm_map.TeleservNotProvParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding TeleservNotProvParam: %w", err)
	}
	return convertWireToTeleservNotProvParam(&w)
}

// parseDataMissingParam decodes BER-encoded bytes into a
// DataMissingParam (errorCode 35).
func parseDataMissingParam(data []byte) (*DataMissingParam, error) {
	var w gsm_map.DataMissingParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding DataMissingParam: %w", err)
	}
	return convertWireToDataMissingParam(&w)
}

// parseAbsentSubscriberParam decodes BER-encoded bytes into an
// AbsentSubscriberParam (errorCode 27). Distinct from
// AbsentSubscriberSMParam (errorCode 6).
func parseAbsentSubscriberParam(data []byte) (*AbsentSubscriberParam, error) {
	var w gsm_map.AbsentSubscriberParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding AbsentSubscriberParam: %w", err)
	}
	return convertWireToAbsentSubscriberParam(&w)
}

// parseIllegalSubscriberParam decodes BER-encoded bytes into an
// IllegalSubscriberParam (errorCode 9).
func parseIllegalSubscriberParam(data []byte) (*IllegalSubscriberParam, error) {
	var w gsm_map.IllegalSubscriberParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding IllegalSubscriberParam: %w", err)
	}
	return convertWireToIllegalSubscriberParam(&w)
}

// parseIllegalEquipmentParam decodes BER-encoded bytes into an
// IllegalEquipmentParam (errorCode 12).
func parseIllegalEquipmentParam(data []byte) (*IllegalEquipmentParam, error) {
	var w gsm_map.IllegalEquipmentParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding IllegalEquipmentParam: %w", err)
	}
	return convertWireToIllegalEquipmentParam(&w)
}

// parseUnexpectedDataParam decodes BER-encoded bytes into an
// UnexpectedDataParam (errorCode 36).
func parseUnexpectedDataParam(data []byte) (*UnexpectedDataParam, error) {
	var w gsm_map.UnexpectedDataParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding UnexpectedDataParam: %w", err)
	}
	return convertWireToUnexpectedDataParam(&w)
}
