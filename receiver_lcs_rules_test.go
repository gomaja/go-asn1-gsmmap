package gsmmap

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/gomaja/go-asn1/runtime/ber"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func lcsRulePSLArg(t *testing.T) *gsm_map.ProvideSubscriberLocationArg {
	t.Helper()
	w, err := convertProvideSubscriberLocationArgToWire(&ProvideSubscriberLocationArg{
		LocationType: LocationType{LocationEstimateType: LocationEstimateCurrentLocation},
		MlcNumber:    "12345", MlcNumberNature: 0x10, MlcNumberPlan: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

// TS 29.002 V19.1.0 §17.7.13: beingInsideArea always means oneTimeEvent.
func TestReceiverT081BeingInsideArea(t *testing.T) {
	for _, inside := range []bool{false, true} {
		for _, occurrence := range []*gsm_map.OccurrenceInfo{nil, new(gsm_map.OccurrenceInfoOneTimeEvent), new(gsm_map.OccurrenceInfoMultipleTimeEvent), new(gsm_map.OccurrenceInfo(99))} {
			w := lcsRulePSLArg(t)
			w.LocationType.DeferredLocationEventType = new(convertDeferredLocationEventTypeToBitString(&DeferredLocationEventType{MsAvailable: true, BeingInsideArea: inside}))
			w.AreaEventInfo = &gsm_map.AreaEventInfo{
				AreaDefinition: gsm_map.AreaDefinition{AreaList: &gsm_map.AreaList{Values: []gsm_map.Area{{AreaType: gsm_map.AreaTypeCountryCode, AreaIdentification: gsm_map.AreaIdentification{0x21, 0xf3}}}}},
				OccurrenceInfo: occurrence,
			}
			got, err := ParseProvideSubscriberLocation(strictBER(t, w))
			if err != nil {
				t.Fatal(err)
			}
			// The converter must allocate the effective occurrence independently.
			converted, err := convertWireToProvideSubscriberLocationArg(w)
			if err != nil {
				t.Fatal(err)
			}
			if inside && converted.AreaEventInfo.OccurrenceInfo == w.AreaEventInfo.OccurrenceInfo {
				t.Fatal("effective occurrence aliases source")
			}
			if inside {
				if got.AreaEventInfo.OccurrenceInfo == nil || *got.AreaEventInfo.OccurrenceInfo != OccurrenceOneTimeEvent {
					t.Errorf("inside=%v occurrence=%v: %+v", inside, occurrence, got.AreaEventInfo)
				}
			} else if (occurrence == nil) != (got.AreaEventInfo.OccurrenceInfo == nil) || occurrence != nil && *occurrence != *got.AreaEventInfo.OccurrenceInfo {
				t.Errorf("occurrence changed without beingInsideArea")
			}
			if occurrence != nil && *occurrence == 99 && *w.AreaEventInfo.OccurrenceInfo != 99 {
				t.Fatal("source occurrence mutated")
			}
		}
		w := lcsRulePSLArg(t)
		w.LocationType.DeferredLocationEventType = new(convertDeferredLocationEventTypeToBitString(&DeferredLocationEventType{BeingInsideArea: inside}))
		got, err := ParseProvideSubscriberLocation(strictBER(t, w))
		if err != nil || got.AreaEventInfo != nil {
			t.Fatalf("absent area: got=%+v err=%v", got, err)
		}
	}
}

// TS 29.002 V19.1.0 §17.7.13: accuracy bit 8 and the periodic product cap constrain senders.
func TestReceiverT084AccuracyPreserved(t *testing.T) {
	for _, v := range []byte{0, 0x7f, 0x80, 0xff} {
		for _, vertical := range []bool{false, true} {
			w := lcsRulePSLArg(t)
			w.LcsQoS = &gsm_map.LCSQoS{}
			if vertical {
				w.LcsQoS.VerticalAccuracy = new(gsm_map.VerticalAccuracy{v})
			} else {
				w.LcsQoS.HorizontalAccuracy = new(gsm_map.HorizontalAccuracy{v})
			}
			got, err := ParseProvideSubscriberLocation(strictBER(t, w))
			if err != nil {
				t.Fatalf("vertical=%v code=%x: %v", vertical, v, err)
			}
			accuracy := got.LcsQoS.HorizontalAccuracy
			wantErr := ErrHorizontalAccuracyReservedBit
			if vertical {
				accuracy = got.LcsQoS.VerticalAccuracy
				wantErr = ErrVerticalAccuracyReservedBit
			}
			if !bytes.Equal(accuracy, []byte{v}) {
				t.Errorf("accuracy changed: %x", accuracy)
			}
			_, err = got.Marshal()
			if v&0x80 != 0 {
				if !errors.Is(err, wantErr) {
					t.Errorf("sender: want %v, got %v", wantErr, err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestReceiverT085PeriodicProductPreserved(t *testing.T) {
	for _, amount := range []int64{1, 4000, PeriodicLDRProductMax} {
		interval := int64(1)
		if amount == 4000 {
			interval = 4000
		}
		periodic := &gsm_map.PeriodicLDRInfo{ReportingAmount: amount, ReportingInterval: interval}
		w := lcsRulePSLArg(t)
		w.PeriodicLDRInfo = periodic
		got, err := ParseProvideSubscriberLocation(strictBER(t, w))
		if err != nil {
			t.Fatalf("PSL: %v", err)
		}
		if got.PeriodicLDRInfo.ReportingAmount != amount || got.PeriodicLDRInfo.ReportingInterval != interval {
			t.Fatal("PSL operands changed")
		}
		_, err = got.Marshal()
		if amount*interval > PeriodicLDRProductMax {
			if !errors.Is(err, ErrPeriodicLDRProductExceeded) {
				t.Errorf("PSL sender: %v", err)
			}
		} else if err != nil {
			t.Fatal(err)
		}
		slr, err := convertSubscriberLocationReportArgToWire(minimalSLRArg())
		if err != nil {
			t.Fatal(err)
		}
		slr.PeriodicLDRInfo = periodic
		report, err := ParseSubscriberLocationReport(strictBER(t, slr))
		if err != nil {
			t.Fatalf("SLR: %v", err)
		}
		if report.PeriodicLDRInfo.ReportingAmount != amount || report.PeriodicLDRInfo.ReportingInterval != interval {
			t.Fatal("SLR operands changed")
		}
		_, err = report.Marshal()
		if amount*interval > PeriodicLDRProductMax {
			if !errors.Is(err, ErrPeriodicLDRProductExceeded) {
				t.Errorf("SLR sender: %v", err)
			}
		} else if err != nil {
			t.Fatal(err)
		}
	}
}

func lcsRuleParseGeography(t *testing.T, slr bool, ext, add, velocity []byte, deferred bool) ([]byte, []byte, error) {
	t.Helper()
	if slr {
		w, err := convertSubscriberLocationReportArgToWire(minimalSLRArg())
		if err != nil {
			t.Fatal(err)
		}
		if ext != nil {
			w.LocationEstimate = new(ext)
		}
		if add != nil {
			w.AddLocationEstimate = new(add)
		}
		if velocity != nil {
			w.VelocityEstimate = new(velocity)
		}
		got, err := ParseSubscriberLocationReport(strictBER(t, w))
		if err != nil {
			return nil, nil, err
		}
		if !bytes.Equal(got.VelocityEstimate, velocity) {
			t.Errorf("velocity changed: %x", got.VelocityEstimate)
		}
		return got.LocationEstimate, got.AddLocationEstimate, nil
	}
	w := &gsm_map.ProvideSubscriberLocationRes{LocationEstimate: ext}
	if add != nil {
		w.AddLocationEstimate = new(add)
	}
	if velocity != nil {
		w.VelocityEstimate = new(velocity)
	}
	w.DeferredmtLrResponseIndicator = boolToNullPtr(deferred)
	got, err := ParseProvideSubscriberLocationRes(strictBER(t, w))
	if err != nil {
		return nil, nil, err
	}
	if got.DeferredmtLrResponseIndicator != deferred || !bytes.Equal(got.VelocityEstimate, velocity) {
		t.Fatal("unrelated PSL fields changed")
	}
	return got.LocationEstimate, got.AddLocationEstimate, nil
}

// TS 29.002 V19.1.0 §17.7.13 Ext/Add exception handling, after deferred-result ignore.
func TestReceiverT086T090T094GeographyPrecedence(t *testing.T) {
	valid := []byte{0, 0, 0, 0, 0, 0, 0}
	bad := []byte{0xf0, 0}
	for _, slr := range []bool{false, true} {
		// Each valid Ext shape independently suppresses semantic validation of Add.
		for shape, length := range map[byte]int{0: 7, 1: 8, 3: 11, 9: 14, 10: 13} {
			v := make([]byte, length)
			v[0] = shape << 4
			for _, add := range [][]byte{valid, bad, {0xff}} {
				ext, gotAdd, err := lcsRuleParseGeography(t, slr, v, add, nil, false)
				if err != nil || !bytes.Equal(ext, v) || gotAdd != nil {
					t.Fatalf("Ext shape%d precedence slr=%v: ext=%x add=%x err=%v", shape, slr, ext, gotAdd, err)
				}
			}
		}
		for _, add := range [][]byte{valid, bad, {0xff}} {
			ext, gotAdd, err := lcsRuleParseGeography(t, slr, valid, add, nil, false)
			if err != nil || !bytes.Equal(ext, valid) || gotAdd != nil {
				t.Fatalf("valid Ext precedence slr=%v: ext=%x add=%x err=%v", slr, ext, gotAdd, err)
			}
		}
		for dummy := range 256 {
			ext, add, err := lcsRuleParseGeography(t, slr, []byte{byte(dummy)}, valid, nil, false)
			if err != nil || ext != nil || !bytes.Equal(add, valid) {
				t.Fatalf("dummy Ext slr=%v byte=%x: ext=%x add=%x err=%v", slr, dummy, ext, add, err)
			}
		}
		_, _, err := lcsRuleParseGeography(t, slr, []byte{0}, bad, nil, false)
		if !errors.Is(err, ErrAddGeographicalInformationInvalidData) {
			t.Errorf("dummy with invalid Add: %v", err)
		}
		_, _, err = lcsRuleParseGeography(t, slr, bad, valid, nil, false)
		if !errors.Is(err, ErrExtGeographicalInformationInvalidData) {
			t.Errorf("invalid Ext with valid Add: %v", err)
		}
	}
	for _, ext := range [][]byte{valid, bad, {0xff}} {
		gotExt, gotAdd, err := lcsRuleParseGeography(t, false, ext, valid, nil, true)
		if err != nil || gotExt != nil || !bytes.Equal(gotAdd, valid) {
			t.Fatalf("deferred ignore: ext=%x add=%x err=%v", gotExt, gotAdd, err)
		}
		gotExt, _, err = lcsRuleParseGeography(t, false, ext, nil, nil, true)
		if err != nil || gotExt != nil {
			t.Fatalf("deferred without Add: ext=%x err=%v", gotExt, err)
		}
	}
	_, _, err := lcsRuleParseGeography(t, false, valid, bad, nil, true)
	if !errors.Is(err, ErrAddGeographicalInformationInvalidData) {
		t.Errorf("ignored Ext must not suppress invalid Add: %v", err)
	}
	ext, add, err := lcsRuleParseGeography(t, true, nil, valid, nil, false)
	if err != nil || ext != nil || !bytes.Equal(add, valid) {
		t.Fatalf("absent SLR Ext: %v", err)
	}
}

// TS 29.002 V19.1.0 §17.7.13: independently exercise each ignore/discard rule.
func TestReceiverGeographyDiscardRules(t *testing.T) {
	valid := []byte{0, 0, 0, 0, 0, 0, 0}
	for _, tc := range []struct {
		name             string
		ext, add         []byte
		deferred         bool
		wantExt, wantAdd []byte
	}{
		{"T086", []byte{0xf0, 0}, nil, true, nil, nil},
		{"T090", []byte{0xff}, valid, false, nil, valid},
		{"T094", valid, []byte{0xff}, false, valid, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ext, add, err := lcsRuleParseGeography(t, false, tc.ext, tc.add, nil, tc.deferred)
			if err != nil || !bytes.Equal(ext, tc.wantExt) || !bytes.Equal(add, tc.wantAdd) {
				t.Fatalf("ext=%x add=%x err=%v", ext, add, err)
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.13; TS 23.032 V19.0.0 §§6.3a,6.5,6.7,7.2,7.3 and figures 5..10.
func TestReceiverT088T095GeographicalCoding(t *testing.T) {
	lengths := map[byte]int{0: 7, 1: 8, 3: 11, 5: 19, 8: 9, 9: 14, 10: 13, 11: 13, 12: 18, 13: 13, 14: 18}
	for _, slr := range []bool{false, true} {
		for shape := byte(0); shape < 16; shape++ {
			length, defined := lengths[shape]
			if !defined {
				length = 7
			}
			v := make([]byte, length)
			v[0] = shape << 4
			if shape == 5 {
				v[0] |= 3
			}
			extAllowed := shape == 0 || shape == 1 || shape == 3 || shape == 9 || shape == 10
			for _, asAdd := range []bool{false, true} {
				wantErr := ErrExtGeographicalInformationInvalidData
				valid := extAllowed
				if asAdd {
					wantErr = ErrAddGeographicalInformationInvalidData
					valid = defined
				}
				parse := func(d []byte) error {
					if asAdd {
						_, _, err := lcsRuleParseGeography(t, slr, []byte{0}, d, nil, false)
						return err
					}
					_, _, err := lcsRuleParseGeography(t, slr, d, nil, nil, false)
					return err
				}
				t.Run(fmt.Sprintf("slr=%v/shape=%d/add=%v", slr, shape, asAdd), func(t *testing.T) {
					err := parse(v)
					if valid {
						if err != nil {
							t.Fatal(err)
						}
					} else if !errors.Is(err, wantErr) {
						t.Fatalf("want %v, got %v", wantErr, err)
					}
					if !valid {
						return
					}
					for _, d := range [][]byte{v[:len(v)-1], append(bytes.Clone(v), 0)} {
						if err := parse(d); !errors.Is(err, wantErr) {
							t.Errorf("wrong length %d: %v", len(d), err)
						}
					}
					if shape != 5 {
						d := bytes.Clone(v)
						d[0] |= 0xf
						if err := parse(d); err != nil {
							t.Errorf("unspecified header spare bits: %v", err)
						}
					}
					var zeroBits []int
					var angles []int
					switch shape {
					case 1:
						zeroBits = []int{7}
					case 3:
						zeroBits = []int{7, 8, 10}
						angles = []int{9}
					case 9:
						zeroBits = []int{9, 10, 12, 13}
						angles = []int{11}
					case 10:
						zeroBits = []int{9, 12}
						angles = []int{10, 11}
					case 11:
						zeroBits = []int{12}
						angles = []int{11}
					case 12:
						zeroBits = []int{9, 15, 17}
						angles = []int{14}
					case 13:
						angles = []int{11}
					case 14:
						zeroBits = []int{9}
						angles = []int{14}
					}
					for _, i := range zeroBits {
						d := bytes.Clone(v)
						d[i] |= 0x80
						if err := parse(d); !errors.Is(err, wantErr) {
							t.Errorf("explicit zero bit octet %d: %v", i+1, err)
						}
					}
					if shape == 12 || shape == 14 {
						d := bytes.Clone(v)
						d[9] |= 0x40
						if err := parse(d); !errors.Is(err, wantErr) {
							t.Errorf("second explicit zero altitude bit: %v", err)
						}
					}
					for _, i := range angles {
						for _, angle := range []byte{179, 180, 255} {
							d := bytes.Clone(v)
							d[i] = angle
							err := parse(d)
							if angle == 179 {
								if err != nil {
									t.Errorf("angle179: %v", err)
								}
							} else if !errors.Is(err, wantErr) {
								t.Errorf("angle%d: %v", angle, err)
							}
						}
					}
					// Confidence 101..127 may mean no information; they do not mandate rejection (§6.5).
					confidences := map[byte][]int{3: {10}, 9: {13}, 10: {12}, 11: {12}, 12: {15, 17}, 13: {12}, 14: {15, 17}}
					for _, i := range confidences[shape] {
						for _, c := range []byte{100, 101, 127} {
							d := bytes.Clone(v)
							d[i] = c
							if err := parse(d); err != nil {
								t.Errorf("confidence%d: %v", c, err)
							}
						}
					}
					if shape == 5 {
						for _, n := range []byte{2, 3, 15} {
							d := make([]byte, 1+6*int(n))
							d[0] = 0x50 | n
							err := parse(d)
							if n >= 3 {
								if err != nil {
									t.Errorf("polygon%d: %v", n, err)
								}
							} else if !errors.Is(err, wantErr) {
								t.Errorf("polygon%d: %v", n, err)
							}
						}
					}
					if shape == 8 || shape == 9 {
						d := bytes.Clone(v)
						d[7] = 0x80
						if err := parse(d); err != nil {
							t.Errorf("altitude direction: %v", err)
						}
					}
					if shape == 12 || shape == 14 {
						for _, altitude := range []int32{-64001, -64000, 1280000, 1280001} {
							d := bytes.Clone(v)
							raw := uint32(altitude) & 0x3fffff
							d[9], d[10], d[11] = byte(raw>>16), byte(raw>>8), byte(raw)
							err := parse(d)
							if altitude >= -64000 && altitude <= 1280000 {
								if err != nil {
									t.Errorf("altitude%d: %v", altitude, err)
								}
							} else if !errors.Is(err, wantErr) {
								t.Errorf("altitude%d: %v", altitude, err)
							}
						}
					}
					if shape >= 11 {
						uncertainties := []int{9, 10}
						if shape == 12 || shape == 14 {
							uncertainties = []int{12, 13, 16}
						}
						for _, i := range uncertainties {
							for code := range 256 {
								d := bytes.Clone(v)
								d[i] = byte(code)
								if err := parse(d); err != nil {
									t.Fatalf("uncertainty octet%d code%d: %v", i+1, code, err)
								}
							}
						}
						for _, coordinate := range []byte{0, 0x7f, 0x80, 0xff} {
							d := bytes.Clone(v)
							for i := 1; i <= 8; i++ {
								d[i] = coordinate
							}
							if err := parse(d); err != nil {
								t.Errorf("coordinate pattern %x: %v", coordinate, err)
							}
						}
					}
					if shape == 13 {
						d := bytes.Clone(v)
						d[12] = 0xff
						if err := parse(d); err != nil {
							t.Errorf("scalable U: %v", err)
						}
					}
					if shape == 14 {
						for flags := range 4 {
							d := bytes.Clone(v)
							d[15] = byte(flags&1) << 7
							d[17] = byte(flags>>1) << 7
							if err := parse(d); err != nil {
								t.Errorf("scalable HU/VU%d: %v", flags, err)
							}
						}
					}
				})
			}
		}
	}
}

// TS 29.002 V19.1.0 §17.7.13; TS 23.032 V19.0.0 §§8.5,8.6: type is the high nibble.
func TestReceiverT092VelocityType(t *testing.T) {
	for _, slr := range []bool{false, true} {
		for typ := byte(0); typ < 16; typ++ {
			length := 4
			if typ < 4 {
				length += int(typ)
			}
			v := make([]byte, length)
			v[0] = typ<<4 | 1
			if typ == 1 || typ == 3 {
				v[0] |= 2
			}
			_, _, err := lcsRuleParseGeography(t, slr, []byte{0, 0, 0, 0, 0, 0, 0}, nil, v, false)
			if typ < 4 {
				if err != nil {
					t.Errorf("type%d slr=%v: %v", typ, slr, err)
				}
			} else if !errors.Is(err, ErrVelocityEstimateInvalidData) {
				t.Errorf("type%d slr=%v: %v", typ, slr, err)
			}
		}
	}
}

// TS 29.002 V19.1.0 §17.7.13: both routing identifiers make the response unsuccessful.
func TestReceiverT100RoutingIdentifiers(t *testing.T) {
	for _, rk := range []bool{false, true} {
		for _, rd := range []bool{false, true} {
			w := &gsm_map.SubscriberLocationReportRes{}
			if rk {
				w.NaESRK = new(gsm_map.ISDNAddressString{0x91, 0x21})
			}
			if rd {
				w.NaESRD = new(gsm_map.ISDNAddressString{0x91, 0x43})
			}
			got, err := ParseSubscriberLocationReportRes(strictBER(t, w))
			if rk && rd {
				if !errors.Is(err, ErrSLRResRoutingIdentifiersReturnError) {
					t.Errorf("both: %v", err)
				}
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			if (got.NaESRK != "") != rk || (got.NaESRD != "") != rd {
				t.Fatalf("routing identity changed: %+v", got)
			}
		}
	}
	w := &gsm_map.SubscriberLocationReportRes{NaESRK: new(gsm_map.ISDNAddressString{0x91}), NaESRD: new(gsm_map.ISDNAddressString{0x91})}
	_, err := ParseSubscriberLocationReportRes(strictBER(t, w))
	if !errors.Is(err, ErrSLRResRoutingIdentifiersReturnError) {
		t.Errorf("presence before digit conversion: %v", err)
	}
}

// TS 29.002 V19.1.0 §17.7.13: ASN.1 constraints still apply to ignored fields.
func TestReceiverLCSCodecConstraints(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*gsm_map.ProvideSubscriberLocationRes)
	}{
		{"ignored Ext empty", func(w *gsm_map.ProvideSubscriberLocationRes) {
			w.LocationEstimate = gsm_map.ExtGeographicalInformation{}
			w.DeferredmtLrResponseIndicator = boolToNullPtr(true)
		}},
		{"ignored Ext oversized", func(w *gsm_map.ProvideSubscriberLocationRes) {
			w.LocationEstimate = make(gsm_map.ExtGeographicalInformation, 21)
			w.DeferredmtLrResponseIndicator = boolToNullPtr(true)
		}},
		{"discarded Add empty", func(w *gsm_map.ProvideSubscriberLocationRes) {
			w.AddLocationEstimate = new(gsm_map.AddGeographicalInformation{})
		}},
		{"discarded Add oversized", func(w *gsm_map.ProvideSubscriberLocationRes) {
			w.AddLocationEstimate = new(make(gsm_map.AddGeographicalInformation, 92))
		}},
		{"velocity short", func(w *gsm_map.ProvideSubscriberLocationRes) {
			w.VelocityEstimate = new(make(gsm_map.VelocityEstimate, 3))
		}},
		{"velocity long", func(w *gsm_map.ProvideSubscriberLocationRes) {
			w.VelocityEstimate = new(make(gsm_map.VelocityEstimate, 8))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &gsm_map.ProvideSubscriberLocationRes{LocationEstimate: gsm_map.ExtGeographicalInformation{0, 0, 0, 0, 0, 0, 0}}
			tc.change(w)
			_, err := ParseProvideSubscriberLocationRes(tolerantBER(t, w))
			var constraint *ber.ConstraintError
			if !errors.As(err, &constraint) {
				t.Fatalf("want codec constraint error, got %v", err)
			}
		})
	}
	for _, amount := range []int64{0, PeriodicLDRProductMax + 1} {
		w := lcsRulePSLArg(t)
		w.PeriodicLDRInfo = &gsm_map.PeriodicLDRInfo{ReportingAmount: amount, ReportingInterval: 1}
		_, err := ParseProvideSubscriberLocation(tolerantBER(t, w))
		var constraint *ber.ConstraintError
		if !errors.As(err, &constraint) {
			t.Fatalf("amount%d: want codec constraint error, got %v", amount, err)
		}
	}
}

// TS 29.002 V19.1.0 §17.7.13: effective geography is preserved or discarded.
func FuzzReceiverGeographicalInformation(f *testing.F) {
	for _, shape := range []byte{0, 1, 3, 5, 8, 9, 10, 11, 12, 13, 14, 15} {
		lengths := map[byte]int{0: 7, 1: 8, 3: 11, 5: 19, 8: 9, 9: 14, 10: 13, 11: 13, 12: 18, 13: 13, 14: 18, 15: 7}
		v := make([]byte, lengths[shape])
		v[0] = shape << 4
		if shape == 5 {
			v[0] |= 3
		}
		f.Add([]byte{0}, v, uint8(0))
		f.Add([]byte{0, 0, 0, 0, 0, 0, 0}, v, uint8(1))
		f.Add(v[:min(len(v), 20)], []byte{0xff}, uint8(2))
	}
	f.Fuzz(func(t *testing.T, ext, add []byte, flags uint8) {
		if len(ext) < 1 || len(ext) > 20 || len(add) < 1 || len(add) > 91 {
			return
		}
		slr, deferred := flags&1 != 0, flags&2 != 0
		gotExt, gotAdd, err := lcsRuleParseGeography(t, slr, ext, add, nil, deferred)
		if err != nil {
			if !errors.Is(err, ErrExtGeographicalInformationInvalidData) && !errors.Is(err, ErrAddGeographicalInformationInvalidData) {
				t.Fatalf("unexpected error: %v", err)
			}
			return
		}
		if !slr && deferred && gotExt != nil {
			t.Fatalf("deferred Ext retained: %x", gotExt)
		}
		if len(ext) == 1 && gotExt != nil {
			t.Fatalf("dummy Ext retained: %x", gotExt)
		}
		if gotExt != nil && (!bytes.Equal(gotExt, ext) || gotAdd != nil) {
			t.Fatal("effective Ext changed or Add retained")
		}
		if gotExt == nil && !bytes.Equal(gotAdd, add) {
			t.Fatal("effective Add changed")
		}
	})
}

// locationEstimate is mandatory. When Parse leaves it empty, beside
// deferredmt-lrResponseIndicator or after discarding a one-octet value
// beside add-LocationEstimate (3GPP TS 29.002 V19.1.0 §17.7.13), Marshal
// sends one octet and the result parses back to the same value.
func TestProvideSubscriberLocationResEmptyLocationEstimateMarshals(t *testing.T) {
	point := ExtGeographicalInformation{0x00, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc}
	add := AddGeographicalInformation{0x10, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0x05} // point with uncertainty circle
	for _, tc := range []struct {
		name string
		in   ProvideSubscriberLocationRes
		ext  byte
	}{
		{"deferred response", ProvideSubscriberLocationRes{LocationEstimate: point, DeferredmtLrResponseIndicator: true}, 0x00},
		{"additional estimate", ProvideSubscriberLocationRes{LocationEstimate: ExtGeographicalInformation{0x50}, AddLocationEstimate: add}, 0x10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.in.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			got, err := ParseProvideSubscriberLocationRes(data)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if len(got.LocationEstimate) != 0 {
				t.Fatalf("LocationEstimate = %x, want empty", got.LocationEstimate)
			}
			again, err := got.Marshal()
			if err != nil {
				t.Fatalf("Marshal of the parsed value: %v", err)
			}
			var w gsm_map.ProvideSubscriberLocationRes
			if err := w.UnmarshalBER(again); err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "locationEstimate", gsm_map.ExtGeographicalInformation{tc.ext}, w.LocationEstimate)
			reparsed, err := ParseProvideSubscriberLocationRes(again)
			if err != nil {
				t.Fatalf("Parse of the re-marshalled value: %v", err)
			}
			wantEqual(t, "result", got, reparsed)
		})
	}
	if _, err := (&ProvideSubscriberLocationRes{}).Marshal(); err == nil {
		t.Error("Marshal with no LocationEstimate: want an error")
	}
}
