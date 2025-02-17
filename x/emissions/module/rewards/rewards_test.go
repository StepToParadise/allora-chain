package rewards_test

import (
	"fmt"
	l "log"
	"testing"
	"time"

	actorutils "github.com/allora-network/allora-chain/x/emissions/keeper/actor_utils"
	inferencesynthesis "github.com/allora-network/allora-chain/x/emissions/keeper/inference_synthesis"

	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	cosmosMath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/allora-network/allora-chain/app/params"
	alloraMath "github.com/allora-network/allora-chain/math"
	alloratestutil "github.com/allora-network/allora-chain/test/testutil"
	"github.com/allora-network/allora-chain/x/emissions/keeper"
	"github.com/allora-network/allora-chain/x/emissions/keeper/msgserver"
	"github.com/allora-network/allora-chain/x/emissions/module"
	"github.com/allora-network/allora-chain/x/emissions/module/rewards"
	"github.com/allora-network/allora-chain/x/emissions/types"
	mintkeeper "github.com/allora-network/allora-chain/x/mint/keeper"
	mint "github.com/allora-network/allora-chain/x/mint/module"
	minttypes "github.com/allora-network/allora-chain/x/mint/types"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	codecAddress "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/stretchr/testify/suite"
)

const (
	multiPerm  = "multiple permissions account"
	randomPerm = "random permission"
)

type RewardsTestSuite struct {
	suite.Suite

	ctx                sdk.Context
	accountKeeper      authkeeper.AccountKeeper
	bankKeeper         bankkeeper.BaseKeeper
	emissionsKeeper    keeper.Keeper
	emissionsAppModule module.AppModule
	mintAppModule      mint.AppModule
	msgServer          types.MsgServiceServer
	key                *storetypes.KVStoreKey
	privKeys           []secp256k1.PrivKey
	addrs              []sdk.AccAddress
	addrsStr           []string
	pubKeyHexStr       []string
}

func (s *RewardsTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey("emissions")
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	// Set logger to show logs from the rewards module too
	logger := log.NewTestLogger(s.T()).With("module", "rewards")
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now()}).WithLogger(logger) // nolint: exhaustruct
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, module.AppModule{})

	maccPerms := map[string][]string{
		"fee_collector":                {"minter"},
		"mint":                         {"minter"},
		types.AlloraStakingAccountName: {"burner", "minter", "staking"},
		types.AlloraRewardsAccountName: {"minter"},
		types.AlloraPendingRewardForDelegatorAccountName: {"minter"},
		"ecosystem":              {"minter"},
		"bonded_tokens_pool":     {"burner", "staking"},
		"not_bonded_tokens_pool": {"burner", "staking"},
		multiPerm:                {"burner", "minter", "staking"},
		randomPerm:               {"random"},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		encCfg.Codec,
		storeService,
		authtypes.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec(params.Bech32PrefixAccAddr),
		params.Bech32PrefixAccAddr,
		authtypes.NewModuleAddress("gov").String(),
	)
	bankKeeper := bankkeeper.NewBaseKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		map[string]bool{},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		log.NewNopLogger(),
	)
	emissionsKeeper := keeper.NewKeeper(
		encCfg.Codec,
		codecAddress.NewBech32Codec(params.Bech32PrefixAccAddr),
		storeService,
		accountKeeper,
		bankKeeper,
		authtypes.FeeCollectorName)
	stakingKeeper := stakingkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress("gov").String(),
		codecAddress.NewBech32Codec(sdk.Bech32PrefixValAddr),
		codecAddress.NewBech32Codec(sdk.Bech32PrefixConsAddr),
	)
	mintKeeper := mintkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		stakingKeeper,
		accountKeeper,
		bankKeeper,
		emissionsKeeper,
		authtypes.FeeCollectorName,
	)

	s.ctx = ctx
	s.accountKeeper = accountKeeper
	s.bankKeeper = bankKeeper
	s.emissionsKeeper = emissionsKeeper
	s.key = key
	emissionsAppModule := module.NewAppModule(encCfg.Codec, s.emissionsKeeper)
	defaultEmissionsGenesis := emissionsAppModule.DefaultGenesis(encCfg.Codec)
	emissionsAppModule.InitGenesis(ctx, encCfg.Codec, defaultEmissionsGenesis)
	s.msgServer = msgserver.NewMsgServerImpl(s.emissionsKeeper)
	s.emissionsAppModule = emissionsAppModule
	mintAppModule := mint.NewAppModule(encCfg.Codec, mintKeeper, accountKeeper)
	defaultMintGenesis := mintAppModule.DefaultGenesis(encCfg.Codec)
	mintAppModule.InitGenesis(ctx, encCfg.Codec, defaultMintGenesis)
	s.mintAppModule = mintAppModule

	// Fund the rewards account generously
	s.FundAccount(10000000000, s.accountKeeper.GetModuleAddress(types.AlloraRewardsAccountName))

	s.privKeys, s.pubKeyHexStr, s.addrs, s.addrsStr = alloratestutil.GenerateTestAccounts(50)
	for _, addr := range s.addrs {
		s.FundAccount(10000000000, addr)
	}

	// Add all tests addresses in whitelists
	for _, addr := range s.addrsStr {
		err := s.emissionsKeeper.AddWhitelistAdmin(ctx, addr)
		s.Require().NoError(err)

		err = s.emissionsKeeper.AddToTopicCreatorWhitelist(ctx, addr)
		s.Require().NoError(err)

		err = s.emissionsKeeper.AddToGlobalWhitelist(ctx, addr)
		s.Require().NoError(err)
	}
	// topic id must not start at 0
	id, err := s.emissionsKeeper.IncrementTopicId(s.ctx)
	s.Require().NoError(err)
	s.Require().NotEqual(0, id)
}

func (s *RewardsTestSuite) FundAccount(amount int64, accAddress sdk.AccAddress) {
	initialStakeCoins := sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, cosmosMath.NewInt(amount)))
	err := s.bankKeeper.MintCoins(s.ctx, types.AlloraStakingAccountName, initialStakeCoins)
	s.Require().NoError(err)
	err = s.bankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.AlloraStakingAccountName, accAddress, initialStakeCoins)
	s.Require().NoError(err)
}

func TestRewardsTestSuite(t *testing.T) {
	suite.Run(t, new(RewardsTestSuite))
}

func (s *RewardsTestSuite) signValueBundle(
	reputerValueBundle *types.ValueBundle,
	privateKey secp256k1.PrivKey,
) []byte {
	require := s.Require()
	src := make([]byte, 0)
	src, err := reputerValueBundle.XXX_Marshal(src, true)
	require.NoError(err, "Marshall reputer value bundle should not return an error")
	valueBundleSignature, err := privateKey.Sign(src)
	require.NoError(err, "Sign should not return an error")
	return valueBundleSignature
}

func (s *RewardsTestSuite) MintTokensToAddress(address sdk.AccAddress, amount cosmosMath.Int) {
	creatorInitialBalanceCoins := sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, amount))

	err := s.bankKeeper.MintCoins(s.ctx, types.AlloraStakingAccountName, creatorInitialBalanceCoins)
	s.Require().NoError(err)
	err = s.bankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.AlloraStakingAccountName, address, creatorInitialBalanceCoins)
	s.Require().NoError(err)
}

func (s *RewardsTestSuite) MintTokensToModule(moduleName string, amount cosmosMath.Int) {
	creatorInitialBalanceCoins := sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, amount))
	err := s.bankKeeper.MintCoins(s.ctx, moduleName, creatorInitialBalanceCoins)
	s.Require().NoError(err)
}

func (s *RewardsTestSuite) RegisterAllWorkersOfPayload(topicId types.TopicId, payload *types.WorkerDataBundle) {
	worker := payload.InferenceForecastsBundle.Inference.Inferer
	// Define sample OffchainNode information for a worker
	workerInfo := types.OffchainNode{
		Owner:       s.addrsStr[4],
		NodeAddress: worker,
	}
	err := s.emissionsKeeper.InsertWorker(s.ctx, topicId, worker, workerInfo)
	s.Require().NoError(err)
	for _, element := range payload.InferenceForecastsBundle.Forecast.ForecastElements {
		worker = element.Inferer
		workerInfo = types.OffchainNode{
			Owner:       worker,
			NodeAddress: worker,
		}
		err := s.emissionsKeeper.InsertWorker(s.ctx, topicId, worker, workerInfo)
		s.Require().NoError(err)
	}
}

func (s *RewardsTestSuite) RegisterAllReputersOfPayload(topicId types.TopicId, payload *types.ReputerValueBundle) {
	reputer := payload.ValueBundle.Reputer
	// Define sample OffchainNode information for a worker
	reputerInfo := types.OffchainNode{
		Owner:       reputer,
		NodeAddress: reputer,
	}
	err := s.emissionsKeeper.InsertReputer(s.ctx, topicId, reputer, reputerInfo)
	s.Require().NoError(err)
}
func (s *RewardsTestSuite) TestStandardRewardEmission() {
	block := int64(600)
	s.ctx = s.ctx.WithBlockHeight(block)

	// Reputer Addresses
	reputerIndexes := s.returnIndexes(0, 5)

	// Worker Addresses
	workerIndexes := s.returnIndexes(5, 5)

	// Create topic
	newTopicMsg := &types.CreateNewTopicRequest{
		Creator:                  s.addrsStr[reputerIndexes[0]],
		Metadata:                 "test",
		LossMethod:               "mse",
		EpochLength:              10800,
		AllowNegative:            false,
		GroundTruthLag:           10800,
		WorkerSubmissionWindow:   10,
		AlphaRegret:              alloraMath.NewDecFromInt64(1),
		PNorm:                    alloraMath.NewDecFromInt64(3),
		Epsilon:                  alloraMath.MustNewDecFromString("0.01"),
		MeritSortitionAlpha:      alloraMath.MustNewDecFromString("0.1"),
		ActiveInfererQuantile:    alloraMath.MustNewDecFromString("0.2"),
		ActiveForecasterQuantile: alloraMath.MustNewDecFromString("0.2"),
		ActiveReputerQuantile:    alloraMath.MustNewDecFromString("0.2"),
		EnableWorkerWhitelist:    true,
		EnableReputerWhitelist:   true,
	}
	res, err := s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	s.Require().NoError(err)

	// Get Topic Id
	topicId := res.TopicId

	// Register 5 workers
	for _, index := range workerIndexes {
		workerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: false,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, workerRegMsg)
		s.Require().NoError(err)
	}

	// Register 5 reputers
	for _, index := range reputerIndexes {
		reputerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: true,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, reputerRegMsg)
		s.Require().NoError(err)
	}

	cosmosOneE18 := inferencesynthesis.CosmosIntOneE18()

	// Add Stake for reputers
	var stakes = []cosmosMath.Int{
		cosmosMath.NewInt(1176644).Mul(cosmosOneE18),
		cosmosMath.NewInt(384623).Mul(cosmosOneE18),
		cosmosMath.NewInt(394676).Mul(cosmosOneE18),
		cosmosMath.NewInt(207999).Mul(cosmosOneE18),
		cosmosMath.NewInt(368582).Mul(cosmosOneE18),
	}
	for i, index := range reputerIndexes {
		s.MintTokensToAddress(s.addrs[index], stakes[i])
		_, err := s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
			Sender:  s.addrsStr[index],
			Amount:  stakes[i],
			TopicId: topicId,
		})
		s.Require().NoError(err)
	}

	// Insert unfullfiled nonces
	err = s.emissionsKeeper.AddWorkerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)

	// Insert inference from workers
	inferenceBundles := generateWorkerDataBundles(s, block, topicId)
	for _, payload := range inferenceBundles {
		_, err = s.msgServer.InsertWorkerPayload(s.ctx, &types.InsertWorkerPayloadRequest{
			Sender:           payload.Worker,
			WorkerDataBundle: payload,
		})
		s.Require().NoError(err)
	}

	topic, err := s.emissionsKeeper.GetTopic(s.ctx, topicId)
	s.Require().NoError(err)

	newBlockheight := block + topic.GroundTruthLag
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(newBlockheight)

	// Insert loss bundle from reputers
	lossBundles := generateLossBundles(s, block, topicId, reputerIndexes)
	for _, payload := range lossBundles.ReputerValueBundles {
		_, _ = s.emissionsKeeper.FulfillWorkerNonce(s.ctx, topicId, payload.ValueBundle.ReputerRequestNonce.ReputerNonce)
		_ = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, payload.ValueBundle.ReputerRequestNonce.ReputerNonce)
		_, err = s.msgServer.InsertReputerPayload(s.ctx, &types.InsertReputerPayloadRequest{
			Sender:             payload.ValueBundle.Reputer,
			ReputerValueBundle: payload,
		})
		s.Require().NoError(err)
	}

	block = 600 + 10800
	s.ctx = s.ctx.WithBlockHeight(block)

	// Trigger end block - rewards distribution
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)
}

func (s *RewardsTestSuite) TestStandardRewardEmissionShouldRewardTopicsWithFulfilledNonces() {
	s.SetParamsForTest()
	block := int64(600)
	s.ctx = s.ctx.WithBlockHeight(block)

	// Reputer Addresses
	reputerIndexes := s.returnIndexes(0, 5)

	// Worker Addresses
	workerIndexes := s.returnIndexes(5, 5)

	// Create topic
	newTopicMsg := &types.CreateNewTopicRequest{
		Creator:                  s.addrsStr[reputerIndexes[0]],
		Metadata:                 "test",
		AllowNegative:            false,
		LossMethod:               "mse",
		EpochLength:              10800,
		GroundTruthLag:           10800,
		WorkerSubmissionWindow:   10,
		AlphaRegret:              alloraMath.NewDecFromInt64(1),
		PNorm:                    alloraMath.NewDecFromInt64(3),
		Epsilon:                  alloraMath.MustNewDecFromString("0.01"),
		MeritSortitionAlpha:      alloraMath.MustNewDecFromString("0.1"),
		ActiveInfererQuantile:    alloraMath.MustNewDecFromString("0.2"),
		ActiveForecasterQuantile: alloraMath.MustNewDecFromString("0.2"),
		ActiveReputerQuantile:    alloraMath.MustNewDecFromString("0.2"),
		EnableWorkerWhitelist:    true,
		EnableReputerWhitelist:   true,
	}
	res, err := s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	s.Require().NoError(err)

	// Get Topic Id
	topicId := res.TopicId

	// Register 5 workers
	for _, index := range workerIndexes {
		workerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: false,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, workerRegMsg)
		s.Require().NoError(err)
	}

	// Register 5 reputers
	for _, index := range reputerIndexes {
		reputerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: true,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, reputerRegMsg)
		s.Require().NoError(err)
	}

	cosmosOneE18 := inferencesynthesis.CosmosIntOneE18()

	// Add Stake for reputers
	var stakes = []cosmosMath.Int{
		cosmosMath.NewInt(1176644).Mul(cosmosOneE18),
		cosmosMath.NewInt(384623).Mul(cosmosOneE18),
		cosmosMath.NewInt(394676).Mul(cosmosOneE18),
		cosmosMath.NewInt(207999).Mul(cosmosOneE18),
		cosmosMath.NewInt(368582).Mul(cosmosOneE18),
	}
	for i, index := range reputerIndexes {
		s.MintTokensToAddress(s.addrs[index], stakes[i])
		_, err := s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
			Sender:  s.addrsStr[index],
			Amount:  stakes[i],
			TopicId: topicId,
		})
		s.Require().NoError(err)
	}

	initialStake := cosmosMath.NewInt(1000)
	s.MintTokensToAddress(s.addrs[reputerIndexes[0]], initialStake)
	fundTopicMessage := types.FundTopicRequest{
		Sender:  s.addrsStr[reputerIndexes[0]],
		TopicId: topicId,
		Amount:  initialStake,
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
	s.Require().NoError(err)
	s.Require().True(
		s.bankKeeper.HasBalance(
			s.ctx,
			s.accountKeeper.GetModuleAddress(minttypes.EcosystemModuleName),
			sdk.NewCoin(params.DefaultBondDenom, initialStake),
		),
		"ecosystem account should have something in it after funding",
	)

	// Insert unfullfiled nonces
	err = s.emissionsKeeper.AddWorkerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)

	// Insert inference from workers
	inferenceBundles := generateWorkerDataBundles(s, block, topicId)
	for _, payload := range inferenceBundles {
		_, err = s.msgServer.InsertWorkerPayload(s.ctx, &types.InsertWorkerPayloadRequest{
			Sender:           payload.Worker,
			WorkerDataBundle: payload,
		})
		s.Require().NoError(err)
	}

	topic, err := s.emissionsKeeper.GetTopic(s.ctx, topicId)
	s.Require().NoError(err)

	newBlockheight := block + topic.GroundTruthLag
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(newBlockheight)

	// Trigger end block - rewards distribution
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)
	// Insert loss bundle from reputers
	lossBundles := generateLossBundles(s, block, topicId, reputerIndexes)
	for _, payload := range lossBundles.ReputerValueBundles {
		_, _ = s.emissionsKeeper.FulfillWorkerNonce(s.ctx, topicId, payload.ValueBundle.ReputerRequestNonce.ReputerNonce)
		_ = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, payload.ValueBundle.ReputerRequestNonce.ReputerNonce)
		_, err = s.msgServer.InsertReputerPayload(s.ctx, &types.InsertReputerPayloadRequest{
			Sender:             payload.ValueBundle.Reputer,
			ReputerValueBundle: payload,
		})
		s.Require().NoError(err)
	}

	// Create topic 2
	// Reputer Addresses
	reputerIndexes = s.returnIndexes(10, 5)

	// Worker Addresses
	workerIndexes = s.returnIndexes(15, 5)

	// Create topic
	newTopicMsg = &types.CreateNewTopicRequest{
		Creator:                  s.addrsStr[reputerIndexes[0]],
		Metadata:                 "test",
		LossMethod:               "mse",
		AllowNegative:            false,
		EpochLength:              10800,
		GroundTruthLag:           10800,
		WorkerSubmissionWindow:   10,
		AlphaRegret:              alloraMath.NewDecFromInt64(1),
		PNorm:                    alloraMath.NewDecFromInt64(3),
		Epsilon:                  alloraMath.MustNewDecFromString("0.01"),
		MeritSortitionAlpha:      alloraMath.MustNewDecFromString("0.1"),
		ActiveInfererQuantile:    alloraMath.MustNewDecFromString("0.2"),
		ActiveForecasterQuantile: alloraMath.MustNewDecFromString("0.2"),
		ActiveReputerQuantile:    alloraMath.MustNewDecFromString("0.2"),
		EnableWorkerWhitelist:    true,
		EnableReputerWhitelist:   true,
	}
	res, err = s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	s.Require().NoError(err)

	// Get Topic Id
	topicId2 := res.TopicId

	// Register 5 workers
	for _, index := range workerIndexes {
		workerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId2,
			IsReputer: false,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, workerRegMsg)
		s.Require().NoError(err)
	}

	// Register 5 reputers
	for _, index := range reputerIndexes {
		reputerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId2,
			IsReputer: true,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, reputerRegMsg)
		s.Require().NoError(err)
	}

	for i, index := range reputerIndexes {
		s.MintTokensToAddress(s.addrs[index], stakes[i])
		_, err := s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
			Sender:  s.addrsStr[index],
			Amount:  stakes[i],
			TopicId: topicId2,
		})
		s.Require().NoError(err)
	}

	initialStake = cosmosMath.NewInt(1000)
	s.MintTokensToAddress(s.addrs[reputerIndexes[0]], initialStake)
	fundTopicMessage = types.FundTopicRequest{
		Sender:  s.addrsStr[reputerIndexes[0]],
		TopicId: topicId2,
		Amount:  initialStake,
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
	s.Require().NoError(err)

	// Insert unfullfiled nonces
	err = s.emissionsKeeper.AddWorkerNonce(s.ctx, topicId2, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId2, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)

	// Do not send bundles for topic 2 yet

	beforeRewardsTopic1FeeRevenue, err := s.emissionsKeeper.GetTopicFeeRevenue(s.ctx, topicId)
	s.Require().NoError(err)
	beforeRewardsTopic2FeeRevenue, err := s.emissionsKeeper.GetTopicFeeRevenue(s.ctx, topicId2)
	s.Require().NoError(err)

	// mint some rewards to give out
	s.MintTokensToModule(types.AlloraRewardsAccountName, cosmosMath.NewInt(1000))

	newBlockheight += topic.GroundTruthLag
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(newBlockheight)

	// Trigger end block - rewards distribution
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)

	afterRewardsTopic1FeeRevenue, err := s.emissionsKeeper.GetTopicFeeRevenue(s.ctx, topicId)
	s.Require().NoError(err)
	afterRewardsTopic2FeeRevenue, err := s.emissionsKeeper.GetTopicFeeRevenue(s.ctx, topicId2)
	s.Require().NoError(err)

	// Topic 1 should have less revenue after rewards distribution -> rewards distributed
	s.Require().True(
		beforeRewardsTopic1FeeRevenue.GT(afterRewardsTopic1FeeRevenue),
		"Topic 1 should lose influence of their fee revenue: %s > %s",
		beforeRewardsTopic1FeeRevenue.String(),
		afterRewardsTopic1FeeRevenue.String(),
	)
	// Topic 2 should also have less revenue after rewards distribution as topic rewards
	// are shared among all topics whose epoch lengths modulo the current block height are 0
	s.Require().True(
		beforeRewardsTopic2FeeRevenue.GT(afterRewardsTopic2FeeRevenue),
		"Topic 2 should lose influence of their fee revenue: %s > %s",
		beforeRewardsTopic2FeeRevenue.String(),
		afterRewardsTopic2FeeRevenue.String(),
	)
}

func (s *RewardsTestSuite) setUpTopic(
	blockHeight int64,
	workerIndexes []int,
	reputerIndexes []int,
	stake cosmosMath.Int,
	alphaRegret alloraMath.Dec,
) uint64 {
	return s.setUpTopicWithEpochLength(blockHeight, workerIndexes, reputerIndexes, stake, alphaRegret, 10800)
}

func (s *RewardsTestSuite) setUpTopicWithEpochLength(
	blockHeight int64,
	workerIndexes []int,
	reputerIndexes []int,
	stake cosmosMath.Int,
	alphaRegret alloraMath.Dec,
	epochLength int64,
) uint64 {
	require := s.Require()
	s.ctx = s.ctx.WithBlockHeight(blockHeight)

	// Create topic
	newTopicMsg := &types.CreateNewTopicRequest{
		Creator:                  s.addrsStr[reputerIndexes[0]],
		Metadata:                 "test",
		LossMethod:               "mse",
		EpochLength:              epochLength,
		AllowNegative:            false,
		GroundTruthLag:           epochLength,
		WorkerSubmissionWindow:   min(10, epochLength-2),
		AlphaRegret:              alphaRegret,
		PNorm:                    alloraMath.NewDecFromInt64(3),
		Epsilon:                  alloraMath.MustNewDecFromString("0.01"),
		MeritSortitionAlpha:      alloraMath.MustNewDecFromString("0.1"),
		ActiveInfererQuantile:    alloraMath.MustNewDecFromString("0.2"),
		ActiveForecasterQuantile: alloraMath.MustNewDecFromString("0.2"),
		ActiveReputerQuantile:    alloraMath.MustNewDecFromString("0.2"),
		EnableWorkerWhitelist:    true,
		EnableReputerWhitelist:   true,
	}
	res, err := s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	require.NoError(err)

	// Get Topic Id
	topicId := res.TopicId

	for _, index := range workerIndexes {
		workerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: false,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, workerRegMsg)
		require.NoError(err)
	}

	for _, index := range reputerIndexes {
		reputerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: true,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, reputerRegMsg)
		require.NoError(err)
	}
	for _, index := range reputerIndexes {
		s.MintTokensToAddress(s.addrs[index], stake)
		_, err := s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
			Sender:  s.addrsStr[index],
			Amount:  stake,
			TopicId: topicId,
		})
		require.NoError(err)
	}

	var initialStake int64 = 1000
	s.MintTokensToAddress(s.addrs[reputerIndexes[0]], cosmosMath.NewInt(initialStake))
	fundTopicMessage := types.FundTopicRequest{
		Sender:  s.addrsStr[reputerIndexes[0]],
		TopicId: topicId,
		Amount:  cosmosMath.NewInt(initialStake),
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
	require.NoError(err)

	return topicId
}

func (s *RewardsTestSuite) getRewardsDistribution(
	topicId uint64,
	workerValues []TestWorkerValue,
	reputerValues []TestWorkerValue,
	workerZeroAddress sdk.AccAddress,
	workerZeroOneOutInfererValue string,
	workerZeroInfererValue string,
) []types.TaskReward {
	require := s.Require()

	params, err := s.emissionsKeeper.GetParams(s.ctx)
	require.NoError(err)

	// Move to end of this epoch block
	nextBlock, _, err := s.emissionsKeeper.GetNextPossibleChurningBlockByTopicId(s.ctx, topicId)
	s.T().Logf("Next block: %v", nextBlock)
	s.Require().NoError(err)
	blockHeight := nextBlock
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(blockHeight)
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	require.NoError(err)

	topic, err := s.emissionsKeeper.GetTopic(s.ctx, topicId)
	s.Require().NoError(err)

	// Advance one to send the worker data
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(blockHeight + 1)

	getIndexesFromValues := func(values []TestWorkerValue) []int {
		indexes := make([]int, 0)
		for _, value := range values {
			indexes = append(indexes, value.Index)
		}
		return indexes
	}

	workerIndexes := getIndexesFromValues(workerValues)

	inferenceBundles := generateSimpleWorkerDataBundles(s, topicId, topic.EpochLastEnded, blockHeight, workerValues, workerIndexes)
	for _, payload := range inferenceBundles {
		s.RegisterAllWorkersOfPayload(topicId, payload)
		_, err = s.msgServer.InsertWorkerPayload(s.ctx, &types.InsertWorkerPayloadRequest{
			Sender:           payload.Worker,
			WorkerDataBundle: payload,
		})
		require.NoError(err)
	}
	// Advance to close the window
	newBlock := blockHeight + topic.WorkerSubmissionWindow
	s.T().Logf("SubmissionWindow Next block: %v", nextBlock)
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(newBlock)
	// EndBlock closes the nonce
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)

	// Insert loss bundle from reputers
	lossBundles := generateSimpleLossBundles(
		s,
		topicId,
		topic.EpochLastEnded,
		workerValues,
		reputerValues,
		workerZeroAddress,
		workerZeroOneOutInfererValue,
		workerZeroInfererValue,
	)

	newBlockheight := blockHeight + topic.GroundTruthLag
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(newBlockheight)
	for _, payload := range lossBundles.ReputerValueBundles {
		s.RegisterAllReputersOfPayload(topicId, payload)
		_, err = s.msgServer.InsertReputerPayload(s.ctx, &types.InsertReputerPayloadRequest{
			Sender:             payload.ValueBundle.Reputer,
			ReputerValueBundle: payload,
		})
		require.NoError(err)
	}
	err = actorutils.CloseReputerNonce(
		&s.emissionsKeeper, s.ctx, topic,
		*lossBundles.ReputerValueBundles[0].ValueBundle.ReputerRequestNonce.ReputerNonce,
	)
	s.Require().NoError(err)

	topicTotalRewards := alloraMath.NewDecFromInt64(1000000)

	rewardsDistributionByTopicParticipant, _, err := rewards.GenerateRewardsDistributionByTopicParticipant(
		rewards.GenerateRewardsDistributionByTopicParticipantArgs{
			Ctx:          s.ctx,
			K:            s.emissionsKeeper,
			TopicId:      topicId,
			TopicReward:  &topicTotalRewards,
			BlockHeight:  blockHeight,
			ModuleParams: params,
		},
	)
	require.NoError(err)

	return rewardsDistributionByTopicParticipant
}

func areTaskRewardsEqualIgnoringTopicId(s *RewardsTestSuite, A []types.TaskReward, B []types.TaskReward) bool {
	if len(A) != len(B) {
		s.Fail("Lengths are different")
	}

	for _, taskRewardA := range A {
		found := false
		for _, taskRewardB := range B {
			if taskRewardA.Address == taskRewardB.Address && taskRewardA.Type == taskRewardB.Type {
				if found {
					s.Fail("Worker %v found twice", taskRewardA.Address)
				}
				found = true
				inDelta, err := alloraMath.InDelta(
					taskRewardA.Reward,
					taskRewardB.Reward,
					alloraMath.MustNewDecFromString("0.00001"),
				)
				if err != nil {
					s.Fail("Error finding out if taskRewardA in delta with taskRewardB",
						taskRewardA.Reward.String(),
						taskRewardB.Reward.String(),
					)
				}
				if !inDelta {
					return false
				}
			}
		}
		if !found {
			s.T().Logf("Worker %v not found", taskRewardA.Address)
			return false
		}
	}

	return true
}

// We have 2 trials with 2 epochs each, and the first worker does better in the 2nd epoch in both trials.
// We show that keeping the TaskRewardAlpha the same means that the worker is rewarded the same amount
// in both cases.
// This is a sanity test to ensure that we are isolating the effect of TaskRewardAlpha in subsequent tests.
func (s *RewardsTestSuite) TestFixingTaskRewardAlphaDoesNotChangePerformanceImportanceOfPastVsPresent() {
	/// SETUP
	require := s.Require()
	k := s.emissionsKeeper

	currentParams, err := k.GetParams(s.ctx)
	require.NoError(err)

	blockHeight0 := int64(100)
	blockHeightDelta := int64(1)
	s.ctx = s.ctx.WithBlockHeight(blockHeight0)

	workerIndexes := s.returnIndexes(0, 3)
	reputerIndexes := s.returnIndexes(3, 3)

	stake := cosmosMath.NewInt(1000000000000000000).Mul(inferencesynthesis.CosmosIntOneE18())

	alphaRegret := alloraMath.MustNewDecFromString("0.1")
	topicId := s.setUpTopic(blockHeight0, workerIndexes, reputerIndexes, stake, alphaRegret)

	workerValues := []TestWorkerValue{
		{Index: 0, Value: "0.1"},
		{Index: 1, Value: "0.2"},
		{Index: 2, Value: "0.3"},
	}

	reputerValues := []TestWorkerValue{
		{Index: 3, Value: "0.1"},
		{Index: 4, Value: "0.2"},
		{Index: 5, Value: "0.3"},
	}

	currentParams.TaskRewardAlpha = alloraMath.MustNewDecFromString(("0.1"))
	err = k.SetParams(s.ctx, currentParams)
	require.NoError(err)

	/// TEST 0 PART A

	rewardsDistribution0_0 := s.getRewardsDistribution(
		topicId,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.1",
	)

	/// TEST 0 PART B

	blockHeight1 := blockHeight0 + blockHeightDelta
	s.ctx = s.ctx.WithBlockHeight(blockHeight1)

	rewardsDistribution0_1 := s.getRewardsDistribution(
		topicId,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.2",
		"0.1",
	)

	/// TEST 1 PART A

	blockHeight2 := blockHeight1 + blockHeightDelta
	s.ctx = s.ctx.WithBlockHeight(blockHeight2)

	topicId1 := s.setUpTopic(blockHeight0, workerIndexes, reputerIndexes, stake, alphaRegret)

	rewardsDistribution1_0 := s.getRewardsDistribution(
		topicId1,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.1",
	)

	/// TEST 1 PART B

	blockHeight3 := blockHeight2 + blockHeightDelta
	s.ctx = s.ctx.WithBlockHeight(blockHeight3)

	rewardsDistribution1_1 := s.getRewardsDistribution(
		topicId1,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.2",
		"0.1",
	)

	require.True(areTaskRewardsEqualIgnoringTopicId(s, rewardsDistribution0_0, rewardsDistribution1_0))
	require.True(areTaskRewardsEqualIgnoringTopicId(s, rewardsDistribution0_1, rewardsDistribution1_1))
}

// We have 2 trials with 2 epochs each, and the first worker does better in the 2nd epoch in both trials,
// due to a worse one out inferer value, indicating that the network is better off with the worker.
// We increase TaskRewardAlpha between the trials to show that weighting current performance more heavily
// means that the worker is rewarded more for their better performance in the 2nd epoch of the 2nd trial.
func (s *RewardsTestSuite) TestIncreasingTaskRewardAlphaIncreasesImportanceOfPresentPerformance() {
	require := s.Require()
	k := s.emissionsKeeper

	currentParams, err := k.GetParams(s.ctx)
	require.NoError(err)

	blockHeight0 := int64(100)
	blockHeightDelta := int64(1)
	s.ctx = s.ctx.WithBlockHeight(blockHeight0)

	workerIndexes := s.returnIndexes(0, 3)
	reputerIndexes := s.returnIndexes(3, 3)

	stake := cosmosMath.NewInt(1000000000000000000).Mul(inferencesynthesis.CosmosIntOneE18())

	alphaRegret := alloraMath.MustNewDecFromString("0.1")
	topicId := s.setUpTopic(blockHeight0, workerIndexes, reputerIndexes, stake, alphaRegret)

	workerValues := []TestWorkerValue{
		{Index: workerIndexes[0], Value: "0.1"},
		{Index: workerIndexes[1], Value: "0.2"},
		{Index: workerIndexes[2], Value: "0.3"},
	}

	reputerValues := []TestWorkerValue{
		{Index: reputerIndexes[0], Value: "0.1"},
		{Index: reputerIndexes[1], Value: "0.2"},
		{Index: reputerIndexes[2], Value: "0.3"},
	}

	currentParams.TaskRewardAlpha = alloraMath.MustNewDecFromString("0.1")
	err = k.SetParams(s.ctx, currentParams)
	require.NoError(err)

	/// TEST 0 PART A

	rewardsDistribution0_0 := s.getRewardsDistribution(
		topicId,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.1",
	)

	/// TEST 0 PART B

	blockHeight1 := blockHeight0 + blockHeightDelta
	s.ctx = s.ctx.WithBlockHeight(blockHeight1)

	rewardsDistribution0_1 := s.getRewardsDistribution(
		topicId,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.2",
		"0.1",
	)

	/// CHANGE TASK REWARD ALPHA

	currentParams.TaskRewardAlpha = alloraMath.MustNewDecFromString(("0.2"))
	err = k.SetParams(s.ctx, currentParams)
	require.NoError(err)

	/// TEST 1 PART A

	blockHeight2 := blockHeight1 + blockHeightDelta
	s.ctx = s.ctx.WithBlockHeight(blockHeight2)

	topicId1 := s.setUpTopic(blockHeight2, workerIndexes, reputerIndexes, stake, alphaRegret)

	rewardsDistribution1_0 := s.getRewardsDistribution(
		topicId1,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.1",
	)

	/// TEST 1 PART B

	blockHeight3 := blockHeight2 + blockHeightDelta
	s.ctx = s.ctx.WithBlockHeight(blockHeight3)

	rewardsDistribution1_1 := s.getRewardsDistribution(
		topicId1,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.2",
		"0.1",
	)

	require.True(areTaskRewardsEqualIgnoringTopicId(s, rewardsDistribution0_0, rewardsDistribution1_0))
	require.False(areTaskRewardsEqualIgnoringTopicId(s, rewardsDistribution0_1, rewardsDistribution1_1))

	var workerReward_0_0_1_Reward alloraMath.Dec
	found := false
	for _, reward := range rewardsDistribution0_1 {
		if reward.Address == s.addrsStr[workerIndexes[0]] && reward.Type == types.WorkerInferenceRewardType {
			found = true
			workerReward_0_0_1_Reward = reward.Reward
		}
	}
	if !found {
		require.Fail("Worker not found")
	}

	var workerReward_0_1_1_Reward alloraMath.Dec
	found = false
	for _, reward := range rewardsDistribution1_1 {
		if reward.Address == s.addrsStr[workerIndexes[0]] && reward.Type == types.WorkerInferenceRewardType {
			found = true
			workerReward_0_1_1_Reward = reward.Reward
		}
	}
	if !found {
		require.Fail("Worker not found")
	}

	require.True(workerReward_0_0_1_Reward.Lt(workerReward_0_1_1_Reward))
}

// We have 2 trials with 2 epochs each, and the first worker does worse in 2nd epoch in both trials,
// enacted by their increasing loss between epochs.
// We increase alpha between the trials to prove that their worsening performance decreases regret.
// This is somewhat counterintuitive, but can be explained by the following passage from the litepaper:
// "A positive regret implies that the inference of worker j is expected by worker k to outperform
// the network’s previously reported accuracy, whereas a negative regret indicates that the network
// is expected to be more accurate."
func (s *RewardsTestSuite) TestIncreasingAlphaRegretIncreasesPresentEffectOnRegret() {
	/// SETUP
	require := s.Require()
	k := s.emissionsKeeper

	currentParams, err := k.GetParams(s.ctx)
	require.NoError(err)

	blockHeight0 := int64(100)
	blockHeightDelta := int64(1)
	s.ctx = s.ctx.WithBlockHeight(blockHeight0)

	workerIndexes := s.returnIndexes(0, 3)
	reputerIndexes := s.returnIndexes(3, 3)

	stake := cosmosMath.NewInt(1000000000000000000).Mul(inferencesynthesis.CosmosIntOneE18())

	alphaRegret := alloraMath.MustNewDecFromString("0.1")
	topicId0 := s.setUpTopic(blockHeight0, workerIndexes, reputerIndexes, stake, alphaRegret)

	workerValues := []TestWorkerValue{
		{Index: workerIndexes[0], Value: "0.1"},
		{Index: workerIndexes[1], Value: "0.2"},
		{Index: workerIndexes[2], Value: "0.3"},
	}

	reputerValues := []TestWorkerValue{
		{Index: reputerIndexes[0], Value: "0.1"},
		{Index: reputerIndexes[1], Value: "0.2"},
		{Index: reputerIndexes[2], Value: "0.3"},
	}

	topic, err := k.GetTopic(s.ctx, topicId0)
	s.Require().NoError(err)
	topic.AlphaRegret = alloraMath.MustNewDecFromString("0.1")
	err = k.SetTopic(s.ctx, topicId0, topic)
	require.NoError(err)

	/// TEST 0 PART A

	s.getRewardsDistribution(
		topicId0,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.1",
	)

	worker0_0, _, err := k.GetInfererNetworkRegret(s.ctx, topicId0, s.addrsStr[workerIndexes[0]])
	require.NoError(err)

	/// TEST 0 PART B

	blockHeight1 := blockHeight0 + blockHeightDelta
	s.ctx = s.ctx.WithBlockHeight(blockHeight1)

	s.getRewardsDistribution(
		topicId0,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.2",
	)

	worker0_0FirstRegret := worker0_0.Value

	worker0_0, _, err = k.GetInfererNetworkRegret(s.ctx, topicId0, s.addrsStr[workerIndexes[0]])
	require.NoError(err)

	worker1_0, _, err := k.GetInfererNetworkRegret(s.ctx, topicId0, s.addrsStr[workerIndexes[1]])
	require.NoError(err)

	worker2_0, _, err := k.GetInfererNetworkRegret(s.ctx, topicId0, s.addrsStr[workerIndexes[2]])
	require.NoError(err)

	worker0_0RegretDecrease, err := worker0_0FirstRegret.Sub(worker0_0.Value)
	require.NoError(err)
	worker0_0RegretDecreaseRate, err := worker0_0RegretDecrease.Quo(worker0_0FirstRegret)
	require.NoError(err)

	require.True(worker0_0RegretDecreaseRate.Equal(alphaRegret))

	/// INCREASE ALPHA REGRET

	alphaRegret = alloraMath.MustNewDecFromString(("0.2"))
	err = k.SetParams(s.ctx, currentParams)
	require.NoError(err)

	/// TEST 1 PART A

	blockHeight2 := blockHeight1 + blockHeightDelta
	s.ctx = s.ctx.WithBlockHeight(blockHeight2)

	topicId1 := s.setUpTopic(blockHeight2, workerIndexes, reputerIndexes, stake, alphaRegret)

	s.getRewardsDistribution(
		topicId1,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.1",
	)

	worker0_1, _, err := k.GetInfererNetworkRegret(s.ctx, topicId1, s.addrsStr[workerIndexes[0]])
	require.NoError(err)

	/// TEST 1 PART B

	blockHeight3 := blockHeight2 + blockHeightDelta
	s.ctx = s.ctx.WithBlockHeight(blockHeight3)

	s.getRewardsDistribution(
		topicId1,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.2",
	)

	worker0_1FirstRegret := worker0_1.Value

	worker0_1, _, err = k.GetInfererNetworkRegret(s.ctx, topicId1, s.addrsStr[workerIndexes[0]])
	require.NoError(err)

	worker1_1, _, err := k.GetInfererNetworkRegret(s.ctx, topicId1, s.addrsStr[workerIndexes[1]])
	require.NoError(err)

	worker2_1, _, err := k.GetInfererNetworkRegret(s.ctx, topicId1, s.addrsStr[workerIndexes[2]])
	require.NoError(err)

	worker0_1RegretDecrease, err := worker0_1FirstRegret.Sub(worker0_1.Value)
	require.NoError(err)
	worker0_1RegretDecreaseRate, err := worker0_1RegretDecrease.Quo(worker0_1FirstRegret)
	require.NoError(err)
	require.True(worker0_1RegretDecreaseRate.Equal(alphaRegret))

	// Check alpha impact in regrets - Topic 0 (alpha 0.1) vs Topic 1 (alpha 0.2)
	require.True(worker0_1RegretDecreaseRate.Gt(worker0_0RegretDecreaseRate))

	require.True(alloraMath.InDelta(worker1_0.Value, worker1_1.Value, alloraMath.MustNewDecFromString("0.00001")))
	require.True(worker2_0.Value.Gt(worker2_1.Value))
}

func (s *RewardsTestSuite) TestGenerateTasksRewardsShouldIncreaseRewardShareIfMoreParticipants() {
	block := int64(100)
	s.ctx = s.ctx.WithBlockHeight(block)

	reputerIndexes := s.returnIndexes(0, 3)
	workerIndexes := s.returnIndexes(5, 5)

	cosmosOneE18 := inferencesynthesis.CosmosIntOneE18()

	stakes := []cosmosMath.Int{
		cosmosMath.NewInt(1000000000000000000).Mul(cosmosOneE18),
		cosmosMath.NewInt(1000000000000000000).Mul(cosmosOneE18),
		cosmosMath.NewInt(1000000000000000000).Mul(cosmosOneE18),
	}

	// Create topic
	newTopicMsg := &types.CreateNewTopicRequest{
		Creator:                  s.addrsStr[reputerIndexes[0]],
		Metadata:                 "test",
		LossMethod:               "mse",
		AllowNegative:            false,
		EpochLength:              10800,
		GroundTruthLag:           10800,
		WorkerSubmissionWindow:   10,
		AlphaRegret:              alloraMath.NewDecFromInt64(1),
		PNorm:                    alloraMath.NewDecFromInt64(3),
		Epsilon:                  alloraMath.MustNewDecFromString("0.01"),
		MeritSortitionAlpha:      alloraMath.MustNewDecFromString("0.1"),
		ActiveInfererQuantile:    alloraMath.MustNewDecFromString("0.2"),
		ActiveForecasterQuantile: alloraMath.MustNewDecFromString("0.2"),
		ActiveReputerQuantile:    alloraMath.MustNewDecFromString("0.2"),
		EnableWorkerWhitelist:    true,
		EnableReputerWhitelist:   true,
	}
	res, err := s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	s.Require().NoError(err)

	// Get Topic Id
	topicId := res.TopicId

	// Register 5 workers
	for _, index := range workerIndexes {
		workerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: false,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, workerRegMsg)
		s.Require().NoError(err)
	}

	// Register 3 reputers
	for _, index := range reputerIndexes {
		reputerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: true,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, reputerRegMsg)
		s.Require().NoError(err)
	}
	// Add Stake for reputers
	for i, index := range reputerIndexes {
		s.MintTokensToAddress(s.addrs[index], stakes[i])
		_, err := s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
			Sender:  s.addrsStr[index],
			Amount:  stakes[i],
			TopicId: topicId,
		})
		s.Require().NoError(err)
	}

	var initialStake int64 = 1000
	s.FundAccount(initialStake, s.addrs[reputerIndexes[0]])
	fundTopicMessage := types.FundTopicRequest{
		Sender:  s.addrsStr[reputerIndexes[0]],
		TopicId: topicId,
		Amount:  cosmosMath.NewInt(initialStake),
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
	s.Require().NoError(err)

	err = s.emissionsKeeper.AddWorkerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)

	// Insert inference from workers
	inferenceBundles := generateWorkerDataBundles(s, block, topicId)
	for _, payload := range inferenceBundles {
		_, err = s.msgServer.InsertWorkerPayload(s.ctx, &types.InsertWorkerPayloadRequest{
			Sender:           payload.Worker,
			WorkerDataBundle: payload,
		})
		s.Require().NoError(err)
	}

	topic, err := s.emissionsKeeper.GetTopic(s.ctx, topicId)
	s.Require().NoError(err)

	err = actorutils.CloseWorkerNonce(&s.emissionsKeeper, s.ctx, topic, *inferenceBundles[0].Nonce)
	s.Require().NoError(err)

	newBlockheight := block + topic.GroundTruthLag
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(newBlockheight)
	// Trigger end block - rewards distribution
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)

	// Insert loss bundle from reputers
	lossBundles := generateLossBundles(s, block, topicId, reputerIndexes)
	for _, payload := range lossBundles.ReputerValueBundles {
		_, _ = s.emissionsKeeper.FulfillWorkerNonce(s.ctx, topicId, payload.ValueBundle.ReputerRequestNonce.ReputerNonce)
		_ = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, payload.ValueBundle.ReputerRequestNonce.ReputerNonce)
		_, err = s.msgServer.InsertReputerPayload(s.ctx, &types.InsertReputerPayloadRequest{
			Sender:             payload.ValueBundle.Reputer,
			ReputerValueBundle: payload,
		})
		s.Require().NoError(err)
	}
	err = actorutils.CloseReputerNonce(
		&s.emissionsKeeper, s.ctx, topic,
		*lossBundles.ReputerValueBundles[0].ValueBundle.ReputerRequestNonce.ReputerNonce,
	)
	s.Require().NoError(err)

	topicTotalRewards := alloraMath.NewDecFromInt64(1000000)
	params, err := s.emissionsKeeper.GetParams(s.ctx)
	s.Require().NoError(err)

	firstRewardsDistribution, firstTotalReputerReward, err := rewards.GenerateRewardsDistributionByTopicParticipant(rewards.GenerateRewardsDistributionByTopicParticipantArgs{
		Ctx:          s.ctx,
		K:            s.emissionsKeeper,
		TopicId:      topicId,
		TopicReward:  &topicTotalRewards,
		BlockHeight:  block,
		ModuleParams: params,
	})
	s.Require().NoError(err)

	calcFirstTotalReputerReward := alloraMath.ZeroDec()
	for _, reward := range firstRewardsDistribution {
		if reward.Type == types.ReputerAndDelegatorRewardType {
			calcFirstTotalReputerReward, err = calcFirstTotalReputerReward.Add(reward.Reward)
			s.Require().NoError(err)
		}
	}
	inDelta, err := alloraMath.InDelta(
		firstTotalReputerReward,
		calcFirstTotalReputerReward,
		alloraMath.MustNewDecFromString("0.0001"),
	)
	s.Require().NoError(err)
	s.Require().True(
		inDelta,
		"expected: %s, got: %s",
		firstTotalReputerReward.String(),
		calcFirstTotalReputerReward.String(),
	)

	block += 1
	s.ctx = s.ctx.WithBlockHeight(block)

	// Add new reputers and stakes
	newReputerIndexes := []int{3, 4}
	reputerIndexes = append(reputerIndexes, newReputerIndexes...)

	// Add Stake for new reputers
	newStakes := []cosmosMath.Int{
		cosmosMath.NewInt(1000000000000000000).Mul(cosmosOneE18),
		cosmosMath.NewInt(1000000000000000000).Mul(cosmosOneE18),
	}
	stakes = append(stakes, newStakes...)

	// Create new topic
	newTopicMsg = &types.CreateNewTopicRequest{
		Creator:                  s.addrsStr[reputerIndexes[0]],
		Metadata:                 "test",
		LossMethod:               "mse",
		AllowNegative:            false,
		EpochLength:              10800,
		GroundTruthLag:           10800,
		WorkerSubmissionWindow:   10,
		AlphaRegret:              alloraMath.NewDecFromInt64(1),
		PNorm:                    alloraMath.NewDecFromInt64(3),
		Epsilon:                  alloraMath.MustNewDecFromString("0.01"),
		MeritSortitionAlpha:      alloraMath.MustNewDecFromString("0.1"),
		ActiveInfererQuantile:    alloraMath.MustNewDecFromString("0.2"),
		ActiveForecasterQuantile: alloraMath.MustNewDecFromString("0.2"),
		ActiveReputerQuantile:    alloraMath.MustNewDecFromString("0.2"),
		EnableWorkerWhitelist:    true,
		EnableReputerWhitelist:   true,
	}
	res, err = s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	s.Require().NoError(err)

	// Get Topic Id
	topicId = res.TopicId

	// Register 5 workers
	for _, index := range workerIndexes {
		workerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: false,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, workerRegMsg)
		s.Require().NoError(err)
	}

	// Register 5 reputers
	for _, index := range reputerIndexes {
		reputerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: true,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, reputerRegMsg)
		s.Require().NoError(err)
	}
	// Add Stake for reputers
	for i, index := range reputerIndexes {
		s.MintTokensToAddress(s.addrs[index], stakes[i])
		_, err := s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
			Sender:  s.addrsStr[index],
			Amount:  stakes[i],
			TopicId: topicId,
		})
		s.Require().NoError(err)
	}

	s.FundAccount(initialStake, s.addrs[reputerIndexes[0]])

	fundTopicMessage = types.FundTopicRequest{
		Sender:  s.addrsStr[reputerIndexes[0]],
		TopicId: topicId,
		Amount:  cosmosMath.NewInt(initialStake),
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
	s.Require().NoError(err)

	err = s.emissionsKeeper.AddWorkerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)

	// Insert inference from workers
	inferenceBundles = generateWorkerDataBundles(s, block, topicId)
	for _, payload := range inferenceBundles {
		_, err = s.msgServer.InsertWorkerPayload(s.ctx, &types.InsertWorkerPayloadRequest{
			Sender:           payload.Worker,
			WorkerDataBundle: payload,
		})
		s.Require().NoError(err)
	}

	topic, err = s.emissionsKeeper.GetTopic(s.ctx, topicId)
	s.Require().NoError(err)

	err = actorutils.CloseWorkerNonce(&s.emissionsKeeper, s.ctx, topic, *inferenceBundles[0].Nonce)
	s.Require().NoError(err)

	newBlockheight += topic.GroundTruthLag - 1
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(newBlockheight)
	// Trigger end block - rewards distribution
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)

	// Insert loss bundle from reputers
	lossBundles = generateLossBundles(s, block, topicId, reputerIndexes)
	for _, payload := range lossBundles.ReputerValueBundles {
		_, _ = s.emissionsKeeper.FulfillWorkerNonce(s.ctx, topicId, payload.ValueBundle.ReputerRequestNonce.ReputerNonce)
		_ = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, payload.ValueBundle.ReputerRequestNonce.ReputerNonce)
		_, err = s.msgServer.InsertReputerPayload(s.ctx, &types.InsertReputerPayloadRequest{
			Sender:             payload.ValueBundle.Reputer,
			ReputerValueBundle: payload,
		})
		s.Require().NoError(err)
	}
	err = actorutils.CloseReputerNonce(
		&s.emissionsKeeper, s.ctx, topic,
		*lossBundles.ReputerValueBundles[0].ValueBundle.ReputerRequestNonce.ReputerNonce,
	)
	s.Require().NoError(err)

	secondRewardsDistribution, secondTotalReputerReward, err := rewards.GenerateRewardsDistributionByTopicParticipant(
		rewards.GenerateRewardsDistributionByTopicParticipantArgs{
			Ctx:          s.ctx,
			K:            s.emissionsKeeper,
			TopicId:      topicId,
			TopicReward:  &topicTotalRewards,
			BlockHeight:  block,
			ModuleParams: params,
		})
	s.Require().NoError(err)

	calcSecondTotalReputerReward := alloraMath.ZeroDec()
	for _, reward := range secondRewardsDistribution {
		if reward.Type == types.ReputerAndDelegatorRewardType {
			calcSecondTotalReputerReward, err = calcSecondTotalReputerReward.Add(reward.Reward)
			s.Require().NoError(err)
		}
	}
	inDelta, err = alloraMath.InDelta(
		secondTotalReputerReward,
		calcSecondTotalReputerReward,
		alloraMath.MustNewDecFromString("0.0001"),
	)
	s.Require().NoError(err)
	s.Require().True(
		inDelta,
		"expected: %s, got: %s",
		secondTotalReputerReward.String(),
		calcSecondTotalReputerReward.String(),
	)

	// Check if the reward share increased
	s.Require().True(secondTotalReputerReward.Gt(firstTotalReputerReward))
}

func (s *RewardsTestSuite) TestRewardsIncreasesBalance() {
	block := int64(600)
	s.ctx = s.ctx.WithBlockHeight(block)
	epochLength := int64(10800)
	s.MintTokensToModule(types.AlloraStakingAccountName, cosmosMath.NewInt(10000000000))

	// Reputer Addresses
	reputerIndexes := s.returnIndexes(0, 5)
	// Worker Addresses
	workerIndexes := s.returnIndexes(5, 5)

	// Create topic
	newTopicMsg := &types.CreateNewTopicRequest{
		Creator:                  s.addrsStr[reputerIndexes[0]],
		Metadata:                 "test",
		LossMethod:               "mse",
		EpochLength:              epochLength,
		AllowNegative:            false,
		GroundTruthLag:           epochLength,
		WorkerSubmissionWindow:   10,
		AlphaRegret:              alloraMath.MustNewDecFromString("0.1"),
		PNorm:                    alloraMath.NewDecFromInt64(3),
		Epsilon:                  alloraMath.MustNewDecFromString("0.01"),
		MeritSortitionAlpha:      alloraMath.MustNewDecFromString("0.1"),
		ActiveInfererQuantile:    alloraMath.MustNewDecFromString("0.2"),
		ActiveForecasterQuantile: alloraMath.MustNewDecFromString("0.2"),
		ActiveReputerQuantile:    alloraMath.MustNewDecFromString("0.2"),
		EnableWorkerWhitelist:    true,
		EnableReputerWhitelist:   true,
	}
	res, err := s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	s.Require().NoError(err)

	// Get Topic Id
	topicId := res.TopicId

	// Register 5 workers
	for _, index := range workerIndexes {
		workerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: false,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, workerRegMsg)
		s.Require().NoError(err)
	}

	// Register 5 reputers
	for _, index := range reputerIndexes {
		reputerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: true,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, reputerRegMsg)
		s.Require().NoError(err)
	}

	cosmosOneE18 := inferencesynthesis.CosmosIntOneE18()

	// Add Stake for reputers
	var stakes = []cosmosMath.Int{
		cosmosMath.NewInt(1176644).Mul(cosmosOneE18),
		cosmosMath.NewInt(984623).Mul(cosmosOneE18),
		cosmosMath.NewInt(994676).Mul(cosmosOneE18),
		cosmosMath.NewInt(907999).Mul(cosmosOneE18),
		cosmosMath.NewInt(868582).Mul(cosmosOneE18),
	}
	for _, index := range reputerIndexes {
		s.MintTokensToAddress(s.addrs[index], stakes[index])
		_, err := s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
			Sender:  s.addrsStr[index],
			Amount:  stakes[index],
			TopicId: topicId,
		})
		s.Require().NoError(err)
	}

	initialStake := cosmosMath.NewInt(1000)
	s.MintTokensToAddress(s.addrs[reputerIndexes[0]], initialStake)
	fundTopicMessage := types.FundTopicRequest{
		Sender:  s.addrsStr[reputerIndexes[0]],
		TopicId: topicId,
		Amount:  initialStake,
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
	s.Require().NoError(err)

	// Insert unfullfiled nonces
	err = s.emissionsKeeper.AddWorkerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)

	reputerBalances := make([]sdk.Coin, 5)
	reputerStake := make([]cosmosMath.Int, 5)
	for _, index := range reputerIndexes {
		reputerBalances[index] = s.bankKeeper.GetBalance(s.ctx, s.addrs[index], params.DefaultBondDenom)
		reputerStake[index], err = s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId, s.addrsStr[index])
		s.Require().NoError(err)
	}

	workerBalances := make(map[int]sdk.Coin)
	for _, index := range workerIndexes {
		workerBalances[index] = s.bankKeeper.GetBalance(s.ctx, s.addrs[index], params.DefaultBondDenom)
	}

	// Insert inference from workers
	inferenceBundles := generateWorkerDataBundles(s, block, topicId)
	for _, payload := range inferenceBundles {
		_, err = s.msgServer.InsertWorkerPayload(s.ctx, &types.InsertWorkerPayloadRequest{
			Sender:           payload.Worker,
			WorkerDataBundle: payload,
		})
		s.Require().NoError(err)
	}

	topic, err := s.emissionsKeeper.GetTopic(s.ctx, topicId)
	s.Require().NoError(err)

	err = actorutils.CloseWorkerNonce(&s.emissionsKeeper, s.ctx, topic, *inferenceBundles[0].Nonce)
	s.Require().NoError(err)

	newBlockheight := block + topic.GroundTruthLag
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(newBlockheight)

	// Trigger end block - rewards distribution
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)

	// Insert loss bundle from reputers
	lossBundles := generateLossBundles(s, block, topicId, reputerIndexes)
	for _, payload := range lossBundles.ReputerValueBundles {
		_, err = s.msgServer.InsertReputerPayload(s.ctx, &types.InsertReputerPayloadRequest{
			Sender:             payload.ValueBundle.Reputer,
			ReputerValueBundle: payload,
		})
		s.Require().NoError(err)
	}

	err = actorutils.CloseReputerNonce(&s.emissionsKeeper, s.ctx, topic, *lossBundles.ReputerValueBundles[0].ValueBundle.ReputerRequestNonce.ReputerNonce)
	s.Require().NoError(err)

	// mint some rewards to give out
	s.MintTokensToModule(types.AlloraRewardsAccountName, cosmosMath.NewInt(1000))

	// Trigger end block - rewards distribution
	newBlockheight += epochLength
	s.ctx = s.ctx.WithBlockHeight(newBlockheight)
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)

	for i, index := range reputerIndexes {
		reputerStakeCurrent, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId, s.addrsStr[index])
		s.Require().NoError(err)
		s.Require().True(
			reputerStakeCurrent.GT(reputerStake[i]),
			"Reputer %s stake did not increase: %s | %s",
			s.addrsStr[index],
			reputerStakeCurrent.String(),
			reputerStake[i].String(),
		)
	}

	for _, index := range workerIndexes {
		s.Require().True(s.bankKeeper.GetBalance(s.ctx, s.addrs[index], params.DefaultBondDenom).Amount.GT(workerBalances[index].Amount))
	}
}

func (s *RewardsTestSuite) TestRewardsHandleStandardDeviationOfZero() {
	block := int64(600)
	s.ctx = s.ctx.WithBlockHeight(block)
	epochLength := int64(10800)

	// Reputer Addresses
	reputerIndexes := s.returnIndexes(0, 5)
	// Worker Addresses
	workerIndexes := s.returnIndexes(5, 5)

	// Create first topic
	newTopicMsg := &types.CreateNewTopicRequest{
		Creator:                  s.addrsStr[reputerIndexes[0]],
		Metadata:                 "test",
		LossMethod:               "mse",
		EpochLength:              epochLength,
		AllowNegative:            false,
		GroundTruthLag:           epochLength,
		WorkerSubmissionWindow:   10,
		AlphaRegret:              alloraMath.NewDecFromInt64(1),
		PNorm:                    alloraMath.NewDecFromInt64(3),
		Epsilon:                  alloraMath.MustNewDecFromString("0.01"),
		MeritSortitionAlpha:      alloraMath.MustNewDecFromString("0.1"),
		ActiveInfererQuantile:    alloraMath.MustNewDecFromString("0.2"),
		ActiveForecasterQuantile: alloraMath.MustNewDecFromString("0.2"),
		ActiveReputerQuantile:    alloraMath.MustNewDecFromString("0.2"),
		EnableWorkerWhitelist:    true,
		EnableReputerWhitelist:   true,
	}
	res, err := s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	s.Require().NoError(err)
	// Get Topic Id for first topic
	topicId1 := res.TopicId
	res, err = s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	s.Require().NoError(err)
	topicId2 := res.TopicId

	// Register 5 workers, first 3 for topic 1 and last 2 for topic 2
	for _, index := range workerIndexes {
		workerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId1,
			IsReputer: false,
			Owner:     s.addrsStr[index],
		}
		// Full registration of all workers in both topics.
		_, err := s.msgServer.Register(s.ctx, workerRegMsg)
		s.Require().NoError(err)

		workerRegMsg.TopicId = topicId2
		_, err = s.msgServer.Register(s.ctx, workerRegMsg)
		s.Require().NoError(err)
	}

	// Register 5 reputers, first 3 for topic 1 and last 2 for topic 2
	for i, index := range reputerIndexes {
		reputerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			Owner:     s.addrsStr[index],
			TopicId:   topicId1,
			IsReputer: true,
		}
		if i > 2 {
			reputerRegMsg.TopicId = topicId2
		}
		_, err := s.msgServer.Register(s.ctx, reputerRegMsg)
		s.Require().NoError(err)
	}

	cosmosOneE18 := inferencesynthesis.CosmosIntOneE18()

	// Add Stake for reputers
	var stakes = []cosmosMath.Int{
		cosmosMath.NewInt(1176644).Mul(cosmosOneE18),
		cosmosMath.NewInt(384623).Mul(cosmosOneE18),
		cosmosMath.NewInt(394676).Mul(cosmosOneE18),
		cosmosMath.NewInt(207999).Mul(cosmosOneE18),
		cosmosMath.NewInt(368582).Mul(cosmosOneE18),
	}
	for i, index := range reputerIndexes {
		addStakeMsg := &types.AddStakeRequest{
			Sender:  s.addrsStr[index],
			Amount:  stakes[i],
			TopicId: topicId1,
		}
		if i > 2 {
			addStakeMsg.TopicId = topicId2
		}
		s.MintTokensToAddress(s.addrs[index], stakes[i])
		_, err := s.msgServer.AddStake(s.ctx, addStakeMsg)
		s.Require().NoError(err)
	}

	// fund topic 1
	var initialStake int64 = 1000
	initialStakeCoins := sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, cosmosMath.NewInt(initialStake)))
	err = s.bankKeeper.MintCoins(s.ctx, types.AlloraStakingAccountName, initialStakeCoins)
	s.Require().NoError(err)
	err = s.bankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.AlloraStakingAccountName, s.addrs[reputerIndexes[0]], initialStakeCoins)
	s.Require().NoError(err)
	fundTopicMessage := types.FundTopicRequest{
		Sender:  s.addrsStr[reputerIndexes[0]],
		TopicId: topicId1,
		Amount:  cosmosMath.NewInt(initialStake),
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
	s.Require().NoError(err)

	// fund topic 2
	err = s.bankKeeper.MintCoins(s.ctx, types.AlloraStakingAccountName, initialStakeCoins)
	s.Require().NoError(err)
	err = s.bankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.AlloraStakingAccountName, s.addrs[reputerIndexes[0]], initialStakeCoins)
	s.Require().NoError(err)
	fundTopicMessage.TopicId = topicId2
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
	s.Require().NoError(err)

	// Insert unfullfiled nonces
	err = s.emissionsKeeper.AddWorkerNonce(s.ctx, topicId1, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId1, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddWorkerNonce(s.ctx, topicId2, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId2, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)

	reputerBalances := make([]sdk.Coin, 5)
	reputerStake := make([]cosmosMath.Int, 5)
	for i, index := range reputerIndexes {
		reputerBalances[i] = s.bankKeeper.GetBalance(s.ctx, s.addrs[index], params.DefaultBondDenom)
		if i > 2 {
			reputerStake[i], err = s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId2, s.addrsStr[index])
			s.Require().NoError(err)
		} else {
			reputerStake[i], err = s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId1, s.addrsStr[index])
			s.Require().NoError(err)
		}
	}

	workerBalances := make([]sdk.Coin, 5)
	for i, index := range workerIndexes {
		workerBalances[i] = s.bankKeeper.GetBalance(s.ctx, s.addrs[index], params.DefaultBondDenom)
	}

	// Insert inference from workers
	inferenceBundles := generateWorkerDataBundles(s, block, topicId1)
	for _, payload := range inferenceBundles {
		s.RegisterAllWorkersOfPayload(topicId1, payload)
		_, err = s.msgServer.InsertWorkerPayload(s.ctx, &types.InsertWorkerPayloadRequest{
			Sender:           payload.Worker,
			WorkerDataBundle: payload,
		})
		s.Require().NoError(err)
	}

	inferenceBundles2 := generateWorkerDataBundles(s, block, topicId2)
	for _, payload := range inferenceBundles2 {
		s.RegisterAllWorkersOfPayload(topicId2, payload)
		_, err = s.msgServer.InsertWorkerPayload(s.ctx, &types.InsertWorkerPayloadRequest{
			Sender:           payload.Worker,
			WorkerDataBundle: payload,
		})
		s.Require().NoError(err)
	}

	topic, err := s.emissionsKeeper.GetTopic(s.ctx, topicId1)
	s.Require().NoError(err)

	newBlockheight := block + topic.GroundTruthLag
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(newBlockheight)

	// Insert loss bundle from reputers
	lossBundles := generateLossBundles(s, block, topicId1, reputerIndexes)
	for i, payload := range lossBundles.ReputerValueBundles {
		s.RegisterAllReputersOfPayload(topicId1, payload)
		if i <= 2 {
			_, _ = s.emissionsKeeper.FulfillWorkerNonce(s.ctx, topicId1, payload.ValueBundle.ReputerRequestNonce.ReputerNonce)
			_ = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId1, payload.ValueBundle.ReputerRequestNonce.ReputerNonce)
			_, err = s.msgServer.InsertReputerPayload(s.ctx, &types.InsertReputerPayloadRequest{
				Sender:             payload.ValueBundle.Reputer,
				ReputerValueBundle: payload,
			})
			s.Require().NoError(err)
		}
	}

	topic2, err := s.emissionsKeeper.GetTopic(s.ctx, topicId2)
	s.Require().NoError(err)

	newBlockheight = block + topic2.GroundTruthLag
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(newBlockheight)

	lossBundles2 := generateLossBundles(s, block, topicId2, reputerIndexes)
	for i, payload := range lossBundles2.ReputerValueBundles {
		s.RegisterAllReputersOfPayload(topicId2, payload)
		if i > 2 {
			_, _ = s.emissionsKeeper.FulfillWorkerNonce(s.ctx, topicId2, payload.ValueBundle.ReputerRequestNonce.ReputerNonce)
			_ = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId2, payload.ValueBundle.ReputerRequestNonce.ReputerNonce)
			_, err = s.msgServer.InsertReputerPayload(s.ctx, &types.InsertReputerPayloadRequest{
				Sender:             payload.ValueBundle.Reputer,
				ReputerValueBundle: payload,
			})
			s.Require().NoError(err)
		}
	}

	block += epochLength * 3
	s.ctx = s.ctx.WithBlockHeight(block)

	// mint some rewards to give out
	s.MintTokensToModule(types.AlloraRewardsAccountName, cosmosMath.NewInt(10000000000))

	// Trigger end block - rewards distribution
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)
}

func (s *RewardsTestSuite) TestStandardRewardEmissionWithOneInfererAndOneReputer() {
	blockHeight := int64(600)
	s.ctx = s.ctx.WithBlockHeight(blockHeight)
	epochLength := int64(10800)

	// Reputer Addresses
	reputer := 0
	// Worker Addresses
	worker := 5

	// Create topic
	newTopicMsg := &types.CreateNewTopicRequest{
		Creator:                  s.addrsStr[reputer],
		Metadata:                 "test",
		LossMethod:               "mse",
		EpochLength:              epochLength,
		GroundTruthLag:           epochLength,
		AllowNegative:            false,
		WorkerSubmissionWindow:   10,
		AlphaRegret:              alloraMath.NewDecFromInt64(1),
		PNorm:                    alloraMath.NewDecFromInt64(3),
		Epsilon:                  alloraMath.MustNewDecFromString("0.01"),
		MeritSortitionAlpha:      alloraMath.MustNewDecFromString("0.1"),
		ActiveInfererQuantile:    alloraMath.MustNewDecFromString("0.2"),
		ActiveForecasterQuantile: alloraMath.MustNewDecFromString("0.2"),
		ActiveReputerQuantile:    alloraMath.MustNewDecFromString("0.2"),
		EnableWorkerWhitelist:    true,
		EnableReputerWhitelist:   true,
	}
	res, err := s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	s.Require().NoError(err)
	// Get Topic Id
	topicId := res.TopicId

	// Register 1 worker
	workerRegMsg := &types.RegisterRequest{
		Sender:    s.addrsStr[worker],
		TopicId:   topicId,
		IsReputer: false,
		Owner:     s.addrsStr[worker],
	}
	_, err = s.msgServer.Register(s.ctx, workerRegMsg)
	s.Require().NoError(err)

	// Register 1 reputer
	reputerRegMsg := &types.RegisterRequest{
		Sender:    s.addrsStr[reputer],
		TopicId:   topicId,
		IsReputer: true,
		Owner:     s.addrsStr[reputer],
	}
	_, err = s.msgServer.Register(s.ctx, reputerRegMsg)
	s.Require().NoError(err)

	cosmosOneE18 := inferencesynthesis.CosmosIntOneE18()

	s.MintTokensToAddress(s.addrs[reputer], cosmosMath.NewInt(1176644).Mul(cosmosOneE18))
	// Add Stake for reputer
	_, err = s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
		Sender:  s.addrsStr[reputer],
		Amount:  cosmosMath.NewInt(1176644).Mul(cosmosOneE18),
		TopicId: topicId,
	})
	s.Require().NoError(err)

	var initialStake int64 = 1000
	initialStakeCoins := sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, cosmosMath.NewInt(initialStake)))
	err = s.bankKeeper.MintCoins(s.ctx, types.AlloraStakingAccountName, initialStakeCoins)
	s.Require().NoError(err)
	err = s.bankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.AlloraStakingAccountName, s.addrs[reputer], initialStakeCoins)
	s.Require().NoError(err)
	fundTopicMessage := types.FundTopicRequest{
		Sender:  s.addrsStr[reputer],
		TopicId: topicId,
		Amount:  cosmosMath.NewInt(initialStake),
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
	s.Require().NoError(err)
	// Insert unfullfiled nonces
	err = s.emissionsKeeper.AddWorkerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: blockHeight,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: blockHeight,
	})
	s.Require().NoError(err)

	// Insert inference from worker
	worker1InferenceForecastBundle := &types.InferenceForecastBundle{
		Inference: &types.Inference{
			TopicId:     topicId,
			BlockHeight: blockHeight,
			Inferer:     s.addrsStr[worker],
			Value:       alloraMath.MustNewDecFromString("0.01127"),
			ExtraData:   []byte("extra data"),
			Proof:       "",
		},
		Forecast: nil,
	}
	worker1Sig, err := signInferenceForecastBundle(worker1InferenceForecastBundle, s.privKeys[worker])
	s.Require().NoError(err)
	worker1Bundle := &types.WorkerDataBundle{
		Worker:                             s.addrsStr[worker],
		TopicId:                            topicId,
		Nonce:                              &types.Nonce{BlockHeight: blockHeight},
		InferenceForecastsBundle:           worker1InferenceForecastBundle,
		InferencesForecastsBundleSignature: worker1Sig,
		Pubkey:                             s.pubKeyHexStr[worker],
	}
	_, err = s.msgServer.InsertWorkerPayload(s.ctx, &types.InsertWorkerPayloadRequest{
		Sender:           s.addrsStr[worker],
		WorkerDataBundle: worker1Bundle,
	})
	s.Require().NoError(err)

	// Insert loss bundle from reputer
	valueBundle := &types.ValueBundle{
		TopicId: topicId,
		ReputerRequestNonce: &types.ReputerRequestNonce{
			ReputerNonce: &types.Nonce{
				BlockHeight: blockHeight,
			},
		},
		ExtraData:                     nil,
		Reputer:                       s.addrsStr[reputer],
		CombinedValue:                 alloraMath.MustNewDecFromString("0.01127"),
		NaiveValue:                    alloraMath.MustNewDecFromString("0.0116"),
		InfererValues:                 []*types.WorkerAttributedValue{{Worker: s.addrsStr[worker], Value: alloraMath.MustNewDecFromString("0.0112")}},
		ForecasterValues:              []*types.WorkerAttributedValue{},
		OneOutInfererValues:           []*types.WithheldWorkerAttributedValue{},
		OneOutForecasterValues:        []*types.WithheldWorkerAttributedValue{},
		OneInForecasterValues:         []*types.WorkerAttributedValue{},
		OneOutInfererForecasterValues: nil,
	}
	sig, err := signValueBundle(valueBundle, s.privKeys[reputer])
	s.Require().NoError(err)
	reputerBundle := &types.ReputerValueBundle{
		Pubkey:      s.pubKeyHexStr[reputer],
		Signature:   sig,
		ValueBundle: valueBundle,
	}

	topic, err := s.emissionsKeeper.GetTopic(s.ctx, topicId)
	s.Require().NoError(err)

	newBlockheight := blockHeight + topic.GroundTruthLag
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(newBlockheight)

	_, _ = s.emissionsKeeper.FulfillWorkerNonce(s.ctx, topicId, reputerBundle.ValueBundle.ReputerRequestNonce.ReputerNonce)
	_ = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, reputerBundle.ValueBundle.ReputerRequestNonce.ReputerNonce)

	_, err = s.msgServer.InsertReputerPayload(s.ctx, &types.InsertReputerPayloadRequest{
		Sender:             s.addrsStr[reputer],
		ReputerValueBundle: reputerBundle,
	})
	s.Require().NoError(err)

	blockHeight += epochLength * 3
	s.ctx = s.ctx.WithBlockHeight(blockHeight)

	// mint some rewards to give out
	s.MintTokensToModule(types.AlloraRewardsAccountName, cosmosMath.NewInt(10000000000))

	// Trigger end block - rewards distribution
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)
}

func (s *RewardsTestSuite) SetParamsForTest() {
	// Setup a sender address
	adminPrivateKey := secp256k1.GenPrivKey()
	adminAddr := sdk.AccAddress(adminPrivateKey.PubKey().Address())
	err := s.emissionsKeeper.AddWhitelistAdmin(s.ctx, adminAddr.String())
	s.Require().NoError(err)

	newParams := &types.OptionalParams{
		MaxTopInferersToReward:  []uint64{24},
		MinEpochLength:          []int64{1},
		RegistrationFee:         []cosmosMath.Int{cosmosMath.NewInt(6)},
		MaxActiveTopicsPerBlock: []uint64{2},
		// the following fields are not set
		Version:                             nil,
		MaxSerializedMsgLength:              nil,
		MinTopicWeight:                      nil,
		RequiredMinimumStake:                nil,
		RemoveStakeDelayWindow:              nil,
		BetaEntropy:                         nil,
		LearningRate:                        nil,
		MaxGradientThreshold:                nil,
		MinStakeFraction:                    nil,
		MaxUnfulfilledWorkerRequests:        nil,
		MaxUnfulfilledReputerRequests:       nil,
		TopicRewardStakeImportance:          nil,
		TopicRewardFeeRevenueImportance:     nil,
		TopicRewardAlpha:                    nil,
		TaskRewardAlpha:                     nil,
		ValidatorsVsAlloraPercentReward:     nil,
		MaxSamplesToScaleScores:             nil,
		MaxTopForecastersToReward:           nil,
		MaxTopReputersToReward:              nil,
		CreateTopicFee:                      nil,
		GradientDescentMaxIters:             nil,
		DefaultPageLimit:                    nil,
		MaxPageLimit:                        nil,
		MinEpochLengthRecordLimit:           nil,
		BlocksPerMonth:                      nil,
		PRewardInference:                    nil,
		PRewardForecast:                     nil,
		PRewardReputer:                      nil,
		CRewardInference:                    nil,
		CRewardForecast:                     nil,
		CNorm:                               nil,
		EpsilonReputer:                      nil,
		HalfMaxProcessStakeRemovalsEndBlock: nil,
		DataSendingFee:                      nil,
		EpsilonSafeDiv:                      nil,
		MaxElementsPerForecast:              nil,
		MaxStringLength:                     nil,
		InitialRegretQuantile:               nil,
		PNormSafeDiv:                        nil,
		GlobalWhitelistEnabled:              nil,
		TopicCreatorWhitelistEnabled:        nil,
		MinExperiencedWorkerRegrets:         nil,
	}

	updateMsg := &types.UpdateParamsRequest{
		Sender: adminAddr.String(),
		Params: newParams,
	}

	response, err := s.msgServer.UpdateParams(s.ctx, updateMsg)
	s.Require().NoError(err)
	s.Require().NotNil(response)
}

func (s *RewardsTestSuite) TestOnlyFewTopActorsGetReward() {
	block := int64(600)
	s.ctx = s.ctx.WithBlockHeight(block)
	epochLength := int64(10800)

	// Reputer Addresses
	var reputerIndexes = make([]int, 0)
	var workerIndexes = make([]int, 0)
	var stakes = make([]cosmosMath.Int, 0)
	cosmosOneE18 := inferencesynthesis.CosmosIntOneE18()

	s.SetParamsForTest()

	for i := 0; i < 25; i++ {
		reputerIndexes = append(reputerIndexes, i)
		workerIndexes = append(workerIndexes, i+25)
		stakes = append(stakes, cosmosMath.NewInt(int64(1000*(i+1))).Mul(cosmosOneE18))
	}

	// Create topic
	newTopicMsg := &types.CreateNewTopicRequest{
		Creator:                  s.addrsStr[reputerIndexes[0]],
		Metadata:                 "test",
		LossMethod:               "mse",
		AllowNegative:            false,
		EpochLength:              epochLength,
		GroundTruthLag:           epochLength,
		WorkerSubmissionWindow:   10,
		AlphaRegret:              alloraMath.NewDecFromInt64(1),
		PNorm:                    alloraMath.NewDecFromInt64(3),
		Epsilon:                  alloraMath.MustNewDecFromString("0.01"),
		MeritSortitionAlpha:      alloraMath.MustNewDecFromString("0.1"),
		ActiveInfererQuantile:    alloraMath.MustNewDecFromString("0.2"),
		ActiveForecasterQuantile: alloraMath.MustNewDecFromString("0.2"),
		ActiveReputerQuantile:    alloraMath.MustNewDecFromString("0.2"),
		EnableWorkerWhitelist:    true,
		EnableReputerWhitelist:   true,
	}
	res, err := s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	s.Require().NoError(err)

	// Get Topic Id
	topicId := res.TopicId

	// Register 25 workers
	for _, index := range workerIndexes {
		workerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: false,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, workerRegMsg)
		s.Require().NoError(err)
	}

	// Register 25 reputers
	for _, index := range reputerIndexes {
		reputerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: true,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, reputerRegMsg)
		s.Require().NoError(err)
	}

	for i, index := range reputerIndexes {
		s.MintTokensToAddress(s.addrs[index], stakes[i])
		_, err := s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
			Sender:  s.addrsStr[index],
			Amount:  stakes[i],
			TopicId: topicId,
		})
		s.Require().NoError(err)
	}

	var initialStake int64 = 1000
	s.FundAccount(initialStake, s.addrs[reputerIndexes[0]])

	fundTopicMessage := types.FundTopicRequest{
		Sender:  s.addrsStr[reputerIndexes[0]],
		TopicId: topicId,
		Amount:  cosmosMath.NewInt(initialStake),
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
	s.Require().NoError(err)

	// Insert unfullfiled nonces
	err = s.emissionsKeeper.AddWorkerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)

	// Insert inference from workers
	inferenceBundles := generateHugeWorkerDataBundles(s, block, topicId, workerIndexes)
	for _, payload := range inferenceBundles {
		_, err = s.msgServer.InsertWorkerPayload(s.ctx, &types.InsertWorkerPayloadRequest{
			Sender:           payload.Worker,
			WorkerDataBundle: payload,
		})
		s.Require().NoError(err)
	}

	topic, err := s.emissionsKeeper.GetTopic(s.ctx, topicId)
	s.Require().NoError(err)

	err = actorutils.CloseWorkerNonce(&s.emissionsKeeper, s.ctx, topic, *inferenceBundles[0].Nonce)
	s.Require().NoError(err)

	newBlockheight := block + topic.GroundTruthLag
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(newBlockheight)

	// Insert loss bundle from reputers
	lossBundles := generateHugeLossBundles(s, block, topicId, reputerIndexes, workerIndexes)
	for _, payload := range lossBundles.ReputerValueBundles {
		_, err = s.msgServer.InsertReputerPayload(s.ctx, &types.InsertReputerPayloadRequest{
			Sender:             s.addrsStr[reputerIndexes[0]],
			ReputerValueBundle: payload,
		})
		s.Require().NoError(err)
	}
	err = actorutils.CloseReputerNonce(&s.emissionsKeeper, s.ctx, topic, *lossBundles.ReputerValueBundles[0].ValueBundle.ReputerRequestNonce.ReputerNonce)
	s.Require().NoError(err)

	params, err := s.emissionsKeeper.GetParams(s.ctx)
	s.Require().NoError(err)

	//scoresAtBlock, err := s.emissionsKeeper.GetReputersScoresAtBlock(s.ctx, topicId, block)
	//s.Require().Equal(len(scoresAtBlock.Scores), int(params.GetMaxTopReputersToReward()), "Only few Top reputers can get reward")

	networkLossBundles, err := s.emissionsKeeper.GetNetworkLossBundleAtBlock(s.ctx, topicId, block)
	s.Require().NoError(err)
	s.Require().NotNil(networkLossBundles)

	infererScores, err := rewards.GenerateInferenceScores(
		s.ctx,
		s.emissionsKeeper,
		topicId,
		block,
		*networkLossBundles)
	s.Require().NoError(err)

	forecasterScores, err := rewards.GenerateForecastScores(
		s.ctx,
		s.emissionsKeeper,
		topicId,
		block,
		*networkLossBundles)
	s.Require().NoError(err)

	s.Require().Equal(uint64(len(infererScores)), params.GetMaxTopInferersToReward(), "Only few Top inferers can get reward")
	s.Require().Equal(uint64(len(forecasterScores)), params.GetMaxTopForecastersToReward(), "Only few Top forecasters can get reward")
}

func (s *RewardsTestSuite) TestTotalInferersRewardFractionGrowsWithMoreInferers() {
	block := int64(100)
	s.ctx = s.ctx.WithBlockHeight(block)

	reputerIndexes := s.returnIndexes(0, 3)
	workerIndexes := s.returnIndexes(5, 5)

	cosmosOneE18 := inferencesynthesis.CosmosIntOneE18()

	stakes := []cosmosMath.Int{
		cosmosMath.NewInt(1000000000000000000).Mul(cosmosOneE18),
		cosmosMath.NewInt(1000000000000000000).Mul(cosmosOneE18),
		cosmosMath.NewInt(1000000000000000000).Mul(cosmosOneE18),
	}

	// Create topic
	newTopicMsg := &types.CreateNewTopicRequest{
		Creator:                  s.addrsStr[reputerIndexes[0]],
		Metadata:                 "test",
		LossMethod:               "mse",
		AllowNegative:            false,
		EpochLength:              10800,
		GroundTruthLag:           10800,
		WorkerSubmissionWindow:   10,
		AlphaRegret:              alloraMath.NewDecFromInt64(1),
		PNorm:                    alloraMath.NewDecFromInt64(3),
		Epsilon:                  alloraMath.MustNewDecFromString("0.01"),
		MeritSortitionAlpha:      alloraMath.MustNewDecFromString("0.1"),
		ActiveInfererQuantile:    alloraMath.MustNewDecFromString("0.2"),
		ActiveForecasterQuantile: alloraMath.MustNewDecFromString("0.2"),
		ActiveReputerQuantile:    alloraMath.MustNewDecFromString("0.2"),
		EnableWorkerWhitelist:    true,
		EnableReputerWhitelist:   true,
	}
	res, err := s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	s.Require().NoError(err)

	// Get Topic Id
	topicId := res.TopicId

	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)

	// Register 5 workers
	for _, index := range workerIndexes {
		workerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: false,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, workerRegMsg)
		s.Require().NoError(err)
	}

	// Register 3 reputers
	for _, index := range reputerIndexes {
		reputerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: true,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, reputerRegMsg)
		s.Require().NoError(err)
	}
	// Add Stake for reputers
	for i, index := range reputerIndexes {
		s.MintTokensToAddress(s.addrs[index], stakes[i])
		_, err := s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
			Sender:  s.addrsStr[index],
			Amount:  stakes[i],
			TopicId: topicId,
		})
		s.Require().NoError(err)
	}

	var initialStake int64 = 1000
	s.FundAccount(initialStake, s.addrs[reputerIndexes[0]])
	fundTopicMessage := types.FundTopicRequest{
		Sender:  s.addrsStr[reputerIndexes[0]],
		TopicId: topicId,
		Amount:  cosmosMath.NewInt(initialStake),
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
	s.Require().NoError(err)

	err = s.emissionsKeeper.AddWorkerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)

	// Insert inference from workers
	inferenceBundles := generateHugeWorkerDataBundles(s, block, topicId, workerIndexes)
	for _, payload := range inferenceBundles {
		s.RegisterAllWorkersOfPayload(topicId, payload)
		_, err = s.msgServer.InsertWorkerPayload(s.ctx, &types.InsertWorkerPayloadRequest{
			Sender:           payload.Worker,
			WorkerDataBundle: payload,
		})
		s.Require().NoError(err)
	}

	topic, err := s.emissionsKeeper.GetTopic(s.ctx, topicId)
	s.Require().NoError(err)

	err = actorutils.CloseWorkerNonce(&s.emissionsKeeper, s.ctx, topic, *inferenceBundles[0].Nonce)
	s.Require().NoError(err)

	// Insert loss bundle from reputers
	lossBundles := generateHugeLossBundles(s, block, topicId, reputerIndexes, workerIndexes)

	block = block + topic.GroundTruthLag
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(block)

	// Trigger end block - rewards distribution
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)

	for _, payload := range lossBundles.ReputerValueBundles {
		_, err = s.msgServer.InsertReputerPayload(s.ctx, &types.InsertReputerPayloadRequest{
			Sender:             payload.ValueBundle.Reputer,
			ReputerValueBundle: payload,
		})
		s.Require().NoError(err)
	}

	err = actorutils.CloseReputerNonce(&s.emissionsKeeper, s.ctx, topic,
		*lossBundles.ReputerValueBundles[0].ValueBundle.ReputerRequestNonce.ReputerNonce)
	s.Require().NoError(err)

	topicTotalRewards := alloraMath.NewDecFromInt64(1000000)
	params, err := s.emissionsKeeper.GetParams(s.ctx)
	s.Require().NoError(err)

	firstRewardsDistribution, _, err := rewards.GenerateRewardsDistributionByTopicParticipant(
		rewards.GenerateRewardsDistributionByTopicParticipantArgs{
			Ctx:          s.ctx,
			K:            s.emissionsKeeper,
			TopicId:      topicId,
			TopicReward:  &topicTotalRewards,
			BlockHeight:  lossBundles.ReputerValueBundles[0].ValueBundle.ReputerRequestNonce.ReputerNonce.BlockHeight,
			ModuleParams: params,
		})
	s.Require().NoError(err)

	totalInferersReward := alloraMath.ZeroDec()
	totalForecastersReward := alloraMath.ZeroDec()
	totalReputersReward := alloraMath.ZeroDec()
	countInferencesAccepted := 0
	countForecastersAccepted := 0
	countReputersAccepted := 0
	for _, reward := range firstRewardsDistribution {
		if reward.Type == types.WorkerInferenceRewardType {
			totalInferersReward, err = totalInferersReward.Add(reward.Reward)
			s.Require().NoError(err)
			countInferencesAccepted++
		} else if reward.Type == types.WorkerForecastRewardType {
			totalForecastersReward, err = totalForecastersReward.Add(reward.Reward)
			s.Require().NoError(err)
			countForecastersAccepted++
		} else if reward.Type == types.ReputerAndDelegatorRewardType {
			totalReputersReward, err = totalReputersReward.Add(reward.Reward)
			s.Require().NoError(err)
			countReputersAccepted++
		}
	}
	s.Require().Equal(len(inferenceBundles), countInferencesAccepted)
	// because the GenerateHugeWorkerDataBundles does not generate forecasts
	// that are accepted, the forecast elements are for inferers that don't exist
	s.Require().Zero(countForecastersAccepted)
	s.Require().Equal(len(lossBundles.ReputerValueBundles), countReputersAccepted)
	totalNonInfererReward, err := totalForecastersReward.Add(totalReputersReward)
	s.Require().NoError(err)
	totalReward, err := totalNonInfererReward.Add(totalInferersReward)
	s.Require().NoError(err)

	firstInfererFraction, err := totalInferersReward.Quo(totalReward)
	s.Require().NoError(err)

	block += 1
	s.ctx = s.ctx.WithBlockHeight(block)

	// Add new worker(inferer) and stakes
	newSecondWorkersIndexes := []int{
		10,
		11,
		12,
	}
	newSecondWorkersIndexes = append(workerIndexes, newSecondWorkersIndexes...)

	// Create new topic
	newTopicMsg = &types.CreateNewTopicRequest{
		Creator:                  s.addrsStr[reputerIndexes[0]],
		Metadata:                 "test",
		LossMethod:               "mse",
		AllowNegative:            false,
		EpochLength:              10800,
		GroundTruthLag:           10800,
		WorkerSubmissionWindow:   10,
		AlphaRegret:              alloraMath.NewDecFromInt64(1),
		PNorm:                    alloraMath.NewDecFromInt64(3),
		Epsilon:                  alloraMath.MustNewDecFromString("0.01"),
		MeritSortitionAlpha:      alloraMath.MustNewDecFromString("0.1"),
		ActiveInfererQuantile:    alloraMath.MustNewDecFromString("0.2"),
		ActiveForecasterQuantile: alloraMath.MustNewDecFromString("0.2"),
		ActiveReputerQuantile:    alloraMath.MustNewDecFromString("0.2"),
		EnableWorkerWhitelist:    true,
		EnableReputerWhitelist:   true,
	}
	res, err = s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	s.Require().NoError(err)

	// Get Topic Id
	topicId = res.TopicId

	// Register 7 workers with 2 new inferers
	for _, index := range newSecondWorkersIndexes {
		workerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: false,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, workerRegMsg)
		s.Require().NoError(err)
	}

	// Register 3 reputers
	for _, index := range reputerIndexes {
		reputerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: true,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, reputerRegMsg)
		s.Require().NoError(err)
	}
	// Add Stake for reputers
	for i, index := range reputerIndexes {
		s.MintTokensToAddress(s.addrs[index], stakes[i])
		_, err := s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
			Sender:  s.addrsStr[index],
			Amount:  stakes[i],
			TopicId: topicId,
		})
		s.Require().NoError(err)
	}

	s.FundAccount(initialStake, s.addrs[reputerIndexes[0]])

	fundTopicMessage = types.FundTopicRequest{
		Sender:  s.addrsStr[reputerIndexes[0]],
		TopicId: topicId,
		Amount:  cosmosMath.NewInt(initialStake),
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
	s.Require().NoError(err)

	block += newTopicMsg.EpochLength
	s.ctx = s.ctx.WithBlockHeight(block)
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddWorkerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)

	block += 1
	s.ctx = s.ctx.WithBlockHeight(block)

	err = s.emissionsKeeper.AddWorkerNonce(
		s.ctx,
		topicId,
		&types.Nonce{BlockHeight: block},
	)
	s.Require().NoError(err)
	// Insert inference from workers
	inferenceBundles = generateHugeWorkerDataBundles(s, block, topicId, newSecondWorkersIndexes)
	// Add more inferer
	newInferenceBundles := generateMoreInferencesDataBundles(s, block, topicId)
	inferenceBundles = append(inferenceBundles, newInferenceBundles...)

	for _, payload := range inferenceBundles {
		s.RegisterAllWorkersOfPayload(topicId, payload)
		_, err = s.msgServer.InsertWorkerPayload(s.ctx, &types.InsertWorkerPayloadRequest{
			Sender:           payload.Worker,
			WorkerDataBundle: payload,
		})
		s.Require().NoError(err)
	}
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)

	topic, err = s.emissionsKeeper.GetTopic(s.ctx, topicId)
	s.Require().NoError(err)

	err = actorutils.CloseWorkerNonce(&s.emissionsKeeper, s.ctx, topic, *inferenceBundles[0].Nonce)
	s.Require().NoError(err)

	lossBundles = generateHugeLossBundles(s, block, topicId, reputerIndexes, newSecondWorkersIndexes)

	block = block + topic.GroundTruthLag
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(block)

	// Insert loss bundle from reputers
	for _, payload := range lossBundles.ReputerValueBundles {
		_, err = s.msgServer.InsertReputerPayload(s.ctx, &types.InsertReputerPayloadRequest{
			Sender:             payload.ValueBundle.Reputer,
			ReputerValueBundle: payload,
		})
		s.Require().NoError(err)
	}
	err = actorutils.CloseReputerNonce(&s.emissionsKeeper, s.ctx, topic,
		*lossBundles.ReputerValueBundles[0].ValueBundle.ReputerRequestNonce.ReputerNonce)
	s.Require().NoError(err)

	topicTotalRewards = alloraMath.NewDecFromInt64(1000000)
	secondRewardsDistribution, _, err := rewards.GenerateRewardsDistributionByTopicParticipant(
		rewards.GenerateRewardsDistributionByTopicParticipantArgs{
			Ctx:          s.ctx,
			K:            s.emissionsKeeper,
			TopicId:      topicId,
			TopicReward:  &topicTotalRewards,
			BlockHeight:  lossBundles.ReputerValueBundles[0].ValueBundle.ReputerRequestNonce.ReputerNonce.BlockHeight,
			ModuleParams: params,
		})
	s.Require().NoError(err)

	totalInferersReward = alloraMath.ZeroDec()
	totalReward = alloraMath.ZeroDec()
	for _, reward := range secondRewardsDistribution {
		if reward.Type == types.WorkerInferenceRewardType {
			totalInferersReward, err = totalInferersReward.Add(reward.Reward)
			s.Require().NoError(err)
		}
		totalReward, err = totalReward.Add(reward.Reward)
		s.Require().NoError(err)
	}
	secondInfererFraction, err := totalInferersReward.Quo(totalReward)
	s.Require().NoError(err)
	s.Require().True(
		firstInfererFraction.Lt(secondInfererFraction),
		"Second inference fraction must be bigger than first fraction %s < %s",
		firstInfererFraction,
		secondInfererFraction,
	)

	// Add new worker(forecaster) and stakes
	newThirdWorkersIndexes := []int{
		10,
		11,
	}
	newThirdWorkersIndexes = append(workerIndexes, newThirdWorkersIndexes...)

	// Create new topic
	newTopicMsg = &types.CreateNewTopicRequest{
		Creator:                  s.addrsStr[reputerIndexes[0]],
		Metadata:                 "test",
		LossMethod:               "mse",
		AllowNegative:            false,
		EpochLength:              10800,
		GroundTruthLag:           10800,
		WorkerSubmissionWindow:   10,
		AlphaRegret:              alloraMath.NewDecFromInt64(1),
		PNorm:                    alloraMath.NewDecFromInt64(3),
		Epsilon:                  alloraMath.MustNewDecFromString("0.01"),
		MeritSortitionAlpha:      alloraMath.MustNewDecFromString("0.1"),
		ActiveInfererQuantile:    alloraMath.MustNewDecFromString("0.2"),
		ActiveForecasterQuantile: alloraMath.MustNewDecFromString("0.2"),
		ActiveReputerQuantile:    alloraMath.MustNewDecFromString("0.2"),
		EnableWorkerWhitelist:    true,
		EnableReputerWhitelist:   true,
	}
	res, err = s.msgServer.CreateNewTopic(s.ctx, newTopicMsg)
	s.Require().NoError(err)

	// Get Topic Id
	topicId = res.TopicId

	// Register 7 workers with 2 new forecasters
	for _, index := range newThirdWorkersIndexes {
		workerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: false,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, workerRegMsg)
		s.Require().NoError(err)
	}

	// Register 3 reputers
	for _, index := range reputerIndexes {
		reputerRegMsg := &types.RegisterRequest{
			Sender:    s.addrsStr[index],
			TopicId:   topicId,
			IsReputer: true,
			Owner:     s.addrsStr[index],
		}
		_, err := s.msgServer.Register(s.ctx, reputerRegMsg)
		s.Require().NoError(err)
	}
	// Add Stake for reputers
	for i, index := range reputerIndexes {
		s.MintTokensToAddress(s.addrs[index], stakes[i])
		_, err := s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
			Sender:  s.addrsStr[index],
			Amount:  stakes[i],
			TopicId: topicId,
		})
		s.Require().NoError(err)
	}

	s.FundAccount(initialStake, s.addrs[reputerIndexes[0]])

	fundTopicMessage = types.FundTopicRequest{
		Sender:  s.addrsStr[reputerIndexes[0]],
		TopicId: topicId,
		Amount:  cosmosMath.NewInt(initialStake),
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
	s.Require().NoError(err)

	block += newTopicMsg.EpochLength
	s.ctx = s.ctx.WithBlockHeight(block)
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)

	block++
	s.ctx = s.ctx.WithBlockHeight(block)

	err = s.emissionsKeeper.AddWorkerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)
	err = s.emissionsKeeper.AddReputerNonce(s.ctx, topicId, &types.Nonce{
		BlockHeight: block,
	})
	s.Require().NoError(err)

	// Insert inference from workers
	inferenceBundles = generateHugeWorkerDataBundles(s, block, topicId, newThirdWorkersIndexes)
	// Add more inferer
	newInferenceBundles = generateMoreForecastersDataBundles(s, block, topicId)
	inferenceBundles = append(inferenceBundles, newInferenceBundles...)
	for _, payload := range inferenceBundles {
		s.RegisterAllWorkersOfPayload(topicId, payload)
		_, err = s.msgServer.InsertWorkerPayload(s.ctx, &types.InsertWorkerPayloadRequest{
			Sender:           payload.Worker,
			WorkerDataBundle: payload,
		})
		s.Require().NoError(err)
	}

	topic, err = s.emissionsKeeper.GetTopic(s.ctx, topicId)
	s.Require().NoError(err)

	err = actorutils.CloseWorkerNonce(&s.emissionsKeeper, s.ctx, topic, *inferenceBundles[0].Nonce)
	s.Require().NoError(err)

	lossBundles = generateHugeLossBundles(s, block, topicId, reputerIndexes, newThirdWorkersIndexes)
	block = block + topic.GroundTruthLag
	s.ctx = sdk.UnwrapSDKContext(s.ctx).WithBlockHeight(block)

	// Insert loss bundle from reputers
	for _, payload := range lossBundles.ReputerValueBundles {
		_, err = s.msgServer.InsertReputerPayload(s.ctx, &types.InsertReputerPayloadRequest{
			Sender:             payload.ValueBundle.Reputer,
			ReputerValueBundle: payload,
		})
		s.Require().NoError(err)
	}
	err = actorutils.CloseReputerNonce(&s.emissionsKeeper, s.ctx, topic,
		*lossBundles.ReputerValueBundles[0].ValueBundle.ReputerRequestNonce.ReputerNonce)
	s.Require().NoError(err)

	topicTotalRewards = alloraMath.NewDecFromInt64(1000000)
	thirdRewardsDistribution, _, err := rewards.GenerateRewardsDistributionByTopicParticipant(
		rewards.GenerateRewardsDistributionByTopicParticipantArgs{
			Ctx:          s.ctx,
			K:            s.emissionsKeeper,
			TopicId:      topicId,
			TopicReward:  &topicTotalRewards,
			BlockHeight:  lossBundles.ReputerValueBundles[0].ValueBundle.ReputerRequestNonce.ReputerNonce.BlockHeight,
			ModuleParams: params,
		})
	s.Require().NoError(err)

	totalInferersReward = alloraMath.ZeroDec()
	totalReward = alloraMath.ZeroDec()
	for _, reward := range thirdRewardsDistribution {
		if reward.Type == types.WorkerInferenceRewardType {
			totalInferersReward, err = totalInferersReward.Add(reward.Reward)
			s.Require().NoError(err)
		}
		totalReward, err = totalReward.Add(reward.Reward)
		s.Require().NoError(err)
	}
	thirdInfererFraction, err := totalInferersReward.Quo(totalReward)
	s.Require().NoError(err)
	s.Require().True(
		firstInfererFraction.Lt(thirdInfererFraction),
		"Third forecaster fraction must be bigger than first fraction %s > %s",
		firstInfererFraction.String(),
		thirdInfererFraction.String(),
	)
}

// TestRewardForTopicGoesUpWhenRelativeStakeGoesUp tests that the reward for a topic increases
// when its relative stake compared to other topics increases.
//
// Setup:
// - Create two topics (topicId0 and topicId1) with identical initial stakes, workers, and reputers
// - Set up identical worker and reputer values for both topics
// - Record initial stakes for reputers on both topics
//
// Expected outcomes:
// 1. Initially, rewards for both topics should be similar due to identical setups
// 2. After increasing stake on one topic:
//   - The reward for the topic with increased stake should be higher
//   - The reward for the topic with unchanged stake should be lower
//
// 3. The total rewards across both topics should remain constant
//
// This test demonstrates that the reward distribution mechanism correctly
// adjusts rewards based on the relative stakes of topics in the network.
func (s *RewardsTestSuite) TestRewardForTopicGoesUpWhenRelativeStakeGoesUp() {
	// setup
	require := s.Require()

	alphaRegret := alloraMath.MustNewDecFromString("0.1")

	block := int64(1)
	s.ctx = s.ctx.WithBlockHeight(block)

	s.SetParamsForTest()
	reputerIndexes := s.returnIndexes(0, 3)
	workerIndexes := s.returnIndexes(3, 3)

	// setup topics
	stake := cosmosMath.NewInt(1000).Mul(inferencesynthesis.CosmosIntOneE18())

	epochLength := int64(100)
	topicId0 := s.setUpTopicWithEpochLength(block, workerIndexes, reputerIndexes, stake, alphaRegret, epochLength)
	topicId1 := s.setUpTopicWithEpochLength(block, workerIndexes, reputerIndexes, stake, alphaRegret, epochLength)
	// setup values to be identical for both topics
	reputerValues := []TestWorkerValue{
		{Index: reputerIndexes[0], Value: "0.2"},
		{Index: reputerIndexes[1], Value: "0.2"},
		{Index: reputerIndexes[2], Value: "0.2"},
	}

	workerValues := []TestWorkerValue{
		{Index: workerIndexes[0], Value: "0.2"},
		{Index: workerIndexes[1], Value: "0.2"},
		{Index: workerIndexes[2], Value: "0.2"},
	}

	// record the stakes on each topic so we can see the reward differences
	reputer0_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrsStr[reputerIndexes[0]])
	require.NoError(err)
	reputer1_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrsStr[reputerIndexes[1]])
	require.NoError(err)
	reputer2_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrsStr[reputerIndexes[2]])
	require.NoError(err)

	reputer3_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId1, s.addrsStr[reputerIndexes[0]])
	require.NoError(err)
	reputer4_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId1, s.addrsStr[reputerIndexes[1]])
	require.NoError(err)
	reputer5_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId1, s.addrsStr[reputerIndexes[2]])
	require.NoError(err)

	// do work on the topics to earn rewards
	s.getRewardsDistribution(
		topicId0,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.1",
	)

	s.getRewardsDistribution(
		topicId1,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.1",
	)

	// force rewards to be distributed
	s.MintTokensToModule(types.AlloraRewardsAccountName, cosmosMath.NewInt(1000))

	nextBlock, _, err := s.emissionsKeeper.GetNextPossibleChurningBlockByTopicId(s.ctx, topicId0)

	s.T().Logf("Next block: %v", nextBlock)
	s.Require().NoError(err)
	s.ctx = s.ctx.WithBlockHeight(nextBlock)

	// Check that the total sum of previous topic weights is equal to the sum of the weights of the two topics
	topic1Weight0, _, err := s.emissionsKeeper.GetPreviousTopicWeight(s.ctx, topicId0)
	s.Require().NoError(err)
	topic0Weight1, _, err := s.emissionsKeeper.GetPreviousTopicWeight(s.ctx, topicId1)
	s.Require().NoError(err)
	totalSumPreviousTopicWeights, err := s.emissionsKeeper.GetTotalSumPreviousTopicWeights(s.ctx)
	s.Require().NoError(err)
	sumWeights, err := topic0Weight1.Add(topic1Weight0)
	s.Require().NoError(err)
	inDelta, err := alloraMath.InDelta(totalSumPreviousTopicWeights, sumWeights, alloraMath.MustNewDecFromString("0.0001"))
	s.Require().NoError(err)
	s.Require().True(inDelta, fmt.Sprintf("Total sum of previous topic weights %s + %s = %s is not equal to the sum of the weights of the two topics %s", topic0Weight1, topic1Weight0, totalSumPreviousTopicWeights, sumWeights))

	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)

	require.NoError(err)

	worker1InclusionNum, err := s.emissionsKeeper.GetCountInfererInclusionsInTopic(s.ctx, topicId0, s.addrsStr[workerIndexes[0]])
	l.Println("worker1InclusionNum", worker1InclusionNum)
	require.NoError(err)
	require.Equal(uint64(1), worker1InclusionNum)
	worker2InclusionNum, err := s.emissionsKeeper.GetCountInfererInclusionsInTopic(s.ctx, topicId0, s.addrsStr[workerIndexes[1]])
	l.Println("worker2InclusionNum", worker2InclusionNum)
	require.NoError(err)
	require.Equal(uint64(1), worker2InclusionNum)
	worker3InclusionNum, err := s.emissionsKeeper.GetCountInfererInclusionsInTopic(s.ctx, topicId0, s.addrsStr[workerIndexes[2]])
	l.Println("worker3InclusionNum", worker3InclusionNum)
	require.Equal(uint64(1), worker3InclusionNum)
	require.NoError(err)

	worker1InclusionNum, err = s.emissionsKeeper.GetCountForecasterInclusionsInTopic(s.ctx, topicId0, s.addrsStr[workerIndexes[0]])
	l.Println("worker1InclusionNum", worker1InclusionNum)
	require.NoError(err)
	require.Equal(uint64(1), worker1InclusionNum)
	worker2InclusionNum, err = s.emissionsKeeper.GetCountForecasterInclusionsInTopic(s.ctx, topicId0, s.addrsStr[workerIndexes[1]])
	l.Println("worker2InclusionNum", worker2InclusionNum)
	require.NoError(err)
	require.Equal(uint64(1), worker2InclusionNum)
	worker3InclusionNum, err = s.emissionsKeeper.GetCountForecasterInclusionsInTopic(s.ctx, topicId0, s.addrsStr[workerIndexes[2]])
	l.Println("worker3InclusionNum", worker3InclusionNum)
	require.Equal(uint64(1), worker3InclusionNum)
	require.NoError(err)
	const topicFundAmount int64 = 1000

	fundTopic := func(topicId uint64, funderAddr sdk.AccAddress, amount int64) {
		s.MintTokensToAddress(funderAddr, cosmosMath.NewInt(amount))
		fundTopicMessage := types.FundTopicRequest{
			Sender:  funderAddr.String(),
			TopicId: topicId,
			Amount:  cosmosMath.NewInt(amount),
		}
		_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
		require.NoError(err)
	}

	fundTopic(topicId0, s.addrs[0], topicFundAmount)
	fundTopic(topicId1, s.addrs[3], topicFundAmount)

	reputer0_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrsStr[reputerIndexes[0]])
	require.NoError(err)
	reputer1_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrsStr[reputerIndexes[1]])
	require.NoError(err)
	reputer2_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrsStr[reputerIndexes[2]])
	require.NoError(err)

	reputer3_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId1, s.addrsStr[reputerIndexes[0]])
	require.NoError(err)
	reputer4_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId1, s.addrsStr[reputerIndexes[1]])
	require.NoError(err)
	reputer5_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId1, s.addrsStr[reputerIndexes[2]])
	require.NoError(err)

	reputer0_Reward0 := reputer0_Stake1.Sub(reputer0_Stake0)
	reputer1_Reward0 := reputer1_Stake1.Sub(reputer1_Stake0)
	reputer2_Reward0 := reputer2_Stake1.Sub(reputer2_Stake0)
	reputer3_Reward0 := reputer3_Stake1.Sub(reputer3_Stake0)
	reputer4_Reward0 := reputer4_Stake1.Sub(reputer4_Stake0)
	reputer5_Reward0 := reputer5_Stake1.Sub(reputer5_Stake0)

	topic0RewardTotal0 := reputer0_Reward0.Add(reputer1_Reward0).Add(reputer2_Reward0)
	topic1RewardTotal0 := reputer3_Reward0.Add(reputer4_Reward0).Add(reputer5_Reward0)

	require.Equal(topic0RewardTotal0, topic1RewardTotal0)

	// Now, in second trial, increase stake for first reputer in topic1
	s.MintTokensToAddress(s.addrs[reputerIndexes[0]], stake)
	_, err = s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
		Sender:  s.addrsStr[reputerIndexes[0]],
		Amount:  stake,
		TopicId: topicId1,
	})
	require.NoError(err)

	// record the updated stakes
	reputer3_Stake1, err = s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId1, s.addrsStr[reputerIndexes[0]])
	require.NoError(err)

	// force rewards to be distributed
	block = s.ctx.BlockHeight()
	block++
	s.ctx = s.ctx.WithBlockHeight(block)

	// do work on the topics to earn rewards
	s.getRewardsDistribution(
		topicId0,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.1",
	)

	s.getRewardsDistribution(
		topicId1,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.1",
	)

	s.MintTokensToModule(types.AlloraRewardsAccountName, cosmosMath.NewInt(1000))

	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	require.NoError(err)

	// record the stakes after
	reputer0_Stake2, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrsStr[reputerIndexes[0]])
	require.NoError(err)
	reputer1_Stake2, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrsStr[reputerIndexes[1]])
	require.NoError(err)
	reputer2_Stake2, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrsStr[reputerIndexes[2]])
	require.NoError(err)

	reputer3_Stake2, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId1, s.addrsStr[reputerIndexes[0]])
	require.NoError(err)
	reputer4_Stake2, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId1, s.addrsStr[reputerIndexes[1]])
	require.NoError(err)
	reputer5_Stake2, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId1, s.addrsStr[reputerIndexes[2]])
	require.NoError(err)

	// calculate rewards
	reputer0_Reward1 := reputer0_Stake2.Sub(reputer0_Stake1)
	reputer1_Reward1 := reputer1_Stake2.Sub(reputer1_Stake1)
	reputer2_Reward1 := reputer2_Stake2.Sub(reputer2_Stake1)

	reputer3_Reward1 := reputer3_Stake2.Sub(reputer3_Stake1)
	reputer4_Reward1 := reputer4_Stake2.Sub(reputer4_Stake1)
	reputer5_Reward1 := reputer5_Stake2.Sub(reputer5_Stake1)

	// calculate total rewards for each topic
	topic0RewardTotal1 := reputer0_Reward1.Add(reputer1_Reward1).Add(reputer2_Reward1)
	topic1RewardTotal1 := reputer3_Reward1.Add(reputer4_Reward1).Add(reputer5_Reward1)

	// in the first round, the rewards should be equal for each topic
	require.True(topic0RewardTotal0.Equal(topic1RewardTotal0), "%s != %s", topic0RewardTotal0, topic1RewardTotal0)
	// for topic 0, the rewards should be less in the second round
	require.True(topic0RewardTotal0.GT(topic0RewardTotal1), "%s <= %s", topic0RewardTotal0, topic0RewardTotal1)
	// in the second round, the rewards should be greater for topic 1
	require.True(topic0RewardTotal1.LT(topic1RewardTotal1), "%s >= %s", topic0RewardTotal1, topic1RewardTotal1)
	// the rewards for topic 1 should be greater in the second round
	require.True(topic1RewardTotal0.LT(topic1RewardTotal1), "%s >= %s", topic1RewardTotal0, topic1RewardTotal1)
}

func (s *RewardsTestSuite) TestReputerAboveConsensusGetsLessRewards() {
	require := s.Require()

	block := int64(1)
	s.ctx = s.ctx.WithBlockHeight(block)

	alphaRegret := alloraMath.MustNewDecFromString("0.1")

	s.SetParamsForTest()

	reputerIndexes := s.returnIndexes(0, 6)
	workerIndexes := s.returnIndexes(6, 3)

	stake := cosmosMath.NewInt(1000).Mul(inferencesynthesis.CosmosIntOneE18())

	topicId0 := s.setUpTopicWithEpochLength(block, workerIndexes, reputerIndexes, stake, alphaRegret, 5)
	s.T().Logf("TopicId: %v", topicId0)
	err := s.emissionsKeeper.SetPreviousTopicWeight(s.ctx, topicId0, alloraMath.MustNewDecFromString("100"))
	require.NoError(err)

	reputerValues := []TestWorkerValue{
		{Index: reputerIndexes[0], Value: "0.1"},
		{Index: reputerIndexes[1], Value: "0.1"},
		{Index: reputerIndexes[2], Value: "0.1"},
		{Index: reputerIndexes[3], Value: "0.1"},
		{Index: reputerIndexes[4], Value: "0.1"},
		{Index: reputerIndexes[5], Value: "0.9"},
	}

	workerValues := []TestWorkerValue{
		{Index: workerIndexes[0], Value: "0.1"},
		{Index: workerIndexes[1], Value: "0.1"},
		{Index: workerIndexes[2], Value: "0.1"},
	}

	reputer0_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[0].String())
	require.NoError(err)
	reputer1_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[1].String())
	require.NoError(err)
	reputer2_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[2].String())
	require.NoError(err)
	reputer3_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[3].String())
	require.NoError(err)
	reputer4_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[4].String())
	require.NoError(err)
	reputer5_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[5].String())
	require.NoError(err)

	s.getRewardsDistribution(
		topicId0,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.1",
	)

	s.MintTokensToModule(types.AlloraRewardsAccountName, cosmosMath.NewInt(1000))

	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	require.NoError(err)

	reputer0_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[0].String())
	require.NoError(err)
	reputer1_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[1].String())
	require.NoError(err)
	reputer2_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[2].String())
	require.NoError(err)
	reputer3_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[3].String())
	require.NoError(err)
	reputer4_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[4].String())
	require.NoError(err)
	reputer5_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[5].String())
	require.NoError(err)

	reputer0Reward := reputer0_Stake1.Sub(reputer0_Stake0)
	reputer1Reward := reputer1_Stake1.Sub(reputer1_Stake0)
	reputer2Reward := reputer2_Stake1.Sub(reputer2_Stake0)
	reputer3Reward := reputer3_Stake1.Sub(reputer3_Stake0)
	reputer4Reward := reputer4_Stake1.Sub(reputer4_Stake0)
	reputer5Reward := reputer5_Stake1.Sub(reputer5_Stake0)

	require.True(reputer0Reward.Equal(reputer1Reward))
	require.True(reputer1Reward.Equal(reputer2Reward))
	require.True(reputer2Reward.Equal(reputer3Reward))
	require.True(reputer3Reward.Equal(reputer4Reward))
	require.True(reputer5Reward.LT(reputer1Reward))
}

func (s *RewardsTestSuite) TestReputerBelowConsensusGetsLessRewards() {
	require := s.Require()

	block := int64(1)
	s.ctx = s.ctx.WithBlockHeight(block)

	alphaRegret := alloraMath.MustNewDecFromString("0.1")

	s.SetParamsForTest()

	reputerIndexes := s.returnIndexes(0, 6)
	workerIndexes := s.returnIndexes(6, 3)

	stake := cosmosMath.NewInt(1000).Mul(inferencesynthesis.CosmosIntOneE18())

	topicId0 := s.setUpTopicWithEpochLength(block, workerIndexes, reputerIndexes, stake, alphaRegret, 5)

	reputerValues := []TestWorkerValue{
		{Index: reputerIndexes[0], Value: "0.9"},
		{Index: reputerIndexes[1], Value: "0.9"},
		{Index: reputerIndexes[2], Value: "0.9"},
		{Index: reputerIndexes[3], Value: "0.9"},
		{Index: reputerIndexes[4], Value: "0.9"},
		{Index: reputerIndexes[5], Value: "0.1"},
	}

	workerValues := []TestWorkerValue{
		{Index: workerIndexes[0], Value: "0.9"},
		{Index: workerIndexes[1], Value: "0.9"},
		{Index: workerIndexes[2], Value: "0.9"},
	}

	reputer0_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[0].String())
	require.NoError(err)
	reputer1_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[1].String())
	require.NoError(err)
	reputer2_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[2].String())
	require.NoError(err)
	reputer3_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[3].String())
	require.NoError(err)
	reputer4_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[4].String())
	require.NoError(err)
	reputer5_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[5].String())
	require.NoError(err)

	s.getRewardsDistribution(
		topicId0,
		workerValues,
		reputerValues,
		s.addrs[workerIndexes[0]],
		"0.9",
		"0.9",
	)

	s.MintTokensToModule(types.AlloraRewardsAccountName, cosmosMath.NewInt(1000))

	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	require.NoError(err)

	reputer0_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[0].String())
	require.NoError(err)
	reputer1_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[1].String())
	require.NoError(err)
	reputer2_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[2].String())
	require.NoError(err)
	reputer3_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[3].String())
	require.NoError(err)
	reputer4_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[4].String())
	require.NoError(err)
	reputer5_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[5].String())
	require.NoError(err)

	reputer0Reward := reputer0_Stake1.Sub(reputer0_Stake0)
	reputer1Reward := reputer1_Stake1.Sub(reputer1_Stake0)
	reputer2Reward := reputer2_Stake1.Sub(reputer2_Stake0)
	reputer3Reward := reputer3_Stake1.Sub(reputer3_Stake0)
	reputer4Reward := reputer4_Stake1.Sub(reputer4_Stake0)
	reputer5Reward := reputer5_Stake1.Sub(reputer5_Stake0)

	require.True(reputer0Reward.Equal(reputer1Reward))
	require.True(reputer1Reward.Equal(reputer2Reward))
	require.True(reputer2Reward.Equal(reputer3Reward))
	require.True(reputer3Reward.Equal(reputer4Reward))
	require.True(reputer5Reward.LT(reputer1Reward))
}

func (s *RewardsTestSuite) TestRewardForRemainingParticipantsGoUpWhenParticipantDropsOut() {
	// SETUP
	require := s.Require()

	block := int64(1)
	s.ctx = s.ctx.WithBlockHeight(block)

	alphaRegret := alloraMath.MustNewDecFromString("0.1")

	s.SetParamsForTest()

	reputerIndexes := s.returnIndexes(0, 3)
	workerIndexes := s.returnIndexes(3, 3)

	stake := cosmosMath.NewInt(1000).Mul(inferencesynthesis.CosmosIntOneE18())

	topicId0 := s.setUpTopicWithEpochLength(block, workerIndexes, reputerIndexes, stake, alphaRegret, 30)

	// Define values to test
	reputer0Values := []TestWorkerValue{
		{Index: reputerIndexes[0], Value: "0.2"},
		{Index: reputerIndexes[1], Value: "0.2"},
		{Index: reputerIndexes[2], Value: "0.2"},
	}

	workerValues := []TestWorkerValue{
		{Index: workerIndexes[0], Value: "0.2"},
		{Index: workerIndexes[1], Value: "0.2"},
		{Index: workerIndexes[2], Value: "0.2"},
	}

	// Define second round values to test with one less reputer
	reputer1Values := []TestWorkerValue{
		{Index: reputerIndexes[0], Value: "0.2"},
		{Index: reputerIndexes[1], Value: "0.2"},
	}

	// record the stakes before rewards
	reputer0_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[0].String())
	require.NoError(err)
	reputer1_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[1].String())
	require.NoError(err)
	reputer2_Stake0, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[2].String())
	require.NoError(err)

	// do work on the current block
	s.getRewardsDistribution(
		topicId0,
		workerValues,
		reputer0Values,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.1",
	)

	// create tokens to reward with
	s.MintTokensToModule(types.AlloraRewardsAccountName, cosmosMath.NewInt(1000))

	nextBlock, _, err := s.emissionsKeeper.GetNextPossibleChurningBlockByTopicId(s.ctx, topicId0)
	s.Require().NoError(err)
	s.ctx = s.ctx.WithBlockHeight(nextBlock)
	// force rewards to be distributed
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	require.NoError(err)

	topic, err := s.emissionsKeeper.GetTopic(s.ctx, topicId0)
	s.Require().NoError(err)
	// record the updated stakes after rewards
	reputer0_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[0].String())
	require.NoError(err)
	reputer1_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[1].String())
	require.NoError(err)
	reputer2_Stake1, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[2].String())
	require.NoError(err)

	// calculate the rewards for each reputer
	reputer0_Reward0 := reputer0_Stake1.Sub(reputer0_Stake0)
	reputer1_Reward0 := reputer1_Stake1.Sub(reputer1_Stake0)
	reputer2_Reward0 := reputer2_Stake1.Sub(reputer2_Stake0)

	// fund the topic again for future rewards
	const topicFundAmount int64 = 1000

	fundTopic := func(topicId uint64, funderAddr sdk.AccAddress, amount int64) {
		s.MintTokensToAddress(funderAddr, cosmosMath.NewInt(amount))
		fundTopicMessage := types.FundTopicRequest{
			Sender:  funderAddr.String(),
			TopicId: topicId,
			Amount:  cosmosMath.NewInt(amount),
		}
		_, err = s.msgServer.FundTopic(s.ctx, &fundTopicMessage)
		require.NoError(err)
	}

	fundTopic(topicId0, s.addrs[0], topicFundAmount)

	// do work on the current block, but with one less reputer
	s.getRewardsDistribution(
		topic.Id,
		workerValues,
		reputer1Values,
		s.addrs[workerIndexes[0]],
		"0.1",
		"0.1",
	)

	// create tokens to reward with
	s.MintTokensToModule(types.AlloraRewardsAccountName, cosmosMath.NewInt(1000))

	nextBlock, _, err = s.emissionsKeeper.GetNextPossibleChurningBlockByTopicId(s.ctx, topicId0)
	s.Require().NoError(err)
	s.ctx = s.ctx.WithBlockHeight(nextBlock)
	// force rewards to be distributed
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	require.NoError(err)

	// check the updated stakes after rewards
	reputer0_Stake2, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[0].String())
	require.NoError(err)
	reputer1_Stake2, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[1].String())
	require.NoError(err)
	reputer2_Stake2, err := s.emissionsKeeper.GetStakeReputerAuthority(s.ctx, topicId0, s.addrs[2].String())
	require.NoError(err)

	// calculate the rewards for each reputer
	reputer0_Reward1 := reputer0_Stake2.Sub(reputer0_Stake1)
	reputer1_Reward1 := reputer1_Stake2.Sub(reputer1_Stake1)
	reputer2_Reward1 := reputer2_Stake2.Sub(reputer2_Stake1)

	// sanity check that participating reputer rewards went up, but non participating reputer
	// rewards went to zero
	require.True(reputer0_Reward1.GT(reputer0_Reward0))
	require.True(reputer1_Reward1.GT(reputer1_Reward0))
	require.True(reputer2_Reward0.GT(cosmosMath.ZeroInt()))
	require.True(reputer2_Reward1.Equal(cosmosMath.ZeroInt()))
}

func (s *RewardsTestSuite) TestRewardIncreaseContiouslyAfterTopicReactivated() {
	// SETUP
	require := s.Require()

	block := int64(1)
	s.ctx = s.ctx.WithBlockHeight(block)

	alphaRegret := alloraMath.MustNewDecFromString("0.1")

	s.SetParamsForTest()

	reputer0Indexes := s.returnIndexes(0, 3)
	worker0Indexes := s.returnIndexes(3, 3)

	reputer1Indexes := s.returnIndexes(6, 3)
	worker1Indexes := s.returnIndexes(9, 3)

	stake := cosmosMath.NewInt(1000).Mul(inferencesynthesis.CosmosIntOneE18())

	topicId0 := s.setUpTopicWithEpochLength(block, worker0Indexes, reputer0Indexes, stake, alphaRegret, 30)
	topicId1 := s.setUpTopicWithEpochLength(block, worker1Indexes, reputer1Indexes, stake, alphaRegret, 60)
	cosmosOneE18 := inferencesynthesis.CosmosIntOneE18()
	// Add Stake for reputers
	var stakes = []cosmosMath.Int{
		cosmosMath.NewInt(1176644).Mul(cosmosOneE18),
		cosmosMath.NewInt(384623).Mul(cosmosOneE18),
		cosmosMath.NewInt(394676).Mul(cosmosOneE18),
		cosmosMath.NewInt(207999).Mul(cosmosOneE18),
		cosmosMath.NewInt(368582).Mul(cosmosOneE18),
	}
	for i, index := range reputer0Indexes {
		s.MintTokensToAddress(s.addrs[index], stakes[i])
		_, err := s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
			Sender:  s.addrs[index].String(),
			Amount:  stakes[i],
			TopicId: topicId0,
		})
		s.Require().NoError(err)
	}
	for i, index := range reputer1Indexes {
		s.MintTokensToAddress(s.addrs[index], stakes[i].MulRaw(2))
		_, err := s.msgServer.AddStake(s.ctx, &types.AddStakeRequest{
			Sender:  s.addrs[index].String(),
			Amount:  stakes[i].MulRaw(2),
			TopicId: topicId1,
		})
		s.Require().NoError(err)
	}

	initialStake := cosmosMath.NewInt(1000)
	s.MintTokensToAddress(s.addrs[reputer0Indexes[0]], initialStake)
	fundTopic0Message := types.FundTopicRequest{
		Sender:  s.addrs[reputer0Indexes[0]].String(),
		TopicId: topicId0,
		Amount:  initialStake,
	}
	_, err := s.msgServer.FundTopic(s.ctx, &fundTopic0Message)
	s.Require().NoError(err)

	s.MintTokensToAddress(s.addrs[reputer1Indexes[0]], initialStake.MulRaw(2))
	fundTopic1Message := types.FundTopicRequest{
		Sender:  s.addrs[reputer1Indexes[0]].String(),
		TopicId: topicId1,
		Amount:  initialStake.MulRaw(2),
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopic1Message)
	s.Require().NoError(err)

	topic0, err := s.emissionsKeeper.GetTopic(s.ctx, topicId0)
	require.NoError(err)
	blocklHeight := topic0.EpochLength
	s.ctx = s.ctx.WithBlockHeight(blocklHeight)
	err = s.emissionsKeeper.SetRewardCurrentBlockEmission(s.ctx, cosmosMath.NewInt(100))
	s.Require().NoError(err)
	err = s.emissionsAppModule.EndBlock(s.ctx)
	s.Require().NoError(err)
	worker0Values := []TestWorkerValue{
		{Index: worker0Indexes[0], Value: "0.1"},
		{Index: worker0Indexes[1], Value: "0.2"},
		{Index: worker0Indexes[2], Value: "0.3"},
	}
	reputer0Values := []TestWorkerValue{
		{Index: reputer0Indexes[0], Value: "0.1"},
		{Index: reputer0Indexes[1], Value: "0.2"},
		{Index: reputer0Indexes[2], Value: "0.3"},
	}
	worker1Values := []TestWorkerValue{
		{Index: worker1Indexes[0], Value: "0.4"},
		{Index: worker1Indexes[1], Value: "0.5"},
		{Index: worker1Indexes[2], Value: "0.6"},
	}
	reputer1Values := []TestWorkerValue{
		{Index: reputer1Indexes[0], Value: "0.4"},
		{Index: reputer1Indexes[1], Value: "0.5"},
		{Index: reputer1Indexes[2], Value: "0.6"},
	}
	rewardsDistribution0_0 := s.getRewardsDistribution(
		topicId0,
		worker0Values,
		reputer0Values,
		s.addrs[worker0Indexes[0]],
		"0.1",
		"0.1",
	)

	_ = s.getRewardsDistribution(
		topicId1,
		worker1Values,
		reputer1Values,
		s.addrs[worker1Indexes[0]],
		"0.2",
		"0.2",
	)

	// Check if first topic is inactivated due to low weight
	isActivated, err := s.emissionsKeeper.IsTopicActive(s.ctx, topicId0)
	require.NoError(err)
	require.False(isActivated)

	// Activate first topic
	s.MintTokensToAddress(s.addrs[reputer0Indexes[0]], initialStake)
	fundTopic0Message = types.FundTopicRequest{
		Sender:  s.addrs[reputer0Indexes[0]].String(),
		TopicId: topicId0,
		Amount:  initialStake.MulRaw(3),
	}
	_, err = s.msgServer.FundTopic(s.ctx, &fundTopic0Message)
	s.Require().NoError(err)

	isActivated, err = s.emissionsKeeper.IsTopicActive(s.ctx, topicId0)
	require.NoError(err)
	require.True(isActivated)

	rewardsDistribution0_1 := s.getRewardsDistribution(
		topicId0,
		worker0Values,
		reputer0Values,
		s.addrs[worker0Indexes[0]],
		"0.1",
		"0.1",
	)
	require.Equal(len(rewardsDistribution0_1), len(rewardsDistribution0_0))
}

func (s *RewardsTestSuite) returnIndexes(start, count int) []int {
	res := make([]int, count)
	for ind := start; ind < start+count; ind++ {
		res[ind-start] = ind
	}
	return res
}

func decPtr(s string) *alloraMath.Dec {
	d := alloraMath.MustNewDecFromString(s)
	return &d
}

func (s *RewardsTestSuite) TestCalcTopicRewards() {
	testCases := []struct {
		name                string
		setupFunc           func() (map[uint64]*alloraMath.Dec, []uint64, alloraMath.Dec, alloraMath.Dec, map[uint64]int64, alloraMath.Dec)
		expectedRewardsFunc func(map[uint64]*alloraMath.Dec) bool
		expectedError       error
	}{
		{
			name: "Happy path - multiple topics",
			setupFunc: func() (map[uint64]*alloraMath.Dec,
				[]uint64,
				alloraMath.Dec,
				alloraMath.Dec,
				map[uint64]int64,
				alloraMath.Dec) {
				weights := map[uint64]*alloraMath.Dec{
					1: decPtr("0.5"),
					2: decPtr("0.3"),
					3: decPtr("0.2"),
				}
				sortedTopics := []uint64{1, 2, 3}
				sumWeight := alloraMath.MustNewDecFromString("1.0")
				totalReward := alloraMath.MustNewDecFromString("1000.0")
				epochLengths := map[uint64]int64{
					1: 100,
					2: 100,
					3: 100,
				}
				currentBlockEmission := alloraMath.MustNewDecFromString("10.0")
				return weights, sortedTopics, sumWeight, totalReward, epochLengths, currentBlockEmission
			},
			expectedRewardsFunc: func(rewards map[uint64]*alloraMath.Dec) bool {
				expected := map[uint64]*alloraMath.Dec{
					1: decPtr("500.0"),
					2: decPtr("300.0"),
					3: decPtr("200.0"),
				}
				return s.compareRewards(rewards, expected)
			},
			expectedError: nil,
		},
		{
			name: "Single topic",
			setupFunc: func() (map[uint64]*alloraMath.Dec, []uint64, alloraMath.Dec, alloraMath.Dec, map[uint64]int64, alloraMath.Dec) {
				weights := map[uint64]*alloraMath.Dec{
					1: decPtr("1.0"),
				}
				sortedTopics := []uint64{1}
				sumWeight := alloraMath.MustNewDecFromString("1.0")
				totalReward := alloraMath.MustNewDecFromString("1000.0")
				epochLengths := map[uint64]int64{
					1: 100,
				}
				currentBlockEmission := alloraMath.MustNewDecFromString("10.0")
				return weights, sortedTopics, sumWeight, totalReward, epochLengths, currentBlockEmission
			},
			expectedRewardsFunc: func(rewards map[uint64]*alloraMath.Dec) bool {
				expected := map[uint64]*alloraMath.Dec{
					1: decPtr("1000.0"),
				}
				return s.compareRewards(rewards, expected)
			},
			expectedError: nil,
		},
		{
			name: "Zero total reward",
			setupFunc: func() (map[uint64]*alloraMath.Dec, []uint64, alloraMath.Dec, alloraMath.Dec, map[uint64]int64, alloraMath.Dec) {
				weights := map[uint64]*alloraMath.Dec{
					1: decPtr("0.5"),
					2: decPtr("0.5"),
				}
				sortedTopics := []uint64{1, 2}
				sumWeight := alloraMath.MustNewDecFromString("1.0")
				totalReward := alloraMath.ZeroDec()
				epochLengths := map[uint64]int64{
					1: 100,
					2: 100,
				}
				currentBlockEmission := alloraMath.MustNewDecFromString("10.0")
				return weights, sortedTopics, sumWeight, totalReward, epochLengths, currentBlockEmission
			},
			expectedRewardsFunc: func(rewards map[uint64]*alloraMath.Dec) bool {
				expected := map[uint64]*alloraMath.Dec{
					1: decPtr("0"),
					2: decPtr("0"),
				}
				return s.compareRewards(rewards, expected)
			},
			expectedError: nil,
		},
		{
			name: "Different epoch lengths",
			setupFunc: func() (map[uint64]*alloraMath.Dec, []uint64, alloraMath.Dec, alloraMath.Dec, map[uint64]int64, alloraMath.Dec) {
				weights := map[uint64]*alloraMath.Dec{
					1: decPtr("0.75"),
					2: decPtr("0.25"),
				}
				sortedTopics := []uint64{1, 2}
				sumWeight := alloraMath.MustNewDecFromString("1")
				totalReward := alloraMath.MustNewDecFromString("1000.0")
				epochLengths := map[uint64]int64{
					1: 100,
					2: 200,
				}
				currentBlockEmission := alloraMath.MustNewDecFromString("1.0")
				return weights, sortedTopics, sumWeight, totalReward, epochLengths, currentBlockEmission
			},
			expectedRewardsFunc: func(rewards map[uint64]*alloraMath.Dec) bool {
				expected := map[uint64]*alloraMath.Dec{
					1: decPtr("75"),
					2: decPtr("50"),
				}
				return s.compareRewards(rewards, expected)
			},
			expectedError: nil,
		},
		{
			name: "Very small weights",
			setupFunc: func() (map[uint64]*alloraMath.Dec, []uint64, alloraMath.Dec, alloraMath.Dec, map[uint64]int64, alloraMath.Dec) {
				weights := map[uint64]*alloraMath.Dec{
					1: decPtr("0.000000000000000001"),
					2: decPtr("0.000000000000000002"),
				}
				sortedTopics := []uint64{1, 2}
				sumWeight := alloraMath.MustNewDecFromString("0.000000000000000003")
				totalReward := alloraMath.MustNewDecFromString("1000.0")
				epochLengths := map[uint64]int64{
					1: 100,
					2: 100,
				}
				currentBlockEmission := alloraMath.MustNewDecFromString("10.0")
				return weights, sortedTopics, sumWeight, totalReward, epochLengths, currentBlockEmission
			},
			expectedRewardsFunc: func(rewards map[uint64]*alloraMath.Dec) bool {
				expected := map[uint64]*alloraMath.Dec{
					1: decPtr("333.333333333333333333"),
					2: decPtr("666.666666666666666667"),
				}
				return s.compareRewards(rewards, expected)
			},
			expectedError: nil,
		},
		{
			name: "Mismatched weights and sorted topics",
			setupFunc: func() (map[uint64]*alloraMath.Dec, []uint64, alloraMath.Dec, alloraMath.Dec, map[uint64]int64, alloraMath.Dec) {
				weights := map[uint64]*alloraMath.Dec{
					1: decPtr("0.5"),
					2: decPtr("0.5"),
				}
				sortedTopics := []uint64{1, 2, 3}
				sumWeight := alloraMath.MustNewDecFromString("1.0")
				totalReward := alloraMath.MustNewDecFromString("1000.0")
				epochLengths := map[uint64]int64{
					1: 100,
					2: 100,
				}
				currentBlockEmission := alloraMath.MustNewDecFromString("10.0")
				return weights, sortedTopics, sumWeight, totalReward, epochLengths, currentBlockEmission
			},
			expectedRewardsFunc: func(rewards map[uint64]*alloraMath.Dec) bool {
				return len(rewards) == 2
			},
			expectedError: types.ErrInvalidValue,
		},
		{
			name: "Treasury lower than rewards",
			setupFunc: func() (map[uint64]*alloraMath.Dec, []uint64, alloraMath.Dec, alloraMath.Dec, map[uint64]int64, alloraMath.Dec) {
				weights := map[uint64]*alloraMath.Dec{
					1: decPtr("0.5"),
					2: decPtr("0.3"),
					3: decPtr("0.2"),
				}
				sortedTopics := []uint64{1, 2, 3}
				sumWeight := alloraMath.MustNewDecFromString("1.0")
				totalReward := alloraMath.MustNewDecFromString("50.0") // Lower than expected rewards
				epochLengths := map[uint64]int64{
					1: 100,
					2: 100,
					3: 100,
				}
				currentBlockEmission := alloraMath.MustNewDecFromString("1.0")
				return weights, sortedTopics, sumWeight, totalReward, epochLengths, currentBlockEmission
			},
			expectedRewardsFunc: func(rewards map[uint64]*alloraMath.Dec) bool {
				expected := map[uint64]*alloraMath.Dec{
					1: decPtr("25.0"),
					2: decPtr("15.0"),
					3: decPtr("10.0"),
				}
				return s.compareRewards(rewards, expected)
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			weights, sortedTopics, sumWeight, totalReward, epochLengths, currentBlockEmission := tc.setupFunc()
			args := rewards.CalcTopicRewardsArgs{
				Ctx:                             s.ctx,
				Weights:                         weights,
				SortedTopics:                    sortedTopics,
				SumTopicWeights:                 sumWeight,
				TotalAvailableInRewardsTreasury: totalReward,
				EpochLengths:                    epochLengths,
				CurrentRewardsEmissionPerBlock:  currentBlockEmission,
			}
			rewards, err := rewards.CalcTopicRewards(args)

			if tc.expectedError != nil {
				s.Require().ErrorIs(err, tc.expectedError)
			} else {
				s.Require().NoError(err)
				s.Require().True(tc.expectedRewardsFunc(rewards))
			}
		})
	}
}

// Helper function to compare rewards
func (s *RewardsTestSuite) compareRewards(actual, expected map[uint64]*alloraMath.Dec) bool {
	if len(actual) != len(expected) {
		return false
	}
	for topicID, expectedReward := range expected {
		actualReward, exists := actual[topicID]
		if !exists {
			return false
		}
		inDelta, err := alloraMath.InDelta(*actualReward, *expectedReward, alloraMath.MustNewDecFromString("0.0001"))
		s.Require().NoError(err)
		s.Require().True(inDelta)
		if !inDelta {
			s.T().Logf("Actual   %v not found", actualReward)
			s.T().Logf("Expected %v not found", expectedReward)
			return false
		}
	}
	return true
}
