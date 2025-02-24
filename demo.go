package pluginTraefik

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Config 是插件的配置结构体
type Config struct {
	MetricName string `json:"metricName"` // 指标名称
}

// CreateConfig 创建默认配置
func CreateConfig() *Config {
	return &Config{
		MetricName: "traefik_url_requests_total", // 默认指标名称
	}
}

// URLMetrics 是插件的主要结构体
type URLMetrics struct {
	next          http.Handler
	name          string
	config        *Config
	requestsTotal *prometheus.CounterVec // Prometheus 计数器
	once          sync.Once              // 用于确保指标只初始化一次
}

// New 创建插件实例
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &URLMetrics{
		next:   next,
		name:   name,
		config: config,
	}, nil
}

// ServeHTTP 处理 HTTP 请求
func (m *URLMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 确保 Prometheus 指标只初始化一次
	m.once.Do(func() {
		m.requestsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: m.config.MetricName,
				Help: "Total number of HTTP requests by URL",
			},
			[]string{"url", "method", "status"}, // 标签：URL、HTTP 方法、状态码
		)
	})

	// 调用下一个处理器
	m.next.ServeHTTP(w, r)

	// 记录请求的 URL、方法和状态码
	m.requestsTotal.WithLabelValues(normalizeUri(safeLabel(r.URL.Path)), r.Method, r.Response.Status).Inc()
}

// normalizeUri 标准化 URI，将动态部分替换为占位符
func normalizeUri(uri string) string {
	re := regexp.MustCompile("/\\d+")
	return re.ReplaceAllString(uri, "/{id}")
}

// safeLabel 处理标签值，使其安全并符合 Prometheus 标签规范
func safeLabel(value string) string {
	// 替换动态路径
	re := regexp.MustCompile("/\\d+")
	value = re.ReplaceAllString(value, "/:id")

	reUUID := regexp.MustCompile("/[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}")
	value = reUUID.ReplaceAllString(value, "/:uuid")

	// 压缩重复下划线
	reDoubleUnderscore := regexp.MustCompile("_{2,}")
	value = reDoubleUnderscore.ReplaceAllString(value, "_")

	// 白名单过滤 + 截断
	value = regexp.MustCompile("[^a-zA-Z0-9_:/]").ReplaceAllString(value, "_")
	value = strings.Trim(value, "_")

	return value
}
