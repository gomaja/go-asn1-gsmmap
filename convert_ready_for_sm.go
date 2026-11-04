package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func convertReadyForSMArgToWire(v *ReadyForSMArg) (*gsm_map.ReadyForSMArg, error) {
	if v == nil {
		return nil, ErrReadyForSMArgNil
	}
	imsi, err := encodeIdentityDigits(identityIMSI, v.IMSI)
	if err != nil {
		return nil, fmt.Errorf("IMSI: %w", err)
	}
	w := &gsm_map.ReadyForSMArg{Imsi: imsi, AlertReason: v.AlertReason, AlertReasonIndicator: boolToNullPtr(v.AlertReasonIndicator), AdditionalAlertReasonIndicator: boolToNullPtr(v.AdditionalAlertReasonIndicator)}
	if len(v.MaximumUEAvailabilityTime) > 0 {
		b := gsm_map.Time(v.MaximumUEAvailabilityTime)
		w.MaximumUeAvailabilityTime = &b
	}
	return w, nil
}

func convertWireToReadyForSMArg(w *gsm_map.ReadyForSMArg) (*ReadyForSMArg, error) {
	imsi, err := decodeIdentityDigits(identityIMSI, w.Imsi)
	if err != nil {
		return nil, fmt.Errorf("IMSI: %w", err)
	}
	v := &ReadyForSMArg{IMSI: imsi, AlertReason: w.AlertReason, AlertReasonIndicator: nullPtrToBool(w.AlertReasonIndicator), AdditionalAlertReasonIndicator: nullPtrToBool(w.AdditionalAlertReasonIndicator)}
	if w.MaximumUeAvailabilityTime != nil {
		v.MaximumUEAvailabilityTime = HexBytes(*w.MaximumUeAvailabilityTime)
	}
	return v, nil
}

func convertReadyForSMResToWire(v *ReadyForSMRes) (*gsm_map.ReadyForSMRes, error) {
	if v == nil {
		return nil, ErrReadyForSMResNil
	}
	return &gsm_map.ReadyForSMRes{}, nil
}

func convertWireToReadyForSMRes(_ *gsm_map.ReadyForSMRes) (*ReadyForSMRes, error) {
	return &ReadyForSMRes{}, nil
}
