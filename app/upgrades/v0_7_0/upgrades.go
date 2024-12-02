package v0_7_0 //nolint:revive // var-naming: don't use an underscore in package name

import (
	"context"

	cosmosmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/allora-network/allora-chain/app/keepers"
	"github.com/allora-network/allora-chain/app/params"
	"github.com/allora-network/allora-chain/app/upgrades"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

const (
	UpgradeName = "v0.7.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			feegrant.StoreKey, feemarkettypes.StoreKey,
		},
		Renamed: nil,
		Deleted: nil,
	},
}

func CreateUpgradeHandler(
	moduleManager *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		vm, err := moduleManager.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		sdkCtx.Logger().Info("CONFIGURING FEE MARKET MODULE")
		if err := ConfigureFeeMarketModule(sdkCtx, keepers); err != nil {
			return vm, err
		}
		sdkCtx.Logger().Info("FEE MARKET MODULE CONFIGURED")

		return vm, nil
	}
}

func ConfigureFeeMarketModule(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	ctx.Logger().Info("SETTING PARAMETERS FOR FEE MARKET MODULE")

	feeMarketParams, err := keepers.FeeMarketKeeper.GetParams(ctx)
	if err != nil {
		return err
	}

	feeMarketParams.Enabled = true
	feeMarketParams.FeeDenom = params.BaseCoinUnit
	feeMarketParams.DistributeFees = true
	feeMarketParams.MinBaseGasPrice = cosmosmath.LegacyMustNewDecFromStr("10")
	if err := keepers.FeeMarketKeeper.SetParams(ctx, feeMarketParams); err != nil {
		return err
	}
	ctx.Logger().Info("SETTING STATE FOR FEE MARKET MODULE")
	state, err := keepers.FeeMarketKeeper.GetState(ctx)
	if err != nil {
		return err
	}

	state.BaseGasPrice = cosmosmath.LegacyMustNewDecFromStr("10")

	return keepers.FeeMarketKeeper.SetState(ctx, state)
}
