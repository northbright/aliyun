package message_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/northbright/aliyun/message"
)

type Config struct {
	AccessKeyID     string   `json:"access_key_id"`
	AccessKeySecret string   `json:"access_key_secret"`
	PhoneNumbers    []string `json:"phone_numbers"`
	SignName        string   `json:"sign_name"`
	TemplateCode    string   `json:"template_code"`
	TemplateParam   string   `json:"template_param"`
}

func Example() {
	var (
		err    error
		config Config
	)

	// Load config from file.
	// You may rename "config.example.json" to "config.json" and modify it.
	// It looks like this:
	// {
	//    "access_key_id":"testId",
	//    "access_key_secret":"testSecret",
	//    "phone_numbers":["15300000001"],
	//    "sign_name":"阿里云短信测试专用",
	//    "template_code":"SMS_71390007",
	//    "template_param":"{\"code\":\"888888\"}"
	//}
	if err = loadConfig("config.json", &config); err != nil {
		log.Printf("loadConfig() error: %v", err)
		return
	}

	// Creates a new client.
	client := message.NewClient(config.AccessKeyID, config.AccessKeySecret)
	log.Printf("client: %v", client)
	// Send SMS.
	ok, resp, err := client.SendSMS(config.PhoneNumbers,
		config.SignName,
		config.TemplateCode,
		config.TemplateParam,
	)
	if err != nil {
		log.Printf("SendSMS() error: %v", err)
		return
	}
	log.Printf("SendSMS() ok: %v, response: %v", ok, resp)

	// Output:
}

// loadConfig loads app config.
func loadConfig(configFile string, config *Config) error {
	// Load Conifg
	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("load config file error: %v", err)

	}

	if err = json.Unmarshal(buf, config); err != nil {
		return fmt.Errorf("parse config err: %v", err)
	}

	return nil
}
