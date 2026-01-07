package sortedagents

import (
	"context"
	"encoding/json"
)

// Tool represents a function that can be called by the agent
type Tool interface {
	Name() string
	Description() string
	Parameters() *JSONSchema
	Execute(ctx context.Context, args map[string]any) (any, error)
}

// ToolFunc is a function signature that can be wrapped as a Tool
type ToolFunc func(ctx context.Context, args map[string]any) (any, error)

// funcTool wraps a function to implement the Tool interface
type funcTool struct {
	name        string
	description string
	parameters  *JSONSchema
	fn          ToolFunc
}

// Name returns the tool's name
func (f *funcTool) Name() string {
	return f.name
}

// Description returns the tool's description
func (f *funcTool) Description() string {
	return f.description
}

// Parameters returns the tool's parameter schema
func (f *funcTool) Parameters() *JSONSchema {
	if f.parameters == nil {
		return &JSONSchema{
			Type:       "object",
			Properties: make(map[string]JSONSchema),
		}
	}
	return f.parameters
}

// Execute runs the wrapped function
func (f *funcTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	return f.fn(ctx, args)
}

// Func creates a Tool from a function with name and description
func Func(name, description string, fn ToolFunc) Tool {
	return &funcTool{
		name:        name,
		description: description,
		fn:          fn,
	}
}

// TypedToolFunc is a generic function signature with a concrete struct for arguments
type TypedToolFunc[T any] func(ctx context.Context, args T) (any, error)

// typedTool wraps a generic function to implement the Tool interface
type typedTool[T any] struct {
	name        string
	description string
	fn          TypedToolFunc[T]
}

func (t *typedTool[T]) Name() string {
	return t.name
}

func (t *typedTool[T]) Description() string {
	return t.description
}

func (t *typedTool[T]) Parameters() *JSONSchema {
	schema, _ := GenerateSchema[T]()
	return schema
}

func (t *typedTool[T]) Execute(ctx context.Context, args map[string]any) (any, error) {
	// Convert map to JSON then to struct T
	// This is a robust way to handle type conversion (e.g. float64 to int)
	jsonData, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	var typedArgs T
	if err := json.Unmarshal(jsonData, &typedArgs); err != nil {
		return nil, err
	}

	return t.fn(ctx, typedArgs)
}

// NewTool creates a type-safe tool from a generic function
func NewTool[T any](name, description string, fn TypedToolFunc[T]) Tool {
	return &typedTool[T]{
		name:        name,
		description: description,
		fn:          fn,
	}
}

// Method creates a Tool from a struct method with compile-time type safety
func Method[T any](instance *T, name, description string, method func(*T, context.Context, map[string]any) (any, error)) Tool {
	return Func(name, description, func(ctx context.Context, args map[string]any) (any, error) {
		return method(instance, ctx, args)
	})
}

// TypedMethod creates a type-safe Tool from a struct method
func TypedMethod[R any, A any](instance *R, name, description string, method func(*R, context.Context, A) (any, error)) Tool {
	return NewTool(name, description, func(ctx context.Context, args A) (any, error) {
		return method(instance, ctx, args)
	})
}

// ToolBuilder provides a type-safe way to build tools from struct methods
type ToolBuilder struct {
	tools []Tool
}

// NewToolBuilder creates a new ToolBuilder instance
func NewToolBuilder() *ToolBuilder {
	return &ToolBuilder{
		tools: make([]Tool, 0),
	}
}

// AddTypedMethod adds a struct method as a type-safe tool
func (b *ToolBuilder) AddTypedMethod(tool Tool) *ToolBuilder {
	b.tools = append(b.tools, tool)
	return b
}

// AddTypedFunc adds a standalone function as a type-safe tool
func (b *ToolBuilder) AddTypedFunc(tool Tool) *ToolBuilder {
	b.tools = append(b.tools, tool)
	return b
}

// AddMethod adds a struct method as a tool with compile-time type safety
// Since Go methods can't have type parameters, we use a standalone function
func (b *ToolBuilder) AddMethod(tool Tool) *ToolBuilder {
	b.tools = append(b.tools, tool)
	return b
}

// AddFunc adds a standalone function as a tool (same as Func but chainable)
func (b *ToolBuilder) AddFunc(name, description string, fn ToolFunc) *ToolBuilder {
	b.tools = append(b.tools, Func(name, description, fn))
	return b
}

// AddTool adds an existing Tool implementation
func (b *ToolBuilder) AddTool(tool Tool) *ToolBuilder {
	b.tools = append(b.tools, tool)
	return b
}

// Build returns the final slice of tools
func (b *ToolBuilder) Build() []Tool {
	return b.tools
}

// Agent represents an AI agent configuration
type Agent interface {
	Name() string
	Instructions() string
	Model() string
	Tools() []Tool
}

// BasicAgent is a simple implementation of the Agent interface
type BasicAgent struct {
	name         string
	instructions string
	model        string
	tools        []Tool
}

// NewAgent creates a new BasicAgent instance
func NewAgent(name, instructions, model string, tools []Tool) *BasicAgent {
	return &BasicAgent{
		name:         name,
		instructions: instructions,
		model:        model,
		tools:        tools,
	}
}

// Name returns the agent's name
func (a *BasicAgent) Name() string {
	return a.name
}

// Instructions returns the agent's instructions
func (a *BasicAgent) Instructions() string {
	return a.instructions
}

// Model returns the model name
func (a *BasicAgent) Model() string {
	return a.model
}

// Tools returns the available tools
func (a *BasicAgent) Tools() []Tool {
	return a.tools
}
