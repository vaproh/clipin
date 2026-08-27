package payout_test

import (
	"testing"

	"clipin/apps/api/internal/payout"
)

func TestNewRazorpayProviderOrStub_WithKeys(t *testing.T) {
	provider := payout.NewRazorpayProviderOrStub("rzp_test_key", "secret123", "acc123")
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
	// Should be a real RazorpayProvider, not a stub
	if _, ok := provider.(*payout.RazorpayProvider); !ok {
		t.Error("expected *RazorpayProvider, got different type")
	}
}

func TestNewRazorpayProviderOrStub_EmptyKeyID(t *testing.T) {
	provider := payout.NewRazorpayProviderOrStub("", "secret123", "acc123")
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
	// Should fall back to stub
	if _, ok := provider.(*payout.RazorpayStub); !ok {
		t.Error("expected *RazorpayStub when key_id is empty")
	}
}

func TestNewRazorpayProviderOrStub_EmptyKeySecret(t *testing.T) {
	provider := payout.NewRazorpayProviderOrStub("rzp_test_key", "", "acc123")
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
	// Should fall back to stub
	if _, ok := provider.(*payout.RazorpayStub); !ok {
		t.Error("expected *RazorpayStub when key_secret is empty")
	}
}

func TestNewRazorpayProviderOrStub_BothEmpty(t *testing.T) {
	provider := payout.NewRazorpayProviderOrStub("", "", "")
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
	if _, ok := provider.(*payout.RazorpayStub); !ok {
		t.Error("expected *RazorpayStub when both keys are empty")
	}
}

func TestMapPayoutStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"processed", "processed", "completed"},
		{"processing", "processing", "processing"},
		{"pending", "pending", "pending"},
		{"reversed", "reversed", "failed"},
		{"failed", "failed", "failed"},
		{"unknown string", "unknown_status", "pending"},
		{"nil value", nil, "pending"},
		{"integer value", 123, "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := payout.MapPayoutStatus(tt.input)
			if result != tt.expected {
				t.Errorf("MapPayoutStatus(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRedactUPI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal UPI", "user@paytm", "us****tm"},
		{"short UPI", "ab@x", "****"},
		{"very short", "a@b", "****"},
		{"long UPI", "verylongusername@bank", "ve****nk"},
		{"exactly 4 chars", "ab@x", "****"},
		{"5 chars", "abc@x", "ab****@x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := payout.RedactUPI(tt.input)
			if result != tt.expected {
				t.Errorf("redactUPI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewRazorpayProvider_Creation(t *testing.T) {
	// Test that provider can be constructed without making API calls
	provider := payout.NewRazorpayProvider("rzp_test_key", "secret123", "acc123")
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestTransferResponse_Structure(t *testing.T) {
	resp := &payout.TransferResponse{
		TransferID: "pout_123",
		Status:     "completed",
		Reference:  "ref-456",
	}
	if resp.TransferID != "pout_123" {
		t.Errorf("expected TransferID 'pout_123', got %s", resp.TransferID)
	}
	if resp.Status != "completed" {
		t.Errorf("expected Status 'completed', got %s", resp.Status)
	}
	if resp.Reference != "ref-456" {
		t.Errorf("expected Reference 'ref-456', got %s", resp.Reference)
	}
}
