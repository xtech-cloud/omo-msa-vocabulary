package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy/nosql"
	"omo.msa.vocabulary/tool"
	"time"
)

type ArchivedInfo struct {
	BaseInfo
	Concept string
	Entity string
	File string
	MD5 string
	Scene string
}

func (mine *cacheContext)CreateArchived(info *EntityInfo) error {
	if info == nil {
		return errors.New("the entity info is nil")
	}
	db := new(nosql.Archived)
	db.UID = primitive.NewObjectID()
	db.CreatedTime = time.Now()
	db.ID = nosql.GetEntityNextID(info.table())
	db.Name = fmt.Sprintf("%s(%s)",info.Name, info.Add)
	db.Concept = info.Concept
	db.Entity = info.UID
	db.Scene = info.Owner
	db.Creator = info.Creator
	db.Operator = info.Operator
	file,er := json.Marshal(info)
	if er != nil {
		return er
	}
	db.File = string(file)
	db.MD5 = tool.CalculateMD5(file)
	er = nosql.CreateArchived(db)
	return er
}

func (mine *cacheContext)GetArchivedByEntity(entity string) *ArchivedInfo {
	if len(entity) < 1 {
		return nil
	}
	db,err := nosql.GetArchivedByEntity(entity)
	if err == nil && db != nil {
		info := new(ArchivedInfo)
		info.initInfo(db)
		return info
	}
	return nil
}

func (mine *cacheContext)GetArchivedList(concept string) []*EntityInfo {
	var array []*nosql.Archived
	var err error
	if len(concept) > 1 {
		array,err = nosql.GetArchivedItems(concept)
	}else{
		array,err = nosql.GetAllArchived()
	}
	if err != nil {
		return make([]*EntityInfo, 0, 1)
	}
	list := make([]*EntityInfo, 0, len(array))
	for _, info := range array {
		entity := new(EntityInfo)
		er := json.Unmarshal([]byte(info.File), entity)
		if er == nil {
			entity.Status = EntityStatusUsable
			list = append(list, entity)
		}
	}
	return list
}

func (mine *cacheContext)GetArchivedEntities(scene, concept string) []*EntityInfo {
	var array []*nosql.Archived
	var err error
	if len(scene) > 1 && len(concept) > 1 {
		array,err = nosql.GetArchivedListBy(scene, concept)
	}else if len(scene) > 1 {
		array,err = nosql.GetArchivedListByScene(scene)
	}else if len(concept) > 1 {
		array,err = nosql.GetArchivedItems(concept)
	}else{
		array,err = nosql.GetAllArchived()
	}
	if err != nil {
		return make([]*EntityInfo, 0, 1)
	}
	list := make([]*EntityInfo, 0, len(array))
	for _, info := range array {
		entity := new(EntityInfo)
		er := json.Unmarshal([]byte(info.File), entity)
		if er == nil {
			entity.Status = EntityStatusUsable
			list = append(list, entity)
		}
	}
	return list
}

func (mine *ArchivedInfo) initInfo(db *nosql.Archived) bool {
	if db == nil {
		return false
	}
	mine.UID = db.UID.Hex()
	mine.ID = db.ID
	mine.CreateTime = db.CreatedTime
	mine.UpdateTime = db.UpdatedTime
	mine.Creator = db.Creator
	mine.Operator = db.Operator
	mine.Concept = db.Concept
	mine.Name = db.Name
	mine.Entity = db.Entity
	mine.File = db.File
	mine.MD5 = db.MD5
	mine.Scene = db.Scene
	return true
}

func (mine *ArchivedInfo)UpdateFile(info *EntityInfo, operator string) error {
	if info == nil {
		return errors.New("the entity info is nil")
	}
	data,er := json.Marshal(info)
	if er != nil {
		return er
	}
	md5 := tool.CalculateMD5(data)
	err := nosql.UpdateArchivedFile(mine.UID, operator, string(data), md5)
	if err == nil {
		mine.File = string(data)
		mine.MD5 = md5
	}
	return err
}


