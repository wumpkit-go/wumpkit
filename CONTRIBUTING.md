# Contributing to Wumpkit

Thank you for considering contributing to Wumpkit, a lightweight Go library for interacting with the Discord API. We value contributions from the community to enhance the library’s functionality, reliability, and documentation. This document provides guidelines for contributing, including our branching strategy, commit conventions, and pull request (PR) process, to ensure a smooth and collaborative development experience.

## Getting Started

Before contributing, ensure you have the following prerequisites set up:
- **Go**: Install the latest stable version of Go (see [golang.org](https://golang.org/doc/install)).
- **Git**: Ensure Git is installed and configured.
- **Development Environment**: Clone the repository and run `go mod tidy` to fetch dependencies.
- **Testing**: Verify the test suite runs successfully with `go test ./...`.

Contributions may include bug fixes, new features, performance optimizations, documentation improvements, or additional tests. To avoid duplicated efforts, please follow the guidelines below based on the scope of your contribution.

## Contribution Workflow

### For Small Changes
Small changes, such as typo corrections or minor documentation updates, can be submitted directly as pull requests without prior discussion.

### For Larger Changes
For significant contributions, such as new features, refactors, or performance optimizations:
1. Check the [issue tracker](https://github.com/wumpkit-go/wumpkit/issues) and [discussions](https://github.com/wumpkit-go/wumpkit/discussions) to ensure your proposed change hasn’t already been reported or discussed.
2. If no existing issue or discussion covers your proposal, create a new issue or start a discussion to outline your idea and engage with maintainers and the community. This step helps align your contribution with the project’s goals and prevents wasted effort or rejection.

### Branching Strategy

Wumpkit follows a branching model inspired by Git Flow, maintaining two primary branches:
- **master**: Represents the production-ready, stable version of the library, updated only with merged changes from the `develop` branch during releases.
- **develop**: Contains the latest development work, serving as the integration point for all new features, fixes, and improvements.

All contributions must be based on the `develop` branch, and pull requests should target `develop`.

Branch names must follow this convention:
- Use a prefix indicating the type of change, followed by a slash and a concise, descriptive name (using hyphens for spaces):
  - `feature/`: For new features (e.g., `feature/add-magic-unicorn-login`).
  - `fix/`: For bug fixes (e.g., `fix/stop-gremlins-from-crashing-server`).
  - `refactor/`: For code refactoring without changing functionality (e.g., `refactor/untangle-spaghetti-code`).
  - `optimize/`: For performance improvements (e.g., `optimize/make-code-run-like-cheetah`).
  - `docs/`: For documentation changes (e.g., `docs/polish-magic-scroll-readme`).
  - `test/`: For adding or improving tests (e.g., `test/guard-against-bug-invasions`).

For more information on this branching model, refer to the [Git Flow branching model](https://nvie.com/posts/a-successful-git-branching-model/).

### Commit Conventions

Commits must adhere to the [Conventional Commits specification](https://www.conventionalcommits.org/en/v1.0.0/) to facilitate automated changelog generation and versioning. The format is:

[optional scope]: 
[optional body]
[optional footer(s)]

Common commit types include:
- `feat`: Introduces a new feature.
- `fix`: Addresses a bug.
- `refactor`: Modifies code without altering functionality.
- `perf`: Improves performance.
- `docs`: Updates documentation.
- `test`: Adds or improves tests.
- `chore`: Updates build processes or auxiliary tools.

Example:

feat(channel): add support for thread channels

Implement ThreadChannel struct
Update channelFromJson to handle thread types

Closes #123

This convention ensures a clear and consistent commit history.

### Submitting Pull Requests

To ensure your pull request is effective and not duplicated:
1. Review [existing PRs](https://github.com/wumpkit-go/wumpkit/pulls) to confirm no one else is working on the same change.
2. For larger changes, verify that an issue or discussion has been opened and approved by maintainers to align with project goals.

To submit a PR:
1. Fork the repository and clone it locally.
2. Create a new branch from `develop`: `git checkout -b feature/your-feature-name develop`.
3. Implement your changes and commit them using Conventional Commits.
4. Push your branch to your fork: `git push origin feature/your-feature-name`.
5. Open a pull request on GitHub, targeting the `develop` branch.
6. Provide a clear title and detailed description, referencing related issues or discussions (e.g., "Closes #123" or "Discussed in #456").
7. Ensure your PR passes all continuous integration (CI) checks, including tests and linters.

For detailed instructions, see GitHub’s guide on [Creating a pull request](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request).

### Code Style and Testing

- Adhere to Go’s coding standards, using `gofmt` for formatting and `go vet` for static analysis.
- Include unit tests for new features or bug fixes, leveraging Go’s `testing` package.
- Verify all tests pass before submitting: `go test ./...`.

### Code Review Process

Pull requests will be reviewed by maintainers for:
- Adherence to coding standards and project goals.
- Sufficient test coverage.
- Clear documentation of changes.
- Compliance with the Conventional Commits specification.

Feedback may be provided, and you may need to make revisions before your PR is merged. We appreciate your patience during this process.

### Licensing

By contributing to Wumpkit, you agree that your contributions are licensed under the [BSD 3-Clause License](LICENSE), as specified in the repository. Ensure any code or documentation you submit complies with this license.

### Reporting Issues

To report a bug or propose a feature, open an issue on GitHub with:
- A clear description of the bug or feature.
- Steps to reproduce (for bugs).
- Expected and actual behavior.
- Relevant logs or code snippets.

For broader ideas or suggestions, start a discussion in the [GitHub Discussions](https://github.com/wumpkit-go/wumpkit/discussions) to collaborate with the community.

## Contact

If you have questions or need assistance, please open an issue or join the [GitHub Discussions](https://github.com/wumpkit-go/wumpkit/discussions). We value your contributions and are committed to making Wumpkit a robust and reliable library for the community.

Thank you for contributing to Wumpkit!
