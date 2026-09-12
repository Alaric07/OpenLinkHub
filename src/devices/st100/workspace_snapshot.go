package st100

import (
	"sort"

	"LumenForge/src/deviceprofilepresentation"
	"LumenForge/src/headsetpresentation"
)

func (d *Device) DeviceProfileDeviceID() string {
	if d == nil {
		return ""
	}
	return d.Serial
}

func (d *Device) HeadsetDeviceID() string {
	if d == nil {
		return ""
	}
	return d.Serial
}

// DeviceProfileSnapshot exposes the existing ST100 profile authority without
// changing its persistence or profile switching semantics.
func (d *Device) DeviceProfileSnapshot() (deviceprofilepresentation.Snapshot, bool) {
	if d == nil || d.DeviceProfile == nil || len(d.UserProfiles) == 0 {
		return deviceprofilepresentation.Snapshot{}, false
	}
	snapshot := deviceprofilepresentation.Snapshot{Supported: true}
	active := 0
	for name, profile := range d.UserProfiles {
		if name == "" || profile == nil {
			return deviceprofilepresentation.Snapshot{}, false
		}
		snapshot.Profiles = append(snapshot.Profiles, name)
		if profile.Active {
			active++
			snapshot.ActiveProfile = name
		}
	}
	if active != 1 {
		return deviceprofilepresentation.Snapshot{}, false
	}
	sort.Strings(snapshot.Profiles)
	return deviceprofilepresentation.WithMutationCapabilities(snapshot, d), true
}

// HeadsetSnapshot adapts ST100's existing ten-band equalizer to the shared
// headset workspace. The IDs and labels are the persisted source contract.
func (d *Device) HeadsetSnapshot() (headsetpresentation.Snapshot, bool) {
	if d == nil || d.DeviceProfile == nil || len(d.DeviceProfile.Equalizers) != 10 {
		return headsetpresentation.Snapshot{}, false
	}
	labels := []string{"32", "64", "125", "250", "500", "1K", "2K", "4K", "8K", "16K"}
	snapshot := headsetpresentation.Snapshot{Equalizer: make([]headsetpresentation.EqualizerBand, 0, len(labels))}
	for index, label := range labels {
		band, ok := d.DeviceProfile.Equalizers[index+1]
		if !ok || band.Name != label || band.Value < -12 || band.Value > 12 {
			return headsetpresentation.Snapshot{}, false
		}
		snapshot.Equalizer = append(snapshot.Equalizer, headsetpresentation.EqualizerBand{ID: index + 1, Label: label, Value: band.Value})
	}
	return snapshot, headsetpresentation.Valid(snapshot)
}
