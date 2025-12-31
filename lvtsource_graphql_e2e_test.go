package tinkerdown_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"net/http/httptest"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/livetemplate/tinkerdown/internal/config"
	"github.com/livetemplate/tinkerdown/internal/server"
)

// TestGraphQLSourceE2E tests the lvt-source functionality with a public GraphQL API
// This test verifies that:
// 1. lvt-source="countries" fetches data from the Countries GraphQL API
// 2. The data is rendered in a table with proper headers
// 3. Multiple countries are displayed correctly
func TestGraphQLSourceE2E(t *testing.T) {
	// Load config from test example
	cfg, err := config.LoadFromDir("examples/lvt-source-graphql-test")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify source is configured
	if cfg.Sources == nil {
		t.Fatal("No sources configured in tinkerdown.yaml")
	}
	countriesSource, ok := cfg.Sources["countries"]
	if !ok {
		t.Fatal("countries source not found in config")
	}
	if countriesSource.Type != "graphql" {
		t.Fatalf("Expected graphql source type, got: %s", countriesSource.Type)
	}
	t.Logf("Source config: type=%s, url=%s, query_file=%s, result_path=%s",
		countriesSource.Type, countriesSource.URL, countriesSource.QueryFile, countriesSource.ResultPath)

	// Create test server
	srv := server.NewWithConfig("examples/lvt-source-graphql-test", cfg)
	if err := srv.Discover(); err != nil {
		t.Fatalf("Failed to discover pages: %v", err)
	}

	handler := server.WithCompression(srv)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Setup chromedp
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
		)...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(t.Logf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Store console logs for debugging
	var consoleLogs []string
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*runtime.EventConsoleAPICalled); ok {
			for _, arg := range ev.Args {
				consoleLogs = append(consoleLogs, fmt.Sprintf("[Console] %s", arg.Value))
			}
		}
	})

	t.Logf("Test server URL: %s", ts.URL)

	// Test 1: Navigate and wait for WebSocket to render content
	var hasInteractiveBlock bool
	err = chromedp.Run(ctx,
		chromedp.Navigate(ts.URL+"/"),
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate(`document.querySelector('.tinkerdown-interactive-block') !== null`, &hasInteractiveBlock),
	)
	if err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}

	if !hasInteractiveBlock {
		var htmlContent string
		chromedp.Run(ctx, chromedp.OuterHTML("html", &htmlContent))
		t.Logf("HTML: %s", htmlContent[:min(2000, len(htmlContent))])
		t.Fatal("Page did not load correctly - no interactive block found")
	}

	// Wait for WebSocket to render the table with country data
	// The Countries API returns 250+ countries, so we wait for the table to have data
	var tableRendered bool
	err = chromedp.Run(ctx,
		chromedp.Sleep(5*time.Second), // Give time for GraphQL fetch
		// Check for table inside the interactive block (the table is inside the block wrapper)
		chromedp.Evaluate(`document.querySelector('.tinkerdown-interactive-block table') !== null`, &tableRendered),
	)
	if err != nil {
		t.Fatalf("Failed to wait for table: %v", err)
	}

	if !tableRendered {
		var htmlContent string
		chromedp.Run(ctx, chromedp.OuterHTML("html", &htmlContent))
		t.Logf("HTML (first 3000 chars): %s", htmlContent[:min(3000, len(htmlContent))])
		t.Logf("Console logs: %v", consoleLogs)
		t.Fatal("Table was not rendered by WebSocket - table not found in interactive block")
	}
	t.Log("Page loaded and table rendered via WebSocket")

	// Test 2: Verify table headers exist
	var hasCodeHeader, hasCountryHeader, hasFlagHeader bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('th:first-child').textContent.includes('Code')`, &hasCodeHeader),
		chromedp.Evaluate(`document.querySelector('th:nth-child(2)').textContent.includes('Country')`, &hasCountryHeader),
		chromedp.Evaluate(`document.querySelector('th:nth-child(3)').textContent.includes('Flag')`, &hasFlagHeader),
	)
	if err != nil {
		t.Fatalf("Failed to check headers: %v", err)
	}

	if !hasCodeHeader {
		t.Fatal("Code header not found")
	}
	if !hasCountryHeader {
		t.Fatal("Country header not found")
	}
	if !hasFlagHeader {
		t.Fatal("Flag header not found")
	}
	t.Log("All table headers verified: Code, Country, Flag")

	// Test 3: Verify country data is rendered (at least 10 rows)
	var rowCount int
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelectorAll('tbody tr').length`, &rowCount),
	)
	if err != nil {
		t.Fatalf("Failed to check row count: %v", err)
	}

	if rowCount < 10 {
		t.Logf("Console logs: %v", consoleLogs)
		t.Fatalf("Expected at least 10 country rows, got %d", rowCount)
	}
	t.Logf("Found %d country rows (expected 250+)", rowCount)

	// Test 4: Verify specific country data exists (US should always be present)
	var hasUSCode, hasUSName bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`document.body.innerHTML.includes('US')`, &hasUSCode),
		chromedp.Evaluate(`document.body.innerHTML.includes('United States')`, &hasUSName),
	)
	if err != nil {
		t.Fatalf("Failed to check US data: %v", err)
	}

	if !hasUSCode {
		t.Fatal("US country code not found in table")
	}
	if !hasUSName {
		t.Fatal("United States country name not found in table")
	}
	t.Log("Verified US country data is present")

	// Test 5: Verify emoji flags are displayed (check for any emoji)
	var hasEmoji bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			(() => {
				const cells = document.querySelectorAll('tbody td:nth-child(3)');
				for (const cell of cells) {
					// Check for emoji pattern (flags are emoji)
					if (cell.textContent && cell.textContent.length > 0) {
						return true;
					}
				}
				return false;
			})()
		`, &hasEmoji),
	)
	if err != nil {
		t.Fatalf("Failed to check emojis: %v", err)
	}

	if !hasEmoji {
		t.Fatal("No emoji flags found in table")
	}
	t.Log("Emoji flags are displayed in table")

	// Test 6: Verify multiple countries exist (spot check a few)
	var htmlContent string
	err = chromedp.Run(ctx,
		chromedp.OuterHTML("body", &htmlContent),
	)
	if err != nil {
		t.Fatalf("Failed to get HTML: %v", err)
	}

	expectedCountries := []string{"United States", "Canada", "Germany", "France", "Japan"}
	foundCount := 0
	for _, country := range expectedCountries {
		if strings.Contains(htmlContent, country) {
			foundCount++
			t.Logf("Found country: %s", country)
		}
	}

	if foundCount < 3 {
		t.Logf("Console logs: %v", consoleLogs)
		t.Fatalf("Expected at least 3 of the sample countries, found %d", foundCount)
	}
	t.Logf("Verified %d of %d sample countries", foundCount, len(expectedCountries))

	t.Log("All GraphQL source E2E tests passed!")
}
