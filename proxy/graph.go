package proxy

type Graph struct {
	Center string
	Nodes  []*Node
	Links  []*Link
}

func (mine *Graph) construct() {
	mine.Nodes = make([]*Node, 0, 20)
	mine.Links = make([]*Link, 0, 10)
}

func (mine *Graph) HadNode(id int64) bool {
	for i := 0; i < len(mine.Nodes); i += 1 {
		if mine.Nodes[i].ID == id {
			return true
		}
	}
	return false
}

func (mine *Graph) HadLink(id int64) bool {
	for i := 0; i < len(mine.Links); i += 1 {
		if mine.Links[i].ID == id {
			return true
		}
	}
	return false
}

func (mine *Graph) AddNode(info *Node) {
	if info == nil {
		return
	}
	if mine.Nodes == nil {
		mine.Nodes = make([]*Node, 0, 20)
	}
	if mine.HadNode(info.ID) {
		return
	}
	mine.Nodes = append(mine.Nodes, info)
}

func (mine *Graph) AddLink(info *Link) {
	if info == nil {
		return
	}
	if mine.Links == nil {
		mine.Links = make([]*Link, 0, 10)
	}
	if mine.HadLink(info.ID) {
		return
	}
	mine.Links = append(mine.Links, info)
}
