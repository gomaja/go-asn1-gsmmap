package gsmmap

import (
	"bytes"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	identity "github.com/gomaja/go-3gpp-identity"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
	"github.com/gomaja/go-sms/encoding/tpdu"
)

// TS 29.002 V19.1.0 §§17.7.6, 17.7.8: SRI-SM with two 7-octet
// addresses and FALSE sm-RP-PRI. SEQUENCE length = 9 + 3 + 9 = 21.
const toleranceSriValid = "30158007912143658709f18101008207912143658709f1"

// SRI-SM with a 10-octet msisdn holding 17 digits; length = 12 + 3 + 9 = 24.
const toleranceSriLong = "3018800a912143658709214365f78101008207912143658709f1"

// MO-ForwardSM: SC DA, msisdn OA, SMS-SUBMIT to 12 with one GSM7 'A'.
// The 9-octet TPDU has no validity period (TS 23.040 §9.2.2.2).
// SEQUENCE length = 4 + 4 + 11 = 19.
const toleranceMoValid = "301384029121820291210409010002912100000141"

// MT-ForwardSM: IMSI DA, SC OA, SMS-DELIVER from 12 with one GSM7 'A'
// and timestamp 2026-01-01 00:00:00 UTC (TS 23.040 §9.2.2.1).
// SEQUENCE length = 10 + 4 + 17 = 31; TPDU length = 15.
const toleranceMtValid = "301f800800010121436587f984029121040f000291210000621010000000000141"

func asTolerantParser[T any](parse func([]byte, ...ParseOption) (*T, error)) func([]byte, ...ParseOption) (marshaler, error) {
	return func(data []byte, opts ...ParseOption) (marshaler, error) {
		v, err := parse(data, opts...)
		if err != nil || v == nil {
			return nil, err
		}
		return any(v).(marshaler), nil
	}
}

var tolerantParsers = []struct {
	name  string
	parse func([]byte, ...ParseOption) (marshaler, error)
	wire  string
}{
	{"SriSm", asTolerantParser(ParseSriSm), toleranceSriValid},
	{"MoFsm", asTolerantParser(ParseMoFsm), toleranceMoValid},
	{"MtFsm", asTolerantParser(ParseMtFsm), toleranceMtValid},
}

func TestParseToleranceConstraints(t *testing.T) {
	for _, tc := range []struct {
		name       string
		parser     int
		wire       string
		path       string
		constraint string
		length     int
		check      func(*testing.T, marshaler)
	}{
		{"SriSm msisdn", 0, toleranceSriLong, "msisdn", "SIZE (1..9)", 10, func(t *testing.T, v marshaler) {
			if got := v.(*SriSm).MSISDN; got != "12345678901234567" {
				t.Fatalf("MSISDN = %q", got)
			}
		}},
		// MO: 10-octet msisdn OA; SEQUENCE length = 4 + 12 + 11 = 27.
		{"MoFsm msisdn", 1, "301b84029121820a912143658709214365f70409010002912100000141", "sm-RP-OA.msisdn", "SIZE (1..9)", 10, func(t *testing.T, v marshaler) {
			m := v.(*MoFsm)
			if m.SmRpOa.MSISDN != "12345678901234567" || m.TPDU.SmsType() != tpdu.SmsSubmit {
				t.Fatalf("MO = %+v", m)
			}
		}},
		// MT: 10-octet smsGmscAddress [3]; SEQUENCE length = 31 + 12 = 43.
		{"MtFsm GMSC", 2, "302b800800010121436587f984029121040f000291210000621010000000000141830a912143658709214365f7", "smsGmscAddress", "SIZE (1..9)", 10, func(t *testing.T, v marshaler) {
			if got := v.(*MtFsm).SmsGmscAddress; got != "12345678901234567" {
				t.Fatalf("SmsGmscAddress = %q", got)
			}
		}},
		// MT: INTEGER smDeliveryTimer 29; SEQUENCE length = 31 + 3 = 34.
		{"MtFsm timer below", 2, "3022800800010121436587f984029121040f00029121000062101000000000014102011d", "smDeliveryTimer", "(30..600)", 0, func(t *testing.T, v marshaler) {
			if got := v.(*MtFsm).SmDeliveryTimer; got == nil || *got != 29 {
				t.Fatalf("SmDeliveryTimer = %v", got)
			}
		}},
		// MT: INTEGER smDeliveryTimer 601; SEQUENCE length = 31 + 4 = 35.
		{"MtFsm timer above", 2, "3023800800010121436587f984029121040f00029121000062101000000000014102020259", "smDeliveryTimer", "(30..600)", 0, func(t *testing.T, v marshaler) {
			if got := v.(*MtFsm).SmDeliveryTimer; got == nil || *got != 601 {
				t.Fatalf("SmDeliveryTimer = %v", got)
			}
		}},
		// MO: unlisted non-extensible SM-DeliveryOutcome [1] 9; length = 19 + 3.
		{"MoFsm enum", 1, "301684029121820291210409010002912100000141810109", "sm-DeliveryOutcome", "ENUMERATED {0, 1, 2}", 0, func(t *testing.T, v marshaler) {
			if got := v.(*MoFsm).SmDeliveryOutcome; got == nil || *got != 9 {
				t.Fatalf("SmDeliveryOutcome = %v", got)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := tolerantParsers[tc.parser].parse
			data := mustHex(t, tc.wire)
			_, err := p(data)
			var validation *ber.ValidationError
			if !errors.As(err, &validation) || validation.Kind != ViolationKindASN1Constraint {
				t.Fatalf("strict error = %v, want ASN.1 ValidationError", err)
			}
			// A successful Parse appends after existing reports, including on reuse.
			violations := []Violation{{Path: "existing"}}
			option := WithTolerance(&violations)
			for i := 1; i <= 2; i++ {
				v, err := p(data, option)
				if err != nil {
					t.Fatal(err)
				}
				tc.check(t, v)
				if len(violations) != i+1 || violations[0].Path != "existing" {
					t.Fatalf("violations = %+v", violations)
				}
				got := violations[i]
				if got.Kind != ViolationKindASN1Constraint || got.Path != tc.path || got.Constraint != tc.constraint {
					t.Fatalf("violation = %+v", got)
				}
				if tc.length != 0 && (got.ObservedLength == nil || *got.ObservedLength != tc.length) {
					t.Fatalf("observed length = %v, want %d", got.ObservedLength, tc.length)
				}
				_, err = v.Marshal()
				if !errors.As(err, &validation) || validation.Kind != ViolationKindASN1Constraint {
					t.Fatalf("Marshal error = %v, want ASN.1 ValidationError", err)
				}
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.1.1 requires short lengths below 128. Only
// msisdn's length uses the long form here: SEQUENCE length = 10 + 3 + 9 = 22.
func TestParseToleranceProtocolRule(t *testing.T) {
	data := mustHex(t, "3016808107912143658709f18101008207912143658709f1")
	_, err := ParseSriSm(data)
	var validation *ber.ValidationError
	if !errors.As(err, &validation) || validation.Kind != ViolationKindProtocolRule {
		t.Fatalf("strict error = %v, want ProtocolRule ValidationError", err)
	}
	var violations []Violation
	v, err := ParseSriSm(data, WithTolerance(&violations))
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 || violations[0].Kind != ViolationKindProtocolRule || violations[0].Rule.Clause != "17.1.1" {
		t.Fatalf("violations = %+v", violations)
	}
	canonical, err := v.Marshal()
	if err != nil || !bytes.Equal(canonical, mustHex(t, toleranceSriValid)) {
		t.Fatalf("Marshal = %x, %v", canonical, err)
	}
}

func TestParseToleranceValid(t *testing.T) {
	for _, p := range tolerantParsers {
		t.Run(p.name, func(t *testing.T) {
			data := mustHex(t, p.wire)
			strict, err := p.parse(data)
			if err != nil {
				t.Fatal(err)
			}
			for _, initial := range [][]Violation{nil, {{Path: "existing"}}} {
				violations := append([]Violation(nil), initial...)
				tolerant, err := p.parse(data, WithTolerance(&violations))
				if err != nil || !reflect.DeepEqual(strict, tolerant) || !reflect.DeepEqual(initial, violations) {
					t.Fatalf("tolerant = %+v, %v; violations = %+v", tolerant, err, violations)
				}
			}
		})
	}
}

func TestParseToleranceFailureDoesNotAppend(t *testing.T) {
	for _, tc := range []struct {
		name   string
		parser int
		wire   string
		want   error
	}{
		// SRI: zero-octet msisdn; SEQUENCE length = 2 + 3 + 9 = 14.
		{"SriSm address", 0, "300e80008101008207912143658709f1", ErrSriSmMissingMSISDN},
		// SRI: optional IMSI of two octets/four digits; length = 21 + 4 = 25.
		{"SriSm identity", 0, "30198007912143658709f18101008207912143658709f18c020001", identity.ErrIMSILength},
		// SRI: CorrelationID [15] with two-octet HLR-Id and nonempty SIP-URI-B;
		// length = 21 + 9 = 30; CorrelationID length = 4 + 3 = 7.
		{"SriSm HLR identity", 0, "301e8007912143658709f18101008207912143658709f1af0780020001820161", identity.ErrIMSILength},
		// MO: zero-octet OA msisdn; SEQUENCE length = 4 + 2 + 11 = 17.
		{"MoFsm address", 1, "30118402912182000409010002912100000141", ErrSmRpOaMSISDNDecodedEmpty},
		// MT: zero-octet GMSC address; SEQUENCE length = 31 + 2 = 33.
		{"MtFsm address", 2, "3021800800010121436587f984029121040f0002912100006210100000000001418300", ErrMtFsmSmsGmscAddressDecodedEmpty},
		// MO: over-length OA and empty SM-RP-UI; length = 4 + 12 + 2 = 18.
		{"MoFsm TPDU", 1, "301284029121820a912143658709214365f70400", nil},
		// MT: empty SM-RP-UI; SEQUENCE length = 10 + 4 + 2 = 16.
		{"MtFsm TPDU", 2, "3010800800010121436587f9840291210400", nil},
		// SRI: over-length msisdn then invalid BOOLEAN contents; length = 24.
		{"codec failure", 0, "3018800a912143658709214365f78102000082069121436587f9", ber.ErrInvalidValue},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, initial := range [][]Violation{nil, {{Path: "existing"}}} {
				violations := append([]Violation(nil), initial...)
				v, err := tolerantParsers[tc.parser].parse(mustHex(t, tc.wire), WithTolerance(&violations))
				if err == nil || v != nil || (tc.want != nil && !errors.Is(err, tc.want)) {
					t.Fatalf("Parse = %+v, %v, want %v", v, err, tc.want)
				}
				if tc.want == nil && !strings.Contains(err.Error(), "unmarshaling TPDU") {
					t.Fatalf("error = %v, want TPDU error", err)
				}
				if !reflect.DeepEqual(initial, violations) {
					t.Fatalf("failed Parse appended: %+v", violations)
				}
			}
		})
	}
}

func TestParseOptionInvalid(t *testing.T) {
	for _, p := range tolerantParsers {
		t.Run(p.name, func(t *testing.T) {
			var violations []Violation
			for _, opts := range [][]ParseOption{
				{nil},
				{WithTolerance(nil)},
				{WithTolerance(&violations), WithTolerance(&violations)},
				{WithTolerance(&violations), WithTolerance(new([]Violation))},
				{WithTolerance(&violations), nil},
			} {
				// Invalid options take precedence even over an undecodable input.
				for _, data := range [][]byte{nil, mustHex(t, p.wire)} {
					v, err := p.parse(data, opts...)
					if !errors.Is(err, ErrParseOptionInvalid) || v != nil || len(violations) != 0 {
						t.Fatalf("Parse = %+v, %v; violations = %+v", v, err, violations)
					}
				}
			}
		})
	}
}

func TestParseToleranceEmptyOptionalOctets(t *testing.T) {
	// SRI: present empty sm-RP-SMEA [9]; length = 21 + 2 = 23.
	data := mustHex(t, "30178007912143658709f18101008207912143658709f18900")
	var violations []Violation
	v, err := ParseSriSm(data, WithTolerance(&violations))
	if err != nil {
		t.Fatal(err)
	}
	if len(v.SmRpSmea) != 0 || len(violations) != 1 || violations[0].Path != "sm-RP-SMEA" {
		t.Fatalf("SMEA = %x; violations = %+v", v.SmRpSmea, violations)
	}
	canonical, err := v.Marshal()
	if err != nil || !bytes.Equal(canonical, mustHex(t, toleranceSriValid)) {
		t.Fatalf("Marshal = %x, %v", canonical, err)
	}
}

func TestParseToleranceMultipleViolations(t *testing.T) {
	// MT: timer 29 and a 10-octet GMSC address. The two independent violations
	// follow the mandatory fields; SEQUENCE length = 31 + 3 + 12 = 46.
	data := mustHex(t, "302e800800010121436587f984029121040f00029121000062101000000000014102011d830a912143658709214365f7")
	violations := []Violation{{Path: "existing"}}
	v, err := ParseMtFsm(data, WithTolerance(&violations))
	if err != nil {
		t.Fatal(err)
	}
	if v.SmDeliveryTimer == nil || *v.SmDeliveryTimer != 29 || v.SmsGmscAddress != "12345678901234567" {
		t.Fatalf("MT = %+v", v)
	}
	if len(violations) != 3 || violations[0].Path != "existing" {
		t.Fatalf("violations = %+v", violations)
	}
	for i, want := range []struct{ path, constraint string }{
		{"smDeliveryTimer", "(30..600)"},
		{"smsGmscAddress", "SIZE (1..9)"},
	} {
		got := violations[i+1]
		if got.Kind != ViolationKindASN1Constraint || got.Path != want.path || got.Constraint != want.constraint {
			t.Fatalf("violation %d = %+v", i, got)
		}
	}
}

func TestParseToleranceEmptyLMSI(t *testing.T) {
	// MO: present zero-octet LMSI DA [1], msisdn OA and SMS-SUBMIT.
	// SEQUENCE length = 2 + 4 + 11 = 17. LMSI has SIZE (4)
	// (TS 29.002 V19.1.0 §17.7.8); an empty public slice selects no DA.
	data := mustHex(t, "30118100820291210409010002912100000141")
	var violations []Violation
	v, err := ParseMoFsm(data, WithTolerance(&violations))
	if err != nil {
		t.Fatal(err)
	}
	if len(v.SmRpDa.LMSI) != 0 || len(violations) != 1 || violations[0].Path != "sm-RP-DA.lmsi" {
		t.Fatalf("DA = %+v; violations = %+v", v.SmRpDa, violations)
	}
	if _, err := v.Marshal(); !errors.Is(err, ErrSmRpDaNoAlternative) {
		t.Fatalf("Marshal = %v, want ErrSmRpDaNoAlternative", err)
	}
}

func TestParseToleranceTimerRepresentation(t *testing.T) {
	// MT: INTEGER smDeliveryTimer 2^31, requiring five octets including sign;
	// SEQUENCE length = 31 + 7 = 38. Public int cannot hold it on 32-bit Go.
	data := mustHex(t, "3026800800010121436587f984029121040f00029121000062101000000000014102050080000000")
	var violations []Violation
	v, err := ParseMtFsm(data, WithTolerance(&violations))
	if strconv.IntSize == 32 {
		if !errors.Is(err, ErrIntOverflow) || v != nil || len(violations) != 0 {
			t.Fatalf("unrepresentable timer: %+v, %v; violations = %+v", v, err, violations)
		}
		return
	}
	if err != nil || v.SmDeliveryTimer == nil || int64(*v.SmDeliveryTimer) != 1<<31 || len(violations) != 1 {
		t.Fatalf("timer: %+v, %v; violations = %+v", v, err, violations)
	}
}

// FuzzParseTolerant compares strict and tolerant decoding without requiring
// tolerated values to satisfy the encoder's strict constraints.
func FuzzParseTolerant(f *testing.F) {
	for _, seed := range berSeeds(f) {
		for i := range tolerantParsers {
			f.Add(uint8(i), seed)
		}
	}
	f.Fuzz(func(t *testing.T, which uint8, data []byte) {
		p := tolerantParsers[int(which)%len(tolerantParsers)]
		strict, strictErr := p.parse(data)
		violations := []Violation{{Path: "existing"}}
		tolerant, tolerantErr := p.parse(data, WithTolerance(&violations))
		if len(violations) == 0 || violations[0].Path != "existing" {
			t.Fatalf("%s changed existing reports", p.name)
		}
		if tolerantErr != nil {
			if tolerant != nil || len(violations) != 1 {
				t.Fatalf("%s failed with a value or reports: %v, %+v", p.name, tolerantErr, violations)
			}
		} else {
			// A tolerated value may be refused by strict Marshal, but may not panic.
			_, _ = tolerant.Marshal()
		}
		if strictErr == nil || (tolerantErr == nil && len(violations) == 1) {
			if strictErr != nil || tolerantErr != nil || len(violations) != 1 || !reflect.DeepEqual(strict, tolerant) {
				t.Fatalf("%s: strict = %+v, %v; tolerant = %+v, %v; reports = %+v", p.name, strict, strictErr, tolerant, tolerantErr, violations)
			}
		}
	})
}
