# terraform-provider-adcm

```terraform
terraform {
  required_providers {
    adcm = {
      source = "github.com/giggsoff/adcm"
    }
  }
}

resource "tls_private_key" "example" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

provider "adcm" {
  host     = "http://127.0.0.1:8000"
  username = "admin"
  password = "admin"
}

data "adcm_bundle" "ssh" {
  name = "SSH Common"
}
data "adcm_provider" "ssh" {
  bundle_id = data.adcm_bundle.ssh.id
}
resource "adcm_host" "h1" {
  provider_id = data.adcm_provider.ssh.id
  fqdn        = "h1"
  config      = jsonencode({
    "ansible_user" : "adcm", "ansible_host" : "127.0.0.1",
    "ansible_ssh_private_key_file" : "${tls_private_key.example.private_key_pem}"
  })
}
data "adcm_bundle" "adpg" {
  name = "ADPG"
}
resource "adcm_cluster" "c1" {
  bundle_id   = data.adcm_bundle.adpg.id
  name        = "c1"
  description = "c1"
  hc_map      = jsonencode({
    "${resource.adcm_host.h1.fqdn}" : [{ "adpg" : ["adpg"] }]
  })
  services_config = jsonencode({
    "adpg" : {
      "datadir" : "/pg_data1"
    }
  })
  cluster_config = jsonencode({
    "disable_firewall" : true,
    "repos" : {
      "use_adpg_repo" : true
    }
  })
}
```