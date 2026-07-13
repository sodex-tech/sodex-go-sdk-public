package enums

type APIKeyType int

const (
	APIKeyTypeUnknown APIKeyType = iota
	APIKeyTypeEVM
)

func (t APIKeyType) String() string {
	if t == APIKeyTypeEVM {
		return "EVM"
	}
	return "UNKNOWN"
}

func ParseAPIKeyType(s string) APIKeyType {
	if s == "EVM" {
		return APIKeyTypeEVM
	}
	return APIKeyTypeUnknown
}
