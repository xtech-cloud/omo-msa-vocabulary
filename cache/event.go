package cache

import (
	"fmt"
	"github.com/micro/go-micro/v2/logger"
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

type EventInfo struct {
	BaseInfo
	Description string
	Parent string
	Date        proxy.DateInfo
	Place       proxy.PlaceInfo
	Assets      []string
	Relations   []proxy.RelationCaseInfo
}

func GetEvent(uid string) *EventInfo {
	for i := 0;i < len(cacheCtx.entities);i += 1 {
		info := cacheCtx.entities[i].GetEvent(uid)
		if info != nil {
			return info
		}
	}
	return nil
}

func RemoveEvent(uid, operator string) error {
	for i := 0;i < len(cacheCtx.entities);i += 1 {
		if cacheCtx.entities[i].HadEvent(uid) {
			return cacheCtx.entities[i].RemoveEvent(uid, operator)
		}
	}
	return nil
}

func (mine *EventInfo)initInfo(db *nosql.Event)  {
	mine.UID = db.UID.Hex()
	mine.ID = db.ID
	mine.CreateTime = db.CreatedTime
	mine.UpdateTime = db.UpdatedTime
	mine.Operator = db.Operator
	mine.Creator = db.Creator
	mine.Name = db.Name
	mine.Parent = db.Entity
	mine.Description = db.Description
	mine.Date = db.Date
	mine.Place = db.Place
	mine.Assets = db.Assets
	mine.Relations = db.Relations
}

func (mine *EventInfo)UpdateBase(name, remark, operator string, date proxy.DateInfo, place proxy.PlaceInfo) error {
	err := nosql.UpdateEventBase(mine.UID, name, remark,operator, date, place)
	if err == nil {
		mine.Name = name
		mine.Description = remark
		mine.Date = date
		mine.Place = place
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EventInfo)AppendAsset(asset string) error {
	err := nosql.AppendEventAsset(mine.UID, asset)
	if err == nil {
		mine.Assets = append(mine.Assets, asset)
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EventInfo)SubtractAsset(asset string) error {
	err := nosql.SubtractEventAsset(mine.UID, asset)
	if err == nil {
		for i := 0;i < len(mine.Assets);i += 1 {
			if mine.Assets[i] == asset {
				mine.Assets = append(mine.Assets[:i], mine.Assets[i+1:]...)
				break
			}
		}
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EventInfo)AppendRelation(relation *proxy.RelationCaseInfo) error {
	relation.UID = fmt.Sprintf("%s-%d", mine.UID, nosql.GetRelationCaseNextID())
	err := nosql.AppendEventRelation(mine.UID, relation)
	if err == nil {
		mine.Relations = append(mine.Relations, *relation)
		mine.UpdateTime = time.Now()
		tmp := GetRelation(relation.Category)
		if tmp != nil {
			go createLink(mine.Parent, relation.Entity, switchRelationToLink(tmp.Kind),tmp.UID, relation.Name, relation.Direction )
		}
	}
	return err
}

func createLink(from, to string, kind LinkType, relationUID, name string, dire uint8)  {
	fromNode := GetGraphNode(from)
	toNode := GetGraphNode(to)
	_, err := cacheCtx.graph.CreateLink(fromNode, toNode, kind, name, relationUID, DirectionType(dire))
	if err != nil {
		logger.Warn(err.Error())
	}
}

func (mine *EventInfo)SubtractRelation(relation string) error {
	err := nosql.SubtractEventRelation(mine.UID, relation)
	if err == nil {
		for i := 0;i < len(mine.Relations);i += 1 {
			if mine.Relations[i].UID == relation {
				mine.Relations = append(mine.Relations[:i], mine.Relations[i+1:]...)
				break
			}
		}
		mine.UpdateTime = time.Now()
	}
	return err
}