package gsmmap

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/warthog618/sms/encoding/tpdu"
)

// GetErrorString converts a MAP error code to its string representation.
func GetErrorString(errCode int64) string {
	return gsm_map.ErrorCode(errCode).String()
}

// HexBytes is a []byte that marshals to/from hex strings in JSON
// instead of the default base64 encoding.
type HexBytes []byte

func (h HexBytes) MarshalJSON() ([]byte, error) {
	if h == nil {
		return []byte("null"), nil
	}
	return json.Marshal(hex.EncodeToString(h))
}

func (h *HexBytes) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*h = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	*h = b
	return nil
}

// SmDeliveryNotIntended per 3GPP TS 29.002.
type SmDeliveryNotIntended int

const (
	SmDeliveryOnlyIMSIRequested   SmDeliveryNotIntended = 0
	SmDeliveryOnlyMCCMNCRequested SmDeliveryNotIntended = 1
)

// SriSmCorrelationID corresponds to CorrelationID SEQUENCE.
type SriSmCorrelationID struct {
	HlrID   HexBytes // HLR-Id, optional
	SipUriA HexBytes // SIP-URI, optional
	SipUriB HexBytes // SIP-URI, mandatory within CorrelationID
}

// AdditionalNumber is the Additional-Number CHOICE.
// Set exactly one of MscNumber or SgsnNumber.
type AdditionalNumber struct {
	MscNumber       string
	MscNumberNature uint8
	MscNumberPlan   uint8
	SgsnNumber       string
	SgsnNumberNature uint8
	SgsnNumberPlan   uint8
}

// NetworkNodeDiameterAddress SEQUENCE.
type NetworkNodeDiameterAddress struct {
	DiameterName  HexBytes // DiameterIdentity
	DiameterRealm HexBytes // DiameterIdentity
}

// IpSmGwGuidance SEQUENCE (IP-SM-GW-Guidance).
type IpSmGwGuidance struct {
	MinimumDeliveryTimeValue     int // SM-DeliveryTimerValue (INTEGER 30..600)
	RecommendedDeliveryTimeValue int // SM-DeliveryTimerValue
}

// SriSm represents a Send Routing Info for Short Message request (opCode 45).
type SriSm struct {
	MSISDN               string
	MSISDNNature         uint8 // address nature indicator (default: International)
	MSISDNPlan           uint8 // numbering plan indicator (default: ISDN)
	SmRpPri              bool
	ServiceCentreAddress string
	SCANature            uint8 // address nature indicator (default: International)
	SCAPlan              uint8 // numbering plan indicator (default: ISDN)

	// Optional fields (post-extension marker).
	GprsSupportIndicator    bool                   // [7] NULL — SMS-GMSC supports receiving two numbers from HLR
	SmRpMti                 *int                   // [8] SM-RP-MTI: 0=SMS Deliver, 1=SMS Status Report (0..10)
	SmRpSmea                HexBytes               // [9] SM-RP-SMEA: 1..12 octets (address per 3GPP TS 23.040)
	SmDeliveryNotIntended   *SmDeliveryNotIntended  // [10] ENUMERATED
	IpSmGwGuidanceIndicator bool                   // [11] NULL
	IMSI                    string                 // [12] optional IMSI for delivery control
	SingleAttemptDelivery   bool                   // [13] NULL
	T4TriggerIndicator      bool                   // [14] NULL
	CorrelationID           *SriSmCorrelationID    // [15] SEQUENCE
	SmsfSupportIndicator    bool                   // [16] NULL
}

// SriSmResp represents a Send Routing Info for Short Message response.
type SriSmResp struct {
	IMSI                 string
	LocationInfoWithLMSI LocationInfoWithLMSI
	IpSmGwGuidance       *IpSmGwGuidance // [5] optional
}

// LocationInfoWithLMSI contains location information with LMSI.
type LocationInfoWithLMSI struct {
	NetworkNodeNumber       string
	NetworkNodeNumberNature uint8 // address nature indicator
	NetworkNodeNumberPlan   uint8 // numbering plan indicator
	LMSI                    HexBytes          // 4 octets; nil if absent
	GprsNodeIndicator       bool              // [5] NULL
	AdditionalNumber        *AdditionalNumber // [6] CHOICE
	NetworkNodeDiameterAddress           *NetworkNodeDiameterAddress // [7]
	AdditionalNetworkNodeDiameterAddress *NetworkNodeDiameterAddress // [8]
	ThirdNumber             *AdditionalNumber           // [9] CHOICE
	ThirdNetworkNodeDiameterAddress      *NetworkNodeDiameterAddress // [10]
	ImsNodeIndicator        bool              // [11] NULL
	Smsf3gppNumber          string            // [12]
	Smsf3gppNumberNature    uint8
	Smsf3gppNumberPlan      uint8
	Smsf3gppDiameterAddress *NetworkNodeDiameterAddress // [13]
	SmsfNon3gppNumber       string            // [14]
	SmsfNon3gppNumberNature uint8
	SmsfNon3gppNumberPlan   uint8
	SmsfNon3gppDiameterAddress *NetworkNodeDiameterAddress // [15]
	Smsf3gppAddressIndicator    bool // [16] NULL
	SmsfNon3gppAddressIndicator bool // [17] NULL
}

// MtFsm represents a Mobile Terminated Forward Short Message (opCode 44).
type MtFsm struct {
	IMSI                   string
	ServiceCentreAddressOA string
	SCAOANature            uint8 // address nature indicator (default: International)
	SCAOAPlan              uint8 // numbering plan indicator (default: ISDN)
	TPDU                   tpdu.TPDU
	MoreMessagesToSend     bool

	// Optional fields (post-extension marker).
	SmDeliveryTimer           *int                        // SM-DeliveryTimerValue: MinSmDeliveryTimer..MaxSmDeliveryTimer seconds
	SmDeliveryStartTime       HexBytes                    // Time octet string; nil if absent
	SmsOverIPOnlyIndicator    bool                        // [0] NULL
	CorrelationID             *SriSmCorrelationID         // [1] reuse SRI-SM type
	MaximumRetransmissionTime HexBytes                    // [2] Time octet string; nil if absent
	SmsGmscAddress            string                      // [3] ISDN-AddressString
	SmsGmscAddressNature      uint8
	SmsGmscAddressPlan        uint8
	SmsGmscDiameterAddress    *NetworkNodeDiameterAddress // [4]
}

// MtFsmResp represents a Mobile Terminated Forward Short Message response.
type MtFsmResp struct {
	SmRpUI HexBytes // optional SM-RP-UI (SignalInfo); nil if absent
}

// SmDeliveryOutcome per 3GPP TS 29.002.
type SmDeliveryOutcome int

const (
	SmDeliveryMemoryCapacityExceeded SmDeliveryOutcome = 0
	SmDeliveryAbsentSubscriber       SmDeliveryOutcome = 1
	SmDeliverySuccessfulTransfer     SmDeliveryOutcome = 2
)

// SmRpDa represents the SM-RP-DA CHOICE (destination address).
// Set exactly one field.
type SmRpDa struct {
	IMSI                   string   // [0] IMSI (TBCD)
	LMSI                   HexBytes // [1] 4 octets
	ServiceCentreAddressDA string   // [4] AddressString
	SCADANature            uint8
	SCADAPlan              uint8
	NoSmRpDa               bool // [5] NULL
}

// SmRpOa represents the SM-RP-OA CHOICE (originator address).
// Set exactly one field.
type SmRpOa struct {
	MSISDN                 string // ISDNAddressString
	MSISDNNature           uint8
	MSISDNPlan             uint8
	ServiceCentreAddressOA string // [4] AddressString
	SCAOANature            uint8
	SCAOAPlan              uint8
	NoSmRpOa               bool // [5] NULL
}

// MoFsm represents a Mobile Originated Forward Short Message (opCode 46).
type MoFsm struct {
	// SM-RP-DA: destination address CHOICE.
	// ServiceCentreAddressDA is the common variant. When SmRpDa is set it
	// overrides ServiceCentreAddressDA and allows any SM-RP-DA alternative
	// (IMSI, LMSI, serviceCentreAddressDA, noSM-RP-DA).
	ServiceCentreAddressDA string
	SCADANature            uint8 // address nature indicator (default: International)
	SCADAPlan              uint8 // numbering plan indicator (default: ISDN)
	SmRpDa                 *SmRpDa // when set, overrides ServiceCentreAddressDA

	// SM-RP-OA: originator address CHOICE.
	// MSISDN is the common variant. When SmRpOa is set it overrides MSISDN
	// and allows any SM-RP-OA alternative (msisdn, serviceCentreAddressOA,
	// noSM-RP-OA).
	MSISDN       string
	MSISDNNature uint8 // address nature indicator (default: International)
	MSISDNPlan   uint8 // numbering plan indicator (default: ISDN)
	SmRpOa       *SmRpOa // when set, overrides MSISDN

	TPDU tpdu.TPDU

	// Optional fields (post-extension marker).
	IMSI              string              // optional IMSI
	CorrelationID     *SriSmCorrelationID // [0] reuse SRI-SM type
	SmDeliveryOutcome *SmDeliveryOutcome  // [1]
}

// MoFsmResp represents a Mobile Originated Forward Short Message response.
type MoFsmResp struct {
	SmRpUI HexBytes // optional SM-RP-UI (SignalInfo); nil if absent
}

// UpdateLocation represents an UpdateLocation request.
type UpdateLocation struct {
	IMSI      string
	MSCNumber string
	MSCNature uint8 // address nature indicator (default: International)
	MSCPlan   uint8 // numbering plan indicator (default: ISDN)
	VLRNumber string
	VLRNature uint8 // address nature indicator (default: International)
	VLRPlan   uint8 // numbering plan indicator (default: ISDN)

	VlrCapability *VlrCapability
}

// VlrCapability contains VLR capability information.
type VlrCapability struct {
	SupportedCamelPhases       *SupportedCamelPhases
	SupportedLCSCapabilitySets *SupportedLCSCapabilitySets
}

// SupportedCamelPhases indicates which CAMEL phases are supported.
type SupportedCamelPhases struct {
	Phase1 bool
	Phase2 bool
	Phase3 bool
	Phase4 bool
}

// SupportedLCSCapabilitySets indicates which LCS capability sets are supported.
type SupportedLCSCapabilitySets struct {
	LcsCapabilitySet1 bool
	LcsCapabilitySet2 bool
	LcsCapabilitySet3 bool
	LcsCapabilitySet4 bool
	LcsCapabilitySet5 bool
}

// UpdateLocationRes represents an UpdateLocation response.
type UpdateLocationRes struct {
	HLRNumber       string
	HLRNumberNature uint8 // address nature indicator
	HLRNumberPlan   uint8 // numbering plan indicator
}

// UpdateGprsLocation represents an UpdateGprsLocation request.
type UpdateGprsLocation struct {
	IMSI        string
	SGSNNumber  string
	SGSNNature  uint8 // address nature indicator (default: International)
	SGSNPlan    uint8 // numbering plan indicator (default: ISDN)
	SGSNAddress string

	SGSNCapability *SGSNCapability
}

// SGSNCapability contains SGSN capability information.
type SGSNCapability struct {
	GprsEnhancementsSupportIndicator bool
	SupportedLCSCapabilitySets       *SupportedLCSCapabilitySets
}

// UpdateGprsLocationRes represents an UpdateGprsLocation response.
type UpdateGprsLocationRes struct {
	HLRNumber       string
	HLRNumberNature uint8 // address nature indicator
	HLRNumberPlan   uint8 // numbering plan indicator
}

// DomainType represents the requested domain.
type DomainType int

const (
	CsDomain DomainType = 0
	PsDomain DomainType = 1
)

// RequestedNodes indicates which network nodes are requested.
type RequestedNodes struct {
	MME  bool
	SGSN bool
}

// SubscriberIdentity represents the subscriber identity CHOICE.
// Set exactly one of IMSI or MSISDN.
type SubscriberIdentity struct {
	IMSI   string
	MSISDN string
}

// RequestedInfo represents the requested information flags for ATI.
type RequestedInfo struct {
	LocationInformation             bool
	SubscriberState                 bool
	CurrentLocation                 bool
	RequestedDomain                 *DomainType
	IMEI                            bool
	MsClassmark                     bool
	MnpRequestedInfo                bool
	LocationInformationEPSSupported bool
	TAdsData                        bool
	RequestedNodes                  *RequestedNodes
	ServingNodeIndication           bool
	LocalTimeZoneRequest            bool
}

// AnyTimeInterrogation represents an ATI request.
type AnyTimeInterrogation struct {
	SubscriberIdentity SubscriberIdentity
	RequestedInfo      RequestedInfo
	GsmSCFAddress      string
	GsmSCFNature       uint8 // address nature indicator (default: International)
	GsmSCFPlan         uint8 // numbering plan indicator (default: ISDN)
}

// AnyTimeInterrogationRes represents an ATI response.
type AnyTimeInterrogationRes struct {
	SubscriberInfo SubscriberInfo
}

// SubscriberInfo contains subscriber information returned by ATI.
type SubscriberInfo struct {
	LocationInformation     *CSLocationInformation
	SubscriberState         *SubscriberStateInfo
	LocationInformationEPS  *EPSLocationInformation
	LocationInformationGPRS *GPRSLocationInformation
	IMEI                    string // decoded TBCD; empty if absent
	MsClassmark2            HexBytes // raw octets; nil if absent
	TimeZone                HexBytes // raw octet; nil if absent
	DaylightSavingTime      *int   // nil if absent; 0=noAdjustment, 1=+1h, 2=+2h
}

// SubscriberStateInfo represents the subscriber state CHOICE.
type SubscriberStateInfo struct {
	State              SubscriberState
	NotReachableReason *int // set only when State == StateNetDetNotReachable
}

// SubscriberState enumerates subscriber state values.
type SubscriberState int

const (
	StateAssumedIdle        SubscriberState = 0
	StateCamelBusy          SubscriberState = 1
	StateNetDetNotReachable SubscriberState = 2
	StateNotProvidedFromVLR SubscriberState = 3
)

// NotReachableReason constants per 3GPP TS 29.002.
const (
	ReasonMsPurged       = 0
	ReasonImsiDetached   = 1
	ReasonRestrictedArea = 2
	ReasonNotRegistered  = 3
)

// CSLocationInformation contains CS domain location data.
type CSLocationInformation struct {
	AgeOfLocationInformation *int   // seconds; nil if absent
	VlrNumber                string // decoded; empty if absent
	VlrNumberNature          uint8
	VlrNumberPlan            uint8
	MscNumber                string // decoded; empty if absent
	MscNumberNature          uint8
	MscNumberPlan            uint8
	GeographicalInformation  *GeographicalInfo // decoded per 3GPP TS 23.032; nil if absent
	GeodeticInformation      HexBytes          // raw 10 octets; nil if absent
	CellGlobalId             HexBytes          // raw fixed-length cell ID or SAI; nil if absent
	LAI                      HexBytes          // raw 5-octet LAI; nil if absent
	LocationNumber           HexBytes          // raw octets; nil if absent
	CurrentLocationRetrieved bool
	SAIPresent               bool
}

// EPSLocationInformation contains EPS/LTE location data.
type EPSLocationInformation struct {
	AgeOfLocationInformation *int              // seconds; nil if absent
	EUtranCellGlobalIdentity HexBytes          // raw 7 octets; nil if absent
	TrackingAreaIdentity     HexBytes          // raw 5 octets; nil if absent
	GeographicalInformation  *GeographicalInfo // decoded per 3GPP TS 23.032; nil if absent
	GeodeticInformation      HexBytes          // raw 10 octets; nil if absent
	CurrentLocationRetrieved bool
	MmeName                  HexBytes          // raw DiameterIdentity; nil if absent
}

// GPRSLocationInformation contains GPRS domain location data.
type GPRSLocationInformation struct {
	AgeOfLocationInformation *int              // seconds; nil if absent
	CellGlobalId             HexBytes          // raw fixed-length cell ID or SAI; nil if absent
	LAI                      HexBytes          // raw 5-octet LAI; nil if absent
	RouteingAreaIdentity     HexBytes          // raw octets; nil if absent
	GeographicalInformation  *GeographicalInfo // decoded per 3GPP TS 23.032; nil if absent
	GeodeticInformation      HexBytes          // raw 10 octets; nil if absent
	SgsnNumber               string // decoded; empty if absent
	SgsnNumberNature         uint8
	SgsnNumberPlan           uint8
	CurrentLocationRetrieved bool
	SAIPresent               bool
}

// --- SendRoutingInfo (opCode 22) supporting types ---

// InterrogationType per 3GPP TS 29.002 clause 17.6.2.
type InterrogationType int

const (
	InterrogationBasicCall  InterrogationType = 0
	InterrogationForwarding InterrogationType = 1
)

// ForwardingReason per 3GPP TS 29.002.
type ForwardingReason int

const (
	ForwardingNotReachable ForwardingReason = 0
	ForwardingBusy         ForwardingReason = 1
	ForwardingNoReply      ForwardingReason = 2
)

// NumberPortabilityStatus per 3GPP TS 29.002.
type NumberPortabilityStatus int

const (
	MnpNotKnownToBePorted                  NumberPortabilityStatus = 0
	MnpOwnNumberPortedOut                  NumberPortabilityStatus = 1
	MnpForeignNumberPortedToForeignNetwork NumberPortabilityStatus = 2
	MnpOwnNumberNotPortedOut               NumberPortabilityStatus = 4
	MnpForeignNumberPortedIn               NumberPortabilityStatus = 5
)

// UnavailabilityCause per 3GPP TS 29.002.
type UnavailabilityCause int

const (
	UnavailBearerServiceNotProvisioned UnavailabilityCause = 1
	UnavailTeleserviceNotProvisioned   UnavailabilityCause = 2
	UnavailAbsentSubscriber            UnavailabilityCause = 3
	UnavailBusySubscriber              UnavailabilityCause = 4
	UnavailCallBarred                  UnavailabilityCause = 5
	UnavailCugReject                   UnavailabilityCause = 6
)

// SuppressMTSSFlags is the SuppressMTSS BIT STRING (bits 0..1 defined).
type SuppressMTSSFlags struct {
	SuppressCUG  bool
	SuppressCCBS bool
}

// AllowedServicesFlags is the AllowedServices BIT STRING.
type AllowedServicesFlags struct {
	FirstServiceAllowed  bool
	SecondServiceAllowed bool
}

// CugCheckInfo corresponds to CUG-CheckInfo SEQUENCE.
type CugCheckInfo struct {
	CugInterlock      HexBytes // CUG-Interlock, 4 octets
	CugOutgoingAccess bool
}

// ExtBasicServiceCode is the Ext-BasicServiceCode CHOICE.
// Set exactly one of ExtBearerService or ExtTeleservice.
type ExtBasicServiceCode struct {
	ExtBearerService HexBytes // Ext-BearerServiceCode, 1..5 octets
	ExtTeleservice   HexBytes // Ext-TeleserviceCode,   1..5 octets
}

// ExternalSignalInfo per ASN.1 SEQUENCE.
type ExternalSignalInfo struct {
	ProtocolID int      // ProtocolId (ENUMERATED: 0=gsm-0408, 1=gsm-0806, 2=gsm-BSSMAP, 3=ets-300102-1)
	SignalInfo HexBytes // octet string
}

// ExtExternalSignalInfo per ASN.1 SEQUENCE.
type ExtExternalSignalInfo struct {
	ExtProtocolID int // Ext-ProtocolId (1=ets-300356)
	SignalInfo    HexBytes
}

// SriCamelInfo mirrors the CamelInfo SEQUENCE used in SRI.
type SriCamelInfo struct {
	SupportedCamelPhases SupportedCamelPhases // reuses existing type
	SuppressTCSI         bool
	OfferedCamel4CSIs    *OfferedCamel4CSIs
}

// ExtendedRoutingInfo is the CHOICE ExtendedRoutingInfo.
// Set exactly one of RoutingInfo or CamelRoutingInfo.
type ExtendedRoutingInfo struct {
	RoutingInfo      *RoutingInfo
	CamelRoutingInfo *CamelRoutingInfo
}

// RoutingInfo is the CHOICE RoutingInfo.
// Set exactly one of RoamingNumber (non-empty) or ForwardingData (non-nil).
type RoutingInfo struct {
	RoamingNumber       string
	RoamingNumberNature uint8
	RoamingNumberPlan   uint8
	ForwardingData      *ForwardingData
}

// ForwardingData SEQUENCE.
type ForwardingData struct {
	ForwardedToNumber       string
	ForwardedToNumberNature uint8
	ForwardedToNumberPlan   uint8
	ForwardedToSubaddress   HexBytes
	ForwardingOptions       HexBytes // 1 octet
	LongForwardedToNumber   HexBytes // FTN-AddressString, opaque
}

// CamelRoutingInfo SEQUENCE.
type CamelRoutingInfo struct {
	ForwardingData            *ForwardingData
	GmscCamelSubscriptionInfo GmscCamelSubscriptionInfo
}

// GmscCamelSubscriptionInfo SEQUENCE.
// Nested CAMEL SEQUENCEs are kept opaque (HexBytes) for now; future work may decompose.
type GmscCamelSubscriptionInfo struct {
	TCSI                      HexBytes
	OCSI                      HexBytes
	DCSI                      HexBytes
	OBcsmCamelTDPCriteriaList HexBytes
	TBcsmCamelTDPCriteriaList HexBytes
}

// CcbsIndicators SEQUENCE.
type CcbsIndicators struct {
	CcbsPossible          bool
	KeepCCBSCallIndicator bool
}

// NaeaPreferredCI SEQUENCE (simplified: opaque CIC).
type NaeaPreferredCI struct {
	NaeaPreferredCIC HexBytes
}

// OfferedCamel4CSIs BIT STRING (7 bits defined).
type OfferedCamel4CSIs struct {
	OCSI            bool
	DCSI            bool
	VTCSI           bool
	TCSI            bool
	MTSMSCSI        bool
	MGCSI           bool
	PsiEnhancements bool
}

// SsCode is an SS-Code (single octet).
type SsCode uint8

// Sri represents a SendRoutingInfo (opCode 22) request.
type Sri struct {
	MSISDN       string
	MSISDNNature uint8
	MSISDNPlan   uint8

	InterrogationType   InterrogationType
	GmscOrGsmSCFAddress string
	GmscNature          uint8
	GmscPlan            uint8

	CugCheckInfo                    *CugCheckInfo
	NumberOfForwarding              *int
	OrInterrogation                 bool
	OrCapability                    *int
	CallReferenceNumber             HexBytes
	ForwardingReason                *ForwardingReason
	BasicServiceGroup               *ExtBasicServiceCode
	BasicServiceGroup2              *ExtBasicServiceCode
	NetworkSignalInfo               *ExternalSignalInfo
	NetworkSignalInfo2              *ExternalSignalInfo
	CamelInfo                       *SriCamelInfo
	SuppressionOfAnnouncement       bool
	AlertingPattern                 HexBytes
	CcbsCall                        bool
	SupportedCCBSPhase              *int
	AdditionalSignalInfo            *ExtExternalSignalInfo
	IstSupportIndicator             *int
	PrePagingSupported              bool
	CallDiversionTreatmentIndicator HexBytes
	LongFTNSupported                bool
	SuppressVTCSI                   bool
	SuppressIncomingCallBarring     bool
	GsmSCFInitiatedCall             bool
	SuppressMTSS                    *SuppressMTSSFlags
	MtRoamingRetrySupported         bool
	CallPriority                    *int
}

// SriResp represents a SendRoutingInfo response.
type SriResp struct {
	IMSI                            string
	ExtendedRoutingInfo             *ExtendedRoutingInfo
	CugCheckInfo                    *CugCheckInfo
	CugSubscriptionFlag             bool
	SubscriberInfo                  *SubscriberInfo // reuses existing ATI type
	SsList                          []SsCode
	BasicService                    *ExtBasicServiceCode
	BasicService2                   *ExtBasicServiceCode
	ForwardingInterrogationRequired bool
	VmscAddress                     string
	VmscNature                      uint8
	VmscPlan                        uint8
	NaeaPreferredCI                 *NaeaPreferredCI
	CcbsIndicators                  *CcbsIndicators
	MSISDN                          string
	MSISDNNature                    uint8
	MSISDNPlan                      uint8
	NumberPortabilityStatus         *NumberPortabilityStatus
	IstAlertTimer                   *int
	SupportedCamelPhasesInVMSC      *SupportedCamelPhases // reuses existing type
	OfferedCamel4CSIsInVMSC         *OfferedCamel4CSIs
	RoutingInfo2                    *RoutingInfo
	SsList2                         []SsCode
	AllowedServices                 *AllowedServicesFlags
	UnavailabilityCause             *UnavailabilityCause
	ReleaseResourcesSupported       bool
	GsmBearerCapability             *ExternalSignalInfo
}

// SM-DeliveryTimerValue range per 3GPP TS 29.002.
const (
	MinSmDeliveryTimer = 30
	MaxSmDeliveryTimer = 600
)

// MAP operation sentinel errors.
var (
	ErrSriMissingMSISDN              = errors.New("sri: MSISDN is empty")
	ErrSriMissingGmsc                = errors.New("sri: GmscOrGsmSCFAddress is empty")
	ErrSriInvalidInterrogationType   = errors.New("sri: InterrogationType must be 0 or 1")
	ErrSriInvalidNumberOfForwarding  = errors.New("sri: NumberOfForwarding must be 1..5")
	ErrSriInvalidOrCapability        = errors.New("sri: OrCapability must be 1..127")
	ErrSriInvalidCallReferenceNumber = errors.New("sri: CallReferenceNumber, if set, must be 1..8 octets")
	ErrSriChoiceMultipleAlternatives = errors.New("sri: CHOICE has multiple alternatives set")
	ErrSriChoiceNoAlternative        = errors.New("sri: CHOICE has no alternative set")

	ErrSriSmMissingSipUriB             = errors.New("sriSm: CorrelationID.SipUriB is mandatory but empty")
	ErrSriSmInvalidDeliveryTimerValue  = errors.New("sriSm: SM-DeliveryTimerValue must be 30..600")

	ErrMtFsmInvalidDeliveryTimer = errors.New("mtFsm: SmDeliveryTimer must be 30..600")

	ErrMoFsmSmRpDaNoAlternative        = errors.New("moFsm: SmRpDa CHOICE has no alternative set")
	ErrMoFsmSmRpDaMultipleAlternatives = errors.New("moFsm: SmRpDa CHOICE has multiple alternatives set")
	ErrMoFsmSmRpOaNoAlternative        = errors.New("moFsm: SmRpOa CHOICE has no alternative set")
	ErrMoFsmSmRpOaMultipleAlternatives = errors.New("moFsm: SmRpOa CHOICE has multiple alternatives set")
)
