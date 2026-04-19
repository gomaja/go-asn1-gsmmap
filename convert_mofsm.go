package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
	"github.com/warthog618/sms"
)

// --- MO-ForwardSM SM-RP-DA/OA converters ---

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
		return gsm_map.SMRPDA{}, ErrMoFsmSmRpDaNoAlternative
	}
	if count > 1 {
		return gsm_map.SMRPDA{}, ErrMoFsmSmRpDaMultipleAlternatives
	}

	switch {
	case da.IMSI != "":
		imsiBytes, err := tbcd.Encode(da.IMSI)
		if err != nil {
			return gsm_map.SMRPDA{}, fmt.Errorf("encoding SmRpDa IMSI: %w", err)
		}
		return gsm_map.NewSMRPDAImsi(gsm_map.IMSI(imsiBytes)), nil
	case len(da.LMSI) > 0:
		if len(da.LMSI) != 4 {
			return gsm_map.SMRPDA{}, fmt.Errorf("SmRpDa LMSI must be exactly 4 octets, got %d", len(da.LMSI))
		}
		return gsm_map.NewSMRPDALmsi(gsm_map.LMSI(da.LMSI)), nil
	case da.ServiceCentreAddressDA != "":
		scaDA, err := encodeAddressField(da.ServiceCentreAddressDA, da.SCADANature, da.SCADAPlan)
		if err != nil {
			return gsm_map.SMRPDA{}, fmt.Errorf("encoding SmRpDa ServiceCentreAddressDA: %w", err)
		}
		return gsm_map.NewSMRPDAServiceCentreAddressDA(gsm_map.AddressString(scaDA)), nil
	default: // da.NoSmRpDa
		return gsm_map.NewSMRPDANoSMRPDA(struct{}{}), nil
	}
}

func convertWireToSmRpDa(w *gsm_map.SMRPDA) (*SmRpDa, error) {
	da := &SmRpDa{}
	switch w.Choice {
	case gsm_map.SMRPDAChoiceImsi:
		if w.Imsi == nil {
			return nil, fmt.Errorf("SMRPDA IMSI is nil")
		}
		imsi, err := tbcd.Decode(*w.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding SmRpDa IMSI: %w", err)
		}
		da.IMSI = imsi
	case gsm_map.SMRPDAChoiceLmsi:
		if w.Lmsi == nil {
			return nil, fmt.Errorf("SMRPDA LMSI is nil")
		}
		if len(*w.Lmsi) != 4 {
			return nil, fmt.Errorf("SmRpDa LMSI must be exactly 4 octets, got %d", len(*w.Lmsi))
		}
		da.LMSI = HexBytes(*w.Lmsi)
	case gsm_map.SMRPDAChoiceServiceCentreAddressDA:
		if w.ServiceCentreAddressDA == nil {
			return nil, fmt.Errorf("SMRPDA ServiceCentreAddressDA is nil")
		}
		sca, nature, plan, err := decodeAddressField(*w.ServiceCentreAddressDA)
		if err != nil {
			return nil, fmt.Errorf("decoding SmRpDa ServiceCentreAddressDA: %w", err)
		}
		da.ServiceCentreAddressDA = sca
		da.SCADANature = nature
		da.SCADAPlan = plan
	case gsm_map.SMRPDAChoiceNoSMRPDA:
		da.NoSmRpDa = true
	default:
		return nil, fmt.Errorf("unexpected SMRPDA choice: %d", w.Choice)
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
		return gsm_map.SMRPOA{}, ErrMoFsmSmRpOaNoAlternative
	}
	if count > 1 {
		return gsm_map.SMRPOA{}, ErrMoFsmSmRpOaMultipleAlternatives
	}

	switch {
	case oa.MSISDN != "":
		msisdn, err := encodeAddressField(oa.MSISDN, oa.MSISDNNature, oa.MSISDNPlan)
		if err != nil {
			return gsm_map.SMRPOA{}, fmt.Errorf("encoding SmRpOa MSISDN: %w", err)
		}
		return gsm_map.NewSMRPOAMsisdn(gsm_map.ISDNAddressString(msisdn)), nil
	case oa.ServiceCentreAddressOA != "":
		scaOA, err := encodeAddressField(oa.ServiceCentreAddressOA, oa.SCAOANature, oa.SCAOAPlan)
		if err != nil {
			return gsm_map.SMRPOA{}, fmt.Errorf("encoding SmRpOa ServiceCentreAddressOA: %w", err)
		}
		return gsm_map.NewSMRPOAServiceCentreAddressOA(gsm_map.AddressString(scaOA)), nil
	default: // oa.NoSmRpOa
		return gsm_map.NewSMRPOANoSMRPOA(struct{}{}), nil
	}
}

func convertWireToSmRpOa(w *gsm_map.SMRPOA) (*SmRpOa, error) {
	oa := &SmRpOa{}
	switch w.Choice {
	case gsm_map.SMRPOAChoiceMsisdn:
		if w.Msisdn == nil {
			return nil, fmt.Errorf("SMRPOA MSISDN is nil")
		}
		msisdn, nature, plan, err := decodeAddressField(*w.Msisdn)
		if err != nil {
			return nil, fmt.Errorf("decoding SmRpOa MSISDN: %w", err)
		}
		oa.MSISDN = msisdn
		oa.MSISDNNature = nature
		oa.MSISDNPlan = plan
	case gsm_map.SMRPOAChoiceServiceCentreAddressOA:
		if w.ServiceCentreAddressOA == nil {
			return nil, fmt.Errorf("SMRPOA ServiceCentreAddressOA is nil")
		}
		sca, nature, plan, err := decodeAddressField(*w.ServiceCentreAddressOA)
		if err != nil {
			return nil, fmt.Errorf("decoding SmRpOa ServiceCentreAddressOA: %w", err)
		}
		oa.ServiceCentreAddressOA = sca
		oa.SCAOANature = nature
		oa.SCAOAPlan = plan
	case gsm_map.SMRPOAChoiceNoSMRPOA:
		oa.NoSmRpOa = true
	default:
		return nil, fmt.Errorf("unexpected SMRPOA choice: %d", w.Choice)
	}
	return oa, nil
}

// --- MO-ForwardSM ---

func convertMoFsmToArg(m *MoFsm) (*gsm_map.MOForwardSMArg, error) {
	// SM-RP-DA
	var smRpDa gsm_map.SMRPDA
	if m.SmRpDa != nil {
		da, err := convertSmRpDaToWire(m.SmRpDa)
		if err != nil {
			return nil, err
		}
		smRpDa = da
	} else {
		if m.ServiceCentreAddressDA == "" {
			return nil, fmt.Errorf("MoFsm: ServiceCentreAddressDA is empty (set SmRpDa for other DA variants)")
		}
		scaDA, err := encodeAddressField(m.ServiceCentreAddressDA, m.SCADANature, m.SCADAPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding ServiceCentreAddressDA: %w", err)
		}
		smRpDa = gsm_map.NewSMRPDAServiceCentreAddressDA(gsm_map.AddressString(scaDA))
	}

	// SM-RP-OA
	var smRpOa gsm_map.SMRPOA
	if m.SmRpOa != nil {
		oa, err := convertSmRpOaToWire(m.SmRpOa)
		if err != nil {
			return nil, err
		}
		smRpOa = oa
	} else {
		if m.MSISDN == "" {
			return nil, fmt.Errorf("MoFsm: MSISDN is empty (set SmRpOa for other OA variants)")
		}
		msisdn, err := encodeAddressField(m.MSISDN, m.MSISDNNature, m.MSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding MSISDN: %w", err)
		}
		smRpOa = gsm_map.NewSMRPOAMsisdn(gsm_map.ISDNAddressString(msisdn))
	}

	tpduBytes, err := m.TPDU.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshaling TPDU: %w", err)
	}

	arg := &gsm_map.MOForwardSMArg{
		SmRPDA: smRpDa,
		SmRPOA: smRpOa,
		SmRPUI: gsm_map.SignalInfo(tpduBytes),
	}

	// Optional fields (post-extension marker).
	if m.IMSI != "" {
		imsiBytes, err := tbcd.Encode(m.IMSI)
		if err != nil {
			return nil, fmt.Errorf(errEncodingIMSI, err)
		}
		v := gsm_map.IMSI(imsiBytes)
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
		v := gsm_map.SMDeliveryOutcome(*m.SmDeliveryOutcome)
		arg.SmDeliveryOutcome = &v
	}

	return arg, nil
}

func convertArgToMoFsm(arg *gsm_map.MOForwardSMArg) (*MoFsm, error) {
	var moFsm MoFsm

	// Extract SM-RP-DA
	switch arg.SmRPDA.Choice {
	case gsm_map.SMRPDAChoiceServiceCentreAddressDA:
		if arg.SmRPDA.ServiceCentreAddressDA == nil {
			return nil, fmt.Errorf("SMRPDA ServiceCentreAddressDA is nil")
		}
		sca, nature, plan, err := decodeAddressField(*arg.SmRPDA.ServiceCentreAddressDA)
		if err != nil {
			return nil, fmt.Errorf("decoding ServiceCentreAddressDA: %w", err)
		}
		moFsm.ServiceCentreAddressDA = sca
		moFsm.SCADANature = nature
		moFsm.SCADAPlan = plan
	case gsm_map.SMRPDAChoiceImsi, gsm_map.SMRPDAChoiceLmsi, gsm_map.SMRPDAChoiceNoSMRPDA:
		da, err := convertWireToSmRpDa(&arg.SmRPDA)
		if err != nil {
			return nil, err
		}
		moFsm.SmRpDa = da
	default:
		return nil, fmt.Errorf("unexpected SMRPDA choice: %d", arg.SmRPDA.Choice)
	}

	// Extract SM-RP-OA
	switch arg.SmRPOA.Choice {
	case gsm_map.SMRPOAChoiceMsisdn:
		if arg.SmRPOA.Msisdn == nil {
			return nil, fmt.Errorf("SMRPOA MSISDN is nil")
		}
		msisdn, nature, plan, err := decodeAddressField(*arg.SmRPOA.Msisdn)
		if err != nil {
			return nil, fmt.Errorf("decoding MSISDN: %w", err)
		}
		moFsm.MSISDN = msisdn
		moFsm.MSISDNNature = nature
		moFsm.MSISDNPlan = plan
	case gsm_map.SMRPOAChoiceServiceCentreAddressOA, gsm_map.SMRPOAChoiceNoSMRPOA:
		oa, err := convertWireToSmRpOa(&arg.SmRPOA)
		if err != nil {
			return nil, err
		}
		moFsm.SmRpOa = oa
	default:
		return nil, fmt.Errorf("unexpected SMRPOA choice: %d", arg.SmRPOA.Choice)
	}

	// Unmarshal TPDU
	tpduResult, tpduErr := sms.Unmarshal(arg.SmRPUI, sms.AsMO)
	if tpduErr != nil {
		return nil, fmt.Errorf("unmarshaling TPDU: %w", tpduErr)
	}
	if tpduResult == nil {
		return nil, fmt.Errorf("unmarshaling TPDU: nil result")
	}
	moFsm.TPDU = *tpduResult

	// Optional fields (post-extension marker).
	if arg.Imsi != nil {
		imsi, err := tbcd.Decode(*arg.Imsi)
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
		v := SmDeliveryOutcome(*arg.SmDeliveryOutcome)
		moFsm.SmDeliveryOutcome = &v
	}

	return &moFsm, nil
}

// --- MO-ForwardSM Response ---

func convertMoFsmRespToRes(r *MoFsmResp) *gsm_map.MOForwardSMRes {
	out := &gsm_map.MOForwardSMRes{}
	if len(r.SmRpUI) > 0 {
		v := gsm_map.SignalInfo(r.SmRpUI)
		out.SmRPUI = &v
	}
	return out
}

func convertResToMoFsmResp(res *gsm_map.MOForwardSMRes) *MoFsmResp {
	out := &MoFsmResp{}
	if res.SmRPUI != nil {
		out.SmRpUI = HexBytes(*res.SmRPUI)
	}
	return out
}
