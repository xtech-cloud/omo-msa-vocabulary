package nosql

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy"
	"time"
)

type VEdge struct {
	UID         primitive.ObjectID `bson:"_id"`
	ID          uint64             `json:"id" bson:"id"`
	CreatedTime time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedTime time.Time          `json:"updatedAt" bson:"updatedAt"`
	DeleteTime  time.Time          `json:"deleteAt" bson:"deleteAt"`
	Created     int64              `json:"created" bson:"created"`
	Updated     int64              `json:"updated" bson:"updated"`
	Deleted     int64              `json:"deleted" bson:"deleted"`
	Creator     string             `json:"creator" bson:"creator"`
	Operator    string             `json:"operator" bson:"operator"`

	Name      string      `json:"name" bson:"name"`
	Direction uint8       `json:"direction" bson:"direction"`
	Weight    uint32      `json:"weight" bson:"weight"`
	Type      uint32      `json:"type" bson:"type"`
	Center    string      `json:"center" bson:"center"`
	Remark    string      `json:"remark" bson:"remark"`
	Catalog   string      `json:"catalog" bson:"catalog"` //关系类型
	Source    string      `json:"source" bson:"source"`   //实体UID或者临时UID
	Target    proxy.VNode `json:"target" bson:"target"`
}

func CreateVEdge(info *VEdge) error {
	_, err := insertOne(TableEdge, info)
	if err != nil {
		return err
	}
	return nil
}

func GetAllVEdges() ([]*VEdge, error) {
	cursor, err1 := findAll(TableEdge, 0)
	if err1 != nil {
		return nil, err1
	}
	var items = make([]*VEdge, 0, 100)
	for cursor.Next(context.Background()) {
		var node = new(VEdge)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetVEdgeNextID() uint64 {
	num, _ := getSequenceNext(TableEdge)
	return num
}

func GetVEdge(uid string) (*VEdge, error) {
	result, err := findOne(TableEdge, uid)
	if err != nil {
		return nil, err
	}
	model := new(VEdge)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetVEdgeCountByEntity(entity string) uint32 {
	filter := bson.M{"center": entity, TimeDeleted: 0}
	count, err := getCountBy(TableEdge, filter)
	if err != nil {
		return 0
	}
	return uint32(count)
}

func GetVEdgesBySource(uid string) ([]*VEdge, error) {
	var items = make([]*VEdge, 0, 20)
	filter := bson.M{"source": uid, TimeDeleted: 0}
	cursor, err1 := findMany(TableEdge, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(VEdge)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetVEdgesByCenter(uid string) ([]*VEdge, error) {
	var items = make([]*VEdge, 0, 20)
	filter := bson.M{"center": uid, TimeDeleted: 0}
	cursor, err1 := findMany(TableEdge, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(VEdge)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func UpdateVEdgeBase(uid, name, remark, relation, operator string, dire uint8, target proxy.VNode) error {
	msg := bson.M{"name": name, "remark": remark, "relation": relation, "target": target, "direction": dire, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableEdge, uid, msg)
	return err
}

func UpdateVEdgeTarget(uid, name, entity, thumb, operator string) error {
	msg := bson.M{"target.name": name, "target.entity": entity, "target.thumb": entity, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableEdge, uid, msg)
	return err
}

func RemoveEdge(uid string, operator string) error {
	_, err := removeOne(TableEdge, uid, operator)
	return err
}
