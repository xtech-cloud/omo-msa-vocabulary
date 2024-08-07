package cache

import (
	"github.com/micro/go-micro/v2/logger"
	"github.com/mozillazg/go-pinyin"
	"omo.msa.vocabulary/config"
	"omo.msa.vocabulary/proxy/graph"
	"omo.msa.vocabulary/proxy/nosql"
	"omo.msa.vocabulary/tool"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	ErrorHadPublished = "the entity had published so can not update"
)

type BaseInfo struct {
	ID       uint64 `json:"id"`
	UID      string `json:"uid"`
	Name     string `json:"name"`
	Creator  string `json:"creator"`
	Operator string `json:"operator"`
	Created  int64
	Updated  int64
}

type WritingInfo struct {
	Number uint16 `json:"number"`
	Hex    string `json:"hex"`
}

type PairInfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Count int32  `json:"count"`
}

type DurationInfo struct {
	Date    time.Time `json:"date"`
	Seconds uint16    `json:"seconds"`
}

type FileInfo struct {
	UID         string    `json:"_id" bson:"_id"`
	UpdatedTime time.Time `json:"uploadDate" bson:"uploadDate"`
	Name        string    `json:"filename" bson:"filename"`
	MD5         string    `json:"md5" bson:"md5"`
	Size        int64     `json:"length" bson:"length"`
	Type        string    `json:"contentType" bson:"contentType"`
}

type NodeTemp struct {
	Entity  string
	Name    string
	Cover   string
	Concept string
}

type LinkTemp struct {
	UUID      string
	From      string
	To        string
	Kind      LinkType
	Relation  string
	Name      string
	Direction uint8
	Weight    uint32
}

type CountMap struct {
	Map   sync.Map
	Count uint32
}

type cacheContext struct {
	graph        *GraphInfo
	entityTables []string
	nodesMap     *CountMap
	linkMap      *CountMap
}

var cacheCtx *cacheContext

func InitData() error {
	cacheCtx = &cacheContext{}

	cacheCtx.entityTables = make([]string, 0, 3)
	cacheCtx.entityTables = append(cacheCtx.entityTables, DefaultEntityTable)
	cacheCtx.entityTables = append(cacheCtx.entityTables, UserEntityTable)
	cacheCtx.graph = new(GraphInfo)
	cacheCtx.nodesMap = new(CountMap)
	cacheCtx.linkMap = new(CountMap)
	cacheCtx.graph.construct()

	err := nosql.InitDB(config.Schema.Database.IP, config.Schema.Database.Port, config.Schema.Database.Name, config.Schema.Database.Type)
	if nil != err {
		return err
	}
	err1 := graph.InitNeo4J(&config.Schema.Graph)
	if err1 != nil {
		return err1
	}

	logger.Infof("init graph!!! node number = %d,link number = %d", len(cacheCtx.graph.nodes), len(cacheCtx.graph.links))
	return nil
}

func Context() *cacheContext {
	return cacheCtx
}

func switchAttributes() {
	info := cacheCtx.GetAttribute("60b9908fa0449d245dbde674")
	if info == nil {
		return
	}
	list := cacheCtx.getEntitiesByAttribute("60a4e2e36956c7f1bbe32414")
	for _, item := range list {
		item.replaceAttribute("60a4e2e36956c7f1bbe32414", "60b9908fa0449d245dbde674")
	}
}

func CheckBoxes() {
	all, _ := nosql.GetBoxes()
	for _, db := range all {
		tmp := new(BoxInfo)
		tmp.initInfo(db)
	}
}

func CheckRepeatedAttribute() {
	all, _ := nosql.GetAllAttributes()
	list := make([]*nosql.Attribute, 0, 100)
	repeats := make([]*nosql.Attribute, 0, 100)
	for _, item := range all {
		if !hadOne(strings.TrimSpace(item.Name), list) {
			list = append(list, item)
		} else {
			repeats = append(repeats, item)
		}
	}
	num := len(repeats)
	logger.Warnf("repeat attribute count = %d", num)
	for _, repeat := range repeats {
		news := getAttributeUID(strings.TrimSpace(repeat.Name), list)
		if news != "" {
			uid := repeat.UID.Hex()
			arr := cacheCtx.getEntitiesByAttribute(uid)
			for _, entity := range arr {
				_ = entity.replaceAttribute(uid, news)
			}
			arr1 := cacheCtx.getArchivedEntitiesByAttribute(uid)
			for _, arch := range arr1 {
				_ = arch.replaceAttribute(uid, news)
			}
			arr2 := cacheCtx.GetConceptsByAttribute(uid)
			for _, item := range arr2 {
				_ = item.ReplaceAttributes(uid, news)
			}
			_ = cacheCtx.RemoveAttribute(uid, repeat.Operator)
		}
	}
}

func getAttributeUID(name string, list []*nosql.Attribute) string {
	for _, info := range list {
		n := strings.TrimSpace(info.Name)
		if n == name {
			return info.UID.Hex()
		}
	}
	return ""
}

func hadOne(name string, list []*nosql.Attribute) bool {
	for _, info := range list {
		n := strings.TrimSpace(info.Name)
		if n == name {
			return true
		}
	}
	return false
}

func checkEntityLetters() {
	for _, table := range cacheCtx.entityTables {
		all, er := nosql.GetEntities(table)
		if er == nil {
			for _, entity := range all {
				if len(entity.FirstLetters) < 2 {
					letter := firstLetter(entity.Name)
					_ = nosql.UpdateEntityLetter(table, entity.UID.Hex(), letter)
				}
			}
		}
	}
}

func checkSequence() {
	arr := make([]string, 0, 6)
	arr = append(arr, "voc_"+nosql.TableArchived)
	arr = append(arr, "voc_"+nosql.TableAttribute)
	arr = append(arr, "voc_"+nosql.TableBox)
	arr = append(arr, "voc_"+nosql.TableConcept)
	arr = append(arr, "voc_"+nosql.TableEvent)
	arr = append(arr, "voc_"+nosql.TableRelation)
	arr = append(arr, "voc_"+nosql.TableRelationCase)
	all, _ := nosql.GetAllSequences()
	for _, s := range all {
		if tool.HasItem(arr, s.Name) {
			k := strings.Replace(s.Name, "voc_", "", 1)
			_ = nosql.UpdateSequenceName(s.UID.Hex(), k)
		}
	}

	arr2 := make([]string, 0, 6)
	arr2 = append(arr2, nosql.TableArchived)
	arr2 = append(arr2, nosql.TableAttribute)
	arr2 = append(arr2, nosql.TableBox)
	arr2 = append(arr2, nosql.TableConcept)
	arr2 = append(arr2, nosql.TableEvent)
	arr2 = append(arr2, nosql.TableRelation)
	arr2 = append(arr2, nosql.TableRelationCase)
	arr2 = append(arr2, nosql.TableSequence)
	arr2 = append(arr2, nosql.TableAddress)
	arr2 = append(arr2, DefaultEntityTable)
	arr2 = append(arr2, DefaultEntityTable+"_school")
	all2, _ := nosql.GetAllSequences()
	for _, s := range all2 {
		if !tool.HasItem(arr2, s.Name) {
			_ = nosql.DeleteSequence(s.UID.Hex())
		}
	}
}

func HadChinese(str string) bool {
	var count int
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			count++
			break
		}
	}
	return count > 0
}

func (mine *cacheContext) addSyncNode(uid, name, concept, cover string) {
	tmp := NodeTemp{
		Entity:  uid,
		Name:    name,
		Concept: concept,
		Cover:   cover,
	}
	mine.nodesMap.Map.Store(uid, &tmp)
	mine.nodesMap.Count += 1
}

func (mine *cacheContext) addSyncLink(from, to, relation, name string, kind LinkType, dir uint8) {
	tmp := LinkTemp{
		UUID:      from + "-" + to,
		From:      from,
		To:        to,
		Relation:  relation,
		Kind:      kind,
		Name:      name,
		Direction: dir,
	}
	mine.linkMap.Map.Store(tmp.UUID, &tmp)
	mine.linkMap.Count += 1
}

func (mine *CountMap) deleteSyncNode(uid string) {
	mine.Map.Delete(uid)
	if mine.Count > 0 {
		mine.Count -= 1
	}
}

func (mine *CountMap) getSyncNode(uid string) *NodeTemp {
	info, ok := mine.Map.Load(uid)
	if ok {
		return info.(*NodeTemp)
	}
	return nil
}

func (mine *cacheContext) EntityTables() []string {
	return mine.entityTables
}

func (mine *cacheContext) CheckSyncNodes() {
	if mine.nodesMap.Count < 1 {
		return
	}
	array := make([]string, 0, 20)
	call := func(key interface{}, val interface{}) bool {
		item := val.(*NodeTemp)
		_, err := mine.graph.CreateNode(0, item.Name, item.Entity, item.Cover, item.Concept, nil)
		if err == nil {
			array = append(array, item.Entity)
		}
		return true
	}
	mine.nodesMap.Map.Range(call)
	for i := 0; i < len(array); i += 1 {
		mine.nodesMap.deleteSyncNode(array[i])
	}
}

func (mine *cacheContext) CheckSyncLinks() {
	if mine.linkMap.Count < 1 {
		return
	}
	array := make([]string, 0, 20)
	call := func(key interface{}, val interface{}) bool {
		item := val.(*LinkTemp)
		err := mine.createLink(item.From, item.To, item.Kind, item.Relation, item.Name, item.Direction, item.Weight)
		if err == nil {
			array = append(array, item.UUID)
		}
		return true
	}
	mine.linkMap.Map.Range(call)
	for i := 0; i < len(array); i += 1 {
		mine.linkMap.deleteSyncNode(array[i])
	}
}

func (mine *cacheContext) createLink(from, to string, kind LinkType, relationUID, name string, dire uint8, weight uint32) error {
	fromNode := mine.GetGraphNode(from)
	toNode := mine.GetGraphNode(to)
	_, err := mine.graph.CreateLink(fromNode, toNode, kind, name, relationUID, DirectionType(dire), weight)
	return err
}

func StringToUint32(str string) uint32 {
	num, _ := strconv.ParseUint(str, 10, 32)
	return uint32(num)
}

func stringToUint16(str string) uint16 {
	num, _ := strconv.ParseUint(str, 10, 16)
	return uint16(num)
}

func stringToUint8(str string) uint8 {
	num, _ := strconv.ParseUint(str, 10, 8)
	return uint8(num)
}

func convertExcelDays(days int64) (year uint16, month uint8) {
	y := days / 366
	rest := days - (y * 365) - (y / 4)
	m := rest / 30
	year = uint16(y + 1900)
	month = uint8(m) + 1
	return year, month
}

func parseDate(date string) (year uint16, month uint8) {
	if strings.Contains(date, "年") {
		array := strings.Split(date, "年")
		if array == nil {
			return year, month
		}
		if len(array) > 0 {
			year = stringToUint16(array[0])
		}
		if len(array) > 1 {
			if strings.Contains(array[1], "月") {
				array1 := strings.Split(array[1], "月")
				if array1 != nil && len(array1) > 0 {
					month = stringToUint8(array1[0])
				}
			} else {
				month = stringToUint8(array[1])
			}
		}
	} else {
		days, err := strconv.ParseInt(date, 10, 32)
		if err == nil {
			return convertExcelDays(days)
		}
	}
	return year, month
}

func firstLetter(name string) string {
	if len(name) < 1 {
		return ""
	}
	//first := string([]rune(name)[:1])
	a := pinyin.NewArgs()
	a.Style = pinyin.FirstLetter
	arr := pinyin.Pinyin(name, a)
	var letter = ""
	for i, _ := range arr {
		letter = letter + arr[i][0]
	}
	return strings.ToUpper(letter)
}

func hadChinese(str string) bool {
	var count int
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			count++
			break
		}
	}
	return count > 0
}

func CheckPage[T any](page, number int32, all []T) (int32, int32, []T) {
	if len(all) < 1 {
		return 0, 0, make([]T, 0, 1)
	}
	if number < 1 {
		number = 10
	}
	total := int32(len(all))
	if len(all) <= int(number) {
		return total, 1, all
	}
	maxPage := total/number + 1
	if page < 1 {
		return total, maxPage, all
	}
	if page > maxPage {
		page = maxPage
	}
	var start = (page - 1) * number
	var end = start + number
	if end >= total {
		end = total
	}
	list := make([]T, 0, number)
	list = append(all[start:end])
	return total, maxPage, list
}

func DateToUTC(date string) int64 {
	if date == "" {
		return 0
	}
	t, e := time.ParseInLocation("2006/01/02", date, time.Local)
	if e != nil {
		return 0
	}
	return t.Unix()
}
