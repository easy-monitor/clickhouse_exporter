package main

type Config struct {
	Modules []*Clickmodule `yaml:"module" json:"modules"`
}

type Clickmodule struct {
	Name     string `yaml:"name" json:"name"`
	User     string `yaml:"user" json:"name"`
	Password string `yaml:"password" json:"password"`
}
