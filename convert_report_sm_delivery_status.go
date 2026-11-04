// convert_report_sm_delivery_status.go
//
// Top-level converters for ReportSMDeliveryStatus (opCode 47) and its
// response. The SMS-GMSC → HLR call that reports an MT-SMS delivery
// outcome and, for an absent subscriber, arms MNRF + the Message-Waiting
// list so AlertServiceCentre fires when the subscriber returns. Reuses
// the address helpers, the SriSmCorrelationID converter, the
// AbsentSubscriberDiagnosticSM helpers (from InformServiceCentre), the
// NetworkNodeDiameterAddress converters and the NULL-flag helpers, and
// converts the SMServingNodeAddressList of both messages. Marshal/Parse
// entry points live in marshal.go / parse.go.

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
	// TS 29.002 V19.1.0 §17.7.6: indicators exclude the corresponding additional outcomes and diagnostics.
	if r.DeliveryOutcomeIndicator && (r.AdditionalSMDeliveryOutcome != nil || r.AdditionalAbsentSubscriberDiagnosticSM != nil) ||
		r.IpSmGwIndicator && (r.DeliveryOutcomeIndicator || r.IpSmGwSmDeliveryOutcome != nil || r.IpSmGwAbsentSubscriberDiagnosticSM != nil) ||
		r.Smsf3gppDeliveryOutcomeIndicator && (r.Smsf3gppDeliveryOutcome != nil || r.Smsf3gppAbsentSubscriberDiagnosticSM != nil) ||
		r.SmsfNon3gppDeliveryOutcomeIndicator && (r.SmsfNon3gppDeliveryOutcome != nil || r.SmsfNon3gppAbsentSubscriberDiagnosticSM != nil) {
		return nil, ErrReportSMDeliveryStatusConflictingOutcomeFields
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
	// [18] failedSMServingNodes.
	nodes, err := convertSMServingNodeAddressListToWire(r.FailedSMServingNodes, "ReportSMDeliveryStatus.FailedSMServingNodes")
	if err != nil {
		return nil, err
	}
	out.FailedSMServingNodes = nodes

	return out, nil
}

func convertArgToReportSMDeliveryStatus(w *gsm_map.ReportSMDeliveryStatusArg) (*ReportSMDeliveryStatus, error) {
	if w == nil {
		return nil, ErrReportSMDeliveryStatusNil
	}

	msisdn, mNature, mPlan, err := decodeAddressWithDigits(w.Msisdn, ErrReportSMDeliveryStatusMSISDNEmpty)
	if err != nil {
		return nil, fmt.Errorf("decoding ReportSMDeliveryStatus.MSISDN: %w", err)
	}
	sca, scaNature, scaPlan, err := decodeAddressWithDigits(w.ServiceCentreAddress, ErrReportSMDeliveryStatusServiceCentreAddressEmpty)
	if err != nil {
		return nil, fmt.Errorf("decoding ReportSMDeliveryStatus.ServiceCentreAddress: %w", err)
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
	nodes, err := convertWireToSMServingNodeAddressList(w.FailedSMServingNodes, "ReportSMDeliveryStatus.FailedSMServingNodes")
	if err != nil {
		return nil, err
	}
	out.FailedSMServingNodes = nodes

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
	// [0] registeredSMServingNodes.
	nodes, err := convertSMServingNodeAddressListToWire(r.RegisteredSMServingNodes, "ReportSMDeliveryStatusRes.RegisteredSMServingNodes")
	if err != nil {
		return nil, err
	}
	out.RegisteredSMServingNodes = nodes
	return out, nil
}

func convertResToReportSMDeliveryStatusRes(w *gsm_map.ReportSMDeliveryStatusRes) (*ReportSMDeliveryStatusRes, error) {
	if w == nil {
		return nil, ErrReportSMDeliveryStatusResNil
	}
	out := &ReportSMDeliveryStatusRes{}
	if w.StoredMSISDN != nil {
		s, nature, plan, err := decodeAddressWithDigits(*w.StoredMSISDN, ErrReportSMDeliveryStatusResStoredMSISDNDecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding ReportSMDeliveryStatusRes.StoredMSISDN: %w", err)
		}
		out.StoredMSISDN = s
		out.StoredMSISDNNature = nature
		out.StoredMSISDNPlan = plan
	}
	nodes, err := convertWireToSMServingNodeAddressList(w.RegisteredSMServingNodes, "ReportSMDeliveryStatusRes.RegisteredSMServingNodes")
	if err != nil {
		return nil, err
	}
	out.RegisteredSMServingNodes = nodes
	return out, nil
}

// ============================================================================
// SMServingNodeAddressList — 3GPP TS 29.002 V19.1.0 §17.7.6
// ============================================================================

// convertSMServingNodeAddressListToWire converts the SMServingNodeAddressList
// field named by field. A nil list is absent; the codec enforces SEQUENCE
// SIZE (1..5) on a present one.
func convertSMServingNodeAddressListToWire(list []SMServingNodeAddress, field string) (*gsm_map.SMServingNodeAddressList, error) {
	if list == nil {
		return nil, nil
	}
	out := &gsm_map.SMServingNodeAddressList{Values: make([]gsm_map.SMServingNodeAddress, len(list))}
	for i := range list {
		v, err := convertSMServingNodeAddressToWire(&list[i])
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", field, i, err)
		}
		out.Values[i] = *v
	}
	return out, nil
}

// convertWireToSMServingNodeAddressList decodes the SMServingNodeAddressList
// field named by field; nil when absent.
func convertWireToSMServingNodeAddressList(w *gsm_map.SMServingNodeAddressList, field string) ([]SMServingNodeAddress, error) {
	if w == nil {
		return nil, nil
	}
	out := make([]SMServingNodeAddress, len(w.Values))
	for i := range w.Values {
		v, err := convertWireToSMServingNodeAddress(&w.Values[i])
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", field, i, err)
		}
		out[i] = *v
	}
	return out, nil
}

// convertSMServingNodeAddressToWire encodes the SMServingNodeAddress CHOICE:
// exactly one of NetworkNodeNumber or DiameterAddress is set. The address
// nature/plan and Diameter ASCII sender rules apply through the shared
// helpers.
func convertSMServingNodeAddressToWire(a *SMServingNodeAddress) (*gsm_map.SMServingNodeAddress, error) {
	hasNumber := a.NetworkNodeNumber != ""
	hasDiameter := a.DiameterAddress != nil
	switch {
	case hasNumber && hasDiameter:
		return nil, ErrSMServingNodeAddressMultipleAlternatives
	case hasNumber:
		encoded, err := encodeAddressField(a.NetworkNodeNumber, a.NetworkNodeNumberNature, a.NetworkNodeNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding NetworkNodeNumber: %w", err)
		}
		v := gsm_map.NewSMServingNodeAddressNetworkNodeNumber(encoded)
		return &v, nil
	case hasDiameter:
		node, err := convertNetworkNodeDiameterAddressToWire(a.DiameterAddress)
		if err != nil {
			return nil, fmt.Errorf("DiameterAddress: %w", err)
		}
		v := gsm_map.NewSMServingNodeAddressDiameterAddress(*node)
		return &v, nil
	default:
		return nil, ErrSMServingNodeAddressNoAlternative
	}
}

// convertWireToSMServingNodeAddress decodes the SMServingNodeAddress CHOICE,
// keeping the received nature, plan and Diameter octets (3GPP TS 29.002
// V19.1.0 §17.1.4). A networkNode-Number without digits would select
// nothing in the public type and is rejected.
func convertWireToSMServingNodeAddress(w *gsm_map.SMServingNodeAddress) (*SMServingNodeAddress, error) {
	switch w.Choice {
	case gsm_map.SMServingNodeAddressChoiceNetworkNodeNumber:
		num, nature, plan, err := decodeAddressWithDigits(*w.NetworkNodeNumber, ErrSMServingNodeAddressNetworkNodeNumberDecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding NetworkNodeNumber: %w", err)
		}
		return &SMServingNodeAddress{NetworkNodeNumber: num, NetworkNodeNumberNature: nature, NetworkNodeNumberPlan: plan}, nil
	case gsm_map.SMServingNodeAddressChoiceDiameterAddress:
		return &SMServingNodeAddress{DiameterAddress: convertWireToNetworkNodeDiameterAddress(w.DiameterAddress)}, nil
	default:
		return nil, ErrSMServingNodeAddressUnknownAlternative
	}
}
