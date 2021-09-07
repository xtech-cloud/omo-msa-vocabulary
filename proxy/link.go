package proxy

import (
	"errors"
	"omo.msa.vocabulary/proxy/graph"
)

type Link struct {
	Direction uint8
	ID        int64
	Label     string
	Relation  string //关系分类UID
	Name      string
	From      int64
	To        int64
}

func (mine *Link) Delete() error {
	if isNeo4j {
		return graph.DeleteLink(mine.ID)
	}
	return errors.New("not support db")
}
