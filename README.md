# git-version-action

A GitHub Action that computes a semver tag version from your repository's
nearest existing semver tag, current branch, and commit distance.

## Outputs

| Name      | Description                                   |
|-----------|-----------------------------------------------|
| `version` | Computed version string (e.g. `v2.0.0+3`) |

## Usage

```yaml
jobs:
  version:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - id: git-version
        uses: your-org/git-version-action@v1

      - run: echo "version is ${{ steps.git-version.outputs.version }}"
```

## Version logic

### If the current commit already carries a semver tag

That tag is returned as-is.

### Otherwise, the next version is calculated

1. Walk the first-parent chain from HEAD until a semver tag `v<major>.<minor>.<patch>` is found.
2. Bump the component according to the branch pattern:

| Branch prefix         | Bump rule                                       |
|-----------------------|-------------------------------------------------|
| `rel/` or `release/`  | `v<major+1>.0.0`                                |
| `feat/` or `feature/` | `v<major>.<minor+1>.0`                          |
| *(anything else)*     | `v<major>.<minor>.<patch+1>`                    |

3. Append the **branch-id** (a 4-digit hash of the branch name) and the
   **commit count** since the tag:

   ```
   v<major>.<minor>.<patch>.<branch-id>+<commits>
   ```

4. If the branch is `main` or `master` the `.<branch-id>` segment is
   omitted:

   ```
   v<major>.<minor>.<patch>+<commits>
   ```

### Branch ID

The branch-id is computed as:

```
sha256(branchName)[0:4] % 10000  →  b%04d
```

For example: `feature/foo` → `b1432`.

## Branch naming reference

| Branch name              | Nearest tag | Result                    |
|--------------------------|-------------|---------------------------|
| `main`                   | `v1.0.0`    | `v1.0.1+3`               |
| `master`                 | `v1.0.0`    | `v1.0.1+3`               |
| `feat/foo`               | `v1.0.0`    | `v1.1.0.b1432+3`         |
| `feature/login`          | `v1.0.0`    | `v1.1.0.babcd+3`         |
| `rel/v2`                 | `v1.0.0`    | `v2.0.0.b5678+3`         |
| `release/v2`             | `v1.0.0`    | `v2.0.0.b9012+3`         |
| `bugfix/issue-42`        | `v1.0.0`    | `v1.0.1.b6264+3`         |

## Local testing

```bash
# Build the binary
go build -o git-version .

# Run inside any git repo that has at least one semver tag
cd /path/to/your/repo
/path/to/git-version-action/git-version
```

The binary reads the repository from the current working directory and
prints the computed version to stdout.

### Example

```bash
$ cd /tmp/demo
$ git init && git commit --allow-empty -m init && git tag v1.0.0
$ git commit --allow-empty -m "second commit"
$ git checkout -b feat/awesome
$ git commit --allow-empty -m "third commit"
$ ~/git-version-action/git-version
v1.1.0.b1432+2
```

## Development

```bash
go build -o git-version .
./git-version
```
