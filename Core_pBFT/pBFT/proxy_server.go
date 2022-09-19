package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	url  string
	node *Node
}

func NewServer(nodeID string) *Server {
	node := NewNode(nodeID)
	server := &Server{node.NodeAddressTable[nodeID], node}

	server.setRoute()

	return server
}

func (server *Server) Start() {
	fmt.Printf("Server will be started at %s...\n", server.url)
	log.Printf("노드 실행 : %s \n", server.url)
	if err := http.ListenAndServe(server.url, nil); err != nil {
		log.Println(err)
		return
	}
}

func (server *Server) setRoute() {
	http.HandleFunc("/req", server.getReq) //Leader Node
	http.HandleFunc("/preprepare", server.getPrePrepare)
	http.HandleFunc("/prepare", server.getPrepare)
	http.HandleFunc("/commit", server.getCommit)
	http.HandleFunc("/getTable", server.getTable) //리더노드가 각각 노드에게 nodeaddresstable 줄 때 쓰는 경로 - 리더 제외 나머지가 받음
	http.HandleFunc("/getView", server.viewID)    //합의에 새로 참여한 노드에게 viewid를 전달해줌
	http.HandleFunc("/pingReq", server.pingReq)   //리더노드가 각 노드에게 정상작동하는지 확인하기 위해 보낸 신호를 받는 경로
	http.HandleFunc("/newNodeAlarm", server.newNodeAlarm)
}

func (server *Server) newNodeAlarm(writer http.ResponseWriter, request *http.Request) {
	checkTable = true
}

func (server *Server) pingReq(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(200)
}

func (server *Server) viewID(writer http.ResponseWriter, request *http.Request) {
	var newView map[string]int64
	json.NewDecoder(request.Body).Decode(&newView)
	server.node.View.ID = newView["view"]
	log.Println("View ID 수신 : ", server.node.View.ID)
}

func (server *Server) getTable(writer http.ResponseWriter, request *http.Request) {
	server.node.NodeAddressTable = make(map[string]string)
	json.NewDecoder(request.Body).Decode(&server.node.NodeAddressTable)
	log.Println("msp에게 NodeAddressTable 수신 : ", server.node.NodeAddressTable)
	f = (len(server.node.NodeAddressTable) - 1) / 3
}

func (server *Server) getReq(writer http.ResponseWriter, request *http.Request) {

	log.Println("허용 가능 악성 노드 : ", f)
	if checkTable {
		checkTable = false
		res, err := http.Post("http://127.0.0.1:9999/tableUpdateAlarm", "application/json",
			bytes.NewBuffer([]byte(`{ "view" : `+fmt.Sprint(server.node.View.ID)+` }`)))
		if err != nil {
			log.Println(err)
		}
		server.node.NodeAddressTable = make(map[string]string)
		json.NewDecoder(res.Body).Decode(&server.node.NodeAddressTable)
		log.Println("new table : ", server.node.NodeAddressTable)
		f = (len(server.node.NodeAddressTable) - 1) / 3
		res.Body.Close()
	}

	server.node.CurrentState = nil

	var msg RequestMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		log.Println(err)
		// writer.Write([]byte("false"))
		return
	}
	// writer.Write([]byte("true"))

	go func() {
		server.node.MsgEntrance <- &msg
	}()

	fmt.Println("server.node.ReplyChan1 : ", server.node.ReplyChan)
	select {
	case msg := <-server.node.ReplyChan:
		fmt.Println("server.node.ReplyChan2 : ", server.node.ReplyChan)

		fmt.Println("msg : ", msg)
		hi := ResultMsg{Txid: msg}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(hi)

	}

}

func (server *Server) getPrePrepare(writer http.ResponseWriter, request *http.Request) {
	server.node.CurrentState = nil

	var msg PrePrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		log.Println(err)
		return
	}
	go func() {
		server.node.MsgEntrance <- &msg
	}()
}

func (server *Server) getPrepare(writer http.ResponseWriter, request *http.Request) {

	var msg VoteMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		log.Println(err)
		return
	}
	go func() {

		server.node.MsgEntrance <- &msg
	}()
}

func (server *Server) getCommit(writer http.ResponseWriter, request *http.Request) {
	var msg VoteMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		log.Println(err)
		return
	}
	go func() {

		server.node.MsgEntrance <- &msg
	}()
}

// func (server *Server) getReply(writer http.ResponseWriter, request *http.Request) {
// 	var msg ReplyMsg
// 	err := json.NewDecoder(request.Body).Decode(&msg)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	server.node.GetReply(&msg)

// }

func send(url string, msg []byte) {
	buff := bytes.NewBuffer(msg)
	http.Post("http://"+url, "application/json", buff)
}
