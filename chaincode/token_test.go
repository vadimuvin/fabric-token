package main

import (
    "testing"
    "github.com/hyperledger/fabric/core/chaincode/shim"
    "github.com/hyperledger/fabric/common/util"
    "github.com/token/chaincode/mock"
    "encoding/json"
    "github.com/token/chaincode/testdata"
    "encoding/binary"
    "errors"
)

var fabricToken = Token{
    Name: "FabricToken",
    Symbol: "FT",
    Decimals: 2,
    TotalSupply: 10000,
}

func initToken(t *testing.T) *mock.FullMockStub {
    token := &TokenChaincode{}

    stub := mock.NewFullMockStub("token", token)
    stub.MockCreator("default", testdata.TestUser1Cert)

    tokenBytes, _ := json.Marshal(fabricToken)
    res := stub.MockInit("1", util.ToChaincodeArgs("init", string(tokenBytes)))
    if res.Status != shim.OK {
        t.Error("Token cc init failed: " + res.Message)
    }

    var tokenDataBytes []byte = nil
    var callerBalanceBytes []byte = nil
    for key, val := range stub.State {
        if key == KeyToken {
            tokenDataBytes = val
        } else {
            callerBalanceBytes = val
        }
    }

    if tokenDataBytes == nil || callerBalanceBytes == nil {
        t.Error("Expected value not found in the state")
        t.FailNow()
    }

    callerBalance := binary.LittleEndian.Uint64(callerBalanceBytes)
    if callerBalance != fabricToken.TotalSupply {
        t.Error("Caller balance should be equal to the token total supply")
    }

    if string(tokenBytes) != string(tokenDataBytes) {
        t.Error("Expected token data to be saved in the state")
    }

    return stub
}

func TestInitToken(t *testing.T) {
    initToken(t)
}

func balance(stub *mock.FullMockStub, cn string) (Balance, error) {
    balanceRq := Balance{User: cn}
    balanceRqBytes, _ := json.Marshal(balanceRq)
    balanceRes := stub.MockInvoke("1", util.ToChaincodeArgs("balance", string(balanceRqBytes)))
    balance := Balance{}
    err := json.Unmarshal(balanceRes.Payload, &balance)
    return balance, err
}

func allAllowances(stub *mock.FullMockStub, cn string) ([]Approve, error) {
    rq := Balance{User: cn}
    rqBytes, _ := json.Marshal(rq)
    allowancesData := stub.MockInvoke("1", util.ToChaincodeArgs("allowances", string(rqBytes)))
    allowances := []Approve{}
    if allowancesData.Status != shim.OK {
        return allowances, errors.New("CC call returned error: " + allowancesData.Message)
    }
    err := json.Unmarshal(allowancesData.Payload, &allowances)
    return allowances, err
}

func TestTransfer(t *testing.T) {
    stub := initToken(t)

    stub.MockCreator("default", testdata.TestUser1Cert)
    transferData := `{"to": "testUser2", "value": 100}`
    res := stub.MockInvoke("1", util.ToChaincodeArgs("transfer", transferData))

    if res.Status != shim.OK {
        t.Errorf("Failed to transfer: %s", res.Message)
        t.FailNow()
    }

    balanceFrom, err := balance(stub, testdata.TestUser1CN)
    balanceTo, err := balance(stub, "testUser2")
    if err != nil {
        t.Error("Could not unmarshal balance")
    }

    if balanceFrom.Value != 9900 || balanceTo.Value != 100 {
        t.Error("Transfer does not work as expected")
    }
}

func TestTransferToMyself(t *testing.T) {
    stub := initToken(t)

    stub.MockCreator("default", testdata.TestUser1Cert)
    transferData := `{"to": "testUser", "value": 100}`
    res := stub.MockInvoke("1", util.ToChaincodeArgs("transfer", transferData))

    if res.Status != shim.OK {
        t.Errorf("Failed to transfer: %s", res.Message)
        t.FailNow()
    }

    balance, err := balance(stub, testdata.TestUser1CN)
    if err != nil {
        t.Error("Could not unmarshal balance")
    }

    if balance.Value != 10000 {
        t.Error("Transfer does not work as expected")
    }
}

func TestApprove(t *testing.T) {
    stub := initToken(t)

    stub.MockCreator("default", testdata.TestUser1Cert)

    approveData := `{"spender": "testUser2", "value": 100}`
    res := stub.MockInvoke("1", util.ToChaincodeArgs("approve", approveData))
    if res.Status != shim.OK {
        t.Errorf("Failed to approve: %s", res.Message)
        t.FailNow()
    }

    approveData = `{"spender": "testUser3", "value": 200}`
    res = stub.MockInvoke("1", util.ToChaincodeArgs("approve", approveData))
    if res.Status != shim.OK {
        t.Errorf("Failed to approve: %s", res.Message)
        t.FailNow()
    }

    allowances, err := allAllowances(stub, testdata.TestUser1CN)
    if err != nil {
        t.Error("Could not get allowances")
        t.FailNow()
    }

    if len(allowances) != 2 {
        t.Error("Expected 2 allowances")
    }

    if allowances[0].Spender != "testUser2" && allowances[0].Value != 100 {
        t.Error("Allowance 0 is invalid")
    }

    if allowances[1].Spender != "testUser3" && allowances[1].Value != 200 {
        t.Error("Allowance 1 is invalid")
    }

    allowances, err = allAllowances(stub, testdata.TestUser3CN)
    if err != nil {
        t.Error("Could not get allowances for user without allowances")
        t.FailNow()
    }
    if len(allowances) != 0 {
        t.Error("Expected 0 allowances if they were not added before")
    }
}

func TestTransferFrom(t *testing.T) {
    stub := initToken(t)

    stub.MockCreator("default", testdata.TestUser1Cert)
    approveData := `{"spender": "testUser2", "value": 500}`
    stub.MockInvoke("1", util.ToChaincodeArgs("approve", approveData))

    stub.MockCreator("default", testdata.TestUser2Cert)
    transferData := `{"from": "testUser", "to": "testUser3", "value": 100}`
    res := stub.MockInvoke("1", util.ToChaincodeArgs("transferFrom", transferData))
    if res.Status != shim.OK {
        t.Errorf("Failed to transfer: %s", res.Message)
        t.FailNow()
    }

    balanceFrom, err := balance(stub, testdata.TestUser1CN)
    balanceTo, err := balance(stub, testdata.TestUser3CN)
    allowances, err := allAllowances(stub, testdata.TestUser1CN)
    if err != nil {
        t.Error("Could not unmarshal balance")
    }

    if balanceFrom.Value != 9900 || balanceTo.Value != 100 || len(allowances) != 1 || allowances[0].Value != 400 {
        t.Errorf("TransferFrom does not work as expected: (%d, %d)", balanceFrom.Value, balanceTo.Value)
    }

    transferData = `{"from": "testUser", "to": "testUser3", "value": 1000}`
    res = stub.MockInvoke("1", util.ToChaincodeArgs("transferFrom", transferData))
    if res.Status == shim.OK {
        t.Error("Should fail when transfer value is too big")
        t.FailNow()
    }

    transferData = `{"from": "testUser2", "to": "testUser3", "value": 1000}`
    res = stub.MockInvoke("1", util.ToChaincodeArgs("transferFrom", transferData))
    if res.Status == shim.OK {
        t.Error("Should fail when no allowance")
        t.FailNow()
    }
}