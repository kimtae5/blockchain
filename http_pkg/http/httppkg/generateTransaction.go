package httppkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"http/txpkg"
	"net/http"
)

//Json 타입으로 리턴해주기 위한 구조체
type JsonResponse struct {
	Txid [32]byte `json:"txid"`
}

type ForSign struct {
	Address string
	Txid    []byte
}

type ResSing struct {
	SignValue []byte
}

var Txs *txpkg.Txs

// Generate Transaction
func ApplyCareer(w http.ResponseWriter, req *http.Request) {
	var body Request
	decoder := json.NewDecoder(req.Body)
	//decoder.DisallowUnknownFields()
	decoder.Decode(&body)
	//트랜잭션 생성
	T := txpkg.NewTx(body.Applier, body.Company, body.CareerStart, body.CareerEnd, body.Payment, body.Job, body.Proof, body.Address)
	//전자서명 생성
	signBody := ForSign{Address: body.Address, Txid: T.TxID[:]}
	jsonSign, _ := json.Marshal(signBody)
	SignRes, err := http.Post("http://127.0.0.1:3000/DigitalSignature", "application/json", bytes.NewBuffer(jsonSign))
	if err != nil {
		fmt.Println(err)
		return
	}
	var HashedTxid ResSing
	json.NewDecoder(SignRes.Body).Decode(&HashedTxid)
	fmt.Println("전자서명: ", HashedTxid.SignValue)
	T.Sign = HashedTxid.SignValue
	Txs.AddTx(T)
	T.PrintTx()
	fmt.Println("Tx-TxID: ", T.TxID)
	jsonForPBFT, _ := json.Marshal(T)
	PBFT_Res, err := http.Post("http://192.168.10.159:10000/req", "application/json", bytes.NewBuffer(jsonForPBFT))
	if err != nil {
		fmt.Println("합의 통신 실패 ", err)
	}

	fmt.Println("PBFT_Res : ", PBFT_Res)
	jsonRe := new(JsonResponse)
	json.NewDecoder(PBFT_Res.Body).Decode(&jsonRe)
	fmt.Println(jsonRe.Txid, "Middle에게 보낼 Txid입니다.")
	jsonResponse := JsonResponse{Txid: jsonRe.Txid}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonResponse)
}
