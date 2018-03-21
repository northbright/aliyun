package sms

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/northbright/uuid"
)

type SMS struct {
	accessKeySecret string
	Params          map[string]string
}

type Param struct {
	f func(sms *SMS)
}

type Response struct {
	RequestID string `json:"RequestId"`
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	BizID     string `json:"BizId"`
}

func New(accessKeyID, accessKeySecret string) *SMS {
	sms := &SMS{accessKeySecret: accessKeySecret, Params: make(map[string]string)}
	sms.Params["AccessKeyId"] = accessKeyID
	return sms
}

func (sms *SMS) SetTimestamp(t time.Time) {
	gmt := t.UTC()
	sms.Params["Timestamp"] = fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ",
		gmt.Year(),
		gmt.Month(),
		gmt.Day(),
		gmt.Hour(),
		gmt.Minute(),
		gmt.Second(),
	)
}

func Timestamp(t time.Time) Param {
	return Param{f: func(sms *SMS) { sms.SetTimestamp(t) }}
}

func SignatureMethod(m string) Param {
	return Param{f: func(sms *SMS) { sms.Params["SignatureMethod"] = m }}
}

func SignatureVersion(v string) Param {
	return Param{f: func(sms *SMS) { sms.Params["SignatureVersion"] = v }}
}

func SignatureNonce(nonce string) Param {
	return Param{f: func(sms *SMS) { sms.Params["SignatureNonce"] = nonce }}
}

func Action(action string) Param {
	return Param{f: func(sms *SMS) { sms.Params["Action"] = action }}
}

func Version(v string) Param {
	return Param{f: func(sms *SMS) { sms.Params["Version"] = v }}
}

func RegionID(ID string) Param {
	return Param{f: func(sms *SMS) { sms.Params["RegionId"] = ID }}
}

func OutID(ID string) Param {
	return Param{f: func(sms *SMS) { sms.Params["OutId"] = ID }}
}

func (sms *SMS) SetPhoneNumbers(nums []string) {
	str := ""
	l := len(nums)
	for i, num := range nums {
		str += num
		if i != l-1 {
			str += ","
		}
	}
	sms.Params["PhoneNumbers"] = str
}

func SpecialURLEncode(str string) string {
	encodedStr := url.QueryEscape(str)
	encodedStr = strings.Replace(encodedStr, "+", "%20", -1)
	encodedStr = strings.Replace(encodedStr, "*", "%2A", -1)
	encodedStr = strings.Replace(encodedStr, "%7E", "~", -1)
	return encodedStr
}

func (sms *SMS) SortedQueryStr() string {
	values := url.Values{}
	for k, v := range sms.Params {
		values.Set(k, v)
	}
	// Encodes the values into “URL encoded” form ("bar=baz&foo=quux") sorted by key.
	return values.Encode()
}

func (sms *SMS) SignedString() string {
	str := "GET&" + url.QueryEscape("/") + "&" + SpecialURLEncode(sms.SortedQueryStr())

	// HMAC-SHA1
	mac := hmac.New(sha1.New, []byte(sms.accessKeySecret+"&"))
	mac.Write([]byte(str))

	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	sign = url.QueryEscape(sign)
	return SpecialURLEncode(sign)
}

func (sms *SMS) Send(phoneNumbers []string, signName, templateCode, templateParam string, params ...Param) (bool, *Response, error) {
	// Set default common parameters
	sms.SetTimestamp(time.Now())
	sms.Params["Format"] = "JSON"
	sms.Params["SignatureMethod"] = "HMAC-SHA1"
	sms.Params["SignatureVersion"] = "1.0"
	UUID, _ := uuid.New()
	sms.Params["SignatureNonce"] = UUID

	// Set default business parameters
	sms.Params["Action"] = "SendSms"
	sms.Params["Version"] = "2017-05-25"
	sms.Params["RegionId"] = "cn-hangzhou"

	// Override default parameters
	for _, param := range params {
		param.f(sms)
	}

	// Set required business parameters
	sms.SetPhoneNumbers(phoneNumbers)
	sms.Params["SignName"] = signName
	sms.Params["TemplateCode"] = templateCode
	sms.Params["TemplateParam"] = templateParam

	// Get signed string
	sign := sms.SignedString()
	sortedQueryStr := sms.SortedQueryStr()
	rawQuery := fmt.Sprintf("Signature=%s&%s", sign, sortedQueryStr)

	u := &url.URL{
		Scheme:   "http",
		Host:     "dysmsapi.aliyuncs.com",
		Path:     "/",
		RawQuery: rawQuery,
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return false, nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, nil, err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, nil, err
	}

	// Parse JSON response
	response := &Response{}
	if err = json.Unmarshal(buf, response); err != nil {
		return false, nil, err
	}

	if strings.ToUpper(response.Code) != "OK" {
		return false, response, nil
	}
	return true, response, nil
}
