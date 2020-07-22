package grpc

import (
	"context"
	"errors"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
	"omo.msa.vocabulary/proxy"
)

type EntityService struct{}

func switchEntity(info *cache.EntityInfo) *pb.EntityInfo {
	tmp := new(pb.EntityInfo)
	tmp.Uid = info.UID
	tmp.Concept = info.Concept
	tmp.Cover = info.Cover
	tmp.Name = info.Name
	tmp.Created = info.CreateTime.Unix()
	tmp.Updated = info.UpdateTime.Unix()
	tmp.Description = info.Description
	tmp.Operator = info.Operator
	tmp.Creator = info.Creator
	tmp.Owner = info.Owner
	tmp.Tags = info.Tags
	tmp.Synonyms = info.Synonyms
	tmp.Add = info.Add
	length := len(info.Properties())
	tmp.Properties = make([]*pb.PropertyInfo, 0, length)
	for _, value := range info.Properties() {
		tmp.Properties = append(tmp.Properties, switchEntityProperty(value))
	}

	return tmp
}

func switchEntityProperty(info *proxy.PropertyInfo) *pb.PropertyInfo {
	tmp := new(pb.PropertyInfo)
	tmp.Key = info.Key
	tmp.Words = make([]*pb.WordInfo, 0, len(info.Words))
	for _, value := range info.Words {
		tmp.Words = append(tmp.Words, &pb.WordInfo{Uid:value.UID, Name:value.Name})
	}
	return tmp
}

func (mine *EntityService)AddOne(ctx context.Context, in *pb.ReqEntityAdd, out *pb.ReplyEntityOne) error {
	inLog("entity.add", in)
	tmp := cache.GetEntityByName(in.Name)
	if tmp != nil {
		if tmp.Concept == in.Name {
			out.ErrorCode = pb.ResultStatus_Repeated
			return errors.New("the entity is existed")
		}
	}
	info := new(cache.EntityInfo)
	info.Name = in.Name
	info.Description = in.Description
	info.Add = in.Add
	info.Creator = in.Creator
	info.Owner = in.Owner
	info.Cover = in.Cover
	info.Concept = in.Concept
	info.Synonyms = in.Synonyms
	info.Tags = in.Tags
	_,err := cache.CreateEntity(info)
	if err == nil {
		out.Info = switchEntity(info)
	}else{
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *EntityService)GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEntityOne) error {
	inLog("entity.one", in)
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the entity uid is empty")
	}
	info := cache.GetEntity(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the entity by uid")
	}
	out.Info = switchEntity(info)
	return nil
}

func (mine *EntityService)RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	inLog("entity.remove", in)
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the entity uid is empty")
	}
	err := cache.RemoveEntity(in.Uid, in.Operator)
	out.Uid = in.Uid
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *EntityService)GetAllByOwner(ctx context.Context, in *pb.ReqEntityBy, out *pb.ReplyEntityAll) error {
	out.Flag = in.Owner
	out.List = make([]*pb.EntityInfo, 0, 10)
	for _, value := range cache.AllEntities() {
		if value.Owner == in.Owner && value.Status == cache.EntityStatus(in.Status) {
			out.List = append(out.List, switchEntity(value))
		}
	}
	return nil
}

func (mine *EntityService)UpdateTags(ctx context.Context, in *pb.ReqEntityUpdate, out *pb.ReplyEntityUpdate) error {
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the entity uid is empty")
	}
	info := cache.GetEntity(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the entity by uid")
	}
	err := info.UpdateTags(in.List, in.Operator)
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *EntityService)UpdateBase(ctx context.Context, in *pb.ReqEntityBase, out *pb.ReplyInfo) error {
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the entity uid is empty")
	}
	info := cache.GetEntity(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the entity by uid")
	}
	var err error
	if len(in.Cover) > 0 {
		err = info.UpdateCover(in.Cover, in.Operator)
	}else{
		err = info.UpdateBase(in.Name, in.Desc, in.Add, in.Concept, in.Operator)
	}

	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *EntityService)UpdateStatus(ctx context.Context, in *pb.ReqEntityStatus, out *pb.ReplyEntityStatus) error {
	if len(in.Uid) < 1 {
		//out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the entity uid is empty")
	}
	info := cache.GetEntity(in.Uid)
	if info == nil {
		//out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the entity by uid")
	}
	err := info.UpdateStatus(cache.EntityStatus(in.Status), in.Operator)
	out.Uid = in.Uid
	out.Status = in.Status
	return err
}

func (mine *EntityService)UpdateSynonyms(ctx context.Context, in *pb.ReqEntityUpdate, out *pb.ReplyEntityUpdate) error {
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the entity uid is empty")
	}
	info := cache.GetEntity(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the entity by uid")
	}
	err := info.UpdateSynonyms(in.List, in.Operator)
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *EntityService)AppendProperty(ctx context.Context, in *pb.ReqEntityProperty, out *pb.ReplyEntityProperties) error {
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the entity uid is empty")
	}
	info := cache.GetEntity(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the entity by uid")
	}
	if info.HadProperty(in.Property.Key) {
		out.ErrorCode = pb.ResultStatus_Repeated
		return errors.New("the key of entity is sxisted")
	}
	words := make([]proxy.WordInfo, 0, len(in.Property.Words))
	for _, value := range in.Property.Words {
		words = append(words, proxy.WordInfo{UID:value.Uid, Name:value.Name})
	}
	err := info.AddProperty(in.Property.Key, words)
	if err == nil {
		out.Properties = make([]*pb.PropertyInfo, 0, len(info.Properties()))
		for _, value := range info.Properties() {
			tmp := switchEntityProperty(value)
			out.Properties = append(out.Properties, tmp)
		}
	}else{
		out.ErrorCode = pb.ResultStatus_DBException
	}

	return err
}

func (mine *EntityService)SubtractProperty(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEntityProperties) error {
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the entity uid is empty")
	}
	info := cache.GetEntity(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the entity by uid")
	}
	err := info.RemoveProperty(in.Key)
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}else{
		out.Properties = make([]*pb.PropertyInfo, 0, len(info.Properties()))
		for _, value := range info.Properties() {
			tmp := switchEntityProperty(value)
			out.Properties = append(out.Properties, tmp)
		}
	}
	return err
}
