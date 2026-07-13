package types

import "github.com/sodex-tech/sodex-go-sdk-public/common/enums"

const AddAPIKeyRequestTypeName = "addAPIKey"
const AddAPIKeyWithBuilderRequestTypeName = "addAPIKeyWithBuilder"
const AddPermissionedAPIKeyRequestTypeName = "addPermissionedAPIKey"

type AddAPIKeyRequest struct {
	AccountID uint64           `json:"accountID"`
	Name      string           `json:"name"`
	Type      enums.APIKeyType `json:"type"`
	PublicKey []byte           `json:"publicKey"`
	ExpiresAt uint64           `json:"expiresAt"`
}

func (req *AddAPIKeyRequest) ActionName() string {
	return AddAPIKeyRequestTypeName
}

type AddAPIKeyWithBuilderRequest struct {
	AccountID uint64           `json:"accountID"`
	Name      string           `json:"name"`
	Type      enums.APIKeyType `json:"type"`
	PublicKey string           `json:"publicKey"`
	ExpiresAt uint64           `json:"expiresAt"`
	Builder   BuilderParams    `json:"builder"`
}

func (req *AddAPIKeyWithBuilderRequest) ActionName() string {
	return AddAPIKeyWithBuilderRequestTypeName
}

type AddPermissionedAPIKeyRequest struct {
	AccountID   uint64           `json:"accountID"`
	Name        string           `json:"name"`
	Type        enums.APIKeyType `json:"type"`
	PublicKey   string           `json:"publicKey"`
	ExpiresAt   uint64           `json:"expiresAt"`
	Permissions uint64           `json:"permissions"`
}

func (req *AddPermissionedAPIKeyRequest) ActionName() string {
	return AddPermissionedAPIKeyRequestTypeName
}
