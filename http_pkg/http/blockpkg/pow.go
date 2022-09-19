package blockpkg

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64
)

const targetBites = 6

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func IntToHex(obj int64) []byte {
	s := fmt.Sprint(obj)
	return []byte(s)
}

func (pow *ProofOfWork) PrepareData(nonce int) []byte {
	data := bytes.Join([][]byte{
		pow.block.PrevHash[:],
		pow.block.Txid[:],
		pow.block.Data,
		pow.block.Timestamp,
		IntToHex(int64(targetBites)),
		IntToHex(int64(nonce)),
	}, []byte{})
	return data
}
func (pow *ProofOfWork) Run() (int, [32]byte) {
	//fmt.Printf("PoW Run Start(targetBites=20): %s\n", time.Now().String())
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	for nonce < maxNonce {
		data := pow.PrepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	//fmt.Printf("PoW Run Finish: %s\n", time.Now().String())
	return nonce, hash
}

func NewProofOfWork(block *Block) *ProofOfWork {

	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBites))
	pow := &ProofOfWork{block, target}

	return pow
}
