/*
 * File: server_option.go
 * Project: server
 * File Created: Sunday, 7th April 2019 12:37:32 am
 * Author: lizongrong (389006500@qq.com)
 * -----
 * Last Modified: Monday, 8th April 2019 2:07:54 pm
 * Modified By: lizongrong (389006500@qq.com>)
 * -----
 * Copyright 2019 - 2019
 */
package server

import (
	"time"

	"github.com/tiancai110a/go-rpc/protocol"
	"github.com/tiancai110a/go-rpc/transport"
)

type Option struct {
	ProtocolType  protocol.ProtocolType
	SerializeType protocol.SerializeType
	CompressType  protocol.CompressType
	TransportType transport.TransportType
	ShutDownWait  time.Duration
	Wrappers      []Wrapper
	ShutDownHooks []ShutDownHook
}

var DefaultOption = Option{
	ProtocolType:  protocol.Default,
	SerializeType: protocol.SerializeTypeJson,
	CompressType:  protocol.CompressTypeNone,
	TransportType: transport.TCPTransport,
	ShutDownWait:  time.Second * 12,
}

type ShutDownHook func(s *SGServer)
