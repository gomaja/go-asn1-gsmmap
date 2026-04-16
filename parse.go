package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ParseSriSm decodes BER-encoded bytes into an SriSm.
func ParseSriSm(data []byte) (*SriSm, error) {
	var arg gsm_map.RoutingInfoForSMArg
	if err := arg.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding RoutingInfoForSMArg: %w", err)
	}
	return convertArgToSriSm(&arg)
}

// ParseSriSmResp decodes BER-encoded bytes into an SriSmResp.
func ParseSriSmResp(data []byte) (*SriSmResp, error) {
	var res gsm_map.RoutingInfoForSMRes
	if err := res.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding RoutingInfoForSMRes: %w", err)
	}
	return convertResToSriSmResp(&res)
}

// ParseMtFsm decodes BER-encoded bytes into an MtFsm.
func ParseMtFsm(data []byte) (*MtFsm, error) {
	var arg gsm_map.MTForwardSMArg
	if err := arg.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding MTForwardSMArg: %w", err)
	}
	return convertArgToMtFsm(&arg)
}

// ParseMoFsm decodes BER-encoded bytes into an MoFsm.
func ParseMoFsm(data []byte) (*MoFsm, error) {
	var arg gsm_map.MOForwardSMArg
	if err := arg.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding MOForwardSMArg: %w", err)
	}
	return convertArgToMoFsm(&arg)
}

// ParseUpdateLocation decodes BER-encoded bytes into an UpdateLocation.
func ParseUpdateLocation(data []byte) (*UpdateLocation, error) {
	var arg gsm_map.UpdateLocationArg
	if err := arg.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding UpdateLocationArg: %w", err)
	}
	return convertArgToUpdateLocation(&arg)
}

// ParseUpdateLocationRes decodes BER-encoded bytes into an UpdateLocationRes.
func ParseUpdateLocationRes(data []byte) (*UpdateLocationRes, error) {
	var res gsm_map.UpdateLocationRes
	if err := res.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding UpdateLocationRes: %w", err)
	}
	return convertResToUpdateLocationRes(&res)
}

// ParseUpdateGprsLocation decodes BER-encoded bytes into an UpdateGprsLocation.
func ParseUpdateGprsLocation(data []byte) (*UpdateGprsLocation, error) {
	var arg gsm_map.UpdateGprsLocationArg
	if err := arg.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding UpdateGprsLocationArg: %w", err)
	}
	return convertArgToUpdateGprsLocation(&arg)
}

// ParseUpdateGprsLocationRes decodes BER-encoded bytes into an UpdateGprsLocationRes.
func ParseUpdateGprsLocationRes(data []byte) (*UpdateGprsLocationRes, error) {
	var res gsm_map.UpdateGprsLocationRes
	if err := res.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding UpdateGprsLocationRes: %w", err)
	}
	return convertResToUpdateGprsLocationRes(&res)
}

// ParseAnyTimeInterrogation decodes BER-encoded bytes into an AnyTimeInterrogation.
func ParseAnyTimeInterrogation(data []byte) (*AnyTimeInterrogation, error) {
	var arg gsm_map.AnyTimeInterrogationArg
	if err := arg.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding AnyTimeInterrogationArg: %w", err)
	}
	return convertArgToATI(&arg)
}

// ParseAnyTimeInterrogationRes decodes BER-encoded bytes into an AnyTimeInterrogationRes.
func ParseAnyTimeInterrogationRes(data []byte) (*AnyTimeInterrogationRes, error) {
	var res gsm_map.AnyTimeInterrogationRes
	if err := res.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding AnyTimeInterrogationRes: %w", err)
	}
	return convertResToATIRes(&res)
}

// ParseSri decodes BER-encoded bytes into an Sri.
func ParseSri(data []byte) (*Sri, error) {
	var arg gsm_map.SendRoutingInfoArg
	if err := arg.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding SendRoutingInfoArg: %w", err)
	}
	return convertArgToSri(&arg)
}

// ParseSriResp decodes BER-encoded bytes into an SriResp.
func ParseSriResp(data []byte) (*SriResp, error) {
	var res gsm_map.SendRoutingInfoRes
	if err := res.UnmarshalBER(data); err != nil {
		return nil, fmt.Errorf("decoding SendRoutingInfoRes: %w", err)
	}
	return convertResToSriResp(&res)
}
