package init_app

import "fmt"

func printIndent(prefix, value string) {
	fmt.Printf("%s%s\n", prefix, value)
}

func formatPtr[T any](p *T) string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", *p)
}

func PrintConfig(c Config) {
	indentCfg := "Config:"
	printIndent("", indentCfg)

	printIndent("  Bridge:      ", formatPtr(c.Bridge))
	printIndent("  BridgeIP:    ", formatPtr(c.BridgeIP))
	printIndent("  Network:     ", formatPtr(c.Network))

	printIndent("  Nameservers:", "")
	printIndent("    NS1:      ", c.Nameservers.NS1)
	printIndent("    NS2:      ", c.Nameservers.NS2)

	for ti, tmpl := range c.Templates {
		printIndent(fmt.Sprintf("  Template #%d:", ti), "")
		printIndent("    CPU:           ", formatPtr(tmpl.CPU))
		printIndent("    SMT:           ", formatPtr(tmpl.SMT))
		printIndent("    Multiplier:    ", formatPtr(tmpl.Multiplier))
		printIndent("    RAM:           ", formatPtr(tmpl.RAM))
		printIndent("    EnableOverlay: ", formatPtr(tmpl.EnableOverlay))
		printIndent("    OverlaySize:   ", formatPtr(tmpl.OverlaySize))
		printIndent("    IsZFS:         ", formatPtr(tmpl.IsZFS))
		printIndent("    ZFSSnapshot:   ", tmpl.ZFSSnapshot)
		printIndent("    ZFSClonePath:  ", tmpl.ZFSClonePath)
		printIndent("    KernelArgs:    ", tmpl.KernelArgs)
		printIndent("    Kernel:        ", tmpl.Kernel)
		printIndent("    RootFS:        ", tmpl.RootFS)
		printIndent("    Username:      ", formatPtr(tmpl.Username))
		printIndent("    EnableIDE:     ", formatPtr(tmpl.EnableIDE))
		printIndent("    IDEPort:       ", formatPtr(tmpl.IDEPort))

		for ni, net := range tmpl.Network {
			printIndent(fmt.Sprintf("    Network #%d:", ni), "")
			printIndent("      IsPrimary:   ", fmt.Sprintf("%v", net.IsPrimary))
			printIndent("      Name:        ", net.Name)
			printIndent("      TAP:         ", net.TAP)
			printIndent("      Bridge:      ", formatPtr(net.Bridge))
			printIndent("      IP:          ", formatPtr(net.IP))
			printIndent("      Mask:        ", fmt.Sprintf("%d", net.Mask))
			printIndent("      Gateway:     ", net.Gateway)
			printIndent("      MAC:         ", net.MAC)
			printIndent("      Nameservers:", "")
			printIndent("        NS1:      ", net.Nameservers.NS1)
			printIndent("        NS2:      ", net.Nameservers.NS2)
		}
	}
}
