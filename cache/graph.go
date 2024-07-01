package cache

import (
	"errors"
	"fmt"
	"omo.msa.vocabulary/config"
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
	"omo.msa.vocabulary/tool"
	"regexp"
)

const (
	GraphTypeFace     = "face"
	GraphTypeAsset    = "asset"
	GraphTypeEntity   = "entity"
	GraphTypeActivity = "activity"
	GraphTypeHonor    = "honor"
	GraphTypeEvent    = "event"
)

type GraphInfo struct {
	center string
	nodes  []*NodeInfo
	links  []*LinkInfo
}

func (mine *cacheContext) GetGraphByNode(node string) (*GraphInfo, error) {
	return mine.graph.GetGraphByCenter(node)
}

func (mine *cacheContext) GetGraphNode(uid string) *NodeInfo {
	return mine.graph.GetNode(uid)
}

func (mine *cacheContext) Graph() *GraphInfo {
	return mine.graph
}

func (mine *cacheContext) checkRelations(old, now *EntityInfo) {
	if now == nil {
		return
	}
	if old == nil {
		edges := mine.GetVEdgesByCenter(now.UID)
		for _, edge := range edges {
			relationKind := Context().GetRelation(edge.Relation)
			if relationKind != nil {
				Context().addSyncLink(now.UID, edge.Source, relationKind.UID, edge.Name, switchRelationToLink(relationKind.Kind), edge.Direction)
			}
		}
	} else {
		oldList := make([]string, 0, 10)
		oldEdges := mine.GetVEdgesByCenter(old.UID)
		for _, edge := range oldEdges {
			oldList = append(oldList, edge.Source)
		}
		newList := make([]string, 0, 10)
		newEdges := mine.GetVEdgesByCenter(now.UID)
		for _, relation := range newEdges {
			newList = append(newList, relation.Source)
		}
		for _, oldR := range oldEdges {
			if !tool.HasItem(newList, oldR.Source) {
				link := Context().graph.GetRelationBy(now.UID, oldR.Source)
				if link != nil {
					_ = Context().graph.RemoveLink(link.ID)
				}
			}
		}
		for _, nowR := range newEdges {
			if !tool.HasItem(oldList, nowR.Source) {
				relationKind := Context().GetRelation(nowR.Source)
				if relationKind != nil {
					Context().addSyncLink(now.UID, nowR.Source, relationKind.UID, nowR.Name, switchRelationToLink(relationKind.Kind), nowR.Direction)
				}
			}
		}
	}
}

func (mine *cacheContext) CreateLink(from, to *NodeInfo, name, relationUID string, direction DirectionType, weight uint32) (*LinkInfo, error) {
	if len(name) > 0 {
		pattern := `^[0-9]*$`
		reg := regexp.MustCompile(pattern)
		if reg.MatchString(name) {
			return nil, errors.New("the relation that all digit letter is baned")
		}
	}
	tmp := mine.GetRelation(relationUID)
	if tmp != nil {
		return mine.graph.CreateLink(from, to, switchRelationToLink(tmp.Kind), name, relationUID, direction, weight)
	} else {
		return nil, errors.New("not found the relation type by uid")
	}
}

func (mine *GraphInfo) construct() {
	mine.nodes = make([]*NodeInfo, 0, 100)
	mine.links = make([]*LinkInfo, 0, 100)
}

func (mine *GraphInfo) initInfo(db *proxy.Graph) {
	if db == nil {
		return
	}
	for i := 0; i < len(db.Nodes); i += 1 {
		node := new(NodeInfo)
		node.initInfo(db.Nodes[i])
		mine.nodes = append(mine.nodes, node)
	}
	for i := 0; i < len(db.Links); i += 1 {
		link := new(LinkInfo)
		link.initInfo(db.Links[i], mine.GetNodeByID(db.Links[i].From).Entity, mine.GetNodeByID(db.Links[i].To).Entity)
		mine.links = append(mine.links, link)
	}
}

func switchGraphNodeType(tp uint32) string {
	if tp == EventHonor {
		return GraphTypeHonor
	} else if tp == EventActivity {
		return GraphTypeActivity
	} else if tp == EventSpec {
		return GraphTypeActivity
	} else {
		return GraphTypeEntity
	}
}

func (mine *GraphInfo) GetGraphByCenter(entity string) (*GraphInfo, error) {
	en := Context().GetEntity(entity)
	if en == nil {
		return nil, errors.New("the entity is nil")
	}
	dbs, er := nosql.GetVEdgesByCenter(entity)
	if er != nil {
		return nil, er
	}
	var g = new(GraphInfo)
	g.construct()
	g.SetCenter(entity)
	_, _ = g.CreateNodeByEntity(en)
	g.checkVirtual()
	for _, db := range dbs {
		if db.Type > 0 {
			_ = g.CreateByVEdge2(fmt.Sprintf(g.center+"_virtual-%d", db.Type), switchGraphNodeType(db.Type), db)
		} else {
			_ = g.CreateByVEdge(switchGraphNodeType(db.Type), db)
		}
	}
	dbs2, err := nosql.GetVEdgesBySource(entity)
	if err == nil {
		for _, db := range dbs2 {
			if db.Type > 0 {
				_ = g.CreateByVEdge2(fmt.Sprintf(g.center+"_virtual-%d", db.Type), switchGraphNodeType(db.Type), db)
			} else {
				_ = g.CreateByVEdge(switchGraphNodeType(db.Type), db)
			}
		}
	}
	dbs3, err3 := nosql.GetEventsByType(entity, EventActivity)
	if err3 == nil {
		for _, db := range dbs3 {
			_ = g.CreateByEvent(g.center+"_virtual-1", GraphTypeActivity, db)
		}
	}
	dbs4, err4 := nosql.GetEventsByType(entity, EventHonor)
	if err4 == nil {
		for _, db := range dbs4 {
			_ = g.CreateByEvent(g.center+"_virtual-2", GraphTypeHonor, db)
		}
	}
	dbs5, err5 := nosql.GetEventsByType(entity, EventSpec)
	if err5 == nil {
		for _, db := range dbs5 {
			_ = g.CreateByEvent(g.center+"_virtual-4", GraphTypeEvent, db)
		}
	}

	//db, err := proxy.FindGraph(entity, switchEntityLabel(en.Concept))
	//if err == nil {
	//	for _, node := range db.Nodes {
	//		n := new(NodeInfo)
	//		n.initInfo(node)
	//		mine.AppendNode(n)
	//		g.AppendNode(n)
	//	}
	//	for _, link := range db.Links {
	//		l := new(LinkInfo)
	//		l.initInfo(link, g.GetNodeByID(link.From).Entity, g.GetNodeByID(link.To).Entity)
	//		mine.AppendEdge(l)
	//		g.AppendEdge(l)
	//	}
	//}
	return g, nil
}

func (mine *GraphInfo) checkVirtual() (string, string, string) {
	honor := mine.center + "_virtual-2"
	act := mine.center + "_virtual-1"
	spec := mine.center + "_virtual-4"
	_, _ = mine.CreateNode(0, config.Schema.Basic.GetName(EventHonor), honor, "", "", nil)
	_, _ = mine.CreateNode(0, config.Schema.Basic.GetName(EventActivity), act, "", "", nil)
	_, _ = mine.CreateNode(0, config.Schema.Basic.GetName(EventSpec), spec, "", "", nil)
	mine.CreateLinkBy(mine.center, honor, "", "", "", DirectionTypeDouble, 0)
	mine.CreateLinkBy(mine.center, act, "", "", "", DirectionTypeDouble, 0)
	mine.CreateLinkBy(mine.center, spec, "", "", "", DirectionTypeDouble, 0)
	return act, honor, spec
}

func (mine *GraphInfo) GetOwnerGraph(owner string) *GraphInfo {
	_, _, list := Context().GetEntitiesByOwner(owner, 0, 1)
	var g = new(GraphInfo)
	g.construct()
	for _, info := range list {
		db, err := proxy.FindGraph(info.UID, switchEntityLabel(info.Concept))
		if err == nil {
			for _, node := range db.Nodes {
				n := new(NodeInfo)
				n.initInfo(node)
				mine.AppendNode(n)
				g.AppendNode(n)
			}
			for _, link := range db.Links {
				l := new(LinkInfo)
				l.initInfo(link, g.GetNodeByID(link.From).Entity, g.GetNodeByID(link.To).Entity)
				mine.AppendEdge(l)
				g.AppendEdge(l)
			}
		}
	}
	return g
}

func (mine *GraphInfo) GetPath(from, to string) (*GraphInfo, error) {
	g := new(GraphInfo)
	g.construct()
	g.center = from
	db, err := proxy.FindPath(from, to)
	if err == nil {
		for _, node := range db.Nodes {
			n := new(NodeInfo)
			n.initInfo(node)
			mine.AppendNode(n)
			g.AppendNode(n)
		}
		for _, link := range db.Links {
			l := new(LinkInfo)
			l.initInfo(link, g.GetNodeByID(link.From).Entity, g.GetNodeByID(link.To).Entity)
			mine.AppendEdge(l)
			g.AppendEdge(l)
		}
		g.initInfo(db)
	}
	return g, err
}

func (mine *GraphInfo) SetCenter(node string) {
	mine.center = node
}

func (mine *GraphInfo) Center() string {
	return mine.center
}

func (mine *GraphInfo) Nodes() []*NodeInfo {
	return mine.nodes
}

func (mine *GraphInfo) Links() []*LinkInfo {
	return mine.links
}

func (mine *GraphInfo) GetNode(uid string) *NodeInfo {
	for i := 0; i < len(mine.nodes); i += 1 {
		if mine.nodes[i].Entity == uid {
			return mine.nodes[i]
		}
	}
	tmp, _ := proxy.GetNode(uid)
	if tmp != nil {
		node := new(NodeInfo)
		node.initInfo(tmp)
		mine.AppendNode(node)
		return node
	}
	return nil
}

func (mine *GraphInfo) GetNodeByID(id int64) *NodeInfo {
	for i := 0; i < len(mine.nodes); i += 1 {
		if mine.nodes[i].ID == id {
			return mine.nodes[i]
		}
	}
	tmp, _ := proxy.GetNodeByID(id)
	if tmp != nil {
		node := new(NodeInfo)
		node.initInfo(tmp)
		mine.AppendNode(node)
		return node
	}
	return nil
}

func (mine *GraphInfo) GetNodeByName(name string) *NodeInfo {
	for i := 0; i < len(mine.nodes); i += 1 {
		if mine.nodes[i].Name == name {
			return mine.nodes[i]
		}
	}
	return nil
}

func (mine *GraphInfo) GetRelationByEntity(node string) *LinkInfo {
	for i := 0; i < len(mine.links); i += 1 {
		if mine.links[i].HadNode(node) {
			return mine.links[i]
		}
	}
	return nil
}

func (mine *GraphInfo) GetRelation(id int64) *LinkInfo {
	for i := 0; i < len(mine.links); i += 1 {
		if mine.links[i].ID == id {
			return mine.links[i]
		}
	}
	return nil
}

func (mine *GraphInfo) HadRelation(from, to string, name string) bool {
	for i := 0; i < len(mine.links); i += 1 {
		if mine.links[i].Name == name && mine.links[i].HadAll(from, to) {
			return true
		}
	}
	link, _ := proxy.GetLink(from, to)
	if link != nil && link.Name == name {
		return true
	}
	return false
}

func (mine *GraphInfo) GetRelationBy(from, to string) *LinkInfo {
	for i := 0; i < len(mine.links); i += 1 {
		if mine.links[i].HadAll(from, to) {
			return mine.links[i]
		}
	}
	link, _ := proxy.GetLink(from, to)
	if link != nil {
		info := new(LinkInfo)
		info.initInfo(link, from, to)
		return info
	}
	return nil
}

func (mine *GraphInfo) HadLinkNode(uid string) bool {
	for i := 0; i < len(mine.links); i += 1 {
		if mine.links[i].HadNode(uid) {
			return true
		}
	}
	return false
}

func (mine *GraphInfo) HadLink(from, to string) bool {
	for i := 0; i < len(mine.links); i += 1 {
		if mine.links[i].From == from && mine.links[i].To == to {
			return true
		}
	}
	return false
}

func (mine *GraphInfo) HadNode(uid string) bool {
	for i := 0; i < len(mine.nodes); i += 1 {
		if mine.nodes[i].Entity == uid {
			return true
		}
	}
	return false
}

func (mine *GraphInfo) UpdateNodeCover(uid string, cover string) error {
	node := mine.GetNode(uid)
	if node == nil {
		return errors.New("not found the node in graph")
	}
	return node.UpdateCover(cover)
}

func (mine *GraphInfo) CreateNodeByEntity(entity *EntityInfo) (*NodeInfo, error) {
	if entity == nil {
		return nil, errors.New("the entity is nil")
	}
	var name = entity.Name
	//if entity.Add != "" {
	//	name = entity.Name + "-" + entity.Add
	//}
	return mine.CreateNode(int64(entity.ID), name, entity.UID, entity.Cover, GraphTypeEntity, entity.Tags)
}

func (mine *GraphInfo) CreateNode(id int64, name, entity, cover, tp string, tags []string) (*NodeInfo, error) {
	//t := mine.GetNodeByName(name)
	//if t != nil {
	//	return nil, errors.New("the node had existed")
	//}
	//node, err := proxy.CreateNode(name, switchEntityLabel(concept), entity)
	//if err != nil {
	//	return nil, err
	//}
	var info = new(NodeInfo)
	info.ID = id
	info.Name = name
	info.Cover = cover
	info.Labels = tags
	if info.Labels == nil {
		info.Labels = make([]string, 0, 1)
	}
	info.Type = tp
	info.Desc = ""
	info.Entity = entity
	mine.AppendNode(info)
	return info, nil
}

func (mine *GraphInfo) CreateLinkBy(from, to, name, kind, label string, direction DirectionType, wei uint32) {
	edge := new(LinkInfo)
	edge.From = from
	edge.To = to
	edge.Direction = direction
	edge.ID = 0
	edge.Name = name
	edge.Relation = kind
	edge.Label = label
	edge.Weight = wei
	mine.AppendEdge(edge)
}

func (mine *GraphInfo) CreateLink(from, to *NodeInfo, kind LinkType, name, relation string, direction DirectionType, weight uint32) (*LinkInfo, error) {
	if from == nil || to == nil {
		return nil, errors.New("from or to node is nil")
	}

	if mine.HadRelation(from.Entity, to.Entity, name) {
		return nil, errors.New("the link existed")
	}
	if kind == "" {
		kind = LinkTypeEmpty
	}
	link, err := proxy.CreateLink(from.ID, to.ID, string(kind), name, relation, uint8(direction), weight)
	if err != nil {
		return nil, err
	}
	var info = new(LinkInfo)
	info.initInfo(link, from.Entity, to.Entity)
	mine.AppendEdge(info)
	return info, nil
}

func (mine *GraphInfo) RemoveLink(id int64) error {
	err := proxy.RemoveLink(id)
	if err == nil {
		for i := 0; i < len(mine.links); i += 1 {
			if mine.links[i].ID == id {
				mine.links = append(mine.links[:i], mine.links[i:]...)
				break
			}
		}
	}
	return err
}

func (mine *GraphInfo) RemoveNode(id int64, label string) error {
	err := proxy.RemoveNode(id, label)
	if err == nil {
		for i := 0; i < len(mine.nodes); i += 1 {
			if mine.nodes[i].ID == id {
				mine.nodes = append(mine.nodes[:i], mine.nodes[i:]...)
				break
			}
		}
	}
	return err
}

func (mine *GraphInfo) AppendNode(node *NodeInfo) {
	if node == nil || mine.HadNode(node.Entity) {
		return
	}
	mine.nodes = append(mine.nodes, node)
}

func (mine *GraphInfo) AppendEdge(link *LinkInfo) {
	if link == nil || mine.HadLink(link.From, link.To) {
		return
	}
	mine.links = append(mine.links, link)
}

func (mine *GraphInfo) CreateByEvent(from, tp string, db *nosql.Event) error {
	if db == nil {
		return errors.New("the event is nil")
	}
	to := db.Quote
	if len(to) < 1 {
		return nil
	}
	_, _ = mine.CreateNode(0, "", to, "", tp, nil)

	edge := new(LinkInfo)
	edge.From = from
	edge.To = to
	edge.Direction = DirectionTypeFromTo
	edge.ID = int64(db.ID)
	edge.Name = db.Name
	edge.Relation = ""
	edge.Label = ""
	edge.Weight = 0
	mine.AppendEdge(edge)
	return nil
}

func (mine *GraphInfo) CreateByVEdge(tp string, db *nosql.VEdge) error {
	if mine.center != db.Source {
		from := cacheCtx.GetEntity(db.Source)
		if from != nil {
			_, er := mine.CreateNodeByEntity(from)
			if er != nil {
				return er
			}
			mine.CreateLinkBy(mine.center, from.UID, "", "", "", DirectionTypeFromTo, 0)
		}
	}
	to := db.Target.Entity
	if len(to) < 1 {
		to = db.Target.UID
	}
	entity := cacheCtx.GetEntity(db.Target.Entity)
	if entity != nil {
		_, er := mine.CreateNodeByEntity(entity)
		if er != nil {
			return er
		}
	} else {
		_, _ = mine.CreateNode(0, db.Target.Name, to, db.Target.Thumb, tp, nil)
	}

	edge := new(LinkInfo)
	edge.From = db.Source
	edge.To = to
	edge.Direction = DirectionType(db.Direction)
	edge.ID = int64(db.ID)
	edge.Name = db.Name
	edge.Relation = db.Catalog
	edge.Label = db.Remark
	edge.Weight = db.Weight
	mine.AppendEdge(edge)
	return nil
}

func (mine *GraphInfo) CreateByVEdge2(from, tp string, db *nosql.VEdge) error {
	to := db.Target.Entity
	if len(to) < 1 {
		to = db.Target.UID
	}
	entity := cacheCtx.GetEntity(db.Target.Entity)
	if entity != nil {
		_, er := mine.CreateNodeByEntity(entity)
		if er != nil {
			return er
		}
	} else {
		_, _ = mine.CreateNode(0, db.Target.Name, to, db.Target.Thumb, tp, nil)
	}

	edge := new(LinkInfo)
	edge.From = from
	edge.To = to
	edge.Direction = DirectionType(db.Direction)
	edge.ID = int64(db.ID)
	edge.Name = db.Name
	edge.Relation = db.Catalog
	edge.Label = db.Remark
	edge.Weight = db.Weight
	mine.AppendEdge(edge)
	return nil
}
