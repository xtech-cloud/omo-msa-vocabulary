package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
)

type graphSample struct {
	Parent string `json:"parent"`
	Name string `json:"name"`
	Center string `json:"center"`
	Nodes []*nodeSample `json:"nodes"`
	Edges []*edgeSample `json:"edges"`
}

type nodeSample struct {
	UID string `json:"uid"`
	Type string `json:"type"`
	Name string `json:"name"`
	Avatar string `json:"avatar"`
}

type edgeSample struct {
	UID string `json:"uid"`
	Direction uint8 `json:"direction"`
	Name string `json:"name"`
	From string `json:"from"`
	To string `json:"to"`
}

func SwitchGraph(info *GraphInfo, parent string) *graphSample {
	sample := new(graphSample)
	sample.Parent = parent
	sample.Center = info.Center()
	sample.Name = info.GetNode(info.center).Name
	nodes := info.Nodes()
	sample.Nodes = make([]*nodeSample, 0, len(nodes))
	for i := 0;i < len(nodes);i += 1 {
		sample.Nodes = append(sample.Nodes, switchNode(nodes[i]))
	}
	links := info.Links()
	sample.Edges = make([]*edgeSample, 0, len(links))
	for i := 0;i < len(links);i += 1 {
		edge := new(edgeSample)
		edge.UID = strconv.FormatInt(links[i].ID,10)
		edge.Name = links[i].Name
		edge.From = links[i].From
		edge.To = links[i].To
		edge.Direction = uint8(links[i].Direction)
		sample.Edges = append(sample.Edges, edge)
	}
	return sample
}

func switchNode(info *NodeInfo) *nodeSample {
	if info == nil {
		return nil
	}
	sample := new (nodeSample)
	sample.Name = info.Name
	sample.UID = info.EntityUID
	tmp := Context().GetEntity(info.EntityUID)
	if tmp != nil {
		sample.Type = switchEntityLabel(tmp.Concept)
		sample.Avatar = tmp.Cover
	}

	return sample
}

func generateGraphJson(uid string, path string)  {
	//if !cacheCtx.graph.HadLinkNode(uid) {
	//	return
	//}
	graph,_ := cacheCtx.graph.GetSubGraph(uid)
	if graph == nil || len(graph.nodes) < 1 {
		return
	}
	tmp := SwitchGraph(graph, "")
	bytes,err := json.Marshal(tmp)
	if err != nil {
		fmt.Println("json marshal error that "+ err.Error())
	}else{
		err1 := ioutil.WriteFile(path + "graph_"+uid+".json", bytes, 0666)
		if err1 != nil {
			fmt.Println("write graph json error that "+ err1.Error())
		}
	}
}
