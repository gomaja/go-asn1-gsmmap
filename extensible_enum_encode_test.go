package gsmmap

import (
	"errors"
	"fmt"
	"math"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// Used-RAT-Type, UE-SRVCC-Capability, SMSRegisterRequest and
// SM-DeliveryNotIntended are extensible ENUMERATEDs without exception text
// (3GPP TS 29.002 V19.1.0 §17.7.1, §17.7.6). Marshal sends only a listed
// value; Parse keeps any other (§17.1.4), which Marshal then refuses, so
// FuzzParse allows these errors.

// extEnumCase builds a message carrying the value v in one field (msg) and
// the same message's wire form (wire), and reads the decoded field back.
type extEnumCase struct {
	name   string
	listed []int64
	want   error
	msg    func(v int64) marshaler
	wire   func(t *testing.T, v int64) berMarshaler
	parse  func([]byte) (marshaler, error)
	field  func(marshaler) int64
}

func extEnumCases() []extEnumCase {
	usedRAT := func(v int64) *UsedRatType { r := UsedRatType(v); return &r }
	gprs := func(v int64, set func(u *UpdateGprsLocation, v int64)) *UpdateGprsLocation {
		u := &UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "31612345678", SGSNAddress: "192.168.31.1"}
		set(u, v)
		return u
	}
	gprsWire := func(set func(u *UpdateGprsLocation, v int64)) func(*testing.T, int64) berMarshaler {
		return func(t *testing.T, v int64) berMarshaler {
			w, err := convertUpdateGprsLocationToArg(gprs(0, set))
			if err != nil {
				t.Fatal(err)
			}
			switch {
			case w.UsedRATType != nil:
				r := gsm_map.UsedRATType(v)
				w.UsedRATType = &r
			case w.UeSrvccCapability != nil:
				c := gsm_map.UESRVCCCapability(v)
				w.UeSrvccCapability = &c
			case w.SmsRegisterRequest != nil:
				s := gsm_map.SMSRegisterRequest(v)
				w.SmsRegisterRequest = &s
			}
			return w
		}
	}
	setRAT := func(u *UpdateGprsLocation, v int64) { u.UsedRatType = usedRAT(v) }
	setSRVCC := func(u *UpdateGprsLocation, v int64) { c := UeSrvccCapability(v); u.UeSrvccCapability = &c }
	setSMSReg := func(u *UpdateGprsLocation, v int64) { s := SmsRegisterRequest(v); u.SmsRegisterRequest = &s }
	ati := func(si SubscriberInfo) *AnyTimeInterrogationRes { return &AnyTimeInterrogationRes{SubscriberInfo: si} }
	edrx := func(v int64) *InsertSubscriberDataArg {
		return &InsertSubscriberDataArg{EDRXCycleLengthList: EDRXCycleLengthList{{RatType: UsedRatType(v), EDRXCycleLengthValue: HexBytes{0x09}}}}
	}
	sriSm := func(v int64) *SriSm {
		n := SmDeliveryNotIntended(v)
		return &SriSm{MSISDN: "31612345678", ServiceCentreAddress: "31611111111", SmDeliveryNotIntended: &n}
	}
	rat := []int64{0, 5}
	return []extEnumCase{
		{
			name: "UpdateGprsLocation usedRAT-Type", listed: rat, want: ErrUsedRATTypeInvalid,
			msg:   func(v int64) marshaler { return gprs(v, setRAT) },
			wire:  gprsWire(setRAT),
			parse: asParser(ParseUpdateGprsLocation),
			field: func(m marshaler) int64 { return int64(*m.(*UpdateGprsLocation).UsedRatType) },
		},
		{
			name: "SubscriberInfo lastRAT-Type", listed: rat, want: ErrUsedRATTypeInvalid,
			msg: func(v int64) marshaler { return ati(SubscriberInfo{LastRATType: usedRAT(v)}) },
			wire: func(t *testing.T, v int64) berMarshaler {
				r := gsm_map.UsedRATType(v)
				return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{LastRATType: &r}}
			},
			parse: asParser(ParseAnyTimeInterrogationRes),
			field: func(m marshaler) int64 { return int64(*m.(*AnyTimeInterrogationRes).SubscriberInfo.LastRATType) },
		},
		{
			name: "LocationInformation5GS rat-Type", listed: rat, want: ErrUsedRATTypeInvalid,
			msg: func(v int64) marshaler {
				return ati(SubscriberInfo{LocationInformation5GS: &LocationInformation5GS{RatType: usedRAT(v)}})
			},
			wire: func(t *testing.T, v int64) berMarshaler {
				r := gsm_map.UsedRATType(v)
				return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{LocationInformation5GS: &gsm_map.LocationInformation5GS{RatType: &r}}}
			},
			parse: asParser(ParseAnyTimeInterrogationRes),
			field: func(m marshaler) int64 {
				return int64(*m.(*AnyTimeInterrogationRes).SubscriberInfo.LocationInformation5GS.RatType)
			},
		},
		{
			name: "EDRX-Cycle-Length rat-Type", listed: rat, want: ErrUsedRATTypeInvalid,
			msg: func(v int64) marshaler { return edrx(v) },
			wire: func(t *testing.T, v int64) berMarshaler {
				w := isdWire(t, edrx(0))
				w.EDRXCycleLengthList.Values[0].RatType = gsm_map.UsedRATType(v)
				return w
			},
			parse: asParser(ParseInsertSubscriberData),
			field: func(m marshaler) int64 {
				return int64(m.(*InsertSubscriberDataArg).EDRXCycleLengthList[0].RatType)
			},
		},
		{
			name: "UpdateGprsLocation ue-srvcc-Capability", listed: []int64{0, 1}, want: ErrUESRVCCCapabilityInvalid,
			msg:   func(v int64) marshaler { return gprs(v, setSRVCC) },
			wire:  gprsWire(setSRVCC),
			parse: asParser(ParseUpdateGprsLocation),
			field: func(m marshaler) int64 { return int64(*m.(*UpdateGprsLocation).UeSrvccCapability) },
		},
		{
			name: "UpdateGprsLocation smsRegisterRequest", listed: []int64{0, 2}, want: ErrSMSRegisterRequestInvalid,
			msg:   func(v int64) marshaler { return gprs(v, setSMSReg) },
			wire:  gprsWire(setSMSReg),
			parse: asParser(ParseUpdateGprsLocation),
			field: func(m marshaler) int64 { return int64(*m.(*UpdateGprsLocation).SmsRegisterRequest) },
		},
		{
			name: "RoutingInfoForSM sm-deliveryNotIntended", listed: []int64{0, 1}, want: ErrSMDeliveryNotIntendedInvalid,
			msg: func(v int64) marshaler { return sriSm(v) },
			wire: func(t *testing.T, v int64) berMarshaler {
				w, err := convertSriSmToArg(sriSm(0))
				if err != nil {
					t.Fatal(err)
				}
				n := gsm_map.SMDeliveryNotIntended(v)
				w.SmDeliveryNotIntended = &n
				return w
			},
			parse: asParser(ParseSriSm),
			field: func(m marshaler) int64 { return int64(*m.(*SriSm).SmDeliveryNotIntended) },
		},
	}
}

func TestExtensibleEnumEncoderStrictDecoderKeeps(t *testing.T) {
	for _, c := range extEnumCases() {
		t.Run(c.name, func(t *testing.T) {
			lo, hi := c.listed[0], c.listed[len(c.listed)-1]
			for _, v := range []int64{lo, hi} {
				data, err := c.msg(v).Marshal()
				if err != nil {
					t.Fatalf("Marshal %d: %v", v, err)
				}
				got, err := c.parse(data)
				if err != nil {
					t.Fatalf("Parse %d: %v", v, err)
				}
				if f := c.field(got); f != v {
					t.Errorf("value %d decoded as %d", v, f)
				}
			}
			for _, v := range []int64{lo - 1, hi + 1, math.MaxInt64, math.MinInt64} {
				if _, err := c.msg(v).Marshal(); !errors.Is(err, c.want) {
					t.Errorf("Marshal %d: err = %v, want %v", v, err, c.want)
				}
				data := strictBER(t, c.wire(t, v))
				got, err := c.parse(data)
				if err != nil {
					t.Fatalf("Parse %d: %v", v, err)
				}
				if f := c.field(got); f != v {
					t.Errorf("unknown value %d decoded as %d, want it kept", v, f)
				}
				if _, err := got.Marshal(); !errors.Is(err, c.want) {
					t.Errorf("Marshal of parsed %d: err = %v, want %v", v, err, c.want)
				}
				if err := checkParseRoundTrip(fmt.Sprintf("%s %d", c.name, v), c.parse, data); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
