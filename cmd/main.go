package main

import (
	"html/template"
	"log"
	"net/http"
	"simdokpol/internal/config"
	"simdokpol/internal/controllers"
	"simdokpol/internal/middleware"
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"simdokpol/internal/services"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	_ "simdokpol/docs" // Import direktori docs yang akan digenerate.

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func runMigrations(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("could not get sql.DB: %v", err)
	}
	driver, err := sqlite.WithInstance(sqlDB, &sqlite.Config{})
	if err != nil {
		log.Fatalf("could not create sqlite driver instance: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "sqlite", driver)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("an error occurred while syncing the database: %v", err)
	}
	log.Println("Database migration finished successfully")
}

// @title SIMDOKPOL API
// @version 1.0
// @description Ini adalah dokumentasi API untuk aplikasi Sistem Informasi Manajemen Dokumen Kepolisian.
// @termsOfService http://swagger.io/terms/

// @contact.name MUHAMMAD YUSUF ABDURROHMAN
// @contact.url https://github.com/dope-s/simdokpol-go
// @contact.email yusuf.aar@gmail.com

// @license.name MIT License
// @license.url https://github.com/dope-s/simdokpol-go/blob/main/LICENSE

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan token JWT Anda dengan format 'Bearer {token}'.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Gagal memuat konfigurasi: %v", err)
	}
	services.JWTSecretKey = []byte(cfg.JWTSecretKey)
	db, err := gorm.Open(gormsqlite.Open(cfg.DBDSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		log.Fatal(err)
	}
	runMigrations(db)
	funcMap := template.FuncMap{"ToUpper": strings.ToUpper}
	templates := template.New("").Funcs(funcMap)
	templates = template.Must(templates.ParseGlob("web/templates/*.html"))
	templates = template.Must(templates.ParseGlob("web/templates/partials/*.html"))
	userRepo := repositories.NewUserRepository(db)
	residentRepo := repositories.NewResidentRepository(db)
	docRepo := repositories.NewLostDocumentRepository(db)
	configRepo := repositories.NewConfigRepository(db)
	auditRepo := repositories.NewAuditLogRepository(db)
	configService := services.NewConfigService(configRepo)
	auditService := services.NewAuditLogService(auditRepo)
	authService := services.NewAuthService(userRepo)
	dashboardService := services.NewDashboardService(docRepo, userRepo, configService)
	docService := services.NewLostDocumentService(db, docRepo, residentRepo, userRepo, auditService, configService)
	userService := services.NewUserService(userRepo, auditService, cfg)
	backupService := services.NewBackupService(cfg, configService, auditService)
	authController := controllers.NewAuthController(authService)
	dashboardController := controllers.NewDashboardController(dashboardService)
	docController := controllers.NewLostDocumentController(docService)
	userController := controllers.NewUserController(userService)
	configController := controllers.NewConfigController(configService, userService)
	auditController := controllers.NewAuditLogController(auditService)
	backupController := controllers.NewBackupController(backupService)
	settingsController := controllers.NewSettingsController(configService, auditService)

	router := gin.Default()
	router.SetHTMLTemplate(templates)
	router.Static("/static", "./web/static")

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/setup", configController.ShowSetupPage)
	apiRoutes := router.Group("/api")
	{
		apiRoutes.POST("/setup", configController.SaveSetup)
	}

	app := router.Group("")
	app.Use(middleware.SetupMiddleware(configService))
	{
		app.GET("/login", func(c *gin.Context) { c.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login"}) })
		app.POST("/api/login", authController.Login)
		app.POST("/api/logout", authController.Logout)

		protected := app.Group("")
		protected.Use(middleware.AuthMiddleware(userRepo))
		{
			protected.GET("/", func(c *gin.Context) { user, _ := c.Get("currentUser"); c.HTML(http.StatusOK, "dashboard.html", gin.H{"Title": "Dasbor", "CurrentUser": user}) })
			protected.GET("/documents", func(c *gin.Context) { user, _ := c.Get("currentUser"); c.HTML(http.StatusOK, "document_list.html", gin.H{"Title": "Daftar Dokumen Aktif", "CurrentUser": user, "PageType": "active"}) })
			protected.GET("/documents/archived", func(c *gin.Context) { user, _ := c.Get("currentUser"); c.HTML(http.StatusOK, "document_list.html", gin.H{"Title": "Arsip Dokumen", "CurrentUser": user, "PageType": "archived"}) })
			protected.GET("/documents/new", func(c *gin.Context) { user, _ := c.Get("currentUser"); c.HTML(http.StatusOK, "document_form.html", gin.H{"Title": "Buat Surat Baru", "CurrentUser": user, "IsEdit": false, "DocID": 0}) })
			protected.GET("/documents/:id/edit", func(c *gin.Context) { user, _ := c.Get("currentUser"); id := c.Param("id"); c.HTML(http.StatusOK, "document_form.html", gin.H{"Title": "Edit Surat", "CurrentUser": user, "IsEdit": true, "DocID": id}) })
			protected.GET("/documents/:id/print", func(c *gin.Context) {
				userInterface, _ := c.Get("currentUser")
				currentUser := userInterface.(*models.User)
				idStr := c.Param("id")
				id, _ := strconv.ParseUint(idStr, 10, 32)
				doc, err := docService.FindByID(uint(id), currentUser.ID)
				if err != nil {
					if err.Error() == "akses ditolak" {
						c.HTML(http.StatusForbidden, "error.html", gin.H{"Title": "Akses Ditolak", "CurrentUser": currentUser, "ErrorMessage": "Anda tidak memiliki izin untuk mengakses dokumen ini."})
					} else {
						c.HTML(http.StatusNotFound, "error.html", gin.H{"Title": "Error", "CurrentUser": currentUser, "ErrorMessage": "Dokumen tidak ditemukan."})
					}
					return
				}
				appConfig, err := configService.GetConfig()
				if err != nil {
					c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Title": "Error", "CurrentUser": currentUser, "ErrorMessage": "Gagal memuat konfigurasi aplikasi."})
					return
				}
				c.HTML(http.StatusOK, "print_preview.html", gin.H{"Document": doc, "Now": time.Now(), "CurrentUser": currentUser, "Config": appConfig})
			})
			protected.GET("/search", func(c *gin.Context) { user, _ := c.Get("currentUser"); query := c.Query("q"); c.HTML(http.StatusOK, "search_results.html", gin.H{"Title": "Hasil Pencarian", "CurrentUser": user, "Query": query}) })
			protected.GET("/tentang", func(c *gin.Context) { user, _ := c.Get("currentUser"); c.HTML(http.StatusOK, "tentang.html", gin.H{"Title": "Tentang Aplikasi", "CurrentUser": user}) })
			protected.GET("/panduan", func(c *gin.Context) { user, _ := c.Get("currentUser"); c.HTML(http.StatusOK, "panduan.html", gin.H{"Title": "Panduan Pengguna", "CurrentUser": user}) })
			protected.GET("/profile", func(c *gin.Context) { user, _ := c.Get("currentUser"); c.HTML(http.StatusOK, "profile.html", gin.H{"Title": "Profil Pengguna", "CurrentUser": user}) })
			adminRoutes := protected.Group("")
			adminRoutes.Use(middleware.AdminAuthMiddleware())
			{
				adminRoutes.GET("/users", func(c *gin.Context) { user, _ := c.Get("currentUser"); c.HTML(http.StatusOK, "user_list.html", gin.H{"Title": "Manajemen Pengguna", "CurrentUser": user}) })
				adminRoutes.GET("/users/new", func(c *gin.Context) { user, _ := c.Get("currentUser"); c.HTML(http.StatusOK, "user_form.html", gin.H{"Title": "Tambah Pengguna", "CurrentUser": user, "IsEdit": false, "UserID": 0}) })
				adminRoutes.GET("/users/:id/edit", func(c *gin.Context) { user, _ := c.Get("currentUser"); id, _ := strconv.Atoi(c.Param("id")); c.HTML(http.StatusOK, "user_form.html", gin.H{"Title": "Edit Pengguna", "CurrentUser": user, "IsEdit": true, "UserID": id}) })
				adminRoutes.GET("/audit-logs", func(c *gin.Context) { user, _ := c.Get("currentUser"); c.HTML(http.StatusOK, "audit_log_list.html", gin.H{"Title": "Log Audit Sistem", "CurrentUser": user}) })
				adminRoutes.GET("/settings", func(c *gin.Context) { user, _ := c.Get("currentUser"); c.HTML(http.StatusOK, "settings.html", gin.H{"Title": "Pengaturan Sistem", "CurrentUser": user}) })
			}

			api := protected.Group("/api")
			{
				api.GET("/notifications/expiring-documents", dashboardController.GetExpiringDocuments)
				api.PUT("/profile", userController.UpdateProfile)
				api.PUT("/profile/password", userController.ChangePassword)
				api.GET("/search", docController.SearchGlobal)
				api.POST("/documents", docController.Create)
				api.GET("/documents", docController.FindAll)
				api.DELETE("/documents/:id", docController.Delete)
				api.GET("/documents/:id", docController.FindByID)
				api.PUT("/documents/:id", docController.Update)
				api.GET("/stats", dashboardController.GetStats)
				api.GET("/stats/monthly-issuance", dashboardController.GetMonthlyChart)
				api.GET("/stats/item-composition", dashboardController.GetItemCompositionChart)
				api.GET("/users/operators", userController.FindOperators)

				userAPI := api.Group("/users")
				userAPI.Use(middleware.AdminAuthMiddleware())
				{
					userAPI.POST("", userController.Create)
					userAPI.GET("", userController.FindAll)
					userAPI.GET("/:id", userController.FindByID)
					userAPI.PUT("/:id", userController.Update)
					userAPI.DELETE("/:id", userController.Delete)
					userAPI.POST("/:id/activate", userController.Activate)
				}

				auditAPI := api.Group("/audit-logs")
				auditAPI.Use(middleware.AdminAuthMiddleware())
				{
					auditAPI.GET("", auditController.FindAll)
				}

				api.POST("/backups", middleware.AdminAuthMiddleware(), backupController.CreateBackup)
				api.POST("/restore", middleware.AdminAuthMiddleware(), backupController.RestoreBackup)

				settingsAPI := api.Group("/settings")
				settingsAPI.Use(middleware.AdminAuthMiddleware())
				{
					settingsAPI.GET("", settingsController.GetSettings)
					settingsAPI.PUT("", settingsController.UpdateSettings)
				}
			}
		}
	}

	port := ":8080"
	log.Printf("Server siap dijalankan di port %s", port)
	if err := router.Run(port); err != nil {
		log.Fatal(err)
	}
}