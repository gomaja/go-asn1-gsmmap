package address

// Extension indicator constants.
const (
	ExtensionNo = 0b10000000 // bit 8 set to 1, indicating no extension
)

// Nature of address indicator constants (bits 7, 6, 5).
const (
	NatureUnknown           = 0b000 << 4
	NatureInternational     = 0b001 << 4
	NatureNational          = 0b010 << 4
	NatureNetworkSpecific   = 0b011 << 4
	NatureSubscriber        = 0b100 << 4
	NatureReserved          = 0b101 << 4
	NatureAbbreviated       = 0b110 << 4
	NatureReservedExtension = 0b111 << 4
)

// Numbering plan indicator constants (bits 4, 3, 2, 1).
const (
	PlanUnknown           = 0b0000
	PlanISDN              = 0b0001
	PlanData              = 0b0011
	PlanTelex             = 0b0100
	PlanLandMobile        = 0b0110
	PlanNational          = 0b1000
	PlanPrivate           = 0b1001
	PlanReservedExtension = 0b1111
)

// Encode builds an AddressString octet sequence from its components.
func Encode(extension, natureOfAddress, numberingPlan uint8, digits []byte) []byte {
	firstOctet := extension | natureOfAddress | (numberingPlan & 0x0F)
	return append([]byte{firstOctet}, digits...)
}

// Decode splits an AddressString octet sequence into its components.
func Decode(encoded []byte) (extension, natureOfAddress, numberingPlan uint8, digits []byte) {
	if len(encoded) == 0 {
		return 0, 0, 0, nil
	}

	firstOctet := encoded[0]
	extension = firstOctet & 0b10000000
	natureOfAddress = firstOctet & 0b01110000
	numberingPlan = firstOctet & 0x0F

	if len(encoded) > 1 {
		digits = encoded[1:]
	} else {
		digits = []byte{}
	}

	return extension, natureOfAddress, numberingPlan, digits
}
