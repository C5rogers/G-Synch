# G-Synch

A powerful database synchronization **CLI** tool built with **Go**. G-Synch enables seamless synchronization between different database systems, ensuring data consistency and integrity across your applications.

## Why G-Synch?

- **Easy to Use**: Simple command-line interface for quick setup and execution for synchronization.

## Features

### Core Features

- **Bidirectional Sync**: Synchronize data in both directions between source and target databases.
- **Conflict Resolution**: Built-in mechanisms to handle data conflicts during synchronization.
- **Logging and Monitoring**: Comprehensive logging and monitoring capabilities to track synchronization progress and troubleshoot issues.

### Core Commands

- `check`: Displays the current difference between the given database with the target database.
- `sync`: Initiates the synchronization process between the specified source and target databases.
- `reverse-check`: Checks for differences in the reverse direction, from target to source database.

### Technical Features

- **Multi-Database Instance Support**: Synchronize different database instances such as `PostgreSQL`.
- **Schema Mapping**: Automatically maps schemas between different database systems.
- **Incremental Sync**: Supports incremental synchronization to minimize data transfer and improve performance.
- **Logging**: Detailed logging of all synchronization activities for audit and troubleshooting purposes.

## 🚀 Quick Start

```bash
git clone https://github.com/c5rogers/G-Synch.git
```

### Basic Usage

- First make sure your system support `task` command. You can install it from [Taskfile.dev](https://taskfile.dev/#/).
- After you setup the environment variables and `task`.

```bash
task generate:config
```

- Then adjust the `config.yml` file generated with the connection string of your target and given database to synchronize.
- To check the differences between the given database and the target database, run the following command:

```bash
task g:synch -- --env=<config environment> audit check <adapter>_<given_db> <adapter>_<target_db>
```

- To perform the synchronization, use the following command:

```bash
task g:synch -- --env=<config environment> audit synch <adapter>_<given_db> <adapter>_<target_db>
```

- To reverse check the difference from the given database to the target database, run the following command:

```bash
task g:synch -- --env=<config environment> audit reverse-check <adapter>_<given_db> <adapter>_<target_db>
```

## 🧪 Testing

- To run the unit tests, use the following command:

```bash
task test
```

## 🤝 Contributing

Contributing is welcome! See the [Contributing Guide](.github__/CONTRIBUTING.md) for guidelines.

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/YourFeature`).
3. Make your changes.
4. Run test `task test` to ensure everything is working.
5. Commit your changes (`git commit -m 'Add some feature'`).
6. Push to the branch (`git push origin feature/YourFeature`).
7. Open a Pull Request.

**Enjoy Coding :)**
