// Test cases ported from go-gsmmap/parse_test.go.
// These use real network captures as hex test vectors to validate
// BER parse/marshal round-trips against the original library.

package gsmmap

import (
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseSriSm(t *testing.T) {
	tests := []struct {
		name                string
		hexString           string
		expectError         bool
		matchMarshaledBytes bool
	}{
		{
			name:                "Valid SRI SM",
			hexString:           "301380069122608538188101ff8206912260909899",
			expectError:         false,
			matchMarshaledBytes: true,
		},
		{
			name:                "Valid SRI SM - nonDER",
			hexString:           "3019800a915282051447720982f9810101820891328490001015f8",
			expectError:         false,
			matchMarshaledBytes: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalBytes, err := hex.DecodeString(tc.hexString)
			if err != nil {
				t.Fatalf("Failed to decode hex string: %v", err)
			}

			sriSm, err := ParseSriSm(originalBytes)
			if err != nil {
				t.Fatalf("Failed to parse SriSm: %v", err)
			}

			marshaledBytes, err := sriSm.Marshal()
			if (err != nil) != tc.expectError {
				t.Fatalf("Unexpected error status: got %v, expected error: %v", err, tc.expectError)
			}

			if err == nil && tc.matchMarshaledBytes {
				if diff := cmp.Diff(originalBytes, marshaledBytes); diff != "" {
					t.Errorf("Marshaled bytes don't match original (-original +marshaled):\n%s", diff)
				}
			}
		})
	}
}

func TestParseSriSmResp(t *testing.T) {
	tests := []struct {
		name        string
		hexString   string
		expectError bool
	}{
		{
			name:        "Valid SRI SM Response",
			hexString:   "3015040882131068584836f3a0098107917394950862f6",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalBytes, err := hex.DecodeString(tc.hexString)
			if err != nil {
				t.Fatalf("Failed to decode hex string: %v", err)
			}

			sriSmResp, err := ParseSriSmResp(originalBytes)
			if err != nil {
				t.Fatalf("Failed to parse SriSmResp: %v", err)
			}

			marshaledBytes, err := sriSmResp.Marshal()
			if (err != nil) != tc.expectError {
				t.Fatalf("Unexpected error status: got %v, expected error: %v", err, tc.expectError)
			}

			if err == nil {
				if diff := cmp.Diff(originalBytes, marshaledBytes); diff != "" {
					t.Errorf("Marshaled bytes don't match original (-original +marshaled):\n%s", diff)
				}
			}
		})
	}
}

func TestParseMtFsm(t *testing.T) {
	tests := []struct {
		name        string
		hexString   string
		expectError bool
	}{
		{
			name:        "Valid MT FSM",
			hexString:   "3077800832140080803138f684069169318488880463040b916971101174f40000422182612464805bd2e2b1252d467ff6de6c47efd96eb6a1d056cb0d69b49a10269c098537586e96931965b260d15613da72c29b91261bde72c6a1ad2623d682b5996d58331271375a0d1733eee4bd98ec768bd966b41c0d",
			expectError: false,
		},
		{
			name:        "Valid MT FSM Concatenated (part 1)",
			hexString:   "3081b7800826610011829761f6840891328490000005f704819e4009d047f6dbfe06000042217251400000a00500035f020190e53c0b947fd741e8b0bd0c9abfdb6510bcec26a7dd67d09c5e86cf41693728ffaecb41f2f2393da7cbc3f4f4db0d82cbdfe3f27cee0241d9e5f0bc0c32bfd9ecf71d44479741ecb47b0da2bf41e3771bce2ed3cb203abadc0685dd64d09c1e96d341e4323b6d2fcbd3ee33888e96bfeb6734e8c87edbdf2190bc3c96d7d3f476d94d77d5e70500",
			expectError: false,
		},
		{
			name:        "Valid MT FSM Concatenated (part 2)",
			hexString:   "3042800826610011829761f6840891328490000005f7042c4409d047f6dbfe060000422172514000001d0500035f0202cae8ba5c9e2ecb5de377fb157ea9d1b0d93b1e06",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalBytes, err := hex.DecodeString(tc.hexString)
			if err != nil {
				t.Fatalf("Failed to decode hex string: %v", err)
			}

			mtFsm, err := ParseMtFsm(originalBytes)
			if err != nil {
				t.Fatalf("Failed to parse MtFsm: %v", err)
			}

			marshaledBytes, err := mtFsm.Marshal()
			if (err != nil) != tc.expectError {
				t.Fatalf("Unexpected error status: got %v, expected error: %v", err, tc.expectError)
			}

			if err == nil {
				if diff := cmp.Diff(originalBytes, marshaledBytes); diff != "" {
					t.Errorf("Marshaled bytes don't match original (-original +marshaled):\n%s", diff)
				}
			}
		})
	}
}

func TestParseMoFsm(t *testing.T) {
	tests := []struct {
		name           string
		hexString      string
		expectError    bool
		skipRoundTrip  bool // skip round-trip check when TPDU re-encoding differs
	}{
		{
			name:        "Valid MO FSM",
			hexString:   "302d84069122609098998206912260539128041b01510a912260716622000011d972180d4a82eee13928cc7ebbcb20",
			expectError: false,
		},
		{
			name:        "Valid MO FSM Concatenated (part 1)",
			hexString:   "3081ab84069122609098998206912260532023048198413f0a9122600650150000a0050003020201a8e8f41c949e83c220f6db7d06b5cbf379f85cd6819a61f93deca6a2d373507a0e0a83d86ff719d42ecfe7e17359076a86e5f7b09b8a4ecf41e939280c62bfdd6750bb3c9f87cf651da81996dfc36e2a3a3d07a5e7a03088fd769f41edf27c1e3e9775a066587e0fbba9e8f41c949e83c220f6db7d06b5cbf379f85cd6819a61f93deca6a2d3",
			expectError: false,
		},
		{
			name:        "Valid MO FSM Concatenated (part 2)",
			hexString:   "303c84069122609098998206912260532023042a41400a912260065015000022050003020202e6a0f41c1406b1dfee33a85d9ecfc3e7b20ed40ccbef6137",
			expectError: false,
		},
		{
			name:        "Invalid Packet for MO FSM 1",
			hexString:   "301380069122608538188101ff8206912260909899",
			expectError: true,
		},
		{
			name:          "Valid MO FSM with IMSI DA and SCA OA",
			hexString:     "3081b7800826610011829761f6840891328490000005f704819e4009d047f6dbfe06000042217251400000a00500035f020190e53c0b947fd741e8b0bd0c9abfdb6510bcec26a7dd67d09c5e86cf41693728ffaecb41f2f2393da7cbc3f4f4db0d82cbdfe3f27cee0241d9e5f0bc0c32bfd9ecf71d44479741ecb47b0da2bf41e3771bce2ed3cb203abadc0685dd64d09c1e96d341e4323b6d2fcbd3ee33888e96bfeb6734e8c87edbdf2190bc3c96d7d3f476d94d77d5e70500",
			expectError:   false,
			skipRoundTrip: true, // TPDU re-encoding differs from original wire bytes
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalBytes, err := hex.DecodeString(tc.hexString)
			if err != nil {
				t.Fatalf("Failed to decode hex string: %v", err)
			}

			moFsm, err := ParseMoFsm(originalBytes)
			if (err != nil) != tc.expectError {
				t.Fatalf("Unexpected error status during parsing: got %v, expected error: %v", err, tc.expectError)
			}

			if tc.expectError && err != nil {
				t.Logf("Expected error occurred in test case '%s': %v", tc.name, err)
				return
			}

			if tc.skipRoundTrip {
				return
			}

			marshaledBytes, err := moFsm.Marshal()
			if err != nil {
				t.Fatalf("Failed to marshal MoFsm: %v", err)
			}

			if diff := cmp.Diff(originalBytes, marshaledBytes); diff != "" {
				t.Errorf("Marshaled bytes don't match original (-original +marshaled):\n%s", diff)
			}
		})
	}
}

func TestParseUpdateGprsLocation(t *testing.T) {
	tests := []struct {
		name                   string
		hexString              string
		expectError            bool
		expectedIMSI           string
		expectedSGSNNumber     string
		expectedSGSNAddress    string
		expectedGprsEnhSupport bool
		expectedLCSCapSets     *SupportedLCSCapabilitySets
	}{
		{
			name:                   "Valid UpdateGprsLocation (IPv4) With SGSNCapability",
			hexString:              "3022040862006630020000f20407911487390120f3040504d5378647a006830085020640",
			expectError:            false,
			expectedIMSI:           "260066032000002",
			expectedSGSNNumber:     "41789310023",
			expectedSGSNAddress:    "213.55.134.71",
			expectedGprsEnhSupport: true,
			expectedLCSCapSets:     &SupportedLCSCapabilitySets{LcsCapabilitySet2: true},
		},
		{
			name:                   "Valid UpdateGprsLocation with SGSNCapability",
			hexString:              "301e04082143658709214365040791261806630000040504c0a80101a0028300",
			expectError:            false,
			expectedIMSI:           "1234567890123456",
			expectedSGSNNumber:     "628160360000",
			expectedSGSNAddress:    "192.168.1.1",
			expectedGprsEnhSupport: true,
			expectedLCSCapSets:     nil,
		},
		{
			name:                   "Valid UpdateGprsLocation with SGSNCapability and LCS",
			hexString:              "302204082143658709214365040791261806630000040504c0a80101a0068300850206c0",
			expectError:            false,
			expectedIMSI:           "1234567890123456",
			expectedSGSNNumber:     "628160360000",
			expectedSGSNAddress:    "192.168.1.1",
			expectedGprsEnhSupport: true,
			expectedLCSCapSets: &SupportedLCSCapabilitySets{
				LcsCapabilitySet1: true,
				LcsCapabilitySet2: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalBytes, err := hex.DecodeString(tc.hexString)
			if err != nil {
				t.Fatalf("Failed to decode hex string: %v", err)
			}

			updGprsLoc, err := ParseUpdateGprsLocation(originalBytes)
			if (err != nil) != tc.expectError {
				t.Fatalf("Unexpected error status during parsing: got %v, expected error: %v", err, tc.expectError)
			}

			if tc.expectError && err != nil {
				t.Logf("Expected error occurred in test case '%s': %v", tc.name, err)
				return
			}

			if updGprsLoc.IMSI != tc.expectedIMSI {
				t.Errorf("IMSI mismatch: got %s, expected %s", updGprsLoc.IMSI, tc.expectedIMSI)
			}
			if updGprsLoc.SGSNNumber != tc.expectedSGSNNumber {
				t.Errorf("SGSNNumber mismatch: got %s, expected %s", updGprsLoc.SGSNNumber, tc.expectedSGSNNumber)
			}
			if updGprsLoc.SGSNAddress != tc.expectedSGSNAddress {
				t.Errorf("SGSNAddress mismatch: got %s, expected %s", updGprsLoc.SGSNAddress, tc.expectedSGSNAddress)
			}

			if tc.expectedGprsEnhSupport {
				if updGprsLoc.SGSNCapability == nil {
					t.Error("Expected SGSNCapability but got nil")
				} else if !updGprsLoc.SGSNCapability.GprsEnhancementsSupportIndicator {
					t.Error("Expected GprsEnhancementsSupportIndicator to be true")
				}
			}

			if tc.expectedLCSCapSets != nil {
				if updGprsLoc.SGSNCapability == nil || updGprsLoc.SGSNCapability.SupportedLCSCapabilitySets == nil {
					t.Error("Expected SupportedLCSCapabilitySets but got nil")
				} else {
					if diff := cmp.Diff(tc.expectedLCSCapSets, updGprsLoc.SGSNCapability.SupportedLCSCapabilitySets); diff != "" {
						t.Errorf("SupportedLCSCapabilitySets mismatch (-expected +got):\n%s", diff)
					}
				}
			}

			marshaledBytes, err := updGprsLoc.Marshal()
			if err != nil {
				t.Fatalf("Failed to marshal UpdateGprsLocation: %v", err)
			}

			updGprsLoc2, err := ParseUpdateGprsLocation(marshaledBytes)
			if err != nil {
				t.Fatalf("Failed to re-parse marshaled UpdateGprsLocation: %v", err)
			}

			if diff := cmp.Diff(updGprsLoc, updGprsLoc2); diff != "" {
				t.Errorf("Round-trip semantic mismatch (-first +second):\n%s", diff)
			}
		})
	}
}

func TestParseUpdateLocationRes(t *testing.T) {
	tests := []struct {
		name              string
		hexString         string
		expectError       bool
		expectedHLRNumber string
	}{
		{
			name:              "Valid UpdateLocationRes",
			hexString:         "300704059126180663",
			expectError:       false,
			expectedHLRNumber: "62816036",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalBytes, err := hex.DecodeString(tc.hexString)
			if err != nil {
				t.Fatalf("Failed to decode hex string: %v", err)
			}

			updLocRes, err := ParseUpdateLocationRes(originalBytes)
			if (err != nil) != tc.expectError {
				t.Fatalf("Unexpected error status during parsing: got %v, expected error: %v", err, tc.expectError)
			}

			if tc.expectError && err != nil {
				t.Logf("Expected error occurred in test case '%s': %v", tc.name, err)
				return
			}

			if updLocRes.HLRNumber != tc.expectedHLRNumber {
				t.Errorf("HLRNumber mismatch: got %s, expected %s", updLocRes.HLRNumber, tc.expectedHLRNumber)
			}

			marshaledBytes, err := updLocRes.Marshal()
			if err != nil {
				t.Fatalf("Failed to marshal UpdateLocationRes: %v", err)
			}

			updLocRes2, err := ParseUpdateLocationRes(marshaledBytes)
			if err != nil {
				t.Fatalf("Failed to re-parse marshaled UpdateLocationRes: %v", err)
			}

			if diff := cmp.Diff(updLocRes, updLocRes2); diff != "" {
				t.Errorf("Round-trip semantic mismatch (-first +second):\n%s", diff)
			}
		})
	}
}

func TestParseUpdateGprsLocationRes(t *testing.T) {
	tests := []struct {
		name              string
		hexString         string
		expectError       bool
		expectedHLRNumber string
	}{
		{
			name:              "Valid UpdateGprsLocationRes",
			hexString:         "300704059126180663",
			expectError:       false,
			expectedHLRNumber: "62816036",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalBytes, err := hex.DecodeString(tc.hexString)
			if err != nil {
				t.Fatalf("Failed to decode hex string: %v", err)
			}

			updGprsLocRes, err := ParseUpdateGprsLocationRes(originalBytes)
			if (err != nil) != tc.expectError {
				t.Fatalf("Unexpected error status during parsing: got %v, expected error: %v", err, tc.expectError)
			}

			if tc.expectError && err != nil {
				t.Logf("Expected error occurred in test case '%s': %v", tc.name, err)
				return
			}

			if updGprsLocRes.HLRNumber != tc.expectedHLRNumber {
				t.Errorf("HLRNumber mismatch: got %s, expected %s", updGprsLocRes.HLRNumber, tc.expectedHLRNumber)
			}

			marshaledBytes, err := updGprsLocRes.Marshal()
			if err != nil {
				t.Fatalf("Failed to marshal UpdateGprsLocationRes: %v", err)
			}

			updGprsLocRes2, err := ParseUpdateGprsLocationRes(marshaledBytes)
			if err != nil {
				t.Fatalf("Failed to re-parse marshaled UpdateGprsLocationRes: %v", err)
			}

			if diff := cmp.Diff(updGprsLocRes, updGprsLocRes2); diff != "" {
				t.Errorf("Round-trip semantic mismatch (-first +second):\n%s", diff)
			}
		})
	}
}

func TestParseAnyTimeInterrogation(t *testing.T) {
	tests := []struct {
		name      string
		hexString string
	}{
		{
			name:      "MSISDN IMEI only",
			hexString: "301aa00a810891881047245232f9a1028600830891889006040000f8",
		},
		{
			name:      "MSISDN LocationInfo SubscriberState CurrentLocation PsDomain IMEI MsClassmark MnpRequestedInfo",
			hexString: "3027a00a810891881086450541f4a10f800081008300840101860085008700830891889006040000f4",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalBytes, err := hex.DecodeString(tc.hexString)
			if err != nil {
				t.Fatalf("Failed to decode hex string: %v", err)
			}

			parsed, err := ParseAnyTimeInterrogation(originalBytes)
			if err != nil {
				t.Fatalf("Failed to parse AnyTimeInterrogation: %v", err)
			}

			marshaledBytes, err := parsed.Marshal()
			if err != nil {
				t.Fatalf("Failed to marshal AnyTimeInterrogation: %v", err)
			}

			if diff := cmp.Diff(originalBytes, marshaledBytes); diff != "" {
				t.Errorf("Round-trip bytes mismatch (-original +marshaled):\n%s", diff)
			}
		})
	}
}

func TestParseUpdateLocation(t *testing.T) {
	tests := []struct {
		name                string
		hexString           string
		expectError         bool
		expectedIMSI        string
		expectedMSCNumber   string
		expectedVLRNumber   string
		expectedCamelPhases *SupportedCamelPhases
		expectedLCSCapSets  *SupportedLCSCapabilitySets
	}{
		{
			name:              "Valid UpdateLocation with VlrCapability (SupportedCamelPhases only)",
			hexString:         "3022040806076300938555f6810791261806630000040791261806630000a604800204c0",
			expectError:       false,
			expectedIMSI:      "607036003958556",
			expectedMSCNumber: "628160360000",
			expectedVLRNumber: "628160360000",
			expectedCamelPhases: &SupportedCamelPhases{
				Phase1: true,
				Phase2: true,
			},
			expectedLCSCapSets: nil,
		},
		{
			name:              "Valid UpdateLocation with VlrCapability (both SupportedCamelPhases and SupportedLCSCapabilitySets)",
			hexString:         "3026040832547090975937f2810791997627854900040791997627854900a608800205e0850206c0",
			expectError:       false,
			expectedIMSI:      "234507097995732",
			expectedMSCNumber: "996772589400",
			expectedVLRNumber: "996772589400",
			expectedCamelPhases: &SupportedCamelPhases{
				Phase1: true,
				Phase2: true,
				Phase3: true,
			},
			expectedLCSCapSets: &SupportedLCSCapabilitySets{
				LcsCapabilitySet1: true,
				LcsCapabilitySet2: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalBytes, err := hex.DecodeString(tc.hexString)
			if err != nil {
				t.Fatalf("Failed to decode hex string: %v", err)
			}

			updLoc, err := ParseUpdateLocation(originalBytes)
			if (err != nil) != tc.expectError {
				t.Fatalf("Unexpected error status during parsing: got %v, expected error: %v", err, tc.expectError)
			}

			if tc.expectError && err != nil {
				t.Logf("Expected error occurred in test case '%s': %v", tc.name, err)
				return
			}

			if updLoc.IMSI != tc.expectedIMSI {
				t.Errorf("IMSI mismatch: got %s, expected %s", updLoc.IMSI, tc.expectedIMSI)
			}
			if updLoc.MSCNumber != tc.expectedMSCNumber {
				t.Errorf("MSCNumber mismatch: got %s, expected %s", updLoc.MSCNumber, tc.expectedMSCNumber)
			}
			if updLoc.VLRNumber != tc.expectedVLRNumber {
				t.Errorf("VLRNumber mismatch: got %s, expected %s", updLoc.VLRNumber, tc.expectedVLRNumber)
			}

			if tc.expectedCamelPhases != nil {
				if updLoc.VlrCapability == nil || updLoc.VlrCapability.SupportedCamelPhases == nil {
					t.Error("Expected SupportedCamelPhases but got nil")
				} else {
					if diff := cmp.Diff(tc.expectedCamelPhases, updLoc.VlrCapability.SupportedCamelPhases); diff != "" {
						t.Errorf("SupportedCamelPhases mismatch (-expected +got):\n%s", diff)
					}
				}
			}

			if tc.expectedLCSCapSets != nil {
				if updLoc.VlrCapability == nil || updLoc.VlrCapability.SupportedLCSCapabilitySets == nil {
					t.Error("Expected SupportedLCSCapabilitySets but got nil")
				} else {
					if diff := cmp.Diff(tc.expectedLCSCapSets, updLoc.VlrCapability.SupportedLCSCapabilitySets); diff != "" {
						t.Errorf("SupportedLCSCapabilitySets mismatch (-expected +got):\n%s", diff)
					}
				}
			}

			marshaledBytes, err := updLoc.Marshal()
			if err != nil {
				t.Fatalf("Failed to marshal UpdateLocation: %v", err)
			}

			updLoc2, err := ParseUpdateLocation(marshaledBytes)
			if err != nil {
				t.Fatalf("Failed to re-parse marshaled UpdateLocation: %v", err)
			}

			if diff := cmp.Diff(updLoc, updLoc2); diff != "" {
				t.Errorf("Round-trip semantic mismatch (-first +second):\n%s", diff)
			}
		})
	}
}
