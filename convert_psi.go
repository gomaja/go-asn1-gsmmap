package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// --- ProvideSubscriberInfo (opCode 70) ---

func validateProvideSubscriberInfo(p *ProvideSubscriberInfo) error {
	if p.IMSI == "" {
		return ErrPsiMissingIMSI
	}

	return nil
}

func convertProvideSubscriberInfoToArg(p *ProvideSubscriberInfo) (*gsm_map.ProvideSubscriberInfoArg, error) {
	if err := validateProvideSubscriberInfo(p); err != nil {
		return nil, err
	}

	imsiBytes, err := encodeIdentityDigits(identityIMSI, p.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	reqInfo, err := buildRequestedInfo(&p.RequestedInfo)
	if err != nil {
		return nil, fmt.Errorf("ProvideSubscriberInfo.RequestedInfo: %w", err)
	}

	arg := &gsm_map.ProvideSubscriberInfoArg{
		Imsi:          gsm_map.IMSI(imsiBytes),
		RequestedInfo: reqInfo,
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
	imsi, err := decodeIdentityDigits(identityIMSI, arg.Imsi)
	if err != nil {
		return nil, fmt.Errorf("decoding IMSI: %w", err)
	}

	out := &ProvideSubscriberInfo{
		IMSI:          imsi,
		RequestedInfo: buildRequestedInfoFromWire(&arg.RequestedInfo),
	}

	// LMSI (optional, must be exactly 4 octets when present).
	if arg.Lmsi != nil {
		lmsi := []byte(*arg.Lmsi)

		out.LMSI = HexBytes(lmsi)
	}

	// CallPriority (optional, 0..15).
	if arg.CallPriority != nil {
		v := int64(*arg.CallPriority)

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
