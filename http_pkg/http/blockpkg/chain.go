package blockpkg

import "reflect"

var ABC int

type Blocks struct {
	BlockChain map[[32]byte]*Block
}

func NewBlockchain() *Blocks {
	newBlocks := &Blocks{}
	newBlocks.BlockChain = make(map[[32]byte]*Block)
	return newBlocks
}

func (bs *Blocks) AddBlock(o *Block) {
	//height 계산
	currentHeight := len(bs.BlockChain) - 1
	//제네시스 블록이 아닐 경우에는 이전 블록의 아이디를 찾아온다.
	prev := [32]byte{}
	for _, value := range bs.BlockChain {
		if value.Height == currentHeight {
			prev = value.Hash
		}
	}
	o.PrevHash = prev
	//height 대입
	o.Height = currentHeight + 1

	bs.BlockChain[o.Hash] = o
}
func (bs *Blocks) GetBlock(blkID [32]byte) *Block {
	return bs.BlockChain[blkID]
}
func (bs *Blocks) FindHashByTx(txid [32]byte) [32]byte {
	for i, ii := range bs.BlockChain {
		if ii.Txid == txid {
			return i
		}
	}
	return [32]byte{}
}

func (bs *Blocks) FindBlock(height int) *Block {
	// 최신부터 돌려보자
	//최신 블록체인의 높이를 구한다
	current_height := len(bs.BlockChain) - 1
	if height == 0 {
		return nil
	}
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
		if blk.Height == height {
			return blk
		} else {
			if reflect.DeepEqual(blk.PrevHash, [32]byte{}) {
				return nil
			}
			curBlockID = blk.PrevHash
		}
	}
}
