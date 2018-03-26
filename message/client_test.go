package message_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/northbright/aliyun/message"
)

type SMSConfig struct {
	PhoneNumbers  []string `json:"phone_numbers"`
	SignName      string   `json:"sign_name"`
	TemplateCode  string   `json:"template_code"`
	TemplateParam string   `json:"template_param"`
}

type SingleCallByTTSConfig struct {
	CalledShowNumber string `json:"called_show_number"`
	CalledNumber     string `json:"called_number"`
	TemplateCode     string `json:"template_code"`
	TemplateParam    string `json:"template_param"`
}

type Config struct {
	AccessKeyID     string                `json:"access_key_id"`
	AccessKeySecret string                `json:"access_key_secret"`
	SMS             SMSConfig             `json:"sms"`
	SingleCallByTTS SingleCallByTTSConfig `json:"single_call_by_tts"`
}

func Example() {
	var (
		err    error
		config Config
	)

	// Load config from file.
	// You may rename "config.example.json" to "config.json" and modify it.
	// It looks like this:
	//{
	//    "access_key_id":"test_key_id",
	//    "access_key_secret":"test_key_secret",
	//    "sms": {
	//        "phone_numbers":["13800138000"],
	//        "sign_name":"测试签名",
	//        "template_code":"SMS_0000",
	//        "template_param":"{\"code\":\"888888\"}"
	//    },
	//    "single_call_by_tts": {
	//        "called_show_number":"025000000",
	//        "called_number":"13800138000",
	//        "template_code":"TTS_0000",
	//        "template_param":"{\"code\":\"888888\"}"
	//    }
	//}
	if err = loadConfig("config.json", &config); err != nil {
		log.Printf("loadConfig() error: %v", err)
		return
	}

	// Creates a new client.
	client := message.NewClient(config.AccessKeyID, config.AccessKeySecret)
	log.Printf("client: %v", client)

	// Send SMS.
	ok, smsResp, err := client.SendSMS(
		config.SMS.PhoneNumbers,
		config.SMS.SignName,
		config.SMS.TemplateCode,
		config.SMS.TemplateParam,
	)
	if err != nil {
		log.Printf("SendSMS() error: %v", err)
		return
	}
	log.Printf("SendSMS() ok: %v, response: %v", ok, smsResp)

	// Make Single Call by TTS.
	ok, vmsResp, err := client.MakeSingleCallByTTS(
		config.SingleCallByTTS.CalledShowNumber,
		config.SingleCallByTTS.CalledNumber,
		config.SingleCallByTTS.TemplateCode,
		config.SingleCallByTTS.TemplateParam,
	)
	if err != nil {
		log.Printf("MakeSingleCallByTTS() error: %v", err)
		return
	}
	log.Printf("MakeSingleCallByTTS() ok: %v, response: %v", ok, vmsResp)

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
