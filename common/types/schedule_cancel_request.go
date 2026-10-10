package types

const ScheduleCancelRequestTypeName = "scheduleCancel"

type ScheduleCancelRequest struct {
	AccountID          uint64  `json:"accountID"`
	ScheduledTimestamp *uint64 `json:"scheduledTimestamp,omitempty"`
}

func (req *ScheduleCancelRequest) ActionName() string {
	return ScheduleCancelRequestTypeName
}
