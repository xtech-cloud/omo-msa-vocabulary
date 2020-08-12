package cache

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

const (
	RelationUnknown RelationType = 0
	RelationPersons RelationType = 1 // 人对人
	RelationEvents  RelationType = 2 // 人与非人
	RelationInhuman RelationType = 3 // 非人与非人
)

type RelationType uint8

type RelationshipInfo struct {
	BaseInfo
	Kind     RelationType
	Remark   string
	Custom   bool
	Parent   string
	children []*RelationshipInfo
}

func AllRelations() []*RelationshipInfo {
	return cacheCtx.relations
}

func CreateRelation(parent, creator string, info *RelationshipInfo) error {
	if info == nil {
		return errors.New("the attribute info is nil")
	}
	db := new(nosql.Relation)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetRelationNextID()
	db.CreatedTime = time.Now()
	db.Creator = creator
	db.Name = info.Name
	db.Remark = info.Remark
	db.Type = uint8(info.Kind)
	db.Parent = parent
	db.Custom = info.Custom
	err := nosql.CreateRelation(db)
	if err == nil {
		info.initInfo(db)
	}
	if len(parent) > 0 {
		top := GetRelation(parent)
		top.children = append(top.children, info)
	} else {
		cacheCtx.relations = append(cacheCtx.relations, info)
	}

	return err
}

func HadRelation(uid string) bool {
	for i := 0; i < len(cacheCtx.relations); i += 1 {
		if cacheCtx.relations[i].UID == uid {
			return true
		}
	}
	return false
}

func HadRelationByName(name string) bool {
	for i := 0; i < len(cacheCtx.relations); i += 1 {
		if cacheCtx.relations[i].Name == name {
			return true
		}
	}
	return false
}

func switchRelationToLink(kind RelationType) LinkType {
	if kind == RelationEvents {
		return LinkTypeEvents
	} else if kind == RelationPersons {
		return LinkTypePersons
	}else if kind == RelationInhuman {
		return LinkTypeInhuman
	}else{
		return LinkTypeEmpty
	}
}

func RemoveRelation(uid, operator string) error {
	err := nosql.RemoveRelation(uid, operator)
	if err == nil {
		for i := 0; i < len(cacheCtx.relations); i += 1 {
			if cacheCtx.relations[i].UID == uid {
				cacheCtx.relations = append(cacheCtx.relations[:i], cacheCtx.relations[i+1:]...)
				break
			} else if cacheCtx.relations[i].HadChild(uid) {
				_ = cacheCtx.relations[i].RemoveChild(uid, operator)
			}
		}
	}
	return err
}

func GetRelation(uid string) *RelationshipInfo {
	for i := 0; i < len(cacheCtx.relations); i += 1 {
		child := cacheCtx.relations[i].GetChild(uid)
		if child != nil {
			return child
		}
	}
	return nil
}

func (mine *RelationshipInfo) initInfo(db *nosql.Relation) {
	mine.UID = db.UID.Hex()
	mine.ID = db.ID
	mine.Name = db.Name
	mine.Remark = db.Remark
	mine.CreateTime = db.CreatedTime
	mine.UpdateTime = db.UpdatedTime
	mine.Creator = db.Creator
	mine.Operator = db.Operator
	mine.Kind = RelationType(db.Type)
	mine.Parent = db.Parent
	array, err := nosql.GetRelationsByParent(mine.UID)
	num := len(array)
	mine.children = make([]*RelationshipInfo, 0, 5)
	if err == nil && num > 0 {
		for i := 0; i < num; i += 1 {
			tmp := RelationshipInfo{}
			tmp.initInfo(array[i])
			mine.children = append(mine.children, &tmp)
		}
	}
}

func (mine *RelationshipInfo) UpdateBase(name, remark, operator string, custom bool, kind uint8) error {
	if len(name) < 1 {
		name = mine.Name
	}
	if len(remark) < 1 {
		remark = mine.Remark
	}
	err := nosql.UpdateRelationBase(mine.UID, name, remark, operator, custom, kind)
	if err == nil {
		mine.Name = name
		mine.Remark = remark
		mine.Operator = operator
		mine.Kind = RelationType(kind)
		mine.Custom = custom
	}
	return err
}

func (mine *RelationshipInfo) deleteChild(uid string) bool {
	for i := 0; i < len(mine.children); i += 1 {
		if mine.children[i].UID == uid {
			mine.children = append(mine.children[:i], mine.children[i+1:]...)
			return true
		}
		if mine.children[i].HadChild(uid) {
			return mine.children[i].deleteChild(uid)
		}
	}
	return false
}

func (mine *RelationshipInfo) HadChild(uid string) bool {
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

func (mine *RelationshipInfo) GetChild(uid string) *RelationshipInfo {
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

func (mine *RelationshipInfo) Children() []*RelationshipInfo {
	return mine.children
}

func (mine *RelationshipInfo) RemoveChild(uid, operator string) error {
	err := nosql.RemoveRelation(uid, operator)
	if err == nil {
		mine.deleteChild(uid)
	}
	return err
}
