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
	m.SCAddressNotIncluded = bs.Has(0)
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
