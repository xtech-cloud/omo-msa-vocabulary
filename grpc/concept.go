package grpc

import (
	"context"
	"errors"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
)

type ConceptService struct {}

func switchConcept(info *cache.ConceptInfo) *pb.ConceptInfo {
	tmp := new(pb.ConceptInfo)
	tmp.Uid = info.UID
	tmp.Updated = info.UpdateTime.Unix()
	tmp.Created = info.CreateTime.Unix()
	tmp.Type = pb.ConceptType(info.Type)
	tmp.Name = info.Name
	tmp.Remark = info.Remark
	tmp.Table = info.Table
	tmp.Cover = info.Cover

	length := len(info.Children())
	if length > 0 {
		tmp.Children = make([]*pb.ConceptInfo, 0, length)
		for _, value := range info.Children() {
			tmp.Children = append(tmp.Children, switchConcept(value))
		}
	}else {
		tmp.Children = make([]*pb.ConceptInfo, 0, 1)
	}

	return tmp
}

func (mine *ConceptService)AddOne(ctx context.Context, in *pb.ReqConceptAdd, out *pb.ReplyConceptInfo) error {
	inLog("concept.add", in)
	if len(in.Parent) > 0 {
		parent := cache.GetConcept(in.Parent)
		if parent == nil {
			out.ErrorCode = pb.ResultStatus_Repeated
			return errors.New("not found the table by parent")
		}

		info := new(cache.ConceptInfo)
		info.Remark = in.Remark
		info.Name = in.Name
		info.Type = cache.ConceptType(in.Type)
		info.Cover = in.Cover
		err := parent.CreateChild(info)
		if err != nil {
			out.ErrorCode = pb.ResultStatus_DBException
			return err
		}else{
			out.Info = switchConcept(info)
		}
	}else{
		if len(in.Table) > 0 && cache.HadTopConceptByTable(in.Table) {
			out.ErrorCode = pb.ResultStatus_Repeated
			return errors.New("the table name is repeated")
		}

		info := new(cache.ConceptInfo)
		info.Table = in.Table
		info.Remark = in.Remark
		info.Name = in.Name
		info.Type = cache.ConceptType(in.Type)
		info.Cover = in.Cover
		err := cache.CreateTopConcept(info)
		if err == nil {
			out.Info = switchConcept(info)
		}else{
			out.ErrorCode = pb.ResultStatus_DBException
		}
	}
	return nil
}

func (mine *ConceptService)GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyConceptInfo) error {
	info := cache.GetConcept(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the concept")
	}
	out.Info = switchConcept(info)
	return nil
}

func (mine *ConceptService)RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	inLog("concept.remove", in)
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the concept uid is empty")
	}

	err := cache.RemoveConcept(in.Uid, in.Operator)
	if err == nil {
		out.Uid = in.Uid
	}else {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *ConceptService)GetAll(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyConceptList) error {
	all := cache.GetTopConcepts()
	out.List = make([]*pb.ConceptInfo, 0, len(all))
	for _, value := range all {
		out.List = append(out.List, switchConcept(value))
	}

	return nil
}

func (mine *ConceptService)Update(ctx context.Context, in *pb.ReqConceptUpdate, out *pb.ReplyConceptInfo) error {
	inLog("concept.update", in)
	info := cache.GetConcept(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the concept")
	}
	err := info.UpdateBase(in.Name, in.Remark, in.Operator)
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
		return err
	}
	out.Info = switchConcept(info)
	return nil
}

func (mine *ConceptService)AppendAttribute(ctx context.Context, in *pb.ReqConceptAttribute, out *pb.ReplyConceptAttribute) error {
	info := cache.GetConcept(in.Concept)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the concept by uid when append attribute")
	}
	attr := cache.GetAttribute(in.Uid)
	if attr == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the attribute by uid")
	}
	err := info.AppendAttribute(attr)
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *ConceptService)RemoveAttribute(ctx context.Context, in *pb.ReqConceptAttribute, out *pb.ReplyConceptAttribute) error {
	info := cache.GetConcept(in.Concept)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the concept by uid when append attribute")
	}
	attr := cache.GetAttribute(in.Uid)
	if attr == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the attribute by uid")
	}
	err := info.RemoveAttribute(attr.UID)
	out.Uid = in.Uid
	out.Concept = in.Concept
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}
