package gsmmap

import gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

// OID is an OBJECT IDENTIFIER in canonical dotted-decimal form, for example
// "0.4.0.0.1.0.19.3". It converts to gomaja/go-asn1-tcap's tcap.OID with a
// type conversion. Marshal refuses malformed or non-canonical text with
// ErrInvalidOID. Parse returns ErrOIDTextLimit from go-asn1's gsm_map/runtime
// package for a valid wire OID whose decimal form exceeds its conversion limit.
type OID string

// MapDialogueAS returns map-DialogueAS (TS 29.002 V19.1.0 §17.4), the direct
// reference of the EXTERNAL that carries a DialoguePDU (§17.1).
func MapDialogueAS() OID { return OID(gsm_map.MapDialogueAS().String()) }

// DialoguePDU is one MAP-DialoguePDU (TS 29.002 V19.1.0 §17.4).
// Exactly one field is set. It goes in an EXTERNAL's single-ASN1-type in
// TCAP dialogue user information, with direct reference MapDialogueAS()
// (§§17.1, 16.1.2.3). Parse accepts unknown extension additions, which
// §17.1.4 forbids rejecting; like the rest of this package, it drops them and
// extension containers, and Marshal never sends either.
type DialoguePDU struct {
	Open          *DialogueOpen          // map-open [0] MAP-OpenInfo
	Accept        *DialogueAccept        // map-accept [1] MAP-AcceptInfo
	Close         *DialogueClose         // map-close [2] MAP-CloseInfo
	Refuse        *DialogueRefuse        // map-refuse [3] MAP-RefuseInfo
	UserAbort     *DialogueUserAbort     // map-userAbort [4] MAP-UserAbortInfo
	ProviderAbort *DialogueProviderAbort // map-providerAbort [5] MAP-ProviderAbortInfo
}

// DialogueOpen holds MAP-OpenInfo (§17.4). The caller applies the service
// restrictions and reference contents in §7.3.1 Tables 7.3/2 and 7.3/3:
// destination references hold IMSI or MSISDN; originating references hold
// ISDN addresses. Empty digits omit a reference, ignoring its nature/plan.
type DialogueOpen struct {
	DestinationReference       string // [0] AddressString OPTIONAL, TBCD digits; "" omits it.
	DestinationReferenceNature uint8
	DestinationReferencePlan   uint8
	OriginationReference       string // [1] AddressString OPTIONAL, TBCD digits; "" omits it.
	OriginationReferenceNature uint8
	OriginationReferencePlan   uint8
}

// DialogueAccept is MAP-AcceptInfo (§17.4), whose only field is an extension container.
type DialogueAccept struct{}

// DialogueClose is MAP-CloseInfo (§17.4), whose only field is an extension container.
type DialogueClose struct{}

// DialogueRefuse holds MAP-RefuseInfo (§17.4). AlternativeApplicationContext
// must not be used in version 2; the application context version is the caller's.
type DialogueRefuse struct {
	Reason                        DialogueRefuseReason
	AlternativeApplicationContext OID // OPTIONAL; "" omits it.
}

// DialogueUserAbort is MAP-UserAbortChoice (§17.4); exactly one field is set.
// User reasons and diagnostics are defined in §7.3.4.
type DialogueUserAbort struct {
	UserSpecificReason               bool                         // [0] NULL
	UserResourceLimitation           bool                         // [1] NULL
	ResourceUnavailable              *ResourceUnavailableReason   // [2]
	ApplicationProcedureCancellation *ProcedureCancellationReason // [3]
}

// DialogueProviderAbort holds MAP-ProviderAbortInfo (§17.4; service in §7.3.5).
type DialogueProviderAbort struct {
	Reason ProviderAbortReason
}

// DialogueRefuseReason is the non-extensible Reason enumeration (§17.4).
type DialogueRefuseReason = gsm_map.Reason

const (
	RefuseReasonNoReasonGiven               DialogueRefuseReason = gsm_map.ReasonNoReasonGiven
	RefuseReasonInvalidDestinationReference DialogueRefuseReason = gsm_map.ReasonInvalidDestinationReference
	RefuseReasonInvalidOriginatingReference DialogueRefuseReason = gsm_map.ReasonInvalidOriginatingReference
)

// ResourceUnavailableReason is the non-extensible diagnostic enumeration (§17.4).
type ResourceUnavailableReason = gsm_map.ResourceUnavailableReason

const (
	ResourceUnavailableShortTermResourceLimitation ResourceUnavailableReason = gsm_map.ResourceUnavailableReasonShortTermResourceLimitation
	ResourceUnavailableLongTermResourceLimitation  ResourceUnavailableReason = gsm_map.ResourceUnavailableReasonLongTermResourceLimitation
)

// ProcedureCancellationReason is the non-extensible diagnostic enumeration (§17.4).
type ProcedureCancellationReason = gsm_map.ProcedureCancellationReason

const (
	ProcedureCancellationHandoverCancellation       ProcedureCancellationReason = gsm_map.ProcedureCancellationReasonHandoverCancellation
	ProcedureCancellationRadioChannelRelease        ProcedureCancellationReason = gsm_map.ProcedureCancellationReasonRadioChannelRelease
	ProcedureCancellationNetworkPathRelease         ProcedureCancellationReason = gsm_map.ProcedureCancellationReasonNetworkPathRelease
	ProcedureCancellationCallRelease                ProcedureCancellationReason = gsm_map.ProcedureCancellationReasonCallRelease
	ProcedureCancellationAssociatedProcedureFailure ProcedureCancellationReason = gsm_map.ProcedureCancellationReasonAssociatedProcedureFailure
	ProcedureCancellationTandemDialogueRelease      ProcedureCancellationReason = gsm_map.ProcedureCancellationReasonTandemDialogueRelease
	ProcedureCancellationRemoteOperationsFailure    ProcedureCancellationReason = gsm_map.ProcedureCancellationReasonRemoteOperationsFailure
)

// ProviderAbortReason is the non-extensible MAP-ProviderAbortReason enumeration (§17.4).
type ProviderAbortReason = gsm_map.MAPProviderAbortReason

const (
	ProviderAbortAbnormalDialogue ProviderAbortReason = gsm_map.MAPProviderAbortReasonAbnormalDialogue
	ProviderAbortInvalidPDU       ProviderAbortReason = gsm_map.MAPProviderAbortReasonInvalidPDU
)
