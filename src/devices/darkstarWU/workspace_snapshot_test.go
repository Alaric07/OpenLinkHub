package darkstarWU

import (
	"LumenForge/src/inputmanager"
	"LumenForge/src/rgb"
	"os"
	"strings"
	"testing"
)

func TestWorkspaceSnapshots(t *testing.T) {
	color := rgb.Color{Red: 1, Green: 2, Blue: 3}
	profiles := map[int]DPIProfile{}
	for i := 0; i < 5; i++ {
		profiles[i] = DPIProfile{Name: "Stage", Value: uint16(800 + i*100)}
	}
	profiles[5] = DPIProfile{Name: "Sniper", Value: 200, Sniper: true}
	buttons := map[int]inputmanager.KeyAssignment{}
	for _, key := range darkstarWUVisibleButtonOrder {
		buttons[key] = inputmanager.KeyAssignment{Name: "Button"}
	}
	d := &Device{Serial: "darkstar-wu", MinDPI: 100, MaxDPI: 26000, DPIAmount: 5, KeyAssignment: buttons, KeyAssignmentTypes: map[int]string{0: "None", 1: "Media"}, SleepModes: map[int]string{15: "15 minutes"}, LiftHeights: map[int]string{2: "Low"}, SwitchModes: map[int]string{0: "Disabled", 1: "Enabled"}, PollingRates: map[int]string{4: "1000 Hz"}, UserProfiles: map[string]*DeviceProfile{"Default": {Active: true}}, DeviceProfile: &DeviceProfile{Profile: 2, DPIColor: &color, Profiles: profiles, SleepMode: 15, PollingRate: 4, ButtonOptimization: 1, LiftHeight: 2, Tilts: map[int]TiltOption{0: {Name: "Left", Value: 20}, 1: {Name: "Right", Value: 20}, 2: {Name: "Forward", Value: 20}, 3: {Name: "Back", Value: 20}}}}
	if s, ok := d.DPISnapshot(); !ok || s.ActiveRegularStageID != "2" || len(s.Stages) != 6 {
		t.Fatal("DPI snapshot")
	}
	if s, ok := d.ButtonsSnapshot(); !ok || len(s.Buttons) != 19 {
		t.Fatal("buttons snapshot")
	}
	if s, ok := d.PerformanceSnapshot(); !ok || s.PollingRate == nil || s.PollingRate.Value != 4 {
		t.Fatal("USB polling snapshot")
	}
	if _, ok := d.SleepTimerSnapshot(); !ok {
		t.Fatal("sleep snapshot")
	}
	if s, ok := d.MouseGesturesSnapshot(); !ok || len(s.Tilts) != 4 {
		t.Fatal("gestures snapshot")
	}
}

func TestSaveMouseDPISettingsPreservesSniperOutputPath(t *testing.T) {
	source, err := os.ReadFile("dpi_workspace_mutation.go")
	if err != nil {
		t.Fatal(err)
	}
	want := "d.saveDeviceProfile()\n\tif !d.SniperMode {\n\t\td.toggleDPI()\n\t}\n\treturn 1"
	if !strings.Contains(string(source), want) {
		t.Fatal("DPI save must persist for both paths and only reapply regular DPI outside sniper mode")
	}
}
