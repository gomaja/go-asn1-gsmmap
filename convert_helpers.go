package gsmmap

import (
	"fmt"
	"math/big"

	identity "github.com/gomaja/go-3gpp-identity"
	"github.com/gomaja/go-asn1/runtime/ber"
	"github.com/gomaja/go-tbcd"

	"github.com/gomaja/go-asn1-gsmmap/address"
)

const errEncodingIMSI = "encoding IMSI: %w"

// keepFirstOctet returns a copy of the first octet of a received
// Ext-SS-Status (3GPP TS 29.002 V19.1.0 §17.7.8) or Ext-ForwOptions
// (§17.7.1), whose octets 2..5 a receiver discards.
func keepFirstOctet(v []byte) HexBytes {
	if len(v) == 0 {
		return nil
	}
	return HexBytes{v[0]}
}

// emlppPriorityFromWire maps spare EMLPP-Priority values to priority 4
// (3GPP TS 29.002 V19.1.0 §17.7.8). ASN.1 range checks belong to the codec.
func emlppPriorityFromWire(v int64) int {
	if v >= 7 && v <= 15 {
		return 4
	}
	return int(v)
}

// encodeAddressField encodes digits into an AddressString (3GPP TS 29.002
// V19.1.0 §17.7.8) with the given nature of address (address.Nature*,
// bits 7..5) and numbering plan (address.Plan*, bits 4..1). Zero is
// address.NatureUnknown / address.PlanUnknown, exactly as on the wire, so a
// decoded nature and plan are retained; encoding canonicalizes the
// extension bit and drops trailing all-filler TBCD octets.
func encodeAddressField(digits string, nature, plan uint8) ([]byte, error) {
	if nature&^0b01110000 != 0 || nature == 0x50 || nature == 0x70 {
		return nil, fmt.Errorf("nature of address 0x%02X: %w", nature, ErrAddressNatureInvalid)
	}
	if plan != 0 && plan != 1 && plan != 3 && plan != 4 && plan != 6 && plan != 8 && plan != 9 {
		return nil, fmt.Errorf("numbering plan 0x%02X: %w", plan, ErrAddressPlanInvalid)
	}
	tbcdBytes, err := tbcd.Encode(digits)
	if err != nil {
		return nil, err
	}
	return address.Encode(address.ExtensionNo, nature, plan, tbcdBytes), nil
}

// RFC 6733 §4.3.1: DiameterIdentity is ASCII; callers supply valid node/realm names.
func validateDiameterIdentityASCII(raw []byte, field string) error {
	for i, b := range raw {
		if b > 0x7f {
			return fmt.Errorf("%s octet %d: %w", field, i, ErrDiameterIdentityNotASCII)
		}
	}
	return nil
}

// validatePLMNIdentityDigits checks the PLMN-Id that raw starts with (3GPP
// TS 29.002 V19.1.0 §§17.7.8, 17.7.13): only MNC digit 3 may be filler. A
// two-octet AreaIdentification holds only the MCC (§17.7.13). The codec
// checks each field's SIZE.
func validatePLMNIdentityDigits(raw []byte, field string) error {
	if len(raw) >= 3 {
		if err := identity.ValidatePLMNIdentity(raw[:3]); err != nil {
			return fmt.Errorf("%s: %w", field, err)
		}
		return nil
	}
	if len(raw) == 2 {
		for i, d := range [3]byte{raw[0] & 0xf, raw[0] >> 4, raw[1] & 0xf} {
			if d > 9 {
				return fmt.Errorf("%s MCC digit %d: %w", field, i+1, identity.ErrPLMNIdentityDigits)
			}
		}
	}
	return nil
}

// decodeAddressField decodes an AddressString byte slice into a phone number string and address components.
// Every AddressString it receives has at least the nature/plan octet: the
// codec checks SIZE (1..n) for fields and SEQUENCE OF elements
// (3GPP TS 29.002 V19.1.0 §17.7.8).
func decodeAddressField(encoded []byte) (digits string, nature, plan uint8, err error) {
	_, nat, pl, rawDigits := address.Decode(encoded)
	digits, err = tbcd.Decode(rawDigits)
	if err != nil {
		return "", 0, 0, err
	}
	return digits, nat, pl, nil
}

// decodeAddressWithDigits decodes an AddressString that is present on the
// wire: a mandatory field, a selected CHOICE alternative or a present
// OPTIONAL field. The public types hold an address as its digits, ""
// meaning missing or absent, so an address with only its nature/plan octet
// (SIZE (1..n) allows it, 3GPP TS 29.002 V19.1.0 §17.7.8) or only filler
// would decode to a value Marshal rejects (mandatory) or drops (OPTIONAL).
// It is rejected with empty: for a mandatory field the sentinel the encoder
// returns for the missing field, otherwise the field's "...DecodedEmpty"
// sentinel. A parsed message therefore marshals back to the same fields.
func decodeAddressWithDigits(encoded []byte, empty error) (digits string, nature, plan uint8, err error) {
	digits, nature, plan, err = decodeAddressField(encoded)
	if err != nil {
		return "", 0, 0, err
	}
	if digits == "" {
		return "", 0, 0, empty
	}
	return digits, nature, plan, nil
}

// boolToNullPtr converts a Go bool into the ASN.1 NULL pointer convention
// used by go-asn1: nil means "absent", non-nil means "present".
func boolToNullPtr(b bool) *struct{} {
	if !b {
		return nil
	}
	v := struct{}{}
	return &v
}

// nullPtrToBool is the inverse of boolToNullPtr.
func nullPtrToBool(p *struct{}) bool { return p != nil }

// intPtrTo64 narrows a *int public-type field to the *int64 wire-type form.
func intPtrTo64(p *int) *int64 {
	if p == nil {
		return nil
	}
	v := int64(*p)
	return &v
}

// int64PtrTo converts an optional INTEGER wire field to the *int public
// form. Its callers pass types whose range the codec enforces and a 32-bit
// int holds: AgeOfLocationInformation (0..32767, 3GPP TS 29.002 V19.1.0
// §17.7.8) and IST-AlertTimerValue (15..255, §17.7.1).
func int64PtrTo(p *int64) *int {
	if p == nil {
		return nil
	}
	v := int(*p)
	return &v
}

func bigIntFromInt64(v int64) *big.Int {
	return big.NewInt(v)
}

func int64FromBigInt(v *big.Int, field string) (int64, error) {
	if v == nil {
		return 0, fmt.Errorf("%w: %s is nil", ErrIntegerMissing, field)
	}
	if !v.IsInt64() {
		return 0, fmt.Errorf("%w: %s overflows int64: %s", ErrInt64Overflow, field, v.String())
	}
	return v.Int64(), nil
}

// identityField is an identity of 3GPP TS 23.003 carried as a TBCD-STRING
// (3GPP TS 29.002 V19.1.0 §17.7.8). The octet SIZE the codec checks admits
// more digit counts than the identity has, so go-3gpp-identity checks the
// decoded digits: received on Parse, sent on Marshal.
type identityField struct {
	received, sent func(digits string) error
}

var (
	// identityIMSI: 6 to 15 digits, 3GPP TS 23.003 V20.1.0 §§2.2, 2.3.
	identityIMSI = identityField{received: identity.ValidateIMSI, sent: identity.ValidateIMSI}

	// identityIMEI: 3GPP TS 23.003 V20.1.0 §6.2.1 composes the IMEI of the
	// TAC ("Its length is 8 digits"), the SNR ("Its length is 6 digits") and
	// the CD/SD, 15 digits; §6.2.2 composes the IMEISV of the TAC, the SNR
	// and the SVN ("Its length is 2 digits"), 16 digits. A MAP field of type
	// IMEI holds either: 3GPP TS 29.002 V19.1.0 §17.7.8 IMEI "Refers to
	// International Mobile Station Equipment Identity and Software Version
	// Number (SVN) [...] If the SVN is not present the last octet shall
	// contain the digit 0 and a filler. If present the SVN shall be included
	// in the last octet."
	//
	// The 15th digit is therefore the spare digit 0, not the Check Digit:
	// TS 23.003 V20.1.0 §6.2.1 "Check Digit (CD) / Spare Digit (SD): If this
	// is the Check Digit see paragraph below; if this digit is Spare Digit it
	// shall be set to zero, when transmitted by the MS." and "The Check Digit
	// is not part of the digits transmitted". Marshal rejects a 15-digit
	// IMEI that does not end in 0 (identity.ErrIMEISpareDigitNotZero). Parse
	// accepts any 15th digit, so a peer that transmits the Check Digit
	// decodes; such a value does not marshal again.
	identityIMEI = identityField{received: imeiField(identity.ValidateIMEI), sent: imeiField(identity.ValidateIMEISpareDigit)}

	// identityIMEISV: the IMEISV parameter (3GPP TS 29.002 V19.1.0
	// §7.6.2.3a, ADD-Info imeisv), 16 digits per 3GPP TS 23.003 V20.1.0
	// §6.2.2.
	identityIMEISV = identityField{received: identity.ValidateIMEISV, sent: identity.ValidateIMEISV}
)

// imeiField checks the digits of a MAP IMEI field: a 16-digit IMEISV, or
// otherwise the 15-digit IMEI that imei checks.
func imeiField(imei func(string) error) func(string) error {
	return func(digits string) error {
		if len(digits) == 16 {
			return identity.ValidateIMEISV(digits)
		}
		return imei(digits)
	}
}

// encodeIdentityDigits TBCD-encodes the digits of an IMSI, IMEI or IMEISV.
// The TBCD-STRING alphabet of TS 29.002 V19.1.0 §17.7.8 also carries
// * # a b c, but these identities are decimal digit strings (TS 23.003), so
// anything else is rejected with identity.ErrNotDigits, and no digits at all
// with identity.ErrEmpty.
func encodeIdentityDigits(id identityField, digits string) ([]byte, error) {
	if err := id.sent(digits); err != nil {
		return nil, err
	}
	return tbcd.Encode(digits)
}

// decodeIdentityDigits is the inverse of encodeIdentityDigits. Octets that
// are all TBCD filler, which tbcd.Decode drops, are identity.ErrEmpty.
func decodeIdentityDigits(id identityField, raw []byte) (string, error) {
	digits, err := tbcd.Decode(raw)
	if err != nil {
		return "", err
	}
	if err := id.received(digits); err != nil {
		return "", err
	}
	return digits, nil
}

// derCodec is a generated go-asn1 type with BER decoding and DER encoding.
type derCodec[T any] interface {
	*T
	UnmarshalBER([]byte, ...ber.DecodeOption) error
	MarshalDER() ([]byte, error)
}

// unmarshalAsDER decodes the BER in raw, then decodes the value again from
// its DER encoding. MarshalBER sends a decoded value that has not changed in
// the form it was received in (https://github.com/gomaja/go-asn1/issues/132),
// which may have long-form or indefinite lengths or constructed strings; a
// value decoded from DER is sent in DER, which meets the transfer syntax of
// 3GPP TS 29.002 V19.1.0 §17.1.1. The types it is used for contain no named
// BIT STRING, whose trailing 0 bits DER would remove.
func unmarshalAsDER[T any, P derCodec[T]](raw []byte) (T, error) {
	var v T
	if err := P(&v).UnmarshalBER(raw); err != nil {
		return v, err
	}
	der, err := P(&v).MarshalDER()
	if err != nil {
		return v, err
	}
	var out T
	err = P(&out).UnmarshalBER(der)
	return out, err
}
