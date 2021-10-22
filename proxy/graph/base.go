package graph

import (
	"errors"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"omo.msa.vocabulary/config"
)

type Neo4JContext struct {
	driver  neo4j.Driver
	session neo4j.Session
}

var neo4jCtx *Neo4JContext

func InitNeo4J(config *config.GraphConfig) error {
	neo4jCtx = new(Neo4JContext)
	// 创建neo4j驱动
	url := "bolt://" + config.IP + ":" + config.Port
	var err error
	configForNeo4j40 := func(conf *neo4j.Config) { conf.Encrypted = false }
	neo4jCtx.driver, err = neo4j.NewDriver(url, neo4j.BasicAuth(config.User, config.Password, ""), configForNeo4j40)
	if err != nil {
		return err
	}

	// 获取neo4j session
	neo4jCtx.session, err = neo4jCtx.driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	return nil
}

func CreateNode(name, label, uid string) (neo4j.Node, error) {
	if neo4jCtx.session == nil {
		return nil, errors.New("the graph session is nil that init first")
	}
	cypher := fmt.Sprintf("CREATE (n:%s{name:$name, uid: $uid}) RETURN n", label)
	result, err := neo4jCtx.session.Run(cypher, map[string]interface{}{"name": name, "uid": uid})
	if err != nil {
		return nil, err
	}
	//fmt.Println("CreateNode..."+ cypher)
	for result.Next() {
		node, ok := result.Record().GetByIndex(0).(neo4j.Node)
		if ok {
			return node, nil
		} else {
			return nil, errors.New("node create failed by unknown error")
		}
	}
	return nil, result.Err()
}

func CreateNodeLabel(id int64, label string) (neo4j.Node, error) {
	if neo4jCtx.session == nil {
		return nil, errors.New("the graph session is nil that init first")
	}
	cypher := fmt.Sprintf("CREATE (a:%s) WHERE id(a) = %d RETURN a", label, id)
	result, err := neo4jCtx.session.Run(cypher, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	//fmt.Println("GetNode..."+ cypher)
	for result.Next() {
		node, ok := result.Record().GetByIndex(0).(neo4j.Node)
		if ok {
			return node, nil
		} else {
			return nil, errors.New("node get failed by unknown error")
		}
	}
	return nil, result.Err()
}

func GetNode(uid string) (neo4j.Node, error) {
	if neo4jCtx.session == nil {
		return nil, errors.New("the graph session is nil that init first")
	}
	cypher := fmt.Sprintf("MATCH (a) WHERE a.uid = '%s' RETURN a", uid)
	result, err := neo4jCtx.session.Run(cypher, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	//fmt.Println("GetNode..."+ cypher)
	for result.Next() {
		node, ok := result.Record().GetByIndex(0).(neo4j.Node)
		if ok {
			return node, nil
		} else {
			return nil, errors.New("node get failed by unknown error")
		}
	}
	return nil, result.Err()
}

func GetNodeByID(id int64) (neo4j.Node, error) {
	if neo4jCtx.session == nil {
		return nil, errors.New("the graph session is nil that init first")
	}
	cypher := fmt.Sprintf("MATCH (a) WHERE id(a)=%d RETURN a", id)
	result, err := neo4jCtx.session.Run(cypher, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	//fmt.Println("GetNodeByID..."+ cypher)
	for result.Next() {
		node, ok := result.Record().GetByIndex(0).(neo4j.Node)
		if ok {
			return node, nil
		} else {
			return nil, errors.New("node get failed by unknown error")
		}
	}
	return nil, result.Err()
}

func CreateLink(from, to int64, kind, name, relation string, direction uint8, weight uint32) (neo4j.Relationship, error) {
	if neo4jCtx.session == nil {
		return nil, errors.New("the graph session is nil that init first")
	}
	if len(kind) < 1 {
		return nil, errors.New("the kind is empty")
	}
	cypher := fmt.Sprintf("MATCH (a),(b) WHERE id(a)=%d AND id(b)=%d CREATE (a)-[r:%s{name:$name, "+
		"direction:$direction, relation:$relation}]->(b) RETURN r", from, to, kind)
	result, err := neo4jCtx.session.Run(cypher, map[string]interface{}{"name": name, "direction": direction, "relation": relation, "weight": weight})
	if err != nil {
		return nil, err
	}
	//fmt.Println("CreateLink..."+ cypher)
	for result.Next() {
		link, ok := result.Record().GetByIndex(0).(neo4j.Relationship)
		if ok {
			return link, nil
		} else {
			return nil, errors.New("link create failed by unknown error")
		}
	}
	return nil, result.Err()
}

func GetLinkByID(id int64) (neo4j.Relationship, error) {
	if neo4jCtx.session == nil {
		return nil, errors.New("the graph session is nil that init first")
	}
	cypher := fmt.Sprintf("MATCH (a)-[r]-(b) WHERE id(r)=%d RETURN r", id)
	result, err := neo4jCtx.session.Run(cypher, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	//fmt.Println("GetLink..."+ cypher)
	for result.Next() {
		link, ok := result.Record().GetByIndex(0).(neo4j.Relationship)
		if ok {
			return link, nil
		} else {
			return nil, errors.New("link create failed by unknown error")
		}
	}
	return nil, result.Err()
}

func GetLink(from, to string) (neo4j.Relationship, error) {
	if neo4jCtx.session == nil {
		return nil, errors.New("the graph session is nil that init first")
	}
	cypher := fmt.Sprintf("MATCH (a{uid:'%s'})-[r]-(b{uid:'%s'}) RETURN r", from, to)
	result, err := neo4jCtx.session.Run(cypher, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	//fmt.Println("GetLink..."+ cypher)
	for result.Next() {
		link, ok := result.Record().GetByIndex(0).(neo4j.Relationship)
		if ok {
			return link, nil
		} else {
			return nil, errors.New("link create failed by unknown error")
		}
	}
	return nil, result.Err()
}

func GetLinks(from, to string) ([]neo4j.Relationship, error) {
	if neo4jCtx.session == nil {
		return nil, errors.New("the graph session is nil that init first")
	}
	cypher := fmt.Sprintf("MATCH (a{uid:%s})-[r]-(b{uid:%s}) RETURN r", from, to)
	result, err := neo4jCtx.session.Run(cypher, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	fmt.Println("GetLinks..." + cypher)
	array := make([]neo4j.Relationship, 0, 3)
	for result.Next() {

		links := result.Record().Values()
		for _, value := range links {
			link := value.(neo4j.Relationship)
			array = append(array, link)
		}
	}
	return array, result.Err()
}

func DeleteNode(id int64, label string) error {
	if neo4jCtx.session == nil {
		return errors.New("the graph session is nil that init first")
	}
	cypher := fmt.Sprintf("MATCH (n:%s) WHERE id(n)=%d DETACH DELETE n RETURN n", label, id)
	result, err := neo4jCtx.session.Run(cypher, map[string]interface{}{})
	if err != nil {
		return err
	}
	for result.Next() {
		return result.Err()
	}
	return result.Err()
}

func DeleteLink(id int64) error {
	if neo4jCtx.session == nil {
		return errors.New("the graph session is nil that init first")
	}
	cypher := fmt.Sprintf("MATCH (a)-[r]-(b) WHERE id(r)=%d DELETE r", id)
	result, err := neo4jCtx.session.Run(cypher, map[string]interface{}{})
	if err != nil {
		return err
	}
	for result.Next() {
		return result.Err()
	}
	return result.Err()
}

func FindPath(from, to string) ([]neo4j.Node, []neo4j.Relationship, error) {
	if neo4jCtx.session == nil {
		return nil, nil, errors.New("the graph session is nil that init first")
	}
	cypher := fmt.Sprintf("MATCH p=shortestPath((a{uid:'%s'})-[*..6]-(b{uid:'%s'})) RETURN p", from, to)
	result, err := neo4jCtx.session.Run(cypher, map[string]interface{}{})
	if err != nil {
		return nil, nil, err
	}
	//fmt.Println("FindPath..."+cypher)
	nodes := make([]neo4j.Node, 0, 5)
	links := make([]neo4j.Relationship, 0, 5)
	for result.Next() {
		path, ok := result.Record().GetByIndex(0).(neo4j.Path)
		if ok {
			for _, node := range path.Nodes() {
				nodes = append(nodes, node)
			}

			for _, link := range path.Relationships() {
				links = append(links, link)
			}
		}
	}
	return nodes, links, result.Err()
}

func FindGraph(uid, label string) ([]neo4j.Node, []neo4j.Relationship, error) {
	if neo4jCtx.session == nil {
		return nil, nil, errors.New("the graph session is nil that init first")
	}
	cypher := fmt.Sprintf("MATCH (a:%s{uid:'%s'})-[r]-(b) RETURN a,r,b", label, uid)
	result, err := neo4jCtx.session.Run(cypher, map[string]interface{}{})
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("FindGraph..." + cypher)
	nodes := make([]neo4j.Node, 0, 5)
	links := make([]neo4j.Relationship, 0, 5)
	for result.Next() {
		if from, ok := result.Record().GetByIndex(0).(neo4j.Node); ok {
			nodes = append(nodes, from)
		}
		if to, ok := result.Record().GetByIndex(2).(neo4j.Node); ok {
			nodes = append(nodes, to)
		}
		if link, ok := result.Record().GetByIndex(1).(neo4j.Relationship); ok {
			links = append(links, link)
		}
	}
	return nodes, links, result.Err()
}
