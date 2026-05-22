// convert_slr_res.go
//
// Top-level converter for SubscriberLocationReportRes (opCode 86) and
// its Marshal()/ParseSubscriberLocationReportRes() entry points. PR G4
// of the staged SLR implementation: completes the SubscriberLocationReport
// operation after the Arg side (#53/#54/#55). Reuses the ISDN-address,
// GSN-address, ReportingPLMNList, and NULL-flag helpers built for
// ProvideSubscriberLocation.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

	"github.com/gomaja/go-asn1-gsmmap/gsn"
)

// convertSubscriberLocationReportResToWire builds the wire-form
// gsm_map.SubscriberLocationReportRes from the public type. Every field
// is optional, so an all-absent input yields an empty (but valid) wire
// response. NaESRK and NaESRD are independent optionals, not a CHOICE.
func convertSubscriberLocationReportResToWire(r *SubscriberLocationReportRes) (*gsm_map.SubscriberLocationReportRes, error) {
	if r == nil {
		return nil, ErrSLRResNil
	}

	out := &gsm_map.SubscriberLocationReportRes{}

	// [0] na-ESRK
	if r.NaESRK != "" {
		isdn, err := encodeAddressField(r.NaESRK, r.NaESRKNature, r.NaESRKPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SubscriberLocationReportRes.NaESRK: %w", err)
		}
		v := gsm_map.ISDNAddressString(isdn)
		out.NaESRK = &v
	}
	// [1] na-ESRD
	if r.NaESRD != "" {
		isdn, err := encodeAddressField(r.NaESRD, r.NaESRDNature, r.NaESRDPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SubscriberLocationReportRes.NaESRD: %w", err)
		}
		v := gsm_map.ISDNAddressString(isdn)
		out.NaESRD = &v
	}
	// [2] h-gmlc-Address
	if r.HGmlcAddress != "" {
		gsnAddr, err := gsn.Build(r.HGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("encoding SubscriberLocationReportRes.HGmlcAddress: %w", err)
		}
		v := gsm_map.GSNAddress(gsnAddr)
		out.HGmlcAddress = &v
	}
	// [3] mo-lrShortCircuitIndicator (NULL flag)
	out.MoLrShortCircuitIndicator = boolToNullPtr(r.MoLrShortCircuitIndicator)
	// [4] reporting-PLMN-List
	if r.ReportingPLMNList != nil {
		v, err := convertReportingPLMNListToWire(r.ReportingPLMNList)
		if err != nil {
			return nil, fmt.Errorf("SubscriberLocationReportRes.ReportingPLMNList: %w", err)
		}
		out.ReportingPLMNList = v
	}
	// [5] lcs-ReferenceNumber (OCTET STRING SIZE 1)
	if len(r.LcsReferenceNumber) > 0 {
		if len(r.LcsReferenceNumber) != 1 {
			return nil, fmt.Errorf("SubscriberLocationReportRes.LcsReferenceNumber len=%d: %w", len(r.LcsReferenceNumber), ErrLCSReferenceNumberInvalidSize)
		}
		v := gsm_map.LCSReferenceNumber(r.LcsReferenceNumber)
		out.LcsReferenceNumber = &v
	}

	return out, nil
}

// convertWireToSubscriberLocationReportRes unmarshals the wire-form
// struct back to the public type. Validation rules:
//   - Round-trip safety: present-but-empty NaESRK/NaESRD decoded values
//     are rejected (cannot round-trip through the string-based API).
//   - LcsReferenceNumber byte size: rejected when != 1, symmetric with
//     the encoder.
//   - ExtensionContainer: dropped (opaque metadata not surfaced; see
//     SubscriberLocationReportRes doc).
func convertWireToSubscriberLocationReportRes(w *gsm_map.SubscriberLocationReportRes) (*SubscriberLocationReportRes, error) {
	if w == nil {
		return nil, ErrSLRResNil
	}

	out := &SubscriberLocationReportRes{}

	if w.NaESRK != nil {
		s, nature, plan, err := decodeAddressField([]byte(*w.NaESRK))
		if err != nil {
			return nil, fmt.Errorf("decoding SubscriberLocationReportRes.NaESRK: %w", err)
		}
		if s == "" {
			return nil, ErrSLRResNaESRKDecodedEmpty
		}
		out.NaESRK = s
		out.NaESRKNature = nature
		out.NaESRKPlan = plan
	}
	if w.NaESRD != nil {
		s, nature, plan, err := decodeAddressField([]byte(*w.NaESRD))
		if err != nil {
			return nil, fmt.Errorf("decoding SubscriberLocationReportRes.NaESRD: %w", err)
		}
		if s == "" {
			return nil, ErrSLRResNaESRDDecodedEmpty
		}
		out.NaESRD = s
		out.NaESRDNature = nature
		out.NaESRDPlan = plan
	}
	if w.HGmlcAddress != nil {
		addr, err := gsn.Parse(*w.HGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("decoding SubscriberLocationReportRes.HGmlcAddress: %w", err)
		}
		out.HGmlcAddress = addr
	}
	out.MoLrShortCircuitIndicator = nullPtrToBool(w.MoLrShortCircuitIndicator)
	if w.ReportingPLMNList != nil {
		v, err := convertWireToReportingPLMNList(w.ReportingPLMNList)
		if err != nil {
			return nil, fmt.Errorf("SubscriberLocationReportRes.ReportingPLMNList: %w", err)
		}
		out.ReportingPLMNList = v
	}
	if w.LcsReferenceNumber != nil {
		if len(*w.LcsReferenceNumber) != 1 {
			return nil, fmt.Errorf("SubscriberLocationReportRes.LcsReferenceNumber len=%d: %w", len(*w.LcsReferenceNumber), ErrLCSReferenceNumberInvalidSize)
		}
		out.LcsReferenceNumber = LCSReferenceNumber(*w.LcsReferenceNumber)
	}

	return out, nil
}
