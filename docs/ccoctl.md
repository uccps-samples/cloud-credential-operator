# Cloud Credential Operator utility

To assist with the creating and maintenance of cloud credentials from outside the cluster (necessary when CCO is put in "Manual" mode), the `ccoctl` tool provides various commands to help with the creation and management of cloud credentials.

## AWS

### Global flags

By default, the tool will output to the directory the command(s) were run in. To specify a directory, use the `--output-dir` flag.

Commands which would otherwise make AWS API calls can be passed the `--dry-run` flag to have `ccoctl` place JSON files on the local filesystem instead of creating/modifying any AWS resources. These JSON files can be reviewed/modified and then applied with the aws CLI tool (using the `--cli-input-json` parameters).

### Creating RSA keys

To generate keys for use when setting up the cluster's OpenID Connect provider, run

```bash
$ ccoctl aws create-key-pair
```

This will write out public/private key files named `serviceaccount-signer.private` and `serviceaccount-signer.public`.

### Creating OpenID Connect Provider

To set up an OpenID Connect provider in the cloud, run

```bash
$ ccoctl aws create-identity-provider --name=<name> --region=<aws-region> --public-key-file=/path/to/public/key/file

```

where `name` is the name used to tag and account any cloud resources that are created. `region` is the aws region in which cloud resources will be created and `public-key-file` is the path to a public key file generated using `ccoctl aws create-key-pair` command.

The above command will write out discovery document file named `02-openid-configuration` and JSON web key set file named `03-keys.json` when `--dry-run` flag is set.

### Creating IAM Roles

To create IAM Roles for each in-cluster component, you need to first extract the list of CredentialsRequest objects from the OpenShift release image

```bash
$ oc adm release extract --credentials-requests --cloud=aws --to=./credrequests quay.io/path/to/uccp-release:version
```

Then you can use `ccoctl` to process each CredentialsRequest object in the `./credrequests` directory (from the example above)

```bash
$ ccoctl aws create-iam-roles --name=<name> --region=<aws-region> --credentials-requests-dir=<path-to-directory-with-list-of-credentials-requests> --identity-provider-arn=<arn-of-identity-provider-created-in-previous-step>
```

This will create one IAM Role for each CredentialsRequest with a trust policy tied to the provided Identity Provider, and a permissions policy as defined in each CredentialsRequest object from the OpenShift release image.

It will also populate the `<output-dir>/manifests` directory with Secret files for each CredentialsRequest that was processed. These can be provided to the installer so that the appropriate Secrets are available for each in-cluster component needing to make cloud API calls.

### Creating all the required resources together

To create all the above mentioned resources in one go, run

```bash
$ oc adm release extract --credentials-requests --cloud=aws --to=./credrequests quay.io/path/to/uccp-release:version
```

Then you can use `ccoctl` to process all CredentialsRequest objects in the `./credrequests` directory (from the example above)

```bash
$ ccoctl aws create-all --name=<name> --region=<aws-region> --credentials-requests-dir=<path-to-directory-with-list-of-credentials-requests>
```

### Deleting resources

To delete resources created by ccoctl, run

```bash
$ ccoctl aws delete --name=<name> --region=<aws-region>

```

where `name` is the name used to tag and account any cloud resources that were created. `region` is the aws region in which cloud resources were created.

## GCP

### Global flags

By default, the tool will output to the directory the command(s) were run in. To specify a directory, use the `--output-dir` flag.

Commands which would otherwise make GCP API calls can be passed the `--dry-run` flag to have `ccoctl` place bash scripts on the local filesystem instead of creating/modifying any GCP resources. These scripts can be reviewed/modified and then run to create cloud resources.

### Creating RSA keys

To generate keys for use when setting up the cluster's OpenID Connect provider, run

```bash
$ ccoctl gcp create-key-pair
```

This will write out public/private key files named `serviceaccount-signer.private` and `serviceaccount-signer.public`.

### Creating Workload Identity Pool

To set up a workload identity pool in the cloud, run 

```bash
$ ccoctl gcp create-workload-identity-pool --name=<name> --project=<gcp-project-id>
```

where `name` is the name prefix for any cloud resources that are created. `project` is the ID of the gcp project.

### Creating Workload Identity Provider

To set up a Workload Identity Provider in the cloud, run

```bash
$ ccoctl gcp create-workload-identity-provider --name=<name> --region=<gcp-region> --project=<gcp-project-id> --public-key-file=/path/to/public/key/file --workload-identity-pool=<pool-id>
```

where `name` is the name prefix for any cloud resources that are created. `region` is the gcp region in which the Google Cloud Storage will be created. `project` is the ID of the gcp project. `workload-identity-pool` is the ID of the pool created using `ccoctl gcp create-workload-identity-pool`. The new provider will be created in this pool.

The above command will write out discovery document file named `02-openid-configuration` and JSON web key set file named `03-keys.json` when `--dry-run` flag is set.


### Creating IAM Service Accounts

To create IAM Service Account for each in-cluster component, you need to first extract the list of CredentialsRequest objects from the OpenShift release image

```bash
$ oc adm release extract --credentials-requests --cloud=gcp --to=./credrequests quay.io/path/to/uccp-release:version
```

Then you can use `ccoctl` to process each CredentialsRequest object in the `./credrequests` directory (from the example above)

```bash
$ ccoctl gcp create-service-accounts --name=<name> --project=<gcp-project-id> --credentials-requests-dir=<path-to-directory-with-list-of-credentials-requests> --workload-identity-pool=<pool-id> --workload-identity-provider=<provider-id>
```

where `name` is the name prefix for any cloud resources that are created. `project` is the ID of the gcp project. `public-key-file` is the path to a public key file generated using `ccoctl gcp create-key-pair` command. `workload-identity-pool` is the ID of the pool created using `ccoctl gcp create-workload-identity-pool` command. `workload-identity-provider` is the ID of the provider created using `ccoctl gcp create-workload-identity-provider` command.

This will create one IAM Service Account for each CredentialsRequest along with appropriate project policy bindings as defined in each CredentialsRequest object from the OpenShift release image.

It will also populate the `<output-dir>/manifests` directory with Secret files for each CredentialsRequest that was processed. These can be provided to the installer so that the appropriate Secrets are available for each in-cluster component needing to make cloud API calls.

### Creating all the required resources together

To create all the above mentioned resources in one go, run

```bash
$ oc adm release extract --credentials-requests --cloud=gcp --to=./credrequests quay.io/path/to/uccp-release:version
```

Then you can use `ccoctl` to process all CredentialsRequest objects in the `./credrequests` directory (from the example above)

```bash
$ ccoctl gcp create-all --name=<name> --region=<gcp-region> --project=<gcp-project-id> --credentials-requests-dir=<path-to-directory-with-list-of-credentials-requests>
```

### Deleting resources

To delete resources created by ccoctl, run

```bash
$ ccoctl gcp delete --name=<name> --project=<gcp-project-id>

```

where `name` is the name prefix used to create cloud resources. `project` is the ID of the gcp project.

## IBMCloud

### Global flags

By default, the tool will output to the directory the command(s) were run in. To specify a directory, use the `--output-dir` flag.

### Extract the Credentials Request objects from the above release image

`ccoctl ibmcloud` can process two kind of credentials requests - `IBMCloudProviderSpec`, `IBMCloudPowerVSProviderSpec` and here are the steps to extract them from the release image

#### IBM Cloud

This extracts the credentials of kind `IBMCloudProviderSpec`

```bash
mkdir credreqs ; oc adm release extract --cloud=ibmcloud --credentials-requests $RELEASE_IMAGE --to=./credreqs
```

#### IBM Cloud Power VS

This extracts the credentials of kind `IBMCloudPowerVSProviderSpec`

```bash
mkdir credreqs ; oc adm release extract --cloud=powervs --credentials-requests $RELEASE_IMAGE --to=./credreqs
```

### Creating Service IDs

This command will create the service ID for each credential request, assign the policies defined, creates an API key in the IBM Cloud and generates the secret.

```bash
ccoctl ibmcloud create-service-id --credentials-requests-dir <path-to-directory-with-list-of-credentials-requests> --name <name> --resource-group-name <resource-group-name>
```

> Note: --resource-group-name option is optional, but it is recommended to use to have finer grained access to the resources. 

### Refresh the API keys for Service ID

```bash
ccoctl ibmcloud refresh-keys --kubeconfig <uccp-kubeconfig-file> --credentials-requests-dir <path-to-directory-with-list-of-credentials-requests> --name <name> 
```

> Note: Any new credential request in the credentials request directory will require the --create parameter.

> **WARNING**: The above command will replace the old API key with newly created api key, hence all the effecting pods need to be recreated after successful of the command. 

### Deleting the Service IDs

This command will delete the service id from the IBM Cloud

```bash
ccoctl ibmcloud delete-service-id --credentials-requests-dir <path-to-directory-with-list-of-credentials-requests> --name <name> 
```

## Alibaba Cloud

This is a guide for using manual mode on alibaba cloud, for more info about manual mode, please refer to [cco-mode-manual](https://github.com/uccps-samples/cloud-credential-operator/blob/master/docs/mode-manual-creds.md).

For alibaba cloud,  the CCO utility (`ccoctl`) binary will create credentials Secret manifests for the OpenShift installer. It will also create a user along with a long-lived RAM AccessKey (AK) for each OpenShift in-cluster component. Every RAM user who owns the AK would be attached RAM policies with the permission defined in each component's CredentialsRequest. To do all this, ccoctl consumes AK of a root RAM user with permissions required for creating user/policy and also attaching the policy to the user.

### Prerequisite

1. Extract and prepare the ccoctl binary from the release image.

2. Choose an existing RAM user who has the below permissions, and get this user's accesskey id/secret for creating the RAM users and policies for each in-cluster component.

   ```bash
   ram:CreatePolicy
   ram:GetPolicy
   ram:CreatePolicyVersion
   ram:DeletePolicy
   ram:DetachPolicyFromUser
   ram:ListPoliciesForUser
   ram:AttachPolicyToUser
   ram:CreateUser
   ram:GetUser
   ram:DeleteUser
   ram:CreateAccessKey
   ram:ListAccessKeys
   ram:DeleteAccessKey
   ```

3. Use the selected RAM user’s accesskey id/secret to configure the Alibaba Cloud SDK client's credential provider chain with [Envionment Creadentials](https://github.com/aliyun/alibaba-cloud-sdk-go/blob/master/docs/2-Client-EN.md#1-environment-credentials) mode or through [Credentials File](https://github.com/aliyun/alibaba-cloud-sdk-go/blob/master/docs/2-Client-EN.md#2-credentials-file) mode

### Procedure

1. Extract the list of CredentialsRequest custom resources (CRs) from the OpenShift Container Platform release image:

   ```bash
   $ oc adm release extract --credentials-requests --cloud=alibabacloud --to=<path_to_directory_with_list_of_credentials_requests>/credrequests quay.io/<path_to>/ocp-release:<version>
   
   ```

   >  step 2&3 are only needed when preparing for upgrading clusters with manually maintained credentials. When doing a fresh installation please skip step 2&3**

2. For each CredentialsRequest CR in the release image, ensure that a namespace that matches the text in the spec.secretRef.namespace field exists in the cluster. This field is where the generated secrets that hold the credentials configuration are stored.

   Sample Alibaba Cloud CredentialsRequest object

   ```yaml
   apiVersion: cloudcredential.uccp.io/v1
   kind: CredentialsRequest
   metadata:
     name: cloud-credential-operator-ram-ro
     namespace: uccp-cloud-credential-operator
   spec:
     providerSpec:
       apiVersion: cloudcredential.uccp.io/v1
       kind: AlibabaCloudProviderSpec
       statementEntries:
       - action:
         - ecs:CopySnapshot
         - ecs:DeleteDisk
         - ecs:DescribeInstanceAttribute
         - ecs:DescribeInstances
         effect: Allow
         resource: '*'
     secretRef:
       namespace: cloud-credential-operator-ram-ro-creds
       name: uccp-cloud-credential-operator
   ```

3. For any `CredentialsRequest` CR for which the cluster does not already have a namespace with the name specified in `spec.secretRef.namespace`, create the namespace:

   ```bash
   $ oc create namespace <component_namespace>
   ```

4. Use the `ccoctl` tool to process all `CredentialsRequest` objects in the `credrequests` directory:

   ```bash
   $ ccoctl alibabacloud create-ram-users --name <name> --region=<region> --credentials-requests-dir=<path_to_directory_with_list_of_credentials_requests>/credrequests --output-dir=xxxxxx
   ```

    where:

   - `name` is the name used to tag any cloud resources that are created for tracking. 
   - `region` is the Alibaba Cloud region in which cloud resources will be created.
   - `credentials-requests-dir` is the directory containing files of component CredentialsRequests.
   - `output-dir`/manifests is the directory containing files of component credentials secret.
    
   > Note:  A ram user can have up to two accesskeys at the same time, so when the `ccoctl alibabacloud create-ram-users` command is run more than twice,  the previous generated manifests secret will become stale and you should apply the new generated secrets again.**

5. Prepare to run the OpenShift Container Platform installer:

   a. Create the install-config.yaml file:
   ```bash
   $ uccp-install create install-config --dir ./path/to/installation/dir
   ```
   b. Configure the cluster to install with the CCO in manual mode:

   ```bash
   $ echo "credentialsMode: Manual" >> ./path/to/installation/dir/install-config.yaml
   ```
   
   c. Create install manifests:

   ```bash
   $ uccp-install create manifests --dir ./path/to/installation/dir
   ```

   d. Copy the generated credential files to the target manifests directory:

   ```bash
   $ cp <output_dir>/manifests/*credentials.yaml ./path/to/installation/dir/manifests/
   ```
6. To delete resources created by ccoctl, run

   ```bash
   $ ccoctl alibabacloud delete-ram-users --name <name> --region=<region>
   ```
   where:
   - `name` is the name used to tag any cloud resources that are created for tracking. 
   - `region` is the Alibaba Cloud region in which cloud resources will be created.   