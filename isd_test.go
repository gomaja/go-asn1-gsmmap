// isd_test.go
//
// Tests for InsertSubscriberData (opCode 7) foundation types and BIT
// STRING helpers. PR A of the staged implementation — top-level ISD
// struct + nested SEQUENCE converters land in follow-up PRs.
package gsmmap

import (
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/runtime"
)

// Compile-smoke: every new public type must be referenceable.
func TestISDTypesCompile(t *testing.T) {
	var _ SubscriberStatus
	var _ NetworkAccessMode
	var _ RegionalSubscriptionResponse
	var _ ODBGeneralData
	var _ ODBHPLMNData
	var _ AccessRestrictionData
	var _ ExtAccessRestrictionData
	var _ SupportedFeatures
	var _ ExtSupportedFeatures

	// Enum constants exist.
	_ = SubscriberStatusServiceGranted
	_ = SubscriberStatusOperatorDeterminedBarring
	_ = NetworkAccessModePacketAndCircuit
	_ = NetworkAccessModeOnlyCircuit
	_ = NetworkAccessModeOnlyPacket
	_ = RegionalSubscriptionResponseNetworkNodeAreaRestricted
	_ = RegionalSubscriptionResponseTooManyZoneCodes
	_ = RegionalSubscriptionResponseZoneCodesConflict
	_ = RegionalSubscriptionResponseRegionalSubscNotSupported
}

// odbGeneralDataAllSet returns an ODBGeneralData with every bit set —
// useful to stress the full-width encoding path.
func odbGeneralDataAllSet() *ODBGeneralData {
	return &ODBGeneralData{
		AllOGCallsBarred: true, InternationalOGCallsBarred: true,
		InternationalOGCallsNotToHPLMNCountryBarred:                     true,
		PremiumRateInformationOGCallsBarred:                             true,
		PremiumRateEntertainmentOGCallsBarred:                           true,
		SSAccessBarred:                                                  true,
		InterzonalOGCallsBarred:                                         true,
		InterzonalOGCallsNotToHPLMNCountryBarred:                        true,
		InterzonalOGCallsAndInternationalOGCallsNotToHPLMNCountryBarred: true,
		AllECTBarred:                                                    true,
		ChargeableECTBarred:                                             true,
		InternationalECTBarred:                                          true,
		InterzonalECTBarred:                                             true,
		DoublyChargeableECTBarred:                                       true,
		MultipleECTBarred:                                               true,
		AllPacketOrientedServicesBarred:                                 true,
		RoamerAccessToHPLMNAPBarred:                                     true,
		RoamerAccessToVPLMNAPBarred:                                     true,
		RoamingOutsidePLMNOGCallsBarred:                                 true,
		AllICCallsBarred:                                                true,
		RoamingOutsidePLMNICCallsBarred:                                 true,
		RoamingOutsidePLMNICountryICCallsBarred:                         true,
		RoamingOutsidePLMNBarred:                                        true,
		RoamingOutsidePLMNCountryBarred:                                 true,
		RegistrationAllCFBarred:                                         true,
		RegistrationCFNotToHPLMNBarred:                                  true,
		RegistrationInterzonalCFBarred:                                  true,
		RegistrationInterzonalCFNotToHPLMNBarred:                        true,
		RegistrationInternationalCFBarred:                               true,
	}
}

func supportedFeaturesAllSet() *SupportedFeatures {
	return &SupportedFeatures{
		OdbAllApn: true, OdbHPLMNApn: true, OdbVPLMNApn: true, OdbAllOg: true,
		OdbAllInternationalOg:        true,
		OdbAllIntOgNotToHPLMNCountry: true, OdbAllInterzonalOg: true,
		OdbAllInterzonalOgNotToHPLMNCountry:              true,
		OdbAllInterzonalOgAndInternatOgNotToHPLMNCountry: true,
		RegSub: true, Trace: true, LcsAllPrivExcep: true, LcsUniversal: true,
		LcsCallSessionRelated: true, LcsCallSessionUnrelated: true,
		LcsPLMNOperator: true, LcsServiceType: true, LcsAllMOLRSS: true,
		LcsBasicSelfLocation:    true,
		LcsAutonomousSelfLocation: true,
		LcsTransferToThirdParty: true,
		SmMoPp: true, BarringOutgoingCalls: true, Baoc: true, Boic: true,
		BoicExHC: true, LocalTimeZoneRetrieval: true, AdditionalMsisdn: true,
		SmsInMME: true, SmsInSGSN: true, UeReachabilityNotification: true,
		StateLocationInformationRetrieval: true,
		PartialPurge:                      true,
		GddInSGSN:                         true,
		SgsnCAMELCapability:               true,
		PcscfRestoration:                  true,
		DedicatedCoreNetworks:             true,
		NonIPPDNTypeAPNs:                  true,
		NonIPPDPTypeAPNs:                  true,
		NrAsSecondaryRAT:                  true,
	}
}

// Each BIT STRING helper must round-trip: encode → decode returns an
// equivalent struct. We test three cases per type: all-zeros (min bits),
// all-set (max bits), and a spec-guidance-representative single-bit case.
func TestISDBitStrings_RoundTrip(t *testing.T) {
	type caseT struct {
		name   string
		encode func() (runtime.BitString, any)
		decode func(runtime.BitString) any
	}
	cases := []caseT{
		{
			name: "ODBGeneralData/empty",
			encode: func() (runtime.BitString, any) {
				in := &ODBGeneralData{}
				return convertODBGeneralDataToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToODBGeneralData(bs) },
		},
		{
			name: "ODBGeneralData/allSet",
			encode: func() (runtime.BitString, any) {
				in := odbGeneralDataAllSet()
				return convertODBGeneralDataToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToODBGeneralData(bs) },
		},
		{
			name: "ODBGeneralData/onlyHighBit",
			encode: func() (runtime.BitString, any) {
				in := &ODBGeneralData{RegistrationInternationalCFBarred: true}
				return convertODBGeneralDataToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToODBGeneralData(bs) },
		},
		{
			name: "ODBHPLMNData/empty",
			encode: func() (runtime.BitString, any) {
				in := &ODBHPLMNData{}
				return convertODBHPLMNDataToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToODBHPLMNData(bs) },
		},
		{
			name: "ODBHPLMNData/allSet",
			encode: func() (runtime.BitString, any) {
				in := &ODBHPLMNData{true, true, true, true}
				return convertODBHPLMNDataToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToODBHPLMNData(bs) },
		},
		{
			name: "AccessRestrictionData/empty",
			encode: func() (runtime.BitString, any) {
				in := &AccessRestrictionData{}
				return convertAccessRestrictionDataToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToAccessRestrictionData(bs) },
		},
		{
			name: "AccessRestrictionData/allSet",
			encode: func() (runtime.BitString, any) {
				in := &AccessRestrictionData{true, true, true, true, true, true, true, true}
				return convertAccessRestrictionDataToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToAccessRestrictionData(bs) },
		},
		{
			name: "AccessRestrictionData/onlyUtran",
			encode: func() (runtime.BitString, any) {
				in := &AccessRestrictionData{UtranNotAllowed: true}
				return convertAccessRestrictionDataToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToAccessRestrictionData(bs) },
		},
		{
			name: "ExtAccessRestrictionData/empty",
			encode: func() (runtime.BitString, any) {
				in := &ExtAccessRestrictionData{}
				return convertExtAccessRestrictionDataToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToExtAccessRestrictionData(bs) },
		},
		{
			name: "ExtAccessRestrictionData/bothSet",
			encode: func() (runtime.BitString, any) {
				in := &ExtAccessRestrictionData{true, true}
				return convertExtAccessRestrictionDataToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToExtAccessRestrictionData(bs) },
		},
		{
			name: "SupportedFeatures/empty",
			encode: func() (runtime.BitString, any) {
				in := &SupportedFeatures{}
				return convertSupportedFeaturesToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToSupportedFeatures(bs) },
		},
		{
			name: "SupportedFeatures/allSet",
			encode: func() (runtime.BitString, any) {
				in := supportedFeaturesAllSet()
				return convertSupportedFeaturesToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToSupportedFeatures(bs) },
		},
		{
			name: "SupportedFeatures/highBit",
			encode: func() (runtime.BitString, any) {
				in := &SupportedFeatures{NrAsSecondaryRAT: true}
				return convertSupportedFeaturesToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToSupportedFeatures(bs) },
		},
		{
			name: "ExtSupportedFeatures/empty",
			encode: func() (runtime.BitString, any) {
				in := &ExtSupportedFeatures{}
				return convertExtSupportedFeaturesToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToExtSupportedFeatures(bs) },
		},
		{
			name: "ExtSupportedFeatures/bitSet",
			encode: func() (runtime.BitString, any) {
				in := &ExtSupportedFeatures{true}
				return convertExtSupportedFeaturesToBitString(in), in
			},
			decode: func(bs runtime.BitString) any { return convertBitStringToExtSupportedFeatures(bs) },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bs, want := tc.encode()
			got := tc.decode(bs)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("round-trip mismatch:\n got  = %#v\n want = %#v\n bs = %+v", got, want, bs)
			}
		})
	}
}

// Each decoder must tolerate a short-encoded peer BIT STRING without
// panicking or setting out-of-range bits. A well-behaved peer that
// doesn't support the newer feature bits may send fewer bytes than
// the spec minimum; the library's decoders rely on runtime.BitString.Has
// (which returns false past BitLength) so no out-of-bounds read occurs.
// Locks in the Postel's-law claim from the ISD foundation commit.
func TestISDBitStrings_DecodesShortInput(t *testing.T) {
	cases := []struct {
		name string
		bs   runtime.BitString
		got  func(runtime.BitString) any
		want any
	}{
		{
			name: "ODBGeneralData/empty",
			bs:   runtime.BitString{},
			got:  func(bs runtime.BitString) any { return convertBitStringToODBGeneralData(bs) },
			want: &ODBGeneralData{},
		},
		{
			name: "ODBGeneralData/oneBit",
			bs:   runtime.BitString{Bytes: []byte{0x80}, BitLength: 1},
			got:  func(bs runtime.BitString) any { return convertBitStringToODBGeneralData(bs) },
			want: &ODBGeneralData{AllOGCallsBarred: true},
		},
		{
			name: "ODBHPLMNData/empty",
			bs:   runtime.BitString{},
			got:  func(bs runtime.BitString) any { return convertBitStringToODBHPLMNData(bs) },
			want: &ODBHPLMNData{},
		},
		{
			name: "AccessRestrictionData/empty",
			bs:   runtime.BitString{},
			got:  func(bs runtime.BitString) any { return convertBitStringToAccessRestrictionData(bs) },
			want: &AccessRestrictionData{},
		},
		{
			name: "AccessRestrictionData/oneBit",
			bs:   runtime.BitString{Bytes: []byte{0x80}, BitLength: 1},
			got:  func(bs runtime.BitString) any { return convertBitStringToAccessRestrictionData(bs) },
			want: &AccessRestrictionData{UtranNotAllowed: true},
		},
		{
			name: "ExtAccessRestrictionData/empty",
			bs:   runtime.BitString{},
			got:  func(bs runtime.BitString) any { return convertBitStringToExtAccessRestrictionData(bs) },
			want: &ExtAccessRestrictionData{},
		},
		{
			name: "SupportedFeatures/empty",
			bs:   runtime.BitString{},
			got:  func(bs runtime.BitString) any { return convertBitStringToSupportedFeatures(bs) },
			want: &SupportedFeatures{},
		},
		{
			name: "SupportedFeatures/truncatedOneByte",
			bs:   runtime.BitString{Bytes: []byte{0x80}, BitLength: 1},
			got:  func(bs runtime.BitString) any { return convertBitStringToSupportedFeatures(bs) },
			want: &SupportedFeatures{OdbAllApn: true},
		},
		{
			name: "ExtSupportedFeatures/empty",
			bs:   runtime.BitString{},
			got:  func(bs runtime.BitString) any { return convertBitStringToExtSupportedFeatures(bs) },
			want: &ExtSupportedFeatures{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.got(tc.bs)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("short-input decode mismatch:\n got  = %#v\n want = %#v", got, tc.want)
			}
		})
	}
}

// BitLength must satisfy each BIT STRING's spec-min when encoding an
// all-zeros value — a receiver checking the encoded size needs to see
// at least the minimum. Uses the spec bounds from MAP-MS-DataTypes.asn.
func TestISDBitStrings_MinLength(t *testing.T) {
	cases := []struct {
		name   string
		bs     runtime.BitString
		minLen int
	}{
		{"ODBGeneralData", convertODBGeneralDataToBitString(&ODBGeneralData{}), 15},
		{"ODBHPLMNData", convertODBHPLMNDataToBitString(&ODBHPLMNData{}), 4},
		{"AccessRestrictionData", convertAccessRestrictionDataToBitString(&AccessRestrictionData{}), 2},
		{"ExtAccessRestrictionData", convertExtAccessRestrictionDataToBitString(&ExtAccessRestrictionData{}), 1},
		{"SupportedFeatures", convertSupportedFeaturesToBitString(&SupportedFeatures{}), 26},
		{"ExtSupportedFeatures", convertExtSupportedFeaturesToBitString(&ExtSupportedFeatures{}), 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.bs.BitLength < tc.minLen {
				t.Errorf("BitLength=%d, want >=%d", tc.bs.BitLength, tc.minLen)
			}
		})
	}
}
