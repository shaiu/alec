package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/your-org/alec/pkg/contracts"
	"github.com/your-org/alec/pkg/services"
)

// TestSearchAndFilteringFunctionality tests the complete search and filtering workflow
// This integration test verifies that the ScriptDiscovery service can properly
// filter scripts by various criteria including name, type, tags, and content
func TestSearchAndFilteringFunctionality(t *testing.T) {
	// Create temporary directory structure for testing
	testDir := t.TempDir()
	scriptsDir := filepath.Join(testDir, "scripts")
	webDir := filepath.Join(scriptsDir, "web")
	dbDir := filepath.Join(scriptsDir, "database")
	deployDir := filepath.Join(scriptsDir, "deploy")

	// Create directory structure
	for _, dir := range []string{webDir, dbDir, deployDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create test scripts with various names, types, and content
	testScripts := map[string]string{
		// Web scripts
		filepath.Join(webDir, "start_server.sh"):     "#!/bin/bash\n# Start web server\nnginx -g 'daemon off;'",
		filepath.Join(webDir, "deploy_web.py"):       "#!/usr/bin/env python3\n# Deploy web application\nprint('Deploying web app')",
		filepath.Join(webDir, "build_assets.js"):     "#!/usr/bin/env node\n// Build frontend assets\nconsole.log('Building assets');",
		filepath.Join(webDir, "restart_nginx.sh"):    "#!/bin/bash\n# Restart nginx server\nsudo systemctl restart nginx",

		// Database scripts
		filepath.Join(dbDir, "backup_db.sh"):         "#!/bin/bash\n# Backup database\nmysqldump -u root mydb > backup.sql",
		filepath.Join(dbDir, "migrate_schema.py"):    "#!/usr/bin/env python3\n# Database migration\nfrom sqlalchemy import *",
		filepath.Join(dbDir, "restore_backup.rb"):    "#!/usr/bin/env ruby\n# Restore database backup\nputs 'Restoring database'",
		filepath.Join(dbDir, "cleanup_logs.pl"):      "#!/usr/bin/perl\n# Clean up database logs\nunlink glob '/var/log/mysql/*.log';",

		// Deploy scripts
		filepath.Join(deployDir, "deploy_prod.sh"):   "#!/bin/bash\n# Production deployment\necho 'Deploying to production'",
		filepath.Join(deployDir, "rollback.py"):      "#!/usr/bin/env python3\n# Rollback deployment\nprint('Rolling back')",
		filepath.Join(deployDir, "health_check.js"):  "#!/usr/bin/env node\n// Health check for deployment\nconsole.log('Health OK');",

		// Root level scripts
		filepath.Join(scriptsDir, "setup.sh"):        "#!/bin/bash\n# Initial setup script\napt-get update && apt-get install -y git",
		filepath.Join(scriptsDir, "test_all.py"):     "#!/usr/bin/env python3\n# Run all tests\nimport subprocess\nsubprocess.run(['pytest'])",
		filepath.Join(scriptsDir, "backup_all.rb"):   "#!/usr/bin/env ruby\n# Backup everything\n['db', 'files'].each { |type| backup(type) }",
	}

	// Create all test scripts
	for path, content := range testScripts {
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			t.Fatalf("Failed to create test script %s: %v", path, err)
		}
	}

	// Initialize service registry
	registry, err := services.NewServiceRegistry()
	if err != nil {
		t.Fatalf("Failed to create service registry: %v", err)
	}

	discovery := registry.GetScriptDiscovery()
	ctx := context.Background()

	// Scan all scripts first
	results, err := discovery.ScanDirectories(ctx, []string{scriptsDir})
	if err != nil {
		t.Fatalf("Failed to scan directories: %v", err)
	}

	// Collect all scripts for filtering tests
	var allScripts []contracts.ScriptInfo
	for _, dir := range results {
		allScripts = append(allScripts, dir.Scripts...)
	}

	if len(allScripts) == 0 {
		t.Fatal("No scripts found during scan")
	}

	t.Run("Filter by Script Name", func(t *testing.T) {
		testCases := []struct {
			query       string
			expectCount int
			description string
		}{
			{"backup", 3, "scripts with 'backup' in name"},
			{"deploy", 3, "scripts with 'deploy' in name"},
			{"server", 1, "scripts with 'server' in name"},
			{"test", 1, "scripts with 'test' in name"},
			{"nonexistent", 0, "scripts with non-existent name"},
			{"", len(allScripts), "all scripts with empty query"},
		}

		for _, tc := range testCases {
			t.Run("name_"+tc.query, func(t *testing.T) {
				filtered := discovery.FilterScripts(allScripts, tc.query)

				if len(filtered) != tc.expectCount {
					t.Errorf("FilterScripts(%q) = %d results, want %d (%s)",
						tc.query, len(filtered), tc.expectCount, tc.description)
				}

				// Verify all results contain the query string in name
				if tc.query != "" {
					for _, script := range filtered {
						if !strings.Contains(strings.ToLower(script.Name), strings.ToLower(tc.query)) {
							t.Errorf("Script %s does not contain query %q in name", script.Name, tc.query)
						}
					}
				}
			})
		}
	})

	t.Run("Filter by Script Type", func(t *testing.T) {
		testCases := []struct {
			query       string
			expectCount int
			scriptType  string
		}{
			{"shell", 6, "shell"},  // .sh scripts
			{"python", 4, "python"}, // .py scripts
			{"node", 2, "node"},     // .js scripts
			{"ruby", 2, "ruby"},     // .rb scripts
			{"perl", 1, "perl"},     // .pl scripts
		}

		for _, tc := range testCases {
			t.Run("type_"+tc.query, func(t *testing.T) {
				filtered := discovery.FilterScripts(allScripts, tc.query)

				if len(filtered) != tc.expectCount {
					t.Errorf("FilterScripts(%q) = %d results, want %d (%s scripts)",
						tc.query, len(filtered), tc.expectCount, tc.scriptType)
				}

				// Verify all results have the expected type
				for _, script := range filtered {
					if script.Type != tc.scriptType {
						t.Errorf("Script %s has type %s, expected type %s", script.Name, script.Type, tc.scriptType)
					}
				}
			})
		}
	})

	t.Run("Case Insensitive Filtering", func(t *testing.T) {
		testCases := []struct {
			lower string
			upper string
			mixed string
		}{
			{"backup", "BACKUP", "BaCkUp"},
			{"deploy", "DEPLOY", "DePlOy"},
			{"shell", "SHELL", "ShElL"},
		}

		for _, tc := range testCases {
			lowerResults := discovery.FilterScripts(allScripts, tc.lower)
			upperResults := discovery.FilterScripts(allScripts, tc.upper)
			mixedResults := discovery.FilterScripts(allScripts, tc.mixed)

			if len(lowerResults) != len(upperResults) || len(lowerResults) != len(mixedResults) {
				t.Errorf("Case insensitive filtering failed for %q: lower=%d, upper=%d, mixed=%d",
					tc.lower, len(lowerResults), len(upperResults), len(mixedResults))
			}

			// Verify same scripts are returned
			for i, script := range lowerResults {
				if i < len(upperResults) && script.Path != upperResults[i].Path {
					t.Errorf("Case insensitive filtering returned different scripts")
				}
				if i < len(mixedResults) && script.Path != mixedResults[i].Path {
					t.Errorf("Case insensitive filtering returned different scripts")
				}
			}
		}
	})

	t.Run("Filter by File Extension", func(t *testing.T) {
		testCases := []struct {
			extension   string
			expectCount int
		}{
			{".sh", 6},   // shell scripts
			{".py", 4},   // python scripts
			{".js", 2},   // javascript scripts
			{".rb", 2},   // ruby scripts
			{".pl", 1},   // perl scripts
			{".xyz", 0},  // non-existent extension
		}

		for _, tc := range testCases {
			t.Run("ext_"+tc.extension, func(t *testing.T) {
				// Filter by extension (which should match type in most cases)
				query := tc.extension[1:] // Remove the dot
				filtered := discovery.FilterScripts(allScripts, query)

				// Count actual scripts with this extension
				actualCount := 0
				for _, script := range allScripts {
					if strings.HasSuffix(script.Path, tc.extension) {
						actualCount++
					}
				}

				// The filter might match by type name rather than extension
				// So we check that the count is reasonable
				if tc.extension != ".xyz" && len(filtered) == 0 && actualCount > 0 {
					t.Errorf("FilterScripts for extension %s returned 0 results but %d scripts exist",
						tc.extension, actualCount)
				}
			})
		}
	})

	t.Run("Complex Query Scenarios", func(t *testing.T) {
		testCases := []struct {
			name  string
			query string
			check func([]contracts.ScriptInfo) bool
		}{
			{
				name:  "web scripts",
				query: "web",
				check: func(scripts []contracts.ScriptInfo) bool {
					// Should find scripts in web directory or with web-related names
					return len(scripts) > 0
				},
			},
			{
				name:  "database scripts",
				query: "db",
				check: func(scripts []contracts.ScriptInfo) bool {
					// Should find scripts in db directory or with db-related names
					return len(scripts) > 0
				},
			},
			{
				name:  "deployment scripts",
				query: "deploy",
				check: func(scripts []contracts.ScriptInfo) bool {
					// Should find deployment-related scripts
					return len(scripts) >= 2 // deploy_web.py and deploy_prod.sh
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				filtered := discovery.FilterScripts(allScripts, tc.query)

				if !tc.check(filtered) {
					t.Errorf("Complex query %q did not pass validation check", tc.query)
				}

				// Verify all results are relevant to the query
				for _, script := range filtered {
					scriptRelevant := strings.Contains(strings.ToLower(script.Name), strings.ToLower(tc.query)) ||
						strings.Contains(strings.ToLower(script.Path), strings.ToLower(tc.query)) ||
						strings.Contains(strings.ToLower(script.Type), strings.ToLower(tc.query))

					if !scriptRelevant {
						t.Errorf("Script %s doesn't seem relevant to query %q", script.Name, tc.query)
					}
				}
			})
		}
	})

	t.Run("Performance with Large Result Sets", func(t *testing.T) {
		// Create a large number of scripts for performance testing
		largeTestDir := filepath.Join(testDir, "large")
		if err := os.MkdirAll(largeTestDir, 0755); err != nil {
			t.Fatalf("Failed to create large test directory: %v", err)
		}

		// Create 100 test scripts
		var largeScriptSet []contracts.ScriptInfo
		for i := 0; i < 100; i++ {
			scriptName := filepath.Join("test_script_", string(rune('0'+i%10)), ".sh")
			scriptPath := filepath.Join(largeTestDir, scriptName)
			content := "#!/bin/bash\necho 'test script " + string(rune('0'+i%10)) + "'"

			if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
				t.Fatalf("Failed to create large test script: %v", err)
			}

			// Create script info manually for performance testing
			largeScriptSet = append(largeScriptSet, contracts.ScriptInfo{
				ID:   "test-" + string(rune('0'+i%10)),
				Name: scriptName,
				Path: scriptPath,
				Type: "shell",
			})
		}

		// Test filtering performance
		start := time.Now()
		filtered := discovery.FilterScripts(largeScriptSet, "script")
		duration := time.Since(start)

		// Should complete quickly (under 100ms for 100 scripts)
		if duration > 100*time.Millisecond {
			t.Errorf("FilterScripts took too long with large set: %v", duration)
		}

		// Should find all scripts that match
		if len(filtered) != 100 {
			t.Errorf("FilterScripts with large set: got %d, want 100", len(filtered))
		}

		// Test more specific filter
		start = time.Now()
		specificFiltered := discovery.FilterScripts(largeScriptSet, "script_5")
		duration = time.Since(start)

		if duration > 50*time.Millisecond {
			t.Errorf("Specific FilterScripts took too long: %v", duration)
		}

		if len(specificFiltered) != 10 { // Should find 10 scripts with "5" in name
			t.Errorf("Specific FilterScripts: got %d, want 10", len(specificFiltered))
		}
	})

	t.Run("Filter Stability and Consistency", func(t *testing.T) {
		// Test that filtering is stable and returns consistent results
		query := "backup"

		results1 := discovery.FilterScripts(allScripts, query)
		results2 := discovery.FilterScripts(allScripts, query)
		results3 := discovery.FilterScripts(allScripts, query)

		// All results should be identical
		if len(results1) != len(results2) || len(results1) != len(results3) {
			t.Errorf("FilterScripts not consistent: %d, %d, %d results",
				len(results1), len(results2), len(results3))
		}

		// Verify same scripts in same order
		for i := range results1 {
			if i < len(results2) && results1[i].Path != results2[i].Path {
				t.Error("FilterScripts order not consistent between calls")
			}
			if i < len(results3) && results1[i].Path != results3[i].Path {
				t.Error("FilterScripts order not consistent between calls")
			}
		}
	})

	t.Run("Empty and Edge Case Queries", func(t *testing.T) {
		edgeCases := []struct {
			name  string
			query string
		}{
			{"empty string", ""},
			{"single space", " "},
			{"multiple spaces", "   "},
			{"special characters", "!@#$%"},
			{"unicode characters", "日本語"},
			{"very long query", strings.Repeat("abcdefghijklmnopqrstuvwxyz", 10)},
			{"query with newlines", "backup\nscript"},
			{"query with tabs", "backup\tscript"},
		}

		for _, tc := range edgeCases {
			t.Run(tc.name, func(t *testing.T) {
				// Should not panic or error
				filtered := discovery.FilterScripts(allScripts, tc.query)

				// Empty query should return all scripts
				if tc.query == "" && len(filtered) != len(allScripts) {
					t.Errorf("Empty query should return all scripts: got %d, want %d",
						len(filtered), len(allScripts))
				}

				// Other edge cases should return valid (possibly empty) results
				if filtered == nil {
					t.Error("FilterScripts should never return nil")
				}
			})
		}
	})

	t.Run("Filter with Script Tags", func(t *testing.T) {
		// Create scripts with tags for testing
		scriptsWithTags := []contracts.ScriptInfo{
			{
				Name: "backup_db.sh",
				Type: "shell",
				Tags: []string{"backup", "database", "mysql"},
			},
			{
				Name: "deploy_web.py",
				Type: "python",
				Tags: []string{"deploy", "web", "production"},
			},
			{
				Name: "test_api.js",
				Type: "node",
				Tags: []string{"test", "api", "web"},
			},
			{
				Name: "cleanup.rb",
				Type: "ruby",
				Tags: []string{"maintenance", "cleanup", "logs"},
			},
		}

		testCases := []struct {
			query       string
			expectCount int
			description string
		}{
			{"backup", 1, "scripts tagged with backup"},
			{"web", 2, "scripts tagged with web"},
			{"production", 1, "scripts tagged with production"},
			{"nonexistent", 0, "scripts with non-existent tag"},
		}

		for _, tc := range testCases {
			t.Run("tag_"+tc.query, func(t *testing.T) {
				filtered := discovery.FilterScripts(scriptsWithTags, tc.query)

				if len(filtered) != tc.expectCount {
					t.Errorf("FilterScripts by tag %q = %d results, want %d (%s)",
						tc.query, len(filtered), tc.expectCount, tc.description)
				}

				// Verify all results have the tag or name/type match
				for _, script := range filtered {
					hasTag := false
					for _, tag := range script.Tags {
						if strings.Contains(strings.ToLower(tag), strings.ToLower(tc.query)) {
							hasTag = true
							break
						}
					}

					hasNameMatch := strings.Contains(strings.ToLower(script.Name), strings.ToLower(tc.query))
					hasTypeMatch := strings.Contains(strings.ToLower(script.Type), strings.ToLower(tc.query))

					if !hasTag && !hasNameMatch && !hasTypeMatch {
						t.Errorf("Script %s doesn't match query %q in tags, name, or type", script.Name, tc.query)
					}
				}
			})
		}
	})
}

// TestSearchFilteringWithRealDiscovery tests filtering with actual script discovery
func TestSearchFilteringWithRealDiscovery(t *testing.T) {
	// Create a realistic script structure
	testDir := t.TempDir()
	projectDir := filepath.Join(testDir, "project")
	scriptDirs := []string{
		filepath.Join(projectDir, "scripts", "build"),
		filepath.Join(projectDir, "scripts", "deploy"),
		filepath.Join(projectDir, "scripts", "test"),
		filepath.Join(projectDir, "tools"),
		filepath.Join(projectDir, "bin"),
	}

	// Create directory structure
	for _, dir := range scriptDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create realistic scripts
	scripts := map[string]string{
		// Build scripts
		filepath.Join(projectDir, "scripts", "build", "compile.sh"):    "#!/bin/bash\ngcc -o app src/*.c",
		filepath.Join(projectDir, "scripts", "build", "package.py"):    "#!/usr/bin/env python3\nimport setuptools",
		filepath.Join(projectDir, "scripts", "build", "minify.js"):     "#!/usr/bin/env node\n// Minify JavaScript",

		// Deploy scripts
		filepath.Join(projectDir, "scripts", "deploy", "staging.sh"):   "#!/bin/bash\nrsync -av . staging:/app/",
		filepath.Join(projectDir, "scripts", "deploy", "production.py"): "#!/usr/bin/env python3\n# Deploy to prod",
		filepath.Join(projectDir, "scripts", "deploy", "rollback.sh"):  "#!/bin/bash\ngit checkout HEAD~1",

		// Test scripts
		filepath.Join(projectDir, "scripts", "test", "unit_tests.py"):  "#!/usr/bin/env python3\nimport pytest",
		filepath.Join(projectDir, "scripts", "test", "e2e_tests.js"):   "#!/usr/bin/env node\n// End to end tests",
		filepath.Join(projectDir, "scripts", "test", "benchmark.rb"):   "#!/usr/bin/env ruby\n# Performance tests",

		// Tools
		filepath.Join(projectDir, "tools", "backup.sh"):                "#!/bin/bash\ntar -czf backup.tar.gz .",
		filepath.Join(projectDir, "tools", "monitor.py"):               "#!/usr/bin/env python3\n# System monitor",

		// Bin
		filepath.Join(projectDir, "bin", "app"):                        "#!/bin/bash\nexec /usr/bin/app \"$@\"",
		filepath.Join(projectDir, "bin", "setup"):                      "#!/bin/bash\n./configure && make",
	}

	// Create all scripts
	for path, content := range scripts {
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			t.Fatalf("Failed to create script %s: %v", path, err)
		}
	}

	// Initialize discovery service
	registry, err := services.NewServiceRegistry()
	if err != nil {
		t.Fatalf("Failed to create service registry: %v", err)
	}

	discovery := registry.GetScriptDiscovery()
	ctx := context.Background()

	// Scan the project directory
	results, err := discovery.ScanDirectories(ctx, []string{projectDir})
	if err != nil {
		t.Fatalf("Failed to scan project directory: %v", err)
	}

	// Collect all discovered scripts
	var allScripts []contracts.ScriptInfo
	for _, dir := range results {
		allScripts = append(allScripts, dir.Scripts...)
	}

	t.Run("Real Discovery Integration", func(t *testing.T) {
		if len(allScripts) == 0 {
			t.Fatal("No scripts discovered in project directory")
		}

		// Test filtering discovered scripts
		buildScripts := discovery.FilterScripts(allScripts, "build")
		deployScripts := discovery.FilterScripts(allScripts, "deploy")
		testScripts := discovery.FilterScripts(allScripts, "test")

		// Verify filtering works with real discovery
		if len(buildScripts) == 0 {
			t.Error("Should find build scripts")
		}

		if len(deployScripts) == 0 {
			t.Error("Should find deploy scripts")
		}

		if len(testScripts) == 0 {
			t.Error("Should find test scripts")
		}

		// Test type filtering
		shellScripts := discovery.FilterScripts(allScripts, "shell")
		pythonScripts := discovery.FilterScripts(allScripts, "python")

		if len(shellScripts) == 0 {
			t.Error("Should find shell scripts")
		}

		if len(pythonScripts) == 0 {
			t.Error("Should find python scripts")
		}
	})

	t.Run("Combined Discovery and Filtering Workflow", func(t *testing.T) {
		// Simulate a user workflow: scan then filter
		start := time.Now()

		// 1. Scan directories
		scanResults, err := discovery.ScanDirectories(ctx, []string{projectDir})
		if err != nil {
			t.Fatalf("Scan step failed: %v", err)
		}

		// 2. Extract scripts
		var scripts []contracts.ScriptInfo
		for _, dir := range scanResults {
			scripts = append(scripts, dir.Scripts...)
		}

		// 3. Filter by query
		filtered := discovery.FilterScripts(scripts, "deploy")

		duration := time.Since(start)

		// Workflow should be fast
		if duration > 1*time.Second {
			t.Errorf("Combined workflow took too long: %v", duration)
		}

		// Should find deployment scripts
		if len(filtered) < 2 {
			t.Errorf("Expected at least 2 deploy scripts, got %d", len(filtered))
		}

		// Verify results are relevant
		for _, script := range filtered {
			if !strings.Contains(strings.ToLower(script.Path), "deploy") &&
				!strings.Contains(strings.ToLower(script.Name), "deploy") {
				t.Errorf("Filtered script %s doesn't seem related to deploy", script.Name)
			}
		}
	})
}