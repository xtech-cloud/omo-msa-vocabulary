package cache

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

const (
	AttributeTypeString AttributeType = 0
	AttributeTypeDate    AttributeType = 1
	AttributeTypeNumber  AttributeType = 2
	AttributeTypeEntity  AttributeType = 3
	AttributeTypeSex     AttributeType = 4
	AttributeTypeAddress AttributeType = 5
)

type AttributeType uint8

type AttributeInfo struct {
	BaseInfo

	Kind  AttributeType
	Key   string
	Name  string
	Remark string
	Begin string
	End   string
}

func (mine *cacheContext)CreateAttribute(info *AttributeInfo) error {
	if info == nil {
		return errors.New("the attribute info is nil")
	}
	db := new(nosql.Attribute)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetAttributeNextID()
	db.CreatedTime = time.Now()
	db.Creator = info.Creator
	db.Key = info.Key
	db.Name = info.Name
	db.Kind = uint8(info.Kind)
	db.Begin = info.Begin
	db.End = info.End
	db.Remark = info.Remark
	err := nosql.CreateAttribute(db)
	if err == nil {
		info.initInfo(db)
		mine.attributes = append(mine.attributes, info)
	}
	return err
}

func (mine *cacheContext)AllAttributes() []*AttributeInfo {
	return mine.attributes
}

func (mine *cacheContext)HadAttributeByName(name string) bool {
	for i := 0;i < len(mine.attributes);i += 1 {
		if mine.attributes[i].Name == name {
			return true
		}
	}
	return false
}

func (mine *cacheContext)GetAttribute(uid string) *AttributeInfo {
	for _, value := range mine.attributes {
		if value.UID == uid {
			return value
		}
	}
	return nil
}

func (mine *cacheContext)GetAttributeByKey(key string) *AttributeInfo {
	for _, value := range mine.attributes {
		if value.Key == key {
			return value
		}
	}
	return nil
}

func (mine *cacheContext)RemoveAttribute(uid, operator string) error {
	if len(uid) <  1 {
		return errors.New("the attribute uid is empty")
	}
	err := nosql.RemoveAttribute(uid, operator)
	if err == nil {
		for i := 0;i < len(mine.attributes);i +=1 {
			if mine.attributes[i].UID == uid {
				mine.attributes = append(mine.attributes[:i], mine.attributes[i+1:]...)
				break
			}
		}
	}
	return err
}

func (mine *AttributeInfo)initInfo(db *nosql.Attribute)  {
	mine.UID = db.UID.Hex()
	mine.ID = db.ID
	mine.Name = db.Name
	mine.Key = db.Key
	mine.Remark = db.Remark
	mine.Kind = AttributeType(db.Kind)
	mine.Begin = db.Begin
	mine.End = db.End
	mine.CreateTime = db.CreatedTime
	mine.UpdateTime = db.UpdatedTime
}

func (mine *AttributeInfo)UpdateBase(name, remark, begin, end, operator string, kind uint8) error {
	err := nosql.UpdateAttributeBase(mine.UID, name, remark, begin, end, operator, kind)
	if err == nil {
		mine.Name = name
		mine.Remark = remark
		mine.Begin = begin
		mine.End = end
		mine.Kind = AttributeType(kind)
		mine.Operator = operator
		mine.UpdateTime = time.Now()
	}
	return err
}