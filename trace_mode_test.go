package gsmmap

import (
	"bytes"
	"errors"
	"fmt"
	identity "github.com/gomaja/go-3gpp-identity"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
	"reflect"
	"strings"
	"testing"
)

// TS 29.002 V19.1.0 §17.7.2: [1] reference 01, [2] trace type 0.
func TestTraceModeMandatory(t *testing.T) {
	wire := mustHex(t, "3006810101820100")
	a := &ActivateTraceModeArg{TraceReference: HexBytes{1}}
	b, err := a.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, wire) {
		t.Fatalf("wire = %x, want %x", b, wire)
	}
	got, err := ParseActivateTraceModeArg(wire)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "activate", a, got)

	// [1] reference 01; no optional fields.
	dwire := mustHex(t, "3003810101")
	d := &DeactivateTraceModeArg{TraceReference: HexBytes{1}}
	b, err = d.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, dwire) {
		t.Fatalf("wire = %x, want %x", b, dwire)
	}
	dg, err := ParseDeactivateTraceModeArg(dwire)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "deactivate", d, dg)
}

func traceArg() *ActivateTraceModeArg { return &ActivateTraceModeArg{TraceReference: HexBytes{1}} }

func traceRawDepth(tag byte, v byte) []byte {
	return traceTLV(0x30, append([]byte{0x81, 1, 1, 0x82, 1, 0}, traceTLV(0xa6, []byte{tag, 1, v})...))
}

func TestTraceModeDepthRules(t *testing.T) {
	// TS 29.002 V19.1.0 §17.7.2: all ten base/extension pairs, tags 0..19.
	for i, name := range []string{"MscSTraceDepth", "MgwTraceDepth", "SgsnTraceDepth", "GgsnTraceDepth", "RncTraceDepth", "BmscTraceDepth", "MmeTraceDepth", "SgwTraceDepth", "PgwTraceDepth", "ENBTraceDepth"} {
		t.Run(name, func(t *testing.T) {
			for _, extension := range []bool{false, true} {
				field, tag := i, byte(0x80+i)
				if extension {
					field += 10
					tag += 10
				}
				for _, value := range []byte{0, 1, 2, 3, 127, 255} {
					got, err := ParseActivateTraceModeArg(traceRawDepth(tag, value))
					if err != nil {
						t.Fatal(err)
					}
					f := reflect.ValueOf(got.TraceDepthList).Elem().Field(field)
					want := int64(value)
					if value == 255 {
						want = -1
					}
					if i != 4 && value == 1 {
						want = 0
					}
					if f.IsNil() || f.Elem().Int() != want {
						t.Fatalf("field %d value %d: got %v, want %d", field, value, f, want)
					}
					_, err = got.Marshal()
					switch {
					case value > 2 && extension:
						if !errors.Is(err, ErrTraceDepthExtensionInvalid) {
							t.Fatal(err)
						}
					case value > 2:
						if !errors.Is(err, ErrTraceDepthInvalid) {
							t.Fatal(err)
						}
					case extension:
						if !errors.Is(err, ErrTraceDepthExtensionMismatch) {
							t.Fatal(err)
						}
					default:
						if err != nil {
							t.Fatal(err)
						}
					}
				}
			}
			for _, values := range [][2]int64{{0, 0}, {1, 1}, {2, 2}, {0, 2}, {2, 0}} {
				a := traceArg()
				a.TraceDepthList = &TraceDepthList{}
				v := reflect.ValueOf(a.TraceDepthList).Elem()
				b, e := TraceDepth(values[0]), TraceDepthExtension(values[1])
				v.Field(i).Set(reflect.ValueOf(&b))
				v.Field(i + 10).Set(reflect.ValueOf(&e))
				_, err := a.Marshal()
				switch {
				case i != 4 && values[0] == 1:
					if !errors.Is(err, ErrTraceDepthMediumNonRNC) {
						t.Fatal(err)
					}
				case values[0] != values[1]:
					if !errors.Is(err, ErrTraceDepthExtensionMismatch) {
						t.Fatal(err)
					}
				default:
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			if i != 4 {
				a := traceArg()
				a.TraceDepthList = &TraceDepthList{}
				b := TraceDepthMedium
				reflect.ValueOf(a.TraceDepthList).Elem().Field(i).Set(reflect.ValueOf(&b))
				_, err := a.Marshal()
				if !errors.Is(err, ErrTraceDepthMediumNonRNC) {
					t.Fatal(err)
				}
			}
			// Medium extension is forbidden even when its base is absent or minimum.
			if i != 4 {
				a := traceArg()
				a.TraceDepthList = &TraceDepthList{}
				e := TraceDepthExtensionMediumWithoutVendorSpecificExtension
				reflect.ValueOf(a.TraceDepthList).Elem().Field(i + 10).Set(reflect.ValueOf(&e))
				_, err := a.Marshal()
				if !errors.Is(err, ErrTraceDepthExtensionMediumNonRNC) {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestTraceModeNilAndResults(t *testing.T) {
	cases := []struct {
		v   marshaler
		err error
	}{
		{(*ActivateTraceModeArg)(nil), ErrActivateTraceModeArgNil},
		{(*ActivateTraceModeRes)(nil), ErrActivateTraceModeResNil},
		{(*DeactivateTraceModeArg)(nil), ErrDeactivateTraceModeArgNil},
		{(*DeactivateTraceModeRes)(nil), ErrDeactivateTraceModeResNil},
	}
	for _, c := range cases {
		_, err := c.v.Marshal()
		if !errors.Is(err, c.err) {
			t.Errorf("got %v, want %v", err, c.err)
		}
	}
	for _, flag := range []bool{false, true} {
		// Empty sequence, or [1] traceSupportIndicator NULL.
		want := mustHex(t, "3000")
		if flag {
			want = mustHex(t, "30028100")
		}
		r := &ActivateTraceModeRes{TraceSupportIndicator: flag}
		b, err := r.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b, want) {
			t.Fatalf("got %x, want %x", b, want)
		}
		got, err := ParseActivateTraceModeRes(want)
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "result", r, got)
	}
	b, err := (&DeactivateTraceModeRes{}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, []byte{0x30, 0}) {
		t.Fatalf("%x", b)
	}
	r, err := ParseDeactivateTraceModeRes(b)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "result", &DeactivateTraceModeRes{}, r)
	for _, parse := range []func([]byte) (marshaler, error){asParser(ParseActivateTraceModeArg), asParser(ParseActivateTraceModeRes), asParser(ParseDeactivateTraceModeArg), asParser(ParseDeactivateTraceModeRes)} {
		if _, err := parse(nil); err == nil {
			t.Fatal("empty input accepted")
		}
	}
}

func TestTraceModeNamedBits(t *testing.T) {
	// §17.7.2: every listed bit independently, an empty list, and the first unlisted bit.
	cases := []struct {
		name  string
		n     int
		max   int
		value any
		put   func(*ActivateTraceModeArg, any)
		get   func(*ActivateTraceModeArg) any
		wrap  func([]byte) []byte
	}{
		{"TraceNETypeList", 10, 16, &TraceNETypeList{}, func(a *ActivateTraceModeArg, v any) { a.TraceNETypeList = v.(*TraceNETypeList) }, func(a *ActivateTraceModeArg) any { return a.TraceNETypeList }, func(b []byte) []byte { return traceTLV(0x87, b) }},
		{"MSCSInterfaceList", 10, 16, &MSCSInterfaceList{}, func(a *ActivateTraceModeArg, v any) {
			a.TraceInterfaceList = &TraceInterfaceList{MscSList: v.(*MSCSInterfaceList)}
		}, func(a *ActivateTraceModeArg) any { return a.TraceInterfaceList.MscSList }, func(b []byte) []byte { return traceTLV(168, traceTLV(128, b)) }},
		{"MGWInterfaceList", 3, 8, &MGWInterfaceList{}, func(a *ActivateTraceModeArg, v any) {
			a.TraceInterfaceList = &TraceInterfaceList{MgwList: v.(*MGWInterfaceList)}
		}, func(a *ActivateTraceModeArg) any { return a.TraceInterfaceList.MgwList }, func(b []byte) []byte { return traceTLV(168, traceTLV(129, b)) }},
		{"SGSNInterfaceList", 11, 16, &SGSNInterfaceList{}, func(a *ActivateTraceModeArg, v any) {
			a.TraceInterfaceList = &TraceInterfaceList{SgsnList: v.(*SGSNInterfaceList)}
		}, func(a *ActivateTraceModeArg) any { return a.TraceInterfaceList.SgsnList }, func(b []byte) []byte { return traceTLV(168, traceTLV(130, b)) }},
		{"GGSNInterfaceList", 3, 8, &GGSNInterfaceList{}, func(a *ActivateTraceModeArg, v any) {
			a.TraceInterfaceList = &TraceInterfaceList{GgsnList: v.(*GGSNInterfaceList)}
		}, func(a *ActivateTraceModeArg) any { return a.TraceInterfaceList.GgsnList }, func(b []byte) []byte { return traceTLV(168, traceTLV(131, b)) }},
		{"RNCInterfaceList", 4, 8, &RNCInterfaceList{}, func(a *ActivateTraceModeArg, v any) {
			a.TraceInterfaceList = &TraceInterfaceList{RncList: v.(*RNCInterfaceList)}
		}, func(a *ActivateTraceModeArg) any { return a.TraceInterfaceList.RncList }, func(b []byte) []byte { return traceTLV(168, traceTLV(132, b)) }},
		{"BMSCInterfaceList", 1, 8, &BMSCInterfaceList{}, func(a *ActivateTraceModeArg, v any) {
			a.TraceInterfaceList = &TraceInterfaceList{BmscList: v.(*BMSCInterfaceList)}
		}, func(a *ActivateTraceModeArg) any { return a.TraceInterfaceList.BmscList }, func(b []byte) []byte { return traceTLV(168, traceTLV(133, b)) }},
		{"MMEInterfaceList", 5, 8, &MMEInterfaceList{}, func(a *ActivateTraceModeArg, v any) {
			a.TraceInterfaceList = &TraceInterfaceList{MmeList: v.(*MMEInterfaceList)}
		}, func(a *ActivateTraceModeArg) any { return a.TraceInterfaceList.MmeList }, func(b []byte) []byte { return traceTLV(168, traceTLV(134, b)) }},
		{"SGWInterfaceList", 5, 8, &SGWInterfaceList{}, func(a *ActivateTraceModeArg, v any) {
			a.TraceInterfaceList = &TraceInterfaceList{SgwList: v.(*SGWInterfaceList)}
		}, func(a *ActivateTraceModeArg) any { return a.TraceInterfaceList.SgwList }, func(b []byte) []byte { return traceTLV(168, traceTLV(135, b)) }},
		{"PGWInterfaceList", 8, 16, &PGWInterfaceList{}, func(a *ActivateTraceModeArg, v any) {
			a.TraceInterfaceList = &TraceInterfaceList{PgwList: v.(*PGWInterfaceList)}
		}, func(a *ActivateTraceModeArg) any { return a.TraceInterfaceList.PgwList }, func(b []byte) []byte { return traceTLV(168, traceTLV(136, b)) }},
		{"ENBInterfaceList", 3, 8, &ENBInterfaceList{}, func(a *ActivateTraceModeArg, v any) {
			a.TraceInterfaceList = &TraceInterfaceList{ENBList: v.(*ENBInterfaceList)}
		}, func(a *ActivateTraceModeArg) any { return a.TraceInterfaceList.ENBList }, func(b []byte) []byte { return traceTLV(168, traceTLV(137, b)) }},
		{"MSCSEventList", 5, 16, &MSCSEventList{}, func(a *ActivateTraceModeArg, v any) { a.TraceEventList = &TraceEventList{MscSList: v.(*MSCSEventList)} }, func(a *ActivateTraceModeArg) any { return a.TraceEventList.MscSList }, func(b []byte) []byte { return traceTLV(169, traceTLV(128, b)) }},
		{"MGWEventList", 1, 8, &MGWEventList{}, func(a *ActivateTraceModeArg, v any) { a.TraceEventList = &TraceEventList{MgwList: v.(*MGWEventList)} }, func(a *ActivateTraceModeArg) any { return a.TraceEventList.MgwList }, func(b []byte) []byte { return traceTLV(169, traceTLV(129, b)) }},
		{"SGSNEventList", 4, 16, &SGSNEventList{}, func(a *ActivateTraceModeArg, v any) { a.TraceEventList = &TraceEventList{SgsnList: v.(*SGSNEventList)} }, func(a *ActivateTraceModeArg) any { return a.TraceEventList.SgsnList }, func(b []byte) []byte { return traceTLV(169, traceTLV(130, b)) }},
		{"GGSNEventList", 2, 8, &GGSNEventList{}, func(a *ActivateTraceModeArg, v any) { a.TraceEventList = &TraceEventList{GgsnList: v.(*GGSNEventList)} }, func(a *ActivateTraceModeArg) any { return a.TraceEventList.GgsnList }, func(b []byte) []byte { return traceTLV(169, traceTLV(131, b)) }},
		{"BMSCEventList", 1, 8, &BMSCEventList{}, func(a *ActivateTraceModeArg, v any) { a.TraceEventList = &TraceEventList{BmscList: v.(*BMSCEventList)} }, func(a *ActivateTraceModeArg) any { return a.TraceEventList.BmscList }, func(b []byte) []byte { return traceTLV(169, traceTLV(132, b)) }},
		{"MMEEventList", 6, 8, &MMEEventList{}, func(a *ActivateTraceModeArg, v any) { a.TraceEventList = &TraceEventList{MmeList: v.(*MMEEventList)} }, func(a *ActivateTraceModeArg) any { return a.TraceEventList.MmeList }, func(b []byte) []byte { return traceTLV(169, traceTLV(133, b)) }},
		{"SGWEventList", 3, 8, &SGWEventList{}, func(a *ActivateTraceModeArg, v any) { a.TraceEventList = &TraceEventList{SgwList: v.(*SGWEventList)} }, func(a *ActivateTraceModeArg) any { return a.TraceEventList.SgwList }, func(b []byte) []byte { return traceTLV(169, traceTLV(134, b)) }},
		{"PGWEventList", 3, 8, &PGWEventList{}, func(a *ActivateTraceModeArg, v any) { a.TraceEventList = &TraceEventList{PgwList: v.(*PGWEventList)} }, func(a *ActivateTraceModeArg) any { return a.TraceEventList.PgwList }, func(b []byte) []byte { return traceTLV(169, traceTLV(135, b)) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for bit := -1; bit < c.n; bit++ {
				v := reflect.New(reflect.TypeOf(c.value).Elem())
				if bit >= 0 {
					v.Elem().Field(bit).SetBool(true)
				}
				a := traceArg()
				c.put(a, v.Interface())
				b, err := a.Marshal()
				if err != nil {
					t.Fatal(err)
				}
				got, err := ParseActivateTraceModeArg(b)
				if err != nil {
					t.Fatal(err)
				}
				wantEqual(t, "named", v.Interface(), c.get(got))
				// Independent wire expectation: ASN.1 bit 0 is MSB. Raw strings use max SIZE.
				raw := make([]byte, 1+(c.max+7)/8)
				raw[0] = byte((8 - c.max%8) % 8)
				if bit >= 0 {
					raw[1+bit/8] |= 0x80 >> uint(bit%8)
				}
				wire := traceTLV(0x30, append(mustHex(t, "810101820100"), c.wrap(raw)...))
				parsed, err := ParseActivateTraceModeArg(wire)
				if err != nil {
					t.Fatal(err)
				}
				wantEqual(t, "raw named", v.Interface(), c.get(parsed))
			}
			// A set bit past the ASN.1 upper bound cannot be normalized away.
			tooLong := make([]byte, 1+(c.max+8)/8)
			tooLong[0] = 7
			tooLong[1+c.max/8] = 0x80 >> uint(c.max%8)
			_, err := ParseActivateTraceModeArg(traceTLV(0x30, append(mustHex(t, "810101820100"), c.wrap(tooLong)...)))
			var violation *ber.ValidationError
			if !errors.As(err, &violation) || !strings.HasPrefix(violation.Constraint, "SIZE (") {
				t.Fatalf("set bit %d: got %v, want codec SIZE error", c.max, err)
			}
			for _, known := range []bool{false, true} {
				raw := make([]byte, 1+(c.max+7)/8)
				raw[0] = byte((8 - c.max%8) % 8)
				if known {
					raw[1] = 0x80
				}
				raw[1+c.n/8] |= 0x80 >> uint(c.n%8)
				wire := traceTLV(0x30, append(mustHex(t, "810101820100"), c.wrap(raw)...))
				got, err := ParseActivateTraceModeArg(wire)
				if err != nil {
					t.Fatal(err)
				}
				want := reflect.New(reflect.TypeOf(c.value).Elem())
				want.Elem().Field(0).SetBool(known)
				wantEqual(t, "unknown discarded", want.Interface(), c.get(got))
				canonical, err := got.Marshal()
				if err != nil {
					t.Fatal(err)
				}
				again, err := ParseActivateTraceModeArg(canonical)
				if err != nil {
					t.Fatal(err)
				}
				wantEqual(t, "canonical", got, again)
			}
		})
	}
}

// Parse preserves sender-invalid extensible enumeration values (§17.7.2).
func traceModeLenientCases() []semLenientCase {
	return []semLenientCase{
		{ErrTraceDepthInvalid, asParser(ParseActivateTraceModeArg), func(*testing.T) semWire {
			d := gsm_map.TraceDepth(3)
			return &gsm_map.ActivateTraceModeArg{TraceReference: []byte{1}, TraceDepthList: &gsm_map.TraceDepthList{RncTraceDepth: &d}}
		}},
		{ErrTraceDepthExtensionInvalid, asParser(ParseActivateTraceModeArg), func(*testing.T) semWire {
			d := gsm_map.TraceDepthExtension(3)
			return &gsm_map.ActivateTraceModeArg{TraceReference: []byte{1}, TraceDepthList: &gsm_map.TraceDepthList{RncTraceDepthExtension: &d}}
		}},
		{ErrTraceDepthExtensionMismatch, asParser(ParseActivateTraceModeArg), func(*testing.T) semWire {
			d := gsm_map.TraceDepthExtensionMaximumWithoutVendorSpecificExtension
			return &gsm_map.ActivateTraceModeArg{TraceReference: []byte{1}, TraceDepthList: &gsm_map.TraceDepthList{RncTraceDepthExtension: &d}}
		}},
	}
}

func traceFullArg() *ActivateTraceModeArg {
	min, max, med := TraceDepthMinimum, TraceDepthMaximum, TraceDepthMedium
	emin, emax, emed := TraceDepthExtensionMinimumWithoutVendorSpecificExtension, TraceDepthExtensionMaximumWithoutVendorSpecificExtension, TraceDepthExtensionMediumWithoutVendorSpecificExtension
	interval, amount, logging, duration := ReportIntervalLte1min, ReportAmountInfinity, LoggingIntervalD61dot44, LoggingDurationD7200sec
	umts, lte, rrmumts, rrmlte := PeriodUMTSD64000ms, PeriodLTED1min, PeriodUMTSD250ms, PeriodLTED1280ms
	rsrp, rsrq, f, i := 97, 34, 165, -25
	return &ActivateTraceModeArg{
		IMSI: "001010123456789", TraceReference: HexBytes{0xaa, 0xbb}, TraceType: 7,
		OMCID: "12345", OMCIDNature: 0x10, OMCIDPlan: 1, TraceReference2: HexBytes{0x11, 0x22, 0x33},
		TraceDepthList: &TraceDepthList{
			MscSTraceDepth:          &min,
			MgwTraceDepth:           &max,
			SgsnTraceDepth:          &min,
			GgsnTraceDepth:          &max,
			RncTraceDepth:           &med,
			BmscTraceDepth:          &max,
			MmeTraceDepth:           &min,
			SgwTraceDepth:           &max,
			PgwTraceDepth:           &min,
			ENBTraceDepth:           &max,
			MscSTraceDepthExtension: &emin,
			MgwTraceDepthExtension:  &emax,
			SgsnTraceDepthExtension: &emin,
			GgsnTraceDepthExtension: &emax,
			RncTraceDepthExtension:  &emed,
			BmscTraceDepthExtension: &emax,
			MmeTraceDepthExtension:  &emin,
			SgwTraceDepthExtension:  &emax,
			PgwTraceDepthExtension:  &emin,
			ENBTraceDepthExtension:  &emax,
		},
		TraceNETypeList: &TraceNETypeList{MscS: true, Mgw: true, Sgsn: true, Ggsn: true, Rnc: true, Bmsc: true, Mme: true, Sgw: true, Pgw: true, ENB: true},
		TraceInterfaceList: &TraceInterfaceList{
			MscSList: &MSCSInterfaceList{A: true, Iu: true, Mc: true, MapG: true, MapB: true, MapE: true, MapF: true, Cap: true, MapD: true, MapC: true},
			MgwList:  &MGWInterfaceList{Mc: true, NbUp: true, IuUp: true},
			SgsnList: &SGSNInterfaceList{Gb: true, Iu: true, Gn: true, MapGr: true, MapGd: true, MapGf: true, Gs: true, Ge: true, S3: true, S4: true, S6d: true},
			GgsnList: &GGSNInterfaceList{Gn: true, Gi: true, Gmb: true},
			RncList:  &RNCInterfaceList{Iu: true, Iur: true, Iub: true, Uu: true},
			BmscList: &BMSCInterfaceList{Gmb: true},
			MmeList:  &MMEInterfaceList{S1Mme: true, S3: true, S6a: true, S10: true, S11: true},
			SgwList:  &SGWInterfaceList{S4: true, S5: true, S8b: true, S11: true, Gxc: true},
			PgwList:  &PGWInterfaceList{S2a: true, S2b: true, S2c: true, S5: true, S6b: true, Gx: true, S8b: true, Sgi: true},
			ENBList:  &ENBInterfaceList{S1Mme: true, X2: true, Uu: true},
		},
		TraceEventList: &TraceEventList{
			MscSList: &MSCSEventList{MoMtCall: true, MoMtSms: true, LuImsiAttachImsiDetach: true, Handovers: true, Ss: true},
			MgwList:  &MGWEventList{Context: true},
			SgsnList: &SGSNEventList{PdpContext: true, MoMtSms: true, RauGprsAttachGprsDetach: true, MbmsContext: true},
			GgsnList: &GGSNEventList{PdpContext: true, MbmsContext: true},
			BmscList: &BMSCEventList{MbmsMulticastServiceActivation: true},
			MmeList:  &MMEEventList{UeInitiatedPDNconectivityRequest: true, ServiceRequestts: true, InitialAttachTrackingAreaUpdateDetach: true, UeInitiatedPDNdisconnection: true, BearerActivationModificationDeletion: true, Handover: true},
			SgwList:  &SGWEventList{PdnConnectionCreation: true, PdnConnectionTermination: true, BearerActivationModificationDeletion: true},
			PgwList:  &PGWEventList{PdnConnectionCreation: true, PdnConnectionTermination: true, BearerActivationModificationDeletion: true},
		},
		TraceCollectionEntity: "192.0.2.1",
		MDTConfiguration: &MDTConfiguration{JobType: JobTypeImmediateMDTAndTrace,
			AreaScope: &AreaScope{
				CgiList:            []HexBytes{{0x21, 0xf3, 0x54, 0, 1, 0, 2}},
				EUtranCgiList:      []HexBytes{{0x21, 0xf3, 0x54, 0, 0, 0, 3}},
				RoutingAreaIdList:  []HexBytes{{0x21, 0xf3, 0x54, 0, 4, 5}},
				LocationAreaIdList: []HexBytes{{0x21, 0xf3, 0x54, 0, 6}},
				TrackingAreaIdList: []HexBytes{{0x21, 0xf3, 0x54, 0, 7}},
			},
			ListOfMeasurements: HexBytes{1, 2, 3, 4}, ReportingTrigger: HexBytes{1},
			ReportInterval: &interval, ReportAmount: &amount, EventThresholdRSRP: &rsrp, EventThresholdRSRQ: &rsrq,
			LoggingInterval: &logging, LoggingDuration: &duration, MeasurementPeriodUMTS: &umts, MeasurementPeriodLTE: &lte,
			CollectionPeriodRRMUMTS: &rrmumts, CollectionPeriodRRMLTE: &rrmlte, PositioningMethod: HexBytes{2}, MeasurementQuantity: HexBytes{3},
			EventThreshold1F: &f, EventThreshold1I: &i, MDTAllowedPLMNList: []HexBytes{{0x21, 0xf3, 0x54}},
		},
	}
}

func TestTraceModeFullWire(t *testing.T) {
	// §17.7.2: all public fields; content 296 octets (depth 60, interfaces 42, events 32, MDT 114 including area content 50).
	wire := mustHex(t, "30820128800800010121436587f98102aabb8201078304912143f58503112233a63c8001008101028201008301028401018501028601008701028801008901028a01008b01028c01008d01028e01018f0102900100910102920100930102870306ffc0a82a800306ffc0810205e0820305ffe0830205e0840204f085020780860203f8870203f8880200ff890205e0a920800203f881020780820204f0830206c084020780850202fc860205e0870205e08a0504c0000201ab720a01033032a009040721f35400010002a109040721f35400000003a208040621f354000405a307040521f3540006a407040521f35400070404010203048001010a011781010702016182012283010784010586010e8701068801008901018a01028b01038c0200a58d01e7ae05040321f354")
	a := traceFullArg()
	got, err := ParseActivateTraceModeArg(wire)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "activate full", a, got)
	for _, v := range []*ActivateTraceModeArg{a, got} {
		b, err := v.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b, wire) {
			t.Fatalf("got %x, want %x", b, wire)
		}
	}
	// [0] IMSI (8), [1] reference (2), [3] reference2 (3): content 19 octets.
	dwire := mustHex(t, "3013800800010121436587f98102aabb8303112233")
	d := &DeactivateTraceModeArg{IMSI: a.IMSI, TraceReference: a.TraceReference, TraceReference2: a.TraceReference2}
	dg, err := ParseDeactivateTraceModeArg(dwire)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "deactivate full", d, dg)
	b, err := d.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, dwire) {
		t.Fatalf("%x", b)
	}
}

func traceTLV(tag byte, b []byte) []byte {
	n := len(b)
	if n < 128 {
		return append([]byte{tag, byte(n)}, b...)
	}
	if n < 256 {
		return append([]byte{tag, 0x81, byte(n)}, b...)
	}
	return append([]byte{tag, 0x82, byte(n >> 8), byte(n)}, b...)
}

// ASN.1 constraints are enforced by the codec on both entry points (§17.7.2).
func TestTraceModeConstraints(t *testing.T) {
	type check struct {
		name, path, constraint string
		values                 []int
		set                    func(*ActivateTraceModeArg, int)
		raw                    func(int) []byte
	}
	cases := []check{
		{"TraceReference", "traceReference", "SIZE (1..2)", []int{0, 1, 2, 3}, func(a *ActivateTraceModeArg, n int) { a.TraceReference = make(HexBytes, n) }, func(n int) []byte { return append(traceTLV(0x81, make([]byte, n)), []byte{0x82, 1, 0}...) }},
		{"TraceReference2", "traceReference2", "SIZE (3)", []int{0, 2, 3, 4}, func(a *ActivateTraceModeArg, n int) { a.TraceReference2 = make(HexBytes, n) }, func(n int) []byte { return append([]byte{0x81, 1, 1, 0x82, 1, 0}, traceTLV(0x85, make([]byte, n))...) }},
		{"TraceType", "traceType", "(0..255)", []int{-1, 0, 255, 256}, func(a *ActivateTraceModeArg, n int) { a.TraceType = n }, func(n int) []byte {
			v := []byte{byte(n)}
			if n >= 128 {
				v = []byte{byte(n >> 8), byte(n)}
			}
			return append([]byte{0x81, 1, 1}, traceTLV(0x82, v)...)
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for i, n := range c.values {
				a := traceArg()
				c.set(a, n)
				_, enc := a.Marshal()
				_, dec := ParseActivateTraceModeArg(traceTLV(0x30, c.raw(n)))
				valid := i == 1 || i == 2
				if c.name == "TraceReference2" {
					valid = n == 3
					if n == 0 { // an empty optional octet string is absent
						if enc != nil {
							t.Fatal(enc)
						}
						wantValidationError(t, dec, c.path, c.constraint)
						continue
					}
				}
				if valid {
					if enc != nil || dec != nil {
						t.Fatalf("%d: encode %v decode %v", n, enc, dec)
					}
				} else {
					wantValidationError(t, enc, c.path, c.constraint)
					wantValidationError(t, dec, c.path, c.constraint)
				}
			}
		})
	}
	// Each octet parameter has its own SIZE constraint; nil and empty are absent.
	for _, c := range []struct {
		field, path string
		size        int
		tag         byte
	}{
		{"ListOfMeasurements", "listOfMeasurements", 4, 4}, {"ReportingTrigger", "reportingTrigger", 1, 0x80},
		{"PositioningMethod", "positioningMethod", 1, 0x8a}, {"MeasurementQuantity", "measurementQuantity", 1, 0x8b},
	} {
		t.Run(c.field, func(t *testing.T) {
			for _, n := range []int{0, c.size, c.size + 1} {
				a := traceArg()
				a.MDTConfiguration = &MDTConfiguration{}
				reflect.ValueOf(a.MDTConfiguration).Elem().FieldByName(c.field).Set(reflect.ValueOf(make(HexBytes, n)))
				_, enc := a.Marshal()
				nested := append([]byte{0x0a, 1, 0}, traceTLV(c.tag, make([]byte, n))...)
				_, dec := ParseActivateTraceModeArg(traceTLV(0x30, append(mustHex(t, "810101820100"), traceTLV(0xab, nested)...)))
				if n == 0 {
					if enc != nil {
						t.Fatal(enc)
					}
					wantValidationError(t, dec, "mdt-Configuration."+c.path, fmt.Sprintf("SIZE (%d)", c.size))
					continue
				}
				if n == c.size {
					if enc != nil || dec != nil {
						t.Fatalf("encode %v decode %v", enc, dec)
					}
				} else {
					wantValidationError(t, enc, "mdt-Configuration."+c.path, fmt.Sprintf("SIZE (%d)", c.size))
					wantValidationError(t, dec, "mdt-Configuration."+c.path, fmt.Sprintf("SIZE (%d)", c.size))
				}
			}
		})
	}
	// All integer ranges, including signed event thresholds.
	for _, c := range []struct {
		field, path string
		lo, hi      int
		tag         byte
	}{
		{"EventThresholdRSRP", "eventThresholdRSRP", 0, 97, 2}, {"EventThresholdRSRQ", "eventThresholdRSRQ", 0, 34, 0x82},
		{"EventThreshold1F", "eventThreshold1F", -120, 165, 0x8c}, {"EventThreshold1I", "eventThreshold1I", -120, -25, 0x8d},
	} {
		t.Run(c.field, func(t *testing.T) {
			for _, n := range []int{c.lo - 1, c.lo, c.hi, c.hi + 1} {
				a := traceArg()
				a.MDTConfiguration = &MDTConfiguration{}
				v := n
				reflect.ValueOf(a.MDTConfiguration).Elem().FieldByName(c.field).Set(reflect.ValueOf(&v))
				_, enc := a.Marshal()
				value := []byte{byte(n)}
				if n > 127 {
					value = []byte{0, byte(n)}
				}
				nested := append([]byte{0x0a, 1, 0}, traceTLV(c.tag, value)...)
				_, dec := ParseActivateTraceModeArg(traceTLV(0x30, append(mustHex(t, "810101820100"), traceTLV(0xab, nested)...)))
				if n >= c.lo && n <= c.hi {
					if enc != nil || dec != nil {
						t.Fatalf("%d: encode %v decode %v", n, enc, dec)
					}
				} else {
					wantValidationError(t, enc, "mdt-Configuration."+c.path, fmt.Sprintf("(%d..%d)", c.lo, c.hi))
					wantValidationError(t, dec, "mdt-Configuration."+c.path, fmt.Sprintf("(%d..%d)", c.lo, c.hi))
				}
			}
		})
	}
	// Non-extensible enumerations belong to the codec, including every optional instance.
	for _, c := range []struct {
		field, path string
		max         int
		tag         byte
	}{
		{"JobType", "jobType", 3, 0x0a}, {"ReportInterval", "reportInterval", 27, 0x0a}, {"ReportAmount", "reportAmount", 7, 0x81},
		{"LoggingInterval", "loggingInterval", 7, 0x83}, {"LoggingDuration", "loggingDuration", 5, 0x84},
		{"MeasurementPeriodUMTS", "measurementPeriodUMTS", 14, 0x86}, {"MeasurementPeriodLTE", "measurementPeriodLTE", 6, 0x87},
		{"CollectionPeriodRRMUMTS", "collectionPeriodRRM-UMTS", 14, 0x88}, {"CollectionPeriodRRMLTE", "collectionPeriodRRM-LTE", 6, 0x89},
	} {
		t.Run(c.field, func(t *testing.T) {
			for _, n := range []int{-1, 0, c.max, c.max + 1} {
				a := traceArg()
				a.MDTConfiguration = &MDTConfiguration{}
				f := reflect.ValueOf(a.MDTConfiguration).Elem().FieldByName(c.field)
				if f.Kind() == reflect.Pointer {
					v := reflect.New(f.Type().Elem())
					v.Elem().SetInt(int64(n))
					f.Set(v)
				} else {
					f.SetInt(int64(n))
				}
				_, enc := a.Marshal()
				nested := traceTLV(c.tag, []byte{byte(n)})
				if c.field != "JobType" {
					nested = append([]byte{0x0a, 1, 0}, nested...)
				}
				_, dec := ParseActivateTraceModeArg(traceTLV(0x30, append(mustHex(t, "810101820100"), traceTLV(0xab, nested)...)))
				if n >= 0 && n <= c.max {
					if enc != nil || dec != nil {
						t.Fatalf("%d: encode %v decode %v", n, enc, dec)
					}
				} else {
					var ve *ber.ValidationError
					if !errors.As(enc, &ve) || ve.Path != "mdt-Configuration."+c.path {
						t.Fatalf("encode %v", enc)
					}
					if !matchesConstraint(dec, ve.Path, ve.Constraint) {
						t.Fatalf("decode %v, want %s %s", dec, ve.Path, ve.Constraint)
					}
				}
			}
		})
	}
}

func TestTraceModeAreaLists(t *testing.T) {
	// All list limits and identity octet limits from §§17.7.2,17.7.8.
	for _, c := range []struct {
		field, path        string
		minSize, size, max int
		tag                byte
	}{
		{"CgiList", "cgi-List", 5, 7, 32, 0xa0}, {"EUtranCgiList", "e-utran-cgi-List", 7, 7, 32, 0xa1},
		{"RoutingAreaIdList", "routingAreaId-List", 6, 6, 8, 0xa2}, {"LocationAreaIdList", "locationAreaId-List", 5, 5, 8, 0xa3},
		{"TrackingAreaIdList", "trackingAreaId-List", 5, 5, 8, 0xa4}, {"MDTAllowedPLMNList", "mdt-Allowed-PLMN-List", 3, 3, 16, 0xae},
	} {
		t.Run(c.field, func(t *testing.T) {
			makeArg := func(list []HexBytes) *ActivateTraceModeArg {
				a := traceArg()
				a.MDTConfiguration = &MDTConfiguration{}
				if c.field == "MDTAllowedPLMNList" {
					a.MDTConfiguration.MDTAllowedPLMNList = list
				} else {
					a.MDTConfiguration.AreaScope = &AreaScope{}
					reflect.ValueOf(a.MDTConfiguration.AreaScope).Elem().FieldByName(c.field).Set(reflect.ValueOf(list))
				}
				return a
			}
			path := "mdt-Configuration."
			if c.field != "MDTAllowedPLMNList" {
				path += "areaScope."
			}
			path += c.path
			raw := func(list []HexBytes) []byte {
				var b []byte
				for _, id := range list {
					b = append(b, traceTLV(4, id)...)
				}
				b = traceTLV(c.tag, b)
				if c.field != "MDTAllowedPLMNList" {
					b = traceTLV(0x30, b)
				}
				return traceTLV(0x30, append(mustHex(t, "810101820100"), traceTLV(0xab, append([]byte{0x0a, 1, 0}, b...))...))
			}
			for _, n := range []int{0, 1, c.max, c.max + 1} {
				list := make([]HexBytes, n)
				for i := range list {
					list[i] = make(HexBytes, c.size)
					copy(list[i], []byte{0x21, 0xf3, 0x54})
				}
				_, enc := makeArg(list).Marshal()
				_, dec := ParseActivateTraceModeArg(raw(list))
				if n > 0 && n <= c.max {
					if enc != nil || dec != nil {
						t.Fatalf("%d encode %v decode %v", n, enc, dec)
					}
				} else {
					wantValidationError(t, enc, path, fmt.Sprintf("SIZE (1..%d)", c.max))
					wantValidationError(t, dec, path, fmt.Sprintf("SIZE (1..%d)", c.max))
				}
			}
			for _, size := range []int{c.minSize - 1, c.minSize, c.size, c.size + 1} {
				id := make(HexBytes, size)
				copy(id, []byte{0x21, 0xf3, 0x54})
				list := []HexBytes{id}
				_, enc := makeArg(list).Marshal()
				_, dec := ParseActivateTraceModeArg(raw(list))
				if size >= c.minSize && size <= c.size {
					if enc != nil || dec != nil {
						t.Fatalf("size %d encode %v decode %v", size, enc, dec)
					}
				} else {
					constraint := fmt.Sprintf("SIZE (%d)", c.size)
					if c.minSize != c.size {
						constraint = fmt.Sprintf("SIZE (%d..%d)", c.minSize, c.size)
					}
					wantValidationError(t, enc, path+".element[0]", constraint)
					wantValidationError(t, dec, path+".element[0]", constraint)
				}
			}
			// PLMN decimal coding is sender-only, as for existing identities of these types.
			id := make(HexBytes, c.size)
			copy(id, []byte{0x2a, 0xf3, 0x54})
			list := []HexBytes{id}
			_, err := makeArg(list).Marshal()
			if !errors.Is(err, identity.ErrPLMNIdentityDigits) {
				t.Fatal(err)
			}
			got, err := ParseActivateTraceModeArg(raw(list))
			if err != nil {
				t.Fatal(err)
			}
			var received []HexBytes
			if c.field == "MDTAllowedPLMNList" {
				received = got.MDTConfiguration.MDTAllowedPLMNList
			} else {
				received = reflect.ValueOf(got.MDTConfiguration.AreaScope).Elem().FieldByName(c.field).Interface().([]HexBytes)
			}
			wantEqual(t, "retained PLMN", list, received)
			_, err = got.Marshal()
			if !errors.Is(err, identity.ErrPLMNIdentityDigits) {
				t.Fatal(err)
			}
		})
	}
}

func TestTraceModeIdentityAddressAndGSN(t *testing.T) {
	for _, deactivate := range []bool{false, true} {
		marshal := func(imsi string) ([]byte, error) {
			if deactivate {
				return (&DeactivateTraceModeArg{IMSI: imsi, TraceReference: HexBytes{1}}).Marshal()
			}
			a := traceArg()
			a.IMSI = imsi
			return a.Marshal()
		}
		parse := func(raw []byte) (string, error) {
			b := traceTLV(0x30, append(traceTLV(0x80, raw), []byte{0x81, 1, 1}...))
			if deactivate {
				v, e := ParseDeactivateTraceModeArg(b)
				if e != nil {
					return "", e
				}
				return v.IMSI, nil
			}
			b = traceTLV(0x30, append(append(traceTLV(0x80, raw), []byte{0x81, 1, 1}...), []byte{0x82, 1, 0}...))
			v, e := ParseActivateTraceModeArg(b)
			if e != nil {
				return "", e
			}
			return v.IMSI, nil
		}
		for _, id := range []string{"123456", "001010123456789"} {
			b, err := marshal(id)
			if err != nil {
				t.Fatal(err)
			}
			if deactivate {
				v, err := ParseDeactivateTraceModeArg(b)
				if err != nil || v.IMSI != id {
					t.Fatalf("%v %v", v, err)
				}
			} else {
				v, err := ParseActivateTraceModeArg(b)
				if err != nil || v.IMSI != id {
					t.Fatalf("%v %v", v, err)
				}
			}
		}
		for _, id := range []string{"1234", "1234567890123456", "12345a"} {
			if _, err := marshal(id); err == nil {
				t.Fatalf("IMSI %q accepted", id)
			}
		}
		for _, raw := range [][]byte{{0xff, 0xff, 0xff}, {0xaa, 0x21, 0x43}, {0x21, 0x43}} {
			if _, err := parse(raw); err == nil {
				t.Fatalf("IMSI %x accepted", raw)
			}
		}
		// An empty TraceReference2 is absent; the wire form reaches SIZE (3).
		if deactivate {
			for _, n := range []int{0, 2, 3, 4} {
				d := &DeactivateTraceModeArg{TraceReference: HexBytes{1}, TraceReference2: make(HexBytes, n)}
				_, enc := d.Marshal()
				_, dec := ParseDeactivateTraceModeArg(traceTLV(0x30, append([]byte{0x81, 1, 1}, traceTLV(0x83, make([]byte, n))...)))
				if n == 0 {
					if enc != nil {
						t.Fatal(enc)
					}
					wantValidationError(t, dec, "traceReference2", "SIZE (3)")
					continue
				}
				if n == 3 {
					if enc != nil || dec != nil {
						t.Fatalf("encode %v decode %v", enc, dec)
					}
				} else {
					wantValidationError(t, enc, "traceReference2", "SIZE (3)")
					wantValidationError(t, dec, "traceReference2", "SIZE (3)")
				}
			}
			d := &DeactivateTraceModeArg{}
			_, err := d.Marshal()
			wantValidationError(t, err, "traceReference", "SIZE (1..2)")
			// [1] zero-length mandatory reference.
			_, err = ParseDeactivateTraceModeArg(mustHex(t, "30028100"))
			wantValidationError(t, err, "traceReference", "SIZE (1..2)")
		}
	}
	for _, c := range []struct {
		digits       string
		nature, plan uint8
		want         error
	}{
		{"1", 0x50, 1, ErrAddressNatureInvalid}, {"1", 0x10, 2, ErrAddressPlanInvalid},
	} {
		a := traceArg()
		a.OMCID = c.digits
		a.OMCIDNature = c.nature
		a.OMCIDPlan = c.plan
		_, err := a.Marshal()
		if !errors.Is(err, c.want) {
			t.Fatal(err)
		}
	}
	a := traceArg()
	a.OMCID = "invalid!"
	if _, err := a.Marshal(); err == nil {
		t.Fatal("invalid OMCID accepted")
	}
	a.OMCID = strings.Repeat("1", 38)
	_, err := a.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	a.OMCID = strings.Repeat("1", 39)
	_, err = a.Marshal()
	wantValidationError(t, err, "omc-Id", "SIZE (1..20)")
	// §17.7.2 [3] AddressString with malformed TBCD, or zero digits.
	for _, raw := range [][]byte{{0x91}, {0x91, 0xff}, {0x91, 0x1f}} {
		_, err := ParseActivateTraceModeArg(traceTLV(0x30, append(mustHex(t, "810101820100"), traceTLV(0x83, raw)...)))
		if err == nil {
			t.Fatalf("OMCID %x accepted", raw)
		}
		if len(raw) == 1 || raw[1] == 0xff {
			if !errors.Is(err, ErrActivateTraceModeArgOMCIDDecodedEmpty) {
				t.Fatal(err)
			}
		}
	}
	a = traceArg()
	a.TraceCollectionEntity = "invalid IP"
	if _, err := a.Marshal(); err == nil {
		t.Fatal("invalid GSN accepted")
	}
	// [10] IPv4 header length 5 with only four address octets; valid ASN.1 SIZE (5).
	_, err = ParseActivateTraceModeArg(mustHex(t, "300d8101018201008a0505c0000201"))
	if err == nil {
		t.Fatal("invalid GSN discriminator accepted")
	}
	// [10] IPv4 address with four octets rather than the mandatory five.
	_, err = ParseActivateTraceModeArg(mustHex(t, "300c8101018201008a0404c00002"))
	wantValidationError(t, err, "traceCollectionEntity", "SIZE (5..17)")
	a = traceArg()
	a.TraceCollectionEntity = "2001:db8::1"
	b, err := a.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	g, err := ParseActivateTraceModeArg(b)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "IPv6", a, g)
}

func TestTraceModeReceiverSeeds(t *testing.T) {
	// [6] base MSC medium: map to minimum (§17.7.2).
	a, err := ParseActivateTraceModeArg(mustHex(t, "300b810101820100a603800101"))
	if err != nil {
		t.Fatal(err)
	}
	if a.TraceDepthList.MscSTraceDepth == nil || *a.TraceDepthList.MscSTraceDepth != TraceDepthMinimum {
		t.Fatal(a.TraceDepthList)
	}
	// [6] extension MSC medium: map to minimum and retain the absent base.
	a, err = ParseActivateTraceModeArg(mustHex(t, "300b810101820100a6038a0101"))
	if err != nil {
		t.Fatal(err)
	}
	if a.TraceDepthList.MscSTraceDepthExtension == nil || *a.TraceDepthList.MscSTraceDepthExtension != TraceDepthExtensionMinimumWithoutVendorSpecificExtension || a.TraceDepthList.MscSTraceDepth != nil {
		t.Fatal(a.TraceDepthList)
	}
	// [7] named MSC bit 0 plus unlisted bit 10; [8]/[9] present empty sequences.
	// MDT [11] has its mandatory JobType=0. Content: 6+5+2+2+5=20 octets.
	a, err = ParseActivateTraceModeArg(mustHex(t, "30148101018201008703008020a800a900ab030a0100"))
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "NE discard", &TraceNETypeList{MscS: true}, a.TraceNETypeList)
	wantEqual(t, "empty interfaces", &TraceInterfaceList{}, a.TraceInterfaceList)
	wantEqual(t, "empty events", &TraceEventList{}, a.TraceEventList)
	wantEqual(t, "minimal MDT", &MDTConfiguration{}, a.MDTConfiguration)
	// All public empty optional sequences remain present after canonical encoding.
	b, err := a.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	again, err := ParseActivateTraceModeArg(b)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "empty sequences", a, again)
}
