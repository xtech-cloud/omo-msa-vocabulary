package cache

import (
	"errors"
	"omo.msa.vocabulary/proxy/nosql"
	"omo.msa.vocabulary/tool"
	"sort"
	"strings"
)

//region Global Entity

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

func (mine *cacheContext) MatchEntities(key string) []*EntityInfo {
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
	array2, err2 := nosql.GetEntitiesByMatch(UserEntityTable, key)
	if err2 == nil {
		for _, entity := range array2 {
			info := new(EntityInfo)
			info.initInfo(entity)
			list = append(list, info)
		}
	}
	return list
}

func (mine *cacheContext) SearchPersonalEntities(key string) []*EntityInfo {
	array, err := nosql.GetEntitiesByMatch(UserEntityTable, key)
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

func (mine *cacheContext) SearchDefaultEntities(owner, key string) []*EntityInfo {
	if owner == "" || key == "" {
		return make([]*EntityInfo, 0, 1)
	}
	array, err := nosql.GetEntitiesByOwnMatch(DefaultEntityTable, key, owner)
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

func (mine *cacheContext) HadEntityByName(name, add, owner string) bool {
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
		db1, err1 := nosql.GetEntitiesByOwnName(UserEntityTable, name, owner)
		if err1 == nil && db1 != nil {
			if len(db1) > 0 {
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

func (mine *cacheContext) GetEntitiesByOwner(owner string, page, num int32) (int32, int32, []*EntityInfo) {
	dbs := make([]*nosql.Entity, 0, 200)
	for _, tb := range mine.EntityTables() {
		array, err := nosql.GetEntitiesByOwner(tb, owner)
		if err == nil {
			dbs = append(dbs, array...)

		}
	}
	total, pages, arr := CheckPage(page, num, dbs)
	list := make([]*EntityInfo, 0, len(arr))
	for _, entity := range arr {
		info := new(EntityInfo)
		info.initInfo(entity)
		list = append(list, info)
	}
	return total, pages, list
}

func (mine *cacheContext) GetEntitiesByConcept(owner, concept string, page, num int32) (int32, int32, []*EntityInfo) {
	list := make([]*EntityInfo, 0, 100)
	for _, table := range mine.entityTables {
		array, err := nosql.GetEntitiesByConcept(table, owner, concept)
		if err != nil {
			return 0, 0, list
		}
		for _, entity := range array {
			info := new(EntityInfo)
			info.initInfo(entity)
			list = append(list, info)
		}
	}

	return CheckPage(page, num, list)
}

func (mine *cacheContext) GetEntitiesByConcept2(concept string) []*EntityInfo {
	list := make([]*EntityInfo, 0, 10)
	for _, table := range mine.entityTables {
		array, err := nosql.GetEntitiesByConcept2(table, concept)
		if err != nil {
			return list
		}
		for _, entity := range array {
			info := new(EntityInfo)
			info.initInfo(entity)
			list = append(list, info)
		}
	}

	return list
}

func (mine *cacheContext) GetEntitiesCountByConcept(tb, concept string) uint32 {
	return uint32(nosql.GetEntitiesCountByConcept(tb, concept))
}

func (mine *cacheContext) GetEntitiesCountByAttribute(attr string) int {
	num := 0
	for _, table := range mine.entityTables {
		array, _ := nosql.GetEntitiesByAttribute(table, attr)
		num += len(array)
	}

	return num
}

func (mine *cacheContext) getEntitiesByAttribute(uid string) []*EntityInfo {
	list := make([]*EntityInfo, 0, 500)
	for _, table := range mine.entityTables {
		array, _ := nosql.GetEntitiesByAttribute(table, uid)
		for _, db := range array {
			info := new(EntityInfo)
			info.initInfo(db)
			list = append(list, info)
		}
	}

	return list
}

func (mine *cacheContext) GetEntitiesByStatus(status EntityStatus, concept string) []*EntityInfo {
	list := make([]*EntityInfo, 0, 100)
	if status == EntityStatusUsable {
		return mine.GetArchivedEntities(DefaultOwner, concept)
	} else {
		for _, tb := range mine.EntityTables() {
			var array []*nosql.Entity
			var err error
			if status == EntityStatusAll {
				array, err = nosql.GetEntitiesByConcept2(tb, concept)
			} else {
				array, err = nosql.GetEntitiesByStatus(tb, uint8(status))
			}

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
			var array []*nosql.Entity
			var err error
			if status == EntityStatusAll {
				array, err = nosql.GetEntitiesByOwner(tb, owner)
			} else {
				array, err = nosql.GetEntitiesByOwnerAndStatus(tb, owner, uint8(status))
			}

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
	db := mine.getEntityFromDB(uid)
	if db != nil {
		info := new(EntityInfo)
		info.initInfo(db)
		return info
	}
	return nil
}

func (mine *cacheContext) GetEntityByProp(name, key, val string) *EntityInfo {
	array, err := nosql.GetEntitiesByProp(UserEntityTable, key, val)
	if err == nil {
		for _, entity := range array {
			if entity.Name == name {
				info := new(EntityInfo)
				info.initInfo(entity)
				return info
			}
		}
	}
	return nil
}

func (mine *cacheContext) GetEntitiesByRegex(key, val string, page, num int32) (int32, int32, []*EntityInfo, error) {
	list := make([]*EntityInfo, 0, 200)
	all := make([]*nosql.Entity, 0, 200)
	for _, table := range mine.entityTables {
		dbs, err := nosql.GetEntitiesByRegex(table, key, val)
		if err == nil {
			all = append(all, dbs...)
		}
	}

	if len(all) > 100 && page < 1 {
		page = 1
		num = 100
	}
	max, pages, arr := CheckPage(page, num, all)
	for _, db := range arr {
		info := new(EntityInfo)
		info.initInfo(db)
		list = append(list, info)
	}
	return max, pages, list, nil
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

func (mine *cacheContext) GetEntitiesByRelate(relate string, page, num int32) (int32, int32, []*EntityInfo) {
	list := make([]*EntityInfo, 0, 200)
	dbs, err := nosql.GetRecordsByRelate(relate, uint8(OptionSwitch))
	if err != nil {
		return 0, 0, list
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

	return CheckPage(page, num, list)
}

func (mine *cacheContext) GetEntitiesByRank(relate string, num int) []*EntityInfo {
	list := make([]*EntityInfo, 0, num)
	dbs, err := nosql.GetRecordsByRelate(relate, uint8(OptionSwitch))
	if err != nil {
		return list
	}
	arr := make([]string, 0, 50)
	pairs := make([]*PairInfo, 0, 50)
	for _, db := range dbs {
		if !tool.HasItem(arr, db.Entity) {
			arr = append(arr, db.Entity)
			events := mine.GetEvents(db.Entity)
			pairs = append(pairs, &PairInfo{Key: db.Entity, Count: int32(len(events))})
		}
	}

	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	for i := 0; i < len(pairs); i += 1 {
		if i < num {
			entity := mine.GetEntity(pairs[i].Key)
			if entity != nil {
				entity.Score = uint32(pairs[i].Count)
				list = append(list, entity)
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

func (mine *cacheContext) GetUserEntitiesByLetter(relate, first string) []*EntityInfo {
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

func (mine *cacheContext) GetUserEntitiesByLetters(relate, letters string) []*EntityInfo {
	list := make([]*EntityInfo, 0, 30)
	dbs, err := nosql.GetEntityByFirstLetter(UserEntityTable, relate, strings.ToUpper(letters))
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
	return count + count1
}

func (mine *cacheContext) GetEntityCount() uint32 {
	count := nosql.GetEntityCount(DefaultEntityTable)
	count1 := nosql.GetEntityCount(UserEntityTable)
	return count + count1
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

func (mine *cacheContext) GetEventByAsset(uid string) *EventInfo {
	event, err := nosql.GetEventByAsset(uid)
	if err == nil && event != nil {
		info := new(EventInfo)
		info.initInfo(event)
		return info
	}

	return nil
}

func (mine *cacheContext) GetEventsByQuote(quote string) []*EventInfo {
	if quote == "" {
		return make([]*EventInfo, 0, 1)
	}
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

func (mine *cacheContext) GetEventByTarget(entity, target string, tp uint8) (*EventInfo, error) {
	if entity == "" || target == "" {
		return nil, errors.New("the entity or target is empty")
	}
	db, err := nosql.GetEventByTarget(entity, target)
	if err != nil {
		return nil, err
	}
	info := new(EventInfo)
	info.initInfo(db)
	return info, nil
}

func hadEvent(list []*EventInfo, uid string) bool {
	for _, item := range list {
		if item.UID == uid {
			return true
		}
	}
	return false
}

func (mine *cacheContext) GetEventsAssetsByQuote(quote string, page, number int32) (int32, int32, []*EventInfo) {
	all, err := nosql.GetEventsByQuote2(quote)
	var list []*EventInfo
	var total int32
	var pages int32
	if err == nil {
		list = make([]*EventInfo, 0, len(all))
		assets := make([]string, 0, len(all)*3)
		for _, db := range all {
			assets = append(assets, db.Assets...)
		}
		var arr []string
		total, pages, arr = CheckPage(page, number, assets)
		for _, uid := range arr {
			eve := getEventByAsset(all, uid)
			if eve != nil && !hadEvent(list, eve.UID.Hex()) {
				info := new(EventInfo)
				info.initInfo(eve)
				list = append(list, info)
			}
		}
	} else {
		list = make([]*EventInfo, 0, 1)
	}

	return total, pages, list
}

func getEventByAsset(list []*nosql.Event, uid string) *nosql.Event {
	for _, event := range list {
		if tool.HasItem(event.Assets, uid) {
			return event
		}
	}
	return nil
}

func (mine *cacheContext) GetEventsByQuotePage(quote string, page, number int32) (int32, int32, []*EventInfo) {
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

	return CheckPage(page, number, list)
}

func (mine *cacheContext) GetEventsByDuration(quote string, from, end int64) []*EventInfo {
	arr, err := nosql.GetEventsByDuration(quote, from, end)
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

func (mine *cacheContext) GetEventsByRegex(quote, key, value string) []*EventInfo {
	arr, err := nosql.GetEventsByRegex(quote, key, value)
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

func (mine *cacheContext) GetEventsByEntityTarget(entity, target string) []*EventInfo {
	if target == "" {
		return nil
	}
	var dbs []*nosql.Event
	var err error
	if len(entity) > 0 {
		dbs, err = nosql.GetEventsByEntityTarget(entity, target)
	} else {
		dbs, err = nosql.GetEventsByTarget(target)
	}

	var list []*EventInfo
	if err == nil {
		list = make([]*EventInfo, 0, len(dbs))
		for _, db := range dbs {
			info := new(EventInfo)
			info.initInfo(db)
			list = append(list, info)
		}
	} else {
		list = make([]*EventInfo, 0, 1)
	}

	return list
}

func (mine *cacheContext) GetEventsBySceneTarget(owner, target string) []*EventInfo {
	if owner == "" || target == "" {
		return nil
	}
	dbs, err := nosql.GetEventsByOwnerTarget(owner, target)
	var list []*EventInfo
	if err == nil {
		list = make([]*EventInfo, 0, len(dbs))
		for _, db := range dbs {
			info := new(EventInfo)
			info.initInfo(db)
			list = append(list, info)
		}
	} else {
		list = make([]*EventInfo, 0, 1)
	}

	return list
}

func (mine *cacheContext) GetEventAssetCountBySceneTarget(owner, target string) uint32 {
	dbs, _ := nosql.GetEventsByOwnerTarget(owner, target)
	list := make([]string, 0, len(dbs)*2)
	for _, db := range dbs {
		for _, asset := range db.Assets {
			if !tool.HasItem(list, asset) {
				list = append(list, asset)
			}
		}
	}
	return uint32(len(list))
}

func (mine *cacheContext) GetEventCountBySceneTarget(owner, target string) uint32 {
	dbs, _ := nosql.GetEventsByOwnerTarget(owner, target)
	return uint32(len(dbs))
}

func (mine *cacheContext) GetAllSystemEvents(page, number int32) (int32, int32, []*EventInfo) {
	arr, err := nosql.GetEventsAllByType(1)
	var list []*EventInfo
	if err == nil {
		list = make([]*EventInfo, 0, len(arr))
		max, pages, dbs := CheckPage(page, number, arr)
		for _, db := range dbs {
			info := new(EventInfo)
			info.initInfo(db)
			list = append(list, info)
		}
		return max, pages, list
	} else {
		list = make([]*EventInfo, 0, 1)
		return 0, 0, list
	}
}

func (mine *cacheContext) GetEventsByWeek(from int64, quotes []string) []*EventInfo {
	end := from + 6*24*3600
	list := make([]*EventInfo, 0, 100)
	for _, quote := range quotes {
		arr, err := nosql.GetEventsByQuote2(quote)
		if err == nil {
			for _, db := range arr {
				now := db.Created
				if now > from && now < end {
					info := new(EventInfo)
					info.initInfo(db)
					list = append(list, info)
				}
			}
		} else {

		}
	}
	return list
}

func (mine *cacheContext) GetEventsByEntityType(entity string, tp, page, number int32) (int32, int32, []*EventInfo) {
	arr, err := nosql.GetEventsByType(entity, uint8(tp))
	var list []*EventInfo
	if err == nil {
		list = make([]*EventInfo, 0, len(arr))
		for _, db := range arr {
			info := new(EventInfo)
			info.initInfo(db)
			list = append(list, info)
		}
		return CheckPage(page, number, list)
	} else {
		return 0, 0, make([]*EventInfo, 0, 1)
	}
}

func (mine *cacheContext) GetEventsByEntity(entity, quote string, tp uint8) []*EventInfo {
	if entity == "" {
		return make([]*EventInfo, 0, 1)
	}
	var arr []*nosql.Event
	var err error
	if quote == "" {
		arr, err = nosql.GetEventsByType(entity, tp)
	} else {
		arr, err = nosql.GetEventsByQuoteType(entity, quote, tp)
	}

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
			if event.Created < latest.Created {
				info := new(EventInfo)
				info.initInfo(event)
				list = append(list, info)
			}
		}
	}
	return list
}

//endregion
