package gsmmap

import (
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime"
)

func convertTraceNETypeListToBitString(v *TraceNETypeList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.TraceNETypeListMscS: v.MscS,
		gsm_map.TraceNETypeListMgw:  v.Mgw,
		gsm_map.TraceNETypeListSgsn: v.Sgsn,
		gsm_map.TraceNETypeListGgsn: v.Ggsn,
		gsm_map.TraceNETypeListRnc:  v.Rnc,
		gsm_map.TraceNETypeListBmSc: v.Bmsc,
		gsm_map.TraceNETypeListMme:  v.Mme,
		gsm_map.TraceNETypeListSgw:  v.Sgw,
		gsm_map.TraceNETypeListPgw:  v.Pgw,
		gsm_map.TraceNETypeListENB:  v.ENB,
	}
	return gsm_map.NewTraceNETypeList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToTraceNETypeList(bs runtime.BitString) *TraceNETypeList {
	return &TraceNETypeList{
		MscS: bs.Has(gsm_map.TraceNETypeListMscS),
		Mgw:  bs.Has(gsm_map.TraceNETypeListMgw),
		Sgsn: bs.Has(gsm_map.TraceNETypeListSgsn),
		Ggsn: bs.Has(gsm_map.TraceNETypeListGgsn),
		Rnc:  bs.Has(gsm_map.TraceNETypeListRnc),
		Bmsc: bs.Has(gsm_map.TraceNETypeListBmSc),
		Mme:  bs.Has(gsm_map.TraceNETypeListMme),
		Sgw:  bs.Has(gsm_map.TraceNETypeListSgw),
		Pgw:  bs.Has(gsm_map.TraceNETypeListPgw),
		ENB:  bs.Has(gsm_map.TraceNETypeListENB),
	}
}

func convertMSCSInterfaceListToBitString(v *MSCSInterfaceList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.MSCSInterfaceListA:    v.A,
		gsm_map.MSCSInterfaceListIu:   v.Iu,
		gsm_map.MSCSInterfaceListMc:   v.Mc,
		gsm_map.MSCSInterfaceListMapG: v.MapG,
		gsm_map.MSCSInterfaceListMapB: v.MapB,
		gsm_map.MSCSInterfaceListMapE: v.MapE,
		gsm_map.MSCSInterfaceListMapF: v.MapF,
		gsm_map.MSCSInterfaceListCap:  v.Cap,
		gsm_map.MSCSInterfaceListMapD: v.MapD,
		gsm_map.MSCSInterfaceListMapC: v.MapC,
	}
	return gsm_map.NewMSCSInterfaceList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToMSCSInterfaceList(bs runtime.BitString) *MSCSInterfaceList {
	return &MSCSInterfaceList{
		A:    bs.Has(gsm_map.MSCSInterfaceListA),
		Iu:   bs.Has(gsm_map.MSCSInterfaceListIu),
		Mc:   bs.Has(gsm_map.MSCSInterfaceListMc),
		MapG: bs.Has(gsm_map.MSCSInterfaceListMapG),
		MapB: bs.Has(gsm_map.MSCSInterfaceListMapB),
		MapE: bs.Has(gsm_map.MSCSInterfaceListMapE),
		MapF: bs.Has(gsm_map.MSCSInterfaceListMapF),
		Cap:  bs.Has(gsm_map.MSCSInterfaceListCap),
		MapD: bs.Has(gsm_map.MSCSInterfaceListMapD),
		MapC: bs.Has(gsm_map.MSCSInterfaceListMapC),
	}
}

func convertMGWInterfaceListToBitString(v *MGWInterfaceList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.MGWInterfaceListMc:   v.Mc,
		gsm_map.MGWInterfaceListNbUp: v.NbUp,
		gsm_map.MGWInterfaceListIuUp: v.IuUp,
	}
	return gsm_map.NewMGWInterfaceList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToMGWInterfaceList(bs runtime.BitString) *MGWInterfaceList {
	return &MGWInterfaceList{
		Mc:   bs.Has(gsm_map.MGWInterfaceListMc),
		NbUp: bs.Has(gsm_map.MGWInterfaceListNbUp),
		IuUp: bs.Has(gsm_map.MGWInterfaceListIuUp),
	}
}

func convertSGSNInterfaceListToBitString(v *SGSNInterfaceList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.SGSNInterfaceListGb:    v.Gb,
		gsm_map.SGSNInterfaceListIu:    v.Iu,
		gsm_map.SGSNInterfaceListGn:    v.Gn,
		gsm_map.SGSNInterfaceListMapGr: v.MapGr,
		gsm_map.SGSNInterfaceListMapGd: v.MapGd,
		gsm_map.SGSNInterfaceListMapGf: v.MapGf,
		gsm_map.SGSNInterfaceListGs:    v.Gs,
		gsm_map.SGSNInterfaceListGe:    v.Ge,
		gsm_map.SGSNInterfaceListS3:    v.S3,
		gsm_map.SGSNInterfaceListS4:    v.S4,
		gsm_map.SGSNInterfaceListS6d:   v.S6d,
	}
	return gsm_map.NewSGSNInterfaceList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToSGSNInterfaceList(bs runtime.BitString) *SGSNInterfaceList {
	return &SGSNInterfaceList{
		Gb:    bs.Has(gsm_map.SGSNInterfaceListGb),
		Iu:    bs.Has(gsm_map.SGSNInterfaceListIu),
		Gn:    bs.Has(gsm_map.SGSNInterfaceListGn),
		MapGr: bs.Has(gsm_map.SGSNInterfaceListMapGr),
		MapGd: bs.Has(gsm_map.SGSNInterfaceListMapGd),
		MapGf: bs.Has(gsm_map.SGSNInterfaceListMapGf),
		Gs:    bs.Has(gsm_map.SGSNInterfaceListGs),
		Ge:    bs.Has(gsm_map.SGSNInterfaceListGe),
		S3:    bs.Has(gsm_map.SGSNInterfaceListS3),
		S4:    bs.Has(gsm_map.SGSNInterfaceListS4),
		S6d:   bs.Has(gsm_map.SGSNInterfaceListS6d),
	}
}

func convertGGSNInterfaceListToBitString(v *GGSNInterfaceList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.GGSNInterfaceListGn:  v.Gn,
		gsm_map.GGSNInterfaceListGi:  v.Gi,
		gsm_map.GGSNInterfaceListGmb: v.Gmb,
	}
	return gsm_map.NewGGSNInterfaceList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToGGSNInterfaceList(bs runtime.BitString) *GGSNInterfaceList {
	return &GGSNInterfaceList{
		Gn:  bs.Has(gsm_map.GGSNInterfaceListGn),
		Gi:  bs.Has(gsm_map.GGSNInterfaceListGi),
		Gmb: bs.Has(gsm_map.GGSNInterfaceListGmb),
	}
}

func convertRNCInterfaceListToBitString(v *RNCInterfaceList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.RNCInterfaceListIu:  v.Iu,
		gsm_map.RNCInterfaceListIur: v.Iur,
		gsm_map.RNCInterfaceListIub: v.Iub,
		gsm_map.RNCInterfaceListUu:  v.Uu,
	}
	return gsm_map.NewRNCInterfaceList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToRNCInterfaceList(bs runtime.BitString) *RNCInterfaceList {
	return &RNCInterfaceList{
		Iu:  bs.Has(gsm_map.RNCInterfaceListIu),
		Iur: bs.Has(gsm_map.RNCInterfaceListIur),
		Iub: bs.Has(gsm_map.RNCInterfaceListIub),
		Uu:  bs.Has(gsm_map.RNCInterfaceListUu),
	}
}

func convertBMSCInterfaceListToBitString(v *BMSCInterfaceList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.BMSCInterfaceListGmb: v.Gmb,
	}
	return gsm_map.NewBMSCInterfaceList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToBMSCInterfaceList(bs runtime.BitString) *BMSCInterfaceList {
	return &BMSCInterfaceList{
		Gmb: bs.Has(gsm_map.BMSCInterfaceListGmb),
	}
}

func convertMMEInterfaceListToBitString(v *MMEInterfaceList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.MMEInterfaceListS1Mme: v.S1Mme,
		gsm_map.MMEInterfaceListS3:    v.S3,
		gsm_map.MMEInterfaceListS6a:   v.S6a,
		gsm_map.MMEInterfaceListS10:   v.S10,
		gsm_map.MMEInterfaceListS11:   v.S11,
	}
	return gsm_map.NewMMEInterfaceList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToMMEInterfaceList(bs runtime.BitString) *MMEInterfaceList {
	return &MMEInterfaceList{
		S1Mme: bs.Has(gsm_map.MMEInterfaceListS1Mme),
		S3:    bs.Has(gsm_map.MMEInterfaceListS3),
		S6a:   bs.Has(gsm_map.MMEInterfaceListS6a),
		S10:   bs.Has(gsm_map.MMEInterfaceListS10),
		S11:   bs.Has(gsm_map.MMEInterfaceListS11),
	}
}

func convertSGWInterfaceListToBitString(v *SGWInterfaceList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.SGWInterfaceListS4:  v.S4,
		gsm_map.SGWInterfaceListS5:  v.S5,
		gsm_map.SGWInterfaceListS8b: v.S8b,
		gsm_map.SGWInterfaceListS11: v.S11,
		gsm_map.SGWInterfaceListGxc: v.Gxc,
	}
	return gsm_map.NewSGWInterfaceList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToSGWInterfaceList(bs runtime.BitString) *SGWInterfaceList {
	return &SGWInterfaceList{
		S4:  bs.Has(gsm_map.SGWInterfaceListS4),
		S5:  bs.Has(gsm_map.SGWInterfaceListS5),
		S8b: bs.Has(gsm_map.SGWInterfaceListS8b),
		S11: bs.Has(gsm_map.SGWInterfaceListS11),
		Gxc: bs.Has(gsm_map.SGWInterfaceListGxc),
	}
}

func convertPGWInterfaceListToBitString(v *PGWInterfaceList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.PGWInterfaceListS2a: v.S2a,
		gsm_map.PGWInterfaceListS2b: v.S2b,
		gsm_map.PGWInterfaceListS2c: v.S2c,
		gsm_map.PGWInterfaceListS5:  v.S5,
		gsm_map.PGWInterfaceListS6b: v.S6b,
		gsm_map.PGWInterfaceListGx:  v.Gx,
		gsm_map.PGWInterfaceListS8b: v.S8b,
		gsm_map.PGWInterfaceListSgi: v.Sgi,
	}
	return gsm_map.NewPGWInterfaceList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToPGWInterfaceList(bs runtime.BitString) *PGWInterfaceList {
	return &PGWInterfaceList{
		S2a: bs.Has(gsm_map.PGWInterfaceListS2a),
		S2b: bs.Has(gsm_map.PGWInterfaceListS2b),
		S2c: bs.Has(gsm_map.PGWInterfaceListS2c),
		S5:  bs.Has(gsm_map.PGWInterfaceListS5),
		S6b: bs.Has(gsm_map.PGWInterfaceListS6b),
		Gx:  bs.Has(gsm_map.PGWInterfaceListGx),
		S8b: bs.Has(gsm_map.PGWInterfaceListS8b),
		Sgi: bs.Has(gsm_map.PGWInterfaceListSgi),
	}
}

func convertENBInterfaceListToBitString(v *ENBInterfaceList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.ENBInterfaceListS1Mme: v.S1Mme,
		gsm_map.ENBInterfaceListX2:    v.X2,
		gsm_map.ENBInterfaceListUu:    v.Uu,
	}
	return gsm_map.NewENBInterfaceList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToENBInterfaceList(bs runtime.BitString) *ENBInterfaceList {
	return &ENBInterfaceList{
		S1Mme: bs.Has(gsm_map.ENBInterfaceListS1Mme),
		X2:    bs.Has(gsm_map.ENBInterfaceListX2),
		Uu:    bs.Has(gsm_map.ENBInterfaceListUu),
	}
}

func convertMSCSEventListToBitString(v *MSCSEventList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.MSCSEventListMoMtCall:               v.MoMtCall,
		gsm_map.MSCSEventListMoMtSms:                v.MoMtSms,
		gsm_map.MSCSEventListLuImsiAttachImsiDetach: v.LuImsiAttachImsiDetach,
		gsm_map.MSCSEventListHandovers:              v.Handovers,
		gsm_map.MSCSEventListSs:                     v.Ss,
	}
	return gsm_map.NewMSCSEventList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToMSCSEventList(bs runtime.BitString) *MSCSEventList {
	return &MSCSEventList{
		MoMtCall:               bs.Has(gsm_map.MSCSEventListMoMtCall),
		MoMtSms:                bs.Has(gsm_map.MSCSEventListMoMtSms),
		LuImsiAttachImsiDetach: bs.Has(gsm_map.MSCSEventListLuImsiAttachImsiDetach),
		Handovers:              bs.Has(gsm_map.MSCSEventListHandovers),
		Ss:                     bs.Has(gsm_map.MSCSEventListSs),
	}
}

func convertMGWEventListToBitString(v *MGWEventList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.MGWEventListContext: v.Context,
	}
	return gsm_map.NewMGWEventList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToMGWEventList(bs runtime.BitString) *MGWEventList {
	return &MGWEventList{
		Context: bs.Has(gsm_map.MGWEventListContext),
	}
}

func convertSGSNEventListToBitString(v *SGSNEventList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.SGSNEventListPdpContext:              v.PdpContext,
		gsm_map.SGSNEventListMoMtSms:                 v.MoMtSms,
		gsm_map.SGSNEventListRauGprsAttachGprsDetach: v.RauGprsAttachGprsDetach,
		gsm_map.SGSNEventListMbmsContext:             v.MbmsContext,
	}
	return gsm_map.NewSGSNEventList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToSGSNEventList(bs runtime.BitString) *SGSNEventList {
	return &SGSNEventList{
		PdpContext:              bs.Has(gsm_map.SGSNEventListPdpContext),
		MoMtSms:                 bs.Has(gsm_map.SGSNEventListMoMtSms),
		RauGprsAttachGprsDetach: bs.Has(gsm_map.SGSNEventListRauGprsAttachGprsDetach),
		MbmsContext:             bs.Has(gsm_map.SGSNEventListMbmsContext),
	}
}

func convertGGSNEventListToBitString(v *GGSNEventList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.GGSNEventListPdpContext:  v.PdpContext,
		gsm_map.GGSNEventListMbmsContext: v.MbmsContext,
	}
	return gsm_map.NewGGSNEventList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToGGSNEventList(bs runtime.BitString) *GGSNEventList {
	return &GGSNEventList{
		PdpContext:  bs.Has(gsm_map.GGSNEventListPdpContext),
		MbmsContext: bs.Has(gsm_map.GGSNEventListMbmsContext),
	}
}

func convertBMSCEventListToBitString(v *BMSCEventList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.BMSCEventListMbmsMulticastServiceActivation: v.MbmsMulticastServiceActivation,
	}
	return gsm_map.NewBMSCEventList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToBMSCEventList(bs runtime.BitString) *BMSCEventList {
	return &BMSCEventList{
		MbmsMulticastServiceActivation: bs.Has(gsm_map.BMSCEventListMbmsMulticastServiceActivation),
	}
}

func convertMMEEventListToBitString(v *MMEEventList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.MMEEventListUeInitiatedPDNconectivityRequest:      v.UeInitiatedPDNConnectivityRequest,
		gsm_map.MMEEventListServiceRequestts:                      v.ServiceRequest,
		gsm_map.MMEEventListInitialAttachTrackingAreaUpdateDetach: v.InitialAttachTrackingAreaUpdateDetach,
		gsm_map.MMEEventListUeInitiatedPDNdisconnection:           v.UeInitiatedPDNdisconnection,
		gsm_map.MMEEventListBearerActivationModificationDeletion:  v.BearerActivationModificationDeletion,
		gsm_map.MMEEventListHandover:                              v.Handover,
	}
	return gsm_map.NewMMEEventList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToMMEEventList(bs runtime.BitString) *MMEEventList {
	return &MMEEventList{
		UeInitiatedPDNConnectivityRequest:     bs.Has(gsm_map.MMEEventListUeInitiatedPDNconectivityRequest),
		ServiceRequest:                        bs.Has(gsm_map.MMEEventListServiceRequestts),
		InitialAttachTrackingAreaUpdateDetach: bs.Has(gsm_map.MMEEventListInitialAttachTrackingAreaUpdateDetach),
		UeInitiatedPDNdisconnection:           bs.Has(gsm_map.MMEEventListUeInitiatedPDNdisconnection),
		BearerActivationModificationDeletion:  bs.Has(gsm_map.MMEEventListBearerActivationModificationDeletion),
		Handover:                              bs.Has(gsm_map.MMEEventListHandover),
	}
}

func convertSGWEventListToBitString(v *SGWEventList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.SGWEventListPdnConnectionCreation:                v.PdnConnectionCreation,
		gsm_map.SGWEventListPdnConnectionTermination:             v.PdnConnectionTermination,
		gsm_map.SGWEventListBearerActivationModificationDeletion: v.BearerActivationModificationDeletion,
	}
	return gsm_map.NewSGWEventList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToSGWEventList(bs runtime.BitString) *SGWEventList {
	return &SGWEventList{
		PdnConnectionCreation:                bs.Has(gsm_map.SGWEventListPdnConnectionCreation),
		PdnConnectionTermination:             bs.Has(gsm_map.SGWEventListPdnConnectionTermination),
		BearerActivationModificationDeletion: bs.Has(gsm_map.SGWEventListBearerActivationModificationDeletion),
	}
}

func convertPGWEventListToBitString(v *PGWEventList) (runtime.BitString, error) {
	bits := []bool{
		gsm_map.PGWEventListPdnConnectionCreation:                v.PdnConnectionCreation,
		gsm_map.PGWEventListPdnConnectionTermination:             v.PdnConnectionTermination,
		gsm_map.PGWEventListBearerActivationModificationDeletion: v.BearerActivationModificationDeletion,
	}
	return gsm_map.NewPGWEventList(namedBitPositions(bits)...)
}

// 3GPP TS 29.002 V19.1.0 §17.7.2: "Other bits than listed above shall be discarded."
func convertBitStringToPGWEventList(bs runtime.BitString) *PGWEventList {
	return &PGWEventList{
		PdnConnectionCreation:                bs.Has(gsm_map.PGWEventListPdnConnectionCreation),
		PdnConnectionTermination:             bs.Has(gsm_map.PGWEventListPdnConnectionTermination),
		BearerActivationModificationDeletion: bs.Has(gsm_map.PGWEventListBearerActivationModificationDeletion),
	}
}
