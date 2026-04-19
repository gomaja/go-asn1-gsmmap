// BIT STRING <-> struct-of-bools helpers shared across operations.

package gsmmap

import (
	"github.com/gomaja/go-asn1/runtime"
)

func convertCamelPhasesToBitString(cp *SupportedCamelPhases) runtime.BitString {
	var b byte
	bitLen := 1
	if cp.Phase1 {
		b |= 0x80
	}
	if cp.Phase2 {
		b |= 0x40
		bitLen = 2
	}
	if cp.Phase3 {
		b |= 0x20
		bitLen = 3
	}
	if cp.Phase4 {
		b |= 0x10
		bitLen = 4
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: bitLen}
}

func convertBitStringToCamelPhases(bs runtime.BitString) *SupportedCamelPhases {
	cp := &SupportedCamelPhases{}
	if bs.BitLength > 0 {
		cp.Phase1 = bs.Has(0)
	}
	if bs.BitLength > 1 {
		cp.Phase2 = bs.Has(1)
	}
	if bs.BitLength > 2 {
		cp.Phase3 = bs.Has(2)
	}
	if bs.BitLength > 3 {
		cp.Phase4 = bs.Has(3)
	}
	return cp
}

func convertLCSCapsToBitString(lcs *SupportedLCSCapabilitySets) runtime.BitString {
	var b byte
	bitLen := 2 // minimum per spec
	if lcs.LcsCapabilitySet1 {
		b |= 0x80
	}
	if lcs.LcsCapabilitySet2 {
		b |= 0x40
	}
	if lcs.LcsCapabilitySet3 {
		b |= 0x20
		bitLen = 3
	}
	if lcs.LcsCapabilitySet4 {
		b |= 0x10
		bitLen = 4
	}
	if lcs.LcsCapabilitySet5 {
		b |= 0x08
		bitLen = 5
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: bitLen}
}

func convertBitStringToLCSCaps(bs runtime.BitString) *SupportedLCSCapabilitySets {
	lcs := &SupportedLCSCapabilitySets{}
	if bs.BitLength > 0 {
		lcs.LcsCapabilitySet1 = bs.Has(0)
	}
	if bs.BitLength > 1 {
		lcs.LcsCapabilitySet2 = bs.Has(1)
	}
	if bs.BitLength > 2 {
		lcs.LcsCapabilitySet3 = bs.Has(2)
	}
	if bs.BitLength > 3 {
		lcs.LcsCapabilitySet4 = bs.Has(3)
	}
	if bs.BitLength > 4 {
		lcs.LcsCapabilitySet5 = bs.Has(4)
	}
	return lcs
}

func convertRequestedNodesToBitString(rn *RequestedNodes) runtime.BitString {
	var b byte
	bitLen := 1
	if rn.MME {
		b |= 0x80
	}
	if rn.SGSN {
		b |= 0x40
		bitLen = 2
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: bitLen}
}

func convertBitStringToRequestedNodes(bs runtime.BitString) *RequestedNodes {
	rn := &RequestedNodes{}
	if bs.BitLength > 0 {
		rn.MME = bs.Has(0)
	}
	if bs.BitLength > 1 {
		rn.SGSN = bs.Has(1)
	}
	return rn
}

// AllowedServices: 2 bits (bit 0 = first, bit 1 = second).
func convertAllowedServicesToBitString(a *AllowedServicesFlags) runtime.BitString {
	var b byte
	if a.FirstServiceAllowed {
		b |= 0x80
	}
	if a.SecondServiceAllowed {
		b |= 0x40
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: 2}
}

func convertBitStringToAllowedServices(bs runtime.BitString) *AllowedServicesFlags {
	a := &AllowedServicesFlags{}
	if bs.BitLength > 0 {
		a.FirstServiceAllowed = bs.Has(0)
	}
	if bs.BitLength > 1 {
		a.SecondServiceAllowed = bs.Has(1)
	}
	return a
}

// SuppressMTSS: 2 bits (bit 0 = suppressCUG, bit 1 = suppressCCBS), min size 2.
func convertSuppressMTSSToBitString(s *SuppressMTSSFlags) runtime.BitString {
	var b byte
	if s.SuppressCUG {
		b |= 0x80
	}
	if s.SuppressCCBS {
		b |= 0x40
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: 2}
}

func convertBitStringToSuppressMTSS(bs runtime.BitString) *SuppressMTSSFlags {
	s := &SuppressMTSSFlags{}
	if bs.BitLength > 0 {
		s.SuppressCUG = bs.Has(0)
	}
	if bs.BitLength > 1 {
		s.SuppressCCBS = bs.Has(1)
	}
	return s
}

// OfferedCamel4CSIs: 7 bits per 3GPP TS 29.002.
// Bit order: 0=o-CSI, 1=d-CSI, 2=vt-CSI, 3=t-CSI, 4=mt-sms-CSI, 5=mg-CSI, 6=psi-enhancements.
func convertOfferedCamel4CSIsToBitString(o *OfferedCamel4CSIs) runtime.BitString {
	var b byte
	if o.OCSI {
		b |= 1 << 7
	}
	if o.DCSI {
		b |= 1 << 6
	}
	if o.VTCSI {
		b |= 1 << 5
	}
	if o.TCSI {
		b |= 1 << 4
	}
	if o.MTSMSCSI {
		b |= 1 << 3
	}
	if o.MGCSI {
		b |= 1 << 2
	}
	if o.PsiEnhancements {
		b |= 1 << 1
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: 7}
}

func convertBitStringToOfferedCamel4CSIs(bs runtime.BitString) *OfferedCamel4CSIs {
	o := &OfferedCamel4CSIs{}
	if bs.BitLength > 0 {
		o.OCSI = bs.Has(0)
	}
	if bs.BitLength > 1 {
		o.DCSI = bs.Has(1)
	}
	if bs.BitLength > 2 {
		o.VTCSI = bs.Has(2)
	}
	if bs.BitLength > 3 {
		o.TCSI = bs.Has(3)
	}
	if bs.BitLength > 4 {
		o.MTSMSCSI = bs.Has(4)
	}
	if bs.BitLength > 5 {
		o.MGCSI = bs.Has(5)
	}
	if bs.BitLength > 6 {
		o.PsiEnhancements = bs.Has(6)
	}
	return o
}

// SupportedRATTypes: bit 0=utran, 1=geran, 2=gan, 3=i-hspa-evolution, 4=e-utran.
func convertSupportedRATTypesToBitString(r *SupportedRATTypes) runtime.BitString {
	var b byte
	if r.UTRAN {
		b |= 0x80
	}
	if r.GERAN {
		b |= 0x40
	}
	if r.GAN {
		b |= 0x20
	}
	if r.IHSPAEvolution {
		b |= 0x10
	}
	if r.EUTRAN {
		b |= 0x08
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: 5}
}

func convertBitStringToSupportedRATTypes(bs runtime.BitString) *SupportedRATTypes {
	r := &SupportedRATTypes{}
	if bs.BitLength > 0 {
		r.UTRAN = bs.Has(0)
	}
	if bs.BitLength > 1 {
		r.GERAN = bs.Has(1)
	}
	if bs.BitLength > 2 {
		r.GAN = bs.Has(2)
	}
	if bs.BitLength > 3 {
		r.IHSPAEvolution = bs.Has(3)
	}
	if bs.BitLength > 4 {
		r.EUTRAN = bs.Has(4)
	}
	return r
}
