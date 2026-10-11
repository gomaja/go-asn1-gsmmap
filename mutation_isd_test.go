package gsmmap

import (
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func mutationISDWire(t *testing.T, in *InsertSubscriberDataArg) *gsm_map.InsertSubscriberDataArg {
	t.Helper()
	data, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var wire gsm_map.InsertSubscriberDataArg
	if err := wire.UnmarshalBER(data); err != nil {
		t.Fatal(err)
	}
	return &wire
}

// VPLMN CSG subscription is OPTIONAL: TS 29.002 V19.1.0 §17.7.1.
func TestMutationISDVPLMNCSGPresence(t *testing.T) {
	for _, list := range []VPLMNCSGSubscriptionDataList{nil, {{CsgID: HexBytes{0x12, 0x34, 0x56, 0x60}, CsgIDBits: 27}}} {
		in := &InsertSubscriberDataArg{VplmnCsgSubscriptionDataList: list}
		data, err := in.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		out, err := ParseInsertSubscriberData(data)
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "VPLMN CSG presence", in, out)
	}
}
