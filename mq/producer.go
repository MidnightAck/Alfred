package mq

import (
	"Alfred/config"
	"log"
	"github.com/streadyway/amqp"

)

var conn *amqp.Connection
var channel *amqp.Channel

// 如果异常关闭，会接收通知
var notifyClose chan *amqp.Error

func init() {
	// 是否开启异步转移功能，开启时才初始化rabbitMQ连接
	if !config.AsyncTransferEnable {
		return
	}
	if initChannel() {
		channel.NotifyClose(notifyClose)
	}
	// 断线自动重连
	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				log.Printf("onNotifyChannelClosed: %+v\n", msg)
				initChannel()
			}
		}
	}()
}

func initChannel() bool {
	if channel != nil {
		return true
	}
	//获得链接
	conn, err := amqp.Dial(config.RabbitURL)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	//打开channel用于链接和接受
	channel, err = conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}

// Publish : 发布消息 接受exchange交换机
func Publish(exchange, routingKey string, msg []byte) bool {
	//判断channel是否正确
	if !initChannel() {
		return false
	}

	//执行消息发布动作
	if nil == channel.Publish(
		exchange,
		routingKey,
		false, // 如果没有对应的queue, 就会丢弃这条小心
		false, //
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg}) {
		return true
	}
	return false
}