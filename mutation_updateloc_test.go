package gsmmap

import (
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// TS 29.002 V19.1.0 §17.7.1: LocationArea chooses LAI or LAC; basicISTSupported(0) is defined.
func TestMutationUpdateLocationPaging(t *testing.T) {
	v := ISTSupportIndicatorBasicISTSupported
	for _, la := range []gsm_map.LocationArea{gsm_map.NewLocationAreaLac(gsm_map.LAC{1, 2}), gsm_map.NewLocationAreaLaiFixedLength(gsm_map.LAIFixedLength{1, 2, 3, 4, 5})} {
		raw := strictBER(t, &la)
		in := &UpdateLocation{IMSI: "001010123456789", MscNumber: "12345", VlrNumber: "67890", PagingArea: []HexBytes{raw}, VlrCapability: &VlrCapability{IstSupportIndicator: &v}}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseUpdateLocation(data)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		wantEqual(t, "paging area", in.PagingArea, got.PagingArea)
		wantEqual(t, "IST", v, *got.VlrCapability.IstSupportIndicator)
	}
}
