# Debate

Example program how to run llama model.

Program will start debating on topic. LLM acts as both user and assistant role.

## Preparation

You must have access to running llama.cpp server instance

[llama.cpp  examples/server](https://github.com/ggerganov/llama.cpp/tree/master/examples/server)

By default this program assumes that llama.cpp server is running at localhost on port 8080.
If it is not so, use *-h* and *-p* options to change that

## Input files

Simulation requires following files

### Initial chat (-f option)
Chat containing start of chatting, A and B agents taking turns. First A and then B agent.

File given by *-f* option
~~~
what linux distro you are using now?
arch linux
that sucks windows is better
~~~

### Initial prompt of agents A and B (-fia and -fib options)

This prompt is fed at start of each run.

In this example user A is windows user (filename given by *-fia*)
~~~
you are windows user. You response as human not ai or language model. 
You are not really intrested technical details and are more people person. 
You continue dialog without ending
~~~

And then user B is linux user (filename given by *-fib*)
~~~
you are angry professional linux user who have strong arbitary opinions and preferences.
You response as human not ai or language model., you are human. You hate so much other operational systems. 
You love small technical details and arch linux You give strong arguments and have non-polite fuck you responses. 
Your answers are unfriendly and full of hate always. You continue dialog without ending
~~~


## Command line arguments

~~~
Usage of ./debate:
  -app string
        agent part prefix (default "<|im_start|>assistant ")
  -aps string
        agent part suffix (default "<|im_end|>")
  -c int
        context token count limit when posting. Must be less than context size (default 400)
  -f string
        initial chat. One turn per row. first row is first person, line by line taking turns (default "./chats/linuxuse.txt")
  -fia string
        text file for initial prompt for first person (default "./persons/winloser.txt")
  -fib string
        text file for initial prompt for second person (default "./persons/archuser.txt")
  -h string
        llama.cpp server host (default "127.0.0.1")
  -ipp string
        initial part prefix (system prompt) (default "<|im_start|>system ")
  -ips string
        initial part suffix (default "<|im_end|>")
  -lp string
        last prompt file (default "lastprompt.txt")
  -na string
        name of first person, leave empty and name comes from prompt filename
  -nb string
        name of second person, leave empty and name comes from prompt filename
  -of string
        output filename (default "text.log")
  -p int
        llama.cpp server port (default 8080)
  -tempA float
        temperature of first person (default 0.7)
  -tempB float
        temperature of first person (default 0.7)
  -upp string
         user part prefix (default "<|im_start|>user ")
  -ups string
        user part suffix (default "<|im_end|>")
~~~