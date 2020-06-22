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
	tmp.Description = info.Description
	tmp.Owner = info.Owner
	l := len(info.GetAssets())
	tmp.Assets = make([]string, 0, l)
	for _, value := range info.GetAssets() {
		tmp.Assets = append(tmp.Assets, value.UID)
	}
	tmp.Tags = info.Tags
	tmp.Synonyms = info.Synonyms
	tmp.Add = info.Add

	num := len(info.AllEvents())
	tmp.Events = make([]*pb.EventInfo, 0, num)
	for _, value := range info.AllEvents() {
		tmp.Events = append(tmp.Events, switchEntityEvent(value))
	}

	length := len(info.Properties())
	tmp.Properties = make([]*pb.PropertyInfo, 0, length)
	for _, value := range info.Properties() {
		tmp.Properties = append(tmp.Properties, switchEntityProperty(value))
	}

	return tmp
}

func switchEntityEvent(info *proxy.EventPoint) *pb.EventInfo {
	tmp := new(pb.EventInfo)
	tmp.Id = info.ID
	tmp.Description = info.Description
	tmp.Date = &pb.DateInfo{Uid:info.Date.UID, Name:info.Date.Name, Begin:info.Date.Begin.String(), End:info.Date.End.String()}
	tmp.Place = &pb.PlaceInfo{Uid:info.Place.UID, Name:info.Place.Name, Location:info.Place.Location}
	tmp.Assets = info.Assets
	tmp.Relations = make([]*pb.RelationshipInfo, 0, len(info.Relations))
	for i := 0;i < len(info.Relations);i +=1 {
		r := new(pb.RelationshipInfo)
		r.Name = info.Relations[i].Name
		r.Uid = info.Relations[i].UID
		r.Entity = info.Relations[i].Entity
		r.Category = info.Relations[i].Category
		r.Direction = pb.DirectionType(int32(info.Relations[i].Direction))
		tmp.Relations = append(tmp.Relations, r)
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
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the entity uid is empty")
	}
	err := cache.RemoveEntity(in.Uid)
	out.Uid = in.Uid
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *EntityService)GetListByOwner(ctx context.Context, in *pb.ReqEntityBy, out *pb.ReplyEntityList) error {
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
	err := info.UpdateTags(in.List)
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
		err = info.UpdateCover(in.Cover)
	}else{
		err = info.UpdateBase(in.Name, in.Desc, in.Add, in.Concept)
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
	err := info.UpdateStatus(cache.EntityStatus(in.Status))
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
	err := info.UpdateSynonyms(in.List)
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *EntityService)AppendAsset(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEntityAsset) error {
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the entity uid is empty")
	}
	if len(in.Key) < 1 {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("the entity asset uid is empty")
	}
	info := cache.GetEntity(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the entity by uid")
	}
	if info.AddAsset(in.Key) {
		return nil
	}else{
		return errors.New("add asset to entity error")
	}
}

func (mine *EntityService)SubtractAsset(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEntityAsset) error {
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the entity uid is empty")
	}
	info := cache.GetEntity(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the entity by uid")
	}
	err := info.RemoveAsset(in.Key)
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *EntityService)AppendEvent(ctx context.Context, in *pb.ReqEntityEvent, out *pb.ReplyEntityEvents) error {
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the entity uid is empty")
	}
	if in.Event == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not event that in is empty")
	}
	info := cache.GetEntity(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the entity by uid")
	}
	begin := proxy.Date{}
	end := proxy.Date{}
	if in.Event.Date != nil {
		_ = begin.Parse(in.Event.Date.Begin)
		_ = end.Parse(in.Event.Date.End)
	}

	date := proxy.DateInfo{UID:in.Event.Date.Uid, Name:in.Event.Date.Name, Begin:begin, End:end}
	place := proxy.PlaceInfo{UID:in.Event.Place.Uid, Name:in.Event.Place.Name, Location:in.Event.Place.Location}
	relations := make([]proxy.RelationInfo, 0, len(in.Event.Relations))
	for _, value := range in.Event.Relations {
		relations = append(relations, proxy.RelationInfo{UID:value.Uid, Direction:uint8(value.Direction),
			Name:value.Name, Category:value.Category, Entity:value.Entity})
	}
	_,err := info.AddEvent(date, place, in.Event.Description, relations, in.Event.Assets)
	if err == nil {
		events := info.AllEvents()
		out.Events = make([]*pb.EventInfo, 0, len(events))
		for _, event := range events {
			tmp := switchEntityEvent(event)
			out.Events = append(out.Events, tmp)
		}
	}else{
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *EntityService)SubtractEvent(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEntityEvents) error {
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the entity uid is empty")
	}
	info := cache.GetEntity(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the entity by uid")
	}
	err := info.RemoveEvent(in.Id)
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}else{
		events := info.AllEvents()
		out.Events = make([]*pb.EventInfo, 0, len(events))
		for _, event := range events {
			tmp := switchEntityEvent(event)
			out.Events = append(out.Events, tmp)
		}
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
