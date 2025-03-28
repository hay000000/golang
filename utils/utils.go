package utils

import (
	"encoding/json"
	"log"
)

var logFn = log.Panic

// HandleErr : 에러가 발생시 처리
func HandleErr(err error) {
	if err != nil {
		logFn(err)
	}
}

// StructToBytes : Struct 데이터를 Bytes 로 변환
func StructToBytes(data interface{}) []byte {
	bytes, err := json.Marshal(data)
	HandleErr(err)
	return bytes
}
