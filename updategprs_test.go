package gsmmap

import (
	"errors"
	"strings"
	"testing"

	"github.com/gomaja/go-asn1-gsmmap/address"
)

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

// TS 29.002 V19.1.0 §17.7.1, PDN-GW-Update: without pdn-gw-Identity the
// HSS ignores the entire EPS-Info, including a supplied APN or contextId.
func TestUpdateGprsParseEpsInfoWithoutPDNGWIdentity(t *testing.T) {
	for _, tc := range []struct {
		name string
		wire string
		want error
	}{
		// Each argument has 22 octets of mandatory TLVs, 8100 before EPS-Info
		// and 8700 after it. EPS-Info [5] explicitly wraps PDN-GW-Update [0].
		// Empty: [5] content 2, outer content 22+2+4+2 = 30.
		{"empty", "301e040800010121436587f90403912143040504c00002018100a502a0008700", nil},
		// APN TLV 4: [0] content 4, [5] content 6, outer content 34.
		{"APN only", "3022040800010121436587f90403912143040504c00002018100a506a004800261708700", nil},
		// ContextId TLV 3: [0] content 3, [5] content 5, outer content 33.
		{"contextId only", "3021040800010121436587f90403912143040504c00002018100a505a0038201018700", nil},
		// Identity [1] content 5 (IPv4 TLV 2+3); [0] content 7,
		// [5] content 9, outer content 37. PDP-Address SIZE admits three
		// octets; the identity converter must reject this IPv4 address.
		{"short IPv4", "3025040800010121436587f90403912143040504c00002018100a509a007a1058003c000028700", ErrPdnGwIdentityIPv4AddressInvalidLength},
		// Identity [1] is empty: [0] content 2, [5] content 4, outer 32.
		{"identity without address", "3020040800010121436587f90403912143040504c00002018100a504a002a1008700", ErrPdnGwIdentityAddressMissing},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseUpdateGprsLocation(mustHex(t, tc.wire))
			if tc.want != nil {
				if got != nil || !errors.Is(err, tc.want) {
					t.Fatalf("Parse = %+v, %v; want nil, %v", got, err, tc.want)
				}
				if !strings.Contains(err.Error(), "EpsInfo") || !strings.Contains(err.Error(), "PdnGwIdentity") {
					t.Errorf("error must name EPS info and its identity: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			wantEqual(t, "argument with EPS info ignored", &UpdateGprsLocation{
				IMSI: "001010123456789", SgsnNumber: "1234",
				SgsnNumberNature: address.NatureInternational, SgsnNumberPlan: address.PlanISDN,
				SGSNAddress: "192.0.2.1", InformPreviousNetworkEntity: true, SkipSubscriberDataUpdate: true,
			}, got)
		})
	}
}
