# Alfred-Personal Data Butler

![banner](https://s1.ax1x.com/2020/04/18/JmVm8I.png)


Alfred是一款私有云盘，通过简单的搭建可以使个人轻松的使用。

Alfred支持各种文件类型，支持大文件上传，支持秒传、分块上传、断点续传、异步队列上传。同时，在数据方面支持将文件上传至多个云端，如阿里云。同时，用户可搭建自己的ceph集群用于存储数据。

## Table of Contents


- [Infrastructure](#Infrastructure)
- [Install](#install)

## Infrastructure
![img](https://s1.ax1x.com/2020/04/18/JmY8ds.png)

系统的架构如上所示，云端本系统选用的是阿里云OSS。通过微服务部署，使得整个系统变得更于维护和升级。

![img](https://s1.ax1x.com/2020/04/18/JmY3Zj.png)

系统的业务流程如上所示。

## Install
```bash
git clone git@github.com:CrowFea/Alfred.git
```

你需要在根目录下创建config文件夹，在其中定义你的连接参数。

```
crowfea@crowfea-HP-Pavilion-Notebook:~/goProject/src/Alfred$ tree
.
├── cache
│   └── redis
│       └── conn.go
├── common
│   └── common.go
├── config
│   ├── ceph.go
│   ├── mysql.go
│   ├── oss.go
│   ├── rabbitmq.go
│   ├── service.go
│   └── store.go
├── db
│   ├── file.go
│   ├── mysql
│   │   └── conn.go
│   ├── userfile.go
│   └── user.go

```
### config setting
```
    //Ceph集群
	// CephAccessKey : 访问Key
	CephAccessKey = ""
	// CephSecretKey : 访问密钥
	CephSecretKey = ""
	// CephGWEndpoint : gateway地址
	CephGWEndpoint = ""

    //mysql配置参数
    userName=""
	password=""
	ip=""
	port=""
	dbName=""
	MySQLSource = ""

    //OSS配置参数
    //OSS bucket名称
    OSSBucket = ""
	// OSSEndpoint : oss endpoint
	OSSEndpoint = ""
	// OSSAccesskeyID : oss访问key
	OSSAccesskeyID = ""
	// OSSAccessKeySecret : oss访问key secret
	OSSAccessKeySecret = ""

    //RabbitMQ配置参数
    // AsyncTransferEnable : 是否开启文件异步转移(默认同步)
	AsyncTransferEnable = true
	// RabbitURL : rabbitmq服务的入口url
	RabbitURL = ""
	// TransExchangeName : 用于文件transfer的交换机
	TransExchangeName = ""
	// TransOSSQueueName : oss转移队列名
	TransOSSQueueName = ""
	// TransOSSErrQueueName : oss转移失败后写入另一个队列的队列名
	TransOSSErrQueueName = ""
	// TransOSSRoutingKey : routingkey
	TransOSSRoutingKey = ""

    // UploadServiceHost : 上传服务监听的地址
	UploadServiceHost = ""

    // TempLocalRootDir : 本地临时存储地址的路径
	TempLocalRootDir = ""
	// TempPartRootDir : 分块文件在本地临时存储地址的路径
	TempPartRootDir = ""
	// CephRootDir : Ceph的存储路径prefix
	CephRootDir = ""
	// OSSRootDir : OSS的存储路径prefix
	OSSRootDir = ""
	// CurrentStoreType : 设置当前文件的存储类型
	CurrentStoreType = 
```


