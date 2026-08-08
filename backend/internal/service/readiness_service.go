package service

import (
	"context"
	"sort"
)

// ReadinessQuery 是就绪检查所需的最小模型端口。
type ReadinessQuery interface {
	Check(context.Context) map[string]error
}

// DependencyStatus 是依赖健康状态响应 DTO。
type DependencyStatus struct {
	Name  string `json:"name"`            // 依赖名称。
	Ready bool   `json:"ready"`           // 依赖是否可用。
	Error string `json:"error,omitempty"` // 不可用时的通用错误标识。
}

// ReadinessResponse 是服务就绪状态响应 DTO。
type ReadinessResponse struct {
	Ready        bool               `json:"ready"`        // 所有依赖是否均可用。
	Dependencies []DependencyStatus `json:"dependencies"` // 各依赖的检查结果。
}

// ReadinessService 负责将基础设施错误转换为安全的就绪响应。
type ReadinessService struct{ repo ReadinessQuery }

// NewReadinessService 创建就绪检查服务。
func NewReadinessService(repo ReadinessQuery) *ReadinessService { return &ReadinessService{repo: repo} }

func (s *ReadinessService) Check(ctx context.Context) ReadinessResponse {
	checks := s.repo.Check(ctx)
	names := make([]string, 0, len(checks))
	for name := range checks {
		names = append(names, name)
	}
	sort.Strings(names)
	ready := true
	dependencies := make([]DependencyStatus, 0, len(names))
	for _, name := range names {
		err := checks[name]
		item := DependencyStatus{Name: name, Ready: err == nil}
		if err != nil {
			// 就绪接口面向编排系统，避免把数据库地址等底层细节返回给外部调用方。
			item.Error = "unavailable"
			ready = false
		}
		dependencies = append(dependencies, item)
	}
	return ReadinessResponse{Ready: ready, Dependencies: dependencies}
}
