package grpc

import (
	"context"
	"errors"
	"fmt"
	pbstaus "github.com/xtech-cloud/omo-msp-status/proto/status"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
	"omo.msa.vocabulary/proxy"
	"strings"
)

type BoxService struct{}

func switchBox(info *cache.BoxInfo) *pb.BoxInfo {
	tmp := new(pb.BoxInfo)
	tmp.Uid = info.UID
	tmp.Created = info.Created
	tmp.Updated = info.Updated
	tmp.Name = info.Name
	tmp.Remark = info.Remark
	tmp.Concept = info.Concept
	tmp.Cover = info.Cover
	tmp.Type = uint32(info.Type)
	tmp.Count = 0
	tmp.Owner = info.Owner
	tmp.Keywords = make([]string, 0, len(info.Contents))
	tmp.Contents = make([]*pb.ContentInfo, 0, len(info.Contents))
	for _, content := range info.Contents {
		tmp.Keywords = append(tmp.Keywords, content.Keyword)
		tmp.Contents = append(tmp.Contents, &pb.ContentInfo{Keywords: content.Keyword,
			Name: content.Name, Status: uint32(content.Status), Count: content.Count})
	}
	tmp.Workflow = info.Workflow
	tmp.Users = info.Users
	tmp.Reviewers = info.Reviewers
	//logger.Info(fmt.Sprintf("the keywords length = %d of name = %s", len(tmp.Keywords), tmp.Name))
	return tmp
}

func (mine *BoxService) AddOne(ctx context.Context, in *pb.ReqBoxAdd, out *pb.ReplyBoxInfo) error {
	path := "box.addOne"
	inLog(path, in)
	in.Name = strings.TrimSpace(in.Name)
	if cache.Context().HadBoxByName(in.Name) {
		out.Status = outError(path, "the box name is repeated", pbstaus.ResultStatus_Repeated)
		return nil
	}

	info := new(cache.BoxInfo)
	info.Remark = in.Remark
	info.Name = in.Name
	info.Type = uint8(in.Type)
	info.Cover = in.Cover
	info.Concept = in.Concept
	info.Creator = in.Operator
	info.Workflow = in.Workflow
	info.Owner = in.Owner
	info.Contents = make([]*proxy.ContentInfo, 0, len(in.Keywords))
	for _, key := range in.Keywords {
		info.Contents = append(info.Contents, &proxy.ContentInfo{
			Name: key, Keyword: "", Count: 0, Status: 0,
		})
	}
	err := cache.Context().CreateBox(info)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	} else {
		out.Info = switchBox(info)
		out.Status = outLog(path, out)
	}
	return nil
}

func (mine *BoxService) GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyBoxInfo) error {
	path := "box.getOne"
	inLog(path, in)
	info := cache.Context().GetBox(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the box by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	out.Info = switchBox(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *BoxService) RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "box.removeOne"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the box uid is empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	info := cache.Context().GetBox(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the box by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	if len(info.Contents) > 0 {
		out.Status = outError(path, "the box is not empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	err := cache.Context().RemoveBox(in.Uid, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Uid = in.Uid
	out.Status = outLog(path, out)
	return nil
}

func (mine *BoxService) GetAll(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyBoxList) error {
	path := "box.getAll"
	inLog(path, in)
	all := cache.Context().GetBoxes(in.Key, uint8(in.Id))
	out.List = make([]*pb.BoxInfo, 0, len(all))
	for _, value := range all {
		out.List = append(out.List, switchBox(value))
	}
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *BoxService) GetListByUser(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyBoxList) error {
	path := "box.getListByUser"
	inLog(path, in)
	var list []*cache.BoxInfo
	if in.Id < 1 {
		list = cache.Context().GetBoxesByUser(in.Uid)
	} else {
		list = cache.Context().GetBoxesByReviewer(in.Uid)
	}

	out.List = make([]*pb.BoxInfo, 0, len(list))
	for _, value := range list {
		out.List = append(out.List, switchBox(value))
	}
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *BoxService) GetByFilter(ctx context.Context, in *pb.RequestFilter, out *pb.ReplyBoxList) error {
	path := "box.GetByFilter"
	inLog(path, in)

	var err error
	var list []*cache.BoxInfo
	var max uint32
	var pages uint32
	if in.Key == "concept" {
		if in.Value == "" {
			list = cache.Context().GetBoxesByOwner(in.Value)
		} else {
			list = cache.Context().GetBoxesByConcept(in.Value)
		}
	} else if in.Key == "entities" {
		list = cache.Context().GetBoxesByEntities(in.Values)
	} else if in.Key == "name" {
		list = cache.Context().GetBoxesByName(in.Value)
	} else if in.Key == "pages" {
		max, pages, list = cache.Context().GetBoxPages(uint32(in.Page), uint32(in.Number))
	} else if in.Key == "usable" {
		max, pages, list = cache.Context().GetUsableBoxPages(uint32(in.Page), uint32(in.Number))
	} else {
		err = errors.New("not define the key")
	}
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.List = make([]*pb.BoxInfo, 0, len(list))
	for _, info := range list {
		out.List = append(out.List, switchBox(info))
	}
	out.Total = max
	out.Pages = pages
	out.Status = outLog(path, fmt.Sprintf("the length = %d", len(out.List)))
	return nil
}

func (mine *BoxService) GetStatistic(ctx context.Context, in *pb.RequestFilter, out *pb.ReplyStatistic) error {
	path := "box.getStatistic"
	inLog(path, in)
	out.Status = outError(path, "param is empty", pbstaus.ResultStatus_Empty)
	return nil
}

func (mine *BoxService) UpdateBase(ctx context.Context, in *pb.ReqBoxUpdate, out *pb.ReplyBoxInfo) error {
	path := "box.updateBase"
	inLog(path, in)
	info := cache.Context().GetBox(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the box by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	in.Name = strings.TrimSpace(in.Name)
	err := info.UpdateBase(in.Name, in.Remark, in.Concept, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	if len(in.Keywords) > 0 {
		err = info.UpdateContents(in.Keywords, in.Operator)
		if err != nil {
			out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
			return nil
		}
	}
	out.Info = switchBox(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *BoxService) AppendKeywords(ctx context.Context, in *pb.ReqBoxKeywords, out *pb.ReplyBoxInfo) error {
	path := "box.appendsKeywords"
	inLog(path, in)
	info := cache.Context().GetBox(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the box by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}

	err := info.AppendKeywords(in.Keywords, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Info = switchBox(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *BoxService) SubtractKeywords(ctx context.Context, in *pb.ReqBoxKeywords, out *pb.ReplyBoxInfo) error {
	path := "box.subtractKeywords"
	inLog(path, in)
	info := cache.Context().GetBox(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the box by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}

	err := info.RemoveKeywords(in.Keywords, in.Operator)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Info = switchBox(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *BoxService) AppendUsers(ctx context.Context, in *pb.ReqBoxKeywords, out *pb.ReplyBoxInfo) error {
	path := "box.appendsUsers"
	inLog(path, in)
	info := cache.Context().GetBox(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the box by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}

	err := info.AppendUsers(in.Keywords, in.Operator, in.Reviewer)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Info = switchBox(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *BoxService) SubtractUsers(ctx context.Context, in *pb.ReqBoxKeywords, out *pb.ReplyBoxInfo) error {
	path := "box.subtractUsers"
	inLog(path, in)
	info := cache.Context().GetBox(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the box by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}

	err := info.RemoveUsers(in.Keywords, in.Operator, in.Reviewer)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Info = switchBox(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *BoxService) UpdateUsers(ctx context.Context, in *pb.ReqBoxKeywords, out *pb.ReplyBoxInfo) error {
	path := "box.updateUsers"
	inLog(path, in)
	info := cache.Context().GetBox(in.Uid)
	if info == nil {
		out.Status = outError(path, "not found the box by uid", pbstaus.ResultStatus_NotExisted)
		return nil
	}

	err := info.UpdateUsers(in.Keywords, in.Operator, in.Reviewer)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Info = switchBox(info)
	out.Status = outLog(path, out)
	return nil
}

func (mine *BoxService) UpdateByFilter(ctx context.Context, in *pb.ReqUpdateFilter, out *pb.ReplyInfo) error {
	path := "box.updateByFilter"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the box uid is empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	box := cache.Context().GetBox(in.Uid)
	if box == nil {
		out.Status = outError(path, "not found the box", pbstaus.ResultStatus_Empty)
		return nil
	}
	var err error
	if in.Key == "reviewers" {
		err = box.UpdateUsers(in.Values, in.Operator, true)
	} else if in.Key == "concept" {
		err = box.UpdateConcept(in.Value, in.Operator)
	} else if in.Key == "fill" {
		if len(in.Values) == 2 {
			err = box.FillContent(in.Values[0], in.Values[1], in.Operator)
		} else {
			err = errors.New("the values is limit when fill box")
		}
	} else {
		err = errors.New("not defined the key when update by filter")
	}
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Updated = uint64(box.Updated)
	out.Status = outLog(path, out)
	return nil
}
