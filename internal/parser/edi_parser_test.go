package parser

import (
	"testing"
)

func TestParseEDI_ValidContent(t *testing.T) {
	content := `CLAIM*CLM001*MEM123*2500
CLAIM*CLM002*MEM456*3000
CLAIM*CLM003*MEM789*1500`

	result, err := ParseEDI(content)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if len(result.Claims) != 3 {
		t.Errorf("Expected 3 claims, got %d", len(result.Claims))
	}

	if result.Summary.TotalClaims != 3 {
		t.Errorf("Expected total_claims 3, got %d", result.Summary.TotalClaims)
	}

	expectedTotal := 7000.0
	if result.Summary.TotalAmount != expectedTotal {
		t.Errorf("Expected total_amount %.2f, got %.2f", expectedTotal, result.Summary.TotalAmount)
	}

	if result.Claims[0].ClaimID != "CLM001" {
		t.Errorf("Expected claim_id CLM001, got %s", result.Claims[0].ClaimID)
	}
	if result.Claims[0].MemberID != "MEM123" {
		t.Errorf("Expected member_id MEM123, got %s", result.Claims[0].MemberID)
	}
	if result.Claims[0].Amount != 2500 {
		t.Errorf("Expected amount 2500, got %.2f", result.Claims[0].Amount)
	}
}

func TestParseEDI_EmptyContent(t *testing.T) {
	content := ""

	result, err := ParseEDI(content)
	if err == nil {
		t.Fatal("Expected error for empty content, got nil")
	}
	if result != nil {
		t.Error("Expected nil result for empty content")
	}
}

func TestParseEDI_InvalidFormat(t *testing.T) {
	content := "INVALID*LINE*FORMAT"

	result, err := ParseEDI(content)
	if err == nil {
		t.Fatal("Expected error for invalid format, got nil")
	}
	if result != nil {
		t.Error("Expected nil result for invalid format")
	}
}

func TestParseEDI_WrongRecordType(t *testing.T) {
	content := "PAYMENT*CLM001*MEM123*2500"

	_, err := ParseEDI(content)
	if err == nil {
		t.Fatal("Expected error for wrong record type, got nil")
	}
}

func TestParseEDI_InvalidAmount(t *testing.T) {
	content := "CLAIM*CLM001*MEM123*INVALID"

	_, err := ParseEDI(content)
	if err == nil {
		t.Fatal("Expected error for invalid amount, got nil")
	}
}

func TestParseEDI_NegativeAmount(t *testing.T) {
	content := "CLAIM*CLM001*MEM123*-2500"

	_, err := ParseEDI(content)
	if err == nil {
		t.Fatal("Expected error for negative amount, got nil")
	}
}

func TestParseEDI_EmptyClaimID(t *testing.T) {
	content := "CLAIM**MEM123*2500"

	_, err := ParseEDI(content)
	if err == nil {
		t.Fatal("Expected error for empty claim ID, got nil")
	}
}

func TestParseEDI_EmptyMemberID(t *testing.T) {
	content := "CLAIM*CLM001**2500"

	_, err := ParseEDI(content)
	if err == nil {
		t.Fatal("Expected error for empty member ID, got nil")
	}
}

func TestParseEDI_WithEmptyLines(t *testing.T) {
	content := `CLAIM*CLM001*MEM123*2500

CLAIM*CLM002*MEM456*3000

`

	result, err := ParseEDI(content)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Claims) != 2 {
		t.Errorf("Expected 2 claims, got %d", len(result.Claims))
	}
}

func TestParseEDI_MixedValidInvalid(t *testing.T) {
	content := `CLAIM*CLM001*MEM123*2500
INVALID*LINE
CLAIM*CLM002*MEM456*3000`

	result, err := ParseEDI(content)
	if err != nil {
		t.Fatalf("Expected no error for mixed content, got: %v", err)
	}

	if len(result.Claims) != 2 {
		t.Errorf("Expected 2 valid claims, got %d", len(result.Claims))
	}

	if result.Summary.TotalAmount != 5500 {
		t.Errorf("Expected total_amount 5500, got %.2f", result.Summary.TotalAmount)
	}
}

func TestParseEDI_CaseInsensitiveRecordType(t *testing.T) {
	content := `claim*CLM001*MEM123*2500
ClAiM*CLM002*MEM456*3000
CLAIM*CLM003*MEM789*1500`

	result, err := ParseEDI(content)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Claims) != 3 {
		t.Errorf("Expected 3 claims, got %d", len(result.Claims))
	}
}

func TestParseEDI_WithWhitespace(t *testing.T) {
	content := `  CLAIM * CLM001 * MEM123 * 2500  
CLAIM*CLM002*MEM456*3000`

	result, err := ParseEDI(content)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Claims) != 2 {
		t.Errorf("Expected 2 claims, got %d", len(result.Claims))
	}

	if result.Claims[0].ClaimID != "CLM001" {
		t.Errorf("Expected trimmed claim_id CLM001, got %s", result.Claims[0].ClaimID)
	}
}

func TestParseEDI_DecimalAmount(t *testing.T) {
	content := "CLAIM*CLM001*MEM123*2500.50"

	result, err := ParseEDI(content)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Claims[0].Amount != 2500.50 {
		t.Errorf("Expected amount 2500.50, got %.2f", result.Claims[0].Amount)
	}
}
