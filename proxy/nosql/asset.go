package nosql

import (
	"context"
	"errors"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io/ioutil"
	"mime/multipart"
	"time"
)

type Asset struct {
	UID         primitive.ObjectID `bson:"_id"`
	ID          uint64             `json:"id" bson:"id"`
	Name        string             `json:"name" bson:"name"`
	CreatedTime time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedTime time.Time          `json:"updatedAt" bson:"updatedAt"`
	DeleteTime  time.Time          `json:"deleteAt" bson:"deleteAt"`
	Type        uint8              `json:"type" bson:"type"`
	Size        uint64             `json:"size" bson:"size"`
	Language    string             `json:"language" bson:"language"`
	Version     string             `json:"version" bson:"version"`
	Format      string             `json:"format" bson:"format"`
	MD5         string             `json:"md5" bson:"md5"`
	Owner       string             `json:"owner" bson:"owner"`
	File        string             `json:"file_uid" bson:"file_uid"`
}

func CreateAsset(info *Asset) (error, string) {
	_, err := insertOne(TableAsset, &info)
	if err != nil {
		return err, ""
	}
	return nil, info.UID.Hex()
}

func GetAssetNextID() uint64 {
	num, _ := getSequenceNext(TableAsset)
	return num
}

func GetAssetFile(uid string) (*FileInfo, error) {
	if len(uid) < 2 {
		return nil, errors.New("db asset.files uid is empty ")
	}
	result, err := findOne(TableAsset, uid)
	if err != nil {
		return nil, err
	}
	info := new(FileInfo)
	err1 := result.Decode(&info)
	return info, err1
}

func RemoveAsset(uid string) error {
	if len(uid) < 2 {
		return errors.New("db asset uid is empty ")
	}
	_, err := removeOne(TableAsset, uid)
	return err
}

func DeleteAssetFile(file string) bool {
	if len(file) < 1 {
		return false
	}
	return false
}

func GetAsset(uid string) (*Asset, error) {
	if len(uid) < 2 {
		return nil, errors.New("db asset uid is empty of GetAsset")
	}

	result, err := findOne(TableAsset, uid)
	if err != nil {
		return nil, err
	}
	model := new(Asset)
	err1 := result.Decode(&model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetAssetsByOwner(owner string) ([]*Asset, error) {
	var items = make([]*Asset, 0, 20)
	def := new(time.Time)
	filter := bson.M{"owner": owner, "deleteAt": def}
	cursor, err1 := findMany(TableAsset, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	for cursor.Next(context.Background()) {
		var node = new(Asset)
		if err := cursor.Decode(&node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func UpdateAssetLanguage(uid string, language string) error {
	msg := bson.M{"language": language, "updatedAt": time.Now()}
	_, err := updateOne(TableAsset, uid, msg)
	return err
}

func CreateAssetInfoFile(from multipart.File, filename string) (*FileInfo, error) {
	var info = new(FileInfo)
	_, err := ioutil.ReadAll(from)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func dropAssetTable() error {
	return dropOne(TableAsset)
}

func parseAssets(data []gjson.Result) bool {
	assets := make([]*Asset, 0)
	for _, value := range data {
		var asset = new(Asset)
		for key, value := range value.Map() {
			switch key {
			case "_id":
				asset.UID, _ = primitive.ObjectIDFromHex(value.String())
			case "createdAt":
				asset.CreatedTime = value.Time()
			case "updatedAt":
				asset.UpdatedTime = value.Time()
			case "deleteAt":
				asset.DeleteTime = value.Time()
			case "name":
				asset.Name = value.String()
			case "size":
				asset.Size = value.Uint()
			case "version":
				asset.Version = value.String()
			case "format":
				asset.Format = value.String()
			case "md5":
				asset.MD5 = value.String()
			case "file_uid":
				asset.File = value.String()
			}
		}
		assets = append(assets, asset)
	}
	if tableExist(TableAsset) {
		err := dropAssetTable()
		if err != nil {
			return false
		}
	}
	for i := 0; i < len(assets); i++ {
		_, _ = CreateAsset(assets[i])
	}
	return true
}
