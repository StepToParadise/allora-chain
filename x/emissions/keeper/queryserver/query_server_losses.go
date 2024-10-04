package queryserver

import (
	"context"
	"time"

	"github.com/allora-network/allora-chain/x/emissions/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/allora-network/allora-chain/x/emissions/types"
)

func (qs queryServer) GetNetworkLossBundleAtBlock(ctx context.Context, req *types.GetNetworkLossBundleAtBlockRequest,
) (
	_ *types.GetNetworkLossBundleAtBlockResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetNetworkLossBundleAtBlock", time.Now(), returnErr == nil)
	topicExists, err := qs.k.TopicExists(ctx, req.TopicId)
	if !topicExists {
		return nil, status.Errorf(codes.NotFound, "topic %v not found", req.TopicId)
	} else if err != nil {
		return nil, err
	}

	networkLoss, err := qs.k.GetNetworkLossBundleAtBlock(ctx, req.TopicId, req.BlockHeight)
	if err != nil {
		return nil, err
	}

	return &types.GetNetworkLossBundleAtBlockResponse{LossBundle: networkLoss}, nil
}

func (qs queryServer) IsReputerNonceUnfulfilled(
	ctx context.Context,
	req *types.IsReputerNonceUnfulfilledRequest,
) (
	_ *types.IsReputerNonceUnfulfilledResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("IsReputerNonceUnfulfilled", time.Now(), returnErr == nil)
	isReputerNonceUnfulfilled, err :=
		qs.k.IsReputerNonceUnfulfilled(ctx, req.TopicId, &types.Nonce{BlockHeight: req.BlockHeight})

	return &types.IsReputerNonceUnfulfilledResponse{IsReputerNonceUnfulfilled: isReputerNonceUnfulfilled}, err
}

func (qs queryServer) GetUnfulfilledReputerNonces(
	ctx context.Context,
	req *types.GetUnfulfilledReputerNoncesRequest,
) (
	_ *types.GetUnfulfilledReputerNoncesResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetUnfulfilledReputerNonces", time.Now(), returnErr == nil)
	unfulfilledNonces, err := qs.k.GetUnfulfilledReputerNonces(ctx, req.TopicId)
	if err != nil {
		return nil, err
	}

	return &types.GetUnfulfilledReputerNoncesResponse{Nonces: &unfulfilledNonces}, nil
}

func (qs queryServer) GetReputerLossBundlesAtBlock(
	ctx context.Context,
	req *types.GetReputerLossBundlesAtBlockRequest,
) (
	_ *types.GetReputerLossBundlesAtBlockResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetReputerLossBundlesAtBlock", time.Now(), returnErr == nil)
	reputerLossBundles, err := qs.k.GetReputerLossBundlesAtBlock(ctx, req.TopicId, req.BlockHeight)
	if err != nil {
		return nil, err
	}

	return &types.GetReputerLossBundlesAtBlockResponse{LossBundles: reputerLossBundles}, nil
}
