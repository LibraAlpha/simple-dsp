package rta

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"simple-dsp/internal/rta"
)

// MockServer 模拟RTA服务器
type MockServer struct {
	server *httptest.Server
}

// NewMockServer 创建新的模拟服务器
func NewMockServer() *MockServer {
	ms := &MockServer{}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/single", ms.handleSingleQuery)
	mux.HandleFunc("/api/batch", ms.handleBatchQuery)

	ms.server = httptest.NewServer(mux)
	return ms
}

// URL 返回模拟服务器的URL
func (ms *MockServer) URL() string {
	return ms.server.URL
}

// Close 关闭模拟服务器
func (ms *MockServer) Close() {
	ms.server.Close()
}

// handleSingleQuery 处理单次查询请求
func (ms *MockServer) handleSingleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求参数
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	// 检查必要参数
	if r.Form.Get("channel") == "" || r.Form.Get("ad_space_id") == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// 模拟响应
	resp := &rta.SingleResponse{
		ErrCode: rta.ErrCodeSuccess,
		Result:  true,
		TaskID:  "mock_task_123",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleBatchQuery 处理批量查询请求
func (ms *MockServer) handleBatchQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求参数
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	// 检查必要参数
	if r.Form.Get("channel") == "" || r.Form.Get("ad_space_id") == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// 检查设备ID列表
	var deviceIDs []string
	if ids := r.Form.Get("imei_md5"); ids != "" {
		deviceIDs = strings.Split(ids, ",")
	} else if ids := r.Form.Get("idfa_md5"); ids != "" {
		deviceIDs = strings.Split(ids, ",")
	} else if ids := r.Form.Get("oaid_md5"); ids != "" {
		deviceIDs = strings.Split(ids, ",")
	}

	if len(deviceIDs) == 0 {
		http.Error(w, "No device IDs provided", http.StatusBadRequest)
		return
	}

	if len(deviceIDs) > 20 {
		http.Error(w, "Too many device IDs", http.StatusBadRequest)
		return
	}

	// 模拟响应
	results := make([]rta.BatchResult, 0, len(deviceIDs))
	for _, id := range deviceIDs {
		results = append(results, rta.BatchResult{
			TaskID:  "mock_task_" + id,
			IMEIMD5: id,
		})
	}

	resp := &rta.BatchResponse{
		ErrCode: rta.ErrCodeSuccess,
		Results: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
