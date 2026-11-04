package gsmmap

import (
	"bytes"
	"errors"
	"testing"

	"github.com/gomaja/go-gad"
)

func testCSLocationEPS() *EPSLocationInformation {
	return &EPSLocationInformation{
		EUtranCellGlobalIdentity: HexBytes{0x62, 0xf2, 0x20, 0, 1, 0x23, 0x45},
		TrackingAreaIdentity:     HexBytes{0x62, 0xf2, 0x20, 0x12, 0x34},
	}
}

// TS 29.002 V19.1.0 §17.7.1: LocationInformation carries EPS location at
// [10], inside SubscriberInfo for ATI/PSI and SRI (§17.7.3), or [2] in
// PurgeMS-Arg (§17.7.1). The E-UTRAN CGI is seven octets and the TAI five.
func TestCSLocationInformationEPSRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name  string
		wire  string
		parse func([]byte) (marshaler, error)
	}{
		// CS content: VLR 9 + EPS 18 = 27; SubscriberInfo 29; outer 31.
		{"ATI-Res", "301f301da01b8107911316111111f1aa10800762f22000012345810562f2201234", asParser(ParseAnyTimeInterrogationRes)},
		{"PSI-Res", "301f301da01b8107911316111111f1aa10800762f22000012345810562f2201234", asParser(ParseProvideSubscriberInfoRes)},
		// SRI content: IMSI 10 + SubscriberInfo [7] 31 = 41.
		{"SRI-Res", "a329890800010121436587f9a71da01b8107911316111111f1aa10800762f22000012345810562f2201234", asParser(ParseSriRes)},
		// PurgeMS content: IMSI 10 + CS location [2] 29 = 39.
		{"PurgeMS-Arg", "a327040800010121436587f9a21b8107911316111111f1aa10800762f22000012345810562f2201234", asParser(ParsePurgeMS)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wire := mustHex(t, tc.wire)
			got, err := tc.parse(wire)
			if err != nil {
				t.Fatal(err)
			}
			var loc *CSLocationInformation
			switch v := got.(type) {
			case *AnyTimeInterrogationRes:
				loc = v.SubscriberInfo.LocationInformation
				if v.SubscriberInfo.LocationInformationEPS != nil {
					t.Fatal("nested EPS decoded as SubscriberInfo.LocationInformationEPS")
				}
			case *ProvideSubscriberInfoRes:
				loc = v.SubscriberInfo.LocationInformation
				if v.SubscriberInfo.LocationInformationEPS != nil {
					t.Fatal("nested EPS decoded as SubscriberInfo.LocationInformationEPS")
				}
			case *SriRes:
				wantEqual(t, "IMSI", "001010123456789", v.IMSI)
				if v.SubscriberInfo == nil {
					t.Fatal("missing SubscriberInfo")
				}
				loc = v.SubscriberInfo.LocationInformation
				if v.SubscriberInfo.LocationInformationEPS != nil {
					t.Fatal("nested EPS decoded as SubscriberInfo.LocationInformationEPS")
				}
			case *PurgeMS:
				wantEqual(t, "IMSI", "001010123456789", v.IMSI)
				loc = v.LocationInformation
				if v.LocationInformationEPS != nil {
					t.Fatal("nested EPS decoded as PurgeMS.LocationInformationEPS")
				}
			default:
				t.Fatalf("unexpected parsed type %T", got)
			}
			wantEqual(t, "CS location", &CSLocationInformation{
				VlrNumber: "31611111111", VlrNumberNature: 0x10, VlrNumberPlan: 1,
				LocationInformationEPS: testCSLocationEPS(),
			}, loc)
			encoded, err := got.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(encoded, wire) {
				t.Fatalf("Marshal = %x, want %x", encoded, wire)
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.1: optional locationInformationEPS [10] is
// omitted for nil. EPS content is 9 + 7 = 16 octets; its TLV is 18.
func TestCSLocationInformationEPSMarshal(t *testing.T) {
	for _, tc := range []struct {
		name string
		eps  *EPSLocationInformation
		wire string
	}{
		{"present", testCSLocationEPS(), "301f301da01b8107911316111111f1aa10800762f22000012345810562f2201234"},
		// VLR TLV 9, CS TLV 11, SubscriberInfo TLV 13.
		{"absent", nil, "300d300ba0098107911316111111f1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loc := &CSLocationInformation{
				VlrNumber: "31611111111", VlrNumberNature: 0x10, VlrNumberPlan: 1,
				LocationInformationEPS: tc.eps,
			}
			in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformation: loc}}
			wire, err := in.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if want := mustHex(t, tc.wire); !bytes.Equal(wire, want) {
				t.Fatalf("Marshal = %x, want %x", wire, want)
			}
			got, err := ParseAnyTimeInterrogationRes(wire)
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "CS location", loc, got.SubscriberInfo.LocationInformation)
		})
	}
}

// TS 29.002 V19.1.0 §17.7.1 excludes these six fields beside EPS location.
// The LAI exception concerns MAP-NOTE-MM-EVENT, not these carriers.
func TestCSLocationInformationEPSSenderRule(t *testing.T) {
	for _, tc := range []struct {
		name string
		loc  CSLocationInformation
	}{
		{"CellGlobalId", CSLocationInformation{CellGlobalId: HexBytes{0x62, 0xf2, 0x20, 0x12, 0x34, 0x56, 0x78}}},
		{"LAI", CSLocationInformation{LAI: HexBytes{0x62, 0xf2, 0x20, 0x12, 0x34}}},
		{"AgeOfLocationInformation", CSLocationInformation{AgeOfLocationInformation: new(1)}},
		{"GeographicalInformation", CSLocationInformation{GeographicalInformation: testGeographicalInformation(t, 0, 0, 1)}},
		{"GeodeticInformation", CSLocationInformation{GeodeticInformation: HexBytes{0, 1, 0, 0, 0, 0, 0, 0, 0, 1}}},
		{"CurrentLocationRetrieved", CSLocationInformation{CurrentLocationRetrieved: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loc := tc.loc
			in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformation: &loc}}
			if _, err := in.Marshal(); err != nil {
				t.Fatalf("field without EPS: %v", err)
			}
			loc.LocationInformationEPS = testCSLocationEPS()
			wire, err := in.Marshal()
			if !errors.Is(err, ErrCSLocationInformationEPSConflictingFields) || wire != nil {
				t.Fatalf("Marshal = %x, %v; want nil, %v", wire, err, ErrCSLocationInformationEPSConflictingFields)
			}
		})
	}
	// The same clause leaves these fields available beside EPS location.
	for _, tc := range []struct {
		name string
		loc  CSLocationInformation
	}{
		{"VlrNumber", CSLocationInformation{VlrNumber: "31611111111", VlrNumberNature: 0x10, VlrNumberPlan: 1}},
		{"MscNumber", CSLocationInformation{MscNumber: "31622222222", MscNumberNature: 0x10, MscNumberPlan: 1}},
		{"LocationNumber", CSLocationInformation{LocationNumber: HexBytes{0, 0}}},
		{"SelectedLSAId", CSLocationInformation{SelectedLSAId: HexBytes{0x12, 0x34, 0x56}}},
		{"UserCSGInformation", CSLocationInformation{UserCSGInformation: &UserCSGInformation{CsgID: HexBytes{0x12, 0x34, 0x56, 0x60}, CsgIDBits: 27}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loc := tc.loc
			loc.LocationInformationEPS = testCSLocationEPS()
			in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{LocationInformation: &loc}}
			wire, err := in.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			got, err := ParseAnyTimeInterrogationRes(wire)
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "CS location", &loc, got.SubscriberInfo.LocationInformation)
		})
	}
}

// TS 29.002 V19.1.0 §17.7.1: the sender restriction must not discard a
// received LAI or nested EPS location. LAI [3] wraps a seven-octet CHOICE
// TLV: 9 + 18 = 27 CS content, 29 SubscriberInfo content, 31 outer content.
func TestCSLocationInformationEPSParseKeepsLAI(t *testing.T) {
	got, err := ParseAnyTimeInterrogationRes(mustHex(t, "301f301da01ba307810562f2201234aa10800762f22000012345810562f2201234"))
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "received CS location", &CSLocationInformation{
		LAI: HexBytes{0x62, 0xf2, 0x20, 0x12, 0x34}, LocationInformationEPS: testCSLocationEPS(),
	}, got.SubscriberInfo.LocationInformation)
	wire, err := got.Marshal()
	if !errors.Is(err, ErrCSLocationInformationEPSConflictingFields) || wire != nil {
		t.Fatalf("Marshal = %x, %v; want nil, %v", wire, err, ErrCSLocationInformationEPSConflictingFields)
	}
}

// testGeographicalInformation codes an ellipsoid point with uncertainty
// circle, the only shape GeographicalInformation carries (TS 29.002 V19.1.0
// §17.7.1), in the eight octets of TS 23.032 V19.0.0 §7.3.2.
func testGeographicalInformation(t *testing.T, lat, lon float64, uncertainty gad.Uncertainty) HexBytes {
	t.Helper()
	point, err := gad.PointFromDegrees(lat, lon)
	if err != nil {
		t.Fatal(err)
	}
	b, err := gad.EllipsoidPointUncertaintyCircle{Point: point, Uncertainty: uncertainty}.Encode()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// GeographicalInformation carries an uncertainty circle: TS 29.002 V19.1.0 §17.7.1.
// The eight-octet representation is TS 23.032 V19.0.0 §7.3.2.
func TestLocationGeographicalInformation(t *testing.T) {
	geo := testGeographicalInformation(t, 0, 0, 1)
	in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{
		LocationInformationEPS:  &EPSLocationInformation{GeographicalInformation: geo},
		LocationInformationGPRS: &GPRSLocationInformation{GeographicalInformation: geo},
	}}
	data, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatal(err)
	}
	for name, got := range map[string]HexBytes{"EPS": out.SubscriberInfo.LocationInformationEPS.GeographicalInformation, "GPRS": out.SubscriberInfo.LocationInformationGPRS.GeographicalInformation} {
		if !bytes.Equal(got, geo) {
			t.Errorf("%s: got %x, want %x", name, got, geo)
		}
		d, err := gad.Decode(got)
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, name+" description", gad.Description(gad.EllipsoidPointUncertaintyCircle{Uncertainty: 1}), d)
	}
}
