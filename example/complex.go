package main

type ComplexStruct struct {
	Name    string   `json:"name"`
	Parents []Parent `json:"parents"`
}

type Parent struct {
	Name     string   `json:"name"`
	Children []*Child `json:"children"`
}

type Child struct {
	Name string `json:"name"`
}
