/* 
*Copyright (c) 2019-2021, Alibaba Group Holding Limited;
*Licensed under the Apache License, Version 2.0 (the "License");
*you may not use this file except in compliance with the License.
*You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*Unless required by applicable law or agreed to in writing, software
*distributed under the License is distributed on an "AS IS" BASIS,
*WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*See the License for the specific language governing permissions and
*limitations under the License.
 */


package bizapis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"

	"github.com/emicklei/go-restful"
	"github.com/pkg/errors"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business"
	"k8s.io/klog/klogr"
)

type bodyLogWriter struct {
	http.ResponseWriter
	Body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}

var logger = klogr.New().WithName("bizApis")

func logging(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	logger.Info("api request begin.", "method", req.Request.Method, "url", req.Request.URL.Path, "body", req.Request.Body)
	blw := &bodyLogWriter{Body: bytes.NewBufferString(""), ResponseWriter: resp.ResponseWriter}
	resp.ResponseWriter = blw
	chain.ProcessFilter(req, resp)
	respStr := blw.Body.String()
	respStr = strings.ReplaceAll(respStr, "\n", "")
	logger.Info("api request end.", "method", req.Request.Method, "url", req.Request.URL.Path, "statusCode", strconv.Itoa(resp.StatusCode()), "body", respStr)
}

func Start() {
	ws := new(restful.WebService)
	ws.Filter(logging)
	ws.Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("namespace/{namespace}/name/{name}/rw_ins").To(SetRw))
	ws.Route(ws.POST("namespace/{namespace}/name/{name}/meta/engine_status").To(SetEngineStatus))

	restful.Add(ws)
}

func SetRw(request *restful.Request, response *restful.Response) {
	requestLogger := logger.WithValues("url", request.Request.URL.Path)
	params, err := ParseInput(request, response, requestLogger, "endpoint")
	if err != nil {
		return
	}

	clusterName, clusterNameSpace := request.PathParameter("name"), request.PathParameter("namespace")

	tBegin := time.Now()
	defer func() {
		tEnd := time.Now()
		spend := tEnd.Sub(tBegin).Seconds()
		requestLogger.Info(fmt.Sprintf("deal request spend [%f] second", spend))
	}()

	endpoint := params[0]

	if endpoint == "" {
		errMsg := "endpoint should not be empty."
		requestLogger.Error(errors.New(errMsg), "")
		if err := response.WriteErrorString(http.StatusBadRequest, errMsg); err != nil {
			requestLogger.Error(err, "WriteErrorString failed")
		}
		return
	}

	service := business.NewSharedStorageClusterService(logger)
	cluster, err := service.GetByName(clusterName, clusterNameSpace)
	if err != nil {
		if err := response.WriteErrorString(http.StatusInternalServerError, err.Error()); err != nil {
			requestLogger.Error(err, "WriteErrorString failed")
		}
		return
	}
	if err = service.SetRw(context.TODO(), endpoint, cluster); err != nil {
		if err := response.WriteErrorString(http.StatusInternalServerError, err.Error()); err != nil {
			requestLogger.Error(err, "WriteErrorString failed")
		}
		return
	}
	if err := response.WriteEntity(ResponseSucInitFactory()); err != nil {
		requestLogger.Error(err, "WriteEntity failed")
	}
	return
}

func SetEngineStatus(request *restful.Request, response *restful.Response) {
	requestLogger := logger.WithValues("url", request.Request.URL.Path)
	params, err := ParseInput(request, response, requestLogger, "endpoint", "engineStatus", "reason")
	if err != nil {
		return
	}

	clusterName, clusterNameSpace := request.PathParameter("name"), request.PathParameter("namespace")

	tBegin := time.Now()
	defer func() {
		tEnd := time.Now()
		spend := tEnd.Sub(tBegin).Seconds()
		requestLogger.Info(fmt.Sprintf("deal request spend [%f] second", spend))
	}()

	endpoint := params[0]
	engineStatus := params[1]
	reason := params[2]

	if endpoint == "" || engineStatus == "" {
		errMsg := "endpoint and engineStatus should not be empty."
		requestLogger.Error(errors.New(errMsg), "")
		if err := response.WriteErrorString(http.StatusBadRequest, errMsg); err != nil {
			requestLogger.Error(err, "WriteErrorString failed")
		}
		return
	}

	service := business.NewSharedStorageClusterService(requestLogger)
	cluster, err := service.GetByName(clusterName, clusterNameSpace)
	if err != nil {
		if err := response.WriteErrorString(http.StatusInternalServerError, err.Error()); err != nil {
			requestLogger.Error(err, "WriteErrorString failed")
		}
		return
	}
	if err = service.SetInsState(context.TODO(), endpoint, engineStatus, "", reason, cluster); err != nil {
		if err := response.WriteErrorString(http.StatusInternalServerError, err.Error()); err != nil {
			requestLogger.Error(err, "WriteErrorString failed")
		}
		return
	}

	if err := response.WriteEntity(ResponseSucInitFactory()); err != nil {
		requestLogger.Error(err, "WriteEntity failed")
	}
	return
}

func ParseInput(request *restful.Request, response *restful.Response, requestLogger logr.Logger, params ...string) ([]string, error) {
	body, err := ReadHttpPostBody(request, response, requestLogger)
	if err != nil {
		return nil, err
	}
	requestLogger.Info("request body", "body", string(body))
	var input map[string]string
	err = json.Unmarshal(body, &input)
	if err != nil {
		if err := response.WriteErrorString(http.StatusBadRequest, err.Error()); err != nil {
			requestLogger.Error(err, "WriteErrorString failed")
		}
		return nil, err
	}
	var values []string
	for _, param := range params {
		if paramValue, ok := input[param]; !ok {
			if err := response.WriteErrorString(http.StatusBadRequest, fmt.Sprintf("please input param: %v", param)); err != nil {
				requestLogger.Error(err, "WriteErrorString failed")
			}
			return nil, errors.Errorf("input param: %v not exists", param)
		} else {
			values = append(values, paramValue)
		}
	}
	return values, nil
}

func ReadHttpPostBody(request *restful.Request, response *restful.Response, requestLogger logr.Logger) ([]byte, error) {
	var reader io.Reader = request.Request.Body
	maxFormSize := int64(1<<63 - 1)
	maxFormSizeReal := int64(10 << 20) // 10 MB is a lot of text.
	reader = io.LimitReader(reader, maxFormSizeReal+1)
	b, e := ioutil.ReadAll(reader)
	if e != nil {
		requestLogger.Error(e, "get post msg fail, %v")
		if err := response.WriteErrorString(http.StatusInternalServerError, e.Error()); err != nil {
			requestLogger.Error(err, "WriteErrorString failed")
		}
		return nil, e
	}
	if int64(len(b)) > maxFormSize {
		err := errors.New("http: POST too large")
		if err := response.WriteErrorString(http.StatusInternalServerError, err.Error()); err != nil {
			requestLogger.Error(err, "WriteErrorString failed")
		}
		return nil, e
	}
	return b, nil
}

type ResponseWithMsg struct {
	Status string                 `json:"status"`
	Msg    map[string]interface{} `json:"msg"`
}

func ResponseSucInitFactory() *ResponseWithMsg {
	return &ResponseWithMsg{Status: "ok", Msg: map[string]interface{}{}}
}
