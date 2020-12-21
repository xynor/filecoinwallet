package wallet

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/xinxuwang/filecoinwallet/rpc"
	"github.com/xinxuwang/filecoinwallet/sigs"
	_ "github.com/xinxuwang/filecoinwallet/sigs/secp"
	"github.com/xinxuwang/filecoinwallet/utils"
	"testing"
)

var token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.kdKU2k6TJ6Xjs6lCY4rPV0oZLaIuM-GbMrRgi4k7PlQ"
var rpcAddr = "http://127.0.0.1:1234/rpc/v0"

func TestNewWallet(t *testing.T) {
	key, err := GenerateKey(rpc.KTSecp256k1)
	if err != nil {
		t.Log(err)
		return
	}
	j, _ := json.Marshal(key)
	t.Logf("key:%v", string(j))
	//{"Type":"secp256k1","PrivateKey":"NNAgY3KZJ/RvMVfgJYgy4ABGXYZFq/bBWU9iZJR6v8k=","PublicKey":"BMdP4u/XfnUspVq8oH1IdhMU+kjbm5LABXrf76SnI8uhWe+BcnoggMKSjP7vMW3rlVvFutUXcFercRzV4amcyC8=","Address":"f1s5oakpdjs7bxmh2hq654zy64jptgbjqz4pcgc4q"}
}

func TestRestoreWallet(t *testing.T) {
	keyStr := `{"Type":"secp256k1","PrivateKey":"NNAgY3KZJ/RvMVfgJYgy4ABGXYZFq/bBWU9iZJR6v8k=","PublicKey":"BMdP4u/XfnUspVq8oH1IdhMU+kjbm5LABXrf76SnI8uhWe+BcnoggMKSjP7vMW3rlVvFutUXcFercRzV4amcyC8=","Address":"f1s5oakpdjs7bxmh2hq654zy64jptgbjqz4pcgc4q"}`
	var keyInfo = rpc.KeyInfo{}
	err := json.Unmarshal([]byte(keyStr), &keyInfo)
	if err != nil {
		t.Error(err)
		return
	}
	key, err := NewKey(keyInfo)
	if err != nil {
		t.Error(err)
		return
	}
	//compare
	keyJson, _ := json.Marshal(key)
	if string(keyJson) == keyStr {
		t.Log("BINGO")
	}
}

func TestSign(t *testing.T) {
	keyStr := `{"Type":"secp256k1","PrivateKey":"NNAgY3KZJ/RvMVfgJYgy4ABGXYZFq/bBWU9iZJR6v8k=","PublicKey":"BMdP4u/XfnUspVq8oH1IdhMU+kjbm5LABXrf76SnI8uhWe+BcnoggMKSjP7vMW3rlVvFutUXcFercRzV4amcyC8=","Address":"f1s5oakpdjs7bxmh2hq654zy64jptgbjqz4pcgc4q"}`
	var keyInfo = rpc.KeyInfo{}
	err := json.Unmarshal([]byte(keyStr), &keyInfo)
	if err != nil {
		t.Error(err)
		return
	}
	ki, err := NewKey(keyInfo)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Address:%s", ki.Address.String())
	//
	msg := []byte(`111111111111`)
	t.Log("msg hex:", hex.EncodeToString(msg))
	signed, err := sigs.Sign(ActSigType(ki.Type), ki.PrivateKey, msg)
	if err != nil {
		t.Error(err)
		return
	}
	sigBytes := append([]byte{byte(signed.Type)}, signed.Data...)
	t.Logf("signed:%+v", hex.EncodeToString(sigBytes))
}

func TestSendTransfer(t *testing.T) {
	c := rpc.NewClient(rpcAddr, token)

	keyStr := `{"Type":"secp256k1","PrivateKey":"NNAgY3KZJ/RvMVfgJYgy4ABGXYZFq/bBWU9iZJR6v8k=","PublicKey":"BMdP4u/XfnUspVq8oH1IdhMU+kjbm5LABXrf76SnI8uhWe+BcnoggMKSjP7vMW3rlVvFutUXcFercRzV4amcyC8=","Address":"f1s5oakpdjs7bxmh2hq654zy64jptgbjqz4pcgc4q"}`
	var keyInfo = rpc.KeyInfo{}
	err := json.Unmarshal([]byte(keyStr), &keyInfo)
	if err != nil {
		t.Error(err)
		return
	}
	ki, err := NewKey(keyInfo)
	if err != nil {
		t.Error(err)
		return
	}

	from, _ := address.NewFromString("f1s5oakpdjs7bxmh2hq654zy64jptgbjqz4pcgc4q")
	to, _ := address.NewFromString("f1fdwixtudzc3s7tn4hv45cpvcfw6u7txvb6e4hqa")
	valueBig, _ := utils.Str2Big("0.01", 18)
	msg := &rpc.Message{
		Version:    0,
		To:         to,
		From:       from,
		Nonce:      0,
		Value:      abi.TokenAmount{Int: valueBig},
		GasLimit:   0,
		GasFeeCap:  abi.TokenAmount{Int: big.Zero().Int},
		GasPremium: abi.TokenAmount{Int: big.Zero().Int},
		Method:     0,
		Params:     nil,
	}
	feeBig, _ := utils.Str2Big("0.00001", 18)
	maxFee := abi.TokenAmount{Int: feeBig}
	msg, err = c.GasEstimateMessageGas(context.Background(), msg, &rpc.MessageSendSpec{MaxFee: maxFee}, nil)
	if err != nil {
		t.Error(err)
	}
	actor, err := c.StateGetActor(context.Background(), msg.From, nil)
	if err != nil {
		t.Error(err)
	}

	msg.Nonce = actor.Nonce
	t.Logf("msg:%+v", msg)

	mb, err := msg.ToStorageBlock()
	if err != nil {
		t.Log(err)
		return
	}
	signed, err := sigs.Sign(ActSigType(ki.Type), ki.PrivateKey, mb.Cid().Bytes())
	if err != nil {
		t.Log(err)
		return
	}
	signedMessage := &rpc.SignedMessage{
		Message:   msg,
		Signature: signed,
		CID:       msg.CID,
	}
	sj, _ := json.Marshal(signedMessage)
	t.Logf("sigedMess:%+v", string(sj))
	cid, err := c.MpoolPush(context.Background(), signedMessage)
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("MpoolPush cid:%v", cid.String())
}

//nonce	低了会返回错误，MpoolPush。 wallet_test.go:220: jsonrpc call: map[code:1 message:minimum expected nonce is 1: message nonce too low]
//nonce 过高可以发送，放在pool池子中，等待nonce增长到该笔。
