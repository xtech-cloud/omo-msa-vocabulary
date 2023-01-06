package cache

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
	"omo.msa.vocabulary/tool"
	"strings"
	"time"
)

//region Global Entity

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
	db.Pushed = 0
	db.Synonyms = info.Synonyms
	db.Events = info.StaticEvents
	db.Relations = info.StaticRelations
	db.Relates = info.Relates
	info.events = make([]*EventInfo, 0, 1)
	if info.Properties == nil {
		info.Properties = make([]*proxy.PropertyInfo, 0, 1)
	}
	db.FirstLetter = firstLetter(info.Name)
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
					} else {
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
					} else {
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

func (mine *cacheContext) GetEntitiesByRelate(relate string) []*EntityInfo {
	list := make([]*EntityInfo, 0, 50)
	dbs, err := nosql.GetRecordsByRelate(relate, uint8(OptionSwitch))
	if err != nil {
		return list
	}
	arr := make([]string, 0, 50)
	for _, db := range dbs {
		if !tool.HasItem(arr, db.Entity) {
			arr = append(arr, db.Entity)
			info := mine.GetEntity(db.Entity)
			if info != nil {
				list = append(list, info)
			}
		}
	}
	return list
}

func (mine *cacheContext) GetEntityCountByRelate(relate string) uint32 {
	dbs, err := nosql.GetRecordsByRelate(relate, uint8(OptionSwitch))
	if err != nil {
		return 0
	}
	arr := make([]string, 0, 50)
	for _, db := range dbs {
		if !tool.HasItem(arr, db.Entity) {
			arr = append(arr, db.Entity)
		}
	}
	return uint32(len(arr))
}

func (mine *cacheContext) GetUserEntityByLetter(relate, first string) []*EntityInfo {
	list := make([]*EntityInfo, 0, 10)
	dbs, err := nosql.GetEntityByFirstLetter(UserEntityTable, relate, strings.ToUpper(first))
	if err == nil {
		for _, db := range dbs {
			info := new(EntityInfo)
			info.initInfo(db)
			list = append(list, info)
		}
	}
	return list
}

func (mine *cacheContext) GetEntityCountByScene(scene string) uint32 {
	if len(scene) < 2 {
		return 0
	}
	count := nosql.GetEntityCountByScene(DefaultEntityTable, scene)
	count1 := nosql.GetEntityCountByScene(UserEntityTable, scene)
	count2 := nosql.GetEntityCountByScene(MuseumEntityTable, scene)
	return count + count1 + count2
}

func (mine *cacheContext) GetEntityCount() uint32 {
	count := nosql.GetEntityCount(DefaultEntityTable)
	count1 := nosql.GetEntityCount(UserEntityTable)
	count2 := nosql.GetEntityCount(MuseumEntityTable)
	return count + count1 + count2
}

//endregion

//region Global Events

func (mine *cacheContext) GetEvent(uid string) *EventInfo {
	event, err := nosql.GetEvent(uid)
	if err == nil && event != nil {
		info := new(EventInfo)
		info.initInfo(event)
		return info
	}

	return nil
}

func (mine *cacheContext) GetEventsByQuote(quote string) []*EventInfo {
	arr, err := nosql.GetEventsByQuote2(quote)
	var list []*EventInfo
	if err == nil {
		list = make([]*EventInfo, 0, len(arr))
		for _, db := range arr {
			info := new(EventInfo)
			info.initInfo(db)
			list = append(list, info)
		}
	} else {
		list = make([]*EventInfo, 0, 1)
	}

	return list
}

func (mine *cacheContext) GetEventsByEntity(entity string, tp uint8) []*EventInfo {
	arr, err := nosql.GetEventsByType(entity, tp)
	var list []*EventInfo
	if err == nil {
		list = make([]*EventInfo, 0, len(arr))
		for _, db := range arr {
			info := new(EventInfo)
			info.initInfo(db)
			list = append(list, info)
		}
	} else {
		list = make([]*EventInfo, 0, 1)
	}

	return list
}

func (mine *cacheContext) GetEventsCountByEntity(entity string) uint32 {
	if len(entity) < 2 {
		return 0
	}
	return nosql.GetEventCountByEntity(entity)
}

func (mine *cacheContext) GetEvents(entity string) []*EventInfo {
	arr, err := nosql.GetEventsByEntity(entity)
	var list []*EventInfo
	if err == nil {
		list = make([]*EventInfo, 0, len(arr))
		for _, db := range arr {
			info := new(EventInfo)
			info.initInfo(db)
			list = append(list, info)
		}
	} else {
		list = make([]*EventInfo, 0, 1)
	}

	return list
}

func (mine *cacheContext) RemoveEvent(uid, operator string) error {
	return nosql.RemoveEvent(uid, operator)
}

/*
  GetEventsByRelate 根据实体的关联信息获取事件列表, 并且要求实体的关联时间要晚于事件的创建时间
*/
func (mine *cacheContext) GetEventsByRelate(entity, relate string) []*EventInfo {
	list := make([]*EventInfo, 0, 50)
	dbs, err := nosql.GetRecordsBy(entity, relate, uint8(OptionSwitch))
	if err != nil || len(dbs) < 1 {
		return list
	}
	latest := dbs[len(dbs)-1]
	eveDBs, er := nosql.GetEventsByEntity(entity)
	if er == nil {
		for _, event := range eveDBs {
			if event.CreatedTime.Unix() < latest.CreatedTime.Unix() {
				info := new(EventInfo)
				info.initInfo(event)
				list = append(list, info)
			}
		}
	}
	return list
}

//endregion