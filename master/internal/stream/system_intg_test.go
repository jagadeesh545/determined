//go:build integration
// +build integration

package stream

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/determined-ai/determined/master/internal/db"
	"github.com/determined-ai/determined/master/pkg/stream"
	"github.com/determined-ai/determined/master/test/streamdata"
)

func simpleUpsert(i interface{}) interface{} {
	return i
}

func recordDeletion(s1, s2 string) interface{} {
	fmt.Println("recording deletion:", s1+" and "+s2)
	return s2
}

type startupReadWriter struct {
	data           []interface{}
	startupMessage *StartupMsg
}

func (t *startupReadWriter) ReadJSON(data interface{}) error {
	if t.startupMessage == nil {
		return fmt.Errorf("startup message has been sent")
	}
	targetMsg, ok := data.(*StartupMsg)
	if !ok {
		return fmt.Errorf("target message type is not a pointer to StartupMsg")
	}
	targetMsg.Known = t.startupMessage.Known
	targetMsg.Subscribe = t.startupMessage.Subscribe
	t.startupMessage = nil
	return nil
}

func (t *startupReadWriter) Write(data interface{}) error {
	t.data = append(t.data, data)
	return nil
}

func TestReadWriter(t *testing.T) {
	startupMessage := StartupMsg{
		Known: KnownKeySet{Trials: "1,2,3"},
		Subscribe: SubscriptionSpecSet{
			Trials: &TrialSubscriptionSpec{
				ExperimentIds: []int{1},
				Since:         0,
			},
		},
	}

	trw := startupReadWriter{
		startupMessage: &startupMessage,
	}

	emptyMsg := StartupMsg{}
	err := trw.ReadJSON(&emptyMsg)
	require.NoError(t, err)
	require.Equal(t, emptyMsg.Known, startupMessage.Known)
	require.Equal(t, emptyMsg.Subscribe, startupMessage.Subscribe)
	require.True(t, trw.startupMessage == nil)

	err = trw.Write("test")
	require.NoError(t, err)
	require.Equal(t, 1, len(trw.data))
	dataStr, ok := trw.data[0].(string)
	require.True(t, ok)
	require.Equal(t, "test", dataStr)
}

func TestStartup(t *testing.T) {
	ctx := context.TODO()
	pgDB, cleanup := db.MustResolveNewPostgresDatabase(t)
	defer func() {
		fmt.Println("cleaning up?")
		cleanup()
	}()

	fmt.Println("START OF TEST!")
	trials := streamdata.GenerateStreamTrials()
	trials.MustMigrate(t, pgDB, "file://../../static/migrations")

	startupMessage := StartupMsg{
		Known: KnownKeySet{
			Trials: "1,2,3",
		},
		Subscribe: SubscriptionSpecSet{
			Trials: &TrialSubscriptionSpec{
				ExperimentIds: []int{1}, // trials 1,2,3
				Since:         0,
			},
		},
	}

	tester := startupReadWriter{
		startupMessage: &startupMessage,
	}
	publisherSet := NewPublisherSet()
	err := publisherSet.entrypoint(ctx, &tester, simpleUpsert, recordDeletion)
	require.NoError(t, err)

	deletions, trialMsgs, err := splitDeletionsAndTrials(tester.data)
	require.NoError(t, err)
	require.Equal(t, 1, len(deletions), "did not receive 1 deletion message")
	require.Equal(t, "0", deletions[0], "expected deleted trials to be 0, not %s", deletions[0])
	require.Equal(t, 0, len(trialMsgs), "received unexpected trial message")

	// don't know about trial 3, and trial 4 doesn't exist
	startupMessage = StartupMsg{
		Known: KnownKeySet{
			Trials: "1,2,4",
		},
		Subscribe: SubscriptionSpecSet{
			Trials: &TrialSubscriptionSpec{
				ExperimentIds: []int{1}, // trials 1,2,3
				Since:         0,
			},
		},
	}
	tester.startupMessage = &startupMessage
	publisherSet = NewPublisherSet()
	err = publisherSet.entrypoint(ctx, &tester, simpleUpsert, recordDeletion)
	require.NoError(t, err)
	deletions, trialMsgs, err = splitDeletionsAndTrials(tester.data)
	require.NoError(t, err)
	fmt.Println("deletions:", deletions)
	require.Equal(t, 1, len(deletions), "did not receive 1 deletion message")
	require.Equal(t, "4", deletions[0], "expected deleted trials to be 4, not %s", deletions[0])
	require.Equal(t, 1, len(trialMsgs), "received unexpected trial message")
	require.Equal(t, 3, trialMsgs[0].ID, "expected trialMsg with ID 3, received ID %d",
		trialMsgs[0].ID)

	fmt.Println("END OF TEST!")

	// Subscribe to all known trials, but 4 doesn't exist
	//startupMessage = StartupMsg{
	//	Known: KnownKeySet{
	//		Trials: "1,2,3,4",
	//	},
	//	Subscribe: SubscriptionSpecSet{
	//		Trials: &TrialSubscriptionSpec{
	//			TrialIds: []int{1, 2, 3, 4},
	//			Since:    0,
	//		},
	//	},
	//}
	//tester.startupMessage = &startupMessage
	//err = publisherSet.entrypoint(ctx, &tester, simpleUpsert, recordDeletion)
	//require.NoError(t, err)
	//deletions, trialMsgs, err = splitDeletionsAndTrials(tester.data)
	//require.NoError(t, err)
	//require.Equal(t, 1, len(deletions), "did not receive 1 deletion message")
	//require.Equal(t, "4", deletions[0], "expected deleted trials to be 4, not %s", deletions[0])
	//require.Equal(t, 0, len(trialMsgs), "received unexpected trial message")
}

func TestStartupTrial(t *testing.T) {
	pgDB, cleanup := db.MustResolveNewPostgresDatabase(t)

	defer cleanup()

	trials := streamdata.GenerateStreamTrials()
	trials.MustMigrate(t, pgDB, "file://../../static/migrations")

	startupMessage := StartupMsg{
		Known: KnownKeySet{
			Trials: "1,2,3",
		},
		Subscribe: SubscriptionSpecSet{
			Trials: &TrialSubscriptionSpec{
				ExperimentIds: []int{1}, // trials 1,2,3
				Since:         0,
			},
		},
	}
	messages := testSubscriptionSetStartup(t, startupMessage)
	deletions, trialMsgs, err := splitDeletionsAndTrials(messages)
	require.NoError(t, err)
	require.Equal(t, 1, len(deletions), "did not receive 1 deletion message")
	require.Equal(t, "0", deletions[0], "expected deleted trials to be 0, not %s", deletions[0])
	require.Equal(t, 0, len(trialMsgs), "received unexpected trial message")

	// don't know about trial 3, and trial 4 doesn't exist
	startupMessage = StartupMsg{
		Known: KnownKeySet{
			Trials: "1,2,4",
		},
		Subscribe: SubscriptionSpecSet{
			Trials: &TrialSubscriptionSpec{
				ExperimentIds: []int{1}, // trials 1,2,3
				Since:         0,
			},
		},
	}
	messages = testSubscriptionSetStartup(t, startupMessage)
	deletions, trialMsgs, err = splitDeletionsAndTrials(messages)
	require.NoError(t, err)
	require.Equal(t, 1, len(deletions), "did not receive 1 deletion message")
	require.Equal(t, "4", deletions[0], "expected deleted trials to be 4, not %s", deletions[0])
	require.Equal(t, 1, len(trialMsgs), "received unexpected trial message")
	require.Equal(t, 3, trialMsgs[0].ID, "expected trialMsg with ID 3, received ID %d",
		trialMsgs[0].ID)

	// Subscribe to all known trials, but 4 doesn't exist
	startupMessage = StartupMsg{
		Known: KnownKeySet{
			Trials: "1,2,3,4",
		},
		Subscribe: SubscriptionSpecSet{
			Trials: &TrialSubscriptionSpec{
				TrialIds: []int{1, 2, 3, 4},
				Since:    0,
			},
		},
	}
	messages = testSubscriptionSetStartup(t, startupMessage)
	deletions, trialMsgs, err = splitDeletionsAndTrials(messages)
	require.NoError(t, err)
	require.Equal(t, 1, len(deletions), "did not receive 1 deletion message")
	require.Equal(t, "4", deletions[0], "expected deleted trials to be 4, not %s", deletions[0])
	require.Equal(t, 0, len(trialMsgs), "received unexpected trial message")
}

func splitDeletionsAndTrials(messages []interface{}) ([]string, []*TrialMsg, error) {
	var deletions []string
	var trialMsgs []*TrialMsg
	for _, msg := range messages {
		if deletion, ok := msg.(string); ok {
			deletions = append(deletions, deletion)
		} else if trialMsg, ok := msg.(*TrialMsg); ok {
			trialMsgs = append(trialMsgs, trialMsg)
		} else {
			return nil, nil, fmt.Errorf("expected a string or *TrialMsg, but received %t",
				reflect.TypeOf(msg))
		}
	}
	return deletions, trialMsgs, nil
}

func testSubscriptionSetStartup(t *testing.T, startupMessage StartupMsg) []interface{} {
	ctx := context.TODO()
	streamer := stream.NewStreamer()
	publisherSet := NewPublisherSet()
	subSet := NewSubscriptionSet(streamer, publisherSet, simpleUpsert, recordDeletion)
	messages, err := subSet.Startup(startupMessage, ctx)
	require.NoError(t, err, "error running startup")

	return messages
}
