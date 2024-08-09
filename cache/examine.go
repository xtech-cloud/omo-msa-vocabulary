package cache

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy/nosql"
	"time"
)

const (
	ExamineTypeBase      ExamineType = 0 //基本信息
	ExamineTypeAttribute ExamineType = 1 //属性
	ExamineTypeEvent     ExamineType = 2 //事件
)

const (
	ExamineStatusIdle   = 0 //待审核
	ExamineStatusFree   = 1 //通过
	ExamineStatusRefuse = 2 //拒绝
)

const (
	ExamineBaseAvatar  = "avatar"
	ExamineBaseDesc    = "desc"
	ExamineBaseSummary = "summary"
	ExamineBaseName    = "name"
)

type ExamineType uint8

type ExamineInfo struct {
	UID  string
	Data *nosql.Examine
}

func (mine *cacheContext) CreateExamine(creator, target, key, val string, tp uint8) (*ExamineInfo, error) {
	db := new(nosql.Examine)
	db.UID = primitive.NewObjectID()
	db.ID = nosql.GetExamineNextID()
	db.Created = time.Now().Unix()
	db.Creator = creator
	db.Key = key
	db.Target = target
	db.Kind = tp
	db.Value = val
	db.Status = ExamineStatusIdle
	err := nosql.CreateExamine(db)
	if err == nil {
		info := new(ExamineInfo)
		info.initInfo(db)
		return info, nil
	}
	return nil, err
}

func (mine *cacheContext) GetExamineCountByStatus(target string, st uint8) uint32 {
	if target == "" {
		return 0
	}
	return nosql.GetExamineCountByStatus(target, st)
}

func (mine *cacheContext) GetExamineCountByType(target string, tp uint8) uint32 {
	if target == "" {
		return 0
	}
	return nosql.GetExamineCountByType(target, tp, ExamineStatusIdle)
}

func (mine *cacheContext) GetExaminesByTarget(target string) []*ExamineInfo {
	if target == "" {
		return nil
	}
	dbs, err := nosql.GetExamineByTarget(target)
	list := make([]*ExamineInfo, 0, len(dbs))
	if err != nil {
		return list
	}
	for _, db := range dbs {
		tmp := new(ExamineInfo)
		tmp.initInfo(db)
		list = append(list, tmp)
	}

	return list
}

func (mine *cacheContext) GetExaminesByStatus(target string, st uint8) []*ExamineInfo {
	if target == "" {
		return nil
	}

	dbs, err := nosql.GetExamineByStatus(target, st)
	list := make([]*ExamineInfo, 0, len(dbs))
	if err != nil {
		return list
	}
	for _, db := range dbs {
		tmp := new(ExamineInfo)
		tmp.initInfo(db)
		list = append(list, tmp)
	}

	return list
}

func (mine *cacheContext) GetExaminesByType(target string, tp ExamineType) []*ExamineInfo {
	if target == "" {
		return nil
	}
	dbs, err := nosql.GetExamineByType(target, uint8(tp), ExamineStatusIdle)
	list := make([]*ExamineInfo, 0, len(dbs))
	if err != nil {
		return list
	}
	for _, db := range dbs {
		tmp := new(ExamineInfo)
		tmp.initInfo(db)
		list = append(list, tmp)
	}
	return list
}

func (mine *cacheContext) GetExamine(uid string) *ExamineInfo {
	if uid == "" {
		return nil
	}
	db, err := nosql.GetExamine(uid)
	if err != nil {
		return nil
	}
	tmp := new(ExamineInfo)
	tmp.initInfo(db)
	return tmp
}

func (mine *ExamineInfo) initInfo(db *nosql.Examine) {
	mine.UID = db.UID.Hex()
	mine.Data = db
}

func (mine *ExamineInfo) UpdateStatus(st uint8, operator string) error {
	if st == ExamineStatusFree {
		_ = updateTargetValue(mine.Data.Target, mine.Data.Key, mine.Data.Value, operator, ExamineType(mine.Data.Kind))
	}
	err := nosql.UpdateExamineStatus(mine.UID, operator, st)
	if err == nil {
		mine.Data.Status = st
		mine.Data.Operator = operator
		mine.Data.Updated = time.Now().Unix()
	}
	return err
}

func updateTargetValue(target, key, val, operator string, tp ExamineType) error {
	entity := cacheCtx.GetEntity(target)
	if entity == nil {
		return nil
	}
	var err error
	var had = false
	if tp == ExamineTypeBase {
		if key == ExamineBaseName {
			err = entity.UpdateName(val, operator)
			had = true
		} else if key == ExamineBaseAvatar {
			err = entity.UpdateCover(val, operator)
			had = true
		} else if key == ExamineBaseSummary {
			err = entity.UpdateRemark(entity.Description, val, operator)
			had = true
		} else if key == ExamineBaseDesc {
			err = entity.UpdateRemark(val, entity.Summary, operator)
			had = true
		}
		if err == nil && had {
			err = entity.UpdateStatus(EntityStatusUsable, operator, "")
		}
	} else if tp == ExamineTypeAttribute {
		att := cacheCtx.GetAttributeByKey(key)
		if att != nil {
			err = entity.UpdateProperty(att.UID, val, operator)
		}
	} else if tp == ExamineTypeEvent {
		event := cacheCtx.GetEvent(key)
		if event != nil {

		}
	}
	return err
}
