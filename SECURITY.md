# Security Policy

This document outlines security procedures and general policies for the Enduro
project.

**Contents**

- [Security Policy](#security-policy)
	- [Reporting a security vulnerability](#reporting-a-security-vulnerability)
	- [Disclosure policy](#disclosure-policy)
	- [Supported versions](#supported-versions)
	- [Reporting general bugs](#reporting-general-bugs)

## Reporting a security vulnerability

The Enduro development team takes security seriously and will investigate all
reported vulnerabilities.

If you would like to report a vulnerability or have a security concern regarding
Enduro, **please do not file a public issue in our GitHub repository.** It is
critical to the safety of other users that security issues are reported in a
secure manner. Instead, please email a report to:

* [security@artefactual.com](mailto:security@artefactual.com)

We will be better able to evaluate and respond to your report if it includes
all the details needed for us to reproduce the issue locally. Please include
the following information in your email:

* The version of Enduro you are using.
* Basic information about your installation environment, including operating
  system and dependency versions.
* Steps to reproduce the issue.
* The resulting error or vulnerability.
* If there are any error logs related to the issue, please include the
  relevant parts as well.

Your report will be acknowledged within 2 business days, and we’ll follow up
with a more detailed response indicating the next steps we intend to take
within 1 week.

If you haven’t received a reply to your submission after 5 business days of
the original report, please email Artefactual's info address:
[info@artefactual.com](info@artefactual.com)

Any information you share with the Enduro development team as a part of this
process will be kept confidential within the team. If we determine that the
vulnerability is located upstream in one of the libraries or dependencies that
Enduro uses, we may need to share some information about the report with the
dependency’s core team - in this case, we will notify you before proceeding.

If the vulnerability is first reported by you, we will credit you with the
discovery in the public disclosure, unless you tell us you would prefer to
remain anonymous.

## Disclosure policy

When the Enduro development team receives a security bug report, we will assign
it to a primary handler. This person will coordinate the fix and release
process, involving the following steps:

* Confirm the problem and determine the affected versions.
* Audit code to find any similar potential problems.
* Prepare fixes for all releases still under maintenance. These fixes will be
  released as fast as possible.

Once new releases and/or security patches have been prepared, tested, and made
publicly available, we will publicize the fix and encourage users to upgrade (or
apply the supplied patch) as soon as possible. Any internal tickets created in
our issue tracker related to the issue will be made public after disclosure, and
referenced in the release notes for the new version(s).

## Supported versions

In the case of a confirmed security issue, we will add the fix to the most
recent stable branch and the current development branch. If the severity of the
issue is high, we may in some cases also backport the fix to the previous stable
branch as well (e.g. `stable/1.11.x`) so that users running a legacy version
have the option of adding the fix as a patch to their local installations. We
will attempt to ensure that fixes, and/or a confirmed workaround that resolves
the security issue, are available prior to disclosing any security issues
publicly.

## Reporting general bugs

If you have discovered an issue in Archivematica that is **not related to a
security vulnerability**, we welcome you to file an issue in the
[Enduro repository](https://github.com/artefactual-sdps/enduro).
