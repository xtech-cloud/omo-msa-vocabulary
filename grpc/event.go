package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/micro/go-micro/v2/logger"
	pbstaus "github.com/xtech-cloud/omo-msp-status/proto/status"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
	"omo.msa.vocabulary/proxy"
	"strconv"
	"time"
)

type EventService struct{}

func switchEntityEvent(info *cache.EventInfo) *pb.EventInfo {
	tmp := new(pb.EventInfo)
	tmp.Id = info.ID
	tmp.Uid = info.UID
	tmp.Operator = info.Operator
	tmp.Creator = info.Creator
	tmp.Created = info.CreateTime.Unix()
	tmp.Updated = info.UpdateTime.Unix()
	tmp.Parent = info.Entity
	tmp.Name = info.Name
	tmp.Quote = info.Quote
	tmp.Type = uint32(info.Type)
	tmp.Description = info.Description
	tmp.Date = &pb.DateInfo{Uid: info.Date.UID, Name: info.Date.Name, Begin: info.Date.Begin.String(), End: info.Date.End.String()}
	tmp.Place = &pb.PlaceInfo{Uid: info.Place.UID, Name: info.Place.Name, Location: info.Place.Location}
	tmp.Assets = info.Assets
	tmp.Tags = info.Tags
	tmp.Cover = info.Cover
	tmp.Access = uint32(info.Access)
	tmp.Relations = make([]*pb.RelationshipInfo, 0, len(info.Relations))
	for i := 0; i < len(info.Relations); i += 1 {
		tmp.Relations = append(tmp.Relations, switchRelationIns(&info.Relations[i]))
	}
	return tmp
}

func switchEntityEventBrief(info *cache.EventInfo) *pb.EventBrief {
	tmp := new(pb.EventBrief)
	tmp.Uid = info.UID
	tmp.Name = info.Name
	tmp.Created = info.CreateTime.Unix()
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

func switchRelationIns(info *proxy.RelationCaseInfo) *pb.RelationshipInfo {
	r := new(pb.RelationshipInfo)
	r.Name = info.Name
	r.Uid = info.UID
	r.Entity = info.Entity
	r.Category = info.Category
	r.Direction = pb.DirectionType(int32(info.Direction))
	return r
}

func (mine *EventService) AddOne(ctx context.Context, in *pb.ReqEventAdd, out *pb.ReplyEventInfo) error {
	path := "event.addOne"
	inLog(path, in)
	info := cache.Context().GetEntity(in.Parent)
	if info == nil {
		out.Status = outError(path, "not found the entity by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	//event := info.GetEventBy(in.Date.Begin, in.Place.Name)
	//if event != nil {
	//	er := event.UpdateInfo(in.Name, in.Description, in.Operator)
	//	if er != nil {
	//		out.Status = outError(path, er.Error(), pbstaus.ResultStatus_DBException)
	//	}else{
	//		out.Info = switchEntityEvent(event)
	//		out.Status = outLog(path, out)
	//	}
	//	return nil
	//}
	begin := proxy.Date{}
	end := proxy.Date{}
	if in.Date != nil {
		_ = begin.Parse(in.Date.Begin)
		_ = end.Parse(in.Date.End)
	}

	date := proxy.DateInfo{UID: in.Date.Uid, Name: in.Date.Name, Begin: begin, End: end}
	place := proxy.PlaceInfo{UID: in.Place.Uid, Name: in.Place.Name, Location: in.Place.Location}
	relations := make([]proxy.RelationCaseInfo, 0, len(in.Relations))
	for _, value := range in.Relations {
		relations = append(relations, proxy.RelationCaseInfo{UID: value.Uid, Direction: uint8(value.Direction),
			Name: value.Name, Category: value.Category, Entity: value.Entity})
	}
	event, err := info.AddEvent(date, place, in.Name, in.Description, in.Cover, in.Quote, in.Operator, uint8(in.Type), uint8(in.Access), relations, in.Tags, in.Assets)
	if err == nil {
		out.Info = switchEntityEvent(event)
		out.Status = outLog(path, out)
	} else {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
	}
	return nil
}

func (mine *EventService) GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEventInfo) error {
	path := "event.getOne"
	inLog(path, in)
	if in.Uid == "" {
		out.Status = outError(path, "the uid is empty", pbstaus.ResultStatus_Empty)
	}
	if in.Key == "" {
		info := cache.Context().GetEvent(in.Uid)
		if info == nil {
			out.Status = outError(path, "not found the event by uid", pbstaus.ResultStatus_NotExisted)
			return nil
		}
		out.Info = switchEntityEvent(info)
		out.Status = outLog(path, out)
	} else if in.Key == "asset" {
		info := cache.Context().GetEventByAsset(in.Uid)
		if info == nil {
			out.Status = outError(path, "not found the event by uid", pbstaus.ResultStatus_NotExisted)
			return nil
		}
		out.Info = switchEntityEvent(info)
		out.Status = outLog(path, out)
	}
	return nil
}

func (mine *EventService) RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "event.removeOne"
	inLog(path, in)
	err := cache.Context().RemoveEvent(in.Uid, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.Status = outLog(path, out)
	return nil
}

func (mine *EventService) GetList(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyEventList) error {
	path := "event.getList"
	inLog(path, in)
	info := cache.Context().GetEntity(in.Uid)
	out.List = make([]*pb.EventInfo, 0, 10)
	if info == nil {
		//out.Status = outError(path, "not found the entity by uid", pbstaus.ResultStatus_NotExisted)
		//return nil
		list := cache.Context().GetEventsByQuote(in.Key)
		for _, value := range list {
			out.List = append(out.List, switchEntityEvent(value))
		}
	} else {
		if in.Operator != "" {
			acc, er := strconv.ParseUint(in.Operator, 10, 32)
			if er == nil {
				list := info.GetEventsByAccess(uint8(in.Id), uint8(acc))
				for _, value := range list {
					out.List = append(out.List, switchEntityEvent(value))
				}
			}
		} else if in.Id > 0 {
			list := info.GetEventsByType(uint8(in.Id), in.Key)
			for _, value := range list {
				out.List = append(out.List, switchEntityEvent(value))
			}
		} else {
			if len(in.Key) > 1 {
				list := info.GetEventsByQuote(in.Key)
				for _, value := range list {
					out.List = append(out.List, switchEntityEvent(value))
				}
			} else {
				for _, value := range info.AllEvents() {
					out.List = append(out.List, switchEntityEvent(value))
				}
			}
		}
	}

	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *EventService) GetByFilter(ctx context.Context, in *pb.RequestFilter, out *pb.ReplyEventList) error {
	path := "event.getByFilter"
	inLog(path, in)
	var err error
	var total int32
	var pages int32
	var list []*cache.EventInfo

	if in.Key == "entities_sys" {
		list = make([]*cache.EventInfo, 0, 200)
		for _, uid := range in.Values {
			arr := cache.Context().GetEventsByEntity(uid, 1)
			list = append(list, arr...)
		}
		total = int32(len(list))
		pages = 1
	} else if in.Key == "relate" {
		list = cache.Context().GetEventsByRelate(in.Parent, in.Value)
	} else if in.Key == "quote" {
		total, pages, list = cache.Context().GetEventsByQuotePage(in.Value, in.Page, in.Number)
	} else if in.Key == "assets" {
		total, pages, list = cache.Context().GetEventsAssetsByQuote(in.Value, in.Page, in.Number)
	} else if in.Key == "type" {
		tp, er := strconv.Atoi(in.Value)
		if er == nil {
			total, pages, list = cache.Context().GetEventsByEntityType(in.Parent, int32(tp), in.Page, in.Number)
		}
	} else if in.Key == "week" {
		//从指定时间开始7天内的活动事件列表
		utc, er := strconv.Atoi(in.Value)
		if er != nil {
			out.Status = outError(path, er.Error(), pbstaus.ResultStatus_DBException)
			return nil
		}
		list = cache.Context().GetEventsByWeek(int64(utc), in.Values)
	} else {
		err = errors.New("not define the key")
	}
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.List = make([]*pb.EventInfo, 0, len(list))
	for _, event := range list {
		out.List = append(out.List, switchEntityEvent(event))
	}
	out.Total = uint32(total)
	out.Pages = uint32(pages)
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *EventService) GetStatistic(ctx context.Context, in *pb.RequestFilter, out *pb.ReplyStatistic) error {
	path := "event.getStatistic"
	inLog(path, in)
	if in.Key == "count" {
		out.Count = cache.Context().GetEventsCountByEntity(in.Parent)
	} else if in.Key == "date" {
		date, err := time.Parse("2006-01-02", in.Value)
		if err != nil {
			out.Status = outError(path, err.Error(), pbstaus.ResultStatus_Empty)
			return nil
		}
		out.List = cache.Context().GetActivityCountBy(in.Values, date)
	} else if in.Key == "analyse" {
		out.List = cache.Context().GetEventCountBy(in.Value)
	} else if in.Key == "today" {
		events := cache.Context().GetEventsByQuote(in.Value)
		now := time.Now()
		from := time.Date(now.Year(), now.Month(), now.Day(), 0, 1, 0, 0, time.UTC).Unix()
		to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, time.UTC).Unix()
		for _, event := range events {
			unix := event.CreateTime.Unix()
			if unix < to && unix > from {
				out.Count += 1
			}
		}
	} else if in.Key == "quote" {
		//参与人数或者被引用次数
		out.Count = uint32(cache.Context().GetEventCountByQuote(in.Value))
	} else if in.Key == "opus" {
		//作品数量
		assets := cache.Context().GetEventAssetsByQuote(in.Value)
		out.Count = uint32(len(assets))
	} else if in.Key == "sex" {
		events := cache.Context().GetEventsByQuote(in.Value)
		var boy uint32 = 0
		var girl uint32 = 0
		sex := cache.Context().GetAttributeByKey("sex")
		if sex != nil {
			for _, event := range events {
				entity := cache.Context().GetEntity(event.Entity)
				if entity != nil {
					prop := entity.GetProperty(sex.UID)
					if prop != nil && len(prop.Words) > 0 {
						logger.Warn(entity.Name + " that sex = " + prop.Words[0].Name)
						if prop.Words[0].Name == "1" || prop.Words[0].Name == "男" {
							boy += 1
						} else {
							girl += 1
						}
					}
				}
			}
		}

		out.List = make([]*pb.StatisticInfo, 0, 2)
		out.List = append(out.List, &pb.StatisticInfo{Key: "1", Count: boy})
		out.List = append(out.List, &pb.StatisticInfo{Key: "0", Count: girl})
	}

	out.Owner = in.Value
	out.Key = in.Key
	out.Status = outLog(path, out)
	return nil
}

func (mine *EventService) UpdateBase(ctx context.Context, in *pb.ReqEventUpdate, out *pb.ReplyEventInfo) error {
	path := "event.updateBase"
	inLog(path, in)
	info := cache.Context().GetEvent(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the event by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	begin := proxy.Date{}
	end := proxy.Date{}
	if in.Date != nil {
		_ = begin.Parse(in.Date.Begin)
		_ = end.Parse(in.Date.End)
	}
	date := proxy.DateInfo{UID: in.Date.Uid, Name: in.Date.Name, Begin: begin, End: end}
	place := proxy.PlaceInfo{UID: in.Place.Uid, Name: in.Place.Name, Location: in.Place.Location}
	err := info.UpdateBase(in.Name, in.Description, in.Operator, uint8(in.Access), date, place, in.Assets)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}

	out.Info = switchEntityEvent(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *EventService) UpdateTags(ctx context.Context, in *pb.RequestList, out *pb.ReplyInfo) error {
	path := "event.updateTags"
	inLog(path, in)
	info := cache.Context().GetEvent(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the event by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateTags("", in.List)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.Status = outLog(path, out)
	return nil
}

func (mine *EventService) UpdateCover(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "event.cover"
	inLog(path, in)
	info := cache.Context().GetEvent(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the event by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateCover(in.Operator, in.Key)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Uid = info.UID
	out.Status = outLog(path, out)
	return nil
}

func (mine *EventService) UpdateQuote(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "event.quote"
	inLog(path, in)
	info := cache.Context().GetEvent(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the event by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateQuote(in.Key, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Uid = info.UID
	out.Status = outLog(path, out)
	return nil
}

func (mine *EventService) UpdateAccess(ctx context.Context, in *pb.ReqEventAccess, out *pb.ReplyInfo) error {
	path := "event.access"
	inLog(path, in)
	info := cache.Context().GetEvent(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the event by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateAccess(in.Operator, uint8(in.Access))
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Uid = info.UID
	out.Status = outLog(path, out)
	return nil
}

func (mine *EventService) UpdateAssets(ctx context.Context, in *pb.RequestList, out *pb.ReplyEventAssets) error {
	path := "event.updateAssets"
	inLog(path, in)
	info := cache.Context().GetEvent(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the event by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateAssets(in.Operator, in.List)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.Assets = info.Assets
	out.Status = outLog(path, out)
	return nil
}

func (mine *EventService) AppendAsset(ctx context.Context, in *pb.ReqEventAsset, out *pb.ReplyEventAssets) error {
	path := "event.appendAsset"
	inLog(path, in)
	info := cache.Context().GetEvent(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the event by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}

	err := info.AppendAsset(in.Asset)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.Assets = info.Assets
	out.Status = outLog(path, out)
	return nil
}

func (mine *EventService) SubtractAsset(ctx context.Context, in *pb.ReqEventAsset, out *pb.ReplyEventAssets) error {
	path := "event.subtractAsset"
	inLog(path, in)
	info := cache.Context().GetEvent(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the event by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}

	err := info.SubtractAsset(in.Asset)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.Assets = info.Assets
	out.Status = outLog(path, out)
	return nil
}

func (mine *EventService) AppendRelation(ctx context.Context, in *pb.ReqEventRelation, out *pb.ReplyEventRelations) error {
	path := "event.appendRelation"
	inLog(path, in)
	info := cache.Context().GetEvent(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the event by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	tmp := proxy.RelationCaseInfo{Direction: uint8(in.Relation.Direction),
		Name: in.Relation.Name, Category: in.Relation.Category, Entity: in.Relation.Entity}
	err := info.AppendRelation(&tmp)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Relations = make([]*pb.RelationshipInfo, 0, len(info.Relations))
	for i := 0; i < len(info.Relations); i += 1 {
		out.Relations = append(out.Relations, switchRelationIns(&info.Relations[i]))
	}
	out.Status = outLog(path, out)
	return nil
}

func (mine *EventService) SubtractRelation(ctx context.Context, in *pb.ReqEventRelation, out *pb.ReplyEventRelations) error {
	path := "event.subtractRelation"
	inLog(path, in)
	info := cache.Context().GetEvent(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the event by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	err := info.SubtractRelation(in.Relation.Uid)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Relations = make([]*pb.RelationshipInfo, 0, len(info.Relations))
	for i := 0; i < len(info.Relations); i += 1 {
		out.Relations = append(out.Relations, switchRelationIns(&info.Relations[i]))
	}
	out.Status = outLog(path, out)
	return nil
}

func (mine *EventService) UpdateByFilter(ctx context.Context, in *pb.ReqUpdateFilter, out *pb.ReplyInfo) error {
	path := "event.updateByFilter"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the uid is empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	entity := cache.Context().GetEntity(in.Uid)
	if entity == nil {
		out.Status = outError(path, "not found the entity", pbstaus.ResultStatus_Empty)
		return nil
	}
	var err error
	if in.Key == "pushed" {
		err = entity.UpdatePushTime(in.Operator)
	}
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Updated = uint64(entity.UpdateTime.Unix())
	out.Status = outLog(path, out)
	return nil
}
