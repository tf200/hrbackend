package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	"hrbackend/config"
	"hrbackend/internal/domain"
	"hrbackend/internal/handler"
	"hrbackend/internal/middleware"
	"hrbackend/internal/repository"
	db "hrbackend/internal/repository/db"
	"hrbackend/internal/service"
	pkgasynq "hrbackend/pkg/asynq"
	pkgjwt "hrbackend/pkg/jwt"
	pkglogger "hrbackend/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Config    config.Config
	Router    *gin.Engine
	DB        *pgxpool.Pool
	TaskQueue domain.TaskQueue
	logger    domain.Logger
}

func Build(ctx context.Context, cfg config.Config) (*App, error) {
	logger, err := pkglogger.Setup(cfg.Environment)
	if err != nil {
		return nil, fmt.Errorf("setup logger: %w", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DbSource)
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
	router := buildRouter(cfg, logger, store, tokenMaker, taskQueue)

	return &App{
		Config:    cfg,
		Router:    router,
		DB:        pool,
		TaskQueue: taskQueue,
		logger:    logger,
	}, nil
}

func (a *App) Close(_ context.Context) error {
	var errs []string

	if a.TaskQueue != nil {
		if err := a.TaskQueue.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("close task queue: %v", err))
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

	departmentRepo := repository.NewDepartmentRepository(store)
	departmentService := service.NewDepartmentService(departmentRepo, logger)

	scheduleRepo := repository.NewScheduleRepository(store)
	scheduleService := service.NewScheduleService(scheduleRepo, taskQueue, logger)

	leaveRepo := repository.NewLeaveRepository(store)
	leaveService := service.NewLeaveService(leaveRepo, logger)

	payoutRepo := repository.NewPayoutRepository(store)
	payoutService := service.NewPayoutService(payoutRepo)

	timeEntryRepo := repository.NewTimeEntryRepository(store)
	timeEntryService := service.NewTimeEntryService(timeEntryRepo, logger)

	handbookRepo := repository.NewHandbookRepository(store)
	handbookService := service.NewHandbookService(handbookRepo, logger)

	authHandler := handler.NewAuthHandler(authService)
	employeeHandler := handler.NewEmployeeHandler(employeeService)
	organizationHandler := handler.NewOrganizationHandler(organizationService)
	departmentHandler := handler.NewDepartmentHandler(departmentService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)
	shiftSwapHandler := handler.NewShiftSwapHandler(scheduleService)
	leaveHandler := handler.NewLeaveHandler(leaveService)
	payoutHandler := handler.NewPayoutHandler(payoutService)
	timeEntryHandler := handler.NewTimeEntryHandler(timeEntryService)
	handbookHandler := handler.NewHandbookHandler(handbookService)

	api := router.Group("/api")
	auth := authMiddleware.Handle()
	requirePermission := permissionMiddleware.Require

	handler.RegisterAuthRoutes(api, authHandler, auth)
	handler.RegisterEmployeeRoutes(api, employeeHandler, auth, requirePermission)
	handler.RegisterOrganizationRoutes(api, organizationHandler, auth, requirePermission)
	handler.RegisterDepartmentRoutes(api, departmentHandler, auth, requirePermission)
	handler.RegisterScheduleRoutes(api, scheduleHandler, auth, requirePermission)
	handler.RegisterShiftSwapRoutes(api, shiftSwapHandler, auth, requirePermission)
	handler.RegisterLeaveRoutes(api, leaveHandler, auth, requirePermission)
	handler.RegisterPayoutRoutes(api, payoutHandler, auth, requirePermission)
	handler.RegisterTimeEntryRoutes(api, timeEntryHandler, auth, requirePermission)
	handler.RegisterHandbookRoutes(api, handbookHandler, auth, requirePermission)

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
