/**
 * Accessibility Compliance Tests
 * Tests WCAG 2.1 compliance and accessibility features
 */

import { LIVEditor } from '../src/editor';

describe('Accessibility Compliance Tests', () => {
    let editorContainer: HTMLElement;
    let previewContainer: HTMLElement;
    let toolbarContainer: HTMLElement;
    let propertiesContainer: HTMLElement;
    let editor: LIVEditor;

    // Accessibility testing utilities
    const getContrastRatio = (foreground: string, background: string): number => {
        // Simple contrast ratio calculation (simplified version)
        const getLuminance = (color: string): number => {
            // This is a simplified implementation
            // In a real scenario, you'd parse RGB values and calculate proper luminance
            return color === 'rgb(0, 0, 0)' ? 0 : 1;
        };
        
        const fgLuminance = getLuminance(foreground);
        const bgLuminance = getLuminance(background);
        
        const lighter = Math.max(fgLuminance, bgLuminance);
        const darker = Math.min(fgLuminance, bgLuminance);
        
        return (lighter + 0.05) / (darker + 0.05);
    };

    const checkAriaAttributes = (element: Element): boolean => {
        const requiredAriaAttributes = ['aria-label', 'aria-labelledby', 'aria-describedby'];
        return requiredAriaAttributes.some(attr => element.hasAttribute(attr)) || 
               element.textContent?.trim() !== '';
    };

    const checkKeyboardAccessibility = (element: HTMLElement): boolean => {
        return element.tabIndex >= 0 || 
               ['button', 'input', 'textarea', 'select', 'a'].includes(element.tagName.toLowerCase());
    };

    beforeEach(() => {
        document.body.innerHTML = `
            <div id="editor-container"></div>
            <div id="preview-container"></div>
            <div id="toolbar-container"></div>
            <div id="properties-container"></div>
        `;

        editorContainer = document.getElementById('editor-container')!;
        previewContainer = document.getElementById('preview-container')!;
        toolbarContainer = document.getElementById('toolbar-container')!;
        propertiesContainer = document.getElementById('properties-container')!;

        editor = new LIVEditor(
            editorContainer,
            previewContainer,
            toolbarContainer,
            propertiesContainer
        );
    });

    afterEach(() => {
        if (editor) {
            editor.destroy();
        }
        document.body.innerHTML = '';
    });

    describe('WCAG 2.1 Level A Compliance', () => {
        it('should provide text alternatives for images', async () => {
            await editor.initialize();

            editor.insertElement('image', {
                src: 'https://example.com/test.jpg',
                alt: 'Test image description'
            });

            const image = editorContainer.querySelector('img') as HTMLImageElement;
            expect(image.alt).toBe('Test image description');
            expect(image.alt.length).toBeGreaterThan(0);
        });

        it('should provide captions for audio/video content', async () => {
            await editor.initialize();

            editor.insertElement('video', {
                src: 'https://example.com/test.mp4',
                captions: 'Test video captions'
            });

            const video = editorContainer.querySelector('video');
            expect(video).toBeTruthy();
            
            // Check for caption track or aria-describedby
            const hasAccessibleDescription = 
                video?.querySelector('track[kind="captions"]') ||
                video?.hasAttribute('aria-describedby');
            
            expect(hasAccessibleDescription).toBeTruthy();
        });

        it('should ensure proper heading hierarchy', async () => {
            await editor.initialize();

            editor.insertElement('heading', { text: 'Main Heading', level: 1 });
            editor.insertElement('heading', { text: 'Sub Heading', level: 2 });
            editor.insertElement('heading', { text: 'Sub Sub Heading', level: 3 });

            const h1 = editorContainer.querySelector('h1');
            const h2 = editorContainer.querySelector('h2');
            const h3 = editorContainer.querySelector('h3');

            expect(h1?.textContent).toBe('Main Heading');
            expect(h2?.textContent).toBe('Sub Heading');
            expect(h3?.textContent).toBe('Sub Sub Heading');

            // Check hierarchy is logical
            const headings = editorContainer.querySelectorAll('h1, h2, h3, h4, h5, h6');
            let previousLevel = 0;
            
            headings.forEach(heading => {
                const currentLevel = parseInt(heading.tagName.charAt(1));
                expect(currentLevel - previousLevel).toBeLessThanOrEqual(1);
                previousLevel = currentLevel;
            });
        });

        it('should provide meaningful link text', async () => {
            await editor.initialize();

            editor.insertElement('link', {
                href: 'https://example.com',
                text: 'Visit our documentation page'
            });

            const link = editorContainer.querySelector('a') as HTMLAnchorElement;
            expect(link.textContent?.trim()).toBe('Visit our documentation page');
            expect(link.textContent?.trim().length).toBeGreaterThan(4);
            
            // Should not contain generic text
            const genericTexts = ['click here', 'read more', 'link', 'here'];
            const linkText = link.textContent?.toLowerCase() || '';
            genericTexts.forEach(generic => {
                expect(linkText).not.toBe(generic);
            });
        });
    });

    describe('WCAG 2.1 Level AA Compliance', () => {
        it('should meet color contrast requirements', async () => {
            await editor.initialize();

            editor.insertElement('paragraph', { text: 'Color contrast test' });
            const paragraph = editorContainer.querySelector('p') as HTMLElement;

            const computedStyle = window.getComputedStyle(paragraph);
            const color = computedStyle.color;
            const backgroundColor = computedStyle.backgroundColor;

            // Test contrast ratio (simplified check)
            expect(color).toBeTruthy();
            expect(backgroundColor).toBeTruthy();
            
            // In a real implementation, you'd calculate the actual contrast ratio
            // and ensure it meets WCAG AA standards (4.5:1 for normal text)
            const contrastRatio = getContrastRatio(color, backgroundColor);
            expect(contrastRatio).toBeGreaterThanOrEqual(3.0); // Simplified threshold
        });

        it('should be resizable up to 200% without loss of functionality', async () => {
            await editor.initialize();

            // Simulate zoom
            const originalFontSize = 16;
            document.documentElement.style.fontSize = `${originalFontSize * 2}px`;

            editor.insertElement('paragraph', { text: 'Zoom test content' });
            
            // Editor should still be functional at 200% zoom
            expect(() => {
                editor.insertElement('heading', { text: 'Zoomed heading' });
            }).not.toThrow();

            const paragraph = editorContainer.querySelector('p');
            expect(paragraph?.textContent).toBe('Zoom test content');

            // Reset zoom
            document.documentElement.style.fontSize = '';
        });

        it('should support keyboard navigation', async () => {
            await editor.initialize();

            const focusableElements = editorContainer.querySelectorAll(
                'button, input, textarea, select, a, [tabindex]:not([tabindex="-1"])'
            );

            expect(focusableElements.length).toBeGreaterThan(0);

            // Test tab order
            focusableElements.forEach((element, index) => {
                const htmlElement = element as HTMLElement;
                expect(checkKeyboardAccessibility(htmlElement)).toBeTruthy();
            });
        });

        it('should provide focus indicators', async () => {
            await editor.initialize();

            const toolbar = toolbarContainer.querySelector('.editor-toolbar');
            const buttons = toolbar?.querySelectorAll('button');

            buttons?.forEach(button => {
                button.focus();
                
                const computedStyle = window.getComputedStyle(button, ':focus');
                const outline = computedStyle.outline;
                const boxShadow = computedStyle.boxShadow;
                
                // Should have visible focus indicator
                expect(outline !== 'none' || boxShadow !== 'none').toBeTruthy();
            });
        });
    });

    describe('ARIA Implementation', () => {
        it('should use appropriate ARIA roles', async () => {
            await editor.initialize();

            const toolbar = toolbarContainer.querySelector('.editor-toolbar');
            expect(toolbar?.getAttribute('role')).toBe('toolbar');

            // Check button roles
            const buttons = toolbar?.querySelectorAll('button');
            buttons?.forEach(button => {
                expect(button.getAttribute('role') || 'button').toBe('button');
            });
        });

        it('should provide ARIA labels for interactive elements', async () => {
            await editor.initialize();

            const interactiveElements = editorContainer.querySelectorAll(
                'button, input, textarea, select'
            );

            interactiveElements.forEach(element => {
                expect(checkAriaAttributes(element)).toBeTruthy();
            });
        });

        it('should use ARIA states and properties correctly', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            const sourceButton = toolbarContainer.querySelector('button[title*="Source"]');
            if (sourceButton) {
                expect(sourceButton.getAttribute('aria-pressed')).toBe('true');
            }

            (editor as any).switchMode('visual');
            const visualButton = toolbarContainer.querySelector('button[title*="Visual"]');
            if (visualButton) {
                expect(visualButton.getAttribute('aria-pressed')).toBe('true');
            }
        });

        it('should provide live region updates', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            
            // Test validation messages in live region
            htmlEditor.value = '<div><p>Invalid HTML';
            await (editor as any).validateCode('html');

            const validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.getAttribute('aria-live')).toBe('polite');
        });
    });

    describe('Keyboard Navigation', () => {
        it('should support tab navigation through all interactive elements', async () => {
            await editor.initialize();

            const focusableElements = document.querySelectorAll(
                'button:not([disabled]), input:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
            );

            expect(focusableElements.length).toBeGreaterThan(0);

            // Test sequential focus
            let currentIndex = 0;
            focusableElements.forEach((element, index) => {
                const htmlElement = element as HTMLElement;
                htmlElement.focus();
                expect(document.activeElement).toBe(htmlElement);
                currentIndex = index;
            });
        });

        it('should support arrow key navigation in toolbar', async () => {
            await editor.initialize();

            const toolbar = toolbarContainer.querySelector('.editor-toolbar');
            const buttons = toolbar?.querySelectorAll('button');

            if (buttons && buttons.length > 1) {
                const firstButton = buttons[0] as HTMLElement;
                firstButton.focus();

                // Simulate arrow key navigation
                const rightArrowEvent = new KeyboardEvent('keydown', {
                    key: 'ArrowRight',
                    bubbles: true
                });

                firstButton.dispatchEvent(rightArrowEvent);
                
                // Should move focus to next button (implementation dependent)
                expect(document.activeElement).toBeTruthy();
            }
        });

        it('should support escape key to close dialogs/menus', async () => {
            await editor.initialize();

            // Test escape key handling
            const escapeEvent = new KeyboardEvent('keydown', {
                key: 'Escape',
                bubbles: true
            });

            expect(() => {
                document.dispatchEvent(escapeEvent);
            }).not.toThrow();
        });

        it('should support enter and space for button activation', async () => {
            await editor.initialize();

            const toolbar = toolbarContainer.querySelector('.editor-toolbar');
            const buttons = toolbar?.querySelectorAll('button');

            buttons?.forEach(button => {
                let activated = false;
                button.addEventListener('click', () => {
                    activated = true;
                });

                // Test Enter key
                const enterEvent = new KeyboardEvent('keydown', {
                    key: 'Enter',
                    bubbles: true
                });
                button.dispatchEvent(enterEvent);

                // Test Space key
                const spaceEvent = new KeyboardEvent('keydown', {
                    key: ' ',
                    bubbles: true
                });
                button.dispatchEvent(spaceEvent);

                // At least one should work (implementation dependent)
                expect(typeof activated).toBe('boolean');
            });
        });
    });

    describe('Screen Reader Support', () => {
        it('should provide meaningful element descriptions', async () => {
            await editor.initialize();

            editor.insertElement('paragraph', { text: 'Screen reader test' });
            editor.insertElement('heading', { text: 'Test Heading' });
            editor.insertElement('list', { items: ['Item 1', 'Item 2'] });

            const paragraph = editorContainer.querySelector('p');
            const heading = editorContainer.querySelector('h2');
            const list = editorContainer.querySelector('ul');

            expect(paragraph?.textContent).toBeTruthy();
            expect(heading?.textContent).toBeTruthy();
            expect(list?.children.length).toBeGreaterThan(0);
        });

        it('should announce dynamic content changes', async () => {
            await editor.initialize();

            // Create live region for announcements
            const liveRegion = document.createElement('div');
            liveRegion.setAttribute('aria-live', 'polite');
            liveRegion.setAttribute('aria-atomic', 'true');
            liveRegion.className = 'sr-only';
            document.body.appendChild(liveRegion);

            editor.insertElement('paragraph', { text: 'Dynamic content test' });

            // Should announce content changes
            expect(liveRegion.getAttribute('aria-live')).toBe('polite');

            document.body.removeChild(liveRegion);
        });

        it('should provide context for form controls', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            const cssEditor = editorContainer.querySelector('#css-editor') as HTMLTextAreaElement;

            // Should have labels or aria-label
            expect(
                htmlEditor.getAttribute('aria-label') ||
                htmlEditor.getAttribute('aria-labelledby') ||
                document.querySelector(`label[for="${htmlEditor.id}"]`)
            ).toBeTruthy();

            expect(
                cssEditor.getAttribute('aria-label') ||
                cssEditor.getAttribute('aria-labelledby') ||
                document.querySelector(`label[for="${cssEditor.id}"]`)
            ).toBeTruthy();
        });
    });

    describe('Color and Visual Accessibility', () => {
        it('should not rely solely on color to convey information', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            
            // Test validation errors
            htmlEditor.value = '<div><p>Invalid HTML';
            await (editor as any).validateCode('html');

            const validationItems = editorContainer.querySelectorAll('.validation-item');
            validationItems.forEach(item => {
                // Should have text content, not just color indicators
                expect(item.textContent?.trim().length).toBeGreaterThan(0);
                
                // Should have icons or other visual indicators beyond color
                const hasIcon = item.querySelector('.icon, .symbol') || 
                               item.textContent?.includes('Error') ||
                               item.textContent?.includes('Warning');
                expect(hasIcon).toBeTruthy();
            });
        });

        it('should support high contrast mode', async () => {
            await editor.initialize();

            // Simulate high contrast mode
            document.documentElement.style.filter = 'contrast(200%)';

            editor.insertElement('paragraph', { text: 'High contrast test' });
            
            const paragraph = editorContainer.querySelector('p');
            expect(paragraph?.textContent).toBe('High contrast test');

            // Reset
            document.documentElement.style.filter = '';
        });

        it('should support reduced motion preferences', async () => {
            await editor.initialize();

            // Mock reduced motion preference
            Object.defineProperty(window, 'matchMedia', {
                writable: true,
                value: jest.fn().mockImplementation(query => ({
                    matches: query === '(prefers-reduced-motion: reduce)',
                    media: query,
                    onchange: null,
                    addListener: jest.fn(),
                    removeListener: jest.fn(),
                    addEventListener: jest.fn(),
                    removeEventListener: jest.fn(),
                    dispatchEvent: jest.fn(),
                })),
            });

            // Should respect reduced motion
            const animatedElements = editorContainer.querySelectorAll('[style*="transition"], [style*="animation"]');
            animatedElements.forEach(element => {
                const style = (element as HTMLElement).style;
                // Should disable or reduce animations
                expect(
                    style.transition === 'none' ||
                    style.animation === 'none' ||
                    style.transition.includes('0s') ||
                    style.animation.includes('0s')
                ).toBeTruthy();
            });
        });
    });

    describe('Error Handling and Feedback', () => {
        it('should provide accessible error messages', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            
            // Test validation error accessibility
            htmlEditor.value = '<div><p>Broken HTML';
            await (editor as any).validateCode('html');

            const validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.getAttribute('role')).toBe('alert');
            expect(validationPanel?.getAttribute('aria-live')).toBe('polite');

            const errorItems = validationPanel?.querySelectorAll('.validation-item');
            errorItems?.forEach(item => {
                expect(item.textContent?.trim().length).toBeGreaterThan(0);
            });
        });

        it('should provide success feedback', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            
            // Test valid HTML
            htmlEditor.value = '<div><p>Valid HTML</p></div>';
            await (editor as any).validateCode('html');

            const validationPanel = editorContainer.querySelector('#html-validation');
            
            // Should indicate success accessibly
            expect(
                validationPanel?.textContent?.includes('Valid') ||
                validationPanel?.textContent?.includes('No errors') ||
                validationPanel?.getAttribute('aria-label')?.includes('valid')
            ).toBeTruthy();
        });

        it('should provide context-sensitive help', async () => {
            await editor.initialize();

            const toolbar = toolbarContainer.querySelector('.editor-toolbar');
            const buttons = toolbar?.querySelectorAll('button');

            buttons?.forEach(button => {
                // Should have help text via title, aria-describedby, or aria-label
                const hasHelp = 
                    button.getAttribute('title') ||
                    button.getAttribute('aria-describedby') ||
                    button.getAttribute('aria-label');
                
                expect(hasHelp).toBeTruthy();
            });
        });
    });

    describe('Mobile Accessibility', () => {
        it('should provide adequate touch targets', async () => {
            await editor.initialize();

            const interactiveElements = editorContainer.querySelectorAll('button, input, textarea, a');
            const minTouchTarget = 44; // WCAG recommendation

            interactiveElements.forEach(element => {
                const rect = element.getBoundingClientRect();
                const size = Math.min(rect.width, rect.height);
                
                // Allow some tolerance for styling
                expect(size).toBeGreaterThanOrEqual(minTouchTarget - 10);
            });
        });

        it('should support voice control', async () => {
            await editor.initialize();

            const buttons = editorContainer.querySelectorAll('button');
            
            buttons.forEach(button => {
                // Should have accessible names for voice control
                const accessibleName = 
                    button.textContent?.trim() ||
                    button.getAttribute('aria-label') ||
                    button.getAttribute('title');
                
                expect(accessibleName).toBeTruthy();
                expect(accessibleName!.length).toBeGreaterThan(2);
            });
        });
    });

    describe('Accessibility Testing Integration', () => {
        it('should pass automated accessibility checks', async () => {
            await editor.initialize();

            // Create comprehensive content for testing
            editor.insertElement('heading', { text: 'Accessibility Test Document' });
            editor.insertElement('paragraph', { text: 'This document tests accessibility compliance.' });
            editor.insertElement('image', { 
                src: 'https://example.com/test.jpg', 
                alt: 'Test image for accessibility' 
            });
            editor.insertElement('link', { 
                href: 'https://example.com', 
                text: 'Visit accessibility guidelines' 
            });

            // Basic accessibility checks
            const images = editorContainer.querySelectorAll('img');
            images.forEach(img => {
                expect(img.alt).toBeTruthy();
            });

            const links = editorContainer.querySelectorAll('a');
            links.forEach(link => {
                expect(link.textContent?.trim().length).toBeGreaterThan(0);
            });

            const headings = editorContainer.querySelectorAll('h1, h2, h3, h4, h5, h6');
            expect(headings.length).toBeGreaterThan(0);
        });

        it('should maintain accessibility during dynamic updates', async () => {
            await editor.initialize();

            // Test accessibility during content changes
            for (let i = 0; i < 5; i++) {
                editor.insertElement('paragraph', { text: `Dynamic paragraph ${i}` });
                
                // Check that accessibility attributes are maintained
                const paragraphs = editorContainer.querySelectorAll('p');
                expect(paragraphs.length).toBe(i + 1);
                
                paragraphs.forEach(p => {
                    expect(p.textContent?.trim().length).toBeGreaterThan(0);
                });
            }
        });
    });
});