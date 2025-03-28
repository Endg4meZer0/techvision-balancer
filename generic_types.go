package main

type Container struct {
	TaskID string
}

type Node struct {
	Hostname   string
	IsManager  bool
	Containers map[string]Container
}

type Nodes map[string]Node

// Returns true (working) if len(Containers) >= 1
func (n *Node) IsWorking() bool { return len(n.Containers) >= 1 }
