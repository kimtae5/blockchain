package main

// type RequestMsg struct {
// 	Timestamp  int64  `json:"timestamp"`
// 	ClientID   string `json:"clientID"`
// 	Operation  string `json:"operation"`
// 	SequenceID int64  `json:"sequenceID"`
// }

//재호님
// type Resp struct {
// 	Timestamp  int64    `json:"timestamp"`
// 	ViewID     int64    `json:"viewID"`
// 	ReplicaNum []string `json:"replicaNum"`
// 	ClientID   string   `json:"clientID"`
// }

// type RequestMsg struct {
// 	TxID        [32]byte `json:"TxID"`
// 	TimeStamp   []byte   `json:"TimeStamp"`   // 블럭 생성 시간
// 	Applier     []byte   `json:"Applier"`     // 신청자
// 	Company     []byte   `json:"Company"`     // 경력회사
// 	CareerStart []byte   `json:"careerStart"` // 경력기간
// 	CareerEnd   []byte   `json:"careerEnd"`
// 	Payment     []byte   `json:"Payment"` // 결제수단
// 	Job         []byte   `json:"Job"`     // 직종, 업무
// 	Proof       []byte   `json:"Proof"`   // 경력증명서 pdf
// 	WAddr       []byte   `json:"Address"` // 지갑 주소
// 	Sign        []byte   `json:"Sign"`
// 	SequenceID  int64    `json:"sequenceID"` //시퀀스아이디
// }

type RequestMsg struct {
	TxID       [32]byte `json:"TxID"`
	TimeStamp  []byte   `json:"TimeStamp"`
	SequenceID int64    `json:"sequenceID"` //시퀀스아이디
}
type ResultMsg struct {
	Txid [32]byte `json:"txid"`
}

type ReplyMsg struct {
	ViewID    int64    `json:"viewID"`
	Timestamp []byte   `json:"timestamp"`
	ClientID  [32]byte `json:"clientID"`
	NodeID    string   `json:"nodeID"`
	Result    string   `json:"result"`
}

type PrePrepareMsg struct {
	ViewID     int64       `json:"viewID"`
	SequenceID int64       `json:"sequenceID"`
	Digest     string      `json:"digest"`
	RequestMsg *RequestMsg `json:"requestMsg"`
}

type VoteMsg struct {
	ViewID     int64  `json:"viewID"`
	SequenceID int64  `json:"sequenceID"`
	Digest     string `json:"digest"`
	NodeID     string `json:"nodeID"`
	MsgType    `json:"msgType"`
}

type MsgType int

const (
	PrepareMsg MsgType = iota
	CommitMsg
)
