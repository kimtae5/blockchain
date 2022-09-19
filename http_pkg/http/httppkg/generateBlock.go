package httppkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"http/blockpkg"
	"http/txpkg"
	"net/http"
)

// Response 데이터를 담을 구조체
type FinalResponse struct {
	TxID [32]byte `json:"Txid"`
}

// Request 구조체
type Request struct {
	Address string `json:"Address"`
	Data    string `json:"Data"`
	//T       *txpkg.Tx `json:"transaction"`
	Applier     string `json:"Applier"`
	Company     string `json:"Company"`
	CareerStart string `json:"CareerStart"`
	CareerEnd   string `json:"CareerEnd"`
	Payment     string `json:"Payment"`
	Job         string `json:"Job"`
	Proof       string `json:"Proof"`
}

type Hi struct {
	Address   string
	Txid      []byte
	SignValue []byte
}
type Che struct {
	Check bool
}

var BlkChain *blockpkg.Blocks
var PrevHash [32]byte

func CreateNewBlock(w http.ResponseWriter, req *http.Request) {
	headerContentTtype := req.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		fmt.Println("content type 오류")
		return
	}

	tempBody := req.Body

	var body Request
	decoder := json.NewDecoder(tempBody)
	decoder.Decode(&body)

	bodyTwo, _ := json.Marshal(body)

	Res, err := http.Post("http://127.0.0.1:5000/Apply/Career", "application/json", bytes.NewBuffer(bodyTwo)) // middle에서 받은 데이터 GenTx한테 주고 Txid 받기
	if err != nil {
		fmt.Println(err)
		return
	}
	var txID FinalResponse
	json.NewDecoder(Res.Body).Decode(&txID)

	height := len(BlkChain.BlockChain)

	data := body.Data
	fmt.Println("Data: ", body.Address)

	//response용 구조체 생성
	// 블록 패키지에 구현해놓은 NewBlock() 실행후 해시값 저장
	t := txpkg.FindTxByTxid(txID.TxID, Txs)
	var bye Hi = Hi{body.Address, txID.TxID[:], t.Sign}
	fmt.Println(t.Sign)
	byebye, _ := json.Marshal(bye)
	hi, errr := http.Post("http://127.0.0.1:3000/VerifySign", "application/json", bytes.NewBuffer(byebye))
	if errr != nil {
		fmt.Println(errr)
	}
	Che := new(Che)
	json.NewDecoder(hi.Body).Decode(&Che)
	if !Che.Check {
		fmt.Println("서명 검증 실패 : ", *Che)
		return
	}
	b := blockpkg.NewBlock(PrevHash, height, txID.TxID, data)
	BlkChain.AddBlock(b)
	PrevHash = b.Hash
	b.PrintBlock()
	//Content Type을 JSON으로 설정
	w.Header().Set("Content-Type", "application/json")
	// response 구조체 JSON으로 인코딩후 전송
	json.NewEncoder(w).Encode(txID)

}
