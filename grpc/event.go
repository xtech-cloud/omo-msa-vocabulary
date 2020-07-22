package grpc

import (
	"context"
	"errors"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
	"omo.msa.vocabulary/proxy"
)

type EventService struct {}

func switchEntityEvent(info *cache.EventInfo) *pb.EventInfo {
	tmp := new(pb.EventInfo)
	tmp.Id = info.ID
	tmp.Uid = info.UID
	tmp.Operator = info.Operator
	tmp.Creator = info.Creator
	tmp.Created = info.CreateTime.Unix()
	tmp.Updated = info.UpdateTime.Unix()
	tmp.Parent = info.Parent
	tmp.Name = info.Name
	tmp.Description = info.Description
	tmp.Date = &pb.DateInfo{Uid:info.Date.UID, Name:info.Date.Name, Begin:info.Date.Begin.String(), End:info.Date.End.String()}
	tmp.Place = &pb.PlaceInfo{Uid:info.Place.UID, Name:info.Place.Name, Location:info.Place.Location}
	tmp.Assets = info.Assets
	tmp.Relations = make([]*pb.RelationshipInfo, 0, len(info.Relations))
	for i := 0;i < len(info.Relations);i +=1 {
		tmp.Relations = append(tmp.Relations, switchRelationIns(&info.Relations[i]))
	}
	return tmp
}

func switchRelationIns(info *proxy.RelationInfo) *pb.RelationshipInfo {
	r := new(pb.RelationshipInfo)
	r.Name = info.Name
	r.Uid = info.UID
	r.Entity = info.Entity
	r.Category = info.Category
	r.Direction = pb.DirectionType(int32(info.Direction))
	return r
}

func (mine *EventService)AddOne(ctx context.Context, in *pb.ReqEventAdd, out *pb.ReplyEventOne) error {
	inLog("event.add", in)
	info := cache.GetEntity(in.Parent)
	if info == nil {
		out.ErrorCode  = pb.ResultStatus_NotExisted
		return errors.New("not found the entity by uid")
	}
	begin := proxy.Date{}
	end := proxy.Date{}
	if in.Date != nil {
		_ = begin.Parse(in.Date.Begin)
		_ = end.Parse(in.Date.End)
	}

	date := proxy.DateInfo{UID:in.Date.Uid, Name:in.Date.Name, Begin:begin, End:end}
	place := proxy.PlaceInfo{UID:in.Place.Uid, Name:in.Place.Name, Location:in.Place.Location}
	relations := make([]proxy.RelationInfo, 0, len(in.Relations))
	for _, value := range in.Relations {
		relations = append(relations, proxy.RelationInfo{UID:value.Uid, Direction:uint8(value.Direction),
			Name:value.Name, Category:value.Category, Entity:value.Entity})
	}
	event,err := info.AddEvent(date, place,in.Name, in.Description, in.Operator, relations, in.Assets)
	if err == nil {
		out.Info = switchEntityEvent(event)
	}else{
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *EventService)GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEventOne) error {
	inLog("event.one", in)
	if len(in.Uid) > 0 {
		info := cache.GetEvent(in.Uid)
		if info == nil {
			out.ErrorCode  = pb.ResultStatus_NotExisted
			return errors.New("not found the attribute by uid")
		}
		out.Info = switchEntityEvent(info)
	}else{
		out.ErrorCode  = pb.ResultStatus_Empty
		return errors.New("the uid or key all is empty")
	}
	return nil
}

func (mine *EventService)RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	inLog("event.remove", in)
	err := cache.RemoveEvent(in.Uid, in.Operator)
	out.Uid = in.Uid
	return err
}

func (mine *EventService)GetList(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEventList) error {
	inLog("event.list", in)
	info := cache.GetEntity(in.Uid)
	if info == nil {
		return errors.New("not found the entity by uid")
	}
	out.List = make([]*pb.EventInfo, 0, 10)
	for _, value := range info.AllEvents() {
		out.List = append(out.List, switchEntityEvent(value))
	}
	return nil
}

func (mine *EventService)Update(ctx context.Context, in *pb.ReqEventUpdate, out *pb.ReplyEventOne) error {
	inLog("event.update", in)
	info := cache.GetEvent(in.Uid)
	if info == nil {
		out.ErrorCode  = pb.ResultStatus_NotExisted
		return errors.New("not found the attribute by uid")
	}
	begin := proxy.Date{}
	end := proxy.Date{}
	if in.Date != nil {
		_ = begin.Parse(in.Date.Begin)
		_ = end.Parse(in.Date.End)
	}
	date := proxy.DateInfo{UID:in.Date.Uid, Name:in.Date.Name, Begin:begin, End:end}
	place := proxy.PlaceInfo{UID:in.Place.Uid, Name:in.Place.Name, Location:in.Place.Location}
	err := info.UpdateBase(in.Name, in.Description, in.Operator, date, place)
	if err != nil {
		return err
	}
	out.Info = switchEntityEvent(info)
	return nil
}

func (mine *EventService)AppendAsset(ctx context.Context, in *pb.ReqEventAsset, out *pb.ReplyEventAsset) error {
	info := cache.GetEvent(in.Uid)
	if info == nil {
		return errors.New("not found the attribute by uid")
	}

	err := info.AppendAsset(in.Asset)
	if err != nil {
		return err
	}
	out.Uid = in.Uid
	out.Assets = info.Assets
	return nil
}

func (mine *EventService)SubtractAsset(ctx context.Context, in *pb.ReqEventAsset, out *pb.ReplyEventAsset) error {
	info := cache.GetEvent(in.Uid)
	if info == nil {
		return errors.New("not found the attribute by uid")
	}

	err := info.SubtractAsset(in.Asset)
	if err != nil {
		return err
	}
	out.Uid = in.Uid
	out.Assets = info.Assets
	return nil
}

func (mine *EventService)AppendRelation(ctx context.Context, in *pb.ReqEventRelation, out *pb.ReplyEventRelation) error {
	info := cache.GetEvent(in.Uid)
	if info == nil {
		return errors.New("not found the attribute by uid")
	}
	tmp := proxy.RelationInfo{UID:in.Relation.Uid, Direction:uint8(in.Relation.Direction),
		Name:in.Relation.Name, Category:in.Relation.Category, Entity:in.Relation.Entity}
	err := info.AppendRelation(tmp)
	if err != nil {
		return err
	}
	out.Relations = make([]*pb.RelationshipInfo, 0, len(info.Relations))
	for i := 0;i < len(info.Relations);i +=1 {
		out.Relations = append(out.Relations, switchRelationIns(&info.Relations[i]))
	}
	return nil
}

func (mine *EventService)SubtractRelation(ctx context.Context, in *pb.ReqEventRelation, out *pb.ReplyEventRelation) error {
	info := cache.GetEvent(in.Uid)
	if info == nil {
		return errors.New("not found the attribute by uid")
	}
	err := info.SubtractRelation(in.Relation.Uid)
	if err != nil {
		return err
	}
	out.Relations = make([]*pb.RelationshipInfo, 0, len(info.Relations))
	for i := 0;i < len(info.Relations);i +=1 {
		out.Relations = append(out.Relations, switchRelationIns(&info.Relations[i]))
	}
	return nil
}