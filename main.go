package main

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	_ "github.com/micro/go-plugins/registry/consul/v2"
	_ "github.com/micro/go-plugins/registry/etcdv3/v2"
	"github.com/robfig/cron"
	proto "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"io"
	"omo.msa.vocabulary/cache"
	"omo.msa.vocabulary/config"
	"omo.msa.vocabulary/grpc"
	"os"
	"path/filepath"
	"time"
)

var (
	BuildVersion string
	BuildTime    string
	CommitID     string
)

func main() {
	config.Setup()
	err := cache.InitData()
	if err != nil {
		panic(err)
	}
	// New Service
	service := micro.NewService(
		micro.Name("omo.msa.vocabulary"),
		micro.Version("latest"),
		micro.RegisterTTL(time.Second*time.Duration(config.Schema.Service.TTL)),
		micro.RegisterInterval(time.Second*time.Duration(config.Schema.Service.Interval)),
		micro.Address(config.Schema.Service.Address),
	)
	// Initialise service
	service.Init()
	// Register Handler
	_ = proto.RegisterEntityServiceHandler(service.Server(), new(grpc.EntityService))
	_ = proto.RegisterConceptServiceHandler(service.Server(), new(grpc.ConceptService))
	_ = proto.RegisterGraphServiceHandler(service.Server(), new(grpc.GraphService))
	_ = proto.RegisterAttributeServiceHandler(service.Server(), new(grpc.AttributeService))
	_ = proto.RegisterRelationServiceHandler(service.Server(), new(grpc.RelationService))
	_ = proto.RegisterEventServiceHandler(service.Server(), new(grpc.EventService))

	checkTimer()

	app, _ := filepath.Abs(os.Args[0])

	BuildVersion := "1.1.2"
	BuildTime := time.Now().String()
	CommitID := "3"
	logger.Info("-------------------------------------------------------------")
	logger.Info("- Micro Service Agent -> Run")
	logger.Info("-------------------------------------------------------------")
	logger.Infof("- version      : %s", BuildVersion)
	logger.Infof("- application  : %s", app)
	logger.Infof("- md5          : %s", md5hex(app))
	logger.Infof("- build        : %s", BuildTime)
	logger.Infof("- commit       : %s", CommitID)
	logger.Info("-------------------------------------------------------------")
	// Run service
	if err := service.Run(); err != nil {
		logger.Fatal(err)
	}
}

func checkTimer() {
	c := cron.New()
	_ = c.AddFunc("*/3 * * * * ?", func() {
		cache.Context().CheckSyncNodes()
		cache.Context().CheckSyncLinks()
	})
	c.Start()
}

func md5hex(_file string) string {
	h := md5.New()

	f, err := os.Open(_file)
	if err != nil {
		return ""
	}
	defer f.Close()

	io.Copy(h, f)

	return hex.EncodeToString(h.Sum(nil))
}
