package gsmmap

import (
	"errors"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// TS 29.002 V19.1.0 §17.7.13: ServingNodeAddress is a CHOICE; accuracy value 1 is defined.
func TestPSLResponseChoices(t *testing.T) {
	v := AccuracyFulfilmentIndicator(1)
	for _, node := range []*ServingNodeAddress{{MscNumber: "12345"}, {SgsnNumber: "67890"}, {MmeNumber: HexBytes("mme.example")}} {
		in := &ProvideSubscriberLocationRes{LocationEstimate: HexBytes{0, 0, 0, 0, 0, 0, 0}, TargetServingNodeForHandover: node, AccuracyFulfilmentIndicator: &v}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseProvideSubscriberLocationRes(data)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		wantEqual(t, "node", node, got.TargetServingNodeForHandover)
		wantEqual(t, "accuracy", v, *got.AccuracyFulfilmentIndicator)
	}
	if _, err := (&ProvideSubscriberLocationRes{TargetServingNodeForHandover: &ServingNodeAddress{MscNumber: "12345", MmeNumber: HexBytes("mme.example")}}).Marshal(); !errors.Is(err, ErrServingNodeAddressMultipleAlternatives) {
		t.Fatalf("multiple node error: got %v, want %v", err, ErrServingNodeAddressMultipleAlternatives)
	}
}

// TS 29.002 V19.1.0 §17.7.13: LCSFormatIndicator has defined values 0..4.
func TestPSLFormats(t *testing.T) {
	for _, v := range []LCSFormatIndicator{0, 4} {
		in := &ProvideSubscriberLocationArg{MlcNumber: "12345", IMSI: "001010123456789", LcsClientID: &LCSClientID{LcsClientName: &LCSClientName{DataCodingScheme: 0x0f, NameString: HexBytes{1}, LcsFormatIndicator: &v}, LcsRequestorID: &LCSRequestorID{DataCodingScheme: 0x0f, RequestorIDString: HexBytes{2}, LcsFormatIndicator: &v}}}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal %d: %v", v, err)
		}
		got, err := ParseProvideSubscriberLocation(data)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		wantEqual(t, "name format", v, *got.LcsClientID.LcsClientName.LcsFormatIndicator)
		wantEqual(t, "requestor format", v, *got.LcsClientID.LcsRequestorID.LcsFormatIndicator)
	}
}

// TS 29.002 V19.1.0 §17.7.13: beingInsideArea is named bit 3.
func TestPSLInsideArea(t *testing.T) {
	w := &gsm_map.ProvideSubscriberLocationArg{MlcNumber: gsm_map.ISDNAddressString{0x80, 0x21}, LocationType: gsm_map.LocationType{LocationEstimateType: LocationEstimateActivateDeferredLocation}}
	in := &ProvideSubscriberLocationArg{MlcNumber: "12", IMSI: "001010123456789", LcsReferenceNumber: LCSReferenceNumber{1}, LocationType: LocationType{LocationEstimateType: LocationEstimateActivateDeferredLocation, DeferredLocationEventType: &DeferredLocationEventType{BeingInsideArea: true}}}
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err = w.UnmarshalBER(data); err != nil {
		t.Fatalf("wire decode: %v", err)
	}
	wantEqual(t, "bit length", 4, w.LocationType.DeferredLocationEventType.BitLength)
	wantEqual(t, "bytes", []byte{0x10}, w.LocationType.DeferredLocationEventType.Bytes)
}
