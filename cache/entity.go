package cache

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

const (
	EntityStatusDraft    EntityStatus = 0
	EntityStatusFirst EntityStatus = 1
	EntityStatusPending EntityStatus = 2
	EntityStatusSpecial  EntityStatus = 3
	EntityStatusUsable  EntityStatus = 4
	EntityStatusFailed  EntityStatus = 10
)

const DefaultEntityTable = "entities"

type EntityStatus uint8

type EntityInfo struct {
	Status EntityStatus
	BaseInfo
	Concept     string
	Summary         string
	Description     string
	Cover           string
	Add             string   //消歧义
	Creator         string   //创建者
	Owner           string   //所属单位
	Mark            string // 标记或者来源
	Quote           string   // 引用
	Synonyms        []string //同义词
	Tags            []string //标签
	Properties      []*proxy.PropertyInfo
	StaticEvents    []*proxy.EventBrief
	StaticRelations []*proxy.RelationCaseInfo
	events          []*EventInfo
}

func switchEntityLabel(concept string) string {
	if len(concept) < 1 {
		return DefaultEntityTable
	} else {
		top := Context().GetConcept(concept)
		if top != nil {
			return top.Label()
		} else {
			return DefaultEntityTable
		}
	}
}

func (mine *cacheContext)CreateEntity(info *EntityInfo) error {
	if info == nil {
		return errors.New("the entity info is nil")
	}
	db := new(nosql.Entity)
	db.UID = primitive.NewObjectID()
	db.CreatedTime = time.Now()
	db.ID = nosql.GetEntityNextID(info.table())
	db.Name = info.Name
	db.Description = info.Description
	db.Scene = info.Owner
	db.Creator = info.Creator
	db.Operator = info.Operator
	db.Add = info.Add
	db.Cover = info.Cover
	db.Summary = info.Summary
	db.Quote = info.Quote
	db.Mark = info.Mark
	db.Concept = info.Concept
	db.Status = uint8(info.Status)
	db.Tags = info.Tags
	db.Synonyms = info.Synonyms
	db.Events = info.StaticEvents
	db.Relations = info.StaticRelations
	info.events = make([]*EventInfo, 0, 1)
	if info.Properties == nil {
		info.Properties = make([]*proxy.PropertyInfo, 0, 1)
	}

	db.Properties = info.Properties
	if db.Tags == nil {
		db.Tags = make([]string, 0, 1)
	}
	if db.Synonyms == nil {
		db.Synonyms = make([]string, 0, 1)
	}
	var err error
	err = nosql.CreateEntity(db, info.table())
	if err == nil {
		info.initInfo(db)
		//mine.entities = append(mine.entities, info)
		mine.syncGraphNode(info)
	}
	return err
}

func (mine *cacheContext)syncGraphNode(info *EntityInfo)  {
	var name = info.Name
	if info.Add != "" {
		name = info.Name + "-" + info.Add
	}
	mine.addSyncNode(info.UID, name, info.Concept, info.Cover)
}

func (mine *cacheContext)AllEntities() []*EntityInfo {
	array,err := nosql.GetEntities(DefaultEntityTable)
	if err != nil {
		return make([]*EntityInfo, 0, 0)
	}
	list := make([]*EntityInfo, 0, len(array))
	for _, entity := range array {
		info := new(EntityInfo)
		info.initInfo(entity)
		list = append(list, info)
	}
	for i := 0; i < len(mine.concepts); i += 1 {
		tb := mine.concepts[i].Table
		if len(tb) > 0 {
			arr, err := nosql.GetEntities(tb)
			if err == nil && arr != nil {
				for _, entity := range arr {
					info := new(EntityInfo)
					info.initInfo(entity)
					list = append(list, info)
				}
			}
		}
	}
	return list
}

func (mine *cacheContext)HadEntityByName(name, add string) bool {
	info := mine.GetEntityByName(name, add)
	if info != nil {
		return true
	}
	return false
}

func (mine *cacheContext)HadEntityByMark(mark string) bool {
	info := mine.GetEntityByMark(mark)
	if info != nil {
		return true
	}
	return false
}

func (mine *cacheContext)GetEntityByName(name, add string) *EntityInfo {
	if len(name) < 1 {
		return nil
	}
	db,err := nosql.GetEntityByName(DefaultEntityTable, name, add)
	if err == nil && db != nil {
		info := new(EntityInfo)
		info.initInfo(db)
		return info
	}
	for i := 0; i < len(mine.concepts); i += 1 {
		tb := mine.concepts[i].Table
		if len(tb) > 0 {
			db, err = nosql.GetEntityByName(tb, name, add)
			if err == nil && db != nil {
				info := new(EntityInfo)
				info.initInfo(db)
				return info
			}
		}
	}
	return nil
}

func (mine *cacheContext)GetEntityByMark(mark string) *EntityInfo {
	if len(mark) < 1 {
		return nil
	}
	db,err := nosql.GetEntityByMark(DefaultEntityTable, mark)
	if err == nil && db != nil {
		info := new(EntityInfo)
		info.initInfo(db)
		return info
	}

	return nil
}

func (mine *cacheContext)GetEntitiesByOwner(owner string) []*EntityInfo {
	list := make([]*EntityInfo, 0, 10)
	array,err := nosql.GetEntitiesByOwner(DefaultEntityTable, owner)
	if err != nil {
		return list
	}

	for _, entity := range array {
		info := new(EntityInfo)
		info.initInfo(entity)
		list = append(list, info)
	}
	for i := 0; i < len(mine.concepts); i += 1 {
		tb := mine.concepts[i].Table
		if len(tb) > 0 {
			arr, err := nosql.GetEntitiesByOwner(tb, owner)
			if err == nil && arr != nil {
				for _, entity := range arr {
					info := new(EntityInfo)
					info.initInfo(entity)
					list = append(list, info)
				}
			}
		}
	}
	//for _, value := range mine.entities {
	//	if value.Owner == owner {
	//		list = append(list, value)
	//	}
	//}
	return list
}

func (mine *cacheContext)GetEntitiesByConcept(concept string) []*EntityInfo {
	list := make([]*EntityInfo, 0, 10)
	array,err := nosql.GetEntitiesByConcept(DefaultEntityTable, concept)
	if err != nil {
		return list
	}

	for _, entity := range array {
		info := new(EntityInfo)
		info.initInfo(entity)
		list = append(list, info)
	}

	return list
}

func (mine *cacheContext)GetEntitiesByStatus(status EntityStatus) []*EntityInfo {
	list := make([]*EntityInfo, 0, 10)
	array,err := nosql.GetEntitiesByStatus(DefaultEntityTable, uint8(status))
	if err != nil {
		return list
	}

	for _, entity := range array {
		info := new(EntityInfo)
		info.initInfo(entity)
		list = append(list, info)
	}
	for i := 0; i < len(mine.concepts); i += 1 {
		tb := mine.concepts[i].Table
		if len(tb) > 0 {
			arr, err := nosql.GetEntitiesByStatus(tb, uint8(status))
			if err == nil && arr != nil {
				for _, entity := range arr {
					info := new(EntityInfo)
					info.initInfo(entity)
					list = append(list, info)
				}
			}
		}
	}
	return list
}

func (mine *cacheContext)GetEntitiesByOwnerStatus(owner string, status EntityStatus) []*EntityInfo {
	list := make([]*EntityInfo, 0, 10)
	array,err := nosql.GetEntitiesByOwnerAndStatus(DefaultEntityTable, owner, uint8(status))
	if err != nil {
		return list
	}

	for _, entity := range array {
		info := new(EntityInfo)
		info.initInfo(entity)
		list = append(list, info)
	}
	for i := 0; i < len(mine.concepts); i += 1 {
		tb := mine.concepts[i].Table
		if len(tb) > 0 {
			arr, err := nosql.GetEntitiesByOwnerAndStatus(tb, owner, uint8(status))
			if err == nil && arr != nil {
				for _, entity := range arr {
					info := new(EntityInfo)
					info.initInfo(entity)
					list = append(list, info)
				}
			}
		}
	}
	return list
}

func (mine *cacheContext)GetEntitiesByProp(key, val string) []*EntityInfo {
	list := make([]*EntityInfo, 0, 10)
	array,err := nosql.GetEntitiesByProp(DefaultEntityTable, key, val)
	if err != nil {
		return list
	}

	for _, entity := range array {
		info := new(EntityInfo)
		info.initInfo(entity)
		list = append(list, info)
	}
	for i := 0; i < len(mine.concepts); i += 1 {
		tb := mine.concepts[i].Table
		if len(tb) > 0 {
			arr, err := nosql.GetEntitiesByProp(tb, key, val)
			if err == nil && arr != nil {
				for _, entity := range arr {
					info := new(EntityInfo)
					info.initInfo(entity)
					list = append(list, info)
				}
			}
		}
	}
	//for _, value := range mine.entities {
	//	prop := value.GetProperty(key)
	//	if prop != nil && prop.HadWordByValue(val) {
	//		list = append(list, value)
	//	}
	//}
	return list
}

func (mine *cacheContext)GetEntity(uid string) *EntityInfo {
	if len(uid) < 1 {
		return nil
	}
	//for i := 0; i < len(mine.entities); i++ {
	//	if mine.entities[i].UID == uid {
	//		return mine.entities[i]
	//	}
	//}
	db := mine.getEntityFromDB(uid)
	if db != nil {
		info := new(EntityInfo)
		info.initInfo(db)
		//mine.entities = append(mine.entities, info)
		return info
	}
	return nil
}

func (mine *cacheContext)getEntityFromDB(uid string) *nosql.Entity {
	for i := 0; i < len(mine.concepts); i += 1 {
		tb := mine.concepts[i].Table
		if len(tb) > 0 {
			db, err := nosql.GetEntity(tb, uid)
			if err == nil && db != nil {
				return db
			}
		}
	}
	db, err := nosql.GetEntity(DefaultEntityTable, uid)
	if err == nil && db != nil {
		return db
	}
	return nil
}

func (mine *cacheContext)HadEntity(uid string) bool {
	db := mine.getEntityFromDB(uid)
	if db != nil {
		return true
	}
	return false
	//for i := 0; i < len(mine.entities); i += 1 {
	//	if mine.entities[i].UID == uid {
	//		return true
	//	}
	//}
	//return false
}

func (mine *cacheContext)RemoveEntity(uid, operator string) error {
	if len(uid) < 1 {
		return errors.New("the micro course uid is empty")
	}
	tmp := mine.GetEntity(uid)
	err := nosql.RemoveEntity(tmp.table(), uid, operator)
	if err == nil {
		//length := len(mine.entities)
		//for i := 0; i < length; i++ {
		//	if mine.entities[i].UID == uid {
		//		mine.entities[i].clear()
		//		mine.entities = append(mine.entities[:i], mine.entities[i+1:]...)
		//		break
		//	}
		//}
	}

	return err
}

func (mine *cacheContext)HadOwnerOfAsset(owner string) bool {
	info := mine.GetEntity(owner)
	if info != nil {
		return true
	}
	return false
}

func (mine *EntityInfo) Construct() {
	mine.Tags = make([]string, 0, 5)
	mine.events = make([]*EventInfo, 0, 10)
	mine.Properties = make([]*proxy.PropertyInfo, 0, 10)
}

func (mine *EntityInfo) initInfo(db *nosql.Entity) bool {
	if db == nil {
		return false
	}
	mine.UID = db.UID.Hex()
	mine.ID = db.ID
	mine.CreateTime = db.CreatedTime
	mine.UpdateTime = db.UpdatedTime
	mine.Creator = db.Creator
	mine.Operator = db.Operator
	mine.Tags = db.Tags
	mine.Name = db.Name
	mine.Add = db.Add
	mine.Description = db.Description
	mine.Concept = db.Concept
	mine.Status = EntityStatus(db.Status)
	mine.Owner = db.Scene
	mine.Cover = db.Cover
	mine.Mark = db.Mark
	mine.Quote = db.Quote
	mine.Summary = db.Summary
	mine.StaticEvents = db.Events
	mine.StaticRelations = db.Relations
	if mine.StaticRelations == nil {
		mine.StaticRelations = make([]*proxy.RelationCaseInfo, 0, 1)
	}
	if mine.StaticEvents == nil {
		mine.StaticEvents = make([]*proxy.EventBrief, 0, 1)
	}

	mine.Properties = make([]*proxy.PropertyInfo, 0, 10)
	if db.Properties != nil {
		mine.Properties = db.Properties
	}
	events, err := nosql.GetEventsByParent(mine.UID)

	if err == nil {
		mine.events = make([]*EventInfo, 0, len(events))
		for _, event := range events {
			tmp := new(EventInfo)
			tmp.initInfo(event)
			mine.events = append(mine.events, tmp)
		}
	} else {
		mine.events = make([]*EventInfo, 0, 2)
	}

	return true
}

func (mine *EntityInfo) clear() {
	mine.UID = ""
}

func (mine *EntityInfo) table() string {
	if len(mine.Concept) < 1 {
		return DefaultEntityTable
	} else {
		top := Context().GetTopConcept(mine.Concept)
		if top != nil {
			if len(top.Table) > 0 {
				return top.Table
			} else {
				return DefaultEntityTable
			}
		} else {
			return DefaultEntityTable
		}
	}
}

func (mine *EntityInfo) UpdateBase(name, remark, add, concept, cover, mark, quote, sum, operator string) error {
	if concept == "" {
		concept = mine.Concept
	}
	if remark == "" {
		remark = mine.Description
	}
	if add == "" {
		add = mine.Add
	}
	if name == "" {
		name = mine.Name
	}
	if mark == "" {
		mark = mine.Mark
	}
	if sum == "" {
		sum = mine.Summary
	}
	if quote == "" {
		quote = mine.Quote
	}
	var err error
	if len(cover) > 0 {
		err = mine.UpdateCover(cover, operator)
	}
	if name != mine.Name || remark != mine.Description || add != mine.Add || concept != mine.Concept {
		err = nosql.UpdateEntityBase(mine.table(), mine.UID, name, remark, add, concept, operator)
		if err == nil {
			mine.Name = name
			mine.Description = remark
			mine.Add = add
			mine.Concept = concept
			mine.Operator = operator
			mine.UpdateTime = time.Now()
		}
	}
	return err
}

func (mine *EntityInfo) UpdateStatic(info *EntityInfo) error {
	mine.UpdateBase(info.Name, info.Description, info.Add, info.Concept, info.Cover, info.Mark, info.Quote, info.Summary, info.Operator)
	err := nosql.UpdateEntityStatic(mine.table(), mine.UID, info.Operator, info.Properties, info.StaticEvents, info.StaticRelations)
	if err == nil {
		mine.Properties = info.Properties
		mine.StaticEvents = info.StaticEvents
		mine.StaticRelations = info.StaticRelations
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EntityInfo) UpdateCover(cover, operator string) error {
	if cover == "" {
		cover = mine.Cover
	}
	err := nosql.UpdateEntityCover(mine.table(), mine.UID, cover, operator)
	if err == nil {
		mine.Cover = cover
		mine.Operator = operator
		mine.UpdateTime = time.Now()
		Context().graph.UpdateNodeCover(mine.UID, cover)
	}
	return err
}

func (mine *EntityInfo) UpdateTags(tags []string, operator string) error {
	err := nosql.UpdateEntityTags(mine.table(), mine.UID, operator, tags)
	if err == nil {
		mine.Tags = tags
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EntityInfo) UpdateSynonyms(list []string, operator string) error {
	err := nosql.UpdateEntitySynonyms(mine.table(), mine.UID, operator, list)
	if err == nil {
		mine.Synonyms = list
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EntityInfo) UpdateStatus(status EntityStatus, operator string) error {
	err := nosql.UpdateEntityStatus(mine.table(), mine.UID, uint8(status), operator)
	if err == nil {
		mine.Status = status
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

//region Event Fun
func (mine *EntityInfo) AllEvents() []*EventInfo {
	return mine.events
}

func (mine *EntityInfo) AddEvent(date proxy.DateInfo, place proxy.PlaceInfo, name, desc, cover, operator string, links []proxy.RelationCaseInfo, tags, assets []string) (*EventInfo, error) {
	if mine.events == nil {
		return nil, errors.New("must call construct fist")
	}

	db := new(nosql.Event)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetEventNextID()
	db.CreatedTime = time.Now()
	db.Creator = operator
	db.Name = name
	db.Date = date
	db.Place = place
	db.Entity = mine.UID
	db.Description = desc
	db.Relations = links
	db.Cover = cover
	db.Tags = tags
	if db.Tags == nil {
		db.Tags = make([]string, 0, 1)
	}
	db.Assets = assets
	if db.Assets == nil {
		db.Assets = make([]string, 0, 1)
	}
	err := nosql.CreateEvent(db)
	if err == nil {
		info := new(EventInfo)
		info.initInfo(db)
		mine.events = append(mine.events, info)

		for i := 0; i < len(links); i += 1 {
			relationKind := Context().GetRelation(links[i].Category)
			if relationKind != nil {
				Context().addSyncLink(mine.UID, links[i].Entity, relationKind.UID, links[i].Name, switchRelationToLink(relationKind.Kind), links[i].Direction)
			}
		}

		return info, nil
	}
	return nil, err
}

func (mine *EntityInfo) HadEvent(uid string) bool {
	for i := 0; i < len(mine.events); i += 1 {
		if mine.events[i].UID == uid {
			return true
		}
	}
	return false
}

func (mine *EntityInfo)HadEventBy(time, place string) bool {
	for _, event := range mine.events {
		if event.Date.Begin.String() == time && event.Place.Name == place {
			return true
		}
	}
	return false
}

func (mine *EntityInfo)GetEventBy(time, place string) *EventInfo {
	for _, event := range mine.events {
		if event.Date.Begin.String() == time && event.Place.Name == place {
			return event
		}
	}
	return nil
}

func (mine *EntityInfo) RemoveEvent(uid, operator string) error {
	if mine.events == nil {
		return errors.New("must call construct fist")
	}
	if !mine.HadEvent(uid) {
		return errors.New("not found the event")
	}
	err := nosql.RemoveEvent(uid, operator)
	if err == nil {
		for i := 0; i < len(mine.events); i += 1 {
			if mine.events[i].UID == uid {
				mine.events = append(mine.events[:i], mine.events[i+1:]...)
				break
			}
		}
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EntityInfo) GetEvent(uid string) *EventInfo {
	if mine.events == nil {
		return nil
	}
	for i := 0; i < len(mine.events); i += 1 {
		if mine.events[i].UID == uid {
			return mine.events[i]
		}
	}
	return nil
}

//endregion

//region Property Fun
func (mine *EntityInfo) addProp(key string, words []proxy.WordInfo) {
	if mine.Properties == nil {
		return
	}
	mine.Properties = append(mine.Properties, &proxy.PropertyInfo{Key: key, Words: words})
}

func (mine *EntityInfo) AddProperty(key string, words []proxy.WordInfo) error {
	if mine.Properties == nil {
		return errors.New("must call construct fist")
	}
	if len(key) < 1 || len(words) < 1 {
		return errors.New("the prop key or value is empty")
	}
	pair := proxy.PropertyInfo{Key: key, Words: words}
	err := nosql.AppendEntityProperty(mine.table(), mine.UID, pair)
	if err == nil {
		mine.Properties = append(mine.Properties, &pair)
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EntityInfo) UpdateProperties(array []*proxy.PropertyInfo, operator string) error {
	err := nosql.UpdateEntityProperties(mine.table(), mine.UID, operator, array)
	if err == nil {
		mine.Properties = array
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EntityInfo) HadProperty(attribute string) bool {
	if mine.Properties == nil {
		return false
	}
	for i := 0; i < len(mine.Properties); i += 1 {
		if mine.Properties[i].Key == attribute {
			return true
		}
	}
	return false
}

func (mine *EntityInfo) HadPropertyByEntity(uid string) bool {
	if mine.Properties == nil {
		return false
	}
	for i := 0; i < len(mine.Properties); i += 1 {
		if mine.Properties[i].HadWordByEntity(uid) {
			return true
		}
	}
	return false
}

func (mine *EntityInfo) RemoveProperty(attribute string) error {
	if mine.Properties == nil {
		return errors.New("must call construct fist")
	}
	if !mine.HadProperty(attribute) {
		return errors.New("not found the property when remove")
	}
	err := nosql.SubtractEntityProperty(mine.table(), mine.UID, attribute)
	if err == nil {
		for i := 0; i < len(mine.Properties); i += 1 {
			if mine.Properties[i].Key == attribute {
				mine.Properties = append(mine.Properties[:i], mine.Properties[i+1:]...)
				break
			}
		}
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EntityInfo) GetProperty(attribute string) *proxy.PropertyInfo {
	if mine.Properties == nil {
		return nil
	}
	for i := 0; i < len(mine.Properties); i += 1 {
		if mine.Properties[i].Key == attribute {
			return mine.Properties[i]
		}
	}
	return nil
}

func (mine *EntityInfo) IsSatisfy(concepts, attributes, tags []string) bool {
	if hadItem(concepts, mine.Concept) {
		return true
	}
	if mine.Properties != nil {
		for i := 0; i < len(mine.Properties); i += 1 {
			if hadItem(attributes, mine.Properties[i].Key) {
				return true
			}
		}
	}
	if mine.Tags != nil {
		for _, tag := range mine.Tags {
			if hadItem(tags, tag) {
				return true
			}
		}
	}

	return false
}
//endregion
