package cache

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy/nosql"
	"strings"
	"time"
)

const (
	AttributeTypeString  AttributeType = 0
	AttributeTypeDate    AttributeType = 1
	AttributeTypeNumber  AttributeType = 2
	AttributeTypeEntity  AttributeType = 3
	AttributeTypeSex     AttributeType = 4
	AttributeTypeAddress AttributeType = 5
)

type AttributeType uint8

type AttributeInfo struct {
	BaseInfo

	Kind   AttributeType
	Key    string
	Name   string
	Remark string
	Begin  string
	End    string
}

func (mine *cacheContext) CreateAttribute(info *AttributeInfo) error {
	if info == nil {
		return errors.New("the attribute info is nil")
	}
	db := new(nosql.Attribute)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetAttributeNextID()
	db.Created = time.Now().Unix()
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
	}
	return err
}

func (mine *cacheContext) AllAttributes() []*AttributeInfo {
	dbs, er := nosql.GetAllAttributes()
	if er == nil {
		all := make([]*AttributeInfo, 0, len(dbs))
		for _, db := range dbs {
			info := new(AttributeInfo)
			info.initInfo(db)
			all = append(all, info)
		}
		return all
	} else {
		return make([]*AttributeInfo, 0, 1)
	}
}

func (mine *cacheContext) HadAttributeByName(name string) bool {
	if name == "" {
		return true
	}
	db, err := nosql.GetAttributeByName(name)
	if err != nil {
		if strings.Contains(err.Error(), "no documents") {
			return false
		}
		return true
	}
	if db == nil {
		return false
	}
	return true
}

func (mine *cacheContext) HadAttributeByKey(key string) bool {
	if key == "" {
		return true
	}
	db, err := nosql.GetAttributeByKey(key)
	if err != nil {
		if strings.Contains(err.Error(), "no documents") {
			return false
		}
		return true
	}
	if db == nil {
		return false
	}
	return true
}

func (mine *cacheContext) GetAttribute(uid string) *AttributeInfo {
	if uid == "" {
		return nil
	}
	db, err := nosql.GetAttribute(uid)
	if err != nil {
		return nil
	}
	tmp := new(AttributeInfo)
	tmp.initInfo(db)
	return tmp
}

func (mine *cacheContext) GetAttributeByKey(key string) *AttributeInfo {
	if key == "" {
		return nil
	}
	db, err := nosql.GetAttributeByKey(strings.ToLower(key))
	if err != nil {
		return nil
	}
	if db == nil {
		return nil
	}
	tmp := new(AttributeInfo)
	tmp.initInfo(db)
	return tmp
}

func (mine *cacheContext) GetAttributeByName(key string) *AttributeInfo {
	k := strings.TrimSpace(key)
	if k == "" {
		return nil
	}
	db, err := nosql.GetAttributeByName(k)
	if err != nil {
		return nil
	}
	tmp := new(AttributeInfo)
	tmp.initInfo(db)
	return tmp
}

func (mine *cacheContext) RemoveAttribute(uid, operator string) error {
	if len(uid) < 1 {
		return errors.New("the attribute uid is empty")
	}
	err := nosql.RemoveAttribute(uid, operator)
	return err
}

func (mine *AttributeInfo) initInfo(db *nosql.Attribute) {
	mine.UID = db.UID.Hex()
	mine.ID = db.ID
	mine.Name = db.Name
	mine.Key = db.Key
	mine.Remark = db.Remark
	mine.Kind = AttributeType(db.Kind)
	mine.Begin = db.Begin
	mine.End = db.End
	mine.Created = db.Created
	mine.Updated = db.Updated
}

func (mine *AttributeInfo) UpdateBase(name, remark, begin, end, operator string, kind uint8) error {
	err := nosql.UpdateAttributeBase(mine.UID, name, remark, begin, end, operator, kind)
	if err == nil {
		mine.Name = name
		mine.Remark = remark
		mine.Begin = begin
		mine.End = end
		mine.Kind = AttributeType(kind)
		mine.Operator = operator
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *AttributeInfo) UpdateKey(key, operator string) error {
	err := nosql.UpdateAttributeKey(mine.UID, key, operator)
	if err == nil {
		mine.Key = key
		mine.Operator = operator
		mine.Updated = time.Now().Unix()
	}
	return err
}
