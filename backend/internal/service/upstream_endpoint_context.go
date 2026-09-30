package service

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

// upstreamEndpointCtxKey 携带"本次上游请求实际打到的端点"（host+path）。
//
// 401/403 等错误路径会把它写进日志：同一个账号在不同接口上的拒绝原因完全不同
// （例如 chatgpt.com/backend-api/codex/responses 与 .../codex/models、
// .../codex/alpha/search 的鉴权口径不一致），没有端点信息就分不清是哪一处失败。
// 仅用于日志与排障，不参与任何业务判定。
type upstreamEndpointCtxKey struct{}

// withUpstreamEndpoint 在 ctx 上记录上游端点（空串视为未提供）。
func withUpstreamEndpoint(ctx context.Context, endpoint string) context.Context {
	if ctx == nil {
		return nil
	}
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ctx
	}
	return context.WithValue(ctx, upstreamEndpointCtxKey{}, endpoint)
}

// upstreamEndpointFromContext 返回本次请求的上游端点，未记录时返回空串。
func upstreamEndpointFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	endpoint, _ := ctx.Value(upstreamEndpointCtxKey{}).(string)
	return endpoint
}

// upstreamEndpointLabel 返回可直接写进日志的端点标签。
func upstreamEndpointLabel(ctx context.Context) string {
	if endpoint := upstreamEndpointFromContext(ctx); endpoint != "" {
		return endpoint
	}
	return "unknown_endpoint"
}

// upstreamEndpointOfURL 把完整 URL 转成与 upstreamEndpointOf 一致的 host+path 标签。
// 用于只有端点常量、拿不到 *http.Response 的错误路径。
func upstreamEndpointOfURL(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed == nil {
		return ""
	}
	host := strings.TrimSpace(parsed.Host)
	path := strings.TrimRight(strings.TrimSpace(parsed.Path), "/")
	switch {
	case host != "" && path != "":
		return host + path
	case host != "":
		return host
	default:
		return path
	}
}

// upstreamEndpointOf 从上游响应中取出实际请求的 host+path；缺失时返回空串。
func upstreamEndpointOf(resp *http.Response) string {
	if resp == nil || resp.Request == nil || resp.Request.URL == nil {
		return ""
	}
	host := strings.TrimSpace(resp.Request.URL.Host)
	path := strings.TrimRight(strings.TrimSpace(resp.Request.URL.Path), "/")
	switch {
	case host != "" && path != "":
		return host + path
	case host != "":
		return host
	default:
		return path
	}
}
