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
	Creator     string             `json:"creator" bson:"creator"`
	Operator    string             `json:"operator" bson:"operator"`

	Name      string      `json:"name" bson:"name"`
	Direction uint8       `json:"direction" bson:"direction"`
	Weight    uint32      `json:"weight" bson:"weight"`
	Center    string      `json:"center" bson:"center"`
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

func GetVEdgesBySource(uid string) ([]*VEdge, error) {
	var items = make([]*VEdge, 0, 20)
	def := new(time.Time)
	filter := bson.M{"source": uid, "deleteAt": def}
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
	def := new(time.Time)
	filter := bson.M{"center": uid, "deleteAt": def}
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

func UpdateVEdgeBase(uid, name, relation, operator string, dire uint8, target proxy.VNode) error {
	msg := bson.M{"name": name, "relation": relation, "target": target, "direction": dire, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(TableEdge, uid, msg)
	return err
}

func RemoveEdge(uid string, operator string) error {
	_, err := removeOne(TableEdge, uid, operator)
	return err
}
