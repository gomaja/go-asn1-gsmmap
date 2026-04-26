package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// --- AlertServiceCentre (opCode 64) ---

// convertAlertServiceCentreToArg converts the public AlertServiceCentre into
// the wire-level gsm_map.AlertServiceCentreArg. The response carries no
// parameters so there is no matching -ToRes helper.
func convertAlertServiceCentreToArg(a *AlertServiceCentre) (*gsm_map.AlertServiceCentreArg, error) {
	if a.MSISDN == "" {
		return nil, ErrAscMissingMSISDN
	}
	if a.ServiceCentreAddress == "" {
		return nil, ErrAscMissingServiceCentreAddress
	}

	msisdn, err := encodeAddressField(a.MSISDN, a.MSISDNNature, a.MSISDNPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding MSISDN: %w", err)
	}
	sca, err := encodeAddressField(a.ServiceCentreAddress, a.SCANature, a.SCAPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding ServiceCentreAddress: %w", err)
	}

	arg := &gsm_map.AlertServiceCentreArg{
		Msisdn:               gsm_map.ISDNAddressString(msisdn),
		ServiceCentreAddress: gsm_map.AddressString(sca),
	}

	if a.IMSI != "" {
		imsiBytes, err := tbcd.Encode(a.IMSI)
		if err != nil {
			return nil, fmt.Errorf(errEncodingIMSI, err)
		}
		v := gsm_map.IMSI(imsiBytes)
		arg.Imsi = &v
	}

	if a.CorrelationID != nil {
		cid, err := convertCorrelationIDToWire(a.CorrelationID)
		if err != nil {
			return nil, fmt.Errorf("CorrelationID: %w", err)
		}
		arg.CorrelationID = cid
	}

	if len(a.MaximumUeAvailabilityTime) > 0 {
		v := gsm_map.Time(a.MaximumUeAvailabilityTime)
		arg.MaximumUeAvailabilityTime = &v
	}

	if a.SmsGmscAlertEvent != nil {
		ev := *a.SmsGmscAlertEvent
		if ev != SmsGmscAlertMsAvailableForMtSms && ev != SmsGmscAlertMsUnderNewServingNode {
			return nil, ErrAscInvalidSmsGmscAlertEvent
		}
		v := ev
		arg.SmsGmscAlertEvent = &v
	}

	if a.SmsGmscDiameterAddress != nil {
		arg.SmsGmscDiameterAddress = convertNetworkNodeDiameterAddressToWire(a.SmsGmscDiameterAddress)
	}

	if a.NewSGSNNumber != "" {
		encoded, err := encodeAddressField(a.NewSGSNNumber, a.NewSGSNNumberNature, a.NewSGSNNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding NewSGSNNumber: %w", err)
		}
		v := gsm_map.ISDNAddressString(encoded)
		arg.NewSGSNNumber = &v
	}

	if a.NewSGSNDiameterAddress != nil {
		arg.NewSGSNDiameterAddress = convertNetworkNodeDiameterAddressToWire(a.NewSGSNDiameterAddress)
	}

	if a.NewMMENumber != "" {
		encoded, err := encodeAddressField(a.NewMMENumber, a.NewMMENumberNature, a.NewMMENumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding NewMMENumber: %w", err)
		}
		v := gsm_map.ISDNAddressString(encoded)
		arg.NewMMENumber = &v
	}

	if a.NewMMEDiameterAddress != nil {
		arg.NewMMEDiameterAddress = convertNetworkNodeDiameterAddressToWire(a.NewMMEDiameterAddress)
	}

	if a.NewMSCNumber != "" {
		encoded, err := encodeAddressField(a.NewMSCNumber, a.NewMSCNumberNature, a.NewMSCNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding NewMSCNumber: %w", err)
		}
		v := gsm_map.ISDNAddressString(encoded)
		arg.NewMSCNumber = &v
	}

	return arg, nil
}

// convertArgToAlertServiceCentre converts a wire-level
// gsm_map.AlertServiceCentreArg back into the public AlertServiceCentre type.
func convertArgToAlertServiceCentre(arg *gsm_map.AlertServiceCentreArg) (*AlertServiceCentre, error) {
	if len(arg.Msisdn) == 0 {
		return nil, ErrAscMissingMSISDN
	}
	if len(arg.ServiceCentreAddress) == 0 {
		return nil, ErrAscMissingServiceCentreAddress
	}

	msisdn, msisdnNature, msisdnPlan, err := decodeAddressField(arg.Msisdn)
	if err != nil {
		return nil, fmt.Errorf("decoding MSISDN: %w", err)
	}
	sca, scaNature, scaPlan, err := decodeAddressField(arg.ServiceCentreAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding ServiceCentreAddress: %w", err)
	}

	out := &AlertServiceCentre{
		MSISDN:               msisdn,
		MSISDNNature:         msisdnNature,
		MSISDNPlan:           msisdnPlan,
		ServiceCentreAddress: sca,
		SCANature:            scaNature,
		SCAPlan:              scaPlan,
	}

	if arg.Imsi != nil {
		imsi, err := tbcd.Decode(*arg.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding optional IMSI: %w", err)
		}
		out.IMSI = imsi
	}

	if arg.CorrelationID != nil {
		cid, err := convertWireToCorrelationID(arg.CorrelationID)
		if err != nil {
			return nil, fmt.Errorf("decoding CorrelationID: %w", err)
		}
		out.CorrelationID = cid
	}

	if arg.MaximumUeAvailabilityTime != nil {
		out.MaximumUeAvailabilityTime = HexBytes(*arg.MaximumUeAvailabilityTime)
	}

	if arg.SmsGmscAlertEvent != nil {
		ev := *arg.SmsGmscAlertEvent
		if ev != SmsGmscAlertMsAvailableForMtSms && ev != SmsGmscAlertMsUnderNewServingNode {
			return nil, ErrAscInvalidSmsGmscAlertEvent
		}
		out.SmsGmscAlertEvent = &ev
	}

	if arg.SmsGmscDiameterAddress != nil {
		out.SmsGmscDiameterAddress = convertWireToNetworkNodeDiameterAddress(arg.SmsGmscDiameterAddress)
	}

	if arg.NewSGSNNumber != nil {
		digits, nature, plan, err := decodeAddressField(*arg.NewSGSNNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding NewSGSNNumber: %w", err)
		}
		out.NewSGSNNumber = digits
		out.NewSGSNNumberNature = nature
		out.NewSGSNNumberPlan = plan
	}

	if arg.NewSGSNDiameterAddress != nil {
		out.NewSGSNDiameterAddress = convertWireToNetworkNodeDiameterAddress(arg.NewSGSNDiameterAddress)
	}

	if arg.NewMMENumber != nil {
		digits, nature, plan, err := decodeAddressField(*arg.NewMMENumber)
		if err != nil {
			return nil, fmt.Errorf("decoding NewMMENumber: %w", err)
		}
		out.NewMMENumber = digits
		out.NewMMENumberNature = nature
		out.NewMMENumberPlan = plan
	}

	if arg.NewMMEDiameterAddress != nil {
		out.NewMMEDiameterAddress = convertWireToNetworkNodeDiameterAddress(arg.NewMMEDiameterAddress)
	}

	if arg.NewMSCNumber != nil {
		digits, nature, plan, err := decodeAddressField(*arg.NewMSCNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding NewMSCNumber: %w", err)
		}
		out.NewMSCNumber = digits
		out.NewMSCNumberNature = nature
		out.NewMSCNumberPlan = plan
	}

	return out, nil
}
