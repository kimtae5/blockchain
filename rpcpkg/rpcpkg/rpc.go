package main

// gRPC server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"net/rpc"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

var wallets = make(map[string]*Wallet)

type RpcServer struct{}

type Reply struct {
	Alias      string
	Address    string
	PublicKey  []byte
	PrivateKey []byte
	Check      bool
	SignValue  []byte
	Sign       []byte
}

// ------- Actual Responsable Functions -------------------------------

func (wRPC *RpcServer) MakeNewWallet(Alias string, reply *Reply) error {
	prvKey, pubKey := NewKeyPair()
	w := MakeWallet(&prvKey, pubKey, Alias)
	reply.Address = w.Address
	reply.PrivateKey = w.PrivateKey
	reply.PublicKey = w.PublicKey
	reply.Alias = w.Alias
	return nil
}
func (wRPC *RpcServer) CheckAddress(Address string, reply *Reply) error {
	// 주소가 존재한다면
	if wallets[Address] != nil {
		reply.Check = true
	} else {
		reply.Check = false
	}
	return nil
}

func (wRPC *RpcServer) GetWallet(Address string, reply *Reply) error {

	w := wallets[Address]
	reply.PrivateKey = w.PrivateKey
	reply.PublicKey = w.PublicKey
	return nil
}

func (wRCP *RpcServer) Signature(request *Request, reply *Reply) error {
	wallet := wallets[request.Address]
	if wallet != nil {
		SignValue, _ := ecdsa.SignASN1(rand.Reader, wallet.ecdsaPrviateKey, request.Txid)
		reply.SignValue = SignValue
	} else {
		fmt.Println("존재하지 않는 지갑주소입니다.")
	}
	return nil
}
func (wRPC *RpcServer) VerifySign(request *Request, reply *Reply) error {
	wallet := wallets[request.Address]
	if wallet == nil {
		return errors.New("no has wallet")
	}
	reply.Check = ecdsa.VerifyASN1(&wallet.ecdsaPrviateKey.PublicKey, request.Txid, request.Sign)
	return nil
}

// ----------------- End of Actual Responsable Functions ----------------

type Args struct {
	Alias   string
	Address string
}

type Wallet struct {
	PrivateKey      []byte
	PublicKey       []byte
	Address         string
	Alias           string
	ecdsaPrviateKey *ecdsa.PrivateKey
}

type Request struct {
	Txid    []byte
	Address string
	Sign    []byte
}

func MakeWallet(prvkey *ecdsa.PrivateKey, pubkey []byte, alias string) *Wallet {
	w := &Wallet{}
	publicRIPEMD160 := HashPubKey(pubkey)
	version := byte(0x00)
	Address := base58.CheckEncode(publicRIPEMD160, version)
	w.PrivateKey = prvkey.D.Bytes()
	w.ecdsaPrviateKey = prvkey
	w.PublicKey = pubkey
	w.Address = Address
	w.Alias = alias
	// walltes 에 방금 만들어진 wallet을 넣기
	wallets[w.Address] = w
	return w
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	prvKey, _ := ecdsa.GenerateKey(curve, rand.Reader)
	pubKey := prvKey.PublicKey
	bpubKey := append(pubKey.X.Bytes(), pubKey.Y.Bytes()...)

	return *prvKey, bpubKey
}

func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	RIPEMD160Hasher.Write(publicSHA256[:])

	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

func (Wallet *Wallet) PrintInfo() {
	fmt.Printf("Alias : %s\n", Wallet.Alias)
	fmt.Printf("Address : %s\n", Wallet.Address)
	fmt.Printf("PublicKey : %x\n", Wallet.PublicKey)
	fmt.Printf("PrivateKey : %s\n", Wallet.PrivateKey)
}

// -------------------- main ----------------------------------------------------

func main() {
	rpc.Register(new(RpcServer))
	In, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer In.Close()
	for {
		conn, err := In.Accept()
		if err != nil {
			continue
		}
		defer conn.Close()

		go rpc.ServeConn(conn)
	}
}
