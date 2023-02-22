package rabbitmq

import (
	"context"
	"fmt"
	"runtime/debug"
	"strconv"
	"testing"
	"time"
)

func ExpectEqual(left interface{}, right interface{}, t *testing.T) {
	if left != right {
		t.Fatalf("expected %v == %v\n%s", left, right, debug.Stack())
	}
}

func ExpectUnEqual(left interface{}, right interface{}, t *testing.T) {
	if left == right {
		t.Fatalf("expected %v != %v\n%s", left, right, debug.Stack())
	}
}

func TestPublish(t *testing.T) {
	ctx := context.Background()
	rabbitmq := NewRabbitMQSimple("newProduct")
	for i := 0; i < 20; i++ {
		rabbitmq.PublishSimple(ctx, []byte("订阅模式生产第"+strconv.Itoa(i)+"条"+"数据"))
		fmt.Println("订阅模式生产第" + strconv.Itoa(i) + "条" + "数据")
		time.Sleep(1 * time.Second)
	}
}
