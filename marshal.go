package gsmmap

import (
	"fmt"
)

// Marshal encodes SriSm into BER-encoded bytes.
func (s *SriSm) Marshal() ([]byte, error) {
	arg, err := convertSriSmToArg(s)
	if err != nil {
		return nil, fmt.Errorf("converting SriSm: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding RoutingInfoForSMArg: %w", err)
	}
	return data, nil
}

// Marshal encodes SriSmResp into BER-encoded bytes.
func (s *SriSmResp) Marshal() ([]byte, error) {
	res, err := convertSriSmRespToRes(s)
	if err != nil {
		return nil, fmt.Errorf("converting SriSmResp: %w", err)
	}
	data, err := res.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding RoutingInfoForSMRes: %w", err)
	}
	return data, nil
}

// Marshal encodes MtFsm into BER-encoded bytes.
func (m *MtFsm) Marshal() ([]byte, error) {
	arg, err := convertMtFsmToArg(m)
	if err != nil {
		return nil, fmt.Errorf("converting MtFsm: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding MTForwardSMArg: %w", err)
	}
	return data, nil
}

// Marshal encodes MtFsmResp into BER-encoded bytes.
func (r *MtFsmResp) Marshal() ([]byte, error) {
	res := convertMtFsmRespToRes(r)
	data, err := res.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding MTForwardSMRes: %w", err)
	}
	return data, nil
}

// Marshal encodes MoFsm into BER-encoded bytes.
func (m *MoFsm) Marshal() ([]byte, error) {
	arg, err := convertMoFsmToArg(m)
	if err != nil {
		return nil, fmt.Errorf("converting MoFsm: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding MOForwardSMArg: %w", err)
	}
	return data, nil
}

// Marshal encodes MoFsmResp into BER-encoded bytes.
func (r *MoFsmResp) Marshal() ([]byte, error) {
	res := convertMoFsmRespToRes(r)
	data, err := res.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding MOForwardSMRes: %w", err)
	}
	return data, nil
}

// Marshal encodes UpdateLocation into BER-encoded bytes.
func (u *UpdateLocation) Marshal() ([]byte, error) {
	arg, err := convertUpdateLocationToArg(u)
	if err != nil {
		return nil, fmt.Errorf("converting UpdateLocation: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding UpdateLocationArg: %w", err)
	}
	return data, nil
}

// Marshal encodes UpdateLocationRes into BER-encoded bytes.
func (u *UpdateLocationRes) Marshal() ([]byte, error) {
	res, err := convertUpdateLocationResToRes(u)
	if err != nil {
		return nil, fmt.Errorf("converting UpdateLocationRes: %w", err)
	}
	data, err := res.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding UpdateLocationRes: %w", err)
	}
	return data, nil
}

// Marshal encodes UpdateGprsLocation into BER-encoded bytes.
func (u *UpdateGprsLocation) Marshal() ([]byte, error) {
	arg, err := convertUpdateGprsLocationToArg(u)
	if err != nil {
		return nil, fmt.Errorf("converting UpdateGprsLocation: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding UpdateGprsLocationArg: %w", err)
	}
	return data, nil
}

// Marshal encodes UpdateGprsLocationRes into BER-encoded bytes.
func (u *UpdateGprsLocationRes) Marshal() ([]byte, error) {
	res, err := convertUpdateGprsLocationResToRes(u)
	if err != nil {
		return nil, fmt.Errorf("converting UpdateGprsLocationRes: %w", err)
	}
	data, err := res.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding UpdateGprsLocationRes: %w", err)
	}
	return data, nil
}

// Marshal encodes AnyTimeInterrogation into BER-encoded bytes.
func (a *AnyTimeInterrogation) Marshal() ([]byte, error) {
	arg, err := convertATIToArg(a)
	if err != nil {
		return nil, fmt.Errorf("converting AnyTimeInterrogation: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding AnyTimeInterrogationArg: %w", err)
	}
	return data, nil
}

// Marshal encodes AnyTimeInterrogationRes into BER-encoded bytes.
func (a *AnyTimeInterrogationRes) Marshal() ([]byte, error) {
	res, err := convertATIResToRes(a)
	if err != nil {
		return nil, fmt.Errorf("converting AnyTimeInterrogationRes: %w", err)
	}
	data, err := res.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding AnyTimeInterrogationRes: %w", err)
	}
	return data, nil
}

// Marshal encodes ProvideSubscriberInfo (opCode 70) into BER-encoded bytes.
func (p *ProvideSubscriberInfo) Marshal() ([]byte, error) {
	arg, err := convertProvideSubscriberInfoToArg(p)
	if err != nil {
		return nil, fmt.Errorf("converting ProvideSubscriberInfo: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding ProvideSubscriberInfoArg: %w", err)
	}
	return data, nil
}

// Marshal encodes ProvideSubscriberInfoRes (opCode 70) into BER-encoded bytes.
func (p *ProvideSubscriberInfoRes) Marshal() ([]byte, error) {
	res, err := convertProvideSubscriberInfoResToRes(p)
	if err != nil {
		return nil, fmt.Errorf("converting ProvideSubscriberInfoRes: %w", err)
	}
	data, err := res.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding ProvideSubscriberInfoRes: %w", err)
	}
	return data, nil
}

// Marshal encodes SriResp into BER-encoded bytes.
func (s *SriResp) Marshal() ([]byte, error) {
	res, err := convertSriRespToRes(s)
	if err != nil {
		return nil, fmt.Errorf("converting SriResp: %w", err)
	}
	data, err := res.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding SendRoutingInfoRes: %w", err)
	}
	return data, nil
}

// Marshal encodes Sri into BER-encoded bytes.
func (s *Sri) Marshal() ([]byte, error) {
	arg, err := convertSriToArg(s)
	if err != nil {
		return nil, fmt.Errorf("converting Sri: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding SendRoutingInfoArg: %w", err)
	}
	return data, nil
}

// Marshal encodes InformServiceCentre into BER-encoded bytes.
// InformServiceCentre (opCode 63) is a one-way MAP operation; no response is
// defined in 3GPP TS 29.002.
func (i *InformServiceCentre) Marshal() ([]byte, error) {
	arg, err := convertInformServiceCentreToArg(i)
	if err != nil {
		return nil, fmt.Errorf("converting InformServiceCentre: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding InformServiceCentreArg: %w", err)
	}
	return data, nil
}

// Marshal encodes AlertServiceCentre into BER-encoded bytes.
// AlertServiceCentre (opCode 64) per 3GPP TS 29.002: Invoke carries the
// arg; the response is an empty RETURN RESULT (no parameters), so no
// response Marshal is defined on the public API.
func (a *AlertServiceCentre) Marshal() ([]byte, error) {
	arg, err := convertAlertServiceCentreToArg(a)
	if err != nil {
		return nil, fmt.Errorf("converting AlertServiceCentre: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding AlertServiceCentreArg: %w", err)
	}
	return data, nil
}

// Marshal encodes PurgeMS (opCode 67) into BER-encoded bytes.
func (p *PurgeMS) Marshal() ([]byte, error) {
	arg, err := convertPurgeMSToArg(p)
	if err != nil {
		return nil, fmt.Errorf("converting PurgeMS: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding PurgeMSArg: %w", err)
	}
	return data, nil
}

// Marshal encodes PurgeMSRes (opCode 67) into BER-encoded bytes.
func (r *PurgeMSRes) Marshal() ([]byte, error) {
	res := convertPurgeMSResToWire(r)
	data, err := res.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding PurgeMSRes: %w", err)
	}
	return data, nil
}

// Marshal encodes SendAuthenticationInfo (opCode 56) into BER-encoded bytes.
func (s *SendAuthenticationInfo) Marshal() ([]byte, error) {
	arg, err := convertSendAuthenticationInfoToArg(s)
	if err != nil {
		return nil, fmt.Errorf("converting SendAuthenticationInfo: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding SendAuthenticationInfoArg: %w", err)
	}
	return data, nil
}

// Marshal encodes SendAuthenticationInfoRes (opCode 56) into BER-encoded bytes.
func (s *SendAuthenticationInfoRes) Marshal() ([]byte, error) {
	res, err := convertSendAuthenticationInfoResToRes(s)
	if err != nil {
		return nil, fmt.Errorf("converting SendAuthenticationInfoRes: %w", err)
	}
	data, err := res.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding SendAuthenticationInfoRes: %w", err)
	}
	return data, nil
}

// Marshal encodes CancelLocation (opCode 3) into BER-encoded bytes.
func (c *CancelLocation) Marshal() ([]byte, error) {
	arg, err := convertCancelLocationToArg(c)
	if err != nil {
		return nil, fmt.Errorf("converting CancelLocation: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding CancelLocationArg: %w", err)
	}
	return data, nil
}

// Marshal encodes CancelLocationRes (opCode 3) into BER-encoded bytes.
func (r *CancelLocationRes) Marshal() ([]byte, error) {
	res := convertCancelLocationResToWire(r)
	data, err := res.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding CancelLocationRes: %w", err)
	}
	return data, nil
}

// Marshal encodes InsertSubscriberDataArg (opCode 7) into BER-encoded bytes.
func (a *InsertSubscriberDataArg) Marshal() ([]byte, error) {
	arg, err := convertInsertSubscriberDataArgToWire(a)
	if err != nil {
		return nil, fmt.Errorf("converting InsertSubscriberDataArg: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding InsertSubscriberDataArg: %w", err)
	}
	return data, nil
}

// Marshal encodes InsertSubscriberDataRes (opCode 7) into BER-encoded bytes.
func (r *InsertSubscriberDataRes) Marshal() ([]byte, error) {
	res, err := convertInsertSubscriberDataResToWire(r)
	if err != nil {
		return nil, fmt.Errorf("converting InsertSubscriberDataRes: %w", err)
	}
	data, err := res.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding InsertSubscriberDataRes: %w", err)
	}
	return data, nil
}

// Marshal encodes ProvideSubscriberLocationArg (opCode 83) into
// BER-encoded bytes.
func (a *ProvideSubscriberLocationArg) Marshal() ([]byte, error) {
	arg, err := convertProvideSubscriberLocationArgToWire(a)
	if err != nil {
		return nil, fmt.Errorf("converting ProvideSubscriberLocationArg: %w", err)
	}
	data, err := arg.MarshalBER()
	if err != nil {
		return nil, fmt.Errorf("encoding ProvideSubscriberLocationArg: %w", err)
	}
	return data, nil
}
