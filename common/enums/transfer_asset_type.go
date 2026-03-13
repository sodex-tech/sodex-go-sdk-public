package enums

type TransferAssetType int

const (
	TransferAssetTypeEVMDeposit TransferAssetType = iota
	TransferAssetTypePerpsDeposit
	TransferAssetTypeEVMWithdraw
	TransferAssetTypePerpsWithdraw
	TransferAssetTypeInternal
	TransferAssetTypeSpotWithdraw
	TransferAssetTypeSpotDeposit

	TransferAssetTypeUnknown TransferAssetType = -1
)

// String returns the string representation of TransferAssetType
func (t TransferAssetType) String() string {
	switch t {
	case TransferAssetTypeEVMDeposit:
		return "EVM_DEPOSIT"
	case TransferAssetTypePerpsDeposit:
		return "PERPS_DEPOSIT"
	case TransferAssetTypeEVMWithdraw:
		return "EVM_WITHDRAW"
	case TransferAssetTypePerpsWithdraw:
		return "PERPS_WITHDRAW"
	case TransferAssetTypeInternal:
		return "INTERNAL"
	case TransferAssetTypeSpotWithdraw:
		return "SPOT_WITHDRAW"
	case TransferAssetTypeSpotDeposit:
		return "SPOT_DEPOSIT"
	default:
		return "UNKNOWN"
	}
}

// ParseTransferAssetType parses a string into TransferAssetType
func ParseTransferAssetType(s string) TransferAssetType {
	switch s {
	case "EVM_DEPOSIT":
		return TransferAssetTypeEVMDeposit
	case "PERPS_DEPOSIT":
		return TransferAssetTypePerpsDeposit
	case "EVM_WITHDRAW":
		return TransferAssetTypeEVMWithdraw
	case "PERPS_WITHDRAW":
		return TransferAssetTypePerpsWithdraw
	case "INTERNAL":
		return TransferAssetTypeInternal
	case "SPOT_WITHDRAW":
		return TransferAssetTypeSpotWithdraw
	case "SPOT_DEPOSIT":
		return TransferAssetTypeSpotDeposit
	default:
		return TransferAssetTypeUnknown
	}
}
