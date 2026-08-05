package gsmmap

import (
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/gomaja/go-sms/encoding/tpdu"
)

func TestMoFsmMarshalRejectsMtTPDU(t *testing.T) {
	mtData, err := hex.DecodeString("3056800822589172230006f7840891328490000033f00440040d91328471112898f3000062805011948422324f2228e90c42a153500c34a3e1643010cd06a2c570391cc8268bd960a0a213548bc16020015990a6cb62b61a")
	if err != nil {
		t.Fatalf("hex decode MT fixture: %v", err)
	}
	mtFsm, err := ParseMtFsm(mtData)
	if err != nil {
		t.Fatalf("ParseMtFsm: %v", err)
	}

	msg := &MoFsm{
		ServiceCentreAddressDA: "2348090000330",
		MSISDN:                 "2348171182893",
		TPDU:                   mtFsm.TPDU,
	}

	if _, err := msg.Marshal(); !errors.Is(err, ErrMoFsmUnexpectedTPDUType) {
		t.Fatalf("Marshal error: got %v, want ErrMoFsmUnexpectedTPDUType", err)
	}
}

func TestMtFsmMarshalRejectsMoTPDU(t *testing.T) {
	moData, err := hex.DecodeString("302d84069122609098998206912260539128041b01510a912260716622000011d972180d4a82eee13928cc7ebbcb20")
	if err != nil {
		t.Fatalf("hex decode MO fixture: %v", err)
	}
	moFsm, err := ParseMoFsm(moData)
	if err != nil {
		t.Fatalf("ParseMoFsm: %v", err)
	}

	msg := &MtFsm{
		IMSI:                   "228519273200607",
		ServiceCentreAddressOA: "2348090000330",
		TPDU:                   moFsm.TPDU,
	}

	if _, err := msg.Marshal(); !errors.Is(err, ErrMtFsmUnexpectedTPDUType) {
		t.Fatalf("Marshal error: got %v, want ErrMtFsmUnexpectedTPDUType", err)
	}
}

func TestMtFsmMarshalAcceptsSubmitReportTPDU(t *testing.T) {
	msg := &MtFsm{
		IMSI:                   "228519273200607",
		ServiceCentreAddressOA: "2348090000330",
		TPDU: tpdu.TPDU{
			Direction:  tpdu.MT,
			FirstOctet: 0x01,
			FCS:        0x12,
			SCTS: tpdu.Timestamp{
				Time: time.Date(2015, time.May, 17, 23, 02, 50, 0, time.FixedZone("SCTS", 8*3600)),
			},
		},
	}

	data, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	got, err := ParseMtFsm(data)
	if err != nil {
		t.Fatalf("ParseMtFsm: %v", err)
	}
	if got.TPDU.SmsType() != tpdu.SmsSubmitReport {
		t.Fatalf("TPDU type: got %v, want %v", got.TPDU.SmsType(), tpdu.SmsSubmitReport)
	}
}
