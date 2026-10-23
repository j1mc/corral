package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuna-os/corral/pkg/config"
)

var (
	contextBackend string
	contextValue   string
)

var contextCmd = &cobra.Command{
	Use:     "context",
	Aliases: []string{"ctx"},
	Short:   "Manage the local, Kubernetes, Incus, libvirt, Proxmox, VirtualBox, vSphere, Xen, and Nutanix contexts aggregated by Corral",
	Long: `A context is an inventory entry: a name, a backend, and that backend's own
address — a kubeconfig context, an Incus remote, a libvirt URI, a vCenter
host, a Prism Central VIP. The built-in "local" qemu context always exists
and cannot be removed.

Corral aggregates every enabled context at once: "context use" only picks the
default destination for commands given no backend of their own, it does not
hide or disable the rest — "context list" keeps listing them and marks the
default with *. "context remove" drops an entry from this inventory; it never
touches the VMs on that backend (deleting a VM is "corral delete").`,
	Example: `  corral context add lab --backend qemu   # a second local QEMU entry
  corral context add prod --backend kubevirt --context prod-cluster
  corral context list        # * marks the default, the others stay usable
  corral context use prod    # default destination only
  corral context get         # which destination that is right now
  corral context remove lab  # forgets the entry, deletes no VM`,
}

var contextGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Print the default destination: name, backend, and backend address",
	Example: `  corral context get
  corral context get | cut -f2   # the backend unqualified commands will use`,
	Args: cobra.NoArgs,
	Run: func(*cobra.Command, []string) {
		c := config.DefaultContext()
		fmt.Printf("%s\t%s\t%s\n", c.Name, c.Backend, c.Context)
	},
}

var contextUseCmd = &cobra.Command{
	Use:     "use <name>",
	Aliases: []string{"set", "switch"},
	Short:   "Select the default destination without hiding other contexts",
	Example: `  corral context use prod
  corral context list   # prod is marked *, every other context is still there`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return config.SetDefaultContext(args[0])
	},
}

var contextAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add or update an inventory context",
	Example: `  corral context add lab --backend qemu                              # local; qemu takes no --context
  corral context add prod --backend kubevirt --context prod-cluster  # a kubeconfig context name
  corral context add hv --backend libvirt --context qemu:///system   # same name again updates the entry`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return config.AddContext(config.ContextConfig{Name: args[0], Backend: contextBackend, Context: contextValue})
	},
}

var contextRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"delete", "rm"},
	Short:   "Forget an inventory context entry (deletes no VM)",
	Example: `  corral context remove lab   # drops the entry; VMs on that backend are untouched
  corral delete myvm          # deleting a VM is a different command`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return config.RemoveContext(args[0])
	},
}

var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List every known context, marking the default with *",
	Example: `  corral context list
  corral context list | grep '^\*'   # the default destination only`,
	Args: cobra.NoArgs,
	Run: func(*cobra.Command, []string) {
		current := config.DefaultContext().Name
		for _, c := range config.Contexts() {
			mark := " "
			if c.Name == current {
				mark = "*"
			}
			fmt.Printf("%s %-18s %-10s %s\n", mark, c.Name, c.Backend, c.Context)
		}
	},
}

func init() {
	contextAddCmd.Flags().StringVarP(&contextBackend, "backend", "b", "", "Backend: qemu, kubevirt, incus, libvirt, proxmox, virtualbox, vsphere, xen, or nutanix")
	contextAddCmd.Flags().StringVar(&contextValue, "context", "", "kubeconfig context, Incus remote, libvirt URI, VirtualBox context label, vCenter host, Xen pool master, or Prism Central VIP")
	_ = contextAddCmd.MarkFlagRequired("backend")
	contextCmd.AddCommand(contextGetCmd, contextUseCmd, contextAddCmd, contextRemoveCmd, contextListCmd)
	rootCmd.AddCommand(contextCmd)
}
