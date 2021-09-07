package grpc

import (
	"encoding/json"
	"github.com/micro/go-micro/v2/logger"
	pb "github.com/xtech-cloud/omo-msp-vocabulary/proto/vocabulary"
	"reflect"
)

func inLog(name, data interface{}) {
	bytes, _ := json.Marshal(data)
	msg := ByteString(bytes)
	logger.Infof("[in.%s]:data = %s", name, msg)
}

func outError(name, msg string, code pb.ResultStatus) *pb.ReplyStatus {
	logger.Warnf("[error.%s]:code = %d, msg = %s", name, code, msg)
	tmp := &pb.ReplyStatus{
		Code: uint32(code),
		Msg:  msg,
	}
	return tmp
}

func outLog(name, data interface{}) *pb.ReplyStatus {
	bytes, _ := json.Marshal(data)
	msg := ByteString(bytes)
	logger.Infof("[out.%s]:data = %s", name, msg)
	tmp := &pb.ReplyStatus{
		Code: 0,
		Msg:  "",
	}
	return tmp
}

func ByteString(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}

func checkPage(page, number int32, all interface{}) (int32, int32, interface{}) {
	if number < 1 {
		number = 10
	}
	array := reflect.ValueOf(all)
	total := int32(array.Len())
	maxPage := total/number + 1
	if page < 1 {
		return total, maxPage, all
	}

	var start = (page - 1) * number
	var end = start + number
	if end > total-1 {
		end = total - 1
	}

	list := array.Slice(int(start), int(end))
	return total, maxPage, list.Interface()
}
