package gsmmap

import "testing"

func TestReadyForSMMinimal(t *testing.T) {
	// SEQUENCE: [0] IMSI and AlertReason ms-Present(0).
	hlrRoundTrip(t, &ReadyForSMArg{IMSI: "001010123456789", AlertReason: AlertReasonMsPresent}, ParseReadyForSMArg, "300d800800010121436587f90a0100")
	// Empty result parameter.
	hlrRoundTrip(t, &ReadyForSMRes{}, ParseReadyForSMRes, "3000")
}

func TestReadyForSMFull(t *testing.T) {
	// IMSI; memoryAvailable; both independent NULL indicators; four Time octets.
	hlrRoundTrip(t, &ReadyForSMArg{IMSI: "001010123456789", AlertReason: AlertReasonMemoryAvailable, AlertReasonIndicator: true, AdditionalAlertReasonIndicator: true, MaximumUEAvailabilityTime: HexBytes{1, 2, 3, 4}}, ParseReadyForSMArg, "3017800800010121436587f90a010105008100040401020304")
}

func TestReadyForSMConstraints(t *testing.T) {
	_, err := (&ReadyForSMArg{IMSI: "001010", AlertReason: AlertReason(2)}).Marshal()
	wantValidationError(t, err, "alertReason", "ENUMERATED {0, 1}")
	_, err = ParseReadyForSMArg(mustHex(t, "300d800800010121436587f90a0102"))
	wantValidationError(t, err, "alertReason", "ENUMERATED {0, 1}")
	for _, n := range []int{0, 3, 5} {
		v := &ReadyForSMArg{IMSI: "001010", MaximumUEAvailabilityTime: make(HexBytes, n)}
		_, err := v.Marshal()
		if n == 0 { // an empty optional time is absent
			if err != nil {
				t.Fatal(err)
			}
		} else {
			wantValidationError(t, err, "maximumUeAvailabilityTime", "SIZE (4)")
		}
		body := append(mustHex(t, "80030001010a0100"), ussdTLV(0x04, make([]byte, n))...)
		_, err = ParseReadyForSMArg(ussdTLV(0x30, body))
		wantValidationError(t, err, "maximumUeAvailabilityTime", "SIZE (4)")
	}
}

// The two indicators are separate components: alertReasonIndicator is an
// untagged NULL and additionalAlertReasonIndicator is [1] (3GPP TS 29.002
// V19.1.0 §17.7.6).
func TestReadyForSMIndicators(t *testing.T) {
	hlrRoundTrip(t, &ReadyForSMArg{IMSI: "001010123456789", AlertReasonIndicator: true}, ParseReadyForSMArg, "300f800800010121436587f90a01000500")
	hlrRoundTrip(t, &ReadyForSMArg{IMSI: "001010123456789", AdditionalAlertReasonIndicator: true}, ParseReadyForSMArg, "300f800800010121436587f90a01008100")
}
