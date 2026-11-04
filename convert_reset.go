package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func convertResetArgToWire(v *ResetArg) (*gsm_map.ResetArg, error) {
	if v == nil {
		return nil, ErrResetArgNil
	}
	node, err := convertSendingNodeNumberToWire(&v.SendingNodeNumber)
	if err != nil {
		return nil, fmt.Errorf("SendingNodeNumber: %w", err)
	}
	// 3GPP TS 29.002 V19.1.0 §8.10.1.3: HLR Id list applies only to HLR/HSS restart.
	if v.HLRList != nil && v.SendingNodeNumber.CSSNumber != "" {
		return nil, ErrResetArgHLRListWithCSSNumber
	}
	// 3GPP TS 29.002 V19.1.0 §8.10.1.3: HLR Id list shall not be present with Reset-ID LIST.
	if v.HLRList != nil && v.ResetIdList != nil {
		return nil, ErrResetArgHLRListWithResetIdList
	}
	// 3GPP TS 29.002 V19.1.0 §8.10.1.3: each subscription parameter shall be absent if the other is present.
	if v.SubscriptionData != nil && v.SubscriptionDataDeletion != nil {
		return nil, ErrResetArgSubscriptionDataConflict
	}
	// 3GPP TS 29.002 V19.1.0 §8.10.1.3: Subscription Data shall be absent if Reset-ID LIST is absent.
	if v.SubscriptionData != nil && v.ResetIdList == nil {
		return nil, ErrResetArgSubscriptionDataWithoutResetIdList
	}
	// 3GPP TS 29.002 V19.1.0 §8.10.1.3: Subscription Data Deletion shall be absent if Reset-ID LIST is absent.
	if v.SubscriptionDataDeletion != nil && v.ResetIdList == nil {
		return nil, ErrResetArgSubscriptionDataDeletionWithoutResetIdList
	}
	w := &gsm_map.ResetArg{SendingNodenumber: *node, ResetIdList: convertResetIdListToWire(v.ResetIdList)}
	if v.HLRList != nil {
		w.HlrList = &gsm_map.HLRList{Values: make([]gsm_map.HLRId, len(v.HLRList))}
		for i, id := range v.HLRList {
			b, err := encodeIdentityDigits(identityIMSI, id)
			if err != nil {
				return nil, fmt.Errorf("HLRList[%d]: %w", i, err)
			}
			w.HlrList.Values[i] = b
		}
	}
	if v.SubscriptionData != nil {
		w.SubscriptionData, err = convertInsertSubscriberDataArgToWire(v.SubscriptionData)
		if err != nil {
			return nil, fmt.Errorf("SubscriptionData: %w", err)
		}
	}
	if v.SubscriptionDataDeletion != nil {
		w.SubscriptionDataDeletion, err = convertDeleteSubscriberDataArgToWire(v.SubscriptionDataDeletion)
		if err != nil {
			return nil, fmt.Errorf("SubscriptionDataDeletion: %w", err)
		}
	}
	return w, nil
}

func convertWireToResetArg(w *gsm_map.ResetArg) (*ResetArg, error) {
	node, err := convertWireToSendingNodeNumber(&w.SendingNodenumber)
	if err != nil {
		return nil, fmt.Errorf("SendingNodeNumber: %w", err)
	}
	v := &ResetArg{SendingNodeNumber: *node, ResetIdList: convertWireToResetIdList(w.ResetIdList)}
	if w.HlrList != nil {
		v.HLRList = make([]string, len(w.HlrList.Values))
		for i, id := range w.HlrList.Values {
			v.HLRList[i], err = decodeIdentityDigits(identityIMSI, id)
			if err != nil {
				return nil, fmt.Errorf("HLRList[%d]: %w", i, err)
			}
		}
	}
	if w.SubscriptionData != nil {
		v.SubscriptionData, err = convertWireToInsertSubscriberDataArg(w.SubscriptionData)
		if err != nil {
			return nil, fmt.Errorf("SubscriptionData: %w", err)
		}
	}
	if w.SubscriptionDataDeletion != nil {
		v.SubscriptionDataDeletion, err = convertWireToDeleteSubscriberDataArg(w.SubscriptionDataDeletion)
		if err != nil {
			return nil, fmt.Errorf("SubscriptionDataDeletion: %w", err)
		}
	}
	return v, nil
}

func convertSendingNodeNumberToWire(v *SendingNodeNumber) (*gsm_map.SendingNodeNumber, error) {
	var w gsm_map.SendingNodeNumber
	switch {
	case v.HLRNumber != "" && v.CSSNumber != "":
		return nil, ErrSendingNodeNumberMultipleAlternatives
	case v.HLRNumber != "":
		b, err := encodeAddressField(v.HLRNumber, v.HLRNumberNature, v.HLRNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("HLRNumber: %w", err)
		}
		w = gsm_map.NewSendingNodeNumberHlrNumber(b)
	case v.CSSNumber != "":
		b, err := encodeAddressField(v.CSSNumber, v.CSSNumberNature, v.CSSNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("CSSNumber: %w", err)
		}
		w = gsm_map.NewSendingNodeNumberCssNumber(b)
	default:
		return nil, ErrSendingNodeNumberNoAlternative
	}
	return &w, nil
}

func convertWireToSendingNodeNumber(w *gsm_map.SendingNodeNumber) (*SendingNodeNumber, error) {
	v := &SendingNodeNumber{}
	var err error
	switch w.Choice {
	case gsm_map.SendingNodeNumberChoiceHlrNumber:
		v.HLRNumber, v.HLRNumberNature, v.HLRNumberPlan, err = decodeAddressWithDigits(*w.HlrNumber, ErrSendingNodeNumberHLRNumberDecodedEmpty)
	case gsm_map.SendingNodeNumberChoiceCssNumber:
		v.CSSNumber, v.CSSNumberNature, v.CSSNumberPlan, err = decodeAddressWithDigits(*w.CssNumber, ErrSendingNodeNumberCSSNumberDecodedEmpty)
	default:
		return nil, ErrSendingNodeNumberUnknownAlternative
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}
