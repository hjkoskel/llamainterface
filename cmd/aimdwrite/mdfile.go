package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	LINESTART_CMD = "$" //Start run from here, use previous as context
)

type MdFile []MdChapter

func LoadMdFileFromFile(fname string) (MdFile, error) {
	rows, errLoad := LoadToRows(fname)
	if errLoad != nil {
		return nil, errLoad
	}
	return LoadMdFromRows(rows), nil
}

func (p *MdFile) ToFileContent() (string, error) {
	var sb strings.Builder
	for chapterIndex, chapter := range *p {
		if chapter.Level == 0 && 0 < chapterIndex {
			return "", fmt.Errorf("chapter %v can not have level %v", chapterIndex, chapter.Level)
		}
		for i := 0; i < chapter.Level; i++ {
			sb.WriteString("#")
		}
		if 0 < chapter.Level {
			if len(chapter.Title) == 0 {
				return "", fmt.Errorf("no title on chapter %v", chapterIndex)
			}
			sb.WriteString(" ")
			sb.WriteString(chapter.Title)
			sb.WriteString("\n")
		}
		sb.WriteString(strings.Join(chapter.Rows, "\n"))
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func LoadToRows(fname string) ([]string, error) {
	byt, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	//Windows losers!!!  \n is enough!
	return strings.Split(strings.ReplaceAll(string(byt), "\r", ""), "\n"), nil
}

func CalcTitleLevel(row string) int {
	result := 0
	s := strings.TrimSpace(row)
	for strings.HasPrefix(s, "#") {
		result++
		s = s[1:]
	}
	if result == 0 {
		return 0
	}
	//Must have space after ## marks
	if s[0] != ' ' {
		return 0
	}
	return result
}

// Keep comments!!! It must be possible to re-build document from this
func LoadMdFromRows(rows []string) MdFile { //Non-hierarchial
	result := []MdChapter{}
	latest := MdChapter{Level: 0, Title: "", Rows: []string{}}

	for _, row := range rows {
		lvl := CalcTitleLevel(row)
		if lvl == 0 { //Continue on same level
			latest.Rows = append(latest.Rows, row)
		} else {
			result = append(result, latest)
			latest = MdChapter{Level: lvl, Title: strings.TrimSpace(strings.ReplaceAll(row, "#", "")), Rows: []string{}}
		}
	}
	result = append(result, latest)
	return result
}

// Get file to cmd, remove siblings etc..
func (p *MdFile) GetCmdChapterRowIndex() (int, int) {
	for i, chapter := range *p {
		r := chapter.CmdRow()
		if 0 < r {
			return i, r
		}
	}
	return -1, -1
}

// First is summary, then content without summaries
func (p *MdFile) GetSummaryFromN(n int) MdFile {
	summaryCounter := 0
	result := []MdChapter{}
	for _, cha := range *p {
		summary, content := cha.ExtractSummary()
		summaryCounter++
		if summaryCounter < n {
			continue
		}
		if summaryCounter == n {
			result = append(result, MdChapter{
				Level: cha.Level,
				Title: cha.Title,
				Rows:  summary,
			})
		} else {
			result = append(result, MdChapter{
				Level: cha.Level,
				Title: cha.Title,
				Rows:  content,
			})
		}
	}
	return result
}

// useSummaryN, 0= do not use summary. 1=Start file from first summary, 2=start file from second summary
func (p *MdFile) GetFileToCmd() MdFile {
	maxChapter, cmdRow := p.GetCmdChapterRowIndex()
	if maxChapter < 0 {
		return nil
	}

	a := (*p)[maxChapter]
	a.Rows = a.Rows[0:cmdRow]
	result := []MdChapter{a}

	for i := maxChapter - 1; 0 <= i; i-- {
		if (*p)[i].Level < result[0].Level {
			result = append([]MdChapter{(*p)[i]}, result...)
		}
	}
	return result
}

// Inserts new stuff
func (p *MdFile) SetToCmd(newStuff string) error {
	chapterIndex, rowIndex := p.GetCmdChapterRowIndex()
	if chapterIndex < 0 {
		return fmt.Errorf("no commands")
	}
	(*p)[chapterIndex].Rows[rowIndex] = strings.Replace((*p)[chapterIndex].Rows[rowIndex], "$", newStuff, 1)
	return nil
}

func (p *MdFile) JoinRows() []string {
	result := []string{}
	for _, chapter := range *p {
		for _, row := range chapter.Rows {
			result = append(result, row)
		}
	}
	return result
}

// Returns what is was before comment start and was it started
func beforeComment(s string) (string, bool) {
	i := strings.Index(s, "<!--")
	if i < 0 {
		return s, false
	}
	return s[0:i], true
}

func afterComment(s string) (string, bool) {
	i := strings.Index(s, "-->")
	if i < 0 {
		return s, false
	}
	return s[i+3:], true
}

func RemoveComments(input string) string {
	var result strings.Builder
	inComment := false
	hyphenCount := 0

	for i := 0; i < len(input); i++ {
		if i+3 < len(input) && input[i:i+4] == "<!--" {
			inComment = true
			hyphenCount = 0
			i += 3
		} else if inComment && input[i] == '-' {
			hyphenCount++
		} else if inComment && input[i] == '>' && hyphenCount > 0 {
			inComment = false
			hyphenCount = 0
		} else if !inComment {
			result.WriteByte(input[i])
		}
	}

	return result.String()
}

func removeCommentsFromRows(rows []string) []string {
	return strings.Split(RemoveComments(strings.Join(rows, "\n")), "\n")
}

// ToCleanText, cleans comments out and all other markdown stuff like pictures. Anything that disturbs LLM etc...
func (p *MdFile) ToCleanText() string {
	allRows := p.JoinRows()
	allRows = removeCommentsFromRows(allRows)
	for i, row := range allRows {
		if strings.HasPrefix(strings.TrimSpace(row), ">") {
			allRows[i] = strings.Replace(row, ">", "", 1)
		}
	}

	return strings.Join(allRows, "\n")
}
