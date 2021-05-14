package nosql

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Archived struct {
	UID         primitive.ObjectID `bson:"_id"`
	ID          uint64             `json:"id" bson:"id"`
	CreatedTime time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedTime time.Time          `json:"updatedAt" bson:"updatedAt"`
	DeleteTime  time.Time          `json:"deleteAt" bson:"deleteAt"`
	Creator     string                `json:"creator" bson:"creator"`
	Operator    string                `json:"operator" bson:"operator"`

	Concept string 				 `json:"concept" bson:"concept"`
	Name   string                `json:"name" bson:"name"`
	Entity  string                `json:"entity" bson:"entity"`
	Scene   string 				  `json:"scene" bson:"scene"`
	File string                	  `json:"file" bson:"file"`
	MD5 string 					  `json:"md5" bson:"md5"`
}

func CreateArchived(info *Archived) error {
	_, err := insertOne(TableArchived, info)
	if err != nil {
		return err
	}
	return nil
}

func GetArchivedNextID() uint64 {
	num, _ := getSequenceNext(TableArchived)
	return num
}

func GetArchivedByEntity(uid string) (*Archived, error) {
	filter := bson.M{"entity": uid}
	result, err := findOneBy(TableArchived, filter)
	if err != nil {
		return nil, err
	}
	model := new(Archived)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetAllArchived() ([]*Archived, error) {
	var items = make([]*Archived, 0, 20)
	def := new(time.Time)
	filter := bson.M{"deleteAt": def}
	cursor, err1 := findMany(TableArchived, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Archived)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetArchivedItems(concept string) ([]*Archived, error) {
	var items = make([]*Archived, 0, 20)
	filter := bson.M{"concept": concept}
	cursor, err1 := findMany(TableArchived, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Archived)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetArchivedListByScene(scene string) ([]*Archived, error) {
	var items = make([]*Archived, 0, 20)
	filter := bson.M{"scene": scene}
	cursor, err1 := findMany(TableArchived, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Archived)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetArchivedListBy(scene, concept string) ([]*Archived, error) {
	var items = make([]*Archived, 0, 20)
	filter := bson.M{"scene": scene, "concept": concept}
	cursor, err1 := findMany(TableArchived, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Archived)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func UpdateArchivedFile(uid, operator, file, md5  string) error {
	msg := bson.M{"file": file, "md5":md5, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(TableArchived, uid, msg)
	return err
}

func HadArchivedByName(name string) (bool, error) {
	msg := bson.M{"name": name}
	return hadOne(TableArchived, msg)
}

