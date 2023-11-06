package cache

import (
	"errors"
	"github.com/micro/go-micro/v2/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy/nosql"
	"omo.msa.vocabulary/tool"
	"time"
)

const DefaultOwner = "system"

type BoxInfo struct {
	Type uint8
	BaseInfo
	Cover     string
	Remark    string
	Concept   string // 针对的实体类型
	Workflow  string
	Owner     string
	Keywords  []string
	Users     []string //采集人
	Reviewers []string //审核人
}

//region Global Fun

func (mine *cacheContext) GetBoxByName(name string) *BoxInfo {
	db, err := nosql.GetBoxByName(name)
	if err == nil {
		info := new(BoxInfo)
		info.initInfo(db)
		return info
	}
	return nil
}

func (mine *cacheContext) GetBox(uid string) *BoxInfo {
	db, err := nosql.GetBox(uid)
	if err == nil {
		info := new(BoxInfo)
		info.initInfo(db)
		return info
	}
	return nil
}

func (mine *cacheContext) GetBoxes(owner string, kind uint8) []*BoxInfo {
	if len(owner) < 1 {
		owner = DefaultOwner
	}
	dbs, _ := nosql.GetBoxesByType(owner, kind)
	list := make([]*BoxInfo, 0, len(dbs))
	for _, db := range dbs {
		box := new(BoxInfo)
		box.initInfo(db)
		list = append(list, box)
	}

	return list
}

func (mine *cacheContext) GetBoxesByUser(user string) []*BoxInfo {
	dbs, _ := nosql.GetBoxesByUser(user)
	list := make([]*BoxInfo, 0, len(dbs))
	for _, db := range dbs {
		box := new(BoxInfo)
		box.initInfo(db)
		list = append(list, box)
	}
	return list
}

func (mine *cacheContext) GetBoxesByReviewer(user string) []*BoxInfo {
	dbs, _ := nosql.GetBoxesByReviewer(user)
	list := make([]*BoxInfo, 0, len(dbs))
	for _, db := range dbs {
		box := new(BoxInfo)
		box.initInfo(db)
		list = append(list, box)
	}
	return list
}

func (mine *cacheContext) GetBoxesByConcept(concept string) []*BoxInfo {
	dbs, _ := nosql.GetBoxesByConcept(concept)
	list := make([]*BoxInfo, 0, len(dbs))
	for _, db := range dbs {
		box := new(BoxInfo)
		box.initInfo(db)
		list = append(list, box)
	}
	return list
}

func (mine *cacheContext) GetBoxesByKeyword(key string) []*BoxInfo {
	dbs, _ := nosql.GetBoxesByKeyword(key)
	list := make([]*BoxInfo, 0, len(dbs))
	for _, db := range dbs {
		box := new(BoxInfo)
		box.initInfo(db)
		list = append(list, box)
	}
	return list
}

func (mine *cacheContext) GetBoxesByOwner(owner string) []*BoxInfo {
	dbs, _ := nosql.GetBoxesByOwner(owner)
	list := make([]*BoxInfo, 0, len(dbs))
	for _, db := range dbs {
		box := new(BoxInfo)
		box.initInfo(db)
		list = append(list, box)
	}
	return list
}

func (mine *cacheContext) GetEntitiesByBox(uid string, st EntityStatus) ([]*EntityInfo, error) {
	box := mine.GetBox(uid)
	if box == nil {
		return nil, errors.New("not found the box that uid = " + uid)
	}

	if box.Keywords == nil || len(box.Keywords) < 1 {
		return nil, errors.New("the box keywords is empty")
	}
	list := make([]*EntityInfo, 0, len(box.Keywords))
	for _, item := range box.Keywords {
		if st == EntityStatusUsable {
			info := mine.GetArchivedByEntity(item)
			if info != nil {
				tmp, er := info.Decode()
				if er == nil {
					list = append(list, tmp)
				} else {
					logger.Warn("decode archive entity failed that uid = " + item + " and error = " + er.Error())
				}
			}
		} else {
			info := mine.GetEntity(item)
			if info != nil {
				list = append(list, info)
			}
		}
	}
	return list, nil
}

func (mine *cacheContext) GetEntitiesByName(name string) ([]*EntityInfo, error) {
	if len(name) < 1 {
		return nil, errors.New("the name is empty")
	}
	array, err := nosql.GetEntitiesByName(DefaultEntityTable, name)
	if err != nil {
		return nil, err
	}
	list := make([]*EntityInfo, 0, len(array))
	for _, entity := range array {
		info := new(EntityInfo)
		info.initInfo(entity)
		list = append(list, info)
	}
	array1, err1 := nosql.GetEntitiesByName(UserEntityTable, name)
	if err1 == nil {
		for _, entity := range array1 {
			info := new(EntityInfo)
			info.initInfo(entity)
			list = append(list, info)
		}
	}

	return list, nil
}

func (mine *cacheContext) CreateBox(info *BoxInfo) error {
	db := new(nosql.Box)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetBoxNextID()
	db.Created = time.Now().Unix()
	db.CreatedTime = time.Now()
	db.Creator = info.Creator
	db.Name = info.Name
	db.Concept = info.Concept
	db.Cover = info.Cover
	db.Remark = info.Remark
	db.Type = info.Type
	db.Owner = info.Owner
	db.Workflow = info.Workflow
	db.Keywords = make([]string, 0, 5)
	db.Users = make([]string, 0, 5)
	err := nosql.CreateBox(db)
	if err == nil {
		info.initInfo(db)
	}
	return err
}

func (mine *cacheContext) RemoveBox(uid, operator string) error {
	err := nosql.RemoveBox(uid, operator)
	if err == nil {
		//for i := 0; i < len(mine.boxes); i += 1 {
		//	if mine.boxes[i].UID == uid {
		//		mine.boxes = append(mine.boxes[:i], mine.boxes[i+1:]...)
		//		break
		//	}
		//}
	}
	return err
}

func (mine *cacheContext) HadBoxByName(name string) bool {
	had, _ := nosql.HadBoxByName(name)
	return had
}

func (mine *cacheContext) checkEntityFromBoxes(uid, name string) {
	boxes := mine.GetBoxesByKeyword(uid)
	for _, box := range boxes {
		_ = box.RemoveKeyword(uid)
		_ = box.AppendKeyword(name)
	}
}

//endregion

//region Base Fun
func (mine *BoxInfo) initInfo(db *nosql.Box) {
	if db == nil {
		return
	}
	mine.ID = db.ID
	mine.UID = db.UID.Hex()
	mine.Name = db.Name
	mine.Type = db.Type
	mine.Remark = db.Remark
	mine.Concept = db.Concept
	mine.Cover = db.Cover
	mine.Updated = db.Updated
	mine.Created = db.Created
	mine.Operator = db.Operator
	mine.Creator = db.Creator
	mine.Owner = db.Owner
	mine.Workflow = db.Workflow
	mine.Keywords = db.Keywords
	mine.Users = db.Users
	if len(mine.Owner) < 1 {
		_ = mine.updateOwner(DefaultOwner)
	}
}

func (mine *BoxInfo) UpdateKeywords(list []string, operator string) error {
	if list == nil {
		return errors.New("the list is nil when update")
	}

	err := nosql.UpdateBoxKeywords(mine.UID, operator, list)
	if err == nil {
		mine.Keywords = list
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *BoxInfo) updateOwner(owner string) error {
	err := nosql.UpdateBoxOwner(mine.UID, owner)
	if err == nil {
		mine.Owner = owner
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *BoxInfo) UpdateUsers(list []string, operator string, reviewer bool) error {
	if list == nil {
		return errors.New("the list is nil when update users")
	}
	var err error
	if reviewer {
		err = nosql.UpdateBoxReviewers(mine.UID, operator, list)
	} else {
		err = nosql.UpdateBoxUsers(mine.UID, operator, list)
	}

	if err == nil {
		if reviewer {
			mine.Reviewers = list
		} else {
			mine.Users = list
		}
		mine.Operator = operator
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *BoxInfo) HadUser(key string) bool {
	if mine.Users == nil {
		return false
	}
	for i := 0; i < len(mine.Users); i += 1 {
		if mine.Users[i] == key {
			return true
		}
	}
	return false
}

func (mine *BoxInfo) HadReviewer(key string) bool {
	if mine.Reviewers == nil {
		return false
	}
	for i := 0; i < len(mine.Reviewers); i += 1 {
		if mine.Reviewers[i] == key {
			return true
		}
	}
	return false
}

func (mine *BoxInfo) AppendUsers(keys []string, operator string, review bool) error {
	var list []string
	var arr []string
	if review {
		arr = mine.Reviewers
	} else {
		arr = mine.Users
	}
	list = make([]string, 0, len(keys)+len(arr))
	list = append(list, arr...)
	for i := 0; i < len(keys); i += 1 {
		if !mine.HadUser(keys[i]) {
			list = append(list, keys[i])
		}
	}
	return mine.UpdateUsers(list, operator, review)
}

func (mine *BoxInfo) RemoveUsers(keys []string, operator string, review bool) error {
	var list []string
	var arr []string
	if review {
		arr = mine.Reviewers
	} else {
		arr = mine.Users
	}
	list = make([]string, 0, len(arr))
	for _, keyword := range arr {
		if !tool.HasItem(keys, keyword) {
			list = append(list, keyword)
		}
	}
	return mine.UpdateUsers(list, operator, review)
}

func (mine *BoxInfo) AppendKeywords(keys []string, operator string) error {
	list := make([]string, 0, len(keys)+len(mine.Keywords))
	list = append(list, mine.Keywords...)
	for i := 0; i < len(keys); i += 1 {
		if !mine.HadKeyword(keys[i]) {
			list = append(list, keys[i])
		}
	}
	err := nosql.UpdateBoxKeywords(mine.UID, operator, list)
	if err == nil {
		mine.Keywords = list
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *BoxInfo) HadKeyword(key string) bool {
	if mine.Keywords == nil {
		return false
	}
	for i := 0; i < len(mine.Keywords); i += 1 {
		if mine.Keywords[i] == key {
			return true
		}
	}
	return false
}

func (mine *BoxInfo) RemoveKeywords(keys []string, operator string) error {
	list := make([]string, 0, len(mine.Keywords))
	for _, keyword := range mine.Keywords {
		if !tool.HasItem(keys, keyword) {
			list = append(list, keyword)
		}
	}
	err := nosql.UpdateBoxKeywords(mine.UID, operator, list)
	if err == nil {
		mine.Keywords = list
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *BoxInfo) AppendKeyword(key string) error {
	if mine.Keywords == nil {
		return errors.New("must call construct fist")
	}
	if !mine.HadKeyword(key) {
		return errors.New("not found the property when remove")
	}
	err := nosql.AppendBoxKeyword(mine.UID, key)
	if err == nil {
		for i := 0; i < len(mine.Keywords); i += 1 {
			if mine.Keywords[i] == key {
				mine.Keywords = append(mine.Keywords[:i], mine.Keywords[i+1:]...)
				break
			}
		}
	}
	return err
}

func (mine *BoxInfo) RemoveKeyword(key string) error {
	if mine.Keywords == nil {
		return errors.New("must call construct fist")
	}
	if !mine.HadKeyword(key) {
		return errors.New("not found the property when remove")
	}
	err := nosql.SubtractBoxKeyword(mine.UID, key)
	if err == nil {
		for i := 0; i < len(mine.Keywords); i += 1 {
			if mine.Keywords[i] == key {
				mine.Keywords = append(mine.Keywords[:i], mine.Keywords[i+1:]...)
				break
			}
		}
	}
	return err
}

func (mine *BoxInfo) UpdateBase(name, remark, concept, operator string) error {
	if mine.Name != name || mine.Remark != remark || mine.Concept != concept {
		err := nosql.UpdateBoxBase(mine.UID, name, remark, concept, operator)
		if err == nil {
			if mine.Concept != concept {
				for _, keyword := range mine.Keywords {
					if !hadChinese(keyword) {
						mine.updateEntity(keyword, concept, operator)
					}
				}
			}
			mine.Name = name
			mine.Remark = remark
			mine.Operator = operator
			mine.Concept = concept
			mine.Updated = time.Now().Unix()
		}
		return err
	}
	return nil
}

func (mine *BoxInfo) updateEntity(uid, concept, operator string) error {
	info := cacheCtx.GetEntity(uid)
	if info != nil {
		return info.updateConcept(concept, operator)
	}
	return nil
}

func (mine *BoxInfo) UpdateCover(cover string) error {
	err := nosql.UpdateBoxCover(mine.UID, cover)
	if err == nil {
		mine.Cover = cover
		mine.Updated = time.Now().Unix()
	}
	return err
}

//endregion
