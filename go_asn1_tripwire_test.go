package gsmmap

import (
	"encoding/hex"
	"testing"
)

// https://github.com/gomaja/go-asn1/issues/82: an implicitly tagged OCTET
// STRING received in constructed form keeps its segment headers on decode
// (ITU-T X.690 §8.7.3, §8.14.4). This pins today's behaviour; when it fails,
// go-asn1 reassembles the value and the test asserts IMSI "00101012345".
func TestGoASN1Issue82ConstructedImplicitOctetString(t *testing.T) {
	// AnyTimeInterrogationArg whose subscriberIdentity imsi [0] carries
	// 00 01 01 21 43 f5 (IMSI 00101012345) in one constructed segment.
	data, err := hex.DecodeString("3016a00aa00804060001012143f5a1028000830491132654")
	if err != nil {
		t.Fatal(err)
	}
	ati, err := ParseAnyTimeInterrogation(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := ati.SubscriberIdentity.IMSI; got != "406000101012345" {
		t.Fatalf("IMSI = %q: go-asn1#82 is resolved; assert \"00101012345\" here and drop the README note", got)
	}
}
