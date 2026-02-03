package sortedagents

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Tool represents a function that can be called by the agent
type Tool interface {
	Name() string
	Description() string
	Parameters() *JSONSchema
	Execute(ctx context.Context, args map[string]any) (any, error)
}

// StructTool represents a tool that uses a specific struct for arguments
type StructTool[T any] interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args T) (any, error)
}

// structToolAdapter adapts a StructTool[T] to the standard Tool interface
type structToolAdapter[T any] struct {
	impl StructTool[T]
}

func (s *structToolAdapter[T]) Name() string {
	return s.impl.Name()
}

func (s *structToolAdapter[T]) Description() string {
	return s.impl.Description()
}

func (s *structToolAdapter[T]) Parameters() *JSONSchema {
	schema, _ := GenerateSchema[T]()
	return schema
}

func (s *structToolAdapter[T]) Execute(ctx context.Context, args map[string]any) (any, error) {
	jsonBytes, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal arguments: %w", err)
	}

	var typedArgs T
	if err := json.Unmarshal(jsonBytes, &typedArgs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments to %T: %w", typedArgs, err)
	}

	return s.impl.Execute(ctx, typedArgs)
}

// NewStructuredTool wraps a StructTool[T] into a standard Tool
func NewStructuredTool[T any](impl StructTool[T]) Tool {
	return &structToolAdapter[T]{
		impl: impl,
	}
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

// AddService inspects the given instance for exported methods and registers them as tools.
// It looks for methods with the signature: func (ctx context.Context, args StructType) (any, error)
// It uses "Description" method if available to map method names to descriptions.
func (b *ToolBuilder) AddService(instance any) *ToolBuilder {
	val := reflect.ValueOf(instance)
	typ := val.Type()

	// Check if instance implements a Description() map[string]string method
	descriptions := make(map[string]string)
	if descMethod := val.MethodByName("Description"); descMethod.IsValid() {
		// Expect signature: func() map[string]string
		if descMethod.Type().NumIn() == 0 && descMethod.Type().NumOut() == 1 {
			if descMethod.Type().Out(0).Kind() == reflect.Map {
				results := descMethod.Call(nil)
				if m, ok := results[0].Interface().(map[string]string); ok {
					descriptions = m
				}
			}
		}
	}

	for i := 0; i < val.NumMethod(); i++ {
		method := val.Method(i)
		methodType := method.Type()
		structMethod := typ.Method(i)
		methodName := structMethod.Name

		// Skip unexported methods or the Description method itself
		// PkgPath is empty for exported methods
		if structMethod.PkgPath != "" || methodName == "Description" {
			continue
		}

		// Expect signature: func (ctx context.Context, args StructType) (any, error)
		if methodType.NumIn() != 2 || methodType.NumOut() != 2 {
			continue
		}

		// Check first arg is Context
		if methodType.In(0).Name() != "Context" && !methodType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
			continue
		}

		// Check return types: (any, error)
		if !methodType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			continue
		}

		// Get argument type
		argType := methodType.In(1)

		// Create a wrapper function that calls the method via reflection
		// We need to capture 'method' and 'argType' in the closure
		toolName := toSnakeCase(methodName)
		toolDesc := descriptions[methodName]
		if toolDesc == "" {
			toolDesc = fmt.Sprintf("Execute %s", methodName)
		}

		// Use NewTool with a dynamic execution wrapper
		// Since we can't instantiate generic NewTool[T] with a runtime type easily without reflection trickery,
		// we'll implement a custom Tool for this reflection-based method.

		b.tools = append(b.tools, &reflectionTool{
			name:        toolName,
			description: toolDesc,
			method:      method,
			argType:     argType,
		})
	}

	return b
}

// reflectionTool adapts a reflected method to the Tool interface
type reflectionTool struct {
	name        string
	description string
	method      reflect.Value
	argType     reflect.Type
}

func (r *reflectionTool) Name() string        { return r.name }
func (r *reflectionTool) Description() string { return r.description }

func (r *reflectionTool) Parameters() *JSONSchema {
	schema, _ := GenerateSchemaReflect(r.argType)
	return schema
}

func (r *reflectionTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	// 1. Marshal args to JSON
	jsonBytes, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	// 2. Unmarshal to the specific arg struct type
	argVal := reflect.New(r.argType) // pointer to new instance
	if err := json.Unmarshal(jsonBytes, argVal.Interface()); err != nil {
		return nil, err
	}

	// 3. Call method: func(ctx, args)
	results := r.method.Call([]reflect.Value{reflect.ValueOf(ctx), argVal.Elem()})

	// 4. Handle returns: (any, error)
	res := results[0].Interface()
	errVal := results[1].Interface()

	if errVal != nil {
		return res, errVal.(error)
	}
	return res, nil
}

// toSnakeCase converts "CamelCase" to "camel_case"
func toSnakeCase(str string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
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
