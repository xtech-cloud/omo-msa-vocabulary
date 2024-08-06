package grpc

import (
	"context"
	"fmt"
	pbstaus "github.com/xtech-cloud/omo-msp-status/proto/status"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
	"strconv"
)

type ExamineService struct{}

func switchExamine(info *cache.ExamineInfo) *pb.ExamineInfo {
	tmp := new(pb.ExamineInfo)
	tmp.Created = info.Data.Created
	tmp.Updated = info.Data.Updated
	tmp.Key = info.Data.Key
	tmp.Uid = info.UID
	tmp.Id = info.Data.ID
	tmp.Target = info.Data.Target
	tmp.Value = info.Data.Value
	tmp.Status = uint32(info.Data.Status)
	tmp.Type = uint32(info.Data.Kind)
	return tmp
}

func (mine *ExamineService) AddOne(ctx context.Context, in *pb.ReqExamineAdd, out *pb.ReplyExamineInfo) error {
	path := "examine.addOne"
	inLog(path, in)

	info, err := cache.Context().CreateExamine(in.Operator, in.Target, in.Key, in.Value, uint8(in.Type))
	if err == nil {
		out.Info = switchExamine(info)
		out.Status = outLog(path, out)
	} else {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
	}
	return nil
}

func (mine *ExamineService) GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyExamineInfo) error {
	path := "examine.getOne"
	inLog(path, in)
	info := cache.Context().GetExamine(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the examine by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	out.Info = switchExamine(info)
	out.Status = outLog(path, out)

	return nil
}

func (mine *ExamineService) GetStatistic(ctx context.Context, in *pb.RequestFilter, out *pb.ReplyStatistic) error {
	path := "examine.getStatistic"
	inLog(path, in)
	if in.Key == "target" {
		out.Count = cache.Context().GetExamineCountByStatus(in.Parent, cache.ExamineStatusIdle)
	} else if in.Key == "type" {
		tp, er := strconv.Atoi(in.Value)
		if er != nil {
			out.Status = outLog(path, er.Error())
			return nil
		}
		out.Count = cache.Context().GetExamineCountByType(in.Parent, uint8(tp))
	} else if in.Key == "status" {
		tp, er := strconv.Atoi(in.Value)
		if er != nil {
			out.Status = outLog(path, er.Error())
			return nil
		}
		out.Count = cache.Context().GetExamineCountByStatus(in.Parent, uint8(tp))
	}
	out.Status = outError(path, "param is empty", pbstaus.ResultStatus_Empty)
	return nil
}

func (mine *ExamineService) RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "examine.removeOne"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the examine uid is empty", pbstaus.ResultStatus_Empty)
		return nil
	}

	out.Uid = in.Uid
	out.Status = outLog(path, out)
	return nil
}

func (mine *ExamineService) GetListByFilter(ctx context.Context, in *pb.RequestFilter, out *pb.ReplyExamineList) error {
	path := "examine.getListByFilter"
	inLog(path, in)
	var list []*cache.ExamineInfo
	if in.Key == "target" {
		list = cache.Context().GetExaminesByTarget(in.Parent)
	} else if in.Key == "type" {
		tp, er := strconv.Atoi(in.Value)
		if er != nil {
			out.Status = outLog(path, er.Error())
			return nil
		}
		list = cache.Context().GetExaminesByType(in.Parent, cache.ExamineType(tp))
	} else if in.Key == "status" {
		tp, er := strconv.Atoi(in.Value)
		if er != nil {
			out.Status = outLog(path, er.Error())
			return nil
		}
		list = cache.Context().GetExaminesByStatus(in.Parent, uint8(tp))
	} else if in.Key == "scene" {

	}
	out.List = make([]*pb.ExamineInfo, 0, len(list))
	for _, value := range list {
		out.List = append(out.List, switchExamine(value))
	}
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *ExamineService) UpdateByFilter(ctx context.Context, in *pb.ReqUpdateFilter, out *pb.ReplyExamineInfo) error {
	path := "examine.updateByFilter"
	inLog(path, in)
	info := cache.Context().GetExamine(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the examine by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	var err error
	if in.Key == "status" {
		st, er := strconv.Atoi(in.Value)
		if er != nil {
			out.Status = outLog(path, er.Error())
			return nil
		}
		err = info.UpdateStatus(uint8(st), in.Operator)
	}
	if err != nil {
		out.Status = outLog(path, err.Error())
		return nil
	}

	out.Info = switchExamine(info)
	out.Status = outLog(path, out)
	return nil
}
