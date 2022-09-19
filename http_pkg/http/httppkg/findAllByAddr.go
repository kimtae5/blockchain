package httppkg

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type findReqBody struct {
	Address string `json:"Address"`
}

type resBody struct {
	TxID        [][32]byte `json:"TxID"`
	CareerStart []string   `json:"CareerStart"`
	CareerEnd   []string   `json:"CareerEnd"`
	Company     []string   `json:"Company"`
	Payment     []string   `json:"Payment"`
	Job         []string   `json:"Job"`
	Proof       []string   `json:"Proof"`
}

func FindAllbyAddr(w http.ResponseWriter, req *http.Request) {
	var body findReqBody
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&body)
	//에러 체크
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("Req Body Address: %s\n", body.Address)
	list := Txs.FindTxByAddr(body.Address, BlkChain)
	for _, v := range list {
		fmt.Printf("txID: %x\n", v.TxID)
	}

	res := &resBody{}
	for i := 0; i < len(list); i++ {
		res.TxID = append(res.TxID, list[i].TxID)
		res.CareerStart = append(res.CareerStart, string(list[i].CareerStart))
		res.CareerEnd = append(res.CareerEnd, string(list[i].CareerEnd))
		res.Company = append(res.Company, string(list[i].Company))
		res.Payment = append(res.Payment, string(list[i].Payment))
		res.Job = append(res.Job, string(list[i].Job))
		res.Proof = append(res.Proof, string(list[i].Proof))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
