package inferencesynthesis

import (
	errorsmod "cosmossdk.io/errors"
	alloraMath "github.com/allora-network/allora-chain/math"
	emissions "github.com/allora-network/allora-chain/x/emissions/types"
)

type RunningWeightedLoss struct {
	UnnormalizedWeightedLoss Loss
	SumWeight                Weight
}

// Update the running information needed to calculate weighted loss per worker
func RunningWeightedAvgUpdate(
	runningWeightedAvg *RunningWeightedLoss,
	nextWeight Weight,
	nextValue Weight,
) (RunningWeightedLoss, error) {
	nextValTimesWeight, err := nextValue.Mul(nextWeight)
	if err != nil {
		return RunningWeightedLoss{}, err
	}
	newUnnormalizedWeightedLoss, err := runningWeightedAvg.UnnormalizedWeightedLoss.Add(nextValTimesWeight)
	if err != nil {
		return RunningWeightedLoss{}, err
	}
	newSumWeight, err := runningWeightedAvg.SumWeight.Add(nextWeight)
	if err != nil {
		return RunningWeightedLoss{}, err
	}
	return RunningWeightedLoss{
		UnnormalizedWeightedLoss: newUnnormalizedWeightedLoss,
		SumWeight:                newSumWeight,
	}, nil
}

// Convert the running weighted average objects to WorkerAttributedValues
func convertMapOfRunningWeightedLossesToWorkerAttributedValue[T emissions.WorkerAttributedValue | emissions.WithheldWorkerAttributedValue](
	runningWeightedLosses map[Worker]*RunningWeightedLoss,
	sortedWorkers []Worker,
) []*T {
	weightedLosses := make([]*T, 0)
	for _, worker := range sortedWorkers {
		runningLoss, ok := runningWeightedLosses[worker]
		if !ok {
			continue
		}
		normalizedWeightedLoss, err := normalizeWeightedLoss(runningLoss)
		if err != nil {
			continue
		}
		weightedLosses = append(weightedLosses, &T{
			Worker: worker,
			Value:  normalizedWeightedLoss,
		})
	}
	return weightedLosses
}

// Assumes stakes are all positive
func CalcNetworkLosses(
	topicId uint64,
	blockHeight int64,
	stakesByReputer map[Worker]Stake,
	reputerReportedLosses emissions.ReputerValueBundles,
) (emissions.ValueBundle, error) {
	// Make map from inferer to their running weighted-average loss
	runningWeightedCombinedLoss := RunningWeightedLoss{alloraMath.ZeroDec(), alloraMath.ZeroDec()}
	runningWeightedInfererLosses := make(map[Worker]*RunningWeightedLoss)
	runningWeightedForecasterLosses := make(map[Worker]*RunningWeightedLoss)
	runningWeightedNaiveLoss := RunningWeightedLoss{alloraMath.ZeroDec(), alloraMath.ZeroDec()}
	runningWeightedOneOutInfererForecasterLosses := make(map[Worker]map[Worker]*RunningWeightedLoss)
	runningWeightedOneOutInfererLosses := make(map[Worker]*RunningWeightedLoss) // Withheld worker -> Forecaster -> Loss
	runningWeightedOneOutForecasterLosses := make(map[Worker]*RunningWeightedLoss)
	runningWeightedOneInForecasterLosses := make(map[Worker]*RunningWeightedLoss)

	for _, report := range reputerReportedLosses.ReputerValueBundles {
		if report.ValueBundle != nil {
			if report.ValueBundle.TopicId != topicId {
				return emissions.ValueBundle{}, errorsmod.Wrapf(
					emissions.ErrInvalidTopicId,
					"all reputer bundles must be for same topic to calculate network losses %d!=%d",
					topicId,
					report.ValueBundle.TopicId,
				)
			}
			if report.ValueBundle.ReputerRequestNonce.ReputerNonce.BlockHeight != blockHeight {
				return emissions.ValueBundle{}, errorsmod.Wrapf(
					emissions.ErrInvalidValue,
					"all reputer bundles must be for same nonce to calculate network losses %d!=%d",
					blockHeight,
					report.ValueBundle.ReputerRequestNonce.ReputerNonce.BlockHeight,
				)
			}
			stakeAmount, err := alloraMath.NewDecFromSdkInt(stakesByReputer[report.ValueBundle.Reputer])
			if err != nil {
				return emissions.ValueBundle{}, err
			}

			// Update combined loss with reputer reported loss and stake
			runningWeightedCombinedLoss, err = RunningWeightedAvgUpdate(&runningWeightedCombinedLoss, stakeAmount, report.ValueBundle.CombinedValue)
			if err != nil {
				return emissions.ValueBundle{}, errorsmod.Wrapf(err, "Error updating running weighted average for next combined loss")
			}

			// Not all reputers may have reported losses on the same set of inferers => important that the code below doesn't assume that!
			// Update inferer losses
			for _, loss := range report.ValueBundle.InfererValues {
				if runningWeightedInfererLosses[loss.Worker] == nil {
					runningWeightedInfererLosses[loss.Worker] = &RunningWeightedLoss{
						UnnormalizedWeightedLoss: alloraMath.ZeroDec(),
						SumWeight:                alloraMath.ZeroDec(),
					}
				}

				nextAvg, err := RunningWeightedAvgUpdate(runningWeightedInfererLosses[loss.Worker], stakeAmount, loss.Value)
				if err != nil {
					return emissions.ValueBundle{}, errorsmod.Wrapf(err, "Error updating running weighted average for inferer")
				}
				runningWeightedInfererLosses[loss.Worker] = &nextAvg
			}

			// Update forecaster losses
			for _, loss := range report.ValueBundle.ForecasterValues {
				if runningWeightedForecasterLosses[loss.Worker] == nil {
					runningWeightedForecasterLosses[loss.Worker] = &RunningWeightedLoss{
						UnnormalizedWeightedLoss: alloraMath.ZeroDec(),
						SumWeight:                alloraMath.ZeroDec(),
					}
				}

				nextAvg, err := RunningWeightedAvgUpdate(runningWeightedForecasterLosses[loss.Worker], stakeAmount, loss.Value)
				if err != nil {
					return emissions.ValueBundle{}, errorsmod.Wrapf(err, "Error updating running weighted average for forecaster")
				}
				runningWeightedForecasterLosses[loss.Worker] = &nextAvg
			}

			// Update naive loss
			runningWeightedNaiveLoss, err = RunningWeightedAvgUpdate(&runningWeightedNaiveLoss, stakeAmount, report.ValueBundle.NaiveValue)
			if err != nil {
				return emissions.ValueBundle{}, errorsmod.Wrapf(err, "Error updating running weighted average for naive loss: ")
			}

			// Update one-out inferer forecaster losses
			for _, losses := range report.ValueBundle.OneOutInfererForecasterValues {
				for _, loss := range losses.OneOutInfererValues {
					if runningWeightedOneOutInfererForecasterLosses[losses.Forecaster][loss.Worker] == nil {
						// initialize first map
						if runningWeightedOneOutInfererForecasterLosses[losses.Forecaster] == nil {
							runningWeightedOneOutInfererForecasterLosses[losses.Forecaster] = make(map[Worker]*RunningWeightedLoss)
						}
						runningWeightedOneOutInfererForecasterLosses[losses.Forecaster][loss.Worker] = &RunningWeightedLoss{
							UnnormalizedWeightedLoss: alloraMath.ZeroDec(),
							SumWeight:                alloraMath.ZeroDec(),
						}
					}

					nextAvg, err := RunningWeightedAvgUpdate(runningWeightedOneOutInfererForecasterLosses[losses.Forecaster][loss.Worker], stakeAmount, loss.Value)
					if err != nil {
						return emissions.ValueBundle{}, errorsmod.Wrapf(err, "Error updating running weighted average for one-out inferer forecaster")
					}
					runningWeightedOneOutInfererForecasterLosses[losses.Forecaster][loss.Worker] = &nextAvg
				}
			}

			// Update one-out inferer losses
			for _, loss := range report.ValueBundle.OneOutInfererValues {
				if runningWeightedOneOutInfererLosses[loss.Worker] == nil {
					runningWeightedOneOutInfererLosses[loss.Worker] = &RunningWeightedLoss{
						UnnormalizedWeightedLoss: alloraMath.ZeroDec(),
						SumWeight:                alloraMath.ZeroDec(),
					}
				}
				nextAvg, err := RunningWeightedAvgUpdate(runningWeightedOneOutInfererLosses[loss.Worker], stakeAmount, loss.Value)
				if err != nil {
					return emissions.ValueBundle{}, errorsmod.Wrapf(err, "Error updating running weighted average for one-out inferer")
				}
				runningWeightedOneOutInfererLosses[loss.Worker] = &nextAvg
			}

			// Update one-out forecaster losses
			for _, loss := range report.ValueBundle.OneOutForecasterValues {
				if runningWeightedOneOutForecasterLosses[loss.Worker] == nil {
					runningWeightedOneOutForecasterLosses[loss.Worker] = &RunningWeightedLoss{
						UnnormalizedWeightedLoss: alloraMath.ZeroDec(),
						SumWeight:                alloraMath.ZeroDec(),
					}
				}

				nextAvg, err := RunningWeightedAvgUpdate(runningWeightedOneOutForecasterLosses[loss.Worker], stakeAmount, loss.Value)
				if err != nil {
					return emissions.ValueBundle{}, errorsmod.Wrapf(err, "Error updating running weighted average for one-out forecaster")
				}
				runningWeightedOneOutForecasterLosses[loss.Worker] = &nextAvg
			}

			// Update one-in forecaster losses
			for _, loss := range report.ValueBundle.OneInForecasterValues {
				if runningWeightedOneInForecasterLosses[loss.Worker] == nil {
					runningWeightedOneInForecasterLosses[loss.Worker] = &RunningWeightedLoss{
						UnnormalizedWeightedLoss: alloraMath.ZeroDec(),
						SumWeight:                alloraMath.ZeroDec(),
					}
				}

				nextAvg, err := RunningWeightedAvgUpdate(runningWeightedOneInForecasterLosses[loss.Worker], stakeAmount, loss.Value)
				if err != nil {
					return emissions.ValueBundle{}, errorsmod.Wrapf(err, "Error updating running weighted average for one-in forecaster")
				}
				runningWeightedOneInForecasterLosses[loss.Worker] = &nextAvg
			}
		}
	}

	sortedInferers := alloraMath.GetSortedKeys(runningWeightedInfererLosses)
	sortedForecasters := alloraMath.GetSortedKeys(runningWeightedForecasterLosses)

	// Normalize the combined loss
	combinedValue, err := normalizeWeightedLoss(&runningWeightedCombinedLoss)
	if err != nil {
		return emissions.ValueBundle{}, errorsmod.Wrapf(err, "Error normalizing combined loss")
	}

	// Normalize the naive loss
	naiveValue, err := normalizeWeightedLoss(&runningWeightedNaiveLoss)
	if err != nil {
		return emissions.ValueBundle{}, errorsmod.Wrapf(err, "Error normalizing naive loss")
	}

	// Convert the running weighted averages to WorkerAttributedValue/WithheldWorkerAttributedValue for inferers and forecasters
	infererLosses := convertMapOfRunningWeightedLossesToWorkerAttributedValue[emissions.WorkerAttributedValue](runningWeightedInfererLosses, sortedInferers)
	forecasterLosses := convertMapOfRunningWeightedLossesToWorkerAttributedValue[emissions.WorkerAttributedValue](runningWeightedForecasterLosses, sortedForecasters)
	oneOutInfererLosses := convertMapOfRunningWeightedLossesToWorkerAttributedValue[emissions.WithheldWorkerAttributedValue](runningWeightedOneOutInfererLosses, sortedInferers)
	oneOutForecasterLosses := convertMapOfRunningWeightedLossesToWorkerAttributedValue[emissions.WithheldWorkerAttributedValue](runningWeightedOneOutForecasterLosses, sortedForecasters)
	oneInForecasterLosses := convertMapOfRunningWeightedLossesToWorkerAttributedValue[emissions.WorkerAttributedValue](runningWeightedOneInForecasterLosses, sortedForecasters)
	oneOutInfererForecasterLosses := make([]*emissions.OneOutInfererForecasterValues, 0)
	for _, forecaster := range sortedForecasters {
		innerMap, ok := runningWeightedOneOutInfererForecasterLosses[forecaster]
		if !ok {
			continue
		}
		oneOutInfererValues := convertMapOfRunningWeightedLossesToWorkerAttributedValue[emissions.WithheldWorkerAttributedValue](innerMap, sortedInferers)
		oneOutInfererForecasterLosses = append(oneOutInfererForecasterLosses, &emissions.OneOutInfererForecasterValues{
			Forecaster:          forecaster,
			OneOutInfererValues: oneOutInfererValues,
		})
	}

	output := emissions.ValueBundle{
		TopicId:             topicId,
		ReputerRequestNonce: &emissions.ReputerRequestNonce{ReputerNonce: &emissions.Nonce{BlockHeight: blockHeight}},
		// we re-use the value bundle for the network loss bundle
		// this is bad practice and we should refactor: PROTO-2369, PROTO-2464
		// for now set the Reputer address to a "zero" address
		Reputer:                       "allo1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqas6usy",
		ExtraData:                     nil,
		CombinedValue:                 combinedValue,
		InfererValues:                 infererLosses,
		ForecasterValues:              forecasterLosses,
		NaiveValue:                    naiveValue,
		OneOutInfererValues:           oneOutInfererLosses,
		OneOutForecasterValues:        oneOutForecasterLosses,
		OneInForecasterValues:         oneInForecasterLosses,
		OneOutInfererForecasterValues: oneOutInfererForecasterLosses,
	}

	return output, nil
}

func normalizeWeightedLoss(
	runningWeightedLossData *RunningWeightedLoss,
) (alloraMath.Dec, error) {
	if runningWeightedLossData.SumWeight.IsZero() {
		return alloraMath.Dec{}, errorsmod.Wrapf(emissions.ErrFractionDivideByZero, "Sum weight for combined naive loss is 0")
	}

	normalizedWeightedLoss, err := runningWeightedLossData.UnnormalizedWeightedLoss.Quo(runningWeightedLossData.SumWeight)
	if err != nil {
		return alloraMath.Dec{}, err
	}

	return normalizedWeightedLoss, nil
}
