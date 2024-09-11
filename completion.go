package llamainterface

import (
	"encoding/json"
	"fmt"
	"image"
	"reflect"
	"strings"
	"time"
)

//Custom JSON. Produce flat JSON

type LogitBiasValue struct {
	Token       int64
	Propability float64
	Never       bool //set as false
}

type GenerationSettings struct { //JSON copupaste
	FrequencyPenalty float64 `json:"frequency_penalty"`
	//TODO `json:"grammar"`
	InoreEos        bool            `json:"ignore_eos"`
	LogitBias       json.RawMessage `json:"logit_bias"`
	Mirostat        int             `json:"mirostat"`
	MirostatEta     float64         `json:"mirostat_eta"`
	MirostatTau     float64         `json:"mirostat_tau"`
	Model           string          `json:"model"`
	NCtx            int             `json:"n_ctx"`
	NKeep           int             `json:"n_keep"`
	NPredict        int             `json:"n_predict"`
	NProbs          int             `json:"n_probs"`
	PenalizeNl      bool            `json:"penalize_nl"`
	PresencePenalty float64         `json:"presence_penalty"`
	RepeatLastN     int             `json:"repeat_last_n"`
	RepeatPenalty   float64         `json:"repeat_penalty"`
	Seed            int64           `json:"seed"`
	Stop            []string        `json:"stop"`
	Stream          bool            `json:"stream"`
	Temperature     float64         `json:"temp"`
	TfsZ            float64         `json:"tfs_z"`
	TopK            float64         `json:"top_k"`
	TopP            float64         `json:"top_p"`
	TypicalP        float64         `json:"typical_p"`
}

// OnlyPrompt is used for escaping json
type OnlyPrompt struct {
	Prompt string `json:"prompt"`
}

/*
system_prompt
 - prompt
 - anti_prompt
 - assistant_name
*/

type QueryCompletion struct {
	Prompt       string //: Provide the prompt for this completion as a string or as an array of strings or numbers representing tokens. Internally, the prompt is compared to the previous completion and only the "unseen" suffix is evaluated. If the prompt is a string or an array with the first element given as a string, a bos token is inserted in the front like main does.
	PromptTokens []int
	Temperature  float64 //: Adjust the randomness of the generated text (default: 0.8).

	Top_k     float64 //Limit the next token selection to the K most probable tokens (default: 40).
	Top_p     float64 //Limit the next token selection to a subset of tokens with a cumulative probability above a threshold P (default: 0.95).
	N_predict int     // Set the maximum number of tokens to predict when generating text. Note: May exceed the set limit slightly if the last token is a partial multibyte character. When 0, no tokens will be generated but the prompt is evaluated into the cache. (default: -1, -1 = infinity).
	N_keep    int     //Specify the number of tokens from the prompt to retain when the context size is exceeded and tokens need to be discarded. By default, this value is set to 0 (meaning no tokens are kept). Use -1 to retain all tokens from the prompt.
	Stream    bool    //: It allows receiving each predicted token in real-time instead of waiting for the completion to finish. To enable this, set to true.

	Stop []string //Specify a JSON array of stopping strings. These words will not be included in the completion, so make sure to add them to the prompt for the next iteration (default: []).

	Tfs_z          float64 //: Enable tail free sampling with parameter z (default: 1.0, 1.0 = disabled).
	Typical_p      float64 //: Enable locally typical sampling with parameter p (default: 1.0, 1.0 = disabled).
	Repeat_penalty float64 //: Control the repetition of token sequences in the generated text (default: 1.1).
	Repeat_last_n  int     //: Last n tokens to consider for penalizing repetition (default: 64, 0 = disabled, -1 = ctx-size).
	Penalize_nl    bool    //: Penalize newline tokens when applying the repeat penalty (default: true).

	Presence_penalty float64 //: Repeat alpha presence penalty (default: 0.0, 0.0 = disabled).

	Frequency_penalty float64 //Repeat alpha frequency penalty (default: 0.0, 0.0 = disabled);

	//TODO CONST
	Mirostat     int     //Enable Mirostat sampling, controlling perplexity during text generation (default: 0, 0 = disabled, 1 = Mirostat, 2 = Mirostat 2.0).
	Mirostat_tau float64 //Set the Mirostat target entropy, parameter tau (default: 5.0).
	Mirostat_eta float64 //Set the Mirostat learning rate, parameter eta (default: 0.1).

	//Grammar TODO `json:"grammar"` //Set grammar for grammar-based sampling (default: no grammar)

	Seed int64 //Set the random number generator (RNG) seed (default: -1, -1 = random seed).

	Ignore_eos bool //Ignore end of stream token and continue generating (default: false).

	Logit_bias []LogitBiasValue //Modify the likelihood of a token appearing in the generated text completion. For example, use "logit_bias": [[15043,1.0]] to increase the likelihood of the token 'Hello', or "logit_bias": [[15043,-1.0]] to decrease its likelihood. Setting the value to false, "logit_bias": [[15043,false]] ensures that the token Hello is never produced (default: []).

	N_probs int //If greater than 0, the response also contains the probabilities of top N tokens for each generated token (default: 0)

	Image_data map[int]image.Image // puts image in correct format BASE64 png?, and assigns ID
}

// Diff report, except prompt and image_data. Used for printing out what have changed
func (p *QueryCompletion) DiffReport(old QueryCompletion) string {
	var sb strings.Builder

	if p.Temperature != old.Temperature {
		sb.WriteString(fmt.Sprintf("Temperature=%v\n", p.Temperature))
	}

	if p.Top_k != old.Top_k {
		sb.WriteString(fmt.Sprintf("Top_k=%v\n", p.Top_k))
	}
	if p.Top_p != old.Top_p {
		sb.WriteString(fmt.Sprintf("Top_p=%v\n", p.Top_p))
	}
	if p.N_predict != old.N_predict {
		sb.WriteString(fmt.Sprintf("N_predict=%v\n", p.N_predict))
	}
	if p.N_keep != old.N_keep {
		sb.WriteString(fmt.Sprintf("N_keep=%v\n", p.N_keep))
	}
	if p.Stream != old.Stream {
		sb.WriteString(fmt.Sprintf("Stream=%v\n", p.Stream))
	}

	if !reflect.DeepEqual(p.Stop, old.Stop) {
		sb.WriteString(fmt.Sprintf("Stop=%v\n", p.Stop))
	}

	if p.Tfs_z != old.Tfs_z {
		sb.WriteString(fmt.Sprintf("Tfs_z=%v\n", p.Tfs_z))
	}
	if p.Typical_p != old.Typical_p {
		sb.WriteString(fmt.Sprintf("Typical_p=%v\n", p.Typical_p))
	}
	if p.Repeat_penalty != old.Repeat_penalty {
		sb.WriteString(fmt.Sprintf("Repeat_penalty=%v\n", p.Repeat_penalty))
	}
	if p.Repeat_last_n != old.Repeat_last_n {
		sb.WriteString(fmt.Sprintf("Repeat_last_n=%v\n", p.Repeat_last_n))
	}
	if p.Penalize_nl != old.Penalize_nl {
		sb.WriteString(fmt.Sprintf("Penalize_nl=%v\n", p.Penalize_nl))
	}

	if p.Presence_penalty != old.Presence_penalty {
		sb.WriteString(fmt.Sprintf("Presence_penalty=%v\n", p.Presence_penalty))
	}

	if p.Frequency_penalty != old.Frequency_penalty {
		sb.WriteString(fmt.Sprintf("Frequency_penalty=%v\n", p.Frequency_penalty))
	}

	if p.Mirostat != old.Mirostat {
		sb.WriteString(fmt.Sprintf("Mirostat=%v\n", p.Mirostat))
	}
	if p.Mirostat_tau != old.Mirostat_tau {
		sb.WriteString(fmt.Sprintf("Mirostat_tau=%v\n", p.Mirostat_tau))
	}
	if p.Mirostat_eta != old.Mirostat_eta {
		sb.WriteString(fmt.Sprintf("Mirostat_eta=%v\n", p.Mirostat_eta))
	}

	if p.Seed != old.Seed {
		sb.WriteString(fmt.Sprintf("Seed=%v\n", p.Seed))
	}

	if p.Ignore_eos != old.Ignore_eos {
		sb.WriteString(fmt.Sprintf("Ignore_eos=%v\n", p.Ignore_eos))
	}

	if !reflect.DeepEqual(p.Logit_bias, old.Logit_bias) {
		sb.WriteString(fmt.Sprintf("Logit_bias=%v\n", p.Logit_bias))
	}

	if p.N_probs != old.N_probs {
		sb.WriteString(fmt.Sprintf("N_probs=%v\n", p.N_probs))
	}
	return sb.String()
}

type CompletionTiming struct {
	PredictedN        int
	PredictedDuration time.Duration

	PromptN        int
	PromptDuration time.Duration
}

type CompletionTimingJson struct { //Stupid way to give so many numbers... just pre-calced
	PredictedMs          float64 `json:"predicted_ms"`
	PredictedN           int     `json:"predicted_n"`
	Predicted_per_second float64 `json:"predicted_per_second"`
	PredictedPerToken    float64 `json:"predicted_per_token_ms"`
	PromptMs             float64 `json:"prompt_ms"`
	PromptN              int     `json:"prompt_n"`
	PromptPerSecond      float64 `json:"prompt_per_second"`
	PromptPerTokenMs     float64 `json:"prompt_per_token_ms"`
}

// TODO allow non-complete when streaming!
// When using streaming mode (`stream`) only `content` and `stop` will be returned until end of completion.
type CompletionResult struct {
	Content         string               `json:"content"`             // Completion result as a string (excluding `stopping_word` if any). In case of streaming mode, will contain the next token as a string.
	Stop            bool                 `json:"stop"`                // Boolean for use with `stream` to check whether the generation has stopped (Note: This is not related to stopping words array `stop` from input options)
	GenSettings     GenerationSettings   `json:"generation_settings"` //`generation_settings`: The provided options above excluding `prompt` but including `n_ctx`, `model`
	Model           string               `json:"model"`               // The path to the model loaded with `-m`
	Prompt          string               `json:"prompt"`              // `prompt`: The provided `prompt`
	StoppedEos      bool                 `json:"stopped_eos"`         // Indicating whether the completion has stopped because it encountered the EOS token
	StoppedLimit    bool                 `json:"stopped_limit"`       // Indicating whether the completion stopped because `n_predict` tokens were generated before stop words or EOS was encountered
	StoppedWord     bool                 `json:"stopped_word"`        // Indicating whether the completion stopped due to encountering a stopping word from `stop` JSON array provided
	StoppingWord    string               `json:"stopping_word"`       // The stopping word encountered which stopped the generation (or "" if not stopped due to a stopping word)
	Timings         CompletionTimingJson `json:"timings"`             // Hash of timing information about the completion such as the number of tokens `predicted_per_second`
	TokensCached    int                  `json:"tokens_cached"`       // Number of tokens from the prompt which could be re-used from previous completion (`n_past`)
	TokensEvaluated int                  `json:"tokens_evaluated"`    // Number of tokens evaluated in total from the prompt
	Truncated       bool                 `json:"truncated"`           //Boolean indicating if the context size was exceeded during generation, i.e. the number of tokens provided in the prompt (`tokens_evaluated`) plus tokens generated (`tokens predicted`) exceeded the context size (`n_ctx`)
	SlotId          int                  `json:"slot_id"`             //Assign the completion task to an specific slot. If is -1 the task will be assigned to a Idle slot (default: -1)
	CachePrompt     bool                 `json:"cache_prompt"`        // Save the prompt and generation for avoid reprocess entire prompt if a part of this isn't change (default: false)
	SystemPrompt    string               `json:"system_prompt"`       // Change the system prompt (initial prompt of all slots), this is useful for chat applications. [See more](#change-system-prompt-on-runtime)
}

func (p *QueryCompletion) MarshalJSON() ([]byte, error) {
	//var sb strings.Builder
	parts := []string{}
	//sb.WriteString("{\n")
	d := DefaultQueryCompletion()

	if len(p.PromptTokens) == 0 {

		byt, errMarsh := json.Marshal(OnlyPrompt{Prompt: p.Prompt})
		if errMarsh != nil {
			return nil, fmt.Errorf("error marshaling prompt text %s", errMarsh)
		}
		s := string(byt)

		parts = append(parts, s[1:len(s)-1])
		//parts = append(parts, fmt.Sprintf("\"prompt\":\"%s\"", p.Prompt))
	} else {
		s := fmt.Sprintf("%#v", p.PromptTokens)
		s = strings.Replace(s, "[]int{", "", 1)
		s = strings.Replace(s, "}", "", 1)

		parts = append(parts, fmt.Sprintf("\"prompt\":[%s]", s)) //TODO WHY NOT WORKING?

	}

	if d.Temperature != p.Temperature {
		parts = append(parts, fmt.Sprintf("\"temperature\":%.3f", p.Temperature))
	}
	if d.Top_k != p.Top_k {
		parts = append(parts, fmt.Sprintf("\"top_k\":%.3f", p.Top_k))
	}
	if d.Top_p != p.Top_p {
		parts = append(parts, fmt.Sprintf("\"top_p\":%.3f", p.Top_p))
	}

	if d.N_predict != p.N_predict {
		parts = append(parts, fmt.Sprintf("\"n_predict\":%v", p.N_predict))
	}

	if d.N_keep != p.N_keep {
		parts = append(parts, fmt.Sprintf("\"n_keep\":%v", p.N_keep))
	}

	if p.Stream {
		parts = append(parts, "\"stream\":true")
	}

	if 0 < len(p.Stop) {
		parts = append(parts, "\"stop\":[\""+strings.Join(p.Stop, "\",\"")+"\"]")
	}

	if d.Tfs_z != p.Tfs_z {
		parts = append(parts, fmt.Sprintf("\"tfs_z\":%.3f", p.Tfs_z))
	}
	if d.Typical_p != p.Typical_p {
		parts = append(parts, fmt.Sprintf("\"typical_p\":%.3f", p.Typical_p))
	}

	if d.Repeat_penalty != p.Repeat_penalty {
		parts = append(parts, fmt.Sprintf("\"repeat_penalty\":%.3f", p.Repeat_penalty))
	}

	if d.Repeat_last_n != p.Repeat_last_n {
		parts = append(parts, fmt.Sprintf("\"repeat_last_n\":%v", p.Repeat_last_n))
	}

	if d.Penalize_nl != p.Penalize_nl {
		parts = append(parts, fmt.Sprintf("\"penalize_nl\":%v", p.Penalize_nl))
	}

	if d.Presence_penalty != p.Presence_penalty {
		parts = append(parts, fmt.Sprintf("\"presence_penalty\":%.3f", p.Presence_penalty))
	}

	if d.Frequency_penalty != p.Frequency_penalty {
		parts = append(parts, fmt.Sprintf("\"frequency_penalty\":%.3f", p.Frequency_penalty))
	}

	//TODO CONST
	if d.Mirostat != p.Mirostat {
		parts = append(parts, fmt.Sprintf("\"mirostat\":%v", p.Mirostat))
	}

	if d.Mirostat_tau != p.Mirostat_tau {
		parts = append(parts, fmt.Sprintf("\"mirostat_tau\":%.3f", p.Mirostat_tau))
	}
	if d.Mirostat_eta != p.Mirostat_eta {
		parts = append(parts, fmt.Sprintf("\"mirostat_eta\":%.3f", p.Mirostat_eta))
	}

	//Grammar TODO `json:"grammar"` //Set grammar for grammar-based sampling (default: no grammar)

	if d.Seed != p.Seed {
		parts = append(parts, fmt.Sprintf("\"seed\":%v", p.Seed))
	}

	if d.Ignore_eos != p.Ignore_eos {
		parts = append(parts, fmt.Sprintf("\"ignore_eos\":%v", p.Ignore_eos))
	}

	if 0 < len(p.Logit_bias) {
		//M example, use "logit_bias": [[15043,1.0]] to increase the likelihood of the token 'Hello', or "logit_bias": [[15043,-1.0]] to decrease its likelihood. Setting the value to false, "logit_bias": [[15043,false]] ensures that the token Hello is never produced (default: []).
		var sb strings.Builder
		sb.WriteString("\"logit_bias\":[")
		for i, a := range p.Logit_bias {
			if i != 0 {
				sb.WriteString(",")
			}
			if a.Never {
				sb.WriteString(fmt.Sprintf("[%v,false]", a.Token))
			} else {
				sb.WriteString(fmt.Sprintf("[%v,%.3f]", a.Token, a.Propability))
			}
		}
		sb.WriteString("]")
		parts = append(parts, sb.String())
	}

	if d.N_probs != p.N_probs {
		parts = append(parts, fmt.Sprintf("\"\":%v", p.N_probs))
	}

	/*
		for imageId, img := range p.Image_data {
			//TODO CONVERT IMAGE FOR MULTIMODAL!
		}*/

	return []byte("{" + strings.Join(parts, ",\n") + "}"), nil
}

func DefaultQueryCompletion() QueryCompletion {
	return QueryCompletion{
		Prompt:      "",
		Temperature: 0.8,

		Top_k:     40.0,
		Top_p:     0.95,
		N_predict: -1, //(default: -1, -1 = infinity).
		N_keep:    0,
		Stream:    false,
		Stop:      []string{},

		Tfs_z:          1.0,
		Typical_p:      1.0,
		Repeat_penalty: 1.1,
		Repeat_last_n:  64,
		Penalize_nl:    true,

		Presence_penalty: 0,

		Frequency_penalty: 0.0,

		//TODO CONST
		Mirostat: 0,

		Mirostat_tau: 5.0,
		Mirostat_eta: 0.1,

		//Grammar TODO `json:"grammar"` //Set grammar for grammar-based sampling (default: no grammar)

		Seed: -1,

		Ignore_eos: false,

		Logit_bias: []LogitBiasValue{},
		N_probs:    0,
		Image_data: nil,
	}
}

func (p *LLamaServer) PostCompletion(q QueryCompletion, updates chan CompletionResult, timeout time.Duration) (CompletionResult, error) {
	queryRaw, errQueryRaw := q.MarshalJSON()
	if errQueryRaw != nil {
		return CompletionResult{}, errQueryRaw
	}

	rowFeed := make(chan string, 100)

	fmt.Printf("query is %s\n\n", queryRaw)

	final, errPost := p.PostQuery("/completion", queryRaw, rowFeed, timeout)
	if errPost != nil {
		if updates != nil {
			close(updates)
		}
		return CompletionResult{}, errPost
	}
	completed := CompletionResult{}

	//fmt.Printf("\n\nFINAL IS %s\n\n", final)

	go func() {
		d := CompletionResult{}
		for row := range rowFeed {
			parseErr := json.Unmarshal([]byte(row), &d)
			if parseErr != nil {
				close(updates)
				break
			}
			updates <- d
		}
	}()

	parseErr := json.Unmarshal([]byte(final), &completed)
	if parseErr != nil {
		if updates != nil {
			close(updates)
		}
		return CompletionResult{}, parseErr
	}
	return completed, nil
}

/*
// PostCompletionRetryOnEmpty  until there is timeout
func (p *LLamaServer) PostCompletionRetryOnEmpty(q QueryCompletion, updates chan CompletionResult, timeout time.Duration) (CompletionResult, error) {
	tStart := time.Now()
	tEnd := tStart.Add(timeout)

	queryRaw, errQueryRaw := q.MarshalJSON()
	if errQueryRaw != nil {
		return CompletionResult{}, errQueryRaw
	}

	rowFeed := make(chan string, 100)

	final, errPost := p.PostQuery("/completion", queryRaw, rowFeed, tEnd.Sub(time.Now()))
	if errPost != nil {
		if updates != nil {
			close(updates)
		}
		return CompletionResult{}, errPost
	}
	completed := CompletionResult{}

	fmt.Printf("\n\nFINAL IS %s\n\n", final)

	go func() {
		d := CompletionResult{}
		for row := range rowFeed {
			strings.TrimSpace()
			parseErr := json.Unmarshal([]byte(row), &d)
			if parseErr != nil {
				close(updates)
				break
			}
			updates <- d
		}
	}()

	parseErr := json.Unmarshal([]byte(final), &completed)
	if parseErr != nil {
		if updates != nil {
			close(updates)
		}
		return CompletionResult{}, parseErr
	}
	return completed, nil
}
*/
