package main

import (
)

type HAConfig struct {
	Configs  Config	
}

type Config struct {
	ProxyList  []Proxy	
}

type Proxy struct {
	Src string
	SrcPort string
	Mode string
	Type string
	KeepAlive int
	CheckTime int
	Counter int
	DstLen int
	Index int
	DstList []DstConfig
}

type DstConfig struct {
	Name string
	Dst string
	DstPort string
	Weight int
	Check bool
	Health bool
}

func (p *Proxy) GetSrcAddr() string {
	return p.Src + ":" + p.SrcPort
}

func (p *Proxy) GetDstAddr(index int) string {
	return p.DstList[index].Dst + ":" + p.DstList[index].DstPort
}