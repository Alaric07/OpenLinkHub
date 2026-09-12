package st100

import (
	"reflect"
	"testing"

	"LumenForge/src/common"
	"LumenForge/src/lightingsettings"
	"LumenForge/src/rgb"
)

func st100WorkspaceDevice() *Device {
	equalizers := map[int]Equalizer{}
	for id, label := range []string{"32", "64", "125", "250", "500", "1K", "2K", "4K", "8K", "16K"} {
		equalizers[id+1] = Equalizer{Name: label, Value: float64(id - 4)}
	}
	brightness := uint8(66)
	return &Device{Serial: "st100-workspace", Rgb: &rgb.RGB{Profiles: map[string]rgb.Profile{"static": {StartColor: rgb.Color{Red: 1, Green: 2, Blue: 3}}}}, DeviceProfile: &DeviceProfile{RGBProfile: "stand", BrightnessSlider: &brightness, Equalizers: equalizers, Stand: st100Stand()}, UserProfiles: map[string]*DeviceProfile{"Default": {Active: true}, "Desk": {}}}
}

func st100Stand() *Stand {
	indices := map[int]int{1: 12, 2: 24, 3: 0, 4: 3, 5: 21, 6: 6, 7: 18, 8: 15, 9: 9}
	rows := map[int]Row{1: {Zones: map[int]Zones{}}, 2: {Zones: map[int]Zones{}}, 3: {Zones: map[int]Zones{}}, 4: {Zones: map[int]Zones{}}}
	for id, index := range indices {
		row := 2
		if id == 1 {
			row = 1
		} else if id >= 5 && id <= 6 {
			row = 3
		} else if id >= 7 {
			row = 4
		}
		rows[row].Zones[id] = Zones{Name: "", Width: 150, Height: 150, Left: id, Top: row, PacketIndex: []int{index}, Color: rgb.Color{Red: float64(id)}}
	}
	logo := rows[1].Zones[1]
	logo.Name, logo.Width, logo.Height, logo.Left, logo.Top = "Logo", 150, 30, 170, -21
	rows[1].Zones[1] = logo
	return &Stand{Row: rows}
}

func TestST100WorkspaceSnapshotsPreserveSourceContracts(t *testing.T) {
	d := st100WorkspaceDevice()
	d.createDevice()
	if d.instance.ProductType != common.ProductTypeST100 {
		t.Fatalf("product type = %d", d.instance.ProductType)
	}
	profiles, ok := d.DeviceProfileSnapshot()
	if !ok || !profiles.CanSwitch || !profiles.CanSave || !profiles.CanDelete || !reflect.DeepEqual(profiles.Profiles, []string{"Default", "Desk"}) || profiles.ActiveProfile != "Default" {
		t.Fatalf("profile snapshot = %#v usable=%t", profiles, ok)
	}
	headset, ok := d.HeadsetSnapshot()
	if !ok || len(headset.Equalizer) != 10 || headset.Equalizer[0].ID != 1 || headset.Equalizer[0].Label != "32" || headset.Equalizer[0].Value != -4 || headset.Equalizer[9].ID != 10 || headset.Equalizer[9].Label != "16K" || headset.Equalizer[9].Value != 5 {
		t.Fatalf("equalizer snapshot = %#v usable=%t", headset, ok)
	}
}

func TestST100SnapshotsFailClosed(t *testing.T) {
	d := st100WorkspaceDevice()
	d.DeviceProfile.Equalizers[10] = Equalizer{Name: "wrong", Value: 0}
	if _, ok := d.HeadsetSnapshot(); ok {
		t.Fatal("invalid equalizer was exposed")
	}
	d = st100WorkspaceDevice()
	d.DeviceProfile.Stand.Row[2].Zones[2] = Zones{}
	if _, ok := d.LightingSnapshot(); ok {
		t.Fatal("invalid authored geometry was exposed")
	}
}

func TestST100AuthoredLightingGeometryAndMutationMapping(t *testing.T) {
	d := st100WorkspaceDevice()
	sourceGeometry := d.DeviceProfile.Stand.Row[2].Zones[2]
	snapshot, ok := d.LightingSnapshot()
	if !ok || snapshot.AuthoredZoneEditor == nil || len(snapshot.AuthoredZoneEditor.Zones) != 9 || !snapshot.AuthoredZoneEditor.HasGroups {
		t.Fatalf("lighting snapshot = %#v usable=%t", snapshot, ok)
	}
	wantPackets := []int{12, 24, 0, 3, 21, 6, 18, 15, 9}
	wantGeometry := [][2]int{{1, 0}, {0, 1}, {1, 1}, {2, 1}, {0, 2}, {2, 2}, {0, 3}, {1, 3}, {2, 3}}
	occupied := map[[2]int]string{}
	for index, zone := range snapshot.AuthoredZoneEditor.Zones {
		id := index + 1
		if zone.ID != string(rune('0'+id)) || !zone.HasGeometry || zone.Left != wantGeometry[index][0] || zone.Top != wantGeometry[index][1] || zone.Width != 1 || zone.Height != 1 || d.st100FindZone(id).PacketIndex[0] != wantPackets[index] {
			t.Fatalf("zone %d = %#v packet=%v", id, zone, d.st100FindZone(id).PacketIndex)
		}
		position := [2]int{zone.Left, zone.Top}
		if previous, exists := occupied[position]; exists {
			t.Fatalf("zones %s and %s overlap at %#v", previous, zone.ID, position)
		}
		occupied[position] = zone.ID
	}
	if len(occupied) != 9 || snapshot.AuthoredZoneEditor.Zones[4].Left != 0 || snapshot.AuthoredZoneEditor.Zones[5].Left != 2 {
		t.Fatalf("normalized ST100 stand layout = %#v", snapshot.AuthoredZoneEditor.Zones)
	}
	if got := d.DeviceProfile.Stand.Row[2].Zones[2]; got.Left != sourceGeometry.Left || got.Top != sourceGeometry.Top || got.Width != sourceGeometry.Width || got.Height != sourceGeometry.Height || !reflect.DeepEqual(got.PacketIndex, sourceGeometry.PacketIndex) {
		t.Fatalf("presentation rewrote legacy source geometry: before %#v after %#v", sourceGeometry, got)
	}
	if snapshot.AuthoredZoneEditor.Zones[0].Label != "Logo" {
		t.Fatalf("logo = %#v", snapshot.AuthoredZoneEditor.Zones[0])
	}
	if id, option, err := d.st100AuthoredMutation("zone", "3", ""); err != nil || id != 3 || option != 0 {
		t.Fatalf("zone mapping = %d %d %v", id, option, err)
	}
	if id, option, err := d.st100AuthoredMutation("group", "", "3"); err != nil || option != 1 || d.st100FindZone(id) == nil || id < 5 || id > 6 {
		t.Fatalf("group mapping = %d %d %v", id, option, err)
	}
	if id, option, err := d.st100AuthoredMutation("all", "", ""); err != nil || id != 0 || option != 2 {
		t.Fatalf("all mapping = %d %d %v", id, option, err)
	}
}

func TestST100LightingMutationWiring(t *testing.T) {
	d := st100WorkspaceDevice()
	previousProfile, previousBrightness, previousZone, previousRGBData := st100SelectRGBProfile, st100SetBrightness, st100SetZoneColor, st100UpdateRGBData
	t.Cleanup(func() {
		st100SelectRGBProfile, st100SetBrightness, st100SetZoneColor, st100UpdateRGBData = previousProfile, previousBrightness, previousZone, previousRGBData
	})
	profile, brightness, zoneID, option := "", uint8(0), -1, -1
	static := rgb.Profile{}
	st100SelectRGBProfile = func(_ *Device, effect string) uint8 { profile = effect; return 1 }
	st100SetBrightness = func(_ *Device, value uint8) uint8 { brightness = value; return 1 }
	st100SetZoneColor = func(_ *Device, id, gotOption int, _ rgb.Color) uint8 { zoneID, option = id, gotOption; return 1 }
	st100UpdateRGBData = func(_ *Device, effect string, got rgb.Profile) uint8 {
		if effect != "static" {
			t.Fatalf("profile effect = %q", effect)
		}
		static = got
		return 1
	}
	if err := d.SetLightingEffect("static"); err != nil || profile != "static" {
		t.Fatalf("effect = %q err=%v", profile, err)
	}
	if err := d.SetLightingBrightness(42); err != nil || brightness != 42 {
		t.Fatalf("brightness = %d err=%v", brightness, err)
	}
	if err := d.SetLightingZoneColor("stand", "zone", "9", "", rgb.Color{}); err != nil || zoneID != 9 || option != 0 {
		t.Fatalf("zone mutation = %d %d %v", zoneID, option, err)
	}
	settings := lightingsettings.EffectSettings{SchemaVersion: lightingsettings.SchemaVersion, EffectID: "static", SingleColor: &lightingsettings.SingleColorSettings{Color: lightingsettings.Color{Red: 7, Green: 8, Blue: 9}}}
	if err := d.SetLightingEffectSettings("static", settings); err != nil || static.StartColor.Red != 7 || static.StartColor.Green != 8 || static.StartColor.Blue != 9 {
		t.Fatalf("static mutation = %#v err=%v", static, err)
	}
	d.DeviceProfile.RGBCluster = true
	if err := d.SetLightingEffect("static"); err == nil {
		t.Fatal("RGB Cluster effect mutation succeeded")
	}
}
