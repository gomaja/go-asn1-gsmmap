package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// --- PurgeMS (opCode 67) ---

// convertPurgeMSToArg converts the public PurgeMS into the wire-level
// gsm_map.PurgeMSArg.
func convertPurgeMSToArg(p *PurgeMS) (*gsm_map.PurgeMSArg, error) {
	if p.IMSI == "" {
		return nil, ErrPurgeMSMissingIMSI
	}

	imsiBytes, err := encodeIdentityDigits(identityIMSI, p.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	arg := &gsm_map.PurgeMSArg{
		Imsi: imsiBytes,
	}

	// [0] VLR-Number
	if p.VlrNumber != "" {
		encoded, err := encodeAddressField(p.VlrNumber, p.VlrNumberNature, p.VlrNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding VlrNumber: %w", err)
		}
		v := encoded
		arg.VlrNumber = &v
	}

	// [1] SGSN-Number
	if p.SgsnNumber != "" {
		encoded, err := encodeAddressField(p.SgsnNumber, p.SgsnNumberNature, p.SgsnNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SgsnNumber: %w", err)
		}
		v := encoded
		arg.SgsnNumber = &v
	}

	// [2] LocationInformation
	if p.LocationInformation != nil {
		loc, err := convertCSLocationToAsn1(p.LocationInformation)
		if err != nil {
			return nil, fmt.Errorf("LocationInformation: %w", err)
		}
		arg.LocationInformation = loc
	}

	// [3] LocationInformationGPRS
	if p.LocationInformationGPRS != nil {
		loc, err := convertGPRSLocationToAsn1(p.LocationInformationGPRS)
		if err != nil {
			return nil, fmt.Errorf("LocationInformationGPRS: %w", err)
		}
		arg.LocationInformationGPRS = loc
	}

	// [4] LocationInformationEPS
	if p.LocationInformationEPS != nil {
		loc, err := convertEPSLocationToAsn1(p.LocationInformationEPS)
		if err != nil {
			return nil, fmt.Errorf("LocationInformationEPS: %w", err)
		}
		arg.LocationInformationEPS = loc
	}

	return arg, nil
}

// convertArgToPurgeMS converts a wire-level gsm_map.PurgeMSArg back into the
// public PurgeMS type.
func convertArgToPurgeMS(arg *gsm_map.PurgeMSArg) (*PurgeMS, error) {
	imsi, err := decodeIdentityDigits(identityIMSI, arg.Imsi)
	if err != nil {
		return nil, fmt.Errorf("decoding IMSI: %w", err)
	}

	out := &PurgeMS{IMSI: imsi}

	if arg.VlrNumber != nil {
		digits, nature, plan, err := decodeAddressWithDigits(*arg.VlrNumber, ErrPurgeMSVLRNumberDecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding VlrNumber: %w", err)
		}
		out.VlrNumber = digits
		out.VlrNumberNature = nature
		out.VlrNumberPlan = plan
	}

	if arg.SgsnNumber != nil {
		digits, nature, plan, err := decodeAddressWithDigits(*arg.SgsnNumber, ErrPurgeMSSGSNNumberDecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding SgsnNumber: %w", err)
		}
		out.SgsnNumber = digits
		out.SgsnNumberNature = nature
		out.SgsnNumberPlan = plan
	}

	if arg.LocationInformation != nil {
		loc, err := convertAsn1ToCSLocation(arg.LocationInformation)
		if err != nil {
			return nil, fmt.Errorf("LocationInformation: %w", err)
		}
		out.LocationInformation = loc
	}

	if arg.LocationInformationGPRS != nil {
		loc, err := convertAsn1ToGPRSLocation(arg.LocationInformationGPRS)
		if err != nil {
			return nil, fmt.Errorf("LocationInformationGPRS: %w", err)
		}
		out.LocationInformationGPRS = loc
	}

	if arg.LocationInformationEPS != nil {
		loc, err := convertAsn1ToEPSLocation(arg.LocationInformationEPS)
		if err != nil {
			return nil, fmt.Errorf("LocationInformationEPS: %w", err)
		}
		out.LocationInformationEPS = loc
	}

	return out, nil
}

// convertPurgeMSResToWire converts the public PurgeMSRes into the wire-level
// gsm_map.PurgeMSRes.
func convertPurgeMSResToWire(r *PurgeMSRes) *gsm_map.PurgeMSRes {
	return &gsm_map.PurgeMSRes{
		FreezeTMSI:  boolToNullPtr(r.FreezeTMSI),
		FreezePTMSI: boolToNullPtr(r.FreezePTMSI),
		FreezeMTMSI: boolToNullPtr(r.FreezeMTMSI),
	}
}

// convertWireToPurgeMSRes converts a wire-level gsm_map.PurgeMSRes back into
// the public PurgeMSRes type.
func convertWireToPurgeMSRes(res *gsm_map.PurgeMSRes) *PurgeMSRes {
	return &PurgeMSRes{
		FreezeTMSI:  nullPtrToBool(res.FreezeTMSI),
		FreezePTMSI: nullPtrToBool(res.FreezePTMSI),
		FreezeMTMSI: nullPtrToBool(res.FreezeMTMSI),
	}
}
