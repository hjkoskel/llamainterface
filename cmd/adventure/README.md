# Adventure

This program demonstrates how to create psychedelic adventure experience with generative AI

Project is still under active development and it requires major refactoring. Whole game design started as proof of concept and example how to use llama library

[Please check example game run](https://smallpdf.com/file#s=544c875d-a258-4987-a323-d34f4dc27f28)

## Installation

Actual building of software is easy, just use go build
~~~ sh
go build
~~~

Software requires llama.cpp server running and stable diffusion or flux.1 server.
Game software will do http post queries to those servers. On simple typical game setup those servers run on same computer.

Reason why large language model and graphic generation are separated from game, is that it is more flexible in this way.
There are user scenarios where language and graphic generation models are runned on separate computers for memory/gpu contraints.
Also it allows user to choose what model is runned on GPU and what on CPU.



### Get image generator server running

There are two ways to run stable diffusion or flux.1 server. Both have compatible interfaces

The first option is to use python script.
~~~ sh
python fluxserver.py
~~~
Script have command line options **-p** for choosing TCP listen port (default: 8800). And model **-m** (black-forest-labs/FLUX.1-schnell)
Installation requires installing [*diffusers* huggingface](https://github.com/huggingface/diffusers) FluxPipeline.
Installing diffusers library might have some challenges and using library purely on CPU is not possible due bug

Usually it is bad idea to run python on production. For that reason there is C++ implementation available.
[https://github.com/leejet/stable-diffusion.cpp](https://github.com/leejet/stable-diffusion.cpp)
Main branch does not yet have ./example/sd-server available. but there is [pull request](https://github.com/leejet/stable-diffusion.cpp/pull/367).
But there is fork available with *server* branch. Where there is sd-server implemented

[https://github.com/stduhpf/stable-diffusion.cpp/tree/server](https://github.com/stduhpf/stable-diffusion.cpp/tree/server)
Follow README instructions and compile examples way what is optimal for your system (nvidia vs amd vs cpu).

Then download the best model suited for games. My favourite is flux.1 (snell or dev). Follow instructions on
[https://github.com/stduhpf/stable-diffusion.cpp/blob/server/docs/flux.md](https://github.com/stduhpf/stable-diffusion.cpp/blob/server/docs/flux.md)

sd-server start command should look something like this
~~~ sh
./sd-server --port 8800 --prompt "test image" --diffusion-model ./models/flux1-schnell-q3_k.gguf --vae ./models/ae.sft --clip_l ./models/clip_l.safetensors --t5xxl ./models/t5xxl_fp16.safetensors --clip-on-cpu --vae-on-cpu --cfg-scale 1.0 --sampling-method euler -v --steps 20
~~~

Disk storage requirements for minimal setup would be something like this
|fname | size approx |
|------|--------------|
|flux1-schnell-q3_k.gguf |  5G|
|ae.sft | 320M |
|clip_l.safetensors | 235|
|t5xxl_fp16.safetensors | 9.2G|


Actual game have following settings that control directly image quality

~~~
  -igh int
        image generation Y-resolution (default 1024)
  -igsteps int
        image generation steps, affects speed vs quality (default 20)
  -igurl string
        image generator URL (default "http://127.0.0.1:8800/txt2img")
  -igw int
        image generation X-resolution (default 1024)
  -igsampler string
        choose sampling method [euler_a,euler,heun,dpm2,dpm++2s_a,dpm++2m,dpm++2mv2,lcm] (default "euler")
~~~
What I have noticed is that with 6GB VRAM, the maximum picture size is 512x512

### Choose way to run LLM

There are many ways to run LLM model for game. It was chosen to support external llama.cpp based program for following reasons
- User might want to use shared LLM server for multiple applications, running on dedicated computing server without exessive RAM usage
- The most performant way to compile LLM interference code depends on user. GPU, apple metal, CPU arch, NPU?
- llama.cpp development is very active.
- There are other ways to run like ollama or llamafile

Actual options are
- Call running *llama.cpp* server by setting
    - IP address *-h*
    - port *-p* 
- use llamafile
    - *-l* defines filename to be started and closed when software ends
        - It is also possible to llamafile as llama.cpp server and give *-p* instead of *-l*
    - *-p* option, is totally optional. Runs llamafile on specific port
    - use custom weights if llamafile do not include correct model *-lm*

So typical command line settings would be like
~~~
./adventure -l /usr/local/bin/llamafile -lm ~/Downloads/llama-3.2-1b-instruct-uncensored.Q8_0.gguf
~~~

Game support llama3 inst models and models following chatml format. It is very important to use instruct tuned models.
Prefered models are llama3.2 instruct models.  If using other kind of model, then *-pf* command line option is required

There are many options like size of models. Larger models are 
One good model is *llama-3.2-1b-instruct-uncensored.Q8_0.gguf. It is preferable not to use uncensored models if minors are playing. Also it is worth of using larger models if there is computing resource or waiting time available.


**NOTICE! Just give command  ./adventure -? and it will tell all command line parameters**

Links
- [llamafile](https://github.com/Mozilla-Ocho/llamafile)
- [llama.cpp](https://github.com/ggerganov/llama.cpp) build llama-server


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

## Optional localization

Game support [https://github.com/hjkoskel/turntongue](turntongue) language language translation server.
Install and start it and it is possible

- Option *-traurl*, translation url. Default http://127.0.0.1:8000/translate
- Option *-lang-*, pick use this language in translation (fin_Latn etc... )
- Option *-tmi*, set this flag and game translates all missing texts



## TODO

- youtube video
- More games, update prompts
- ML language Translation
- support for scrolling pages back on local ui
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


