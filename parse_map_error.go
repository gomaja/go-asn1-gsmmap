// parse_map_error.go
//
// Parse* helpers for the BER-encoded TCAP ReturnError.Parameter
// payload, plus a dispatcher (ParseReturnErrorParameter) that selects
// the right parser based on the MAP error opcode.
//
// PR F2 of the staged ReturnError.Parameter implementation, building
// on PR #49 (types only) and convert_map_error.go (this PR's wire ↔
// public-type converters).

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ParseReturnErrorParameter decodes the BER-encoded parameter of a
// TCAP ReturnError component into a wrapper-level diagnostic struct.
//
// errorCode is the MAP error opcode from TCAP ReturnError.ErrorCode;
// data is TCAP ReturnError.Parameter. The concrete type of the
// returned value depends on errorCode:
//
//	MapErrorUnknownSubscriber             (1)  → *UnknownSubscriberParam
//	MapErrorAbsentSubscriberSM            (6)  → *AbsentSubscriberSMParam
//	MapErrorRoamingNotAllowed             (8)  → *RoamingNotAllowedParam
//	MapErrorTeleserviceNotProvisioned     (11) → *TeleservNotProvParam
//	MapErrorCallBarred                    (13) → *CallBarredParam
//	MapErrorFacilityNotSupported          (21) → *FacilityNotSupParam
//	MapErrorAbsentSubscriber              (27) → *AbsentSubscriberParam
//	MapErrorSystemFailure                 (34) → *SystemFailureParam
//	MapErrorDataMissing                   (35) → *DataMissingParam
//	MapErrorUnauthorizedRequestingNetwork (52) → *UnauthorizedRequestingNetworkParam
//
// Returns (nil, nil) for unhandled error codes or when data is empty,
// so callers can safely call this for every ReturnError without
// branching first.
//
// errorCode is typed as int64 to match TCAP ReturnError.ErrorCode on
// the wire. Callers using the typed MapErrorCode constants can pass
// them with an explicit cast (int64(MapErrorAbsentSubscriberSM)) or
// use the untyped numeric value directly.
func ParseReturnErrorParameter(errorCode int64, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	switch MapErrorCode(errorCode) {
	case MapErrorUnknownSubscriber:
		return ParseUnknownSubscriberParam(data)
	case MapErrorAbsentSubscriberSM:
		return ParseAbsentSubscriberSMParam(data)
	case MapErrorRoamingNotAllowed:
		return ParseRoamingNotAllowedParam(data)
	case MapErrorTeleserviceNotProvisioned:
		return ParseTeleservNotProvParam(data)
	case MapErrorCallBarred:
		return ParseCallBarredParam(data)
	case MapErrorFacilityNotSupported:
		return ParseFacilityNotSupParam(data)
	case MapErrorAbsentSubscriber:
		return ParseAbsentSubscriberParam(data)
	case MapErrorSystemFailure:
		return ParseSystemFailureParam(data)
	case MapErrorDataMissing:
		return ParseDataMissingParam(data)
	case MapErrorUnauthorizedRequestingNetwork:
		return ParseUnauthorizedRequestingNetworkParam(data)
	default:
		return nil, nil
	}
}

// ParseAbsentSubscriberSMParam decodes BER-encoded bytes into an
// AbsentSubscriberSMParam (errorCode 6).
func ParseAbsentSubscriberSMParam(data []byte) (*AbsentSubscriberSMParam, error) {
	var w gsm_map.AbsentSubscriberSMParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding AbsentSubscriberSMParam: %w", err)
	}
	return convertWireToAbsentSubscriberSMParam(&w)
}

// ParseUnknownSubscriberParam decodes BER-encoded bytes into an
// UnknownSubscriberParam (errorCode 1).
func ParseUnknownSubscriberParam(data []byte) (*UnknownSubscriberParam, error) {
	var w gsm_map.UnknownSubscriberParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding UnknownSubscriberParam: %w", err)
	}
	return convertWireToUnknownSubscriberParam(&w)
}

// ParseCallBarredParam decodes BER-encoded bytes into a CallBarredParam
// (errorCode 13). Handles both legacy (CallBarringCause alone) and
// extensible (ExtensibleCallBarredParam) CHOICE variants.
func ParseCallBarredParam(data []byte) (*CallBarredParam, error) {
	var w gsm_map.CallBarredParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding CallBarredParam: %w", err)
	}
	return convertWireToCallBarredParam(&w)
}

// ParseSystemFailureParam decodes BER-encoded bytes into a
// SystemFailureParam (errorCode 34). Handles both legacy (NetworkResource
// alone) and extensible (ExtensibleSystemFailureParam) CHOICE variants.
func ParseSystemFailureParam(data []byte) (*SystemFailureParam, error) {
	var w gsm_map.SystemFailureParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding SystemFailureParam: %w", err)
	}
	return convertWireToSystemFailureParam(&w)
}

// ParseRoamingNotAllowedParam decodes BER-encoded bytes into a
// RoamingNotAllowedParam (errorCode 8).
func ParseRoamingNotAllowedParam(data []byte) (*RoamingNotAllowedParam, error) {
	var w gsm_map.RoamingNotAllowedParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding RoamingNotAllowedParam: %w", err)
	}
	return convertWireToRoamingNotAllowedParam(&w)
}

// ParseUnauthorizedRequestingNetworkParam decodes BER-encoded bytes
// into an UnauthorizedRequestingNetworkParam (errorCode 52).
func ParseUnauthorizedRequestingNetworkParam(data []byte) (*UnauthorizedRequestingNetworkParam, error) {
	var w gsm_map.UnauthorizedRequestingNetworkParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding UnauthorizedRequestingNetworkParam: %w", err)
	}
	return convertWireToUnauthorizedRequestingNetworkParam(&w)
}

// ParseFacilityNotSupParam decodes BER-encoded bytes into a
// FacilityNotSupParam (errorCode 21).
func ParseFacilityNotSupParam(data []byte) (*FacilityNotSupParam, error) {
	var w gsm_map.FacilityNotSupParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding FacilityNotSupParam: %w", err)
	}
	return convertWireToFacilityNotSupParam(&w)
}

// ParseTeleservNotProvParam decodes BER-encoded bytes into a
// TeleservNotProvParam (errorCode 11).
func ParseTeleservNotProvParam(data []byte) (*TeleservNotProvParam, error) {
	var w gsm_map.TeleservNotProvParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding TeleservNotProvParam: %w", err)
	}
	return convertWireToTeleservNotProvParam(&w)
}

// ParseDataMissingParam decodes BER-encoded bytes into a
// DataMissingParam (errorCode 35).
func ParseDataMissingParam(data []byte) (*DataMissingParam, error) {
	var w gsm_map.DataMissingParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding DataMissingParam: %w", err)
	}
	return convertWireToDataMissingParam(&w)
}

// ParseAbsentSubscriberParam decodes BER-encoded bytes into an
// AbsentSubscriberParam (errorCode 27). Distinct from
// AbsentSubscriberSMParam (errorCode 6).
func ParseAbsentSubscriberParam(data []byte) (*AbsentSubscriberParam, error) {
	var w gsm_map.AbsentSubscriberParam
	if err := w.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding AbsentSubscriberParam: %w", err)
	}
	return convertWireToAbsentSubscriberParam(&w)
}
