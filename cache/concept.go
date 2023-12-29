package cache

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

const (
	ConceptTypeUnknown  = 0
	ConceptTypePersonal = 1  //人物
	ConceptTypeUtensil  = 2  // 器物
	ConceptTypeEvent    = 3  //事件
	ConceptTypeOrganize = 4  // 组织机构
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
	Type       uint8
	Cover      string
	Remark     string
	Table      string
	Parent     string
	Scene      uint8 //针对的场景类型
	Count      int
	attributes []string //所有支持的属性
	privates   []string //隐藏属性
	Children   []*ConceptInfo
}

//region Global Fun

func (mine *cacheContext) GetTopConcept(uid string) *ConceptInfo {
	tops := mine.GetTopConcepts()
	for i := 0; i < len(tops); i += 1 {
		if tops[i].HadChild(uid) {
			return tops[i]
		}
	}
	return nil
}

func (mine *cacheContext) GetConceptByName(name string) *ConceptInfo {
	tops := mine.GetTopConcepts()
	for i := 0; i < len(tops); i += 1 {
		child := tops[i].GetChildByName(name)
		if child != nil {
			return child
		}
	}
	return nil
}

func (mine *cacheContext) GetConceptsByAttribute(uid string) []*ConceptInfo {
	dbs, err := nosql.GetConceptsByAttribute(uid)
	list := make([]*ConceptInfo, 0, len(dbs))
	if err == nil {
		for _, db := range dbs {
			info := new(ConceptInfo)
			info.initInfo(db)
			list = append(list, info)
		}
	}
	return list
}

func (mine *cacheContext) GetConcept(uid string) *ConceptInfo {
	if uid == "" {
		return nil
	}
	db, _ := nosql.GetConcept(uid)
	if db != nil {
		tmp := new(ConceptInfo)
		tmp.initInfo(db)
		return tmp
	}
	return nil
}

func (mine *cacheContext) GetTopConcepts() []*ConceptInfo {
	dbs, _ := nosql.GetTopConcepts()
	all := make([]*ConceptInfo, 0, len(dbs))
	for _, db := range dbs {
		tmp := new(ConceptInfo)
		tmp.initInfo(db)
		all = append(all, tmp)
	}
	return all
}

func (mine *cacheContext) CreateTopConcept(info *ConceptInfo) error {
	//if len(info.Table) < 1 {
	//	return errors.New("the table must not null")
	//}
	db := new(nosql.Concept)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetConceptNextID()
	db.Created = time.Now().Unix()
	db.Creator = info.Creator
	db.Name = info.Name
	db.Table = info.Table
	db.Cover = info.Cover
	db.Remark = info.Remark
	db.Parent = ""
	db.Scene = info.Scene
	db.Type = info.Type
	db.Attributes = make([]string, 0, 1)
	err := nosql.CreateConcept(db)
	if err == nil {
		info.initInfo(db)
	}
	return err
}

func (mine *cacheContext) RemoveConcept(uid, operator string) error {
	if uid == "" {
		return errors.New("the concept uid is empty")
	}
	err := nosql.RemoveConcept(uid, operator)
	if err == nil {
		//for i := 0; i < len(mine.concepts); i += 1 {
		//	if mine.concepts[i].UID == uid {
		//		mine.concepts = append(mine.concepts[:i], mine.concepts[i+1:]...)
		//		break
		//	} else if mine.concepts[i].HadChild(uid) {
		//		_ = mine.concepts[i].RemoveChild(uid)
		//	}
		//}
	}
	return err
}

func (mine *cacheContext) HadConceptByTable(table string) bool {
	dbs, _ := nosql.GetTopConcepts()
	for _, db := range dbs {
		if db.Table == table {
			return true
		}
	}
	return false
}

func (mine *cacheContext) HadConceptByName(name, parent string) bool {
	if parent == "" {
		dbs, _ := nosql.GetTopConcepts()
		for i := 0; i < len(dbs); i += 1 {
			if dbs[i].Name == name {
				return true
			}
		}
	} else {
		p := mine.GetConcept(parent)
		if p != nil && p.HadChildByName(name) {
			return true
		}
	}

	return false
}

func (mine *cacheContext) HadConceptProperty(uid, key string) bool {
	var had = false
	dbs, _ := nosql.GetTopConcepts()
	for i := 0; i < len(dbs); i += 1 {
		tmp := new(ConceptInfo)
		tmp.initInfo(dbs[i])
		if tmp.HadChild(uid) {
			had = tmp.HadAttribute(key)
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
	mine.Updated = db.Updated
	mine.Created = db.Created
	mine.Operator = db.Operator
	mine.Creator = db.Creator
	mine.Parent = db.Parent
	mine.attributes = db.Attributes
	mine.Scene = db.Scene
	mine.privates = db.Privates
	mine.Count = cacheCtx.GetEntitiesCountByConcept(mine.UID)
	if len(mine.Table) < 2 && len(mine.Parent) < 2 {
		mine.Table = DefaultEntityTable
	}
	array, err := nosql.GetConceptsByParent(mine.UID)
	num := len(array)
	mine.Children = make([]*ConceptInfo, 0, 5)
	if err == nil && num > 0 {
		for i := 0; i < num; i += 1 {
			tmp := ConceptInfo{}
			tmp.initInfo(array[i])
			mine.Children = append(mine.Children, &tmp)
		}
	}
}

func (mine *ConceptInfo) CreateChild(info *ConceptInfo) error {
	db := new(nosql.Concept)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetConceptNextID()
	db.Created = time.Now().Unix()
	db.CreatedTime = time.Now()
	db.Creator = info.Creator
	db.Name = info.Name
	db.Table = ""
	db.Cover = info.Cover
	db.Remark = info.Remark
	db.Parent = mine.UID
	db.Scene = info.Scene
	db.Attributes = make([]string, 0, 5)
	db.Privates = make([]string, 0, 1)
	err := nosql.CreateConcept(db)
	if err == nil {
		info.initInfo(db)
		mine.Children = append(mine.Children, info)
	}
	return err
}

func (mine *ConceptInfo) Label() string {
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
		return DefaultEntityTable
	}
}

func (mine *ConceptInfo) RemoveChild(uid string) bool {
	for i := 0; i < len(mine.Children); i += 1 {
		if mine.Children[i].UID == uid {
			mine.Children = append(mine.Children[:i], mine.Children[i+1:]...)
			return true
		}
		if mine.Children[i].HadChild(uid) {
			return mine.Children[i].RemoveChild(uid)
		}
	}
	return false
}

func (mine *ConceptInfo) HadChild(uid string) bool {
	if mine.UID == uid {
		return true
	}
	for i := 0; i < len(mine.Children); i += 1 {
		if mine.Children[i].HadChild(uid) {
			return true
		}
	}
	return false
}

func (mine *ConceptInfo) HadChildByName(name string) bool {
	if mine.Name == name {
		return true
	}
	for i := 0; i < len(mine.Children); i += 1 {
		if mine.Children[i].HadChildByName(name) {
			return true
		}
	}
	return false
}

func (mine *ConceptInfo) GetChild(uid string) *ConceptInfo {
	if mine.UID == uid {
		return mine
	}
	for i := 0; i < len(mine.Children); i += 1 {
		t := mine.Children[i].GetChild(uid)
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
	for i := 0; i < len(mine.Children); i += 1 {
		t := mine.Children[i].GetChildByName(name)
		if t != nil {
			return t
		}
	}
	return nil
}

func (mine *ConceptInfo) Attributes() []string {
	return mine.attributes
}

func (mine *ConceptInfo) Privates() []string {
	return mine.privates
}

func (mine *ConceptInfo) CreateAttribute(key, val, begin, end string, kind AttributeType) error {
	if mine.attributes == nil {
		return errors.New("must call construct fist")
	}
	if Context().HadAttributeByName(key) {
		return errors.New("the attribute name is repeated")
	}

	info := new(AttributeInfo)
	info.Key = key
	info.Name = val
	info.Kind = kind
	info.Begin = begin
	info.End = end
	var err error
	err = Context().CreateAttribute(info)
	if err == nil {
		err = nosql.AppendConceptAttribute(mine.UID, info.UID)
		if err == nil {
			mine.attributes = append(mine.attributes, info.UID)
		}
	}
	return err
}

func (mine *ConceptInfo) UpdateAttributes(arr []string) error {
	if arr == nil {
		arr = make([]string, 0, 1)
	}

	err := nosql.UpdateConceptAttributes(mine.UID, arr)
	if err == nil {
		mine.attributes = arr
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *ConceptInfo) ReplaceAttributes(old, news string) error {
	arr := make([]string, 0, len(mine.attributes))
	arr = append(arr, mine.attributes...)
	for i, item := range arr {
		if item == old {
			if i == len(arr)-1 {
				arr = append(arr[:i])
			} else {
				arr = append(arr[:i], arr[i+1:]...)
			}
		}
	}
	arr = append(arr, news)
	err := nosql.UpdateConceptAttributes(mine.UID, arr)
	if err == nil {
		mine.attributes = arr
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *ConceptInfo) UpdatePrivates(arr []string) error {
	if arr == nil {
		arr = make([]string, 0, 1)
	}

	err := nosql.UpdateConceptPrivates(mine.UID, arr)
	if err == nil {
		mine.privates = arr
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *ConceptInfo) AppendAttribute(info *AttributeInfo) error {
	if info == nil {
		return errors.New("the attribute is nil when append")
	}
	if mine.HadAttributeByUID(info.UID) {
		return nil
	}
	err := nosql.AppendConceptAttribute(mine.UID, info.UID)
	if err == nil {
		mine.attributes = append(mine.attributes, info.UID)
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *ConceptInfo) GetAttributeName(key string) string {
	for i := 0; i < len(mine.attributes); i += 1 {
		t := Context().GetAttribute(mine.attributes[i])
		if t != nil && t.Key == key {
			return t.Name
		}
	}
	return ""
}

func (mine *ConceptInfo) GetAttribute(key string) *AttributeInfo {
	for i := 0; i < len(mine.attributes); i += 1 {
		t := Context().GetAttribute(mine.attributes[i])
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
		t := Context().GetAttribute(mine.attributes[i])
		if t != nil && t.Key == key {
			return true
		}
	}
	for i := 0; i < len(mine.Children); i += 1 {
		if mine.Children[i].HadAttribute(key) {
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
				if i == len(mine.attributes)-1 {
					mine.attributes = append(mine.attributes[:i])
				} else {
					mine.attributes = append(mine.attributes[:i], mine.attributes[i+1:]...)
				}
				break
			}
		}
	}
	return err
}

func (mine *ConceptInfo) UpdateBase(name, remark, operator string, kind, scene uint8) error {
	err := nosql.UpdateConceptBase(mine.UID, name, remark, operator, kind, scene)
	if err == nil {
		mine.Name = name
		mine.Remark = remark
		mine.Operator = operator
		mine.Type = kind
		mine.Scene = scene
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *ConceptInfo) UpdateCover(cover string) error {
	err := nosql.UpdateConceptCover(mine.UID, cover)
	if err == nil {
		mine.Cover = cover
		mine.Updated = time.Now().Unix()
	}
	return err
}

//endregion
