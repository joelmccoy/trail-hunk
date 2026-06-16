package app

import (
	"fmt"
	"io"
)

func Run(stdout io.Writer, cfg Config) error {
	_, err := fmt.Fprintf(stdout, "trail-hunk provider=%s model=%s\n", cfg.Provider, cfg.Model)
	return err
}
