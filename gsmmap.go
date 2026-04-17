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
// (opCode 56). Exactly one of Triplets or Quintuplets must be non-nil when
// the list is set; both nil is treated as "no alternative" and both non-nil
// as "multiple alternatives" during encode.
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

	ErrSaiMissingIMSI                                = errors.New("sai: IMSI is empty")
	ErrSaiInvalidNumberOfRequestedVectors            = errors.New("sai: NumberOfRequestedVectors must be 1..5")
	ErrSaiInvalidNumberOfRequestedAdditionalVectors  = errors.New("sai: NumberOfRequestedAdditionalVectors must be 1..5")
	ErrSaiInvalidUeUsageType                         = errors.New("sai: UeUsageType must be exactly 4 octets")
	ErrSaiInvalidPLMNId                              = errors.New("sai: RequestingPLMNId must be exactly 3 octets")
	ErrSaiAuthSetListChoiceMultipleAlternatives      = errors.New("sai: AuthenticationSetList CHOICE has multiple alternatives set")
	ErrSaiAuthSetListChoiceNoAlternative             = errors.New("sai: AuthenticationSetList CHOICE has no alternative set")
	ErrSaiInvalidRequestingNodeType                  = errors.New("sai: RequestingNodeType must be one of vlr(0), sgsn(1), s-cscf(2), bsf(3), gan-aaa-server(4), wlan-aaa-server(5), mme(16), mme-sgsn(17)")
	ErrSaiInvalidEpsAuthSetListSize                  = errors.New("sai: EpsAuthenticationSetList must have 1..5 entries")
)
