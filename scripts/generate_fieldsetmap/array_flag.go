package main

import "flag"

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func ArrayFlag(name string, initial []string, desc string) *arrayFlags {
	vals := &arrayFlags{}
	for _, initialVal := range initial {
		vals.Set(initialVal)
	}
	flag.Var(vals, name, desc)
	return vals
}
