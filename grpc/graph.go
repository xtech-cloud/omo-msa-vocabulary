package grpc

import (
	"context"
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

func (mine *GraphService)AddNode(ctx context.Context, in *pb.ReqNodeAdd, out *pb.ReplyNodeInfo) error {
	path := "graph.addNode"
	inLog(path, in)
	node ,err := cache.Graph().CreateNode(in.Name, in.Entity, in.Cover, in.Label)
	if err != nil {
		out.Status = outError(path,err.Error(), pb.ResultStatus_DBException)
		return nil
	}
	out.Info = switchNode(node)
	out.Status = outLog(path, out)
	return nil
}

func (mine *GraphService)AddLink(ctx context.Context, in *pb.ReqLinkAdd, out *pb.ReplyLinkInfo) error {
	path := "graph.addLink"
	inLog(path, in)
	var err error
	from := cache.GetGraphNode(in.From)
	if from == nil {
		out.Status = outError(path, "not found the from node", pb.ResultStatus_NotExisted)
		return nil
	}
	to := cache.GetGraphNode(in.To)
	if to == nil {
		out.Status = outError(path, "not found the to node", pb.ResultStatus_NotExisted)
		return nil
	}

	link,err := cache.CreateLink(from, to, cache.LinkType(in.Key), in.Name,in.Key, cache.DirectionType(in.Direction))
	if err == nil {
		out.Info = switchLink(link)
		out.Status = outLog(path, out)
	}else{
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
	}
	return nil
}

func (mine *GraphService)RemoveNode(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "graph.removeNode"
	inLog(path, in)
	err := cache.Graph().RemoveNode(int64(in.Id), in.Key)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
	}else{
		out.Status = outLog(path, out)
	}
	return nil
}

func (mine *GraphService)RemoveLink(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "graph.removeLink"
	inLog(path, in)
	err := cache.Graph().RemoveLink(int64(in.Id))
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
	}else{
		out.Status = outLog(path, out)
	}
	return nil
}

func (mine *GraphService)FindPath(ctx context.Context, in *pb.ReqGraphPath, out *pb.ReplyGraphInfo) error {
	path := "graph.findPath"
	inLog(path, in)
	if len(in.From) < 1 {
		out.Status = outError(path,"the from node is empty", pb.ResultStatus_Empty)
		return nil
	}
	if len(in.To) < 1 {
		out.Status = outError(path,"the to node is empty", pb.ResultStatus_Empty)
		return nil
	}
	graph,err:= cache.Graph().GetPath(in.From, in.To)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
	}else{
		out.Graph = switchGraph(graph)
		out.Status = outLog(path, out)
	}
	return nil
}

func (mine *GraphService)FindGraph(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyGraphInfo) error {
	path := "graph.findGraph"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path,"the node uid is empty", pb.ResultStatus_Empty)
		return nil
	}
	graph,err:= cache.Graph().GetSubGraph(in.Uid)
	if err != nil {
		out.Status = outError(path, err.Error(), pb.ResultStatus_DBException)
	}else{
		out.Graph = switchGraph(graph)
		out.Status = outLog(path, out)
	}
	return nil
}
