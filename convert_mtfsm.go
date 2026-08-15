package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1-gsmmap/tbcd"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	sms "github.com/gomaja/go-sms"
)

// --- MT-ForwardSM ---

func convertMtFsmToArg(m *MtFsm) (*gsm_map.MTForwardSMArg, error) {
	// sm-RP-DA and sm-RP-OA are non-OPTIONAL CHOICEs in MT-ForwardSM-Arg
	// per 3GPP TS 29.002 v19.1.0 clause 17.7.6. IMSI and
	// ServiceCentreAddressOA remain the common public fields; SmRpDa/SmRpOa
	// expose the full CHOICEs.
	var smRpDa gsm_map.SMRPDA
	if m.SmRpDa != nil {
		da, err := convertMtSmRpDaToWire(m.SmRpDa)
		if err != nil {
			return nil, err
		}
		smRpDa = da
	} else {
		if m.IMSI == "" {
			return nil, ErrMtFsmMissingIMSI
		}
		imsiBytes, err := tbcd.Encode(m.IMSI)
		if err != nil {
			return nil, fmt.Errorf(errEncodingIMSI, err)
		}
		smRpDa = gsm_map.NewSMRPDAImsi(gsm_map.IMSI(imsiBytes))
	}

	var smRpOa gsm_map.SMRPOA
	if m.SmRpOa != nil {
		oa, err := convertMtSmRpOaToWire(m.SmRpOa)
		if err != nil {
			return nil, err
		}
		smRpOa = oa
	} else {
		if m.ServiceCentreAddressOA == "" {
			return nil, ErrMtFsmMissingServiceCentreAddressOA
		}
		scaOA, err := encodeAddressField(m.ServiceCentreAddressOA, m.SCAOANature, m.SCAOAPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding ServiceCentreAddressOA: %w", err)
		}
		smRpOa = gsm_map.NewSMRPOAServiceCentreAddressOA(gsm_map.AddressString(scaOA))
	}

	if err := validateMtForwardSMArgTPDU(m.TPDU); err != nil {
		return nil, err
	}
	tpduBytes, err := m.TPDU.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshaling TPDU: %w", err)
	}

	arg := &gsm_map.MTForwardSMArg{
		SmRPDA: smRpDa,
		SmRPOA: smRpOa,
		SmRPUI: gsm_map.SignalInfo(tpduBytes),
	}

	arg.MoreMessagesToSend = boolToNullPtr(m.MoreMessagesToSend)

	// Optional fields (post-extension marker).
	if m.SmDeliveryTimer != nil {
		v := *m.SmDeliveryTimer
		if v < MinSmDeliveryTimer || v > MaxSmDeliveryTimer {
			return nil, ErrMtFsmInvalidDeliveryTimer
		}
		val := gsm_map.SMDeliveryTimerValue(v)
		arg.SmDeliveryTimer = &val
	}
	if len(m.SmDeliveryStartTime) > 0 {
		v := gsm_map.Time(m.SmDeliveryStartTime)
		arg.SmDeliveryStartTime = &v
	}
	arg.SmsOverIPOnlyIndicator = boolToNullPtr(m.SmsOverIPOnlyIndicator)
	if m.CorrelationID != nil {
		cid, err := convertCorrelationIDToWire(m.CorrelationID)
		if err != nil {
			return nil, fmt.Errorf("encoding CorrelationID: %w", err)
		}
		arg.CorrelationID = cid
	}
	if len(m.MaximumRetransmissionTime) > 0 {
		v := gsm_map.Time(m.MaximumRetransmissionTime)
		arg.MaximumRetransmissionTime = &v
	}
	if m.SmsGmscAddress != "" {
		encoded, err := encodeAddressField(m.SmsGmscAddress, m.SmsGmscAddressNature, m.SmsGmscAddressPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SmsGmscAddress: %w", err)
		}
		v := gsm_map.ISDNAddressString(encoded)
		arg.SmsGmscAddress = &v
	}
	if m.SmsGmscDiameterAddress != nil {
		arg.SmsGmscDiameterAddress = convertNetworkNodeDiameterAddressToWire(m.SmsGmscDiameterAddress)
	}

	return arg, nil
}

func convertArgToMtFsm(arg *gsm_map.MTForwardSMArg) (*MtFsm, error) {
	var mtFsm MtFsm

	// Extract SM-RP-DA.
	switch arg.SmRPDA.Choice {
	case gsm_map.SMRPDAChoiceImsi:
		if arg.SmRPDA.Imsi == nil {
			return nil, fmt.Errorf("SMRPDA IMSI is nil")
		}
		imsi, err := tbcd.Decode(*arg.SmRPDA.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding IMSI: %w", err)
		}
		mtFsm.IMSI = imsi
	case gsm_map.SMRPDAChoiceLmsi, gsm_map.SMRPDAChoiceServiceCentreAddressDA, gsm_map.SMRPDAChoiceNoSMRPDA:
		da, err := convertWireToSmRpDa(&arg.SmRPDA)
		if err != nil {
			return nil, err
		}
		mtFsm.SmRpDa = da
	default:
		return nil, fmt.Errorf("unexpected SMRPDA choice: %d", arg.SmRPDA.Choice)
	}

	// Extract SM-RP-OA.
	switch arg.SmRPOA.Choice {
	case gsm_map.SMRPOAChoiceServiceCentreAddressOA:
		if arg.SmRPOA.ServiceCentreAddressOA == nil {
			return nil, fmt.Errorf("SMRPOA ServiceCentreAddressOA is nil")
		}
		sca, nature, plan, err := decodeAddressField(*arg.SmRPOA.ServiceCentreAddressOA)
		if err != nil {
			return nil, fmt.Errorf("decoding ServiceCentreAddressOA: %w", err)
		}
		mtFsm.ServiceCentreAddressOA = sca
		mtFsm.SCAOANature = nature
		mtFsm.SCAOAPlan = plan
	case gsm_map.SMRPOAChoiceMsisdn, gsm_map.SMRPOAChoiceNoSMRPOA:
		oa, err := convertWireToSmRpOa(&arg.SmRPOA)
		if err != nil {
			return nil, err
		}
		mtFsm.SmRpOa = oa
	default:
		return nil, fmt.Errorf("unexpected SMRPOA choice: %d", arg.SmRPOA.Choice)
	}

	// Unmarshal TPDU
	tpduResult, tpduErr := sms.Unmarshal(arg.SmRPUI, sms.AsMT)
	if tpduErr != nil {
		return nil, fmt.Errorf("unmarshaling TPDU: %w", tpduErr)
	}
	if tpduResult == nil {
		return nil, fmt.Errorf("unmarshaling TPDU: nil result")
	}
	if err := validateMtForwardSMArgTPDU(*tpduResult); err != nil {
		return nil, err
	}
	mtFsm.TPDU = *tpduResult

	// MoreMessagesToSend
	mtFsm.MoreMessagesToSend = nullPtrToBool(arg.MoreMessagesToSend)

	// Optional fields (post-extension marker).
	if arg.SmDeliveryTimer != nil {
		v := int(*arg.SmDeliveryTimer)
		if v < MinSmDeliveryTimer || v > MaxSmDeliveryTimer {
			return nil, ErrMtFsmInvalidDeliveryTimer
		}
		mtFsm.SmDeliveryTimer = &v
	}
	if arg.SmDeliveryStartTime != nil {
		mtFsm.SmDeliveryStartTime = HexBytes(*arg.SmDeliveryStartTime)
	}
	mtFsm.SmsOverIPOnlyIndicator = nullPtrToBool(arg.SmsOverIPOnlyIndicator)
	if arg.CorrelationID != nil {
		cid, err := convertWireToCorrelationID(arg.CorrelationID)
		if err != nil {
			return nil, fmt.Errorf("decoding CorrelationID: %w", err)
		}
		mtFsm.CorrelationID = cid
	}
	if arg.MaximumRetransmissionTime != nil {
		mtFsm.MaximumRetransmissionTime = HexBytes(*arg.MaximumRetransmissionTime)
	}
	if arg.SmsGmscAddress != nil {
		addr, nature, plan, err := decodeAddressField([]byte(*arg.SmsGmscAddress))
		if err != nil {
			return nil, fmt.Errorf("decoding SmsGmscAddress: %w", err)
		}
		mtFsm.SmsGmscAddress = addr
		mtFsm.SmsGmscAddressNature = nature
		mtFsm.SmsGmscAddressPlan = plan
	}
	if arg.SmsGmscDiameterAddress != nil {
		mtFsm.SmsGmscDiameterAddress = convertWireToNetworkNodeDiameterAddress(arg.SmsGmscDiameterAddress)
	}

	return &mtFsm, nil
}

// --- MT-ForwardSM Response ---

func convertMtFsmRespToRes(r *MtFsmResp) *gsm_map.MTForwardSMRes {
	out := &gsm_map.MTForwardSMRes{}
	if len(r.SmRpUI) > 0 {
		v := gsm_map.SignalInfo(r.SmRpUI)
		out.SmRPUI = &v
	}
	return out
}

func convertResToMtFsmResp(res *gsm_map.MTForwardSMRes) *MtFsmResp {
	out := &MtFsmResp{}
	if res.SmRPUI != nil {
		out.SmRpUI = HexBytes(*res.SmRPUI)
	}
	return out
}
