// convert_ussd.go
//
// Converters between the public USSDArg / USSDRes and the go-asn1 wire types
// USSD-Arg / USSD-Res, 3GPP TS 29.002 V19.1.0 §17.7.4.

package gsmmap

import (
	"fmt"
	"slices"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// validateAlertingPatternToWire accepts the eight values defined in
// 3GPP TS 29.002 V19.1.0 §17.7.8; decode is lenient, encode is strict.
func validateAlertingPatternToWire(p AlertingPattern) error {
	switch p {
	case AlertingLevel0, AlertingLevel1, AlertingLevel2,
		AlertingCategory1, AlertingCategory2, AlertingCategory3, AlertingCategory4, AlertingCategory5:
		return nil
	}
	return fmt.Errorf("value 0x%02X: %w", uint8(p), ErrAlertingPatternReserved)
}

func convertUSSDArgToWire(a *USSDArg) (*gsm_map.USSDArg, error) {
	if a == nil {
		return nil, ErrUSSDArgNil
	}
	out := &gsm_map.USSDArg{
		UssdDataCodingScheme: []byte{byte(a.DataCodingScheme)},
		UssdString:           slices.Clone([]byte(a.USSDString)),
	}
	if a.AlertingPattern != nil {
		if err := validateAlertingPatternToWire(*a.AlertingPattern); err != nil {
			return nil, fmt.Errorf("USSDArg.AlertingPattern: %w", err)
		}
		out.AlertingPattern = &gsm_map.AlertingPattern{byte(*a.AlertingPattern)}
	}
	if a.MSISDN != "" {
		enc, err := encodeAddressField(a.MSISDN, a.MSISDNNature, a.MSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding USSDArg.MSISDN: %w", err)
		}
		v := gsm_map.ISDNAddressString(enc)
		out.Msisdn = &v
	}
	return out, nil
}

func convertWireToUSSDArg(w *gsm_map.USSDArg) (*USSDArg, error) {
	if w == nil {
		return nil, ErrUSSDArgNil
	}
	dcs := USSDDataCodingScheme(w.UssdDataCodingScheme[0])
	out := &USSDArg{
		DataCodingScheme: dcs,
		USSDString:       slices.Clone([]byte(w.UssdString)),
	}
	if w.AlertingPattern != nil {
		p := AlertingPattern((*w.AlertingPattern)[0])
		out.AlertingPattern = &p
	}
	if w.Msisdn != nil {
		digits, nature, plan, err := decodeAddressField([]byte(*w.Msisdn))
		if err != nil {
			return nil, fmt.Errorf("decoding USSDArg.MSISDN: %w", err)
		}
		if digits == "" {
			return nil, ErrUSSDMSISDNDecodedEmpty
		}
		out.MSISDN = digits
		out.MSISDNNature = nature
		out.MSISDNPlan = plan
	}
	return out, nil
}

func convertUSSDResToWire(r *USSDRes) (*gsm_map.USSDRes, error) {
	if r == nil {
		return nil, ErrUSSDResNil
	}
	return &gsm_map.USSDRes{
		UssdDataCodingScheme: []byte{byte(r.DataCodingScheme)},
		UssdString:           slices.Clone([]byte(r.USSDString)),
	}, nil
}

func convertWireToUSSDRes(w *gsm_map.USSDRes) (*USSDRes, error) {
	if w == nil {
		return nil, ErrUSSDResNil
	}
	dcs := USSDDataCodingScheme(w.UssdDataCodingScheme[0])
	return &USSDRes{
		DataCodingScheme: dcs,
		USSDString:       slices.Clone([]byte(w.UssdString)),
	}, nil
}
