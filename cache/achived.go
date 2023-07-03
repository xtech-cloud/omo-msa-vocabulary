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
	Access uint8
	BaseInfo
	Concept string
	Entity  string
	File    string
	MD5     string
	Scene   string
	Score   uint32
}

func (mine *cacheContext) CreateArchived(info *EntityInfo) error {
	if info == nil {
		return errors.New("the entity info is nil")
	}
	db := new(nosql.Archived)
	db.UID = primitive.NewObjectID()
	db.CreatedTime = time.Now()
	db.ID = nosql.GetEntityNextID(info.table())
	db.Name = fmt.Sprintf("%s(%s)", info.Name, info.Add)
	db.Concept = info.Concept
	db.Entity = info.UID
	db.Scene = info.Owner
	db.Creator = info.Creator
	db.Operator = info.Operator
	db.Access = 0
	db.Score = 0
	file, er := json.Marshal(info)
	if er != nil {
		return er
	}
	db.File = string(file)
	db.MD5 = tool.CalculateMD5(file)
	er = nosql.CreateArchived(db)
	return er
}

func (mine *cacheContext) GetArchivedByEntity(entity string) *ArchivedInfo {
	if len(entity) < 1 {
		return nil
	}
	db, err := nosql.GetArchivedByEntity(entity)
	if err == nil && db != nil {
		info := new(ArchivedInfo)
		info.initInfo(db)
		return info
	}
	return nil
}

func (mine *cacheContext) HadArchivedByEntity(entity string) bool {
	if len(entity) < 1 {
		return false
	}
	return nosql.HadArchivedItem(entity)
}

func (mine *cacheContext) GetArchivedByList(list []string) ([]*EntityInfo, error) {
	if list == nil || len(list) < 1 {
		return nil, errors.New("the array is empty")
	}
	arr := make([]*EntityInfo, 0, len(list))
	for _, key := range list {
		info := mine.GetArchivedByEntity(key)
		if info != nil {
			arr = append(arr, info.GetEntity())
		}
	}
	return arr, nil
}

func (mine *cacheContext) GetArchivedList(name string) []*EntityInfo {
	var array []*nosql.Archived
	var err error
	if len(name) > 1 {
		array, err = nosql.GetArchivedItems(name)
	} else {
		array, err = nosql.GetAllArchived()
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

func (mine *cacheContext) GetArchivedEntities(scene, concept string) []*EntityInfo {
	var array []*nosql.Archived
	var err error
	if len(scene) > 1 && len(concept) > 1 {
		array, err = nosql.GetArchivedListBy(scene, concept)
	} else if len(scene) > 1 {
		array, err = nosql.GetArchivedListByScene(scene)
	} else if len(concept) > 1 {
		array, err = nosql.GetArchivedItems(concept)
	} else {
		array, err = nosql.GetAllArchived()
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
	mine.Access = db.Access
	//if strings.Contains(mine.File,"http://rdp-down.suii.cn/") {
	//	f := strings.Replace(mine.File, "http://rdp-down.suii.cn/", "", 1)
	//	_ = mine.setFile(f)
	//}
	return true
}

func (mine *ArchivedInfo) setFile(file string) error {
	md5 := tool.CalculateMD5([]byte(file))
	err := nosql.UpdateArchivedFile(mine.UID, mine.Operator, file, md5)
	if err == nil {
		mine.File = file
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *ArchivedInfo) UpdateFile(info *EntityInfo, operator string) error {
	if info == nil {
		return errors.New("the entity info is nil")
	}
	data, er := json.Marshal(info)
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

func (mine *ArchivedInfo) UpdateAccess(operator string, acc uint8) error {
	err := nosql.UpdateArchivedAccess(mine.UID, operator, acc)
	if err == nil {
		mine.Access = acc
		mine.Operator = operator
	}
	return err
}

func (mine *ArchivedInfo) GetEntity() *EntityInfo {
	entity := new(EntityInfo)
	er := json.Unmarshal([]byte(mine.File), entity)
	if er != nil {
		return nil
	}
	now := cacheCtx.GetEntity(entity.UID)
	entity.Status = now.Status
	entity.Published = true
	entity.Access = mine.Access
	return entity
}
