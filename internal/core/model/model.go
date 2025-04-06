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

type TransactionLimit struct {
	ID				int			`json:"id,omitempty"`
	Category		string 		`json:"category,omitempty"`
	CardNumber		string 		`json:"card_number,omitempty"`
	TransactionId	string 		`json:"transaction_id,omitempty"`
	Mcc				string 		`json:"mcc,omitempty"`
	Status			string 		`json:"status,omitempty"`
	TransactionAt	time.Time 	`json:"transaction_at,omitempty"`		
	Currency		string 		`json:"currency,omitempty"`
	Amount			float64 	`json:"amount,omitempty"`
	SumAmount		float64 	`json:"transaction_sum_amount,omitempty"`
	SumCount		int 		`json:"transaction_sum_count,omitempty"`
	TenantId		string 		`json:"tenant_id,omitempty"`
}

type SpendLimit struct {
	Category	string 		`json:"category,omitempty"`
	Mcc			string 		`json:"mcc,omitempty"`
	LimitAmount	float64 	`json:"limit_amount,omitempty"`
	LimitDay	int 		`json:"limit_day,omitempty"`
	LimitHour	int 		`json:"limit_hour,omitempty"`
	LimitMinute	int 		`json:"limit_minute,omitempty"`
}

type BreachLimit struct {
	ID				int			`json:"id,omitempty"`
	FkIdTransLimit	int 		`json:"fk_id_transaction_limit,omitempty"`
	TransactionId	string 		`json:"transaction_id,omitempty"`
	Mcc				string 		`json:"mcc,omitempty"`
	Status			string 		`json:"status,omitempty"`
	Amount			float64 	`json:"breach_amount,omitempty"`
	Count			int 		`json:"breach_count,omitempty"`
	CreatedAt		time.Time 	`json:"created_at,omitempty"`
	TenantId		string 		`json:"tenant_id,omitempty"`
}