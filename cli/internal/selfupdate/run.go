package selfupdate

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

// Phase names one moment in an upgrade run the UI layer might want to
// reflect (progress lines, spinner labels). Mirrors the shape of
// internal/install's Phase/Event/EventSink.
type Phase string

const (
	PhaseCheckingLatest Phase = "checking_latest"
	PhaseDownloading    Phase = "downloading"
	PhaseVerifyingSum   Phase = "verifying_checksum"
	PhaseInstalling     Phase = "installing"
	// PhaseBrewUpdating fires while refreshing Homebrew's local tap
	// metadata (`brew update`), before PhaseBrewUpgrading or a formula switch.
	PhaseBrewUpdating Phase = "brew_updating"
	// PhaseBrewUpgrading fires while running `brew upgrade <formula>`.
	PhaseBrewUpgrading Phase = "brew_upgrading"
	// PhaseBrewUninstalling / PhaseBrewInstalling fire when beta's winning
	// version lives on the other formula (pre → stable or stable → pre).
	PhaseBrewUninstalling Phase = "brew_uninstalling"
	PhaseBrewInstalling   Phase = "brew_installing"
	PhaseError            Phase = "error"
)

// Event is a single progress notification emitted while running
// ResolvePlan/Apply.
type Event struct {
	Phase Phase
	Msg   string
	Err   error
}

// EventSink receives progress events. Callers that don't care about
// progress (tests, scripts, --json consumers) can pass nil.
type EventSink func(Event)

// emit is a nil-safe helper so call-sites don't have to guard.
func (s EventSink) emit(ev Event) {
	if s != nil {
		s(ev)
	}
}

// Plan is what fetching the latest release and comparing versions resolves
// to, before any network download or filesystem change happens. Both
// `--dry-run` and the real upgrade path build the same Plan so the two can
// never disagree about what would happen.
type Plan struct {
	CurrentVersion   string
	LatestVersion    string // bare, no "v" prefix, e.g. "2.17.0"
	LatestTag        string // as published, e.g. "v2.17.0"
	UpgradeAvailable bool
	Homebrew         bool
	Channel          string
	Formula          string // target brew formula when Homebrew; empty otherwise
	CurrentFormula   string // installed brew formula when detected
	BrewHint         string // exact brew (or CLI) command notices/dry-run print
	SwitchFormula    bool
	AssetName        string
	AssetURL         string
	ChecksumsURL     string
}

// ResolvePlan fetches the latest *stable* release and decides whether an
// upgrade is available. Equivalent to ResolvePlanForChannel(..., ChannelStable).
func ResolvePlan(client *http.Client, repo, currentVersion, exePath string, sink EventSink) (*Plan, error) {
	return ResolvePlanForChannel(client, repo, currentVersion, exePath, ChannelStable, sink)
}

// ResolvePlanForChannel is ResolvePlan using LatestReleaseForChannel so
// beta picks the higher of latest stable vs latest prerelease. The brew
// formula follows that winner, not the channel name.
func ResolvePlanForChannel(client *http.Client, repo, currentVersion, exePath, channel string, sink EventSink) (*Plan, error) {
	if repo == "" {
		repo = DefaultRepo
	}
	channel = NormalizeChannel(channel)
	sink.emit(Event{Phase: PhaseCheckingLatest})
	rel, err := LatestReleaseForChannel(client, repo, channel)
	if err != nil {
		sink.emit(Event{Phase: PhaseError, Err: err})
		return nil, err
	}

	latest := rel.Version()
	homebrew := IsHomebrewManaged(exePath)
	target := FormulaForRelease(rel)
	currentFormula := ""
	if homebrew {
		currentFormula = InstalledFormula(exePath)
	}
	action := PlanBrewAction(currentFormula, target)
	plan := &Plan{
		CurrentVersion:   currentVersion,
		LatestVersion:    latest,
		LatestTag:        rel.TagName,
		UpgradeAvailable: IsUpgradeAvailable(currentVersion, latest),
		Homebrew:         homebrew,
		Channel:          channel,
		BrewHint:         RecommendedUpgradeCommand(homebrew, currentFormula, target),
	}
	if homebrew {
		plan.Formula = target
		plan.CurrentFormula = currentFormula
		plan.SwitchFormula = action.NeedsSwitch()
	}
	if !plan.UpgradeAvailable {
		return plan, nil
	}

	assetName, err := CurrentAssetName(latest)
	if err != nil {
		sink.emit(Event{Phase: PhaseError, Err: err})
		return nil, err
	}
	asset, ok := rel.Asset(assetName)
	if !ok {
		err := fmt.Errorf("release %s has no asset named %s", rel.TagName, assetName)
		sink.emit(Event{Phase: PhaseError, Err: err})
		return nil, err
	}
	checksums, ok := rel.Asset(ChecksumsAssetName)
	if !ok {
		err := fmt.Errorf("release %s has no %s asset", rel.TagName, ChecksumsAssetName)
		sink.emit(Event{Phase: PhaseError, Err: err})
		return nil, err
	}

	plan.AssetName = assetName
	plan.AssetURL = asset.BrowserDownloadURL
	plan.ChecksumsURL = checksums.BrowserDownloadURL
	return plan, nil
}

// Apply performs the download/verify/extract/swap pipeline for a
// non-Homebrew install. cacheDir is a scratch directory for the downloaded
// archive and extracted binary (typically the CLI's own cache dir);
// exePath is the currently running executable's path, which is also the
// swap target.
func Apply(client *http.Client, plan *Plan, cacheDir, exePath string, sink EventSink) error {
	if client == nil {
		client = NewHTTPClient()
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	sink.emit(Event{Phase: PhaseDownloading, Msg: plan.AssetName})
	archivePath := filepath.Join(cacheDir, plan.AssetName)
	if err := downloadToFile(client, plan.AssetURL, archivePath); err != nil {
		sink.emit(Event{Phase: PhaseError, Err: err})
		return err
	}

	sink.emit(Event{Phase: PhaseVerifyingSum})
	sums, err := downloadChecksums(client, plan.ChecksumsURL)
	if err != nil {
		sink.emit(Event{Phase: PhaseError, Err: err})
		return err
	}
	if err := VerifyChecksum(archivePath, plan.AssetName, sums); err != nil {
		sink.emit(Event{Phase: PhaseError, Err: err})
		return err
	}

	sink.emit(Event{Phase: PhaseInstalling})
	binName := BinaryName(runtime.GOOS)
	extractedPath := filepath.Join(cacheDir, binName+".new")
	if err := ExtractBinary(archivePath, binName, extractedPath); err != nil {
		sink.emit(Event{Phase: PhaseError, Err: err})
		return err
	}
	if err := ReplaceBinary(exePath, extractedPath); err != nil {
		sink.emit(Event{Phase: PhaseError, Err: err})
		return err
	}
	return nil
}
