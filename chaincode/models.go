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
