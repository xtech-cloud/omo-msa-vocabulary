package cache

import "omo.msa.vocabulary/proxy"

type NodeInfo struct {
	/**
	BASE:ID 自动生成
	*/
	ID int64
	/**
	BASE:节点名称
	*/
	Name string
	/**
	BASE:节点标签
	*/
	Labels []string

	/**
	Property:节点封面
	*/
	Cover string

	Type string

	Desc string

	/**
	Property：数据库的实体UID
	*/
	Entity string
}

func (mine *NodeInfo) initInfo(db *proxy.Node) {
	mine.Name = db.Name
	mine.Entity = db.UID
	mine.ID = db.ID
	entity := Context().GetEntity(db.UID)
	if entity != nil {
		mine.Cover = entity.Cover
	}
	mine.Labels = db.Labels
}

func (mine *NodeInfo) UpdateCover(cover string) error {
	mine.Cover = cover
	return nil
}

func (mine *NodeInfo) UpdateName(name string) error {
	return nil
}
