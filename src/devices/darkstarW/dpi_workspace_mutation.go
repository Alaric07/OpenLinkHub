package darkstarW

import "LumenForge/src/rgb"

func (d *Device) SelectMouseDPIStage(stage int) uint8 {
	if d == nil || d.DeviceProfile == nil {
		return 0
	}
	p, ok := d.DeviceProfile.Profiles[stage]
	if !ok || p.Sniper {
		return 0
	}
	if d.DeviceProfile.Profile == stage {
		return 1
	}
	d.DeviceProfile.Profile = stage
	d.saveDeviceProfile()
	if !d.SniperMode {
		d.toggleDPI()
	}
	return 1
}
func (d *Device) SetMouseSniperMode(active bool) uint8 {
	if d == nil || d.DeviceProfile == nil {
		return 0
	}
	for _, p := range d.DeviceProfile.Profiles {
		if p.Sniper {
			if d.SniperMode != active {
				d.sniperMode(active)
			}
			return 1
		}
	}
	return 0
}
func (d *Device) SaveMouseDPISettings(stages map[int]uint16, colors map[int]rgb.Color) uint8 {
	if d == nil || d.DeviceProfile == nil || d.DeviceProfile.DPIColor == nil || len(stages) != len(d.DeviceProfile.Profiles) || len(colors) != len(d.DeviceProfile.Profiles) {
		return 0
	}
	var color rgb.Color
	first := true
	for key, p := range d.DeviceProfile.Profiles {
		value, hasValue := stages[key]
		c, hasColor := colors[key]
		if !hasValue || !hasColor || value < uint16(d.MinDPI) || value > uint16(d.MaxDPI) || !darkstarWColorOK(c) {
			return 0
		}
		if first {
			color = c
			first = false
		} else if c.Red != color.Red || c.Green != color.Green || c.Blue != color.Blue {
			return 0
		}
		_ = p
	}
	for key := range stages {
		if _, ok := d.DeviceProfile.Profiles[key]; !ok {
			return 0
		}
	}
	for key := range colors {
		if _, ok := d.DeviceProfile.Profiles[key]; !ok {
			return 0
		}
	}
	for key, value := range stages {
		p := d.DeviceProfile.Profiles[key]
		p.Value = value
		d.DeviceProfile.Profiles[key] = p
	}
	d.DeviceProfile.DPIColor.Red, d.DeviceProfile.DPIColor.Green, d.DeviceProfile.DPIColor.Blue = color.Red, color.Green, color.Blue
	d.saveDeviceProfile()
	if !d.SniperMode {
		d.toggleDPI()
	}
	return 1
}
