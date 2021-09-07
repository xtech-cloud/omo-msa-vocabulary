package proxy

import (
	"errors"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"omo.msa.vocabulary/proxy/graph"
)

var isNeo4j = true

func switchNode(info neo4j.Node) *Node {
	if info == nil {
		return nil
	}
	node := new(Node)
	props := info.Props()
	node.Name, _ = props["name"].(string)
	node.UID, _ = props["uid"].(string)
	node.Labels = info.Labels()
	node.ID = info.Id()
	return node
}

func switchLink(info neo4j.Relationship) *Link {
	if info == nil {
		return nil
	}
	link := new(Link)
	link.ID = info.Id()
	link.Label = info.Type()
	props := info.Props()
	link.Name = props["name"].(string)
	link.Relation = props["relation"].(string)
	link.Direction = uint8(props["direction"].(int64))
	link.From = info.StartId()
	link.To = info.EndId()
	return link
}

func CreateNode(name, label, uid string) (*Node, error) {
	if isNeo4j {
		node, err := graph.CreateNode(name, label, uid)
		return switchNode(node), err
	}
	return nil, errors.New("not support db")
}

func CreateLink(from, to int64, kind, name, relation string, direction uint8) (*Link, error) {
	if isNeo4j {
		link, err := graph.CreateLink(from, to, kind, name, direction, relation)
		return switchLink(link), err
	}
	return nil, errors.New("not support db")
}

func RemoveNode(id int64, label string) error {
	if isNeo4j {
		return graph.DeleteNode(id, label)
	}
	return errors.New("not support db")
}

func RemoveLink(id int64) error {
	if isNeo4j {
		return graph.DeleteLink(id)
	}
	return errors.New("not support db")
}

func GetNode(uid string) (*Node, error) {
	if isNeo4j {
		node, err := graph.GetNode(uid)
		return switchNode(node), err
	}
	return nil, errors.New("not support db")
}

func GetNodeByID(id int64) (*Node, error) {
	if isNeo4j {
		node, err := graph.GetNodeByID(id)
		return switchNode(node), err
	}
	return nil, errors.New("not support db")
}

func GetLink(from, to string) (*Link, error) {
	if isNeo4j {
		link, err := graph.GetLink(from, to)
		return switchLink(link), err
	}
	return nil, errors.New("not support db")
}

func GetLinkByID(id int64) (*Link, error) {
	if isNeo4j {
		link, err := graph.GetLinkByID(id)
		return switchLink(link), err
	}
	return nil, errors.New("not support db")
}

func FindPath(from, to string) (*Graph, error) {
	var tmp = new(Graph)
	tmp.construct()
	tmp.Center = from
	var err error
	if isNeo4j {
		nodes, links, err1 := graph.FindPath(from, to)
		for i := 0; i < len(nodes); i += 1 {
			tmp.AddNode(switchNode(nodes[i]))
		}
		for i := 0; i < len(links); i += 1 {
			tmp.AddLink(switchLink(links[i]))
		}
		err = err1
	}
	return tmp, err
}

func FindGraph(uid, label string) (*Graph, error) {
	var tmp = new(Graph)
	tmp.construct()
	tmp.Center = uid
	var err error
	if isNeo4j {
		nodes, links, err1 := graph.FindGraph(uid, label)
		for i := 0; i < len(nodes); i += 1 {
			tmp.AddNode(switchNode(nodes[i]))
		}
		for i := 0; i < len(links); i += 1 {
			tmp.AddLink(switchLink(links[i]))
		}
		err = err1
	}
	return tmp, err
}
