package enums

// SignatureType is a single-byte discriminator prepended to every wire-format
// signature. It allows the exchange server to select the correct verification
// algorithm without inspecting the signature body.
//
// The value is stored as the first byte of the 66-byte signature slice:
//
//	byte[0]    – SignatureType
//	byte[1:66] – 65-byte raw ECDSA signature (r ‖ s ‖ v)
type SignatureType int

const (
	SignatureTypeUnknown SignatureType = iota // Unrecognised or uninitialised type; always rejected by the server.
	SignatureTypeEIP712                       // EIP-712 structured-data signature using the engine-specific domain.
)

// String returns the canonical string representation of SignatureType as
// expected by the exchange API.
func (t SignatureType) String() string {
	switch t {
	case SignatureTypeEIP712:
		return "EIP712"
	default:
		return "UNKNOWN"
	}
}

// ParseSignatureType converts the canonical API string back to a SignatureType.
// Unrecognised values return SignatureTypeUnknown.
func ParseSignatureType(s string) SignatureType {
	switch s {
	case "EIP712":
		return SignatureTypeEIP712
	default:
		return SignatureTypeUnknown
	}
}
