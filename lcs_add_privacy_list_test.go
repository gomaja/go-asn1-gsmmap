package gsmmap

import (
	"errors"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// 3GPP TS 29.002 V19.1.0 §17.7.1 LCSInformation: "add-lcs-PrivacyExceptionList
// may be sent only if lcs-PrivacyExceptionList is present and contains four
// instances of LCS-PrivacyClass. If the mentioned condition is not satisfied
// the receiving node shall discard add-lcs-PrivacyExceptionList. If an
// LCS-PrivacyClass is received both in lcs-PrivacyExceptionList and in
// add-lcs-PrivacyExceptionList with the same SS-Code, then the error
// unexpected data value shall be returned."

// LCS privacy SS-Codes (3GPP TS 29.002 V19.1.0 §17.7.5).
const (
	ssUniversal            SsCode = 0xB1
	ssCallSessionRelated   SsCode = 0xB2
	ssCallSessionUnrelated SsCode = 0xB3
	ssPLMNOperator         SsCode = 0xB4
	ssServiceType          SsCode = 0xB5
)

func lcsPrivacyClasses(codes ...SsCode) LCSPrivacyExceptionList {
	out := make(LCSPrivacyExceptionList, len(codes))
	for i, c := range codes {
		out[i] = LCSPrivacyClass{SsCode: c, SsStatus: HexBytes{0x05}}
	}
	return out
}

func wireLCSPrivacyClasses(codes ...SsCode) *gsm_map.LCSPrivacyExceptionList {
	if len(codes) == 0 {
		return nil
	}
	l := &gsm_map.LCSPrivacyExceptionList{}
	for _, c := range codes {
		l.Values = append(l.Values, gsm_map.LCSPrivacyClass{SsCode: []byte{byte(c)}, SsStatus: []byte{0x05}})
	}
	return l
}

var fullLCSPrivacyList = []SsCode{ssUniversal, ssCallSessionRelated, ssCallSessionUnrelated, ssPLMNOperator}

func TestMarshalAddLCSPrivacyExceptionList(t *testing.T) {
	for _, tc := range []struct {
		name      string
		base, add []SsCode
		want      error
	}{
		{"without the list", nil, []SsCode{ssServiceType}, ErrLCSAddPrivacyExceptionListNotAllowed},
		{"with three classes", fullLCSPrivacyList[:3], []SsCode{ssServiceType}, ErrLCSAddPrivacyExceptionListNotAllowed},
		{"same SS-Code in both", fullLCSPrivacyList, []SsCode{ssCallSessionRelated}, ErrLCSPrivacyClassDuplicateSSCode},
		{"with four classes", fullLCSPrivacyList, []SsCode{ssServiceType}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := &InsertSubscriberDataArg{LcsInformation: &LCSInformation{
				AddLcsPrivacyExceptionList: lcsPrivacyClasses(tc.add...),
			}}
			if len(tc.base) > 0 {
				in.LcsInformation.LcsPrivacyExceptionList = lcsPrivacyClasses(tc.base...)
			}
			if tc.want != nil {
				if _, err := in.Marshal(); !errors.Is(err, tc.want) {
					t.Fatalf("Marshal: err = %v, want %v", err, tc.want)
				}
				return
			}
			semWantEqual(t, "ISD", in, semMarshalParseISD(t, in))
		})
	}
}

func TestParseAddLCSPrivacyExceptionList(t *testing.T) {
	for _, tc := range []struct {
		name      string
		base, add []SsCode
		kept      bool
		want      error
	}{
		{"without the list: discarded", nil, []SsCode{ssServiceType}, false, nil},
		{"with three classes: discarded", fullLCSPrivacyList[:3], []SsCode{ssServiceType}, false, nil},
		{"with three classes and a shared SS-Code: discarded first", fullLCSPrivacyList[:3], []SsCode{ssUniversal}, false, nil},
		{"same SS-Code in both: rejected", fullLCSPrivacyList, []SsCode{ssCallSessionRelated}, false, ErrLCSPrivacyClassDuplicateSSCode},
		{"with four classes: kept", fullLCSPrivacyList, []SsCode{ssServiceType}, true, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := semParseISDWire(t, &gsm_map.InsertSubscriberDataArg{LcsInformation: &gsm_map.LCSInformation{
				LcsPrivacyExceptionList:    wireLCSPrivacyClasses(tc.base...),
				AddLcsPrivacyExceptionList: wireLCSPrivacyClasses(tc.add...),
			}})
			if !errors.Is(err, tc.want) {
				t.Fatalf("Parse: err = %v, want %v", err, tc.want)
			}
			if err != nil {
				return
			}
			if kept := got.LcsInformation.AddLcsPrivacyExceptionList != nil; kept != tc.kept {
				t.Fatalf("AddLcsPrivacyExceptionList kept = %t, want %t", kept, tc.kept)
			}
			if _, err := got.Marshal(); err != nil {
				t.Errorf("Marshal of the parsed value: %v", err)
			}
		})
	}
}
