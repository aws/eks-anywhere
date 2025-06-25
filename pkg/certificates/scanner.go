package certificates

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	clusterNameLabel  = "cluster.x-k8s.io/cluster-name"
	controlPlaneLabel = "cluster.x-k8s.io/control-plane"
	externalEtcdLabel = "cluster.x-k8s.io/etcd-cluster"
)

// CertificateScanner defines the interface for checking certificate expiration.
type CertificateScanner interface {
	CheckCertificateExpiry(ctx context.Context, cluster *anywherev1.Cluster) ([]anywherev1.ClusterCertificateInfo, error)
	UpdateClusterCertificateStatus(ctx context.Context, cluster *anywherev1.Cluster) error
}

// Scanner implements the CertificateScanner interface and provides certificate checking functionality.
type Scanner struct {
	client client.Client
	logger logr.Logger
}

// NewCertificateScanner creates a new certificate service.
func NewCertificateScanner(client client.Client, logger logr.Logger) *Scanner {
	return &Scanner{
		client: client,
		logger: logger,
	}
}

// MachineInfo holds machine name and IP information.
type MachineInfo struct {
	Name string
	IP   string
}

// CheckCertificateExpiry checks the certificate expiration for control plane and etcd machines.
func (s *Scanner) CheckCertificateExpiry(ctx context.Context, cluster *anywherev1.Cluster) ([]anywherev1.ClusterCertificateInfo, error) {
	var certificateInfos []anywherev1.ClusterCertificateInfo

	if cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachines, err := s.getEtcdMachines(ctx, cluster)
		if err != nil {
			return nil, fmt.Errorf("getting etcd machines: %w", err)
		}

		etcdCertInfos := s.getMachinesCertificateInfo(etcdMachines, "2379")
		certificateInfos = append(certificateInfos, etcdCertInfos...)
	}

	controlPlaneMachines, err := s.getControlPlaneMachines(ctx, cluster)
	if err != nil {
		return nil, fmt.Errorf("getting control plane machines: %w", err)
	}

	controlPlaneCertInfos := s.getMachinesCertificateInfo(controlPlaneMachines, "6443")
	certificateInfos = append(certificateInfos, controlPlaneCertInfos...)

	return certificateInfos, nil
}

func (s *Scanner) getControlPlaneMachines(ctx context.Context, cluster *anywherev1.Cluster) ([]MachineInfo, error) {
	var machines []MachineInfo

	machineList := &clusterv1.MachineList{}
	selector := client.MatchingLabels{
		clusterNameLabel:  cluster.Name,
		controlPlaneLabel: "",
	}

	if err := s.client.List(ctx, machineList, selector); err != nil {
		return nil, fmt.Errorf("listing control plane machines: %w", err)
	}

	for _, machine := range machineList.Items {
		for _, address := range machine.Status.Addresses {
			if address.Type == clusterv1.MachineExternalIP {
				machines = append(machines, MachineInfo{
					Name: machine.Name,
					IP:   address.Address,
				})
				break
			}
		}
	}

	return machines, nil
}

func (s *Scanner) getEtcdMachines(ctx context.Context, cluster *anywherev1.Cluster) ([]MachineInfo, error) {
	var machines []MachineInfo

	machineList := &clusterv1.MachineList{}
	selector := client.MatchingLabels{
		clusterNameLabel:  cluster.Name,
		externalEtcdLabel: cluster.Name + "-etcd",
	}

	if err := s.client.List(ctx, machineList, selector); err != nil {
		return nil, fmt.Errorf("listing etcd machines: %w", err)
	}

	for _, machine := range machineList.Items {
		for _, address := range machine.Status.Addresses {
			if address.Type == clusterv1.MachineExternalIP {
				machines = append(machines, MachineInfo{
					Name: machine.Name,
					IP:   address.Address,
				})
				break
			}
		}
	}

	return machines, nil
}

func (s *Scanner) getMachinesCertificateInfo(machines []MachineInfo, port string) []anywherev1.ClusterCertificateInfo {
	var certificateInfos []anywherev1.ClusterCertificateInfo

	for _, machine := range machines {
		expiryDays, err := s.checkMachineCertificateExpiry(machine.IP, port)
		if err != nil {
			s.logger.Error(err, "checking certificate expiry", "machine", machine.Name, "ip", machine.IP, "port", port)
			continue
		}

		certificateInfos = append(certificateInfos, anywherev1.ClusterCertificateInfo{
			Machine:       machine.Name,
			ExpiresInDays: expiryDays,
		})
	}

	return certificateInfos
}

func (s *Scanner) checkMachineCertificateExpiry(ip, port string) (int, error) {
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(ip, port), &tls.Config{
		InsecureSkipVerify: true, // We just want to get the certificate, not verify it
	})
	if err != nil {
		return 0, fmt.Errorf("connecting to %s:%s: %w", ip, port, err)
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return 0, fmt.Errorf("no certificates found for %s:%s", ip, port)
	}

	// Use the first certificate (leaf certificate)
	cert := certs[0]

	now := time.Now()
	daysUntilExpiry := int(cert.NotAfter.Sub(now).Hours() / 24)

	return daysUntilExpiry, nil
}

// UpdateClusterCertificateStatus updates the cluster status with certificate information.
func (s *Scanner) UpdateClusterCertificateStatus(ctx context.Context, cluster *anywherev1.Cluster) error {
	certificateInfo, err := s.CheckCertificateExpiry(ctx, cluster)
	if err != nil {
		return fmt.Errorf("checking certificate expiry: %w", err)
	}

	cluster.Status.ClusterCertificateInfo = certificateInfo

	return nil
}
