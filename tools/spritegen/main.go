// Command spritegen converts text-grid sprite sources (*.txt) into PNG
// sheets plus a JSON frame index (DESIGN §8). Run via go generate in assets.
//
//	spritegen -in sprites -out gen
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/comprofix/alleycat/internal/spritegrid"
)

func main() {
	in := flag.String("in", "sprites", "directory of *.txt sprite sources")
	out := flag.String("out", "gen", "output directory for *.png and *.json")
	flag.Parse()
	if err := run(*in, *out); err != nil {
		fmt.Fprintln(os.Stderr, "spritegen:", err)
		os.Exit(1)
	}
}

func run(inDir, outDir string) error {
	srcs, err := filepath.Glob(filepath.Join(inDir, "*.txt"))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	for _, src := range srcs {
		if err := convert(src, outDir); err != nil {
			return fmt.Errorf("%s: %w", src, err)
		}
	}
	return nil
}

func convert(src, outDir string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	sprites, err := spritegrid.Parse(f)
	if err != nil {
		return err
	}
	img, index, err := spritegrid.Pack(sprites)
	if err != nil {
		return err
	}

	base := filepath.Join(outDir, strings.TrimSuffix(filepath.Base(src), ".txt"))
	pf, err := os.Create(base + ".png")
	if err != nil {
		return err
	}
	if err := png.Encode(pf, img); err != nil {
		pf.Close()
		return err
	}
	if err := pf.Close(); err != nil {
		return err
	}

	// Stable key order keeps the JSON diff-friendly.
	names := make([]string, 0, len(index))
	for n := range index {
		names = append(names, n)
	}
	sort.Strings(names)
	ordered := make(map[string]spritegrid.Entry, len(index))
	for _, n := range names {
		ordered[n] = index[n]
	}
	data, err := json.MarshalIndent(ordered, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(base+".json", append(data, '\n'), 0o644)
}
