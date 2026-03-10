package handler

import (
	"errors"
	"net/http"

	"github.com/DangeL187/erax"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"

	"back/internal/features/transaction/domain"
	"back/internal/features/transaction/usecase"
)

type TransactionHandler struct {
	uc *usecase.TransactionUseCase
}

func NewTransactionHandler(uc *usecase.TransactionUseCase) *TransactionHandler {
	return &TransactionHandler{uc: uc}
}

func (h *TransactionHandler) ConfirmWithdrawal(c *echo.Context) error {
	withdrawalID := c.Param("id")
	userID := c.Get("user_id").(int)

	err := h.uc.ConfirmWithdrawal(c.Request().Context(), userID, withdrawalID)
	if err != nil {
		zap.L().Error("\n" + erax.Format(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to confirm withdrawal"})
	}

	return c.NoContent(http.StatusOK)
}

func (h *TransactionHandler) CreateWithdrawal(c *echo.Context) error {
	userID := c.Get("user_id").(int)

	var req domain.CreateWithdrawalRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.Amount <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "amount must be > 0"})
	}

	w, err := h.uc.CreateWithdrawal(c.Request().Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInsufficientBalance):
			return c.JSON(http.StatusConflict, map[string]string{"error": "insufficient balance"})
		case errors.Is(err, domain.ErrIdempotencyConflict):
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "idempotency key conflict"})
		default:
			zap.L().Error("\n" + erax.Format(err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create withdrawal"})
		}
	}

	return c.JSON(http.StatusOK, w)
}

func (h *TransactionHandler) GetWithdrawal(c *echo.Context) error {
	withdrawalID := c.Param("id")
	userID := c.Get("user_id").(int)

	w, err := h.uc.GetWithdrawalByID(c.Request().Context(), userID, withdrawalID)
	if err != nil {
		if errors.Is(err, domain.ErrWithdrawalNotFound) || errors.Is(err, domain.ErrWithdrawalUserMismatch) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		zap.L().Error("\n" + erax.Format(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch withdrawal"})
	}

	return c.JSON(http.StatusOK, w)
}
