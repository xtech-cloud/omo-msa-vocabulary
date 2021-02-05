package grpc

import (
	"context"
	"fmt"
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
	tmp.Status = uint32(info.Status)
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
	tmp.Uid = info.Key
	tmp.Words = make([]*pb.WordInfo, 0, len(info.Words))
	for _, value := range info.Words {
		tmp.Words = append(tmp.Words, &pb.WordInfo{Uid:value.UID, Name:value.Name})
	}
	return tmp
}

func (mine *EntityService)AddOne(ctx context.Context, in *pb.ReqEntityAdd, out *pb.ReplyEntityInfo) error {
	path := "entity.addOne"
	inLog(path, in)
	if in.Name == "" {
		out.Status = outError(path,"the entity name is empty", pb.ResultStatus_Empty)
		return nil
	}
	if cache.Context().HadEntityByName(in.Name, in.Add){
		out.Status = outError(path,"the entity name is repeated", pb.ResultStatus_Repeated)
		return nil
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
	info.Status = cache.EntityStatusIdle
	err := cache.Context().CreateEntity(info)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Info = switchEntity(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService)GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEntityInfo) error {
	path := "entity.getOne"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	out.Info = switchEntity(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService)GetByName(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEntityInfo) error {
	path := "entity.getByName"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntityByName(in.Uid, in.Key)
	if info == nil {
		out.Status = outError(path,"not found the entity by name", pb.ResultStatus_NotExisted)
		return nil
	}
	out.Info = switchEntity(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService)RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "entity.removeOne"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	err := cache.Context().RemoveEntity(in.Uid, in.Operator)
	out.Uid = in.Uid
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService)GetAllByOwner(ctx context.Context, in *pb.ReqEntityBy, out *pb.ReplyEntityList) error {
	path := "entity.getByOwner"
	inLog(path, in)
	out.Flag = in.Owner
	out.List = make([]*pb.EntityInfo, 0, 10)
	if len(in.Owner) > 0 {
		for _, value := range cache.Context().AllEntities() {
			if value.Owner == in.Owner && value.Status == cache.EntityStatus(in.Status) {
				out.List = append(out.List, switchEntity(value))
			}
		}
	}else{
		for _, value := range cache.Context().AllEntities() {
			if value.Status == cache.EntityStatus(in.Status) {
				out.List = append(out.List, switchEntity(value))
			}
		}
	}
	out.Page = uint32(in.Page)
	out.Total = uint32(len(out.List))
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *EntityService)SearchPublic(ctx context.Context, in *pb.ReqEntitySearch, out *pb.ReplyEntityList) error {
	path := "entity.searchPublic"
	inLog(path, in)
	out.Flag = ""
	out.List = make([]*pb.EntityInfo, 0, 200)
	for _, value := range cache.Context().AllEntities() {
		if value.Status == cache.EntityStatusUsable && value.IsSatisfy(in.Concept, in.Attribute, in.Tags){
			out.List = append(out.List, switchEntity(value))
		}
	}
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *EntityService)UpdateTags(ctx context.Context, in *pb.RequestList, out *pb.ReplyEntityUpdate) error {
	path := "entity.updateTags"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateTags(in.List, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Uid = info.UID
	out.List = info.Tags
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService)UpdateProperties(ctx context.Context, in *pb.ReqEntityProperties, out *pb.ReplyEntityProperties) error {
	path := "entity.updateProps"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	list := make([]*proxy.PropertyInfo, 0, len(in.Properties))
	for _, value := range in.Properties {
		prop := new(proxy.PropertyInfo)
		prop.Key = value.Uid
		prop.Words = make([]proxy.WordInfo, 0, len(value.Words))
		for _, word := range value.Words {
			prop.Words = append(prop.Words, proxy.WordInfo{UID:word.Uid, Name:word.Name})
		}
		list = append(list, prop)
	}
	err := info.UpdateProperties(list, "")
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Uid = info.UID
	out.Properties = make([]*pb.PropertyInfo, 0, len(info.Properties()))
	for _, value := range info.Properties() {
		tmp := switchEntityProperty(value)
		out.Properties = append(out.Properties, tmp)
	}
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService)UpdateBase(ctx context.Context, in *pb.ReqEntityBase, out *pb.ReplyInfo) error {
	path := "entity.updateBase"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	if in.Name != "" && cache.Context().HadEntityByName(in.Name, in.Add){
		out.Status = outError(path,"the entity name is repeated", pb.ResultStatus_Repeated)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateBase(in.Name, in.Desc, in.Add, in.Concept, in.Cover, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService)UpdateCover(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "entity.updateCover"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateCover(in.Key, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService)UpdateStatus(ctx context.Context, in *pb.ReqEntityStatus, out *pb.ReplyEntityStatus) error {
	path := "entity.updateStatus"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateStatus(cache.EntityStatus(in.Status), in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.State = in.Status
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService)UpdateSynonyms(ctx context.Context, in *pb.RequestList, out *pb.ReplyEntityUpdate) error {
	path := "entity.updateSynonyms"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateSynonyms(in.List, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Uid = info.UID
	out.List = info.Synonyms
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService)AppendProperty(ctx context.Context, in *pb.ReqEntityProperty, out *pb.ReplyEntityProperties) error {
	path := "entity.appendProp"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	if info.HadProperty(in.Property.Uid) {
		out.Status = outError(path, "the property had existed", pb.ResultStatus_Repeated)
		return nil
	}
	words := make([]proxy.WordInfo, 0, len(in.Property.Words))
	for _, value := range in.Property.Words {
		words = append(words, proxy.WordInfo{UID:value.Uid, Name:value.Name})
	}
	err := info.AddProperty(in.Property.Uid, words)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Properties = make([]*pb.PropertyInfo, 0, len(info.Properties()))
	for _, value := range info.Properties() {
		tmp := switchEntityProperty(value)
		out.Properties = append(out.Properties, tmp)
	}
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService)SubtractProperty(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEntityProperties) error {
	path := "entity.subtractProp"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.RemoveProperty(in.Key)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Properties = make([]*pb.PropertyInfo, 0, len(info.Properties()))
	for _, value := range info.Properties() {
		tmp := switchEntityProperty(value)
		out.Properties = append(out.Properties, tmp)
	}
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService)GetByProperty(ctx context.Context, in *pb.ReqEntityByProp, out *pb.ReplyEntityList) error {
	path := "entity.getByProperty"
	inLog(path, in)
	if len(in.Key) < 1 || len(in.Value) < 1{
		out.Status = outError(path, "the key or value is empty", pb.ResultStatus_Empty)
		return nil
	}
	list := cache.Context().GetEntitiesByProp(in.Key, in.Value)
	out.List = make([]*pb.EntityInfo, 0, 5)
	for _, value := range list {
		out.List = append(out.List, switchEntity(value))
	}

	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}
