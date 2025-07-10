package certificates

import (
	"fmt"
)

func buildBRImagePullCmd() string {
	return `IMAGE_ID=$(apiclient get | apiclient exec admin jq -r '.settings["host-containers"]["kubeadm-bootstrap"].source')
ctr image pull ${IMAGE_ID}`
}

func buildBRControlPlaneBackupCertsCmd(_ string, hasExternalEtcd bool, backupDir, certDir string) string {
	var script string
	if hasExternalEtcd {
		script = fmt.Sprintf(`mkdir -p '/etc/kubernetes/pki.bak_%[1]s'
cp -r %[2]s/* '/etc/kubernetes/pki.bak_%[1]s/'
rm -rf '/etc/kubernetes/pki.bak_%[1]s/etcd'`, backupDir, certDir)
	} else {
		script = fmt.Sprintf("cp -r '%s' '/etc/kubernetes/pki.bak_%s'", certDir, backupDir)
	}
	return script
}

func buildBRControlPlaneRenewCertsCmd() string {
	script := `ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs renew all`
	return script
}

func buildBRControlPlaneCheckCertsCmd() string {
	script := `ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm ${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs check-expiration`
	return script
}

func buildBRControlPlaneCopyCertsFromTmpCmd() string {
	script := `if [ -d "/run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs" ]; then
    cp /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs/apiserver-etcd-client.crt /var/lib/kubeadm/pki/server-etcd-client.crt || exit 1
    cp /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs/apiserver-etcd-client.key /var/lib/kubeadm/pki/apiserver-etcd-client.key || exit 1
    chmod 600 /var/lib/kubeadm/pki/server-etcd-client.crt || exit 1
    chmod 600 /var/lib/kubeadm/pki/apiserver-etcd-client.key || exit 1
    rm -rf /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/etcd-client-certs
fi`
	return script
}

func buildBRControlPlaneRestartPodsCmd() string {
	script := `
apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=false
sleep 10
apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=true
`
	return script
}

func buildBREtcdBackupCertsCmd(backupDir string) string {
	script := fmt.Sprintf(`cp -r /var/lib/etcd/pki /var/lib/etcd/pki.bak_%[1]s
rm /var/lib/etcd/pki/*
cp /var/lib/etcd/pki.bak_%[1]s/ca.* /var/lib/etcd/pki`, backupDir)
	return script
}

func buildBREtcdRenewCertsCmd() string {
	script := `ctr run \
--mount type=bind,src=/var/lib/etcd/pki,dst=/etc/etcd/pki,options=rbind:rw \
--net-host \
--rm \
${IMAGE_ID} tmp-cert-renew \
/opt/bin/etcdadm join phase certificates http://eks-a-etcd-dumb-url --init-system kubelet`
	return script
}

func buildBREtcdRenewChecksCmd() string {
	script := `ETCD_CONTAINER_ID=$(ctr -n k8s.io c ls | grep -w "etcd-io" | cut -d " " -f1 | tail -1)
ctr -n k8s.io t exec --exec-id etcd ${ETCD_CONTAINER_ID} etcdctl \
     --cacert=/var/lib/etcd/pki/ca.crt \
     --cert=/var/lib/etcd/pki/server.crt \
     --key=/var/lib/etcd/pki/server.key \
     member list`
	return script
}

func buildBREtcdCopyCertsToTmpCmd(tempDir string) string {
	script := fmt.Sprintf(`cp /var/lib/etcd/pki/apiserver-etcd-client.* %[1]s/ || exit 1
chmod 600 %[1]s/apiserver-etcd-client.crt || exit 1
chmod 600 %[1]s/apiserver-etcd-client.key || exit 1`, tempDir)
	return script
}

func buildBREtcdCleanupTmpCmd(tempDir string) string {
	script := fmt.Sprintf(`rm -f %s/apiserver-etcd-client.*`, tempDir)
	return script
}

func buildBRCreateTmpDirCmd(dirName string) string {
	script := fmt.Sprintf(`TARGET_DIR="/run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp/%[1]s"
mkdir -p "${TARGET_DIR}" || exit 1
chmod 755 "${TARGET_DIR}" || exit 1`, dirName)
	return script
}

func buildBRWriteCertToTmpCmd(certBase64 string) string {
	script := fmt.Sprintf(`cat <<'CRT_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.crt"
%s
CRT_END
[ $? -eq 0 ] || exit 1`, certBase64)
	return script
}

func buildBRWriteKeyToTmpCmd(keyBase64 string) string {
	script := fmt.Sprintf(`cat <<'KEY_END' | base64 -d > "${TARGET_DIR}/apiserver-etcd-client.key"
%s
KEY_END
[ $? -eq 0 ] || exit 1`, keyBase64)
	return script
}

func buildBRSetTmpCertPermissionsCmd() string {
	script := `chmod 600 "${TARGET_DIR}/apiserver-etcd-client.crt" || exit 1
chmod 600 "${TARGET_DIR}/apiserver-etcd-client.key" || exit 1`
	return script
}

func buildBRListTmpFilesCmd(tempDir string) string {
	script := fmt.Sprintf(`sudo sheltie << 'EOF'
ls -l %s/apiserver-etcd-client.*
exit
EOF`, tempDir)
	return script
}

func buildBRReadTmpCertCmd(tempDir string) string {
	script := fmt.Sprintf(`sudo sheltie << 'EOF'
cat %s/apiserver-etcd-client.crt
exit
EOF`, tempDir)
	return script
}

func buildBRReadTmpKeyCmd(tempDir string) string {
	script := fmt.Sprintf(`sudo sheltie << 'EOF'
cat %s/apiserver-etcd-client.key
exit
EOF`, tempDir)
	return script
}
