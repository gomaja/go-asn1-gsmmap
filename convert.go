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

	return &gsm_map.RoutingInfoForSMArg{
		Msisdn:               gsm_map.ISDNAddressString(msisdn),
		SmRPPRI:               s.SmRpPri,
		ServiceCentreAddress: gsm_map.AddressString(sca),
	}, nil
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

	return &SriSm{
		MSISDN:               msisdn,
		MSISDNNature:         msisdnNature,
		MSISDNPlan:           msisdnPlan,
		SmRpPri:              arg.SmRPPRI,
		ServiceCentreAddress: sca,
		SCANature:            scaNature,
		SCAPlan:              scaPlan,
	}, nil
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

	return &gsm_map.RoutingInfoForSMRes{
		Imsi: gsm_map.IMSI(imsiBytes),
		LocationInfoWithLMSI: gsm_map.LocationInfoWithLMSI{
			NetworkNodeNumber: gsm_map.ISDNAddressString(nnn),
		},
	}, nil
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

	return &SriSmResp{
		IMSI: imsi,
		LocationInfoWithLMSI: LocationInfoWithLMSI{
			NetworkNodeNumber:       nnn,
			NetworkNodeNumberNature: nnnNature,
			NetworkNodeNumberPlan:   nnnPlan,
		},
	}, nil
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

	if m.MoreMessagesToSend {
		marker := struct{}{}
		arg.MoreMessagesToSend = &marker
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
	mtFsm.MoreMessagesToSend = arg.MoreMessagesToSend != nil

	return &mtFsm, nil
}

// --- MO-ForwardSM ---

func convertMoFsmToArg(m *MoFsm) (*gsm_map.MOForwardSMArg, error) {
	scaDA, err := encodeAddressField(m.ServiceCentreAddressDA, m.SCADANature, m.SCADAPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding ServiceCentreAddressDA: %w", err)
	}

	msisdn, err := encodeAddressField(m.MSISDN, m.MSISDNNature, m.MSISDNPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding MSISDN: %w", err)
	}

	tpduBytes, err := m.TPDU.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshaling TPDU: %w", err)
	}

	return &gsm_map.MOForwardSMArg{
		SmRPDA: gsm_map.NewSMRPDAServiceCentreAddressDA(gsm_map.AddressString(scaDA)),
		SmRPOA: gsm_map.NewSMRPOAMsisdn(gsm_map.ISDNAddressString(msisdn)),
		SmRPUI: gsm_map.SignalInfo(tpduBytes),
	}, nil
}

func convertArgToMoFsm(arg *gsm_map.MOForwardSMArg) (*MoFsm, error) {
	var moFsm MoFsm

	// Extract ServiceCentreAddressDA from SM-RP-DA
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
	default:
		return nil, fmt.Errorf("unexpected SMRPDA choice: %d", arg.SmRPDA.Choice)
	}

	// Extract MSISDN from SM-RP-OA
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

	return &moFsm, nil
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
	res := &gsm_map.AnyTimeInterrogationRes{}

	si := &res.SubscriberInfo

	// LocationInformation (CS)
	if atiRes.SubscriberInfo.LocationInformation != nil {
		locInfo, err := convertCSLocationToAsn1(atiRes.SubscriberInfo.LocationInformation)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformation: %w", err)
		}
		si.LocationInformation = locInfo
	}

	// SubscriberState
	if atiRes.SubscriberInfo.SubscriberState != nil {
		si.SubscriberState = convertSubscriberStateToAsn1(atiRes.SubscriberInfo.SubscriberState)
	}

	// LocationInformationEPS
	if atiRes.SubscriberInfo.LocationInformationEPS != nil {
		locEPS, err := convertEPSLocationToAsn1(atiRes.SubscriberInfo.LocationInformationEPS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationEPS: %w", err)
		}
		si.LocationInformationEPS = locEPS
	}

	// LocationInformationGPRS
	if atiRes.SubscriberInfo.LocationInformationGPRS != nil {
		locGPRS, err := convertGPRSLocationToAsn1(atiRes.SubscriberInfo.LocationInformationGPRS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationGPRS: %w", err)
		}
		si.LocationInformationGPRS = locGPRS
	}

	// IMEI
	if atiRes.SubscriberInfo.IMEI != "" {
		imeiBytes, err := tbcd.Encode(atiRes.SubscriberInfo.IMEI)
		if err != nil {
			return nil, fmt.Errorf("encoding IMEI: %w", err)
		}
		imei := gsm_map.IMEI(imeiBytes)
		si.Imei = &imei
	}

	// MsClassmark2
	if atiRes.SubscriberInfo.MsClassmark2 != nil {
		mc := gsm_map.MSClassmark2(atiRes.SubscriberInfo.MsClassmark2)
		si.MsClassmark2 = &mc
	}

	// TimeZone
	if atiRes.SubscriberInfo.TimeZone != nil {
		tz := gsm_map.TimeZone(atiRes.SubscriberInfo.TimeZone)
		si.TimeZone = &tz
	}

	// DaylightSavingTime
	if atiRes.SubscriberInfo.DaylightSavingTime != nil {
		dst := gsm_map.DaylightSavingTime(*atiRes.SubscriberInfo.DaylightSavingTime)
		si.DaylightSavingTime = &dst
	}

	return res, nil
}

func convertResToATIRes(res *gsm_map.AnyTimeInterrogationRes) (*AnyTimeInterrogationRes, error) {
	atiRes := &AnyTimeInterrogationRes{}
	si := &res.SubscriberInfo

	// LocationInformation (CS)
	if si.LocationInformation != nil {
		locInfo, err := convertAsn1ToCSLocation(si.LocationInformation)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformation: %w", err)
		}
		atiRes.SubscriberInfo.LocationInformation = locInfo
	}

	// SubscriberState
	if si.SubscriberState != nil {
		atiRes.SubscriberInfo.SubscriberState = convertAsn1ToSubscriberState(si.SubscriberState)
	}

	// LocationInformationEPS
	if si.LocationInformationEPS != nil {
		locEPS, err := convertAsn1ToEPSLocation(si.LocationInformationEPS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationEPS: %w", err)
		}
		atiRes.SubscriberInfo.LocationInformationEPS = locEPS
	}

	// LocationInformationGPRS
	if si.LocationInformationGPRS != nil {
		locGPRS, err := convertAsn1ToGPRSLocation(si.LocationInformationGPRS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationGPRS: %w", err)
		}
		atiRes.SubscriberInfo.LocationInformationGPRS = locGPRS
	}

	// IMEI
	if si.Imei != nil && len(*si.Imei) > 0 {
		imei, err := tbcd.Decode(*si.Imei)
		if err != nil {
			return nil, fmt.Errorf("decoding IMEI: %w", err)
		}
		atiRes.SubscriberInfo.IMEI = imei
	}

	// MsClassmark2
	if si.MsClassmark2 != nil {
		atiRes.SubscriberInfo.MsClassmark2 = []byte(*si.MsClassmark2)
	}

	// TimeZone
	if si.TimeZone != nil {
		atiRes.SubscriberInfo.TimeZone = []byte(*si.TimeZone)
	}

	// DaylightSavingTime
	if si.DaylightSavingTime != nil {
		v := int(*si.DaylightSavingTime)
		atiRes.SubscriberInfo.DaylightSavingTime = &v
	}

	return atiRes, nil
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
		cgid := gsm_map.CellGlobalIdOrServiceAreaIdFixedLength(loc.CellGlobalId)
		li.CellGlobalIdOrServiceAreaIdOrLAI = &gsm_map.CellGlobalIdOrServiceAreaIdOrLAI{
			Choice:                                 gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceCellGlobalIdOrServiceAreaIdFixedLength,
			CellGlobalIdOrServiceAreaIdFixedLength: &cgid,
		}
	} else if loc.LAI != nil {
		lai := gsm_map.LAIFixedLength(loc.LAI)
		li.CellGlobalIdOrServiceAreaIdOrLAI = &gsm_map.CellGlobalIdOrServiceAreaIdOrLAI{
			Choice:         gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceLaiFixedLength,
			LaiFixedLength: &lai,
		}
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
	s := &gsm_map.SubscriberState{}
	switch ss.State {
	case StateAssumedIdle:
		s.Choice = gsm_map.SubscriberStateChoiceAssumedIdle
		s.AssumedIdle = &struct{}{}
	case StateCamelBusy:
		s.Choice = gsm_map.SubscriberStateChoiceCamelBusy
		s.CamelBusy = &struct{}{}
	case StateNetDetNotReachable:
		s.Choice = gsm_map.SubscriberStateChoiceNetDetNotReachable
		if ss.NotReachableReason != nil {
			reason := gsm_map.NotReachableReason(*ss.NotReachableReason)
			s.NetDetNotReachable = &reason
		} else {
			reason := gsm_map.NotReachableReason(0)
			s.NetDetNotReachable = &reason
		}
	case StateNotProvidedFromVLR:
		s.Choice = gsm_map.SubscriberStateChoiceNotProvidedFromVLR
		s.NotProvidedFromVLR = &struct{}{}
	}
	return s
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
		cgid := gsm_map.CellGlobalIdOrServiceAreaIdFixedLength(loc.CellGlobalId)
		li.CellGlobalIdOrServiceAreaIdOrLAI = &gsm_map.CellGlobalIdOrServiceAreaIdOrLAI{
			Choice:                                 gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceCellGlobalIdOrServiceAreaIdFixedLength,
			CellGlobalIdOrServiceAreaIdFixedLength: &cgid,
		}
	} else if loc.LAI != nil {
		lai := gsm_map.LAIFixedLength(loc.LAI)
		li.CellGlobalIdOrServiceAreaIdOrLAI = &gsm_map.CellGlobalIdOrServiceAreaIdOrLAI{
			Choice:         gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceLaiFixedLength,
			LaiFixedLength: &lai,
		}
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
