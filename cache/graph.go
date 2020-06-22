package cache

import (
	"errors"
	"regexp"
	"omo.msa.vocabulary/proxy"
)

type GraphInfo struct {
	center string
	nodes  []*NodeInfo
	links  []*LinkInfo
}

func GetGraphByNode(node string) (*GraphInfo,error) {
	return cacheCtx.graph.GetSubGraph(node)
}

func GetGraphNode(uid string) *NodeInfo {
	return cacheCtx.graph.GetNode(uid)
}

func Graph() *GraphInfo {
	return cacheCtx.graph
}

func CreateLink(from, to *NodeInfo,kind LinkType, name string, relation string, direction DirectionType) (*LinkInfo, error) {
	if from == nil || to == nil {
		return nil, errors.New("the source node or target node is nil")
	}
	if len(name) > 0 {
		pattern := `^[0-9]*$`
		reg := regexp.MustCompile(pattern)
		if reg.MatchString(name){
			return nil, errors.New("the relation that all digit letter is baned")
		}
	}
	return cacheCtx.graph.CreateLink(from, to, kind, name,relation, direction)
}

func (mine *GraphInfo) construct()  {
	mine.nodes = make([]*NodeInfo, 0 ,100)
	mine.links = make([]*LinkInfo, 0 ,100)
}

func (mine *GraphInfo)initInfo(db *proxy.Graph)  {
	if db == nil {
		return
	}
	for i := 0;i < len(db.Nodes);i += 1 {
		node := new(NodeInfo)
		node.initInfo(db.Nodes[i])
		mine.nodes = append(mine.nodes, node)
	}
	for i := 0;i < len(db.Links);i += 1 {
		link := new(LinkInfo)
		link.initInfo(db.Links[i], mine.GetNodeByID(db.Links[i].From).EntityUID, mine.GetNodeByID(db.Links[i].To).EntityUID)
		mine.links = append(mine.links, link)
	}
}

func (mine *GraphInfo)GetSubGraph(entity string) (*GraphInfo,error) {
	center := mine.GetNode(entity)
	if center == nil {
		return nil,errors.New("the node is nil")
	}
	en := GetEntity(entity)
	if en == nil {
		return nil, errors.New("the entity is nil")
	}
	var g = new(GraphInfo)
	g.construct()
	g.SetCenter(center.EntityUID)
	db,err := proxy.FindGraph(entity, switchEntityName(en.Concept))
	if err == nil {
		for _, node := range db.Nodes {
			n := new(NodeInfo)
			n.initInfo(node)
			mine.AddNode(n)
			g.AddNode(n)
		}
		for _, link := range db.Links {
			l := new(LinkInfo)
			l.initInfo(link,g.GetNodeByID(link.From).EntityUID, g.GetNodeByID(link.To).EntityUID)
			mine.AddEdge(l)
			g.AddEdge(l)
		}
	}
	return g,err
}

func (mine *GraphInfo)GetPath(from ,to string) (*GraphInfo,error) {
	g := new(GraphInfo)
	g.construct()
	g.center = from
	db,err := proxy.FindPath(from, to)
	if err == nil {
		for _, node := range db.Nodes {
			n := new(NodeInfo)
			n.initInfo(node)
			mine.AddNode(n)
			g.AddNode(n)
		}
		for _, link := range db.Links {
			l := new(LinkInfo)
			l.initInfo(link,g.GetNodeByID(link.From).EntityUID, g.GetNodeByID(link.To).EntityUID)
			mine.AddEdge(l)
			g.AddEdge(l)
		}
		g.initInfo(db)
	}
	return g,err
}

func (mine *GraphInfo)SetCenter(node string) {
	mine.center = node
}

func (mine *GraphInfo)Center() string {
	return mine.center
}

func (mine *GraphInfo)Nodes() []*NodeInfo {
	return mine.nodes
}

func (mine *GraphInfo)Links() []*LinkInfo {
	return mine.links
}

func (mine *GraphInfo)GetNode(uid string) *NodeInfo {
	for i := 0;i < len(mine.nodes);i += 1 {
		if mine.nodes[i].EntityUID == uid {
			return mine.nodes[i]
		}
	}
	tmp,_ := proxy.GetNode(uid)
	if tmp != nil {
		node := new(NodeInfo)
		node.initInfo(tmp)
		mine.nodes = append(mine.nodes, node)
		return node
	}
	return nil
}

func (mine *GraphInfo)GetNodeByID(id int64) *NodeInfo {
	for i := 0;i < len(mine.nodes);i += 1 {
		if mine.nodes[i].ID == id {
			return mine.nodes[i]
		}
	}
	tmp,_ := proxy.GetNodeByID(id)
	if tmp != nil {
		node := new(NodeInfo)
		node.initInfo(tmp)
		mine.nodes = append(mine.nodes, node)
		return node
	}
	return nil
}

func (mine *GraphInfo)GetNodeByName(name string) *NodeInfo {
	for i := 0;i < len(mine.nodes);i += 1 {
		if mine.nodes[i].Name == name {
			return mine.nodes[i]
		}
	}
	return nil
}

func (mine *GraphInfo)GetRelation(node string) *LinkInfo {
	for i := 0;i < len(mine.links);i += 1 {
		if mine.links[i].HadNode(node) {
			return mine.links[i]
		}
	}
	return nil
}

func (mine *GraphInfo)Relation(id int64) *LinkInfo {
	for i := 0;i < len(mine.links);i += 1 {
		if mine.links[i].ID == id {
			return mine.links[i]
		}
	}
	return nil
}

func (mine *GraphInfo)HadRelation(from, to string, name string) bool {
	for i := 0;i < len(mine.links);i += 1 {
		if mine.links[i].Name == name && mine.links[i].HadAll(from, to){
			return true
		}
	}
	link,_ := proxy.GetLink(from, to)
	if link != nil && link.Name == name {
		return true
	}
	return false
}

func (mine *GraphInfo) HadLinkNode(uid string) bool {
	for i := 0;i < len(mine.links);i += 1 {
		if mine.links[i].HadNode(uid) {
			return true
		}
	}
	return false
}

func (mine *GraphInfo) HadLink(id int64) bool {
	for i := 0;i < len(mine.links);i += 1 {
		if mine.links[i].ID == id {
			return true
		}
	}
	return false
}

func (mine *GraphInfo)HadNode(uid string) bool {
	for i := 0;i < len(mine.nodes);i += 1 {
		if mine.nodes[i].EntityUID == uid {
			return true
		}
	}
	return false
}

func (mine *GraphInfo)UpdateNodeCover(uid string, cover string) error {
	node := mine.GetNode(uid)
	if node == nil {
		return errors.New("not found the node in graph")
	}
	return node.UpdateCover(cover)
}

func (mine *GraphInfo) CreateNodeByEntity(entity *EntityInfo) (*NodeInfo,error) {
	if entity == nil {
		return nil, errors.New("the entity is nil")
	}
	return mine.CreateNode(entity.Name, entity.UID, entity.Cover, entity.table())
}

func (mine *GraphInfo) CreateNode(name ,entity, cover, label string) (*NodeInfo,error) {
	if mine.GetNodeByName(name) != nil {
		return nil, errors.New("the node had existed")
	}
	node,err := proxy.CreateNode(name, label, entity)
	if err != nil {
		return nil, err
	}
	var info = new(NodeInfo)
	info.ID = node.ID
	info.Name = node.Name
	info.Cover = cover
	info.Labels = node.Labels
	info.EntityUID = entity
	mine.AddNode(info)
	return info, nil
}

func (mine *GraphInfo) CreateLink(from,to *NodeInfo,kind LinkType, name, relation string, direction DirectionType) (*LinkInfo, error) {
	if from == nil || to == nil {
		return nil, errors.New("from or to node is nil")
	}

	if mine.HadRelation(from.EntityUID, to.EntityUID, name) {
		return nil, errors.New("the link existed")
	}
	link,err := proxy.CreateLink(from.ID, to.ID, string(kind), name,relation, uint8(direction))
	if err != nil {
		return nil, err
	}
	var info = new(LinkInfo)
	info.initInfo(link, from.EntityUID, to.EntityUID)
	mine.AddEdge(info)
	return info,nil
}

func (mine *GraphInfo)RemoveLink(id int64) error {
	return proxy.RemoveLink(id)
}

func (mine *GraphInfo)RemoveNode(id int64, label string) error {
	return proxy.RemoveNode(id, label)
}

func (mine *GraphInfo)AddNode(node *NodeInfo)  {
	if node == nil || mine.HadNode(node.EntityUID){
		return
	}
	mine.nodes = append(mine.nodes, node)
}

func (mine *GraphInfo)AddEdge(link *LinkInfo)  {
	if link == nil || mine.HadLink(link.ID){
		return
	}
	mine.links = append(mine.links, link)
}


