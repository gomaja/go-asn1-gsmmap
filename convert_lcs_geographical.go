package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
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
	if len(ext) > 1 && validGeographicalInformation(ext, false) {
		add = nil
	} else if !ignoreExt && (len(ext) > 1 || len(ext) == 1 && add == nil) {
		return nil, nil, ErrExtGeographicalInformationInvalidData
	}
	if ignoreExt || len(ext) == 1 {
		ext = nil
	}
	if add == nil {
		return ExtGeographicalInformation(ext), nil, nil
	}
	if !validGeographicalInformation(*add, true) {
		return nil, nil, ErrAddGeographicalInformationInvalidData
	}
	return ExtGeographicalInformation(ext), AddGeographicalInformation(*add), nil
}

// TS 23.032 V19.0.0 §7.3 figures define lengths and explicitly zero coding bits.
// Confidence 101..127 does not mandate rejection (§6.5).
func validGeographicalInformation(d []byte, add bool) bool {
	if len(d) == 0 {
		return false
	}
	switch d[0] >> 4 {
	case 0: // §7.3.1, figure 5.
		return len(d) == 7
	case 1: // §7.3.2, figure 6.
		return len(d) == 8 && d[7]&0x80 == 0
	case 3: // §7.3.3, figure 6a.
		return len(d) == 11 && (d[7]|d[8]|d[10])&0x80 == 0 && d[9] <= 179
	case 5: // §7.3.4, figure 7.
		n := int(d[0] & 0xf)
		return add && n >= 3 && n <= 15 && len(d) == 1+6*n
	case 8: // §7.3.5, figure 8; altitude includes a direction bit.
		return add && len(d) == 9
	case 9: // §7.3.6, figure 9.
		return len(d) == 14 && (d[9]|d[10]|d[12]|d[13])&0x80 == 0 && d[11] <= 179
	case 10: // §7.3.7, figure 10; arc angles follow §6.7.
		return len(d) == 13 && (d[9]|d[12])&0x80 == 0 && d[10] <= 179 && d[11] <= 179
	case 11: // §7.3.3a, figure 7.3.3a-1.
		return add && len(d) == 13 && d[11] <= 179 && d[12]&0x80 == 0
	case 13: // §7.3.3b, figure 7.3.3b-1; bit 8 selects uncertainty range.
		return add && len(d) == 13 && d[11] <= 179
	case 12, 14: // §§7.3.6a,7.3.6b, figures 7.3.6a-1 and 7.3.6b-1.
		if !add || len(d) != 18 || d[9]&0xc0 != 0 || d[14] > 179 {
			return false
		}
		if d[0]>>4 == 12 && (d[15]|d[17])&0x80 != 0 {
			return false
		}
		// §6.3a: signed 22-bit altitude codes range from -64000 to 1280000.
		altitude := int32(d[9]&0x3f)<<16 | int32(d[10])<<8 | int32(d[11])
		if altitude&0x200000 != 0 {
			altitude -= 1 << 22
		}
		return altitude >= -64000 && altitude <= 1280000
	default:
		return false
	}
}

// TS 29.002 V19.1.0 §17.7.13; TS 23.032 V19.0.0 §§8.5,8.6.
func receiveVelocityEstimate(v *gsm_map.VelocityEstimate) (VelocityEstimate, error) {
	if v == nil {
		return nil, nil
	}
	if (*v)[0]>>4 > 3 {
		return nil, ErrVelocityEstimateInvalidData
	}
	return VelocityEstimate(*v), nil
}

// TS 29.002 V19.1.0 §17.7.13; TS 23.032 V19.0.0 §7.3 define sender geography.
func validateSendGeographicalInformation(ext, add []byte, ignoreExt bool) error {
	// Values outside ASN.1 SIZE bounds are left to the BER codec.
	if !ignoreExt && len(ext) > 1 && len(ext) <= 20 && !validGeographicalInformation(ext, false) {
		return fmt.Errorf("LocationEstimate: %w", ErrExtGeographicalInformationInvalid)
	}
	if len(add) > 0 && len(add) <= 91 && !validGeographicalInformation(add, true) {
		return fmt.Errorf("AddLocationEstimate: %w", ErrAddGeographicalInformationInvalid)
	}
	return nil
}

// TS 29.002 V19.1.0 §17.7.13; TS 23.032 V19.0.0 §8.6 define four velocity types.
func validateSendVelocityEstimate(v []byte) error {
	if len(v) > 0 && v[0]>>4 > 3 {
		return ErrVelocityEstimateTypeInvalid
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
