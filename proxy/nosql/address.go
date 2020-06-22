package nosql

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Address struct {
	UID         primitive.ObjectID `bson:"_id"`
	ID          uint64             `json:"id" bson:"id"`
	CreatedTime time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedTime time.Time          `json:"updatedAt" bson:"updatedAt"`
	DeleteTime  time.Time          `json:"deleteAt" bson:"deleteAt"`
	Country     string             `json:"country" bson:"country"`
	Province    string             `json:"province" bson:"province"`
	City        string             `json:"city" bson:"city"`
	District    string             `json:"district" bson:"district"`
	Town        string             `json:"town" bson:"town"`
	Village     string             `json:"village" bson:"village"`
	Street      string             `json:"street" bson:"street"`
	Number      string             `json:"number" bson:"number"`
	User        string             `json:"user" bson:"user"`
}

func CreateAddress(info *Address) error {
	_, err := insertOne(TableAddress, info)
	if err != nil {
		return err
	}
	return nil
}

func GetAddressNextID() uint64 {
	num, _ := getSequenceNext(TableAddress)
	return num
}

func GetAddress(uid string) (*Address, error) {
	result, err := findOne(TableAddress, uid)
	if err != nil {
		return nil, err
	}
	model := new(Address)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func HadAddress(longitude float32, latitude float32) (bool, error) {
	msg := bson.M{"longitude": longitude, "latitude": latitude}
	had, err := hadOne(TableAddress, msg)
	if err != nil {
		return false, err
	}
	return had, nil
}

func GetAllAddresses() ([]*Address, error) {
	cursor, err1 := findAll(TableAddress, 0)
	if err1 != nil {
		return nil, err1
	}
	var items = make([]*Address, 0, 200)
	for cursor.Next(context.Background()) {
		var node = new(Address)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func dropAddress() error {
	err := dropOne(TableAddress)
	if err != nil {
		return err
	}
	return nil
}
