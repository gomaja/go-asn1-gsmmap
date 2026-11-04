package gsmmap

import (
	"errors"
	"math"
	"math/big"
	"strings"
	"testing"
)

// int64FromBigInt implements the public int64 boundary for unconstrained
// INTEGERs (TS 29.002 V19.1.0 §17.7.1). No public Parse path reaches its
// nil case: the codec rejects missing mandatory INTEGERs and converters
// skip absent optional INTEGERs. Test ErrIntegerMissing directly here.
func TestInt64FromBigInt(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   *big.Int
		want int64
		err  error
	}{
		{"missing", nil, 0, ErrIntegerMissing},
		{"zero", big.NewInt(0), 0, nil},
		{"maximum", big.NewInt(math.MaxInt64), math.MaxInt64, nil},
		{"minimum", big.NewInt(math.MinInt64), math.MinInt64, nil},
		{"positive overflow", new(big.Int).Lsh(big.NewInt(1), 63), 0, ErrInt64Overflow},
		{"negative overflow", new(big.Int).Sub(big.NewInt(math.MinInt64), big.NewInt(1)), 0, ErrInt64Overflow},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := int64FromBigInt(tc.in, "priority-level")
			if got != tc.want || !errors.Is(err, tc.err) {
				t.Fatalf("int64FromBigInt = %d, %v; want %d, %v", got, err, tc.want, tc.err)
			}
			if err != nil && !strings.Contains(err.Error(), "priority-level") {
				t.Errorf("error must name priority-level: %v", err)
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.8: CorrelationID's SIP URI A is optional.
func TestSharedCorrelationOptional(t *testing.T) {
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
