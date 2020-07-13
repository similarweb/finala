# Contributing

Welcome to Finala! If you are interested in contributing to the 
Finala's code repo then checkout this Contributor's Guide. 

For help on how to get started visit our [developer guide](https://raw.githubusercontent.com/similarweb/finala/docs/update-readme-md/docs/developers).

### Versioning

We use standard [semver](https://semver.org/) to mark versions. 

### Publish

The publish of a new release is done with [Github Releases](https://github.com/similarweb/finala/releases) for new versions and patches. 

The Docker image version is updated automatically in Dockerhub on every release (see [finala](https://hub.docker.com/r/similarweb/finala)

## Branches

We use `master` branch as the branch for the latest working **development** version. 
For stable version please only see [releases](https://github.com/similarweb/finala/releases).

For contributing to Finala please follow the next steps: 

1. Create a fork on your branch from `master` branch.
2. Checkout to your branch using our branch name conventions.
3. Once done, create a Pull Request following our guide lines. Use the `master` branch as the target.
4. For a merge: CI tests must succeed and 1 review from a core maintainer.

### Standard branch name conventions (git-flow):

```
hotfix/your_fix_name
bug/your_bug_name
feature/your_feature_name
docs/doc_theme
maintenance/your_change_name
```

