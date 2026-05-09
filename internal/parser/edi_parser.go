package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sarabjeet/golang-backend-task/internal/models"
)

func ParseEDI(content string) (*models.Result, error) {
	if content == "" {
		return nil, fmt.Errorf("empty file content")
	}

	lines := strings.Split(content, "\n")
	claims := make([]models.Claim, 0)
	var totalAmount float64
	var parseErrors []string

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		claim, err := parseLine(line)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("line %d: %s", lineNum+1, err.Error()))
			continue
		}

		claims = append(claims, *claim)
		totalAmount += claim.Amount
	}

	if len(claims) == 0 {
		if len(parseErrors) > 0 {
			return nil, fmt.Errorf("failed to parse any valid claims: %s", strings.Join(parseErrors, "; "))
		}
		return nil, fmt.Errorf("no claims found in file")
	}

	result := &models.Result{
		Claims: claims,
		Summary: models.Summary{
			TotalClaims: len(claims),
			TotalAmount: totalAmount,
		},
	}

	return result, nil
}

func parseLine(line string) (*models.Claim, error) {
	parts := strings.Split(line, "*")

	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid format: expected 4 fields separated by *, got %d fields", len(parts))
	}

	if strings.ToUpper(strings.TrimSpace(parts[0])) != "CLAIM" {
		return nil, fmt.Errorf("invalid record type: expected 'CLAIM', got '%s'", parts[0])
	}

	claimID := strings.TrimSpace(parts[1])
	if claimID == "" {
		return nil, fmt.Errorf("empty claim ID")
	}

	memberID := strings.TrimSpace(parts[2])
	if memberID == "" {
		return nil, fmt.Errorf("empty member ID")
	}

	amountStr := strings.TrimSpace(parts[3])
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid amount '%s': %w", amountStr, err)
	}

	if amount < 0 {
		return nil, fmt.Errorf("negative amount not allowed: %f", amount)
	}

	return &models.Claim{
		ClaimID:  claimID,
		MemberID: memberID,
		Amount:   amount,
	}, nil
}
