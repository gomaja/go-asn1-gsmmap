package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/runtime"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// --- InformServiceCentre (opCode 63) ---

// MwStatus: 6 bits per 3GPP TS 29.002.
// Bit 0=scAddressNotIncluded, 1=mnrfSet, 2=mcefSet, 3=mnrgSet, 4=mnr5gSet, 5=mnr5gn3gSet.
func convertMwStatusToBitString(m *MwStatusFlags) runtime.BitString {
	var b byte
	if m.SCAddressNotIncluded {
		b |= 1 << 7
	}
	if m.MnrfSet {
		b |= 1 << 6
	}
	if m.McefSet {
		b |= 1 << 5
	}
	if m.MnrgSet {
		b |= 1 << 4
	}
	if m.Mnr5gSet {
		b |= 1 << 3
	}
	if m.Mnr5gn3gSet {
		b |= 1 << 2
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: 6}
}

func convertBitStringToMwStatus(bs runtime.BitString) *MwStatusFlags {
	m := &MwStatusFlags{}
	if bs.BitLength > 0 {
		m.SCAddressNotIncluded = bs.Has(0)
	}
	if bs.BitLength > 1 {
		m.MnrfSet = bs.Has(1)
	}
	if bs.BitLength > 2 {
		m.McefSet = bs.Has(2)
	}
	if bs.BitLength > 3 {
		m.MnrgSet = bs.Has(3)
	}
	if bs.BitLength > 4 {
		m.Mnr5gSet = bs.Has(4)
	}
	if bs.BitLength > 5 {
		m.Mnr5gn3gSet = bs.Has(5)
	}
	return m
}

func validateAbsentSubscriberDiagnosticSM(p *int) error {
	if p == nil {
		return nil
	}
	if *p < 0 || *p > 255 {
		return ErrIscInvalidAbsentSubscriberDiagnosticSM
	}
	return nil
}

// absentDiagToWire validates and converts a public *int AbsentSubscriberDiagnosticSM
// to the go-asn1 wire type. Returns nil, nil when the input is nil.
func absentDiagToWire(field string, p *int) (*gsm_map.AbsentSubscriberDiagnosticSM, error) {
	if p == nil {
		return nil, nil
	}
	if err := validateAbsentSubscriberDiagnosticSM(p); err != nil {
		return nil, fmt.Errorf("%s: %w", field, err)
	}
	v := gsm_map.AbsentSubscriberDiagnosticSM(int64(*p))
	return &v, nil
}

// absentDiagFromWire validates and converts a wire AbsentSubscriberDiagnosticSM
// back to *int. Range-checks on int64 so 32-bit builds cannot silently narrow
// an oversized value into the 0..255 range.
func absentDiagFromWire(field string, p *gsm_map.AbsentSubscriberDiagnosticSM) (*int, error) {
	if p == nil {
		return nil, nil
	}
	v := int64(*p)
	if v < 0 || v > 255 {
		return nil, fmt.Errorf("%s: %w", field, ErrIscInvalidAbsentSubscriberDiagnosticSM)
	}
	iv := int(v)
	return &iv, nil
}

func convertInformServiceCentreToArg(i *InformServiceCentre) (*gsm_map.InformServiceCentreArg, error) {
	arg := &gsm_map.InformServiceCentreArg{}

	if i.StoredMSISDN != "" {
		encoded, err := encodeAddressField(i.StoredMSISDN, i.StoredMSISDNNature, i.StoredMSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding StoredMSISDN: %w", err)
		}
		v := gsm_map.ISDNAddressString(encoded)
		arg.StoredMSISDN = &v
	}

	if i.MwStatus != nil {
		bs := convertMwStatusToBitString(i.MwStatus)
		arg.MwStatus = &bs
	}

	diagFields := []struct {
		name string
		src  *int
		dst  **gsm_map.AbsentSubscriberDiagnosticSM
	}{
		{"AbsentSubscriberDiagnosticSM", i.AbsentSubscriberDiagnosticSM, &arg.AbsentSubscriberDiagnosticSM},
		{"AdditionalAbsentSubscriberDiagnosticSM", i.AdditionalAbsentSubscriberDiagnosticSM, &arg.AdditionalAbsentSubscriberDiagnosticSM},
		{"Smsf3gppAbsentSubscriberDiagnosticSM", i.Smsf3gppAbsentSubscriberDiagnosticSM, &arg.Smsf3gppAbsentSubscriberDiagnosticSM},
		{"SmsfNon3gppAbsentSubscriberDiagnosticSM", i.SmsfNon3gppAbsentSubscriberDiagnosticSM, &arg.SmsfNon3gppAbsentSubscriberDiagnosticSM},
	}
	for _, f := range diagFields {
		w, err := absentDiagToWire(f.name, f.src)
		if err != nil {
			return nil, err
		}
		*f.dst = w
	}

	return arg, nil
}

func convertArgToInformServiceCentre(arg *gsm_map.InformServiceCentreArg) (*InformServiceCentre, error) {
	out := &InformServiceCentre{}

	if arg.StoredMSISDN != nil {
		digits, nature, plan, err := decodeAddressField(*arg.StoredMSISDN)
		if err != nil {
			return nil, fmt.Errorf("decoding StoredMSISDN: %w", err)
		}
		out.StoredMSISDN = digits
		out.StoredMSISDNNature = nature
		out.StoredMSISDNPlan = plan
	}

	if arg.MwStatus != nil {
		// MW-Status per 3GPP TS 29.002 has SIZE (6..16). Reject malformed
		// wire values outside this range to avoid silently normalizing a
		// short BIT STRING into a valid-looking flag struct.
		if arg.MwStatus.BitLength < 6 || arg.MwStatus.BitLength > 16 {
			return nil, fmt.Errorf("MwStatus: BitLength must be 6..16, got %d", arg.MwStatus.BitLength)
		}
		// Capacity check: BitLength must fit within the provided byte slice.
		if int64(arg.MwStatus.BitLength) > int64(len(arg.MwStatus.Bytes))*8 {
			return nil, fmt.Errorf("MwStatus: BitLength %d exceeds len(Bytes)*8 = %d",
				arg.MwStatus.BitLength, len(arg.MwStatus.Bytes)*8)
		}
		out.MwStatus = convertBitStringToMwStatus(*arg.MwStatus)
	}

	diagFields := []struct {
		name string
		src  *gsm_map.AbsentSubscriberDiagnosticSM
		dst  **int
	}{
		{"AbsentSubscriberDiagnosticSM", arg.AbsentSubscriberDiagnosticSM, &out.AbsentSubscriberDiagnosticSM},
		{"AdditionalAbsentSubscriberDiagnosticSM", arg.AdditionalAbsentSubscriberDiagnosticSM, &out.AdditionalAbsentSubscriberDiagnosticSM},
		{"Smsf3gppAbsentSubscriberDiagnosticSM", arg.Smsf3gppAbsentSubscriberDiagnosticSM, &out.Smsf3gppAbsentSubscriberDiagnosticSM},
		{"SmsfNon3gppAbsentSubscriberDiagnosticSM", arg.SmsfNon3gppAbsentSubscriberDiagnosticSM, &out.SmsfNon3gppAbsentSubscriberDiagnosticSM},
	}
	for _, f := range diagFields {
		v, err := absentDiagFromWire(f.name, f.src)
		if err != nil {
			return nil, err
		}
		*f.dst = v
	}

	return out, nil
}
