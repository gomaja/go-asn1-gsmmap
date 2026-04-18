package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// --- ProvideSubscriberInfo (opCode 70) ---

func validateProvideSubscriberInfo(p *ProvideSubscriberInfo) error {
	if p.IMSI == "" {
		return ErrPsiMissingIMSI
	}
	if len(p.LMSI) != 0 && len(p.LMSI) != 4 {
		return ErrPsiInvalidLMSI
	}
	if p.CallPriority != nil {
		v := *p.CallPriority
		if v < 0 || v > 15 {
			return ErrPsiInvalidCallPriority
		}
	}
	return nil
}

func convertProvideSubscriberInfoToArg(p *ProvideSubscriberInfo) (*gsm_map.ProvideSubscriberInfoArg, error) {
	if err := validateProvideSubscriberInfo(p); err != nil {
		return nil, err
	}

	imsiBytes, err := tbcd.Encode(p.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	arg := &gsm_map.ProvideSubscriberInfoArg{
		Imsi:          gsm_map.IMSI(imsiBytes),
		RequestedInfo: buildMSRequestedInfo(&p.RequestedInfo),
	}

	// LMSI (optional, 4 octets).
	if len(p.LMSI) > 0 {
		v := gsm_map.LMSI(p.LMSI)
		arg.Lmsi = &v
	}

	// CallPriority (optional, 0..15).
	if p.CallPriority != nil {
		v := gsm_map.EMLPPPriority(int64(*p.CallPriority))
		arg.CallPriority = &v
	}

	return arg, nil
}

func convertArgToProvideSubscriberInfo(arg *gsm_map.ProvideSubscriberInfoArg) (*ProvideSubscriberInfo, error) {
	imsi, err := tbcd.Decode(arg.Imsi)
	if err != nil {
		return nil, fmt.Errorf("decoding IMSI: %w", err)
	}
	if imsi == "" {
		return nil, ErrPsiMissingIMSI
	}

	out := &ProvideSubscriberInfo{
		IMSI:          imsi,
		RequestedInfo: buildRequestedInfoFromWire(&arg.RequestedInfo),
	}

	// LMSI (optional, must be exactly 4 octets when present).
	if arg.Lmsi != nil {
		lmsi := []byte(*arg.Lmsi)
		if len(lmsi) != 4 {
			return nil, ErrPsiInvalidLMSI
		}
		out.LMSI = HexBytes(lmsi)
	}

	// CallPriority (optional, 0..15).
	if arg.CallPriority != nil {
		v := int64(*arg.CallPriority)
		if v < 0 || v > 15 {
			return nil, ErrPsiInvalidCallPriority
		}
		iv := int(v)
		out.CallPriority = &iv
	}

	return out, nil
}

func convertProvideSubscriberInfoResToRes(p *ProvideSubscriberInfoRes) (*gsm_map.ProvideSubscriberInfoRes, error) {
	si, err := convertSubscriberInfoToWire(&p.SubscriberInfo)
	if err != nil {
		return nil, err
	}
	return &gsm_map.ProvideSubscriberInfoRes{SubscriberInfo: *si}, nil
}

func convertResToProvideSubscriberInfoRes(res *gsm_map.ProvideSubscriberInfoRes) (*ProvideSubscriberInfoRes, error) {
	si, err := convertWireToSubscriberInfo(&res.SubscriberInfo)
	if err != nil {
		return nil, err
	}
	return &ProvideSubscriberInfoRes{SubscriberInfo: *si}, nil
}
