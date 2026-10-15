package gsmmap

import (
	"bytes"
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-gad"
	"github.com/gomaja/go-tbcd"
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
	if err := validateExtPDPAddressPresence(c.PdpAddress != nil, c.ExtPdpAddress != nil); err != nil {
		return fmt.Errorf("PDPContextInfo.ExtPdpAddress: %w", err)
	}
	if c.ExtPdpAddress != nil {
		if err := validateSenderPDPAddressFamilies(*c.PdpAddress, *c.ExtPdpAddress); err != nil {
			return err
		}
	}
	for _, qos := range []struct {
		present [4]bool
		fields  [4]string
	}{
		{[4]bool{c.QosSubscribed != nil, c.Qos2Subscribed != nil, c.Qos3Subscribed != nil, c.Qos4Subscribed != nil}, [4]string{"PDPContextInfo.QosSubscribed", "PDPContextInfo.Qos2Subscribed", "PDPContextInfo.Qos3Subscribed", "PDPContextInfo.Qos4Subscribed"}},
		{[4]bool{c.QosRequested != nil, c.Qos2Requested != nil, c.Qos3Requested != nil, c.Qos4Requested != nil}, [4]string{"PDPContextInfo.QosRequested", "PDPContextInfo.Qos2Requested", "PDPContextInfo.Qos3Requested", "PDPContextInfo.Qos4Requested"}},
		{[4]bool{c.QosNegotiated != nil, c.Qos2Negotiated != nil, c.Qos3Negotiated != nil, c.Qos4Negotiated != nil}, [4]string{"PDPContextInfo.QosNegotiated", "PDPContextInfo.Qos2Negotiated", "PDPContextInfo.Qos3Negotiated", "PDPContextInfo.Qos4Negotiated"}},
	} {
		if err := validateExtQoSHierarchy(qos.present, qos.fields); err != nil {
			return err
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

// validateSenderGeographicalInformation checks a GeographicalInformation
// before sending: 3GPP TS 29.002 V19.1.0 §17.7.1 allows only an ellipsoid
// point with uncertainty circle, coded per 3GPP TS 23.032 V19.0.0 §7.3.2.
// The codec checks SIZE (8).
func validateSenderGeographicalInformation(b []byte) error {
	if len(b) != 8 {
		return nil
	}
	if shape := gad.Shape(b[0] >> 4); shape != gad.ShapeEllipsoidPointUncertaintyCircle {
		return fmt.Errorf("%w: %v", ErrGeographicalInformationInvalid, shape)
	}
	if err := gad.Validate(b); err != nil {
		return fmt.Errorf("%w: %w", ErrGeographicalInformationInvalid, err)
	}
	return nil
}

// validateSenderRouteingNumber checks that a RouteingNumber is the TBCD
// coding of its digits: TS 29.002 V19.1.0 §17.7.8 TBCD-STRING uses filler
// only after the final odd digit, so neither inside the digits nor as a
// trailing filler octet.
func validateSenderRouteingNumber(b []byte) error {
	digits, err := tbcd.Decode(b)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRouteingNumberTBCDInvalid, err)
	}
	// Decode returns digits of the TBCD alphabet only, which Encode accepts.
	if enc, _ := tbcd.Encode(digits); !bytes.Equal(enc, b) {
		return fmt.Errorf("%w: trailing filler octet", ErrRouteingNumberTBCDInvalid)
	}
	return nil
}
