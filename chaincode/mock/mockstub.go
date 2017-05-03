/*
Copyright Vadim Uvin (Swisscom AG). 2017 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
