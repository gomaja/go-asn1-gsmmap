package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// TS 29.002 V19.1.0 §17.7.1 PDP-Type and Ext-PDP-Type; TS 29.060 V19.0.0 §7.7.27 Figures35..39.
func validateSenderPDPTypes(base, ext []byte) error {
	if len(base) == 2 {
		switch [2]byte{base[0], base[1]} {
		case [2]byte{0xf0, 1}, [2]byte{0xf0, 2}, [2]byte{0xf1, 0x21}, [2]byte{0xf1, 0x57}:
		default:
			return ErrPDPTypeInvalid
		}
	}
	if len(ext) == 2 && (ext[0] != 0xf1 || ext[1] != 0x8d) {
		return ErrExtPDPTypeInvalid
	}
	return nil
}

// TS 29.002 V19.1.0 §17.7.1 PDP-Context/PDP-ContextInfo: paired IP families differ.
func validateSenderPDPAddressFamilies(base, ext []byte) error {
	if base == nil || ext == nil || len(base) < 1 || len(base) > 16 || len(ext) < 1 || len(ext) > 16 {
		return nil
	}
	if (len(base) == 4 && len(ext) == 16) || (len(base) == 16 && len(ext) == 4) {
		return nil
	}
	return ErrPDPAddressFamilyCombinationInvalid
}

// TS 29.002 V19.1.0 §17.7.1 PDP-ContextInfo permits only successive QoS extensions.
func validateSenderPDPContextInfo(c *gsm_map.PDPContextInfo) error {
	var ext []byte
	if c.ExtPdpType != nil {
		ext = *c.ExtPdpType
	}
	if err := validateSenderPDPTypes(c.PdpType, ext); err != nil {
		return err
	}
	if c.ExtPdpAddress != nil {
		if c.PdpAddress == nil {
			return ErrPDPContextInfoExtPDPAddressWithoutPDPAddress
		}
		if err := validateSenderPDPAddressFamilies(*c.PdpAddress, *c.ExtPdpAddress); err != nil {
			return err
		}
	}
	for _, p := range []struct {
		next, prev         bool
		field, predecessor string
	}{
		{c.Qos2Subscribed != nil, c.QosSubscribed != nil, "Qos2Subscribed", "QosSubscribed"},
		{c.Qos2Requested != nil, c.QosRequested != nil, "Qos2Requested", "QosRequested"},
		{c.Qos2Negotiated != nil, c.QosNegotiated != nil, "Qos2Negotiated", "QosNegotiated"},
		{c.Qos3Subscribed != nil, c.Qos2Subscribed != nil, "Qos3Subscribed", "Qos2Subscribed"},
		{c.Qos3Requested != nil, c.Qos2Requested != nil, "Qos3Requested", "Qos2Requested"},
		{c.Qos3Negotiated != nil, c.Qos2Negotiated != nil, "Qos3Negotiated", "Qos2Negotiated"},
		{c.Qos4Subscribed != nil, c.Qos3Subscribed != nil, "Qos4Subscribed", "Qos3Subscribed"},
		{c.Qos4Requested != nil, c.Qos3Requested != nil, "Qos4Requested", "Qos3Requested"},
		{c.Qos4Negotiated != nil, c.Qos3Negotiated != nil, "Qos4Negotiated", "Qos3Negotiated"},
	} {
		if p.next && !p.prev {
			return fmt.Errorf("%s requires %s: %w", p.field, p.predecessor, ErrPDPContextInfoQoSWithoutPredecessor)
		}
	}
	return nil
}

// TS 29.002 V19.1.0 §17.7.1 NumberPortabilityStatus restricts4/5 to ATI.
func validateNumberPortabilityOutsideATI(status *NumberPortabilityStatus) error {
	if status != nil && (*status == MnpOwnNumberNotPortedOut || *status == MnpForeignNumberPortedIn) {
		return ErrNumberPortabilityStatusATIOnly
	}
	return nil
}
func validateSubscriberInfoOutsideATI(s *SubscriberInfo) error {
	if s != nil && s.MnpInfoRes != nil {
		return validateNumberPortabilityOutsideATI(s.MnpInfoRes.NumberPortabilityStatus)
	}
	return nil
}

// TS 29.002 V19.1.0 §17.7.1 GeodeticInformation; Q.763 (1999) §3.88(c), Figure82.1.
func validateSenderGeodeticInformation(b []byte) error {
	if len(b) == 10 && b[1]&0x7f != 1 {
		return ErrGeodeticInformationShapeInvalid
	}
	return nil
}

// TS 29.002 V19.1.0 §17.7.8 TBCD-STRING uses filler only after the final odd digit.
func validateSenderRouteingNumber(b []byte) error {
	for i, v := range b {
		if v&0x0f == 0x0f || (v>>4 == 0x0f && i != len(b)-1) {
			return fmt.Errorf("octet %d: %w", i, ErrRouteingNumberTBCDInvalid)
		}
	}
	return nil
}
