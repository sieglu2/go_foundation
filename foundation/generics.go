package foundation

import (
	"math/rand"
	"regexp"
	"strings"
)

// Compile the regex once at package level
var placeholderRegex = regexp.MustCompile(`\${([^}]+)}`)

func ReplaceNamedPlaceholders(template string, replacements map[string]string) string {
	// Use a strings.Builder for efficient string concatenation
	var builder strings.Builder
	lastIndex := 0

	// Find all matches
	matches := placeholderRegex.FindAllStringSubmatchIndex(template, -1)
	if matches == nil {
		return template // No matches, return original
	}

	// Pre-allocate approximate capacity
	builder.Grow(len(template))

	for _, match := range matches {
		// Append everything before the match
		builder.WriteString(template[lastIndex:match[0]])

		// Extract key name (without the ${})
		key := template[match[2]:match[3]]

		// Replace with value if exists, otherwise keep original
		if val, ok := replacements[key]; ok {
			builder.WriteString(val)
		} else {
			builder.WriteString(template[match[0]:match[1]])
		}

		lastIndex = match[1]
	}

	// Append the remainder of the string
	builder.WriteString(template[lastIndex:])

	return builder.String()
}

func RandomString(length int) string {
	const charset = "1234567890abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func SanitizeToRFC1123Subdomain(input string) string {
	if input == "" {
		return "default"
	}

	// Convert to lowercase
	sanitized := strings.ToLower(input)

	// Replace underscores with hyphens
	sanitized = strings.ReplaceAll(sanitized, "_", "-")

	// Replace any character that's not alphanumeric, dot, or hyphen
	reg := regexp.MustCompile("[^a-z0-9\\.\\-]")
	sanitized = reg.ReplaceAllString(sanitized, "-")

	// Replace multiple consecutive hyphens with a single hyphen
	reg = regexp.MustCompile("-+")
	sanitized = reg.ReplaceAllString(sanitized, "-")

	// Replace multiple consecutive dots with a single dot
	reg = regexp.MustCompile("\\.+")
	sanitized = reg.ReplaceAllString(sanitized, ".")

	// Ensure it doesn't start with a hyphen or dot
	reg = regexp.MustCompile("^[\\.\\-]+")
	sanitized = reg.ReplaceAllString(sanitized, "")

	// Ensure it doesn't end with a hyphen or dot
	reg = regexp.MustCompile("[\\.\\-]+$")
	sanitized = reg.ReplaceAllString(sanitized, "")

	// Ensure segments between dots don't start or end with hyphens
	segments := strings.Split(sanitized, ".")
	for i, segment := range segments {
		segment = strings.Trim(segment, "-")
		if segment == "" {
			segment = "x" // Replace empty segments with a valid character
		}
		segments[i] = segment
	}
	sanitized = strings.Join(segments, ".")

	// If we've stripped everything, provide a default
	if sanitized == "" {
		return "default"
	}

	return sanitized
}

func MergeMaps(base, override map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy base map
	for k, v := range base {
		result[k] = v
	}

	// Override with second map
	for k, v := range override {
		if baseVal, exists := result[k]; exists {
			if baseMap, baseIsMap := baseVal.(map[string]interface{}); baseIsMap {
				if overrideMap, overrideIsMap := v.(map[string]interface{}); overrideIsMap {
					result[k] = MergeMaps(baseMap, overrideMap)
					continue
				}
			}
		}
		result[k] = v
	}

	return result
}

func DeepMerge(base, override map[string]any) map[string]any {
	result := make(map[string]any)

	// Copy all base values
	for k, v := range base {
		result[k] = DeepCopy(v)
	}

	// Deep merge override values
	for k, v := range override {
		if existing, exists := result[k]; exists {
			if existingMap, ok := existing.(map[string]any); ok {
				if overrideMap, ok := v.(map[string]any); ok {
					result[k] = DeepMerge(existingMap, overrideMap)
					continue
				}
			}
		}
		result[k] = DeepCopy(v)
	}

	return result
}

func DeepCopy(src any) any {
	if src == nil {
		return nil
	}

	switch v := src.(type) {
	case map[string]any:
		result := make(map[string]any)
		for k, val := range v {
			result[k] = DeepCopy(val)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = DeepCopy(val)
		}
		return result
	default:
		return v
	}
}
