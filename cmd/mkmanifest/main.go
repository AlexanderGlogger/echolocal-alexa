package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/ygelfand/echolocal/internal/host/profile"
	"github.com/ygelfand/echolocal/internal/update"
)

func main() {
	var (
		m update.Manifest

		from  = flag.String("from", "", "where this release's assets can be fetched from")
		arm64 = flag.String("arm64", "", "the arm64 build, hashed and measured")
		arm   = flag.String("arm", "", "the arm build, hashed and measured")
		out   = flag.String("out", "", "where to write the manifest, or stdout")
	)
	flag.StringVar(&m.Version, "version", "", "version as Home Assistant will compare it")
	flag.StringVar(&m.Title, "title", "", "title for Home Assistant's update card")
	flag.StringVar(&m.Notes, "notes", "", "release notes, shown on the card")
	flag.StringVar(&m.ReleaseURL, "release-url", "", "what the card's link points at")

	var perBoard boards
	flag.Var(&perBoard, "board", "a build for one board only, as codename:arch:path; repeatable")
	flag.Parse()

	if err := run(
		m,
		*from,
		map[string]string{"arm64": *arm64, "arm": *arm},
		perBoard,
		*out,
	); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type boards []boardBuild

func (b *boards) String() string { return fmt.Sprint([]boardBuild(*b)) }

func (b *boards) Set(v string) error {
	parts := strings.Split(v, ":")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return fmt.Errorf("want codename:arch:path, got %q", v)
	}
	*b = append(*b, boardBuild{Board: parts[0], Arch: parts[1], Path: parts[2]})
	return nil
}

type boardBuild struct {
	Board, Arch, Path string
}

func run(
	m update.Manifest,
	from string,
	builds map[string]string,
	perBoard []boardBuild,
	out string,
) error {
	if m.Version == "" || from == "" || builds["arm64"] == "" {
		return fmt.Errorf("mkmanifest: -version, -from and -arm64 are all required")
	}

	m.Binaries = make(map[string]update.Binary, len(builds))
	for arch, path := range builds {
		if path == "" {
			continue
		}
		b, err := measure(path)
		if err != nil {
			return err
		}
		b.URL = from + "/" + filepath.Base(path)
		m.Binaries[arch] = b
	}

	m.Boards = make(map[string]map[string]update.Binary, len(profile.Boards()))
	for _, b := range profile.Boards() {
		m.Boards[b.Codename] = maps.Clone(m.Binaries)
	}
	own := map[string]map[string]update.Binary{}

	for _, pb := range perBoard {
		if _, ok := m.Boards[pb.Board]; !ok {
			return fmt.Errorf(
				"mkmanifest: -board names %q, which is not a board this build installs onto",
				pb.Board,
			)
		}
		b, err := measure(pb.Path)
		if err != nil {
			return err
		}
		b.URL = from + "/" + filepath.Base(pb.Path)

		if own[pb.Board] == nil {
			own[pb.Board] = map[string]update.Binary{}
			m.Boards[pb.Board] = own[pb.Board]
		}
		own[pb.Board][pb.Arch] = b
	}

	arm64 := m.Binaries["arm64"]
	m.URL, m.SHA256, m.Size = arm64.URL, arm64.SHA256, arm64.Size

	if m.URL == "" || len(m.Binaries) == 0 {
		return fmt.Errorf(
			"mkmanifest: %s carries no shared binaries, which no device already in the field could install",
			m.Version,
		)
	}

	if err := m.Valid(); err != nil {
		return err
	}

	encoded, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')

	if out == "" {
		_, err = os.Stdout.Write(encoded)
		return err
	}
	return os.WriteFile(out, encoded, 0o644)
}

func measure(path string) (update.Binary, error) {
	f, err := os.Open(path)
	if err != nil {
		return update.Binary{}, err
	}
	defer f.Close()

	sum := sha256.New()
	size, err := io.Copy(sum, f)
	if err != nil {
		return update.Binary{}, err
	}
	return update.Binary{SHA256: hex.EncodeToString(sum.Sum(nil)), Size: size}, nil
}
