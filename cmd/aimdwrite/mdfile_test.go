package main

import (
	"testing"
)

func TestToFileContent(t *testing.T) {
	testMdFile := MdFile{
		{Level: 1, Title: "Chapter 1", Rows: []string{"Content 1", "Content 2"}},
		{Level: 2, Title: "Subchapter 1.1", Rows: []string{"Subcontent 1", "Subcontent 2"}},
	}

	expectedContent := "# Chapter 1\nContent 1\nContent 2\n## Subchapter 1.1\nSubcontent 1\nSubcontent 2\n"
	result, err := testMdFile.ToFileContent()
	if err != nil {
		t.Errorf("Error converting MdFile to file content: %v", err)
	}

	if result != expectedContent {
		t.Errorf("Unexpected file content. Got:\n%#v\n\nExpected:\n%#v", result, expectedContent)
	}
}

func TestCalcTitleLevel(t *testing.T) {
	testCases := []struct {
		Input    string
		Expected int
	}{
		{"# Title", 1},
		{"## Subtitle", 2},
		{"### Sub-subtitle", 3},
		{"No Title", 0},
		{"#Invalid Space", 0},
	}

	for _, tc := range testCases {
		result := CalcTitleLevel(tc.Input)
		if result != tc.Expected {
			t.Errorf("Unexpected title level for input '%s'. Got %d, want %d", tc.Input, result, tc.Expected)
		}
	}
}

func TestRemoveCommentsFromRows(t *testing.T) {
	// Assuming you have a test case with rows containing comments
	testRows := []string{
		"Some content",
		"<!-- Comment 1 -->",
		"More content <!-- Comment 2 -->",
		"stuff goes here just ok",
		"then start comment <!---- and here it is",
		"continues to next line",
		"and then ends --->DONE",
		"Final line"}

	expectedResult := []string{"Some content", "", "More content ", "stuff goes here just ok", "then start comment DONE", "Final line"}

	result := removeCommentsFromRows(testRows)

	if !stringSlicesEqual(result, expectedResult) {
		t.Errorf("Unexpected result after removing comments. Got:\n%#v\n\nExpected:\n%#v", result, expectedResult)
	}
}

// Helper function to check if two string slices are equal
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
