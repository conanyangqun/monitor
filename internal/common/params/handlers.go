package params

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// generateRandomString generates a random string of specified length
func (pr *ParameterResolver) generateRandomString(length int, charset string) string {
	if charset == "" {
		charset = DefaultCharset
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[pr.randSource.Intn(len(charset))]
	}
	return string(result)
}

// generateSecureRandomString generates a cryptographically secure random string
func (pr *ParameterResolver) generateSecureRandomString(length int) string {
	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(DefaultCharset))))
		result[i] = DefaultCharset[num.Int64()]
	}
	return string(result)
}

// formatTimeWithPattern converts Go time format patterns to actual values
func (pr *ParameterResolver) formatTimeWithPattern(pattern string) string {
	// Convert strftime-like patterns to Go time format
	goFormat := pattern
	for strftime, goFmt := range TimeFormatReplacements {
		if goFmt != "" {
			goFormat = strings.ReplaceAll(goFormat, strftime, goFmt)
		}
	}

	// Handle special cases that don't have direct Go equivalents
	if strings.Contains(pattern, "%U") || strings.Contains(pattern, "%W") {
		_, week := pr.currentTime.ISOWeek()
		weekStr := fmt.Sprintf("%02d", week)
		goFormat = strings.ReplaceAll(goFormat, "%U", weekStr)
		goFormat = strings.ReplaceAll(goFormat, "%W", weekStr)
	}

	if strings.Contains(pattern, "%w") {
		weekday := int(pr.currentTime.Weekday())
		goFormat = strings.ReplaceAll(goFormat, "%w", fmt.Sprintf("%d", weekday))
	}

	if strings.Contains(pattern, "%s") {
		timestamp := fmt.Sprintf("%d", pr.currentTime.Unix())
		goFormat = strings.ReplaceAll(goFormat, "%s", timestamp)
	}

	return pr.currentTime.Format(goFormat)
}

func (pr *ParameterResolver) getLocalIP() string {
	adders, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range adders {
		// Skip loopback and down interfaces
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}

	return ""
}

func (pr *ParameterResolver) getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

func (pr *ParameterResolver) generateUserAgent() string {
	return UserAgents[pr.randSource.Intn(len(UserAgents))]
}

func (pr *ParameterResolver) base64Encode(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func (pr *ParameterResolver) urlEncode(input string) string {
	return url.QueryEscape(input)
}

func (pr *ParameterResolver) jsonEscape(input string) string {
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return input
	}
	// Remove the surrounding quotes from JSON marshal
	result := string(jsonBytes)
	if len(result) >= 2 && result[0] == '"' && result[len(result)-1] == '"' {
		return result[1 : len(result)-1]
	}
	return result
}

func (pr *ParameterResolver) mathOperation(input, operation string) string {
	parts := strings.Split(input, ",")
	if len(parts) != 2 {
		return "0"
	}

	num1, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	num2, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)

	if err1 != nil || err2 != nil {
		return "0"
	}

	var result float64
	switch operation {
	case "add":
		result = num1 + num2
	case "sub":
		result = num1 - num2
	case "mul":
		result = num1 * num2
	case "div":
		if num2 != 0 {
			result = num1 / num2
		} else {
			return "0"
		}
	default:
		return "0"
	}

	return fmt.Sprintf("%f", result)
}

func (pr *ParameterResolver) reverseString(input string) string {
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func (pr *ParameterResolver) subString(input string) string {
	parts := strings.Split(input, ",")
	if len(parts) != 3 {
		return ""
	}

	str := strings.TrimSpace(parts[0])
	start, err1 := strconv.Atoi(strings.TrimSpace(parts[1]))
	length, err2 := strconv.Atoi(strings.TrimSpace(parts[2]))

	if err1 != nil || err2 != nil || start < 0 || length < 0 {
		return ""
	}

	if start+length > len(str) {
		length = len(str) - start
	}

	return str[start : start+length]
}

func (pr *ParameterResolver) generateHexColor() string {
	return fmt.Sprintf("#%06x", pr.randSource.Intn(0xFFFFFF))
}

func (pr *ParameterResolver) generateRGBColor() string {
	r := pr.randSource.Intn(256)
	g := pr.randSource.Intn(256)
	b := pr.randSource.Intn(256)
	return fmt.Sprintf("rgb(%d,%d,%d)", r, g, b)
}

func (pr *ParameterResolver) generateHSLColor() string {
	h := pr.randSource.Intn(360)
	s := pr.randSource.Intn(101)
	l := pr.randSource.Intn(101)
	return fmt.Sprintf("hsl(%d,%d%%,%d%%)", h, s, l)
}

func (pr *ParameterResolver) generateMimeType() string {
	return MimeTypes[pr.randSource.Intn(len(MimeTypes))]
}

func (pr *ParameterResolver) generateFileExtension() string {
	return FileExtensions[pr.randSource.Intn(len(FileExtensions))]
}

func (pr *ParameterResolver) generateFakeEmail() string {
	return fmt.Sprintf("user%d@example.com", pr.randSource.Intn(10000))
}

func (pr *ParameterResolver) generateFakePhone() string {
	return fmt.Sprintf("+1-800-%04d-%04d", pr.randSource.Intn(10000), pr.randSource.Intn(10000))
}

func (pr *ParameterResolver) generateFakeName() string {
	return fmt.Sprintf("%s %s", FirstNames[pr.randSource.Intn(len(FirstNames))], LastNames[pr.randSource.Intn(len(LastNames))])
}

func (pr *ParameterResolver) generateFakeDomain() string {
	return fmt.Sprintf("www.%s", FakeDomains[pr.randSource.Intn(len(FakeDomains))])
}

func (pr *ParameterResolver) timeCalculation(input, operation string) string {
	parts := strings.Split(input, ",")
	if len(parts) != 2 {
		return ""
	}

	timeStr := strings.TrimSpace(parts[0])
	valueStr := strings.TrimSpace(parts[1])

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return ""
	}

	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return ""
	}

	var result time.Time
	switch operation {
	case "add":
		result = t.Add(time.Duration(value) * time.Second)
	case "sub":
		result = t.Add(-time.Duration(value) * time.Second)
	default:
		return ""
	}

	return result.Format(layout)
}
