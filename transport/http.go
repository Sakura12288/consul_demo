package transport

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-kit/kit/log"
	kittransport "github.com/go-kit/kit/transport"
	mymux "github.com/gorilla/mux"

	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
	endpts "service/endpoint"
)

var (
	ErrBadRequest = errors.New("无效的请求参数")
)

func decodeSayHelloRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return endpts.SayHelloRequest{}, nil
}

func decodeDiscoveryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	svcName := r.URL.Query().Get("service_name")
	if svcName == "" {
		return nil, ErrBadRequest
	}
	return endpts.DiscoveryRequest{ServiceName: svcName}, nil
}

func decodeHealthRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return endpts.HealthCheckRequest{}, nil
}

// 编码内容
func encodeJsonResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	err := json.NewEncoder(w).Encode(response)
	return err
}

// 编码错误
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]string{
		"errors": err.Error(),
	})
}

func MakeHttpHandler(ctx context.Context, endpoints endpts.DiscoveryEndpoints, logger log.Logger) http.Handler {
	r := mymux.NewRouter()

	//定义处理处理器
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(kittransport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}
	//say-hello 接口
	r.Methods("GET").Path("/say-hello").Handler(httptransport.NewServer(endpoints.SayHelloEndpoint,
		decodeSayHelloRequest, encodeJsonResponse, options...))

	//服务发现 接口
	r.Methods("GET").Path("/discovery").Handler(httptransport.NewServer(endpoints.DiscoveryEndpoint,
		decodeDiscoveryRequest, encodeJsonResponse, options...))

	//健康检查 接口
	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(endpoints.HealthCheckEndpoint,
		decodeHealthRequest, encodeJsonResponse, options...))

	return r
}
