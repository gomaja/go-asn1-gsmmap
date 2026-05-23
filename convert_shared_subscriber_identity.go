// convert_shared_subscriber_identity.go
//
// Shared converter for the SubscriberIdentity CHOICE (IMSI or MSISDN)
// per TS 29.002 MAP-CommonDataTypes.asn. Used by SendRoutingInfoForLCS
// (opCode 85) and AnyTimeInterrogation (opCode 71). MSISDN carries its
// AddressString Nature/Plan (defaulting to International/ISDN when zero),
// so a non-international MSISDN survives a decode→encode round-trip.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// convertSubscriberIdentityToWire encodes the SubscriberIdentity CHOICE.
// Exactly one of IMSI or MSISDN must be set.
func convertSubscriberIdentityToWire(s SubscriberIdentity) (gsm_map.SubscriberIdentity, error) {
	imsiSet := s.IMSI != ""
	msisdnSet := s.MSISDN != ""
	switch {
	case !imsiSet && !msisdnSet:
		return gsm_map.SubscriberIdentity{}, ErrSubscriberIdentityNoAlt
	case imsiSet && msisdnSet:
		return gsm_map.SubscriberIdentity{}, ErrSubscriberIdentityMultipleAlts
	}

	if imsiSet {
		imsiBytes, err := tbcd.Encode(s.IMSI)
		if err != nil {
			return gsm_map.SubscriberIdentity{}, fmt.Errorf(errEncodingIMSI, err)
		}
		return gsm_map.NewSubscriberIdentityImsi(gsm_map.IMSI(imsiBytes)), nil
	}
	msisdnBytes, err := encodeAddressField(s.MSISDN, s.MSISDNNature, s.MSISDNPlan)
	if err != nil {
		return gsm_map.SubscriberIdentity{}, fmt.Errorf("encoding MSISDN: %w", err)
	}
	return gsm_map.NewSubscriberIdentityMsisdn(gsm_map.ISDNAddressString(msisdnBytes)), nil
}

// convertWireToSubscriberIdentity decodes the SubscriberIdentity CHOICE.
// A present-but-empty decoded value is rejected so the string-based
// public type round-trips faithfully.
func convertWireToSubscriberIdentity(w gsm_map.SubscriberIdentity) (SubscriberIdentity, error) {
	var out SubscriberIdentity
	switch w.Choice {
	case gsm_map.SubscriberIdentityChoiceImsi:
		if w.Imsi == nil {
			return out, ErrSubscriberIdentityUnknownChoice
		}
		imsi, err := tbcd.Decode(*w.Imsi)
		if err != nil {
			return out, fmt.Errorf("decoding SubscriberIdentity.IMSI: %w", err)
		}
		if imsi == "" {
			return out, ErrSubscriberIdentityIMSIDecodedEmpty
		}
		out.IMSI = imsi
	case gsm_map.SubscriberIdentityChoiceMsisdn:
		if w.Msisdn == nil {
			return out, ErrSubscriberIdentityUnknownChoice
		}
		msisdn, nature, plan, err := decodeAddressField(*w.Msisdn)
		if err != nil {
			return out, fmt.Errorf("decoding SubscriberIdentity.MSISDN: %w", err)
		}
		if msisdn == "" {
			return out, ErrSubscriberIdentityMSISDNDecodedEmpty
		}
		out.MSISDN = msisdn
		out.MSISDNNature = nature
		out.MSISDNPlan = plan
	default:
		return out, fmt.Errorf("SubscriberIdentity choice=%d: %w", w.Choice, ErrSubscriberIdentityUnknownChoice)
	}
	return out, nil
}
