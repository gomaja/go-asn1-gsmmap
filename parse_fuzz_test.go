package gsmmap

import (
	"encoding/hex"
	"errors"
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

// FuzzParse drives every Parse function. None may panic, and a value that
// parses and marshals again parses back to the same value.
func FuzzParse(f *testing.F) {
	for _, s := range berSeeds(f) {
		for i := range parsers {
			f.Add(uint8(i), s)
		}
	}
	f.Fuzz(func(t *testing.T, which uint8, data []byte) {
		p := parsers[int(which)%len(parsers)]
		v, err := p.parse(data)
		if err != nil {
			if v != nil {
				t.Fatalf("%s: error %v came with a value", p.name, err)
			}
			return
		}
		enc, err := v.Marshal()
		if err != nil {
			return
		}
		again, err := p.parse(enc)
		if err != nil {
			t.Fatalf("%s: re-parsing marshalled value: %v\nwire %x", p.name, err, enc)
		}
		if diff := cmp.Diff(v, again, equateTPDU, cmpopts.EquateEmpty()); diff != "" {
			t.Fatalf("%s round trip (-first +second):\n%s\nwire %x", p.name, diff, data)
		}
		if reflect.TypeOf(v) != reflect.TypeOf(again) {
			t.Fatalf("%s: type changed", p.name)
		}
	})
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
