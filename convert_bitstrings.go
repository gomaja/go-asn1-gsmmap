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

// packBits is a shared helper for wide BIT STRINGs. It walks the set-bit
// table (in ASN.1 bit-position order), packs each bit into the encoded
// byte stream MSB-first, and returns Bytes padded to cover minBits plus
// a BitLength equal to max(minBits, one past the highest set bit). This
// matches the encoding convention used by go-asn1's BER codec.
func packBits(set []bool, minBits int) runtime.BitString {
	// highest-set bit index, or -1 if none set
	high := -1
	for i, v := range set {
		if v {
			high = i
		}
	}
	bitLen := minBits
	if h := high + 1; h > bitLen {
		bitLen = h
	}
	nBytes := (bitLen + 7) / 8
	if nBytes == 0 {
		return runtime.BitString{}
	}
	out := make([]byte, nBytes)
	for i, v := range set {
		if v && i < bitLen {
			out[i/8] |= 1 << (7 - (i % 8))
		}
	}
	return runtime.BitString{Bytes: out, BitLength: bitLen}
}

// ODBGeneralData: 29 named bits (SIZE 15..32) per MAP-MS-DataTypes.asn:1776.
func convertODBGeneralDataToBitString(o *ODBGeneralData) runtime.BitString {
	bits := []bool{
		o.AllOGCallsBarred,                                                // 0
		o.InternationalOGCallsBarred,                                      // 1
		o.InternationalOGCallsNotToHPLMNCountryBarred,                     // 2
		o.PremiumRateInformationOGCallsBarred,                             // 3
		o.PremiumRateEntertainmentOGCallsBarred,                           // 4
		o.SSAccessBarred,                                                  // 5
		o.InterzonalOGCallsBarred,                                         // 6
		o.InterzonalOGCallsNotToHPLMNCountryBarred,                        // 7
		o.InterzonalOGCallsAndInternationalOGCallsNotToHPLMNCountryBarred, // 8
		o.AllECTBarred,                                                    // 9
		o.ChargeableECTBarred,                                             // 10
		o.InternationalECTBarred,                                          // 11
		o.InterzonalECTBarred,                                             // 12
		o.DoublyChargeableECTBarred,                                       // 13
		o.MultipleECTBarred,                                               // 14
		o.AllPacketOrientedServicesBarred,                                 // 15
		o.RoamerAccessToHPLMNAPBarred,                                     // 16
		o.RoamerAccessToVPLMNAPBarred,                                     // 17
		o.RoamingOutsidePLMNOGCallsBarred,                                 // 18
		o.AllICCallsBarred,                                                // 19
		o.RoamingOutsidePLMNICCallsBarred,                                 // 20
		o.RoamingOutsidePLMNICountryICCallsBarred,                         // 21
		o.RoamingOutsidePLMNBarred,                                        // 22
		o.RoamingOutsidePLMNCountryBarred,                                 // 23
		o.RegistrationAllCFBarred,                                         // 24
		o.RegistrationCFNotToHPLMNBarred,                                  // 25
		o.RegistrationInterzonalCFBarred,                                  // 26
		o.RegistrationInterzonalCFNotToHPLMNBarred,                        // 27
		o.RegistrationInternationalCFBarred,                               // 28
	}
	return packBits(bits, 15) // spec min
}

func convertBitStringToODBGeneralData(bs runtime.BitString) *ODBGeneralData {
	o := &ODBGeneralData{}
	o.AllOGCallsBarred = bs.Has(0)
	o.InternationalOGCallsBarred = bs.Has(1)
	o.InternationalOGCallsNotToHPLMNCountryBarred = bs.Has(2)
	o.PremiumRateInformationOGCallsBarred = bs.Has(3)
	o.PremiumRateEntertainmentOGCallsBarred = bs.Has(4)
	o.SSAccessBarred = bs.Has(5)
	o.InterzonalOGCallsBarred = bs.Has(6)
	o.InterzonalOGCallsNotToHPLMNCountryBarred = bs.Has(7)
	o.InterzonalOGCallsAndInternationalOGCallsNotToHPLMNCountryBarred = bs.Has(8)
	o.AllECTBarred = bs.Has(9)
	o.ChargeableECTBarred = bs.Has(10)
	o.InternationalECTBarred = bs.Has(11)
	o.InterzonalECTBarred = bs.Has(12)
	o.DoublyChargeableECTBarred = bs.Has(13)
	o.MultipleECTBarred = bs.Has(14)
	o.AllPacketOrientedServicesBarred = bs.Has(15)
	o.RoamerAccessToHPLMNAPBarred = bs.Has(16)
	o.RoamerAccessToVPLMNAPBarred = bs.Has(17)
	o.RoamingOutsidePLMNOGCallsBarred = bs.Has(18)
	o.AllICCallsBarred = bs.Has(19)
	o.RoamingOutsidePLMNICCallsBarred = bs.Has(20)
	o.RoamingOutsidePLMNICountryICCallsBarred = bs.Has(21)
	o.RoamingOutsidePLMNBarred = bs.Has(22)
	o.RoamingOutsidePLMNCountryBarred = bs.Has(23)
	o.RegistrationAllCFBarred = bs.Has(24)
	o.RegistrationCFNotToHPLMNBarred = bs.Has(25)
	o.RegistrationInterzonalCFBarred = bs.Has(26)
	o.RegistrationInterzonalCFNotToHPLMNBarred = bs.Has(27)
	o.RegistrationInternationalCFBarred = bs.Has(28)
	return o
}

// ODBHPLMNData: 4 named bits (SIZE 4..32) per MAP-MS-DataTypes.asn:1812.
func convertODBHPLMNDataToBitString(o *ODBHPLMNData) runtime.BitString {
	bits := []bool{o.PlmnSpecificBarringType1, o.PlmnSpecificBarringType2, o.PlmnSpecificBarringType3, o.PlmnSpecificBarringType4}
	return packBits(bits, 4)
}

func convertBitStringToODBHPLMNData(bs runtime.BitString) *ODBHPLMNData {
	return &ODBHPLMNData{
		PlmnSpecificBarringType1: bs.Has(0),
		PlmnSpecificBarringType2: bs.Has(1),
		PlmnSpecificBarringType3: bs.Has(2),
		PlmnSpecificBarringType4: bs.Has(3),
	}
}

// AccessRestrictionData: 8 named bits (SIZE 2..8) per MAP-MS-DataTypes.asn:1454.
func convertAccessRestrictionDataToBitString(a *AccessRestrictionData) runtime.BitString {
	bits := []bool{
		a.UtranNotAllowed, a.GeranNotAllowed, a.GanNotAllowed, a.IHSPAEvolutionNotAllowed,
		a.WBEUtranNotAllowed, a.HoToNon3GPPAccessNotAllowed, a.NBIoTNotAllowed, a.EnhancedCoverageNotAllowed,
	}
	return packBits(bits, 2)
}

func convertBitStringToAccessRestrictionData(bs runtime.BitString) *AccessRestrictionData {
	return &AccessRestrictionData{
		UtranNotAllowed:             bs.Has(0),
		GeranNotAllowed:             bs.Has(1),
		GanNotAllowed:               bs.Has(2),
		IHSPAEvolutionNotAllowed:    bs.Has(3),
		WBEUtranNotAllowed:          bs.Has(4),
		HoToNon3GPPAccessNotAllowed: bs.Has(5),
		NBIoTNotAllowed:             bs.Has(6),
		EnhancedCoverageNotAllowed:  bs.Has(7),
	}
}

// ExtAccessRestrictionData: 2 named bits (SIZE 1..32) per MAP-MS-DataTypes.asn:1471.
func convertExtAccessRestrictionDataToBitString(e *ExtAccessRestrictionData) runtime.BitString {
	bits := []bool{e.NrAsSecondaryRATNotAllowed, e.UnlicensedSpectrumAsSecondaryRATNotAllowed}
	return packBits(bits, 1)
}

func convertBitStringToExtAccessRestrictionData(bs runtime.BitString) *ExtAccessRestrictionData {
	return &ExtAccessRestrictionData{
		NrAsSecondaryRATNotAllowed:                 bs.Has(0),
		UnlicensedSpectrumAsSecondaryRATNotAllowed: bs.Has(1),
	}
}

// SupportedFeatures: 40 named bits (SIZE 26..40) per MAP-MS-DataTypes.asn:642.
func convertSupportedFeaturesToBitString(s *SupportedFeatures) runtime.BitString {
	bits := []bool{
		s.OdbAllApn, s.OdbHPLMNApn, s.OdbVPLMNApn, s.OdbAllOg, s.OdbAllInternationalOg,
		s.OdbAllIntOgNotToHPLMNCountry, s.OdbAllInterzonalOg, s.OdbAllInterzonalOgNotToHPLMNCountry,
		s.OdbAllInterzonalOgAndInternatOgNotToHPLMNCountry, s.RegSub, s.Trace, s.LcsAllPrivExcep,
		s.LcsUniversal, s.LcsCallSessionRelated, s.LcsCallSessionUnrelated, s.LcsPLMNOperator,
		s.LcsServiceType, s.LcsAllMOLRSS, s.LcsBasicSelfLocation, s.LcsAutonomousSelfLocation,
		s.LcsTransferToThirdParty, s.SmMoPp, s.BarringOutgoingCalls, s.Baoc, s.Boic, s.BoicExHC,
		s.LocalTimeZoneRetrieval, s.AdditionalMsisdn, s.SmsInMME, s.SmsInSGSN,
		s.UeReachabilityNotification, s.StateLocationInformationRetrieval, s.PartialPurge,
		s.GddInSGSN, s.SgsnCAMELCapability, s.PcscfRestoration, s.DedicatedCoreNetworks,
		s.NonIPPDNTypeAPNs, s.NonIPPDPTypeAPNs, s.NrAsSecondaryRAT,
	}
	return packBits(bits, 26)
}

func convertBitStringToSupportedFeatures(bs runtime.BitString) *SupportedFeatures {
	return &SupportedFeatures{
		OdbAllApn:                                        bs.Has(0),
		OdbHPLMNApn:                                      bs.Has(1),
		OdbVPLMNApn:                                      bs.Has(2),
		OdbAllOg:                                         bs.Has(3),
		OdbAllInternationalOg:                            bs.Has(4),
		OdbAllIntOgNotToHPLMNCountry:                     bs.Has(5),
		OdbAllInterzonalOg:                               bs.Has(6),
		OdbAllInterzonalOgNotToHPLMNCountry:              bs.Has(7),
		OdbAllInterzonalOgAndInternatOgNotToHPLMNCountry: bs.Has(8),
		RegSub:                            bs.Has(9),
		Trace:                             bs.Has(10),
		LcsAllPrivExcep:                   bs.Has(11),
		LcsUniversal:                      bs.Has(12),
		LcsCallSessionRelated:             bs.Has(13),
		LcsCallSessionUnrelated:           bs.Has(14),
		LcsPLMNOperator:                   bs.Has(15),
		LcsServiceType:                    bs.Has(16),
		LcsAllMOLRSS:                      bs.Has(17),
		LcsBasicSelfLocation:              bs.Has(18),
		LcsAutonomousSelfLocation:         bs.Has(19),
		LcsTransferToThirdParty:           bs.Has(20),
		SmMoPp:                            bs.Has(21),
		BarringOutgoingCalls:              bs.Has(22),
		Baoc:                              bs.Has(23),
		Boic:                              bs.Has(24),
		BoicExHC:                          bs.Has(25),
		LocalTimeZoneRetrieval:            bs.Has(26),
		AdditionalMsisdn:                  bs.Has(27),
		SmsInMME:                          bs.Has(28),
		SmsInSGSN:                         bs.Has(29),
		UeReachabilityNotification:        bs.Has(30),
		StateLocationInformationRetrieval: bs.Has(31),
		PartialPurge:                      bs.Has(32),
		GddInSGSN:                         bs.Has(33),
		SgsnCAMELCapability:               bs.Has(34),
		PcscfRestoration:                  bs.Has(35),
		DedicatedCoreNetworks:             bs.Has(36),
		NonIPPDNTypeAPNs:                  bs.Has(37),
		NonIPPDPTypeAPNs:                  bs.Has(38),
		NrAsSecondaryRAT:                  bs.Has(39),
	}
}

// ExtSupportedFeatures: 1 named bit (SIZE 1..40) per MAP-MS-DataTypes.asn:687.
func convertExtSupportedFeaturesToBitString(e *ExtSupportedFeatures) runtime.BitString {
	bits := []bool{e.UnlicensedSpectrumAsSecondaryRAT}
	return packBits(bits, 1)
}

func convertBitStringToExtSupportedFeatures(bs runtime.BitString) *ExtSupportedFeatures {
	return &ExtSupportedFeatures{
		UnlicensedSpectrumAsSecondaryRAT: bs.Has(0),
	}
}
