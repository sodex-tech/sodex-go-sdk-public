package types

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	ErrInvalidSignatureLength = errors.New("invalid signature length")

	ErrInvalidSignatureType = errors.New("invalid signature type")

	ErrInvalidPublicKey = errors.New("invalid public key")
)

var (
	SpotDomainName = "spot"

	PerpsDomainName = "futures"

	UniversalDomainName = "universal"
)

type EIP712Domain struct {
	Name              string
	Version           string
	ChainID           *big.Int
	VerifyingContract common.Address

	separator *common.Hash
}

func NewEIP712Domain(name string, chainID uint64) EIP712Domain {
	return EIP712Domain{
		Name:              name,
		Version:           "1",
		ChainID:           big.NewInt(int64(chainID)),
		VerifyingContract: common.Address{},
	}
}

func DefaultSparkDomain() EIP712Domain {
	return NewEIP712Domain(SpotDomainName, 286623)
}

func DefaultBoltDomain() EIP712Domain {
	return NewEIP712Domain(PerpsDomainName, 286623)
}

func DefaultUniversalDomain() EIP712Domain {
	return NewEIP712Domain(UniversalDomainName, 286623)
}

func (d *EIP712Domain) DomainSeparator() common.Hash {
	if d.separator != nil {
		return *d.separator
	}

	hash := crypto.Keccak256Hash(
		crypto.Keccak256(
			[]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
		),
		crypto.Keccak256([]byte(d.Name)),
		crypto.Keccak256([]byte(d.Version)),
		math.U256Bytes(d.ChainID),
		common.LeftPadBytes(d.VerifyingContract.Bytes(), 32),
	)
	d.separator = &hash
	return hash
}

type ExchangeAction struct {
	PayloadHash common.Hash

	Nonce uint64
}

var ExchangeActionTypeHash = crypto.Keccak256Hash([]byte("ExchangeAction(bytes32 payloadHash,uint64 nonce)"))

func (ea *ExchangeAction) StructHash() common.Hash {
	nonceBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(nonceBytes[24:], ea.Nonce)

	return crypto.Keccak256Hash(
		ExchangeActionTypeHash.Bytes(),
		ea.PayloadHash.Bytes(),
		nonceBytes,
	)
}

func (ea *ExchangeAction) Hash(domain *EIP712Domain) common.Hash {
	domainSeparator := domain.DomainSeparator()
	structHash := ea.StructHash()

	return crypto.Keccak256Hash(
		[]byte{0x19, 0x01},
		domainSeparator.Bytes(),
		structHash.Bytes(),
	)
}

func RecoverExchangeActionSigner(payloadHash common.Hash, nonce uint64, domain *EIP712Domain, signature []byte) (common.Address, error) {
	if len(signature) != 65 {
		return common.Address{}, ErrInvalidSignatureLength
	}

	ea := &ExchangeAction{
		PayloadHash: payloadHash,
		Nonce:       nonce,
	}

	hash := ea.Hash(domain)

	pubKey, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}
	address := crypto.PubkeyToAddress(*pubKey)
	if address == (common.Address{}) {
		return common.Address{}, ErrInvalidPublicKey
	}
	return address, nil
}

type UserSignedAction struct {
	ChainID uint64 `json:"chainID"`
	Nonce   uint64 `json:"nonce"`
}

func publicKeyBytes(publicKey string) []byte {
	return common.FromHex(publicKey)
}

type UserSignedAddAPIKeyAction struct {
	UserSignedAction
	AddAPIKeyRequest
}

var UserSignedAddAPIKeyActionTypeHash = crypto.Keccak256Hash([]byte("UserSignedAddAPIKeyAction(uint64 chainID,uint64 nonce,uint64 accountID,string name,uint8 keyType,bytes publicKey,uint64 expiresAt)"))

func (a *UserSignedAddAPIKeyAction) StructHash() common.Hash {
	chainIDBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(chainIDBytes[24:], a.ChainID)
	nonceBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(nonceBytes[24:], a.Nonce)
	accountIDBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(accountIDBytes[24:], a.AccountID)
	keyTypeBytes := make([]byte, 32)
	keyTypeBytes[31] = uint8(a.Type)
	expiresAtBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(expiresAtBytes[24:], a.ExpiresAt)

	return crypto.Keccak256Hash(
		UserSignedAddAPIKeyActionTypeHash.Bytes(),
		chainIDBytes,
		nonceBytes,
		accountIDBytes,
		crypto.Keccak256([]byte(a.Name)),
		keyTypeBytes,
		crypto.Keccak256(a.PublicKey),
		expiresAtBytes,
	)
}

func (a *UserSignedAddAPIKeyAction) Hash(domain *EIP712Domain) common.Hash {
	return crypto.Keccak256Hash([]byte{0x19, 0x01}, domain.DomainSeparator().Bytes(), a.StructHash().Bytes())
}

type UserSignedAddPermissionedAPIKeyAction struct {
	UserSignedAction
	AddPermissionedAPIKeyRequest
}

var UserSignedAddPermissionedAPIKeyActionTypeHash = crypto.Keccak256Hash([]byte("UserSignedAddPermissionedAPIKeyAction(uint64 chainID,uint64 nonce,uint64 accountID,string name,uint8 keyType,bytes publicKey,uint64 expiresAt,uint64 permissions)"))

func (a *UserSignedAddPermissionedAPIKeyAction) StructHash() common.Hash {
	chainIDBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(chainIDBytes[24:], a.ChainID)
	nonceBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(nonceBytes[24:], a.Nonce)
	accountIDBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(accountIDBytes[24:], a.AccountID)
	keyTypeBytes := make([]byte, 32)
	keyTypeBytes[31] = uint8(a.Type)
	expiresAtBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(expiresAtBytes[24:], a.ExpiresAt)
	permissionsBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(permissionsBytes[24:], a.Permissions)

	return crypto.Keccak256Hash(
		UserSignedAddPermissionedAPIKeyActionTypeHash.Bytes(),
		chainIDBytes,
		nonceBytes,
		accountIDBytes,
		crypto.Keccak256([]byte(a.Name)),
		keyTypeBytes,
		crypto.Keccak256(publicKeyBytes(a.PublicKey)),
		expiresAtBytes,
		permissionsBytes,
	)
}

func (a *UserSignedAddPermissionedAPIKeyAction) Hash(domain *EIP712Domain) common.Hash {
	return crypto.Keccak256Hash([]byte{0x19, 0x01}, domain.DomainSeparator().Bytes(), a.StructHash().Bytes())
}

type AddAPIKey struct {
	AddAPIKeyRequest
	Nonce uint64
}

var AddAPIKeyTypeHash = crypto.Keccak256Hash([]byte("AddAPIKey(uint64 accountID,string name,uint8 keyType,bytes publicKey,uint64 expiresAt,uint64 nonce)"))

func (a *AddAPIKey) StructHash() common.Hash {
	accountIDBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(accountIDBytes[24:], a.AccountID)
	keyTypeBytes := make([]byte, 32)
	keyTypeBytes[31] = uint8(a.Type)
	expiresAtBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(expiresAtBytes[24:], a.ExpiresAt)
	nonceBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(nonceBytes[24:], a.Nonce)

	return crypto.Keccak256Hash(
		AddAPIKeyTypeHash.Bytes(),
		accountIDBytes,
		crypto.Keccak256([]byte(a.Name)),
		keyTypeBytes,
		crypto.Keccak256(a.PublicKey),
		expiresAtBytes,
		nonceBytes,
	)
}

func (a *AddAPIKey) Hash(domain *EIP712Domain) common.Hash {
	return crypto.Keccak256Hash([]byte{0x19, 0x01}, domain.DomainSeparator().Bytes(), a.StructHash().Bytes())
}

type AddAPIKeyWithBuilderAction struct {
	ChainID uint64
	Nonce   uint64
	AddAPIKeyWithBuilderRequest
}

var AddAPIKeyWithBuilderActionTypeHash = crypto.Keccak256Hash([]byte("AddAPIKeyWithBuilder(uint64 chainID,uint64 nonce,uint64 accountID,string name,uint8 keyType,bytes publicKey,uint64 expiresAt,uint64 builderID,uint64 maxFeeRate)"))

func (a *AddAPIKeyWithBuilderAction) StructHash() common.Hash {
	chainIDBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(chainIDBytes[24:], a.ChainID)
	nonceBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(nonceBytes[24:], a.Nonce)
	accountIDBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(accountIDBytes[24:], a.AccountID)
	keyTypeBytes := make([]byte, 32)
	keyTypeBytes[31] = uint8(a.Type)
	expiresAtBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(expiresAtBytes[24:], a.ExpiresAt)
	builderIDBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(builderIDBytes[24:], a.Builder.BuilderID)
	maxFeeRateBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(maxFeeRateBytes[24:], a.Builder.FeeRate)

	return crypto.Keccak256Hash(
		AddAPIKeyWithBuilderActionTypeHash.Bytes(),
		chainIDBytes,
		nonceBytes,
		accountIDBytes,
		crypto.Keccak256([]byte(a.Name)),
		keyTypeBytes,
		crypto.Keccak256(publicKeyBytes(a.PublicKey)),
		expiresAtBytes,
		builderIDBytes,
		maxFeeRateBytes,
	)
}

func (a *AddAPIKeyWithBuilderAction) Hash(domain *EIP712Domain) common.Hash {
	return crypto.Keccak256Hash([]byte{0x19, 0x01}, domain.DomainSeparator().Bytes(), a.StructHash().Bytes())
}

type ApproveBuilderFeeAction struct {
	ChainID    uint64
	Nonce      uint64
	AccountID  uint64
	BuilderID  uint64
	MaxFeeRate uint64
}

var ApproveBuilderFeeActionTypeHash = crypto.Keccak256Hash([]byte("ApproveBuilderFeeAction(uint64 chainID,uint64 nonce,uint64 accountID,uint64 builderID,uint64 maxFeeRate)"))

func (a *ApproveBuilderFeeAction) StructHash() common.Hash {
	chainIDBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(chainIDBytes[24:], a.ChainID)
	nonceBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(nonceBytes[24:], a.Nonce)
	accountIDBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(accountIDBytes[24:], a.AccountID)
	builderIDBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(builderIDBytes[24:], a.BuilderID)
	maxFeeRateBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(maxFeeRateBytes[24:], a.MaxFeeRate)

	return crypto.Keccak256Hash(
		ApproveBuilderFeeActionTypeHash.Bytes(),
		chainIDBytes,
		nonceBytes,
		accountIDBytes,
		builderIDBytes,
		maxFeeRateBytes,
	)
}

func (a *ApproveBuilderFeeAction) Hash(domain *EIP712Domain) common.Hash {
	return crypto.Keccak256Hash([]byte{0x19, 0x01}, domain.DomainSeparator().Bytes(), a.StructHash().Bytes())
}

func RecoverAddAPIKeySigner(action *AddAPIKey, domain *EIP712Domain, signature []byte) (common.Address, error) {
	if len(signature) != 65 {
		return common.Address{}, ErrInvalidSignatureLength
	}
	publicKey, err := crypto.SigToPub(action.Hash(domain).Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}
	address := crypto.PubkeyToAddress(*publicKey)
	if address == (common.Address{}) {
		return common.Address{}, ErrInvalidPublicKey
	}
	return address, nil
}
