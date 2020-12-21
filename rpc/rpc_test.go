package rpc

import (
	"context"
	"encoding/json"
	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
	"testing"
)

var token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.kdKU2k6TJ6Xjs6lCY4rPV0oZLaIuM-GbMrRgi4k7PlQ"
var rpcAddr = "http://127.0.0.1:1234/rpc/v0"

func TestClient_ChainHead(t *testing.T) {
	c := NewClient(rpcAddr, token)
	tpset, err := c.ChainHead(context.Background())
	if err != nil {
		t.Log(err)
		return
	}
	b, _ := json.Marshal(tpset)
	t.Log(string(b))
}

//Deprecated
func TestClient_ChainGetTipSetByHeight(t *testing.T) {
	c := NewClient(rpcAddr, token)
	ts, err := c.ChainGetTipSetByHeight(context.Background(), 161916, nil)
	if err != nil {
		t.Error(err)
	}
	for _, cid := range ts.Cids {
		blockMessages, err := c.ChainGetBlockMessages(context.Background(), cid)
		if err != nil {
			t.Error(err)
			return
		}
		for _, msgCid := range blockMessages.Cids {
			//state
			stateReplay, err := c.GetStateReplay(context.Background(), TipSetKey{}, msgCid)
			if err != nil {
				t.Error(err)
				return
			}
			if stateReplay.Msg.Method == 0 { //SEND
				if stateReplay.MsgRct.ExitCode == 0 { //OK
					t.Logf("OK SEND:hash[%v],from[%v],to[%v],value[%v],gas[%v],bcid[%v]", msgCid,
						stateReplay.Msg.From, stateReplay.Msg.To,
						stateReplay.Msg.Value, stateReplay.GasCost.TotalCost, cid)
				}
			}
		}
	}
	//注意bafy2bzaceaq73lj6qwnoeazl2ajy5bv2dcw2eoind5y5dq226edq2gx56posc出现在三个cid里面。
	/*
	   TestClient_ChainGetTipSetByHeight: rpc_test.go:45: OK SEND:hash[bafy2bzaceaq73lj6qwnoeazl2ajy5bv2dcw2eoind5y5dq226edq2gx56posc],from[t13sb4pa34qzf35txnan4fqjfkwwqgldz6ekh5trq],to[t1kxjjy3vizrg44swc5gmzwi4jh7inq4nuuswy47y],value[52622559635572906044],gas[439268],bcid[bafy2bzacecvadxxdzd5ils7kxcxbkqsg6l2g2ht33ft6mjlaecytub4bpzr5u]
	   TestClient_ChainGetTipSetByHeight: rpc_test.go:45: OK SEND:hash[bafy2bzaceaq73lj6qwnoeazl2ajy5bv2dcw2eoind5y5dq226edq2gx56posc],from[t13sb4pa34qzf35txnan4fqjfkwwqgldz6ekh5trq],to[t1kxjjy3vizrg44swc5gmzwi4jh7inq4nuuswy47y],value[52622559635572906044],gas[439268],bcid[bafy2bzaceanc2tbykfosbvbds3p3nq2ushomhz7m5m4yc54guer37sz7jokoq]
	   TestClient_ChainGetTipSetByHeight: rpc_test.go:45: OK SEND:hash[bafy2bzacecn2tocmeadtveh7ougm4ohnlk6h7xbp4umbhmdd3afg7zmebkp4m],from[t1susdjal2rogayvfr6m7wypfhawcmuc3ny7v7osi],to[t3sg22lqqjewwczqcs2cjr3zp6htctbovwugzzut2nkvb366wzn5tp2zkfvu5xrfqhreowiryxump7l5e6jaaq],value[200000000000000000000],gas[466268],bcid[bafy2bzaced7fadosawhuwjpjkbgxlu2qlsa2jiihsdfui5tjjpg2k6fcbjsfc]
	   TestClient_ChainGetTipSetByHeight: rpc_test.go:45: OK SEND:hash[bafy2bzaceaq73lj6qwnoeazl2ajy5bv2dcw2eoind5y5dq226edq2gx56posc],from[t13sb4pa34qzf35txnan4fqjfkwwqgldz6ekh5trq],to[t1kxjjy3vizrg44swc5gmzwi4jh7inq4nuuswy47y],value[52622559635572906044],gas[439268],bcid[bafy2bzaced7fadosawhuwjpjkbgxlu2qlsa2jiihsdfui5tjjpg2k6fcbjsfc]
	*/
}

func TestClient_ChainGetTipSetByHeightPro(t *testing.T) {
	c := NewClient(rpcAddr, token)
	ts, err := c.ChainGetTipSetByHeight(context.Background(), 161916, nil)
	if err != nil {
		t.Error(err)
	}
	for _, cid := range ts.Cids {
		blockMessages, err := c.ChainGetBlockMessages(context.Background(), cid)
		if err != nil {
			t.Error(err)
			return
		}
		for _, v := range blockMessages.BlsMessages {
			if v.Method == 0 { //SEND
				//checkState
				stateReplay, err := c.GetStateReplay(context.Background(), TipSetKey{}, v.CID)
				if err != nil {
					t.Error(err)
					return
				}
				if stateReplay.MsgCid != v.CID {
					t.Error("HOLY SHIT")
				}
				if stateReplay.MsgRct.ExitCode == 0 { //OK
					t.Logf("OK SEND in bls:hash[%v],from[%v],to[%v],value[%v],gas[%v],bcid[%v]", v.CID,
						stateReplay.Msg.From, stateReplay.Msg.To,
						stateReplay.Msg.Value, stateReplay.GasCost.TotalCost, cid)
				}
			}
		}
		for _, v := range blockMessages.SecpkMessages {
			if v.Message.Method == 0 {
				//checkState
				stateReplay, err := c.GetStateReplay(context.Background(), TipSetKey{}, v.CID)
				if err != nil {
					t.Error(err)
					return
				}
				if stateReplay.MsgCid != v.CID {
					t.Error("HOLY SHIT")
				}
				if stateReplay.MsgRct.ExitCode == 0 { //OK
					t.Logf("OK SEND in secpk:hash[%v],from[%v],to[%v],value[%v],gas[%v],bcid[%v]", v.CID,
						stateReplay.Msg.From, stateReplay.Msg.To,
						stateReplay.Msg.Value, stateReplay.GasCost.TotalCost, cid)
				}
			}
		}
	}
	//注意bafy2bzaceaq73lj6qwnoeazl2ajy5bv2dcw2eoind5y5dq226edq2gx56posc出现在三个cid里面。
	/*
		TestClient_ChainGetTipSetByHeightPro: rpc_test.go:96: OK SEND in secpk:hash[bafy2bzaceaq73lj6qwnoeazl2ajy5bv2dcw2eoind5y5dq226edq2gx56posc],from[t13sb4pa34qzf35txnan4fqjfkwwqgldz6ekh5trq],to[t1kxjjy3vizrg44swc5gmzwi4jh7inq4nuuswy47y],value[52622559635572906044],gas[439268],bcid[bafy2bzacecvadxxdzd5ils7kxcxbkqsg6l2g2ht33ft6mjlaecytub4bpzr5u]
		TestClient_ChainGetTipSetByHeightPro: rpc_test.go:96: OK SEND in secpk:hash[bafy2bzaceaq73lj6qwnoeazl2ajy5bv2dcw2eoind5y5dq226edq2gx56posc],from[t13sb4pa34qzf35txnan4fqjfkwwqgldz6ekh5trq],to[t1kxjjy3vizrg44swc5gmzwi4jh7inq4nuuswy47y],value[52622559635572906044],gas[439268],bcid[bafy2bzaceanc2tbykfosbvbds3p3nq2ushomhz7m5m4yc54guer37sz7jokoq]
		TestClient_ChainGetTipSetByHeightPro: rpc_test.go:96: OK SEND in secpk:hash[bafy2bzacecn2tocmeadtveh7ougm4ohnlk6h7xbp4umbhmdd3afg7zmebkp4m],from[t1susdjal2rogayvfr6m7wypfhawcmuc3ny7v7osi],to[t3sg22lqqjewwczqcs2cjr3zp6htctbovwugzzut2nkvb366wzn5tp2zkfvu5xrfqhreowiryxump7l5e6jaaq],value[200000000000000000000],gas[466268],bcid[bafy2bzaced7fadosawhuwjpjkbgxlu2qlsa2jiihsdfui5tjjpg2k6fcbjsfc]
		TestClient_ChainGetTipSetByHeightPro: rpc_test.go:96: OK SEND in secpk:hash[bafy2bzaceaq73lj6qwnoeazl2ajy5bv2dcw2eoind5y5dq226edq2gx56posc],from[t13sb4pa34qzf35txnan4fqjfkwwqgldz6ekh5trq],to[t1kxjjy3vizrg44swc5gmzwi4jh7inq4nuuswy47y],value[52622559635572906044],gas[439268],bcid[bafy2bzaced7fadosawhuwjpjkbgxlu2qlsa2jiihsdfui5tjjpg2k6fcbjsfc]
	*/
}

func TestBlockMessages(t *testing.T) {
	c := NewClient(rpcAddr, token)
	cids, _ := cid.Decode("bafy2bzacecvadxxdzd5ils7kxcxbkqsg6l2g2ht33ft6mjlaecytub4bpzr5u")
	msg, err := c.ChainGetBlockMessages(context.Background(), cids)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("len msg:", len(msg.BlsMessages))
	t.Log("len SecpkMessages:", len(msg.SecpkMessages))
	t.Log("len cids:", len(msg.Cids))
}

func TestClient_GetStateReplay(t *testing.T) {
	c := NewClient(rpcAddr, token)
	cids, _ := cid.Decode("bafy2bzacedc6b6bedcabeeo6kaad5kgem7fsmgzzzqrdknfsbcrn6an5hvvsm") //error
	resp, err := c.GetStateReplay(context.Background(), TipSetKey{}, cids)
	if err != nil {
		t.Log(err)
		//	return
	}
	respB, _ := json.Marshal(resp)
	t.Logf("%+v", string(respB))

	cids, _ = cid.Decode("bafy2bzacecgloawd525yw7n7oxyrard6arwb3nydhhv26lqrpclrgl2645hu6") //ok
	resp, err = c.GetStateReplay(context.Background(), TipSetKey{}, cids)
	if err != nil {
		t.Log(err)
		return
	}
	respB, _ = json.Marshal(resp)
	t.Logf("%+v", string(respB))
}

func TestClient_WalletBalance(t *testing.T) {
	c := NewClient(rpcAddr, token)
	addr, _ := address.NewFromString("f1tybjx2ri2khuugqpsc4f34vbvcb64gavhfgu6pq")
	d, err := c.WalletBalance(context.Background(), addr)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("b:", d)
}

func TestClient_StateGetActor(t *testing.T) {
	c := NewClient(rpcAddr, token)
	addr, _ := address.NewFromString("f1tybjx2ri2khuugqpsc4f34vbvcb64gavhfgu6pq")
	act, err := c.StateGetActor(context.Background(), addr, TipSetKey{})
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", act)
}
