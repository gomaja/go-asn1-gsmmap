// BIT STRING <-> struct-of-bools helpers shared across operations.

package gsmmap

import (
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime"
)

// bitStringToWire builds the BIT STRING of a field the public type carries
// as raw octets and a bit length. The codec checks their consistency per
// ITU-T X.690 (02/2021) §8.6.2.2.
func bitStringToWire(octets []byte, bits int) runtime.BitString {
	return runtime.BitString{Bytes: clearPaddingBits(octets, bits), BitLength: bits}
}

// bitStringFromWire copies decoded octets; the codec clears unused bits.
func bitStringFromWire(bs runtime.BitString) HexBytes {
	return append(HexBytes(nil), bs.Bytes...)
}

// clearPaddingBits copies the octets of a bits-long BIT STRING and clears
// the unused bits of the last octet. BER leaves them unconstrained on
// receipt (ITU-T X.690 (02/2021) §11.2.1 requires zero in DER and CER).
// The sender uses zero padding.
func clearPaddingBits(octets []byte, bits int) HexBytes {
	out := append(HexBytes(nil), octets...)
	if r := bits % 8; r != 0 && len(out) > 0 {
		out[len(out)-1] &= 0xFF << uint(8-r)
	}
	return out
}

func convertCamelPhasesToBitString(cp *SupportedCamelPhases) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.SupportedCamelPhasesPhase1: cp.Phase1,
		gsm_map.SupportedCamelPhasesPhase2: cp.Phase2,
		gsm_map.SupportedCamelPhasesPhase3: cp.Phase3,
		gsm_map.SupportedCamelPhasesPhase4: cp.Phase4,
	}
	return gsm_map.NewSupportedCamelPhases(namedBitPositions(bits)...)
}

func convertBitStringToCamelPhases(bs runtime.BitString) *SupportedCamelPhases {
	cp := &SupportedCamelPhases{}
	cp.Phase1 = bs.Has(gsm_map.SupportedCamelPhasesPhase1)
	cp.Phase2 = bs.Has(gsm_map.SupportedCamelPhasesPhase2)
	cp.Phase3 = bs.Has(gsm_map.SupportedCamelPhasesPhase3)
	cp.Phase4 = bs.Has(gsm_map.SupportedCamelPhasesPhase4)
	return cp
}

func convertLCSCapsToBitString(lcs *SupportedLCSCapabilitySets) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.SupportedLCSCapabilitySetsLcsCapabilitySet1: lcs.LcsCapabilitySet1,
		gsm_map.SupportedLCSCapabilitySetsLcsCapabilitySet2: lcs.LcsCapabilitySet2,
		gsm_map.SupportedLCSCapabilitySetsLcsCapabilitySet3: lcs.LcsCapabilitySet3,
		gsm_map.SupportedLCSCapabilitySetsLcsCapabilitySet4: lcs.LcsCapabilitySet4,
		gsm_map.SupportedLCSCapabilitySetsLcsCapabilitySet5: lcs.LcsCapabilitySet5,
	}
	return gsm_map.NewSupportedLCSCapabilitySets(namedBitPositions(bits)...)
}

func convertBitStringToLCSCaps(bs runtime.BitString) *SupportedLCSCapabilitySets {
	lcs := &SupportedLCSCapabilitySets{}
	lcs.LcsCapabilitySet1 = bs.Has(gsm_map.SupportedLCSCapabilitySetsLcsCapabilitySet1)
	lcs.LcsCapabilitySet2 = bs.Has(gsm_map.SupportedLCSCapabilitySetsLcsCapabilitySet2)
	lcs.LcsCapabilitySet3 = bs.Has(gsm_map.SupportedLCSCapabilitySetsLcsCapabilitySet3)
	lcs.LcsCapabilitySet4 = bs.Has(gsm_map.SupportedLCSCapabilitySetsLcsCapabilitySet4)
	lcs.LcsCapabilitySet5 = bs.Has(gsm_map.SupportedLCSCapabilitySetsLcsCapabilitySet5)
	return lcs
}

func convertRequestedNodesToBitString(rn *RequestedNodes) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.RequestedNodesMme:  rn.MME,
		gsm_map.RequestedNodesSgsn: rn.SGSN,
	}
	return gsm_map.NewRequestedNodes(namedBitPositions(bits)...)
}

func convertBitStringToRequestedNodes(bs runtime.BitString) *RequestedNodes {
	rn := &RequestedNodes{}
	rn.MME = bs.Has(gsm_map.RequestedNodesMme)
	rn.SGSN = bs.Has(gsm_map.RequestedNodesSgsn)
	return rn
}

// AllowedServices: 2 bits (bit 0 = first, bit 1 = second).
func convertAllowedServicesToBitString(a *AllowedServicesFlags) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.AllowedServicesFirstServiceAllowed:  a.FirstServiceAllowed,
		gsm_map.AllowedServicesSecondServiceAllowed: a.SecondServiceAllowed,
	}
	return gsm_map.NewAllowedServices(namedBitPositions(bits)...)
}

func convertBitStringToAllowedServices(bs runtime.BitString) *AllowedServicesFlags {
	a := &AllowedServicesFlags{}
	a.FirstServiceAllowed = bs.Has(gsm_map.AllowedServicesFirstServiceAllowed)
	a.SecondServiceAllowed = bs.Has(gsm_map.AllowedServicesSecondServiceAllowed)
	return a
}

// SuppressMTSS: 2 bits (bit 0 = suppressCUG, bit 1 = suppressCCBS), min size 2.
func convertSuppressMTSSToBitString(s *SuppressMTSSFlags) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.SuppressMTSSSuppressCUG:  s.SuppressCUG,
		gsm_map.SuppressMTSSSuppressCCBS: s.SuppressCCBS,
	}
	return gsm_map.NewSuppressMTSS(namedBitPositions(bits)...)
}

func convertBitStringToSuppressMTSS(bs runtime.BitString) *SuppressMTSSFlags {
	s := &SuppressMTSSFlags{}
	s.SuppressCUG = bs.Has(gsm_map.SuppressMTSSSuppressCUG)
	s.SuppressCCBS = bs.Has(gsm_map.SuppressMTSSSuppressCCBS)
	return s
}

// OfferedCamel4CSIs: 7 bits per 3GPP TS 29.002.
// Bit order: 0=o-CSI, 1=d-CSI, 2=vt-CSI, 3=t-CSI, 4=mt-sms-CSI, 5=mg-CSI, 6=psi-enhancements.
func convertOfferedCamel4CSIsToBitString(o *OfferedCamel4CSIs) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.OfferedCamel4CSIsOCsi:            o.OCSI,
		gsm_map.OfferedCamel4CSIsDCsi:            o.DCSI,
		gsm_map.OfferedCamel4CSIsVtCsi:           o.VTCSI,
		gsm_map.OfferedCamel4CSIsTCsi:            o.TCSI,
		gsm_map.OfferedCamel4CSIsMtSmsCsi:        o.MTSMSCSI,
		gsm_map.OfferedCamel4CSIsMgCsi:           o.MGCSI,
		gsm_map.OfferedCamel4CSIsPsiEnhancements: o.PsiEnhancements,
	}
	return gsm_map.NewOfferedCamel4CSIs(namedBitPositions(bits)...)
}

func convertBitStringToOfferedCamel4CSIs(bs runtime.BitString) *OfferedCamel4CSIs {
	o := &OfferedCamel4CSIs{}
	o.OCSI = bs.Has(gsm_map.OfferedCamel4CSIsOCsi)
	o.DCSI = bs.Has(gsm_map.OfferedCamel4CSIsDCsi)
	o.VTCSI = bs.Has(gsm_map.OfferedCamel4CSIsVtCsi)
	o.TCSI = bs.Has(gsm_map.OfferedCamel4CSIsTCsi)
	o.MTSMSCSI = bs.Has(gsm_map.OfferedCamel4CSIsMtSmsCsi)
	o.MGCSI = bs.Has(gsm_map.OfferedCamel4CSIsMgCsi)
	o.PsiEnhancements = bs.Has(gsm_map.OfferedCamel4CSIsPsiEnhancements)
	return o
}

// SupportedRATTypes: bit 0=utran, 1=geran, 2=gan, 3=i-hspa-evolution,
// 4=e-utran, 5=nb-iot.
func convertSupportedRATTypesToBitString(r *SupportedRATTypes) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.SupportedRATTypesUtran:          r.UTRAN,
		gsm_map.SupportedRATTypesGeran:          r.GERAN,
		gsm_map.SupportedRATTypesGan:            r.GAN,
		gsm_map.SupportedRATTypesIHspaEvolution: r.IHSPAEvolution,
		gsm_map.SupportedRATTypesEUtran:         r.EUTRAN,
		gsm_map.SupportedRATTypesNbIot:          r.NBIoT,
	}
	return gsm_map.NewSupportedRATTypes(namedBitPositions(bits)...)
}

func convertBitStringToSupportedRATTypes(bs runtime.BitString) *SupportedRATTypes {
	r := &SupportedRATTypes{}
	r.UTRAN = bs.Has(gsm_map.SupportedRATTypesUtran)
	r.GERAN = bs.Has(gsm_map.SupportedRATTypesGeran)
	r.GAN = bs.Has(gsm_map.SupportedRATTypesGan)
	r.IHSPAEvolution = bs.Has(gsm_map.SupportedRATTypesIHspaEvolution)
	r.EUTRAN = bs.Has(gsm_map.SupportedRATTypesEUtran)
	r.NBIoT = bs.Has(gsm_map.SupportedRATTypesNbIot)
	return r
}

// namedBitPositions collects the set positions of a table keyed by named-bit constants.
func namedBitPositions(bits []bool) []int {
	var positions []int
	for bit, set := range bits {
		if set {
			positions = append(positions, bit)
		}
	}
	return positions
}

// ODBGeneralData: 29 named bits (SIZE 15..32) per 3GPP TS 29.002 V19.1.0 §17.7.1.
func convertODBGeneralDataToBitString(o *ODBGeneralData) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.ODBGeneralDataAllOGCallsBarred:                                                o.AllOGCallsBarred,
		gsm_map.ODBGeneralDataInternationalOGCallsBarred:                                      o.InternationalOGCallsBarred,
		gsm_map.ODBGeneralDataInternationalOGCallsNotToHPLMNCountryBarred:                     o.InternationalOGCallsNotToHPLMNCountryBarred,
		gsm_map.ODBGeneralDataPremiumRateInformationOGCallsBarred:                             o.PremiumRateInformationOGCallsBarred,
		gsm_map.ODBGeneralDataPremiumRateEntertainementOGCallsBarred:                          o.PremiumRateEntertainmentOGCallsBarred,
		gsm_map.ODBGeneralDataSsAccessBarred:                                                  o.SSAccessBarred,
		gsm_map.ODBGeneralDataInterzonalOGCallsBarred:                                         o.InterzonalOGCallsBarred,
		gsm_map.ODBGeneralDataInterzonalOGCallsNotToHPLMNCountryBarred:                        o.InterzonalOGCallsNotToHPLMNCountryBarred,
		gsm_map.ODBGeneralDataInterzonalOGCallsAndInternationalOGCallsNotToHPLMNCountryBarred: o.InterzonalOGCallsAndInternationalOGCallsNotToHPLMNCountryBarred,
		gsm_map.ODBGeneralDataAllECTBarred:                                                    o.AllECTBarred,
		gsm_map.ODBGeneralDataChargeableECTBarred:                                             o.ChargeableECTBarred,
		gsm_map.ODBGeneralDataInternationalECTBarred:                                          o.InternationalECTBarred,
		gsm_map.ODBGeneralDataInterzonalECTBarred:                                             o.InterzonalECTBarred,
		gsm_map.ODBGeneralDataDoublyChargeableECTBarred:                                       o.DoublyChargeableECTBarred,
		gsm_map.ODBGeneralDataMultipleECTBarred:                                               o.MultipleECTBarred,
		gsm_map.ODBGeneralDataAllPacketOrientedServicesBarred:                                 o.AllPacketOrientedServicesBarred,
		gsm_map.ODBGeneralDataRoamerAccessToHPLMNAPBarred:                                     o.RoamerAccessToHPLMNAPBarred,
		gsm_map.ODBGeneralDataRoamerAccessToVPLMNAPBarred:                                     o.RoamerAccessToVPLMNAPBarred,
		gsm_map.ODBGeneralDataRoamingOutsidePLMNOGCallsBarred:                                 o.RoamingOutsidePLMNOGCallsBarred,
		gsm_map.ODBGeneralDataAllICCallsBarred:                                                o.AllICCallsBarred,
		gsm_map.ODBGeneralDataRoamingOutsidePLMNICCallsBarred:                                 o.RoamingOutsidePLMNICCallsBarred,
		gsm_map.ODBGeneralDataRoamingOutsidePLMNICountryICCallsBarred:                         o.RoamingOutsidePLMNICountryICCallsBarred,
		gsm_map.ODBGeneralDataRoamingOutsidePLMNBarred:                                        o.RoamingOutsidePLMNBarred,
		gsm_map.ODBGeneralDataRoamingOutsidePLMNCountryBarred:                                 o.RoamingOutsidePLMNCountryBarred,
		gsm_map.ODBGeneralDataRegistrationAllCFBarred:                                         o.RegistrationAllCFBarred,
		gsm_map.ODBGeneralDataRegistrationCFNotToHPLMNBarred:                                  o.RegistrationCFNotToHPLMNBarred,
		gsm_map.ODBGeneralDataRegistrationInterzonalCFBarred:                                  o.RegistrationInterzonalCFBarred,
		gsm_map.ODBGeneralDataRegistrationInterzonalCFNotToHPLMNBarred:                        o.RegistrationInterzonalCFNotToHPLMNBarred,
		gsm_map.ODBGeneralDataRegistrationInternationalCFBarred:                               o.RegistrationInternationalCFBarred,
	}
	return gsm_map.NewODBGeneralData(namedBitPositions(bits)...)
}

func convertBitStringToODBGeneralData(bs runtime.BitString) *ODBGeneralData {
	o := &ODBGeneralData{}
	o.AllOGCallsBarred = bs.Has(gsm_map.ODBGeneralDataAllOGCallsBarred)
	o.InternationalOGCallsBarred = bs.Has(gsm_map.ODBGeneralDataInternationalOGCallsBarred)
	o.InternationalOGCallsNotToHPLMNCountryBarred = bs.Has(gsm_map.ODBGeneralDataInternationalOGCallsNotToHPLMNCountryBarred)
	o.PremiumRateInformationOGCallsBarred = bs.Has(gsm_map.ODBGeneralDataPremiumRateInformationOGCallsBarred)
	o.PremiumRateEntertainmentOGCallsBarred = bs.Has(gsm_map.ODBGeneralDataPremiumRateEntertainementOGCallsBarred)
	o.SSAccessBarred = bs.Has(gsm_map.ODBGeneralDataSsAccessBarred)
	o.InterzonalOGCallsBarred = bs.Has(gsm_map.ODBGeneralDataInterzonalOGCallsBarred)
	o.InterzonalOGCallsNotToHPLMNCountryBarred = bs.Has(gsm_map.ODBGeneralDataInterzonalOGCallsNotToHPLMNCountryBarred)
	o.InterzonalOGCallsAndInternationalOGCallsNotToHPLMNCountryBarred = bs.Has(gsm_map.ODBGeneralDataInterzonalOGCallsAndInternationalOGCallsNotToHPLMNCountryBarred)
	o.AllECTBarred = bs.Has(gsm_map.ODBGeneralDataAllECTBarred)
	o.ChargeableECTBarred = bs.Has(gsm_map.ODBGeneralDataChargeableECTBarred)
	o.InternationalECTBarred = bs.Has(gsm_map.ODBGeneralDataInternationalECTBarred)
	o.InterzonalECTBarred = bs.Has(gsm_map.ODBGeneralDataInterzonalECTBarred)
	o.DoublyChargeableECTBarred = bs.Has(gsm_map.ODBGeneralDataDoublyChargeableECTBarred)
	o.MultipleECTBarred = bs.Has(gsm_map.ODBGeneralDataMultipleECTBarred)
	o.AllPacketOrientedServicesBarred = bs.Has(gsm_map.ODBGeneralDataAllPacketOrientedServicesBarred)
	o.RoamerAccessToHPLMNAPBarred = bs.Has(gsm_map.ODBGeneralDataRoamerAccessToHPLMNAPBarred)
	o.RoamerAccessToVPLMNAPBarred = bs.Has(gsm_map.ODBGeneralDataRoamerAccessToVPLMNAPBarred)
	o.RoamingOutsidePLMNOGCallsBarred = bs.Has(gsm_map.ODBGeneralDataRoamingOutsidePLMNOGCallsBarred)
	o.AllICCallsBarred = bs.Has(gsm_map.ODBGeneralDataAllICCallsBarred)
	o.RoamingOutsidePLMNICCallsBarred = bs.Has(gsm_map.ODBGeneralDataRoamingOutsidePLMNICCallsBarred)
	o.RoamingOutsidePLMNICountryICCallsBarred = bs.Has(gsm_map.ODBGeneralDataRoamingOutsidePLMNICountryICCallsBarred)
	o.RoamingOutsidePLMNBarred = bs.Has(gsm_map.ODBGeneralDataRoamingOutsidePLMNBarred)
	o.RoamingOutsidePLMNCountryBarred = bs.Has(gsm_map.ODBGeneralDataRoamingOutsidePLMNCountryBarred)
	o.RegistrationAllCFBarred = bs.Has(gsm_map.ODBGeneralDataRegistrationAllCFBarred)
	o.RegistrationCFNotToHPLMNBarred = bs.Has(gsm_map.ODBGeneralDataRegistrationCFNotToHPLMNBarred)
	o.RegistrationInterzonalCFBarred = bs.Has(gsm_map.ODBGeneralDataRegistrationInterzonalCFBarred)
	o.RegistrationInterzonalCFNotToHPLMNBarred = bs.Has(gsm_map.ODBGeneralDataRegistrationInterzonalCFNotToHPLMNBarred)
	o.RegistrationInternationalCFBarred = bs.Has(gsm_map.ODBGeneralDataRegistrationInternationalCFBarred)
	return o
}

// ODBHPLMNData: 4 named bits (SIZE 4..32) per 3GPP TS 29.002 V19.1.0 §17.7.1.
func convertODBHPLMNDataToBitString(o *ODBHPLMNData) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.ODBHPLMNDataPlmnSpecificBarringType1: o.PLMNSpecificBarringType1,
		gsm_map.ODBHPLMNDataPlmnSpecificBarringType2: o.PLMNSpecificBarringType2,
		gsm_map.ODBHPLMNDataPlmnSpecificBarringType3: o.PLMNSpecificBarringType3,
		gsm_map.ODBHPLMNDataPlmnSpecificBarringType4: o.PLMNSpecificBarringType4,
	}
	return gsm_map.NewODBHPLMNData(namedBitPositions(bits)...)
}

func convertBitStringToODBHPLMNData(bs runtime.BitString) *ODBHPLMNData {
	return &ODBHPLMNData{
		PLMNSpecificBarringType1: bs.Has(gsm_map.ODBHPLMNDataPlmnSpecificBarringType1),
		PLMNSpecificBarringType2: bs.Has(gsm_map.ODBHPLMNDataPlmnSpecificBarringType2),
		PLMNSpecificBarringType3: bs.Has(gsm_map.ODBHPLMNDataPlmnSpecificBarringType3),
		PLMNSpecificBarringType4: bs.Has(gsm_map.ODBHPLMNDataPlmnSpecificBarringType4),
	}
}

// AccessRestrictionData: 8 named bits (SIZE 2..8) per 3GPP TS 29.002 V19.1.0 §17.7.1.
func convertAccessRestrictionDataToBitString(a *AccessRestrictionData) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.AccessRestrictionDataUtranNotAllowed:             a.UtranNotAllowed,
		gsm_map.AccessRestrictionDataGeranNotAllowed:             a.GeranNotAllowed,
		gsm_map.AccessRestrictionDataGanNotAllowed:               a.GanNotAllowed,
		gsm_map.AccessRestrictionDataIHspaEvolutionNotAllowed:    a.IHSPAEvolutionNotAllowed,
		gsm_map.AccessRestrictionDataWbEUtranNotAllowed:          a.WBEUtranNotAllowed,
		gsm_map.AccessRestrictionDataHoToNon3GPPAccessNotAllowed: a.HoToNon3GPPAccessNotAllowed,
		gsm_map.AccessRestrictionDataNbIotNotAllowed:             a.NBIoTNotAllowed,
		gsm_map.AccessRestrictionDataEnhancedCoverageNotAllowed:  a.EnhancedCoverageNotAllowed,
	}
	return gsm_map.NewAccessRestrictionData(namedBitPositions(bits)...)
}

func convertBitStringToAccessRestrictionData(bs runtime.BitString) *AccessRestrictionData {
	return &AccessRestrictionData{
		UtranNotAllowed:             bs.Has(gsm_map.AccessRestrictionDataUtranNotAllowed),
		GeranNotAllowed:             bs.Has(gsm_map.AccessRestrictionDataGeranNotAllowed),
		GanNotAllowed:               bs.Has(gsm_map.AccessRestrictionDataGanNotAllowed),
		IHSPAEvolutionNotAllowed:    bs.Has(gsm_map.AccessRestrictionDataIHspaEvolutionNotAllowed),
		WBEUtranNotAllowed:          bs.Has(gsm_map.AccessRestrictionDataWbEUtranNotAllowed),
		HoToNon3GPPAccessNotAllowed: bs.Has(gsm_map.AccessRestrictionDataHoToNon3GPPAccessNotAllowed),
		NBIoTNotAllowed:             bs.Has(gsm_map.AccessRestrictionDataNbIotNotAllowed),
		EnhancedCoverageNotAllowed:  bs.Has(gsm_map.AccessRestrictionDataEnhancedCoverageNotAllowed),
	}
}

// ExtAccessRestrictionData: 2 named bits (SIZE 1..32) per 3GPP TS 29.002 V19.1.0 §17.7.1.
func convertExtAccessRestrictionDataToBitString(e *ExtAccessRestrictionData) (runtime.BitString, error) {
	bits := make([]bool, max(2, len(e.UnknownBits)*8))
	bits[gsm_map.ExtAccessRestrictionDataNrAsSecondaryRATNotAllowed], bits[gsm_map.ExtAccessRestrictionDataUnlicensedSpectrumAsSecondaryRATNotAllowed] = e.NrAsSecondaryRATNotAllowed, e.UnlicensedSpectrumAsSecondaryRATNotAllowed
	for i := gsm_map.ExtAccessRestrictionDataUnlicensedSpectrumAsSecondaryRATNotAllowed + 1; i < len(bits); i++ {
		bits[i] = e.UnknownBits[i/8]&(0x80>>uint(i%8)) != 0
	}
	return gsm_map.NewExtAccessRestrictionData(namedBitPositions(bits)...)
}

func convertBitStringToExtAccessRestrictionData(bs runtime.BitString) *ExtAccessRestrictionData {
	out := &ExtAccessRestrictionData{
		NrAsSecondaryRATNotAllowed:                 bs.Has(gsm_map.ExtAccessRestrictionDataNrAsSecondaryRATNotAllowed),
		UnlicensedSpectrumAsSecondaryRATNotAllowed: bs.Has(gsm_map.ExtAccessRestrictionDataUnlicensedSpectrumAsSecondaryRATNotAllowed),
	}
	// TS 29.002 V19.1.0 §17.7.1 Ext-AccessRestrictionData allows
	// unnamed bits within SIZE (1..32), with no receiver discard rule.
	if len(bs.Bytes) > 0 {
		out.UnknownBits = bitStringFromWire(bs)
		out.UnknownBits[0] &^= 1<<(7-gsm_map.ExtAccessRestrictionDataNrAsSecondaryRATNotAllowed) | 1<<(7-gsm_map.ExtAccessRestrictionDataUnlicensedSpectrumAsSecondaryRATNotAllowed)
		for len(out.UnknownBits) > 0 && out.UnknownBits[len(out.UnknownBits)-1] == 0 {
			out.UnknownBits = out.UnknownBits[:len(out.UnknownBits)-1]
		}
		if len(out.UnknownBits) == 0 {
			out.UnknownBits = nil
		}
	}
	return out
}

// SupportedFeatures: 40 named bits (SIZE 26..40) per 3GPP TS 29.002 V19.1.0 §17.7.1.
func convertSupportedFeaturesToBitString(s *SupportedFeatures) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.SupportedFeaturesOdbAllApn:                                        s.OdbAllApn,
		gsm_map.SupportedFeaturesOdbHPLMNAPN:                                      s.OdbHPLMNApn,
		gsm_map.SupportedFeaturesOdbVPLMNAPN:                                      s.OdbVPLMNApn,
		gsm_map.SupportedFeaturesOdbAllOg:                                         s.OdbAllOg,
		gsm_map.SupportedFeaturesOdbAllInternationalOg:                            s.OdbAllInternationalOg,
		gsm_map.SupportedFeaturesOdbAllIntOgNotToHPLMNCountry:                     s.OdbAllIntOgNotToHPLMNCountry,
		gsm_map.SupportedFeaturesOdbAllInterzonalOg:                               s.OdbAllInterzonalOg,
		gsm_map.SupportedFeaturesOdbAllInterzonalOgNotToHPLMNCountry:              s.OdbAllInterzonalOgNotToHPLMNCountry,
		gsm_map.SupportedFeaturesOdbAllInterzonalOgAndInternatOgNotToHPLMNCountry: s.OdbAllInterzonalOgAndInternatOgNotToHPLMNCountry,
		gsm_map.SupportedFeaturesRegSub:                                           s.RegSub,
		gsm_map.SupportedFeaturesTrace:                                            s.Trace,
		gsm_map.SupportedFeaturesLcsAllPrivExcep:                                  s.LcsAllPrivExcep,
		gsm_map.SupportedFeaturesLcsUniversal:                                     s.LcsUniversal,
		gsm_map.SupportedFeaturesLcsCallSessionRelated:                            s.LcsCallSessionRelated,
		gsm_map.SupportedFeaturesLcsCallSessionUnrelated:                          s.LcsCallSessionUnrelated,
		gsm_map.SupportedFeaturesLcsPLMNOperator:                                  s.LcsPLMNOperator,
		gsm_map.SupportedFeaturesLcsServiceType:                                   s.LcsServiceType,
		gsm_map.SupportedFeaturesLcsAllMOLRSS:                                     s.LcsAllMOLRSS,
		gsm_map.SupportedFeaturesLcsBasicSelfLocation:                             s.LcsBasicSelfLocation,
		gsm_map.SupportedFeaturesLcsAutonomousSelfLocation:                        s.LcsAutonomousSelfLocation,
		gsm_map.SupportedFeaturesLcsTransferToThirdParty:                          s.LcsTransferToThirdParty,
		gsm_map.SupportedFeaturesSmMoPp:                                           s.SmMoPp,
		gsm_map.SupportedFeaturesBarringOutgoingCalls:                             s.BarringOutgoingCalls,
		gsm_map.SupportedFeaturesBaoc:                                             s.Baoc,
		gsm_map.SupportedFeaturesBoic:                                             s.Boic,
		gsm_map.SupportedFeaturesBoicExHC:                                         s.BoicExHC,
		gsm_map.SupportedFeaturesLocalTimeZoneRetrieval:                           s.LocalTimeZoneRetrieval,
		gsm_map.SupportedFeaturesAdditionalMsisdn:                                 s.AdditionalMsisdn,
		gsm_map.SupportedFeaturesSmsInMME:                                         s.SmsInMME,
		gsm_map.SupportedFeaturesSmsInSGSN:                                        s.SmsInSGSN,
		gsm_map.SupportedFeaturesUeReachabilityNotification:                       s.UeReachabilityNotification,
		gsm_map.SupportedFeaturesStateLocationInformationRetrieval:                s.StateLocationInformationRetrieval,
		gsm_map.SupportedFeaturesPartialPurge:                                     s.PartialPurge,
		gsm_map.SupportedFeaturesGddInSGSN:                                        s.GddInSGSN,
		gsm_map.SupportedFeaturesSgsnCAMELCapability:                              s.SgsnCAMELCapability,
		gsm_map.SupportedFeaturesPcscfRestoration:                                 s.PcscfRestoration,
		gsm_map.SupportedFeaturesDedicatedCoreNetworks:                            s.DedicatedCoreNetworks,
		gsm_map.SupportedFeaturesNonIPPDNTypeAPNs:                                 s.NonIPPDNTypeAPNs,
		gsm_map.SupportedFeaturesNonIPPDPTypeAPNs:                                 s.NonIPPDPTypeAPNs,
		gsm_map.SupportedFeaturesNrAsSecondaryRAT:                                 s.NrAsSecondaryRAT,
	}
	return gsm_map.NewSupportedFeatures(namedBitPositions(bits)...)
}

func convertBitStringToSupportedFeatures(bs runtime.BitString) *SupportedFeatures {
	out := &SupportedFeatures{
		OdbAllApn:                           bs.Has(gsm_map.SupportedFeaturesOdbAllApn),
		OdbHPLMNApn:                         bs.Has(gsm_map.SupportedFeaturesOdbHPLMNAPN),
		OdbVPLMNApn:                         bs.Has(gsm_map.SupportedFeaturesOdbVPLMNAPN),
		OdbAllOg:                            bs.Has(gsm_map.SupportedFeaturesOdbAllOg),
		OdbAllInternationalOg:               bs.Has(gsm_map.SupportedFeaturesOdbAllInternationalOg),
		OdbAllIntOgNotToHPLMNCountry:        bs.Has(gsm_map.SupportedFeaturesOdbAllIntOgNotToHPLMNCountry),
		OdbAllInterzonalOg:                  bs.Has(gsm_map.SupportedFeaturesOdbAllInterzonalOg),
		OdbAllInterzonalOgNotToHPLMNCountry: bs.Has(gsm_map.SupportedFeaturesOdbAllInterzonalOgNotToHPLMNCountry),
		OdbAllInterzonalOgAndInternatOgNotToHPLMNCountry: bs.Has(gsm_map.SupportedFeaturesOdbAllInterzonalOgAndInternatOgNotToHPLMNCountry),
		RegSub:                            bs.Has(gsm_map.SupportedFeaturesRegSub),
		Trace:                             bs.Has(gsm_map.SupportedFeaturesTrace),
		LcsAllPrivExcep:                   bs.Has(gsm_map.SupportedFeaturesLcsAllPrivExcep),
		LcsUniversal:                      bs.Has(gsm_map.SupportedFeaturesLcsUniversal),
		LcsCallSessionRelated:             bs.Has(gsm_map.SupportedFeaturesLcsCallSessionRelated),
		LcsCallSessionUnrelated:           bs.Has(gsm_map.SupportedFeaturesLcsCallSessionUnrelated),
		LcsPLMNOperator:                   bs.Has(gsm_map.SupportedFeaturesLcsPLMNOperator),
		LcsServiceType:                    bs.Has(gsm_map.SupportedFeaturesLcsServiceType),
		LcsAllMOLRSS:                      bs.Has(gsm_map.SupportedFeaturesLcsAllMOLRSS),
		LcsBasicSelfLocation:              bs.Has(gsm_map.SupportedFeaturesLcsBasicSelfLocation),
		LcsAutonomousSelfLocation:         bs.Has(gsm_map.SupportedFeaturesLcsAutonomousSelfLocation),
		LcsTransferToThirdParty:           bs.Has(gsm_map.SupportedFeaturesLcsTransferToThirdParty),
		SmMoPp:                            bs.Has(gsm_map.SupportedFeaturesSmMoPp),
		BarringOutgoingCalls:              bs.Has(gsm_map.SupportedFeaturesBarringOutgoingCalls),
		Baoc:                              bs.Has(gsm_map.SupportedFeaturesBaoc),
		Boic:                              bs.Has(gsm_map.SupportedFeaturesBoic),
		BoicExHC:                          bs.Has(gsm_map.SupportedFeaturesBoicExHC),
		LocalTimeZoneRetrieval:            bs.Has(gsm_map.SupportedFeaturesLocalTimeZoneRetrieval),
		AdditionalMsisdn:                  bs.Has(gsm_map.SupportedFeaturesAdditionalMsisdn),
		SmsInMME:                          bs.Has(gsm_map.SupportedFeaturesSmsInMME),
		SmsInSGSN:                         bs.Has(gsm_map.SupportedFeaturesSmsInSGSN),
		UeReachabilityNotification:        bs.Has(gsm_map.SupportedFeaturesUeReachabilityNotification),
		StateLocationInformationRetrieval: bs.Has(gsm_map.SupportedFeaturesStateLocationInformationRetrieval),
		PartialPurge:                      bs.Has(gsm_map.SupportedFeaturesPartialPurge),
		GddInSGSN:                         bs.Has(gsm_map.SupportedFeaturesGddInSGSN),
		SgsnCAMELCapability:               bs.Has(gsm_map.SupportedFeaturesSgsnCAMELCapability),
		PcscfRestoration:                  bs.Has(gsm_map.SupportedFeaturesPcscfRestoration),
		DedicatedCoreNetworks:             bs.Has(gsm_map.SupportedFeaturesDedicatedCoreNetworks),
		NonIPPDNTypeAPNs:                  bs.Has(gsm_map.SupportedFeaturesNonIPPDNTypeAPNs),
		NonIPPDPTypeAPNs:                  bs.Has(gsm_map.SupportedFeaturesNonIPPDPTypeAPNs),
		NrAsSecondaryRAT:                  bs.Has(gsm_map.SupportedFeaturesNrAsSecondaryRAT),
	}
	return out
}

// ExtSupportedFeatures: bit 0 is named; positions 1..39 are
// unnamed but valid (3GPP TS 29.002 V19.1.0 §17.7.1, SIZE (1..40)).
func convertExtSupportedFeaturesToBitString(e *ExtSupportedFeatures) (runtime.BitString, error) {
	bits := make([]bool, len(e.UnknownBits)*8)
	if len(bits) == 0 {
		bits = make([]bool, 1)
	}
	bits[gsm_map.ExtSupportedFeaturesUnlicensedSpectrumAsSecondaryRAT] = e.UnlicensedSpectrumAsSecondaryRAT
	for i := gsm_map.ExtSupportedFeaturesUnlicensedSpectrumAsSecondaryRAT + 1; i < len(bits); i++ {
		bits[i] = e.UnknownBits[i/8]&(0x80>>uint(i%8)) != 0
	}
	return gsm_map.NewExtSupportedFeatures(namedBitPositions(bits)...)
}

func convertBitStringToExtSupportedFeatures(bs runtime.BitString) *ExtSupportedFeatures {
	out := &ExtSupportedFeatures{
		UnlicensedSpectrumAsSecondaryRAT: bs.Has(gsm_map.ExtSupportedFeaturesUnlicensedSpectrumAsSecondaryRAT),
	}
	if len(bs.Bytes) > 0 {
		out.UnknownBits = bitStringFromWire(bs)
		out.UnknownBits[0] &^= 1 << (7 - gsm_map.ExtSupportedFeaturesUnlicensedSpectrumAsSecondaryRAT)
		for len(out.UnknownBits) > 0 && out.UnknownBits[len(out.UnknownBits)-1] == 0 {
			out.UnknownBits = out.UnknownBits[:len(out.UnknownBits)-1]
		}
		if len(out.UnknownBits) == 0 {
			out.UnknownBits = nil
		}
	}
	return out
}

// AdditionalSubscriptions: 3 named bits (SIZE 3..8) per 3GPP TS 29.002 V19.1.0 §17.7.1.
// "Other bits than listed above shall be discarded" per spec.
func convertAdditionalSubscriptionsToBitString(a *AdditionalSubscriptions) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.AdditionalSubscriptionsPrivilegedUplinkRequest: a.PrivilegedUplinkRequest,
		gsm_map.AdditionalSubscriptionsEmergencyUplinkRequest:  a.EmergencyUplinkRequest,
		gsm_map.AdditionalSubscriptionsEmergencyReset:          a.EmergencyReset,
	}
	return gsm_map.NewAdditionalSubscriptions(namedBitPositions(bits)...)
}

func convertBitStringToAdditionalSubscriptions(bs runtime.BitString) *AdditionalSubscriptions {
	return &AdditionalSubscriptions{
		PrivilegedUplinkRequest: bs.Has(gsm_map.AdditionalSubscriptionsPrivilegedUplinkRequest),
		EmergencyUplinkRequest:  bs.Has(gsm_map.AdditionalSubscriptionsEmergencyUplinkRequest),
		EmergencyReset:          bs.Has(gsm_map.AdditionalSubscriptionsEmergencyReset),
	}
}
