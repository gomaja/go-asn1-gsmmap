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

		arg.VlrCapability = vlrCap
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

		// Only set if at least one capability was parsed
		if vlrCap.SupportedCamelPhases != nil || vlrCap.SupportedLCSCapabilitySets != nil {
			u.VlrCapability = vlrCap
		}
	}

	return u, nil
}

// --- UpdateLocationRes ---

func convertUpdateLocationResToRes(u *UpdateLocationRes) (*gsm_map.UpdateLocationRes, error) {
	hlr, err := encodeAddressField(u.HLRNumber, u.HLRNumberNature, u.HLRNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding HLRNumber: %w", err)
	}

	return &gsm_map.UpdateLocationRes{
		HlrNumber: gsm_map.ISDNAddressString(hlr),
	}, nil
}

func convertResToUpdateLocationRes(res *gsm_map.UpdateLocationRes) (*UpdateLocationRes, error) {
	hlr, nature, plan, err := decodeAddressField(res.HlrNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding HLRNumber: %w", err)
	}

	return &UpdateLocationRes{
		HLRNumber:       hlr,
		HLRNumberNature: nature,
		HLRNumberPlan:   plan,
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
		sgsnCap := &gsm_map.SGSNCapability{}

		if u.SGSNCapability.GprsEnhancementsSupportIndicator {
			marker := struct{}{}
			sgsnCap.GprsEnhancementsSupportIndicator = &marker
		}

		if u.SGSNCapability.SupportedLCSCapabilitySets != nil {
			bs := convertLCSCapsToBitString(u.SGSNCapability.SupportedLCSCapabilitySets)
			sgsnCap.SupportedLCSCapabilitySets = &bs
		}

		arg.SgsnCapability = sgsnCap
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
		sgsnCap := &SGSNCapability{}

		sgsnCap.GprsEnhancementsSupportIndicator = arg.SgsnCapability.GprsEnhancementsSupportIndicator != nil

		if arg.SgsnCapability.SupportedLCSCapabilitySets != nil && arg.SgsnCapability.SupportedLCSCapabilitySets.BitLength > 0 {
			sgsnCap.SupportedLCSCapabilitySets = convertBitStringToLCSCaps(*arg.SgsnCapability.SupportedLCSCapabilitySets)
		}

		if sgsnCap.GprsEnhancementsSupportIndicator || sgsnCap.SupportedLCSCapabilitySets != nil {
			u.SGSNCapability = sgsnCap
		}
	}

	return u, nil
}

// --- UpdateGprsLocationRes ---

func convertUpdateGprsLocationResToRes(u *UpdateGprsLocationRes) (*gsm_map.UpdateGprsLocationRes, error) {
	hlr, err := encodeAddressField(u.HLRNumber, u.HLRNumberNature, u.HLRNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding HLRNumber: %w", err)
	}

	return &gsm_map.UpdateGprsLocationRes{
		HlrNumber: gsm_map.ISDNAddressString(hlr),
	}, nil
}

func convertResToUpdateGprsLocationRes(res *gsm_map.UpdateGprsLocationRes) (*UpdateGprsLocationRes, error) {
	hlr, nature, plan, err := decodeAddressField(res.HlrNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding HLRNumber: %w", err)
	}

	return &UpdateGprsLocationRes{
		HLRNumber:       hlr,
		HLRNumberNature: nature,
		HLRNumberPlan:   plan,
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

	if s.LocationInformationEPS != nil {
		locEPS, err := convertEPSLocationToAsn1(s.LocationInformationEPS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationEPS: %w", err)
		}
		si.LocationInformationEPS = locEPS
	}

	if s.LocationInformationGPRS != nil {
		locGPRS, err := convertGPRSLocationToAsn1(s.LocationInformationGPRS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationGPRS: %w", err)
		}
		si.LocationInformationGPRS = locGPRS
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

	if s.TimeZone != nil {
		tz := gsm_map.TimeZone(s.TimeZone)
		si.TimeZone = &tz
	}

	if s.DaylightSavingTime != nil {
		dst := gsm_map.DaylightSavingTime(*s.DaylightSavingTime)
		si.DaylightSavingTime = &dst
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

	if si.LocationInformationEPS != nil {
		locEPS, err := convertAsn1ToEPSLocation(si.LocationInformationEPS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationEPS: %w", err)
		}
		out.LocationInformationEPS = locEPS
	}

	if si.LocationInformationGPRS != nil {
		locGPRS, err := convertAsn1ToGPRSLocation(si.LocationInformationGPRS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationGPRS: %w", err)
		}
		out.LocationInformationGPRS = locGPRS
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

	if si.TimeZone != nil {
		out.TimeZone = []byte(*si.TimeZone)
	}

	if si.DaylightSavingTime != nil {
		v := int(*si.DaylightSavingTime)
		out.DaylightSavingTime = &v
	}

	return out, nil
}
