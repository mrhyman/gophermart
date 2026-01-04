package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapAccrualStatusToOrderStatus(t *testing.T) {
	tests := []struct {
		name          string
		accrualStatus AccrualStatus
		want          OrderStatus
		wantErr       error
	}{
		{
			name:          "NEW maps to NEW",
			accrualStatus: AccrualStatusNew,
			want:          OrderStatusNew,
			wantErr:       nil,
		},
		{
			name:          "PROCESSING maps to NEW",
			accrualStatus: AccrualStatusProcessing,
			want:          OrderStatusNew,
			wantErr:       nil,
		},
		{
			name:          "INVALID maps to INVALID",
			accrualStatus: AccrualStatusInvalid,
			want:          OrderStatusInvalid,
			wantErr:       nil,
		},
		{
			name:          "PROCESSED maps to PROCESSED",
			accrualStatus: AccrualStatusProcessed,
			want:          OrderStatusProcessed,
			wantErr:       nil,
		},
		{
			name:          "unknown status returns error",
			accrualStatus: "UNKNOWN",
			want:          "",
			wantErr:       ErrUnknownAccrualStatus,
		},
		{
			name:          "empty status returns error",
			accrualStatus: "",
			want:          "",
			wantErr:       ErrUnknownAccrualStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapAccrualStatusToOrderStatus(tt.accrualStatus)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNewOrder(t *testing.T) {
	t.Run("create valid order", func(t *testing.T) {
		id := uuid.New()
		userID := uuid.New()
		number := "ORDER-123"
		status := OrderStatusNew
		createdAt := time.Now()

		order, err := NewOrder(id, userID, number, status, 0, createdAt)

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, id, order.ID)
		assert.Equal(t, userID, order.UserID)
		assert.Equal(t, number, order.Number)
		assert.Equal(t, status, order.Status)
		assert.Equal(t, createdAt, order.CreatedAt)
	})

	t.Run("create order with different statuses", func(t *testing.T) {
		statuses := []OrderStatus{
			OrderStatusNew,
			OrderStatusInvalid,
			OrderStatusProcessed,
		}

		for _, status := range statuses {
			t.Run(string(status), func(t *testing.T) {
				order, err := NewOrder(
					uuid.New(),
					uuid.New(),
					"ORDER-123",
					status,
					0,
					time.Now(),
				)

				require.NoError(t, err)
				assert.Equal(t, status, order.Status)
			})
		}
	})

	t.Run("create order with empty number", func(t *testing.T) {
		order, err := NewOrder(
			uuid.New(),
			uuid.New(),
			"",
			OrderStatusNew,
			0,
			time.Now(),
		)

		require.NoError(t, err)
		assert.Empty(t, order.Number)
	})

	t.Run("create order with zero time", func(t *testing.T) {
		order, err := NewOrder(
			uuid.New(),
			uuid.New(),
			"ORDER-123",
			OrderStatusNew,
			0,
			time.Time{},
		)

		require.NoError(t, err)
		assert.True(t, order.CreatedAt.IsZero())
	})

	t.Run("create order with nil UUID", func(t *testing.T) {
		order, err := NewOrder(
			uuid.Nil,
			uuid.New(),
			"ORDER-123",
			OrderStatusNew,
			0,
			time.Now(),
		)

		require.NoError(t, err)
		assert.Equal(t, uuid.Nil, order.ID)
	})
}

func TestOrderStatus_Constants(t *testing.T) {
	t.Run("order status values", func(t *testing.T) {
		assert.Equal(t, OrderStatus("NEW"), OrderStatusNew)
		assert.Equal(t, OrderStatus("INVALID"), OrderStatusInvalid)
		assert.Equal(t, OrderStatus("PROCESSED"), OrderStatusProcessed)
	})
}

func TestAccrualStatus_Constants(t *testing.T) {
	t.Run("accrual status values", func(t *testing.T) {
		assert.Equal(t, AccrualStatus("NEW"), AccrualStatusNew)
		assert.Equal(t, AccrualStatus("PROCESSING"), AccrualStatusProcessing)
		assert.Equal(t, AccrualStatus("INVALID"), AccrualStatusInvalid)
		assert.Equal(t, AccrualStatus("PROCESSED"), AccrualStatusProcessed)
	})
}
