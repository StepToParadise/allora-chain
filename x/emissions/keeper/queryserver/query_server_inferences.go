package queryserver

import (
	"context"
	"time"

	"github.com/allora-network/allora-chain/x/emissions/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	alloraMath "github.com/allora-network/allora-chain/math"
	synth "github.com/allora-network/allora-chain/x/emissions/keeper/inference_synthesis"
	emissionstypes "github.com/allora-network/allora-chain/x/emissions/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GetWorkerLatestInferenceByTopicId handles the query for the latest inference by a specific worker for a given topic.
func (qs queryServer) GetWorkerLatestInferenceByTopicId(ctx context.Context, req *emissionstypes.GetWorkerLatestInferenceByTopicIdRequest) (
	_ *emissionstypes.GetWorkerLatestInferenceByTopicIdResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetWorkerLatestInferenceByTopicId", "rpc", time.Now(), returnErr == nil)
	if err := qs.k.ValidateStringIsBech32(req.WorkerAddress); err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}
	topicExists, err := qs.k.TopicExists(ctx, req.TopicId)
	if !topicExists {
		return nil, status.Errorf(codes.NotFound, "topic %v not found", req.TopicId)
	} else if err != nil {
		return nil, err
	}

	inference, err := qs.k.GetWorkerLatestInferenceByTopicId(ctx, req.TopicId, req.WorkerAddress)
	if err != nil {
		return nil, err
	}

	return &emissionstypes.GetWorkerLatestInferenceByTopicIdResponse{LatestInference: &inference}, nil
}

func (qs queryServer) GetInferencesAtBlock(ctx context.Context, req *emissionstypes.GetInferencesAtBlockRequest) (
	_ *emissionstypes.GetInferencesAtBlockResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetInferencesAtBlock", "rpc", time.Now(), returnErr == nil)
	topicExists, err := qs.k.TopicExists(ctx, req.TopicId)
	if !topicExists {
		return nil, status.Errorf(codes.NotFound, "topic %v not found", req.TopicId)
	} else if err != nil {
		return nil, err
	}

	inferences, err := qs.k.GetInferencesAtBlock(ctx, req.TopicId, req.BlockHeight)
	if err != nil {
		return nil, err
	}

	return &emissionstypes.GetInferencesAtBlockResponse{Inferences: inferences}, nil
}

// Return full set of inferences in I_i from the chain
func (qs queryServer) GetNetworkInferencesAtBlock(
	ctx context.Context,
	req *emissionstypes.GetNetworkInferencesAtBlockRequest,
) (
	_ *emissionstypes.GetNetworkInferencesAtBlockResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetNetworkInferencesAtBlock", "rpc", time.Now(), returnErr == nil)
	topic, err := qs.k.GetTopic(ctx, req.TopicId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "topic %v not found", req.TopicId)
	}
	if topic.EpochLastEnded == 0 {
		return nil, status.Errorf(codes.NotFound, "network inference not available for topic %v", req.TopicId)
	}

	networkInferences, _, _, _, _, _, err := synth.GetNetworkInferences(
		sdk.UnwrapSDKContext(ctx),
		qs.k,
		req.TopicId,
		&req.BlockHeightLastInference,
	)
	if err != nil {
		return nil, err
	}

	return &emissionstypes.GetNetworkInferencesAtBlockResponse{NetworkInferences: networkInferences}, nil
}

// Return full set of inferences in I_i from the chain, as well as weights and forecast implied inferences
func (qs queryServer) GetLatestNetworkInferences(
	ctx context.Context,
	req *emissionstypes.GetLatestNetworkInferencesRequest,
) (
	_ *emissionstypes.GetLatestNetworkInferencesResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetLatestNetworkInferences", "rpc", time.Now(), returnErr == nil)
	topicExists, err := qs.k.TopicExists(ctx, req.TopicId)
	if !topicExists {
		return nil, status.Errorf(codes.NotFound, "topic %v not found", req.TopicId)
	} else if err != nil {
		return nil, err
	}

	networkInferences, forecastImpliedInferenceByWorker, infererWeights, forecasterWeights, inferenceBlockHeight, lossBlockHeight, err := synth.GetNetworkInferences(
		sdk.UnwrapSDKContext(ctx),
		qs.k,
		req.TopicId,
		nil,
	)
	if err != nil {
		return nil, err
	}

	ciRawPercentiles, ciValues, err :=
		qs.GetConfidenceIntervalsForInferenceData(
			networkInferences,
			forecastImpliedInferenceByWorker,
			infererWeights,
			forecasterWeights,
		)
	if err != nil {
		return nil, err
	}

	if ciRawPercentiles == nil {
		ciRawPercentiles = []alloraMath.Dec{}
	}

	if ciValues == nil {
		ciValues = []alloraMath.Dec{}
	}

	inferers := alloraMath.GetSortedKeys(infererWeights)
	forecasters := alloraMath.GetSortedKeys(forecasterWeights)

	return &emissionstypes.GetLatestNetworkInferencesResponse{
		NetworkInferences:                networkInferences,
		InfererWeights:                   synth.ConvertWeightsToArrays(inferers, infererWeights),
		ForecasterWeights:                synth.ConvertWeightsToArrays(forecasters, forecasterWeights),
		ForecastImpliedInferences:        synth.ConvertForecastImpliedInferencesToArrays(forecasters, forecastImpliedInferenceByWorker),
		InferenceBlockHeight:             inferenceBlockHeight,
		LossBlockHeight:                  lossBlockHeight,
		ConfidenceIntervalRawPercentiles: ciRawPercentiles,
		ConfidenceIntervalValues:         ciValues,
	}, nil
}

func (qs queryServer) GetLatestAvailableNetworkInferences(
	ctx context.Context,
	req *emissionstypes.GetLatestAvailableNetworkInferencesRequest,
) (
	_ *emissionstypes.GetLatestAvailableNetworkInferencesResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetLatestAvailableNetworkInferences", "rpc", time.Now(), returnErr == nil)
	lastWorkerCommit, err := qs.k.GetWorkerTopicLastCommit(ctx, req.TopicId)
	if err != nil {
		return nil, err
	}

	lastReputerCommit, err := qs.k.GetReputerTopicLastCommit(ctx, req.TopicId)
	if err != nil {
		return nil, err
	}

	networkInferences, forecastImpliedInferenceByWorker, infererWeights, forecasterWeights, _, _, err :=
		synth.GetNetworkInferences(
			sdk.UnwrapSDKContext(ctx),
			qs.k,
			req.TopicId,
			&lastWorkerCommit.Nonce.BlockHeight,
		)
	if err != nil {
		return nil, err
	}

	ciRawPercentiles, ciValues, err :=
		qs.GetConfidenceIntervalsForInferenceData(
			networkInferences,
			forecastImpliedInferenceByWorker,
			infererWeights,
			forecasterWeights,
		)
	if err != nil {
		return nil, err
	}

	if ciRawPercentiles == nil {
		ciRawPercentiles = []alloraMath.Dec{}
	}

	if ciValues == nil {
		ciValues = []alloraMath.Dec{}
	}

	inferers := alloraMath.GetSortedKeys(infererWeights)
	forecasters := alloraMath.GetSortedKeys(forecasterWeights)

	return &emissionstypes.GetLatestAvailableNetworkInferencesResponse{
		NetworkInferences:                networkInferences,
		InfererWeights:                   synth.ConvertWeightsToArrays(inferers, infererWeights),
		ForecasterWeights:                synth.ConvertWeightsToArrays(forecasters, forecasterWeights),
		ForecastImpliedInferences:        synth.ConvertForecastImpliedInferencesToArrays(forecasters, forecastImpliedInferenceByWorker),
		InferenceBlockHeight:             lastWorkerCommit.Nonce.BlockHeight,
		LossBlockHeight:                  lastReputerCommit.Nonce.BlockHeight,
		ConfidenceIntervalRawPercentiles: ciRawPercentiles,
		ConfidenceIntervalValues:         ciValues,
	}, nil
}

func (qs queryServer) GetConfidenceIntervalsForInferenceData(
	networkInferences *emissionstypes.ValueBundle,
	forecastImpliedInferenceByWorker map[string]*emissionstypes.Inference,
	infererWeights map[string]alloraMath.Dec,
	forecasterWeights map[string]alloraMath.Dec,
) (_ []alloraMath.Dec, _ []alloraMath.Dec, returnErr error) {
	defer metrics.RecordMetrics("GetConfidenceIntervalsForInferenceData", "rpc", time.Now(), returnErr == nil)
	var inferences []alloraMath.Dec // from inferers + forecast-implied inferences
	var weights []alloraMath.Dec    // weights of all workers

	for _, inference := range networkInferences.InfererValues {
		weight, exists := infererWeights[inference.Worker]
		if exists {
			inferences = append(inferences, inference.Value)
			weights = append(weights, weight)
		}
	}

	for _, forecast := range networkInferences.ForecasterValues {
		weight, exists := forecasterWeights[forecast.Worker]
		if exists {
			inferences = append(inferences, forecastImpliedInferenceByWorker[forecast.Worker].Value)
			weights = append(weights, weight)
		}
	}

	ciRawPercentiles := []alloraMath.Dec{
		alloraMath.MustNewDecFromString("2.28"),
		alloraMath.MustNewDecFromString("15.87"),
		alloraMath.MustNewDecFromString("50"),
		alloraMath.MustNewDecFromString("84.13"),
		alloraMath.MustNewDecFromString("97.72"),
	}

	var ciValues []alloraMath.Dec
	var err error
	if len(inferences) == 0 {
		ciRawPercentiles = []alloraMath.Dec{}
		ciValues = []alloraMath.Dec{}
	} else {
		ciValues, err = alloraMath.WeightedPercentile(inferences, weights, ciRawPercentiles)
		if err != nil {
			return nil, nil, err
		}
	}

	return ciRawPercentiles, ciValues, nil
}

func (qs queryServer) GetLatestTopicInferences(
	ctx context.Context,
	req *emissionstypes.GetLatestTopicInferencesRequest,
) (
	_ *emissionstypes.GetLatestTopicInferencesResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetLatestTopicInferences", "rpc", time.Now(), returnErr == nil)
	topicExists, err := qs.k.TopicExists(ctx, req.TopicId)
	if !topicExists {
		return nil, status.Errorf(codes.NotFound, "topic %v not found", req.TopicId)
	} else if err != nil {
		return nil, err
	}

	inferences, blockHeight, err := qs.k.GetLatestTopicInferences(ctx, req.TopicId)
	if err != nil {
		return nil, err
	}

	return &emissionstypes.GetLatestTopicInferencesResponse{Inferences: inferences, BlockHeight: blockHeight}, nil
}

func (qs queryServer) IsWorkerNonceUnfulfilled(
	ctx context.Context,
	req *emissionstypes.IsWorkerNonceUnfulfilledRequest,
) (
	_ *emissionstypes.IsWorkerNonceUnfulfilledResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("IsWorkerNonceUnfulfilled", "rpc", time.Now(), returnErr == nil)
	isWorkerNonceUnfulfilled, err :=
		qs.k.IsWorkerNonceUnfulfilled(ctx, req.TopicId, &emissionstypes.Nonce{BlockHeight: req.BlockHeight})

	return &emissionstypes.IsWorkerNonceUnfulfilledResponse{IsWorkerNonceUnfulfilled: isWorkerNonceUnfulfilled}, err
}

func (qs queryServer) GetUnfulfilledWorkerNonces(
	ctx context.Context,
	req *emissionstypes.GetUnfulfilledWorkerNoncesRequest,
) (
	_ *emissionstypes.GetUnfulfilledWorkerNoncesResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetUnfulfilledWorkerNonces", "rpc", time.Now(), returnErr == nil)
	unfulfilledNonces, err := qs.k.GetUnfulfilledWorkerNonces(ctx, req.TopicId)
	if err != nil {
		return nil, err
	}

	return &emissionstypes.GetUnfulfilledWorkerNoncesResponse{Nonces: &unfulfilledNonces}, nil
}

func (qs queryServer) GetInfererNetworkRegret(
	ctx context.Context,
	req *emissionstypes.GetInfererNetworkRegretRequest,
) (
	_ *emissionstypes.GetInfererNetworkRegretResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetInfererNetworkRegret", "rpc", time.Now(), returnErr == nil)
	infererNetworkRegret, _, err := qs.k.GetInfererNetworkRegret(ctx, req.TopicId, req.ActorId)
	if err != nil {
		return nil, err
	}

	return &emissionstypes.GetInfererNetworkRegretResponse{Regret: &infererNetworkRegret}, nil
}

func (qs queryServer) GetForecasterNetworkRegret(
	ctx context.Context,
	req *emissionstypes.GetForecasterNetworkRegretRequest,
) (
	_ *emissionstypes.GetForecasterNetworkRegretResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetForecasterNetworkRegret", "rpc", time.Now(), returnErr == nil)
	forecasterNetworkRegret, _, err := qs.k.GetForecasterNetworkRegret(ctx, req.TopicId, req.Worker)
	if err != nil {
		return nil, err
	}

	return &emissionstypes.GetForecasterNetworkRegretResponse{Regret: &forecasterNetworkRegret}, nil
}

func (qs queryServer) GetOneInForecasterNetworkRegret(
	ctx context.Context,
	req *emissionstypes.GetOneInForecasterNetworkRegretRequest,
) (
	_ *emissionstypes.GetOneInForecasterNetworkRegretResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetOneInForecasterNetworkRegret", "rpc", time.Now(), returnErr == nil)
	oneInForecasterNetworkRegret, _, err := qs.k.GetOneInForecasterNetworkRegret(ctx, req.TopicId, req.Forecaster, req.Inferer)
	if err != nil {
		return nil, err
	}

	return &emissionstypes.GetOneInForecasterNetworkRegretResponse{Regret: &oneInForecasterNetworkRegret}, nil
}

func (qs queryServer) GetNaiveInfererNetworkRegret(
	ctx context.Context,
	req *emissionstypes.GetNaiveInfererNetworkRegretRequest,
) (
	_ *emissionstypes.GetNaiveInfererNetworkRegretResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetNaiveInfererNetworkRegret", "rpc", time.Now(), returnErr == nil)
	regret, _, err := qs.k.GetNaiveInfererNetworkRegret(ctx, req.TopicId, req.Inferer)
	if err != nil {
		return nil, err
	}

	return &emissionstypes.GetNaiveInfererNetworkRegretResponse{Regret: &regret}, nil
}

func (qs queryServer) GetOneOutInfererInfererNetworkRegret(
	ctx context.Context,
	req *emissionstypes.GetOneOutInfererInfererNetworkRegretRequest,
) (
	_ *emissionstypes.GetOneOutInfererInfererNetworkRegretResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetOneOutInfererInfererNetworkRegret", "rpc", time.Now(), returnErr == nil)
	regret, _, err := qs.k.GetOneOutInfererInfererNetworkRegret(ctx, req.TopicId, req.OneOutInferer, req.Inferer)
	if err != nil {
		return nil, err
	}

	return &emissionstypes.GetOneOutInfererInfererNetworkRegretResponse{Regret: &regret}, nil
}

func (qs queryServer) GetOneOutInfererForecasterNetworkRegret(
	ctx context.Context,
	req *emissionstypes.GetOneOutInfererForecasterNetworkRegretRequest,
) (
	_ *emissionstypes.GetOneOutInfererForecasterNetworkRegretResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetOneOutInfererForecasterNetworkRegret", "rpc", time.Now(), returnErr == nil)
	regret, _, err := qs.k.GetOneOutInfererForecasterNetworkRegret(ctx, req.TopicId, req.OneOutInferer, req.Forecaster)
	if err != nil {
		return nil, err
	}

	return &emissionstypes.GetOneOutInfererForecasterNetworkRegretResponse{Regret: &regret}, nil
}

func (qs queryServer) GetOneOutForecasterInfererNetworkRegret(
	ctx context.Context,
	req *emissionstypes.GetOneOutForecasterInfererNetworkRegretRequest,
) (
	_ *emissionstypes.GetOneOutForecasterInfererNetworkRegretResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetOneOutForecasterInfererNetworkRegret", "rpc", time.Now(), returnErr == nil)
	regret, _, err := qs.k.GetOneOutForecasterInfererNetworkRegret(ctx, req.TopicId, req.OneOutForecaster, req.Inferer)
	if err != nil {
		return nil, err
	}

	return &emissionstypes.GetOneOutForecasterInfererNetworkRegretResponse{Regret: &regret}, nil
}

func (qs queryServer) GetOneOutForecasterForecasterNetworkRegret(
	ctx context.Context,
	req *emissionstypes.GetOneOutForecasterForecasterNetworkRegretRequest,
) (
	_ *emissionstypes.GetOneOutForecasterForecasterNetworkRegretResponse,
	returnErr error,
) {
	defer metrics.RecordMetrics("GetOneOutForecasterForecasterNetworkRegret", "rpc", time.Now(), returnErr == nil)
	regret, _, err := qs.k.GetOneOutForecasterForecasterNetworkRegret(ctx, req.TopicId, req.OneOutForecaster, req.Forecaster)
	if err != nil {
		return nil, err
	}

	return &emissionstypes.GetOneOutForecasterForecasterNetworkRegretResponse{Regret: &regret}, nil
}
