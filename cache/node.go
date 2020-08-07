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

	/**
	Property：数据库的实体UID
	*/
	EntityUID string
}

func (mine *NodeInfo)initInfo(db *proxy.Node)  {
	mine.Name = db.Name
	mine.EntityUID = db.UID
	mine.ID = db.ID
	mine.Labels = db.Labels
}

func (mine *NodeInfo)UpdateCover(cover string) error {
	mine.Cover = cover
	return nil
}

func (mine *NodeInfo)UpdateName(name string) error {
	return nil
}

