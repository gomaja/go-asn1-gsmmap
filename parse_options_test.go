package gsmmap

import (
	"bytes"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	identity "github.com/gomaja/go-3gpp-identity"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
	"github.com/gomaja/go-sms/encoding/tpdu"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

// directTolerant decodes with go-asn1's tolerance and converts without the
// option code, as an oracle for Parse with WithTolerance.
func directTolerant[W, T any](decode func(*W, []byte, ...ber.DecodeOption) error, convert func(*W) (*T, error)) func([]byte) (marshaler, []Violation, error) {
	return func(data []byte) (marshaler, []Violation, error) {
		var log ber.ViolationLog
		var w W
		if err := decode(&w, data, ber.WithTolerance(&log)); err != nil {
			return nil, nil, err
		}
		v, err := convert(&w)
		if err != nil {
			return nil, nil, err
		}
		return any(v).(marshaler), log.Snapshot(), nil
	}
}

var tolerantParsers = []struct {
	name   string
	parse  func([]byte, ...ParseOption) (marshaler, error)
	direct func([]byte) (marshaler, []Violation, error)
	wire   string
}{
	{"SriSm", asTolerantParser(ParseSriSm), directTolerant((*gsm_map.RoutingInfoForSMArg).UnmarshalBER, convertArgToSriSm), toleranceSriValid},
	{"MoFsm", asTolerantParser(ParseMoFsm), directTolerant((*gsm_map.MOForwardSMArg).UnmarshalBER, convertArgToMoFsm), toleranceMoValid},
	{"MtFsm", asTolerantParser(ParseMtFsm), directTolerant((*gsm_map.MTForwardSMArg).UnmarshalBER, convertArgToMtFsm), toleranceMtValid},
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
		// SRI: 11-octet msisdn with 20 digits, sm-RP-PRI and a 7-octet
		// service-centre address; SEQUENCE length = 13 + 3 + 9 = 25.
		{"SriSm msisdn 20 digits", 0, "3019800b91214365870921436587098101008207912143658709f1", "msisdn", "SIZE (1..9)", 11, func(t *testing.T, v marshaler) {
			if got := v.(*SriSm).MSISDN; got != "12345678901234567890" {
				t.Fatalf("MSISDN = %q", got)
			}
		}},
		// SRI: 7-octet msisdn, sm-RP-PRI and a 21-octet service-centre address
		// with 40 digits, over AddressString SIZE (1..20); length = 9 + 3 + 23 = 35.
		{"SriSm service centre", 0, "30238007912143658709f18101008215912143658709214365870921436587092143658709", "serviceCentreAddress", "SIZE (1..20)", 21, func(t *testing.T, v marshaler) {
			if got := v.(*SriSm).ServiceCentreAddress; got != strings.Repeat("1234567890", 4) {
				t.Fatalf("ServiceCentreAddress = %q", got)
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
	for _, tc := range []struct {
		name, wire, sent, path string
		parser                 int
		field                  func(marshaler) HexBytes
	}{
		// SRI: present empty sm-RP-SMEA [9]; length = 21 + 2 = 23.
		{"SriSm sm-RP-SMEA", "30178007912143658709f18101008207912143658709f18900", toleranceSriValid, "sm-RP-SMEA", 0,
			func(v marshaler) HexBytes { return v.(*SriSm).SmRpSmea }},
		// MT: empty smDeliveryStartTime after the TPDU; length = 31 + 2 = 33.
		{"MtFsm smDeliveryStartTime", "3021800800010121436587f984029121040f0002912100006210100000000001410400", toleranceMtValid, "smDeliveryStartTime", 2,
			func(v marshaler) HexBytes { return v.(*MtFsm).SmDeliveryStartTime }},
		// MT: empty maximumRetransmissionTime [2]; length = 31 + 2 = 33.
		{"MtFsm maximumRetransmissionTime", "3021800800010121436587f984029121040f0002912100006210100000000001418200", toleranceMtValid, "maximumRetransmissionTime", 2,
			func(v marshaler) HexBytes { return v.(*MtFsm).MaximumRetransmissionTime }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var violations []Violation
			v, err := tolerantParsers[tc.parser].parse(mustHex(t, tc.wire), WithTolerance(&violations))
			if err != nil {
				t.Fatal(err)
			}
			if got := tc.field(v); got != nil {
				t.Errorf("field = %#v, want nil", got)
			}
			if len(violations) != 1 || violations[0].Path != tc.path {
				t.Errorf("violations = %+v, want one at %s", violations, tc.path)
			}
			sent, err := v.Marshal()
			if err != nil || !bytes.Equal(sent, mustHex(t, tc.sent)) {
				t.Errorf("Marshal = %x, %v; want %s", sent, err, tc.sent)
			}
		})
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
	// (TS 29.002 V19.1.0 §17.7.8); an empty LMSI would select no DA.
	data := mustHex(t, "30118100820291210409010002912100000141")
	_, err := ParseMoFsm(data)
	wantValidationError(t, err, "sm-RP-DA.lmsi", "SIZE (4)")
	violations := []Violation{{Path: "existing"}}
	v, err := ParseMoFsm(data, WithTolerance(&violations))
	if v != nil || !errors.Is(err, ErrSmRpDaLMSIDecodedEmpty) {
		t.Fatalf("Parse = %+v, %v; want nil, %v", v, err, ErrSmRpDaLMSIDecodedEmpty)
	}
	if len(violations) != 1 || violations[0].Path != "existing" {
		t.Fatalf("violations = %+v, want the existing report only", violations)
	}
}

func TestParseToleranceEmptyLMSIMt(t *testing.T) {
	// MT: present zero-octet LMSI DA [1], service-centre OA [4] and an
	// SMS-DELIVER; SEQUENCE length = 2 + 4 + 17 = 23.
	data := mustHex(t, "3017810084029121040f000291210000621010000000000141")
	violations := []Violation{{Path: "existing"}}
	v, err := ParseMtFsm(data, WithTolerance(&violations))
	if v != nil || !errors.Is(err, ErrSmRpDaLMSIDecodedEmpty) {
		t.Fatalf("Parse = %+v, %v; want nil, %v", v, err, ErrSmRpDaLMSIDecodedEmpty)
	}
	if len(violations) != 1 || violations[0].Path != "existing" {
		t.Fatalf("violations = %+v, want the existing report only", violations)
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
// tolerated values to satisfy the encoder's strict constraints, and compares
// tolerant Parse with go-asn1's tolerant decode followed by the same
// conversion: the same outcome, the same value and the same reports.
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
		want, wantViolations, directErr := p.direct(data)
		if (directErr == nil) != (tolerantErr == nil) {
			t.Fatalf("%s: direct tolerant decode err = %v, Parse err = %v", p.name, directErr, tolerantErr)
		}
		if directErr == nil {
			if !reflect.DeepEqual(want, tolerant) {
				t.Fatalf("%s: direct = %+v, Parse = %+v", p.name, want, tolerant)
			}
			if diff := cmp.Diff(wantViolations, violations[1:], cmpopts.IgnoreUnexported(Violation{}), cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("%s reports (-direct +Parse):\n%s", p.name, diff)
			}
		}
		if strictErr == nil || (tolerantErr == nil && len(violations) == 1) {
			if strictErr != nil || tolerantErr != nil || len(violations) != 1 || !reflect.DeepEqual(strict, tolerant) {
				t.Fatalf("%s: strict = %+v, %v; tolerant = %+v, %v; reports = %+v", p.name, strict, strictErr, tolerant, tolerantErr, violations)
			}
		}
	})
}

// X.680 (02/2021) §41.2 Table 9 and X.690 (02/2021) §8.23.4:
// NumericString admits digits and space. An open extType containing 'A'
// is refused strictly and reported once under tolerance.
func TestParseToleranceOpenNumericString(t *testing.T) {
	for _, tc := range []struct {
		name   string
		parser int
		wire   string
		offset int
	}{
		// PrivateExtension = OID TLV 5 + NumericString TLV 3 = 8;
		// list = SEQUENCE TLV 10, container = [0] TLV 12, [6] TLV 14.
		// SRI-SM SEQUENCE length = mandatory fields 21 + 14 = 35.
		{"SriSm", 0, "30238007912143658709f18101008207912143658709f1a60ca00a300806032a0304120141", 34},
		// The same container: MO-ForwardSM length = 19 + 14 = 33.
		{"MoFsm", 1, "302184029121820291210409010002912100000141300ca00a300806032a0304120141", 32},
		// The same container: MT-ForwardSM length = 31 + 14 = 45.
		{"MtFsm", 2, "302d800800010121436587f984029121040f000291210000621010000000000141300ca00a300806032a0304120141", 44},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := tolerantParsers[tc.parser].parse
			data := mustHex(t, tc.wire)
			_, err := p(data)
			var element *ber.ElementError
			if !errors.Is(err, ber.ErrInvalidValue) || !errors.As(err, &element) {
				t.Errorf("strict error = %v, want ErrInvalidValue in *ber.ElementError", err)
			} else if element.Offset() != tc.offset || !reflect.DeepEqual(element.ChildIndexes(), []int{3, 0, 0, 1}) {
				t.Errorf("strict location = %d %v", element.Offset(), element.ChildIndexes())
			}
			violations := []Violation{{Path: "existing"}}
			value, err := p(data, WithTolerance(&violations))
			if err != nil || value == nil {
				t.Fatalf("tolerant Parse = %v, %v", value, err)
			}
			if len(violations) != 2 || violations[0].Path != "existing" {
				t.Fatalf("violations = %+v, want one appended report", violations)
			}
			got := violations[1]
			if got.Kind != ViolationKindASN1Constraint || got.Rule != (ber.RuleRef{Document: "ITU-T X.690", Edition: "02/2021", Clause: "8.23.4"}) ||
				got.Source != "MAP-ExtensionDataTypes.PrivateExtension.extType" ||
				got.Path != "extensionContainer.privateExtensionList.element[0].extType" ||
				got.Constraint != "NumericString digits and space (X.680 (02/2021) §41.2 Table 9; X.690 (02/2021) §8.23.4)" ||
				got.Offset == nil || *got.Offset != tc.offset || !reflect.DeepEqual(got.ChildIndexes, []int{3, 0, 0, 1}) ||
				got.ObservedLength != nil || got.ObservedValue != "41" || got.ObservedValueLength != 2 || got.OctetOffset == nil || *got.OctetOffset != 0 {
				t.Fatalf("violation = %+v", got)
			}
		})
	}
}

// A constructed NumericString is checked as its concatenated contents "A1B"
// (X.690 (02/2021) §§8.23.3, 8.23.6), and tolerance reports each offending
// octet: 'A' at contents offset 0 and 'B' at 2.
func TestParseToleranceOpenSegmentedNumericString(t *testing.T) {
	// Segments TLV 4 "A1" and TLV 3 "B": payload 9=2+7. PrivateExtension
	// 14=5+9, list TLV 16, container [6] content 18, SRI-SM length 21+20=41.
	data := mustHex(t, "30298007912143658709f18101008207912143658709f1a612a010300e06032a0304320704024131040142")
	_, err := ParseSriSm(data)
	var element *ber.ElementError
	if !errors.Is(err, ber.ErrInvalidValue) || !errors.As(err, &element) || element.Offset() != 34 {
		t.Fatalf("strict error = %v, want ErrInvalidValue in a *ber.ElementError at offset 34", err)
	}
	var violations []Violation
	if _, err := ParseSriSm(data, WithTolerance(&violations)); err != nil {
		t.Fatalf("tolerant Parse = %v", err)
	}
	if len(violations) != 2 {
		t.Fatalf("violations = %+v, want two reports", violations)
	}
	for i, want := range []struct {
		value  string
		offset int
	}{{"41", 0}, {"42", 2}} {
		got := violations[i]
		if got.Path != "extensionContainer.privateExtensionList.element[0].extType" || got.ObservedValue != want.value ||
			got.OctetOffset == nil || *got.OctetOffset != want.offset || got.Offset == nil || *got.Offset != 34 {
			t.Errorf("violation %d = %+v, want %s at contents offset %d", i, got, want.value, want.offset)
		}
	}
}

// Under tolerance the other alphabet types are reported like NumericString,
// while width, scalar and time errors are refused in every mode (X.680
// (02/2021) §41 Tables 8–10, §§46.3, 47.3; X.690 (02/2021) §§8.23.4–8.23.8,
// 8.25). Each wire is an SRI-SM whose private extType holds the value; the
// string TLV starts at offset 34.
func TestParseToleranceOpenRestrictedTypes(t *testing.T) {
	for _, tc := range []struct {
		name, wire, alphabet, clause, observed string
	}{
		// Payload 3; SRI-SM length 32+3=35.
		{"PrintableString at sign", "30238007912143658709f18101008207912143658709f1a60ca00a300806032a0304130140", "PrintableString", "8.23.4", "40"},
		// Payload 3; SRI-SM length 32+3=35.
		{"IA5String high octet", "30238007912143658709f18101008207912143658709f1a60ca00a300806032a0304160180", "IA5String", "8.23.5", "80"},
		// Payload 3; SRI-SM length 32+3=35.
		{"VisibleString control", "30238007912143658709f18101008207912143658709f1a60ca00a300806032a03041a011f", "VisibleString", "8.23.5", "1f"},
		// Payload 3; SRI-SM length 32+3=35.
		{"BMPString odd width", "30238007912143658709f18101008207912143658709f1a60ca00a300806032a03041e0100", "", "", ""},
		// Payload 6; SRI-SM length 32+6=38.
		{"UniversalString above U+10FFFF", "30268007912143658709f18101008207912143658709f1a60fa00d300b06032a03041c0400110000", "", "", ""},
		// Payload 13; SRI-SM length 32+13=45.
		{"UTCTime impossible date", "302d8007912143658709f18101008207912143658709f1a616a014301206032a0304170b323630323330303330345a", "", "", ""},
		// Payload 17; SRI-SM length 32+17=49.
		{"GeneralizedTime impossible date", "30318007912143658709f18101008207912143658709f1a61aa018301606032a0304180f32303236303233303033303430355a", "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := mustHex(t, tc.wire)
			if _, err := ParseSriSm(data); !errors.Is(err, ber.ErrInvalidValue) {
				t.Fatalf("strict Parse = %v, want ErrInvalidValue", err)
			}
			var violations []Violation
			_, err := ParseSriSm(data, WithTolerance(&violations))
			if tc.alphabet == "" {
				if !errors.Is(err, ber.ErrInvalidValue) || len(violations) != 0 {
					t.Fatalf("tolerant Parse = %v with %d reports, want ErrInvalidValue and none", err, len(violations))
				}
				return
			}
			if err != nil || len(violations) != 1 {
				t.Fatalf("tolerant Parse = %v with reports %+v, want one report", err, violations)
			}
			got := violations[0]
			if got.Kind != ViolationKindASN1Constraint || got.Rule.Clause != tc.clause || !strings.HasPrefix(got.Constraint, tc.alphabet+" ") ||
				got.ObservedValue != tc.observed || got.Offset == nil || *got.Offset != 34 ||
				got.Path != "extensionContainer.privateExtensionList.element[0].extType" {
				t.Errorf("violation = %+v", got)
			}
		})
	}
}

// X.680 (02/2021) §§41.15, 41.16, X.690 (02/2021) §§8.23.5.3, 8.23.9–8.23.10
// and Unicode 3.2.0 (UAX #27 D7b): the BMPString repertoire, noncharacters in
// BMPString and UniversalString, and BMPString/UniversalString control errors
// remain fatal under tolerance. UTF8String shifts, escapes and noncharacters
// and IA5String escapes produce one report at the offending character; for an
// escape sequence that is its final character, for a noncharacter its first
// octet. Segmentation preserves the contents offset (X.690 §8.23.6).
func TestParseToleranceOpenStringControls(t *testing.T) {
	for _, tc := range []struct {
		name         string
		wires        [3]string
		alphabet     string
		observed     []string // one report per entry, in order
		octetOffsets []int
	}{
		// Payload 4=2+2; private SEQUENCE content 9=5+4, list TLV 13, container TLV 15.
		// Message contents: SRI 36=21+15, MO 34=19+15, MT 46=31+15.
		{"BMPString U+FDD0", [3]string{
			"30248007912143658709f18101008207912143658709f1a60da00b300906032a03041e02fdd0",
			"302284029121820291210409010002912100000141300da00b300906032a03041e02fdd0",
			"302e800800010121436587f984029121040f000291210000621010000000000141300da00b300906032a03041e02fdd0",
		}, "", nil, nil},
		// Payload 6=2+4; private SEQUENCE content 11=5+6, list TLV 15, container TLV 17.
		// Message contents: SRI 38=21+17, MO 36=19+17, MT 48=31+17.
		{"UniversalString U+1FFFE", [3]string{
			"30268007912143658709f18101008207912143658709f1a60fa00d300b06032a03041c040001fffe",
			"302484029121820291210409010002912100000141300fa00d300b06032a03041c040001fffe",
			"3030800800010121436587f984029121040f000291210000621010000000000141300fa00d300b06032a03041c040001fffe",
		}, "", nil, nil},
		// Payload 5=2+3; private SEQUENCE content 10=5+5, list TLV 14, container TLV 16.
		// Message contents: SRI 37=21+16, MO 35=19+16, MT 47=31+16.
		{"UTF8String noncharacter U+FDD0", [3]string{
			"30258007912143658709f18101008207912143658709f1a60ea00c300a06032a03040c03efb790",
			"302384029121820291210409010002912100000141300ea00c300a06032a03040c03efb790",
			"302f800800010121436587f984029121040f000291210000621010000000000141300ea00c300a06032a03040c03efb790",
		}, "noncharacter", []string{"ef"}, []int{0}},
		// Payload 5=2+3; private SEQUENCE content 10=5+5, list TLV 14, container TLV 16.
		// Message contents: SRI 37=21+16, MO 35=19+16, MT 47=31+16.
		{"UTF8String noncharacter U+FFFE", [3]string{
			"30258007912143658709f18101008207912143658709f1a60ea00c300a06032a03040c03efbfbe",
			"302384029121820291210409010002912100000141300ea00c300a06032a03040c03efbfbe",
			"302f800800010121436587f984029121040f000291210000621010000000000141300ea00c300a06032a03040c03efbfbe",
		}, "noncharacter", []string{"ef"}, []int{0}},
		// Payload 9=2+7; private SEQUENCE content 14=5+9, list TLV 18, container TLV 20.
		// Message contents: SRI 41=21+20, MO 39=19+20, MT 51=31+20.
		{"UTF8String two noncharacters", [3]string{
			"30298007912143658709f18101008207912143658709f1a612a010300e06032a03040c07efb790f48fbfbf",
			"3027840291218202912104090100029121000001413012a010300e06032a03040c07efb790f48fbfbf",
			"3033800800010121436587f984029121040f0002912100006210100000000001413012a010300e06032a03040c07efb790f48fbfbf",
		}, "noncharacter", []string{"ef", "f4"}, []int{0, 3}},
		// Payload 9=2+7; private SEQUENCE content 14=5+9, list TLV 18, container TLV 20.
		// Message contents: SRI 41=21+20, MO 39=19+20, MT 51=31+20.
		{"split UTF8String noncharacter", [3]string{
			"30298007912143658709f18101008207912143658709f1a612a010300e06032a03042c070401ef0402bfbe",
			"3027840291218202912104090100029121000001413012a010300e06032a03042c070401ef0402bfbe",
			"3033800800010121436587f984029121040f0002912100006210100000000001413012a010300e06032a03042c070401ef0402bfbe",
		}, "noncharacter", []string{"ef"}, []int{0}},
		// Payload 4=2+2; private SEQUENCE content 5+4=9, list TLV 13, container TLV 15.
		// Message contents: SRI 21+15=36, MO 19+15=34, MT 31+15=46 octets.
		{"BMPString U+FFFE", [3]string{
			"30248007912143658709f18101008207912143658709f1a60da00b300906032a03041e02fffe",
			"302284029121820291210409010002912100000141300da00b300906032a03041e02fffe",
			"302e800800010121436587f984029121040f000291210000621010000000000141300da00b300906032a03041e02fffe",
		}, "", nil, nil},
		// Payload 4=2+2; private SEQUENCE content 5+4=9, list TLV 13, container TLV 15.
		// Message contents: SRI 21+15=36, MO 19+15=34, MT 31+15=46 octets.
		{"BMPString U+FFFF", [3]string{
			"30248007912143658709f18101008207912143658709f1a60da00b300906032a03041e02ffff",
			"302284029121820291210409010002912100000141300da00b300906032a03041e02ffff",
			"302e800800010121436587f984029121040f000291210000621010000000000141300da00b300906032a03041e02ffff",
		}, "", nil, nil},
		// Payload 4=2+2; private SEQUENCE content 5+4=9, list TLV 13, container TLV 15.
		// Message contents: SRI 21+15=36, MO 19+15=34, MT 31+15=46 octets.
		{"BMPString SO", [3]string{
			"30248007912143658709f18101008207912143658709f1a60da00b300906032a03041e02000e",
			"302284029121820291210409010002912100000141300da00b300906032a03041e02000e",
			"302e800800010121436587f984029121040f000291210000621010000000000141300da00b300906032a03041e02000e",
		}, "", nil, nil},
		// Payload 6=2+4; private SEQUENCE content 5+6=11, list TLV 15, container TLV 17.
		// Message contents: SRI 21+17=38, MO 19+17=36, MT 31+17=48 octets.
		{"UniversalString SI", [3]string{
			"30268007912143658709f18101008207912143658709f1a60fa00d300b06032a03041c040000000f",
			"302484029121820291210409010002912100000141300fa00d300b06032a03041c040000000f",
			"3030800800010121436587f984029121040f000291210000621010000000000141300fa00d300b06032a03041c040000000f",
		}, "", nil, nil},
		// Payload 8=2+6; private SEQUENCE content 5+8=13, list TLV 17, container TLV 19.
		// Message contents: SRI 21+19=40, MO 19+19=38, MT 31+19=50 octets.
		{"BMPString designating escape", [3]string{
			"30288007912143658709f18101008207912143658709f1a611a00f300d06032a03041e06001b00280042",
			"3026840291218202912104090100029121000001413011a00f300d06032a03041e06001b00280042",
			"3032800800010121436587f984029121040f0002912100006210100000000001413011a00f300d06032a03041e06001b00280042",
		}, "", nil, nil},
		// Payload 14=2+12; private SEQUENCE content 5+14=19, list TLV 23, container TLV 25.
		// Message contents: SRI 21+25=46, MO 19+25=44, MT 31+25=56 octets.
		{"UniversalString designating escape", [3]string{
			"302e8007912143658709f18101008207912143658709f1a617a015301306032a03041c0c0000001b0000002800000042",
			"302c840291218202912104090100029121000001413017a015301306032a03041c0c0000001b0000002800000042",
			"3038800800010121436587f984029121040f0002912100006210100000000001413017a015301306032a03041c0c0000001b0000002800000042",
		}, "", nil, nil},
		// Payload 5=2+3; private SEQUENCE content 5+5=10, list TLV 14, container TLV 16.
		// Message contents: SRI 21+16=37, MO 19+16=35, MT 31+16=47 octets.
		{"UTF8String escape sequence", [3]string{
			"30258007912143658709f18101008207912143658709f1a60ea00c300a06032a03040c031b2842",
			"302384029121820291210409010002912100000141300ea00c300a06032a03040c031b2842",
			"302f800800010121436587f984029121040f000291210000621010000000000141300ea00c300a06032a03040c031b2842",
		}, "UTF8String", []string{"42"}, []int{2}},
		// Payload 3=2+1; private SEQUENCE content 5+3=8, list TLV 12, container TLV 14.
		// Message contents: SRI 21+14=35, MO 19+14=33, MT 31+14=45 octets.
		{"UTF8String SO", [3]string{
			"30238007912143658709f18101008207912143658709f1a60ca00a300806032a03040c010e",
			"302184029121820291210409010002912100000141300ca00a300806032a03040c010e",
			"302d800800010121436587f984029121040f000291210000621010000000000141300ca00a300806032a03040c010e",
		}, "UTF8String", []string{"0e"}, []int{0}},
		// Payload 5=2+3; private SEQUENCE content 5+5=10, list TLV 14, container TLV 16.
		// Message contents: SRI 21+16=37, MO 19+16=35, MT 31+16=47 octets.
		{"IA5String designating escape", [3]string{
			"30258007912143658709f18101008207912143658709f1a60ea00c300a06032a030416031b2842",
			"302384029121820291210409010002912100000141300ea00c300a06032a030416031b2842",
			"302f800800010121436587f984029121040f000291210000621010000000000141300ea00c300a06032a030416031b2842",
		}, "IA5String", []string{"42"}, []int{2}},
		// Payload 4=2+2; private SEQUENCE content 5+4=9, list TLV 13, container TLV 15.
		// Message contents: SRI 21+15=36, MO 19+15=34, MT 31+15=46 octets.
		{"IA5String ESC N", [3]string{
			"30248007912143658709f18101008207912143658709f1a60da00b300906032a030416021b4e",
			"302284029121820291210409010002912100000141300da00b300906032a030416021b4e",
			"302e800800010121436587f984029121040f000291210000621010000000000141300da00b300906032a030416021b4e",
		}, "IA5String", []string{"4e"}, []int{1}},
		// Payload 9=2+7; private SEQUENCE content 5+9=14, list TLV 18, container TLV 20.
		// Message contents: SRI 21+20=41, MO 19+20=39, MT 31+20=51 octets.
		{"split UTF8String escape sequence", [3]string{
			"30298007912143658709f18101008207912143658709f1a612a010300e06032a03042c0704011b04022842",
			"3027840291218202912104090100029121000001413012a010300e06032a03042c0704011b04022842",
			"3033800800010121436587f984029121040f0002912100006210100000000001413012a010300e06032a03042c0704011b04022842",
		}, "UTF8String", []string{"42"}, []int{2}},
		// Payload 9=2+7; private SEQUENCE content 5+9=14, list TLV 18, container TLV 20.
		// Message contents: SRI 21+20=41, MO 19+20=39, MT 31+20=51 octets.
		{"split IA5String designating escape", [3]string{
			"30298007912143658709f18101008207912143658709f1a612a010300e06032a0304360704011b04022842",
			"3027840291218202912104090100029121000001413012a010300e06032a0304360704011b04022842",
			"3033800800010121436587f984029121040f0002912100006210100000000001413012a010300e06032a0304360704011b04022842",
		}, "IA5String", []string{"42"}, []int{2}},
		// Payload 3=2+1; private SEQUENCE content 5+3=8, list TLV 12, container TLV 14.
		// Message contents: SRI 21+14=35, MO 19+14=33, MT 31+14=45 octets.
		{"UTF8String SI", [3]string{
			"30238007912143658709f18101008207912143658709f1a60ca00a300806032a03040c010f",
			"302184029121820291210409010002912100000141300ca00a300806032a03040c010f",
			"302d800800010121436587f984029121040f000291210000621010000000000141300ca00a300806032a03040c010f",
		}, "UTF8String", []string{"0f"}, []int{0}},
		// Two shifts give two reports. Payload 4=2+2; private SEQUENCE content
		// 5+4=9, list TLV 13, container TLV 15.
		// Message contents: SRI 21+15=36, MO 19+15=34, MT 31+15=46 octets.
		{"UTF8String SO and SI", [3]string{
			"30248007912143658709f18101008207912143658709f1a60da00b300906032a03040c020e0f",
			"302284029121820291210409010002912100000141300da00b300906032a03040c020e0f",
			"302e800800010121436587f984029121040f000291210000621010000000000141300da00b300906032a03040c020e0f",
		}, "UTF8String", []string{"0e", "0f"}, []int{0, 1}},
		// Two escape sequences give two reports, each at its final octet.
		// Payload 8=2+6; private SEQUENCE content 5+8=13, list TLV 17, container TLV 19.
		// Message contents: SRI 21+19=40, MO 19+19=38, MT 31+19=50 octets.
		{"IA5String two designating escapes", [3]string{
			"30288007912143658709f18101008207912143658709f1a611a00f300d06032a030416061b28421b2842",
			"3026840291218202912104090100029121000001413011a00f300d06032a030416061b28421b2842",
			"3032800800010121436587f984029121040f0002912100006210100000000001413011a00f300d06032a030416061b28421b2842",
		}, "IA5String", []string{"42", "42"}, []int{2, 5}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for i, parser := range tolerantParsers {
				t.Run(parser.name, func(t *testing.T) {
					data := mustHex(t, tc.wires[i])
					checkRefusal := func(mode string, err error) {
						t.Helper()
						var element *ber.ElementError
						if !errors.Is(err, ber.ErrInvalidValue) || !errors.As(err, &element) {
							t.Errorf("%s Parse = %v, want ErrInvalidValue in *ber.ElementError", mode, err)
						}
						for _, other := range []error{ber.ErrUnexpectedElement, ber.ErrTruncated, ber.ErrInvalidLength, ber.ErrInvalidTag, ber.ErrExtraData, ber.ErrIndefiniteLength, ber.ErrUnsupported, ber.ErrResourceLimit, ber.ErrInvalidCodec, ber.ErrInvalidBinding} {
							if errors.Is(err, other) {
								t.Errorf("%s Parse = %v also matches %v", mode, err, other)
							}
						}
					}
					_, err := parser.parse(data)
					checkRefusal("strict", err)
					violations := []Violation{{Path: "existing"}}
					value, err := parser.parse(data, WithTolerance(&violations))
					if tc.alphabet == "" {
						checkRefusal("tolerant", err)
						if value != nil || !reflect.DeepEqual(violations, []Violation{{Path: "existing"}}) {
							t.Errorf("tolerant Parse = %+v with reports %+v, want nil and no appended report", value, violations)
						}
						return
					}
					if err != nil || value == nil || len(violations) != 1+len(tc.observed) || violations[0].Path != "existing" {
						t.Fatalf("tolerant Parse = %+v, %v; reports %+v, want %d appended reports", value, err, violations, len(tc.observed))
					}
					offset := []int{34, 32, 44}[i]
					clause := "8.23.10"
					constraint := "UTF8String forbids escapes and code-extension shifts (X.690 (02/2021) §8.23.10; ISO/IEC 10646:2003 §15)"
					if tc.alphabet == "IA5String" {
						clause = "8.23.5.3"
						constraint = "IA5String forbids explicit escape sequences (X.690 (02/2021) §8.23.5.3/Table 3)"
					}
					rule := ber.RuleRef{Document: "ITU-T X.690", Edition: "02/2021", Clause: clause}
					if tc.alphabet == "noncharacter" {
						rule = ber.RuleRef{Document: "ISO/IEC 10646", Edition: "2003", Clause: "7 a), 8"}
						constraint = "UCS repertoire excludes permanently reserved noncharacters (X.680 (02/2021) §§41.6, 41.15–41.16; ISO/IEC 10646:2003 §§7 a), 8)"
					}
					for j, observed := range tc.observed {
						octetOffset := tc.octetOffsets[j]
						want := Violation{
							Kind:       ViolationKindASN1Constraint,
							Rule:       rule,
							Source:     "MAP-ExtensionDataTypes.PrivateExtension.extType",
							Path:       "extensionContainer.privateExtensionList.element[0].extType",
							Constraint: constraint, Offset: &offset, ChildIndexes: []int{3, 0, 0, 1},
							ObservedValue: observed, ObservedValueLength: 2, OctetOffset: &octetOffset,
						}
						if diff := cmp.Diff(want, violations[1+j], cmpopts.IgnoreUnexported(Violation{})); diff != "" {
							t.Errorf("violation %d (-want +got):\n%s", j, diff)
						}
					}
				})
			}
		})
	}
}

// A GeneralizedTime leap second in an open value (X.680 (02/2021) §§3.7 h),
// 46.3 a) 2)) parses in both modes, with no report.
func TestParseToleranceOpenLeapSecond(t *testing.T) {
	// Payload 17=2+15; private SEQUENCE content 5+17=22, list TLV 26, container TLV 28.
	// Message contents: SRI 21+28=49, MO 19+28=47, MT 31+28=59 octets.
	wires := [3]string{
		"30318007912143658709f18101008207912143658709f1a61aa018301606032a0304180f32303136313233313233353936305a",
		"302f84029121820291210409010002912100000141301aa018301606032a0304180f32303136313233313233353936305a",
		"303b800800010121436587f984029121040f000291210000621010000000000141301aa018301606032a0304180f32303136313233313233353936305a",
	}
	for i, parser := range tolerantParsers {
		t.Run(parser.name, func(t *testing.T) {
			data := mustHex(t, wires[i])
			if _, err := parser.parse(data); err != nil {
				t.Fatalf("strict Parse = %v, want accepted", err)
			}
			var violations []Violation
			value, err := parser.parse(data, WithTolerance(&violations))
			if err != nil || value == nil || len(violations) != 0 {
				t.Fatalf("tolerant Parse = %+v, %v; reports %+v, want accepted with none", value, err, violations)
			}
		})
	}
}
