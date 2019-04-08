/*
 * File: testservice.go
 * Project: service
 * File Created: Sunday, 7th April 2019 5:59:05 pm
 * Author: lizongrong (389006500@qq.com)
 * -----
 * Last Modified: Monday, 8th April 2019 2:08:12 pm
 * Modified By: lizongrong (389006500@qq.com>)
 * -----
 * Copyright 2019 - 2019
 */
package service

type TestRequest struct {
	A int //发送的参数
	B int
}

type TestResponse struct {
	Reply int //返回的参数
}

type TestService struct {
}

func (t TestService) Add(req *TestRequest, res *TestResponse) {
	res.Reply = req.A + req.B
}