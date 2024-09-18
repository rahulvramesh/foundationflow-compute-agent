# Server Monitoring Agent

## Table of Contents
1. [Overview](#overview)
2. [Features](#features)
3. [Requirements](#requirements)
4. [Installation](#installation)
5. [Usage](#usage)
6. [Build Instructions](#build-instructions)
7. [Configuration](#configuration)
8. [Data Collection](#data-collection)
9. [Security Considerations](#security-considerations)
10. [Troubleshooting](#troubleshooting)
11. [Contributing](#contributing)
12. [License](#license)

## Overview

The Server Monitoring Agent is a robust, lightweight tool written in Go that collects and reports various system metrics from Linux servers. It's designed to gather detailed information about system resources and send this data to a specified endpoint, making it ideal for centralized monitoring solutions.

## Features

- Collects detailed system information including:
  - Memory usage
  - Swap usage
  - Storage details for all attached drives
  - CPU usage (overall and per-core)
  - GPU usage (for NVIDIA GPUs)
  - Detailed CPU information via `lscpu`
- Stores collected data locally in an SQLite database
- Sends collected data to a specified remote endpoint
- Configurable reporting frequency
- Cross-platform compatibility (builds available for multiple Linux architectures)
- Secure communication with customizable authentication token

## Requirements

- Go 1.15 or higher
- SQLite
- NVIDIA drivers (for GPU monitoring, optional)
- Linux operating system

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/server-monitoring-agent.git
   cd server-monitoring-agent
   ```

2. Install dependencies:
   ```
   make deps
   ```

3. Build the agent:
   ```
   make build-local
   ```

## Usage

Run the agent with the following command:

```
./server-monitor-agent -url https://your-reporting-endpoint.com/api -token your-auth-token -freq 5
```

Command-line flags:
- `-url`: The URL to send reports to (required)
- `-token`: Authentication token for the reporting endpoint (required)
- `-freq`: Reporting frequency in minutes (default: 5)

## Build Instructions

To build for all supported Linux architectures:

```
make build-linux
```

This will create binaries for various architectures in the `build` directory and a tarball `server-monitor-agent-linux-binaries.tar.gz`.

Other useful make commands:
- `make test`: Run tests
- `make clean`: Clean up built binaries and directories
- `make run`: Build and run the agent locally

## Configuration

The agent is configured via command-line flags. Ensure you set the correct reporting URL and authentication token when running the agent.

## Data Collection

The agent collects the following data:

1. Memory Information:
   - Total memory
   - Available memory
   - Used memory

2. Swap Information:
   - Total swap
   - Used swap
   - Free swap

3. Storage Information (for each attached drive):
   - Device name
   - Total space
   - Used space
   - Free space

4. CPU Information:
   - Number of cores
   - Usage percentage (overall and per-core)

5. GPU Information (if NVIDIA GPU is present):
   - Usage percentage

6. Detailed CPU information from `lscpu` command

## Security Considerations

- Always use HTTPS for the reporting endpoint in production environments.
- The agent currently disables SSL certificate verification. This should be enabled in production environments.
- Keep the authentication token secure and rotate it regularly.
- Ensure the SQLite database file has appropriate file permissions.

## Troubleshooting

- If you encounter "command not found" errors, ensure Go is correctly installed and your PATH is set correctly.
- For GPU-related errors, check that NVIDIA drivers are installed and configured correctly.
- If the agent fails to send reports, check your network configuration and firewall settings.

## Contributing

Contributions to the Server Monitoring Agent are welcome! Please feel free to submit pull requests, create issues or spread the word.

## License

[Include your chosen license here, e.g., MIT, GPL, etc.]