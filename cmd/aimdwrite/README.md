# AI MD Write

This software is for using LLMs in way of text editor.

**This software is at early stage of development** and just acts as example for llamainterface

Idea is to start creating markdown document on your text editor of choice and then start **aimdwrite** helper program that starts generating new content to your document when it finds line starting with dollar sign $

LLMS have limitations on context length. Context length is for example typically 2048 tokens for LLMA2. Actual input text and output text must fit in that context length. AiMdWrite refuses working if text takes too much context.

AiMdWrite have few ways to handle this issue
- Hierarcial document. Use markdown chapter titles #,##,###
    - context is built from joining actual chapter where $ is and parent chapters "up" from there
    - This requires specific style of writing
- Summary code
    - If chapter have *summary* named code block, software will use that instead of normal content of chapter
    - like start code block with **~~~summary** instead like **~~~python**
- Comment out text
    - aimdwrite skips commented out text and directives 
    - normal markdown comment out syntax
    - can be used also when doing re-runs. Hide old response into comments and try again by adding $

Important to note

Some LLMS use # when operating in chat-dialog mode like #user: Those parts must be inside code or quotation ">" block
because hastag at start of line conflicts with markdown syntax


# Usage

aimdwrite is built with normal "go build" command.

For using software running [llama.cpp server](https://github.com/ggerganov/llama.cpp/tree/master/examples/server) is required
Server is found on 

Typical use case is to run llama.cpp server on high end computing server and use *aimdwrite* on much lightweight client.
Model like *Wizard-Vicuna-13B-Uncensored.Q5_K_M.gguf* is good for general use.

User also needs editor that can refresh view if open file is changed by some other program. Visual studio code can do that


## Command line options
Using aimdwrite is straightforward. If llama.cpp server is running on localhost, only -f is required.

Number of max tokens -mt should kept low enough so response is

~~~
Usage of ./aimdwrite:
  -f string
        file for reading and writing (default "example.md")
  -h string
        llama.cpp server host (default "127.0.0.1")
  -mt int
        max tokens in input prompt (default 1024)
  -p int
        llama.cpp server port (default 8080)
~~~

When*aimdwrite* is kept running it will notice when file is saved and it will check is there $ somewhere at start of line.

## Directives

The Dollar sign $ is the most used directive but there are few other magic words that allow change sampling settings

Following directives with their default values and explanation

* **$TEMP=0.4** sets temperature of model from that point forward
* **$N=-1** sets number of  N_predict, -1 no limit
* **$TOPK=40**	Limit the next token selection to the K most probable tokens (default: 40).
* **$TOPP=0.95**	Limit the next token selection to a subset of tokens with a cumulative probability above a threshold P (default: 0.95).
* **$NKEEP=0** Specify the number of tokens from the prompt to retain when the context size is exceeded and tokens need to be discarded. By default, this value is set to 0 (meaning no tokens are kept). Use -1 to retain all tokens from the prompt.
* **$TFS_Z=1.0** ,	Tfs_z          float64 //: Enable tail free sampling with parameter z (default: 1.0, 1.0 = disabled).
* **$TYPICALP=1.0** Typical_p      float64 //: Enable locally typical sampling with parameter p (default: 1.0, 1.0 = disabled).
* **$REPEATPENALTY=1.1** //Repeat_penalty float64 //: Control the repetition of token sequences in the generated text (default: 1.1).
* **$REPEATLASTN=64**	 Last n tokens to consider for penalizing repetition (default: 64, 0 = disabled, -1 = ctx-size).
* **$PENALIZENL=true** Penalize_nl    bool    //: Penalize newline tokens when applying the repeat penalty (default: true).
* **$PRESENCEPENALTY=0.0** // Presence_penalty float64 //: Repeat alpha presence penalty (default: 0.0, 0.0 = disabled).
* **$FREQUENCYPEBALTY=0** ,Repeat alpha frequency penalty (default: 0.0, 0.0 = disabled);
* **$MIROSTAT=0** Enable Mirostat sampling, controlling perplexity during text generation (default: 0, 0 = disabled, 1 = Mirostat, 2 = Mirostat 2.0).
* **$MIROSTATTAU=5**
* **$MIROSTATETA=0.1** Set the Mirostat learning rate, parameter eta (default: 0.1).
* **$SEED=-1**, Set the random number generator (RNG) seed (default: -1, -1 = random seed).


# Example run

Check separate [example documentation](examplerun.md) 
