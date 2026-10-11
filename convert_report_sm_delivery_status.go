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
)

// optOutcomeToWire converts an optional SmDeliveryOutcome to the wire
// pointer when present (3GPP TS 29.002 V19.1.0 §17.7.6).
func optOutcomeToWire(o *SmDeliveryOutcome) *gsm_map.SMDeliveryOutcome {
	if o == nil {
		return nil
	}
	v := *o
	return &v
}

// optOutcomeFromWire converts an optional wire SmDeliveryOutcome back to
// the public pointer when present (3GPP TS 29.002 V19.1.0 §17.7.6).
func optOutcomeFromWire(o *gsm_map.SMDeliveryOutcome) *SmDeliveryOutcome {
	if o == nil {
		return nil
	}
	v := *o
	return &v
}

func convertReportSMDeliveryStatusToArg(r *ReportSMDeliveryStatus) (*gsm_map.ReportSMDeliveryStatusArg, error) {
	if r == nil {
		return nil, ErrReportSMDeliveryStatusNil
	}
	if r.MSISDN == "" {
		return nil, ErrReportSMDeliveryStatusMSISDNEmpty
	}
	if r.ServiceCentreAddress == "" {
		return nil, ErrReportSMDeliveryStatusServiceCentreAddressEmpty
	}

	msisdn, err := encodeAddressField(r.MSISDN, r.MSISDNNature, r.MSISDNPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding ReportSMDeliveryStatus.MSISDN: %w", err)
	}
	sca, err := encodeAddressField(r.ServiceCentreAddress, r.ServiceCentreAddressNature, r.ServiceCentreAddressPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding ReportSMDeliveryStatus.ServiceCentreAddress: %w", err)
	}

	out := &gsm_map.ReportSMDeliveryStatusArg{
		Msisdn:               msisdn,
		ServiceCentreAddress: sca,
		SmDeliveryOutcome:    r.SmDeliveryOutcome,
	}

	// [0] / [5] / [8] / [14] / [17] absent-subscriber diagnostics.
	out.AbsentSubscriberDiagnosticSM = absentDiagToWire(r.AbsentSubscriberDiagnosticSM)
	out.AdditionalAbsentSubscriberDiagnosticSM = absentDiagToWire(r.AdditionalAbsentSubscriberDiagnosticSM)
	out.IpSmGwAbsentSubscriberDiagnosticSM = absentDiagToWire(r.IpSmGwAbsentSubscriberDiagnosticSM)
	out.Smsf3gppAbsentSubscriberDiagSM = absentDiagToWire(r.Smsf3gppAbsentSubscriberDiagnosticSM)
	out.SmsfNon3gppAbsentSubscriberDiagSM = absentDiagToWire(r.SmsfNon3gppAbsentSubscriberDiagnosticSM)

	// [4] / [7] / [13] / [16] additional outcomes.
	out.AdditionalSMDeliveryOutcome = optOutcomeToWire(r.AdditionalSMDeliveryOutcome)
	out.IpSmGwSmDeliveryOutcome = optOutcomeToWire(r.IpSmGwSmDeliveryOutcome)
	out.Smsf3gppDeliveryOutcome = optOutcomeToWire(r.Smsf3gppDeliveryOutcome)
	out.SmsfNon3gppDeliveryOutcome = optOutcomeToWire(r.SmsfNon3gppDeliveryOutcome)

	// NULL flags.
	out.GprsSupportIndicator = boolToNullPtr(r.GprsSupportIndicator)
	out.DeliveryOutcomeIndicator = boolToNullPtr(r.DeliveryOutcomeIndicator)
	out.IpSmGwIndicator = boolToNullPtr(r.IpSmGwIndicator)
	out.SingleAttemptDelivery = boolToNullPtr(r.SingleAttemptDelivery)
	out.Smsf3gppDeliveryOutcomeIndicator = boolToNullPtr(r.Smsf3gppDeliveryOutcomeIndicator)
	out.SmsfNon3gppDeliveryOutcomeIndicator = boolToNullPtr(r.SmsfNon3gppDeliveryOutcomeIndicator)

	// [9] imsi
	if r.IMSI != "" {
		imsiBytes, err := encodeIdentityDigits(identityIMSI, r.IMSI)
		if err != nil {
			return nil, fmt.Errorf("encoding ReportSMDeliveryStatus.IMSI: %w", err)
		}
		v := imsiBytes
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

	msisdn, mNature, mPlan, err := decodeAddressField(w.Msisdn)
	if err != nil {
		return nil, fmt.Errorf("decoding ReportSMDeliveryStatus.MSISDN: %w", err)
	}
	if msisdn == "" {
		return nil, ErrReportSMDeliveryStatusMSISDNEmpty
	}
	sca, scaNature, scaPlan, err := decodeAddressField(w.ServiceCentreAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding ReportSMDeliveryStatus.ServiceCentreAddress: %w", err)
	}
	if sca == "" {
		return nil, ErrReportSMDeliveryStatusServiceCentreAddressEmpty
	}

	out := &ReportSMDeliveryStatus{
		MSISDN:                     msisdn,
		MSISDNNature:               mNature,
		MSISDNPlan:                 mPlan,
		ServiceCentreAddress:       sca,
		ServiceCentreAddressNature: scaNature,
		ServiceCentreAddressPlan:   scaPlan,
		SmDeliveryOutcome:          w.SmDeliveryOutcome,
	}

	out.AbsentSubscriberDiagnosticSM = absentDiagFromWire(w.AbsentSubscriberDiagnosticSM)
	out.AdditionalAbsentSubscriberDiagnosticSM = absentDiagFromWire(w.AdditionalAbsentSubscriberDiagnosticSM)
	out.IpSmGwAbsentSubscriberDiagnosticSM = absentDiagFromWire(w.IpSmGwAbsentSubscriberDiagnosticSM)
	out.Smsf3gppAbsentSubscriberDiagnosticSM = absentDiagFromWire(w.Smsf3gppAbsentSubscriberDiagSM)
	out.SmsfNon3gppAbsentSubscriberDiagnosticSM = absentDiagFromWire(w.SmsfNon3gppAbsentSubscriberDiagSM)

	out.AdditionalSMDeliveryOutcome = optOutcomeFromWire(w.AdditionalSMDeliveryOutcome)
	out.IpSmGwSmDeliveryOutcome = optOutcomeFromWire(w.IpSmGwSmDeliveryOutcome)
	out.Smsf3gppDeliveryOutcome = optOutcomeFromWire(w.Smsf3gppDeliveryOutcome)
	out.SmsfNon3gppDeliveryOutcome = optOutcomeFromWire(w.SmsfNon3gppDeliveryOutcome)

	out.GprsSupportIndicator = nullPtrToBool(w.GprsSupportIndicator)
	out.DeliveryOutcomeIndicator = nullPtrToBool(w.DeliveryOutcomeIndicator)
	out.IpSmGwIndicator = nullPtrToBool(w.IpSmGwIndicator)
	out.SingleAttemptDelivery = nullPtrToBool(w.SingleAttemptDelivery)
	out.Smsf3gppDeliveryOutcomeIndicator = nullPtrToBool(w.Smsf3gppDeliveryOutcomeIndicator)
	out.SmsfNon3gppDeliveryOutcomeIndicator = nullPtrToBool(w.SmsfNon3gppDeliveryOutcomeIndicator)

	if w.Imsi != nil {
		imsi, err := decodeIdentityDigits(identityIMSI, *w.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding ReportSMDeliveryStatus.IMSI: %w", err)
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
		v := enc
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
		s, nature, plan, err := decodeAddressField(*w.StoredMSISDN)
		if err != nil {
			return nil, fmt.Errorf("decoding ReportSMDeliveryStatusRes.StoredMSISDN: %w", err)
		}
		if s == "" {
			return nil, ErrReportSMDeliveryStatusResStoredMSISDNDecodedEmpty
		}
		out.StoredMSISDN = s
		out.StoredMSISDNNature = nature
		out.StoredMSISDNPlan = plan
	}
	return out, nil
}
