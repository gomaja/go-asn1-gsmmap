package gsmmap

import (
	"bytes"
	"testing"

	"github.com/gomaja/go-gad"
)

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
