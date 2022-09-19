package httppkg

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"http/txpkg"
	"net/http"
)

// Request 구조체
type DetailTxRequest struct {
	TxId string `json:"txID"`
}

type JsonDetailResponse struct {
	BeforeTxid  [32]byte `json:"BeforeTxid"`  //
	Company     string   `json:"Company"`     //
	CareerStart string   `json:"CareerStart"` //
	CareerEnd   string   `json:"CareerEnd"`   //
	Job         string   `json:"Job"`         //
	Proof       string   `json:"Proof"`       //
	Txid        string   `json:"Txid"`
	Payment     string   `json:"Payment"`
	Address     string   `json:"Address"`
	Sign        string   `json:"Sign"`
	Hash        string   `json:"Hash"`
}

func DetailTx(w http.ResponseWriter, req *http.Request) {
	var body JsonDetailResponse

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&body)
	if err != nil {
		fmt.Println(err)
		return
	}

	byteTxid, _ := hex.DecodeString(body.Txid)
	var Txid32 [32]byte
	copy(Txid32[:], byteTxid)
	t := txpkg.FindTxByTxid(Txid32, Txs)
	bHash := BlkChain.FindHashByTx(t.TxID)
	fmt.Println("Sign", t.Sign)

	if t != nil {
		var response = JsonDetailResponse{
			Company: string(t.Company), CareerStart: string(t.CareerStart), CareerEnd: string(t.CareerEnd),
			Job: string(t.Job), Proof: string(t.Proof), Txid: body.Txid, Hash: hex.EncodeToString(bHash[:]), Sign: hex.EncodeToString(t.Sign),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		fmt.Println("Txid 가 없습니다.")
	}

}
