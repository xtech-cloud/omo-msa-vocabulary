package cache

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy/nosql"
	"omo.msa.vocabulary/tool"
	"time"
)

type BoxInfo struct {
	Type uint8
	BaseInfo
	Cover      string
	Remark     string
	Concept    string  // 针对的场景类型
	Keywords []string
}

//region Global Fun

func (mine *cacheContext)GetBoxByName(name string) *BoxInfo {
	for i := 0; i < len(mine.boxes); i += 1 {
		if mine.boxes[i].Name == name {
			return mine.boxes[i]
		}
	}
	return nil
}

func (mine *cacheContext)GetBox(uid string) *BoxInfo {
	for i := 0; i < len(mine.boxes); i += 1 {
		if mine.boxes[i].UID == uid {
			return mine.boxes[i]
		}
	}
	return nil
}

func (mine *cacheContext)GetBoxes(kind uint8) []*BoxInfo {
	list := make([]*BoxInfo, 0, 10)
	for _, box := range mine.boxes {
		if box.Type == kind {
			list = append(list, box)
		}
	}
	return list
}

func (mine *cacheContext) CreateBox(info *BoxInfo) error {
	db := new(nosql.Box)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetBoxNextID()
	db.CreatedTime = time.Now()
	db.Creator = info.Creator
	db.Name = info.Name
	db.Concept = info.Concept
	db.Cover = info.Cover
	db.Remark = info.Remark
	db.Type = info.Type
	db.Keywords = make([]string, 0, 5)
	err := nosql.CreateBox(db)
	if err == nil {
		info.initInfo(db)
		mine.boxes = append(mine.boxes, info)
	}
	return err
}

func (mine *cacheContext)RemoveBox(uid, operator string) error {
	err := nosql.RemoveBox(uid, operator)
	if err == nil {
		for i := 0; i < len(mine.boxes); i += 1 {
			if mine.boxes[i].UID == uid {
				mine.boxes = append(mine.boxes[:i], mine.boxes[i+1:]...)
				break
			}
		}
	}
	return err
}

func (mine *cacheContext)HadBoxByName(name string) bool {
	for i := 0; i < len(mine.boxes); i += 1 {
		if mine.boxes[i].Name == name {
			return true
		}
	}
	return false
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
	mine.UpdateTime = db.UpdatedTime
	mine.CreateTime = db.CreatedTime
	mine.Operator = db.Operator
	mine.Creator = db.Creator
	mine.Keywords = db.Keywords
}

func (mine *BoxInfo)UpdateKeywords(list []string) error {
	if list == nil {
		return errors.New("the list is nil when update")
	}

	err := nosql.UpdateBoxKeywords(mine.UID, list)
	if err == nil {
		mine.Keywords = list
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *BoxInfo)AppendKeywords(keys []string) error {
	list := make([]string, 0, len(keys) +len(mine.Keywords))
	list = append(list, mine.Keywords...)
	for i := 0;i < len(keys);i += 1 {
		if !mine.HadKeyword(keys[i]){
			list = append(list, keys[i])
		}
	}
	err := nosql.UpdateBoxKeywords(mine.UID, list)
	if err == nil {
		mine.Keywords = list
		mine.UpdateTime = time.Now()
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

func (mine *BoxInfo) RemoveKeywords(keys []string) error {
	list := make([]string, 0, len(mine.Keywords))
	for _, keyword := range mine.Keywords {
		if !tool.HasItem(keys, keyword) {
			list = append(list, keyword)
		}
	}
	err := nosql.UpdateBoxKeywords(mine.UID, list)
	if err == nil {
		mine.Keywords = list
		mine.UpdateTime = time.Now()
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

func (mine *BoxInfo) UpdateBase(name, remark,operator, concept string) error {
	err := nosql.UpdateBoxBase(mine.UID, name, remark, concept, operator)
	if err == nil {
		mine.Name = name
		mine.Remark = remark
		mine.Operator = operator
		mine.Concept = concept
		mine.UpdateTime = time.Now()
	}
	return err
}

func (mine *BoxInfo) UpdateCover(cover string) error {
	err := nosql.UpdateBoxCover(mine.UID, cover)
	if err == nil {
		mine.Cover = cover
		mine.UpdateTime = time.Now()
	}
	return err
}
//endregion
