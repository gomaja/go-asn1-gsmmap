package gsmmap

import "testing"

// Bandwidth and BandwidthExt include zero: TS 29.002 V19.1.0 §17.7.1.
func TestMutationAMBRZeroBandwidth(t *testing.T) {
	zero := int64(0)
	in := &InsertSubscriberDataArg{EpsSubscriptionData: &EPSSubscriptionData{Ambr: &AMBR{
		MaxRequestedBandwidthUL: 0, MaxRequestedBandwidthDL: 0,
		ExtendedMaxRequestedBandwidthUL: &zero, ExtendedMaxRequestedBandwidthDL: &zero,
	}}}
	data, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ParseInsertSubscriberData(data)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "zero bandwidth", in, out)
}
