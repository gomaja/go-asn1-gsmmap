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
type SmDeliveryNotIntended = gsm_map.SMDeliveryNotIntended

const (
	SmDeliveryOnlyIMSIRequested   = gsm_map.SMDeliveryNotIntendedOnlyIMSIRequested
	SmDeliveryOnlyMCCMNCRequested = gsm_map.SMDeliveryNotIntendedOnlyMCCMNCRequested
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
type SmDeliveryOutcome = gsm_map.SMDeliveryOutcome

const (
	SmDeliveryMemoryCapacityExceeded = gsm_map.SMDeliveryOutcomeMemoryCapacityExceeded
	SmDeliveryAbsentSubscriber       = gsm_map.SMDeliveryOutcomeAbsentSubscriber
	SmDeliverySuccessfulTransfer     = gsm_map.SMDeliveryOutcomeSuccessfulTransfer
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

// UsedRatType per 3GPP TS 29.002 (opCode 23). Aliased from go-asn1 per
// project rule "GSM-MAP spec constants must come from go-asn1 library,
// not defined locally".
type UsedRatType = gsm_map.UsedRATType

const (
	UsedRatUTRAN          = gsm_map.UsedRATTypeUtran
	UsedRatGERAN          = gsm_map.UsedRATTypeGeran
	UsedRatGAN            = gsm_map.UsedRATTypeGan
	UsedRatIHSPAEvolution = gsm_map.UsedRATTypeIHspaEvolution
	UsedRatEUTRAN         = gsm_map.UsedRATTypeEUtran
	UsedRatNBIOT          = gsm_map.UsedRATTypeNbIot
)

// UeSrvccCapability per 3GPP TS 29.002 (opCode 23). Aliased from go-asn1.
type UeSrvccCapability = gsm_map.UESRVCCCapability

const (
	UeSrvccNotSupported = gsm_map.UESRVCCCapabilityUeSrvccNotSupported
	UeSrvccSupported    = gsm_map.UESRVCCCapabilityUeSrvccSupported
)

// SmsRegisterRequest per 3GPP TS 29.002 (opCode 23). Aliased from go-asn1.
type SmsRegisterRequest = gsm_map.SMSRegisterRequest

const (
	SmsRegistrationRequired     = gsm_map.SMSRegisterRequestSmsRegistrationRequired
	SmsRegistrationNotPreferred = gsm_map.SMSRegisterRequestSmsRegistrationNotPreferred
	SmsRegistrationNoPreference = gsm_map.SMSRegisterRequestNoPreference
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
type DomainType = gsm_map.DomainType

const (
	CsDomain = gsm_map.DomainTypeCsDomain
	PsDomain = gsm_map.DomainTypePsDomain
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
type ImsVoiceOverPSSessionsIndication = gsm_map.IMSVoiceOverPSSessionsInd

const (
	IMSVoiceOverPSNotSupported = gsm_map.IMSVoiceOverPSSessionsIndImsVoiceOverPSSessionsNotSupported
	IMSVoiceOverPSSupported    = gsm_map.IMSVoiceOverPSSessionsIndImsVoiceOverPSSessionsSupported
	IMSVoiceOverPSUnknown      = gsm_map.IMSVoiceOverPSSessionsIndUnknown
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
type InterrogationType = gsm_map.InterrogationType

const (
	InterrogationBasicCall  = gsm_map.InterrogationTypeBasicCall
	InterrogationForwarding = gsm_map.InterrogationTypeForwarding
)

// ForwardingReason per 3GPP TS 29.002.
type ForwardingReason = gsm_map.ForwardingReason

const (
	ForwardingNotReachable = gsm_map.ForwardingReasonNotReachable
	ForwardingBusy         = gsm_map.ForwardingReasonBusy
	ForwardingNoReply      = gsm_map.ForwardingReasonNoReply
)

// NumberPortabilityStatus per 3GPP TS 29.002.
type NumberPortabilityStatus = gsm_map.NumberPortabilityStatus

const (
	MnpNotKnownToBePorted                  = gsm_map.NumberPortabilityStatusNotKnownToBePorted
	MnpOwnNumberPortedOut                  = gsm_map.NumberPortabilityStatusOwnNumberPortedOut
	MnpForeignNumberPortedToForeignNetwork = gsm_map.NumberPortabilityStatusForeignNumberPortedToForeignNetwork
	MnpOwnNumberNotPortedOut               = gsm_map.NumberPortabilityStatusOwnNumberNotPortedOut
	MnpForeignNumberPortedIn               = gsm_map.NumberPortabilityStatusForeignNumberPortedIn
)

// UnavailabilityCause per 3GPP TS 29.002.
type UnavailabilityCause = gsm_map.UnavailabilityCause

const (
	UnavailBearerServiceNotProvisioned = gsm_map.UnavailabilityCauseBearerServiceNotProvisioned
	UnavailTeleserviceNotProvisioned   = gsm_map.UnavailabilityCauseTeleserviceNotProvisioned
	UnavailAbsentSubscriber            = gsm_map.UnavailabilityCauseAbsentSubscriber
	UnavailBusySubscriber              = gsm_map.UnavailabilityCauseBusySubscriber
	UnavailCallBarred                  = gsm_map.UnavailabilityCauseCallBarred
	UnavailCugReject                   = gsm_map.UnavailabilityCauseCugReject
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
type OBcsmTriggerDetectionPoint = gsm_map.OBcsmTriggerDetectionPoint

const (
	OBcsmTriggerCollectedInfo      = gsm_map.OBcsmTriggerDetectionPointCollectedInfo
	OBcsmTriggerRouteSelectFailure = gsm_map.OBcsmTriggerDetectionPointRouteSelectFailure
)

// TBcsmTriggerDetectionPoint per 3GPP TS 29.002.
type TBcsmTriggerDetectionPoint = gsm_map.TBcsmTriggerDetectionPoint

const (
	TBcsmTriggerTermAttemptAuthorized = gsm_map.TBcsmTriggerDetectionPointTermAttemptAuthorized
	TBcsmTriggerTBusy                 = gsm_map.TBcsmTriggerDetectionPointTBusy
	TBcsmTriggerTNoAnswer             = gsm_map.TBcsmTriggerDetectionPointTNoAnswer
)

// DefaultCallHandling per 3GPP TS 29.002.
type DefaultCallHandling = gsm_map.DefaultCallHandling

const (
	DefaultCallHandlingContinueCall = gsm_map.DefaultCallHandlingContinueCall
	DefaultCallHandlingReleaseCall  = gsm_map.DefaultCallHandlingReleaseCall
)

// CallTypeCriteria per 3GPP TS 29.002 (O-BcsmCamelTDP-Criteria).
type CallTypeCriteria = gsm_map.CallTypeCriteria

const (
	CallTypeCriteriaForwarded    = gsm_map.CallTypeCriteriaForwarded
	CallTypeCriteriaNotForwarded = gsm_map.CallTypeCriteriaNotForwarded
)

// MatchType per 3GPP TS 29.002 (DestinationNumberCriteria).
type MatchType = gsm_map.MatchType

const (
	MatchTypeInhibiting = gsm_map.MatchTypeInhibiting
	MatchTypeEnabling   = gsm_map.MatchTypeEnabling
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
type DefaultSMSHandling = gsm_map.DefaultSMSHandling

const (
	DefaultSMSHandlingContinueTransaction = gsm_map.DefaultSMSHandlingContinueTransaction
	DefaultSMSHandlingReleaseTransaction  = gsm_map.DefaultSMSHandlingReleaseTransaction
)

// SMSTriggerDetectionPoint per 3GPP TS 29.002 MAP-MS-DataTypes.asn:2487.
// ENUMERATED { sms-CollectedInfo(1), sms-DeliveryRequest(2), ... }.
type SMSTriggerDetectionPoint = gsm_map.SMSTriggerDetectionPoint

const (
	SMSTriggerDetectionPointSmsCollectedInfo   = gsm_map.SMSTriggerDetectionPointSmsCollectedInfo
	SMSTriggerDetectionPointSmsDeliveryRequest = gsm_map.SMSTriggerDetectionPointSmsDeliveryRequest
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
type MTSMSTPDUType = gsm_map.MTSMSTPDUType

const (
	MTSMSTPDUTypeSmsDELIVER      = gsm_map.MTSMSTPDUTypeSmsDELIVER
	MTSMSTPDUTypeSmsSUBMITREPORT = gsm_map.MTSMSTPDUTypeSmsSUBMITREPORT
	MTSMSTPDUTypeSmsSTATUSREPORT = gsm_map.MTSMSTPDUTypeSmsSTATUSREPORT
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
type CliRestrictionOption = gsm_map.CliRestrictionOption

const (
	CliRestrictionPermanent                  = gsm_map.CliRestrictionOptionPermanent
	CliRestrictionTemporaryDefaultRestricted = gsm_map.CliRestrictionOptionTemporaryDefaultRestricted
	CliRestrictionTemporaryDefaultAllowed    = gsm_map.CliRestrictionOptionTemporaryDefaultAllowed
)

// OverrideCategory per TS 29.002 MAP-SS-DataTypes.asn:182.
// ENUMERATED { overrideEnabled(0), overrideDisabled(1) }.
type OverrideCategory = gsm_map.OverrideCategory

const (
	OverrideEnabled  = gsm_map.OverrideCategoryOverrideEnabled
	OverrideDisabled = gsm_map.OverrideCategoryOverrideDisabled
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
type IntraCUGOptions = gsm_map.IntraCUGOptions

const (
	IntraCUGNoRestrictions = gsm_map.IntraCUGOptionsNoCUGRestrictions
	IntraCUGICCallBarred   = gsm_map.IntraCUGOptionsCugICCallBarred
	IntraCUGOGCallBarred   = gsm_map.IntraCUGOptionsCugOGCallBarred
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
type SmsGmscAlertEvent = gsm_map.SmsGmscAlertEvent

const (
	SmsGmscAlertMsAvailableForMtSms   = gsm_map.SmsGmscAlertEventMsAvailableForMtSms
	SmsGmscAlertMsUnderNewServingNode = gsm_map.SmsGmscAlertEventMsUnderNewServingNode
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
type RequestingNodeType = gsm_map.RequestingNodeType

const (
	RequestingNodeVlr           = gsm_map.RequestingNodeTypeVlr
	RequestingNodeSgsn          = gsm_map.RequestingNodeTypeSgsn
	RequestingNodeSCscf         = gsm_map.RequestingNodeTypeSCscf
	RequestingNodeBsf           = gsm_map.RequestingNodeTypeBsf
	RequestingNodeGanAAAServer  = gsm_map.RequestingNodeTypeGanAaaServer
	RequestingNodeWlanAAAServer = gsm_map.RequestingNodeTypeWlanAaaServer
	RequestingNodeMme           = gsm_map.RequestingNodeTypeMme
	RequestingNodeMmeSgsn       = gsm_map.RequestingNodeTypeMmeSgsn
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
type CancellationType = gsm_map.CancellationType

const (
	CancellationTypeUpdateProcedure        = gsm_map.CancellationTypeUpdateProcedure
	CancellationTypeSubscriptionWithdraw   = gsm_map.CancellationTypeSubscriptionWithdraw
	CancellationTypeInitialAttachProcedure = gsm_map.CancellationTypeInitialAttachProcedure
)

// TypeOfUpdate per 3GPP TS 29.002 (opCode 3). This field is only valid
// when CancellationType is updateProcedure or initialAttachProcedure.
// The constraint is enforced on both the encode and decode paths and
// returns ErrCancelLocTypeOfUpdateNotApplicable when violated.
type TypeOfUpdate = gsm_map.TypeOfUpdate

const (
	TypeOfUpdateSgsnChange = gsm_map.TypeOfUpdateSgsnChange
	TypeOfUpdateMmeChange  = gsm_map.TypeOfUpdateMmeChange
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
type SubscriberStatus = gsm_map.SubscriberStatus

const (
	SubscriberStatusServiceGranted            = gsm_map.SubscriberStatusServiceGranted
	SubscriberStatusOperatorDeterminedBarring = gsm_map.SubscriberStatusOperatorDeterminedBarring
)

// NetworkAccessMode per 3GPP TS 29.002 (MAP-MS-DataTypes.asn:1509).
// ENUMERATED { packetAndCircuit(0), onlyCircuit(1), onlyPacket(2) }.
type NetworkAccessMode = gsm_map.NetworkAccessMode

const (
	NetworkAccessModePacketAndCircuit = gsm_map.NetworkAccessModePacketAndCircuit
	NetworkAccessModeOnlyCircuit      = gsm_map.NetworkAccessModeOnlyCircuit
	NetworkAccessModeOnlyPacket       = gsm_map.NetworkAccessModeOnlyPacket
)

// RegionalSubscriptionResponse per 3GPP TS 29.002 (MAP-MS-DataTypes.asn:2091).
// ENUMERATED { networkNode-AreaRestricted(0), tooManyZoneCodes(1),
// zoneCodesConflict(2), regionalSubscNotSupported(3) }.
type RegionalSubscriptionResponse = gsm_map.RegionalSubscriptionResponse

const (
	RegionalSubscriptionResponseNetworkNodeAreaRestricted = gsm_map.RegionalSubscriptionResponseNetworkNodeAreaRestricted
	RegionalSubscriptionResponseTooManyZoneCodes          = gsm_map.RegionalSubscriptionResponseTooManyZoneCodes
	RegionalSubscriptionResponseZoneCodesConflict         = gsm_map.RegionalSubscriptionResponseZoneCodesConflict
	RegionalSubscriptionResponseRegionalSubscNotSupported = gsm_map.RegionalSubscriptionResponseRegionalSubscNotSupported
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
	CsgId              HexBytes   // mandatory: 27-bit BIT STRING (4 octets carrying 27 bits)
	CsgIdBitLength     int        // mandatory: must be set to 27; any other value (including 0) is rejected by the encoder
	ExpirationDate     HexBytes   // optional: Time (UTCTime/GeneralizedTime BER-encoded)
	LipaAllowedAPNList []HexBytes // [0] optional: list of APN OCTET STRINGs (SIZE 2..63), 1..50 entries when present
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
	// RatType: currently defined values are 0..5 (UsedRatUTRAN..UsedRatNBIOT);
	// the spec marks the enum as extensible, so unknown values are preserved
	// across round-trip per Postel's law.
	RatType              UsedRatType // [0] mandatory
	EDRXCycleLengthValue HexBytes    // [1] mandatory: exactly 1 octet
}

// EDRXCycleLengthList (SEQUENCE SIZE 1..8 OF EDRX-Cycle-Length) per TS 29.002
// MAP-MS-DataTypes.asn:1207.
type EDRXCycleLengthList []EDRXCycleLength

// MaxNumOfEDRXCycleLength is the upper bound on EDRXCycleLengthList per TS 29.002.
const MaxNumOfEDRXCycleLength = 8

// ResetIdList (SEQUENCE SIZE 1..50 OF Reset-Id) per TS 29.002
// MAP-MS-DataTypes.asn:1223. Each Reset-Id is an OCTET STRING (SIZE 1..4)
// unique within the HPLMN.
type ResetIdList []HexBytes

// MaxNumOfResetId is the upper bound on ResetIdList per TS 29.002.
const MaxNumOfResetId = 50

// MaxResetIdOctets is the upper bound on a single Reset-Id per TS 29.002
// (OCTET STRING SIZE 1..4).
const MaxResetIdOctets = 4

// AMBR (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1386. The two
// mandatory bandwidth fields are Bandwidth INTEGER (bits per second);
// the extended pair carries kbps values for >4 Gbps profiles.
type AMBR struct {
	MaxRequestedBandwidthUL         int64 // [0] mandatory, bits per second
	MaxRequestedBandwidthDL         int64 // [1] mandatory, bits per second
	ExtendedMaxRequestedBandwidthUL *int64 // [3] optional, kilobits per second
	ExtendedMaxRequestedBandwidthDL *int64 // [4] optional, kilobits per second
}

// SIPTOPermission (ENUMERATED) per TS 29.002 MAP-MS-DataTypes.asn:1567.
// Constants alias the go-asn1 spec exports per project rule "GSM-MAP
// spec constants must come from go-asn1 library, not defined locally".
type SIPTOPermission = gsm_map.SIPTOPermission

const (
	SIPTOAboveRanAllowed    = gsm_map.SIPTOPermissionSiptoAboveRanAllowed
	SIPTOAboveRanNotAllowed = gsm_map.SIPTOPermissionSiptoAboveRanNotAllowed
)

// SIPTOLocalNetworkPermission (ENUMERATED) per TS 29.002
// MAP-MS-DataTypes.asn:1572. Aliased from go-asn1.
type SIPTOLocalNetworkPermission = gsm_map.SIPTOLocalNetworkPermission

const (
	SIPTOAtLocalNetworkAllowed    = gsm_map.SIPTOLocalNetworkPermissionSiptoAtLocalNetworkAllowed
	SIPTOAtLocalNetworkNotAllowed = gsm_map.SIPTOLocalNetworkPermissionSiptoAtLocalNetworkNotAllowed
)

// LIPAPermission (ENUMERATED) per TS 29.002 MAP-MS-DataTypes.asn:1577.
// Aliased from go-asn1.
type LIPAPermission = gsm_map.LIPAPermission

const (
	LIPAProhibited  = gsm_map.LIPAPermissionLipaProhibited
	LIPAOnly        = gsm_map.LIPAPermissionLipaOnly
	LIPAConditional = gsm_map.LIPAPermissionLipaConditional
)

// NIDDMechanism (ENUMERATED) per TS 29.002 MAP-MS-DataTypes.asn:1362.
// Default (when absent) is sGi-based-data-delivery (0) per spec.
// Aliased from go-asn1.
type NIDDMechanism = gsm_map.NIDDMechanism

const (
	NIDDSGiBasedDataDelivery  = gsm_map.NIDDMechanismSGiBasedDataDelivery
	NIDDSCEFBasedDataDelivery = gsm_map.NIDDMechanismSCEFBasedDataDelivery
)

// PDPContext (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1522.
// Mandatory fields: PdpContextId, PdpType, QosSubscribed, Apn.
// VplmnAddressAllowed is an OPTIONAL ASN.1 NULL — true means present.
// All other fields are optional pointers / slices; the zero-value
// (nil pointer, nil slice, false bool) means the field is omitted from
// the wire encoding.
//
// Field ordering follows the ASN.1 tag order from the spec
// (mandatory base fields first, then extensions [0]..[14]).
//
// The Ext-QoS-Subscribed extension chain is hierarchical per
// MAP-MS-DataTypes.asn:1534-1538: Ext2 requires Ext, Ext3 requires
// Ext2, Ext4 requires Ext3.
type PDPContext struct {
	PdpContextId        int      // mandatory, ContextId 1..50
	PdpType             HexBytes // [16] mandatory, OCTET STRING SIZE 2
	PdpAddress          HexBytes // [17] optional, OCTET STRING SIZE 1..16 (nil = absent)
	QosSubscribed       HexBytes // [18] mandatory, OCTET STRING SIZE 3
	VplmnAddressAllowed bool     // [19] optional NULL — true when present
	Apn                 HexBytes // [20] mandatory, OCTET STRING SIZE 2..63

	ExtQoSSubscribed            HexBytes                     // [0] optional, SIZE 1..9
	PdpChargingCharacteristics  HexBytes                     // [1] optional, SIZE 2
	Ext2QoSSubscribed           HexBytes                     // [2] optional, SIZE 1..3 (requires ExtQoSSubscribed)
	Ext3QoSSubscribed           HexBytes                     // [3] optional, SIZE 1..2 (requires Ext2QoSSubscribed)
	Ext4QoSSubscribed           HexBytes                     // [4] optional, SIZE 1 (requires Ext3QoSSubscribed)
	ApnOiReplacement            HexBytes                     // [5] optional, SIZE 9..100
	ExtPdpType                  HexBytes                     // [6] optional, SIZE 2
	ExtPdpAddress               HexBytes                     // [7] optional, SIZE 1..16
	SiptoPermission             *SIPTOPermission             // [8] optional
	LipaPermission              *LIPAPermission              // [9] optional
	Ambr                        *AMBR                        // [10] optional
	RestorationPriority         HexBytes                     // [11] optional, SIZE 1
	SiptoLocalNetworkPermission *SIPTOLocalNetworkPermission // [12] optional
	NIDDMechanism               *NIDDMechanism               // [13] optional
	SCEFID                      HexBytes                     // [14] optional, FQDN SIZE 9..255
}

// GPRSDataList (SEQUENCE SIZE 1..50 OF PDP-Context) per TS 29.002
// MAP-MS-DataTypes.asn:1517.
type GPRSDataList []PDPContext

// GPRSSubscriptionData (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1585.
type GPRSSubscriptionData struct {
	CompleteDataListIncluded bool         // optional NULL — true when present
	GprsDataList             GPRSDataList // [1] mandatory, 1..50 entries
	ApnOiReplacement         HexBytes     // [3] optional, OCTET STRING SIZE 9..100
}

// LSAOnlyAccessIndicator (ENUMERATED) per TS 29.002 MAP-MS-DataTypes.asn:1702.
// Aliased from go-asn1.
type LSAOnlyAccessIndicator = gsm_map.LSAOnlyAccessIndicator

const (
	LSAAccessOutsideAllowed    = gsm_map.LSAOnlyAccessIndicatorAccessOutsideLSAsAllowed
	LSAAccessOutsideRestricted = gsm_map.LSAOnlyAccessIndicatorAccessOutsideLSAsRestricted
)

// LSAData (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1711.
type LSAData struct {
	LsaIdentity            HexBytes // [0] mandatory, OCTET STRING SIZE 3
	LsaAttributes          HexBytes // [1] mandatory, OCTET STRING SIZE 1
	LsaActiveModeIndicator bool     // [2] optional NULL — true when present
}

// LSADataList (SEQUENCE SIZE 1..20 OF LSAData) per TS 29.002
// MAP-MS-DataTypes.asn:1706.
type LSADataList []LSAData

// LSAInformation (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1718.
type LSAInformation struct {
	CompleteDataListIncluded bool                    // optional NULL — true when present
	LsaOnlyAccessIndicator   *LSAOnlyAccessIndicator // [1] optional
	LsaDataList              LSADataList             // [2] optional, 1..20 entries when present
}

// PDNConnectionContinuity (ENUMERATED) per TS 29.002
// MAP-MS-DataTypes.asn:1356. Aliased from go-asn1.
type PDNConnectionContinuity = gsm_map.PDNConnectionContinuity

const (
	PDNConnectionMaintain                            = gsm_map.PDNConnectionContinuityMaintainPDNConnection
	PDNConnectionDisconnectWithReactivationRequest   = gsm_map.PDNConnectionContinuityDisconnectPDNConnectionWithReactivationRequest
	PDNConnectionDisconnectWithoutReactivationRequest = gsm_map.PDNConnectionContinuityDisconnectPDNConnectionWithoutReactivationRequest
)

// PDNGWAllocationType (ENUMERATED) per TS 29.002 MAP-MS-DataTypes.asn:1437.
// Aliased from go-asn1.
type PDNGWAllocationType = gsm_map.PDNGWAllocationType

const (
	PDNGWAllocationStatic  = gsm_map.PDNGWAllocationTypeStatic
	PDNGWAllocationDynamic = gsm_map.PDNGWAllocationTypeDynamic
)

// WLANOffloadabilityIndication (ENUMERATED) per TS 29.002.
// Aliased from go-asn1.
type WLANOffloadabilityIndication = gsm_map.WLANOffloadabilityIndication

const (
	WLANOffloadabilityNotAllowed = gsm_map.WLANOffloadabilityIndicationNotAllowed
	WLANOffloadabilityAllowed    = gsm_map.WLANOffloadabilityIndicationAllowed
)

// AllocationRetentionPriority (SEQUENCE) per TS 29.002
// MAP-MS-DataTypes.asn:1420. PriorityLevel is an opaque INTEGER per spec
// (3GPP TS 29.212 defines actual semantics). Pre-emption flags are
// optional BOOLEANs.
type AllocationRetentionPriority struct {
	PriorityLevel           int64 // [0] mandatory
	PreEmptionCapability    *bool // [1] optional
	PreEmptionVulnerability *bool // [2] optional
}

// EPSQoSSubscribed (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1380.
// QoS-Class-Identifier is INTEGER (1..9) per asn:1415.
type EPSQoSSubscribed struct {
	QosClassIdentifier          int                         // [0] mandatory, 1..9
	AllocationRetentionPriority AllocationRetentionPriority // [1] mandatory
}

// SpecificAPNInfo (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1403.
// Reuses the pre-existing PdnGwIdentity public type (gsmmap.go:366),
// which enforces strict spec sizes (IPv4=4, IPv6=16) and the
// "at least one identity present" rule.
type SpecificAPNInfo struct {
	Apn           HexBytes      // [0] mandatory, APN SIZE 2..63
	PdnGwIdentity PdnGwIdentity // [1] mandatory
}

// SpecificAPNInfoList (SEQUENCE SIZE 1..50 OF SpecificAPNInfo) per
// TS 29.002 MAP-MS-DataTypes.asn:1398. The field is OPTIONAL on the
// wire; in this public API absence is represented by a nil slice.
// A non-nil empty slice (len == 0) is rejected as a size violation —
// callers must use nil rather than `SpecificAPNInfoList{}` to mean
// "absent". This matches the package convention for OPTIONAL
// SEQUENCE OF lists with SIZE (1..N) constraints (e.g.
// CSGSubscriptionDataList, EPSDataList, GPRSDataList).
type SpecificAPNInfoList []SpecificAPNInfo

// WLANOffloadability (SEQUENCE) per TS 29.002. Both fields optional.
type WLANOffloadability struct {
	WlanOffloadabilityEUTRAN *WLANOffloadabilityIndication // [0] optional
	WlanOffloadabilityUTRAN  *WLANOffloadabilityIndication // [1] optional
}

// APNConfiguration (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1327.
// Mandatory fields: ContextId, PdnType, Apn, EpsQosSubscribed.
// VplmnAddressAllowed and NonIPPDNTypeIndicator are OPTIONAL ASN.1 NULL
// fields modeled as bool (true means present).
//
// Field ordering follows the ASN.1 tag order. Tag [11]
// (extensionContainer) is intentionally omitted from the public type
// per the package-wide convention that ExtensionContainer is opaque
// metadata not surfaced to callers. It is dropped on decode and
// emitted as absent on encode — callers requiring opaque pass-through
// must add it at a higher layer.
type APNConfiguration struct {
	ContextId                int              // [0] mandatory, ContextId 1..50
	PdnType                  HexBytes         // [1] mandatory, OCTET STRING SIZE 1
	ServedPartyIPIPv4Address HexBytes         // [2] optional, PDP-Address SIZE 1..16
	Apn                      HexBytes         // [3] mandatory, APN SIZE 2..63
	EpsQosSubscribed         EPSQoSSubscribed // [4] mandatory
	PdnGwIdentity            *PdnGwIdentity   // [5] optional
	PdnGwAllocationType      *PDNGWAllocationType // [6] optional
	VplmnAddressAllowed      bool             // [7] optional NULL — true when present
	ChargingCharacteristics  HexBytes         // [8] optional, OCTET STRING SIZE 2
	Ambr                     *AMBR            // [9] optional
	SpecificAPNInfoList      SpecificAPNInfoList // [10] optional, 1..50 entries when present
	ServedPartyIPIPv6Address HexBytes                     // [12] optional, PDP-Address SIZE 1..16
	ApnOiReplacement         HexBytes                     // [13] optional, SIZE 9..100
	SiptoPermission          *SIPTOPermission             // [14] optional
	LipaPermission           *LIPAPermission              // [15] optional
	RestorationPriority      HexBytes                     // [16] optional, SIZE 1
	SiptoLocalNetworkPermission *SIPTOLocalNetworkPermission // [17] optional
	WlanOffloadability       *WLANOffloadability          // [18] optional
	NonIPPDNTypeIndicator    bool                         // [19] optional NULL — true when present
	NIDDMechanism            *NIDDMechanism               // [20] optional
	SCEFID                   HexBytes                     // [21] optional, FQDN SIZE 9..255
	PdnConnectionContinuity  *PDNConnectionContinuity     // [22] optional
}

// EPSDataList (SEQUENCE SIZE 1..50 OF APN-Configuration) per TS 29.002
// MAP-MS-DataTypes.asn:1320.
type EPSDataList []APNConfiguration

// APNConfigurationProfile (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1308.
type APNConfigurationProfile struct {
	DefaultContext           int         // mandatory, ContextId 1..50
	CompleteDataListIncluded bool        // optional NULL — true when present
	EpsDataList              EPSDataList // [1] mandatory, 1..50 entries
	AdditionalDefaultContext *int        // [3] optional, ContextId 1..50
}

// EPSSubscriptionData (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1283.
// All fields are OPTIONAL per spec. ApnConfigurationProfile typically
// carries the substantive payload, but the spec does not mandate it;
// callers requiring its presence should validate at their layer.
// The MpsCSPriority, MpsEPSPriority, and SubscribedVsrvcc fields are
// OPTIONAL ASN.1 NULL flags modeled as bool.
type EPSSubscriptionData struct {
	ApnOiReplacement        HexBytes                 // [0] optional, SIZE 9..100
	RfspId                  *int                     // [2] optional, INTEGER 1..256
	Ambr                    *AMBR                    // [3] optional
	ApnConfigurationProfile *APNConfigurationProfile // [4] optional
	// StnSr [6] OPTIONAL ISDN-AddressString. Empty digits string means
	// absent; a present wire frame that decodes to empty digits is
	// rejected on the decode path to keep round-trip semantics stable.
	StnSr       string // [6] optional, ISDN-AddressString digits
	StnSrNature uint8  // ISDN-AddressString nature-of-address octet
	StnSrPlan   uint8  // ISDN-AddressString numbering-plan octet
	MpsCSPriority           bool                     // [7] optional NULL — true when present
	MpsEPSPriority          bool                     // [8] optional NULL — true when present
	SubscribedVsrvcc        bool                     // [9] optional NULL — true when present
}

// EPS-DataList and SpecificAPNInfoList are bounded by the upstream
// constants gsm_map.MaxNumOfAPNConfigurations (50) and
// gsm_map.MaxNumOfSpecificAPNInfos (50) respectively — converters
// reference those constants directly per project rule
// "GSM-MAP spec constants must come from go-asn1, not defined locally".

// MaxRFSPID is the upper bound on RFSP-ID per TS 29.002
// MAP-MS-DataTypes.asn:1306 (`RFSP-ID ::= INTEGER (1..256)`). go-asn1
// v0.1.8 does not export this bound (`type RFSPID = int64`), so it is
// defined here pending upstream surfacing.
const MaxRFSPID = 256

// ============================================================================
// LCS-Information (TS 29.002 MAP-MS-DataTypes.asn:1490)
// ============================================================================

// GMLCRestriction (ENUMERATED) per TS 29.002 MAP-MS-DataTypes.asn:2027.
// Aliased from go-asn1.
type GMLCRestriction = gsm_map.GMLCRestriction

const (
	GMLCRestrictionGmlcList    = gsm_map.GMLCRestrictionGmlcList
	GMLCRestrictionHomeCountry = gsm_map.GMLCRestrictionHomeCountry
)

// NotificationToMSUser (ENUMERATED) per TS 29.002 MAP-MS-DataTypes.asn:2035.
// Aliased from go-asn1.
type NotificationToMSUser = gsm_map.NotificationToMSUser

const (
	NotifyLocationAllowed                         = gsm_map.NotificationToMSUserNotifyLocationAllowed
	NotifyAndVerifyLocationAllowedIfNoResponse    = gsm_map.NotificationToMSUserNotifyAndVerifyLocationAllowedIfNoResponse
	NotifyAndVerifyLocationNotAllowedIfNoResponse = gsm_map.NotificationToMSUserNotifyAndVerifyLocationNotAllowedIfNoResponse
	NotificationLocationNotAllowed                = gsm_map.NotificationToMSUserLocationNotAllowed
)

// LCSClientInternalID (ENUMERATED) per TS 29.002
// MAP-CommonDataTypes.asn (gsm_map.LCSClientInternalID).
// Aliased from go-asn1.
type LCSClientInternalID = gsm_map.LCSClientInternalID

const (
	LCSClientBroadcastService          = gsm_map.LCSClientInternalIDBroadcastService
	LCSClientOAndMHPLMN                = gsm_map.LCSClientInternalIDOAndMHPLMN
	LCSClientOAndMVPLMN                = gsm_map.LCSClientInternalIDOAndMVPLMN
	LCSClientAnonymousLocation         = gsm_map.LCSClientInternalIDAnonymousLocation
	LCSClientTargetMSsubscribedService = gsm_map.LCSClientInternalIDTargetMSsubscribedService
)

// LCSClientExternalID (SEQUENCE) per TS 29.002 MAP-CommonDataTypes.asn:642
// in the go-asn1 library. Surfaces the optional ISDN-AddressString as a
// digits string + nature/plan triple consistent with the rest of the
// public API.
type LCSClientExternalID struct {
	ExternalAddress       string // optional ISDN-AddressString digits (empty = absent)
	ExternalAddressNature uint8
	ExternalAddressPlan   uint8
}

// ExternalClient (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:2018.
type ExternalClient struct {
	ClientIdentity       LCSClientExternalID  // mandatory
	GmlcRestriction      *GMLCRestriction     // [0] optional
	NotificationToMSUser *NotificationToMSUser // [1] optional
}

// ExternalClientList (SEQUENCE SIZE 0..5 OF ExternalClient) per
// TS 29.002 MAP-MS-DataTypes.asn:2003.
//
// Note: spec allows 0 entries (the only such list in the package),
// so an empty slice is valid here unlike elsewhere.
type ExternalClientList []ExternalClient

// ExtExternalClientList (SEQUENCE SIZE 1..35 OF ExternalClient) per
// TS 29.002 MAP-MS-DataTypes.asn:2013.
type ExtExternalClientList []ExternalClient

// PLMNClientList (SEQUENCE SIZE 1..5 OF LCSClientInternalID) per
// TS 29.002 MAP-MS-DataTypes.asn:2008.
type PLMNClientList []LCSClientInternalID

// ServiceType (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:2050.
type ServiceType struct {
	ServiceTypeIdentity  int64                 // mandatory, LCSServiceTypeID INTEGER
	GmlcRestriction      *GMLCRestriction      // [0] optional
	NotificationToMSUser *NotificationToMSUser // [1] optional
}

// ServiceTypeList (SEQUENCE SIZE 1..32 OF ServiceType) per TS 29.002
// MAP-MS-DataTypes.asn:2045.
type ServiceTypeList []ServiceType

// LCSPrivacyClass (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1976.
// SsCode is a single-octet SS-Code; SsStatus is the Ext-SS-Status
// OCTET STRING (SIZE 1..5) shared with PR D's Ext-SS-Info tree.
type LCSPrivacyClass struct {
	SsCode                SsCode                // mandatory
	SsStatus              HexBytes              // mandatory, Ext-SS-Status 1..5 octets
	NotificationToMSUser  *NotificationToMSUser // [0] optional
	ExternalClientList    ExternalClientList    // [1] optional, 0..5 entries
	PlmnClientList        PLMNClientList        // [2] optional, 1..5 entries when present
	ExtExternalClientList ExtExternalClientList // [4] optional, 1..35 entries when present
	ServiceTypeList       ServiceTypeList       // [5] optional, 1..32 entries when present
}

// LCSPrivacyExceptionList (SEQUENCE SIZE 1..4 OF LCS-PrivacyClass) per
// TS 29.002 MAP-MS-DataTypes.asn:1971.
type LCSPrivacyExceptionList []LCSPrivacyClass

// MOLRClass (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:2064.
type MOLRClass struct {
	SsCode   SsCode   // mandatory
	SsStatus HexBytes // mandatory, Ext-SS-Status 1..5 octets
}

// MOLRList (SEQUENCE SIZE 1..3 OF MOLR-Class) per TS 29.002
// MAP-MS-DataTypes.asn:2059.
type MOLRList []MOLRClass

// GMLCAddress represents an ISDN-AddressString entry in a GMLC-List.
type GMLCAddress struct {
	Address string // mandatory ISDN-AddressString digits
	Nature  uint8
	Plan    uint8
}

// GMLCList (SEQUENCE SIZE 1..5 OF ISDN-AddressString) per TS 29.002
// MAP-MS-DataTypes.asn:1503.
type GMLCList []GMLCAddress

// LCSInformation (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1490.
// All four lists are OPTIONAL. AddLcsPrivacyExceptionList may only be
// present alongside LcsPrivacyExceptionList (extension list per LCS
// release). Callers requiring that invariant should validate at their
// layer.
type LCSInformation struct {
	GmlcList                   GMLCList                // [0] optional, 1..5 entries when present
	LcsPrivacyExceptionList    LCSPrivacyExceptionList // [1] optional, 1..4 entries when present
	MolrList                   MOLRList                // [2] optional, 1..3 entries when present
	AddLcsPrivacyExceptionList LCSPrivacyExceptionList // [3] optional, 1..4 entries when present
}

// ============================================================================
// ProvideSubscriberLocation foundation types (TS 29.002 MAP-LCS-DataTypes.asn)
// ============================================================================
//
// First PR of a staged ProvideSubscriberLocation (opCode 83) implementation.
// Lands the foundation LCS types so follow-up PRs can build top-level
// converters without a monolithic diff. No PSL Arg/Res top-level types yet —
// those land in a later PR alongside their converters.

// LocationEstimateType (ENUMERATED) per TS 29.002 MAP-LCS-DataTypes.asn:153.
// Extensible enum; decoders preserve unknown values per Postel's law.
// Aliased from go-asn1.
type LocationEstimateType = gsm_map.LocationEstimateType

const (
	LocationEstimateCurrentLocation              = gsm_map.LocationEstimateTypeCurrentLocation
	LocationEstimateCurrentOrLastKnownLocation   = gsm_map.LocationEstimateTypeCurrentOrLastKnownLocation
	LocationEstimateInitialLocation              = gsm_map.LocationEstimateTypeInitialLocation
	LocationEstimateActivateDeferredLocation     = gsm_map.LocationEstimateTypeActivateDeferredLocation
	LocationEstimateCancelDeferredLocation       = gsm_map.LocationEstimateTypeCancelDeferredLocation
	LocationEstimateNotificationVerificationOnly = gsm_map.LocationEstimateTypeNotificationVerificationOnly
)

// DeferredLocationEventType (BIT STRING SIZE 1..16) per TS 29.002
// MAP-LCS-DataTypes.asn:165. 5 named bits (msAvailable through periodicLDR).
// Surfaced as a bools-only struct to match the package's BIT STRING surrogate
// pattern (e.g., SupportedCamelPhases). Codec lives with PSL converters.
type DeferredLocationEventType struct {
	MsAvailable       bool // bit 0
	EnteringIntoArea  bool // bit 1
	LeavingFromArea   bool // bit 2
	BeingInsideArea   bool // bit 3
	PeriodicLDR       bool // bit 4
}

// LocationType (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:148.
// Mandatory in PSL-Arg.
type LocationType struct {
	LocationEstimateType      LocationEstimateType       // [0] mandatory
	DeferredLocationEventType *DeferredLocationEventType // [1] optional, present only past the extensibility marker
}

// LCSClientType (ENUMERATED) per TS 29.002 MAP-LCS-DataTypes.asn:188.
// Extensible enum. Aliased from go-asn1.
type LCSClientType = gsm_map.LCSClientType

const (
	LCSClientTypeEmergencyServices       = gsm_map.LCSClientTypeEmergencyServices
	LCSClientTypeValueAddedServices      = gsm_map.LCSClientTypeValueAddedServices
	LCSClientTypePlmnOperatorServices    = gsm_map.LCSClientTypePlmnOperatorServices
	LCSClientTypeLawfulInterceptServices = gsm_map.LCSClientTypeLawfulInterceptServices
)

// LCSFormatIndicator (ENUMERATED) per TS 29.002 MAP-LCS-DataTypes.asn:224.
// Extensible enum. Aliased from go-asn1.
type LCSFormatIndicator = gsm_map.LCSFormatIndicator

const (
	LCSFormatLogicalName  = gsm_map.LCSFormatIndicatorLogicalName
	LCSFormatEMailAddress = gsm_map.LCSFormatIndicatorEMailAddress
	LCSFormatMsisdn       = gsm_map.LCSFormatIndicatorMsisdn
	LCSFormatUrl          = gsm_map.LCSFormatIndicatorUrl
	LCSFormatSipUrl       = gsm_map.LCSFormatIndicatorSipUrl
)

// LCSClientName (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:199.
// NameString is a USSD-String of 1..63 octets per maxNameStringLength.
// Surfaced as an opaque byte slice; the USSD-DataCodingScheme is preserved
// verbatim so callers can decode per 3GPP TS 23.038 if needed.
//
// Note: the spec assigns tags [0]/[2]/[3] (skipping [1]) for these fields;
// the gap is intentional in the ASN.1 module and is not an off-by-one in
// this Go surface.
type LCSClientName struct {
	DataCodingScheme   uint8               // [0] mandatory, USSD-DataCodingScheme single octet
	NameString         HexBytes            // [2] mandatory, NameString 1..63 octets
	LcsFormatIndicator *LCSFormatIndicator // [3] optional, present only past the extensibility marker
}

// LCSRequestorID (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:214.
// RequestorIDString is a USSD-String of 1..63 octets per
// maxRequestorIDStringLength.
type LCSRequestorID struct {
	DataCodingScheme   uint8               // [0] mandatory, USSD-DataCodingScheme single octet
	RequestorIDString  HexBytes            // [1] mandatory, RequestorIDString 1..63 octets
	LcsFormatIndicator *LCSFormatIndicator // [2] optional, present only past the extensibility marker
}

// LCSClientID (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:178.
// LcsClientDialedByMS is an AddressString surfaced as digits + nature/plan
// triple consistent with the rest of the public API.
type LCSClientID struct {
	LcsClientType       LCSClientType        // [0] mandatory
	LcsClientExternalID *LCSClientExternalID // [1] optional
	// LcsClientDialedByMS, if present, is conveyed as an AddressString.
	// Empty digits = absent.
	LcsClientDialedByMS       string // [2] optional, AddressString digits
	LcsClientDialedByMSNature uint8
	LcsClientDialedByMSPlan   uint8
	LcsClientInternalID       *LCSClientInternalID // [3] optional
	LcsClientName             *LCSClientName       // [4] optional
	LcsAPN                    HexBytes             // [5] optional, APN OCTET STRING (past extensibility marker)
	LcsRequestorID            *LCSRequestorID      // [6] optional (past extensibility marker)
}

// ResponseTimeCategory (ENUMERATED) per TS 29.002 MAP-LCS-DataTypes.asn:266.
// Extensible enum; spec exception: unknown values shall be treated as
// delaytolerant(1) on decode. Aliased from go-asn1; the lenient remap
// happens in the decoder.
type ResponseTimeCategory = gsm_map.ResponseTimeCategory

const (
	ResponseTimeLowdelay      = gsm_map.ResponseTimeCategoryLowdelay
	ResponseTimeDelaytolerant = gsm_map.ResponseTimeCategoryDelaytolerant
)

// ResponseTime (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:261.
// An expandable SEQUENCE per spec, currently carrying only the category.
type ResponseTime struct {
	ResponseTimeCategory ResponseTimeCategory // mandatory
}

// LCSQoS (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:237.
// All fields optional. Horizontal/Vertical-Accuracy are 1-octet uncertainty
// codes per 3GPP TS 23.032; surfaced as raw single-octet HexBytes.
//
// Note: the ASN.1 definition includes an optional ExtensionContainer at
// tag [4]; consistent with the package-wide convention (see
// APNConfiguration), it is opaque metadata not surfaced to callers.
// It is dropped on decode and emitted as absent on encode — callers
// requiring opaque pass-through must add it at a higher layer.
type LCSQoS struct {
	HorizontalAccuracy        HexBytes      // [0] optional, 1 octet per TS 23.032
	VerticalCoordinateRequest bool          // [1] optional NULL; true when present, false when absent
	VerticalAccuracy          HexBytes      // [2] optional, 1 octet per TS 23.032
	ResponseTime              *ResponseTime // [3] optional
	VelocityRequest           bool          // [5] optional NULL; true when present, false when absent; present only past the extensibility marker
}

// PrivacyCheckRelatedAction (ENUMERATED) per TS 29.002
// MAP-LCS-DataTypes.asn:307. Aliased from go-asn1.
type PrivacyCheckRelatedAction = gsm_map.PrivacyCheckRelatedAction

const (
	PrivacyCheckAllowedWithoutNotification = gsm_map.PrivacyCheckRelatedActionAllowedWithoutNotification
	PrivacyCheckAllowedWithNotification    = gsm_map.PrivacyCheckRelatedActionAllowedWithNotification
	PrivacyCheckAllowedIfNoResponse        = gsm_map.PrivacyCheckRelatedActionAllowedIfNoResponse
	PrivacyCheckRestrictedIfNoResponse     = gsm_map.PrivacyCheckRelatedActionRestrictedIfNoResponse
	PrivacyCheckNotAllowed                 = gsm_map.PrivacyCheckRelatedActionNotAllowed
)

// LCSPrivacyCheck (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:302.
type LCSPrivacyCheck struct {
	CallSessionUnrelated PrivacyCheckRelatedAction  // [0] mandatory
	CallSessionRelated   *PrivacyCheckRelatedAction // [1] optional
}

// LCSCodeword (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:293.
// LcsCodewordString is a USSD-String of 1..20 octets per
// maxLCSCodewordStringLength.
type LCSCodeword struct {
	DataCodingScheme  uint8    // [0] mandatory, USSD-DataCodingScheme single octet
	LcsCodewordString HexBytes // [1] mandatory, LCSCodewordString 1..20 octets
}

// AccuracyFulfilmentIndicator (ENUMERATED) per TS 29.002
// MAP-LCS-DataTypes.asn:457. Extensible enum. Aliased from go-asn1.
type AccuracyFulfilmentIndicator = gsm_map.AccuracyFulfilmentIndicator

const (
	AccuracyFulfilmentRequestedAccuracyFulfilled    = gsm_map.AccuracyFulfilmentIndicatorRequestedAccuracyFulfilled
	AccuracyFulfilmentRequestedAccuracyNotFulfilled = gsm_map.AccuracyFulfilmentIndicatorRequestedAccuracyNotFulfilled
)

// SupportedGADShapes (BIT STRING SIZE 7..16) per TS 29.002
// MAP-LCS-DataTypes.asn:280. 7 named bits per 3GPP TS 23.032.
// Surfaced as a bools-only struct following the package BIT STRING pattern.
type SupportedGADShapes struct {
	EllipsoidPoint                                  bool // bit 0
	EllipsoidPointWithUncertaintyCircle             bool // bit 1
	EllipsoidPointWithUncertaintyEllipse            bool // bit 2
	Polygon                                         bool // bit 3
	EllipsoidPointWithAltitude                      bool // bit 4
	EllipsoidPointWithAltitudeAndUncertaintyEllipsoid bool // bit 5
	EllipsoidArc                                    bool // bit 6
}

// LCSPriority (OCTET STRING SIZE 1) per TS 29.002 MAP-LCS-DataTypes.asn:232.
// Per spec: 0 = highest, 1 = normal, all other values treated as 1.
type LCSPriority = HexBytes

// LCSReferenceNumber (OCTET STRING SIZE 1) per TS 29.002
// MAP-CommonDataTypes.asn — single-octet PSL/SLR correlation reference.
type LCSReferenceNumber = HexBytes

// LCSCodewordStringMaxLen is the maxLCSCodewordStringLength constant
// from TS 29.002 MAP-LCS-DataTypes.asn:300.
const LCSCodewordStringMaxLen = 20

// NameStringMaxLen is the maxNameStringLength constant from TS 29.002
// MAP-LCS-DataTypes.asn:212.
const NameStringMaxLen = 63

// RequestorIDStringMaxLen is the maxRequestorIDStringLength constant
// from TS 29.002 MAP-LCS-DataTypes.asn:222.
const RequestorIDStringMaxLen = 63

// ============================================================================
// PSL geographical / positioning data types (TS 29.002 MAP-LCS-DataTypes.asn)
// ============================================================================
//
// Second PR of the staged ProvideSubscriberLocation (opCode 83)
// implementation. Lands the OCTET STRING / INTEGER types referenced by
// PSL-Res and used in deferred-MT-LR responses. Contents are opaque
// per the cited 3GPP specs (TS 23.032 for geographical/velocity data,
// TS 49.031 for GERAN/GANSS positioning data, TS 25.413 for UTRAN
// positioning data) — this package preserves them verbatim and leaves
// interpretation to callers.
//
// All types are byte aliases (= HexBytes) or int64 wire surrogates so
// they compose cleanly with the existing public-API patterns
// (extensionContainer-style opaque pass-through).

// ExtGeographicalInformation (OCTET STRING SIZE 1..20) per TS 29.002
// MAP-LCS-DataTypes.asn:462. Carries a 3GPP TS 23.032 geographical
// information element; only a subset of TS 23.032 shapes is allowed
// in this field (see TS 29.002 MAP-LCS-DataTypes.asn:466).
type ExtGeographicalInformation = HexBytes

// AddGeographicalInformation (OCTET STRING SIZE 1..91) per TS 29.002
// MAP-LCS-DataTypes.asn:601. Carries a 3GPP TS 23.032 geographical
// information element; all TS 23.032 shapes are allowed in this field
// (the wider size bound vs Ext-GeographicalInformation reflects that).
type AddGeographicalInformation = HexBytes

// VelocityEstimate (OCTET STRING SIZE 4..7) per TS 29.002
// MAP-LCS-DataTypes.asn:522. Carries a 3GPP TS 23.032 velocity
// description element.
type VelocityEstimate = HexBytes

// PositioningDataInformation (OCTET STRING SIZE 2..10) per TS 29.002
// MAP-LCS-DataTypes.asn:552. GERAN positioning data per 3GPP TS 49.031.
type PositioningDataInformation = HexBytes

// UtranPositioningDataInfo (OCTET STRING SIZE 3..11) per TS 29.002
// MAP-LCS-DataTypes.asn:560. UTRAN positioning data
// (positioningDataDiscriminator + positioningDataSet) per 3GPP TS 25.413.
type UtranPositioningDataInfo = HexBytes

// GeranGANSSpositioningData (OCTET STRING SIZE 2..10) per TS 29.002
// MAP-LCS-DataTypes.asn:568. GERAN GANSS positioning data per
// 3GPP TS 49.031.
type GeranGANSSpositioningData = HexBytes

// UtranGANSSpositioningData (OCTET STRING SIZE 1..9) per TS 29.002
// MAP-LCS-DataTypes.asn:576. UTRAN GANSS positioning data
// (GANSS-PositioningDataSet only) per 3GPP TS 25.413.
type UtranGANSSpositioningData = HexBytes

// UtranAdditionalPositioningData (OCTET STRING SIZE 1..8) per TS 29.002
// MAP-LCS-DataTypes.asn:584. UTRAN Additional-PositioningDataSet only,
// per 3GPP TS 25.413.
type UtranAdditionalPositioningData = HexBytes

// UtranCivicAddress (OCTET STRING) per TS 29.002 MAP-LCS-DataTypes.asn:597.
// CivicAddress only, per 3GPP TS 25.413. The spec puts no explicit size
// bound on the wire; size validation is left to the caller.
type UtranCivicAddress = HexBytes

// UtranBaroPressureMeas (INTEGER 30000..115000) per TS 29.002
// MAP-LCS-DataTypes.asn:592. UTRAN BarometricPressureMeasurement per
// 3GPP TS 25.413. Raw value per the cited spec; no scaling is applied
// here. Aliased from go-asn1's gsm_map.UtranBaroPressureMeas, which is
// int64-backed.
type UtranBaroPressureMeas = gsm_map.UtranBaroPressureMeas

// Size constants for PSL geographical / positioning data fields, per
// TS 29.002 MAP-LCS-DataTypes.asn:518/619/522/552/557/560/565/568/573/
// 576/581/584/589.
//
// Both Min and Max bounds are surfaced explicitly (including Min=1 for
// SIZE(1..N) fields) so the codec PRs can validate without magic
// numbers.
const (
	ExtGeographicalInformationMinLen     = 1
	ExtGeographicalInformationMaxLen     = 20
	AddGeographicalInformationMinLen     = 1
	AddGeographicalInformationMaxLen     = 91
	VelocityEstimateMinLen               = 4
	VelocityEstimateMaxLen               = 7
	PositioningDataInformationMinLen     = 2
	PositioningDataInformationMaxLen     = 10
	UtranPositioningDataInfoMinLen       = 3
	UtranPositioningDataInfoMaxLen       = 11
	GeranGANSSpositioningDataMinLen      = 2
	GeranGANSSpositioningDataMaxLen      = 10
	UtranGANSSpositioningDataMinLen      = 1
	UtranGANSSpositioningDataMaxLen      = 9
	UtranAdditionalPositioningDataMinLen = 1
	UtranAdditionalPositioningDataMaxLen = 8

	// UtranBaroPressureMeas range bounds (TS 29.002 MAP-LCS-DataTypes.asn:592).
	// Typed as UtranBaroPressureMeas so future range checks compose without
	// explicit casts even if the alias is later replaced by a defined type.
	UtranBaroPressureMeasMin UtranBaroPressureMeas = 30000
	UtranBaroPressureMeasMax UtranBaroPressureMeas = 115000
)

// ============================================================================
// PSL area-event / periodic / reporting-PLMN / serving-node types
// (TS 29.002 MAP-LCS-DataTypes.asn)
// ============================================================================
//
// Third PR of the staged ProvideSubscriberLocation (opCode 83)
// implementation. Lands the SEQUENCE/CHOICE/ENUMERATED types referenced
// by PSL-Arg's areaEventInfo, periodicLDRInfo, and reportingPLMNList
// fields, and PSL-Res's targetServingNodeForHandover CHOICE.
//
// Top-level Arg/Res structs and codec arrive in subsequent PRs.

// AreaType (ENUMERATED) per TS 29.002 MAP-LCS-DataTypes.asn:337.
// Extensible enum. Aliased from go-asn1.
type AreaType = gsm_map.AreaType

const (
	AreaTypeCountryCode    = gsm_map.AreaTypeCountryCode
	AreaTypePlmnId         = gsm_map.AreaTypePlmnId
	AreaTypeLocationAreaId = gsm_map.AreaTypeLocationAreaId
	AreaTypeRoutingAreaId  = gsm_map.AreaTypeRoutingAreaId
	AreaTypeCellGlobalId   = gsm_map.AreaTypeCellGlobalId
	AreaTypeUtranCellId    = gsm_map.AreaTypeUtranCellId
)

// AreaIdentification (OCTET STRING SIZE 2..7) per TS 29.002
// MAP-LCS-DataTypes.asn:346. Internal structure per the spec comment
// (MCC/MNC/LAC/CI etc., depending on AreaType).
type AreaIdentification = HexBytes

// Area (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:332.
type Area struct {
	AreaType           AreaType           // [0] mandatory
	AreaIdentification AreaIdentification // [1] mandatory, 2..7 octets
}

// AreaList (SEQUENCE SIZE 1..10 OF Area) per TS 29.002
// MAP-LCS-DataTypes.asn:328 (maxNumOfAreas).
type AreaList []Area

// AreaDefinition (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:324.
type AreaDefinition struct {
	AreaList AreaList // [0] mandatory, 1..10 entries
}

// OccurrenceInfo (ENUMERATED) per TS 29.002 MAP-LCS-DataTypes.asn:361.
// Extensible enum. Aliased from go-asn1.
type OccurrenceInfo = gsm_map.OccurrenceInfo

const (
	OccurrenceOneTimeEvent      = gsm_map.OccurrenceInfoOneTimeEvent
	OccurrenceMultipleTimeEvent = gsm_map.OccurrenceInfoMultipleTimeEvent
)

// IntervalTime (INTEGER 1..32767) per TS 29.002 MAP-LCS-DataTypes.asn:366.
// Minimum interval time between area reports, in seconds. Aliased from
// go-asn1 to int64.
type IntervalTime = gsm_map.IntervalTime

// AreaEventInfo (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:318.
type AreaEventInfo struct {
	AreaDefinition AreaDefinition  // [0] mandatory
	OccurrenceInfo *OccurrenceInfo // [1] optional
	IntervalTime   *IntervalTime   // [2] optional, 1..32767 seconds
}

// ReportingAmount (INTEGER 1..8639999) per TS 29.002
// MAP-LCS-DataTypes.asn:380 (maxReportingAmount). Aliased from go-asn1
// to int64.
type ReportingAmount = gsm_map.ReportingAmount

// ReportingInterval (INTEGER 1..8639999) per TS 29.002
// MAP-LCS-DataTypes.asn:384 (maxReportingInterval). Value is in seconds.
// Aliased from go-asn1 to int64.
type ReportingInterval = gsm_map.ReportingInterval

// PeriodicLDRInfo (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:369.
//
// Per spec: ReportingInterval × ReportingAmount must not exceed
// 8639999 (99 days, 23 hours, 59 minutes, 59 seconds) for compatibility
// with OMA MLP and RLP. Validation lives in the codec PR.
//
// Note: the ASN.1 definition includes an optional
// reportingOptionMilliseconds at tag [0] past the extensibility marker;
// not surfaced by this API. It is dropped on decode and emitted as
// absent on encode — callers requiring millisecond-resolution
// reporting intervals must add it at a higher layer.
type PeriodicLDRInfo struct {
	ReportingAmount   ReportingAmount   // mandatory, 1..8639999
	ReportingInterval ReportingInterval // mandatory, 1..8639999 seconds
}

// RANTechnology (ENUMERATED) per TS 29.002 MAP-LCS-DataTypes.asn:420.
// Extensible enum. Aliased from go-asn1.
type RANTechnology = gsm_map.RANTechnology

const (
	RANTechnologyGsm  = gsm_map.RANTechnologyGsm
	RANTechnologyUmts = gsm_map.RANTechnologyUmts
)

// ReportingPLMN (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:414.
// PlmnId is a 3-octet PLMN-Id per TS 23.003.
type ReportingPLMN struct {
	PlmnId                     HexBytes       // [0] mandatory, 3 octets
	RanTechnology              *RANTechnology // [1] optional
	RanPeriodicLocationSupport bool           // [2] optional NULL; true when present, false when absent
}

// PLMNList (SEQUENCE SIZE 1..20 OF ReportingPLMN) per TS 29.002
// MAP-LCS-DataTypes.asn:409 (maxNumOfReportingPLMN).
type PLMNList []ReportingPLMN

// ReportingPLMNList (SEQUENCE) per TS 29.002 MAP-LCS-DataTypes.asn:404.
type ReportingPLMNList struct {
	PlmnListPrioritized bool     // [0] optional NULL; true when present, false when absent
	PlmnList            PLMNList // [1] mandatory, 1..20 entries
}

// TerminationCause (ENUMERATED) per TS 29.002 MAP-LCS-DataTypes.asn:696.
// Extensible enum. Aliased from go-asn1.
type TerminationCause = gsm_map.TerminationCause

const (
	TerminationNormal                              = gsm_map.TerminationCauseNormal
	TerminationErrorundefined                      = gsm_map.TerminationCauseErrorundefined
	TerminationInternalTimeout                     = gsm_map.TerminationCauseInternalTimeout
	TerminationCongestion                          = gsm_map.TerminationCauseCongestion
	TerminationMtLrRestart                         = gsm_map.TerminationCauseMtLrRestart
	TerminationPrivacyViolation                    = gsm_map.TerminationCausePrivacyViolation
	TerminationShapeOfLocationEstimateNotSupported = gsm_map.TerminationCauseShapeOfLocationEstimateNotSupported
	TerminationSubscriberTermination               = gsm_map.TerminationCauseSubscriberTermination
	TerminationUETermination                       = gsm_map.TerminationCauseUETermination
	TerminationNetworkTermination                  = gsm_map.TerminationCauseNetworkTermination
)

// ServingNodeAddress (CHOICE) per TS 29.002 MAP-LCS-DataTypes.asn (used
// in PSL-Res targetServingNodeForHandover field). Set exactly one of
// MscNumber, SgsnNumber, or MmeNumber.
//
// Following the existing CHOICE pattern (AdditionalNumber,
// CancelLocationIdentity), the selected alternative is inferred from
// which field is set: empty `MscNumber` digits, empty `SgsnNumber`
// digits, and nil/empty `MmeNumber` mean "absent". Set exactly one.
//
// MscNumber and SgsnNumber are ISDN-AddressString digits + Nature/Plan
// triples (consistent with the rest of the public API). MmeNumber is a
// DiameterIdentity (FQDN, 9..255 octets per RFC 6733). The field name
// matches the ASN.1 spec literal `mme-Number [2] DiameterIdentity`
// even though the type is a name/FQDN — this preserves the
// match-upstream-spec convention used elsewhere in the package.
type ServingNodeAddress struct {
	MscNumber        string // ISDN-AddressString digits; "" = alternative not selected
	MscNumberNature  uint8
	MscNumberPlan    uint8
	SgsnNumber       string // ISDN-AddressString digits; "" = alternative not selected
	SgsnNumberNature uint8
	SgsnNumberPlan   uint8
	MmeNumber        HexBytes // DiameterIdentity octets; nil/empty = alternative not selected
}

// Spec-derived size / range constants for PR C types, per TS 29.002
// MAP-LCS-DataTypes.asn:328/330/346/366/380/382/384/387/409/412.
const (
	AreaIdentificationMinLen = 2
	AreaIdentificationMaxLen = 7

	AreaListMinEntries = 1
	AreaListMaxEntries = 10 // maxNumOfAreas

	IntervalTimeMin IntervalTime = 1
	IntervalTimeMax IntervalTime = 32767

	ReportingAmountMin   ReportingAmount   = 1
	ReportingAmountMax   ReportingAmount   = 8639999 // maxReportingAmount
	ReportingIntervalMin ReportingInterval = 1
	ReportingIntervalMax ReportingInterval = 8639999 // maxReportingInterval

	// PeriodicLDRInfo combined cap: ReportingInterval × ReportingAmount
	// must not exceed this value (99 days, 23 hours, 59 minutes, 59
	// seconds) for compatibility with OMA MLP and RLP.
	PeriodicLDRProductMax int64 = 8639999

	PLMNListMinEntries = 1
	PLMNListMaxEntries = 20 // maxNumOfReportingPLMN
)

// ============================================================================
// ProvideSubscriberLocationArg — TS 29.002 MAP-LCS-DataTypes.asn:425
// ============================================================================
//
// Top-level PSL-Arg public type, opCode 83. Wires the PSL leaf,
// LCS-Client, and area-event/periodic/PLMN-list converters from PRs
// #43, #44, and #45 into a single public struct. Marshal()/Parse()
// entry points are in marshal.go / parse.go.

// ProvideSubscriberLocationArg represents a ProvideSubscriberLocation
// request (opCode 83) per TS 29.002 MAP-LCS-DataTypes.asn:425.
//
// Mandatory fields: LocationType, MlcNumber (digits string + Nature/Plan
// triple); empty MlcNumber digits are rejected on encode.
//
// Optional fields follow the package-wide conventions:
//   - Optional string fields (e.g., MSISDN, IMSI, IMEI, HGmlcAddress):
//     "" = absent.
//   - Pointer fields: nil = absent.
//   - bool NULL flags: false = absent, true = present.
//
// Note: ExtensionContainer at tag [8] is opaque metadata not surfaced
// to callers (per the package convention; see APNConfiguration). It
// is dropped on decode and emitted as absent on encode — callers
// requiring opaque pass-through must add it at a higher layer.
type ProvideSubscriberLocationArg struct {
	// Mandatory.
	LocationType    LocationType
	MlcNumber       string // ISDN-AddressString digits
	MlcNumberNature uint8  // address nature indicator (default: International when 0)
	MlcNumberPlan   uint8  // numbering plan indicator (default: ISDN when 0)

	// Optional.
	LcsClientID               *LCSClientID
	PrivacyOverride           bool   // [1] NULL flag
	IMSI                      string // TBCD-decoded digits; "" = absent (5..15 BCD digits per TS 29.002, TBCD-STRING SIZE 3..8 octets)
	MSISDN                    string // ISDN-AddressString digits; "" = absent
	MSISDNNature              uint8  // address nature indicator (default: International when 0)
	MSISDNPlan                uint8  // numbering plan indicator (default: ISDN when 0)
	LMSI                      HexBytes // 4 octets opaque
	IMEI                      string   // TBCD-decoded digits; "" = absent (15 BCD digits per TS 29.002)
	LcsPriority               LCSPriority
	LcsQoS                    *LCSQoS
	SupportedGADShapes        *SupportedGADShapes
	LcsReferenceNumber        LCSReferenceNumber
	LcsServiceTypeID          *int64 // 0..127 per LCSServiceTypeID INTEGER
	LcsCodeword               *LCSCodeword
	LcsPrivacyCheck           *LCSPrivacyCheck
	AreaEventInfo             *AreaEventInfo
	HGmlcAddress              string // GSN-Address as IP string (built via gsn.Build); "" = absent
	MoLrShortCircuitIndicator bool   // [16] NULL flag
	PeriodicLDRInfo           *PeriodicLDRInfo
	ReportingPLMNList         *ReportingPLMNList
}

// ============================================================================
// ProvideSubscriberLocationRes — TS 29.002 MAP-LCS-DataTypes.asn:425
// ============================================================================

// ProvideSubscriberLocationRes represents a ProvideSubscriberLocation
// response (opCode 83) per TS 29.002 MAP-LCS-DataTypes.asn:425.
//
// Mandatory: LocationEstimate (Ext-GeographicalInformation, 1..20 octets).
// Optional fields follow the package-wide conventions:
//   - Pointer fields: nil = absent.
//   - HexBytes / string fields: nil/empty = absent.
//   - bool NULL flags: false = absent, true = present.
//
// Note: ExtensionContainer at tag [1] is opaque metadata not surfaced
// to callers (per the package convention; see APNConfiguration). It
// is dropped on decode and emitted as absent on encode.
//
// CellIdOrSai is a CHOICE between CGI/SAI (7 octets) and LAI (5 octets);
// set exactly one of CellGlobalId or LAI on encode (matching the
// existing CSLocationInformation pattern). Empty/nil on both = absent.
type ProvideSubscriberLocationRes struct {
	// Mandatory.
	LocationEstimate ExtGeographicalInformation // 1..20 octets per TS 23.032

	// Optional.
	AgeOfLocationEstimate         *int64                     // [0] minutes since location was acquired
	AddLocationEstimate           AddGeographicalInformation // [2] 1..91 octets; nil/empty = absent
	DeferredmtLrResponseIndicator bool                       // [3] NULL flag
	GeranPositioningData          PositioningDataInformation // [4] 2..10 octets; nil/empty = absent
	UtranPositioningData          UtranPositioningDataInfo   // [5] 3..11 octets; nil/empty = absent

	// CellIdOrSai CHOICE [6] (explicit). Set exactly one of:
	CellGlobalId HexBytes // CGI or SAI fixed-length 7 octets
	LAI          HexBytes // LAI fixed-length 5 octets

	SaiPresent                     bool                          // [7] NULL flag
	AccuracyFulfilmentIndicator    *AccuracyFulfilmentIndicator  // [8] extensible enum
	VelocityEstimate               VelocityEstimate              // [9] 4..7 octets
	MoLrShortCircuitIndicator      bool                          // [10] NULL flag
	GeranGANSSpositioningData      GeranGANSSpositioningData     // [11] 2..10 octets
	UtranGANSSpositioningData      UtranGANSSpositioningData     // [12] 1..9 octets
	TargetServingNodeForHandover   *ServingNodeAddress           // [13] explicit CHOICE
	UtranAdditionalPositioningData UtranAdditionalPositioningData // [14] 1..8 octets
	UtranBaroPressureMeas          *UtranBaroPressureMeas        // [15] INTEGER 30000..115000
	UtranCivicAddress              UtranCivicAddress             // [16] CivicAddress per TS 25.413
}

// ============================================================================
// MAP ReturnError diagnostics (TS 29.002 §17.6 / MAP-ER-DataTypes.asn)
// ============================================================================
//
// TCAP ReturnError carries an opcode (e.g. absentSubscriberSM, callBarred)
// plus a BER-encoded Parameter that holds operationally-important
// diagnostic detail (which network resource broke; why a subscriber is
// absent; etc.). The wrapper-level types below mirror the gsm_map.*Param
// counterparts while keeping diagnostic enums as named upstream types
// so callers can call String() for free, without dropping down to
// gsm_map.* directly.

// MapErrorCode is the typed MAP ReturnError opcode per TS 29.002 §17.6.
// Aliased from the upstream go-asn1 ErrorCode so callers can use either
// the local or upstream constants interchangeably and the existing
// GetErrorString helper continues to delegate to the upstream String()
// method. Values not surfaced as constants below are still valid; they
// can be cast from int64 (e.g. MapErrorCode(48) for orNotAllowed) or
// referenced via gsm_map.<Name>.
type MapErrorCode = gsm_map.ErrorCode

// MAP error opcodes per TS 29.002 §17.6, scoped to the SRI-SM / SRI /
// ATI-relevant subset surfaced by ParseReturnErrorParameter. Aliased
// from the upstream gsm_map.<Name> constants. The full set lives in
// the gsm_map package for callers needing less-common opcodes.
const (
	MapErrorUnknownSubscriber             = gsm_map.UnknownSubscriber             // 1
	MapErrorAbsentSubscriberSM            = gsm_map.AbsentSubscriberSM            // 6
	MapErrorRoamingNotAllowed             = gsm_map.RoamingNotAllowed             // 8
	MapErrorTeleserviceNotProvisioned     = gsm_map.TeleserviceNotProvisioned     // 11
	MapErrorCallBarred                    = gsm_map.CallBarred                    // 13
	MapErrorFacilityNotSupported          = gsm_map.FacilityNotSupported          // 21
	MapErrorAbsentSubscriber              = gsm_map.AbsentSubscriber              // 27
	MapErrorSystemFailure                 = gsm_map.SystemFailure                 // 34
	MapErrorDataMissing                   = gsm_map.DataMissing                   // 35
	MapErrorUnauthorizedRequestingNetwork = gsm_map.UnauthorizedRequestingNetwork // 52
)
//
// Parsers (Parse*Param functions) and the dispatcher
// (ParseReturnErrorParameter) live in parse.go; see follow-up PRs.
//
// Coverage scope is the SRI-SM / SRI / ATI-relevant errors observed
// on roaming networks: absentSubscriberSM, unknownSubscriber,
// callBarred, systemFailure, roamingNotAllowed,
// unauthorizedRequestingNetwork, facilityNotSupported,
// teleserviceNotProvisioned, dataMissing.

// AbsentSubscriberSMParam (SEQUENCE) per TS 29.002 MAP-ER-DataTypes.asn.
// Returned with errorCode 6 (absentSubscriberSM) by SRI-SM and
// MT-ForwardSM. The diagnostic fields explain why the subscriber is
// absent (phone off, out of coverage, purged from HLR, etc.) and
// drive different remediation paths on the caller side.
//
// AbsentSubscriberDiagnosticSM is currently a type alias to int64 in
// upstream go-asn1, so it has no String() method yet. Callers can
// still test specific values via the gsm_map.AbsentSubscriberDiagnosticSM*
// constants.
type AbsentSubscriberSMParam struct {
	AbsentSubscriberDiagnosticSM           *gsm_map.AbsentSubscriberDiagnosticSM // untagged
	AdditionalAbsentSubscriberDiagnosticSM *gsm_map.AbsentSubscriberDiagnosticSM // [0]
	IMSI                                   string                                // [1] TBCD-decoded digits; "" = absent
	RequestedRetransmissionTime            HexBytes                              // [2] opaque GeneralizedTime octets; nil = absent
	UserIdentifierAlert                    string                                // [3] TBCD-decoded digits; "" = absent
}

// UnknownSubscriberParam (SEQUENCE) per TS 29.002 MAP-ER-DataTypes.asn.
// Returned with errorCode 1 (unknownSubscriber). The diagnostic
// distinguishes "we never had this MSISDN" from "this MSISDN is not
// provisioned for the queried service".
type UnknownSubscriberParam struct {
	UnknownSubscriberDiagnostic *gsm_map.UnknownSubscriberDiagnostic
}

// CallBarredParam (CHOICE) per TS 29.002 MAP-ER-DataTypes.asn.
// Returned with errorCode 13 (callBarred). The CHOICE is between a
// bare CallBarringCause (legacy) and the extensible variant. Set
// exactly one of CallBarringCause or ExtensibleCallBarredParam on
// encode; both fields are populated mutually exclusively on decode.
type CallBarredParam struct {
	CallBarringCause          *gsm_map.CallBarringCause  // legacy alternative
	ExtensibleCallBarredParam *ExtensibleCallBarredParam // extensible alternative
}

// ExtensibleCallBarredParam (SEQUENCE) — extensible variant of the
// callBarred CHOICE.
type ExtensibleCallBarredParam struct {
	CallBarringCause              *gsm_map.CallBarringCause // untagged
	UnauthorisedMessageOriginator bool                      // [1] NULL flag
	AnonymousCallRejection        bool                      // [2] NULL flag
}

// SystemFailureParam (CHOICE) per TS 29.002 MAP-ER-DataTypes.asn.
// Returned with errorCode 34 (systemFailure). Identifies which network
// node broke — critical for incident triage. The CHOICE is between a
// bare NetworkResource (legacy) and the extensible variant. Set
// exactly one on encode; both fields are populated mutually
// exclusively on decode.
type SystemFailureParam struct {
	NetworkResource              *gsm_map.NetworkResource      // legacy alternative
	ExtensibleSystemFailureParam *ExtensibleSystemFailureParam // extensible alternative
}

// ExtensibleSystemFailureParam (SEQUENCE) — extensible variant of the
// systemFailure CHOICE.
type ExtensibleSystemFailureParam struct {
	NetworkResource           *gsm_map.NetworkResource           // untagged
	AdditionalNetworkResource *gsm_map.AdditionalNetworkResource // [0]
	FailureCauseParam         *gsm_map.FailureCauseParam         // [1]
}

// RoamingNotAllowedParam (SEQUENCE) per TS 29.002 MAP-ER-DataTypes.asn.
// Returned with errorCode 8 (roamingNotAllowed). Mandatory cause +
// optional additional cause distinguish PLMN-roaming-not-allowed from
// operator-determined-barring.
type RoamingNotAllowedParam struct {
	RoamingNotAllowedCause           gsm_map.RoamingNotAllowedCause            // untagged, mandatory
	AdditionalRoamingNotAllowedCause *gsm_map.AdditionalRoamingNotAllowedCause // [0]
}

// UnauthorizedRequestingNetworkParam (SEQUENCE) per TS 29.002
// MAP-ER-DataTypes.asn. Returned with errorCode 52
// (unauthorizedRequestingNetwork). Carries only ExtensionContainer
// in the spec; the public type is empty (placeholder for opaque
// pass-through callers).
type UnauthorizedRequestingNetworkParam struct{}

// FacilityNotSupParam (SEQUENCE) per TS 29.002 MAP-ER-DataTypes.asn.
// Returned with errorCode 21 (facilityNotSupported). Optional
// indicators identify which sub-facility is unsupported.
type FacilityNotSupParam struct {
	ShapeOfLocationEstimateNotSupported          bool // [0] NULL flag
	NeededLcsCapabilityNotSupportedInServingNode bool // [1] NULL flag
}

// TeleservNotProvParam (SEQUENCE) per TS 29.002 MAP-ER-DataTypes.asn.
// Returned with errorCode 11 (teleserviceNotProvisioned). Carries
// only ExtensionContainer in the spec; the public type is empty.
type TeleservNotProvParam struct{}

// DataMissingParam (SEQUENCE) per TS 29.002 MAP-ER-DataTypes.asn.
// Returned with errorCode 35 (dataMissing). Carries only
// ExtensionContainer in the spec; the public type is empty.
type DataMissingParam struct{}

// AbsentSubscriberParam (SEQUENCE) per TS 29.002 MAP-ER-DataTypes.asn.
// Returned with errorCode 27 (absentSubscriber) by SRI and PSI on the
// GSM CS side. Distinct from AbsentSubscriberSMParam (errorCode 6),
// which is the SMS-side variant. The optional AbsentSubscriberReason
// distinguishes imsiDetach / pageReceiveFailure / etc.
type AbsentSubscriberParam struct {
	AbsentSubscriberReason *gsm_map.AbsentSubscriberReason // [0]
}

// ============================================================================
// SGSN-CAMEL-SubscriptionInfo (TS 29.002 MAP-MS-DataTypes.asn:1596)
// ============================================================================

// GPRSTriggerDetectionPoint (ENUMERATED) per TS 29.002
// MAP-MS-DataTypes.asn (extensible enum). Aliased from go-asn1.
type GPRSTriggerDetectionPoint = gsm_map.GPRSTriggerDetectionPoint

const (
	GPRSTDPAttach                                 = gsm_map.GPRSTriggerDetectionPointAttach
	GPRSTDPAttachChangeOfPosition                 = gsm_map.GPRSTriggerDetectionPointAttachChangeOfPosition
	GPRSTDPPdpContextEstablishment                = gsm_map.GPRSTriggerDetectionPointPdpContextEstablishment
	GPRSTDPPdpContextEstablishmentAcknowledgement = gsm_map.GPRSTriggerDetectionPointPdpContextEstablishmentAcknowledgement
	GPRSTDPPdpContextChangeOfPosition             = gsm_map.GPRSTriggerDetectionPointPdpContextChangeOfPosition
)

// DefaultGPRSHandling (ENUMERATED) per TS 29.002 MAP-MS-DataTypes.asn:1634.
// Per spec exception clause, decoders MUST treat values >1 as
// releaseTransaction. Aliased from go-asn1; the lenient remap happens
// in the decoder.
type DefaultGPRSHandling = gsm_map.DefaultGPRSHandling

const (
	DefaultGPRSContinueTransaction = gsm_map.DefaultGPRSHandlingContinueTransaction
	DefaultGPRSReleaseTransaction  = gsm_map.DefaultGPRSHandlingReleaseTransaction
)

// GPRSCamelTDPData (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1625.
// All four fields are mandatory per spec.
type GPRSCamelTDPData struct {
	GprsTriggerDetectionPoint GPRSTriggerDetectionPoint // [0] mandatory
	ServiceKey                int64                     // [1] mandatory, 0..2147483647 per CAMEL convention
	GsmSCFAddress             string                    // [2] mandatory ISDN-AddressString digits
	GsmSCFAddressNature       uint8
	GsmSCFAddressPlan         uint8
	DefaultSessionHandling    DefaultGPRSHandling // [3] mandatory
}

// GPRSCamelTDPDataList (SEQUENCE SIZE 1..10 OF GPRS-CamelTDPData) per
// TS 29.002 MAP-MS-DataTypes.asn:1620.
type GPRSCamelTDPDataList []GPRSCamelTDPData

// GPRSCSI (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:1606.
// Per spec clause 8.8.x, when GPRSCSI is present both
// GprsCamelTDPDataList and CamelCapabilityHandling SHALL be set;
// otherwise all fields are optional.
type GPRSCSI struct {
	GprsCamelTDPDataList    GPRSCamelTDPDataList // [0] optional, 1..10 entries when present
	CamelCapabilityHandling *int                 // [1] optional, CAMEL phase 1..4
	NotificationToCSE       bool                 // [3] optional NULL — true when present
	CsiActive               bool                 // [4] optional NULL — true when present
}

// MGCSI (SEQUENCE) per TS 29.002 MAP-MS-DataTypes.asn:2528.
// MobilityTriggers SIZE 1..10, each entry MM-Code SIZE 1.
type MGCSI struct {
	MobilityTriggers    []HexBytes // mandatory, 1..10 entries; each MM-Code SIZE 1
	ServiceKey          int64      // mandatory, 0..2147483647 per CAMEL convention
	GsmSCFAddress       string     // [0] mandatory ISDN-AddressString digits
	GsmSCFAddressNature uint8
	GsmSCFAddressPlan   uint8
	NotificationToCSE   bool // [2] optional NULL — true when present
	CsiActive           bool // [3] optional NULL — true when present
}

// SGSNCAMELSubscriptionInfo (SEQUENCE) per TS 29.002
// MAP-MS-DataTypes.asn:1596. All fields optional.
type SGSNCAMELSubscriptionInfo struct {
	GprsCSI                   *GPRSCSI                // [0] optional
	MoSmsCSI                  *SMSCSI                 // [1] optional, reuses PR C type
	MtSmsCSI                  *SMSCSI                 // [3] optional, reuses PR C type
	MtSmsCAMELTDPCriteriaList []MTSmsCAMELTDPCriteria // [4] optional, reuses PR C type
	MgCsi                     *MGCSI                  // [5] optional
}

// ============================================================================
// InsertSubscriberData (opCode 7) — TS 29.002 MAP-MS-DataTypes.asn:1738
// ============================================================================

// InsertSubscriberDataArg (SEQUENCE) per TS 29.002. All fields are
// OPTIONAL per spec. The sub-types covered by PRs E1a/E1b1/E1b2/E1b3
// are referenced here with their public Go equivalents; opaque scalar
// fields are surfaced as `HexBytes` or typed integers.
type InsertSubscriberDataArg struct {
	// Identification (typically present together)
	IMSI                                           HexBytes // [0] optional, IMSI octets (TBCD-encoded)
	MSISDN                                         string   // [1] optional, ISDN-AddressString digits ("" = absent)
	MSISDNNature                                   uint8
	MSISDNPlan                                     uint8

	// Subscriber profile
	Category         HexBytes          // [2] optional, OCTET STRING SIZE 1
	SubscriberStatus *SubscriberStatus // [3] optional

	// Service lists
	BearerServiceList []HexBytes // [4] optional, list of Ext-BearerServiceCode (1..5 octets each)
	TeleserviceList   []HexBytes // [6] optional, list of Ext-TeleserviceCode (1..5 octets each)
	ProvisionedSS     []ExtSSInfo // [7] optional, ExtSSInfoList

	// Operator-determined barring
	OdbData                                   *ODBData // [8] optional
	RoamingRestrictionDueToUnsupportedFeature bool     // [9] optional NULL

	// Voice services and zones
	RegionalSubscriptionData ZoneCodeList // [10] optional, 1..10 entries
	VbsSubscriptionData      VBSDataList  // [11] optional, 1..50 entries
	VgcsSubscriptionData     VGCSDataList // [12] optional, 1..50 entries

	// CAMEL VLR-side
	VlrCamelSubscriptionInfo *VlrCamelSubscriptionInfo // [13] optional

	// PR E1a sub-types
	NaeaPreferredCI *NaeaPreferredCI // [15] optional

	// PR E1b1 sub-types
	GprsSubscriptionData                           *GPRSSubscriptionData // [16] optional
	RoamingRestrictedInSgsnDueToUnsupportedFeature bool                  // [23] optional NULL
	NetworkAccessMode                              *NetworkAccessMode    // [24] optional
	LsaInformation                                 *LSAInformation       // [25] optional

	// LCS / IST / supercharger
	LmuIndicator               bool                       // [21] optional NULL
	LcsInformation             *LCSInformation            // [22] optional
	IstAlertTimer              *int64                     // [26] optional, ISTAlertTimerValue
	SuperChargerSupportedInHLR HexBytes                   // [27] optional, AgeIndicator OCTET STRING (SIZE 1..6)
	McSSInfo                   *MCSSInfo                  // [28] optional
	CsAllocationRetentionPriority HexBytes                // [29] optional, OCTET STRING SIZE 1
	SgsnCAMELSubscriptionInfo  *SGSNCAMELSubscriptionInfo // [17] optional

	// Charging / access restriction
	ChargingCharacteristics HexBytes               // [18] optional, OCTET STRING SIZE 2
	AccessRestrictionData   *AccessRestrictionData // [19] optional
	IcsIndicator            *bool                  // [20] optional

	// EPS subscription
	EpsSubscriptionData     *EPSSubscriptionData    // [31] optional
	CsgSubscriptionDataList CSGSubscriptionDataList // [32] optional, 1..50 entries

	// Reachability + network identifiers
	UeReachabilityRequestIndicator bool   // [33] optional NULL
	SgsnNumber                     string // [34] optional ISDN-AddressString
	SgsnNumberNature               uint8
	SgsnNumberPlan                 uint8
	MmeName                        HexBytes // [35] optional DiameterIdentity (FQDN)

	SubscribedPeriodicRAUTAUtimer *int64 // [36] optional INTEGER
	VplmnLIPAAllowed              bool   // [37] optional NULL
	MdtUserConsent                *bool  // [38] optional
	SubscribedPeriodicLAUtimer    *int64 // [39] optional INTEGER

	// VPLMN CSG / additional MSISDN
	VplmnCsgSubscriptionDataList VPLMNCSGSubscriptionDataList // [40] optional, 1..50 entries
	AdditionalMSISDN             string                       // [41] optional ISDN-AddressString
	AdditionalMSISDNNature       uint8
	AdditionalMSISDNPlan         uint8

	// Service provision flags (NULL)
	PsAndSMSOnlyServiceProvision bool // [42]
	SmsInSGSNAllowed             bool // [43]
	CsToPsSRVCCAllowedIndicator  bool // [44]
	PcscfRestorationRequest      bool // [45]

	// PR E1a sub-types (continued)
	AdjacentAccessRestrictionDataList AdjacentAccessRestrictionDataList // [46] optional, 1..50 entries
	ImsiGroupIdList                   IMSIGroupIdList                   // [47] optional, 1..50 entries

	UeUsageType                           HexBytes            // [48] optional OCTET STRING
	UserPlaneIntegrityProtectionIndicator bool                // [49] optional NULL
	DlBufferingSuggestedPacketCount       *int64              // [50] optional INTEGER
	ResetIdList                           ResetIdList         // [51] optional, 1..50 entries
	EDRXCycleLengthList                   EDRXCycleLengthList // [52] optional, 1..8 entries

	ExtAccessRestrictionData     *ExtAccessRestrictionData // [53] optional
	IabOperationAllowedIndicator bool                      // [54] optional NULL
}

// InsertSubscriberDataRes (SEQUENCE) per TS 29.002. All fields are
// OPTIONAL per spec. Returns the HLR's view of the subscriber's
// services + ODB + supported CAMEL phases / features.
type InsertSubscriberDataRes struct {
	TeleserviceList              []HexBytes                    // [1] optional, list of Ext-TeleserviceCode
	BearerServiceList            []HexBytes                    // [2] optional, list of Ext-BearerServiceCode
	SsList                       []SsCode                      // [3] optional, list of SS-Code
	OdbGeneralData               *ODBGeneralData               // [4] optional
	RegionalSubscriptionResponse *RegionalSubscriptionResponse // [5] optional
	SupportedCamelPhases         *SupportedCamelPhases         // [6] optional
	OfferedCamel4CSIs            *OfferedCamel4CSIs            // [8] optional
	SupportedFeatures            *SupportedFeatures            // [9] optional
	ExtSupportedFeatures         *ExtSupportedFeatures         // [10] optional
}

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
	ErrMCSSInfoSsCodeInvalidSize = errors.New("mcSSInfo: SsCode must be exactly 1 octet per TS 29.002 (mandatory tag [0])")

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

	ErrPDPContextIdOutOfRange       = errors.New("pdpContext: PdpContextId must be 1..50 (maxNumOfPDP-Contexts) per TS 29.002")
	ErrPDPTypeInvalidSize           = errors.New("pdpContext: PdpType must be exactly 2 octets per TS 29.002 MAP-MS-DataTypes.asn:1657")
	ErrQoSSubscribedInvalidSize     = errors.New("pdpContext: QosSubscribed must be exactly 3 octets per TS 29.002 MAP-MS-DataTypes.asn:1673 (mandatory tag [18])")
	ErrExtQoSSubscribedInvalidSize  = errors.New("pdpContext: ExtQoSSubscribed must be 1..9 octets per TS 29.002 MAP-MS-DataTypes.asn:1677")
	ErrExt2QoSSubscribedInvalidSize = errors.New("pdpContext: Ext2QoSSubscribed must be 1..3 octets per TS 29.002 MAP-MS-DataTypes.asn:1685")
	ErrExt3QoSSubscribedInvalidSize = errors.New("pdpContext: Ext3QoSSubscribed must be 1..2 octets per TS 29.002 MAP-MS-DataTypes.asn:1690")
	ErrExt4QoSSubscribedInvalidSize = errors.New("pdpContext: Ext4QoSSubscribed must be exactly 1 octet per TS 29.002 MAP-MS-DataTypes.asn:1693")
	ErrExtQoSHierarchyViolated      = errors.New("pdpContext: Ext{2,3,4}-QoS-Subscribed must follow the spec hierarchy per TS 29.002 MAP-MS-DataTypes.asn:1534-1538 (Ext2 requires Ext, Ext3 requires Ext2, Ext4 requires Ext3)")
	ErrExtPDPAddressWithoutPDPAddress = errors.New("pdpContext: ExtPdpAddress may be present only if PdpAddress is present per TS 29.002 MAP-MS-DataTypes.asn:1549")
	ErrExtPDPTypeInvalidSize        = errors.New("pdpContext: ExtPdpType must be exactly 2 octets per TS 29.002 MAP-MS-DataTypes.asn:1661")
	ErrPDPAddressInvalidSize      = errors.New("pdpContext: PdpAddress must be 1..16 octets per TS 29.002 MAP-MS-DataTypes.asn:1665")
	ErrExtPDPAddressInvalidSize   = errors.New("pdpContext: ExtPdpAddress must be 1..16 octets per TS 29.002 MAP-MS-DataTypes.asn:1665 (PDP-Address)")
	ErrPDPChargingCharsInvalidSize = errors.New("pdpContext: PdpChargingCharacteristics must be exactly 2 octets per TS 29.002")
	ErrAPNOIReplacementInvalidSize = errors.New("apnOIReplacement: must be 9..100 octets per TS 29.002 MAP-MS-DataTypes.asn:1303")
	ErrFQDNInvalidSize            = errors.New("fqdn: must be 9..255 octets per TS 29.002 MAP-MS-DataTypes.asn:1434")
	ErrRestorationPriorityInvalidSize = errors.New("pdpContext: RestorationPriority must be exactly 1 octet per TS 29.002")
	ErrGPRSDataListSize           = errors.New("gprsDataList: must contain 1..50 entries (maxNumOfPDP-Contexts) per TS 29.002")
	ErrGPRSSubscriptionDataMissingList = errors.New("gprsSubscriptionData: GprsDataList is mandatory and must contain at least one entry")
	ErrAMBRBandwidthOutOfRange    = errors.New("ambr: bandwidth fields must be non-negative")
	ErrSIPTOPermissionInvalid     = errors.New("pdpContext: SiptoPermission must be siptoAboveRanAllowed(0) or siptoAboveRanNotAllowed(1)")
	ErrSIPTOLocalNetworkPermissionInvalid = errors.New("pdpContext: SiptoLocalNetworkPermission must be siptoAtLocalNetworkAllowed(0) or siptoAtLocalNetworkNotAllowed(1)")
	ErrLIPAPermissionInvalid      = errors.New("pdpContext: LipaPermission must be lipaProhibited(0), lipaOnly(1), or lipaConditional(2)")
	ErrNIDDMechanismInvalid       = errors.New("pdpContext: NIDDMechanism must be sGi-based-data-delivery(0) or sCEF-based-data-delivery(1)")

	ErrLSAIdentityInvalidSize       = errors.New("lsaData: LsaIdentity must be exactly 3 octets per TS 29.002 MAP-MS-DataTypes.asn:1728")
	ErrLSAAttributesInvalidSize     = errors.New("lsaData: LsaAttributes must be exactly 1 octet per TS 29.002 MAP-MS-DataTypes.asn:1731")
	ErrLSADataListSize              = errors.New("lsaDataList: must contain 1..20 entries (maxNumOfLSAs) per TS 29.002")
	ErrLSAOnlyAccessIndicatorInvalid = errors.New("lsaInformation: LsaOnlyAccessIndicator must be accessOutsideLSAsAllowed(0) or accessOutsideLSAsRestricted(1)")

	ErrPDNTypeInvalidSize           = errors.New("apnConfiguration: PdnType must be exactly 1 octet per TS 29.002 MAP-MS-DataTypes.asn:1369")
	ErrQoSClassIdentifierOutOfRange = errors.New("epsQoSSubscribed: QosClassIdentifier must be 1..9 per TS 29.002 MAP-MS-DataTypes.asn:1415")
	ErrRFSPIDOutOfRange             = errors.New("epsSubscriptionData: RfspId must be 1..MaxRFSPID (256) per TS 29.002 MAP-MS-DataTypes.asn:1306")
	ErrPDNGWAllocationTypeInvalid   = errors.New("apnConfiguration: PdnGwAllocationType must be static(0) or dynamic(1) per TS 29.002 MAP-MS-DataTypes.asn:1437")
	ErrPDNConnectionContinuityInvalid = errors.New("apnConfiguration: PdnConnectionContinuity must be 0..2 per TS 29.002 MAP-MS-DataTypes.asn:1356")
	ErrWLANOffloadabilityIndicationInvalid = errors.New("wlanOffloadability: WLAN-Offloadability-Indication must be notAllowed(0) or allowed(1)")
	ErrSpecificAPNInfoListSize      = errors.New("specificAPNInfoList: must contain 1..50 entries (maxNumOfSpecificAPNInfos) per TS 29.002")
	ErrEPSDataListSize              = errors.New("epsDataList: must contain 1..50 entries (maxNumOfAPN-Configurations) per TS 29.002")
	ErrAPNConfigurationProfileMissingList = errors.New("apnConfigurationProfile: EpsDataList is mandatory and must contain at least one entry")

	ErrGMLCListSize                       = errors.New("gmlcList: must contain 1..5 entries (maxNumOfGMLC) per TS 29.002")
	ErrLCSPrivacyExceptionListSize        = errors.New("lcsPrivacyExceptionList: must contain 1..4 entries (maxNumOfPrivacyClass) per TS 29.002")
	ErrExternalClientListSize             = errors.New("externalClientList: must contain 0..5 entries (maxNumOfExternalClient) per TS 29.002")
	ErrPLMNClientListSize                 = errors.New("plmnClientList: must contain 1..5 entries (maxNumOfPLMNClient) per TS 29.002")
	ErrExtExternalClientListSize          = errors.New("extExternalClientList: must contain 1..35 entries (maxNumOfExt-ExternalClient) per TS 29.002")
	ErrServiceTypeListSize                = errors.New("serviceTypeList: must contain 1..32 entries (maxNumOfServiceType) per TS 29.002")
	ErrMOLRListSize                       = errors.New("molrList: must contain 1..3 entries (maxNumOfMOLR-Class) per TS 29.002")
	ErrGMLCRestrictionInvalid             = errors.New("externalClient: GmlcRestriction must be gmlcList(0) or homeCountry(1)")
	ErrNotificationToMSUserInvalid        = errors.New("notificationToMSUser: must be 0..3 per TS 29.002 MAP-MS-DataTypes.asn:2035")
	ErrLCSClientInternalIDInvalid         = errors.New("plmnClientList: LCSClientInternalID must be 0..4 per TS 29.002 MAP-CommonDataTypes.asn")
	ErrServiceTypeIdentityRange           = errors.New("serviceType: ServiceTypeIdentity must be 0..127 per TS 29.002 MAP-CommonDataTypes.asn:436 (LCSServiceTypeID INTEGER (0..127))")
	ErrLCSPrivacyClassSsCodeInvalidSize   = errors.New("lcsPrivacyClass: SsCode must be exactly 1 octet per TS 29.002 (mandatory SS-Code)")
	ErrMOLRClassSsCodeInvalidSize         = errors.New("molrClass: SsCode must be exactly 1 octet per TS 29.002 (mandatory SS-Code)")
	ErrGMLCAddressEmpty                   = errors.New("gmlcAddress: Address is mandatory; empty digits are not permitted on encode or decode")
	ErrSGSNMtSmsCAMELTDPCriteriaListSize  = errors.New("sgsnCAMELSubscriptionInfo: MtSmsCAMELTDPCriteriaList must contain 1..10 entries (maxNumOfCamelTDPData) per TS 29.002 MAP-MS-DataTypes.asn:2199")

	ErrIsdArgNil                       = errors.New("insertSubscriberDataArg: argument must not be nil")
	ErrIsdResNil                       = errors.New("insertSubscriberDataRes: argument must not be nil")
	ErrIsdCategoryInvalidSize          = errors.New("insertSubscriberDataArg: Category must be exactly 1 octet per TS 29.002")
	ErrIsdChargingCharsInvalidSize     = errors.New("insertSubscriberDataArg: ChargingCharacteristics must be exactly 2 octets per TS 29.002")
	ErrIsdCsAllocRetentionInvalidSize  = errors.New("insertSubscriberDataArg: CsAllocationRetentionPriority must be exactly 1 octet per TS 29.002")
	ErrIsdAgeIndicatorInvalidSize      = errors.New("insertSubscriberDataArg: SuperChargerSupportedInHLR (AgeIndicator) must be 1..6 octets per TS 29.002")
	ErrIsdBearerServiceCodeSize        = errors.New("insertSubscriberDataArg: each Ext-BearerServiceCode must be 1..5 octets per TS 29.002")
	ErrIsdTeleserviceCodeSize          = errors.New("insertSubscriberDataArg: each Ext-TeleserviceCode must be 1..5 octets per TS 29.002")
	ErrIsdBearerServiceListSize        = errors.New("insertSubscriberDataArg: BearerServiceList must contain 1..50 entries (maxNumOfBearerServices) per TS 29.002")
	ErrIsdTeleserviceListSize          = errors.New("insertSubscriberDataArg: TeleserviceList must contain 1..20 entries (maxNumOfTeleservices) per TS 29.002")
	ErrIsdProvisionedSSListSize        = errors.New("insertSubscriberDataArg: ProvisionedSS must contain 1..30 entries (maxNumOfSS) per TS 29.002 MAP-MS-DataTypes.asn:1508")
	ErrIsdResSsListSize                = errors.New("insertSubscriberDataRes: SsList entries must each be exactly 1 octet (SS-Code) per TS 29.002")
	ErrIsdMSISDNDecodedEmpty           = errors.New("insertSubscriberDataArg: present wire ISDN-AddressString decoded to empty digits; presence cannot round-trip through string-based API")
	ErrIsdIMSIInvalidSize              = errors.New("insertSubscriberDataArg: IMSI must be 3..8 octets per TS 29.002 MAP-CommonDataTypes.asn:327 (TBCD-STRING SIZE 3..8)")

	ErrGPRSCamelTDPDataListSize           = errors.New("gprsCamelTDPDataList: must contain 1..10 entries (maxNumOfCamelTDPData) per TS 29.002")
	ErrGPRSTriggerDetectionPointInvalid   = errors.New("gprsCamelTDPData: GprsTriggerDetectionPoint must be 1, 2, 11, 12, or 14 per TS 29.002 (extensible enum: unknown values preserved on decode)")
	ErrDefaultGPRSHandlingInvalid         = errors.New("gprsCamelTDPData: DefaultSessionHandling encoder requires continueTransaction(0) or releaseTransaction(1); decoder applies spec exception clause TS 29.002 MAP-MS-DataTypes.asn:1638-1640 (values 2..31 → continueTransaction; >31 → releaseTransaction)")
	ErrCamelCapabilityHandlingOutOfRange  = errors.New("gprsCSI/mgCSI: CamelCapabilityHandling must be 1..4 per TS 29.078")
	ErrGPRSCSIRequiresTDPListAndPhase     = errors.New("gprsCSI: when GPRS-CSI is present, GprsCamelTDPDataList AND CamelCapabilityHandling SHALL both be present per TS 29.002 MAP-MS-DataTypes.asn:1615-1616")
	ErrMobilityTriggersSize               = errors.New("mgCSI: MobilityTriggers must contain 1..10 entries (maxNumOfMobilityTriggers) per TS 29.002")
	ErrMMCodeInvalidSize                  = errors.New("mgCSI: each MobilityTriggers entry (MM-Code) must be exactly 1 octet per TS 29.002 MAP-MS-DataTypes.asn:2544")

	ErrLocationEstimateTypeInvalid       = errors.New("locationType: LocationEstimateType must be 0..5 per TS 29.002 MAP-LCS-DataTypes.asn:153 (extensible enum: unknown values preserved on decode)")
	ErrLCSClientTypeInvalid              = errors.New("lcsClientID: LcsClientType must be 0..3 per TS 29.002 MAP-LCS-DataTypes.asn:188 (extensible enum: unknown values preserved on decode)")
	ErrLCSFormatIndicatorInvalid         = errors.New("lcsClientName/lcsRequestorID: LCSFormatIndicator must be 0..4 per TS 29.002 MAP-LCS-DataTypes.asn:224 (extensible enum: unknown values preserved on decode)")
	ErrPrivacyCheckRelatedActionInvalid  = errors.New("lcsPrivacyCheck: PrivacyCheckRelatedAction must be 0..4 per TS 29.002 MAP-LCS-DataTypes.asn:307")
	ErrAccuracyFulfilmentIndicatorInvalid = errors.New("psl: AccuracyFulfilmentIndicator must be 0..1 per TS 29.002 MAP-LCS-DataTypes.asn:457 (extensible enum: unknown values preserved on decode)")
	ErrResponseTimeCategoryInvalid       = errors.New("responseTime: ResponseTimeCategory encoder requires lowdelay(0) or delaytolerant(1); decoder applies spec exception clause TS 29.002 MAP-LCS-DataTypes.asn:270-271 (unrecognized values → delaytolerant)")
	ErrLCSPriorityInvalidSize            = errors.New("psl: LCSPriority must be exactly 1 octet per TS 29.002 MAP-LCS-DataTypes.asn:232")
	ErrLCSReferenceNumberInvalidSize     = errors.New("psl: LCSReferenceNumber must be exactly 1 octet per TS 29.002 MAP-CommonDataTypes.asn (LCS-ReferenceNumber)")
	ErrHorizontalAccuracyInvalidSize     = errors.New("lcsQoS: HorizontalAccuracy must be exactly 1 octet per TS 29.002 MAP-LCS-DataTypes.asn:249 (7-bit Uncertainty Code per TS 23.032)")
	ErrHorizontalAccuracyReservedBit     = errors.New("lcsQoS: HorizontalAccuracy bit 8 must be 0 per TS 29.002 MAP-LCS-DataTypes.asn:250 (only the low 7 bits encode the uncertainty code per TS 23.032)")
	ErrVerticalAccuracyInvalidSize       = errors.New("lcsQoS: VerticalAccuracy must be exactly 1 octet per TS 29.002 MAP-LCS-DataTypes.asn:255 (7-bit Vertical Uncertainty Code per TS 23.032)")
	ErrVerticalAccuracyReservedBit       = errors.New("lcsQoS: VerticalAccuracy bit 8 must be 0 per TS 29.002 MAP-LCS-DataTypes.asn:256 (only the low 7 bits encode the vertical uncertainty code per TS 23.032)")
	ErrUSSDDataCodingSchemeInvalidSize   = errors.New("ussd: USSD-DataCodingScheme must be exactly 1 octet on the wire per TS 29.002 MAP-SS-DataTypes.asn (USSD-DataCodingScheme ::= OCTET STRING (SIZE (1)))")
	ErrLCSCodewordStringSize             = errors.New("lcsCodeword: LcsCodewordString must be 1..20 octets (maxLCSCodewordStringLength) per TS 29.002 MAP-LCS-DataTypes.asn:298")
	ErrLCSClientNameNameStringSize       = errors.New("lcsClientName: NameString must be 1..63 octets (maxNameStringLength) per TS 29.002 MAP-LCS-DataTypes.asn:210")
	ErrLCSRequestorIDStringSize          = errors.New("lcsRequestorID: RequestorIDString must be 1..63 octets (maxRequestorIDStringLength) per TS 29.002 MAP-LCS-DataTypes.asn:220")
	ErrDeferredLocationEventTypeSize     = errors.New("locationType: DeferredLocationEventType BIT STRING must be 1..16 bits per TS 29.002 MAP-LCS-DataTypes.asn:165 (5 named bits, padded to multiple of 8 on the wire)")
	ErrSupportedGADShapesSize            = errors.New("psl: SupportedGADShapes BIT STRING must be 7..16 bits per TS 29.002 MAP-LCS-DataTypes.asn:280 (7 named bits, padded to multiple of 8 on the wire)")
	ErrLCSClientIDDialedByMSEmpty        = errors.New("lcsClientID: LcsClientDialedByMSNature/Plan must not be set when LcsClientDialedByMS digits are empty (presence cannot round-trip through string-based API)")

	ErrExtGeographicalInformationSize     = errors.New("psl: ExtGeographicalInformation must be 1..20 octets (maxExt-GeographicalInformation) per TS 29.002 MAP-LCS-DataTypes.asn:462")
	ErrAddGeographicalInformationSize     = errors.New("psl: AddGeographicalInformation must be 1..91 octets (maxAdd-GeographicalInformation) per TS 29.002 MAP-LCS-DataTypes.asn:601")
	ErrVelocityEstimateSize               = errors.New("psl: VelocityEstimate must be 4..7 octets per TS 29.002 MAP-LCS-DataTypes.asn:522")
	ErrPositioningDataInformationSize     = errors.New("psl: PositioningDataInformation must be 2..10 octets (maxPositioningDataInformation) per TS 29.002 MAP-LCS-DataTypes.asn:552")
	ErrUtranPositioningDataInfoSize       = errors.New("psl: UtranPositioningDataInfo must be 3..11 octets (maxUtranPositioningDataInfo) per TS 29.002 MAP-LCS-DataTypes.asn:560")
	ErrGeranGANSSpositioningDataSize      = errors.New("psl: GeranGANSSpositioningData must be 2..10 octets (maxGeranGANSSpositioningData) per TS 29.002 MAP-LCS-DataTypes.asn:568")
	ErrUtranGANSSpositioningDataSize      = errors.New("psl: UtranGANSSpositioningData must be 1..9 octets (maxUtranGANSSpositioningData) per TS 29.002 MAP-LCS-DataTypes.asn:576")
	ErrUtranAdditionalPositioningDataSize = errors.New("psl: UtranAdditionalPositioningData must be 1..8 octets (maxUtranAdditionalPositioningData) per TS 29.002 MAP-LCS-DataTypes.asn:584")
	ErrUtranBaroPressureMeasOutOfRange    = errors.New("psl: UtranBaroPressureMeas must be 30000..115000 per TS 29.002 MAP-LCS-DataTypes.asn:592")

	ErrAreaTypeInvalid                   = errors.New("area: AreaType must be 0..5 per TS 29.002 MAP-LCS-DataTypes.asn:337 (extensible enum: unknown values preserved on decode)")
	ErrAreaIdentificationSize            = errors.New("area: AreaIdentification must be 2..7 octets per TS 29.002 MAP-LCS-DataTypes.asn:346")
	ErrAreaListSize                      = errors.New("areaDefinition: AreaList must contain 1..10 entries (maxNumOfAreas) per TS 29.002 MAP-LCS-DataTypes.asn:328-330")
	ErrOccurrenceInfoInvalid             = errors.New("areaEventInfo: OccurrenceInfo must be 0..1 per TS 29.002 MAP-LCS-DataTypes.asn:361 (extensible enum: unknown values preserved on decode)")
	ErrIntervalTimeOutOfRange            = errors.New("areaEventInfo: IntervalTime must be 1..32767 seconds per TS 29.002 MAP-LCS-DataTypes.asn:366")
	ErrReportingAmountOutOfRange         = errors.New("periodicLDRInfo: ReportingAmount must be 1..8639999 (maxReportingAmount) per TS 29.002 MAP-LCS-DataTypes.asn:380-382")
	ErrReportingIntervalOutOfRange       = errors.New("periodicLDRInfo: ReportingInterval must be 1..8639999 seconds (maxReportingInterval) per TS 29.002 MAP-LCS-DataTypes.asn:384-387")
	ErrPeriodicLDRProductExceeded        = errors.New("periodicLDRInfo: ReportingInterval × ReportingAmount must not exceed 8639999 (99d 23h 59m 59s) per TS 29.002 MAP-LCS-DataTypes.asn:375-376")
	ErrRANTechnologyInvalid              = errors.New("reportingPLMN: RanTechnology must be 0..1 per TS 29.002 MAP-LCS-DataTypes.asn:420 (extensible enum: unknown values preserved on decode)")
	ErrPLMNListSize                      = errors.New("reportingPLMNList: PlmnList must contain 1..20 entries (maxNumOfReportingPLMN) per TS 29.002 MAP-LCS-DataTypes.asn:409-412")
	ErrTerminationCauseInvalid           = errors.New("deferredmt-lrData: TerminationCause must be 0..9 per TS 29.002 MAP-LCS-DataTypes.asn:696 (extensible enum: unknown values preserved on decode)")
	ErrServingNodeAddressMultipleAlts    = errors.New("servingNodeAddress: CHOICE has multiple alternatives set; pick exactly one of MscNumber, SgsnNumber, or MmeNumber")
	ErrServingNodeAddressNoAlt           = errors.New("servingNodeAddress: CHOICE has no alternative set; pick exactly one of MscNumber, SgsnNumber, or MmeNumber")
	ErrServingNodeAddressMmeNumberSize   = errors.New("servingNodeAddress: MmeNumber must be 9..255 octets (DiameterIdentity per RFC 6733) per TS 29.002 MAP-MS-DataTypes.asn:1434")
	ErrServingNodeAddressMscNumberDecodedEmpty  = errors.New("servingNodeAddress: present wire MscNumber decoded to empty digits; presence cannot round-trip through string-based API")
	ErrServingNodeAddressSgsnNumberDecodedEmpty = errors.New("servingNodeAddress: present wire SgsnNumber decoded to empty digits; presence cannot round-trip through string-based API")

	ErrPSLArgNil                         = errors.New("provideSubscriberLocationArg: argument must not be nil")
	ErrPSLArgMlcNumberEmpty              = errors.New("provideSubscriberLocationArg: MlcNumber digits are mandatory; empty value is not permitted on encode")
	ErrPSLArgMlcNumberDecodedEmpty       = errors.New("provideSubscriberLocationArg: present wire ISDN-AddressString decoded to empty digits; presence cannot round-trip through string-based API")
	ErrPSLArgMSISDNDecodedEmpty          = errors.New("provideSubscriberLocationArg: present wire MSISDN decoded to empty digits; presence cannot round-trip through string-based API")
	ErrPSLArgIMSIDecodedEmpty            = errors.New("provideSubscriberLocationArg: present wire IMSI decoded to empty digits; presence cannot round-trip through string-based API")
	ErrPSLArgIMSIInvalidSize             = errors.New("provideSubscriberLocationArg: IMSI must be 5..15 BCD digits per TS 29.002 MAP-CommonDataTypes.asn (TBCD-STRING SIZE 3..8 octets per ITU E.212)")
	ErrPSLArgIMEIDecodedEmpty            = errors.New("provideSubscriberLocationArg: present wire IMEI decoded to empty digits; presence cannot round-trip through string-based API")
	ErrPSLArgIMEIInvalidSize             = errors.New("provideSubscriberLocationArg: IMEI must be exactly 15 BCD digits per 3GPP TS 23.003 (TBCD-STRING SIZE 8 octets)")
	ErrPSLArgLMSIInvalidSize             = errors.New("provideSubscriberLocationArg: LMSI must be exactly 4 octets per TS 29.002 MAP-CommonDataTypes.asn")
	ErrPSLArgLcsServiceTypeIDOutOfRange  = errors.New("provideSubscriberLocationArg: LcsServiceTypeID must be 0..127 per TS 29.002 MAP-CommonDataTypes.asn:436 (LCSServiceTypeID INTEGER (0..127))")

	ErrPSLResNil                         = errors.New("provideSubscriberLocationRes: argument must not be nil")
	ErrPSLResLocationEstimateMissing     = errors.New("provideSubscriberLocationRes: LocationEstimate is mandatory; nil/empty value is not permitted on encode")
	ErrPSLResCellGlobalIdSize            = errors.New("provideSubscriberLocationRes: CellGlobalId must be exactly 7 octets per TS 29.002 MAP-CommonDataTypes.asn (CellGlobalIdOrServiceAreaIdFixedLength)")
	ErrPSLResLAIInvalidSize              = errors.New("provideSubscriberLocationRes: LAI must be exactly 5 octets per TS 29.002 MAP-CommonDataTypes.asn (LAIFixedLength)")
	ErrPSLResCellGlobalIdAndLAIMutex     = errors.New("provideSubscriberLocationRes: CellGlobalId and LAI are mutually exclusive (CellIdOrSai CHOICE); set exactly one")
	ErrPSLResCellIdOrSaiInvalidChoice    = errors.New("provideSubscriberLocationRes: CellIdOrSai CHOICE has unknown or empty selected alternative on the wire; cannot decode")
)
