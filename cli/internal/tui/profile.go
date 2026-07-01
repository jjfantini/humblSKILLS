package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// ProfileHeaderSpec lets the caller override the profile editor's top header
// so sub-screens rendered from the dashboard can show a consistent breadcrumb
// + status line instead of a bare "Profile" crumb.
type ProfileHeaderSpec struct {
	Section string // e.g. "Profile" or "Dashboard > Profile"
	Meta    string // right-anchored meta — typically tui.RenderStatusMeta(...)
}

// RunProfileEditor opens a two-pane bubbletea TUI for editing the user
// profile. The left pane lists settings; pressing enter on a row moves focus
// into the right pane so the user can toggle checkboxes (platforms) or pick
// an option (scope) inline. No sub-forms pop on top of the screen.
//
// Returns the edited profile and whether any change was made. Save is the
// caller's responsibility — this function only mutates the returned value.
func RunProfileEditor(theme *ui.Theme, adapterList []adapters.Adapter, p *profile.Profile) (*profile.Profile, bool, error) {
	return RunProfileEditorWith(theme, adapterList, p, ProfileHeaderSpec{})
}

// RunProfileEditorWith is RunProfileEditor plus a header override. A zero
// ProfileHeaderSpec falls back to the default "Profile" crumb.
func RunProfileEditorWith(theme *ui.Theme, adapterList []adapters.Adapter, p *profile.Profile, h ProfileHeaderSpec) (*profile.Profile, bool, error) {
	if p == nil {
		p = &profile.Profile{SchemaVersion: profile.SchemaVersion}
	}
	initial := *p

	section := h.Section
	if section == "" {
		section = "Profile"
	}

	m := profileModel{
		theme:      theme,
		version:    versionString,
		adapters:   adapterList,
		profile:    initial,
		focus:      focusSettings,
		settingIdx: 0,
		section:    section,
		meta:       h.Meta,
	}
	out, err := Run(m)
	if err != nil {
		return &initial, false, err
	}
	final, ok := out.(profileModel)
	if !ok {
		return &initial, false, nil
	}
	return &final.profile, final.changed, nil
}

// versionString is stamped in at link time via -ldflags; an empty value
// makes the header drop the version tag, which is fine for this editor.
var versionString = ""

// SetProfileEditorVersion lets the caller inject the CLI version string so
// the profile TUI header shows `humblskills v2.1.0 · Profile`. Call once
// from main/root setup; safe no-op if left empty.
func SetProfileEditorVersion(v string) { versionString = v }

// --- model ------------------------------------------------------------------

type profileFocus int

const (
	focusSettings profileFocus = iota
	focusValue
)

type profileModel struct {
	theme    *ui.Theme
	version  string
	adapters []adapters.Adapter

	profile profile.Profile
	changed bool

	focus      profileFocus
	settingIdx int
	valueIdx   int

	section string // header crumb (defaults to "Profile")
	meta    string // right-anchored header text (typically the status line)

	width, height int
	quit          bool
}

func (m profileModel) Init() tea.Cmd { return nil }

func (m profileModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.focus == focusSettings {
			return m.updateSettings(msg)
		}
		return m.updateValue(msg)
	}
	return m, nil
}

func (m profileModel) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		m.quit = true
		return m, tea.Quit
	case "up", "k":
		if m.settingIdx > 0 {
			m.settingIdx--
		}
	case "down", "j":
		if m.settingIdx < len(profileSettings)-1 {
			m.settingIdx++
		}
	case "enter", "e", "l", "right", "tab":
		m.focus = focusValue
		m.valueIdx = m.currentSelectionIndex()
	}
	return m, nil
}

func (m profileModel) updateValue(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	n := m.valueCount()
	switch msg.String() {
	case "esc", "h", "left", "shift+tab":
		m.focus = focusSettings
		return m, nil
	case "q", "ctrl+c":
		m.quit = true
		return m, tea.Quit
	case "up", "k":
		if m.valueIdx > 0 {
			m.valueIdx--
		}
	case "down", "j":
		if m.valueIdx < n-1 {
			m.valueIdx++
		}
	case " ", "x":
		m = m.toggleCurrent()
	case "enter":
		// Space is the only toggle key, consistent with every other
		// multi-select surface in the TUI (e.g. the install platform
		// modal). For radio-style settings (scope), enter commits the
		// highlighted option and returns to the settings pane; for
		// multi-select (platforms) it never toggles — it just returns,
		// same as esc.
		if profileSettings[m.settingIdx].kind == settingRadio {
			m = m.toggleCurrent()
		}
		m.focus = focusSettings
	}
	return m, nil
}

func (m profileModel) toggleCurrent() profileModel {
	switch m.settingIdx {
	case 0: // platforms
		if m.valueIdx >= len(m.adapters) {
			return m
		}
		name := m.adapters[m.valueIdx].Name
		idx := -1
		for i, p := range m.profile.DefaultPlatforms {
			if p == name {
				idx = i
				break
			}
		}
		if idx >= 0 {
			m.profile.DefaultPlatforms = append(
				m.profile.DefaultPlatforms[:idx],
				m.profile.DefaultPlatforms[idx+1:]...,
			)
		} else {
			m.profile.DefaultPlatforms = append(m.profile.DefaultPlatforms, name)
		}
		m.changed = true
	case 1: // scope
		if m.valueIdx >= 0 && m.valueIdx < len(scopeSettingOpts) {
			m.profile.DefaultScope = scopeSettingOpts[m.valueIdx].value
			m.changed = true
		}
	}
	return m
}

// scopeSettingOpts is the profile's full scope picker — unlike the
// per-install modal (3 concrete choices only), the profile also offers
// "adapter default" as an explicit, deliberate opt-in: it can't show a
// concrete location at install time, so it only belongs here, not in the
// interactive install flow.
var scopeSettingOpts = []struct{ label, value string }{
	{"global humblskills (recommended)", profile.ScopeGlobal},
	{"user", profile.ScopeUser},
	{"project", profile.ScopeProject},
	{"adapter default", profile.ScopeAdapterDefault},
}

// currentSelectionIndex returns the index in the right-pane options list that
// represents the profile's current value for the focused setting. Used to
// place the cursor on the already-selected option when the user drills in.
func (m profileModel) currentSelectionIndex() int {
	switch m.settingIdx {
	case 0:
		return 0
	case 1:
		resolved := m.profile.ResolvedScope()
		for i, opt := range scopeSettingOpts {
			if opt.value == resolved {
				return i
			}
		}
	}
	return 0
}

func (m profileModel) valueCount() int {
	switch m.settingIdx {
	case 0:
		return len(m.adapters)
	case 1:
		return len(scopeSettingOpts)
	}
	return 0
}

// --- view -------------------------------------------------------------------

type settingKind int

const (
	settingMulti settingKind = iota
	settingRadio
)

type profileSetting struct {
	key   string
	label string
	kind  settingKind
}

var profileSettings = []profileSetting{
	{key: "platforms", label: "default platforms", kind: settingMulti},
	{key: "scope", label: "default scope", kind: settingRadio},
}

func (m profileModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	th := m.theme

	section := m.section
	if section == "" {
		section = "Profile"
	}
	header := Header(th, HeaderSpec{
		Version: m.version,
		Section: section,
		Meta:    m.meta,
	}, m.width)

	leftW, rightW := m.paneWidths()

	left := m.renderLeft(leftW)
	right := m.renderRight(rightW)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	focused := th.Meta.Render("focused: ") +
		th.Brand.Render(profileSettings[m.settingIdx].label)
	footer := Footer(th, m.hints(), focused, m.width)

	bodyH := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 1
	if bodyH < 5 {
		bodyH = 5
	}
	return Frame(header, body, footer, bodyH)
}

func (m profileModel) paneWidths() (int, int) {
	left := 32
	if cap := m.width / 3; cap > 0 && left > cap {
		left = cap
	}
	if left < 22 {
		left = 22
	}
	right := m.width - left
	if right < 20 {
		right = 20
	}
	return left, right
}

func (m profileModel) renderLeft(width int) string {
	th := m.theme
	pad := func(s string) string {
		w := lipgloss.Width(s)
		if w >= width {
			return s
		}
		return s + strings.Repeat(" ", width-w)
	}

	title := "  " + th.SectionTitle.Render("SETTINGS")
	var sb strings.Builder
	sb.WriteString(pad(title))
	sb.WriteString("\n\n")

	for i, s := range profileSettings {
		selected := i == m.settingIdx && m.focus == focusSettings
		cursorHere := i == m.settingIdx
		dot := th.DotOK.Render("●")
		if m.settingValueEmpty(s.key) {
			dot = th.DotNo.Render("●")
		}
		name := th.RowUnselected.Render(s.label)
		if selected {
			name = th.RowSelected.Render(s.label)
		} else if cursorHere && m.focus == focusValue {
			// cursor still anchored here, but pane isn't focused — use a
			// muted selected style so the user can see what they drilled in on.
			name = th.RowDim.Render(s.label)
		}
		badge := Badge(th, BadgeGhost, m.settingBadge(s.key))

		row := m.layoutRow(dot+" "+name, badge, width-2)

		var line string
		if cursorHere && m.focus == focusSettings {
			line = th.Bullet.Render("▌") + " " + row
		} else {
			line = "  " + row
		}
		sb.WriteString(pad(line))
		if i < len(profileSettings)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (m profileModel) renderRight(width int) string {
	th := m.theme
	bar := th.Divider.Render("│")
	title := th.SectionTitle.Render("VALUE")

	var body []string
	body = append(body, title)
	body = append(body, bar)

	setting := profileSettings[m.settingIdx]
	body = append(body, bar+" "+th.DetailTitle.Render(setting.label))

	var hint string
	if m.focus == focusSettings {
		hint = "press enter to edit"
	} else if setting.kind == settingMulti {
		hint = "space to toggle · enter/esc to return"
	} else {
		hint = "enter to select · esc to return"
	}
	body = append(body, bar+" "+th.DetailSub.Render(hint))
	body = append(body, bar)

	switch m.settingIdx {
	case 0:
		body = append(body, m.renderPlatformOptions(bar, width)...)
	case 1:
		body = append(body, m.renderScopeOptions(bar, width)...)
	}
	return strings.Join(body, "\n")
}

func (m profileModel) renderPlatformOptions(bar string, width int) []string {
	th := m.theme
	selected := map[string]bool{}
	for _, n := range m.profile.DefaultPlatforms {
		selected[n] = true
	}
	rows := make([]string, 0, len(m.adapters)+2)
	if len(m.profile.DefaultPlatforms) == 0 {
		rows = append(rows, bar+" "+th.Detail.Render(
			"No defaults set — installs target every detected platform."))
		rows = append(rows, bar)
	}
	rows = append(rows, bar+" "+th.SectionTitle.Render("PLATFORMS"))
	for i, a := range m.adapters {
		cursorHere := i == m.valueIdx && m.focus == focusValue
		box := "[ ]"
		label := a.Name
		if selected[a.Name] {
			box = "[✓]"
		}
		var styled string
		switch {
		case cursorHere:
			styled = th.RowSelected.Render(box + "  " + label)
		case selected[a.Name]:
			styled = th.Success.Render(box) + "  " + th.RowUnselected.Render(label)
		default:
			styled = th.RowDim.Render(box) + "  " + th.RowUnselected.Render(label)
		}
		prefix := bar + "   "
		if cursorHere {
			prefix = bar + " " + th.Bullet.Render("▸") + " "
		}
		rows = append(rows, prefix+styled)
	}
	_ = width
	return rows
}

func (m profileModel) renderScopeOptions(bar string, width int) []string {
	th := m.theme
	resolved := m.profile.ResolvedScope()
	rows := make([]string, 0, len(scopeSettingOpts)+4)
	rows = append(rows, bar+" "+th.Detail.Render(
		"Which scope installs default to. Global humblskills installs one canonical "+
			"copy and symlinks it to every selected platform — recommended for most "+
			"setups. User/project pin a concrete platform-native location instead."))
	rows = append(rows, bar)
	rows = append(rows, bar+" "+th.SectionTitle.Render("OPTIONS"))
	for i, opt := range scopeSettingOpts {
		cursorHere := i == m.valueIdx && m.focus == focusValue
		isCurrent := opt.value == resolved
		marker := "( )"
		if isCurrent {
			marker = "(●)"
		}
		var styled string
		switch {
		case cursorHere:
			styled = th.RowSelected.Render(marker + "  " + opt.label)
		case isCurrent:
			styled = th.Success.Render(marker) + "  " + th.RowUnselected.Render(opt.label)
		default:
			styled = th.RowDim.Render(marker) + "  " + th.RowUnselected.Render(opt.label)
		}
		prefix := bar + "   "
		if cursorHere {
			prefix = bar + " " + th.Bullet.Render("▸") + " "
		}
		rows = append(rows, prefix+styled)
	}
	if resolved == profile.ScopeAdapterDefault {
		rows = append(rows, bar)
		rows = append(rows, bar+" "+th.Detail.Render(
			"Note: adapter default can't show a concrete location up front — "+
				"every platform here happens to default to \"user\" today, but the "+
				"install screen will still ask you to pick a scope since it can't "+
				"display one for this setting."))
	}
	_ = width
	return rows
}

func (m profileModel) layoutRow(label, badge string, width int) string {
	lw := lipgloss.Width(label)
	bw := lipgloss.Width(badge)
	if width < 10 || width-lw < bw+1 {
		if lw >= width {
			return label
		}
		return label + strings.Repeat(" ", width-lw)
	}
	gap := width - lw - bw
	return label + strings.Repeat(" ", gap) + badge
}

func (m profileModel) settingBadge(key string) string {
	switch key {
	case "platforms":
		if len(m.profile.DefaultPlatforms) == 0 {
			return "all detected"
		}
		return fmt.Sprintf("%d platform%s", len(m.profile.DefaultPlatforms),
			plural2(len(m.profile.DefaultPlatforms)))
	case "scope":
		switch resolved := m.profile.ResolvedScope(); resolved {
		case profile.ScopeGlobal:
			return "global humblskills"
		case profile.ScopeAdapterDefault:
			return "adapter default"
		default:
			return resolved
		}
	}
	return ""
}

// settingValueEmpty reports whether a setting has no meaningful value yet.
// Scope is never "empty" — an unset DefaultScope already resolves to a
// concrete, deliberate default (global humblskills) via ResolvedScope.
func (m profileModel) settingValueEmpty(key string) bool {
	switch key {
	case "platforms":
		return len(m.profile.DefaultPlatforms) == 0
	}
	return false
}

func (m profileModel) hints() []KeyHint {
	if m.focus == focusSettings {
		return []KeyHint{
			{Keys: "↑↓", Label: "select"},
			{Keys: "enter", Label: "edit"},
			{Keys: "q", Label: "quit"},
		}
	}
	setting := profileSettings[m.settingIdx]
	if setting.kind == settingMulti {
		return []KeyHint{
			{Keys: "↑↓", Label: "select"},
			{Keys: "space", Label: "toggle"},
			{Keys: "esc", Label: "back"},
			{Keys: "q", Label: "quit"},
		}
	}
	return []KeyHint{
		{Keys: "↑↓", Label: "select"},
		{Keys: "enter", Label: "choose"},
		{Keys: "esc", Label: "back"},
		{Keys: "q", Label: "quit"},
	}
}

func plural2(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// --- per-install modal (moved) ----------------------------------------------

// InstallModalResult is what the install platform picker returns. The modal
// implementation lives in install_modal.go (hand-rolled bubbletea model).
type InstallModalResult struct {
	Platforms   []string
	Scope       string
	Global      bool
	Confirmed   bool
	EditProfile bool
}
