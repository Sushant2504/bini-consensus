package main

import (
    "fmt"
    "math/rand"
    "time"
)

type NodeState int

const (
    Follower NodeState = iota
    Candidate
    Leader
)

type Node struct {
    ID          int
    State       NodeState
    CurrentTerm int
    VotedFor    int
    Peers       []int
    MsgCh       chan Message
    Channels    []chan Message
    Log         []string
    VoteCount   int
    Timeout     time.Duration
}

func NewNode(id int, peers []int, msgCh chan Message, allChans []chan Message) *Node {
    return &Node{
        ID:       id,
        State:    Follower,
        VotedFor: -1,
        Peers:    peers,
        MsgCh:    msgCh,
        Channels: allChans,
        Timeout:  time.Duration(1500+rand.Intn(1500)) * time.Millisecond,
    }
}

func (n *Node) broadcast(msg Message) {
    for _, peerID := range n.Peers {
        n.Channels[peerID] <- msg
    }
}

func (n *Node) resetElectionTimer() <-chan time.Time {
    return time.After(n.Timeout)
}

func (n *Node) Run() {
    timer := n.resetElectionTimer()
    for {
        select {
        case msg := <-n.MsgCh:
            n.handleMessage(msg)
        case <-timer:
            n.startElection()
            timer = n.resetElectionTimer()
        }
    }
}

func (n *Node) handleMessage(msg Message) {
    switch msg.Type {
    case RequestVote:
        if msg.Term > n.CurrentTerm && (n.VotedFor == -1 || n.VotedFor == msg.From) {
            n.VotedFor = msg.From
            n.CurrentTerm = msg.Term
            fmt.Printf("Node %d voted for Node %d (Term %d)\n", n.ID, msg.From, msg.Term)
            n.Channels[msg.From] <- Message{From: n.ID, To: msg.From, Type: Vote, Term: msg.Term}
        }
    case Vote:
        if n.State == Candidate && msg.Term == n.CurrentTerm {
            n.VoteCount++
            if n.VoteCount > len(n.Peers)/2 {
                n.becomeLeader()
            }
        }
    case AppendEntry:
        if msg.Term >= n.CurrentTerm {
            n.CurrentTerm = msg.Term
            n.State = Follower
            n.VotedFor = msg.From
            n.Log = append(n.Log, msg.LogEntry)
            fmt.Printf("Node %d appended log '%s' from Leader %d\n", n.ID, msg.LogEntry, msg.From)
        }
    case Heartbeat:
        if msg.Term >= n.CurrentTerm {
            n.State = Follower
            n.CurrentTerm = msg.Term
            n.VotedFor = msg.From
        }
    }
}

func (n *Node) startElection() {
    n.State = Candidate
    n.CurrentTerm++
    n.VotedFor = n.ID
    n.VoteCount = 1
    fmt.Printf("Node %d started election for Term %d\n", n.ID, n.CurrentTerm)

    voteReq := Message{From: n.ID, Type: RequestVote, Term: n.CurrentTerm}
    n.broadcast(voteReq)
}

func (n *Node) becomeLeader() {
    n.State = Leader
    fmt.Printf("Node %d became Leader for Term %d\n", n.ID, n.CurrentTerm)
    go n.sendHeartbeats()
    go n.appendEntries()
}

func (n *Node) sendHeartbeats() {
    for n.State == Leader {
        hb := Message{From: n.ID, Type: Heartbeat, Term: n.CurrentTerm}
        n.broadcast(hb)
        time.Sleep(500 * time.Millisecond)
    }
}

func (n *Node) appendEntries() {
    counter := 1
    for n.State == Leader {
        entry := fmt.Sprintf("cmd%d", counter)
        msg := Message{From: n.ID, Type: AppendEntry, Term: n.CurrentTerm, LogEntry: entry}
        n.broadcast(msg)
        n.Log = append(n.Log, entry)
        fmt.Printf("Leader %d broadcast log '%s'\n", n.ID, entry)
        counter++
        time.Sleep(2000 * time.Millisecond)
    }
}
