package cache

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

const (
	TemplateStudent TemplateType = 0
	TemplateTeacher TemplateType = 1 
)

type TemplateType uint8

type TemplateInfo struct {
	BaseInfo
	Url        string `json:"url"`
	Comments   string `json:"comments"`
	Type       TemplateType
	SkipRows   int32;
    Columns    map[string]int32;
}

func (mine *TemplateInfo) initInfo(db *nosql.Template) {
	mine.UID = db.UID.Hex()
	mine.ID = db.ID
	mine.Name = db.Name
	mine.Url = db.Url
	mine.Type = TemplateType(db.Type)
	mine.SkipRows = db.SkipRows
	mine.Columns = db.Columns
	mine.Comments = db.Comments
	mine.Created = db.Created
	mine.Updated = db.Updated
	mine.Creator = db.Creator
	mine.Operator = db.Operator
}

func (mine *cacheContext) AllTemplates() []*TemplateInfo {
	dbs, _ := nosql.GetTemplates()
	all := make([]*TemplateInfo, 0, len(dbs))
	for _, db := range dbs {
		tmp := new(TemplateInfo)
		tmp.initInfo(db)
		all = append(all, tmp)
	}
	return all
}

func (mine *cacheContext) HadTemplateByName(name string) bool {
	db, _ := nosql.GetTemplateByName(name)
	if db == nil {
		return false
	} else {
		return true
	}
}

func (mine *cacheContext) GetTemplateByName(name string) *TemplateInfo {
	db, _ := nosql.GetTemplateByName(name)
	if db != nil {
		tmp := new(TemplateInfo)
		tmp.initInfo(db)
		return tmp
	}
	return nil
}

func (mine *cacheContext) GetTemplate(uid string) *TemplateInfo {
	db, _ := nosql.GetTemplate(uid)
	if db != nil {
		tmp := new(TemplateInfo)
		tmp.initInfo(db)
		return tmp
	}
	return nil
}

func (mine *cacheContext) CreateTemplate(creator string, info *TemplateInfo) error {
	if info == nil {
		return errors.New("the attribute info is nil")
	}
	db := new(nosql.Template)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetTemplateNextID()
	db.Created = time.Now().Unix()
	db.CreatedTime = time.Now()
	db.Creator = creator
	db.Name = info.Name
	db.Type = uint8(info.Type)
	db.SkipRows = info.SkipRows
	db.Columns = info.Columns
	db.Url = info.Url
	db.Comments = info.Comments
	err := nosql.CreateTemplate(db)
	if err == nil {
		info.initInfo(db)
	}
	return err
}

func (mine *cacheContext) RemoveTemplate(uid, operator string) error {
	if uid == "" {
		return errors.New("the template uid is empty")
	}
	err := nosql.RemoveTemplate(uid, operator)
	if err == nil {
	}
	return err
}

func (mine *TemplateInfo) UpdateBase(name string, _type uint8, skipRows int32, columns map[string]int32, url, comments, operator string) error {
	if len(name) < 1 {
		name = mine.Name
	}
	err := nosql.UpdateTemplateBase(mine.UID, name, _type, skipRows, columns, url, comments, operator)
	if err == nil {
		mine.Name = name
		mine.Type = TemplateType(_type)
		mine.SkipRows = skipRows
		mine.Columns = columns
		mine.Url = url
		mine.Comments = comments
		mine.Updated = time.Now().Unix()
		mine.Operator = operator
	}
	return err
}
