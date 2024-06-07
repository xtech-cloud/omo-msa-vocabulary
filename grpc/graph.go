package grpc

import (
	"context"
	pbstaus "github.com/xtech-cloud/omo-msp-status/proto/status"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"omo.msa.vocabulary/cache"
)

type GraphService struct{}

func switchNode(info *cache.NodeInfo) *pb.NodeInfo {
	tmp := new(pb.NodeInfo)
	tmp.Entity = info.Entity
	tmp.Name = info.Name
	tmp.Id = info.ID
	tmp.Type = info.Type
	tmp.Desc = info.Desc
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

func (mine *GraphService) AddNode(ctx context.Context, in *pb.ReqNodeAdd, out *pb.ReplyNodeInfo) error {
	path := "graph.addNode"
	inLog(path, in)
	node, err := cache.Context().Graph().CreateNode(-1, in.Name, in.Entity, in.Cover, in.Label, nil)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
		return nil
	}
	out.Info = switchNode(node)
	out.Status = outLog(path, out)
	return nil
}

func (mine *GraphService) AddLink(ctx context.Context, in *pb.ReqLinkAdd, out *pb.ReplyLinkInfo) error {
	path := "graph.addLink"
	inLog(path, in)
	var err error
	from := cache.Context().GetGraphNode(in.From)
	if from == nil {
		out.Status = outError(path, "not found the from node", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	to := cache.Context().GetGraphNode(in.To)
	if to == nil {
		out.Status = outError(path, "not found the to node", pbstaus.ResultStatus_NotExisted)
		return nil
	}

	link, err := cache.Context().CreateLink(from, to, in.Name, in.Relation, cache.DirectionType(in.Direction), in.Weight)
	if err == nil {
		out.Info = switchLink(link)
		out.Status = outLog(path, out)
	} else {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
	}
	return nil
}

func (mine *GraphService) GetNode(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyNodeInfo) error {
	path := "graph.getNode"
	inLog(path, in)
	var node *cache.NodeInfo
	if in.Id > 0 {
		node = cache.Context().Graph().GetNodeByID(int64(in.Id))
	} else if len(in.Uid) > 0 {
		node = cache.Context().Graph().GetNode(in.Uid)
	} else if len(in.Key) > 0 {
		node = cache.Context().Graph().GetNodeByName(in.Key)
	}

	if node == nil {
		out.Status = outError(path, "not found the node", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	out.Info = switchNode(node)
	out.Status = outLog(path, out)
	return nil
}

func (mine *GraphService) GetLink(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyLinkInfo) error {
	path := "graph.getLink"
	inLog(path, in)
	var link *cache.LinkInfo
	if in.Id > 0 {
		link = cache.Context().Graph().GetRelation(int64(in.Id))
	} else if len(in.Uid) > 0 {
		link = cache.Context().Graph().GetRelationByEntity(in.Uid)
	}

	if link == nil {
		out.Status = outError(path, "not found the link", pbstaus.ResultStatus_NotExisted)
		return nil
	}
	out.Info = switchLink(link)
	out.Status = outLog(path, out)
	return nil
}

func (mine *GraphService) GetStatistic(ctx context.Context, in *pb.RequestFilter, out *pb.ReplyStatistic) error {
	path := "graph.getStatistic"
	inLog(path, in)
	out.Status = outError(path, "param is empty", pbstaus.ResultStatus_Empty)
	return nil
}

func (mine *GraphService) RemoveNode(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "graph.removeNode"
	inLog(path, in)
	err := cache.Context().Graph().RemoveNode(int64(in.Id), in.Key)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
	} else {
		out.Status = outLog(path, out)
	}
	return nil
}

func (mine *GraphService) RemoveLink(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyInfo) error {
	path := "graph.removeLink"
	inLog(path, in)
	err := cache.Context().Graph().RemoveLink(int64(in.Id))
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
	} else {
		out.Status = outLog(path, out)
	}
	return nil
}

func (mine *GraphService) FindPath(ctx context.Context, in *pb.ReqGraphPath, out *pb.ReplyGraphInfo) error {
	path := "graph.findPath"
	inLog(path, in)
	if len(in.From) < 1 {
		out.Status = outError(path, "the from node is empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	if len(in.To) < 1 {
		out.Status = outError(path, "the to node is empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	graph, err := cache.Context().Graph().GetPath(in.From, in.To)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
	} else {
		out.Graph = switchGraph(graph)
		out.Status = outLog(path, out)
	}
	return nil
}

func (mine *GraphService) FindGraph(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyGraphInfo) error {
	path := "graph.findGraph"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the node uid is empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	graph, err := cache.Context().Graph().GetGraphByCenter(in.Uid)
	if err != nil {
		out.Status = outError(path, err.Error(), pbstaus.ResultStatus_DBException)
	} else {
		out.Graph = switchGraph(graph)
		out.Status = outLog(path, out)
	}
	return nil
}

func (mine *GraphService) FindNodes(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyGraphInfo) error {
	path := "graph.findNodes"
	inLog(path, in)
	if len(in.Uid) < 1 {
		out.Status = outError(path, "the owner uid is empty", pbstaus.ResultStatus_Empty)
		return nil
	}
	graph := cache.Context().Graph().GetOwnerGraph(in.Uid)
	out.Graph = switchGraph(graph)
	out.Status = outLog(path, out)
	return nil
}
