package cache

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

type RelationshipInfo struct {
	BaseInfo
	Kind uint8
	Remark string
	Custom bool
	Parent string
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
	db.Type = info.Kind
	db.Parent = parent
	db.Custom = info.Custom
	err := nosql.CreateRelation(db)
	if err == nil {
		info.initInfo(db)
	}
	if len(parent) > 0 {
		 top := GetRelation(parent)
		 top.children = append(top.children, info)
	}else{
		cacheCtx.relations = append(cacheCtx.relations, info)
	}

	return err
}

func HadRelation(uid string) bool {
	for i := 0;i < len(cacheCtx.attributes);i += 1 {
		if cacheCtx.relations[i].UID == uid {
			return true
		}
	}
	return false
}

func HadRelationByName(name string) bool {
	for i := 0;i < len(cacheCtx.attributes);i += 1 {
		if cacheCtx.relations[i].Name == name {
			return true
		}
	}
	return false
}

func RemoveRelation(uid, operator string) error {
	err := nosql.RemoveRelation(uid, operator)
	if err == nil {
		for i := 0;i < len(cacheCtx.relations);i += 1 {
			if cacheCtx.relations[i].UID == uid {
				cacheCtx.relations = append(cacheCtx.relations[:i], cacheCtx.relations[i+1:]...)
				break
			}
		}
	}
	return err
}

func GetRelation(uid string) *RelationshipInfo {
	for _, value := range cacheCtx.relations {
		if value.UID == uid {
			return value
		}
	}
	return nil
}

func (mine *RelationshipInfo)initInfo(db *nosql.Relation)  {
	mine.UID = db.UID.Hex()
	mine.ID = db.ID
	mine.Name = db.Name
	mine.Remark = db.Remark
	mine.CreateTime = db.CreatedTime
	mine.UpdateTime = db.UpdatedTime
	mine.Creator = db.Creator
	mine.Operator = db.Operator
	mine.Kind = db.Type
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

func (mine *RelationshipInfo)UpdateBase(name, remark, operator string, custom bool, kind uint8) error {
	if len(name) < 1{
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
		mine.Custom = custom
	}
	return err
}

func (mine *RelationshipInfo)Children() []*RelationshipInfo {
	return mine.children
}

func (mine *RelationshipInfo)RemoveChild(uid, operator string) error {
	err := nosql.RemoveRelation(uid, operator)
	if err == nil {
		for i := 0;i < len(mine.children);i += 1 {
			if mine.children[i].UID == uid {
				mine.children = append(mine.children[:i], mine.children[i+1:]...)
				break
			}
		}
	}
	return err
}
