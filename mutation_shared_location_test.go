package gsmmap

import "testing"

// GeographicalInformation carries an uncertainty circle: TS 29.002 V19.1.0 §17.7.1.
// The eight-octet representation is TS 23.032 V19.0.0 §7.3.2.
func TestMutationLocationGeographicalInformation(t *testing.T) {
	uncertainty := uint8(1)
	geo := &GeographicalInfo{ShapeType: ShapeEllipsoidPointUncertainty, Latitude: 0, Longitude: 0, UncertaintyCode: &uncertainty}
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
	for name, got := range map[string]*GeographicalInfo{"EPS": out.SubscriberInfo.LocationInformationEPS.GeographicalInformation, "GPRS": out.SubscriberInfo.LocationInformationGPRS.GeographicalInformation} {
		wantEqual(t, name+" shape", ShapeEllipsoidPointUncertainty, got.ShapeType)
		wantEqual(t, name+" latitude", 0.0, got.Latitude)
		wantEqual(t, name+" longitude", 0.0, got.Longitude)
		wantEqual(t, name+" uncertainty", &uncertainty, got.UncertaintyCode)
	}
}
