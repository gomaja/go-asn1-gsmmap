package gsmmap

import (
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/gomaja/go-asn1/runtime"
)

// 3GPP TS 29.002 V19.1.0 §17.7.13 DeferredLocationEventType: "a
// ProvideSubscriberLocation-Arg containing other values than listed above in
// DeferredLocationEventType shall be rejected by the receiver with a return
// error cause of unexpected data value". The listed values are msAvailable(0)
// to periodicLDR(4).

// The probe wire: a ProvideSubscriberLocation-Arg whose
// deferredLocationEventType sets bits 0 and 5.
func TestParsePSLRejectsUnlistedDeferredLocationEventBitWire(t *testing.T) {
	data, err := hex.DecodeString("30123007800100810202840407911316325476f8")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := ParseProvideSubscriberLocation(data); !errors.Is(err, ErrDeferredLocationEventTypeUnrecognized) || got != nil {
		t.Errorf("got %v, %v; want ErrDeferredLocationEventTypeUnrecognized", got, err)
	}
}

func TestParsePSLDeferredLocationEventType(t *testing.T) {
	for _, tc := range []struct {
		name string
		bs   runtime.BitString
		want *DeferredLocationEventType
	}{
		{"msAvailable", runtime.BitString{Bytes: []byte{0x80}, BitLength: 1}, &DeferredLocationEventType{MsAvailable: true}},
		{"all listed", runtime.BitString{Bytes: []byte{0xf8}, BitLength: 5}, &DeferredLocationEventType{
			MsAvailable: true, EnteringIntoArea: true, LeavingFromArea: true, BeingInsideArea: true, PeriodicLDR: true,
		}},
		// Bits past periodicLDR that are not set carry no value.
		{"trailing zero bits", runtime.BitString{Bytes: []byte{0x08, 0x00}, BitLength: 16}, &DeferredLocationEventType{PeriodicLDR: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := pslWire(t, pslArg())
			bs := tc.bs
			w.LocationType.DeferredLocationEventType = &bs
			got, err := ParseProvideSubscriberLocation(strictBER(t, w))
			if err != nil {
				t.Fatalf("ParseProvideSubscriberLocation: %v", err)
			}
			wantEqual(t, "DeferredLocationEventType", tc.want, got.LocationType.DeferredLocationEventType)
		})
	}
	for bit := 5; bit < 16; bit++ {
		t.Run(fmt.Sprintf("bit %d", bit), func(t *testing.T) {
			w := pslWire(t, pslArg())
			b := []byte{0x80, 0x00}
			b[bit/8] |= 0x80 >> (bit % 8)
			w.LocationType.DeferredLocationEventType = &runtime.BitString{Bytes: b, BitLength: 16}
			if got, err := ParseProvideSubscriberLocation(strictBER(t, w)); !errors.Is(err, ErrDeferredLocationEventTypeUnrecognized) || got != nil {
				t.Errorf("got %v, %v; want ErrDeferredLocationEventTypeUnrecognized", got, err)
			}
		})
	}
}

// The clause names the ProvideSubscriberLocation-Arg only. The
// deferredmt-lrData of a SubscriberLocationReport-Arg keeps the listed bits
// and drops the others.
func TestParseSLRDeferredLocationEventTypeIgnoresUnlistedBits(t *testing.T) {
	a := minimalSLRArg()
	a.DeferredmtLrData = &DeferredmtLrData{DeferredLocationEventType: DeferredLocationEventType{EnteringIntoArea: true}}
	w, err := convertSubscriberLocationReportArgToWire(a)
	if err != nil {
		t.Fatalf("convertSubscriberLocationReportArgToWire: %v", err)
	}
	w.DeferredmtLrData.DeferredLocationEventType = runtime.BitString{Bytes: []byte{0x44}, BitLength: 6}
	got, err := ParseSubscriberLocationReport(strictBER(t, w))
	if err != nil {
		t.Fatalf("ParseSubscriberLocationReport: %v", err)
	}
	wantEqual(t, "DeferredLocationEventType", DeferredLocationEventType{EnteringIntoArea: true}, got.DeferredmtLrData.DeferredLocationEventType)
}
