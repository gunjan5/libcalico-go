crds.yaml is applied before running the tests to initialize CRDs (CustomResourceDefinitions)
for the Kubernetes backend.
This manifest is applied in the Makefile once kubernetes API server is running.
crds.yaml creates the following CRDs:
  - GlobalConfig
  - GlobalBGPPeer
  - GlobalBgpConfig
  - IPPool
  - SystemNetworkPolicy

These CRDs must be created in advance for any Calico deployment with Kubernetes backend,
typically as part of the same manifest used to setup Calico.