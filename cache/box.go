package cache

import (
	"errors"
	"github.com/micro/go-micro/v2/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math"
	"omo.msa.vocabulary/proxy"
	"omo.msa.vocabulary/proxy/nosql"
	"omo.msa.vocabulary/tool"
	"time"
)

const DefaultOwner = "system"

type BoxInfo struct {
	Type uint8
	BaseInfo
	Cover    string
	Remark   string
	Concept  string // 针对的实体类型
	Workflow string
	Owner    string

	Users     []string             //采集人
	Reviewers []string             //审核人
	Contents  []*proxy.ContentInfo //内容
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

func (mine *cacheContext) GetAllBoxes() []*BoxInfo {
	dbs, _ := nosql.GetBoxes()
	list := make([]*BoxInfo, 0, len(dbs))
	for _, db := range dbs {
		box := new(BoxInfo)
		box.initInfo(db)
		list = append(list, box)
	}
	return list
}

func (mine *cacheContext) GetBoxesByEntities(arr []string) []*BoxInfo {
	all := mine.GetAllBoxes()
	list := make([]*BoxInfo, 0, len(all))
	for _, item := range all {
		if item.hadEntities(arr) {
			list = append(list, item)
		}
	}
	return list
}

func (mine *cacheContext) GetBoxesByName(val string) []*BoxInfo {
	dbs, _ := nosql.GetBoxesByRegex("name", val)
	list := make([]*BoxInfo, 0, len(dbs))
	for _, db := range dbs {
		box := new(BoxInfo)
		box.initInfo(db)
		list = append(list, box)
	}
	return list
}

func (mine *cacheContext) GetBoxPages(page, number uint32) (uint32, uint32, []*BoxInfo) {
	if page < 1 {
		page = 1
	}
	if number < 1 {
		number = 10
	}
	start := (page - 1) * number
	array, err := nosql.GetBoxesByPage(int64(start), int64(number))
	total := nosql.GetBoxCount()
	pages := math.Ceil(float64(total) / float64(number))
	if err == nil {
		list := make([]*BoxInfo, 0, len(array))
		for _, item := range array {
			info := new(BoxInfo)
			info.initInfo(item)
			list = append(list, info)
		}
		return uint32(total), uint32(pages), list
	}
	return 0, 0, make([]*BoxInfo, 0, 1)
}

func (mine *cacheContext) GetUsableBoxPages(page, number uint32) (uint32, uint32, []*BoxInfo) {
	if page < 1 {
		page = 1
	}
	if number < 1 {
		number = 10
	}
	array, err := nosql.GetBoxes()
	if err == nil {
		list := make([]*BoxInfo, 0, len(array))
		for _, item := range array {
			info := new(BoxInfo)
			info.initInfo(item)
			if info.HadPublished() {
				list = append(list, info)
			}
		}
		total, pages, arr := CheckPage(int32(page), int32(number), list)
		return uint32(total), uint32(pages), arr
	}
	return 0, 0, make([]*BoxInfo, 0, 1)
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

	if len(box.Contents) < 1 {
		return nil, errors.New("the box contents is empty")
	}
	list := make([]*EntityInfo, 0, len(box.Contents))
	for _, item := range box.Contents {
		if st == EntityStatusUsable {
			info := mine.GetArchivedByEntity(item.Keyword)
			if info != nil {
				tmp, er := info.Decode()
				if er == nil {
					list = append(list, tmp)
				} else {
					logger.Warn("decode archive entity failed that uid = " + item.Keyword + " and error = " + er.Error())
				}
			}
		} else {
			info := mine.GetEntity(item.Keyword)
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

func (mine *cacheContext) GetEntitiesByAdditional(add string) ([]*EntityInfo, error) {
	if len(add) < 1 {
		return nil, errors.New("the entity add is empty")
	}
	array, err := nosql.GetEntitiesByAdditional(DefaultEntityTable, add)
	if err != nil {
		return nil, err
	}
	list := make([]*EntityInfo, 0, len(array))
	for _, entity := range array {
		info := new(EntityInfo)
		info.initInfo(entity)
		list = append(list, info)
	}
	array1, err1 := nosql.GetEntitiesByAdditional(UserEntityTable, add)
	if err1 == nil {
		for _, entity := range array1 {
			info := new(EntityInfo)
			info.initInfo(entity)
			list = append(list, info)
		}
	}

	return list, nil
}

func (mine *cacheContext) GetEntitiesByConceptNum(concept string, num int32) ([]*EntityInfo, error) {
	array, err := nosql.GetEntitiesByRankConcept(DefaultEntityTable, concept, int64(num))
	if err != nil {
		return nil, err
	}
	list := make([]*EntityInfo, 0, len(array))
	for _, entity := range array {
		info := new(EntityInfo)
		info.initInfo(entity)
		list = append(list, info)
	}

	return list, nil
}

func (mine *cacheContext) GetEntitiesByType(tp uint32, num int32) []*EntityInfo {
	tps := mine.GetConceptsByType(tp)
	all := make([]*EntityInfo, 0, 100)
	for _, tp1 := range tps {
		arr, _ := mine.GetEntitiesByConceptNum(tp1.UID, num)
		all = append(all, arr...)
	}
	return all
}

func (mine *cacheContext) UpdateBoxContentStatus(entity string, st EntityStatus, publish bool) {
	boxes := mine.GetBoxesByKeyword(entity)
	var pub uint32 = 0
	if publish {
		pub = 1
	}
	for _, box := range boxes {
		_ = box.updateContentStatus(entity, st, pub)
	}
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
	//db.Keywords = make([]string, 0, 5)
	db.Users = make([]string, 0, 5)
	db.Contents = info.Contents
	if db.Contents == nil {
		db.Contents = make([]*proxy.ContentInfo, 0, 1)
	}
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
	mine.Users = db.Users
	if len(mine.Owner) < 1 {
		_ = mine.updateOwner(DefaultOwner)
	}
	mine.Contents = db.Contents
	if len(db.Contents) < 1 && len(db.Keywords) > 0 {
		contents := make([]*proxy.ContentInfo, 0, len(db.Keywords))
		for _, key := range db.Keywords {
			if hadChinese(key) {
				contents = append(contents, &proxy.ContentInfo{Keyword: "", Name: key, Count: 0, Status: uint8(EntityStatusDraft)})
			} else {
				entity := cacheCtx.GetEntity(key)
				if entity == nil {
					contents = append(contents, &proxy.ContentInfo{Keyword: key, Name: "", Count: 0, Status: 0})
				} else {
					var pub uint32 = 0
					if entity.Published {
						pub = 1
					}
					contents = append(contents, &proxy.ContentInfo{Keyword: key, Name: entity.Name, Count: pub, Status: uint8(entity.Status)})
				}
			}
		}
		_ = mine.updateContents(contents, mine.Operator)
	}
}

func (mine *BoxInfo) hadEntities(arr []string) bool {
	for _, content := range mine.Contents {
		if tool.HasItem(arr, content.Keyword) {
			return true
		}
	}
	return false
}

func (mine *BoxInfo) updateContents(list []*proxy.ContentInfo, operator string) error {
	if list == nil {
		return errors.New("the list is nil when update")
	}

	err := nosql.UpdateBoxContents(mine.UID, operator, list)
	if err == nil {
		mine.Contents = list
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *BoxInfo) updateContentStatus(entity string, st EntityStatus, publish uint32) error {
	list := make([]*proxy.ContentInfo, 0, len(mine.Contents))
	list = append(list, mine.Contents...)
	for _, info := range list {
		if info.Keyword == entity {
			info.Status = uint8(st)
			info.Count = publish
			break
		}
	}
	return mine.updateContents(list, mine.Operator)
}

func (mine *BoxInfo) UpdateContents(arr []string, operator string) error {
	if arr == nil {
		return errors.New("the arr is nil when update")
	}
	list := make([]*proxy.ContentInfo, 0, len(arr))
	for _, item := range arr {
		list = append(list, &proxy.ContentInfo{
			Keyword: item, Name: "", Count: 0, Status: 0,
		})
	}
	err := nosql.UpdateBoxContents(mine.UID, operator, list)
	if err == nil {
		mine.Contents = list
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

func (mine *BoxInfo) UpdateConcept(con, operator string) error {
	if mine.Concept == con {
		return nil
	}
	er := nosql.UpdateBoxConcept(mine.UID, operator, con)
	if er != nil {
		return er
	}
	for _, item := range mine.Contents {
		if !hadChinese(item.Keyword) {
			_ = mine.updateEntityConcept(item.Keyword, con, operator)
		}
	}
	return nil
}

func (mine *BoxInfo) HadPublished() bool {
	for _, content := range mine.Contents {
		if content.Status == uint8(EntityStatusUsable) {
			return true
		}
	}
	return false
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
	list := make([]*proxy.ContentInfo, 0, len(keys)+len(mine.Contents))
	list = append(list, mine.Contents...)
	for i := 0; i < len(keys); i += 1 {
		if !mine.HadContent(keys[i]) {
			key := ""
			name := ""
			if hadChinese(keys[i]) {
				name = keys[i]
			} else {
				key = keys[i]
			}
			list = append(list, &proxy.ContentInfo{
				Keyword: key,
				Name:    name,
				Count:   0,
				Status:  0,
			})
		}
	}
	err := nosql.UpdateBoxContents(mine.UID, operator, list)
	if err == nil {
		mine.Contents = list
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *BoxInfo) HadContent(key string) bool {
	if mine.Contents == nil {
		return false
	}
	for _, content := range mine.Contents {
		if content.Keyword == key || content.Name == key {
			return true
		}
	}
	return false
}

func (mine *BoxInfo) FillContent(name, entity, operator string) error {
	list := make([]*proxy.ContentInfo, 0, len(mine.Contents))
	list = append(list, mine.Contents...)
	for _, content := range mine.Contents {
		if content.Name == name {
			content.Keyword = entity
			content.Status = uint8(EntityStatusDraft)
			break
		}
	}
	return mine.updateContents(list, operator)
}

func (mine *BoxInfo) RemoveKeywords(keys []string, operator string) error {
	list := make([]*proxy.ContentInfo, 0, len(mine.Contents))
	for _, item := range mine.Contents {
		if !tool.HasItem(keys, item.Keyword) && !tool.HasItem(keys, item.Name) {
			list = append(list, item)
		}
	}
	err := nosql.UpdateBoxContents(mine.UID, operator, list)
	if err == nil {
		mine.Contents = list
		mine.Updated = time.Now().Unix()
	}
	return err
}

func (mine *BoxInfo) AppendKeyword(key string) error {
	if mine.Contents == nil {
		return errors.New("must call construct fist")
	}
	if !mine.HadContent(key) {
		return errors.New("not found the property when remove")
	}
	err := nosql.AppendBoxKeyword(mine.UID, key)
	if err == nil {
		for i := 0; i < len(mine.Contents); i += 1 {
			if mine.Contents[i].Keyword == key {
				mine.Contents = append(mine.Contents[:i], mine.Contents[i+1:]...)
				break
			}
		}
	}
	return err
}

func (mine *BoxInfo) RemoveKeyword(key string) error {
	if mine.Contents == nil {
		return errors.New("must call construct fist")
	}
	if !mine.HadContent(key) {
		return errors.New("not found the property when remove")
	}
	err := nosql.SubtractBoxKeyword(mine.UID, key)
	if err == nil {
		for i := 0; i < len(mine.Contents); i += 1 {
			if mine.Contents[i].Keyword == key {
				mine.Contents = append(mine.Contents[:i], mine.Contents[i+1:]...)
				break
			}
		}
	}
	return err
}

func (mine *BoxInfo) UpdateBase(name, remark, concept, operator string) error {
	if mine.Name != name || mine.Remark != remark {
		err := nosql.UpdateBoxBase(mine.UID, name, remark, operator)
		if err != nil {
			return err
		}
		mine.Name = name
		mine.Remark = remark
		mine.Operator = operator
		mine.Updated = time.Now().Unix()
	}
	err := mine.UpdateConcept(concept, operator)
	if err != nil {
		return err
	}
	return nil
}

func (mine *BoxInfo) updateEntityConcept(uid, concept, operator string) error {
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
