# AWS Multi-Account Credentials Rotator

Welcome to the **AWS Multi-Account Credentials Rotator**! This tool helps you seamlessly rotate your AWS credentials across multiple AWS accounts, ensuring enhanced security and compliance. Each profile configured using the AWS CLI is expected to relate to a separate AWS account, but uses the same username.

## Features

- **Profile Selection**: Easily select which AWS profiles to update, with options to update all profiles or all except the default.
- **Automated Backup**: Automatically create a backup of your AWS credentials file before making any changes.
- **Session Validation**: Ensure new credentials are valid before removing old ones.
- **User-Friendly Interface**: Interactive command-line interface for seamless user experience.
- **Detailed Status Messages**: Get clear feedback on the status of each profile update.

## Usage

### Prerequisites

Ensure you have the following installed:

- [Go](https://golang.org/dl/)
- [AWS CLI](https://aws.amazon.com/cli/)

### Building the Application

Clone the repository and build the application:

```bash
git clone https://bitbucket.org/neilcroll/aws-rotate.git
cd aws-rotate
go mod tidy
go build
```

### Running the Application

To run the application, use the following command:
```powershell
aws-rotate.exe 
```

or
```bash
./aws-rotate
```

There are command-line flags available to customize the application:
1. ``-c`` or ``--concurrent`` (default): Update the chosen profiles concurrently.
2. ``-s`` or ``--sequential``: Update the chosen profiles sequentially. This approach will be slower than updating concurrently, but may be required if you encounter issues with concurrent updates. 

### Example Workflow

1. **Profile Selection**: Choose the profiles you want to update. Options include updating specific profiles, all profiles or all except the default.
2. **Username Input**: Enter the username for the AWS accounts.
3. **Password Input**: Enter the new password for the AWS accounts.
4. **Backup Creation**: A backup of your existing credentials file is created.
5. **Concurrent Profile Updates**: The application updates the credentials for each selected profile concurrently.
6. **Completion**: Upon successful updates, a message is displayed confirming the update of each profile.

## Contributing

I welcome contributions! If you have any ideas, suggestions, or bug reports, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License.
