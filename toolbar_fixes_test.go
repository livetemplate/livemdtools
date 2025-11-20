package livepage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestToolbarFixes verifies toolbar consolidation and fixes
func TestToolbarFixes(t *testing.T) {
	// Create chrome context with console logging
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var html string

	t.Log("Test 1: Verify toolbar is at the top and sticky")
	var toolbarTop float64
	var toolbarPosition string
	var toolbarZIndex string

	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/"),
		chromedp.Sleep(2*time.Second),

		// Get HTML for debugging
		chromedp.OuterHTML("html", &html),

		// Check toolbar position
		chromedp.Evaluate(`window.getComputedStyle(document.querySelector('.tutorial-nav-bottom')).position`, &toolbarPosition),
		chromedp.Evaluate(`window.getComputedStyle(document.querySelector('.tutorial-nav-bottom')).zIndex`, &toolbarZIndex),
		chromedp.Evaluate(`document.querySelector('.tutorial-nav-bottom').getBoundingClientRect().top`, &toolbarTop),
	)

	if err != nil {
		t.Fatalf("Failed to check toolbar position: %v", err)
	}

	// Save HTML for debugging
	if err := os.WriteFile("/tmp/toolbar-test-initial.html", []byte(html), 0644); err != nil {
		t.Logf("Warning: Could not save HTML: %v", err)
	}

	t.Logf("Toolbar position: %s (expected: sticky)", toolbarPosition)
	t.Logf("Toolbar z-index: %s (expected: 100)", toolbarZIndex)
	t.Logf("Toolbar top: %.2f (expected: 0)", toolbarTop)

	if toolbarPosition != "sticky" {
		t.Error("Toolbar should have position: sticky")
	}

	if toolbarZIndex != "100" {
		t.Error("Toolbar should have z-index: 100")
	}

	if toolbarTop != 0 {
		t.Errorf("Toolbar should be at top (0), got %.2f", toolbarTop)
	}

	t.Log("Test 2: Verify progress text doesn't wrap")
	var progressTextWrapped bool
	var progressHeight float64

	err = chromedp.Run(ctx,
		// Check if progress text height is single line (< 30px typically)
		chromedp.Evaluate(`document.querySelector('.progress-text').getBoundingClientRect().height`, &progressHeight),
		chromedp.Evaluate(`document.querySelector('.progress-text').scrollHeight > document.querySelector('.progress-text').clientHeight`, &progressTextWrapped),
	)

	if err != nil {
		t.Fatalf("Failed to check progress text: %v", err)
	}

	t.Logf("Progress text height: %.2f (should be single line, < 30px)", progressHeight)
	t.Logf("Progress text wrapped: %v (should be false)", progressTextWrapped)

	if progressTextWrapped {
		t.Error("Progress text should not wrap to multiple lines")
	}

	if progressHeight > 30 {
		t.Errorf("Progress text height too large (%.2f), suggests wrapping", progressHeight)
	}

	t.Log("Test 3: Verify theme and presentation buttons are in toolbar")
	var hasThemeLight bool
	var hasThemeDark bool
	var hasThemeAuto bool
	var hasPresentationBtn bool
	var topThemeToggleHidden bool

	err = chromedp.Run(ctx,
		// Check theme buttons exist in toolbar
		chromedp.Evaluate(`document.querySelector('.tutorial-nav-bottom #theme-light') !== null`, &hasThemeLight),
		chromedp.Evaluate(`document.querySelector('.tutorial-nav-bottom #theme-dark') !== null`, &hasThemeDark),
		chromedp.Evaluate(`document.querySelector('.tutorial-nav-bottom #theme-auto') !== null`, &hasThemeAuto),
		chromedp.Evaluate(`document.querySelector('.tutorial-nav-bottom #presentation-toggle') !== null`, &hasPresentationBtn),

		// Check top theme toggle is hidden
		chromedp.Evaluate(`window.getComputedStyle(document.querySelector('.theme-toggle')).display === 'none'`, &topThemeToggleHidden),
	)

	if err != nil {
		t.Fatalf("Failed to check theme buttons: %v", err)
	}

	t.Logf("Has theme light button: %v", hasThemeLight)
	t.Logf("Has theme dark button: %v", hasThemeDark)
	t.Logf("Has theme auto button: %v", hasThemeAuto)
	t.Logf("Has presentation button: %v", hasPresentationBtn)
	t.Logf("Top theme toggle hidden: %v", topThemeToggleHidden)

	if !hasThemeLight || !hasThemeDark || !hasThemeAuto {
		t.Error("All theme buttons should be in the toolbar")
	}

	if !hasPresentationBtn {
		t.Error("Presentation button should be in the toolbar")
	}

	if !topThemeToggleHidden {
		t.Error("Original top theme toggle should be hidden")
	}

	t.Log("Test 4: Navigate to step 7 and verify TOC highlighting")
	var currentStep string
	var tocItems []map[string]interface{}

	err = chromedp.Run(ctx,
		// Navigate through steps to reach step 7 (beyond "How It Works")
		chromedp.Click("#nav-next"),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click("#nav-next"),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click("#nav-next"),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click("#nav-next"),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click("#nav-next"),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click("#nav-next"),
		chromedp.Sleep(500*time.Millisecond),

		// Get current step
		chromedp.Text("#current-step", &currentStep),

		// Get all TOC items with their classes and text
		chromedp.Evaluate(`Array.from(document.querySelectorAll('.toc-link')).map(el => ({
			text: el.textContent.trim(),
			classList: Array.from(el.classList),
			href: el.getAttribute('href'),
			visited: el.classList.contains('visited'),
			active: el.classList.contains('active')
		}))`, &tocItems),

		// Get HTML after navigation
		chromedp.OuterHTML("html", &html),
	)

	if err != nil {
		t.Fatalf("Failed to navigate and check TOC: %v", err)
	}

	// Save HTML for debugging
	if err := os.WriteFile("/tmp/toolbar-test-step7.html", []byte(html), 0644); err != nil {
		t.Logf("Warning: Could not save HTML: %v", err)
	}

	t.Logf("Current step: %s (expected: 7)", currentStep)
	t.Logf("TOC items found: %d", len(tocItems))

	for i, item := range tocItems {
		t.Logf("  TOC[%d]: %s - visited=%v, active=%v, classes=%v",
			i, item["text"], item["visited"], item["active"], item["classList"])
	}

	if currentStep != "7" {
		t.Errorf("Expected to be at step 7, got step %s", currentStep)
	}

	// Check if any TOC items beyond "How It Works" (index 5) are marked as visited/active
	hasVisitedBeyondHowItWorks := false
	for i, item := range tocItems {
		visited := item["visited"].(bool)
		active := item["active"].(bool)

		// Check items after "How It Works" (index 5)
		if i > 5 && (visited || active) {
			hasVisitedBeyondHowItWorks = true
			break
		}
	}

	t.Logf("Has visited/active items beyond 'How It Works': %v", hasVisitedBeyondHowItWorks)

	if !hasVisitedBeyondHowItWorks {
		t.Error("TOC highlighting should work beyond 'How It Works' section")
	}

	t.Log("Test 5: Check for mermaid diagram errors")
	var mermaidErrors []string

	err = chromedp.Run(ctx,
		// Navigate back to where the mermaid diagram is (step 6 - "How It Works")
		chromedp.Click("#nav-prev"),
		chromedp.Sleep(1*time.Second),

		// Get HTML
		chromedp.OuterHTML("html", &html),

		// Look for mermaid error messages (specifically for "Syntax error in text")
		chromedp.Evaluate(`Array.from(document.querySelectorAll('.mermaid')).map(el => el.textContent.trim()).filter(t => t.includes('Syntax error'))`, &mermaidErrors),
	)

	if err != nil {
		t.Fatalf("Failed to check mermaid errors: %v", err)
	}

	// Save HTML for debugging
	if err := os.WriteFile("/tmp/toolbar-test-mermaid.html", []byte(html), 0644); err != nil {
		t.Logf("Warning: Could not save HTML: %v", err)
	}

	t.Logf("Mermaid errors found: %d", len(mermaidErrors))
	for i, err := range mermaidErrors {
		t.Logf("  Mermaid error[%d]: %s", i, err)
	}

	if len(mermaidErrors) > 0 {
		t.Errorf("Mermaid diagram has errors: %v", mermaidErrors)
	}

	t.Log("✓ Toolbar fixes test complete!")
}
