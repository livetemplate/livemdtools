package livepage

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestNavigationSystem verifies that the navigation system works correctly
func TestNavigationSystem(t *testing.T) {
	// Start the server
	serverCmd := exec.Command("./livepage", "serve", "examples/counter", "--port", "8090")
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
	time.Sleep(3 * time.Second)

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

	var sidebarExists bool
	var navBottomExists bool
	var tocItemCount int
	var currentStepText string
	var totalStepsText string
	var prevBtnDisabled bool
	var nextBtnDisabled bool
	var html string

	err := chromedp.Run(ctx,
		// Navigate to page
		chromedp.Navigate("http://localhost:8090/"),
		chromedp.Sleep(2*time.Second), // Wait for page to load

		// Get the HTML for debugging
		chromedp.OuterHTML("html", &html),

		// 1. Verify sidebar exists
		chromedp.Evaluate(`document.querySelector('#tutorial-sidebar') !== null`, &sidebarExists),

		// 2. Verify bottom navigation exists
		chromedp.Evaluate(`document.querySelector('#tutorial-nav-bottom') !== null`, &navBottomExists),

		// 3. Count TOC items (should be 8 H2 sections in counter tutorial)
		chromedp.Evaluate(`document.querySelectorAll('.toc-item').length`, &tocItemCount),

		// 4. Check progress indicator
		chromedp.Text("#current-step", &currentStepText),
		chromedp.Text("#total-steps", &totalStepsText),

		// 5. Check prev button is disabled (we're on first section)
		chromedp.Evaluate(`document.querySelector('#nav-prev').disabled`, &prevBtnDisabled),

		// 6. Check next button is enabled
		chromedp.Evaluate(`document.querySelector('#nav-next').disabled`, &nextBtnDisabled),
	)

	if err != nil {
		t.Fatalf("Failed to run chromedp: %v", err)
	}

	// Save HTML for debugging
	if err := os.WriteFile("/tmp/navigation-test.html", []byte(html), 0644); err != nil {
		t.Logf("Warning: Could not save HTML: %v", err)
	}

	// Assertions
	t.Logf("Sidebar exists: %v", sidebarExists)
	t.Logf("Bottom nav exists: %v", navBottomExists)
	t.Logf("TOC items: %d", tocItemCount)
	t.Logf("Current step: %s", currentStepText)
	t.Logf("Total steps: %s", totalStepsText)
	t.Logf("Prev button disabled: %v", prevBtnDisabled)
	t.Logf("Next button disabled: %v", nextBtnDisabled)

	if !sidebarExists {
		t.Error("Sidebar does not exist")
	}

	if !navBottomExists {
		t.Error("Bottom navigation does not exist")
	}

	if tocItemCount != 8 {
		t.Errorf("Expected 8 TOC items, got %d", tocItemCount)
	}

	if currentStepText != "1" {
		t.Errorf("Expected current step to be 1, got %s", currentStepText)
	}

	if totalStepsText != "8" {
		t.Errorf("Expected total steps to be 8, got %s", totalStepsText)
	}

	if !prevBtnDisabled {
		t.Error("Prev button should be disabled on first section")
	}

	if nextBtnDisabled {
		t.Error("Next button should be enabled when not on last section")
	}

	// Test navigation interactions
	var currentStep2 int
	var prevBtnDisabled2 bool
	var firstTocActive bool
	var secondTocActive bool

	err = chromedp.Run(ctx,
		// Click next button
		chromedp.Click("#nav-next"),
		chromedp.Sleep(2*time.Second), // Wait for navigation and smooth scroll

		// Check that we moved to step 2
		chromedp.Evaluate(`parseInt(document.querySelector('#current-step').textContent)`, &currentStep2),

		// Check prev button is now enabled
		chromedp.Evaluate(`document.querySelector('#nav-prev').disabled`, &prevBtnDisabled2),

		// Check TOC active states
		chromedp.Evaluate(`document.querySelectorAll('.toc-link')[0].classList.contains('active')`, &firstTocActive),
		chromedp.Evaluate(`document.querySelectorAll('.toc-link')[1].classList.contains('active')`, &secondTocActive),
	)

	if err != nil {
		t.Fatalf("Failed to test navigation: %v", err)
	}

	t.Logf("After clicking next - Current step: %d", currentStep2)
	t.Logf("Prev button disabled: %v", prevBtnDisabled2)
	t.Logf("First TOC active: %v", firstTocActive)
	t.Logf("Second TOC active: %v", secondTocActive)

	if currentStep2 != 2 {
		t.Errorf("Expected step 2 after clicking next, got %d", currentStep2)
	}

	if prevBtnDisabled2 {
		t.Error("Prev button should be enabled after moving to step 2")
	}

	if firstTocActive {
		t.Error("First TOC item should not be active after moving to step 2")
	}

	if !secondTocActive {
		t.Error("Second TOC item should be active after moving to step 2")
	}

	// Test keyboard navigation
	var currentStep3 int

	err = chromedp.Run(ctx,
		// Press right arrow key
		chromedp.SendKeys("body", "\ue014"), // Right arrow key code
		chromedp.Sleep(1*time.Second),

		// Check we moved to step 3
		chromedp.Evaluate(`parseInt(document.querySelector('#current-step').textContent)`, &currentStep3),
	)

	if err != nil {
		t.Fatalf("Failed to test keyboard navigation: %v", err)
	}

	t.Logf("After right arrow - Current step: %d", currentStep3)

	if currentStep3 != 3 {
		t.Errorf("Expected step 3 after pressing right arrow, got %d", currentStep3)
	}

	// Test TOC click navigation
	var currentStep1Again int
	var urlHash string

	err = chromedp.Run(ctx,
		// Click on first TOC item
		chromedp.Click(".toc-link:first-child"),
		chromedp.Sleep(1*time.Second),

		// Check we're back at step 1
		chromedp.Evaluate(`parseInt(document.querySelector('#current-step').textContent)`, &currentStep1Again),

		// Check URL hash is updated
		chromedp.Evaluate(`window.location.hash`, &urlHash),
	)

	if err != nil {
		t.Fatalf("Failed to test TOC click: %v", err)
	}

	t.Logf("After TOC click - Current step: %d", currentStep1Again)
	t.Logf("URL hash: %s", urlHash)

	if currentStep1Again != 1 {
		t.Errorf("Expected step 1 after clicking first TOC item, got %d", currentStep1Again)
	}

	if urlHash == "" {
		t.Error("URL hash should be set after navigation")
	}

	// Test visited tracking
	var firstVisited bool
	var secondVisited bool
	var thirdVisited bool

	err = chromedp.Run(ctx,
		// Check visited states (we've been to steps 1, 2, 3)
		chromedp.Evaluate(`document.querySelectorAll('.toc-link')[0].classList.contains('visited')`, &firstVisited),
		chromedp.Evaluate(`document.querySelectorAll('.toc-link')[1].classList.contains('visited')`, &secondVisited),
		chromedp.Evaluate(`document.querySelectorAll('.toc-link')[2].classList.contains('visited')`, &thirdVisited),
	)

	if err != nil {
		t.Fatalf("Failed to test visited tracking: %v", err)
	}

	t.Logf("Visited states - 1:%v, 2:%v, 3:%v", firstVisited, secondVisited, thirdVisited)

	if !firstVisited {
		t.Error("First section should be marked as visited")
	}

	if !secondVisited {
		t.Error("Second section should be marked as visited")
	}

	if !thirdVisited {
		t.Error("Third section should be marked as visited")
	}

	// Test localStorage persistence
	var localStorageData string

	err = chromedp.Run(ctx,
		chromedp.Evaluate(`localStorage.getItem('livepage-tutorial-progress')`, &localStorageData),
	)

	if err != nil {
		t.Fatalf("Failed to check localStorage: %v", err)
	}

	t.Logf("localStorage data: %s", localStorageData)

	if localStorageData == "" || localStorageData == "null" {
		t.Error("Progress should be saved to localStorage")
	}

	t.Logf("✓ Navigation system working correctly!")
}
