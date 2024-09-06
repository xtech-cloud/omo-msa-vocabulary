package nosql

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

/**
审核
*/
type Examine struct {
	UID      primitive.ObjectID `bson:"_id"`
	ID       uint64             `json:"id" bson:"id"`
	Created  int64              `json:"created" bson:"created"`
	Updated  int64              `json:"updated" bson:"updated"`
	Deleted  int64              `json:"deleted" bson:"deleted"`
	Creator  string             `json:"creator" bson:"creator"`
	Operator string             `json:"operator" bson:"operator"`

	Kind   uint8  `json:"type" bson:"type"`
	Key    string `json:"key" bson:"key"`
	Target string `json:"target" bson:"target"`
	Value  string `json:"value" bson:"value"`
	Status uint8  `json:"status" bson:"status"`
}

func CreateExamine(info *Examine) error {
	_, err := insertOne(TableExamine, info)
	if err != nil {
		return err
	}
	return nil
}

func GetExamineNextID() uint64 {
	num, _ := getSequenceNext(TableExamine)
	return num
}

func GetExamine(uid string) (*Examine, error) {
	result, err := findOne(TableExamine, uid)
	if err != nil {
		return nil, err
	}
	model := new(Examine)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetExamineBy(target, key string, st, tp uint8) (*Examine, error) {
	filter := bson.M{"target": target, "key": key, "status": st, "type": tp, TimeDeleted: 0}
	result, err := findOneBy(TableExamine, filter)
	if err != nil {
		return nil, err
	}
	model := new(Examine)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetExaminesByTarget(target string) ([]*Examine, error) {
	var items = make([]*Examine, 0, 20)
	filter := bson.M{"target": target, TimeDeleted: 0}
	cursor, err1 := findMany(TableExamine, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Examine)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetExamineCountByStatus(target string, st uint8) uint32 {
	msg := bson.M{"target": target, "status": st, TimeDeleted: 0}
	num, err := getCountBy(TableExamine, msg)
	if err != nil {
		return 0
	}
	return uint32(num)
}

func GetExamineCountByType(target string, tp, st uint8) uint32 {
	msg := bson.M{"target": target, "type": tp, "status": st, TimeDeleted: 0}
	num, err := getCountBy(TableExamine, msg)
	if err != nil {
		return 0
	}
	return uint32(num)
}

func GetExaminesByStatus(target string, st uint8) ([]*Examine, error) {
	var items = make([]*Examine, 0, 20)
	filter := bson.M{"target": target, "status": st, TimeDeleted: 0}
	cursor, err1 := findMany(TableExamine, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Examine)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetExaminesByType(target string, tp, st uint8) ([]*Examine, error) {
	var items = make([]*Examine, 0, 20)
	msg := bson.M{"target": target, "type": tp, "status": st, TimeDeleted: 0}
	cursor, err1 := findMany(TableExamine, msg, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Examine)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func RemoveExamine(uid, operator string) error {
	_, err := removeOne(TableExamine, uid, operator)
	return err
}

func UpdateExamineStatus(uid, operator string, st uint8) error {
	msg := bson.M{"status": st, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableExamine, uid, msg)
	return err
}

func UpdateExamineValue(uid, val, operator string) error {
	msg := bson.M{"value": val, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableExamine, uid, msg)
	return err
}
