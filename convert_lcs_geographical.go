package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-gad"
)

// receiveGeographicalInformation applies the exception handling of 3GPP TS
// 29.002 V19.1.0 §17.7.13 to a received location estimate (ext) and
// additional location estimate (add):
//   - an Add received together with a valid Ext is discarded, valid or not;
//   - a one-octet Ext is discarded beside an Add, and is invalid data
//     without one;
//   - any other invalid Ext, and an invalid Add without a valid Ext, are
//     invalid data.
//
// With ignoreExt, set by deferredmt-lrResponseIndicator in a PSL result,
// "locationEstimate is ignored": it is dropped and never invalid data.
func receiveGeographicalInformation(ext []byte, add *gsm_map.AddGeographicalInformation, ignoreExt bool) (ExtGeographicalInformation, AddGeographicalInformation, error) {
	switch {
	case len(ext) > 1:
		if err := checkGeographicalInformation(ext, false); err == nil {
			add = nil
		} else if !ignoreExt {
			return nil, nil, fmt.Errorf("%w: %w", ErrExtGeographicalInformationInvalidData, err)
		}
	case len(ext) == 1 && add == nil && !ignoreExt:
		return nil, nil, ErrExtGeographicalInformationInvalidData
	}
	if ignoreExt || len(ext) == 1 {
		ext = nil
	}
	if add == nil {
		return ExtGeographicalInformation(ext), nil, nil
	}
	if err := checkGeographicalInformation(*add, true); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrAddGeographicalInformationInvalidData, err)
	}
	return ExtGeographicalInformation(ext), AddGeographicalInformation(*add), nil
}

// checkGeographicalInformation returns why d is not a valid
// Ext-GeographicalInformation, or with add an Add-GeographicalInformation:
// 3GPP TS 29.002 V19.1.0 §17.7.13 allows cases (a) to (e) in Ext and every
// shape in Add, each coded per 3GPP TS 23.032 V19.0.0 §7.3, which go-gad
// checks.
func checkGeographicalInformation(d []byte, add bool) error {
	if err := gad.Validate(d); err != nil {
		return err
	}
	if add {
		return nil
	}
	switch shape := gad.Shape(d[0] >> 4); shape {
	case gad.ShapeEllipsoidPointUncertaintyCircle, gad.ShapeEllipsoidPointUncertaintyEllipse,
		gad.ShapeEllipsoidPointAltitudeUncertaintyEllipsoid, gad.ShapeEllipsoidArc, gad.ShapeEllipsoidPoint:
		return nil
	default:
		return fmt.Errorf("%v is not a shape of Ext-GeographicalInformation", shape)
	}
}

// receiveVelocityEstimate applies 3GPP TS 29.002 V19.1.0 §17.7.13 to a
// received VelocityEstimate: "Any other value in octet 1 shall be treated
// as invalid", other than the four velocity types of 3GPP TS 23.032
// V19.0.0 §8.6.
func receiveVelocityEstimate(v *gsm_map.VelocityEstimate) (VelocityEstimate, error) {
	if v == nil {
		return nil, nil
	}
	if gad.VelocityType((*v)[0]>>4) > gad.VelocityHorizontalVerticalUncertainty {
		return nil, ErrVelocityEstimateInvalidData
	}
	return VelocityEstimate(*v), nil
}

// TS 29.002 V19.1.0 §17.7.13; TS 23.032 V19.0.0 §7.3 define sender geography.
func validateSendGeographicalInformation(ext, add []byte, ignoreExt bool) error {
	// Values outside ASN.1 SIZE bounds are left to the BER codec.
	if !ignoreExt && len(ext) > 1 && len(ext) <= int(gsm_map.MaxExtGeographicalInformation) {
		if err := checkGeographicalInformation(ext, false); err != nil {
			return fmt.Errorf("LocationEstimate: %w: %w", ErrExtGeographicalInformationInvalid, err)
		}
	}
	if len(add) > 0 && len(add) <= int(gsm_map.MaxAddGeographicalInformation) {
		if err := checkGeographicalInformation(add, true); err != nil {
			return fmt.Errorf("AddLocationEstimate: %w: %w", ErrAddGeographicalInformationInvalid, err)
		}
	}
	return nil
}

// locationEstimateBesideAdd returns the locationEstimate sent beside a
// non-empty add. 3GPP TS 29.002 V19.1.0 §7.6.11.11 composes a location
// estimate of its type of shape and that shape's encoding, and puts the
// encoding of a shape locationEstimate does not carry in add-LocationEstimate;
// locationEstimate is then one octet (§§13A.2.3, 13A.3.3), which the
// receiver discards (§17.7.13). An empty ext is sent as add's type of shape.
// An add outside its SIZE is left to the BER codec.
func locationEstimateBesideAdd(ext, add []byte) ([]byte, error) {
	switch {
	case len(ext) == 0:
		return []byte{add[0] & 0xF0}, nil
	case len(ext) == 1 && len(add) <= int(gsm_map.MaxAddGeographicalInformation) && ext[0]>>4 != add[0]>>4:
		return nil, ErrLocationEstimateShapeMismatch
	}
	return ext, nil
}

// validateSendVelocityEstimate checks a VelocityEstimate before sending: one
// of the four velocity types 3GPP TS 29.002 V19.1.0 §17.7.13 allows, coded
// per 3GPP TS 23.032 V19.0.0 §8, which go-gad checks. The codec checks
// SIZE (4..7).
func validateSendVelocityEstimate(v []byte) error {
	if len(v) < 4 || len(v) > 7 {
		return nil
	}
	if err := gad.ValidateVelocity(v); err != nil {
		return fmt.Errorf("%w: %w", ErrVelocityEstimateInvalid, err)
	}
	return nil
}

// TS 29.002 V19.1.0 §17.7.13: either IMSI or MSISDN identifies the subscriber.
func validateLCSSubscriberIdentity(imsi, msisdn string) error {
	if imsi == "" && msisdn == "" {
		return ErrLCSSubscriberIdentityMissing
	}
	return nil
}
