package gsmmap

import "testing"

// TS 29.002 V19.1.0 §17.7.3: UnavailabilityCause defines values 1 through 6.
func TestMutationSriResponseBoundaries(t *testing.T) {
	for _, v := range []UnavailabilityCause{1, 6} {
		in := &SriResp{UnavailabilityCause: &v, SubscriberInfo: &SubscriberInfo{}, ExtendedRoutingInfo: &ExtendedRoutingInfo{CamelRoutingInfo: &CamelRoutingInfo{ForwardingData: &ForwardingData{ForwardedToNumber: "12345"}}}}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal %d: %v", v, err)
		}
		got, err := ParseSriResp(data)
		if err != nil {
			t.Fatalf("Parse %d: %v", v, err)
		}
		wantEqual(t, "UnavailabilityCause", v, *got.UnavailabilityCause)
		wantEqual(t, "SubscriberInfo", in.SubscriberInfo, got.SubscriberInfo)
		wantEqual(t, "routing info", in.ExtendedRoutingInfo, got.ExtendedRoutingInfo)
	}
}

// TS 29.002 V19.1.0 §§17.7.1,17.7.3: basicISTSupported(0) is defined.
func TestMutationSriISTZero(t *testing.T) {
	v := ISTSupportIndicatorBasicISTSupported
	in := &Sri{MSISDN: "12345", GmscOrGsmSCFAddress: "67890", IstSupportIndicator: &v}
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseSri(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	wantEqual(t, "IST", v, *got.IstSupportIndicator)
}
