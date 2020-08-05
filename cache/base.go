package cache

import (
	"github.com/micro/go-micro/v2/logger"
	"mime/multipart"
	"omo.msa.vocabulary/config"
	"omo.msa.vocabulary/proxy/graph"
	"omo.msa.vocabulary/proxy/nosql"
	"strconv"
	"strings"
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

type cacheContext struct {
	graph		*GraphInfo
	entities  []*EntityInfo
	concerts   []*ConceptInfo
	attributes []*AttributeInfo
	relations []*RelationshipInfo
}

var cacheCtx *cacheContext

func InitData() error {
	cacheCtx = &cacheContext{}
	cacheCtx.entities = make([]*EntityInfo, 0, 1000)
	cacheCtx.concerts = make([]*ConceptInfo, 0, 50)
	cacheCtx.attributes = make([]*AttributeInfo, 0, 100)
	cacheCtx.relations = make([]*RelationshipInfo, 0, 100)
	cacheCtx.graph = new(GraphInfo)
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

	relations,_ := nosql.GetAllRelations()
	for i := 0; i < len(relations); i += 1 {
		info := new(RelationshipInfo)
		info.initInfo(relations[i])
		cacheCtx.relations = append(cacheCtx.relations, info)
	}
	logger.Infof("init relation!!! number = %d", len(cacheCtx.attributes))

	concerts,_ := nosql.GetTopConcepts()
	for i := 0; i < len(concerts); i += 1 {
		info := new(ConceptInfo)
		info.initInfo(concerts[i])
		cacheCtx.concerts = append(cacheCtx.concerts, info)
	}
	logger.Infof("init concerts!!! number = %d", len(cacheCtx.concerts))

	//for _, kind := range cacheCtx.concerts {
	//	entities,_ := nosql.GetEntities(kind.Table)
	//	for i := 0; i < len(entities); i += 1 {
	//		info := new(EntityInfo)
	//		info.initInfo(entities[i])
	//		cacheCtx.entities = append(cacheCtx.entities, info)
	//	}
	//}
	entities,_ := nosql.GetEntities(DefaultEntityTable)
	for i := 0; i < len(entities); i += 1 {
		info := new(EntityInfo)
		info.initInfo(entities[i])
		cacheCtx.entities = append(cacheCtx.entities, info)
	}
	logger.Infof("init entities!!! number = %d", len(cacheCtx.entities))
	//initDefConcepts()
	//readLocalExcels()
	//exportLocalJsons()
	return nil
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
