package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
	"github.com/warthog618/sms"
)

// --- MT-ForwardSM ---

func convertMtFsmToArg(m *MtFsm) (*gsm_map.MTForwardSMArg, error) {
	imsiBytes, err := tbcd.Encode(m.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	scaOA, err := encodeAddressField(m.ServiceCentreAddressOA, m.SCAOANature, m.SCAOAPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding ServiceCentreAddressOA: %w", err)
	}

	tpduBytes, err := m.TPDU.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshaling TPDU: %w", err)
	}

	arg := &gsm_map.MTForwardSMArg{
		SmRPDA: gsm_map.NewSMRPDAImsi(gsm_map.IMSI(imsiBytes)),
		SmRPOA: gsm_map.NewSMRPOAServiceCentreAddressOA(gsm_map.AddressString(scaOA)),
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

	// Extract IMSI from SM-RP-DA
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
	default:
		return nil, fmt.Errorf("unexpected SMRPDA choice: %d", arg.SmRPDA.Choice)
	}

	// Extract ServiceCentreAddressOA from SM-RP-OA
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
