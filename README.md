# go-asn1-gsmmap

High-level Go library for GSM MAP (3GPP TS 29.002) — parse, build, and marshal MAP operations using clean Go types.

Built on [go-asn1](https://github.com/gomaja/go-asn1)'s generated ASN.1 structs for correct BER encoding/decoding, with [warthog618/sms](https://github.com/warthog618/sms) for SMS TPDU handling.

## Supported Operations

| Operation | Request | Response |
|---|---|---|
| **SendRoutingInfoForSM** (SRI-SM) | `SriSm` | `SriSmResp` |
| **MT-ForwardSM** | `MtFsm` | - |
| **MO-ForwardSM** | `MoFsm` | - |
| **UpdateLocation** | `UpdateLocation` | `UpdateLocationRes` |
| **UpdateGprsLocation** | `UpdateGprsLocation` | `UpdateGprsLocationRes` |
| **AnyTimeInterrogation** (ATI) | `AnyTimeInterrogation` | `AnyTimeInterrogationRes` |
| **SendRoutingInfo** (SRI) | `Sri` | `SriResp` |

## Install

```bash
go get github.com/gomaja/go-asn1-gsmmap
```

## Usage

### Parse BER-encoded MAP data

```go
import gsmmap "github.com/gomaja/go-asn1-gsmmap"

// Parse a SendRoutingInfoForSM request
sriSm, err := gsmmap.ParseSriSm(berData)
if err != nil {
    log.Fatal(err)
}
fmt.Println(sriSm.MSISDN)              // "1234567890"
fmt.Println(sriSm.ServiceCentreAddress) // "9876543210"

// Parse an MT-ForwardSM
mtFsm, err := gsmmap.ParseMtFsm(berData)
if err != nil {
    log.Fatal(err)
}
fmt.Println(mtFsm.IMSI) // "001010123456789"
// mtFsm.TPDU contains the decoded SMS TPDU
```

### Build and marshal MAP data

```go
import gsmmap "github.com/gomaja/go-asn1-gsmmap"

sriSm := &gsmmap.SriSm{
    MSISDN:               "1234567890",
    SmRpPri:              true,
    ServiceCentreAddress: "9876543210",
}

berData, err := sriSm.Marshal()
if err != nil {
    log.Fatal(err)
}
// berData is ready to send over TCAP/SCTP
```

### AnyTimeInterrogation

```go
// Build an ATI request
ati := &gsmmap.AnyTimeInterrogation{
    SubscriberIdentity: gsmmap.SubscriberIdentity{IMSI: "001010123456789"},
    RequestedInfo: gsmmap.RequestedInfo{
        LocationInformation: true,
        SubscriberState:     true,
    },
    GsmSCFAddress: "1234567890",
}

data, err := ati.Marshal()

// Parse an ATI response
atiRes, err := gsmmap.ParseAnyTimeInterrogationRes(data)
if atiRes.SubscriberInfo.LocationInformation != nil {
    fmt.Println(atiRes.SubscriberInfo.LocationInformation.VlrNumber)
}
if atiRes.SubscriberInfo.SubscriberState != nil {
    fmt.Println(atiRes.SubscriberInfo.SubscriberState.State) // e.g. StateAssumedIdle
}
```

### SendRoutingInfo (opCode 22)

```go
// Build an SRI request
sri := &gsmmap.Sri{
    MSISDN:              "972501234567",
    InterrogationType:   gsmmap.InterrogationBasicCall,
    GmscOrGsmSCFAddress: "972531111111",
}
data, err := sri.Marshal()

// Parse an SRI response
resp, err := gsmmap.ParseSriResp(data)
if resp.NumberPortabilityStatus != nil {
    fmt.Println(*resp.NumberPortabilityStatus) // e.g. MnpOwnNumberPortedOut
}
if resp.ExtendedRoutingInfo != nil && resp.ExtendedRoutingInfo.RoutingInfo != nil {
    ri := resp.ExtendedRoutingInfo.RoutingInfo
    if ri.RoamingNumber != "" {
        fmt.Println("Roaming number:", ri.RoamingNumber)
    } else if ri.ForwardingData != nil {
        fmt.Println("Forwarded to:", ri.ForwardingData.ForwardedToNumber)
    }
}
```

## Design

This library provides a **layered API**:

- **Public types** (`SriSm`, `MtFsm`, etc.) use plain Go types — strings for phone numbers, bools for flags, `tpdu.TPDU` for SMS data.
- **Internally**, these are converted to/from [go-asn1](https://github.com/gomaja/go-asn1)'s generated `gsm_map.*` structs for BER encoding.
- **OpCode constants** can be imported directly from `github.com/gomaja/go-asn1/telecom/ss7/gsm_map` if needed for TCAP integration.

### Address handling

Phone numbers are stored as plain digit strings. Address nature and numbering plan indicators are preserved via companion fields (e.g., `MSISDNNature`, `MSISDNPlan`), defaulting to International + ISDN (E.164) when zero.

### Sub-packages

| Package | Purpose |
|---|---|
| `tbcd` | TBCD (Telephony BCD) encoding/decoding |
| `address` | MAP AddressString encoding/decoding |
| `gsn` | GSN address (IPv4/IPv6) encoding per 3GPP TS 23.003 |

## Requirements

- Go 1.21+
- [gomaja/go-asn1](https://github.com/gomaja/go-asn1) v0.1.2+

## License

MIT
