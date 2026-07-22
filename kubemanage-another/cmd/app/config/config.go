package config

// SysConfig 系统配置，全局对象
var SysConfig *Config

// Config 配置对象
type Config struct {
	Default DefaultOptions `mapstructure:"default"`
	Mysql   MysqlOptions   `mapstructure:"mysql"`
	Log     LogConfig      `mapstructure:"log"`
	Grafana GrafanaOptions `mapstructure:"grafana"`
}

// GrafanaOptions Grafana 单点登录/反向代理配置
type GrafanaOptions struct {
	// Upstream 集群内 Grafana 地址（已开启子路径），后端反代目标
	Upstream string `mapstructure:"upstream"`
	// DefaultRole 未在 RoleMapping 命中的角色使用的默认 Grafana 角色；为空则拒绝访问
	DefaultRole string `mapstructure:"defaultRole"`
	// RoleMapping Kubemanage 角色ID(字符串) -> Grafana Org 角色(Viewer/Editor/Admin)
	RoleMapping map[string]string `mapstructure:"roleMapping"`
	// TicketTTL 一次性票据有效期（秒）
	TicketTTL int `mapstructure:"ticketTTL"`
	// SessionTTL 代理会话 Cookie 有效期（秒）
	SessionTTL int `mapstructure:"sessionTTL"`
}

// DefaultOptions 默认配置选项
type DefaultOptions struct {
	PodLogTailLine       string `mapstructure:"podLogTailLine"`
	ListenAddr           string `mapstructure:"listenAddr"`
	WebSocketListenAddr  string `mapstructure:"webSocketListenAddr"`
	JWTSecret            string `mapstructure:"JWTSecret"`
	ExpireTime           int64  `mapstructure:"expireTime"`
	KubernetesConfigFile string `mapstructure:"kubernetesConfigFile"`
}

// MysqlOptions mysql配置选项
type MysqlOptions struct {
	Host         string `mapstructure:"host"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Port         string `mapstructure:"port"`
	Name         string `mapstructure:"name"`
	MaxOpenConns int    `mapstructure:"maxOpenConns"`
	MaxLifetime  int    `mapstructure:"maxLifetime"`
	MaxIdleConns int    `mapstructure:"maxIdleConns"`
}

// LogConfig 日志配置选项
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}
