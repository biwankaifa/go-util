package rabbitmq

import (
	//"errors"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

// 定义全局变量,指针类型
var mqConn *amqp.Connection

//var mqChan *amqp.Channel

// Producer 定义生产者接口
type Producer interface {
	MsgContent() string
}

// RetryProducer 定义生产者接口
type RetryProducer interface {
	MsgContent() string
}

// Receiver 定义接收者接口
type Receiver interface {
	Options() QueueExchange
	Consumer([]byte) error
	FailAction(error, []byte) error
	Send(...interface{}) error
	Recv(int)
}

// RabbitMQ 定义RabbitMQ对象
type RabbitMQ struct {
	connection        *amqp.Connection
	Channel           *amqp.Channel
	dns               string
	QueueName         string // 队列名称
	RoutingKey        string // key名称
	ExchangeName      string // 交换机名称
	ExchangeType      string // 交换机类型
	producerList      []Producer
	retryProducerList []RetryProducer
	receiverList      []Receiver
}

// QueueExchange 定义队列交换机对象
type QueueExchange struct {
	QuName string // 队列名称
	RtKey  string // key值
	ExName string // 交换机名称
	ExType string // 交换机类型
	Dns    string //链接地址
}

// MqConnect 链接rabbitMQ
func (mq *RabbitMQ) MqConnect() (err error) {

	mqConn, err = amqp.Dial(mq.dns)
	mq.connection = mqConn // 赋值给RabbitMQ对象

	if err != nil {
		fmt.Printf("关闭mq链接失败  :%s \n", err)
	}

	return
}

// CloseMqConnect 关闭mq链接
func (mq *RabbitMQ) CloseMqConnect() (err error) {

	err = mq.connection.Close()
	if err != nil {
		fmt.Printf("关闭mq链接失败  :%s \n", err)
	}
	return
}

// MqOpenChannel 链接rabbitMQ
func (mq *RabbitMQ) MqOpenChannel() (err error) {
	mqConn := mq.connection
	mq.Channel, err = mqConn.Channel()
	//defer mqChan.Close()
	if err != nil {
		fmt.Printf("MQ打开管道失败:%s \n", err)
	}
	return err
}

// CloseMqChannel 链接rabbitMQ
func (mq *RabbitMQ) CloseMqChannel() (err error) {
	err = mq.Channel.Close()
	if err != nil {
		fmt.Printf("关闭mq链接失败  :%s \n", err)
	}
	return err
}

// NewMq 创建一个新的操作对象
func NewMq(q QueueExchange) RabbitMQ {
	return RabbitMQ{
		QueueName:    q.QuName,
		RoutingKey:   q.RtKey,
		ExchangeName: q.ExName,
		ExchangeType: q.ExType,
		dns:          q.Dns,
	}
}

// sendMsg
func (mq *RabbitMQ) sendMsg(body string) {
	err := mq.MqOpenChannel()
	ch := mq.Channel
	if err != nil {
		log.Printf("Channel err  :%s \n", err)
	}

	defer func(Channel *amqp.Channel) {
		err := Channel.Close()
		if err != nil {

		}
	}(mq.Channel)
	if mq.ExchangeName != "" {
		if mq.ExchangeType == "" {
			mq.ExchangeType = "direct"
		}
		err = ch.ExchangeDeclare(mq.ExchangeName, mq.ExchangeType, true, false, false, false, nil)
		if err != nil {
			log.Printf("ExchangeDeclare err  :%s \n", err)
		}
	}

	// 用于检查队列是否存在,已经存在不需要重复声明
	_, err = ch.QueueDeclare(mq.QueueName, true, false, false, false, nil)
	if err != nil {
		log.Printf("QueueDeclare err :%s \n", err)
	}
	// 绑定任务
	if mq.RoutingKey != "" && mq.ExchangeName != "" {
		err = ch.QueueBind(mq.QueueName, mq.RoutingKey, mq.ExchangeName, false, nil)
		if err != nil {
			log.Printf("QueueBind err :%s \n", err)
		}
	}

	if mq.ExchangeName != "" && mq.RoutingKey != "" {
		err = mq.Channel.Publish(
			mq.ExchangeName, // exchange
			mq.RoutingKey,   // routing key
			false,           // mandatory
			false,           // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
	} else {
		err = mq.Channel.Publish(
			"",           // exchange
			mq.QueueName, // routing key
			false,        // mandatory
			false,        // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
	}

}

func (mq *RabbitMQ) sendRetryMsg(body string, retryNums int32, args ...string) {
	err := mq.MqOpenChannel()
	ch := mq.Channel
	if err != nil {
		log.Printf("Channel err  :%s \n", err)
	}
	defer func(Channel *amqp.Channel) {
		_ = Channel.Close()
	}(mq.Channel)

	if mq.ExchangeName != "" {
		if mq.ExchangeType == "" {
			mq.ExchangeType = "direct"
		}
		err = ch.ExchangeDeclare(mq.ExchangeName, mq.ExchangeType, true, false, false, false, nil)
		if err != nil {
			log.Printf("ExchangeDeclare err  :%s \n", err)
		}
	}

	//原始路由key
	oldRoutingKey := args[0]
	//原始交换机名
	oldExchangeName := args[1]

	table := make(map[string]interface{}, 3)
	table["x-dead-letter-routing-key"] = oldRoutingKey
	if oldExchangeName != "" {
		table["x-dead-letter-exchange"] = oldExchangeName
	} else {
		mq.ExchangeName = ""
		table["x-dead-letter-exchange"] = ""
	}

	table["x-message-ttl"] = int64(20000)

	//fmt.Printf("%+v",table)
	//fmt.Printf("%+v",mq)
	// 用于检查队列是否存在,已经存在不需要重复声明
	_, err = ch.QueueDeclare(mq.QueueName, true, false, false, false, table)
	if err != nil {
		log.Printf("QueueDeclare err :%s \n", err)
	}
	// 绑定任务
	if mq.RoutingKey != "" && mq.ExchangeName != "" {
		err = ch.QueueBind(mq.QueueName, mq.RoutingKey, mq.ExchangeName, false, nil)
		if err != nil {
			log.Printf("QueueBind err :%s \n", err)
		}
	}

	header := make(map[string]interface{}, 1)

	header["retry_nums"] = retryNums + int32(1)

	var ttlExchange string
	var ttlRoutkey string

	if mq.ExchangeName != "" {
		ttlExchange = mq.ExchangeName
	} else {
		ttlExchange = ""
	}

	if mq.RoutingKey != "" && mq.ExchangeName != "" {
		ttlRoutkey = mq.RoutingKey
	} else {
		ttlRoutkey = mq.QueueName
	}

	//fmt.Printf("ttl_exchange:%s,ttl_routkey:%s \n",ttl_exchange,ttl_routkey)
	err = mq.Channel.Publish(
		ttlExchange, // exchange
		ttlRoutkey,  // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
			Headers:     header,
		})
	if err != nil {
		fmt.Printf("MQ任务发送失败:%s \n", err)

	}

}

// ListenReceiver 监听接收者接收任务 消费者
func (mq *RabbitMQ) ListenReceiver(receiver Receiver, routineNum int) {
	err := mq.MqOpenChannel()
	ch := mq.Channel
	if err != nil {
		log.Printf("Channel err  :%s \n", err)
	}
	defer func(Channel *amqp.Channel) {
		_ = Channel.Close()
	}(mq.Channel)
	if mq.ExchangeName != "" {
		if mq.ExchangeType == "" {
			mq.ExchangeType = "direct"
		}
		err = ch.ExchangeDeclare(mq.ExchangeName, mq.ExchangeType, true, false, false, false, nil)
		if err != nil {
			log.Printf("ExchangeDeclare err  :%s \n", err)
		}
	}

	// 用于检查队列是否存在,已经存在不需要重复声明
	_, err = ch.QueueDeclare(mq.QueueName, true, false, false, false, nil)
	if err != nil {
		log.Printf("QueueDeclare err :%s \n", err)
	}
	// 绑定任务
	if mq.RoutingKey != "" && mq.ExchangeName != "" {
		err = ch.QueueBind(mq.QueueName, mq.RoutingKey, mq.ExchangeName, false, nil)
		if err != nil {
			log.Printf("QueueBind err :%s \n", err)
		}
	}
	// 获取消费通道,确保rabbitMQ一个一个发送消息
	err = ch.Qos(1, 0, false)
	msgList, err := ch.Consume(mq.QueueName, "sgen-1", false, false, false, false, nil)
	if err != nil {
		log.Printf("Consume err :%s \n", err)
	}
	for msg := range msgList {
		retryNums, ok := msg.Headers["retry_nums"].(int32)
		if !ok {
			retryNums = int32(0)
		}
		// 处理数据
		err := receiver.Consumer(msg.Body)
		if err != nil {
			//消息处理失败 进入延时尝试机制
			if retryNums < 3 {
				retryMsg(msg.Body, retryNums, QueueExchange{
					mq.QueueName,
					mq.RoutingKey,
					mq.ExchangeName,
					mq.ExchangeType,
					mq.dns,
				})
				fmt.Println(err)
			} else {
				// TODO 消息失败 入库db
				_ = receiver.FailAction(err, msg.Body)
			}
			err = msg.Ack(true)
			if err != nil {
				fmt.Printf("确认消息未完成异常:%s \n", err)
			}
		} else {
			// 确认消息,必须为false
			err = msg.Ack(true)

			if err != nil {
				fmt.Printf("消息消费ack失败 err :%s \n", err)
			}
		}

	}
}

// retryMsg 消息处理失败之后 延时尝试
func retryMsg(msg []byte, retryNums int32, queueExchange QueueExchange) {
	//原始队列名称 交换机名称
	oldQName := queueExchange.QuName
	oldExchangeName := queueExchange.ExName
	oldRoutingKey := queueExchange.RtKey
	if oldRoutingKey == "" || oldExchangeName == "" {
		oldRoutingKey = oldQName
	}

	if queueExchange.QuName != "" {
		queueExchange.QuName = queueExchange.QuName + "_retry_3"
	}

	if queueExchange.RtKey != "" {
		queueExchange.RtKey = queueExchange.RtKey + "_retry_3"
	} else {
		queueExchange.RtKey = queueExchange.QuName + "_retry_3"
	}

	//fmt.Printf("%+v",queueExchange)

	mq := NewMq(queueExchange)
	_ = mq.MqConnect()

	defer func() {
		_ = mq.CloseMqConnect()
	}()
	//fmt.Printf("%+v",queueExchange)
	mq.sendRetryMsg(string(msg), retryNums, oldRoutingKey, oldExchangeName)

}

// Send 生产者
func Send(queueExchange QueueExchange, msg string) {
	mq := NewMq(queueExchange)
	_ = mq.MqConnect()

	defer func() {
		_ = mq.CloseMqConnect()
	}()
	mq.sendMsg(msg)

}

// Recv 消费者
// runNums  开启并发执行任务数量
func Recv(queueExchange QueueExchange, receiver Receiver, runNums int) {
	mq := NewMq(queueExchange)
	_ = mq.MqConnect()

	defer func() {
		_ = mq.CloseMqConnect()
	}()

	forever := make(chan bool)
	for i := 1; i <= runNums; i++ {
		go func(routineNum int) {
			defer func(Channel *amqp.Channel) {
				if Channel != nil {
					_ = Channel.Close()
				}
			}(mq.Channel)
			// 验证链接是否正常
			_ = mq.MqOpenChannel()
			mq.ListenReceiver(receiver, routineNum)
		}(i)
	}
	<-forever
}
