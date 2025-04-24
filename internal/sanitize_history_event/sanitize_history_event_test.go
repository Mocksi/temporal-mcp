package sanitize_history_event

import (
	"bufio"
	"context"
	"fmt"
	"github.com/mocksi/temporal-mcp/internal/config"
	"github.com/mocksi/temporal-mcp/internal/temporal"
	"github.com/stretchr/testify/require"
	temporal_enums "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"strings"
	"testing"
)

const TEST_DIR = "test_data"
const ORIGINAL_SUFFIX = "_original.jsonl"

func TestSanitizeHistoryEvent(t *testing.T) {
	// To generate new test files from a real workflow history, uncomment the following line
	// generateTestJson(t, "localhost:7233", "default", "someWorkflowID")

	workflowIDs := make([]string, 0)
	entries, err := os.ReadDir(TEST_DIR)
	require.NoError(t, err)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), ORIGINAL_SUFFIX) {
			workflowIDs = append(workflowIDs, entry.Name()[0:len(entry.Name())-len(ORIGINAL_SUFFIX)])
		}
	}

	for _, workflowID := range workflowIDs {
		t.Run(fmt.Sprintf("history for %s", workflowID), func(t *testing.T) {
			original, sanitized := getTestFilenames(workflowID)

			originalEvents := readEvents(t, original)
			sanitizedEvents := readEvents(t, sanitized)
			require.Equal(t, len(originalEvents), len(sanitizedEvents))

			for i, actualEvent := range originalEvents {
				SanitizeHistoryEvent(actualEvent)
				require.Equal(t, sanitizedEvents[i], actualEvent)
			}
		})
	}
}

func generateTestJson(t *testing.T, hostport string, namespace string, workflowID string) {
	tClient, err := temporal.NewTemporalClient(config.TemporalConfig{
		HostPort:         hostport,
		Namespace:        namespace,
		Environment:      "local",
		DefaultTaskQueue: "unused",
	})
	require.NoError(t, err)

	iter := tClient.GetWorkflowHistory(context.Background(), workflowID, "", false, temporal_enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)

	original, sanitized := getTestFilenames(workflowID)

	originalFile, err := os.Create(original)
	require.NoError(t, err)
	defer originalFile.Close()

	sanitizedFile, err := os.Create(sanitized)
	require.NoError(t, err)
	defer sanitizedFile.Close()

	for iter.HasNext() {
		event, err := iter.Next()
		require.NoError(t, err)

		writeEvent(t, originalFile, event)
		SanitizeHistoryEvent(event)
		writeEvent(t, sanitizedFile, event)
	}
}

func writeEvent(t *testing.T, file *os.File, event *history.HistoryEvent) {
	bytes, err := protojson.Marshal(event)
	require.NoError(t, err)

	bytes = append(bytes, '\n')

	n, err := file.Write(bytes)
	require.NoError(t, err)
	require.Equal(t, len(bytes), n)
}

func readEvents(t *testing.T, filename string) []*history.HistoryEvent {
	f, err := os.Open(filename)
	require.NoError(t, err)
	defer f.Close()

	var events []*history.HistoryEvent
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		eventJson := scanner.Text()
		event := &history.HistoryEvent{}
		err := protojson.Unmarshal([]byte(eventJson), event)
		require.NoError(t, err)
		events = append(events, event)
	}

	return events
}

func getTestFilenames(workflowID string) (string, string) {
	original := fmt.Sprintf("%s/%s%s", TEST_DIR, workflowID, ORIGINAL_SUFFIX)
	sanitized := fmt.Sprintf("%s/%s_sanitized.jsonl", TEST_DIR, workflowID)
	return original, sanitized
}
