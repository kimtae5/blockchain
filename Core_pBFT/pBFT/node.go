package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

type Node struct {
	NodeID           string
	NodeAddressTable map[string]string // key=nodeID, value=url
	View             *View
	CurrentState     *State
	CommittedMsgs    []*RequestMsg // kinda block.
	MsgBuffer        *MsgBuffer
	MsgEntrance      chan interface{}
	MsgDelivery      chan interface{}
	Alarm            chan bool
	NewNodeAlarm     chan []byte
	ReplyChan        chan [32]byte
}

type MsgBuffer struct {
	ReqMsgs        []*RequestMsg
	PrePrepareMsgs []*PrePrepareMsg
	PrepareMsgs    []*VoteMsg
	CommitMsgs     []*VoteMsg
}

type View struct {
	ID      int64
	Primary string
}
type Add struct {
	Ip   string `json:"address"`
	Port string `json:"node"`
}

var f int

var checkTable bool = true

const ResolvingTimeDuration = time.Millisecond * 1000 // 1 second.

func NewNode(nodeID string) *Node {

	var viewID int64 = 0 // temporary.
	ip := GetIP()

	node := &Node{
		// Hard-coded for test.
		NodeID: nodeID,
		NodeAddressTable: map[string]string{
			"10000":        "127.0.0.1:10000",
			string(nodeID): ip + ":" + nodeID,
		},
		View: &View{
			ID:      viewID,
			Primary: "10000",
		},

		// Consensus-related struct
		CurrentState:  nil,
		CommittedMsgs: make([]*RequestMsg, 0),
		MsgBuffer: &MsgBuffer{
			ReqMsgs:        make([]*RequestMsg, 0),
			PrePrepareMsgs: make([]*PrePrepareMsg, 0),
			PrepareMsgs:    make([]*VoteMsg, 0),
			CommitMsgs:     make([]*VoteMsg, 0),
		},

		// Channels
		MsgEntrance:  make(chan interface{}, 40000),
		MsgDelivery:  make(chan interface{}, 40000),
		NewNodeAlarm: make(chan []byte, 40000),
		Alarm:        make(chan bool),
		ReplyChan:    make(chan [32]byte, 40000),
	}

	// Start message dispatcher
	go node.dispatchMsg()

	// Start alarm trigger
	go node.alarmToDispatcher()

	// Start message resolver
	go node.resolveMsg()

	addNodeJson := &Add{Ip: ip, Port: nodeID}
	e, err := json.Marshal(addNodeJson)
	if err != nil {
		log.Println(err)
	}

	buff := bytes.NewBuffer(e)
	if nodeID != "10000" {
		http.Post("http://127.0.0.1:9999/newNode", "application/json", buff) //msp로 보냄
	}

	return node
}

//내부 아이피 가져오는 함수
func GetIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func (node *Node) Broadcast(msg interface{}, path string) map[string]error {
	errorMap := make(map[string]error)

	for nodeID, url := range node.NodeAddressTable {
		if nodeID == node.NodeID {
			continue
		}

		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			log.Println(err)
			errorMap[nodeID] = err
			continue
		}

		send(url+path, jsonMsg)
	}

	if len(errorMap) == 0 {
		return nil
	} else {
		return errorMap
	}
}

func (node *Node) Reply(msg *ReplyMsg) error {
	j++
	// Print all committed messages.
	for _, value := range node.CommittedMsgs {
		fmt.Printf("#Committed value# TxID : %x, TimeStamp : %s \n",
			value.TxID[:], string(value.TimeStamp))
		node.ReplyChan <- value.TxID

		fmt.Println("value.TxID : ", value.TxID)
	}
	fmt.Print("\n")

	// jsonMsg, err := json.Marshal(msg)
	// if err != nil {
	// 	log.Println(err)
	// 	return err
	// }
	//여기
	// Client가 없으므로, 일단 Primary에게 보내는 걸로 처리.
	// send(node.NodeAddressTable[node.View.Primary]+"/reply", jsonMsg)
	// send("192.168.10.99:5000/reply", jsonMsg)
	node.CurrentState = nil
	node.CommittedMsgs = make([]*RequestMsg, 0)
	return nil
}

// GetReq can be called when the node's CurrentState is nil.
// Consensus start procedure for the Primary.
func (node *Node) GetReq(reqMsg *RequestMsg) error {
	LogMsg(reqMsg)

	// Create a new state for the new
	err := node.createStateForNewConsensus()
	if err != nil {
		log.Println(err)
		return err
	}

	// Start the consensus process.
	prePrepareMsg, err := node.CurrentState.StartConsensus(reqMsg)
	if err != nil {
		log.Println(err)
		return err
	}

	LogStage(fmt.Sprintf("Consensus Process (ViewID:%d)", node.CurrentState.ViewID), false)

	// Send getPrePrepare message
	if prePrepareMsg != nil {
		node.Broadcast(prePrepareMsg, "/preprepare")
		LogStage("Pre-prepare", true)
	}

	return nil
}

// GetPrePrepare can be called when the node's CurrentState is nil.
// Consensus start procedure for normal participants.
func (node *Node) GetPrePrepare(prePrepareMsg *PrePrepareMsg) error {
	LogMsg(prePrepareMsg)

	// Create a new state for the new
	err := node.createStateForNewConsensus()
	if err != nil {
		log.Println(err)
		return err
	}

	prePareMsg, err := node.CurrentState.PrePrepare(prePrepareMsg)
	if err != nil {
		log.Println(err)
		return err
	}

	if prePareMsg != nil {
		// Attach node ID to the message
		prePareMsg.NodeID = node.NodeID

		LogStage("Pre-prepare", true)
		node.Broadcast(prePareMsg, "/prepare")
		LogStage("Prepare", false)
	}

	return nil
}

func (node *Node) GetPrepare(prepareMsg *VoteMsg) error {
	LogMsg(prepareMsg)

	commitMsg, err := node.CurrentState.Prepare(prepareMsg)
	if err != nil {
		log.Println(err)
		return err
	}

	if commitMsg != nil {
		// Attach node ID to the message
		commitMsg.NodeID = node.NodeID

		LogStage("Prepare", true)
		node.Broadcast(commitMsg, "/commit")
		LogStage("Commit", false)
	}

	return nil
}

func (node *Node) GetCommit(commitMsg *VoteMsg) error {
	LogMsg(commitMsg)

	replyMsg, committedMsg, err := node.CurrentState.Commit(commitMsg)
	if err != nil {
		log.Println(err)
		return err
	}

	if replyMsg != nil {
		if committedMsg == nil {

			return errors.New("committed message is nil, even though the reply message is not nil")
		}

		// Attach node ID to the message
		replyMsg.NodeID = node.NodeID

		// Save the last version of committed messages to node.
		node.CommittedMsgs = append(node.CommittedMsgs, committedMsg)

		LogStage("Commit", true)
		node.Reply(replyMsg)
		LogStage("Reply", true)

	}

	return nil
}

func (node *Node) GetReply(msg *ReplyMsg) {

	log.Printf("Result: %s by %s\n", msg.Result, msg.NodeID)

}

func (node *Node) createStateForNewConsensus() error {
	// Check if there is an ongoing consensus process.
	if node.CurrentState != nil {
		return errors.New("another consensus is ongoing")
	}

	// Get the last sequence ID
	var lastSequenceID int64
	if len(node.CommittedMsgs) == 0 {
		lastSequenceID = -1
	} else {
		lastSequenceID = node.CommittedMsgs[len(node.CommittedMsgs)-1].SequenceID
	}
	node.View.ID++
	fmt.Println("VIEW ID : ", node.View.ID)
	// Create a new state for this new consensus process in the Primary
	node.CurrentState = CreateState(node.View.ID, lastSequenceID)

	LogStage("Create the replica status", true)

	return nil
}

func (node *Node) dispatchMsg() {
	for {
		select {
		case msg := <-node.MsgEntrance:
			err := node.routeMsg(msg)
			if err != nil {
				log.Println(err)
				// TODO: send err to ErrorChannel
			}
		case <-node.Alarm:
			err := node.routeMsgWhenAlarmed()
			if err != nil {
				log.Println(err)
				// TODO: send err to ErrorChannel
			}

		}
	}
}

func (node *Node) routeMsg(msg interface{}) []error {
	switch msg.(type) {
	case *RequestMsg:
		if node.CurrentState == nil {
			// Copy buffered messages first.
			msgs := make([]*RequestMsg, len(node.MsgBuffer.ReqMsgs))
			copy(msgs, node.MsgBuffer.ReqMsgs)

			// Append a newly arrived message.
			msgs = append(msgs, msg.(*RequestMsg))

			// Empty the buffer.
			node.MsgBuffer.ReqMsgs = make([]*RequestMsg, 0)

			// Send messages.
			go func() {
				node.MsgDelivery <- msgs
			}()
		} else {
			log.Println("add Req to buffer")
			node.MsgBuffer.ReqMsgs = append(node.MsgBuffer.ReqMsgs, msg.(*RequestMsg))
		}
	case *PrePrepareMsg:
		if node.CurrentState == nil {
			// Copy buffered messages first.
			msgs := make([]*PrePrepareMsg, len(node.MsgBuffer.PrePrepareMsgs))
			copy(msgs, node.MsgBuffer.PrePrepareMsgs)

			// Append a newly arrived message.
			msgs = append(msgs, msg.(*PrePrepareMsg))

			// Empty the buffer.
			node.MsgBuffer.PrePrepareMsgs = make([]*PrePrepareMsg, 0)

			// Send messages.
			go func() {

				node.MsgDelivery <- msgs
			}()
		} else {
			log.Println("add PrePrePare to buffer")
			node.MsgBuffer.PrePrepareMsgs = append(node.MsgBuffer.PrePrepareMsgs, msg.(*PrePrepareMsg))
		}
	case *VoteMsg:
		if msg.(*VoteMsg).MsgType == PrepareMsg {

			if node.CurrentState == nil || node.CurrentState.CurrentStage != PrePrepared {
				log.Println("add PrePare to buffer")
				node.MsgBuffer.PrepareMsgs = append(node.MsgBuffer.PrepareMsgs, msg.(*VoteMsg))
			} else {
				// Copy buffered messages first.
				msgs := make([]*VoteMsg, len(node.MsgBuffer.PrepareMsgs))
				copy(msgs, node.MsgBuffer.PrepareMsgs)

				// Append a newly arrived message.
				msgs = append(msgs, msg.(*VoteMsg))

				// Empty the buffer.
				node.MsgBuffer.PrepareMsgs = make([]*VoteMsg, 0)

				// Send messages.
				go func() {

					node.MsgDelivery <- msgs
				}()
			}
		} else if msg.(*VoteMsg).MsgType == CommitMsg {
			if node.CurrentState == nil || node.CurrentState.CurrentStage != Prepared {
				log.Println("add Commit to buffer")
				node.MsgBuffer.CommitMsgs = append(node.MsgBuffer.CommitMsgs, msg.(*VoteMsg))
			} else {
				// Copy buffered messages first.
				msgs := make([]*VoteMsg, len(node.MsgBuffer.CommitMsgs))
				copy(msgs, node.MsgBuffer.CommitMsgs)

				// Append a newly arrived message.
				msgs = append(msgs, msg.(*VoteMsg))

				// Empty the buffer.
				node.MsgBuffer.CommitMsgs = make([]*VoteMsg, 0)

				// Send messages.
				go func() {

					node.MsgDelivery <- msgs
				}()
			}
		}
	}

	return nil
}

func (node *Node) routeMsgWhenAlarmed() []error {
	if node.CurrentState == nil {
		// Check ReqMsgs, send them.
		if len(node.MsgBuffer.ReqMsgs) != 0 {
			log.Println("get REQUEST from buffer")
			msgs := make([]*RequestMsg, len(node.MsgBuffer.ReqMsgs))
			copy(msgs, node.MsgBuffer.ReqMsgs)
			go func() {

				node.MsgDelivery <- msgs
			}()
		}

		// Check PrePrepareMsgs, send them.
		if len(node.MsgBuffer.PrePrepareMsgs) != 0 {
			log.Println("get PREPREPARE from buffer")
			msgs := make([]*PrePrepareMsg, len(node.MsgBuffer.PrePrepareMsgs))
			copy(msgs, node.MsgBuffer.PrePrepareMsgs)

			go func() {

				node.MsgDelivery <- msgs
			}()
		}
	} else {
		switch node.CurrentState.CurrentStage {
		case PrePrepared:
			// Check PrepareMsgs, send them.
			if len(node.MsgBuffer.PrepareMsgs) != 0 {
				log.Println("get PREPARE from buffer")
				msgs := make([]*VoteMsg, len(node.MsgBuffer.PrepareMsgs))
				copy(msgs, node.MsgBuffer.PrepareMsgs)

				go func() {

					node.MsgDelivery <- msgs
				}()
			}
		case Prepared:
			// Check CommitMsgs, send them.
			if len(node.MsgBuffer.CommitMsgs) != 0 {
				log.Println("get COMMIT from buffer")
				msgs := make([]*VoteMsg, len(node.MsgBuffer.CommitMsgs))
				copy(msgs, node.MsgBuffer.CommitMsgs)

				go func() {

					node.MsgDelivery <- msgs
				}()
			}
		}
	}

	return nil
}

func (node *Node) resolveMsg() {
	for {
		// Get buffered messages from the dispatcher.
		msgs := <-node.MsgDelivery
		switch msgs.(type) {
		case []*RequestMsg:
			errs := node.resolveRequestMsg(msgs.([]*RequestMsg))
			if len(errs) != 0 {
				for _, err := range errs {
					log.Println(err)
				}
				// TODO: send err to ErrorChannel
			}
		case []*PrePrepareMsg:
			errs := node.resolvePrePrepareMsg(msgs.([]*PrePrepareMsg))
			if len(errs) != 0 {
				for _, err := range errs {
					log.Println(err)
				}
				// TODO: send err to ErrorChannel
			}
		case []*VoteMsg:
			voteMsgs := msgs.([]*VoteMsg)
			if len(voteMsgs) == 0 {
				break
			}

			if voteMsgs[0].MsgType == PrepareMsg {
				errs := node.resolvePrepareMsg(voteMsgs)
				if len(errs) != 0 {
					for _, err := range errs {
						log.Println(err)
					}
					// TODO: send err to ErrorChannel
				}
			} else if voteMsgs[0].MsgType == CommitMsg {
				errs := node.resolveCommitMsg(voteMsgs)
				if len(errs) != 0 {
					for _, err := range errs {
						log.Println(err)
					}
					// TODO: send err to ErrorChannel
				}
			}
		}
	}
}

func (node *Node) alarmToDispatcher() {
	for {
		time.Sleep(ResolvingTimeDuration)
		node.Alarm <- true
	}
}

func (node *Node) resolveRequestMsg(msgs []*RequestMsg) []error {
	errs := make([]error, 0)

	// Resolve messages
	for _, reqMsg := range msgs {
		err := node.GetReq(reqMsg)
		if err != nil {
			log.Println(err)
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func (node *Node) resolvePrePrepareMsg(msgs []*PrePrepareMsg) []error {
	errs := make([]error, 0)

	// Resolve messages
	for _, prePrepareMsg := range msgs {
		err := node.GetPrePrepare(prePrepareMsg)
		if err != nil {
			log.Println(err)
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func (node *Node) resolvePrepareMsg(msgs []*VoteMsg) []error {
	errs := make([]error, 0)

	// Resolve messages
	for _, prepareMsg := range msgs {
		err := node.GetPrepare(prepareMsg)
		if err != nil {
			log.Println(err)
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func (node *Node) resolveCommitMsg(msgs []*VoteMsg) []error {
	errs := make([]error, 0)

	// Resolve messages
	for _, commitMsg := range msgs {
		err := node.GetCommit(commitMsg)
		if err != nil {
			log.Println(err)
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}
