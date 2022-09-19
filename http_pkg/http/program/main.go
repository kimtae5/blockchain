package main

import (
	"http/blockpkg"
	"http/httppkg"
	"http/txpkg"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	GB := blockpkg.GenesisBlock()
	httppkg.PrevHash = GB.Hash
	httppkg.BlkChain = blockpkg.NewBlockchain()
	httppkg.BlkChain.AddBlock(GB)
	httppkg.Txs = txpkg.CreateTxDB()
	router := mux.NewRouter()
	// GenerateTransaction (트랜젝션 생성)
	router.HandleFunc("/Apply/Career", httppkg.ApplyCareer).Methods("Post")
	// GenerateBlock (블록생성)
	router.HandleFunc("/newBlk", httppkg.CreateNewBlock).Methods("Post")
	// FindAllTxByAddress ( 전체조회)
	router.HandleFunc("/refTx", httppkg.FindAllbyAddr).Methods("Post")
	// findTxbyTxid (상세조회)
	router.HandleFunc("/detailTx", httppkg.DetailTx).Methods("Post")

	http.ListenAndServe(":5000", router)
}
