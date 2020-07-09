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
	tmp.Created = info.CreateTime.Unix()
	tmp.Updated = info.UpdateTime.Unix()
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

func (mine *EventService)AddOne(ctx context.Context, in *pb.ReqEventAdd, out *pb.ReplyEventOne) error {
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
	err := cache.RemoveEvent(in.Uid, in.Operator)
	out.Uid = in.Uid
	return err
}

func (mine *EventService)GetList(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEventList) error {
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