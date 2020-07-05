# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
- Add the option to filter resources by tags in the API [[PR-73](https://github.com/similarweb/finala/pull/73)]
- Ignore Neptune based instances for RDS instances collection
- Implement golangci-lint and golang fmt

## [0.3.3]
### Added
- Add version command [[PR-66](https://github.com/similarweb/finala/pull/66)]

## [0.3.2]
### Added
- Add the total price for notification group in slack [[PR-63](https://github.com/similarweb/finala/pull/63)]
### Fixed
- Fix a bug wich supports mutilple regions prices [[PR-64](https://github.com/similarweb/finala/pull/64)]

## [0.3.1]
### Fixed
- Fix a bug where we did not return the correct collector resource status

## [0.3.0] 
- Add the option to notify by tags a notification group [[PR-56](https://github.com/similarweb/finala/pull/56)]
- Fix Dynamodb price calculation [[PR-57](https://github.com/similarweb/finala/pull/57)]
- Add Kinesis unused resource detection [[PR-59](https://github.com/similarweb/finala/pull/59)]

## [0.2.1] 
- Add Neptune unused database detection [[PR-52](https://github.com/similarweb/finala/pull/52)]

## [0.2.0] 
- Split components to UI, API and collector. Support elasticsearch storage [[PR-53](https://github.com/similarweb/finala/pull/53)]

## [0.1.8] 
- Show History view [[PR-43](https://github.com/similarweb/finala/pull/43)]

## [0.1.7] 
- New dashboard. [[PR-43](https://github.com/similarweb/finala/pull/43)]
- Export table to CSV. [[PR-43](https://github.com/similarweb/finala/pull/43)]

## [0.1.6] - 2020-03-15
### Added
- Detect last activity of IAM user. [[PR-36](https://github.com/similarweb/finala/pull/36)]
- Notify new Finala version release. [[PR-31](https://github.com/similarweb/finala/pull/31)]
- Improve CI process. [[PR-33](https://github.com/similarweb/finala/pull/33)]
- Add AWS ELBV2 detection. [[PR-30](https://github.com/similarweb/finala/pull/30)]
- Dockerize the project. [[PR-40](https://github.com/similarweb/finala/pull/40)]
- Notify when new Finala version release. [[PR-43](https://github.com/similarweb/finala/pull/43)]

## [0.1.5] 
### Added
-  Support MySQL. [[PR-27](https://github.com/similarweb/finala/pull/27)]

## [0.1.4] 
### Added
-  Add AWS unused volumes detection. [[PR-24](https://github.com/similarweb/finala/pull/24)]

## [0.1.3] - 2019-15-11
### Added
-  Add AWS token session as optional parameter for session authentication.

## [0.1.2] - 2019-12-11
### Added
-  Add a possibility to use environment variables for AWS credentials instead of storing key/secret in config.yaml. [[PR-15](https://github.com/similarweb/finala/pull/15)]

## [0.1.1] - 2019-12-03
### Added
- Print analyze result to stdout

## [0.1.0] - 2019-12-03
### Added
- Init project
