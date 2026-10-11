package gsmmap

import "testing"

// Extension external clients and PLMN clients: TS 29.002 V19.1.0 §17.7.1.
func TestLCSExtExternalClients(t *testing.T) {
	for _, clients := range []ExtExternalClientList{nil, {makeExternalClient()}} {
		privacy := LCSPrivacyClass{SsCode: 0xb3, SsStatus: HexBytes{1}, ExtExternalClientList: clients}
		if clients != nil {
			privacy.ExternalClientList = ExternalClientList{makeExternalClient(), makeExternalClient(), makeExternalClient(), makeExternalClient(), makeExternalClient()}
		}
		in := &InsertSubscriberDataArg{LcsInformation: &LCSInformation{LcsPrivacyExceptionList: LCSPrivacyExceptionList{privacy, {SsCode: 0xb4, SsStatus: HexBytes{1}, PlmnClientList: PLMNClientList{LCSClientTargetMSsubscribedService}}}}}
		data, err := in.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		out, err := ParseInsertSubscriberData(data)
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "privacy clients", in, out)
	}
}
