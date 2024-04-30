package cache

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/micro/go-micro/v2/logger"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
	"omo.msa.vocabulary/tool"
	"strconv"
	"time"
)

const (
	EntityStatusDraft   EntityStatus = 0
	EntityStatusFirst   EntityStatus = 1
	EntityStatusPending EntityStatus = 2
	EntityStatusSpecial EntityStatus = 3
	EntityStatusUsable  EntityStatus = 4 //审核通过
	EntityStatusFailed  EntityStatus = 10
	EntityStatusAll     EntityStatus = 99
)

const (
	DefaultEntityTable = "entities"
	UserEntityTable    = "entities_school"
)

const (
	OptionAgree  OptionType = 1 //审核同意
	OptionRefuse OptionType = 2 //审核拒绝
	OptionSwitch OptionType = 3 //切换关联
)

const (
	AccessAll    = 0 //全部可以访问
	AccessIgnore = 1 //
)

type EntityStatus uint8

type OptionType uint8

type EntityInfo struct {
	Status EntityStatus `json:"-"`
	Pushed int64        `json:"-"`
	BaseInfo
	FirstLetters string `json:"letters"` //名称首字母
	Concept      string `json:"concept"`
	Summary      string `json:"summary"`
	Description  string `json:"description"`
	Cover        string `json:"cover"`
	Add          string `json:"add"`   //消歧义
	Owner        string `json:"owner"` //所属单位
	Mark         string `json:"mark"`  // 标记采集来源
	Quote        string `json:"quote"` // 引用外部链接，或者群晖路径
	Published    bool   `json:"published"`
	Thumb        string `json:"thumb"` //图谱头像
	Access       uint8  `json:"-"`     //是否可被第三方访问，默认0是可以被访问的
	Score        uint32 `json:"-"`
	dbTable      string
	Links        []string `json:"links" bson:"links"` //可与其他实体链接
	Synonyms     []string `json:"synonyms"`           //同义词
	Tags         []string `json:"tags"`               //标签
	Relates      []string `json:"relates"`            //关联的一些数据，可以是社区，场景等

	Properties   []*proxy.PropertyInfo `json:"properties"`
	StaticEvents []*proxy.EventBrief   `json:"events"`
	StaticVEdges []*VEdgeInfo          `json:"relations"`
	events       []*EventInfo          `json:"-"`
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

func (mine *cacheContext) CreateEntity(info *EntityInfo, relations []*pb.VEdgeInfo) error {
	if info == nil {
		return errors.New("the entity info is nil")
	}
	db := new(nosql.Entity)
	db.UID = primitive.NewObjectID()
	db.Created = time.Now().Unix()
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
	db.Thumb = ""
	db.Access = info.Access
	db.Synonyms = info.Synonyms
	db.Events = info.StaticEvents
	//db.Relations = info.StaticRelations
	db.Relates = info.Relates
	if db.Relates == nil {
		db.Relates = make([]string, 0, 1)
	}
	info.events = make([]*EventInfo, 0, 1)
	if info.Properties == nil {
		info.Properties = make([]*proxy.PropertyInfo, 0, 1)
	}
	db.FirstLetters = firstLetter(info.Name)
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
		_ = info.UpdateStaticRelations(info.Operator, relations)
		mine.syncGraphNode(info)
	}
	return err
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
	mine.Created = db.Created
	mine.Updated = db.Updated
	mine.Creator = db.Creator
	mine.Operator = db.Operator
	mine.Tags = db.Tags
	mine.Name = db.Name
	mine.FirstLetters = db.FirstLetters
	mine.Add = db.Add
	mine.Pushed = db.Pushed
	mine.Description = db.Description
	mine.Concept = db.Concept
	mine.Status = EntityStatus(db.Status)
	mine.Owner = db.Scene
	mine.Cover = db.Cover
	mine.Mark = db.Mark
	mine.Quote = db.Quote
	mine.Access = db.Access
	mine.Summary = db.Summary
	mine.Thumb = db.Thumb
	mine.Links = db.Links
	mine.dbTable = db.Table
	if mine.Links == nil {
		mine.Links = make([]string, 0, 1)
	}
	mine.Relates = db.Relates
	if mine.Relates == nil {
		mine.Relates = make([]string, 0, 1)
	}
	if cacheCtx.HadArchivedByEntity(mine.UID) {
		mine.Published = true
	} else {
		mine.Published = false
	}

	if len(db.Relations) > 0 {
		mine.relationsToVEdges(db.Relations)
	}
	//mine.StaticRelations = mine.GetVEdges()

	mine.StaticEvents = db.Events
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
	if len(mine.dbTable) > 0 {
		return mine.dbTable
	}

	if len(mine.Concept) < 2 {
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

func (mine *EntityInfo) updateConcept(concept, operator string) error {
	if mine.Concept != concept {
		err := nosql.UpdateEntityConcept(mine.table(), mine.UID, concept, operator)
		if err == nil {
			mine.Concept = concept
			mine.Operator = operator
		}
		return err
	} else {
		return nil
	}
}

func (mine *EntityInfo) replaceAttribute(old, news string) error {
	props := make([]*proxy.PropertyInfo, 0, len(mine.Properties))
	props = append(props, mine.Properties...)
	for _, prop := range props {
		if prop.Key == old {
			prop.Key = news
		}
	}
	err := nosql.UpdateEntityProperties(mine.table(), mine.UID, mine.Operator, props)
	if err == nil {
		mine.Properties = props
	}
	return err
}

func (mine *EntityInfo) relationsToVEdges(list []*proxy.RelationCaseInfo) {
	if len(list) < 1 {
		return
	}
	logger.Warn("relationsToVEdges...uid = " + mine.UID + "; edges count = " + strconv.Itoa(len(list)))
	for _, item := range list {
		target := proxy.VNode{Name: "", Entity: "", UID: "", Thumb: ""}
		if hadChinese(item.Entity) {
			target.Name = item.Entity
			target.Entity = ""
		} else {
			target.Name = ""
			target.Entity = item.Entity
		}
		_, _ = mine.CreateVEdge(mine.UID, item.Name, "", item.Category, mine.Operator, uint32(item.Direction), item.Weight, target)
	}
	_ = nosql.UpdateEntityRelations(mine.table(), mine.UID, mine.Operator, make([]*proxy.RelationCaseInfo, 0, 1))
}

func (mine *EntityInfo) GetVEdges() []*VEdgeInfo {
	return cacheCtx.GetVEdgesByCenter(mine.UID)
}

func (mine *EntityInfo) UpdateBase(name, desc, add, concept, cover, mark, quote, sum, operator string) error {
	if mine.Status != EntityStatusDraft {
		return errors.New("the entity is not draft so can not update")
	}
	if concept == "" {
		concept = mine.Concept
	}
	if name == "" {
		name = mine.Name
	}
	if mark == "" {
		mark = mine.Mark
	}
	if quote == "" {
		quote = mine.Quote
	}
	var err error
	if len(cover) > 0 && cover != mine.Cover {
		err = mine.UpdateCover(cover, operator)
	}
	if desc != mine.Description || sum != mine.Summary {
		err = nosql.UpdateEntityRemark(mine.table(), mine.UID, desc, sum, operator)
		if err == nil {
			mine.Description = desc
			mine.Summary = sum
			mine.Operator = operator
			mine.Updated = time.Now().Unix()
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
			mine.Updated = time.Now().Unix()
		}
	}
	return err
}

func (mine *EntityInfo) UpdateStatic(info *EntityInfo, relations []*pb.VEdgeInfo) error {
	if mine.Status != EntityStatusDraft {
		return errors.New("the entity is not draft so can not update")
	}
	_ = mine.UpdateBase(info.Name, info.Description, info.Add, info.Concept, info.Cover, info.Mark, info.Quote, info.Summary, info.Operator)
	err := nosql.UpdateEntityStatic(mine.table(), mine.UID, info.Operator, info.Tags, info.Properties)
	if err == nil {
		mine.Tags = info.Tags
		mine.Properties = info.Properties
		mine.Updated = time.Now().Unix()
	}
	if len(info.StaticEvents) > 0 {
		_ = mine.UpdateStaticEvents(info.Operator, info.StaticEvents)
	}
	if len(relations) > 0 {
		_ = mine.UpdateStaticRelations(info.Operator, relations)
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
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *EntityInfo) GetEventCount() uint32 {
	return uint32(len(mine.events)) + nosql.GetEventCountByEntity(mine.UID)
}

func (mine *EntityInfo) GetVEdgeCount() uint32 {
	return nosql.GetVEdgeCountByEntity(mine.UID)
}

func (mine *EntityInfo) UpdateStaticRelations(operator string, list []*pb.VEdgeInfo) error {
	if mine.Status != EntityStatusDraft {
		return errors.New("the entity is not draft so can not update")
	}

	for _, brief := range list {
		target := proxy.VNode{
			Name:   brief.Target.Name,
			Entity: brief.Target.Entity,
			Thumb:  brief.Target.Thumb,
			UID:    brief.Target.Uid,
		}
		if brief.Uid != "" {
			err := nosql.UpdateVEdgeBase(brief.Uid, brief.Name, brief.Remark, brief.Category, operator, uint8(brief.Direction), target)
			if err != nil {
				return err
			}
		} else {
			_, err := mine.CreateVEdge(brief.Source, brief.Name, brief.Remark, brief.Category, operator, brief.Direction, brief.Weight, target)
			if err != nil {
				return err
			}
		}
	}
	//err := nosql.UpdateEntityRelations(mine.table(), mine.UID, operator, list)
	//if err == nil {
	//	mine.Operator = operator
	//mine.StaticRelations = list
	//	mine.Updated = time.Now().Unix()
	//}
	return nil
}

func (mine *EntityInfo) UpdateCover(cover, operator string) error {
	if cover == "" || cover == mine.Cover {
		return nil
	}
	err := nosql.UpdateEntityCover(mine.table(), mine.UID, cover, operator)
	if err == nil {
		mine.Cover = cover
		mine.Operator = operator
		mine.Updated = time.Now().Unix()
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
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *EntityInfo) UpdateThumb(thumb, operator string) error {
	if thumb == "" || thumb == mine.Thumb {
		return nil
	}
	err := nosql.UpdateEntityThumb(mine.table(), mine.UID, thumb, operator)
	if err == nil {
		mine.Thumb = thumb
		mine.Operator = operator
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *EntityInfo) UpdateMark(mark, operator string) error {
	if mark == mine.Mark {
		return nil
	}
	err := nosql.UpdateEntityMark(mine.table(), mine.UID, mark, operator)
	if err == nil {
		mine.Mark = mark
		mine.Operator = operator
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *EntityInfo) UpdateQuote(quote, operator string) error {
	if quote == mine.Quote {
		return nil
	}
	err := nosql.UpdateEntityQuote(mine.table(), mine.UID, quote, operator)
	if err == nil {
		mine.Quote = quote
		mine.Operator = operator
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *EntityInfo) UpdateProperty(uid, val, operator string) error {
	arr := make([]*proxy.PropertyInfo, 0, len(mine.Properties))
	arr = append(arr, mine.Properties...)
	had := false
	for _, info := range arr {
		if info.Key == uid {
			if len(info.Words) > 0 {
				info.Words[0].Name = val
			} else {
				info.Words = make([]proxy.WordInfo, 0, 1)
				info.Words = append(info.Words, proxy.WordInfo{Name: val, UID: ""})
			}
			had = true
			break
		}
	}
	if !had {
		info := new(proxy.PropertyInfo)
		info.Key = uid
		info.Words = make([]proxy.WordInfo, 0, 1)
		info.Words = append(info.Words, proxy.WordInfo{Name: val, UID: ""})
		arr = append(arr, info)
	}

	return mine.UpdateProperties(arr, operator)
}

func (mine *EntityInfo) GetRecords() ([]*nosql.Record, error) {
	dbs, err := nosql.GetRecords(mine.UID)
	if err != nil {
		return nil, err
	}
	list := make([]*nosql.Record, 0, len(dbs))
	for _, db := range dbs {
		if db.Option == uint8(OptionAgree) || db.Option == uint8(OptionRefuse) {
			list = append(list, db)
		}
	}
	return list, nil
}

func (mine *EntityInfo) UpdateTags(tags []string, operator string) error {
	if mine.Status == EntityStatusUsable {
		return errors.New("the entity had published so can not update")
	}
	err := nosql.UpdateEntityTags(mine.table(), mine.UID, operator, tags)
	if err == nil {
		mine.Tags = tags
		mine.Operator = operator
		mine.Updated = time.Now().Unix()
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
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *EntityInfo) createRecord(operator, remark string, from, to EntityStatus) {
	opt := OptionAgree
	if to > from {
		opt = OptionAgree
	} else {
		opt = OptionRefuse
	}

	_ = mine.insertRecord(operator, remark, fmt.Sprintf("%d", from), fmt.Sprintf("%d", to), opt)
}

func (mine *EntityInfo) insertRecord(operator, remark, from, to string, opt OptionType) error {
	db := new(nosql.Record)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetRecordNextID()
	db.Creator = operator
	db.Created = time.Now().Unix()
	db.Entity = mine.UID
	db.From = from
	db.To = to
	db.Option = uint8(opt)
	db.Remark = remark
	return nosql.CreateRecord(db)
}

func (mine *EntityInfo) HadRecord(to string) bool {
	dbs, err := nosql.GetRecordsBy(mine.UID, to, uint8(OptionSwitch))
	if err != nil {
		return false
	}
	if len(dbs) > 0 {
		return true
	} else {
		return false
	}
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
		mine.Published = true
		tmp := Context().GetArchivedByEntity(mine.UID)
		if tmp == nil {
			err = Context().CreateArchived(mine)
			if err != nil {
				return err
			}
			//cacheCtx.checkRelations(nil, mine)
		} else {
			_, er := tmp.Decode()
			if er != nil {
				return er
			}
			err = tmp.UpdateFile(mine, operator)
			if err != nil {
				return err
			}
			//cacheCtx.checkRelations(old, mine)
		}
	}
	cacheCtx.UpdateBoxContentStatus(mine.UID, status, mine.Published)
	mine.Status = status
	mine.Updated = time.Now().Unix()
	return nil
}

func (mine *EntityInfo) encode() (string, string, uint32, error) {
	mine.StaticVEdges = mine.GetVEdges()
	bts, er := json.Marshal(mine)
	if er != nil {
		return "", "", 0, er
	}
	md5 := tool.CalculateMD5(bts)
	size := len(bts)
	data := base64.StdEncoding.EncodeToString(bts)
	return data, md5, uint32(size), nil
}

func (mine *EntityInfo) UpdatePushTime(operator string) error {
	err := nosql.UpdateEntityPushed(mine.table(), mine.UID, operator)
	if err != nil {
		return err
	}
	mine.Operator = operator
	mine.Pushed = time.Now().Unix()
	mine.Updated = time.Now().Unix()
	return nil
}

func (mine *EntityInfo) UpdateRelates(operator string, list []string) error {
	if list == nil {
		list = make([]string, 0, 1)
	}
	err := nosql.UpdateEntityRelates(mine.table(), mine.UID, operator, list)
	if err != nil {
		return err
	}
	from, to := tool.DifferenceStrings(mine.Relates, list)
	if len(to) > 2 && from != to {
		_ = mine.insertRecord(operator, "", from, to, OptionSwitch)
	}
	mine.Operator = operator
	mine.Relates = list
	mine.Updated = time.Now().Unix()
	return nil
}

func (mine *EntityInfo) UpdateLinks(operator string, list []string) error {
	if list == nil {
		list = make([]string, 0, 1)
	}
	err := nosql.UpdateEntityLinks(mine.table(), mine.UID, operator, list)
	if err != nil {
		return err
	}

	mine.Operator = operator
	mine.Links = list
	mine.Updated = time.Now().Unix()
	return nil
}

func (mine *EntityInfo) UpdateAccess(operator string, acc uint8) error {
	info := cacheCtx.GetArchivedByEntity(mine.UID)
	if info == nil {
		return errors.New("not found the archived file by entity")
	}
	err := nosql.UpdateArchivedAccess(info.UID, operator, acc)
	if err != nil {
		return err
	}

	info.Operator = operator
	info.Access = acc
	//mine.Updated = time.Now().Unix()
	return nil
}

//region VEdge fun
func (mine *EntityInfo) CreateVEdge(source, name, remark, relation, operator string, dire, weight uint32, target proxy.VNode) (*VEdgeInfo, error) {
	if target.Name == "" {
		return nil, errors.New("the target is empty")
	}
	target.UID = "temp-" + primitive.NewObjectID().Hex()
	db := new(nosql.VEdge)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetVEdgeNextID()
	db.Created = time.Now().Unix()
	db.Creator = operator
	db.Name = name
	db.Source = source
	db.Center = mine.UID
	db.Catalog = relation
	db.Target = target
	db.Direction = uint8(dire)
	db.Weight = weight
	db.Remark = remark
	if db.Source == "" {
		db.Source = mine.UID
	}
	err := nosql.CreateVEdge(db)
	if err != nil {
		return nil, err
	}
	info := new(VEdgeInfo)
	info.initInfo(db)
	return info, nil
}

func (mine *EntityInfo) GetPublicEdges() []*VEdgeInfo {
	if mine.StaticVEdges != nil {
		return mine.StaticVEdges
	}
	return cacheCtx.GetVEdgesByCenter(mine.UID)
}

//endregion

//region Event Fun
func (mine *EntityInfo) initEvents() {
	if mine.events != nil {
		return
	}
	events, err := nosql.GetEventsByEntity(mine.UID)
	if err == nil {
		mine.events = make([]*EventInfo, 0, len(events))
		for _, event := range events {
			tmp := new(EventInfo)
			tmp.initInfo(event)
			mine.events = append(mine.events, tmp)
		}
	} else {
		mine.events = make([]*EventInfo, 0, 10)
	}
	if len(mine.StaticEvents) > 0 {
		for _, event := range mine.StaticEvents {
			tmp := new(EventInfo)
			tmp.initByBrief(mine.UID, event)
			mine.events = append(mine.events, tmp)
		}
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
		arr, err = nosql.GetEventsByTypeQuote(mine.UID, quote, tp)
	} else {
		arr, err = nosql.GetEventsByType(mine.UID, tp)
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

func (mine *EntityInfo) GetEventsByQuote(quote string) []*EventInfo {
	arr, err := nosql.GetEventsByQuote(mine.UID, quote)
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

func (mine *EntityInfo) GetEventsByAccess(tp, access uint8) []*EventInfo {
	var arr []*nosql.Event
	var err error
	var list = make([]*EventInfo, 0, 50)
	if tp > EventCustom {
		arr, err = nosql.GetEventsByTypeAndAccess(mine.UID, tp, access)
	} else {
		arr, err = nosql.GetEventsByAccess(mine.UID, access)
	}
	if len(mine.StaticEvents) > 0 {
		for _, event := range mine.StaticEvents {
			tmp := new(EventInfo)
			tmp.initByBrief(mine.UID, event)
			list = append(list, tmp)
		}
	}
	if err == nil {
		for _, db := range arr {
			info := new(EventInfo)
			info.initInfo(db)
			list = append(list, info)
		}
	}

	return list
}

func (mine *EntityInfo) GetPublicEvents() []*EventInfo {
	var list = make([]*EventInfo, 0, 50)
	arr, _ := nosql.GetEventsByAccess(mine.UID, AccessRead)
	for _, event := range arr {
		info := new(EventInfo)
		info.initInfo(event)
		list = append(list, info)
	}
	arr2, _ := nosql.GetEventsByAccess(mine.UID, AccessWR)
	for _, event := range arr2 {
		info := new(EventInfo)
		info.initInfo(event)
		list = append(list, info)
	}

	if len(mine.StaticEvents) > 0 {
		for _, event := range mine.StaticEvents {
			tmp := new(EventInfo)
			tmp.initByBrief(mine.UID, event)
			list = append(list, tmp)
		}
	}

	return list
}

func (mine *EntityInfo) AddEvent(data *pb.ReqEventAdd) (*EventInfo, error) {
	if mine.Status == EntityStatusUsable {
		return nil, errors.New("the entity had published so can not update")
	}
	mine.initEvents()
	begin := proxy.Date{}
	end := proxy.Date{}
	if data.Date != nil {
		_ = begin.Parse(data.Date.Begin)
		_ = end.Parse(data.Date.End)
	}

	date := proxy.DateInfo{UID: data.Date.Uid, Name: data.Date.Name, Begin: begin, End: end}
	place := proxy.PlaceInfo{UID: data.Place.Uid, Name: data.Place.Name, Location: data.Place.Location}
	relations := make([]proxy.RelationCaseInfo, 0, len(data.Relations))
	for _, value := range data.Relations {
		relations = append(relations, proxy.RelationCaseInfo{UID: value.Uid, Direction: uint8(value.Direction),
			Name: value.Name, Category: value.Category, Entity: value.Entity})
	}
	db := new(nosql.Event)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetEventNextID()
	db.CreatedTime = time.Now()
	db.Created = time.Now().Unix()
	db.Creator = data.Operator
	db.Name = data.Name
	db.Date = date
	db.Place = place
	db.Type = uint8(data.Type)
	db.Subtype = uint8(data.Sub)
	db.Entity = mine.UID
	db.Parent = ""
	db.Quote = data.Quote
	db.Certify = data.Certify
	db.Description = data.Description
	db.Relations = relations
	db.Cover = data.Cover
	db.Tags = data.Tags
	db.Access = uint8(data.Access)
	db.Owner = data.Owner
	if db.Tags == nil {
		db.Tags = make([]string, 0, 1)
	}
	db.Assets = data.Assets
	if db.Assets == nil {
		db.Assets = make([]string, 0, 1)
	}
	db.Targets = data.Targets
	if db.Targets == nil {
		db.Targets = make([]string, 0, 1)
	}
	err := nosql.CreateEvent(db)
	if err == nil {
		info := new(EventInfo)
		info.initInfo(db)
		mine.events = append(mine.events, info)

		for i := 0; i < len(relations); i += 1 {
			relationKind := Context().GetRelation(relations[i].Category)
			if relationKind != nil {
				Context().addSyncLink(mine.UID, relations[i].Entity, relationKind.UID, relations[i].Name, switchRelationToLink(relationKind.Kind), relations[i].Direction)
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
		mine.Updated = time.Now().Unix()
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
		mine.Updated = time.Now().Unix()
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
		mine.Updated = time.Now().Unix()
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
				if i == len(mine.Properties)-1 {
					mine.Properties = append(mine.Properties[:i])
				} else {
					mine.Properties = append(mine.Properties[:i], mine.Properties[i+1:]...)
				}

				break
			}
		}
		mine.Updated = time.Now().Unix()
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
