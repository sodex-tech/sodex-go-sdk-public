package enums

// APIKeyPermission is the bit index used by an API key permission mask.
// A set bit disables the corresponding permission.
type APIKeyPermission uint8

const (
	APIKeyPermissionTrade APIKeyPermission = iota
	APIKeyPermissionCancel
	APIKeyPermissionWithdraw
	APIKeyPermissionTransfer
)

func (p APIKeyPermission) Mask() uint64 {
	return 1 << p
}
