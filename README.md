# go-asn1-gsmmap

High-level Go library for GSM MAP (3GPP TS 29.002) — parse, build, and marshal MAP operations using clean Go types.

Built on [go-asn1](https://github.com/gomaja/go-asn1)'s generated ASN.1 structs for correct BER encoding/decoding, with [go-sms](https://github.com/gomaja/go-sms) for SMS TPDU handling.

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
| **ProvideSubscriberLocation** (PSL) | 83 | `ProvideSubscriberLocationArg` | `ProvideSubscriberLocationRes` |
| **SendRoutingInfoForLCS** (SRI-LCS) | 85 | `SriLcs` | `SriLcsResp` |
| **SubscriberLocationReport** (SLR) | 86 | `SubscriberLocationReportArg` | `SubscriberLocationReportRes` |
| **ReportSMDeliveryStatus** | 47 | `ReportSMDeliveryStatus` | `ReportSMDeliveryStatusRes` |
| **processUnstructuredSS-Request** (USSD) | 59 | `USSDArg` | `USSDRes` |
| **unstructuredSS-Request** (USSD) | 60 | `USSDArg` | `USSDRes` |
| **unstructuredSS-Notify** (USSD) | 61 | `USSDArg` | — |

MAP ReturnError parameters are decoded with `ParseReturnErrorParameter` (see [MAP errors](#map-errors)).

## Install

Depend on the main branch:

```bash
go get github.com/gomaja/go-asn1-gsmmap@main
```

The main branch is the only supported version. The API follows the current
3GPP specifications and the main branches of
[go-asn1](https://github.com/gomaja/go-asn1) and
[go-sms](https://github.com/gomaja/go-sms), and keeps no backward-compatible
aliases or wrappers: when something changes, there is one way to do it.

The module was tagged v1.0.0, v1.0.1 and v1.0.2 before it moved to the main
branch. Those tags and their GitHub releases are deleted, and v1.0.3 was
published only to retract all four (see the `retract` block in `go.mod`).
What that means for a consumer:

- proxy.golang.org and sum.golang.org keep v1.0.0–v1.0.3 permanently; a
  deleted tag cannot be removed from them. Because they are retracted, the go
  command does not offer them: `go get github.com/gomaja/go-asn1-gsmmap` with
  no version, `@latest` and `go list -m -u` resolve to the main branch.
- Without tags, a version of the main branch is a pseudo-version
  `v0.0.0-<UTC time>-<commit>`, which sorts below v1.0.x. `go get -u` never
  moves such a pseudo-version back to v1.0.x (the go command does not
  "upgrade" to a chronologically older version), but minimal version
  selection does compare by version: if any module in your build still
  requires v1.0.x, that retracted release wins over the main branch, and
  `go build` and `go mod tidy` say nothing about it. `go get` warns that the
  selected version is retracted, and `go list -m -u all` marks it
  `(retracted)`; update the module that requires it.
- A project that requires v1.0.x is warned that it is retracted. `go get -u`
  leaves it there, because the main branch's pseudo-versions sort below
  v1.0.x; it moves to the main branch with the command above. The v1.0.x API
  differs (see the migration notes below).

### Migrating from v1.0.x

**Dependencies and errors**

- The main branch requires the current `go-asn1` and `go-sms` modules. Codec-enforced ASN.1 sizes (including SEQUENCE OF elements), integer ranges and non-extensible ENUMERATED sets return `*ber.ConstraintError` (`errors.As`). A BIT STRING whose octet count disagrees with its bit length returns an error wrapping `ber.ErrInvalidValue` (`errors.Is`); when the bit length also breaks the SIZE, the error can be that SIZE's `*ber.ConstraintError` instead ([go-asn1#137](https://github.com/gomaja/go-asn1/issues/137)). Package semantic checks return named `Err*` sentinels (`errors.Is`).
- Removed constraint and bounds sentinels include `ErrSaiInvalidPLMNId`, `ErrIntervalTimeOutOfRange`, `ErrGPRSDataListSize`, and `ErrUtranBaroPressureMeasOutOfRange`. Sender rules retain dedicated sentinels for restrictions narrower than the ASN.1 constraints.
- Element-size sentinels from v1.0.2 replaced by `*ber.ConstraintError`: `ErrCamelInvalidCauseValueOctetLength`, `ErrZoneCodeInvalidSize`, `ErrAPNInvalidSize`, `ErrResetIdInvalidSize`, `ErrIsdBearerServiceCodeSize`, `ErrIsdTeleserviceCodeSize`, `ErrIsdResSsListSize`, `ErrMMCodeInvalidSize`.
- Non-extensible ENUMERATED sentinels from v1.0.2 replaced by `*ber.ConstraintError`: `ErrSriInvalidInterrogationType`, `ErrAscInvalidSmsGmscAlertEvent`, `ErrCamelInvalidCallTypeCriteria`, `ErrCamelInvalidMatchType`, `ErrIntraCUGOptionsInvalidValue`, `ErrCliRestrictionOptionInvalidValue`, `ErrOverrideCategoryInvalidValue`, `ErrSIPTOPermissionInvalid`, `ErrSIPTOLocalNetworkPermissionInvalid`, `ErrLIPAPermissionInvalid`, `ErrNIDDMechanismInvalid`, `ErrLSAOnlyAccessIndicatorInvalid`, `ErrPDNGWAllocationTypeInvalid`, `ErrPDNConnectionContinuityInvalid`, `ErrWLANOffloadabilityIndicationInvalid`, `ErrReportSMDeliveryStatusOutcomeInvalid`.
- `ErrCamelInvalidCamelCapabilityHandling` is `ErrCamelCapabilityHandlingOutOfRange`. Other removed `Err*` names for duplicated presence, identity, size, receiver, and digitless-address checks are replaced by the applicable current semantic sentinel or codec error; update `errors.Is` checks by condition.
- `ErrReportSMDeliveryStatusResStoredMSISDNEmpty` → `ErrReportSMDeliveryStatusResStoredMSISDNDecodedEmpty`; `ErrLCSClientIDDialedByMSEmpty` splits into `ErrLCSClientIDDialedByMSDecodedEmpty` on decode and `ErrLCSClientIDDialedByMSNaturePlanWithoutDigits` on encode.
- Error names follow current Go fields: `ErrUpdateLocationMissingMSCNumber` → `ErrUpdateLocationMissingMscNumber` and `ErrUpdateLocationMissingVLRNumber` → `ErrUpdateLocationMissingVlrNumber`.
- A digitless mandatory value returns the sentinel `Marshal` returns for the missing field: `ErrPSLArgMlcNumberDecodedEmpty` → `ErrPSLArgMlcNumberEmpty`, `ErrSriLcsMlcNumberDecodedEmpty` → `ErrSriLcsMlcNumberEmpty`, `ErrLCSLocationInfoNetworkNodeDecodedEmpty` → `ErrLCSLocationInfoNetworkNodeNumberEmpty`, `ErrReportSMDeliveryStatusMSISDNDecodedEmpty` → `ErrReportSMDeliveryStatusMSISDNEmpty` and `ErrReportSMDeliveryStatusSCADecodedEmpty` → `ErrReportSMDeliveryStatusServiceCentreAddressEmpty`. `ErrCSGIdInvalidSize` is replaced by the codec's `*ber.ConstraintError` for SIZE (27), or an error wrapping `ber.ErrInvalidValue` for inconsistent bit length and octet count.
- `ErrSriMissingGmsc` → `ErrSriMissingGmscOrGsmSCFAddress`; SM-RP-DA/OA service-centre address sentinels add `DA`/`OA` to match their fields. `ErrLCSLocationInfoNetworkNodeEmpty` → `ErrLCSLocationInfoNetworkNodeNumberEmpty`, and `ErrReportSMDeliveryStatusSCAEmpty` → `ErrReportSMDeliveryStatusServiceCentreAddressEmpty`.
- CHOICE errors are named after their CHOICE type: `Err<Type>NoAlternative` and `Err<Type>MultipleAlternatives` from `Marshal`, `Err<Type>UnknownAlternative` from `Parse`. `ErrSriChoice*` becomes `ErrRoutingInfo*`, `ErrExtendedRoutingInfo*`, `ErrAdditionalNumber*`, `ErrExtBasicServiceCode*` or `ErrEpsInfo*` by type; `ErrMoFsmSmRp{Da,Oa}*` and `ErrMtFsmSmRp{Da,Oa}*` become `ErrSmRp{Da,Oa}*`; `ErrAtiPsSubscriberState*` becomes `ErrPsSubscriberState*`; `ErrCancelLocIdentityChoice{NoAlternative,Multiple}` becomes `ErrCancelLocationIdentity{NoAlternative,MultipleAlternatives}`; `ErrExtSSInfoChoice*`, `ErrSSSubscriptionOptionChoice*` and `ErrSaiAuthSetListChoice*` become `ErrExtSSInfo*`, `ErrSSSubscriptionOption*` and `ErrAuthenticationSetList*`; `ErrServingNodeAddress{NoAlt,MultipleAlts}` and `ErrSubscriberIdentity{NoAlt,MultipleAlts,UnknownChoice}` take the full suffixes; `ErrPSLResCellIdOrSaiInvalidChoice` becomes `ErrCellGlobalIdOrServiceAreaIdOrLAIUnknownAlternative`, and `ErrPSLResCellGlobalIdAndLAIMutex` and `ErrSLRArgCellGlobalIdAndLAIMutex` become `ErrCellGlobalIdOrServiceAreaIdOrLAIMultipleAlternatives`, which CS and GPRS location information now also return.
- `ErrExtNoRepCondTimeOutOfRange` becomes `ErrNoReplyConditionTimeOutOfRange` for the sender range 5 to 30, and `ErrGMLCAddressEmpty` becomes `ErrGMLCListEntryEmpty`.

**Identities and addresses**

- Marshal rejects reserved address natures `0x50`/`0x70` with `ErrAddressNatureInvalid` and accepts only numbering plans `0`, `1`, `3`, `4`, `6`, `8`, `9`, returning `ErrAddressPlanInvalid` for other values (TS 29.002 §17.7.8). Parse preserves received nature and plan values.
- Marshal rejects non-ASCII Diameter names and realms with `ErrDiameterIdentityNotASCII` (RFC 6733 §4.3.1). This checks the encoding; callers still supply valid FQDNs, realms, and unique node identities. Parse preserves the received octets.
- Marshal rejects nondecimal MCC/MNC digits in PLMN identifiers, location-identifier prefixes, and LCS areas with `ErrPLMNIdentityInvalidDigits`; only the third MNC digit may be `0xf`. Two-octet country areas check the three MCC digits; longer areas also check the included MNC (TS 29.002 §§17.7.8, 17.7.13). Parse preserves the received octets.
- `InsertSubscriberDataArg.IMSI` and `SriSmCorrelationID.HlrID` change from `HexBytes` to decimal digit strings. IMSI, IMEI, IMEISV, and HLR-Id follow TS 23.003 digit counts; `GroupId` and `LongGroupId` use TBCD digit strings. Parse accepts any 15th IMEI digit, while Marshal requires the spare digit 0.
- Identity failures use `ErrIMSIInvalidLength`, `ErrIMEIInvalidLength`, `ErrIMEISVInvalidLength`, `ErrIdentityEmpty`, or `ErrIdentityNotDigits`. The per-message identity sentinels (`ErrIsdIMSIInvalidSize`, `ErrPSLArgIMSI*`, `ErrPSLArgIMEI*`, `ErrSLRArgIMSI*`, `ErrSLRArgIMEI*`, `ErrReportSMDeliveryStatusIMSI*`, `ErrSubscriberIdentityIMSIDecodedEmpty`) are replaced by these. A missing or digitless IMSI returns `ErrIdentityEmpty` on both `Marshal` and `Parse`; `ErrSaiMissingIMSI`, `ErrPurgeMSMissingIMSI`, `ErrUpdateLocationMissingIMSI`, `ErrPsiMissingIMSI`, and `ErrCancelLocIdentityMissingIMSI` are gone.
- `ForwardingData.LongForwardedToNumber` changes from `HexBytes` to digits plus `LongForwardedToNumberNature/Plan`; `ExtForwFeature.ForwardedToNature/Plan` become `ForwardedToNumberNature/Plan`, and its long FTN has separate nature/plan fields.
- Address nature and plan zero mean unknown. Set them explicitly for outgoing messages; parsing can normalize a trailing all-filler octet away and set the address extension bit to 1 when remarshal occurs.
- A present optional wire address that decodes to no digits fails with an `Err*DecodedEmpty` sentinel. Mandatory digitless addresses return the corresponding missing-field or identity sentinel.

**Receiver and sender rules**

- `UpdateGprsLocation.Marshal` rejects SGSN LCS capability set1 (`ErrSGSNCapabilityLCSSet1NotAllowed`), PDN-GW updates lacking an identity or lacking both APN and ContextID (`ErrPdnGwUpdateMissingIdentity`, `ErrPdnGwUpdateMissingAPNOrContextID`), and set ISR bits outside 0..2 (`ErrISRInformationUnknownBits`). Parse retains SGSN set1 and identity-only updates; its existing PDN identity and ISR receiver rules still apply.
- `InsertSubscriberDataArg.Marshal` requires PDN types 01..04 (`ErrPDNTypeInvalid`). Typed and opaque PDP contexts require base types F001/F002/F121/F157, extended type F18D and paired addresses of different IPv4/IPv6 families (`ErrPDPTypeInvalid`, `ErrExtPDPTypeInvalid`, `ErrPDPAddressFamilyCombinationInvalid`). Typed and opaque PDP contexts share the QoS predecessor and base address rules (`ErrExtQoSHierarchyViolated`, `ErrExtPDPAddressWithoutPDPAddress`); Parse preserves received values.
- `ProvideSubscriberInfoRes.Marshal` rejects IMS voice-over-PS `unknown` (`ErrPSIResImsVoiceOverPSSessionsIndicationUnknown`). Number portability values 4/5 may be sent only by ATI responses; PSI responses and top-level or nested SRI results return `ErrNumberPortabilityStatusATIOnly`. Parse retains those received values.
- ATI and PSI request encoders enforce `RequestedInfo` dependencies: CurrentLocation and LocationInformationEPSSupported require LocationInformation; RequestedNodes is forbidden with explicit cs-Domain; ServingNodeIndication requires LocationInformation and excludes CurrentLocation. Violations return `ErrRequestedInfoCurrentLocationWithoutLocationInformation`, `ErrRequestedInfoEPSSupportedWithoutLocationInformation`, `ErrRequestedInfoRequestedNodesInCSDomain`, `ErrRequestedInfoServingNodeWithoutLocationInformation` or `ErrRequestedInfoServingNodeWithCurrentLocation`; Parse retains received combinations.
- All subscriber location encoders require the Q.763 uncertainty-circle geodetic shape (`ErrGeodeticInformationShapeInvalid`). MNP routing-number TBCD permits filler only in the final high nibble (`ErrRouteingNumberTBCDInvalid`). Parse retains received octets. PSI and SRI call-priority encoders reject spare priorities 7..15 with `ErrEMLPPPriorityOutOfRange`; out-of-ASN priority ranges return the codec constraint error, including in EMLPPInfo.
- `SriSmResp.Marshal` permits at most one of its four network node indicators (`ErrSriSmRespNodeIndicatorsMutuallyExclusive`) and requires `AdditionalNumber` and `ThirdNumber` to carry different CHOICE types (`ErrSriSmRespAdditionalNumbersSameType`). Parse preserves received overlaps.
- `ReportSMDeliveryStatus.Marshal` rejects indicator/outcome/diagnostic conflicts with `ErrReportSMDeliveryStatusConflictingOutcomeFields`. `InformServiceCentre.Marshal` requires a primary diagnostic when its additional diagnostic is present (`ErrIscAdditionalDiagnosticWithoutPrimary`); Parse keeps those received fields.
- `ProvideSubscriberLocationArg.Marshal` and `SubscriberLocationReportArg.Marshal` require IMSI or MSISDN (`ErrLCSSubscriberIdentityMissing`), the applicable deferred reference (`ErrLCSDeferredReferenceNumberMissing`), and LCSClientName DCS `0x0f` (`ErrLCSClientNameDataCodingSchemeInvalid`). RequestorID and Codeword retain their own DCS definitions; Parse preserves received sender violations.
- `LCSLocationInfo` rejects its GPRS indicator with dummy node `0` on Marshal (`ErrLCSLocationInfoGPRSIndicatorWithDummyNode`). A deferred SLR response requires `DeferredmtLrData`, other SLR events exclude it (`ErrSLRArgDeferredmtLrDataPresence`), and its nested location requires termination cause `mt-lrRestart` (`ErrDeferredmtLrDataLocationInfoRequiresRestart`). `SubscriberLocationReportRes.Marshal` rejects both routing identifiers (`ErrSLRResRoutingIdentifiersMutuallyExclusive`); Parse retains its return-error handling.
- PSL results and SLR arguments reject undefined or incorrectly coded geographic shapes, velocity types and velocity lengths other than 4, 5, 5 and 7 octets for types 0 to 3 on Marshal (`ErrExtGeographicalInformationInvalid`, `ErrAddGeographicalInformationInvalid`, `ErrVelocityEstimateTypeInvalid`, `ErrVelocityEstimateLengthInvalid`); Parse keeps a velocity whose length does not match its type, since §17.7.13 makes only an unlisted type invalid data. Geography supports all current TS 23.032 shapes in Add, including high accuracy scalable forms; Ext supports its five MAP shapes. ASN.1 SIZE remains enforced by the codec, and Parse keeps its existing discard and invalid-data rules.
- Decoders apply TS 29.002 receiver rules after strict BER validation. BER syntax errors and ASN.1 constraint violations fail before ignore or discard rules; semantic checks apply to effective values. Encoders accept only sender values. For example, CAMEL capability handling above 4 decodes as 4, while NoReplyConditionTime encodes only 5 through 30.
- `Marshal` rejects reserved SS-Code assignments outside the ISD result echo lists (`ErrSSCodeReserved`) and SRI ccbs-A/ccbs-B (`ErrSSCodeOperationInvalid`). Forwarding options (`ErrExtForwOptionsUnusedBits`), InterCUG restrictions (`ErrInterCUGRestrictionsUnusedBits`), ISDN subaddress headers (`ErrISDNSubaddressHeaderInvalid`), and Ext-SS-Status (`ErrExtSSStatusUnusedBits`) enforce their sender bit assignments; Ext-SS-Status and Ext-ForwOptions send one octet only (`ErrExtSSStatusReservedOctets`, `ErrExtForwOptionsReservedOctets`). Received reserved bits remain available where the receiver rules retain them.
- ISD LCS privacy fields must match their SS-Code: notification and external clients use call-session codes, PLMN clients use `0xb4`, and service types use `0xb5`. Extended external clients require more than five clients in total. The new `ErrLCSPrivacy*` sentinels identify violations; callers ensure the required peer release and SCCP support. LCS service type 8 returns `ErrLCSServiceTypeIDReserved`, and SRI alerting patterns now share `ErrAlertingPatternReserved`.
- Extended bearer and teleservice codes send one octet; later octets return `ErrExtBearerServiceCodeReservedOctets` or `ErrExtTeleserviceCodeReservedOctets`. ISD arguments and SRI reject compound service groups with `ErrCompoundBearerServiceCodeNotAllowed` or `ErrCompoundTeleserviceCodeNotAllowed`. Bearer codes with bit 8 set require the PLMN-specific `0xd0..0xdf` family (`ErrBearerServiceCodeUnusedBit`). ISD result lists echo unsupported SS, bearer and teleservice codes with only ASN.1 SIZE checks (TS 29.002 §8.8.1.3).
- `MTSMSTPDUTypeSmsSUBMITREPORT` is removed: CAMEL phase 4 permits only `sms-DELIVER` and `sms-STATUS-REPORT` in MT-SMS TPDU criteria; the sender rejects that value and the receiver ignores the entire TPDU-TypeCriterion containing it.
- `MtFsm` and `MoFsm` expose SM-RP-DA and SM-RP-OA through `SmRpDa` and `SmRpOa` CHOICE fields. Their former shorthand identity and service-centre fields, including nature/plan fields, are removed. `ErrMtFsmMissingIMSI` is gone: an `SmRpDa` with no alternative set returns `ErrSmRpDaNoAlternative`.
- CSI types (`OCSI`, `TCSI`, `DCSI`, `SSCSI`, `MCSI`, `SMSCSI`, `GPRSCSI`, `MGCSI`) omit `NotificationToCSE` and `CsiActive`; Parse drops those wire fields. Segmented CSI and SRI presence rules that depend on other segments remain the caller's responsibility, so `ErrCamelSMSCSIMissingTDPData`, `ErrCamelSMSCSIMissingCapabilityHandling` and `ErrGPRSCSIRequiresTDPListAndPhase` are removed.
- `RoamingNotAllowedParam.RoamingNotAllowedCause` is a pointer, nil when the additional cause is present after successful strict decoding. `LCSQoS.LcsQosClass` decodes unrecognized values as `LCSQoSClassBestEffort`.
- `ParseReturnErrorParameter(MapErrorCode, []byte)` replaces the per-error `Parse*Param` functions, and `MapErrorCode(code).String()` replaces `GetErrorString(code)`.
- `ParseReturnErrorParameter(MapErrorRoamingNotAllowed, nil)` or an empty parameter returns `ErrRoamingNotAllowedParameterMissing`; the caller constructs the TCAP mistyped-parameter reject required by TS 29.002 §17.1.2.
- Parse preserves both CancelLocation MTRF flags, TypeOfUpdate with any CancellationType, independent PDP QoS/address extensions, repeated CAMEL TDP entries, D-CSI capability without a list, DestinationNumberCriteria without either list, both short and long group IDs, QoS accuracy high bits, and periodic reporting operands whose product exceeds the sender cap. Marshal keeps its sender restrictions.
- Parse ignores EPS info whose PDN-GW update lacks an identity, clears unknown ISR bits, retains only the first received forwarding-options and Ext-SS-Status octet, masks InterCUGRestrictions to its low two bits, and maps call priorities 7..15 to 4. Opaque PDP contexts keep their private extensions: only the caller knows which it supports (§17.7.11). A PSL request whose only area event is beingInsideArea gets AreaEventInfo.OccurrenceInfo oneTimeEvent, as §17.7.13 treats that event.
- Location estimates follow TS 29.002 §17.7.13 in PSL results and SLR arguments: a valid LocationEstimate discards AddLocationEstimate; a one-octet LocationEstimate is discarded beside AddLocationEstimate and is invalid data without one; a PSL result with DeferredmtLrResponseIndicator ignores its LocationEstimate. Invalid data, including an invalid estimate's shape or a velocity type outside 0..3, returns `ErrExtGeographicalInformationInvalidData`, `ErrAddGeographicalInformationInvalidData` or `ErrVelocityEstimateInvalidData`. Marshal sends an empty PSL LocationEstimate beside AddLocationEstimate or DeferredmtLrResponseIndicator as one octet, and rejects a one-octet one without them (`ErrLocationEstimateOneOctetWithoutAdd`). SLR responses containing both NA-ESRK and NA-ESRD return `ErrSLRResRoutingIdentifiersReturnError`; the caller treats the response as unsuccessful.
- `GeographicalInfo.Encode` uses the TS 23.032 §6.1 floor bins and accepts latitudes ±90, encoding each pole in the maximum magnitude bin; decoding still returns the bin's lower edge, and -0 for the southern bin 0, which encodes back to it. `ErrGeoLatitudeEncodingOverflow`, `ErrGeoLongitudeEncodingOverflow` and `ErrGeoLongitudeEncodingCollapse` are removed; these sentinels were not present in v1.0.2. `ShapeEllipsoidPointAltitude` becomes `ShapeEllipsoidPointAltitudeUncertaintyEllipsoid`: shape 9 is the point with altitude and uncertainty ellipsoid.

**Renamed or retyped fields**

- `SubscriberStateInfo.NotReachableReason` changes from `*int` to `*NotReachableReason`, an alias of `gsm_map.NotReachableReason`.
- `PsSubscriberState.NetDetNotReachable` changes from `*int` to `*NotReachableReason`; this also applies when used as `SubscriberInfo.EpsSubscriberState`.
- `SubscriberInfo.DaylightSavingTime` changes from `*int` to `*DaylightSavingTime`, an alias of `gsm_map.DaylightSavingTime`, with `DaylightSavingTimeNoAdjustment`, `DaylightSavingTimePlusOneHourAdjustment`, and `DaylightSavingTimePlusTwoHoursAdjustment` constants.
- `VlrCapability.IstSupportIndicator` changes from `*int` to `*ISTSupportIndicator`, an alias of `gsm_map.ISTSupportIndicator`, with `ISTSupportIndicatorBasicISTSupported` and `ISTSupportIndicatorIstCommandSupported` constants.
- `Sri.IstSupportIndicator` changes from `*int` to `*ISTSupportIndicator`.
- `ExternalSignalInfo.ProtocolID` changes from `int` to `ProtocolID`, an alias of `gsm_map.ProtocolId`, with `ProtocolIDGsm0408`, `ProtocolIDGsm0806`, `ProtocolIDGsmBSSMAP`, and `ProtocolIDEts3001021` constants. The reserved `gsm-BSSMAP` sender rule still applies.
- `ExtExternalSignalInfo.ExtProtocolID` changes from `int` to `ExtProtocolID`, an alias of `gsm_map.ExtProtocolId`, with the `ExtProtocolIDEts300356` constant.
- `CSGSubscriptionData.CsgId/CsgIdBitLength` → `CsgID/CsgIDBits`; set `CsgIDBits` to 27, the CSG-Id SIZE (27), since zero is rejected. `VoiceGroupCallData.AdditionalInfo` gains `AdditionalInfoBits` and keeps values that are not byte-aligned. `SGSNCapability.SupportedFeatures/ExtSupportedFeatures` become pointers to named feature structs; `ExtSupportedFeatures` gains `UnknownBits`, making it non-comparable, and both feature structs preserve bit lengths. `SupportedFeaturesBits` and `ExtSupportedFeaturesBits` are removed.
- `ExtAccessRestrictionData` gains `BitLength` and `UnknownBits`, preserving unnamed bits 2..31 and trailing zero bits on Parse/Marshal. It is no longer comparable; use keyed struct literals. `UnknownBits` uses MSB-first positions with bits 0..1 ignored, following `ExtSupportedFeatures.UnknownBits`.
- `ProvideSubscriberLocationRes.AgeOfLocationEstimate` and `SubscriberLocationReportArg.AgeOfLocationEstimate` change from `*int64` to `*int`; so does `InsertSubscriberDataArg.IstAlertTimer`. `MCSI.MobilityTriggers` and `MGCSI.MobilityTriggers` become `[]MMCode`.
- `GsmSCFNature/Plan` → `GsmSCFAddressNature/Plan` in ATI and CAMEL structs. `SCANature/Plan` → `ServiceCentreAddressNature/Plan` in SRI-SM, AlertServiceCentre, and ReportSMDeliveryStatus; `SCADANature/Plan` and `SCAOANature/Plan` follow their full ServiceCentreAddress field names.
- `UpdateLocation.MSCNumber/VLRNumber` → `MscNumber/VlrNumber`; `UpdateGprsLocation.SGSNNumber` → `SgsnNumber`; `PurgeMS.VLRNumber/SGSNNumber` → `VlrNumber/SgsnNumber`; `UpdateLocationRes.HLRNumber` and `UpdateGprsLocationRes.HLRNumber` → `HlrNumber`. Each associated nature/plan field follows the new base name.
- `Sri.GmscNature/Plan` → `GmscOrGsmSCFAddressNature/Plan`, and `SriResp.VmscNature/Plan` → `VmscAddressNature/Plan`.
- LCS client and codeword data coding schemes use `USSDDataCodingScheme`, with strings decoded by `DataCodingScheme.Decode`.

**Removed identifiers**

- `ReasonMsPurged` → `NotReachableReasonMsPurged`, a typed go-asn1 constant.
- `ReasonImsiDetached` → `NotReachableReasonImsiDetached`, a typed go-asn1 constant.
- `ReasonRestrictedArea` → `NotReachableReasonRestrictedArea`, a typed go-asn1 constant.
- `ReasonNotRegistered` → `NotReachableReasonNotRegistered`, a typed go-asn1 constant.
- 53 exported min/max and bound constants are removed; examples include `AreaListMaxEntries`, `MaxNumOfResetId`, `ReportingIntervalMax`, and `VelocityEstimateMinLen`. Use the codec constraints and current semantic errors rather than these constants.
- `ErrCamelInvalidMobilityTriggerOctet`, `CSGIdBitLength`, and `MaxResetIdOctets` are removed. `GMLCAddress` is removed: `GMLCList` is a `[]ISDNNumber`, like the CAMEL `DestinationNumberList`. Wrong-size MM-Code values return the codec's `*ber.ConstraintError`.

## Usage

### Parse BER-encoded MAP data

Each `Parse` function decodes the BER of one argument or result parameter. TS 29.002 marks several results optional, for example those of CancelLocation, InsertSubscriberData, PurgeMS, SendAuthenticationInfo, ReportSM-DeliveryStatus and unstructuredSS-Request. Such a return result may carry no parameter, and then there is nothing to parse.

```go
import gsmmap "github.com/gomaja/go-asn1-gsmmap"

// Parse a SendRoutingInfoForSM request
var berData []byte // BER request bytes received from the network
sriSm, err := gsmmap.ParseSriSm(berData)
if err != nil {
    log.Fatal(err)
}
fmt.Println(sriSm.MSISDN)               // 1234567890
fmt.Println(sriSm.ServiceCentreAddress) // 9876543210

// Parse an MT-ForwardSM
var mtFsmBytes []byte // BER MT-ForwardSM bytes received from the network
mtFsm, err := gsmmap.ParseMtFsm(mtFsmBytes)
if err != nil {
    log.Fatal(err)
}
fmt.Println(mtFsm.SmRpDa.IMSI) // "001010123456789" when SM-RP-DA is an IMSI
// mtFsm.TPDU contains the decoded SMS TPDU
```

### Build and marshal MAP data

```go
import (
    gsmmap "github.com/gomaja/go-asn1-gsmmap"
    "github.com/gomaja/go-asn1-gsmmap/address"
)

sriSm := &gsmmap.SriSm{
    MSISDN:               "1234567890",
    MSISDNNature:         address.NatureInternational,
    MSISDNPlan:           address.PlanISDN,
    SmRpPri:              true,
    ServiceCentreAddress: "9876543210",
    ServiceCentreAddressNature:            address.NatureInternational,
    ServiceCentreAddressPlan:              address.PlanISDN,
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
    GsmSCFAddressNature:  address.NatureInternational,
    GsmSCFAddressPlan:    address.PlanISDN,
}

data, err := ati.Marshal()
if err != nil {
    log.Fatal(err)
}
_ = data // send the request bytes to the HLR

// Parse an ATI response received from the HLR
var respBytes []byte // BER response bytes received from the HLR
atiRes, err := gsmmap.ParseAnyTimeInterrogationRes(respBytes)
if err != nil {
    log.Fatal(err)
}
if atiRes.SubscriberInfo.LocationInformation != nil {
    fmt.Println(atiRes.SubscriberInfo.LocationInformation.VlrNumber)
}
if atiRes.SubscriberInfo.SubscriberState != nil {
    fmt.Println(atiRes.SubscriberInfo.SubscriberState.State) // 0
}
```

### InformServiceCentre (opCode 63)

```go
// Build an InformServiceCentre notification. ISC is a one-way MAP operation
// (no response is defined in 3GPP TS 29.002).
absent := 5 // AbsentSubscriberDiagnosticSM (0..255)

isc := &gsmmap.InformServiceCentre{
    StoredMSISDN:       "31612345678",
    StoredMSISDNNature: address.NatureInternational,
    StoredMSISDNPlan:   address.PlanISDN,
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
    MSISDNNature:         address.NatureInternational,
    MSISDNPlan:           address.PlanISDN,
    ServiceCentreAddress: "31611111111",
    ServiceCentreAddressNature:            address.NatureInternational,
    ServiceCentreAddressPlan:              address.PlanISDN,
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
// Build a PurgeMS request. The VLR or SGSN sends it to the HLR when
// deleting a subscriber record. The HLR replies with freeze-TMSI flags.
purge := &gsmmap.PurgeMS{
    IMSI:      "204080012345678",
    VlrNumber: "31611111111",
    VlrNumberNature: address.NatureInternational,
    VlrNumberPlan:   address.PlanISDN,
}
data, err := purge.Marshal()
if err != nil {
    log.Fatal(err)
}

// Parse a PurgeMS response received from the HLR
respBytes := []byte{ /* PurgeMS-Res BER bytes from the HLR */ }
resp, err := gsmmap.ParsePurgeMSRes(respBytes)
if err != nil {
    log.Fatal(err)
}
if resp.FreezeTMSI {
    fmt.Println("HLR asked VLR to freeze the TMSI")
}
if resp.FreezePTMSI {
    fmt.Println("HLR asked SGSN to freeze the P-TMSI")
}
if resp.FreezeMTMSI {
    fmt.Println("Freeze M-TMSI requested")
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
// Build a ProvideSubscriberInfo request. PSI is sent by the HLR to
// the VLR/SGSN/MME to retrieve subscriber info (location, state, etc.)
// given an IMSI (+optional LMSI). The set of fields returned is governed
// by RequestedInfo — identical to the one used by ATI (opCode 71).
domain := gsmmap.PsDomain
prio := 3 // EMLPP-Priority sender values (0..6)

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
    NewMSCNumber:       "31611111111",
    NewMSCNumberNature: address.NatureInternational,
    NewMSCNumberPlan:   address.PlanISDN,
    NewVLRNumber:       "31622222222",
    NewVLRNumberNature: address.NatureInternational,
    NewVLRNumberPlan:   address.PlanISDN,
    NewLMSI:            gsmmap.HexBytes{0x11, 0x22, 0x33, 0x44},
    ReattachRequired:   true,
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
    MSISDNNature:        address.NatureInternational,
    MSISDNPlan:          address.PlanISDN,
    InterrogationType:   gsmmap.InterrogationBasicCall,
    GmscOrGsmSCFAddress: "31201111111",
    GmscOrGsmSCFAddressNature:          address.NatureInternational,
    GmscOrGsmSCFAddressPlan:            address.PlanISDN,
}
data, err := sri.Marshal()
if err != nil {
    log.Fatal(err)
}
_ = data // send the request bytes to the network

// Parse an SRI response received from the network
var respBytes []byte // BER response bytes received from the network
resp, err := gsmmap.ParseSriResp(respBytes)
if err != nil {
    log.Fatal(err)
}
if resp.NumberPortabilityStatus != nil {
    fmt.Println(*resp.NumberPortabilityStatus) // ownNumberPortedOut
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
D-CSI, and BCSM-CAMEL-TDP criteria lists). Every field of these types has
a Go field except extensionContainer, which the package does not surface.
Receiver mappings can normalize enum values, and ignored or unsurfaced wire
fields are dropped on Parse, so marshalled bytes need not match the input.

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
                            GsmSCFAddressNature:        address.NatureInternational,
                            GsmSCFAddressPlan:          address.PlanISDN,
                            DefaultCallHandling:        gsmmap.DefaultCallHandlingContinueCall,
                        },
                    },
                    CamelCapabilityHandling: &phase,
                },
            },
        },
    },
}
data, err := resp.Marshal()
if err != nil {
    log.Fatal(err)
}
_ = data // send the response bytes
```

### USSD (opCodes 59, 60, 61)

The three USSD operations of 3GPP TS 29.002 share one argument and one result
type ([§17.7.4](https://www.3gpp.org/ftp/Specs/archive/29_series/29.002/)):

- **processUnstructuredSS-Request** (59, §11.9): mobile-initiated USSD. Argument `USSDArg`, result `USSDRes`. `MSISDN` applies to this operation only.
- **unstructuredSS-Request** (60, §11.10): network-initiated request for input. Argument `USSDArg`, result `USSDRes` (which the peer may omit). `AlertingPattern` applies to 60 and 61 only.
- **unstructuredSS-Notify** (61, §11.11): network-initiated notification. Argument `USSDArg`; the result has no parameter.

All three use the application context networkUnstructuredSsContext-v2
(0.4.0.0.1.0.19.2, §17.3.2.20), returned by `NetworkUnstructuredSsContextV2()`
as a `[]uint64`.

```go
// A USSD gateway receiving processUnstructuredSS-Request (opCode 59)
var invokeParameter []byte // BER invoke parameter received from the network
arg, err := gsmmap.ParseUSSDArg(invokeParameter)
if err != nil {
    log.Fatal(err)
}
text, err := arg.DataCodingScheme.Decode(arg.USSDString) // "*100#"
if err != nil {
    // e.g. errors.Is(err, gsmmap.ErrUSSDUnsupportedDataCodingScheme):
    log.Fatal(err) // answer with MapErrorUnknownAlphabet in a TCAP gateway
}
fmt.Println(text)

// Answer with USSD-Res
reply, err := gsmmap.USSDDataCodingSchemeGSM7.Encode("Balance: 10.00 EUR")
if err != nil {
    log.Fatal(err)
}
res := &gsmmap.USSDRes{
    DataCodingScheme: gsmmap.USSDDataCodingSchemeGSM7,
    USSDString:       reply,
}
resultParameter, err := res.Marshal()
if err != nil {
    log.Fatal(err)
}
_ = resultParameter // send the result bytes
```

`USSDDataCodingScheme` interprets the data coding scheme as the Cell
Broadcast Data Coding Scheme of 3GPP TS 23.038 §5, as TS 29.002 §7.6.4.36
requires:

| Coding | Decode | Encode |
|---|---|---|
| GSM 7 bit default alphabet (`0000 xxxx`, e.g. `0x0F`; `0010 0000`–`0010 0100`; `010x 00xx`; `1111 00xx`) | yes, with the USSD packing of TS 23.038 §6.1.2.3.1 (a final `<CR>` pad is removed) and the extension table | yes |
| UCS2 (`010x 10xx`, e.g. `0x48`) | yes | yes, for characters up to U+FFFF |
| Reserved codings (e.g. `0010 0101`–`0011 1111`, `1111 1xxx`) | as GSM 7 bit, which §5 requires of a receiving entity | no (a sender must not use them) |
| Language indication (`0x10`, `0x11`, `0x12`), compressed, 8 bit data, UDH, I1, WAP | `ErrUSSDUnsupportedDataCodingScheme` | `ErrUSSDUnsupportedDataCodingScheme` |

The 7 bit packing and the character tables come from go-sms
(`encoding/gsm7`, `encoding/ucs2`).

### MAP errors

`MapErrorCode` names the MAP local error codes (`String()` gives the ASN.1
name, e.g. `"ussd-Busy"`), and `ParseReturnErrorParameter` decodes the
parameter of a TCAP ReturnError into the error's parameter type:

```go
var errorCode int64 // TCAP ReturnError local error code
var parameter []byte // BER parameter received with the error
p, err := gsmmap.ParseReturnErrorParameter(gsmmap.MapErrorCode(errorCode), parameter)
if err != nil {
    log.Fatal(err)
}
switch v := p.(type) {
case *gsmmap.SystemFailureParam:
    fmt.Println(v)
case *gsmmap.UnexpectedDataParam:
    fmt.Println(v.UnexpectedSubscriber)
}
```

It returns `(nil, nil)` for errors without a parameter (unknownAlphabet 71,
ussd-Busy 72), for codes it does not decode, and for an empty parameter.

## Design

This library provides a **layered API**:

- **Public types** (`SriSm`, `MtFsm`, etc.) use plain Go types — strings for phone numbers, bools for flags, `tpdu.TPDU` for SMS data.
- **Internally**, these are converted to/from [go-asn1](https://github.com/gomaja/go-asn1)'s generated `gsm_map.*` structs for BER encoding.
- **OpCode constants** can be imported directly from `github.com/gomaja/go-asn1/telecom/ss7/gsm_map` if needed for TCAP integration.
- **Validation** of ASN.1 constraints (sizes, including SEQUENCE OF elements, integer ranges, and non-extensible ENUMERATED values) and BIT STRING length consistency is done by go-asn1's strict BER codec on encode and decode. This package also checks TS 23.003 identity digit counts and TS 29.002 semantics (presence rules, sender rules, receiver mappings, and the TBCD alphabet).
- **Encoding.** `Parse` accepts any BER of ITU-T X.690, including constructed strings and indefinite or long-form lengths. `Marshal` follows TS 29.002 §17.1.1: OCTET and BIT STRINGs in the primitive form, and lengths in the definite form, short below 128 octets and otherwise in the fewest octets. The opaque PDP-ContextInfo and PagingArea entries are returned by `Parse` and sent by `Marshal` in DER, whatever BER they arrive in. X.680 §22.7 lets encoding rules add or remove trailing 0 bits of a named BIT STRING: go-asn1 pads a received one shorter than its SIZE lower bound with 0 bits, so `Parse` accepts it and `Marshal` sends it at that bound, but it rejects one padded with 0 bits beyond its SIZE upper bound ([go-asn1#129](https://github.com/gomaja/go-asn1/issues/129)).

### Address handling

Phone numbers are stored as digit strings in the TBCD alphabet of TS 29.002 §17.7.8 (`0-9 * # a b c`). The nature of address and numbering plan are companion fields (e.g., `MSISDNNature`, `MSISDNPlan`) holding the `address.Nature*` / `address.Plan*` values. Zero is unknown, exactly as on the wire. Parse preserves the decoded digits, nature, and plan, including reserved assignments. Marshal rejects `address.NatureReserved` and `address.NatureReservedExtension` with `ErrAddressNatureInvalid`, and undefined numbering plans, including `address.PlanReservedExtension`, with `ErrAddressPlanInvalid`. Remarshalling an accepted address drops a trailing all-filler octet (for example, `91 21 43 FF` becomes `91 21 43`) and sets an address extension bit of 0 to 1. Set the nature and plan explicitly when building a message.

### Sub-packages

| Package | Purpose |
|---|---|
| `tbcd` | TBCD-STRING encoding/decoding per TS 29.002 §17.7.8 |
| `address` | MAP AddressString encoding/decoding |
| `gsn` | GSN address (IPv4/IPv6) encoding per 3GPP TS 23.003 |

## Requirements

- Go 1.26.9 or later (the `go` directive in `go.mod`)
- The main branches of [gomaja/go-asn1](https://github.com/gomaja/go-asn1) and [gomaja/go-sms](https://github.com/gomaja/go-sms), which `go.mod` pins by commit

## License

MIT — see [LICENSE](LICENSE).
