// LIV Sample Document JavaScript

// Initialize the document when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    console.log('LIV Sample Document loaded');
    
    // Add some interactive behavior
    initializeInteractivity();
    
    // Add smooth scrolling
    addSmoothScrolling();
    
    // Add keyboard shortcuts
    addKeyboardShortcuts();
});

function initializeInteractivity() {
    // Add click handlers to feature list items
    const featureItems = document.querySelectorAll('.features li');
    featureItems.forEach(item => {
        item.addEventListener('click', function() {
            this.style.transform = 'scale(1.05)';
            this.style.transition = 'transform 0.2s ease';
            
            setTimeout(() => {
                this.style.transform = 'scale(1)';
            }, 200);
        });
    });
    
    // Add hover effects to sections
    const sections = document.querySelectorAll('section');
    sections.forEach(section => {
        section.addEventListener('mouseenter', function() {
            this.style.transform = 'translateY(-2px)';
            this.style.transition = 'transform 0.3s ease';
        });
        
        section.addEventListener('mouseleave', function() {
            this.style.transform = 'translateY(0)';
        });
    });
}

function toggleAnimation() {
    const demoBox = document.getElementById('demo-box');
    demoBox.classList.toggle('animate');
    
    // Update button text
    const button = demoBox.querySelector('button');
    if (demoBox.classList.contains('animate')) {
        button.textContent = 'Stop Animation';
    } else {
        button.textContent = 'Start Animation';
    }
    
    // Add some visual feedback
    button.style.transform = 'scale(0.95)';
    setTimeout(() => {
        button.style.transform = 'scale(1)';
    }, 100);
}

function addSmoothScrolling() {
    // Add smooth scrolling to any anchor links
    const links = document.querySelectorAll('a[href^="#"]');
    links.forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            
            const targetId = this.getAttribute('href').substring(1);
            const targetElement = document.getElementById(targetId);
            
            if (targetElement) {
                targetElement.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });
}

function addKeyboardShortcuts() {
    document.addEventListener('keydown', function(e) {
        // Space bar to toggle animation
        if (e.code === 'Space' && !e.target.matches('input, textarea, button')) {
            e.preventDefault();
            toggleAnimation();
        }
        
        // 'R' key to reset animations
        if (e.code === 'KeyR' && !e.target.matches('input, textarea')) {
            const demoBox = document.getElementById('demo-box');
            demoBox.classList.remove('animate');
            const button = demoBox.querySelector('button');
            button.textContent = 'Start Animation';
        }
    });
}

// Add some dynamic content updates
function updateTimestamp() {
    const now = new Date();
    const timestamp = now.toLocaleString();
    
    // Add timestamp to footer if it doesn't exist
    const footer = document.querySelector('footer');
    let timestampElement = footer.querySelector('.timestamp');
    
    if (!timestampElement) {
        timestampElement = document.createElement('div');
        timestampElement.className = 'timestamp';
        timestampElement.style.fontSize = '0.8rem';
        timestampElement.style.marginTop = '0.5rem';
        timestampElement.style.opacity = '0.7';
        footer.appendChild(timestampElement);
    }
    
    timestampElement.textContent = `Last updated: ${timestamp}`;
}

// Update timestamp every minute
setInterval(updateTimestamp, 60000);
updateTimestamp(); // Initial call

// Add some console messages for debugging
console.log('LIV Document Features:');
console.log('- Interactive animations');
console.log('- Smooth scrolling');
console.log('- Keyboard shortcuts (Space, R)');
console.log('- Dynamic content updates');

// Export functions for potential WASM integration
if (typeof window !== 'undefined') {
    window.LIVSample = {
        toggleAnimation,
        updateTimestamp,
        version: '1.0.0'
    };
}