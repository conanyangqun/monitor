package params

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestNewParameterResolver(t *testing.T) {
	pr := NewParameterResolver()
	if pr == nil {
		t.Fatal("NewParameterResolver() returned nil")
	}
	if pr.randSource == nil {
		t.Error("randSource should not be nil")
	}
	if pr.currentTime.IsZero() {
		t.Error("currentTime should not be zero")
	}
}

func TestResolveSpecialParameter_UUID(t *testing.T) {
	pr := NewParameterResolver()

	// Test UUID generation
	uuid := pr.resolveSpecialParameter("uuid")
	if len(uuid) != 36 {
		t.Errorf("UUID length should be 36, got %d", len(uuid))
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).MatchString(uuid) {
		t.Errorf("Invalid UUID format: %s", uuid)
	}

	// Test short UUID generation
	uuidShort := pr.resolveSpecialParameter("uuid_short")
	if len(uuidShort) != 32 {
		t.Errorf("Short UUID length should be 32, got %d", len(uuidShort))
	}
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(uuidShort) {
		t.Errorf("Invalid short UUID format: %s", uuidShort)
	}
}

func TestResolveSpecialParameter_Random(t *testing.T) {
	pr := NewParameterResolver()

	// Test basic random number
	rand := pr.resolveSpecialParameter("rand")
	if randInt, err := strconv.Atoi(rand); err != nil || randInt < 0 || randInt >= 1000000 {
		t.Errorf("Random number should be between 0 and 999999, got %s", rand)
	}

	// Test random int
	randInt := pr.resolveSpecialParameter("rand_int")
	if ri, err := strconv.Atoi(randInt); err != nil || ri < 0 || ri >= 2147483647 {
		t.Errorf("Random int should be between 0 and 2147483646, got %s", randInt)
	}

	// Test random range
	randRange := pr.resolveSpecialParameter("rand(10,20)")
	if rr, err := strconv.Atoi(randRange); err != nil || rr < 10 || rr >= 20 {
		t.Errorf("Random range should be between 10 and 19, got %s", randRange)
	}

	// Test invalid random range
	invalidRange := pr.resolveSpecialParameter("rand(invalid,range)")
	if ir, err := strconv.Atoi(invalidRange); err != nil || ir < 0 || ir >= 1000000 {
		t.Errorf("Invalid random range should fallback to default range, got %s", invalidRange)
	}
}

func TestResolveSpecialParameter_RandomString(t *testing.T) {
	pr := NewParameterResolver()

	// Test default random string
	randStr := pr.resolveSpecialParameter("rand_str")
	if len(randStr) != 8 {
		t.Errorf("Random string length should be 8, got %d", len(randStr))
	}

	// Test secure random string
	secureStr := pr.resolveSpecialParameter("rand_str_secure")
	if len(secureStr) != 16 {
		t.Errorf("Secure random string length should be 16, got %d", len(secureStr))
	}

	// Test custom length random string
	customStr := pr.resolveSpecialParameter("rand_str(12)")
	if len(customStr) != 12 {
		t.Errorf("Custom random string length should be 12, got %d", len(customStr))
	}

	// Test hex random string
	hexStr := pr.resolveSpecialParameter("rand_hex(8)")
	if len(hexStr) != 8 {
		t.Errorf("Hex string length should be 8, got %d", len(hexStr))
	}
	if !regexp.MustCompile(`^[0-9a-f]+$`).MatchString(hexStr) {
		t.Errorf("Hex string should only contain hex characters, got %s", hexStr)
	}
}

func TestResolveSpecialParameter_Environment(t *testing.T) {
	pr := NewParameterResolver()

	// Set test environment variable
	testKey := "TEST_PARAM_KEY"
	testValue := "test_value"
	if err := os.Setenv(testKey, testValue); err != nil {
		return
	}
	defer func(key string) {
		if err := os.Unsetenv(key); err != nil {
			t.Errorf("Failed to unset environment variable %s: %v", key, err)
		}
	}(testKey)

	// Test environment variable resolution
	envResult := pr.resolveSpecialParameter("env(" + testKey + ")")
	if envResult != testValue {
		t.Errorf("Environment variable should be %s, got %s", testValue, envResult)
	}

	// Test non-existent environment variable
	nonExistent := pr.resolveSpecialParameter("env(NON_EXISTENT_VAR)")
	if nonExistent != "" {
		t.Errorf("Non-existent environment variable should return empty string, got %s", nonExistent)
	}
}

func TestResolveSpecialParameter_Sequence(t *testing.T) {
	pr := NewParameterResolver()

	// Test sequence number
	seq := pr.resolveSpecialParameter("seq")
	if _, err := strconv.Atoi(seq); err != nil {
		t.Errorf("Sequence should be a valid integer, got %s", seq)
	}

	// Test daily sequence
	seqDaily := pr.resolveSpecialParameter("seq_daily")
	if _, err := strconv.Atoi(seqDaily); err != nil {
		t.Errorf("Daily sequence should be a valid integer, got %s", seqDaily)
	}
}

func TestResolveSpecialParameter_Hash(t *testing.T) {
	pr := NewParameterResolver()

	// Test short hash
	hashShort := pr.resolveSpecialParameter("hash_short")
	if !regexp.MustCompile(`^[0-9a-f]+$`).MatchString(hashShort) {
		t.Errorf("Short hash should be hex format, got %s", hashShort)
	}

	// Test MD5-like hash
	hashMD5 := pr.resolveSpecialParameter("hash_md5_like")
	if len(hashMD5) != 32 {
		t.Errorf("MD5-like hash should be 32 characters, got %d", len(hashMD5))
	}
}

func TestResolveSpecialParameter_Network(t *testing.T) {
	pr := NewParameterResolver()

	// Test local IP (maybe empty if no network interface)
	localIP := pr.resolveSpecialParameter("local_ip")
	if localIP != "" && !regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`).MatchString(localIP) {
		t.Errorf("Local IP should be valid IPv4 format or empty, got %s", localIP)
	}

	// Test hostname (should not be empty)
	hostname := pr.resolveSpecialParameter("hostname")
	if hostname == "" {
		t.Error("Hostname should not be empty")
	}

	// Test user agent
	userAgent := pr.resolveSpecialParameter("user_agent")
	if userAgent == "" {
		t.Error("User agent should not be empty")
	}

	// Test HTTP method
	httpMethod := pr.resolveSpecialParameter("http_method")
	validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true, "HEAD": true, "OPTIONS": true}
	if !validMethods[httpMethod] {
		t.Errorf("HTTP method should be valid, got %s", httpMethod)
	}
}

func TestResolveSpecialParameter_Encoding(t *testing.T) {
	pr := NewParameterResolver()

	// Test Base64 encoding
	base64Result := pr.resolveSpecialParameter("base64(hello)")
	expected := "aGVsbG8="
	if base64Result != expected {
		t.Errorf("Base64 encoding should be %s, got %s", expected, base64Result)
	}

	// Test URL encoding
	urlResult := pr.resolveSpecialParameter("url_encode(hello world)")
	expectedURL := "hello+world"
	if urlResult != expectedURL {
		t.Errorf("URL encoding should be %s, got %s", expectedURL, urlResult)
	}

	// Test JSON escape
	jsonResult := pr.resolveSpecialParameter("json_escape(hello\"world)")
	expectedJSON := "hello\\\"world"
	if jsonResult != expectedJSON {
		t.Errorf("JSON escape should be %s, got %s", expectedJSON, jsonResult)
	}
}

func TestResolveSpecialParameter_Math(t *testing.T) {
	pr := NewParameterResolver()

	// Test addition
	addResult := pr.resolveSpecialParameter("add(5,3)")
	if !strings.Contains(addResult, "8") {
		t.Errorf("Addition result should contain 8, got %s", addResult)
	}

	// Test subtraction
	subResult := pr.resolveSpecialParameter("sub(10,3)")
	if !strings.Contains(subResult, "7") {
		t.Errorf("Subtraction result should contain 7, got %s", subResult)
	}

	// Test multiplication
	mulResult := pr.resolveSpecialParameter("mul(4,3)")
	if !strings.Contains(mulResult, "12") {
		t.Errorf("Multiplication result should contain 12, got %s", mulResult)
	}

	// Test division
	divResult := pr.resolveSpecialParameter("div(15,3)")
	if !strings.Contains(divResult, "5") {
		t.Errorf("Division result should contain 5, got %s", divResult)
	}

	// Test division by zero
	divZeroResult := pr.resolveSpecialParameter("div(10,0)")
	if divZeroResult != "0" {
		t.Errorf("Division by zero should return 0, got %s", divZeroResult)
	}
}

func TestResolveSpecialParameter_TextProcessing(t *testing.T) {
	pr := NewParameterResolver()

	// Test uppercase
	upperResult := pr.resolveSpecialParameter("upper(hello)")
	if upperResult != "HELLO" {
		t.Errorf("Uppercase should be HELLO, got %s", upperResult)
	}

	// Test lowercase
	lowerResult := pr.resolveSpecialParameter("lower(HELLO)")
	if lowerResult != "hello" {
		t.Errorf("Lowercase should be hello, got %s", lowerResult)
	}

	// Test reverse
	reverseResult := pr.resolveSpecialParameter("reverse(hello)")
	if reverseResult != "olleh" {
		t.Errorf("Reverse should be olleh, got %s", reverseResult)
	}

	// Test substring
	substrResult := pr.resolveSpecialParameter("substr(hello,1,3)")
	if substrResult != "ell" {
		t.Errorf("Substring should be ell, got %s", substrResult)
	}
}

func TestResolveSpecialParameter_Colors(t *testing.T) {
	pr := NewParameterResolver()

	// Test hex color
	hexColor := pr.resolveSpecialParameter("color_hex")
	if !regexp.MustCompile(`^#[0-9a-f]{6}$`).MatchString(hexColor) {
		t.Errorf("Hex color should match pattern #xxxxxx, got %s", hexColor)
	}

	// Test RGB color
	rgbColor := pr.resolveSpecialParameter("color_rgb")
	if !regexp.MustCompile(`^rgb\(\d+,\d+,\d+\)$`).MatchString(rgbColor) {
		t.Errorf("RGB color should match pattern rgb(x,y,z), got %s", rgbColor)
	}

	// Test HSL color
	hslColor := pr.resolveSpecialParameter("color_hsl")
	if !regexp.MustCompile(`^hsl\(\d+,\d+%,\d+%\)$`).MatchString(hslColor) {
		t.Errorf("HSL color should match pattern hsl(x,y%%,z%%), got %s", hslColor)
	}
}

func TestResolveSpecialParameter_FileTypes(t *testing.T) {
	pr := NewParameterResolver()

	// Test MIME type
	mimeType := pr.resolveSpecialParameter("mime_type")
	if mimeType == "" {
		t.Error("MIME type should not be empty")
	}

	// Test file extension
	fileExt := pr.resolveSpecialParameter("file_ext")
	if !strings.HasPrefix(fileExt, ".") {
		t.Errorf("File extension should start with dot, got %s", fileExt)
	}
}

func TestResolveSpecialParameter_FakeData(t *testing.T) {
	pr := NewParameterResolver()

	// Test fake email
	fakeEmail := pr.resolveSpecialParameter("fake_email")
	if !strings.Contains(fakeEmail, "@") {
		t.Errorf("Fake email should contain @, got %s", fakeEmail)
	}

	// Test fake phone
	fakePhone := pr.resolveSpecialParameter("fake_phone")
	if !strings.HasPrefix(fakePhone, "+1-800-") {
		t.Errorf("Fake phone should start with +1-800-, got %s", fakePhone)
	}

	// Test fake name
	fakeName := pr.resolveSpecialParameter("fake_name")
	if !strings.Contains(fakeName, " ") {
		t.Errorf("Fake name should contain space, got %s", fakeName)
	}

	// Test fake domain
	fakeDomain := pr.resolveSpecialParameter("fake_domain")
	if !strings.HasPrefix(fakeDomain, "www.") {
		t.Errorf("Fake domain should start with www., got %s", fakeDomain)
	}
}

func TestResolveSpecialParameter_TimeFormat(t *testing.T) {
	pr := NewParameterResolver()

	// Test year format
	yearResult := pr.resolveSpecialParameter("%Y")
	currentYear := strconv.Itoa(pr.currentTime.Year())
	if yearResult != currentYear {
		t.Errorf("Year format should be %s, got %s", currentYear, yearResult)
	}

	// Test month format
	monthResult := pr.resolveSpecialParameter("%m")
	expectedMonth := pr.currentTime.Format("01")
	if monthResult != expectedMonth {
		t.Errorf("Month format should be %s, got %s", expectedMonth, monthResult)
	}
}

func TestResolveParameters(t *testing.T) {
	pr := NewParameterResolver()

	// Test simple parameter replacement
	result := pr.ResolveParameters("Hello {{uuid}}")
	if !strings.HasPrefix(result, "Hello ") {
		t.Errorf("Result should start with 'Hello ', got %s", result)
	}

	// Test multiple parameters
	result = pr.ResolveParameters("{{%Y}}-{{%m}}-{{%d}}")
	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(result) {
		t.Errorf("Date format should match YYYY-MM-DD, got %s", result)
	}

	// Test no parameters
	result = pr.ResolveParameters("No parameters here")
	if result != "No parameters here" {
		t.Errorf("String without parameters should remain unchanged, got %s", result)
	}
}

func TestMaskSensitiveValue(t *testing.T) {
	pr := NewParameterResolver()

	// Test short value masking
	result := pr.maskSensitiveValue("abc")
	if result != "***" {
		t.Errorf("Short value should be fully masked, got %s", result)
	}

	// Test longer value masking
	result = pr.maskSensitiveValue("secret123456")
	if !strings.HasPrefix(result, "s") || !strings.HasSuffix(result, "6") {
		t.Errorf("Long value should show first 1 and last 1 char, got %s", result)
	}
}

func TestHighlightChanges(t *testing.T) {
	pr := NewParameterResolver()

	// Test highlighting with parameters
	result, segments := pr.HighlightChanges("Hello {{uuid}} world")

	if len(segments) != 3 {
		t.Errorf("Should have 3 segments, got %d", len(segments))
	}

	if segments[0].Text != "Hello " || segments[0].IsHighlight {
		t.Errorf("First segment should be 'Hello ' and not highlighted")
	}

	if !segments[1].IsHighlight {
		t.Errorf("Second segment should be highlighted")
	}

	if segments[2].Text != " world" || segments[2].IsHighlight {
		t.Errorf("Third segment should be ' world' and not highlighted")
	}

	// Test no parameters
	result, segments = pr.HighlightChanges("No parameters")
	if result != "No parameters" || len(segments) != 0 {
		t.Errorf("String without parameters should return original string and no segments")
	}
}

func TestGetResolvedValue(t *testing.T) {
	pr := NewParameterResolver()

	result := pr.GetResolvedValue("{{%Y}}")
	currentYear := strconv.Itoa(pr.currentTime.Year())
	if result != currentYear {
		t.Errorf("Resolved value should be current year, got %s", result)
	}
}

func TestGetOriginalValue(t *testing.T) {
	pr := NewParameterResolver()

	original := "{{uuid}}"
	result := pr.GetOriginalValue(original)
	if result != original {
		t.Errorf("Original value should remain unchanged, got %s", result)
	}
}

func TestResolveSpecialParameterForDisplay(t *testing.T) {
	pr := NewParameterResolver()

	// Set up test environment variable with sensitive name
	testKey := "API_SECRET"
	testValue := "very_secret_key_123456"
	if err := os.Setenv(testKey, testValue); err != nil {
		return
	}
	defer func(key string) {
		if err := os.Unsetenv(key); err != nil {
			t.Errorf("Failed to unset environment variable %s: %v", key, err)
		}
	}(testKey)

	// Test that sensitive env vars are masked for display
	result := pr.resolveSpecialParameterWithSecret("env(" + testKey + ")")
	if result == testValue {
		t.Error("Sensitive environment variable should be masked for display")
	}
	if !strings.Contains(result, "*") {
		t.Errorf("Masked value should contain asterisks, got %s", result)
	}
}

func TestTimeCalculation(t *testing.T) {
	pr := NewParameterResolver()

	// Test time addition
	result := pr.resolveSpecialParameter("time_add(2023-01-01 12:00:00,3600)")
	if result != "2023-01-01 13:00:00" {
		t.Errorf("Time addition should add 1 hour, got %s", result)
	}

	// Test time subtraction
	result = pr.resolveSpecialParameter("time_sub(2023-01-01 12:00:00,1800)")
	if result != "2023-01-01 11:30:00" {
		t.Errorf("Time subtraction should subtract 30 minutes, got %s", result)
	}

	// Test invalid time format
	result = pr.resolveSpecialParameter("time_add(invalid,3600)")
	if result != "" {
		t.Errorf("Invalid time format should return empty string, got %s", result)
	}
}

// Test nested tag resolution
func TestNestedTagResolution(t *testing.T) {
	pr := NewParameterResolver()

	// Test simple nested tags
	result := pr.ResolveParameters("{{base64({{uuid}})}}")
	if len(result) == 0 {
		t.Error("Nested base64(uuid) should produce a result")
	}

	// Test multiple levels of nesting
	result = pr.ResolveParameters("{{upper({{base64(test)}})}}")
	expected := "DGVZDA==" // base64 of "test" in uppercase
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test nested with random content
	result = pr.ResolveParameters("{{url_encode({{fake_email}})}}")
	if !strings.Contains(result, "%40") { // @ symbol should be encoded as %40
		t.Error("URL encoded email should contain %40")
	}

	// Test complex nesting
	result = pr.ResolveParameters("{{base64({{upper({{fake_name}})}})}}")
	if len(result) == 0 {
		t.Error("Complex nested expression should produce a result")
	}

	// Test mathematical operations with nested values
	result = pr.ResolveParameters("{{add({{rand(1,10)}},{{rand(1,10)}})}}")
	if len(result) == 0 {
		t.Error("Nested mathematical operation should produce a result")
	}

	// Test string manipulation with nested content
	result = pr.ResolveParameters("{{substr({{uuid}},0,8)}}")
	if len(result) != 8 {
		t.Errorf("Substring of UUID should be 8 characters, got %d", len(result))
	}
}

// Test cycle detection
func TestNestedTagCycleDetection(t *testing.T) {
	pr := NewParameterResolver()

	// Test that cycles are properly detected and handled
	result := pr.ResolveParameters("{{upper({{upper(test)}})}}")
	expected := "TEST"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test self-referencing prevention (should not cause infinite loop)
	result = pr.ResolveParameters("test")
	if result != "test" {
		t.Errorf("Non-parameterized string should remain unchanged")
	}
}

// Test recursion depth limits
func TestNestedTagDepthLimit(t *testing.T) {
	pr := NewParameterResolver()

	// Create a deeply nested structure
	deepNested := "test"
	for i := 0; i < 15; i++ { // More than maxDepth (10)
		deepNested = fmt.Sprintf("{{upper(%s)}}", deepNested)
	}

	result := pr.ResolveParameters(deepNested)
	// Should not crash and should return some reasonable result
	if len(result) == 0 {
		t.Error("Deep nesting should not result in empty string")
	}
}

// Test mixed nested and non-nested parameters
func TestMixedNestedParameters(t *testing.T) {
	pr := NewParameterResolver()

	// Mix nested and simple parameters
	result := pr.ResolveParameters("ID: {{uuid}} - Encoded: {{base64({{fake_name}})}} - Time: {{%Y}}")

	parts := strings.Split(result, " - ")
	if len(parts) != 3 {
		t.Errorf("Expected 3 parts separated by ' - ', got %d", len(parts))
	}

	if !strings.HasPrefix(parts[0], "ID: ") {
		t.Error("First part should start with 'ID: '")
	}

	if !strings.HasPrefix(parts[1], "Encoded: ") {
		t.Error("Second part should start with 'Encoded: '")
	}

	if !strings.HasPrefix(parts[2], "Time: ") {
		t.Error("Third part should start with 'Time: '")
	}
}
