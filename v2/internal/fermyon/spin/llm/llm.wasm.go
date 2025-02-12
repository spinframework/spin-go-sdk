// Code generated by wit-bindgen-go. DO NOT EDIT.

package llm

import (
	"go.bytecodealliance.org/cm"
)

// This file contains wasmimport and wasmexport declarations for "fermyon:spin".

//go:wasmimport fermyon:spin/llm infer
//go:noescape
func wasmimport_Infer(model0 *uint8, model1 uint32, prompt0 *uint8, prompt1 uint32, params0 uint32, params1 uint32, params2 float32, params3 uint32, params4 float32, params5 uint32, params6 float32, result *cm.Result[InferencingResultShape, InferencingResult, Error])

//go:wasmimport fermyon:spin/llm generate-embeddings
//go:noescape
func wasmimport_GenerateEmbeddings(model0 *uint8, model1 uint32, text0 *string, text1 uint32, result *cm.Result[EmbeddingsResultShape, EmbeddingsResult, Error])
