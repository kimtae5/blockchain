package txpkg

import "sync"

type Txs struct {
	TxMap map[[32]byte]*Tx
}

var OnlyOnce sync.Once

// Txs(트랜잭션 DB 대용) 생성(최초 한번만 실행)
func CreateTxDB() *Txs {
	txs := &Txs{}
	txs.TxMap = make(map[[32]byte]*Tx)
	return txs
}

// Txs에 TX 저장
func (txs *Txs) AddTx(tx *Tx) [32]byte {
	txs.TxMap[tx.TxID] = tx
	return tx.TxID
}
