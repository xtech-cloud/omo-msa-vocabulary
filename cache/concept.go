package cache

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy/nosql"
	"omo.msa.vocabulary/tool"
	"time"
)

const (
	ConceptTypeUnknown = 0
	ConceptTypePersonal = 1
	ConceptTypeUtensil  = 2  // 器物
	ConceptTypeEvent    = 3  //事件
	ConceptTypeOrganize = 4  // 组织
	ConceptTypeIdea     = 5  //思想理论
	ConceptTypeBook     = 6  //经籍著作
	ConceptTypeCulture  = 7  //文化
	ConceptTypeFaction  = 8  //派别
	ConceptTypeNature   = 9  //自然
	ConceptTypeHonor    = 10 //荣誉奖项
	ConceptTypePlace    = 11 //地理位置
	ConceptTypeEra      = 12 // 时代
)

type ConceptInfo struct {
	BaseInfo
	Type uint8
	Cover    string
	Remark   string
	Table    string
	Parent   string
	Scene    uint8  // 针对的场景类型
	attributes    []string
	children []*ConceptInfo
}

//region Global Fun
func initDefConcepts()  {
	bytes, err1 := tool.ReadFile("conf/def_concept.json")
	if err1 != nil {
		fmt.Println("read default concept error::"+err1.Error())
		return
	}
	result := gjson.Parse(string(bytes))
	for _, value := range result.Array() {
		info := new(ConceptInfo)
		info.Name = value.Get("name").String()
		info.Table = value.Get("table").String()
		array := value.Get("attributes").Array()
		if !HadConceptByTable(info.Table) {
			err := CreateTopConcept(info)
			if err == nil {
				for _, result := range array {
					key := result.Get("key").String()
					if !info.HadAttribute(key) {
						_ = info.CreateAttribute(key, result.Get("value").String(),
							"", "",  AttributeType(result.Get("type").Uint()))
					}
				}
			}
		}
	}
}

func GetTopConcept(uid string) *ConceptInfo {
	for i := 0; i < len(cacheCtx.concerts); i += 1 {
		if cacheCtx.concerts[i].HadChild(uid) {
			return cacheCtx.concerts[i]
		}
	}
	return nil
}

func GetConceptByName(name string) *ConceptInfo {
	for i := 0; i < len(cacheCtx.concerts); i += 1 {
		child := cacheCtx.concerts[i].GetChildByName(name)
		if child != nil {
			return child
		}
	}
	return nil
}

func GetConcept(uid string) *ConceptInfo {
	for i := 0; i < len(cacheCtx.concerts); i += 1 {
		child := cacheCtx.concerts[i].GetChild(uid)
		if child != nil {
			return child
		}
	}
	return nil
}

func GetTopConcepts() []*ConceptInfo {
	return cacheCtx.concerts
}

func CreateTopConcept(info *ConceptInfo) error {
	//if len(info.Table) < 1 {
	//	return errors.New("the table must not null")
	//}
	db := new(nosql.Concept)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetConceptNextID()
	db.CreatedTime = time.Now()
	db.Creator = info.Creator
	db.Name = info.Name
	db.Table = info.Table
	db.Cover = info.Cover
	db.Remark = info.Remark
	db.Parent = ""
	db.Scene = info.Scene
	db.Type = info.Type
	db.Attributes = make([]string, 0, 5)
	err := nosql.CreateConcept(db)
	if err == nil {
		info.initInfo(db)
		cacheCtx.concerts = append(cacheCtx.concerts, info)
	}
	return err
}

func RemoveConcept(uid, operator string) error {
	err := nosql.RemoveConcept(uid, operator)
	if err == nil {
		for i := 0; i < len(cacheCtx.concerts); i += 1 {
			if cacheCtx.concerts[i].UID == uid {
				cacheCtx.concerts = append(cacheCtx.concerts[:i], cacheCtx.concerts[i+1:]...)
				break
			}else if cacheCtx.concerts[i].HadChild(uid) {
				_ = cacheCtx.concerts[i].RemoveChild(uid)
			}
		}
	}
	return err
}

func HadConceptByTable(table string) bool {
	for i := 0; i < len(cacheCtx.concerts); i += 1 {
		if cacheCtx.concerts[i].Table == table {
			return true
		}
	}
	return false
}

func HadConceptByName(name string) bool {
	for i := 0; i < len(cacheCtx.concerts); i += 1 {
		if cacheCtx.concerts[i].Name == name {
			return true
		}
	}
	return false
}

func HadConceptProperty(uid, key string) bool {
	var had = false
	for i := 0;i < len(cacheCtx.concerts);i += 1 {
		if cacheCtx.concerts[i].HadChild(uid) {
			had = cacheCtx.concerts[i].HadAttribute(key)
			break
		}
	}
	return had
}
//endregion

//region Base Fun
func (mine *ConceptInfo) initInfo(db *nosql.Concept) {
	if db == nil {
		return
	}
	mine.ID = db.ID
	mine.UID = db.UID.Hex()
	mine.Name = db.Name
	mine.Remark = db.Remark
	mine.Table = db.Table
	mine.Cover = db.Cover
	mine.UpdateTime = db.UpdatedTime
	mine.CreateTime = db.CreatedTime
	mine.Operator = db.Operator
	mine.Creator = db.Creator
	mine.Parent = db.Parent
	mine.attributes = db.Attributes
	mine.Scene = db.Scene

	array, err := nosql.GetConceptsByParent(mine.UID)
	num := len(array)
	mine.children = make([]*ConceptInfo, 0, 5)
	if err == nil && num > 0 {
		for i := 0; i < num; i += 1 {
			tmp := ConceptInfo{}
			tmp.initInfo(array[i])
			mine.children = append(mine.children, &tmp)
		}
	}
}

func (mine *ConceptInfo) Children() []*ConceptInfo {
	return mine.children
}

func (mine *ConceptInfo) CreateChild(info *ConceptInfo) error {
	db := new(nosql.Concept)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetConceptNextID()
	db.CreatedTime = time.Now()
	db.Creator = info.Creator
	db.Name = info.Name
	db.Table = ""
	db.Cover = info.Cover
	db.Remark = info.Remark
	db.Parent = mine.UID
	db.Attributes = make([]string, 0, 5)
	err := nosql.CreateConcept(db)
	if err == nil {
		info.initInfo(db)
		mine.children = append(mine.children, info)
	}
	return err
}

func (mine *ConceptInfo)Label() string {
	if mine.Type == ConceptTypePersonal {
		return "personals"
	} else if mine.Type == ConceptTypeUtensil {
		return "utensils"
	} else if mine.Type == ConceptTypeEvent {
		return "events"
	} else if mine.Type == ConceptTypeOrganize {
		return "organizations"
	} else if mine.Type == ConceptTypeIdea {
		return "ideas"
	} else if mine.Type == ConceptTypeBook {
		return "books"
	} else if mine.Type == ConceptTypeCulture {
		return "culture"
	} else if mine.Type == ConceptTypeFaction {
		return "factions"
	} else if mine.Type == ConceptTypeNature {
		return "nature"
	} else if mine.Type == ConceptTypeHonor {
		return "honors"
	} else if mine.Type == ConceptTypePlace {
		return "places"
	} else if mine.Type == ConceptTypeEra {
		return "eras"
	} else {
		return "others"
	}
}

func (mine *ConceptInfo) RemoveChild(uid string) bool {
	for i := 0; i < len(mine.children); i += 1 {
		if mine.children[i].UID == uid {
			mine.children = append(mine.children[:i], mine.children[i+1:]...)
			return true
		}
		if mine.children[i].HadChild(uid) {
			return  mine.children[i].RemoveChild(uid)
		}
	}
	return false
}

func (mine *ConceptInfo) HadChild(uid string) bool {
	if mine.UID == uid {
		return true
	}
	for i := 0; i < len(mine.children); i += 1 {
		if mine.children[i].HadChild(uid) {
			return true
		}
	}
	return false
}

func (mine *ConceptInfo) GetChild(uid string) *ConceptInfo {
	if mine.UID == uid {
		return mine
	}
	for i := 0; i < len(mine.children); i += 1 {
		t := mine.children[i].GetChild(uid)
		if t != nil {
			return t
		}
	}
	return nil
}

func (mine *ConceptInfo) GetChildByName(name string) *ConceptInfo {
	if mine.Name == name {
		return mine
	}
	for i := 0; i < len(mine.children); i += 1 {
		t := mine.children[i].GetChildByName(name)
		if t != nil {
			return t
		}
	}
	return nil
}

func (mine *ConceptInfo) Attributes() []string {
	return mine.attributes
}

func (mine *ConceptInfo) CreateAttribute(key, val, begin, end string,kind AttributeType) error {
	if mine.attributes == nil {
		return errors.New("must call construct fist")
	}
	if HadAttributeByName(key) {
		return errors.New("the attribute name is repeated")
	}

	info := new(AttributeInfo)
	info.Key = key
	info.Name = val
	info.Kind = kind
	info.Begin = begin
	info.End = end
	var err error
	err = CreateAttribute(info)
	if err == nil {
		err = nosql.AppendConceptAttribute(mine.UID, info.UID)
		if err == nil {
			mine.attributes = append(mine.attributes, info.UID)
		}
	}
	return err
}

func (mine *ConceptInfo)UpdateAttributes(attributes []string) error {
	if attributes == nil {
		return errors.New("the attributes is nil when update")
	}

	err := nosql.UpdateConceptAttributes(mine.UID, attributes)
	if err == nil {
		mine.attributes = attributes
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *ConceptInfo)AppendAttribute(info *AttributeInfo) error {
	if info == nil {
		return errors.New("the attribute is nil when append")
	}
	if mine.HadAttributeByUID(info.UID) {
		return nil
	}
	err := nosql.AppendConceptAttribute(mine.UID, info.UID)
	if err == nil {
		mine.attributes = append(mine.attributes, info.UID)
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *ConceptInfo) GetAttributeName(key string) string {
	for i := 0;i < len(mine.attributes);i += 1 {
		t := GetAttribute(mine.attributes[i])
		if t != nil && t.Key == key {
			return t.Name
		}
	}
	return ""
}

func (mine *ConceptInfo) GetAttribute(key string) *AttributeInfo {
	for i := 0;i < len(mine.attributes);i += 1 {
		t := GetAttribute(mine.attributes[i])
		if t != nil && t.Key == key {
			return t
		}
	}
	return nil
}

func (mine *ConceptInfo) HadAttribute(key string) bool {
	if mine.attributes == nil {
		return false
	}
	for i := 0; i < len(mine.attributes); i += 1 {
		t := GetAttribute(mine.attributes[i])
		if t != nil && t.Key == key {
			return true
		}
	}
	for i := 0;i < len(mine.children);i += 1{
		if mine.children[i].HadAttribute(key) {
			return true
		}
	}
	return false
}

func (mine *ConceptInfo) HadAttributeByUID(uid string) bool {
	if mine.attributes == nil {
		return false
	}
	for i := 0; i < len(mine.attributes); i += 1 {
		if mine.attributes[i] == uid {
			return true
		}
	}
	return false
}

func (mine *ConceptInfo) RemoveAttribute(uid string) error {
	if mine.attributes == nil {
		return errors.New("must call construct fist")
	}
	if !mine.HadAttributeByUID(uid) {
		return errors.New("not found the property when remove")
	}
	err := nosql.SubtractConceptAttribute(mine.UID, uid)
	if err == nil {
		for i := 0; i < len(mine.attributes); i += 1 {
			if mine.attributes[i] == uid {
				mine.attributes = append(mine.attributes[:i], mine.attributes[i+1:]...)
				break
			}
		}
	}
	return err
}

func (mine *ConceptInfo) UpdateBase(name, remark,operator string, kind, scene uint8) error {
	err := nosql.UpdateConceptBase(mine.UID, name, remark, operator, kind, scene)
	if err == nil {
		mine.Name = name
		mine.Remark = remark
		mine.Operator = operator
		mine.Type = kind
		mine.Scene = scene
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *ConceptInfo) UpdateCover(cover string) error {
	err := nosql.UpdateConceptCover(mine.UID, cover)
	if err == nil {
		mine.Cover = cover
		mine.UpdateTime = time.Now()
	}
	return err
}
//endregion
