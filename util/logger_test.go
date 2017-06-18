package util

import (
    "testing"
)

func TestCanLog(t *testing.T) {
    logger, err := NewLogger(2, true, "prefix")
    if err != nil {
        t.Fatalf("Failed to create logger: %s", err)
    }
    logger.Error("Test error message with args: %d", 1)
    logger.Info("Test info message with args: %d", 1)
}
