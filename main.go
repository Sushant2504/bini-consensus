package main

import (
    "math/rand"
    "sync"
    "time"
)

const nodeCount = 4

func main() {
    rand.Seed(time.Now().UnixNano())

    nodes := make([]*Node, nodeCount)
    channels := make([]chan Message, nodeCount)

    for i := 0; i < nodeCount; i++ {
        channels[i] = make(chan Message, 100)
    }

    for i := 0; i < nodeCount; i++ {
        peerIDs := make([]int, 0)
        for j := 0; j < nodeCount; j++ {
            if i != j {
                peerIDs = append(peerIDs, j)
            }
        }
        nodes[i] = NewNode(i, peerIDs, channels[i], channels)
    }

    var wg sync.WaitGroup
    for _, node := range nodes {
        wg.Add(1)
        go func(n *Node) {
            defer wg.Done()
            n.Run()
        }(node)
    }

    wg.Wait()
}
