package st100

import (
	"fmt"
	"sort"
	"strconv"

	"LumenForge/src/lightingpresentation"
	"LumenForge/src/lightingsettings"
	"LumenForge/src/rgb"
)

func (d *Device) LightingDeviceID() string {
	if d == nil {
		return ""
	}
	return d.Serial
}

func (d *Device) SupportsLightingEffect(effect string) bool {
	if effect == "stand" {
		return true
	}
	for _, candidate := range rgbModes {
		if candidate == effect {
			_, known := rgb.SoftwareEffectDescriptorByID(effect)
			return known
		}
	}
	return false
}

// LightingSnapshot deliberately reads ST100's existing RGB profile state. It
// does not introduce canonical stores or reinterpret the legacy renderer.
func (d *Device) LightingSnapshot() (lightingpresentation.Snapshot, bool) {
	if d == nil || d.DeviceProfile == nil || d.DeviceProfile.BrightnessSlider == nil || d.DeviceProfile.Stand == nil {
		return lightingpresentation.Snapshot{}, false
	}
	effect := d.DeviceProfile.RGBProfile
	if !d.SupportsLightingEffect(effect) {
		return lightingpresentation.Snapshot{}, false
	}
	snapshot := lightingpresentation.Snapshot{
		TargetKind:        "native",
		ConfiguredEffect:  effect,
		EffectSupported:   true,
		HasBrightness:     true,
		Brightness:        *d.DeviceProfile.BrightnessSlider,
		ClusterControlled: d.DeviceProfile.RGBCluster,
	}
	for _, id := range rgbModes {
		if !d.SupportsLightingEffect(id) {
			continue
		}
		label := id
		if id == "stand" {
			label = "Stand"
		} else if descriptor, ok := rgb.SoftwareEffectDescriptorByID(id); ok {
			label = descriptor.Label
		}
		snapshot.SupportedEffects = append(snapshot.SupportedEffects, lightingpresentation.EffectOption{ID: id, Label: label})
	}
	if effect == "stand" {
		snapshot.AuthoredZoneEditor = st100AuthoredZoneEditor(d.DeviceProfile.Stand)
		return snapshot, snapshot.AuthoredZoneEditor != nil
	}
	// Static is the one existing ST100 profile setting with an exact shared
	// representation. Other renderer profiles remain selectable, but expose no
	// lossy generic settings editor.
	if effect == "static" {
		profile := d.GetRgbProfile("static")
		if profile == nil {
			return lightingpresentation.Snapshot{}, false
		}
		snapshot.PaletteKind = string(rgb.LightingPaletteStaticSingle)
		snapshot.SingleColorHex = st100ColorHex(profile.StartColor)
	}
	return snapshot, true
}

func st100AuthoredZoneEditor(stand *Stand) *lightingpresentation.AuthoredZoneEditor {
	if stand == nil || len(stand.Row) == 0 {
		return nil
	}
	rowIDs := make([]int, 0, len(stand.Row))
	for rowID := range stand.Row {
		rowIDs = append(rowIDs, rowID)
	}
	sort.Ints(rowIDs)
	editor := &lightingpresentation.AuthoredZoneEditor{EffectID: "stand", Heading: "Zones", Description: "Select one or more zones, choose a color, then apply it to the selected zones.", HasGroups: true}
	for rowIndex, rowID := range rowIDs {
		row := stand.Row[rowID]
		if len(row.Zones) == 0 {
			return nil
		}
		zoneIDs := make([]int, 0, len(row.Zones))
		for zoneID := range row.Zones {
			zoneIDs = append(zoneIDs, zoneID)
		}
		sort.Ints(zoneIDs)
		for zoneIndex, zoneID := range zoneIDs {
			zone := row.Zones[zoneID]
			if len(zone.PacketIndex) == 0 || zone.Width <= 0 || zone.Height <= 0 {
				return nil
			}
			left, top, width, height := st100NormalizedZoneGeometry(rowIndex, len(zoneIDs), zoneIndex)
			label := zone.Name
			if label == "" {
				label = "Zone " + strconv.Itoa(zoneID)
			}
			// ST100's stored Left and Top are offsets within its legacy
			// row-by-row template, not coordinates in one shared canvas. Use
			// only source-backed row membership to construct the modern grid;
			// never rewrite the persisted legacy geometry.
			editor.Zones = append(editor.Zones, lightingpresentation.AuthoredZone{ID: strconv.Itoa(zoneID), Label: label, ColorHex: st100ColorHex(zone.Color), GroupID: strconv.Itoa(rowID), GroupLabel: "Row " + strconv.Itoa(rowID), HasGeometry: true, Left: left, Top: top, Width: width, Height: height})
		}
	}
	return editor
}

// st100NormalizedZoneGeometry maps ascending source rows and zones to the
// physical stand layout on a compact grid: centered logo; three,
// two-with-center-gap, then three zones. Legacy zone coordinates cannot be
// used globally because they restart within every row.
func st100NormalizedZoneGeometry(rowIndex, rowLength, zoneIndex int) (left, top, width, height int) {
	top, width, height = rowIndex, 1, 1
	switch rowLength {
	case 1:
		return 1, top, width, height
	case 2:
		return zoneIndex * 2, top, width, height
	default:
		return zoneIndex, top, width, height
	}
}

func st100ColorHex(color rgb.Color) string {
	return fmt.Sprintf("#%02x%02x%02x", uint8(color.Red), uint8(color.Green), uint8(color.Blue))
}

func st100StaticSettings(profile *rgb.Profile) (lightingsettings.EffectSettings, error) {
	if profile == nil {
		return lightingsettings.EffectSettings{}, fmt.Errorf("ST100 static profile is unavailable")
	}
	return lightingsettings.EffectSettings{SchemaVersion: lightingsettings.SchemaVersion, EffectID: "static", SingleColor: &lightingsettings.SingleColorSettings{Color: lightingsettings.Color{Red: profile.StartColor.Red, Green: profile.StartColor.Green, Blue: profile.StartColor.Blue}}}, nil
}
