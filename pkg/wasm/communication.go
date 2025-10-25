package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// CommunicationBridge provides secure communication between Go and WASM
type CommunicationBridge struct {
	runtime         *WASMRuntime
	messageHandlers map[string]MessageHandler
	eventListeners  map[string][]EventListener
	messageQueue    chan *Message
	responseMap     map[string]chan *Message
	responseMutex   sync.RWMutex
	logger          core.Logger
	metrics         core.MetricsCollector
	shutdownChan    chan struct{}
	processingWG    sync.WaitGroup
}

// Message represents a communication message between Go and WASM
type Message struct {
	ID        string                 `json:"id"`
	Type      MessageType            `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
	Response  bool                   `json:"response"`
}

// MessageType defines the type of communication message
type MessageType string

const (
	MessageTypeFunction    MessageType = "function_call"
	MessageTypeEvent       MessageType = "event"
	MessageTypeData        MessageType = "data"
	MessageTypeControl     MessageType = "control"
	MessageTypeResponse    MessageType = "response"
	MessageTypeError       MessageType = "error"
	MessageTypeHeartbeat   MessageType = "heartbeat"
)

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, message *Message) (*Message, error)

// EventListener listens for specific events
type EventListener func(ctx context.Context, event *Message)

// CommunicationConfig holds configuration for the communication bridge
type CommunicationConfig struct {
	MessageQueueSize    int           `json:"message_queue_size"`
	ResponseTimeout     time.Duration `json:"response_timeout"`
	HeartbeatInterval   time.Duration `json:"heartbeat_interval"`
	MaxMessageSize      int           `json:"max_message_size"`
	EnableCompression   bool          `json:"enable_compression"`
	EnableEncryption    bool          `json:"enable_encryption"`
	LogAllMessages      bool          `json:"log_all_messages"`
}

// NewCommunicationBridge creates a new communication bridge
func NewCommunicationBridge(runtime *WASMRuntime, logger core.Logger, metrics core.MetricsCollector) *CommunicationBridge {
	return &CommunicationBridge{
		runtime:         runtime,
		messageHandlers: make(map[string]MessageHandler),
		eventListeners:  make(map[string][]EventListener),
		messageQueue:    make(chan *Message, 1000),
		responseMap:     make(map[string]chan *Message),
		logger:          logger,
		metrics:         metrics,
		shutdownChan:    make(chan struct{}),
	}
}

// Start starts the communication bridge
func (cb *CommunicationBridge) Start(ctx context.Context) error {
	cb.logger.Info("starting WASM communication bridge")

	// Register default message handlers
	cb.registerDefaultHandlers()

	// Start message processing loop
	cb.processingWG.Add(1)
	go cb.messageProcessingLoop(ctx)

	// Start heartbeat loop
	cb.processingWG.Add(1)
	go cb.heartbeatLoop(ctx)

	cb.logger.Info("WASM communication bridge started")
	return nil
}

// Stop stops the communication bridge
func (cb *CommunicationBridge) Stop() error {
	cb.logger.Info("stopping WASM communication bridge")

	close(cb.shutdownChan)
	cb.processingWG.Wait()

	// Close all pending response channels
	cb.responseMutex.Lock()
	for _, ch := range cb.responseMap {
		close(ch)
	}
	cb.responseMap = make(map[string]chan *Message)
	cb.responseMutex.Unlock()

	cb.logger.Info("WASM communication bridge stopped")
	return nil
}

// SendMessage sends a message to a WASM module
func (cb *CommunicationBridge) SendMessage(ctx context.Context, sandboxID string, moduleName string, message *Message) (*Message, error) {
	if message == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	// Set message metadata
	message.ID = cb.generateMessageID()
	message.Source = "go_runtime"
	message.Target = fmt.Sprintf("%s:%s", sandboxID, moduleName)
	message.Timestamp = time.Now()

	// Validate message size
	messageData, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	if len(messageData) > 1024*1024 { // 1MB limit
		return nil, fmt.Errorf("message size %d exceeds limit", len(messageData))
	}

	// Create response channel if expecting response
	var responseChan chan *Message
	if !message.Response {
		responseChan = make(chan *Message, 1)
		cb.responseMutex.Lock()
		cb.responseMap[message.ID] = responseChan
		cb.responseMutex.Unlock()

		// Clean up response channel after timeout
		go func() {
			time.Sleep(30 * time.Second) // 30 second timeout
			cb.responseMutex.Lock()
			delete(cb.responseMap, message.ID)
			cb.responseMutex.Unlock()
			close(responseChan)
		}()
	}

	// Send message to processing queue
	select {
	case cb.messageQueue <- message:
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, fmt.Errorf("message queue is full")
	}

	cb.logger.Debug("message sent to WASM module",
		"message_id", message.ID,
		"type", message.Type,
		"target", message.Target,
	)

	// For function calls, we need to process the message immediately
	if message.Type == MessageTypeFunction {
		// Process the function call directly
		handler, exists := cb.messageHandlers[string(message.Type)]
		if exists {
			response, err := handler(ctx, message)
			if err != nil {
				return nil, err
			}
			return response, nil
		}
	}

	// Wait for response if expected
	if responseChan != nil {
		select {
		case response := <-responseChan:
			return response, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second): // Reduced timeout for testing
			return nil, fmt.Errorf("response timeout for message %s", message.ID)
		}
	}

	return nil, nil
}

// CallWASMFunction calls a function in a WASM module with communication bridge
func (cb *CommunicationBridge) CallWASMFunction(ctx context.Context, sandboxID string, moduleName string, functionName string, args map[string]interface{}) (interface{}, error) {
	message := &Message{
		Type: MessageTypeFunction,
		Payload: map[string]interface{}{
			"function": functionName,
			"args":     args,
		},
	}

	response, err := cb.SendMessage(ctx, sandboxID, moduleName, message)
	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, fmt.Errorf("no response received")
	}

	if response.Type == MessageTypeError {
		return nil, fmt.Errorf("WASM function error: %v", response.Payload["error"])
	}

	return response.Payload["result"], nil
}

// SendEvent sends an event to WASM modules
func (cb *CommunicationBridge) SendEvent(ctx context.Context, eventType string, eventData map[string]interface{}) error {
	message := &Message{
		Type:     MessageTypeEvent,
		Response: true, // Events don't expect responses
		Payload: map[string]interface{}{
			"event_type": eventType,
			"data":       eventData,
		},
	}

	// Broadcast to all active sandboxes
	sandboxes := cb.runtime.ListActiveSandboxes()
	for _, sandboxID := range sandboxes {
		sandboxInfo, err := cb.runtime.GetSandboxInfo(sandboxID)
		if err != nil {
			continue
		}

		for moduleName := range sandboxInfo.LoadedModules {
			_, err := cb.SendMessage(ctx, sandboxID, moduleName, message)
			if err != nil {
				cb.logger.Warn("failed to send event to module",
					"sandbox_id", sandboxID,
					"module", moduleName,
					"event", eventType,
					"error", err,
				)
			}
		}
	}

	return nil
}

// RegisterMessageHandler registers a handler for a specific message type
func (cb *CommunicationBridge) RegisterMessageHandler(messageType string, handler MessageHandler) {
	cb.messageHandlers[messageType] = handler
	cb.logger.Debug("message handler registered", "type", messageType)
}

// RegisterEventListener registers a listener for a specific event type
func (cb *CommunicationBridge) RegisterEventListener(eventType string, listener EventListener) {
	if cb.eventListeners[eventType] == nil {
		cb.eventListeners[eventType] = []EventListener{}
	}
	cb.eventListeners[eventType] = append(cb.eventListeners[eventType], listener)
	cb.logger.Debug("event listener registered", "type", eventType)
}

// ProcessIncomingMessage processes a message from WASM
func (cb *CommunicationBridge) ProcessIncomingMessage(ctx context.Context, messageData []byte) error {
	var message Message
	if err := json.Unmarshal(messageData, &message); err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	cb.logger.Debug("processing incoming message",
		"message_id", message.ID,
		"type", message.Type,
		"source", message.Source,
	)

	// Handle response messages
	if message.Response {
		cb.responseMutex.RLock()
		responseChan, exists := cb.responseMap[message.ID]
		cb.responseMutex.RUnlock()

		if exists {
			select {
			case responseChan <- &message:
			default:
				cb.logger.Warn("response channel full", "message_id", message.ID)
			}
		}
		return nil
	}

	// Handle regular messages
	handler, exists := cb.messageHandlers[string(message.Type)]
	if !exists {
		cb.logger.Warn("no handler for message type", "type", message.Type)
		return fmt.Errorf("no handler for message type: %s", message.Type)
	}

	response, err := handler(ctx, &message)
	if err != nil {
		cb.logger.Error("message handler error",
			"message_id", message.ID,
			"type", message.Type,
			"error", err,
		)

		// Send error response
		errorResponse := &Message{
			ID:       message.ID,
			Type:     MessageTypeError,
			Source:   "go_runtime",
			Target:   message.Source,
			Response: true,
			Payload: map[string]interface{}{
				"error": err.Error(),
			},
			Timestamp: time.Now(),
		}

		cb.sendResponse(errorResponse)
		return err
	}

	// Send response if provided
	if response != nil {
		response.ID = message.ID
		response.Response = true
		response.Target = message.Source
		response.Timestamp = time.Now()
		cb.sendResponse(response)
	}

	return nil
}

// Helper methods

func (cb *CommunicationBridge) registerDefaultHandlers() {
	// Function call handler
	cb.RegisterMessageHandler(string(MessageTypeFunction), cb.handleFunctionCall)

	// Data handler
	cb.RegisterMessageHandler(string(MessageTypeData), cb.handleDataMessage)

	// Control handler
	cb.RegisterMessageHandler(string(MessageTypeControl), cb.handleControlMessage)

	// Event handler
	cb.RegisterMessageHandler(string(MessageTypeEvent), cb.handleEventMessage)
}

func (cb *CommunicationBridge) handleFunctionCall(ctx context.Context, message *Message) (*Message, error) {
	functionName, ok := message.Payload["function"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid function name in message")
	}

	args, ok := message.Payload["args"].(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}

	// Parse target (sandboxID:moduleName)
	sandboxID, moduleName, err := cb.parseTarget(message.Target)
	if err != nil {
		return nil, err
	}

	// Convert args to slice for WASM call
	argSlice := make([]interface{}, 0, len(args))
	for _, v := range args {
		argSlice = append(argSlice, v)
	}

	// Execute function in runtime
	result, err := cb.runtime.ExecuteInSandbox(ctx, sandboxID, moduleName, functionName, argSlice...)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:   MessageTypeResponse,
		Source: "go_runtime",
		Payload: map[string]interface{}{
			"result": result,
		},
	}, nil
}

func (cb *CommunicationBridge) handleDataMessage(ctx context.Context, message *Message) (*Message, error) {
	// Handle data transfer messages
	dataType, ok := message.Payload["type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid data type in message")
	}

	data := message.Payload["data"]

	cb.logger.Debug("handling data message",
		"data_type", dataType,
		"source", message.Source,
	)

	// Process data based on type
	switch dataType {
	case "memory_update":
		// Handle memory updates from WASM
		return cb.handleMemoryUpdate(ctx, message, data)
	case "resource_request":
		// Handle resource requests from WASM
		return cb.handleResourceRequest(ctx, message, data)
	default:
		cb.logger.Warn("unknown data type", "type", dataType)
	}

	return &Message{
		Type:   MessageTypeResponse,
		Source: "go_runtime",
		Payload: map[string]interface{}{
			"status": "processed",
		},
	}, nil
}

func (cb *CommunicationBridge) handleControlMessage(ctx context.Context, message *Message) (*Message, error) {
	command, ok := message.Payload["command"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid command in control message")
	}

	cb.logger.Debug("handling control message", "command", command)

	switch command {
	case "pause":
		// Handle pause command
		return &Message{
			Type:   MessageTypeResponse,
			Source: "go_runtime",
			Payload: map[string]interface{}{
				"status": "paused",
			},
		}, nil
	case "resume":
		// Handle resume command
		return &Message{
			Type:   MessageTypeResponse,
			Source: "go_runtime",
			Payload: map[string]interface{}{
				"status": "resumed",
			},
		}, nil
	case "terminate":
		// Handle terminate command
		sandboxID, _, err := cb.parseTarget(message.Target)
		if err == nil {
			cb.runtime.TerminateSandbox(sandboxID)
		}
		return &Message{
			Type:   MessageTypeResponse,
			Source: "go_runtime",
			Payload: map[string]interface{}{
				"status": "terminated",
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown control command: %s", command)
	}
}

func (cb *CommunicationBridge) handleEventMessage(ctx context.Context, message *Message) (*Message, error) {
	eventType, ok := message.Payload["event_type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid event type in message")
	}

	// Notify event listeners
	if listeners, exists := cb.eventListeners[eventType]; exists {
		for _, listener := range listeners {
			go listener(ctx, message)
		}
	}

	cb.logger.Debug("event processed", "type", eventType, "listeners", len(cb.eventListeners[eventType]))

	return nil, nil // Events don't need responses
}

func (cb *CommunicationBridge) handleMemoryUpdate(ctx context.Context, message *Message, data interface{}) (*Message, error) {
	// Handle memory usage updates from WASM modules
	memoryData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid memory data format")
	}

	usage, ok := memoryData["usage"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid memory usage value")
	}

	cb.logger.Debug("memory usage update", "usage", usage)

	return &Message{
		Type:   MessageTypeResponse,
		Source: "go_runtime",
		Payload: map[string]interface{}{
			"acknowledged": true,
		},
	}, nil
}

func (cb *CommunicationBridge) handleResourceRequest(ctx context.Context, message *Message, data interface{}) (*Message, error) {
	// Handle resource requests from WASM modules
	requestData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid resource request format")
	}

	resourceType, ok := requestData["type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid resource type")
	}

	cb.logger.Debug("resource request", "type", resourceType)

	// Process resource request based on type
	switch resourceType {
	case "memory":
		// Handle memory allocation request
		return &Message{
			Type:   MessageTypeResponse,
			Source: "go_runtime",
			Payload: map[string]interface{}{
				"granted": true,
				"amount":  requestData["amount"],
			},
		}, nil
	default:
		return &Message{
			Type:   MessageTypeResponse,
			Source: "go_runtime",
			Payload: map[string]interface{}{
				"granted": false,
				"reason":  "unsupported resource type",
			},
		}, nil
	}
}

func (cb *CommunicationBridge) parseTarget(target string) (string, string, error) {
	// Parse target format: "sandboxID:moduleName"
	parts := make([]string, 2)
	colonIndex := -1
	for i, char := range target {
		if char == ':' {
			colonIndex = i
			break
		}
	}

	if colonIndex == -1 {
		return "", "", fmt.Errorf("invalid target format: %s", target)
	}

	parts[0] = target[:colonIndex]
	parts[1] = target[colonIndex+1:]

	return parts[0], parts[1], nil
}

func (cb *CommunicationBridge) sendResponse(response *Message) {
	select {
	case cb.messageQueue <- response:
	default:
		cb.logger.Warn("failed to send response: queue full", "message_id", response.ID)
	}
}

func (cb *CommunicationBridge) generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

func (cb *CommunicationBridge) messageProcessingLoop(ctx context.Context) {
	defer cb.processingWG.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cb.shutdownChan:
			return
		case message := <-cb.messageQueue:
			cb.processMessage(ctx, message)
		}
	}
}

func (cb *CommunicationBridge) processMessage(ctx context.Context, message *Message) {
	if cb.metrics != nil {
		cb.metrics.RecordSecurityEvent("message_processed", map[string]interface{}{
			"message_type": string(message.Type),
			"source":       message.Source,
			"target":       message.Target,
		})
	}

	// Log message if enabled
	cb.logger.Debug("processing message",
		"id", message.ID,
		"type", message.Type,
		"source", message.Source,
		"target", message.Target,
	)
}

func (cb *CommunicationBridge) heartbeatLoop(ctx context.Context) {
	defer cb.processingWG.Done()

	ticker := time.NewTicker(30 * time.Second) // Heartbeat every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cb.shutdownChan:
			return
		case <-ticker.C:
			cb.sendHeartbeat()
		}
	}
}

func (cb *CommunicationBridge) sendHeartbeat() {
	heartbeat := &Message{
		Type:     MessageTypeHeartbeat,
		Source:   "go_runtime",
		Target:   "all",
		Response: true,
		Payload: map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"status":    "alive",
		},
		Timestamp: time.Now(),
	}

	select {
	case cb.messageQueue <- heartbeat:
	default:
		cb.logger.Warn("failed to send heartbeat: queue full")
	}
}