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
	if ati.GsmSCFAddress == "" {
		return nil, ErrAtiMissingGsmSCFAddress
	}
	subId, err := convertSubscriberIdentityToWire(ati.SubscriberIdentity)
	if err != nil {
		return nil, fmt.Errorf("AnyTimeInterrogation.SubscriberIdentity: %w", err)
	}

	reqInfo, err := buildRequestedInfo(&ati.RequestedInfo)
	if err != nil {
		return nil, fmt.Errorf("AnyTimeInterrogation.RequestedInfo: %w", err)
	}

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

// buildRequestedInfo converts the public RequestedInfo to gsm_map.RequestedInfo.
// Shared between ATI (opCode 71) and PSI (opCode 70).
func buildRequestedInfo(ri *RequestedInfo) (gsm_map.RequestedInfo, error) {
	var wire gsm_map.RequestedInfo

	nullMarker := &struct{}{}

	if ri.LocationInformation {
		wire.LocationInformation = nullMarker
	}
	if ri.SubscriberState {
		wire.SubscriberState = nullMarker
	}
	if ri.CurrentLocation {
		wire.CurrentLocation = nullMarker
	}
	if ri.RequestedDomain != nil {
		// A sender uses only the listed domains; a receiver maps any value
		// above ps-Domain to cs-Domain (buildRequestedInfoFromWire).
		dt := *ri.RequestedDomain
		if dt != CsDomain && dt != PsDomain {
			return gsm_map.RequestedInfo{}, fmt.Errorf("%w (got %d)", ErrRequestedDomainInvalid, dt)
		}
		wire.RequestedDomain = &dt
	}
	if ri.MsClassmark {
		wire.MsClassmark = nullMarker
	}
	if ri.IMEI {
		wire.Imei = nullMarker
	}
	if ri.MnpRequestedInfo {
		wire.MnpRequestedInfo = nullMarker
	}
	if ri.LocationInformationEPSSupported {
		wire.LocationInformationEPSSupported = nullMarker
	}
	if ri.TAdsData {
		wire.TAdsData = nullMarker
	}
	if ri.RequestedNodes != nil {
		bs := convertRequestedNodesToBitString(ri.RequestedNodes)
		wire.RequestedNodes = &bs
	}
	if ri.ServingNodeIndication {
		wire.ServingNodeIndication = nullMarker
	}
	if ri.LocalTimeZoneRequest {
		wire.LocalTimeZoneRequest = nullMarker
	}

	return wire, nil
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
	scf, scfNature, scfPlan, err := decodeAddressWithDigits(arg.GsmSCFAddress, ErrAtiMissingGsmSCFAddress)
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

// buildRequestedInfoFromWire converts gsm_map.RequestedInfo to the public
// RequestedInfo type. Shared between ATI (opCode 71) and PSI (opCode 70).
func buildRequestedInfoFromWire(ri *gsm_map.RequestedInfo) RequestedInfo {
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
		// 3GPP TS 29.002 V19.1.0 §17.7.1 DomainType: "reception of values
		// > 1 shall be mapped to 'cs-Domain'".
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
