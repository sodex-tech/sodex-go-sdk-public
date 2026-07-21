package types

const RevokeAPIKeyRequestTypeName = "revokeAPIKey"

type RevokeAPIKeyRequest struct {
	AccountID uint64 `json:"accountID"`
	Name      string `json:"name"`
}

func (req *RevokeAPIKeyRequest) ActionName() string {
	return RevokeAPIKeyRequestTypeName
}
