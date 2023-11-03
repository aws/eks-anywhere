## [Version]

<!---
* EACH ENTRY SHOULD BE A SUMMARY OF THE CHANGE; BIG CHANGES GENERALLY REQUIRE MORE EXPLANATION.
* EACH ENTRY SHOULD BE ACCOMPANIED BY EITHER A PUBLIC ISSUE OR A PR (PREFER ISSUES OVER PRS).
* REREAD ALL ENTRIES FROM THE USERS PERSPECTIVE. WILL IT MAKE SENSE?
* IF A SECTION CONTAINS NO ENTRIES, REMOVE THE SECTION.
* IF ALL SECTIONS ARE EMPTY, YOU SHOULDN'T BE DOING A RELEASE!
-->

### Must read before upgrade
<!--
Discuss things a user _really_ needs to know before they upgrade. Its the kind of thing that if unread could
be disasterous.
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
Highlight new additions to Custom Resource Definitions, CLI or any tool we provide for customers.
-->

### Tool Upgrade
<!--
Highlight all upgrades to tooling. Most of this information comes from the build tooling repo. Format as follows.

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
Highlight any other changes such as tweaks to behavior.
-->

### Supported Operating Systems
<!--
List all supported OSs for each provider.
-->

|                     | vSphere | Bare Metal | Nutanix | CloudStack | Snow  |
| :----------:        | :-----: | :--------: | :-----: | :--------: | :---: |
| Ubuntu 20.04        | ✔       | ✔          | ✔       | —          | ✔     |
| Ubuntu 22.04        | ✔       | ✔          | ✔       | —          | —     |
| Bottlerocket 1.15.1 | ✔       | ✔          | —       | —          | —     |
| RHEL 8.7            | ✔       | ✔          | ✔       | ✔          | —     |
| RHEL 9.x            | —       | —          | ✔       | —          | —     |