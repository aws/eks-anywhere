#!/bin/bash

set -euo pipefail
trap 'echo "❌ Script failed at line $LINENO: $BASH_COMMAND"; exit 1' ERR

# Constants
readonly ETCD_CERT_DIR="/etc/etcd"
readonly BACKUP_DATE=$(date '+%Y%m%d_%H%M%S')
readonly BACKUP_DIR="${ETCD_CERT_DIR}/pki.bak_${BACKUP_DATE}"
readonly CONTROL_PLANE_CERT_DIR="/etc/kubernetes/pki"
readonly TEMP_LOCAL_ETCD_CERTS_DIR="etcd-client-certs"
readonly CONTROL_PLANE_MANIFESTS_DIR="/etc/kubernetes/manifests"
readonly SSH_OPTS="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
readonly BACKUP_FOLDER="./certificate_backup_${BACKUP_DATE}"
readonly KUBEADM_CONFIG_BACKUP="${BACKUP_FOLDER}/kubeadm-config.yaml"

# Global variables for cluster access
declare cluster_name
declare ssh_user
declare ssh_key

# Input validation
function validate_inputs() {
    if [[ $# -ne 3 ]]; then
        echo "Usage: $0 <cluster-name> <ssh-user> <path-to-ssh-key>" >&2
        exit 1
    fi
    
    # Set global variables after validation
    cluster_name="$1"
    ssh_user="$2"
    ssh_key="$3"
}

function check_sudo_access() {
    if ! sudo -n true 2>/dev/null; then
        echo "❌ Error: This script requires sudo access. Please run with a user that has sudo privileges." >&2
        exit 1
    fi
}


# Node retrieval functions
function get_etcd_nodes() {
    echo "Retrieving etcd node IPs for cluster: ${cluster_name}..."

    ETCD_NODES=($(kubectl -n eksa-system get machines \
        --selector "cluster.x-k8s.io/cluster-name=${cluster_name},cluster.x-k8s.io/etcd-cluster=${cluster_name}-etcd" \
        -ojsonpath='{.items[*].status.addresses[?(@.type=="ExternalIP")].address}'))
}

function get_control_plane_nodes() {
    echo "Retrieving control plane node IPs for cluster: ${cluster_name}..."

    CONTROL_PLANE_NODES=($(kubectl -n eksa-system get machines \
                            --selector "cluster.x-k8s.io/cluster-name=${cluster_name},cluster.x-k8s.io/control-plane" \
                            -o json | jq -r '.items[].status.addresses | map(select(.type=="ExternalIP"))[0].address'))

}

# Certificate management functions
function backup_etcd_certs() {
    cat <<EOF
# Backup certificates
cd ${ETCD_CERT_DIR}
sudo cp -r pki pki.bak_${BACKUP_DATE}
sudo rm -rf pki/*
sudo cp pki.bak_${BACKUP_DATE}/ca.* pki/
EOF
}

function renew_etcd_certs() {
    cat <<EOF
# Renew certificates
sudo etcdadm join phase certificates http://eks-a-etcd-dumb-url
EOF
}

function validate_etcd_renewal() {
    cat <<EOF
# Validate certificates
sudo etcdctl --cacert=${ETCD_CERT_DIR}/pki/ca.crt \
             --cert=${ETCD_CERT_DIR}/pki/etcdctl-etcd-client.crt \
             --key=${ETCD_CERT_DIR}/pki/etcdctl-etcd-client.key \
             endpoint health
EOF
}

function process_etcd_node() {
    local node_ip="$1"

    echo "Processing etcd node: ${node_ip}..."
    
    ssh ${SSH_OPTS} -i "${ssh_key}" "${ssh_user}@${node_ip}" bash <<EOF
set -euo pipefail
$(backup_etcd_certs)
$(renew_etcd_certs)

sudo cp /etc/etcd/pki/apiserver-etcd-client.key /home/${ssh_user}/
sudo chown ${ssh_user}:${ssh_user} /home/${ssh_user}/apiserver-etcd-client.key
$(validate_etcd_renewal)
EOF
    
    scp ${SSH_OPTS} -i "${ssh_key}" "${ssh_user}@${node_ip}:${ETCD_CERT_DIR}/pki/apiserver-etcd-client.crt" "${BACKUP_FOLDER}/${TEMP_LOCAL_ETCD_CERTS_DIR}/" || exit 1
    scp ${SSH_OPTS} -i "${ssh_key}" "${ssh_user}@${node_ip}:/home/${ssh_user}/apiserver-etcd-client.key" "${BACKUP_FOLDER}/${TEMP_LOCAL_ETCD_CERTS_DIR}/" || exit 1

    ssh ${SSH_OPTS} -i "${ssh_key}" "${ssh_user}@${node_ip}" "rm -f /home/${ssh_user}/apiserver-etcd-client.key"

    echo "✅ Completed renewing certificate for the ETCD node: ${node_ip}."
    echo "---------------------------------------------"
}

function update_apiserver_etcd_client_secret() {
    local base64_cmd

    if [[ "$OSTYPE" == "darwin"* ]]; then
        base64_cmd="base64 | tr -d '\n'"
    else
        base64_cmd="base64 -w 0"
    fi

    local crt_base64
    local key_base64
    crt_base64=$(cat "${BACKUP_FOLDER}/${TEMP_LOCAL_ETCD_CERTS_DIR}/apiserver-etcd-client.crt" | eval "${base64_cmd}")
    key_base64=$(cat "${BACKUP_FOLDER}/${TEMP_LOCAL_ETCD_CERTS_DIR}/apiserver-etcd-client.key" | eval "${base64_cmd}")

    kubectl patch secret "${cluster_name}-apiserver-etcd-client" -n eksa-system --type='merge' -p="
data:
  tls.crt: ${crt_base64}
  tls.key: ${key_base64}
"
    echo "✅ Successfully updated ${cluster_name}-apiserver-etcd-client secret."
}

function transfer_certs_to_control_plane() {
    local node_ip="$1"

    echo "Transferring apiserver-etcd-client certificates to control plane node: ${node_ip}..."
    sudo scp ${SSH_OPTS} -i "${ssh_key}" -r "${BACKUP_FOLDER}/${TEMP_LOCAL_ETCD_CERTS_DIR}" "${ssh_user}@${node_ip}:."
    echo "External certificates transferred to control plane node: ${node_ip}."
}

function process_control_plane_node() {
    local node_ip="$1"

    echo "Processing control plane node: ${node_ip}..."
    
    ssh ${SSH_OPTS} -i "${ssh_key}" "${ssh_user}@${node_ip}" bash <<EOF
set -euo pipefail

sudo cp -r '${CONTROL_PLANE_CERT_DIR}' '/etc/kubernetes/pki.bak_${BACKUP_DATE}'

sudo kubeadm certs renew all
sudo kubeadm certs check-expiration

# Only copy etcd client certificates if external etcd exists
if [[ -d "${TEMP_LOCAL_ETCD_CERTS_DIR}" ]]; then
  sudo cp '${TEMP_LOCAL_ETCD_CERTS_DIR}/apiserver-etcd-client.crt' '${CONTROL_PLANE_CERT_DIR}'
  sudo cp '${TEMP_LOCAL_ETCD_CERTS_DIR}/apiserver-etcd-client.key' '${CONTROL_PLANE_CERT_DIR}'
  rm -rf ${TEMP_LOCAL_ETCD_CERTS_DIR}
fi

sudo mkdir -p /tmp/manifests
sudo mv ${CONTROL_PLANE_MANIFESTS_DIR}/* /tmp/manifests/
sleep 20
sudo mv /tmp/manifests/* ${CONTROL_PLANE_MANIFESTS_DIR}/

EOF

    echo "✅ Completed renewing certificate for the control node: ${node_ip}."
    echo "---------------------------------------------"
}

function check_api_server_reachability() {
    echo "✅ Checking if Kubernetes API server is reachable..."
    for i in {1..5}; do
        kubectl version --request-timeout=2m &>/dev/null && return 0
        sleep 10
    done

    echo "❌ Error: Kubernetes API server is not reachable. Aborting." >&2
    exit 1
}

function backup_kubeadm_config() {
    mkdir -p "${BACKUP_FOLDER}"
    echo "✅ Backing up kubeadm-config ConfigMap..."

    if ! kubectl -n kube-system get cm kubeadm-config -o yaml > "${KUBEADM_CONFIG_BACKUP}"; then
        echo "❌ Failed to backup kubeadm-config." >&2
        exit 1
    fi

    echo "✅ kubeadm-config backed up to ${KUBEADM_CONFIG_BACKUP}"
}

function cleanup_on_success() {
    if check_api_server_reachability; then
        echo "✅ Cleaning up temporary files..."
        rm -rf "${BACKUP_FOLDER}"
        echo "✅ All temporary files removed."
    else
        echo "❌ API server unreachable — skipping cleanup to preserve debug data." >&2
    fi
}

function main() {
    validate_inputs "$@"
    check_sudo_access
    check_api_server_reachability
    backup_kubeadm_config

    # ETCD cert renewal
    echo "Starting etcd certificate renewal process..."
    get_etcd_nodes
    mkdir -p "${BACKUP_FOLDER}"
    mkdir -p "${BACKUP_FOLDER}/${TEMP_LOCAL_ETCD_CERTS_DIR}"
    
    if [[ ${#ETCD_NODES[@]} -eq 0 ]]; then
        echo "Cluster ${cluster_name} does not have external ETCD." >&2
    else
        for node_ip in "${ETCD_NODES[@]}"; do
            process_etcd_node "${node_ip}"
        done

        update_apiserver_etcd_client_secret
        echo "🎉 Etcd certificate renewal process completed successfully."
    fi

    # Control plane cert renewal
    echo "Starting control plane certificate renewal process..."
    get_control_plane_nodes

    if [[ ${#CONTROL_PLANE_NODES[@]} -eq 0 ]]; then
        echo "❌ Error: No control plane node IPs found for cluster: ${cluster_name}." >&2
        exit 1
    else
        for node_ip in "${CONTROL_PLANE_NODES[@]}"; do
            if [[ ${#ETCD_NODES[@]} -ne 0 ]]; then
                transfer_certs_to_control_plane "${node_ip}"
            fi
            check_api_server_reachability
            process_control_plane_node "${node_ip}"
        done
    fi

    echo "🎉 Control plane certificate renewal process completed successfully."

    cleanup_on_success
}

main "$@"
