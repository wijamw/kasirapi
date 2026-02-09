package repositories

import (
	"database/sql"
	"fmt"
	"kasir-api/models"
	"time"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repo *TransactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	var (
		res *models.Transaction
	)

	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	totalAmount := 0
	// inisialisasi modeling transactionDetails -> nanti kita insert ke db
	details := make([]models.TransactionDetail, 0)
	// loop setiap item
	for _, item := range items {
		var catName, catDesc string
		var catID int
		// get product dapet pricing
		err := tx.QueryRow("SELECT id, name, description FROM categories WHERE id=$1", item.CategoryID).Scan(&catID, &catName, &catDesc)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Category id %d not found", item.CategoryID)
		}

		if err != nil {
			return nil, err
		}

		// tambah jumlah pick
		totalAmount = item.Quantity

		// item nya dimasukkin ke transactionDetails
		details = append(details, models.TransactionDetail{
			CategoryID:   catID,
			CategoryName: catName,
			Quantity:    item.Quantity,
		})
	}

	// insert transaction
	var transactionID int
	err = tx.QueryRow("INSERT INTO transactions (total_amount) VALUES ($1) RETURNING ID", totalAmount).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	// insert transaction details
	for i := range details {
		details[i].TransactionID = transactionID
		_, err := tx.Exec("INSERT INTO transaction_details (transaction_id, category_id, quantity) VALUES ($1,$2,$3)", transactionID, details[i].CategoryID, details[i].Quantity)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	res = &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		Details:     details,
	}

	return res, nil
}

func (repo *TransactionRepository) GetDailyReport() (*models.DailyCategoryReport, error) {
    var (
        report *models.DailyCategoryReport
    )

    // Get today's date range in WIB (UTC+7)
    // Start: today 00:00:00 WIB, End: today 23:59:59 WIB
    todayStart := time.Now().UTC().Add(7 * time.Hour).Truncate(24 * time.Hour).Add(-7 * time.Hour)
    todayEnd := todayStart.Add(24 * time.Hour)

    // 1. Count total transactions for today
	var totalTransaction int
    err := repo.db.QueryRow(`
        SELECT COUNT(*) 
        FROM transactions 
        WHERE created_at >= $1 AND created_at < $2
    `, todayStart, todayEnd).Scan(&report.TotalTransaction)
    
    if err != nil {
        return nil, fmt.Errorf("failed to count transactions: %w", err)
    }

    // 2. Get most picked category with its total quantity
	var mostPickedCategory string
    var totalPicked int
    err = repo.db.QueryRow(`
        SELECT c.name, SUM(td.quantity) as total
        FROM transaction_details td
        JOIN transactions t ON td.transaction_id = t.id
        JOIN categories c ON td.category_id = c.id
        WHERE t.created_at >= $1 AND t.created_at < $2
        GROUP BY c.id, c.name
        ORDER BY total DESC
        LIMIT 1
    `, todayStart, todayEnd).Scan(&report.MostPickedCategory, &report.TotalPicked)
    
    if err == sql.ErrNoRows {
        // No transactions today, return zeros
        mostPickedCategory = ""
        totalPicked = 0
    } else if err != nil {
        return nil, fmt.Errorf("failed to get most picked category: %w", err)
    }

	report = &models.DailyCategoryReport{
        TotalTransaction:   totalTransaction,
        MostPickedCategory: mostPickedCategory,
        TotalPicked:        totalPicked,
    }

    return report, nil
}