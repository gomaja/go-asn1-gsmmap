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
		b |= 1 << (7 - gsm_map.MWStatusScAddressNotIncluded)
	}
	if m.MnrfSet {
		b |= 1 << (7 - gsm_map.MWStatusMnrfSet)
	}
	if m.McefSet {
		b |= 1 << (7 - gsm_map.MWStatusMcefSet)
	}
	if m.MnrgSet {
		b |= 1 << (7 - gsm_map.MWStatusMnrgSet)
	}
	if m.Mnr5gSet {
		b |= 1 << (7 - gsm_map.MWStatusMnr5gSet)
	}
	if m.Mnr5gn3gSet {
		b |= 1 << (7 - gsm_map.MWStatusMnr5gn3gSet)
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: 6}
}

func convertBitStringToMwStatus(bs runtime.BitString) *MwStatusFlags {
	m := &MwStatusFlags{}
	m.SCAddressNotIncluded = bs.Has(gsm_map.MWStatusScAddressNotIncluded)
	if bs.BitLength > gsm_map.MWStatusMnrfSet {
		m.MnrfSet = bs.Has(gsm_map.MWStatusMnrfSet)
	}
	if bs.BitLength > gsm_map.MWStatusMcefSet {
		m.McefSet = bs.Has(gsm_map.MWStatusMcefSet)
	}
	if bs.BitLength > gsm_map.MWStatusMnrgSet {
		m.MnrgSet = bs.Has(gsm_map.MWStatusMnrgSet)
	}
	if bs.BitLength > gsm_map.MWStatusMnr5gSet {
		m.Mnr5gSet = bs.Has(gsm_map.MWStatusMnr5gSet)
	}
	if bs.BitLength > gsm_map.MWStatusMnr5gn3gSet {
		m.Mnr5gn3gSet = bs.Has(gsm_map.MWStatusMnr5gn3gSet)
	}
	return m
}

// absentDiagToWire converts an optional diagnostic to its wire type.
func absentDiagToWire(p *int) *gsm_map.AbsentSubscriberDiagnosticSM {
	if p == nil {
		return nil
	}
	v := gsm_map.AbsentSubscriberDiagnosticSM(*p)
	return &v
}

// absentDiagFromWire converts an optional wire diagnostic to int.
func absentDiagFromWire(p *gsm_map.AbsentSubscriberDiagnosticSM) *int {
	if p == nil {
		return nil
	}
	v := int(*p)
	return &v
}

func convertInformServiceCentreToArg(i *InformServiceCentre) (*gsm_map.InformServiceCentreArg, error) {
	// TS 29.002 V19.1.0 §17.7.6: the additional diagnostic requires the primary diagnostic.
	if i.AdditionalAbsentSubscriberDiagnosticSM != nil && i.AbsentSubscriberDiagnosticSM == nil {
		return nil, ErrIscAdditionalDiagnosticWithoutPrimary
	}
	arg := &gsm_map.InformServiceCentreArg{}

	if i.StoredMSISDN != "" {
		encoded, err := encodeAddressField(i.StoredMSISDN, i.StoredMSISDNNature, i.StoredMSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding StoredMSISDN: %w", err)
		}
		v := encoded
		arg.StoredMSISDN = &v
	}

	if i.MwStatus != nil {
		bs := convertMwStatusToBitString(i.MwStatus)
		arg.MwStatus = &bs
	}

	diagFields := []struct {
		src *int
		dst **gsm_map.AbsentSubscriberDiagnosticSM
	}{
		{i.AbsentSubscriberDiagnosticSM, &arg.AbsentSubscriberDiagnosticSM},
		{i.AdditionalAbsentSubscriberDiagnosticSM, &arg.AdditionalAbsentSubscriberDiagnosticSM},
		{i.Smsf3gppAbsentSubscriberDiagnosticSM, &arg.Smsf3gppAbsentSubscriberDiagnosticSM},
		{i.SmsfNon3gppAbsentSubscriberDiagnosticSM, &arg.SmsfNon3gppAbsentSubscriberDiagnosticSM},
	}
	for _, f := range diagFields {
		*f.dst = absentDiagToWire(f.src)
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
		if digits == "" {
			return nil, ErrIscStoredMSISDNDecodedEmpty
		}
		out.StoredMSISDN = digits
		out.StoredMSISDNNature = nature
		out.StoredMSISDNPlan = plan
	}

	if arg.MwStatus != nil {
		out.MwStatus = convertBitStringToMwStatus(*arg.MwStatus)
	}

	diagFields := []struct {
		src *gsm_map.AbsentSubscriberDiagnosticSM
		dst **int
	}{
		{arg.AbsentSubscriberDiagnosticSM, &out.AbsentSubscriberDiagnosticSM},
		{arg.AdditionalAbsentSubscriberDiagnosticSM, &out.AdditionalAbsentSubscriberDiagnosticSM},
		{arg.Smsf3gppAbsentSubscriberDiagnosticSM, &out.Smsf3gppAbsentSubscriberDiagnosticSM},
		{arg.SmsfNon3gppAbsentSubscriberDiagnosticSM, &out.SmsfNon3gppAbsentSubscriberDiagnosticSM},
	}
	for _, f := range diagFields {
		*f.dst = absentDiagFromWire(f.src)
	}

	return out, nil
}
