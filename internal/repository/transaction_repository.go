package repository

import (
	"context"
	"expense-tracker/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(ctx context.Context, t *models.Transaction) error
	FindByID(ctx context.Context, id, userID uuid.UUID) (*models.Transaction, error)
	List(ctx context.Context, userID uuid.UUID, f *models.TransactionFilter) ([]models.Transaction, int64, error)
	Update(ctx context.Context, t *models.Transaction) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	Summary(ctx context.Context, userID uuid.UUID, from, to time.Time) (*SummaryResult, error)
	DailyTotals(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]DailyTotal, error)
	CategoryBreakdown(ctx context.Context, userID uuid.UUID, from, to time.Time, txType models.TransactionType) ([]CategoryTotal, error)
	MonthlyTrends(ctx context.Context, userID uuid.UUID, months int) ([]MonthlyTrend, error)
}

type SummaryResult struct {
	TotalIncome  float64 `json:"totalIncome"`
	TotalExpense float64 `json:"totalExpense"`
	NetBalance   float64 `json:"netBalance"`
	SavingsRate  float64 `json:"savingsRate"`
}

type DailyTotal struct {
	Date    time.Time `json:"date"`
	Income  float64   `json:"income"`
	Expense float64   `json:"expense"`
}

type CategoryTotal struct {
	CategoryID    uuid.UUID `json:"categoryId"`
	CategoryName  string    `json:"categoryName"`
	CategoryIcon  string    `json:"categoryIcon"`
	CategoryColor string    `json:"categoryColor"`
	Total         float64   `json:"total"`
	Count         int64     `json:"count"`
	Percentage    float64   `json:"percentage"`
}

type MonthlyTrend struct {
	Year    int     `json:"year"`
	Month   int     `json:"month"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, t *models.Transaction) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *transactionRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*models.Transaction, error) {
	var t models.Transaction
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("id = ? AND user_id = ?", id, userID).
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *transactionRepository) List(ctx context.Context, userID uuid.UUID, f *models.TransactionFilter) ([]models.Transaction, int64, error) {
	q := r.db.WithContext(ctx).
		Preload("Category").
		Where("user_id = ?", userID)

	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	if f.CategoryID != nil {
		q = q.Where("category_id = ?", *f.CategoryID)
	}
	if f.FromDate != nil {
		q = q.Where("date >= ?", *f.FromDate)
	}
	if f.ToDate != nil {
		q = q.Where("date <= ?", *f.ToDate)
	}
	if f.Search != "" {
		q = q.Where("description ILIKE ?", "%"+f.Search+"%")
	}

	var total int64
	if err := q.Model(&models.Transaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var transactions []models.Transaction
	err := q.Order("date DESC, created_at DESC").
		Limit(f.PerPage).
		Offset(f.Offset()).
		Find(&transactions).Error

	return transactions, total, err
}

func (r *transactionRepository) Update(ctx context.Context, t *models.Transaction) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *transactionRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.Transaction{}).Error
}

func (r *transactionRepository) Summary(ctx context.Context, userID uuid.UUID, from, to time.Time) (*SummaryResult, error) {
	type row struct {
		Type  models.TransactionType
		Total float64
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Select("type, COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, from, to).
		Group("type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := &SummaryResult{}
	for _, row := range rows {
		if row.Type == models.TransactionIncome {
			result.TotalIncome = row.Total
		} else {
			result.TotalExpense = row.Total
		}
	}
	result.NetBalance = result.TotalIncome - result.TotalExpense
	if result.TotalIncome > 0 {
		result.SavingsRate = (result.NetBalance / result.TotalIncome) * 100
	}
	return result, nil
}

func (r *transactionRepository) DailyTotals(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]DailyTotal, error) {
	var rows []struct {
		Date    time.Time
		Type    models.TransactionType
		Total   float64
	}
	err := r.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Select("DATE(date) as date, type, COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, from, to).
		Group("DATE(date), type").
		Order("DATE(date)").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	dailyMap := make(map[time.Time]*DailyTotal)
	for _, row := range rows {
		d := row.Date.UTC().Truncate(24 * time.Hour)
		if _, ok := dailyMap[d]; !ok {
			dailyMap[d] = &DailyTotal{Date: d}
		}
		if row.Type == models.TransactionIncome {
			dailyMap[d].Income = row.Total
		} else {
			dailyMap[d].Expense = row.Total
		}
	}

	result := make([]DailyTotal, 0, len(dailyMap))
	for _, v := range dailyMap {
		result = append(result, *v)
	}
	return result, nil
}

func (r *transactionRepository) CategoryBreakdown(ctx context.Context, userID uuid.UUID, from, to time.Time, txType models.TransactionType) ([]CategoryTotal, error) {
	var rows []CategoryTotal
	err := r.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Select(`
			transactions.category_id,
			categories.name as category_name,
			categories.icon as category_icon,
			categories.color as category_color,
			COALESCE(SUM(transactions.amount), 0) as total,
			COUNT(*) as count
		`).
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Where("transactions.user_id = ? AND transactions.type = ? AND transactions.date BETWEEN ? AND ?", userID, txType, from, to).
		Group("transactions.category_id, categories.name, categories.icon, categories.color").
		Order("total DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	var grandTotal float64
	for _, row := range rows {
		grandTotal += row.Total
	}
	for i := range rows {
		if grandTotal > 0 {
			rows[i].Percentage = (rows[i].Total / grandTotal) * 100
		}
	}
	return rows, nil
}

func (r *transactionRepository) MonthlyTrends(ctx context.Context, userID uuid.UUID, months int) ([]MonthlyTrend, error) {
	var rows []struct {
		Year    int
		Month   int
		Type    models.TransactionType
		Total   float64
	}
	err := r.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Select("EXTRACT(YEAR FROM date)::int as year, EXTRACT(MONTH FROM date)::int as month, type, COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND date >= NOW() - INTERVAL '1 month' * ?", userID, months).
		Group("year, month, type").
		Order("year, month").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	type key struct{ Year, Month int }
	trendMap := make(map[key]*MonthlyTrend)
	for _, row := range rows {
		k := key{row.Year, row.Month}
		if _, ok := trendMap[k]; !ok {
			trendMap[k] = &MonthlyTrend{Year: row.Year, Month: row.Month}
		}
		if row.Type == models.TransactionIncome {
			trendMap[k].Income = row.Total
		} else {
			trendMap[k].Expense = row.Total
		}
	}

	result := make([]MonthlyTrend, 0, len(trendMap))
	for _, v := range trendMap {
		result = append(result, *v)
	}
	return result, nil
}
