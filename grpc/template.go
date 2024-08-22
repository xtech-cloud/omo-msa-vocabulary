package grpc

import (
	"context"
	"fmt"
	pbstaus "github.com/xtech-cloud/omo-msp-status/proto/status"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
	"strings"
)

type TemplateService struct{}


  
func switchTemplate(info *cache.TemplateInfo) *pb.TemplateInfo {
	tmp := new(pb.TemplateInfo)
	tmp.Uid = info.UID
	tmp.Name = info.Name
	tmp.Type = pb.TemplateType(info.Type)
	tmp.SkipRows = int32(info.SkipRows)
	tmp.Columns = info.Columns
	tmp.Url = info.Url
	tmp.Comments = info.Comments
	tmp.Creator = info.Creator
	tmp.Operator = info.Operator
	return tmp
}

func (mine *TemplateService) AddOne(ctx context.Context, in *pb.ReqTemplateAdd, out *pb.ReplyTemplateInfo) error {
	path := "template.addOne"
	inLog(path, in)
	in.Name = strings.TrimSpace(in.Name)

	if cache.Context().HadTemplateByName(in.Name) {
		out.Status = outError(path, "the template name had existed", pbstaus.ResultStatus_Repeated)
		return nil
	}
	info := new(cache.TemplateInfo)
	info.Name = in.Name
	info.Type = cache.TemplateType(in.Type)
	info.SkipRows = in.SkipRows
	info.Columns = in.Columns
	info.Url = in.Url
	info.Comments = in.Comments
	err := cache.Context().CreateTemplate(in.Operator, info)
	if err == nil {
		out.Info = switchTemplate(info)
		out.Status = outLog(path, out)
	} else {
		out.Status = outError(path, "the template name had existed", pbstaus.ResultStatus_DBException)
	}
	return nil
}

func (mine *TemplateService) GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyTemplateInfo) error {
	path := "template.getOne"
	inLog(path, in)
	if len(in.Uid) > 0 {
		info := cache.Context().GetTemplate(in.Uid)
		if info == nil {
			out.Status = outError(path, "not found the template by uid", pbstaus.ResultStatus_NotExisted)
			return nil
		}
		out.Info = switchTemplate(info)
		out.Status = outLog(path, out)
	} else if len(in.Key) > 0 {
		info := cache.Context().GetTemplateByName(in.Key)
		if info == nil {
			out.Status = outError(path, "not found the template by key", pbstaus.ResultStatus_NotExisted)
			return nil
		}
		out.Info = switchTemplate(info)
		out.Status = outLog(path, out)
	} else {
		out.Status = outError(path, "param is empty", pbstaus.ResultStatus_Empty)
	}
	return nil
}

func (mine *TemplateService) RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "template.removeOne"
	inLog(path, in)
	err := cache.Context().RemoveTemplate(in.Uid, in.Operator)
	out.Uid = in.Uid
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Status = outLog(path, out)
	return nil
}

func (mine *TemplateService) GetAll(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyTemplateList) error {
	path := "template.getAll"
	inLog(path, in)
	array := cache.Context().AllTemplates()
	out.List = make([]*pb.TemplateInfo, 0, len(array))
	for _, value := range array {
		out.List = append(out.List, switchTemplate(value))
	}
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *TemplateService) UpdateOne(ctx context.Context, in *pb.ReqTemplateUpdate, out *pb.ReplyTemplateInfo) error {
	path := "template.updateOne"
	inLog(path, in)
	var err error
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the uid is empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	in.Name = strings.TrimSpace(in.Name)

	info := cache.Context().GetTemplate(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the template by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	err = info.UpdateBase(in.Name, uint8(in.Type), in.SkipRows, in.Columns, in.Url, in.Comments, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Info = switchTemplate(info)
	out.Status = outLog(path, out)
	return nil
}
