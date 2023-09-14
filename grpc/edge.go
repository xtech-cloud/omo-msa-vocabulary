package grpc

import (
	"context"
	"fmt"
	pbstaus "github.com/xtech-cloud/omo-msp-status/proto/status"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
	"omo.msa.vocabulary/proxy"
)

type VEdgeService struct{}

func switchVEdge(info *cache.VEdgeInfo) *pb.VEdgeInfo {
	tmp := new(pb.VEdgeInfo)
	tmp.Uid = info.UID
	tmp.Name = info.Name
	tmp.Id = int64(info.ID)
	tmp.Uid = info.UID
	tmp.Operator = info.Operator
	tmp.Creator = info.Creator
	tmp.Created = info.CreateTime.Unix()
	tmp.Updated = info.UpdateTime.Unix()
	tmp.Source = info.Source
	tmp.Target = &pb.VNode{
		Name:   info.Target.Name,
		Uid:    info.Target.UID,
		Entity: info.Target.Entity,
		Thumb:  info.Target.Thumb,
	}
	tmp.Direction = uint32(info.Direction)
	tmp.Category = info.Relation
	tmp.Weight = info.Weight
	tmp.Center = info.Center
	return tmp
}

func (mine *VEdgeService) AddOne(ctx context.Context, in *pb.ReqVEdgeAdd, out *pb.ReplyVEdgeInfo) error {
	path := "vedge.addOne"
	inLog(path, in)
	entity := cache.Context().GetEntity(in.Center)
	if entity == nil {
		out.Status = outError(path, "not found the entity by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	info, err := entity.CreateVEdge(in.Source, in.Name, in.Relation, in.Operator, in.Direction, in.Weight, proxy.VNode{Entity: in.Target, Name: in.Label})
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Info = switchVEdge(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *VEdgeService) GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyVEdgeInfo) error {
	path := "vedge.getOne"
	inLog(path, in)
	info, err := cache.Context().GetVEdge(in.Uid)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_NotExisted)
		return nil
	}
	out.Info = switchVEdge(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *VEdgeService) RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "vedge.removeOne"
	inLog(path, in)
	var err error
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the uid is empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	err = cache.Context().RemoveVEdge(in.Uid, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.Key = in.Key
	out.Status = outLog(path, out)
	return nil
}

func (mine *VEdgeService) GetAll(ctx context.Context, in *pb.RequestFilter, out *pb.ReplyVEdgeList) error {
	path := "vedge.getAll"
	inLog(path, in)
	var array []*cache.VEdgeInfo
	if in.Key == "" {
		array = cache.Context().GetVEdgesByCenter(in.Parent)
	} else if in.Key == "public" {

	}

	out.List = make([]*pb.VEdgeInfo, 0, len(array))
	for _, value := range array {
		out.List = append(out.List, switchVEdge(value))
	}
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *VEdgeService) GetStatistic(ctx context.Context, in *pb.RequestFilter, out *pb.ReplyStatistic) error {
	path := "vedge.getStatistic"
	inLog(path, in)
	out.Status = outError(path, "param is empty", pbstaus.ResultStatus_Empty)
	return nil
}

func (mine *VEdgeService) UpdateInfo(ctx context.Context, in *pb.ReqVEdgeUpdate, out *pb.ReplyVEdgeInfo) error {
	path := "vedge.updateInfo"
	inLog(path, in)
	var err error
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the uid is empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	info, err := cache.Context().GetVEdge(in.Uid)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_NotExisted)
		return nil
	}
	err = info.UpdateBase(in.Name, in.Relation, in.Operator, uint8(in.Direction), proxy.VNode{UID: info.Target.UID, Name: in.Label, Entity: in.Target, Thumb: in.Thumb})
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Info = switchVEdge(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *VEdgeService) UpdateByFilter(ctx context.Context, in *pb.ReqUpdateFilter, out *pb.ReplyInfo) error {
	path := "vedge.updateByFilter"
	inLog(path, in)
	var err error
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the uid is empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	_, err = cache.Context().GetVEdge(in.Uid)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_NotExisted)
		return nil
	}

	out.Status = outLog(path, out)
	return nil
}
