package gsmmap

import (
	"errors"
	"testing"

	"github.com/gomaja/go-sms/encoding/tpdu"
)

// TS 29.002 V19.1.0 §17.7.6: each SM-RP address selects one CHOICE alternative.
func TestMoFsmChoices(t *testing.T) {
	for _, tc := range []struct {
		name string
		da   SmRpDa
		oa   SmRpOa
	}{
		{"no addresses", SmRpDa{NoSmRpDa: true}, SmRpOa{NoSmRpOa: true}},
		{"LMSI", SmRpDa{LMSI: HexBytes{1, 2, 3, 4}}, SmRpOa{NoSmRpOa: true}},
		{"service centres", SmRpDa{ServiceCentreAddressDA: "12345"}, SmRpOa{ServiceCentreAddressOA: "67890"}},
		{"subscriber", SmRpDa{IMSI: "001010123456789"}, SmRpOa{MSISDN: "12345"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := &MoFsm{SmRpDa: tc.da, SmRpOa: tc.oa, TPDU: tpdu.TPDU{Direction: tpdu.MO, FirstOctet: 1, DA: tpdu.Address{Addr: "12345"}}}
			data, err := in.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			got, err := ParseMoFsm(data)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			wantEqual(t, "DA", tc.da, got.SmRpDa)
			wantEqual(t, "OA", tc.oa, got.SmRpOa)
		})
	}
	if _, err := (&MoFsm{SmRpDa: SmRpDa{ServiceCentreAddressDA: "12345", NoSmRpDa: true}}).Marshal(); !errors.Is(err, ErrSmRpDaMultipleAlternatives) {
		t.Fatalf("multiple DA error: got %v, want %v", err, ErrSmRpDaMultipleAlternatives)
	}
}
