package grpc

import (
	"context"
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
	tmp.Remark = info.Remark
	tmp.Begin = info.Begin
	tmp.End = info.End
	tmp.Name = info.Name
	tmp.Type = pb.AttributeType(info.Kind)
	return tmp
}

func (mine *AttributeService)AddOne(ctx context.Context, in *pb.ReqAttributeAdd, out *pb.ReplyAttributeInfo) error {
	path := "attribute.addOne"
	inLog(path, in)
	if cache.HadAttribute(in.Key) {
		out.Status = outError(path, "the key of attribute is repeated", pb.ResultStatus_Repeated)
		return nil
	}
	if cache.HadAttributeByName(in.Name) {
		out.Status = outError(path, "the name of attribute is repeated", pb.ResultStatus_Repeated)
		return nil
	}
	info := new(cache.AttributeInfo)
	info.Name = in.Name
	info.Key = in.Key
	info.Kind = cache.AttributeType(in.Type)
	info.Remark = in.Remark
	info.Begin = in.Begin
	info.End = in.End
	err := cache.CreateAttribute(info)
	if err == nil {
		out.Info = switchAttribute(info)
		out.Status = outLog(path, out)
	}else{
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
	}
	return nil
}

func (mine *AttributeService)GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyAttributeInfo) error {
	path := "attribute.getOne"
	inLog(path, in)
	if len(in.Uid) > 0 {
		info := cache.GetAttribute(in.Uid)
		if info == nil {
			out.Status = outError(path,"not found the attribute by uid", pb.ResultStatus_NotExisted)
			return nil
		}
		out.Info = switchAttribute(info)
		out.Status = outLog(path, out)
	}else if len(in.Key) > 0 {
		info := cache.GetAttributeByKey(in.Key)
		if info == nil {
			out.Status = outError(path,"not found the attribute by key", pb.ResultStatus_NotExisted)
			return nil
		}
		out.Info = switchAttribute(info)
		out.Status = outLog(path, out)
	}else{
		out.Status = outError(path,"param is empty", pb.ResultStatus_Empty)
	}
	return nil
}

func (mine *AttributeService)RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "attribute.removeOne"
	inLog(path, in)
	err := cache.RemoveAttribute(in.Uid, in.Operator)
	if err != nil {
		out.Status = outError(path,err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.Status = outLog(path, out)
	return nil
}

func (mine *AttributeService)All(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyAttributeList) error {
	out.List = make([]*pb.AttributeInfo, 0, 10)
	for _, value := range cache.AllAttributes() {
		out.List = append(out.List, switchAttribute(value))
	}
	out.Status = &pb.ReplyStatus{Code: 0, Msg: ""}
	return nil
}

func (mine *AttributeService)Update(ctx context.Context, in *pb.ReqAttributeUpdate, out *pb.ReplyAttributeInfo) error {
	path := "attribute.update"
	inLog(path, in)
	info := cache.GetAttribute(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the attribute by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateBase(in.Name, in.Remark, in.Begin, in.End, in.Operator)
	if err != nil {
		out.Status = outError(path,err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Info = switchAttribute(info)
	out.Status = outLog(path, out)
	return nil
}

