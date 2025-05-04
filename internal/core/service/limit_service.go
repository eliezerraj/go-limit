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

var tracerProvider go_core_observ.TracerProvider
var childLogger = log.With().Str("component","go-limit").Str("package","internal.core.service").Logger()

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

// About create a person
func (s *WorkerService) GetTransactionLimit(ctx context.Context, transactionLimit model.TransactionLimit) (*model.TransactionLimit, error){
	childLogger.Info().Str("func","GetTransactionLimit").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("transactionLimit", transactionLimit).Send()

	// trace
	span := tracerProvider.Span(ctx, "service.GetTransactionLimit")
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

	// Businness rule
	if transactionLimit.TransactionAt.IsZero() {
		transactionLimit.TransactionAt = time.Now()
	}
	transactionLimit.Status = "REQUESTED"

	// save the transaction
	res_transactionLimit, err := s.workerRepository.AddTransactionLimit(ctx, tx, transactionLimit)
	if err != nil {
		return nil, err
	}
	transactionLimit.ID = res_transactionLimit.ID

	// get the current spending limit
	res_spendLimit, err := s.workerRepository.GetSpendLimit(ctx, transactionLimit)
	if err != nil {
		return nil, err
	}

	// get the transaction limit status
	res_transactionLimit, err = s.workerRepository.GetTransactionLimit(ctx, transactionLimit)
	if err != nil {
		return nil, err
	}
	transactionLimit.SumAmount = res_transactionLimit.SumAmount
	transactionLimit.SumCount = res_transactionLimit.SumCount

	childLogger.Info().Interface("res_spendLimit",res_spendLimit ).Send()
	childLogger.Info().Interface("res_transactionLimit",res_transactionLimit ).Send()

	// check the breach
	if res_spendLimit.LimitAmount < transactionLimit.SumAmount || res_spendLimit.LimitHour <  transactionLimit.SumCount{
		transactionLimit.Status = "BREACH_LIMIT:" + transactionLimit.Category

		breach_limit := model.BreachLimit{	FkIdTransLimit: 	transactionLimit.ID ,
											TransactionId: 		transactionLimit.TransactionId ,
											Mcc: 				transactionLimit.Mcc,
											Status: 			transactionLimit.Status,
											Amount: 			res_spendLimit.LimitAmount - transactionLimit.SumAmount,
											Count: 				transactionLimit.SumCount }

		_, err = s.workerRepository.AddBreachLimit(ctx, tx, breach_limit)
		if err != nil {
			return nil, err
		}
	}

	return &transactionLimit, nil
}