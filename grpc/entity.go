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
	tmp.Brief = switchEntityBrief(info)
	tmp.Events = make([]*pb.EventBrief, 0, len(info.StaticEvents))
	for _, event := range info.StaticEvents {
		tmp.Events = append(tmp.Events, switchEventBriefToPB(event))
	}
	tmp.Relations = make([]*pb.RelationBrief, 0, len(info.StaticRelations))
	for _, item := range info.StaticRelations {
		tmp.Relations = append(tmp.Relations, switchRRelationBrief(item))
	}
	length := len(info.Properties)
	tmp.Properties = make([]*pb.PropertyInfo, 0, length)
	for _, value := range info.Properties {
		tmp.Properties = append(tmp.Properties, switchPropertyToPB(value))
	}

	return tmp
}

func switchUserEntity(info *cache.EntityInfo, more bool) *pb.EntityInfo {
	tmp := new(pb.EntityInfo)
	tmp.Brief = switchEntityBrief(info)
	if more {
		events := cache.Context().GetEventsByEntity(info.UID, cache.EventCustom)
		tmp.Events = make([]*pb.EventBrief, 0, len(events))
		for _, event := range events {
			if event.Access == cache.AccessPublic {
				tmp.Events = append(tmp.Events, switchEntityEventBrief(event))
			}
		}
		events2 := cache.Context().GetEventsByEntity(info.UID, cache.EventActivity)
		for _, event := range events2 {
			tmp.Events = append(tmp.Events, switchEntityEventBrief(event))
		}
	}
	tmp.Properties = make([]*pb.PropertyInfo, 0, len(info.Properties))
	for _, value := range info.Properties {
		tmp.Properties = append(tmp.Properties, switchPropertyToPB(value))
	}
	return tmp
}

func switchEntityBrief(info *cache.EntityInfo) *pb.EntityBrief {
	tmp := new(pb.EntityBrief)
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
	tmp.Summary = info.Summary
	tmp.Mark = info.Mark
	tmp.Quote = info.Quote
	tmp.Published = info.Published

	//length := len(info.Properties)
	//tmp.Properties = make([]*pb.PropertyInfo, 0, length)
	//for _, value := range info.Properties {
	//	tmp.Properties = append(tmp.Properties, switchEntityProperty(value))
	//}
	return tmp
}

func switchPropertyToPB(info *proxy.PropertyInfo) *pb.PropertyInfo {
	tmp := new(pb.PropertyInfo)
	tmp.Uid = info.Key
	tmp.Words = make([]*pb.WordInfo, 0, len(info.Words))
	for _, value := range info.Words {
		tmp.Words = append(tmp.Words, &pb.WordInfo{Uid: value.UID, Name: value.Name})
	}
	return tmp
}

func switchPropertyFromPB(info *pb.PropertyInfo) *proxy.PropertyInfo {
	tmp := new(proxy.PropertyInfo)
	tmp.Key = info.Uid
	tmp.Words = make([]proxy.WordInfo, 0, len(info.Words))
	for _, value := range info.Words {
		tmp.Words = append(tmp.Words, proxy.WordInfo{UID: value.Uid, Name: value.Name})
	}
	return tmp
}

func switchEventBriefFromPB(info *pb.EventBrief) *proxy.EventBrief {
	tmp := new(proxy.EventBrief)
	tmp.Name = info.Name
	tmp.Quote = info.Quote
	tmp.Description = info.Desc
	tmp.Place.Name = info.Place.Name
	tmp.Place.UID = info.Place.Uid
	tmp.Place.Location = info.Place.Location
	tmp.Assets = info.Assets
	if tmp.Assets == nil {
		tmp.Assets = make([]string, 0, 0)
	}
	tmp.Tags = info.Tags
	if tmp.Tags == nil {
		tmp.Tags = make([]string, 0, 0)
	}
	tmp.Date.UID = info.Date.Uid
	tmp.Date.Name = info.Date.Name
	tmp.Date.Begin.Parse(info.Date.Begin)
	tmp.Date.End.Parse(info.Date.End)
	return tmp
}

func switchEventBriefToPB(info *proxy.EventBrief) *pb.EventBrief {
	tmp := new(pb.EventBrief)
	tmp.Name = info.Name
	tmp.Quote = info.Quote
	tmp.Desc = info.Description
	tmp.Place = new(pb.PlaceInfo)
	tmp.Place.Name = info.Place.Name
	tmp.Place.Uid = info.Place.UID
	tmp.Place.Location = info.Place.Location
	tmp.Assets = info.Assets
	tmp.Tags = info.Tags
	tmp.Date = new(pb.DateInfo)
	tmp.Date.Uid = info.Date.UID
	tmp.Date.Name = info.Date.Name
	tmp.Date.Begin = info.Date.Begin.String()
	tmp.Date.End = info.Date.End.String()
	return tmp
}

func switchRelationBrief(info *pb.RelationBrief) *proxy.RelationCaseInfo {
	tmp := new(proxy.RelationCaseInfo)
	tmp.Name = info.Name
	tmp.Entity = info.Entity
	tmp.Category = info.Type
	tmp.Direction = uint8(info.Direction)
	tmp.Weight = info.Weight
	return tmp
}

func switchRRelationBrief(info *proxy.RelationCaseInfo) *pb.RelationBrief {
	tmp := new(pb.RelationBrief)
	tmp.Name = info.Name
	tmp.Entity = info.Entity
	tmp.Type = info.Category
	tmp.Direction = uint32(info.Direction)
	tmp.Weight = info.Weight
	return tmp
}

func (mine *EntityService) AddOne(ctx context.Context, in *pb.ReqEntityAdd, out *pb.ReplyEntityInfo) error {
	path := "entity.addOne"
	inLog(path, in)
	if in.Name == "" {
		out.Status = outError(path, "the entity name is empty", pb.ResultStatus_Empty)
		return nil
	}
	if cache.Context().HadEntityByName(in.Name, in.Add) {
		out.Status = outError(path, "the entity name is repeated", pb.ResultStatus_Repeated)
		return nil
	}
	//if len(in.Mark) > 0 && cache.Context().HadEntityByMark(in.Mark) {
	//	out.Status = outError(path,"the entity mark is repeated", pb.ResultStatus_Repeated)
	//	return nil
	//}
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
	info.Quote = in.Quote
	info.Summary = in.Summary
	info.Status = cache.EntityStatus(in.Status)
	info.Mark = in.Mark
	info.StaticEvents = make([]*proxy.EventBrief, 0, len(in.Events))
	for _, event := range in.Events {
		info.StaticEvents = append(info.StaticEvents, switchEventBriefFromPB(event))
	}
	info.StaticRelations = make([]*proxy.RelationCaseInfo, 0, len(in.Relations))
	for _, relation := range in.Relations {
		info.StaticRelations = append(info.StaticRelations, switchRelationBrief(relation))
	}
	info.Properties = make([]*proxy.PropertyInfo, 0, len(in.Relations))
	for _, prop := range in.Properties {
		info.Properties = append(info.Properties, switchPropertyFromPB(prop))
	}

	err := cache.Context().CreateEntity(info)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Info = switchEntity(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEntityInfo) error {
	path := "entity.getOne"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	out.Info = switchEntity(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) GetBrief(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEntityBrief) error {
	path := "entity.getBrief"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	out.Info = switchEntityBrief(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) GetByName(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEntityInfo) error {
	path := "entity.getByName"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity name is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntityByName(in.Uid, in.Key)
	if info == nil {
		out.Status = outError(path, "not found the entity by name", pb.ResultStatus_NotExisted)
		return nil
	}
	out.Info = switchEntity(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) GetByMark(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEntityInfo) error {
	path := "entity.getByMark"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity mark is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntityByMark(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the entity by mark", pb.ResultStatus_NotExisted)
		return nil
	}
	out.Info = switchEntity(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "entity.removeOne"
	inLog(path, in)
	err := cache.Context().RemoveEntity(in.Uid, in.Operator)
	out.Uid = in.Uid
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) GetAllByOwner(ctx context.Context, in *pb.ReqEntityBy, out *pb.ReplyEntityList) error {
	path := "entity.getByOwner"
	inLog(path, in)
	out.Flag = in.Owner
	if len(in.Owner) > 1 {
		array := cache.Context().GetEntitiesByOwnerStatus(in.Owner, in.Concept, cache.EntityStatus(in.Status))
		total, _, list := checkPage(in.Page, in.Number, array)
		out.List = make([]*pb.EntityInfo, 0, in.Number)
		out.Total = uint32(total)
		for _, value := range list.([]*cache.EntityInfo) {
			out.List = append(out.List, switchEntity(value))
		}
	} else {
		array := cache.Context().GetEntitiesByStatus(cache.EntityStatus(in.Status), in.Concept)
		total, _, list := checkPage(in.Page, in.Number, array)
		out.List = make([]*pb.EntityInfo, 0, in.Number)
		out.Total = uint32(total)
		for _, value := range list.([]*cache.EntityInfo) {
			out.List = append(out.List, switchEntity(value))
		}
	}
	out.Page = uint32(in.Page)
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *EntityService) GetListByBox(ctx context.Context, in *pb.RequestPage, out *pb.ReplyEntityList) error {
	path := "entity.getListByBox"
	inLog(path, in)
	out.Flag = in.Parent
	array, err := cache.Context().GetEntitiesByBox(in.Parent, cache.EntityStatus(in.Status))
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	total, _, list := checkPage(in.Page, in.Number, array)
	out.List = make([]*pb.EntityInfo, 0, in.Number)
	out.Total = uint32(total)
	for _, value := range list.([]*cache.EntityInfo) {
		out.List = append(out.List, switchEntity(value))
	}
	out.Page = uint32(in.Page)
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *EntityService) GetListByName(ctx context.Context, in *pb.RequestList, out *pb.ReplyEntityList) error {
	path := "entity.getListByName"
	inLog(path, in)
	if len(in.List) < 1 {
		out.List = make([]*pb.EntityInfo, 0, 1)
		return nil
	}
	out.List = make([]*pb.EntityInfo, 0, len(in.List))
	for i := 0; i < len(in.List); i += 1 {
		array, err := cache.Context().GetEntitiesByName(in.List[i])
		if err == nil {
			for _, info := range array {
				out.List = append(out.List, switchEntity(info))
			}
		}
	}

	out.Total = uint32(len(out.List))
	out.Page = 0
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *EntityService) GetByList(ctx context.Context, in *pb.RequestList, out *pb.ReplyEntityList) error {
	path := "entity.getByList"
	inLog(path, in)

	array, err := cache.Context().GetEntitiesByList(cache.EntityStatus(in.Status), in.List)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}

	out.List = make([]*pb.EntityInfo, 0, len(array))
	out.Total = uint32(len(array))
	for _, value := range array {
		out.List = append(out.List, switchEntity(value))
	}
	out.Page = 0
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *EntityService) GetPublishList(ctx context.Context, in *pb.RequestList, out *pb.ReplyEntityPublic) error {
	path := "entity.getPublishList"
	inLog(path, in)

	out.Systems = make([]*pb.EntityInfo, 0, len(in.List))
	out.Users = make([]*pb.EntityInfo, 0, len(in.List))
	if in.Status == 1 {
		array, err := cache.Context().GetEntitiesByList(cache.EntityStatusUsable, in.List)
		if err == nil {
			for _, value := range array {
				out.Systems = append(out.Systems, switchEntity(value))
			}
		}
	}else if in.Status == 2 {
		list, err := cache.Context().GetCustomEntitiesByList(in.List)
		if err == nil {
			for _, value := range list {
				out.Users = append(out.Users, switchUserEntity(value, true))
			}
		}
	}else{
		array, err := cache.Context().GetEntitiesByList(cache.EntityStatusUsable, in.List)
		rest := make([]string, 0, len(in.List))
		for _, key := range in.List {
			rest = append(rest, key)
		}
		if err == nil {
			for _, value := range array {
				out.Systems = append(out.Systems, switchEntity(value))
				for i := 0; i < len(rest); i += 1 {
					if rest[i] == value.UID {
						rest = append(rest[:i], rest[i+1:]...)
						break
					}
				}
			}
		}
		list, err := cache.Context().GetCustomEntitiesByList(rest)
		if err == nil {
			for _, value := range list {
				out.Users = append(out.Users, switchUserEntity(value, false))
			}
		}
	}

	out.Status = outLog(path, fmt.Sprintf("the system length = %d; user length = %d", len(out.Systems), len(out.Users)))
	return nil
}

func (mine *EntityService) GetByFilter(ctx context.Context, in *pb.RequestFilter, out *pb.ReplyEntityList) error {
	path := "entity.getByFilter"
	inLog(path, in)
	out.Flag = ""
	out.List = make([]*pb.EntityInfo, 0, 200)

	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *EntityService) SearchPublic(ctx context.Context, in *pb.ReqEntitySearch, out *pb.ReplyEntityList) error {
	path := "entity.searchPublic"
	inLog(path, in)
	out.Flag = ""
	out.List = make([]*pb.EntityInfo, 0, 200)
	list := cache.Context().GetArchivedList(in.Name)
	for _, value := range list {
		//if value.IsSatisfy(in.Concept, in.Attribute, in.Tags) {
		//	out.List = append(out.List, switchEntity(value))
		//}
		out.List = append(out.List, switchEntity(value))
	}
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *EntityService) SearchMatch(ctx context.Context, in *pb.ReqEntityMatch, out *pb.ReplyEntityList) error {
	path := "entity.searchMatch"
	inLog(path, in)
	out.Flag = ""
	array := cache.Context().SearchEntities(in.Keywords)
	total, _, list := checkPage(in.Page, in.Number, array)
	out.List = make([]*pb.EntityInfo, 0, total)
	for _, value := range list.([]*cache.EntityInfo) {
		out.List = append(out.List, switchEntity(value))
	}
	out.Page = uint32(in.Page)
	out.Total = uint32(total)
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *EntityService) UpdateTags(ctx context.Context, in *pb.RequestList, out *pb.ReplyEntityUpdate) error {
	path := "entity.updateTags"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateTags(in.List, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Uid = info.UID
	out.List = info.Tags
	out.Updated = uint64(info.UpdateTime.Unix())
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) UpdateProperties(ctx context.Context, in *pb.ReqEntityProperties, out *pb.ReplyEntityProperties) error {
	path := "entity.updateProps"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	list := make([]*proxy.PropertyInfo, 0, len(in.Properties))
	for _, value := range in.Properties {
		prop := new(proxy.PropertyInfo)
		prop.Key = value.Uid
		prop.Words = make([]proxy.WordInfo, 0, len(value.Words))
		for _, word := range value.Words {
			prop.Words = append(prop.Words, proxy.WordInfo{UID: word.Uid, Name: word.Name})
		}
		list = append(list, prop)
	}
	err := info.UpdateProperties(list, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Uid = info.UID
	out.Properties = make([]*pb.PropertyInfo, 0, len(info.Properties))
	for _, value := range info.Properties {
		tmp := switchPropertyToPB(value)
		out.Properties = append(out.Properties, tmp)
	}
	out.Updated = uint64(info.UpdateTime.Unix())
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) UpdateBase(ctx context.Context, in *pb.ReqEntityBase, out *pb.ReplyInfo) error {
	path := "entity.updateBase"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	if in.Name != "" && in.Name != info.Name && cache.Context().HadEntityByName(in.Name, in.Add) {
		out.Status = outError(path, "the entity name is repeated", pb.ResultStatus_Repeated)
		return nil
	}
	err := info.UpdateBase(in.Name, in.Desc, in.Add, in.Concept, in.Cover, in.Mark, in.Quote, in.Summary, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Updated = uint64(info.UpdateTime.Unix())
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) UpdateCover(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "entity.updateCover"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateCover(in.Key, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.Updated = uint64(info.UpdateTime.Unix())
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) UpdateStatus(ctx context.Context, in *pb.ReqEntityStatus, out *pb.ReplyEntityStatus) error {
	path := "entity.updateStatus"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateStatus(cache.EntityStatus(in.Status), in.Operator, in.Remark)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.State = in.Status
	out.Updated = uint64(info.UpdateTime.Unix())
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) UpdateSynonyms(ctx context.Context, in *pb.RequestList, out *pb.ReplyEntityUpdate) error {
	path := "entity.updateSynonyms"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateSynonyms(in.List, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Uid = info.UID
	out.List = info.Synonyms
	out.Updated = uint64(info.UpdateTime.Unix())
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) AppendProperty(ctx context.Context, in *pb.ReqEntityProperty, out *pb.ReplyEntityProperties) error {
	path := "entity.appendProp"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	if info.HadProperty(in.Property.Uid) {
		out.Status = outError(path, "the property had existed", pb.ResultStatus_Repeated)
		return nil
	}
	words := make([]proxy.WordInfo, 0, len(in.Property.Words))
	for _, value := range in.Property.Words {
		words = append(words, proxy.WordInfo{UID: value.Uid, Name: value.Name})
	}
	err := info.AddProperty(in.Property.Uid, words)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Properties = make([]*pb.PropertyInfo, 0, len(info.Properties))
	for _, value := range info.Properties {
		tmp := switchPropertyToPB(value)
		out.Properties = append(out.Properties, tmp)
	}
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) SubtractProperty(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEntityProperties) error {
	path := "entity.subtractProp"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the entity uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetEntity(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the entity by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.RemoveProperty(in.Key)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Properties = make([]*pb.PropertyInfo, 0, len(info.Properties))
	for _, value := range info.Properties {
		tmp := switchPropertyToPB(value)
		out.Properties = append(out.Properties, tmp)
	}
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) GetByProperty(ctx context.Context, in *pb.ReqEntityByProp, out *pb.ReplyEntityList) error {
	path := "entity.getByProperty"
	inLog(path, in)
	if len(in.Key) < 1 || len(in.Value) < 1 {
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

func (mine *EntityService) UpdateStatic(ctx context.Context, in *pb.ReqEntityStatic, out *pb.ReplyInfo) error {
	path := "entity.updateStatic"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	entity := cache.Context().GetEntity(in.Uid)
	if entity == nil {
		out.Status = outError(path, "not found the entity", pb.ResultStatus_Empty)
		return nil
	}
	if entity.Name != in.Name || entity.Add != in.Add {
		if cache.Context().HadEntityByName(in.Name, in.Add) {
			out.Status = outError(path, "the entity name is repeated", pb.ResultStatus_Repeated)
			return nil
		}
	}

	info := new(cache.EntityInfo)
	info.Name = in.Name
	info.Description = in.Desc
	info.Add = in.Add
	info.Operator = in.Operator
	info.Cover = in.Cover
	info.Concept = in.Concept
	info.Synonyms = in.Synonyms
	info.Tags = in.Tags
	info.Quote = in.Quote
	info.Summary = in.Summary
	info.Mark = in.Mark
	info.StaticEvents = make([]*proxy.EventBrief, 0, len(in.Events))
	for _, event := range in.Events {
		info.StaticEvents = append(info.StaticEvents, switchEventBriefFromPB(event))
	}
	info.StaticRelations = make([]*proxy.RelationCaseInfo, 0, len(in.Relations))
	for _, relation := range in.Relations {
		info.StaticRelations = append(info.StaticRelations, switchRelationBrief(relation))
	}
	info.Properties = make([]*proxy.PropertyInfo, 0, len(in.Properties))
	for _, prop := range in.Properties {
		info.Properties = append(info.Properties, switchPropertyFromPB(prop))
	}
	err := entity.UpdateStatic(info)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Updated = uint64(info.UpdateTime.Unix())
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) UpdateRelations(ctx context.Context, in *pb.ReqEntityRelations, out *pb.ReplyInfo) error {
	path := "entity.updateStRelations"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	entity := cache.Context().GetEntity(in.Uid)
	if entity == nil {
		out.Status = outError(path, "not found the entity", pb.ResultStatus_Empty)
		return nil
	}

	relations := make([]*proxy.RelationCaseInfo, 0, len(in.Relations))
	for _, relation := range in.Relations {
		relations = append(relations, switchRelationBrief(relation))
	}

	err := entity.UpdateStaticRelations(in.Operator, relations)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Updated = uint64(entity.UpdateTime.Unix())
	out.Status = outLog(path, out)
	return nil
}

func (mine *EntityService) UpdateEvents(ctx context.Context, in *pb.ReqEntityEvents, out *pb.ReplyInfo) error {
	path := "entity.updateStEvents"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	entity := cache.Context().GetEntity(in.Uid)
	if entity == nil {
		out.Status = outError(path, "not found the entity", pb.ResultStatus_Empty)
		return nil
	}
	events := make([]*proxy.EventBrief, 0, len(in.Events))
	for _, event := range in.Events {
		events = append(events, switchEventBriefFromPB(event))
	}

	err := entity.UpdateStaticEvents(in.Operator, events)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Updated = uint64(entity.UpdateTime.Unix())
	out.Status = outLog(path, out)
	return nil
}
