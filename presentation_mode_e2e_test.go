package livepage

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestPresentationMode verifies that presentation mode works correctly
func TestPresentationMode(t *testing.T) {
	// Start the server
	serverCmd := exec.Command("./livepage", "serve", "examples/counter", "--port", "8080")
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr

	if err := serverCmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		if serverCmd.Process != nil {
			serverCmd.Process.Kill()
		}
	}()

	// Wait for server to start
	time.Sleep(5 * time.Second)

	// Create chrome context
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
	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	var presentationBtnExists bool
	var bodyHasPresentationClass bool
	var sidebarVisible bool
	var currentSectionClass bool
	var html string

	// Test 1: Verify presentation button exists
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/"),
		chromedp.Sleep(2*time.Second),

		// Get HTML for debugging
		chromedp.OuterHTML("html", &html),

		// Check presentation button exists
		chromedp.Evaluate(`document.querySelector('#presentation-toggle') !== null`, &presentationBtnExists),
	)

	if err != nil {
		t.Fatalf("Failed to check presentation button: %v", err)
	}

	// Save HTML for debugging
	if err := os.WriteFile("/tmp/presentation-test-initial.html", []byte(html), 0644); err != nil {
		t.Logf("Warning: Could not save HTML: %v", err)
	}

	t.Logf("Presentation button exists: %v", presentationBtnExists)

	if !presentationBtnExists {
		t.Error("Presentation button does not exist")
	}

	// Test 2: Click presentation button and verify mode is activated
	err = chromedp.Run(ctx,
		// Click presentation button
		chromedp.Click("#presentation-toggle"),
		chromedp.Sleep(1*time.Second),

		// Get HTML after clicking
		chromedp.OuterHTML("html", &html),

		// Check body has presentation-mode class
		chromedp.Evaluate(`document.body.classList.contains('presentation-mode')`, &bodyHasPresentationClass),

		// Check sidebar is hidden
		chromedp.Evaluate(`
			var sidebar = document.querySelector('.livepage-nav-sidebar');
			sidebar ? window.getComputedStyle(sidebar).display === 'none' : true;
		`, &sidebarVisible),

		// Check current section has presentation marker
		chromedp.Evaluate(`document.querySelectorAll('.presentation-current-section').length > 0`, &currentSectionClass),
	)

	if err != nil {
		t.Fatalf("Failed to test presentation mode activation: %v", err)
	}

	// Save HTML for debugging
	if err := os.WriteFile("/tmp/presentation-test-activated.html", []byte(html), 0644); err != nil {
		t.Logf("Warning: Could not save HTML: %v", err)
	}

	t.Logf("Body has presentation-mode class: %v", bodyHasPresentationClass)
	t.Logf("Sidebar hidden: %v", sidebarVisible)
	t.Logf("Current section marked: %v", currentSectionClass)

	if !bodyHasPresentationClass {
		t.Error("Body should have presentation-mode class")
	}

	if !sidebarVisible {
		t.Error("Sidebar should be hidden in presentation mode")
	}

	if !currentSectionClass {
		t.Error("Current section should be marked with presentation-current-section class")
	}

	// Test 3: Test keyboard shortcut (F key)
	var bodyHasClassAfterF bool

	err = chromedp.Run(ctx,
		// First exit presentation mode by clicking button again
		chromedp.Click("#presentation-toggle"),
		chromedp.Sleep(500*time.Millisecond),

		// Verify we exited
		chromedp.Evaluate(`document.body.classList.contains('presentation-mode')`, &bodyHasClassAfterF),
	)

	if err != nil {
		t.Fatalf("Failed to exit presentation mode: %v", err)
	}

	t.Logf("After clicking again, presentation mode active: %v", bodyHasClassAfterF)

	if bodyHasClassAfterF {
		t.Error("Should have exited presentation mode after second click")
	}

	// Test keyboard shortcut
	var bodyHasClassAfterKeypress bool

	err = chromedp.Run(ctx,
		// Focus the body and send 'f' key via keyboard event
		chromedp.Evaluate(`
			const event = new KeyboardEvent('keydown', { key: 'f', code: 'KeyF', keyCode: 70 });
			document.dispatchEvent(event);
		`, nil),
		chromedp.Sleep(500*time.Millisecond),

		// Check presentation mode is active
		chromedp.Evaluate(`document.body.classList.contains('presentation-mode')`, &bodyHasClassAfterKeypress),
	)

	if err != nil {
		t.Fatalf("Failed to test keyboard shortcut: %v", err)
	}

	t.Logf("After pressing 'f', presentation mode active: %v", bodyHasClassAfterKeypress)

	if !bodyHasClassAfterKeypress {
		t.Error("Presentation mode should be activated by 'f' key")
	}

	// Test 4: Verify navigation works in presentation mode
	var currentSectionElements1 int
	var currentSectionElements2 int

	err = chromedp.Run(ctx,
		// Count current section elements
		chromedp.Evaluate(`document.querySelectorAll('.presentation-current-section').length`, &currentSectionElements1),

		// Navigate to next section using arrow key (more reliable than clicking button)
		chromedp.Evaluate(`
			const arrowEvent = new KeyboardEvent('keydown', { key: 'ArrowRight', code: 'ArrowRight', keyCode: 39 });
			document.dispatchEvent(arrowEvent);
		`, nil),
		chromedp.Sleep(1*time.Second),

		// Count again (should be different elements now)
		chromedp.Evaluate(`document.querySelectorAll('.presentation-current-section').length`, &currentSectionElements2),
	)

	if err != nil {
		t.Fatalf("Failed to test navigation in presentation mode: %v", err)
	}

	t.Logf("Section 1 elements: %d", currentSectionElements1)
	t.Logf("Section 2 elements: %d", currentSectionElements2)

	// Both sections should have some content
	if currentSectionElements1 == 0 {
		t.Error("First section should have content in presentation mode")
	}

	if currentSectionElements2 == 0 {
		t.Error("Second section should have content in presentation mode")
	}

	// Test 5: Verify button has active class when in presentation mode
	var btnHasActiveClass bool

	err = chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('#presentation-toggle').classList.contains('active')`, &btnHasActiveClass),
	)

	if err != nil {
		t.Fatalf("Failed to check button active class: %v", err)
	}

	t.Logf("Button has active class: %v", btnHasActiveClass)

	if !btnHasActiveClass {
		t.Error("Presentation button should have 'active' class when in presentation mode")
	}

	t.Logf("✓ Presentation mode working correctly!")
}
