package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// --- SRI-SM ---

func convertSriSmToArg(s *SriSm) (*gsm_map.RoutingInfoForSMArg, error) {
	msisdn, err := encodeAddressField(s.MSISDN, s.MSISDNNature, s.MSISDNPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding MSISDN: %w", err)
	}

	sca, err := encodeAddressField(s.ServiceCentreAddress, s.SCANature, s.SCAPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding ServiceCentreAddress: %w", err)
	}

	arg := &gsm_map.RoutingInfoForSMArg{
		Msisdn:               gsm_map.ISDNAddressString(msisdn),
		SmRPPRI:               s.SmRpPri,
		ServiceCentreAddress: gsm_map.AddressString(sca),
	}

	// Optional fields (post-extension marker).
	arg.GprsSupportIndicator = boolToNullPtr(s.GprsSupportIndicator)

	if s.SmRpMti != nil {
		v := gsm_map.SMRPMTI(*s.SmRpMti)
		arg.SmRPMTI = &v
	}

	if len(s.SmRpSmea) > 0 {
		v := gsm_map.SMRPSMEA(s.SmRpSmea)
		arg.SmRPSMEA = &v
	}

	if s.SmDeliveryNotIntended != nil {
		v := gsm_map.SMDeliveryNotIntended(*s.SmDeliveryNotIntended)
		arg.SmDeliveryNotIntended = &v
	}

	arg.IpSmGwGuidanceIndicator = boolToNullPtr(s.IpSmGwGuidanceIndicator)

	if s.IMSI != "" {
		imsiBytes, err := tbcd.Encode(s.IMSI)
		if err != nil {
			return nil, fmt.Errorf("encoding IMSI: %w", err)
		}
		v := gsm_map.IMSI(imsiBytes)
		arg.Imsi = &v
	}

	arg.SingleAttemptDelivery = boolToNullPtr(s.SingleAttemptDelivery)
	arg.T4TriggerIndicator = boolToNullPtr(s.T4TriggerIndicator)

	if s.CorrelationID != nil {
		cid, err := convertCorrelationIDToWire(s.CorrelationID)
		if err != nil {
			return nil, fmt.Errorf("CorrelationID: %w", err)
		}
		arg.CorrelationID = cid
	}

	arg.SmsfSupportIndicator = boolToNullPtr(s.SmsfSupportIndicator)

	return arg, nil
}

func convertArgToSriSm(arg *gsm_map.RoutingInfoForSMArg) (*SriSm, error) {
	msisdn, msisdnNature, msisdnPlan, err := decodeAddressField(arg.Msisdn)
	if err != nil {
		return nil, fmt.Errorf("decoding MSISDN: %w", err)
	}

	sca, scaNature, scaPlan, err := decodeAddressField(arg.ServiceCentreAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding ServiceCentreAddress: %w", err)
	}

	s := &SriSm{
		MSISDN:               msisdn,
		MSISDNNature:         msisdnNature,
		MSISDNPlan:           msisdnPlan,
		SmRpPri:              arg.SmRPPRI,
		ServiceCentreAddress: sca,
		SCANature:            scaNature,
		SCAPlan:              scaPlan,
	}

	// Optional fields (post-extension marker).
	s.GprsSupportIndicator = nullPtrToBool(arg.GprsSupportIndicator)

	if arg.SmRPMTI != nil {
		v := int(*arg.SmRPMTI)
		s.SmRpMti = &v
	}

	if arg.SmRPSMEA != nil {
		s.SmRpSmea = HexBytes(*arg.SmRPSMEA)
	}

	if arg.SmDeliveryNotIntended != nil {
		v := SmDeliveryNotIntended(*arg.SmDeliveryNotIntended)
		s.SmDeliveryNotIntended = &v
	}

	s.IpSmGwGuidanceIndicator = nullPtrToBool(arg.IpSmGwGuidanceIndicator)

	if arg.Imsi != nil {
		imsi, err := tbcd.Decode(*arg.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding optional IMSI: %w", err)
		}
		s.IMSI = imsi
	}

	s.SingleAttemptDelivery = nullPtrToBool(arg.SingleAttemptDelivery)
	s.T4TriggerIndicator = nullPtrToBool(arg.T4TriggerIndicator)

	if arg.CorrelationID != nil {
		s.CorrelationID = convertWireToCorrelationID(arg.CorrelationID)
	}

	s.SmsfSupportIndicator = nullPtrToBool(arg.SmsfSupportIndicator)

	return s, nil
}

// --- SRI-SM Response ---

func convertSriSmRespToRes(s *SriSmResp) (*gsm_map.RoutingInfoForSMRes, error) {
	imsiBytes, err := tbcd.Encode(s.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	nnn, err := encodeAddressField(
		s.LocationInfoWithLMSI.NetworkNodeNumber,
		s.LocationInfoWithLMSI.NetworkNodeNumberNature,
		s.LocationInfoWithLMSI.NetworkNodeNumberPlan,
	)
	if err != nil {
		return nil, fmt.Errorf("encoding NetworkNodeNumber: %w", err)
	}

	li := gsm_map.LocationInfoWithLMSI{
		NetworkNodeNumber: gsm_map.ISDNAddressString(nnn),
	}

	// LMSI
	if len(s.LocationInfoWithLMSI.LMSI) > 0 {
		v := gsm_map.LMSI(s.LocationInfoWithLMSI.LMSI)
		li.Lmsi = &v
	}

	// GprsNodeIndicator
	li.GprsNodeIndicator = boolToNullPtr(s.LocationInfoWithLMSI.GprsNodeIndicator)

	// AdditionalNumber
	if s.LocationInfoWithLMSI.AdditionalNumber != nil {
		an, err := convertAdditionalNumberToWire(s.LocationInfoWithLMSI.AdditionalNumber)
		if err != nil {
			return nil, fmt.Errorf("encoding AdditionalNumber: %w", err)
		}
		li.AdditionalNumber = an
	}

	// NetworkNodeDiameterAddress
	if s.LocationInfoWithLMSI.NetworkNodeDiameterAddress != nil {
		li.NetworkNodeDiameterAddress = convertNetworkNodeDiameterAddressToWire(s.LocationInfoWithLMSI.NetworkNodeDiameterAddress)
	}

	// AdditionalNetworkNodeDiameterAddress
	if s.LocationInfoWithLMSI.AdditionalNetworkNodeDiameterAddress != nil {
		li.AdditionalNetworkNodeDiameterAddress = convertNetworkNodeDiameterAddressToWire(s.LocationInfoWithLMSI.AdditionalNetworkNodeDiameterAddress)
	}

	// ThirdNumber
	if s.LocationInfoWithLMSI.ThirdNumber != nil {
		tn, err := convertAdditionalNumberToWire(s.LocationInfoWithLMSI.ThirdNumber)
		if err != nil {
			return nil, fmt.Errorf("encoding ThirdNumber: %w", err)
		}
		li.ThirdNumber = tn
	}

	// ThirdNetworkNodeDiameterAddress
	if s.LocationInfoWithLMSI.ThirdNetworkNodeDiameterAddress != nil {
		li.ThirdNetworkNodeDiameterAddress = convertNetworkNodeDiameterAddressToWire(s.LocationInfoWithLMSI.ThirdNetworkNodeDiameterAddress)
	}

	// ImsNodeIndicator
	li.ImsNodeIndicator = boolToNullPtr(s.LocationInfoWithLMSI.ImsNodeIndicator)

	// Smsf3gppNumber
	if s.LocationInfoWithLMSI.Smsf3gppNumber != "" {
		encoded, err := encodeAddressField(s.LocationInfoWithLMSI.Smsf3gppNumber, s.LocationInfoWithLMSI.Smsf3gppNumberNature, s.LocationInfoWithLMSI.Smsf3gppNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding Smsf3gppNumber: %w", err)
		}
		v := gsm_map.ISDNAddressString(encoded)
		li.Smsf3gppNumber = &v
	}

	// Smsf3gppDiameterAddress
	if s.LocationInfoWithLMSI.Smsf3gppDiameterAddress != nil {
		li.Smsf3gppDiameterAddress = convertNetworkNodeDiameterAddressToWire(s.LocationInfoWithLMSI.Smsf3gppDiameterAddress)
	}

	// SmsfNon3gppNumber
	if s.LocationInfoWithLMSI.SmsfNon3gppNumber != "" {
		encoded, err := encodeAddressField(s.LocationInfoWithLMSI.SmsfNon3gppNumber, s.LocationInfoWithLMSI.SmsfNon3gppNumberNature, s.LocationInfoWithLMSI.SmsfNon3gppNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SmsfNon3gppNumber: %w", err)
		}
		v := gsm_map.ISDNAddressString(encoded)
		li.SmsfNon3gppNumber = &v
	}

	// SmsfNon3gppDiameterAddress
	if s.LocationInfoWithLMSI.SmsfNon3gppDiameterAddress != nil {
		li.SmsfNon3gppDiameterAddress = convertNetworkNodeDiameterAddressToWire(s.LocationInfoWithLMSI.SmsfNon3gppDiameterAddress)
	}

	// Smsf3gppAddressIndicator
	li.Smsf3gppAddressIndicator = boolToNullPtr(s.LocationInfoWithLMSI.Smsf3gppAddressIndicator)

	// SmsfNon3gppAddressIndicator
	li.SmsfNon3gppAddressIndicator = boolToNullPtr(s.LocationInfoWithLMSI.SmsfNon3gppAddressIndicator)

	out := &gsm_map.RoutingInfoForSMRes{
		Imsi:                 gsm_map.IMSI(imsiBytes),
		LocationInfoWithLMSI: li,
	}

	// IpSmGwGuidance
	if s.IpSmGwGuidance != nil {
		gw, err := convertIpSmGwGuidanceToWire(s.IpSmGwGuidance)
		if err != nil {
			return nil, fmt.Errorf("IpSmGwGuidance: %w", err)
		}
		out.IpSmGwGuidance = gw
	}

	return out, nil
}

func convertResToSriSmResp(res *gsm_map.RoutingInfoForSMRes) (*SriSmResp, error) {
	imsi, err := tbcd.Decode(res.Imsi)
	if err != nil {
		return nil, fmt.Errorf("decoding IMSI: %w", err)
	}

	nnn, nnnNature, nnnPlan, err := decodeAddressField(res.LocationInfoWithLMSI.NetworkNodeNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding NetworkNodeNumber: %w", err)
	}

	li := LocationInfoWithLMSI{
		NetworkNodeNumber:       nnn,
		NetworkNodeNumberNature: nnnNature,
		NetworkNodeNumberPlan:   nnnPlan,
	}

	// LMSI
	if res.LocationInfoWithLMSI.Lmsi != nil {
		if len(*res.LocationInfoWithLMSI.Lmsi) != 4 {
			return nil, fmt.Errorf("LocationInfoWithLMSI.LMSI must be exactly 4 octets, got %d", len(*res.LocationInfoWithLMSI.Lmsi))
		}
		li.LMSI = HexBytes(*res.LocationInfoWithLMSI.Lmsi)
	}

	// GprsNodeIndicator
	li.GprsNodeIndicator = nullPtrToBool(res.LocationInfoWithLMSI.GprsNodeIndicator)

	// AdditionalNumber
	if res.LocationInfoWithLMSI.AdditionalNumber != nil {
		an, err := convertWireToAdditionalNumber(res.LocationInfoWithLMSI.AdditionalNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding AdditionalNumber: %w", err)
		}
		li.AdditionalNumber = an
	}

	// NetworkNodeDiameterAddress
	if res.LocationInfoWithLMSI.NetworkNodeDiameterAddress != nil {
		li.NetworkNodeDiameterAddress = convertWireToNetworkNodeDiameterAddress(res.LocationInfoWithLMSI.NetworkNodeDiameterAddress)
	}

	// AdditionalNetworkNodeDiameterAddress
	if res.LocationInfoWithLMSI.AdditionalNetworkNodeDiameterAddress != nil {
		li.AdditionalNetworkNodeDiameterAddress = convertWireToNetworkNodeDiameterAddress(res.LocationInfoWithLMSI.AdditionalNetworkNodeDiameterAddress)
	}

	// ThirdNumber
	if res.LocationInfoWithLMSI.ThirdNumber != nil {
		tn, err := convertWireToAdditionalNumber(res.LocationInfoWithLMSI.ThirdNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding ThirdNumber: %w", err)
		}
		li.ThirdNumber = tn
	}

	// ThirdNetworkNodeDiameterAddress
	if res.LocationInfoWithLMSI.ThirdNetworkNodeDiameterAddress != nil {
		li.ThirdNetworkNodeDiameterAddress = convertWireToNetworkNodeDiameterAddress(res.LocationInfoWithLMSI.ThirdNetworkNodeDiameterAddress)
	}

	// ImsNodeIndicator
	li.ImsNodeIndicator = nullPtrToBool(res.LocationInfoWithLMSI.ImsNodeIndicator)

	// Smsf3gppNumber
	if res.LocationInfoWithLMSI.Smsf3gppNumber != nil {
		num, nature, plan, err := decodeAddressField(*res.LocationInfoWithLMSI.Smsf3gppNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding Smsf3gppNumber: %w", err)
		}
		li.Smsf3gppNumber = num
		li.Smsf3gppNumberNature = nature
		li.Smsf3gppNumberPlan = plan
	}

	// Smsf3gppDiameterAddress
	if res.LocationInfoWithLMSI.Smsf3gppDiameterAddress != nil {
		li.Smsf3gppDiameterAddress = convertWireToNetworkNodeDiameterAddress(res.LocationInfoWithLMSI.Smsf3gppDiameterAddress)
	}

	// SmsfNon3gppNumber
	if res.LocationInfoWithLMSI.SmsfNon3gppNumber != nil {
		num, nature, plan, err := decodeAddressField(*res.LocationInfoWithLMSI.SmsfNon3gppNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding SmsfNon3gppNumber: %w", err)
		}
		li.SmsfNon3gppNumber = num
		li.SmsfNon3gppNumberNature = nature
		li.SmsfNon3gppNumberPlan = plan
	}

	// SmsfNon3gppDiameterAddress
	if res.LocationInfoWithLMSI.SmsfNon3gppDiameterAddress != nil {
		li.SmsfNon3gppDiameterAddress = convertWireToNetworkNodeDiameterAddress(res.LocationInfoWithLMSI.SmsfNon3gppDiameterAddress)
	}

	// Smsf3gppAddressIndicator
	li.Smsf3gppAddressIndicator = nullPtrToBool(res.LocationInfoWithLMSI.Smsf3gppAddressIndicator)

	// SmsfNon3gppAddressIndicator
	li.SmsfNon3gppAddressIndicator = nullPtrToBool(res.LocationInfoWithLMSI.SmsfNon3gppAddressIndicator)

	resp := &SriSmResp{
		IMSI:                 imsi,
		LocationInfoWithLMSI: li,
	}

	// IpSmGwGuidance
	if res.IpSmGwGuidance != nil {
		resp.IpSmGwGuidance = convertWireToIpSmGwGuidance(res.IpSmGwGuidance)
	}

	return resp, nil
}
