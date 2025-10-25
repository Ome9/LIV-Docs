// Simple interactive functionality for test document
document.addEventListener('DOMContentLoaded', function() {
    const interactiveElement = document.getElementById('test-element');
    
    if (interactiveElement) {
        interactiveElement.addEventListener('click', function() {
            this.style.transform = this.style.transform === 'scale(1.05)' ? 'scale(1)' : 'scale(1.05)';
            this.style.transition = 'transform 0.2s ease';
        });
    }
    
    console.log('LIV Test Document loaded successfully');
});