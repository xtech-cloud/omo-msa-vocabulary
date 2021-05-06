package grpc

import (
	"context"
	"fmt"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
)

type BoxService struct {}

func switchBox(info *cache.BoxInfo) *pb.BoxInfo {
	tmp := new(pb.BoxInfo)
	tmp.Uid = info.UID
	tmp.Updated = info.UpdateTime.Unix()
	tmp.Created = info.CreateTime.Unix()
	tmp.Name = info.Name
	tmp.Remark = info.Remark
	tmp.Concept = info.Concept
	tmp.Cover = info.Cover
	tmp.Type = uint32(info.Type)
	tmp.Keywords = info.Keywords
	return tmp
}

func (mine *BoxService)AddOne(ctx context.Context, in *pb.ReqBoxAdd, out *pb.ReplyBoxInfo) error {
	path := "box.addOne"
	inLog(path, in)
	if cache.Context().HadBoxByName(in.Name) {
		out.Status = outError(path,"the box name is repeated", pb.ResultStatus_Repeated)
		return nil
	}

	info := new(cache.BoxInfo)
	info.Remark = in.Remark
	info.Name = in.Name
	info.Type = uint8(in.Type)
	info.Cover = in.Cover
	info.Concept = in.Concept
	info.Creator = in.Operator
	err := cache.Context().CreateBox(info)
	if err != nil {
		out.Status = outError(path,err.Error(), pb.ResultStatus_DBException)
		return nil
	}else{
		out.Info = switchBox(info)
		out.Status = outLog(path, out)
	}
	return nil
}

func (mine *BoxService)GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyBoxInfo) error {
	path := "box.getOne"
	inLog(path, in)
	info := cache.Context().GetBox(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the box by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	out.Info = switchBox(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *BoxService)RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "box.removeOne"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path,"the box uid is empty", pb.ResultStatus_Empty)
		return nil
	}

	err := cache.Context().RemoveBox(in.Uid, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.Status = outLog(path, out)
	return nil
}

func (mine *BoxService)GetAll(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyBoxList) error {
	path := "box.getAll"
	inLog(path, in)
	all := cache.Context().GetBoxes(uint8(in.Id))
	out.List = make([]*pb.BoxInfo, 0, len(all))
	for _, value := range all {
		out.List = append(out.List, switchBox(value))
	}
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *BoxService)Update(ctx context.Context, in *pb.ReqBoxUpdate, out *pb.ReplyBoxInfo) error {
	path := "box.update"
	inLog(path, in)
	info := cache.Context().GetBox(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the box by uid", pb.ResultStatus_NotExisted)
		return nil
	}
	err := info.UpdateBase(in.Name, in.Remark, in.Operator, in.Concept)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Info = switchBox(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *BoxService)Appends(ctx context.Context, in *pb.ReqBoxKeywords, out *pb.ReplyBoxInfo) error {
	path := "box.appendsKeywords"
	inLog(path, in)
	info := cache.Context().GetBox(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the box by uid", pb.ResultStatus_NotExisted)
		return nil
	}

	err := info.AppendKeywords(in.Keywords)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Info = switchBox(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *BoxService)Subtracts(ctx context.Context, in *pb.ReqBoxKeywords, out *pb.ReplyBoxInfo) error {
	path := "box.sub"
	inLog(path, in)
	info := cache.Context().GetBox(in.Uid)
	if info == nil {
		out.Status = outError(path,"not found the box by uid", pb.ResultStatus_NotExisted)
		return nil
	}

	err := info.RemoveKeywords(in.Keywords)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Info = switchBox(info)
	out.Status = outLog(path, out)
	return nil
}

