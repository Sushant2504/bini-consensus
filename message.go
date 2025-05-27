package main

type MessageType int

const (
    RequestVote MessageType = iota
    Vote
    AppendEntry
    Ack
    Heartbeat
)

type Message struct {
    From     int
    To       int
    Type     MessageType
    Term     int
    LogEntry string
}
