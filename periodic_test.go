package gsmmap

import (
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// TS 29.002 V19.1.0 §17.7.13: the sender reportingAmount × reportingInterval may not exceed 8639999.
func TestPeriodicProductErrors(t *testing.T) {
	p := &PeriodicLDRInfo{ReportingAmount: 100, ReportingInterval: 86400}
	want := "ProvideSubscriberLocationArg.PeriodicLDRInfo: PeriodicLDRInfo: ReportingAmount(100) × ReportingInterval(86400) = 8640000: " + ErrPeriodicLDRProductExceeded.Error()
	in := &ProvideSubscriberLocationArg{MlcNumber: "12345", IMSI: "001010123456789", PeriodicLDRInfo: p}
	_, err := in.Marshal()
	if err == nil {
		t.Fatal("Marshal accepted excessive product")
	}
	wantEqual(t, "Marshal error", "converting ProvideSubscriberLocationArg: "+want, err.Error())
	imsi := gsm_map.IMSI{0, 1, 1, 0x21, 0x43, 0x65, 0x87, 0xf9}
	w := &gsm_map.ProvideSubscriberLocationArg{Imsi: &imsi, LocationType: gsm_map.LocationType{}, MlcNumber: gsm_map.ISDNAddressString{0x80, 0x21}, PeriodicLDRInfo: &gsm_map.PeriodicLDRInfo{ReportingAmount: 100, ReportingInterval: 86400}}
	got, err := ParseProvideSubscriberLocation(strictBER(t, w))
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "received amount", p.ReportingAmount, got.PeriodicLDRInfo.ReportingAmount)
	wantEqual(t, "received interval", p.ReportingInterval, got.PeriodicLDRInfo.ReportingInterval)
}
