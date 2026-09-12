package darkstarWU

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
	var c rgb.Color
	first := true
	for k := range d.DeviceProfile.Profiles {
		v, hv := stages[k]
		color, hc := colors[k]
		if !hv || !hc || v < uint16(d.MinDPI) || v > uint16(d.MaxDPI) || !darkstarWUColorOK(color) {
			return 0
		}
		if first {
			c = color
			first = false
		} else if c.Red != color.Red || c.Green != color.Green || c.Blue != color.Blue {
			return 0
		}
	}
	for k := range stages {
		if _, ok := d.DeviceProfile.Profiles[k]; !ok {
			return 0
		}
	}
	for k := range colors {
		if _, ok := d.DeviceProfile.Profiles[k]; !ok {
			return 0
		}
	}
	for k, v := range stages {
		p := d.DeviceProfile.Profiles[k]
		p.Value = v
		d.DeviceProfile.Profiles[k] = p
	}
	d.DeviceProfile.DPIColor.Red, d.DeviceProfile.DPIColor.Green, d.DeviceProfile.DPIColor.Blue = c.Red, c.Green, c.Blue
	d.saveDeviceProfile()
	if !d.SniperMode {
		d.toggleDPI()
	}
	return 1
}
