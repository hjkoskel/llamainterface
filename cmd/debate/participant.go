package main

import (
	"fmt"
	"strings"
)

/*
 builds up its own prompt, depending on what other person says
 Does not delete old.

 Builds whole input for LLM with required number of words etc..
From point of view of Me or Another

*/

/*
type PromptFragmentSetting struct {
	Prefix string // Like "User:"  or <|im_start|>user
	Suffix string // Empty or <|im_end|>
}*/

type ParticipantConf struct {
	InitialPrompt string //Like system prompt, repeats every time
	Name          string //For user printout
}

type PrefixSuffix struct {
	Prefix string
	Suffix string
}

type Arena struct {
	InitialPart PrefixSuffix //System prompt
	AgentPart   PrefixSuffix
	UserPart    PrefixSuffix

	MyConf      ParticipantConf
	AnotherConf ParticipantConf

	Turns                 []string // Me is the first
	StartContextTurnIndex int      //Increase this, limit tokens

}

// Repeating  conversation jams into "I apologize" or "Good day" loop
func (p *Arena) Repeating() bool {
	if len(p.Turns) < 4 {
		return false
	}

	last := len(p.Turns) - 1
	return p.Turns[last] == p.Turns[last-2] && p.Turns[last-1] == p.Turns[last-3]
}

func (p *Arena) MyTurnNext() bool {
	return (len(p.Turns) % 2) == 0 //capsulate rule in one place
}

func (p *Arena) Append(s string, isay bool) error { //Just keep pace and get USER/assistant correct
	txt := strings.TrimSpace(s)
	if len(txt) == 0 {
		return fmt.Errorf("string is empty")
	}

	//Check correct turn
	if p.MyTurnNext() != isay {
		return fmt.Errorf("turn mismatch isay=%v got lines=%v", isay, len(p.Turns))
	}

	p.Turns = append(p.Turns, txt)
	return nil
}

func (p *Arena) Prompt() string {
	var sb strings.Builder
	if p.MyTurnNext() {
		sb.WriteString(p.InitialPart.Prefix + p.MyConf.InitialPrompt + p.InitialPart.Suffix)
	} else {
		sb.WriteString(p.InitialPart.Prefix + p.AnotherConf.InitialPrompt + p.InitialPart.Suffix)
	}
	side := !p.MyTurnNext() //this is the first

	for index, s := range p.Turns {
		if p.StartContextTurnIndex <= index {
			if !side {
				sb.WriteString(p.AgentPart.Prefix + s + p.AgentPart.Suffix)
			} else {
				sb.WriteString(p.UserPart.Prefix + s + p.UserPart.Suffix)
			}
		}
		side = !side //Just toggle in turns
	}
	sb.WriteString(p.AgentPart.Prefix) //In reality, user is the lastTurn

	return sb.String() //clean?
}

func (p *Arena) userPrintRow(i int) string {
	if len(p.Turns) <= i {
		return "" //or error?

	}
	if i%2 == 0 {
		return fmt.Sprintf("%s: %s", p.MyConf.Name, p.Turns[i])
	}
	return fmt.Sprintf("%s: %s", p.AnotherConf.Name, p.Turns[i])

}

func (p *Arena) LatestUserPrintout() string {
	return p.userPrintRow(len(p.Turns) - 1)
}

func (p *Arena) FullUserPrintout() string { //Color support?
	var sb strings.Builder
	for i, _ := range p.Turns {
		sb.WriteString(p.userPrintRow(i))
		sb.WriteString("\n\n")
	}
	return sb.String()
}
