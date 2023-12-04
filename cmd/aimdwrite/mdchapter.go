package main

import "strings"

type MdChapter struct {
	Level int    //0 not title before  1=# 2=## etc...
	Title string //Important to remove titles from context so LLM does not go crazy
	Rows  []string
}

func lineIsSummaryStart(line string) bool {
	s := strings.ReplaceAll(strings.TrimSpace(line), " ", "")
	return s == "~~~summary" || s == "```summary"
}

func lineIsSummaryEnd(line string) bool {
	s := strings.TrimSpace(line)
	return s == "~~~" || s == "```" //TODO allows now mixing~~~and ```   :(
}

// Summary and what is left.
func (p *MdChapter) ExtractSummary() ([]string, []string) {
	rows := removeCommentsFromRows(p.Rows)

	summaryRows := []string{}
	contentRows := []string{}
	nowSummary := false
	for _, row := range rows {
		if nowSummary {
			if lineIsSummaryEnd(row) {
				nowSummary = false
			} else {
				summaryRows = append(summaryRows, row)
			}
		} else {
			if lineIsSummaryStart(row) {
				nowSummary = true
			} else {
				contentRows = append(contentRows, row)
			}
		}
	}
	return summaryRows, contentRows
}

func (p *MdChapter) CmdRow() int {
	for i, row := range p.Rows {
		if strings.TrimSpace(row) == LINESTART_CMD {
			return i
		}
	}
	return -1
}
