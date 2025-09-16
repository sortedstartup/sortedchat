package llamacpp

/*
#cgo CFLAGS: -I./llama.cpp/include -I./llama.cpp/ggml/include
#cgo LDFLAGS: -L./llama.cpp/build/src/ -L./llama.cpp/build/ggml/src/ -l:libllama.a -l:libggml.a -l:libggml-base.a -l:libggml-cpu.a -lm -fopenmp -lstdc++
#include "llama.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"path/filepath"
	"sync"
	"unsafe"
)

const modelDir = "/home/xask/projects/sortedstartup/experiments/llamacpp/llama.cpp/"

// Global model cache
var (
	modelCache = make(map[string]*Model)
	cacheMutex sync.RWMutex
)

// Model represents a loaded LLaMA model with its associated resources
type Model struct {
	model   *C.struct_llama_model
	vocab   *C.struct_llama_vocab
	ctx     *C.struct_llama_context
	sampler *C.struct_llama_sampler
}

// LoadModel loads a LLaMA model from the specified file path
func LoadModel(modelPath string, nGpuLayers int) (*Model, error) {
	// Load dynamic backends
	C.ggml_backend_load_all()

	// Initialize the model
	modelParams := C.llama_model_default_params()
	modelParams.n_gpu_layers = C.int(nGpuLayers)

	cModelPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cModelPath))

	model := C.llama_model_load_from_file(cModelPath, modelParams)
	if model == nil {
		return nil, fmt.Errorf("unable to load model from %s", modelPath)
	}

	// Get vocabulary
	vocab := C.llama_model_get_vocab(model)

	// Initialize the context with default parameters
	ctxParams := C.llama_context_default_params()
	ctxParams.n_ctx = 4096  // Default context size
	ctxParams.n_batch = 512 // Default batch size
	ctxParams.no_perf = C.bool(false)

	ctx := C.llama_init_from_model(model, ctxParams)
	if ctx == nil {
		C.llama_model_free(model)
		return nil, fmt.Errorf("failed to create llama context")
	}

	// Initialize the sampler with better parameters to avoid repetition
	sparams := C.llama_sampler_chain_default_params()
	sparams.no_perf = C.bool(false)
	sampler := C.llama_sampler_chain_init(sparams)

	// Add repetition penalty to prevent loops
	C.llama_sampler_chain_add(sampler, C.llama_sampler_init_penalties(
		64,   // last_n tokens to consider for repetition penalty
		1.10, // repeat_penalty (> 1.0 discourages repetition)
		0.0,  // freq_penalty
		0.0,  // present_penalty
	))

	// Add top-k sampling to introduce some randomness
	C.llama_sampler_chain_add(sampler, C.llama_sampler_init_top_k(40))

	// Add temperature sampling for controlled randomness
	C.llama_sampler_chain_add(sampler, C.llama_sampler_init_temp(0.8))

	// Keep greedy as fallback
	C.llama_sampler_chain_add(sampler, C.llama_sampler_init_greedy())

	return &Model{
		model:   model,
		vocab:   vocab,
		ctx:     ctx,
		sampler: sampler,
	}, nil
}

// Close frees all resources associated with the model
func (m *Model) Close() {
	if m.sampler != nil {
		C.llama_sampler_free(m.sampler)
		m.sampler = nil
	}
	if m.ctx != nil {
		C.llama_free(m.ctx)
		m.ctx = nil
	}
	if m.model != nil {
		C.llama_model_free(m.model)
		m.model = nil
	}
}

// LoadModelByName loads a model by name from the modelDir with caching
func LoadModelByName(modelName string) (*Model, error) {
	// Check if model is already cached
	cacheMutex.RLock()
	if cachedModel, exists := modelCache[modelName]; exists {
		cacheMutex.RUnlock()
		return cachedModel, nil
	}
	cacheMutex.RUnlock()

	// Load model if not cached
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Double-check in case another goroutine loaded it while we were waiting
	if cachedModel, exists := modelCache[modelName]; exists {
		return cachedModel, nil
	}

	modelPath := filepath.Join(modelDir, modelName+".gguf")
	fmt.Println("Loading model from path:", modelPath)
	// Use 0 GPU layers for CPU-only inference
	model, err := LoadModel(modelPath, 0)
	if err != nil {
		return nil, err
	}

	// Cache the model
	modelCache[modelName] = model
	return model, nil
}

// UnloadModel removes a model from cache and frees its resources
func UnloadModel(modelName string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if model, exists := modelCache[modelName]; exists {
		model.Close()
		delete(modelCache, modelName)
	}
}

// UnloadAllModels removes all models from cache and frees their resources
func UnloadAllModels() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	for name, model := range modelCache {
		model.Close()
		delete(modelCache, name)
	}
}

// Predict generates text based on the input prompt and streams the output through a channel
func Predict(model *Model, text string) (<-chan string, error) {
	if model == nil || model.model == nil {
		return nil, fmt.Errorf("model is not loaded")
	}

	// Create output channel
	output := make(chan string, 100)

	go func() {
		defer close(output)

		// Tokenize the prompt
		cPrompt := C.CString(text)
		defer C.free(unsafe.Pointer(cPrompt))

		// Find the number of tokens in the prompt
		nPrompt := int(-C.llama_tokenize(model.vocab, cPrompt, C.int32_t(len(text)), nil, 0, C.bool(true), C.bool(true)))
		if nPrompt <= 0 {
			output <- fmt.Sprintf("Error: failed to tokenize prompt")
			return
		}

		// Allocate space for tokens and tokenize the prompt
		promptTokens := make([]C.llama_token, nPrompt)
		var promptTokensPtr *C.llama_token
		if nPrompt > 0 {
			promptTokensPtr = &promptTokens[0]
		}

		if C.llama_tokenize(model.vocab, cPrompt, C.int32_t(len(text)), promptTokensPtr, C.int32_t(nPrompt), C.bool(true), C.bool(true)) < 0 {
			output <- fmt.Sprintf("Error: failed to tokenize the prompt")
			return
		}

		// Update context size based on prompt length
		nPredict := 512 // Default prediction length
		ctxParams := C.llama_context_default_params()
		ctxParams.n_ctx = C.uint32_t(nPrompt + nPredict)
		ctxParams.n_batch = C.uint32_t(nPrompt)
		ctxParams.no_perf = C.bool(false)

		// Send prompt tokens as output first (echo the input)
		for _, token := range promptTokens {
			buf := make([]C.char, 128)
			n := C.llama_token_to_piece(model.vocab, token, &buf[0], C.int32_t(len(buf)), 0, C.bool(true))
			if n > 0 {
				s := C.GoStringN(&buf[0], n)
				output <- s
			}
		}

		// Prepare a batch for the prompt
		batch := C.llama_batch_get_one(promptTokensPtr, C.int(len(promptTokens)))

		// Main generation loop
		nDecode := 0
		var newTokenId C.llama_token

		for nPos := 0; nPos+int(batch.n_tokens) < nPrompt+nPredict; {
			// Evaluate the current batch with the transformer model
			if C.llama_decode(model.ctx, batch) != 0 {
				output <- fmt.Sprintf("Error: failed to decode")
				return
			}

			nPos += int(batch.n_tokens)

			// Sample the next token
			newTokenId = C.llama_sampler_sample(model.sampler, model.ctx, -1)

			// Check if it's end of generation
			if C.llama_vocab_is_eog(model.vocab, newTokenId) {
				break
			}

			// Convert token to text and send to channel
			buf := make([]C.char, 128)
			n := C.llama_token_to_piece(model.vocab, newTokenId, &buf[0], C.int32_t(len(buf)), 0, C.bool(true))
			if n > 0 {
				s := C.GoStringN(&buf[0], n)
				output <- s
			}

			// Prepare the next batch with the sampled token
			batch = C.llama_batch_get_one(&newTokenId, 1)
			nDecode++
		}
	}()

	return output, nil
}
