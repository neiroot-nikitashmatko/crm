package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

var productionCategoryByNomenclature = map[string]string{
	"Перетяжка руля":    "Перетяжка",
	"Установка чехлов":  "Установка",
	"Установка накидок": "Установка",
	"Полировка фар":     "Стёкла",
	"Полировка стёкол":  "Стёкла",
	"Ремонт стёкол":     "Стёкла",
	"Пошив ковриков":    "Коврики",
}

var productionShareCategories = []string{
	"Перетяжка",
	"Установка",
	"Стёкла",
	"Коврики",
}

const productionShareOtherCategory = "Прочее"

func productionCategoryForNomenclature(nomenclature string) string {
	trimmed := strings.TrimSpace(nomenclature)
	if category, ok := productionCategoryByNomenclature[trimmed]; ok {
		return category
	}
	return productionShareOtherCategory
}

var (
	ErrAnalyticsPeriodRequired = errors.New("укажите период")
	ErrInvalidAnalyticsPeriod  = errors.New("некорректный период")
)

type AnalyticsService struct {
	repo *repository.AnalyticsRepository
}

func NewAnalyticsService(repo *repository.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) parsePeriod(fromMs int64, toMs int64) (time.Time, time.Time, error) {
	if fromMs <= 0 || toMs <= 0 {
		return time.Time{}, time.Time{}, ErrAnalyticsPeriodRequired
	}
	if fromMs > toMs {
		return time.Time{}, time.Time{}, ErrInvalidAnalyticsPeriod
	}
	return time.UnixMilli(fromMs), time.UnixMilli(toMs), nil
}

func (s *AnalyticsService) LeadsTraffic(ctx context.Context, fromMs int64, toMs int64) ([]model.TrafficSourceMetric, error) {
	from, to, err := s.parsePeriod(fromMs, toMs)
	if err != nil {
		return nil, err
	}
	return s.repo.CountLeadsByTrafficSource(ctx, from, to)
}

func (s *AnalyticsService) DealsTraffic(ctx context.Context, fromMs int64, toMs int64) ([]model.TrafficSourceMetric, error) {
	from, to, err := s.parsePeriod(fromMs, toMs)
	if err != nil {
		return nil, err
	}
	return s.repo.CountDealsByTrafficSource(ctx, from, to)
}

func (s *AnalyticsService) LeadToDealConversion(ctx context.Context, fromMs int64, toMs int64) (model.LeadToDealConversion, error) {
	from, to, err := s.parsePeriod(fromMs, toMs)
	if err != nil {
		return model.LeadToDealConversion{}, err
	}

	leadsCount, convertedCount, err := s.repo.CountLeadToDealConversion(ctx, from, to)
	if err != nil {
		return model.LeadToDealConversion{}, err
	}

	result := model.LeadToDealConversion{
		LeadsCount:     leadsCount,
		ConvertedCount: convertedCount,
	}
	if leadsCount > 0 {
		result.Percent = math.Round(float64(convertedCount)*1000/float64(leadsCount)) / 10
	}
	return result, nil
}

func (s *AnalyticsService) FailedLeadShare(ctx context.Context, fromMs int64, toMs int64) (model.FailedLeadShare, error) {
	from, to, err := s.parsePeriod(fromMs, toMs)
	if err != nil {
		return model.FailedLeadShare{}, err
	}

	leadsCount, failedCount, err := s.repo.CountFailedLeadShare(ctx, from, to)
	if err != nil {
		return model.FailedLeadShare{}, err
	}

	result := model.FailedLeadShare{
		LeadsCount:  leadsCount,
		FailedCount: failedCount,
	}
	if leadsCount > 0 {
		result.Percent = math.Round(float64(failedCount)*1000/float64(leadsCount)) / 10
	}
	return result, nil
}

func (s *AnalyticsService) FailedDealShare(ctx context.Context, fromMs int64, toMs int64) (model.FailedDealShare, error) {
	from, to, err := s.parsePeriod(fromMs, toMs)
	if err != nil {
		return model.FailedDealShare{}, err
	}

	dealsCount, failedCount, err := s.repo.CountFailedDealShare(ctx, from, to)
	if err != nil {
		return model.FailedDealShare{}, err
	}

	result := model.FailedDealShare{
		DealsCount:  dealsCount,
		FailedCount: failedCount,
	}
	if dealsCount > 0 {
		result.Percent = math.Round(float64(failedCount)*1000/float64(dealsCount)) / 10
	}
	return result, nil
}

func (s *AnalyticsService) ClosedDealsProductionShare(
	ctx context.Context,
	fromMs int64,
	toMs int64,
) ([]model.ProductionCategoryMetric, error) {
	from, to, err := s.parsePeriod(fromMs, toMs)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.CountClosedDealsByNomenclature(ctx, from, to)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int, len(productionShareCategories)+1)
	for _, category := range productionShareCategories {
		counts[category] = 0
	}

	otherCount := 0
	for _, row := range rows {
		nomenclature := strings.TrimSpace(row.Nomenclature)
		if category, ok := productionCategoryByNomenclature[nomenclature]; ok {
			counts[category] += row.Count
			continue
		}
		otherCount += row.Count
	}

	items := make([]model.ProductionCategoryMetric, 0, len(productionShareCategories)+1)
	for _, category := range productionShareCategories {
		items = append(items, model.ProductionCategoryMetric{
			Category: category,
			Count:    counts[category],
		})
	}
	if otherCount > 0 {
		items = append(items, model.ProductionCategoryMetric{
			Category: productionShareOtherCategory,
			Count:    otherCount,
		})
	}

	return items, nil
}

func (s *AnalyticsService) ClosedDealsEmployeeShare(
	ctx context.Context,
	fromMs int64,
	toMs int64,
) ([]model.EmployeeShareMetric, error) {
	from, to, err := s.parsePeriod(fromMs, toMs)
	if err != nil {
		return nil, err
	}
	return s.repo.CountClosedDealsByEmployee(ctx, from, to)
}

func (s *AnalyticsService) ClosedDealsList(
	ctx context.Context,
	fromMs int64,
	toMs int64,
	requireEmployee bool,
	requireProduction bool,
) ([]model.ClosedDealListItem, error) {
	from, to, err := s.parsePeriod(fromMs, toMs)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListClosedDeals(ctx, from, to, requireEmployee, requireProduction)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].Category = productionCategoryForNomenclature(items[index].Nomenclature)
	}
	return items, nil
}
