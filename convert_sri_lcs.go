// convert_sri_lcs.go
//
// Top-level converters for SendRoutingInfoForLCS (opCode 85):
// RoutingInfoForLCS-Arg and -Res. The GMLC → HLR location-routing
// query that precedes ProvideSubscriberLocation (83). Reuses the shared
// SubscriberIdentity CHOICE converter, the LCSLocationInfo converter
// (from the SubscriberLocationReport work), the ISDN-address helpers,
// and the gsn IP codec. Marshal()/Parse() entry points live in
// marshal.go / parse.go.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

	"github.com/gomaja/go-asn1-gsmmap/gsn"
)

// ============================================================================
// RoutingInfoForLCS-Arg
// ============================================================================

func convertSriLcsToArg(a *SriLcs) (*gsm_map.RoutingInfoForLCSArg, error) {
	if a == nil {
		return nil, ErrSriLcsNil
	}
	if a.MlcNumber == "" {
		return nil, ErrSriLcsMlcNumberEmpty
	}
	mlc, err := encodeAddressField(a.MlcNumber, a.MlcNumberNature, a.MlcNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding SriLcs.MlcNumber: %w", err)
	}
	target, err := convertSubscriberIdentityToWire(a.TargetMS)
	if err != nil {
		return nil, fmt.Errorf("SriLcs.TargetMS: %w", err)
	}
	return &gsm_map.RoutingInfoForLCSArg{
		MlcNumber: gsm_map.ISDNAddressString(mlc),
		TargetMS:  target,
	}, nil
}

func convertArgToSriLcs(w *gsm_map.RoutingInfoForLCSArg) (*SriLcs, error) {
	if w == nil {
		return nil, ErrSriLcsNil
	}
	mlc, nature, plan, err := decodeAddressField([]byte(w.MlcNumber))
	if err != nil {
		return nil, fmt.Errorf("decoding SriLcs.MlcNumber: %w", err)
	}
	if mlc == "" {
		return nil, ErrSriLcsMlcNumberDecodedEmpty
	}
	target, err := convertWireToSubscriberIdentity(w.TargetMS)
	if err != nil {
		return nil, fmt.Errorf("SriLcs.TargetMS: %w", err)
	}
	return &SriLcs{
		MlcNumber:       mlc,
		MlcNumberNature: nature,
		MlcNumberPlan:   plan,
		TargetMS:        target,
	}, nil
}

// ============================================================================
// RoutingInfoForLCS-Res
// ============================================================================

func convertSriLcsRespToRes(r *SriLcsResp) (*gsm_map.RoutingInfoForLCSRes, error) {
	if r == nil {
		return nil, ErrSriLcsRespNil
	}
	target, err := convertSubscriberIdentityToWire(r.TargetMS)
	if err != nil {
		return nil, fmt.Errorf("SriLcsResp.TargetMS: %w", err)
	}
	locInfo, err := convertLCSLocationInfoToWire(&r.LcsLocationInfo)
	if err != nil {
		return nil, fmt.Errorf("SriLcsResp.LcsLocationInfo: %w", err)
	}

	out := &gsm_map.RoutingInfoForLCSRes{
		TargetMS:        target,
		LcsLocationInfo: *locInfo,
	}

	if err := encodeOptionalGSN(r.VGmlcAddress, "SriLcsResp.VGmlcAddress", &out.VGmlcAddress); err != nil {
		return nil, err
	}
	if err := encodeOptionalGSN(r.HGmlcAddress, "SriLcsResp.HGmlcAddress", &out.HGmlcAddress); err != nil {
		return nil, err
	}
	if err := encodeOptionalGSN(r.PprAddress, "SriLcsResp.PprAddress", &out.PprAddress); err != nil {
		return nil, err
	}
	if err := encodeOptionalGSN(r.AdditionalVGmlcAddress, "SriLcsResp.AdditionalVGmlcAddress", &out.AdditionalVGmlcAddress); err != nil {
		return nil, err
	}
	return out, nil
}

func convertResToSriLcsResp(w *gsm_map.RoutingInfoForLCSRes) (*SriLcsResp, error) {
	if w == nil {
		return nil, ErrSriLcsRespNil
	}
	target, err := convertWireToSubscriberIdentity(w.TargetMS)
	if err != nil {
		return nil, fmt.Errorf("SriLcsResp.TargetMS: %w", err)
	}
	locInfo, err := convertWireToLCSLocationInfo(&w.LcsLocationInfo)
	if err != nil {
		return nil, fmt.Errorf("SriLcsResp.LcsLocationInfo: %w", err)
	}

	out := &SriLcsResp{
		TargetMS:        target,
		LcsLocationInfo: *locInfo,
	}

	if out.VGmlcAddress, err = decodeOptionalGSN(w.VGmlcAddress, "SriLcsResp.VGmlcAddress"); err != nil {
		return nil, err
	}
	if out.HGmlcAddress, err = decodeOptionalGSN(w.HGmlcAddress, "SriLcsResp.HGmlcAddress"); err != nil {
		return nil, err
	}
	if out.PprAddress, err = decodeOptionalGSN(w.PprAddress, "SriLcsResp.PprAddress"); err != nil {
		return nil, err
	}
	if out.AdditionalVGmlcAddress, err = decodeOptionalGSN(w.AdditionalVGmlcAddress, "SriLcsResp.AdditionalVGmlcAddress"); err != nil {
		return nil, err
	}
	return out, nil
}

// encodeOptionalGSN builds an optional GSN-Address field from an IP
// string ("" = absent), assigning the wire pointer in place.
func encodeOptionalGSN(ipStr, field string, dst **gsm_map.GSNAddress) error {
	if ipStr == "" {
		return nil
	}
	b, err := gsn.Build(ipStr)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", field, err)
	}
	v := gsm_map.GSNAddress(b)
	*dst = &v
	return nil
}

// decodeOptionalGSN parses an optional GSN-Address field to an IP string
// (nil pointer = "" absent).
func decodeOptionalGSN(src *gsm_map.GSNAddress, field string) (string, error) {
	if src == nil {
		return "", nil
	}
	s, err := gsn.Parse(*src)
	if err != nil {
		return "", fmt.Errorf("decoding %s: %w", field, err)
	}
	return s, nil
}
