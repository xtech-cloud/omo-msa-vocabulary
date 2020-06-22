package grpc

import (
	"context"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
)

type AssetService struct {}

func (mine *AssetService)AddOne(ctx context.Context, in *pb.ReqAssetAdd, out *pb.ReplyAssetOne) error {
	var err error
	return err
}

func (mine *AssetService)GetOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyAssetOne) error {
	var err error
	return err
}

func (mine *AssetService)RemoveOne(ctx context.Context, in *pb.RequestInfo, out *pb.ReplyAssetOne) error {
	var err error
	return err
}

func (mine *AssetService)GetList(ctx context.Context, in *pb.ReqAssetList, out *pb.ReplyAssetList) error {
	var err error
	return err
}

