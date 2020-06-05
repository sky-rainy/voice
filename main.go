package main

import (
	"time"
	"fmt"
	"strings"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"net"
	"net/url"
	"strconv"
	"crypto/tls"
)

const POST_REMOTE_TIMEOUT = 30

const (
	// APP相关的两个KEY,需要从百度AI开放平台申请
	API_KEY = ""
	SECRET_KEY = ""

	// 发音人选择, 0为普通女声，1为普通男生，3为情感合成-度逍遥，4为情感合成-度丫丫，默认为普通女声
	PER = 4
	// 语速，取值0-15，默认为5中语速
	SPD = 5
	// 音调，取值0-15，默认为5中语调
	PIT = 5
	// 音量，取值0-9，默认为5中音量
	VOL = 9
	// 下载的文件格式, 3：mp3(default) 4： pcm-16k 5： pcm-8k 6. wav
	AUE = 3

	CUID = "123456PYTHON"
	TTS_URL = "http://tsn.baidu.com/text2audio"
	TOKEN_URL = "http://openapi.baidu.com/oauth/2.0/token"
)

type CommResp struct {
	AccessToken     string    `json:"access_token"`
	SessionKey      string    `json:"session_key"`
	Scope           string    `json:"scope"`
	RefreshToken    string    `json:"refresh_token"`
	SessionSecret   string    `json:"session_secret"`
	ExpiresIn       string    `json:"expires_in"`
}

func main() {
	token := fetch_token()

	tex := "嘟嘟嘟，嘟嘟嘟！"

	var format_map = make(map[int]string, 0)
	format_map[3] = "mp3"
	format_map[4] = "pcm"
	format_map[5] = "pcm"
	format_map[6] = "wav"

	params := make(map[string]interface{}, 0)
	params["tok"] = token
	params["tex"] = tex
	params["per"] = PER
	params["spd"] = SPD
	params["pit"] = PIT
	params["vol"] = VOL
	params["aue"] = AUE
	params["cuid"] = CUID
	params["lan"] = "zh"
	params["ctp"] = 1

	response := sendRequest(TTS_URL, params, "GET")
	name := fmt.Sprintf("result.%v", format_map[AUE])
	WriteWithIoutil(name, response)
}
// 获取百度TOKEN
func fetch_token() string {
	params := make(map[string]interface{}, 0)
	params["grant_type"] = "client_credentials"
	params["client_id"] = API_KEY
	params["client_secret"] = SECRET_KEY

	response := sendRequest(TOKEN_URL, params, "GET")

	var res CommResp
	if err := json.Unmarshal([]byte(response), &res); err == nil {
		fmt.Printf("res=[%v]\n", res)
	}

	return res.AccessToken
}

func WriteWithIoutil(name, content string) {
	data :=  []byte(content)
	if ioutil.WriteFile("./mp3/"+name,data,0644) == nil {
		fmt.Println("生成音频文件成功:", name)
	}
}
// post 设置请求超时
func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, time.Second*POST_REMOTE_TIMEOUT)
}
// 转换类型
func typeSwitcher(t interface{}) string {
	switch v := t.(type) {
	case int:
		return strconv.Itoa(v)
	case string:
		return v
	case int64:
		return strconv.Itoa(int(v))
	case []string:
		return "typeArray"
	case map[string]interface{}:
		return "typeMap"
	default:
		return ""
	}
}
// get 拼接参数
func ParamsToStr(params map[string]interface{}) string {
	isfirst := true
	requesturl := ""
	for k, v := range params {
		if !isfirst {
			requesturl = requesturl + "&"
		}

		isfirst = false
		if strings.Contains(k, "_") {
			strings.Replace(k, ".", "_", -1)
		}
		v := typeSwitcher(v)
		requesturl = requesturl + k + "=" + url.QueryEscape(v)
	}

	return requesturl
}
// get请求
func httpGet(url string) string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr, Timeout: time.Duration(3) * time.Second}
	fmt.Println(url)
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}

	defer resp.Body.Close()
	body, erro := ioutil.ReadAll(resp.Body)
	if erro != nil {
		fmt.Println("http wrong erro")
		return erro.Error()
	}

	return string(body)
}
// post请求
func httpPost(requesturl string, params map[string]interface{}) string {
	b, err := json.Marshal(params)
	if err != nil {
		fmt.Errorf("json.Marshal failed[%v]", err)
		return err.Error()
	}

	req, err := http.NewRequest("POST", requesturl, strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")

	transport := http.Transport {
		Dial: dialTimeout,
		DisableKeepAlives: true,
	}

	client := &http.Client{Transport: &transport, Timeout: time.Duration(30) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}

	defer resp.Body.Close()
	body, erro := ioutil.ReadAll(resp.Body)
	if erro != nil {
		fmt.Println("http wrong erro")
		return erro.Error()
	}

	return string(body)
}
// 集合get或post请求方式
func sendRequest(requesturl string, params map[string]interface{}, method string) string {
	response := ""
	if method == "GET" {
		params_str := "?" + ParamsToStr(params)
		requesturl = requesturl + params_str
		response = httpGet(requesturl)
	} else if method == "POST" {
		response = httpPost(requesturl, params)
	} else {
		fmt.Println("unsuppported http method")
		return "unsuppported http method"
	}

	return response
}