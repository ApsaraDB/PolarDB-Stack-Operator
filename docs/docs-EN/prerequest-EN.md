# Deployment Prerequisites
## Hardware Requirements
| Item | Minimum Requirements |
| --- | --- |
| Number of Machines | 3 |
| Server Architecture | Supports the x86 and ARM architecture |
| Hard Disk | SSD is used as the system disk and its capacity should be 300 GB and above. Two separate NVMe disks (at least 200 GB capacity for each one for test only and at least 1 TB for each one in the production environment) are used as the storage disk of the control data and the redline data disk and are mounted under the directory /disk1 and /disk2, respectively. |
| Storage | Shared storage, distributed storage, or SAN. It is required to support the SCSI or NVME protocol and PR locking and be mounted on all machines. You can use the command `multipath -ll ` to view all shared disks. |

## Software Requirements
| Item | Requirements |
| --- | --- |
| Docker | / |
| Kubernetes | 1.14.8+ |
| Root Password | All servers should use the same root password and set up password-less access to each other. |
| Clock Synchronization | The operating system has been configured with NTP clock synchronization. |