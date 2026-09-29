package gsmmap

import (
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// forwardSMFuzzSeeds are the MO-/MT-ForwardSM-Arg golden captures of
// parse_test.go. Each fuzzer is seeded with all of them, so it also sees
// the other operation's argument.
var forwardSMFuzzSeeds = []string{
	"302d84069122609098998206912260539128041b01510a912260716622000011d972180d4a82eee13928cc7ebbcb20",
	"303c84069122609098998206912260532023042a41400a912260065015000022050003020202e6a0f41c1406b1dfee33a85d9ecfc3e7b20ed40ccbef6137",
	"3042800826610011829761f6840891328490000005f7042c4409d047f6dbfe060000422172514000001d0500035f0202cae8ba5c9e2ecb5de377fb157ea9d1b0d93b1e06",
	"3056800822589172230006f7840891328490000033f00440040d91328471112898f3000062805011948422324f2228e90c42a153500c34a3e1643010cd06a2c570391cc8268bd960a0a213548bc16020015990a6cb62b61a",
	"3067850085000461040b916971101911f900006280517153902159d2e2b1252d467ff6de6c47efd18c38980d27ac1589b9a1d02814d26cb1d92d17a40983311aee560bde6ec39b8d4623da8a4623cd188bc18a441b4c981b0e6d4259ac3814eee4bd5b4d26c3cd6c36",
	"3077800832140080803138f684069169318488880463040b916971101174f40000422182612464805bd2e2b1252d467ff6de6c47efd96eb6a1d056cb0d69b49a10269c098537586e96931965b260d15613da72c29b91261bde72c6a1ad2623d682b5996d58331271375a0d1733eee4bd98ec768bd966b41c0d",
}

func addForwardSMSeeds(f *testing.F) {
	for _, s := range forwardSMFuzzSeeds {
		data, err := hex.DecodeString(s)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Add([]byte{0x30, 0x00})
	f.Add([]byte{0x30, 0x80, 0x00, 0x00})
}

// FuzzParseMtFsm: ParseMtFsm never panics, and a value it returns that
// Marshal accepts decodes back to the same value.
func FuzzParseMtFsm(f *testing.F) {
	addForwardSMSeeds(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		m, err := ParseMtFsm(data)
		if err != nil {
			return
		}
		enc, err := m.Marshal()
		if err != nil {
			return
		}
		again, err := ParseMtFsm(enc)
		if err != nil {
			t.Fatalf("re-parsing marshalled MtFsm: %v", err)
		}
		if diff := cmp.Diff(m, again, equateTPDU); diff != "" {
			t.Fatalf("MtFsm round trip (-first +second):\n%s", diff)
		}
	})
}

// FuzzParseMoFsm: ParseMoFsm never panics, and a value it returns that
// Marshal accepts decodes back to the same value.
func FuzzParseMoFsm(f *testing.F) {
	addForwardSMSeeds(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		m, err := ParseMoFsm(data)
		if err != nil {
			return
		}
		enc, err := m.Marshal()
		if err != nil {
			return
		}
		again, err := ParseMoFsm(enc)
		if err != nil {
			t.Fatalf("re-parsing marshalled MoFsm: %v", err)
		}
		if diff := cmp.Diff(m, again, equateTPDU); diff != "" {
			t.Fatalf("MoFsm round trip (-first +second):\n%s", diff)
		}
	})
}
