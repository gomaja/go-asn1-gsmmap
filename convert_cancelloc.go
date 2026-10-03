package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
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
// CHOICE validation is performed up-front by validateCancelLocation; this helper
// assumes its input has been validated and focuses on conversion.
func convertCancelLocationIdentityToWire(id *CancelLocationIdentity) (gsm_map.Identity, error) {
	if id.IMSI != "" {
		imsiBytes, err := encodeIdentityDigits(identityIMSI, id.IMSI)
		if err != nil {
			return gsm_map.Identity{}, fmt.Errorf(errEncodingIMSI, err)
		}
		return gsm_map.NewIdentityImsi(imsiBytes), nil
	}

	imsiBytes, err := encodeIdentityDigits(identityIMSI, id.IMSIWithLMSI.IMSI)
	if err != nil {
		return gsm_map.Identity{}, fmt.Errorf(errEncodingIMSI, err)
	}
	iwl := gsm_map.IMSIWithLMSI{
		Imsi: imsiBytes,
		Lmsi: gsm_map.LMSI(id.IMSIWithLMSI.LMSI),
	}
	return gsm_map.NewIdentityImsiWithLMSI(iwl), nil
}

// convertWireToCancelLocationIdentity decodes the wire-level Identity CHOICE.
func convertWireToCancelLocationIdentity(id gsm_map.Identity) (CancelLocationIdentity, error) {
	switch id.Choice {
	case gsm_map.IdentityChoiceImsi:
		if id.Imsi == nil {
			return CancelLocationIdentity{}, ErrCancelLocIdentityChoiceNoAlternative
		}
		imsi, err := decodeIdentityDigits(identityIMSI, *id.Imsi)
		if err != nil {
			return CancelLocationIdentity{}, fmt.Errorf("decoding IMSI: %w", err)
		}
		return CancelLocationIdentity{IMSI: imsi}, nil
	case gsm_map.IdentityChoiceImsiWithLMSI:
		if id.ImsiWithLMSI == nil {
			return CancelLocationIdentity{}, ErrCancelLocIdentityChoiceNoAlternative
		}

		imsi, err := decodeIdentityDigits(identityIMSI, id.ImsiWithLMSI.Imsi)
		if err != nil {
			return CancelLocationIdentity{}, fmt.Errorf("decoding IMSI: %w", err)
		}
		lmsi := id.ImsiWithLMSI.Lmsi

		return CancelLocationIdentity{
			IMSIWithLMSI: &CancelLocationIMSIWithLMSI{IMSI: imsi, LMSI: HexBytes(lmsi)},
		}, nil
	default:
		return CancelLocationIdentity{}, ErrCancelLocIdentityChoiceNoAlternative
	}
}

// validateCancelLocation checks the Identity CHOICE, sender ENUMERATED
// values, TypeOfUpdate applicability, and mutually exclusive MTRF flags.
func validateCancelLocation(c *CancelLocation) error {
	imsiSet := c.Identity.IMSI != ""
	withLmsiSet := c.Identity.IMSIWithLMSI != nil
	switch {
	case imsiSet && withLmsiSet:
		return ErrCancelLocIdentityChoiceMultiple
	case !imsiSet && !withLmsiSet:
		return ErrCancelLocIdentityChoiceNoAlternative
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
		ct := *c.CancellationType
		arg.CancellationType = &ct
	}

	if c.TypeOfUpdate != nil {
		t := *c.TypeOfUpdate
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
		v := encoded
		arg.NewMSCNumber = &v
	}

	// [4] NewVLR-Number
	if c.NewVLRNumber != "" {
		encoded, err := encodeAddressField(c.NewVLRNumber, c.NewVLRNumberNature, c.NewVLRNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding NewVLRNumber: %w", err)
		}
		v := encoded
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

	// CancellationType and TypeOfUpdate are extensible ENUMERATEDs (3GPP TS
	// 29.002 V19.1.0 §17.7.1). Unknown values are copied unless the
	// TypeOfUpdate applicability rule below rejects the combination.
	if arg.CancellationType != nil {
		ct := *arg.CancellationType
		out.CancellationType = &ct
	}

	if arg.TypeOfUpdate != nil {
		t := *arg.TypeOfUpdate
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
		digits, nature, plan, err := decodeAddressWithDigits(*arg.NewMSCNumber, ErrCancelLocNewMSCNumberDecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding NewMSCNumber: %w", err)
		}
		out.NewMSCNumber = digits
		out.NewMSCNumberNature = nature
		out.NewMSCNumberPlan = plan
	}

	if arg.NewVLRNumber != nil {
		digits, nature, plan, err := decodeAddressWithDigits(*arg.NewVLRNumber, ErrCancelLocNewVLRNumberDecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding NewVLRNumber: %w", err)
		}
		out.NewVLRNumber = digits
		out.NewVLRNumberNature = nature
		out.NewVLRNumberPlan = plan
	}

	if arg.NewLmsi != nil {
		lmsi := *arg.NewLmsi

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
