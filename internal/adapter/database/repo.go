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

var (
	tracerProvider go_core_observ.TracerProvider
	childLogger = log.With().Str("component","go-limit").Str("package","internal.core.database").Logger()
)

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

// Above get stats from database
func (w WorkerRepository) Stat(ctx context.Context) (go_core_pg.PoolStats){
	childLogger.Info().Str("func","Stat").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()
	
	stats := w.DatabasePGServer.Stat()

	resPoolStats := go_core_pg.PoolStats{
		AcquireCount:         stats.AcquireCount(),
		AcquiredConns:        stats.AcquiredConns(),
		CanceledAcquireCount: stats.CanceledAcquireCount(),
		ConstructingConns:    stats.ConstructingConns(),
		EmptyAcquireCount:    stats.EmptyAcquireCount(),
		IdleConns:            stats.IdleConns(),
		MaxConns:             stats.MaxConns(),
		TotalConns:           stats.TotalConns(),
	}

	return resPoolStats
}

// Above get type limit
func (w WorkerRepository) GetTypeLimit(ctx context.Context, typeLimit model.TypeLimit) (*model.TypeLimit, error){
	childLogger.Info().Str("func","GetTypeLimit").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.GetTypeLimit")
	defer span.End()

	// prepare database
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Send()
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// prepare query
	res_type_limit := model.TypeLimit{}

	query := `select code,
					 category,
					 created_at	
			  from type_limit
			  where code = $1`

	rows, err := conn.Query(ctx, 
							query, 
							typeLimit.Code)
	if err != nil {
		childLogger.Error().Err(err).Send()
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	// execute	
	for rows.Next() {
		err := rows.Scan( 	&res_type_limit.Code,
							&res_type_limit.Category, 
							&res_type_limit.CreateAt,
						)
		if err != nil {
			childLogger.Error().Err(err).Send()
			return nil, errors.New(err.Error())
        }
		return &res_type_limit, nil
	}
	
	return nil, erro.ErrNotFound
}

func (w WorkerRepository) GetOrderLimit(ctx context.Context, orderLimit model.OrderLimit) (*[]model.OrderLimit, error){
	childLogger.Info().Str("func","GetOrderLimit").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()
	// trace
	span := tracerProvider.Span(ctx, "database.GetOrderLimit")
	defer span.End()

	// prepare database
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Send()
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// prepare query
	res_lis_order_limit := []model.OrderLimit{}

	query := `select fk_type_limit_code,
					 fk_counter_limit_code,
					 type,
					 amount	
			  from order_limit
			  where fk_type_limit_code = $1
			  and type = $2`

	rows, err := conn.Query(ctx, 
							query, 
							orderLimit.TypeLimit,
							orderLimit.CounterLimit)
	if err != nil {
		childLogger.Error().Err(err).Send()
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	// execute	
	for rows.Next() {
			
		res_order_limit := model.OrderLimit{}

		err := rows.Scan( 	&res_order_limit.TypeLimit,
							&res_order_limit.CounterLimit, 
							&res_order_limit.Type,
							&res_order_limit.Amount,
						)
		if err != nil {
			childLogger.Error().Err(err).Send()
			return nil, errors.New(err.Error())
        }

		res_lis_order_limit = append(res_lis_order_limit, res_order_limit)		
	}
	
	return &res_lis_order_limit, nil
}

// Above get the transaction limit response
func (w WorkerRepository) GetLimitTransactionPerKey(ctx context.Context, limit model.Limit) (*model.Limit, error){
	childLogger.Info().Str("func","GetLimitTransactionPerKey").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.GetLimitTransactionPerKey")
	defer span.End()

	// prepare database
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Send()
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// prepare query
	res_limit := model.Limit{}

	query := `select coalesce( sum(amount), 0) as transaction_sum_amount,
					 coalesce( count(1), 0) as transaction_sum_count
				from public.limit_transaction
				where key = $1
				and fk_type_limit_code = $2
				and fk_order_limit_type = $3
				and fk_counter_limit_code = $4
				and created_at between (now() - interval '1 minute') and now()`

	// execute			
	rows, err := conn.Query(ctx, 
							query, 
							limit.Key,
							limit.TypeLimit,
							limit.OrderLimit,
							limit.CounterLimit,
						)
	if err != nil {
		childLogger.Error().Err(err).Send()
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( &res_limit.Amount,
						  &res_limit.Quantity )
		if err != nil {
			childLogger.Error().Err(err).Send()
			return nil, errors.New(err.Error())
        }
		return &res_limit, nil
	}
	
	return nil, erro.ErrNotFound
}

// Above add transaction limit
func (w WorkerRepository) AddLimitTransaction(ctx context.Context, tx pgx.Tx, limitTransaction model.LimitTransaction) (*model.LimitTransaction, error){
	childLogger.Info().Str("func","AddLimitTransaction").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.AddLimitTransaction")
	defer span.End()

	// prepare
	limitTransaction.CreareAt = time.Now()

	//query
	query := `INSERT INTO limit_transaction (transaction_id,
											key, 
											fk_type_limit_code,
											fk_counter_limit_code,
											fk_order_limit_type,
											status,
											amount,
											created_at) 
											VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	// execute
	row := tx.QueryRow(ctx, query,  limitTransaction.TransactionId, 
									limitTransaction.Key,
									limitTransaction.TypeLimit,
									limitTransaction.CounterLimit,
									limitTransaction.OrderLimit,
									limitTransaction.Status,
									limitTransaction.Amount,
									limitTransaction.CreareAt,
									)

	var id int
	
	if err := row.Scan(&id); err != nil {
		childLogger.Error().Err(err).Send()
		return nil, errors.New(err.Error())
	}

	limitTransaction.ID = id

	return &limitTransaction, nil
}
