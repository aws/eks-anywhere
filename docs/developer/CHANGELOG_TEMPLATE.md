<!--

How to use this template
------------------------

1. Copy the contents into the desired location.
2. If you're creating the GitHub release, remove the ## [Version]
3. Add entries to the various sections ensuring you summarise the change; big changes generally
   require more explanation.
4. Append an issue or PR to each entry (prefer issues).
5. If a section doesn't contain any entries, remove the section.
6. Re-read all changes from the users perspective. Ensure the summary is written so that almost
   anyone with a foundational knowledge of EKS Anywhere can appreciate what the change means.
7. If all sections are empty, you shouldn't be creating a release.

When gathering feedback and applying suggestions:

* Re-read the suggest edits in the same fashion as (6).
* Double check for duplication, punctuation and grammar (ask for help as needed).

General tips
------------

* Avoid acronyms unless they're accepted industry wide. CRD is domain specific, use Custom
  Resource Definition instead. I/O is known industry wide as Input/Output so is fine.
* Use plain language - avoid jargon.
* Keep sentences concise but clear.

-->
## [Version](GitHub URL)

### Must read before upgrade
<!--
Discuss caveats a user _really_ needs to know before they upgrade. Its the kind of thing that if
unread could be disasterous. Perhaps the user needs to execute some commands before or after
running the upgrade; detail that here.
-->

### Deprecation
<!--
Highlight features, APIs or behavior that we no longer want the user to use/rely on.
-->

### API Change
<!--
Highlight Custom Resource Definition and CLI changes (additions go under Features).
-->

### Feature
<!--
Highlight new additions to Custom Resource Definitions, CLI or any tool we maintain for customers.
-->

### Tool Upgrade
<!--
Highlight all upgrades to tooling. Most of this information comes from the build tooling repo.
Format as follows:

* Tool Name: <from version> to  <to version>

If we support multiple versions format as follows:

* Tool Name:
  * <from version> to <to version>
  * <from version> to <to version>
  * ...
-->

### Bug
<!--
Highlight bug fixes for all applications.
-->

### Other
<!--
Highlight any other changes that would be useful for the user to know.
-->

### Supported Operating Systems
<!--
List all supported operating systems for each provider.
-->

|                     | vSphere | Bare Metal | Nutanix | CloudStack | Snow  |
| :----------:        | :-----: | :--------: | :-----: | :--------: | :---: |
| Ubuntu 20.04        | ✔       | ✔          | ✔       | —          | ✔     |
| Ubuntu 22.04        | ✔       | ✔          | ✔       | —          | —     |
| Bottlerocket 1.15.1 | ✔       | ✔          | —       | —          | —     |
| RHEL 8.7            | ✔       | ✔          | ✔       | ✔          | —     |
| RHEL 9.x            | —       | —          | ✔       | —          | —     |