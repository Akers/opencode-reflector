package model

// GetMetadataFloat extracts a float64 from session metadata map
func GetMetadataFloat(meta map[string]interface{}, key string, defaultVal float64) float64 {
	if meta == nil {
		return defaultVal
	}
	val, ok := meta[key]
	if !ok {
		return defaultVal
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return defaultVal
	}
}

// GetMetadataString extracts a string from session metadata map
func GetMetadataString(meta map[string]interface{}, key string, defaultVal string) string {
	if meta == nil {
		return defaultVal
	}
	val, ok := meta[key]
	if !ok {
		return defaultVal
	}
	str, ok := val.(string)
	if !ok {
		return defaultVal
	}
	return str
}

// GetMetadataMap extracts a nested map from session metadata
func GetMetadataMap(meta map[string]interface{}, key string) map[string]interface{} {
	if meta == nil {
		return nil
	}
	val, ok := meta[key]
	if !ok {
		return nil
	}
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil
	}
	return m
}