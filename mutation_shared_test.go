package gsmmap

import "testing"

// TS 29.002 V19.1.0 §17.7.8: CorrelationID's SIP URI A is optional.
func TestMutationSharedCorrelationOptional(t *testing.T) {
	in := &SriSm{MSISDN: "12345", ServiceCentreAddress: "67890", CorrelationID: &SriSmCorrelationID{SipUriB: HexBytes("sip:b@example")}}
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseSriSm(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	wantEqual(t, "correlation", in.CorrelationID, got.CorrelationID)
}
