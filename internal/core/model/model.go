package model

import (
	"time"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"
	go_core_observ "github.com/eliezerraj/go-core/observability" 
)

type AppServer struct {
	InfoPod 		*InfoPod 					`json:"info_pod"`
	Server     		*Server     				`json:"server"`
	ConfigOTEL		*go_core_observ.ConfigOTEL	`json:"otel_config"`
	DatabaseConfig	*go_core_pg.DatabaseConfig  `json:"database"`
}

type InfoPod struct {
	PodName				string 	`json:"pod_name"`
	ApiVersion			string 	`json:"version"`
	OSPID				string 	`json:"os_pid"`
	IPAddress			string 	`json:"ip_address"`
	AvailabilityZone 	string 	`json:"availabilityZone"`
	IsAZ				bool   	`json:"is_az"`
	Env					string `json:"enviroment,omitempty"`
	OtelTraces			bool   	`json:"otel_traces"`
	OtelMetrics			bool   	`json:"otel_metrics"`
	OtelLogs			bool   	`json:"otel_logs"`
	AccountID			string `json:"account_id,omitempty"`
}

type Server struct {
	Port 			int `json:"port"`
	ReadTimeout		int `json:"readTimeout"`
	WriteTimeout	int `json:"writeTimeout"`
	IdleTimeout		int `json:"idleTimeout"`
	CtxTimeout		int `json:"ctxTimeout"`
}

type MessageRouter struct {
	Message			string `json:"message"`
}

type TypeLimit struct {
	Code			string 		`json:"code,omitempty"`
	Category		string 		`json:"category,omitempty"`
	CreateAt		time.Time 	`json:"created_at,omitempty"`	
}

type OrderLimit struct {
	TypeLimit		string 		`json:"type_limit,omitempty"`
	CounterLimit	string 		`json:"counter_limit,omitempty"`
	Type			string 		`json:"type,omitempty"`
	Amount			int 		`json:"amount,omitempty"`
	CreateAt		time.Time 	`json:"created_at,omitempty"`	
}

type Limit struct {
	TransactionId	string 		`json:"transaction_id,omitempty"`
	Key				string 		`json:"key,omitempty"`
	TypeLimit		string 		`json:"type_limit,omitempty"`
	OrderLimit		string 		`json:"order_limit,omitempty"`
	CounterLimit	string 		`json:"counter_limit,omitempty"`	
	Amount			float64 	`json:"amount,omitempty"`
	Quantity		int 		`json:"quantity,omitempty"`
}

type LimitTransaction struct {
	ID				int			`json:"id,omitempty"`
	TransactionId	string 		`json:"transaction_id,omitempty"`
	Key				string 		`json:"key,omitempty"`
	TypeLimit		string 		`json:"type_limit,omitempty"`
	CounterLimit	string 		`json:"counter_limit,omitempty"`	
	OrderLimit		string 		`json:"order_limit,omitempty"`	
	Status			string 		`json:"status,omitempty"`
	Amount			float64 	`json:"amount,omitempty"`
	CreareAt		time.Time 	`json:"created_at,omitempty"`			
}
