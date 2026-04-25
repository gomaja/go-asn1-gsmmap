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

// AddInfo corresponds to ADD-Info SEQUENCE (opCode 2).
type AddInfo struct {
	IMEISV                   string // IMEI (TBCD-decoded)
	SkipSubscriberDataUpdate bool
}

// SuperChargerInfo is the SuperChargerInfo CHOICE (opCode 2).
// Set exactly one: SendSubscriberData=true, or SubscriberDataStored (non-nil).
type SuperChargerInfo struct {
	SendSubscriberData   bool
	SubscriberDataStored HexBytes // AgeIndicator; nil if not this alternative
}

// SupportedRATTypes BIT STRING (5 bits defined) (opCode 2).
type SupportedRATTypes struct {
	UTRAN          bool // bit 0
	GERAN          bool // bit 1
	GAN            bool // bit 2
	IHSPAEvolution bool // bit 3
	EUTRAN         bool // bit 4
}

// UpdateLocation represents an UpdateLocation request (opCode 2).
type UpdateLocation struct {
	IMSI      string
	MSCNumber string
	MSCNature uint8 // address nature indicator (default: International)
	MSCPlan   uint8 // numbering plan indicator (default: ISDN)
	VLRNumber string
	VLRNature uint8 // address nature indicator (default: International)
	VLRPlan   uint8 // numbering plan indicator (default: ISDN)

	VlrCapability *VlrCapability

	// Optional fields.
	LMSI                        HexBytes                    // [10] 4 octets; nil if absent
	InformPreviousNetworkEntity bool                        // [11] NULL
	CsLCSNotSupportedByUE       bool                        // [12] NULL
	VGmlcAddress                string                      // [2] GSN-Address; empty if absent
	AddInfo                     *AddInfo                    // [13]
	PagingArea                  []HexBytes                  // [14] list of LocationArea (kept opaque)
	SkipSubscriberDataUpdate    bool                        // [15] NULL
	RestorationIndicator        bool                        // [16] NULL
	EplmnList                   []HexBytes                  // [3] list of PLMNId (3 octets each)
	MmeDiameterAddress          *NetworkNodeDiameterAddress // [4]
}

// VlrCapability contains VLR capability information (opCode 2).
type VlrCapability struct {
	SupportedCamelPhases       *SupportedCamelPhases       // [0]
	SupportedLCSCapabilitySets *SupportedLCSCapabilitySets // [5]

	SolsaSupportIndicator                      bool              // [2] NULL
	IstSupportIndicator                        *int              // [1] 0=basicISTSupported, 1=istCommandSupported
	SuperChargerSupportedInServingNetworkEntity *SuperChargerInfo // [3] CHOICE
	LongFTNSupported                           bool              // [4] NULL
	OfferedCamel4CSIs                          *OfferedCamel4CSIs // [6]
	SupportedRATTypesIndicator                 *SupportedRATTypes // [7]
	LongGroupIDSupported                       bool              // [8] NULL
	MtRoamingForwardingSupported               bool              // [9] NULL
	MsisdnLessOperationSupported               bool              // [10] NULL
	ResetIdsSupported                          bool              // [11] NULL
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

// UpdateLocationRes represents an UpdateLocation response (opCode 2).
type UpdateLocationRes struct {
	HLRNumber            string
	HLRNumberNature      uint8 // address nature indicator
	HLRNumberPlan        uint8 // numbering plan indicator
	AddCapability        bool  // NULL
	PagingAreaCapability bool  // [0] NULL
}

// UsedRatType per 3GPP TS 29.002 (opCode 23).
type UsedRatType int

const (
	UsedRatUTRAN          UsedRatType = 0
	UsedRatGERAN          UsedRatType = 1
	UsedRatGAN            UsedRatType = 2
	UsedRatIHSPAEvolution UsedRatType = 3
	UsedRatEUTRAN         UsedRatType = 4
	UsedRatNBIOT          UsedRatType = 5
)

// UeSrvccCapability per 3GPP TS 29.002 (opCode 23).
type UeSrvccCapability int

const (
	UeSrvccNotSupported UeSrvccCapability = 0
	UeSrvccSupported    UeSrvccCapability = 1
)

// SmsRegisterRequest per 3GPP TS 29.002 (opCode 23).
type SmsRegisterRequest int

const (
	SmsRegistrationRequired     SmsRegisterRequest = 0
	SmsRegistrationNotPreferred SmsRegisterRequest = 1
	SmsRegistrationNoPreference SmsRegisterRequest = 2
)

// EpsInfo is the EPS-Info CHOICE (opCode 23).
// Set exactly one alternative: either PdnGwUpdate (non-nil) or
// IsrInformationBits > 0 (IsrInformation carries the BIT STRING bytes).
type EpsInfo struct {
	PdnGwUpdate        *PdnGwUpdate
	IsrInformation     HexBytes // BIT STRING content
	IsrInformationBits int      // BitLength; 0 means unset
}

// PdnGwUpdate SEQUENCE (opCode 23).
type PdnGwUpdate struct {
	APN           HexBytes       // [0] optional
	PdnGwIdentity *PdnGwIdentity // [1] optional
	ContextID     *int           // [2] optional
}

// PdnGwIdentity SEQUENCE (opCode 23).
// Per spec at least one of the address variants (or PdnGwName) should be set.
type PdnGwIdentity struct {
	IPv4Address HexBytes // [0] 4 octets
	IPv6Address HexBytes // [1] 16 octets
	Name        HexBytes // [2] FQDN
}

// UpdateGprsLocation represents an UpdateGprsLocation request (opCode 23).
type UpdateGprsLocation struct {
	IMSI        string
	SGSNNumber  string
	SGSNNature  uint8 // address nature indicator (default: International)
	SGSNPlan    uint8 // numbering plan indicator (default: ISDN)
	SGSNAddress string

	SGSNCapability *SGSNCapability

	// Optional fields (post-extension marker).
	InformPreviousNetworkEntity    bool               // [1] NULL
	PsLCSNotSupportedByUE          bool               // [2] NULL
	VGmlcAddress                   string             // [3] GSN-Address (IP string)
	AddInfo                        *AddInfo           // [4]
	EpsInfo                        *EpsInfo           // [5] CHOICE
	ServingNodeTypeIndicator       bool               // [6] NULL
	SkipSubscriberDataUpdate       bool               // [7] NULL
	UsedRatType                    *UsedRatType       // [8]
	GprsSubscriptionDataNotNeeded  bool               // [9] NULL
	NodeTypeIndicator              bool               // [10] NULL
	AreaRestricted                 bool               // [11] NULL
	UeReachableIndicator           bool               // [12] NULL
	EpsSubscriptionDataNotNeeded   bool               // [13] NULL
	UeSrvccCapability              *UeSrvccCapability // [14]
	EplmnList                      []HexBytes         // [15] list of 3-octet PLMNIds
	MmeNumberForMTSMS              string             // [16] ISDN-AddressString
	MmeNumberForMTSMSNature        uint8
	MmeNumberForMTSMSPlan          uint8
	SmsRegisterRequest             *SmsRegisterRequest // [17]
	SmsOnly                        bool                // [18] NULL
	SgsnName                       HexBytes            // [19] DiameterIdentity
	SgsnRealm                      HexBytes            // [20] DiameterIdentity
	LgdSupportIndicator            bool                // [21] NULL
	RemovalofMMERegistrationforSMS bool                // [22] NULL
	AdjacentPLMNList               []HexBytes          // [23] list of 3-octet PLMNIds
}

// SGSNCapability indicates SGSN capabilities per 3GPP TS 29.002 (opCode 23).
type SGSNCapability struct {
	SolsaSupportIndicator                       bool              // untagged (first field) NULL
	SuperChargerSupportedInServingNetworkEntity *SuperChargerInfo // [2] CHOICE
	GprsEnhancementsSupportIndicator            bool              // [3] NULL
	SupportedCamelPhases                        *SupportedCamelPhases
	SupportedLCSCapabilitySets                  *SupportedLCSCapabilitySets
	OfferedCamel4CSIs                           *OfferedCamel4CSIs
	SmsCallBarringSupportIndicator              bool // [7] NULL
	SupportedRATTypesIndicator                  *SupportedRATTypes
	SupportedFeatures                           HexBytes // raw BIT STRING bytes [9]
	SupportedFeaturesBits                       int      // BitLength; 0 means unset
	TAdsDataRetrieval                           bool     // [10] NULL
	HomogeneousSupportOfIMSVoiceOverPSSessions  *bool    // [11] 3-state
	CancellationTypeInitialAttach               bool     // [12] NULL
	MsisdnLessOperationSupported                bool     // [14] NULL
	UpdateofHomogeneousSupportOfIMSVoiceOverPSSessions bool // [15] NULL
	ResetIdsSupported                           bool     // [16] NULL
	ExtSupportedFeatures                        HexBytes // raw BIT STRING bytes [17]
	ExtSupportedFeaturesBits                    int      // BitLength; 0 means unset
}

// UpdateGprsLocationRes represents an UpdateGprsLocation response (opCode 23).
type UpdateGprsLocationRes struct {
	HLRNumber       string
	HLRNumberNature uint8 // address nature indicator
	HLRNumberPlan   uint8 // numbering plan indicator

	AddCapability              bool // untagged NULL
	SgsnMmeSeparationSupported bool // [0] NULL
	MmeRegisteredforSMS        bool // [1] NULL
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

// AnyTimeInterrogation represents an ATI request (opCode 71).
type AnyTimeInterrogation struct {
	SubscriberIdentity SubscriberIdentity
	RequestedInfo      RequestedInfo
	GsmSCFAddress      string
	GsmSCFNature       uint8 // address nature indicator (default: International)
	GsmSCFPlan         uint8 // numbering plan indicator (default: ISDN)
}

// AnyTimeInterrogationRes represents an ATI response (opCode 71).
type AnyTimeInterrogationRes struct {
	SubscriberInfo SubscriberInfo
}

// ProvideSubscriberInfo represents a ProvideSubscriberInfo request (opCode 70).
// PSI queries subscriber info (location, state, etc.) given an IMSI — similar
// to ATI but keyed by IMSI+LMSI rather than by MSISDN/IMSI identity.
type ProvideSubscriberInfo struct {
	IMSI          string        // mandatory (TBCD)
	LMSI          HexBytes      // [1] 4 octets; nil if absent
	RequestedInfo RequestedInfo // mandatory (reuses ATI RequestedInfo)
	CallPriority  *int          // [4] EMLPP-Priority 0..15; nil if absent
}

// ProvideSubscriberInfoRes represents a ProvideSubscriberInfo response (opCode 70).
type ProvideSubscriberInfoRes struct {
	SubscriberInfo SubscriberInfo // mandatory (reuses ATI SubscriberInfo)
}

// SubscriberInfo contains subscriber information returned by ATI (opCode 71).
type SubscriberInfo struct {
	LocationInformation     *CSLocationInformation  // [0]
	SubscriberState         *SubscriberStateInfo    // [1]
	LocationInformationGPRS *GPRSLocationInformation // [3]
	PsSubscriberState       *PsSubscriberState      // [4] CHOICE
	IMEI                    string                  // [5] decoded TBCD; empty if absent
	MsClassmark2            HexBytes                // [6] raw octets; nil if absent
	GprsMSClass             *GprsMSClass            // [7]
	MnpInfoRes              *MnpInfoRes             // [8]
	ImsVoiceOverPSSessionsIndication *ImsVoiceOverPSSessionsIndication // [9]
	LastUEActivityTime      HexBytes                // [10] Time octet string; nil if absent
	LastRATType             *UsedRatType            // [11]
	EpsSubscriberState      *PsSubscriberState      // [12] CHOICE
	LocationInformationEPS  *EPSLocationInformation // [13]
	TimeZone                HexBytes                // [14] raw octet; nil if absent
	DaylightSavingTime      *int                    // [15] nil if absent; 0=noAdjustment, 1=+1h, 2=+2h
	LocationInformation5GS  *LocationInformation5GS // [16]
}

// PsSubscriberState is the PS-SubscriberState CHOICE (opCode 71).
// Set exactly one alternative. The PDP-ContextInfoList alternatives carry
// opaque BER-encoded gsm_map.PDPContextInfo bytes.
type PsSubscriberState struct {
	NotProvidedFromSGSNorMME         bool       // [0] NULL
	PsDetached                       bool       // [1] NULL
	PsAttachedNotReachableForPaging  bool       // [2] NULL
	PsAttachedReachableForPaging     bool       // [3] NULL
	PsPDPActiveNotReachableForPaging []HexBytes // [4] opaque PDP-ContextInfoList
	PsPDPActiveReachableForPaging    []HexBytes // [5] opaque PDP-ContextInfoList
	NetDetNotReachable               *int       // untagged, NotReachableReason
}

// MnpInfoRes is the MNPInfoRes SEQUENCE (opCode 71, number portability result).
type MnpInfoRes struct {
	RouteingNumber          HexBytes                 // [0] raw bytes
	IMSI                    string                   // [1] TBCD-decoded
	MSISDN                  string                   // [2] ISDN
	MSISDNNature            uint8                    // address nature indicator
	MSISDNPlan              uint8                    // numbering plan indicator
	NumberPortabilityStatus *NumberPortabilityStatus // [3] enum
}

// ImsVoiceOverPSSessionsIndication per 3GPP TS 29.002 (opCode 71).
type ImsVoiceOverPSSessionsIndication int

const (
	IMSVoiceOverPSNotSupported ImsVoiceOverPSSessionsIndication = 0
	IMSVoiceOverPSSupported    ImsVoiceOverPSSessionsIndication = 1
	IMSVoiceOverPSUnknown      ImsVoiceOverPSSessionsIndication = 2
)

// GprsMSClass is the GPRSMSClass SEQUENCE (opCode 71).
type GprsMSClass struct {
	MSNetworkCapability     HexBytes // [0] mandatory
	MSRadioAccessCapability HexBytes // [1] optional
}

// UserCSGInformation is the UserCSGInformation SEQUENCE (opCode 71).
type UserCSGInformation struct {
	CsgID      HexBytes // [0] CSG-Id BIT STRING (raw bytes)
	CsgIDBits  int      // BitLength for the BIT STRING
	AccessMode HexBytes // [2]
	CMI        HexBytes // [3]
}

// LocationInformation5GS is the LocationInformation5GS SEQUENCE (opCode 71).
type LocationInformation5GS struct {
	NrCellGlobalIdentity     HexBytes          // [0]
	EUtranCellGlobalIdentity HexBytes          // [1]
	GeographicalInformation  *GeographicalInfo // [2]
	GeodeticInformation      HexBytes          // [3]
	AmfAddress               HexBytes          // [4] FQDN
	TrackingAreaIdentity     HexBytes          // [5]
	CurrentLocationRetrieved bool              // [6] NULL
	AgeOfLocationInformation *int              // [7]
	VplmnID                  HexBytes          // [8] 3 octets
	LocalTimeZone            HexBytes          // [9]
	RatType                  *UsedRatType      // [10]
	NrTrackingAreaIdentity   HexBytes          // [12]
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

// CSLocationInformation contains CS domain location data (opCode 71).
type CSLocationInformation struct {
	AgeOfLocationInformation *int   // seconds; nil if absent
	VlrNumber                string // decoded; empty if absent
	VlrNumberNature          uint8
	VlrNumberPlan            uint8
	MscNumber                string // decoded; empty if absent
	MscNumberNature          uint8
	MscNumberPlan            uint8
	GeographicalInformation  *GeographicalInfo   // decoded per 3GPP TS 23.032; nil if absent
	GeodeticInformation      HexBytes            // raw 10 octets; nil if absent
	CellGlobalId             HexBytes            // raw fixed-length cell ID or SAI; nil if absent
	LAI                      HexBytes            // raw 5-octet LAI; nil if absent
	LocationNumber           HexBytes            // raw octets; nil if absent
	SelectedLSAId            HexBytes            // [5] LSAIdentity; nil if absent
	UserCSGInformation       *UserCSGInformation // [11]
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

// GPRSLocationInformation contains GPRS domain location data (opCode 71).
type GPRSLocationInformation struct {
	AgeOfLocationInformation *int                // seconds; nil if absent
	CellGlobalId             HexBytes            // raw fixed-length cell ID or SAI; nil if absent
	LAI                      HexBytes            // raw 5-octet LAI; nil if absent
	RouteingAreaIdentity     HexBytes            // raw octets; nil if absent
	GeographicalInformation  *GeographicalInfo   // decoded per 3GPP TS 23.032; nil if absent
	GeodeticInformation      HexBytes            // raw 10 octets; nil if absent
	SgsnNumber               string              // decoded; empty if absent
	SgsnNumberNature         uint8
	SgsnNumberPlan           uint8
	SelectedLSAIdentity      HexBytes            // [4] LSAIdentity; nil if absent
	UserCSGInformation       *UserCSGInformation // [10]
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

// OBcsmTriggerDetectionPoint per 3GPP TS 29.002. Subset of values used in
// the MAP CAMEL subscription info; additional TDPs exist in CAP itself.
type OBcsmTriggerDetectionPoint int

const (
	OBcsmTriggerCollectedInfo      OBcsmTriggerDetectionPoint = 2
	OBcsmTriggerRouteSelectFailure OBcsmTriggerDetectionPoint = 4
)

// TBcsmTriggerDetectionPoint per 3GPP TS 29.002.
type TBcsmTriggerDetectionPoint int

const (
	TBcsmTriggerTermAttemptAuthorized TBcsmTriggerDetectionPoint = 12
	TBcsmTriggerTBusy                 TBcsmTriggerDetectionPoint = 13
	TBcsmTriggerTNoAnswer             TBcsmTriggerDetectionPoint = 14
)

// DefaultCallHandling per 3GPP TS 29.002.
type DefaultCallHandling int

const (
	DefaultCallHandlingContinueCall DefaultCallHandling = 0
	DefaultCallHandlingReleaseCall  DefaultCallHandling = 1
)

// CallTypeCriteria per 3GPP TS 29.002 (O-BcsmCamelTDP-Criteria).
type CallTypeCriteria int

const (
	CallTypeCriteriaForwarded    CallTypeCriteria = 0
	CallTypeCriteriaNotForwarded CallTypeCriteria = 1
)

// MatchType per 3GPP TS 29.002 (DestinationNumberCriteria).
type MatchType int

const (
	MatchTypeInhibiting MatchType = 0
	MatchTypeEnabling   MatchType = 1
)

// DestinationNumberCriteria per 3GPP TS 29.002.
// At least one of DestinationNumberList or DestinationNumberLengthList must
// be present when this criteria SEQUENCE is set.
type DestinationNumberCriteria struct {
	MatchType                   MatchType    // mandatory
	DestinationNumberList       []ISDNNumber // [1] list of destination numbers
	DestinationNumberLengthList []int        // [2] list of number lengths (1..15)
}

// ISDNNumber represents an ISDN-AddressString with its nature/plan indicators.
// Reused for DestinationNumberList entries in CAMEL criteria.
type ISDNNumber struct {
	Digits string
	Nature uint8 // default: International
	Plan   uint8 // default: ISDN
}

// OBcsmCamelTDPData per 3GPP TS 29.002. Originating BCSM CAMEL Trigger
// Detection Point descriptor — each entry pairs a trigger with the gsmSCF
// address to notify and a default call-handling action.
type OBcsmCamelTDPData struct {
	OBcsmTriggerDetectionPoint OBcsmTriggerDetectionPoint // mandatory
	ServiceKey                 int64                      // mandatory (0..2147483647)
	GsmSCFAddress              string                     // mandatory ISDN-AddressString
	GsmSCFAddressNature        uint8                      // default: International
	GsmSCFAddressPlan          uint8                      // default: ISDN
	DefaultCallHandling        DefaultCallHandling        // mandatory
}

// OCSI (O-CSI) per 3GPP TS 29.002. Originating CAMEL Subscription Info.
type OCSI struct {
	OBcsmCamelTDPDataList   []OBcsmCamelTDPData // mandatory, 1..10 entries
	CamelCapabilityHandling *int                // [0] phase (1..4); nil if absent
	NotificationToCSE       bool                // [1] NULL
	CsiActive               bool                // [2] NULL
}

// OBcsmCamelTDPCriteria per 3GPP TS 29.002. Selection criteria for an
// O-BcsmCamelTDP invocation. Empty optional fields are omitted on the wire.
type OBcsmCamelTDPCriteria struct {
	OBcsmTriggerDetectionPoint OBcsmTriggerDetectionPoint // mandatory
	DestinationNumberCriteria  *DestinationNumberCriteria // [0]
	BasicServiceCriteria       []ExtBasicServiceCode      // [1]
	CallTypeCriteria           *CallTypeCriteria          // [2]
	OCauseValueCriteria        []int                      // [3] list of CauseValue bytes (0..127)
}

// TBcsmCamelTDPData per 3GPP TS 29.002. Terminating BCSM CAMEL TDP descriptor.
type TBcsmCamelTDPData struct {
	TBcsmTriggerDetectionPoint TBcsmTriggerDetectionPoint // mandatory
	ServiceKey                 int64                      // mandatory
	GsmSCFAddress              string                     // mandatory ISDN-AddressString
	GsmSCFAddressNature        uint8                      // default: International
	GsmSCFAddressPlan          uint8                      // default: ISDN
	DefaultCallHandling        DefaultCallHandling        // mandatory
}

// TCSI (T-CSI) per 3GPP TS 29.002. Terminating CAMEL Subscription Info.
type TCSI struct {
	TBcsmCamelTDPDataList   []TBcsmCamelTDPData // mandatory, 1..10 entries
	CamelCapabilityHandling *int                // [0] phase (1..4); nil if absent
	NotificationToCSE       bool                // [1] NULL
	CsiActive               bool                // [2] NULL
}

// TBcsmCamelTDPCriteria per 3GPP TS 29.002. Selection criteria for a
// T-BCSM CAMEL TDP invocation.
type TBcsmCamelTDPCriteria struct {
	TBcsmTriggerDetectionPoint TBcsmTriggerDetectionPoint // mandatory
	BasicServiceCriteria       []ExtBasicServiceCode      // [0]
	TCauseValueCriteria        []int                      // [1] list of CauseValue bytes (0..127)
}

// DPAnalysedInfoCriterium per 3GPP TS 29.002. Entry in DCSI's
// DPAnalysedInfoCriteriaList — fires when a dialled number matches.
type DPAnalysedInfoCriterium struct {
	DialledNumber         string              // mandatory ISDN-AddressString
	DialledNumberNature   uint8               // default: International
	DialledNumberPlan     uint8               // default: ISDN
	ServiceKey            int64               // mandatory
	GsmSCFAddress         string              // mandatory
	GsmSCFAddressNature   uint8               // default: International
	GsmSCFAddressPlan     uint8               // default: ISDN
	DefaultCallHandling   DefaultCallHandling // mandatory
}

// DCSI (D-CSI) per 3GPP TS 29.002. Dialled-number CAMEL Subscription Info.
type DCSI struct {
	DPAnalysedInfoCriteriaList []DPAnalysedInfoCriterium // [0] 1..10 entries
	CamelCapabilityHandling    *int                      // [1] phase (1..4)
	NotificationToCSE          bool                      // [3] NULL
	CsiActive                  bool                      // [4] NULL
}

// GmscCamelSubscriptionInfo per 3GPP TS 29.002. Carries the CAMEL
// subscription information reported to the GMSC for call routing.
// Fields are typed SEQUENCEs with full field coverage; the lossy opaque
// HexBytes representation used in earlier versions has been replaced.
type GmscCamelSubscriptionInfo struct {
	TCSI                      *TCSI                   // [0]
	OCSI                      *OCSI                   // [1]
	OBcsmCamelTDPCriteriaList []OBcsmCamelTDPCriteria // [3]
	TBcsmCamelTDPCriteriaList []TBcsmCamelTDPCriteria // [4]
	DCSI                      *DCSI                   // [5]
}

// SSCSI (SS-CSI) per 3GPP TS 29.002 MAP-MS-DataTypes.asn:2254.
// Supplementary Service CAMEL Subscription Info.
//
// NotificationToCSE and CsiActive are spec-forbidden in messages sent
// toward the VLR; they're only legal in ATSI/ATM-ack/NSDC messages.
// The public API exposes them as bools for those cases.
type SSCSI struct {
	SsEventList       []SsCode // mandatory, 1..10 entries
	GsmSCFAddress     string   // mandatory ISDN-AddressString
	GsmSCFNature      uint8    // default: International
	GsmSCFPlan        uint8    // default: ISDN
	NotificationToCSE bool     // [0] NULL (ATSI/ATM/NSDC only)
	CsiActive         bool     // [1] NULL (ATSI/ATM/NSDC only)
}

// MCSI (M-CSI) per 3GPP TS 29.002 MAP-MS-DataTypes.asn:2517.
// Mobility-events CAMEL Subscription Info.
type MCSI struct {
	MobilityTriggers  []byte // mandatory 1..10 MM-Code octets (1 byte each)
	ServiceKey        int64  // mandatory 0..2147483647
	GsmSCFAddress     string // [0] mandatory ISDN-AddressString
	GsmSCFNature      uint8  // default: International
	GsmSCFPlan        uint8  // default: ISDN
	NotificationToCSE bool   // [2] NULL (ATSI/ATM/NSDC only)
	CsiActive         bool   // [3] NULL (ATSI/ATM/NSDC only)
}

// DefaultSMSHandling per 3GPP TS 29.002 MAP-MS-DataTypes.asn:2509.
// ENUMERATED { continueTransaction(0), releaseTransaction(1), ... }.
// Per spec exception handling, values 2..31 are treated as
// continueTransaction and values > 31 as releaseTransaction on decode —
// the decoder maps them accordingly and the encoder rejects anything
// outside 0..1.
type DefaultSMSHandling int

const (
	DefaultSMSHandlingContinueTransaction DefaultSMSHandling = 0
	DefaultSMSHandlingReleaseTransaction  DefaultSMSHandling = 1
)

// SMSTriggerDetectionPoint per 3GPP TS 29.002 MAP-MS-DataTypes.asn:2487.
// ENUMERATED { sms-CollectedInfo(1), sms-DeliveryRequest(2), ... }.
type SMSTriggerDetectionPoint int

const (
	SMSTriggerDetectionPointSmsCollectedInfo   SMSTriggerDetectionPoint = 1
	SMSTriggerDetectionPointSmsDeliveryRequest SMSTriggerDetectionPoint = 2
)

// SMSCAMELTDPData per 3GPP TS 29.002 MAP-MS-DataTypes.asn:2478.
// One SMS CAMEL trigger detection point entry.
type SMSCAMELTDPData struct {
	SmsTriggerDetectionPoint SMSTriggerDetectionPoint // [0] mandatory
	ServiceKey               int64                    // [1] mandatory 0..2147483647
	GsmSCFAddress            string                   // [2] mandatory ISDN-AddressString
	GsmSCFNature             uint8                    // default: International
	GsmSCFPlan               uint8                    // default: ISDN
	DefaultSMSHandling       DefaultSMSHandling       // [3] mandatory
}

// SMSCSI (SMS-CSI) per 3GPP TS 29.002 MAP-MS-DataTypes.asn:2458.
// Used for both mo-sms-CSI and mt-sms-CSI fields on the VLR.
//
// Per spec, SmsCAMELTDPDataList and CamelCapabilityHandling SHALL be
// present in an SMS-CSI sequence (spec clause 8.8.1). The encoder
// enforces that invariant.
type SMSCSI struct {
	SmsCAMELTDPDataList     []SMSCAMELTDPData // [0] mandatory 1..10 entries
	CamelCapabilityHandling *int              // [1] mandatory phase (1..4)
	NotificationToCSE       bool              // [3] NULL (ATSI/ATM/NSDC only)
	CsiActive               bool              // [4] NULL (ATSI/ATM/NSDC only)
}

// MTSMSTPDUType per 3GPP TS 29.002 MAP-MS-DataTypes.asn:2213.
// ENUMERATED { sms-DELIVER(0), sms-SUBMIT-REPORT(1), sms-STATUS-REPORT(2), ... }.
type MTSMSTPDUType int

const (
	MTSMSTPDUTypeSmsDELIVER      MTSMSTPDUType = 0
	MTSMSTPDUTypeSmsSUBMITREPORT MTSMSTPDUType = 1
	MTSMSTPDUTypeSmsSTATUSREPORT MTSMSTPDUType = 2
)

// MTSmsCAMELTDPCriteria per 3GPP TS 29.002 MAP-MS-DataTypes.asn:2202.
// Selection criteria for an MT-SMS CAMEL invocation.
type MTSmsCAMELTDPCriteria struct {
	SmsTriggerDetectionPoint SMSTriggerDetectionPoint // mandatory
	TpduTypeCriterion        []MTSMSTPDUType          // [0] optional, 1..5 entries when present
}

// VlrCamelSubscriptionInfo per 3GPP TS 29.002 MAP-MS-DataTypes.asn:2183.
// Full typed coverage; all 11 fields are exposed. Fields with dedicated
// ASN.1 CHOICE/SEQUENCE types delegate to their own domain structs.
//
// TifCSI is a NULL marker (spec-level "tif supported"); all other
// fields are optional by spec and nil-checked by the encoder.
type VlrCamelSubscriptionInfo struct {
	OCSI                      *OCSI                   // [0]
	SsCSI                     *SSCSI                  // [2]
	OBcsmCamelTDPCriteriaList []OBcsmCamelTDPCriteria // [4]
	TifCSI                    bool                    // [3] NULL
	MCSI                      *MCSI                   // [5]
	MoSmsCSI                  *SMSCSI                 // [6]
	VtCSI                     *TCSI                   // [7]
	TBcsmCamelTDPCriteriaList []TBcsmCamelTDPCriteria // [8]
	DCSI                      *DCSI                   // [9]
	MtSmsCSI                  *SMSCSI                 // [10]
	MtSmsCAMELTDPCriteriaList []MTSmsCAMELTDPCriteria // [11] 1..5 entries
}

// --- Ext-SS-Info (MAP-MS-DataTypes.asn:1826) — 5-alternative CHOICE ---

// MaxNumOfExtBasicServiceGroups is the spec upper bound on the various
// Ext-BasicServiceGroupList instances throughout TS 29.002 (asn:1942).
const MaxNumOfExtBasicServiceGroups = 32

// MaxNumOfCUG is the upper bound on CUG-SubscriptionList per TS 29.002
// (asn:1934). Note the lower bound is 0 (peer is allowed to send an
// empty CUG-SubscriptionList, unlike most other lists).
const MaxNumOfCUG = 10

// CliRestrictionOption per TS 29.002 MAP-SS-DataTypes.asn:177.
// ENUMERATED { permanent(0), temporaryDefaultRestricted(1),
// temporaryDefaultAllowed(2) }.
type CliRestrictionOption int

const (
	CliRestrictionPermanent                  CliRestrictionOption = 0
	CliRestrictionTemporaryDefaultRestricted CliRestrictionOption = 1
	CliRestrictionTemporaryDefaultAllowed    CliRestrictionOption = 2
)

// OverrideCategory per TS 29.002 MAP-SS-DataTypes.asn:182.
// ENUMERATED { overrideEnabled(0), overrideDisabled(1) }.
type OverrideCategory int

const (
	OverrideEnabled  OverrideCategory = 0
	OverrideDisabled OverrideCategory = 1
)

// SSSubscriptionOption is the SS-SubscriptionOption CHOICE
// (MAP-SS-DataTypes.asn:173). Set exactly one alternative.
type SSSubscriptionOption struct {
	CliRestriction *CliRestrictionOption // [2] cliRestrictionOption
	Override       *OverrideCategory     // [1] overrideCategory
}

// IntraCUGOptions per TS 29.002 MAP-MS-DataTypes.asn:1929.
// ENUMERATED { noCUG-Restrictions(0), cugIC-CallBarred(1),
// cugOG-CallBarred(2) }.
type IntraCUGOptions int

const (
	IntraCUGNoRestrictions IntraCUGOptions = 0
	IntraCUGICCallBarred   IntraCUGOptions = 1
	IntraCUGOGCallBarred   IntraCUGOptions = 2
)

// CUGSubscription per TS 29.002 MAP-MS-DataTypes.asn:1916.
type CUGSubscription struct {
	CugIndex              int                    // mandatory 0..32767
	CugInterlock          HexBytes               // mandatory, exactly 4 octets
	IntraCUGOptions       IntraCUGOptions        // mandatory
	BasicServiceGroupList []ExtBasicServiceCode  // optional, 1..32 entries when present
}

// CUGFeature per TS 29.002 MAP-MS-DataTypes.asn:1944.
type CUGFeature struct {
	BasicService          *ExtBasicServiceCode // optional
	PreferentialCUGIndex  *int                 // optional 0..32767
	InterCUGRestrictions  uint8                // mandatory; 1 octet bit-encoded per spec
}

// CUGInfo per TS 29.002 MAP-MS-DataTypes.asn:1907.
type CUGInfo struct {
	CugSubscriptionList []CUGSubscription // mandatory but spec allows SIZE(0..10) on the wire
	CugFeatureList      []CUGFeature      // optional, 1..32 entries when present
}

// ExtCallBarringFeature per TS 29.002 MAP-MS-DataTypes.asn:1901.
type ExtCallBarringFeature struct {
	BasicService *ExtBasicServiceCode // optional
	SsStatus     HexBytes             // [4] mandatory, 1..5 octets per Ext-SS-Status
}

// ExtCallBarInfo per TS 29.002 MAP-MS-DataTypes.asn:1892.
type ExtCallBarInfo struct {
	SsCode                 SsCode                  // mandatory
	CallBarringFeatureList []ExtCallBarringFeature // mandatory, 1..32 entries
}

// ExtForwFeature per TS 29.002 MAP-MS-DataTypes.asn:1842.
//
// ForwardedToNumber + ForwardingOptions + NoReplyConditionTime are all
// optional on the wire; the encoder writes whatever subset the caller
// populated. ForwardingOptions is 1..5 octets per spec; the encoder
// rejects anything outside that range. NoReplyConditionTime is 1..100;
// the lenient decoder maps 1..4 → 5 and 31..100 → 30 per spec exception
// handling.
type ExtForwFeature struct {
	BasicService          *ExtBasicServiceCode // optional
	SsStatus              HexBytes             // [4] mandatory, 1..5 octets per Ext-SS-Status
	ForwardedToNumber     string               // [5] optional ISDN-AddressString
	ForwardedToNature     uint8                // default: International
	ForwardedToPlan       uint8                // default: ISDN
	ForwardedToSubaddress HexBytes             // [8] optional ISDN-SubaddressString
	ForwardingOptions     HexBytes             // [6] optional 1..5 octets
	NoReplyConditionTime  *int                 // [7] optional 1..100 (post-decode normalised to 5..30)
	LongForwardedToNumber string               // [10] optional FTN-AddressString
}

// ExtForwInfo per TS 29.002 MAP-MS-DataTypes.asn:1833.
type ExtForwInfo struct {
	SsCode                SsCode           // mandatory
	ForwardingFeatureList []ExtForwFeature // mandatory, 1..32 entries
}

// ExtSSData per TS 29.002 MAP-MS-DataTypes.asn:1963.
type ExtSSData struct {
	SsCode                SsCode                // mandatory
	SsStatus              HexBytes              // [4] mandatory, 1..5 octets per Ext-SS-Status
	SsSubscriptionOption  *SSSubscriptionOption // optional CHOICE
	BasicServiceGroupList []ExtBasicServiceCode // optional, 1..32 entries when present
}

// EMLPPInfo per TS 29.002 MAP-CommonDataTypes.asn:607.
//
// Both priorities are EMLPP-Priority INTEGER (0..15); per spec exception
// handling, values 7..15 are spare and shall be mapped to 4 on decode.
// The encoder accepts anything 0..6 directly (the spec's mapped/named
// range — A=6, B=5, 0..4=0..4) and rejects 7..15 to avoid silently
// emitting a value the receiver will rewrite.
type EMLPPInfo struct {
	MaximumEntitledPriority int // mandatory 0..6 (post-mapping)
	DefaultPriority         int // mandatory 0..6 (post-mapping)
}

// ExtSSInfo is the Ext-SS-InfoList element CHOICE per TS 29.002
// MAP-MS-DataTypes.asn:1826. Exactly one alternative must be set.
type ExtSSInfo struct {
	ForwardingInfo  *ExtForwInfo    // [0] forwardingInfo
	CallBarringInfo *ExtCallBarInfo // [1] callBarringInfo
	CugInfo         *CUGInfo        // [2] cug-Info
	SsData          *ExtSSData      // [3] ss-Data
	EmlppInfo       *EMLPPInfo      // [4] emlpp-Info
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

// MwStatusFlags is the MW-Status BIT STRING (6 bits defined).
// Bit 0=scAddressNotIncluded, 1=mnrfSet, 2=mcefSet, 3=mnrgSet, 4=mnr5gSet, 5=mnr5gn3gSet.
// Per 3GPP TS 29.002, BIT STRING size is 6..16; bits 6-15 are reserved.
type MwStatusFlags struct {
	SCAddressNotIncluded bool
	MnrfSet              bool
	McefSet              bool
	MnrgSet              bool
	Mnr5gSet             bool
	Mnr5gn3gSet          bool
}

// SmsGmscAlertEvent per 3GPP TS 29.002 (opCode 64).
type SmsGmscAlertEvent int

const (
	SmsGmscAlertMsAvailableForMtSms   SmsGmscAlertEvent = 0
	SmsGmscAlertMsUnderNewServingNode SmsGmscAlertEvent = 1
)

// AlertServiceCentre represents an AlertServiceCentre request (opCode 64)
// per 3GPP TS 29.002. The response is an empty acknowledgement with no
// parameters (RETURN RESULT TRUE), so no response type is defined here.
type AlertServiceCentre struct {
	MSISDN               string // mandatory
	MSISDNNature         uint8  // default: International
	MSISDNPlan           uint8  // default: ISDN
	ServiceCentreAddress string // mandatory
	SCANature            uint8  // default: International
	SCAPlan              uint8  // default: ISDN

	// Optional fields (post-extension marker).
	IMSI                      string                      // optional IMSI (TBCD)
	CorrelationID             *SriSmCorrelationID         // SEQUENCE (reuses SRI-SM type)
	MaximumUeAvailabilityTime HexBytes                    // [0] Time octet string; nil if absent
	SmsGmscAlertEvent         *SmsGmscAlertEvent          // [1] ENUMERATED
	SmsGmscDiameterAddress    *NetworkNodeDiameterAddress // [2]
	NewSGSNNumber             string                      // [3] ISDN-AddressString
	NewSGSNNumberNature       uint8
	NewSGSNNumberPlan         uint8
	NewSGSNDiameterAddress    *NetworkNodeDiameterAddress // [4]
	NewMMENumber              string                      // [5] ISDN-AddressString
	NewMMENumberNature        uint8
	NewMMENumberPlan          uint8
	NewMMEDiameterAddress     *NetworkNodeDiameterAddress // [6]
	NewMSCNumber              string                      // [7] ISDN-AddressString
	NewMSCNumberNature        uint8
	NewMSCNumberPlan          uint8
}

// InformServiceCentre represents an InformServiceCentre request (opCode 63).
// This is a one-way MAP operation; no response is defined in 3GPP TS 29.002.
type InformServiceCentre struct {
	StoredMSISDN       string
	StoredMSISDNNature uint8 // address nature indicator (default: International)
	StoredMSISDNPlan   uint8 // numbering plan indicator (default: ISDN)

	MwStatus *MwStatusFlags // MW-Status BIT STRING (6 bits defined)

	AbsentSubscriberDiagnosticSM            *int // 0..255
	AdditionalAbsentSubscriberDiagnosticSM  *int // [0] 0..255
	Smsf3gppAbsentSubscriberDiagnosticSM    *int // [1] 0..255
	SmsfNon3gppAbsentSubscriberDiagnosticSM *int // [2] 0..255
}

// PurgeMS represents a PurgeMS request (opCode 67) per 3GPP TS 29.002.
// It is sent by the HLR to the VLR/SGSN to purge subscriber data when the
// subscriber has been deactivated or is permanently unreachable.
type PurgeMS struct {
	IMSI string // mandatory (TBCD)

	// Optional fields.
	VLRNumber  string // [0] ISDN-AddressString
	VLRNature  uint8  // address nature indicator (default: International)
	VLRPlan    uint8  // numbering plan indicator (default: ISDN)
	SGSNNumber string // [1] ISDN-AddressString
	SGSNNature uint8  // address nature indicator (default: International)
	SGSNPlan   uint8  // numbering plan indicator (default: ISDN)

	// Optional fields (post-extension marker).
	LocationInformation     *CSLocationInformation   // [2]
	LocationInformationGPRS *GPRSLocationInformation // [3]
	LocationInformationEPS  *EPSLocationInformation  // [4]
}

// PurgeMSRes represents a PurgeMS response (opCode 67) per 3GPP TS 29.002.
// The VLR/SGSN may reply with freeze-TMSI flags indicating which TMSIs the
// HLR should block.
type PurgeMSRes struct {
	FreezeTMSI  bool // [0] NULL
	FreezePTMSI bool // [1] NULL
	FreezeMTMSI bool // [2] NULL (post-extension marker)
}

// RequestingNodeType per 3GPP TS 29.002 (opCode 56).
type RequestingNodeType int

const (
	RequestingNodeVlr           RequestingNodeType = 0
	RequestingNodeSgsn          RequestingNodeType = 1
	RequestingNodeSCscf         RequestingNodeType = 2
	RequestingNodeBsf           RequestingNodeType = 3
	RequestingNodeGanAAAServer  RequestingNodeType = 4
	RequestingNodeWlanAAAServer RequestingNodeType = 5
	RequestingNodeMme           RequestingNodeType = 16
	RequestingNodeMmeSgsn       RequestingNodeType = 17
)

// ReSynchronisationInfo per 3GPP TS 29.002 (opCode 56).
// Carries re-synchronisation parameters produced by the USIM when a
// previously issued authentication vector is out of sequence.
type ReSynchronisationInfo struct {
	RAND HexBytes // 16 octets
	AUTS HexBytes // 14 octets
}

// AuthenticationTriplet is a 2G (GSM) authentication triplet (opCode 56).
type AuthenticationTriplet struct {
	RAND HexBytes // 16 octets
	SRES HexBytes // 4 octets
	Kc   HexBytes // 8 octets
}

// AuthenticationQuintuplet is a UMTS/3G authentication quintuplet (opCode 56).
type AuthenticationQuintuplet struct {
	RAND HexBytes // 16 octets
	XRES HexBytes // 4..16 octets
	CK   HexBytes // 16 octets
	IK   HexBytes // 16 octets
	AUTN HexBytes // 16 octets
}

// EpcAV is an LTE/EPS authentication vector (opCode 56).
type EpcAV struct {
	RAND  HexBytes // 16 octets
	XRES  HexBytes // 4..16 octets
	AUTN  HexBytes // 16 octets
	KASME HexBytes // 32 octets
}

// AuthenticationSetList is a CHOICE between 2G triplets and 3G quintuplets
// (opCode 56). Exactly one of Triplets or Quintuplets must be non-empty when
// the list is set; both empty/nil is treated as "no alternative" and both
// non-empty as "multiple alternatives" during encode. An explicitly-set
// empty slice counts as absent.
type AuthenticationSetList struct {
	Triplets    []AuthenticationTriplet
	Quintuplets []AuthenticationQuintuplet
}

// SendAuthenticationInfo represents a SendAuthenticationInfo request
// (opCode 56) per 3GPP TS 29.002. Sent by the VLR/SGSN/MME to the HLR/HSS
// to retrieve authentication vectors for subscriber authentication.
type SendAuthenticationInfo struct {
	IMSI                     string // mandatory (TBCD)
	NumberOfRequestedVectors int    // mandatory 1..5

	// Optional fields.
	SegmentationProhibited     bool                   // NULL
	ImmediateResponsePreferred bool                   // [1] NULL
	ReSynchronisationInfo      *ReSynchronisationInfo // SEQUENCE

	// Optional fields (post-extension marker).
	RequestingNodeType                 *RequestingNodeType // [3]
	RequestingPLMNId                   HexBytes            // [4] 3 octets
	NumberOfRequestedAdditionalVectors *int                // [5] 1..5
	AdditionalVectorsAreForEPS         bool                // [6] NULL
	UeUsageTypeRequestIndication       bool                // [7] NULL
}

// SendAuthenticationInfoRes represents a SendAuthenticationInfo response
// (opCode 56) per 3GPP TS 29.002.
type SendAuthenticationInfoRes struct {
	AuthenticationSetList    *AuthenticationSetList // CHOICE: Triplets | Quintuplets
	EpsAuthenticationSetList []EpcAV                // [2] EPS-AuthenticationSetList
	UeUsageType              HexBytes               // [3] UE-UsageType, exactly 4 octets
}

// CancellationType per 3GPP TS 29.002 (opCode 3). Indicates why the HLR
// is asking the VLR/SGSN to cancel the subscriber's location record.
type CancellationType int

const (
	CancellationTypeUpdateProcedure        CancellationType = 0
	CancellationTypeSubscriptionWithdraw   CancellationType = 1
	CancellationTypeInitialAttachProcedure CancellationType = 2
)

// TypeOfUpdate per 3GPP TS 29.002 (opCode 3). This field is only valid
// when CancellationType is updateProcedure or initialAttachProcedure.
// The constraint is enforced on both the encode and decode paths and
// returns ErrCancelLocTypeOfUpdateNotApplicable when violated.
type TypeOfUpdate int

const (
	TypeOfUpdateSgsnChange TypeOfUpdate = 0
	TypeOfUpdateMmeChange  TypeOfUpdate = 1
)

// CancelLocationIdentity is the CHOICE between IMSI alone and IMSI+LMSI.
// Exactly one of IMSI or IMSIWithLMSI must be set when encoding.
type CancelLocationIdentity struct {
	IMSI         string                   // alternative: imsi (TBCD)
	IMSIWithLMSI *CancelLocationIMSIWithLMSI // alternative: imsi-WithLMSI
}

// CancelLocationIMSIWithLMSI carries both an IMSI and the LMSI assigned
// by the VLR for that subscriber.
type CancelLocationIMSIWithLMSI struct {
	IMSI string   // mandatory (TBCD)
	LMSI HexBytes // mandatory, 4 octets
}

// CancelLocation represents a CancelLocation request (opCode 3) per
// 3GPP TS 29.002. It is sent by the HLR to the VLR/SGSN to remove the
// subscriber's location record — e.g. after a successful location update
// in another VLR, on subscription withdrawal, or on initial EPS attach.
type CancelLocation struct {
	Identity         CancelLocationIdentity // mandatory CHOICE
	CancellationType *CancellationType      // optional ENUMERATED

	// Optional fields (post-extension marker).
	TypeOfUpdate                  *TypeOfUpdate // [0] ENUMERATED
	MtrfSupportedAndAuthorized    bool          // [1] NULL
	MtrfSupportedAndNotAuthorized bool          // [2] NULL (mutually exclusive with the above)
	NewMSCNumber                  string        // [3] ISDN-AddressString
	NewMSCNumberNature            uint8         // address nature indicator (default: International)
	NewMSCNumberPlan              uint8         // numbering plan indicator (default: ISDN)
	NewVLRNumber                  string        // [4] ISDN-AddressString
	NewVLRNumberNature            uint8         // address nature indicator (default: International)
	NewVLRNumberPlan              uint8         // numbering plan indicator (default: ISDN)
	NewLMSI                       HexBytes      // [5] LMSI, 4 octets when present
	ReattachRequired              bool          // [6] NULL
}

// --- InsertSubscriberData (opCode 7) — foundation types ---

// SubscriberStatus per 3GPP TS 29.002 (MAP-MS-DataTypes.asn:1756).
// ENUMERATED { serviceGranted(0), operatorDeterminedBarring(1) }.
type SubscriberStatus int

const (
	SubscriberStatusServiceGranted            SubscriberStatus = 0
	SubscriberStatusOperatorDeterminedBarring SubscriberStatus = 1
)

// NetworkAccessMode per 3GPP TS 29.002 (MAP-MS-DataTypes.asn:1509).
// ENUMERATED { packetAndCircuit(0), onlyCircuit(1), onlyPacket(2) }.
type NetworkAccessMode int

const (
	NetworkAccessModePacketAndCircuit NetworkAccessMode = 0
	NetworkAccessModeOnlyCircuit      NetworkAccessMode = 1
	NetworkAccessModeOnlyPacket       NetworkAccessMode = 2
)

// RegionalSubscriptionResponse per 3GPP TS 29.002 (MAP-MS-DataTypes.asn:2091).
// ENUMERATED { networkNode-AreaRestricted(0), tooManyZoneCodes(1),
// zoneCodesConflict(2), regionalSubscNotSupported(3) }.
type RegionalSubscriptionResponse int

const (
	RegionalSubscriptionResponseNetworkNodeAreaRestricted RegionalSubscriptionResponse = 0
	RegionalSubscriptionResponseTooManyZoneCodes          RegionalSubscriptionResponse = 1
	RegionalSubscriptionResponseZoneCodesConflict         RegionalSubscriptionResponse = 2
	RegionalSubscriptionResponseRegionalSubscNotSupported RegionalSubscriptionResponse = 3
)

// ODBGeneralData (BIT STRING SIZE 15..32) per TS 29.002 MAP-MS-DataTypes.asn:1776.
// 29 named bits covering operator-determined barring of outgoing calls,
// explicit call transfer, packet-oriented services, and roaming. Unknown
// bits received from peers are treated as unsupported-ODB per spec
// exception handling.
type ODBGeneralData struct {
	AllOGCallsBarred                                                   bool // bit 0
	InternationalOGCallsBarred                                         bool // bit 1
	InternationalOGCallsNotToHPLMNCountryBarred                        bool // bit 2
	PremiumRateInformationOGCallsBarred                                bool // bit 3
	PremiumRateEntertainmentOGCallsBarred                              bool // bit 4
	SSAccessBarred                                                     bool // bit 5
	InterzonalOGCallsBarred                                            bool // bit 6
	InterzonalOGCallsNotToHPLMNCountryBarred                           bool // bit 7
	InterzonalOGCallsAndInternationalOGCallsNotToHPLMNCountryBarred    bool // bit 8
	AllECTBarred                                                       bool // bit 9
	ChargeableECTBarred                                                bool // bit 10
	InternationalECTBarred                                             bool // bit 11
	InterzonalECTBarred                                                bool // bit 12
	DoublyChargeableECTBarred                                          bool // bit 13
	MultipleECTBarred                                                  bool // bit 14
	AllPacketOrientedServicesBarred                                    bool // bit 15
	RoamerAccessToHPLMNAPBarred                                        bool // bit 16
	RoamerAccessToVPLMNAPBarred                                        bool // bit 17
	RoamingOutsidePLMNOGCallsBarred                                    bool // bit 18
	AllICCallsBarred                                                   bool // bit 19
	RoamingOutsidePLMNICCallsBarred                                    bool // bit 20
	RoamingOutsidePLMNICountryICCallsBarred                            bool // bit 21
	RoamingOutsidePLMNBarred                                           bool // bit 22
	RoamingOutsidePLMNCountryBarred                                    bool // bit 23
	RegistrationAllCFBarred                                            bool // bit 24
	RegistrationCFNotToHPLMNBarred                                     bool // bit 25
	RegistrationInterzonalCFBarred                                     bool // bit 26
	RegistrationInterzonalCFNotToHPLMNBarred                           bool // bit 27
	RegistrationInternationalCFBarred                                  bool // bit 28
}

// ODBHPLMNData (BIT STRING SIZE 4..32) per TS 29.002 MAP-MS-DataTypes.asn:1812.
// Carries HPLMN-specific ODB barring types. Unknown bits received from
// peers are treated as unsupported-ODB per spec exception handling.
type ODBHPLMNData struct {
	PLMNSpecificBarringType1 bool // bit 0
	PLMNSpecificBarringType2 bool // bit 1
	PLMNSpecificBarringType3 bool // bit 2
	PLMNSpecificBarringType4 bool // bit 3
}

// AccessRestrictionData (BIT STRING SIZE 2..8) per TS 29.002
// MAP-MS-DataTypes.asn:1454. Access-type restrictions applied to the
// subscriber. Per spec, nodes shall ignore restrictions for access types
// they do not support.
type AccessRestrictionData struct {
	UtranNotAllowed            bool // bit 0
	GeranNotAllowed            bool // bit 1
	GanNotAllowed              bool // bit 2
	IHSPAEvolutionNotAllowed   bool // bit 3
	WBEUtranNotAllowed         bool // bit 4
	HoToNon3GPPAccessNotAllowed bool // bit 5
	NBIoTNotAllowed            bool // bit 6
	EnhancedCoverageNotAllowed bool // bit 7
}

// ExtAccessRestrictionData (BIT STRING SIZE 1..32) per TS 29.002
// MAP-MS-DataTypes.asn:1471. Additional access-type restrictions that
// don't fit in the 8-bit AccessRestrictionData.
type ExtAccessRestrictionData struct {
	NrAsSecondaryRATNotAllowed                bool // bit 0
	UnlicensedSpectrumAsSecondaryRATNotAllowed bool // bit 1
}

// SupportedFeatures (BIT STRING SIZE 26..40) per TS 29.002
// MAP-MS-DataTypes.asn:642. HSS/HLR-advertised feature support;
// see 3GPP TS 29.272 for each bit's meaning.
type SupportedFeatures struct {
	OdbAllApn                                        bool // bit 0
	OdbHPLMNApn                                      bool // bit 1
	OdbVPLMNApn                                      bool // bit 2
	OdbAllOg                                         bool // bit 3
	OdbAllInternationalOg                            bool // bit 4
	OdbAllIntOgNotToHPLMNCountry                     bool // bit 5
	OdbAllInterzonalOg                               bool // bit 6
	OdbAllInterzonalOgNotToHPLMNCountry              bool // bit 7
	OdbAllInterzonalOgAndInternatOgNotToHPLMNCountry bool // bit 8
	RegSub                                           bool // bit 9
	Trace                                            bool // bit 10
	LcsAllPrivExcep                                  bool // bit 11
	LcsUniversal                                     bool // bit 12
	LcsCallSessionRelated                            bool // bit 13
	LcsCallSessionUnrelated                          bool // bit 14
	LcsPLMNOperator                                  bool // bit 15
	LcsServiceType                                   bool // bit 16
	LcsAllMOLRSS                                     bool // bit 17
	LcsBasicSelfLocation                             bool // bit 18
	LcsAutonomousSelfLocation                        bool // bit 19
	LcsTransferToThirdParty                          bool // bit 20
	SmMoPp                                           bool // bit 21
	BarringOutgoingCalls                             bool // bit 22
	Baoc                                             bool // bit 23
	Boic                                             bool // bit 24
	BoicExHC                                         bool // bit 25
	LocalTimeZoneRetrieval                           bool // bit 26
	AdditionalMsisdn                                 bool // bit 27
	SmsInMME                                         bool // bit 28
	SmsInSGSN                                        bool // bit 29
	UeReachabilityNotification                       bool // bit 30
	StateLocationInformationRetrieval                bool // bit 31
	PartialPurge                                     bool // bit 32
	GddInSGSN                                        bool // bit 33
	SgsnCAMELCapability                              bool // bit 34
	PcscfRestoration                                 bool // bit 35
	DedicatedCoreNetworks                            bool // bit 36
	NonIPPDNTypeAPNs                                 bool // bit 37
	NonIPPDPTypeAPNs                                 bool // bit 38
	NrAsSecondaryRAT                                 bool // bit 39
}

// ExtSupportedFeatures (BIT STRING SIZE 1..40) per TS 29.002
// MAP-MS-DataTypes.asn:687. Extension to SupportedFeatures for newer
// feature bits; only 1 bit is currently defined.
type ExtSupportedFeatures struct {
	UnlicensedSpectrumAsSecondaryRAT bool // bit 0
}

// ODBData per TS 29.002 MAP-MS-DataTypes.asn:1770. Wraps the general
// ODB bit-string with an optional HPLMN-specific overlay.
type ODBData struct {
	OdbGeneralData *ODBGeneralData // mandatory
	OdbHPLMNData   *ODBHPLMNData   // optional
}

// ZoneCode (OCTET STRING SIZE 2) per TS 29.002 MAP-MS-DataTypes.asn:2073.
// Internal structure defined in 3GPP TS 23.003.
type ZoneCode HexBytes

// ZoneCodeList (SEQUENCE SIZE 1..10 OF ZoneCode) per TS 29.002
// MAP-MS-DataTypes.asn:2070.
type ZoneCodeList []ZoneCode

// MaxNumOfZoneCodes is the upper bound on ZoneCodeList per TS 29.002.
const MaxNumOfZoneCodes = 10

// MaxNumOfVBSGroupIds is the upper bound on VBSDataList per TS 29.002.
const MaxNumOfVBSGroupIds = 50

// MaxNumOfVGCSGroupIds is the upper bound on VGCSDataList per TS 29.002.
const MaxNumOfVGCSGroupIds = 50

// AdditionalSubscriptions (BIT STRING SIZE 3..8) per TS 29.002
// MAP-MS-DataTypes.asn:2711. Carries VGCS uplink-request privileges.
// Bits other than the three listed below shall be discarded by the
// receiver per spec.
type AdditionalSubscriptions struct {
	PrivilegedUplinkRequest bool // bit 0
	EmergencyUplinkRequest  bool // bit 1
	EmergencyReset          bool // bit 2
}

// VoiceBroadcastData per TS 29.002 MAP-MS-DataTypes.asn:2717.
// GroupId must encode to exactly 3 TBCD octets (6 hex nibbles); pass
// "ffffff" as the required filler when LongGroupId is present per spec.
// LongGroupId must encode to exactly 4 TBCD octets (8 hex nibbles).
// The encoder rejects inputs that violate these invariants.
type VoiceBroadcastData struct {
	GroupId                  string // mandatory TBCD, exactly 6 hex digits
	BroadcastInitEntitlement bool   // NULL marker
	LongGroupId              string // optional TBCD, exactly 8 hex digits
}

// VoiceGroupCallData per TS 29.002 MAP-MS-DataTypes.asn:2695.
// GroupId must encode to exactly 3 TBCD octets (6 hex nibbles); pass
// "ffffff" as the required filler when LongGroupId is present per spec.
// LongGroupId must encode to exactly 4 TBCD octets (8 hex nibbles).
// The encoder rejects inputs that violate these invariants.
//
// AdditionalInfo is an opaque BIT STRING (SIZE 1..136 per TS 43.068),
// modeled here as HexBytes. This representation only preserves
// byte-aligned values — the encoder sets BitLength to len(bytes)*8,
// and the decoder discards any trailing sub-byte bits. The encoder
// rejects values exceeding the 17-octet (136-bit) maximum.
type VoiceGroupCallData struct {
	GroupId                 string                   // mandatory TBCD, exactly 6 hex digits
	AdditionalSubscriptions *AdditionalSubscriptions // optional
	AdditionalInfo          HexBytes                 // optional, byte-aligned only, at most 17 octets per TS 43.068
	LongGroupId             string                   // optional TBCD, exactly 8 hex digits
}

// Octet size constants for VBS/VGCS TBCD identifiers and AdditionalInfo
// per TS 29.002 MAP-MS-DataTypes.asn:2729-2738 and TS 43.068.
const (
	GroupIdOctets             = 3
	LongGroupIdOctets         = 4
	MaxAdditionalInfoOctets   = 17 // 136 bits
)

// VBSDataList per TS 29.002 MAP-MS-DataTypes.asn:2685 (SIZE 1..50).
type VBSDataList []VoiceBroadcastData

// VGCSDataList per TS 29.002 MAP-MS-DataTypes.asn:2688 (SIZE 1..50).
type VGCSDataList []VoiceGroupCallData

// CancelLocationRes represents a CancelLocation response (opCode 3) per
// 3GPP TS 29.002. The response body carries only an optional
// ExtensionContainer; the wire response is effectively empty in practice.
type CancelLocationRes struct{}

// MCSSInfo (SEQUENCE) per TS 29.002 MAP-CommonDataTypes.asn:627.
// Carries multicall supplementary-service status and bearer counts.
type MCSSInfo struct {
	SsCode   SsCode   // [0] mandatory
	SsStatus HexBytes // [1] mandatory: Ext-SS-Status (1..5 octets)
	NbrSB    int      // [2] mandatory: MaxMC-Bearers (2..7)
	NbrUser  int      // [3] mandatory: MC-Bearers (1..7)
}

// CSGSubscriptionData (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1262.
// CsgId is a 27-bit Closed Subscriber Group identifier per TS 23.003.
type CSGSubscriptionData struct {
	CsgId              HexBytes // mandatory: 27-bit BIT STRING (4 octets carrying 27 bits)
	CsgIdBitLength     int      // 0 means 27 (the spec-mandated default)
	ExpirationDate     HexBytes // optional: Time (UTCTime/GeneralizedTime BER-encoded)
	LipaAllowedAPNList []HexBytes // [0] optional: list of APN OCTET STRINGs (SIZE 2..63)
	PlmnId             HexBytes   // [1] optional: PLMN-Id (3 octets)
}

// CSGSubscriptionDataList (SEQUENCE SIZE 1..50 OF CSG-SubscriptionData) per
// TS 29.002 MAP-MS-DataTypes.asn:1259.
type CSGSubscriptionDataList []CSGSubscriptionData

// VPLMNCSGSubscriptionDataList (SEQUENCE SIZE 1..50 OF CSG-SubscriptionData)
// per TS 29.002 MAP-MS-DataTypes.asn:1271. Same shape as CSGSubscriptionDataList.
type VPLMNCSGSubscriptionDataList []CSGSubscriptionData

// MaxNumOfCSGSubscriptions is the upper bound on CSGSubscriptionDataList and
// VPLMNCSGSubscriptionDataList per TS 29.002.
const MaxNumOfCSGSubscriptions = 50

// CSGIdBitLength is the spec-mandated bit length for CSG-Id per
// TS 29.002 MAP-MS-DataTypes.asn:1274 (BIT STRING SIZE 27).
const CSGIdBitLength = 27

// AdjacentAccessRestrictionData (SEQUENCE) per TS 29.002
// MAP-MS-DataTypes.asn:1478.
type AdjacentAccessRestrictionData struct {
	PlmnId                   HexBytes                  // [0] mandatory: 3-octet PLMN-Id
	AccessRestrictionData    AccessRestrictionData     // [1] mandatory
	ExtAccessRestrictionData *ExtAccessRestrictionData // [2] optional
}

// AdjacentAccessRestrictionDataList (SEQUENCE SIZE 1..50 OF
// AdjacentAccessRestrictionData) per TS 29.002 MAP-MS-DataTypes.asn:1475.
type AdjacentAccessRestrictionDataList []AdjacentAccessRestrictionData

// MaxNumOfAdjacentPLMN is the upper bound on AdjacentAccessRestrictionDataList
// per TS 29.002.
const MaxNumOfAdjacentPLMN = 50

// IMSIGroupId (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1245.
type IMSIGroupId struct {
	GroupServiceID uint32   // [0] mandatory: 0..4294967295
	PlmnId         HexBytes // [1] mandatory: 3-octet PLMN-Id
	LocalGroupID   HexBytes // [2] mandatory: 1..10 octets
}

// IMSIGroupIdList (SEQUENCE SIZE 1..50 OF IMSI-GroupId) per TS 29.002
// MAP-MS-DataTypes.asn:1242.
type IMSIGroupIdList []IMSIGroupId

// MaxNumOfIMSIGroupId is the upper bound on IMSIGroupIdList per TS 29.002.
const MaxNumOfIMSIGroupId = 50

// EDRXCycleLength (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1210.
// EDRXCycleLengthValue is a single-octet code per 3GPP TS 29.272 clause 7.3.216.
type EDRXCycleLength struct {
	RatType              UsedRATType // [0] mandatory: 0..5
	EDRXCycleLengthValue HexBytes    // [1] mandatory: exactly 1 octet
}

// EDRXCycleLengthList (SEQUENCE SIZE 1..8 OF EDRX-Cycle-Length) per TS 29.002
// MAP-MS-DataTypes.asn:1207.
type EDRXCycleLengthList []EDRXCycleLength

// MaxNumOfEDRXCycleLength is the upper bound on EDRXCycleLengthList per TS 29.002.
const MaxNumOfEDRXCycleLength = 8

// UsedRATType (ENUMERATED) per TS 29.002 MAP-MS-DataTypes.asn:4241 in the
// go-asn1 library. Values are extensible per spec; unknown values must be
// preserved on round-trip per Postel's law.
type UsedRATType int

const (
	UsedRATTypeUtran          UsedRATType = 0
	UsedRATTypeGeran          UsedRATType = 1
	UsedRATTypeGan            UsedRATType = 2
	UsedRATTypeIHspaEvolution UsedRATType = 3
	UsedRATTypeEUtran         UsedRATType = 4
	UsedRATTypeNbIot          UsedRATType = 5
)

// ResetIdList (SEQUENCE SIZE 1..50 OF Reset-Id) per TS 29.002
// MAP-MS-DataTypes.asn:1223. Each Reset-Id is an OCTET STRING (SIZE 1..4)
// unique within the HPLMN.
type ResetIdList []HexBytes

// MaxNumOfResetId is the upper bound on ResetIdList per TS 29.002.
const MaxNumOfResetId = 50

// MaxResetIdOctets is the upper bound on a single Reset-Id per TS 29.002
// (OCTET STRING SIZE 1..4).
const MaxResetIdOctets = 4

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

	ErrSuperChargerInfoNoAlternative        = errors.New("updateLocation: SuperChargerInfo CHOICE has no alternative set")
	ErrSuperChargerInfoMultipleAlternatives = errors.New("updateLocation: SuperChargerInfo CHOICE has multiple alternatives set")

	ErrAtiPsSubscriberStateNoAlternative        = errors.New("ati: PsSubscriberState CHOICE has no alternative set")
	ErrAtiPsSubscriberStateMultipleAlternatives = errors.New("ati: PsSubscriberState CHOICE has multiple alternatives set")

	ErrIscInvalidAbsentSubscriberDiagnosticSM = errors.New("informServiceCentre: value must be 0..255")

	ErrAscMissingMSISDN               = errors.New("alertServiceCentre: MSISDN is empty")
	ErrAscMissingServiceCentreAddress = errors.New("alertServiceCentre: ServiceCentreAddress is empty")
	ErrAscInvalidSmsGmscAlertEvent    = errors.New("alertServiceCentre: SmsGmscAlertEvent must be 0 or 1")

	ErrPurgeMSMissingIMSI = errors.New("purgeMS: IMSI is empty")

	ErrUpdateLocationMissingIMSI      = errors.New("updateLocation: IMSI is empty")
	ErrUpdateLocationMissingMSCNumber = errors.New("updateLocation: MSCNumber is empty")
	ErrUpdateLocationMissingVLRNumber = errors.New("updateLocation: VLRNumber is empty")

	ErrMtFsmMissingIMSI                   = errors.New("mtFsm: IMSI is empty")
	ErrMtFsmMissingServiceCentreAddressOA = errors.New("mtFsm: ServiceCentreAddressOA is empty")

	ErrSriSmMissingMSISDN               = errors.New("sriSm: MSISDN is empty")
	ErrSriSmMissingServiceCentreAddress = errors.New("sriSm: ServiceCentreAddress is empty")

	ErrSaiMissingIMSI                                = errors.New("sai: IMSI is empty")
	ErrSaiInvalidNumberOfRequestedVectors            = errors.New("sai: NumberOfRequestedVectors must be 1..5")
	ErrSaiInvalidNumberOfRequestedAdditionalVectors  = errors.New("sai: NumberOfRequestedAdditionalVectors must be 1..5")
	ErrSaiInvalidUeUsageType                         = errors.New("sai: UeUsageType must be exactly 4 octets")
	ErrSaiInvalidPLMNId                              = errors.New("sai: RequestingPLMNId must be exactly 3 octets")
	ErrSaiAuthSetListChoiceMultipleAlternatives      = errors.New("sai: AuthenticationSetList CHOICE has multiple alternatives set")
	ErrSaiAuthSetListChoiceNoAlternative             = errors.New("sai: AuthenticationSetList CHOICE has no alternative set")
	ErrSaiInvalidRequestingNodeType                  = errors.New("sai: RequestingNodeType must be one of vlr(0), sgsn(1), s-cscf(2), bsf(3), gan-aaa-server(4), wlan-aaa-server(5), mme(16), mme-sgsn(17)")
	ErrSaiInvalidEpsAuthSetListSize                  = errors.New("sai: EpsAuthenticationSetList size must be at most 5 entries when present")

	ErrPsiMissingIMSI         = errors.New("psi: IMSI is empty")
	ErrPsiInvalidLMSI         = errors.New("psi: LMSI, if set, must be exactly 4 octets")
	ErrPsiInvalidCallPriority = errors.New("psi: CallPriority must be 0..15")

	ErrCancelLocIdentityChoiceNoAlternative = errors.New("cancelLocation: Identity CHOICE has no alternative set")
	ErrCancelLocIdentityChoiceMultiple      = errors.New("cancelLocation: Identity CHOICE has multiple alternatives set")
	ErrCancelLocIdentityMissingIMSI         = errors.New("cancelLocation: IMSIWithLMSI.IMSI is empty")
	ErrCancelLocIdentityInvalidLMSI         = errors.New("cancelLocation: IMSIWithLMSI.LMSI must be exactly 4 octets")
	ErrCancelLocInvalidCancellationType     = errors.New("cancelLocation: CancellationType must be one of updateProcedure(0), subscriptionWithdraw(1), initialAttachProcedure(2)")
	ErrCancelLocInvalidTypeOfUpdate         = errors.New("cancelLocation: TypeOfUpdate must be one of sgsn-change(0), mme-change(1)")
	ErrCancelLocTypeOfUpdateNotApplicable   = errors.New("cancelLocation: TypeOfUpdate is only valid when CancellationType is updateProcedure or initialAttachProcedure")
	ErrCancelLocMtrfBothSet                 = errors.New("cancelLocation: MtrfSupportedAndAuthorized and MtrfSupportedAndNotAuthorized are mutually exclusive")
	ErrCancelLocInvalidNewLMSI              = errors.New("cancelLocation: NewLMSI, if set, must be exactly 4 octets")

	ErrCamelInvalidOTriggerPoint             = errors.New("camel: O-BcsmTriggerDetectionPoint must be collectedInfo(2) or routeSelectFailure(4)")
	ErrCamelInvalidTTriggerPoint             = errors.New("camel: T-BcsmTriggerDetectionPoint must be termAttemptAuthorized(12), tBusy(13), or tNoAnswer(14)")
	ErrCamelInvalidDefaultCallHandling       = errors.New("camel: DefaultCallHandling must be continueCall(0) or releaseCall(1)")
	ErrCamelInvalidCallTypeCriteria          = errors.New("camel: CallTypeCriteria must be forwarded(0) or notForwarded(1)")
	ErrCamelInvalidMatchType                 = errors.New("camel: MatchType must be inhibiting(0) or enabling(1)")
	ErrCamelInvalidServiceKey                = errors.New("camel: ServiceKey must be 0..2147483647")
	ErrCamelMissingGsmSCFAddress             = errors.New("camel: GsmSCFAddress is mandatory and must be non-empty")
	ErrCamelMissingDialledNumber             = errors.New("camel: DialledNumber is mandatory on DPAnalysedInfoCriterium")
	ErrCamelInvalidCamelCapabilityHandling   = errors.New("camel: CamelCapabilityHandling must be 1..4 when set")
	ErrCamelInvalidTDPDataListSize           = errors.New("camel: TDP data list must contain 1..10 entries")
	ErrCamelInvalidDPAnalysedInfoListSize    = errors.New("camel: DPAnalysedInfoCriteriaList must contain 1..10 entries when present")
	ErrCamelInvalidCauseValue                = errors.New("camel: CauseValue must be 0..127")
	ErrCamelInvalidCauseValueOctetLength     = errors.New("camel: CauseValue is OCTET STRING (SIZE(1)) — each entry must be exactly 1 octet")
	ErrCamelInvalidCauseValueListSize        = errors.New("camel: CauseValueCriteria must contain 1..5 entries when present")
	ErrCamelInvalidDestinationNumberLength   = errors.New("camel: DestinationNumberLength must be 1..15")
	ErrCamelMissingDestinationNumber         = errors.New("camel: DestinationNumberList entry must have non-empty Digits")
	ErrCamelMissingDestinationNumberCriteria = errors.New("camel: DestinationNumberCriteria requires at least one of DestinationNumberList or DestinationNumberLengthList")
	ErrCamelInvalidCriteriaListSize          = errors.New("camel: TDP-CriteriaList must contain 1..10 entries when present")
	ErrCamelInvalidSSEventListSize           = errors.New("camel: SsEventList must contain 1..10 entries")
	ErrCamelInvalidMobilityTriggersSize      = errors.New("camel: MobilityTriggers must contain 1..10 single-octet entries")
	ErrCamelInvalidMobilityTriggerOctet      = errors.New("camel: each MobilityTriggers entry must be exactly 1 octet")
	ErrCamelInvalidSMSTDPDataListSize        = errors.New("camel: SmsCAMELTDPDataList must contain 1..10 entries")
	ErrCamelSMSCSIMissingTDPData             = errors.New("camel: SMS-CSI must include SmsCAMELTDPDataList per TS 29.002 clause 8.8.1")
	ErrCamelSMSCSIMissingCapabilityHandling  = errors.New("camel: SMS-CSI must include CamelCapabilityHandling per TS 29.002 clause 8.8.1")
	ErrCamelInvalidSMSTriggerDetectionPoint  = errors.New("camel: SmsTriggerDetectionPoint must be sms-CollectedInfo(1) or sms-DeliveryRequest(2)")
	ErrCamelInvalidDefaultSMSHandling        = errors.New("camel: DefaultSMSHandling must be continueTransaction(0) or releaseTransaction(1)")
	ErrCamelInvalidMTSmsCAMELCriteriaSize    = errors.New("camel: MtSmsCAMELTDPCriteriaList must contain 1..5 entries when present")
	ErrCamelInvalidTPDUTypeCriterionSize     = errors.New("camel: TpduTypeCriterion must contain 1..5 entries when present")
	ErrCamelInvalidMTSMSTPDUType             = errors.New("camel: MT-SMS-TPDU-Type must be sms-DELIVER(0), sms-SUBMIT-REPORT(1), or sms-STATUS-REPORT(2)")

	// Ext-SS-Info CHOICE / nested SEQUENCE validation
	ErrExtSSInfoChoiceNoAlternative           = errors.New("extSSInfo: exactly one of ForwardingInfo, CallBarringInfo, CugInfo, SsData, EmlppInfo must be set")
	ErrExtSSInfoChoiceMultipleAlternatives    = errors.New("extSSInfo: only one of ForwardingInfo, CallBarringInfo, CugInfo, SsData, EmlppInfo may be set")
	ErrExtSSStatusInvalidSize                 = errors.New("extSSInfo: SsStatus (Ext-SS-Status) must be 1..5 octets")
	ErrExtForwOptionsInvalidSize              = errors.New("extSSInfo: ForwardingOptions (Ext-ForwOptions) must be 1..5 octets")
	ErrExtNoRepCondTimeOutOfRange             = errors.New("extSSInfo: NoReplyConditionTime must be 1..100 per Ext-NoRepCondTime")
	ErrExtForwSubaddressInvalidSize           = errors.New("extSSInfo: ForwardedToSubaddress (ISDN-SubaddressString) must be 1..21 octets")
	ErrExtForwFeatureListInvalidSize          = errors.New("extSSInfo: ForwardingFeatureList must contain 1..32 entries")
	ErrExtCallBarFeatureListInvalidSize       = errors.New("extSSInfo: CallBarringFeatureList must contain 1..32 entries")
	ErrExtBasicServiceGroupListInvalidSize    = errors.New("extSSInfo: BasicServiceGroupList must contain 1..32 entries when present")
	ErrCUGSubscriptionListInvalidSize         = errors.New("extSSInfo: CugSubscriptionList must contain 0..10 entries")
	ErrCUGFeatureListInvalidSize              = errors.New("extSSInfo: CugFeatureList must contain 1..32 entries when present")
	ErrCUGIndexOutOfRange                     = errors.New("extSSInfo: CugIndex must be 0..32767")
	ErrCUGInterlockInvalidSize                = errors.New("extSSInfo: CugInterlock must be exactly 4 octets")
	ErrIntraCUGOptionsInvalidValue            = errors.New("extSSInfo: IntraCUGOptions must be noCUG-Restrictions(0), cugIC-CallBarred(1), or cugOG-CallBarred(2)")
	ErrSSSubscriptionOptionChoiceNoAlternative        = errors.New("extSSInfo: SsSubscriptionOption requires exactly one of CliRestriction or Override")
	ErrSSSubscriptionOptionChoiceMultipleAlternatives  = errors.New("extSSInfo: SsSubscriptionOption may only have one of CliRestriction or Override set")
	ErrCliRestrictionOptionInvalidValue       = errors.New("extSSInfo: CliRestrictionOption must be permanent(0), temporaryDefaultRestricted(1), or temporaryDefaultAllowed(2)")
	ErrOverrideCategoryInvalidValue           = errors.New("extSSInfo: OverrideCategory must be overrideEnabled(0) or overrideDisabled(1)")
	ErrEMLPPPriorityOutOfRange                = errors.New("extSSInfo: EMLPP priority must be 0..6 per TS 29.002 (values 7..15 are spare and would be silently remapped on decode)")

	ErrODBDataMissingGeneralData       = errors.New("odbData: OdbGeneralData is mandatory and must be non-nil")
	ErrZoneCodeInvalidSize             = errors.New("zoneCode: each entry must be exactly 2 octets")
	ErrZoneCodeListInvalidSize         = errors.New("zoneCode: ZoneCodeList must contain 1..10 entries")
	ErrVBSDataListInvalidSize          = errors.New("vbsData: VBSDataList must contain 1..50 entries")
	ErrVGCSDataListInvalidSize         = errors.New("vgcsData: VGCSDataList must contain 1..50 entries")
	ErrGroupIdMissingWithoutLong       = errors.New("voiceGroupCallData/voiceBroadcastData: GroupId is mandatory")
	ErrGroupIdFillerRequired           = errors.New("voiceGroupCallData/voiceBroadcastData: when LongGroupId is present, GroupId must be the six TBCD fillers \"ffffff\" per TS 29.002")
	ErrGroupIdInvalidEncodedLength     = errors.New("voiceGroupCallData/voiceBroadcastData: GroupId must encode to exactly 3 TBCD octets")
	ErrLongGroupIdInvalidEncodedLength = errors.New("voiceGroupCallData/voiceBroadcastData: LongGroupId must encode to exactly 4 TBCD octets")
	ErrAdditionalInfoTooLong           = errors.New("voiceGroupCallData: AdditionalInfo exceeds the TS 43.068 maximum of 17 octets / 136 bits")

	ErrMCSSInfoNbrSBOutOfRange   = errors.New("mcSSInfo: NbrSB (MaxMC-Bearers) must be 2..7 per TS 29.002")
	ErrMCSSInfoNbrUserOutOfRange = errors.New("mcSSInfo: NbrUser (MC-Bearers) must be 1..7 per TS 29.002")
	ErrMCSSInfoMissingSsCode     = errors.New("mcSSInfo: SsCode is mandatory ([0]) and must be present on the wire")

	ErrCSGIdInvalidSize            = errors.New("csgSubscriptionData: CsgId BIT STRING (SIZE 27) requires exactly 4 octets carrying 27 bits; CsgIdBitLength must be set to 27")
	ErrCSGSubscriptionDataListSize = errors.New("csgSubscriptionDataList: must contain 1..50 entries when present")
	ErrLipaAllowedAPNListSize      = errors.New("csgSubscriptionData: LipaAllowedAPNList must contain 1..50 entries when present per TS 29.002")
	ErrAPNInvalidSize              = errors.New("apn: each entry must be 2..63 octets per TS 29.002 MAP-MS-DataTypes.asn:1654")
	ErrPlmnIdInvalidSize           = errors.New("plmnId must be exactly 3 octets per TS 23.003")

	ErrAdjacentAccessRestrictionListSize = errors.New("adjacentAccessRestrictionDataList: must contain 1..50 entries when present")

	ErrIMSIGroupIdListSize         = errors.New("imsiGroupIdList: must contain 1..50 entries when present")
	ErrIMSIGroupServiceIDOverflow  = errors.New("imsiGroupId: GroupServiceID must fit in 0..4294967295")
	ErrLocalGroupIDInvalidSize     = errors.New("imsiGroupId: LocalGroupID must be 1..10 octets per TS 29.002")

	ErrEDRXCycleLengthListSize    = errors.New("eDRXCycleLengthList: must contain 1..8 entries when present")
	ErrEDRXCycleLengthValueSize   = errors.New("eDRXCycleLength: EDRXCycleLengthValue must be exactly 1 octet per TS 29.002")

	ErrResetIdListSize     = errors.New("resetIdList: must contain 1..50 entries when present")
	ErrResetIdInvalidSize  = errors.New("resetId: each entry must be 1..4 octets per TS 29.002")
)
