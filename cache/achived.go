package cache

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/micro/go-micro/v2/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy/nosql"
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
	Size    uint32
}

func (mine *cacheContext) CreateArchived(info *EntityInfo) error {
	if info == nil {
		return errors.New("the entity info is nil")
	}
	db := new(nosql.Archived)
	db.UID = primitive.NewObjectID()
	db.Created = time.Now().Unix()
	db.CreatedTime = time.Now()
	db.ID = nosql.GetArchivedNextID()
	db.Name = fmt.Sprintf("%s(%s)", info.Name, info.Add)
	db.Concept = info.Concept
	db.Entity = info.UID
	db.Scene = info.Owner
	db.Creator = info.Creator
	db.Operator = info.Operator

	db.Access = 0
	db.Score = 0
	var er error
	db.File, db.MD5, db.Size, er = info.encode()
	if er != nil {
		return er
	}
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

func (mine *cacheContext) GetPublicEntity(entity string) (*EntityInfo, error) {
	if len(entity) < 1 {
		return nil, errors.New("the entity uid is empty")
	}
	db, err := nosql.GetArchivedByEntity(entity)
	if err == nil && db != nil {
		info := new(ArchivedInfo)
		info.initInfo(db)
		return info.Decode()
	}
	for _, table := range mine.entityTables {
		db2, err1 := nosql.GetEntity(table, entity)
		if err1 == nil && db2 != nil && db2.Scene != DefaultOwner {
			info := new(EntityInfo)
			info.initInfo(db2)
			return info, nil
		}
	}

	return nil, errors.New("not found the public entity of entity = " + entity)
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
			tmp, er := info.Decode()
			if er == nil {
				arr = append(arr, tmp)
			} else {
				logger.Warn("decode archive entity failed that uid = " + key + " and error = " + er.Error())
			}
		}
	}
	return arr, nil
}

func (mine *cacheContext) GetArchivedList(name string) []*EntityInfo {
	var array []*nosql.Archived
	var err error
	if len(name) > 0 {
		array, err = nosql.GetArchivedItems(name)
	} else {
		array, err = nosql.GetAllArchived()
	}
	if err != nil {
		return make([]*EntityInfo, 0, 1)
	}
	list := make([]*EntityInfo, 0, len(array))
	for _, db := range array {
		info := new(ArchivedInfo)
		info.initInfo(db)
		entity, er := info.Decode()
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
	for _, db := range array {
		info := new(ArchivedInfo)
		info.initInfo(db)
		entity, er := info.Decode()
		if er == nil {
			entity.Status = EntityStatusUsable
			list = append(list, entity)
		}
	}
	return list
}

func (mine *cacheContext) getArchivedEntitiesByAttribute(attr string) []*ArchivedInfo {
	array, err := nosql.GetAllArchived()
	if err != nil {
		return make([]*ArchivedInfo, 0, 1)
	}
	list := make([]*ArchivedInfo, 0, len(array))
	for _, db := range array {
		info := new(ArchivedInfo)
		info.initInfo(db)
		entity, er := info.Decode()
		if er == nil && entity.HadProperty(attr) {
			list = append(list, info)
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
	mine.Created = db.Created
	mine.Updated = db.Updated
	mine.Creator = db.Creator
	mine.Operator = db.Operator
	mine.Concept = db.Concept
	mine.Name = db.Name
	mine.Entity = db.Entity
	mine.File = db.File
	mine.MD5 = db.MD5
	mine.Scene = db.Scene
	mine.Access = db.Access
	mine.Size = db.Size
	//if strings.Contains(mine.File,"http://rdp-down.suii.cn/") {
	//	f := strings.Replace(mine.File, "http://rdp-down.suii.cn/", "", 1)
	//	_ = mine.setFile(f)
	//}
	return true
}

func (mine *ArchivedInfo) UpdateFile(info *EntityInfo, operator string) error {
	if info == nil {
		return errors.New("the entity info is nil")
	}
	data, md5, size, er := info.encode()
	if er != nil {
		return er
	}
	err := nosql.UpdateArchivedFile(mine.UID, operator, string(data), md5, size)
	if err == nil {
		mine.File = data
		mine.MD5 = md5
		mine.Size = size
		mine.Updated = time.Now().Unix()
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

func (mine *ArchivedInfo) replaceAttribute(old, news string) error {
	entity, err := mine.Decode()
	if err != nil {
		return err
	}
	for _, prop := range entity.Properties {
		if prop.Key == old {
			prop.Key = news
		}
	}
	return mine.UpdateFile(entity, mine.Operator)
}

func (mine *ArchivedInfo) Decode() (*EntityInfo, error) {
	entity := new(EntityInfo)
	var data []byte
	var err error
	if mine.Size > 0 {
		data, err = base64.StdEncoding.DecodeString(mine.File)
		if err != nil {
			return nil, err
		}
	} else {
		data = []byte(mine.File)
	}
	er := json.Unmarshal(data, entity)
	if er != nil {
		return nil, er
	}
	now := cacheCtx.GetEntity(entity.UID)
	entity.Status = now.Status
	entity.Published = true
	entity.Access = mine.Access
	return entity, nil
}
