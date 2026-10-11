package gsmmap

import (
	"fmt"
	"testing"
)

// DefaultGPRS-Handling lists 0 and 1: TS 29.002 V19.1.0 §17.7.1.
func TestGPRSReleaseTransaction(t *testing.T) {
	in := camelISD()
	in.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList[0].DefaultSessionHandling = DefaultGPRSReleaseTransaction
	data, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ParseInsertSubscriberData(data)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "release transaction", in, out)
}

// DefaultGPRS-Handling receiver exceptions: TS 29.002 V19.1.0 §17.7.1.
func TestGPRSDefaultMapping(t *testing.T) {
	cases := []struct{ wire, want DefaultGPRSHandling }{{0, 0}, {1, 1}, {2, 0}, {31, 0}, {32, 1}}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.wire), func(t *testing.T) {
			w := marshalledISDWire(t, camelISD())
			w.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList.Values[0].DefaultSessionHandling = tc.wire
			out := parseISD(t, w)
			wantEqual(t, "default GPRS handling", tc.want, out.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList[0].DefaultSessionHandling)
		})
	}
}
