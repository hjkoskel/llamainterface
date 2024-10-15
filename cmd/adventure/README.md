# Adventure

This program demonstrates how to create psychedelic adventure experience with generative AI

Project is still under active development and it requires major refactoring. Whole game design started as proof of concept and example how to use llama library

[Please check example game run](https://smallpdf.com/file#s=544c875d-a258-4987-a323-d34f4dc27f28)

## Installation

Actual building of software is easy, just use go build
~~~ sh
go build
~~~




### Get image generator running

~~~ sh
python fluxserver.py
~~~

Try starting that and install missing libraries
TODO LIST OR MAKE 
Script downloads required weights from huggingface

This *fluxserver.py* runs server on port 8800, server will react any http POST and use body of query as prompt. Return response is 1024x1024  *image/png*
Remember that it is possible to run flux.1 server on different compyter than where game runs

This is total hack

Deploying python is pain. Weights should be shared via IPFS or similar source, not one centralized place

C/C++ based solution without python is under investigation.

OPTION!!
Is it also possible use internal stable diffusion feature *-dmf* option. It is very slow and runs on CPU.
It does not yet support flux.1

Anyway flux.1 (snell or dev) is the best model out there imho

### Choose way to run LLM

There are many ways to run LLM model for game. It was chosen to support external llama.cpp based program for following reasons
- User might want to use shared LLM server for multiple applications, running on dedicated computing server without exessive RAM usage
- The most performant way to compile LLM interference code depends on user. GPU, apple metal, CPU arch, NPU?
- llama.cpp development is very active.
- There are other ways to run like ollama or llamafile

Actual options are
- Call *llama.cpp* server by setting
    - IP address *-h*
    - port *-p* 
- use llamafile
    - *-l* defines filename to be started and closed when software ends
        - It is also possible to llamafile as llama.cpp server and give *-p* instead of *-l*
    - *-p* option, is totally optional. Runs llamafile on specific port
    - use custom weights if llamafile do not include correct model *-lm*

**NOTICE! Just give command  ./adventure -? and it will tell all command line parameters**


### load models

Game works best on  llama-3.1 or llama-3.2 instruct models. Model have to be **Instruct** tuned
Even small model like 1B works fine. For example *Llama-3.2-1B-Instruct-Q5_K_M.gguf*

### Choose in between native vs web interfaces

Game support two user interfaces. Native raylib based full screen custom UI and server side rendered web browser UI.

At this point is not clear what kind of user interface is the best. Native provides the best *real computer game* like experience.
Web user interface is great for gaming over home wifi on mobile phone. When one turn of game can take few minutes, gaming can happen everywhere inside house.

For starting native interface
- use *-uip* value 0 (if it is not default). If want to use web interface give some other number than 0, that is HTML server port number
- Optionally load specific game with option *-load*



Common options
- Option *-tm* runs actual game in text mode (skips graphic generation). Generative AI is needed only menu pictures
- Option *-gmi* generate missing pictures from chosen adventure. Typical use is to play game with *-tm* and later generate pictures for story book
- Option *-tip* changes temperature in graphics generation
- Option *-tp* changes temperature for text generation, more temperature more randomness to generation
- Option *-rpp* clear picture prompts from loaded game. Allows to experiment prompt settings of game

## Running

Software have menu system for starting new game or continuing old games.

Newgames are created based on what json files are found from **./games** folder and old games are loaded from **./savegames**

## TODO

- More games, update prompts
- Dockerize flux.1 python
- support for scrolling pages back on local ui
- Use flux.1 interference on C/C++ (problem: leejet/stable-diffusion.cpp is much slower than python even on GPU)
- Handle interence picture with different resolutions
    - Picture resolution is different than 1024, smaller for testing?
    - Display resolution is different
- Support other image generation api than very dummy fluxserver.py 
- Store seed number of image generation. Allows re-calculation of picture
- Add command line option to reset all picture prompts and re-generate those and pictures. Important when testing image generation prompting
- store languea model used on each page
- llm or image synthesis duration for benchmarking
- Speech and/or music generation
- Internal LLM model execution for easier deployment (CPU)
- Video mode resolution selection on command prompt
- Utilize summary when building prompt and context lenght is filling up
- Better local ui
    - Browse back/forward while calculating next
- Webui, with more javascript
    - Show status
    - Query whole game json and operate from there instead of "server side rendering"


