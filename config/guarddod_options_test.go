package config

import (
    "testing"
)

func TestGO(t *testing.T) {
    o := NewGuarddogOptions()
    _ = o.Validate()
}
