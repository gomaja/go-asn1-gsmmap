package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/runtime"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

	"github.com/gomaja/go-asn1-gsmmap/address"
	"github.com/gomaja/go-asn1-gsmmap/gsn"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"

	"github.com/warthog618/sms"
)

const errEncodingIMSI = "encoding IMSI: %w"

// natureOrDefault returns the given nature if non-zero, otherwise International.
func natureOrDefault(nature uint8) uint8 {
	if nature == 0 {
		return address.NatureInternational
	}
	return nature
}

// planOrDefault returns the given plan if non-zero, otherwise ISDN.
func planOrDefault(plan uint8) uint8 {
	if plan == 0 {
		return address.PlanISDN
	}
	return plan
}

// encodeAddressField encodes a phone number string into an AddressString byte slice.
func encodeAddressField(digits string, nature, plan uint8) ([]byte, error) {
	tbcdBytes, err := tbcd.Encode(digits)
	if err != nil {
		return nil, err
	}
	return address.Encode(address.ExtensionNo, natureOrDefault(nature), planOrDefault(plan), tbcdBytes), nil
}

// decodeAddressField decodes an AddressString byte slice into a phone number string and address components.
func decodeAddressField(encoded []byte) (digits string, nature, plan uint8, err error) {
	_, nat, pl, rawDigits := address.Decode(encoded)
	digits, err = tbcd.Decode(rawDigits)
	if err != nil {
		return "", 0, 0, err
	}
	return digits, nat, pl, nil
}

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

// --- SRI-SM helper converters ---

func convertAdditionalNumberToWire(a *AdditionalNumber) (*gsm_map.AdditionalNumber, error) {
	hasMsc := a.MscNumber != ""
	hasSgsn := a.SgsnNumber != ""
	switch {
	case hasMsc && hasSgsn:
		return nil, ErrSriChoiceMultipleAlternatives
	case hasMsc:
		encoded, err := encodeAddressField(a.MscNumber, a.MscNumberNature, a.MscNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding MscNumber: %w", err)
		}
		v := gsm_map.NewAdditionalNumberMscNumber(gsm_map.ISDNAddressString(encoded))
		return &v, nil
	case hasSgsn:
		encoded, err := encodeAddressField(a.SgsnNumber, a.SgsnNumberNature, a.SgsnNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SgsnNumber: %w", err)
		}
		v := gsm_map.NewAdditionalNumberSgsnNumber(gsm_map.ISDNAddressString(encoded))
		return &v, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertWireToAdditionalNumber(w *gsm_map.AdditionalNumber) (*AdditionalNumber, error) {
	an := &AdditionalNumber{}
	switch w.Choice {
	case gsm_map.AdditionalNumberChoiceMscNumber:
		if w.MscNumber == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		num, nature, plan, err := decodeAddressField(*w.MscNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding MscNumber: %w", err)
		}
		an.MscNumber = num
		an.MscNumberNature = nature
		an.MscNumberPlan = plan
	case gsm_map.AdditionalNumberChoiceSgsnNumber:
		if w.SgsnNumber == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		num, nature, plan, err := decodeAddressField(*w.SgsnNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding SgsnNumber: %w", err)
		}
		an.SgsnNumber = num
		an.SgsnNumberNature = nature
		an.SgsnNumberPlan = plan
	default:
		return nil, ErrSriChoiceNoAlternative
	}
	return an, nil
}

func convertNetworkNodeDiameterAddressToWire(n *NetworkNodeDiameterAddress) *gsm_map.NetworkNodeDiameterAddress {
	return &gsm_map.NetworkNodeDiameterAddress{
		DiameterName:  gsm_map.DiameterIdentity(n.DiameterName),
		DiameterRealm: gsm_map.DiameterIdentity(n.DiameterRealm),
	}
}

func convertWireToNetworkNodeDiameterAddress(w *gsm_map.NetworkNodeDiameterAddress) *NetworkNodeDiameterAddress {
	return &NetworkNodeDiameterAddress{
		DiameterName:  HexBytes(w.DiameterName),
		DiameterRealm: HexBytes(w.DiameterRealm),
	}
}

func convertCorrelationIDToWire(c *SriSmCorrelationID) (*gsm_map.CorrelationID, error) {
	if len(c.SipUriB) == 0 {
		return nil, ErrSriSmMissingSipUriB
	}
	out := &gsm_map.CorrelationID{
		SipUriB: gsm_map.SIPURI(c.SipUriB),
	}
	if len(c.HlrID) > 0 {
		v := gsm_map.HLRId(c.HlrID)
		out.HlrId = &v
	}
	if len(c.SipUriA) > 0 {
		v := gsm_map.SIPURI(c.SipUriA)
		out.SipUriA = &v
	}
	return out, nil
}

func convertWireToCorrelationID(w *gsm_map.CorrelationID) *SriSmCorrelationID {
	c := &SriSmCorrelationID{
		SipUriB: HexBytes(w.SipUriB),
	}
	if w.HlrId != nil {
		c.HlrID = HexBytes(*w.HlrId)
	}
	if w.SipUriA != nil {
		c.SipUriA = HexBytes(*w.SipUriA)
	}
	return c
}

func convertIpSmGwGuidanceToWire(g *IpSmGwGuidance) (*gsm_map.IPSMGWGuidance, error) {
	if g.MinimumDeliveryTimeValue < MinSmDeliveryTimer || g.MinimumDeliveryTimeValue > MaxSmDeliveryTimer ||
		g.RecommendedDeliveryTimeValue < MinSmDeliveryTimer || g.RecommendedDeliveryTimeValue > MaxSmDeliveryTimer {
		return nil, ErrSriSmInvalidDeliveryTimerValue
	}
	return &gsm_map.IPSMGWGuidance{
		MinimumDeliveryTimeValue:     gsm_map.SMDeliveryTimerValue(g.MinimumDeliveryTimeValue),
		RecommendedDeliveryTimeValue: gsm_map.SMDeliveryTimerValue(g.RecommendedDeliveryTimeValue),
	}, nil
}

func convertWireToIpSmGwGuidance(w *gsm_map.IPSMGWGuidance) *IpSmGwGuidance {
	return &IpSmGwGuidance{
		MinimumDeliveryTimeValue:     int(w.MinimumDeliveryTimeValue),
		RecommendedDeliveryTimeValue: int(w.RecommendedDeliveryTimeValue),
	}
}

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
		mtFsm.CorrelationID = convertWireToCorrelationID(arg.CorrelationID)
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
		moFsm.CorrelationID = convertWireToCorrelationID(arg.CorrelationID)
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

// --- UpdateLocation ---

func convertUpdateLocationToArg(u *UpdateLocation) (*gsm_map.UpdateLocationArg, error) {
	imsiBytes, err := tbcd.Encode(u.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	mscNumber, err := encodeAddressField(u.MSCNumber, u.MSCNature, u.MSCPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding MSCNumber: %w", err)
	}

	vlrNumber, err := encodeAddressField(u.VLRNumber, u.VLRNature, u.VLRPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding VLRNumber: %w", err)
	}

	arg := &gsm_map.UpdateLocationArg{
		Imsi:      gsm_map.IMSI(imsiBytes),
		MscNumber: gsm_map.ISDNAddressString(mscNumber),
		VlrNumber: gsm_map.ISDNAddressString(vlrNumber),
	}

	if u.VlrCapability != nil {
		vlrCap := &gsm_map.VLRCapability{}

		if u.VlrCapability.SupportedCamelPhases != nil {
			bs := convertCamelPhasesToBitString(u.VlrCapability.SupportedCamelPhases)
			vlrCap.SupportedCamelPhases = &bs
		}

		if u.VlrCapability.SupportedLCSCapabilitySets != nil {
			bs := convertLCSCapsToBitString(u.VlrCapability.SupportedLCSCapabilitySets)
			vlrCap.SupportedLCSCapabilitySets = &bs
		}

		vlrCap.SolsaSupportIndicator = boolToNullPtr(u.VlrCapability.SolsaSupportIndicator)

		if u.VlrCapability.IstSupportIndicator != nil {
			v := gsm_map.ISTSupportIndicator(int64(*u.VlrCapability.IstSupportIndicator))
			vlrCap.IstSupportIndicator = &v
		}

		if u.VlrCapability.SuperChargerSupportedInServingNetworkEntity != nil {
			sc, err := convertSuperChargerInfoToWire(u.VlrCapability.SuperChargerSupportedInServingNetworkEntity)
			if err != nil {
				return nil, fmt.Errorf("SuperChargerInfo: %w", err)
			}
			vlrCap.SuperChargerSupportedInServingNetworkEntity = sc
		}

		vlrCap.LongFTNSupported = boolToNullPtr(u.VlrCapability.LongFTNSupported)

		if u.VlrCapability.OfferedCamel4CSIs != nil {
			bs := convertOfferedCamel4CSIsToBitString(u.VlrCapability.OfferedCamel4CSIs)
			vlrCap.OfferedCamel4CSIs = &bs
		}

		if u.VlrCapability.SupportedRATTypesIndicator != nil {
			bs := convertSupportedRATTypesToBitString(u.VlrCapability.SupportedRATTypesIndicator)
			vlrCap.SupportedRATTypesIndicator = &bs
		}

		vlrCap.LongGroupIDSupported = boolToNullPtr(u.VlrCapability.LongGroupIDSupported)
		vlrCap.MtRoamingForwardingSupported = boolToNullPtr(u.VlrCapability.MtRoamingForwardingSupported)
		vlrCap.MsisdnLessOperationSupported = boolToNullPtr(u.VlrCapability.MsisdnLessOperationSupported)
		vlrCap.ResetIdsSupported = boolToNullPtr(u.VlrCapability.ResetIdsSupported)

		arg.VlrCapability = vlrCap
	}

	// Optional fields.
	if len(u.LMSI) > 0 {
		if len(u.LMSI) != 4 {
			return nil, fmt.Errorf("UpdateLocation: LMSI must be exactly 4 octets, got %d", len(u.LMSI))
		}
		v := gsm_map.LMSI(u.LMSI)
		arg.Lmsi = &v
	}

	arg.InformPreviousNetworkEntity = boolToNullPtr(u.InformPreviousNetworkEntity)
	arg.CsLCSNotSupportedByUE = boolToNullPtr(u.CsLCSNotSupportedByUE)

	if u.VGmlcAddress != "" {
		gsnAddr, err := gsn.Build(u.VGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("encoding VGmlcAddress: %w", err)
		}
		v := gsm_map.GSNAddress(gsnAddr)
		arg.VGmlcAddress = &v
	}

	if u.AddInfo != nil {
		ai, err := convertAddInfoToWire(u.AddInfo)
		if err != nil {
			return nil, fmt.Errorf("AddInfo: %w", err)
		}
		arg.AddInfo = ai
	}

	if len(u.PagingArea) > 0 {
		pa := make(gsm_map.PagingArea, len(u.PagingArea))
		for i, raw := range u.PagingArea {
			// Each raw HexBytes is BER-encoded LocationArea CHOICE.
			var la gsm_map.LocationArea
			if err := la.UnmarshalBER(raw); err != nil {
				return nil, fmt.Errorf("PagingArea[%d]: %w", i, err)
			}
			pa[i] = la
		}
		arg.PagingArea = pa
	}

	arg.SkipSubscriberDataUpdate = boolToNullPtr(u.SkipSubscriberDataUpdate)
	arg.RestorationIndicator = boolToNullPtr(u.RestorationIndicator)

	if len(u.EplmnList) > 0 {
		list := make(gsm_map.EPLMNList, len(u.EplmnList))
		for i, raw := range u.EplmnList {
			if len(raw) != 3 {
				return nil, fmt.Errorf("UpdateLocation: EplmnList[%d] PLMNId must be exactly 3 octets, got %d", i, len(raw))
			}
			list[i] = gsm_map.PLMNId(raw)
		}
		arg.EplmnList = list
	}

	if u.MmeDiameterAddress != nil {
		arg.MmeDiameterAddress = convertNetworkNodeDiameterAddressToWire(u.MmeDiameterAddress)
	}

	return arg, nil
}

func convertArgToUpdateLocation(arg *gsm_map.UpdateLocationArg) (*UpdateLocation, error) {
	imsi, err := tbcd.Decode(arg.Imsi)
	if err != nil {
		return nil, fmt.Errorf("decoding IMSI: %w", err)
	}

	msc, mscNature, mscPlan, err := decodeAddressField(arg.MscNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding MSCNumber: %w", err)
	}

	vlr, vlrNature, vlrPlan, err := decodeAddressField(arg.VlrNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding VLRNumber: %w", err)
	}

	u := &UpdateLocation{
		IMSI:      imsi,
		MSCNumber: msc,
		MSCNature: mscNature,
		MSCPlan:   mscPlan,
		VLRNumber: vlr,
		VLRNature: vlrNature,
		VLRPlan:   vlrPlan,
	}

	if arg.VlrCapability != nil {
		vlrCap := &VlrCapability{}

		if arg.VlrCapability.SupportedCamelPhases != nil && arg.VlrCapability.SupportedCamelPhases.BitLength > 0 {
			vlrCap.SupportedCamelPhases = convertBitStringToCamelPhases(*arg.VlrCapability.SupportedCamelPhases)
		}

		if arg.VlrCapability.SupportedLCSCapabilitySets != nil && arg.VlrCapability.SupportedLCSCapabilitySets.BitLength > 0 {
			vlrCap.SupportedLCSCapabilitySets = convertBitStringToLCSCaps(*arg.VlrCapability.SupportedLCSCapabilitySets)
		}

		vlrCap.SolsaSupportIndicator = nullPtrToBool(arg.VlrCapability.SolsaSupportIndicator)

		if arg.VlrCapability.IstSupportIndicator != nil {
			v := int(int64(*arg.VlrCapability.IstSupportIndicator))
			vlrCap.IstSupportIndicator = &v
		}

		if arg.VlrCapability.SuperChargerSupportedInServingNetworkEntity != nil {
			sc, err := convertWireToSuperChargerInfo(arg.VlrCapability.SuperChargerSupportedInServingNetworkEntity)
			if err != nil {
				return nil, fmt.Errorf("SuperChargerInfo: %w", err)
			}
			vlrCap.SuperChargerSupportedInServingNetworkEntity = sc
		}

		vlrCap.LongFTNSupported = nullPtrToBool(arg.VlrCapability.LongFTNSupported)

		if arg.VlrCapability.OfferedCamel4CSIs != nil && arg.VlrCapability.OfferedCamel4CSIs.BitLength > 0 {
			vlrCap.OfferedCamel4CSIs = convertBitStringToOfferedCamel4CSIs(*arg.VlrCapability.OfferedCamel4CSIs)
		}

		if arg.VlrCapability.SupportedRATTypesIndicator != nil && arg.VlrCapability.SupportedRATTypesIndicator.BitLength > 0 {
			if arg.VlrCapability.SupportedRATTypesIndicator.BitLength < 2 || arg.VlrCapability.SupportedRATTypesIndicator.BitLength > 8 {
				return nil, fmt.Errorf("UpdateLocation: SupportedRATTypes BitLength must be 2..8, got %d", arg.VlrCapability.SupportedRATTypesIndicator.BitLength)
			}
			vlrCap.SupportedRATTypesIndicator = convertBitStringToSupportedRATTypes(*arg.VlrCapability.SupportedRATTypesIndicator)
		}

		vlrCap.LongGroupIDSupported = nullPtrToBool(arg.VlrCapability.LongGroupIDSupported)
		vlrCap.MtRoamingForwardingSupported = nullPtrToBool(arg.VlrCapability.MtRoamingForwardingSupported)
		vlrCap.MsisdnLessOperationSupported = nullPtrToBool(arg.VlrCapability.MsisdnLessOperationSupported)
		vlrCap.ResetIdsSupported = nullPtrToBool(arg.VlrCapability.ResetIdsSupported)

		u.VlrCapability = vlrCap
	}

	// Optional fields.
	if arg.Lmsi != nil {
		if len(*arg.Lmsi) != 4 {
			return nil, fmt.Errorf("UpdateLocation: LMSI must be exactly 4 octets, got %d", len(*arg.Lmsi))
		}
		u.LMSI = HexBytes(*arg.Lmsi)
	}

	u.InformPreviousNetworkEntity = nullPtrToBool(arg.InformPreviousNetworkEntity)
	u.CsLCSNotSupportedByUE = nullPtrToBool(arg.CsLCSNotSupportedByUE)

	if arg.VGmlcAddress != nil {
		addr, err := gsn.Parse(*arg.VGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("decoding VGmlcAddress: %w", err)
		}
		u.VGmlcAddress = addr
	}

	if arg.AddInfo != nil {
		ai, err := convertWireToAddInfo(arg.AddInfo)
		if err != nil {
			return nil, fmt.Errorf("AddInfo: %w", err)
		}
		u.AddInfo = ai
	}

	if len(arg.PagingArea) > 0 {
		pa := make([]HexBytes, len(arg.PagingArea))
		for i, la := range arg.PagingArea {
			encoded, err := la.MarshalDER()
			if err != nil {
				return nil, fmt.Errorf("PagingArea[%d]: %w", i, err)
			}
			pa[i] = HexBytes(encoded)
		}
		u.PagingArea = pa
	}

	u.SkipSubscriberDataUpdate = nullPtrToBool(arg.SkipSubscriberDataUpdate)
	u.RestorationIndicator = nullPtrToBool(arg.RestorationIndicator)

	if len(arg.EplmnList) > 0 {
		list := make([]HexBytes, len(arg.EplmnList))
		for i, plmn := range arg.EplmnList {
			if len(plmn) != 3 {
				return nil, fmt.Errorf("UpdateLocation: EplmnList[%d] PLMNId must be exactly 3 octets, got %d", i, len(plmn))
			}
			list[i] = HexBytes(plmn)
		}
		u.EplmnList = list
	}

	if arg.MmeDiameterAddress != nil {
		u.MmeDiameterAddress = convertWireToNetworkNodeDiameterAddress(arg.MmeDiameterAddress)
	}

	return u, nil
}

// --- UpdateLocationRes ---

func convertUpdateLocationResToRes(u *UpdateLocationRes) (*gsm_map.UpdateLocationRes, error) {
	hlr, err := encodeAddressField(u.HLRNumber, u.HLRNumberNature, u.HLRNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding HLRNumber: %w", err)
	}

	res := &gsm_map.UpdateLocationRes{
		HlrNumber:            gsm_map.ISDNAddressString(hlr),
		AddCapability:        boolToNullPtr(u.AddCapability),
		PagingAreaCapability: boolToNullPtr(u.PagingAreaCapability),
	}
	return res, nil
}

func convertResToUpdateLocationRes(res *gsm_map.UpdateLocationRes) (*UpdateLocationRes, error) {
	hlr, nature, plan, err := decodeAddressField(res.HlrNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding HLRNumber: %w", err)
	}

	return &UpdateLocationRes{
		HLRNumber:            hlr,
		HLRNumberNature:      nature,
		HLRNumberPlan:        plan,
		AddCapability:        nullPtrToBool(res.AddCapability),
		PagingAreaCapability: nullPtrToBool(res.PagingAreaCapability),
	}, nil
}

// --- UpdateGprsLocation ---

func convertUpdateGprsLocationToArg(u *UpdateGprsLocation) (*gsm_map.UpdateGprsLocationArg, error) {
	imsiBytes, err := tbcd.Encode(u.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	sgsnNumber, err := encodeAddressField(u.SGSNNumber, u.SGSNNature, u.SGSNPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding SGSNNumber: %w", err)
	}

	sgsnAddr, err := gsn.Build(u.SGSNAddress)
	if err != nil {
		return nil, fmt.Errorf("encoding SGSNAddress: %w", err)
	}

	arg := &gsm_map.UpdateGprsLocationArg{
		Imsi:        gsm_map.IMSI(imsiBytes),
		SgsnNumber:  gsm_map.ISDNAddressString(sgsnNumber),
		SgsnAddress: gsm_map.GSNAddress(sgsnAddr),
	}

	if u.SGSNCapability != nil {
		sgsnCap, err := convertSGSNCapabilityToWire(u.SGSNCapability)
		if err != nil {
			return nil, fmt.Errorf("SGSNCapability: %w", err)
		}
		arg.SgsnCapability = sgsnCap
	}

	// [1] informPreviousNetworkEntity / [2] psLCSNotSupportedByUE
	arg.InformPreviousNetworkEntity = boolToNullPtr(u.InformPreviousNetworkEntity)
	arg.PsLCSNotSupportedByUE = boolToNullPtr(u.PsLCSNotSupportedByUE)

	// [3] v-gmlc-Address
	if u.VGmlcAddress != "" {
		gsnAddr, err := gsn.Build(u.VGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("encoding VGmlcAddress: %w", err)
		}
		v := gsm_map.GSNAddress(gsnAddr)
		arg.VGmlcAddress = &v
	}

	// [4] add-info
	if u.AddInfo != nil {
		ai, err := convertAddInfoToWire(u.AddInfo)
		if err != nil {
			return nil, fmt.Errorf("AddInfo: %w", err)
		}
		arg.AddInfo = ai
	}

	// [5] eps-info
	if u.EpsInfo != nil {
		ei, err := convertEpsInfoToWire(u.EpsInfo)
		if err != nil {
			return nil, fmt.Errorf("EpsInfo: %w", err)
		}
		arg.EpsInfo = ei
	}

	// [6]..[13] simple NULL flags
	arg.ServingNodeTypeIndicator = boolToNullPtr(u.ServingNodeTypeIndicator)
	arg.SkipSubscriberDataUpdate = boolToNullPtr(u.SkipSubscriberDataUpdate)

	// [8] usedRatType
	if u.UsedRatType != nil {
		v := gsm_map.UsedRATType(int64(*u.UsedRatType))
		arg.UsedRATType = &v
	}

	arg.GprsSubscriptionDataNotNeeded = boolToNullPtr(u.GprsSubscriptionDataNotNeeded)
	arg.NodeTypeIndicator = boolToNullPtr(u.NodeTypeIndicator)
	arg.AreaRestricted = boolToNullPtr(u.AreaRestricted)
	arg.UeReachableIndicator = boolToNullPtr(u.UeReachableIndicator)
	arg.EpsSubscriptionDataNotNeeded = boolToNullPtr(u.EpsSubscriptionDataNotNeeded)

	// [14] ue-SRVCC-Capability
	if u.UeSrvccCapability != nil {
		v := gsm_map.UESRVCCCapability(int64(*u.UeSrvccCapability))
		arg.UeSrvccCapability = &v
	}

	// [15] eplmn-List
	if len(u.EplmnList) > 0 {
		list := make(gsm_map.EPLMNList, len(u.EplmnList))
		for i, raw := range u.EplmnList {
			if len(raw) != 3 {
				return nil, fmt.Errorf("UpdateGprsLocation: EplmnList[%d] PLMNId must be exactly 3 octets, got %d", i, len(raw))
			}
			list[i] = gsm_map.PLMNId(raw)
		}
		arg.EplmnList = list
	}

	// [16] mme-Number-for-MT-SMS
	if u.MmeNumberForMTSMS != "" {
		mme, err := encodeAddressField(u.MmeNumberForMTSMS, u.MmeNumberForMTSMSNature, u.MmeNumberForMTSMSPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding MmeNumberForMTSMS: %w", err)
		}
		v := gsm_map.ISDNAddressString(mme)
		arg.MmeNumberforMTSMS = &v
	}

	// [17] smsRegisterRequest
	if u.SmsRegisterRequest != nil {
		v := gsm_map.SMSRegisterRequest(int64(*u.SmsRegisterRequest))
		arg.SmsRegisterRequest = &v
	}

	arg.SmsOnly = boolToNullPtr(u.SmsOnly)

	// [19]/[20] DiameterIdentity
	if len(u.SgsnName) > 0 {
		v := gsm_map.DiameterIdentity(u.SgsnName)
		arg.SgsnName = &v
	}
	if len(u.SgsnRealm) > 0 {
		v := gsm_map.DiameterIdentity(u.SgsnRealm)
		arg.SgsnRealm = &v
	}

	arg.LgdSupportIndicator = boolToNullPtr(u.LgdSupportIndicator)
	arg.RemovalofMMERegistrationforSMS = boolToNullPtr(u.RemovalofMMERegistrationforSMS)

	// [23] adjacentPLMNList
	if len(u.AdjacentPLMNList) > 0 {
		list := make(gsm_map.AdjacentPLMNList, len(u.AdjacentPLMNList))
		for i, raw := range u.AdjacentPLMNList {
			if len(raw) != 3 {
				return nil, fmt.Errorf("UpdateGprsLocation: AdjacentPLMNList[%d] PLMNId must be exactly 3 octets, got %d", i, len(raw))
			}
			list[i] = gsm_map.PLMNId(raw)
		}
		arg.AdjacentPLMNList = list
	}

	return arg, nil
}

func convertArgToUpdateGprsLocation(arg *gsm_map.UpdateGprsLocationArg) (*UpdateGprsLocation, error) {
	imsi, err := tbcd.Decode(arg.Imsi)
	if err != nil {
		return nil, fmt.Errorf("decoding IMSI: %w", err)
	}

	sgsnNum, sgsnNature, sgsnPlan, err := decodeAddressField(arg.SgsnNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding SGSNNumber: %w", err)
	}

	sgsnAddr, err := gsn.Parse(arg.SgsnAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding SGSNAddress: %w", err)
	}

	u := &UpdateGprsLocation{
		IMSI:        imsi,
		SGSNNumber:  sgsnNum,
		SGSNNature:  sgsnNature,
		SGSNPlan:    sgsnPlan,
		SGSNAddress: sgsnAddr,
	}

	if arg.SgsnCapability != nil {
		sc, err := convertWireToSGSNCapability(arg.SgsnCapability)
		if err != nil {
			return nil, fmt.Errorf("SGSNCapability: %w", err)
		}
		u.SGSNCapability = sc
	}

	u.InformPreviousNetworkEntity = nullPtrToBool(arg.InformPreviousNetworkEntity)
	u.PsLCSNotSupportedByUE = nullPtrToBool(arg.PsLCSNotSupportedByUE)

	if arg.VGmlcAddress != nil {
		addr, err := gsn.Parse(*arg.VGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("decoding VGmlcAddress: %w", err)
		}
		u.VGmlcAddress = addr
	}

	if arg.AddInfo != nil {
		ai, err := convertWireToAddInfo(arg.AddInfo)
		if err != nil {
			return nil, fmt.Errorf("AddInfo: %w", err)
		}
		u.AddInfo = ai
	}

	if arg.EpsInfo != nil {
		ei, err := convertWireToEpsInfo(arg.EpsInfo)
		if err != nil {
			return nil, fmt.Errorf("EpsInfo: %w", err)
		}
		u.EpsInfo = ei
	}

	u.ServingNodeTypeIndicator = nullPtrToBool(arg.ServingNodeTypeIndicator)
	u.SkipSubscriberDataUpdate = nullPtrToBool(arg.SkipSubscriberDataUpdate)

	if arg.UsedRATType != nil {
		v := UsedRatType(int64(*arg.UsedRATType))
		u.UsedRatType = &v
	}

	u.GprsSubscriptionDataNotNeeded = nullPtrToBool(arg.GprsSubscriptionDataNotNeeded)
	u.NodeTypeIndicator = nullPtrToBool(arg.NodeTypeIndicator)
	u.AreaRestricted = nullPtrToBool(arg.AreaRestricted)
	u.UeReachableIndicator = nullPtrToBool(arg.UeReachableIndicator)
	u.EpsSubscriptionDataNotNeeded = nullPtrToBool(arg.EpsSubscriptionDataNotNeeded)

	if arg.UeSrvccCapability != nil {
		v := UeSrvccCapability(int64(*arg.UeSrvccCapability))
		u.UeSrvccCapability = &v
	}

	if len(arg.EplmnList) > 0 {
		list := make([]HexBytes, len(arg.EplmnList))
		for i, plmn := range arg.EplmnList {
			if len(plmn) != 3 {
				return nil, fmt.Errorf("UpdateGprsLocation: EplmnList[%d] PLMNId must be exactly 3 octets, got %d", i, len(plmn))
			}
			list[i] = HexBytes(plmn)
		}
		u.EplmnList = list
	}

	if arg.MmeNumberforMTSMS != nil {
		mme, nature, plan, err := decodeAddressField(*arg.MmeNumberforMTSMS)
		if err != nil {
			return nil, fmt.Errorf("decoding MmeNumberForMTSMS: %w", err)
		}
		u.MmeNumberForMTSMS = mme
		u.MmeNumberForMTSMSNature = nature
		u.MmeNumberForMTSMSPlan = plan
	}

	if arg.SmsRegisterRequest != nil {
		v := SmsRegisterRequest(int64(*arg.SmsRegisterRequest))
		u.SmsRegisterRequest = &v
	}

	u.SmsOnly = nullPtrToBool(arg.SmsOnly)

	if arg.SgsnName != nil {
		u.SgsnName = HexBytes(*arg.SgsnName)
	}
	if arg.SgsnRealm != nil {
		u.SgsnRealm = HexBytes(*arg.SgsnRealm)
	}

	u.LgdSupportIndicator = nullPtrToBool(arg.LgdSupportIndicator)
	u.RemovalofMMERegistrationforSMS = nullPtrToBool(arg.RemovalofMMERegistrationforSMS)

	if len(arg.AdjacentPLMNList) > 0 {
		list := make([]HexBytes, len(arg.AdjacentPLMNList))
		for i, plmn := range arg.AdjacentPLMNList {
			if len(plmn) != 3 {
				return nil, fmt.Errorf("UpdateGprsLocation: AdjacentPLMNList[%d] PLMNId must be exactly 3 octets, got %d", i, len(plmn))
			}
			list[i] = HexBytes(plmn)
		}
		u.AdjacentPLMNList = list
	}

	return u, nil
}

// --- SGSNCapability ---

func convertSGSNCapabilityToWire(s *SGSNCapability) (*gsm_map.SGSNCapability, error) {
	out := &gsm_map.SGSNCapability{}

	out.SolsaSupportIndicator = boolToNullPtr(s.SolsaSupportIndicator)

	if s.SuperChargerSupportedInServingNetworkEntity != nil {
		sc, err := convertSuperChargerInfoToWire(s.SuperChargerSupportedInServingNetworkEntity)
		if err != nil {
			return nil, fmt.Errorf("SuperChargerInfo: %w", err)
		}
		out.SuperChargerSupportedInServingNetworkEntity = sc
	}

	out.GprsEnhancementsSupportIndicator = boolToNullPtr(s.GprsEnhancementsSupportIndicator)

	if s.SupportedCamelPhases != nil {
		bs := convertCamelPhasesToBitString(s.SupportedCamelPhases)
		out.SupportedCamelPhases = &bs
	}

	if s.SupportedLCSCapabilitySets != nil {
		bs := convertLCSCapsToBitString(s.SupportedLCSCapabilitySets)
		out.SupportedLCSCapabilitySets = &bs
	}

	if s.OfferedCamel4CSIs != nil {
		bs := convertOfferedCamel4CSIsToBitString(s.OfferedCamel4CSIs)
		out.OfferedCamel4CSIs = &bs
	}

	out.SmsCallBarringSupportIndicator = boolToNullPtr(s.SmsCallBarringSupportIndicator)

	if s.SupportedRATTypesIndicator != nil {
		bs := convertSupportedRATTypesToBitString(s.SupportedRATTypesIndicator)
		out.SupportedRATTypesIndicator = &bs
	}

	if s.SupportedFeaturesBits > 0 {
		if len(s.SupportedFeatures) == 0 || s.SupportedFeaturesBits > len(s.SupportedFeatures)*8 {
			return nil, fmt.Errorf("SGSNCapability: SupportedFeaturesBits (%d) inconsistent with bytes (%d)", s.SupportedFeaturesBits, len(s.SupportedFeatures))
		}
		bs := runtime.BitString{Bytes: append([]byte(nil), s.SupportedFeatures...), BitLength: s.SupportedFeaturesBits}
		out.SupportedFeatures = &bs
	}

	out.TAdsDataRetrieval = boolToNullPtr(s.TAdsDataRetrieval)

	if s.HomogeneousSupportOfIMSVoiceOverPSSessions != nil {
		v := *s.HomogeneousSupportOfIMSVoiceOverPSSessions
		out.HomogeneousSupportOfIMSVoiceOverPSSessions = &v
	}

	out.CancellationTypeInitialAttach = boolToNullPtr(s.CancellationTypeInitialAttach)
	out.MsisdnLessOperationSupported = boolToNullPtr(s.MsisdnLessOperationSupported)
	out.UpdateofHomogeneousSupportOfIMSVoiceOverPSSessions = boolToNullPtr(s.UpdateofHomogeneousSupportOfIMSVoiceOverPSSessions)
	out.ResetIdsSupported = boolToNullPtr(s.ResetIdsSupported)

	if s.ExtSupportedFeaturesBits > 0 {
		if len(s.ExtSupportedFeatures) == 0 || s.ExtSupportedFeaturesBits > len(s.ExtSupportedFeatures)*8 {
			return nil, fmt.Errorf("SGSNCapability: ExtSupportedFeaturesBits (%d) inconsistent with bytes (%d)", s.ExtSupportedFeaturesBits, len(s.ExtSupportedFeatures))
		}
		bs := runtime.BitString{Bytes: append([]byte(nil), s.ExtSupportedFeatures...), BitLength: s.ExtSupportedFeaturesBits}
		out.ExtSupportedFeatures = &bs
	}

	return out, nil
}

func convertWireToSGSNCapability(w *gsm_map.SGSNCapability) (*SGSNCapability, error) {
	out := &SGSNCapability{}

	out.SolsaSupportIndicator = nullPtrToBool(w.SolsaSupportIndicator)

	if w.SuperChargerSupportedInServingNetworkEntity != nil {
		sc, err := convertWireToSuperChargerInfo(w.SuperChargerSupportedInServingNetworkEntity)
		if err != nil {
			return nil, fmt.Errorf("SuperChargerInfo: %w", err)
		}
		out.SuperChargerSupportedInServingNetworkEntity = sc
	}

	out.GprsEnhancementsSupportIndicator = nullPtrToBool(w.GprsEnhancementsSupportIndicator)

	if w.SupportedCamelPhases != nil && w.SupportedCamelPhases.BitLength > 0 {
		out.SupportedCamelPhases = convertBitStringToCamelPhases(*w.SupportedCamelPhases)
	}

	if w.SupportedLCSCapabilitySets != nil && w.SupportedLCSCapabilitySets.BitLength > 0 {
		out.SupportedLCSCapabilitySets = convertBitStringToLCSCaps(*w.SupportedLCSCapabilitySets)
	}

	if w.OfferedCamel4CSIs != nil && w.OfferedCamel4CSIs.BitLength > 0 {
		out.OfferedCamel4CSIs = convertBitStringToOfferedCamel4CSIs(*w.OfferedCamel4CSIs)
	}

	out.SmsCallBarringSupportIndicator = nullPtrToBool(w.SmsCallBarringSupportIndicator)

	if w.SupportedRATTypesIndicator != nil && w.SupportedRATTypesIndicator.BitLength > 0 {
		if w.SupportedRATTypesIndicator.BitLength < 2 || w.SupportedRATTypesIndicator.BitLength > 8 {
			return nil, fmt.Errorf("SGSNCapability: SupportedRATTypes BitLength must be 2..8, got %d", w.SupportedRATTypesIndicator.BitLength)
		}
		out.SupportedRATTypesIndicator = convertBitStringToSupportedRATTypes(*w.SupportedRATTypesIndicator)
	}

	if w.SupportedFeatures != nil && w.SupportedFeatures.BitLength > 0 {
		out.SupportedFeatures = HexBytes(append([]byte(nil), w.SupportedFeatures.Bytes...))
		out.SupportedFeaturesBits = w.SupportedFeatures.BitLength
	}

	out.TAdsDataRetrieval = nullPtrToBool(w.TAdsDataRetrieval)

	if w.HomogeneousSupportOfIMSVoiceOverPSSessions != nil {
		v := *w.HomogeneousSupportOfIMSVoiceOverPSSessions
		out.HomogeneousSupportOfIMSVoiceOverPSSessions = &v
	}

	out.CancellationTypeInitialAttach = nullPtrToBool(w.CancellationTypeInitialAttach)
	out.MsisdnLessOperationSupported = nullPtrToBool(w.MsisdnLessOperationSupported)
	out.UpdateofHomogeneousSupportOfIMSVoiceOverPSSessions = nullPtrToBool(w.UpdateofHomogeneousSupportOfIMSVoiceOverPSSessions)
	out.ResetIdsSupported = nullPtrToBool(w.ResetIdsSupported)

	if w.ExtSupportedFeatures != nil && w.ExtSupportedFeatures.BitLength > 0 {
		out.ExtSupportedFeatures = HexBytes(append([]byte(nil), w.ExtSupportedFeatures.Bytes...))
		out.ExtSupportedFeaturesBits = w.ExtSupportedFeatures.BitLength
	}

	return out, nil
}

// --- EpsInfo CHOICE ---

func convertEpsInfoToWire(e *EpsInfo) (*gsm_map.EPSInfo, error) {
	hasPdn := e.PdnGwUpdate != nil
	hasIsr := e.IsrInformationBits > 0
	if hasPdn && hasIsr {
		return nil, ErrSriChoiceMultipleAlternatives
	}
	if !hasPdn && !hasIsr {
		return nil, ErrSriChoiceNoAlternative
	}
	if hasPdn {
		pgu, err := convertPdnGwUpdateToWire(e.PdnGwUpdate)
		if err != nil {
			return nil, err
		}
		v := gsm_map.NewEPSInfoPdnGwUpdate(*pgu)
		return &v, nil
	}
	if len(e.IsrInformation) == 0 || e.IsrInformationBits > len(e.IsrInformation)*8 {
		return nil, fmt.Errorf("EpsInfo: IsrInformationBits (%d) inconsistent with bytes (%d)", e.IsrInformationBits, len(e.IsrInformation))
	}
	bs := runtime.BitString{Bytes: append([]byte(nil), e.IsrInformation...), BitLength: e.IsrInformationBits}
	v := gsm_map.NewEPSInfoIsrInformation(bs)
	return &v, nil
}

func convertWireToEpsInfo(w *gsm_map.EPSInfo) (*EpsInfo, error) {
	switch w.Choice {
	case gsm_map.EPSInfoChoicePdnGwUpdate:
		if w.PdnGwUpdate == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		return &EpsInfo{PdnGwUpdate: convertWireToPdnGwUpdate(w.PdnGwUpdate)}, nil
	case gsm_map.EPSInfoChoiceIsrInformation:
		out := &EpsInfo{}
		if w.IsrInformation != nil {
			out.IsrInformation = HexBytes(append([]byte(nil), w.IsrInformation.Bytes...))
			out.IsrInformationBits = w.IsrInformation.BitLength
		}
		return out, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertPdnGwUpdateToWire(p *PdnGwUpdate) (*gsm_map.PDNGWUpdate, error) {
	out := &gsm_map.PDNGWUpdate{}
	if len(p.APN) > 0 {
		apn := gsm_map.APN(append([]byte(nil), p.APN...))
		out.Apn = &apn
	}
	if p.PdnGwIdentity != nil {
		id, err := convertPdnGwIdentityToWire(p.PdnGwIdentity)
		if err != nil {
			return nil, err
		}
		out.PdnGwIdentity = id
	}
	if p.ContextID != nil {
		v := gsm_map.ContextId(int64(*p.ContextID))
		out.ContextId = &v
	}
	return out, nil
}

func convertWireToPdnGwUpdate(w *gsm_map.PDNGWUpdate) *PdnGwUpdate {
	out := &PdnGwUpdate{}
	if w.Apn != nil {
		out.APN = HexBytes(append([]byte(nil), (*w.Apn)...))
	}
	if w.PdnGwIdentity != nil {
		out.PdnGwIdentity = convertWireToPdnGwIdentity(w.PdnGwIdentity)
	}
	if w.ContextId != nil {
		v := int(*w.ContextId)
		out.ContextID = &v
	}
	return out
}

func convertPdnGwIdentityToWire(p *PdnGwIdentity) (*gsm_map.PDNGWIdentity, error) {
	if len(p.IPv4Address) > 0 && len(p.IPv4Address) != 4 {
		return nil, fmt.Errorf("PdnGwIdentity: IPv4Address must be exactly 4 octets, got %d", len(p.IPv4Address))
	}
	if len(p.IPv6Address) > 0 && len(p.IPv6Address) != 16 {
		return nil, fmt.Errorf("PdnGwIdentity: IPv6Address must be exactly 16 octets, got %d", len(p.IPv6Address))
	}
	if len(p.IPv4Address) == 0 && len(p.IPv6Address) == 0 && len(p.Name) == 0 {
		return nil, fmt.Errorf("PdnGwIdentity: at least one of IPv4Address, IPv6Address, or Name must be set")
	}
	out := &gsm_map.PDNGWIdentity{}
	if len(p.IPv4Address) > 0 {
		v := gsm_map.PDPAddress(append([]byte(nil), p.IPv4Address...))
		out.PdnGwIpv4Address = &v
	}
	if len(p.IPv6Address) > 0 {
		v := gsm_map.PDPAddress(append([]byte(nil), p.IPv6Address...))
		out.PdnGwIpv6Address = &v
	}
	if len(p.Name) > 0 {
		v := gsm_map.FQDN(append([]byte(nil), p.Name...))
		out.PdnGwName = &v
	}
	return out, nil
}

func convertWireToPdnGwIdentity(w *gsm_map.PDNGWIdentity) *PdnGwIdentity {
	out := &PdnGwIdentity{}
	if w.PdnGwIpv4Address != nil {
		out.IPv4Address = HexBytes(append([]byte(nil), (*w.PdnGwIpv4Address)...))
	}
	if w.PdnGwIpv6Address != nil {
		out.IPv6Address = HexBytes(append([]byte(nil), (*w.PdnGwIpv6Address)...))
	}
	if w.PdnGwName != nil {
		out.Name = HexBytes(append([]byte(nil), (*w.PdnGwName)...))
	}
	return out
}

// --- UpdateGprsLocationRes ---

func convertUpdateGprsLocationResToRes(u *UpdateGprsLocationRes) (*gsm_map.UpdateGprsLocationRes, error) {
	hlr, err := encodeAddressField(u.HLRNumber, u.HLRNumberNature, u.HLRNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding HLRNumber: %w", err)
	}

	return &gsm_map.UpdateGprsLocationRes{
		HlrNumber:                  gsm_map.ISDNAddressString(hlr),
		AddCapability:              boolToNullPtr(u.AddCapability),
		SgsnMmeSeparationSupported: boolToNullPtr(u.SgsnMmeSeparationSupported),
		MmeRegisteredforSMS:        boolToNullPtr(u.MmeRegisteredforSMS),
	}, nil
}

func convertResToUpdateGprsLocationRes(res *gsm_map.UpdateGprsLocationRes) (*UpdateGprsLocationRes, error) {
	hlr, nature, plan, err := decodeAddressField(res.HlrNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding HLRNumber: %w", err)
	}

	return &UpdateGprsLocationRes{
		HLRNumber:                  hlr,
		HLRNumberNature:            nature,
		HLRNumberPlan:              plan,
		AddCapability:              nullPtrToBool(res.AddCapability),
		SgsnMmeSeparationSupported: nullPtrToBool(res.SgsnMmeSeparationSupported),
		MmeRegisteredforSMS:        nullPtrToBool(res.MmeRegisteredforSMS),
	}, nil
}

// --- AnyTimeInterrogation ---

func convertATIToArg(ati *AnyTimeInterrogation) (*gsm_map.AnyTimeInterrogationArg, error) {
	if ati.SubscriberIdentity.IMSI == "" && ati.SubscriberIdentity.MSISDN == "" {
		return nil, fmt.Errorf("subscriber identity required: set either IMSI or MSISDN")
	}
	if ati.SubscriberIdentity.IMSI != "" && ati.SubscriberIdentity.MSISDN != "" {
		return nil, fmt.Errorf("subscriber identity ambiguous: set either IMSI or MSISDN, not both")
	}

	var subId gsm_map.SubscriberIdentity
	if ati.SubscriberIdentity.IMSI != "" {
		imsiBytes, err := tbcd.Encode(ati.SubscriberIdentity.IMSI)
		if err != nil {
			return nil, fmt.Errorf(errEncodingIMSI, err)
		}
		subId = gsm_map.NewSubscriberIdentityImsi(gsm_map.IMSI(imsiBytes))
	} else {
		msisdnBytes, err := encodeAddressField(ati.SubscriberIdentity.MSISDN, 0, 0)
		if err != nil {
			return nil, fmt.Errorf("encoding MSISDN: %w", err)
		}
		subId = gsm_map.NewSubscriberIdentityMsisdn(gsm_map.ISDNAddressString(msisdnBytes))
	}

	reqInfo := buildMSRequestedInfo(&ati.RequestedInfo)

	scfAddr, err := encodeAddressField(ati.GsmSCFAddress, ati.GsmSCFNature, ati.GsmSCFPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding GsmSCFAddress: %w", err)
	}

	return &gsm_map.AnyTimeInterrogationArg{
		SubscriberIdentity: subId,
		RequestedInfo:      reqInfo,
		GsmSCFAddress:      gsm_map.ISDNAddressString(scfAddr),
	}, nil
}

func buildMSRequestedInfo(ri *RequestedInfo) gsm_map.MSRequestedInfo {
	var msri gsm_map.MSRequestedInfo

	nullMarker := &struct{}{}

	if ri.LocationInformation {
		msri.LocationInformation = nullMarker
	}
	if ri.SubscriberState {
		msri.SubscriberState = nullMarker
	}
	if ri.CurrentLocation {
		msri.CurrentLocation = nullMarker
	}
	if ri.RequestedDomain != nil {
		dt := gsm_map.DomainType(*ri.RequestedDomain)
		msri.RequestedDomain = &dt
	}
	if ri.MsClassmark {
		msri.MsClassmark = nullMarker
	}
	if ri.IMEI {
		msri.Imei = nullMarker
	}
	if ri.MnpRequestedInfo {
		msri.MnpRequestedInfo = nullMarker
	}
	if ri.LocationInformationEPSSupported {
		msri.LocationInformationEPSSupported = nullMarker
	}
	if ri.TAdsData {
		msri.TAdsData = nullMarker
	}
	if ri.RequestedNodes != nil {
		bs := convertRequestedNodesToBitString(ri.RequestedNodes)
		msri.RequestedNodes = &bs
	}
	if ri.ServingNodeIndication {
		msri.ServingNodeIndication = nullMarker
	}
	if ri.LocalTimeZoneRequest {
		msri.LocalTimeZoneRequest = nullMarker
	}

	return msri
}

func convertArgToATI(arg *gsm_map.AnyTimeInterrogationArg) (*AnyTimeInterrogation, error) {
	var ati AnyTimeInterrogation

	// SubscriberIdentity
	switch arg.SubscriberIdentity.Choice {
	case gsm_map.SubscriberIdentityChoiceImsi:
		if arg.SubscriberIdentity.Imsi == nil {
			return nil, fmt.Errorf("SubscriberIdentity IMSI is nil")
		}
		imsi, err := tbcd.Decode(*arg.SubscriberIdentity.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding IMSI: %w", err)
		}
		ati.SubscriberIdentity.IMSI = imsi
	case gsm_map.SubscriberIdentityChoiceMsisdn:
		if arg.SubscriberIdentity.Msisdn == nil {
			return nil, fmt.Errorf("SubscriberIdentity MSISDN is nil")
		}
		msisdn, _, _, err := decodeAddressField(*arg.SubscriberIdentity.Msisdn)
		if err != nil {
			return nil, fmt.Errorf("decoding MSISDN: %w", err)
		}
		ati.SubscriberIdentity.MSISDN = msisdn
	default:
		return nil, fmt.Errorf("unknown SubscriberIdentity choice: %d", arg.SubscriberIdentity.Choice)
	}

	// RequestedInfo
	ri := arg.RequestedInfo
	ati.RequestedInfo.LocationInformation = ri.LocationInformation != nil
	ati.RequestedInfo.SubscriberState = ri.SubscriberState != nil
	ati.RequestedInfo.CurrentLocation = ri.CurrentLocation != nil
	ati.RequestedInfo.MsClassmark = ri.MsClassmark != nil
	ati.RequestedInfo.IMEI = ri.Imei != nil
	ati.RequestedInfo.MnpRequestedInfo = ri.MnpRequestedInfo != nil
	ati.RequestedInfo.LocationInformationEPSSupported = ri.LocationInformationEPSSupported != nil
	ati.RequestedInfo.TAdsData = ri.TAdsData != nil
	ati.RequestedInfo.ServingNodeIndication = ri.ServingNodeIndication != nil
	ati.RequestedInfo.LocalTimeZoneRequest = ri.LocalTimeZoneRequest != nil

	if ri.RequestedDomain != nil {
		domain := DomainType(*ri.RequestedDomain)
		// Per spec: values > 1 shall be mapped to cs-Domain
		if domain > PsDomain {
			domain = CsDomain
		}
		ati.RequestedInfo.RequestedDomain = &domain
	}

	if ri.RequestedNodes != nil && ri.RequestedNodes.BitLength > 0 {
		ati.RequestedInfo.RequestedNodes = convertBitStringToRequestedNodes(*ri.RequestedNodes)
	}

	// GsmSCFAddress
	scf, scfNature, scfPlan, err := decodeAddressField(arg.GsmSCFAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding GsmSCFAddress: %w", err)
	}
	ati.GsmSCFAddress = scf
	ati.GsmSCFNature = scfNature
	ati.GsmSCFPlan = scfPlan

	return &ati, nil
}

// --- AnyTimeInterrogation Response ---

func convertATIResToRes(atiRes *AnyTimeInterrogationRes) (*gsm_map.AnyTimeInterrogationRes, error) {
	si, err := convertSubscriberInfoToWire(&atiRes.SubscriberInfo)
	if err != nil {
		return nil, err
	}
	return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: *si}, nil
}

func convertResToATIRes(res *gsm_map.AnyTimeInterrogationRes) (*AnyTimeInterrogationRes, error) {
	si, err := convertWireToSubscriberInfo(&res.SubscriberInfo)
	if err != nil {
		return nil, err
	}
	return &AnyTimeInterrogationRes{SubscriberInfo: *si}, nil
}

// --- CS Location conversion ---

func convertCSLocationToAsn1(loc *CSLocationInformation) (*gsm_map.LocationInformation, error) {
	li := &gsm_map.LocationInformation{}

	if loc.AgeOfLocationInformation != nil {
		age := gsm_map.AgeOfLocationInformation(*loc.AgeOfLocationInformation)
		li.AgeOfLocationInformation = &age
	}

	if loc.VlrNumber != "" {
		vlr, err := encodeAddressField(loc.VlrNumber, loc.VlrNumberNature, loc.VlrNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding VlrNumber: %w", err)
		}
		vlrAddr := gsm_map.ISDNAddressString(vlr)
		li.VlrNumber = &vlrAddr
	}

	if loc.MscNumber != "" {
		msc, err := encodeAddressField(loc.MscNumber, loc.MscNumberNature, loc.MscNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding MscNumber: %w", err)
		}
		mscAddr := gsm_map.ISDNAddressString(msc)
		li.MscNumber = &mscAddr
	}

	if loc.GeographicalInformation != nil {
		raw, err := loc.GeographicalInformation.Encode()
		if err != nil {
			return nil, fmt.Errorf("encoding GeographicalInformation: %w", err)
		}
		gi := gsm_map.GeographicalInformation(raw)
		li.GeographicalInformation = &gi
	}

	if loc.GeodeticInformation != nil {
		gd := gsm_map.GeodeticInformation(loc.GeodeticInformation)
		li.GeodeticInformation = &gd
	}

	if loc.CellGlobalId != nil {
		v := gsm_map.NewCellGlobalIdOrServiceAreaIdOrLAICellGlobalIdOrServiceAreaIdFixedLength(
			gsm_map.CellGlobalIdOrServiceAreaIdFixedLength(loc.CellGlobalId),
		)
		li.CellGlobalIdOrServiceAreaIdOrLAI = &v
	} else if loc.LAI != nil {
		v := gsm_map.NewCellGlobalIdOrServiceAreaIdOrLAILaiFixedLength(
			gsm_map.LAIFixedLength(loc.LAI),
		)
		li.CellGlobalIdOrServiceAreaIdOrLAI = &v
	}

	if loc.LocationNumber != nil {
		ln := gsm_map.LocationNumber(loc.LocationNumber)
		li.LocationNumber = &ln
	}

	if loc.SelectedLSAId != nil {
		lsa := gsm_map.LSAIdentity(loc.SelectedLSAId)
		li.SelectedLSAId = &lsa
	}

	if loc.UserCSGInformation != nil {
		csg, err := convertUserCSGInformationToWire(loc.UserCSGInformation)
		if err != nil {
			return nil, fmt.Errorf("UserCSGInformation: %w", err)
		}
		li.UserCSGInformation = csg
	}

	if loc.CurrentLocationRetrieved {
		li.CurrentLocationRetrieved = &struct{}{}
	}

	if loc.SAIPresent {
		li.SaiPresent = &struct{}{}
	}

	return li, nil
}

func convertAsn1ToCSLocation(li *gsm_map.LocationInformation) (*CSLocationInformation, error) {
	loc := &CSLocationInformation{}

	if li.AgeOfLocationInformation != nil {
		v := int(*li.AgeOfLocationInformation)
		loc.AgeOfLocationInformation = &v
	}

	if li.VlrNumber != nil {
		vlr, nature, plan, err := decodeAddressField(*li.VlrNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding VlrNumber: %w", err)
		}
		loc.VlrNumber = vlr
		loc.VlrNumberNature = nature
		loc.VlrNumberPlan = plan
	}

	if li.MscNumber != nil {
		msc, nature, plan, err := decodeAddressField(*li.MscNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding MscNumber: %w", err)
		}
		loc.MscNumber = msc
		loc.MscNumberNature = nature
		loc.MscNumberPlan = plan
	}

	if li.GeographicalInformation != nil {
		gi, err := DecodeGeographicalInfo([]byte(*li.GeographicalInformation))
		if err != nil {
			return nil, fmt.Errorf("decoding GeographicalInformation: %w", err)
		}
		loc.GeographicalInformation = gi
	}

	if li.GeodeticInformation != nil {
		loc.GeodeticInformation = []byte(*li.GeodeticInformation)
	}

	if li.CellGlobalIdOrServiceAreaIdOrLAI != nil {
		choice := li.CellGlobalIdOrServiceAreaIdOrLAI
		switch choice.Choice {
		case gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceCellGlobalIdOrServiceAreaIdFixedLength:
			if choice.CellGlobalIdOrServiceAreaIdFixedLength != nil {
				loc.CellGlobalId = []byte(*choice.CellGlobalIdOrServiceAreaIdFixedLength)
			}
		case gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceLaiFixedLength:
			if choice.LaiFixedLength != nil {
				loc.LAI = []byte(*choice.LaiFixedLength)
			}
		}
	}

	if li.LocationNumber != nil {
		loc.LocationNumber = []byte(*li.LocationNumber)
	}

	if li.SelectedLSAId != nil {
		loc.SelectedLSAId = []byte(*li.SelectedLSAId)
	}

	if li.UserCSGInformation != nil {
		loc.UserCSGInformation = convertWireToUserCSGInformation(li.UserCSGInformation)
	}

	loc.CurrentLocationRetrieved = li.CurrentLocationRetrieved != nil
	loc.SAIPresent = li.SaiPresent != nil

	return loc, nil
}

// --- SubscriberState conversion ---

func convertSubscriberStateToAsn1(ss *SubscriberStateInfo) *gsm_map.SubscriberState {
	var s gsm_map.SubscriberState
	switch ss.State {
	case StateAssumedIdle:
		s = gsm_map.NewSubscriberStateAssumedIdle(struct{}{})
	case StateCamelBusy:
		s = gsm_map.NewSubscriberStateCamelBusy(struct{}{})
	case StateNetDetNotReachable:
		if ss.NotReachableReason != nil {
			s = gsm_map.NewSubscriberStateNetDetNotReachable(gsm_map.NotReachableReason(*ss.NotReachableReason))
		} else {
			s = gsm_map.NewSubscriberStateNetDetNotReachable(gsm_map.NotReachableReason(0))
		}
	case StateNotProvidedFromVLR:
		s = gsm_map.NewSubscriberStateNotProvidedFromVLR(struct{}{})
	}
	return &s
}

func convertAsn1ToSubscriberState(ss *gsm_map.SubscriberState) *SubscriberStateInfo {
	info := &SubscriberStateInfo{}
	switch ss.Choice {
	case gsm_map.SubscriberStateChoiceAssumedIdle:
		info.State = StateAssumedIdle
	case gsm_map.SubscriberStateChoiceCamelBusy:
		info.State = StateCamelBusy
	case gsm_map.SubscriberStateChoiceNetDetNotReachable:
		info.State = StateNetDetNotReachable
		if ss.NetDetNotReachable != nil {
			reason := int(*ss.NetDetNotReachable)
			info.NotReachableReason = &reason
		}
	case gsm_map.SubscriberStateChoiceNotProvidedFromVLR:
		info.State = StateNotProvidedFromVLR
	}
	return info
}

// --- EPS Location conversion ---

func convertEPSLocationToAsn1(loc *EPSLocationInformation) (*gsm_map.LocationInformationEPS, error) {
	li := &gsm_map.LocationInformationEPS{}

	if loc.AgeOfLocationInformation != nil {
		age := gsm_map.AgeOfLocationInformation(*loc.AgeOfLocationInformation)
		li.AgeOfLocationInformation = &age
	}

	if loc.EUtranCellGlobalIdentity != nil {
		cgi := gsm_map.EUTRANCGI(loc.EUtranCellGlobalIdentity)
		li.EUtranCellGlobalIdentity = &cgi
	}

	if loc.TrackingAreaIdentity != nil {
		ta := gsm_map.TAId(loc.TrackingAreaIdentity)
		li.TrackingAreaIdentity = &ta
	}

	if loc.GeographicalInformation != nil {
		raw, err := loc.GeographicalInformation.Encode()
		if err != nil {
			return nil, fmt.Errorf("encoding GeographicalInformation: %w", err)
		}
		gi := gsm_map.GeographicalInformation(raw)
		li.GeographicalInformation = &gi
	}

	if loc.GeodeticInformation != nil {
		gd := gsm_map.GeodeticInformation(loc.GeodeticInformation)
		li.GeodeticInformation = &gd
	}

	if loc.CurrentLocationRetrieved {
		li.CurrentLocationRetrieved = &struct{}{}
	}

	if loc.MmeName != nil {
		mm := gsm_map.DiameterIdentity(loc.MmeName)
		li.MmeName = &mm
	}

	return li, nil
}

func convertAsn1ToEPSLocation(li *gsm_map.LocationInformationEPS) (*EPSLocationInformation, error) {
	loc := &EPSLocationInformation{}

	if li.AgeOfLocationInformation != nil {
		v := int(*li.AgeOfLocationInformation)
		loc.AgeOfLocationInformation = &v
	}

	if li.EUtranCellGlobalIdentity != nil {
		loc.EUtranCellGlobalIdentity = []byte(*li.EUtranCellGlobalIdentity)
	}

	if li.TrackingAreaIdentity != nil {
		loc.TrackingAreaIdentity = []byte(*li.TrackingAreaIdentity)
	}

	if li.GeographicalInformation != nil {
		gi, err := DecodeGeographicalInfo([]byte(*li.GeographicalInformation))
		if err != nil {
			return nil, fmt.Errorf("decoding GeographicalInformation: %w", err)
		}
		loc.GeographicalInformation = gi
	}

	if li.GeodeticInformation != nil {
		loc.GeodeticInformation = []byte(*li.GeodeticInformation)
	}

	loc.CurrentLocationRetrieved = li.CurrentLocationRetrieved != nil

	if li.MmeName != nil {
		loc.MmeName = []byte(*li.MmeName)
	}

	return loc, nil
}

// --- GPRS Location conversion ---

func convertGPRSLocationToAsn1(loc *GPRSLocationInformation) (*gsm_map.LocationInformationGPRS, error) {
	li := &gsm_map.LocationInformationGPRS{}

	if loc.AgeOfLocationInformation != nil {
		age := gsm_map.AgeOfLocationInformation(*loc.AgeOfLocationInformation)
		li.AgeOfLocationInformation = &age
	}

	if loc.CellGlobalId != nil {
		v := gsm_map.NewCellGlobalIdOrServiceAreaIdOrLAICellGlobalIdOrServiceAreaIdFixedLength(
			gsm_map.CellGlobalIdOrServiceAreaIdFixedLength(loc.CellGlobalId),
		)
		li.CellGlobalIdOrServiceAreaIdOrLAI = &v
	} else if loc.LAI != nil {
		v := gsm_map.NewCellGlobalIdOrServiceAreaIdOrLAILaiFixedLength(
			gsm_map.LAIFixedLength(loc.LAI),
		)
		li.CellGlobalIdOrServiceAreaIdOrLAI = &v
	}

	if loc.RouteingAreaIdentity != nil {
		ra := gsm_map.RAIdentity(loc.RouteingAreaIdentity)
		li.RouteingAreaIdentity = &ra
	}

	if loc.GeographicalInformation != nil {
		raw, err := loc.GeographicalInformation.Encode()
		if err != nil {
			return nil, fmt.Errorf("encoding GeographicalInformation: %w", err)
		}
		gi := gsm_map.GeographicalInformation(raw)
		li.GeographicalInformation = &gi
	}

	if loc.GeodeticInformation != nil {
		gd := gsm_map.GeodeticInformation(loc.GeodeticInformation)
		li.GeodeticInformation = &gd
	}

	if loc.SgsnNumber != "" {
		sgsn, err := encodeAddressField(loc.SgsnNumber, loc.SgsnNumberNature, loc.SgsnNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SgsnNumber: %w", err)
		}
		sgsnAddr := gsm_map.ISDNAddressString(sgsn)
		li.SgsnNumber = &sgsnAddr
	}

	if loc.SelectedLSAIdentity != nil {
		lsa := gsm_map.LSAIdentity(loc.SelectedLSAIdentity)
		li.SelectedLSAIdentity = &lsa
	}

	if loc.UserCSGInformation != nil {
		csg, err := convertUserCSGInformationToWire(loc.UserCSGInformation)
		if err != nil {
			return nil, fmt.Errorf("UserCSGInformation: %w", err)
		}
		li.UserCSGInformation = csg
	}

	if loc.CurrentLocationRetrieved {
		li.CurrentLocationRetrieved = &struct{}{}
	}

	if loc.SAIPresent {
		li.SaiPresent = &struct{}{}
	}

	return li, nil
}

func convertAsn1ToGPRSLocation(li *gsm_map.LocationInformationGPRS) (*GPRSLocationInformation, error) {
	loc := &GPRSLocationInformation{}

	if li.AgeOfLocationInformation != nil {
		v := int(*li.AgeOfLocationInformation)
		loc.AgeOfLocationInformation = &v
	}

	if li.CellGlobalIdOrServiceAreaIdOrLAI != nil {
		choice := li.CellGlobalIdOrServiceAreaIdOrLAI
		switch choice.Choice {
		case gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceCellGlobalIdOrServiceAreaIdFixedLength:
			if choice.CellGlobalIdOrServiceAreaIdFixedLength != nil {
				loc.CellGlobalId = []byte(*choice.CellGlobalIdOrServiceAreaIdFixedLength)
			}
		case gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceLaiFixedLength:
			if choice.LaiFixedLength != nil {
				loc.LAI = []byte(*choice.LaiFixedLength)
			}
		}
	}

	if li.RouteingAreaIdentity != nil {
		loc.RouteingAreaIdentity = []byte(*li.RouteingAreaIdentity)
	}

	if li.GeographicalInformation != nil {
		gi, err := DecodeGeographicalInfo([]byte(*li.GeographicalInformation))
		if err != nil {
			return nil, fmt.Errorf("decoding GeographicalInformation: %w", err)
		}
		loc.GeographicalInformation = gi
	}

	if li.GeodeticInformation != nil {
		loc.GeodeticInformation = []byte(*li.GeodeticInformation)
	}

	if li.SgsnNumber != nil {
		sgsn, nature, plan, err := decodeAddressField(*li.SgsnNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding SgsnNumber: %w", err)
		}
		loc.SgsnNumber = sgsn
		loc.SgsnNumberNature = nature
		loc.SgsnNumberPlan = plan
	}

	if li.SelectedLSAIdentity != nil {
		loc.SelectedLSAIdentity = []byte(*li.SelectedLSAIdentity)
	}

	if li.UserCSGInformation != nil {
		loc.UserCSGInformation = convertWireToUserCSGInformation(li.UserCSGInformation)
	}

	loc.CurrentLocationRetrieved = li.CurrentLocationRetrieved != nil
	loc.SAIPresent = li.SaiPresent != nil

	return loc, nil
}

// --- BitString helpers ---

func convertCamelPhasesToBitString(cp *SupportedCamelPhases) runtime.BitString {
	var b byte
	bitLen := 1
	if cp.Phase1 {
		b |= 0x80
	}
	if cp.Phase2 {
		b |= 0x40
		bitLen = 2
	}
	if cp.Phase3 {
		b |= 0x20
		bitLen = 3
	}
	if cp.Phase4 {
		b |= 0x10
		bitLen = 4
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: bitLen}
}

func convertBitStringToCamelPhases(bs runtime.BitString) *SupportedCamelPhases {
	cp := &SupportedCamelPhases{}
	if bs.BitLength > 0 {
		cp.Phase1 = bs.Has(0)
	}
	if bs.BitLength > 1 {
		cp.Phase2 = bs.Has(1)
	}
	if bs.BitLength > 2 {
		cp.Phase3 = bs.Has(2)
	}
	if bs.BitLength > 3 {
		cp.Phase4 = bs.Has(3)
	}
	return cp
}

func convertLCSCapsToBitString(lcs *SupportedLCSCapabilitySets) runtime.BitString {
	var b byte
	bitLen := 2 // minimum per spec
	if lcs.LcsCapabilitySet1 {
		b |= 0x80
	}
	if lcs.LcsCapabilitySet2 {
		b |= 0x40
	}
	if lcs.LcsCapabilitySet3 {
		b |= 0x20
		bitLen = 3
	}
	if lcs.LcsCapabilitySet4 {
		b |= 0x10
		bitLen = 4
	}
	if lcs.LcsCapabilitySet5 {
		b |= 0x08
		bitLen = 5
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: bitLen}
}

func convertBitStringToLCSCaps(bs runtime.BitString) *SupportedLCSCapabilitySets {
	lcs := &SupportedLCSCapabilitySets{}
	if bs.BitLength > 0 {
		lcs.LcsCapabilitySet1 = bs.Has(0)
	}
	if bs.BitLength > 1 {
		lcs.LcsCapabilitySet2 = bs.Has(1)
	}
	if bs.BitLength > 2 {
		lcs.LcsCapabilitySet3 = bs.Has(2)
	}
	if bs.BitLength > 3 {
		lcs.LcsCapabilitySet4 = bs.Has(3)
	}
	if bs.BitLength > 4 {
		lcs.LcsCapabilitySet5 = bs.Has(4)
	}
	return lcs
}

func convertRequestedNodesToBitString(rn *RequestedNodes) runtime.BitString {
	var b byte
	bitLen := 1
	if rn.MME {
		b |= 0x80
	}
	if rn.SGSN {
		b |= 0x40
		bitLen = 2
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: bitLen}
}

func convertBitStringToRequestedNodes(bs runtime.BitString) *RequestedNodes {
	rn := &RequestedNodes{}
	if bs.BitLength > 0 {
		rn.MME = bs.Has(0)
	}
	if bs.BitLength > 1 {
		rn.SGSN = bs.Has(1)
	}
	return rn
}

// --- SRI BitString helpers ---

// AllowedServices: 2 bits (bit 0 = first, bit 1 = second).
func convertAllowedServicesToBitString(a *AllowedServicesFlags) runtime.BitString {
	var b byte
	if a.FirstServiceAllowed {
		b |= 0x80
	}
	if a.SecondServiceAllowed {
		b |= 0x40
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: 2}
}

func convertBitStringToAllowedServices(bs runtime.BitString) *AllowedServicesFlags {
	a := &AllowedServicesFlags{}
	if bs.BitLength > 0 {
		a.FirstServiceAllowed = bs.Has(0)
	}
	if bs.BitLength > 1 {
		a.SecondServiceAllowed = bs.Has(1)
	}
	return a
}

// SuppressMTSS: 2 bits (bit 0 = suppressCUG, bit 1 = suppressCCBS), min size 2.
func convertSuppressMTSSToBitString(s *SuppressMTSSFlags) runtime.BitString {
	var b byte
	if s.SuppressCUG {
		b |= 0x80
	}
	if s.SuppressCCBS {
		b |= 0x40
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: 2}
}

func convertBitStringToSuppressMTSS(bs runtime.BitString) *SuppressMTSSFlags {
	s := &SuppressMTSSFlags{}
	if bs.BitLength > 0 {
		s.SuppressCUG = bs.Has(0)
	}
	if bs.BitLength > 1 {
		s.SuppressCCBS = bs.Has(1)
	}
	return s
}

// OfferedCamel4CSIs: 7 bits per 3GPP TS 29.002.
// Bit order: 0=o-CSI, 1=d-CSI, 2=vt-CSI, 3=t-CSI, 4=mt-sms-CSI, 5=mg-CSI, 6=psi-enhancements.
func convertOfferedCamel4CSIsToBitString(o *OfferedCamel4CSIs) runtime.BitString {
	var b byte
	if o.OCSI {
		b |= 1 << 7
	}
	if o.DCSI {
		b |= 1 << 6
	}
	if o.VTCSI {
		b |= 1 << 5
	}
	if o.TCSI {
		b |= 1 << 4
	}
	if o.MTSMSCSI {
		b |= 1 << 3
	}
	if o.MGCSI {
		b |= 1 << 2
	}
	if o.PsiEnhancements {
		b |= 1 << 1
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: 7}
}

func convertBitStringToOfferedCamel4CSIs(bs runtime.BitString) *OfferedCamel4CSIs {
	o := &OfferedCamel4CSIs{}
	if bs.BitLength > 0 {
		o.OCSI = bs.Has(0)
	}
	if bs.BitLength > 1 {
		o.DCSI = bs.Has(1)
	}
	if bs.BitLength > 2 {
		o.VTCSI = bs.Has(2)
	}
	if bs.BitLength > 3 {
		o.TCSI = bs.Has(3)
	}
	if bs.BitLength > 4 {
		o.MTSMSCSI = bs.Has(4)
	}
	if bs.BitLength > 5 {
		o.MGCSI = bs.Has(5)
	}
	if bs.BitLength > 6 {
		o.PsiEnhancements = bs.Has(6)
	}
	return o
}

// --- UpdateLocation helpers ---

// SupportedRATTypes: bit 0=utran, 1=geran, 2=gan, 3=i-hspa-evolution, 4=e-utran.
func convertSupportedRATTypesToBitString(r *SupportedRATTypes) runtime.BitString {
	var b byte
	if r.UTRAN {
		b |= 0x80
	}
	if r.GERAN {
		b |= 0x40
	}
	if r.GAN {
		b |= 0x20
	}
	if r.IHSPAEvolution {
		b |= 0x10
	}
	if r.EUTRAN {
		b |= 0x08
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: 5}
}

func convertBitStringToSupportedRATTypes(bs runtime.BitString) *SupportedRATTypes {
	r := &SupportedRATTypes{}
	if bs.BitLength > 0 {
		r.UTRAN = bs.Has(0)
	}
	if bs.BitLength > 1 {
		r.GERAN = bs.Has(1)
	}
	if bs.BitLength > 2 {
		r.GAN = bs.Has(2)
	}
	if bs.BitLength > 3 {
		r.IHSPAEvolution = bs.Has(3)
	}
	if bs.BitLength > 4 {
		r.EUTRAN = bs.Has(4)
	}
	return r
}

func convertSuperChargerInfoToWire(s *SuperChargerInfo) (*gsm_map.SuperChargerInfo, error) {
	hasSend := s.SendSubscriberData
	hasStored := len(s.SubscriberDataStored) > 0
	if hasSend && hasStored {
		return nil, ErrSuperChargerInfoMultipleAlternatives
	}
	if !hasSend && !hasStored {
		return nil, ErrSuperChargerInfoNoAlternative
	}
	if hasSend {
		v := gsm_map.NewSuperChargerInfoSendSubscriberData(struct{}{})
		return &v, nil
	}
	v := gsm_map.NewSuperChargerInfoSubscriberDataStored(gsm_map.AgeIndicator(s.SubscriberDataStored))
	return &v, nil
}

func convertWireToSuperChargerInfo(w *gsm_map.SuperChargerInfo) (*SuperChargerInfo, error) {
	if w.SendSubscriberData != nil && w.SubscriberDataStored != nil {
		return nil, ErrSuperChargerInfoMultipleAlternatives
	}
	out := &SuperChargerInfo{}
	if w.SendSubscriberData != nil {
		out.SendSubscriberData = true
	} else if w.SubscriberDataStored != nil {
		out.SubscriberDataStored = HexBytes(*w.SubscriberDataStored)
	} else {
		return nil, ErrSuperChargerInfoNoAlternative
	}
	return out, nil
}

func convertAddInfoToWire(a *AddInfo) (*gsm_map.ADDInfo, error) {
	imeisvBytes, err := tbcd.Encode(a.IMEISV)
	if err != nil {
		return nil, fmt.Errorf("encoding IMEISV: %w", err)
	}
	out := &gsm_map.ADDInfo{
		Imeisv:                   gsm_map.IMEI(imeisvBytes),
		SkipSubscriberDataUpdate: boolToNullPtr(a.SkipSubscriberDataUpdate),
	}
	return out, nil
}

func convertWireToAddInfo(w *gsm_map.ADDInfo) (*AddInfo, error) {
	imeisv, err := tbcd.Decode(w.Imeisv)
	if err != nil {
		return nil, fmt.Errorf("decoding IMEISV: %w", err)
	}
	return &AddInfo{
		IMEISV:                   imeisv,
		SkipSubscriberDataUpdate: nullPtrToBool(w.SkipSubscriberDataUpdate),
	}, nil
}

// --- SRI nested SEQUENCE helpers ---

func boolToNullPtr(b bool) *struct{} {
	if !b {
		return nil
	}
	v := struct{}{}
	return &v
}

func nullPtrToBool(p *struct{}) bool { return p != nil }

func intPtrTo64(p *int) *int64 {
	if p == nil {
		return nil
	}
	v := int64(*p)
	return &v
}

func int64PtrTo(p *int64) *int {
	if p == nil {
		return nil
	}
	v := int(*p)
	return &v
}

func convertForwardingDataToWire(f *ForwardingData) (*gsm_map.ForwardingData, error) {
	out := &gsm_map.ForwardingData{}
	if f.ForwardedToNumber != "" {
		enc, err := encodeAddressField(f.ForwardedToNumber, f.ForwardedToNumberNature, f.ForwardedToNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding ForwardedToNumber: %w", err)
		}
		as := gsm_map.ISDNAddressString(enc)
		out.ForwardedToNumber = &as
	}
	if len(f.ForwardedToSubaddress) > 0 {
		sa := gsm_map.ISDNSubaddressString(f.ForwardedToSubaddress)
		out.ForwardedToSubaddress = &sa
	}
	if len(f.ForwardingOptions) > 0 {
		fo := gsm_map.ForwardingOptions(f.ForwardingOptions)
		out.ForwardingOptions = &fo
	}
	if len(f.LongForwardedToNumber) > 0 {
		ln := gsm_map.FTNAddressString(f.LongForwardedToNumber)
		out.LongForwardedToNumber = &ln
	}
	return out, nil
}

func convertWireToForwardingData(w *gsm_map.ForwardingData) (*ForwardingData, error) {
	out := &ForwardingData{}
	if w.ForwardedToNumber != nil {
		digits, nat, pl, err := decodeAddressField(*w.ForwardedToNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding ForwardedToNumber: %w", err)
		}
		out.ForwardedToNumber = digits
		out.ForwardedToNumberNature = nat
		out.ForwardedToNumberPlan = pl
	}
	if w.ForwardedToSubaddress != nil {
		out.ForwardedToSubaddress = HexBytes(*w.ForwardedToSubaddress)
	}
	if w.ForwardingOptions != nil {
		out.ForwardingOptions = HexBytes(*w.ForwardingOptions)
	}
	if w.LongForwardedToNumber != nil {
		out.LongForwardedToNumber = HexBytes(*w.LongForwardedToNumber)
	}
	return out, nil
}

func convertCcbsIndicatorsToWire(c *CcbsIndicators) *gsm_map.CCBSIndicators {
	return &gsm_map.CCBSIndicators{
		CcbsPossible:          boolToNullPtr(c.CcbsPossible),
		KeepCCBSCallIndicator: boolToNullPtr(c.KeepCCBSCallIndicator),
	}
}

func convertWireToCcbsIndicators(w *gsm_map.CCBSIndicators) *CcbsIndicators {
	return &CcbsIndicators{
		CcbsPossible:          nullPtrToBool(w.CcbsPossible),
		KeepCCBSCallIndicator: nullPtrToBool(w.KeepCCBSCallIndicator),
	}
}

func convertCugCheckInfoToWire(c *CugCheckInfo) *gsm_map.CUGCheckInfo {
	return &gsm_map.CUGCheckInfo{
		CugInterlock:      gsm_map.CUGInterlock(c.CugInterlock),
		CugOutgoingAccess: boolToNullPtr(c.CugOutgoingAccess),
	}
}

func convertWireToCugCheckInfo(w *gsm_map.CUGCheckInfo) *CugCheckInfo {
	return &CugCheckInfo{
		CugInterlock:      HexBytes(w.CugInterlock),
		CugOutgoingAccess: nullPtrToBool(w.CugOutgoingAccess),
	}
}

// --- SRI CHOICE helpers ---

func convertExtBasicServiceCodeToWire(e *ExtBasicServiceCode) (*gsm_map.ExtBasicServiceCode, error) {
	hasBearer := len(e.ExtBearerService) > 0
	hasTele := len(e.ExtTeleservice) > 0
	switch {
	case hasBearer && hasTele:
		return nil, ErrSriChoiceMultipleAlternatives
	case hasBearer:
		v := gsm_map.NewExtBasicServiceCodeExtBearerService(gsm_map.ExtBearerServiceCode(e.ExtBearerService))
		return &v, nil
	case hasTele:
		v := gsm_map.NewExtBasicServiceCodeExtTeleservice(gsm_map.ExtTeleserviceCode(e.ExtTeleservice))
		return &v, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertWireToExtBasicServiceCode(w *gsm_map.ExtBasicServiceCode) (*ExtBasicServiceCode, error) {
	switch w.Choice {
	case gsm_map.ExtBasicServiceCodeChoiceExtBearerService:
		if w.ExtBearerService == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		return &ExtBasicServiceCode{ExtBearerService: HexBytes(*w.ExtBearerService)}, nil
	case gsm_map.ExtBasicServiceCodeChoiceExtTeleservice:
		if w.ExtTeleservice == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		return &ExtBasicServiceCode{ExtTeleservice: HexBytes(*w.ExtTeleservice)}, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertRoutingInfoToWire(r *RoutingInfo) (*gsm_map.RoutingInfo, error) {
	hasRoaming := r.RoamingNumber != ""
	hasFwd := r.ForwardingData != nil
	switch {
	case hasRoaming && hasFwd:
		return nil, ErrSriChoiceMultipleAlternatives
	case hasRoaming:
		enc, err := encodeAddressField(r.RoamingNumber, r.RoamingNumberNature, r.RoamingNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding RoamingNumber: %w", err)
		}
		v := gsm_map.NewRoutingInfoRoamingNumber(gsm_map.ISDNAddressString(enc))
		return &v, nil
	case hasFwd:
		fw, err := convertForwardingDataToWire(r.ForwardingData)
		if err != nil {
			return nil, err
		}
		v := gsm_map.NewRoutingInfoForwardingData(*fw)
		return &v, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertWireToRoutingInfo(w *gsm_map.RoutingInfo) (*RoutingInfo, error) {
	switch w.Choice {
	case gsm_map.RoutingInfoChoiceRoamingNumber:
		if w.RoamingNumber == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		digits, nat, pl, err := decodeAddressField(*w.RoamingNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding RoamingNumber: %w", err)
		}
		return &RoutingInfo{RoamingNumber: digits, RoamingNumberNature: nat, RoamingNumberPlan: pl}, nil
	case gsm_map.RoutingInfoChoiceForwardingData:
		if w.ForwardingData == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		fd, err := convertWireToForwardingData(w.ForwardingData)
		if err != nil {
			return nil, err
		}
		return &RoutingInfo{ForwardingData: fd}, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertExtendedRoutingInfoToWire(e *ExtendedRoutingInfo) (*gsm_map.ExtendedRoutingInfo, error) {
	hasRI := e.RoutingInfo != nil
	hasCamel := e.CamelRoutingInfo != nil
	switch {
	case hasRI && hasCamel:
		return nil, ErrSriChoiceMultipleAlternatives
	case hasRI:
		ri, err := convertRoutingInfoToWire(e.RoutingInfo)
		if err != nil {
			return nil, err
		}
		v := gsm_map.NewExtendedRoutingInfoRoutingInfo(*ri)
		return &v, nil
	case hasCamel:
		cri, err := convertCamelRoutingInfoToWire(e.CamelRoutingInfo)
		if err != nil {
			return nil, err
		}
		v := gsm_map.NewExtendedRoutingInfoCamelRoutingInfo(*cri)
		return &v, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertWireToExtendedRoutingInfo(w *gsm_map.ExtendedRoutingInfo) (*ExtendedRoutingInfo, error) {
	switch w.Choice {
	case gsm_map.ExtendedRoutingInfoChoiceRoutingInfo:
		if w.RoutingInfo == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		ri, err := convertWireToRoutingInfo(w.RoutingInfo)
		if err != nil {
			return nil, err
		}
		return &ExtendedRoutingInfo{RoutingInfo: ri}, nil
	case gsm_map.ExtendedRoutingInfoChoiceCamelRoutingInfo:
		if w.CamelRoutingInfo == nil {
			return nil, ErrSriChoiceNoAlternative
		}
		cri, err := convertWireToCamelRoutingInfo(w.CamelRoutingInfo)
		if err != nil {
			return nil, err
		}
		return &ExtendedRoutingInfo{CamelRoutingInfo: cri}, nil
	default:
		return nil, ErrSriChoiceNoAlternative
	}
}

func convertCamelRoutingInfoToWire(c *CamelRoutingInfo) (*gsm_map.CamelRoutingInfo, error) {
	out := &gsm_map.CamelRoutingInfo{
		GmscCamelSubscriptionInfo: convertGmscCamelSubInfoToWire(&c.GmscCamelSubscriptionInfo),
	}
	if c.ForwardingData != nil {
		fd, err := convertForwardingDataToWire(c.ForwardingData)
		if err != nil {
			return nil, err
		}
		out.ForwardingData = fd
	}
	return out, nil
}

func convertWireToCamelRoutingInfo(w *gsm_map.CamelRoutingInfo) (*CamelRoutingInfo, error) {
	out := &CamelRoutingInfo{
		GmscCamelSubscriptionInfo: convertWireToGmscCamelSubInfo(&w.GmscCamelSubscriptionInfo),
	}
	if w.ForwardingData != nil {
		fd, err := convertWireToForwardingData(w.ForwardingData)
		if err != nil {
			return nil, err
		}
		out.ForwardingData = fd
	}
	return out, nil
}

// GmscCamelSubscriptionInfo: nested CAMEL SEQUENCEs (T-CSI, O-CSI, D-CSI,
// criteria lists) are not yet decomposed. These stubs silently drop CAMEL
// subscription data on round-trip. Full CAMEL support is deferred to future work.
// TODO: implement field mappings for GmscCamelSubscriptionInfo.
func convertGmscCamelSubInfoToWire(_ *GmscCamelSubscriptionInfo) gsm_map.GmscCamelSubscriptionInfo {
	return gsm_map.GmscCamelSubscriptionInfo{}
}

func convertWireToGmscCamelSubInfo(_ *gsm_map.GmscCamelSubscriptionInfo) GmscCamelSubscriptionInfo {
	return GmscCamelSubscriptionInfo{}
}

// --- SRI remaining helpers ---

func convertExternalSignalInfoToWire(e *ExternalSignalInfo) *gsm_map.ExternalSignalInfo {
	return &gsm_map.ExternalSignalInfo{
		ProtocolId: gsm_map.ProtocolId(int64(e.ProtocolID)),
		SignalInfo:  gsm_map.SignalInfo(e.SignalInfo),
	}
}

func convertWireToExternalSignalInfo(w *gsm_map.ExternalSignalInfo) *ExternalSignalInfo {
	return &ExternalSignalInfo{
		ProtocolID: int(w.ProtocolId),
		SignalInfo:  HexBytes(w.SignalInfo),
	}
}

func convertExtExternalSignalInfoToWire(e *ExtExternalSignalInfo) *gsm_map.ExtExternalSignalInfo {
	return &gsm_map.ExtExternalSignalInfo{
		ExtProtocolId: gsm_map.ExtProtocolId(int64(e.ExtProtocolID)),
		SignalInfo:    gsm_map.SignalInfo(e.SignalInfo),
	}
}

func convertWireToExtExternalSignalInfo(w *gsm_map.ExtExternalSignalInfo) *ExtExternalSignalInfo {
	return &ExtExternalSignalInfo{
		ExtProtocolID: int(w.ExtProtocolId),
		SignalInfo:    HexBytes(w.SignalInfo),
	}
}

func convertSriCamelInfoToWire(c *SriCamelInfo) *gsm_map.CamelInfo {
	out := &gsm_map.CamelInfo{
		SupportedCamelPhases: convertCamelPhasesToBitString(&c.SupportedCamelPhases),
		SuppressTCSI:         boolToNullPtr(c.SuppressTCSI),
	}
	if c.OfferedCamel4CSIs != nil {
		bs := convertOfferedCamel4CSIsToBitString(c.OfferedCamel4CSIs)
		out.OfferedCamel4CSIs = &bs
	}
	return out
}

func convertWireToSriCamelInfo(w *gsm_map.CamelInfo) *SriCamelInfo {
	out := &SriCamelInfo{
		SupportedCamelPhases: *convertBitStringToCamelPhases(w.SupportedCamelPhases),
		SuppressTCSI:         nullPtrToBool(w.SuppressTCSI),
	}
	if w.OfferedCamel4CSIs != nil {
		out.OfferedCamel4CSIs = convertBitStringToOfferedCamel4CSIs(*w.OfferedCamel4CSIs)
	}
	return out
}

func convertNaeaPreferredCIToWire(n *NaeaPreferredCI) *gsm_map.NAEAPreferredCI {
	return &gsm_map.NAEAPreferredCI{NaeaPreferredCIC: gsm_map.NAEACIC(n.NaeaPreferredCIC)}
}

func convertWireToNaeaPreferredCI(w *gsm_map.NAEAPreferredCI) *NaeaPreferredCI {
	return &NaeaPreferredCI{NaeaPreferredCIC: HexBytes(w.NaeaPreferredCIC)}
}

// --- SRI (SendRoutingInfo) full converters ---

func validateSri(s *Sri) error {
	if s.MSISDN == "" {
		return ErrSriMissingMSISDN
	}
	if s.GmscOrGsmSCFAddress == "" {
		return ErrSriMissingGmsc
	}
	if s.InterrogationType != InterrogationBasicCall && s.InterrogationType != InterrogationForwarding {
		return ErrSriInvalidInterrogationType
	}
	if s.NumberOfForwarding != nil {
		if *s.NumberOfForwarding < 1 || *s.NumberOfForwarding > 5 {
			return ErrSriInvalidNumberOfForwarding
		}
	}
	if s.OrCapability != nil {
		if *s.OrCapability < 1 || *s.OrCapability > 127 {
			return ErrSriInvalidOrCapability
		}
	}
	if len(s.CallReferenceNumber) > 8 {
		return ErrSriInvalidCallReferenceNumber
	}
	return nil
}

func convertSriToArg(s *Sri) (*gsm_map.SendRoutingInfoArg, error) {
	if err := validateSri(s); err != nil {
		return nil, err
	}

	msisdn, err := encodeAddressField(s.MSISDN, s.MSISDNNature, s.MSISDNPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding MSISDN: %w", err)
	}
	gmsc, err := encodeAddressField(s.GmscOrGsmSCFAddress, s.GmscNature, s.GmscPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding GmscOrGsmSCFAddress: %w", err)
	}

	arg := &gsm_map.SendRoutingInfoArg{
		Msisdn:              gsm_map.ISDNAddressString(msisdn),
		InterrogationType:   gsm_map.InterrogationType(int64(s.InterrogationType)),
		GmscOrGsmSCFAddress: gsm_map.ISDNAddressString(gmsc),
	}

	// CugCheckInfo
	if s.CugCheckInfo != nil {
		arg.CugCheckInfo = convertCugCheckInfoToWire(s.CugCheckInfo)
	}

	// NumberOfForwarding
	if s.NumberOfForwarding != nil {
		v := int64(*s.NumberOfForwarding)
		arg.NumberOfForwarding = &v
	}

	// OrInterrogation
	arg.OrInterrogation = boolToNullPtr(s.OrInterrogation)

	// OrCapability
	if s.OrCapability != nil {
		v := int64(*s.OrCapability)
		arg.OrCapability = &v
	}

	// CallReferenceNumber
	if len(s.CallReferenceNumber) > 0 {
		v := gsm_map.CallReferenceNumber(s.CallReferenceNumber)
		arg.CallReferenceNumber = &v
	}

	// ForwardingReason
	if s.ForwardingReason != nil {
		v := gsm_map.ForwardingReason(int64(*s.ForwardingReason))
		arg.ForwardingReason = &v
	}

	// BasicServiceGroup
	if s.BasicServiceGroup != nil {
		bsg, err := convertExtBasicServiceCodeToWire(s.BasicServiceGroup)
		if err != nil {
			return nil, fmt.Errorf("encoding BasicServiceGroup: %w", err)
		}
		arg.BasicServiceGroup = bsg
	}

	// BasicServiceGroup2
	if s.BasicServiceGroup2 != nil {
		bsg2, err := convertExtBasicServiceCodeToWire(s.BasicServiceGroup2)
		if err != nil {
			return nil, fmt.Errorf("encoding BasicServiceGroup2: %w", err)
		}
		arg.BasicServiceGroup2 = bsg2
	}

	// NetworkSignalInfo
	if s.NetworkSignalInfo != nil {
		arg.NetworkSignalInfo = convertExternalSignalInfoToWire(s.NetworkSignalInfo)
	}

	// NetworkSignalInfo2
	if s.NetworkSignalInfo2 != nil {
		arg.NetworkSignalInfo2 = convertExternalSignalInfoToWire(s.NetworkSignalInfo2)
	}

	// CamelInfo
	if s.CamelInfo != nil {
		arg.CamelInfo = convertSriCamelInfoToWire(s.CamelInfo)
	}

	// SuppressionOfAnnouncement
	if s.SuppressionOfAnnouncement {
		v := gsm_map.SuppressionOfAnnouncement{}
		arg.SuppressionOfAnnouncement = &v
	}

	// AlertingPattern
	if len(s.AlertingPattern) > 0 {
		v := gsm_map.AlertingPattern(s.AlertingPattern)
		arg.AlertingPattern = &v
	}

	// CcbsCall
	arg.CcbsCall = boolToNullPtr(s.CcbsCall)

	// SupportedCCBSPhase
	if s.SupportedCCBSPhase != nil {
		v := int64(*s.SupportedCCBSPhase)
		arg.SupportedCCBSPhase = &v
	}

	// AdditionalSignalInfo
	if s.AdditionalSignalInfo != nil {
		arg.AdditionalSignalInfo = convertExtExternalSignalInfoToWire(s.AdditionalSignalInfo)
	}

	// IstSupportIndicator
	if s.IstSupportIndicator != nil {
		v := gsm_map.ISTSupportIndicator(int64(*s.IstSupportIndicator))
		arg.IstSupportIndicator = &v
	}

	// PrePagingSupported
	arg.PrePagingSupported = boolToNullPtr(s.PrePagingSupported)

	// CallDiversionTreatmentIndicator
	if len(s.CallDiversionTreatmentIndicator) > 0 {
		v := gsm_map.CallDiversionTreatmentIndicator(s.CallDiversionTreatmentIndicator)
		arg.CallDiversionTreatmentIndicator = &v
	}

	// LongFTNSupported
	arg.LongFTNSupported = boolToNullPtr(s.LongFTNSupported)

	// SuppressVTCSI
	arg.SuppressVTCSI = boolToNullPtr(s.SuppressVTCSI)

	// SuppressIncomingCallBarring
	arg.SuppressIncomingCallBarring = boolToNullPtr(s.SuppressIncomingCallBarring)

	// GsmSCFInitiatedCall
	arg.GsmSCFInitiatedCall = boolToNullPtr(s.GsmSCFInitiatedCall)

	// SuppressMTSS
	if s.SuppressMTSS != nil {
		v := convertSuppressMTSSToBitString(s.SuppressMTSS)
		arg.SuppressMTSS = &v
	}

	// MtRoamingRetrySupported
	arg.MtRoamingRetrySupported = boolToNullPtr(s.MtRoamingRetrySupported)

	// CallPriority
	if s.CallPriority != nil {
		v := int64(*s.CallPriority)
		arg.CallPriority = &v
	}

	return arg, nil
}

func convertArgToSri(arg *gsm_map.SendRoutingInfoArg) (*Sri, error) {
	msisdn, msisdnNature, msisdnPlan, err := decodeAddressField(arg.Msisdn)
	if err != nil {
		return nil, fmt.Errorf("decoding MSISDN: %w", err)
	}

	gmsc, gmscNature, gmscPlan, err := decodeAddressField(arg.GmscOrGsmSCFAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding GmscOrGsmSCFAddress: %w", err)
	}

	s := &Sri{
		MSISDN:              msisdn,
		MSISDNNature:        msisdnNature,
		MSISDNPlan:          msisdnPlan,
		InterrogationType:   InterrogationType(int(arg.InterrogationType)),
		GmscOrGsmSCFAddress: gmsc,
		GmscNature:          gmscNature,
		GmscPlan:            gmscPlan,
	}

	// CugCheckInfo
	if arg.CugCheckInfo != nil {
		s.CugCheckInfo = convertWireToCugCheckInfo(arg.CugCheckInfo)
	}

	// NumberOfForwarding
	if arg.NumberOfForwarding != nil {
		v := int(*arg.NumberOfForwarding)
		s.NumberOfForwarding = &v
	}

	// OrInterrogation
	s.OrInterrogation = nullPtrToBool(arg.OrInterrogation)

	// OrCapability
	if arg.OrCapability != nil {
		v := int(*arg.OrCapability)
		s.OrCapability = &v
	}

	// CallReferenceNumber
	if arg.CallReferenceNumber != nil {
		s.CallReferenceNumber = HexBytes(*arg.CallReferenceNumber)
	}

	// ForwardingReason
	if arg.ForwardingReason != nil {
		v := ForwardingReason(int(*arg.ForwardingReason))
		s.ForwardingReason = &v
	}

	// BasicServiceGroup
	if arg.BasicServiceGroup != nil {
		bsg, err := convertWireToExtBasicServiceCode(arg.BasicServiceGroup)
		if err != nil {
			return nil, fmt.Errorf("decoding BasicServiceGroup: %w", err)
		}
		s.BasicServiceGroup = bsg
	}

	// BasicServiceGroup2
	if arg.BasicServiceGroup2 != nil {
		bsg2, err := convertWireToExtBasicServiceCode(arg.BasicServiceGroup2)
		if err != nil {
			return nil, fmt.Errorf("decoding BasicServiceGroup2: %w", err)
		}
		s.BasicServiceGroup2 = bsg2
	}

	// NetworkSignalInfo
	if arg.NetworkSignalInfo != nil {
		s.NetworkSignalInfo = convertWireToExternalSignalInfo(arg.NetworkSignalInfo)
	}

	// NetworkSignalInfo2
	if arg.NetworkSignalInfo2 != nil {
		s.NetworkSignalInfo2 = convertWireToExternalSignalInfo(arg.NetworkSignalInfo2)
	}

	// CamelInfo
	if arg.CamelInfo != nil {
		s.CamelInfo = convertWireToSriCamelInfo(arg.CamelInfo)
	}

	// SuppressionOfAnnouncement
	s.SuppressionOfAnnouncement = arg.SuppressionOfAnnouncement != nil

	// AlertingPattern
	if arg.AlertingPattern != nil {
		s.AlertingPattern = HexBytes(*arg.AlertingPattern)
	}

	// CcbsCall
	s.CcbsCall = nullPtrToBool(arg.CcbsCall)

	// SupportedCCBSPhase
	if arg.SupportedCCBSPhase != nil {
		v := int(*arg.SupportedCCBSPhase)
		s.SupportedCCBSPhase = &v
	}

	// AdditionalSignalInfo
	if arg.AdditionalSignalInfo != nil {
		s.AdditionalSignalInfo = convertWireToExtExternalSignalInfo(arg.AdditionalSignalInfo)
	}

	// IstSupportIndicator
	if arg.IstSupportIndicator != nil {
		v := int(*arg.IstSupportIndicator)
		s.IstSupportIndicator = &v
	}

	// PrePagingSupported
	s.PrePagingSupported = nullPtrToBool(arg.PrePagingSupported)

	// CallDiversionTreatmentIndicator
	if arg.CallDiversionTreatmentIndicator != nil {
		s.CallDiversionTreatmentIndicator = HexBytes(*arg.CallDiversionTreatmentIndicator)
	}

	// LongFTNSupported
	s.LongFTNSupported = nullPtrToBool(arg.LongFTNSupported)

	// SuppressVTCSI
	s.SuppressVTCSI = nullPtrToBool(arg.SuppressVTCSI)

	// SuppressIncomingCallBarring
	s.SuppressIncomingCallBarring = nullPtrToBool(arg.SuppressIncomingCallBarring)

	// GsmSCFInitiatedCall
	s.GsmSCFInitiatedCall = nullPtrToBool(arg.GsmSCFInitiatedCall)

	// SuppressMTSS
	if arg.SuppressMTSS != nil && arg.SuppressMTSS.BitLength > 0 {
		s.SuppressMTSS = convertBitStringToSuppressMTSS(*arg.SuppressMTSS)
	}

	// MtRoamingRetrySupported
	s.MtRoamingRetrySupported = nullPtrToBool(arg.MtRoamingRetrySupported)

	// CallPriority
	if arg.CallPriority != nil {
		v := int(*arg.CallPriority)
		s.CallPriority = &v
	}

	return s, nil
}

// --- SRI Response (SendRoutingInfoRes) full converters ---

func convertSriRespToRes(s *SriResp) (*gsm_map.SendRoutingInfoRes, error) {
	imsiBytes, err := tbcd.Encode(s.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	out := &gsm_map.SendRoutingInfoRes{
		Imsi: (*gsm_map.IMSI)(&imsiBytes),
	}

	// ExtendedRoutingInfo
	if s.ExtendedRoutingInfo != nil {
		eri, err := convertExtendedRoutingInfoToWire(s.ExtendedRoutingInfo)
		if err != nil {
			return nil, fmt.Errorf("encoding ExtendedRoutingInfo: %w", err)
		}
		out.ExtendedRoutingInfo = eri
	}

	// CugCheckInfo
	if s.CugCheckInfo != nil {
		out.CugCheckInfo = convertCugCheckInfoToWire(s.CugCheckInfo)
	}

	// CugSubscriptionFlag
	out.CugSubscriptionFlag = boolToNullPtr(s.CugSubscriptionFlag)

	// SubscriberInfo
	if s.SubscriberInfo != nil {
		si, err := convertSubscriberInfoToWire(s.SubscriberInfo)
		if err != nil {
			return nil, fmt.Errorf("encoding SubscriberInfo: %w", err)
		}
		out.SubscriberInfo = si
	}

	// SsList
	if len(s.SsList) > 0 {
		out.SsList = make(gsm_map.SSList, len(s.SsList))
		for i, c := range s.SsList {
			out.SsList[i] = gsm_map.SSCode{byte(c)}
		}
	}

	// BasicService
	if s.BasicService != nil {
		bs, err := convertExtBasicServiceCodeToWire(s.BasicService)
		if err != nil {
			return nil, fmt.Errorf("encoding BasicService: %w", err)
		}
		out.BasicService = bs
	}

	// ForwardingInterrogationRequired
	out.ForwardingInterrogationRequired = boolToNullPtr(s.ForwardingInterrogationRequired)

	// VmscAddress
	if s.VmscAddress != "" {
		enc, err := encodeAddressField(s.VmscAddress, s.VmscNature, s.VmscPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding VmscAddress: %w", err)
		}
		v := gsm_map.ISDNAddressString(enc)
		out.VmscAddress = &v
	}

	// NaeaPreferredCI
	if s.NaeaPreferredCI != nil {
		out.NaeaPreferredCI = convertNaeaPreferredCIToWire(s.NaeaPreferredCI)
	}

	// CcbsIndicators
	if s.CcbsIndicators != nil {
		out.CcbsIndicators = convertCcbsIndicatorsToWire(s.CcbsIndicators)
	}

	// Msisdn
	if s.MSISDN != "" {
		enc, err := encodeAddressField(s.MSISDN, s.MSISDNNature, s.MSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding MSISDN: %w", err)
		}
		v := gsm_map.ISDNAddressString(enc)
		out.Msisdn = &v
	}

	// NumberPortabilityStatus
	if s.NumberPortabilityStatus != nil {
		v := gsm_map.NumberPortabilityStatus(int64(*s.NumberPortabilityStatus))
		out.NumberPortabilityStatus = &v
	}

	// IstAlertTimer
	out.IstAlertTimer = intPtrTo64(s.IstAlertTimer)

	// SupportedCamelPhasesInVMSC
	if s.SupportedCamelPhasesInVMSC != nil {
		bs := convertCamelPhasesToBitString(s.SupportedCamelPhasesInVMSC)
		out.SupportedCamelPhasesInVMSC = &bs
	}

	// OfferedCamel4CSIsInVMSC
	if s.OfferedCamel4CSIsInVMSC != nil {
		bs := convertOfferedCamel4CSIsToBitString(s.OfferedCamel4CSIsInVMSC)
		out.OfferedCamel4CSIsInVMSC = &bs
	}

	// RoutingInfo2
	if s.RoutingInfo2 != nil {
		ri, err := convertRoutingInfoToWire(s.RoutingInfo2)
		if err != nil {
			return nil, fmt.Errorf("encoding RoutingInfo2: %w", err)
		}
		out.RoutingInfo2 = ri
	}

	// SsList2
	if len(s.SsList2) > 0 {
		out.SsList2 = make(gsm_map.SSList, len(s.SsList2))
		for i, c := range s.SsList2 {
			out.SsList2[i] = gsm_map.SSCode{byte(c)}
		}
	}

	// BasicService2
	if s.BasicService2 != nil {
		bs2, err := convertExtBasicServiceCodeToWire(s.BasicService2)
		if err != nil {
			return nil, fmt.Errorf("encoding BasicService2: %w", err)
		}
		out.BasicService2 = bs2
	}

	// AllowedServices
	if s.AllowedServices != nil {
		bs := convertAllowedServicesToBitString(s.AllowedServices)
		out.AllowedServices = &bs
	}

	// UnavailabilityCause
	if s.UnavailabilityCause != nil {
		v := gsm_map.UnavailabilityCause(int64(*s.UnavailabilityCause))
		out.UnavailabilityCause = &v
	}

	// ReleaseResourcesSupported
	out.ReleaseResourcesSupported = boolToNullPtr(s.ReleaseResourcesSupported)

	// GsmBearerCapability
	if s.GsmBearerCapability != nil {
		out.GsmBearerCapability = convertExternalSignalInfoToWire(s.GsmBearerCapability)
	}

	return out, nil
}

func convertResToSriResp(res *gsm_map.SendRoutingInfoRes) (*SriResp, error) {
	out := &SriResp{}

	// Imsi
	if res.Imsi != nil {
		imsi, err := tbcd.Decode(*res.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding IMSI: %w", err)
		}
		out.IMSI = imsi
	}

	// ExtendedRoutingInfo
	if res.ExtendedRoutingInfo != nil {
		eri, err := convertWireToExtendedRoutingInfo(res.ExtendedRoutingInfo)
		if err != nil {
			return nil, fmt.Errorf("decoding ExtendedRoutingInfo: %w", err)
		}
		out.ExtendedRoutingInfo = eri
	}

	// CugCheckInfo
	if res.CugCheckInfo != nil {
		out.CugCheckInfo = convertWireToCugCheckInfo(res.CugCheckInfo)
	}

	// CugSubscriptionFlag
	out.CugSubscriptionFlag = nullPtrToBool(res.CugSubscriptionFlag)

	// SubscriberInfo
	if res.SubscriberInfo != nil {
		si, err := convertWireToSubscriberInfo(res.SubscriberInfo)
		if err != nil {
			return nil, fmt.Errorf("decoding SubscriberInfo: %w", err)
		}
		out.SubscriberInfo = si
	}

	// SsList
	if len(res.SsList) > 0 {
		out.SsList = make([]SsCode, len(res.SsList))
		for i, c := range res.SsList {
			if len(c) > 0 {
				out.SsList[i] = SsCode(c[0])
			}
		}
	}

	// BasicService
	if res.BasicService != nil {
		bs, err := convertWireToExtBasicServiceCode(res.BasicService)
		if err != nil {
			return nil, fmt.Errorf("decoding BasicService: %w", err)
		}
		out.BasicService = bs
	}

	// ForwardingInterrogationRequired
	out.ForwardingInterrogationRequired = nullPtrToBool(res.ForwardingInterrogationRequired)

	// VmscAddress
	if res.VmscAddress != nil {
		digits, nat, pl, err := decodeAddressField(*res.VmscAddress)
		if err != nil {
			return nil, fmt.Errorf("decoding VmscAddress: %w", err)
		}
		out.VmscAddress = digits
		out.VmscNature = nat
		out.VmscPlan = pl
	}

	// NaeaPreferredCI
	if res.NaeaPreferredCI != nil {
		out.NaeaPreferredCI = convertWireToNaeaPreferredCI(res.NaeaPreferredCI)
	}

	// CcbsIndicators
	if res.CcbsIndicators != nil {
		out.CcbsIndicators = convertWireToCcbsIndicators(res.CcbsIndicators)
	}

	// Msisdn
	if res.Msisdn != nil {
		digits, nat, pl, err := decodeAddressField(*res.Msisdn)
		if err != nil {
			return nil, fmt.Errorf("decoding MSISDN: %w", err)
		}
		out.MSISDN = digits
		out.MSISDNNature = nat
		out.MSISDNPlan = pl
	}

	// NumberPortabilityStatus
	if res.NumberPortabilityStatus != nil {
		v := NumberPortabilityStatus(int(*res.NumberPortabilityStatus))
		out.NumberPortabilityStatus = &v
	}

	// IstAlertTimer
	out.IstAlertTimer = int64PtrTo(res.IstAlertTimer)

	// SupportedCamelPhasesInVMSC
	if res.SupportedCamelPhasesInVMSC != nil && res.SupportedCamelPhasesInVMSC.BitLength > 0 {
		out.SupportedCamelPhasesInVMSC = convertBitStringToCamelPhases(*res.SupportedCamelPhasesInVMSC)
	}

	// OfferedCamel4CSIsInVMSC
	if res.OfferedCamel4CSIsInVMSC != nil && res.OfferedCamel4CSIsInVMSC.BitLength > 0 {
		out.OfferedCamel4CSIsInVMSC = convertBitStringToOfferedCamel4CSIs(*res.OfferedCamel4CSIsInVMSC)
	}

	// RoutingInfo2
	if res.RoutingInfo2 != nil {
		ri, err := convertWireToRoutingInfo(res.RoutingInfo2)
		if err != nil {
			return nil, fmt.Errorf("decoding RoutingInfo2: %w", err)
		}
		out.RoutingInfo2 = ri
	}

	// SsList2
	if len(res.SsList2) > 0 {
		out.SsList2 = make([]SsCode, len(res.SsList2))
		for i, c := range res.SsList2 {
			if len(c) > 0 {
				out.SsList2[i] = SsCode(c[0])
			}
		}
	}

	// BasicService2
	if res.BasicService2 != nil {
		bs2, err := convertWireToExtBasicServiceCode(res.BasicService2)
		if err != nil {
			return nil, fmt.Errorf("decoding BasicService2: %w", err)
		}
		out.BasicService2 = bs2
	}

	// AllowedServices
	if res.AllowedServices != nil && res.AllowedServices.BitLength > 0 {
		out.AllowedServices = convertBitStringToAllowedServices(*res.AllowedServices)
	}

	// UnavailabilityCause
	if res.UnavailabilityCause != nil {
		v := UnavailabilityCause(int(*res.UnavailabilityCause))
		out.UnavailabilityCause = &v
	}

	// ReleaseResourcesSupported
	out.ReleaseResourcesSupported = nullPtrToBool(res.ReleaseResourcesSupported)

	// GsmBearerCapability
	if res.GsmBearerCapability != nil {
		out.GsmBearerCapability = convertWireToExternalSignalInfo(res.GsmBearerCapability)
	}

	return out, nil
}

// --- SubscriberInfo helpers (shared between ATI and SRI response) ---

func convertSubscriberInfoToWire(s *SubscriberInfo) (*gsm_map.SubscriberInfo, error) {
	si := &gsm_map.SubscriberInfo{}

	if s.LocationInformation != nil {
		locInfo, err := convertCSLocationToAsn1(s.LocationInformation)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformation: %w", err)
		}
		si.LocationInformation = locInfo
	}

	if s.SubscriberState != nil {
		si.SubscriberState = convertSubscriberStateToAsn1(s.SubscriberState)
	}

	if s.LocationInformationGPRS != nil {
		locGPRS, err := convertGPRSLocationToAsn1(s.LocationInformationGPRS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationGPRS: %w", err)
		}
		si.LocationInformationGPRS = locGPRS
	}

	if s.PsSubscriberState != nil {
		ps, err := convertPsSubscriberStateToWire(s.PsSubscriberState)
		if err != nil {
			return nil, fmt.Errorf("converting PsSubscriberState: %w", err)
		}
		si.PsSubscriberState = ps
	}

	if s.IMEI != "" {
		imeiBytes, err := tbcd.Encode(s.IMEI)
		if err != nil {
			return nil, fmt.Errorf("encoding IMEI: %w", err)
		}
		imei := gsm_map.IMEI(imeiBytes)
		si.Imei = &imei
	}

	if s.MsClassmark2 != nil {
		mc := gsm_map.MSClassmark2(s.MsClassmark2)
		si.MsClassmark2 = &mc
	}

	if s.GprsMSClass != nil {
		si.GprsMSClass = convertGprsMSClassToWire(s.GprsMSClass)
	}

	if s.MnpInfoRes != nil {
		mnp, err := convertMnpInfoResToWire(s.MnpInfoRes)
		if err != nil {
			return nil, fmt.Errorf("converting MnpInfoRes: %w", err)
		}
		si.MnpInfoRes = mnp
	}

	if s.ImsVoiceOverPSSessionsIndication != nil {
		v := gsm_map.IMSVoiceOverPSSessionsInd(int64(*s.ImsVoiceOverPSSessionsIndication))
		si.ImsVoiceOverPSSessionsIndication = &v
	}

	if s.LastUEActivityTime != nil {
		t := gsm_map.Time(s.LastUEActivityTime)
		si.LastUEActivityTime = &t
	}

	if s.LastRATType != nil {
		v := gsm_map.UsedRATType(int64(*s.LastRATType))
		si.LastRATType = &v
	}

	if s.EpsSubscriberState != nil {
		ps, err := convertPsSubscriberStateToWire(s.EpsSubscriberState)
		if err != nil {
			return nil, fmt.Errorf("converting EpsSubscriberState: %w", err)
		}
		si.EpsSubscriberState = ps
	}

	if s.LocationInformationEPS != nil {
		locEPS, err := convertEPSLocationToAsn1(s.LocationInformationEPS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationEPS: %w", err)
		}
		si.LocationInformationEPS = locEPS
	}

	if s.TimeZone != nil {
		tz := gsm_map.TimeZone(s.TimeZone)
		si.TimeZone = &tz
	}

	if s.DaylightSavingTime != nil {
		dst := gsm_map.DaylightSavingTime(*s.DaylightSavingTime)
		si.DaylightSavingTime = &dst
	}

	if s.LocationInformation5GS != nil {
		loc5gs, err := convertLocationInformation5GSToWire(s.LocationInformation5GS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformation5GS: %w", err)
		}
		si.LocationInformation5GS = loc5gs
	}

	return si, nil
}

func convertWireToSubscriberInfo(si *gsm_map.SubscriberInfo) (*SubscriberInfo, error) {
	out := &SubscriberInfo{}

	if si.LocationInformation != nil {
		locInfo, err := convertAsn1ToCSLocation(si.LocationInformation)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformation: %w", err)
		}
		out.LocationInformation = locInfo
	}

	if si.SubscriberState != nil {
		out.SubscriberState = convertAsn1ToSubscriberState(si.SubscriberState)
	}

	if si.LocationInformationGPRS != nil {
		locGPRS, err := convertAsn1ToGPRSLocation(si.LocationInformationGPRS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationGPRS: %w", err)
		}
		out.LocationInformationGPRS = locGPRS
	}

	if si.PsSubscriberState != nil {
		ps, err := convertWireToPsSubscriberState(si.PsSubscriberState)
		if err != nil {
			return nil, fmt.Errorf("converting PsSubscriberState: %w", err)
		}
		out.PsSubscriberState = ps
	}

	if si.Imei != nil && len(*si.Imei) > 0 {
		imei, err := tbcd.Decode(*si.Imei)
		if err != nil {
			return nil, fmt.Errorf("decoding IMEI: %w", err)
		}
		out.IMEI = imei
	}

	if si.MsClassmark2 != nil {
		out.MsClassmark2 = []byte(*si.MsClassmark2)
	}

	if si.GprsMSClass != nil {
		out.GprsMSClass = convertWireToGprsMSClass(si.GprsMSClass)
	}

	if si.MnpInfoRes != nil {
		mnp, err := convertWireToMnpInfoRes(si.MnpInfoRes)
		if err != nil {
			return nil, fmt.Errorf("converting MnpInfoRes: %w", err)
		}
		out.MnpInfoRes = mnp
	}

	if si.ImsVoiceOverPSSessionsIndication != nil {
		v := ImsVoiceOverPSSessionsIndication(int(*si.ImsVoiceOverPSSessionsIndication))
		out.ImsVoiceOverPSSessionsIndication = &v
	}

	if si.LastUEActivityTime != nil {
		out.LastUEActivityTime = []byte(*si.LastUEActivityTime)
	}

	if si.LastRATType != nil {
		v := UsedRatType(int(*si.LastRATType))
		out.LastRATType = &v
	}

	if si.EpsSubscriberState != nil {
		ps, err := convertWireToPsSubscriberState(si.EpsSubscriberState)
		if err != nil {
			return nil, fmt.Errorf("converting EpsSubscriberState: %w", err)
		}
		out.EpsSubscriberState = ps
	}

	if si.LocationInformationEPS != nil {
		locEPS, err := convertAsn1ToEPSLocation(si.LocationInformationEPS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationEPS: %w", err)
		}
		out.LocationInformationEPS = locEPS
	}

	if si.TimeZone != nil {
		out.TimeZone = []byte(*si.TimeZone)
	}

	if si.DaylightSavingTime != nil {
		v := int(*si.DaylightSavingTime)
		out.DaylightSavingTime = &v
	}

	if si.LocationInformation5GS != nil {
		loc5gs, err := convertWireToLocationInformation5GS(si.LocationInformation5GS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformation5GS: %w", err)
		}
		out.LocationInformation5GS = loc5gs
	}

	return out, nil
}

// --- PS-SubscriberState (opCode 71) ---

// psSubscriberStateCount returns the number of alternatives set in p.
func psSubscriberStateCount(p *PsSubscriberState) int {
	c := 0
	if p.NotProvidedFromSGSNorMME {
		c++
	}
	if p.PsDetached {
		c++
	}
	if p.PsAttachedNotReachableForPaging {
		c++
	}
	if p.PsAttachedReachableForPaging {
		c++
	}
	if len(p.PsPDPActiveNotReachableForPaging) > 0 {
		c++
	}
	if len(p.PsPDPActiveReachableForPaging) > 0 {
		c++
	}
	if p.NetDetNotReachable != nil {
		c++
	}
	return c
}

func convertPsSubscriberStateToWire(p *PsSubscriberState) (*gsm_map.PSSubscriberState, error) {
	n := psSubscriberStateCount(p)
	if n == 0 {
		return nil, ErrAtiPsSubscriberStateNoAlternative
	}
	if n > 1 {
		return nil, ErrAtiPsSubscriberStateMultipleAlternatives
	}

	switch {
	case p.NotProvidedFromSGSNorMME:
		v := gsm_map.NewPSSubscriberStateNotProvidedFromSGSNorMME(struct{}{})
		return &v, nil
	case p.PsDetached:
		v := gsm_map.NewPSSubscriberStatePsDetached(struct{}{})
		return &v, nil
	case p.PsAttachedNotReachableForPaging:
		v := gsm_map.NewPSSubscriberStatePsAttachedNotReachableForPaging(struct{}{})
		return &v, nil
	case p.PsAttachedReachableForPaging:
		v := gsm_map.NewPSSubscriberStatePsAttachedReachableForPaging(struct{}{})
		return &v, nil
	case len(p.PsPDPActiveNotReachableForPaging) > 0:
		list, err := decodePDPContextInfoList(p.PsPDPActiveNotReachableForPaging)
		if err != nil {
			return nil, fmt.Errorf("decoding PsPDPActiveNotReachableForPaging: %w", err)
		}
		v := gsm_map.NewPSSubscriberStatePsPDPActiveNotReachableForPaging(list)
		return &v, nil
	case len(p.PsPDPActiveReachableForPaging) > 0:
		list, err := decodePDPContextInfoList(p.PsPDPActiveReachableForPaging)
		if err != nil {
			return nil, fmt.Errorf("decoding PsPDPActiveReachableForPaging: %w", err)
		}
		v := gsm_map.NewPSSubscriberStatePsPDPActiveReachableForPaging(list)
		return &v, nil
	case p.NetDetNotReachable != nil:
		v := gsm_map.NewPSSubscriberStateNetDetNotReachable(gsm_map.NotReachableReason(int64(*p.NetDetNotReachable)))
		return &v, nil
	}
	return nil, ErrAtiPsSubscriberStateNoAlternative
}

func convertWireToPsSubscriberState(w *gsm_map.PSSubscriberState) (*PsSubscriberState, error) {
	out := &PsSubscriberState{}
	switch w.Choice {
	case gsm_map.PSSubscriberStateChoiceNotProvidedFromSGSNorMME:
		out.NotProvidedFromSGSNorMME = true
	case gsm_map.PSSubscriberStateChoicePsDetached:
		out.PsDetached = true
	case gsm_map.PSSubscriberStateChoicePsAttachedNotReachableForPaging:
		out.PsAttachedNotReachableForPaging = true
	case gsm_map.PSSubscriberStateChoicePsAttachedReachableForPaging:
		out.PsAttachedReachableForPaging = true
	case gsm_map.PSSubscriberStateChoicePsPDPActiveNotReachableForPaging:
		enc, err := encodePDPContextInfoList(w.PsPDPActiveNotReachableForPaging)
		if err != nil {
			return nil, fmt.Errorf("encoding PsPDPActiveNotReachableForPaging: %w", err)
		}
		out.PsPDPActiveNotReachableForPaging = enc
	case gsm_map.PSSubscriberStateChoicePsPDPActiveReachableForPaging:
		enc, err := encodePDPContextInfoList(w.PsPDPActiveReachableForPaging)
		if err != nil {
			return nil, fmt.Errorf("encoding PsPDPActiveReachableForPaging: %w", err)
		}
		out.PsPDPActiveReachableForPaging = enc
	case gsm_map.PSSubscriberStateChoiceNetDetNotReachable:
		if w.NetDetNotReachable == nil {
			return nil, fmt.Errorf("PsSubscriberState: NetDetNotReachable alternative selected but reason is nil")
		}
		v := int(*w.NetDetNotReachable)
		out.NetDetNotReachable = &v
	default:
		return nil, fmt.Errorf("PsSubscriberState: unknown CHOICE value %d", w.Choice)
	}
	return out, nil
}

// encodePDPContextInfoList serializes each gsm_map.PDPContextInfo entry to
// its BER-encoded bytes, keeping them opaque from the caller's perspective.
func encodePDPContextInfoList(list gsm_map.PDPContextInfoList) ([]HexBytes, error) {
	if len(list) == 0 {
		return nil, nil
	}
	out := make([]HexBytes, len(list))
	for i := range list {
		ctx := list[i]
		enc, err := ctx.MarshalBER()
		if err != nil {
			return nil, fmt.Errorf("PDPContextInfo[%d]: %w", i, err)
		}
		out[i] = enc
	}
	return out, nil
}

// decodePDPContextInfoList deserializes each opaque PDPContextInfo entry
// back into its gsm_map.PDPContextInfo struct.
func decodePDPContextInfoList(list []HexBytes) (gsm_map.PDPContextInfoList, error) {
	if len(list) == 0 {
		return nil, nil
	}
	out := make(gsm_map.PDPContextInfoList, len(list))
	for i, b := range list {
		var ctx gsm_map.PDPContextInfo
		if err := ctx.UnmarshalBER(b); err != nil {
			return nil, fmt.Errorf("PDPContextInfo[%d]: %w", i, err)
		}
		out[i] = ctx
	}
	return out, nil
}

// --- MNPInfoRes (opCode 71) ---

func convertMnpInfoResToWire(m *MnpInfoRes) (*gsm_map.MNPInfoRes, error) {
	out := &gsm_map.MNPInfoRes{}

	if m.RouteingNumber != nil {
		rn := gsm_map.RouteingNumber(m.RouteingNumber)
		out.RouteingNumber = &rn
	}

	if m.IMSI != "" {
		b, err := tbcd.Encode(m.IMSI)
		if err != nil {
			return nil, fmt.Errorf(errEncodingIMSI, err)
		}
		imsi := gsm_map.IMSI(b)
		out.Imsi = &imsi
	}

	if m.MSISDN != "" {
		enc, err := encodeAddressField(m.MSISDN, m.MSISDNNature, m.MSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding MSISDN: %w", err)
		}
		as := gsm_map.ISDNAddressString(enc)
		out.Msisdn = &as
	}

	if m.NumberPortabilityStatus != nil {
		v := gsm_map.NumberPortabilityStatus(int64(*m.NumberPortabilityStatus))
		out.NumberPortabilityStatus = &v
	}

	return out, nil
}

func convertWireToMnpInfoRes(w *gsm_map.MNPInfoRes) (*MnpInfoRes, error) {
	out := &MnpInfoRes{}

	if w.RouteingNumber != nil {
		out.RouteingNumber = []byte(*w.RouteingNumber)
	}

	if w.Imsi != nil && len(*w.Imsi) > 0 {
		imsi, err := tbcd.Decode(*w.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding IMSI: %w", err)
		}
		out.IMSI = imsi
	}

	if w.Msisdn != nil {
		digits, nat, pl, err := decodeAddressField(*w.Msisdn)
		if err != nil {
			return nil, fmt.Errorf("decoding MSISDN: %w", err)
		}
		out.MSISDN = digits
		out.MSISDNNature = nat
		out.MSISDNPlan = pl
	}

	if w.NumberPortabilityStatus != nil {
		v := NumberPortabilityStatus(int(*w.NumberPortabilityStatus))
		out.NumberPortabilityStatus = &v
	}

	return out, nil
}

// --- GprsMSClass (opCode 71) ---

func convertGprsMSClassToWire(g *GprsMSClass) *gsm_map.GPRSMSClass {
	out := &gsm_map.GPRSMSClass{
		MSNetworkCapability: gsm_map.MSNetworkCapability(g.MSNetworkCapability),
	}
	if g.MSRadioAccessCapability != nil {
		rac := gsm_map.MSRadioAccessCapability(g.MSRadioAccessCapability)
		out.MSRadioAccessCapability = &rac
	}
	return out
}

func convertWireToGprsMSClass(w *gsm_map.GPRSMSClass) *GprsMSClass {
	out := &GprsMSClass{
		MSNetworkCapability: []byte(w.MSNetworkCapability),
	}
	if w.MSRadioAccessCapability != nil {
		out.MSRadioAccessCapability = []byte(*w.MSRadioAccessCapability)
	}
	return out
}

// --- UserCSGInformation (opCode 71) ---

func convertUserCSGInformationToWire(u *UserCSGInformation) (*gsm_map.UserCSGInformation, error) {
	if u.CsgIDBits < 0 {
		return nil, fmt.Errorf("CsgIDBits (%d) must be non-negative", u.CsgIDBits)
	}
	if len(u.CsgID) > 0 && u.CsgIDBits == 0 {
		return nil, fmt.Errorf("CsgIDBits must be set when CsgID has bytes (got len %d)", len(u.CsgID))
	}
	if u.CsgIDBits > len(u.CsgID)*8 {
		return nil, fmt.Errorf("CsgIDBits (%d) exceeds len(CsgID)*8 (%d)", u.CsgIDBits, len(u.CsgID)*8)
	}
	out := &gsm_map.UserCSGInformation{
		CsgId: runtime.BitString{
			Bytes:     append([]byte(nil), u.CsgID...),
			BitLength: u.CsgIDBits,
		},
	}
	if u.AccessMode != nil {
		out.AccessMode = []byte(u.AccessMode)
	}
	if u.CMI != nil {
		out.Cmi = []byte(u.CMI)
	}
	return out, nil
}

func convertWireToUserCSGInformation(w *gsm_map.UserCSGInformation) *UserCSGInformation {
	out := &UserCSGInformation{
		CsgID:     append([]byte(nil), w.CsgId.Bytes...),
		CsgIDBits: w.CsgId.BitLength,
	}
	if w.AccessMode != nil {
		out.AccessMode = []byte(w.AccessMode)
	}
	if w.Cmi != nil {
		out.CMI = []byte(w.Cmi)
	}
	return out
}

// --- LocationInformation5GS (opCode 71) ---

func convertLocationInformation5GSToWire(l *LocationInformation5GS) (*gsm_map.LocationInformation5GS, error) {
	out := &gsm_map.LocationInformation5GS{}

	if l.NrCellGlobalIdentity != nil {
		cgi := gsm_map.NRCGI(l.NrCellGlobalIdentity)
		out.NrCellGlobalIdentity = &cgi
	}

	if l.EUtranCellGlobalIdentity != nil {
		cgi := gsm_map.EUTRANCGI(l.EUtranCellGlobalIdentity)
		out.EUtranCellGlobalIdentity = &cgi
	}

	if l.GeographicalInformation != nil {
		raw, err := l.GeographicalInformation.Encode()
		if err != nil {
			return nil, fmt.Errorf("encoding GeographicalInformation: %w", err)
		}
		gi := gsm_map.GeographicalInformation(raw)
		out.GeographicalInformation = &gi
	}

	if l.GeodeticInformation != nil {
		gd := gsm_map.GeodeticInformation(l.GeodeticInformation)
		out.GeodeticInformation = &gd
	}

	if l.AmfAddress != nil {
		amf := gsm_map.FQDN(l.AmfAddress)
		out.AmfAddress = &amf
	}

	if l.TrackingAreaIdentity != nil {
		ta := gsm_map.TAId(l.TrackingAreaIdentity)
		out.TrackingAreaIdentity = &ta
	}

	out.CurrentLocationRetrieved = boolToNullPtr(l.CurrentLocationRetrieved)

	if l.AgeOfLocationInformation != nil {
		age := gsm_map.AgeOfLocationInformation(*l.AgeOfLocationInformation)
		out.AgeOfLocationInformation = &age
	}

	if l.VplmnID != nil {
		if len(l.VplmnID) != 3 {
			return nil, fmt.Errorf("LocationInformation5GS: VplmnID must be exactly 3 octets, got %d", len(l.VplmnID))
		}
		p := gsm_map.PLMNId(l.VplmnID)
		out.VplmnId = &p
	}

	if l.LocalTimeZone != nil {
		tz := gsm_map.TimeZone(l.LocalTimeZone)
		out.LocaltimeZone = &tz
	}

	if l.RatType != nil {
		v := gsm_map.UsedRATType(int64(*l.RatType))
		out.RatType = &v
	}

	if l.NrTrackingAreaIdentity != nil {
		ta := gsm_map.NRTAId(l.NrTrackingAreaIdentity)
		out.NrTrackingAreaIdentity = &ta
	}

	return out, nil
}

func convertWireToLocationInformation5GS(w *gsm_map.LocationInformation5GS) (*LocationInformation5GS, error) {
	out := &LocationInformation5GS{}

	if w.NrCellGlobalIdentity != nil {
		out.NrCellGlobalIdentity = []byte(*w.NrCellGlobalIdentity)
	}

	if w.EUtranCellGlobalIdentity != nil {
		out.EUtranCellGlobalIdentity = []byte(*w.EUtranCellGlobalIdentity)
	}

	if w.GeographicalInformation != nil {
		gi, err := DecodeGeographicalInfo([]byte(*w.GeographicalInformation))
		if err != nil {
			return nil, fmt.Errorf("decoding GeographicalInformation: %w", err)
		}
		out.GeographicalInformation = gi
	}

	if w.GeodeticInformation != nil {
		out.GeodeticInformation = []byte(*w.GeodeticInformation)
	}

	if w.AmfAddress != nil {
		out.AmfAddress = []byte(*w.AmfAddress)
	}

	if w.TrackingAreaIdentity != nil {
		out.TrackingAreaIdentity = []byte(*w.TrackingAreaIdentity)
	}

	out.CurrentLocationRetrieved = nullPtrToBool(w.CurrentLocationRetrieved)

	if w.AgeOfLocationInformation != nil {
		v := int(*w.AgeOfLocationInformation)
		out.AgeOfLocationInformation = &v
	}

	if w.VplmnId != nil {
		out.VplmnID = []byte(*w.VplmnId)
	}

	if w.LocaltimeZone != nil {
		out.LocalTimeZone = []byte(*w.LocaltimeZone)
	}

	if w.RatType != nil {
		v := UsedRatType(int(*w.RatType))
		out.RatType = &v
	}

	if w.NrTrackingAreaIdentity != nil {
		out.NrTrackingAreaIdentity = []byte(*w.NrTrackingAreaIdentity)
	}

	return out, nil
}

// --- InformServiceCentre (opCode 63) ---

// MwStatus: 6 bits per 3GPP TS 29.002.
// Bit 0=scAddressNotIncluded, 1=mnrfSet, 2=mcefSet, 3=mnrgSet, 4=mnr5gSet, 5=mnr5gn3gSet.
func convertMwStatusToBitString(m *MwStatusFlags) runtime.BitString {
	var b byte
	if m.SCAddressNotIncluded {
		b |= 1 << 7
	}
	if m.MnrfSet {
		b |= 1 << 6
	}
	if m.McefSet {
		b |= 1 << 5
	}
	if m.MnrgSet {
		b |= 1 << 4
	}
	if m.Mnr5gSet {
		b |= 1 << 3
	}
	if m.Mnr5gn3gSet {
		b |= 1 << 2
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: 6}
}

func convertBitStringToMwStatus(bs runtime.BitString) *MwStatusFlags {
	m := &MwStatusFlags{}
	if bs.BitLength > 0 {
		m.SCAddressNotIncluded = bs.Has(0)
	}
	if bs.BitLength > 1 {
		m.MnrfSet = bs.Has(1)
	}
	if bs.BitLength > 2 {
		m.McefSet = bs.Has(2)
	}
	if bs.BitLength > 3 {
		m.MnrgSet = bs.Has(3)
	}
	if bs.BitLength > 4 {
		m.Mnr5gSet = bs.Has(4)
	}
	if bs.BitLength > 5 {
		m.Mnr5gn3gSet = bs.Has(5)
	}
	return m
}

func validateAbsentSubscriberDiagnosticSM(p *int) error {
	if p == nil {
		return nil
	}
	if *p < 0 || *p > 255 {
		return ErrIscInvalidAbsentSubscriberDiagnosticSM
	}
	return nil
}

func convertInformServiceCentreToArg(i *InformServiceCentre) (*gsm_map.InformServiceCentreArg, error) {
	// Validate AbsentSubscriberDiagnosticSM fields.
	if err := validateAbsentSubscriberDiagnosticSM(i.AbsentSubscriberDiagnosticSM); err != nil {
		return nil, fmt.Errorf("AbsentSubscriberDiagnosticSM: %w", err)
	}
	if err := validateAbsentSubscriberDiagnosticSM(i.AdditionalAbsentSubscriberDiagnosticSM); err != nil {
		return nil, fmt.Errorf("AdditionalAbsentSubscriberDiagnosticSM: %w", err)
	}
	if err := validateAbsentSubscriberDiagnosticSM(i.Smsf3gppAbsentSubscriberDiagnosticSM); err != nil {
		return nil, fmt.Errorf("Smsf3gppAbsentSubscriberDiagnosticSM: %w", err)
	}
	if err := validateAbsentSubscriberDiagnosticSM(i.SmsfNon3gppAbsentSubscriberDiagnosticSM); err != nil {
		return nil, fmt.Errorf("SmsfNon3gppAbsentSubscriberDiagnosticSM: %w", err)
	}

	arg := &gsm_map.InformServiceCentreArg{}

	if i.StoredMSISDN != "" {
		encoded, err := encodeAddressField(i.StoredMSISDN, i.StoredMSISDNNature, i.StoredMSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding StoredMSISDN: %w", err)
		}
		v := gsm_map.ISDNAddressString(encoded)
		arg.StoredMSISDN = &v
	}

	if i.MwStatus != nil {
		bs := convertMwStatusToBitString(i.MwStatus)
		arg.MwStatus = &bs
	}

	if i.AbsentSubscriberDiagnosticSM != nil {
		v := gsm_map.AbsentSubscriberDiagnosticSM(int64(*i.AbsentSubscriberDiagnosticSM))
		arg.AbsentSubscriberDiagnosticSM = &v
	}
	if i.AdditionalAbsentSubscriberDiagnosticSM != nil {
		v := gsm_map.AbsentSubscriberDiagnosticSM(int64(*i.AdditionalAbsentSubscriberDiagnosticSM))
		arg.AdditionalAbsentSubscriberDiagnosticSM = &v
	}
	if i.Smsf3gppAbsentSubscriberDiagnosticSM != nil {
		v := gsm_map.AbsentSubscriberDiagnosticSM(int64(*i.Smsf3gppAbsentSubscriberDiagnosticSM))
		arg.Smsf3gppAbsentSubscriberDiagnosticSM = &v
	}
	if i.SmsfNon3gppAbsentSubscriberDiagnosticSM != nil {
		v := gsm_map.AbsentSubscriberDiagnosticSM(int64(*i.SmsfNon3gppAbsentSubscriberDiagnosticSM))
		arg.SmsfNon3gppAbsentSubscriberDiagnosticSM = &v
	}

	return arg, nil
}

func convertArgToInformServiceCentre(arg *gsm_map.InformServiceCentreArg) (*InformServiceCentre, error) {
	out := &InformServiceCentre{}

	if arg.StoredMSISDN != nil {
		digits, nature, plan, err := decodeAddressField(*arg.StoredMSISDN)
		if err != nil {
			return nil, fmt.Errorf("decoding StoredMSISDN: %w", err)
		}
		out.StoredMSISDN = digits
		out.StoredMSISDNNature = nature
		out.StoredMSISDNPlan = plan
	}

	if arg.MwStatus != nil {
		out.MwStatus = convertBitStringToMwStatus(*arg.MwStatus)
	}

	if arg.AbsentSubscriberDiagnosticSM != nil {
		v := int(*arg.AbsentSubscriberDiagnosticSM)
		if v < 0 || v > 255 {
			return nil, fmt.Errorf("AbsentSubscriberDiagnosticSM: %w", ErrIscInvalidAbsentSubscriberDiagnosticSM)
		}
		out.AbsentSubscriberDiagnosticSM = &v
	}
	if arg.AdditionalAbsentSubscriberDiagnosticSM != nil {
		v := int(*arg.AdditionalAbsentSubscriberDiagnosticSM)
		if v < 0 || v > 255 {
			return nil, fmt.Errorf("AdditionalAbsentSubscriberDiagnosticSM: %w", ErrIscInvalidAbsentSubscriberDiagnosticSM)
		}
		out.AdditionalAbsentSubscriberDiagnosticSM = &v
	}
	if arg.Smsf3gppAbsentSubscriberDiagnosticSM != nil {
		v := int(*arg.Smsf3gppAbsentSubscriberDiagnosticSM)
		if v < 0 || v > 255 {
			return nil, fmt.Errorf("Smsf3gppAbsentSubscriberDiagnosticSM: %w", ErrIscInvalidAbsentSubscriberDiagnosticSM)
		}
		out.Smsf3gppAbsentSubscriberDiagnosticSM = &v
	}
	if arg.SmsfNon3gppAbsentSubscriberDiagnosticSM != nil {
		v := int(*arg.SmsfNon3gppAbsentSubscriberDiagnosticSM)
		if v < 0 || v > 255 {
			return nil, fmt.Errorf("SmsfNon3gppAbsentSubscriberDiagnosticSM: %w", ErrIscInvalidAbsentSubscriberDiagnosticSM)
		}
		out.SmsfNon3gppAbsentSubscriberDiagnosticSM = &v
	}

	return out, nil
}
