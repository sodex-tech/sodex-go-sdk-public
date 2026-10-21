package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDcaAccountPayloads verifies DCA-only snapshots/updates in both markets preserve all fields and lifecycle statuses.
func TestDcaAccountPayloads(t *testing.T) {
	for _, status := range []string{"ACTIVE", "CANCELLED", "TERMINATED", "PRUNED"} {
		t.Run(status, func(t *testing.T) {
			data := []byte(fmt.Sprintf(`{"DO":[{"s":"BTC-USD","i":9007199254740993,"a":"100","S":"BUY","sc":{"kind":"monthly","hour":0,"day":31},"lp":"1","hp":"90000","r":false,"im":true,"z":"0.001","v":"90","f":2,"ct":1700000000000,"nt":1700086400000,"st":%q}]}`, status))
			var spotState WsSpotAccountState
			var spotUpdate WsSpotAccountUpdate
			var perpsState WsPerpsAccountState
			var perpsUpdate WsPerpsAccountUpdate
			for _, payload := range []any{&spotState, &spotUpdate, &perpsState, &perpsUpdate} {
				require.NoError(t, json.Unmarshal(data, payload))
			}
			for _, plans := range [][]*DcaOrder{spotState.Dcas, spotUpdate.Dcas, perpsState.Dcas, perpsUpdate.Dcas} {
				require.Len(t, plans, 1)
				plan := plans[0]
				require.EqualValues(t, 9007199254740993, plan.OrderID)
				require.Equal(t, status, plan.Status)
				require.Equal(t, "BTC-USD", plan.Symbol)
				require.Equal(t, "100", plan.Amount)
				require.Equal(t, "BUY", plan.Side)
				require.EqualValues(t, 0, plan.Schedule.Hour)
				require.EqualValues(t, 31, *plan.Schedule.Day)
				require.Equal(t, "1", *plan.MinPrice)
				require.Equal(t, "90000", *plan.MaxPrice)
				require.False(t, plan.Randomize)
				require.True(t, plan.Immediate)
				require.Equal(t, "0.001", plan.ExecutedQty)
				require.Equal(t, "90", plan.ExecutedValue)
				require.EqualValues(t, 2, plan.Failures)
				require.EqualValues(t, 1700000000000, plan.CreatedAt)
				require.EqualValues(t, 1700086400000, plan.NextActiveAt)
			}
		})
	}
	var plan DcaOrder
	require.NoError(t, json.Unmarshal([]byte(`{"sc":{"kind":"daily","hour":0}}`), &plan))
	require.Nil(t, plan.MinPrice)
	require.Nil(t, plan.MaxPrice)
	require.Nil(t, plan.Schedule.Day)
}
