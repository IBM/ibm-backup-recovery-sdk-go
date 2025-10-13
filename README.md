# IBM Cloud Backup and Recovery Service SDK GO

This GO SDK for IBM Cloud Backup and Recovery enables customers to leverage a managed service that provides backup solutions for various customer workloads running on IBM Cloud.

**Note**:
IBM Cloud Backup and recovery supporting IKS/ROKS workload is only available in select regions. For more details/Questions about products, sales, support etc at [IBM Help](https://www.ibm.com/contact/global)

## Table of Contents

- [Documentation](#documentation)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [Go modules](#go-modules)
  - [Get SDK package](#get-sdk-package)
- [Questions](#questions)
- [Issues](#issues)
- [License](#license)

## Documentation

- [Documentation for IBM Cloud Backup and Recovery](https://cloud.ibm.com/docs/backup-recovery?topic=backup-recovery-getting-started-backup-recovery)
- [REST API Reference and Code Examples](https://cloud.ibm.com/docs/backup-recovery?topic=backup-recovery-using-go)
- [Go Backup and Recovery API reference documentation](https://pkg.go.dev/github.com/IBM/ibm-backup-recovery-sdk-go)

## Prerequisites

- An [IBM Cloud](https://cloud.ibm.com/registration) account.
- An IAM API key to allow the SDK to access your account. Create one [here](https://cloud.ibm.com/iam/apikeys).
- Go version 1.23 or newer.

## Installation

### Go modules

If your application uses Go modules for dependency management (recommended), just add an import for each service
that you will use in your application.  
Here is an example:

```go
import (
 "github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
)
```

Next, run `go build` or `go mod tidy` to download and install the new dependencies and update your application's
`go.mod` file.  

In the example above, the `backuprecoveryv1` part of the import path is the package name
associated with the Example Service.
See the service table above to find the appropriate package name for the services used by your application.

### Get SDK package

Alternatively, you can use the `go get` command to download and install the appropriate packages needed by your application:

``` go
go get -u github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1
```

## Questions

If you are having difficulties using this SDK or have a question about the IBM Cloud services,
please ask a question at
[Stack Overflow](http://stackoverflow.com/questions/ask?tags=ibm-cloud).

## Issues

If you encounter an issue with the project, you are welcome to submit a
[bug report](https://github.com/IBM/ibm-backup-recovery-sdk-go/issues).
Before that, please search for similar issues. It's possible that someone has already reported the problem.

## Language Support Policy

IBM supports [current public releases](https://golang.org/doc/devel/release.html). IBM will deprecate language versions 90 days after a version reaches end-of-life. All clients will need to upgrade to a supported version before the end of the grace period.

## License

This SDK project is released under the Apache 2.0 license.
The license's full text can be found in [LICENSE](LICENSE).
