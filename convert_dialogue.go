package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime"
)

func convertDialoguePDUToWire(p *DialoguePDU) (*gsm_map.MAPDialoguePDU, error) {
	// TS 29.002 V19.1.0 §17.4: MAP-DialoguePDU selects exactly one alternative.
	if p == nil {
		return nil, ErrDialoguePDUNoAlternative
	}
	count := 0
	for _, set := range []bool{p.Open != nil, p.Accept != nil, p.Close != nil, p.Refuse != nil, p.UserAbort != nil, p.ProviderAbort != nil} {
		if set {
			count++
		}
	}
	if count == 0 {
		return nil, ErrDialoguePDUNoAlternative
	}
	if count > 1 {
		return nil, ErrDialoguePDUMultipleAlternatives
	}
	var w gsm_map.MAPDialoguePDU
	switch {
	case p.Open != nil:
		o, err := convertDialogueOpenToWire(p.Open)
		if err != nil {
			return nil, err
		}
		w = gsm_map.NewMAPDialoguePDUMapOpen(*o)
	case p.Accept != nil:
		w = gsm_map.NewMAPDialoguePDUMapAccept(gsm_map.MAPAcceptInfo{})
	case p.Close != nil:
		w = gsm_map.NewMAPDialoguePDUMapClose(gsm_map.MAPCloseInfo{})
	case p.Refuse != nil:
		r := gsm_map.MAPRefuseInfo{Reason: p.Refuse.Reason}
		if p.Refuse.AlternativeApplicationContext != "" {
			// TS 29.002 V19.1.0 §17.4: optional OBJECT IDENTIFIER.
			// ParseObjectIdentifier accepts only canonical dotted-decimal text.
			oid, err := runtime.ParseObjectIdentifier(string(p.Refuse.AlternativeApplicationContext))
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrInvalidOID, err)
			}
			r.AlternativeApplicationContext = &oid
		}
		w = gsm_map.NewMAPDialoguePDUMapRefuse(r)
	case p.UserAbort != nil:
		u, err := convertDialogueUserAbortToWire(p.UserAbort)
		if err != nil {
			return nil, err
		}
		w = gsm_map.NewMAPDialoguePDUMapUserAbort(gsm_map.MAPUserAbortInfo{MapUserAbortChoice: *u})
	default:
		w = gsm_map.NewMAPDialoguePDUMapProviderAbort(gsm_map.MAPProviderAbortInfo{MapProviderAbortReason: p.ProviderAbort.Reason})
	}
	// Extension containers and unknown additions are not exposed, so the
	// wire values above are built without them.
	return &w, nil
}

func convertWireToDialoguePDU(w *gsm_map.MAPDialoguePDU) (*DialoguePDU, error) {
	p := &DialoguePDU{}
	// TS 29.002 V19.1.0 §17.4: the CHOICE has six alternatives.
	switch w.Choice {
	case gsm_map.MAPDialoguePDUChoiceMapOpen:
		o, err := convertWireToDialogueOpen(w.MapOpen)
		if err != nil {
			return nil, err
		}
		p.Open = o
	case gsm_map.MAPDialoguePDUChoiceMapAccept:
		p.Accept = &DialogueAccept{}
	case gsm_map.MAPDialoguePDUChoiceMapClose:
		p.Close = &DialogueClose{}
	case gsm_map.MAPDialoguePDUChoiceMapRefuse:
		r := &DialogueRefuse{Reason: w.MapRefuse.Reason}
		if w.MapRefuse.AlternativeApplicationContext != nil {
			// TS 29.002 V19.1.0 §17.4: retain the OID without narrowing its arcs;
			// StringChecked reports the runtime's decimal conversion resource limits.
			s, err := w.MapRefuse.AlternativeApplicationContext.StringChecked()
			if err != nil {
				return nil, fmt.Errorf("decoding AlternativeApplicationContext: %w", err)
			}
			r.AlternativeApplicationContext = OID(s)
		}
		p.Refuse = r
	case gsm_map.MAPDialoguePDUChoiceMapUserAbort:
		u, err := convertWireToDialogueUserAbort(&w.MapUserAbort.MapUserAbortChoice)
		if err != nil {
			return nil, err
		}
		p.UserAbort = u
	case gsm_map.MAPDialoguePDUChoiceMapProviderAbort:
		p.ProviderAbort = &DialogueProviderAbort{Reason: w.MapProviderAbort.MapProviderAbortReason}
	default:
		return nil, ErrDialoguePDUUnknownAlternative
	}
	// TS 29.002 V19.1.0 §17.1.4: the codec accepts unsupported extension
	// additions; this package drops them, and extension containers, here.
	return p, nil
}

func convertDialogueOpenToWire(o *DialogueOpen) (*gsm_map.MAPOpenInfo, error) {
	w := &gsm_map.MAPOpenInfo{}
	// TS 29.002 V19.1.0 §§17.4, 17.7.8: optional AddressStrings;
	// empty public digits mean absent, including when nature/plan are set.
	if o.DestinationReference != "" {
		b, err := encodeAddressField(o.DestinationReference, o.DestinationReferenceNature, o.DestinationReferencePlan)
		if err != nil {
			return nil, fmt.Errorf("encoding DestinationReference: %w", err)
		}
		w.DestinationReference = &b
	}
	if o.OriginationReference != "" {
		b, err := encodeAddressField(o.OriginationReference, o.OriginationReferenceNature, o.OriginationReferencePlan)
		if err != nil {
			return nil, fmt.Errorf("encoding OriginationReference: %w", err)
		}
		w.OriginationReference = &b
	}
	return w, nil
}

func convertWireToDialogueOpen(w *gsm_map.MAPOpenInfo) (*DialogueOpen, error) {
	o := &DialogueOpen{}
	// TS 29.002 V19.1.0 §§17.4, 17.7.8: a present optional AddressString
	// must decode to digits to preserve presence through this string API.
	if w.DestinationReference != nil {
		s, n, p, err := decodeAddressWithDigits(*w.DestinationReference, ErrDialogueOpenDestinationReferenceDecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding DestinationReference: %w", err)
		}
		o.DestinationReference, o.DestinationReferenceNature, o.DestinationReferencePlan = s, n, p
	}
	if w.OriginationReference != nil {
		s, n, p, err := decodeAddressWithDigits(*w.OriginationReference, ErrDialogueOpenOriginationReferenceDecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding OriginationReference: %w", err)
		}
		o.OriginationReference, o.OriginationReferenceNature, o.OriginationReferencePlan = s, n, p
	}
	return o, nil
}

func convertDialogueUserAbortToWire(u *DialogueUserAbort) (*gsm_map.MAPUserAbortChoice, error) {
	// TS 29.002 V19.1.0 §17.4: MAP-UserAbortChoice selects one alternative.
	count := 0
	for _, set := range []bool{u.UserSpecificReason, u.UserResourceLimitation, u.ResourceUnavailable != nil, u.ApplicationProcedureCancellation != nil} {
		if set {
			count++
		}
	}
	if count == 0 {
		return nil, ErrDialogueUserAbortNoAlternative
	}
	if count > 1 {
		return nil, ErrDialogueUserAbortMultipleAlternatives
	}
	var w gsm_map.MAPUserAbortChoice
	switch {
	case u.UserSpecificReason:
		w = gsm_map.NewMAPUserAbortChoiceUserSpecificReason(struct{}{})
	case u.UserResourceLimitation:
		w = gsm_map.NewMAPUserAbortChoiceUserResourceLimitation(struct{}{})
	case u.ResourceUnavailable != nil:
		w = gsm_map.NewMAPUserAbortChoiceResourceUnavailable(*u.ResourceUnavailable)
	default:
		w = gsm_map.NewMAPUserAbortChoiceApplicationProcedureCancellation(*u.ApplicationProcedureCancellation)
	}
	return &w, nil
}

func convertWireToDialogueUserAbort(w *gsm_map.MAPUserAbortChoice) (*DialogueUserAbort, error) {
	u := &DialogueUserAbort{}
	// TS 29.002 V19.1.0 §17.4: the CHOICE has four alternatives.
	switch w.Choice {
	case gsm_map.MAPUserAbortChoiceChoiceUserSpecificReason:
		u.UserSpecificReason = true
	case gsm_map.MAPUserAbortChoiceChoiceUserResourceLimitation:
		u.UserResourceLimitation = true
	case gsm_map.MAPUserAbortChoiceChoiceResourceUnavailable:
		v := *w.ResourceUnavailable
		u.ResourceUnavailable = &v
	case gsm_map.MAPUserAbortChoiceChoiceApplicationProcedureCancellation:
		v := *w.ApplicationProcedureCancellation
		u.ApplicationProcedureCancellation = &v
	default:
		return nil, ErrDialogueUserAbortUnknownAlternative
	}
	return u, nil
}
