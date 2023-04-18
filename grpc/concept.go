package grpc

import (
	"context"
	"fmt"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
)

type ConceptService struct{}

func switchConcept(info *cache.ConceptInfo) *pb.ConceptInfo {
	tmp := new(pb.ConceptInfo)
	tmp.Uid = info.UID
	tmp.Updated = info.UpdateTime.Unix()
	tmp.Created = info.CreateTime.Unix()
	tmp.Type = uint32(info.Type)
	tmp.Name = info.Name
	tmp.Remark = info.Remark
	tmp.Table = info.Table
	tmp.Cover = info.Cover
	tmp.Parent = info.Parent
	tmp.Scene = uint32(info.Scene)
	tmp.Attributes = info.Attributes()
	tmp.Privates = info.Privates()
	length := len(info.Children)
	if length > 0 {
		tmp.Children = make([]*pb.ConceptInfo, 0, length)
		for _, value := range info.Children {
			tmp.Children = append(tmp.Children, switchConcept(value))
		}
	} else {
		tmp.Children = make([]*pb.ConceptInfo, 0, 1)
	}

	return tmp
}

func (mine *ConceptService) AddOne(ctx context.Context, in *pb.ReqConceptAdd, out *pb.ReplyConceptInfo) error {
	path := "concept.addOne"
	inLog(path, in)
	if len(in.Parent) > 0 {
		parent := cache.Context().GetConcept(in.Parent)
		if parent == nil {
			out.Status = outError(path, "not found the parent concept", pb.ResultStatus_NotExisted)
			return nil
		}
		if parent.HadChildByName(in.Name) {
			out.Status = outError(path, "the concept child name is repeated", pb.ResultStatus_Repeated)
			return nil
		}

		info := new(cache.ConceptInfo)
		info.Remark = in.Remark
		info.Name = in.Name
		info.Type = uint8(in.Type)
		info.Cover = in.Cover
		info.Scene = uint8(in.Scene)
		info.Creator = in.Operator
		err := parent.CreateChild(info)
		if err != nil {
			out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
			return nil
		} else {
			out.Info = switchConcept(info)
			out.Status = outLog(path, out)
		}
	} else {
		//if len(in.Table) > 0 && cache.Context().HadConceptByTable(in.Table) {
		//	out.Status = outError(path,"the table name is repeated", pb.ResultStatus_Repeated)
		//	return nil
		//}

		if cache.Context().HadConceptByName(in.Name, in.Parent) {
			out.Status = outError(path, "the concept name is repeated", pb.ResultStatus_Repeated)
			return nil
		}

		info := new(cache.ConceptInfo)
		info.Table = in.Table
		info.Remark = in.Remark
		info.Name = in.Name
		info.Type = uint8(in.Type)
		info.Cover = in.Cover
		err := cache.Context().CreateTopConcept(info)
		if err == nil {
			out.Info = switchConcept(info)
			out.Status = outLog(path, out)
		} else {
			out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		}
	}
	return nil
}

func (mine *ConceptService) GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyConceptInfo) error {
	path := "concept.getOne"
	inLog(path, in)
	if len(in.Uid) > 0 {
		info := cache.Context().GetConcept(in.Uid)
		if info == nil {
			out.Status = outError(path, "not found the concept by uid", pb.ResultStatus_NotExisted)
			return nil
		}
		out.Info = switchConcept(info)
		out.Status = outLog(path, out)
	} else if len(in.Key) > 0 {
		info := cache.Context().GetConceptByName(in.Key)
		if info == nil {
			out.Status = outError(path, "not found the concept by key", pb.ResultStatus_NotExisted)
			return nil
		}
		out.Info = switchConcept(info)
		out.Status = outLog(path, out)
	} else {
		out.Status = outError(path, "param is empty", pb.ResultStatus_Empty)
	}
	return nil
}

func (mine *ConceptService) RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "concept.removeOne"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the concept uid is empty", pb.ResultStatus_Empty)
		return nil
	}

	err := cache.Context().RemoveConcept(in.Uid, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.Status = outLog(path, out)
	return nil
}

func (mine *ConceptService) GetAll(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyConceptList) error {
	path := "concept.getAll"
	inLog(path, in)
	all := cache.Context().GetTopConcepts()
	out.List = make([]*pb.ConceptInfo, 0, len(all))
	for _, value := range all {
		out.List = append(out.List, switchConcept(value))
	}
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *ConceptService) GetStatistic(ctx context.Context, in *pb.RequestFilter, out *pb.ReplyStatistic) error {
	path := "concept.getStatistic"
	inLog(path, in)
	out.Status = outError(path, "param is empty", pb.ResultStatus_Empty)
	return nil
}

func (mine *ConceptService) Update(ctx context.Context, in *pb.ReqConceptUpdate, out *pb.ReplyConceptInfo) error {
	path := "concept.update"
	inLog(path, in)
	info := cache.Context().GetConcept(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the concept by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateBase(in.Name, in.Remark, in.Operator, uint8(in.Type), uint8(in.Scene))
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Info = switchConcept(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *ConceptService) UpdateAttributes(ctx context.Context, in *pb.RequestList, out *pb.ReplyConceptAttrs) error {
	path := "concept.updateAttributes"
	inLog(path, in)
	info := cache.Context().GetConcept(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the concept by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	if in.Status == 1 {
		err := info.UpdatePrivates(in.List)
		if err != nil {
			out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
			return nil
		}
		out.Attributes = info.Privates()
	} else {
		err := info.UpdateAttributes(in.List)
		if err != nil {
			out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
			return nil
		}
		out.Attributes = info.Attributes()
	}

	out.Status = outLog(path, out)
	return nil
}
