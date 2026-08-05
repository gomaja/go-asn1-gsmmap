package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-sms/encoding/tpdu"
)

func validateMtForwardSMArgTPDU(p tpdu.TPDU) error {
	// 3GPP TS 29.002 v19.1.0 MAP-SM-DataTypes.asn defines
	// MT-ForwardSM-Arg sm-RP-UI as SignalInfo. Its SMS TPDU must be
	// SC-to-MS per 3GPP TS 23.040 v19.0.0 clause 9.2.2.
	smsType := p.SmsType()
	switch smsType {
	case tpdu.SmsDeliver, tpdu.SmsSubmitReport, tpdu.SmsStatusReport:
		return nil
	default:
		return fmt.Errorf("%w: got %s", ErrMtFsmUnexpectedTPDUType, smsType)
	}
}

func validateMoForwardSMArgTPDU(p tpdu.TPDU) error {
	// 3GPP TS 29.002 v19.1.0 MAP-SM-DataTypes.asn defines
	// MO-ForwardSM-Arg sm-RP-UI as SignalInfo. Its SMS TPDU must be
	// MS-to-SC per 3GPP TS 23.040 v19.0.0 clause 9.2.2.
	smsType := p.SmsType()
	switch smsType {
	case tpdu.SmsSubmit, tpdu.SmsCommand:
		return nil
	default:
		return fmt.Errorf("%w: got %s", ErrMoFsmUnexpectedTPDUType, smsType)
	}
}
