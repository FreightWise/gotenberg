package pdftoppm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"go.uber.org/zap"

	"github.com/gotenberg/gotenberg/v8/pkg/gotenberg"
)

func init() {
	gotenberg.MustRegisterModule(new(PdfToPpm))
}

type PdfToPpm struct {
	binPath string
}

func (engine *PdfToPpm) Descriptor() gotenberg.ModuleDescriptor {
	return gotenberg.ModuleDescriptor{
		ID:  "pdftoppm",
		New: func() gotenberg.Module { return new(PdfToPpm) },
	}
}

func (engine *PdfToPpm) Provision(ctx *gotenberg.Context) error {
	binPath, ok := os.LookupEnv("PDFTOPPM_BIN_PATH")
	if !ok {
		return errors.New("PDFTOPPM_BIN_PATH environment variable is not set")
	}

	engine.binPath = binPath

	return nil
}

func (engine *PdfToPpm) Validate() error {
	_, err := os.Stat(engine.binPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("pdftoppm binary path does not exist: %w", err)
	}

	return nil
}

func (engine *PdfToPpm) Debug() map[string]interface{} {
	debug := make(map[string]interface{})

	cmd := exec.Command(engine.binPath, "-v")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	output, err := cmd.Output()
	if err != nil {
		debug["version"] = err.Error()
		return debug
	}

	debug["version"] = string(output)
	return debug
}

func (engine *PdfToPpm) ConvertToImage(ctx context.Context, logger *zap.Logger, inputPath, outputDirPath string, format string) ([]string, error) {
	dpi := 203
	if envDpi := os.Getenv("PDFTOPPM_DPI"); envDpi != "" {
		if parsedDpi, err := strconv.Atoi(envDpi); err == nil {
			dpi = parsedDpi
		}
	}
	
	antialiasing := "no"
	if envAA := os.Getenv("PDFTOPPM_ANTIALIASING"); envAA != "" {
		antialiasing = envAA
	}
	
	return engine.ConvertToImageWithOptions(ctx, logger, inputPath, outputDirPath, format, dpi, antialiasing)
}

func (engine *PdfToPpm) ConvertToImageWithOptions(ctx context.Context, logger *zap.Logger, inputPath, outputDirPath string, format string, dpi int, antialiasing string) ([]string, error) {
	if format != "png" {
		return nil, fmt.Errorf("convert PDF to '%s' with pdftoppm: %w", format, gotenberg.ErrPdfEngineMethodNotSupported)
	}

	outputPrefix := filepath.Join(outputDirPath, "page")
	
	args := []string{
		"-png",
		"-r", strconv.Itoa(dpi),
		"-aa", antialiasing,
	}
	
	args = append(args, inputPath, outputPrefix)

	cmd, err := gotenberg.CommandContext(ctx, logger, engine.binPath, args...)
	if err != nil {
		return nil, fmt.Errorf("create command: %w", err)
	}

	_, err = cmd.Exec()
	if err != nil {
		return nil, fmt.Errorf("convert PDF to PNG with pdftoppm: %w", err)
	}

	pattern := fmt.Sprintf("%s-*.png", outputPrefix)
	outputPaths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("find output files: %w", err)
	}

	return outputPaths, nil
}

func (engine *PdfToPpm) ImageConverter() gotenberg.ImageConverter {
	return engine
}

var (
	_ gotenberg.Module                 = (*PdfToPpm)(nil)
	_ gotenberg.Provisioner            = (*PdfToPpm)(nil)
	_ gotenberg.Validator              = (*PdfToPpm)(nil)
	_ gotenberg.Debuggable             = (*PdfToPpm)(nil)
	_ gotenberg.ImageConverter         = (*PdfToPpm)(nil)
	_ gotenberg.ImageConverterProvider = (*PdfToPpm)(nil)
)
