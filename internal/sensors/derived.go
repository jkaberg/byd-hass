package sensors

// DeriveChargingStatus derives a human-readable charging state from the raw
// Diplus metrics. The logic is as follows:
//  1. If ChargeGunState is nil or not equal to 2 → "disconnected".
//  2. If ChargeGunState == 2 *and* EnginePower > -1 → "charging".
//  3. Otherwise (gun connected but power <= -1) → "connected".
//
// This helper lives in the sensors package so that other components (MQTT
// transmitter, ABRP, etc.) can reuse the logic without duplicating it.
func DeriveChargingStatus(data *SensorData) string {
	if data == nil || data.ChargeGunState == nil || *data.ChargeGunState != 2 {
		return "disconnected"
	}

	// At this point the charge gun is physically connected.
	if data.EnginePower != nil && *data.EnginePower > -1 {
		return "charging"
	}

	return "connected"
}
