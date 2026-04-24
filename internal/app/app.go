package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	"hrbackend/config"
	docs "hrbackend/docs"
	"hrbackend/internal/domain"
	"hrbackend/internal/handler"
	"hrbackend/internal/middleware"
	"hrbackend/internal/repository"
	db "hrbackend/internal/repository/db"
	"hrbackend/internal/service"
	"hrbackend/internal/ws"
	pkgasynq "hrbackend/pkg/asynq"
	pkgjwt "hrbackend/pkg/jwt"
	pkglogger "hrbackend/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type App struct {
	Config    config.Config
	Router    *gin.Engine
	DB        *pgxpool.Pool
	TaskQueue domain.TaskQueue
	WSHub     *ws.Hub
	WSTickets domain.WebSocketTicketStore
	logger    domain.Logger
}

func Build(ctx context.Context, cfg config.Config) (*App, error) {
	logger, err := pkglogger.Setup(cfg.Environment)
	if err != nil {
		return nil, fmt.Errorf("setup logger: %w", err)
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DbSource)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}
	poolConfig.AfterConnect = registerPostgresTypes

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	store := db.NewStore(pool)

	tokenMaker, err := newTokenMaker(cfg)
	if err != nil {
		pool.Close()
		return nil, err
	}

	taskQueue := buildTaskQueue(cfg)
	wsHub := ws.NewHub()
	go wsHub.Run()
	wsTicketStore := newWebSocketTicketStore(cfg)
	router := buildRouter(cfg, logger, store, tokenMaker, taskQueue, wsHub, wsTicketStore)

	return &App{
		Config:    cfg,
		Router:    router,
		DB:        pool,
		TaskQueue: taskQueue,
		WSHub:     wsHub,
		WSTickets: wsTicketStore,
		logger:    logger,
	}, nil
}

func registerPostgresTypes(ctx context.Context, conn *pgx.Conn) error {
	for _, typeName := range postgresEnumTypeNames {
		typ, err := conn.LoadType(ctx, typeName)
		if err != nil {
			return fmt.Errorf("load postgres type %q: %w", typeName, err)
		}
		conn.TypeMap().RegisterType(typ)

		arrayType, err := conn.LoadType(ctx, "_"+typeName)
		if err != nil {
			return fmt.Errorf("load postgres array type %q: %w", "_"+typeName, err)
		}
		conn.TypeMap().RegisterType(arrayType)
	}

	return nil
}

var postgresEnumTypeNames = []string{
	"location_type_enum",
	"permission_override_effect",
	"notification_type_enum",
	"gender_enum",
	"employee_contract_type_enum",
	"irregular_hours_profile_enum",
	"training_assignment_status_enum",
	"handbook_step_kind_enum",
	"handbook_assignment_status_enum",
	"handbook_step_status_enum",
	"handbook_template_status_enum",
	"handbook_assignment_event_enum",
	"time_entry_status_enum",
	"time_entry_hour_type_enum",
	"shift_swap_status_enum",
	"leave_request_type_enum",
	"leave_request_status_enum",
	"payout_request_status_enum",
	"expense_request_category_enum",
	"expense_request_status_enum",
	"pay_period_status_enum",
	"calendar_event_kind_enum",
	"calendar_event_status_enum",
	"calendar_event_work_approval_status_enum",
	"attendee_response_enum",
	"reminder_channel_enum",
	"performance_assessment_status_enum",
	"performance_work_assignment_status_enum",
}

func (a *App) Close(_ context.Context) error {
	var errs []string

	if a.TaskQueue != nil {
		if err := a.TaskQueue.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("close task queue: %v", err))
		}
	}

	if a.WSHub != nil {
		a.WSHub.Shutdown()
	}

	if a.WSTickets != nil {
		if err := a.WSTickets.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("close websocket tickets: %v", err))
		}
	}

	if a.DB != nil {
		a.DB.Close()
	}

	if syncer, ok := a.logger.(interface{ Sync() error }); ok {
		if err := syncer.Sync(); err != nil {
			errs = append(errs, fmt.Sprintf("sync logger: %v", err))
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

func buildRouter(
	cfg config.Config,
	logger domain.Logger,
	store *db.Store,
	tokenMaker domain.TokenMaker,
	taskQueue domain.TaskQueue,
	wsHub *ws.Hub,
	wsTicketStore domain.WebSocketTicketStore,
) *gin.Engine {
	setGinMode(cfg.Environment)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))
	router.Use(middleware.NewRequestContextMiddleware(logger).Handle())
	router.Use(middleware.NewRequestLoggingMiddleware(logger, cfg.Environment).Handle())

	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	docs.SwaggerInfo.BasePath = "/api"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	authMiddleware := middleware.NewAuthMiddleware(tokenMaker, logger)
	permissionMiddleware := middleware.NewPermissionMiddleware(store, logger)

	authRepo := repository.NewAuthRepository(store)
	authService := service.NewAuthService(
		authRepo,
		tokenMaker,
		logger,
		cfg.AccessTokenDuration,
		cfg.RefreshTokenDuration,
		cfg.TwoFATokenDuration,
	)

	employeeRepo := repository.NewEmployeeRepository(store)
	employeeService := service.NewEmployeeService(employeeRepo, logger)

	organizationRepo := repository.NewOrganizationRepository(store)
	organizationService := service.NewOrganizationService(organizationRepo, logger)

	settingsRepo := repository.NewSettingsRepository(store)
	settingsService := service.NewSettingsService(settingsRepo, logger)

	departmentRepo := repository.NewDepartmentRepository(store)
	departmentService := service.NewDepartmentService(departmentRepo, logger)

	roleRepo := repository.NewRoleRepository(store)
	roleService := service.NewRoleService(roleRepo, logger)

	scheduleRepo := repository.NewScheduleRepository(store)
	scheduleService := service.NewScheduleService(scheduleRepo, taskQueue, logger)

	leaveRepo := repository.NewLeaveRepository(store)
	leaveService := service.NewLeaveService(leaveRepo, logger)

	payoutRepo := repository.NewPayoutRepository(store)
	payoutService := service.NewPayoutService(payoutRepo, logger)

	expenseRepo := repository.NewExpenseRepository(store)
	expenseService := service.NewExpenseService(expenseRepo, logger)

	timeEntryRepo := repository.NewTimeEntryRepository(store)
	timeEntryService := service.NewTimeEntryService(timeEntryRepo, logger)

	performanceRepo := repository.NewPerformanceRepository(store)
	performanceService := service.NewPerformanceService(performanceRepo, logger)

	handbookRepo := repository.NewHandbookRepository(store)
	handbookService := service.NewHandbookService(handbookRepo, logger)

	trainingRepo := repository.NewTrainingRepository(store)
	trainingService := service.NewTrainingService(trainingRepo, logger)

	authHandler := handler.NewAuthHandler(authService)
	wsAuthService := service.NewWebSocketAuthService(wsTicketStore, logger, cfg.WsTicketTTL)
	wsHandler := handler.NewWebSocketHandler(wsAuthService, wsHub, logger, cfg.WsAllowedOrigins)
	employeeHandler := handler.NewEmployeeHandler(employeeService)
	organizationHandler := handler.NewOrganizationHandler(organizationService)
	settingsHandler := handler.NewSettingsHandler(settingsService)
	departmentHandler := handler.NewDepartmentHandler(departmentService)
	roleHandler := handler.NewRoleHandler(roleService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)
	shiftSwapHandler := handler.NewShiftSwapHandler(scheduleService)
	leaveHandler := handler.NewLeaveHandler(leaveService)
	payoutHandler := handler.NewPayoutHandler(payoutService)
	expenseHandler := handler.NewExpenseHandler(expenseService)
	timeEntryHandler := handler.NewTimeEntryHandler(timeEntryService)
	performanceHandler := handler.NewPerformanceHandler(performanceService)
	handbookHandler := handler.NewHandbookHandler(handbookService)
	trainingHandler := handler.NewTrainingHandler(trainingService)

	api := router.Group("/api")
	auth := authMiddleware.Handle()
	requirePermission := permissionMiddleware.Require

	handler.RegisterAuthRoutes(api, authHandler, auth)
	handler.RegisterWebSocketRoutes(api, wsHandler, auth)
	handler.RegisterEmployeeRoutes(api, employeeHandler, auth, requirePermission)
	handler.RegisterOrganizationRoutes(api, organizationHandler, auth, requirePermission)
	handler.RegisterSettingsRoutes(api, settingsHandler, auth, requirePermission)
	handler.RegisterDepartmentRoutes(api, departmentHandler, auth, requirePermission)
	handler.RegisterRoleRoutes(api, roleHandler, auth, requirePermission)
	handler.RegisterScheduleRoutes(api, scheduleHandler, auth, requirePermission)
	handler.RegisterShiftSwapRoutes(api, shiftSwapHandler, auth, requirePermission)
	handler.RegisterLeaveRoutes(api, leaveHandler, auth, requirePermission)
	handler.RegisterPayoutRoutes(api, payoutHandler, auth, requirePermission)
	handler.RegisterExpenseRoutes(api, expenseHandler, auth, requirePermission)
	handler.RegisterTimeEntryRoutes(api, timeEntryHandler, auth, requirePermission)
	handler.RegisterPerformanceRoutes(api, performanceHandler, auth, requirePermission)
	handler.RegisterHandbookRoutes(api, handbookHandler, auth, requirePermission)
	handler.RegisterTrainingRoutes(api, trainingHandler, auth, requirePermission)

	return router
}

func setGinMode(environment string) {
	switch strings.ToLower(environment) {
	case "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}

func newTokenMaker(cfg config.Config) (domain.TokenMaker, error) {
	maker, err := pkgjwt.New(
		cfg.AccessTokenSecretKey,
		cfg.RefreshTokenSecretKey,
		cfg.TwoFATokenSecretKey,
	)
	if err != nil {
		return nil, fmt.Errorf("create token maker: %w", err)
	}

	return &tokenMakerAdapter{maker: maker}, nil
}

func buildTaskQueue(cfg config.Config) domain.TaskQueue {
	if strings.TrimSpace(cfg.RedisHost) == "" {
		return nil
	}

	var tlsConfig *tls.Config
	if cfg.Remote {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	client := pkgasynq.NewClient(cfg.RedisHost, "", cfg.RedisPassword, tlsConfig)
	return &taskQueueAdapter{client: client}
}
