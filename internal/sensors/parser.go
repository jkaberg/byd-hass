package sensors

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// APIResponse represents the response from the Diplus API
type APIResponse struct {
	Success bool   `json:"success"`
	Val     string `json:"val"`
}

// ParseAPIResponse parses the API response and populates a SensorData struct
func ParseAPIResponse(responseBody []byte) (*SensorData, error) {
	var apiResp APIResponse
	if err := json.Unmarshal(responseBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("API request failed: success=false")
	}

	sensorData := &SensorData{
		Timestamp: time.Now(),
	}

	if err := parseValueString(apiResp.Val, sensorData); err != nil {
		return nil, fmt.Errorf("failed to parse sensor values: %w", err)
	}

	return sensorData, nil
}

// parseValueString parses the pipe-separated key:value string from the API
func parseValueString(valString string, sensorData *SensorData) error {
	if valString == "" {
		return fmt.Errorf("empty value string")
	}

	// Split by pipe separator
	pairs := strings.Split(valString, "|")

	// Use reflection to set struct fields
	v := reflect.ValueOf(sensorData).Elem()

	for _, pair := range pairs {
		// Split key:value
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue // Skip malformed pairs
		}

		key := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(parts[1])

		// Lookup the struct field by the authoritative key directly; no fallback
		// conversion is needed because Diplus now echoes back exactly what we
		// requested.
		field := v.FieldByName(key)
		if !field.IsValid() || !field.CanSet() {
			// Field not found or not settable; skip.
			continue
		}

		// Determine scaling factor based on sensor metadata (defaults to 1)
		scaleFactor := GetScaleFactor(ToSnakeCase(key))

		// Parse the value and set the field with scaling applied where necessary
		if err := setFieldValue(field, valueStr, scaleFactor); err != nil {
			// Log error but continue with other fields
			continue
		}
	}

	return nil
}

// setFieldValue sets a reflect.Value field with the parsed string value
func setFieldValue(field reflect.Value, valueStr string, scaleFactor float64) error {
	// Normalize the value string for European formats
	normalizedValue := normalizeNumericValue(valueStr)

	// If the normalized value is empty, treat it as null/not present
	if normalizedValue == "" {
		return nil // Leave the pointer nil
	}
	if field.Kind() != reflect.Ptr {
		return fmt.Errorf("field is not a pointer")
	}

	// Get the type of the pointer's element
	elemType := field.Type().Elem()

	// Create a new pointer to the element type
	newVal := reflect.New(elemType)

	switch elemType.Kind() {
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(normalizedValue, 64)
		if err != nil {
			return fmt.Errorf("failed to parse float value '%s': %w", normalizedValue, err)
		}
		newVal.Elem().SetFloat(floatVal * scaleFactor)
	case reflect.String:
		newVal.Elem().SetString(normalizedValue)
	default:
		// We currently only expect *float64 and *string fields in SensorData.
		// Unknown types are ignored rather than treated as errors to keep the
		// parser resilient to future struct changes.
		return nil
	}

	field.Set(newVal)

	return nil
}

// normalizeNumericValue converts European number formats to standard formats
func normalizeNumericValue(value string) string {
	if value == "" {
		return ""
	}

	// Replace Unicode minus sign with standard minus
	value = strings.ReplaceAll(value, "−", "-")

	// Replace European decimal comma with dot
	value = strings.ReplaceAll(value, ",", ".")

	return value
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// ToSnakeCase converts a CamelCase string to snake_case.
func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// GetSensorIDs returns a minimal set of essential sensor IDs for basic vehicle monitoring.
func GetSensorIDs() []int {
	return []int{
		33, // BatteryPercentage
		2,  // Speed
		52, // ChargingStatus
		10, // EnginePower
		26, // OutsideTemperature
		3,  // Mileage
		14, // CabinTemperature
		53, // LeftFrontTirePressure
		54, // RightFrontTirePressure
		55, // LeftRearTirePressure
		56, // RightRearTirePressure
	}
}

// GetAllSensorIDs returns all available sensor IDs
func GetAllSensorIDs() []int {
	var ids []int
	for _, sensor := range AllSensors {
		ids = append(ids, sensor.ID)
	}
	return ids
}

// ValidateSensorData performs basic validation on sensor data
func ValidateSensorData(data *SensorData) []string {
	var warnings []string

	// Check for reasonable battery percentage
	if data.BatteryPercentage != nil {
		if *data.BatteryPercentage < 0 || *data.BatteryPercentage > 100 {
			warnings = append(warnings, fmt.Sprintf("Battery percentage out of range: %.1f%%", *data.BatteryPercentage))
		}
	}

	// Check for reasonable speed
	if data.Speed != nil {
		if *data.Speed < 0 || *data.Speed > 300 { // 300 km/h max reasonable speed
			warnings = append(warnings, fmt.Sprintf("Speed out of reasonable range: %.1f km/h", *data.Speed))
		}
	}

	// Check for reasonable temperatures
	if data.CabinTemperature != nil {
		if *data.CabinTemperature < -40 || *data.CabinTemperature > 80 {
			warnings = append(warnings, fmt.Sprintf("Cabin temperature out of reasonable range: %.1f°C", *data.CabinTemperature))
		}
	}

	if data.OutsideTemperature != nil {
		if *data.OutsideTemperature < -50 || *data.OutsideTemperature > 60 {
			warnings = append(warnings, fmt.Sprintf("Outside temperature out of reasonable range: %.1f°C", *data.OutsideTemperature))
		}
	}

	return warnings
}

// GetNonNilFields returns a map of field names to values for all non-nil fields
func GetNonNilFields(data *SensorData) map[string]interface{} {
	result := make(map[string]interface{})

	v := reflect.ValueOf(data).Elem()
	t := reflect.TypeOf(data).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip timestamp field
		if fieldType.Name == "Timestamp" {
			continue
		}

		// Check if pointer field is not nil
		if field.Kind() == reflect.Ptr && !field.IsNil() {
			jsonTag := fieldType.Tag.Get("json")
			if jsonTag != "" {
				// Extract field name from json tag
				tagParts := strings.Split(jsonTag, ",")
				fieldName := tagParts[0]
				result[fieldName] = field.Elem().Interface()
			}
		}
	}

	return result
}

// CompareRawVsParsed compares the raw API response map with the parsed SensorData struct.
func CompareRawVsParsed(responseBody []byte, parsedData *SensorData) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("RAW API vs PARSED VALUES COMPARISON")
	fmt.Println(strings.Repeat("=", 80))

	// Parse the API response
	var apiResp APIResponse
	if err := json.Unmarshal(responseBody, &apiResp); err != nil {
		fmt.Printf("ERROR: Failed to unmarshal API response: %v\n", err)
		return
	}

	if !apiResp.Success {
		fmt.Println("ERROR: API returned success=false")
		return
	}

	// Parse the raw value string into key-value pairs
	rawValues := parseRawValues(apiResp.Val)

	fmt.Printf("Found %d raw values from API\n", len(rawValues))
	fmt.Printf("Parsed %d non-nil fields in struct\n", countNonNilFields(parsedData))

	// Get reflection info for the parsed data
	v := reflect.ValueOf(parsedData).Elem()

	var successCount, failCount, mismatchCount int

	fmt.Println("\n📊 VALUE-BY-VALUE COMPARISON:")

	for key, rawValue := range rawValues {
		// Direct match only; we no longer support automatic key conversion.
		fieldName := key
		field := v.FieldByName(fieldName)

		if !field.IsValid() {
			fmt.Printf("❓ UNKNOWN: %s = '%s' (no matching field)\n", key, rawValue)
			continue
		}

		// Check if field is set (not nil)
		if field.IsNil() {
			fmt.Printf("❌ FAILED: %s = '%s' -> nil (parsing failed)\n", key, rawValue)
			failCount++
			continue
		}

		// Get the actual parsed value
		parsedValue := field.Elem().Interface()

		// Determine expected vs actual types
		expectedType := getExpectedType(rawValue)
		actualType := fmt.Sprintf("%T", parsedValue)

		if expectedType != actualType {
			fmt.Printf("⚠️  MISMATCH: %s = '%s' -> %v (%s) [expected: %s]\n",
				key, rawValue, parsedValue, actualType, expectedType)
			mismatchCount++
		} else {
			fmt.Printf("✅ SUCCESS: %s = '%s' -> %v (%s)\n",
				key, rawValue, parsedValue, actualType)
			successCount++
		}
	}

	// Summary
	fmt.Println("\n" + strings.Repeat("-", 80))
	fmt.Printf("📈 SUMMARY:\n")
	fmt.Printf("  ✅ Successful: %d\n", successCount)
	fmt.Printf("  ⚠️  Type Mismatches: %d\n", mismatchCount)
	fmt.Printf("  ❌ Parse Failures: %d\n", failCount)
	fmt.Printf("  Total Compared: %d\n", successCount+mismatchCount+failCount)

	if mismatchCount > 0 {
		fmt.Printf("\n🔧 TYPE MISMATCH FIXES NEEDED:\n")
		fmt.Printf("Review the ⚠️  MISMATCH entries above and fix the struct field types accordingly.\n")
	}

	if failCount > 0 {
		fmt.Printf("\n🐛 PARSING FAILURES:\n")
		fmt.Printf("Review the ❌ FAILED entries above - these values couldn't be parsed at all.\n")
	}

	fmt.Println(strings.Repeat("=", 80))
}

// parseRawValues parses the raw API value string into key-value pairs
func parseRawValues(valString string) map[string]string {
	values := make(map[string]string)

	if valString == "" {
		return values
	}

	// Split by pipe separator
	pairs := strings.Split(valString, "|")

	for _, pair := range pairs {
		// Split key:value
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			values[key] = value
		}
	}

	return values
}

// getExpectedType determines what Go type a raw string value should be
func getExpectedType(rawValue string) string {
	// Check for empty strings or obvious string values (file paths, etc.)
	if rawValue == "" || strings.Contains(rawValue, "/") || strings.Contains(rawValue, "\\") {
		return "string"
	}

	// Check if it's a number (our sensor data is mostly numeric and we use float64 for all numeric values)
	if _, err := strconv.ParseFloat(rawValue, 64); err == nil {
		// All numeric values in our BYD sensor data are now float64
		return "float64"
	}

	// Check if it's a boolean-like value (though we don't currently use bool types)
	if rawValue == "true" || rawValue == "false" {
		return "bool"
	}

	// Default to string
	return "string"
}

// countNonNilFields counts how many fields in the sensor data are not nil
func countNonNilFields(data *SensorData) int {
	count := 0
	v := reflect.ValueOf(data).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Ptr && !field.IsNil() {
			count++
		}
	}

	return count
}
