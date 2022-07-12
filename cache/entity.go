package cache

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

const (
	EntityStatusDraft   EntityStatus = 0
	EntityStatusFirst   EntityStatus = 1
	EntityStatusPending EntityStatus = 2
	EntityStatusSpecial EntityStatus = 3
	EntityStatusUsable  EntityStatus = 4 //审核通过
	EntityStatusFailed  EntityStatus = 10
)

const (
	DefaultEntityTable = "entities"
	SchoolEntityTable = "entities_school"
)

const (
	OptionAgree OptionType = 1
	OptionRefuse OptionType = 2

)
type EntityStatus uint8

type OptionType uint8

type EntityInfo struct {
	Status EntityStatus `json:"_"`
	BaseInfo
	Concept     string `json:"concept"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Cover       string `json:"cover"`
	Add         string `json:"add"` //消歧义
	Owner           string                    `json:"owner"`    //所属单位
	Mark            string                    `json:"mark"`     // 标记或者来源
	Quote           string                    `json:"quote"`    // 引用
	Synonyms        []string                  `json:"synonyms"` //同义词
	Tags            []string                  `json:"tags"`     //标签
	Published       bool                      `json:"published"`
	Properties      []*proxy.PropertyInfo     `json:"properties"`
	StaticEvents    []*proxy.EventBrief       `json:"events"`
	StaticRelations []*proxy.RelationCaseInfo `json:"relations"`
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

func (mine *cacheContext) CreateEntity(info *EntityInfo) error {
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

func (mine *cacheContext) syncGraphNode(info *EntityInfo) {
	var name = info.Name
	if info.Add != "" {
		name = info.Name + "-" + info.Add
	}
	mine.addSyncNode(info.UID, name, info.Concept, info.Cover)
}

func (mine *cacheContext) AllEntities() []*EntityInfo {
	list := make([]*EntityInfo, 0, 200)
	for _, tb := range mine.EntityTables() {
		array, err := nosql.GetEntities(tb)
		if err == nil {
			for _, entity := range array {
				info := new(EntityInfo)
				info.initInfo(entity)
				list = append(list, info)
			}
		}
	}
	return list
}

func (mine *cacheContext) SearchEntities(key string) []*EntityInfo {
	array, err := nosql.GetEntitiesByMatch(DefaultEntityTable, key)
	if err != nil {
		return make([]*EntityInfo, 0, 0)
	}
	list := make([]*EntityInfo, 0, len(array))
	for _, entity := range array {
		info := new(EntityInfo)
		info.initInfo(entity)
		list = append(list, info)
	}
	return list
}

func (mine *cacheContext) HadEntityByName(name, add string) bool {
	if len(name) < 1 {
		return true
	}
	if len(add) > 0 {
		info := mine.GetEntityByName(name, add)
		if info != nil {
			return true
		} else {
			return false
		}
	} else {
		db, err := nosql.GetEntitiesByName(DefaultEntityTable, name)
		if err == nil && db != nil {
			if len(db) > 0 {
				return true
			}
		}
		return false
	}
}

func (mine *cacheContext) HadEntityByMark(mark string) bool {
	info := mine.GetEntityByMark(mark)
	if info != nil {
		return true
	}
	return false
}

func (mine *cacheContext) GetEntityByName(name, add string) *EntityInfo {
	if len(name) < 1 {
		return nil
	}

	for _, tb := range mine.EntityTables() {
		db, err := nosql.GetEntityByName(tb, name, add)
		if err == nil && db != nil {
			info := new(EntityInfo)
			info.initInfo(db)
			return info
		}
	}
	return nil
}

func (mine *cacheContext) GetEntityByMark(mark string) *EntityInfo {
	if len(mark) < 1 {
		return nil
	}
	db, err := nosql.GetEntityByMark(DefaultEntityTable, mark)
	if err == nil && db != nil {
		info := new(EntityInfo)
		info.initInfo(db)
		return info
	}

	return nil
}

func (mine *cacheContext) GetEntitiesByOwner(owner string) []*EntityInfo {
	list := make([]*EntityInfo, 0, 30)
	for _, tb := range mine.EntityTables() {
		array, err := nosql.GetEntitiesByOwner(tb, owner)
		if err == nil {
			for _, entity := range array {
				info := new(EntityInfo)
				info.initInfo(entity)
				list = append(list, info)
			}
		}
	}
	return list
}

func (mine *cacheContext) GetEntitiesByConcept(concept string) []*EntityInfo {
	list := make([]*EntityInfo, 0, 10)
	array, err := nosql.GetEntitiesByConcept(DefaultEntityTable, concept)
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

func (mine *cacheContext) GetEntitiesByStatus(status EntityStatus, concept string) []*EntityInfo {
	list := make([]*EntityInfo, 0, 100)
	if status == EntityStatusUsable {
		return mine.GetArchivedEntities("", concept)
	} else {
		for _, tb := range mine.EntityTables() {
			array, err := nosql.GetEntitiesByStatus(tb, uint8(status))
			if err == nil {
				for _, entity := range array {
					if concept != "" {
						if entity.Concept == concept {
							info := new(EntityInfo)
							info.initInfo(entity)
							list = append(list, info)
						}
					}else{
						info := new(EntityInfo)
						info.initInfo(entity)
						list = append(list, info)
					}
				}
			}
		}
	}
	return list
}

func (mine *cacheContext) GetEntitiesByOwnerStatus(owner, concept string, status EntityStatus) []*EntityInfo {
	list := make([]*EntityInfo, 0, 50)
	if status == EntityStatusUsable {
		return mine.GetArchivedEntities(owner, concept)
	} else {
		for _, tb := range mine.EntityTables() {
			array, err := nosql.GetEntitiesByOwnerAndStatus(tb, owner, uint8(status))
			if err == nil {
				for _, entity := range array {
					if concept != "" {
						if entity.Concept == concept {
							info := new(EntityInfo)
							info.initInfo(entity)
							list = append(list, info)
						}
					}else{
						info := new(EntityInfo)
						info.initInfo(entity)
						list = append(list, info)
					}
				}
			}
		}
	}
	return list
}

func (mine *cacheContext) GetEntitiesByProp(key, val string) []*EntityInfo {
	list := make([]*EntityInfo, 0, 10)
	for _, tb := range mine.EntityTables() {
		array, err := nosql.GetEntitiesByProp(tb, key, val)
		if err == nil {
			for _, entity := range array {
				info := new(EntityInfo)
				info.initInfo(entity)
				list = append(list, info)
			}
		}
	}
	return list
}

func (mine *cacheContext) GetEntity(uid string) *EntityInfo {
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

func (mine *cacheContext) GetEntitiesByList(st EntityStatus, array []string) ([]*EntityInfo, error) {
	if array == nil || len(array) < 1 {
		return nil, errors.New("the list is empty")
	}
	if st == EntityStatusUsable {
		return mine.GetArchivedByList(array)
	} else {
		list := make([]*EntityInfo, 0, len(array))
		for _, item := range array {
			info := mine.GetEntity(item)
			if info != nil {
				list = append(list, info)
			}
		}
		return list, nil
	}
}

func (mine *cacheContext) GetCustomEntitiesByList(array []string) ([]*EntityInfo, error) {
	if array == nil || len(array) < 1 {
		return nil, errors.New("the list is empty")
	}
	list := make([]*EntityInfo, 0, len(array))
	for _, item := range array {
		info := mine.GetEntity(item)
		if info != nil {
			list = append(list, info)
		}
	}
	return list, nil
}

func (mine *cacheContext) getEntityFromDB(uid string) *nosql.Entity {
	//db, err := nosql.GetEntity(DefaultEntityTable, uid)
	//if err == nil && db != nil {
	//	return db
	//}
	//logger.Error("getEntityFromDB in entities that error =" + err.Error())
	for _, tb := range mine.EntityTables() {
		db1, er := nosql.GetEntity(tb, uid)
		if er == nil && db1 != nil {
			return db1
		}
	}

	return nil
}

func (mine *cacheContext) HadEntity(uid string) bool {
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

func (mine *cacheContext) RemoveEntity(uid, operator string) error {
	if len(uid) < 1 {
		return errors.New("the entity uid is empty")
	}
	tmp := mine.GetEntity(uid)
	if tmp == nil {
		return nil
	}
	if tmp.Status != EntityStatusDraft {
		return errors.New("the entity status not equal 0 ")
	}

	err := nosql.RemoveEntity(tmp.table(), uid, operator)
	if err == nil {
		t, _ := nosql.GetArchivedByEntity(uid)
		if t != nil {
			_ = nosql.RemoveArchived(t.UID.Hex(), operator)
			//return errors.New("the entity had published")
		}
		mine.checkEntityFromBoxes(uid, tmp.Name)
	}
	return err
}

func (mine *cacheContext) HadOwnerOfAsset(owner string) bool {
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
	if cacheCtx.HadArchivedByEntity(mine.UID) {
		mine.Published = true
	} else {
		mine.Published = false
	}

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
	//if strings.Contains(mine.Cover,"http://rdp-down.suii.cn/") {
	//	cover := strings.Replace(mine.Cover, "http://rdp-down.suii.cn/", "", 1)
	//	_ = mine.setCover(cover, mine.Operator)
	//}
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

func (mine *EntityInfo)updateConcept(concept, operator string) error {
	//if mine.Status != EntityStatusDraft {
	//	return errors.New("the entity is not draft so can not update")
	//}
	if mine.Concept != concept {
		err := nosql.UpdateEntityConcept(mine.table(), mine.UID, concept, operator)
		if err == nil {
			mine.Concept = concept
			mine.Operator = operator
		}
		return err
	} else{
		return nil
	}
}

func (mine *EntityInfo) UpdateBase(name, desc, add, concept, cover, mark, quote, sum, operator string) error {
	if mine.Status != EntityStatusDraft {
		return errors.New("the entity is not draft so can not update")
	}
	if concept == "" {
		concept = mine.Concept
	}
	if desc == "" {
		desc = mine.Description
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
	if desc != mine.Description || sum != mine.Summary {
		err = nosql.UpdateEntityRemark(mine.table(), mine.UID, desc, sum, operator)
		if err == nil {
			mine.Description = desc
			mine.Summary = sum
			mine.Operator = operator
			mine.UpdateTime = time.Now()
		}
	}
	if name != mine.Name || add != mine.Add || concept != mine.Concept || quote != mine.Quote {
		err = nosql.UpdateEntityBase(mine.table(), mine.UID, name, add, concept, quote, mark, operator)
		if err == nil {
			mine.Name = name
			mine.Add = add
			mine.Quote = quote
			mine.Concept = concept
			mine.Mark = mark
			mine.Operator = operator
			mine.UpdateTime = time.Now()
		}
	}
	return err
}

func (mine *EntityInfo) UpdateStatic(info *EntityInfo) error {
	if mine.Status != EntityStatusDraft {
		return errors.New("the entity is not draft so can not update")
	}
	_ = mine.UpdateBase(info.Name, info.Description, info.Add, info.Concept, info.Cover, info.Mark, info.Quote, info.Summary, info.Operator)
	err := nosql.UpdateEntityStatic(mine.table(), mine.UID, info.Operator, info.Tags, info.Properties)
	if err == nil {
		mine.Tags = info.Tags
		mine.Properties = info.Properties
		mine.UpdateTime = time.Now()
	}
	if len(info.StaticEvents) > 0 {
		_ = mine.UpdateStaticEvents(info.Operator, info.StaticEvents)
	}
	if len(info.StaticRelations) > 0 {
		_ = mine.UpdateStaticRelations(info.Operator, info.StaticRelations)
	}
	return err
}

func (mine *EntityInfo) UpdateStaticEvents(operator string, events []*proxy.EventBrief) error {
	if mine.Status != EntityStatusDraft {
		return errors.New("the entity is not draft so can not update")
	}
	err := nosql.UpdateEntityEvents(mine.table(), mine.UID, operator, events)
	if err == nil {
		mine.Operator = operator
		mine.StaticEvents = events
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EntityInfo) UpdateStaticRelations(operator string, list []*proxy.RelationCaseInfo) error {
	if mine.Status != EntityStatusDraft {
		return errors.New("the entity is not draft so can not update")
	}
	err := nosql.UpdateEntityRelations(mine.table(), mine.UID, operator, list)
	if err == nil {
		mine.Operator = operator
		mine.StaticRelations = list
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EntityInfo) UpdateCover(cover, operator string) error {
	if mine.Status == EntityStatusUsable {
		return errors.New("the entity had published so can not update")
	}

	if cover == "" || cover == mine.Cover {
		return nil
	}
	err := nosql.UpdateEntityCover(mine.table(), mine.UID, cover, operator)
	if err == nil {
		mine.Cover = cover
		mine.Operator = operator
		mine.UpdateTime = time.Now()
		//go Context().graph.UpdateNodeCover(mine.UID, cover)
	}
	return err
}

func (mine *EntityInfo) setCover(cover, operator string) error {
	if cover == "" || cover == mine.Cover {
		return nil
	}
	err := nosql.UpdateEntityCover(mine.table(), mine.UID, cover, operator)
	if err == nil {
		mine.Cover = cover
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EntityInfo)GetRecords() ([]*nosql.Record,error) {
	return nosql.GetRecords(mine.UID)
}

func (mine *EntityInfo) UpdateTags(tags []string, operator string) error {
	if mine.Status == EntityStatusUsable {
		return errors.New("the entity had published so can not update")
	}
	err := nosql.UpdateEntityTags(mine.table(), mine.UID, operator, tags)
	if err == nil {
		mine.Tags = tags
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EntityInfo) UpdateSynonyms(list []string, operator string) error {
	if mine.Status == EntityStatusUsable {
		return errors.New("the entity had published so can not update")
	}
	err := nosql.UpdateEntitySynonyms(mine.table(), mine.UID, operator, list)
	if err == nil {
		mine.Synonyms = list
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *EntityInfo)createRecord(operator, remark string, from, to EntityStatus)  {
	db := new(nosql.Record)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetRecordNextID()
	db.Creator = operator
	db.Entity = mine.UID
	if to > from {
		db.Option = uint8(OptionAgree)
	}else{
		db.Option = uint8(OptionRefuse)
	}

	db.Remark = remark
	_ = nosql.CreateRecord(db)
}

func (mine *EntityInfo) UpdateStatus(status EntityStatus, operator, remark string) error {
	if mine.Status == status {
		return nil
	}
	err := nosql.UpdateEntityStatus(mine.table(), mine.UID, uint8(status), operator)
	if err != nil {
		return err
	}
	mine.Operator = operator
	mine.createRecord(operator, remark, mine.Status, status)
	if status == EntityStatusUsable {
		tmp := Context().GetArchivedByEntity(mine.UID)
		if tmp == nil {
			err = Context().CreateArchived(mine)
			if err != nil {
				return err
			}
			cacheCtx.checkRelations(nil, mine)
		} else {
			old := tmp.GetEntity()
			err = tmp.UpdateFile(mine, operator)
			if err != nil {
				return err
			}
			cacheCtx.checkRelations(old, mine)
		}
	}
	mine.Status = status
	mine.UpdateTime = time.Now()
	return nil
}

//region Event Fun
func (mine *EntityInfo) initEvents() {
	if mine.events != nil {
		return
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
		mine.events = make([]*EventInfo, 0, 1)
	}
}

func (mine *EntityInfo) AllEvents() []*EventInfo {
	mine.initEvents()
	return mine.events
}

func (mine *EntityInfo) GetEventsByType(tp uint8, quote string) []*EventInfo {
	var err error
	var arr []*nosql.Event
	if len(quote) > 1 {
		arr,err = nosql.GetEventsByTypeQuote(mine.UID, quote, tp)
	}else{
		arr,err = nosql.GetEventsByType(mine.UID, tp)
	}

	var list []*EventInfo
	if err == nil {
		list = make([]*EventInfo, 0, len(arr))
		for _, db := range arr {
			info := new(EventInfo)
			info.initInfo(db)
			list = append(list, info)
		}
	}else{
		list = make([]*EventInfo, 0, 1)
	}

	return list
}

func (mine *EntityInfo) GetEventsByQuote(quote string) []*EventInfo {
	arr,err := nosql.GetEventsByQuote(mine.UID, quote)
	var list []*EventInfo
	if err == nil {
		list = make([]*EventInfo, 0, len(arr))
		for _, db := range arr {
			info := new(EventInfo)
			info.initInfo(db)
			list = append(list, info)
		}
	}else{
		list = make([]*EventInfo, 0, 1)
	}

	return list
}

func (mine *EntityInfo) GetEventsByAccess(tp, access uint8) []*EventInfo {
	var arr []*nosql.Event
	var err error
	if tp > 0 {
		arr,err = nosql.GetEventsByTypeAndAccess(mine.UID,tp, access)
	}else{
		arr,err = nosql.GetEventsByAccess(mine.UID, access)
	}

	var list []*EventInfo
	if err == nil {
		list = make([]*EventInfo, 0, len(arr))
		for _, db := range arr {
			info := new(EventInfo)
			info.initInfo(db)
			list = append(list, info)
		}
	}else{
		list = make([]*EventInfo, 0, 1)
	}

	return list
}

func (mine *EntityInfo) AddEvent(date proxy.DateInfo, place proxy.PlaceInfo, name, desc, cover, quote, operator string, tp, access uint8, links []proxy.RelationCaseInfo, tags, assets []string) (*EventInfo, error) {
	if mine.Status == EntityStatusUsable {
		return nil, errors.New("the entity had published so can not update")
	}
	mine.initEvents()

	db := new(nosql.Event)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetEventNextID()
	db.CreatedTime = time.Now()
	db.Creator = operator
	db.Name = name
	db.Date = date
	db.Place = place
	db.Type = tp
	db.Entity = mine.UID
	db.Quote = quote
	db.Description = desc
	db.Relations = links
	db.Cover = cover
	db.Tags = tags
	db.Access = access
	if db.Tags == nil {
		db.Tags = make([]string, 0, 1)
	}
	db.Assets = assets
	if db.Assets == nil {
		db.Assets = make([]string, 0, 1)
	}
	err := nosql.CreateEvent(db)
	if err == nil {
		mine.initEvents()
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
	mine.initEvents()
	for i := 0; i < len(mine.events); i += 1 {
		if mine.events[i].UID == uid {
			return true
		}
	}
	return false
}

func (mine *EntityInfo) HadEventBy(time, place string) bool {
	mine.initEvents()
	for _, event := range mine.events {
		if event.Date.Begin.String() == time && event.Place.Name == place {
			return true
		}
	}
	return false
}

func (mine *EntityInfo) GetEventBy(time, place string) *EventInfo {
	mine.initEvents()
	for _, event := range mine.events {
		if event.Date.Begin.String() == time && event.Place.Name == place {
			return event
		}
	}
	return nil
}

func (mine *EntityInfo) RemoveEvent(uid, operator string) error {
	if mine.Status == EntityStatusUsable {
		return errors.New("the entity had published so can not update")
	}
	mine.initEvents()
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
	mine.initEvents()
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
	if mine.Status != EntityStatusDraft {
		return errors.New("the entity is not draft so can not update")
	}
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
	if mine.Status != EntityStatusDraft {
		return errors.New("the entity is not draft so can not update")
	}
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
	if mine.Status == EntityStatusUsable {
		return errors.New("the entity had published so can not update")
	}
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
