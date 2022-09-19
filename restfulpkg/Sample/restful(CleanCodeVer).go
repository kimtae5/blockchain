package main

// go version  go 1.18.4 window/amd64
//Restful API

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
)

type Args struct {
	Alias   string
	Address string
}
type Request struct {
	Alias   string
	Address string
	Txid    []byte
	Sign    []byte
}

type ReqVerify struct {
	Address   string
	Txid      []byte
	SignValue []byte
}

type Response struct {
	Address     string
	PublicKey   []byte
	PrivateKey  []byte
	Check       bool
	Txid        [][32]byte
	CareerStart []string
	CareerEnd   []string
	Company     []string
	Payment     []string
	Job         []string
	Proof       []string
	StringTxid  []string
	SignValue   []byte
}

type JsonDetailResponse struct { //경력 상세 조회 요청시 서비스로 돌려줄 구조체
	// Hash        [32]byte `json:"blockID"`
	// Data        string `json:"Data"`
	// Timestamp   string `json:"Timestamp"`
	// Txid        string `json:"StringTxid"` /////////////////////////////
	// Applier     string `json:"Applier"`
	Company     string `json:"Company"`
	CareerStart string `json:"CareerStart"`
	CareerEnd   string `json:"CareerEnd"`
	Job         string `json:"Job"`
	Proof       string `json:"Proof"`
	Sign        string `json:"Sign"`
	Hash        string `json:"Hash"`
}

func main() {
	r := &Request{}
	setRouter(r)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func setRouter(r *Request) {
	// if /mdware/MakeWallet으로 요청이 들어오면 r.ConnectWallet 실행
	http.HandleFunc("/MakeWallet", r.GenerateWallet)
	// if /mdware/CheckAddress으로 요청이 들어오면 r.ConnectTransaction 실행
	http.HandleFunc("/CheckAddress", r.CheckAddress)
	// if /mdware/RegisterCareer 요청이 들어오면
	http.HandleFunc("/RegisterCareer", r.RegisterCareer)
	// if /mdware/FindAllTxByAddress 요청이 들어오면
	http.HandleFunc("/GetWalletInfo", r.GetWallet)
	// if /mdware/digitalSignature
	http.HandleFunc("/DigitalSignature", r.DigitalSigniture)

	http.HandleFunc("/VerifySign", r.VerifySign)

	http.HandleFunc("/detailTx", r.findDetail)
}
func (r *Request) GenerateWallet(w http.ResponseWriter, re *http.Request) {
	//------------ Json 으로 들어온 Alias 확인 ( 서버에서 Send)
	headerContentType := re.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		// json 타입이 아니라면
		fmt.Println("Type is not application/json not applicable")
	}
	decoder := json.NewDecoder(re.Body)
	var request Request
	err := decoder.Decode(&request) // request Body에 들어있는 json 데이터를 해독하고 저장
	if err != nil {
		return
	}
	// -----------------------------------------JSON 해독 끝 -------------------------
	// ------------------------- RPC 서버 연결 ---------------------
	Client, err := rpc.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer Client.Close()
	response := new(Response) // 연결후 return을 받기 위해 빈 바구니 생성
	err = Client.Call("RpcServer.MakeNewWallet", request.Alias, response)
	if err != nil {
		fmt.Println(err, "Client.Call 에서 에러가 났음 ")
		return
	}
	// Wallet.go에서 받아온 데이터 요청한 서비스로 다시 돌려주기
	// 돌려주기 위해서 Json Parsing
	PrivateKey := hex.EncodeToString(response.PrivateKey)

	PublicKey := hex.EncodeToString(response.PublicKey)

	// fmt.Println(PrivateKey, "PrvateKey")
	// fmt.Println(PublicKey, "PublicKey")
	value := map[string]interface{}{
		"Alias":      request.Alias,
		"Address":    response.Address,
		"PublicKey":  PublicKey,
		"PrivateKey": PrivateKey,
	}
	// json_data, err := json.Marshal(value) // Parsing 완료
	// fmt.Println(json_data, "json 파싱한 후 데이터 ")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(value)
}

// 주소 검증 후 findAllTxByAddress 요청 보내기
func (r *Request) CheckAddress(w http.ResponseWriter, re *http.Request) {
	headerContentType := re.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		// json 타입이 아니라면
		fmt.Println("Json 타입이 아닙니다!!")
	}
	decoder := json.NewDecoder(re.Body)
	var request Request
	err := decoder.Decode(&request) // request Body에 들어있는 json 데이터를 해독하고 저장
	fmt.Println("request :", request)
	if err != nil {
		fmt.Println(err)
		return
	}

	Client, err := rpc.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer Client.Close()
	response := new(Response)
	fmt.Println("request.Address :", request.Address)
	err = Client.Call("RpcServer.CheckAddress", request.Address, response)
	fmt.Println(" response.check :", response.Check)
	if err != nil {
		fmt.Println(err)
		return
	}
	if response.Check {
		value := map[string]interface{}{
			"Address": request.Address,
		}
		json_data, _ := json.Marshal(value)
		fmt.Println("json_data : ", string(json_data))
		res, err := http.Post("http://127.0.0.1:5000/refTx", "application/json", bytes.NewBuffer(json_data))
		if err != nil {
			fmt.Println(err)
			return
		} // 받아온 Txs를 돌려줌
		response := new(Response)
		Decoder := json.NewDecoder(res.Body)
		Decoder.Decode(&response)
		for _, ii := range response.Txid {
			response.StringTxid = append(response.StringTxid, fmt.Sprintf("%x", ii[:]))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		fmt.Println("존재하지 않는 지갑주소입니다.")
	}
}

// 지갑 주소를 주고 그 주소에 해당하는 지갑을 받아오기
func (r *Request) GetWallet(w http.ResponseWriter, re *http.Request) {
	headerContentType := re.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		// json 타입이 아니라면
		fmt.Println("Json 타입이 아닙니다!!")
	}
	decoder := json.NewDecoder(re.Body)
	var request Request
	err := decoder.Decode(&request) // request Body에 들어있는 json 데이터를 해독하고 저장
	if err != nil {
		fmt.Println(err)
		return
	}

	Client, err := rpc.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer Client.Close()
	response := new(Response)
	err = Client.Call("RpcServer.GetWallet", request.Address, response)
	if err != nil {
		fmt.Println(err)
		return
	}
	value := map[string]string{
		"Address":    response.Address,
		"PublicKey":  string(response.PublicKey),
		"PrivateKey": string(response.PrivateKey),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(value)
}

// Digital Signature Function
func (r *Request) DigitalSigniture(w http.ResponseWriter, re *http.Request) {
	headerContentType := re.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		// json 타입이 아니라면
		fmt.Println("Json 타입이 아닙니다!!")
	}
	decoder := json.NewDecoder(re.Body)
	var request Request
	err := decoder.Decode(&request) // request Body에 들어있는 json 데이터를 해독하고 저장
	if err != nil {
		fmt.Println(err)
		return
	}
	Client, err := rpc.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer Client.Close()

	type tempStruct struct {
		Address string
		Txid    []byte
		// Sign    []byte
	}

	ts := tempStruct{Address: request.Address, Txid: request.Txid}
	response := new(Response)
	err = Client.Call("RpcServer.Signature", ts, response)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (r *Request) VerifySign(w http.ResponseWriter, re *http.Request) {
	headerContentType := re.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		// json 타입이 아니라면
		fmt.Println("Json 타입이 아닙니다!!")
	}
	decoder := json.NewDecoder(re.Body)
	var request ReqVerify
	err := decoder.Decode(&request) // request Body에 들어있는 json 데이터를 해독하고 저장

	type reqTemp struct {
		Address string
		Txid    []byte
		Sign    []byte
	}

	temp := reqTemp{Address: request.Address, Txid: request.Txid, Sign: request.SignValue}

	if err != nil {
		fmt.Println(err)
		return
	}
	Client, err := rpc.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer Client.Close()
	response := new(Response)
	err = Client.Call("RpcServer.VerifySign", temp, response)
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 경력 등록 요청을 받으면 Apply/Career로 보내기
func (r *Request) RegisterCareer(w http.ResponseWriter, re *http.Request) {
	Res, err := http.Post("http://127.0.0.1:5000/newBlk", "application/json", re.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Txid 만 돌려줌
	type txidStruct struct {
		Txid [32]byte `json:"Txid"`
	}
	response := new(txidStruct)
	Decoder := json.NewDecoder(Res.Body)
	Decoder.Decode(&response)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (r *Request) findDetail(w http.ResponseWriter, re *http.Request) {
	Res, err := http.Post("http://127.0.0.1:5000/detailTx", "application/json", re.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var detailStruct JsonDetailResponse
	json.NewDecoder(Res.Body).Decode(&detailStruct)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detailStruct)
}
