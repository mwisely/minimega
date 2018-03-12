// Copyright (2012) Sandia Corporation.
// Under the terms of Contract DE-AC04-94AL85000 with Sandia Corporation,
// the U.S. Government retains certain rights in this software.

package main

import (
	"errors"
	"fmt"
	"io"
	log "minilog"
	"strconv"
	"strings"
	"unicode"
)

// NetConfig contains all the network-related config for an interface. The IP
// addresses are automagically populated by snooping ARP traffic. The bandwidth
// stats and IP addresses are updated on-demand by calling the UpdateNetworks
// function of BaseConfig.
type NetConfig struct {
	Alias  string
	VLAN   int
	Bridge string
	Tap    string
	MAC    string
	Driver string
	IP4    string
	IP6    string

	RxRate, TxRate float64 // Most recent bandwidth measurements for Tap

	// Raw string that we used when creating this network config will be
	// reparsed if we ever clone the VM that has this config.
	Raw string
}

type NetConfigs []NetConfig

func NewVMConfig() VMConfig {
	c := VMConfig{}
	c.Clear(Wildcard)
	return c
}

// ParseNetConfig processes the input specifying the bridge, VLAN alias, and
// mac for one interface to a VM and updates the vm config accordingly. The
// VLAN alias must be resolved using the active namespace. This takes a bit of
// parsing, because the entry can be in a few forms:
//
// 	vlan alias
//
//	vlan alias,mac
//	bridge,vlan alias
//	vlan alias,driver
//
//	bridge,vlan alias,mac
//	vlan alias,mac,driver
//	bridge,vlan alias,driver
//
//	bridge,vlan alias,mac,driver
//
// If there are 2 or 3 fields, just the last field for the presence of a mac
func ParseNetConfig(spec string) (res NetConfig, err error) {
	// example: my_bridge,100,00:00:00:00:00:00
	f := strings.Split(spec, ",")

	var b, v, m, d string
	switch len(f) {
	case 1:
		v = f[0]
	case 2:
		if isMac(f[1]) {
			// vlan, mac
			v, m = f[0], f[1]
		} else if isNetworkDriver(f[1]) {
			// vlan, driver
			v, d = f[0], f[1]
		} else {
			// bridge, vlan
			b, v = f[0], f[1]
		}
	case 3:
		if isMac(f[2]) {
			// bridge, vlan, mac
			b, v, m = f[0], f[1], f[2]
		} else if isMac(f[1]) {
			// vlan, mac, driver
			v, m, d = f[0], f[1], f[2]
		} else {
			// bridge, vlan, driver
			b, v, d = f[0], f[1], f[2]
		}
	case 4:
		b, v, m, d = f[0], f[1], f[2], f[3]
	default:
		return NetConfig{}, errors.New("malformed netspec")
	}

	if d != "" && !isNetworkDriver(d) {
		return NetConfig{}, errors.New("malformed netspec, invalid driver: " + d)
	}

	log.Info(`got bridge="%v", alias="%v", mac="%v", driver="%v"`, b, v, m, d)

	if m != "" && !isMac(m) {
		return NetConfig{}, errors.New("malformed netspec, invalid mac address: " + m)
	}

	// warn on valid but not allocated macs
	if m != "" && !allocatedMac(m) {
		log.Warn("unallocated mac address: %v", m)
	}

	if b == "" {
		b = DefaultBridge
	}
	if d == "" {
		d = DefaultKVMDriver
	}

	return NetConfig{
		Alias:  v,
		Bridge: b,
		MAC:    strings.ToLower(m),
		Driver: d,
	}, nil
}

// String representation of NetConfig, should be able to parse back into a
// NetConfig.
func (c NetConfig) String() string {
	parts := []string{}

	prep := func(s string) string {
		if strings.IndexFunc(s, unicode.IsSpace) > -1 {
			return strconv.Quote(s)
		}

		return s
	}

	if c.Bridge != "" && c.Bridge != DefaultBridge {
		parts = append(parts, prep(c.Bridge))
	}

	parts = append(parts, prep(c.Alias))

	if c.MAC != "" {
		// shouldn't need to prep MAC since it is a valid MAC
		parts = append(parts, c.MAC)
	}

	if c.Driver != "" && c.Driver != DefaultKVMDriver {
		parts = append(parts, prep(c.Driver))
	}

	return strings.Join(parts, ",")
}

func (c NetConfigs) String() string {
	parts := []string{}
	for _, n := range c {
		parts = append(parts, n.String())
	}

	return strings.Join(parts, " ")
}

func (c NetConfigs) WriteConfig(w io.Writer) error {
	if len(c) > 0 {
		_, err := fmt.Fprintf(w, "vm config networks %v\n", c)
		return err
	}

	return nil
}
