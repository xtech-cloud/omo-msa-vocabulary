package cache

import (
	"errors"
	"github.com/qiniu/api.v7/v7/storage"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/config"
	"omo.msa.vocabulary/proxy/nosql"
	"strings"
	"time"
)

const (
	//视频
	AssetTypeVideo = 1
	//音频
	AssetTypeAudio = 2
	AssetTypeImage = 3
	AssetTypeDBFile = 4
	AssetTypeModd = 5
)

const (
	AssetPlatAll = 1
	AssetPlatWin = 2
	AssetPlatAndroid = 4
	AssetPlatWeb = 8
)

const (
	AssetLanguageEN = "en"
	AssetLanguageCN = "zh"
)

type AssetInfo struct {
	Type uint8  `json:"type"`
	Size uint64 `json:"size"`
	BaseInfo
	Version  string             `json:"version"`
	Language string             `json:"language"`
	Format   string             `json:"format"`
	MD5      string             `json:"md5"`
	File     *nosql.FileInfo `json:"file"`
}

func CreateAsset(owner string, info *AssetInfo) error {
	var db = new(nosql.Asset)
	db.CreatedTime = time.Now()
	db.UID = primitive.NewObjectID()
	db.File = info.File.UID
	db.ID = nosql.GetAssetNextID()
	db.Type = info.Type
	db.Name = info.Name
	db.Version = info.Version
	db.Format = info.Format
	db.MD5 = info.MD5
	db.Language = info.Language
	db.Owner = owner
	db.Size = info.Size

	err, uid := nosql.CreateAsset(db)
	if err != nil {
		return err
	}
	info.UID = uid
	return nil
}

func CheckAssetType(format string) uint8 {
	if format == "mp3" || format == "ogg" {
		return AssetTypeAudio
	} else if format == "mp4" {
		return AssetTypeVideo
	} else if format == "jpg" || format == "png" {
		return AssetTypeImage
	} else if format == "json" {
		return AssetTypeDBFile
	} else {
		return 0
	}
}

func RemoveAsset(uid string, file string) error {
	if len(uid) < 2 {
		return errors.New("the asset uid is empty")
	}
	err := nosql.RemoveAsset(uid)
	if err == nil {
		if config.Schema.Cache.Kind == UploadMongo {
			nosql.DeleteAssetFile(file)
		} else if config.Schema.Cache.Kind == UploadQN {
			err1 := DeleteContentFromCloud(file)
			RefreshCDN(config.Schema.Cache.Domain + "/" + file)
			if err1 != nil {
				return err1
			}
		}
	}
	return err
}

func RemoveAssetByUID(uid string) error {
	asset,err := nosql.GetAsset(uid)
	if err != nil {
		return err
	}
	if asset == nil {
		return errors.New("not found the asset")
	}
	err1 := nosql.RemoveAsset(uid)
	if err1 == nil {
		if config.Schema.Cache.Kind == UploadMongo {
			nosql.DeleteAssetFile(asset.File)
		} else if config.Schema.Cache.Kind == UploadQN {
			err2 := DeleteContentFromCloud(asset.File)
			RefreshCDN(config.Schema.Cache.Domain + "/" + asset.File)
			if err2 != nil {
				return err2
			}
		}
	}
	return err1
}

func (mine *AssetInfo) initAsset(uid string) bool {
	asset,err := nosql.GetAsset(uid)
	if err != nil {
		return false
	}
	return mine.initInfo(asset)
}

func (mine *AssetInfo) initInfo(db *nosql.Asset) bool {
	if db == nil {
		return false
	}
	mine.UID = db.UID.Hex()
	mine.Name = db.Name
	mine.Type = db.Type
	mine.Language = db.Language
	if config.Schema.Cache.Kind == UploadMongo {
		file,err := nosql.GetAssetFile(db.File)
		if err == nil {
			mine.File = file
		}
	} else if config.Schema.Cache.Kind == UploadQN {
		mine.File = new(nosql.FileInfo)
		fileInfo := GetContentFromCloud(db.File)
		if fileInfo != nil {
			mine.File.Type = fileInfo.MimeType
			mine.File.MD5 = fileInfo.Hash
			mine.File.Size = fileInfo.Fsize
		}
	}
	if mine.File != nil {
		mine.File.UID = db.File
	}
	mine.Size = db.Size
	mine.Format = db.Format
	mine.MD5 = db.MD5
	mine.Version = db.Version
	return true
}

func (mine *AssetInfo) UpdateBase(language string) error {
	err := nosql.UpdateAssetLanguage(mine.UID, language)
	if err == nil {
		mine.Language = language
	}
	return err
}

func (mine *AssetInfo) URL() string {
	if strings.Contains(mine.File.UID,"http") {
		return mine.File.UID
	}
	if config.Schema.Cache.Kind == UploadQN {
		return storage.MakePublicURL(config.Schema.Cache.Domain, mine.File.UID)
	} else {
		return mine.UID
	}
}

func (mine *AssetInfo) IsSupport(plat uint16) bool {
	p := mine.Platform()
	if p == AssetPlatAll {
		return true
	}
	if p == plat {
		return true
	}else {
		return false
	}
}

func (mine *AssetInfo) Platform() uint16 {
	if mine.Type == AssetTypeAudio || mine.Type == AssetTypeVideo {
		return AssetPlatAll
	}else {
		return 0
	}
}