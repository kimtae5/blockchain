package txpkg

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"http/blockpkg"
	"reflect"
	"time"
)

type Tx struct {
	TxID        [32]byte `json:"TxID"`
	TimeStamp   []byte   `json:"Timestamp"`   // 블럭 생성 시간
	Applier     []byte   `json:"Applier"`     // 신청자
	Company     []byte   `json:"Company"`     // 경력회사
	CareerStart []byte   `json:"CareerStart"` // 경력기간
	CareerEnd   []byte   `json:"CareerEnd"`
	Payment     []byte   `json:"Payment"` // 결제수단
	Job         []byte   `json:"Job"`     // 직종, 업무
	Proof       []byte   `json:"Proof"`   // 경력증명서 pdf
	WAddr       []byte   `json:"Address"` // 지갑 주소
	Sign        []byte   `json:"Sign"`
}

//TX Hash 데이터 생성
func (tx *Tx) prepareData() []byte {
	data := bytes.Join([][]byte{
		tx.TimeStamp,
		tx.Payment,
		tx.Applier,
		tx.Company,
		tx.CareerStart,
		tx.CareerEnd,
		tx.Job,
		tx.Proof,
		tx.WAddr,
	}, []byte{})
	return data
}

//새로운 트랜잭션 생성
func NewTx(applier, company, careerStart, careerEnd, payment, job, proof, wAddr string) *Tx {
	newTx := &Tx{}

	newTx.Applier = []byte(applier)
	newTx.Company = []byte(company)
	newTx.CareerStart = []byte(careerStart)
	newTx.CareerEnd = []byte(careerEnd)
	newTx.Payment = []byte(payment)
	newTx.Job = []byte(job)
	newTx.Proof = []byte(proof)
	newTx.WAddr = []byte(wAddr)

	loc, _ := time.LoadLocation("Asia/Seoul")
	now := time.Now()
	t := now.In(loc)
	newTx.TimeStamp = []byte(t.String())

	data := newTx.prepareData()
	newTx.TxID = sha256.Sum256(data)

	return newTx
}

// 트랜잭션 ID를 이용해 Block 조회
func FindBlockByTx(txID [32]byte, bs *blockpkg.Blocks) *blockpkg.Block {
	// 최신부터 돌려보자
	//최신 블록체인의 높이를 구한다
	current_height := len(bs.BlockChain) - 1

	// 최신 블록ID를 찾는다
	curBlockID := [32]byte{}
	for _, v := range bs.BlockChain {
		if v.Height == current_height {
			curBlockID = v.Hash
			break
		}
	}

	for {
		blk := bs.BlockChain[curBlockID]
		if blk.IsExisted(txID) {
			return blk
		} else {
			if reflect.DeepEqual(blk.PrevHash, [32]byte{}) {
				return nil
			}
			curBlockID = blk.PrevHash
		}
	}
}

// 트랜잭션 ID를 이용해 트랜잭션 조회
func FindTxByTxid(txID [32]byte, txs *Txs) *Tx {
	return txs.TxMap[txID]
}

func (txs *Txs) FindTxByAddr(wAddr string, bs *blockpkg.Blocks) []*Tx {
	// 최신부터 돌려보자
	//최신 블록체인의 높이를 구한다
	current_height := len(bs.BlockChain) - 1
	// 최신 블록ID를 찾는다
	curBlockID := [32]byte{}
	for _, v := range bs.BlockChain {
		if v.Height == current_height {
			curBlockID = v.Hash
			break
		}
	}
	res := []*Tx{}
	for {
		blk := bs.BlockChain[curBlockID]
		if blk.Height != 0 {
			if string(txs.TxMap[blk.Txid].WAddr) == wAddr {
				res = append(res, txs.TxMap[blk.Txid])
			}
			curBlockID = blk.PrevHash
		} else {
			break
		}
	}
	return res
}

// 트랜잭션 정보 출력
func (t *Tx) PrintTx() {
	fmt.Println("==========Transaction Info=============")
	fmt.Printf("TxId: %x\n Applier: %s\n Company: %s\n CareerStart: %s\n CareerEnd: %s\n Payment: %s\n TimeStamp: %d\n Job: %s\n Proof: %s\nWallet Address: %s\n\n", t.TxID, t.Applier, t.Company, t.CareerStart, t.CareerEnd, t.Payment, t.TimeStamp, t.Job, t.Proof, t.WAddr)
}
