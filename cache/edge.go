package cache

import (
	"errors"
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

type VEdgeInfo struct {
	BaseInfo
	Center    string //中心实体或者根节点
	Direction uint8  // 方向
	Weight    uint32
	Source    string //实体对象或者临时UID
	Relation  string //关系类型
	Target    proxy.VNode
}

func (mine *cacheContext) GetVEdge(uid string) (*VEdgeInfo, error) {
	if uid == "" {
		return nil, errors.New("the uid is empty")
	}
	db, err := nosql.GetVEdge(uid)
	if err != nil {
		return nil, err
	}
	info := new(VEdgeInfo)
	info.initInfo(db)
	return info, nil
}

func (mine *cacheContext) RemoveVEdge(uid, operator string) error {
	if uid == "" {
		return errors.New("the uid is empty")
	}
	err := nosql.RemoveEdge(uid, operator)
	if err != nil {
		return err
	}
	return nil
}

func (mine *cacheContext) GetVEdgesBySource(entity string) []*VEdgeInfo {
	if entity == "" {
		return make([]*VEdgeInfo, 0, 1)
	}
	dbs, _ := nosql.GetVEdgesBySource(entity)
	list := make([]*VEdgeInfo, 0, len(dbs))
	for _, db := range dbs {
		info := new(VEdgeInfo)
		info.initInfo(db)
	}
	return list
}

func (mine *cacheContext) GetVEdgesByCenter(entity string) []*VEdgeInfo {
	if entity == "" {
		return make([]*VEdgeInfo, 0, 1)
	}
	dbs, _ := nosql.GetVEdgesByCenter(entity)
	list := make([]*VEdgeInfo, 0, len(dbs))
	for _, db := range dbs {
		info := new(VEdgeInfo)
		info.initInfo(db)
		list = append(list, info)
	}
	return list
}

func (mine *VEdgeInfo) initInfo(db *nosql.VEdge) {
	mine.UID = db.UID.Hex()
	mine.ID = db.ID
	mine.CreateTime = db.CreatedTime
	mine.UpdateTime = db.UpdatedTime
	mine.Operator = db.Operator
	mine.Creator = db.Creator
	mine.Name = db.Name
	mine.Source = db.Source
	mine.Center = db.Center
	mine.Target = db.Target
	mine.Weight = db.Weight
	mine.Direction = db.Direction
	mine.Relation = db.Catalog
}

func (mine *VEdgeInfo) UpdateBase(name, relation, operator string, dire uint8, target proxy.VNode) error {
	//if name == "" {
	//	name = mine.Name
	//}
	if relation == "" {
		relation = mine.Relation
	}

	err := nosql.UpdateVEdgeBase(mine.UID, name, relation, operator, dire, target)
	if err == nil {
		mine.Name = name
		mine.Relation = relation
		mine.Direction = dire
		mine.Target = target
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}
