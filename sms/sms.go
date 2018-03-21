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

type Client struct {
	http.Client
	accessKeySecret string
	Params          map[string]string
}

type Param struct {
	f func(c *Client)
}

type Response struct {
	RequestID string `json:"RequestId"`
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	BizID     string `json:"BizId"`
}

func NewClient(accessKeyID, accessKeySecret string) *Client {
	c := &Client{accessKeySecret: accessKeySecret, Params: make(map[string]string)}
	c.Params["AccessKeyId"] = accessKeyID
	return c
}

func (c *Client) SetTimestamp(t time.Time) {
	gmt := t.UTC()
	c.Params["Timestamp"] = fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ",
		gmt.Year(),
		gmt.Month(),
		gmt.Day(),
		gmt.Hour(),
		gmt.Minute(),
		gmt.Second(),
	)
}

func Timestamp(t time.Time) Param {
	return Param{f: func(c *Client) { c.SetTimestamp(t) }}
}

func SignatureMethod(m string) Param {
	return Param{f: func(c *Client) { c.Params["SignatureMethod"] = m }}
}

func SignatureVersion(v string) Param {
	return Param{f: func(c *Client) { c.Params["SignatureVersion"] = v }}
}

func SignatureNonce(nonce string) Param {
	return Param{f: func(c *Client) { c.Params["SignatureNonce"] = nonce }}
}

func Action(action string) Param {
	return Param{f: func(c *Client) { c.Params["Action"] = action }}
}

func Version(v string) Param {
	return Param{f: func(c *Client) { c.Params["Version"] = v }}
}

func RegionID(ID string) Param {
	return Param{f: func(c *Client) { c.Params["RegionId"] = ID }}
}

func OutID(ID string) Param {
	return Param{f: func(c *Client) { c.Params["OutId"] = ID }}
}

func (c *Client) SetPhoneNumbers(nums []string) {
	str := ""
	l := len(nums)
	for i, num := range nums {
		str += num
		if i != l-1 {
			str += ","
		}
	}
	c.Params["PhoneNumbers"] = str
}

func SpecialURLEncode(str string) string {
	encodedStr := url.QueryEscape(str)
	encodedStr = strings.Replace(encodedStr, "+", "%20", -1)
	encodedStr = strings.Replace(encodedStr, "*", "%2A", -1)
	encodedStr = strings.Replace(encodedStr, "%7E", "~", -1)
	return encodedStr
}

func (c *Client) SortedQueryStr() string {
	values := url.Values{}
	for k, v := range c.Params {
		values.Set(k, v)
	}
	// Encodes the values into “URL encoded” form ("bar=baz&foo=quux") sorted by key.
	return values.Encode()
}

func (c *Client) SignedString() string {
	str := "GET&" + url.QueryEscape("/") + "&" + SpecialURLEncode(c.SortedQueryStr())

	// HMAC-SHA1
	mac := hmac.New(sha1.New, []byte(c.accessKeySecret+"&"))
	mac.Write([]byte(str))

	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return SpecialURLEncode(sign)
}

func (c *Client) Send(phoneNumbers []string, signName, templateCode, templateParam string, params ...Param) (bool, *Response, error) {
	// Set default common parameters
	c.SetTimestamp(time.Now())
	c.Params["Format"] = "JSON"
	c.Params["SignatureMethod"] = "HMAC-SHA1"
	c.Params["SignatureVersion"] = "1.0"
	UUID, _ := uuid.New()
	c.Params["SignatureNonce"] = UUID

	// Set default business parameters
	c.Params["Action"] = "SendSms"
	c.Params["Version"] = "2017-05-25"
	c.Params["RegionId"] = "cn-hangzhou"

	// Override default parameters
	for _, param := range params {
		param.f(c)
	}

	// Set required business parameters
	c.SetPhoneNumbers(phoneNumbers)
	c.Params["SignName"] = signName
	c.Params["TemplateCode"] = templateCode
	c.Params["TemplateParam"] = templateParam

	// Get signature
	sign := c.SignedString()

	// Get query string
	sortedQueryStr := c.SortedQueryStr()

	// Make final query string with signature
	rawQuery := fmt.Sprintf("Signature=%s&%s", sign, sortedQueryStr)

	// New a URL with host, raw query
	u := &url.URL{
		Scheme:   "http",
		Host:     "dysmsapi.aliyuncs.com",
		Path:     "/",
		RawQuery: rawQuery,
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return false, nil, err
	}
	resp, err := c.Do(req)
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
