package gsmmap

import "testing"

// TS 29.002 V19.1.0 §17.7.1: PDN-GW identity can contain IPv4, IPv6, or a name.
func TestUpdateGprsPDNIdentity(t *testing.T) {
	contextID := 1
	for _, tc := range []struct {
		name string
		id   *PdnGwIdentity
	}{
		{"IPv4", &PdnGwIdentity{IPv4Address: HexBytes{192, 0, 2, 1}}},
		{"IPv6", &PdnGwIdentity{IPv6Address: HexBytes{0x20, 1, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}}},
		{"name", &PdnGwIdentity{Name: HexBytes("pgw.example")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := &UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "12345", SGSNAddress: "192.0.2.2", EpsInfo: &EpsInfo{PdnGwUpdate: &PdnGwUpdate{PdnGwIdentity: tc.id, ContextID: &contextID}}}
			data, err := in.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			got, err := ParseUpdateGprsLocation(data)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			wantEqual(t, "PDN update", in.EpsInfo, got.EpsInfo)
		})
	}
}
