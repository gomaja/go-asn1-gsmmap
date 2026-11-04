package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	sms "github.com/gomaja/go-sms"
)

// --- SM-RP-DA/OA converters, shared by MO-ForwardSM and MT-ForwardSM ---

func convertSmRpDaToWire(da *SmRpDa) (gsm_map.SMRPDA, error) {
	count := 0
	if da.IMSI != "" {
		count++
	}
	if len(da.LMSI) > 0 {
		count++
	}
	if da.ServiceCentreAddressDA != "" {
		count++
	}
	if da.NoSmRpDa {
		count++
	}
	if count == 0 {
		return gsm_map.SMRPDA{}, ErrSmRpDaNoAlternative
	}
	if count > 1 {
		return gsm_map.SMRPDA{}, ErrSmRpDaMultipleAlternatives
	}

	switch {
	case da.IMSI != "":
		imsiBytes, err := encodeIdentityDigits(identityIMSI, da.IMSI)
		if err != nil {
			return gsm_map.SMRPDA{}, fmt.Errorf("encoding SmRpDa IMSI: %w", err)
		}
		return gsm_map.NewSMRPDAImsi(imsiBytes), nil
	case len(da.LMSI) > 0:

		return gsm_map.NewSMRPDALmsi(gsm_map.LMSI(da.LMSI)), nil
	case da.ServiceCentreAddressDA != "":
		scaDA, err := encodeAddressField(da.ServiceCentreAddressDA, da.ServiceCentreAddressDANature, da.ServiceCentreAddressDAPlan)
		if err != nil {
			return gsm_map.SMRPDA{}, fmt.Errorf("encoding SmRpDa ServiceCentreAddressDA: %w", err)
		}
		return gsm_map.NewSMRPDAServiceCentreAddressDA(scaDA), nil
	default: // da.NoSmRpDa
		return gsm_map.NewSMRPDANoSMRPDA(struct{}{}), nil
	}
}

func convertWireToSmRpDa(w *gsm_map.SMRPDA) (*SmRpDa, error) {
	da := &SmRpDa{}
	switch w.Choice {
	case gsm_map.SMRPDAChoiceImsi:
		imsi, err := decodeIdentityDigits(identityIMSI, *w.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding SmRpDa IMSI: %w", err)
		}
		da.IMSI = imsi
	case gsm_map.SMRPDAChoiceLmsi:

		da.LMSI = HexBytes(*w.Lmsi)
	case gsm_map.SMRPDAChoiceServiceCentreAddressDA:
		sca, nature, plan, err := decodeAddressWithDigits(*w.ServiceCentreAddressDA, ErrSmRpDaServiceCentreAddressDADecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding SmRpDa ServiceCentreAddressDA: %w", err)
		}
		da.ServiceCentreAddressDA = sca
		da.ServiceCentreAddressDANature = nature
		da.ServiceCentreAddressDAPlan = plan
	case gsm_map.SMRPDAChoiceNoSMRPDA:
		da.NoSmRpDa = true
	default:
		return nil, fmt.Errorf("%w: unexpected SMRPDA choice: %d", ErrSmRpDaUnknownAlternative, w.Choice)
	}
	return da, nil
}

func convertSmRpOaToWire(oa *SmRpOa) (gsm_map.SMRPOA, error) {
	count := 0
	if oa.MSISDN != "" {
		count++
	}
	if oa.ServiceCentreAddressOA != "" {
		count++
	}
	if oa.NoSmRpOa {
		count++
	}
	if count == 0 {
		return gsm_map.SMRPOA{}, ErrSmRpOaNoAlternative
	}
	if count > 1 {
		return gsm_map.SMRPOA{}, ErrSmRpOaMultipleAlternatives
	}

	switch {
	case oa.MSISDN != "":
		msisdn, err := encodeAddressField(oa.MSISDN, oa.MSISDNNature, oa.MSISDNPlan)
		if err != nil {
			return gsm_map.SMRPOA{}, fmt.Errorf("encoding SmRpOa MSISDN: %w", err)
		}
		return gsm_map.NewSMRPOAMsisdn(msisdn), nil
	case oa.ServiceCentreAddressOA != "":
		scaOA, err := encodeAddressField(oa.ServiceCentreAddressOA, oa.ServiceCentreAddressOANature, oa.ServiceCentreAddressOAPlan)
		if err != nil {
			return gsm_map.SMRPOA{}, fmt.Errorf("encoding SmRpOa ServiceCentreAddressOA: %w", err)
		}
		return gsm_map.NewSMRPOAServiceCentreAddressOA(scaOA), nil
	default: // oa.NoSmRpOa
		return gsm_map.NewSMRPOANoSMRPOA(struct{}{}), nil
	}
}

func convertWireToSmRpOa(w *gsm_map.SMRPOA) (*SmRpOa, error) {
	oa := &SmRpOa{}
	switch w.Choice {
	case gsm_map.SMRPOAChoiceMsisdn:
		msisdn, nature, plan, err := decodeAddressWithDigits(*w.Msisdn, ErrSmRpOaMSISDNDecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding SmRpOa MSISDN: %w", err)
		}
		oa.MSISDN = msisdn
		oa.MSISDNNature = nature
		oa.MSISDNPlan = plan
	case gsm_map.SMRPOAChoiceServiceCentreAddressOA:
		sca, nature, plan, err := decodeAddressWithDigits(*w.ServiceCentreAddressOA, ErrSmRpOaServiceCentreAddressOADecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding SmRpOa ServiceCentreAddressOA: %w", err)
		}
		oa.ServiceCentreAddressOA = sca
		oa.ServiceCentreAddressOANature = nature
		oa.ServiceCentreAddressOAPlan = plan
	case gsm_map.SMRPOAChoiceNoSMRPOA:
		oa.NoSmRpOa = true
	default:
		return nil, fmt.Errorf("%w: unexpected SMRPOA choice: %d", ErrSmRpOaUnknownAlternative, w.Choice)
	}
	return oa, nil
}

// --- MO-ForwardSM ---

func convertMoFsmToArg(m *MoFsm) (*gsm_map.MOForwardSMArg, error) {
	// sm-RP-DA and sm-RP-OA are non-OPTIONAL CHOICEs in MO-ForwardSM-Arg
	// per 3GPP TS 29.002 v19.1.0 clause 17.7.6.
	smRpDa, err := convertSmRpDaToWire(&m.SmRpDa)
	if err != nil {
		return nil, err
	}
	smRpOa, err := convertSmRpOaToWire(&m.SmRpOa)
	if err != nil {
		return nil, err
	}

	if err := validateMoForwardSMArgTPDU(m.TPDU); err != nil {
		return nil, err
	}
	tpduBytes, err := m.TPDU.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshaling TPDU: %w", err)
	}

	arg := &gsm_map.MOForwardSMArg{
		SmRPDA: smRpDa,
		SmRPOA: smRpOa,
		SmRPUI: tpduBytes,
	}

	// Optional fields (post-extension marker).
	if m.IMSI != "" {
		imsiBytes, err := encodeIdentityDigits(identityIMSI, m.IMSI)
		if err != nil {
			return nil, fmt.Errorf(errEncodingIMSI, err)
		}
		v := imsiBytes
		arg.Imsi = &v
	}
	if m.CorrelationID != nil {
		cid, err := convertCorrelationIDToWire(m.CorrelationID)
		if err != nil {
			return nil, fmt.Errorf("encoding CorrelationID: %w", err)
		}
		arg.CorrelationID = cid
	}
	if m.SmDeliveryOutcome != nil {
		v := *m.SmDeliveryOutcome
		arg.SmDeliveryOutcome = &v
	}

	return arg, nil
}

func convertArgToMoFsm(arg *gsm_map.MOForwardSMArg) (*MoFsm, error) {
	var moFsm MoFsm

	da, err := convertWireToSmRpDa(&arg.SmRPDA)
	if err != nil {
		return nil, err
	}
	moFsm.SmRpDa = *da
	oa, err := convertWireToSmRpOa(&arg.SmRPOA)
	if err != nil {
		return nil, err
	}
	moFsm.SmRpOa = *oa

	// Unmarshal TPDU
	tpduResult, tpduErr := sms.Unmarshal(arg.SmRPUI, sms.AsMO)
	if tpduErr != nil {
		return nil, fmt.Errorf("unmarshaling TPDU: %w", tpduErr)
	}
	if tpduResult == nil {
		return nil, fmt.Errorf("%w: unmarshaling TPDU: nil result", ErrTPDUDecodedNil)
	}
	if err := validateMoForwardSMArgTPDU(*tpduResult); err != nil {
		return nil, err
	}
	moFsm.TPDU = *tpduResult

	// Optional fields (post-extension marker).
	if arg.Imsi != nil {
		imsi, err := decodeIdentityDigits(identityIMSI, *arg.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding IMSI: %w", err)
		}
		moFsm.IMSI = imsi
	}
	if arg.CorrelationID != nil {
		cid, err := convertWireToCorrelationID(arg.CorrelationID)
		if err != nil {
			return nil, fmt.Errorf("decoding CorrelationID: %w", err)
		}
		moFsm.CorrelationID = cid
	}
	if arg.SmDeliveryOutcome != nil {
		v := *arg.SmDeliveryOutcome
		moFsm.SmDeliveryOutcome = &v
	}

	return &moFsm, nil
}

// --- MO-ForwardSM Response ---

func convertMoFsmResToRes(r *MoFsmRes) *gsm_map.MOForwardSMRes {
	out := &gsm_map.MOForwardSMRes{}
	if len(r.SmRpUI) > 0 {
		v := gsm_map.SignalInfo(r.SmRpUI)
		out.SmRPUI = &v
	}
	return out
}

func convertResToMoFsmRes(res *gsm_map.MOForwardSMRes) *MoFsmRes {
	out := &MoFsmRes{}
	if res.SmRPUI != nil {
		out.SmRpUI = HexBytes(*res.SmRPUI)
	}
	return out
}
