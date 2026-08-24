package client

import (
	"context"
	"fmt"
	"strings"

	ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"
)

// ApproveBuilderFee approves a builder's maximum fee rate for both Spot and Perps.
func (c *Client) ApproveBuilderFee(ctx context.Context, userAddress string, request *ctypes.ApproveBuilderFeeRequest) error {
	if c.spotSgn == nil {
		return ErrNotAuthenticated
	}
	if !strings.EqualFold(userAddress, c.Address()) {
		return fmt.Errorf("client: builder approval must be signed by the master wallet")
	}
	nonce := c.nonce()
	sig, err := c.spotSgn.SignApproveBuilderFeeRequest(request, nonce, nil)
	if err != nil {
		return fmt.Errorf("client: sign approve builder fee: %w", err)
	}
	return c.postSigned(
		ctx,
		fmt.Sprintf("/api/v1/user/%s/builders", userAddress),
		request,
		sig,
		nonce,
		nil,
	)
}
