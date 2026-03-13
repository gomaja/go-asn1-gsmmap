package gsmmap

import (
	"encoding/hex"
	"encoding/json"

	"github.com/warthog618/sms/encoding/tpdu"
)

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

// SriSm represents a Send Routing Info for Short Message request.
type SriSm struct {
	MSISDN               string
	MSISDNNature         uint8 // address nature indicator (default: International)
	MSISDNPlan           uint8 // numbering plan indicator (default: ISDN)
	SmRpPri              bool
	ServiceCentreAddress string
	SCANature            uint8 // address nature indicator (default: International)
	SCAPlan              uint8 // numbering plan indicator (default: ISDN)
}

// SriSmResp represents a Send Routing Info for Short Message response.
type SriSmResp struct {
	IMSI                 string
	LocationInfoWithLMSI LocationInfoWithLMSI
}

// LocationInfoWithLMSI contains location information with LMSI.
type LocationInfoWithLMSI struct {
	NetworkNodeNumber       string
	NetworkNodeNumberNature uint8 // address nature indicator
	NetworkNodeNumberPlan   uint8 // numbering plan indicator
}

// MtFsm represents a Mobile Terminated Forward Short Message.
type MtFsm struct {
	IMSI                   string
	ServiceCentreAddressOA string
	SCAOANature            uint8 // address nature indicator (default: International)
	SCAOAPlan              uint8 // numbering plan indicator (default: ISDN)
	TPDU                   tpdu.TPDU
	MoreMessagesToSend     bool
}

// MoFsm represents a Mobile Originated Forward Short Message.
type MoFsm struct {
	ServiceCentreAddressDA string
	SCADANature            uint8 // address nature indicator (default: International)
	SCADAPlan              uint8 // numbering plan indicator (default: ISDN)
	MSISDN                 string
	MSISDNNature           uint8 // address nature indicator (default: International)
	MSISDNPlan             uint8 // numbering plan indicator (default: ISDN)
	TPDU                   tpdu.TPDU
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
	GeographicalInformation  HexBytes // raw 8 octets; nil if absent
	GeodeticInformation      HexBytes // raw 10 octets; nil if absent
	CellGlobalId             HexBytes // raw fixed-length cell ID or SAI; nil if absent
	LAI                      HexBytes // raw 5-octet LAI; nil if absent
	LocationNumber           HexBytes // raw octets; nil if absent
	CurrentLocationRetrieved bool
	SAIPresent               bool
}

// EPSLocationInformation contains EPS/LTE location data.
type EPSLocationInformation struct {
	AgeOfLocationInformation *int   // seconds; nil if absent
	EUtranCellGlobalIdentity HexBytes // raw 7 octets; nil if absent
	TrackingAreaIdentity     HexBytes // raw 5 octets; nil if absent
	GeographicalInformation  HexBytes // raw 8 octets; nil if absent
	GeodeticInformation      HexBytes // raw 10 octets; nil if absent
	CurrentLocationRetrieved bool
	MmeName                  HexBytes // raw DiameterIdentity; nil if absent
}

// GPRSLocationInformation contains GPRS domain location data.
type GPRSLocationInformation struct {
	AgeOfLocationInformation *int   // seconds; nil if absent
	CellGlobalId             HexBytes // raw fixed-length cell ID or SAI; nil if absent
	LAI                      HexBytes // raw 5-octet LAI; nil if absent
	RouteingAreaIdentity     HexBytes // raw octets; nil if absent
	GeographicalInformation  HexBytes // raw 8 octets; nil if absent
	GeodeticInformation      HexBytes // raw 10 octets; nil if absent
	SgsnNumber               string // decoded; empty if absent
	SgsnNumberNature         uint8
	SgsnNumberPlan           uint8
	CurrentLocationRetrieved bool
	SAIPresent               bool
}
