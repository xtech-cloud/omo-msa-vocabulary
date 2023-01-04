package cache

import (
	"fmt"
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

const (
	EventCustom   = 0
	EventActivity = 1
	EventHonor    = 2
)

const (
	AccessPublic  = 0
	AccessPrivate = 1
)

type EventInfo struct {
	Type   uint8
	Access uint8 // 可访问对象
	BaseInfo
	Description string // 描述
	Entity      string //实体对象
	Parent      string //父级事件
	Cover       string //封面
	Quote       string // 引用或者备注，活动，超链接
	Date        proxy.DateInfo
	Place       proxy.PlaceInfo
	Tags        []string
	Assets      []string
	Relations   []proxy.RelationCaseInfo
}

func (mine *EventInfo) initInfo(db *nosql.Event) {
	mine.UID = db.UID.Hex()
	mine.ID = db.ID
	mine.Type = db.Type
	mine.CreateTime = db.CreatedTime
	mine.UpdateTime = db.UpdatedTime
	mine.Operator = db.Operator
	mine.Creator = db.Creator
	mine.Name = db.Name
	mine.Entity = db.Entity
	mine.Parent = db.Parent
	mine.Description = db.Description
	mine.Date = db.Date
	mine.Place = db.Place
	mine.Cover = db.Cover
	mine.Quote = db.Quote
	mine.Assets = db.Assets
	mine.Access = db.Access
	mine.Tags = db.Tags
	mine.Relations = db.Relations
}

func (mine *EventInfo) UpdateBase(name, remark, operator string, access uint8, date proxy.DateInfo, place proxy.PlaceInfo, assets []string) error {
	if name == "" {
		name = mine.Name
	}
	if remark == "" {
		remark = mine.Description
	}
	if assets == nil {
		assets = mine.Assets
	}
	err := nosql.UpdateEventBase(mine.UID, name, remark, operator, access, date, place, assets)
	if err == nil {
		mine.Name = name
		mine.Description = remark
		mine.Date = date
		mine.Place = place
		mine.Access = access
		mine.Assets = assets
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EventInfo) UpdateInfo(name, remark, operator string) error {
	if name == "" {
		name = mine.Name
	}
	if remark == "" {
		remark = mine.Description
	}
	err := nosql.UpdateEventInfo(mine.UID, name, remark, operator)
	if err == nil {
		mine.Name = name
		mine.Description = remark
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EventInfo) UpdateAssets(operator string, list []string) error {
	if operator == "" {
		operator = mine.Operator
	}
	err := nosql.UpdateEventAssets(mine.UID, operator, list)
	if err == nil {
		mine.Assets = list
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EventInfo) UpdateTags(operator string, tags []string) error {
	if operator == "" {
		operator = mine.Operator
	}
	err := nosql.UpdateEventTags(mine.UID, operator, tags)
	if err == nil {
		mine.Tags = tags
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EventInfo) UpdateAccess(operator string, access uint8) error {
	if operator == "" {
		operator = mine.Operator
	}
	if mine.Access == access {
		return nil
	}
	err := nosql.UpdateEventAccess(mine.UID, operator, access)
	if err == nil {
		mine.Access = access
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EventInfo) UpdateQuote(quote, operator string) error {
	if operator == "" {
		operator = mine.Operator
	}
	if mine.Quote == quote {
		return nil
	}
	err := nosql.UpdateEventQuote(mine.UID, quote, operator)
	if err == nil {
		mine.Quote = quote
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EventInfo) UpdateCover(operator, cover string) error {
	if operator == "" {
		operator = mine.Operator
	}
	if mine.Cover == cover {
		return nil
	}
	err := nosql.UpdateEventCover(mine.UID, operator, cover)
	if err == nil {
		mine.Cover = cover
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EventInfo) hadAsset(asset string) bool {
	for _, s := range mine.Assets {
		if s == asset {
			return true
		}
	}
	return false
}

func (mine *EventInfo) AppendAsset(asset string) error {
	if mine.hadAsset(asset) {
		return nil
	}
	err := nosql.AppendEventAsset(mine.UID, asset)
	if err == nil {
		mine.Assets = append(mine.Assets, asset)
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EventInfo) SubtractAsset(asset string) error {
	if !mine.hadAsset(asset) {
		return nil
	}
	err := nosql.SubtractEventAsset(mine.UID, asset)
	if err == nil {
		for i := 0; i < len(mine.Assets); i += 1 {
			if mine.Assets[i] == asset {
				mine.Assets = append(mine.Assets[:i], mine.Assets[i+1:]...)
				break
			}
		}
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EventInfo) AppendRelation(relation *proxy.RelationCaseInfo) error {
	relation.UID = fmt.Sprintf("%s-%d", mine.UID, nosql.GetRelationCaseNextID())
	err := nosql.AppendEventRelation(mine.UID, relation)
	if err == nil {
		mine.Relations = append(mine.Relations, *relation)
		mine.UpdateTime = time.Now()
		tmp := Context().GetRelation(relation.Category)
		if tmp != nil {
			Context().addSyncLink(mine.Entity, relation.Entity, tmp.UID, relation.Name, switchRelationToLink(tmp.Kind), relation.Direction)
		}
	}
	return err
}

func (mine *EventInfo) SubtractRelation(relation string) error {
	err := nosql.SubtractEventRelation(mine.UID, relation)
	if err == nil {
		for i := 0; i < len(mine.Relations); i += 1 {
			if mine.Relations[i].UID == relation {
				mine.Relations = append(mine.Relations[:i], mine.Relations[i+1:]...)
				break
			}
		}
		mine.UpdateTime = time.Now()
	}
	return err
}
