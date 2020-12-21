package wallet

import (
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/xinxuwang/filecoinwallet/rpc"
	"github.com/xinxuwang/filecoinwallet/sigs"
	_ "github.com/xinxuwang/filecoinwallet/sigs/secp"
)

func GenerateKey(typ rpc.KeyType) (*Key, error) {
	ctyp := ActSigType(typ)
	if ctyp == crypto.SigTypeUnknown {
		return nil, fmt.Errorf("unknown sig type: %s", typ)
	}
	pk, err := sigs.Generate(ctyp)
	if err != nil {
		return nil, err
	}
	ki := rpc.KeyInfo{
		Type:       typ,
		PrivateKey: pk,
	}
	return NewKey(ki)
}

type Key struct {
	rpc.KeyInfo

	PublicKey []byte
	Address   address.Address
}

func NewKey(keyinfo rpc.KeyInfo) (*Key, error) {
	k := &Key{
		KeyInfo: keyinfo,
	}

	var err error
	k.PublicKey, err = sigs.ToPublic(ActSigType(k.Type), k.PrivateKey)
	if err != nil {
		return nil, err
	}

	switch k.Type {
	case rpc.KTSecp256k1:
		k.Address, err = address.NewSecp256k1Address(k.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("converting Secp256k1 to address: %w", err)
		}
	case rpc.KTBLS:
		k.Address, err = address.NewBLSAddress(k.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("converting BLS to address: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported key type: %s", k.Type)
	}
	return k, nil

}

func ActSigType(typ rpc.KeyType) crypto.SigType {
	switch typ {
	case rpc.KTBLS:
		return crypto.SigTypeBLS
	case rpc.KTSecp256k1:
		return crypto.SigTypeSecp256k1
	default:
		return crypto.SigTypeUnknown
	}
}

func init() {
	address.CurrentNetwork = address.Mainnet
}
