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

package main

type Token struct {
	Standard    string `json:"standard"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    uint16 `json:"decimals"`
	TotalSupply uint64 `json:"totalSupply"`
}

type Balance struct {
	User  string `json:"user"`
	Value uint64 `json:"value"`
}

type Transfer struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint64 `json:"value"`
}

type Approve struct {
	Spender string `json:"spender"`
	Value   uint64 `json:"value"`
}
