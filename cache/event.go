package cache

import (
	"fmt"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
	"omo.msa.vocabulary/tool"
	"sort"
	"strconv"
	"time"
)

const (
	EventCustom   = 0 //普通事件
	EventActivity = 1 //活动
	EventHonor    = 2 //荣誉
	EventCert     = 3 //证书
	EventSpec     = 4 //特殊事件
)

const (
	AccessRead    = 0 //可读
	AccessPrivate = 1 //
	AccessWR      = 2 //可读写
)

const (
	SubEventCommon     EventSubtype = 0 //通用
	SubEventRecitation EventSubtype = 1 //朗诵
	SubEventReading    EventSubtype = 2 //阅读
	SubEventPlace      EventSubtype = 3 //实践地点
	SubEventCert       EventSubtype = 4 //证书
	SubEventWords      EventSubtype = 5 //留言
)

type EventSubtype uint8

type EventInfo struct {
	Type    uint8
	Access  uint8 // 可访问对象
	Subtype uint8 //子类型
	BaseInfo
	Description string // 描述
	Entity      string //实体对象
	Parent      string //父级事件
	Cover       string //封面
	Quote       string // 引用或者备注，活动，超链接
	Owner       string //所属场景或者组织机构
	Certify     string //证书实例

	Date      proxy.DateInfo
	Place     proxy.PlaceInfo
	Targets   []string //关联的其他实体对象
	Tags      []string
	Assets    []string
	Relations []proxy.RelationCaseInfo
	Labels    []proxy.PairInfo
}

func (mine *cacheContext) GetActivityCountBy(arr []string, date time.Time) []*pb.StatisticInfo {
	list := make([]*pb.StatisticInfo, 0, len(arr))
	for _, item := range arr {
		num := mine.GetActivityCountByDate(item, date)
		list = append(list, &pb.StatisticInfo{Key: item, Count: uint32(num)})
	}
	return list
}

func (mine *cacheContext) GetEventCountBy(entity string) []*pb.StatisticInfo {
	list := make([]*pb.StatisticInfo, 0, 3)
	arr := mine.GetEvents(entity)
	expCount := new(pb.StatisticInfo)
	expCount.Key = strconv.Itoa(EventCustom)
	honorCount := new(pb.StatisticInfo)
	honorCount.Key = strconv.Itoa(EventHonor)
	actCount := new(pb.StatisticInfo)
	actCount.Key = strconv.Itoa(EventActivity)
	for _, item := range arr {
		if item.Type == EventCustom {
			expCount.Count = expCount.Count + 1
		} else if item.Type == EventActivity {
			actCount.Count = actCount.Count + 1
		} else {
			honorCount.Count = honorCount.Count + 1
		}
	}
	list = append(list, expCount)
	list = append(list, honorCount)
	list = append(list, actCount)
	return list
}

func (mine *cacheContext) GetActivityCountByDate(entity string, date time.Time) int {
	count := 0
	dbs, err := nosql.GetEventsByType(entity, EventActivity)
	if err != nil {
		return count
	}
	for _, db := range dbs {
		utc := time.Unix(db.Created, 0)
		if utc.Month() == date.Month() {
			count += 1
		}
	}
	return count
}

func (mine *cacheContext) GetEventCountByQuote(quote string) uint32 {
	count := nosql.GetEventCountByQuote(quote)
	return count
}

func (mine *cacheContext) GetEventEntityCountByQuote(quote string) uint32 {
	dbs, _ := nosql.GetEventsByQuote2(quote)
	arr := make([]string, 0, 100)
	for _, db := range dbs {
		if !tool.HasItem(arr, db.Entity) {
			arr = append(arr, db.Entity)
		}
	}
	return uint32(len(arr))
}

func (mine *cacheContext) GetEventCountByEntityTarget(scene, target string, entities []string) uint32 {
	var count uint32
	for _, entity := range entities {
		if len(scene) > 0 {
			num := nosql.GetEventCountByEntityTarget2(scene, entity, target)
			count += num
		} else {
			num := nosql.GetEventCountByEntityTarget(entity, target)
			count += num
		}
	}

	return count
}

func (mine *cacheContext) GetEventRanks(owner string, num int) []*PairInfo {
	dbs, _ := nosql.GetEventsByOwner(owner)
	list := make([]*PairInfo, 0, len(dbs))
	getPair := func(key string, arr []*PairInfo) *PairInfo {
		for _, info := range arr {
			if info.Key == key {
				return info
			}
		}
		return nil
	}
	for _, db := range dbs {
		pair := getPair(db.Entity, list)
		if pair == nil {
			pair = &PairInfo{Key: db.Entity, Count: 1}
			list = append(list, pair)
		} else {
			pair.Count += 1
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Count > list[j].Count
	})
	if len(list) < num {
		return list
	} else {
		return list[:num]
	}
}

func (mine *cacheContext) GetEventAssetsByQuote(quote string) []string {
	assets := make([]string, 0, 10)
	dbs, err := nosql.GetEventsByQuote2(quote)
	if err != nil {
		return assets
	}
	for _, db := range dbs {
		if len(db.Assets) > 0 {
			assets = append(assets, db.Assets...)
		}
	}
	return assets
}

func (mine *EventInfo) initInfo(db *nosql.Event) {
	mine.UID = db.UID.Hex()
	mine.ID = db.ID
	mine.Type = db.Type
	mine.Subtype = db.Subtype
	mine.Created = db.Created
	mine.Updated = db.Updated
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
	mine.Owner = db.Owner
	mine.Assets = db.Assets
	mine.Access = db.Access
	mine.Tags = db.Tags
	mine.Certify = db.Certify
	mine.Relations = db.Relations
	mine.Targets = db.Targets
	if db.Targets == nil {
		mine.Targets = make([]string, 0, 1)
		_ = nosql.UpdateEventTargets(mine.UID, mine.Operator, make([]string, 0, 1))
	}
}

func (mine *EventInfo) initByBrief(entity string, info *proxy.EventBrief) {
	msg := info.Name
	if len(msg) < 1 {
		msg = info.Description
	}
	mine.UID = "static_" + tool.StrToMD5(msg)
	mine.ID = 0
	mine.Type = 99
	mine.Operator = ""
	mine.Creator = "system"
	mine.Name = info.Name
	mine.Entity = entity
	mine.Parent = ""
	mine.Description = info.Description
	mine.Date = info.Date
	mine.Place = info.Place
	mine.Cover = ""
	mine.Quote = info.Quote
	mine.Assets = info.Assets
	mine.Access = AccessRead
	mine.Tags = info.Tags
	mine.Targets = make([]string, 0, 1)
	mine.Relations = make([]proxy.RelationCaseInfo, 0, 1)
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
		mine.Updated = time.Now().Unix()
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
		mine.Updated = time.Now().Unix()
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
		mine.Updated = time.Now().Unix()
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
		mine.Updated = time.Now().Unix()
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
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *EventInfo) UpdateOwner(owner, operator string) error {
	if operator == "" {
		operator = mine.Operator
	}
	if mine.Owner == owner {
		return nil
	}
	err := nosql.UpdateEventOwner(mine.UID, owner, operator)
	if err == nil {
		mine.Owner = owner
		mine.Operator = operator
		mine.Updated = time.Now().Unix()
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
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *EventInfo) UpdateCertify(uid, operator string) error {
	if operator == "" {
		operator = mine.Operator
	}
	if mine.Certify == uid {
		return nil
	}
	err := nosql.UpdateEventCertify(mine.UID, uid, operator)
	if err == nil {
		mine.Certify = uid
		mine.Operator = operator
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *EventInfo) UpdateTargets(operator string, arr []string) error {
	if operator == "" {
		operator = mine.Operator
	}
	err := nosql.UpdateEventTargets(mine.UID, operator, arr)
	if err == nil {
		mine.Targets = arr
		mine.Operator = operator
		mine.Updated = time.Now().Unix()
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
		mine.Updated = time.Now().Unix()
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
		mine.Updated = time.Now().Unix()
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
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *EventInfo) AppendRelation(relation *proxy.RelationCaseInfo) error {
	relation.UID = fmt.Sprintf("%s-%d", mine.UID, nosql.GetRelationCaseNextID())
	err := nosql.AppendEventRelation(mine.UID, relation)
	if err == nil {
		mine.Relations = append(mine.Relations, *relation)
		mine.Updated = time.Now().Unix()
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
		mine.Updated = time.Now().Unix()
	}
	return err
}
