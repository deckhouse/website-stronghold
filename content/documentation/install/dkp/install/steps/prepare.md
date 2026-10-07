---
title: "Prepare the environment"
description: "Steps to prepare cluster nodes before installing the platform: operating system, registry access, SSH access, and a technical user."
weight: 10
---

Before installing the platform, complete the following steps:

1. **Install the operating system:**
   - Install one of the [supported operating systems](../../../../about/requirements/#supported-os) on each cluster node. Pay attention to the system version and architecture.

1. **Ensure access to the container registry:**
   - Make sure that the container image registry is reachable from each cluster node. By default, `registry.deckhouse.io` is used. Configure the network connection and the security policies required to access this registry.

1. **Configure SSH access to the nodes:**
   - Disable password login for users to protect the system from unauthorized access.
   - Generate an SSH key that will be used when installing the cluster nodes. Example of generating a key:

     ```bash
     ssh-keygen -t rsa -b 4096 -f cluster-node -N "" -C "cluster-node" -v
     ```

1. **Create a technical user:**
   - On each node, create a technical user with administrator privileges, for example, `install-user`. This user is required for automatic installation and configuration of cluster components.
   - Add the previously generated public SSH key for this user to enable passwordless authentication.

Once all the steps are completed, the cluster nodes are ready for further installation and configuration of the platform.
