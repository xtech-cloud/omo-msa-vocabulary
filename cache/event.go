package cache

import (
	"fmt"
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

type EventInfo struct {
	BaseInfo
	Description string
	Parent      string
	Cover       string
	Date        proxy.DateInfo
	Place       proxy.PlaceInfo
	Tags        []string
	Assets      []string
	Relations   []proxy.RelationCaseInfo
}

func (mine *cacheContext)GetEvent(uid string) *EventInfo {
	event,err := nosql.GetEvent(uid)
	if err == nil && event != nil {
		info := new(EventInfo)
		info.initInfo(event)
		return info
	}

	return nil
}

func (mine *cacheContext)RemoveEvent(uid, operator string) error {
	return nosql.RemoveEvent(uid, operator)
}

func (mine *EventInfo) initInfo(db *nosql.Event) {
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
	mine.Cover = db.Cover
	mine.Assets = db.Assets
	mine.Tags = db.Tags
	mine.Relations = db.Relations
}

func (mine *EventInfo) UpdateBase(name, remark, operator string, date proxy.DateInfo, place proxy.PlaceInfo, assets []string) error {
	if name == "" {
		name = mine.Name
	}
	if remark == "" {
		remark = mine.Description
	}
	if assets == nil {
		assets = mine.Assets
	}
	err := nosql.UpdateEventBase(mine.UID, name, remark, operator, date, place, assets)
	if err == nil {
		mine.Name = name
		mine.Description = remark
		mine.Date = date
		mine.Place = place
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

func (mine *EventInfo)hadAsset(asset string) bool {
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
			Context().addSyncLink(mine.Parent, relation.Entity, tmp.UID, relation.Name, switchRelationToLink(tmp.Kind), relation.Direction)
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
