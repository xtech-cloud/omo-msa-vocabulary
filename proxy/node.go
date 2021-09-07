package proxy

import (
	"errors"
	"omo.msa.vocabulary/proxy/graph"
)

type Node struct {
	ID     int64
	UID    string
	Name   string
	Labels []string
}

func (mine *Node) AddLabel(label string) (*Node, error) {
	if isNeo4j {
		node, err := graph.CreateNodeLabel(mine.ID, label)
		return switchNode(node), err
	}
	return nil, errors.New("not support db")
}

func (mine *Node) Delete() error {
	if isNeo4j {
		return graph.DeleteNode(mine.ID, mine.Labels[0])
	}
	return errors.New("not support db")
}
