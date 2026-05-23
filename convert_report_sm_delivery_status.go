// convert_report_sm_delivery_status.go
//
// Top-level converters for ReportSMDeliveryStatus (opCode 47) and its
// response. The SMS-GMSC → HLR call that reports an MT-SMS delivery
// outcome and, for an absent subscriber, arms MNRF + the Message-Waiting
// list so AlertServiceCentre fires when the subscriber returns. Reuses
// the address helpers, the SriSmCorrelationID converter, the
// AbsentSubscriberDiagnosticSM helpers (from InformServiceCentre), and
// the NULL-flag helpers. Marshal/Parse entry points live in
// marshal.go / parse.go.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// SmDeliveryOutcome is a non-extensible ENUMERATED 0..2; validated on
// both encode and decode.
const (
	smDeliveryOutcomeMin = 0
	smDeliveryOutcomeMax = 2
)

func validateSmDeliveryOutcome(o SmDeliveryOutcome) error {
	if int64(o) < smDeliveryOutcomeMin || int64(o) > smDeliveryOutcomeMax {
		return fmt.Errorf("value=%d: %w", int64(o), ErrReportSMDeliveryStatusOutcomeInvalid)
	}
	return nil
}

// optOutcomeToWire converts an optional SmDeliveryOutcome to the wire
// pointer, range-checking when present.
func optOutcomeToWire(o *SmDeliveryOutcome) (*gsm_map.SMDeliveryOutcome, error) {
	if o == nil {
		return nil, nil
	}
	if err := validateSmDeliveryOutcome(*o); err != nil {
		return nil, err
	}
	v := *o
	return &v, nil
}

// optOutcomeFromWire converts an optional wire SmDeliveryOutcome back to
// the public pointer, range-checking when present.
func optOutcomeFromWire(o *gsm_map.SMDeliveryOutcome) (*SmDeliveryOutcome, error) {
	if o == nil {
		return nil, nil
	}
	if err := validateSmDeliveryOutcome(*o); err != nil {
		return nil, err
	}
	v := *o
	return &v, nil
}

func convertReportSMDeliveryStatusToArg(r *ReportSMDeliveryStatus) (*gsm_map.ReportSMDeliveryStatusArg, error) {
	if r == nil {
		return nil, ErrReportSMDeliveryStatusNil
	}
	if r.MSISDN == "" {
		return nil, ErrReportSMDeliveryStatusMSISDNEmpty
	}
	if r.ServiceCentreAddress == "" {
		return nil, ErrReportSMDeliveryStatusSCAEmpty
	}
	if err := validateSmDeliveryOutcome(r.SmDeliveryOutcome); err != nil {
		return nil, fmt.Errorf("ReportSMDeliveryStatus.SmDeliveryOutcome: %w", err)
	}

	msisdn, err := encodeAddressField(r.MSISDN, r.MSISDNNature, r.MSISDNPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding ReportSMDeliveryStatus.MSISDN: %w", err)
	}
	sca, err := encodeAddressField(r.ServiceCentreAddress, r.SCANature, r.SCAPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding ReportSMDeliveryStatus.ServiceCentreAddress: %w", err)
	}

	out := &gsm_map.ReportSMDeliveryStatusArg{
		Msisdn:               gsm_map.ISDNAddressString(msisdn),
		ServiceCentreAddress: gsm_map.AddressString(sca),
		SmDeliveryOutcome:    r.SmDeliveryOutcome,
	}

	// [0] / [5] / [8] / [14] / [17] absent-subscriber diagnostics.
	if out.AbsentSubscriberDiagnosticSM, err = absentDiagToWire("ReportSMDeliveryStatus.AbsentSubscriberDiagnosticSM", r.AbsentSubscriberDiagnosticSM); err != nil {
		return nil, err
	}
	if out.AdditionalAbsentSubscriberDiagnosticSM, err = absentDiagToWire("ReportSMDeliveryStatus.AdditionalAbsentSubscriberDiagnosticSM", r.AdditionalAbsentSubscriberDiagnosticSM); err != nil {
		return nil, err
	}
	if out.IpSmGwAbsentSubscriberDiagnosticSM, err = absentDiagToWire("ReportSMDeliveryStatus.IpSmGwAbsentSubscriberDiagnosticSM", r.IpSmGwAbsentSubscriberDiagnosticSM); err != nil {
		return nil, err
	}
	if out.Smsf3gppAbsentSubscriberDiagSM, err = absentDiagToWire("ReportSMDeliveryStatus.Smsf3gppAbsentSubscriberDiagnosticSM", r.Smsf3gppAbsentSubscriberDiagnosticSM); err != nil {
		return nil, err
	}
	if out.SmsfNon3gppAbsentSubscriberDiagSM, err = absentDiagToWire("ReportSMDeliveryStatus.SmsfNon3gppAbsentSubscriberDiagnosticSM", r.SmsfNon3gppAbsentSubscriberDiagnosticSM); err != nil {
		return nil, err
	}

	// [4] / [7] / [13] / [16] additional outcomes.
	if out.AdditionalSMDeliveryOutcome, err = optOutcomeToWire(r.AdditionalSMDeliveryOutcome); err != nil {
		return nil, fmt.Errorf("ReportSMDeliveryStatus.AdditionalSMDeliveryOutcome: %w", err)
	}
	if out.IpSmGwSmDeliveryOutcome, err = optOutcomeToWire(r.IpSmGwSmDeliveryOutcome); err != nil {
		return nil, fmt.Errorf("ReportSMDeliveryStatus.IpSmGwSmDeliveryOutcome: %w", err)
	}
	if out.Smsf3gppDeliveryOutcome, err = optOutcomeToWire(r.Smsf3gppDeliveryOutcome); err != nil {
		return nil, fmt.Errorf("ReportSMDeliveryStatus.Smsf3gppDeliveryOutcome: %w", err)
	}
	if out.SmsfNon3gppDeliveryOutcome, err = optOutcomeToWire(r.SmsfNon3gppDeliveryOutcome); err != nil {
		return nil, fmt.Errorf("ReportSMDeliveryStatus.SmsfNon3gppDeliveryOutcome: %w", err)
	}

	// NULL flags.
	out.GprsSupportIndicator = boolToNullPtr(r.GprsSupportIndicator)
	out.DeliveryOutcomeIndicator = boolToNullPtr(r.DeliveryOutcomeIndicator)
	out.IpSmGwIndicator = boolToNullPtr(r.IpSmGwIndicator)
	out.SingleAttemptDelivery = boolToNullPtr(r.SingleAttemptDelivery)
	out.Smsf3gppDeliveryOutcomeIndicator = boolToNullPtr(r.Smsf3gppDeliveryOutcomeIndicator)
	out.SmsfNon3gppDeliveryOutcomeIndicator = boolToNullPtr(r.SmsfNon3gppDeliveryOutcomeIndicator)

	// [9] imsi (digit count bounded as elsewhere in the package:
	// TBCD-STRING SIZE 3..8 octets = 5..15 BCD digits).
	if r.IMSI != "" {
		if len(r.IMSI) < pslIMSIDigitsMin || len(r.IMSI) > pslIMSIDigitsMax {
			return nil, fmt.Errorf("ReportSMDeliveryStatus.IMSI digits=%d: %w", len(r.IMSI), ErrReportSMDeliveryStatusIMSIInvalidSize)
		}
		imsiBytes, err := tbcd.Encode(r.IMSI)
		if err != nil {
			return nil, fmt.Errorf("encoding ReportSMDeliveryStatus.IMSI: %w", err)
		}
		v := gsm_map.IMSI(imsiBytes)
		out.Imsi = &v
	}
	// [11] correlationID.
	if r.CorrelationID != nil {
		cid, err := convertCorrelationIDToWire(r.CorrelationID)
		if err != nil {
			return nil, fmt.Errorf("ReportSMDeliveryStatus.CorrelationID: %w", err)
		}
		out.CorrelationID = cid
	}

	return out, nil
}

func convertArgToReportSMDeliveryStatus(w *gsm_map.ReportSMDeliveryStatusArg) (*ReportSMDeliveryStatus, error) {
	if w == nil {
		return nil, ErrReportSMDeliveryStatusNil
	}

	msisdn, mNature, mPlan, err := decodeAddressField([]byte(w.Msisdn))
	if err != nil {
		return nil, fmt.Errorf("decoding ReportSMDeliveryStatus.MSISDN: %w", err)
	}
	if msisdn == "" {
		return nil, ErrReportSMDeliveryStatusMSISDNDecodedEmpty
	}
	sca, scaNature, scaPlan, err := decodeAddressField([]byte(w.ServiceCentreAddress))
	if err != nil {
		return nil, fmt.Errorf("decoding ReportSMDeliveryStatus.ServiceCentreAddress: %w", err)
	}
	if sca == "" {
		return nil, ErrReportSMDeliveryStatusSCADecodedEmpty
	}
	if err := validateSmDeliveryOutcome(w.SmDeliveryOutcome); err != nil {
		return nil, fmt.Errorf("ReportSMDeliveryStatus.SmDeliveryOutcome: %w", err)
	}

	out := &ReportSMDeliveryStatus{
		MSISDN:               msisdn,
		MSISDNNature:         mNature,
		MSISDNPlan:           mPlan,
		ServiceCentreAddress: sca,
		SCANature:            scaNature,
		SCAPlan:              scaPlan,
		SmDeliveryOutcome:    w.SmDeliveryOutcome,
	}

	if out.AbsentSubscriberDiagnosticSM, err = absentDiagFromWire("ReportSMDeliveryStatus.AbsentSubscriberDiagnosticSM", w.AbsentSubscriberDiagnosticSM); err != nil {
		return nil, err
	}
	if out.AdditionalAbsentSubscriberDiagnosticSM, err = absentDiagFromWire("ReportSMDeliveryStatus.AdditionalAbsentSubscriberDiagnosticSM", w.AdditionalAbsentSubscriberDiagnosticSM); err != nil {
		return nil, err
	}
	if out.IpSmGwAbsentSubscriberDiagnosticSM, err = absentDiagFromWire("ReportSMDeliveryStatus.IpSmGwAbsentSubscriberDiagnosticSM", w.IpSmGwAbsentSubscriberDiagnosticSM); err != nil {
		return nil, err
	}
	if out.Smsf3gppAbsentSubscriberDiagnosticSM, err = absentDiagFromWire("ReportSMDeliveryStatus.Smsf3gppAbsentSubscriberDiagnosticSM", w.Smsf3gppAbsentSubscriberDiagSM); err != nil {
		return nil, err
	}
	if out.SmsfNon3gppAbsentSubscriberDiagnosticSM, err = absentDiagFromWire("ReportSMDeliveryStatus.SmsfNon3gppAbsentSubscriberDiagnosticSM", w.SmsfNon3gppAbsentSubscriberDiagSM); err != nil {
		return nil, err
	}

	if out.AdditionalSMDeliveryOutcome, err = optOutcomeFromWire(w.AdditionalSMDeliveryOutcome); err != nil {
		return nil, fmt.Errorf("ReportSMDeliveryStatus.AdditionalSMDeliveryOutcome: %w", err)
	}
	if out.IpSmGwSmDeliveryOutcome, err = optOutcomeFromWire(w.IpSmGwSmDeliveryOutcome); err != nil {
		return nil, fmt.Errorf("ReportSMDeliveryStatus.IpSmGwSmDeliveryOutcome: %w", err)
	}
	if out.Smsf3gppDeliveryOutcome, err = optOutcomeFromWire(w.Smsf3gppDeliveryOutcome); err != nil {
		return nil, fmt.Errorf("ReportSMDeliveryStatus.Smsf3gppDeliveryOutcome: %w", err)
	}
	if out.SmsfNon3gppDeliveryOutcome, err = optOutcomeFromWire(w.SmsfNon3gppDeliveryOutcome); err != nil {
		return nil, fmt.Errorf("ReportSMDeliveryStatus.SmsfNon3gppDeliveryOutcome: %w", err)
	}

	out.GprsSupportIndicator = nullPtrToBool(w.GprsSupportIndicator)
	out.DeliveryOutcomeIndicator = nullPtrToBool(w.DeliveryOutcomeIndicator)
	out.IpSmGwIndicator = nullPtrToBool(w.IpSmGwIndicator)
	out.SingleAttemptDelivery = nullPtrToBool(w.SingleAttemptDelivery)
	out.Smsf3gppDeliveryOutcomeIndicator = nullPtrToBool(w.Smsf3gppDeliveryOutcomeIndicator)
	out.SmsfNon3gppDeliveryOutcomeIndicator = nullPtrToBool(w.SmsfNon3gppDeliveryOutcomeIndicator)

	if w.Imsi != nil {
		imsi, err := tbcd.Decode(*w.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding ReportSMDeliveryStatus.IMSI: %w", err)
		}
		if imsi == "" {
			return nil, ErrReportSMDeliveryStatusIMSIDecodedEmpty
		}
		if len(imsi) < pslIMSIDigitsMin || len(imsi) > pslIMSIDigitsMax {
			return nil, fmt.Errorf("ReportSMDeliveryStatus.IMSI digits=%d: %w", len(imsi), ErrReportSMDeliveryStatusIMSIInvalidSize)
		}
		out.IMSI = imsi
	}
	if w.CorrelationID != nil {
		cid, err := convertWireToCorrelationID(w.CorrelationID)
		if err != nil {
			return nil, fmt.Errorf("ReportSMDeliveryStatus.CorrelationID: %w", err)
		}
		out.CorrelationID = cid
	}

	return out, nil
}

// ============================================================================
// ReportSMDeliveryStatus-Res
// ============================================================================

func convertReportSMDeliveryStatusResToRes(r *ReportSMDeliveryStatusRes) (*gsm_map.ReportSMDeliveryStatusRes, error) {
	if r == nil {
		return nil, ErrReportSMDeliveryStatusResNil
	}
	out := &gsm_map.ReportSMDeliveryStatusRes{}
	if r.StoredMSISDN != "" {
		enc, err := encodeAddressField(r.StoredMSISDN, r.StoredMSISDNNature, r.StoredMSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding ReportSMDeliveryStatusRes.StoredMSISDN: %w", err)
		}
		v := gsm_map.ISDNAddressString(enc)
		out.StoredMSISDN = &v
	}
	return out, nil
}

func convertResToReportSMDeliveryStatusRes(w *gsm_map.ReportSMDeliveryStatusRes) (*ReportSMDeliveryStatusRes, error) {
	if w == nil {
		return nil, ErrReportSMDeliveryStatusResNil
	}
	out := &ReportSMDeliveryStatusRes{}
	if w.StoredMSISDN != nil {
		s, nature, plan, err := decodeAddressField([]byte(*w.StoredMSISDN))
		if err != nil {
			return nil, fmt.Errorf("decoding ReportSMDeliveryStatusRes.StoredMSISDN: %w", err)
		}
		if s == "" {
			return nil, ErrReportSMDeliveryStatusResStoredMSISDNEmpty
		}
		out.StoredMSISDN = s
		out.StoredMSISDNNature = nature
		out.StoredMSISDNPlan = plan
	}
	return out, nil
}
