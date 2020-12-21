package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	block "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/xerrors"
	"io"
	"time"
)

type TipSetKey []cid.Cid

type Version struct {
	Version    string
	APIVersion uint32
	BlockDelay uint64
}

type BeaconEntry struct {
	Round uint64
	Data  []byte
}

type IpldObject struct {
	Cid cid.Cid
	Obj interface{}
}

type HeadChange struct {
	Type string
	Val  *TipSet
}

type ObjStat struct {
	Size  uint64
	Links uint64
}

type MessageSendSpec struct {
	MaxFee abi.TokenAmount `json:"MaxFee"`
}

type SignedMessage struct {
	Message   *Message          `json:"Message"`
	Signature *crypto.Signature `json:"Signature"`
	CID       cid.Cid           `json:"CID"`
}

type BlockMessages struct {
	BlsMessages   []*Message       `json:"BlsMessages"`
	SecpkMessages []*SignedMessage `json:"SecpkMessages"`
	Cids          []cid.Cid        `json:"Cids"`
}

type Actor struct {
	Code    cid.Cid         `json:"Code"`
	Head    cid.Cid         `json:"Head"`
	Nonce   uint64          `json:"Nonce"`
	Balance abi.TokenAmount `json:"Balance"`
}

type Ticket struct {
	VRFProof []byte
}

type BlockHeader struct {
	Miner                 string
	Ticket                *Ticket
	Parents               []cid.Cid
	ParentWeight          abi.TokenAmount
	Height                int64
	ParentStateRoot       cid.Cid
	ParentMessageReceipts cid.Cid
	Messages              cid.Cid
	BLSAggregate          *Signature
	Timestamp             uint64
	BlockSig              *Signature
	ForkSignaling         uint64
	ParentBaseFee         abi.TokenAmount
}

type TipSet struct {
	Cids   []cid.Cid
	Blocks []*BlockHeader
	Height int64
}

type MessageReceipt struct {
	// Ok = ExitCode(0)
	ExitCode int64 // 状态为0表示成功
	Return   []byte
	GasUsed  int64
}

type Loc struct {
	File     string
	Line     int
	Function string
}

type GasTrace struct {
	Name string

	Location          []Loc `json:"loc"`
	TotalGas          int64 `json:"tg"`
	ComputeGas        int64 `json:"cg"`
	StorageGas        int64 `json:"sg"`
	TotalVirtualGas   int64 `json:"vtg"`
	VirtualComputeGas int64 `json:"vcg"`
	VirtualStorageGas int64 `json:"vsg"`

	TimeTaken time.Duration `json:"tt"`
	Extra     interface{}   `json:"ex,omitempty"`
}

type ExecutionTrace struct {
	Msg        *Message
	MsgRct     *MessageReceipt
	Error      string
	Duration   time.Duration
	GasCharges []*GasTrace
	Subcalls   []ExecutionTrace
}

type StateReplay struct {
	MsgCid         cid.Cid         `json:"MsgCid"`
	Msg            *Message        `json:"Msg"`
	MsgRct         *MessageReceipt `json:"MsgRct"`
	GasCost        GasCost         `json:"GasCost"`
	ExecutionTrace ExecutionTrace  `json:"ExecutionTrace"`
	Error          string          `json:"Error"`
	Duration       time.Duration
}

type GasCost struct {
	TotalCost string `json:"TotalCost"`
}

type MsgLookup struct {
	Message   cid.Cid // Can be different than requested, in case it was replaced, but only gas values changed
	Receipt   MessageReceipt
	ReturnDec interface{}
	TipSet    TipSetKey
	Height    int64
}

type Message struct {
	Version    uint64          `json:"Version"`
	To         address.Address `json:"To"`
	From       address.Address `json:"From"`
	Nonce      uint64          `json:"Nonce"`
	Value      abi.TokenAmount `json:"Value"`
	GasLimit   int64           `json:"GasLimit"`
	GasFeeCap  abi.TokenAmount `json:"GasFeeCap"`
	GasPremium abi.TokenAmount `json:"GasPremium"`
	Method     uint64          `json:"Method"`
	Params     []byte          `json:"Params"`
	CID        cid.Cid         `json:"CID"`
}

func (m *Message) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.MarshalCBOR(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var lengthBufMessage = []byte{138}

func (t *Message) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufMessage); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Version (uint64) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Version)); err != nil {
		return err
	}

	// t.To (address.Address) (struct)
	if err := t.To.MarshalCBOR(w); err != nil {
		return err
	}

	// t.From (address.Address) (struct)
	if err := t.From.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Nonce (uint64) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Nonce)); err != nil {
		return err
	}

	// t.Value (big.Int) (struct)
	if err := t.Value.MarshalCBOR(w); err != nil {
		return err
	}

	// t.GasLimit (int64) (int64)
	if t.GasLimit >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.GasLimit)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.GasLimit-1)); err != nil {
			return err
		}
	}

	// t.GasFeeCap (big.Int) (struct)
	if err := t.GasFeeCap.MarshalCBOR(w); err != nil {
		return err
	}

	// t.GasPremium (big.Int) (struct)
	if err := t.GasPremium.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Method (abi.MethodNum) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Method)); err != nil {
		return err
	}

	// t.Params ([]uint8) (slice)
	if len(t.Params) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.Params was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.Params))); err != nil {
		return err
	}

	if _, err := w.Write(t.Params[:]); err != nil {
		return err
	}
	return nil
}

func (m *Message) ToStorageBlock() (block.Block, error) {
	data, err := m.Serialize()
	if err != nil {
		return nil, err
	}

	c, err := abi.CidBuilder.Sum(data)
	if err != nil {
		return nil, err
	}

	return block.NewBlockWithCid(data, c)
}

func (m *Message) Cid() cid.Cid {
	b, err := m.ToStorageBlock()
	if err != nil {
		panic(fmt.Sprintf("failed to marshal message: %s", err)) // I think this is maybe sketchy, what happens if we try to serialize a message with an undefined address in it?
	}

	return b.Cid()
}

////////////////////////

var (
	ErrKeyInfoNotFound = fmt.Errorf("key info not found")
	ErrKeyExists       = fmt.Errorf("key already exists")
)

// KeyType defines a type of a key
type KeyType string

func (kt *KeyType) UnmarshalJSON(bb []byte) error {
	{
		// first option, try unmarshaling as string
		var s string
		err := json.Unmarshal(bb, &s)
		if err == nil {
			*kt = KeyType(s)
			return nil
		}
	}

	{
		var b byte
		err := json.Unmarshal(bb, &b)
		if err != nil {
			return fmt.Errorf("could not unmarshal KeyType either as string nor integer: %w", err)
		}
		bst := crypto.SigType(b)

		switch bst {
		case crypto.SigTypeBLS:
			*kt = KTBLS
		case crypto.SigTypeSecp256k1:
			*kt = KTSecp256k1
		default:
			return fmt.Errorf("unknown sigtype: %d", bst)
		}
		return nil
	}
}

const (
	KTBLS             KeyType = "bls"
	KTSecp256k1       KeyType = "secp256k1"
	KTSecp256k1Ledger KeyType = "secp256k1-ledger"
)

// KeyInfo is used for storing keys in KeyStore
type KeyInfo struct {
	Type       KeyType
	PrivateKey []byte
}
