package cmd

// import (
// 	"context"
// 	"fmt"
// 	"log"

// 	"github.com/spf13/cobra"
// 	"github.com/spf13/pflag"
// 	"github.com/spf13/viper"

// 	"github.com/aws/eks-anywhere/pkg/config"
// 	"github.com/aws/eks-anywhere/pkg/dependencies"
// 	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
// )

// func init() {
// 	rootCmd.AddCommand(vSphereSetupUserCmd)
// 	vSphereSetupUserCmd.Flags().StringVarP(&lio.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
// 	err := vSphereSetupUserCmd.MarkFlagRequired("filename")
// 	if err != nil {
// 		log.Fatalf("Error marking filename flag as required: %v", err)
// 	}
// }

// var vSphereSetupUserCmd = &cobra.Command{
// 	Use:   "vsphere-user-setup",
// 	Short: "Foo",
// 	Long:  "Bar",
// 	PreRunE: func(cmd *cobra.Command, args []string) error {
// 		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
// 			if err := viper.BindPFlag(flag.Name, flag); err != nil {
// 				log.Fatalf("Error initializing flags: %v", err)
// 			}
// 		})
// 		return nil
// 	},
// 	SilenceUsage: true,
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		return vSphereSetupUser(cmd.Context(), lio.fileName)
// 	},
// }

// func vSphereSetupUser(ctx context.Context, spec string) error {
// 	config := config.VSphereSetupUserConfig{
// 		Domain:          "vsphere.local",
// 		Username:        "jwmeiertest004",
// 		GroupName:       "JwmeierTestGroup4",
// 		GlobalRole:      "JwmeierEKSAGlobalTest004",
// 		UserRole:        "JwmeierEKSAUserTest004",
// 		CloudAdmin:      "JwmeierEKSACloudAdminTest004",
// 		Datastore:       "/Datacenter/datastore/datastore1",
// 		ResourcePool:    "/Datacenter/host/Cluster-01/Resources/TestResourcePool",
// 		Network:         "/Datacenter/network/VM Network",
// 		VirtualMachines: "/Datacenter/vm/jwmeier/permissiontest",
// 		Templates:       "/Datacenter/vm/Templates/",
// 	}

// 	deps, err := dependencies.NewFactory().WithGovc().Build(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	defer close(ctx, deps)

// 	err = vsphere.ConfigureVSphere(ctx, deps.Govc, &config)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("Done!")

// 	return nil
// }
