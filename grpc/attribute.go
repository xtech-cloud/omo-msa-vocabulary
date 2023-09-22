package grpc

import (
	"context"
	"fmt"
	pbstaus "github.com/xtech-cloud/omo-msp-status/proto/status"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
	"strconv"
	"strings"
)

type AttributeService struct{}

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
	tmp.Type = uint32(info.Kind)
	return tmp
}

func (mine *AttributeService) AddOne(ctx context.Context, in *pb.ReqAttributeAdd, out *pb.ReplyAttributeInfo) error {
	path := "attribute.addOne"
	inLog(path, in)
	in.Name = strings.TrimSpace(in.Name)
	if cache.Context().HadAttributeByName(in.Name) {
		out.Status = outError(path, "the name of attribute is repeated", pbstaus.ResultStatus_Repeated)
		return nil
	}
	key := strings.ToLower(in.Key)
	if len(in.Key) > 1 && cache.Context().HadAttributeByKey(key) {
		out.Status = outError(path, "the key of attribute is repeated", pbstaus.ResultStatus_Empty)
		return nil
	}
	info := new(cache.AttributeInfo)
	info.Name = in.Name
	info.Key = key
	info.Kind = cache.AttributeType(in.Type)
	info.Remark = in.Remark
	info.Begin = in.Begin
	info.End = in.End
	info.Creator = in.Operator
	err := cache.Context().CreateAttribute(info)
	if err == nil {
		out.Info = switchAttribute(info)
		out.Status = outLog(path, out)
	} else {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
	}
	return nil
}

func (mine *AttributeService) GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyAttributeInfo) error {
	path := "attribute.getOne"
	inLog(path, in)
	if len(in.Uid) > 0 {
		info := cache.Context().GetAttribute(in.Uid)
		if info == nil {
			out.Status = outError(path, "not found the attribute by uid", pbstaus.ResultStatus_NotExisted)
			return nil
		}
		out.Info = switchAttribute(info)
		out.Status = outLog(path, out)
	} else if len(in.Key) > 0 {
		info := cache.Context().GetAttributeByKey(in.Key)
		if info == nil {
			out.Status = outError(path, "not found the attribute by key", pbstaus.ResultStatus_NotExisted)
			return nil
		}
		out.Info = switchAttribute(info)
		out.Status = outLog(path, out)
	} else {
		out.Status = outError(path, "param is empty", pbstaus.ResultStatus_Empty)
	}
	return nil
}

func (mine *AttributeService) GetStatistic(ctx context.Context, in *pb.RequestFilter, out *pb.ReplyStatistic) error {
	path := "attribute.getStatistic"
	inLog(path, in)
	out.Status = outError(path, "param is empty", pbstaus.ResultStatus_Empty)
	return nil
}

func (mine *AttributeService) RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "attribute.removeOne"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the attribute uid is empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	num := cache.Context().GetEntitiesCountByAttribute(in.Uid)
	if num > 0 {
		out.Status = outError(path, "the attribute have entities used that num = "+strconv.Itoa(num), pbstaus.ResultStatus_Prohibition)
		return nil
	}
	err := cache.Context().RemoveAttribute(in.Uid, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.Status = outLog(path, out)
	return nil
}

func (mine *AttributeService) GetAll(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyAttributeList) error {
	path := "attribute.getAll"
	inLog(path, in)
	out.List = make([]*pb.AttributeInfo, 0, 10)
	for _, value := range cache.Context().AllAttributes() {
		out.List = append(out.List, switchAttribute(value))
	}
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *AttributeService) Update(ctx context.Context, in *pb.ReqAttributeUpdate, out *pb.ReplyAttributeInfo) error {
	path := "attribute.update"
	inLog(path, in)
	info := cache.Context().GetAttribute(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the attribute by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	in.Name = strings.TrimSpace(in.Name)
	err := info.UpdateBase(in.Name, in.Remark, in.Begin, in.End, in.Operator, uint8(in.Type))
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	key := strings.ToLower(in.Key)
	if in.Key != "" && info.Key != key {
		if !cache.Context().HadAttributeByKey(in.Key) {
			_ = info.UpdateKey(key, in.Operator)
		}
	}
	out.Info = switchAttribute(info)
	out.Status = outLog(path, out)
	return nil
}
