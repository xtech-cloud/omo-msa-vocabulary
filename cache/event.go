package cache

import (
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
)

type EventInfo struct {
	BaseInfo
	Description string
	Date        proxy.DateInfo
	Place       proxy.PlaceInfo
	Assets      []string
	Relations   []proxy.RelationInfo
}

func (mine *EventInfo)initInfo(db *nosql.Event)  {
	mine.UID = db.UID.Hex()
	mine.ID = db.ID
	mine.CreateTime = db.CreatedTime
	mine.UpdateTime = db.UpdatedTime
	mine.Name = db.Name
	mine.Description = db.Description
	mine.Date = db.Date
	mine.Place = db.Place
	mine.Assets = db.Assets
	mine.Relations = db.Relations
}

func (mine *EventInfo)UpdateBase(name, remark string, date proxy.DateInfo, place proxy.PlaceInfo) error {
	err := nosql.UpdateEventBase(mine.UID, name, remark, date, place)
	if err == nil {
		mine.Name = name
		mine.Description = remark
		mine.Date = date
		mine.Place = place
	}
	return err
}

func (mine *EventInfo)AppendAsset(asset string) error {
	err := nosql.AppendEventAsset(mine.UID, asset)
	if err == nil {
		mine.Assets = append(mine.Assets, asset)
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
	}
	return err
}

func (mine *EventInfo)AppendRelation(relation proxy.RelationInfo) error {
	err := nosql.AppendEventRelation(mine.UID, relation)
	if err == nil {
		mine.Relations = append(mine.Relations, relation)
	}
	return err
}

func (mine *EventInfo)SubtractRelation(relation string) error {
	err := nosql.SubtractEventRelation(mine.UID, relation)
	if err == nil {
		for i := 0;i < len(mine.Assets);i += 1 {
			if mine.Relations[i].UID == relation {
				mine.Relations = append(mine.Relations[:i], mine.Relations[i+1:]...)
				break
			}
		}
	}
	return err
}