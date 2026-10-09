package types

type TwapOrder struct {
	Symbol        string `json:"s"`
	OrderID       uint64 `json:"i"`
	Side          string `json:"S"`
	Quantity      string `json:"q"`
	Minutes       uint64 `json:"m"`
	Randomize     bool   `json:"r"`
	ReduceOnly    bool   `json:"R"`
	ExecutedQty   string `json:"z"`
	ExecutedValue string `json:"v"`
	CreatedAt     uint64 `json:"ct"`
	NextActiveAt  uint64 `json:"nt"`
	Active        bool   `json:"a"`
}
