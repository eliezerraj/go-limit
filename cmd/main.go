package main

import(
	"time"
	"context"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-limit/internal/infra/configuration"
	"github.com/go-limit/internal/core/model"
	"github.com/go-limit/internal/core/service"
	"github.com/go-limit/internal/infra/server"
	"github.com/go-limit/internal/adapter/api"
	"github.com/go-limit/internal/adapter/database"

	go_core_pg "github.com/eliezerraj/go-core/database/pg"
)

var(
	logLevel = 	zerolog.InfoLevel // zerolog.InfoLevel zerolog.DebugLevel
	appServer	model.AppServer
	databaseConfig 		go_core_pg.DatabaseConfig
	databasePGServer 	go_core_pg.DatabasePGServer

	childLogger = log.With().Str("component","go-limit").Str("package", "main").Logger()
)

// Above init
func init(){
	childLogger.Info().Str("func","init").Send()
	zerolog.SetGlobalLevel(logLevel)

	infoPod, server := configuration.GetInfoPod()
	configOTEL 		:= configuration.GetOtelEnv()
	databaseConfig 	:= configuration.GetDatabaseEnv() 

	appServer.InfoPod = &infoPod
	appServer.Server = &server
	appServer.ConfigOTEL = &configOTEL
	appServer.DatabaseConfig = &databaseConfig
}

// Above main
func main (){
	childLogger.Info().Str("func","main").Interface("appServer",appServer).Send()
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Open Database
	count := 1
	var err error
	for {
		databasePGServer, err = databasePGServer.NewDatabasePGServer(ctx, *appServer.DatabaseConfig)
		if err != nil {
			if count < 3 {
				log.Error().Err(err).Msg("error open database... trying again !!")
			} else {
				log.Error().Err(err).Msg("fatal error open Database aborting")
				panic(err)
			}
			time.Sleep(3 * time.Second) //backoff
			count = count + 1
			continue
		}
		break
	}

	// wire	
	database := database.NewWorkerRepository(&databasePGServer)
	workerService := service.NewWorkerService(database)
	httpRouters := api.NewHttpRouters(workerService, time.Duration(appServer.Server.CtxTimeout))

	// start server
	httpServer := server.NewHttpAppServer(appServer.Server)
	httpServer.StartHttpAppServer(ctx, &httpRouters, &appServer)
}