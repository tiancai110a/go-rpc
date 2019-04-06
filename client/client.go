/*
 * File: client.go
 * Project: client
 * File Created: Friday, 5th April 2019 5:50:17 pm
 * Author: lizongrong (389006500@qq.com)
 * -----
 * Last Modified: Friday, 5th April 2019 5:50:27 pm
 * Modified By: lizongrong (389006500@qq.com>)
 * -----
 * Copyright lizongrong - 2019
 */
package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"sync/atomic"

	"github.com/golang/glog"
	"github.com/tiancai110a/go-rpc/transport"
)

//用来传递参数的通用结构体
type Test struct {
	Seq   uint64
	A     int //发送的参数
	B     int
	Reply *int //返回的参数
}

type Call struct {
	ServiceMethod string     // 服务名.方法名
	Error         error      // 错误信息
	Done          chan *Call // 在调用结束时激活
	Payload       *Test
}

func (c *Call) done() {
	c.Done <- c
}

//Recv 暂时用来处理response的handler
func Recv(conn transport.Transport) (error, *Test) {
	data := make([]byte, 10000)
	n, err := conn.Read(data)

	if err != nil {
		return err, nil
	}
	t := Test{}
	err = json.Unmarshal(data[:n], &t)

	if err != nil {
		glog.Error("read failed: ", err)
		return err, nil
	}
	return err, &t
}

//RPCClient
type RPCClient interface {
	Call(ctx context.Context, a int, b int, reply *int) error
	Close() error
}

//SimpleClient
type SimpleClient struct {
	rwc          io.ReadWriteCloser
	pendingCalls sync.Map
	seq          uint64
}

func (c *SimpleClient) input(s transport.Transport) {
	var err error
	for err == nil {
		var t *Test
		err, t = Recv(s)

		if err != nil {
			break
		}

		seq := t.Seq
		TestInterface, _ := c.pendingCalls.Load(seq)
		call := TestInterface.(*Call)
		c.pendingCalls.Delete(seq)

		switch {
		case call == nil:
			glog.Error("call is canceled before")
		default:
			*(call.Payload.Reply) = *(t.Reply)
			call.done()
		}

	}

}

func NewRPCClient(network string, addr string) (RPCClient, error) {
	c := SimpleClient{}
	tr := transport.Socket{}
	err := tr.Dial(network, addr)
	if err != nil {
		glog.Error("Connect err:", err)
		return nil, err
	}
	c.rwc = &tr

	go c.input(&tr)
	return &c, nil
}

//Close 关闭连接
func (c *SimpleClient) Close() error {
	err := c.rwc.Close()
	if err != nil {
		glog.Info("socket already clsosed")
	}
	return err
}

//Call call是调用rpc的入口，pack打包request，send负责序列化和发送
//TODO 加入超时限制
func (c *SimpleClient) Call(ctx context.Context, a int, b int, reply *int) error {
	seq := atomic.AddUint64(&c.seq, 1)
	ctx = context.WithValue(ctx, "RequestSeqKey", seq)

	done := make(chan *Call, 1)

	call := c.pack(ctx, done, a, b, reply)
	select {
	case <-ctx.Done():
		c.pendingCalls.Delete(seq)
		call.Error = errors.New("client request time out")
	case <-call.Done:
	}
	return call.Error
}

func (c *SimpleClient) pack(ctx context.Context, done chan *Call, a, b int, reply *int) *Call {
	call := new(Call)
	call.ServiceMethod = "test" //服务名加方法名

	t := Test{}
	t.A = a
	t.B = b
	t.Reply = reply
	call.Payload = &t
	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		if cap(done) == 0 {
			panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done

	c.send(ctx, call)

	return call
}

func (c *SimpleClient) send(ctx context.Context, call *Call) error {
	seq := ctx.Value("RequestSeqKey").(uint64)
	call.Payload.Seq = seq
	c.pendingCalls.Store(seq, call)

	data, err := json.Marshal(call.Payload)

	if err != nil {
		glog.Error("Marshal failed", err)
		return err
	}

	_, err = c.rwc.Write(data)
	if err != nil {
		glog.Error("Write failed", err)
		return err
	}
	return err

}