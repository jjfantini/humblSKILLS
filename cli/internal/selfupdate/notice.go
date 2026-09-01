package selfupdate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// NoticeHeadline is the CLI lead-in when a newer channel release exists.
// The full line is "{NoticeHeadline}: {current} → {latest} ({channel}) — run `{cmd}`".
const NoticeHeadline = "newer version available"

// NoticeCacheTTL is how long a GitHub latest-release snapshot stays fresh
// on disk. Long enough that doctor/start/list in the same day don't each
// hit the API; short enough that a real release shows up the next day.
const NoticeCacheTTL = 12 * time.Hour

// NoticeHTTPTimeout is the fail-fast budget for a background/CLI notice
// check. Upgrade itself keeps DefaultHTTPTimeout; a hung notice must
// never stall a normal command.
const NoticeHTTPTimeout = 3 * time.Second

const (
	noticeCacheFilePrefix = "latest-"
	noticeCacheFileSuffix = ".json"
)

// Notice is a "newer version available" result built from the same
// channel resolution upgrade uses (LatestReleaseForChannel +
// IsUpgradeAvailable + FormulaForVersion + RecommendedUpgradeCommand).
// Available=false means stay quiet — current, failed check, or skipped.
type Notice struct {
	CurrentVersion string
	LatestVersion  string
	LatestTag      string
	Channel        string
	Homebrew       bool
	Formula        string
	CurrentFormula string
	Available      bool
}

// UpdateCommand is the exact command the notice tells the user to run.
// Same helper upgrade --dry-run uses: brew stays on `brew upgrade
// <formula>` unless beta's winner lives on the other formula, in which
// case it is `brew uninstall <old> && brew install <new>`.
func (n Notice) UpdateCommand() string {
	return RecommendedUpgradeCommand(n.Homebrew, n.CurrentFormula, n.Formula)
}

// CLILine is the user-visible CLI notice. Empty when nothing is available.
func (n Notice) CLILine() string {
	if !n.Available {
		return ""
	}
	return fmt.Sprintf("%s: %s → %s (%s) — run `%s`",
		NoticeHeadline,
		DisplayVersion(n.CurrentVersion),
		DisplayVersion(n.LatestVersion),
		NormalizeChannel(n.Channel),
		n.UpdateCommand())
}

// CheckOptions configures Check. Zero-value client/TTL/Now use defaults.
type CheckOptions struct {
	Client         *http.Client
	Repo           string
	CurrentVersion string
	ExePath        string
	Channel        string
	CacheDir       string
	TTL            time.Duration
	Now            func() time.Time
}

type noticeSnapshot struct {
	Channel       string    `json:"channel"`
	LatestVersion string    `json:"latest_version"`
	LatestTag     string    `json:"latest_tag"`
	CheckedAt     time.Time `json:"checked_at"`
}

type noticeMemKey struct {
	cacheDir string
	channel  string
}

var (
	noticeMemMu sync.Mutex
	noticeMem   = map[noticeMemKey]noticeSnapshot{}
)

// NewNoticeHTTPClient is the short-timeout client notice checks use so a
// slow GitHub never blocks doctor/start/list the way upgrade may.
func NewNoticeHTTPClient() *http.Client {
	return &http.Client{Timeout: NoticeHTTPTimeout}
}

// SkipNetwork, when true, makes Check return a quiet Notice without
// touching GitHub. Command-level tests that aren't exercising the notice
// flip this so doctor/list/start don't each wait on the network.
var SkipNetwork bool

// Check looks up the latest release for channel and reports whether the
// running binary is behind it. Failures (network, decode, empty cache
// dir) return a quiet Notice — no error is surfaced to the caller.
//
// Latest-release bytes are cached in memory (once per process + cache
// dir + channel) and on disk under cacheDir/selfupdate/latest-<channel>.json
// for NoticeCacheTTL. The comparison against CurrentVersion is always
// live so a just-upgraded binary does not keep advertising itself.
func Check(opts CheckOptions) Notice {
	channel := NormalizeChannel(opts.Channel)
	n := Notice{
		CurrentVersion: opts.CurrentVersion,
		Channel:        channel,
	}
	if SkipNetwork {
		return n
	}
	if opts.ExePath != "" {
		n.Homebrew = IsHomebrewManaged(opts.ExePath)
		if n.Homebrew {
			n.CurrentFormula = InstalledFormula(opts.ExePath)
		}
	}

	snap, ok := loadNoticeSnapshot(opts, channel)
	if !ok {
		return n
	}
	n.LatestVersion = snap.LatestVersion
	n.LatestTag = snap.LatestTag
	n.Formula = FormulaForVersion(snap.LatestVersion)
	n.Available = IsUpgradeAvailable(opts.CurrentVersion, snap.LatestVersion)
	return n
}

// InvalidateNoticeCache drops the in-memory and on-disk snapshots for
// cacheDir so the next Check refetches. Called after a successful
// upgrade so the notice does not keep advertising the version we just
// installed.
func InvalidateNoticeCache(cacheDir string) {
	if cacheDir == "" {
		return
	}
	noticeMemMu.Lock()
	for k := range noticeMem {
		if k.cacheDir == cacheDir {
			delete(noticeMem, k)
		}
	}
	noticeMemMu.Unlock()
	for _, ch := range []string{ChannelStable, ChannelBeta} {
		_ = os.Remove(noticeCachePath(cacheDir, ch))
	}
}

func loadNoticeSnapshot(opts CheckOptions, channel string) (noticeSnapshot, bool) {
	now := time.Now
	if opts.Now != nil {
		now = opts.Now
	}
	ttl := opts.TTL
	if ttl <= 0 {
		ttl = NoticeCacheTTL
	}

	key := noticeMemKey{cacheDir: opts.CacheDir, channel: channel}
	noticeMemMu.Lock()
	if snap, ok := noticeMem[key]; ok && !noticeSnapshotExpired(snap, now(), ttl) {
		noticeMemMu.Unlock()
		return snap, true
	}
	noticeMemMu.Unlock()

	if disk, ok := readNoticeCache(opts.CacheDir, channel); ok {
		if !noticeSnapshotExpired(disk, now(), ttl) {
			storeNoticeMem(key, disk)
			return disk, true
		}
		// Stale disk is a fallback if the network check fails.
		if snap, ok := fetchNoticeSnapshot(opts, channel, now()); ok {
			writeNoticeCache(opts.CacheDir, snap)
			storeNoticeMem(key, snap)
			return snap, true
		}
		storeNoticeMem(key, disk)
		return disk, true
	}

	snap, ok := fetchNoticeSnapshot(opts, channel, now())
	if !ok {
		return noticeSnapshot{}, false
	}
	writeNoticeCache(opts.CacheDir, snap)
	storeNoticeMem(key, snap)
	return snap, true
}

func fetchNoticeSnapshot(opts CheckOptions, channel string, now time.Time) (noticeSnapshot, bool) {
	client := opts.Client
	if client == nil {
		client = NewNoticeHTTPClient()
	}
	repo := opts.Repo
	if repo == "" {
		repo = DefaultRepo
	}
	rel, err := LatestReleaseForChannel(client, repo, channel)
	if err != nil || rel == nil || rel.TagName == "" {
		return noticeSnapshot{}, false
	}
	return noticeSnapshot{
		Channel:       channel,
		LatestVersion: rel.Version(),
		LatestTag:     rel.TagName,
		CheckedAt:     now.UTC(),
	}, true
}

func storeNoticeMem(key noticeMemKey, snap noticeSnapshot) {
	noticeMemMu.Lock()
	defer noticeMemMu.Unlock()
	noticeMem[key] = snap
}

func noticeSnapshotExpired(snap noticeSnapshot, now time.Time, ttl time.Duration) bool {
	if snap.CheckedAt.IsZero() || snap.LatestVersion == "" {
		return true
	}
	return now.Sub(snap.CheckedAt) >= ttl
}

func noticeCachePath(cacheDir, channel string) string {
	return filepath.Join(cacheDir, "selfupdate", noticeCacheFilePrefix+NormalizeChannel(channel)+noticeCacheFileSuffix)
}

func readNoticeCache(cacheDir, channel string) (noticeSnapshot, bool) {
	if cacheDir == "" {
		return noticeSnapshot{}, false
	}
	data, err := os.ReadFile(noticeCachePath(cacheDir, channel))
	if err != nil {
		return noticeSnapshot{}, false
	}
	var snap noticeSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return noticeSnapshot{}, false
	}
	if snap.LatestVersion == "" {
		return noticeSnapshot{}, false
	}
	snap.Channel = NormalizeChannel(snap.Channel)
	return snap, true
}

func writeNoticeCache(cacheDir string, snap noticeSnapshot) {
	if cacheDir == "" || snap.LatestVersion == "" {
		return
	}
	path := noticeCachePath(cacheDir, snap.Channel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	data, err := json.Marshal(snap)
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}
