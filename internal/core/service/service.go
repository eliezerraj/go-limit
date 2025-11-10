package service

import(
	"time"
	"context"

	"github.com/rs/zerolog/log"

	"github.com/go-limit/internal/core/model"
	"github.com/go-limit/internal/adapter/database"

	go_core_pg "github.com/eliezerraj/go-core/database/pg"
	go_core_observ "github.com/eliezerraj/go-core/observability"
)

var (
	tracerProvider go_core_observ.TracerProvider
	childLogger = log.With().Str("component","go-limit").Str("package","internal.core.service").Logger()
)

type WorkerService struct {
	workerRepository 	*database.WorkerRepository}

// About create a new worker service
func NewWorkerService(	workerRepository *database.WorkerRepository) *WorkerService{
	childLogger.Info().Str("func","NewWorkerService").Send()

	return &WorkerService{
		workerRepository: workerRepository,
	}
}

// About handle/convert http status code
func (s *WorkerService) Stat(ctx context.Context) (go_core_pg.PoolStats){
	childLogger.Info().Str("func","Stat").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	return s.workerRepository.Stat(ctx)
}

// About check the limit
func (s *WorkerService) CheckLimitTransaction(ctx context.Context, limit model.Limit) (*[]model.LimitTransaction, error){
	childLogger.Info().Str("func","CheckLimitTransaction").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("limit", limit).Send()

	// trace
	ctx, span := tracerProvider.SpanCtx(ctx, "service.CheckLimitTransaction")
	defer span.End()
	
	// prepare batabase
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.workerRepository.DatabasePGServer.ReleaseTx(conn)
	
	// handle connection
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		span.End()
	}()

	// check the type limit
	type_limit := model.TypeLimit{Code: limit.TypeLimit}
	_, err = s.workerRepository.GetTypeLimit(ctx, type_limit)
	if err != nil {
		return nil, err
	}

	// get list order limit
	res_order_limit := model.OrderLimit{TypeLimit: limit.TypeLimit,
										CounterLimit: limit.OrderLimit}

	res_lis_order_limit, err := s.workerRepository.GetOrderLimit(ctx, res_order_limit)
	if err != nil {
		return nil, err
	}

	//childLogger.Info().Interface("== 1 ===> res_lis_order_limit", res_lis_order_limit ).Send()

	// Create a list of limit transaction
	list_limitTransaction := []model.LimitTransaction{}

	// for each order limit check the limit transaction 
	for _, val := range *res_lis_order_limit{

		limit.TypeLimit = val.TypeLimit
		limit.OrderLimit = val.Type
		limit.CounterLimit = val.CounterLimit
		
		// get all transaction per key and per count limit
		res_limit_trans_per_key, err := s.workerRepository.GetLimitTransactionPerKey(ctx, limit)
		if err != nil {
				return nil, err
		}

		//childLogger.Info().Interface("== 22 ===> res_limit_trans_per_key", res_limit_trans_per_key ).Send()

		// check if the limit is breach
		var tmp_amount float64
		var tmp_status = "LIMIT:APROVED"

		if val.CounterLimit == "VALUE" {
			//childLogger.Info().Interface("== 22 VALUE ===> float64(res_limit_trans_per_key.Amount) ", float64(res_limit_trans_per_key.Amount) ).Send()
			//childLogger.Info().Interface("== 22 VALUE ===> val.Amount ", val.Amount ).Send()

			tmp_amount =  limit.Amount 
			if float64(res_limit_trans_per_key.Amount) > float64(val.Amount) {
				tmp_status = "LIMIT:VALUE:BREACH"
			} else {
				tmp_status  = "LIMIT:VALUE:APPROVED"
			}
		}

		if val.CounterLimit == "QUANTITY" {
			//childLogger.Info().Interface("== 22 QUANTITY ===> res_limit_trans_per_key", float64(res_limit_trans_per_key.Amount) ).Send()
			//childLogger.Info().Interface("== 22 QUANTITY ===> val.Amount", float64(val.Amount) ).Send()

			tmp_amount =  float64(limit.Quantity)
			if float64(res_limit_trans_per_key.Amount) > float64(val.Amount) {
				tmp_status = "LIMIT:QUANTITY:BREACH"
			} else {
				tmp_status  = "LIMIT:QUANTITY:APPROVED"
			}
		}

		if val.CounterLimit == "MINUTE" {
			continue
		}

		limitTransaction := model.LimitTransaction{	TransactionId: limit.TransactionId,
													Key: limit.Key,
													TypeLimit: limit.TypeLimit,
													CounterLimit: val.CounterLimit,
													OrderLimit: limit.OrderLimit,
													Status: tmp_status,
													Amount: tmp_amount,
													CreareAt: time.Now(), 
												} 
			
		// save the transaction
		res_limit_transaction, err := s.workerRepository.AddLimitTransaction(ctx, tx, limitTransaction)
		if err != nil {
			return nil, err
		}

		//childLogger.Info().Interface("---3333--->res_limit_transaction",res_limit_transaction ).Send()

		limitTransaction.ID = res_limit_transaction.ID
		list_limitTransaction = append(list_limitTransaction, limitTransaction)
	}

	return &list_limitTransaction, nil
}
