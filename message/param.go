package message

import (
	"fmt"
	"net/url"
	"time"
)

// Param is the parameter for HTTP request of aliyun API.
// Use param helper functions to get specified Param. e.g. Timestamp(), SignatureNonce().
type Param struct {
	f func(v url.Values)
}

// Timestamp specifies the timestamp.
// aliyun requires GMT but not local time.
// It will generate timestamp automatically by default if no one specifed.
func Timestamp(t time.Time) Param {
	return Param{f: func(v url.Values) {
		v.Set("Timestamp", GenTimestamp(t))
	}}
}

// SignatureMethod specifies the signature method.
// It's "HMAC-SHA1" by default if no one specifed.
func SignatureMethod(m string) Param {
	return Param{f: func(v url.Values) { v.Set("SignatureMethod", m) }}
}

// SignatureVersion specifies the signature version.
// It's "1.0" by default if no one specifed.
func SignatureVersion(ver string) Param {
	return Param{f: func(v url.Values) { v.Set("SignatureVersion", ver) }}
}

// SignatureNonce specifies the nonce.
// It will generate UUID as nonce automatically by default if no one specified.
func SignatureNonce(nonce string) Param {
	return Param{f: func(v url.Values) { v.Set("SignatureNonce", nonce) }}
}

// Action specifies the action.
// It's "SendSms" by default if no one specified.
func Action(action string) Param {
	return Param{f: func(v url.Values) { v.Set("Action", action) }}
}

// Version specifies the version.
// It's "2017-05-25" by default if no one specified.
func Version(ver string) Param {
	return Param{f: func(v url.Values) { v.Set("Version", ver) }}
}

// RegionID specifies the region ID.
// It's "cn-hangzhou" by default if no one specified.
func RegionID(ID string) Param {
	return Param{f: func(v url.Values) { v.Set("RegionId", ID) }}
}

// PhoneNumbers specifies the phone numbers to send SMS.
func PhoneNumbers(nums []string) Param {
	return Param{f: func(v url.Values) {
		v.Set("PhoneNumbers", GenPhoneNumbersStr(nums))
	}}
}

// GenTimestamp generates the timestamp for aliyun services.
// aliyun requires GMT but not local time.
func GenTimestamp(t time.Time) string {
	gmt := t.UTC()
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ",
		gmt.Year(),
		gmt.Month(),
		gmt.Day(),
		gmt.Hour(),
		gmt.Minute(),
		gmt.Second(),
	)
}

// GenPhoneNumbersStr generates the parameter string for one or more phone numbers.
// Delimeter is ",".
func GenPhoneNumbersStr(nums []string) string {
	str := ""
	l := len(nums)
	for i, num := range nums {
		str += num
		if i != l-1 {
			str += ","
		}
	}
	return str
}
