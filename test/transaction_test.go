package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"back/internal/features/transaction/domain"
	"back/internal/features/transaction/handler"
	"back/internal/features/transaction/usecase"
	"back/internal/shared/test"
)

var (
	testOnce sync.Once
	testEcho *echo.Echo
	testDB   *test.DB
)

func TestMain(m *testing.M) {
	testOnce.Do(func() {
		logger, _ := zap.NewDevelopment()
		zap.ReplaceGlobals(logger)
		defer func() {
			_ = logger.Sync()
		}()

		var err error
		testDB, err = test.SetupTestDB()
		if err != nil {
			zap.L().Fatal("Failed to setup test database", zap.Error(err))
		}

		uc := usecase.NewTransactionUseCase(testDB.BalanceRepo, testDB.LedgerRepo, testDB.WithdrawalRepo, testDB.TxManager)
		h := handler.NewTransactionHandler(uc)

		testEcho = echo.New()
		testEcho.POST("/withdrawals", func(c *echo.Context) error {
			c.Set("user_id", 1)
			return h.CreateWithdrawal(c)
		})
		testEcho.POST("/withdrawals/:id/confirm", func(c *echo.Context) error {
			c.Set("user_id", 1)
			return h.ConfirmWithdrawal(c)
		})
		testEcho.GET("/withdrawals/:id", func(c *echo.Context) error {
			c.Set("user_id", 1)
			return h.GetWithdrawal(c)
		})
	})

	code := m.Run()
	os.Exit(code)
}

func beforeTest(t *testing.T) {
	// Clear DB:
	require.NoError(t, testDB.DB.Exec(`
		TRUNCATE balances, withdrawals, ledger_entries, users RESTART IDENTITY CASCADE
	`).Error)

	// Create user:
	require.NoError(t, testDB.DB.Exec(`
		INSERT INTO users (username)
		VALUES ('test_user')
	`).Error)
}

func setupTestServer(t *testing.T) (*echo.Echo, *test.DB) {
	beforeTest(t)
	return testEcho, testDB
}

func TestCreateWithdrawalEndpoint_Success(t *testing.T) {
	e, db := setupTestServer(t)

	// Create balance:
	_, err := db.BalanceRepo.UpdateBalance(context.Background(), 1, "USDT", 100)
	require.NoError(t, err)

	reqBody := map[string]interface{}{
		"amount":          50,
		"currency":        "USDT",
		"destination":     "addr1",
		"idempotency_key": "ik1",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/withdrawals", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp domain.Withdrawal
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 50.0, resp.Amount)
	assert.Equal(t, "USDT", resp.Currency)
}

func TestCreateWithdrawalEndpoint_InsufficientBalance(t *testing.T) {
	e, db := setupTestServer(t)

	reqBody := map[string]interface{}{
		"amount":          1000,
		"currency":        "USDT",
		"destination":     "addr1",
		"idempotency_key": "ik1",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	// 1. No balance:
	req1 := httptest.NewRequest(http.MethodPost, "/withdrawals", bytes.NewReader(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()

	e.ServeHTTP(rec1, req1)

	assert.Equal(t, http.StatusConflict, rec1.Code)

	// 2. Insufficient balance:

	// Create balance:
	_, err := db.BalanceRepo.UpdateBalance(context.Background(), 1, "USDT", 100)
	require.NoError(t, err)

	reqBody["idempotency_key"] = "ik2"
	bodyBytes, _ = json.Marshal(reqBody)

	req2 := httptest.NewRequest(http.MethodPost, "/withdrawals", bytes.NewReader(bodyBytes))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()

	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusConflict, rec2.Code)
}

func TestCreateWithdrawalEndpoint_Idempotency(t *testing.T) {
	e, db := setupTestServer(t)

	// Create balance:
	_, err := db.BalanceRepo.UpdateBalance(context.Background(), 1, "USDT", 100)
	require.NoError(t, err)

	reqBody := map[string]interface{}{
		"amount":          30,
		"currency":        "USDT",
		"destination":     "addr1",
		"idempotency_key": "ik1",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// 1. First request:
	req1 := httptest.NewRequest(http.MethodPost, "/withdrawals", bytes.NewReader(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()

	e.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// 2. The same request:
	req2 := httptest.NewRequest(http.MethodPost, "/withdrawals", bytes.NewReader(bodyBytes))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()

	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)

	// 3. Same idempotency key, different payload:
	reqBody["amount"] = 40

	bodyBytesConflict, _ := json.Marshal(reqBody)

	req3 := httptest.NewRequest(http.MethodPost, "/withdrawals", bytes.NewReader(bodyBytesConflict))
	req3.Header.Set("Content-Type", "application/json")
	rec3 := httptest.NewRecorder()

	e.ServeHTTP(rec3, req3)
	assert.Equal(t, http.StatusUnprocessableEntity, rec3.Code)
}

func TestCreateWithdrawalEndpoint_Concurrent(t *testing.T) {
	e, db := setupTestServer(t)

	// Create balance
	_, err := db.BalanceRepo.UpdateBalance(context.Background(), 1, "USDT", 100)
	require.NoError(t, err)

	wg := sync.WaitGroup{}

	for i := 0; i < 2; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			reqBody := map[string]interface{}{
				"amount":          80,
				"currency":        "USDT",
				"destination":     "addr" + strconv.Itoa(i),
				"idempotency_key": "ik-conc-" + strconv.Itoa(i),
			}

			bodyBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/withdrawals", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

		}(i)
	}

	wg.Wait()

	var count int64
	err = db.DB.Model(&domain.Withdrawal{}).Count(&count).Error
	require.NoError(t, err)

	assert.Equal(t, int64(1), count)

	var balance float64
	err = db.DB.Raw(`
		SELECT balance
		FROM balances
		WHERE user_id = 1 AND currency = 'USDT'
	`).Scan(&balance).Error

	require.NoError(t, err)

	assert.Equal(t, 20.0, balance)
}
