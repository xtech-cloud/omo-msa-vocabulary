package cache

import (
	"errors"
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

type VEdgeInfo struct {
	BaseInfo
	Type      uint8
	Remark    string      `json:"remark"`
	Center    string      `json:"center"`    //中心实体或者根节点
	Direction uint8       `json:"direction"` // 方向
	Weight    uint32      `json:"weight"`
	Source    string      `json:"source"`   //from实体对象或者临时UID
	Relation  string      `json:"relation"` //关系类型或者名称
	Target    proxy.VNode `json:"target"`   //目标对象to
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
	mine.Created = db.Created
	mine.Updated = db.Updated
	mine.Operator = db.Operator
	mine.Creator = db.Creator
	mine.Name = db.Name
	mine.Source = db.Source
	mine.Center = db.Center
	mine.Target = db.Target
	mine.Weight = db.Weight
	mine.Type = uint8(db.Type)
	mine.Direction = db.Direction
	mine.Relation = db.Catalog
	mine.Remark = db.Remark
}

func (mine *VEdgeInfo) UpdateBase(name, remark, relation, operator string, dire uint8, target proxy.VNode) error {
	//if name == "" {
	//	name = mine.Name
	//}
	if relation == "" {
		relation = mine.Relation
	}

	err := nosql.UpdateVEdgeBase(mine.UID, name, remark, relation, operator, dire, target)
	if err == nil {
		mine.Name = name
		mine.Relation = relation
		mine.Direction = dire
		mine.Target = target
		mine.Operator = operator
		mine.Updated = time.Now().Unix()
	}
	return err
}
