package loanpro

import (
	"encoding/json"
	"testing"
)

// Mock API response data from actual LoanPro API
const mockTransactionsResponse = `{
    "d": {
        "results": [
            {
                "__metadata": {
                    "uri": "https://loanpro.simnang.com/api/public/api/1/odata.svc/LoanTransactions(id=247006)",
                    "type": "Entity.LoanTransaction"
                },
                "id": 247006,
                "txId": "630-0-info-origin",
                "entityType": "Entity.Loan",
                "entityId": 630,
                "modId": 0,
                "date": "/Date(1764547200)/",
                "period": 0,
                "periodStart": "/Date(-62169984000)/",
                "periodEnd": "/Date(-62169984000)/",
                "title": "Loan Origination",
                "type": "origination",
                "infoOnly": 1,
                "infoDetails": "{\"amount\":\"75000.00\",\"underwriting\":\"0.00\",\"discount\":\"0.00\"}",
                "paymentId": 0,
                "paymentDisplayId": 0,
                "paymentAmount": "0",
                "paymentInterest": "0",
                "paymentPrincipal": "0",
                "paymentDiscount": "0",
                "paymentFees": "0",
                "feesPaidDetails": null,
                "paymentEscrow": "0",
                "paymentEscrowBreakdown": null,
                "chargeAmount": "0",
                "chargeInterest": "0",
                "chargePrincipal": "0",
                "chargeDiscount": "0",
                "chargeFees": "0",
                "chargeEscrow": "0",
                "chargeEscrowBreakdown": null,
                "future": 0,
                "principalOnly": 0,
                "advancement": 0,
                "payoffFee": 0,
                "chargeOff": 0,
                "paymentType": 0,
                "adbDays": 0,
                "adb": "0",
                "principalBalance": "0",
                "displayOrder": "0"
            },
            {
                "__metadata": {
                    "uri": "https://loanpro.simnang.com/api/public/api/1/odata.svc/LoanTransactions(id=247538)",
                    "type": "Entity.LoanTransaction"
                },
                "id": 247538,
                "txId": "630-26-info-mod",
                "entityType": "Entity.Loan",
                "entityId": 630,
                "modId": 26,
                "date": "/Date(1764547200)/",
                "period": 0,
                "periodStart": "/Date(-62169984000)/",
                "periodEnd": "/Date(-62169984000)/",
                "title": "Loan Modification",
                "type": "modification",
                "infoOnly": 1,
                "infoDetails": "{\"amount\":\"75000.00\",\"underwriting\":\"0.00\",\"discount\":\"0.00\"}",
                "paymentId": 0,
                "paymentDisplayId": 0,
                "paymentAmount": "0",
                "paymentInterest": "0",
                "paymentPrincipal": "0",
                "paymentDiscount": "0",
                "paymentFees": "0",
                "feesPaidDetails": null,
                "paymentEscrow": "0",
                "paymentEscrowBreakdown": null,
                "chargeAmount": "0",
                "chargeInterest": "0",
                "chargePrincipal": "0",
                "chargeDiscount": "0",
                "chargeFees": "0",
                "chargeEscrow": "0",
                "chargeEscrowBreakdown": null,
                "future": 0,
                "principalOnly": 0,
                "advancement": 0,
                "payoffFee": 0,
                "chargeOff": 0,
                "paymentType": 0,
                "adbDays": 0,
                "adb": "0",
                "principalBalance": "0",
                "displayOrder": "0"
            },
            {
                "__metadata": {
                    "uri": "https://loanpro.simnang.com/api/public/api/1/odata.svc/LoanTransactions(id=247770)",
                    "type": "Entity.LoanTransaction"
                },
                "id": 247770,
                "txId": "630-26-pay24360",
                "entityType": "Entity.Loan",
                "entityId": 630,
                "modId": 26,
                "date": "/Date(1764892800)/",
                "period": 0,
                "periodStart": "/Date(1764547200)/",
                "periodEnd": "/Date(1767139200)/",
                "title": "Payment: Payoff - 2025/12/05 Other",
                "type": "payment",
                "infoOnly": 0,
                "infoDetails": null,
                "paymentId": 2436,
                "paymentDisplayId": 6975,
                "paymentAmount": "75111.98",
                "paymentInterest": "111.98",
                "paymentPrincipal": "75000",
                "paymentDiscount": "0",
                "paymentFees": "0",
                "feesPaidDetails": null,
                "paymentEscrow": "0",
                "paymentEscrowBreakdown": "{\"subsets\":{\"2\":0,\"3\":0,\"4\":0,\"5\":0,\"6\":0,\"7\":0,\"8\":0,\"9\":0,\"10\":0,\"11\":0,\"12\":0,\"13\":0,\"14\":0,\"15\":0,\"16\":0}}",
                "chargeAmount": "0",
                "chargeInterest": "0",
                "chargePrincipal": "0",
                "chargeDiscount": "0",
                "chargeFees": "0",
                "chargeEscrow": "0",
                "chargeEscrowBreakdown": null,
                "future": 0,
                "principalOnly": 0,
                "advancement": 0,
                "payoffFee": 0,
                "chargeOff": 0,
                "paymentType": 0,
                "adbDays": 30,
                "adb": "0",
                "principalBalance": "0",
                "displayOrder": "5"
            },
            {
                "__metadata": {
                    "uri": "https://loanpro.simnang.com/api/public/api/1/odata.svc/LoanTransactions(id=247539)",
                    "type": "Entity.LoanTransaction"
                },
                "id": 247539,
                "txId": "630-26-spm0",
                "entityType": "Entity.Loan",
                "entityId": 630,
                "modId": 26,
                "date": "/Date(1767225600)/",
                "period": 0,
                "periodStart": "/Date(1764547200)/",
                "periodEnd": "/Date(1767139200)/",
                "title": "Scheduled Payment: 1",
                "type": "scheduledPayment",
                "infoOnly": 0,
                "infoDetails": null,
                "paymentId": 0,
                "paymentDisplayId": 0,
                "paymentAmount": "0",
                "paymentInterest": "0",
                "paymentPrincipal": "0",
                "paymentDiscount": "0",
                "paymentFees": "0",
                "feesPaidDetails": null,
                "paymentEscrow": "0",
                "paymentEscrowBreakdown": null,
                "chargeAmount": "86.69",
                "chargeInterest": "86.69",
                "chargePrincipal": "0",
                "chargeDiscount": "0",
                "chargeFees": "0",
                "chargeEscrow": "0",
                "chargeEscrowBreakdown": "{\"subsets\":{\"2\":0,\"3\":0,\"4\":0,\"5\":0,\"6\":0,\"7\":0,\"8\":0,\"9\":0,\"10\":0,\"11\":0,\"12\":0,\"13\":0,\"14\":0,\"15\":0,\"16\":0}}",
                "future": 1,
                "principalOnly": 0,
                "advancement": 0,
                "payoffFee": 0,
                "chargeOff": 0,
                "paymentType": 0,
                "adbDays": 30,
                "adb": "9677.42",
                "principalBalance": "0",
                "displayOrder": "0"
            },
            {
                "__metadata": {
                    "uri": "https://loanpro.simnang.com/api/public/api/1/odata.svc/LoanTransactions(id=247594)",
                    "type": "Entity.LoanTransaction"
                },
                "id": 247594,
                "txId": "630-26-info-int-rate-chg-152",
                "entityType": "Entity.Loan",
                "entityId": 630,
                "modId": 26,
                "date": "/Date(1796083200)/",
                "period": 0,
                "periodStart": "/Date(-62169984000)/",
                "periodEnd": "/Date(-62169984000)/",
                "title": "Interest Rate Change",
                "type": "intRateChange",
                "infoOnly": 1,
                "infoDetails": "{\"type\":\"loan.interest.rate.change.type.fixed\",\"rate\":\"21.5000000\"}",
                "paymentId": 0,
                "paymentDisplayId": 0,
                "paymentAmount": "0",
                "paymentInterest": "0",
                "paymentPrincipal": "0",
                "paymentDiscount": "0",
                "paymentFees": "0",
                "feesPaidDetails": null,
                "paymentEscrow": "0",
                "paymentEscrowBreakdown": null,
                "chargeAmount": "0",
                "chargeInterest": "0",
                "chargePrincipal": "0",
                "chargeDiscount": "0",
                "chargeFees": "0",
                "chargeEscrow": "0",
                "chargeEscrowBreakdown": null,
                "future": 0,
                "principalOnly": 0,
                "advancement": 0,
                "payoffFee": 0,
                "chargeOff": 0,
                "paymentType": 0,
                "adbDays": 0,
                "adb": "0",
                "principalBalance": "0",
                "displayOrder": "0"
            },
            {
                "__metadata": {
                    "uri": "https://loanpro.simnang.com/api/public/api/1/odata.svc/LoanTransactions(id=247599)",
                    "type": "Entity.LoanTransaction"
                },
                "id": 247599,
                "txId": "630-26-fee800",
                "entityType": "Entity.Loan",
                "entityId": 630,
                "modId": 26,
                "date": "/Date(1796083200)/",
                "period": 12,
                "periodStart": "/Date(1796083200)/",
                "periodEnd": "/Date(1798675200)/",
                "title": "Fee: Extension fee",
                "type": "fee",
                "infoOnly": 0,
                "infoDetails": null,
                "paymentId": 800,
                "paymentDisplayId": 5959,
                "paymentAmount": "0",
                "paymentInterest": "0",
                "paymentPrincipal": "0",
                "paymentDiscount": "0",
                "paymentFees": "0",
                "feesPaidDetails": null,
                "paymentEscrow": "0",
                "paymentEscrowBreakdown": null,
                "chargeAmount": "1000",
                "chargeInterest": "0",
                "chargePrincipal": "0",
                "chargeDiscount": "0",
                "chargeFees": "1000",
                "chargeEscrow": "0",
                "chargeEscrowBreakdown": null,
                "future": 0,
                "principalOnly": 0,
                "advancement": 0,
                "payoffFee": 0,
                "chargeOff": 0,
                "paymentType": 0,
                "adbDays": 30,
                "adb": "0",
                "principalBalance": "0",
                "displayOrder": "1"
            }
        ],
        "summary": {
            "start": 0,
            "pageSize": 50,
            "total": 6
        }
    }
}`

func TestTransactionUnmarshal(t *testing.T) {
	var response ODataResponse
	err := json.Unmarshal([]byte(mockTransactionsResponse), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal mock response: %v", err)
	}

	// Marshal and unmarshal the D field to get transactions
	transactionsData, err := json.Marshal(response.D)
	if err != nil {
		t.Fatalf("Failed to marshal transaction data: %v", err)
	}

	var wrapper TransactionsWrapper
	err = json.Unmarshal(transactionsData, &wrapper)
	if err != nil {
		t.Fatalf("Failed to unmarshal transactions wrapper: %v", err)
	}

	// Verify we got all 6 transactions
	if len(wrapper.Results) != 6 {
		t.Errorf("Expected 6 transactions, got %d", len(wrapper.Results))
	}

	// Test first transaction (Loan Origination)
	tx1 := wrapper.Results[0]
	if string(tx1.ID) != "247006" {
		t.Errorf("Expected ID 247006, got %s", string(tx1.ID))
	}
	if tx1.TxID != "630-0-info-origin" {
		t.Errorf("Expected TxID '630-0-info-origin', got %s", tx1.TxID)
	}
	if string(tx1.EntityID) != "630" {
		t.Errorf("Expected EntityID 630, got %s", string(tx1.EntityID))
	}
	if tx1.Title != "Loan Origination" {
		t.Errorf("Expected Title 'Loan Origination', got %s", tx1.Title)
	}
	if tx1.Type != "origination" {
		t.Errorf("Expected Type 'origination', got %s", tx1.Type)
	}
	if string(tx1.InfoOnly) != "1" {
		t.Errorf("Expected InfoOnly '1', got %s", string(tx1.InfoOnly))
	}

	// Test third transaction (Payment)
	tx3 := wrapper.Results[2]
	if string(tx3.ID) != "247770" {
		t.Errorf("Expected ID 247770, got %s", string(tx3.ID))
	}
	if tx3.TxID != "630-26-pay24360" {
		t.Errorf("Expected TxID '630-26-pay24360', got %s", tx3.TxID)
	}
	if tx3.Type != "payment" {
		t.Errorf("Expected Type 'payment', got %s", tx3.Type)
	}
	if tx3.PaymentAmount != "75111.98" {
		t.Errorf("Expected PaymentAmount '75111.98', got %s", tx3.PaymentAmount)
	}
	if tx3.PaymentPrincipal != "75000" {
		t.Errorf("Expected PaymentPrincipal '75000', got %s", tx3.PaymentPrincipal)
	}
	if tx3.PaymentInterest != "111.98" {
		t.Errorf("Expected PaymentInterest '111.98', got %s", tx3.PaymentInterest)
	}
	if string(tx3.PaymentID) != "2436" {
		t.Errorf("Expected PaymentID '2436', got %s", string(tx3.PaymentID))
	}
	if string(tx3.PaymentDisplayID) != "6975" {
		t.Errorf("Expected PaymentDisplayID '6975', got %s", string(tx3.PaymentDisplayID))
	}

	// Test fourth transaction (Scheduled Payment with charges)
	tx4 := wrapper.Results[3]
	if tx4.Type != "scheduledPayment" {
		t.Errorf("Expected Type 'scheduledPayment', got %s", tx4.Type)
	}
	if tx4.ChargeAmount != "86.69" {
		t.Errorf("Expected ChargeAmount '86.69', got %s", tx4.ChargeAmount)
	}
	if tx4.ChargeInterest != "86.69" {
		t.Errorf("Expected ChargeInterest '86.69', got %s", tx4.ChargeInterest)
	}
	if string(tx4.Future) != "1" {
		t.Errorf("Expected Future '1', got %s", string(tx4.Future))
	}

	// Test sixth transaction (Fee)
	tx6 := wrapper.Results[5]
	if tx6.Type != "fee" {
		t.Errorf("Expected Type 'fee', got %s", tx6.Type)
	}
	if tx6.ChargeAmount != "1000" {
		t.Errorf("Expected ChargeAmount '1000', got %s", tx6.ChargeAmount)
	}
	if tx6.ChargeFees != "1000" {
		t.Errorf("Expected ChargeFees '1000', got %s", tx6.ChargeFees)
	}
	if string(tx6.Period) != "12" {
		t.Errorf("Expected Period '12', got %s", string(tx6.Period))
	}
}

func TestTransactionMethods(t *testing.T) {
	// Create a sample payment transaction
	paymentTx := Transaction{
		ID:               json.Number("247770"),
		TxID:             "630-26-pay24360",
		EntityID:         json.Number("630"),
		Date:             "/Date(1764892800)/",
		PeriodStart:      "/Date(1764547200)/",
		PeriodEnd:        "/Date(1767139200)/",
		Title:            "Payment: Payoff",
		Type:             "payment",
		InfoOnly:         json.Number("0"),
		PaymentAmount:    "75111.98",
		PaymentPrincipal: "75000",
		PaymentInterest:  "111.98",
		PaymentFees:      "0",
		PaymentEscrow:    "0",
		Future:           json.Number("0"),
	}

	// Test getter methods
	if paymentTx.GetID() != "247770" {
		t.Errorf("Expected ID '247770', got %s", paymentTx.GetID())
	}
	if paymentTx.GetTxID() != "630-26-pay24360" {
		t.Errorf("Expected TxID '630-26-pay24360', got %s", paymentTx.GetTxID())
	}
	if paymentTx.GetEntityID() != "630" {
		t.Errorf("Expected EntityID '630', got %s", paymentTx.GetEntityID())
	}
	if paymentTx.GetTitle() != "Payment: Payoff" {
		t.Errorf("Expected Title 'Payment: Payoff', got %s", paymentTx.GetTitle())
	}
	if paymentTx.GetType() != "payment" {
		t.Errorf("Expected Type 'payment', got %s", paymentTx.GetType())
	}

	// Test date parsing
	expectedDate := "2025-12-05"
	if paymentTx.GetDate() != expectedDate {
		t.Errorf("Expected Date '%s', got %s", expectedDate, paymentTx.GetDate())
	}

	// Test payment breakdown getters
	if paymentTx.GetPaymentAmount() != "75111.98" {
		t.Errorf("Expected PaymentAmount '75111.98', got %s", paymentTx.GetPaymentAmount())
	}
	if paymentTx.GetPaymentPrincipal() != "75000" {
		t.Errorf("Expected PaymentPrincipal '75000', got %s", paymentTx.GetPaymentPrincipal())
	}
	if paymentTx.GetPaymentInterest() != "111.98" {
		t.Errorf("Expected PaymentInterest '111.98', got %s", paymentTx.GetPaymentInterest())
	}

	// Test boolean methods
	if paymentTx.IsInfoOnly() {
		t.Error("Expected IsInfoOnly to be false")
	}
	if paymentTx.IsFuture() {
		t.Error("Expected IsFuture to be false")
	}

	// Test HasPaymentBreakdown
	if !paymentTx.HasPaymentBreakdown() {
		t.Error("Expected HasPaymentBreakdown to be true for payment transaction")
	}
}

func TestTransactionInfoOnly(t *testing.T) {
	// Create an info-only transaction (Loan Origination)
	infoTx := Transaction{
		ID:          json.Number("247006"),
		TxID:        "630-0-info-origin",
		EntityID:    json.Number("630"),
		Title:       "Loan Origination",
		Type:        "origination",
		InfoOnly:    json.Number("1"),
		InfoDetails: `{"amount":"75000.00","underwriting":"0.00","discount":"0.00"}`,
	}

	if !infoTx.IsInfoOnly() {
		t.Error("Expected IsInfoOnly to be true for origination transaction")
	}
	if infoTx.GetInfoDetails() != `{"amount":"75000.00","underwriting":"0.00","discount":"0.00"}` {
		t.Errorf("Expected InfoDetails, got %s", infoTx.GetInfoDetails())
	}
}

func TestTransactionChargeBreakdown(t *testing.T) {
	// Create a scheduled payment with charges
	chargeTx := Transaction{
		ID:              json.Number("247539"),
		Type:            "scheduledPayment",
		ChargeAmount:    "86.69",
		ChargeInterest:  "86.69",
		ChargePrincipal: "0",
		ChargeFees:      "0",
		ChargeEscrow:    "0",
		Future:          json.Number("1"),
	}

	// Test charge breakdown getters
	if chargeTx.GetChargeAmount() != "86.69" {
		t.Errorf("Expected ChargeAmount '86.69', got %s", chargeTx.GetChargeAmount())
	}
	if chargeTx.GetChargeInterest() != "86.69" {
		t.Errorf("Expected ChargeInterest '86.69', got %s", chargeTx.GetChargeInterest())
	}

	// Test HasChargeBreakdown
	if !chargeTx.HasChargeBreakdown() {
		t.Error("Expected HasChargeBreakdown to be true for scheduled payment")
	}

	// Test IsFuture
	if !chargeTx.IsFuture() {
		t.Error("Expected IsFuture to be true for future scheduled payment")
	}
}

func TestTransactionFee(t *testing.T) {
	// Create a fee transaction
	feeTx := Transaction{
		ID:              json.Number("247599"),
		Type:            "fee",
		Title:           "Fee: Extension fee",
		ChargeAmount:    "1000",
		ChargeFees:      "1000",
		ChargeInterest:  "0",
		ChargePrincipal: "0",
		Period:          json.Number("12"),
	}

	if feeTx.GetType() != "fee" {
		t.Errorf("Expected Type 'fee', got %s", feeTx.GetType())
	}
	if feeTx.GetChargeAmount() != "1000" {
		t.Errorf("Expected ChargeAmount '1000', got %s", feeTx.GetChargeAmount())
	}
	if feeTx.GetChargeFees() != "1000" {
		t.Errorf("Expected ChargeFees '1000', got %s", feeTx.GetChargeFees())
	}
	if !feeTx.HasChargeBreakdown() {
		t.Error("Expected HasChargeBreakdown to be true for fee transaction")
	}
}

func TestTransactionZeroAmounts(t *testing.T) {
	// Create a transaction with all zero amounts
	zeroTx := Transaction{
		PaymentAmount:    "0",
		PaymentPrincipal: "0.00",
		PaymentInterest:  "0",
		PaymentFees:      "0",
		ChargeAmount:     "0.00",
		ChargeInterest:   "0",
	}

	// Test that zero amounts don't trigger breakdown flags
	if zeroTx.HasPaymentBreakdown() {
		t.Error("Expected HasPaymentBreakdown to be false for zero amounts")
	}
	if zeroTx.HasChargeBreakdown() {
		t.Error("Expected HasChargeBreakdown to be false for zero amounts")
	}
}
