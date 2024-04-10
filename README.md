# IBM Backup and Recovery Service SDK GO

**Note:**
IBM Cloud Backup and Recovery is the Limited Availability (LA) offering in the present release and currently not available under the "General Availability (GA)".  Only after the GA release, it would be available through the "IBM Global Catalog" for delivery and consumption from all available Data center Region/Zones.  For more details/Questions about products, sales, support etc at [IBM Help](https://www.ibm.com/contact/global).

Go client library to interact with the various [IBM Cloud Backup and Recovery service APIs](https://cloud.ibm.com/docs/allowlist/backup-recovery?topic=backup-recovery-compatibility-api).

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [Go modules](#go-modules)
  - [`go get` command](#go-get-command)
- [Questions](#questions)
- [Issues](#issues)
- [License](#license)

## Overview

The IBM Cloud Backup and Recovery service Go SDK allows developers to programmatically interact with the following IBM Cloud services:

Service Name | Package name
--- | ---
[Backup Recovery](https://cloud.ibm.com/docs/allowlist/backup-recovery) | backuprecoveryv1

## Prerequisites

[ibm-cloud-onboarding]: https://cloud.ibm.com/registration

- An [IBM Cloud][ibm-cloud-onboarding] account.
- An IAM API key to allow the SDK to access your account. Create one [here](https://cloud.ibm.com/iam/apikeys).
- Go version 1.21.0 or above.

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

### `go get` command  

Alternatively, you can use the `go get` command to download and install the appropriate packages needed by your application:

``` go
go get -u github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1
```

Be sure to use the appropriate package name from the service table above for the services used by your application.

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
