# go-asn1-gsmmap

High-level Go library for GSM MAP (3GPP TS 29.002) — parse, build, and marshal MAP operations using clean Go types.

Built on [go-asn1](https://github.com/gomaja/go-asn1)'s generated ASN.1 structs for correct BER encoding/decoding, with [warthog618/sms](https://github.com/warthog618/sms) for SMS TPDU handling.

## Supported Operations

| Operation | OpCode | Request | Response |
|---|---|---|---|
| **SendRoutingInfoForSM** (SRI-SM) | 45 | `SriSm` | `SriSmResp` |
| **MT-ForwardSM** | 44 | `MtFsm` | `MtFsmResp` |
| **MO-ForwardSM** | 46 | `MoFsm` | `MoFsmResp` |
| **UpdateLocation** | 2 | `UpdateLocation` | `UpdateLocationRes` |
| **UpdateGprsLocation** | 23 | `UpdateGprsLocation` | `UpdateGprsLocationRes` |
| **AnyTimeInterrogation** (ATI) | 71 | `AnyTimeInterrogation` | `AnyTimeInterrogationRes` |
| **SendRoutingInfo** (SRI) | 22 | `Sri` | `SriResp` |
| **InformServiceCentre** (ISC) | 63 | `InformServiceCentre` | — |
| **AlertServiceCentre** (ASC) | 64 | `AlertServiceCentre` | — |
| **PurgeMS** | 67 | `PurgeMS` | `PurgeMSRes` |

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

### InformServiceCentre (opCode 63)

```go
// Build an InformServiceCentre notification. ISC is a one-way MAP operation
// (no response is defined in 3GPP TS 29.002).
absent := 5 // AbsentSubscriberDiagnosticSM (0..255)

isc := &gsmmap.InformServiceCentre{
    StoredMSISDN: "31612345678",
    MwStatus: &gsmmap.MwStatusFlags{
        MnrfSet: true,
        McefSet: true,
    },
    AbsentSubscriberDiagnosticSM: &absent,
}
data, err := isc.Marshal()
if err != nil {
    log.Fatal(err)
}

// Parse an InformServiceCentre received from the network
parsed, err := gsmmap.ParseInformServiceCentre(data)
if err != nil {
    log.Fatal(err)
}
if parsed.MwStatus != nil && parsed.MwStatus.McefSet {
    fmt.Println("MCEF flag set for stored MSISDN:", parsed.StoredMSISDN)
}
```

### AlertServiceCentre (opCode 64)

```go
// Build an AlertServiceCentre notification. ASC is sent by the HLR to the
// SMSC to trigger retry of pending short messages once the subscriber
// becomes available again (e.g., after the MNRF flag is cleared). The
// response is an empty acknowledgement — no response type is defined on
// the public API.
event := gsmmap.SmsGmscAlertMsAvailableForMtSms

asc := &gsmmap.AlertServiceCentre{
    MSISDN:               "31612345678",
    ServiceCentreAddress: "31611111111",
    SmsGmscAlertEvent:    &event,
}
data, err := asc.Marshal()
if err != nil {
    log.Fatal(err)
}

// Parse an AlertServiceCentre received from the network
parsed, err := gsmmap.ParseAlertServiceCentre(data)
if err != nil {
    log.Fatal(err)
}
fmt.Println("SMS retry triggered for MSISDN:", parsed.MSISDN)
```

### PurgeMS (opCode 67)

```go
// Build a PurgeMS request. PurgeMS is sent by the HLR to the VLR/SGSN to
// purge subscriber data when the subscriber has been deactivated or is
// permanently unreachable. The VLR/SGSN may reply with freeze-TMSI flags
// indicating which TMSIs should be blocked.
purge := &gsmmap.PurgeMS{
    IMSI:      "204080012345678",
    VLRNumber: "31611111111",
}
data, err := purge.Marshal()
if err != nil {
    log.Fatal(err)
}

// Parse a PurgeMS response received from the network
resp, err := gsmmap.ParsePurgeMSRes(respBytes)
if err != nil {
    log.Fatal(err)
}
if resp.FreezeTMSI {
    fmt.Println("VLR asked HLR to freeze the TMSI")
}
if resp.FreezePTMSI {
    fmt.Println("SGSN asked HLR to freeze the P-TMSI")
}
if resp.FreezeMTMSI {
    fmt.Println("MME asked HLR to freeze the M-TMSI")
}
```

### SendRoutingInfo (opCode 22)

```go
// Build an SRI request
sri := &gsmmap.Sri{
    MSISDN:              "31612345678",
    InterrogationType:   gsmmap.InterrogationBasicCall,
    GmscOrGsmSCFAddress: "31201111111",
}
data, err := sri.Marshal()

// Parse an SRI response received from the network
resp, err := gsmmap.ParseSriResp(respBytes)
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
