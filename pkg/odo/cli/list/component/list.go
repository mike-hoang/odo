package component

import (
	"context"
	"errors"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"

	"github.com/redhat-developer/odo/pkg/component"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended list name
const RecommendedCommandName = "component"

var listExample = ktemplates.Examples(`  # List all components in the application
%[1]s
  `)

// ListOptions ...
type ListOptions struct {
	// Clients
	clientset *clientset.Clientset

	// Local variables
	namespaceFilter string

	// Flags
	namespaceFlag string
}

var _ genericclioptions.Runnable = (*ListOptions)(nil)
var _ genericclioptions.JsonOutputter = (*ListOptions)(nil)

// NewListOptions ...
func NewListOptions() *ListOptions {
	return &ListOptions{}
}

func (o *ListOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete ...
func (lo *ListOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	// If the namespace flag has been passed, we will search there.
	// if it hasn't, we will search from the default project / namespace.
	if lo.namespaceFlag != "" {
		if lo.clientset.KubernetesClient == nil {
			return errors.New("cluster is non accessible")
		}
		lo.namespaceFilter = lo.namespaceFlag
	} else if lo.clientset.KubernetesClient != nil {
		lo.namespaceFilter = odocontext.GetNamespace(ctx)
	}

	return nil
}

// Validate ...
func (lo *ListOptions) Validate(ctx context.Context) (err error) {
	if lo.clientset.KubernetesClient == nil {
		log.Warning("No connection to cluster defined")
	}
	return nil
}

// Run has the logic to perform the required actions as part of command
func (lo *ListOptions) Run(ctx context.Context) error {
	listSpinner := log.Spinnerf("Listing components from namespace '%s'", lo.namespaceFilter)
	defer listSpinner.End(false)

	list, err := lo.run(ctx)
	if err != nil {
		return err
	}

	listSpinner.End(true)

	HumanReadableOutput(list)
	return nil
}

// Run contains the logic for the odo command
func (lo *ListOptions) RunForJsonOutput(ctx context.Context) (out interface{}, err error) {
	return lo.run(ctx)
}

func (lo *ListOptions) run(ctx context.Context) (api.ResourcesList, error) {
	var (
		devfileObj    = odocontext.GetDevfileObj(ctx)
		componentName = odocontext.GetComponentName(ctx)
	)
	devfileComponents, componentInDevfile, err := component.ListAllComponents(
		lo.clientset.KubernetesClient, lo.namespaceFilter, devfileObj, componentName)
	if err != nil {
		return api.ResourcesList{}, err
	}
	return api.ResourcesList{
		ComponentInDevfile: componentInDevfile,
		Components:         devfileComponents,
	}, nil
}

// NewCmdList implements the list odo command
func NewCmdComponentList(name, fullName string) *cobra.Command {
	o := NewListOptions()

	var listCmd = &cobra.Command{
		Use:         name,
		Short:       "List all components in the current namespace",
		Long:        "List all components in the current namespace.",
		Example:     fmt.Sprintf(listExample, fullName),
		Args:        genericclioptions.NoArgsAndSilenceJSON,
		Annotations: map[string]string{"command": "management"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, cmd, args)
		},
		Aliases: []string{"components"},
	}
	clientset.Add(listCmd, clientset.KUBERNETES_NULLABLE, clientset.FILESYSTEM)

	listCmd.Flags().StringVar(&o.namespaceFlag, "namespace", "", "Namespace for odo to scan for components")

	commonflags.UseOutputFlag(listCmd)

	return listCmd
}

func HumanReadableOutput(list api.ResourcesList) {
	components := list.Components
	if len(components) == 0 {
		log.Error("There are no components deployed.")
		return
	}

	t := ui.NewTable()

	// Create the header and then sort accordingly
	t.AppendHeader(table.Row{"NAME", "PROJECT TYPE", "RUNNING IN", "MANAGED"})
	t.SortBy([]table.SortBy{
		{Name: "MANAGED", Mode: table.Asc},
		{Name: "NAME", Mode: table.Dsc},
	})

	// Go through each component and add it to the table
	for _, comp := range components {

		// Mark the name as yellow in the index to it's easier to see.
		name := text.Colors{text.FgHiYellow}.Sprint(comp.Name)

		// Get the managed by label
		managedBy := comp.ManagedBy
		if managedBy == "" {
			managedBy = api.TypeUnknown
		}

		// Get the mode (dev or deploy)
		mode := comp.RunningIn.String()

		// Get the type of the component
		componentType := comp.Type
		if componentType == "" {
			componentType = api.TypeUnknown
		}

		// If we find our local unpushed component, let's change the output appropriately.
		if list.ComponentInDevfile == comp.Name {
			name = fmt.Sprintf("* %s", name)
		}
		if comp.ManagedByVersion != "" {
			managedBy += fmt.Sprintf(" (%s)", comp.ManagedByVersion)
		}
		// If we are managing that component, output it as blue (our logo colour) to indicate it's used by odo
		if comp.ManagedBy == "odo" {
			managedBy = text.Colors{text.FgBlue}.Sprintf(managedBy)
		}

		t.AppendRow(table.Row{name, componentType, mode, managedBy})
	}
	t.Render()

}
