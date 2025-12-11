package params

import (
	"fmt"
	mathrand "math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wcy-dt/ponghub/internal/types/types/highlight"
)

// ParameterResolver handles dynamic parameter resolution in configuration
type ParameterResolver struct {
	currentTime time.Time
	randSource  *mathrand.Rand
}

// NewParameterResolver creates a new parameter resolver with current time
func NewParameterResolver() *ParameterResolver {
	return &ParameterResolver{
		currentTime: time.Now(),
		randSource:  mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
	}
}

// resolveSpecialParameter resolves non-datetime special parameters
func (pr *ParameterResolver) resolveSpecialParameter(param string) string {
	// Handle different types of special parameters
	switch {
	// UUID generation
	case param == "uuid":
		return uuid.New().String()
	case param == "uuid_short":
		return strings.ReplaceAll(uuid.New().String(), "-", "")

	// Random numbers
	case param == "rand":
		return fmt.Sprintf("%d", pr.randSource.Intn(1000000))
	case param == "rand_int":
		return fmt.Sprintf("%d", pr.randSource.Intn(2147483647))
	case strings.HasPrefix(param, "rand(") && strings.HasSuffix(param, ")"):
		// Handle rand(min,max) format
		content := param[5 : len(param)-1]
		parts := strings.Split(content, ",")
		if len(parts) == 2 {
			paramMin, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			paramMax, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil && paramMax > paramMin {
				return fmt.Sprintf("%d", pr.randSource.Intn(paramMax-paramMin)+paramMin)
			}
		}
		return fmt.Sprintf("%d", pr.randSource.Intn(1000000))

	// Random strings
	case param == "rand_str":
		return pr.generateRandomString(8, "")
	case param == "rand_str_secure":
		return pr.generateSecureRandomString(16)
	case strings.HasPrefix(param, "rand_str(") && strings.HasSuffix(param, ")"):
		// Handle rand_str(length) format
		lengthStr := param[9 : len(param)-1]
		if length, err := strconv.Atoi(strings.TrimSpace(lengthStr)); err == nil && length > 0 {
			return pr.generateRandomString(length, "")
		}
		return pr.generateRandomString(8, "")
	case strings.HasPrefix(param, "rand_hex(") && strings.HasSuffix(param, ")"):
		// Handle rand_hex(length) format
		lengthStr := param[9 : len(param)-1]
		if length, err := strconv.Atoi(strings.TrimSpace(lengthStr)); err == nil && length > 0 {
			return pr.generateRandomString(length, HexCharset)
		}
		return pr.generateRandomString(8, HexCharset)

	// Environment variables
	case strings.HasPrefix(param, "env(") && strings.HasSuffix(param, ")"):
		envVar := param[4 : len(param)-1]
		if value := os.Getenv(envVar); value != "" {
			return value
		}
		return ""

	// Sequence numbers (based on current time)
	case param == "seq":
		return fmt.Sprintf("%d", pr.currentTime.UnixNano()%1000000)
	case param == "seq_daily":
		// Daily sequence: seconds since midnight
		midnight := time.Date(pr.currentTime.Year(), pr.currentTime.Month(), pr.currentTime.Day(), 0, 0, 0, 0, pr.currentTime.Location())
		return fmt.Sprintf("%d", int(pr.currentTime.Sub(midnight).Seconds()))

	// Hash-like values
	case param == "hash_short":
		return fmt.Sprintf("%x", pr.currentTime.UnixNano()%0xFFFFFF)
	case param == "hash_md5_like":
		return fmt.Sprintf("%032x", pr.currentTime.UnixNano())

	// Network and System Information
	case param == "local_ip":
		return pr.getLocalIP()
	case param == "hostname":
		return pr.getHostname()
	case param == "user_agent":
		return pr.generateUserAgent()
	case param == "http_method":
		return HTTPMethods[pr.randSource.Intn(len(HTTPMethods))]

	// Encoding and Decoding
	case strings.HasPrefix(param, "base64(") && strings.HasSuffix(param, ")"):
		content := param[7 : len(param)-1]
		return pr.base64Encode(content)
	case strings.HasPrefix(param, "url_encode(") && strings.HasSuffix(param, ")"):
		content := param[11 : len(param)-1]
		return pr.urlEncode(content)
	case strings.HasPrefix(param, "json_escape(") && strings.HasSuffix(param, ")"):
		content := param[12 : len(param)-1]
		return pr.jsonEscape(content)

	// Mathematical Operations
	case strings.HasPrefix(param, "add(") && strings.HasSuffix(param, ")"):
		return pr.mathOperation(param[4:len(param)-1], "add")
	case strings.HasPrefix(param, "sub(") && strings.HasSuffix(param, ")"):
		return pr.mathOperation(param[4:len(param)-1], "sub")
	case strings.HasPrefix(param, "mul(") && strings.HasSuffix(param, ")"):
		return pr.mathOperation(param[4:len(param)-1], "mul")
	case strings.HasPrefix(param, "div(") && strings.HasSuffix(param, ")"):
		return pr.mathOperation(param[4:len(param)-1], "div")

	// Text Processing
	case strings.HasPrefix(param, "upper(") && strings.HasSuffix(param, ")"):
		content := param[6 : len(param)-1]
		return strings.ToUpper(content)
	case strings.HasPrefix(param, "lower(") && strings.HasSuffix(param, ")"):
		content := param[6 : len(param)-1]
		return strings.ToLower(content)
	case strings.HasPrefix(param, "reverse(") && strings.HasSuffix(param, ")"):
		content := param[8 : len(param)-1]
		return pr.reverseString(content)
	case strings.HasPrefix(param, "substr(") && strings.HasSuffix(param, ")"):
		return pr.subString(param[7 : len(param)-1])

	// Color and CSS
	case param == "color_hex":
		return pr.generateHexColor()
	case param == "color_rgb":
		return pr.generateRGBColor()
	case param == "color_hsl":
		return pr.generateHSLColor()

	// File and MIME types
	case param == "mime_type":
		return pr.generateMimeType()
	case param == "file_ext":
		return pr.generateFileExtension()

	// Fake Data Generation
	case param == "fake_email":
		return pr.generateFakeEmail()
	case param == "fake_phone":
		return pr.generateFakePhone()
	case param == "fake_name":
		return pr.generateFakeName()
	case param == "fake_domain":
		return pr.generateFakeDomain()

	// Time calculations
	case strings.HasPrefix(param, "time_add(") && strings.HasSuffix(param, ")"):
		return pr.timeCalculation(param[9:len(param)-1], "add")
	case strings.HasPrefix(param, "time_sub(") && strings.HasSuffix(param, ")"):
		return pr.timeCalculation(param[9:len(param)-1], "sub")

	default:
		// If it's a time format, try to format it
		if strings.Contains(param, "%") {
			return pr.formatTimeWithPattern(param)
		}
		return param // Return as-is if not recognized
	}
}

// resolveSpecialParameterWithSecret resolves parameters for display with sensitive data masking
func (pr *ParameterResolver) resolveSpecialParameterWithSecret(param string) string {
	// Handle different types of special parameters
	switch {
	// Environment variables
	case strings.HasPrefix(param, "env(") && strings.HasSuffix(param, ")"):
		envVar := param[4 : len(param)-1]
		if value := os.Getenv(envVar); value != "" {
			return pr.maskSensitiveValue(value)
		}
		return ""

	// For other parameters, use normal resolution
	default:
		return pr.resolveSpecialParameter(param)
	}
}

// maskSensitiveValue creates a masked version of a sensitive value
func (pr *ParameterResolver) maskSensitiveValue(value string) string {
	if len(value) == 0 {
		return value
	}

	if len(value) <= 6 {
		return strings.Repeat("*", len(value))
	}

	// Show first 1 and last 1 character, mask the middle
	visible := 1

	prefix := value[:visible]
	suffix := value[len(value)-visible:]
	maskLength := len(value) - 2*visible

	return prefix + strings.Repeat("*", maskLength) + suffix
}

// ResolveParameters resolves dynamic parameters in a string
func (pr *ParameterResolver) ResolveParameters(input string) string {
	return pr.resolveParametersWithDepth(input, 0, make(map[string]bool))
}

// resolveParametersWithDepth resolves parameters with recursion depth control and cycle detection
func (pr *ParameterResolver) resolveParametersWithDepth(input string, depth int, resolved map[string]bool) string {
	const maxDepth = 10 // Prevent infinite recursion

	if depth >= maxDepth {
		return input // Return as-is if max depth reached
	}

	// Check for cycles in resolution
	if resolved[input] {
		return input // Return as-is if we've already processed this exact string
	}

	// Use regex to find parameters in {{...}} format
	re := regexp.MustCompile(`\{\{([^{}]+)}}`)

	// If no matches found, return the input
	if !re.MatchString(input) {
		return input
	}

	// Mark this input as being resolved to detect cycles
	resolved[input] = true

	// Find all matches and their positions
	matches := re.FindAllStringSubmatchIndex(input, -1)
	if len(matches) == 0 {
		delete(resolved, input)
		return input
	}

	result := input

	// Process matches from right to left to preserve indices
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		fullMatchStart, fullMatchEnd := match[0], match[1]
		paramStart, paramEnd := match[2], match[3]

		// Extract the parameter content
		param := strings.TrimSpace(input[paramStart:paramEnd])

		// First, recursively resolve any nested parameters within this parameter
		resolvedParam := pr.resolveParametersWithDepth(param, depth+1, resolved)

		// Then resolve the parameter itself
		var resolvedValue string
		if strings.HasPrefix(resolvedParam, "%") {
			resolvedValue = pr.formatTimeWithPattern(resolvedParam)
		} else {
			resolvedValue = pr.resolveSpecialParameter(resolvedParam)
		}

		// Replace the match in the result
		result = result[:fullMatchStart] + resolvedValue + result[fullMatchEnd:]
	}

	// Remove this input from resolved map
	delete(resolved, input)

	// Check if the result contains more parameters that need resolution
	if re.MatchString(result) && result != input {
		// Recursively resolve the result if it still contains parameters
		return pr.resolveParametersWithDepth(result, depth, resolved)
	}

	return result
}

// GetResolvedValue returns the resolved value for display purposes
func (pr *ParameterResolver) GetResolvedValue(original string) string {
	return pr.ResolveParameters(original)
}

// GetOriginalValue returns the original template value for configuration display
func (pr *ParameterResolver) GetOriginalValue(original string) string {
	return original
}

// HighlightChanges creates a highlighted version showing what parts were replaced
func (pr *ParameterResolver) HighlightChanges(originalURL string) (string, []highlight.Segment) {
	re := regexp.MustCompile(`\{\{([^}]+)}}`)
	matches := re.FindAllStringSubmatchIndex(originalURL, -1)

	if len(matches) == 0 {
		return originalURL, nil
	}

	var segments []highlight.Segment
	result := ""
	lastEnd := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		// Add unchanged part before the match
		if start > lastEnd {
			unchanged := originalURL[lastEnd:start]
			result += unchanged
			segments = append(segments, highlight.Segment{
				Text:        unchanged,
				IsHighlight: false,
			})
		}

		// Replace the match with resolved value (with masking for sensitive data)
		param := strings.TrimSpace(originalURL[match[2]:match[3]]) // Extract parameter without {{...}}
		var resolved string

		if strings.HasPrefix(param, "%") {
			resolved = pr.formatTimeWithPattern(param)
		} else {
			// Use display version with masking for environment variables
			resolved = pr.resolveSpecialParameterWithSecret(param)
		}

		result += resolved
		segments = append(segments, highlight.Segment{
			Text:        resolved,
			IsHighlight: true,
		})

		lastEnd = end
	}

	// Add any remaining unchanged part
	if lastEnd < len(originalURL) {
		unchanged := originalURL[lastEnd:]
		result += unchanged
		segments = append(segments, highlight.Segment{
			Text:        unchanged,
			IsHighlight: false,
		})
	}

	return result, segments
}
