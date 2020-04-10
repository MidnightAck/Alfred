package ceph

import (
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
	cfg "Alfred/config"
)

var cephConn *s3.S3

//获取Ceph的connection方法
func GetCephConnection() *s3.S3{
	if cephConn!=nil {
		return cephConn
	}
	//初始化Ceph信息
	
	auth:=aws.Auth{
		AccessKey: cfg.CephAccessKey,
		SecretKey: cfg.CephSecretKey,
	}
	
	curRegion:=aws.Region{
		Name:                 "default",
		EC2Endpoint:          cfg.CephGWEndpoint,
		S3Endpoint:           cfg.CephGWEndpoint,
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

// PutObject : 上传文件到ceph集群
func PutObject(bucket string, path string, data []byte) error {
	return GetCephBucket(bucket).Put(path, data, "octet-stream", s3.PublicRead)
}