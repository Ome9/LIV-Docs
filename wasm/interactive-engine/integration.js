/**
 * JavaScript integration layer for the LIV Interactive Engine WASM module
 * 
 * This file demonstrates how to integrate the Rust WASM interactive engine
 * with the JavaScript viewer layer for secure interactive content execution.
 */

class LIVInteractiveEngine {
    constructor() {
        this.wasmModule = null;
        this.isInitialized = false;
        this.animationFrameId = null;
        this.lastFrameTime = 0;
    }

    /**
     * Initialize the WASM interactive engine with security permissions
     * @param {Object} permissions - Security permissions for the engine
     * @returns {Promise<void>}
     */
    async initialize(permissions = {}) {
        try {
            // Load the WASM module (this would be the actual WASM file in production)
            // For now, we'll simulate the interface
            this.wasmModule = await this.loadWASMModule();
            
            // Default permissions with security-first approach
            const defaultPermissions = {
                memory_limit: 4 * 1024 * 1024, // 4MB
                allowed_imports: ["env"],
                cpu_time_limit: 5000, // 5 seconds
                allow_networking: false,
                allow_file_system: false,
                allowed_interactions: ["Click", "Hover", "Touch", "Scroll", "DataUpdate"],
                max_data_size: 64 * 1024, // 64KB
                max_elements: 1000
            };

            const finalPermissions = { ...defaultPermissions, ...permissions };
            
            // Initialize the WASM engine with permissions
            await this.wasmModule.init_interactive_engine(JSON.stringify(finalPermissions));
            
            this.isInitialized = true;
            console.log('LIV Interactive Engine initialized successfully');
            
        } catch (error) {
            console.error('Failed to initialize LIV Interactive Engine:', error);
            throw error;
        }
    }

    /**
     * Process a user interaction event
     * @param {Object} event - The interaction event
     * @returns {Promise<Object>} Render update instructions
     */
    async processInteraction(event) {
        if (!this.isInitialized) {
            throw new Error('Engine not initialized');
        }

        try {
            const interactionEvent = {
                event_type: this.mapEventType(event.type),
                target_element: event.target?.id || null,
                position: event.clientX !== undefined ? {
                    x: event.clientX,
                    y: event.clientY
                } : null,
                data: this.extractEventData(event),
                timestamp: performance.now()
            };

            const updateJson = await this.wasmModule.process_interaction(JSON.stringify(interactionEvent));
            return JSON.parse(updateJson);
            
        } catch (error) {
            console.error('Failed to process interaction:', error);
            throw error;
        }
    }

    /**
     * Render a frame (called by animation loop)
     * @param {number} timestamp - Current timestamp
     * @returns {Promise<Object>} Render update instructions
     */
    async renderFrame(timestamp) {
        if (!this.isInitialized) {
            return { dom_operations: [], style_changes: [], animation_updates: [] };
        }

        try {
            const updateJson = await this.wasmModule.render_frame(timestamp);
            return JSON.parse(updateJson);
            
        } catch (error) {
            console.error('Failed to render frame:', error);
            return { dom_operations: [], style_changes: [], animation_updates: [] };
        }
    }

    /**
     * Update data source
     * @param {string} dataSourceId - ID of the data source
     * @param {Object} data - New data
     * @returns {Promise<void>}
     */
    async updateData(dataSourceId, data) {
        if (!this.isInitialized) {
            throw new Error('Engine not initialized');
        }

        try {
            const dataBytes = new TextEncoder().encode(JSON.stringify(data));
            await this.wasmModule.update_data(dataSourceId, dataBytes);
            
        } catch (error) {
            console.error('Failed to update data:', error);
            throw error;
        }
    }

    /**
     * Get performance statistics
     * @returns {Promise<Object>} Performance stats
     */
    async getPerformanceStats() {
        if (!this.isInitialized) {
            return null;
        }

        try {
            const statsJson = await this.wasmModule.get_performance_stats();
            return JSON.parse(statsJson);
            
        } catch (error) {
            console.error('Failed to get performance stats:', error);
            return null;
        }
    }

    /**
     * Start the animation loop
     */
    startAnimationLoop() {
        if (this.animationFrameId) {
            return; // Already running
        }

        const animate = async (timestamp) => {
            try {
                const renderUpdate = await this.renderFrame(timestamp);
                this.applyRenderUpdate(renderUpdate);
                
                this.lastFrameTime = timestamp;
                this.animationFrameId = requestAnimationFrame(animate);
                
            } catch (error) {
                console.error('Animation loop error:', error);
                this.stopAnimationLoop();
            }
        };

        this.animationFrameId = requestAnimationFrame(animate);
    }

    /**
     * Stop the animation loop
     */
    stopAnimationLoop() {
        if (this.animationFrameId) {
            cancelAnimationFrame(this.animationFrameId);
            this.animationFrameId = null;
        }
    }

    /**
     * Apply render updates to the DOM
     * @param {Object} renderUpdate - Render update from WASM engine
     */
    applyRenderUpdate(renderUpdate) {
        // Apply DOM operations
        for (const operation of renderUpdate.dom_operations || []) {
            this.applyDOMOperation(operation);
        }

        // Apply style changes
        for (const styleChange of renderUpdate.style_changes || []) {
            this.applyStyleChange(styleChange);
        }

        // Apply animation updates
        for (const animationUpdate of renderUpdate.animation_updates || []) {
            this.applyAnimationUpdate(animationUpdate);
        }
    }

    /**
     * Apply a single DOM operation
     * @param {Object} operation - DOM operation
     */
    applyDOMOperation(operation) {
        try {
            switch (operation.type || 'Update') {
                case 'Create':
                    this.createElement(operation);
                    break;
                case 'Update':
                    this.updateElement(operation);
                    break;
                case 'Remove':
                    this.removeElement(operation);
                    break;
                case 'Move':
                    this.moveElement(operation);
                    break;
            }
        } catch (error) {
            console.error('Failed to apply DOM operation:', error);
        }
    }

    /**
     * Apply a style change
     * @param {Object} styleChange - Style change
     */
    applyStyleChange(styleChange) {
        try {
            const element = document.getElementById(styleChange.element_id);
            if (element) {
                element.style[styleChange.property] = styleChange.value;
            }
        } catch (error) {
            console.error('Failed to apply style change:', error);
        }
    }

    /**
     * Apply an animation update
     * @param {Object} animationUpdate - Animation update
     */
    applyAnimationUpdate(animationUpdate) {
        try {
            // Apply current animation values
            for (const [property, value] of Object.entries(animationUpdate.current_values || {})) {
                // This would apply the animated values to the target elements
                // Implementation depends on the specific animation system
            }
        } catch (error) {
            console.error('Failed to apply animation update:', error);
        }
    }

    /**
     * Create a new DOM element
     * @param {Object} operation - Create operation
     */
    createElement(operation) {
        const element = document.createElement(operation.tag);
        element.id = operation.element_id;
        
        if (operation.parent_id) {
            const parent = document.getElementById(operation.parent_id);
            if (parent) {
                parent.appendChild(element);
            }
        }
    }

    /**
     * Update an existing DOM element
     * @param {Object} operation - Update operation
     */
    updateElement(operation) {
        const element = document.getElementById(operation.element_id);
        if (element && operation.attributes) {
            for (const [attr, value] of Object.entries(operation.attributes)) {
                element.setAttribute(attr, value);
            }
        }
    }

    /**
     * Remove a DOM element
     * @param {Object} operation - Remove operation
     */
    removeElement(operation) {
        const element = document.getElementById(operation.element_id);
        if (element && element.parentNode) {
            element.parentNode.removeChild(element);
        }
    }

    /**
     * Move a DOM element
     * @param {Object} operation - Move operation
     */
    moveElement(operation) {
        const element = document.getElementById(operation.element_id);
        const newParent = document.getElementById(operation.new_parent_id);
        
        if (element && newParent) {
            newParent.appendChild(element);
        }
    }

    /**
     * Map DOM event types to WASM event types
     * @param {string} domEventType - DOM event type
     * @returns {string} WASM event type
     */
    mapEventType(domEventType) {
        const eventMap = {
            'click': 'Click',
            'mouseenter': 'Hover',
            'mouseleave': 'Hover',
            'touchstart': 'Touch',
            'touchend': 'Touch',
            'scroll': 'Scroll',
            'keydown': 'Keyboard',
            'keyup': 'Keyboard',
            'resize': 'Resize'
        };

        return eventMap[domEventType] || 'Unknown';
    }

    /**
     * Extract relevant data from DOM events
     * @param {Event} event - DOM event
     * @returns {Object} Event data
     */
    extractEventData(event) {
        const data = {};

        if (event.key) {
            data.key = event.key;
        }

        if (event.deltaX !== undefined || event.deltaY !== undefined) {
            data.deltaX = event.deltaX || 0;
            data.deltaY = event.deltaY || 0;
        }

        if (event.touches) {
            data.touches = Array.from(event.touches).map(touch => ({
                x: touch.clientX,
                y: touch.clientY,
                identifier: touch.identifier
            }));
        }

        return data;
    }

    /**
     * Load the WASM module (placeholder for actual WASM loading)
     * @returns {Promise<Object>} WASM module interface
     */
    async loadWASMModule() {
        // In a real implementation, this would load the actual WASM file
        // For now, return a mock interface
        return {
            init_interactive_engine: async (permissions) => {
                console.log('Mock: Initializing engine with permissions:', permissions);
            },
            process_interaction: async (event) => {
                console.log('Mock: Processing interaction:', event);
                return JSON.stringify({
                    dom_operations: [],
                    style_changes: [],
                    animation_updates: [],
                    timestamp: performance.now()
                });
            },
            render_frame: async (timestamp) => {
                return JSON.stringify({
                    dom_operations: [],
                    style_changes: [],
                    animation_updates: [],
                    timestamp: timestamp
                });
            },
            update_data: async (dataSourceId, data) => {
                console.log('Mock: Updating data source:', dataSourceId, data);
            },
            get_performance_stats: async () => {
                return JSON.stringify({
                    interactions_per_second: 0,
                    renders_per_second: 60,
                    total_interactions: 0,
                    total_renders: 0,
                    uptime_ms: performance.now()
                });
            },
            destroy_engine: () => {
                console.log('Mock: Destroying engine');
            }
        };
    }

    /**
     * Destroy the engine and clean up resources
     */
    async destroy() {
        this.stopAnimationLoop();
        
        if (this.wasmModule && this.isInitialized) {
            await this.wasmModule.destroy_engine();
        }
        
        this.wasmModule = null;
        this.isInitialized = false;
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = LIVInteractiveEngine;
} else if (typeof window !== 'undefined') {
    window.LIVInteractiveEngine = LIVInteractiveEngine;
}

// Example usage
async function exampleUsage() {
    const engine = new LIVInteractiveEngine();
    
    try {
        // Initialize with custom permissions
        await engine.initialize({
            memory_limit: 8 * 1024 * 1024, // 8MB
            allowed_interactions: ["Click", "Hover", "Touch"],
            max_elements: 500
        });
        
        // Start animation loop
        engine.startAnimationLoop();
        
        // Set up event listeners
        document.addEventListener('click', async (event) => {
            try {
                const renderUpdate = await engine.processInteraction(event);
                console.log('Render update:', renderUpdate);
            } catch (error) {
                console.error('Interaction processing failed:', error);
            }
        });
        
        // Update data periodically
        setInterval(async () => {
            try {
                await engine.updateData('chart_data', {
                    values: [Math.random() * 100, Math.random() * 100, Math.random() * 100]
                });
            } catch (error) {
                console.error('Data update failed:', error);
            }
        }, 1000);
        
        // Get performance stats periodically
        setInterval(async () => {
            const stats = await engine.getPerformanceStats();
            if (stats) {
                console.log('Performance stats:', stats);
            }
        }, 5000);
        
    } catch (error) {
        console.error('Engine initialization failed:', error);
    }
}

// Uncomment to run example
// exampleUsage();