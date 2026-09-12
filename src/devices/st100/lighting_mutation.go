package st100

import (
	"fmt"
	"strconv"

	"LumenForge/src/lightingsettings"
	"LumenForge/src/rgb"
)

// These narrow call seams retain ST100's established mutation authorities and
// let the presentation adapter be characterized without opening hardware.
var (
	st100SelectRGBProfile = func(device *Device, effect string) uint8 { return device.UpdateRgbProfile(0, effect) }
	st100SetBrightness    = func(device *Device, value uint8) uint8 { return device.ChangeDeviceBrightnessValue(value) }
	st100SetZoneColor     = func(device *Device, id, option int, color rgb.Color) uint8 {
		return device.UpdateDeviceColor(id, option, color, nil)
	}
	st100UpdateRGBData = func(device *Device, effect string, profile rgb.Profile) uint8 {
		return device.UpdateRgbProfileData(effect, profile)
	}
)

func (d *Device) SetLightingEffect(effect string) error {
	if d == nil || d.DeviceProfile == nil || !d.SupportsLightingEffect(effect) || d.DeviceProfile.RGBCluster {
		return fmt.Errorf("ST100 lighting is unavailable")
	}
	if st100SelectRGBProfile(d, effect) != 1 {
		return fmt.Errorf("unable to select ST100 RGB profile")
	}
	return nil
}

func (d *Device) SetLightingBrightness(value uint8) error {
	if d == nil || d.DeviceProfile == nil || st100SetBrightness(d, value) != 1 {
		return fmt.Errorf("unable to update ST100 brightness")
	}
	return nil
}

func (d *Device) ResolveLightingEffectSettings(effect string) (lightingsettings.EffectSettings, error) {
	if d == nil || effect != "static" || !d.SupportsLightingEffect(effect) {
		return lightingsettings.EffectSettings{}, fmt.Errorf("ST100 effect settings are unavailable")
	}
	return st100StaticSettings(d.GetRgbProfile(effect))
}

func (d *Device) SetLightingEffectSettings(effect string, settings lightingsettings.EffectSettings) error {
	if d == nil || effect != "static" || settings.EffectID != effect || settings.SingleColor == nil {
		return fmt.Errorf("ST100 effect settings are unavailable")
	}
	if err := lightingsettings.Validate(settings); err != nil {
		return err
	}
	profile := d.GetRgbProfile(effect)
	if profile == nil {
		return fmt.Errorf("ST100 static profile is unavailable")
	}
	updated := *profile
	updated.StartColor.Red = settings.SingleColor.Color.Red
	updated.StartColor.Green = settings.SingleColor.Color.Green
	updated.StartColor.Blue = settings.SingleColor.Color.Blue
	if st100UpdateRGBData(d, effect, updated) != 1 {
		return fmt.Errorf("unable to update ST100 static profile")
	}
	return nil
}

// ST100 RGB profiles have no separate reset authority; retaining the legacy
// profile data untouched is safer than inventing one.
func (d *Device) ResetLightingEffectSettings(string) error {
	return fmt.Errorf("ST100 effect reset is unavailable")
}

func (d *Device) SetLightingZoneColor(effect, scope, zoneID, groupID string, color rgb.Color) error {
	if d == nil || d.DeviceProfile == nil || d.DeviceProfile.Stand == nil || effect != "stand" || !st100ValidColor(color) {
		return fmt.Errorf("ST100 authored lighting is unavailable")
	}
	keyID, option, err := d.st100AuthoredMutation(scope, zoneID, groupID)
	if err != nil || st100SetZoneColor(d, keyID, option, color) != 1 {
		return fmt.Errorf("unable to update ST100 authored lighting")
	}
	return nil
}

func (d *Device) SetLightingZoneColors(effect string, zoneIDs []string, color rgb.Color) error {
	if d == nil || d.DeviceProfile == nil || d.DeviceProfile.Stand == nil || effect != "stand" || len(zoneIDs) == 0 || !st100ValidColor(color) {
		return fmt.Errorf("ST100 authored lighting is unavailable")
	}
	seen := make(map[int]struct{}, len(zoneIDs))
	ids := make([]int, 0, len(zoneIDs))
	for _, id := range zoneIDs {
		zoneID, err := strconv.Atoi(id)
		if err != nil || d.st100FindZone(zoneID) == nil {
			return fmt.Errorf("unknown ST100 authored zone")
		}
		if _, duplicate := seen[zoneID]; duplicate {
			return fmt.Errorf("duplicate ST100 authored zone")
		}
		seen[zoneID] = struct{}{}
		ids = append(ids, zoneID)
	}
	for _, id := range ids {
		if st100SetZoneColor(d, id, 0, color) != 1 {
			return fmt.Errorf("unable to update ST100 authored lighting")
		}
	}
	return nil
}

func (d *Device) st100AuthoredMutation(scope, zoneID, groupID string) (int, int, error) {
	switch scope {
	case "zone":
		id, err := strconv.Atoi(zoneID)
		if err != nil || groupID != "" || d.st100FindZone(id) == nil {
			return 0, 0, fmt.Errorf("unknown ST100 authored zone")
		}
		return id, 0, nil
	case "group":
		rowID, err := strconv.Atoi(groupID)
		row, ok := d.DeviceProfile.Stand.Row[rowID]
		if err != nil || zoneID != "" || !ok || len(row.Zones) == 0 {
			return 0, 0, fmt.Errorf("unknown ST100 authored row")
		}
		for id := range row.Zones {
			return id, 1, nil
		}
	case "all":
		if zoneID == "" && groupID == "" && len(d.DeviceProfile.Stand.Row) > 0 {
			return 0, 2, nil
		}
	}
	return 0, 0, fmt.Errorf("invalid ST100 authored selection")
}

func (d *Device) st100FindZone(id int) *Zones {
	for _, row := range d.DeviceProfile.Stand.Row {
		if zone, ok := row.Zones[id]; ok {
			copy := zone
			return &copy
		}
	}
	return nil
}

func st100ValidColor(color rgb.Color) bool {
	return color.Red >= 0 && color.Red <= 255 && color.Green >= 0 && color.Green <= 255 && color.Blue >= 0 && color.Blue <= 255
}
