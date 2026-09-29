// ussd_tripwire_test.go
//
// Tripwires for a known gap in the go-asn1 dependency: OCTET STRING SIZE
// constraints are not enforced on encode or decode. The USSD codec does not
// re-check them; these tests record today's behaviour and fail once go-asn1
// enforces the constraint, at which point they must assert rejection instead.
// Tracked upstream in https://github.com/gomaja/go-asn1/issues/64.

package gsmmap

import (
	"bytes"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func TestUSSDStringSizeTripwire(t *testing.T) {
	for _, n := range []int{0, 161} {
		str := bytes.Repeat([]byte{0x41}, n)

		// go-asn1 does not enforce SIZE (1..maxUSSD-StringLength); this test fails once it does — then assert rejection instead
		arg := &USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: str}
		data, err := arg.Marshal()
		if err != nil {
			t.Errorf("USSDArg.Marshal with %d-octet USSD-String: %v", n, err)
			continue
		}
		back, err := ParseUSSDArg(data)
		if err != nil {
			t.Errorf("ParseUSSDArg with %d-octet USSD-String: %v", n, err)
		} else if len(back.USSDString) != n {
			t.Errorf("ParseUSSDArg kept %d octets, want %d", len(back.USSDString), n)
		}

		// go-asn1 does not enforce SIZE (1..maxUSSD-StringLength); this test fails once it does — then assert rejection instead
		res := &USSDRes{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: str}
		data, err = res.Marshal()
		if err != nil {
			t.Errorf("USSDRes.Marshal with %d-octet USSD-String: %v", n, err)
			continue
		}
		if _, err := ParseUSSDRes(data); err != nil {
			t.Errorf("ParseUSSDRes with %d-octet USSD-String: %v", n, err)
		}
	}
}

func TestUSSDMSISDNSizeTripwire(t *testing.T) {
	// go-asn1 does not enforce SIZE (1..maxISDN-AddressLength); this test fails once it does — then assert rejection instead
	// 19 digits + the nature/plan octet = 11 octets, above the 9-octet maximum.
	m := gsm_map.ISDNAddressString(append([]byte{0x91}, bytes.Repeat([]byte{0x21}, 10)...))
	data, err := (&gsm_map.USSDArg{UssdDataCodingScheme: []byte{0x0F}, UssdString: []byte{0x31}, Msisdn: &m}).MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	if _, err := ParseUSSDArg(data); err != nil {
		t.Errorf("ParseUSSDArg with an 11-octet MSISDN: %v", err)
	}
}
