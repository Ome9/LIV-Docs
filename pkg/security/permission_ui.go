// Permission Management UI HTML Generation
// Generates the web interface for permission management

package security

// generatePermissionDashboardHTML generates the HTML for the permission management dashboard
func (pm *PermissionManager) generatePermissionDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LIV Security Permission Management</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: #f5f5f7;
            color: #1d1d1f;
            line-height: 1.6;
        }

        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 2rem 0;
            text-align: center;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }

        .header h1 {
            font-size: 2.5rem;
            font-weight: 600;
            margin-bottom: 0.5rem;
        }

        .header p {
            font-size: 1.1rem;
            opacity: 0.9;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 2rem;
        }

        .dashboard-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 2rem;
            margin-bottom: 2rem;
        }

        .card {
            background: white;
            border-radius: 12px;
            padding: 1.5rem;
            box-shadow: 0 4px 20px rgba(0,0,0,0.08);
            border: 1px solid #e5e5e7;
        }

        .card h2 {
            font-size: 1.5rem;
            margin-bottom: 1rem;
            color: #1d1d1f;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .card-icon {
            width: 24px;
            height: 24px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            border-radius: 6px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-weight: bold;
        }

        .permission-form {
            display: flex;
            flex-direction: column;
            gap: 1rem;
        }

        .form-group {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
        }

        .form-group label {
            font-weight: 500;
            color: #424245;
        }

        .form-group input,
        .form-group select,
        .form-group textarea {
            padding: 0.75rem;
            border: 1px solid #d2d2d7;
            border-radius: 8px;
            font-size: 1rem;
            transition: border-color 0.2s;
        }

        .form-group input:focus,
        .form-group select:focus,
        .form-group textarea:focus {
            outline: none;
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }

        .checkbox-group {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .checkbox-group input[type="checkbox"] {
            width: auto;
        }

        .btn {
            padding: 0.75rem 1.5rem;
            border: none;
            border-radius: 8px;
            font-size: 1rem;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s;
        }

        .btn-primary {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }

        .btn-primary:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
        }

        .btn-secondary {
            background: #f5f5f7;
            color: #1d1d1f;
            border: 1px solid #d2d2d7;
        }

        .btn-secondary:hover {
            background: #e8e8ed;
        }

        .evaluation-result {
            margin-top: 1rem;
            padding: 1rem;
            border-radius: 8px;
            display: none;
        }

        .evaluation-result.granted {
            background: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
        }

        .evaluation-result.denied {
            background: #f8d7da;
            border: 1px solid #f5c6cb;
            color: #721c24;
        }

        .warning-list {
            margin-top: 0.5rem;
        }

        .warning-item {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            color: #856404;
            padding: 0.5rem;
            border-radius: 4px;
            margin-bottom: 0.5rem;
            font-size: 0.9rem;
        }

        .restriction-list {
            margin-top: 0.5rem;
        }

        .restriction-item {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            color: #495057;
            padding: 0.5rem;
            border-radius: 4px;
            margin-bottom: 0.5rem;
            font-size: 0.9rem;
        }

        .template-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 1rem;
            margin-top: 1rem;
        }

        .template-card {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 8px;
            padding: 1rem;
            cursor: pointer;
            transition: all 0.2s;
        }

        .template-card:hover {
            background: #e9ecef;
            transform: translateY(-2px);
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }

        .template-card.selected {
            background: #e7f3ff;
            border-color: #667eea;
        }

        .template-card h3 {
            font-size: 1.1rem;
            margin-bottom: 0.5rem;
            color: #1d1d1f;
        }

        .template-card p {
            font-size: 0.9rem;
            color: #6c757d;
            margin-bottom: 0.5rem;
        }

        .template-card .use-case {
            font-size: 0.8rem;
            color: #495057;
            font-style: italic;
        }

        .policy-list {
            max-height: 300px;
            overflow-y: auto;
        }

        .policy-item {
            padding: 0.75rem;
            border-bottom: 1px solid #e5e5e7;
            cursor: pointer;
            transition: background-color 0.2s;
        }

        .policy-item:hover {
            background: #f8f9fa;
        }

        .policy-item.selected {
            background: #e7f3ff;
            border-left: 3px solid #667eea;
        }

        .policy-name {
            font-weight: 500;
            margin-bottom: 0.25rem;
        }

        .policy-description {
            font-size: 0.9rem;
            color: #6c757d;
        }

        .loading {
            display: none;
            text-align: center;
            padding: 2rem;
            color: #6c757d;
        }

        .spinner {
            border: 3px solid #f3f3f3;
            border-top: 3px solid #667eea;
            border-radius: 50%;
            width: 30px;
            height: 30px;
            animation: spin 1s linear infinite;
            margin: 0 auto 1rem;
        }

        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }

        .full-width {
            grid-column: 1 / -1;
        }

        @media (max-width: 768px) {
            .dashboard-grid {
                grid-template-columns: 1fr;
            }
            
            .container {
                padding: 1rem;
            }
            
            .header h1 {
                font-size: 2rem;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🔐 LIV Security Permission Management</h1>
        <p>Manage granular permissions, inheritance, and security policies</p>
    </div>

    <div class="container">
        <div class="dashboard-grid">
            <!-- Permission Evaluation -->
            <div class="card">
                <h2>
                    <div class="card-icon">🔍</div>
                    Permission Evaluation
                </h2>
                <form class="permission-form" id="evaluationForm">
                    <div class="form-group">
                        <label for="documentId">Document ID</label>
                        <input type="text" id="documentId" placeholder="Enter document ID" required>
                    </div>
                    
                    <div class="form-group">
                        <label for="moduleName">WASM Module Name (Optional)</label>
                        <input type="text" id="moduleName" placeholder="Enter module name">
                    </div>
                    
                    <div class="form-group">
                        <label for="policySelect">Security Policy</label>
                        <select id="policySelect" required>
                            <option value="">Select a policy...</option>
                        </select>
                    </div>
                    
                    <div class="form-group">
                        <label for="memoryLimit">Memory Limit (bytes)</label>
                        <input type="number" id="memoryLimit" value="16777216" min="1024" max="134217728">
                    </div>
                    
                    <div class="form-group">
                        <label for="cpuTimeLimit">CPU Time Limit (ms)</label>
                        <input type="number" id="cpuTimeLimit" value="5000" min="100" max="30000">
                    </div>
                    
                    <div class="checkbox-group">
                        <input type="checkbox" id="allowNetworking">
                        <label for="allowNetworking">Allow Networking</label>
                    </div>
                    
                    <div class="checkbox-group">
                        <input type="checkbox" id="allowFileSystem">
                        <label for="allowFileSystem">Allow File System Access</label>
                    </div>
                    
                    <div class="form-group">
                        <label for="allowedImports">Allowed Imports (comma-separated)</label>
                        <input type="text" id="allowedImports" placeholder="console, dom, events" value="console">
                    </div>
                    
                    <div class="form-group">
                        <label for="justification">Justification</label>
                        <textarea id="justification" rows="3" placeholder="Explain why these permissions are needed..."></textarea>
                    </div>
                    
                    <button type="submit" class="btn btn-primary">Evaluate Permissions</button>
                </form>
                
                <div id="evaluationResult" class="evaluation-result">
                    <h3 id="resultTitle"></h3>
                    <p id="resultMessage"></p>
                    <div id="warningsList" class="warning-list"></div>
                    <div id="restrictionsList" class="restriction-list"></div>
                </div>
            </div>

            <!-- Permission Templates -->
            <div class="card">
                <h2>
                    <div class="card-icon">📋</div>
                    Permission Templates
                </h2>
                <p>Select a pre-configured permission template:</p>
                <div id="templateGrid" class="template-grid">
                    <!-- Templates will be loaded here -->
                </div>
                <button id="applyTemplate" class="btn btn-secondary" style="margin-top: 1rem;" disabled>
                    Apply Selected Template
                </button>
            </div>

            <!-- Security Policies -->
            <div class="card">
                <h2>
                    <div class="card-icon">🛡️</div>
                    Security Policies
                </h2>
                <div id="policyList" class="policy-list">
                    <!-- Policies will be loaded here -->
                </div>
                <button id="refreshPolicies" class="btn btn-secondary" style="margin-top: 1rem;">
                    Refresh Policies
                </button>
            </div>

            <!-- Trust Chain Validation -->
            <div class="card">
                <h2>
                    <div class="card-icon">🔗</div>
                    Trust Chain Validation
                </h2>
                <div class="form-group">
                    <label for="trustDocumentId">Document ID for Trust Validation</label>
                    <input type="text" id="trustDocumentId" placeholder="Enter document ID">
                </div>
                <button id="validateTrust" class="btn btn-primary">Validate Trust Chain</button>
                <div id="trustResult" style="margin-top: 1rem; display: none;">
                    <h3>Trust Chain</h3>
                    <div id="trustChainList"></div>
                </div>
            </div>
        </div>

        <div class="loading" id="loading">
            <div class="spinner"></div>
            <p>Processing request...</p>
        </div>
    </div>

    <script>
        // Global state
        let selectedTemplate = null;
        let selectedPolicy = null;
        let policies = [];
        let templates = [];

        // Initialize the application
        document.addEventListener('DOMContentLoaded', function() {
            loadPolicies();
            loadTemplates();
            setupEventListeners();
        });

        // Setup event listeners
        function setupEventListeners() {
            document.getElementById('evaluationForm').addEventListener('submit', handlePermissionEvaluation);
            document.getElementById('applyTemplate').addEventListener('click', applySelectedTemplate);
            document.getElementById('refreshPolicies').addEventListener('click', loadPolicies);
            document.getElementById('validateTrust').addEventListener('click', validateTrustChain);
        }

        // Load security policies
        async function loadPolicies() {
            try {
                showLoading(true);
                const response = await fetch('/api/permissions/policies');
                policies = await response.json();
                
                renderPolicies();
                populatePolicySelect();
            } catch (error) {
                console.error('Failed to load policies:', error);
                alert('Failed to load security policies');
            } finally {
                showLoading(false);
            }
        }

        // Load permission templates
        async function loadTemplates() {
            try {
                const response = await fetch('/api/permissions/templates');
                templates = await response.json();
                renderTemplates();
            } catch (error) {
                console.error('Failed to load templates:', error);
                alert('Failed to load permission templates');
            }
        }

        // Render policies in the policy list
        function renderPolicies() {
            const policyList = document.getElementById('policyList');
            policyList.innerHTML = '';

            policies.forEach(policy => {
                const policyItem = document.createElement('div');
                policyItem.className = 'policy-item';
                policyItem.innerHTML = '<div class="policy-name">' + policy.name + '</div><div class="policy-description">' + policy.description + '</div>';
                
                policyItem.addEventListener('click', () => selectPolicy(policy));
                policyList.appendChild(policyItem);
            });
        }

        // Populate policy select dropdown
        function populatePolicySelect() {
            const policySelect = document.getElementById('policySelect');
            policySelect.innerHTML = '<option value="">Select a policy...</option>';

            policies.forEach(policy => {
                const option = document.createElement('option');
                option.value = policy.id;
                option.textContent = policy.name;
                policySelect.appendChild(option);
            });
        }

        // Render permission templates
        function renderTemplates() {
            const templateGrid = document.getElementById('templateGrid');
            templateGrid.innerHTML = '';

            templates.forEach(template => {
                const templateCard = document.createElement('div');
                templateCard.className = 'template-card';
                const h3 = document.createElement('h3');
                h3.textContent = template.name;
                const p = document.createElement('p');
                p.textContent = template.description;
                const useCase = document.createElement('div');
                useCase.className = 'use-case';
                useCase.textContent = 'Use case: ' + template.use_case;
                templateCard.appendChild(h3);
                templateCard.appendChild(p);
                templateCard.appendChild(useCase);
                
                templateCard.addEventListener('click', () => selectTemplate(template, templateCard));
                templateGrid.appendChild(templateCard);
            });
        }

        // Select a policy
        function selectPolicy(policy) {
            selectedPolicy = policy;
            
            // Update UI
            document.querySelectorAll('.policy-item').forEach(item => {
                item.classList.remove('selected');
            });
            event.currentTarget.classList.add('selected');
            
            // Update form
            document.getElementById('policySelect').value = policy.id;
        }

        // Select a template
        function selectTemplate(template, cardElement) {
            selectedTemplate = template;
            
            // Update UI
            document.querySelectorAll('.template-card').forEach(card => {
                card.classList.remove('selected');
            });
            cardElement.classList.add('selected');
            
            document.getElementById('applyTemplate').disabled = false;
        }

        // Apply selected template to form
        function applySelectedTemplate() {
            if (!selectedTemplate) return;
            
            const perms = selectedTemplate.permissions;
            document.getElementById('memoryLimit').value = perms.memory_limit;
            document.getElementById('cpuTimeLimit').value = perms.cpu_time_limit;
            document.getElementById('allowNetworking').checked = perms.allow_networking;
            document.getElementById('allowFileSystem').checked = perms.allow_file_system;
            document.getElementById('allowedImports').value = perms.allowed_imports.join(', ');
        }

        // Handle permission evaluation
        async function handlePermissionEvaluation(event) {
            event.preventDefault();
            
            const formData = new FormData(event.target);
            const request = {
                document_id: document.getElementById('documentId').value,
                module_name: document.getElementById('moduleName').value,
                policy_id: document.getElementById('policySelect').value,
                requested_permissions: {
                    memory_limit: parseInt(document.getElementById('memoryLimit').value),
                    cpu_time_limit: parseInt(document.getElementById('cpuTimeLimit').value),
                    allow_networking: document.getElementById('allowNetworking').checked,
                    allow_file_system: document.getElementById('allowFileSystem').checked,
                    allowed_imports: document.getElementById('allowedImports').value.split(',').map(s => s.trim()).filter(s => s)
                },
                user_context: {
                    user_id: 'current-user',
                    session_id: 'current-session',
                    ip_address: '127.0.0.1',
                    roles: ['user']
                },
                justification: document.getElementById('justification').value,
                requested_at: new Date().toISOString()
            };

            try {
                showLoading(true);
                const response = await fetch('/api/permissions/evaluate', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(request)
                });

                const evaluation = await response.json();
                displayEvaluationResult(evaluation);
            } catch (error) {
                console.error('Permission evaluation failed:', error);
                alert('Permission evaluation failed');
            } finally {
                showLoading(false);
            }
        }

        // Display evaluation result
        function displayEvaluationResult(evaluation) {
            const resultDiv = document.getElementById('evaluationResult');
            const titleEl = document.getElementById('resultTitle');
            const messageEl = document.getElementById('resultMessage');
            const warningsEl = document.getElementById('warningsList');
            const restrictionsEl = document.getElementById('restrictionsList');

            resultDiv.className = 'evaluation-result ' + (evaluation.granted ? 'granted' : 'denied');
            titleEl.textContent = evaluation.granted ? '✅ Permissions Granted' : '❌ Permissions Denied';
            
            let message = evaluation.granted ? 'The requested permissions have been granted.' : 'The requested permissions have been denied.';
            if (evaluation.inherited_from) {
                message += ' Permissions inherited from policy: ' + evaluation.inherited_from;
            }
            messageEl.textContent = message;

            // Display warnings
            warningsEl.innerHTML = '';
            if (evaluation.warnings && evaluation.warnings.length > 0) {
                evaluation.warnings.forEach(warning => {
                    const warningDiv = document.createElement('div');
                    warningDiv.className = 'warning-item';
                    const strong = document.createElement('strong');
                    strong.textContent = warning.type + ': ';
                    const text = document.createTextNode(warning.description);
                    const br = document.createElement('br');
                    const small = document.createElement('small');
                    small.textContent = 'Recommendation: ' + warning.recommendation;
                    warningDiv.appendChild(strong);
                    warningDiv.appendChild(text);
                    warningDiv.appendChild(br);
                    warningDiv.appendChild(small);
                    warningsEl.appendChild(warningDiv);
                });
            }

            // Display restrictions
            restrictionsEl.innerHTML = '';
            if (evaluation.restrictions && evaluation.restrictions.length > 0) {
                evaluation.restrictions.forEach(restriction => {
                    const restrictionDiv = document.createElement('div');
                    restrictionDiv.className = 'restriction-item';
                    restrictionDiv.textContent = restriction;
                    restrictionsEl.appendChild(restrictionDiv);
                });
            }

            resultDiv.style.display = 'block';
        }

        // Validate trust chain
        async function validateTrustChain() {
            const documentId = document.getElementById('trustDocumentId').value;
            if (!documentId) {
                alert('Please enter a document ID');
                return;
            }

            try {
                showLoading(true);
                const response = await fetch('/api/permissions/trust-chain?document_id=' + encodeURIComponent(documentId));
                const trustChain = await response.json();
                displayTrustChain(trustChain);
            } catch (error) {
                console.error('Trust chain validation failed:', error);
                alert('Trust chain validation failed');
            } finally {
                showLoading(false);
            }
        }

        // Display trust chain
        function displayTrustChain(trustChain) {
            const trustResult = document.getElementById('trustResult');
            const trustChainList = document.getElementById('trustChainList');

            trustChainList.innerHTML = '';
            trustChain.forEach((signer, index) => {
                const signerDiv = document.createElement('div');
                signerDiv.className = 'restriction-item';
                const strong = document.createElement('strong');
                strong.textContent = signer.name;
                const trustLevel = document.createTextNode(' (' + signer.trust_level + ')');
                const br1 = document.createElement('br');
                const small = document.createElement('small');
                small.textContent = 'Valid: ' + new Date(signer.valid_from).toLocaleDateString() + ' - ' + new Date(signer.valid_until).toLocaleDateString();
                signerDiv.appendChild(strong);
                signerDiv.appendChild(trustLevel);
                signerDiv.appendChild(br1);
                signerDiv.appendChild(small);
                if (signer.revoked) {
                    const br2 = document.createElement('br');
                    const revokedSpan = document.createElement('span');
                    revokedSpan.style.color = 'red';
                    revokedSpan.textContent = '⚠️ REVOKED';
                    signerDiv.appendChild(br2);
                    signerDiv.appendChild(revokedSpan);
                }
                trustChainList.appendChild(signerDiv);
            });

            trustResult.style.display = 'block';
        }

        // Show/hide loading indicator
        function showLoading(show) {
            document.getElementById('loading').style.display = show ? 'block' : 'none';
        }
    </script>
</body>
</html>`
}
