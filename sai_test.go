package gsmmap

import (
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// TS 29.002 V19.1.0 §17.7.1: RequestingNodeType maps 6..15 to VLR and >17 to SGSN.
func TestSAINodeBoundaries(t *testing.T) {
	for _, tc := range []struct{ raw, want RequestingNodeType }{{5, 5}, {6, 0}, {15, 0}, {16, 16}, {17, 17}, {18, 1}} {
		v := tc.raw
		w := &gsm_map.SendAuthenticationInfoArg{Imsi: gsm_map.IMSI{0, 0x11, 0x10, 0x32, 0x54, 0x76, 0x98, 0xF0}, NumberOfRequestedVectors: 1, RequestingNodeType: &v}
		got, err := ParseSendAuthenticationInfo(strictBER(t, w))
		if err != nil {
			t.Fatalf("raw %d: %v", v, err)
		}
		wantEqual(t, "RequestingNodeType", tc.want, *got.RequestingNodeType)
	}
}
