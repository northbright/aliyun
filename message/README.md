# message

[![Build Status](https://travis-ci.org/northbright/aliyun.svg?branch=master)](https://travis-ci.org/northbright/aliyun)
[![GoDoc](https://godoc.org/github.com/northbright/aliyun/message?status.svg)](https://godoc.org/github.com/northbright/aliyun/message)

message is a [Golang](https://golang.org) SDK for aliyun message services(阿里云通信服务).

#### Supported Services
* [SMS(短消息服务)](https://www.aliyun.com/product/sms)
* [Voice service(语音服务)](https://www.aliyun.com/product/sms)

#### Example of Sending SMS

    // Create a new client.
    c := messgae.NewClient(accessKeyID, accessKeySecret)

    // Specify one or more phone numbers.
    numbers := []string{"13800138000"}
    
    // Pass phone numbers, signature name, template code, template param(JSON) to Send().
    ok, resp, err := c.Send(numbers, "my_product", "SMS_0000", `{"code":"1234","product":"ytx"}`)

#### Documentation
* [API References](https://godoc.org/github.com/northbright/aliyun/message)

#### License
* [MIT License](../LICENSE)

