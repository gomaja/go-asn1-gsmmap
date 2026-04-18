package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// --- PurgeMS (opCode 67) ---

// convertPurgeMSToArg converts the public PurgeMS into the wire-level
// gsm_map.PurgeMSArg.
func convertPurgeMSToArg(p *PurgeMS) (*gsm_map.PurgeMSArg, error) {
	if p.IMSI == "" {
		return nil, ErrPurgeMSMissingIMSI
	}

	imsiBytes, err := tbcd.Encode(p.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	arg := &gsm_map.PurgeMSArg{
		Imsi: gsm_map.IMSI(imsiBytes),
	}

	// [0] VLR-Number
	if p.VLRNumber != "" {
		encoded, err := encodeAddressField(p.VLRNumber, p.VLRNature, p.VLRPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding VLRNumber: %w", err)
		}
		v := gsm_map.ISDNAddressString(encoded)
		arg.VlrNumber = &v
	}

	// [1] SGSN-Number
	if p.SGSNNumber != "" {
		encoded, err := encodeAddressField(p.SGSNNumber, p.SGSNNature, p.SGSNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SGSNNumber: %w", err)
		}
		v := gsm_map.ISDNAddressString(encoded)
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
	if len(arg.Imsi) == 0 {
		return nil, ErrPurgeMSMissingIMSI
	}

	imsi, err := tbcd.Decode(arg.Imsi)
	if err != nil {
		return nil, fmt.Errorf("decoding IMSI: %w", err)
	}

	out := &PurgeMS{IMSI: imsi}

	if arg.VlrNumber != nil {
		digits, nature, plan, err := decodeAddressField(*arg.VlrNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding VLRNumber: %w", err)
		}
		out.VLRNumber = digits
		out.VLRNature = nature
		out.VLRPlan = plan
	}

	if arg.SgsnNumber != nil {
		digits, nature, plan, err := decodeAddressField(*arg.SgsnNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding SGSNNumber: %w", err)
		}
		out.SGSNNumber = digits
		out.SGSNNature = nature
		out.SGSNPlan = plan
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
