package newstress_test

import (
	"fmt"
	"testing"
	"time"
	"sync"

	testcommon "github.com/allora-network/allora-chain/test/common"
)

const topicFunds int64 = 1e6
const stakeToAdd uint64 = 9e4

func TestNewStressTestSuite(t *testing.T) {
	t.Log(">>> Environment <<<")
	seed := testcommon.LookupEnvInt(t, "SEED", 1)
	rpcMode := testcommon.LookupRpcMode(t, "RPC_MODE", testcommon.SingleRpc)
	rpcEndpoints := testcommon.LookupEnvStringArray("RPC_URLS", []string{"http://localhost:26657"})

	testConfig := testcommon.NewTestConfig(
		t,
		rpcMode,
		rpcEndpoints,
		"../localnet/genesis",
		seed,
	)

	// Read env vars with defaults
	maxIterations := testcommon.LookupEnvInt(t, "MAX_ITERATIONS", 1000)
	epochLength := testcommon.LookupEnvInt(t, "EPOCH_LENGTH", 12) // in blocks

	
	numTopics := testcommon.LookupEnvInt(t, "NUM_TOPICS", 10)
	workersPerTopic := testcommon.LookupEnvInt(t, "WORKERS_PER_TOPIC", 5)
	reputersPerTopic := testcommon.LookupEnvInt(t, "REPUTERS_PER_TOPIC", 4)
	createTopicsSameBlock := testcommon.LookupEnvBool(t, "CREATE_TOPICS_SAME_BLOCK", false)

	t.Log("Max Iterations: ", maxIterations)
	t.Log("Epoch Length: ", epochLength)
	t.Log("Number of Topics: ", numTopics)
	t.Log("Workers per Topic: ", workersPerTopic)
	t.Log("Reputers per Topic: ", reputersPerTopic)
	t.Log("Create Topics in Same Block: ", createTopicsSameBlock)
	t.Log(">>> Starting Test <<<")
	timestr := fmt.Sprintf(">>> Starting %s <<<", time.Now().Format(time.RFC850))
	t.Log(timestr)

	numActors := workersPerTopic + reputersPerTopic
	_, simulationData := simulateSetUp(&testConfig, numActors, epochLength)

	topicIds, err := startCreateTopicsAndRegister(
		&testConfig,
		simulationData,
		numTopics,
		workersPerTopic,
		reputersPerTopic,
		createTopicsSameBlock,
	)
	requireNoError(t, simulationData.failOnErr, err)

	fmt.Println("Topic IDs: ", topicIds)

	err = simulateAutomatic(
		&testConfig,
		simulationData,
		topicIds,
	)
	requireNoError(t, simulationData.failOnErr, err)
}

// startCreateTopicsAndRegister creates topics and registers workers and reputers according to env variables
func startCreateTopicsAndRegister(
	m *testcommon.TestConfig,
	data *SimulationData,
	numTopics int,
	workersPerTopic int,
	reputersPerTopic int,
	createTopicsSameBlock bool,
) ([]uint64, error) {
	actor := data.actors[0]
	topicIds, err := createTopics(m, actor, numTopics, data.epochLength, createTopicsSameBlock)
	if err != nil {
		return nil, err
	}

	for _, topicId := range topicIds {
		fmt.Println("Registering workers in topic: ", workersPerTopic)
		err = registerWorkers(m, data.actors, topicId, data, workersPerTopic)
		if err != nil {
			fmt.Println("Error registering workers: ", err)
			return nil, err
		}
		fmt.Println("Registering reputers in topic: ", reputersPerTopic)
		err = registerReputers(m, data.actors, topicId, data, reputersPerTopic)
		if err != nil {
			return nil, err
		}
		fmt.Println("Staking reputers in topic: ", topicId)
		err = stakeReputers(m, topicId, data.actors, stakeToAdd)
		if err != nil {
			return nil, err
		}
	}

	err = fundTopics(m, topicIds, actor, topicFunds)
	if err != nil {
		return nil, err
	}

	return topicIds, nil
}

func simulateAutomatic(
	m *testcommon.TestConfig,
	data *SimulationData,
	topicIds []uint64,
) error {
	// Create wait group to track all goroutines
	var wg sync.WaitGroup
	// Create error channel to catch any errors from goroutines
	errChan := make(chan error, len(topicIds)*2)

	// For each topic, start a worker routine and a reputer routine
	for _, topicId := range topicIds {
		// Add 2 to waitgroup (1 for worker, 1 for reputer)
		wg.Add(2)

		// Start worker routine
		go func(tid uint64) {
			defer wg.Done()
			err := runTopicWorkersLoop(m, data, tid)
			if err != nil {
				errChan <- fmt.Errorf("worker routine failed for topic %d: %w", tid, err)
			}
		}(topicId)

		// Start reputer routine
		go func(tid uint64) {
			defer wg.Done()
			err := runReputersProcess(m, data, tid)
			if err != nil {
				errChan <- fmt.Errorf("reputer routine failed for topic %d: %w", tid, err)
			}
		}(topicId)
	}

	// Create a channel that closes when all goroutines are done
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for either completion or an error
	select {
	case err := <-errChan:
		return err
	case <-done:
		return nil
	case <-time.After(5 * time.Minute): //TODO: Add a timeout to prevent infinite running in ENV VARIABLES
		return fmt.Errorf("simulation timed out after 5 minutes")
	}
}
