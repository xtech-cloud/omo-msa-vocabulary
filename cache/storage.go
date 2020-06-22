package cache

import (
	"bytes"
	"context"
	"errors"
	"github.com/micro/go-micro/v2/logger"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/cdn"
	"github.com/qiniu/api.v7/v7/storage"
	"go.uber.org/zap"
	"io/ioutil"
	"mime/multipart"
	"omo.msa.vocabulary/config"
	"omo.msa.vocabulary/proxy/nosql"
	"omo.msa.vocabulary/tool"
)

const (
	UploadMongo = 0
	UploadQN    = 1
)

func RefreshCDN(url string) bool {
	mac := qbox.NewMac(config.Schema.Cache.AccessKey, config.Schema.Cache.SecretKey)
	cdnManager := cdn.NewCdnManager(mac)

	urlsToRefresh := []string{
		url,
	}
	_, err := cdnManager.RefreshUrls(urlsToRefresh)
	if err != nil {
		logger.Info("cache: refresh cdn failed from qiniu cache!!!", zap.String("url", url))
		return false
	}
	return true
}

func DeleteContentFromCloud(key string) error {
	if len(key) < 1 {
		return errors.New("cache: the key is empty")
	}
	mac := qbox.NewMac(config.Schema.Cache.AccessKey, config.Schema.Cache.SecretKey)
	cfg := storage.Config{
		// 是否使用https域名进行资源管理
		UseHTTPS: false,
	}
	// 指定空间所在的区域，如果不指定将自动探测
	// 如果没有特殊需求，默认不需要指定
	//cfg.Zone=&storage.ZoneHuabei
	bucketManager := storage.NewBucketManager(mac, &cfg)
	err := bucketManager.Delete(config.Schema.Cache.Bucket, key)
	if err != nil {
		return err
	}
	return nil
}

func GetContentFromCloud(key string) *storage.FileInfo {
	mac := qbox.NewMac(config.Schema.Cache.AccessKey, config.Schema.Cache.SecretKey)
	cfg := storage.Config{
		// 是否使用https域名进行资源管理
		UseHTTPS: false,
	}
	// 指定空间所在的区域，如果不指定将自动探测
	// 如果没有特殊需求，默认不需要指定
	//cfg.Zone=&storage.ZoneHuabei
	bucketManager := storage.NewBucketManager(mac, &cfg)
	fileInfo, err := bucketManager.Stat(config.Schema.Cache.Bucket, key)
	if err == nil {
		return &fileInfo
	}
	logger.Warn("cache: check file info failed from qiniu cache!!!", zap.String("key", key))
	return nil
}

func UploadFileInfo(file multipart.File, filename string, kind int) (*nosql.FileInfo, error) {
	var info = new(nosql.FileInfo)
	var err error
	if kind == UploadMongo {
		info, err = nosql.CreateAssetInfoFile(file, filename)
	} else if kind == UploadQN {
		info, err = CreateFileByQN(file, filename)
	}
	return info, err
}

func CreateFileByQN(from multipart.File, filename string) (*nosql.FileInfo, error) {
	logger.Info("QN create file by qi niu :: ", zap.Any("filename", filename))
	// 自定义返回值结构体
	type MyPutRet struct {
		Key      string
		Hash     string
		Fsize    int
		MimeType string
		Name     string
	}

	var info = new(nosql.FileInfo)
	data, err := ioutil.ReadAll(from)
	if err != nil {
		return info, err
	}

	info.MD5 = tool.CalculateMD5(data)
	putPolicy := storage.PutPolicy{
		Scope:      config.Schema.Cache.Bucket,
		ReturnBody: `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"mimeType":"$(mimeType)","name":"$(x:name)"}`,
	}
	mac := qbox.NewMac(config.Schema.Cache.AccessKey, config.Schema.Cache.SecretKey)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = &storage.Zone_z2
	// 是否使用https域名
	cfg.UseHTTPS = false
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	formUploader := storage.NewFormUploader(&cfg)
	ret := MyPutRet{}
	putExtra := storage.PutExtra{
		Params: map[string]string{
			"x:name": filename,
		},
	}

	key := tool.CreateUUID()
	er := formUploader.Put(context.Background(), &ret, upToken, key, bytes.NewReader(data), int64(len(data)), &putExtra)
	if er != nil {
		return info, err
	}

	info.UID = ret.Key
	info.Name = ret.Name
	info.Type = ret.MimeType
	//info.MD5 = ret.Hash
	info.Size = int64(ret.Fsize)
	return info, nil
}
