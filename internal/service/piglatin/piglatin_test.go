package piglatin

import (
	"testing"
)

func TestToPigLatin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single word starting with vowel",
			input:    "apple",
			expected: "appleway",
		},
		{
			name:     "single word starting with consonant",
			input:    "banana",
			expected: "ananabay",
		},
		{
			name:     "multiple words",
			input:    "hello world",
			expected: "ellohay orldway",
		},
		{
			name:     "sentence with punctuation",
			input:    "Hello, world!",
			expected: "Ellohay, orldway!",
		},
		{
			name:     "capitalized word",
			input:    "Hello",
			expected: "Ellohay",
		},
		{
			name:     "word with no vowels",
			input:    "rhythm",
			expected: "rhythmay",
		},
		{
			name:     "complex sentence",
			input:    "The quick brown fox jumps over the lazy dog.",
			expected: "Ethay uickqay ownbray oxfay umpsjay overway ethay azylay ogday.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPigLatin(tt.input)
			if result != tt.expected {
				t.Errorf("ToPigLatin(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFromPigLatin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "word with way suffix",
			input:    "appleway",
			expected: "apple",
		},
		{
			name:     "word with ay suffix",
			input:    "ananabay",
			expected: "banana",
		},
		{
			name:     "multiple words",
			input:    "ellohay orldway",
			expected: "hello world",
		},
		{
			name:     "sentence with punctuation",
			input:    "Ellohay, orldway!",
			expected: "Hello, world!",
		},
		{
			name:     "capitalized word",
			input:    "Ellohay",
			expected: "Hello",
		},
		{
			name:     "complex sentence",
			input:    "Ethay uickqay ownbray oxfay umpsjay overway ethay azylay ogday.",
			expected: "The quick brown fox jumps over the lazy dog.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromPigLatin(tt.input)
			if result != tt.expected {
				t.Errorf("FromPigLatin(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWordToPigLatin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "word starting with vowel",
			input:    "apple",
			expected: "appleway",
		},
		{
			name:     "word starting with consonant",
			input:    "banana",
			expected: "ananabay",
		},
		{
			name:     "word with punctuation",
			input:    "hello!",
			expected: "ellohay!",
		},
		{
			name:     "capitalized word",
			input:    "Hello",
			expected: "Ellohay",
		},
		{
			name:     "word with no vowels",
			input:    "rhythm",
			expected: "rhythmay",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wordToPigLatin(tt.input)
			if result != tt.expected {
				t.Errorf("wordToPigLatin(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWordFromPigLatin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "short word",
			input:    "hi",
			expected: "hi",
		},
		{
			name:     "word with way suffix",
			input:    "appleway",
			expected: "apple",
		},
		{
			name:     "word with ay suffix",
			input:    "ananabay",
			expected: "banana",
		},
		{
			name:     "word with punctuation",
			input:    "ellohay!",
			expected: "hello!",
		},
		{
			name:     "capitalized word",
			input:    "Ellohay",
			expected: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wordFromPigLatin(tt.input)
			if result != tt.expected {
				t.Errorf("wordFromPigLatin(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsVowel(t *testing.T) {
	vowels := []rune{'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U'}
	consonants := []rune{'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x', 'y', 'z',
		'B', 'C', 'D', 'F', 'G', 'H', 'J', 'K', 'L', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'V', 'W', 'X', 'Y', 'Z'}

	for _, vowel := range vowels {
		if !isVowel(vowel) {
			t.Errorf("isVowel(%q) = false, expected true", vowel)
		}
	}

	for _, consonant := range consonants {
		if isVowel(consonant) {
			t.Errorf("isVowel(%q) = true, expected false", consonant)
		}
	}

	// Test non-alphabetic characters
	nonAlpha := []rune{'1', '2', '3', '!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '-', '_', '+', '=', '{', '}', '[', ']', '|', '\\', ':', ';', '"', '\'', '<', '>', ',', '.', '?', '/'}
	for _, char := range nonAlpha {
		if isVowel(char) {
			t.Errorf("isVowel(%q) = true, expected false", char)
		}
	}
}

// TestRoundTrip tests that converting to Pig Latin and back results in the original text
func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "single word",
			input: "hello",
		},
		{
			name:  "multiple words",
			input: "hello world",
		},
		{
			name:  "sentence with punctuation",
			input: "Hello, world!",
		},
		{
			name:  "complex sentence",
			input: "The quick brown fox jumps over the lazy dog.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pigLatin := ToPigLatin(tt.input)
			result := FromPigLatin(pigLatin)
			if result != tt.input {
				t.Errorf("FromPigLatin(ToPigLatin(%q)) = %q, expected %q", tt.input, result, tt.input)
			}
		})
	}
}
