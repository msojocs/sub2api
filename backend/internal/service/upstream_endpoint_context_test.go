//go:build unit

package service

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpstreamEndpointLabel(t *testing.T) {
	t.Run("nil context falls back to unknown_endpoint", func(t *testing.T) {
		require.Equal(t, "unknown_endpoint", upstreamEndpointLabel(nil))
	})

	t.Run("empty context falls back to unknown_endpoint", func(t *testing.T) {
		require.Equal(t, "unknown_endpoint", upstreamEndpointLabel(context.Background()))
	})

	t.Run("blank endpoint is not recorded", func(t *testing.T) {
		ctx := withUpstreamEndpoint(context.Background(), "   ")
		require.Equal(t, "unknown_endpoint", upstreamEndpointLabel(ctx))
	})

	t.Run("endpoint is trimmed and recorded", func(t *testing.T) {
		ctx := withUpstreamEndpoint(context.Background(), "  chatgpt.com/backend-api/codex/responses  ")
		require.Equal(t, "chatgpt.com/backend-api/codex/responses", upstreamEndpointLabel(ctx))
	})
}

func TestWithUpstreamEndpointNilContext(t *testing.T) {
	// 上游错误路径可能拿到 nil ctx（例如流式失败回退到 context.Background 之前），
	// 这里必须原样返回 nil，不能 panic。
	//nolint:staticcheck // 故意传入 nil context 验证防御性行为
	require.Nil(t, withUpstreamEndpoint(nil, "chatgpt.com/backend-api/codex/responses"))
}

func TestUpstreamEndpointOf(t *testing.T) {
	t.Run("nil response", func(t *testing.T) {
		require.Empty(t, upstreamEndpointOf(nil))
	})

	t.Run("response without request", func(t *testing.T) {
		require.Empty(t, upstreamEndpointOf(&http.Response{}))
	})

	t.Run("response without url", func(t *testing.T) {
		require.Empty(t, upstreamEndpointOf(&http.Response{Request: &http.Request{}}))
	})

	t.Run("host and path", func(t *testing.T) {
		resp := newUpstreamResponseForTest(t, "https://chatgpt.com/backend-api/codex/responses")
		require.Equal(t, "chatgpt.com/backend-api/codex/responses", upstreamEndpointOf(resp))
	})

	t.Run("trailing slash trimmed", func(t *testing.T) {
		resp := newUpstreamResponseForTest(t, "https://chatgpt.com/backend-api/codex/models/")
		require.Equal(t, "chatgpt.com/backend-api/codex/models", upstreamEndpointOf(resp))
	})

	t.Run("query string is dropped", func(t *testing.T) {
		resp := newUpstreamResponseForTest(t, "https://chatgpt.com/backend-api/codex/responses?stream=true")
		require.Equal(t, "chatgpt.com/backend-api/codex/responses", upstreamEndpointOf(resp))
	})
}

func TestUpstreamEndpointOfURL(t *testing.T) {
	t.Run("matches response derived label", func(t *testing.T) {
		raw := "https://chatgpt.com/backend-api/codex/models/"
		resp := newUpstreamResponseForTest(t, raw)
		require.Equal(t, upstreamEndpointOf(resp), upstreamEndpointOfURL(raw))
	})

	t.Run("invalid url yields empty", func(t *testing.T) {
		require.Empty(t, upstreamEndpointOfURL("://not-a-url"))
	})

	t.Run("blank url yields empty", func(t *testing.T) {
		require.Empty(t, upstreamEndpointOfURL("   "))
	})

	t.Run("codex models manifest constant", func(t *testing.T) {
		require.Equal(t, "chatgpt.com/backend-api/codex/models", upstreamEndpointOfURL(chatgptCodexModelsURL))
	})
}

func newUpstreamResponseForTest(t *testing.T, rawURL string) *http.Response {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)
	return &http.Response{Request: &http.Request{URL: parsed}}
}
