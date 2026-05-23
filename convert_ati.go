package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// --- AnyTimeInterrogation ---

func convertATIToArg(ati *AnyTimeInterrogation) (*gsm_map.AnyTimeInterrogationArg, error) {
	if ati == nil {
		return nil, ErrAnyTimeInterrogationNil
	}
	subId, err := convertSubscriberIdentityToWire(ati.SubscriberIdentity)
	if err != nil {
		return nil, fmt.Errorf("AnyTimeInterrogation.SubscriberIdentity: %w", err)
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

	subId, err := convertWireToSubscriberIdentity(arg.SubscriberIdentity)
	if err != nil {
		return nil, fmt.Errorf("AnyTimeInterrogation.SubscriberIdentity: %w", err)
	}
	ati.SubscriberIdentity = subId

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
