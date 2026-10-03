package gsmmap

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// marshaler is the Marshal method every parsed type has.
type marshaler interface {
	Marshal() ([]byte, error)
}

func asParser[T any](parse func([]byte) (*T, error)) func([]byte) (marshaler, error) {
	return func(data []byte) (marshaler, error) {
		v, err := parse(data)
		if err != nil || v == nil {
			return nil, err
		}
		m, ok := any(v).(marshaler)
		if !ok {
			panic("parsed type has no Marshal method")
		}
		return m, nil
	}
}

// parsers lists every exported Parse function of the package.
var parsers = []struct {
	name  string
	parse func([]byte) (marshaler, error)
}{
	{"SriSm", asParser(ParseSriSm)},
	{"SriSmResp", asParser(ParseSriSmResp)},
	{"MtFsm", asParser(ParseMtFsm)},
	{"MtFsmResp", asParser(ParseMtFsmResp)},
	{"MoFsm", asParser(ParseMoFsm)},
	{"MoFsmResp", asParser(ParseMoFsmResp)},
	{"UpdateLocation", asParser(ParseUpdateLocation)},
	{"UpdateLocationRes", asParser(ParseUpdateLocationRes)},
	{"UpdateGprsLocation", asParser(ParseUpdateGprsLocation)},
	{"UpdateGprsLocationRes", asParser(ParseUpdateGprsLocationRes)},
	{"AnyTimeInterrogation", asParser(ParseAnyTimeInterrogation)},
	{"AnyTimeInterrogationRes", asParser(ParseAnyTimeInterrogationRes)},
	{"ProvideSubscriberInfo", asParser(ParseProvideSubscriberInfo)},
	{"ProvideSubscriberInfoRes", asParser(ParseProvideSubscriberInfoRes)},
	{"Sri", asParser(ParseSri)},
	{"SriResp", asParser(ParseSriResp)},
	{"InformServiceCentre", asParser(ParseInformServiceCentre)},
	{"AlertServiceCentre", asParser(ParseAlertServiceCentre)},
	{"PurgeMS", asParser(ParsePurgeMS)},
	{"PurgeMSRes", asParser(ParsePurgeMSRes)},
	{"SendAuthenticationInfo", asParser(ParseSendAuthenticationInfo)},
	{"SendAuthenticationInfoRes", asParser(ParseSendAuthenticationInfoRes)},
	{"CancelLocation", asParser(ParseCancelLocation)},
	{"CancelLocationRes", asParser(ParseCancelLocationRes)},
	{"InsertSubscriberData", asParser(ParseInsertSubscriberData)},
	{"InsertSubscriberDataRes", asParser(ParseInsertSubscriberDataRes)},
	{"ProvideSubscriberLocation", asParser(ParseProvideSubscriberLocation)},
	{"ProvideSubscriberLocationRes", asParser(ParseProvideSubscriberLocationRes)},
	{"SubscriberLocationReport", asParser(ParseSubscriberLocationReport)},
	{"SubscriberLocationReportRes", asParser(ParseSubscriberLocationReportRes)},
	{"SriLcs", asParser(ParseSriLcs)},
	{"SriLcsResp", asParser(ParseSriLcsResp)},
	{"ReportSMDeliveryStatus", asParser(ParseReportSMDeliveryStatus)},
	{"ReportSMDeliveryStatusRes", asParser(ParseReportSMDeliveryStatusRes)},
	{"USSDArg", asParser(ParseUSSDArg)},
	{"USSDRes", asParser(ParseUSSDRes)},
}

// TestParsersCoverEveryParseFunction keeps the fuzz table in step with
// the package: every exported Parse function returning a pointer is listed.
func TestParsersCoverEveryParseFunction(t *testing.T) {
	src, err := os.ReadFile("parse.go")
	if err != nil {
		t.Fatal(err)
	}
	names := regexp.MustCompile(`(?m)^func Parse(\w+)\(data \[\]byte\) \(\*`).FindAllSubmatch(src, -1)
	listed := map[string]bool{}
	for _, p := range parsers {
		listed[p.name] = true
	}
	for _, m := range names {
		if !listed[string(m[1])] {
			t.Errorf("Parse%s is missing from the fuzz table", m[1])
		}
	}
	if len(names) != len(parsers) {
		t.Errorf("parse.go has %d Parse functions, the fuzz table %d", len(names), len(parsers))
	}
}

// berSeeds collects the BER hex vectors of the package's tests.
func berSeeds(tb testing.TB) [][]byte {
	files, err := filepath.Glob("*_test.go")
	if err != nil {
		tb.Fatal(err)
	}
	re := regexp.MustCompile(`"((?:30|a[0-9a-f]|3081)[0-9a-fA-F]{6,})"`)
	seen := map[string]bool{}
	var seeds [][]byte
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			tb.Fatal(err)
		}
		for _, m := range re.FindAllSubmatch(src, -1) {
			b, err := hex.DecodeString(string(m[1]))
			if err != nil || seen[string(b)] {
				continue
			}
			seen[string(b)] = true
			seeds = append(seeds, b)
		}
	}
	return seeds
}

// strictEncodeErrors are the only errors Marshal may return for a value a
// Parse function returned. Each marks a value the package decodes
// leniently, as 3GPP TS 29.002 V19.1.0 tells a receiver to, but never sends:
var strictEncodeErrors = []error{
	// A reserved AlertingPattern (§17.7.8) is decoded and not sent.
	ErrAlertingPatternReserved,
	// A 15-digit IMEI whose last digit is a peer's Check Digit instead of
	// the spare digit 0 (§17.7.8 IMEI, TS 23.003 §6.2.1).
	ErrIMEISpareDigitNotZero,
	// Unknown values of extensible ENUMERATEDs are kept (§17.1.4); negative
	// values lie outside the ranges the exception handling maps.
	ErrCancelLocInvalidCancellationType,
	ErrCancelLocInvalidTypeOfUpdate,
	ErrCamelInvalidTTriggerPoint,
	ErrCamelInvalidDefaultCallHandling,
	ErrCamelInvalidDefaultSMSHandling,
	ErrDefaultGPRSHandlingInvalid,
	ErrSaiInvalidRequestingNodeType,
	ErrRequestedDomainInvalid,
	ErrISTSupportIndicatorInvalid,
	ErrUnavailabilityCauseInvalid,
	ErrLCSClientInternalIDInvalid,
	ErrLCSFormatIndicatorInvalid,
	ErrAccuracyFulfilmentIndicatorInvalid,
	ErrAreaTypeInvalid,
	ErrOccurrenceInfoInvalid,
	ErrRANTechnologyInvalid,
	// An unrecognized LCSClientType is kept only in a
	// ProvideSubscriberLocation-Arg with privacyOverride (§17.7.13).
	ErrLCSClientTypeInvalid,
}

// isStrictEncodeError reports whether err is one of strictEncodeErrors.
func isStrictEncodeError(err error) bool {
	for _, e := range strictEncodeErrors {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}

// FuzzParse drives every Parse function. None may panic, a value that
// parses marshals again unless Marshal reports one of strictEncodeErrors,
// and the marshalled value parses back to the same value.
func FuzzParse(f *testing.F) {
	for _, s := range berSeeds(f) {
		for i := range parsers {
			f.Add(uint8(i), s)
		}
	}
	f.Fuzz(func(t *testing.T, which uint8, data []byte) {
		p := parsers[int(which)%len(parsers)]
		if err := checkParseRoundTrip(p.name, p.parse, data); err != nil {
			t.Fatal(err)
		}
	})
}

// checkParseRoundTrip applies the FuzzParse property to data and parse:
// an error comes without a value; a parsed value marshals, unless Marshal
// reports one of strictEncodeErrors; and the marshalled octets parse back
// to an equal value of the same type.
func checkParseRoundTrip(name string, parse func([]byte) (marshaler, error), data []byte) error {
	v, err := parse(data)
	if err != nil {
		if v != nil {
			return fmt.Errorf("%s: error %v came with a value", name, err)
		}
		return nil
	}
	enc, err := v.Marshal()
	if err != nil {
		if isStrictEncodeError(err) {
			return nil
		}
		return fmt.Errorf("%s: parsed value does not marshal: %w\nwire %x", name, err, data)
	}
	again, err := parse(enc)
	if err != nil {
		return fmt.Errorf("%s: re-parsing marshalled value: %w\nwire %x", name, err, enc)
	}
	if diff := cmp.Diff(v, again, equateTPDU, cmpopts.EquateEmpty()); diff != "" {
		return fmt.Errorf("%s round trip (-first +second):\n%s\nwire %x", name, diff, data)
	}
	if reflect.TypeOf(v) != reflect.TypeOf(again) {
		return fmt.Errorf("%s: type changed", name)
	}
	return nil
}

// A wire storedMSISDN that holds only its nature/plan octet has no digits;
// StoredMSISDN "" means absent, so it is rejected rather than dropped.
func TestParseInformServiceCentreStoredMSISDNWithoutDigits(t *testing.T) {
	data, err := hex.DecodeString("30030401" + "91")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseInformServiceCentre(data); !errors.Is(err, ErrIscStoredMSISDNDecodedEmpty) {
		t.Errorf("err = %v, want ErrIscStoredMSISDNDecodedEmpty", err)
	}
}
