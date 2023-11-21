package tools

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HttpRequest(url string, method string, postParams []byte, headers map[string]string) ([]byte, error) {
	httpClient := &http.Client{}
	var reader io.Reader
	if len(postParams) > 0 {
		reader = strings.NewReader(string(postParams))
		if headers == nil {
			headers = map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
		}
	} else {
		reader = nil
	}

	//构建request
	request, err := http.NewRequest(method, url, reader)
	if nil != err {
		return nil, fmt.Errorf("NewRequest error. %v", err)
	}

	//添加header
	for key, value := range headers {
		request.Header.Add(key, value)
	}

	// 发出请求
	response, err := httpClient.Do(request)
	if nil != err {
		return nil, fmt.Errorf("do the request error. %v", err)
	}

	defer response.Body.Close()

	// 解析响应内容
	body, err := io.ReadAll(response.Body)
	if nil != err {
		return nil, fmt.Errorf("ReadAll response.Body error. %v", err)
	}

	return body, nil
}
