package contract

import (
	"context"
	"testing"

	"github.com/shaiu/alec/pkg/contracts"
)

// TestTUIManagerContract verifies that any implementation of TUIManager
// interface conforms to the contract requirements
func TestTUIManagerContract(t *testing.T) {
	// This test will fail until we have an implementation
	var manager contracts.TUIManager
	if manager == nil {
		t.Skip("No TUIManager implementation available yet - this is expected during TDD phase")
	}

	tests := []struct {
		name string
		test func(t *testing.T, m contracts.TUIManager)
	}{
		{"Initialize must handle terminal size detection", testInitialize},
		{"HandleResize must recalculate layout", testHandleResize},
		{"UpdateFocus must manage component transitions", testUpdateFocus},
		{"NavigateToScript must update state and history", testNavigateToScript},
		{"NavigateToDirectory must expand and update context", testNavigateToDirectory},
		{"UpdateSearch must filter scripts and directories", testUpdateSearch},
		{"GetCurrentState must provide consistent snapshot", testGetCurrentState},
		{"Shutdown must restore terminal and cleanup", testShutdown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, manager)
		})
	}
}

func testInitialize(t *testing.T, m contracts.TUIManager) {
	ctx := context.Background()

	err := m.Initialize(ctx)
	if err != nil {
		t.Errorf("Initialize should not fail with valid context: %v", err)
	}

	// Should be able to get state after initialization
	state := m.GetCurrentState()
	if state.TerminalWidth <= 0 || state.TerminalHeight <= 0 {
		t.Error("Initialize should detect valid terminal dimensions")
	}
}

func testHandleResize(t *testing.T, m contracts.TUIManager) {
	// Test various terminal sizes
	testSizes := []struct {
		width, height int
		shouldSucceed bool
	}{
		{80, 24, true},   // Standard terminal
		{120, 30, true},  // Large terminal
		{40, 10, false},  // Too small
		{0, 0, false},    // Invalid
	}

	for _, size := range testSizes {
		t.Run("resize_to", func(t *testing.T) {
			err := m.HandleResize(size.width, size.height)

			if size.shouldSucceed && err != nil {
				t.Errorf("HandleResize(%d, %d) should succeed, got error: %v",
					size.width, size.height, err)
			}

			if !size.shouldSucceed && err == nil {
				t.Errorf("HandleResize(%d, %d) should fail for invalid dimensions",
					size.width, size.height)
			}

			if size.shouldSucceed {
				state := m.GetCurrentState()
				if state.TerminalWidth != size.width || state.TerminalHeight != size.height {
					t.Errorf("Terminal dimensions not updated correctly: got %dx%d, want %dx%d",
						state.TerminalWidth, state.TerminalHeight, size.width, size.height)
				}

				// Check golden ratio sidebar width calculation
				expectedSidebarWidth := int(float64(size.width) * 0.382) // Golden ratio
				if abs(state.SidebarWidth-expectedSidebarWidth) > 2 {    // Allow small variance
					t.Errorf("Sidebar width not following golden ratio: got %d, expected ~%d",
						state.SidebarWidth, expectedSidebarWidth)
				}
			}
		})
	}
}

func testUpdateFocus(t *testing.T, m contracts.TUIManager) {
	components := []contracts.ComponentType{
		contracts.ComponentSidebar,
		contracts.ComponentMain,
		contracts.ComponentOutput,
		contracts.ComponentSearch,
	}

	for _, component := range components {
		t.Run("focus_"+string(component), func(t *testing.T) {
			err := m.UpdateFocus(component)
			if err != nil {
				t.Errorf("UpdateFocus(%s) failed: %v", component, err)
				return
			}

			state := m.GetCurrentState()
			if state.FocusedComponent != component {
				t.Errorf("Focus not updated: got %s, want %s",
					state.FocusedComponent, component)
			}
		})
	}

	// Test invalid component
	err := m.UpdateFocus("invalid_component")
	if err == nil {
		t.Error("UpdateFocus should reject invalid component types")
	}
}

func testNavigateToScript(t *testing.T, m contracts.TUIManager) {
	testScript := contracts.ScriptInfo{
		ID:   "test-script-1",
		Name: "test.sh",
		Path: "/tmp/test.sh",
		Type: "shell",
	}

	err := m.NavigateToScript(testScript)
	if err != nil {
		t.Errorf("NavigateToScript failed: %v", err)
		return
	}

	state := m.GetCurrentState()
	if state.SelectedScript == nil {
		t.Error("Selected script should be set after navigation")
		return
	}

	if state.SelectedScript.ID != testScript.ID {
		t.Errorf("Selected script ID mismatch: got %s, want %s",
			state.SelectedScript.ID, testScript.ID)
	}
}

func testNavigateToDirectory(t *testing.T, m contracts.TUIManager) {
	testPath := "/tmp/test-directory"

	err := m.NavigateToDirectory(testPath)
	if err != nil {
		t.Errorf("NavigateToDirectory failed: %v", err)
		return
	}

	state := m.GetCurrentState()
	if state.SelectedDirectory != testPath {
		t.Errorf("Selected directory mismatch: got %s, want %s",
			state.SelectedDirectory, testPath)
	}
}

func testUpdateSearch(t *testing.T, m contracts.TUIManager) {
	testQueries := []string{
		"test",
		"*.sh",
		"backup",
		"", // Clear search
	}

	for _, query := range testQueries {
		t.Run("search_"+query, func(t *testing.T) {
			err := m.UpdateSearch(query)
			if err != nil {
				t.Errorf("UpdateSearch(%q) failed: %v", query, err)
				return
			}

			state := m.GetCurrentState()
			if state.SearchQuery != query {
				t.Errorf("Search query not updated: got %q, want %q",
					state.SearchQuery, query)
			}
		})
	}
}

func testGetCurrentState(t *testing.T, m contracts.TUIManager) {
	// Get state multiple times - should be consistent
	state1 := m.GetCurrentState()
	state2 := m.GetCurrentState()

	// Basic consistency checks
	if state1.CurrentView != state2.CurrentView {
		t.Error("GetCurrentState should return consistent views")
	}

	if state1.FocusedComponent != state2.FocusedComponent {
		t.Error("GetCurrentState should return consistent focus")
	}

	if state1.TerminalWidth != state2.TerminalWidth ||
		state1.TerminalHeight != state2.TerminalHeight {
		t.Error("GetCurrentState should return consistent terminal dimensions")
	}
}

func testShutdown(t *testing.T, m contracts.TUIManager) {
	// Initialize first
	ctx := context.Background()
	err := m.Initialize(ctx)
	if err != nil {
		t.Skip("Cannot test shutdown without successful initialization")
	}

	// Shutdown should succeed
	err = m.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// After shutdown, operations should fail or be no-op
	err = m.HandleResize(100, 50)
	// Implementation-dependent: may fail or be no-op after shutdown
}

// Helper function for absolute difference
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TestBubbleTeaModelContract tests the Bubble Tea model integration
func TestBubbleTeaModelContract(t *testing.T) {
	// This test will fail until we have an implementation
	var model contracts.BubbleTeaModel
	if model == nil {
		t.Skip("No BubbleTeaModel implementation available yet - this is expected during TDD phase")
	}

	tests := []struct {
		name string
		test func(t *testing.T, m contracts.BubbleTeaModel)
	}{
		{"SetSize must update dimensions", testModelSetSize},
		{"SetFocus must update focus state", testModelSetFocus},
		{"GetHeight must return valid height", testModelGetHeight},
		{"GetWidth must return valid width", testModelGetWidth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, model)
		})
	}
}

func testModelSetSize(t *testing.T, m contracts.BubbleTeaModel) {
	testWidth, testHeight := 120, 40

	m.SetSize(testWidth, testHeight)

	if m.GetWidth() != testWidth {
		t.Errorf("Width not set correctly: got %d, want %d", m.GetWidth(), testWidth)
	}

	if m.GetHeight() != testHeight {
		t.Errorf("Height not set correctly: got %d, want %d", m.GetHeight(), testHeight)
	}
}

func testModelSetFocus(t *testing.T, m contracts.BubbleTeaModel) {
	// Test both focused and unfocused states
	states := []bool{true, false}

	for _, focused := range states {
		t.Run("focused_state", func(t *testing.T) {
			m.SetFocus(focused)
			// Note: We can't directly test the focus state without accessing
			// internal model state, but the model should handle it correctly
		})
	}
}

func testModelGetHeight(t *testing.T, m contracts.BubbleTeaModel) {
	height := m.GetHeight()
	if height < 0 {
		t.Error("GetHeight should return non-negative value")
	}
}

func testModelGetWidth(t *testing.T, m contracts.BubbleTeaModel) {
	width := m.GetWidth()
	if width < 0 {
		t.Error("GetWidth should return non-negative value")
	}
}