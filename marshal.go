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
