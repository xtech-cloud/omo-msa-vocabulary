package nosql

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Box struct {
	UID         primitive.ObjectID `bson:"_id"`
	ID          uint64             `json:"id" bson:"id"`
	CreatedTime time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedTime time.Time          `json:"updatedAt" bson:"updatedAt"`
	DeleteTime  time.Time          `json:"deleteAt" bson:"deleteAt"`
	Creator     string                `json:"creator" bson:"creator"`
	Operator    string                `json:"operator" bson:"operator"`

	Name   string                `json:"name" bson:"name"`
	Type   uint8 				 `json:"type" bson:"type"`
	Cover  string                `json:"cover" bson:"cover"`
	Remark string                `json:"remark" bson:"remark"`
	Concept string                `json:"concept" bson:"concept"`
	Workflow string `json:"workflow" bson:"workflow"`
	Keywords  []string `json:"keywords" bson:"keywords"`
}

func CreateBox(info *Box) error {
	_, err := insertOne(TableBox, info)
	if err != nil {
		return err
	}
	return nil
}

func GetBoxNextID() uint64 {
	num, _ := getSequenceNext(TableBox)
	return num
}

func GetBox(uid string) (*Box, error) {
	result, err := findOne(TableBox, uid)
	if err != nil {
		return nil, err
	}
	model := new(Box)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetBoxes() ([]*Box, error) {
	var items = make([]*Box, 0, 20)
	cursor, err1 := findAll(TableBox, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Box)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func HadBoxByName(name string) (bool, error) {
	msg := bson.M{"name": name}
	return hadOne(TableBox, msg)
}

func UpdateBoxBase(uid, name, desc, concept, operator string) error {
	msg := bson.M{"name": name, "remark": desc,"operator":operator,"concept": concept, "updatedAt": time.Now()}
	_, err := updateOne(TableBox, uid, msg)
	return err
}

func UpdateBoxCover(uid string, icon string) error {
	msg := bson.M{"cover": icon, "updatedAt": time.Now()}
	_, err := updateOne(TableBox, uid, msg)
	return err
}

func UpdateBoxKeywords(uid string, list []string) error {
	msg := bson.M{"keywords": list, "updatedAt": time.Now()}
	_, err := updateOne(TableBox, uid, msg)
	return err
}

func RemoveBox(uid, operator string) error {
	_, err := removeOne(TableBox, uid, operator)
	return err
}

func AppendBoxKeywords(uid string, attr []string) error {
	msg := bson.M{"keywords": attr}
	_, err := appendElement(TableBox, uid, msg)
	return err
}

func SubtractBoxKeyword(uid string, attr string) error {
	msg := bson.M{"keywords": attr}
	_, err := removeElement(TableBox, uid, msg)
	return err
}
