package common

import (
	"testing"

	"bytes"
	"os"
	"path/filepath"
)

func consumeString(s string) (out chan byte) {
	out = make(chan byte)
	go func() {
		for _, ch := range s {
			out <- byte(ch)
		}
		close(out)
	}()
	return
}

func uncommentFromString(s string) string {
	inCh := consumeString(s)
	out := uncomment(inCh)

	result := ""
	for reading := true; reading; {
		if item, ok := <-out; !ok {
			break
		} else {
			result += string(item)
		}
	}
	return result
}

func TestParserUncomment(t *testing.T) {
	var result string

	test_0 := "this is a straight-up line"
	result_0 := test_0

	result = uncommentFromString(test_0)
	if result != result_0 {
		t.Errorf("Expected %#v, received %#v", result_0, result)
	}

	test_1 := "this is a straight-up line with a newline\n"
	result_1 := test_1

	result = uncommentFromString(test_1)
	if result != result_1 {
		t.Errorf("Expected %#v, received %#v", result_1, result)
	}

	test_2 := "this line has a comment # at its end"
	result_2 := "this line has a comment "

	result = uncommentFromString(test_2)
	if result != result_2 {
		t.Errorf("Expected %#v, received %#v", result_2, result)
	}

	test_3 := "# this whole line is commented"
	result_3 := ""

	result = uncommentFromString(test_3)
	if result != result_3 {
		t.Errorf("Expected %#v, received %#v", result_3, result)
	}

	test_4 := "this\nhas\nmultiple\nlines"
	result_4 := test_4

	result = uncommentFromString(test_4)
	if result != result_4 {
		t.Errorf("Expected %#v, received %#v", result_4, result)
	}

	test_5 := "this only has\n# one line"
	result_5 := "this only has\n"
	result = uncommentFromString(test_5)
	if result != result_5 {
		t.Errorf("Expected %#v, received %#v", result_5, result)
	}

	test_6 := "this is\npartially # commented"
	result_6 := "this is\npartially "

	result = uncommentFromString(test_6)
	if result != result_6 {
		t.Errorf("Expected %#v, received %#v", result_6, result)
	}

	test_7 := "this # has\nmultiple # lines\ncommented # out"
	result_7 := "this \nmultiple \ncommented "

	result = uncommentFromString(test_7)
	if result != result_7 {
		t.Errorf("Expected %#v, received %#v", result_7, result)
	}
}

func tokenizeDhcpConfigFromString(s string) []string {
	inCh := consumeString(s)
	out := tokenizeDhcpConfig(inCh)

	result := make([]string, 0)
	for {
		if item, ok := <-out; !ok {
			break
		} else {
			result = append(result, item)
		}
	}
	return result
}

func TestParserTokenizeDhcp(t *testing.T) {

	test_1 := `
subnet 127.0.0.0 netmask 255.255.255.252 {
    item 1234 5678;
	tabbed item-1 1234;
    quoted item-2 "hola mundo.";
}
`
	expected := []string{
		"subnet", "127.0.0.0", "netmask", "255.255.255.252", "{",
		"item", "1234", "5678", ";",
		"tabbed", "item-1", "1234", ";",
		"quoted", "item-2", "\"hola mundo.\"", ";",
		"}",
	}
	result := tokenizeDhcpConfigFromString(test_1)

	t.Logf("testing for: %v", expected)
	t.Logf("checking out: %v", result)
	if len(result) != len(expected) {
		t.Fatalf("length of token lists do not match (%d != %d)", len(result), len(expected))
	}

	for index := range expected {
		if expected[index] != result[index] {
			t.Errorf("unexpected token at index %d: %v != %v", index, expected[index], result[index])
		}
	}
}

func consumeTokens(tokes []string) chan string {
	out := make(chan string)
	go func() {
		for _, item := range tokes {
			out <- item
		}
		out <- ";"
		close(out)
	}()
	return out
}

func TestParserDhcpParameters(t *testing.T) {
	var ch chan string

	test_1 := []string{"option", "whee", "whooo"}
	ch = consumeTokens(test_1)

	result := parseTokenParameter(ch)
	if result.name != "option" {
		t.Errorf("expected name %s, got %s", test_1[0], result.name)
	}
	if len(result.operand) == 2 {
		if result.operand[0] != "whee" {
			t.Errorf("expected operand[%d] as %s, got %s", 0, "whee", result.operand[0])
		}
		if result.operand[1] != "whooo" {
			t.Errorf("expected operand[%d] as %s, got %s", 0, "whooo", result.operand[1])
		}
	} else {
		t.Errorf("expected %d operands, got %d", 2, len(result.operand))
	}

	test_2 := []string{"whaaa", "whoaaa", ";", "wooops"}
	ch = consumeTokens(test_2)

	result = parseTokenParameter(ch)
	if result.name != "whaaa" {
		t.Errorf("expected name %s, got %s", test_2[0], result.name)
	}
	if len(result.operand) == 1 {
		if result.operand[0] != "whoaaa" {
			t.Errorf("expected operand[%d] as %s, got %s", 0, "whoaaa", result.operand[0])
		}
	} else {
		t.Errorf("expected %d operands, got %d", 1, len(result.operand))
	}

	test_3 := []string{"optionz", "only", "{", "culled"}
	ch = consumeTokens(test_3)

	result = parseTokenParameter(ch)
	if result.name != "optionz" {
		t.Errorf("expected name %s, got %s", test_3[0], result.name)
	}
	if len(result.operand) == 1 {
		if result.operand[0] != "only" {
			t.Errorf("expected operand[%d] as %s, got %s", 0, "only", result.operand[0])
		}
	} else {
		t.Errorf("expected %d operands, got %d", 1, len(result.operand))
	}
}

func consumeDhcpConfig(items []string) (tkGroup, error) {
	out := make(chan string)
	tch := consumeTokens(items)

	go func() {
		for item := range tch {
			out <- item
		}
		close(out)
	}()

	return parseDhcpConfig(out)
}

func compareSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestParserDhcpConfigParse(t *testing.T) {
	test_1 := []string{
		"allow", "unused-option", ";",
		"lease-option", "1234", ";",
		"more", "options", "hi", ";",
	}
	result_1, err := consumeDhcpConfig(test_1)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if len(result_1.params) != 4 {
		t.Fatalf("expected %d params, got %d", 3, len(result_1.params))
	}
	if result_1.params[0].name != "allow" {
		t.Errorf("expected %s, got %s", "allow", result_1.params[0].name)
	}
	if !compareSlice(result_1.params[0].operand, []string{"unused-option"}) {
		t.Errorf("unexpected options parsed: %v", result_1.params[0].operand)
	}
	if result_1.params[1].name != "lease-option" {
		t.Errorf("expected %s, got %s", "lease-option", result_1.params[1].name)
	}
	if !compareSlice(result_1.params[1].operand, []string{"1234"}) {
		t.Errorf("unexpected options parsed: %v", result_1.params[1].operand)
	}
	if result_1.params[2].name != "more" {
		t.Errorf("expected %s, got %s", "lease-option", result_1.params[2].name)
	}
	if !compareSlice(result_1.params[2].operand, []string{"options", "hi"}) {
		t.Errorf("unexpected options parsed: %v", result_1.params[2].operand)
	}

	test_2 := []string{
		"first-option", ";",
		"child", "group", "{",
		"blah", ";",
		"meh", ";",
		"}",
		"hidden", "option", "57005", ";",
		"host", "device", "two", "{",
		"skipped", "option", ";",
		"more", "skipped", "options", ";",
		"}",
		"last", "option", "but", "unterminated",
	}
	result_2, err := consumeDhcpConfig(test_2)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if len(result_2.groups) != 2 {
		t.Fatalf("expected %d groups, got %d", 2, len(result_2.groups))
	}

	if len(result_2.params) != 3 {
		t.Errorf("expected %d options, got %d", 3, len(result_2.params))
	}

	group0 := result_2.groups[0]
	if group0.id.name != "child" {
		t.Errorf("expected group %s, got %s", "child", group0.id.name)
	}
	if len(group0.id.operand) != 1 {
		t.Errorf("expected group operand %d, got %d", 1, len(group0.params))
	}
	if len(group0.params) != 2 {
		t.Errorf("expected group params %d, got %d", 2, len(group0.params))
	}

	group1 := result_2.groups[1]
	if group1.id.name != "host" {
		t.Errorf("expected group %s, got %s", "host", group1.id.name)
	}
	if len(group1.id.operand) != 2 {
		t.Errorf("expected group operand %d, got %d", 2, len(group1.params))
	}
	if len(group1.params) != 2 {
		t.Errorf("expected group params %d, got %d", 2, len(group1.params))
	}
}

func TestParserReadDhcpConfig(t *testing.T) {
	expected := []string{
		`{global}
grants : map[unknown-clients:0]
parameters : map[default-lease-time:1800 max-lease-time:7200]
`,

		`{subnet4 172.33.33.0/24},{global}
address : range4:172.33.33.128-172.33.33.254
options : map[broadcast-address:172.33.33.255 routers:172.33.33.2]
grants : map[unknown-clients:0]
parameters : map[default-lease-time:2400 max-lease-time:9600]
`,

		`{host name:vmnet8},{global}
address : hardware-address:ethernet[00:50:56:c0:00:08],fixed-address4:172.33.33.1
options : map[domain-name:"packer.test"]
grants : map[unknown-clients:0]
parameters : map[default-lease-time:1800 max-lease-time:7200]
`,
	}

	f, err := os.Open(filepath.Join("testdata", "dhcpd-example.conf"))
	if err != nil {
		t.Fatalf("Unable to open dhcpd.conf sample: %s", err)
	}
	defer f.Close()

	config, err := ReadDhcpConfiguration(f)
	if err != nil {
		t.Fatalf("Unable to read dhcpd.conf samplpe: %s", err)
	}

	if len(config) != 3 {
		t.Fatalf("expected %d entries, got %d", 3, len(config))
	}

	for index, item := range config {
		if item.repr() != expected[index] {
			t.Errorf("Parsing of declaration %d did not match what was expected", index)
			t.Logf("Result from parsing:\n%s", item.repr())
			t.Logf("Expected to parse:\n%s", expected[index])
		}
	}
}

func TestParserTokenizeNetworkMap(t *testing.T) {

	test_1 := "group.attribute = \"string\""
	expected := []string{
		"group.attribute", "=", "\"string\"",
	}
	result := tokenizeDhcpConfigFromString(test_1)
	if len(result) != len(expected) {
		t.Fatalf("length of token lists do not match (%d != %d)", len(result), len(expected))
	}

	for index := range expected {
		if expected[index] != result[index] {
			t.Errorf("unexpected token at index %d: %v != %v", index, expected[index], result[index])
		}
	}

	test_2 := "attribute == \""
	expected = []string{
		"attribute", "==", "\"",
	}
	result = tokenizeDhcpConfigFromString(test_2)
	if len(result) != len(expected) {
		t.Fatalf("length of token lists do not match (%d != %d)", len(result), len(expected))
	}

	test_3 := "attribute ....... ======\nnew lines should make no difference"
	expected = []string{
		"attribute", ".......", "======", "new", "lines", "should", "make", "no", "difference",
	}
	result = tokenizeDhcpConfigFromString(test_3)
	if len(result) != len(expected) {
		t.Fatalf("length of token lists do not match (%d != %d)", len(result), len(expected))
	}

	test_4 := "\t\t\t\t    thishadwhitespacebeforebeingparsed\t \t \t \t\n\n"
	expected = []string{
		"thishadwhitespacebeforebeingparsed",
	}
	result = tokenizeDhcpConfigFromString(test_4)
	if len(result) != len(expected) {
		t.Fatalf("length of token lists do not match (%d != %d)", len(result), len(expected))
	}
}

func TestParserReadNetworkMap(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "netmap-example.conf"))
	if err != nil {
		t.Fatalf("Unable to open netmap.conf sample: %s", err)
	}
	defer f.Close()

	netmap, err := ReadNetworkMap(f)
	if err != nil {
		t.Fatalf("Unable to read netmap.conf samplpe: %s", err)
	}

	expected_keys := []string{"device", "name"}
	for _, item := range netmap {
		for _, name := range expected_keys {
			_, ok := item[name]
			if !ok {
				t.Errorf("unable to find expected key %v in map: %v", name, item)
			}
		}
	}

	expected_vmnet0 := [][]string{
		[]string{"device", "vmnet0"},
		[]string{"name", "Bridged"},
	}
	for _, item := range netmap {
		if item["device"] != "vmnet0" {
			continue
		}
		for _, expectpair := range expected_vmnet0 {
			name := expectpair[0]
			value := expectpair[1]
			if item[name] != value {
				t.Errorf("expected value %v for attribute %v, got %v", value, name, item[name])
			}
		}
	}

	expected_vmnet1 := [][]string{
		[]string{"device", "vmnet1"},
		[]string{"name", "HostOnly"},
	}
	for _, item := range netmap {
		if item["device"] != "vmnet1" {
			continue
		}
		for _, expectpair := range expected_vmnet1 {
			name := expectpair[0]
			value := expectpair[1]
			if item[name] != value {
				t.Errorf("expected value %v for attribute %v, got %v", value, name, item[name])
			}
		}
	}

	expected_vmnet8 := [][]string{
		[]string{"device", "vmnet8"},
		[]string{"name", "NAT"},
	}
	for _, item := range netmap {
		if item["device"] != "vmnet8" {
			continue
		}
		for _, expectpair := range expected_vmnet8 {
			name := expectpair[0]
			value := expectpair[1]
			if item[name] != value {
				t.Errorf("expected value %v for attribute %v, got %v", value, name, item[name])
			}
		}
	}
}

func collectIntoString(in chan byte) string {
	result := ""
	for item := range in {
		result += string(item)
	}
	return result
}

func TestParserConsumeUntilSentinel(t *testing.T) {

	test_1 := "consume until a semicolon; yeh?"
	expected_1 := "consume until a semicolon"

	ch := consumeString(test_1)
	resultch, _ := consumeUntilSentinel(';', ch)
	result := string(resultch)
	if expected_1 != result {
		t.Errorf("expected %#v, got %#v", expected_1, result)
	}

	test_2 := "; this is only a semi"
	expected_2 := ""

	ch = consumeString(test_2)
	resultch, _ = consumeUntilSentinel(';', ch)
	result = string(resultch)
	if expected_2 != result {
		t.Errorf("expected %#v, got %#v", expected_2, result)
	}
}

func TestParserFilterCharacters(t *testing.T) {

	test_1 := []string{" ", "ignore all spaces"}
	expected_1 := "ignoreallspaces"

	ch := consumeString(test_1[1])
	outch := filterOutCharacters(bytes.NewBufferString(test_1[0]).Bytes(), ch)
	result := collectIntoString(outch)
	if result != expected_1 {
		t.Errorf("expected %#v, got %#v", expected_1, result)
	}

	test_2 := []string{"\n\v\t\r ", "ignore\nall\rwhite\v\v space                "}
	expected_2 := "ignoreallwhitespace"

	ch = consumeString(test_2[1])
	outch = filterOutCharacters(bytes.NewBufferString(test_2[0]).Bytes(), ch)
	result = collectIntoString(outch)
	if result != expected_2 {
		t.Errorf("expected %#v, got %#v", expected_2, result)
	}
}

func TestParserConsumeOpenClosePair(t *testing.T) {
	test_1 := "(everything)"
	expected_1 := []string{"", test_1}

	testch := consumeString(test_1)
	prefix, ch := consumeOpenClosePair('(', ')', testch)
	if string(prefix) != expected_1[0] {
		t.Errorf("expected prefix %#v, got %#v", expected_1[0], prefix)
	}
	result := collectIntoString(ch)
	if result != expected_1[1] {
		t.Errorf("expected %#v, got %#v", expected_1[1], test_1)
	}

	test_2 := "prefixed (everything)"
	expected_2 := []string{"prefixed ", "(everything)"}

	testch = consumeString(test_2)
	prefix, ch = consumeOpenClosePair('(', ')', testch)
	if string(prefix) != expected_2[0] {
		t.Errorf("expected prefix %#v, got %#v", expected_2[0], prefix)
	}
	result = collectIntoString(ch)
	if result != expected_2[1] {
		t.Errorf("expected %#v, got %#v", expected_2[1], test_2)
	}

	test_3 := "this(is()suffixed"
	expected_3 := []string{"this", "(is()"}

	testch = consumeString(test_3)
	prefix, ch = consumeOpenClosePair('(', ')', testch)
	if string(prefix) != expected_3[0] {
		t.Errorf("expected prefix %#v, got %#v", expected_3[0], prefix)
	}
	result = collectIntoString(ch)
	if result != expected_3[1] {
		t.Errorf("expected %#v, got %#v", expected_3[1], test_2)
	}
}

func TestParserCombinators(t *testing.T) {

	test_1 := "across # ignore\nmultiple lines;"
	expected_1 := "across multiple lines"

	ch := consumeString(test_1)
	inch := uncomment(ch)
	whch := filterOutCharacters([]byte{'\n'}, inch)
	resultch, _ := consumeUntilSentinel(';', whch)
	result := string(resultch)
	if expected_1 != result {
		t.Errorf("expected %#v, got %#v", expected_1, result)
	}

	test_2 := "lease blah {\n    blah\r\n# skipping this line\nblahblah  # ignore semicolon;\n last item;\n\n };;;;;;"
	expected_2 := []string{"lease blah ", "{    blahblahblah   last item; }"}

	ch = consumeString(test_2)
	inch = uncomment(ch)
	whch = filterOutCharacters([]byte{'\n', '\v', '\r'}, inch)
	prefix, pairch := consumeOpenClosePair('{', '}', whch)

	result = collectIntoString(pairch)
	if string(prefix) != expected_2[0] {
		t.Errorf("expected prefix %#v, got %#v", expected_2[0], prefix)
	}
	if result != expected_2[1] {
		t.Errorf("expected %#v, got %#v", expected_2[1], result)
	}

	test_3 := "lease blah { # comment\n item 1;\n item 2;\n } not imortant"
	expected_3_prefix := "lease blah "
	expected_3 := []string{"{  item 1", " item 2", " }"}

	sch := consumeString(test_3)
	inch = uncomment(sch)
	wch := filterOutCharacters([]byte{'\n', '\v', '\r'}, inch)
	lease, itemch := consumeOpenClosePair('{', '}', wch)
	if string(lease) != expected_3_prefix {
		t.Errorf("expected %#v, got %#v", expected_3_prefix, string(lease))
	}

	result_3 := []string{}
	for reading := true; reading; {
		item, ok := consumeUntilSentinel(';', itemch)
		result_3 = append(result_3, string(item))
		if !ok {
			reading = false
		}
	}

	for index := range expected_3 {
		if expected_3[index] != result_3[index] {
			t.Errorf("expected index %d as %#v, got %#v", index, expected_3[index], result_3[index])
		}
	}
}
