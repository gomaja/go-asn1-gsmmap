package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

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
		dt := *ri.RequestedDomain
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
	ati.RequestedInfo = buildRequestedInfoFromWire(&arg.RequestedInfo)

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

// buildRequestedInfoFromWire converts gsm_map.MSRequestedInfo to the public
// RequestedInfo type. Shared between ATI (opCode 71) and PSI (opCode 70).
func buildRequestedInfoFromWire(ri *gsm_map.MSRequestedInfo) RequestedInfo {
	var out RequestedInfo
	out.LocationInformation = ri.LocationInformation != nil
	out.SubscriberState = ri.SubscriberState != nil
	out.CurrentLocation = ri.CurrentLocation != nil
	out.MsClassmark = ri.MsClassmark != nil
	out.IMEI = ri.Imei != nil
	out.MnpRequestedInfo = ri.MnpRequestedInfo != nil
	out.LocationInformationEPSSupported = ri.LocationInformationEPSSupported != nil
	out.TAdsData = ri.TAdsData != nil
	out.ServingNodeIndication = ri.ServingNodeIndication != nil
	out.LocalTimeZoneRequest = ri.LocalTimeZoneRequest != nil

	if ri.RequestedDomain != nil {
		domain := *ri.RequestedDomain
		// Per spec: values > 1 shall be mapped to cs-Domain
		if domain > PsDomain {
			domain = CsDomain
		}
		out.RequestedDomain = &domain
	}

	if ri.RequestedNodes != nil && ri.RequestedNodes.BitLength > 0 {
		out.RequestedNodes = convertBitStringToRequestedNodes(*ri.RequestedNodes)
	}
	return out
}
