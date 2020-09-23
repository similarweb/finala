# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v0.4.0 (September 23,2020)
### Added
- Build Finala Notifier URL according to defined Tags [[PR-104](https://github.com/similarweb/finala/pull/104)]
- Add AWS profile authentication [[PR-139](https://github.com/similarweb/finala/pull/139)]
- Support direct table filter link [[PR-146](https://github.com/similarweb/finala/pull/146)]
- Add tests for ElasticSearch storage [[PR-138](https://github.com/similarweb/finala/pull/138)]
### Changed
- Revamp Finala User Interface [[PR-126](https://github.com/similarweb/finala/pull/126)]
- Change download csv file with the name of the resource [[PR-141](https://github.com/similarweb/finala/pull/141)]
### Fixed
- Make sure all Kinesis streams are retrieved [[PR-145](https://github.com/similarweb/finala/pull/145)]
- Fix API exact match query [[PR-151](https://github.com/similarweb/finala/pull/151)]
- Fix docker compose UI build [[PR-153](https://github.com/similarweb/finala/pull/153)]
- Add storage for RDS total pricing [[PR-143](https://github.com/similarweb/finala/pull/143)]

## v0.3.5 (August 13,2020)
### Added
- Add NAT Gateways detection [[PR-128](https://github.com/similarweb/finala/pull/128)]
- Support Assume Role [[PR-130](https://github.com/similarweb/finala/pull/130)]

## v0.3.4 (July 29,2020)
### Added
- Add the option to filter resources by tags in the API [[PR-73](https://github.com/similarweb/finala/pull/73)]
- Add Redshift resource detection [[PR-74](https://github.com/similarweb/finala/pull/74)]
- Implement golangci-lint and golang fmt [[PR-78](https://github.com/similarweb/finala/pull/78)]
- Add CodeCov [[PR-79](https://github.com/similarweb/finala/pull/79)]
- Improve CI/CD [[PR-81](https://github.com/similarweb/finala/pull/81)]
- Add elastic ip resource detection [[PR-97](https://github.com/similarweb/finala/pull/97)]
- Add elasticsearch resource detection [[PR-98](https://github.com/similarweb/finala/pull/98)]
- Enable/Disable specific metrics detection [[PR-100](https://github.com/similarweb/finala/pull/100)]
- Load dynamically all collector resources [[PR-101](https://github.com/similarweb/finala/pull/101)]
- Add support to compare between executions API [[PR-106](https://github.com/similarweb/finala/pull/106)]
- Add APIGateway detection [[PR-111](https://github.com/similarweb/finala/pull/111)]
- Add support for Version notification on a new version [[PR-114](https://github.com/similarweb/finala/pull/114)]
- Create index per day [[PR-115](https://github.com/similarweb/finala/pull/115)]
### Changed
- Update README.md [[PR-89](https://github.com/similarweb/finala/pull/89)]
- Remove Notifier from Docker compose first run [[PR-90](https://github.com/similarweb/finala/pull/90)]
- Improve pricing tests [[PR-109](https://github.com/similarweb/finala/pull/109)]
- Improve dobdb tests [[PR-110](https://github.com/similarweb/finala/pull/110)]
### Fixed
- Ignore Neptune based instances for RDS instances collection [[PR-71](https://github.com/similarweb/finala/pull/71)]
- Fix the pricing for ELB/ELBV2 [[PR-76](https://github.com/similarweb/finala/pull/76)]
- Change GoRelease to have the correct semver tags [[PR-82](https://github.com/similarweb/finala/pull/82)]
- Fix the Dockerfile release download [[PR-83](https://github.com/similarweb/finala/pull/83)]
- Fix version preview [[PR-84](https://github.com/similarweb/finala/pull/84)]
- Don't change the Tag Key , use the original from AWS [[PR-86](https://github.com/similarweb/finala/pull/86)]
- Fix total records response [[PR-113](https://github.com/similarweb/finala/pull/113)]
- Fix wrong pricing api filters for aurora-mysql [[PR-123](https://github.com/similarweb/finala/pull/123)]
### Removed
- Remove totalSpend field [[PR-99](https://github.com/similarweb/finala/pull/99)]

## v0.3.3 (June 17,2020)
### Added
- Add version command [[PR-66](https://github.com/similarweb/finala/pull/66)]

## v0.3.2 (June 16,2020)
### Added
- Add the total price for notification group in slack [[PR-63](https://github.com/similarweb/finala/pull/63)]
### Fixed
- Fix a bug wich supports mutilple regions prices [[PR-64](https://github.com/similarweb/finala/pull/64)]

## v0.3.1 (June 14,2020)
### Fixed
- Fix a bug where we did not return the correct collector resource status

## v0.3.0 (June 14,2020)
- Add the option to notify by tags a notification group [[PR-56](https://github.com/similarweb/finala/pull/56)]
- Fix Dynamodb price calculation [[PR-57](https://github.com/similarweb/finala/pull/57)]
- Add Kinesis unused resource detection [[PR-59](https://github.com/similarweb/finala/pull/59)]

## v0.2.1 (June 2,2020)
- Add Neptune unused database detection [[PR-52](https://github.com/similarweb/finala/pull/52)]

## v0.2.0 (May 26,2020)
- Split components to UI, API and collector. Support elasticsearch storage [[PR-53](https://github.com/similarweb/finala/pull/53)]

## v0.1.8 (May 26,2020)
- Show History view [[PR-43](https://github.com/similarweb/finala/pull/43)]

## v0.1.7 (April 19,2020)
- New dashboard. [[PR-43](https://github.com/similarweb/finala/pull/43)]
- Export table to CSV. [[PR-43](https://github.com/similarweb/finala/pull/43)]

## v0.1.6 (March 15,2020)
### Added
- Detect last activity of IAM user. [[PR-36](https://github.com/similarweb/finala/pull/36)]
- Notify new Finala version release. [[PR-31](https://github.com/similarweb/finala/pull/31)]
- Improve CI process. [[PR-33](https://github.com/similarweb/finala/pull/33)]
- Add AWS ELBV2 detection. [[PR-30](https://github.com/similarweb/finala/pull/30)]
- Dockerize the project. [[PR-40](https://github.com/similarweb/finala/pull/40)]
- Notify when new Finala version release. [[PR-43](https://github.com/similarweb/finala/pull/43)]

## v0.1.5 (February 4,2020)
### Added
-  Support MySQL. [[PR-27](https://github.com/similarweb/finala/pull/27)]

## v0.1.4 (December 15,2019)
### Added
-  Add AWS unused volumes detection. [[PR-24](https://github.com/similarweb/finala/pull/24)]

## v0.1.3 (December 15,2019)
### Added
-  Add AWS token session as optional parameter for session authentication.

## v0.1.2 (December 15,2019)
### Added
-  Add a possibility to use environment variables for AWS credentials instead of storing key/secret in config.yaml. [[PR-15](https://github.com/similarweb/finala/pull/15)]

## v0.1.1 (December 3,2019)
### Added
- Print analyze result to stdout

## v0.1.0 (December 2,2019)
### Added
- Init project
