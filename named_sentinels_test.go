package gsmmap

import (
	"errors"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func TestInterimRejectionsExposeSentinels(t *testing.T) {
	forwardingReason := ForwardingReason(3)
	mnpStatus := NumberPortabilityStatus(3)
	callBarringCause := gsm_map.CallBarringCause(2)
	tests := []struct {
		name  string
		check func() error
		want  error
	}{
		{"forwarding reason", func() error {
			return validateSri(&Sri{MSISDN: "123456", GmscOrGsmSCFAddress: "123456", InterrogationType: InterrogationBasicCall, ForwardingReason: &forwardingReason})
		}, ErrSriForwardingReasonInvalid},
		{"number portability status", func() error {
			_, err := convertMnpInfoResToWire(&MnpInfoRes{NumberPortabilityStatus: &mnpStatus})
			return err
		}, ErrNumberPortabilityStatusInvalid},
		{"IPv4 element length", func() error {
			_, err := convertPdnGwIdentityToWire(&PdnGwIdentity{IPv4Address: HexBytes{1, 2, 3}})
			return err
		}, ErrPdnGwIdentityIPv4AddressInvalidLength},
		{"call barring cause", func() error {
			_, err := convertWireToCallBarredParam(&gsm_map.CallBarredParam{Choice: gsm_map.CallBarredParamChoiceCallBarringCause, CallBarringCause: &callBarringCause})
			return err
		}, ErrCallBarringCauseInvalid},
		{"LCS address metadata without digits", func() error {
			_, err := convertLCSClientIDToWire(&LCSClientID{LcsClientDialedByMSNature: 1})
			return err
		}, ErrLCSClientIDDialedByMSNaturePlanWithoutDigits},
		{"present digitless LCS address", func() error {
			addr := gsm_map.ISDNAddressString{0x91}
			_, err := convertWireToLCSClientID(&gsm_map.LCSClientID{LcsClientDialedByMS: &addr})
			return err
		}, ErrLCSClientIDDialedByMSDecodedEmpty},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.check(); !errors.Is(err, tt.want) {
				t.Fatalf("error %v does not wrap %v", err, tt.want)
			}
		})
	}
}
