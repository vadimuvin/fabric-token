package mock

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/msp"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type FullMockStub struct {
	shim.MockStub

	cc          shim.Chaincode
	mockCreator []byte
}

func NewFullMockStub(name string, cc shim.Chaincode) *FullMockStub {
	s := shim.NewMockStub(name, cc)
	fs := new(FullMockStub)
	fs.MockStub = *s
	fs.cc = cc
	return fs
}

func (stub *FullMockStub) MockCreator(mspID string, cert string) {
	stub.mockCreator, _ = msp.NewSerializedIdentity(mspID, []byte(cert))
}

func (stub *FullMockStub) MockInit(uuid string, args [][]byte) pb.Response {
	// this is a hack here to set MockStub.args, because its not accessible otherwise
	stub.MockStub.MockInvoke(uuid, args)

	stub.MockTransactionStart(uuid)
	res := stub.cc.Init(stub)
	stub.MockTransactionEnd(uuid)

	return res
}

func (stub *FullMockStub) MockInvoke(uuid string, args [][]byte) pb.Response {
	// this is a hack here to set MockStub.args, because its not accessible otherwise
	stub.MockStub.MockInvoke(uuid, args)

	// now do the invoke with the correct stub
	stub.MockTransactionStart(uuid)
	res := stub.cc.Invoke(stub)
	stub.MockTransactionEnd(uuid)

	return res
}

func (stub *FullMockStub) GetCreator() ([]byte, error) {
	return stub.mockCreator, nil
}
