locals {
  cluster_name = "ha-gce.example.com"
  project      = "testproject"
  region       = "us-test1"
}

output "cluster_name" {
  value = "ha-gce.example.com"
}

output "project" {
  value = "testproject"
}

output "region" {
  value = "us-test1"
}

provider "google" {
  project = "testproject"
  region  = "us-test1"
}

provider "aws" {
  alias  = "files"
  region = "us-test-1"
}

resource "aws_s3_object" "cluster-completed-spec" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_cluster-completed.spec_content")
  key                    = "tests/ha-gce.example.com/cluster-completed.spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-events_content")
  key                    = "tests/ha-gce.example.com/backups/etcd/events/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-main_content")
  key                    = "tests/ha-gce.example.com/backups/etcd/main/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "ha-gce-example-com-addons-bootstrap" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_ha-gce.example.com-addons-bootstrap_content")
  key                    = "tests/ha-gce.example.com/addons/bootstrap-channel.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "ha-gce-example-com-addons-core-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_ha-gce.example.com-addons-core.addons.k8s.io_content")
  key                    = "tests/ha-gce.example.com/addons/core.addons.k8s.io/v1.4.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "ha-gce-example-com-addons-coredns-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_ha-gce.example.com-addons-coredns.addons.k8s.io-k8s-1.12_content")
  key                    = "tests/ha-gce.example.com/addons/coredns.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "ha-gce-example-com-addons-dns-controller-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_ha-gce.example.com-addons-dns-controller.addons.k8s.io-k8s-1.12_content")
  key                    = "tests/ha-gce.example.com/addons/dns-controller.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "ha-gce-example-com-addons-kops-controller-addons-k8s-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_ha-gce.example.com-addons-kops-controller.addons.k8s.io-k8s-1.16_content")
  key                    = "tests/ha-gce.example.com/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "ha-gce-example-com-addons-kubelet-api-rbac-addons-k8s-io-k8s-1-9" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_ha-gce.example.com-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content")
  key                    = "tests/ha-gce.example.com/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "ha-gce-example-com-addons-limit-range-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_ha-gce.example.com-addons-limit-range.addons.k8s.io_content")
  key                    = "tests/ha-gce.example.com/addons/limit-range.addons.k8s.io/v1.5.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "ha-gce-example-com-addons-metadata-proxy-addons-k8s-io-v0-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_ha-gce.example.com-addons-metadata-proxy.addons.k8s.io-v0.1.12_content")
  key                    = "tests/ha-gce.example.com/addons/metadata-proxy.addons.k8s.io/v0.1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "ha-gce-example-com-addons-rbac-addons-k8s-io-k8s-1-8" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_ha-gce.example.com-addons-rbac.addons.k8s.io-k8s-1.8_content")
  key                    = "tests/ha-gce.example.com/addons/rbac.addons.k8s.io/k8s-1.8.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "ha-gce-example-com-addons-storage-gce-addons-k8s-io-v1-7-0" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_ha-gce.example.com-addons-storage-gce.addons.k8s.io-v1.7.0_content")
  key                    = "tests/ha-gce.example.com/addons/storage-gce.addons.k8s.io/v1.7.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "kops-version-txt" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_kops-version.txt_content")
  key                    = "tests/ha-gce.example.com/kops-version.txt"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-events_content")
  key                    = "tests/ha-gce.example.com/manifests/etcd/events.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-main_content")
  key                    = "tests/ha-gce.example.com/manifests/etcd/main.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-static-kube-apiserver-healthcheck" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-static-kube-apiserver-healthcheck_content")
  key                    = "tests/ha-gce.example.com/manifests/static/kube-apiserver-healthcheck.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-master-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-master-us-test1-a_content")
  key                    = "tests/ha-gce.example.com/igconfig/master/master-us-test1-a/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-master-us-test1-b" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-master-us-test1-b_content")
  key                    = "tests/ha-gce.example.com/igconfig/master/master-us-test1-b/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-master-us-test1-c" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-master-us-test1-c_content")
  key                    = "tests/ha-gce.example.com/igconfig/master/master-us-test1-c/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-nodes" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-nodes_content")
  key                    = "tests/ha-gce.example.com/igconfig/node/nodes/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "google_compute_disk" "d1-etcd-events-ha-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "ha-gce-example-com"
    "k8s-io-etcd-events"  = "1-2f1-2c2-2c3"
    "k8s-io-role-master"  = "master"
  }
  name = "d1-etcd-events-ha-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-a"
}

resource "google_compute_disk" "d1-etcd-main-ha-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "ha-gce-example-com"
    "k8s-io-etcd-main"    = "1-2f1-2c2-2c3"
    "k8s-io-role-master"  = "master"
  }
  name = "d1-etcd-main-ha-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-a"
}

resource "google_compute_disk" "d2-etcd-events-ha-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "ha-gce-example-com"
    "k8s-io-etcd-events"  = "2-2f1-2c2-2c3"
    "k8s-io-role-master"  = "master"
  }
  name = "d2-etcd-events-ha-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-b"
}

resource "google_compute_disk" "d2-etcd-main-ha-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "ha-gce-example-com"
    "k8s-io-etcd-main"    = "2-2f1-2c2-2c3"
    "k8s-io-role-master"  = "master"
  }
  name = "d2-etcd-main-ha-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-b"
}

resource "google_compute_disk" "d3-etcd-events-ha-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "ha-gce-example-com"
    "k8s-io-etcd-events"  = "3-2f1-2c2-2c3"
    "k8s-io-role-master"  = "master"
  }
  name = "d3-etcd-events-ha-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-c"
}

resource "google_compute_disk" "d3-etcd-main-ha-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "ha-gce-example-com"
    "k8s-io-etcd-main"    = "3-2f1-2c2-2c3"
    "k8s-io-role-master"  = "master"
  }
  name = "d3-etcd-main-ha-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-c"
}

resource "google_compute_firewall" "kubernetes-master-https-ha-gce-example-com" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "kubernetes-master-https-ha-gce-example-com"
  network       = google_compute_network.ha-gce-example-com.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "kubernetes-master-https-ipv6-ha-gce-example-com" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "kubernetes-master-https-ipv6-ha-gce-example-com"
  network       = google_compute_network.ha-gce-example-com.name
  source_ranges = ["::/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "master-to-master-ha-gce-example-com" {
  allow {
    protocol = "tcp"
  }
  allow {
    protocol = "udp"
  }
  allow {
    protocol = "icmp"
  }
  allow {
    protocol = "esp"
  }
  allow {
    protocol = "ah"
  }
  allow {
    protocol = "sctp"
  }
  disabled    = false
  name        = "master-to-master-ha-gce-example-com"
  network     = google_compute_network.ha-gce-example-com.name
  source_tags = ["ha-gce-example-com-k8s-io-role-master"]
  target_tags = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "master-to-node-ha-gce-example-com" {
  allow {
    protocol = "tcp"
  }
  allow {
    protocol = "udp"
  }
  allow {
    protocol = "icmp"
  }
  allow {
    protocol = "esp"
  }
  allow {
    protocol = "ah"
  }
  allow {
    protocol = "sctp"
  }
  disabled    = false
  name        = "master-to-node-ha-gce-example-com"
  network     = google_compute_network.ha-gce-example-com.name
  source_tags = ["ha-gce-example-com-k8s-io-role-master"]
  target_tags = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "node-to-master-ha-gce-example-com" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  allow {
    ports    = ["3988"]
    protocol = "tcp"
  }
  disabled    = false
  name        = "node-to-master-ha-gce-example-com"
  network     = google_compute_network.ha-gce-example-com.name
  source_tags = ["ha-gce-example-com-k8s-io-role-node"]
  target_tags = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "node-to-node-ha-gce-example-com" {
  allow {
    protocol = "tcp"
  }
  allow {
    protocol = "udp"
  }
  allow {
    protocol = "icmp"
  }
  allow {
    protocol = "esp"
  }
  allow {
    protocol = "ah"
  }
  allow {
    protocol = "sctp"
  }
  disabled    = false
  name        = "node-to-node-ha-gce-example-com"
  network     = google_compute_network.ha-gce-example-com.name
  source_tags = ["ha-gce-example-com-k8s-io-role-node"]
  target_tags = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "nodeport-external-to-node-ha-gce-example-com" {
  allow {
    ports    = ["30000-32767"]
    protocol = "tcp"
  }
  allow {
    ports    = ["30000-32767"]
    protocol = "udp"
  }
  disabled      = true
  name          = "nodeport-external-to-node-ha-gce-example-com"
  network       = google_compute_network.ha-gce-example-com.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "nodeport-external-to-node-ipv6-ha-gce-example-com" {
  allow {
    ports    = ["30000-32767"]
    protocol = "tcp"
  }
  allow {
    ports    = ["30000-32767"]
    protocol = "udp"
  }
  disabled      = true
  name          = "nodeport-external-to-node-ipv6-ha-gce-example-com"
  network       = google_compute_network.ha-gce-example-com.name
  source_ranges = ["::/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "ssh-external-to-master-ha-gce-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-master-ha-gce-example-com"
  network       = google_compute_network.ha-gce-example-com.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "ssh-external-to-master-ipv6-ha-gce-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-master-ipv6-ha-gce-example-com"
  network       = google_compute_network.ha-gce-example-com.name
  source_ranges = ["::/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "ssh-external-to-node-ha-gce-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-node-ha-gce-example-com"
  network       = google_compute_network.ha-gce-example-com.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "ssh-external-to-node-ipv6-ha-gce-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-node-ipv6-ha-gce-example-com"
  network       = google_compute_network.ha-gce-example-com.name
  source_ranges = ["::/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_instance_group_manager" "a-master-us-test1-a-ha-gce-example-com" {
  base_instance_name = "master-us-test1-a"
  name               = "a-master-us-test1-a-ha-gce-example-com"
  target_size        = 1
  version {
    instance_template = google_compute_instance_template.master-us-test1-a-ha-gce-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_group_manager" "a-nodes-ha-gce-example-com" {
  base_instance_name = "nodes"
  name               = "a-nodes-ha-gce-example-com"
  target_size        = 1
  version {
    instance_template = google_compute_instance_template.nodes-ha-gce-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_group_manager" "b-master-us-test1-b-ha-gce-example-com" {
  base_instance_name = "master-us-test1-b"
  name               = "b-master-us-test1-b-ha-gce-example-com"
  target_size        = 1
  version {
    instance_template = google_compute_instance_template.master-us-test1-b-ha-gce-example-com.self_link
  }
  zone = "us-test1-b"
}

resource "google_compute_instance_group_manager" "b-nodes-ha-gce-example-com" {
  base_instance_name = "nodes"
  name               = "b-nodes-ha-gce-example-com"
  target_size        = 1
  version {
    instance_template = google_compute_instance_template.nodes-ha-gce-example-com.self_link
  }
  zone = "us-test1-b"
}

resource "google_compute_instance_group_manager" "c-master-us-test1-c-ha-gce-example-com" {
  base_instance_name = "master-us-test1-c"
  name               = "c-master-us-test1-c-ha-gce-example-com"
  target_size        = 1
  version {
    instance_template = google_compute_instance_template.master-us-test1-c-ha-gce-example-com.self_link
  }
  zone = "us-test1-c"
}

resource "google_compute_instance_group_manager" "c-nodes-ha-gce-example-com" {
  base_instance_name = "nodes"
  name               = "c-nodes-ha-gce-example-com"
  target_size        = 0
  version {
    instance_template = google_compute_instance_template.nodes-ha-gce-example-com.self_link
  }
  zone = "us-test1-c"
}

resource "google_compute_instance_template" "master-us-test1-a-ha-gce-example-com" {
  can_ip_forward = true
  disk {
    auto_delete  = true
    boot         = true
    device_name  = "persistent-disks-0"
    disk_name    = ""
    disk_size_gb = 64
    disk_type    = "pd-standard"
    interface    = ""
    mode         = "READ_WRITE"
    source       = ""
    source_image = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-57-9202-64-0"
    type         = "PERSISTENT"
  }
  labels = {
    "k8s-io-cluster-name"   = "ha-gce-example-com"
    "k8s-io-instance-group" = "master-us-test1-a-ha-gce-example-com"
    "k8s-io-role-master"    = ""
  }
  machine_type = "n1-standard-1"
  metadata = {
    "cluster-name"                    = "ha-gce.example.com"
    "kops-k8s-io-instance-group-name" = "master-us-test1-a"
    "ssh-keys"                        = file("${path.module}/data/google_compute_instance_template_master-us-test1-a-ha-gce-example-com_metadata_ssh-keys")
    "startup-script"                  = file("${path.module}/data/google_compute_instance_template_master-us-test1-a-ha-gce-example-com_metadata_startup-script")
  }
  name_prefix = "master-us-test1-a-ha-gce--ke5ah6-"
  network_interface {
    access_config {
    }
    network    = google_compute_network.ha-gce-example-com.name
    subnetwork = google_compute_subnetwork.us-test1-ha-gce-example-com.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
  }
  service_account {
    email  = google_service_account.control-plane.email
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/ndev.clouddns.readwrite"]
  }
  tags = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_instance_template" "master-us-test1-b-ha-gce-example-com" {
  can_ip_forward = true
  disk {
    auto_delete  = true
    boot         = true
    device_name  = "persistent-disks-0"
    disk_name    = ""
    disk_size_gb = 64
    disk_type    = "pd-standard"
    interface    = ""
    mode         = "READ_WRITE"
    source       = ""
    source_image = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-57-9202-64-0"
    type         = "PERSISTENT"
  }
  labels = {
    "k8s-io-cluster-name"   = "ha-gce-example-com"
    "k8s-io-instance-group" = "master-us-test1-b-ha-gce-example-com"
    "k8s-io-role-master"    = ""
  }
  machine_type = "n1-standard-1"
  metadata = {
    "cluster-name"                    = "ha-gce.example.com"
    "kops-k8s-io-instance-group-name" = "master-us-test1-b"
    "ssh-keys"                        = file("${path.module}/data/google_compute_instance_template_master-us-test1-b-ha-gce-example-com_metadata_ssh-keys")
    "startup-script"                  = file("${path.module}/data/google_compute_instance_template_master-us-test1-b-ha-gce-example-com_metadata_startup-script")
  }
  name_prefix = "master-us-test1-b-ha-gce--c8u7qq-"
  network_interface {
    access_config {
    }
    network    = google_compute_network.ha-gce-example-com.name
    subnetwork = google_compute_subnetwork.us-test1-ha-gce-example-com.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
  }
  service_account {
    email  = google_service_account.control-plane.email
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/ndev.clouddns.readwrite"]
  }
  tags = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_instance_template" "master-us-test1-c-ha-gce-example-com" {
  can_ip_forward = true
  disk {
    auto_delete  = true
    boot         = true
    device_name  = "persistent-disks-0"
    disk_name    = ""
    disk_size_gb = 64
    disk_type    = "pd-standard"
    interface    = ""
    mode         = "READ_WRITE"
    source       = ""
    source_image = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-57-9202-64-0"
    type         = "PERSISTENT"
  }
  labels = {
    "k8s-io-cluster-name"   = "ha-gce-example-com"
    "k8s-io-instance-group" = "master-us-test1-c-ha-gce-example-com"
    "k8s-io-role-master"    = ""
  }
  machine_type = "n1-standard-1"
  metadata = {
    "cluster-name"                    = "ha-gce.example.com"
    "kops-k8s-io-instance-group-name" = "master-us-test1-c"
    "ssh-keys"                        = file("${path.module}/data/google_compute_instance_template_master-us-test1-c-ha-gce-example-com_metadata_ssh-keys")
    "startup-script"                  = file("${path.module}/data/google_compute_instance_template_master-us-test1-c-ha-gce-example-com_metadata_startup-script")
  }
  name_prefix = "master-us-test1-c-ha-gce--3unp7l-"
  network_interface {
    access_config {
    }
    network    = google_compute_network.ha-gce-example-com.name
    subnetwork = google_compute_subnetwork.us-test1-ha-gce-example-com.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
  }
  service_account {
    email  = google_service_account.control-plane.email
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/ndev.clouddns.readwrite"]
  }
  tags = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_instance_template" "nodes-ha-gce-example-com" {
  can_ip_forward = true
  disk {
    auto_delete  = true
    boot         = true
    device_name  = "persistent-disks-0"
    disk_name    = ""
    disk_size_gb = 128
    disk_type    = "pd-standard"
    interface    = ""
    mode         = "READ_WRITE"
    source       = ""
    source_image = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-57-9202-64-0"
    type         = "PERSISTENT"
  }
  labels = {
    "k8s-io-cluster-name"   = "ha-gce-example-com"
    "k8s-io-instance-group" = "nodes-ha-gce-example-com"
    "k8s-io-role-node"      = ""
  }
  machine_type = "n1-standard-2"
  metadata = {
    "cluster-name"                    = "ha-gce.example.com"
    "kops-k8s-io-instance-group-name" = "nodes"
    "ssh-keys"                        = file("${path.module}/data/google_compute_instance_template_nodes-ha-gce-example-com_metadata_ssh-keys")
    "startup-script"                  = file("${path.module}/data/google_compute_instance_template_nodes-ha-gce-example-com_metadata_startup-script")
  }
  name_prefix = "nodes-ha-gce-example-com-"
  network_interface {
    access_config {
    }
    network    = google_compute_network.ha-gce-example-com.name
    subnetwork = google_compute_subnetwork.us-test1-ha-gce-example-com.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
  }
  service_account {
    email  = google_service_account.node.email
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/devstorage.read_only"]
  }
  tags = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_network" "ha-gce-example-com" {
  auto_create_subnetworks = false
  name                    = "ha-gce-example-com"
}

resource "google_compute_subnetwork" "us-test1-ha-gce-example-com" {
  ip_cidr_range = "10.0.16.0/20"
  name          = "us-test1-ha-gce-example-com"
  network       = google_compute_network.ha-gce-example-com.name
  region        = "us-test1"
}

resource "google_project_iam_binding" "serviceaccount-control-plane" {
  members = ["serviceAccount:control-plane-ha-gce-ex-mr702t@testproject.iam.gserviceaccount.com"]
  project = "testproject"
  role    = "roles/container.serviceAgent"
}

resource "google_project_iam_binding" "serviceaccount-nodes" {
  members = ["serviceAccount:node-ha-gce-example-com@testproject.iam.gserviceaccount.com"]
  project = "testproject"
  role    = "roles/compute.viewer"
}

resource "google_service_account" "control-plane" {
  account_id   = "control-plane-ha-gce-ex-mr702t"
  description  = "kubernetes control-plane instances"
  display_name = "control-plane"
  project      = "testproject"
}

resource "google_service_account" "node" {
  account_id   = "node-ha-gce-example-com"
  description  = "kubernetes worker nodes"
  display_name = "node"
  project      = "testproject"
}

terraform {
  required_version = ">= 0.15.0"
  required_providers {
    google = {
      "source"  = "hashicorp/google"
      "version" = ">= 2.19.0"
    }
  }
}
