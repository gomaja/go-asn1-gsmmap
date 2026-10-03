// semantic_receiver_test.go
//
// Receiver rules of 3GPP TS 29.002 V19.1.0 for ENUMERATEDs whose unknown
// values a receiver discards or ignores. Parse applies the rule; Marshal,
// where the type is sent, sends only the listed values.
package gsmmap

import (
	"bytes"
	"errors"
	"math"
	"testing"

	"github.com/gomaja/go-asn1/runtime/ber"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// §17.7.1 NetworkAccessMode: "if unknown values are received in
// NetworkAccessMode they shall be discarded."
func TestNetworkAccessModeDiscardsUnknown(t *testing.T) {
	for _, v := range []NetworkAccessMode{0, 1, 2} {
		mode := v
		in := &InsertSubscriberDataArg{IMSI: "204080012345678", NetworkAccessMode: &mode}
		got := semMarshalParseISD(t, in)
		semWantEqual(t, "ISD", in, got)
	}
	for _, v := range []NetworkAccessMode{3, -1, math.MaxInt64} {
		mode := v
		if _, err := (&InsertSubscriberDataArg{NetworkAccessMode: &mode}).Marshal(); !errors.Is(err, ErrNetworkAccessModeInvalid) {
			t.Errorf("Marshal %d: err = %v, want ErrNetworkAccessModeInvalid", v, err)
		}
		imsi := gsm_map.IMSI{0x02, 0x04, 0x08, 0x10, 0x32, 0x54, 0x76, 0xF8}
		got, err := semParseISDWire(t, &gsm_map.InsertSubscriberDataArg{Imsi: &imsi, NetworkAccessMode: &mode})
		if err != nil {
			t.Errorf("Parse %d: %v", v, err)
			continue
		}
		if got.NetworkAccessMode != nil || got.IMSI != "204080012345678" {
			t.Errorf("Parse %d: NetworkAccessMode = %v, IMSI = %q; want discarded mode, IMSI kept", v, got.NetworkAccessMode, got.IMSI)
		}
	}
}

// §17.7.8 Ext-ProtocolId: "For Ext-ExternalSignalInfo sequences containing
// this parameter with any other value than the ones listed the receiver
// shall ignore the whole Ext-ExternalSignalInfo sequence."
func TestExtExternalSignalInfoIgnoredForUnknownProtocol(t *testing.T) {
	base := func() *Sri {
		return &Sri{MSISDN: "31612345678", GmscOrGsmSCFAddress: "31600000001"}
	}
	in := base()
	in.AdditionalSignalInfo = &ExtExternalSignalInfo{ExtProtocolID: 1, SignalInfo: HexBytes{0x01, 0x02}}
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseSri(data)
	if err != nil {
		t.Fatalf("ParseSri: %v", err)
	}
	semWantEqual(t, "AdditionalSignalInfo", in.AdditionalSignalInfo, got.AdditionalSignalInfo)

	for _, v := range []int64{0, 2, -1, math.MaxInt64} {
		bad := base()
		bad.AdditionalSignalInfo = &ExtExternalSignalInfo{ExtProtocolID: int(v), SignalInfo: HexBytes{0x01}}
		if _, err := bad.Marshal(); !errors.Is(err, ErrExtProtocolIDInvalid) {
			t.Errorf("Marshal %d: err = %v, want ErrExtProtocolIDInvalid", v, err)
		}
		w, err := convertSriToArg(base())
		if err != nil {
			t.Fatalf("convertSriToArg: %v", err)
		}
		w.AdditionalSignalInfo = &gsm_map.ExtExternalSignalInfo{ExtProtocolId: gsm_map.ExtProtocolId(v), SignalInfo: gsm_map.SignalInfo{0x01}}
		data, err := w.MarshalBER()
		if err != nil {
			t.Fatalf("MarshalBER: %v", err)
		}
		got, err := ParseSri(data)
		if err != nil {
			t.Errorf("ParseSri %d: %v", v, err)
			continue
		}
		if got.AdditionalSignalInfo != nil || got.MSISDN != "31612345678" {
			t.Errorf("ParseSri %d: AdditionalSignalInfo = %+v, MSISDN = %q; want ignored, MSISDN kept", v, got.AdditionalSignalInfo, got.MSISDN)
		}
	}
}

func semParseError(t *testing.T, code MapErrorCode, w interface {
	MarshalBER(...ber.EncodeOption) ([]byte, error)
}) any {
	t.Helper()
	data, err := w.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	v, err := ParseReturnErrorParameter(code, data)
	if err != nil {
		t.Fatalf("ParseReturnErrorParameter: %v", err)
	}
	return v
}

// §17.7.7 UnknownSubscriberDiagnostic: "if unknown values are received in
// UnknownSubscriberDiagnostic they shall be discarded".
func TestUnknownSubscriberDiagnosticDiscardsUnknown(t *testing.T) {
	for _, tc := range []struct {
		wire gsm_map.UnknownSubscriberDiagnostic
		kept bool
	}{{0, true}, {1, true}, {2, true}, {3, false}, {-1, false}, {math.MaxInt64, false}} {
		d := tc.wire
		got := semParseError(t, MapErrorUnknownSubscriber, &gsm_map.UnknownSubscriberParam{UnknownSubscriberDiagnostic: &d}).(*UnknownSubscriberParam)
		if kept := got.UnknownSubscriberDiagnostic != nil; kept != tc.kept || kept && *got.UnknownSubscriberDiagnostic != tc.wire {
			t.Errorf("%d: UnknownSubscriberDiagnostic = %v, want kept %t", tc.wire, got.UnknownSubscriberDiagnostic, tc.kept)
		}
	}
}

// §17.7.8 AdditionalNetworkResource: "if unknown value is received in
// AdditionalNetworkResource it shall be ignored." §17.7.7
// FailureCauseParam: "if unknown value is received in FailureCauseParam it
// shall be ignored".
func TestExtensibleSystemFailureParamIgnoresUnknown(t *testing.T) {
	for _, tc := range []struct {
		resource gsm_map.AdditionalNetworkResource
		cause    gsm_map.FailureCauseParam
		kept     bool
	}{{0, 0, true}, {7, 0, true}, {8, 1, false}, {-1, -1, false}, {math.MaxInt64, math.MaxInt64, false}} {
		r, c := tc.resource, tc.cause
		w := gsm_map.NewSystemFailureParamExtensibleSystemFailureParam(gsm_map.ExtensibleSystemFailureParam{AdditionalNetworkResource: &r, FailureCauseParam: &c})
		got := semParseError(t, MapErrorSystemFailure, &w).(*SystemFailureParam).ExtensibleSystemFailureParam
		if kept := got.AdditionalNetworkResource != nil; kept != tc.kept || kept && *got.AdditionalNetworkResource != tc.resource {
			t.Errorf("%d: AdditionalNetworkResource = %v, want kept %t", tc.resource, got.AdditionalNetworkResource, tc.kept)
		}
		if kept := got.FailureCauseParam != nil; kept != tc.kept || kept && *got.FailureCauseParam != tc.cause {
			t.Errorf("%d: FailureCauseParam = %v, want kept %t", tc.cause, got.FailureCauseParam, tc.kept)
		}
	}
}

// §17.7.7 AbsentSubscriberReason: "at reception of other values than the
// ones listed the AbsentSubscriberReason shall be ignored."
func TestAbsentSubscriberReasonIgnoresUnknown(t *testing.T) {
	for _, tc := range []struct {
		wire gsm_map.AbsentSubscriberReason
		kept bool
	}{{0, true}, {5, true}, {6, false}, {-1, false}, {math.MaxInt64, false}} {
		r := tc.wire
		got := semParseError(t, MapErrorAbsentSubscriber, &gsm_map.AbsentSubscriberParam{AbsentSubscriberReason: &r}).(*AbsentSubscriberParam)
		if kept := got.AbsentSubscriberReason != nil; kept != tc.kept || kept && *got.AbsentSubscriberReason != tc.wire {
			t.Errorf("%d: AbsentSubscriberReason = %v, want kept %t", tc.wire, got.AbsentSubscriberReason, tc.kept)
		}
	}
}

// §17.7.7: "if the additionalRoamingNotallowedCause is received by the
// MSC/VLR or SGSN then the roamingNotAllowedCause shall be discarded."
// Parse leaves the discarding to the MSC/VLR or SGSN: with the additional
// cause present it passes RoamingNotAllowedCause through unchecked, exactly
// as received, so the parsed fields encode back to the received octets.
func TestRoamingNotAllowedCausePassedThroughWithAdditionalCause(t *testing.T) {
	additional := gsm_map.AdditionalRoamingNotAllowedCauseSupportedRATTypesNotAllowed
	for _, cause := range []gsm_map.RoamingNotAllowedCause{1, 2, -1, math.MaxInt64} {
		data, err := (&gsm_map.RoamingNotAllowedParam{
			RoamingNotAllowedCause:           cause,
			AdditionalRoamingNotAllowedCause: &additional,
		}).MarshalBER()
		if err != nil {
			t.Fatalf("MarshalBER: %v", err)
		}
		v, err := ParseReturnErrorParameter(MapErrorRoamingNotAllowed, data)
		if err != nil {
			t.Fatalf("cause %d: ParseReturnErrorParameter: %v", cause, err)
		}
		got := v.(*RoamingNotAllowedParam)
		if got.RoamingNotAllowedCause != cause {
			t.Errorf("cause %d: RoamingNotAllowedCause = %d, want the wire value", cause, got.RoamingNotAllowedCause)
		}
		if got.AdditionalRoamingNotAllowedCause == nil || *got.AdditionalRoamingNotAllowedCause != additional {
			t.Errorf("cause %d: AdditionalRoamingNotAllowedCause = %v", cause, got.AdditionalRoamingNotAllowedCause)
		}
		// The package has no encoder for error parameters, so the parsed
		// fields go back through the wire type.
		again, err := (&gsm_map.RoamingNotAllowedParam{
			RoamingNotAllowedCause:           got.RoamingNotAllowedCause,
			AdditionalRoamingNotAllowedCause: got.AdditionalRoamingNotAllowedCause,
		}).MarshalBER()
		if err != nil {
			t.Fatalf("cause %d: MarshalBER of the parsed value: %v", cause, err)
		}
		if !bytes.Equal(again, data) {
			t.Errorf("cause %d: parsed value encodes to %x, want %x", cause, again, data)
		}
		// Without the additional cause the non-extensible cause is checked.
		data, err = (&gsm_map.RoamingNotAllowedParam{RoamingNotAllowedCause: cause}).MarshalBER()
		if err != nil {
			t.Fatalf("MarshalBER: %v", err)
		}
		if _, err := ParseReturnErrorParameter(MapErrorRoamingNotAllowed, data); err == nil {
			t.Errorf("cause %d without the additional cause: Parse accepted it", cause)
		}
	}
}
