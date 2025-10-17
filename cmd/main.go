/**
* FILE HEADER: cmd/main.go
*
* PURPOSE:
* Titik masuk utama (entrypoint) untuk aplikasi SIMDOKPOL.
* Bertanggung jawab untuk:
* 1. Menjalankan aplikasi sebagai ikon di system tray.
* 2. Menampilkan notifikasi saat startup.
* 3. Menyediakan menu untuk membuka aplikasi dan keluar.
* 4. Menjalankan server web Gin di background.
*/
package main

import (
  "errors"
  "fmt"
  "html/template"
  "io/ioutil"
  "log"
  "net/http"
  "os/exec"
  "runtime"
  "simdokpol/internal/config"
  "simdokpol/internal/controllers"
  "simdokpol/internal/middleware"
  "simdokpol/internal/repositories"
  "simdokpol/internal/services"
  "strconv"
  "strings"
  "time"

  _ "simdokpol/docs"
  _ "time/tzdata"

  "github.com/gin-gonic/gin"
  "github.com/getlantern/systray"
  "github.com/golang-migrate/migrate/v4"
  "github.com/golang-migrate/migrate/v4/database/sqlite"
  _ "github.com/golang-migrate/migrate/v4/source/file"
  swaggerFiles "github.com/swaggo/files"
  ginSwagger "github.com/swaggo/gin-swagger"
  gormsqlite "gorm.io/driver/sqlite"
  "gorm.io/gorm"
  "gorm.io/gorm/logger"
)

const (
  port = ":8080"
  url = "http://localhost:8080"
)

// main sekarang menjadi entrypoint untuk aplikasi system tray.
func main() {
  systray.Run(onReady, onExit)
}

// onReady adalah fungsi yang akan dijalankan saat ikon tray siap.
// Di sinilah semua logika utama aplikasi kita akan dimulai.
func onReady() {
  // Setup ikon dan tooltip
  systray.SetIcon(getIcon("web/static/img/icon.png"))
  systray.SetTitle("SIMDOKPOL")
  systray.SetTooltip("Sistem Informasi Manajemen Dokumen Kepolisian")

  // Tambahkan menu items
  mOpen: = systray.AddMenuItem("Buka Aplikasi", "Buka SIMDOKPOL di browser")
  systray.AddSeparator()
  mQuit: = systray.AddMenuItem("Keluar", "Tutup aplikasi")

  // Jalankan server web di sebuah goroutine agar tidak memblokir UI tray.
  go startWebServer()

  // Tampilkan notifikasi startup.
  systray.ShowNotification("SIMDOKPOL", "Aplikasi sedang berjalan di latar belakang.")

  // Buka browser secara otomatis saat pertama kali dijalankan.
  openBrowser(url)

  // Loop untuk menangani event klik pada menu.
  go func() {
    for {
      select {
        case <-mOpen.ClickedCh:
          // Jika menu "Buka Aplikasi" diklik, panggil fungsi openBrowser.
          openBrowser(url)
        case <-mQuit.ClickedCh:
          // Jika menu "Keluar" diklik, hentikan aplikasi tray,
          // yang juga akan memicu onExit.
          systray.Quit()
      }
    }
  }()
}

// onExit akan dipanggil saat aplikasi ditutup.
func onExit() {
  log.Println("INFO: Aplikasi SIMDOKPOL ditutup.")
}

// startWebServer berisi semua logika setup dan run server Gin.
// Ini adalah isi dari fungsi main() Anda sebelumnya.
func startWebServer() {
  cfg,
  err: = config.Load()
  if err != nil {
    log.Fatalf("FATAL: Gagal memuat konfigurasi: %v", err)
  }

  db,
  err: = setupDatabase(cfg.DBDSN)
  if err != nil {
    log.Fatalf("FATAL: Gagal terhubung ke database: %v", err)
  }

  repos,
  svcs,
  ctrls: = setupDependencies(db, cfg)
  router: = setupRouter(repos.UserRepo, svcs, ctrls)

  log.Printf("INFO: Server web dimulai di %s", url)
  if err: = router.Run(port); err != nil {
    log.Fatalf("FATAL: Gagal menjalankan server: %v", err)
  }
}

// getIcon adalah helper untuk membaca file ikon dari disk.
func getIcon(s string) []byte {
  b,
  err: = ioutil.ReadFile(s)
  if err != nil {
    log.Printf("PERINGATAN: Gagal membaca file ikon: %v", err)
    return nil
  }
  return b
}

// openBrowser membuka URL di browser default pengguna.
func openBrowser(url string) {
  var err error
  time.Sleep(1 * time.Second) // Beri jeda agar server siap

  switch runtime.GOOS {
  case "linux":
    err = exec.Command("xdg-open", url).Start()
  case "windows":
    err = exec.Command("cmd", "/c", "start", url).Start()
  case "darwin":
    err = exec.Command("open", url).Start()
  default:
    err = fmt.Errorf("sistem operasi tidak didukung: %s", runtime.GOOS)
  }

  if err != nil {
    log.Printf("PERINGATAN: Gagal membuka browser secara otomatis: %v", err)
  }
}

// --- SEMUA FUNGSI SETUP LAINNYA TETAP SAMA ---

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

// setupDatabase, setupDependencies, setupRouter, setupPageRoutes, setupAPIRoutes,
// dan struct Repositories, Services, Controllers tidak perlu diubah.
// (Kode di bawah ini sama persis dengan versi sebelumnya)

func setupDatabase(dsn string) (*gorm.DB, error) {
  db,
  err: = gorm.Open(gormsqlite.Open(dsn), &gorm.Config {
    Logger: logger.Default.LogMode(logger.Info)})
  if err != nil {
    return nil,
    err
  }
  sqlDB,
  err: = db.DB(); if err != nil {
    return nil,
    err
  }
  driver,
  err: = sqlite.WithInstance(sqlDB, &sqlite.Config {}); if err != nil {
    return nil,
    err
  }
  m,
  err: = migrate.NewWithDatabaseInstance("file://migrations", "sqlite", driver); if err != nil {
    return nil,
    err
  }
  if err: = m.Up(); err != nil && err != migrate.ErrNoChange {
    return nil,
    err
  }
  log.Println("INFO: Migrasi database berhasil dijalankan."); return db,
  nil
}

func setupDependencies(db *gorm.DB, cfg *config.Config) (Repositories, Services, Controllers) {
  userRepo: = repositories.NewUserRepository(db); residentRepo: = repositories.NewResidentRepository(db); docRepo: = repositories.NewLostDocumentRepository(db); configRepo: = repositories.NewConfigRepository(db); auditRepo: = repositories.NewAuditLogRepository(db)
  services.JWTSecretKey = []byte(cfg.JWTSecretKey); configService: = services.NewConfigService(configRepo); auditService: = services.NewAuditLogService(auditRepo); authService: = services.NewAuthService(userRepo); dashboardService: = services.NewDashboardService(docRepo, userRepo, configService); docService: = services.NewLostDocumentService(db, docRepo, residentRepo, userRepo, auditService, configService); userService: = services.NewUserService(userRepo, auditService, cfg); backupService: = services.NewBackupService(cfg, configService, auditService)
  authController: = controllers.NewAuthController(authService); dashboardController: = controllers.NewDashboardController(dashboardService); docController: = controllers.NewLostDocumentController(docService); userController: = controllers.NewUserController(userService); configController: = controllers.NewConfigController(configService, userService); auditController: = controllers.NewAuditLogController(auditService); backupController: = controllers.NewBackupController(backupService); settingsController: = controllers.NewSettingsController(configService, auditService)
  return Repositories {
    UserRepo: userRepo
  },
  Services {
    ConfigService: configService,
    DocService: docService
  },
  Controllers {
    AuthController: authController,
    DashboardController: dashboardController,
    DocController: docController,
    UserController: userController,
    ConfigController: configController,
    AuditController: auditController,
    BackupController: backupController,
    SettingsController: settingsController
  }
}

func setupRouter(userRepo repositories.UserRepository, svcs Services, ctrls Controllers) *gin.Engine {
  router: = gin.Default(); funcMap: = template.FuncMap {
    "ToUpper": strings.ToUpper
  }; templates: = template.New("").Funcs(funcMap); templates = template.Must(templates.ParseGlob("web/templates/*.html")); templates = template.Must(templates.ParseGlob("web/templates/partials/*.html")); router.SetHTMLTemplate(templates); router.Static("/static", "./web/static"); router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)); router.GET("/setup", ctrls.ConfigController.ShowSetupPage); router.POST("/api/setup", ctrls.ConfigController.SaveSetup); app: = router.Group(""); app.Use(middleware.SetupMiddleware(svcs.ConfigService)); {
    app.GET("/login", func(c *gin.Context) {
      c.HTML(http.StatusOK, "login.html", gin.H {
        "Title": "Login"
      })
    }); app.POST("/api/login", ctrls.AuthController.Login); app.POST("/api/logout", ctrls.AuthController.Logout); protected: = app.Group(""); protected.Use(middleware.AuthMiddleware(userRepo)); {
      setupPageRoutes(protected, svcs); setupAPIRoutes(protected, ctrls)
    }
  }; return router
}

func setupPageRoutes(router *gin.RouterGroup, svcs Services) {
  getUser: = func(c *gin.Context) interface {} {
    user,
    _: = c.Get("currentUser"); return user
  }; router.GET("/", func(c *gin.Context) {
      c.HTML(http.StatusOK, "dashboard.html", gin.H {
        "Title": "Dasbor", "CurrentUser": getUser(c)})
  }); router.GET("/documents", func(c *gin.Context) {
    c.HTML(http.StatusOK, "document_list.html", gin.H {
      "Title": "Daftar Dokumen Aktif", "CurrentUser": getUser(c), "PageType": "active"
    })
  }); router.GET("/documents/archived", func(c *gin.Context) {
    c.HTML(http.StatusOK, "document_list.html", gin.H {
      "Title": "Arsip Dokumen", "CurrentUser": getUser(c), "PageType": "archived"
    })
  }); router.GET("/documents/new", func(c *gin.Context) {
    c.HTML(http.StatusOK, "document_form.html", gin.H {
      "Title": "Buat Surat Baru", "CurrentUser": getUser(c), "IsEdit": false, "DocID": 0
    })
  }); router.GET("/documents/:id/edit", func(c *gin.Context) {
    id: = c.Param("id"); c.HTML(http.StatusOK, "document_form.html", gin.H {
      "Title": "Edit Surat", "CurrentUser": getUser(c), "IsEdit": true, "DocID": id
    })
  }); router.GET("/search", func(c *gin.Context) {
    query: = c.Query("q"); c.HTML(http.StatusOK, "search_results.html", gin.H {
      "Title": "Hasil Pencarian", "CurrentUser": getUser(c), "Query": query
    })
  }); router.GET("/profile", func(c *gin.Context) {
    c.HTML(http.StatusOK, "profile.html", gin.H {
      "Title": "Profil Pengguna", "CurrentUser": getUser(c)})
  }); router.GET("/panduan", func(c *gin.Context) {
    c.HTML(http.StatusOK, "panduan.html", gin.H {
      "Title": "Panduan Pengguna", "CurrentUser": getUser(c)})
  }); router.GET("/tentang", func(c *gin.Context) {
    c.HTML(http.StatusOK, "tentang.html", gin.H {
      "Title": "Tentang Aplikasi", "CurrentUser": getUser(c)})
  }); router.GET("/documents/:id/print", func(c *gin.Context) {
    id, err: = strconv.ParseUint(c.Param("id"), 10, 32); if err != nil {
      c.HTML(http.StatusBadRequest, "error.html", gin.H {
        "Title": "Error", "CurrentUser": getUser(c), "ErrorMessage": "ID Dokumen tidak valid."
    }); return
  }; doc,
  err: = svcs.DocService.FindByID(uint(id), c.GetUint("userID")); if err != nil {
    status: = http.StatusNotFound; message: = "Dokumen tidak ditemukan."; if errors.Is(err, services.ErrAccessDenied) {
      status = http.StatusForbidden; message = "Anda tidak memiliki izin untuk mengakses dokumen ini."
    }; c.HTML(status, "error.html", gin.H {
        "Title": "Error", "CurrentUser": getUser(c), "ErrorMessage": message
      }); return
  }; appConfig,
  err: = svcs.ConfigService.GetConfig(); if err != nil {
    c.HTML(http.StatusInternalServerError, "error.html", gin.H {
      "Title": "Error", "CurrentUser": getUser(c), "ErrorMessage": "Gagal memuat konfigurasi aplikasi."
    }); return
  }; c.HTML(http.StatusOK, "print_preview.html", gin.H {
      "Document": doc, "Now": time.Now(), "CurrentUser": getUser(c), "Config": appConfig
    })
}); adminRoutes: = router.Group(""); adminRoutes.Use(middleware.AdminAuthMiddleware()); {
adminRoutes.GET("/users",
  func(c *gin.Context) {
    c.HTML(http.StatusOK, "user_list.html", gin.H {
      "Title": "Manajemen Pengguna", "CurrentUser": getUser(c)})
  }); adminRoutes.GET("/users/new",
  func(c *gin.Context) {
    c.HTML(http.StatusOK, "user_form.html", gin.H {
      "Title": "Tambah Pengguna", "CurrentUser": getUser(c), "IsEdit": false, "UserID": 0
    })
  }); adminRoutes.GET("/users/:id/edit",
  func(c *gin.Context) {
    id,
    _: = strconv.Atoi(c.Param("id")); c.HTML(http.StatusOK, "user_form.html", gin.H {
      "Title": "Edit Pengguna", "CurrentUser": getUser(c), "IsEdit": true, "UserID": id
    })
  }); adminRoutes.GET("/audit-logs",
  func(c *gin.Context) {
    c.HTML(http.StatusOK, "audit_log_list.html", gin.H {
      "Title": "Log Audit Sistem", "CurrentUser": getUser(c)})
  }); adminRoutes.GET("/settings",
  func(c *gin.Context) {
    c.HTML(http.StatusOK, "settings.html", gin.H {
      "Title": "Pengaturan Sistem", "CurrentUser": getUser(c)})
  })
}
}

func setupAPIRoutes(router *gin.RouterGroup, ctrls Controllers) {
api: = router.Group("/api"); {
api.GET("/stats",
ctrls.DashboardController.GetStats); api.GET("/stats/monthly-issuance",
ctrls.DashboardController.GetMonthlyChart); api.GET("/stats/item-composition",
ctrls.DashboardController.GetItemCompositionChart); api.GET("/notifications/expiring-documents",
ctrls.DashboardController.GetExpiringDocuments); api.PUT("/profile",
ctrls.UserController.UpdateProfile); api.PUT("/profile/password",
ctrls.UserController.ChangePassword); api.GET("/search",
ctrls.DocController.SearchGlobal); api.POST("/documents",
ctrls.DocController.Create); api.GET("/documents",
ctrls.DocController.FindAll); api.GET("/documents/:id",
ctrls.DocController.FindByID); api.PUT("/documents/:id",
ctrls.DocController.Update); api.DELETE("/documents/:id",
ctrls.DocController.Delete); adminAPI: = api.Group(""); adminAPI.Use(middleware.AdminAuthMiddleware()); {
adminAPI.GET("/users/operators",
ctrls.UserController.FindOperators); adminAPI.POST("/users",
ctrls.UserController.Create); adminAPI.GET("/users",
ctrls.UserController.FindAll); adminAPI.GET("/users/:id",
ctrls.UserController.FindByID); adminAPI.PUT("/users/:id",
ctrls.UserController.Update); adminAPI.DELETE("/users/:id",
ctrls.UserController.Delete); adminAPI.POST("/users/:id/activate",
ctrls.UserController.Activate); adminAPI.GET("/audit-logs",
ctrls.AuditController.FindAll); adminAPI.POST("/backups",
ctrls.BackupController.CreateBackup); adminAPI.POST("/restore",
ctrls.BackupController.RestoreBackup); adminAPI.GET("/settings",
ctrls.SettingsController.GetSettings); adminAPI.PUT("/settings",
ctrls.SettingsController.UpdateSettings)
}
}
}

type Repositories struct {
UserRepo repositories.UserRepository
}
type Services struct {
ConfigService services.ConfigService; DocService services.LostDocumentService
}
type Controllers struct {
AuthController *controllers.AuthController; DashboardController *controllers.DashboardController; DocController *controllers.LostDocumentController; UserController *controllers.UserController; ConfigController *controllers.ConfigController; AuditController *controllers.AuditLogController; BackupController *controllers.BackupController; SettingsController *controllers.SettingsController
}