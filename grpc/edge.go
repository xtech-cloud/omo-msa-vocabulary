package grpc

import (
	"context"
	"fmt"
	pbstaus "github.com/xtech-cloud/omo-msp-status/proto/status"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
	"omo.msa.vocabulary/proxy"
	"strings"
)

type VEdgeService struct{}

func switchVEdge(info *cache.VEdgeInfo) *pb.VEdgeInfo {
	tmp := new(pb.VEdgeInfo)
	tmp.Uid = info.UID
	tmp.Name = info.Name
	tmp.Id = int64(info.ID)
	tmp.Type = uint32(info.Type)
	tmp.Operator = info.Operator
	tmp.Creator = info.Creator
	tmp.Created = info.Created
	tmp.Updated = info.Updated
	tmp.Source = info.Source
	tmp.Target = &pb.VNode{
		Name:   info.Target.Name,
		Uid:    info.Target.UID,
		Entity: info.Target.Entity,
		Thumb:  info.Target.Thumb,
		Desc:   info.Target.Desc,
	}
	tmp.Direction = uint32(info.Direction)
	tmp.Category = info.Relation
	tmp.Weight = info.Weight
	tmp.Center = info.Center
	tmp.Remark = info.Remark
	if tmp.Source == "" {
		tmp.Source = tmp.Center
	}
	return tmp
}

func (mine *VEdgeService) AddOne(ctx context.Context, in *pb.ReqVEdgeAdd, out *pb.ReplyVEdgeInfo) error {
	path := "vedge.addOne"
	inLog(path, in)
	in.Name = strings.TrimSpace(in.Name)

	if len(in.Center) < 1 {
		out.Status = outError(path, "the center is empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	node := proxy.VNode{Entity: in.Target, Name: in.Label, Desc: in.Desc, Thumb: in.Thumb}
	info, err := cache.Context().CreateVEdge(in.Center, in.Source, in.Name, in.Remark, in.Relation, in.Operator, in.Direction, in.Weight, in.Type, node)
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
	if in.Key == "public" {
		entity, _ := cache.Context().GetPublicEntity(in.Parent)
		if entity != nil {
			array = entity.GetPublicEdges()
		}
	} else {
		array = cache.Context().GetVEdgesByCenter(in.Parent)
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
	in.Name = strings.TrimSpace(in.Name)

	info, err := cache.Context().GetVEdge(in.Uid)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_NotExisted)
		return nil
	}
	node := proxy.VNode{UID: info.Target.UID, Name: in.Label, Entity: in.Target, Thumb: in.Thumb, Desc: in.Desc}
	err = info.UpdateBase(in.Name, in.Remark, in.Relation, in.Operator, uint8(in.Direction), node)
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
