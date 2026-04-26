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
| **SendAuthenticationInfo** (SAI) | 56 | `SendAuthenticationInfo` | `SendAuthenticationInfoRes` |
| **ProvideSubscriberInfo** (PSI) | 70 | `ProvideSubscriberInfo` | `ProvideSubscriberInfoRes` |
| **CancelLocation** | 3 | `CancelLocation` | `CancelLocationRes` |
| **InsertSubscriberData** (ISD) | 7 | `InsertSubscriberDataArg` | `InsertSubscriberDataRes` |
| _SubscriberLocationReport_ | 83 | _planned_ | _planned_ |
| _SendRoutingInfoForLCS_ | 85 | _planned_ | _planned_ |
| _ProvideSubscriberLocation_ | 86 | _planned_ | _planned_ |

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
respBytes := []byte{ /* PurgeMS-Res BER bytes from the VLR/SGSN */ }
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

### SendAuthenticationInfo (opCode 56)

```go
// Build a SendAuthenticationInfo request. SAI is sent by the VLR/SGSN/MME
// to the HLR/HSS to retrieve authentication vectors used for subscriber
// authentication and key agreement.
//
// In this example an MME requests 2 EPS authentication vectors for LTE
// authentication, identifying itself as an MME from a specific PLMN.
node := gsmmap.RequestingNodeMme

sai := &gsmmap.SendAuthenticationInfo{
    IMSI:                       "204080012345678",
    NumberOfRequestedVectors:   2,
    ImmediateResponsePreferred: true,
    AdditionalVectorsAreForEPS: true,
    RequestingNodeType:         &node,
    RequestingPLMNId:           gsmmap.HexBytes{0x62, 0xf2, 0x20}, // PLMN-Id, 3 octets
}
data, err := sai.Marshal()
if err != nil {
    log.Fatal(err)
}

// Parse a SendAuthenticationInfo response received from the HLR/HSS.
respBytes := []byte{ /* SAI-Res BER bytes from the HLR/HSS */ }
resp, err := gsmmap.ParseSendAuthenticationInfoRes(respBytes)
if err != nil {
    log.Fatal(err)
}

// Access the LTE/EPS authentication vectors.
for i, av := range resp.EpsAuthenticationSetList {
    fmt.Printf("EPS-AV[%d] RAND=%x KASME=%x\n", i, []byte(av.RAND), []byte(av.KASME))
}

// Or, if the HLR returned 2G/3G vectors:
if resp.AuthenticationSetList != nil {
    if len(resp.AuthenticationSetList.Quintuplets) > 0 {
        fmt.Println("Got 3G UMTS quintuplets:", len(resp.AuthenticationSetList.Quintuplets))
    }
    if len(resp.AuthenticationSetList.Triplets) > 0 {
        fmt.Println("Got 2G GSM triplets:", len(resp.AuthenticationSetList.Triplets))
    }
}
```

### ProvideSubscriberInfo (opCode 70)

```go
// Build a ProvideSubscriberInfo request. PSI is sent by the HLR/gsmSCF to
// the VLR/SGSN/MME to retrieve subscriber info (location, state, etc.)
// given an IMSI (+optional LMSI). The set of fields returned is governed
// by RequestedInfo — identical to the one used by ATI (opCode 71).
domain := gsmmap.PsDomain
prio := 3 // EMLPP-Priority (0..15)

psi := &gsmmap.ProvideSubscriberInfo{
    IMSI: "310150123456789",
    LMSI: gsmmap.HexBytes{0x01, 0x02, 0x03, 0x04}, // 4 octets
    RequestedInfo: gsmmap.RequestedInfo{
        LocationInformation:             true,
        SubscriberState:                 true,
        CurrentLocation:                 true,
        RequestedDomain:                 &domain,
        LocationInformationEPSSupported: true,
        RequestedNodes: &gsmmap.RequestedNodes{
            MME:  true,
            SGSN: true,
        },
    },
    CallPriority: &prio,
}
data, err := psi.Marshal()
if err != nil {
    log.Fatal(err)
}

// Parse a ProvideSubscriberInfo response received from the VLR/SGSN/MME.
respBytes := []byte{ /* PSI-Res BER bytes from the VLR/SGSN/MME */ }
resp, err := gsmmap.ParseProvideSubscriberInfoRes(respBytes)
if err != nil {
    log.Fatal(err)
}
if resp.SubscriberInfo.LocationInformation != nil {
    fmt.Println("VLR:", resp.SubscriberInfo.LocationInformation.VlrNumber)
}
if resp.SubscriberInfo.SubscriberState != nil {
    fmt.Println("State:", resp.SubscriberInfo.SubscriberState.State)
}
```

### CancelLocation (opCode 3)

```go
// Build a CancelLocation request. CancelLocation is sent by the HLR to the
// VLR/SGSN/MME to remove a subscriber's location record — e.g. after a
// successful location update in another VLR, on subscription withdrawal,
// or on initial EPS attach. The Identity field is a CHOICE: either an IMSI
// alone, or an IMSI paired with the LMSI previously assigned by the VLR.
ct := gsmmap.CancellationTypeUpdateProcedure
tu := gsmmap.TypeOfUpdateSgsnChange

cl := &gsmmap.CancelLocation{
    Identity:         gsmmap.CancelLocationIdentity{IMSI: "204080012345678"},
    CancellationType: &ct,
    TypeOfUpdate:     &tu,
    NewMSCNumber:     "31611111111",
    NewVLRNumber:     "31622222222",
    NewLMSI:          gsmmap.HexBytes{0x11, 0x22, 0x33, 0x44},
    ReattachRequired: true,
}
data, err := cl.Marshal()
if err != nil {
    log.Fatal(err)
}

// Build a CancelLocation using the IMSI-with-LMSI alternative of the
// Identity CHOICE — often used when the HLR already knows the LMSI the
// VLR previously assigned to the subscriber.
clWithLmsi := &gsmmap.CancelLocation{
    Identity: gsmmap.CancelLocationIdentity{
        IMSIWithLMSI: &gsmmap.CancelLocationIMSIWithLMSI{
            IMSI: "204080012345678",
            LMSI: gsmmap.HexBytes{0xA1, 0xB2, 0xC3, 0xD4}, // 4 octets
        },
    },
}
if _, err := clWithLmsi.Marshal(); err != nil {
    log.Fatal(err)
}

// Parse a CancelLocation response received from the VLR/SGSN/MME. The wire
// response is effectively empty in practice — only an optional
// ExtensionContainer is defined in 3GPP TS 29.002.
respBytes := []byte{ /* CancelLocation-Res BER bytes */ }
if _, err := gsmmap.ParseCancelLocationRes(respBytes); err != nil {
    log.Fatal(err)
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

#### CAMEL subscription info in SRI responses

The `ExtendedRoutingInfo` CHOICE carries a `CamelRoutingInfo` alternative
that exposes the GMSC's full CAMEL subscription information (T-CSI, O-CSI,
D-CSI, and BCSM-CAMEL-TDP criteria lists) with field-level coverage. Every
nested SEQUENCE, enum, and trigger-detection-point round-trips between Go
and BER without data loss.

```go
phase := 2
resp := &gsmmap.SriResp{
    IMSI: "310260123456789",
    ExtendedRoutingInfo: &gsmmap.ExtendedRoutingInfo{
        CamelRoutingInfo: &gsmmap.CamelRoutingInfo{
            GmscCamelSubscriptionInfo: gsmmap.GmscCamelSubscriptionInfo{
                OCSI: &gsmmap.OCSI{
                    OBcsmCamelTDPDataList: []gsmmap.OBcsmCamelTDPData{
                        {
                            OBcsmTriggerDetectionPoint: gsmmap.OBcsmTriggerCollectedInfo,
                            ServiceKey:                 42,
                            GsmSCFAddress:              "31611111111",
                            DefaultCallHandling:        gsmmap.DefaultCallHandlingContinueCall,
                        },
                    },
                    CamelCapabilityHandling: &phase,
                    NotificationToCSE:       true,
                },
            },
        },
    },
}
data, err := resp.Marshal()
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
