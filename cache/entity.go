package cache

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

const (
	EntityStatusIdle    EntityStatus = 0
	EntityStatusPending EntityStatus = 1
	EntityStatusUsable  EntityStatus = 2
	EntityStatusFailed  EntityStatus = 3
)

const DefaultEntityTable = "entities"

type EntityStatus uint8

type EntityInfo struct {
	Status EntityStatus
	BaseInfo
	Concept     string
	Description string
	Cover       string
	Add         string   //消歧义
	Creator     string   //创建者
	Owner       string   //所属单位
	Synonyms    []string //同义词
	Tags        []string //标签
	properties  []*proxy.PropertyInfo
	events      []*EventInfo
}

func switchEntityName(concept string) string {
	if len(concept) < 1 {
		return "entities"
	} else {
		top := GetTopConcept(concept)
		if top != nil {
			return top.Table
		} else {
			return "entities"
		}
	}
}

func CreateEntity(info *EntityInfo) error {
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
	db.Concept = info.Concept
	db.Status = uint8(info.Status)
	db.Tags = info.Tags
	db.Synonyms = info.Synonyms
	info.events = make([]*EventInfo, 0, 1)
	info.properties = make([]*proxy.PropertyInfo, 0, 1)
	db.Properties = info.properties
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
		cacheCtx.entities = append(cacheCtx.entities, info)
		go createGraphNode(info)
	}
	return err
}

func createGraphNode(info *EntityInfo) (*NodeInfo,error) {
	node, err := cacheCtx.graph.CreateNodeByEntity(info)
	return node,err
}

func AllEntities() []*EntityInfo {
	return cacheCtx.entities
}

func HadEntityByName(name string) bool {
	for i := 0; i < len(cacheCtx.entities); i++ {
		if cacheCtx.entities[i].Name == name {
			return true
		}
	}
	return false
}

func GetEntityByName(name string) *EntityInfo {
	if len(name) < 1 {
		return nil
	}
	for i := 0; i < len(cacheCtx.entities); i++ {
		if cacheCtx.entities[i].Name == name {
			return cacheCtx.entities[i]
		}
	}
	return nil
}

func GetEntity(uid string) *EntityInfo {
	if len(uid) < 1 {
		return nil
	}
	for i := 0; i < len(cacheCtx.entities); i++ {
		if cacheCtx.entities[i].UID == uid {
			return cacheCtx.entities[i]
		}
	}
	db := getEntityFromDB(uid)
	if db != nil {
		info := new(EntityInfo)
		info.initInfo(db)
		cacheCtx.entities = append(cacheCtx.entities, info)
		return info
	}
	return nil
}

func getEntityFromDB(uid string) *nosql.Entity {
	for i := 0; i < len(cacheCtx.concerts); i += 1 {
		tb := cacheCtx.concerts[i].Table
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

func HadEntity(uid string) bool {
	for i := 0; i < len(cacheCtx.entities); i += 1 {
		if cacheCtx.entities[i].UID == uid {
			return true
		}
	}
	return false
}

func RemoveEntity(uid, operator string) error {
	if len(uid) < 1 {
		return errors.New("the micro course uid is empty")
	}
	tmp := GetEntity(uid)
	err := nosql.RemoveEntity(tmp.table(), uid, operator)
	if err == nil {
		length := len(cacheCtx.entities)
		for i := 0; i < length; i++ {
			if cacheCtx.entities[i].UID == uid {
				cacheCtx.entities[i].clear()
				cacheCtx.entities = append(cacheCtx.entities[:i], cacheCtx.entities[i+1:]...)
				break
			}
		}

	}

	return err
}

func HadOwnerOfAsset(owner string) bool {
	info := GetEntity(owner)
	if info != nil {
		return true
	}
	return false
}

func createSampleEntity(name string, concept string) (*EntityInfo, *NodeInfo, error) {
	if len(name) < 1 {
		return nil, nil, errors.New("the entity name is nil")
	}
	var info *EntityInfo
	info = GetEntityByName(name)
	var node *NodeInfo
	var err error
	if info == nil {
		info = new(EntityInfo)
		info.Construct()
		info.Name = name
		info.Concept = concept
		info.Cover = ""
		err = CreateEntity(info)
		if err == nil {
			node,err = createGraphNode(info)
		}
	}

	if node == nil {
		node = createNode(name, info.UID)
	}
	return info, node, err
}

func createNode(name string, entity string) *NodeInfo {
	node := cacheCtx.graph.GetNode(entity)
	if node != nil {
		//fmt.Println("the node(" + name + ") is exist !")
		return node
	}
	node, _ = cacheCtx.graph.CreateNode(name, entity, name+".jpg", "")
	return node
}

func (mine *EntityInfo) Construct() {
	mine.Tags = make([]string, 0, 5)
	mine.events = make([]*EventInfo, 0, 10)
	mine.properties = make([]*proxy.PropertyInfo, 0, 10)
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

	mine.properties = make([]*proxy.PropertyInfo, 0, 10)
	if db.Properties != nil {
		mine.properties = db.Properties
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
		return "entities"
	} else {
		top := GetTopConcept(mine.Concept)
		if top != nil {
			if len(top.Table) > 0 {
				return top.Table
			}else{
				return DefaultEntityTable
			}
		} else {
			return DefaultEntityTable
		}
	}
}

func (mine *EntityInfo) UpdateBase(name, remark, add, concept, operator string) error {
	err := nosql.UpdateEntityBase(mine.table(), mine.UID, name, remark, add, concept, operator)
	if err == nil {
		mine.Name = name
		mine.Description = remark
		mine.Add = add
		mine.Concept = concept
		mine.Operator = operator
	}
	return err
}

func (mine *EntityInfo) UpdateCover(cover, operator string) error {
	err := nosql.UpdateEntityCover(mine.table(), mine.UID, cover, operator)
	if err == nil {
		mine.Cover = cover
		mine.Operator = operator
		cacheCtx.graph.UpdateNodeCover(mine.UID, cover)
	}
	return err
}

func (mine *EntityInfo) UpdateTags(tags []string, operator string) error {
	err := nosql.UpdateEntityTags(mine.table(), mine.UID, operator, tags)
	if err == nil {
		mine.Tags = tags
		mine.Operator = operator
	}
	return err
}

func (mine *EntityInfo) UpdateSynonyms(list []string, operator string) error {
	err := nosql.UpdateEntitySynonyms(mine.table(), mine.UID, operator, list)
	if err == nil {
		mine.Synonyms = list
		mine.Operator = operator
	}
	return err
}

func (mine *EntityInfo) UpdateStatus(status EntityStatus ,operator string) error {
	err := nosql.UpdateEntityStatus(mine.table(), mine.UID, uint8(status), operator)
	if err == nil {
		mine.Status = status
		mine.Operator = operator
	}
	return err
}

//region Event Fun
func (mine *EntityInfo) AllEvents() []*EventInfo {
	return mine.events
}

func (mine *EntityInfo) AddEvent(date proxy.DateInfo, place proxy.PlaceInfo,name, desc, operator string, links []proxy.RelationCaseInfo, assets []string) (*EventInfo, error) {
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
	db.Assets = assets
	err := nosql.CreateEvent(db)
	if err == nil {
		info := new(EventInfo)
		info.initInfo(db)
		mine.events = append(mine.events, info)
		//from := GetGraphNode(mine.UID)
		//for i := 0; i < len(links); i += 1 {
		//	node := GetGraphNode(links[i].Entity)
		//	_, _ = CreateLink(from, node, LinkType(links[i].Category), links[i].Name, "", DirectionType(links[i].Direction))
		//}
		//concept := GetConceptByName("地理")
		//_, _, _ = createSampleEntity(place.Name, concept.Name)
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
	if mine.properties == nil {
		return
	}
	mine.properties = append(mine.properties, &proxy.PropertyInfo{Key: key, Words: words})
}

func (mine *EntityInfo) AddProperty(key string, words []proxy.WordInfo) error {
	if mine.properties == nil {
		return errors.New("must call construct fist")
	}
	if len(key) < 1 || len(words) < 1 {
		return errors.New("the prop key or value is empty")
	}
	pair := proxy.PropertyInfo{Key: key, Words: words}
	err := nosql.AppendEntityProperty(mine.table(), mine.UID, pair)
	if err == nil {
		mine.properties = append(mine.properties, &pair)
	}
	return err
}

func (mine *EntityInfo) AddProperties(array []*proxy.PropertyInfo, operator string) error {
	err := nosql.UpdateEntityProperties(mine.table(), mine.UID, operator, array)
	if err == nil {
		mine.properties = array
	}
	return err
}

func (mine *EntityInfo) Properties() []*proxy.PropertyInfo {
	return mine.properties
}

func (mine *EntityInfo) HadProperty(key string) bool {
	if mine.properties == nil {
		return false
	}
	for i := 0; i < len(mine.properties); i += 1 {
		if mine.properties[i].Key == key {
			return true
		}
	}
	return false
}

func (mine *EntityInfo) HadPropertyByEntity(uid string) bool {
	if mine.properties == nil {
		return false
	}
	for i := 0; i < len(mine.properties); i += 1 {
		if mine.properties[i].HadWordByEntity(uid) {
			return true
		}
	}
	return false
}

func (mine *EntityInfo) RemoveProperty(key string) error {
	if mine.properties == nil {
		return errors.New("must call construct fist")
	}
	if !mine.HadProperty(key) {
		return errors.New("not found the property when remove")
	}
	err := nosql.SubtractEntityProperty(mine.table(), mine.UID, key)
	if err == nil {
		for i := 0; i < len(mine.properties); i += 1 {
			if mine.properties[i].Key == key {
				mine.properties = append(mine.properties[:i], mine.properties[i+1:]...)
				break
			}
		}
	}
	return err
}

func (mine *EntityInfo) GetProperty(key string) *proxy.PropertyInfo {
	if mine.properties == nil {
		return nil
	}
	for i := 0; i < len(mine.properties); i += 1 {
		if mine.properties[i].Key == key {
			return mine.properties[i]
		}
	}
	return nil
}

//endregion
