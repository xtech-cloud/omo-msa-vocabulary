package grpc

import (
	"context"
	"errors"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
)

type AttributeService struct {}

func switchAttribute(info *cache.AttributeInfo) *pb.AttributeInfo {
	tmp := new(pb.AttributeInfo)
	tmp.Created = info.CreateTime.Unix()
	tmp.Updated = info.UpdateTime.Unix()
	tmp.Key = info.Key
	tmp.Uid = info.UID
	tmp.Id = info.ID
	tmp.Begin = info.Begin
	tmp.End = info.End
	tmp.Name = info.Name
	tmp.Type = pb.AttributeType(info.Kind)
	return tmp
}

func (mine *AttributeService)AddOne(ctx context.Context, in *pb.ReqAttributeAdd, out *pb.ReplyAttributeOne) error {
	inLog("attribute.add", in)
	if cache.HadAttribute(in.Key) {
		out.ErrorCode = pb.ResultStatus_Repeated
		return errors.New("the key of attribute is repeated")
	}
	info := new(cache.AttributeInfo)
	info.Name = in.Name
	info.Key = in.Key
	info.Kind = cache.AttributeType(in.Type)
	info.Remark = ""
	info.Begin = in.Begin
	info.End = in.End
	err := cache.CreateAttribute(info)
	if err == nil {
		out.Info = switchAttribute(info)
	}
	return err
}

func (mine *AttributeService)GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyAttributeOne) error {
	if len(in.Uid) > 0 {
		info := cache.GetAttribute(in.Uid)
		if info == nil {
			out.ErrorCode  = pb.ResultStatus_NotExisted
			return errors.New("not found the attribute by uid")
		}
		out.Info = switchAttribute(info)
	}else if len(in.Key) > 0 {
		info := cache.GetAttributeByKey(in.Key)
		if info == nil {
			out.ErrorCode  = pb.ResultStatus_NotExisted
			return errors.New("nou found the attribute by key")
		}
		out.Info = switchAttribute(info)
	}else{
		out.ErrorCode  = pb.ResultStatus_Empty
		return errors.New("the uid or key all is empty")
	}
	return nil
}

func (mine *AttributeService)RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	inLog("attribute.remove", in)
	err := cache.RemoveAttribute(in.Uid, in.Operator)
	out.Uid = in.Uid
	return err
}

func (mine *AttributeService)All(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyAttributeList) error {
	out.List = make([]*pb.AttributeInfo, 0, 10)
	for _, value := range cache.AllAttributes() {
		out.List = append(out.List, switchAttribute(value))
	}
	return nil
}

func (mine *AttributeService)Update(ctx context.Context, in *pb.ReqAttributeUpdate, out *pb.ReplyAttributeOne) error {
	inLog("attribute.update", in)
	info := cache.GetAttribute(in.Uid)
	if info == nil {
		out.ErrorCode  = pb.ResultStatus_NotExisted
		return errors.New("not found the attribute by uid")
	}
	err := info.UpdateBase(in.Name, in.Remark, in.Begin, in.End, in.Operator)
	if err != nil {
		return err
	}
	out.Info = switchAttribute(info)
	return nil
}

