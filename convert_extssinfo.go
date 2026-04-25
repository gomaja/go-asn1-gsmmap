// Ext-SS-Info CHOICE converters (ISD PR D).
//
// Covers TS 29.002 MAP-MS-DataTypes.asn:1826-onwards: the 5-alternative
// CHOICE used inside Ext-SS-InfoList plus all directly-referenced
// nested SEQUENCEs (Ext-ForwInfo, Ext-CallBarInfo, CUG-Info,
// Ext-SS-Data, EMLPP-Info) and CHOICEs (SS-SubscriptionOption).
//
// Reuses ExtBasicServiceCode (PR #7) and the SS-Code typedef.
//
// CHOICE pattern: each public CHOICE struct has separate optional
// pointer fields per alternative. The encoder counts the populated
// alternatives and returns ErrXxxChoiceMultipleAlternatives /
// ErrXxxChoiceNoAlternative if the caller violated the exactly-one
// invariant. The decoder switches on the wire-side `.Choice` constant.

package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// --- Ext-SS-Status helpers (OCTET STRING SIZE 1..5) ---

func validateExtSSStatus(b HexBytes, field string) error {
	if len(b) < 1 || len(b) > 5 {
		return fmt.Errorf("%s: %w (got %d)", field, ErrExtSSStatusInvalidSize, len(b))
	}
	return nil
}

// --- SSSubscriptionOption (CHOICE) ---

func convertSSSubscriptionOptionToWire(o *SSSubscriptionOption) (*gsm_map.SSSubscriptionOption, error) {
	hasCli := o.CliRestriction != nil
	hasOver := o.Override != nil
	switch {
	case hasCli && hasOver:
		return nil, ErrSSSubscriptionOptionChoiceMultipleAlternatives
	case hasCli:
		if !isValidCliRestrictionOption(*o.CliRestriction) {
			return nil, ErrCliRestrictionOptionInvalidValue
		}
		v := gsm_map.NewSSSubscriptionOptionCliRestrictionOption(gsm_map.CliRestrictionOption(int64(*o.CliRestriction)))
		return &v, nil
	case hasOver:
		if !isValidOverrideCategory(*o.Override) {
			return nil, ErrOverrideCategoryInvalidValue
		}
		v := gsm_map.NewSSSubscriptionOptionOverrideCategory(gsm_map.OverrideCategory(int64(*o.Override)))
		return &v, nil
	default:
		return nil, ErrSSSubscriptionOptionChoiceNoAlternative
	}
}

func convertWireToSSSubscriptionOption(w *gsm_map.SSSubscriptionOption) (*SSSubscriptionOption, error) {
	switch w.Choice {
	case gsm_map.SSSubscriptionOptionChoiceCliRestrictionOption:
		if w.CliRestrictionOption == nil {
			return nil, ErrSSSubscriptionOptionChoiceNoAlternative
		}
		raw, err := narrowInt64(int64(*w.CliRestrictionOption))
		if err != nil {
			return nil, fmt.Errorf("CliRestrictionOption: %w", err)
		}
		v := CliRestrictionOption(raw)
		if !isValidCliRestrictionOption(v) {
			return nil, ErrCliRestrictionOptionInvalidValue
		}
		return &SSSubscriptionOption{CliRestriction: &v}, nil
	case gsm_map.SSSubscriptionOptionChoiceOverrideCategory:
		if w.OverrideCategory == nil {
			return nil, ErrSSSubscriptionOptionChoiceNoAlternative
		}
		raw, err := narrowInt64(int64(*w.OverrideCategory))
		if err != nil {
			return nil, fmt.Errorf("OverrideCategory: %w", err)
		}
		v := OverrideCategory(raw)
		if !isValidOverrideCategory(v) {
			return nil, ErrOverrideCategoryInvalidValue
		}
		return &SSSubscriptionOption{Override: &v}, nil
	default:
		return nil, ErrSSSubscriptionOptionChoiceNoAlternative
	}
}

func isValidCliRestrictionOption(v CliRestrictionOption) bool {
	switch v {
	case CliRestrictionPermanent,
		CliRestrictionTemporaryDefaultRestricted,
		CliRestrictionTemporaryDefaultAllowed:
		return true
	}
	return false
}

func isValidOverrideCategory(v OverrideCategory) bool {
	return v == OverrideEnabled || v == OverrideDisabled
}

// --- Ext-BasicServiceGroupList (SIZE 1..32) ---

func convertExtBasicServiceGroupListToWire(in []ExtBasicServiceCode) (gsm_map.ExtBasicServiceGroupList, error) {
	if in == nil {
		return nil, nil
	}
	if len(in) < 1 || len(in) > MaxNumOfExtBasicServiceGroups {
		return nil, ErrExtBasicServiceGroupListInvalidSize
	}
	out := make(gsm_map.ExtBasicServiceGroupList, len(in))
	for i := range in {
		w, err := convertExtBasicServiceCodeToWire(&in[i])
		if err != nil {
			return nil, fmt.Errorf("BasicServiceGroupList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToExtBasicServiceGroupList(w gsm_map.ExtBasicServiceGroupList) ([]ExtBasicServiceCode, error) {
	if w == nil {
		return nil, nil
	}
	if len(w) < 1 || len(w) > MaxNumOfExtBasicServiceGroups {
		return nil, ErrExtBasicServiceGroupListInvalidSize
	}
	out := make([]ExtBasicServiceCode, len(w))
	for i := range w {
		d, err := convertWireToExtBasicServiceCode(&w[i])
		if err != nil {
			return nil, fmt.Errorf("BasicServiceGroupList[%d]: %w", i, err)
		}
		out[i] = *d
	}
	return out, nil
}

// --- Ext-ForwFeature / Ext-ForwInfo ---

func convertExtForwFeatureToWire(f *ExtForwFeature) (gsm_map.ExtForwFeature, error) {
	if err := validateExtSSStatus(f.SsStatus, "Ext-ForwFeature.SsStatus"); err != nil {
		return gsm_map.ExtForwFeature{}, err
	}
	out := gsm_map.ExtForwFeature{SsStatus: gsm_map.ExtSSStatus(f.SsStatus)}
	if f.BasicService != nil {
		bs, err := convertExtBasicServiceCodeToWire(f.BasicService)
		if err != nil {
			return gsm_map.ExtForwFeature{}, fmt.Errorf("BasicService: %w", err)
		}
		out.BasicService = bs
	}
	if f.ForwardedToNumber != "" {
		enc, err := encodeAddressField(f.ForwardedToNumber, f.ForwardedToNature, f.ForwardedToPlan)
		if err != nil {
			return gsm_map.ExtForwFeature{}, fmt.Errorf("ForwardedToNumber: %w", err)
		}
		v := gsm_map.ISDNAddressString(enc)
		out.ForwardedToNumber = &v
	}
	if f.ForwardedToSubaddress != nil {
		// ISDN-SubaddressString SIZE(1..21) per TS 29.002. Reject a non-nil
		// empty slice rather than silently omitting it (PR #29 pattern).
		if len(f.ForwardedToSubaddress) < 1 || len(f.ForwardedToSubaddress) > 21 {
			return gsm_map.ExtForwFeature{}, ErrExtForwSubaddressInvalidSize
		}
		v := gsm_map.ISDNSubaddressString(f.ForwardedToSubaddress)
		out.ForwardedToSubaddress = &v
	}
	if f.ForwardingOptions != nil {
		// Ext-ForwOptions OCTET STRING (SIZE 1..5) per TS 29.002.
		if len(f.ForwardingOptions) < 1 || len(f.ForwardingOptions) > 5 {
			return gsm_map.ExtForwFeature{}, ErrExtForwOptionsInvalidSize
		}
		v := gsm_map.ExtForwOptions(f.ForwardingOptions)
		out.ForwardingOptions = &v
	}
	if f.NoReplyConditionTime != nil {
		v64 := int64(*f.NoReplyConditionTime)
		if v64 < 1 || v64 > 100 {
			return gsm_map.ExtForwFeature{}, ErrExtNoRepCondTimeOutOfRange
		}
		v := gsm_map.ExtNoRepCondTime(v64)
		out.NoReplyConditionTime = &v
	}
	if f.LongForwardedToNumber != "" {
		enc, err := encodeAddressField(f.LongForwardedToNumber, f.ForwardedToNature, f.ForwardedToPlan)
		if err != nil {
			return gsm_map.ExtForwFeature{}, fmt.Errorf("LongForwardedToNumber: %w", err)
		}
		v := gsm_map.FTNAddressString(enc)
		out.LongForwardedToNumber = &v
	}
	return out, nil
}

func convertWireToExtForwFeature(w *gsm_map.ExtForwFeature) (ExtForwFeature, error) {
	if err := validateExtSSStatus(HexBytes(w.SsStatus), "Ext-ForwFeature.SsStatus"); err != nil {
		return ExtForwFeature{}, err
	}
	out := ExtForwFeature{SsStatus: HexBytes(w.SsStatus)}
	if w.BasicService != nil {
		bs, err := convertWireToExtBasicServiceCode(w.BasicService)
		if err != nil {
			return ExtForwFeature{}, fmt.Errorf("BasicService: %w", err)
		}
		out.BasicService = bs
	}
	if w.ForwardedToNumber != nil {
		digits, nat, plan, err := decodeAddressField(*w.ForwardedToNumber)
		if err != nil {
			return ExtForwFeature{}, fmt.Errorf("ForwardedToNumber: %w", err)
		}
		out.ForwardedToNumber = digits
		out.ForwardedToNature = nat
		out.ForwardedToPlan = plan
	}
	if w.ForwardedToSubaddress != nil {
		// ISDN-SubaddressString SIZE(1..21) per TS 29.002.
		if len(*w.ForwardedToSubaddress) < 1 || len(*w.ForwardedToSubaddress) > 21 {
			return ExtForwFeature{}, ErrExtForwSubaddressInvalidSize
		}
		out.ForwardedToSubaddress = HexBytes(*w.ForwardedToSubaddress)
	}
	if w.ForwardingOptions != nil {
		if len(*w.ForwardingOptions) < 1 || len(*w.ForwardingOptions) > 5 {
			return ExtForwFeature{}, ErrExtForwOptionsInvalidSize
		}
		out.ForwardingOptions = HexBytes(*w.ForwardingOptions)
	}
	if w.NoReplyConditionTime != nil {
		// Per TS 29.002: 1..4 → 5; 31..100 → 30; outside 1..100 is
		// out of spec entirely. Apply the lenient mapping in int64
		// space so 32-bit narrowing can't bypass it.
		v64 := int64(*w.NoReplyConditionTime)
		if v64 < 1 || v64 > 100 {
			return ExtForwFeature{}, ErrExtNoRepCondTimeOutOfRange
		}
		switch {
		case v64 >= 1 && v64 <= 4:
			v64 = 5
		case v64 >= 31 && v64 <= 100:
			v64 = 30
		}
		v := int(v64)
		out.NoReplyConditionTime = &v
	}
	if w.LongForwardedToNumber != nil {
		// FTN-AddressString carries its own ext+ton+npi octet (TS 29.002).
		// The public type shares ForwardedToNature / ForwardedToPlan
		// between the short and long numbers, so the encoder reuses
		// whichever pair is populated. To preserve round-trip fidelity
		// when only LongForwardedToNumber is present, capture its
		// decoded nat/plan into the shared fields. When ForwardedToNumber
		// is also present, its values were already written above and
		// take precedence (consistent with the encoder's behavior).
		digits, nat, plan, err := decodeAddressField(*w.LongForwardedToNumber)
		if err != nil {
			return ExtForwFeature{}, fmt.Errorf("LongForwardedToNumber: %w", err)
		}
		out.LongForwardedToNumber = digits
		if w.ForwardedToNumber == nil {
			out.ForwardedToNature = nat
			out.ForwardedToPlan = plan
		}
	}
	return out, nil
}

func convertExtForwInfoToWire(f *ExtForwInfo) (*gsm_map.ExtForwInfo, error) {
	if len(f.ForwardingFeatureList) < 1 || len(f.ForwardingFeatureList) > MaxNumOfExtBasicServiceGroups {
		return nil, ErrExtForwFeatureListInvalidSize
	}
	list := make(gsm_map.ExtForwFeatureList, len(f.ForwardingFeatureList))
	for i := range f.ForwardingFeatureList {
		w, err := convertExtForwFeatureToWire(&f.ForwardingFeatureList[i])
		if err != nil {
			return nil, fmt.Errorf("ForwardingFeatureList[%d]: %w", i, err)
		}
		list[i] = w
	}
	return &gsm_map.ExtForwInfo{
		SsCode:                gsm_map.SSCode{byte(f.SsCode)},
		ForwardingFeatureList: list,
	}, nil
}

func convertWireToExtForwInfo(w *gsm_map.ExtForwInfo) (*ExtForwInfo, error) {
	if len(w.SsCode) != 1 {
		return nil, fmt.Errorf("Ext-ForwInfo.SsCode: must be 1 octet, got %d", len(w.SsCode))
	}
	if len(w.ForwardingFeatureList) < 1 || len(w.ForwardingFeatureList) > MaxNumOfExtBasicServiceGroups {
		return nil, ErrExtForwFeatureListInvalidSize
	}
	out := &ExtForwInfo{
		SsCode:                SsCode(w.SsCode[0]),
		ForwardingFeatureList: make([]ExtForwFeature, len(w.ForwardingFeatureList)),
	}
	for i := range w.ForwardingFeatureList {
		d, err := convertWireToExtForwFeature(&w.ForwardingFeatureList[i])
		if err != nil {
			return nil, fmt.Errorf("ForwardingFeatureList[%d]: %w", i, err)
		}
		out.ForwardingFeatureList[i] = d
	}
	return out, nil
}

// --- Ext-CallBarringFeature / Ext-CallBarInfo ---

func convertExtCallBarringFeatureToWire(f *ExtCallBarringFeature) (gsm_map.ExtCallBarringFeature, error) {
	if err := validateExtSSStatus(f.SsStatus, "Ext-CallBarringFeature.SsStatus"); err != nil {
		return gsm_map.ExtCallBarringFeature{}, err
	}
	out := gsm_map.ExtCallBarringFeature{SsStatus: gsm_map.ExtSSStatus(f.SsStatus)}
	if f.BasicService != nil {
		bs, err := convertExtBasicServiceCodeToWire(f.BasicService)
		if err != nil {
			return gsm_map.ExtCallBarringFeature{}, fmt.Errorf("BasicService: %w", err)
		}
		out.BasicService = bs
	}
	return out, nil
}

func convertWireToExtCallBarringFeature(w *gsm_map.ExtCallBarringFeature) (ExtCallBarringFeature, error) {
	if err := validateExtSSStatus(HexBytes(w.SsStatus), "Ext-CallBarringFeature.SsStatus"); err != nil {
		return ExtCallBarringFeature{}, err
	}
	out := ExtCallBarringFeature{SsStatus: HexBytes(w.SsStatus)}
	if w.BasicService != nil {
		bs, err := convertWireToExtBasicServiceCode(w.BasicService)
		if err != nil {
			return ExtCallBarringFeature{}, fmt.Errorf("BasicService: %w", err)
		}
		out.BasicService = bs
	}
	return out, nil
}

func convertExtCallBarInfoToWire(c *ExtCallBarInfo) (*gsm_map.ExtCallBarInfo, error) {
	if len(c.CallBarringFeatureList) < 1 || len(c.CallBarringFeatureList) > MaxNumOfExtBasicServiceGroups {
		return nil, ErrExtCallBarFeatureListInvalidSize
	}
	list := make(gsm_map.ExtCallBarFeatureList, len(c.CallBarringFeatureList))
	for i := range c.CallBarringFeatureList {
		w, err := convertExtCallBarringFeatureToWire(&c.CallBarringFeatureList[i])
		if err != nil {
			return nil, fmt.Errorf("CallBarringFeatureList[%d]: %w", i, err)
		}
		list[i] = w
	}
	return &gsm_map.ExtCallBarInfo{
		SsCode:                 gsm_map.SSCode{byte(c.SsCode)},
		CallBarringFeatureList: list,
	}, nil
}

func convertWireToExtCallBarInfo(w *gsm_map.ExtCallBarInfo) (*ExtCallBarInfo, error) {
	if len(w.SsCode) != 1 {
		return nil, fmt.Errorf("Ext-CallBarInfo.SsCode: must be 1 octet, got %d", len(w.SsCode))
	}
	if len(w.CallBarringFeatureList) < 1 || len(w.CallBarringFeatureList) > MaxNumOfExtBasicServiceGroups {
		return nil, ErrExtCallBarFeatureListInvalidSize
	}
	out := &ExtCallBarInfo{
		SsCode:                 SsCode(w.SsCode[0]),
		CallBarringFeatureList: make([]ExtCallBarringFeature, len(w.CallBarringFeatureList)),
	}
	for i := range w.CallBarringFeatureList {
		d, err := convertWireToExtCallBarringFeature(&w.CallBarringFeatureList[i])
		if err != nil {
			return nil, fmt.Errorf("CallBarringFeatureList[%d]: %w", i, err)
		}
		out.CallBarringFeatureList[i] = d
	}
	return out, nil
}

// --- CUG-Subscription / CUG-Feature / CUG-Info ---

func isValidIntraCUGOptions(v IntraCUGOptions) bool {
	switch v {
	case IntraCUGNoRestrictions, IntraCUGICCallBarred, IntraCUGOGCallBarred:
		return true
	}
	return false
}

func convertCUGSubscriptionToWire(s *CUGSubscription) (gsm_map.CUGSubscription, error) {
	if s.CugIndex < 0 || s.CugIndex > 32767 {
		return gsm_map.CUGSubscription{}, ErrCUGIndexOutOfRange
	}
	if len(s.CugInterlock) != 4 {
		return gsm_map.CUGSubscription{}, ErrCUGInterlockInvalidSize
	}
	if !isValidIntraCUGOptions(s.IntraCUGOptions) {
		return gsm_map.CUGSubscription{}, ErrIntraCUGOptionsInvalidValue
	}
	out := gsm_map.CUGSubscription{
		CugIndex:        gsm_map.CUGIndex(int64(s.CugIndex)),
		CugInterlock:    gsm_map.CUGInterlock(s.CugInterlock),
		IntraCUGOptions: gsm_map.IntraCUGOptions(int64(s.IntraCUGOptions)),
	}
	if s.BasicServiceGroupList != nil {
		bsgl, err := convertExtBasicServiceGroupListToWire(s.BasicServiceGroupList)
		if err != nil {
			return gsm_map.CUGSubscription{}, fmt.Errorf("BasicServiceGroupList: %w", err)
		}
		out.BasicServiceGroupList = bsgl
	}
	return out, nil
}

func convertWireToCUGSubscription(w *gsm_map.CUGSubscription) (CUGSubscription, error) {
	idxRaw, err := narrowInt64(int64(w.CugIndex))
	if err != nil {
		return CUGSubscription{}, fmt.Errorf("CugIndex: %w", err)
	}
	if idxRaw < 0 || idxRaw > 32767 {
		return CUGSubscription{}, ErrCUGIndexOutOfRange
	}
	if len(w.CugInterlock) != 4 {
		return CUGSubscription{}, ErrCUGInterlockInvalidSize
	}
	optRaw, err := narrowInt64(int64(w.IntraCUGOptions))
	if err != nil {
		return CUGSubscription{}, fmt.Errorf("IntraCUGOptions: %w", err)
	}
	opt := IntraCUGOptions(optRaw)
	if !isValidIntraCUGOptions(opt) {
		return CUGSubscription{}, ErrIntraCUGOptionsInvalidValue
	}
	out := CUGSubscription{
		CugIndex:        idxRaw,
		CugInterlock:    HexBytes(w.CugInterlock),
		IntraCUGOptions: opt,
	}
	if w.BasicServiceGroupList != nil {
		bsgl, err := convertWireToExtBasicServiceGroupList(w.BasicServiceGroupList)
		if err != nil {
			return CUGSubscription{}, fmt.Errorf("BasicServiceGroupList: %w", err)
		}
		out.BasicServiceGroupList = bsgl
	}
	return out, nil
}

func convertCUGFeatureToWire(f *CUGFeature) (gsm_map.CUGFeature, error) {
	out := gsm_map.CUGFeature{
		InterCUGRestrictions: gsm_map.InterCUGRestrictions{f.InterCUGRestrictions},
	}
	if f.BasicService != nil {
		bs, err := convertExtBasicServiceCodeToWire(f.BasicService)
		if err != nil {
			return gsm_map.CUGFeature{}, fmt.Errorf("BasicService: %w", err)
		}
		out.BasicService = bs
	}
	if f.PreferentialCUGIndex != nil {
		v := *f.PreferentialCUGIndex
		if v < 0 || v > 32767 {
			return gsm_map.CUGFeature{}, ErrCUGIndexOutOfRange
		}
		idx := gsm_map.CUGIndex(int64(v))
		out.PreferentialCUGIndicator = &idx
	}
	return out, nil
}

func convertWireToCUGFeature(w *gsm_map.CUGFeature) (CUGFeature, error) {
	if len(w.InterCUGRestrictions) != 1 {
		return CUGFeature{}, fmt.Errorf("CUG-Feature.InterCUGRestrictions: must be exactly 1 octet, got %d", len(w.InterCUGRestrictions))
	}
	out := CUGFeature{InterCUGRestrictions: w.InterCUGRestrictions[0]}
	if w.BasicService != nil {
		bs, err := convertWireToExtBasicServiceCode(w.BasicService)
		if err != nil {
			return CUGFeature{}, fmt.Errorf("BasicService: %w", err)
		}
		out.BasicService = bs
	}
	if w.PreferentialCUGIndicator != nil {
		idx, err := narrowInt64(int64(*w.PreferentialCUGIndicator))
		if err != nil {
			return CUGFeature{}, fmt.Errorf("PreferentialCUGIndicator: %w", err)
		}
		if idx < 0 || idx > 32767 {
			return CUGFeature{}, ErrCUGIndexOutOfRange
		}
		out.PreferentialCUGIndex = &idx
	}
	return out, nil
}

func convertCUGInfoToWire(c *CUGInfo) (*gsm_map.CUGInfo, error) {
	// Per spec the SubscriptionList SIZE is 0..10 (lower bound is 0).
	if len(c.CugSubscriptionList) > MaxNumOfCUG {
		return nil, ErrCUGSubscriptionListInvalidSize
	}
	out := &gsm_map.CUGInfo{}
	if c.CugSubscriptionList != nil {
		subs := make(gsm_map.CUGSubscriptionList, len(c.CugSubscriptionList))
		for i := range c.CugSubscriptionList {
			w, err := convertCUGSubscriptionToWire(&c.CugSubscriptionList[i])
			if err != nil {
				return nil, fmt.Errorf("CugSubscriptionList[%d]: %w", i, err)
			}
			subs[i] = w
		}
		out.CugSubscriptionList = subs
	}
	if c.CugFeatureList != nil {
		if len(c.CugFeatureList) < 1 || len(c.CugFeatureList) > MaxNumOfExtBasicServiceGroups {
			return nil, ErrCUGFeatureListInvalidSize
		}
		feats := make(gsm_map.CUGFeatureList, len(c.CugFeatureList))
		for i := range c.CugFeatureList {
			w, err := convertCUGFeatureToWire(&c.CugFeatureList[i])
			if err != nil {
				return nil, fmt.Errorf("CugFeatureList[%d]: %w", i, err)
			}
			feats[i] = w
		}
		out.CugFeatureList = feats
	}
	return out, nil
}

func convertWireToCUGInfo(w *gsm_map.CUGInfo) (*CUGInfo, error) {
	if len(w.CugSubscriptionList) > MaxNumOfCUG {
		return nil, ErrCUGSubscriptionListInvalidSize
	}
	out := &CUGInfo{}
	if w.CugSubscriptionList != nil {
		out.CugSubscriptionList = make([]CUGSubscription, len(w.CugSubscriptionList))
		for i := range w.CugSubscriptionList {
			d, err := convertWireToCUGSubscription(&w.CugSubscriptionList[i])
			if err != nil {
				return nil, fmt.Errorf("CugSubscriptionList[%d]: %w", i, err)
			}
			out.CugSubscriptionList[i] = d
		}
	}
	if w.CugFeatureList != nil {
		if len(w.CugFeatureList) < 1 || len(w.CugFeatureList) > MaxNumOfExtBasicServiceGroups {
			return nil, ErrCUGFeatureListInvalidSize
		}
		out.CugFeatureList = make([]CUGFeature, len(w.CugFeatureList))
		for i := range w.CugFeatureList {
			d, err := convertWireToCUGFeature(&w.CugFeatureList[i])
			if err != nil {
				return nil, fmt.Errorf("CugFeatureList[%d]: %w", i, err)
			}
			out.CugFeatureList[i] = d
		}
	}
	return out, nil
}

// --- Ext-SS-Data ---

func convertExtSSDataToWire(d *ExtSSData) (*gsm_map.ExtSSData, error) {
	if err := validateExtSSStatus(d.SsStatus, "Ext-SS-Data.SsStatus"); err != nil {
		return nil, err
	}
	out := &gsm_map.ExtSSData{
		SsCode:   gsm_map.SSCode{byte(d.SsCode)},
		SsStatus: gsm_map.ExtSSStatus(d.SsStatus),
	}
	if d.SsSubscriptionOption != nil {
		w, err := convertSSSubscriptionOptionToWire(d.SsSubscriptionOption)
		if err != nil {
			return nil, fmt.Errorf("SsSubscriptionOption: %w", err)
		}
		out.SsSubscriptionOption = w
	}
	if d.BasicServiceGroupList != nil {
		bsgl, err := convertExtBasicServiceGroupListToWire(d.BasicServiceGroupList)
		if err != nil {
			return nil, fmt.Errorf("BasicServiceGroupList: %w", err)
		}
		out.BasicServiceGroupList = bsgl
	}
	return out, nil
}

func convertWireToExtSSData(w *gsm_map.ExtSSData) (*ExtSSData, error) {
	if len(w.SsCode) != 1 {
		return nil, fmt.Errorf("Ext-SS-Data.SsCode: must be 1 octet, got %d", len(w.SsCode))
	}
	if err := validateExtSSStatus(HexBytes(w.SsStatus), "Ext-SS-Data.SsStatus"); err != nil {
		return nil, err
	}
	out := &ExtSSData{
		SsCode:   SsCode(w.SsCode[0]),
		SsStatus: HexBytes(w.SsStatus),
	}
	if w.SsSubscriptionOption != nil {
		d, err := convertWireToSSSubscriptionOption(w.SsSubscriptionOption)
		if err != nil {
			return nil, fmt.Errorf("SsSubscriptionOption: %w", err)
		}
		out.SsSubscriptionOption = d
	}
	if w.BasicServiceGroupList != nil {
		bsgl, err := convertWireToExtBasicServiceGroupList(w.BasicServiceGroupList)
		if err != nil {
			return nil, fmt.Errorf("BasicServiceGroupList: %w", err)
		}
		out.BasicServiceGroupList = bsgl
	}
	return out, nil
}

// --- EMLPP-Info ---

// emlppPriorityRange validates a domain-side EMLPP priority. The encoder
// accepts only the spec's named range 0..6 so the wire never carries a
// 7..15 value the receiver would silently rewrite to 4 per spec exception
// handling.
func emlppPriorityRange(v int, field string) error {
	if v < 0 || v > 6 {
		return fmt.Errorf("%s: %w (got %d)", field, ErrEMLPPPriorityOutOfRange, v)
	}
	return nil
}

func convertEMLPPInfoToWire(e *EMLPPInfo) (*gsm_map.EMLPPInfo, error) {
	if err := emlppPriorityRange(e.MaximumEntitledPriority, "MaximumEntitledPriority"); err != nil {
		return nil, err
	}
	if err := emlppPriorityRange(e.DefaultPriority, "DefaultPriority"); err != nil {
		return nil, err
	}
	return &gsm_map.EMLPPInfo{
		MaximumentitledPriority: gsm_map.EMLPPPriority(int64(e.MaximumEntitledPriority)),
		DefaultPriority:         gsm_map.EMLPPPriority(int64(e.DefaultPriority)),
	}, nil
}

func convertWireToEMLPPInfo(w *gsm_map.EMLPPInfo) (*EMLPPInfo, error) {
	// Lenient decode per TS 29.002: 7..15 → 4. Apply in int64 space.
	mapPriority := func(field string, v int64) (int, error) {
		if v < 0 {
			return 0, fmt.Errorf("%s: %w (got %d)", field, ErrEMLPPPriorityOutOfRange, v)
		}
		if v >= 7 && v <= 15 {
			return 4, nil
		}
		if v > 15 {
			return 0, fmt.Errorf("%s: %w (got %d)", field, ErrEMLPPPriorityOutOfRange, v)
		}
		return int(v), nil
	}
	maxP, err := mapPriority("MaximumEntitledPriority", int64(w.MaximumentitledPriority))
	if err != nil {
		return nil, err
	}
	defP, err := mapPriority("DefaultPriority", int64(w.DefaultPriority))
	if err != nil {
		return nil, err
	}
	return &EMLPPInfo{MaximumEntitledPriority: maxP, DefaultPriority: defP}, nil
}

// --- Ext-SS-Info CHOICE orchestrator ---

func convertExtSSInfoToWire(i *ExtSSInfo) (*gsm_map.ExtSSInfo, error) {
	count := 0
	if i.ForwardingInfo != nil {
		count++
	}
	if i.CallBarringInfo != nil {
		count++
	}
	if i.CugInfo != nil {
		count++
	}
	if i.SsData != nil {
		count++
	}
	if i.EmlppInfo != nil {
		count++
	}
	switch count {
	case 0:
		return nil, ErrExtSSInfoChoiceNoAlternative
	case 1:
		// fall through
	default:
		return nil, ErrExtSSInfoChoiceMultipleAlternatives
	}
	switch {
	case i.ForwardingInfo != nil:
		w, err := convertExtForwInfoToWire(i.ForwardingInfo)
		if err != nil {
			return nil, fmt.Errorf("ForwardingInfo: %w", err)
		}
		v := gsm_map.NewExtSSInfoForwardingInfo(*w)
		return &v, nil
	case i.CallBarringInfo != nil:
		w, err := convertExtCallBarInfoToWire(i.CallBarringInfo)
		if err != nil {
			return nil, fmt.Errorf("CallBarringInfo: %w", err)
		}
		v := gsm_map.NewExtSSInfoCallBarringInfo(*w)
		return &v, nil
	case i.CugInfo != nil:
		w, err := convertCUGInfoToWire(i.CugInfo)
		if err != nil {
			return nil, fmt.Errorf("CugInfo: %w", err)
		}
		v := gsm_map.NewExtSSInfoCugInfo(*w)
		return &v, nil
	case i.SsData != nil:
		w, err := convertExtSSDataToWire(i.SsData)
		if err != nil {
			return nil, fmt.Errorf("SsData: %w", err)
		}
		v := gsm_map.NewExtSSInfoSsData(*w)
		return &v, nil
	default: // EmlppInfo
		w, err := convertEMLPPInfoToWire(i.EmlppInfo)
		if err != nil {
			return nil, fmt.Errorf("EmlppInfo: %w", err)
		}
		v := gsm_map.NewExtSSInfoEmlppInfo(*w)
		return &v, nil
	}
}

func convertWireToExtSSInfo(w *gsm_map.ExtSSInfo) (*ExtSSInfo, error) {
	switch w.Choice {
	case gsm_map.ExtSSInfoChoiceForwardingInfo:
		if w.ForwardingInfo == nil {
			return nil, ErrExtSSInfoChoiceNoAlternative
		}
		d, err := convertWireToExtForwInfo(w.ForwardingInfo)
		if err != nil {
			return nil, fmt.Errorf("ForwardingInfo: %w", err)
		}
		return &ExtSSInfo{ForwardingInfo: d}, nil
	case gsm_map.ExtSSInfoChoiceCallBarringInfo:
		if w.CallBarringInfo == nil {
			return nil, ErrExtSSInfoChoiceNoAlternative
		}
		d, err := convertWireToExtCallBarInfo(w.CallBarringInfo)
		if err != nil {
			return nil, fmt.Errorf("CallBarringInfo: %w", err)
		}
		return &ExtSSInfo{CallBarringInfo: d}, nil
	case gsm_map.ExtSSInfoChoiceCugInfo:
		if w.CugInfo == nil {
			return nil, ErrExtSSInfoChoiceNoAlternative
		}
		d, err := convertWireToCUGInfo(w.CugInfo)
		if err != nil {
			return nil, fmt.Errorf("CugInfo: %w", err)
		}
		return &ExtSSInfo{CugInfo: d}, nil
	case gsm_map.ExtSSInfoChoiceSsData:
		if w.SsData == nil {
			return nil, ErrExtSSInfoChoiceNoAlternative
		}
		d, err := convertWireToExtSSData(w.SsData)
		if err != nil {
			return nil, fmt.Errorf("SsData: %w", err)
		}
		return &ExtSSInfo{SsData: d}, nil
	case gsm_map.ExtSSInfoChoiceEmlppInfo:
		if w.EmlppInfo == nil {
			return nil, ErrExtSSInfoChoiceNoAlternative
		}
		d, err := convertWireToEMLPPInfo(w.EmlppInfo)
		if err != nil {
			return nil, fmt.Errorf("EmlppInfo: %w", err)
		}
		return &ExtSSInfo{EmlppInfo: d}, nil
	default:
		return nil, ErrExtSSInfoChoiceNoAlternative
	}
}
