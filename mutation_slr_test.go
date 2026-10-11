package gsmmap

import "testing"

// TS 29.002 V19.1.0 §17.7.13: periodicLDRInfo is optional and accuracy value 1 is defined.
func TestMutationSLROptionals(t *testing.T) {
	for _, p := range []*PeriodicLDRInfo{nil, {ReportingAmount: 2, ReportingInterval: 30}} {
		v := AccuracyFulfilmentIndicator(1)
		in := minimalSLRArg()
		in.PeriodicLDRInfo = p
		in.AccuracyFulfilmentIndicator = &v
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseSubscriberLocationReport(data)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		wantEqual(t, "periodic", p, got.PeriodicLDRInfo)
		wantEqual(t, "accuracy", v, *got.AccuracyFulfilmentIndicator)
	}
}
