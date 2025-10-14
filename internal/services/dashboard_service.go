package services

import (
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"time"
)

type DashboardStatsDTO struct {
	DocsMonthly int64 `json:"docs_monthly"`
	DocsYearly  int64 `json:"docs_yearly"`
	DocsToday   int64 `json:"docs_today"`
	ActiveUsers int64 `json:"active_users"`
}

type ChartDataDTO struct {
	Labels []string `json:"labels"`
	Data   []int    `json:"data"`
}

type PieChartDataDTO struct {
	Labels           []string `json:"labels"`
	Data             []int    `json:"data"`
	BackgroundColors []string `json:"background_colors"`
}

type DashboardService interface {
	GetDashboardStats() (*DashboardStatsDTO, error)
	GetMonthlyIssuanceChartData() (*ChartDataDTO, error)
	GetItemCompositionPieChartData() (*PieChartDataDTO, error)
	GetExpiringDocumentsForUser(userID uint, notificationWindowDays int) ([]models.LostDocument, error) // <-- METHOD BARU
}

type dashboardService struct {
	docRepo       repositories.LostDocumentRepository
	userRepo      repositories.UserRepository
	configService ConfigService
}

func NewDashboardService(docRepo repositories.LostDocumentRepository, userRepo repositories.UserRepository, configService ConfigService) DashboardService {
	return &dashboardService{
		docRepo:       docRepo,
		userRepo:      userRepo,
		configService: configService,
	}
}

// === FUNGSI BARU UNTUK NOTIFIKASI ===
func (s *dashboardService) GetExpiringDocumentsForUser(userID uint, notificationWindowDays int) ([]models.LostDocument, error) {
	appConfig, err := s.configService.GetConfig()
	if err != nil {
		return nil, err
	}
	loc, err := s.configService.GetLocation()
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)

	archiveDuration := time.Duration(appConfig.ArchiveDurationDays) * 24 * time.Hour
	notificationWindow := time.Duration(notificationWindowDays) * 24 * time.Hour

	// Menghitung rentang waktu. Kita mencari dokumen yang tanggal laporannya berada di antara:
	// Awal rentang: (Sekarang - Durasi Arsip) -> Misal, 15 hari yang lalu
	// Akhir rentang: (Sekarang - Durasi Arsip + Jendela Notifikasi) -> Misal, 12 hari yang lalu (15-3)
	// Artinya kita mencari dokumen yang dibuat antara 15 s/d 12 hari yang lalu.
	expiryDateStart := now.Add(-archiveDuration)
	expiryDateEnd := expiryDateStart.Add(notificationWindow)

	return s.docRepo.FindExpiringDocumentsForUser(userID, expiryDateStart, expiryDateEnd)
}
// === AKHIR FUNGSI BARU ===

func (s *dashboardService) GetDashboardStats() (*DashboardStatsDTO, error) {
	loc, err := s.configService.GetLocation()
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)

	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.AddDate(0, 0, 1).Add(-time.Nanosecond)
	docsToday, err := s.docRepo.CountByDateRange(startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)
	docsMonthly, err := s.docRepo.CountByDateRange(startOfMonth, endOfMonth)
	if err != nil {
		return nil, err
	}

	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
	endOfYear := startOfYear.AddDate(1, 0, 0).Add(-time.Nanosecond)
	docsYearly, err := s.docRepo.CountByDateRange(startOfYear, endOfYear)
	if err != nil {
		return nil, err
	}

	activeUsers, err := s.userRepo.CountAll()
	if err != nil {
		return nil, err
	}

	stats := &DashboardStatsDTO{
		DocsMonthly: docsMonthly,
		DocsYearly:  docsYearly,
		DocsToday:   docsToday,
		ActiveUsers: activeUsers,
	}

	return stats, nil
}

func (s *dashboardService) GetMonthlyIssuanceChartData() (*ChartDataDTO, error) {
	loc, err := s.configService.GetLocation()
	if err != nil {
		loc = time.UTC
	}
	currentYear := time.Now().In(loc).Year()

	counts, err := s.docRepo.GetMonthlyIssuanceForYear(currentYear)
	if err != nil {
		return nil, err
	}

	labels := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Ags", "Sep", "Okt", "Nov", "Des"}
	data := make([]int, 12)

	for _, count := range counts {
		if count.Month >= 1 && count.Month <= 12 {
			data[count.Month-1] = count.Count
		}
	}

	return &ChartDataDTO{Labels: labels, Data: data}, nil
}

func (s *dashboardService) GetItemCompositionPieChartData() (*PieChartDataDTO, error) {
	stats, err := s.docRepo.GetItemCompositionStats()
	if err != nil {
		return nil, err
	}

	var labels []string
	var data []int
	colors := []string{"#4e73df", "#1cc88a", "#36b9cc", "#f6c23e", "#e74a3b", "#858796"}

	limit := 5
	othersCount := 0
	for i, stat := range stats {
		if i < limit {
			labels = append(labels, stat.NamaBarang)
			data = append(data, stat.Count)
		} else {
			othersCount += stat.Count
		}
	}

	if othersCount > 0 {
		labels = append(labels, "Lainnya")
		data = append(data, othersCount)
	}

	finalColors := colors[:len(labels)]

	return &PieChartDataDTO{
		Labels:           labels,
		Data:             data,
		BackgroundColors: finalColors,
	}, nil
}