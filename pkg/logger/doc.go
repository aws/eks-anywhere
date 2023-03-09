/*
Package logger implements a simple way to init a global logger and access it
through a logr.Logger interface.

Message:

All messages should start with a capital letter.

Log level:

The loggers only support verbosity levels (V-levels) instead of semantic levels.
Level zero, the default, is important for the end user.
  - 0: You always want to see this.
  - 1: Common logging that you don't want to show by default.
  - 2: Useful steady state information about the operation and important log messages that may correlate to significant changes in the system.
  - 3: Extended information about changes. Somehow useful information to the user that is not important enough for level 2.
  - 4: Debugging information. Starting from this level, all logs are oriented to developers and troubleshooting.
  - 5: Traces. Information to follow the code path.
  - 6: Information about interaction with external resources. External binary commands, api calls.
  - 7: Extra information passed to external systems. Configuration files, kubernetes manifests, etc.
  - 8: Truncated external binaries and clients output/responses.
  - 9: Full external binaries and clients output/responses.

Logging WithValues:

Logging WithValues should be preferred to embedding values into log messages because it allows
machine readability.

Variables name should start with a capital letter.

Logging WithNames:

Logging WithNames should be used carefully.
Please consider that practices like prefixing the logs with something indicating which part of code
is generating the log entry might be useful for developers, but it can create confusion for
the end users because it increases the verbosity without providing information the user can understand/take benefit from.

Logging errors:

A proper error management should always be preferred to the usage of log.Error.
*/
package logger
