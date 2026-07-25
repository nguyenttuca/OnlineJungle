/**
 * Editor.js - Monaco Editor Integration for Top OJ
 */

document.addEventListener('DOMContentLoaded', () => {
    // Only initialize if we have an editor container
    const editorContainer = document.getElementById('monaco-editor-container');
    const sourceCodeTextarea = document.getElementById('source_code');
    const languageSelect = document.getElementById('language');
    
    if (!editorContainer || !sourceCodeTextarea) return;

    // We will use requireJS provided by Monaco loader
    require.config({ paths: { 'vs': 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.45.0/min/vs' }});

    require(['vs/editor/editor.main'], function() {
        // Map OJ languages to Monaco languages
        const getMonacoLanguage = (ojLang) => {
            switch(ojLang) {
                case 'cpp': return 'cpp';
                case 'c': return 'c';
                case 'python': return 'python';
                case 'java': return 'java';
                case 'pascal': return 'pascal';
                default: return 'plaintext';
            }
        };

        // Determine theme based on html data-bs-theme
        const getMonacoTheme = () => {
            return document.documentElement.getAttribute('data-bs-theme') === 'dark' ? 'vs-dark' : 'vs';
        };

        // Initialize editor
        const editor = monaco.editor.create(editorContainer, {
            value: sourceCodeTextarea.value,
            language: getMonacoLanguage(languageSelect ? languageSelect.value : 'cpp'),
            theme: getMonacoTheme(),
            automaticLayout: true,
            fontSize: 14,
            fontFamily: "'JetBrains Mono', monospace",
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            roundedSelection: true,
            padding: { top: 16, bottom: 16 },
            renderLineHighlight: 'all',
            cursorBlinking: 'smooth',
            cursorSmoothCaretAnimation: true,
            formatOnPaste: true,
        });

        // Sync editor content to hidden textarea
        editor.onDidChangeModelContent(() => {
            sourceCodeTextarea.value = editor.getValue();
            // Save draft to localStorage based on problem ID/Slug if available
            const form = sourceCodeTextarea.closest('form');
            if (form) {
                const action = form.getAttribute('action');
                if (action) {
                    localStorage.setItem(`draft_${action}`, editor.getValue());
                }
            }
        });

        // Load draft if editor is empty
        if (!editor.getValue()) {
            const form = sourceCodeTextarea.closest('form');
            if (form) {
                const action = form.getAttribute('action');
                if (action) {
                    const draft = localStorage.getItem(`draft_${action}`);
                    if (draft) {
                        editor.setValue(draft);
                    }
                }
            }
        }

        // Handle language change
        if (languageSelect) {
            languageSelect.addEventListener('change', (e) => {
                monaco.editor.setModelLanguage(editor.getModel(), getMonacoLanguage(e.target.value));
            });
        }

        // Handle theme toggle if it exists
        const themeToggleBtn = document.getElementById('themeToggleBtn');
        if (themeToggleBtn) {
            themeToggleBtn.addEventListener('click', () => {
                // Short delay to let the HTML attribute update first
                setTimeout(() => {
                    monaco.editor.setTheme(getMonacoTheme());
                }, 10);
            });
        }

        // Handle Ctrl+Enter to submit
        editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, function() {
            const form = sourceCodeTextarea.closest('form');
            if (form) {
                // Show loading state on submit button if it exists
                const submitBtn = form.querySelector('button[type="submit"]');
                if (submitBtn) {
                    const originalText = submitBtn.innerHTML;
                    submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Submitting...';
                    submitBtn.disabled = true;
                }
                form.submit();
            }
        });
        
        // Export to window for external access if needed
        window.ojEditor = editor;
    });
});

/**
 * Utility to make read-only editors (for submission details)
 */
function initReadOnlyEditor(containerId, sourceCodeId, language) {
    const container = document.getElementById(containerId);
    const sourceEl = document.getElementById(sourceCodeId);
    
    if (!container || !sourceEl) return;
    
    require.config({ paths: { 'vs': 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.45.0/min/vs' }});
    
    require(['vs/editor/editor.main'], function() {
        const getMonacoLanguage = (ojLang) => {
            switch(ojLang) {
                case 'cpp': return 'cpp';
                case 'c': return 'c';
                case 'python': return 'python';
                case 'java': return 'java';
                case 'pascal': return 'pascal';
                default: return 'plaintext';
            }
        };

        const getMonacoTheme = () => {
            return document.documentElement.getAttribute('data-bs-theme') === 'dark' ? 'vs-dark' : 'vs';
        };

        const editor = monaco.editor.create(container, {
            value: sourceEl.textContent || sourceEl.value,
            language: getMonacoLanguage(language),
            theme: getMonacoTheme(),
            automaticLayout: true,
            readOnly: true,
            fontSize: 14,
            fontFamily: "'JetBrains Mono', monospace",
            minimap: { enabled: true },
            scrollBeyondLastLine: false,
        });

        // Handle theme toggle
        const themeToggleBtn = document.getElementById('themeToggleBtn');
        if (themeToggleBtn) {
            themeToggleBtn.addEventListener('click', () => {
                setTimeout(() => {
                    monaco.editor.setTheme(getMonacoTheme());
                }, 10);
            });
        }
    });
}
