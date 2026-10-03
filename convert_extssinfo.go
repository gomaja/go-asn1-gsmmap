// Ext-SS-Info CHOICE converters for InsertSubscriberData.
//
// Covers TS 29.002 MAP-MS-DataTypes.asn:1826-onwards: the 5-alternative
// CHOICE used inside Ext-SS-InfoList plus all directly-referenced
// nested SEQUENCEs (Ext-ForwInfo, Ext-CallBarInfo, CUG-Info,
// Ext-SS-Data, EMLPP-Info) and CHOICEs (SS-SubscriptionOption).
//
// Reuses ExtBasicServiceCode and the SS-Code typedef.
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
		v := gsm_map.NewSSSubscriptionOptionCliRestrictionOption(*o.CliRestriction)
		return &v, nil
	case hasOver:
		if !isValidOverrideCategory(*o.Override) {
			return nil, ErrOverrideCategoryInvalidValue
		}
		v := gsm_map.NewSSSubscriptionOptionOverrideCategory(*o.Override)
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

func convertExtBasicServiceGroupListToWire(in []ExtBasicServiceCode) (*gsm_map.ExtBasicServiceGroupList, error) {
	if in == nil {
		return nil, nil
	}

	out := gsm_map.ExtBasicServiceGroupList{Values: make([]gsm_map.ExtBasicServiceCode, len(in))}
	for i := range in {
		w, err := convertExtBasicServiceCodeToWire(&in[i])
		if err != nil {
			return nil, fmt.Errorf("BasicServiceGroupList[%d]: %w", i, err)
		}
		out.Values[i] = *w
	}
	return &out, nil
}

func convertWireToExtBasicServiceGroupList(w *gsm_map.ExtBasicServiceGroupList) ([]ExtBasicServiceCode, error) {
	if w == nil {
		return nil, nil
	}

	out := make([]ExtBasicServiceCode, len(w.Values))
	for i := range w.Values {
		d, err := convertWireToExtBasicServiceCode(&w.Values[i])
		if err != nil {
			return nil, fmt.Errorf("BasicServiceGroupList[%d]: %w", i, err)
		}
		out[i] = *d
	}
	return out, nil
}

// --- Ext-ForwFeature / Ext-ForwInfo ---

func convertExtForwFeatureToWire(f *ExtForwFeature) (gsm_map.ExtForwFeature, error) {
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
		v := enc
		out.ForwardedToNumber = &v
	}
	if f.ForwardedToSubaddress != nil {
		v := gsm_map.ISDNSubaddressString(f.ForwardedToSubaddress)
		out.ForwardedToSubaddress = &v
	}
	if f.ForwardingOptions != nil {
		v := gsm_map.ExtForwOptions(f.ForwardingOptions)
		out.ForwardingOptions = &v
	}
	if f.NoReplyConditionTime != nil {
		// 3GPP TS 29.002 V19.1.0 §17.7.1 Ext-NoRepCondTime: "Only values
		// 5-30 are used. Values in the ranges 1-4 and 31-100 are reserved
		// for future use".
		if *f.NoReplyConditionTime < 5 || *f.NoReplyConditionTime > 30 {
			return gsm_map.ExtForwFeature{}, fmt.Errorf("NoReplyConditionTime: %w (got %d)", ErrNoReplyConditionTimeOutOfRange, *f.NoReplyConditionTime)
		}
		v64 := int64(*f.NoReplyConditionTime)

		v := v64
		out.NoReplyConditionTime = &v
	}
	if f.LongForwardedToNumber != "" {
		enc, err := encodeAddressField(f.LongForwardedToNumber, f.ForwardedToNature, f.ForwardedToPlan)
		if err != nil {
			return gsm_map.ExtForwFeature{}, fmt.Errorf("LongForwardedToNumber: %w", err)
		}
		v := enc
		out.LongForwardedToNumber = &v
	}
	return out, nil
}

func convertWireToExtForwFeature(w *gsm_map.ExtForwFeature) (ExtForwFeature, error) {
	out := ExtForwFeature{SsStatus: HexBytes(w.SsStatus)}
	if w.BasicService != nil {
		bs, err := convertWireToExtBasicServiceCode(w.BasicService)
		if err != nil {
			return ExtForwFeature{}, fmt.Errorf("BasicService: %w", err)
		}
		out.BasicService = bs
	}
	if w.ForwardedToNumber != nil {
		digits, nat, plan, err := decodeAddressWithDigits(*w.ForwardedToNumber, ErrExtForwFeatureForwardedToNumberDecodedEmpty)
		if err != nil {
			return ExtForwFeature{}, fmt.Errorf("ForwardedToNumber: %w", err)
		}
		out.ForwardedToNumber = digits
		out.ForwardedToNature = nat
		out.ForwardedToPlan = plan
	}
	if w.ForwardedToSubaddress != nil {
		out.ForwardedToSubaddress = HexBytes(*w.ForwardedToSubaddress)
	}
	if w.ForwardingOptions != nil {
		out.ForwardingOptions = HexBytes(*w.ForwardingOptions)
	}
	if w.NoReplyConditionTime != nil {
		// 3GPP TS 29.002 V19.1.0 §17.7.1 Ext-NoRepCondTime: "If received:
		// values 1-4 shall be mapped on to value 5", "values 31-100 shall be
		// mapped on to value 30". The codec has enforced INTEGER (1..100).
		v64 := *w.NoReplyConditionTime

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
		digits, nat, plan, err := decodeAddressWithDigits(*w.LongForwardedToNumber, ErrExtForwFeatureLongForwardedToNumberDecodedEmpty)
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
	list := gsm_map.ExtForwFeatureList{Values: make([]gsm_map.ExtForwFeature, len(f.ForwardingFeatureList))}
	for i := range f.ForwardingFeatureList {
		w, err := convertExtForwFeatureToWire(&f.ForwardingFeatureList[i])
		if err != nil {
			return nil, fmt.Errorf("ForwardingFeatureList[%d]: %w", i, err)
		}
		list.Values[i] = w
	}
	return &gsm_map.ExtForwInfo{
		SsCode:                gsm_map.SSCode{byte(f.SsCode)},
		ForwardingFeatureList: &list,
	}, nil
}

func convertWireToExtForwInfo(w *gsm_map.ExtForwInfo) (*ExtForwInfo, error) {
	out := &ExtForwInfo{
		SsCode:                SsCode(w.SsCode[0]),
		ForwardingFeatureList: make([]ExtForwFeature, len(w.ForwardingFeatureList.Values)),
	}
	for i := range w.ForwardingFeatureList.Values {
		d, err := convertWireToExtForwFeature(&w.ForwardingFeatureList.Values[i])
		if err != nil {
			return nil, fmt.Errorf("ForwardingFeatureList[%d]: %w", i, err)
		}
		out.ForwardingFeatureList[i] = d
	}
	return out, nil
}

// --- Ext-CallBarringFeature / Ext-CallBarInfo ---

func convertExtCallBarringFeatureToWire(f *ExtCallBarringFeature) (gsm_map.ExtCallBarringFeature, error) {
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
	list := gsm_map.ExtCallBarFeatureList{Values: make([]gsm_map.ExtCallBarringFeature, len(c.CallBarringFeatureList))}
	for i := range c.CallBarringFeatureList {
		w, err := convertExtCallBarringFeatureToWire(&c.CallBarringFeatureList[i])
		if err != nil {
			return nil, fmt.Errorf("CallBarringFeatureList[%d]: %w", i, err)
		}
		list.Values[i] = w
	}
	return &gsm_map.ExtCallBarInfo{
		SsCode:                 gsm_map.SSCode{byte(c.SsCode)},
		CallBarringFeatureList: &list,
	}, nil
}

func convertWireToExtCallBarInfo(w *gsm_map.ExtCallBarInfo) (*ExtCallBarInfo, error) {
	out := &ExtCallBarInfo{
		SsCode:                 SsCode(w.SsCode[0]),
		CallBarringFeatureList: make([]ExtCallBarringFeature, len(w.CallBarringFeatureList.Values)),
	}
	for i := range w.CallBarringFeatureList.Values {
		d, err := convertWireToExtCallBarringFeature(&w.CallBarringFeatureList.Values[i])
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
	if !isValidIntraCUGOptions(s.IntraCUGOptions) {
		return gsm_map.CUGSubscription{}, ErrIntraCUGOptionsInvalidValue
	}
	out := gsm_map.CUGSubscription{
		CugIndex:        int64(s.CugIndex),
		CugInterlock:    gsm_map.CUGInterlock(s.CugInterlock),
		IntraCUGOptions: s.IntraCUGOptions,
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
	optRaw, err := narrowInt64(int64(w.IntraCUGOptions))
	if err != nil {
		return CUGSubscription{}, fmt.Errorf("IntraCUGOptions: %w", err)
	}
	opt := IntraCUGOptions(optRaw)
	if !isValidIntraCUGOptions(opt) {
		return CUGSubscription{}, ErrIntraCUGOptionsInvalidValue
	}
	out := CUGSubscription{
		CugIndex:        int(w.CugIndex),
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
		idx := int64(v)
		out.PreferentialCUGIndicator = &idx
	}
	return out, nil
}

func convertWireToCUGFeature(w *gsm_map.CUGFeature) (CUGFeature, error) {
	out := CUGFeature{InterCUGRestrictions: w.InterCUGRestrictions[0]}
	if w.BasicService != nil {
		bs, err := convertWireToExtBasicServiceCode(w.BasicService)
		if err != nil {
			return CUGFeature{}, fmt.Errorf("BasicService: %w", err)
		}
		out.BasicService = bs
	}
	if w.PreferentialCUGIndicator != nil {
		idx := int(*w.PreferentialCUGIndicator)
		out.PreferentialCUGIndex = &idx
	}
	return out, nil
}

func convertCUGInfoToWire(c *CUGInfo) (*gsm_map.CUGInfo, error) {
	// cug-SubscriptionList is mandatory and CUG-SubscriptionList ::=
	// SEQUENCE SIZE (0..maxNumOfCUG) (3GPP TS 29.002 V19.1.0 §17.7.1), so
	// a nil list is the empty list and is always encoded.
	subs := gsm_map.CUGSubscriptionList{Values: make([]gsm_map.CUGSubscription, len(c.CugSubscriptionList))}
	for i := range c.CugSubscriptionList {
		w, err := convertCUGSubscriptionToWire(&c.CugSubscriptionList[i])
		if err != nil {
			return nil, fmt.Errorf("CugSubscriptionList[%d]: %w", i, err)
		}
		subs.Values[i] = w
	}
	out := &gsm_map.CUGInfo{CugSubscriptionList: &subs}
	if c.CugFeatureList != nil {
		feats := gsm_map.CUGFeatureList{Values: make([]gsm_map.CUGFeature, len(c.CugFeatureList))}
		for i := range c.CugFeatureList {
			w, err := convertCUGFeatureToWire(&c.CugFeatureList[i])
			if err != nil {
				return nil, fmt.Errorf("CugFeatureList[%d]: %w", i, err)
			}
			feats.Values[i] = w
		}
		out.CugFeatureList = &feats
	}
	return out, nil
}

func convertWireToCUGInfo(w *gsm_map.CUGInfo) (*CUGInfo, error) {
	out := &CUGInfo{}
	// An empty list decodes to nil, the zero value that encodes it.
	if w.CugSubscriptionList != nil && len(w.CugSubscriptionList.Values) > 0 {
		out.CugSubscriptionList = make([]CUGSubscription, len(w.CugSubscriptionList.Values))
		for i := range w.CugSubscriptionList.Values {
			d, err := convertWireToCUGSubscription(&w.CugSubscriptionList.Values[i])
			if err != nil {
				return nil, fmt.Errorf("CugSubscriptionList[%d]: %w", i, err)
			}
			out.CugSubscriptionList[i] = d
		}
	}
	if w.CugFeatureList != nil {
		out.CugFeatureList = make([]CUGFeature, len(w.CugFeatureList.Values))
		for i := range w.CugFeatureList.Values {
			d, err := convertWireToCUGFeature(&w.CugFeatureList.Values[i])
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
		MaximumentitledPriority: int64(e.MaximumEntitledPriority),
		DefaultPriority:         int64(e.DefaultPriority),
	}, nil
}

func convertWireToEMLPPInfo(w *gsm_map.EMLPPInfo) (*EMLPPInfo, error) {
	// Spare values 7..15 map to priority 4 (3GPP TS 29.002 V19.1.0 §17.7.8).
	mapPriority := func(v int64) int {
		if v >= 7 && v <= 15 {
			return 4
		}
		return int(v)
	}
	maxP := mapPriority(w.MaximumentitledPriority)
	defP := mapPriority(w.DefaultPriority)
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
