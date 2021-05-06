package cache

import (
	"github.com/micro/go-micro/v2/logger"
	"mime/multipart"
	"omo.msa.vocabulary/config"
	"omo.msa.vocabulary/proxy/graph"
	"omo.msa.vocabulary/proxy/nosql"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BaseInfo struct {
	ID         uint64 `json:"-"`
	UID        string `json:"uid"`
	Name       string `json:"name"`
	CreateTime time.Time
	UpdateTime time.Time
	Creator string
	Operator string
}

type WritingInfo struct {
	Number uint16 `json:"number"`
	Hex    string `json:"hex"`
}

type PairInfo struct {
	Key string `json:"key"`
	Value string `json:"value"`
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
	Entity string
	Name string
	Cover string
	Concept string
}

type LinkTemp struct {
	UUID string
	From string
	To string
	Kind LinkType
	Relation string
	Name string
	Direction uint8
}

type CountMap struct {
	Map sync.Map
	Count uint32
}

type cacheContext struct {
	graph      *GraphInfo
	//entities   []*EntityInfo
	concepts   []*ConceptInfo
	boxes      []*BoxInfo
	attributes []*AttributeInfo
	relations  []*RelationshipInfo
	nodesMap   *CountMap
	linkMap    *CountMap
}

var cacheCtx *cacheContext

func InitData() error {
	cacheCtx = &cacheContext{}
	//cacheCtx.entities = make([]*EntityInfo, 0, 1000)
	cacheCtx.concepts = make([]*ConceptInfo, 0, 50)
	cacheCtx.attributes = make([]*AttributeInfo, 0, 100)
	cacheCtx.relations = make([]*RelationshipInfo, 0, 100)
	cacheCtx.boxes = make([]*BoxInfo, 0, 50)
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

	attributes,_ := nosql.GetAllAttributes()
	for i := 0; i < len(attributes); i += 1 {
		info := new(AttributeInfo)
		info.initInfo(attributes[i])
		cacheCtx.attributes = append(cacheCtx.attributes, info)
	}
	logger.Infof("init attribute!!! number = %d", len(cacheCtx.attributes))

	relations,_ := nosql.GetTopRelations()
	for i := 0; i < len(relations); i += 1 {
		info := new(RelationshipInfo)
		info.initInfo(relations[i])
		cacheCtx.relations = append(cacheCtx.relations, info)
	}
	logger.Infof("init relation!!! number = %d", len(cacheCtx.relations))
	concerts,_ := nosql.GetTopConcepts()
	for i := 0; i < len(concerts); i += 1 {
		info := new(ConceptInfo)
		info.initInfo(concerts[i])
		cacheCtx.concepts = append(cacheCtx.concepts, info)
	}
	logger.Infof("init concerts!!! number = %d", len(cacheCtx.concepts))

	boxes,_ := nosql.GetBoxes()
	for i := 0; i < len(boxes); i += 1 {
		info := new(BoxInfo)
		info.initInfo(boxes[i])
		cacheCtx.boxes = append(cacheCtx.boxes, info)
	}
	logger.Infof("init boxes!!! number = %d", len(cacheCtx.boxes))

	//for _, kind := range cacheCtx.concerts {
	//	if kind.Table != "" {
	//		entities,_ := nosql.GetEntities(kind.Table)
	//		for i := 0; i < len(entities); i += 1 {
	//			info := new(EntityInfo)
	//			info.initInfo(entities[i])
	//			cacheCtx.entities = append(cacheCtx.entities, info)
	//		}
	//	}
	//}
	//entities,_ := nosql.GetEntities(DefaultEntityTable)
	//for i := 0; i < len(entities); i += 1 {
	//	info := new(EntityInfo)
	//	info.initInfo(entities[i])
	//	cacheCtx.entities = append(cacheCtx.entities, info)
	//}
	//logger.Infof("init entities!!! number = %d", len(cacheCtx.entities))
	logger.Infof("init graph!!! node number = %d,link number = %d", len(cacheCtx.graph.nodes), len(cacheCtx.graph.links))

	return nil
}

func Context() *cacheContext {
	return cacheCtx
}

func (mine *cacheContext)addSyncNode(uid, name, concept, cover string) {
	tmp := NodeTemp{
		Entity: uid,
		Name: name,
		Concept: concept,
		Cover: cover,
	}
	mine.nodesMap.Map.Store(uid, &tmp)
	mine.nodesMap.Count += 1
}

func (mine *cacheContext)addSyncLink(from, to, relation, name string, kind LinkType, dir uint8) {
	tmp := LinkTemp{
		UUID: from + "-" + to,
		From: from,
		To: to,
		Relation: relation,
		Kind: kind,
		Name: name,
		Direction: dir,
	}
	mine.nodesMap.Map.Store(tmp.UUID, &tmp)
	mine.nodesMap.Count += 1
}

func (mine *CountMap)deleteSyncNode(uid string)  {
	mine.Map.Delete(uid)
	if mine.Count > 0 {
		mine.Count -= 1
	}
}

func (mine *CountMap)getSyncNode(uid string) *NodeTemp {
	info, ok := mine.Map.Load(uid)
	if ok {
		return info.(*NodeTemp)
	}
	return nil
}

func (mine *cacheContext)CheckSyncNodes()  {
	if mine.nodesMap.Count < 1 {
		return
	}
	array := make([]string, 0, 20)
	call := func(key interface{}, val interface{}) bool {
		item := val.(*NodeTemp)
		_,err := mine.graph.CreateNode(item.Name, item.Entity, item.Cover, item.Concept)
		if err == nil {
			array = append(array, item.Entity)
		}
		return true
	}
	mine.nodesMap.Map.Range(call)
	for i := 0;i < len(array);i+=1 {
		mine.nodesMap.deleteSyncNode(array[i])
	}
}

func (mine *cacheContext)CheckSyncLinks()  {
	if mine.linkMap.Count < 1 {
		return
	}
	array := make([]string, 0, 20)
	call := func(key interface{}, val interface{}) bool {
		item := val.(*LinkTemp)
		err := mine.createLink(item.From, item.To, item.Kind, item.Relation, item.Name, item.Direction)
		if err == nil {
			array = append(array, item.UUID)
		}
		return true
	}
	mine.linkMap.Map.Range(call)
	for i := 0;i < len(array);i+=1 {
		mine.linkMap.deleteSyncNode(array[i])
	}
}

func (mine *cacheContext)createLink(from, to string, kind LinkType, relationUID, name string, dire uint8) error {
	fromNode := mine.GetGraphNode(from)
	toNode := mine.GetGraphNode(to)
	_, err := mine.graph.CreateLink(fromNode, toNode, kind, name, relationUID, DirectionType(dire))
	return err
}

func stringToUint32(str string) uint32 {
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
	m := rest/30
	year = uint16(y + 1900)
	month = uint8(m) + 1
	return year,month
}

func parseDate(date string) (year uint16, month uint8) {
	if strings.Contains(date,"年") {
		array := strings.Split(date, "年")
		if array == nil {
			return year, month
		}
		if len(array) > 0 {
			year = stringToUint16(array[0])
		}
		if len(array) > 1 {
			if strings.Contains(array[1],"月") {
				array1 := strings.Split(array[1], "月")
				if array1 != nil && len(array1) > 0 {
					month = stringToUint8(array1[0])
				}
			}else{
				month = stringToUint8(array[1])
			}
		}
	} else{
		days,err := strconv.ParseInt(date, 10, 32)
		if err == nil{
			return convertExcelDays(days)
		}
	}
	return year,month
}

func ImportDatabase(table string, file multipart.File) error {
	return nosql.ImportDatabase(table, file)
}
