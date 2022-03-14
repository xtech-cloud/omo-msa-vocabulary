package nosql

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Sequence struct {
	UID         primitive.ObjectID `bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	CreatedTime time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedTime time.Time          `json:"updatedAt" bson:"updatedAt"`
	DeleteTime  time.Time          `json:"deleteAt" bson:"deleteAt"`
	Count       uint64             `json:"count" bson:"count"`
}

func createSequence(name string) error {
	var db = new(Sequence)
	db.Count = 0
	db.UID = primitive.NewObjectID()
	db.Name = name
	db.CreatedTime = time.Now()
	_, err := insertOne(TableSequence, db)
	if err != nil {
		return err
	}
	return nil
}

func getSequenceNext(name string) (uint64, error) {
	num, _ := getSequenceCount(name)
	if num < 1 {
		_ = createSequence(name)
	}
	filter := bson.M{"name": name}
	update := bson.M{"$inc": bson.M{"count": 1}, "$set": bson.M{"updatedAt": time.Now()}}
	_, err := updateOneBy(TableSequence, filter, update)
	if err != nil {
		return 0, err
	}
	return num + 1, nil
}

func getSequenceCount(name string) (uint64, error) {
	filter := bson.M{"name": name}
	result, err := findOneBy(TableSequence, filter)
	if err != nil {
		return 0, err
	}
	model := new(Sequence)
	err1 := result.Decode(model)
	if err1 != nil {
		return 0, err1
	}
	return model.Count, nil
}

func GetAllSequences() ([]*Sequence, error) {
	var items = make([]*Sequence, 0, 100)
	cursor, err1 := findAll(TableSequence, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Sequence)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func UpdateSequenceName(uid, name string) error {
	msg := bson.M{"name": name, "updatedAt": time.Now()}
	_, err := updateOne(TableSequence, uid, msg)
	return err
}


func DeleteSequence(uid string) error {
	_,err := deleteOne(TableSequence, uid)
	return err
}
