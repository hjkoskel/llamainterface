package llamainterface

import "fmt"

// Command line options for starting llama server
type ServerCommandLineFlags struct {
	Threads                     int    //--threads N , -t N: Set the number of threads to use during generation.
	ThreadsBatch                int    //-tb N, --threads-batch N : Set the number of threads to use during batch and prompt processing. If not specified, the number of threads will be set to the number of threads used for generation.
	ModelFilename               string //-m FNAME ,  --model FNAME : Specify the path to the LLaMA model file (e.g.,  models/7B/ggml-model.gguf ).
	Alias                       string //-m ALIAS ,  --alias ALIAS : Set an alias for the model. The alias will be returned in API responses.
	CtxSize                     int    //-c N ,  --ctx-size N : Set the size of the prompt context. The default is 512, but LLaMA models were built with a context of 2048, which will provide better results for longer input/inference. The size may differ in other models, for example, baichuan models were build with a context of 4096.
	NGPULayers                  int    //-ngl N ,  --n-gpu-layers N : When compiled with appropriate support (currently CLBlast or cuBLAS), this option allows offloading some layers to the GPU for computation. Generally results in increased performance.
	MainGpu                     int    //-mg i, --main-gpu i : When using multiple GPUs this option controls which GPU is used for small tensors for which the overhead of splitting the computation across all GPUs is not worthwhile. The GPU in question will use slightly more VRAM to store a scratch buffer for temporary results. By default GPU 0 is used. Requires cuBLAS.
	TensorSplit                 []int  //-ts SPLIT, --tensor-split SPLIT : When using multiple GPUs this option controls how large tensors should be split across all GPUs.  SPLIT  is a comma-separated list of non-negative values that assigns the proportion of data that each GPU should get in order. For example, "3,2" will assign 60% of the data to GPU 0 and 40% to GPU 1. By default the data is split in proportion to VRAM but this may not be optimal for performance. Requires cuBLAS.
	BatchSize                   int    //-b N ,  --batch-size N : Set the batch size for prompt processing. Default:  512 .
	MemoryF32                   bool   //--memory-f32 : Use 32-bit floats instead of 16-bit floats for memory key+value. Not recommended.
	Mlock                       bool   //--mlock : Lock the model in memory, preventing it from being swapped out when memory-mapped.
	NoMMap                      bool   //--no-mmap : Do not memory-map the model. By default, models are mapped into memory, which allows the system to load only the necessary parts of the model as needed.
	Numa                        bool   //--numa : Attempt optimizations that help on some NUMA systems.
	LoraFileName                string //--lora FNAME : Apply a LoRA (Low-Rank Adaptation) adapter to the model (implies --no-mmap). This allows you to adapt the pretrained model to specific tasks or domains.
	LoraBaseFileName            string //--lora-base FNAME : Optional model to use as a base for the layers modified by the LoRA adapter. This flag is used in conjunction with the  --lora  flag, and specifies the base model for the adaptation.
	Timeout                     int    //-to N ,  --timeout N : Server read/write timeout in seconds. Default  600 .
	ListenHost                  string //--host : Set the hostname or ip address to listen. Default  127.0.0.1 .
	ListenPort                  int    //--port: Set the port to listen. Default:  8080.
	StaticFilePath              string //--path: path from which to serve static files (default examples/server/public)
	EmbeddingExtraction         bool   //  --embedding : Enable embedding extraction, Default: disabled.
	ParallelSlots               int    // -np N ,  --parallel N : Set the number of slots for process requests (default: 1)
	ContinousBatching           bool   // -cb ,  --cont-batching : enable continuous batching (a.k.a dynamic batching) (default: disabled)
	SystemPromptFileName        string //-    -spf FNAME ,  --system-prompt-file FNAME  Set a file to load "a system prompt (initial prompt of all slots), this is useful for chat applications. [See more](#change-system-prompt-on-runtime)
	MultimodalProjectorFileName string //   --mmproj MMPROJ_FILE : Path to a multimodal projector file for LLaVA.
}

func (p *ServerCommandLineFlags) ToDefaults() {
	p.Threads = -1 //TODO
	p.ThreadsBatch = -1
	p.ModelFilename = "models/7B/ggml-model.gguf"
	p.Alias = ""
	p.CtxSize = 512
	p.NGPULayers = -1
	p.MainGpu = -1
	p.TensorSplit = []int{}
	p.BatchSize = 512
	p.MemoryF32 = false
	p.Mlock = false
	p.NoMMap = false
	p.Numa = false
	p.LoraFileName = ""
	p.LoraBaseFileName = ""
	p.Timeout = 600
	p.ListenHost = "127.0.0.1"
	p.ListenPort = 8080
	p.StaticFilePath = "examples/server/public"
	p.EmbeddingExtraction = false
	p.ParallelSlots = 1
	p.ContinousBatching = false
	p.SystemPromptFileName = ""
	p.MultimodalProjectorFileName = ""
}

func (p *ServerCommandLineFlags) GetArgs() ([]string, error) {
	d := ServerCommandLineFlags{}
	d.ToDefaults()

	fmt.Printf("\n\nDefaults are %#v\n\n", d)

	result := []string{}
	if p.Threads != d.Threads {
		result = append(result, "--threads", fmt.Sprintf("%v", p.Threads))
	}
	if p.ThreadsBatch != d.ThreadsBatch {
		result = append(result, "-tb", fmt.Sprintf("%v", p.ThreadsBatch))
	}
	if p.ModelFilename != d.ModelFilename {
		result = append(result, "-m", fmt.Sprintf("%v", p.ModelFilename))
	}
	if p.Alias != d.Alias {
		result = append(result, "--alias", fmt.Sprintf("%v", p.Alias))
	}
	if p.CtxSize != d.CtxSize {
		result = append(result, "-c", fmt.Sprintf("%v", p.CtxSize))
	}
	if p.NGPULayers != d.NGPULayers {
		result = append(result, "-ngl", fmt.Sprintf("%v", p.NGPULayers))
	}
	if p.MainGpu != d.MainGpu {
		result = append(result, "-mg", fmt.Sprintf("%v", p.MainGpu))
	}
	if 0 < len(p.TensorSplit) {
		s := fmt.Sprintf("%v", p.TensorSplit)
		result = append(result, "-ts", s[1:len(s)-2])
	}
	if p.BatchSize != d.BatchSize {
		result = append(result, "-b", fmt.Sprintf("%v", p.BatchSize))
	}
	if p.MemoryF32 != d.MemoryF32 {
		result = append(result, "--memory-f32", fmt.Sprintf("%v", p.MemoryF32))
	}
	if p.Mlock != d.Mlock {
		result = append(result, "--mlock", fmt.Sprintf("%v", p.Mlock))
	}
	if p.NoMMap != d.NoMMap {
		result = append(result, "--no-mmap", fmt.Sprintf("%v", p.NoMMap))
	}
	if p.Numa != d.Numa {
		result = append(result, "--numa", fmt.Sprintf("%v", p.Numa))
	}
	if p.LoraFileName != d.LoraFileName {
		result = append(result, "--lora", fmt.Sprintf("%v", p.LoraFileName))
	}
	if p.LoraBaseFileName != d.LoraBaseFileName {
		result = append(result, "--lora-base", fmt.Sprintf("%v", p.LoraBaseFileName))
	}
	if p.Timeout != d.Timeout {
		result = append(result, "-to", fmt.Sprintf("%v", p.Timeout))
	}
	if p.ListenHost != d.ListenHost {
		result = append(result, "--host", fmt.Sprintf("%v", p.ListenHost))
	}
	if p.ListenPort != d.ListenPort {

		result = append(result, "--port", fmt.Sprintf("%v", p.ListenPort))
	}
	if p.StaticFilePath != d.StaticFilePath {
		result = append(result, "--path", fmt.Sprintf("%v", p.StaticFilePath))
	}
	if p.EmbeddingExtraction != d.EmbeddingExtraction {
		result = append(result, "--embedding", fmt.Sprintf("%v", p.EmbeddingExtraction))
	}
	if p.ParallelSlots != d.ParallelSlots {
		result = append(result, "-np", fmt.Sprintf("%v", p.ParallelSlots))
	}
	if p.ContinousBatching != d.ContinousBatching {
		result = append(result, "-cb", fmt.Sprintf("%v", p.ContinousBatching))
	}
	if p.SystemPromptFileName != d.SystemPromptFileName {
		result = append(result, "-spf", fmt.Sprintf("%v", p.SystemPromptFileName))
	}
	if p.MultimodalProjectorFileName != d.MultimodalProjectorFileName {
		result = append(result, "--mmproj", fmt.Sprintf("%v", p.MultimodalProjectorFileName))
	}
	return result, nil
}

func (p *ServerCommandLineFlags) GetPort() error {
	var err error
	p.ListenPort, err = GetFreePort()
	return err
}
