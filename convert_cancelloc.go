package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// --- CancelLocation (opCode 3) ---

// isValidCancellationType reports whether v is one of the CancellationType
// values defined in 3GPP TS 29.002 (updateProcedure=0, subscriptionWithdraw=1,
// initialAttachProcedure=2).
func isValidCancellationType(v CancellationType) bool {
	switch v {
	case CancellationTypeUpdateProcedure,
		CancellationTypeSubscriptionWithdraw,
		CancellationTypeInitialAttachProcedure:
		return true
	}
	return false
}

// isValidTypeOfUpdate reports whether v is one of the TypeOfUpdate values
// defined in 3GPP TS 29.002 (sgsn-change=0, mme-change=1).
func isValidTypeOfUpdate(v TypeOfUpdate) bool {
	switch v {
	case TypeOfUpdateSgsnChange, TypeOfUpdateMmeChange:
		return true
	}
	return false
}

// convertCancelLocationIdentityToWire encodes the Identity CHOICE.
// All field-level validation (exactly-one-alternative, non-empty nested IMSI,
// LMSI length) is performed up-front by validateCancelLocation; this helper
// assumes its input has been validated and focuses on conversion.
func convertCancelLocationIdentityToWire(id *CancelLocationIdentity) (gsm_map.Identity, error) {
	if id.IMSI != "" {
		imsiBytes, err := tbcd.Encode(id.IMSI)
		if err != nil {
			return gsm_map.Identity{}, fmt.Errorf(errEncodingIMSI, err)
		}
		return gsm_map.NewIdentityImsi(gsm_map.IMSI(imsiBytes)), nil
	}

	imsiBytes, err := tbcd.Encode(id.IMSIWithLMSI.IMSI)
	if err != nil {
		return gsm_map.Identity{}, fmt.Errorf(errEncodingIMSI, err)
	}
	iwl := gsm_map.IMSIWithLMSI{
		Imsi: gsm_map.IMSI(imsiBytes),
		Lmsi: gsm_map.LMSI(id.IMSIWithLMSI.LMSI),
	}
	return gsm_map.NewIdentityImsiWithLMSI(iwl), nil
}

// convertWireToCancelLocationIdentity decodes the wire-level Identity CHOICE.
func convertWireToCancelLocationIdentity(id gsm_map.Identity) (CancelLocationIdentity, error) {
	switch id.Choice {
	case gsm_map.IdentityChoiceImsi:
		if id.Imsi == nil || len(*id.Imsi) == 0 {
			return CancelLocationIdentity{}, ErrCancelLocIdentityChoiceNoAlternative
		}
		imsi, err := tbcd.Decode(*id.Imsi)
		if err != nil {
			return CancelLocationIdentity{}, fmt.Errorf("decoding IMSI: %w", err)
		}
		if imsi == "" {
			return CancelLocationIdentity{}, ErrCancelLocIdentityChoiceNoAlternative
		}
		return CancelLocationIdentity{IMSI: imsi}, nil
	case gsm_map.IdentityChoiceImsiWithLMSI:
		if id.ImsiWithLMSI == nil {
			return CancelLocationIdentity{}, ErrCancelLocIdentityChoiceNoAlternative
		}
		if len(id.ImsiWithLMSI.Imsi) == 0 {
			return CancelLocationIdentity{}, ErrCancelLocIdentityMissingIMSI
		}
		imsi, err := tbcd.Decode(id.ImsiWithLMSI.Imsi)
		if err != nil {
			return CancelLocationIdentity{}, fmt.Errorf("decoding IMSI: %w", err)
		}
		if imsi == "" {
			return CancelLocationIdentity{}, ErrCancelLocIdentityMissingIMSI
		}
		lmsi := []byte(id.ImsiWithLMSI.Lmsi)
		if len(lmsi) != 4 {
			return CancelLocationIdentity{}, ErrCancelLocIdentityInvalidLMSI
		}
		return CancelLocationIdentity{
			IMSIWithLMSI: &CancelLocationIMSIWithLMSI{IMSI: imsi, LMSI: HexBytes(lmsi)},
		}, nil
	default:
		return CancelLocationIdentity{}, ErrCancelLocIdentityChoiceNoAlternative
	}
}

// validateCancelLocation enforces every field-level and cross-field
// constraint on a CancelLocation: the Identity CHOICE (exactly-one
// alternative, non-empty nested IMSI, 4-octet LMSI), enum ranges, the
// TypeOfUpdate applicability rule (only with updateProcedure or
// initialAttachProcedure, per 3GPP TS 29.002), the MTRF mutex, and the
// new-lmsi length. All identity-related errors funnel through the
// CHOICE-specific sentinels for a consistent API.
func validateCancelLocation(c *CancelLocation) error {
	imsiSet := c.Identity.IMSI != ""
	withLmsiSet := c.Identity.IMSIWithLMSI != nil
	switch {
	case imsiSet && withLmsiSet:
		return ErrCancelLocIdentityChoiceMultiple
	case !imsiSet && !withLmsiSet:
		return ErrCancelLocIdentityChoiceNoAlternative
	case withLmsiSet:
		if c.Identity.IMSIWithLMSI.IMSI == "" {
			return ErrCancelLocIdentityMissingIMSI
		}
		if len(c.Identity.IMSIWithLMSI.LMSI) != 4 {
			return ErrCancelLocIdentityInvalidLMSI
		}
	}
	if c.CancellationType != nil && !isValidCancellationType(*c.CancellationType) {
		return ErrCancelLocInvalidCancellationType
	}
	if c.TypeOfUpdate != nil {
		if !isValidTypeOfUpdate(*c.TypeOfUpdate) {
			return ErrCancelLocInvalidTypeOfUpdate
		}
		// Per TS 29.002: TypeOfUpdate is only valid with updateProcedure
		// or initialAttachProcedure.
		if c.CancellationType == nil ||
			(*c.CancellationType != CancellationTypeUpdateProcedure &&
				*c.CancellationType != CancellationTypeInitialAttachProcedure) {
			return ErrCancelLocTypeOfUpdateNotApplicable
		}
	}
	if c.MtrfSupportedAndAuthorized && c.MtrfSupportedAndNotAuthorized {
		return ErrCancelLocMtrfBothSet
	}
	if len(c.NewLMSI) > 0 && len(c.NewLMSI) != 4 {
		return ErrCancelLocInvalidNewLMSI
	}
	return nil
}

// convertCancelLocationToArg converts the public CancelLocation into the
// wire-level gsm_map.CancelLocationArg.
func convertCancelLocationToArg(c *CancelLocation) (*gsm_map.CancelLocationArg, error) {
	if err := validateCancelLocation(c); err != nil {
		return nil, err
	}

	id, err := convertCancelLocationIdentityToWire(&c.Identity)
	if err != nil {
		return nil, err
	}

	arg := &gsm_map.CancelLocationArg{Identity: id}

	if c.CancellationType != nil {
		ct := gsm_map.CancellationType(*c.CancellationType)
		arg.CancellationType = &ct
	}

	if c.TypeOfUpdate != nil {
		t := gsm_map.TypeOfUpdate(*c.TypeOfUpdate)
		arg.TypeOfUpdate = &t
	}

	arg.MtrfSupportedAndAuthorized = boolToNullPtr(c.MtrfSupportedAndAuthorized)
	arg.MtrfSupportedAndNotAuthorized = boolToNullPtr(c.MtrfSupportedAndNotAuthorized)

	// [3] NewMSC-Number
	if c.NewMSCNumber != "" {
		encoded, err := encodeAddressField(c.NewMSCNumber, c.NewMSCNumberNature, c.NewMSCNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding NewMSCNumber: %w", err)
		}
		v := gsm_map.ISDNAddressString(encoded)
		arg.NewMSCNumber = &v
	}

	// [4] NewVLR-Number
	if c.NewVLRNumber != "" {
		encoded, err := encodeAddressField(c.NewVLRNumber, c.NewVLRNumberNature, c.NewVLRNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding NewVLRNumber: %w", err)
		}
		v := gsm_map.ISDNAddressString(encoded)
		arg.NewVLRNumber = &v
	}

	// [5] new-lmsi
	if len(c.NewLMSI) > 0 {
		v := gsm_map.LMSI(c.NewLMSI)
		arg.NewLmsi = &v
	}

	arg.ReattachRequired = boolToNullPtr(c.ReattachRequired)

	return arg, nil
}

// convertArgToCancelLocation converts a wire-level gsm_map.CancelLocationArg
// back into the public CancelLocation type.
func convertArgToCancelLocation(arg *gsm_map.CancelLocationArg) (*CancelLocation, error) {
	id, err := convertWireToCancelLocationIdentity(arg.Identity)
	if err != nil {
		return nil, err
	}

	out := &CancelLocation{Identity: id}

	if arg.CancellationType != nil {
		ct := CancellationType(*arg.CancellationType)
		if !isValidCancellationType(ct) {
			return nil, ErrCancelLocInvalidCancellationType
		}
		out.CancellationType = &ct
	}

	if arg.TypeOfUpdate != nil {
		t := TypeOfUpdate(*arg.TypeOfUpdate)
		if !isValidTypeOfUpdate(t) {
			return nil, ErrCancelLocInvalidTypeOfUpdate
		}
		// TS 29.002: TypeOfUpdate only valid with updateProcedure/initialAttachProcedure.
		if out.CancellationType == nil ||
			(*out.CancellationType != CancellationTypeUpdateProcedure &&
				*out.CancellationType != CancellationTypeInitialAttachProcedure) {
			return nil, ErrCancelLocTypeOfUpdateNotApplicable
		}
		out.TypeOfUpdate = &t
	}

	out.MtrfSupportedAndAuthorized = nullPtrToBool(arg.MtrfSupportedAndAuthorized)
	out.MtrfSupportedAndNotAuthorized = nullPtrToBool(arg.MtrfSupportedAndNotAuthorized)
	if out.MtrfSupportedAndAuthorized && out.MtrfSupportedAndNotAuthorized {
		return nil, ErrCancelLocMtrfBothSet
	}

	if arg.NewMSCNumber != nil {
		digits, nature, plan, err := decodeAddressField(*arg.NewMSCNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding NewMSCNumber: %w", err)
		}
		out.NewMSCNumber = digits
		out.NewMSCNumberNature = nature
		out.NewMSCNumberPlan = plan
	}

	if arg.NewVLRNumber != nil {
		digits, nature, plan, err := decodeAddressField(*arg.NewVLRNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding NewVLRNumber: %w", err)
		}
		out.NewVLRNumber = digits
		out.NewVLRNumberNature = nature
		out.NewVLRNumberPlan = plan
	}

	if arg.NewLmsi != nil {
		lmsi := []byte(*arg.NewLmsi)
		if len(lmsi) != 4 {
			return nil, ErrCancelLocInvalidNewLMSI
		}
		out.NewLMSI = HexBytes(lmsi)
	}

	out.ReattachRequired = nullPtrToBool(arg.ReattachRequired)

	return out, nil
}

// convertCancelLocationResToWire converts the public CancelLocationRes into
// the wire-level gsm_map.CancelLocationRes. The response body is empty in
// practice (only an optional ExtensionContainer is defined).
func convertCancelLocationResToWire(_ *CancelLocationRes) *gsm_map.CancelLocationRes {
	return &gsm_map.CancelLocationRes{}
}

// convertWireToCancelLocationRes converts a wire-level gsm_map.CancelLocationRes
// back into the public CancelLocationRes type.
func convertWireToCancelLocationRes(_ *gsm_map.CancelLocationRes) *CancelLocationRes {
	return &CancelLocationRes{}
}
