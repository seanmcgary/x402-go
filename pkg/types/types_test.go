package types

import (
	"encoding/json"
	"testing"
)

func TestPaymentRequirementsJSON(t *testing.T) {
	// Test marshaling
	pr := PaymentRequirements{
		Scheme:            "exact",
		Network:           "base-sepolia",
		MaxAmountRequired: "10000",
		Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
		Resource:          "https://api.example.com/premium-data",
		Description:       "Access to premium market data",
		MimeType:          "application/json",
		MaxTimeoutSeconds: 60,
		Extra: map[string]interface{}{
			"name":    "USDC",
			"version": "2",
		},
	}

	data, err := json.Marshal(pr)
	if err != nil {
		t.Fatalf("Failed to marshal PaymentRequirements: %v", err)
	}

	// Test unmarshaling
	var pr2 PaymentRequirements
	err = json.Unmarshal(data, &pr2)
	if err != nil {
		t.Fatalf("Failed to unmarshal PaymentRequirements: %v", err)
	}

	// Verify fields
	if pr2.Scheme != pr.Scheme {
		t.Errorf("Expected scheme %s, got %s", pr.Scheme, pr2.Scheme)
	}
	if pr2.Network != pr.Network {
		t.Errorf("Expected network %s, got %s", pr.Network, pr2.Network)
	}
	if pr2.MaxAmountRequired != pr.MaxAmountRequired {
		t.Errorf("Expected maxAmountRequired %s, got %s", pr.MaxAmountRequired, pr2.MaxAmountRequired)
	}
	if pr2.Asset != pr.Asset {
		t.Errorf("Expected asset %s, got %s", pr.Asset, pr2.Asset)
	}
	if pr2.PayTo != pr.PayTo {
		t.Errorf("Expected payTo %s, got %s", pr.PayTo, pr2.PayTo)
	}
	if pr2.Resource != pr.Resource {
		t.Errorf("Expected resource %s, got %s", pr.Resource, pr2.Resource)
	}
	if pr2.Description != pr.Description {
		t.Errorf("Expected description %s, got %s", pr.Description, pr2.Description)
	}
	if pr2.MimeType != pr.MimeType {
		t.Errorf("Expected mimeType %s, got %s", pr.MimeType, pr2.MimeType)
	}
	if pr2.MaxTimeoutSeconds != pr.MaxTimeoutSeconds {
		t.Errorf("Expected maxTimeoutSeconds %d, got %d", pr.MaxTimeoutSeconds, pr2.MaxTimeoutSeconds)
	}
}

func TestPaymentRequirementsResponseJSON(t *testing.T) {
	prr := PaymentRequirementsResponse{
		X402Version: 1,
		Error:       "X-PAYMENT header is required",
		Accepts: []PaymentRequirements{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/premium-data",
				Description:       "Access to premium market data",
				MimeType:          "application/json",
				MaxTimeoutSeconds: 60,
			},
		},
	}

	data, err := json.Marshal(prr)
	if err != nil {
		t.Fatalf("Failed to marshal PaymentRequirementsResponse: %v", err)
	}

	var prr2 PaymentRequirementsResponse
	err = json.Unmarshal(data, &prr2)
	if err != nil {
		t.Fatalf("Failed to unmarshal PaymentRequirementsResponse: %v", err)
	}

	if prr2.X402Version != prr.X402Version {
		t.Errorf("Expected x402Version %d, got %d", prr.X402Version, prr2.X402Version)
	}
	if prr2.Error != prr.Error {
		t.Errorf("Expected error %s, got %s", prr.Error, prr2.Error)
	}
	if len(prr2.Accepts) != len(prr.Accepts) {
		t.Errorf("Expected %d accepts, got %d", len(prr.Accepts), len(prr2.Accepts))
	}
}

func TestAuthorizationJSON(t *testing.T) {
	auth := Authorization{
		From:        "0x857b06519E91e3A54538791bDbb0E22373e36b66",
		To:          "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
		Value:       "10000",
		ValidAfter:  "1740672089",
		ValidBefore: "1740672154",
		Nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
	}

	data, err := json.Marshal(auth)
	if err != nil {
		t.Fatalf("Failed to marshal Authorization: %v", err)
	}

	var auth2 Authorization
	err = json.Unmarshal(data, &auth2)
	if err != nil {
		t.Fatalf("Failed to unmarshal Authorization: %v", err)
	}

	if auth2.From != auth.From {
		t.Errorf("Expected from %s, got %s", auth.From, auth2.From)
	}
	if auth2.To != auth.To {
		t.Errorf("Expected to %s, got %s", auth.To, auth2.To)
	}
	if auth2.Value != auth.Value {
		t.Errorf("Expected value %s, got %s", auth.Value, auth2.Value)
	}
	if auth2.ValidAfter != auth.ValidAfter {
		t.Errorf("Expected validAfter %s, got %s", auth.ValidAfter, auth2.ValidAfter)
	}
	if auth2.ValidBefore != auth.ValidBefore {
		t.Errorf("Expected validBefore %s, got %s", auth.ValidBefore, auth2.ValidBefore)
	}
	if auth2.Nonce != auth.Nonce {
		t.Errorf("Expected nonce %s, got %s", auth.Nonce, auth2.Nonce)
	}
}

func TestPaymentPayloadJSON(t *testing.T) {
	pp := PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "base-sepolia",
		Payload: SchemePayload{
			Signature: "0x2d6a7588d6acca505cbf0d9a4a227e0c52c6c34008c8e8986a1283259764173608a2ce6496642e377d6da8dbbf5836e9bd15092f9ecab05ded3d6293af148b571c",
			Authorization: Authorization{
				From:        "0x857b06519E91e3A54538791bDbb0E22373e36b66",
				To:          "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Value:       "10000",
				ValidAfter:  "1740672089",
				ValidBefore: "1740672154",
				Nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
			},
		},
	}

	data, err := json.Marshal(pp)
	if err != nil {
		t.Fatalf("Failed to marshal PaymentPayload: %v", err)
	}

	var pp2 PaymentPayload
	err = json.Unmarshal(data, &pp2)
	if err != nil {
		t.Fatalf("Failed to unmarshal PaymentPayload: %v", err)
	}

	if pp2.X402Version != pp.X402Version {
		t.Errorf("Expected x402Version %d, got %d", pp.X402Version, pp2.X402Version)
	}
	if pp2.Scheme != pp.Scheme {
		t.Errorf("Expected scheme %s, got %s", pp.Scheme, pp2.Scheme)
	}
	if pp2.Network != pp.Network {
		t.Errorf("Expected network %s, got %s", pp.Network, pp2.Network)
	}
	if pp2.Payload.Signature != pp.Payload.Signature {
		t.Errorf("Expected signature %s, got %s", pp.Payload.Signature, pp2.Payload.Signature)
	}
	if pp2.Payload.Authorization.From != pp.Payload.Authorization.From {
		t.Errorf("Expected from %s, got %s", pp.Payload.Authorization.From, pp2.Payload.Authorization.From)
	}
}

func TestSettlementResponseJSON(t *testing.T) {
	// Test successful settlement
	sr := SettlementResponse{
		Success:     true,
		Transaction: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		Network:     "base-sepolia",
		Payer:       "0x857b06519E91e3A54538791bDbb0E22373e36b66",
	}

	data, err := json.Marshal(sr)
	if err != nil {
		t.Fatalf("Failed to marshal SettlementResponse: %v", err)
	}

	var sr2 SettlementResponse
	err = json.Unmarshal(data, &sr2)
	if err != nil {
		t.Fatalf("Failed to unmarshal SettlementResponse: %v", err)
	}

	if sr2.Success != sr.Success {
		t.Errorf("Expected success %v, got %v", sr.Success, sr2.Success)
	}
	if sr2.Transaction != sr.Transaction {
		t.Errorf("Expected transaction %s, got %s", sr.Transaction, sr2.Transaction)
	}
	if sr2.Network != sr.Network {
		t.Errorf("Expected network %s, got %s", sr.Network, sr2.Network)
	}
	if sr2.Payer != sr.Payer {
		t.Errorf("Expected payer %s, got %s", sr.Payer, sr2.Payer)
	}

	// Test failed settlement
	srFailed := SettlementResponse{
		Success:     false,
		ErrorReason: "insufficient_funds",
		Transaction: "",
		Network:     "base-sepolia",
		Payer:       "0x857b06519E91e3A54538791bDbb0E22373e36b66",
	}

	data, err = json.Marshal(srFailed)
	if err != nil {
		t.Fatalf("Failed to marshal failed SettlementResponse: %v", err)
	}

	var srFailed2 SettlementResponse
	err = json.Unmarshal(data, &srFailed2)
	if err != nil {
		t.Fatalf("Failed to unmarshal failed SettlementResponse: %v", err)
	}

	if srFailed2.Success != srFailed.Success {
		t.Errorf("Expected success %v, got %v", srFailed.Success, srFailed2.Success)
	}
	if srFailed2.ErrorReason != srFailed.ErrorReason {
		t.Errorf("Expected errorReason %s, got %s", srFailed.ErrorReason, srFailed2.ErrorReason)
	}
}

func TestVerifyRequestJSON(t *testing.T) {
	vr := VerifyRequest{
		PaymentPayload: PaymentPayload{
			X402Version: 1,
			Scheme:      "exact",
			Network:     "base-sepolia",
			Payload: SchemePayload{
				Signature: "0x2d6a7588d6acca505cbf0d9a4a227e0c52c6c34008c8e8986a1283259764173608a2ce6496642e377d6da8dbbf5836e9bd15092f9ecab05ded3d6293af148b571c",
				Authorization: Authorization{
					From:        "0x857b06519E91e3A54538791bDbb0E22373e36b66",
					To:          "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
					Value:       "10000",
					ValidAfter:  "1740672089",
					ValidBefore: "1740672154",
					Nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
				},
			},
		},
		PaymentRequirements: PaymentRequirements{
			Scheme:            "exact",
			Network:           "base-sepolia",
			MaxAmountRequired: "10000",
			Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
			Resource:          "https://api.example.com/premium-data",
			Description:       "Access to premium market data",
			MimeType:          "application/json",
			MaxTimeoutSeconds: 60,
		},
	}

	data, err := json.Marshal(vr)
	if err != nil {
		t.Fatalf("Failed to marshal VerifyRequest: %v", err)
	}

	var vr2 VerifyRequest
	err = json.Unmarshal(data, &vr2)
	if err != nil {
		t.Fatalf("Failed to unmarshal VerifyRequest: %v", err)
	}

	if vr2.PaymentPayload.Scheme != vr.PaymentPayload.Scheme {
		t.Errorf("Expected scheme %s, got %s", vr.PaymentPayload.Scheme, vr2.PaymentPayload.Scheme)
	}
	if vr2.PaymentRequirements.MaxAmountRequired != vr.PaymentRequirements.MaxAmountRequired {
		t.Errorf("Expected maxAmountRequired %s, got %s", vr.PaymentRequirements.MaxAmountRequired, vr2.PaymentRequirements.MaxAmountRequired)
	}
}

func TestVerifyResponseJSON(t *testing.T) {
	// Test valid response
	vr := VerifyResponse{
		IsValid: true,
		Payer:   "0x857b06519E91e3A54538791bDbb0E22373e36b66",
	}

	data, err := json.Marshal(vr)
	if err != nil {
		t.Fatalf("Failed to marshal VerifyResponse: %v", err)
	}

	var vr2 VerifyResponse
	err = json.Unmarshal(data, &vr2)
	if err != nil {
		t.Fatalf("Failed to unmarshal VerifyResponse: %v", err)
	}

	if vr2.IsValid != vr.IsValid {
		t.Errorf("Expected isValid %v, got %v", vr.IsValid, vr2.IsValid)
	}
	if vr2.Payer != vr.Payer {
		t.Errorf("Expected payer %s, got %s", vr.Payer, vr2.Payer)
	}

	// Test invalid response
	vrInvalid := VerifyResponse{
		IsValid:       false,
		InvalidReason: "insufficient_funds",
		Payer:         "0x857b06519E91e3A54538791bDbb0E22373e36b66",
	}

	data, err = json.Marshal(vrInvalid)
	if err != nil {
		t.Fatalf("Failed to marshal invalid VerifyResponse: %v", err)
	}

	var vrInvalid2 VerifyResponse
	err = json.Unmarshal(data, &vrInvalid2)
	if err != nil {
		t.Fatalf("Failed to unmarshal invalid VerifyResponse: %v", err)
	}

	if vrInvalid2.IsValid != vrInvalid.IsValid {
		t.Errorf("Expected isValid %v, got %v", vrInvalid.IsValid, vrInvalid2.IsValid)
	}
	if vrInvalid2.InvalidReason != vrInvalid.InvalidReason {
		t.Errorf("Expected invalidReason %s, got %s", vrInvalid.InvalidReason, vrInvalid2.InvalidReason)
	}
}

func TestSupportedResponseJSON(t *testing.T) {
	sr := SupportedResponse{
		Kinds: []SupportedKind{
			{
				X402Version: 1,
				Scheme:      "exact",
				Network:     "base-sepolia",
			},
			{
				X402Version: 1,
				Scheme:      "exact",
				Network:     "base",
			},
		},
	}

	data, err := json.Marshal(sr)
	if err != nil {
		t.Fatalf("Failed to marshal SupportedResponse: %v", err)
	}

	var sr2 SupportedResponse
	err = json.Unmarshal(data, &sr2)
	if err != nil {
		t.Fatalf("Failed to unmarshal SupportedResponse: %v", err)
	}

	if len(sr2.Kinds) != len(sr.Kinds) {
		t.Errorf("Expected %d kinds, got %d", len(sr.Kinds), len(sr2.Kinds))
	}

	for i, kind := range sr.Kinds {
		if sr2.Kinds[i].X402Version != kind.X402Version {
			t.Errorf("Expected x402Version %d, got %d", kind.X402Version, sr2.Kinds[i].X402Version)
		}
		if sr2.Kinds[i].Scheme != kind.Scheme {
			t.Errorf("Expected scheme %s, got %s", kind.Scheme, sr2.Kinds[i].Scheme)
		}
		if sr2.Kinds[i].Network != kind.Network {
			t.Errorf("Expected network %s, got %s", kind.Network, sr2.Kinds[i].Network)
		}
	}
}

func TestDiscoveryResponseJSON(t *testing.T) {
	dr := DiscoveryResponse{
		X402Version: 1,
		Items: []DiscoveredResource{
			{
				Resource:    "https://api.example.com/premium-data",
				Type:        "http",
				X402Version: 1,
				Accepts: []PaymentRequirements{
					{
						Scheme:            "exact",
						Network:           "base-sepolia",
						MaxAmountRequired: "10000",
						Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
						PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
						Resource:          "https://api.example.com/premium-data",
						Description:       "Access to premium market data",
						MimeType:          "application/json",
						MaxTimeoutSeconds: 60,
					},
				},
				LastUpdated: 1703123456,
				Metadata: map[string]interface{}{
					"category": "finance",
					"provider": "Example Corp",
				},
			},
		},
		Pagination: Pagination{
			Limit:  10,
			Offset: 0,
			Total:  1,
		},
	}

	data, err := json.Marshal(dr)
	if err != nil {
		t.Fatalf("Failed to marshal DiscoveryResponse: %v", err)
	}

	var dr2 DiscoveryResponse
	err = json.Unmarshal(data, &dr2)
	if err != nil {
		t.Fatalf("Failed to unmarshal DiscoveryResponse: %v", err)
	}

	if dr2.X402Version != dr.X402Version {
		t.Errorf("Expected x402Version %d, got %d", dr.X402Version, dr2.X402Version)
	}
	if len(dr2.Items) != len(dr.Items) {
		t.Errorf("Expected %d items, got %d", len(dr.Items), len(dr2.Items))
	}
	if dr2.Pagination.Limit != dr.Pagination.Limit {
		t.Errorf("Expected limit %d, got %d", dr.Pagination.Limit, dr2.Pagination.Limit)
	}
	if dr2.Pagination.Offset != dr.Pagination.Offset {
		t.Errorf("Expected offset %d, got %d", dr.Pagination.Offset, dr2.Pagination.Offset)
	}
	if dr2.Pagination.Total != dr.Pagination.Total {
		t.Errorf("Expected total %d, got %d", dr.Pagination.Total, dr2.Pagination.Total)
	}
}

func TestErrorCodeConstants(t *testing.T) {
	// Test that error code constants are defined correctly
	errorCodes := []string{
		ErrorInsufficientFunds,
		ErrorInvalidExactEVMPayloadAuthorizationValidAfter,
		ErrorInvalidExactEVMPayloadAuthorizationValidBefore,
		ErrorInvalidExactEVMPayloadAuthorizationValue,
		ErrorInvalidExactEVMPayloadSignature,
		ErrorInvalidExactEVMPayloadRecipientMismatch,
		ErrorInvalidNetwork,
		ErrorInvalidPayload,
		ErrorInvalidPaymentRequirements,
		ErrorInvalidScheme,
		ErrorUnsupportedScheme,
		ErrorInvalidX402Version,
		ErrorInvalidTransactionState,
		ErrorUnexpectedVerifyError,
		ErrorUnexpectedSettleError,
	}

	expectedCodes := []string{
		"insufficient_funds",
		"invalid_exact_evm_payload_authorization_valid_after",
		"invalid_exact_evm_payload_authorization_valid_before",
		"invalid_exact_evm_payload_authorization_value",
		"invalid_exact_evm_payload_signature",
		"invalid_exact_evm_payload_recipient_mismatch",
		"invalid_network",
		"invalid_payload",
		"invalid_payment_requirements",
		"invalid_scheme",
		"unsupported_scheme",
		"invalid_x402_version",
		"invalid_transaction_state",
		"unexpected_verify_error",
		"unexpected_settle_error",
	}

	for i, code := range errorCodes {
		if code != expectedCodes[i] {
			t.Errorf("Expected error code %s, got %s", expectedCodes[i], code)
		}
	}
}
