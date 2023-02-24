package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// rabbitMQ结构体
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	//队列名称
	QueueName string
	//交换机名称
	Exchange string
	//bind Key 名称
	Key string
	//连接信息
	Mqurl string
	Queue amqp.Queue
	// 通知
	notifyClose   chan *amqp.Error       // 如果异常关闭，会接收数据
	notifyConfirm chan amqp.Confirmation // 消息发送成功确认，会接收到数据
}

// 创建结构体实例
func NewRabbitMQ(queueName string, exchange string, key string) *RabbitMQ {
	return &RabbitMQ{QueueName: queueName, Exchange: exchange, Key: key, Mqurl: MqUrl}
}

// 断开channel 和 connection
func (r *RabbitMQ) Destroy() {
	r.channel.Close()
	r.conn.Close()
}

// 错误处理函数
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}

// 创建简单模式下RabbitMQ实例
func NewRabbitMQSimple(queueName string) *RabbitMQ {
	// 创建RabbitMQ实例
	rabbitmq := NewRabbitMQ(queueName, "", "")
	var err error
	// 获取connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "failed to connect rabbitmq!")
	// 获取channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a channel")
	// 注册监听
	rabbitmq.channel.NotifyClose(rabbitmq.notifyClose)
	rabbitmq.channel.NotifyPublish(rabbitmq.notifyConfirm)
	return rabbitmq
}

// PublishSimple 简单模式队列生产
func (r *RabbitMQ) PublishSimple(ctx context.Context, message []byte) error {
	//1.申请队列，如果队列不存在会自动创建，存在则跳过创建
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		// 是否持久化
		false,
		// 是否自动删除
		false,
		// 是否具有排他性
		false,
		// 是否阻塞处理
		false,
		// 额外的属性
		nil,
	)
	if err != nil {
		fmt.Println(err)
		logger.Errorf("MQ 生产者错误：%v", err.Error())
		return err
	}
	// 调用channel 发送消息到队列中
	err = r.channel.PublishWithContext(
		ctx,
		r.Exchange,
		r.QueueName,
		// 如果为true，根据自身exchange类型和routekey规则无法找到符合条件的队列会把消息返还给发送者
		false,
		// 如果为true，当exchange发送消息到队列后发现队列上没有消费者，则会把消息返还给发送者
		false,
		amqp.Publishing{
			ContentType: "application/json", //设置消息请求头为json
			Body:        message,
			Timestamp:   time.Now(),
		})
	if err != nil {
		logger.Errorf("MQ 生产者错误：%v", err.Error())
		return err
	}
	return nil
}

// ConsumeSimple simple 模式下消费者
func (r *RabbitMQ) ConsumeSimple() (<-chan amqp.Delivery, error) {
	//1.申请队列，如果队列不存在会自动创建，存在则跳过创建
	q, err := r.channel.QueueDeclare(
		r.QueueName,
		// 是否持久化
		false,
		// 是否自动删除
		false,
		// 是否具有排他性
		false,
		// 是否阻塞处理
		false,
		// 额外的属性
		nil,
	)
	if err != nil {
		logger.Errorf("MQ 消费者错误：%v", err.Error())
		return nil, err
	}

	//接收消息
	msgs, err := r.channel.Consume(
		q.Name, // queue
		// 用来区分多个消费者
		"", // consumer
		// 是否自动应答
		true, // auto-ack
		// 是否独有
		false, // exclusive
		// 设置为true，表示 不能将同一个Connection中生产者发送的消息传递给这个Connection中的消费者
		false, // no-local
		// 列是否阻塞
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		logger.Errorf("MQ 消费者错误：%v", err.Error())
		return nil, err
	}
	return msgs, nil
}

func (r *RabbitMQ) DeclareQueue() error {
	q, err := r.channel.QueueDeclare(
		r.QueueName,
		//是否持久化
		false,
		//是否自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞处理
		false,
		//额外的属性
		nil,
	)
	if err != nil {
		logger.Errorln(err.Error())
	}
	r.Queue = q
	return err
}
