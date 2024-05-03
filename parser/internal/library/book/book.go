package book

import "regexp"

type TitleList struct {
	Genre  string
	Author string
	Title  string
}

// NewTitleList extracts the genre, author, and title from a string and returns a BookTitle object.
//
// The function takes a string as input and uses a regular expression pattern to match and extract the genre, author, and title from the string. If the pattern matches successfully and there are at least 4 matches, a new BookTitle object is created with the extracted values and returned. If the pattern does not match or there are less than 4 matches, a new BookTitle object is created with the input string as the title and returned.
//
// Parameters:
// - str: The input string from which to extract the genre, author, and title.
//
// Returns:
// - *TitleList: A pointer to a TitleList object containing the extracted genre, author, and title. If the pattern does not match or there are less than 4 matches, the title field of the BookTitle object will contain the input string.
func NewTitleList(str string) *TitleList {
	const pattern = `([^_]+)_([^—]+) — (.+)`
	matches := regexp.MustCompile(pattern).FindStringSubmatch(str)
	if len(matches) > 3 {
		return &TitleList{
			Genre:  matches[1],
			Author: matches[2],
			Title:  matches[3],
		}
	}
	return &TitleList{Title: str}
}
