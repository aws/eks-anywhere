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

# Global variables for cluster access
declare cluster_name
declare etcd_ssh_user
declare etcd_ssh_key
declare cp_ssh_user
declare cp_ssh_key
declare ETCD_NODES
declare CONTROL_PLANE_NODES

# Function to display usage information
function show_usage() {
    echo "Usage: $0 -f <config-file-path>" >&2
    echo "" >&2
    echo "Options:" >&2
    echo "  -f <config-file-path>  Path to the YAML configuration file (required)" >&2
    exit 1
}

# Function to check if a command exists
function command_exists() {
    command -v "$1" &> /dev/null
}

# Function to check if yq is installed
function check_yq() {
    if ! command_exists yq; then
        echo "❌ Error: 'yq' is required but not installed." >&2
        echo "Please install yq using one of the following methods:" >&2
        echo "  - Homebrew (macOS): brew install yq" >&2
        echo "  - Go: go install github.com/mikefarah/yq/v4@latest" >&2
        echo "  - Download binary from: https://github.com/mikefarah/yq/releases" >&2
        exit 1
    fi
}

# Input validation
function validate_inputs() {
    local config_file=""
    
    # Parse command line options
    while getopts ":f:" opt; do
        case ${opt} in
            f)
                config_file=$OPTARG
                ;;
            \?)
                echo "❌ Error: Invalid option: -$OPTARG" >&2
                show_usage
                ;;
            :)
                echo "❌ Error: Option -$OPTARG requires an argument." >&2
                show_usage
                ;;
        esac
    done
    
    # Check if config file is provided
    if [[ -z "$config_file" ]]; then
        echo "❌ Error: Config file (-f) is required." >&2
        show_usage
    fi
    
    # Check if config file exists
    if [[ ! -f "$config_file" ]]; then
        echo "❌ Error: Config file '$config_file' does not exist." >&2
        exit 1
    fi
    
    # Check if yq is installed
    check_yq
    
    # Parse config file
    parse_config_file "$config_file"
    
    # Validate parsed configuration
    validate_config
}

# Function to parse the config file
function parse_config_file() {
    local config_file="$1"
    
    # Extract cluster name
    cluster_name=$(yq eval '.clusterName' "$config_file")
    
    # Extract etcd nodes
    ETCD_NODES=$(yq eval '.etcd.nodes[]' "$config_file" 2>/dev/null | tr '\n' ' ' || echo "")
    
    # Extract etcd SSH credentials
    etcd_ssh_user=$(yq eval '.etcd.sshUser' "$config_file" 2>/dev/null || echo "")
    etcd_ssh_key=$(yq eval '.etcd.sshKey' "$config_file" 2>/dev/null || echo "")

    
    # Extract control plane nodes
    CONTROL_PLANE_NODES=$(yq eval '.controlPlane.nodes[]' "$config_file" | tr '\n' ' ')
    
    # Extract control plane SSH credentials
    cp_ssh_user=$(yq eval '.controlPlane.sshUser' "$config_file" 2>/dev/null || echo "")
    cp_ssh_key=$(yq eval '.controlPlane.sshKey' "$config_file" 2>/dev/null || echo "")
}

# Function to validate the parsed configuration
function validate_config() {
    # Check if cluster name is provided
    if [[ -z "$cluster_name" ]]; then
        echo "❌ Error: Cluster name is missing in the config file." >&2
        exit 1
    fi
    
    # Check if etcd SSH credentials are provided only if etcd nodes exist
    if [[ -n "$ETCD_NODES" ]]; then
        if [[ -z "$etcd_ssh_user" ]]; then
            echo "❌ Error: etcd SSH user is missing in the config file." >&2
            exit 1
        fi
        
        if [[ -z "$etcd_ssh_key" ]]; then
            echo "❌ Error: etcd SSH key is missing in the config file." >&2
            exit 1
        fi

        # Check if etcd SSH key file exists
        if [[ ! -f "$etcd_ssh_key" ]]; then
            echo "❌ Error: etcd SSH key file '$etcd_ssh_key' does not exist." >&2
            exit 1
        fi
    fi
   
    
    # Check if control plane SSH credentials are provided
    if [[ -z "$cp_ssh_user" ]]; then
        echo "❌ Error: Control plane SSH user is missing in the config file." >&2
        exit 1
    fi
    
    if [[ -z "$cp_ssh_key" ]]; then
        echo "❌ Error: Control plane SSH key is missing in the config file." >&2
        exit 1
    fi
    
    # Check if control plane SSH key file exists
    if [[ ! -f "$cp_ssh_key" ]]; then
        echo "❌ Error: Control plane SSH key file '$cp_ssh_key' does not exist." >&2
        exit 1
    fi
    
    # Check if at least one control plane node is provided
    if [[ -z "$CONTROL_PLANE_NODES" ]]; then
        echo "❌ Error: No control plane nodes specified in the config file." >&2
        exit 1
    fi
}

function check_sudo_access() {
    if ! sudo -n true 2>/dev/null; then
        echo "❌ Error: This script requires sudo access. Please run with a user that has sudo privileges." >&2
        exit 1
    fi
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
    
    ssh ${SSH_OPTS} -i "${etcd_ssh_key}" "${etcd_ssh_user}@${node_ip}" bash <<EOF
set -euo pipefail
$(backup_etcd_certs)
$(renew_etcd_certs)

sudo cp /etc/etcd/pki/apiserver-etcd-client.key /home/${etcd_ssh_user}/
sudo chown ${etcd_ssh_user}:${etcd_ssh_user} /home/${etcd_ssh_user}/apiserver-etcd-client.key
$(validate_etcd_renewal)
EOF
    
    scp ${SSH_OPTS} -i "${etcd_ssh_key}" "${etcd_ssh_user}@${node_ip}:${ETCD_CERT_DIR}/pki/apiserver-etcd-client.crt" "${BACKUP_FOLDER}/${TEMP_LOCAL_ETCD_CERTS_DIR}/" || exit 1
    scp ${SSH_OPTS} -i "${etcd_ssh_key}" "${etcd_ssh_user}@${node_ip}:/home/${etcd_ssh_user}/apiserver-etcd-client.key" "${BACKUP_FOLDER}/${TEMP_LOCAL_ETCD_CERTS_DIR}/" || exit 1

    ssh ${SSH_OPTS} -i "${etcd_ssh_key}" "${etcd_ssh_user}@${node_ip}" "rm -f /home/${etcd_ssh_user}/apiserver-etcd-client.key"

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
    echo "✅ Successfully patched ${cluster_name}-apiserver-etcd-client secret."
}

function transfer_certs_to_control_plane() {
    local node_ip="$1"

    echo "Transferring apiserver-etcd-client certificates to control plane node: ${node_ip}..."
    sudo scp ${SSH_OPTS} -i "${cp_ssh_key}" -r "${BACKUP_FOLDER}/${TEMP_LOCAL_ETCD_CERTS_DIR}" "${cp_ssh_user}@${node_ip}:."
    echo "External certificates transferred to control plane node: ${node_ip}."
}

function process_control_plane_node() {
    local node_ip="$1"

    echo "Processing control plane node: ${node_ip}..."
    
    ssh ${SSH_OPTS} -i "${cp_ssh_key}" "${cp_ssh_user}@${node_ip}" bash <<EOF
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
    echo "Checking if Kubernetes API server is reachable..."
    for i in {1..5}; do
        kubectl version --request-timeout=2m &>/dev/null && return 0
        sleep 10
    done
    return 1
}

function cleanup_on_success() {
    echo "Cleaning up temporary files..."
    rm -rf "${BACKUP_FOLDER}"
    echo "All temporary files removed."
}

function main() {
    validate_inputs "$@"
    check_sudo_access

    # ETCD cert renewal
    echo "Starting etcd certificate renewal process..."
    mkdir -p "${BACKUP_FOLDER}"
    mkdir -p "${BACKUP_FOLDER}/${TEMP_LOCAL_ETCD_CERTS_DIR}"
    
    if [[ -z "$ETCD_NODES" ]]; then
        echo "Cluster ${cluster_name} does not have external ETCD." >&2
    else
        for node_ip in ${ETCD_NODES}; do
            process_etcd_node "${node_ip}"
        done

        echo "🎉 Etcd certificate renewal process completed successfully."
    fi

    # Control plane cert renewal
    echo "Starting control plane certificate renewal process..."

    if [[ -z "$CONTROL_PLANE_NODES" ]]; then
        echo "❌ Error: No control plane node IPs found for cluster: ${cluster_name}." >&2
        exit 1
    else
        for node_ip in ${CONTROL_PLANE_NODES}; do
            if [[ -n "$ETCD_NODES" ]]; then
                transfer_certs_to_control_plane "${node_ip}"
            fi
            process_control_plane_node "${node_ip}"
        done
    fi

    echo "🎉 Control plane certificate renewal process completed successfully."

    if [[ -n "$ETCD_NODES" ]]; then
        if check_api_server_reachability; then
            update_apiserver_etcd_client_secret
        else
            echo "❌ API server unreachable — could not patch ${cluster_name}-apiserver-etcd-client secret. Please patch it manually!"
        fi
    fi

    cleanup_on_success
}

main "$@"
