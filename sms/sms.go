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

// Client is the SMS client.
// A client should be resused to send SMS.
type Client struct {
	// Use http.Client.Do().
	http.Client
	// accessKeySecret is the access key secrete generated by user.
	accessKeySecret string
	// Params contains parameters used for HTTP request of sending SMS.
	Params map[string]string
}

// Param is the parameter for HTTP request of sending SMS.
// Use param helper functions to get specified Param. e.g. Timestamp(), SignatureNonce().
type Param struct {
	f func(c *Client)
}

// Response is the response of HTTP request of sending SMS.
type Response struct {
	// RequestID is the request ID. e.g. "8906582E-6722".
	RequestID string `json:"RequestId"`
	// Code is the status code. e.g. "OK", "SignatureDoesNotMatch".
	Code string `json:"Code"`
	// Message is the detail message for the status code. e.g. "OK", Specified signature is not matched with our calculation...".
	Message string `json:"Message"`
	// BizID is the business ID. It can be used to query the status of SMS. e.g. "134523^4351232".
	BizID string `json:"BizId"`
}

// NewClient creates a new client to send SMS.
//
// It accepts 2 parameters: access key ID and secret.
// Both of them are generated by user in aliyun control panel.
func NewClient(accessKeyID, accessKeySecret string) *Client {
	c := &Client{accessKeySecret: accessKeySecret, Params: make(map[string]string)}
	c.Params["AccessKeyId"] = accessKeyID
	return c
}

// SetTimestamp sets the timestamp parameter.
// aliyun requires GMT but not local time.
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

// Timestamp specifies the timestamp.
// aliyun requires GMT but not local time.
// Send() will generate timestamp automatically.
// You may also use your own timestamp and pass it to Send().
func Timestamp(t time.Time) Param {
	return Param{f: func(c *Client) { c.SetTimestamp(t) }}
}

// SignatureMethod specifies the signature method.
// It's "HMAC-SHA1" by default if no one specifed.
func SignatureMethod(m string) Param {
	return Param{f: func(c *Client) { c.Params["SignatureMethod"] = m }}
}

// SignatureVersion specifies the signature version.
// It's "1.0" by default if no one specifed.
func SignatureVersion(v string) Param {
	return Param{f: func(c *Client) { c.Params["SignatureVersion"] = v }}
}

// SignatureNonce specifies the nonce.
// Send() will generate UUID as nonce automatically.
// You may also use your own nonce and pass it to Send().
func SignatureNonce(nonce string) Param {
	return Param{f: func(c *Client) { c.Params["SignatureNonce"] = nonce }}
}

// Action specifies the action.
// It's "SendSms" by default if no one specified.
func Action(action string) Param {
	return Param{f: func(c *Client) { c.Params["Action"] = action }}
}

// Version specifies the version.
// It's "2017-05-25" by default if no one specified.
func Version(v string) Param {
	return Param{f: func(c *Client) { c.Params["Version"] = v }}
}

// RegionID specifies the region ID.
// It's "cn-hangzhou" by default if no one specified.
func RegionID(ID string) Param {
	return Param{f: func(c *Client) { c.Params["RegionId"] = ID }}
}

// OutID specifies the out ID.
func OutID(ID string) Param {
	return Param{f: func(c *Client) { c.Params["OutId"] = ID }}
}

// SetPhoneNumbers set phone numbers to send SMS.
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

// SpecialURLEncode follows aliyun's POP protocol to do special URL encoding.
func SpecialURLEncode(str string) string {
	encodedStr := url.QueryEscape(str)
	encodedStr = strings.Replace(encodedStr, "+", "%20", -1)
	encodedStr = strings.Replace(encodedStr, "*", "%2A", -1)
	encodedStr = strings.Replace(encodedStr, "%7E", "~", -1)
	return encodedStr
}

// SortedQueryStr gets the query string sorted by keys.
func (c *Client) SortedQueryStr() string {
	values := url.Values{}
	for k, v := range c.Params {
		values.Set(k, v)
	}
	// Encodes the values into “URL encoded” form ("bar=baz&foo=quux") sorted by key.
	return values.Encode()
}

// SignedString follow aliyun's POP protocol to generate the signature.
func (c *Client) SignedString() string {
	str := "GET&" + url.QueryEscape("/") + "&" + SpecialURLEncode(c.SortedQueryStr())

	// HMAC-SHA1
	// aliyun requires appending "&" after access key secret.
	mac := hmac.New(sha1.New, []byte(c.accessKeySecret+"&"))
	mac.Write([]byte(str))

	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return SpecialURLEncode(sign)
}

// Send sends the SMS to phone numbers.
//
// phoneNumbers: one or more phone numbers. aliyun recommends to send SMS to only one phone number once for validation code.
// signName: permitted signature name. You may apply one ore more signature names in aliyun's control panel.
// templateCode: permitted template code. You may apply one or more template code in aliyun's control panel.
// templateParam: JSON to render the template. e.g. {"code":"1234","product":"ytx"}.
// params: optional parameters for sending SMS. In most case, no need to pass params.
// You may also specify params by helper functions. e.g. Timestamp(), SignatureNonce().
//
// It returns success status, response and error.
//
// For example:
//
// c := sms.NewClient(accessKeyID, accessKeySecret)
//
// ok, resp, err := c.Send([]string{"13800138000"}, "my_product", "SMS_0000", `{"code":"1234","product":"ytx"}`)
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
