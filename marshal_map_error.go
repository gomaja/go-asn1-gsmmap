package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
)

// Marshal encodes AbsentSubscriberSMParam as a ReturnError parameter.
func (p *AbsentSubscriberSMParam) Marshal() ([]byte, error) {
	w, err := convertAbsentSubscriberSMParamToWire(p)
	if err != nil {
		return nil, fmt.Errorf("converting AbsentSubscriberSMParam: %w", err)
	}
	data, err := w.MarshalBER(ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding AbsentSubscriberSMParam: %w", err)
	}
	return data, nil
}

// Marshal encodes UnknownSubscriberParam as a ReturnError parameter.
func (p *UnknownSubscriberParam) Marshal() ([]byte, error) {
	w, err := convertUnknownSubscriberParamToWire(p)
	if err != nil {
		return nil, fmt.Errorf("converting UnknownSubscriberParam: %w", err)
	}
	data, err := w.MarshalBER(ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding UnknownSubscriberParam: %w", err)
	}
	return data, nil
}

// Marshal encodes CallBarredParam as a ReturnError parameter.
func (p *CallBarredParam) Marshal() ([]byte, error) {
	w, err := convertCallBarredParamToWire(p)
	if err != nil {
		return nil, fmt.Errorf("converting CallBarredParam: %w", err)
	}
	data, err := w.MarshalBER(ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding CallBarredParam: %w", err)
	}
	return data, nil
}

// Marshal encodes SystemFailureParam as a ReturnError parameter.
func (p *SystemFailureParam) Marshal() ([]byte, error) {
	w, err := convertSystemFailureParamToWire(p)
	if err != nil {
		return nil, fmt.Errorf("converting SystemFailureParam: %w", err)
	}
	data, err := w.MarshalBER(ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding SystemFailureParam: %w", err)
	}
	return data, nil
}

// Marshal encodes RoamingNotAllowedParam as a ReturnError parameter.
func (p *RoamingNotAllowedParam) Marshal() ([]byte, error) {
	w, err := convertRoamingNotAllowedParamToWire(p)
	if err != nil {
		return nil, fmt.Errorf("converting RoamingNotAllowedParam: %w", err)
	}
	data, err := w.MarshalBER(ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding RoamingNotAllowedParam: %w", err)
	}
	return data, nil
}

// Marshal encodes UnauthorizedRequestingNetworkParam as a ReturnError parameter.
func (p *UnauthorizedRequestingNetworkParam) Marshal() ([]byte, error) {
	w, err := convertUnauthorizedRequestingNetworkParamToWire(p)
	if err != nil {
		return nil, fmt.Errorf("converting UnauthorizedRequestingNetworkParam: %w", err)
	}
	data, err := w.MarshalBER(ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding UnauthorizedRequestingNetworkParam: %w", err)
	}
	return data, nil
}

// Marshal encodes FacilityNotSupParam as a ReturnError parameter.
func (p *FacilityNotSupParam) Marshal() ([]byte, error) {
	w, err := convertFacilityNotSupParamToWire(p)
	if err != nil {
		return nil, fmt.Errorf("converting FacilityNotSupParam: %w", err)
	}
	data, err := w.MarshalBER(ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding FacilityNotSupParam: %w", err)
	}
	return data, nil
}

// Marshal encodes TeleservNotProvParam as a ReturnError parameter.
func (p *TeleservNotProvParam) Marshal() ([]byte, error) {
	w, err := convertTeleservNotProvParamToWire(p)
	if err != nil {
		return nil, fmt.Errorf("converting TeleservNotProvParam: %w", err)
	}
	data, err := w.MarshalBER(ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding TeleservNotProvParam: %w", err)
	}
	return data, nil
}

// Marshal encodes DataMissingParam as a ReturnError parameter.
func (p *DataMissingParam) Marshal() ([]byte, error) {
	w, err := convertDataMissingParamToWire(p)
	if err != nil {
		return nil, fmt.Errorf("converting DataMissingParam: %w", err)
	}
	data, err := w.MarshalBER(ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding DataMissingParam: %w", err)
	}
	return data, nil
}

// Marshal encodes AbsentSubscriberParam as a ReturnError parameter.
func (p *AbsentSubscriberParam) Marshal() ([]byte, error) {
	w, err := convertAbsentSubscriberParamToWire(p)
	if err != nil {
		return nil, fmt.Errorf("converting AbsentSubscriberParam: %w", err)
	}
	data, err := w.MarshalBER(ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding AbsentSubscriberParam: %w", err)
	}
	return data, nil
}

// Marshal encodes IllegalSubscriberParam as a ReturnError parameter.
func (p *IllegalSubscriberParam) Marshal() ([]byte, error) {
	w, err := convertIllegalSubscriberParamToWire(p)
	if err != nil {
		return nil, fmt.Errorf("converting IllegalSubscriberParam: %w", err)
	}
	data, err := w.MarshalBER(ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding IllegalSubscriberParam: %w", err)
	}
	return data, nil
}

// Marshal encodes IllegalEquipmentParam as a ReturnError parameter.
func (p *IllegalEquipmentParam) Marshal() ([]byte, error) {
	w, err := convertIllegalEquipmentParamToWire(p)
	if err != nil {
		return nil, fmt.Errorf("converting IllegalEquipmentParam: %w", err)
	}
	data, err := w.MarshalBER(ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding IllegalEquipmentParam: %w", err)
	}
	return data, nil
}

// Marshal encodes UnexpectedDataParam as a ReturnError parameter.
func (p *UnexpectedDataParam) Marshal() ([]byte, error) {
	w, err := convertUnexpectedDataParamToWire(p)
	if err != nil {
		return nil, fmt.Errorf("converting UnexpectedDataParam: %w", err)
	}
	data, err := w.MarshalBER(ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding UnexpectedDataParam: %w", err)
	}
	return data, nil
}
