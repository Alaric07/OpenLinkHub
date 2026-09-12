package katarproW

import (
	"LumenForge/src/buttonspresentation"
	"LumenForge/src/deviceprofilepresentation"
	"LumenForge/src/dpipresentation"
	"LumenForge/src/performancepresentation"
	"LumenForge/src/sleeptimerpresentation"
	"fmt"
	"sort"
	"strconv"
)

var katarProWVisibleButtonOrder = []int{1, 2, 4, 8, 16, 32}

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
func (d *Device) DPISnapshot() (dpipresentation.Snapshot, bool) {
	if d == nil || d.DeviceProfile == nil || d.MinDPI < 1 || d.MaxDPI < d.MinDPI || d.DPIAmount < 1 || len(d.DeviceProfile.Profiles) != d.DPIAmount {
		return dpipresentation.Snapshot{}, false
	}
	keys := make([]int, 0, len(d.DeviceProfile.Profiles))
	for k := range d.DeviceProfile.Profiles {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	s := dpipresentation.Snapshot{MinimumDPI: d.MinDPI, MaximumDPI: d.MaxDPI, Stages: make([]dpipresentation.Stage, 0, len(keys))}
	sniper := false
	for _, k := range keys {
		p := d.DeviceProfile.Profiles[k]
		if p.Name == "" || p.Color == nil || p.Value < uint16(d.MinDPI) || p.Value > uint16(d.MaxDPI) {
			return dpipresentation.Snapshot{}, false
		}
		active := !p.Sniper && k == d.DeviceProfile.Profile
		if active {
			s.ActiveRegularStageID = strconv.Itoa(k)
		}
		sniper = sniper || p.Sniper
		s.Stages = append(s.Stages, dpipresentation.Stage{ID: strconv.Itoa(k), Name: p.Name, DPI: p.Value, ColorHex: fmt.Sprintf("#%02x%02x%02x", uint8(p.Color.Red), uint8(p.Color.Green), uint8(p.Color.Blue)), Sniper: p.Sniper, Active: active || (p.Sniper && d.SniperMode)})
	}
	return s, sniper && s.ActiveRegularStageID != ""
}
func (d *Device) ButtonsSnapshot() (buttonspresentation.Snapshot, bool) {
	if d == nil || len(d.KeyAssignmentTypes) == 0 {
		return buttonspresentation.Snapshot{}, false
	}
	s := buttonspresentation.Snapshot{Buttons: make([]buttonspresentation.Button, 0, 6)}
	for _, k := range katarProWVisibleButtonOrder {
		a, ok := d.KeyAssignment[k]
		if !ok || a.Name == "" {
			return buttonspresentation.Snapshot{}, false
		}
		s.Buttons = append(s.Buttons, buttonspresentation.Button{KeyIndex: k, Name: a.Name, Default: a.Default, PressAndHold: a.ActionHold, OnRelease: a.OnRelease, ActionType: a.ActionType, ActionCommand: a.ActionCommand, IsMacro: a.IsMacro, ProfileSwitch: a.ProfileSwitch})
	}
	for _, k := range []int{0, 1, 2, 3, 8, 9, 10, 11} {
		label, ok := d.KeyAssignmentTypes[k]
		if !ok || label == "" {
			return buttonspresentation.Snapshot{}, false
		}
		s.AssignmentTypes = append(s.AssignmentTypes, buttonspresentation.AssignmentType{ID: uint8(k), Label: label})
	}
	return s, true
}
func katarProWOptions(m map[int]string) []performancepresentation.Option {
	if len(m) == 0 {
		return nil
	}
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	o := make([]performancepresentation.Option, 0, len(keys))
	for _, k := range keys {
		if m[k] == "" {
			return nil
		}
		o = append(o, performancepresentation.Option{Value: k, Label: m[k]})
	}
	return o
}
func (d *Device) PerformanceSnapshot() (performancepresentation.Snapshot, bool) {
	if d == nil || d.DeviceProfile == nil {
		return performancepresentation.Snapshot{}, false
	}
	poll, opt := katarProWOptions(d.PollingRates), katarProWOptions(d.SwitchModes)
	if poll == nil || opt == nil {
		return performancepresentation.Snapshot{}, false
	}
	if _, ok := d.PollingRates[d.DeviceProfile.PollingRate]; !ok {
		return performancepresentation.Snapshot{}, false
	}
	if _, ok := d.SwitchModes[d.DeviceProfile.ButtonOptimization]; !ok {
		return performancepresentation.Snapshot{}, false
	}
	return performancepresentation.Snapshot{PollingRate: &performancepresentation.SelectSetting{Value: d.DeviceProfile.PollingRate, Options: poll}, ButtonOptimization: &performancepresentation.SelectSetting{Value: d.DeviceProfile.ButtonOptimization, Options: opt}}, true
}
func (d *Device) DeviceProfileSnapshot() (deviceprofilepresentation.Snapshot, bool) {
	if d == nil || len(d.UserProfiles) == 0 {
		return deviceprofilepresentation.Snapshot{}, false
	}
	s := deviceprofilepresentation.Snapshot{Supported: true}
	for n, p := range d.UserProfiles {
		if n == "" || p == nil {
			return deviceprofilepresentation.Snapshot{}, false
		}
		s.Profiles = append(s.Profiles, n)
		if p.Active {
			if s.ActiveProfile != "" {
				return deviceprofilepresentation.Snapshot{}, false
			}
			s.ActiveProfile = n
		}
	}
	sort.Strings(s.Profiles)
	s = deviceprofilepresentation.WithMutationCapabilities(s, d)
	return s, s.ActiveProfile != ""
}
func (d *Device) SleepTimerSnapshot() (sleeptimerpresentation.Snapshot, bool) {
	if d == nil || d.DeviceProfile == nil || !d.Connected {
		return sleeptimerpresentation.Snapshot{}, false
	}
	o := katarProWOptions(d.SleepModes)
	if o == nil {
		return sleeptimerpresentation.Snapshot{}, false
	}
	s := sleeptimerpresentation.Snapshot{Value: d.GetSleepMode(), Options: make([]sleeptimerpresentation.Option, len(o))}
	found := false
	for i, v := range o {
		s.Options[i] = sleeptimerpresentation.Option{Value: v.Value, Label: v.Label}
		found = found || v.Value == s.Value
	}
	return s, found
}
