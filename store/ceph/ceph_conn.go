package ceph

import (
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

var cephConn *s3.S3

//获取Ceph的connection方法
func GetCephConnection() *s3.S3{
	if cephConn!=nil {
		return cephConn
	}
	//初始化Ceph信息
	
	auth:=aws.Auth{
		AccessKey: "C7MBLZSLY75BAO49O7UY",
		SecretKey: "byPyebXSdLqac4LRWANF9nDEziKysPn2GieIRWdo",
	}
	
	curRegion:=aws.Region{
		Name:                 "default",
		EC2Endpoint:          "http://127.0.0.1:9080",
		S3Endpoint:           "http://127.0.0.1:9080",
		S3BucketEndpoint:     "",
		S3LocationConstraint: false,
		S3LowercaseBucket:    false,
		Sign:                 aws.SignV2,
	}

	//创建S3类型的链接
	return s3.New(auth,curRegion)
	
}

func GetCephBucket(bucket string) *s3.Bucket {
	conn:=GetCephConnection()
	return conn.Bucket(bucket)
}