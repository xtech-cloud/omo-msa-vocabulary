package grpc

import (
	"context"
	"errors"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
)

type RelationService struct {}

func switchRelation(info *cache.RelationshipInfo) *pb.RelationInfo {
	tmp := new(pb.RelationInfo)
	tmp.Uid = info.UID
	tmp.Name = info.Name
	tmp.Type = info.Key
	tmp.Remark = info.Remark
	tmp.Custom = info.Custom
	tmp.Time = info.CreateTime.Unix()
	children := info.Children()
	num := len(children)
	if num > 0 {
		tmp.Children = make([]*pb.RelationInfo, 0, num)
		for i := 0;i < num;i += 1{
			tmp.Children = append(tmp.Children, switchRelation(children[i]))
		}
	}else{
		tmp.Children = make([]*pb.RelationInfo, 0, 1)
	}
	return tmp
}

func (mine *RelationService)AddOne(ctx context.Context, in *pb.ReqRelationAdd, out *pb.ReplyRelationOne) error {
	inLog("relation.add", in)
	info := new(cache.RelationshipInfo)
	info.Name = in.Name
	info.Remark = in.Remark
	err := cache.CreateRelation(in.Parent, in.Operator, info)
	if err == nil{
		out.Info = switchRelation(info)
	}else{
		out.ErrorCode = pb.ResultStatus_DBException
	}

	return err
}

func (mine *RelationService)GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyRelationOne) error {
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the relation uid is empty")
	}
	info := cache.GetRelation(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the relation by uid")
	}
	out.Info = switchRelation(info)
	return nil
}

func (mine *RelationService)RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	inLog("relation.remove", in)
	var err error
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the relation uid is empty")
	}
	if len(in.Key) > 0 {
		parent := cache.GetRelation(in.Key)
		if parent == nil {
			out.ErrorCode = pb.ResultStatus_NotExisted
			return errors.New("not found the relation by parent")
		}
		err = parent.RemoveChild(in.Uid,in.Operator)
	}else{
		err = cache.RemoveRelation(in.Uid,in.Operator)
	}
	out.Uid = in.Uid
	out.Key = in.Key
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *RelationService)GetAll(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyRelationList) error {
	array := cache.AllRelations()
	out.List = make([]*pb.RelationInfo, 0, len(array))
	for _, value := range array {
		out.List = append(out.List, switchRelation(value))
	}
	return nil
}

func (mine *RelationService)UpdateInfo(ctx context.Context, in *pb.ReqRelationUpdate, out *pb.ReplyRelationOne) error {
	inLog("relation.update", in)
	var err error
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the relation uid is empty")
	}
	info := cache.GetRelation(in.Uid)
	if info == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the relation by uid")
	}
	err = info.UpdateBase(in.Name, in.Remark, in.Operator, in.Custom)
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}
