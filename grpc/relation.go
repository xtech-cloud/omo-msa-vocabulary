package grpc

import (
	"context"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
)

type RelationService struct {}

func switchRelation(info *cache.RelationshipInfo) *pb.RelationInfo {
	tmp := new(pb.RelationInfo)
	tmp.Uid = info.UID
	tmp.Name = info.Name
	tmp.Type = uint32(info.Kind)
	tmp.Remark = info.Remark
	tmp.Custom = info.Custom
	tmp.Parent = info.Parent
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

func (mine *RelationService)AddOne(ctx context.Context, in *pb.ReqRelationAdd, out *pb.ReplyRelationInfo) error {
	path := "relation.addOne"
	inLog(path, in)
	if cache.HadRelationByName(in.Name) {
		out.Status = outError(path,"the relation name had existed", pb.ResultStatus_Repeated)
		return nil
	}
	info := new(cache.RelationshipInfo)
	info.Name = in.Name
	info.Remark = in.Remark
	info.Kind = uint8(in.Type)
	info.Custom = in.Custom
	err := cache.CreateRelation(in.Parent, in.Operator, info)
	if err == nil{
		out.Info = switchRelation(info)
		out.Status = outLog(path, out)
	}else{
		out.Status = outError(path,"the relation name had existed", pb.ResultStatus_DBException)
	}
	return nil
}

func (mine *RelationService)GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyRelationInfo) error {
	path := "relation.getOne"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path,"the uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.GetRelation(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the relation by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	out.Info = switchRelation(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *RelationService)RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "relation.removeOne"
	inLog(path, in)
	var err error
	if len(in.Uid) < 1 {
		out.Status = outError(path,"the uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	if len(in.Key) > 0 {
		parent := cache.GetRelation(in.Key)
		if parent == nil {
			out.Status = outError(path,"not found the relation by parent", pb.ResultStatus_NotExisted)
			return nil
		}
		err = parent.RemoveChild(in.Uid,in.Operator)
	}else{
		err = cache.RemoveRelation(in.Uid,in.Operator)
	}
	out.Uid = in.Uid
	out.Key = in.Key
	if err != nil {
		out.Status = outError(path,err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Status = outLog(path, out)
	return nil
}

func (mine *RelationService)GetAll(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyRelationList) error {
	array := cache.AllRelations()
	out.List = make([]*pb.RelationInfo, 0, len(array))
	for _, value := range array {
		out.List = append(out.List, switchRelation(value))
	}
	out.Status = &pb.ReplyStatus{Code: 0, Msg: ""}
	return nil
}

func (mine *RelationService)UpdateInfo(ctx context.Context, in *pb.ReqRelationUpdate, out *pb.ReplyRelationInfo) error {
	path := "relation.updateInfo"
	inLog(path, in)
	var err error
	if len(in.Uid) < 1 {
		out.Status = outError(path,"the uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	info := cache.GetRelation(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the relation by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err = info.UpdateBase(in.Name, in.Remark, in.Operator, in.Custom, uint8(in.Type))
	if err != nil {
		out.Status = outError(path,err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Info = switchRelation(info)
	out.Status = outLog(path, out)
	return nil
}
