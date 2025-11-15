/**
 * MonacoEditor - Wrapper for Monaco Editor
 */

import * as monaco from "monaco-editor";
import { EditorOptions } from "../types";

export class MonacoEditor {
  private editor: monaco.editor.IStandaloneCodeEditor | null = null;
  private container: HTMLElement;
  private options: EditorOptions;
  private onChangeCallback: ((code: string) => void) | null = null;

  constructor(container: HTMLElement, initialCode: string, options: EditorOptions) {
    this.container = container;
    this.options = options;

    this.initialize(initialCode);
  }

  private initialize(initialCode: string): void {
    // Create editor container
    const editorDiv = document.createElement("div");
    editorDiv.className = "livepage-monaco-editor";
    editorDiv.style.height = "300px"; // Default height
    editorDiv.style.width = "100%";
    this.container.appendChild(editorDiv);

    // Initialize Monaco Editor
    this.editor = monaco.editor.create(editorDiv, {
      value: initialCode,
      language: this.options.language,
      theme: this.options.theme || "vs-dark",
      readOnly: this.options.readonly,
      minimap: {
        enabled: this.options.minimap ?? true,
      },
      lineNumbers: this.options.lineNumbers !== false ? "on" : "off",
      scrollBeyondLastLine: false,
      automaticLayout: true,
      fontSize: 14,
      tabSize: 4,
      insertSpaces: false, // Use tabs
    });

    // Listen for content changes
    this.editor.onDidChangeModelContent(() => {
      if (this.onChangeCallback && this.editor) {
        this.onChangeCallback(this.editor.getValue());
      }
    });
  }

  /**
   * Get current code from editor
   */
  getValue(): string {
    return this.editor?.getValue() || "";
  }

  /**
   * Set code in editor
   */
  setValue(code: string): void {
    this.editor?.setValue(code);
  }

  /**
   * Set onChange callback
   */
  onChange(callback: (code: string) => void): void {
    this.onChangeCallback = callback;
  }

  /**
   * Set read-only mode
   */
  setReadOnly(readonly: boolean): void {
    this.editor?.updateOptions({ readOnly: readonly });
  }

  /**
   * Focus the editor
   */
  focus(): void {
    this.editor?.focus();
  }

  /**
   * Layout the editor (call after resize)
   */
  layout(): void {
    this.editor?.layout();
  }

  /**
   * Destroy the editor
   */
  destroy(): void {
    this.editor?.dispose();
    this.editor = null;
  }

  /**
   * Get the underlying Monaco editor instance
   */
  getEditor(): monaco.editor.IStandaloneCodeEditor | null {
    return this.editor;
  }
}

/**
 * Initialize Monaco environment (call once on page load)
 */
export function initializeMonaco(): void {
  // Set Monaco environment for webpack/bundlers
  if (typeof window !== "undefined") {
    (window as any).MonacoEnvironment = {
      getWorkerUrl: function (_moduleId: string, label: string) {
        if (label === "json") {
          return "/assets/monaco/json.worker.js";
        }
        if (label === "css" || label === "scss" || label === "less") {
          return "/assets/monaco/css.worker.js";
        }
        if (label === "html" || label === "handlebars" || label === "razor") {
          return "/assets/monaco/html.worker.js";
        }
        if (label === "typescript" || label === "javascript") {
          return "/assets/monaco/ts.worker.js";
        }
        return "/assets/monaco/editor.worker.js";
      },
    };
  }
}
