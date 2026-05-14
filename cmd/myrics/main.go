package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "myrics",
		Short: "Myrics dev CLI — build, install, and run the Myrics overlay",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	root.AddCommand(
		cmdInstallDeps(),
		cmdConfig(),
		cmdBuild(),
		cmdInstall(),
		cmdRun(),
		cmdDev(),
		cmdTest(),
		cmdClean(),
		cmdUninstall(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// ── commands ──────────────────────────────────────────────────────────────────

func cmdInstallDeps() *cobra.Command {
	return &cobra.Command{
		Use:   "install-deps",
		Short: "Install wails CLI and download Go module dependencies",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := run("go", "install", "github.com/wailsapp/wails/v2/cmd/wails@latest"); err != nil {
				return err
			}
			if err := run("go", "mod", "download"); err != nil {
				return err
			}
			fmt.Println()
			fmt.Println("NOTE: also ensure these are installed:")
			fmt.Println("  Windows: WebView2 runtime (pre-installed on Win10/11)")
			fmt.Println("  Windows: C compiler — MinGW-w64 (winget install MSYS2.MSYS2)")
			fmt.Println("  Linux:   portaudio-devel, webkit2gtk4.1-devel")
			fmt.Println()
			fmt.Println("Done. Next: myrics config && myrics install")
			return nil
		},
	}
}

func cmdConfig() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Copy config.yaml.example → configs/config.yaml (skips if already exists)",
		RunE: func(_ *cobra.Command, _ []string) error {
			dst := filepath.Join("configs", "config.yaml")
			if _, err := os.Stat(dst); err == nil {
				fmt.Println("configs/config.yaml already exists — skipping.")
				return nil
			}
			if err := copyFile(filepath.Join("configs", "config.yaml.example"), dst); err != nil {
				return err
			}
			fmt.Println("Created configs/config.yaml — edit it if you need ACRCloud keys.")
			return nil
		},
	}
}

func cmdBuild() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Tidy modules and build the Myrics binary via wails build",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := run("go", "mod", "tidy"); err != nil {
				return err
			}
			return run("wails", "build", "-clean")
		},
	}
}

func cmdInstall() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Build and install myrics-overlay to GOPATH/bin",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := run("go", "mod", "tidy"); err != nil {
				return err
			}
			if err := run("wails", "build", "-clean"); err != nil {
				return err
			}

			src := buildOutputPath()
			if _, err := os.Stat(src); os.IsNotExist(err) {
				return fmt.Errorf("build output not found at %s", src)
			}

			dst := overlayInstallPath()
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return err
			}
			if err := copyFile(src, dst); err != nil {
				return err
			}
			fmt.Printf("Installed to %s\n", dst)
			return nil
		},
	}
}

func cmdRun() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Launch the installed myrics-overlay binary",
		RunE: func(_ *cobra.Command, _ []string) error {
			p := overlayInstallPath()
			if _, err := os.Stat(p); os.IsNotExist(err) {
				return fmt.Errorf("not installed — run 'myrics install' first")
			}
			return run(p)
		},
	}
}

func cmdDev() *cobra.Command {
	return &cobra.Command{
		Use:   "dev",
		Short: "Run in development mode with hot reload (wails dev)",
		RunE: func(_ *cobra.Command, _ []string) error {
			return run("wails", "dev")
		},
	}
}

func cmdTest() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Run Go unit tests",
		RunE: func(_ *cobra.Command, _ []string) error {
			return run("go", "test", "./internal/...", "-v")
		},
	}
}

func cmdClean() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Remove build/ and frontend/dist/",
		RunE: func(_ *cobra.Command, _ []string) error {
			for _, dir := range []string{"build", filepath.Join("frontend", "dist")} {
				if err := os.RemoveAll(dir); err != nil {
					return err
				}
				fmt.Printf("Removed %s\n", dir)
			}
			return nil
		},
	}
}

func cmdUninstall() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove myrics-overlay from GOPATH/bin",
		RunE: func(_ *cobra.Command, _ []string) error {
			p := overlayInstallPath()
			if _, err := os.Stat(p); os.IsNotExist(err) {
				fmt.Printf("Nothing to remove at %s\n", p)
				return nil
			}
			if err := os.Remove(p); err != nil {
				return err
			}
			fmt.Printf("Removed %s\n", p)
			return nil
		},
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func gobin() string {
	if gp := os.Getenv("GOPATH"); gp != "" {
		return filepath.Join(gp, "bin")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "go", "bin")
}

func exeName(base string) string {
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

func buildOutputPath() string {
	return filepath.Join("build", "bin", exeName("myrics-overlay"))
}

func overlayInstallPath() string {
	return filepath.Join(gobin(), exeName("myrics-overlay"))
}
