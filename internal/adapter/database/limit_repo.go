package database

import (
	"context"
	"time"
	"errors"
	
	"github.com/go-limit/internal/core/model"
	"github.com/go-limit/internal/core/erro"

	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

var tracerProvider go_core_observ.TracerProvider
var childLogger = log.With().Str("component","go-limit").Str("package","internal.core.database").Logger()

type WorkerRepository struct {
	DatabasePGServer *go_core_pg.DatabasePGServer
}

// Above new worker
func NewWorkerRepository(databasePGServer *go_core_pg.DatabasePGServer) *WorkerRepository{
	childLogger.Info().Str("func","NewWorkerRepository").Send()

	return &WorkerRepository{
		DatabasePGServer: databasePGServer,
	}
}

func (w WorkerRepository) AddTransactionLimit(ctx context.Context, tx pgx.Tx, transactionLimit model.TransactionLimit) (*model.TransactionLimit, error){
	childLogger.Info().Str("func","AddTransactionLimit").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.AddTransactionLimit")
	defer span.End()

	// prepare
	if transactionLimit.TransactionAt.IsZero() {
		transactionLimit.TransactionAt = time.Now()
	}

	//query
	query := `INSERT INTO transaction_limit (transaction_id,
											category, 
											card_number,
											mcc, 
											status,
											transaction_at, 
											currency,
											amount,
											tenant_id) 
											VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	// execute
	row := tx.QueryRow(ctx, query,  transactionLimit.TransactionId, 
									transactionLimit.Category,
									transactionLimit.CardNumber,
									transactionLimit.Mcc,
									transactionLimit.Status,
									transactionLimit.TransactionAt,
									transactionLimit.Currency,
									transactionLimit.Amount,
									transactionLimit.TenantID,
									)

	var id int
	
	if err := row.Scan(&id); err != nil {
		return nil, errors.New(err.Error())
	}

	transactionLimit.ID = id

	return &transactionLimit, nil
}

// Above create a breach limit log
func (w WorkerRepository) AddBreachLimit(ctx context.Context, tx pgx.Tx, breachLimit model.BreachLimit) (*model.BreachLimit, error){
	childLogger.Info().Str("func","AddBreachLimit").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.AddBreachLimit")
	defer span.End()

	// prepare
	breachLimit.CreatedAt = time.Now()

	//query
	query := `INSERT INTO breach_limit (fk_id_trans_limit,
										transaction_id,
										mcc, 
										status,
										amount, 
										count,
										created_at, 
										tenant_id) 
										VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	// execute
	row := tx.QueryRow(ctx, query,  breachLimit.FkIdTransLimit,  
									breachLimit.TransactionId,
									breachLimit.Mcc,
									breachLimit.Status,
									breachLimit.Amount,
									breachLimit.Count,
									breachLimit.CreatedAt,
									breachLimit.TenantID,
									)

	var id int
	
	if err := row.Scan(&id); err != nil {
		return nil, errors.New(err.Error())
	}

	breachLimit.ID = id

	return &breachLimit, nil
}

// Above get the transaction limit response
func (w WorkerRepository) GetTransactionLimit(ctx context.Context, transactionLimit model.TransactionLimit) (*model.TransactionLimit, error){
	childLogger.Info().Str("func","GetTransactionLimit").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()
	
	// trace
	span := tracerProvider.Span(ctx, "database.GetTransactionLimit")
	defer span.End()

	// prepare database
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// prepare query
	res_transactionLimit := model.TransactionLimit{}

	query := `select coalesce( sum(amount), 0) as transaction_sum_amount,
					 coalesce( count(1), 0) as transaction_sum_count
				from transaction_limit
				where category = $2
				and card_number = $1
				and mcc = $3
				and transaction_at between (now() - interval '0.5 hour') and now() `

	// execute			
	rows, err := conn.Query(ctx, 
							query, 
							transactionLimit.CardNumber,
							transactionLimit.Category,
							transactionLimit.Mcc )
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&res_transactionLimit.SumAmount,
							&res_transactionLimit.SumCount )
		if err != nil {
			return nil, errors.New(err.Error())
        }
		return &res_transactionLimit, nil
	}
	
	return nil, erro.ErrNotFound
}

// Above get the spending limit
func (w WorkerRepository) GetSpendLimit(ctx context.Context, transactionLimit model.TransactionLimit) (*model.SpendLimit, error){
	childLogger.Info().Str("func","GetSpendLimit").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.GetSpendLimit")
	defer span.End()

	// prepare database
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// prepare query
	res_spendLimit := model.SpendLimit{}

	query := `select coalesce( sum(amount) , 0) as limit_amount,
					coalesce( sum(day), 0) as limit_day,
					coalesce( sum(hour), 0) as limit_hour,
					coalesce( sum(minute), 0) as limit_minute
			from spend_limit
			where category = $1
			and mcc = $2`

	rows, err := conn.Query(ctx, 
							query, 
							transactionLimit.Category,
							transactionLimit.Mcc )
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	// execute	
	for rows.Next() {
		err := rows.Scan( 	&res_spendLimit.LimitAmount,
							&res_spendLimit.LimitDay, 
							&res_spendLimit.LimitHour,
							&res_spendLimit.LimitMinute,
						)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		return &res_spendLimit, nil
	}
	
	return nil, erro.ErrNotFound
}
