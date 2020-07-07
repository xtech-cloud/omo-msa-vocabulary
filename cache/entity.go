package cache

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/config"
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
	assets      []*AssetInfo
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

func CreateEntity(info *EntityInfo) (*NodeInfo, error) {
	if info == nil {
		return nil, errors.New("the entity info is nil")
	}
	db := new(nosql.Entity)
	db.UID = primitive.NewObjectID()
	db.CreatedTime = time.Now()
	db.ID = nosql.GetEntityNextID(info.table())
	db.Name = info.Name
	db.Description = info.Description
	db.Owner = ""
	db.Add = info.Add
	db.Cover = info.Cover
	db.Concept = info.Concept
	db.Status = uint8(EntityStatusIdle)
	db.Tags = info.Tags
	info.events = make([]*EventInfo, 0, 1)
	info.properties = make([]*proxy.PropertyInfo, 0, 1)
	db.Properties = info.properties
	info.assets = make([]*AssetInfo, 0, 1)
	if db.Tags == nil {
		db.Tags = make([]string, 0, 1)
	}
	var err error
	err = nosql.CreateEntity(db, info.table())
	var node *NodeInfo
	if err == nil {
		info.initInfo(db)
		cacheCtx.entities = append(cacheCtx.entities, info)
		node, err = cacheCtx.graph.CreateNodeByEntity(info)
	}
	return node, err
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
		db, err := nosql.GetEntity(cacheCtx.concerts[i].Table, uid)
		if err == nil && db != nil {
			return db
		}
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

func RemoveEntity(uid string) error {
	if len(uid) < 1 {
		return errors.New("the micro course uid is empty")
	}
	tmp := GetEntity(uid)
	err := nosql.RemoveEntity(tmp.table(), uid)
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
		node, err = CreateEntity(info)
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
	mine.assets = make([]*AssetInfo, 0, 2)
}

func (mine *EntityInfo) initInfo(db *nosql.Entity) bool {
	if db == nil {
		return false
	}
	mine.UID = db.UID.Hex()
	mine.ID = db.ID
	mine.CreateTime = db.CreatedTime
	mine.UpdateTime = db.UpdatedTime
	mine.Tags = db.Tags
	mine.Name = db.Name
	mine.Add = db.Add
	mine.Description = db.Description
	mine.Concept = db.Concept
	mine.Status = EntityStatus(db.Status)
	mine.Owner = db.Owner
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

	assets, _ := nosql.GetAssetsByOwner(mine.UID)
	mine.assets = make([]*AssetInfo, 0, 2)
	for i := 0; i < len(assets); i += 1 {
		var info = new(AssetInfo)
		if info.initInfo(assets[i]) {
			mine.assets = append(mine.assets, info)
		}
	}
	return true
}

func (mine *EntityInfo) clear() {
	mine.UID = ""
	mine.assets = mine.assets[0:0]
	mine.assets = nil
}

func (mine *EntityInfo) table() string {
	if len(mine.Concept) < 1 {
		return "entities"
	} else {
		top := GetTopConcept(mine.Concept)
		if top != nil {
			return top.Table
		} else {
			return "entities"
		}
	}
}

func (mine *EntityInfo) UpdateBase(name, remark, add, concept string) error {
	err := nosql.UpdateEntityBase(mine.table(), mine.UID, name, remark, add, concept)
	if err == nil {
		mine.Name = name
		mine.Description = remark
		mine.Add = add
		mine.Concept = concept
	}
	return err
}

func (mine *EntityInfo) UpdateCover(cover string) error {
	err := nosql.UpdateEntityCover(mine.table(), mine.UID, cover)
	if err == nil {
		mine.Cover = cover
		cacheCtx.graph.UpdateNodeCover(mine.UID, cover)
	}
	return err
}

func (mine *EntityInfo) UpdateTags(tags []string) error {
	if len(tags) > config.Schema.Basic.TagMax {
		return errors.New(fmt.Sprintf("the tag max number is %d", config.Schema.Basic.TagMax))
	}
	err := nosql.UpdateEntityTags(mine.table(), mine.UID, tags)
	if err == nil {
		mine.Tags = tags
	}
	return err
}

func (mine *EntityInfo) UpdateSynonyms(list []string) error {
	if len(list) > config.Schema.Basic.SynonymMax {
		return errors.New(fmt.Sprintf("the tag max number is %d", config.Schema.Basic.SynonymMax))
	}
	err := nosql.UpdateEntitySynonyms(mine.table(), mine.UID, list)
	if err == nil {
		mine.Synonyms = list
	}
	return err
}

func (mine *EntityInfo) UpdateStatus(status EntityStatus) error {
	if mine.Status != EntityStatusPending {
		return errors.New("the micro course had deal")
	}
	err := nosql.UpdateEntityStatus(mine.table(), mine.UID, uint8(status))
	if err == nil {
		mine.Status = status
	}
	return err
}

func (mine *EntityInfo) DefaultURL() string {
	info := mine.GetAssetByFilter(AssetTypeVideo, AssetLanguageCN)
	if info != nil {
		return info.URL()
	}
	return ""
}

//region Event Fun
func (mine *EntityInfo) AllEvents() []*EventInfo {
	return mine.events
}

func (mine *EntityInfo) AddEvent(date proxy.DateInfo, place proxy.PlaceInfo, desc string, links []proxy.RelationInfo, assets []string) (*EventInfo, error) {
	if mine.events == nil {
		return nil, errors.New("must call construct fist")
	}

	db := new(nosql.Event)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetEventNextID()
	db.CreatedTime = time.Now()
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
		from := GetGraphNode(mine.UID)
		for i := 0; i < len(links); i += 1 {
			node := GetGraphNode(links[i].Entity)
			_, _ = CreateLink(from, node, LinkType(links[i].Category), links[i].Name, "", DirectionType(links[i].Direction))
		}
		concept := GetConceptByName("地理")
		_, _, _ = createSampleEntity(place.Name, concept.Name)
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

func (mine *EntityInfo) RemoveEvent(uid string) error {
	if mine.events == nil {
		return errors.New("must call construct fist")
	}
	if !mine.HadEvent(uid) {
		return errors.New("not found the event")
	}
	err := nosql.RemoveEvent(uid)
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

func (mine *EntityInfo) AddProperties(array []*proxy.PropertyInfo) error {
	err := nosql.UpdateEntityProperties(mine.table(), mine.UID, array)
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

//region Asset Fun
func (mine *EntityInfo) GetAssets() []*AssetInfo {
	return mine.assets
}

func (mine *EntityInfo) HadAsset(kind uint8, language string) bool {
	for i := 0; i < len(mine.assets); i += 1 {
		if mine.assets[i].Type == kind && mine.assets[i].Language == language {
			return true
		}
	}
	return false
}

func (mine *EntityInfo) AddAsset(uid string) bool {
	var info = new(AssetInfo)
	if !info.initAsset(uid) {
		return false
	}
	mine.assets = append(mine.assets, info)
	return true
}

func (mine *EntityInfo) AddAssetInfo(asset *AssetInfo) bool {
	if asset == nil {
		return false
	}
	mine.assets = append(mine.assets, asset)
	return true
}

func (mine *EntityInfo) GetAssetByFilter(kind uint8, language string) *AssetInfo {
	for i := 0; i < len(mine.assets); i += 1 {
		if mine.assets[i].Type == kind && mine.assets[i].Language == language {
			return mine.assets[i]
		}
	}
	return nil
}

func (mine *EntityInfo) GetAssetByType(kind uint8) *AssetInfo {
	for i := 0; i < len(mine.assets); i += 1 {
		if mine.assets[i].Type == kind {
			return mine.assets[i]
		}
	}
	return nil
}

func (mine *EntityInfo) GetAsset(uid string) *AssetInfo {
	for i := 0; i < len(mine.assets); i += 1 {
		if mine.assets[i].UID == uid {
			return mine.assets[i]
		}
	}
	return nil
}

func (mine *EntityInfo) RemoveAsset(uid string) error {
	err := nosql.RemoveAsset(uid)
	if err == nil {
		length := len(mine.assets)
		for i := 0; i < length; i++ {
			if mine.assets[i].UID == uid {
				mine.assets = append(mine.assets[:i], mine.assets[i+1:]...)
				break
			}
		}
	}
	return err
}

//endregion
