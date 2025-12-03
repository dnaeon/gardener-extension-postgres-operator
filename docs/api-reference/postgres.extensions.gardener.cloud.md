# API Reference

## Packages
- [postgres.extensions.gardener.cloud/v1alpha1](#postgresextensionsgardenercloudv1alpha1)


## postgres.extensions.gardener.cloud/v1alpha1

Package v1alpha1 provides the v1alpha1 version of the external API types.





#### PostgresConfigSpec



PostgresConfigSpec defines the desired state of [PostgresConfig]



_Appears in:_
- [PostgresConfig](#postgresconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `volumeSize` _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#quantity-resource-api)_ | VolumeSize specifies the size of the persistent volume for the<br />cluster nodes. |  |  |
| `replicas` _integer_ | Replicas specifies the number of cluster instances. |  |  |
| `users` _object (keys:string, values:string array)_ | Users specifies the database users and their roles. |  |  |
| `databases` _object (keys:string, values:string)_ | Databases specifies the database names and their owners. |  |  |
| `postgresVersion` _string_ | PostgresVersion specifies the PostgreSQL version of the cluster. |  |  |


