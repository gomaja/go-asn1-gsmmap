package gsmmap

import (
	"encoding/hex"
	"errors"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// An SM-RP-DA or SM-RP-OA address alternative can be present on the wire
// yet carry no digits. These cases check the corresponding parse errors for
// one-octet AddressStrings.
func TestForwardSMChoiceAddressDecodedEmpty(t *testing.T) {
	golden, err := hex.DecodeString(forwardSMFuzzSeeds[0]) // MO-ForwardSM-Arg
	if err != nil {
		t.Fatal(err)
	}
	var base gsm_map.MOForwardSMArg
	if err := base.UnmarshalBER(golden); err != nil {
		t.Fatal(err)
	}
	natureOnly := []byte{0x91}
	cases := []struct {
		name string
		edit func(a *gsm_map.MOForwardSMArg)
		want error
	}{
		{"SM-RP-DA serviceCentreAddressDA without digits", func(a *gsm_map.MOForwardSMArg) {
			a.SmRPDA = gsm_map.NewSMRPDAServiceCentreAddressDA(natureOnly)
		}, ErrSmRpDaServiceCentreAddressDecodedEmpty},
		{"SM-RP-OA msisdn without digits", func(a *gsm_map.MOForwardSMArg) {
			a.SmRPOA = gsm_map.NewSMRPOAMsisdn(natureOnly)
		}, ErrSmRpOaMSISDNDecodedEmpty},
		{"SM-RP-OA serviceCentreAddressOA without digits", func(a *gsm_map.MOForwardSMArg) {
			a.SmRPOA = gsm_map.NewSMRPOAServiceCentreAddressOA(natureOnly)
		}, ErrSmRpOaServiceCentreAddressDecodedEmpty},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			arg := base
			tc.edit(&arg)
			data, err := arg.MarshalBER()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParseMoFsm(data); !errors.Is(err, tc.want) {
				t.Errorf("ParseMoFsm: err = %v, want %v", err, tc.want)
			}
			// The same CHOICE types and converters serve MT-ForwardSM-Arg.
			mt := gsm_map.MTForwardSMArg{SmRPDA: arg.SmRPDA, SmRPOA: arg.SmRPOA, SmRPUI: arg.SmRPUI}
			mtData, err := mt.MarshalBER()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParseMtFsm(mtData); !errors.Is(err, tc.want) {
				t.Errorf("ParseMtFsm: err = %v, want %v", err, tc.want)
			}
		})
	}
}
