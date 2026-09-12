package darkstarW

import (
	"LumenForge/src/buttonspresentation"
	"LumenForge/src/deviceprofilepresentation"
	"LumenForge/src/dpipresentation"
	"LumenForge/src/mousegesturepresentation"
	"LumenForge/src/performancepresentation"
	"LumenForge/src/rgb"
	"LumenForge/src/sleeptimerpresentation"
	"fmt"
	"sort"
	"strconv"
)

var darkstarWVisibleButtonOrder = []int{262144, 131072, 65536, 32768, 16384, 8192, 4096, 2048, 1024, 512, 256, 128, 64, 32, 16, 8, 4, 2, 1}

func (d *Device) DPIDeviceID() string {
	if d == nil {
		return ""
	}
	return d.Serial
}
func (d *Device) ButtonsDeviceID() string {
	if d == nil {
		return ""
	}
	return d.Serial
}
func (d *Device) PerformanceDeviceID() string {
	if d == nil {
		return ""
	}
	return d.Serial
}
func (d *Device) DeviceProfileDeviceID() string {
	if d == nil {
		return ""
	}
	return d.Serial
}
func (d *Device) SleepTimerDeviceID() string {
	if d == nil {
		return ""
	}
	return d.Serial
}
func (d *Device) MouseGesturesDeviceID() string {
	if d == nil {
		return ""
	}
	return d.Serial
}
func darkstarWColorOK(c rgb.Color) bool {
	return c.Red >= 0 && c.Red <= 255 && c.Green >= 0 && c.Green <= 255 && c.Blue >= 0 && c.Blue <= 255
}
func (d *Device) DPISnapshot() (dpipresentation.Snapshot, bool) {
	if d == nil || d.DeviceProfile == nil || d.MinDPI < 1 || d.MaxDPI < d.MinDPI || d.DPIAmount < 1 || len(d.DeviceProfile.Profiles) != d.DPIAmount+1 || d.DeviceProfile.DPIColor == nil || !darkstarWColorOK(*d.DeviceProfile.DPIColor) {
		return dpipresentation.Snapshot{}, false
	}
	keys := make([]int, 0, len(d.DeviceProfile.Profiles))
	for key := range d.DeviceProfile.Profiles {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	s := dpipresentation.Snapshot{MinimumDPI: d.MinDPI, MaximumDPI: d.MaxDPI, Stages: make([]dpipresentation.Stage, 0, len(keys))}
	sniper := false
	for _, key := range keys {
		p := d.DeviceProfile.Profiles[key]
		if p.Name == "" || p.Value < uint16(d.MinDPI) || p.Value > uint16(d.MaxDPI) {
			return dpipresentation.Snapshot{}, false
		}
		active := !p.Sniper && key == d.DeviceProfile.Profile
		if active {
			s.ActiveRegularStageID = strconv.Itoa(key)
		}
		sniper = sniper || p.Sniper
		s.Stages = append(s.Stages, dpipresentation.Stage{ID: strconv.Itoa(key), Name: p.Name, DPI: p.Value, ColorHex: fmt.Sprintf("#%02x%02x%02x", uint8(d.DeviceProfile.DPIColor.Red), uint8(d.DeviceProfile.DPIColor.Green), uint8(d.DeviceProfile.DPIColor.Blue)), Sniper: p.Sniper, Active: active || (p.Sniper && d.SniperMode)})
	}
	return s, sniper && s.ActiveRegularStageID != ""
}
func (d *Device) ButtonsSnapshot() (buttonspresentation.Snapshot, bool) {
	if d == nil || len(d.KeyAssignmentTypes) == 0 {
		return buttonspresentation.Snapshot{}, false
	}
	s := buttonspresentation.Snapshot{Buttons: make([]buttonspresentation.Button, 0, len(darkstarWVisibleButtonOrder))}
	for _, key := range darkstarWVisibleButtonOrder {
		a, ok := d.KeyAssignment[key]
		if !ok || a.Name == "" {
			return buttonspresentation.Snapshot{}, false
		}
		s.Buttons = append(s.Buttons, buttonspresentation.Button{KeyIndex: key, Name: a.Name, Default: a.Default, PressAndHold: a.ActionHold, OnRelease: a.OnRelease, ActionType: a.ActionType, ActionCommand: a.ActionCommand, IsMacro: a.IsMacro, ProfileSwitch: a.ProfileSwitch})
	}
	keys := make([]int, 0, len(d.KeyAssignmentTypes))
	for key := range d.KeyAssignmentTypes {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	for _, key := range keys {
		if d.KeyAssignmentTypes[key] == "" {
			return buttonspresentation.Snapshot{}, false
		}
		s.AssignmentTypes = append(s.AssignmentTypes, buttonspresentation.AssignmentType{ID: uint8(key), Label: d.KeyAssignmentTypes[key]})
	}
	return s, true
}
func darkstarWOptions(options map[int]string) []performancepresentation.Option {
	if len(options) == 0 {
		return nil
	}
	keys := make([]int, 0, len(options))
	for key := range options {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	out := make([]performancepresentation.Option, 0, len(keys))
	for _, key := range keys {
		if options[key] == "" {
			return nil
		}
		out = append(out, performancepresentation.Option{Value: key, Label: options[key]})
	}
	return out
}
func (d *Device) PerformanceSnapshot() (performancepresentation.Snapshot, bool) {
	if d == nil || d.DeviceProfile == nil || d.DeviceProfile.AngleSnapping < 0 || d.DeviceProfile.AngleSnapping > 1 {
		return performancepresentation.Snapshot{}, false
	}
	options := darkstarWOptions(d.SwitchModes)
	if options == nil {
		return performancepresentation.Snapshot{}, false
	}
	if _, ok := d.SwitchModes[d.DeviceProfile.ButtonOptimization]; !ok {
		return performancepresentation.Snapshot{}, false
	}
	lift := darkstarWOptions(d.LiftHeights)
	if lift == nil {
		return performancepresentation.Snapshot{}, false
	}
	if _, ok := d.LiftHeights[d.DeviceProfile.LiftHeight]; !ok {
		return performancepresentation.Snapshot{}, false
	}
	return performancepresentation.Snapshot{AngleSnapping: &performancepresentation.ToggleSetting{Enabled: d.DeviceProfile.AngleSnapping == 1}, ButtonOptimization: &performancepresentation.SelectSetting{Value: d.DeviceProfile.ButtonOptimization, Options: options}, LiftHeight: &performancepresentation.SelectSetting{Value: d.DeviceProfile.LiftHeight, Options: lift}}, true
}
func (d *Device) DeviceProfileSnapshot() (deviceprofilepresentation.Snapshot, bool) {
	if d == nil || len(d.UserProfiles) == 0 {
		return deviceprofilepresentation.Snapshot{}, false
	}
	s := deviceprofilepresentation.Snapshot{Supported: true}
	for name, p := range d.UserProfiles {
		if name == "" || p == nil {
			return deviceprofilepresentation.Snapshot{}, false
		}
		s.Profiles = append(s.Profiles, name)
		if p.Active {
			if s.ActiveProfile != "" {
				return deviceprofilepresentation.Snapshot{}, false
			}
			s.ActiveProfile = name
		}
	}
	sort.Strings(s.Profiles)
	s = deviceprofilepresentation.WithMutationCapabilities(s, d)
	return s, s.ActiveProfile != ""
}
func (d *Device) SleepTimerSnapshot() (sleeptimerpresentation.Snapshot, bool) {
	if d == nil || d.DeviceProfile == nil {
		return sleeptimerpresentation.Snapshot{}, false
	}
	opts := darkstarWOptions(d.SleepModes)
	if opts == nil {
		return sleeptimerpresentation.Snapshot{}, false
	}
	s := sleeptimerpresentation.Snapshot{Value: d.GetSleepMode(), Options: make([]sleeptimerpresentation.Option, len(opts))}
	found := false
	for i, o := range opts {
		s.Options[i] = sleeptimerpresentation.Option{Value: o.Value, Label: o.Label}
		found = found || o.Value == s.Value
	}
	return s, found
}
func (d *Device) MouseGesturesSnapshot() (mousegesturepresentation.Snapshot, bool) {
	if d == nil || d.DeviceProfile == nil || d.DeviceProfile.MultiGesture < 0 || d.DeviceProfile.MultiGesture > 1 || len(d.DeviceProfile.Tilts) != 4 {
		return mousegesturepresentation.Snapshot{}, false
	}
	keys := make([]int, 0, len(d.DeviceProfile.Tilts))
	for key := range d.DeviceProfile.Tilts {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	s := mousegesturepresentation.Snapshot{Enabled: d.DeviceProfile.MultiGesture == 1, Tilts: make([]mousegesturepresentation.Tilt, 0, len(keys))}
	for _, key := range keys {
		t := d.DeviceProfile.Tilts[key]
		if t.Name == "" || t.Value < 10 || t.Value > 80 {
			return mousegesturepresentation.Snapshot{}, false
		}
		s.Tilts = append(s.Tilts, mousegesturepresentation.Tilt{ID: strconv.Itoa(key), Name: t.Name, Value: t.Value})
	}
	return s, true
}
