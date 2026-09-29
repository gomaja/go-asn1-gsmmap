// address_nature_plan_test.go
//
// Nature of address and numbering plan handling for AddressString fields
// (3GPP TS 29.002 V19.1.0 §17.7.8): zero is "unknown" exactly as on the wire,
// a decoded address encodes back to the same octets, and out-of-range values
// are rejected on encode.

package gsmmap

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/gomaja/go-asn1-gsmmap/address"
)

// natPlanCases are the AddressString first octets exercised by the lossless
// round-trip tests: no-extension bit set, then nature (bits 7..5) and plan
// (bits 4..1).
var natPlanCases = []struct {
	name         string
	first        byte
	nature, plan uint8
}{
	{"unknown nature, unknown plan (0x80)", 0x80, address.NatureUnknown, address.PlanUnknown},
	{"unknown nature, ISDN plan (0x81)", 0x81, address.NatureUnknown, address.PlanISDN},
	{"international nature, unknown plan (0x90)", 0x90, address.NatureInternational, address.PlanUnknown},
	{"subscriber nature, national plan (0xC8)", 0xC8, address.NatureSubscriber, address.PlanNational},
	{"international nature, ISDN plan (0x91)", 0x91, address.NatureInternational, address.PlanISDN},
}

func natPlanMustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("hex decode %q: %v", s, err)
	}
	return b
}

// patchOnce replaces the single occurrence of old in data with repl.
func patchOnce(t *testing.T, data, old, repl []byte) []byte {
	t.Helper()
	if n := bytes.Count(data, old); n != 1 {
		t.Fatalf("pattern %x occurs %d times in %x, want exactly 1", old, n, data)
	}
	return bytes.Replace(data, old, repl, 1)
}

// assertMarshalReproduces marshals and requires the exact input octets.
func assertMarshalReproduces(t *testing.T, want []byte, marshal func() ([]byte, error)) {
	t.Helper()
	got, err := marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("Marshal did not reproduce the input\n got %x\nwant %x", got, want)
	}
}

func TestUSSDArgMSISDNNaturePlanLossless(t *testing.T) {
	base := &USSDArg{
		DataCodingScheme: USSDDataCodingSchemeGSM7,
		USSDString:       ussdText(t, USSDDataCodingSchemeGSM7, "*100#"),
		MSISDN:           "31612345678",
		MSISDNNature:     address.NatureInternational,
		MSISDNPlan:       address.PlanISDN,
	}
	baseWire, err := base.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	addr, err := encodeAddressField(base.MSISDN, address.NatureInternational, address.PlanISDN)
	if err != nil {
		t.Fatal(err)
	}
	// msisdn [0] carries the address octets verbatim.
	if !bytes.Contains(baseWire, append([]byte{0x80, byte(len(addr))}, addr...)) {
		t.Fatalf("msisdn [0] not found in %x", baseWire)
	}
	for _, c := range natPlanCases {
		t.Run(c.name, func(t *testing.T) {
			patched := append([]byte{c.first}, addr[1:]...)
			wire := patchOnce(t, baseWire, addr, patched)
			got, err := ParseUSSDArg(wire)
			if err != nil {
				t.Fatalf("ParseUSSDArg: %v", err)
			}
			if got.MSISDN != base.MSISDN || got.MSISDNNature != c.nature || got.MSISDNPlan != c.plan {
				t.Errorf("got MSISDN=%q nature=0x%02X plan=0x%02X, want %q 0x%02X 0x%02X",
					got.MSISDN, got.MSISDNNature, got.MSISDNPlan, base.MSISDN, c.nature, c.plan)
			}
			assertMarshalReproduces(t, wire, got.Marshal)
		})
	}
}

func TestMtFsmSmRpOaNaturePlanLossless(t *testing.T) {
	// Golden MT-ForwardSM (parse_test.go); SM-RP-OA is serviceCentreAddressOA
	// [4] = 84 06 91 69 31 84 88 88.
	golden := natPlanMustHex(t, knownMtFsmHex)
	oldAddr := natPlanMustHex(t, "8406916931848888")
	for _, c := range natPlanCases {
		t.Run(c.name, func(t *testing.T) {
			repl := append([]byte(nil), oldAddr...)
			repl[2] = c.first
			wire := patchOnce(t, golden, oldAddr, repl)
			got, err := ParseMtFsm(wire)
			if err != nil {
				t.Fatalf("ParseMtFsm: %v", err)
			}
			if got.SmRpOa.ServiceCentreAddressOA == "" {
				t.Fatalf("serviceCentreAddressOA not decoded: %+v", got.SmRpOa)
			}
			if got.SmRpOa.SCAOANature != c.nature || got.SmRpOa.SCAOAPlan != c.plan {
				t.Errorf("got nature=0x%02X plan=0x%02X, want 0x%02X 0x%02X",
					got.SmRpOa.SCAOANature, got.SmRpOa.SCAOAPlan, c.nature, c.plan)
			}
			assertMarshalReproduces(t, wire, got.Marshal)
		})
	}
}

func TestSriSmNaturePlanLossless(t *testing.T) {
	// Golden SRI-SM (parse_test.go): msisdn [0] 91 22 60 85 38 18,
	// sm-RP-PRI [1] TRUE, serviceCentreAddress [2] 91 22 60 90 98 99.
	const golden = "301380069122608538188101ff8206912260909899"
	for _, c := range natPlanCases {
		t.Run(c.name, func(t *testing.T) {
			wire := natPlanMustHex(t, golden)
			wire = patchOnce(t, wire, natPlanMustHex(t, "8006912260853818"), natPlanMustHex(t, "8006"+hex.EncodeToString([]byte{c.first})+"2260853818"))
			wire = patchOnce(t, wire, natPlanMustHex(t, "8206912260909899"), natPlanMustHex(t, "8206"+hex.EncodeToString([]byte{c.first})+"2260909899"))
			got, err := ParseSriSm(wire)
			if err != nil {
				t.Fatalf("ParseSriSm: %v", err)
			}
			if got.MSISDNNature != c.nature || got.MSISDNPlan != c.plan {
				t.Errorf("MSISDN nature=0x%02X plan=0x%02X, want 0x%02X 0x%02X", got.MSISDNNature, got.MSISDNPlan, c.nature, c.plan)
			}
			if got.SCANature != c.nature || got.SCAPlan != c.plan {
				t.Errorf("SCA nature=0x%02X plan=0x%02X, want 0x%02X 0x%02X", got.SCANature, got.SCAPlan, c.nature, c.plan)
			}
			assertMarshalReproduces(t, wire, got.Marshal)
		})
	}
}

func TestUpdateLocationResNaturePlanLossless(t *testing.T) {
	base := &UpdateLocationRes{
		HLRNumber:       "31612345678",
		HLRNumberNature: address.NatureInternational,
		HLRNumberPlan:   address.PlanISDN,
	}
	baseWire, err := base.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	addr, err := encodeAddressField(base.HLRNumber, address.NatureInternational, address.PlanISDN)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range natPlanCases {
		t.Run(c.name, func(t *testing.T) {
			wire := patchOnce(t, baseWire, addr, append([]byte{c.first}, addr[1:]...))
			got, err := ParseUpdateLocationRes(wire)
			if err != nil {
				t.Fatalf("ParseUpdateLocationRes: %v", err)
			}
			if got.HLRNumber != base.HLRNumber || got.HLRNumberNature != c.nature || got.HLRNumberPlan != c.plan {
				t.Errorf("got %q 0x%02X 0x%02X, want %q 0x%02X 0x%02X",
					got.HLRNumber, got.HLRNumberNature, got.HLRNumberPlan, base.HLRNumber, c.nature, c.plan)
			}
			assertMarshalReproduces(t, wire, got.Marshal)
		})
	}
}

// Zero nature and plan encode as 0x80: no extension, unknown, unknown.
func TestEncodeAddressFieldZeroIsUnknown(t *testing.T) {
	got, err := encodeAddressField("31612345678", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != 0x80 {
		t.Errorf("first octet = 0x%02X, want 0x80", got[0])
	}
	if got[0] == 0x91 {
		t.Error("zero nature/plan must not be rewritten to international/ISDN")
	}
}

func TestEncodeAddressFieldRejectsInvalidNature(t *testing.T) {
	for _, n := range []uint8{0x01, 0x05, 0x0F, 0x11, 0x80, 0x90, 0xF0, 0xFF} {
		_, err := encodeAddressField("31612345678", n, address.PlanISDN)
		if !errors.Is(err, ErrAddressNatureInvalid) {
			t.Errorf("nature 0x%02X: err = %v, want ErrAddressNatureInvalid", n, err)
		}
		if errors.Is(err, ErrAddressPlanInvalid) {
			t.Errorf("nature 0x%02X: err must not be ErrAddressPlanInvalid", n)
		}
	}
}

func TestEncodeAddressFieldRejectsInvalidPlan(t *testing.T) {
	for _, p := range []uint8{0x10, 0x11, 0x1F, 0x80, 0xFF} {
		_, err := encodeAddressField("31612345678", address.NatureInternational, p)
		if !errors.Is(err, ErrAddressPlanInvalid) {
			t.Errorf("plan 0x%02X: err = %v, want ErrAddressPlanInvalid", p, err)
		}
		if errors.Is(err, ErrAddressNatureInvalid) {
			t.Errorf("plan 0x%02X: err must not be ErrAddressNatureInvalid", p)
		}
	}
}

func TestEncodeAddressFieldAcceptsEveryConstant(t *testing.T) {
	natures := []uint8{
		address.NatureUnknown, address.NatureInternational, address.NatureNational,
		address.NatureNetworkSpecific, address.NatureSubscriber, address.NatureReserved,
		address.NatureAbbreviated, address.NatureReservedExtension,
	}
	plans := []uint8{
		address.PlanUnknown, address.PlanISDN, address.PlanData, address.PlanTelex,
		address.PlanLandMobile, address.PlanNational, address.PlanPrivate,
		address.PlanReservedExtension,
	}
	for _, n := range natures {
		for _, p := range plans {
			got, err := encodeAddressField("31612345678", n, p)
			if err != nil {
				t.Errorf("nature 0x%02X plan 0x%02X: %v", n, p, err)
				continue
			}
			if want := byte(0x80) | n | p; got[0] != want {
				t.Errorf("nature 0x%02X plan 0x%02X: first octet 0x%02X, want 0x%02X", n, p, got[0], want)
			}
		}
	}
	// Every 4-bit plan value is on the wire and must be accepted.
	for p := uint8(0); p <= 0x0F; p++ {
		if _, err := encodeAddressField("1", address.NatureUnknown, p); err != nil {
			t.Errorf("plan 0x%02X: %v", p, err)
		}
	}
}

func TestMarshalRejectsInvalidNaturePlan(t *testing.T) {
	gsm := ussdText(t, USSDDataCodingSchemeGSM7, "*100#")
	t.Run("USSDArg nature", func(t *testing.T) {
		_, err := (&USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: gsm, MSISDN: "31612345678", MSISDNNature: 0x05}).Marshal()
		if !errors.Is(err, ErrAddressNatureInvalid) {
			t.Errorf("err = %v, want ErrAddressNatureInvalid", err)
		}
	})
	t.Run("USSDArg plan", func(t *testing.T) {
		_, err := (&USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: gsm, MSISDN: "31612345678", MSISDNPlan: 0x10}).Marshal()
		if !errors.Is(err, ErrAddressPlanInvalid) {
			t.Errorf("err = %v, want ErrAddressPlanInvalid", err)
		}
	})
	t.Run("SriSm nature", func(t *testing.T) {
		_, err := (&SriSm{MSISDN: "31612345678", MSISDNNature: 0x80, ServiceCentreAddress: "31611111111"}).Marshal()
		if !errors.Is(err, ErrAddressNatureInvalid) {
			t.Errorf("err = %v, want ErrAddressNatureInvalid", err)
		}
	})
}
