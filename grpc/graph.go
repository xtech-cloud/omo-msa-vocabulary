package grpc

import (
	"context"
	"errors"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
)

type GraphService struct {}

func switchNode(info *cache.NodeInfo) *pb.NodeInfo {
	tmp := new(pb.NodeInfo)
	tmp.Entity = info.EntityUID
	tmp.Name = info.Name
	tmp.Id = info.ID
	tmp.Labels = info.Labels
	tmp.Cover = info.Cover
	return tmp
}

func switchLink(info *cache.LinkInfo) *pb.LinkInfo {
	tmp := new(pb.LinkInfo)
	tmp.Id = info.ID
	tmp.Name = info.Name
	tmp.Relation = info.Relation
	tmp.Direction = pb.DirectionType(info.Direction)
	tmp.From = info.From
	tmp.To = info.To
	return tmp
}

func switchGraph(info *cache.GraphInfo) *pb.GraphInfo {
	tmp := new(pb.GraphInfo)
	tmp.Center = info.Center()
	tmp.Nodes = make([]*pb.NodeInfo, 0, len(info.Nodes()))
	for _, node := range info.Nodes() {
		tmp.Nodes = append(tmp.Nodes, switchNode(node))
	}
	tmp.Links = make([]*pb.LinkInfo, 0, len(info.Links()))
	for _, link := range info.Links() {
		tmp.Links = append(tmp.Links, switchLink(link))
	}
	return tmp
}

func (mine *GraphService)AddNode(ctx context.Context, in *pb.ReqNodeAdd, out *pb.ReplyNodeOne) error {
	node ,err := cache.Graph().CreateNode(in.Name, in.Entity, in.Cover, in.Label)
	if err == nil {
		out.Info = switchNode(node)
	}else{
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *GraphService)AddLink(ctx context.Context, in *pb.ReqLinkAdd, out *pb.ReplyLinkOne) error {
	var err error
	from := cache.GetGraphNode(in.From)
	if from == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the from node")
	}
	to := cache.GetGraphNode(in.To)
	if to == nil {
		out.ErrorCode = pb.ResultStatus_NotExisted
		return errors.New("not found the to node")
	}

	link,err := cache.CreateLink(from, to, cache.LinkType(in.Key), in.Name,in.Key, cache.DirectionType(in.Direction))
	if err == nil {
		out.Info = switchLink(link)
	}else{
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *GraphService)RemoveNode(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	err := cache.Graph().RemoveNode(int64(in.Id), in.Key)
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *GraphService)RemoveLink(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	err := cache.Graph().RemoveLink(int64(in.Id))
	if err != nil {
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *GraphService)FindPath(ctx context.Context, in *pb.ReqGraphPath, out *pb.ReplyGraphInfo) error {
	if len(in.From) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the node of from is empty")
	}
	if len(in.To) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the node of to is empty")
	}
	graph,err:= cache.Graph().GetPath(in.From, in.To)
	if err == nil{
		out.Graph = switchGraph(graph)
	}else{
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}

func (mine *GraphService)FindGraph(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyGraphInfo) error {
	if len(in.Uid) < 1 {
		out.ErrorCode = pb.ResultStatus_Empty
		return errors.New("the node of uid is empty")
	}
	graph,err:= cache.Graph().GetSubGraph(in.Uid)
	if err == nil{
		out.Graph = switchGraph(graph)
	}else{
		out.ErrorCode = pb.ResultStatus_DBException
	}
	return err
}
