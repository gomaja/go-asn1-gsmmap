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

// MAP error opcodes per TS 29.002 §17.6 — covers the SRI-SM / SRI /
// ATI-relevant subset surfaced by this package. The full list is
// defined in upstream go-asn1; we use raw int64 to keep this PR's
// surface minimal (a typed MapErrorCode enum lands in a follow-up).
const (
	mapErrorCodeUnknownSubscriber             int64 = 1
	mapErrorCodeAbsentSubscriberSM            int64 = 6
	mapErrorCodeRoamingNotAllowed             int64 = 8
	mapErrorCodeTeleserviceNotProvisioned     int64 = 11
	mapErrorCodeCallBarred                    int64 = 13
	mapErrorCodeFacilityNotSupported          int64 = 21
	mapErrorCodeSystemFailure                 int64 = 34
	mapErrorCodeDataMissing                   int64 = 35
	mapErrorCodeUnauthorizedRequestingNetwork int64 = 52
)

// ParseReturnErrorParameter decodes the BER-encoded parameter of a
// TCAP ReturnError component into a wrapper-level diagnostic struct.
//
// errorCode is the MAP error opcode from TCAP ReturnError.ErrorCode;
// data is TCAP ReturnError.Parameter. The concrete type of the
// returned value depends on errorCode:
//
//	errorCode=1  (unknownSubscriber)             → *UnknownSubscriberParam
//	errorCode=6  (absentSubscriberSM)            → *AbsentSubscriberSMParam
//	errorCode=8  (roamingNotAllowed)             → *RoamingNotAllowedParam
//	errorCode=11 (teleserviceNotProvisioned)     → *TeleservNotProvParam
//	errorCode=13 (callBarred)                    → *CallBarredParam
//	errorCode=21 (facilityNotSupported)          → *FacilityNotSupParam
//	errorCode=34 (systemFailure)                 → *SystemFailureParam
//	errorCode=35 (dataMissing)                   → *DataMissingParam
//	errorCode=52 (unauthorizedRequestingNetwork) → *UnauthorizedRequestingNetworkParam
//
// Returns (nil, nil) for unhandled error codes or when data is empty,
// so callers can safely call this for every ReturnError without
// branching first.
func ParseReturnErrorParameter(errorCode int64, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	switch errorCode {
	case mapErrorCodeUnknownSubscriber:
		return ParseUnknownSubscriberParam(data)
	case mapErrorCodeAbsentSubscriberSM:
		return ParseAbsentSubscriberSMParam(data)
	case mapErrorCodeRoamingNotAllowed:
		return ParseRoamingNotAllowedParam(data)
	case mapErrorCodeTeleserviceNotProvisioned:
		return ParseTeleservNotProvParam(data)
	case mapErrorCodeCallBarred:
		return ParseCallBarredParam(data)
	case mapErrorCodeFacilityNotSupported:
		return ParseFacilityNotSupParam(data)
	case mapErrorCodeSystemFailure:
		return ParseSystemFailureParam(data)
	case mapErrorCodeDataMissing:
		return ParseDataMissingParam(data)
	case mapErrorCodeUnauthorizedRequestingNetwork:
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
